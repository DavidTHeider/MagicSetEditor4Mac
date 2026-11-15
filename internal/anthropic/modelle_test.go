package anthropic_test

// **Die Modelle und was sie belegen** — Werkstück A8 von Phase 5.3
// (#1447).
//
// Die Gestalt der Antwort ist am 2026-09-11 unter
// <https://platform.claude.com/docs/en/api/models/list> nachgelesen; die
// Beispiele darunter sind **selbst geschrieben** (§3).

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Fähigkeit kommt aus der Schnittstelle und nicht aus einer Liste.**
//
// # Warum das der Kern ist
//
// **`GET /v1/models` nennt sie selbst** — `structured_outputs.supported` und
// `image_input.supported` stehen je Modell in der Antwort. A6 verlangt für
// eine Fähigkeit einen **Beleg**; eine abgeschriebene Liste aus einer
// Dokumentationsseite wäre keiner, sondern eine Behauptung mit Verfallsdatum
// ([0099](../../docs/entscheidungen/0099-was-anthropic-belegt.md)).
func TestDieFaehigkeitKommtAusDerSchnittstelle(t *testing.T) {
	t.Parallel()

	modelle, befunde, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	// **Zwei von dreien**: das dritte nennt gar keine Fähigkeiten, und daraus
	// folgt nichts.
	if len(modelle) != 2 {
		t.Fatalf("%d Modelle, erwartet 2: %+v", len(modelle), modelle)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}

	erstes := modelle[0]

	if erstes.Alias != "claude-opus-5" || erstes.Anbieter != anthropic.Anbieter {
		t.Errorf("das erste Modell ist %+v", erstes)
	}

	for _, kann := range []domain.KIFaehigkeit{
		domain.KITextgenerierung, domain.KIStrukturierteGenerierung,
		domain.KIVision,
	} {
		if !erstes.Kann(kann) {
			t.Errorf("%q kann %q nicht", erstes.Alias, string(kann))
		}
	}
}

// **Ein Modell ohne die Fähigkeit bekommt sie nicht.**
//
// **Die Gegenprobe**: ohne sie sagte der Lauf darüber nur, dass irgendetwas
// eingetragen wird — nicht, dass die Auskunft gelesen wird.
func TestEinModellOhneFaehigkeitBekommtSieNicht(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	zweites := modelle[1]

	if zweites.Alias != "claude-nurtext-1" {
		t.Fatalf("das zweite Modell ist %q", zweites.Alias)
	}

	// **Text kann jedes** — das ist die Messages-API selbst.
	if !zweites.Kann(domain.KITextgenerierung) {
		t.Error("das Textmodell kann keinen Text")
	}

	for _, kannNicht := range []domain.KIFaehigkeit{
		domain.KIStrukturierteGenerierung, domain.KIVision,
	} {
		if zweites.Kann(kannNicht) {
			t.Errorf("%q behauptet %q", zweites.Alias, string(kannNicht))
		}
	}
}

// **Ein Modell ohne Auskunft kommt als Befund und nicht als Modell.**
//
// **Weil „null" nicht „kann nichts" heißt, sondern „es steht nichts fest".**
// A6: „Unbekannt bleibt unbekannt." Ein Modell mit null Fähigkeiten sähe im
// Katalog aus wie eines, das geprüft wurde und nichts kann.
func TestEinModellOhneAuskunftKommtAlsBefund(t *testing.T) {
	t.Parallel()

	modelle, befunde, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, modell := range modelle {
		if modell.Alias == "claude-ohneauskunft-1" {
			t.Fatal("das Modell ohne Auskunft steht im Katalog")
		}
	}

	if len(befunde) != 1 || befunde[0].Datei != "claude-ohneauskunft-1" {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Der Ort ist die Cloud, und das steht fest.**
//
// **Anders als bei Ollama gibt es hier keine Einstufung nach Adresse**: die
// Messages-API liegt bei Anthropic, und ein Modell dort rechnet dort. Unter
// »nur lokal« wird es deshalb gar nicht erst gefragt.
func TestDerOrtIstDieCloud(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, modell := range modelle {
		if modell.Ort != domain.KIOrtCloud {
			t.Errorf("%q rechnet %q", modell.Alias, modell.Ort)
		}
	}
}

// **Die Herkunft ist unbekannt — und das steht da.**
//
// # Warum hier keine Kette entsteht
//
// **Weil es keine gibt.** Anthropic nennt weder ein Artefakt noch eine
// Revision noch eine Datei mit Prüfsumme; ein Modellname ist kein Digest. A6
// sagt es scharf: „**kein erratener HF-Commit aus einem Ollama-Alias**" — und
// dasselbe gilt hier, nur noch deutlicher: es gibt gar nichts zu erraten.
func TestDieHerkunftIstUnbekanntUndSagtEs(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if !modelle[0].Herkunft.Unbekannt() {
		t.Errorf("die Herkunft ist %+v", modelle[0].Herkunft)
	}
}

// **Ohne Schlüssel wird nicht gefragt.**
func TestOhneSchluesselWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{})

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Fatal("ohne Schlüssel kamen Modelle")
	}

	if gefragt {
		t.Error("der Dienst wurde ohne Schlüssel gefragt")
	}
}

// **Jede Anfrage nennt Schlüssel und Fassung.**
//
// **Nachgelesen am 2026-09-11**: `x-api-key` und `anthropic-version` sind
// beide Pflicht. Ohne die Fassung antwortet der Dienst mit einem 400, und die
// Meldung spräche von der Anfrage statt von der fehlenden Kopfzeile.
func TestJedeAnfrageNenntSchluesselUndFassung(t *testing.T) {
	t.Parallel()

	gesehen := make(chan http.Header, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			gesehen <- gestellt.Header.Clone()
			schreibeDatei(t, schreiben, "modelle.json")
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	kopf := <-gesehen

	if kopf.Get("x-api-key") != "sk-probe" {
		t.Errorf("der Schlüssel ist %q", kopf.Get("x-api-key"))
	}

	if kopf.Get("anthropic-version") != anthropic.Fassung {
		t.Errorf("die Fassung ist %q", kopf.Get("anthropic-version"))
	}

	// **Und keine `Authorization`** — Anthropic nimmt `x-api-key`, und eine
	// zweite Kopfzeile daneben wäre ein zweiter Zugang, der niemandem gehört.
	if kopf.Get("Authorization") != "" {
		t.Errorf("es ging eine Authorization hinaus: %q", kopf.Get("Authorization"))
	}
}

// gegenstelle antwortet wie die Schnittstelle von Anthropic.
func gegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == anthropic.Modellweg {
				schreibeDatei(t, schreiben, "modelle.json")

				return
			}

			schreibeDatei(t, schreiben, "antwort.json")
		}))

	t.Cleanup(server.Close)

	return server
}

// schreibeDatei schickt eine Datei aus `testdaten`.
func schreibeDatei(t *testing.T, schreiben http.ResponseWriter, name string) {
	t.Helper()

	// **Der Name kommt aus dem Lauf, nicht von außen** — ein fester
	// Ordner im Repositorium, und jeder Aufrufer steht daneben.
	roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // siehe oben
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
}

// klientFuer baut einen Klienten gegen eine Gegenstelle.
func klientFuer(server *httptest.Server) *anthropic.Klient {
	return anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "sk-probe"})
}
