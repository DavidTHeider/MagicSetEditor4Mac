package openai

// **Die Anfrage an die Responses-API** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Eine eigene Datei
//
// `reden.go` stand mit ihr an der Zeilengrenze aus §2, und der Schnitt folgt
// der Frage: dort steht, **was zurückkommt** und wie es zu deuten ist — hier,
// **was hinausgeht**.

import (
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dieMedientypen sind die Bildformate, die diese Schnittstelle liest.
//
// **Nachgelesen am 2026-09-12** — vier, und ein fünftes wird abgewiesen,
// bevor es Geld kostet.
func dieMedientypen() []string {
	return []string{"image/png", "image/jpeg", "image/webp", "image/gif"}
}

// anfrage ist der Rumpf, der hinausgeht.
type anfrage struct {
	Modell string `json:"model"`

	// Anweisung ist `instructions` — sie hat Vorrang vor `input`.
	Anweisung string `json:"instructions,omitempty"`

	Eingabe []eingabestueck `json:"input"`

	// Hoechstens begrenzt die Ausgabe.
	Hoechstens int `json:"max_output_tokens"`

	// Form ist `text` — mit dem Schema darin, wenn es eines gibt.
	Form *textform `json:"text,omitempty"`

	// Ablegen ist `store`, und es steht **ausdrücklich** da.
	//
	// **Die Vorgabe der Schnittstelle ist `true`.** Ohne diese Zeile bliebe
	// jede Kartenanfrage beim Anbieter liegen — und `omitempty` machte aus
	// `false` ein Weglassen und damit ein `true`. Genau derselbe Fall wie
	// `truncate` bei Ollama (A7, [0098]).
	Ablegen bool `json:"store"`
}

// eingabestueck ist eine Nachricht der Eingabe.
type eingabestueck struct {
	Rolle  string       `json:"role"`
	Inhalt []inhaltteil `json:"content"`
}

// inhaltteil ist ein Stück einer Nachricht — Text oder Bild.
type inhaltteil struct {
	Art  string `json:"type"`
	Text string `json:"text,omitempty"`

	// Bildadresse ist `image_url` — bei einem mitgeschickten Bild eine
	// Datenadresse und keine Netzadresse.
	//
	// **Und das ist der Punkt**: eine Netzadresse ließe den Anbieter selbst
	// etwas holen, und wohin, entschiede dann ein Kartenfeld.
	Bildadresse string `json:"image_url,omitempty"`

	// Genauigkeit ist `detail` — `auto` überlässt es dem Anbieter.
	Genauigkeit string `json:"detail,omitempty"`
}

// textform ist `text` mit seiner Formangabe.
type textform struct {
	Format formangabe `json:"format"`
}

// formangabe ist `text.format`.
type formangabe struct {
	Art string `json:"type"`

	// Name ist Pflicht bei `json_schema` — er benennt das Schema.
	Name string `json:"name,omitempty"`

	Schema map[string]any `json:"schema,omitempty"`

	// Streng ist `strict`: der Dienst hält die Antwort **im** Schema, statt
	// sie bloß daran zu messen.
	Streng bool `json:"strict,omitempty"`
}

// derSchemaname ist, wie das Schema in der Anfrage heißt.
//
// **Ein fester Name, weil es hier nur eines gibt**: die Felder der Karte.
const derSchemaname = "kartenfelder"

// alsAnfrage macht aus der Frage den Rumpf.
func (f Frage) alsAnfrage() anfrage {
	// **Die Bilder zuerst, der Auftrag danach.**
	teile := make([]inhaltteil, 0, len(f.Bilder)+1)

	for _, bild := range f.Bilder {
		teile = append(teile, inhaltteil{
			Art:         "input_image",
			Bildadresse: "data:" + bild.Medientyp + ";base64," + bild.Daten,
			Genauigkeit: "auto",
		})
	}

	teile = append(teile, inhaltteil{Art: "input_text", Text: f.Auftrag})

	gestellt := anfrage{
		Modell: f.Alias, Anweisung: f.System,
		Eingabe:    []eingabestueck{{Rolle: "user", Inhalt: teile}},
		Hoechstens: hoechstensZeichen,

		// **`false`, und die Zeile steht ausdrücklich da.**
		Ablegen: false,
	}

	if len(f.Schema) > 0 {
		gestellt.Form = &textform{Format: formangabe{
			Art: "json_schema", Name: derSchemaname,
			Schema: f.Schema, Streng: true,
		}}
	}

	return gestellt
}

// pruefe weist eine Frage ab, die niemand beantworten kann.
//
// **Vor dem Ruf und nicht danach**: eine Anfrage ohne Modell kostet Geld und
// bekommt einen 400, den diese Seite selbst kennt.
func (f Frage) pruefe() error {
	if strings.TrimSpace(f.Alias) == "" {
		return fmt.Errorf("%w: eine Frage an OpenAI nennt ihr Modell",
			domain.ErrValidation)
	}

	if strings.TrimSpace(f.Auftrag) == "" {
		return fmt.Errorf("%w: eine Frage an OpenAI hat einen Auftrag",
			domain.ErrValidation)
	}

	for _, bild := range f.Bilder {
		if !slices.Contains(dieMedientypen(), bild.Medientyp) {
			return fmt.Errorf(
				"%w: OpenAI liest %s nicht — möglich sind %s",
				domain.ErrValidation, bild.Medientyp,
				strings.Join(dieMedientypen(), ", "))
		}

		if bild.Daten == "" {
			return fmt.Errorf("%w: ein Bild ohne Daten ist keines",
				domain.ErrValidation)
		}
	}

	return nil
}
