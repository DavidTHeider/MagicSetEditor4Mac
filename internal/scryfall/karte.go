// Package scryfall liest Kartendaten von Scryfall
// (#1005).
//
// # Was hier liegt
//
// Die **zweite und dritte** der drei Schichten aus AGENTS.md §16: das fremde
// DTO — die Gestalt, die Scryfall schickt, mit seinen Feldnamen — und die
// Zuordnung auf Kartenfelder. Der Transport steht in `internal/netz`.
//
// **Ein Blatt** (§2): es kennt `domain` und `netz` und sonst nichts aus dem
// Projekt.
//
// # Die Bedingungen, nachgelesen am 2026-08-23
//
// Nicht erinnert — `curl` gegen `https://scryfall.com/docs/api` und
// `/docs/api/rate-limits`, beide HTTP 200:
//
//   - `User-Agent` **und** `Accept` sind Pflicht
//   - `/cards/search`, `/cards/named`, `/cards/random`, `/cards/collection`:
//     **2/s**; `/cards/manifest`: 10/min; alles Übrige: 10/s
//   - ein 429 sperrt **30 Sekunden**; „It is not acceptable to ignore HTTP 429
//     responses"
//   - zwischenspeichern **mindestens 24 Stunden**
package scryfall

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bilder sind die Adressen, unter denen eine Karte oder Seite abgebildet ist.
//
// **Nur drei von sieben.** Scryfall liefert `small`, `normal`, `large`, `png`,
// `art_crop`, `border_crop`; hier stehen die, für die es einen Verwender gibt.
// Ein Feld ohne Verwender wäre eine Zusage, die niemand einlöst.
type Bilder struct {
	Klein  string
	Normal string
	// Ausschnitt ist `art_crop` — die Illustration ohne Rahmen.
	//
	// **Er darf nur mit Nennung erscheinen.** Die Bedingungen verlangen
	// Künstlername und Copyright in derselben Oberfläche; das ist eine
	// Anforderung an das Fenster und an die Ausfuhr, nicht an dieses Paket.
	Ausschnitt string
}

// Seite ist eine Seite einer mehrseitigen Karte.
type Seite struct {
	Name       string
	Manakosten string
	Typzeile   string
	Regeltext  string
	Aromatext  string
	Staerke    string
	Widerstand string
	Kuenstler  string
	Bilder     Bilder
}

