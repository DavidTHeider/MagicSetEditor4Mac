package werkzeuge

import (
	"encoding/json"
	"strings"
	"testing"
)

// **Jede Vorabfassung hat ihre Ausstiegsfrage** — J03
// (#2267).
//
// Eine Abhängigkeit mit Vorabkennung (`-alpha`, `-beta`, `-pre`, eine
// Pseudofassung) oder mit der Hauptfassung 0 sagt nichts über ihre
// Schnittstelle zu. Der Abschnitt nennt je eine **wann** gewechselt oder
// zurückgegangen wird, und dieser Wächter hält ihn vollständig: im Programm
// — Go direkt und npm `dependencies` — und mit der gepinnten Fassung.
//
// **Was er nicht hält**: ob Pflege, Drift und Auslöser stimmen. Sie sind an
// einem Tag nachgelesen, und der steht über der Tafel. Werkzeuge der
// Pipeline zählt er nicht mit — sie gehen nicht ins Programm.
const abschnittVorab = "Vorabfassungen und ihre Ausstiegsfragen"

// vorabfassung sagt, ob eine Fassung nichts über ihre Schnittstelle zusagt.
func vorabfassung(fassung string) bool {
	fassung = strings.TrimPrefix(fassung, "v")

	return strings.HasPrefix(fassung, "0.") || strings.Contains(fassung, "-")
}

// vorabzeile ist eine Zeile der Tafel: Fassung und Auslöser.
type vorabzeile struct{ fassung, ausloeser string }

func vorabtafel(t *testing.T) map[string]vorabzeile {
	t.Helper()

	var (
		zeilen      = map[string]vorabzeile{}
		imAbschnitt bool
		kopf        bool
	)

	for _, roheZeile := range strings.Split(string(projektdatei(t, inventardatei)), "\n") {
		if titel, ok := strings.CutPrefix(roheZeile, "## "); ok {
			imAbschnitt, kopf = strings.TrimSpace(titel) == abschnittVorab, false

			continue
		}

		if !imAbschnitt || !strings.HasPrefix(strings.TrimSpace(roheZeile), "|") {
			continue
		}

		spalten := tabellenspalten(roheZeile)

		switch {
		case len(spalten) != 7:
			t.Fatalf("%s, %q: die Zeile hat %d Spalten statt sieben: %s",
				inventardatei, abschnittVorab, len(spalten), roheZeile)
		case !kopf:
			kopf = true
		case strings.HasPrefix(spalten[0], "---"):
		default:
			zeilen[ohneAuszeichnung(spalten[0])] = vorabzeile{
				fassung:   ohneAuszeichnung(spalten[1]),
				ausloeser: strings.TrimSpace(spalten[6]),
			}
		}
	}

	return zeilen
}

func TestJedeVorabfassungHatIhreAusstiegsfrage(t *testing.T) {
	gesucht := map[string]string{}

	for modul, fassung := range direkteAnforderungen(t, "go.mod") {
		if vorabfassung(fassung) {
			gesucht[modul] = fassung
		}
	}

	var paket struct {
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(projektdatei(t, "frontend/package.json"), &paket); err != nil {
		t.Fatalf("frontend/package.json: %v", err)
	}

	sperre := sperrdatei(t)

	for name := range paket.Dependencies {
		if fassung := sperre[name].Version; vorabfassung(fassung) {
			gesucht[name] = fassung
		}
	}

	tafel := vorabtafel(t)

	for name, fassung := range gesucht {
		zeile, da := tafel[name]

		switch {
		case !da:
			t.Errorf("%s ist auf %s gepinnt, steht aber nicht unter %q in %s", name, fassung, abschnittVorab, inventardatei)
		case zeile.fassung != fassung:
			t.Errorf("%s: die Tafel nennt %s, gepinnt ist %s", name, zeile.fassung, fassung)
		case zeile.ausloeser == "":
			t.Errorf("%s: ohne Auslöser für Wechsel und Rückweg", name)
		}
	}

	for name := range tafel {
		if _, da := gesucht[name]; !da {
			t.Errorf("%s steht unter %q, ist aber keine Vorabfassung im Programm mehr", name, abschnittVorab)
		}
	}
}

func TestVorabfassungErkenntDieKlasse(t *testing.T) {
	for fassung, erwartet := range map[string]bool{
		"v0.11.0-alpha.6": true,
		"v3.0.0-beta.27":  true,
		"3.0.0-beta.27":   true,
		"v1.103.0-pre.0.20260916030321-a2263542f260": true,
		"v0.35.4": true,
		"0.8.15":  true,
		"v1.3.2":  false,
		"19.3.0":  false,
		"v10.0.0": false,
	} {
		if vorabfassung(fassung) != erwartet {
			t.Errorf("vorabfassung(%q) = %v, erwartet %v", fassung, !erwartet, erwartet)
		}
	}
}
