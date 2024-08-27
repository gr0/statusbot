package util

import (
	"strings"
	"time"
)

func ConnectValues(values []string) string {
	s := strings.Join(values, ", ")
	return strings.TrimRight(s, ", ")
}

func NotReportedUsers(reported, present []string) []string {
	reportedUsers := make(map[string]bool)
	for _, user := range reported {
		reportedUsers[user] = true
	}

	notReported := make([]string, 0)
	for _, user := range present {
		p := reportedUsers[user]
		if !p {
			notReported = append(notReported, user)
		}
	}
	return notReported
}

func GetUserFromMessage(user, text string) string {
	if len(user) > 0 {
		return user
	} else {
		start := strings.IndexAny(text, "<@")
		end := strings.IndexAny(text, ">")
		if start+2 < end {
			return text[start+2 : end]
		}
	}
	return ""
}

func ReadIgnored(ignored string) map[string]bool {
	ignoredUsers := make(map[string]bool)
	users := strings.Split(ignored, ",")
	for _, user := range users {
		ignoredUsers[user] = true
	}
	return ignoredUsers
}

func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
