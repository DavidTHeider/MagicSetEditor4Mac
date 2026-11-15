// Wächter über das Abhängigkeitsinventar — ENG-04
// (#1729).
//
// `docs/ABHAENGIGKEITEN.md` nennt jede **direkte** Abhängigkeit mit Fassung,
// Lizenz und der Stelle, die sie verantwortet (AK-01). Das ist eine Liste, die
// Menschen lesen — und eine solche Liste veraltet still: ein falscher Eintrag
// macht nichts kaputt. Also zählt hier jemand nach, **in beide Richtungen**:
// jede Pinnung hat ihre Zeile, jede Zeile ihre Pinnung, in derselben Fassung.
//
// **Und die Spalte „Verantwortet" wird am Quelltext gemessen**, nicht
// geglaubt: eine Go-Abhängigkeit wird genau von den Paketen importiert, die
// dort stehen, ein npm-Paket nur unterhalb der genannten Stellen. Das hält
// zugleich AK-04 — **eine unbenutzte Abhängigkeit wird rot**, weil sie keinen
// Importeur hat, der sie verantworten könnte.
//
// # Was dieser Wächter nicht prüft
//
//   - **Die Lizenz einer Go-Abhängigkeit.** Sie steht nachgelesen in der
//     Tabelle, aber lesen ließe sie sich nur aus dem Modulzwischenspeicher, und
//     der gehört nicht in einen Einheitstest. Bei npm steht sie in der
//     Sperrdatei und wird verglichen.
//   - **Die transitiven Abhängigkeiten.** Ihre Bilanz im Inventar trägt ein
//     Datum; vollständig stehen sie in den Stücklisten jeder Freigabe.
//   - **Ob die Begründung einer neuen Abhängigkeit trägt** (AK-04). Dass sie
//     eine Zeile und einen Verantwortlichen hat, hält er; ob Größe, Lizenz
//     und Pflege gut abgewogen sind, sieht nur ein Mensch.
//   - **Ob ein Commit-Hash zu der Fassung gehört, die sein Kommentar nennt.**
//     Dafür bräuchte es das Netz — wie in `frontendpinnung_test.go`.
package werkzeuge

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// Die Überschriften der Tabellen im Inventar.
const (
	abschnittGo         = "Go — direkt"
	abschnittWerkzeuge  = "Go — Werkzeuge der Pipeline"
	abschnittProgramm   = "npm — im Programm"
	abschnittEntwickeln = "npm — nur beim Entwickeln"
	abschnittActions    = "Actions"
)

// stilleNutzung ist der Grund, aus dem ein npm-Paket gebraucht wird, das kein
// Quelltext einführt: entweder verlangt ein anderes direktes Paket es als
// Peer, oder eine Datei nennt es auf andere Weise — dann steht dort ein Beleg.
type stilleNutzung struct {
	peerVon string
	datei   string
	beleg   string
}

// stilleNutzer sind die Pakete ohne Einfuhr. **Jeder Eintrag wird geprüft**:
// der Peer steht in der Sperrdatei, der Beleg in der Datei — und wer das
// Paket später doch einführt, macht den Eintrag rot, statt ihn stehen zu
// lassen.
var stilleNutzer = map[string]stilleNutzung{
	// `@mantine/core` verlangt beide in derselben Fassung (§1).
	"@mantine/hooks": {peerVon: "@mantine/core"},
	// `@testing-library/react` 16 bringt `dom` nicht mehr mit.
	"@testing-library/dom": {peerVon: "@testing-library/react"},

	"@types/node":         {datei: "frontend/tsconfig.json", beleg: `"node"`},
	"@types/react":        {datei: "frontend/tsconfig.json", beleg: `"jsx": "react-jsx"`},
	"@types/react-dom":    {datei: "frontend/src/main.tsx", beleg: `'react-dom/client'`},
	"@vitest/coverage-v8": {datei: "frontend/vite.config.ts", beleg: `provider: 'v8'`},
	"jsdom":               {datei: "frontend/vite.config.ts", beleg: `environment: 'jsdom'`},
	"eslint":              {datei: "frontend/package.json", beleg: `"lint": "eslint`},
	"prettier":            {datei: "frontend/package.json", beleg: `"format": "prettier`},
}

func TestJedeDirekteGoAbhaengigkeitStehtImInventar(t *testing.T) {
	gleicheMengen(t, abschnittGo, direkteAnforderungen(t, "go.mod"), inventar(t, abschnittGo))
}

