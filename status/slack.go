package status

import (
	"context"
	"errors"
	"fmt"
	"gr0/statusbot/status/util"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type statusBot struct {
	client       *slack.Client
	socketClient *socketmode.Client
	scheduler    *gocron.Scheduler
	channelID    string
	reports      map[string]bool
	ignored      map[string]bool
	debug        bool
	mu           sync.Mutex
}

func NewStatusBot(botToken, appToken, channelID string, debug bool, ignored map[string]bool) *statusBot {
	client := slack.New(botToken, slack.OptionDebug(debug), slack.OptionAppLevelToken(appToken))
	bot := &statusBot{
		client: client,
		socketClient: socketmode.New(
			client,
			socketmode.OptionDebug(debug),
			socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
		),
		scheduler: gocron.NewScheduler(time.UTC),
		channelID: channelID,
		reports:   make(map[string]bool),
		ignored:   ignored,
		debug:     debug,
	}

	slog.Info("Starting backfilling data")
	err := bot.backfillData()
	if err != nil {
		slog.Error("Error backfilling data", "error", err)
	}
	slog.Info("Finished backfilling data")

	_, err = bot.scheduler.Every(1).Day().At("23:55").Do(func() { bot.sendSummaryMessage() })
	if err != nil {
		slog.Error("Error scheduling task", "error", err)
	}
	bot.scheduler.StartAsync()

	return bot
}

func (b *statusBot) Run(ctx context.Context) {
	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				switch event.Type {
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						slog.Error("Could not type cast the event to the EventsAPIEvent", "event", event)
						continue
					}
					socketClient.Ack(*event.Request)

					err := b.handleEventMessage(eventsAPIEvent)
					if err != nil {
						slog.Error("Error handling event message", "error", err)
					}
				}
			}
		}
	}(ctx, b.client, b.socketClient)

	log.Println("Started bot")
	b.socketClient.Run()
}

func (b *statusBot) handleEventMessage(event slackevents.EventsAPIEvent) error {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			msg, ok := innerEvent.Data.(*slackevents.MessageEvent)
			if !ok {
				return errors.New("could not type cast the event to the MessageEvent")
			}
			err := b.handleUserStatusUpdate(msg.User, msg.Text)
			if err != nil {
				return fmt.Errorf("failed to handle user status update: %w", err)
			}
		}

	default:
		return errors.New("unsupported event type")
	}
	return nil
}

func (b *statusBot) handleUserStatusUpdate(user, text string) error {
	// get user name
	u := util.GetUserFromMessage(user, text)
	usr, err := b.client.GetUserInfo(u)
	if err != nil {
		slog.Error("Error getting user info", "error", err)
		return nil
	}

	b.reportUserStatus(usr.Profile.DisplayName)
	return nil
}

func (b *statusBot) reportUserStatus(user string) {
	// don't report ignored user status
	if b.ignored[user] {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.reports[user] = true
	slog.Info("User reported status", "user", user)
}

func (b *statusBot) getReportedUsers() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	users := make([]string, 0, len(b.reports))
	for user := range b.reports {
		users = append(users, user)
	}

	b.reports = make(map[string]bool)
	return users
}

func (b *statusBot) getChannelUsers() []string {
	users, _, err := b.client.GetUsersInConversation(&slack.GetUsersInConversationParameters{
		ChannelID: b.channelID,
	})
	if err != nil {
		slog.Error("Error getting users in conversation", "error", err)
		return nil
	}

	userNames := make([]string, 0)
	for _, user := range users {
		usr, err := b.client.GetUserInfo(user)
		if err != nil {
			slog.Error("Error getting user info", "error", err)
			return nil
		}
		userNames = append(userNames, usr.Profile.DisplayName)
	}

	return userNames
}

func (b *statusBot) sendSummaryMessage() error {
	slog.Info("Sending summary message")

	reported := b.getReportedUsers()
	reportedUsers := ""
	if len(reported) > 0 {
		reportedUsers = util.ConnectValues(reported)
	} else {
		reportedUsers = "No one reported status"
	}

	channelUsers := b.getChannelUsers()
	notRportedUsers := util.ConnectValues(util.NotReportedUsers(reported, channelUsers))

	if !util.IsWeekend(time.Now()) {
		attachment := slack.Attachment{}
		attachment.Fields = []slack.AttachmentField{
			{
				Title: "Date",
				Value: time.Now().Format("2006-01-02"),
			},
			{
				Title: "Team members that reported status",
				Value: reportedUsers,
			},
			{
				Title: "Team members that didn't report status",
				Value: notRportedUsers,
			},
		}

		attachment.Text = "Daily status report"
		attachment.Color = "#4af030"

		_, _, err := b.client.PostMessage(b.channelID, slack.MsgOptionAttachments(attachment))
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
	} else {
		slog.Info("It's weekend, not sending summary message")
	}

	return nil
}

func (b *statusBot) backfillData() error {
	d := time.Now().Truncate(24 * time.Hour).UTC().Unix()
	rsp, err := b.client.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: b.channelID,
		Oldest:    strconv.FormatInt(d, 10),
		Limit:     1000,
	})
	if err != nil {
		return fmt.Errorf("failed to get conversation history: %w", err)
	}

	for _, msg := range rsp.Messages {
		err := b.handleUserStatusUpdate(msg.User, msg.Text)
		if err != nil {
			slog.Error("Error handling user status update", "error", err)
		}
	}

	return nil
}
