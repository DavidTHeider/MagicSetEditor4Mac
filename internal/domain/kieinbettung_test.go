package domain_test

// **Der Einbettungs- und Neuordnungsvertrag** — Werkstück A6 von Phase 5.3
// (#1445).
//
// Das Issue verlangt einen **unabhängigen** Vertragstest „ohne
// Generierungsprovider": was hier geprüft wird, ist die **Form** der Antwort
// und nichts über ihren Inhalt.

import (
	"errors"
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einVertrag baut einen gültigen Einbettungsvertrag.
func einVertrag() domain.KIEinbettungsvertrag {
	return domain.KIEinbettungsvertrag{
		Dimension:         768,
		Pooling:           domain.KIPoolingMittel,
		Normalisiert:      true,
		Zerteiler:         "sentencepiece:beispiel",
		Frageanweisung:    "Frage: ",
		Dokumentanweisung: "Passage: ",
	}
}

// **Ein Vektor der zugesagten Länge geht durch.**
func TestEinPassenderVektorGehtDurch(t *testing.T) {
	t.Parallel()

	vektor := make([]float32, 768)
	vektor[0] = 1

	if fehler := einVertrag().PruefeVektor(vektor); fehler != nil {
		t.Fatalf("der Vektor wurde abgewiesen: %v", fehler)
	}
}

// **Eine falsche Länge wird abgewiesen.**
//
// Das Issue nennt es als Abnahme: „falsche Dimension … abweisen". Ein Vektor
// anderer Länge ist kein schlechterer Vektor — er kommt aus einem anderen
// Modell, und wer ihn in denselben Index legte, verglich Unvergleichbares.
func TestEineFalscheDimensionWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, laenge := range map[string]int{
		"zu kurz": 767, "zu lang": 769, "leer": 0,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := einVertrag().PruefeVektor(make([]float32, laenge)); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **NaN und Unendlich werden abgewiesen.**
//
// **Sie sind keine Zahlen, und sie stecken an**: ein NaN in einem Vektor macht
// jedes Skalarprodukt damit zu NaN, und eine Ähnlichkeitssuche gäbe danach
// stumm keine Treffer mehr.
func TestKeineZahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, wert := range map[string]float32{
		"NaN":             float32(math.NaN()),
		"Unendlich":       float32(math.Inf(1)),
		"minus Unendlich": float32(math.Inf(-1)),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vektor := make([]float32, 768)
			vektor[42] = wert

			if fehler := einVertrag().PruefeVektor(vektor); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Ein Vertrag ohne Dimension oder mit unbekanntem Pooling wird abgewiesen.**
func TestEinUnvollstaendigerVertragWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.KIEinbettungsvertrag){
		"ohne Dimension":          func(v *domain.KIEinbettungsvertrag) { v.Dimension = 0 },
		"mit negativer Dimension": func(v *domain.KIEinbettungsvertrag) { v.Dimension = -1 },
		"mit erfundenem Pooling":  func(v *domain.KIEinbettungsvertrag) { v.Pooling = "raten" },
		"ohne Zerteiler":          func(v *domain.KIEinbettungsvertrag) { v.Zerteiler = "" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vertrag := einVertrag()
			kaputt(&vertrag)

			if fehler := vertrag.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%s kam durch: %v", name, fehler)
			}
		})
	}
}

