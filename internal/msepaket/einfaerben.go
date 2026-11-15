package msepaket

import "strings"

// **Die Einfärbung einer Paketdatei** —
// #1181.
//
// # Der Befund
//
// Der Vorlagenbereich färbte **gar nicht**. `Quelltextfeld` nimmt eine
// Sprache, und dort stand `keine`: Zeilennummern ja, Farbe nein.
//
// Das war richtig herum geirrt — eine `.mse-style` ist das **Paketformat**
// und nicht die Sprache: `game: magic` ist ein Schlüssel mit einem Wert, kein
// Ausdruck. Der Skript-Lexer liest daraus Bezeichner und Doppelpunkte und
// färbt entsprechend nichts Sinnvolles, und keine Farbe ist besser als eine
// falsche.
//
// # Warum das durch den Zerteiler geht
//
// **`zerteiler.go` liest dieses Format schon** — samt Einrückung,
// mehrzeiligen Werten und der Frage, wo ein Schlüssel aufhört. Ein zweiter
// Leser daneben liefe an dem Tag auseinander, an dem einer von beiden eine
// Regel lernt (§10).
//
// # Die vier Arten
//
//   - **Kommentar** — `#` bis zum Zeilenende
//   - **Schlüssel** — der Teil vor dem `:`
//   - **Wert** — der Rest
//   - **Skriptwert** — ein Wert, dessen Schlüssel MSE-Script erwartet
//
// # Warum die vierte hier nur benannt und nicht gefärbt wird
//
// **Weil `msepaket` `msescript` nicht kennen darf** (§2): es ist ein Blatt,
// und der Schichtenwächter weist den Import ab. Wer sie zusammensetzt, ist
// `internal/service` — es kennt beide, und dort steht `Paketfarben`.
//
// Das ist die richtige Trennung und nicht bloß die erlaubte: **was ein
// Skriptwert ist**, weiß das Paketformat; **wie ein Skript aussieht**, weiß
// die Sprache. Ein `msepaket`, das den Lexer riefe, wäre die Vermischung, vor
// der §2 bei `msescript` ausdrücklich warnt.

// Farbmerkmal ist ein Stück Quelltext mit seiner Art und seiner Stelle.
//
// **Zeile und Spalte zählen ab eins**, wie überall in diesem Programm.
type Farbmerkmal struct {
	Art    Farbart
	Zeile  int
	Spalte int
	Breite int

	// Text ist der Inhalt — **bei [FarbeSkriptwert] und [FarbeTextskriptwert]**.
	//
	// Wer ihn weiterfärbt, braucht ihn; ihn an jedem Merkmal mitzuführen wäre
	// die Datei ein zweites Mal.
	//
	// **Bei einem Rumpf über mehrere Zeilen sind sie mit `\n` verbunden** und
	// tragen die Einrückung des Paketformats nicht mehr — in MSE-Script
	// bedeutet sie nichts. Wohin jede Zeile in der Datei gehört, sagt
	// [Farbmerkmal.Zeilen].
	Text string

	// Zeilen gibt zu jeder Zeile von [Farbmerkmal.Text] ihre Stelle in der
	// **Datei** — **bei [FarbeSkriptwert] und [FarbeTextskriptwert]**.
	//
	// # Warum das nicht aus Zeile und Spalte folgt
	//
	// Weil die Einrückung wegfällt. Zeile 2 des Rumpfes steht in der Datei um
	// einen oder zwei Tabulatoren eingerückt, und ohne diese Liste zeigte
	// jede Fehlerstelle darin auf die falsche Spalte.
	//
	// **Auch der einzeilige Fall trägt sie** — mit genau einem Eintrag. Ein
	// Sonderfall weniger ist eine Verzweigung weniger bei jedem Aufrufer.
	Zeilen []Skriptzeile
}

// Farbart benennt, was ein Stück ist.
type Farbart string

