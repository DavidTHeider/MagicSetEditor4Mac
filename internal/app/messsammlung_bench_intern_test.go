//go:build fremdpakete

// **Was fünftausend Exemplare kosten** — Werkstück 1 von Phase 4.5
// (#1292).
//
// # Warum es diese Messung gibt
//
// Die Abnahme von #1292 verlangt ausdrücklich **eine benannte Schranke für
// Aufbau und Suche, gemessen an einem Bestand dieser Größe, nicht an zwanzig
// Zeilen**. Am 2026-08-30 stand die Sammlung dieses Rechners auf **null**
// Exemplaren (`mse sammlung --json` gab `{"exemplare": []}`), und jeder Lauf
// über sie kam in unter 10 ms zurück. Eine Zusage daraus wäre eine Aussage
// über eine leere Tabelle (§10: die Reichweite einer Messung ist die
// Reichweite ihrer Aussage).
//
// # Was gemessen wird und was nicht
//
// Gemessen ist der **Weg durch die Schichten** bis an die Naht — Speicher,
// Fachkern, DTO. Was diese Datei **nicht** sieht:
//
//   - **Die JSON-Serialisierung der Wails-Naht.** Sie liegt zwischen den
//     Prüfebenen (§3), und dafür gibt es `scripts/sammlungsmass.mjs`: derselbe
//     Bestand, aber durch die laufende Brücke.
//   - **Das Zeichnen der Seite.** Fünftausend Zeilen in den DOM zu schreiben
//     ist ein Aufwand, den Go nicht kennt. Auch das misst die Brücke.
//   - **Ob 5.000 die richtige Zahl ist.** Sie ist die aus dem Issue; wer
//     50.000 besitzt, misst neu.
package app

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandsgroesse ist die Zahl der Exemplare, an der gemessen wird.
//
// **Fünftausend, weil #1292 sie nennt.** Das ist die Größe, ab der jemand
// seine Sammlung nicht mehr im Kopf hat.
const bestandsgroesse = 5000

// stueckzahl ist die Summe der Mengen — was `gefuellteSammlung` wirklich legt.
//
// **Gerechnet und nicht geschätzt**: die Menge läuft `1+zaehler%4`, also 1, 2,
// 3, 4, 1, … — im Mittel 2,5 Stück je Zeile.
const stueckzahl = bestandsgroesse / 4 * (1 + 2 + 3 + 4)

// Die Schwellen — gemessen am 2026-08-30 auf darwin/arm64 (Apple M5), bei
// `bestandsgroesse` Exemplaren **ohne einen einzigen Preis**, `-benchtime=10x`,
// drei Läufe:
//
//	Liste (ganz)       4,46–4,56 ms →   7 ms   (Faktor 1,5)
//	Liste (ein Druck) 17,1–17,3  µs → 100 µs   (Faktor 6 — siehe unten)
//	Portfolio          4,26–4,29 ms →   7 ms   (Faktor 1,6)
//
// **Sie gilt für diese Maschine** (0007): gemessen ist ein Rechner, keine
// Plattform. Ein Wächter gegen Rückschritt, keine Zusage an Nutzende.
//
// # Warum die mittlere so viel weiter liegt
//
// Sie misst eine Abfrage über einen Index, und 17 µs sind Rauschen — eine
// Schwelle bei 26 µs fiele bei jedem zweiten Lauf um. Was sie fangen soll, ist
// der Rückschritt auf einen **Tabellenscan**, und der kostet dasselbe wie die
// ganze Liste: 4,5 ms, also das 260-fache. 100 µs liegt weit darunter und
// steht trotzdem still.
//
// # Was hier zuerst stand
//
// `120 ms · 2 ms · 130 ms` — geraten, bevor gemessen war, und um das
// 25-fache daneben. Sie stehen hier, weil §10 dasselbe für Wächter sagt: eine
// Schwelle, die nichts fängt, sieht aus wie eine, die etwas fängt. Die
// **erste** Messung mit ihnen war grün und hat den Befund verdeckt, der
// darunter lag — 89 ms für ein Portfolio über eine leere Preistabelle, weil
// es einmal je Exemplar fragte
// (#1292).
const (
	schwelleSammlungGanz  = 7 * time.Millisecond
	schwelleSammlungDruck = 100 * time.Microsecond
	schwellePortfolio     = 7 * time.Millisecond
)

// **Die ganze Sammlung laden.** Das ist der Aufbau der Seite.
func BenchmarkWegSammlungsliste(b *testing.B) {
	naht := gefuellteSammlung(b)

	for b.Loop() {
		exemplare, err := naht.Sammlung(context.Background())
		if err != nil {
			b.Fatalf("die Sammlung lesen: %v", err)
		}

		// **Und es sind wirklich fünftausend.** Ein Lauf über eine leere
		// Tabelle ist schnell — und ohne diese Zeile grün.
		if len(exemplare) != bestandsgroesse {
			b.Fatalf("%d Exemplare, erwartet %d", len(exemplare), bestandsgroesse)
		}
	}

	haelt(b, "die ganze Sammlung laden", schwelleSammlungGanz)
}

