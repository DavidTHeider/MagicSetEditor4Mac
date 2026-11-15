package msepaket

// **`add cards script:` — Karten aus einem Skript**
// (#1337,
// Werkstück B von Phase 4.11).
//
// # Was nachgelesen wurde (§15)
//
// `src/data/add_cards_script.hpp`, Zeilen 12–26 (Fassung f9d9356 vom
// 2021-07-06):
//
//	class AddCardsScript : public IntrusivePtrBase<AddCardsScript> {
//	public:
//	  String           name;
//	  String           description;
//	  Scriptable<bool> enabled;
//	  OptionalScript   script;
//
// **`enabled` ist selbst ein Skript.** Es entscheidet je Set, ob der Eintrag
// zur Wahl steht — `magic.mse-game` nimmt das für die Zyklen, die nur bei
// einem passenden Stil sinnvoll sind.
//
// # Was das Skript zurückgibt
//
// Eine **Liste** von Karten. `AddCardsScript::perform`
// (`src/data/add_cards_script.cpp`, Zeilen 20–35) iteriert über das Ergebnis
// und nimmt jeden Eintrag als Karte; **eine Karte, die schon im Set liegt,
// wird kopiert** statt geteilt. Das steht dort so:
//
//	if (contains(set.cards,card) || contains(out,card)) {
//	  card = make_intrusive<Card>(*card);
//	}
//
// Das Ausführen gehört nach `internal/service` — hier steht nur, was in der
// Datei steht.
//
// # Und warum es hier steht und nicht in der Domäne
//
// **Anti-Korruptionsschicht** (§2, §16): [Kartenskript] ist die Gestalt der
// **fremden Datei** — mit ihrem `&`-Kürzel und ihrem Skripttext. Was `service`
// davon braucht, bekommt es als Domänentyp; eine MSE2-Struktur, die bis in die
// Fachlogik durchschlüge, machte jede Änderung an ihrem Format zu einer
// Änderung überall.

import "strings"

// Kartenskript ist ein `add cards script:`-Block.
type Kartenskript struct {
	// Name ist der angezeigte Name — **ohne** das `&` des Tastenkürzels.
	Name string

	// Kuerzel ist der Buchstabe, den das `&` markiert hat — leer, wenn keines
	// dastand.
	//
	// **Getrennt vom Namen und nicht darin gelassen**: „&Basic Lands" in einem
	// Menü, das keine wxWidgets-Kürzel kennt, wäre ein sichtbares Kaufmanns-Und.
	Kuerzel string

	// Beschreibung ist der Satz darunter — was der Eintrag tut.
	Beschreibung string

	// Bedingung ist `enabled:` — ein **Skript**, kein Wahrheitswert.
	//
	// **Leer heißt: immer.** So ist es im Altcode; ein Eintrag, der ohne Grund
	// fehlte, wäre schlimmer als einer zu viel.
	Bedingung string

	// Skript ist der Ausdruck, der die Karten gibt.
	Skript string
}

// Kartenskripte liest die `add cards script:`-Blöcke eines Spielpakets.
//
// **Über den ganzen Kopf und nicht über eine Datei.** `magic.mse-game` legt
// sie in `add_cards_scripts` und bindet die Datei über `include file:` ein;
// ein Leser, der `game` allein läse, fände dort keinen einzigen Block.
func (p *Paket) Kartenskripte() []Kartenskript {
	bloecke := p.Kopf.Alle("add cards script")
	skripte := make([]Kartenskript, 0, len(bloecke))

	for _, knoten := range bloecke {
		name, kuerzel := nameUndKuerzel(wertOderLeer(knoten.Ersten("name")))

		skripte = append(skripte, Kartenskript{
			Name:         name,
			Kuerzel:      kuerzel,
			Beschreibung: wertOderLeer(knoten.Ersten("description")),
			Bedingung:    skripttextOderLeer(knoten.Ersten("enabled")),
			Skript:       skripttextOderLeer(knoten.Ersten("script")),
		})
	}

	return skripte
}

// nameUndKuerzel trennt das wxWidgets-Kürzel vom Namen.
//
// # Warum das hier geschieht und nicht in der Oberfläche
//
// Weil es eine Eigenschaft der **Datei** ist. `&Basic Lands` heißt bei
// wxWidgets: zeige „Basic Lands" und nimm ⌥B als Kürzel. Eine Fläche, die den
// rohen Namen bekäme, müsste dieselbe Regel noch einmal kennen — und die
// zweite Fassung liefe irgendwann auseinander.
//
// **`&&` ist ein echtes Kaufmanns-Und.** So entkommt wxWidgets ihm, und ein
// Name wie „Rock && Roll" soll auch einer bleiben.
func nameUndKuerzel(roh string) (name, kuerzel string) {
	var gebaut strings.Builder

	zeichen := []rune(roh)

	for i := 0; i < len(zeichen); i++ {
		if zeichen[i] != '&' {
			gebaut.WriteRune(zeichen[i])

			continue
		}

		// **Ein `&` am Ende ist keines**: es markiert nichts und gehört
		// trotzdem nicht in den Namen.
		if i+1 >= len(zeichen) {
			continue
		}

		if zeichen[i+1] == '&' {
			gebaut.WriteRune('&')
			i++

			continue
		}

		// **Nur das erste markiert.** wxWidgets nimmt ebenfalls das erste;
		// ein zweites wäre eine Angabe, die niemand auflösen kann.
		if kuerzel == "" {
			kuerzel = string(zeichen[i+1])
		}
	}

	return gebaut.String(), kuerzel
}
