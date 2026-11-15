package msepaket

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Exportvorlage ist der Inhalt eines `.mse-export-template`.
//
// # Was hier roh bleibt
//
// **Das Skript.** Eine Exportvorlage ist selbst skriptgesteuert: was am Ende
// herauskommt, ist der Wert ihres letzten Ausdrucks. Dieser Leser gibt den Text
// heraus; was er bedeutet, weiß `msescript`, und wer ihn auswertet, ist
// `service` — dieselbe Naht wie beim Stil (AGENTS.md §2).
type Exportvorlage struct {
	// Kartenfelder sind die `card field:`-Blöcke des **Spielpakets**.
	//
	// **Sie stehen hier aus demselben Grund wie [Exportvorlage.Spielskript]**:
	// nicht im Paket der Vorlage, sondern in dem des Spiels; wer die Vorlage
	// liest, legt sie dazu.
	//
	// # Wozu die Ausfuhr sie braucht
	//
	// **In MSE trägt jede Karte jedes Feld ihres Spiels** — der Konstruktor
	// legt für jeden erklärten Schlüssel einen Wert an:
	//
	//	Card::Card() … { data.init(game_for_reading()->card_fields); }
	//
	// `src/data/card.cpp`, Zeilen 20–30, und `IndexMap::init`
	// (`src/util/index_map.hpp`, Zeilen 48–62), Fassung f9d9356 vom
	// 2021-07-06. `card.notes` gibt es dort **immer**, leer, wenn die Datei
	// nichts geschrieben hat.
	//
	// Ohne diese Liste fehlte ein ausgelassenes Feld in der Abbildung, und
	// seit ein fehlendes Glied ein **Fehler** ist statt `nil`, scheiterte
	// daran jede Vorlage, die es liest
	// (#609).
	Kartenfelder []domain.Field

	// Symbolfassungen sind die `variation:`-Blöcke des **Spielpakets** — die
	// Farben, in denen `symbol_variation` das Setsymbol zeichnet.
	//
	// **Sie stehen hier aus demselben Grund wie [Exportvorlage.Spielskript]**:
	// nicht im Paket der Vorlage, sondern in dem des Spiels. Beim Zeichnen
	// einer Karte kommen sie über den **Stil** dazu (`stilumgebung.go`, #323);
	// in der Ausfuhr gibt es keinen Stil
	// (#660).
	Symbolfassungen []Symbolfassung

	// Setfelder sind die `set field:`-Blöcke des Spielpakets.
	//
	// **Dieselbe Regel wie an der Karte, eine Ebene höher**: MSEs `Set` legt
	// für jedes erklärte Feld einen Wert an —
	//
	//	Set::Set(const GameP& game) … { data.init(game->set_fields); }
	//
	// `src/data/set.cpp`, Zeilen 33–39, dieselbe Fassung. `set.title` gibt es
	// dort immer, leer, wenn niemand etwas eingetragen hat.
	Setfelder []domain.Field

	// Spielskript ist das `init script:` des Spielpakets, samt seinen Modulen.
	//
	// **Leer, solange niemand es einsetzt.** Es steht nicht im Paket der
	// Vorlage, sondern in dem des Spiels; wer die Vorlage liest, legt es dazu
	// (#216).
	Spielskript string

	// Paket ist der Name des Pakets, aus dem die Vorlage stammt.
	Paket string

	// Ort ist der **Pfad** dieses Pakets.
	//
	// **`copy_file` liest daraus** — die Funktion holt eine Datei aus dem
	// Vorlagenpaket selbst und legt sie neben die Ausgabe
	// (`src/script/functions/export.cpp`, Zeilen 369–383, Fassung f9d9356 vom
	// 2021-07-06). Ohne den Ort wüsste niemand, welches Paket gemeint ist:
	// [Exportvorlage.Paket] ist ein Name und kann in mehreren Wurzeln stehen
	// (#657).
	Ort string

	Spiel      string
	KurzerName string
	VollerName string

	// Dateityp ist der Filter des Dateidialogs, wie MSE ihn schreibt:
	// `*.xml|*.xml|*.*|*.*`. Vier der 19 Vorlagen des Packs führen `*.txt`,
	// die übrigen `*.xml` oder `*.dat`.
	Dateityp string

	// Verzeichnis sagt, ob die Ausfuhr ein Verzeichnis anlegt statt einer
	// Datei (`create directory:`). Zehn der Vorlagen sagen ja, vier nein.
	Verzeichnis bool

	// Skript ist der Text des `script:`-Blocks.
	Skript string

	Optionen []Exportoption
}

// Exportoption ist ein `option field:`-Block — eine Eingabe des Skripts.
//
// Das Skript liest sie als `options.<name>`, mit Unterstrichen statt
// Leerzeichen: `options.nur_namen` für `name: nur namen`. Dieselbe Regel wie
// bei den Kartenfeldern (`docs/FORMATE.md`).
type Exportoption struct {
	Art          string
	Name         string
	Beschreibung string

	// Vorgabe ist, was gilt, wenn niemand wählt.
	//
	// **Zwei Schlüssel, ein Sinn**: eine Auswahl schreibt `initial:`, ein Text
	// `default:`. Beide stehen in den Vorlagen des Packs.
	Vorgabe string

	// Auswahl sind die Möglichkeiten einer `choice`.
	Auswahl []string
}

