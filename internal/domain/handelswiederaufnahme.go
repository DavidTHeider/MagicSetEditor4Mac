package domain

// **Was nach einem Abriss fehlt** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum der Abdruck entscheidet und nicht die Zahl
//
// **Weil „wir sind beide bei Revision 3" nichts sagt.** Zwei Seiten können
// gegen dieselbe Basis verschiedene Dinge vorgeschlagen haben; dann gibt es
// zwei Revisionen 3, und sie haben verschiedene Abdrücke. Wer die Zahlen
// vergliche, lieferte munter nach — auf einen Verlauf, den es bei ihm nicht
// gibt.
//
// # Und warum ein Widerspruch offen bleibt
//
// **Weil ihn aufzulösen hieße, einen Stand zu verlieren.** Welcher der beiden
// gilt, ist eine Entscheidung zwischen zwei Menschen und keine Rechnung; das
// Programm hält beide fest und sagt, dass es zwei gibt
// ([HandelKlaerung]).

// Handelsfortschritt ist, was eine Seite als bestätigt kennt.
type Handelsfortschritt struct {
	Vorgang string

	// Revision ist die höchste Revision, die sie bestätigt hat — **0 heißt:
	// noch keine**.
	Revision int

	// Abdruck ist der Abdruck **dieser** Revision — er ist die Aussage.
	Abdruck string

	// Folge ist die höchste Folge, die von der **anderen** Seite bestätigt
	// angekommen ist.
	Folge int64
}

// Nachzuliefern sagt, ab welcher Folge zu liefern ist — oder dass es klemmt.
//
// **Zwei Rückgaben und kein Fehler.** „Es klemmt" ist kein Fehlerweg, sondern
// ein Ergebnis: beide Seiten haben ordentlich gearbeitet, und trotzdem stehen
// zwei Verläufe da. Ein `error` machte daraus eine Störung, die jemand
// wegprüft.
func Nachzuliefern(meiner, seiner Handelsfortschritt) (abFolge int64, klaerung bool) {
	if meiner.Vorgang != seiner.Vorgang {
		return 0, true
	}

	// **Ein Anfang ist keine Klärung.** Wer noch nichts bestätigt hat, kann
	// nichts widersprechen — und bekommt alles ab der ersten Folge.
	if seiner.Revision == 0 {
		return 1, false
	}

	// **Widersprochen wird nur auf derselben Höhe.** Ist eine Seite weiter,
	// kennt die andere ihren Abdruck noch gar nicht; das ist ein Rückstand
	// und kein Streit.
	if meiner.Revision == seiner.Revision && meiner.Abdruck != seiner.Abdruck {
		return 0, true
	}

	return seiner.Folge + 1, false
}
