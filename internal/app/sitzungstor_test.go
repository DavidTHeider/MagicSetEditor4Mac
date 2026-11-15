package app_test

// **Das Tor der Sitzung an der Bindungsgrenze** — Zweitblick über ID-02,
// Befund 4 (#1701).
//
// **Gesperrt oder abgemeldet hält jede gebundene Methode an**, bevor sie
// etwas tut — außer den wenigen, die die Sperranzeige und die Anmeldung
// brauchen. Wails hat keinen Haken vor einem Aufruf (beta.23,
// `pkg/application/services.go`); es fängt aber eine Panik in einer
// gebundenen Methode und gibt sie als Aufruffehler zurück
// (`pkg/application/bindings.go`, `(*BoundMethod).Call`).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// haelt sagt, ob der Ruf am Tor hielt.
func haelt(t *testing.T, ruf func()) (gehalten bool) {
	t.Helper()

	defer func() {
		if wert := recover(); wert != nil {
			fehler, istFehler := wert.(error)
			gehalten = istFehler && errors.Is(fehler, app.ErrSitzungVerschlossen)

			if !gehalten {
				t.Errorf("eine andere Panik: %v", wert)
			}
		}
	}()

	ruf()

	return false
}

func TestGesperrtHaeltJedeBindungAmTor(t *testing.T) {
	anwendung := faktoranwendung(t)
	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.KennwortWechseln(t.Context(), "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	// **Offen hält nichts** — die Probe, dass das Tor nicht immer zu ist.
	if haelt(t, func() { bindungen.Sets(t.Context()) }) {
		t.Fatal("offen hielt das Tor")
	}

	if antwort := bindungen.Sperren(t.Context()); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	for name, ruf := range map[string]func(){
		"Sets":           func() { bindungen.Sets(t.Context()) },
		"NutzerWechseln": func() { bindungen.NutzerWechseln(t.Context(), "1", "") },
		"PlatzFreigeben": func() { bindungen.PlatzFreigeben(t.Context(), "vorschauen") },
		"RESTStarten":    func() { bindungen.RESTStarten(t.Context(), "127.0.0.1", 0) },
		"KennwortWechseln": func() {
			bindungen.KennwortWechseln(t.Context(), probekennwort, "noch-ein-langes-kennwort")
		},
	} {
		if !haelt(t, ruf) {
			t.Errorf("%s ging gesperrt durch", name)
		}
	}

	// **Was die Sperranzeige braucht, geht durch.**
	for name, ruf := range map[string]func(){
		"Sitzungsstand": func() { bindungen.Sitzungsstand(t.Context()) },
		"Nutzer":        func() { bindungen.Nutzer(t.Context()) },
		"Einstellungen": func() { bindungen.Einstellungen(t.Context()) },
	} {
		if haelt(t, ruf) {
			t.Errorf("%s hielt am Tor", name)
		}
	}

	if antwort := bindungen.Entsperren(t.Context(), probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if haelt(t, func() { bindungen.Sets(t.Context()) }) {
		t.Error("entsperrt hielt das Tor")
	}
}
