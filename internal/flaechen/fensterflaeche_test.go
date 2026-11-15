// **Das Fenster als vierte Fläche** — C04
// (#2229).
//
// Die Läufe hier fuhren Kommandozeile, MCP und Dienst als Prozesse; das
// Fenster stand bis C04 in genau zwei Läufen daneben, und jeder Kopf sagte,
// seine Seite prüfe `internal/app`. Damit verglich kein Lauf, ob die vier
// Flächen **dieselbe** Antwort geben — nur, dass drei es tun.
//
// **Wie**: dieselbe Fassade, die die Hülle dem Fenster gibt (`app.NeueApp`,
// `Starten`, `NeueBindungen`), über **demselben** Bestand wie die Prozesse
// und im selben Zuhause. **Was das nicht ist**: das native Fenster — die
// Serialisierung durch Wails und die Oberfläche prüfen die Hüllenprobe und
// der Vertragsrundgang (B01, C01).
package flaechen_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// fensterAus öffnet die Bindungen des Fensters über dem Bestand einer Umgebung.
//
// `dialog` ist, was ein Dateidialog antwortet — wie die Hülle es über
// `MitDateiwahl` setzt; ohne Angabe bricht jeder Dialog ab.
func fensterAus(t *testing.T, umgebung []string, dialog ...string) *app.Bindungen {
	t.Helper()

	// **Dieselbe Umgebung wie die Prozesse** — Bestand, Dienstordner,
	// Schlüsselbund, Schleife —, und nie das Zuhause der Nutzenden (§3):
	// nennt die Umgebung keines, bekommt dieser Lauf ein eigenes.
	// **Ein echtes Zuhause aus der Umgebung wird übergangen**, nicht
	// übernommen (Zweitblick über Phase C, Befund 5).
	t.Setenv("HOME", t.TempDir())
	echt := echtesZuhause(t)

	for _, zeile := range umgebung {
		name, wert, gefunden := strings.Cut(zeile, "=")
		if gefunden && name != "" && os.Getenv(name) != wert && (name != "HOME" || wert != "" && wert != echt) {
			t.Setenv(name, wert)
		}
	}

	anwendung := app.NeueApp(bestandsortAus(t, umgebung))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return strings.Join(dialog, ""), nil
	})
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("das Fenster startet nicht: %v", fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	return app.NeueBindungen(anwendung)
}

// wieWails ruft eine Bindung, wie Wails sie ruft: eine Panik des Tors wird zu
// dem Fehler, den die Oberfläche aus ihrer Marke liest (`schaltertor.go`).
// Eine andere Panik bleibt eine.
func wieWails(ruf func() *app.Fehler) (fehler *app.Fehler) {
	defer func() {
		grund := recover()
		if grund == nil {
			return
		}

		text, istFehler := grund.(error)
		if !istFehler {
			panic(grund)
		}

		roh, markiert := strings.CutPrefix(text.Error(), app.Fehlermarke)
		if !markiert || json.Unmarshal([]byte(roh), &fehler) != nil {
			panic(grund)
		}
	}()

	return ruf()
}

// bindungRufen ruft eine Bindung beim Namen und mit Nullwerten — für Fälle,
// in denen das Tor antwortet, bevor ein Argument zählt (die Schalter).
func bindungRufen(t *testing.T, bindungen *app.Bindungen, name string) *app.Fehler {
	t.Helper()

	methode := reflect.ValueOf(bindungen).MethodByName(name)
	if !methode.IsValid() {
		t.Fatalf("keine Bindung %q", name)
	}

	argumente := []reflect.Value{reflect.ValueOf(t.Context())}
	for stelle := 1; stelle < methode.Type().NumIn(); stelle++ {
		argumente = append(argumente, reflect.Zero(methode.Type().In(stelle)))
	}

	for _, antwort := range methode.Call(argumente) {
		if fehler, ist := antwort.Interface().(*app.Fehler); ist {
			return fehler
		}

		if antwort.Kind() == reflect.Struct {
			if feld := antwort.FieldByName("Fehler"); feld.IsValid() {
				fehler, ist := feld.Interface().(*app.Fehler)
				if !ist {
					t.Fatalf("%s: das Feld Fehler ist %s und kein *app.Fehler", name, feld.Type())
				}

				return fehler
			}
		}
	}

	return nil
}
