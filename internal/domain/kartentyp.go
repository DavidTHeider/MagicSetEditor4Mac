package domain

// **Die Typzeile als Daten** — Werkstück M2 der Regelmaschine
// (#1604).
//
// # Wogegen gelesen wurde
//
// CR 205.1 (die Typzeile steht unter dem Bild), CR 205.2a (die **fünfzehn**
// Kartentypen), CR 205.3a (Untertypen stehen hinter dem Gedankenstrich),
// CR 205.4a (die **fünf** Supertypen) — Fassung 2026-08-19.
//
// # Warum die Untertypen eine offene Liste sind
//
// **Weil die CR sie je Kartentyp führt und sie mit jedem Set wächst.**
// CR 205.3 zählt sie in zwanzig Unterabschnitten auf; eine eingefrorene Liste
// hier wäre am Tag des nächsten Sets falsch, und ein Paket, das einen neuen
// erfindet, ist bei MSE der **Normalfall** — dafür gibt es das Programm.
//
// # Und warum ein unbekannter **Kartentyp** trotzdem ein Befund ist
//
// **Weil die Menge geschlossen ist.** CR 205.2a zählt fünfzehn auf, und ein
// sechzehnter ist entweder ein Tippfehler oder eine Fassung, die dieses
// Programm nicht kennt — beides gehört gesagt (§8) und nicht verschluckt.

import "strings"

// Kartentyp ist einer der fünfzehn — CR 205.2a (Fassung 2026-08-19).
type Kartentyp string

const (
	TypArtefakt      Kartentyp = "artifact"
	TypKampf         Kartentyp = "battle"
	TypVerschwoerung Kartentyp = "conspiracy"
	TypKreatur       Kartentyp = "creature"
	TypVerlies       Kartentyp = "dungeon"
	TypVerzauberung  Kartentyp = "enchantment"
	TypSpontan       Kartentyp = "instant"
	TypSippe         Kartentyp = "kindred"
	TypLand          Kartentyp = "land"
	TypPhaenomen     Kartentyp = "phenomenon"
	TypEbene         Kartentyp = "plane"
	TypPlaneswalker  Kartentyp = "planeswalker"
	TypPlan          Kartentyp = "scheme"
	TypHexerei       Kartentyp = "sorcery"
	TypVorhut        Kartentyp = "vanguard"
)

// Kartentypen gibt die fünfzehn in der Reihenfolge von CR 205.2a.
func Kartentypen() []Kartentyp {
	return []Kartentyp{
		TypArtefakt, TypKampf, TypVerschwoerung, TypKreatur, TypVerlies,
		TypVerzauberung, TypSpontan, TypSippe, TypLand, TypPhaenomen,
		TypEbene, TypPlaneswalker, TypPlan, TypHexerei, TypVorhut,
	}
}

// Supertyp ist einer der fünf — CR 205.4a (Fassung 2026-08-19).
type Supertyp string

const (
	SuperGrund     Supertyp = "basic"
	SuperLegendaer Supertyp = "legendary"
	SuperLaufend   Supertyp = "ongoing"
	SuperSchnee    Supertyp = "snow"
	SuperWelt      Supertyp = "world"
)

// Supertypen gibt die fünf in der Reihenfolge von CR 205.4a.
func Supertypen() []Supertyp {
	return []Supertyp{
		SuperGrund, SuperLegendaer, SuperLaufend, SuperSchnee, SuperWelt,
	}
}

// Typzeile ist die gedeutete Zeile unter dem Bild — CR 205.1.
type Typzeile struct {
	Supertypen []Supertyp
	Typen      []Kartentyp

	// Untertypen bleiben Zeichenketten — der Grund steht im Kopf.
	Untertypen []string
}

// Ist sagt, ob diese Zeile diesen Kartentyp führt.
func (t Typzeile) Ist(typ Kartentyp) bool {
	for _, eigener := range t.Typen {
		if eigener == typ {
			return true
		}
	}

	return false
}

// Hat sagt, ob diese Zeile diesen Supertyp führt.
func (t Typzeile) Hat(super Supertyp) bool {
	for _, eigener := range t.Supertypen {
		if eigener == super {
			return true
		}
	}

	return false
}