// **Die Exemplare eines Drucks.** Das ist „besitze ich diese Karte?".
func BenchmarkWegSammlungFuerEinenDruck(b *testing.B) {
	naht := gefuellteSammlung(b)

	for b.Loop() {
		exemplare, err := naht.SammlungFuer(context.Background(), druckkennung(4711))
		if err != nil {
			b.Fatalf("den Druck lesen: %v", err)
		}

		if len(exemplare) == 0 {
			b.Fatal("kein Exemplar zu diesem Druck — die Vorrichtung misst nichts")
		}
	}

	haelt(b, "die Exemplare eines Drucks", schwelleSammlungDruck)
}

// **Die Sammlung zusammenrechnen.**
func BenchmarkWegPortfolio(b *testing.B) {
	naht := gefuellteSammlung(b)

	for b.Loop() {
		_, ohnePreis, err := naht.Portfolio(context.Background())
		if err != nil {
			b.Fatalf("das Portfolio rechnen: %v", err)
		}

		// **Ohne Preise sind alle ohne Preis.** Diese Vorrichtung holt nichts
		// aus dem Netz (§16); was gemessen wird, ist das Zusammenlegen von
		// fünftausend Zeilen, nicht die Preisquelle.
		//
		// **Und gezählt sind Stücke, nicht Zeilen** — `Portfoliozeile.Exemplare`
		// sagt es ausdrücklich. Der erste Anlauf erwartete hier 5.000 und bekam
		// 12.500: die Mengen 1 bis 4 summieren sich. Die Zahl steht deshalb
		// gerechnet da, statt als zweite Konstante daneben.
		if ohnePreis != stueckzahl {
			b.Fatalf("%d Stücke ohne Preis, erwartet %d", ohnePreis, stueckzahl)
		}
	}

	haelt(b, "die Sammlung zusammenrechnen", schwellePortfolio)
}

// gefuellteSammlung baut einen Bestand mit `bestandsgroesse` Exemplaren.
//
// **Ein eigenes Zuhause und eine eigene Datei** (§3). `SammlungsnahtUeber`
// legt das Scryfall-Verzeichnis unter `os.UserConfigDir()` an; ohne
// `b.Setenv("HOME", …)` schriebe dieser Lauf in den **echten** Bestand dessen,
// der misst. Gefunden am 2026-08-24 in
// #888, und
// am 2026-08-30 in dieser Phase noch einmal, als drei Probepakete im echten
// Paketverzeichnis lagen.
func gefuellteSammlung(b *testing.B) *Sammlungsnaht {
	b.Helper()

	b.Setenv("HOME", b.TempDir())

	db, err := storage.Oeffnen(context.Background(),
		filepath.Join(b.TempDir(), "bestand.db"))
	if err != nil {
		b.Fatalf("den Bestand öffnen: %v", err)
	}

	b.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			b.Errorf("den Bestand schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(context.Background(), db, nil); fehler != nil {
		b.Fatalf("migrieren: %v", fehler)
	}

	// **Über den Speicher und nicht über die Naht.** Die Naht prüft jedes
	// Exemplar einzeln und wäre hier die Messung des Füllens statt des Lesens;
	// gemessen werden soll, was ein **vorhandener** Bestand kostet.
	speicher := storage.NeuerSammlungsspeicher(db)

	for zaehler := range bestandsgroesse {
		exemplar := &domain.Exemplar{
			Printing: druckkennung(zaehler),
			Menge:    1 + zaehler%4,
			Zustand:  domain.ZustandNearMint,
			Sprache:  domain.Vorgabesprache,
			Finish:   domain.FinishAus(zaehler%7 == 0),
			Name:     fmt.Sprintf("Probekarte %d", zaehler),
			Setcode:  fmt.Sprintf("p%02d", zaehler%50),
			Nummer:   fmt.Sprintf("%d", zaehler%250),
		}

		if _, fehler := speicher.Anlegen(context.Background(), exemplar); fehler != nil {
			b.Fatalf("Exemplar %d anlegen: %v", zaehler, fehler)
		}
	}

	naht, err := SammlungsnahtUeber(db)
	if err != nil {
		b.Fatalf("die Naht bauen: %v", err)
	}

	return naht
}

// druckkennung gibt eine Kennung in der Gestalt, die Scryfall benutzt.
//
// **Eine UUID-Form und keine Zahl.** Die Spalte trägt Kennungen von 36
// Zeichen; über kurze Zeichenketten zu messen sagt nichts über den Vergleich,
// den die Abfrage wirklich anstellt.
func druckkennung(zaehler int) string {
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", zaehler)
}
