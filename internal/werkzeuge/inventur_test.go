package werkzeuge_test

// **Die Inventur jeder Seite ist maschinenlesbar und belegt** — UX-01.AK-01
// (#1673).
//
// # Was er prüft
//
//  1. **Jeder Schlüssel aus `navigation.ts` steht genau einmal in
//     `docs/inventur.json`**, und keiner steht dort, den es nicht gibt.
//  2. **Jeder Zustand hat einen Stand, und der Stand trägt seine Deckung**:
//     `belegt` und `anwendungsweit` nennen eine Testdatei und einen
//     Testnamen, und **beide gibt es**; `offen` und `nicht vorhanden` nennen
//     einen Grund.
//  3. **Die native Prüfung ist offen, bis ein Abzug sie belegt** — ein
//     anderer Stand verlangt eine Datei, die es gibt.
//
// # Was er nicht prüft
//
// **Ob der genannte Test den Zustand wirklich zeigt.** Er sieht, dass es den
// Test gibt, nicht was er zusichert; das bleibt die Durchsicht, die die
// Inventur gebaut hat. Und **ob eine Seite Unteransichten hat, die hier
// fehlen** — die Aufgaben stammen aus `UX-ARCHITEKTUR.md`, nicht aus dem
// Quelltext.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// inventurzustand ist ein Zustand einer Seite.
type inventurzustand struct {
	Stand string `json:"stand"`
	Beleg string `json:"beleg"`
	Grund string `json:"grund"`
}

// inventurseite ist eine Zeile der Matrix.
type inventurseite struct {
	Art        string                     `json:"art"`
	Komponente string                     `json:"komponente"`
	Aufgaben   []string                   `json:"aufgaben"`
	Einstiege  []string                   `json:"einstiege"`
	Zustaende  map[string]inventurzustand `json:"zustaende"`
	Pruefung   map[string]inventurzustand `json:"pruefung"`
}

// inventurzustaende sind die sieben Zustände aus S12.NAT-03 (#1672) — die
// fünf aus UX-01.AK-01 und dazu **befüllt** und **fehlende Berechtigung**,
// die das Audit zu 0.6.0 (`seiten-zustaende.csv`) neben ihnen führt.
var inventurzustaende = []string{
	"leer", "befuellt", "laden", "fehler", "offline", "nurLesen", "fehlendeBerechtigung",
}

func TestDieInventurFuehrtJedeSeiteMitBelegtenZustaenden(t *testing.T) {
	t.Parallel()

	var inventur struct {
		Seiten map[string]inventurseite `json:"seiten"`
	}

	if fehler := json.Unmarshal([]byte(lies(t, "docs/inventur.json")), &inventur); fehler != nil {
		t.Fatalf("docs/inventur.json ist kein JSON: %v", fehler)
	}

	gefuehrt := navigationsseiten(t)

	for _, schluessel := range gefuehrt {
		if _, da := inventur.Seiten[schluessel]; !da {
			t.Errorf("navigation.ts führt %q — die Inventur nicht", schluessel)
		}
	}

	for schluessel, seite := range inventur.Seiten {
		if !slices.Contains(gefuehrt, schluessel) {
			t.Errorf("die Inventur führt %q — navigation.ts kennt die Seite nicht", schluessel)

			continue
		}

		pruefeInventurseite(t, schluessel, seite)
	}
}

// pruefeInventurseite hält eine Zeile gegen den Baum.
func pruefeInventurseite(t *testing.T, schluessel string, seite inventurseite) {
	t.Helper()

	if _, fehler := os.Stat(filepath.Join("..", "..", seite.Komponente)); fehler != nil {
		t.Errorf("%s: die Komponente %q gibt es nicht", schluessel, seite.Komponente)
	}

	if len(seite.Einstiege) == 0 || len(seite.Aufgaben) == 0 ||
		slices.Contains(seite.Aufgaben, "—") {
		t.Errorf("%s: ohne Einstieg oder ohne Aufgabe — ein Strich ist keine", schluessel)
	}

	if len(seite.Zustaende) != len(inventurzustaende) {
		t.Errorf("%s: %d Zustände, verlangt sind %v", schluessel, len(seite.Zustaende), inventurzustaende)
	}

	for _, name := range inventurzustaende {
		zustand, da := seite.Zustaende[name]
		if !da {
			t.Errorf("%s: der Zustand %q fehlt", schluessel, name)

			continue
		}

		pruefeDeckung(t, schluessel+"."+name, zustand,
			[]string{"belegt", "anwendungsweit"}, []string{"offen", "nicht vorhanden"})
	}

	nativ := seite.Pruefung["nativ"]
	pruefeDeckung(t, schluessel+".nativ", nativ, []string{"bestanden", "befund"}, []string{"offen"})
}

// pruefeDeckung verlangt zu jedem Stand, was ihn trägt.
func pruefeDeckung(t *testing.T, wo string, zustand inventurzustand, mitBeleg, mitGrund []string) {
	t.Helper()

	switch {
	case slices.Contains(mitBeleg, zustand.Stand):
		pruefeBeleg(t, wo, zustand.Beleg)
	case slices.Contains(mitGrund, zustand.Stand):
		if strings.TrimSpace(zustand.Grund) == "" {
			t.Errorf("%s steht auf %q ohne Grund", wo, zustand.Stand)
		}
	default:
		t.Errorf("%s: der Stand %q ist keiner", wo, zustand.Stand)
	}
}

// pruefeBeleg verlangt Datei und Testnamen — oder eine Datei allein.
func pruefeBeleg(t *testing.T, wo, beleg string) {
	t.Helper()

	datei, name, mitName := strings.Cut(beleg, " :: ")

	roh, fehler := os.ReadFile(filepath.Join("..", "..", datei)) //nolint:gosec // ein Pfad aus der Inventur im Baum
	if fehler != nil {
		t.Errorf("%s: der Beleg %q liegt nicht im Baum", wo, datei)

		return
	}

	// **Ein Go-Test heißt wie seine Funktion** — seit S12.NAT-03 belegt ein
	// Lauf in `internal/app` die fehlende Berechtigung anwendungsweit.
	gesucht := "'" + name + "'"
	if strings.HasSuffix(datei, ".go") {
		gesucht = "func " + name + "("
	}

	if mitName && !strings.Contains(string(roh), gesucht) {
		t.Errorf("%s: %s kennt keinen Test %q", wo, datei, name)
	}
}
