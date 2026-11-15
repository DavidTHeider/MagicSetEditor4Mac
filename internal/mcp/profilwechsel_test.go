package mcp_test

// **Der Profilwechsel über MCP** — Werkstück 3 von Phase 4.12
// (#1351).

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Zwei Werkzeuge: eines nennt das gezeigte Profil, eines wechselt.**
func TestDieProfilwerkzeugeTunWasSieSagen(t *testing.T) {
	t.Run("lesen", func(t *testing.T) {
		probe := &nutzerprobe{
			gezeigt: mcp.Profilzeile{ID: "10", NutzerID: "1", Name: "Mein Profil"},
		}

		antwort := rufeRoh(t, mitNutzern(t, probe, "lesen.profil"), `{}`)

		if !strings.Contains(antwort, `"name": "Mein Profil"`) {
			t.Errorf("die Antwort nennt das Profil nicht: %s", antwort)
		}
	})

	// **Wer arbeitet, und der Wechsel dorthin** — Werkstück 17 (#1400).
	t.Run("wer arbeitet", func(t *testing.T) {
		antwort := rufeRoh(t, mitNutzern(t, &nutzerprobe{}, "lesen.werarbeitet"), `{}`)

		if !strings.Contains(antwort, `"nutzerID": "7"`) {
			t.Errorf("die Antwort nennt niemanden: %s", antwort)
		}
	})

	// **Der Arbeitskontext** — Werkstück O1 von Phase 5.2 (#1425).
	t.Run("der Arbeitskontext", func(t *testing.T) {
		antwort := rufeRoh(t, mitNutzern(t, &nutzerprobe{}, "lesen.kontext"), `{}`)

		if !strings.Contains(antwort, `"rolle": "verwalter"`) || !strings.Contains(antwort, `"revision": 3`) {
			t.Errorf("die Antwort nennt den Kontext nicht: %s", antwort)
		}
	})

	t.Run("den Nutzer wechseln", func(t *testing.T) {
		probe := &nutzerprobe{}

		antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.nutzer.wechseln"),
			`{"nutzerID":"2"}`)

		if probe.nutzergewechselt != "2" {
			t.Errorf("gewechselt wurde auf %q", probe.nutzergewechselt)
		}

		if !strings.Contains(antwort, `"nutzerID": "2"`) {
			t.Errorf("die Antwort nennt den Menschen nicht: %s", antwort)
		}
	})

	t.Run("wechseln", func(t *testing.T) {
		probe := &nutzerprobe{}

		antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.profil.wechseln"),
			`{"profilID":"10"}`)

		if probe.gewechselt != "10" {
			t.Errorf("gewechselt wurde auf %q", probe.gewechselt)
		}

		if !strings.Contains(antwort, `"id": "10"`) {
			t.Errorf("die Antwort nennt das Profil nicht: %s", antwort)
		}
	})
}

