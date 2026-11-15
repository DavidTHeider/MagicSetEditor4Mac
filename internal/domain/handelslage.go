package domain

// **Wo ein Handel für mich steht** — TRADE-01.AK-01/02
// (#1715).
//
// # Vier Lagen über dreizehn Zuständen
//
// **Weil die Liste eine andere Frage beantwortet als die Matrix.** Die
// Matrix sagt, was geschehen ist; die Liste soll sagen, ob jemand etwas
// **tun** muss — und wer. „Angeboten" heißt für den, der zustimmen soll,
// „du bist dran", und für den, der schon zugestimmt hat, „wartet". Die
// dreizehn Zustände bleiben darunter stehen (TRADE-02.AK-01); die Lage wird
// aus ihnen und den Stimmen **abgeleitet** und nirgends gespeichert.
//
// Portierungsmarke **new**: MSE2 handelt nicht (gesucht nach `trade`,
// `offer`, `exchange` in `/Users/david/mse2-quelle/src`).

import "slices"

// Handelslage sagt, ob ich an einem Vorgang etwas tun muss.
type Handelslage string

const (
	// LageDran heißt: ein Schritt des üblichen Weges ist meiner.
	LageDran Handelslage = "dran"

	// LageWartet heißt: der nächste Schritt ist der des Gegenübers.
	LageWartet Handelslage = "wartet"

	// LageAbgeschlossen ist „beidseitig verbucht" — das einzige gute Ende.
	LageAbgeschlossen Handelslage = "abgeschlossen"

	// LageAbgebrochen sind die drei anderen Enden.
	LageAbgebrochen Handelslage = "abgebrochen"
)

// Handelslagen gibt die vier in der Reihenfolge, in der eine Liste sie
// zeigt: was zu tun ist, zuerst.
func Handelslagen() []Handelslage {
	return []Handelslage{LageDran, LageWartet, LageAbgeschlossen, LageAbgebrochen}
}

// Gueltig sagt, ob es die Lage gibt.
func (l Handelslage) Gueltig() bool { return slices.Contains(Handelslagen(), l) }

// nebenwege sind die Schritte, die einen Vorgang verlassen oder von ihm
// abzweigen, statt ihn weiterzuführen — **wer nur sie hat, ist nicht dran**.
//
// **Das Gegenangebot gehört dazu**: wer einem Angebot zugestimmt hat, darf
// noch eines machen, aber er muss nicht — der Weg wartet auf die Stimme der
// anderen.
var nebenwege = []Handelsschritt{
	SchrittGegenangebot, SchrittAblehnen, SchrittAbbrechen, SchrittAblaufen, SchrittKlaeren,
}

// Moegliche gibt die Schritte, die `ich` aus dem heutigen Zustand gehen
// darf — in der Reihenfolge der Matrix, der übliche Weg vor den Ausgängen.
//
// **Eine abgegebene Stimme ist gegangen**: wo alle Parteien den Schritt tun
// müssen, fehlt er dem, der schon zugestimmt hat. Dieselbe Prüfung wie in
// [Handelsangebot.Schreite], nur ohne ihn zu gehen.
func (a Handelsangebot) Moegliche(ich string) []Handelsschritt {
	if !slices.Contains(a.Parteien, ich) {
		return nil
	}

	var schritte []Handelsschritt

	for _, uebergang := range Handelsmatrix() {
		if uebergang.Von != a.Zustand || !a.darf(uebergang.Wer, ich) ||
			(uebergang.Alle && slices.Contains(a.Zustimmungen, ich)) {
			continue
		}

		schritte = append(schritte, uebergang.Schritt)
	}

	return schritte
}

// Naechster gibt den Schritt des üblichen Weges, den `ich` gehen darf —
// leer, wenn es keinen gibt. **Aus einem Zustand führt höchstens einer**:
// die Matrix hat je Zustand genau einen Übergang, der nicht abzweigt oder
// endet.
func (a Handelsangebot) Naechster(ich string) Handelsschritt {
	for _, schritt := range a.Moegliche(ich) {
		if !slices.Contains(nebenwege, schritt) {
			return schritt
		}
	}

	return ""
}

// Lage sagt, wo der Vorgang für `ich` steht.
//
// **Eine Klärung ist immer dran** — sie ist der Fall, den niemand
// automatisch löst ([HandelKlaerung]), und ein Mensch auf jeder Seite muss
// sie ansehen.
func (a Handelsangebot) Lage(ich string) Handelslage {
	switch a.Zustand {
	case HandelBeidseitigVerbucht:
		return LageAbgeschlossen

	case HandelAbgelehnt, HandelAbgebrochen, HandelAbgelaufen:
		return LageAbgebrochen

	case HandelKlaerung:
		if slices.Contains(a.Parteien, ich) {
			return LageDran
		}
	}

	if a.Naechster(ich) != "" {
		return LageDran
	}

	return LageWartet
}

// Unbestaetigt sagt, ob dieser Bestand fortgeschrieben ist und der des
// Gegenübers noch nicht bestätigt — **einseitig verbucht ist nicht fertig**
// (AK-04).
func (a Handelsangebot) Unbestaetigt() bool {
	return a.Zustand == HandelLokalVerbucht || a.Zustand == HandelAbgleichOffen
}
