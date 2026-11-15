// **Eine eigene Übersetzung — an jeder Fläche** — LOC-02
// (#1688).
//
// # Der Ablauf
//
// Anlegen über Englisch → einen Satz und die Lizenz setzen → die Liste
// lesen → als Datei ausgeben → löschen → wieder einlesen → zeigen.
//
// **Dasselbe Ergebnis und dasselbe Nein**: danach ist genau ein Satz
// übersetzt, und eine eingebaute Sprache als Kennung ist an jeder Fläche ein
// fachliches Nein (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sprachsatz ist der Pfad, der übersetzt wird — ein Satz des deutschen
// Katalogs ohne Platzhalter.
const sprachsatz = "allgemein.arbeitsplatz.keinSet"

func TestEineEigeneUebersetzungGehtUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := eigenesZuhause(t)

		mussKommandozeile(t, umgebung, "sprachpaket", "anlegen", "pt-BR", "Português", "en")
		mussKommandozeile(t, umgebung, "sprachpaket", "setze", "pt-BR", "--lizenz=CC0-1.0",
			sprachsatz+"=Nenhum set aberto")

		if liste := mussKommandozeile(t, umgebung, "sprachpaket", "liste", "--json"); !strings.Contains(liste, `"kennung": "pt-BR"`) {
			t.Errorf("die Liste: %s", liste)
		}

		pfad := filepath.Join(t.TempDir(), "pt.json")
		if err := os.WriteFile(pfad, []byte(mussKommandozeile(t, umgebung, "sprachpaket", "ausgeben", "pt-BR")), 0o600); err != nil {
			t.Fatal(err)
		}

		mussKommandozeile(t, umgebung, "sprachpaket", "loeschen", "pt-BR")
		mussKommandozeile(t, umgebung, "sprachpaket", "einlesen", pfad)
		pruefeUebersetzung(t, mussKommandozeile(t, umgebung, "sprachpaket", "zeige", "pt-BR"))

		if _, _, code := kommandozeile(t, umgebung, "sprachpaket", "anlegen", "ja", "Japonês", "en"); code != 1 {
			t.Errorf("eine eingebaute Sprache gab %d, erwartet 1 (fachliches Nein)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := eigenesZuhause(t)

		mussMCP(t, umgebung, "schreiben.sprachpaket.anlegen",
			map[string]any{"kennung": "pt-BR", "name": "Português", "basis": "en"})
		mussMCP(t, umgebung, "schreiben.sprachpaket.aendern", map[string]any{
			"kennung": "pt-BR", "lizenz": "CC0-1.0", "texte": map[string]string{sprachsatz: "Nenhum set aberto"},
		})

		if liste := mussMCP(t, umgebung, "lesen.sprachpakete", map[string]any{}); !strings.Contains(liste, `"kennung": "pt-BR"`) {
			t.Errorf("die Liste: %s", liste)
		}

		var datei struct {
			Inhalt string `json:"inhalt"`
		}

		if err := json.Unmarshal([]byte(mussMCP(t, umgebung, "lesen.sprachpaket.datei",
			map[string]any{"kennung": "pt-BR"})), &datei); err != nil {
			t.Fatal(err)
		}

		mussMCP(t, umgebung, "schreiben.sprachpaket.loeschen", map[string]any{"kennung": "pt-BR"})
		mussMCP(t, umgebung, "schreiben.sprachpaket.einlesen", map[string]any{"inhalt": datei.Inhalt})
		pruefeUebersetzung(t, mussMCP(t, umgebung, "lesen.sprachpaket", map[string]any{"kennung": "pt-BR"}))

		if roh, istFehler := mcpRuf(t, umgebung, "schreiben.sprachpaket.anlegen",
			map[string]any{"kennung": "ja", "name": "Japonês", "basis": "en"}); !istFehler {
			t.Errorf("eine eingebaute Sprache kam durch: %s", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, eigenesZuhause(t))

		mussREST(t, an, http.MethodPost, "/sprachpakete", `{"kennung":"pt-BR","name":"Português","basis":"en"}`)
		mussREST(t, an, http.MethodPatch, "/sprachpakete/pt-BR",
			`{"lizenz":"CC0-1.0","texte":{"`+sprachsatz+`":"Nenhum set aberto"}}`)

		if liste := mussREST(t, an, http.MethodGet, "/sprachpakete", ""); !strings.Contains(liste, `"kennung":"pt-BR"`) {
			t.Errorf("die Liste: %s", liste)
		}

		var datei struct {
			Inhalt string `json:"inhalt"`
		}

		if err := json.Unmarshal([]byte(mussREST(t, an, http.MethodGet, "/sprachpakete/pt-BR/datei", "")), &datei); err != nil {
			t.Fatal(err)
		}

		mussREST(t, an, http.MethodDelete, "/sprachpakete/pt-BR", "")

		rumpf, err := json.Marshal(map[string]any{"inhalt": datei.Inhalt})
		if err != nil {
			t.Fatal(err)
		}

		mussREST(t, an, http.MethodPost, "/sprachpakete/einlesen", string(rumpf))
		pruefeUebersetzung(t, mussREST(t, an, http.MethodGet, "/sprachpakete/pt-BR", ""))

		if status, antwort := an.ruf(t, http.MethodPost, "/sprachpakete",
			`{"kennung":"ja","name":"Japonês","basis":"en"}`); status < 400 || status >= 500 {
			t.Errorf("eine eingebaute Sprache gab %d (%s), erwartet ein fachliches Nein", status, antwort)
		}
	})
}

// pruefeUebersetzung sichert zu, dass genau der eine Satz übersetzt ist —
// mit dem deutschen, gegen den er übersetzt wurde.
func pruefeUebersetzung(t *testing.T, roh string) {
	t.Helper()

	var antwort struct {
		Paket struct {
			Uebersetzt int    `json:"uebersetzt"`
			Lizenz     string `json:"lizenz"`
		} `json:"paket"`
		Texte map[string]struct {
			Text   string `json:"text"`
			Quelle string `json:"quelle"`
		} `json:"texte"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("keine Übersetzung: %v (%s)", err, roh)
	}

	satz := antwort.Texte[sprachsatz]
	if antwort.Paket.Uebersetzt != 1 || antwort.Paket.Lizenz != "CC0-1.0" ||
		satz.Text != "Nenhum set aberto" || satz.Quelle != "Kein Set offen" {
		t.Errorf("die Übersetzung: %+v", antwort)
	}
}
