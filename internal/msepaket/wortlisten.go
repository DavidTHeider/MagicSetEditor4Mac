package msepaket

// Die `word list:`-Blöcke eines Spiels
// (#1074).
//
// # Der Anlass
//
// `magic.mse-game/word_lists` führt **51** Listen mit **432** Wörtern auf
// oberster Ebene und **445** in Untergruppen — gemessen am 2026-08-24 am
// bezogenen `M15-Magic-Pack`. Wer in MSE2 auf ein Typfeld klickt, wählt daraus;
// hier tippte man den Typ ab und hoffte auf die richtige Schreibung.
//
// # Was hier steht und was nicht
//
// **Hier steht, was in der Datei steht.** Welche Stelle eines Feldwerts zu
// welcher Liste gehört, steht im Wert selbst — als `<word-list-NAME>`-Marke —
// und wird in `service` gesucht: dort ist die Skriptauswertung, die ein Wort
// mit `script:` braucht.

// Wortliste ist eine benannte Auswahlliste eines Spiels.
//
// **Sie trägt nur Namen und Wörter.** `IMPLEMENT_REFLECTION_NO_SCRIPT(WordList)
// { REFLECT(name); REFLECT(words); }` — `src/data/word_list.cpp`, Zeilen 32–35,
// Fassung f9d9356 vom 2021-07-06. `line below`, `is prefix` und `script` gibt
// es erst an den Wörtern.
type Wortliste struct {
	Name    string
	Woerter []Wortlisteneintrag
}

// Wortlisteneintrag ist ein Wort in einer Liste.
//
// Nachgelesen an `WordListWord`, `src/data/word_list.hpp`, Zeilen 22–34,
// Fassung f9d9356 vom 2021-07-06.
type Wortlisteneintrag struct {
	// Name ist das Wort, wie es eingesetzt wird.
	//
	// **Ein Leerzeichen am Ende gehört dazu.** Im Pack heißt es `Legendary `
	// und nicht `Legendary`, denn ein Präfix wird vor das nächste Wort
	// gesetzt; wer es abschnitte, schriebe `LegendaryCreature`.
	Name string

	// Skript setzt an die Stelle des Namens eine **gerechnete** Liste.
	//
	// `OptionalScript script` — „Generate words using a script"
	// (`word_list.hpp`, Zeile 29). Das Ergebnis ist eine Zeichenkette mit
	// Kommas darin; zerlegt und sortiert wird sie in `service`.
	Skript string

	// IstPraefix macht aus dem Eintrag ein **Kästchen** statt eines Punktes:
	// er wird vor das gewählte Wort gesetzt und lässt sich mit anderen
	// verbinden.
	IstPraefix bool

	// ZeileDarunter zieht im Menü einen Strich unter den Eintrag.
	ZeileDarunter bool

	// Woerter macht den Eintrag zu einer **Gruppe** — einem Untermenü.
	//
	// `inline bool isGroup() const { return !words.empty(); }`
	// (`word_list.hpp`, Zeile 31).
	Woerter []Wortlisteneintrag
}

// IstGruppe sagt, ob der Eintrag ein Untermenü öffnet.
func (e Wortlisteneintrag) IstGruppe() bool {
	return len(e.Woerter) > 0
}

// Wortlisten gibt die `word list:`-Blöcke des Kopfes.
//
// **Ohne Namen wird übergangen.** Eine Liste wird über ihren Namen gefunden
// (`findWordLists`, `src/gui/value/text.cpp`, Zeilen 1362–1367, Fassung f9d9356
// vom 2021-07-06); eine namenlose ist keiner Marke zuzuordnen.
func (p *Paket) Wortlisten() []Wortliste {
	knoten := p.Kopf.Alle("word list")

	listen := make([]Wortliste, 0, len(knoten))

	for _, eine := range knoten {
		name := WertOderBlock(eine.Ersten("name"))
		if name == "" {
			continue
		}

		listen = append(listen, Wortliste{Name: name, Woerter: woerterAus(eine)})
	}

	return listen
}

// woerterAus liest die `word:`-Kinder eines Knotens.
//
// **Ein bloßer Wert ist der Name.** `REFLECT_IF_READING_SINGLE_VALUE {
// REFLECT_NAMELESS(name); }` — `src/data/word_list.cpp`, Zeilen 19–29, Fassung
// f9d9356 vom 2021-07-06. `word: Creature` und ein Block mit `name: Creature`
// bedeuten dasselbe.
func woerterAus(knoten *Knoten) []Wortlisteneintrag {
	kinder := knoten.Alle("word")

	woerter := make([]Wortlisteneintrag, 0, len(kinder))

	for _, kind := range kinder {
		praefix, _ := wahrheitMit(kind, "is prefix", false)
		darunter, _ := wahrheitMit(kind, "line below", false)

		woerter = append(woerter, Wortlisteneintrag{
			Name:          namenAus(kind),
			Skript:        WertOderBlock(kind.Ersten("script")),
			IstPraefix:    praefix,
			ZeileDarunter: darunter,
			Woerter:       woerterAus(kind),
		})
	}

	return woerter
}

// namenAus gibt den Namen eines Wortes — den bloßen Wert oder das `name:`-Feld.
func namenAus(knoten *Knoten) string {
	if knoten.Wert != "" {
		return knoten.Wert
	}

	return WertOderBlock(knoten.Ersten("name"))
}