// TestJedeGoAbhaengigkeitWirdVonIhremVerantwortlichenImportiert hält die
// Spalte „Verantwortet" gegen die Importe: **genau** die genannten Pakete.
//
// **Genau, nicht mindestens.** Wer ein weiteres Paket an eine Abhängigkeit
// bindet, weitet ihre Reichweite — und das soll im Inventar zu sehen sein,
// nicht erst beim nächsten Heben, wenn es an einer unerwarteten Stelle bricht.
func TestJedeGoAbhaengigkeitWirdVonIhremVerantwortlichenImportiert(t *testing.T) {
	direkt := direkteAnforderungen(t, "go.mod")
	importeure := goImporteure(t, direkt)

	for modul, zeile := range inventar(t, abschnittGo) {
		t.Run(modul, func(t *testing.T) {
			tatsaechlich := make([]string, 0, len(importeure[modul]))
			for paket := range importeure[modul] {
				tatsaechlich = append(tatsaechlich, paket)
			}

			slices.Sort(tatsaechlich)

			genannt := slices.Sorted(slices.Values(zeile.verantwortet))

			if len(tatsaechlich) == 0 {
				t.Fatalf("kein Paket importiert %s — eine unbenutzte Abhängigkeit wird entfernt (AK-04)", modul)
			}

			if !slices.Equal(genannt, tatsaechlich) {
				t.Errorf("verantwortet laut Inventar: %v, importiert von: %v", genannt, tatsaechlich)
			}
		})
	}
}

// TestJedesWerkzeugDerPipelineStehtImInventar hält die `tool`-Direktiven in
// `tools/go.mod` gegen die Tabelle — und dass die verantwortenden Skripte
// das Werkzeug wirklich rufen.
func TestJedesWerkzeugDerPipelineStehtImInventar(t *testing.T) {
	werkzeugmodul := gomodLesen(t, "tools/go.mod")
	gepinnt := map[string]string{}
	aufrufe := map[string]string{}

	for _, werkzeug := range werkzeugmodul.werkzeuge {
		modul := modulZu(werkzeug, werkzeugmodul.anforderungen)
		gepinnt[modul] = werkzeugmodul.anforderungen[modul]
		aufrufe[modul] = werkzeug[strings.LastIndex(werkzeug, "/")+1:]
	}

	zeilen := inventar(t, abschnittWerkzeuge)
	gleicheMengen(t, abschnittWerkzeuge, gepinnt, zeilen)

	for modul, zeile := range zeilen {
		for _, skript := range zeile.verantwortet {
			if aufruf := aufrufe[modul]; aufruf != "" && !strings.Contains(string(projektdatei(t, skript)), aufruf) {
				t.Errorf("%s: %s ruft %s nicht", modul, skript, aufruf)
			}
		}
	}
}

// TestJedesNpmPaketStehtImInventar trennt, was ins Programm geht, von dem,
// was nur beim Entwickeln gebraucht wird — so, wie `package.json` es trennt.
// Fassung und Lizenz kommen aus der **Sperrdatei** (§1).
func TestJedesNpmPaketStehtImInventar(t *testing.T) {
	var paket struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(projektdatei(t, "frontend/package.json"), &paket); err != nil {
		t.Fatalf("frontend/package.json: %v", err)
	}

	sperre := sperrdatei(t)

	for abschnitt, pakete := range map[string]map[string]string{
		abschnittProgramm:   paket.Dependencies,
		abschnittEntwickeln: paket.DevDependencies,
	} {
		gesperrt := map[string]string{}
		for name := range pakete {
			gesperrt[name] = sperre[name].Version
		}

		zeilen := inventar(t, abschnitt)
		gleicheMengen(t, abschnitt, gesperrt, zeilen)

		for name, zeile := range zeilen {
			if lizenz := sperre[name].License; pakete[name] != "" && lizenz != zeile.lizenz {
				t.Errorf("%s: das Inventar nennt die Lizenz %s, die Sperrdatei %s", name, zeile.lizenz, lizenz)
			}
		}
	}
}

// TestJedesNpmPaketWirdDortBenutztWoEsVerantwortetIst hält die Spalte
// „Verantwortet" gegen die Einfuhren: jede liegt unterhalb einer genannten
// Stelle, und jede genannte Stelle führt das Paket wirklich ein.
func TestJedesNpmPaketWirdDortBenutztWoEsVerantwortetIst(t *testing.T) {
	importeure := npmImporteure(t)
	sperre := sperrdatei(t)

	for _, abschnitt := range []string{abschnittProgramm, abschnittEntwickeln} {
		for name, zeile := range inventar(t, abschnitt) {
			t.Run(name, func(t *testing.T) {
				if still, ok := stilleNutzer[name]; ok {
					pruefeStilleNutzung(t, name, still, zeile, importeure[name], sperre)

					return
				}

				pruefeEinfuhren(t, name, zeile.verantwortet, importeure[name])
			})
		}
	}
}

