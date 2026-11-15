// **Aus einer Karte wieder zwei** — der Schreibweg mehrseitiger Karten.
//
// # Warum eine eigene Datei
//
// `kartenseiten.go` beantwortet die Frage „wie werden zwei MTGJSON-Karten zu
// **einer** MSE-Karte?"; hier steht die Umkehrung. Zusammen gingen sie über
// die Zeilengrenze aus §2 — und der Schnitt folgt der Frage, nicht der Zahl.
//
// **Was oben belegt ist, gilt auch hier**: MSE führt eine mehrseitige Karte
// als eine Karte mit numerierten Feldreihen, und die Belege dafür stehen im
// Kopf von `kartenseiten.go`.

package mtgjson

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TrenneKartenwerte macht aus den Werten einer zusammengelegten Karte die
// Werte ihrer Seiten.
//
// # Die Umkehrung von [fuehreSeitenZusammen]
//
// Wenn das Lesen zwei MTGJSON-Karten zu **einer** MSE-Karte macht, muss das
// Schreiben aus **einer** MSE-Karte wieder zwei machen.
//
// # Warum vor der Zuordnung und nicht danach
//
// **Weil der Zusatz ` 2` ein MSE-Feldname ist.** Nach der Zuordnung heißt das
// Feld `manaCost`, und `casting cost 2` ist dort nie angekommen — der erste
// Anlauf trennte danach und verlor damit jeden Wert der zweiten Seite außer
// dem Namen. Gemessen: `manaCost = {1}{R} / <nil>`.
//
// **Eine leere zweite Liste heißt: einseitig.** Der Aufrufer schreibt dann
// eine Karte, wie er es immer getan hat.
//
// # Getrennt wird nur, wenn `name 2` einen Wert trägt
//
// **Befund 1 des Zweitblicks über Phase 3.6**
// (#1167).
// Hier stand „jedes Feld mit dem Zusatz ` 2`", und das war zu breit:
// `magic.mse-game` führt **40** Felder mit einer Zahl, die **kein Grundfeld**
// haben — `loyalty cost 1` bis `loyalty cost 16` sind die Zeilen eines
// Planeswalkers, `additional credit 2` ist die zweite Zeile einer Danksagung
// (gezählt am 2026-08-27, siehe [domain.Game.Feldreihen]).
//
// **Ein einseitiger Planeswalker wurde damit in zwei MTG-JSON-Karten zerlegt**,
// verbunden über `otherFaceIds` — still, und in jeder Ausfuhr.
//
// **`name 2` ist die richtige Marke**, und zwar aus zwei Gründen:
//
//   - [fuehreSeitenZusammen] setzt sie selbst: es legt beim Lesen für jede
//     zusammengelegte Karte einen Namen an die zweite Stelle.
//   - MTG JSON verlangt für jedes Kartenobjekt einen Namen ([kennungFuer]
//     bildet die Kennung daraus). Eine Rückseite ohne ihn ließe sich gar
//     nicht schreiben.
//
// **Was das nicht löst**, und das gehört ausgesprochen: an einer **wirklich**
// zweiseitigen Karte wandert `loyalty cost 2` weiter nach hinten, obwohl es
// dort die zweite Fähigkeit der **Vorderseite** meint. Das zu trennen bräuchte
// die Feldliste des Spiels, und die hat der Schreiber nicht (`Zuordnung` kennt
// den Spielnamen, nicht seine Felder). Der Fall ist selten und die Wirkung
// klein — der behobene war häufig und die Wirkung groß.
func TrenneKartenwerte(werte []domain.Value) (vorn, hinten []domain.Value) {
	if !domain.WerteHabenZweiteSeite(werte) {
		return werte, nil
	}

	for _, wert := range werte {
		if ohne, hat := ohneZusatz(wert.Field); hat {
			hinten = append(hinten, domain.Value{Field: ohne, Text: wert.Text})

			continue
		}

		vorn = append(vorn, wert)
	}

	return vorn, hinten
}

