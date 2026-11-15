// **Die Anbieterkonten an jeder Fläche** — Werkstück 4 von Phase 4.12
// (#1351).
//
// # Was hier läuft und was nicht
//
// **Der Anmeldefluss läuft hier nicht.** Er öffnet einen Browser und wartet
// auf einen Menschen; ein Lauf, der das täte, risse dem Prüfenden das Fenster
// auf. Was er tut, prüft `internal/app/anmeldenaht_test.go` **ganz** — mit
// `state`, PKCE und Rückruf-Horcher, gegen eine Gegenstelle.
//
// **Hier läuft, was ohne Menschen geht**: die Anbieterliste, die Konten und
// der Rückweg. Und zwar über **jede** Fläche mit demselben Ergebnis (§16).
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Die Anbieterliste steht an der Kommandozeile — ohne Kontakt nach draußen.**
func TestDieAnbieterlisteStehtAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	aus := mussKommandozeile(t, umgebung, "anbieter")

	for _, name := range []string{"github", "google", "microsoft", "x"} {
		if !strings.Contains(aus, name) {
			t.Errorf("%q fehlt: %q", name, aus)
		}
	}

	// **Apple steht belegt nicht darin** (Vermerk 0066).
	if strings.Contains(aus, "apple") {
		t.Errorf("Apple steht in der Liste: %q", aus)
	}
}

// **Ohne eingetragene Kennung ist die Anmeldung ein fachliches Nein.**
//
// **Und das ist der Normalfall**: dieses Programm liefert keine Zugangsdaten
// mit. Die Meldung sagt, was zu tun ist — sie ist die Fläche der Zusage.
func TestOhneKennungIstDieAnmeldungEinNein(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	_, fehler, code := kommandozeile(t, umgebung, "anmelden", "github")
	if code != 1 {
		t.Fatalf("Rückgabewert %d, erwartet war 1", code)
	}

	if !strings.Contains(fehler, "eigene Anwendung") {
		t.Errorf("die Meldung sagt nicht, was zu tun ist: %q", fehler)
	}
}

// **Die Konten sind an jeder Fläche leer — und der Rückweg meldet dasselbe.**
//
// # Warum „leer" die Abnahme ist
//
// **Weil ein Konto nur über eine echte Anmeldung entsteht**, und die braucht
// einen Anbieter, eine Kennung und einen Menschen. Was hier zählt, ist die
// **Naht**: dass jede Fläche dieselbe Antwort gibt und dieselbe Ablehnung.
func TestDieKontenAntwortenAnJederFlaecheGleich(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	t.Run("Kommandozeile", func(t *testing.T) {
		if aus := mussKommandozeile(t, umgebung, "konten", "1", "--json"); !strings.Contains(
			aus, `"konten": []`) {
			t.Errorf("die Liste ist nicht leer: %s", aus)
		}

		// **Und ein Konto, das es nicht gibt, zu lösen ist ein Nein.**
		if _, _, code := kommandozeile(t, umgebung, "konten", "1", "github"); code != 1 {
			t.Errorf("Rückgabewert %d, erwartet war 1", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		if aus := mussMCP(t, umgebung, "lesen.anbieterkonten",
			map[string]any{"nutzerID": "1"}); !strings.Contains(aus, `"konten": []`) {
			t.Errorf("die Liste ist nicht leer: %s", aus)
		}

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.konto.loesen",
			map[string]any{"nutzerID": "1", "anbieter": "github"})
		if !istFehler {
			t.Errorf("das fehlende Konto kam durch: %s", roh)
		}
	})

	t.Run("Dienst", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		if aus := mussREST(t, an, http.MethodGet, "/nutzer/1/konten", ""); !strings.Contains(
			aus, `"konten":[]`) {
			t.Errorf("die Liste ist nicht leer: %s", aus)
		}

		status, antwort := an.ruf(t, http.MethodDelete, "/nutzer/1/konten/github", "")
		if status != http.StatusNotFound {
			t.Errorf("Status %d, erwartet war 404: %s", status, antwort)
		}
	})
}
