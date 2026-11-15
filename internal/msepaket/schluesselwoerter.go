package msepaket

// Die `keyword:`-Blöcke eines Pakets
// (#1060).
//
// # Der Anlass
//
// `magic.mse-game/keywords` führt **334** Schlüsselwörter, **134** davon mit
// Parameter, dazu **sechs** Modi und mehrere Parameterarten — gemessen am
// 2026-08-24 am bezogenen `M15-Magic-Pack`. Keines davon erreichte bis dahin
// die Anwendung: ein Set mit Schlüsselwörtern überlebte den Rundlauf als
// Rohblock, sichtbar war keines.
//
// # Was hier steht und was nicht
//
// **Hier steht, was in der Datei steht.** Was ein Schlüsselwort mit einem
// Regeltext *tut* — den Erinnerungstext einsetzen —, steht in
// #1061 und
// gehört nach `service`: es braucht die Skriptauswertung, und ein Paketleser,
// der Skripte auswertete, wäre die Vermischung, die §2 verhindert.

// Schluesselwort ist ein Schlüsselwort eines Spiels oder eines Sets.
//
// # Die Felder
//
// Nachgelesen an `src/data/keyword.cpp:86–93` (`IMPLEMENT_REFLECTION(Keyword)`),
// Fassung f9d9356 vom 2021-07-06: `keyword`, `match`, `reminder`, `rules`,
// `mode`. Mehr steht nicht in der Datei; alles Übrige an `Keyword` — der
// Ausdruck, die Parameterliste, `valid` — wird beim Vorbereiten **gerechnet**.
type Schluesselwort struct {
	// Wort ist der Name, „only for human use" (`keyword.hpp:93`).
	Wort string

	// Muster ist `match` — der Text, auf den es passt, mit
	// `<atom-param>art</atom-param>` an den Stellen der Parameter.
	Muster string

	// Erinnerung ist der Erinnerungstext. **Ein Skript**, kein fester Text:
	// `StringScript reminder` (`keyword.hpp:95`), und im echten Pack steht
	// darin `{this_or_that("creature", upper:true)}`.
	Erinnerung string

	// Regeln ist `rules` — eine Erläuterung für Menschen, die nicht auf der
	// Karte erscheint.
	Regeln string

	// Modus ist der Name eines [Schluesselwortmodus].
	Modus string

	// Fest sagt, dass es aus dem **Spiel** kommt und nicht geändert werden
	// kann — `Keyword::fixed`, `keyword.hpp:104`: „true for game keywods,
	// false for set keywords".
	Fest bool
}

// Schluesselwortmodus ist eine Verwendungsart — `core`, `expert`, `action`.
//
// Nachgelesen an `src/data/keyword.cpp:50–54`. (Fassung f9d9356 vom 2021-07-06)
type Schluesselwortmodus struct {
	Name         string
	Beschreibung string

	// Vorgabe sagt, welchen Modus ein **neues** Schlüsselwort bekommt
	// (`keyword.hpp:81`). Im Pack ist das `custom`.
	Vorgabe bool
}

// Schluesselwortparameterart ist die Art eines Parameters — `mana`, `cost`.
//
// # Warum nur sechs Felder
//
// `KeywordParam` führt dreizehn (`src/data/keyword.cpp:35–49`). Hier stehen (Fassung f9d9356 vom 2021-07-06)
// die, für die es einen Verwender gibt — dieselbe Überlegung wie bei den
// Bulkeinträgen in §16. Was fehlt, ist benannt: `script`, `reminder_script`,
// `separator_script`, `example`, `refer_scripts` und `placeholder` gehören zur
// **Auswertung** und kommen mit
// #1061.
type Schluesselwortparameterart struct {
	Name         string
	Beschreibung string

	// Muster ist ein **regulärer Ausdruck**, kein Text — im Pack etwa
	// `[HSVCTQXYZI0-9WUBRG/|]+`.
	Muster string

	// TrennerDavor und TrennerDanach sind ebenfalls Ausdrücke
	// (`separator before is`, `separator after is`).
	TrennerDavor  string
	TrennerDanach string

	// Wahlfrei sagt, ob der Parameter fehlen darf.
	//
	// **Ohne Angabe wahr.**
	//
	//	KeywordParam::KeywordParam()
	//	  : optional(true)
	//	  , eat_separator(true)
	//
	// `src/data/keyword.cpp`, Zeilen 24–27, Fassung f9d9356 vom 2021-07-06.
	//
	// **Von 15 Parameterarten im echten Pack tragen 13 kein `optional:`**;
	// die Vorgabe entscheidet über fast alle.
	Wahlfrei bool

	// TrennerFressen sagt, ob der Trenner aus dem festen Text genommen wird,
	// statt zweimal dazustehen — `eat separator:`, **ohne Angabe wahr**
	// (ebd.).
	//
	// **Im echten Pack setzt es keine einzige Parameterart**; die Vorgabe
	// entscheidet in allen drei Arten mit Trenner.
	TrennerFressen bool

	// Verweisskripte sind die `refer script:`-Blöcke — die Formen, in denen
	// ein Erinnerungstext auf diesen Parameter zeigen kann
	// (#1342).
	//
	// **Leer heißt: nur die schlichte Form.** MSE bietet dann `{paramN}` an
	// (`KeywordsPanel::runRefScript`, `src/gui/set/keywords_panel.cpp`
	// Zeilen 236–261, Fassung f9d9356 vom 2021-07-06). **Gemessen am
	// 2026-08-31**: `magic.mse-game` bringt 15 Parameterarten mit, 6 davon
	// mit Verweisen, zusammen 19 Skripte — gezählt **nach** dem Auflösen von
	// `include localized file:`. Über alle Dateien des Packs sind es 161, 46
	// und 147; diese Zahl zählt jede Sprachfassung mit.
	Verweisskripte []Schluesselwortverweis
}

