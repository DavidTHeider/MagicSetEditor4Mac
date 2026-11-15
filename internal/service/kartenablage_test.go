package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Karten kopieren und einfügen** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # ported, und die Fundstelle ist genau
//
//	wxDataFormat CardsDataObject::format = _("application/x-mse-cards");
//
//	struct WrappedCards {
//	  Game*         expected_game;
//	  String        game_name;
//	  vector<CardP> cards;
//	};
//
//	IMPLEMENT_REFLECTION(WrappedCards) {
//	  REFLECT(game_name);
//	  if (game_name == expected_game->name()) {
//	    WITH_DYNAMIC_ARG(game_for_reading, expected_game);
//	    REFLECT(cards);
//	  }
//	}
//
// `src/data/format/clipboard.cpp`, Zeilen 42–60, Fassung f9d9356 vom
// 2021-07-06. **Drei Dinge stehen darin, und alle drei gelten hier**: das
// Spiel steht **vor** den Karten, die Karten sind MSEs eigenes Textformat, und
// **stimmt das Spiel nicht, werden die Karten gar nicht erst gelesen.**

// ablagestand baut Ablage, Bestand und Anleger.
type ablagedoppel struct {
	karten   map[int64][]*domain.CardRow
	koepfe   []domain.SetHead
	angelegt []int64
	werte    map[int64][]domain.Value
	fehler   error

	// schreibfehler scheitert **nur** beim Setzen der Werte — damit der Lauf
	// bis dorthin kommt, wo die Karte schon entstanden ist.
	schreibfehler error

	// anlegefehler scheitert **nur** beim Anlegen.
	anlegefehler error

	// abSchreibzug lässt das Schreiben erst ab dem n-ten Zug scheitern.
	//
	// **Für den Teilerfolg** — ohne ihn scheitert immer schon die erste Karte,
	// und der Fall „die erste ist da, die zweite nicht" kommt in keinem Lauf
	// vor. Genau der stand hinter Befund 2 des Nachlaufs
	// (#1282).
	// Null heißt: keine Schranke.
	abSchreibzug int

	// schreibzuege zählt die Schreibversuche.
	schreibzuege int

	// kartenfehler scheitert **nur** beim Lesen der Karten — die Sets kommen
	// dann noch durch, und der Lauf erreicht die Stelle dahinter.
	kartenfehler error
}

func (a *ablagedoppel) Karten(_ context.Context, setID int64) ([]*domain.CardRow, error) {
	if a.kartenfehler != nil {
		return nil, a.kartenfehler
	}

	return a.karten[setID], a.fehler
}

func (a *ablagedoppel) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return a.koepfe, a.fehler
}

func (a *ablagedoppel) KarteAnlegen(_ context.Context, _ int64) (int64, error) {
	if a.fehler != nil {
		return 0, a.fehler
	}

	if a.anlegefehler != nil {
		return 0, a.anlegefehler
	}

	neue := int64(100 + len(a.angelegt))
	a.angelegt = append(a.angelegt, neue)

	return neue, nil
}

func (a *ablagedoppel) FelderVonAussen(
	_ context.Context, _, karteID int64, werte []domain.Value, _ string,
) (kennung string, geaendert bool, fehler error) {
	if a.fehler != nil {
		return "", false, a.fehler
	}

	a.schreibzuege++

	if a.schreibfehler != nil && a.schreibzuege >= max(a.abSchreibzug, 1) {
		return "", false, a.schreibfehler
	}

	if a.werte == nil {
		a.werte = map[int64][]domain.Value{}
	}

	a.werte[karteID] = werte

	return "", false, nil
}

func kartenablage(doppel *ablagedoppel) *service.Kartenablage {
	return service.NeueKartenablage(doppel, doppel, doppel)
}

func zweiSets() *ablagedoppel {
	return &ablagedoppel{
		koepfe: []domain.SetHead{
			{ID: 1, Name: "Erstes", Game: "magic"},
			{ID: 2, Name: "Zweites", Game: "magic"},
			{ID: 3, Name: "Fremdes", Game: "vanguard"},
		},
		karten: map[int64][]*domain.CardRow{
			1: {
				{ID: 11, SetID: 1, Values: []domain.Value{
					{Field: "name", Text: "Erste"},
					{Field: "rule text", Text: "Fliegend"},
				}},
				{ID: 12, SetID: 1, Values: []domain.Value{
					{Field: "name", Text: "Zweite"},
				}},
			},
		},
	}
}

// **Kopiert wird MSEs eigenes Textformat, mit dem Spiel voran.**
func TestKopierenGibtDasFormatVonMSE(t *testing.T) {
	t.Parallel()

	inhalt, err := kartenablage(zweiSets()).Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	if !strings.HasPrefix(inhalt, "game_name: magic\n") {
		t.Errorf("der Inhalt beginnt mit %q", inhalt)
	}

	if !strings.Contains(inhalt, "card:\n\tname: Erste\n\trule text: Fliegend\n") {
		t.Errorf("Inhalt:\n%s", inhalt)
	}
}

