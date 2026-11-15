package domain

// **Das Urteil über ein Deck** — Werkstück 5 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Zwei Fragen, die nichts miteinander zu tun haben
//
// **Passt die Zahl?** — das rechnet dieses Paket aus, aus dem Deck und den
// Zahlen des Formats. Es braucht dafür niemanden.
//
// **Darf diese Karte hier liegen?** — das weiß nur Scryfall, und zwar je
// **Druck**. Was von dort kommt, steht als Wort in einem [Kartenurteil]; dieses
// Paket wertet es aus, holt es aber nicht.
//
// # Und die Zusage, für die es den Vermerk gibt
//
// **`ungeprüft` ist kein `nicht zulässig`.** Eine selbst entworfene Karte ist
// naturgemäß in keinem Format der Welt gelistet; sie deshalb rot zu färben
// hieße, jedem eigenen Deck zu sagen, es sei kaputt.

import (
	"fmt"
	"slices"
	"strings"
)

// Zulaessigkeit ist die Antwort auf „darf diese Karte in dieses Format?".
//
// **Vier Wörter von Scryfall und eines von hier.** Die vier sind nicht
// übersetzt — eine Übersetzung machte aus einer Auskunft eine Auslegung.
type Zulaessigkeit string

const (
	// ZulaessigLegal ist Scryfalls `legal`.
	ZulaessigLegal Zulaessigkeit = "legal"

	// ZulaessigNichtGelistet ist Scryfalls `not_legal` — der Druck ist in
	// diesem Format nicht spielbar.
	ZulaessigNichtGelistet Zulaessigkeit = "not_legal"

	// ZulaessigEingeschraenkt ist Scryfalls `restricted`.
	//
	// **Er macht ein Deck nicht unzulässig.** Wie oft eine eingeschränkte
	// Karte vorkommen darf, steht in den Turnierregeln — und die hat dieses
	// Programm nicht gelesen. Es meldet den Zustand und erfindet keine Zahl.
	ZulaessigEingeschraenkt Zulaessigkeit = "restricted"

	// ZulaessigVerboten ist Scryfalls `banned`.
	ZulaessigVerboten Zulaessigkeit = "banned"

	// ZulaessigUngeprueft heißt: dieses Programm weiß es nicht.
	//
	// **Das ist eine Antwort und kein Fehlen einer.** Sie steht für eine
	// eigene Karte ohne Verweis, für ein Format ohne Scryfallnamen und für
	// einen Druck, den Scryfall zu diesem Format nicht führt.
	ZulaessigUngeprueft Zulaessigkeit = "ungeprüft"
)

// Kartenurteil ist die Antwort für **einen** Platz im Deck.
type Kartenurteil struct {
	PlatzID int64

	// Bezeichnung ist, wie die Karte heißt — soweit bekannt.
	Bezeichnung string

	Menge   int
	Bereich Deckbereich

	Zulaessigkeit Zulaessigkeit

	// Grund sagt, **warum** — vor allem bei `ungeprüft`.
	Grund string

	// Quelle nennt, woher die Antwort kommt; leer bei `ungeprüft`.
	Quelle string
}

// Groessenbefund ist ein Verstoß gegen eine Zahl des Formats.
type Groessenbefund struct {
	// Feld ist `hauptdeck`, `sideboard` oder `gleiche`.
	Feld    string
	Meldung string
}

// Deckurteil ist das ganze Ergebnis einer Formatprüfung.
type Deckurteil struct {
	Formatname string

	Hauptdeck int
	Sideboard int

	Urteile  []Kartenurteil
	Groessen []Groessenbefund
}

// Zulaessig sagt, ob nichts dagegen steht.
//
// **`ungeprüft` und `restricted` stehen nicht dagegen.** Warum, steht an den
// beiden Konstanten.
func (d Deckurteil) Zulaessig() bool {
	if len(d.Groessen) > 0 {
		return false
	}

	for _, urteil := range d.Urteile {
		if urteil.Zulaessigkeit == ZulaessigVerboten ||
			urteil.Zulaessigkeit == ZulaessigNichtGelistet {
			return false
		}
	}

	return true
}

