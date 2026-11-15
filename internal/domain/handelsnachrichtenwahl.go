package domain

// **Welche Nachricht zu welchem Stand gehört** — Werkstück `T5` von Phase 5.6
// (#1470),
// angeschlossen mit
// #1637.
//
// # Warum die Wahl hier steht und nicht im Absender
//
// **Weil sie eine Aussage über den Vorgang ist und keine über den
// Versandweg.** Wer sie im Absender träfe, träfe sie je Fläche neu — und die
// zweite Fläche schickte irgendwann eine andere Art für denselben Stand. Die
// Gegenseite deutet die Art; zwei Deutungen desselben Standes sind zwei
// Verläufe.
//
// # Und warum nicht jeder Zustand eine hat
//
// **Weil drei von dreizehn niemanden drüben angehen** — der Grund steht je
// Fall an [NachrichtenartFuer]. Ein Zustand, der stumm bleibt, ist eine
// Entscheidung; eine Nachricht, die man hinschicken könnte, weil das Feld
// gerade leer ist, wäre keine.

// NachrichtenartFuer sagt, welche Nachricht diesen Stand hinausträgt.
//
// **Zwei Rückgaben und kein leerer Wert.** „Aus diesem Zustand geht nichts
// hinaus" ist eine Antwort; wer sie als leere Art ausgäbe, ließe sie beim
// ersten `if art != ""` wie einen Fehler aussehen.
func NachrichtenartFuer(zustand Handelszustand) (Handelsnachrichtenart, bool) {
	switch zustand {
	case HandelAngeboten:
		return NachrichtVorschlag, true

	case HandelVerhandlung:
		return NachrichtGegenvorschlag, true

	case HandelAngenommen:
		return NachrichtAnnahme, true

	case HandelVorbereitet:
		return NachrichtVorbereitung, true

	case HandelUebergabeOffen:
		return NachrichtUebergabe, true

	// **Zwei Zustände, eine Art**: `lokalVerbucht` und `abgleichOffen` sagen
	// der Gegenseite dasselbe — „bei mir ist gebucht". Der Unterschied
	// zwischen beiden ist, ob hier schon jemand nachgefragt hat, und das ist
	// eine Sache dieser Seite.
	case HandelLokalVerbucht, HandelAbgleichOffen:
		return NachrichtCommitstand, true

	case HandelBeidseitigVerbucht:
		return NachrichtBeleg, true

	case HandelAbgelehnt:
		return NachrichtAblehnung, true

	case HandelAbgebrochen:
		return NachrichtAbbruch, true
	}

	// **Die drei stummen**: `entwurf` hat die Gegenseite nie gesehen;
	// `abgelaufen` ist eine Beobachtung an **dieser** Uhr, und drüben geht
	// eine andere; `klaerung` ist der Fall, den zwei Menschen lösen — eine
	// Nachricht darüber ersetzte das Gespräch nicht und sähe aus, als täte
	// sie es.
	return "", false
}
