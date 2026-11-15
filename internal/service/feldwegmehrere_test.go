package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Mehrere übernommene Felder gehen in einer Reihe hinaus** —
// #1026.
//
// Eine Reihe ist **eine** Transaktion (#714). Zwei Aufrufe wären zwei, und
// scheiterte der zweite, stünde das erste Feld neu da und das zweite alt.
func TestMehrereFelderGehenInEinerReihe(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{}}, wegfelder{felder: m15felder()}, ablage)

	if _, _, fehler := weg.FelderVonAussen(t.Context(), 1, []domain.Value{
		{Field: "name", Text: "Prüfstein"},
		{Field: "rule text", Text: "Fliegend."},
	}, "https://beispiel.invalid/1"); fehler != nil {
		t.Fatalf("FelderVonAussen: %v", fehler)
	}

	if ablage.reihen != 1 {
		t.Errorf("%d Reihen, erwartet war eine", ablage.reihen)
	}

	if got := strings.Join(ablage.geschrieben, " "); got != "name=Prüfstein rule text=Fliegend." {
		t.Errorf("geschrieben %q", got)
	}

	if len(ablage.herkuenfte) != 1 || ablage.herkuenfte[0] != "https://beispiel.invalid/1" {
		t.Errorf("Herkünfte %v", ablage.herkuenfte)
	}
}

// **Ein Verbundfeld zerfällt auch hier in seine Teile** — und in derselben
// Reihe wie die übrigen.
//
// # Warum das der eigentliche Punkt ist
//
// Die Auflösung steht an **einer** Stelle (`aufloesen`). Hätte dieser Weg
// seine eigene, liefe sie irgendwann auseinander: `text` würde einzeln
// geschrieben zerlegt und in einem Zug roh abgelegt — und niemand sähe es,
// weil beide Wege grün wären.
func TestEinVerbundfeldZerfaelltAuchInDerReihe(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{}}, wegfelder{felder: m15felder()}, ablage)

	if _, _, fehler := weg.FelderVonAussen(t.Context(), 1, []domain.Value{
		{Field: "name", Text: "Prüfstein"},
		{Field: "text", Text: "Fliegend.<sep><line>\n</line></sep>Ein Spruch."},
	}, "quelle"); fehler != nil {
		t.Fatalf("FelderVonAussen: %v", fehler)
	}

	if ablage.reihen != 1 {
		t.Errorf("%d Reihen, erwartet war eine", ablage.reihen)
	}

	got := strings.Join(ablage.geschrieben, " | ")
	if got != "name=Prüfstein | rule text=Fliegend. | flavor text=Ein Spruch." {
		t.Errorf("geschrieben %q", got)
	}
}

// **Ein nicht speicherbares Feld hält die ganze Reihe an.**
//
// Und zwar **bevor** irgendetwas geschrieben ist: ein Wert, der nicht abgelegt
// werden kann, ist ein Befund über die Eingabe, nicht über die Hälfte davon.
func TestEinUnspeicherbaresFeldHaeltDieReiheAn(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{}}, wegfelder{felder: m15felder()}, ablage)

	_, _, fehler := weg.FelderVonAussen(t.Context(), 1, []domain.Value{
		{Field: "name", Text: "Prüfstein"},
		{Field: "card code", Text: "geht nicht"},
	}, "quelle")
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v — kein Validierungsbefund", fehler)
	}

	if len(ablage.geschrieben) != 0 {
		t.Errorf("es wurde geschrieben: %v", ablage.geschrieben)
	}
}

// **Ohne lesbare Vorlage gehen die Werte roh hinaus** — dieselbe Duldung wie
// beim einzelnen Schreiben, und aus demselben Grund: ein Set aus MTG JSON
// bliebe sonst unbearbeitbar, solange sein Spielpaket fehlt.
func TestOhneVorlageGehenDieWerteRohHinaus(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{}},
		wegfelder{fehler: errors.New("kein Paket")}, ablage)

	if _, _, fehler := weg.FelderVonAussen(t.Context(), 1, []domain.Value{
		{Field: "text", Text: "roh"},
	}, "quelle"); fehler != nil {
		t.Fatalf("FelderVonAussen: %v", fehler)
	}

	if got := strings.Join(ablage.geschrieben, " "); got != "text=roh" {
		t.Errorf("geschrieben %q", got)
	}
}
