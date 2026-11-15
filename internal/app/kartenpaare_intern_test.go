//go:build fremdpakete

package app

// **Dieselbe Karte vor, sofort und nach anderen** — PERF-04
// (#2171).
//
// Das alte Wiederholungsmaß verglich die erste Karte mit dem Median
// **anderer** Karten. Hier wird gepaart: Karte A, sofort noch einmal A, dann
// fünf andere, dann wieder A — fünf Runden über das **behaltene** Werk der
// Flächen ohne Fenster ([Zeichennaht]). Je Zeichnung stehen da: die Dauer, der
// SHA-256 der PNG-Datei und **was die Zwischenspeicher dabei gezählt haben**
// — abgetastete Ebenen, Paketbilder, Spielbilder ([werkspeicher]). Die
// Cacheklasse kommt aus den Zählern und nicht aus der Wandzeit.
//
// **Was es nicht gibt, wird nicht vermutet**: einen Speicher für das fertige
// Kartenbild hat der Kartenweg nicht (`werkspeicher.go` nennt drei, und das
// Spielbildfach gehört dem Tisch). Ein wiederholtes A rechnet seinen Plan und
// rastert neu; was es sparen kann, sind Ebenen und Paketbilder.
//
// Läuft nur mit `MSE_KARTENPAARE_BELEG` (dort landet der Beleg als JSON) —
// aus `scripts/leistungsschnitte.mjs`, Schnitt `paare`.
//
// **Kein Messweg mit Schwelle**, und deshalb nicht `…_bench_intern_test.go`
// (`internal/werkzeuge/messschwellen_test.go`): PERF-04 ist eine Diagnose
// ohne Budget. Eine Schwelle hier wäre eine Zusage, die niemand gegeben hat.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// paarschritt ist eine Zeichnung und was sie gekostet hat.
type paarschritt struct {
	Runde    int              `json:"runde"`
	Schritt  string           `json:"schritt"`
	Karte    int64            `json:"karte"`
	Ms       float64          `json:"ms"`
	Sha256   string           `json:"sha256"`
	Speicher map[string][]int `json:"speicher"` // Name → [Treffer, Fehlgriffe] dieser Zeichnung
}

func BenchmarkKartenpaare(b *testing.B) {
	ziel := os.Getenv("MSE_KARTENPAARE_BELEG")
	if ziel == "" {
		b.Skip("MSE_KARTENPAARE_BELEG nicht gesetzt — ein Messlauf, kein Wächter")
	}

	lauf := kopflosAufbauen(b, "Archenemy.mse-set")
	werk := kopflosNaht(b, lauf)

	karten, err := storage.NeuerSetLeser(lauf.db).Karten(lauf.ctx, lauf.setID)
	if err != nil || len(karten) < 13 {
		b.Fatalf("Karten: %v (%d)", err, len(karten))
	}

	quellen, _ := werkspeicher(werk)
	zaehle := func() map[string][2]int64 {
		stand := map[string][2]int64{}
		for _, quelle := range quellen {
			z := quelle()
			stand[z.Name] = [2]int64{z.Treffer, z.Fehlgriffe}
		}

		return stand
	}

	var schritte []paarschritt

	zeichne := func(runde int, schritt string, karte int64) {
		vorher := zaehle()
		anfang := time.Now()

		if _, fehler := werk.SchreibeBild(lauf.ctx, karte, lauf.ziel, render.GueteMittel); fehler != nil {
			b.Fatalf("SchreibeBild %d: %v", karte, fehler)
		}

		dauer := time.Since(anfang)

		roh, fehler := os.ReadFile(lauf.ziel)
		if fehler != nil || len(roh) == 0 {
			b.Fatalf("kein Bild für Karte %d: %v", karte, fehler)
		}

		summe := sha256.Sum256(roh)
		speicher := map[string][]int{}
		for name, nachher := range zaehle() {
			speicher[name] = []int{int(nachher[0] - vorher[name][0]), int(nachher[1] - vorher[name][1])}
		}

		schritte = append(schritte, paarschritt{
			Runde: runde, Schritt: schritt, Karte: karte, Ms: float64(dauer.Microseconds()) / 1000,
			Sha256: hex.EncodeToString(summe[:]), Speicher: speicher,
		})
	}

	a := karten[3].ID

	for b.Loop() {
		schritte = schritte[:0]
		zeichne(0, "A kalt", a)

		for runde := 1; runde <= 5; runde++ {
			zeichne(runde, "A vorher", a)
			zeichne(runde, "A sofort", a)

			for andere := range 5 {
				zeichne(runde, "B", karten[3+runde+andere].ID)
			}

			zeichne(runde, "A nachher", a)
		}
	}

	roh, err := json.MarshalIndent(schritte, "", "  ")
	if err != nil {
		b.Fatal(err)
	}

	if err := os.WriteFile(ziel, roh, 0o600); err != nil { // #nosec G703 -- der Pfad kommt aus der Umgebung dessen, der misst
		b.Fatal(err)
	}
}
