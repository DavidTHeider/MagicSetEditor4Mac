package domain

// **Die quellneutrale Druckidentität** — Werkstück A2 von Phase 5.3
// (#1441),
// entschieden in [0086](../../docs/entscheidungen/0086-sieben-begriffe-fuer-eine-karte.md).
//
// # Sieben Begriffe, und keiner ist der andere
//
// Ein **Druck** ist eine Ausgabe einer Karte in einem Set mit einer Nummer —
// die lokale Zeile, an der Sammlung, Deckplätze, Preise und Verweise hängen.
// Wie **andere** ihn nennen, sind [Druckkennung]en: ein System (Scryfall,
// MTGJSON, Cardmarket, CardTrader, TCGplayer) und eine **Semantik**, denn
// nicht jede Kennung nennt einen Druck. Eine Oracle-ID nennt das
// **Konzept**, das dreißig Drucke teilen; eine Cardmarket-ID nennt ein
// **Produkt**, und ein Blueprint bei CardTrader trägt mehrere davon; eine
// Illustrations-ID nennt das **Kunstwerk**. Nur die Semantik »druck« nennt
// genau einen — und nur für sie gilt Eindeutigkeit je System.
//
// # Warum es keinen Kennungstyp gab und jetzt einen gibt
//
// Bis zu diesem Werkstück war die Scryfall-UUID die Identität — an fünf
// Stellen als `printing TEXT`, geprüft an einer, getrimmt an einer anderen.
// Was Scryfall nicht kennt, hatte keine Identität; was ein Händler anders
// schneidet, hatte keinen Platz. Der Druck bleibt lokal, seine Kennungen
// sind viele — und `printing` bleibt als **Scryfall-Alias** ([Druck.Scryfall]).

import (
	"regexp"
	"strconv"
	"strings"
)

// Kennungssystem ist, wer eine Kennung vergeben hat.
type Kennungssystem string

// Die Systeme, deren Kennungen dieses Programm führt.
//
// **Nachgelesen am 2026-09-09** (0086): CardTrader nennt je Blueprint
// `scryfall_id`, `card_market_ids` (Array) und `tcg_player_id`; MTGJSON
// führt `scryfallId`, `scryfallOracleId`, `scryfallIllustrationId`,
// `mtgjsonV4Id`, `mcmId`, `cardtraderId`, `tcgplayerProductId` und weitere.
const (
	SystemScryfall   Kennungssystem = "scryfall"
	SystemMTGJSON    Kennungssystem = "mtgjson"
	SystemCardmarket Kennungssystem = "cardmarket"
	SystemCardTrader Kennungssystem = "cardtrader"
	SystemTCGplayer  Kennungssystem = "tcgplayer"

	// SystemScrydex ist die Kennung, unter der Scrydex eine Karte führt
	// (A5, #1444).
	//
	// **Eigen und nicht Scryfalls**: sein Preisverlauf hängt an
	// `/magicthegathering/v1/cards/<id>/price_history`, und `<id>` ist
	// **seine** Kennung. Eine Scryfall-UUID dort fragte nach einer Karte, die
	// es dort nicht gibt.
	SystemScrydex Kennungssystem = "scrydex"

	// SystemGatherer ist die **Multiverse-ID**, unter der der Herausgeber
	// eine Karte führt (T8, #1473).
	//
	// **Sie ist die einzige Kennung, die Gatherer kennt** — gemessen am
	// 2026-09-19: `Details.aspx?multiverseid=191089` führt auf die Karte,
	// und einen anderen Weg dorthin gibt es nicht. Ohne sie ist Gatherer
	// **nicht verfügbar**, und das Issue sagt es wörtlich.
	SystemGatherer Kennungssystem = "gatherer"
)

// Kennungssysteme gibt die Systeme in ihrer Reihenfolge.
//
// **Als Funktion und nicht als offene Aufzählung**: die Liste steht ein
// zweites Mal in der Wanderung (`CHECK (system IN …)`), und ein Wächter in
// `internal/storage` hält beide zusammen — A5 hat gezeigt, was ein
// Auseinanderlaufen kostet (#1444).
func Kennungssysteme() []Kennungssystem {
	kopie := make([]Kennungssystem, len(kennungssysteme))
	copy(kopie, kennungssysteme)

	return kopie
}

// Semantik sagt, **was** eine Kennung nennt.
type Semantik string