const (
	// FarbeKommentar ist eine Zeile, die mit `#` beginnt.
	FarbeKommentar Farbart = "Kommentar"

	// FarbeSchluessel ist der Teil vor dem `:`.
	FarbeSchluessel Farbart = "Schlüssel"

	// FarbeWert ist der Rest der Zeile.
	FarbeWert Farbart = "Wert"

	// FarbeSkriptwert ist ein Wert, dessen Schlüssel MSE-Script erwartet.
	//
	// **Er wird hier nur benannt, nicht gefärbt** — warum, steht im Dateikopf.
	FarbeSkriptwert Farbart = "Skriptwert"

	// FarbeTextskriptwert ist ein Wert, den MSE2 als **Zeichenkette mit
	// eingebetteten Ausdrücken** liest — ohne Anführungszeichen, `{…}` ist ein
	// Ausdruck, `\{` eine Klammer. Zwei Stellen tun das
	// (#1934):
	//
	//   - ein Bildwert mit `{` in der Mitte — `parse(*this, true)` in
	//     `Reader::handle(ScriptableImage&)`, `src/script/image.cpp` Zeile 73,
	//     Fassung f9d9356 vom 2021-07-06,
	//   - `script:` unter `refer script:` — ein `StringScript`,
	//     `src/data/keyword.hpp` Zeile 30, dieselbe Fassung.
	//
	// **Gefärbt wird er wie ein Wert**; geprüft wird er streng
	// (`msescript.TextskriptFehler`).
	FarbeTextskriptwert Farbart = "Textskriptwert"
)

// skriptschluessel sind die Schlüssel, deren Wert MSE-Script ist.
//
// **Gezählt und nicht geraten**: `docs/FORMATE.md` führt sie, und der
// Zerteiler behandelt ihren Wert schon als Skriptrumpf. `default:` steht
// dabei, weil es eines **ist** — alle 142 Vorgabewerte von
// `magic.mse-game/card_fields` beginnen mit `stylesheet; `, und das ist eine
// Folge, keine Zeichenkette.
// **In der kanonischen Form, mit Unterstrich** — nachgeschlagen wird
// `Kanonisch(schluessel)`; in der Form der Datei fänden sich `init script`,
// `choice images`, `game script` und jeder andere Schlüssel aus zwei Wörtern
// nie (U7, #1418).
var skriptschluessel = map[string]bool{
	"script":             true,
	"default":            true,
	"init_script":        true,
	"game_script":        true,
	"sort_script":        true,
	"visible":            true,
	"condition":          true,
	"enabled":            true,
	"generate_script":    true,
	"filter":             true,
	"style_script":       true,
	"text_filter":        true,
	"variation":          true,
	"card_list_visible":  true,
	"tagged":             true,
	"editor_script":      true,
	"export_script":      true,
	"selection_choices":  true,
	"symbol_variation":   true,
	"positioning_script": true,
}

// bildschluessel sind die Schlüssel, deren Wert ein **Bild** ist — und die
// Einträge unter `choice images:` sind es auch.
//
// **Ein Bildwert ist ein Dateiname, bis er geklammert ist.** MSE2 liest ihn
// als Skript, wenn er mit `script:` beginnt oder eine Klammer trägt, sonst als
// Dateinamen (`Reader::handle(ScriptableImage&)`, `src/script/image.cpp`
// Zeilen 67–78, Fassung f9d9356 vom 2021-07-06); der Zeichenweg dieses
// Programms fragt, ob er mit `{` beginnt (`bildangabe` in
// `service/stilbild.go`), und `script:` steht im Pack nur als Unterblock.
// Hier standen `image`, `mask` und `choice_images` unter den
// Skriptschlüsseln, und `/magic-genevensis-00-main.mse-style/pride_mask.png`
// oder `hell: hell.png` galten der Werkstatt als kaputtes Skript — gefunden an
// der Hüllenprobe `fmt02` (FMT-02, #1691).
var bildschluessel = map[string]bool{
	"image": true,
	"mask":  true,
}

// Einfaerben zerlegt eine Paketdatei in ihre Farbmerkmale.
//
// **Ohne Fehler.** Eine Datei, die sich nicht zerteilen lässt, färbt so weit,
// wie sie lesbar ist — dieselbe Haltung wie bei [msescript.Einfaerben]. Wo ein
// Fehler steht, sagt der Zerteiler; das Fenster fragt ohnehin beides.
func Einfaerben(quelle string) []Farbmerkmal {
	baum, err := Zerteile("einfärben", []byte(quelle))
	if err != nil || baum == nil {
		return nil
	}

	var merkmale []Farbmerkmal

	sammle(baum.Wurzel, &merkmale, "")

	return merkmale
}

