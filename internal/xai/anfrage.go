package xai

// **Die Anfrage an `/v1/chat/completions`** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Und hier ist der Grund, warum dieses Paket nicht abschreibt
//
// **Derselbe Anbieter hat zwei Bildformen**, nachgelesen am 2026-09-12:
//
//	/v1/responses         {"type":"input_image","image_url":"data:…"}
//	/v1/chat/completions  {"type":"image_url","image_url":{"url":"data:…"}}
//
// Die erste ist die von OpenAI, die zweite die alte. Wer die Form aus
// `internal/openai` übernähme, schickte an **diesen** Weg Unsinn — und der
// Fehler käme erst vom Dienst zurück, also nach dem Ruf.
//
// **Und die Systemanweisung ist hier eine Nachricht**, keine eigene
// Eigenschaft: `instructions` gibt es an diesem Weg nicht.

import (
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dieMedientypen sind die Bildformate, die xAI liest.
//
// **Zwei und nicht vier.** Nachgelesen am 2026-09-12: „jpg/jpeg or png".
// OpenAI nimmt zusätzlich webp und GIF, Anthropic ebenso — hier wäre beides
// ein 400 nach dem Ruf.
func dieMedientypen() []string { return []string{"image/jpeg", "image/png"} }

// anfrage ist der Rumpf, der hinausgeht.
type anfrage struct {
	Modell      string       `json:"model"`
	Nachrichten []nachricht  `json:"messages"`
	Hoechstens  int          `json:"max_tokens"`
	Form        *antwortform `json:"response_format,omitempty"`
}

// nachricht ist eine Nachricht der Unterhaltung.
//
// **Der Inhalt ist `any`, und das ist keine Bequemlichkeit**: die
// Systemnachricht trägt eine Zeichenkette, die Nutzernachricht eine Liste
// von Teilen. Beides heißt `content`.
type nachricht struct {
	Rolle  string `json:"role"`
	Inhalt any    `json:"content"`
}

// inhaltteil ist ein Stück einer Nachricht — Text oder Bild.
type inhaltteil struct {
	Art  string `json:"type"`
	Text string `json:"text,omitempty"`

	// Bild ist `image_url` — an diesem Weg ein **Objekt**.
	Bild *bildadresse `json:"image_url,omitempty"`
}

// bildadresse ist das Objekt hinter `image_url`.
type bildadresse struct {
	// Adresse ist eine Datenadresse und keine Netzadresse.
	//
	// **Und das ist der Punkt**: eine Netzadresse ließe den Anbieter selbst
	// etwas holen, und wohin, entschiede dann ein Kartenfeld.
	Adresse string `json:"url"`

	// Genauigkeit ist `detail`.
	//
	// **`high` und nicht `auto`**: xAI dokumentiert für diesen Weg die Werte
	// `high`, `low` und `auto` — und eine Karte, die gelesen werden soll,
	// wird gelesen.
	Genauigkeit string `json:"detail,omitempty"`
}

// antwortform ist `response_format`.
type antwortform struct {
	Art string `json:"type"`

	// Schema ist `json_schema` — mit Name, Schema und `strict` darin.
	//
	// **Eine Ebene tiefer als bei OpenAI**: dort steht `text.format` mit den
	// Feldern unmittelbar darin, hier ein eigenes Objekt.
	Schema *schemaangabe `json:"json_schema,omitempty"`
}

// schemaangabe ist der Inhalt von `response_format.json_schema`.
type schemaangabe struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`

	// Streng ist `strict`: der Dienst hält die Antwort **im** Schema, statt
	// sie bloß daran zu messen.
	Streng bool `json:"strict"`
}

// derSchemaname ist, wie das Schema in der Anfrage heißt.
//
// **Ein fester Name, weil es hier nur eines gibt**: die Felder der Karte.
const derSchemaname = "kartenfelder"

// alsAnfrage macht aus der Frage den Rumpf.
func (f Frage) alsAnfrage() anfrage {
	nachrichten := make([]nachricht, 0, 2)

	if strings.TrimSpace(f.System) != "" {
		nachrichten = append(nachrichten,
			nachricht{Rolle: "system", Inhalt: f.System})
	}

	// **Die Bilder zuerst, der Auftrag danach.** Nachgelesen am 2026-09-12:
	// „place the image(s) before the text".
	teile := make([]inhaltteil, 0, len(f.Bilder)+1)

	for _, bild := range f.Bilder {
		teile = append(teile, inhaltteil{Art: "image_url", Bild: &bildadresse{
			Adresse:     "data:" + bild.Medientyp + ";base64," + bild.Daten,
			Genauigkeit: "high",
		}})
	}

	teile = append(teile, inhaltteil{Art: "text", Text: f.Auftrag})
	nachrichten = append(nachrichten,
		nachricht{Rolle: "user", Inhalt: teile})

	gestellt := anfrage{
		Modell: f.Alias, Nachrichten: nachrichten,
		Hoechstens: hoechstensZeichen,
	}

	if len(f.Schema) > 0 {
		gestellt.Form = &antwortform{
			Art: "json_schema",
			Schema: &schemaangabe{
				Name: derSchemaname, Schema: f.Schema, Streng: true,
			},
		}
	}

	return gestellt
}

// pruefe weist eine Frage ab, die niemand beantworten kann.
//
// **Vor dem Ruf und nicht danach**: eine Anfrage ohne Modell kostet Geld und
// bekommt einen 400, den diese Seite selbst kennt.
func (f Frage) pruefe() error {
	if strings.TrimSpace(f.Alias) == "" {
		return fmt.Errorf("%w: eine Frage an xAI nennt ihr Modell",
			domain.ErrValidation)
	}

	if strings.TrimSpace(f.Auftrag) == "" {
		return fmt.Errorf("%w: eine Frage an xAI hat einen Auftrag",
			domain.ErrValidation)
	}

	// **Das Schema, bevor es Geld kostet** ([pruefeSchema]).
	if fehler := pruefeSchema(f.Schema); fehler != nil {
		return fehler
	}

	for _, bild := range f.Bilder {
		if !slices.Contains(dieMedientypen(), bild.Medientyp) {
			return fmt.Errorf("%w: xAI liest %s nicht — möglich sind %s",
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