// **Und jeder Weg meldet seinen Fehler.**
func TestJedesProfilwerkzeugMeldetSeinenFehler(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *nutzerprobe
	}{
		"lesen: die Naht klagt": {
			"lesen.profil", `{}`, &nutzerprobe{fehler: errors.New("kaputt")},
		},
		"wechseln: krummer Rumpf": {"schreiben.profil.wechseln", `{`, &nutzerprobe{}},

		// **Und dieselben drei für den Nutzerwechsel** — Werkstück 17 von
		// Phase 4.12 (#1400).
		// Zwei Werkzeuge, die gleich aussehen und Verschiedenes meinen; ein
		// Lauf über nur eines ließe die Wege des anderen unbetreten.
		"wer arbeitet: die Naht klagt": {
			"lesen.werarbeitet", `{}`, &nutzerprobe{fehler: errors.New("kaputt")},
		},
		"kontext: die Naht klagt": {
			"lesen.kontext", `{}`, &nutzerprobe{fehler: errors.New("kaputt")},
		},
		"nutzerwechsel: krummer Rumpf": {
			"schreiben.nutzer.wechseln", `{`, &nutzerprobe{},
		},
		"nutzerwechsel: die Naht klagt": {
			"schreiben.nutzer.wechseln", `{"nutzerID":"1"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
		"wechseln: die Naht klagt": {
			"schreiben.profil.wechseln", `{"profilID":"1"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if !rufeMitFehler(
				t, mitNutzern(t, fall.probe, fall.werkzeug), fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Der Server stattet jeden Ruf aus** — Werkstück 3 (#1351).
//
// # Was er hält
//
// Die Naht wird **je Ruf** befragt, nicht einmal beim Start. Warum das der
// Unterschied ist, steht an `mcp.Server.ausstatten`: eine Sitzung lebt lange,
// und `schreiben.profil.wechseln` ändert den Stand, aus dem der nächste Ruf
// liest.
//
// **Zwei Rufe in einer Sitzung** — mit einem wäre die Aussage nicht zu
// treffen: einmal ausstatten sähe genauso aus.
func TestDerServerStattetJedenRufAus(t *testing.T) {
	gezaehlt := 0

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend,
		Erklaerung: "gibt zurück, was im Kontext steht",
		Schema:     map[string]any{"type": "object"},
		Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
			gesehen, gesetzt := ctx.Value(probeschluessel{}).(string)
			if !gesetzt {
				return "", errors.New("der Kontext trug nichts")
			}

			return gesehen, nil
		},
	}}).MitAusstattung(func(ctx context.Context) (context.Context, error) {
		gezaehlt++

		return context.WithValue(ctx, probeschluessel{},
			"Ruf "+strconv.Itoa(gezaehlt)), nil
	})

	zeilen := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if len(zeilen) != 2 {
		t.Fatalf("%d Antworten: %v", len(zeilen), zeilen)
	}

	for nummer, erwartet := range []string{"Ruf 1", "Ruf 2"} {
		if !strings.Contains(zeilen[nummer], erwartet) {
			t.Errorf("Antwort %d nennt %q nicht: %s", nummer+1, erwartet, zeilen[nummer])
		}
	}
}

// **Und ein Fehler beim Ausstatten kommt als fachlicher Fehler an.**
//
// **Nicht als Protokollfehler**: das Gegenüber soll ihn dem Modell zeigen
// können, statt die Verbindung für kaputt zu halten — dieselbe Regel wie bei
// jedem anderen Fehler eines Werkzeugs.
//
// **Und das Werkzeug läuft gar nicht.** Ein Ruf ohne Profil, der trotzdem
// stattfindet, ist genau der Fehler, gegen den die Naht steht.
func TestEinFehlerBeimAusstattenKommtAlsFachlicherAn(t *testing.T) {
	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Erklaerung: "tut nichts",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			t.Error("das Werkzeug lief, obwohl das Ausstatten scheiterte")

			return "", nil
		},
	}}).MitAusstattung(func(ctx context.Context) (context.Context, error) {
		return ctx, errors.New("kein Profil")
	})

	zeilen := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if len(zeilen) != 1 || !strings.Contains(zeilen[0], "kein Profil") ||
		!strings.Contains(zeilen[0], `"isError":true`) {
		t.Errorf("der Fehler ging verloren: %v", zeilen)
	}
}

// probeschluessel trägt den Probewert durch den Kontext.
type probeschluessel struct{}

// antworten schickt Zeilen an einen Server und gibt seine Antwortzeilen.
func antworten(t *testing.T, server *mcp.Server, zeilen ...string) []string {
	t.Helper()

	var aus strings.Builder

	if err := server.Bediene(t.Context(),
		strings.NewReader(strings.Join(zeilen, "\n")+"\n"), &aus); err != nil {
		t.Fatalf("bedienen: %v", err)
	}

	ausgabe := strings.TrimSpace(aus.String())
	if ausgabe == "" {
		return nil
	}

	return strings.Split(ausgabe, "\n")
}