// VerbindeSeiten setzt die Angaben, die zwei Kartenobjekte verbinden.
//
// # Was MTGJSON von einer Seite erwartet
//
// Nachgelesen am 2026-08-24 unter
// `https://mtgjson.com/data-models/card/card-set/`:
//
//   - `name` trägt **beide** Namen mit `//` — „Cards with multiple faces are
//     given a `//` delimiter."
//   - `faceName` den Namen dieser Seite.
//   - `side` ist `"a"` oder `"b"`.
//   - `otherFaceIds` nennt die Kennungen der anderen Seiten.
func VerbindeSeiten(vorn, hinten map[string]any) {
	// **Der zusammengesetzte Name steht in beiden.** So führt MTGJSON ihn, und
	// ein Leser erkennt daran, dass die Karte mehrseitig ist, bevor er
	// `otherFaceIds` ansieht.
	nameVorn := zeichenketteAus(vorn["name"])
	nameHinten := zeichenketteAus(hinten["name"])
	beide := zusammengesetzt(nameVorn, nameHinten)

	kennungVorn := kennungFuer(beide, "a")
	kennungHinten := kennungFuer(beide, "b")

	vorn["faceName"] = nameVorn
	vorn["name"] = beide
	vorn["side"] = "a"
	vorn["uuid"] = kennungVorn
	vorn["otherFaceIds"] = []any{kennungHinten}

	hinten["faceName"] = nameHinten
	hinten["name"] = beide
	hinten["side"] = "b"
	hinten["uuid"] = kennungHinten
	hinten["otherFaceIds"] = []any{kennungVorn}
}

// ohneZusatz nimmt einem Schlüssel den Zusatz der zweiten Seite.
//
// **Nur am Ende und nur mit Leerzeichen.** `name 2` ist die zweite Seite;
// `name 20` wäre ein Feld dieses Namens, und `2 name` gar keines.
func ohneZusatz(schluessel string) (string, bool) {
	const zusatz = " 2"

	if !strings.HasSuffix(schluessel, zusatz) {
		return schluessel, false
	}

	return strings.TrimSuffix(schluessel, zusatz), true
}

// kennungFuer bildet eine stabile Kennung aus Name und Seite.
//
// **Aus dem Namen und nicht gewürfelt.** Eine Zufallskennung ließe jeden
// Rundlauf eine andere Datei schreiben, und ein Vergleich zweier Läufe wäre
// nicht mehr möglich (§3: „ein Rundlauf, dessen Reihenfolge bei jedem Lauf
// anders ist, ließe sich nicht vergleichen" — hier gilt dasselbe für den
// Inhalt).
//
// **Keine UUID.** MTGJSON schreibt dort eine; wir können keine erfinden, die
// zu Scryfall passt, und eine erfundene UUID sähe echter aus als sie ist. Die
// Kennung sagt, woraus sie gebildet ist.
func kennungFuer(name, seite string) string {
	summe := sha256.Sum256([]byte(name + "|" + seite))

	return "mse4mac-" + hex.EncodeToString(summe[:8]) + "-" + seite
}

// zeichenketteAus gibt einen Wert als Zeichenkette — oder nichts.
//
// **Kein Ersatz.** Ein Name, den die Karte nicht führt, bleibt leer; ihn zu
// erfinden hieße, eine Angabe zu machen, die die Karte nicht macht.
func zeichenketteAus(wert any) string {
	text, gut := wert.(string)
	if !gut {
		return ""
	}

	return text
}

// zusammengesetzt baut den Namen, den MTGJSON für eine mehrseitige Karte führt.
//
// **`//` nur zwischen zwei Namen.** Fehlt der der Vorderseite, wäre
// `" // Tear"` eine Behauptung über eine Vorderseite, die keinen Namen hat —
// der Trenner ohne die Seiten sagt nichts, und ein Leser hielte das
// Leerzeichen für den Namen.
//
// # Und eine Rückseite hat immer einen Namen
//
// [TrenneKartenwerte] trennt nur, **wenn `name 2` einen Wert trägt** (Befund
// 1 des Zweitblicks über Phase 3.6): gibt es eine Rückseite, hat sie einen
// Namen. Ein Zweig `hinten == ""` wäre einer, den kein Lauf betritt — eine
// Behauptung darüber, was im Ernstfall geschieht (§3).
func zusammengesetzt(vorn, hinten string) string {
	if vorn == "" {
		return hinten
	}

	return vorn + " // " + hinten
}
