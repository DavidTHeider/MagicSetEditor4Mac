// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ersetzdoppel ist Kartenquelle, Feldquelle und Schreiber in einem.
type ersetzdoppel struct {
	karten        []*domain.CardRow
	felder        []domain.Field
	lesefehler    error
	feldfehler    error
	schreibfehler error

	geschrieben []domain.Feldwechsel
}

func (e *ersetzdoppel) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return e.karten, e.lesefehler
}

func (e *ersetzdoppel) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return e.felder, e.feldfehler
}

func (e *ersetzdoppel) WerteSetzen(_ context.Context, wechsel []domain.Feldwechsel) error {
	if e.schreibfehler != nil {
		return e.schreibfehler
	}

	e.geschrieben = append(e.geschrieben, wechsel...)

	return nil
}

// dreiKarten gibt ein Doppel mit drei Karten und den üblichen Feldarten.
func dreiKarten() *ersetzdoppel {
	return &ersetzdoppel{
		karten: []*domain.CardRow{
			{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Fliegender Drache"},
				{Field: "rule text", Text: "<i>Fliegend</i>. Sehr schnell."},
				{Field: "rarity", Text: "common"},
			}},
			{ID: 2, Number: 1, Values: []domain.Value{
				{Field: "name", Text: "Eisball"},
				{Field: "rule text", Text: "Kalt."},
			}},
			{ID: 3, Number: 2, Values: []domain.Value{
				{Field: "name", Text: "Fliegende Kuh"},
			}},
		},
		felder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "rule text", Kind: domain.KindText},
			{Name: "rarity", Kind: domain.KindChoice, Choices: []string{"common", "rare"}},
		},
	}
}

func ersetzer(doppel *ersetzdoppel) *service.Ersetzer {
	return service.NeuerErsetzer(doppel, doppel, doppel)
}

// **Die Vorschau nennt jede betroffene Stelle — und schreibt nichts.**
func TestDieVorschauNenntJedeStelleUndSchreibtNichts(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()

	schau, fehler := ersetzer(doppel).Vorschau(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	// Karte 1 zweimal (`name` und `rule text`), Karte 3 einmal.
	if len(schau.Stellen) != 3 || schau.Karten != 2 || schau.Treffer != 3 {
		t.Fatalf("%d Stellen, %d Karten, %d Treffer",
			len(schau.Stellen), schau.Karten, schau.Treffer)
	}

	if len(doppel.geschrieben) != 0 {
		t.Errorf("die Vorschau hat geschrieben: %+v", doppel.geschrieben)
	}

	erste := schau.Stellen[0]
	if erste.KarteID != 1 || erste.Feld != "name" ||
		erste.Vorher != "Fliegender Drache" || erste.Nachher != "Rennender Drache" {
		t.Errorf("erste Stelle: %+v", erste)
	}

	// **Die Auszeichnung steht nicht in der Vorschau** — sie ist Werkzeug und
	// kein Text; wer sie sieht, hält sie für einen Fehler.
	zweite := schau.Stellen[1]
	if zweite.Feld != "rule text" || zweite.Vorher != "Fliegend. Sehr schnell." ||
		zweite.Nachher != "Rennend. Sehr schnell." {
		t.Errorf("zweite Stelle: %+v", zweite)
	}
}

// **Nur Textfelder werden durchsucht.**
//
//	virtual bool search(FindInfo& find, bool from_start) { return false; }
//
// `src/gui/value/editor.hpp`, Zeile 115, Fassung f9d9356 vom 2021-07-06: die
// Grundfassung findet nichts, und nur `TextValueEditor` überschreibt sie
// (`src/gui/value/text.hpp`, Zeile 85). Eine Ersetzung in einem `choice:`-Feld
// schriebe eine Wahl, die es nicht gibt — und die Karte zeichnete nichts mehr.
func TestNurTextfelderWerdenDurchsucht(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()

	schau, fehler := ersetzer(doppel).Vorschau(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "common", Ersatz: "rare"})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if len(schau.Stellen) != 0 {
		t.Errorf("das Wahlfeld steht in der Vorschau: %+v", schau.Stellen)
	}
}