// trennung ist, womit MSE und die CR Ober- von Untertypen trennen.
//
// **Drei Schreibweisen, und alle drei kommen vor**: der Gedankenstrich der
// CR (U+2014), der kürzere (U+2013) und der Bindestrich, den jemand tippt,
// dem das Zeichen fehlt.
var trennung = []string{"—", "–", " - "}

// TypzeileAus deutet eine Typzeile aus **zwei** Feldern.
//
// # Warum zwei und nicht eines
//
// **Weil MSE sie getrennt führt.** `docs/FORMATE.md` (Zeile 1832 ff.) zählt
// für das Spiel `magic` `super type` **und** `sub type` auf: im M15-Paket
// steht im ersten „Legendary Creature" und im zweiten „Human Wizard" — der
// Strich dazwischen wird beim Zeichnen gesetzt. Daneben gibt es `type` mit der
// ganzen Zeile.
//
// **Ist `unter` leer, wird `ober` am Strich getrennt** (CR 205.3a). Damit
// nimmt dieselbe Funktion beide Gestalten entgegen, und der Aufrufer muss
// nicht wissen, welches Feld sein Paket gefüllt hat.
//
// **Zwei Rückgaben und kein Fehler**: eine Zeile mit einem unbekannten Wort
// ist lesbar — CR 205.2a kennt es nicht, der Rest der Karte steht trotzdem da.
// Das ist §8 in seiner gewöhnlichen Gestalt.
func TypzeileAus(ober, unter string) (Typzeile, []Befund) {
	oben, unten := ober, unter
	if strings.TrimSpace(unten) == "" {
		oben, unten = zerteileTypzeile(ober)
	}

	zeile := Typzeile{}

	var befunde []Befund

	for _, wort := range strings.Fields(oben) {
		deuteOberwort(&zeile, &befunde, wort, ober)
	}

	zeile.Untertypen = append(zeile.Untertypen, strings.Fields(unten)...)

	return zeile, befunde
}

// zerteileTypzeile trennt an dem Strich, den CR 205.3a verlangt.
func zerteileTypzeile(roh string) (oben, unten string) {
	for _, strich := range trennung {
		if vor, nach, gefunden := strings.Cut(roh, strich); gefunden {
			return vor, nach
		}
	}

	return roh, ""
}

// deuteOberwort trägt ein Wort vor dem Strich ein.
func deuteOberwort(
	zeile *Typzeile, befunde *[]Befund, wort, roh string,
) {
	klein := strings.ToLower(wort)

	for _, super := range Supertypen() {
		if string(super) == klein {
			zeile.Supertypen = append(zeile.Supertypen, super)

			return
		}
	}

	for _, typ := range Kartentypen() {
		if string(typ) == klein {
			zeile.Typen = append(zeile.Typen, typ)

			return
		}
	}

	*befunde = append(*befunde, Befund{
		Datei: roh, Ort: wort, Art: ArtVorlage,
		Meldung: "CR 205.2a kennt diesen Kartentyp nicht — " +
			"weder als Kartentyp noch als Supertyp (CR 205.4a)",
	})
}

// TypzeileText baut aus zwei Feldern eines MSE-Pakets die Zeile, die
// CR 205.3a beschreibt — Obertypen, Gedankenstrich, Untertypen.
//
// # Wofür es sie gibt
//
// **Weil die Höchstzahl gleicher Karten nach der Zeile fragt** und nicht nach
// zwei Feldern: [IstStandardland] schneidet am Gedankenstrich (CR 205.4a) und
// findet ohne ihn den falschen Teil. Ohne diese Stelle baute jeder Aufrufer
// die Zeile selbst — und der erste mit einem Bindestrich
// (#1622).
//
// **Der Gedankenstrich ist derselbe wie bei Scryfall** (U+2014), damit eine
// gebaute Zeile und eine geholte dieselbe Gestalt haben.
//
// **Leer bleibt leer**: stehen beide Felder leer, entsteht keine Zeile aus
// einem einsamen Strich — „nicht bekannt" ist eine Auskunft, „—" wäre keine.
func TypzeileText(ober, unter string) string {
	oben, unten := strings.TrimSpace(ober), strings.TrimSpace(unter)

	switch {
	case oben == "" && unten == "":
		return ""

	case unten == "":
		return oben
	}

	return strings.TrimSpace(oben + " " + gedankenstrich + " " + unten)
}
