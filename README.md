# Slack Status Bot

A **very** simple status bot for monitoring user Slack activity, that was created for internal use. 
The user activity on a given channel and reporting who pasted the status using a workflow once a day.

## Configuration

To configure the bot you need to create the `.env` file and provide the following configuration 
options:

- `SLACK_BOT_TOKEN` - Slack bot token
- `SLACK_APP_TOKEN` - Slack application token
- `SLACK_CHANNEL_ID` - The identifier of the Slack channel where the bot will be reporting
- `DEBUG` - boolean value, set to `true` for Slack communication debugging
- `IGNORED_USERS` - list of users that should be ignored from reporting