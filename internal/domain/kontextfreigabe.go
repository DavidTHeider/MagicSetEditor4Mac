package domain

// **Was privat ist, geht nur hinaus, wenn jemand es ausgewählt hat** —
// Werkstück A13.4 von Phase 5.3
// (#1452).
//
// # Der Satz, um den herum das gebaut ist
//
// **„Referenzkarten/Decks/private Setnotizen nur ausdrücklich ausgewählt."**
//
// Bis hierher entschied allein der **Fall**, in welchen Korpora gesucht wird
// ([Recherchefall.Quellen]) — und `aehnlich` suchte von sich aus in den
// eigenen Karten. Wer eine Karte mit einem Modell verglich, schickte damit
// seine eigenen mit, ohne es zu sagen und ohne es zu sehen.
//
// # Warum die Freigabe an der Anfrage hängt und nicht am Fall
//
// **Weil sie eine Entscheidung ist und keine Eigenschaft der Frage.** Derselbe
// Fall darf einmal mit und einmal ohne den eigenen Bestand laufen; wer das an
// den Fall hinge, bräuchte zwei Fälle für dieselbe Frage.
//
// # Und warum eine wirkungslose Freigabe ein Fehler ist
//
// **Weil ein Kreuz ohne Wirkung eine Zusage ist, die niemand einlöst** (§10).
// Wer „meine Karten sind freigegeben" ankreuzt und eine Regelrecherche fährt,
// hat etwas anderes entschieden, als er glaubt — und bekäme eine Antwort, die
// aussieht, als sei sein Kreuz berücksichtigt worden.

import (
	"fmt"
	"slices"
)

// IstPrivat sagt, ob diese Quelle eigenen Inhalt führt.
//
// **Was niemand kennt, gilt als privat.** Dieselbe Regel wie bei der
// Datenkategorie in `service`: eine neue Quelle stillschweigend als
// öffentlich zu führen wäre die Gestalt, in der sie unbemerkt hinausginge.
func (q Retrievalquelle) IstPrivat() bool {
	switch q {
	case QuelleOracle, QuelleRuling, QuelleRegelwerk:
		return false

	case QuelleEigeneKarte, QuelleEigenerText:
		return true
	}

	return true
}

// Suchquellen sind die Korpora, in denen dieser Ruf **wirklich** sucht.
//
// **Die des Falls, abzüglich der privaten ohne Freigabe.**
func (a Rechercheanfrage) Suchquellen() []Retrievalquelle {
	moegliche := a.Fall.Quellen()
	gesucht := make([]Retrievalquelle, 0, len(moegliche))

	for _, korpus := range moegliche {
		if korpus.IstPrivat() && !slices.Contains(a.Freigegeben, korpus) {
			continue
		}

		gesucht = append(gesucht, korpus)
	}

	return gesucht
}

// AusgelasseneQuellen sind die privaten, die dieser Fall suchen könnte und
// für die keine Freigabe vorliegt.
//
// **Sie stehen in der Antwort als Einschränkung** (§8): ein Korpus, in dem
// nicht gesucht wurde, erklärt, warum etwas fehlt — und „nichts gefunden"
// allein erklärt es nicht.
func (a Rechercheanfrage) AusgelasseneQuellen() []Retrievalquelle {
	ausgelassen := make([]Retrievalquelle, 0, 2)

	for _, korpus := range a.Fall.Quellen() {
		if korpus.IstPrivat() && !slices.Contains(a.Freigegeben, korpus) {
			ausgelassen = append(ausgelassen, korpus)
		}
	}

	return ausgelassen
}

// pruefeFreigabe weist eine Freigabe ab, die nichts bewirkt.
//
// **Die Schleifenvariable heißt `korpus` und nicht `quelle`.** Der Wächter
// über die Meldungswege geht nach dem **Bezeichner** und hielte einen Wert
// namens `quelle` für einen Dateipfad — er stünde damit in der Ausnahmeliste
// statt im Code. Ein Name, der eine Prüfung falsch auslöst, wird geändert;
// eine Ausnahme dafür wäre eine Zeile, die niemand mehr liest.
func (a Rechercheanfrage) pruefeFreigabe() error {
	moegliche := a.Fall.Quellen()

	for _, korpus := range a.Freigegeben {
		if !korpus.IstPrivat() {
			return fmt.Errorf(
				"%w: %q ist kein privater Korpus — er wird ohnehin gesucht, "+
					"und eine Freigabe dafür bedeutet nichts",
				ErrValidation, korpus)
		}

		if !slices.Contains(moegliche, korpus) {
			return fmt.Errorf(
				"%w: eine Recherche nach %q sucht nicht in %q — die Freigabe "+
					"dafür bliebe ohne Wirkung",
				ErrValidation, a.Fall, korpus)
		}
	}

	return nil
}
