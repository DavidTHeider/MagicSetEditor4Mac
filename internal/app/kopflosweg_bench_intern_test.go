//go:build fremdpakete

// **Der Weg der Flächen ohne Fenster** — eine Karte und ein Ausdruck, wie
// `mse-mcp`, `mse-rest` und `mse` sie rufen
// (#2005).
//
// # Warum eine eigene Messung
//
// `messwege_bench_intern_test.go` misst den Weg des **Fensters**: eine
// [App], die ihr Zeichenwerk einmal baut. Die Flächen ohne Fenster riefen
// [Zeichennaht] je Ruf — und der Leistungsvergleich (QA-04) fand dort 1,8 s
// je Karte gegen 43–61 ms im Fenster. Diese Messung trennt den Ruf in seine
// Teile, damit die Ursache **gemessen** und nicht vermutet ist.
//
// # Unter welchen Bedingungen
//
// **Wie im Leistungsvergleich**: ein eigenes Zuhause, dessen
// Paketverzeichnis das Pack als APFS-Klon trägt — Pakete **und** `Magic -
// Fonts` —, und ein eigener Bestand darin. Der Klon kostet einige Sekunden
// und liegt außerhalb jeder Messung.
package app

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kopfloserLauf ist ein Bestand mit einem eingefahrenen Set über dem Pack.
type kopfloserLauf struct {
	ctx    context.Context
	db     *sql.DB
	karte  int64
	ziel   string
	setID  int64
	karten int
}

