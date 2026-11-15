package scryfall

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gedankenstrich trennt bei Scryfall Ober- von Untertypen.
//
// **Ein Halbgeviertstrich (U+2014), kein Bindestrich.** Ein Untertyp kann
// einen Bindestrich enthalten; wer an ihm träfe, zerschnitte den Typ.
const gedankenstrich = "—"

// Feldpaar ist eine Zeile der Zuordnung.
//
// **Ein leeres `Feld` heißt: absichtlich nicht übernommen** — und braucht dann
// einen `Grund`. Ohne diese Möglichkeit wäre jedes bewusst übergangene Feld von
// einem vergessenen nicht zu unterscheiden. Dieselbe Gestalt wie bei
// `internal/mtgjson`, und aus demselben Grund: auch eine Ablehnung braucht
// Deckung (§10).
type Feldpaar struct {
	// Quelle ist der Name im DTO — der deutsche Feldname dieses Pakets.
	Quelle string

	// Feld ist das Kartenfeld, in das der Wert geht. Leer: nicht übernommen.
	Feld string

	// Grund sagt, warum nicht übernommen wird. Pflicht, wenn `Feld` leer ist.
	Grund string

	// Lies holt den Wert aus einer Karte. Nil bei einem übergangenen Feld.
	Lies func(*Karte) string
}

// Zuordnung sagt, welches Scryfall-Feld auf welches Kartenfeld eines Spiels
// geht.
//
// **An einer Stelle je Spiel** (§2, §16). Über den Code verteilt wäre nicht zu
// sehen, was ein Spiel übernimmt und was nicht.
type Zuordnung struct {
	Spiel  string
	Felder []Feldpaar
}

// Werte gibt die Kartenfeldwerte zu einer Karte.
//
// **Leere Werte fallen weg.** Ein Feld, das Scryfall nicht liefert, soll ein
// vorhandenes nicht überschreiben — §16, Regel 6: nichts wird still
// übernommen, und ein leerer Wert ist keine Angabe.
func (z Zuordnung) Werte(karte *Karte) map[string]string {
	if len(karte.Seiten) < 2 {
		return z.werteAus(karte, "")
	}

	// **Die erste Seite füllt die gewöhnlichen Felder, die zweite die mit
	// ` 2`** (#1040).
	werte := z.werteAus(karte.MitSeite(0), "")

	for feld, wert := range z.werteAus(karte.MitSeite(1), " 2") {
		werte[feld] = wert
	}

	return werte
}

// werteAus liest die Karte und hängt jedem Feldnamen einen Zusatz an.
func (z Zuordnung) werteAus(karte *Karte, zusatz string) map[string]string {
	werte := make(map[string]string, len(z.Felder))

	for _, paar := range z.Felder {
		if paar.Feld == "" || paar.Lies == nil {
			continue
		}

		if wert := paar.Lies(karte); wert != "" {
			werte[paar.Feld+zusatz] = wert
		}
	}

	return werte
}

