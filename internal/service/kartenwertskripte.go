// **Die Wertskripte einer Karte** — `script:` an einem `card field:`.
//
// # Warum eine eigene Datei
//
// `kartenvorgaben.go` stand an der Zeilengrenze aus §2, als das laufende Feld
// dem Schlüsselwortvermerk angesagt wurde
// (#1354).
// Der Schnitt folgt der Frage: *was ein Feld bekommt, wenn es leer ist*, steht
// dort; *was ein Feld aus sich selbst rechnet*, hier.

package service

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// mitWertskripten rechnet die Felder aus, die ein `script:` tragen.
//
// # Der Unterschied zu einem Vorgabewert
//
// `default:` **füllt** einen leeren Wert, `script:` **rechnet** ihn — MSE
// wendet beide an, und in dieser Reihenfolge:
//
//	field().default_script.invokeOnDefault(ctx, value)
//	| field().        script.invokeOn(ctx, value)
//
// `src/data/field/text.cpp`, Zeilen 178–179, Fassung f9d9356 vom 2021-07-06.
// Der bisherige Wert steht dem Skript als `value` zur Verfügung
// (`OptionalScript::invokeOn`, `src/script/scriptable.hpp`, Zeilen 57–69).
//
// # Warum das zählt
//
// Das Feld `text` von `magic.mse-game` trägt `save value: false` und entsteht
// **erst hier**:
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                field2: card.flavor_text)
//
// Regeltext, Blockgrenze, Flavortext. Ohne diesen Schritt fehlt auf jeder
// Karte der Flavortext
// (#345).
//
// # Ein gescheitertes Skript lässt den Wert stehen
//
// `invokeOn` weist nur zu, was herauskommt; wo nichts herauskommt, bleibt der
// alte Wert. Das ist §8 und zugleich die Rettung für die **72** Kartenfelder,
// deren Skript das noch nicht gebaute `text_filter` ruft: sie behalten, was
// die Karte führt, statt leer zu werden.
func mitWertskripten(
	ctx context.Context, gefuellt []msescript.Eintrag, felder []domain.Field,
	umgebung *msescript.Umgebung, vermerk *Schluesselwortverwendung,
	karteNeu func([]msescript.Eintrag),
) ([]msescript.Eintrag, []domain.Befund) {
	var befunde []domain.Befund

	// **In der Reihenfolge des Spiels, einmal und ohne Nachziehen** — wie MSE
	// beim **Laden** (#2396): `SetScriptManager::updateAll` rechnet je Karte
	// `FOR_EACH(v, card->data)`, und `card->data` legt
	// `data.init(game.card_fields)` in der Reihenfolge der Spielfelder an
	// (`src/script/script_manager.cpp`, Zeilen 265–300; `src/data/card.cpp`,
	// Zeile 37). Ein Feld, das ein **späteres** liest, sieht dessen
	// gespeicherten Wert.
	//
	// **Das trägt, weil gespeichert ist, was gerechnet wurde.** Beim
	// **Bearbeiten** rechnet MSE das geänderte Feld und zieht die abhängigen
	// nach (`updateValue`, ebd. Zeilen 249–262), und gespeichert wird der
	// gerechnete Wert; dasselbe tut seit #2396 `Feldweg` (`wertnachzug.go`).
	// Bis dahin speicherte diese Anwendung den eingegebenen Wert und rechnete
	// hier wie nach einer Bearbeitung — mit der Ladereihenfolge hätte sonst
	// auf jedem hier bearbeiteten Planeswalker der ungefilterte Text
	// gestanden (`9fbbf1065`, zurückgenommen).
	//
	// **Über die Felder, die dastehen**, in der Reihenfolge der erklärten:
	// [mitVorgaben] hat jedes erklärte Feld angelegt, und ein Eintrag, den das
	// Spiel nicht erklärt, trägt kein Skript und rückt ans Ende.
	//
	// Fassung f9d9356 vom 2021-07-06.
	skripte := make(map[string]string, len(felder))
	for i := range felder {
		if roh := strings.TrimSpace(felder[i].Skript); roh != "" {
			skripte[strings.ReplaceAll(felder[i].Name, " ", "_")] = roh
		}
	}

	gelesen := gelesenevonKarte(skripte)

	for _, stelle := range inSpielreihenfolge(gefuellt, felder) {
		if ctx.Err() != nil {
			return gefuellt, befunde
		}

		roh, da := skripte[gefuellt[stelle].Name]
		if !da {
			continue
		}

		geaendert, befund := wertskriptRechnen(ctx, &gefuellt[stelle], roh, umgebung, vermerk)
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		// **Wer gelesen wird, wird neu eingetragen.**
		//
		// `card` ist eine **Kopie**: [kartenabbildung] baut die Einträge neu,
		// sobald ein Farb- oder Bildfeld dabei ist — und das ist in jedem
		// gesichteten Spielpaket der Fall. Ein Wertskript sah deshalb den
		// Wert, den ein anderes Feld **vor** seinem eigenen Skript hatte.
		//
		// Sichtbar wurde es am Flavortext: `magic.mse-game` klammert ihn in
		// `flavor_text_filter` in `<i-flavor>`, und `text` setzt danach mit
		// `combined_editor(field2: card.flavor_text)` zusammen — und bekam den
		// ungeklammerten. Auf einer aus MTG JSON übernommenen Karte stand der
		// Flavortext deshalb aufrecht
		// (#288).
		//
		// **Der Kanonlauf hat es nie gesehen**: in einer `.mse-set`-Datei steht
		// die Auszeichnung schon drin — `Red.mse-set` enthält sie 4.508-mal.
		//
		// **Nur für Felder, die jemand liest.** Nach jedem der 142 Wertskripte
		// des Packs neu einzutragen hieße, die Abbildung 142-mal je Karte zu
		// bauen; gelesen werden davon eine Handvoll (siehe
		// [gelesenevonKarte]).
		//
		// **Und nur, wenn sich der Wert geändert hat** (PERF-03, #2170): sonst
		// steht in `card` schon, was ein späteres Skript lesen kann. Gemessen
		// an der Ausfuhr von 70 Karten: 728 MB über vier Läufe allein in
		// diesem Neubau (`docs/LEISTUNGSSCHNITTE.md`).
		if gelesen[gefuellt[stelle].Name] && geaendert {
			karteNeu(gefuellt)
		}
	}

	// **Ein Befund je Meldung, nicht je Feld.** `magic.mse-game` ruft in **72**
	// Kartenfeldern dasselbe `text_filter`; 72-mal derselbe Satz verdeckte
	// alles andere, und eine Liste, die niemand zu Ende liest, ist keine
	// Auskunft (§8).
	return gefuellt, entdoppelt(befunde)
}

