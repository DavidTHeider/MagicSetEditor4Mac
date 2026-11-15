// **Die Regeln als Baum und nicht als Liste** — PLAY-02.AK-02
// (#1710):
// „Ein Regelbaum gruppiert Optionen nach Spielablauf und Mechanik;
// notwendige Abhängigkeiten sind sichtbar, unvereinbare Kombinationen
// werden erklärt."
//
// # Warum eine Gruppierung und nicht dreizehn Zeilen
//
// **Weil dreizehn gleichrangige Namen keine Ordnung sind.** Die
// [Regelmatrix] zählt sie auf; wer sie **bedienen** soll, braucht die Frage
// dahinter: geht es um den **Ablauf** einer Partie — wann bin ich dran, was
// geschieht von selbst — oder um die **Mechanik** einzelner Objekte?
//
// # Warum man nur abwählen kann und nicht zuwählen
//
// **Weil Zuwählen eine Zusage wäre, die niemand halten kann.** Was dieses
// Programm durchsetzt, sagt die Matrix; ein Schalter, der einen Bereich auf
// „automatisch" stellte, ohne dass Code dahinter stünde, wäre die Gestalt
// aus §10 — eine Behauptung, die wie ein Beleg aussieht. Abwählen dagegen
// ist immer einlösbar: **das Programm hält sich heraus, und die Runde macht
// es selbst.**
//
// # Und warum eine Abwahl nichts aufhebt
//
// **PLAY-02.AK-04, wörtlich**: „Abschaltungen heben Integritäts- und
// Sicherheitsprüfungen nicht auf." Eine abgewählte Regel wird zu
// [UnterstuetzungManuell] — das ist dasselbe, was ein Bereich ohnehin ist,
// den dieses Programm nicht kennt. Die Invarianten (siehe
// [InvarianteSichtbarkeit]) hängen an keiner dieser Zahlen.

package domain

import (
	"fmt"
	"slices"
)

// Regelgruppenkennung ist ein Ast des Baums.
type Regelgruppenkennung = string

// Die zwei Äste.
//
// **Zwei und nicht dreizehn Einzelne**: mehr Äste wären eine feinere
// Ordnung, die niemand gelesen hätte — und weniger wäre die Liste von
// vorher.
const (
	// GruppeAblauf — wann etwas geschieht und wer handeln darf.
	GruppeAblauf Regelgruppenkennung = "ablauf"

	// GruppeMechanik — was an einem einzelnen Objekt gilt.
	GruppeMechanik Regelgruppenkennung = "mechanik"
)

// Regelgruppe ist ein Ast samt seinen Bereichen.
type Regelgruppe struct {
	Kennung  Regelgruppenkennung
	Bereiche []Regelbereich
}

// Regelbaum gibt die Bereiche in ihren zwei Gruppen.
//
// **Die Reihenfolge innerhalb einer Gruppe ist die des Spiels** und nicht
// die des Alphabets: eine Partie fängt mit Zonen an, dann kommt der Zug,
// dann, wer dran ist.
func Regelbaum() []Regelgruppe {
	return []Regelgruppe{
		{
			Kennung: GruppeAblauf,
			Bereiche: []Regelbereich{
				BereichZonen, BereichZugstruktur, BereichPrioritaet,
				BereichStapel, BereichCombat, BereichZustandsbasiert,
			},
		},
		{
			Kennung: GruppeMechanik,
			Bereiche: []Regelbereich{
				BereichTypen, BereichMarken, BereichAnhaengsel,
				BereichSpielsteine, BereichMana, BereichWechselwirkungen,
				BereichCommander,
			},
		},
	}
}

// Regelabhaengigkeit sagt, dass ein Bereich ohne einen anderen nicht geht.
//
// **Mit Quelle**, und das ist der ganze Unterschied zu einer Meinung: jede
// dieser sechs Zeilen steht auf einer Regelnummer, die dieses Projekt
// nachgelesen und in `docs/REGELABDECKUNG.md` belegt hat (§11).
type Regelabhaengigkeit struct {
	// Bereich geht nicht ohne Braucht.
	Bereich Regelbereich
	Braucht Regelbereich

	Quelle Regelquelle
}

