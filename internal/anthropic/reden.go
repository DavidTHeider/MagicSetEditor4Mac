package anthropic

// **Die Unterhaltung in fester Form** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum die Form der Schnittstelle überlassen wird
//
// **Weil Anthropic sie halten kann und eine Bitte nicht.** `output_config.format`
// nimmt ein JSON-Schema, und der Dienst zwingt die Ausgabe hinein
// (**nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/build-with-claude/structured-outputs>).
//
// **Und nicht über `output_format`.** Das ist der alte Weg: „The
// `output_format` parameter has moved to `output_config.format`, and beta
// headers are no longer required." Er wird übergangsweise noch angenommen —
// wer ihn heute neu schriebe, schriebe eine Übergangsfassung.
//
// # Und warum ohne Strom
//
// **Weil ein Strom aus Teilen besteht.** Diese Anwendung braucht das ganze
// Ergebnis auf einmal; ein halber Vorschlag ist keiner. `stream` bleibt
// deshalb ganz weg — es ist voreingestellt aus, und ein ausgeschriebenes
// `false` behauptete eine Entscheidung, wo die Vorgabe reicht.

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// hoechstensZeichen ist die Obergrenze einer Antwort.
//
// **`max_tokens` ist Pflicht** — die Messages-API verlangt es bei jedem Ruf.
// Viertausend reichen für jedes Kartenfeld und lassen eine Antwort, die aus
// dem Ruder läuft, an der Grenze enden statt am Geldbeutel.
//
// **Und es kostet nichts, sie hoch zu setzen**: „The `max_tokens` parameter
// does not factor into OTPM rate limit calculations" (nachgelesen am
// 2026-09-11). Was sie begrenzt, ist der Schaden, nicht das Kontingent.
const hoechstensZeichen = 4096

// dieMedientypen sind die vier, die Anthropic für Bilder nimmt.
//
// **Nachgelesen am 2026-09-11**: JPEG, PNG, GIF, WebP — und sonst keiner.
func dieMedientypen() []string {
	return []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
}

// Bild ist ein Bild, wie es mit einer Frage hinausgeht.
//
// **Rohbytes und kein Pfad**: ein Blatt, das Dateien läse, wäre wieder an
// diese Maschine gebunden (§2).
type Bild struct {
	Medientyp string
	Daten     []byte
}

// Frage ist, was hinausgeht.
type Frage struct {
	Alias   string
	System  string
	Auftrag string
	Bilder  []Bild

	// Schema ist das JSON-Schema, in das die Antwort gezwungen wird — leer
	// heißt: keine Form.
	//
	// **Die Teilmenge ist begrenzt** (nachgelesen am 2026-09-11): keine
	// Längen- oder Zahlgrenzen, `additionalProperties` nur `false`,
	// `minItems` nur 0 oder 1. Was dieses Programm baut
	// (`app.alsJSONSchema`), bleibt darunter — Grundtypen, `enum` und
	// Listen.
	Schema map[string]any
}

// Ausgang ist, wie eine Antwort geendet hat.
//
// # Vier Klassen und nicht sieben
//
// Die Schnittstelle nennt sieben `stop_reason`-Werte; für diese Anwendung
// zerfallen sie in vier Fragen, und das Issue verlangt drei davon
// ausdrücklich getrennt: „unvollständig, Refusal und unbekannt getrennt
// behandeln."
type Ausgang string

// Die vier Klassen.
const (
	// AusgangFertig ist `end_turn` und `stop_sequence`.
	AusgangFertig Ausgang = "fertig"

	// AusgangUnvollstaendig ist `max_tokens` und
	// `model_context_window_exceeded` — die Antwort **bricht ab** und sieht
	// aus wie eine ganze.
	AusgangUnvollstaendig Ausgang = "unvollstaendig"

	// AusgangAbgelehnt ist `refusal` — das Modell hat nicht geantwortet, und
	// das ist **kein Fehler dieses Programms**.
	AusgangAbgelehnt Ausgang = "abgelehnt"

	// AusgangUnbekannt ist alles andere — `tool_use`, `pause_turn` und jeder
	// Wert, den es heute nicht gibt.
	//
	// **Er ist die Zusage gegen Drift** (§16): ein neuer Wert wird **nicht**
	// als „fertig" gelesen, nur weil er nicht in der Liste steht.
	AusgangUnbekannt Ausgang = "unbekannt"
)

// Antwort ist, was zurückkommt.
type Antwort struct {
	Inhalt  string
	Ausgang Ausgang

	// Grund ist der rohe `stop_reason` — für ein Protokoll, das mehr wissen
	// will als die Klasse.
	Grund string

	// Naeheres ist `stop_details` bei einer Ablehnung — leer sonst.
	Naeheres string

	// Modell ist, was **tatsächlich** geantwortet hat.
	Modell string

	Verbrauch Verbrauch
}

// nachrichtenanfrage ist der Rumpf von `/v1/messages`.
type nachrichtenanfrage struct {
	Modell      string         `json:"model"`
	Zeichen     int            `json:"max_tokens"`
	System      string         `json:"system,omitempty"`
	Nachrichten []nachricht    `json:"messages"`
	Aufbau      *ausgabeaufbau `json:"output_config,omitempty"`
}

// ausgabeaufbau ist `output_config`.
type ausgabeaufbau struct {
	Form ausgabeform `json:"format"`
}

