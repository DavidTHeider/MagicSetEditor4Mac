package ollama_test

// **Die installierten Modelle** — Werkstück A7 von Phase 5.3
// (#1446).
//
// Die Gestalt der Antworten ist am 2026-09-11 unter
// <https://docs.ollama.com/api/tags.md> und
// <https://docs.ollama.com/api-reference/show-model-details.md> nachgelesen;
// die Beispiele darunter sind **selbst geschrieben** (§3).

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Was installiert ist, kommt als Modell mit Herkunft und Fähigkeiten.**
func TestDieInstalliertenModelleKommenAn(t *testing.T) {
	t.Parallel()

	modelle, befunde, fehler := klientFuer(gegenstelle(t)).
		Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	// **Drei von vieren**: das vierte hat keinen Digest, und damit ist nicht
	// zu sagen, was dort wirklich liegt.
	if len(modelle) != 3 {
		t.Fatalf("%d Modelle, erwartet 3: %+v", len(modelle), modelle)
	}

	// **Zwei Befunde**: der fehlende Digest und der Einbetter, den diese
	// Anwendung nicht einsetzt.
	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet 2: %+v", len(befunde), befunde)
	}

	erstes := modelle[0]

	if erstes.Alias != "gemma4:8b" || erstes.Anbieter != ollama.Anbieter {
		t.Errorf("das erste Modell ist %+v", erstes)
	}

	// **Die Laufzeit steht dabei** — dieselbe Datei tut unter zwei Fassungen
	// nicht dasselbe, und eine Fähigkeit ist immer die einer Kombination.
	if !strings.Contains(erstes.Laufzeit, "ollama") {
		t.Errorf("die Laufzeit ist %q", erstes.Laufzeit)
	}
}

// **Der Digest ist die Revision des ausgeführten Artefakts.**
//
// # Warum das der Kern ist
//
// A6 verlangt eine Kette bis zu dem, was **wirklich geladen** wurde
// ([0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)).
// Ollama nennt sie: `digest` aus `/api/tags`. Damit ist die Herkunft hier
// **nicht unbekannt**, sondern ein Glied lang — und mehr behauptet sie auch
// nicht: woraus das Artefakt gebaut wurde, sagt Ollama nicht.
func TestDerDigestIstDieRevisionDesArtefakts(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	kette := modelle[0].Herkunft

	if kette.Unbekannt() {
		t.Fatal("die Herkunft gilt als unbekannt")
	}

	if fehler := kette.Pruefe(); fehler != nil {
		t.Fatalf("die Kette wurde abgewiesen: %v", fehler)
	}

	if len(kette.Glieder) != 1 || kette.Glieder[0].Art != domain.KIArtefakt {
		t.Fatalf("die Kette ist %+v", kette.Glieder)
	}

	glied := kette.Glieder[0]

	if glied.Revision != "c6eb396dbd5992bbe3f5cdb947e8bbc0ee413d7c17e2beaae69f5d569cf982eb" {
		t.Errorf("die Revision ist %q", glied.Revision)
	}

	// **Die Lizenz kommt aus `/api/show`** und nicht aus einer Vermutung.
	if !strings.Contains(glied.Lizenz, "Gemma Terms of Use") {
		t.Errorf("die Lizenz ist %q", glied.Lizenz)
	}
}

// **Ein Modell ohne Digest wird abgewiesen.**
//
// **Das Issue verlangt es**: „unbekannte Herkunft wird abgewiesen". Ohne
// Digest ist nicht zu sagen, welche Datei dort liegt — und dann ist auch die
// Fähigkeit, die daran hängt, an nichts gebunden.
func TestEinModellOhneDigestWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	gefunden := false

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "ohne-digest:1b") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Die Fähigkeiten kommen von der Laufzeit — und werden übersetzt.**
//
// Ollama nennt `completion`, `vision`, `embedding` und weitere; dieses
// Programm kennt sieben ([domain.KIFaehigkeit]). Was sich nicht zuordnen
// lässt, wird **nicht** erfunden.
func TestDieFaehigkeitenKommenVonDerLaufzeit(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if !modelle[0].Kann(domain.KITextgenerierung) ||
		!modelle[0].Kann(domain.KIVision) {
		t.Errorf("die Fähigkeiten sind %+v", modelle[0].Geprueft)
	}

	// **Und was es nicht meldet, kann es hier nicht.**
	if modelle[0].Kann(domain.KIEinbettung) {
		t.Error("eine nicht gemeldete Fähigkeit gilt als vorhanden")
	}

	// **Und der Einbetter kann hier nichts** — der Vertrag dazu ist aus dem,
	// was Ollama dokumentiert, nicht zu füllen. Der Grund steht als Befund
	// daneben und nicht in einem Kommentar.
	if modelle[1].Kann(domain.KIEinbettung) {
		t.Errorf("die Einbettung gilt als vorhanden: %+v", modelle[1].Geprueft)
	}

	// **Der Beleg nennt Ollama und das Datum** — nicht „geprüft".
	for _, beleg := range modelle[0].Geprueft {
		if !strings.Contains(beleg.Beleg, "Ollama") || beleg.Datum == "" {
			t.Errorf("der Beleg ist %+v", beleg)
		}
	}
}