// Regelabhaengigkeiten gibt die sechs.
//
// **Sechs und nicht dreizehn**: die meisten Bereiche stehen für sich. Eine
// Abhängigkeit wird hier nur eingetragen, wo eine **Regel** sie verlangt —
// nicht, wo zwei Dinge miteinander zu tun haben.
func Regelabhaengigkeiten() []Regelabhaengigkeit {
	ausCR := func(nummern ...string) Regelquelle {
		return Regelquelle{
			Dokument: DokumentCR,
			Fassung:  Grundprofilquelle,
			Nummern:  nummern,
		}
	}

	return []Regelabhaengigkeit{
		// **Ein Abschnitt endet, wenn der Stapel leer ist und alle
		// nacheinander passen** — CR 500.2. Ohne Zugstruktur gibt es den
		// Abschnitt nicht, in dem Priorität überhaupt wandert.
		{BereichPrioritaet, BereichZugstruktur, ausCR("500")},

		// **Das oberste Objekt verrechnet sich, wenn alle nacheinander
		// passen** — CR 608.1. Ein Stapel ohne Priorität verrechnete nie.
		{BereichStapel, BereichPrioritaet, ausCR("608")},

		// **Geprüft wird, wann immer jemand Priorität bekäme** — CR 704.3.
		{BereichZustandsbasiert, BereichPrioritaet, ausCR("704")},

		// **Der Kampf ist ein Abschnitt des Zuges** — CR 506.1.
		{BereichCombat, BereichZugstruktur, ausCR("506")},

		// **Der Vorrat leert sich am Ende jedes Abschnitts** — CR 500.5.
		{BereichMana, BereichZugstruktur, ausCR("106", "500")},

		// **Commander braucht die Kommandozone** — CR 903.
		{BereichCommander, BereichZonen, ausCR("903")},
	}
}

// Regelwahl ist, was eine Runde **nicht** durchsetzen lassen will.
//
// **Eine Abwahl und keine Zuwahl** — der Grund steht im Kopf dieser Datei.
type Regelwahl struct {
	Abgewaehlt []Regelbereich
}

// Pruefe weist ab, was gar kein Bereich ist.
//
// **Ein unbekannter Name ist ein Fehler und kein stilles Übergehen** (§8):
// wer sich vertippt, hätte sonst eine Runde nach Regeln gespielt, die er
// abgewählt zu haben glaubte.
func (w Regelwahl) Pruefe() error {
	for _, bereich := range w.Abgewaehlt {
		if !IstRegelbereich(bereich) {
			return fmt.Errorf(
				"%w: %q ist kein Regelbereich", ErrValidation, bereich)
		}
	}

	return nil
}

// Regelwiderspruch ist eine Abwahl, an der noch etwas hängt.
type Regelwiderspruch struct {
	// Bereich ist der abgewählte.
	Bereich Regelbereich

	// Haengt ist der, der ihn braucht und **nicht** abgewählt wurde.
	Haengt Regelbereich

	Quelle Regelquelle
}

// Widersprueche nennt jede Abwahl, an der noch etwas hängt.
//
// **Erklärt statt abgewiesen** (AK-02: „unvereinbare Kombinationen werden
// erklärt"): wer die Priorität abwählt und den Stapel behält, bekommt den
// Satz dazu — und kann dann entscheiden, ob er beides abwählt oder keines.
//
// **Eine leere Liste und kein `nil`**: „nichts widerspricht sich" ist eine
// Antwort, und sie sieht anders aus als keine.
func (w Regelwahl) Widersprueche() []Regelwiderspruch {
	widersprueche := []Regelwiderspruch{}

	for _, abhaengig := range Regelabhaengigkeiten() {
		if slices.Contains(w.Abgewaehlt, abhaengig.Braucht) &&
			!slices.Contains(w.Abgewaehlt, abhaengig.Bereich) {
			widersprueche = append(widersprueche, Regelwiderspruch{
				Bereich: abhaengig.Braucht,
				Haengt:  abhaengig.Bereich,
				Quelle:  abhaengig.Quelle,
			})
		}
	}

	return widersprueche
}

// Wirksam gibt die Matrix, die nach dieser Wahl gilt.
//
// **Abgewählt heißt [UnterstuetzungManuell]** und nicht `unbekannt`: der
// Unterschied ist, dass jemand **entschieden** hat, es selbst zu machen.
//
// **Die Vorlage bleibt unberührt** — wer hier änderte, änderte das Profil
// jeder anderen Partie mit.
func (w Regelwahl) Wirksam(matrix Regelmatrix) Regelmatrix {
	gewandelt := Regelmatrix{
		Eintraege: make([]Matrixeintrag, 0, len(matrix.Eintraege)),
	}

	for _, eintrag := range matrix.Eintraege {
		if slices.Contains(w.Abgewaehlt, eintrag.Bereich) {
			eintrag.Grad = UnterstuetzungManuell
			eintrag.Quelle = Regelquelle{}

			// **Die Grenzen gehen mit weg** — sie sagen, was von einer
			// **Zusage** nicht gedeckt ist. Ohne Zusage sind sie eine
			// Auskunft über nichts.
			eintrag.Grenzen = nil
		}

		gewandelt.Eintraege = append(gewandelt.Eintraege, eintrag)
	}

	return gewandelt
}
