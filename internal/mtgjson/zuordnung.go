package mtgjson

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Wandlung formt die rohen Werte eines Feldpaars in einen Kartenfeldwert um.
//
// Sie bekommt die Werte in der Reihenfolge der Schlüssel aus `Feldpaar.JSON`;
// ein Schlüssel, den die Karte nicht führt, kommt als `nil`.
type Wandlung func(werte []json.RawMessage) (string, error)

// Feldpaar ist eine Zeile der Zuordnung.
//
// **Ein leeres `Feld` heißt: absichtlich nicht übernommen** — und braucht dann
// einen `Grund`. Ohne diese Möglichkeit wäre jedes bewusst übergangene Feld von
// einem unbekannten nicht zu unterscheiden, und der Leser meldete einen Befund
// über etwas, das längst entschieden ist. Dass der Grund dasteht, prüft
// `TestJedesUebergangeneFeldNenntSeinenGrund` — auch eine Ablehnung braucht
// Deckung (§10).
type Feldpaar struct {
	// JSON sind die Schlüssel, die dieses Paar verbraucht. Mehrere, weil MSE
	// Felder führt, die aus mehr als einer Angabe entstehen.
	JSON []string

	// Feld ist das Kartenfeld, in das der Wert geht. Leer: nicht übernommen.
	Feld string

	// Grund sagt, warum nicht übernommen wird. Pflicht, wenn `Feld` leer ist.
	Grund string

	// Wandeln formt um. Nil heißt: der einzelne Wert wörtlich als Zeichenkette.
	Wandeln Wandlung

	// Rueckwaerts formt beim Schreiben zurück. Nil heißt: als Zeichenkette.
	Rueckwaerts Rueckwandlung

	// Ziel ist der JSON-Schlüssel, in den geschrieben wird. Leer: der erste aus
	// `JSON`.
	//
	// Nötig, weil Lesen und Schreiben verschiedene Reihenfolgen brauchen: beim
	// Lesen bestimmt die Reihenfolge in `JSON`, wie die Werte verbunden werden
	// (`supertypes` vor `types` ergibt „Legendary Creature"); beim Schreiben
	// geht alles in **einen** Schlüssel, und das ist `types`.
	Ziel string

	// Verlust sagt, was beim Schreiben **nicht** mitgeht. Leer heißt: nichts.
	//
	// Er steht hier und nicht in einem Kommentar im Code, weil er am Ergebnis
	// erscheinen soll (#142): was nicht abbildbar ist, wird benannt, nicht
	// stillschweigend weggelassen.
	Verlust string
}

// Rueckwandlung formt einen Kartenfeldwert in den JSON-Wert zurück.
type Rueckwandlung func(text string) any

// Zuordnung sagt, welches MTG-JSON-Feld auf welches Kartenfeld eines Spiels
// geht.
//
// **Sie steht nicht fest im Code verstreut, sondern hier an einer Stelle.** Ein
// MSE-Spiel bestimmt seine Felder selbst (`.mse-game`, `card field`); eine
// erfundene Zuordnung ließe genau die Pakete nicht importieren, für die jemand
// sie braucht. Weitere Spiele kommen über `Aufnehmen` dazu.
type Zuordnung struct {
	Spiel  string
	Felder []Feldpaar
}

// Feld gibt das Feldpaar zu einem JSON-Schlüssel; der zweite Rückgabewert sagt,
// ob die Zuordnung ihn kennt.
func (z Zuordnung) Feld(schluessel string) (Feldpaar, bool) {
	for _, paar := range z.Felder {
		for _, kandidat := range paar.JSON {
			if kandidat == schluessel {
				return paar, true
			}
		}
	}

	return Feldpaar{}, false
}