// Karte ist eine Karte, wie Scryfall sie schickt.
//
// **Die Feldnamen sind deutsch, die JSON-Schlüssel englisch** — §9: der
// Bezeichner folgt dem Vokabular, und dies hier ist die Gestalt **dieses**
// Programms. Was Scryfall schickt, steht in den Marken.
type Karte struct {
	Kennung string

	// Orakelkennung ist `oracle_id` — **dieselbe für beide Namen** eines
	// austauschbaren Paares
	// (#1375).
	//
	// # Ein Zusatzfeld, kein Pflichtfeld
	//
	// §16 sagt: Pflichtfelder streng, Zusatzfelder tolerant. Diese Kennung
	// fehlt an Objekten, die keine Karte im Sinne des Orakels sind — und eine
	// Antwort deshalb abzuweisen, hieße eine Deckprüfung an einer Kennung
	// scheitern zu lassen, die niemand braucht, solange der Name da ist.
	//
	// **Was sie kann, sagt `domain.Deckposten.Orakelkennung`**: sie trägt die
	// austauschbaren Namen. Sichtbar ist sie nur am Kartenobjekt; die Suche
	// über `oracleid:` gibt je Gruppe den Orakelnamen und verbirgt das Paar.
	Orakelkennung string

	Name       string
	Sprache    string
	Gestalt    string
	Manakosten string

	// Manawert ist Scryfalls `cmc`, gerundet — Werkstück U10 von Phase 5.1
	// (#1421).
	//
	// **Nachgelesen und nicht gerechnet** (§15): eine zweiseitige Karte hat
	// bei Scryfall **leere** `mana_cost` und trotzdem einen `cmc`. Wer ihn
	// aus den Kosten rechnete, bekäme dort eine Null.
	//
	// **Gerundet, weil ein Deck ganze Karten zählt.** Scryfall schickt eine
	// Gleitkommazahl, und in den Un-Sets gibt es `0.5`; eine Kurve über
	// halbe Balken wäre keine Auskunft.
	Manawert int

	Typzeile   string
	Regeltext  string
	Aromatext  string
	Staerke    string
	Widerstand string
	Seltenheit string
	Set        string
	Setname    string
	Nummer     string
	Kuenstler  string
	Bilder     Bilder

	// Seiten ist leer bei einer einseitigen Karte.
	//
	// **`card_faces` und `image_uris` sind unabhängig wahlfrei.** Eine
	// gewöhnliche Karte hat Bilder und keine Seiten; eine zweiseitige hat
	// Seiten und keine Bilder auf der obersten Ebene; eine geteilte hat
	// beides. Wer eines aus dem anderen schließt, liest ein Drittel der Karten
	// falsch.
	Seiten []Seite

	// Adresse ist die Seite bei Scryfall — für die Herkunftsangabe.
	Adresse string

	// Farben sind die Farben der Karte, Farbidentitaet ihre Identität nach
	// Regel 903.4 — Werkstück A15 von Phase 5.3
	// (#1514).
	//
	// **Zwei Felder, weil es zwei Dinge sind.** Die Farbe folgt Manakosten
	// und Farbanzeigern; die Identität zählt zusätzlich jedes Manasymbol im
	// Regeltext. Eine Karte mit `{R}` im Text ist farblos und rot-identisch —
	// wer beides zusammenlegte, machte jede Kommandeur-Frage falsch.
	//
	// **Zusatzfelder, nicht Pflicht** (§16): eine zweiseitige Karte führt
	// `colors` nur je Seite, und ein Token hat keines von beidem. Leer heißt
	// **farblos oder unbekannt** — welches, sagt die Karte selbst.
	Farben         []string
	Farbidentitaet []string

	// Legalitaeten sagt je Format, ob dieser Druck dort zulässig ist.
	//
	// **Scryfalls eigene Wörter**, nicht übersetzt: `legal`, `not_legal`,
	// `restricted`, `banned`. Eine Übersetzung hier machte aus einer Auskunft
	// eine Auslegung — und dieses Programm erfindet keine Zulässigkeit
	// ([0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md)).
	//
	// **Leer heißt: nichts bekannt.** Nicht „in keinem Format zulässig".
	Legalitaeten map[string]string
}

// rohbilder ist `image_uris`.
type rohbilder struct {
	Klein      string `json:"small"`
	Normal     string `json:"normal"`
	Ausschnitt string `json:"art_crop"`
}

// rohseite ist ein Eintrag aus `card_faces`.
type rohseite struct {
	Name       string     `json:"name"`
	Manakosten string     `json:"mana_cost"`
	Typzeile   string     `json:"type_line"`
	Regeltext  string     `json:"oracle_text"`
	Aromatext  string     `json:"flavor_text"`
	Staerke    string     `json:"power"`
	Widerstand string     `json:"toughness"`
	Kuenstler  string     `json:"artist"`
	Bilder     *rohbilder `json:"image_uris"`
}

// rohkarte ist die Gestalt, die Scryfall schickt.
//
// # Warum Zeiger bei den wahlfreien Verbünden
//
// Damit „nicht da" von „da und leer" zu unterscheiden ist. Eine zweiseitige
// Karte hat **kein** `image_uris`; ein Nullwert-Verbund sähe aus wie eines mit
// leeren Adressen, und der Leser käme nicht auf die Idee, bei den Seiten
// nachzusehen.
type rohkarte struct {
	Objekt        string `json:"object"`
	Kennung       string `json:"id"`
	Orakelkennung string `json:"oracle_id"`

	Name       string `json:"name"`
	Sprache    string `json:"lang"`
	Gestalt    string `json:"layout"`
	Manakosten string `json:"mana_cost"`

	// Manawert ist Scryfalls `cmc` — Werkstück U10 von Phase 5.1 (#1421).
	//
	// **Als Gleitkommazahl gelesen, weil Scryfall sie so schickt**: es gibt
	// Karten mit `cmc: 0.5` (Un-Sets). Was hier ankommt, ist die Zahl des
	// Anbieters; gerundet wird erst dort, wo eine ganze Zahl gebraucht wird.
	Manawert float64 `json:"cmc"`

	Typzeile   string     `json:"type_line"`
	Regeltext  string     `json:"oracle_text"`
	Aromatext  string     `json:"flavor_text"`
	Staerke    string     `json:"power"`
	Widerstand string     `json:"toughness"`
	Seltenheit string     `json:"rarity"`
	Set        string     `json:"set"`
	Setname    string     `json:"set_name"`
	Nummer     string     `json:"collector_number"`
	Kuenstler  string     `json:"artist"`
	Bilder     *rohbilder `json:"image_uris"`
	Seiten     []rohseite `json:"card_faces"`
	Adresse    string     `json:"scryfall_uri"`

	// Farben ist `colors`, Farbidentitaet `color_identity`.
	Farben         []string `json:"colors"`
	Farbidentitaet []string `json:"color_identity"`

	// Legalitaeten ist `legalities` — je Format ein Wort.
	Legalitaeten map[string]string `json:"legalities"`
}

