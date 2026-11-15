package domain

// **Der Ablauf eines Zuges** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Warum es das gibt
//
// **Weil G1 bis G8 einen Tisch gebaut haben und kein Spiel.** Jeder Sitz
// durfte jederzeit alles; `zugstruktur` und `prioritaet` standen in der
// Regelmatrix ausdrücklich auf `manuell` (G4, #1461). Auf die Frage, wer am
// Zug ist, gab dieses Programm keine Antwort.
//
// # Was hier steht und was nicht
//
// **Der Rahmen, nicht die Karte.** Der Stapel (CR 405), das Wirken von
// Zaubern (CR 601), die zustandsbasierten Aktionen (CR 704) und der
// Kampfschaden (CR 510) gehören zu Phase 5.8. Was ohne sie trägt, steht hier
// — und das ist mehr, als es klingt: solange der Stapel leer bleibt, ist
// „alle passen nacheinander" das einzige, was einen Abschnitt beendet, und
// **das ist nach CR 500.2 genau richtig** und keine Vereinfachung.
//
// # Und warum ein Abschnitt und nicht zwei Felder
//
// **Weil Phase und Schritt zusammengehören.** Eine Hauptphase hat keinen
// Schritt (CR 500.1 zählt nur Anfang, Kampf und Ende zu denen mit Schritten);
// zwei Felder nebeneinander ließen „Hauptphase im Enttappen-Schritt" zu, und
// niemand sähe es.

// Phase ist eine der fünf Phasen eines Zuges — CR 500.1 (Fassung
// 2026-08-19).
type Phase string

const (
	// PhaseAnfang ist die Anfangsphase — CR 501.
	PhaseAnfang Phase = "anfang"

	// PhaseHauptVorKampf ist die erste Hauptphase — CR 505.1 nennt sie
	// „precombat main phase".
	PhaseHauptVorKampf Phase = "hauptVorKampf"

	// PhaseKampf ist die Kampfphase — CR 506.
	PhaseKampf Phase = "kampf"

	// PhaseHauptNachKampf ist die zweite — CR 505.1, „postcombat main phase".
	PhaseHauptNachKampf Phase = "hauptNachKampf"

	// PhaseEnde ist die Endphase — CR 512.
	PhaseEnde Phase = "ende"
)

// Schritt ist ein Abschnitt innerhalb einer Phase.
//
// **Die leere Zeichenkette heißt: keiner** — und das ist kein Fehlwert,
// sondern die Hauptphase (CR 505.1).
type Schritt string

const (
	// Die drei Schritte der Anfangsphase — CR 501.1.
	SchrittEnttappen  Schritt = "enttappen"
	SchrittVersorgung Schritt = "versorgung"
	SchrittZiehen     Schritt = "ziehen"

	// Die fünf Schritte der Kampfphase — CR 506.1.
	//
	// **Sie stehen hier, obwohl G9 den Kampf nicht baut.** CR 500.1 verlangt
	// jede Phase in jedem Zug; ein Ablauf, der den Kampf ausließe, wäre
	// schneller und falsch. Was in ihnen **geschieht**, ist Phase 5.8.
	SchrittKampfbeginn  Schritt = "kampfbeginn"
	SchrittAngreifer    Schritt = "angreifer"
	SchrittBlocker      Schritt = "blocker"
	SchrittKampfschaden Schritt = "kampfschaden"
	SchrittKampfende    Schritt = "kampfende"

	// Die zwei Schritte der Endphase — CR 512.1.
	SchrittEnde       Schritt = "endeschritt"
	SchrittAufraeumen Schritt = "aufraeumen"
)

// Abschnitt ist eine Stelle im Zug — eine Phase und, wo es einen gibt, ihr
// Schritt.
type Abschnitt struct {
	Phase   Phase
	Schritt Schritt
}

// zugfolge ist die Reihenfolge eines Zuges — CR 500.1, 501.1, 506.1, 512.1.
//
// **Eine Aufzählung und keine Rechnung.** Die Reihenfolge steht in den
// Regeln; sie aus Phasen und Schritten zusammenzusetzen hieße, dieselbe
// Auskunft an zwei Stellen zu führen.
var zugfolge = []Abschnitt{
	{Phase: PhaseAnfang, Schritt: SchrittEnttappen},
	{Phase: PhaseAnfang, Schritt: SchrittVersorgung},
	{Phase: PhaseAnfang, Schritt: SchrittZiehen},
	{Phase: PhaseHauptVorKampf},
	{Phase: PhaseKampf, Schritt: SchrittKampfbeginn},
	{Phase: PhaseKampf, Schritt: SchrittAngreifer},
	{Phase: PhaseKampf, Schritt: SchrittBlocker},
	{Phase: PhaseKampf, Schritt: SchrittKampfschaden},
	{Phase: PhaseKampf, Schritt: SchrittKampfende},
	{Phase: PhaseHauptNachKampf},
	{Phase: PhaseEnde, Schritt: SchrittEnde},
	{Phase: PhaseEnde, Schritt: SchrittAufraeumen},
}

// Zugfolge gibt die zwölf Abschnitte eines Zuges in ihrer Reihenfolge.
//
// **Eine Kopie und nicht die Liste selbst**: ein Aufrufer, der sie umsortiert,
// sortierte die Regeln um.
func Zugfolge() []Abschnitt {
	folge := make([]Abschnitt, len(zugfolge))
	copy(folge, zugfolge)

	return folge
}

// VergibtPrioritaet sagt, ob in diesem Abschnitt jemand Priorität bekommt.
//
// **Zwei von zwölf tun es nicht**: das Enttappen (CR 502.4, und CR 117.3a
// sagt es von der anderen Seite) und das Aufräumen (CR 514.3).
//
// **Die Ausnahme CR 514.3a steht hier nicht**, und das ist Absicht: sie
// greift, wenn zustandsbasierte Aktionen anstünden oder Auslöser warteten —
// beides gibt es in dieser Phase nicht. Sie kommt mit 5.8, und bis dahin ist
// „normalerweise keiner" hier ohne Ausnahme wahr.
func (a Abschnitt) VergibtPrioritaet() bool {
	return a.Schritt != SchrittEnttappen && a.Schritt != SchrittAufraeumen
}

// Naechster gibt den folgenden Abschnitt und ob dabei ein neuer Zug anfängt.
//
// **Ein unbekannter Abschnitt fängt von vorn an.** Er kann nur aus einem
// Bestand kommen, den jemand von Hand geändert hat; eine Partie, die daran
// stehenbliebe, wäre unbenutzbar, und der Anfang eines Zuges ist der sichere
// Ausgang (§8).
func (a Abschnitt) Naechster() (Abschnitt, bool) {
	for stelle, dieser := range zugfolge {
		if dieser != a {
			continue
		}

		if stelle+1 < len(zugfolge) {
			return zugfolge[stelle+1], false
		}

		break
	}

	return zugfolge[0], true
}
