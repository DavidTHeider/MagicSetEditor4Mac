package domain

import "strings"

// FieldKind ist die Art eines Kartenfelds.
//
// Sie stammt aus dem Spielpaket, nicht von hier: ein MSE-Spiel definiert seine
// Felder selbst. Diese Aufzählung nennt nur die Arten, die die Anwendung
// unterscheiden **muss** — alles andere ist `KindUnknown` und wird als Text
// behandelt. Eine Art zu erfinden, die kein Paket kennt, hilft niemandem.
type FieldKind string

const (
	KindUnknown FieldKind = ""
	KindText    FieldKind = "text"
	KindChoice  FieldKind = "choice"
	KindImage   FieldKind = "image"
	KindColor   FieldKind = "color"
	KindBoolean FieldKind = "boolean"
	KindInfo    FieldKind = "info"

	// KindMultipleChoice ist `multiple choice` — mehrere Wahlen zugleich,
	// durch `, ` getrennt
	// (#556).
	//
	// **Sie ist keine Spielart von [KindChoice].** Ihr Wert steht bei MSE in
	// einer **Normalform**: der Reihenfolge, in der die Wahlen erklärt sind
	// (`MultipleChoiceValue::normalForm`,
	// `src/data/field/multiple_choice.cpp`, Zeilen 82–110, Fassung f9d9356
	// vom 2021-07-06). Im Magic-Spielpaket sind **26** Felder so erklärt.
	KindMultipleChoice FieldKind = "multiple choice"

	// KindPackageChoice ist `package choice` — die Wahl unter den
	// **installierten Paketen**, die auf ein Muster passen
	// (#632).
	//
	// **Die häufigste unbediente Art des Packs**: 627 Vorkommen in 349 von
	// 376 Stilpaketen (#630).
	// Bis dahin fiel sie als `KindUnknown` durch und stand als Textfeld da, in
	// das jemand einen Paketdateinamen hätte tippen müssen.
	KindPackageChoice FieldKind = "package choice"

	// KindSymbol ist `symbol` — eine **Zeichnung**, die als Datei im Paket
	// liegt (#1339).
	//
	// **Der Wert ist ein Dateiname, keine eingebettete Zeichnung.** So steht
	// es im Altcode (`src/data/field/symbol.hpp`, Fassung f9d9356 vom
	// 2021-07-06):
	//
	//	ValueType filename;  ///< Filename of the symbol (in the current package)
	//
	// Das Feld selbst trägt **keine eigenen Angaben** — dort steht wörtlich
	// „no extra data"; was es aussieht, entscheidet der Stil.
	//
	// **Gemessen am 2026-08-31** über `full-magic-pack`: `type: symbol` kommt
	// neunmal in sieben Dateien vor, und **jedes Mal als `set field:`**. Ein
	// Kartenfeld dieser Art hat im Pack keinen Vertreter — die Art gibt es
	// trotzdem, und die sechs Spielpakete mit einem Setsymbolfeld fielen hier
	// bis dahin als [KindUnknown] durch.
	KindSymbol FieldKind = "symbol"
)

