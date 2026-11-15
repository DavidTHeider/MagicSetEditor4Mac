package fenster

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// huelle ist, was die Anwendung vom laufenden Fenster braucht.
//
// **Eine Schnittstelle und kein fester Ruf**: jede dieser sechs Handlungen
// braucht eine laufende Wails-Anwendung, und ein Lauf hat keine. Über die
// Schnittstelle prüft er die Verdrahtung — dass jede Naht der Anwendung hier
// ankommt, mit welchen Werten —, und nur [wailshuelle] bleibt ungemessen.
type huelle interface {
	melden(name string, daten []any)
	menueSetzen(leiste *application.Menu)
	dateiWaehlen(auswahl app.Dateiauswahl) (string, error)
	ablageLesen() (string, error)
	ablageSchreiben(text string) error

	// teilen zeigt den Teilen-Dialog des Systems (NET-04, #1708).
	teilen(pfad, text string) error
}

// errAblageStumm heißt: macOS hat die Zwischenablage nicht hergegeben.
//
// **Ein eigener Fehler, weil Wails 3 nur ein `bool` gibt** — Wails 2 gab einen
// `error`, und die Anwendung reicht ihn als Befund weiter (#1168).
var errAblageStumm = errors.New("die Zwischenablage hat nicht geantwortet")

// verdrahten setzt die Nähte der Anwendung auf das Fenster.
//
// **Die fünf Nähte an einer Stelle** — unter Wails 2 standen sie als
// `…Echt`-Funktionen in `internal/app` und wurden von `NeueApp` gesetzt. Seit
// Wails 3 kennt `internal/app` die Hülle nicht mehr (ENG-02), und die Nähte
// werden hier gesetzt, bevor das Fenster aufgeht.
//
// **Und die erste Leiste**, über denselben Halter wie jede spätere — ein
// eigenes Menü beim Start bliebe liegen, sobald die erste Meldung kommt.
func verdrahten(anwendung *app.App, fenster huelle, mit umsetzer) {
	anwendung.MitMeldung(func(_ context.Context, name string, daten ...any) {
		fenster.melden(name, nutzlast(daten))
	})

	halter := &menuehalter{}
	anwendung.MitMenue(func(_ context.Context, leiste *app.Menue) {
		if menue, neu := halter.erneuern(leiste, mit); neu {
			fenster.menueSetzen(menue)
		}
	})
	erste, _ := halter.erneuern(app.Menueleiste(
		anwendung.Fensterhandlungen(), app.MenuetexteFuer("de"), nil), mit)
	fenster.menueSetzen(erste)

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		return fenster.dateiWaehlen(auswahl)
	})

	anwendung.MitZwischenablage(
		func(context.Context) (string, error) { return fenster.ablageLesen() },
		func(_ context.Context, text string) error { return fenster.ablageSchreiben(text) },
	)

	anwendung.MitTeilnaht(func(_ context.Context, pfad, text string) error {
		return fenster.teilen(pfad, text)
	})
}
