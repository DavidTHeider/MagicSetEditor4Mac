package app

// **Die langen Datei-Läufe des Fensters im Auftragsbuch** — FMT-03.AK-02
// (#1692).
//
// Ein Druckbogen, eine MTG-JSON-Datei, eine Exportvorlage, ein Installer:
// an der Kommandozeile zeigen sie ihren Stand am Terminal und enden mit ⌃C,
// über MCP und am Dienst über das Protokoll. **Am Fenster tragen sie sich
// ins Auftragsbuch ein** — dasselbe, das Preise und Sammeldaten führt —,
// melden jedes gezählte Stück als [EreignisAuftrag] und lassen sich über
// [Bindungen.AuftragAbbrechen] anhalten. Gehalten wird zwischen zwei
// Stücken (`service.Schritt`), damit nur Ganzes stehen bleibt (AK-04).
//
// **Wiederholt wird, indem man den Knopf noch einmal drückt** — wie an der
// Kommandozeile derselbe Befehl. Einen Lauf ohne den Menschen davor, der
// ihn wieder anstößt, gibt es am Fenster nicht.

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// EreignisAuftrag meldet den Stand eines Datei-Laufs — beim Beginn, bei jedem
// gezählten Stück und am Ende. Die Nutzlast ist ein [AuftragDTO].
const EreignisAuftrag = "auftrag:stand"

// imDateilauf fährt einen Lauf des Fensters im Buch — `was` ist der Name der
// Datei, um die es geht, und `fehler` liest aus der Antwort, wie er endete.
func imDateilauf[T any](
	a *App, ctx context.Context, art, pfad string,
	fahren func(context.Context) T, fehler func(T) *Fehler,
) T {
	lauf, ende := a.dateilauf(ctx, art, filepath.Base(pfad))
	ergebnis := fahren(lauf)
	ende(fehlerAus(fehler(ergebnis)))

	return ergebnis
}

// dateilauf trägt einen Lauf ins Buch ein und gibt seinen Kontext — mit
// Abbruch und Melder — und was ihn beendet.
func (a *App) dateilauf(ctx context.Context, art, was string) (lauf context.Context, ende func(error)) {
	buch := a.auftragsbuch()
	lauf, abbrechen := context.WithCancel(ctx)
	id := buch.Beginnen(art, was)

	buch.Abbrechbar(id, abbrechen)
	a.meldeAuftrag(id)

	lauf = service.MitFortschritt(lauf, func(stand domain.Fortschritt) {
		buch.Zaehlen(id, stand)
		a.meldeAuftrag(id)
	})

	return lauf, func(fehler error) {
		buch.Beenden(id, fehler)
		abbrechen()
		a.meldeAuftrag(id)
	}
}

// meldeAuftrag schickt dem Fenster den Stand eines Auftrags.
func (a *App) meldeAuftrag(id int64) {
	for _, auftrag := range a.auftragsbuch().Auftraege() {
		if auftrag.ID == id {
			a.melden(a.fenster, EreignisAuftrag, auftragAlsDTO(auftrag, a.uhr()))
		}
	}
}

// AuftragAbbrechen hält einen laufenden Datei-Lauf an und gibt das Buch.
func (a *App) AuftragAbbrechen(ctx context.Context, id string) AuftraegeAntwort {
	ctx = a.imProfil(ctx)

	kennung, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		err = fmt.Errorf("%w: %q ist keine Kennung eines Auftrags", domain.ErrValidation, id)
	} else {
		err = a.auftragsbuch().Abbrechen(kennung)
	}

	if err != nil {
		return AuftraegeAntwort{Auftraege: []AuftragDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return a.Auftraege(ctx)
}

// AuftragAbbrechen — die Fensterbindung.
func (b *Bindungen) AuftragAbbrechen(ctx context.Context, id string) AuftraegeAntwort {
	return b.app(ctx).AuftragAbbrechen(b.ruf(ctx), id)
}
