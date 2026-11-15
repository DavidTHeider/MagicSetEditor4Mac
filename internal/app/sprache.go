package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bereit nimmt den Wails-Kontext entgegen.
//
// **Er kommt erst mit dem Fenster.** `Starten` öffnet den Bestand, bevor es ein
// Fenster gibt; alles, was zur Laufzeit ins Fenster greift — und die Menüleiste
// tut das —, braucht diesen hier.
// **Und hier bekommt die Menüleiste ihre Rückrufe** (#162). Bis dahin trägt
// sie die, die `main.go` beim Bauen mitgegeben hat — dieselben, aber ohne
// Kontext: gemeldet wird mit dem des Fensters, und den gibt es erst jetzt.
func (a *App) Bereit(ctx context.Context) {
	ctx = a.imProfil(ctx)

	a.fenster = ctx
	a.menuhandlungen = a.Fensterhandlungen()

	// **Und erst jetzt die Datei, die vor dem Fenster kam** (#891). Der Finder
	// schickt sie beim Start; ohne Kontext gibt es niemanden, dem man es
	// melden könnte.
	a.wartendeNachholen()
}

// SpracheSetzen stellt die Sprache der nativen Menüleiste ein.
//
// # Warum das eine Bindung ist
//
// Die eingestellte Sprache führt die Oberfläche — sie steht dort, wo sie
// gewählt wird. Die Menüleiste ist nativ und liegt in Go. Ohne diesen Weg
// bliebe die Leiste bei der Sprache, mit der die Anwendung gestartet ist,
// während der Rest des Fensters wechselt: eine Anwendung in zwei Sprachen.
func (a *App) SpracheSetzen(sprache string) SchreibAntwort {
	if !IstSprache(sprache) {
		return SchreibAntwort{Fehler: Nutzerfehler(domain.NewFieldError(
			"sprache", domain.SchluesselValidierung,
			"Die Sprache %q wird nicht geführt.", sprache))}
	}

	a.sprache = sprache
	a.menueNeu(a.fensterkontext())

	return SchreibAntwort{}
}

// MitMenue setzt die Naht zum Fenster.
func (a *App) MitMenue(setzen func(context.Context, *Menue)) { a.menueSetzen = setzen }

// MitMenuhandlungen hinterlegt die Rückrufe der Menüleiste.
//
// **Sie werden hier hinterlegt und nicht bei jedem Bauen mitgegeben**, weil die
// Leiste beim Sprachwechsel neu entsteht — und eine Leiste ohne Rückrufe ist
// eine, die aussieht wie vorher und nichts mehr tut.
func (a *App) MitMenuhandlungen(handlungen Menuhandlungen) { a.menuhandlungen = handlungen }
