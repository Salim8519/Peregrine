package main

import (
	"testing"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func TestNotificationAction(t *testing.T) {
	cases := []struct {
		name       string
		r          wruntime.NotificationResponse
		method, id string
	}{
		{"open file", wruntime.NotificationResponse{ID: "done-a1", ActionIdentifier: actionOpenFile, UserInfo: map[string]any{"id": "a1"}}, "openFile", "a1"},
		{"folder via id only", wruntime.NotificationResponse{ID: "done-b2", ActionIdentifier: actionOpenFolder}, "showInFolder", "b2"},
		{"body click", wruntime.NotificationResponse{ID: "done-a1", ActionIdentifier: "DEFAULT_ACTION"}, "", ""},
		{"other notification", wruntime.NotificationResponse{ID: "tray-hint", ActionIdentifier: actionOpenFile}, "", ""},
		{"empty id", wruntime.NotificationResponse{ID: "done-", ActionIdentifier: actionOpenFile}, "", ""},
	}
	for _, c := range cases {
		m, id := notificationAction(c.r)
		if m != c.method || id != c.id {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", c.name, m, id, c.method, c.id)
		}
	}
}
