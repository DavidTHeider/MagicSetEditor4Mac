package flaechen_test

// **Ein Flächenlauf je Schalter** — OPS-10.AK-02
// (#1908).
//
// # Woher die Wege kommen
//
// **Aus der Tafel, die mit dem Programm ausgeliefert wird**, und nicht von
// Hand gewählt: je Schalter und Fläche der erste Weg, den **nur** dieser
// Schalter abschaltet. Eine Auswahl von Hand prüfte, was jemand für
// wichtig hielt.
//
// # Was er nicht hält
//
// **Jeden Weg.** Je Schalter und Fläche läuft einer durch den echten
// Prozess; dass jede Bindung des Fensters abweist, hält
// `internal/app/schaltertor_intern_test.go`, dass die Tafel vollständig ist,
// `internal/werkzeuge/schalterwege_test.go`.

import (
	"net/http"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestJederSchalterWeistAnDreiFlaechenGleichAb(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	tafel := domain.EingebauteWegetafel()
	fenster := fensterAus(t, umgebung)

	for _, schalter := range domain.Schalterverzeichnis() {
		kennung := schalter.Kennung
		befehl := nurUnter(tafel[domain.FlaecheCLI], kennung)
		werkzeug := nurUnter(tafel[domain.FlaecheMCP], kennung)
		endpunkt := nurUnter(tafel[domain.FlaecheREST], kennung)
		bindung := nurUnter(tafel[domain.FlaecheFenster], kennung)

		// **Dienst und Werkzeugserver haben keine Wege an diesen drei
		// Flächen** — sie schalten die Fläche selbst; ihr Lauf steht unten.
		if befehl == "" && werkzeug == "" && endpunkt == "" {
			continue
		}

		if befehl == "" || werkzeug == "" || endpunkt == "" {
			t.Errorf("%s: an einer Fläche kein eigener Weg — CLI %q, MCP %q, REST %q",
				kennung, befehl, werkzeug, endpunkt)

			continue
		}

		mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.SchalterEinstellung(kennung), domain.SchalterAus)

		abweisungGleich(t, kennung, abweisung{
			klasse: klasseAbgeschaltet,
			kern:   kennung,
			cli: func() (string, int) {
				aus, fehler, code := kommandozeile(t, umgebung, befehlsworte(befehl)...)

				return aus + fehler, code
			},
			mcp: func() (string, bool) {
				return mcpRuf(t, umgebung, werkzeug, map[string]any{})
			},
			rest: func() (int, string) {
				methode, weg, _ := strings.Cut(endpunkt, " ")

				return an.ruf(t, methode, strings.ReplaceAll(weg, "{}", "1"), "")
			},
			fenster: func() *app.Fehler {
				if bindung == "" {
					t.Fatalf("%s: die Wegetafel führt keine Bindung des Fensters", kennung)
				}

				return bindungRufen(t, fenster, bindung)
			},
		})

		mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.SchalterEinstellung(kennung), domain.SchalterAn)
	}
}

// **Der Dienst endet mit seinem Schalter** — OPS-10.AK-03: die Anfrage, die
// es bemerkt, bekommt die Abweisung, und danach antwortet niemand mehr.
func TestDerDienstEndetMitSeinemSchalter(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	if status, rumpf := an.ruf(t, http.MethodGet, "/grenzen", ""); status != http.StatusOK {
		t.Fatalf("vorher: %d %s", status, rumpf)
	}

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.SchalterEinstellung("dienst"), domain.SchalterAus)

	if status, rumpf := an.ruf(t, http.MethodGet, "/grenzen", ""); status != http.StatusForbidden ||
		!strings.Contains(rumpf, `"code":"abgeschaltet"`) {
		t.Fatalf("abgeschaltet: %d %s", status, rumpf)
	}

	anfrage, fehler := http.NewRequestWithContext(t.Context(), http.MethodGet, an.adresse+"/grenzen", http.NoBody)
	if fehler != nil {
		t.Fatal(fehler)
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)

	if kam, fehler := an.klient.Do(anfrage); fehler == nil {
		if fehler := kam.Body.Close(); fehler != nil {
			t.Error(fehler)
		}

		t.Errorf("der Dienst antwortet weiter: %d", kam.StatusCode)
	}
}

// nurUnter gibt den ersten Weg, den nur dieser Schalter abschaltet — jede
// Fähigkeit, die ihn nennt, hängt an ihm.
func nurUnter(wege map[string][][]string, kennung string) string {
	namen := make([]string, 0, len(wege))
	for weg := range wege {
		namen = append(namen, weg)
	}

	sort.Strings(namen)

	for _, weg := range namen {
		if !strings.HasPrefix(weg, "mse-") && !strings.Contains(weg, "--") &&
			!slices.ContainsFunc(wege[weg], func(schalter []string) bool {
				return !slices.Contains(schalter, kennung)
			}) {
			return weg
		}
	}

	return ""
}

// befehlsworte macht aus einem Befehlsweg die Argumente — ein Platzhalter
// wird eine Eins.
func befehlsworte(weg string) []string {
	woerter := strings.Fields(strings.TrimPrefix(weg, "mse "))
	for stelle, wort := range woerter {
		if strings.HasPrefix(wort, "<") {
			woerter[stelle] = "1"
		}
	}

	return woerter
}

// **Die Schalter sind an drei Flächen dieselben** — und sagen dasselbe über
// einen, der gerade aus ist.
func TestDieSchalterSindAnDreiFlaechenDieselben(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.SchalterEinstellung("ki"), domain.SchalterAus)

	zeile := mussKommandozeile(t, umgebung, "schalter")
	text, istFehler := mcpRuf(t, umgebung, "lesen.schalter", map[string]any{})
	status, rumpf := an.ruf(t, http.MethodGet, "/schalter", "")

	if istFehler || status != http.StatusOK {
		t.Fatalf("MCP %v %s, REST %d %s", istFehler, text, status, rumpf)
	}

	if !strings.Contains(zeile, "ki\tfunktion\tstabil\taus") || !strings.Contains(zeile, "spielen\tbereich\tstabil\tan") {
		t.Errorf("Kommandozeile:\n%s", zeile)
	}

	for wer, antwort := range map[string]string{"MCP": text, "REST": rumpf} {
		for _, schalter := range domain.Schalterverzeichnis() {
			if !strings.Contains(antwort, schalter.Kennung) {
				t.Errorf("%s nennt %q nicht:\n%s", wer, schalter.Kennung, antwort)
			}
		}

		if !strings.Contains(strings.Join(strings.Fields(antwort), ""), `"kennung":"ki","art":"funktion","reife":"stabil","vorgabe":"an","an":false`) {
			t.Errorf("%s sagt nicht, dass ki aus ist:\n%s", wer, antwort)
		}
	}
}
