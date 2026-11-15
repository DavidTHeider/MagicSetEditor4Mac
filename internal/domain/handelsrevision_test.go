package domain_test

// **Die Revision eines Angebots** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum das neben dem Angebot steht
//
// **Weil es eine andere Frage beantwortet** (§2): nebenan steht, **was** ein
// Angebot ist und woran seine Unterschrift hängt; hier steht, **was aus ihm
// wird**, wenn jemand etwas ändert.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine geteilte Bewertung geht in die Unterschrift ein, eine stille nicht.**
//
// # Die Entscheidung, die dahintersteht
//
// **Eine Marktbasis ist eine Anzeige.** Wer eine andere Ansicht wählt, ändert
// keine unterschriebene Einigung — sonst machte jeder neue Preis eine gültige
// Einigung ungültig. **Ist die Bewertung aber Vertragsinhalt**, dann ist ihre
// Änderung eine Änderung des Angebots und braucht eine neue Revision.
//
// **Beides ist vertretbar; was nicht geht, ist es offenzulassen** — deshalb
// entscheidet es ein Feld und nicht die Gewohnheit des Aufrufers.
func TestEineGeteilteBewertungBindetUndEineStilleNicht(t *testing.T) {
	t.Parallel()

	still := einAngebot()
	stillTeurer := still
	stillTeurer.Positionen = []domain.Handelsposition{still.Positionen[0]}
	stillTeurer.Positionen[0].Bewertung = 4_200

	if stillTeurer.Abdruck() != still.Abdruck() {
		t.Error("eine stille Bewertung ändert den Abdruck")
	}

	geteilt := einAngebot()
	geteilt.BewertungGeteilt = true

	geteiltTeurer := geteilt
	geteiltTeurer.Positionen = []domain.Handelsposition{geteilt.Positionen[0]}
	geteiltTeurer.Positionen[0].Bewertung = 4_200

	if geteiltTeurer.Abdruck() == geteilt.Abdruck() {
		t.Error("eine geteilte Bewertung ändert den Abdruck nicht")
	}

	// **Und der Schalter selbst bindet auch.** Sonst hätten „geteilt, aber
	// null" und „still" denselben Abdruck, und die Unterschrift sagte nicht,
	// worüber sie geschlossen wurde.
	if geteilt.Abdruck() == still.Abdruck() {
		t.Error("der Schalter für die geteilte Bewertung bindet nicht")
	}
}

// **Eine Änderung erzeugt die nächste Revision und nimmt beide Zustimmungen.**
//
// # Warum die Zustimmungen fallen
//
// **Weil sie einem Stand galten, der nicht mehr der letzte ist.** Wer sie
// mitnähme, hätte eine Einigung über etwas, das niemand gelesen hat — und
// genau das ist der Fall, in dem hinterher niemand mehr sagen kann, worauf
// man sich geeinigt hatte.
func TestEineAenderungNimmtBeideZustimmungen(t *testing.T) {
	t.Parallel()

	erste := einAngebot()
	erste.Zustand = domain.HandelAngeboten
	erste.Zustimmungen = []string{"anna", "bert"}

	anders := erste.Positionen[0]
	anders.Menge = 5

	zweite := erste.Fortgeschrieben([]domain.Handelsposition{anders})

	if zweite.Revision != 2 {
		t.Errorf("die nächste Revision ist %d", zweite.Revision)
	}

	if zweite.Eltern != erste.Abdruck() {
		t.Errorf("die Eltern sind %q, der Abdruck der ersten ist %q",
			zweite.Eltern, erste.Abdruck())
	}

	if len(zweite.Zustimmungen) != 0 {
		t.Errorf("die Zustimmungen stehen noch: %v", zweite.Zustimmungen)
	}

	if zweite.Zustand != domain.HandelVerhandlung {
		t.Errorf("der Zustand ist %q", zweite.Zustand)
	}

	if zweite.Positionen[0].Menge != 5 {
		t.Errorf("die Menge ist %d", zweite.Positionen[0].Menge)
	}

	// **Und die erste Revision bleibt, wie sie war** — sonst wäre eine
	// Zustimmung eine Aussage über etwas Veränderliches.
	if erste.Positionen[0].Menge != 2 || len(erste.Zustimmungen) != 2 {
		t.Errorf("die erste Revision hat sich geändert: %+v", erste)
	}

	if fehler := zweite.Pruefe(); fehler != nil {
		t.Errorf("die fortgeschriebene Revision prüft nicht: %v", fehler)
	}
}

// **Und die fortgeschriebene Revision hängt an genau einer Elternrevision.**
//
// **Zwei Vorschläge gegen denselben Stand ergeben zwei verschiedene
// Angebote** — und beide nennen dieselben Eltern. Wer daraus zwei Revisionen
// mit der Nummer 2 machte, hätte zwei Wahrheiten; wer sequenziert, hat einen
// **Konflikt**, und den sieht man an dieser Stelle: die Abdrücke sind
// verschieden, die Eltern gleich.
func TestZweiVorschlaegeGegenDenselbenStandSindEinKonflikt(t *testing.T) {
	t.Parallel()

	erste := einAngebot()

	vonAnna := erste.Positionen[0]
	vonAnna.Menge = 3

	vonBert := erste.Positionen[0]
	vonBert.Menge = 4

	eine := erste.Fortgeschrieben([]domain.Handelsposition{vonAnna})
	andere := erste.Fortgeschrieben([]domain.Handelsposition{vonBert})

	if eine.Eltern != andere.Eltern {
		t.Error("die beiden Vorschläge nennen verschiedene Eltern")
	}

	if eine.Revision != andere.Revision {
		t.Error("die beiden Vorschläge tragen verschiedene Revisionen")
	}

	if eine.Abdruck() == andere.Abdruck() {
		t.Error("die beiden Vorschläge haben denselben Abdruck")
	}
}
