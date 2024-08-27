package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectValues(t *testing.T) {
	s := ConnectValues([]string{"abc 123", "xyz 345", "ghy 23"})
	assert.Equal(t, s, "abc 123, xyz 345, ghy 23")
}

func TestNotReportedUsers(t *testing.T) {
	reported := []string{"abc 123", "xyz 345", "ghy 23"}
	present := []string{"abc 123", "kjl", "xyz 345", "ghy 23", "xxd123"}
	notReported := NotReportedUsers(reported, present)
	assert.Equal(t, []string{"kjl", "xxd123"}, notReported)
}

func TestGetUserFromMessage(t *testing.T) {
	usr := GetUserFromMessage("U01J9JZQZ8G", "Submission from test")
	assert.Equal(t, usr, "U01J9JZQZ8G")

	usr = GetUserFromMessage("", "Submission from <@U01J9JZQZ8G>")
	assert.Equal(t, usr, "U01J9JZQZ8G")

	usr = GetUserFromMessage("", "Submission from someone")
	assert.Equal(t, usr, "")
}

func TestIsWeekend(t *testing.T) {
	// Test for a Saturday
	saturday := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, IsWeekend(saturday))

	// Test for a Sunday
	sunday := time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC)
	assert.True(t, IsWeekend(sunday))

	// Test for a weekday
	weekday := time.Date(2022, time.January, 3, 0, 0, 0, 0, time.UTC)
	assert.False(t, IsWeekend(weekday))
}