// **Ein Cloudmodell wird als solches geführt.**
//
// **Am Suffix und nicht an der Adresse** (nachgelesen am 2026-09-11):
// Cloudmodelle heißen `…-cloud`, und Ollama lagert sie an seinen eigenen
// Dienst aus. Sie liegen auf demselben Port wie die lokalen — die Adresse
// unterscheidet sie **nicht**.
func TestEinCloudmodellWirdAlsSolchesGefuehrt(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	nachAlias := map[string]domain.KIOrt{}
	for _, modell := range modelle {
		nachAlias[modell.Alias] = modell.Ort
	}

	if nachAlias["gemma4:8b"] != domain.KIOrtLokal {
		t.Errorf("gemma4:8b rechnet %q", nachAlias["gemma4:8b"])
	}

	if nachAlias["gpt-oss:120b-cloud"] != domain.KIOrtCloud {
		t.Errorf("das Cloudmodell rechnet %q", nachAlias["gpt-oss:120b-cloud"])
	}
}

// **Jedes Modell ist gültig** — sonst käme es nicht in den Katalog.
func TestJedesGemeldeteModellIstGueltig(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, modell := range modelle {
		if befunde := modell.Validate(); len(befunde) != 0 {
			t.Errorf("%q wurde abgewiesen: %v", modell.Alias, befunde)
		}
	}
}

// gegenstelle antwortet wie eine lokale Ollama-Laufzeit.
func gegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Fassungsweg {
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			if gestellt.URL.Path == ollama.Modellweg {
				schreibeDatei(t, schreiben, "tags.json")

				return
			}

			rumpf := make([]byte, 16384)
			gelesen, _ := gestellt.Body.Read(rumpf) //nolint:errcheck // die Gegenstelle

			if gestellt.URL.Path == ollama.Schauweg {
				if strings.Contains(string(rumpf[:gelesen]), "einbetter") {
					schreibeDatei(t, schreiben, "show-einbetter.json")

					return
				}

				schreibeDatei(t, schreiben, "show-gemma.json")

				return
			}

			// **Der fünfte Weg gehört dazu**: ohne ihn führe der Spion über
			// die Positivliste rot, und das zu Recht.
			if gestellt.URL.Path == ollama.Einbettungsweg {
				schreibeVektoren(t, schreiben, string(rumpf[:gelesen]))

				return
			}

			schreibeDatei(t, schreiben, "chat.json")
		}))

	t.Cleanup(server.Close)

	return server
}

// schreibeDatei schickt eine Datei aus `testdaten`.
func schreibeDatei(t *testing.T, schreiben http.ResponseWriter, name string) {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // ein Name aus diesem Test
	if fehler != nil {
		t.Errorf("%s lesen: %v", name, fehler)

		return
	}

	_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
}

// klientFuer baut einen Klienten gegen eine Gegenstelle.
func klientFuer(server *httptest.Server) *ollama.Klient {
	return ollama.NeuerKlient(server.URL, netz.KlientUeber(server.Client()))
}

// **Ein lokales Textmodell kann auch strukturiert antworten.**
//
// # Warum das eine Fähigkeit der Laufzeit ist
//
// **Weil `format` die Ausgabe erzwingt und nicht darum bittet.**
// **Nachgelesen am 2026-09-11** unter
// <https://docs.ollama.com/capabilities/structured-outputs>: Ollama zwingt die
// Antwort in ein JSON-Schema. Das ist eine Eigenschaft der **Kombination** aus
// Laufzeit und Modell — genau das, was A6 unter einem Beleg versteht.
func TestEinLokalesTextmodellKannAuchStrukturiert(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if !modelle[0].Kann(domain.KIStrukturierteGenerierung) {
		t.Errorf("%q kann sie nicht: %+v", modelle[0].Alias, modelle[0].Geprueft)
	}
}

// **Ein Cloudmodell kann es nicht — und das steht nicht hier, sondern dort.**
//
// **Nachgelesen am 2026-09-11**, ebenda: „Ollama's Cloud currently does not
// support structured outputs." Eine Fähigkeit, die dieses Programm für ein
// Cloudmodell behauptete, führte zu einer Antwort, die die Form nicht hält —
// und die fiele erst an der Schemaprüfung auf, als Fehler des Modells.
func TestEinCloudmodellKannNichtStrukturiert(t *testing.T) {
	t.Parallel()

	modelle, _, fehler := klientFuer(gegenstelle(t)).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	for _, modell := range modelle {
		if modell.Ort != domain.KIOrtCloud {
			continue
		}

		if modell.Kann(domain.KIStrukturierteGenerierung) {
			t.Errorf("%q behauptet strukturierte Ausgabe", modell.Alias)
		}

		// **Und die Textfähigkeit bleibt** — die Gegenprobe: ein Modell, das
		// gar nichts könnte, machte diesen Lauf ebenfalls grün.
		if !modell.Kann(domain.KITextgenerierung) {
			t.Errorf("%q kann auch keinen Text mehr", modell.Alias)
		}
	}
}
