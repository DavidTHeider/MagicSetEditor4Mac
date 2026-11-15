package domain

// **Wann ein Turnier abgeschlossen ist, und was es protokolliert** —
// DOM-04.AK-04 (#1699),
// Vermerk [0139](../../docs/entscheidungen/0139-wann-ein-turnier-abgeschlossen-ist.md).
//
// Portierungsmarke **new** — MSE2 kennt keine Turniere (Vermerk 0138).

// IstAbgeschlossen sagt, ob das Turnier ein Abschlussdatum trägt.
func (t Turnier) IstAbgeschlossen() bool { return t.Abgeschlossen != "" }

// Abschliessbar sagt, ob sich dieses Turnier abschließen lässt: **mindestens
// eine Runde, und kein offener Tisch**. Ein Bye ist ein Ergebnis
// (Vermerk 0062, Entscheidung 8).
func (t Turnier) Abschliessbar() error {
	if len(t.Runden) == 0 {
		return NewFieldError("turnierID", SchluesselTurnierNichtFertig,
			"Ein Turnier ohne Runde lässt sich nicht abschließen.")
	}

	for _, runde := range t.Runden {
		for _, paarung := range runde.Paarungen {
			if !paarung.IstBye() && paarung.Ergebnis == nil {
				return NewFieldError("turnierID", SchluesselTurnierNichtFertig,
					"In Runde %d hat eine Paarung noch kein Ergebnis.", runde.Nummer)
			}
		}
	}

	return nil
}

// TurnierIstAbgeschlossen ist das Nein an ein abgeschlossenes Turnier — ein
// Konflikt, denn die Eingabe ist in Ordnung, der Stand steht ihr entgegen.
func TurnierIstAbgeschlossen(name string) error {
	return NewFieldError("turnierID", SchluesselTurnierAbgeschlossen,
		"Das Turnier »%s« ist abgeschlossen — wer etwas ändern will, eröffnet es wieder.", name)
}

// Turniervorgang ist, was eine Zeile des Protokolls festhält.
type Turniervorgang string

// Die Vorgänge — **dieselben Wörter wie in der Tafel** (Wanderung 0089).
const (
	TurnierAngelegt        Turniervorgang = "angelegt"
	TurnierEingelesen      Turniervorgang = "eingelesen"
	TurnierAngemeldet      Turniervorgang = "angemeldet"
	TurnierAusgestiegen    Turniervorgang = "ausgestiegen"
	TurnierEingestiegen    Turniervorgang = "eingestiegen"
	TurnierRundeGelegt     Turniervorgang = "runde"
	TurnierErgebnis        Turniervorgang = "ergebnis"
	TurnierZurueckgenommen Turniervorgang = "zurueckgenommen"
	TurnierAbgeschlossen   Turniervorgang = "abgeschlossen"
	TurnierWiedereroeffnet Turniervorgang = "wiedereroeffnet"
)

// Protokolleintrag ist eine Zeile des Protokolls.
//
// **Angaben sind sprachfrei** — ein Name, eine Rundennummer oder ein
// Spielstand `2:0:0`. Den Satz dazu baut die Fläche in ihrer Sprache.
type Protokolleintrag struct {
	Zeit    string
	Vorgang Turniervorgang
	Angaben string

	// PaarungID nennt bei einem Ergebnis die Paarung — sonst null.
	PaarungID int64
}