// ZuordnungMagic ist die Zuordnung für das Spiel `magic`.
//
// # Woher die Zielnamen stammen
//
// Aus `card_fields` des Pakets `magic.mse-game` — einer **Datendatei**, nicht
// aus MSE2-Quelltext (§11). Die Fundstellen stehen in `docs/FORMATE.md`.
//
// # Warum drei Felder fehlen, die es gibt
//
// `type`, `pt` und `card number` tragen in `card_fields` die Angabe
// `save value: false`: MSE **leitet** sie ab und schreibt sie nicht in die
// Datei. Wer dorthin zuordnete, schriebe einen Wert, den MSE beim nächsten
// Öffnen überschreibt — und ließe die Felder leer, aus denen er entsteht.
var ZuordnungMagic = Zuordnung{
	Spiel: "magic",
	Felder: []Feldpaar{
		// **`faceName` vor `name`** (#888), nachgelesen am 2026-08-24 unter
		// `https://mtgjson.com/data-models/card/card-set/`:
		//
		//	name     — The name of the card. Cards with multiple faces are
		//	           given a `//` delimiter. Example: "Wear // Tear"
		//	faceName — The name on the face of the card.
		//
		// MTGJSON legt **je Seite ein eigenes Kartenobjekt** an, und beide
		// tragen in `name` denselben zusammengesetzten Namen. Wer nur `name`
		// läse, bekäme zwei MSE-Karten namens „Wear // Tear" — und keine von
		// beiden trüge den Namen, der auf ihr steht.
		//
		// **Der Rückweg schreibt `name`**, denn eine MSE-Karte *ist* eine
		// Seite und führt ihren Namen dort. Den zusammengesetzten könnte man
		// nur erfinden: die andere Seite steht in einer anderen Karte.
		{
			JSON: []string{"faceName", "name"}, Ziel: "name", Feld: "name",
			Wandeln: Seitenname,
			// **Kein Verlust** (Zweitblick über Phase V, 2026-08-24): die andere
			// Seite steht in **derselben** Karte, und `VerbindeSeiten` setzt den
			// zusammengesetzten Namen beim Schreiben wieder zusammen.
			//
			// Eine Verlustangabe, die keinen Verlust beschreibt, ist schlimmer
			// als keine — sie sagt jemandem, er solle etwas nachprüfen, das
			// stimmt.
		},
		{
			JSON: []string{"manaCost"}, Feld: "casting cost",
			Wandeln: Manakosten, Rueckwaerts: Manasymbole,
		},

		// `super type` trägt bei MSE Ober- **und** Kartentypen zusammen:
		// `add_cards_scripts` legt Standardländer mit
		// `super_type: "Basic Land"` an — „Basic" ist der Obertyp, „Land" der
		// Kartentyp.
		{
			JSON: []string{"supertypes", "types"}, Ziel: "types", Feld: "super type",
			Wandeln: Woerter, Rueckwaerts: WoerterZurueck,
			Verlust: "MSE führt Ober- und Kartentypen in einem Feld; MTG JSON trennt sie in »supertypes« und »types«. " +
				"Die Trennung steht im Wert nicht drin und wird nicht erfunden — alles geht als »types« hinaus",
		},
		{
			JSON: []string{"subtypes"}, Feld: "sub type",
			Wandeln: Woerter, Rueckwaerts: WoerterZurueck,
		},

		{JSON: []string{"text"}, Feld: "rule text"},
		{JSON: []string{"flavorText"}, Feld: "flavor text"},
		{JSON: []string{"power"}, Feld: "power"},
		{JSON: []string{"toughness"}, Feld: "toughness"},
		{JSON: []string{"loyalty"}, Feld: "loyalty"},
		{
			JSON: []string{"rarity"}, Feld: "rarity",
			Wandeln: Seltenheit, Rueckwaerts: SeltenheitZurueck,
		},
		{JSON: []string{"artist"}, Feld: "illustrator"},

		{
			JSON:  []string{"type"},
			Grund: "MSE setzt die Typzeile aus »super type« und »sub type« zusammen (card_fields: save value: false)",
		},
		{
			JSON:  []string{"number"},
			Grund: "MSE bildet »card number« aus der Stellung im Set (card_fields: save value: false)",
		},

		// **Die drei, die zu einer Kartenseite gehören, aber kein Feld sind.**
		{
			JSON: []string{"side"},
			Grund: "die Angabe wird beim Lesen **benutzt**, aber in kein Kartenfeld " +
				"übernommen: sie sagt, welche Seite vorn steht. MSE führt die " +
				"Reihenfolge im Feldnamen — »name« und »name 2«",
		},
		{
			JSON: []string{"otherFaceIds"},
			Grund: "die Angabe wird beim Lesen **benutzt**, aber in kein Kartenfeld " +
				"übernommen: sie sagt, welche Kartenobjekte zu einer MSE-Karte " +
				"zusammengehören (siehe fuehreSeitenZusammen). Beim Schreiben wird " +
				"sie wieder gesetzt",
		},

		// **Und die beiden, die anderswo hingehören.**
		{
			JSON: []string{"rulings"},
			Grund: "MSE hat kein Feld für Regelhinweise. Sie stehen in der Anwendung " +
				"neben der Karte — über »lesen.scryfall.regeln« und den gleichnamigen " +
				"Kasten im Fenster, und dort sind sie aktueller als in einer Datei",
		},
		{
			JSON: []string{"identifiers"},
			Grund: "MSE hat kein Feld für fremde Kennungen. Die Herkunft einer " +
				"übernommenen Angabe steht seit #888 im Verlauf — dort, wo sie " +
				"nachprüfbar ist, statt in einem Kartenfeld, das jede Ausfuhr mitschleppte",
		},
	},
}

