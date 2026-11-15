// **Das Zeichenwerk für eine Fläche ohne Fenster** (#886).
//
// Geprüft wird, was `internal/app` beiträgt: die Zusammensetzung steht, die
// Endung bestimmt das Format, und eine unbekannte Endung ist eine
// **Validierung** — damit eine Fläche sie von „die Karte gibt es nicht"
// unterscheiden kann (§8).
//
// **Was hier nicht geprüft wird**: dass die Karte richtig aussieht. Das prüft
// `internal/render` mit Goldbildern, und `cmd/mse` fährt den ganzen Weg bis zu
// den Bytes.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieZeichennahtStehtVollstaendig(t *testing.T) {
	// **Ein eigenes Zuhause, und ein Paket darin** (§3).
	//
	// Bis zum 2026-08-25 stand hier keins, und die Zusicherung darunter — „das
	// Zeichenwerk hat Paketwurzeln" — war eine Aussage über **diesen**
	// Rechner: `Paketverzeichnis()` zeigt in das Benutzerverzeichnis, und dort
	// liegt hier `M15-Magic-Pack`. Mit `HOME` auf ein leeres Verzeichnis
	// gesetzt war der Lauf **rot** — auf einem frischen Klon also immer.
	//
	// Jetzt legt der Test das Paket selbst hin. Damit prüft die Zusicherung
	// die **Verdrahtung** und nicht mehr den Stand der Maschine.
	paketMitEigenemZuhause(t)

	werk, err := app.Zeichennaht(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("die Zeichennaht: %v", err)
	}

	// **Jedes Teil einzeln**, denn ein fehlendes fiele sonst erst beim
	// Zeichnen auf — und dann als Absturz.
	teile := map[string]bool{
		"Bildquelle":  werk.Bildquelle == nil,
		"Schriften":   werk.Schriften == nil,
		"Stile":       werk.Stile == nil,
		"Planbauer":   werk.Planbauer == nil,
		"Rasterer":    werk.Rasterer == nil,
		"Bildausfuhr": werk.Bildausfuhr == nil,
		"Bogenleger":  werk.Bogenleger == nil,
		"Vorlagen":    werk.Vorlagen == nil,
	}

	for name, fehlt := range teile {
		if fehlt {
			t.Errorf("%s fehlt im Zeichenwerk", name)
		}
	}

	if len(werk.Wurzeln) == 0 {
		t.Error("das Zeichenwerk hat keine Paketwurzeln")
	}
}

// **Eine unbekannte Endung ist eine Validierung.**
func TestSchreibeBildWeistEineUnbekannteEndungAb(t *testing.T) {
	werk, err := app.Zeichennaht(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("die Zeichennaht: %v", err)
	}

	_, fehler := werk.SchreibeBild(probekontext(t), 1,
		filepath.Join(t.TempDir(), "karte.xyz"), render.GueteMittel)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war ErrValidation", fehler)
	}

	// **Die Meldung nennt die Datei und die Endungen, die gehen.** Eine
	// Abweisung ohne Ausweg ist eine Sackgasse.
	if !contains(fehler.Error(), "karte.xyz") || !contains(fehler.Error(), ".png") {
		t.Errorf("die Meldung hilft nicht weiter: %v", fehler)
	}
}

// **Eine Karte, die es nicht gibt, ist kein Bedienfehler.**
//
// Der Unterschied zählt: ein Skript soll „mein Aufruf stimmt nicht" von „die
// Antwort ist nein" unterscheiden können.
func TestSchreibeBildEinerUnbekanntenKarteIstKeineValidierung(t *testing.T) {
	werk, err := app.Zeichennaht(offeneProbedatenbank(t))
	if err != nil {
		t.Fatalf("die Zeichennaht: %v", err)
	}

	_, fehler := werk.SchreibeBild(probekontext(t), 4711,
		filepath.Join(t.TempDir(), "karte.png"), render.GueteMittel)
	if fehler == nil {
		t.Fatal("die unbekannte Karte wurde gezeichnet")
	}

	if errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die fehlende Karte kam als Validierung: %v", fehler)
	}
}

