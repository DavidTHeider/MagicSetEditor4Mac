package app

// **Die Naht zum Blatt `internal/xai`** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Warum der Lauf im Paket steht
//
// **Dieselbe Begründung wie bei den beiden Nachbarn**: [App.xainaht] und
// [xaiquelle] sind die Verdrahtung, und wer sie von außen erreichte, könnte
// den Riegel umgehen.
//
// # Und der Dienst wird dabei nicht gerufen
//
// **Jeder Lauf hier geht an einen `httptest.Server`.** Ein Lauf gegen
// `api.x.ai` brauchte ein Konto mit Guthaben.

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

// eineXAIgegenstelle antwortet wie `/v1/chat/completions`.
func eineXAIgegenstelle(t *testing.T, rumpf string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "language-models") {
				_, _ = schreiben.Write([]byte(`{"models":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// dieGuteGrokantwort ist eine vollständige Antwort dieses Weges.
const dieGuteGrokantwort = `{"model":"grok-probe","choices":[{"message":
  {"content":"{\"name\": \"Blitzschlag\"}","refusal":null},
  "finish_reason":"stop"}],"usage":{"cost_in_usd_ticks":1220000}}`

// xaiAn schaltet den Dienst ein und gibt den Bestand.
func xaiAn(t *testing.T) *sql.DB {
	t.Helper()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.DienstEinstellung(domain.DienstXAI),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	return db
}

// **Die echte Naht fragt und übersetzt die Antwort.**
func TestDieXAInahtFragtUndUebersetzt(t *testing.T) {
	t.Parallel()

	server := eineXAIgegenstelle(t, dieGuteGrokantwort)

	quelle := mitSchluessel(xaiAn(t)).xainaht(server.URL)

	// **Die Fassung nennt den Weg** — und das ist hier die Auskunft: xAI hat
	// zwei, und nur einer legt nichts ab.
	fassung, fehler := quelle.Fassung(riegelkontext(t))
	if fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	if !strings.Contains(fassung, "/v1/chat/completions") {
		t.Errorf("die Fassung ist %q", fassung)
	}

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "grok-probe", "x",
		map[string]any{"type": "object"})
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !strings.Contains(antwort.Inhalt, "Blitzschlag") ||
		antwort.Modell != "grok-probe" || antwort.Abgeschnitten {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **`end_turn` ist ein Ergebnis und kein unbekannter Ausgang.**
//
// **Das ist der Fall, den ein abgeschriebenes Paket verloren hätte**: OpenAI
// kennt das Wort nicht, und wer seine Liste übernähme, machte aus einer
// fertigen Antwort eine Verweigerung.
func TestEndTurnIstFuerDieXAInahtEinErgebnis(t *testing.T) {
	t.Parallel()

	server := eineXAIgegenstelle(t, `{"model":"m","choices":[{"message":
	  {"content":"{}"},"finish_reason":"end_turn"}]}`)

	antwort, fehler := mitSchluessel(xaiAn(t)).xainaht(server.URL).
		Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if antwort.Abgeschnitten {
		t.Errorf("die Antwort gilt als abgeschnitten: %+v", antwort)
	}
}

// **Eine abgeschnittene Antwort steht als solche da.**
func TestEineAbgeschnitteneGrokantwortStehtAlsSolcheDa(t *testing.T) {
	t.Parallel()

	server := eineXAIgegenstelle(t, `{"model":"m","choices":[{"message":
	  {"content":"{\"na"},"finish_reason":"length"}]}`)

	antwort, fehler := mitSchluessel(xaiAn(t)).xainaht(server.URL).
		Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !antwort.Abgeschnitten {
		t.Errorf("die Antwort gilt als vollständig: %+v", antwort)
	}
}

// **Ablehnung und unbekannter Ausgang sind Fehler.**
//
// **Und die Ablehnung nennt ihren Satz, der stille Fall seinen Stand**: zwei
// Wege führen zu derselben Antwort, und wer nur den ersten zeigte, zeigte im
// zweiten Fall nichts.
func TestAblehnungUndUnbekannterAusgangSindXAIFehler(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rumpf   string
		enthält string
	}{
		"Ablehnung neben stop": {
			`{"model":"m","choices":[{"message":{"content":"",
			  "refusal":"tue ich nicht"},"finish_reason":"stop"}]}`,
			"tue ich nicht",
		},
		// **Eine leere Ablehnung ist keine.** Sie fällt auf den Grund
		// zurück — und der ist hier keiner, den dieses Programm kennt.
		"eine Ablehnung aus Leerzeichen": {
			`{"model":"m","choices":[{"message":{"content":"",
			  "refusal":" "},"finish_reason":"content_filter"}]}`,
			"content_filter",
		},
		"unbekannter Grund": {
			`{"model":"m","choices":[{"message":{"content":"{}"},
			  "finish_reason":"etwasneues"}]}`,
			"etwasneues",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := eineXAIgegenstelle(t, fall.rumpf)

			_, fehler := mitSchluessel(xaiAn(t)).xainaht(server.URL).
				Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x", nil)
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
func TestDerRiegelHaengtAnDerXAInaht(t *testing.T) {
	t.Parallel()

	server := eineXAIgegenstelle(t, dieGuteGrokantwort)

	t.Run("der Schalter ist aus", func(t *testing.T) {
		t.Parallel()

		quelle := mitSchluessel(bestandFuerDenRiegel(t)).xainaht(server.URL)

		if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x",
			nil); !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Fatalf("der Fehler ist %v", fehler)
		}
	})

	t.Run("es gilt nur lokal", func(t *testing.T) {
		t.Parallel()

		db := xaiAn(t)

		if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
			domain.RichtlinieNurLokal, domain.NurLokalAn); fehler != nil {
			t.Fatalf("nur lokal setzen: %v", fehler)
		}

		_, fehler := mitSchluessel(db).xainaht(server.URL).
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
func TestOhneBestandFragtDieXAInahtNicht(t *testing.T) {
	t.Parallel()

	server := eineXAIgegenstelle(t, dieGuteGrokantwort)

	quelle := NeueApp("").MitMarktbund(erfundenerBund{}).xainaht(server.URL)

	if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "m", "x",
		nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