// Ungeprueft gibt die Zahl der **Karten**, über die nichts bekannt ist.
//
// **Karten, nicht Plätze** — dieselbe Zählweise wie überall sonst im Deck.
//
// **Ein Wort, das dieses Programm nicht kennt, zählt mit** (Befund 3 des
// Zweitblicks über Phase 4.6). Scryfall darf ein sechstes schicken; der
// Parser meldet es als Befund und reicht es durch, und ohne diese Zeile
// stünde am Kopf „null ungeprüft" über Karten, über die nichts bekannt ist.
func (d Deckurteil) Ungeprueft() int {
	gezaehlt := 0

	for _, urteil := range d.Urteile {
		if !Gedeutet(urteil.Zulaessigkeit) {
			gezaehlt += urteil.Menge
		}
	}

	return gezaehlt
}

// scryfallwoerter sind die vier, die eine **Aussage** tragen.
//
// **`ungeprüft` steht nicht dabei**, und das ist der Punkt: es ist die
// Abwesenheit einer Aussage und kein fünftes Wort mit eigener Bedeutung.
var scryfallwoerter = []Zulaessigkeit{
	ZulaessigLegal, ZulaessigNichtGelistet,
	ZulaessigEingeschraenkt, ZulaessigVerboten,
}

// Gedeutet sagt, ob dieses Programm aus einem Zulässigkeitswort eine Aussage
// machen kann.
//
// **Alles andere zählt als ungeprüft** — `ungeprüft` selbst und jedes Wort,
// das Scryfall künftig hinzufügt. Ein unbekanntes Wort ist eine Bedeutung, die
// hier fehlt, und keine Freigabe.
func Gedeutet(zulaessigkeit Zulaessigkeit) bool {
	return slices.Contains(scryfallwoerter, zulaessigkeit)
}

// PruefeGroessen rechnet ein Deck gegen die Zahlen eines Formats.
//
// **Null ist keine Grenze** — bei drei der vier Zahlen. Die vierte ist das
// Sideboard: dort heißt null **keines**, denn CR 903.5e sagt: Commanderpartien
// kennen kein Sideboard, und ein zusätzliches Feld nur dafür wäre eines zu
// viel.
//
// **Sie bekommt [Deckposten] und nicht das Deck**, seit
// #1370: ohne
// die Auskunft, was ein Standardland ist, rechnet sie falsch.
func PruefeGroessen(posten []Deckposten, format Format) []Groessenbefund {
	var (
		befunde   []Groessenbefund
		hauptdeck = zaehlePosten(posten, BereichHauptdeck)
		sideboard = zaehlePosten(posten, BereichSideboard)
	)

	if format.Mindestgroesse > 0 && hauptdeck < format.Mindestgroesse {
		befunde = append(befunde, Groessenbefund{"hauptdeck", fmt.Sprintf(
			"Das Hauptdeck hat %d Karten; %q verlangt mindestens %d.",
			hauptdeck, format.Name, format.Mindestgroesse)})
	}

	if format.Hoechstgroesse > 0 && hauptdeck > format.Hoechstgroesse {
		befunde = append(befunde, Groessenbefund{"hauptdeck", fmt.Sprintf(
			"Das Hauptdeck hat %d Karten; %q erlaubt höchstens %d.",
			hauptdeck, format.Name, format.Hoechstgroesse)})
	}

	// **Ein offenes Sideboard beanstandet nichts** — CR 100.4b kennt dort
	// keine Höchstzahl (Befund 2 des Zweitblicks über Phase 4.6).
	if !format.SideboardOffen && sideboard > format.Sideboardgroesse {
		befunde = append(befunde, sideboardbefund(sideboard, format))
	}

	return append(befunde, gleichebefunde(posten, format)...)
}

// zaehlePosten zählt die **Karten** eines Bereichs — mit ihrer Menge.
func zaehlePosten(posten []Deckposten, bereich Deckbereich) int {
	gezaehlt := 0

	for _, einer := range posten {
		if einer.Platz.Bereich == bereich {
			gezaehlt += einer.Platz.Menge
		}
	}

	return gezaehlt
}

