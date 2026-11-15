package app

// **Die Probleme jedes Sets am Fenster** — AP-8, Teil 2
// (#2365).
// PERF02 (#2352) hat den Vertrag gelegt und die drei Flächen ohne Fenster
// bedient; hier kommt das Fenster dazu, mit seinem Aufrufer: den Kacheln der
// Übersicht.
//
// # Gelesen wird ohne zu prüfen, geprüft wird auf Anstoß
//
// **Die Prüfung läuft nie je Aufruf der Startseite** (`LEISTUNGSSCHNITTE.md`
// §11): `Setprobleme` liest die Stände, wie sie sind — ungeprüft, veraltet,
// unvollständig oder aktuell. `SetproblemePruefen` prüft jedes Set, als
// Auftrag im Buch der Hintergrundaufgaben, mit Fortschritt über die Karten
// aller Sets und mit Abbruch. **Ein abgebrochener Lauf nennt, was er schaffte,
// und die übrigen Sets, wie sie standen** — keines gilt danach als geprüft,
// das niemand prüfte.

import (
	"context"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SetproblemstandDTO ist der Stand der Prüfung eines Sets, mit seinem Namen.
type SetproblemstandDTO struct {
	SetID string `json:"setID"`
	Name  string `json:"name"`
	ProblemstandDTO
}

// SetproblemeAntwort trägt die Stände aller Sets.
type SetproblemeAntwort struct {
	Staende []SetproblemstandDTO `json:"staende"`
	Fehler  *Fehler              `json:"fehler"`
}

// ProblemwegAus baut den Weg über einer Prüfung — für das Fenster über dem
// Zeichenwerk, das auch die Vorschau zeichnet.
func ProblemwegAus(pruefung *service.Problempruefung) *Problemweg {
	return &Problemweg{pruefung: pruefung, uhr: time.Now}
}

// MitProblemweg setzt die Naht.
func (a *App) MitProblemweg(weg *Problemweg) { a.probleme = weg }

// Setprobleme gibt den Stand der Prüfung jedes Sets — **ohne zu prüfen**.
func (a *App) Setprobleme(ctx context.Context) SetproblemeAntwort {
	ctx = a.imProfil(ctx)

	return a.setprobleme(ctx, func(kopf domain.SetHead) (ProblemstandDTO, error) {
		return a.probleme.Stand(ctx, kopf.ID)
	})
}

// SetproblemePruefen prüft jedes Set und gibt die Stände dieses Laufs.
func (a *App) SetproblemePruefen(ctx context.Context) SetproblemeAntwort {
	ctx = a.imProfil(ctx)

	if a.probleme == nil {
		return SetproblemeAntwort{Staende: []SetproblemstandDTO{}}
	}

	lauf, ende := a.dateilauf(ctx, "probleme", "")
	erledigt, gesamt := 0, 0
	gezaehlt := false

	antwort := a.setprobleme(lauf, func(kopf domain.SetHead) (ProblemstandDTO, error) {
		vorher := erledigt

		// **Nach einem Abbruch wird gelesen, nicht geprüft** — so steht
		// jedes Set in der Antwort, wie es ist.
		if lauf.Err() != nil {
			return a.probleme.Stand(context.WithoutCancel(lauf), kopf.ID)
		}

		dto, err := a.probleme.Pruefen(lauf, kopf.ID, func(geprueft, _ int) {
			erledigt = vorher + geprueft
			service.Melde(lauf, domain.Fortschritt{Erledigt: erledigt, Gesamt: gesamt, Einheit: domain.EinheitKarte})
		})

		return dto, err
	}, func(koepfe []domain.SetHead) {
		// **Gezählt wird über die Karten aller Sets** — die Zahl steht im
		// Kopf, ohne ein Set zu lesen.
		for _, kopf := range koepfe {
			gesamt += kopf.Cards
		}

		gezaehlt = true
	})

	// **Ein Abbruch sagt sich**, auch wenn er vor dem ersten Set kam und
	// kein Ruf ihn meldete.
	if lauf.Err() != nil && antwort.Fehler == nil {
		antwort.Fehler = Uebersetze(ctx, lauf.Err())
	}

	if gezaehlt && antwort.Fehler == nil {
		service.Melde(lauf, domain.Fortschritt{Erledigt: gesamt, Gesamt: gesamt, Einheit: domain.EinheitKarte})
	}

	ende(fehlerAus(antwort.Fehler))

	return antwort
}

// setprobleme geht über die Sets und fragt `stand` je Set. Der erste Fehler
// steht in der Antwort; die übrigen Sets stehen trotzdem darin.
func (a *App) setprobleme(
	ctx context.Context,
	stand func(domain.SetHead) (ProblemstandDTO, error),
	vorab ...func([]domain.SetHead),
) SetproblemeAntwort {
	antwort := SetproblemeAntwort{Staende: []SetproblemstandDTO{}}

	if a.probleme == nil {
		return antwort
	}

	koepfe, err := a.leser().Setkoepfe(ctx)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	for _, zuerst := range vorab {
		zuerst(koepfe)
	}

	for _, kopf := range koepfe {
		dto, err := stand(kopf)
		antwort.Staende = append(antwort.Staende, SetproblemstandDTO{
			SetID: strconv.FormatInt(kopf.ID, 10), Name: kopf.Name, ProblemstandDTO: dto,
		})

		if err != nil && antwort.Fehler == nil {
			antwort.Fehler = Uebersetze(ctx, err)
		}
	}

	return antwort
}

// Setprobleme — die Fensterbindung.
func (b *Bindungen) Setprobleme(ctx context.Context) SetproblemeAntwort {
	return b.app(ctx).Setprobleme(b.ruf(ctx))
}

// SetproblemePruefen — die Fensterbindung.
func (b *Bindungen) SetproblemePruefen(ctx context.Context) SetproblemeAntwort {
	return b.app(ctx).SetproblemePruefen(b.ruf(ctx))
}