// Die vier Semantiken.
const (
	// SemantikDruck — der Wert nennt genau diesen Druck; eindeutig je System.
	SemantikDruck Semantik = "druck"
	// SemantikKonzept — der Wert nennt das Konzept, das mehrere Drucke teilen.
	SemantikKonzept Semantik = "konzept"
	// SemantikProdukt — der Wert nennt ein Anbieterprodukt; mehrere je Druck.
	SemantikProdukt Semantik = "produkt"
	// SemantikKunstwerk — der Wert nennt die Illustration, über Drucke hinweg.
	SemantikKunstwerk Semantik = "kunstwerk"
)

var (
	kennungssysteme = []Kennungssystem{
		SystemScryfall, SystemMTGJSON, SystemCardmarket, SystemCardTrader,
		SystemTCGplayer, SystemScrydex, SystemGatherer,
	}
	semantiken = []Semantik{SemantikDruck, SemantikKonzept, SemantikProdukt, SemantikKunstwerk}

	// uuidkennung ist die Gestalt einer Scryfall- oder MTGJSON-Kennung —
	// dieselbe Prüfung wie am Deckplatz (`druckkennung` in deck.go).
	uuidkennung = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// Druckkennung ist, wie ein System einen Druck, sein Konzept, ein Produkt
// oder sein Kunstwerk nennt.
type Druckkennung struct {
	System   Kennungssystem
	Semantik Semantik
	Wert     string
}

// Normalisieren trimmt — und macht UUIDs klein, weil dieselbe Kennung
// dieselbe Zeichenkette sein soll. Eine Händlernummer bleibt, wie sie ist.
func (k *Druckkennung) Normalisieren() {
	k.Wert = strings.TrimSpace(k.Wert)

	if k.System == SystemScryfall || k.System == SystemMTGJSON {
		k.Wert = strings.ToLower(k.Wert)
	}
}

// Validate prüft System, Semantik und Wert.
func (k Druckkennung) Validate() []*FieldError {
	var befunde []*FieldError

	if !kennt(kennungssysteme, k.System) {
		befunde = append(befunde, NewFieldError("system", SchluesselValidierung,
			"%q ist kein bekanntes Kennungssystem.", string(k.System)))
	}

	if !kennt(semantiken, k.Semantik) {
		befunde = append(befunde, NewFieldError("semantik", SchluesselValidierung,
			"%q sagt nicht, was die Kennung nennt.", string(k.Semantik)))
	}

	wert := strings.TrimSpace(k.Wert)

	switch {
	case wert == "":
		befunde = append(befunde, NewFieldError("wert", SchluesselPflichtfeld,
			"Eine Kennung hat einen Wert."))
	case k.uuidVerlangt() && !uuidkennung.MatchString(strings.ToLower(wert)):
		befunde = append(befunde, NewFieldError("wert", SchluesselValidierung,
			"%q ist keine Kennung von %s — erwartet wird eine UUID.", wert, string(k.System)))
	}

	return befunde
}

// uuidVerlangt sagt, ob das System seine Kennungen als UUID vergibt.
func (k Druckkennung) uuidVerlangt() bool {
	return k.System == SystemScryfall || k.System == SystemMTGJSON
}

// Eindeutig sagt, ob die Kennung genau einen Druck nennt.
func (k Druckkennung) Eindeutig() bool { return k.Semantik == SemantikDruck }

// gleich sagt, ob zwei Kennungen dieselbe sind — nach Normalisierung.
func (k Druckkennung) gleich(andere Druckkennung) bool {
	k.Normalisieren()
	andere.Normalisieren()

	return k == andere
}

func kennt[T comparable](alle []T, wert T) bool {
	for _, eines := range alle {
		if eines == wert {
			return true
		}
	}

	return false
}

// Druck ist die lokale Zeile, an der alles hängt, was einen Druck meint.
type Druck struct {
	ID int64

	Name    string
	Setcode string
	Nummer  string
	Sprache string

	// Kennungen sind, wie andere Systeme diesen Druck nennen — in der
	// Reihenfolge ihrer Aufnahme.
	Kennungen []Druckkennung
}

// Normalisieren trimmt und macht den Setcode klein — wie überall sonst.
func (d *Druck) Normalisieren() {
	d.Name = strings.TrimSpace(d.Name)
	d.Setcode = strings.ToLower(strings.TrimSpace(d.Setcode))
	d.Nummer = strings.TrimSpace(d.Nummer)
	d.Sprache = strings.ToLower(strings.TrimSpace(d.Sprache))

	for i := range d.Kennungen {
		d.Kennungen[i].Normalisieren()
	}
}

// Validate prüft den Druck und jede Kennung — mit der Stelle der Kennung.
func (d Druck) Validate() []*FieldError {
	var befunde []*FieldError

	if strings.TrimSpace(d.Name) == "" && d.Vorlaeufig() {
		befunde = append(befunde, NewFieldError("name", SchluesselPflichtfeld,
			"Ein Druck hat einen Namen oder eine Kennung, die ihn nennt."))
	}

	gesehen := map[Kennungssystem]bool{}

	for i, kennung := range d.Kennungen {
		eigene := kennung.Validate()
		for _, befund := range eigene {
			befunde = append(befunde, NewFieldError(
				"kennungen["+strconv.Itoa(i)+"]."+befund.Field, befund.Key, "%s", befund.Message))
		}

		// **Eine Kennung, die selbst nicht stimmt, zählt nicht doppelt**: der
		// Befund an ihr genügt, ein zweiter über die Eindeutigkeit verwirrte.
		if len(eigene) > 0 || !kennung.Eindeutig() {
			continue
		}

		if gesehen[kennung.System] {
			befunde = append(befunde, NewFieldError("kennungen["+strconv.Itoa(i)+"].wert", SchluesselValidierung,
				"Der Druck trägt zwei Druckkennungen von %s — ein Druck hat je System genau eine.",
				string(kennung.System)))
		}

		gesehen[kennung.System] = true
	}

	return befunde
}

// Vorlaeufig sagt, ob dem Druck noch jede eindeutige Kennung fehlt.
//
// **Ein Zustand, kein Fehler** (0086): eine eigene Karte, ein Exemplar aus
// einer Liste ohne Kennung. Eine spätere Auflösung hängt die Kennung an —
// die Zeile bleibt dieselbe.
func (d Druck) Vorlaeufig() bool {
	for _, kennung := range d.Kennungen {
		if kennung.Eindeutig() && strings.TrimSpace(kennung.Wert) != "" {
			return false
		}
	}

	return true
}

// Kennung gibt den Wert zu System und Semantik — den ersten, falls es mehrere gibt.
func (d Druck) Kennung(system Kennungssystem, semantik Semantik) (string, bool) {
	werte := d.Werte(system, semantik)
	if len(werte) == 0 {
		return "", false
	}

	return werte[0], true
}

// Werte gibt alle Werte zu System und Semantik — bei Produkten sind es mehrere.
func (d Druck) Werte(system Kennungssystem, semantik Semantik) []string {
	var werte []string

	for _, kennung := range d.Kennungen {
		if kennung.System == system && kennung.Semantik == semantik {
			werte = append(werte, kennung.Wert)
		}
	}

	return werte
}

// Scryfall gibt den Scryfall-Alias — das heutige `printing` — oder "".
func (d Druck) Scryfall() string {
	wert, _ := d.Kennung(SystemScryfall, SemantikDruck)

	return wert
}

// Druckschluessel sagt, welchen Druck eine Zeile meint — der Alias zuerst,
// sonst die lokale Zeile, und "" für eine Zeile ohne Druck.
//
// **Für Gleichheit und nicht für die Ausfuhr**: `druck:4` ist eine Nummer
// dieses Bestands (0086, Entscheidung 4). Wer zwei Exemplare oder zwei
// Deckplätze vergleicht, vergleicht das hier — sonst fielen alle vorläufigen
// Drucke, die noch keinen Alias haben, in **eine** Gruppe.
func Druckschluessel(druckID int64, printing string) string {
	if printing != "" {
		return printing
	}

	if druckID != 0 {
		return "druck:" + strconv.FormatInt(druckID, 10)
	}

	return ""
}

// druckVorlaeufig sagt, ob eine Zeile einen Druck nennt, dem der Alias fehlt.
//
// **Vorläufig heißt an den Nähten: ohne Scryfall-Alias.** Die Flächen kennen
// heute genau diese Kennung; ein Druck, den nur MTGJSON eindeutig nennt,
// zeigt sich dort ebenso als vorläufig, bis A4 die Kennungen an die Flächen
// bringt ([Druck.Vorlaeufig] ist die Frage an den Druck selbst).
func druckVorlaeufig(druckID int64, printing string) bool {
	return druckID != 0 && strings.TrimSpace(printing) == ""
}
