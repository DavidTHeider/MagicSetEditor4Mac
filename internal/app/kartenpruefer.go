package app

// **Die Problemprüfung über ein ganzes Set, am Zeichenwerk** — PERF02
// (#2352).
//
// **Dieselben Teile wie die Vorschau**: Planbauer, Rasterer und Schriftsuche
// des Werks. Eine zweite Zusammensetzung gäbe eine zweite Antwort auf die
// Frage, ob eine Karte ein Problem hat. **Ohne die Nennungen**
// ([App.nennungen]): ihr Befund entsteht nur, wenn der Nachschlag eines
// Bildnachweises scheitert, und er hängt am Fenster.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// pruefgroesse ist die Vergrößerung, mit der die Prüfung eine Karte rastert —
// **die halbe der Vorschau** (PERF02, #2352). Was ein Mangel ist — eine
// fehlende Schrift, ein Bild, eine Maske, ein Symbol —, hängt nicht an der
// Größe, und die Rasterung ist der Teil, der hinter der Pforte steht: in
// voller Größe brauchte `Black` mit fünf Arbeitern 87 s, in halber 51 s.
// **Abweichen kann ein Befund, der aus dem Runden entsteht** — zwei Bilder,
// die nur in einer Größe verschieden groß ausfallen; das ist der Preis, und er
// steht hier.
const pruefgroesse = 0.5

// problemarbeiter ist, wie viele Karten das Werk zugleich prüft — die Hälfte
// der Kerne, mindestens zwei. **Der Plan steht nicht hinter der Pforte des
// Rasterers** und kostet so viel wie die Rasterung: an `Black` 31,6 ms Plan
// und 33,5 ms Raster je Karte, und mit zwei Arbeitern brauchten 3.788 Karten
// 131 s.
func problemarbeiter() int {
	return max(2, runtime.GOMAXPROCS(0)/2)
}

// problemteile hält die Prüfung eines Werks — gebaut beim ersten Ruf.
type problemteile struct {
	einmal   sync.Once
	pruefung *service.Problempruefung
}

// Problempruefung gibt die Prüfung dieses Werks — beim ersten Ruf gebaut,
// danach dieselbe, damit ihre Stände bleiben.
func (z *Zeichenwerk) Problempruefung() *service.Problempruefung {
	z.probleme.einmal.Do(func() {
		z.probleme.pruefung = service.NeueProblempruefung(z.Leser.Karten, kartenpruefer(z.Planbauer, z.Rasterer), z.problemabdruck, time.Now).
			MitArbeitern(problemarbeiter())
	})

	return z.probleme.pruefung
}

// planquelle und rasterquelle sind, was die Prüfung von Planbauer und
// Rasterer braucht — eine Methode je Teil.
type (
	planquelle interface {
		Kartenplan(ctx context.Context, karte *domain.CardRow) (*render.Plan, []domain.Befund, error)
	}
	rasterquelle interface {
		Zeichne(ctx context.Context, plan *render.Plan, vergroesserung float64, guete render.Guete) (*image.NRGBA, []render.Befund, error)
	}
)

// kartenpruefer zeichnet eine Karte wie die Vorschau und gibt, was dabei
// auffiel — die Befunde des Plans zuerst, wie in [App.zeichneVorschau].
func kartenpruefer(plaene planquelle, raster rasterquelle) func(context.Context, *domain.CardRow) ([]string, error) {
	return func(ctx context.Context, karte *domain.CardRow) ([]string, error) {
		plan, planbefunde, err := plaene.Kartenplan(ctx, karte)
		if err != nil {
			return nil, err
		}

		_, befunde, err := raster.Zeichne(ctx, plan, pruefgroesse, render.GueteMittel)
		if err != nil {
			return nil, err
		}

		meldungen := make([]string, 0, len(planbefunde)+len(befunde))
		for _, gefunden := range planbefunde {
			meldungen = append(meldungen, gefunden.String())
		}

		for _, gefunden := range befunde {
			meldungen = append(meldungen, gefunden.String())
		}

		return meldungen, nil
	}
}

// problemabdruck fasst zusammen, wovon eine Prüfung abhängt: den Setstand aus
// dem Bestand, den Stand von Schriftorten und Paketwurzeln, und **jede Datei
// der Pakete und Bilder des Sets** — eine bearbeitete Datei in einem
// vorhandenen Paket ändert die Zeit seiner Wurzel nicht (Zweitblick über
// Phase 3, Befund 2; [service.Stilspeicher.Ressourcenstand]).
func (z *Zeichenwerk) problemabdruck(ctx context.Context, setID int64) (string, error) {
	stand, err := z.Leser.Setstand(ctx, setID)
	if err != nil {
		return "", err
	}

	summe := sha256.Sum256([]byte(strconv.FormatInt(stand, 10) + "\x00" + z.Schriften.Stand() +
		"\x00" + z.Stile.Ressourcenstand(ctx, setID)))

	return hex.EncodeToString(summe[:8]), nil
}
