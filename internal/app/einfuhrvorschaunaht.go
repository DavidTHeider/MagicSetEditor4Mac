package app

// **Die Sammelvorschau an der Naht** — DOM-03.AK-01
// (#1698).
//
// **Eine eigene Datei neben `listeneinfuhrnaht.go`**: dort steht, wie
// gedeutet und eingelesen wird, hier, was eine Einfuhr **vorher** sagt — und
// die Wahl, was mit einer vorhandenen Zeile geschieht. Die Regel dahinter
// steht in `service/einfuhrvorschau.go`.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// VorschaupostenDTO ist eine Zeile der Vorschau.
type VorschaupostenDTO struct {
	// Zeile ist die Stelle in der Eingabe — beim CSV zählt die Kopfzeile mit.
	Zeile       int    `json:"zeile"`
	Bezeichnung string `json:"bezeichnung"`
	Menge       int    `json:"menge"`
	Zustand     string `json:"zustand"`
	Sprache     string `json:"sprache"`
	Wunsch      bool   `json:"wunsch"`

	// Art ist `neu`, `zusammen` oder `wiederholt`.
	Art string `json:"art"`

	// Vorher und Nachher sind die Mengen der getroffenen Zeile beim Addieren.
	Vorher  int `json:"vorher"`
	Nachher int `json:"nachher"`
}

// Sammelvorschau sagt, was eine Einfuhr täte.
type Sammelvorschau struct {
	Posten  []VorschaupostenDTO `json:"posten"`
	Befunde []string            `json:"befunde"`
}

// Sammelvorschauantwort trägt eine Vorschau oder einen Fehler.
type Sammelvorschauantwort struct {
	Vorschau *Sammelvorschau `json:"vorschau"`
	Fehler   *Fehler         `json:"fehler"`
}

// alsSammelvorschau bildet die Vorschau des Dienstes ab.
func alsSammelvorschau(vorschau service.Sammelvorschau) *Sammelvorschau {
	dto := &Sammelvorschau{
		Posten:  make([]VorschaupostenDTO, 0, len(vorschau.Posten)),
		Befunde: befundtexte(vorschau.Befunde),
	}

	for _, posten := range vorschau.Posten {
		dto.Posten = append(dto.Posten, VorschaupostenDTO{
			Zeile: posten.Zeile, Bezeichnung: posten.Exemplar.Bezeichnung(),
			Menge: posten.Exemplar.Menge, Zustand: string(posten.Exemplar.Zustand),
			Sprache: posten.Exemplar.Sprache, Wunsch: posten.Exemplar.Wunsch,
			Art: string(posten.Art), Vorher: posten.Vorher, Nachher: posten.Nachher,
		})
	}

	return dto
}

// Vorschauen sagt, was das Einlesen der entschiedenen Zeilen täte.
func (l *Listeneinfuhrnaht) Vorschauen(
	ctx context.Context, zeilen []Listeneinfuhrzeile,
) (*Sammelvorschau, error) {
	vorschau, err := l.einfuhr.Vorschauen(ctx, alsListenzeilen(zeilen))
	if err != nil {
		return nil, err
	}

	return alsSammelvorschau(vorschau), nil
}

// CSVVorschauen sagt, was das Einlesen eines CSV täte.
func (s *Sammlungsausfuhrnaht) CSVVorschauen(ctx context.Context, text string) (
	*Sammelvorschau, error,
) {
	vorschau, err := s.ausfuhr.Vorschauen(ctx, text)
	if err != nil {
		return nil, err
	}

	return alsSammelvorschau(vorschau), nil
}

// alsListenzeilen übersetzt die Zeilen der Naht in die des Dienstes —
// **eine** Stelle für Einlesen und Vorschau.
func alsListenzeilen(zeilen []Listeneinfuhrzeile) []service.Listenzeile {
	uebersetzt := make([]service.Listenzeile, 0, len(zeilen))

	for _, zeile := range zeilen {
		uebersetzt = append(uebersetzt, service.Listenzeile{
			Printing: zeile.Printing,
			Menge:    zeile.Menge,
			Zustand:  domain.Zustand(zeile.Zustand),
			Sprache:  zeile.Sprache,
			Finish:   domain.Finish(zeile.Finish),
			Name:     zeile.Name,
			Setcode:  zeile.Setcode,
			Nummer:   zeile.Nummer,
		})
	}

	return uebersetzt
}