// **Zwei Kontowerkzeuge: eines nennt, eines löst.**
//
// **Und `anmelden` gibt es nicht** — der Fluss öffnet einen Browser und
// wartet, bis ein Mensch dort etwas tut. Ein Sprachmodell hat keinen Browser
// und keine Hände; ein Werkzeug dafür wäre eine Zusage ohne Einlösung (§10).
func TestDieKontowerkzeugeTunWasSieSagen(t *testing.T) {
	t.Run("lesen", func(t *testing.T) {
		probe := &nutzerprobe{
			konten: []mcp.Kontozeile{{Anbieter: "github", Anzeigename: "Wer"}},
		}

		antwort := rufeRoh(t, mitNutzern(t, probe, "lesen.anbieterkonten"),
			`{"nutzerID":"1"}`)

		if !strings.Contains(antwort, `"anbieter": "github"`) {
			t.Errorf("die Antwort nennt das Konto nicht: %s", antwort)
		}
	})

	// **Eine leere Liste ist eine Antwort, `null` ist keine** (§3).
	t.Run("lesen: leer", func(t *testing.T) {
		antwort := rufeRoh(t, mitNutzern(t, &nutzerprobe{}, "lesen.anbieterkonten"),
			`{"nutzerID":"1"}`)

		if strings.Contains(antwort, `"konten": null`) {
			t.Errorf("die Liste ist null: %s", antwort)
		}
	})

	t.Run("lösen", func(t *testing.T) {
		probe := &nutzerprobe{}

		antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.konto.loesen"),
			`{"nutzerID":"1","anbieter":"github"}`)

		if probe.geloest != "github" {
			t.Errorf("gelöst wurde %q", probe.geloest)
		}

		if !strings.Contains(antwort, "github") {
			t.Errorf("die Antwort nennt den Anbieter nicht: %s", antwort)
		}
	})
}

// **Und jeder Weg meldet seinen Fehler.**
func TestJedesKontowerkzeugMeldetSeinenFehler(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *nutzerprobe
	}{
		"lesen: krummer Rumpf": {"lesen.anbieterkonten", `{`, &nutzerprobe{}},
		"lesen: die Naht klagt": {
			"lesen.anbieterkonten", `{"nutzerID":"1"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
		"lösen: krummer Rumpf": {"schreiben.konto.loesen", `{`, &nutzerprobe{}},
		"lösen: die Naht klagt": {
			"schreiben.konto.loesen", `{"nutzerID":"1","anbieter":"github"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if !rufeMitFehler(
				t, mitNutzern(t, fall.probe, fall.werkzeug), fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Die vier Pflegewerkzeuge tun, was sie sagen.**
func TestDiePflegewerkzeugeTunWasSieSagen(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		erwartet  string
	}{
		"Nutzer ändern": {
			"schreiben.nutzer.aendern", `{"nutzerID":"1","name":"Umbenannt"}`, "Umbenannt",
		},
		"Profil ändern": {
			"schreiben.profil.aendern", `{"profilID":"1","name":"Legacy"}`, "Legacy",
		},
		"Nutzer löschen": {
			"schreiben.nutzer.loeschen", `{"nutzerID":"7","name":"Wer"}`, "7",
		},
		"Profil löschen": {
			"schreiben.profil.loeschen", `{"profilID":"8","name":"Wer"}`, "8",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			probe := &nutzerprobe{}

			antwort := rufeRoh(t, mitNutzern(t, probe, fall.werkzeug), fall.argumente)

			if !strings.Contains(antwort, fall.erwartet) {
				t.Errorf("die Antwort nennt %q nicht: %s", fall.erwartet, antwort)
			}
		})
	}
}

// **Und jeder Weg meldet seinen Fehler.**
func TestJedesPflegewerkzeugMeldetSeinenFehler(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *nutzerprobe
	}{
		"Nutzer ändern: krummer Rumpf":  {"schreiben.nutzer.aendern", `{`, &nutzerprobe{}},
		"Profil ändern: krummer Rumpf":  {"schreiben.profil.aendern", `{`, &nutzerprobe{}},
		"Nutzer löschen: krummer Rumpf": {"schreiben.nutzer.loeschen", `{`, &nutzerprobe{}},
		"Profil löschen: krummer Rumpf": {"schreiben.profil.loeschen", `{`, &nutzerprobe{}},
		"Nutzer ändern: die Naht klagt": {
			"schreiben.nutzer.aendern", `{"nutzerID":"1","name":"X"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
		"Profil ändern: die Naht klagt": {
			"schreiben.profil.aendern", `{"profilID":"1","name":"X"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
		"Nutzer löschen: die Naht klagt": {
			"schreiben.nutzer.loeschen", `{"nutzerID":"1","name":"X"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
		"Profil löschen: die Naht klagt": {
			"schreiben.profil.loeschen", `{"profilID":"1","name":"X"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if !rufeMitFehler(
				t, mitNutzern(t, fall.probe, fall.werkzeug), fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
