package flaechen_test

// **Die Beispiele der Administration laufen** — OPS04
// (#2350),
// Befund A-22 des Audits 0.9.0: dort stand ein `curl` mit einer Kopfzeile, die
// kein Code mehr kannte, und niemand hatte es gemerkt.
//
// # Was er hält
//
//   - **Jedes `curl`** in `docs/ADMINISTRATION.md`: es trägt
//     `Authorization: Bearer $MSE_MARKE`, fragt `http://127.0.0.1:7777`, und
//     sein Weg antwortet am echten Dienst eines Wegwerfbestands mit 200 — oder
//     mit dem Code, den sein Kommentar nennt.
//   - **Jedes `mse`-Beispiel** nennt einen Befehl, den es gibt; die lesenden
//     ohne Platzhalter laufen am Wegwerfbestand mit dem Rückgabewert, den die
//     Doku nennt.
//
// # Was er nicht hält
//
// **Die schreibenden Beispiele laufen nicht** — sie brauchen Pfade und
// Nachweise, die ein Mensch einsetzt —, und `mse aktualisierung` nicht, denn
// es fragt das Netz. Für sie gilt nur, dass es den Befehl gibt.

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// codebloecke gibt die Zeilen der `bash`-Blöcke einer Doku — Fortsetzungen
// mit `\` zu einer Zeile verbunden.
func codebloecke(t *testing.T, datei string) []string {
	t.Helper()

	roh, err := os.ReadFile(datei) //nolint:gosec // eine Datei des Baums
	if err != nil {
		t.Fatal(err)
	}

	var (
		zeilen  []string
		imBlock bool
		offen   string
	)

	for _, zeile := range strings.Split(string(roh), "\n") {
		switch {
		case strings.HasPrefix(zeile, "```bash"):
			imBlock = true
		case strings.HasPrefix(zeile, "```"):
			imBlock = false
		case imBlock && strings.HasSuffix(strings.TrimSpace(zeile), `\`):
			offen += strings.TrimSuffix(strings.TrimSpace(zeile), `\`) + " "
		case imBlock:
			zeilen = append(zeilen, offen+strings.TrimSpace(zeile))
			offen = ""
		}
	}

	return zeilen
}

var (
	curlweg        = regexp.MustCompile(`http://127\.0\.0\.1:7777(/[^\s"']*)`)
	erwarteteCodes = regexp.MustCompile(`#\s*(\d{3})(?:\s+oder\s+(\d{3}))?`)
)

func TestDieCurlBeispieleDerAdministrationLaufen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)
	gefahren := 0

	for _, zeile := range codebloecke(t, "../../docs/ADMINISTRATION.md") {
		if !strings.HasPrefix(zeile, "curl ") {
			continue
		}

		if !strings.Contains(zeile, `-H "Authorization: Bearer $MSE_MARKE"`) {
			t.Errorf("ohne die Marke des Dienstes: %s", zeile)
		}

		weg := curlweg.FindStringSubmatch(zeile)
		if weg == nil {
			t.Errorf("nicht an http://127.0.0.1:7777: %s", zeile)

			continue
		}

		erlaubt := []int{http.StatusOK}
		if codes := erwarteteCodes.FindStringSubmatch(zeile); codes != nil {
			erlaubt = nil

			for _, code := range codes[1:] {
				if code != "" {
					wert, _ := strconv.Atoi(code) //nolint:errcheck // drei Ziffern aus dem Muster
					erlaubt = append(erlaubt, wert)
				}
			}
		}

		if status, rumpf := an.ruf(t, http.MethodGet, weg[1], ""); !slices.Contains(erlaubt, status) {
			t.Errorf("%s: %d (%.200s), erwartet %v", weg[1], status, rumpf, erlaubt)
		}

		gefahren++
	}

	if gefahren < 5 {
		t.Errorf("nur %d curl-Beispiele gefahren — liest der Lauf die Doku noch?", gefahren)
	}
}

// lesendeBeispiele laufen am Wegwerfbestand — je mit den Rückgabewerten, die
// ihr Kommentar in der Doku nennt.
var lesendeBeispiele = map[string][]int{
	"mse gesundheit":                      {0, 1},
	"mse grenzen":                         {0},
	"mse kosten anbieter":                 {0},
	"mse meta":                            {0},
	"mse pflege-vorschau":                 {0},
	"mse rueckkehren":                     {0},
	"mse ruecksetzung-vorschau anwendung": {0},
	"mse ruecksetzung-vorschau profil":    {0},
	"mse ruecksetzungsarten":              {0},
	"mse ruecksprung-arten":               {0},
	"mse vertraeglichkeit":                {0, 1, 2},
}

func TestDieMseBeispieleDerAdministrationGibtEs(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	befehle := befehleDerKommandozeile(t)
	gelaufen := map[string]bool{}

	for _, zeile := range codebloecke(t, "../../docs/ADMINISTRATION.md") {
		if !strings.HasPrefix(zeile, "mse ") {
			continue
		}

		ohneKommentar := strings.TrimSpace(strings.SplitN(zeile, "#", 2)[0])
		worte := strings.Fields(ohneKommentar)

		if !befehle[worte[1]] {
			t.Errorf("»%s« ist kein Befehl von mse: %s", worte[1], zeile)
		}

		codes, lesend := lesendeBeispiele[ohneKommentar]
		if !lesend {
			continue
		}

		if _, fehler, code := kommandozeile(t, umgebung, worte[1:]...); !slices.Contains(codes, code) {
			t.Errorf("%s: Rückgabewert %d (%s), erwartet %v", ohneKommentar, code, fehler, codes)
		}

		gelaufen[ohneKommentar] = true
	}

	for beispiel := range lesendeBeispiele {
		if !gelaufen[beispiel] {
			t.Errorf("%s steht nicht mehr in der Doku — aus der Liste nehmen", beispiel)
		}
	}
}

// befehleDerKommandozeile liest die Befehle der Kommandozeile aus ihrer Verteilung.
func befehleDerKommandozeile(t *testing.T) map[string]bool {
	t.Helper()

	dateien, err := filepath.Glob("../../cmd/mse/*.go")
	if err != nil {
		t.Fatal(err)
	}

	worte := map[string]bool{}
	fall := regexp.MustCompile(`case ((?:"[a-z-]+"(?:,\s*)?)+):`)
	wort := regexp.MustCompile(`"([a-z-]+)"`)

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		roh, err := os.ReadFile(datei) //nolint:gosec // eine Datei des Baums
		if err != nil {
			t.Fatal(err)
		}

		for _, treffer := range fall.FindAllStringSubmatch(string(roh), -1) {
			for _, w := range wort.FindAllStringSubmatch(treffer[1], -1) {
				worte[w[1]] = true
			}
		}
	}

	return worte
}
