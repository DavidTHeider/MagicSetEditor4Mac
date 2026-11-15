package app

// **Auch das Nachrichten der Leserverbindung kehrt vom Profilschloss
// zurück** — dieselbe Klasse wie `TestAbgebrochenerRufKehrtVomProfilschlossZurueck`
// (BK-01, Zweitblick über A und B, Befund 4), auf dem zweiten Weg zur Sperre:
// hat sich die Revision der Rechte bewegt, will [App.verbindungNachziehen]
// die Verbindung neu richten und wartet dafür exklusiv auf `umschaltung`.

import (
	"context"
	"testing"
	"time"
)

func TestDasNachrichtenKehrtVomGehaltenenProfilschlossZurueck(t *testing.T) {
	anwendung := gestartet(t)

	// **Die Revision hat sich bewegt** — sonst kehrt das Nachziehen vor der
	// Sperre um, und dieser Weg bliebe unbetreten.
	anwendung.gerichtet.Store(-1)

	anwendung.umschaltung.Lock()
	t.Cleanup(anwendung.umschaltung.Unlock)

	ruf, abbrechen := context.WithCancel(t.Context())
	zurueck := make(chan struct{})

	go func() {
		anwendung.verbindungNachziehen(ruf)
		close(zurueck)
	}()

	select {
	case <-zurueck:
		t.Fatal("das Nachziehen kam am gehaltenen Profilschloss vorbei")
	case <-time.After(rufwartezeit):
	}

	abbrechen()

	select {
	case <-zurueck:
	case <-time.After(rufrueckkehr):
		t.Fatal("das abgebrochene Nachziehen wartet weiter am Profilschloss")
	}

	// **Und es hat nichts gerichtet** — die Revision steht, wie sie stand.
	if gerichtet := anwendung.gerichtet.Load(); gerichtet != -1 {
		t.Errorf("gerichtet wurde auf %d", gerichtet)
	}
}