// Schluesselwortverweis ist ein `refer script:`-Block einer Parameterart.
type Schluesselwortverweis struct {
	Name         string
	Beschreibung string

	// Skript rechnet den einzufügenden Text — mit `input` auf `paramN`.
	//
	// **Roh und nicht ausgewertet.** `\{` ist MSEs Maskierung einer
	// geschweiften Klammer im Text; das Auswerten gehört nach `service`.
	Skript string
}

// Schluesselwoerter gibt die `keyword:`-Blöcke des Kopfes.
//
// **Ohne Muster wird übergangen.** `Keyword::prepare` endet mit
// `valid = !match_re.matches("")` — ein Schlüsselwort, dessen Ausdruck auf den
// leeren Text passt, wird abgeschaltet, „that would make MSE hang"
// (`src/data/keyword.cpp:207–208`). Ein leeres `match` ist genau dieser Fall. (Fassung f9d9356 vom 2021-07-06)
//
// **Fest heißt: aus einem Spielpaket.** Ein Set trägt eigene, und die sind
// änderbar; die Unterscheidung hängt an der Art des Pakets und nicht am Block.
func (p *Paket) Schluesselwoerter() []Schluesselwort {
	fest := p.Art == ArtSpiel

	woerter := make([]Schluesselwort, 0, len(p.Kopf.Alle("keyword")))

	for _, knoten := range p.Kopf.Alle("keyword") {
		muster := WertOderBlock(knoten.Ersten("match"))
		if muster == "" {
			continue
		}

		woerter = append(woerter, Schluesselwort{
			Wort:       WertOderBlock(knoten.Ersten("keyword")),
			Muster:     muster,
			Erinnerung: WertOderBlock(knoten.Ersten("reminder")),
			Regeln:     WertOderBlock(knoten.Ersten("rules")),
			Modus:      WertOderBlock(knoten.Ersten("mode")),
			Fest:       fest,
		})
	}

	return woerter
}

// Schluesselwortmodi gibt die `keyword mode:`-Blöcke des Kopfes.
func (p *Paket) Schluesselwortmodi() []Schluesselwortmodus {
	knoten := p.Kopf.Alle("keyword mode")
	modi := make([]Schluesselwortmodus, 0, len(knoten))

	for _, einer := range knoten {
		vorgabe, _ := wahrheitMit(einer, "is default", false)

		modi = append(modi, Schluesselwortmodus{
			Name:         WertOderBlock(einer.Ersten("name")),
			Beschreibung: WertOderBlock(einer.Ersten("description")),
			Vorgabe:      vorgabe,
		})
	}

	return modi
}

// Schluesselwortparameterarten gibt die `keyword parameter type:`-Blöcke.
func (p *Paket) Schluesselwortparameterarten() []Schluesselwortparameterart {
	knoten := p.Kopf.Alle("keyword parameter type")
	arten := make([]Schluesselwortparameterart, 0, len(knoten))

	for _, einer := range knoten {
		wahlfrei, _ := wahrheitMit(einer, "optional", true)
		fressen, _ := wahrheitMit(einer, "eat separator", true)

		arten = append(arten, Schluesselwortparameterart{
			Name:           WertOderBlock(einer.Ersten("name")),
			Beschreibung:   WertOderBlock(einer.Ersten("description")),
			Muster:         WertOderBlock(einer.Ersten("match")),
			TrennerDavor:   WertOderBlock(einer.Ersten("separator before is")),
			TrennerDanach:  WertOderBlock(einer.Ersten("separator after is")),
			Wahlfrei:       wahlfrei,
			TrennerFressen: fressen,
			Verweisskripte: verweisskripteAus(einer),
		})
	}

	return arten
}

// verweisskripteAus liest die `refer script:`-Blöcke einer Parameterart.
func verweisskripteAus(knoten *Knoten) []Schluesselwortverweis {
	bloecke := knoten.Alle("refer script")
	verweise := make([]Schluesselwortverweis, 0, len(bloecke))

	for _, block := range bloecke {
		verweise = append(verweise, Schluesselwortverweis{
			Name:         WertOderBlock(block.Ersten("name")),
			Beschreibung: WertOderBlock(block.Ersten("description")),
			Skript:       WertOderBlock(block.Ersten("script")),
		})
	}

	return verweise
}
