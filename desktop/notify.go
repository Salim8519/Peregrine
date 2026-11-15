package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/Salim8519/Peregrine/engine"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// The "download complete" notification carries two buttons, Open file and Show in folder,
// so a finished download can be used without opening the app (ADR 0021). Clicking the
// notification itself still opens the app.

const (
	doneCategory     = "peregrine-done"
	actionOpenFile   = "open-file"
	actionOpenFolder = "open-folder"
	donePrefix       = "done-" // notification ID: done-<download id>
)

// registerDoneCategory (re)registers the buttons in the user's language.
func (l *life) registerDoneCategory(lang string) {
	l.mu.Lock()
	if !l.notifyOK || l.notifyLang == lang && l.notifyLangSet {
		l.mu.Unlock()
		return
	}
	l.notifyLang, l.notifyLangSet = lang, true
	l.mu.Unlock()
	s := stringsFor(lang)
	err := wruntime.RegisterNotificationCategory(l.app.ctx, wruntime.NotificationCategory{
		ID: doneCategory,
		Actions: []wruntime.NotificationAction{
			{ID: actionOpenFile, Title: s.get("openFile")},
			{ID: actionOpenFolder, Title: s.get("openFolder")},
		},
	})
	if err != nil {
		log.Printf("peregrine: notification buttons: %v", err)
	}
}

// notifyDone shows the completed notification with its buttons.
func (l *life) notifyDone(d engine.Download) {
	st := l.app.m.Settings()
	if !l.shouldNotify(st) {
		return
	}
	go func() {
		err := wruntime.SendNotificationWithActions(l.app.ctx, wruntime.NotificationOptions{
			ID:         donePrefix + d.ID,
			Title:      stringsFor(st.Language).get("completed"),
			Body:       d.Filename,
			CategoryID: doneCategory,
			Data:       map[string]any{"id": d.ID},
		})
		if err != nil {
			log.Printf("peregrine: notification: %v", err)
		}
	}()
}

// notificationAction maps a click on a notification to an API method and download ID.
// method is "" for a click on the notification body (open the app).
func notificationAction(r wruntime.NotificationResponse) (method, id string) {
	switch r.ActionIdentifier {
	case actionOpenFile:
		method = "openFile"
	case actionOpenFolder:
		method = "showInFolder"
	default:
		return "", ""
	}
	if v, ok := r.UserInfo["id"].(string); ok && v != "" {
		id = v
	} else {
		id = strings.TrimPrefix(r.ID, donePrefix)
	}
	if id == "" || id == r.ID {
		return "", "" // not one of our done notifications
	}
	return method, id
}

// onNotification handles a click on any notification.
func (l *life) onNotification(res wruntime.NotificationResult) {
	if res.Error != nil {
		log.Printf("peregrine: notification response: %v", res.Error)
	}
	method, id := notificationAction(res.Response)
	if method == "" || l.app.svc == nil {
		l.app.showWindow()
		return
	}
	params, _ := json.Marshal(map[string]string{"id": id})
	if _, err := l.app.svc.Call(context.Background(), method, params); err != nil {
		log.Printf("peregrine: %s from notification: %v", method, err)
		l.app.showWindow() // e.g. the file was moved: show the app instead of nothing
	}
}
