package domain

// **Was geteilt wird** — NET-04
// (#1708),
// entschieden in
// [0155](../../docs/entscheidungen/0155-teilen-ist-eine-kopie.md).
//
// **Teilen ist eine Kopie**: keine Eigentumsübertragung und keine
// angenommene Einladung (AK-04). Was hier steht, ist, **was mitgeht** und
// **was verloren geht** — als Schlüssel, die die Oberfläche in ihre Sprache
// setzt.

import (
	"fmt"
	"regexp"
	"strings"
)

// Teilart ist, was geteilt wird.
type Teilart string

const (
	// TeilartEinladung ist der Code einer Kontakteinladung — als Text.
	TeilartEinladung Teilart = "einladung"

	// TeilartKartenbild ist das Bild einer Karte als PNG.
	TeilartKartenbild Teilart = "kartenbild"

	// TeilartSet ist ein Set als `.mse-set`-Archiv.
	TeilartSet Teilart = "set"

	// TeilartDeck ist ein Deck als Liste.
	TeilartDeck Teilart = "deck"

	// TeilartSammlung ist die Sammlung als CSV.
	TeilartSammlung Teilart = "sammlung"
)

// Teilarten sind alle, in der Reihenfolge der Oberfläche.
func Teilarten() []Teilart {
	return []Teilart{TeilartEinladung, TeilartKartenbild, TeilartSet, TeilartDeck, TeilartSammlung}
}

// PruefeTeilart liest eine Art — oder sagt, dass es keine ist.
func PruefeTeilart(wert string) (Teilart, error) {
	for _, art := range Teilarten() {
		if string(art) == wert {
			return art, nil
		}
	}

	return "", fmt.Errorf("%w: %q ist nichts, was sich teilen lässt", ErrValidation, wert)
}

// Mitgehendes nennt, was in einer geteilten Datei dieser Art steht.
//
// **Eine Sammlung als CSV trägt Privates** — Erwerbspreise, Lagerorte,
// Notizen und Wünsche. Das CSV ist die verlustfreie Form (`service.FormCSV`);
// wer es teilt, soll vorher lesen, dass es all das enthält.
func (a Teilart) Mitgehendes() []string {
	return map[Teilart][]string{
		TeilartEinladung:  {"name", "adresse", "zweck", "ablauf"},
		TeilartKartenbild: {"bild"},
		TeilartSet:        {"karten", "setangaben", "schluesselwoerter", "bilder"},
		TeilartDeck:       {"mengen", "namen"},
		TeilartSammlung:   {"exemplare", "erwerb", "lagerorte", "notizen", "wuensche"},
	}[a]
}

// Verluste nennt, was beim Teilen dieser Art nicht mitgeht.
func (a Teilart) Verluste() []string {
	return map[Teilart][]string{
		TeilartKartenbild: {"kartendaten"},
		TeilartSet:        {"marken"},
		TeilartDeck:       {"zuordnungen", "format", "notiz"},
	}[a]
}

// Mailgrenze ist, ab wann die Vorschau einen anderen Weg als Mail empfiehlt.
//
// **Zwanzig Megabyte**: darüber nehmen viele Maildienste einen Anhang nicht
// mehr an, und Mail auf macOS bietet dann selbst Mail Drop an — ein Umweg,
// den nicht jedes Konto hat. AirDrop und „Datei sichern" kennen keine Grenze.
const Mailgrenze int64 = 20 << 20

// Teilgut ist, was vor dem Teilen gezeigt wird.
type Teilgut struct {
	Art       Teilart
	Titel     string
	Dateiname string
	Groesse   int64

	// Text ist, was als Text geteilt wird — bei einer Einladung der Code.
	Text string
}

// ZuGrossFuerMail sagt, ob ein anderer Weg als Mail der bessere ist (AK-03).
func (g Teilgut) ZuGrossFuerMail() bool { return g.Groesse > Mailgrenze }

// teilkennung ist die Gestalt einer Kennung: 32 Kleinbuchstaben aus Hex.
var teilkennung = regexp.MustCompile(`^[0-9a-f]{32}$`)

// PruefeTeilkennung weist ab, was keine Kennung ist.
//
// **Sie wird zu einem Pfadteil** — eine Kennung aus dem Fenster, die `..`
// enthielte, zeigte aus dem Teilordner heraus.
func PruefeTeilkennung(kennung string) error {
	if !teilkennung.MatchString(kennung) {
		return fmt.Errorf("%w: %q ist keine Kennung einer Teilung", ErrValidation, kennung)
	}

	return nil
}

// Teildateiname macht aus einem Titel einen sicheren Dateinamen.
//
// **Ohne Schrägstrich, Doppelpunkt und Punkt vorn**: der Name landet im
// Teilordner und als Anhang bei jemand anderem; ein `..` oder ein `/` darin
// wäre ein Weg und kein Name. Höchstens 120 Zeichen.
func Teildateiname(titel, endung string) string {
	name := strings.Map(func(zeichen rune) rune {
		if zeichen == '/' || zeichen == ':' || zeichen == '\\' || zeichen < ' ' {
			return '-'
		}

		return zeichen
	}, titel)

	name = strings.TrimLeft(strings.TrimSpace(name), ".")

	if runen := []rune(name); len(runen) > 120 {
		name = string(runen[:120])
	}

	if name == "" {
		name = "teilen"
	}

	return name + endung
}
