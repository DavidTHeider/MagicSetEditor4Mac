package applefm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Frage ist eine Bitte an das Gerätemodell.
type Frage struct {
	// System ist die Anweisung, Auftrag die Bitte selbst.
	System  string
	Auftrag string

	// Schema zwingt die Antwort in eine Form — `response_format: json_schema`,
	// gemessen am 2026-10-03: der Inhalt ist dann JSON nach dem Schema.
	Schema map[string]any
}

// Ausgang sagt, wie eine Antwort endete.
type Ausgang string

const (
	// AusgangFertig ist eine ganze Antwort.
	AusgangFertig Ausgang = "fertig"

	// AusgangAbgelehnt ist eine Ablehnung durch die Schutzregeln — **in
	// beiden gemessenen Arten** (Vermerk 0166).
	AusgangAbgelehnt Ausgang = "abgelehnt"

	// AusgangUnvollstaendig ist eine abgeschnittene Antwort (`length`).
	AusgangUnvollstaendig Ausgang = "unvollstaendig"
)

// Antwort ist, was das Gerätemodell gab.
type Antwort struct {
	Inhalt  string
	Modell  string
	Ausgang Ausgang

	// Naeheres sagt bei einer Ablehnung, woran sie zu erkennen war.
	Naeheres string
}

// schutzmeldung ist der Anfang der Meldung, mit der `fm serve` eine
// Ablehnung als HTTP 500 gibt — gemessen am 2026-10-03.
const schutzmeldung = "The model's safety guardrails were triggered"

// bekannteAblehnungen sind die Sätze, mit denen das Modell **im Inhalt**
// ablehnt — gemessen, nicht vollständig: ein unbekannter Satz scheitert in
// einer Frage mit Schema am Schema und wird dort abgewiesen.
var bekannteAblehnungen = []string{
	"Diese Anfrage kann nicht bearbeitet werden.",
}

// antwortDTO ist die Gestalt von `chat.completion`.
type antwortDTO struct {
	Model   string `json:"model"`
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string  `json:"content"`
			Refusal *string `json:"refusal"`
		} `json:"message"`
	} `json:"choices"`
}

// statusServerfehler ist HTTP 500 — mit ihm meldet `fm serve` die Schutzregeln
// (Vermerk 0166). **Als Zahl und nicht über `net/http`**: das Netz kennt in
// diesem Projekt nur `internal/netz` (§16).
const statusServerfehler = 500

// Rede stellt eine Frage — **immer ohne Strom**: ohne `stream: false` gäbe
// `fm serve` einen.
func (k *Klient) Rede(ctx context.Context, frage Frage) (Antwort, error) {
	nachrichten := []map[string]string{}
	if frage.System != "" {
		nachrichten = append(nachrichten, map[string]string{"role": "system", "content": frage.System})
	}

	nachrichten = append(nachrichten, map[string]string{"role": "user", "content": frage.Auftrag})

	rumpf := map[string]any{"model": Modell, "stream": false, "messages": nachrichten}
	if frage.Schema != nil {
		rumpf["response_format"] = map[string]any{
			"type":        "json_schema",
			"json_schema": map[string]any{"name": "antwort", "schema": frage.Schema},
		}
	}

	roh, fehler := k.frage(ctx, Unterhaltungsweg, rumpf)
	if fehler != nil {
		return Antwort{}, fehler
	}

	if roh.Status == statusServerfehler &&
		strings.HasPrefix(meldungAus(roh.Rumpf), schutzmeldung) {
		return Antwort{Modell: Modell, Ausgang: AusgangAbgelehnt, Naeheres: meldungAus(roh.Rumpf)}, nil
	}

	if roh.Status != netz.StatusGut {
		return Antwort{}, fehlerAus(roh, Unterhaltungsweg)
	}

	var dto antwortDTO
	if fehler := jsonLesen(roh.Rumpf, &dto); fehler != nil {
		return Antwort{}, fehler
	}

	// **Ein verschwundenes Pflichtfeld bricht laut** (§16) — ohne `choices`
	// ist es eine andere Gestalt und kein leerer Vorschlag.
	if len(dto.Choices) == 0 {
		return Antwort{}, fmt.Errorf("%w: die Antwort des Gerätemodells hat keine Wahl (choices)",
			domain.ErrValidation)
	}

	return antwortAus(dto), nil
}

// antwortAus ordnet das fremde DTO zu — an einer Stelle (§16).
func antwortAus(dto antwortDTO) Antwort {
	antwort := Antwort{Modell: dto.Model, Ausgang: AusgangFertig}
	wahl := dto.Choices[0]
	antwort.Inhalt = wahl.Message.Content

	switch {
	case wahl.Message.Refusal != nil && *wahl.Message.Refusal != "":
		antwort.Ausgang, antwort.Naeheres = AusgangAbgelehnt, *wahl.Message.Refusal
	case istBekannteAblehnung(wahl.Message.Content):
		antwort.Ausgang, antwort.Naeheres = AusgangAbgelehnt, strings.TrimSpace(wahl.Message.Content)
	case wahl.FinishReason == "length":
		antwort.Ausgang = AusgangUnvollstaendig
	}

	return antwort
}

func istBekannteAblehnung(inhalt string) bool {
	for _, satz := range bekannteAblehnungen {
		if strings.TrimSpace(inhalt) == satz {
			return true
		}
	}

	return false
}

// jsonLesen liest eine Antwort und gibt einen Fehler mit Grund.
func jsonLesen(roh []byte, ziel any) error {
	if fehler := json.Unmarshal(roh, ziel); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler, "die Antwort des Gerätemodells lesen")
	}

	return nil
}
