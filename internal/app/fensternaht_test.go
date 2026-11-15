package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Naht, über die das Fenster dem Frontend etwas sagt.**
//
// # Warum ein Ereignis und nicht ein Aufruf ins Backend
//
// Der Menüeintrag **löst aus**; was zu tun ist, entscheidet das Frontend
// (#162).
//
// Die andere Möglichkeit wäre gewesen, dass das Frontend nach jedem Wechsel
// meldet, welches Set offen ist, und Go es sich merkt. Das ist die größere
// Bauweise und die schlechtere: sie legt eine **fachliche** Entscheidung —
// „welches Set ist gemeint?" — in eine Schicht, die keine treffen soll (§2),
// und sie führt einen Zustand doppelt. Zwei Stellen, die dasselbe wissen,
// laufen auseinander; die Frage ist nur, wann.
//
// Der Preis gehört dazu: **Go weiß nach dieser Wahl nicht, ob etwas geschehen
// ist.** Ein ⌘S ohne offenes Set löst ein Ereignis aus, das niemand beantwortet
// — und das ist die richtige Antwort, denn „nichts tun" ist genau, was das
// Issue verlangt.
func TestDieMenueleisteMeldetDemFenster(t *testing.T) {
	anwendung := app.NeueApp(":memory:")

	var gemeldet []string

	anwendung.MitMeldung(func(_ context.Context, ereignis string, _ ...any) {
		gemeldet = append(gemeldet, ereignis)
	})

	handlungen := anwendung.Fensterhandlungen()

	handlungen.EinstellungenOeffnen()
	handlungen.SetSichern()
	handlungen.SetSichernUnter()
	handlungen.BestandSichern()
	handlungen.SetNeu()
	handlungen.SetZuletzt("7")
	handlungen.SetOeffnen()
	handlungen.SetSchliessen()
	handlungen.Drucken()
	handlungen.HilfeOeffnen()
	handlungen.SeitenleisteUmschalten()
	handlungen.SeiteZurueck()
	handlungen.SeiteVor()
	handlungen.NavigatorUmschalten()
	handlungen.InspectorUmschalten()

	erwartet := []string{
		app.EreignisEinstellungen, app.EreignisSetSichern, app.EreignisSetSichernUnter,
		app.EreignisBestandSichern, app.EreignisSetNeu, app.EreignisSetZuletzt,
		app.EreignisSetOeffnen,
		app.EreignisSetSchliessen,
		app.EreignisDrucken, app.EreignisHilfe, app.EreignisSeitenleiste,
		app.EreignisSeiteZurueck, app.EreignisSeiteVor,
		app.EreignisNavigator, app.EreignisInspector,
	}

	if len(gemeldet) != len(erwartet) {
		t.Fatalf("%d Meldungen, erwartet %d: %v", len(gemeldet), len(erwartet), gemeldet)
	}

	for stelle, name := range erwartet {
		if gemeldet[stelle] != name {
			t.Errorf("Meldung %d: %q, erwartet %q", stelle, gemeldet[stelle], name)
		}
	}
}

// **Die Ereignisnamen sind verschieden** — sonst löste ⌘S die Einstellungen aus.
//
// Ein Wächter über zehn Zeichenketten sieht albern aus. Er ist es nicht: sie
// stehen einmal hier und einmal im Frontend, und ein Tippfehler in einer von
// ihnen ergibt eine Anwendung, in der ein Kürzel nichts tut — ohne Fehler,
// ohne Meldung, ohne roten Lauf.
func TestDieEreignisnamenSindVerschieden(t *testing.T) {
	namen := map[string]bool{}

	for _, name := range []string{
		app.EreignisEinstellungen, app.EreignisSetSichern, app.EreignisSetSichernUnter,
		app.EreignisBestandSichern, app.EreignisSetNeu, app.EreignisSetZuletzt,
		app.EreignisSetOeffnen,
		app.EreignisSetSchliessen,
		app.EreignisDrucken, app.EreignisHilfe, app.EreignisSeitenleiste,
		app.EreignisSeiteZurueck, app.EreignisSeiteVor,
		app.EreignisNavigator, app.EreignisInspector,
	} {
		if name == "" {
			t.Error("ein leerer Ereignisname")
		}

		if namen[name] {
			t.Errorf("der Name %q kommt zweimal vor", name)
		}

		namen[name] = true
	}
}

// **Navigator und Inspector sind Ereignisse wie die Seitenleiste** — UXR-F14
// (#2029).
// Die Namen stehen einmal hier und einmal in `frontend/src/fenster.ts`; ein
// Tippfehler in einer der beiden ergäbe einen Eintrag, der nichts tut — ohne
// Fehler, ohne Meldung, ohne roten Lauf. Der Gegenpart steht in
// `fenster.test.ts`.
func TestNavigatorUndInspectorSindEreignisseMitFestemNamen(t *testing.T) {
	t.Parallel()

	if app.EreignisNavigator != "menue:navigator" || app.EreignisInspector != "menue:inspector" {
		t.Errorf("Ereignisse %q und %q", app.EreignisNavigator, app.EreignisInspector)
	}

	handlungen := app.NeueApp("").Fensterhandlungen()
	if handlungen.NavigatorUmschalten == nil || handlungen.InspectorUmschalten == nil {
		t.Error("das Fenster bekommt keine Handlung dafür")
	}
}
