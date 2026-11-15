package service

// **Die Zufallsnaht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Ein Lauf im Paket**, weil `zufallAus` die Naht ist, an der die Bytes
// hereinkommen — und weil das Ablehnungsverfahren sich nur von dort aus
// betreten lässt: bei einer Bibliothek trifft es rechnerisch nie zu, und ein
// Zweig, den kein Lauf betritt, ist kein geprüfter Zweig (§3).

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Mischung ist eine Permutation — jede Stelle genau einmal.**
func TestEchterZufallMischtVollstaendig(t *testing.T) {
	t.Parallel()

	for _, anzahl := range []int{0, 1, 2, 7, 60} {
		mischung, fehler := NeuerZufall().Mische(anzahl)
		if fehler != nil {
			t.Fatalf("%d Karten mischen: %v", anzahl, fehler)
		}

		if fehler := pruefeMischung(mischung, anzahl); fehler != nil {
			t.Errorf("%d Karten: %v (%v)", anzahl, fehler, mischung)
		}
	}
}

// **Zweimal mischen gibt zweimal Verschiedenes.**
//
// **Die Gegenprobe zum Lauf darüber**: `Mische` gäbe auch dann eine gültige
// Permutation, wenn es gar nicht mischte — `[0, 1, 2, …]` ist eine. Bei 60
// Karten ist die Wahrscheinlichkeit, zweimal dieselbe zu ziehen, 1/60!, und
// das ist die Zahl, die diesen Lauf trägt.
func TestEchterZufallMischtWirklich(t *testing.T) {
	t.Parallel()

	const karten = 60

	erste, fehler := NeuerZufall().Mische(karten)
	if fehler != nil {
		t.Fatalf("erste Mischung: %v", fehler)
	}

	zweite, fehler := NeuerZufall().Mische(karten)
	if fehler != nil {
		t.Fatalf("zweite Mischung: %v", fehler)
	}

	gleich := 0

	for i := range erste {
		if erste[i] == zweite[i] {
			gleich++
		}
	}

	if gleich == karten {
		t.Errorf("zweimal dieselbe Mischung: %v", erste)
	}

	// **Und keine von beiden ist die Ausgangsreihe.**
	unveraendert := true

	for i, stelle := range erste {
		if stelle != i {
			unveraendert = false

			break
		}
	}

	if unveraendert {
		t.Errorf("die Mischung ist die Ausgangsreihe: %v", erste)
	}
}

// **Eine negative Menge ist keine.**
//
// **Der Zweig trägt die Umwandlung darunter** — ohne ihn wäre `uint64(i)`
// eine Umdeutung, die bei einem Vorzeichen um die ganze Breite springt.
func TestEchterZufallWeistEineNegativeMengeAb(t *testing.T) {
	t.Parallel()

	_, fehler := NeuerZufall().Mische(-1)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", fehler)
	}
}

// **Eine Quelle, die nichts mehr hergibt, mischt nicht still weiter.**
func TestEchterZufallMeldetEineLeereQuelle(t *testing.T) {
	t.Parallel()

	_, fehler := zufallAus(bytes.NewReader(nil)).Mische(5)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet eine Abweisung", fehler)
	}
}

// **Eine Zahl oberhalb der Schranke wird abgelehnt und nicht genommen.**
//
// # Wie dieser Lauf die Schranke trifft
//
// Bei `grenze = 3` ist das größte Vielfache von 3 unter 2⁶⁴ genau
// `2⁶⁴ − 1`; die einzige abzulehnende Zahl ist also `2⁶⁴ − 1` selbst. Die
// Quelle gibt sie zuerst und danach eine brauchbare — kommt die abgelehnte
// trotzdem durch, steht am Ende eine andere Stelle da.
func TestEchterZufallLehntEineVerzerrendeZahlAb(t *testing.T) {
	t.Parallel()

	// **Sechzehn Byte in einem Stück**: acht für die Zahl, die abgelehnt
	// wird, acht für die, die durchkommt.
	roh := make([]byte, 16)
	binary.BigEndian.PutUint64(roh[:8], ^uint64(0))
	binary.BigEndian.PutUint64(roh[8:], 7)

	quelle := bytes.NewReader(roh)

	gezogen, fehler := zufallAus(quelle).stelleUnter(3)
	if fehler != nil {
		t.Fatalf("ziehen: %v", fehler)
	}

	if gezogen != 1 {
		t.Errorf("gezogen %d, erwartet 7 mod 3 = 1 — die abgelehnte Zahl "+
			"hätte %d gegeben", gezogen, ^uint64(0)%3)
	}
}

// **Was `pruefeMischung` durchlässt und was nicht.**
//
// **Der Erfolgsweg steht zuerst** — ohne ihn sagte „wird abgewiesen" nur,
// dass irgendetwas fehlt (§10).
func TestPruefeMischung(t *testing.T) {
	t.Parallel()

	if fehler := pruefeMischung([]int{2, 0, 1}, 3); fehler != nil {
		t.Fatalf("eine gültige Mischung: %v", fehler)
	}

	for name, fall := range map[string]struct {
		mischung []int
		anzahl   int
	}{
		"zu kurz":       {[]int{0, 1}, 3},
		"zu lang":       {[]int{0, 1, 2, 3}, 3},
		"doppelt":       {[]int{0, 0, 1}, 3},
		"zu groß":       {[]int{0, 1, 3}, 3},
		"unter null":    {[]int{0, 1, -1}, 3},
		"leer bei drei": {nil, 3},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := pruefeMischung(fall.mischung, fall.anzahl); fehler == nil {
				t.Errorf("%v für %d Karten kam durch", fall.mischung, fall.anzahl)
			}
		})
	}
}

// **Eine Quelle ohne Bytes lässt niemanden würfeln** — M15
// (#1617).
//
// **Derselbe Weg wie beim Mischen** (`stelleUnter`), und deshalb derselbe
// Ausgang: ein Würfel, der stillschweigend eins gäbe, sähe aus wie ein
// Würfel.
func TestEineQuelleOhneBytesLaesstNiemandenWuerfeln(t *testing.T) {
	t.Parallel()

	if _, fehler := zufallAus(bytes.NewReader(nil)).Zahl(6); fehler == nil {
		t.Error("es wurde gewürfelt, obwohl keine Bytes da sind")
	}
}
