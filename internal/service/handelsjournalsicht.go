package service

// **Das Journal, wie ein Mensch es liest** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum der Name beim Lesen entsteht und nicht beim Buchen
//
// **Weil das Journal den öffentlichen Ausschnitt führt** — Druck, Menge,
// Zustand, Sprache, Ausführung. Ein eingefrorener Name wäre falsch, sobald
// jemand den Druck berichtigt, und er stünde in einem Buch, aus dem nichts
// herausgenommen wird.
//
// **Eine eigene Datei**, weil `handelsverbuchung.go` mit ihr über die
// Zeilengrenze ginge (§2). Der Schnitt folgt der Frage: dort, **was ein
// Abschluss tut** — hier, **was danach zu lesen ist**.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Druckbezeichnungen sagt, wie ein Druck heißt.
//
// **Nachgeschlagen wird beim Lesen und nicht beim Buchen.** Das Journal führt
// den **öffentlichen** Ausschnitt — Druck, Menge, Zustand, Sprache,
// Ausführung —, und ein eingefrorener Name wäre falsch, sobald jemand den
// Druck berichtigt. Wer nichts findet, behält die Kennung: sie sagt niemandem
// etwas, aber sie ist eine Auskunft.
type Druckbezeichnungen interface {
	NachScryfall(ctx context.Context, printing string) (*domain.Druck, error)
}

// Journaleintrag ist eine Bewegung samt ihrem lesbaren Namen.
type Journaleintrag struct {
	Zeile domain.Journalzeile

	// Bezeichnung ist, wie die Karte heißt — oder die Kennung des Drucks.
	Bezeichnung string
}

// Journal gibt die Bewegungen zu einem Vorgang — mit lesbaren Namen.
//
// # Der Fund, für den es diese Auflösung gibt
//
// **Am Fenster stand die Kennung des Drucks.** Gesehen bei der Sichtprüfung
// am 2026-09-18: „−2 × a1b2c3d4-0000-4000-8000-00000000000a · tausch". Ein
// DOM-Lauf fängt das nicht — „steht die Menge da" trifft auch dann zu —, und
// ein Journal, das niemand lesen kann, ist keins.
//
// **Ein Druck, der nicht zu finden ist, behält seine Kennung.** Ein
// empfangenes Stück nennt nur seinen Druck; wenn dieser Rechner ihn nicht
// kennt, gibt es keinen Namen — und einen zu erfinden wäre schlimmer.
func (v *Verbuchungsstelle) Journal(
	ctx context.Context, vorgang string,
) ([]Journaleintrag, error) {
	zeilen, fehler := v.bestand.Journal(ctx, vorgang)
	if fehler != nil {
		return nil, fehler
	}

	// **Je Druck einmal nachgeschlagen.** Ein Journal nennt denselben Druck
	// in beiden Richtungen und über viele Handel; eine Abfrage je Zeile
	// wüchse mit der Zahl der Bewegungen statt mit der der Drucke.
	namen := map[string]string{}
	eintraege := make([]Journaleintrag, 0, len(zeilen))

	for _, zeile := range zeilen {
		name, gesehen := namen[zeile.Printing]
		if !gesehen {
			gelesen, fehler := v.drucke.NachScryfall(ctx, zeile.Printing)
			if fehler != nil {
				return nil, fehler
			}

			name = zeile.Printing
			if gelesen != nil {
				name = domain.Kartenbezeichnung(zeile.Printing, gelesen.Name,
					gelesen.Setcode, gelesen.Nummer)
			}

			namen[zeile.Printing] = name
		}

		eintraege = append(eintraege,
			Journaleintrag{Zeile: zeile, Bezeichnung: name})
	}

	return eintraege, nil
}