// **Ein unbekanntes Feld gilt als Text** — dieselbe Duldung wie überall.
//
// `KindUnknown` heißt „diese Vorlage nennt eine Art, die wir nicht
// unterscheiden"; sie als undurchsuchbar zu behandeln nähme einem fremden Set
// die Ersetzung, obwohl der Wert Text ist.
func TestEinUnbekanntesFeldGiltAlsText(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()
	doppel.felder = []domain.Field{{Name: "name", Kind: domain.KindUnknown}}

	schau, fehler := ersetzer(doppel).Vorschau(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if len(schau.Stellen) != 2 {
		t.Errorf("%d Stellen, erwartet zwei", len(schau.Stellen))
	}
}

// **Die Frage kann sich auf ein Feld beschränken.**
func TestDieFrageKannSichAufEinFeldBeschraenken(t *testing.T) {
	t.Parallel()

	schau, fehler := ersetzer(dreiKarten()).Vorschau(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn", Feld: "name"})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if len(schau.Stellen) != 2 {
		t.Fatalf("%d Stellen, erwartet zwei", len(schau.Stellen))
	}

	for _, stelle := range schau.Stellen {
		if stelle.Feld != "name" {
			t.Errorf("Stelle in %q", stelle.Feld)
		}
	}
}

// **Ausführen schreibt genau die Stellen der Vorschau — in einem Zug.**
func TestAusfuehrenSchreibtDieStellenDerVorschau(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()

	schau, fehler := ersetzer(doppel).Ausfuehren(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"})
	if fehler != nil {
		t.Fatalf("Ausführen: %v", fehler)
	}

	if schau.Treffer != 3 || len(doppel.geschrieben) != 3 {
		t.Fatalf("%d Treffer, %d geschrieben", schau.Treffer, len(doppel.geschrieben))
	}

	// **Der geschriebene Wert trägt seine Auszeichnung weiter.**
	if doppel.geschrieben[1].Wert != "<i>Rennend</i>. Sehr schnell." {
		t.Errorf("geschrieben wurde %q", doppel.geschrieben[1].Wert)
	}
}

// **Nichts gefunden schreibt nichts** — und ist kein Fehler.
func TestOhneTrefferWirdNichtsGeschrieben(t *testing.T) {
	t.Parallel()

	doppel := dreiKarten()

	schau, fehler := ersetzer(doppel).Ausfuehren(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Schwimm", Ersatz: "Renn"})
	if fehler != nil {
		t.Fatalf("Ausführen: %v", fehler)
	}

	if schau.Treffer != 0 || len(doppel.geschrieben) != 0 {
		t.Errorf("%d Treffer, %d geschrieben", schau.Treffer, len(doppel.geschrieben))
	}
}

// **Eine Frage ohne Suchtext wird abgewiesen — auf beiden Wegen.**
func TestEineLeereFrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gestellt := domain.Ersetzungsfrage{Text: "  ", Ersatz: "X"}

	if _, fehler := ersetzer(dreiKarten()).Vorschau(t.Context(), 1, gestellt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Vorschau: %v", fehler)
	}

	if _, fehler := ersetzer(dreiKarten()).Ausfuehren(t.Context(), 1, gestellt); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("Ausführen: %v", fehler)
	}
}

// **Ein Lesefehler kommt durch, und zwar an jedem der beiden Wege.**
func TestEinLesefehlerDerErsetzungKommtDurch(t *testing.T) {
	t.Parallel()

	gestellt := domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"}

	for _, fall := range []struct {
		name string
		baue func() *ersetzdoppel
	}{
		{"Karten unlesbar", func() *ersetzdoppel {
			doppel := dreiKarten()
			doppel.lesefehler = domain.ErrBelegt

			return doppel
		}},
		{"Schreiben scheitert", func() *ersetzdoppel {
			doppel := dreiKarten()
			doppel.schreibfehler = domain.ErrBelegt

			return doppel
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			doppel := fall.baue()

			// Die Vorschau liest, schreibt aber nicht — ein Schreibfehler
			// erreicht sie deshalb gar nicht.
			if fall.name != "Schreiben scheitert" {
				if _, fehler := ersetzer(doppel).Vorschau(
					t.Context(), 1, gestellt); !errors.Is(fehler, domain.ErrBelegt) {
					t.Errorf("Vorschau: %v", fehler)
				}
			}

			if _, fehler := ersetzer(doppel).Ausfuehren(
				t.Context(), 1, gestellt); !errors.Is(fehler, domain.ErrBelegt) {
				t.Errorf("Ausführen: %v", fehler)
			}
		})
	}
}

// **Ein Set ohne lesbare Vorlage lässt sich trotzdem durchsuchen — und sagt
// es.**
//
// Dieselbe Duldung wie beim Feldschreiben (`Feldweg.FelderVonAussen`): ein Set
// aus MTG JSON bliebe sonst unbearbeitbar, solange sein Spielpaket fehlt.
// **Dann gilt jedes Feld als Text** — und das ist die einzige Annahme, die man
// ohne Vorlage treffen kann.
//
// **Beide Gestalten**: eine leere Liste und ein Lesefehler. Bis zum 2026-08-27
// galt die Duldung nur der ersten — und ein fehlendes Spielpaket gibt die
// zweite. Der Satz war geschrieben und erreichte den Fall nicht, für den er
// da war.
func TestOhneVorlageGiltJedesFeldAlsText(t *testing.T) {
	t.Parallel()

	for name, verletze := range map[string]func(*ersetzdoppel){
		"keine Felder":    func(d *ersetzdoppel) { d.felder = nil },
		"Felder unlesbar": func(d *ersetzdoppel) { d.feldfehler = domain.ErrNotFound },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doppel := dreiKarten()
			verletze(doppel)

			schau, fehler := ersetzer(doppel).Vorschau(t.Context(), 1,
				domain.Ersetzungsfrage{Text: "common", Ersatz: "rare"})
			if fehler != nil {
				t.Fatalf("Vorschau: %v", fehler)
			}

			// **Das Wahlfeld ist jetzt dabei** — ohne Vorlage weiß niemand,
			// dass es eines ist.
			if len(schau.Stellen) != 1 {
				t.Fatalf("%d Stellen, erwartet eine", len(schau.Stellen))
			}

			// **Und die Antwort sagt, dass geraten wurde.**
			if !schau.OhneVorlage {
				t.Error("die Antwort verschweigt die fehlende Vorlage")
			}
		})
	}
}

