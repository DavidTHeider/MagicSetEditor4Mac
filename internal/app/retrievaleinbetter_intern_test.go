package app

// **Die Einbetterquelle am Fenster** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Der Schnitt folgt der Frage**: dort steht, wie die Naht verdrahtet ist;
// hier, was sie über ein Modell sagt — ein Textmodell taugt nicht, ein
// Einbetter schon, und der Unterschied ist ein **Befund** und kein Fehler.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Modell, das kein Einbetter ist, kommt als Befund und nicht als Raum.**
//
// **Der Unterschied gehört dem Aufrufer** (§10): ein Befund sagt „dieses
// Modell taugt nicht", ein Fehler sagt „es kam gar nichts". Ohne diesen Lauf
// stünde der eine Zweig da, ohne je betreten zu werden.
func TestEinTextmodellKommtAlsBefundUndNichtAlsRaum(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				// **`completion` und kein `embedding`** — genau der Fall.
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["completion"],"license":"MIT"}`))

				return
			}

			_, _ = schreiben.Write([]byte(`{"version":"0.33.3"}`)) //nolint:errcheck // dito
		}))
	t.Cleanup(gegenstelle.Close)

	quelle := einbetterquelle{
		bauen: eingerichteteAnwendung(t, gegenstelle.URL).retrievallaufzeit,
	}

	einbetter, befunde, fehler := quelle.Einbetter(imProbeprofil(t), "ein-textmodell")
	if fehler != nil {
		t.Fatalf("Einbetter: %v", fehler)
	}

	if einbetter != nil || len(befunde) == 0 {
		t.Fatalf("es kam %v und %+v", einbetter, befunde)
	}
}

// **Ein Einbettungsmodell gibt einen Einbetter — und keinen Befund.**
//
// **Die Gegenprobe zum Lauf darüber**: einer, der nur das Nein zeigt, sagt
// nichts darüber, ob das Ja je erreicht wird (§10).
func TestEinEinbettungsmodellGibtEinenEinbetter(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(einbettendeGegenstelle))
	t.Cleanup(gegenstelle.Close)

	quelle := einbetterquelle{
		bauen: eingerichteteAnwendung(t, gegenstelle.URL).retrievallaufzeit,
	}

	einbetter, befunde, fehler := quelle.Einbetter(imProbeprofil(t), "einbetter")
	if fehler != nil || len(befunde) > 0 {
		t.Fatalf("Einbetter: %v, %+v", fehler, befunde)
	}

	if einbetter == nil {
		t.Fatal("es kam kein Einbetter")
	}

	// **Und er rechnet in dem Raum, den er nennt.**
	if raum := einbetter.Raum(); raum.Modell != "einbetter" ||
		raum.Modellrevision == "" || raum.Vertrag.Dimension != 4 {
		t.Fatalf("der Raum ist %+v", raum)
	}
}

// einbettendeGegenstelle antwortet wie eine Laufzeit mit einem Einbetter.
//
// **Selbst geschrieben** (§3): dieselbe Gestalt wie Ollama sie schickt,
// eigener Inhalt. Die Vektoren liegen auf Länge eins — sonst meldete die
// Messung „nicht normalisiert", und der Raum wäre ein anderer.
func einbettendeGegenstelle(schreiben http.ResponseWriter, anfrage *http.Request) {
	switch anfrage.URL.Path {
	case ollama.Schauweg:
		_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
			`{"capabilities":["embedding"],"license":"Apache-2.0",` +
				`"model_info":{"bert.pooling_type":1,` +
				`"tokenizer.ggml.model":"bert"}}`))

	case ollama.Modellweg:
		_, _ = schreiben.Write([]byte( //nolint:errcheck // dito
			`{"models":[{"name":"einbetter:latest","digest":"abcdef123456",` +
				`"details":{"family":"bert"}}]}`))

	case ollama.Fassungsweg:
		_, _ = schreiben.Write([]byte(`{"version":"0.33.3"}`)) //nolint:errcheck // dito

	default:
		// **Je Eingabe ein Vektor, und verschiedene Texte verschiedene** —
		// die Messung fragt einmal mit zwei Texten und prüft, ob beide
		// dasselbe geben. Wer immer denselben Vektor schickte, fiele als
		// vertragsbrüchig auf, und der Lauf prüfte den falschen Zweig.
		var gestellt struct {
			Eingaben []string `json:"input"`
		}

		if fehler := json.NewDecoder(anfrage.Body).Decode(&gestellt); fehler != nil {
			schreiben.WriteHeader(http.StatusBadRequest)

			return
		}

		vektoren := make([][]float32, 0, len(gestellt.Eingaben))

		for stelle := range gestellt.Eingaben {
			// **Auf Länge eins**, sonst meldete die Messung „nicht
			// normalisiert" und der Raum wäre ein anderer.
			eins := float32(0.5)
			if stelle%2 == 1 {
				eins = -0.5
			}

			vektoren = append(vektoren,
				[]float32{0.5, eins, 0.5, -eins})
		}

		roh, _ := json.Marshal(map[string]any{"embeddings": vektoren}) //nolint:errcheck // Zahlen

		_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
	}
}

// eingerichteteAnwendung gibt eine Anwendung mit eingeschalteter Laufzeit.
func eingerichteteAnwendung(t *testing.T, adresse string) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = einProbebestand(t)
	anwendung.MitEinstellungen(service.NeuerEinstellungsverwalter(
		storage.NeueEinstellungen(anwendung.db)))

	for schluessel, wert := range map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
		domain.KIAdresse:  adresse,
	} {
		if fehler := anwendung.einstellungen.Setzen(imProbeprofil(t),
			schluessel, wert); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	return anwendung
}