// sideboardbefund sagt, was mit dem Sideboard nicht stimmt.
//
// **Zwei Sätze für zwei Lagen**: „zu viele" und „hier gibt es gar keines".
// Ein Format mit `sideboardgroesse: 0` erlaubt keines — bei Commander steht
// das ausdrücklich in CR 903.5e —, und „höchstens 0" wäre dafür ein Satz, den
// niemand versteht.
func sideboardbefund(sideboard int, format Format) Groessenbefund {
	if format.Sideboardgroesse == 0 {
		return Groessenbefund{"sideboard", fmt.Sprintf(
			"Das Sideboard hat %d Karten; %q kennt kein Sideboard.",
			sideboard, format.Name)}
	}

	return Groessenbefund{"sideboard", fmt.Sprintf(
		"Das Sideboard hat %d Karten; %q erlaubt höchstens %d.",
		sideboard, format.Name, format.Sideboardgroesse)}
}

// gleichebefunde meldet, wovon zu viele im Deck liegen.
//
// # Über **beide** Bereiche gezählt, und das ist nachgelesen
//
// CR 100.4a: die Vierergrenze aus CR 100.2a gilt für Deck und Sideboard
// zusammen. Vier im Hauptdeck und eine im Sideboard sind **fünf**.
//
// # Standardländer sind ausgenommen
//
// CR 100.2a nimmt sie aus (Standardländer beliebig oft), CR 903.5b
// ebenso. Ob ein Posten eines ist, sagt [Deckposten.Standardland] — und wer
// mit [Posten] prüft, sagt damit, dass er es nicht weiß, und bekommt sie
// mitgezählt.
func gleichebefunde(posten []Deckposten, format Format) []Groessenbefund {
	if format.HoechstzahlGleich <= 0 {
		return nil
	}

	var (
		befunde  []Groessenbefund
		gezaehlt = map[string]int{}
		reihe    []string
		anzeige  = map[string][]string{}
	)

	for _, einer := range posten {
		if einer.Standardland {
			continue
		}

		schluessel := gruppe(einer)

		if gezaehlt[schluessel] == 0 {
			reihe = append(reihe, schluessel)
		}

		// **Jeder Name der Gruppe, nicht der erste.** Bei austauschbaren
		// Namen liegen acht Karten unter **zwei** Namen; „Von Mind Flayer,
		// the Shadow liegen 8 im Deck" schickte jemanden acht davon suchen,
		// wo nur vier stehen (#1375).
		if name := benennung(einer); !enthaeltName(anzeige[schluessel], name) {
			anzeige[schluessel] = append(anzeige[schluessel], name)
		}

		gezaehlt[schluessel] += einer.Platz.Menge
	}

	// **In der Reihenfolge der Plätze und nicht der Map** — sonst käme
	// dieselbe Eingabe zweimal anders heraus (#1289).
	for _, schluessel := range reihe {
		if gezaehlt[schluessel] > format.HoechstzahlGleich {
			befunde = append(befunde, Groessenbefund{"gleiche", fmt.Sprintf(
				"Von %s liegen %d im Deck; %q erlaubt höchstens %d.",
				strings.Join(anzeige[schluessel], " / "), gezaehlt[schluessel],
				format.Name, format.HoechstzahlGleich)})
		}
	}

	return befunde
}

