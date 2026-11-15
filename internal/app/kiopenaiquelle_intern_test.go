package app

// **Die Naht zum Blatt `internal/openai`** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum der Lauf im Paket steht
//
// **Dieselbe Begründung wie bei `kianthropicquelle_intern_test.go`**:
// [App.openainaht] und [openaiquelle] sind die Verdrahtung, und wer sie von
// außen erreichte, könnte den Riegel umgehen.
//
// # Und der Dienst wird dabei nicht gerufen
//
// **Jeder Lauf hier geht an einen `httptest.Server`.** Ein Lauf gegen
// `api.openai.com` brauchte ein Konto mit Guthaben.

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineOpenAIgegenstelle antwortet wie die Responses-API.
func eineOpenAIgegenstelle(t *testing.T, rumpf string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/models") {
				_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// dieGuteAntwort ist eine vollständige Antwort der Responses-API.
const dieGuteAntwort = `{"status":"completed","model":"gpt-probe",
  "output":[{"type":"message",
  "content":[{"type":"output_text","text":"{\"name\": \"Blitzschlag\"}"}]}]}`

// openaiAn schaltet den Dienst ein und gibt den Bestand.
func openaiAn(t *testing.T) *sql.DB {
	t.Helper()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.DienstEinstellung(domain.DienstOpenAI),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	return db
}

// **Die echte Naht fragt und übersetzt die Antwort.**
func TestDieOpenAInahtFragtUndUebersetzt(t *testing.T) {
	t.Parallel()

	server := eineOpenAIgegenstelle(t, dieGuteAntwort)

	quelle := mitSchluessel(openaiAn(t)).openainaht(server.URL)

	// **Die Fassung ist eine Konstante und keine Anfrage** — die
	// Responses-API trägt keine Fassungsangabe, anders als Anthropic.
	fassung, fehler := quelle.Fassung(riegelkontext(t))
	if fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	if !strings.Contains(fassung, "/v1/responses") {
		t.Errorf("die Fassung ist %q", fassung)
	}

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "gpt-probe", "x",
		map[string]any{"type": "object"})
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !strings.Contains(antwort.Inhalt, "Blitzschlag") ||
		antwort.Modell != "gpt-probe" || antwort.Abgeschnitten {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Eine abgeschnittene Antwort steht als solche da.**
func TestEineAbgeschnitteneOpenAIAntwortStehtAlsSolcheDa(t *testing.T) {
	t.Parallel()

	server := eineOpenAIgegenstelle(t, `{"status":"incomplete","model":"m",
	  "incomplete_details":{"reason":"max_output_tokens"},
	  "output":[{"type":"message",
	  "content":[{"type":"output_text","text":"{\"na"}]}]}`)

	quelle := mitSchluessel(openaiAn(t)).openainaht(server.URL)

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !antwort.Abgeschnitten {
		t.Errorf("die Antwort gilt als vollständig: %+v", antwort)
	}
}

// **Ablehnung, Filter und unbekannter Ausgang sind Fehler.**
//
// **Und die Ablehnung nennt ihren Satz, der Filter seinen Stand**: zwei Wege
// führen zu derselben Antwort, und wer nur den ersten zeigte, zeigte im
// zweiten Fall nichts.
func TestAblehnungFilterUndUnbekannterAusgangSindOpenAIFehler(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rumpf   string
		enthält string
	}{
		"Ablehnung": {
			`{"status":"completed","model":"m","output":[{"type":"message",
			  "content":[{"type":"refusal","refusal":"tue ich nicht"}]}]}`,
			"tue ich nicht",
		},
		"vom Filter angehalten": {
			`{"status":"incomplete","model":"m",
			  "incomplete_details":{"reason":"content_filter"},
			  "output":[{"type":"message","content":[]}]}`,
			"content_filter",
		},
		"unbekannter Stand": {
			`{"status":"etwasneues","model":"m","output":[]}`,
			"etwasneues",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := eineOpenAIgegenstelle(t, fall.rumpf)

			quelle := mitSchluessel(openaiAn(t)).openainaht(server.URL)

			_, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.enthält) {
				t.Errorf("die Meldung nennt %q nicht: %v", fall.enthält, fehler)
			}
		})
	}
}

