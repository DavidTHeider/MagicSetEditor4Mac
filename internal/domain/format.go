package domain

// **Das Format** — Werkstück 4 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md):
// „Daten, nicht Code".
//
// # Warum als Daten
//
// Wer ein eigenes Format spielt, soll es anlegen können, ohne dass jemand das
// Programm ändert. Ein `switch` über Formatnamen im Code hätte für jedes
// Küchenformat eine Änderung am Programm verlangt.
//
// # Und warum hier keine Formatliste steht
//
// Weil sie erfunden wäre. Die **Namen** der Formate kann man messen — sie
// stehen als Schlüssel in Scryfalls `legalities` (`internal/scryfall`); die
// **Zahlen** stehen dort nicht. Nachgelesen sind sie nur für drei Bauarten,
// und die stehen unten mit ihrer Regel. Jede weitere Zuordnung — „ist Pioneer
// konstruiert?" — ist eine Entscheidung dessen, der das Format anlegt, und
// keine, die dieses Programm für ihn trifft.

import (
	"regexp"
	"strings"
)

// scryfallschluessel trifft einen Schlüssel aus `legalities`.
//
// **Gemessen am 2026-09-01** an `api.scryfall.com/cards/named`: alle 23
// Schlüssel bestehen aus Kleinbuchstaben und sonst nichts —
// `standardbrawl`, `paupercommander`, `predh`.
var scryfallschluessel = regexp.MustCompile(`^[a-z]+$`)

// Format ist ein Regelwerk, gegen das ein Deck geprüft wird.
type Format struct {
	// ID ist die Kennung im Bestand — 0 für eines, das es dort noch nicht gibt.
	ID int64

	// Fassung steigt mit jedem Schreiben (§7, Regel 2).
	Fassung int64

	Name string

	// Spiel ist `magic`, `space`, `vs` — ein Format ist immer das eines Spiels.
	Spiel string

	// Mindestgroesse ist die Untergrenze des Hauptdecks; 0 heißt: keine.
	Mindestgroesse int

	// Hoechstgroesse ist die Obergrenze; 0 heißt: keine.
	Hoechstgroesse int

	// Sideboardgroesse ist die Obergrenze des Sideboards; 0 heißt: keines —
	// **außer** wenn [Format.SideboardOffen] gesetzt ist.
	Sideboardgroesse int

	// SideboardOffen heißt: ein Sideboard **ohne** Obergrenze.
	//
	// # Warum es dieses Feld gibt und nicht eine dritte Bedeutung der Null
	//
	// **Weil „kein Sideboard" und „Sideboard ohne Grenze" zwei Dinge sind**,
	// und eine Zahl nur eines von beiden sagen kann. Commander verbietet es
	// (CR 903.5e: Commanderpartien kennen kein Sideboard), Limited hat eines
	// ohne Höchstzahl (CR 100.4b: im Limited ist Sideboard,
	// was vom Kartenvorrat nicht im Deck steckt).
	//
	// **Der erste Anlauf setzte für Limited eine Null** — und jedes Deck mit
	// einer übrigen Poolkarte galt als unzulässig. Der Kommentar daneben sagte
	// das Richtige und die Zahl das Gegenteil (Befund 2 des Zweitblicks über
	// Phase 4.6).
	SideboardOffen bool

	// HoechstzahlGleich ist, wie oft dieselbe Karte vorkommen darf; 0 heißt:
	// beliebig oft.
	HoechstzahlGleich int

	// Scryfallname ist der Schlüssel in `legalities` — leer heißt, dass
	// Scryfall dieses Format nicht kennt und jede Karte `ungeprüft` bleibt.
	Scryfallname string
}

// Normalisieren glättet, was aus einer Eingabe kommt.
func (f *Format) Normalisieren() {
	f.Name = strings.TrimSpace(f.Name)
	f.Spiel = strings.TrimSpace(f.Spiel)

	// **Kleingeschrieben und nicht abgewiesen**: der Name kommt aus einer
	// Auswahl, und ein „Standard" von dort ist gemeint, nicht falsch.
	f.Scryfallname = strings.ToLower(strings.TrimSpace(f.Scryfallname))
}

// Validate prüft ein Format.
func (f Format) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(f.Name) == "" {
		befunde = append(befunde, NewFieldError("name", SchluesselPflichtfeld,
			"Ein Format hat einen Namen."))
	}

	if strings.TrimSpace(f.Spiel) == "" {
		befunde = append(befunde, NewFieldError("spiel", SchluesselPflichtfeld,
			"Ein Format gehört zu einem Spiel."))
	}

	befunde = append(befunde, f.zahlenPruefen()...)

	if name := strings.TrimSpace(f.Scryfallname); name != "" &&
		!scryfallschluessel.MatchString(name) {
		befunde = append(befunde, NewFieldError("scryfallname", SchluesselValidierung,
			"%q ist kein Schlüssel aus Scryfalls legalities.", f.Scryfallname))
	}

	return befunde
}

