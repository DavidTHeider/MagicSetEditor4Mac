package werkzeuge

// **Die README nennt jede Fähigkeit mit ihren Wegen an Kommandozeile, Dienst
// und MCP** — der Abschluss von 0.6.0 (#1660).
//
// # Der Anlass
//
// Sie nannte von 372 Endpunkten des Dienstes **sechs** und von 264 Befehlen
// der Kommandozeile **keinen** — nur die MCP-Werkzeuge standen vollständig da,
// weil `readmewerkzeuge_test.go` sie hält. Wer wissen wollte, ob es etwas an
// einer Fläche gibt, fand es nicht in der README, sondern in
// `docs/FAEHIGKEITEN.md`.
//
// # Warum erzeugt und nicht geschrieben
//
// **Weil es eine zweite Aufzählung wäre** — die, die niemand nachführt (§9).
// `docs/FAEHIGKEITEN.md` steht unter zwei Wächtern, die jeden Weg gegen den
// Quelltext halten; die Liste hier ist ihre verkürzte Gestalt, ohne die
// Fensterbindungen und ohne die Gründe. `SCHNITTSTELLEN_SCHREIBEN=1` schreibt
// sie zwischen die beiden Marken.
//
// # Was er nicht prüft
//
//   - **Ob ein Weg das Richtige tut.** Das sagen die Flächenläufe
//     (`internal/flaechen`).
//   - **Die Prosa um die Liste.** Sie steht außerhalb der Marken.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	schnittstellenAnfang = "<!-- schnittstellen:anfang — erzeugt aus docs/FAEHIGKEITEN.md; " +
		"SCHNITTSTELLEN_SCHREIBEN=1 go test ./internal/werkzeuge -run Schnittstelle -->"
	schnittstellenEnde = "<!-- schnittstellen:ende -->"
)

// schnittstellenabschnitt ist ein Abschnitt des Katalogs mit seinen Zeilen.
type schnittstellenabschnitt struct {
	titel  string
	zeilen []string
}

// schnittstellenliste baut die Liste aus dem Katalog.
func schnittstellenliste(t *testing.T) (liste string, gezaehlt int) {
	t.Helper()

	var abschnitte []schnittstellenabschnitt

	for _, zeile := range strings.Split(lies(t, "docs/FAEHIGKEITEN.md"), "\n") {
		if titel, ok := strings.CutPrefix(zeile, "## "); ok {
			abschnitte = append(abschnitte, schnittstellenabschnitt{titel: titel})

			continue
		}

		// **Dieselbe Auswahl wie `faehigkeitszeilen`** — sonst zählten die
		// beiden Wächter verschiedene Kataloge.
		if !strings.HasPrefix(zeile, "| ") || strings.Count(zeile, "|") != spaltenzahl ||
			strings.HasPrefix(zeile, "| ---") || len(abschnitte) == 0 {
			continue
		}

		zellen := strings.Split(zeile, "|")
		if strings.TrimSpace(zellen[1]) == "Fähigkeit" {
			continue
		}

		letzter := &abschnitte[len(abschnitte)-1]
		letzter.zeilen = append(letzter.zeilen, fmt.Sprintf("| %s | %s | %s | %s | %s |",
			strings.TrimSpace(zellen[1]), amFenster(zellen[2]),
			wegOderStrich(zellen[4]), wegOderStrich(zellen[5]), wegOderStrich(zellen[3])))
	}

	var aufbau strings.Builder

	for _, abschnitt := range abschnitte {
		if len(abschnitt.zeilen) == 0 {
			continue
		}

		gezaehlt += len(abschnitt.zeilen)

		fmt.Fprintf(&aufbau, "<details>\n<summary><b>%s</b> · %d Fähigkeiten</summary>\n\n",
			abschnitt.titel, len(abschnitt.zeilen))
		aufbau.WriteString("| Fähigkeit | Fenster | Kommandozeile | Dienst (REST) | MCP |\n")
		aufbau.WriteString("| --- | :---: | --- | --- | --- |\n")
		aufbau.WriteString(strings.Join(abschnitt.zeilen, "\n"))
		aufbau.WriteString("\n\n</details>\n\n")
	}

	return strings.TrimSpace(aufbau.String()), gezaehlt
}

// amFenster macht aus den Bindungen ein Zeichen: die Namen sagen einem Leser
// der README nichts, ob es die Fähigkeit dort gibt, schon.
func amFenster(zelle string) string {
	if wegOderStrich(zelle) == "—" {
		return "—"
	}

	return "✓"
}

// wegOderStrich gibt die Zelle, oder einen Strich, wo sie leer ist.
func wegOderStrich(zelle string) string {
	zelle = strings.TrimSpace(zelle)
	if zelle == "" {
		return "—"
	}

	return zelle
}

func TestDieReadmeNenntJedeSchnittstelle(t *testing.T) {
	erwartet, gezaehlt := schnittstellenliste(t)

	if gezaehlt != len(faehigkeitszeilen(t)) {
		t.Fatalf("die Liste hat %d Zeilen, der Katalog %d Fähigkeiten — der Leser greift daneben",
			gezaehlt, len(faehigkeitszeilen(t)))
	}

	liesmich := lies(t, "README.md")

	vorn, rest, gefunden := strings.Cut(liesmich, schnittstellenAnfang)
	if !gefunden {
		t.Fatalf("README.md hat die Marke %q nicht", schnittstellenAnfang)
	}

	vorhanden, hinten, gefunden := strings.Cut(rest, schnittstellenEnde)
	if !gefunden {
		t.Fatalf("README.md hat die Marke %q nicht", schnittstellenEnde)
	}

	if os.Getenv("SCHNITTSTELLEN_SCHREIBEN") == "1" {
		neu := vorn + schnittstellenAnfang + "\n\n" + erwartet + "\n\n" + schnittstellenEnde + hinten

		if fehler := os.WriteFile(filepath.Join(wurzel, "README.md"), []byte(neu), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		return
	}

	if strings.TrimSpace(vorhanden) != erwartet {
		t.Errorf("die Schnittstellenliste in README.md ist nicht, was docs/FAEHIGKEITEN.md ergibt — " +
			"SCHNITTSTELLEN_SCHREIBEN=1 schreibt sie neu")
	}
}
