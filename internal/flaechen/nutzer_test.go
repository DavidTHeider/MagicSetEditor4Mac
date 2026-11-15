// **Nutzer und Profile über jede Fläche** — Werkstück 1 von Phase 4.12
// (#1351).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Die Zusage, um die es geht
//
// Wer über Kommandozeile, MCP oder REST einen Nutzer anlegt, sieht ihn danach
// **an jeder** Fläche — und dieselbe fachliche Abweisung bekommt er überall.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestEinNutzerWirdAnJederFlaecheAngelegt(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if _, fehler, code := kommandozeile(t, umgebung,
			"nutzer-neu", "Zweite"); code != 0 {
			t.Fatalf("anlegen: %d (%s)", code, fehler)
		}

		if liste := mussKommandozeile(t, umgebung, "nutzer"); !strings.Contains(
			liste, "Zweite") {
			t.Errorf("die Liste führt den Nutzer nicht:\n%s", liste)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.nutzer.neu",
			map[string]any{"name": "Zweite"}); istFehler {
			t.Fatalf("anlegen: %s", text)
		}

		if text, istFehler := mcpRuf(t, umgebung, "lesen.nutzer",
			map[string]any{}); istFehler || !strings.Contains(text, "Zweite") {
			t.Errorf("die Liste führt den Nutzer nicht: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/nutzer",
			`{"name":"Zweite"}`); status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		status, rumpf := an.ruf(t, http.MethodGet, "/nutzer", "")
		if status != http.StatusOK || !strings.Contains(rumpf, "Zweite") {
			t.Errorf("GET: %d (%s)", status, rumpf)
		}
	})
}

// **Und dieselbe fachliche Abweisung: ein Name mit Rändern** (§16).
//
// Nicht beschnitten, sondern abgewiesen — „David" und „David " wären zwei
// Zeilen mit demselben sichtbaren Namen in der Profilwahl.
func TestEinNameMitRaendernWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if _, _, code := kommandozeile(t, umgebung,
			"nutzer-neu", " David "); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.nutzer.neu",
			map[string]any{"name": " David "}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/nutzer",
			`{"name":" David "}`); status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// **Ein frischer Bestand führt genau einen Nutzer mit einem Profil.**
//
// Die Zusage aus Vermerk 0066, Entscheidung 4 — und sie gilt an jeder Fläche,
// weil sie in der Wanderung steht und nicht in einer Fläche.
func TestEinFrischerBestandFuehrtEinenNutzerAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	liste := mussKommandozeile(t, umgebung, "nutzer", "--json")

	// **Ein Nutzer, ein Profil** — nicht null, sonst zeigte die Anwendung beim
	// ersten Start eine Anmeldemaske, wo vorher ein Fenster aufging.
	if strings.Count(liste, `"name"`) < 2 {
		t.Errorf("die Liste führt keinen Nutzer mit Profil:\n%s", liste)
	}
}
