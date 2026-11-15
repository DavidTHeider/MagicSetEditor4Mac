package app

// **Ein abgebrochener Ruf des Fensters kehrt zurück** — BK-01
// (#2109).
//
// Der Bestand hat **eine** Verbindung (`SetMaxOpenConns(1)` in
// `storage.oeffnenMit`). Wer sie hält, lässt jeden anderen Ruf warten — und
// bis BK-01 wartete ein Ruf des Fensters mit dem Kontext des **Fensters**, den
// nichts abbricht: Wails brach den Ruf ab, und die Goroutine darunter hing
// weiter am Pool.

import (
	"bytes"
	"context"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// rufwartezeit ist, wie lange ein Ruf mindestens warten muss, damit er als
// „hängt am Pool" gilt — und wie lange ein abgebrochener höchstens braucht.
const (
	rufwartezeit    = 200 * time.Millisecond
	rufrueckkehr    = 5 * time.Second
	rufnachfreigabe = 5 * time.Second
)

func TestAbgebrochenerRufKehrtVomGehaltenenPoolZurueck(t *testing.T) {
	anwendung := gestartet(t)
	bindungen := NeueBindungen(anwendung)

	// **Erst die Wirkung über denselben Weg** (§3): ohne Griff antwortet er.
	if vorher := bindungen.Sets(t.Context()); vorher.Fehler != nil {
		t.Fatalf("ohne Griff: %+v", vorher.Fehler)
	}

	griff, fehler := anwendung.db.Conn(t.Context())
	if fehler != nil {
		t.Fatalf("den Pool greifen: %v", fehler)
	}

	freigegeben := false
	freigeben := func() {
		if !freigegeben {
			freigegeben = true
			_ = griff.Close() //nolint:errcheck // Rückgabe an den Pool
		}
	}
	// **Vor dem Beenden der Anwendung** (LIFO): ein hängender Ruf alter Art
	// käme sonst nie los, und der Test hinge statt rot zu werden.
	t.Cleanup(freigeben)

	// **Ein Abbruch ist kein klemmender Bestand** — und stünde sonst als
	// Fehler im Protokoll, wann immer sich ein Fenster schließt.
	protokoll := &bytes.Buffer{}
	vorher := slog.Default()
	t.Cleanup(func() { slog.SetDefault(vorher) })
	slog.SetDefault(slog.New(slog.NewTextHandler(protokoll, nil)))

	ruf, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	zurueck := make(chan SetsAntwort, 1)

	go func() { zurueck <- bindungen.Sets(ruf) }()

	select {
	case antwort := <-zurueck:
		t.Fatalf("der Ruf kam am gehaltenen Pool vorbei: %+v", antwort)
	case <-time.After(rufwartezeit):
	}

	abbrechen()

	select {
	case antwort := <-zurueck:
		if antwort.Fehler == nil || antwort.Fehler.Code != CodeAbgebrochen {
			t.Errorf("abgebrochen: %+v, erwartet Code %q", antwort.Fehler, CodeAbgebrochen)
		}
	case <-time.After(rufrueckkehr):
		t.Fatalf("der abgebrochene Ruf kehrte nach %v nicht zurück — er wartet weiter am Pool", rufrueckkehr)
	}

	if bytes.Contains(protokoll.Bytes(), []byte("level=ERROR")) {
		t.Errorf("der Abbruch steht als Fehler im Protokoll:\n%s", protokoll)
	}

	// **Und die Gegenprobe**: nach der Freigabe geht derselbe Weg wieder.
	freigeben()

	nachher, fertig := context.WithTimeout(t.Context(), rufnachfreigabe)
	defer fertig()

	if danach := bindungen.Sets(nachher); danach.Fehler != nil {
		t.Errorf("nach der Freigabe: %+v", danach.Fehler)
	}
}

// **Und ein Ruf, der auf einen fremden Schreiber wartet, ebenso** — und was
// er schreiben wollte, unterbleibt. Ein zweiter Prozess (`mse`, MCP, REST)
// hält die Schreibsperre; `busy_timeout(5000)` ließe den Ruf bis zu fünf
// Sekunden warten und **danach schreiben**, obwohl das Fenster ihn längst
// aufgegeben hat.
func TestAbgebrochenerRufSchreibtNichtNachDemFremdenSchreiber(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")
	anwendung := NeueApp(pfad)

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen

	bindungen := NeueBindungen(anwendung)
	vorher := einstellung(t, bindungen, domain.EinstellungBildkonflikt)

	neu := ""
	for _, wert := range vorher.Werte {
		if wert != vorher.Wert {
			neu = wert
		}
	}

	// **Erst die Wirkung über denselben Weg**: ohne fremden Schreiber
	// schreibt er — und wieder zurück.
	if antwort := bindungen.EinstellungSetzen(t.Context(), domain.EinstellungBildkonflikt, neu); antwort.Fehler != nil {
		t.Fatalf("ohne fremden Schreiber: %+v", antwort.Fehler)
	}

	if antwort := bindungen.EinstellungSetzen(t.Context(), domain.EinstellungBildkonflikt, vorher.Wert); antwort.Fehler != nil {
		t.Fatalf("zurück: %+v", antwort.Fehler)
	}

	fremd := fremderSchreiber(t, pfad)

	ruf, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	zurueck := make(chan SchreibAntwort, 1)

	go func() { zurueck <- bindungen.EinstellungSetzen(ruf, domain.EinstellungBildkonflikt, neu) }()

	select {
	case antwort := <-zurueck:
		t.Fatalf("der Ruf kam am fremden Schreiber vorbei: %+v", antwort)
	case <-time.After(rufwartezeit):
	}

	abgebrochen := time.Now()

	abbrechen()

	select {
	case <-zurueck:
		if gebraucht := time.Since(abgebrochen); gebraucht > time.Second {
			t.Errorf("zurück erst nach %v — der Ruf wartete weiter auf die Sperre", gebraucht)
		}
	case <-time.After(rufrueckkehr):
		t.Fatalf("der abgebrochene Ruf kehrte nach %v nicht zurück", rufrueckkehr)
	}

	fremd()

	if danach := einstellung(t, bindungen, domain.EinstellungBildkonflikt); danach.Wert != vorher.Wert {
		t.Errorf("der abgebrochene Ruf hat doch geschrieben: %q statt %q", danach.Wert, vorher.Wert)
	}
}

// einstellung liest eine Einstellung über die Bindung.
func einstellung(t *testing.T, bindungen *Bindungen, schluessel string) EinstellungDTO {
	t.Helper()

	antwort := bindungen.Einstellungen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Einstellungen: %+v", antwort.Fehler)
	}

	for _, eintrag := range antwort.Einstellungen {
		if eintrag.Schluessel == schluessel {
			if len(eintrag.Werte) < 2 {
				t.Fatalf("%s hat nur die Werte %v", schluessel, eintrag.Werte)
			}

			return eintrag
		}
	}

	t.Fatalf("%s steht nicht in den Einstellungen", schluessel)

	return EinstellungDTO{}
}

