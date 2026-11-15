package domain

// **Was auf einem Kartenbild zu sehen ist** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Die drei Sätze, für die es diesen Typ gibt
//
// Das Issue verlangt: „Vision/Kartenerkennung liefert mögliche Identitäten
// **mit Unsicherheit** und **expliziter Zuordnung** über A2; Menge/Exemplar
// wird **nicht automatisch** angelegt." Und weiter: „OCR/Layout/Face/Set/
// Kunstwerk als **getrennte Evidenz** behandeln, unsichere Druckidentität
// bleibt **Auswahl des Nutzers**."
//
// # Und warum die Unsicherheit keine Zahl ist
//
// **Weil sie zählbar ist.** „Das Modell hat »Lightning Bolt« gelesen, und
// darauf passen 72 Drucke" ist eine Tatsache; „87 % sicher" ist eine
// Rechenausgabe, die wie eine Messung aussieht. Dieselbe Überlegung wie bei
// der Balanceanalyse ([0108](../../docs/entscheidungen/0108-eine-balanceanalyse-hat-keine-zahl.md)),
// und hier trägt sie weiter: die Zahl der passenden Drucke **kommt aus der
// Suche**, nicht aus dem Modell.
//
// # Was hier nicht steht
//
// **Kein Kandidat.** Die Kandidaten kommen aus A2 — derselbe Weg, den eine
// getippte Kartenliste geht. Was das Modell beisteuert, sind **Beobachtungen**,
// und die sind der Suchbegriff, nicht das Ergebnis.

import "fmt"

// Evidenzart ist, **woran** das Modell etwas erkannt haben will.
//
// **Getrennt und nicht verrechnet**: ein gelesener Name und ein erkanntes
// Setsymbol sind zwei Beobachtungen mit verschiedener Verlässlichkeit, und
// wer sie zu einer Zahl verrührt, kann hinterher nicht sagen, woran es lag.
type Evidenzart string

const (
	// EvidenzName ist der gelesene Kartenname — die stärkste Evidenz, weil
	// er als Text auf der Karte steht.
	EvidenzName Evidenzart = "name"

	// EvidenzSetcode ist das gelesene oder am Symbol erkannte Set.
	EvidenzSetcode Evidenzart = "setcode"

	// EvidenzNummer ist die Sammlernummer aus der Fußzeile.
	EvidenzNummer Evidenzart = "nummer"

	// EvidenzTypzeile ist die gelesene Typzeile.
	EvidenzTypzeile Evidenzart = "typzeile"

	// EvidenzLayout ist die **Bauart** — normal, geteilt, Saga, beidseitig.
	//
	// **Sie identifiziert keine Karte** und steht trotzdem hier: sie
	// widerlegt Kandidaten, und ein Widerspruch ist eine Auskunft.
	EvidenzLayout Evidenzart = "layout"

	// EvidenzKunstwerk ist, was auf der Illustration zu sehen ist.
	//
	// **Die schwächste**: dieselbe Illustration erscheint über viele Drucke,
	// und eine Beschreibung ist keine Kennung. Sie steht hier, weil sie bei
	// einer abgeschnittenen Karte das Einzige sein kann, was bleibt.
	EvidenzKunstwerk Evidenzart = "kunstwerk"
)

// Evidenzarten nennt die sechs in fester Ordnung — stärkste zuerst.
func Evidenzarten() []Evidenzart {
	return []Evidenzart{
		EvidenzName, EvidenzSetcode, EvidenzNummer,
		EvidenzTypzeile, EvidenzLayout, EvidenzKunstwerk,
	}
}

// IstEvidenzart sagt, ob es diese Art gibt.
func IstEvidenzart(art Evidenzart) bool {
	for _, eine := range Evidenzarten() {
		if eine == art {
			return true
		}
	}

	return false
}

// Beobachtung ist **eine** Sache, die das Modell auf dem Bild gelesen hat.
//
// # Warum „gelesen" und nicht „erkannt"
//
// **Weil der Unterschied der Punkt ist.** Ein Modell, das den Namen liest,
// gibt einen Text wieder, der dasteht; eines, das aus der Illustration auf
// eine Karte schließt, rät. Beides darf vorkommen — aber nicht unter
// derselben Überschrift.
type Beobachtung struct {
	Art Evidenzart

	// Wert ist, was das Modell gelesen hat — **roh**, so wie es dastand.
	Wert string

	// Lesbar sagt, ob die Stelle überhaupt zu sehen war.
	//
	// **Ein eigener Zustand neben dem leeren Wert**: „die Fußzeile ist
	// abgeschnitten" und „dort steht nichts" sind zwei Auskünfte, und nur die
	// erste erklärt, warum ein Kandidat offenbleibt.
	Lesbar bool
}

// BefundOhneBeobachtung steht für ein Bild, zu dem das Modell nichts
// gelesen hat.
const BefundOhneBeobachtung = "ohnebeobachtung"

