package flaechen_test

// **Derselbe Druck zweimal nennt die Zeile — an jeder Fläche** — DOM-03,
// S04.NAT-01 und 03 (#1698).
//
// Das Fenster fragt nach der Menge; ohne Fenster gibt es keinen Knopf, der zur
// vorhandenen Zeile führt, und **ihre Kennung ist dort der Weg**. Geprüft wird
// deshalb, dass die Absage an Kommandozeile, MCP und Dienst die Kennung der
// ersten Aufnahme nennt und die Menge unverändert lässt.

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestDerselbeDruckZweimalNenntDieZeileAnJederFlaeche(t *testing.T) {
	for name, aufnehmen := range map[string]func(t *testing.T, umgebung []string) (antwort string, abgewiesen bool){
		"Kommandozeile": func(t *testing.T, umgebung []string) (string, bool) {
			t.Helper()

			aus, fehler, code := kommandozeile(t, umgebung, "exemplar", probedruck, "2",
				"played", "de", "foil", "2026-08-01", "4.17", "EUR", "aus der Abnahme")

			return aus + fehler, code != 0
		},
		"MCP": func(t *testing.T, umgebung []string) (string, bool) {
			t.Helper()

			return mcpRuf(t, umgebung, "schreiben.exemplar", vollesExemplar())
		},
		"Dienst": func(t *testing.T, umgebung []string) (string, bool) {
			t.Helper()

			status, rumpf := starteDienst(t, umgebung).ruf(t, http.MethodPost, "/sammlung",
				alsRumpf(t, vollesExemplar()))

			return rumpf, status != http.StatusCreated
		},
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := eigenesZuhause(t)

			if antwort, abgewiesen := aufnehmen(t, umgebung); abgewiesen {
				t.Fatalf("die erste Aufnahme: %s", antwort)
			}

			erste := ersteAusHuelle(t, mussKommandozeile(t, umgebung, "--json", "sammlung"))

			antwort, abgewiesen := aufnehmen(t, umgebung)
			if !abgewiesen {
				t.Fatalf("derselbe Druck kam zweimal hinein: %s", antwort)
			}

			if erwartet := fmt.Sprintf("Exemplar %d", erste.ID); !strings.Contains(antwort, erwartet) {
				t.Errorf("die Absage nennt %q nicht: %s", erwartet, antwort)
			}

			pruefeExemplar(t, ersteAusHuelle(t, mussKommandozeile(t, umgebung, "--json", "sammlung")), 2)
		})
	}
}