// LiesKarte liest eine Kartenantwort.
//
// # Pflichtfelder streng, Zusatzfelder tolerant (§16)
//
// **Streng**: fehlt `object`, `id` oder `name`, bricht es — und die Meldung
// sagt, welches Feld fehlt. Ohne diese Strenge stünde nach einer Änderung bei
// Scryfall eine Karte ohne Namen im Bestand, und niemand wüsste, woher sie
// kommt.
//
// **Tolerant**: ein neues Feld bricht nichts und wird auch **nicht gemeldet**.
// Scryfall darf erweitern; ein Befund je neuem Feld machte aus jeder
// Erweiterung dort eine Meldung hier.
//
// # Und die Mitte: ein unbekannter Aufzählungswert
//
// Der wird als **Befund** geführt. Ein neues Feld ist eine Erweiterung; ein
// neuer Wert für `layout` ist eine Bedeutung, die dieses Programm nicht kennt
// — wer den verschluckt, zeichnet die Karte falsch.
//
// **Er bricht trotzdem nicht.** Dieselbe Haltung wie bei `msepaket` (§8): 300
// lesbare Karten und zwei Befunde, nicht null Karten.
func LiesKarte(roh []byte) (*Karte, []domain.Befund, error) {
	var gelesen rohkarte
	if err := json.Unmarshal(roh, &gelesen); err != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, err,
			"die Antwort von Scryfall ist kein gültiges JSON")
	}

	if fehler := pflichtfelder(gelesen); fehler != nil {
		return nil, nil, fehler
	}

	befunde := aufzaehlungen(gelesen)

	return &Karte{
		Kennung: gelesen.Kennung, Orakelkennung: gelesen.Orakelkennung,
		Name: gelesen.Name, Sprache: gelesen.Sprache,
		Gestalt: gelesen.Gestalt, Manakosten: gelesen.Manakosten,
		Manawert: int(math.Round(gelesen.Manawert)),
		Typzeile: gelesen.Typzeile, Regeltext: gelesen.Regeltext,
		Aromatext: gelesen.Aromatext, Staerke: gelesen.Staerke,
		Widerstand: gelesen.Widerstand, Seltenheit: gelesen.Seltenheit,
		Set: gelesen.Set, Setname: gelesen.Setname, Nummer: gelesen.Nummer,
		Kuenstler: gelesen.Kuenstler, Bilder: alsBilder(gelesen.Bilder),
		Seiten: alsSeiten(gelesen.Seiten), Adresse: gelesen.Adresse,
		Farben: gelesen.Farben, Farbidentitaet: gelesen.Farbidentitaet,
		Legalitaeten: gelesen.Legalitaeten,
	}, befunde, nil
}

