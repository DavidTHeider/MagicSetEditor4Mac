//go:build fremdpakete

package app

// **Die Ausfuhr über eine Vorlage, gemessen** — PERF-01
// (#2168).
//
// Der Weg der Flächen ohne Fenster: [Vorlagenweg.Ausfuehren] über einem
// eingefahrenen Satz des Kanons und dem vollen Pack — derselbe Aufbau wie
// `kopflosweg_bench_intern_test.go`. **Es gab ihn als Messung nicht**, und
// PERF-03 setzt genau hier an: ohne ihn stünde eine Änderung an der
// Kartenabbildung der Ausfuhr ohne Vorher-Zahl aus diesem Repositorium da.
//
// **Archenemy, weil der Befund dort lag**: der Auditlauf maß ein Set mit Spiel
// `magic` und Stil `archenemy`, und die berechneten Felder eines Stils sind
// die Mehrarbeit, die eine Optimierung nicht streichen darf (VK-01b, #2182).

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ausfuhrvorlage ist die schlichteste Vorlage des Packs — sie schreibt je
// Karte eine Zeile und misst damit den Weg und nicht die Vorlage.
const ausfuhrvorlage = "magic-simple-exporter.mse-export-template"

func BenchmarkWegAusfuhr(b *testing.B) {
	lauf := kopflosAufbauen(b, "Archenemy.mse-set")
	b.Logf("Archenemy.mse-set mit %d Karten, Vorlage %s", lauf.karten, ausfuhrvorlage)

	weg, err := VorlagenwegUeber(lauf.db, domain.UrheberKommandozeile)
	if err != nil {
		b.Fatalf("Vorlagenweg: %v", err)
	}

	ziel := filepath.Join(b.TempDir(), "ausfuhr.txt")
	ausfuehren := func() {
		ergebnis, fehler := weg.Ausfuehren(lauf.ctx, lauf.setID, ausfuhrvorlage, ziel, nil)
		if fehler != nil {
			b.Fatalf("Ausfuhr: %v %v", fehler, ergebnis.Befunde)
		}
	}

	// **Gemessen wird der zweite und jeder weitere Lauf** — der erste lädt
	// Spiel, Stil und Vorlage und steht in „Karte, frisches Werk" daneben.
	ausfuehren()

	for b.Loop() {
		ausfuehren()
	}

	haelt(b, "eine Ausfuhr über eine Vorlage", schwelleAusfuhr)
}

// schwelleAusfuhr gilt für eine Ausfuhr der 70 Archenemy-Karten über
// `magic-simple-exporter`.
//
// **Gemessen am 2026-10-07** (`docs/LEISTUNGSSCHNITTE.md`): 685–721 ms über
// vier Läufe. Die Schwelle fängt eine Verdopplung und lässt einem belasteten
// Rechner Luft — dieselbe Abwägung wie bei den übrigen Wegen.
const schwelleAusfuhr = 1500 * time.Millisecond