// ZuordnungMagic ist die Zuordnung für das Spiel `magic`.
//
// # Woher die Zielnamen stammen
//
// Aus `card_fields` des Pakets `magic.mse-game` — einer **Datendatei**, nicht
// aus MSE2-Quelltext (§11). Es sind dieselben, die `internal/mtgjson` benutzt;
// die Fundstellen stehen in `docs/FORMATE.md`.
//
// # Warum drei Felder fehlen, die es gibt
//
// `type`, `pt` und `card number` tragen in `card_fields` die Angabe
// `save value: false`: MSE **leitet** sie ab. Wer dorthin zuordnete, schriebe
// einen Wert, den MSE beim nächsten Öffnen überschreibt.
var ZuordnungMagic = Zuordnung{
	Spiel: "magic",
	Felder: []Feldpaar{
		{Quelle: "Name", Feld: "name", Lies: func(k *Karte) string { return k.Name }},
		{
			Quelle: "Manakosten", Feld: "casting cost",
			Lies: func(k *Karte) string { return domain.ManakostenOhneKlammern(k.Manakosten) },
		},

		// **Scryfall liefert eine Zeile, MSE führt zwei Felder.** Getrennt
		// wird am Gedankenstrich — genau die Stelle, an der MTG JSON
		// `types` und `subtypes` trennt.
		{
			Quelle: "Typzeile", Feld: "super type",
			Lies: func(k *Karte) string { ober, _ := typen(k.Typzeile); return ober },
		},
		{
			Quelle: "Typzeile/Untertyp", Feld: "sub type",
			Lies: func(k *Karte) string { _, unter := typen(k.Typzeile); return unter },
		},

		// **Beide gehen durch [MitSymbolen].** Scryfall schreibt `{T}`, MSE
		// schreibt `<sym>T</sym>` — ohne die Wandlung stand der rohe Text auf
		// der Karte. Gemeldet an der laufenden Anwendung an
		// `Birds of Paradise`.
		{
			Quelle: "Regeltext", Feld: "rule text",
			Lies: func(k *Karte) string { return MitSymbolen(k.Regeltext) },
		},
		{
			Quelle: "Aromatext", Feld: "flavor text",
			Lies: func(k *Karte) string { return MitSymbolen(k.Aromatext) },
		},
		{Quelle: "Staerke", Feld: "power", Lies: func(k *Karte) string { return k.Staerke }},
		{Quelle: "Widerstand", Feld: "toughness", Lies: func(k *Karte) string { return k.Widerstand }},
		{
			Quelle: "Seltenheit", Feld: "rarity",
			Lies: func(k *Karte) string { return seltenheit(k.Seltenheit) },
		},
		{Quelle: "Kuenstler", Feld: "illustrator", Lies: func(k *Karte) string { return k.Kuenstler }},

		// Was absichtlich nicht in ein Kartenfeld geht.
		{
			Quelle: "Kennung",
			Grund: "die Scryfall-Kennung ist die Herkunft, kein Kartenwert — sie gehört an " +
				"die Angabe, woher ein Feld kam, nicht in die Karte",
		},
		{
			Quelle: "Orakelkennung",
			Grund: "`oracle_id` benennt die Karte hinter den Drucken und ist kein Wert, " +
				"der auf einer Karte steht. Gebraucht wird sie von der Höchstzahl: " +
				"austauschbare Namen — „Universes Beyond\" und „Universes Within\" — " +
				"teilen sie sich, und CR 100.2a zählt sie deshalb zusammen (#1375)",
		},
		{
			Quelle: "Sprache",
			Grund:  "die Sprache steht am Set, nicht an der Karte; MSE führt kein Sprachfeld je Karte",
		},
		{
			Quelle: "Gestalt",
			Grund: "`layout` sagt, wie die Karte gebaut ist, nicht was auf ihr steht — " +
				"welcher Stilbogen dazu passt, entscheidet das Set",
		},
		{
			Quelle: "Set",
			Grund:  "MSE führt das Set als eigene Einheit; ein Kartenfeld dafür gibt es nicht",
		},
		{
			Quelle: "Setname",
			Grund:  "dito — der Name steht in den Setangaben",
		},
		{
			Quelle: "Nummer",
			Grund:  "MSE bildet »card number« aus der Stellung im Set (card_fields: save value: false)",
		},
		{
			Quelle: "Bilder",
			Grund: "ein Bild ist kein Textfeld. Es wird gesondert übernommen — und ein " +
				"`art_crop` nur mit Künstlername und Copyright in derselben Oberfläche",
		},
		{
			Quelle: "Seiten",
			Grund: "eine mehrseitige Karte wird nicht in eine einseitige gepresst; welche " +
				"Seite in welche Karte geht, entscheidet der Aufrufer",
		},
		{
			Quelle: "Legalitaeten",
			Grund: "die Zulässigkeit gehört zum **Druck**, nicht zum Entwurf — sie " +
				"ändert sich, ohne dass sich die Karte ändert, und ein Kartenfeld " +
				"dafür wäre am Tag der nächsten Bannliste falsch. Sie wird bei der " +
				"Formatprüfung eines Decks gefragt (Phase 4.6)",
		},
		{
			Quelle: "Adresse",
			Grund:  "die Adresse bei Scryfall ist die Herkunft, kein Kartenwert",
		},
		{
			Quelle: "Farben",
			Grund: "die Farbe einer Karte ist **abgeleitet** — MSE rechnet sie " +
				"aus `casting cost` und dem Farbanzeiger, und ein Kartenfeld " +
				"daneben wäre eine zweite Wahrheit. Gelesen wird sie trotzdem: " +
				"der Retrievalindex filtert danach, und dort ist sie kein " +
				"Kartenwert, sondern ein Suchmerkmal (A15, #1514)",
		},
		{
			Quelle: "Farbidentitaet",
			Grund: "dito, und sie ist **nicht dasselbe wie die Farbe**: Regel " +
				"903.4 zählt jedes Manasymbol im Regeltext mit. MSE kennt sie " +
				"gar nicht — sie ist eine Frage an ein Deck, nicht an einen " +
				"Entwurf",
		},
		{
			Quelle: "Manawert",
			Grund: "der Manawert ist **gerechnet** und kein eigener Wert — MSE " +
				"rechnet ihn aus `casting cost`, und ein Feld dafür wäre eine " +
				"zweite Wahrheit neben den Kosten. Gelesen wird er trotzdem: " +
				"die Deckauswertung braucht ihn, weil eine zweiseitige Karte " +
				"bei Scryfall leere Kosten und trotzdem einen Wert hat " +
				"(Werkstück U10 von Phase 5.1, #1421)",
		},
	},
}

