package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Salim8519/Peregrine/engine"
)

// afterDoneMethods are the API methods that carry out the add dialog's "when done" choices
// (ADR 0023), in order: open the file, then show it in its folder.
func afterDoneMethods(ad engine.AfterDone) []string {
	var out []string
	if ad.OpenFile {
		out = append(out, "openFile")
	}
	if ad.OpenFolder {
		out = append(out, "showInFolder")
	}
	return out
}

// finished runs when a download completes: the "when done" choices, then the "download
// complete" popup or, without it, the system notification.
func (l *life) finished(d engine.Download) {
	if ms := afterDoneMethods(d.AfterDone); len(ms) > 0 && l.app.svc != nil {
		go func() {
			params, _ := json.Marshal(map[string]string{"id": d.ID})
			for _, m := range ms {
				if _, err := l.app.svc.Call(context.Background(), m, params); err != nil {
					log.Printf("peregrine: %s when done: %v", m, err)
				}
			}
		}()
	}
	if d.AfterDone.Popup && l.app.showDone(d) {
		return // the popup replaces the notification
	}
	l.notifyDone(d)
}