// pflichtfelder prüft, was dasein muss.
//
// # Die Reihenfolge ist festgeschrieben, und das ist der Punkt
//
// Der erste Anlauf ging über ein `map`-Literal. Go würfelt dessen Reihenfolge
// bei **jedem** Durchlauf; eine Antwort, der zwei Pflichtfelder fehlen, bekam
// damit mal die eine und mal die andere Meldung — bei derselben Eingabe, im
// selben Prozess.
//
// Wer eine Karte zweimal abruft und zwei verschiedene Gründe liest, sucht an
// der falschen Stelle, und ein Fehlerbericht darüber ist nicht nachzustellen.
//
// **`object` vor `id` vor `name`** — die Reihenfolge, in der Scryfall sie in
// seiner Antwort führt. Sie ist willkürlich und deshalb aufgeschrieben: was
// zählt, ist dass sie sich nicht ändert.
//
// Gefunden vom Nachlauf über Phase 4.2
// (#1289),
// über die Zusage „zum Ergebnis gehört auch das Nein".
func pflichtfelder(gelesen rohkarte) error {
	for _, pflicht := range []struct{ feld, wert string }{
		{"object", gelesen.Objekt},
		{"id", gelesen.Kennung},
		{"name", gelesen.Name},
	} {
		if pflicht.wert == "" {
			return fmt.Errorf(
				"%w: die Antwort von Scryfall hat kein Feld %q — das ist ein Pflichtfeld",
				domain.ErrValidation, pflicht.feld)
		}
	}

	return nil
}

// alsBilder übersetzt `image_uris`.
func alsBilder(roh *rohbilder) Bilder {
	if roh == nil {
		return Bilder{}
	}

	return Bilder{Klein: roh.Klein, Normal: roh.Normal, Ausschnitt: roh.Ausschnitt}
}

// alsSeiten übersetzt `card_faces`.
func alsSeiten(roh []rohseite) []Seite {
	if len(roh) == 0 {
		return nil
	}

	seiten := make([]Seite, 0, len(roh))
	for _, eine := range roh {
		seiten = append(seiten, Seite{
			Name: eine.Name, Manakosten: eine.Manakosten, Typzeile: eine.Typzeile,
			Regeltext: eine.Regeltext, Aromatext: eine.Aromatext,
			Staerke: eine.Staerke, Widerstand: eine.Widerstand,
			Kuenstler: eine.Kuenstler, Bilder: alsBilder(eine.Bilder),
		})
	}

	return seiten
}

// MitSeite gibt die Karte, wie sie mit den Werten einer ihrer Seiten aussähe.
//
// # Wofür
//
// **Damit dieselbe Zuordnung beide Seiten lesen kann**
// (#1040).
// Bei einer zweiseitigen Karte sind `mana_cost`, `type_line` und `oracle_text`
// auf der obersten Ebene **leer**; sie stehen je Seite. Wer nur oben läse,
// bekäme fast nichts — und genau so war es.
//
// # Was **nicht** von der Seite kommt
//
// **Seltenheit, Set und Nummer.** Sie sind Eigenschaften des Drucks, nicht
// einer Seite, und Scryfall führt sie nur oben. Die Referenz tut dasselbe:
// `push_alt!("rarity", …)` liest die Seltenheit der **Karte**
// (`magic-set-generator`, `src/mse.rs`, Zeilen 256–266, MIT — §11).
//
// **Und der Name der Karte ist nicht der Name der Seite**: oben steht
// `A // B`, an der Seite `A`. Für ein Kartenfeld ist die Seite gemeint.
func (k *Karte) MitSeite(stelle int) *Karte {
	if stelle < 0 || stelle >= len(k.Seiten) {
		return k
	}

	seite := k.Seiten[stelle]

	// **Eine Kopie, keine Änderung.** Die Karte kommt aus dem
	// Zwischenspeicher und wird mehrfach gelesen; wer sie hier umschriebe,
	// veränderte den Stand für jeden anderen Leser.
	kopie := *k

	kopie.Name = seite.Name
	kopie.Manakosten = seite.Manakosten
	kopie.Typzeile = seite.Typzeile
	kopie.Regeltext = seite.Regeltext
	kopie.Aromatext = seite.Aromatext
	kopie.Staerke = seite.Staerke
	kopie.Widerstand = seite.Widerstand
	kopie.Bilder = seite.Bilder

	// **Der Künstler nur, wenn die Seite einen nennt.** Manche Seiten führen
	// ihn, manche nicht; ein leerer Name überschriebe den der Karte.
	if seite.Kuenstler != "" {
		kopie.Kuenstler = seite.Kuenstler
	}

	return &kopie
}