// Field ist die Definition eines Kartenfelds aus einem Spielpaket.
type Field struct {
	Name    string
	Kind    FieldKind
	Default string
	Choices []string

	// Leerwahl ist der Wert einer [KindMultipleChoice], an der nichts gewählt
	// ist — `empty choice:`
	// (#2006).
	//
	//	// empty choice name
	//	if (val.empty()) { val = field().empty_choice; }
	//
	// `MultipleChoiceValue::normalForm`, `src/data/field/multiple_choice.cpp`,
	// Zeilen 110–113, Fassung f9d9356 vom 2021-07-06. **Ohne Angabe die leere
	// Zeichenkette.** `magic.mse-game` erklärt `colorless` an sieben
	// Kartenfeldern (`card color` bis `card color 4` und `indicator` bis `indicator 3`).
	Leerwahl string

	// Wahlfarben sind die `color:`-Angaben der `choice:`-Blöcke eines
	// **Farbfeldes** — sie stehen neben [Field.Choices], weil eine Wahl dort
	// einen Namen **und** eine Farbe hat.
	//
	//	value( !field->initial.isDefault() ? field->initial()
	//	     : !field->choices.empty()     ? field->choices[0]->color
	//	     :                               Color()
	//	     , true)
	//
	// `ColorValue::ColorValue`, `src/data/field/color.cpp`, Zeilen 84–88,
	// Fassung f9d9356 vom 2021-07-06. **Die erste Wahl ist die Vorgabe** —
	// ohne sie bekam `recolor_image` keine Farbe, und vier Ebenen je Karte
	// wurden in ihrer Ausgangsfarbe gezeichnet
	// (#270).
	Wahlfarben []string

	// Initial ist der **Anfangswert** eines Auswahlfeldes — `initial:`.
	//
	// **Er steht vor der ersten Wahl, nicht daneben.** MSE vergibt den Wert in
	// drei Stufen: `initial`, sonst die erste Wahl, sonst der leere Text
	// (`src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356 vom
	// 2021-07-06).
	//
	// 17-mal in `M15-Magic-Pack`, darunter `rarity: initial: common` — dessen
	// erste Wahl heißt `basic land`.
	Initial string

	// Skript ist das **Wertskript** eines Feldes — `script:`.
	//
	// **Es rechnet den Wert aus, `default:` füllt ihn nur.** MSE wendet beide
	// an, und in dieser Reihenfolge:
	//
	//	field().default_script.invokeOnDefault(ctx, value)
	//	| field().        script.invokeOn(ctx, value)
	//
	// `src/data/field/text.cpp`, Zeilen 178–179, Fassung f9d9356 vom
	// 2021-07-06. Der bisherige Wert steht dem Skript als `value` zur
	// Verfügung.
	//
	// **Daran hängt der halbe Kartentext.** Das Feld `text` von
	// `magic.mse-game` trägt `save value: false` und entsteht erst hier — aus
	// Regeltext und Flavortext, getrennt durch `<line>`
	// (#345).
	Skript string

	// Sortierskript ist das `sort script:` eines Feldes —
	// #1232.
	//
	//	inline String getSortKey() const {
	//	  return fieldP->sort_script ? sort_value : toString();
	//	}
	//
	// `src/data/field.hpp`, Zeilen 234–237, Fassung f9d9356 vom 2021-07-06.
	// Wo es steht, ordnet MSE die Kartenliste nach seinem **Ergebnis** und
	// nicht nach dem Wert.
	//
	// **15-mal in `magic.mse-game`**: vier Farbfelder, vier Indikatoren, drei
	// Seltenheiten und die Kartennummer. Sichtbar wird es an der Seltenheit —
	// ohne das Skript steht `common, mythic rare, rare, uncommon` da, also das
	// Alphabet statt der Ordnung des Spiels.
	Sortierskript string

	// Editierbar ist `editable:` — ob jemand dieses Feld bearbeiten darf.
	//
	//	bool      editable;    ///< Can values of this field be edited?
	//	bool      save_value;  ///< Should values of this field be written to files?
	//
	// `src/data/field.hpp`, Zeilen 45–52, Fassung f9d9356 vom 2021-07-06.
	// **Zwei Eigenschaften und nicht eine**: `save value: false` heißt „steht
	// nicht in der Datei", `editable: false` heißt „niemand tippt hier".
	//
	//	bool is_enabled(ValueViewerP const& v) {
	//	  return v->getField()->editable && v->isVisible();
	//	}
	//
	// `src/gui/control/card_editor.cpp`, Zeilen 155–156: MSE lässt ein
	// solches Feld nicht einmal auswählen.
	//
	// **Ohne Angabe wahr**; `magic.mse-game` setzt es an **24** der 249
	// Kartenfelder auf `false` — `full name`, `cmc`, `card number`, `shape`,
	// `color category`, `exact color`, `color identity`, `color count` und
	// weitere. Der Editor bot sie an, und der Schreibweg wies sie danach ab
	// (#708).
	Editierbar bool

	// Gespeichert ist `save value:` — ob der Wert dieses Feldes in die Datei
	// gehört.
	//
	//	class Field {
	//	  …
	//	  bool save_value; ///< Should the value of this field be written to the file?
	//
	// `src/data/field.hpp`, Fassung f9d9356 vom 2021-07-06. **Ohne Angabe
	// wahr**; `magic.mse-game` setzt es an genau einer Art von Feld auf
	// `false`: an denen, die aus anderen **entstehen**.
	//
	// **`text` ist der Fall, für den es das gibt.** Es ist eine Sicht auf
	// `rule text` und `flavor text`, zusammengesetzt über `combined_editor` —
	// und wer es speicherte, legte einen Wert ab, den MSE nie liest
	// (#696).
	Gespeichert bool

	// Mehrzeilig ist `multi line:` — ob das Feld mehr als eine Zeile trägt.
	//
	//	TextField::TextField()
	//	  : multi_line(false)
	//
	// `src/data/field/text.cpp`, Zeilen 16–24, Fassung f9d9356 vom
	// 2021-07-06. **Ohne Angabe einzeilig**; über `magic.mse-game` steht es
	// **68**-mal auf `true`.
	//
	// **Für die Oberfläche und für den Satz.** Das Eingabefeld wählt danach
	// sein Steuerelement
	// (#560),
	// und MSE bricht ein einzeiliges Feld beim Zeichnen nicht um
	// (`style.field().multi_line`, `src/render/text/viewer.cpp`, Zeile 618).
	// Hier stand bis zum 2026-10-11 „nicht für den Satz" — erschlossen, und
	// vom Zweitblick in J05 (#2269) widerlegt.
	Mehrzeilig bool

	// Beschreibung ist `description:` — der Satz, den MSE als Hilfetext zeigt.
	//
	// **Nicht übersetzt** (§9): er steht so im Paket. MSE hält ihn als
	// `LocalizedString` (`src/data/field.hpp`, Zeile 48, Fassung f9d9356 vom
	// 2021-07-06); die Übersetzungen stünden im Paket unter
	// `localized_description:` und stehen dort nicht.
	Beschreibung string

	// Statistik ist `show statistics:` — ob MSE für dieses Kartenfeld eine
	// eigene Dimension anlegt (`Game::validate`, `src/data/game.cpp`, Zeilen
	// 68–77). **Ohne Angabe ja**, bei einem Bildfeld nein
	// (`src/data/field.cpp` Zeile 28, `src/data/field/image.hpp` Zeile 27;
	// Fassung f9d9356 vom 2021-07-06) — #1985.
	Statistik bool

	// Statistikfarben ist `choice colors:` eines Auswahlfeldes — Wahl auf
	// Farbe, in MSEs Schreibweise. **Nicht dasselbe wie [Field.Wahlfarben]**:
	// jene sind die Farben, die ein Farbfeld zur Wahl stellt, diese färben die
	// Balken seiner Statistik (`src/data/statistics.cpp`, Zeile 40, Fassung
	// f9d9356 vom 2021-07-06). `nil`, wo keine dastehen.
	Statistikfarben map[string]string

	// Beschriftung ist `caption:` — **die Aufschrift, nicht der Bezeichner**.
	//
	// MSE beschriftet ein Eingabefeld damit (`NativeLookEditor`,
	// `src/gui/control/native_look_editor.cpp`, Zeile 42, Fassung f9d9356 vom
	// 2021-07-06); der Name ist, was Skripte ansprechen. Leer heißt: aus dem
	// Namen abgeleitet ([Field.Beschriftet]).
	Beschriftung string

	// Uebersetzungen ist `localized caption:` — je Sprachkürzel eine
	// Aufschrift.
	//
	// **Das ist die eine Stelle, an der ein Paket seine Oberflächentexte
	// selbst übersetzt.** §9 sagt: was MSE in seinen Paketen mitbringt, wird
	// von dort gelesen und nicht nachgebaut. In `M15-Magic-Pack` stehen
	// zehn davon, alle an Setfeldern und alle nur auf Chinesisch — die
	// Kartenfelder tragen keine, und deshalb zeigt **auch MSE2** dort
	// englische Wörter.
	Uebersetzungen map[string]string

	// Listenname ist `card list name:` — die Aufschrift der **Spalte** in der
	// Kartenliste.
	//
	//	if (card_list_name.default_.empty())
	//	  card_list_name.default_ = capitalize(caption.default_);
	//
	// `src/data/field.cpp`, Zeile 70, Fassung f9d9356 vom 2021-07-06. Bei MSE
	// hängt er **an der Aufschrift**, nicht umgekehrt — 243 der 249
	// Kartenfelder von `magic.mse-game` schreiben ihn trotzdem aus.
	Listenname string

	// Listenuebersetzungen ist `localized card list name:`.
	//
	// **Hier steht das einzige Deutsch, das ein Kartenfeld dieses Packs
	// mitbringt** (#699):
	// 25 Felder, jedes mit `de:` — „Name", „Manakosten", „Typ", „Seltenheit",
	// „S/W", „MV". `localized caption:` gibt es an **keinem** Kartenfeld.
	Listenuebersetzungen map[string]string

	// **Was das Spiel über die Spalte dieses Feldes sagt** — die
	// `card list …:`-Angaben stehen in [Listenangaben] (#1168).
	Listenangaben
}

