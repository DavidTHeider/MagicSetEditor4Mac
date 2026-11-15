package domain

// **Was ein Standardland ist** —
// #1370,
// Nacharbeit zu Werkstück 5 von Phase 4.6.
//
// # Warum es diese Datei gibt
//
// Weil die Höchstzahl gleicher Karten sie ausnimmt, und zwar in **beiden**
// nachgelesenen Bauarten:
//
//   - CR 100.2a: ein Constructed-Deck trägt **Standardländer beliebig oft**
//     und jede andere Karte eines englischen Namens höchstens viermal —
//     **die Ausnahme gilt nur für Standardländer**.
//   - CR 903.5b: **außer Standardländern** trägt jede Karte eines
//     Commanderdecks einen anderen englischen Namen.
//
// Ohne diese Ausnahme bekäme jedes gewöhnliche Deck mit vierundzwanzig Wald
// einen Verstoß gemeldet.
//
// *Magic: The Gathering Comprehensive Rules*, in Kraft seit dem
// 2026-08-07, bezogen am 2026-09-01 von `media.wizards.com`.

import "strings"

// IstStandardland sagt, ob eine Typzeile ein Standardland beschreibt.
//
// # Die Regel, wörtlich nachgelesen
//
// CR 205.4c: ein Land mit dem Supertyp „basic" ist ein Standardland; jedes
// Land ohne diesen Supertyp ist keines — **auch dann nicht, wenn es einen
// Landgrundtyp trägt**. — der zweite Satz ist der ganze Grund für die Sorgfalt:
// „Dryad Arbor" trägt `Forest` als Untertyp und ist trotzdem keines.
//
// CR 205.4a: die Supertypen stehen unmittelbar vor den Kartentypen.
// — deshalb wird **vor** dem Gedankenstrich gesucht und nicht in der ganzen
// Zeile.
//
// # Was hier nicht behauptet wird
//
// **Dass eine leere Typzeile etwas bedeutet.** Sie heißt „nicht bekannt", und
// Unbekanntes wird gezählt, nicht ausgenommen — eine Ausnahme auf Verdacht
// ließe ein Deck mit fünf gleichen Nicht-Ländern durchgehen.
//
// # Gemessen
//
// Am 2026-09-01 an `api.scryfall.com/cards/named`: `Forest` gibt
// `Basic Land — Forest`, `Wastes` gibt `Basic Land` (ohne Untertyp),
// `Snow-Covered Forest` gibt `Basic Snow Land — Forest`, und `Dryad Arbor`
// gibt `Land Creature — Forest Dryad`.
func IstStandardland(typzeile string) bool {
	vorne, _, _ := strings.Cut(typzeile, gedankenstrich)

	woerter := strings.Fields(strings.ToLower(vorne))

	var basic, land bool

	for _, wort := range woerter {
		switch wort {
		case "basic":
			basic = true
		case "land":
			land = true
		}
	}

	return basic && land
}

// gedankenstrich trennt Ober- von Untertypen — derselbe wie bei Scryfall.
const gedankenstrich = "—"

// Deckposten ist ein Platz samt dem, was über ihn bekannt ist.
//
// # Warum die Prüfung nicht mehr das Deck allein bekommt
//
// Weil sie ohne die Typen falsch rechnet. Ob eine Karte ein Standardland ist,
// steht in ihrem Typ — und den kennt dieses Paket nicht: ein [Deckplatz]
// verweist nur, er kopiert nichts. Wer prüft, bringt die Auskunft mit oder
// sagt mit [Posten] ausdrücklich, dass er keine hat.
type Deckposten struct {
	Platz Deckplatz

	// Standardland sagt, ob CR 100.2a und CR 903.5b diese Karte von der
	// Höchstzahl ausnehmen.
	Standardland bool

	// Name ist der **englische Kartenname** — leer, wenn er nicht bekannt ist.
	//
	// # Warum die Höchstzahl ihn braucht
	//
	// **CR 100.2a begrenzt Karten mit einem Namen, nicht Drucke**: höchstens
	// vier je **englischem Namen**, Standardländer ausgenommen. CR 903.5b
	// sagt dasselbe für Commander: jede Karte eines Commanderdecks trägt
	// einen **anderen englischen Namen**.
	//
	// **Der erste Anlauf zählte den Druck** — vier verschiedene Drucke von
	// „Lightning Bolt" waren vier Gruppen zu je einem und kamen durch
	// (Befund 1 des Zweitblicks über Phase 4.6).
	//
	// **Leer heißt: nicht bekannt**, und dann entscheidet der Druck. Zwei
	// namenlose Drucke als dieselbe Karte zu zählen wäre eine Erfindung; sie
	// getrennt zu zählen ist das, was man weiß.
	Name string

	// Orakelkennung ist Scryfalls `oracle_id` — leer, wenn sie nicht bekannt
	// ist.
	//
	// # Wozu sie neben dem Namen steht
	//
	// **CR 100.2a hört nicht bei „no more than four" auf.** Der letzte Satz
	// lautet: für den Deckbau gelten Karten mit austauschbaren
	// Namen als derselbe englische Name (CR 201.3). Und
	// CR 201.3b sagt dasselbe für die Formatzulässigkeit.
	//
	// Gemeint sind die Paare aus „Universes Beyond" und „Universes Within":
	// **dieselbe Karte, zwei englische Namen.** Nach dem Namen allein gingen
	// vier von einem und vier vom anderen durch — acht derselben Karte
	// (#1375).
	//
	// **Scryfall verbindet sie über `oracle_id`**, nachgemessen am 2026-09-04:
	//
	//	Mind Flayer, the Shadow  ↔ Arvinox, the Mind Flail        9abecd23…
	//	Eleven, the Mage         ↔ Cecily, Haunted Mage           a726c27d…
	//	Dustin, Gadget Genius    ↔ Hargilde, Kindly Runechanter   747009e7…
	//
	// **Leer heißt: nicht gefragt oder nicht bekannt**, und dann entscheidet
	// weiter der Name. Ein Deck aus eigenen Entwürfen fragt Scryfall gar
	// nicht; dort wäre eine erzwungene Kennung eine Erfindung.
	Orakelkennung string
}

// Posten macht aus einem Deck Posten **ohne** Auskunft über die Typen.
//
// **Ein eigener Name und keine Vorgabe**: wer ihn ruft, sagt damit, dass er
// nichts über Standardländer weiß — und bekommt sie mitgezählt. Ein
// stillschweigender Vorgabewert wäre derselbe Mangel, den
// #1370
// behoben hat, nur schwerer zu finden.
func Posten(deck Deck) []Deckposten {
	posten := make([]Deckposten, 0, len(deck.Plaetze))

	for _, platz := range deck.Plaetze {
		posten = append(posten, Deckposten{Platz: platz})
	}

	return posten
}
