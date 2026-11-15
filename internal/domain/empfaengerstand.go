package domain

// **Was schon da war, kommt nicht zweimal** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum ein Stand und keine Liste
//
// **Weil eine Reihenfolge lückenlos ist.** Der Wirt legt sie fest; wer
// empfängt, braucht sich nicht zu merken, was alles kam — es genügt, was
// **zuletzt** kam. Eine Menge aller gesehenen Folgen wüchse mit der Partie und
// beantwortete keine Frage, die diese zwei Zahlen nicht beantworten.
//
// # Und warum er blockiert, statt zu überspringen
//
// **Das Issue sagt es wörtlich**: „Doppelte/geänderte IDs, Lücken, falsche
// Hashes und alte Epochs blockieren statt weiterzuspielen." Wer eine Lücke
// überspränge, hätte einen Stand, den niemand nachrechnen kann — und der
// Fehler zeigte sich erst Züge später an einer Karte, die woanders liegt.

import "fmt"

// Folgeurteil sagt, was mit einer Nachricht geschieht.
type Folgeurteil int

// Die vier Urteile.
const (
	// FolgeNeu ist die erwartete nächste — sie wird genommen.
	FolgeNeu Folgeurteil = iota

	// FolgeSchonDa war schon da. **Sie wird verworfen und nicht gemeldet**:
	// eine zweite Zustellung ist der Normalfall auf einer Leitung, die
	// wiederholt.
	FolgeSchonDa

	// FolgeLuecke liegt jenseits der nächsten — dazwischen fehlt etwas.
	FolgeLuecke

	// FolgeAlteEpoche gehört zu einem Lauf, den es nicht mehr gibt.
	FolgeAlteEpoche
)

// Fehler macht aus einem Urteil einen Fehler — oder `nil`, wo keiner ist.
//
// **Für den Aufrufer, der nicht selbst entscheiden will.** Wer einen Strom
// verarbeitet, will das Urteil; wer einen einzelnen Befehl schickt, will
// wissen, ob er ankam.
func (u Folgeurteil) Fehler(epoche, folge uint64) error {
	if u == FolgeNeu {
		return nil
	}

	return fmt.Errorf("%w: Epoche %d, Folge %d — %s", ErrConflict, epoche, folge, u)
}

// String gibt dem Urteil einen Namen — für Meldungen und Protokolle.
func (u Folgeurteil) String() string {
	switch u {
	case FolgeNeu:
		return "neu"
	case FolgeSchonDa:
		return "schon da"
	case FolgeLuecke:
		return "Lücke"
	case FolgeAlteEpoche:
		return "alte Epoche"
	}

	return "unbekannt"
}

// Empfaengerstand merkt sich, was zuletzt ankam.
type Empfaengerstand struct {
	epoche uint64
	folge  uint64
}

// NeuerEmpfaengerstand fängt in einer Epoche an, ohne eine Folge gesehen zu
// haben.
func NeuerEmpfaengerstand(epoche uint64) *Empfaengerstand {
	return &Empfaengerstand{epoche: epoche}
}

// Epoche gibt den Lauf, in dem dieser Stand ist.
func (s *Empfaengerstand) Epoche() uint64 { return s.epoche }

// Folge gibt die zuletzt genommene.
func (s *Empfaengerstand) Folge() uint64 { return s.folge }

// Pruefe urteilt, ohne den Stand zu ändern.
func (s *Empfaengerstand) Pruefe(epoche, folge uint64) Folgeurteil {
	switch {
	case epoche < s.epoche:
		return FolgeAlteEpoche

	// **Eine neue Epoche fängt von vorn an.** Der Wirt hat neu angefangen;
	// die Folge beginnt wieder bei eins, und wer die alte weiterzählte, sähe
	// eine Lücke, wo keine ist.
	case epoche > s.epoche:
		if folge == 1 {
			return FolgeNeu
		}

		return FolgeLuecke

	case folge <= s.folge:
		return FolgeSchonDa

	case folge > s.folge+1:
		return FolgeLuecke
	}

	return FolgeNeu
}

// Nimm schreibt den Stand fort — und gibt das Urteil zurück.
//
// **Ein Urteil und kein Fehler**, denn „schon da" ist keiner: eine zweite
// Zustellung ist auf einer Leitung, die wiederholt, der Normalfall. Wer einen
// Fehler will, fragt das Urteil danach ([Folgeurteil.Fehler]).
//
// **Und dieselbe Antwort wie [Empfaengerstand.Pruefe]**, damit der Aufrufer
// nicht beides hintereinander rufen muss: der erste Anlauf tat das, und der
// zweite Aufruf konnte gar nicht mehr klagen — ein Zweig, den kein Lauf je
// beträte (§3).
func (s *Empfaengerstand) Nimm(epoche, folge uint64) Folgeurteil {
	urteil := s.Pruefe(epoche, folge)
	if urteil != FolgeNeu {
		return urteil
	}

	s.epoche = epoche
	s.folge = folge

	return FolgeNeu
}