// pruefeEinfuhren prüft ein Paket, das Quelltext einführt.
func pruefeEinfuhren(t *testing.T, name string, verantwortet, einfuhren []string) {
	t.Helper()

	if len(einfuhren) == 0 {
		t.Fatalf("keine Datei führt %s ein — eine unbenutzte Abhängigkeit wird entfernt (AK-04)", name)
	}

	for _, datei := range einfuhren {
		if !slices.ContainsFunc(verantwortet, func(stelle string) bool { return strings.HasPrefix(datei, stelle) }) {
			t.Errorf("%s führt %s ein und liegt außerhalb von %v", datei, name, verantwortet)
		}
	}

	for _, stelle := range verantwortet {
		if !slices.ContainsFunc(einfuhren, func(datei string) bool { return strings.HasPrefix(datei, stelle) }) {
			t.Errorf("%s ist als verantwortlich genannt, führt %s aber nicht ein", stelle, name)
		}
	}
}

// pruefeStilleNutzung prüft ein Paket ohne Einfuhr gegen seinen Grund.
func pruefeStilleNutzung(t *testing.T, name string, still stilleNutzung, zeile inventarzeile,
	einfuhren []string, sperre map[string]gesperrtesPaket,
) {
	t.Helper()

	if len(einfuhren) > 0 {
		t.Errorf("%s wird eingeführt (%v) — der Eintrag in stilleNutzer ist überholt", name, einfuhren)
	}

	if still.peerVon != "" {
		if _, verlangt := sperre[still.peerVon].PeerDependencies[name]; !verlangt {
			t.Errorf("%s verlangt %s nicht als Peer", still.peerVon, name)
		}

		if !slices.Equal(zeile.verantwortet, []string{still.peerVon}) {
			t.Errorf("verantwortet ist %s, das Inventar nennt %v", still.peerVon, zeile.verantwortet)
		}

		return
	}

	if !strings.Contains(string(projektdatei(t, still.datei)), still.beleg) {
		t.Errorf("%s nennt %s nicht mehr — der Grund für %s ist weg", still.datei, still.beleg, name)
	}

	if !slices.Equal(zeile.verantwortet, []string{still.datei}) {
		t.Errorf("verantwortet ist %s, das Inventar nennt %v", still.datei, zeile.verantwortet)
	}
}

// TestJedeActionIstAnEinenCommitGebundenUndInventarisiert hält die Actions
// gegen die Tabelle — samt der Fassung aus dem Kommentar hinter dem Hash.
func TestJedeActionIstAnEinenCommitGebundenUndInventarisiert(t *testing.T) {
	fassungen, orte := actionsLesen(t)
	zeilen := inventar(t, abschnittActions)

	gleicheMengen(t, abschnittActions, fassungen, zeilen)

	for action, zeile := range zeilen {
		if ort := orte[action]; ort != "" && !slices.Contains(zeile.verantwortet, ort) {
			t.Errorf("%s steht in %s, das Inventar nennt %v", action, ort, zeile.verantwortet)
		}
	}
}

// gleicheMengen vergleicht die Pinnungen mit einer Tabelle — in beide
// Richtungen und mit der Fassung. Eine Zeile ohne Lizenz oder ohne
// Verantwortlichen ist ein Befund (AK-01 verlangt beides).
func gleicheMengen(t *testing.T, abschnitt string, gepinnt map[string]string, zeilen map[string]inventarzeile) {
	t.Helper()

	for name, fassung := range gepinnt {
		zeile, steht := zeilen[name]

		switch {
		case !steht:
			t.Errorf("%s ist gepinnt, fehlt aber in %s unter %q", name, inventardatei, abschnitt)
		case strings.TrimPrefix(zeile.fassung, "v") != strings.TrimPrefix(fassung, "v"):
			t.Errorf("%s: das Inventar nennt %s, gepinnt ist %s", name, zeile.fassung, fassung)
		}
	}

	for name, zeile := range zeilen {
		if _, gibtEs := gepinnt[name]; !gibtEs {
			t.Errorf("%s steht in %s unter %q, ist aber nicht gepinnt", name, inventardatei, abschnitt)
		}

		if zeile.lizenz == "" || len(zeile.verantwortet) == 0 {
			t.Errorf("%s: Lizenz und Verantwortlicher gehören in die Zeile (AK-01)", name)
		}
	}
}

// gesperrtesPaket ist ein Eintrag der Sperrdatei, soweit er hier gebraucht wird.
type gesperrtesPaket struct {
	Version          string            `json:"version"`
	License          string            `json:"license"`
	PeerDependencies map[string]string `json:"peerDependencies"`
}

// sperrdatei liest die direkt installierten Pakete aus `package-lock.json`.
func sperrdatei(t *testing.T) map[string]gesperrtesPaket {
	t.Helper()

	var sperre struct {
		Packages map[string]gesperrtesPaket `json:"packages"`
	}

	if err := json.Unmarshal(projektdatei(t, "frontend/package-lock.json"), &sperre); err != nil {
		t.Fatalf("frontend/package-lock.json: %v", err)
	}

	pakete := map[string]gesperrtesPaket{}

	for pfad, eintrag := range sperre.Packages {
		if name, direkt := strings.CutPrefix(pfad, "node_modules/"); direkt && !strings.Contains(name, "/node_modules/") {
			pakete[name] = eintrag
		}
	}

	return pakete
}
