package app_test

// **Jede gebundene Methode hält am Tor** — Nachlauf über ID-02, Befund 1
// (#1701).
//
// # Was der Nachlauf fand
//
// **Das Tor stand in `kontext()`**, und zwei Wege gingen daran vorbei:
// `IndexAbbrechen` und `KIBildAbbrechen` rufen es nie, und `SpracheSetzen`
// änderte die Sprache, bevor es kam — dort erkannte das Tor seinen Rufer
// nicht, weil zwischen ihm und der Bindung eine Methode der Anwendung lag.
// `TestGesperrtHaeltJedeBindungAmTor` prüfte fünf Bindungen; über die
// übrigen stand im Kopf eine Zusage, die kein Lauf hielt.
//
// # Was hier steht
//
// **Alle**, über Reflexion — außer denen, die die Sperranzeige braucht. Die
// Liste steht hier ein zweites Mal und mit Absicht: `sitzungstor_intern_test.go`
// hält sie gegen den Quelltext, dieser Lauf gegen das Verhalten.

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sperranzeige sind die Bindungen, die gesperrt durchgehen.
var sperranzeige = map[string]bool{
	"Sitzungsstand": true, "Entsperren": true, "EntsperrenMitSensor": true,
	"Abmelden": true, "LokalAnmelden": true, "KennwortWiederherstellen": true,
	"Nutzer": true, "Einstellungen": true, "SpracheSetzen": true, "Startbefund": true,
}

func TestGesperrtHaeltJedeGebundeneMethode(t *testing.T) {
	// **Nach draußen greift keiner** (§3): hielte eine Bindung nicht, liefe
	// sie mit Nullwerten los — und dann nicht gegen dieses Zuhause und
	// nicht gegen Scryfall.
	t.Setenv("HOME", t.TempDir())

	vorher := app.MitScryfalladresse("http://127.0.0.1:1/")
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	anwendung := faktoranwendung(t)
	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.KennwortWechseln(t.Context(), "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if antwort := bindungen.Sperren(t.Context()); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	typ := reflect.TypeOf(bindungen)
	wert := reflect.ValueOf(bindungen)
	gehalten := 0

	for i := range typ.NumMethod() {
		methode := typ.Method(i)
		if sperranzeige[methode.Name] {
			continue
		}

		argumente := make([]reflect.Value, methode.Type.NumIn()-1)
		for stelle := range argumente {
			argumente[stelle] = reflect.Zero(methode.Type.In(stelle + 1))
		}

		if !haelt(t, func() { wert.Method(i).Call(argumente) }) {
			t.Errorf("%s ging gesperrt durch", methode.Name)

			continue
		}

		gehalten++
	}

	// **Die Reichweite gehört zur Aussage** (§10): ein Lauf über keine
	// Bindung wäre grün.
	if gehalten < 400 {
		t.Errorf("nur %d Bindungen gehalten — die Aufzählung ist zu kurz", gehalten)
	}

	// **Die Sprache der Menüleiste geht gesperrt durch** — die Sperranzeige
	// spricht die Sprache des Fensters, und die Leiste soll es auch.
	if antwort := bindungen.SpracheSetzen(t.Context(), "en"); antwort.Fehler != nil {
		t.Errorf("SpracheSetzen gesperrt: %+v", antwort.Fehler)
	}
}

// TestDasTorBrauchtKeinenBestand hält fest, dass ein Tor ohne gestartete
// Anwendung nicht selbst scheitert — die Verdrahtung ruft Bindungen, bevor
// es eine Sitzung gibt.
func TestDasTorBrauchtKeinenBestand(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	bindungen := app.NeueBindungen(anwendung)

	defer func() {
		if wert := recover(); wert != nil {
			if fehler, ist := wert.(error); ist && errors.Is(fehler, app.ErrSitzungVerschlossen) {
				t.Errorf("ohne Sitzung hielt das Tor: %v", fehler)
			}
		}
	}()

	bindungen.Sitzungsstand(t.Context())
	bindungen.IndexAbbrechen(t.Context())
}