// zahlenPruefen prüft die vier Zahlen eines Formats.
//
// **Getrennt, weil `Validate` sonst über der Grenze aus §2 stünde** — und weil
// „sind die Zahlen stimmig?" eine eigene Frage ist.
func (f Format) zahlenPruefen() []*FieldError {
	var befunde []*FieldError

	for _, zahl := range []struct {
		feld string
		wert int
	}{
		{"mindestgroesse", f.Mindestgroesse},
		{"hoechstgroesse", f.Hoechstgroesse},
		{"sideboardgroesse", f.Sideboardgroesse},
		{"hoechstzahlGleich", f.HoechstzahlGleich},
	} {
		if zahl.wert < 0 {
			befunde = append(befunde, NewFieldError(zahl.feld, SchluesselValidierung,
				"Die Zahl ist %d — negativ gibt es nicht.", zahl.wert))
		}
	}

	// **Null ist keine Grenze**, und darf deshalb kein Widerspruch sein.
	if f.Hoechstgroesse > 0 && f.Hoechstgroesse < f.Mindestgroesse {
		befunde = append(befunde, NewFieldError("hoechstgroesse", SchluesselValidierung,
			"Die Obergrenze %d liegt unter der Untergrenze %d.",
			f.Hoechstgroesse, f.Mindestgroesse))
	}

	return befunde
}

// Bauart ist eine Deckbauart, deren Zahlen in den Comprehensive Rules stehen.
//
// **Drei und nicht mehr.** Was dort nicht steht, wird hier nicht erfunden;
// wer ein Format anlegt, das keiner der drei folgt, setzt seine Zahlen selbst.
type Bauart string

const (
	// BauartKonstruiert ist „constructed play" — CR 100.2a und 100.4a.
	BauartKonstruiert Bauart = "konstruiert"

	// BauartCommander ist Commander — CR 903.5a, 903.5b und 903.5e.
	BauartCommander Bauart = "commander"

	// BauartLimitiert ist „limited play" — CR 100.2b.
	BauartLimitiert Bauart = "limitiert"
)

// bauarten trägt die nachgelesenen Zahlen.
//
// # Die Quelle, mit Fassung
//
// *Magic: The Gathering Comprehensive Rules*, in Kraft seit dem
// 2026-08-07, bezogen am 2026-09-01 von `media.wizards.com`. Die Regelnummern
// stehen an jedem Eintrag; wer eine Zahl anzweifelt, liest dort nach.
var bauarten = map[Bauart]Format{
	// CR 100.2a: im Constructed mindestens sechzig Karten; Standardländer
	// beliebig oft, jede andere Karte eines englischen Namens höchstens
	// viermal.
	// CR 100.4a: ein Sideboard im Constructed fasst höchstens fünfzehn
	// Karten.
	//
	// **Keine Obergrenze**: die Regel nennt nur ein Minimum.
	BauartKonstruiert: {
		Mindestgroesse: 60, Hoechstgroesse: 0, Sideboardgroesse: 15,
		HoechstzahlGleich: 4,
	},

	// CR 903.5a: genau hundert Karten samt Commander — Mindest- und
	// Höchstzahl fallen zusammen.
	// CR 903.5b: außer Standardländern trägt jede Karte einen anderen Namen —
	// also **einmal**.
	// CR 903.5e: Commanderpartien kennen kein Sideboard.
	BauartCommander: {
		Mindestgroesse: 100, Hoechstgroesse: 100, Sideboardgroesse: 0,
		HoechstzahlGleich: 1,
	},

	// CR 100.2b: im Limited mindestens vierzig Karten, und eine Karte darf
	// so oft vorkommen, wie das Produkt sie mitbringt.
	//
	// **Keine Zahl für gleiche Karten**: „so viele, wie im Produkt sind" ist
	// keine feste Grenze — und eine erfundene stünde hier falsch.
	// **Ein Sideboard ohne Höchstmaß**: CR 100.4b legt alles Übrige ins
	// Sideboard, ohne es zu begrenzen — deshalb `SideboardOffen` und nicht
	// die Null, die „kein Sideboard" bedeutet.
	BauartLimitiert: {
		Mindestgroesse: 40, Hoechstgroesse: 0, Sideboardgroesse: 0,
		HoechstzahlGleich: 0, SideboardOffen: true,
	},
}

// VonBauart gibt die Zahlen einer Bauart — Name und Spiel bleiben leer.
//
// **Ein Vorschlag, keine Vorschrift**: wer ein Format anlegt, darf jede Zahl
// danach ändern.
func VonBauart(bauart Bauart) (Format, bool) {
	format, da := bauarten[bauart]

	return format, da
}

// Bauarten gibt die Namen der Bauarten — **geordnet**, damit jede Fläche
// dieselbe Reihenfolge zeigt.
func Bauarten() []Bauart {
	return []Bauart{BauartKonstruiert, BauartCommander, BauartLimitiert}
}
