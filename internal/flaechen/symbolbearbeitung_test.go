// **Ein Symbol der Bibliothek bearbeiten — an jeder Fläche** — DOM-02.AK-02
// (#1697).
//
// # Der Ablauf
//
// Ein Symbol trägt ein Set → die Teile lesen → eine neue Fassung sichern: die
// Antwort nennt das Set als veraltet → nachziehen nennt es → zurücknehmen
// nennt es wieder, denn jetzt trägt es die ersetzte Fassung. **Dieselbe
// Antwort an jeder Fläche** (§16).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// bearbeitungswege sind die Schritte, je Fläche einmal geschrieben.
type bearbeitungswege struct {
	teile         func(id string) string
	sichern       func(id, teile string) string
	nachziehen    func(id string) string
	zuruecknehmen func(id string) string
}

// eingesetztesSymbol legt über die Kommandozeile ein Symbol in die Bibliothek
// und in ein Set — der Aufbau, nicht der Gegenstand.
func eingesetztesSymbol(t *testing.T, umgebung []string) (id string) {
	t.Helper()

	setID := neuesSet(t, umgebung, "Kronenset")
	datei := filepath.Join(t.TempDir(), "krone.mse-symbol")

	if err := os.WriteFile(datei, msepaket.SchreibeSymbolzeichnung(msepaket.Standardsymbol()), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	id = strings.TrimSpace(mussKommandozeile(t, umgebung, "symbolbibliothek", "einlesen", datei, "Krone"))
	mussKommandozeile(t, umgebung, "symbolbibliothek", "einsetzen", id, setID)

	return id
}

// kreisteile rechnet die Teile eines Kreises — dieselbe Rechnung an jeder
// Fläche, hier über die Kommandozeile.
func kreisteile(t *testing.T, umgebung []string) string {
	t.Helper()

	aus, fehler, code := mitEingabeUndCode(t, umgebung, `{"art":"Kreis"}`,
		"symbol-rechnen", "lesen.symbol.grundform", "--json")
	if code != 0 {
		t.Fatalf("die Grundform: %d (%s)", code, fehler)
	}

	return teileAus(t, aus)
}

func pruefeBearbeitung(t *testing.T, flaeche, id, kreis string, wege bearbeitungswege) {
	t.Helper()

	if teile := wege.teile(id); !strings.Contains(teile, `"teile"`) {
		t.Errorf("%s: die Teile: %q", flaeche, teile)
	}

	// **Der Reihe nach**: jeder Schritt setzt den vorigen voraus — und nach
	// dem Sichern und dem Zurücknehmen wird nachgesehen, was gilt.
	gesichert := wege.sichern(id, kreis)
	nachSichern := wege.teile(id)
	nachgezogen := wege.nachziehen(id)
	zurueck := wege.zuruecknehmen(id)
	nachZurueck := wege.teile(id)

	for name, antwort := range map[string]string{
		"sichern": gesichert, "nachziehen": nachgezogen, "zurücknehmen": zurueck,
	} {
		if !strings.Contains(antwort, "Kronenset") {
			t.Errorf("%s: %s nennt das Set nicht: %q", flaeche, name, antwort)
		}
	}

	if !strings.Contains(nachSichern, "Ellipse") || strings.Contains(nachZurueck, "Ellipse") {
		t.Errorf("%s: nach dem Sichern %q, nach dem Zurücknehmen %q", flaeche, nachSichern, nachZurueck)
	}
}

func TestEinSymbolWirdUeberallGleichBearbeitet(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		id := eingesetztesSymbol(t, umgebung)

		pruefeBearbeitung(t, "Kommandozeile", id, kreisteile(t, umgebung), bearbeitungswege{
			teile: func(id string) string { return mussKommandozeile(t, umgebung, "symbolbibliothek", "teile", id) },
			sichern: func(id, teile string) string {
				aus, fehler, code := mitEingabeUndCode(t, umgebung, teile, "symbolbibliothek", "sichern", id, "--json")
				if code != 0 {
					t.Fatalf("sichern: %d (%s)", code, fehler)
				}

				return aus
			},
			nachziehen: func(id string) string {
				return mussKommandozeile(t, umgebung, "symbolbibliothek", "nachziehen", id, "--json")
			},
			zuruecknehmen: func(id string) string {
				return mussKommandozeile(t, umgebung, "symbolbibliothek", "zuruecknehmen", id, "--json")
			},
		})
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		id := eingesetztesSymbol(t, umgebung)

		pruefeBearbeitung(t, "MCP", id, kreisteile(t, umgebung), bearbeitungswege{
			teile: func(id string) string {
				return mussMCP(t, umgebung, "lesen.symbolbibliothek.teile", map[string]any{"id": zahlAus(t, id)})
			},
			sichern: func(id, teile string) string {
				return mussMCP(t, umgebung, "schreiben.symbolbibliothek.sichern",
					map[string]any{"id": zahlAus(t, id), "teile": json.RawMessage(teile)})
			},
			nachziehen: func(id string) string {
				return mussMCP(t, umgebung, "schreiben.symbolbibliothek.nachziehen", map[string]any{"id": zahlAus(t, id)})
			},
			zuruecknehmen: func(id string) string {
				return mussMCP(t, umgebung, "schreiben.symbolbibliothek.zuruecknehmen", map[string]any{"id": zahlAus(t, id)})
			},
		})
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		id := eingesetztesSymbol(t, umgebung)
		an := starteDienst(t, umgebung)

		pruefeBearbeitung(t, "REST", id, kreisteile(t, umgebung), bearbeitungswege{
			teile: func(id string) string { return mussREST(t, an, http.MethodGet, "/symbole/"+id+"/teile", "") },
			sichern: func(id, teile string) string {
				return mussREST(t, an, http.MethodPut, "/symbole/"+id+"/teile", `{"teile":`+teile+`}`)
			},
			nachziehen: func(id string) string {
				return mussREST(t, an, http.MethodPost, "/symbole/"+id+"/nachzug", "")
			},
			zuruecknehmen: func(id string) string {
				return mussREST(t, an, http.MethodPost, "/symbole/"+id+"/zuruecknahme", "")
			},
		})
	})
}
