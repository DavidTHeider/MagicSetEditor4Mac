package domain

// Die Stiloptionen eines Stilpakets — `styling field:`.
//
// **Eine eigene Datei**, weil `karte.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, was eine **Karte** ist, hier, was
// ein **Set** an seinem Stil einstellen kann.

import "strings"

// StyleField ist eine Stiloption aus einem Stilpaket — `styling field:`.
//
// # Warum englisch
//
// Der Begriff ist ein **fremder**: jede Vorlage schreibt `styling field:`, und
// ein deutscher Typname daneben verschleierte, dass es dasselbe ist
// ([0015](../../docs/entscheidungen/0015-bezeichner-folgen-dem-vokabular.md)
// nennt `StyleField` als Beispiel).
//
// # Warum nicht [Field]
//
// Beide beschreiben ein Feld, und trotzdem sind es zwei Gegenstände: ein
// [Field] gehört zu einer **Karte** und steht im Spielpaket; ein StyleField
// gehört zum **Set** und steht im Stilpaket. Der Vorgabewert heißt dort
// `initial:` statt `default:`, und die Beschreibung — einen Satz, den das Paket
// mitbringt — führt [Field] gar nicht. Wer beides zusammenlegte, hätte einen
// Typ mit zwei Herkünften und zwei Bedeutungen je Feld.
type StyleField struct {
	Name string
	Kind FieldKind

	// Description ist der Satz aus dem Paket. **Nicht übersetzt** (§9): er
	// steht so in der Vorlage, und eine Übersetzung machte jedes vorhandene
	// Paket unlesbar.
	Description string

	// Initial ist der Wert, den ein Set ohne eigene Angabe hat.
	Initial string

	// Default ist der **Ausdruck** aus `default:` — roh, wie er dasteht.
	//
	// **Ein Skript und keine Zeichenkette.** Bei MSE ist es ein
	// `OptionalScript` (`src/data/field/color.hpp`, Zeile 33, Fassung f9d9356
	// vom 2021-07-06); `"black"` ist deshalb eine Zeichenkette, `rgb(0,0,0)`
	// ein Funktionsaufruf und das nackte `black` ein Fehler. Ausgewertet wird
	// in `internal/service` — hier steht nur, was in der Datei stand.
	Default string

	Choices []string

	// Wahlfarben sind die `color:`-Angaben der `choice:`-Blöcke — dieselbe
	// Sache wie [Field.Wahlfarben], hier für ein `styling field:` mit
	// `type: color`. Der Stil des Packs führt sieben davon, darunter
	// `text_color` und `textbox_color`.
	Wahlfarben []string

	// Match ist das Muster einer [KindPackageChoice] — `match:`.
	//
	// Es nennt Dateinamen samt Endung, mit `*` als Platzhalter:
	// `magic-overlay-*.mse-include`. Welche Pakete darauf passen, entscheidet
	// der Katalog und nicht diese Entität.
	Match string

	// Required sagt, ob eine Wahl **verlangt** ist — `required:`.
	//
	// **Ohne Angabe gilt `true`.** So steht es im Kopf von
	// `PackageChoiceField` (`src/data/field/package_choice.hpp`, Zeile 27,
	// Fassung f9d9356 vom 2021-07-06):
	//
	//	PackageChoiceField() : required(true), empty_name(_("none")) {}
	//
	// Über das Pack sind 331 der 627 `package choice` ohne Angabe und damit
	// Pflicht, 296 tragen `required: false`.
	Required bool

	// EigeneFarbe sagt, ob eine Farbe außerhalb der Liste erlaubt ist —
	// `allow custom:` einer [KindColor]
	// (#634).
	//
	// **Ohne Angabe gilt `true`**:
	//
	//	ColorField::ColorField()
	//	  : allow_custom(true)
	//	  , default_name(_("Default"))
	//	{}
	//
	// `src/data/field/color.cpp`, Zeilen 15–18, Fassung f9d9356 vom
	// 2021-07-06. Über das ganze Pack ist sie **nirgends** abgeschaltet: alle
	// 105 Farbstiloptionen lassen die eigene Farbe zu, und 16 von ihnen
	// nennen überhaupt keine Wahl — für die ist der Farbwähler das Einzige,
	// was es gibt (#630).
	EigeneFarbe bool

	// Leerwahl ist der Wert, der geschrieben wird, wenn an einer
	// [KindMultipleChoice] nichts gewählt ist — `empty choice:`
	// (#637).
	//
	// **Ohne Angabe die leere Zeichenkette**, wie das Feld im Kopf von
	// `MultipleChoiceField` (`src/data/field/multiple_choice.hpp`, Zeile 27,
	// Fassung f9d9356 vom 2021-07-06). Über das Pack erklärt keine der 33
	// Mehrfachwahlen eine; gelesen wird sie trotzdem, denn eine Angabe, die
	// man nicht liest, kann man nicht einhalten.
	Leerwahl string

	// EmptyName ist die Beschriftung des leeren Eintrags — `empty name:`.
	//
	// **Ohne Angabe `none`**, aus derselben Zeile. Sie zählt nur, wenn
	// [StyleField.Required] falsch ist: sonst gibt es den Eintrag nicht.
	EmptyName string
}

// Wahlmoeglichkeit ist eine Möglichkeit einer Stiloption — was gespeichert
// wird und was dasteht.
//
// # Warum beides und nicht nur eines
//
// **Bei einer [KindChoice] sind sie dasselbe**: die Wahl heißt, wie sie heißt.
// Bei einer [KindPackageChoice] nicht — gespeichert wird
// `magic-overlay-foil.mse-include`, dasteht „Foil"
// (#632).
// Bei einer [KindColor] auch nicht: gespeichert wird `rgb(0,0,0)`, dasteht
// „schwarz" (#634).
//
// **Ein Typ für alle drei**, weil es eine Frage ist. Zwei Listen mit derselben
// Gestalt liefen auseinander, sobald eine von beiden einen Fall dazubekäme.
type Wahlmoeglichkeit struct {
	Wert         string `json:"wert"`
	Beschriftung string `json:"beschriftung"`
}

// Validate prüft eine Stiloption.
func (s StyleField) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return NewFieldError("name", SchluesselPflichtfeld, "Eine Stiloption braucht einen Namen.")
	}

	if s.Kind == KindChoice && len(s.Choices) == 0 {
		return NewFieldError("choices", SchluesselPflichtfeld,
			"Die Stiloption %q ist eine Auswahl und nennt keine Möglichkeiten.", s.Name)
	}

	return nil
}