// **Und der Riegel hängt wirklich daran.**
func TestDerRiegelHaengtAnDerOpenAInaht(t *testing.T) {
	t.Parallel()

	server := eineOpenAIgegenstelle(t, dieGuteAntwort)

	t.Run("der Schalter ist aus", func(t *testing.T) {
		t.Parallel()

		quelle := mitSchluessel(bestandFuerDenRiegel(t)).openainaht(server.URL)

		if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x",
			nil); !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Fatalf("der Fehler ist %v", fehler)
		}
	})

	t.Run("es gilt nur lokal", func(t *testing.T) {
		t.Parallel()

		db := openaiAn(t)

		if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
			domain.RichtlinieNurLokal, domain.NurLokalAn); fehler != nil {
			t.Fatalf("nur lokal setzen: %v", fehler)
		}

		_, fehler := mitSchluessel(db).openainaht(server.URL).
			Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
		if !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Fatalf("der Fehler ist %v", fehler)
		}

		if !strings.Contains(fehler.Error(), "nur lokal") {
			t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
		}
	})
}

// **Ohne Bestand fragt sie nicht.**
func TestOhneBestandFragtDieOpenAInahtNicht(t *testing.T) {
	t.Parallel()

	server := eineOpenAIgegenstelle(t, dieGuteAntwort)

	quelle := NeueApp("").MitMarktbund(erfundenerBund{}).openainaht(server.URL)

	if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x",
		nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und sie erzeugt Bilder** — Werkstück A12 (#1451).
//
// # Warum der Lauf hier steht und nicht im Blatt
//
// **Weil er die Naht prüft und nicht den Klienten.** Ob `internal/openai`
// Base64 richtig liest, steht dort; hier steht, dass **diese** Naht die
// einzige ist, die [KIBildquelle] erfüllt — und dass sie den Wunsch
// unverändert weiterreicht.
func TestDieOpenAInahtErzeugtBilder(t *testing.T) {
	t.Parallel()

	// **Ein winziges PNG als Base64** — acht mal acht, einfarbig.
	const einBild = `{"created":1,"output_format":"png","data":[{"b64_json":` +
		`"iVBORw0KGgoAAAANSUhEUgAAAAgAAAAIAQAAAADrjVsBAAAAEUlEQVR4nGNgYGD4z` +
		`8DAwAAABAAB/1nZlgAAAABJRU5ErkJggg==","revised_prompt":"ein Blitz"}]}`

	server := eineOpenAIgegenstelle(t, einBild)

	quelle := mitSchluessel(openaiAn(t)).openainaht(server.URL)

	bildquelle, kann := quelle.(KIBildquelle)
	if !kann {
		t.Fatal("die Naht zu OpenAI erfüllt KIBildquelle nicht")
	}

	roh, fehler := bildquelle.Bild(riegelkontext(t), KIBildwunsch{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
		Groesse: "1024x1536", Guete: "medium",
	})
	if fehler != nil {
		t.Fatalf("Bild: %v", fehler)
	}

	if roh.Medientyp != "image/png" || len(roh.Inhalt) == 0 {
		t.Fatalf("das Bild ist %q mit %d Bytes", roh.Medientyp, len(roh.Inhalt))
	}

	// **Die Umschrift reist mit** — sie ist Herkunft.
	if roh.Umschrift != "ein Blitz" {
		t.Errorf("die Umschrift ist %q", roh.Umschrift)
	}

	// **Und ein Fehler kommt als Fehler zurück.**
	if _, fehler := bildquelle.Bild(riegelkontext(t), KIBildwunsch{
		Auftrag: "ohne Modell",
	}); fehler == nil {
		t.Error("der Auftrag ohne Modell kam durch")
	}
}