// zuordnungen führt die bekannten Zuordnungen je Spiel.
var zuordnungen = map[string]Zuordnung{
	ZuordnungMagic.Spiel: ZuordnungMagic,
}

// ZuordnungFuer gibt die Zuordnung eines Spiels; der zweite Rückgabewert sagt,
// ob es eine gibt.
//
// **Ein unbekanntes Spiel bekommt keine erfundene Zuordnung.** Wer eines
// braucht, nimmt es über `Aufnehmen` auf — mit Fundstelle in `docs/FORMATE.md`.
func ZuordnungFuer(spiel string) (Zuordnung, bool) {
	zuordnung, bekannt := zuordnungen[spiel]

	return zuordnung, bekannt
}

// Aufnehmen fügt eine Zuordnung hinzu. Sie ersetzt eine gleichnamige.
func Aufnehmen(zuordnung Zuordnung) { zuordnungen[zuordnung.Spiel] = zuordnung }

// Spiele nennt die Spiele, für die es eine Zuordnung gibt.
func Spiele() []string {
	namen := make([]string, 0, len(zuordnungen))
	for spiel := range zuordnungen {
		namen = append(namen, spiel)
	}

	return namen
}

// Manakosten macht aus `{2}{R}` die Schreibweise `2R`.
//
// **Die Regel selbst steht in [domain.ManakostenOhneKlammern]**, seit es einen
// zweiten Verbraucher gibt: Scryfall schreibt dieselbe Form
// (#1005).
// Hier bleibt nur, was zu diesem Format gehört — das Auspacken des JSON-Werts.
func Manakosten(werte []json.RawMessage) (string, error) {
	roh, err := ersteZeichenkette(werte)
	if err != nil || roh == "" {
		return "", err
	}

	return domain.ManakostenOhneKlammern(roh), nil
}

// Woerter verbindet die Zeichenketten mehrerer Listen mit einem Leerzeichen.
func Woerter(werte []json.RawMessage) (string, error) {
	var alle []string

	for _, roh := range werte {
		if roh == nil {
			continue
		}

		var liste []string
		if err := json.Unmarshal(roh, &liste); err != nil {
			return "", fmt.Errorf("Liste von Wörtern erwartet: %w", err)
		}

		alle = append(alle, liste...)
	}

	return strings.Join(alle, " "), nil
}

// Seltenheit bildet die Seltenheiten von MTG JSON auf die Auswahl von MSE ab.
//
// **Genau ein Name unterscheidet sich:** `card_fields` führt die Möglichkeiten
// `basic land`, `common`, `uncommon`, `rare`, `mythic rare`, `special`,
// `masterpiece`; MTG JSON schreibt `mythic`. Alles andere geht wörtlich durch —
// eine Seltenheit, die MSE nicht kennt, ist ein Fall für die Prüfung gegen das
// Spielpaket, nicht für stilles Verbiegen hier.
func Seltenheit(werte []json.RawMessage) (string, error) {
	roh, err := ersteZeichenkette(werte)
	if err != nil || roh != "mythic" {
		return roh, err
	}

	return "mythic rare", nil
}

// ersteZeichenkette liest den ersten Wert als Zeichenkette.
func ersteZeichenkette(werte []json.RawMessage) (string, error) {
	if len(werte) == 0 || werte[0] == nil {
		return "", nil
	}

	var text string
	if err := json.Unmarshal(werte[0], &text); err != nil {
		return "", fmt.Errorf("Zeichenkette erwartet: %w", err)
	}

	return text, nil
}

// Seitenname gibt den Namen der Kartenseite.
//
// **`faceName` gewinnt, `name` ist der Rückfall.** Eine gewöhnliche Karte hat
// kein `faceName` und behält damit ihren Namen; eine Seite ohne `faceName` —
// eine von Hand geschriebene Datei kann `side` führen und `faceName` nicht —
// bekommt den zusammengesetzten, statt namenlos zu bleiben (§8).
func Seitenname(werte []json.RawMessage) (string, error) {
	for _, roh := range werte {
		if roh == nil {
			continue
		}

		var wert string
		if fehler := json.Unmarshal(roh, &wert); fehler != nil {
			return "", fmt.Errorf("den Kartennamen lesen: %w", fehler)
		}

		if wert != "" {
			return wert, nil
		}
	}

	return "", nil
}
