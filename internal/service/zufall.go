// **Woher die Mischung kommt** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum das eine Naht ist und kein Aufruf
//
// Das Issue verlangt es wörtlich: „Zufall wird im Betrieb sicher gezogen,
// **für Tests eingesetzt** und als verborgener Zustand reproduzierbar
// gehalten." Ein `rand.Shuffle` mitten im Regelausleger erfüllte den ersten
// Teil und keinen der beiden anderen — und ein Lauf, der eine Partie
// nachspielt, käme nie zweimal auf dasselbe Brett.
//
// # Was hier ausdrücklich **nicht** entsteht
//
// **Kein beweisbar faires Mischen.** G1 schließt es aus: „kein
// kryptografisch beweisbares faires Shuffle in v1". Der Wirt mischt, und er
// **kennt** die Reihenfolge danach — das ist dieselbe Vertrauensgrenze, die
// vor dem Beitritt offengelegt wird, und keine Lücke, die hier still
// entstünde.

package service

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zufallsquelle gibt eine Mischung.
//
// **Eine Permutation und nicht ein gemischtes Etwas**: der Aufrufer weiß, was
// er mischt, und will nur wissen, wohin jedes Stück kommt. So bleibt diese
// Naht frei von allem, was eine Partie ausmacht — und ein Lauf setzt eine
// Reihe ein, statt eine Bibliothek nachzubauen.
type Zufallsquelle interface {
	// Mische gibt eine Permutation von 0 bis anzahl-1.
	//
	// **Mit Fehler**, obwohl `crypto/rand` keinen gibt: die Quelle ist
	// eingesetzt, und eine eingesetzte kann brechen. Ein Mischen, das
	// stillschweigend die alte Reihenfolge behielte, wäre die schlimmste
	// aller Antworten — es sähe aus wie ein Mischen.
	Mische(anzahl int) ([]int, error)

	// Zahl gibt eine gleichverteilte Zahl in `[0, grenze)` — CR 705, CR 706
	// (Fassung 2026-08-19), Werkstück M15
	// (#1617).
	//
	// # Warum dieselbe Naht und keine zweite
	//
	// **Weil es dieselbe Frage ist.** Ein Münzwurf ist eine Zahl unter zwei,
	// ein sechsseitiger Würfel eine unter sechs, und eine Mischung ist eine
	// Folge davon. Zwei Quellen wären zwei Antworten auf dieselbe Frage —
	// und ein Lauf müsste zwei einsetzen, um eine Partie nachzuspielen.
	//
	// **Mit Fehler, aus demselben Grund wie bei [Zufallsquelle.Mische]**:
	// ein Wurf, der stillschweigend immer „Kopf" gäbe, sähe aus wie ein
	// Wurf.
	Zahl(grenze int) (int, error)
}

// EchterZufall mischt über die Quelle des Betriebssystems.
type EchterZufall struct {
	// bytes ist, woraus die Zahlen kommen — im Betrieb `crypto/rand.Reader`.
	bytes io.Reader
}

// NeuerZufall baut ihn für den Betrieb.
func NeuerZufall() *EchterZufall { return &EchterZufall{bytes: rand.Reader} }

// zufallAus baut ihn über einer eingesetzten Quelle — für Läufe (§3).
func zufallAus(bytes io.Reader) *EchterZufall {
	return &EchterZufall{bytes: bytes}
}

// Mische gibt eine gleichverteilte Permutation.
//
// **Fisher und Yates, rückwärts.** Jede der `anzahl!` Anordnungen ist gleich
// wahrscheinlich — vorausgesetzt, [EchterZufall.stelleUnter] verzerrt nicht,
// und genau dafür steht dort das Ablehnungsverfahren.
func (z *EchterZufall) Mische(anzahl int) ([]int, error) {
	// **Eine negative Zahl ist keine Menge.** Der Zweig steht hier, weil er
	// die Umwandlung darunter **tragbar** macht: ohne ihn wäre `uint64(i)`
	// eine Umdeutung, die bei einem Vorzeichen um die ganze Breite springt
	// (gosec G115). Eine Schranke, die man ohnehin will, ist besser als eine
	// Unterdrückung.
	if anzahl < 0 {
		return nil, fmt.Errorf(
			"%w: eine Mischung über %d Karten gibt es nicht",
			domain.ErrValidation, anzahl)
	}

	reihe := make([]int, anzahl)
	for i := range reihe {
		reihe[i] = i
	}

	for i := anzahl - 1; i > 0; i-- {
		gezogen, fehler := z.stelleUnter(uint64(i) + 1)
		if fehler != nil {
			return nil, fehler
		}

		// **Ohne Rückumwandlung**: Go erlaubt jeden ganzzahligen Typ als
		// Index, und `gezogen` ist kleiner als `i` — also kleiner als die
		// Länge.
		reihe[i], reihe[gezogen] = reihe[gezogen], reihe[i]
	}

	return reihe, nil
}

