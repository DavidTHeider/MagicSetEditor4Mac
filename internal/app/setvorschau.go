package app

// **Ein Set vorab ansehen** — FMT-03.AK-01
// (#1692).
//
// Was das Einfahren einer Datei täte — Name, Spiel und Stil, wie viele
// Karten, was sich nicht lesen ließ, ob die Quelle schon im Bestand liegt
// und **welche Pakete fehlen** —, bevor es geschieht. Gelesen wird auf
// demselben Weg wie beim Einfahren (`service/setlesen.go`); geschrieben wird
// nichts.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SetvorschauDTO ist die Vorschau, wie jede Fläche sie sieht.
type SetvorschauDTO struct {
	Pfad    string   `json:"pfad"`
	Name    string   `json:"name"`
	Spiel   string   `json:"spiel"`
	Stil    string   `json:"stil"`
	Karten  int      `json:"karten"`
	Befunde []string `json:"befunde"`

	// SchonOffen ist die Kennung des Sets aus derselben Quelle — leer, wenn
	// das Einfahren ein neues anlegte.
	SchonOffen string `json:"schonOffen"`

	// Bedarf sind die Pakete, die fehlen oder zu alt sind — wie
	// `Paketbedarf` sie für ein eingefahrenes Set nennt.
	Bedarf []AbhaengigkeitDTO `json:"bedarf"`
}

// SetvorschauAntwort ist die Vorschau samt Umschlag.
type SetvorschauAntwort struct {
	Vorschau    *SetvorschauDTO `json:"vorschau"`
	Abgebrochen bool            `json:"abgebrochen"`
	Fehler      *Fehler         `json:"fehler"`
}

// Setvorschauleser liest eine Datei, wie das Einfahren sie läse.
type Setvorschauleser interface {
	Vorschau(ctx context.Context, pfad string) (*service.Setvorschau, error)
}

// Kopfbedarfsquelle nennt, was ein Setkopf an Paketen braucht.
type Kopfbedarfsquelle interface {
	Kopfbedarf(ctx context.Context, kopf domain.SetHead) ([]service.Abhaengigkeitsbefund, error)
}

// SetVorschau sagt, was das Einfahren der Datei `pfad` täte.
func (a *App) SetVorschau(ctx context.Context, pfad string) SetvorschauAntwort {
	ctx = a.imProfil(ctx)

	if pfad == "" {
		return SetvorschauAntwort{Abgebrochen: true}
	}

	vorschau, err := setvorschau(ctx, a.importeur, a.pakete, pfad)
	if err != nil {
		return SetvorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SetvorschauAntwort{Vorschau: vorschau}
}

// Setvorschau ist dieselbe Vorschau für die Flächen ohne Fenster — über dem
// Bestand, in den das Einfahren schriebe.
func (w *Paketweg) Setvorschau(ctx context.Context, db *sql.DB, pfad string) (*SetvorschauDTO, error) {
	return setvorschau(ctx, Einfuhrnaht(db), w.verwalter, pfad)
}

// setvorschau liest die Datei und fragt nach dem Bedarf ihres Kopfs.
func setvorschau(
	ctx context.Context, leser Setvorschauleser, pakete Kopfbedarfsquelle, pfad string,
) (*SetvorschauDTO, error) {
	gelesen, err := leser.Vorschau(ctx, pfad)
	if err != nil {
		return nil, err
	}

	bedarf, err := pakete.Kopfbedarf(ctx, domain.SetHead{
		Name: gelesen.Name, Game: gelesen.Spiel, Style: gelesen.Stil,
	})
	if err != nil {
		return nil, err
	}

	dto := &SetvorschauDTO{
		Pfad: pfad, Name: gelesen.Name, Spiel: gelesen.Spiel, Stil: gelesen.Stil,
		Karten: gelesen.Karten, Befunde: []string{}, Bedarf: alsAbhaengigkeiten(bedarf),
	}

	for _, befund := range gelesen.Befunde {
		dto.Befunde = append(dto.Befunde, befund.String())
	}

	if gelesen.SchonOffen != 0 {
		dto.SchonOffen = strconv.FormatInt(gelesen.SchonOffen, 10)
	}

	return dto, nil
}