// wertskriptRechnen rechnet das Wertskript eines Eintrags — `roh` ist sein
// Quelltext — und sagt, ob sich der Wert geändert hat. **Geteilt** von
// [mitWertskripten] (alle Felder, wie MSE beim Laden) und [nachziehen] (das
// geänderte und seine Leser, wie MSE beim Bearbeiten, #2396).
func wertskriptRechnen(
	ctx context.Context, eintrag *msescript.Eintrag, roh string,
	umgebung *msescript.Umgebung, vermerk *Schluesselwortverwendung,
) (geaendert bool, befund *domain.Befund) {
	// **`value` ist der bisherige Wert.** Die Skripte des Packs rechnen
	// damit — `text_filter(value, …)` steht 72-mal in `card_fields`.
	eigene := umgebung.Darueber()
	eigene.Setze("value", eintrag.Wert)

	// **Und das laufende Feld wird angesagt** (#1354).
	//
	// MSE hält es als dynamisches Argument (`value_being_updated()`,
	// `src/data/keyword.cpp` Zeile 278, Fassung `f9d9356` vom 2021-07-06)
	// und benutzt es zweimal: als
	// Schlüssel des Vermerks und als das, was `remove_from_stats` vor dem
	// Einsetzen wegräumt. Ohne den Feldnamen räumte ein Neuzeichnen von
	// `rule text` auch weg, was `rule text 2` beigetragen hat.
	if vermerk != nil {
		vermerk.BeiFeld(eintrag.Name)
	}

	wert, err := msescript.AuswertenQuelle(ctx, skriptrumpfOder(roh), eigene)
	if err != nil {
		// **Benannt und nicht still.** Ein Feld, dessen Wertskript
		// scheitert, behält seinen alten Wert — und wo es keinen gab, ist
		// es leer. Ein leerer Textkasten ohne Meldung sieht aus wie eine
		// Karte ohne Regeltext (§8).
		return false, &domain.Befund{
			Ort:     eintrag.Name,
			Meldung: "das Wertskript ließ sich nicht ausrechnen: " + err.Error(),
		}
	}

	// **Ein Fehlerwert wird gemeldet, nicht eingesetzt.**
	//
	// Bei MSE endet ein Wertskript in einer Zuweisung, und die ruft
	// `toString()`; bei einem `ScriptDelayedError` wirft die
	// (`src/script/value.cpp`, Zeile 148, Fassung f9d9356 vom
	// 2021-07-06). `Value::update` fängt den Fehler je Feld, und der alte
	// Wert bleibt stehen.
	//
	// Ohne diese Prüfung landete der Wortlaut `Fehler` — der Name der
	// Wertart — als Kartenwert im Plan
	// (#345).
	if wert.IstFehler() {
		return false, &domain.Befund{
			Ort:     eintrag.Name,
			Meldung: "das Wertskript ließ sich nicht ausrechnen: " + wert.Text,
		}
	}

	alt := eintrag.Wert
	eintrag.Wert = wert

	return !wertBleibt(alt, wert), nil
}

// inSpielreihenfolge gibt die Stellen der Einträge in der Reihenfolge, in der
// das Spiel ihre Felder erklärt — siehe [mitWertskripten]. **Stabil**: was das
// Spiel nicht erklärt, behält seine Reihenfolge hinter allen erklärten.
func inSpielreihenfolge(gefuellt []msescript.Eintrag, felder []domain.Field) []int {
	rang := make(map[string]int, len(felder))
	for i := range felder {
		rang[strings.ReplaceAll(felder[i].Name, " ", "_")] = i
	}

	reihe := make([]int, len(gefuellt))
	for i := range reihe {
		reihe[i] = i
	}

	slices.SortStableFunc(reihe, func(a, b int) int {
		return cmp.Compare(rangOder(rang, gefuellt[a].Name, len(felder)),
			rangOder(rang, gefuellt[b].Name, len(felder)))
	})

	return reihe
}

// rangOder gibt den Rang eines Feldes oder, wo das Spiel es nicht erklärt, den
// Platz hinter allen erklärten.
func rangOder(rang map[string]int, name string, hinten int) int {
	if r, da := rang[name]; da {
		return r
	}

	return hinten
}

// skriptrumpfOder schneidet die geschweiften Klammern ab, wenn welche dastehen.
//
// **Beide Formen kommen vor.** `magic.mse-game` schreibt sein `script:` als
// eingerückten Block ohne Klammern; andere Felder klammern.
func skriptrumpfOder(roh string) string {
	if rumpf, geklammert := skriptrumpf(roh); geklammert {
		return rumpf
	}

	return roh
}