// kopflosAufbauen legt Zuhause, Pakete und Bestand an und fährt `satz` ein.
func kopflosAufbauen(b *testing.B, satz string) kopfloserLauf {
	b.Helper()

	quelle := os.Getenv("MSE_FREMDPAKETE")
	if quelle == "" {
		b.Skip("MSE_FREMDPAKETE nicht gesetzt — ohne Stilpaket misst der Rasterer den Behelfsplan")
	}

	setpfad := kanonsatz(b, satz)

	zuhause := kopflosZuhause(b, quelle)
	b.Setenv("HOME", zuhause)
	b.Setenv(Bestandsumgebung, "")

	ort, err := Datenbankpfad()
	if err != nil {
		b.Fatalf("Bestandsort: %v", err)
	}

	// **Jeder Lauf mit eigenem Bestand**, auch in einem geteilten Zuhause.
	for _, rest := range []string{"", "-wal", "-shm"} {
		_ = os.Remove(ort + rest) //nolint:errcheck // meist gibt es ihn nicht
	}

	ctx := storage.MitProfil(context.Background(), 1)

	db, err := storage.Oeffnen(ctx, ort)
	if err != nil {
		b.Fatalf("öffnen: %v", err)
	}

	b.Cleanup(func() { _ = db.Close() })

	if err = storage.Migriere(ctx, db, nil); err != nil {
		b.Fatalf("migrieren: %v", err)
	}

	eingefahren, err := Einfuhrnaht(db).Set(ctx, setpfad)
	if err != nil {
		b.Fatalf("einfahren: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(ctx, eingefahren.SetID)
	if err != nil || len(karten) == 0 {
		b.Fatalf("Karten: %v (%d)", err, len(karten))
	}

	return kopfloserLauf{
		ctx: ctx, db: db, karte: karten[0].ID, setID: eingefahren.SetID,
		karten: len(karten), ziel: filepath.Join(b.TempDir(), "karte.png"),
	}
}

// schwelleKopflosKarte gilt für eine Karte über die Naht der Flächen ohne
// Fenster — also für den Ruf, den `schreiben.karte.zeichnen` macht.
//
// **Gemessen am 2026-10-02** (docs/LEISTUNGSVERGLEICH.md §8): vor #2005
// 1,8 s je Ruf im Leistungsvergleich, weil jeder ein frisches Werk baute; mit
// dem behaltenen Werk 65 ms bei Last 12. Die Schwelle fängt den Rückfall auf
// ein Werk je Ruf mit weitem Abstand und lässt einem belasteten Rechner Luft —
// eine Verdopplung des warmen Wegs fängt `BenchmarkWegRastern`.
const schwelleKopflosKarte = 250 * time.Millisecond

// schwelleKopflosSkript gilt für einen Ausdruck über `lesen.skript`.
//
// **Gemessen am 2026-10-02**: vor #2005 458–482 ms im Leistungsvergleich,
// danach 16 ms bei Last 12.
const schwelleKopflosSkript = 100 * time.Millisecond

// schwelleFrischesWerk gilt für die **erste** Karte über einem frischen Werk —
// die erste einer Sitzung und jede über die Kommandozeile.
//
// **Gemessen am 2026-10-02**: 0.5.0 1,45 s, mit einem Gang über das Pack
// statt dreien 623 ms, beides bei Last 10–12; `b6a1d664` lag in §5 bei 1,87 s
// für die erste Karte einer Sitzung. Die Schwelle fängt die Rückkehr einer
// zweiten Wurzel über demselben Baum.
const schwelleFrischesWerk = 1200 * time.Millisecond

// BenchmarkWegKopflos misst eine Karte und einen Ausdruck, wie eine Fläche
// ohne Fenster sie ruft — und die Teile, aus denen der Ruf besteht.
func BenchmarkWegKopflos(b *testing.B) {
	lauf := kopflosAufbauen(b, "Archenemy.mse-set")
	b.Logf("Archenemy.mse-set mit %d Karten, Karte %d", lauf.karten, lauf.karte)

	// **Der Ruf, wie `schreiben.karte.zeichnen` ihn macht**: Naht, dann Bild.
	//
	// **Nach einem ersten Ruf** — gemessen wird der zweite und jeder weitere
	// einer Sitzung, die Navigation „warm" aus `docs/LEISTUNGSVERGLEICH.md`
	// §5. Den ersten misst „Karte, frisches Werk".
	b.Run("Karte je Ruf", func(b *testing.B) {
		kopflosZeichnen(b, lauf, kopflosNaht(b, lauf))

		defer rechenzeit(b)()

		for b.Loop() {
			kopflosZeichnen(b, lauf, kopflosNaht(b, lauf))
		}

		haelt(b, "eine Karte über die Naht ohne Fenster", schwelleKopflosKarte)
	})

	// **Der Ruf, wie `lesen.skript` ihn macht** — ebenso nach einem ersten.
	b.Run("Skript je Ruf", func(b *testing.B) {
		kopflosAuswerten(b, lauf)

		defer rechenzeit(b)()

		for b.Loop() {
			kopflosAuswerten(b, lauf)
		}

		haelt(b, "ein Ausdruck über die Naht ohne Fenster", schwelleKopflosSkript)
	})

	// **Ein frisches Werk und seine erste Karte** — wie die Naht es baut,
	// aber an ihrem Vorrat vorbei: die erste Karte einer Sitzung, und jede
	// über die Kommandozeile.
	b.Run("Karte, frisches Werk", func(b *testing.B) {
		defer rechenzeit(b)()

		for b.Loop() {
			kopflosZeichnen(b, lauf, frischesWerk(b, lauf))
		}

		haelt(b, "die erste Karte über einem frischen Werk", schwelleFrischesWerk)
	})

	// **Das Werk gehalten — der Weg des Fensters**, über denselben Bestand.
	b.Run("Karte, Werk gehalten", func(b *testing.B) {
		werk := frischesWerk(b, lauf)
		kopflosZeichnen(b, lauf, werk)

		defer rechenzeit(b)()

		for b.Loop() {
			kopflosZeichnen(b, lauf, werk)
		}
	})

	// **Die Teile eines frischen Werks**, je einmal mit der Uhr.
	b.Run("Teile", func(b *testing.B) {
		for b.Loop() {
			kopflosTeile(b, lauf)
		}
	})
}

// kopflosAuswerten wertet einen Ausdruck aus, wie `lesen.skript` es tut.
func kopflosAuswerten(b *testing.B, lauf kopfloserLauf) {
	b.Helper()

	konsole, err := SkriptnahtUeber(lauf.db)
	if err != nil {
		b.Fatalf("Skriptnaht: %v", err)
	}

	if _, err := konsole.Auswerten(lauf.ctx, lauf.karte, "card.name"); err != nil {
		b.Fatalf("Auswerten: %v", err)
	}
}

// kopflosNaht ruft die Naht der Flächen ohne Fenster.
func kopflosNaht(b *testing.B, lauf kopfloserLauf) *Zeichenwerk {
	b.Helper()

	werk, err := Zeichennaht(lauf.db)
	if err != nil {
		b.Fatalf("Zeichennaht: %v", err)
	}

	return werk
}

// frischesWerk baut ein Werk, wie die Naht es baut — ohne ihren Vorrat.
func frischesWerk(b *testing.B, lauf kopfloserLauf) *Zeichenwerk {
	b.Helper()

	pakete, err := Paketverzeichnis()
	if err != nil {
		b.Fatalf("Paketverzeichnis: %v", err)
	}

	bilder, err := Bilderverzeichnis()
	if err != nil {
		b.Fatalf("Bilderverzeichnis: %v", err)
	}

	return NeuesZeichenwerk(storage.NeuerSetLeser(lauf.db), paketwurzelnFuer(pakete), bilder).
		MitPruefer(AutorisierungAus(lauf.db))
}

// kopflosZeichnen schreibt die Karte des Laufs.
func kopflosZeichnen(b *testing.B, lauf kopfloserLauf, werk *Zeichenwerk) {
	b.Helper()

	if _, err := werk.SchreibeBild(lauf.ctx, lauf.karte, lauf.ziel, render.GueteMittel); err != nil {
		b.Fatalf("SchreibeBild: %v", err)
	}
}

// rechenzeit meldet die Rechenzeit des Prozesses je Durchlauf — **neben** der
// Wandzeit, weil ein geteilter Rechner die Wandzeit streckt und die
// Rechenzeit kaum (Benutzer- und Systemzeit zusammen).
func rechenzeit(b *testing.B) func() {
	b.Helper()

	vorher := verbraucht(b)

	return func() {
		b.ReportMetric(float64((verbraucht(b)-vorher).Milliseconds())/float64(max(b.N, 1)), "cpu-ms/op")
	}
}

// verbraucht gibt Benutzer- und Systemzeit des Prozesses.
func verbraucht(b *testing.B) time.Duration {
	b.Helper()

	var nutzung syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &nutzung); err != nil {
		b.Fatalf("getrusage: %v", err)
	}

	return time.Duration(nutzung.Utime.Nano() + nutzung.Stime.Nano())
}

// kopflosTeile stoppt die Schritte eines Rufs einzeln und schreibt sie ins
// Protokoll.
func kopflosTeile(b *testing.B, lauf kopfloserLauf) {
	b.Helper()

	pakete, err := Paketverzeichnis()
	if err != nil {
		b.Fatalf("Paketverzeichnis: %v", err)
	}

	var wurzeln []string

	wurzelzeit := misszeit(b, func() { wurzeln = paketwurzelnFuer(pakete) })

	var werk *Zeichenwerk

	nahtzeit := misszeit(b, func() { werk = frischesWerk(b, lauf) })

	karte, err := storage.NeuerSetLeser(lauf.db).KarteLesen(lauf.ctx, lauf.karte)
	if err != nil {
		b.Fatalf("KarteLesen: %v", err)
	}

	// **Ein gescheiterter Schritt ist kein schneller Schritt**: jeder Fehler
	// bricht die Messung ab, statt als kurze Zeit in den Bericht zu gehen.
	var schrittfehler error

	planzeit := misszeit(b, func() {
		_, _, schrittfehler = werk.Planbauer.Kartenplan(lauf.ctx, karte)
	})
	if schrittfehler != nil {
		b.Fatalf("Kartenplan kalt: %v", schrittfehler)
	}

	plan, _, err := werk.Planbauer.Kartenplan(lauf.ctx, karte)
	if err != nil {
		b.Fatalf("Kartenplan: %v", err)
	}

	rasterzeit := misszeit(b, func() {
		_, _, schrittfehler = werk.Rasterer.Zeichne(lauf.ctx, plan, 1, render.GueteMittel)
	})
	if schrittfehler != nil {
		b.Fatalf("Zeichnen kalt: %v", schrittfehler)
	}

	warmzeit := misszeit(b, func() {
		_, _, schrittfehler = werk.Rasterer.Zeichne(lauf.ctx, plan, 1, render.GueteMittel)
	})
	if schrittfehler != nil {
		b.Fatalf("Zeichnen warm: %v", schrittfehler)
	}

	b.Logf("Wurzeln %d in %v · Werk bauen %v · Plan kalt %v · Zeichnen kalt %v · Zeichnen warm %v",
		len(wurzeln), wurzelzeit.Round(time.Microsecond), nahtzeit.Round(time.Microsecond),
		planzeit.Round(time.Millisecond), rasterzeit.Round(time.Millisecond),
		warmzeit.Round(time.Millisecond))

	// **Wo die Schriftsuche geht**: eine Schrift, die es nicht gibt, nennt im
	// Fehler jede Wurzel — und jede ist ein Gang über ihren Baum.
	if _, err := werk.Schriften.Suche("gibt-es-nicht-2005", false); err != nil {
		b.Logf("Schriftsuche: %v", err)
	}
}

// kopflosZuhause gibt ein Zuhause mit dem Pack als APFS-Klon — oder das, was
// `MSE_PAKETZUHAUSE` nennt.
//
// **Einmal je Messlauf und nicht je Prozess** (PERF-01, #2168): ein Klon von
// 34.683 Dateien weckt `fseventsd`, gemessen am 2026-10-07 mit 250 bis 280 %
// CPU **während** der gemessenen Schleife des Prozesses, der ihn anlegte.
// `scripts/leistungsschnitte.mjs` klont deshalb einmal vorab und wartet, bis
// es ruhig ist.
func kopflosZuhause(b *testing.B, quelle string) string {
	b.Helper()

	if vorbereitet := os.Getenv("MSE_PAKETZUHAUSE"); vorbereitet != "" {
		return vorbereitet
	}

	zuhause := b.TempDir()

	pack := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "full-magic-pack")
	if err := os.MkdirAll(pack, 0o750); err != nil {
		b.Fatalf("Paketverzeichnis: %v", err)
	}

	// **Ein Klon und kein Verweis**: `filepath.WalkDir` folgt keinem — dieselbe
	// Aufstellung wie im Leistungsvergleich (`scripts/leistungsvergleich.mjs`).
	for ziel, herkunft := range map[string]string{
		"data":          quelle,
		"Magic - Fonts": filepath.Join(filepath.Dir(quelle), "Magic - Fonts"),
	} {
		//nolint:gosec // ein Messlauf hinter `fremdpakete`, Pfade aus der Umgebung dessen, der misst
		if ausgabe, err := exec.CommandContext(context.Background(), "cp", "-cR", herkunft, filepath.Join(pack, ziel)).CombinedOutput(); err != nil {
			b.Fatalf("klonen %s: %v %s", herkunft, err, ausgabe)
		}
	}

	return zuhause
}
