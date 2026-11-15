package app

// **Die Naht zum Blatt `internal/ollama`** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum der Lauf im Paket steht
//
// **Weil [App.ollamalaufzeit] und [ollamaquelle] die Verdrahtung sind** —
// nicht öffentlich, und das zu Recht: wer sie von außen erreichte, könnte
// den Riegel umgehen. Was hier geprüft wird, ist, dass die Hülle die
// Aufschrift richtig übersetzt und der Riegel wirklich daranhängt.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineLaufzeitsgegenstelle antwortet wie eine Ollama-Installation.
func eineLaufzeitsgegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == "/api/version" {
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"gemma4:8b","message":{"role":"assistant",
				  "content":"{\"name\": \"Blitzschlag\"}"},"done":true,
				  "done_reason":"stop"}`))
		}))

	t.Cleanup(server.Close)

	return server
}

// **Die echte Naht fragt und übersetzt die Antwort.**
func TestDieEchteNahtFragtUndUebersetzt(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	server := eineLaufzeitsgegenstelle(t)

	quelle := KIAnwendungUeber(db).ollamalaufzeit(server.URL)

	fassung, fehler := quelle.Fassung(riegelkontext(t))
	if fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	if fassung != "0.12.3" {
		t.Errorf("die Fassung ist %q", fassung)
	}

	antwort, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "gemma4:8b", "x",
		map[string]any{"type": "object"})
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if !strings.Contains(antwort.Inhalt, "Blitzschlag") ||
		antwort.Modell != "gemma4:8b" || antwort.Abgeschnitten {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Und der Riegel hängt wirklich daran.**
//
// # Die Gegenprobe zur Verdrahtung
//
// **Ohne sie sagte der Lauf darüber nur, dass die Hülle übersetzt** — nicht,
// dass ein abgeschalteter Schalter den Ruf aufhält. Geprüft wird beides an
// **derselben** Gegenstelle: sie antwortet, und trotzdem kommt nichts an.
func TestDerRiegelHaengtAnDerEchtenNaht(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	server := eineLaufzeitsgegenstelle(t)

	quelle := KIAnwendungUeber(db).ollamalaufzeit(server.URL)

	if _, fehler := quelle.Fassung(riegelkontext(t)); !errors.Is(
		fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if _, fehler := quelle.Vorschlagen(riegelkontext(t), KIKartenanweisung, "gemma4:8b", "x",
		nil); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