// **Der Fingerabdruck ändert sich, wenn sich der Vertrag ändert.**
//
// # Wofür es ihn gibt
//
// **Für A15** (#1514):
// der Index entscheidet daran, ob er neu gebaut werden muss. Das Issue sagt
// „Cache-/Fingerprintänderungen melden; A15 entscheidet Reindexierung" — also
// muss er sich bei **jedem** Teil ändern, der das Ergebnis verschiebt.
func TestDerFingerabdruckHaengtAnJedemTeil(t *testing.T) {
	t.Parallel()

	grund := einVertrag().Fingerabdruck()

	for name, aendern := range map[string]func(*domain.KIEinbettungsvertrag){
		"die Dimension":         func(v *domain.KIEinbettungsvertrag) { v.Dimension = 1024 },
		"das Pooling":           func(v *domain.KIEinbettungsvertrag) { v.Pooling = domain.KIPoolingErster },
		"die Normalisierung":    func(v *domain.KIEinbettungsvertrag) { v.Normalisiert = false },
		"der Zerteiler":         func(v *domain.KIEinbettungsvertrag) { v.Zerteiler = "anders" },
		"die Frageanweisung":    func(v *domain.KIEinbettungsvertrag) { v.Frageanweisung = "Query: " },
		"die Dokumentanweisung": func(v *domain.KIEinbettungsvertrag) { v.Dokumentanweisung = "Doc: " },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vertrag := einVertrag()
			aendern(&vertrag)

			if vertrag.Fingerabdruck() == grund {
				t.Errorf("%s ändert den Fingerabdruck nicht", name)
			}
		})
	}

	// **Und zweimal derselbe Vertrag gibt denselben Abdruck** — ohne das wäre
	// jeder Vergleich ein Fehlalarm.
	if einVertrag().Fingerabdruck() != grund {
		t.Error("derselbe Vertrag gibt zwei verschiedene Abdrücke")
	}
}

// **Frage und Dokument sind zwei Anweisungen, nicht eine.**
//
// Das Issue verlangt „getrennte Query-/Dokument-Embeddingkonfiguration".
// Viele Modelle wollen einen Text je nach Rolle anders eingeleitet sehen; wer
// beides gleich behandelte, bekäme Vektoren, die systematisch danebenliegen —
// ohne dass etwas scheitert.
func TestFrageUndDokumentSindZweiAnweisungen(t *testing.T) {
	t.Parallel()

	vertrag := einVertrag()

	if vertrag.FuerFrage("Blitzschlag") != "Frage: Blitzschlag" {
		t.Errorf("die Frage lautet %q", vertrag.FuerFrage("Blitzschlag"))
	}

	if vertrag.FuerDokument("Blitzschlag") != "Passage: Blitzschlag" {
		t.Errorf("das Dokument lautet %q", vertrag.FuerDokument("Blitzschlag"))
	}

	// **Leer heißt: keine Anweisung** und nicht „ein leerer Vorspann".
	ohne := domain.KIEinbettungsvertrag{}

	if ohne.FuerFrage("x") != "x" || ohne.FuerDokument("x") != "x" {
		t.Error("ein leerer Vertrag setzte trotzdem etwas davor")
	}
}

// **Eine Neuordnung gibt Bewertungen, keine Reihenfolge.**
func TestEineNeuordnungGibtBewertungen(t *testing.T) {
	t.Parallel()

	bewertungen := []domain.KIBewertung{
		{Kandidat: "eins", Wert: 0.9},
		{Kandidat: "zwei", Wert: 0.1},
	}

	if fehler := domain.PruefeBewertungen(bewertungen, 2); fehler != nil {
		t.Fatalf("die Bewertungen wurden abgewiesen: %v", fehler)
	}
}

// **Mehr Bewertungen als Kandidaten werden abgewiesen.**
//
// Das Issue nennt es als Abnahme: „überzählige Vektoren … abweisen". Eine
// Bewertung zu einem Kandidaten, den niemand geschickt hat, ist eine
// erfundene Aussage.
func TestUeberzaehligeBewertungenWerdenAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		bewertungen []domain.KIBewertung
		kandidaten  int
	}{
		"mehr als Kandidaten": {
			[]domain.KIBewertung{{Kandidat: "a", Wert: 1}, {Kandidat: "b", Wert: 1}}, 1,
		},
		"ohne Kandidatennamen": {
			[]domain.KIBewertung{{Wert: 1}}, 1,
		},
		"mit demselben Kandidaten zweimal": {
			[]domain.KIBewertung{{Kandidat: "a", Wert: 1}, {Kandidat: "a", Wert: 2}}, 2,
		},
		"mit einem Wert, der keine Zahl ist": {
			[]domain.KIBewertung{{Kandidat: "a", Wert: math.NaN()}}, 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := domain.PruefeBewertungen(
				fall.bewertungen, fall.kandidaten); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("%s kam durch: %v", name, fehler)
			}
		})
	}
}