// Beschriftet gibt die Aufschrift eines Feldes in der gefragten Sprache.
//
//	String const& LocalizedString::get(String const& locale) const {
//	  auto it = translations.find(locale);
//	  if (it != translations.end()) return it->second;
//	  else return default_;
//	}
//
// `src/data/locale.cpp`, Zeilen 125–132, mit dem Rückfall aus
//
//	if (caption.default_.empty()) caption.default_ = name_to_caption(name);
//
// `src/data/field.cpp`, Zeile 69. Beides Fassung f9d9356 vom 2021-07-06.
//
// # adapted: der übersetzte Listenname zählt mit
//
// Gezählt an `magic.mse-game/card_fields`: 249 Kartenfelder, **null** mit
// `caption:`, **null** mit `localized caption:` — und **25** mit `localized
// card list name:`, jedes davon mit `de:`. Ohne diesen Zweig stünde über jedem
// Feld eines deutschen Fensters ein englisches Wort, obwohl das Paket ein
// deutsches mitbringt.
//
// **Bei MSE hängt der Listenname an der Aufschrift, nicht umgekehrt**
// (`src/data/field.cpp`, Zeile 70) — sein Formularfeld hieße auch auf einem
// deutschen MSE „Type". Die Abweichung ist begründet und eng geschnitten:
// genommen wird **allein** der Eintrag in der gemeldeten Sprache. Der
// unübersetzte Listenname zählt nicht, und ohne Eintrag bleibt alles, wie MSE
// es hat — eine erfundene Übersetzung wäre etwas anderes und nach §9 verboten,
// denn der Feldname ist der Schlüssel, unter dem der Wert in der Datei steht
// (#699).
func (f Field) Beschriftet(sprache string) string {
	sprache = Paketsprache(sprache)

	if text, da := f.Uebersetzungen[sprache]; da {
		return text
	}

	if f.Beschriftung != "" {
		return f.Beschriftung
	}

	if text, da := f.Listenuebersetzungen[sprache]; da {
		return text
	}

	// **Der unübersetzte Listenname zählt nicht.** Er ist eine Spaltenüberschrift
	// und darf kürzer sein als eine Aufschrift: `cmc` heißt in der Liste `MV`,
	// bei MSE im Formular aber `Cmc`. Wer ihn nähme, wiche ohne Not ab — die
	// Begründung oben trägt genau einen Zweig, und das ist der übersetzte.
	return NameZuBeschriftung(f.Name)
}