// BefundErfundeneEvidenz steht für eine Beobachtung einer Art, die es
// nicht gibt.
const BefundErfundeneEvidenz = "erfundeneevidenz"

// BefundWertOhneLesbarkeit steht für einen Wert an einer Stelle, die
// angeblich unlesbar war.
//
// **Das ist ein Widerspruch und kein Formfehler**: wer sagt, die Fußzeile sei
// abgeschnitten, und trotzdem eine Nummer nennt, hat sie nicht gelesen,
// sondern geraten.
const BefundWertOhneLesbarkeit = "wertohnelesbarkeit"

// Bilderkennungsfelder nennt die Felder, die eine Bilderkennung trägt.
//
// **Damit ein Lauf über sie gehen kann** — die Zusage „keine Zahl, die wie
// eine Messung aussieht" ist eine über die **Gestalt**. Wer ein Feld
// hinzufügt, trägt es hier ein und stößt dabei auf die Frage, ob es eine
// erfundene Sicherheit ist.
func Bilderkennungsfelder() []string {
	return []string{"Beobachtungen", "NichtGelesen"}
}

// Bilderkennung ist, was ein Modell zu einem Kartenbild gesagt hat.
//
// **Ohne Kandidaten und ohne Kennung.** Wer daraus einen Druck macht, ist die
// Suche aus A2 — und die Auswahl trifft ein Mensch.
type Bilderkennung struct {
	Beobachtungen []Beobachtung

	// NichtGelesen sind die Arten, zu denen das Modell **nichts** sagen
	// konnte.
	//
	// **Sie stehen ausdrücklich da**, weil ihr Fehlen sonst wie ein
	// Versehen aussieht: „zur Sammlernummer steht nichts" ist eine Antwort,
	// „die Liste ist kürzer" ist keine.
	NichtGelesen []Evidenzart
}

// PruefeBilderkennung hält, was ein Modell über ein Bild sagt, gegen die Form.
//
// **Sie prüft nicht, ob es stimmt** — das kann niemand hier. Sie prüft, dass
// die Auskunft **überhaupt eine ist**: eine Art, die es gibt, und kein Wert
// an einer Stelle, die angeblich nicht zu sehen war.
func PruefeBilderkennung(erkennung Bilderkennung) []Belegbefund {
	befunde := make([]Belegbefund, 0, len(erkennung.Beobachtungen))

	if len(erkennung.Beobachtungen) == 0 {
		return append(befunde, Belegbefund{
			Art: BefundOhneBeobachtung,
			Meldung: "Zu diesem Bild steht keine einzige Beobachtung. Ohne " +
				"eine gelesene Stelle gibt es nichts zu suchen.",
		})
	}

	for _, beobachtung := range erkennung.Beobachtungen {
		if !IstEvidenzart(beobachtung.Art) {
			befunde = append(befunde, Belegbefund{
				Art: BefundErfundeneEvidenz, Stelle: string(beobachtung.Art),
				Meldung: fmt.Sprintf(
					"%q ist keine Evidenzart. Es gibt nur die sechs aus "+
						"»Evidenzarten«.", string(beobachtung.Art)),
			})

			continue
		}

		if !beobachtung.Lesbar && beobachtung.Wert != "" {
			befunde = append(befunde, Belegbefund{
				Art: BefundWertOhneLesbarkeit, Stelle: string(beobachtung.Art),
				Meldung: fmt.Sprintf(
					"Zu %q steht ein Wert, obwohl die Stelle als unlesbar "+
						"gemeldet wurde — das ist geraten und nicht gelesen.",
					string(beobachtung.Art)),
			})
		}
	}

	return befunde
}

// Suchbegriff macht aus den Beobachtungen die Frage an die Suche.
//
// # Warum nur die lesbaren mit Wert
//
// **Weil eine unlesbare Stelle keinen Suchbegriff hergibt.** Sie steht in der
// Antwort, damit jemand sieht, **warum** der Kandidat offenblieb — aber sie
// darf die Suche nicht verengen.
//
// # Und warum das Kunstwerk nicht mitgeht
//
// **Weil es kein Feld der Suche ist.** Scryfall sucht über Namen, Set und
// Nummer; eine Bildbeschreibung dorthin zu geben, machte aus jeder Suche eine
// Volltextsuche über den Regeltext — und die fände etwas anderes, als sie
// verspricht.
func (b Bilderkennung) Suchbegriff() string {
	begriff := ""

	for _, art := range []Evidenzart{EvidenzName, EvidenzSetcode, EvidenzNummer} {
		for _, beobachtung := range b.Beobachtungen {
			if beobachtung.Art != art || !beobachtung.Lesbar ||
				beobachtung.Wert == "" {
				continue
			}

			if begriff != "" {
				begriff += " "
			}

			begriff += beobachtung.Wert
		}
	}

	return begriff
}