// **Eine Karte, die nicht zu diesem Set gehört, wird nicht kopiert.**
func TestKopierenNimmtNurKartenDesSets(t *testing.T) {
	t.Parallel()

	_, err := kartenablage(zweiSets()).Kopieren(t.Context(), 1, []int64{99})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler war %v — will ErrNotFound", err)
	}
}

// **Ohne Karten gibt es nichts zu kopieren.**
func TestKopierenOhneKartenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, err := kartenablage(zweiSets()).Kopieren(t.Context(), 1, nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", err)
	}
}

// **Eingefügt wird in ein anderes Set desselben Spiels.**
func TestEinfuegenLegtDieKartenImZielsetAn(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	ablagen := kartenablage(doppel)

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11, 12})
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	ergebnis, err := ablagen.Einfuegen(t.Context(), 2, inhalt)
	if err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	if len(ergebnis.Karten) != 2 {
		t.Fatalf("es entstanden %d Karten", len(ergebnis.Karten))
	}

	// **Die Werte reisen mit** — sonst wäre es ein Anlegen und kein Einfügen.
	erste := doppel.werte[ergebnis.Karten[0]]
	if len(erste) != 2 || erste[0].Text != "Erste" {
		t.Errorf("die erste Karte trägt %v", erste)
	}
}

// **Ein anderes Spiel wird abgewiesen — und die Karten gar nicht erst
// gelesen.**
//
// Das ist MSEs `if (game_name == expected_game->name())`: die Bedingung steht
// **um** das Lesen herum, nicht dahinter. Eine Karte eines anderen Spiels hat
// Felder, die es hier nicht gibt.
func TestEinfuegenWeistEinFremdesSpielAb(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	ablagen := kartenablage(doppel)

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	if _, fehler := ablagen.Einfuegen(t.Context(), 3, inhalt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("Fehler war %v — will ErrValidation", fehler)
	}

	if len(doppel.angelegt) != 0 {
		t.Errorf("es wurden %d Karten angelegt", len(doppel.angelegt))
	}
}

// **Was nicht von hier stammt, wird abgewiesen.**
//
// Die Zwischenablage trägt, was jemand zuletzt kopiert hat — ein Stück Text
// aus einem Browser genauso wie eine Karte. Ohne diese Wache entstünde daraus
// eine Karte ohne Felder.
func TestEinfuegenWeistFremdenTextAb(t *testing.T) {
	t.Parallel()

	for name, inhalt := range map[string]string{
		"leer":            "",
		"nur Leerzeichen": "   \n",
		"fremder Text":    "Hallo, ich bin ein Einkaufszettel\n",
		"ohne Spiel":      "card:\n\tname: Erste\n",
		"ohne Karten":     "game_name: magic\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := zweiSets()

			if _, fehler := kartenablage(doppel).Einfuegen(
				t.Context(), 2, inhalt); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("Fehler war %v — will ErrValidation", fehler)
			}

			if len(doppel.angelegt) != 0 {
				t.Errorf("es wurden %d Karten angelegt", len(doppel.angelegt))
			}
		})
	}
}

// **Ein mehrzeiliger Wert überlebt den Rundlauf.**
//
// # Warum das eine eigene Zusage ist
//
// Ein Wert mit Umbrüchen steht im Paketformat **unter** seinem Schlüssel, nicht
// dahinter — und der Regeltext einer gewöhnlichen Karte ist einer
// (#398). Wer
// beim Einfügen nur `Wert` läse, bekäme ein leeres Feld: **ohne Befund**, weil
// ein Textfeld ohne Wert gar nicht gezeichnet wird.
func TestEinMehrzeiligerWertUeberlebtDenRundlauf(t *testing.T) {
	t.Parallel()

	doppel := zweiSets()
	doppel.karten[1] = []*domain.CardRow{{ID: 11, SetID: 1, Values: []domain.Value{
		{Field: "name", Text: "Erste"},
		{Field: "rule text", Text: "Fliegend\nWachsamkeit"},
	}}}

	ablagen := kartenablage(doppel)

	inhalt, err := ablagen.Kopieren(t.Context(), 1, []int64{11})
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	ergebnis, err := ablagen.Einfuegen(t.Context(), 2, inhalt)
	if err != nil {
		t.Fatalf("einfügen: %v", err)
	}

	nach := map[string]string{}
	for _, wert := range doppel.werte[ergebnis.Karten[0]] {
		nach[wert.Field] = wert.Text
	}

	if nach["rule text"] != "Fliegend\nWachsamkeit" {
		t.Errorf("rule text = %q", nach["rule text"])
	}
}