// **Mit Vorlage steht der Hinweis nicht da.**
//
// Die Gegenprobe zum Test darüber: ein Hinweis, der immer dasteht, sagt nichts.
func TestMitVorlageStehtKeinHinweis(t *testing.T) {
	t.Parallel()

	schau, fehler := ersetzer(dreiKarten()).Vorschau(t.Context(), 1,
		domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if schau.OhneVorlage {
		t.Error("der Hinweis steht auch mit Vorlage da")
	}
}

// **Die Vorschau hat eine Obergrenze, und sie sagt es.**
//
// Ein Set mit 3.788 Karten und einem häufigen Wort gäbe eine Antwort, die
// niemand liest und die durch die Naht muss. **Ausgeführt wird trotzdem
// alles** — die Grenze gilt der Anzeige, nicht der Änderung.
func TestDieVorschauHatEineObergrenze(t *testing.T) {
	t.Parallel()

	doppel := &ersetzdoppel{felder: []domain.Field{{Name: "name", Kind: domain.KindText}}}

	for i := range domain.GrenzeErsetzungsvorschau + 5 {
		doppel.karten = append(doppel.karten, &domain.CardRow{
			ID: int64(i + 1), Number: i,
			Values: []domain.Value{{Field: "name", Text: "Fliegend"}},
		})
	}

	gestellt := domain.Ersetzungsfrage{Text: "Flieg", Ersatz: "Renn"}

	schau, fehler := ersetzer(doppel).Vorschau(t.Context(), 1, gestellt)
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if len(schau.Stellen) != domain.GrenzeErsetzungsvorschau || !schau.Weitere {
		t.Errorf("%d Stellen, weitere=%v", len(schau.Stellen), schau.Weitere)
	}

	// **Gezählt wird alles**, auch was nicht gezeigt wird.
	if schau.Treffer != domain.GrenzeErsetzungsvorschau+5 {
		t.Errorf("%d Treffer", schau.Treffer)
	}

	if _, fehler := ersetzer(doppel).Ausfuehren(t.Context(), 1, gestellt); fehler != nil {
		t.Fatalf("Ausführen: %v", fehler)
	}

	if len(doppel.geschrieben) != domain.GrenzeErsetzungsvorschau+5 {
		t.Errorf("%d geschrieben — die Grenze gilt der Anzeige", len(doppel.geschrieben))
	}
}