// GeleseneFelder nennt die Felder, die das DTO aus einer Antwort liest.
//
// # Wozu es das gibt
//
// Damit ein Wächter prüfen kann, dass **jedes** gelesene Feld in der Zuordnung
// erwähnt wird — mit Ziel oder mit Grund. Ein Feld, das gelesen und dann
// nirgends genannt wird, ist unsichtbar: es steht im DTO, niemand ordnet es
// zu, und niemandem fällt es auf.
//
// **Von Hand geführt und nicht aus dem Typ abgeleitet.** Go gibt die Felder
// eines Structs zur Laufzeit her, aber nicht die Absicht dahinter — und
// `Typzeile` geht in **zwei** Kartenfelder. Dass die Liste nicht zurückbleibt,
// hält `TestJedesGeleseneFeldStehtInDerZuordnung` zusammen mit dem Wächter
// über die Structfelder.
func GeleseneFelder() []string {
	return []string{
		"Kennung", "Orakelkennung", "Name", "Sprache", "Gestalt", "Manakosten", "Manawert", "Typzeile",
		"Regeltext", "Aromatext", "Staerke", "Widerstand", "Seltenheit",
		"Set", "Setname", "Nummer", "Kuenstler", "Bilder", "Seiten", "Adresse",
		"Farben", "Farbidentitaet", "Legalitaeten",
	}
}

// typen trennt eine Typzeile an ihrem Gedankenstrich.
//
// **Nur am ersten.** „Kreatur — Held — Mensch" gibt es bei Scryfall nicht,
// aber wenn es käme, gehörte alles nach dem ersten Strich in den Untertyp —
// zerschneiden wäre schlimmer als zusammenlassen.
func typen(zeile string) (ober, unter string) {
	ober, unter, gefunden := strings.Cut(zeile, gedankenstrich)
	if !gefunden {
		return strings.TrimSpace(zeile), ""
	}

	return strings.TrimSpace(ober), strings.TrimSpace(unter)
}

// seltenheit bildet die Seltenheiten von Scryfall auf die Auswahl von MSE ab.
//
// **Nur `mythic` heißt anders.** Die anderen fünf — `common`, `uncommon`,
// `rare`, `special`, `bonus` — stehen bei MSE genauso; `card_fields` führt für
// die höchste `mythic rare`. Dieselbe Abbildung wie bei `internal/mtgjson`.
func seltenheit(roh string) string {
	if roh == "mythic" {
		return "mythic rare"
	}

	return roh
}

// zuordnungen führt die bekannten Zuordnungen je Spiel.
var zuordnungen = map[string]Zuordnung{ZuordnungMagic.Spiel: ZuordnungMagic}

// ZuordnungFuer gibt die Zuordnung eines Spiels.
//
// **Ein unbekanntes Spiel bekommt keine erfundene Zuordnung.** Der zweite
// Rückgabewert sind die Befunde — heute keine; er steht da, weil eine
// Zuordnung, die etwas nicht abbilden kann, es sagen soll, statt es
// wegzulassen.
func ZuordnungFuer(spiel string) (Zuordnung, []domain.Befund, bool) {
	zuordnung, bekannt := zuordnungen[spiel]

	return zuordnung, nil, bekannt
}