// ausgabeform ist `output_config.format`.
type ausgabeform struct {
	Art    string         `json:"type"`
	Schema map[string]any `json:"schema"`
}

// nachricht ist eine Zeile der Unterhaltung.
type nachricht struct {
	Rolle  string  `json:"role"`
	Inhalt []block `json:"content"`
}

// block ist ein Inhaltsblock — Text oder Bild.
type block struct {
	Art    string      `json:"type"`
	Text   string      `json:"text,omitempty"`
	Quelle *bildquelle `json:"source,omitempty"`
}

// bildquelle ist `source` eines Bildblocks.
type bildquelle struct {
	Art       string `json:"type"`
	Medientyp string `json:"media_type"`
	Daten     string `json:"data"`
}

// nachrichtenantwort ist, was `/v1/messages` zurückgibt.
type nachrichtenantwort struct {
	Modell string `json:"model"`
	Inhalt []struct {
		Art  string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Grund    string `json:"stop_reason"`
	Naeheres struct {
		Art string `json:"type"`
	} `json:"stop_details"`
	Verbrauch Verbrauch `json:"usage"`
}

// Rede stellt eine Frage und gibt die ganze Antwort.
func (k *Klient) Rede(ctx context.Context, frage Frage) (Antwort, error) {
	anfrage, fehler := frage.alsAnfrage()
	if fehler != nil {
		return Antwort{}, fehler
	}

	var antwort nachrichtenantwort

	if fehler := k.hole(ctx, Nachrichtenweg, nil, anfrage, &antwort); fehler != nil {
		return Antwort{}, fehler
	}

	return Antwort{
		Inhalt:    textAus(antwort),
		Ausgang:   ausgangFuer(antwort.Grund),
		Grund:     antwort.Grund,
		Naeheres:  antwort.Naeheres.Art,
		Modell:    antwort.Modell,
		Verbrauch: antwort.Verbrauch,
	}, nil
}

// textAus fügt die Textblöcke der Antwort zusammen.
//
// **Zusammengefügt und nicht der erste genommen**: die Antwort ist eine
// **Liste** von Blöcken, und bei nachdenkenden Modellen steht der Text nicht
// zwingend allein. Ein `thinking`-Block ist kein Text und wird übergangen.
func textAus(antwort nachrichtenantwort) string {
	var teile []string

	for _, block := range antwort.Inhalt {
		if block.Art == "text" {
			teile = append(teile, block.Text)
		}
	}

	return strings.Join(teile, "")
}

// ausgangFuer ordnet einen `stop_reason` seiner Klasse zu.
//
// **Nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/api/handling-stop-reasons>: sieben
// Werte, und hier vier Klassen.
func ausgangFuer(grund string) Ausgang {
	switch grund {
	case "end_turn", "stop_sequence":
		return AusgangFertig

	case "max_tokens", "model_context_window_exceeded":
		return AusgangUnvollstaendig

	case "refusal":
		return AusgangAbgelehnt
	}

	return AusgangUnbekannt
}

// pruefe weist eine Frage ab, die niemand beantworten kann.
func (f Frage) pruefe() error {
	if strings.TrimSpace(f.Alias) == "" {
		return fmt.Errorf("%w: eine Frage an Anthropic nennt ihr Modell",
			domain.ErrValidation)
	}

	if strings.TrimSpace(f.Auftrag) == "" {
		return fmt.Errorf("%w: eine Frage an Anthropic hat einen Auftrag",
			domain.ErrValidation)
	}

	for _, bild := range f.Bilder {
		if !slices.Contains(dieMedientypen(), bild.Medientyp) {
			return fmt.Errorf(
				"%w: %q ist kein Medientyp, den Anthropic nimmt — es sind %v",
				domain.ErrValidation, bild.Medientyp, dieMedientypen())
		}
	}

	return nil
}

// alsAnfrage baut den Rumpf.
func (f Frage) alsAnfrage() (nachrichtenanfrage, error) {
	if fehler := f.pruefe(); fehler != nil {
		return nachrichtenanfrage{}, fehler
	}

	anfrage := nachrichtenanfrage{
		Modell:      f.Alias,
		Zeichen:     hoechstensZeichen,
		System:      f.System,
		Nachrichten: []nachricht{{Rolle: "user", Inhalt: f.bloecke()}},
	}

	if len(f.Schema) > 0 {
		anfrage.Aufbau = &ausgabeaufbau{Form: ausgabeform{
			Art: "json_schema", Schema: f.Schema,
		}}
	}

	return anfrage, nil
}

// bloecke baut die Inhaltsblöcke der Frage.
//
// **Bilder vor dem Text.** Die Dokumentation empfiehlt es, und die
// Reihenfolge ist eine Aussage über die Frage: erst das, worüber geredet
// wird, dann was daran zu tun ist.
func (f Frage) bloecke() []block {
	bloecke := make([]block, 0, len(f.Bilder)+1)

	for _, bild := range f.Bilder {
		bloecke = append(bloecke, block{Art: "image", Quelle: &bildquelle{
			Art:       "base64",
			Medientyp: bild.Medientyp,
			Daten:     base64.StdEncoding.EncodeToString(bild.Daten),
		}})
	}

	return append(bloecke, block{Art: "text", Text: f.Auftrag})
}
