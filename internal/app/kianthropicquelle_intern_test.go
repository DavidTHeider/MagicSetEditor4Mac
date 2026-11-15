package app

// **Die Naht zum Blatt `internal/anthropic`** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum der Lauf im Paket steht
//
// **Dieselbe Begründung wie bei `kiquelle_intern_test.go`**: [App.anthropicnaht]
// und [anthropicquelle] sind die Verdrahtung. Wer sie von außen erreichte,
// könnte den Riegel umgehen — und geprüft wird hier, dass die Hülle die
// Aufschrift übersetzt und der Riegel wirklich daranhängt.
//
// # Und der Dienst wird dabei nicht gerufen
//
// **Jeder Lauf hier geht an einen `httptest.Server`.** Ein Lauf gegen
// `api.anthropic.com` brauchte ein Konto mit Guthaben und kostete je Ruf
// Geld; was er belegte, stünde in `docs/HAERTUNG.md` unter der fünften Ebene
// und nicht hier.

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineAnthropicgegenstelle antwortet wie die Messages-API.
func eineAnthropicgegenstelle(t *testing.T, grund string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/models") {
				_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"claude-probe","stop_reason":"` + grund + `",
				  "content":[{"type":"text","text":"{\"name\": \"Blitzschlag\"}"}],
				  "usage":{"input_tokens":11,"output_tokens":18}}`))
		}))

	t.Cleanup(server.Close)

	return server
}

// anthropicAn schaltet den Dienst ein und gibt den Bestand.
func anthropicAn(t *testing.T) *sql.DB {
	t.Helper()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.DienstEinstellung(domain.DienstAnthropic),
		domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	return db
}

// mitSchluessel gibt die Anwendung mit einem **erfundenen** Schlüssel.
//
// # Warum kein Lauf hier den echten Schlüsselbund anfasst
//
// **§3 verlangt es**: ein Test ohne eingesetzten Bund läse den Bund dieses
// Rechners — und wäre grün, weil er wirklich funktioniert hat. Der erste
// Anlauf tat genau das und fiel nur daran auf, dass die Meldung „es fehlt
// der API-Schlüssel" lautete: auf **diesem** Rechner liegt keiner.
//
// **Und ohne Schlüssel prüfte kein Lauf den Riegel.** Das Blatt weist eine
// Anfrage ohne Zugang ab, **bevor** sie eine wird — der Riegel am Transport
// käme nie an die Reihe, und »der Schalter ist aus« sähe aus wie »der
// Schlüssel fehlt«.
func mitSchluessel(db *sql.DB) *App {
	return KIAnwendungUeber(db).MitMarktbund(erfundenerBund{})
}

// erfundenerBund gibt für jeden Dienst denselben erfundenen Schlüssel.
type erfundenerBund struct{}

func (erfundenerBund) Lies(context.Context, string, string) (string, error) {
	return "nicht-echt", nil
}

// **Die echte Naht fragt und übersetzt die Antwort.**
func TestDieAnthropicnahtFragtUndUebersetzt(t *testing.T) {
	t.Parallel()

	db := anthropicAn(t)

	server := eineAnthropicgegenstelle(t, "end_turn")

	quelle := mitSchluessel(db).anthropicnaht(server.URL)

	// **Die Fassung ist eine Konstante und keine Anfrage** — anders als bei
	// Ollama gibt es hier kein Programm, dessen Fassung man lesen könnte.
	fassung, fehler := quelle.Fassung(riegelkontext(t))
	if fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	if !strings.HasPrefix(fassung, "anthropic-version ") {
		t.Errorf("die Fassung ist %q", fassung)
	}

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe", "x",
		map[string]any{"type": "object"})
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !strings.Contains(antwort.Inhalt, "Blitzschlag") ||
		antwort.Modell != "claude-probe" || antwort.Abgeschnitten {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Eine abgeschnittene Antwort steht als solche da.**
//
// **Und nicht als Fehler**: die halbe Antwort ist brauchbar, sie ist nur
// nicht vollständig — und wer das verschwiege, ließe jemanden ein halbes
// Feld übernehmen, ohne es zu wissen.
func TestEineAbgeschnitteneAntwortStehtAlsSolcheDa(t *testing.T) {
	t.Parallel()

	db := anthropicAn(t)

	server := eineAnthropicgegenstelle(t, "max_tokens")

	quelle := mitSchluessel(db).anthropicnaht(server.URL)

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe", "x", nil)
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !antwort.Abgeschnitten {
		t.Errorf("die Antwort gilt als vollständig: %+v", antwort)
	}
}

// **Eine Ablehnung und ein unbekannter Ausgang sind Fehler.**
//
// **Getrennt und nicht zusammen** — das Issue verlangt es: „unvollständig,
// Refusal und unbekannt getrennt behandeln". Wer eine Ablehnung als leere
// Antwort durchreichte, bekäme eine Schemaprüfung, die von fehlenden Feldern
// spräche statt davon, dass das Modell nicht wollte.
func TestAblehnungUndUnbekannterAusgangSindFehler(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		grund   string
		enthält string
	}{
		"Ablehnung":         {"refusal", "abgelehnt"},
		"unbekannter Grund": {"etwas_neues", "etwas_neues"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := anthropicAn(t)

			server := eineAnthropicgegenstelle(t, fall.grund)

			quelle := mitSchluessel(db).anthropicnaht(server.URL)

			_, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe", "x", nil)
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
//
// # Die Gegenprobe zur Verdrahtung
//
// **Geprüft wird an derselben Gegenstelle**: sie antwortet, und trotzdem
// kommt nichts an. Zwei Gründe halten auf, und beide werden gefahren — der
// Schalter und »nur lokal«.
func TestDerRiegelHaengtAnDerAnthropicnaht(t *testing.T) {
	t.Parallel()

	server := eineAnthropicgegenstelle(t, "end_turn")

	t.Run("der Schalter ist aus", func(t *testing.T) {
		t.Parallel()

		quelle := mitSchluessel(bestandFuerDenRiegel(t)).anthropicnaht(server.URL)

		if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe",
			"x", nil); !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Fatalf("der Fehler ist %v", fehler)
		}
	})

	t.Run("es gilt nur lokal", func(t *testing.T) {
		t.Parallel()

		db := anthropicAn(t)

		if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
			domain.RichtlinieNurLokal, domain.NurLokalAn); fehler != nil {
			t.Fatalf("nur lokal setzen: %v", fehler)
		}

		quelle := mitSchluessel(db).anthropicnaht(server.URL)

		_, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe", "x", nil)
		if !errors.Is(fehler, domain.ErrErlaubnis) {
			t.Fatalf("der Fehler ist %v", fehler)
		}

		if !strings.Contains(fehler.Error(), "nur lokal") {
			t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
		}
	})
}

// **Ohne Bestand gibt es keinen Riegel und deshalb keinen Ruf.**
func TestOhneBestandFragtDieAnthropicnahtNicht(t *testing.T) {
	t.Parallel()

	server := eineAnthropicgegenstelle(t, "end_turn")

	quelle := NeueApp("").MitMarktbund(erfundenerBund{}).anthropicnaht(server.URL)

	if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "claude-probe",
		"x", nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