// contains ist strings.Contains — hier ohne den Import.
func contains(heuhaufen, nadel string) bool {
	for i := 0; i+len(nadel) <= len(heuhaufen); i++ {
		if heuhaufen[i:i+len(nadel)] == nadel {
			return true
		}
	}

	return false
}

// **Ohne Benutzerverzeichnis kommt die Zeichennaht nicht zustande.**
//
// Sie braucht Paket- und Bilderort; beide kommen von dort. Ein stiller
// Weiterlauf gäbe ein Zeichenwerk mit leeren Wurzeln — und die Karte sähe leer
// aus, ohne dass jemand den Grund erführe.
func TestOhneBenutzerverzeichnisGibtEsKeineZeichennaht(t *testing.T) {
	db := offeneProbedatenbank(t)

	t.Setenv("HOME", "")

	if _, err := app.Zeichennaht(db); err == nil {
		t.Error("die Zeichennaht kam ohne Benutzerverzeichnis zustande")
	}
}

// **Und der ganze Weg trägt: aus einer Karte wird eine Bilddatei.**
//
// # Warum das Zeichenwerk hier von Hand zusammengesetzt wird
//
// [app.Zeichennaht] nimmt die Paketwurzeln des Rechners — in einem Test wäre
// das der Bestand der Nutzenden, und das Ergebnis hinge davon ab, was dort
// installiert ist. Hier steht die Saat des Repositoriums als Wurzel; sie ist
// selbst geschrieben (§11) und liegt in jedem Klon.
func TestSchreibeBildZeichnetEineKarte(t *testing.T) {
	db := offeneProbedatenbank(t)

	saat, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	ergebnis, fehler := app.Einfuhrnaht(db).Set(probekontext(t),
		filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if fehler != nil {
		t.Fatalf("einfahren: %v", fehler)
	}

	karten, fehler := storage.NeuerSetLeser(db).Karten(probekontext(t), ergebnis.SetID)
	if fehler != nil || len(karten) == 0 {
		t.Fatalf("Karten lesen: %v (%d)", fehler, len(karten))
	}

	werk := app.NeuesZeichenwerk(storage.NeuerSetLeser(db), []string{saat}, t.TempDir())

	// **Ohne Verzeichnis im Ziel** — der Zweig, der `.` einsetzt, ist sonst
	// unbetreten, und ein Aufruf mit blankem Dateinamen ist der Normalfall
	// einer Kommandozeile.
	vorher, err := os.Getwd()
	if err != nil {
		t.Fatalf("Arbeitsverzeichnis: %v", err)
	}

	ort := t.TempDir()
	if err := os.Chdir(ort); err != nil {
		t.Fatalf("wechseln: %v", err)
	}

	t.Cleanup(func() { _ = os.Chdir(vorher) }) //nolint:errcheck // nur zum Aufräumen

	geschrieben, fehler := werk.SchreibeBild(probekontext(t), karten[0].ID, "karte.png",
		render.GueteMittel)
	if fehler != nil {
		t.Fatalf("zeichnen: %v", fehler)
	}

	if geschrieben.Breite == 0 || geschrieben.Hoehe == 0 {
		t.Errorf("das Bild ist %d × %d", geschrieben.Breite, geschrieben.Hoehe)
	}

	if _, err := os.Stat(filepath.Join(ort, "karte.png")); err != nil {
		t.Errorf("die Bilddatei fehlt: %v", err)
	}
}

// paketMitEigenemZuhause gibt dem Lauf ein leeres Zuhause und legt ein Paket
// dorthin, wo `Paketverzeichnis()` es sucht.
//
// **Ein Verzeichnis mit einer `.mse-*`-Endung genügt** — `traegtPakete` fragt
// nach nichts weiter (`internal/msepaket/paketwurzeln.go`). Was darin steht,
// prüft diese Naht nicht; sie sagt, **wo** gesucht wird.
func paketMitEigenemZuhause(t *testing.T) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	pakete := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "probe.mse-game")

	if err := os.MkdirAll(pakete, 0o750); err != nil {
		t.Fatalf("Paket anlegen: %v", err)
	}
}
