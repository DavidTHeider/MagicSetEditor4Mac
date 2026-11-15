package msescript

import "slices"

// Umgebung hält die Namen, die ein Ausdruck sehen kann.
//
// # Warum eine Kette und keine Abbildung
//
// Eine Auswertung legt Namen an (`a := 1`), ohne die Umgebung des Aufrufers zu
// verändern. Zugleich muss sie dessen Namen sehen — und zwar die zum Zeitpunkt
// des **Aufrufs**, nicht zum Zeitpunkt der Definition. Das ist der Kern von
// `docs/FORMATE.md`, Abschnitt „Eine geschweifte Klammer ist ein Funktionswert":
//
//	mana_sort := {
//	    if styling.unsorted_casting_cost then input
//	    else sort_text@(order: "…")()
//	}
//
// `input` steht in keiner Argumentliste und in keiner Zuweisung. Es kann nur
// von dort kommen, wo `mana_sort` **benutzt** wird — MSE setzt es je Feld.
// Deshalb schließt eine geschweifte Klammer nichts ein: sie bekommt ihre
// Umgebung gereicht.
type Umgebung struct {
	namen  map[string]Wert
	darauf *Umgebung

	// meldungen sammelt, was `trace`, `warning` und `error` melden.
	//
	// **Ein Zeiger, damit er vererbt wird**: `Darueber` reicht ihn weiter,
	// und eine Meldung aus einem Funktionsrumpf kommt oben an. Siehe
	// `meldung.go`.
	meldungen *[]Meldung

	// ordnungen merkt die Ordnungen, die `position(…, order_by:)` gebaut hat.
	//
	// **Gesucht wird die Kette hinauf** ([Umgebung.speicherDerOrdnungen]) und
	// nicht kopiert: der Speicher darf auch auf eine Ebene gelegt werden, unter
	// der schon Umgebungen hängen.
	ordnungen *ordnungsspeicher
}

// NeueUmgebung baut eine Umgebung aus benannten Werten.
func NeueUmgebung(namen map[string]Wert) *Umgebung {
	eigene := make(map[string]Wert, len(namen))
	for name, wert := range namen {
		eigene[name] = wert
	}

	return &Umgebung{namen: eigene}
}

// Darueber baut eine Umgebung, die diese sieht und eigene Namen anlegen darf.
//
// **Sie legt einen Meldungssammler an, wenn es noch keinen gibt.** Damit
// sammelt jede Auswertung, die über der Bibliothek beginnt — und wer eine
// eingebaute Funktion von Hand ruft, sammelt weiterhin nichts.
func (u *Umgebung) Darueber() *Umgebung {
	neu := &Umgebung{namen: map[string]Wert{}, darauf: u}

	// **Eine Frage und nicht zwei.** Der erste Anlauf fragte zusätzlich
	// `u.Meldungen() == nil` — dieselbe Frage in anderer Gestalt, und der
	// zweite Zweig war damit unerreichbar. Sichtbar wurde es an der
	// Abdeckung.
	if !u.sammelt() {
		neu.meldungen = &[]Meldung{}
	}

	return neu
}

// sammelt sagt, ob irgendeine Ebene schon einen Sammler hat.
func (u *Umgebung) sammelt() bool {
	for ebene := u; ebene != nil; ebene = ebene.darauf {
		if ebene.meldungen != nil {
			return true
		}
	}

	return false
}

// Suche geht die Kette hinauf.
func (u *Umgebung) Suche(name string) (Wert, bool) {
	for ebene := u; ebene != nil; ebene = ebene.darauf {
		if wert, da := ebene.namen[name]; da {
			return wert, true
		}
	}

	return Nichts, false
}

// Namen gibt die Namen dieser Ebene, unsortiert.
//
// **Für Wächter, nicht für die Auswertung.** Ein Ausdruck fragt nach einem
// Namen, nicht nach allen; wer alle braucht, prüft die Bibliothek selbst —
// etwa darauf, dass sie keine Dateifunktion führt
// (#51).
func (u *Umgebung) Namen() []string {
	namen := make([]string, 0, len(u.namen))
	for name := range u.namen {
		namen = append(namen, name)
	}

	return namen
}

// Setze legt einen Namen auf **dieser** Ebene an.
//
// Nicht dort, wo er schon steht: eine Zuweisung in einem Block änderte sonst
// die Umgebung des Aufrufers, und zwei Felder, die dieselbe Vorlage benutzen,
// beeinflussten einander. Das wäre nicht zu finden.
func (u *Umgebung) Setze(name string, wert Wert) {
	u.namen[name] = wert
}

// grundnamen sind die Namen, die die Sprache selbst mitbringt.
//
// `true` und `false` stehen in `kamen-decade.mse-style/style` Zeile 63 im
// Skript (`then true`). Sie sind **keine** Schlüsselwörter des Lexers, weil
// nichts belegt, dass eine Vorlage sie nicht überschreiben darf — als Namen
// verhalten sie sich richtig, ohne dass eine Regel erfunden werden muss.
var grundnamen = map[string]Wert{
	"true":  Wahrheit(true),
	"false": Wahrheit(false),
}

// AlleNamen gibt die Namen **jeder** Ebene, sortiert und ohne Doppel.
//
// # Warum es das neben [Umgebung.Namen] gibt
//
// `Namen` gibt eine Ebene — das genügt einem Wächter, der die Bibliothek
// prüft. Wer **vervollständigen** will, braucht alles, was ein Ausdruck an
// dieser Stelle nennen darf: die Bibliothek **und** `card`, `set`,
// `stylesheet` darüber
// (#1164).
//
// **Ohne Doppel, und die untere Ebene gewinnt.** Ein Name, den eine Vorlage
// überschreibt, ist derselbe Name; zweimal genannt wäre er eine Auskunft über
// den Aufbau der Umgebung und nicht über die Sprache.
//
// **Sortiert**, damit zwei Aufrufe dieselbe Liste geben. Eine Reihenfolge aus
// einer Map ist bei jedem Lauf eine andere, und eine Vervollständigung, die
// springt, ist schlimmer als keine.
func (u *Umgebung) AlleNamen() []string {
	gesehen := map[string]bool{}

	var namen []string

	for ebene := u; ebene != nil; ebene = ebene.darauf {
		for name := range ebene.namen {
			if !gesehen[name] {
				gesehen[name] = true

				namen = append(namen, name)
			}
		}
	}

	slices.Sort(namen)

	return namen
}
