package xai_test

// **Was eine Antwort von xAI heißt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Der Fund, den dieses Werkstück von den anderen trennt
//
// **`end_turn` neben `stop`.** Anthropic kennt es, OpenAI nicht, und xAI
// führt es. Wer die Ausgangsliste von OpenAI übernähme, hielte eine fertige
// Antwort für einen unbekannten Ausgang — und das sähe aus wie ein Modell,
// das sich verweigert.
//
// **Und der Preis steht dabei.** `usage.cost_in_usd_ticks` sagt, was dieser
// eine Ruf gekostet hat. Das tut sonst keiner.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Eine fertige Antwort kommt mit Inhalt, Modell und Preis zurück.**
func TestEineFertigeAntwortTraegtIhrenPreis(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(eineGegenstelle(t).URL,
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	antwort, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "grok-probe", Auftrag: "nenne eine Karte",
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Ausgang != xai.AusgangFertig {
		t.Errorf("der Ausgang ist %q (%q)", antwort.Ausgang, antwort.Grund)
	}

	if antwort.Inhalt != `{"name": "Blitzschlag"}` {
		t.Errorf("der Inhalt ist %q", antwort.Inhalt)
	}

	if antwort.Modell != "grok-probe" {
		t.Errorf("das Modell ist %q", antwort.Modell)
	}

	// **Der Preis dieses einen Rufs** — 1.220.000 Ticks sind 0,000122 $.
	if antwort.Verbrauch.Kosten != 1_220_000 ||
		antwort.Verbrauch.Zusammen != 29 {
		t.Errorf("der Verbrauch ist %+v", antwort.Verbrauch)
	}
}

// **Die vier Ausgänge, und `end_turn` ist der, den sonst keiner kennt.**
func TestDieAusgaengeEinerAntwort(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		grund     string
		ablehnung string
		erwartet  xai.Ausgang
	}{
		{"stop heißt fertig", "stop", "", xai.AusgangFertig},
		{"end_turn heißt auch fertig", "end_turn", "", xai.AusgangFertig},
		{"length heißt abgeschnitten", "length", "", xai.AusgangUnvollstaendig},
		{"null heißt unbekannt", "", "", xai.AusgangUnbekannt},
		{
			"ein fremdes Wort heißt unbekannt", "tool_calls", "",
			xai.AusgangUnbekannt,
		},

		// **Die Ablehnung steht neben `stop`** — wer den Grund zuerst läse,
		// hielte sie für ein Ergebnis.
		{
			"eine Ablehnung neben stop bleibt eine Ablehnung", "stop",
			"das tue ich nicht", xai.AusgangAbgelehnt,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			antwort, fehler := eineAntwortMit(t, fall.grund, fall.ablehnung)
			if fehler != nil {
				t.Fatalf("Rede: %v", fehler)
			}

			if antwort.Ausgang != fall.erwartet {
				t.Errorf("der Ausgang ist %q, erwartet war %q",
					antwort.Ausgang, fall.erwartet)
			}

			if fall.ablehnung != "" && antwort.Naeheres != fall.ablehnung {
				t.Errorf("das Nähere ist %q", antwort.Naeheres)
			}
		})
	}
}

// **Eine Antwort ohne Wahl ist kein leeres Ergebnis, sondern ein Fehler.**
//
// **Der Unterschied ist der ganze Fall**: ein leerer Inhalt sähe aus wie ein
// Modell, das nichts gefunden hat — und das wäre eine Auskunft über die
// Karte statt über den Dienst.
func TestEineAntwortOhneWahlIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(
		einDienstDer(t, http.StatusOK, `{"model":"grok-probe","choices":[]}`),
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), xai.Frage{Alias: "m", Auftrag: "x"})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Eine unlesbare Antwort ist ein Fehler und keine leere.**
func TestEineUnlesbareAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(einDienstDer(t, http.StatusOK, "kein JSON"),
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), xai.Frage{Alias: "m", Auftrag: "x"})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// eineAntwortMit fragt einen Dienst, der genau diesen Ausgang meldet.
func eineAntwortMit(t *testing.T, grund, ablehnung string) (xai.Antwort, error) {
	t.Helper()

	rumpf := `{"model":"grok-probe","choices":[{"message":{"content":"x",` +
		`"refusal":` + alsJSON(ablehnung) + `},"finish_reason":` +
		alsJSON(grund) + `}]}`

	klient := xai.NeuerKlient(einDienstDer(t, http.StatusOK, rumpf),
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	return klient.Rede(t.Context(), xai.Frage{Alias: "m", Auftrag: "x"})
}

// alsJSON macht aus einer leeren Zeichenkette `null` und sonst einen Text.
//
// **`null` und nicht `""`** — so steht es in der Dokumentation, und ein
// Leser, der nur `""` kennte, stolperte über den echten Fall.
func alsJSON(wert string) string {
	if wert == "" {
		return "null"
	}

	return `"` + wert + `"`
}

// einDienstDer antwortet mit genau diesem Status und Rumpf.
func einDienstDer(t *testing.T, status int, rumpf string) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(status)
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server.URL
}
