package ollama

// **Die Unterhaltung** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum die Form der Laufzeit überlassen wird
//
// **Weil Ollama sie halten kann und eine Bitte nicht.** Der Parameter `format`
// nimmt ein JSON-Schema, und die Laufzeit zwingt die Ausgabe hinein — statt
// im Auftrag darum zu bitten und nachher zu hoffen. **Nachgelesen am
// 2026-09-11** unter <https://docs.ollama.com/capabilities/structured-outputs>.
//
// # Und warum ohne Strom
//
// **Weil ein Strom aus Teilen besteht.** `/api/chat` strömt, wenn man nichts
// sagt; wer die erste Zeile für die Antwort nähme, bekäme eine Silbe. Diese
// Anwendung braucht das ganze Ergebnis auf einmal — ein halber Vorschlag ist
// kein Vorschlag.

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Frage ist, was hinausgeht.
type Frage struct {
	// Alias ist das Modell, wie es in `/api/tags` steht.
	Alias string

	// System ist die Anweisung vor dem Auftrag — leer erlaubt.
	System string

	// Auftrag ist die eigentliche Frage.
	Auftrag string

	// Bilder sind die Rohbytes, nicht Pfade.
	//
	// **Das Paket kennt keine Datei**: ein Blatt, das Pfade läse, wäre wieder
	// an diese Maschine gebunden (§2).
	Bilder [][]byte

	// Schema ist das JSON-Schema, in das die Antwort gezwungen wird.
	//
	// **Leer heißt: keine Form** — und dann geht `format` gar nicht erst
	// hinaus. Ein leeres Schema wäre etwas anderes als keines.
	Schema map[string]any
}

// Antwort ist, was zurückkommt.
type Antwort struct {
	// Inhalt ist der Text — bei gesetztem Schema das JSON.
	Inhalt string

	// Grund ist `done_reason` der Laufzeit: „stop", „length", „load".
	Grund string

	// Fertig ist `done`.
	Fertig bool

	// Modell ist, was **tatsächlich** geantwortet hat.
	//
	// **Nicht dasselbe wie der Alias der Frage**: wer ihn nicht vergleicht,
	// merkt einen stillen Tausch nicht.
	Modell string
}

// derAbbruch ist der Grund, der eine abgeschnittene Antwort kennzeichnet.
//
// **Nachgelesen am 2026-09-11**: `done_reason` ist „stop" bei einem regulären
// Ende und „length", wenn das Fenster voll war.
const derAbbruch = "length"

// Abgeschnitten sagt, ob die Antwort am Fensterende abbrach.
//
// **Warum das eine eigene Frage ist**: eine abgeschnittene Antwort sieht aus
// wie eine ganze. Das Issue führt sie unter denen, die **keine
// Kartenänderung** erzeugen dürfen — wer `done_reason` überginge, nähme eine
// halbe Antwort für eine ganze und schriebe sie in ein Feld.
func (a Antwort) Abgeschnitten() bool { return a.Grund == derAbbruch }

// nachricht ist eine Zeile der Unterhaltung, wie Ollama sie erwartet.
type nachricht struct {
	Rolle  string   `json:"role"`
	Inhalt string   `json:"content"`
	Bilder []string `json:"images,omitempty"`
}

// redeanfrage ist der Rumpf von `/api/chat`.
type redeanfrage struct {
	Modell      string         `json:"model"`
	Nachrichten []nachricht    `json:"messages"`
	Stroemt     bool           `json:"stream"`
	Format      map[string]any `json:"format,omitempty"`
	Optionen    map[string]any `json:"options,omitempty"`

	// Denken steht immer auf `false` — **ohne `omitempty`**, denn ein
	// fehlendes Feld heißt bei Ollama nicht „nein": gemessen am 2026-10-04
	// (AI-02, #1743) über die 32 Fragen des Testsatzes dachte `qwen3:0.6b`
	// dann vor jeder Antwort — Median 267 Token und 1,8 s, eine Frage nach
	// drei Minuten noch ohne Inhalt; mit `false` 14 Token und 0,16 s. Dieses
	// Paket liest `message.thinking` nicht; Denktext wäre bezahlt und
	// weggeworfen. **Gemessen an zwei Modellen**; Ollama sagt, `false` gelte,
	// wenn das Modell es erlaubt — für eines mit benannten Stufen ist die
	// Wirkung eine Annahme (Vermerk 0098, Z6 des Zweitblicks über 6.11).
	Denken bool `json:"think"`
}

// redeantwort ist, was `/api/chat` zurückgibt.
type redeantwort struct {
	Modell    string `json:"model"`
	Nachricht struct {
		Rolle  string `json:"role"`
		Inhalt string `json:"content"`
	} `json:"message"`
	Fertig bool   `json:"done"`
	Grund  string `json:"done_reason"`
}

// Rede stellt eine Frage und gibt die ganze Antwort.
func (k *Klient) Rede(ctx context.Context, frage Frage) (Antwort, error) {
	if fehler := frage.pruefe(); fehler != nil {
		return Antwort{}, fehler
	}

	var antwort redeantwort

	if fehler := k.hole(ctx, Redeweg, frage.alsAnfrage(), &antwort); fehler != nil {
		return Antwort{}, fehler
	}

	return Antwort{
		Inhalt: antwort.Nachricht.Inhalt,
		Grund:  antwort.Grund,
		Fertig: antwort.Fertig,
		Modell: antwort.Modell,
	}, nil
}

// pruefe weist eine Frage ab, die niemand beantworten kann.
//
// **Vor dem Ruf und nicht danach**: eine Anfrage ohne Modell erzeugte eine
// Last bei der Laufzeit und einen Fehler, den diese Seite selbst kennt.
func (f Frage) pruefe() error {
	if strings.TrimSpace(f.Alias) == "" {
		return fmt.Errorf("%w: eine Frage an Ollama nennt ihr Modell",
			domain.ErrValidation)
	}

	if strings.TrimSpace(f.Auftrag) == "" {
		return fmt.Errorf("%w: eine Frage an Ollama hat einen Auftrag",
			domain.ErrValidation)
	}

	return nil
}

// alsAnfrage baut den Rumpf.
func (f Frage) alsAnfrage() redeanfrage {
	anfrage := redeanfrage{
		Modell:  f.Alias,
		Stroemt: false,
		Format:  f.Schema,
	}

	if strings.TrimSpace(f.System) != "" {
		anfrage.Nachrichten = append(anfrage.Nachrichten,
			nachricht{Rolle: "system", Inhalt: f.System})
	}

	auftrag := nachricht{Rolle: "user", Inhalt: f.Auftrag}

	for _, bild := range f.Bilder {
		auftrag.Bilder = append(auftrag.Bilder,
			base64.StdEncoding.EncodeToString(bild))
	}

	anfrage.Nachrichten = append(anfrage.Nachrichten, auftrag)

	// **Bei fester Form wird die Temperatur gesenkt.** Die Dokumentation
	// empfiehlt es ausdrücklich für verlässliche Form; ohne Schema bleibt sie
	// unberührt, denn dann ist Streuung keine Gefahr, sondern der Zweck.
	if len(f.Schema) > 0 {
		anfrage.Optionen = map[string]any{"temperature": 0}
	}

	return anfrage
}