// NameZuBeschriftung macht aus einem Feldnamen eine Aufschrift.
//
//	String name_to_caption(const String& str) {
//	  bool leading = true, first = true;
//	  for (wxUniChar c : str) {
//	    if ((c == '_' || c == ' ')) { ret += leading ? c : ' '; }
//	    else if (first) { ret += toUpper(c); leading = false; first = false; }
//	    else { ret += c; leading = false; }
//	  }
//	}
//
// `src/util/string.cpp`, Zeilen 133–150, Fassung f9d9356 vom 2021-07-06.
//
// **Nur das erste Zeichen wird groß**, und führende Trenner bleiben, wie sie
// dastehen: `_intern` wird `_Intern`, nicht ` Intern`.
func NameZuBeschriftung(name string) string {
	var gebaut strings.Builder

	fuehrend, erstes := true, true

	for _, zeichen := range name {
		switch {
		case zeichen == '_' || zeichen == ' ':
			if fuehrend {
				gebaut.WriteRune(zeichen)
			} else {
				gebaut.WriteRune(' ')
			}
		case erstes:
			gebaut.WriteString(strings.ToUpper(string(zeichen)))

			fuehrend, erstes = false, false
		default:
			gebaut.WriteRune(zeichen)

			fuehrend = false
		}
	}

	return gebaut.String()
}

// Validate prüft eine Felddefinition.
func (f Field) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return NewFieldError("name", SchluesselPflichtfeld, "Ein Feld braucht einen Namen.")
	}

	if f.Kind == KindChoice && len(f.Choices) == 0 {
		return NewFieldError("choices", SchluesselPflichtfeld,
			"Das Feld %q ist eine Auswahl und nennt keine Möglichkeiten.", f.Name)
	}

	return nil
}
