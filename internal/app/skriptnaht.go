// Die Skriptkonsole für die Flächen ohne Fenster
// (#1070).

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KonsolenDTO ist, was eine Fläche ohne Fenster von der Konsole zurückbekommt.
//
// **Ein eigener Verbund und nicht [service.Konsolenergebnis]**: `cmd/mse`,
// `cmd/mse-mcp` und `cmd/mse-rest` kennen `service` nicht (§2).
type KonsolenDTO struct {
	// Quelltext ist das Ergebnis als Quelltext — `"a"` und nicht `a`.
	Quelltext string `json:"quelltext"`

	// Text ist dasselbe in gewöhnlicher Form.
	Text string `json:"text"`

	// Art sagt, was für ein Wert es ist.
	Art string `json:"art"`

	// Befunde sind die Vorbehalte beim Aufbau der Umgebung.
	Befunde []string `json:"befunde"`

	// Eintraege sind die Einträge einer Abbildung, rekursiv — leer für jeden
	// anderen Wert (Werkstück U13 von Phase 5.1, #1424).
	Eintraege []KonsolenEintrag `json:"eintraege"`
}

// KonsolenEintrag ist ein Eintrag einer Abbildung im Ergebnis — mit seinen
// eigenen Einträgen, wenn er selbst eine ist. Das Fenster zeigt den Baum.
type KonsolenEintrag struct {
	Name      string            `json:"name"`
	Quelltext string            `json:"quelltext"`
	Art       string            `json:"art"`
	Eintraege []KonsolenEintrag `json:"eintraege"`
}

// eintraegeAls übersetzt die Einträge des Services — **nie nil**, damit über
// die Wails-Bindung `[]` steht und nicht `null`.
func eintraegeAls(eintraege []service.Konsoleneintrag) []KonsolenEintrag {
	uebersetzt := make([]KonsolenEintrag, 0, len(eintraege))
	for _, eintrag := range eintraege {
		uebersetzt = append(uebersetzt, KonsolenEintrag{
			Name: eintrag.Name, Quelltext: eintrag.Quelltext, Art: eintrag.Art,
			Eintraege: eintraegeAls(eintrag.Eintraege),
		})
	}

	return uebersetzt
}

// Skriptnaht ist die Konsole, wie eine Fläche ohne Fenster sie ruft.
type Skriptnaht struct{ konsole *service.Skriptkonsole }

// SkriptnahtUeber baut sie über einem Bestand.
//
// **Mit dem vollen Zeichenwerk**, weil die Umgebung eines Ausdrucks dieselbe
// sein soll wie die, in der eine Karte gezeichnet wird. Eine Konsole mit einer
// ärmeren Umgebung beantwortete Fragen über ein Programm, das es nicht gibt.
func SkriptnahtUeber(db *sql.DB) (*Skriptnaht, error) {
	werk, err := Zeichennaht(db)
	if err != nil {
		return nil, err
	}

	return &Skriptnaht{
		konsole: service.NeueSkriptkonsole(storage.NeuerSetLeser(db), werk.Planbauer),
	}, nil
}

// Auswerten rechnet einen Ausdruck im Kontext einer Karte aus.
func (s *Skriptnaht) Auswerten(ctx context.Context, karteID int64, ausdruck string) (
	KonsolenDTO, error,
) {
	ergebnis, fehler := s.konsole.Auswerten(ctx, karteID, ausdruck)
	if fehler != nil {
		return KonsolenDTO{}, fehler
	}

	return KonsolenDTO{
		Quelltext: ergebnis.Quelltext,
		Text:      ergebnis.Text,
		Art:       ergebnis.Art,
		Befunde:   befundtexte(ergebnis.Befunde),
		Eintraege: eintraegeAls(ergebnis.Eintraege),
	}, nil
}

// Namen gibt, was ein Ausdruck an dieser Karte nennen darf
// (#1164).
//
// **Dieselbe Umgebung wie beim Auswerten** — sie kommt aus demselben
// Planbauer. Eine zweite Quelle für die Namen liefe an dem Tag auseinander, an
// dem eine Vorlage etwas dazulegt, und die Vervollständigung böte dann etwas
// an, das nicht läuft.
func (s *Skriptnaht) Namen(ctx context.Context, karteID int64) ([]string, error) {
	return s.konsole.Namen(ctx, karteID)
}