// gruppe gibt den Schlüssel, unter dem ein Posten zählt.
//
// # Der Name schlägt den Druck
//
// **CR 100.2a und CR 903.5b begrenzen Karten mit einem Namen.** Vier
// verschiedene Drucke von „Lightning Bolt" sind vier Karten desselben
// Namens — und genau das zählte der erste Anlauf nicht (Befund 1 des
// Zweitblicks über Phase 4.6).
//
// **Ohne bekannten Namen entscheidet der Bezug**, und das ist die ehrliche
// Vorgabe: wer den Namen nicht kennt, kann nicht nach ihm gruppieren.
//
// # Austauschbare Namen zählen als derselbe Name
//
// CR 100.2a hört nicht bei „no more than four" auf. Der letzte Satz lautet:
//
//	For the purposes of deck construction, cards with interchangeable names
//	have the same English name (see rule 201.3).
//
// Und CR 201.3b: für Deckbau und Formatzulässigkeit gelten Karten mit
// austauschbaren Namen als dieselbe Karte. Zwei Karten mit
// **verschiedenen** englischen Namen können also dieselbe Karte sein —
// gemeint sind die Paare aus „Universes Beyond" und „Universes Within".
//
// **Deshalb schlägt die Orakelkennung den Namen.** Scryfall gibt beiden
// Namen desselben Paares **eine** `oracle_id`; nachgemessen am 2026-09-04
// gegen `api.scryfall.com` an drei Paaren:
//
//	Mind Flayer, the Shadow  ↔ Arvinox, the Mind Flail        9abecd23…
//	Eleven, the Mage         ↔ Cecily, Haunted Mage           a726c27d…
//	Dustin, Gadget Genius    ↔ Hargilde, Kindly Runechanter   747009e7…
//
// # Warum die Suche es nicht zeigt
//
// `oracleid:<id>` gibt über
// `/cards/search` genau einen Namen je Gruppe — der Suchindex führt den
// Orakelnamen. Sichtbar wird das Paar erst am **Kartenobjekt**: ein
// Namensschlag auf „Mind Flayer, the Shadow" liefert eine Karte dieses
// Namens mit der `oracle_id` von „Arvinox, the Mind Flail". Wer nur sucht,
// sieht es nie.
//
// # Was die Messung nicht belegt
//
// **Drei Paare sind drei Paare.** Dass **jedes** austauschbare Paar sich eine
// `oracle_id` teilt, ist damit nicht gezeigt — nur dass diese drei es tun und
// dass die Quelle das Merkmal überhaupt trägt. Ein vollständiger Abgleich
// bräuchte Scryfalls Bulk-Datei; er ist nicht gefahren.
//
// **Die Richtung stimmt trotzdem.** Zwei Karten mit derselben `oracle_id`
// **sind** dieselbe Karte — das ist die Bedeutung des Feldes, nicht eine
// Vermutung darüber. Wo die Kennung fehlt, entscheidet weiter der Name; die
// Umstellung kann deshalb ein falsches Ja beheben, aber keines erzeugen.
func gruppe(einer Deckposten) string {
	// **Die Orakelkennung zuerst** — sie trägt die austauschbaren Namen mit,
	// der Name allein nicht.
	if orakel := strings.TrimSpace(einer.Orakelkennung); orakel != "" {
		return "orakel:" + strings.ToLower(orakel)
	}

	if name := strings.TrimSpace(einer.Name); name != "" {
		return "name:" + strings.ToLower(name)
	}

	if einer.Platz.KarteID != 0 {
		return fmt.Sprintf("karte:%d", einer.Platz.KarteID)
	}

	return "druck:" + Druckschluessel(einer.Platz.DruckID, einer.Platz.Printing)
}

// benennung gibt, wie ein Posten in einer Meldung heißt.
//
// **Der Name, wenn es einen gibt** — „Von Lightning Bolt liegen 5 im Deck"
// sagt mehr als „Von einer Karte".
func benennung(einer Deckposten) string {
	if name := strings.TrimSpace(einer.Name); name != "" {
		return name
	}

	return "einer Karte"
}

// enthaeltName sagt, ob ein Name schon in der Anzeige steht.
//
// **Nach Groß- und Kleinschreibung unterschieden wie [gruppe] es tut** —
// „Lightning Bolt" und „lightning bolt" sind eine Karte, und zweimal derselbe
// Name in einer Meldung ist ein Fehler, den niemand sich erklären kann.
func enthaeltName(schon []string, name string) bool {
	for _, eintrag := range schon {
		if strings.EqualFold(eintrag, name) {
			return true
		}
	}

	return false
}