// Exportmodell liest eine Exportvorlage aus ihrem Paket.
func (p *Paket) Exportmodell() Exportvorlage {
	vorlage := Exportvorlage{
		Paket:       paketname(p.Pfad),
		Ort:         p.Pfad,
		Spiel:       kopfwert(p.Kopf, "game"),
		KurzerName:  p.KurzerName,
		VollerName:  p.VollerName,
		Dateityp:    kopfwert(p.Kopf, "file type"),
		Verzeichnis: kopfwert(p.Kopf, "create directory") == "true",
	}

	if knoten := p.Kopf.Ersten("script"); knoten != nil {
		vorlage.Skript = knoten.Skripttext()
	}

	for _, block := range p.Kopf.Alle("option field") {
		vorlage.Optionen = append(vorlage.Optionen, optionAus(block))
	}

	return vorlage
}

// optionAus liest einen `option field:`-Block.
func optionAus(knoten *Knoten) Exportoption {
	option := Exportoption{}

	for _, kind := range knoten.Kinder {
		if kind.Zeilenart != ArtEintrag {
			continue
		}

		// **Kanonisch wie beim Feldstil**, auch wenn hier jeder Schlüssel ein
		// Wort ist — ein zweiter, der dazukommt, soll nicht roh verglichen
		// werden (#1972).
		switch Entkanonisch(kind.Schluessel) {
		case "type":
			option.Art = kind.Wert
		case "name":
			option.Name = kind.Wert
		case "description":
			option.Beschreibung = kind.Wert
		case "choice":
			option.Auswahl = append(option.Auswahl, wahlname(kind))

		// **`default` und `initial` meinen dasselbe.** Ein Text schreibt das
		// eine, eine Auswahl das andere; beide stehen in den Vorlagen des
		// Packs, und beide sagen, was gilt, wenn niemand wählt.
		case "default":
			option.Vorgabe = skriptzeichenkette(kind.Wert)
		case "initial":
			option.Vorgabe = kind.Wert
		}
	}

	// **Eine boolesche Option hat zwei Wahlen: `yes` und `no`.**
	//
	//	BooleanField::BooleanField() {
	//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
	//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
	//
	// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom
	// 2021-07-06. Dieselbe Regel wie an `spiel.go`
	// (#456)
	// und an `stil.go`
	// (#531) —
	// eine Exportvorlage führt in `option_fields` **dieselbe** Feldklasse
	// (`src/data/export_template.hpp`, Zeile 32).
	//
	// **Sie steht hinter der Schleife**, damit ein ausdrückliches `choice:`
	// im Block sie nicht verdoppelt — im Pack schreibt keines eines, und
	// falls doch, gewinnt das Geschriebene.
	if option.Art == "boolean" && len(option.Auswahl) == 0 {
		option.Auswahl = []string{"yes", "no"}
	}

	return option
}

// Anfangswert gibt, was für diese Option gilt, wenn niemand wählt.
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : Value(field)
//	  , value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	         :                           _("")
//	         , true)
//
// `src/data/field/choice.cpp`, Zeilen 279–285, mit
// `initial_first_choice = true` als Vorgabe (`choice.hpp`, Zeile 192).
// Fassung f9d9356 vom 2021-07-06.
//
// **Getrennt von [Exportoption.Vorgabe]**, weil beides nicht dasselbe ist:
// dort steht, was die Datei sagt, hier, was gilt. Ein Leser, der nur das
// Feld sähe, hielte eine Option ohne `initial:` für unbelegt.
//
// **Über das Pack**: 7 boolesche Optionen ohne `default:`/`initial:` und 21
// `choice`-Optionen ohne. Sie standen alle leer da
// (#555).
func (o Exportoption) Anfangswert() string {
	if o.Vorgabe != "" {
		return o.Vorgabe
	}

	if len(o.Auswahl) > 0 {
		return o.Auswahl[0]
	}

	// **Eine boolesche Option kommt auch ohne Wahlliste hierher.**
	// [optionAus] legt sie an, aber die Zusage gehört an die Stelle, die sie
	// einlöst: ein anderswo gebautes [Exportoption] gäbe sonst still die
	// leere Zeichenkette, wo MSE `yes` liest (§10).
	if o.Art == "boolean" {
		return "yes"
	}

	return ""
}

// skriptzeichenkette gibt den Wert eines `default:`, der eine Zeichenkette in
// Anführungszeichen ist — und alles andere, wie es dasteht.
//
// **`default:` ist bei MSE2 ein Skript** (`REFLECT_N("default",
// default_script)`, `src/data/field/text.cpp` Zeile 34, Fassung f9d9356 vom
// 2021-07-06): `"Meine Karten"` ist eine Zeichenkette, `Meine Karten` zwei
// Namen und damit kein gültiges Skript. Dieser Leser wertet nichts aus; er
// nimmt den Inhalt einer Zeichenkette und lässt einen Namen wie `Custom`
// wörtlich stehen, weil er ihn nicht nachschlagen kann.
func skriptzeichenkette(wert string) string {
	if len(wert) < 2 || !strings.HasPrefix(wert, `"`) || !strings.HasSuffix(wert, `"`) {
		return wert
	}

	var inhalt strings.Builder

	zeichen := []rune(wert[1 : len(wert)-1])
	for stelle := 0; stelle < len(zeichen); stelle++ {
		if zeichen[stelle] == '\\' && stelle+1 < len(zeichen) {
			stelle++
		}

		inhalt.WriteRune(zeichen[stelle])
	}

	return inhalt.String()
}