// fremderSchreiber hält die Schreibsperre aus einer **zweiten** Verbindung —
// so, wie es ein anderer Prozess täte — und gibt sie mit dem Rückgabewert frei.
func fremderSchreiber(t *testing.T, pfad string) func() {
	t.Helper()

	fremd, fehler := storage.Oeffnen(t.Context(), pfad)
	if fehler != nil {
		t.Fatalf("zweite Verbindung: %v", fehler)
	}

	griff, fehler := fremd.Conn(t.Context())
	if fehler != nil {
		t.Fatalf("Griff: %v", fehler)
	}

	if _, fehler := griff.ExecContext(t.Context(), "BEGIN IMMEDIATE"); fehler != nil {
		t.Fatalf("Schreibsperre: %v", fehler)
	}

	frei := false
	freigeben := func() {
		if frei {
			return
		}

		frei = true

		_, _ = griff.ExecContext(context.Background(), "ROLLBACK") //nolint:errcheck // Aufräumen
		_ = griff.Close()                                          //nolint:errcheck // Aufräumen
		_ = fremd.Close()                                          //nolint:errcheck // Aufräumen
	}
	t.Cleanup(freigeben)

	return freigeben
}

// **Und ein Ruf, der am Profilschloss wartet, ebenso** — Zweitblick über A
// und B, Befund 4: ein Wechsel des Profils hält `umschaltung` exklusiv, und
// `sync.RWMutex` kennt keinen Kontext. Der abgebrochene Ruf kam erst zurück,
// als der Wechsel fertig war.
func TestAbgebrochenerRufKehrtVomProfilschlossZurueck(t *testing.T) {
	anwendung := gestartet(t)
	bindungen := NeueBindungen(anwendung)

	if vorher := bindungen.Sets(t.Context()); vorher.Fehler != nil {
		t.Fatalf("ohne Sperre: %+v", vorher.Fehler)
	}

	anwendung.umschaltung.Lock()

	frei := false
	freigeben := func() {
		if !frei {
			frei = true

			anwendung.umschaltung.Unlock()
		}
	}
	t.Cleanup(freigeben)

	ruf, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	zurueck := make(chan SetsAntwort, 1)

	go func() { zurueck <- bindungen.Sets(ruf) }()

	select {
	case antwort := <-zurueck:
		t.Fatalf("der Ruf kam am gehaltenen Profilschloss vorbei: %+v", antwort)
	case <-time.After(rufwartezeit):
	}

	abbrechen()

	select {
	case antwort := <-zurueck:
		if antwort.Fehler == nil || antwort.Fehler.Code != CodeAbgebrochen {
			t.Errorf("abgebrochen: %+v", antwort.Fehler)
		}
	case <-time.After(rufrueckkehr):
		t.Fatal("der abgebrochene Ruf wartet weiter am Profilschloss")
	}

	freigeben()

	if danach := bindungen.Sets(t.Context()); danach.Fehler != nil {
		t.Errorf("nach der Freigabe: %+v", danach.Fehler)
	}
}
