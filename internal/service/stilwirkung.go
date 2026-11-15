package service

// **Eine Wahrheit für die setweiten Stiloptionen** — F05
// (#2247),
// gefunden beim Beleg für Befund 1 des Nachlaufs über Phase G, zweiter Teil.
//
// # Was schiefging
//
// Die Rasterung las den Block `styling:` der Datei, die Optionsliste die hier
// gesetzten Werte (`styling_values`) — und keine die andere. Eine hier
// gesetzte Option änderte die Karte nicht, das Bild blieb Byte für Byte
// gleich; ein Set aus einer Datei zeichnete »dunkel« und nannte »hell«.
//
// # Die eine Reihenfolge
//
// **Vorgabe des Stils, darüber der Block der Datei, darüber, was hier gesetzt
// wurde, darüber die Karte.** MSE2 hat nur einen Ort, `Set::stylingDataFor`
// (`src/data/set.cpp`, Zeilen 99–105, Fassung f9d9356 vom 2021-07-06); hier
// sind es zwei, weil ein Set aus einer Datei seinen Block mitbringt
// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md))
// und eine Option hier einzeln gesetzt wird. Was hier gesetzt wurde, ist das
// jüngere und gilt; beim Sichern kommt es in den Block ([stilblockAngleichen]).

import (
	"context"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Stilwertquelle gibt die Stiloptionen, die in dieser Anwendung gesetzt
// wurden — implementiert in `storage` ([Stilablage] liest dieselbe Tabelle).
type Stilwertquelle interface {
	Stilwerte(ctx context.Context, setID int64) ([]domain.Value, error)
}

// MitStilwerten legt die hier gesetzten Stiloptionen über den Block der Datei.
//
// **Ohne sie gilt nur der Block** — wie bisher, und für einen Speicher, der
// eine ungespeicherte Vorlage zeigt, ist das richtig.
func (s *Stilspeicher) MitStilwerten(quelle Stilwertquelle) *Stilspeicher {
	s.stilwerte = quelle

	return s
}

// gesetzteStilwerte trägt die hier gesetzten Stiloptionen in die Vorlage ein.
//
// **Jedes Mal gelesen, nie gemerkt** — dieselbe Regel wie für die
// Setangaben: sie ändern sich im selben Fenster. **Ein Fehler ist ein Befund**
// und hält die Karte nicht auf (§8).
func (s *Stilspeicher) gesetzteStilwerte(
	ctx context.Context, setID int64, vorlage *Vorlage, befunde []domain.Befund,
) []domain.Befund {
	if s.stilwerte == nil {
		return befunde
	}

	werte, err := s.stilwerte.Stilwerte(ctx, setID)
	if err != nil {
		return append(befunde, befundliste("die Stiloptionen lesen: %v", err)...)
	}

	vorlage.Stilwerte = werte

	return befunde
}

// stilwerteNachName gibt Stiloptionen unter dem Namen, unter dem ein Skript
// sie liest — `rahmen zeigen` als `rahmen_zeigen`.
func stilwerteNachName(werte []domain.Value) map[string]string {
	nachName := make(map[string]string, len(werte))

	for _, wert := range werte {
		nachName[strings.ReplaceAll(wert.Field, " ", "_")] = wert.Text
	}

	return nachName
}

// alsStilwerte macht aus einer Abbildung wieder Werte — **in fester
// Reihenfolge**, denn wer sie überlagert, hängt Unbekannte hinten an.
func alsStilwerte(nachName map[string]string) []domain.Value {
	namen := make([]string, 0, len(nachName))
	for name := range nachName {
		namen = append(namen, name)
	}

	sort.Strings(namen)

	werte := make([]domain.Value, 0, len(namen))
	for _, name := range namen {
		werte = append(werte, domain.Value{Field: name, Text: nachName[name]})
	}

	return werte
}