// sammle geht den Baum durch und legt je Knoten seine Merkmale ab.
//
// **Rekursiv und in der Reihenfolge der Datei** — die Anzeige legt die
// Abschnitte hintereinander, und eine Umsortierung machte aus ihr ein Rätsel.
//
// **`eltern` ist der kanonische Schlüssel des Blocks, in dem die Knoten
// stehen** — unter `choice images:` ist ein Eintrag eine Wahl mit einem Bild,
// unter `refer script:` ist `script:` ein Textskript.
func sammle(knoten []*Knoten, nach *[]Farbmerkmal, eltern string) {
	for _, eins := range knoten {
		*nach = append(*nach, merkmaleVon(eins, eltern)...)

		// **Unter einem Skriptschlüssel wird nicht weitergegangen.** Sein
		// Rumpf ist **ein** Wert, und [skriptwert] hat ihn schon ganz
		// genommen; wer hier weiterliefe, färbte jede seiner Zeilen ein
		// zweites Mal — als gewöhnlichen Text, über der richtigen Farbe.
		if skriptschluessel[Kanonisch(eins.Schluessel)] {
			continue
		}

		sammle(eins.Kinder, nach, Kanonisch(eins.Schluessel))
	}
}

// merkmaleVon gibt die Farbmerkmale eines einzelnen Knotens.
func merkmaleVon(knoten *Knoten, eltern string) []Farbmerkmal {
	zeile, spalte := knoten.Stelle.Zeile, knoten.Stelle.Spalte

	switch knoten.Zeilenart {
	case ArtLeer:
		// **Eine leere Zeile bekommt kein Merkmal.** Es gibt nichts zu färben,
		// und ein Abschnitt der Breite null legte sich über den Umbruch.
		return nil

	case ArtKommentar:
		return []Farbmerkmal{{
			Art: FarbeKommentar, Zeile: zeile, Spalte: spalte,
			Breite: len([]rune(knoten.Wert)),
		}}

	case ArtFortsetzung:
		// **Sie gehört zum Wert des Blocks über ihr.**
		//
		// **Ohne Prüfung auf leer**: eine Fortsetzung aus lauter Zwischenraum
		// gibt es nicht — der Zerteiler nennt sie [ArtLeer], und die steht
		// oben. Ein Zweig dafür wäre eine Behauptung darüber, was im Ernstfall
		// geschieht, die kein Lauf betreten kann (§3).
		return []Farbmerkmal{{
			Art: FarbeWert, Zeile: zeile, Spalte: spalte,
			Breite: len([]rune(knoten.Wert)),
		}}
	}

	// **Der letzte Fall ohne `case`.** `ArtEintrag` ist die vierte und letzte
	// Zeilenart; ein `default` daneben wäre ein Zweig, den kein Lauf betritt.
	return eintragsmerkmale(knoten, zeile, spalte, eltern)
}

// eintragsmerkmale gibt Schlüssel und Wert einer `schlüssel: wert`-Zeile.
func eintragsmerkmale(knoten *Knoten, zeile, spalte int, eltern string) []Farbmerkmal {
	// **Platz für zwei**: Schlüssel und Wert, mehr wird eine Eintragszeile nie.
	merkmale := make([]Farbmerkmal, 0, 2)
	merkmale = append(merkmale, Farbmerkmal{
		Art: FarbeSchluessel, Zeile: zeile, Spalte: spalte,
		Breite: len([]rune(knoten.Schluessel)),
	})

	// **Der Skriptschlüssel wird vor dem leeren Wert geprüft.** Bei einem
	// Rumpf über mehrere Zeilen steht hinter dem Doppelpunkt nichts, und der
	// Rücksprung darunter ließ genau die Form fallen, um die es geht.
	schluessel := Kanonisch(knoten.Schluessel)
	bildwert := eltern == "choice_images"

	// **Unter `refer script:` ist `script:` ein Textskript** — MSE2 führt es
	// als `StringScript` (`src/data/keyword.hpp` Zeile 30, Fassung f9d9356 vom
	// 2021-07-06): Text mit eingebetteten Ausdrücken, `\{` ist eine Klammer.
	// Als Skript zerteilt meldete jede Parameterart der Saat einen Fehler.
	textskript := eltern == "refer_script" && schluessel == "script"

	if !bildwert && !textskript && skriptschluessel[schluessel] {
		return append(merkmale, skriptwert(knoten, zeile, wertspalte(knoten, spalte)))
	}

	if textskript {
		return append(merkmale, textskriptwert(knoten, zeile, wertspalte(knoten, spalte)))
	}

	// **Jede `{` genügt, auch eine geschützte** — MSE2 fragt
	// `find_first_of('{')`, und `\{` ist danach eine Klammer im Namen.
	if (bildwert || bildschluessel[schluessel]) && strings.Contains(knoten.Wert, "{") {
		return append(merkmale, bildskriptwert(knoten, zeile, wertspalte(knoten, spalte)))
	}

	if strings.TrimSpace(knoten.Wert) == "" {
		return merkmale
	}

	return append(merkmale, Farbmerkmal{
		Art: FarbeWert, Zeile: zeile, Spalte: wertspalte(knoten, spalte),
		Breite: len([]rune(knoten.Wert)),
		Text:   knoten.Wert,
	})
}