// Zahl gibt eine gleichverteilte Zahl in `[0, grenze)` — CR 705, CR 706.
//
// **Dieselbe Rechnung wie beim Mischen**: [EchterZufall.stelleUnter] mit
// seinem Ablehnungsverfahren. Der Grund steht dort, und er gilt hier
// genauso — ein Würfel, der die kleinen Zahlen bevorzugt, ist kein Würfel.
func (z *EchterZufall) Zahl(grenze int) (int, error) {
	if grenze <= 0 {
		return 0, fmt.Errorf(
			"%w: eine Zahl unter %d gibt es nicht", domain.ErrValidation,
			grenze)
	}

	gezogen, fehler := z.stelleUnter(uint64(grenze))
	if fehler != nil {
		return 0, fehler
	}

	// **Die Umdeutung ist hier tragbar, und gosec kann es nicht sehen.**
	// [EchterZufall.stelleUnter] gibt einen Rest modulo `grenze`, also eine
	// Zahl **kleiner** als `grenze` — und `grenze` ist ein `int`, der die
	// Schranke oben schon passiert hat. Eine zusätzliche Prüfung wäre ein
	// Zweig, den kein Weg betritt (§10); der Kommentar steht hier statt
	// seiner.
	//
	// **`#nosec` und nicht `//nolint`**: gosec läuft eigenständig und liest
	// die Anweisung an golangci-lint nicht (§4).
	return int(gezogen), nil // #nosec G115 -- gezogen < grenze, und grenze ist ein int
}

// stelleUnter gibt eine gleichverteilte Zahl in [0, grenze).
//
// # Warum nicht einfach modulo
//
// **Weil der Rest die kleinen Zahlen bevorzugt.** 2⁶⁴ ist durch 40 nicht
// teilbar; wer den Rest nähme, zöge die ersten Karten häufiger. Der
// Unterschied ist winzig und **systematisch**, und systematisch ist das
// Schlimmere: er verschwindet nicht dadurch, dass oft gemischt wird.
//
// **Abgelehnt wird deshalb alles oberhalb des größten Vielfachen.** Wie oft
// das vorkommt, hängt von `grenze` ab und ist bei einer Bibliothek
// verschwindend; dass es vorkommen **kann**, ist der Grund für die Schleife.
func (z *EchterZufall) stelleUnter(grenze uint64) (uint64, error) {
	// **Alles in `uint64`, und nichts wandert hinüber.** Der erste Anlauf
	// nahm und gab `int` und wandelte zweimal; gosec strich beide Stellen an
	// (G115), und beide Male zu Recht: eine Umdeutung, die niemand braucht,
	// ist ein Fehler, der auf eine Plattform wartet.
	schranke := ^uint64(0) / grenze * grenze

	for {
		var roh [8]byte

		if _, fehler := io.ReadFull(z.bytes, roh[:]); fehler != nil {
			return 0, domain.MitGrund(domain.ErrValidation, fehler,
				"Zufall für das Mischen ziehen")
		}

		if gezogen := binary.BigEndian.Uint64(roh[:]); gezogen < schranke {
			return gezogen % grenze, nil
		}
	}
}

// pruefeMischung sagt, ob eine Permutation eine ist.
//
// # Warum eine eingesetzte Quelle geprüft wird
//
// **Weil sie eingesetzt ist.** Eine Quelle, die `[0, 0, 0]` gibt, legte drei
// Karten auf dieselbe Stelle — und die Bibliothek verlöre zwei, ohne dass
// irgendwo ein Fehler stünde. Das ist die Klasse „still falsch", und sie
// gehört an der Naht abgefangen, nicht beim Leser der Bibliothek.
func pruefeMischung(mischung []int, anzahl int) error {
	if len(mischung) != anzahl {
		return fmt.Errorf(
			"%w: die Mischung hat %d Stellen für %d Karten",
			domain.ErrValidation, len(mischung), anzahl)
	}

	gesehen := make([]bool, anzahl)

	for _, stelle := range mischung {
		if stelle < 0 || stelle >= anzahl || gesehen[stelle] {
			return fmt.Errorf(
				"%w: die Mischung nennt die Stelle %d zweimal oder gar nicht",
				domain.ErrValidation, stelle)
		}

		gesehen[stelle] = true
	}

	return nil
}
