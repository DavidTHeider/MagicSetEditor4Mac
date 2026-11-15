package scryfall

import (
	"fmt"
	"maps"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gestalten sind die Werte, die `layout` haben kann.
//
// # Woher die Liste stammt
//
// Nachgelesen am 2026-08-23 unter `https://scryfall.com/docs/api/layouts`.
// Sie steht hier und nicht als Kommentar, weil ein unbekannter Wert **gemeldet**
// werden soll — und dafür muss jemand wissen, welche bekannt sind.
//
// # Was ein unbekannter Wert bedeutet
//
// Nicht „kaputt", sondern „die Karte sieht anders aus, als dieses Programm
// annimmt". `layout` sagt, wie viele Seiten sie hat und wie sie gezeichnet
// wird; wer einen neuen Wert verschluckt, zeichnet sie falsch.
var gestalten = []string{
	"normal", "split", "flip", "transform", "modal_dfc", "meld", "leveler",
	"class", "case", "saga", "adventure", "mutate", "prototype", "battle",
	"planar", "scheme", "vanguard", "token", "double_faced_token", "emblem",
	"augment", "host", "art_series", "reversible_card",
}

// seltenheiten sind die Werte, die `rarity` haben kann.
//
// Nachgelesen am selben Tag unter `https://scryfall.com/docs/api/cards`.
var seltenheiten = []string{"common", "uncommon", "rare", "special", "mythic", "bonus"}

// zulaessigkeiten sind die Werte, die ein Eintrag in `legalities` haben kann.
//
// # Woher die Liste stammt
//
// **Gemessen am 2026-09-01** an `api.scryfall.com/cards/named`: über drei
// Karten mit absichtlich verschiedener Lage — „Black Lotus" (banned,
// restricted, not_legal), „Channel" (alle vier) und „Llanowar Elves" (legal,
// not_legal) — kamen genau diese vier Wörter vor.
//
// **Was hier nicht behauptet wird**: dass es keine weiteren gibt. Genau
// deshalb wird ein unbekanntes **gemeldet** statt verschluckt.
var zulaessigkeiten = []string{"legal", "not_legal", "restricted", "banned"}

// aufzaehlungen meldet Werte, die dieses Programm nicht kennt.
//
// # Warum als Befund und nicht als Fehler
//
// Dieselbe Haltung wie bei `msepaket` (§8): ein Set, von dem 300 Karten lesbar
// sind und zwei nicht, wird mit 300 Karten und zwei benannten Befunden
// geöffnet, nicht abgewiesen. Eine Karte mit unbekanntem `layout` ist
// benutzbar — sie sieht nur womöglich anders aus.
//
// # Was er nicht meldet
//
// **Einen leeren Wert.** Ein `layout`, das gar nicht dasteht, ist kein
// unbekannter Wert, sondern ein fehlendes Feld — und `layout` ist nach den
// Bedingungen keines der drei Pflichtfelder. Es dort aufzunehmen hieße, eine
// Strenge zu behaupten, die Scryfall nicht zusagt.
func aufzaehlungen(gelesen rohkarte) []domain.Befund {
	var befunde []domain.Befund

	for _, gepruefte := range []struct {
		feld    string
		wert    string
		bekannt []string
	}{
		{"layout", gelesen.Gestalt, gestalten},
		{"rarity", gelesen.Seltenheit, seltenheiten},
	} {
		if gepruefte.wert == "" || slices.Contains(gepruefte.bekannt, gepruefte.wert) {
			continue
		}

		befunde = append(befunde, domain.Befund{
			Datei: gelesen.Kennung,
			Ort:   gepruefte.feld,
			Art:   domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Scryfall meldet für %s den Wert %q, den dieses Programm nicht kennt. "+
					"Die Karte wird trotzdem übernommen; sie kann anders aussehen als erwartet.",
				gepruefte.feld, gepruefte.wert),
		})
	}

	return append(befunde, zulaessigkeitsbefunde(gelesen)...)
}

// zulaessigkeitsbefunde meldet unbekannte Wörter aus `legalities`.
//
// # Warum getrennt und nicht in der Schleife darüber
//
// Weil dort **ein Feld einen Wert** hat und hier **eine Abbildung viele**. Der
// Befund nennt deshalb das Format dazu: „für `kuechentisch` den Wert
// `vielleicht`" — ohne den Formatnamen müsste jemand die ganze Antwort
// durchsehen.
//
// **Geordnet**, damit dieselbe Antwort zweimal dieselben Befunde in derselben
// Reihenfolge ergibt. Go würfelt die Reihenfolge einer Map bei jedem
// Durchlauf; ohne `slices.Sorted` wäre die Meldung bei gleicher Eingabe eine
// andere — derselbe Fehler wie bei den Pflichtfeldern (#1289).
func zulaessigkeitsbefunde(gelesen rohkarte) []domain.Befund {
	var befunde []domain.Befund

	for _, format := range slices.Sorted(maps.Keys(gelesen.Legalitaeten)) {
		wert := gelesen.Legalitaeten[format]

		if wert == "" || slices.Contains(zulaessigkeiten, wert) {
			continue
		}

		befunde = append(befunde, domain.Befund{
			Datei: gelesen.Kennung,
			Ort:   "legalities." + format,
			Art:   domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Scryfall meldet für das Format %q den Wert %q, den dieses Programm "+
					"nicht kennt. Dieses Format bleibt für diesen Druck ungeprüft.",
				format, wert),
		})
	}

	return befunde
}