// textskriptwert ist ein [skriptwert] der Art [FarbeTextskriptwert].
func textskriptwert(knoten *Knoten, zeile, wertspalte int) Farbmerkmal {
	merkmal := skriptwert(knoten, zeile, wertspalte)
	merkmal.Art = FarbeTextskriptwert

	return merkmal
}

// bildskriptwert gibt einen Bildwert mit Klammer — ganz geklammert ein Skript,
// sonst ein Textskript.
//
// **Ganz geklammert ist ein Skript**, wie im Zeichenweg (`skriptrumpf` in
// `service/stilbild.go`). MSE2 läse auch ihn im Zeichenkettenmodus; bei
// **einem** Ausdruck ist das Ergebnis dasselbe. Gemessen am 2026-09-29 über
// `full-magic-pack`: 5.191 Bildwerte beginnen mit `{` und enden auf `}`, und
// keiner besteht aus mehr als einem Ausdruck.
func bildskriptwert(knoten *Knoten, zeile, wertspalte int) Farbmerkmal {
	wert := strings.TrimSpace(knoten.Wert)
	if strings.HasPrefix(wert, "{") && strings.HasSuffix(wert, "}") {
		return skriptwert(knoten, zeile, wertspalte)
	}

	return textskriptwert(knoten, zeile, wertspalte)
}

// wertspalte gibt die Spalte, in der der Wert einer Eintragszeile beginnt.
//
// **Aus der rohen Zeile gelesen und nicht gerechnet**: zwischen `:` und Wert
// dürfen Leerzeichen stehen, und wie viele, weiß nur die Datei.
func wertspalte(knoten *Knoten, spalte int) int {
	vorspann := len([]rune(knoten.Schluessel)) + 1
	roh := []rune(knoten.Roh)[spalte-1:]

	for vorspann < len(roh) && (roh[vorspann] == ' ' || roh[vorspann] == '\t') {
		vorspann++
	}

	return spalte + vorspann
}

// skriptwert baut das Merkmal eines Skriptwerts — **beide Formen**.
//
// Ein Skript steht entweder hinter seinem Doppelpunkt oder eingerückt
// darunter. Bis zum 2026-09-02 sah diese Datei nur die erste Form; die
// zweite kam als Folge von [ArtFortsetzung]-Zeilen an und wurde wie
// gewöhnlicher Text gefärbt. Im vollen Pack ist das die Form von **2.240**
// der **5.553** Skriptwerte — gezählt an Zeilen mit einem `script:`-
// Schlüssel, und damit über Vorlagen hinweg, nicht über eine.
func skriptwert(knoten *Knoten, zeile, wertspalte int) Farbmerkmal {
	zeilen := knoten.Skriptzeilen()

	if len(zeilen) == 0 {
		return Farbmerkmal{Art: FarbeSkriptwert, Zeile: zeile, Spalte: wertspalte}
	}

	texte := make([]string, 0, len(zeilen))
	for _, eine := range zeilen {
		texte = append(texte, eine.Text)
	}

	return Farbmerkmal{
		Art: FarbeSkriptwert, Zeile: zeilen[0].Zeile, Spalte: zeilen[0].Spalte,
		Breite: len([]rune(zeilen[0].Text)),
		Text:   strings.Join(texte, "\n"),
		Zeilen: zeilen,
	}
}
