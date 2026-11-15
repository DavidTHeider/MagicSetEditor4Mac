//go:build fremdpakete

// **Wie lange die Kartenliste braucht, und wie groß ihre Antwort ist** —
// #891,
// Werkstücke 5 und 6 von Phase Y.
//
// # Warum hinter einem Build-Tag
//
// Er braucht ein **echtes** Set — `Black.mse-set` hat 3.788 Karten —, und die
// kanonischen Sets liegen nicht im Repositorium (§11). Und er dauert: das
// Einfahren allein ist Sekunden. In der Pipeline hätte er nichts zu suchen
// (§4); auf die Liste in `docs/RELEASE.md` gehört er.
//
// # Wie er gefahren wird
//
//	MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets \
//	  go test -tags fremdpakete ./internal/app/ -run XXX -bench Kartenliste -benchtime 3x -v
//
// **`-run XXX`**, damit nur der Benchmark läuft und nicht die Prüfläufe daneben.
package app_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// BenchmarkKartenliste misst den Weg, den das Fenster beim Öffnen eines Sets
// geht.
//
// **Gemessen wird beides**: die Zeit **und** die Größe der Antwort, wie sie
// über die Wails-Naht geht. Die zweite Zahl ist die, um die es geht — 4,7 MB
// JSON je Aufruf sind kein Rechenproblem, sondern ein Übertragungsproblem.
func BenchmarkKartenliste(b *testing.B) {
	kanon := os.Getenv("MSE_KANON")
	if kanon == "" {
		b.Skip("MSE_KANON nicht gesetzt — die kanonischen Sets liegen neben dem Repositorium (§11)")
	}

	// **Der Pfad kommt aus der Umgebung dessen, der misst** — er nennt das
	// Verzeichnis der kanonischen Sets, die neben dem Repositorium liegen
	// (§11). Ein Messlauf hinter `fremdpakete` läuft nie in der Pipeline und
	// nie auf fremde Eingabe.
	quelle := filepath.Join(kanon, "Black.mse-set")
	if _, err := os.Stat(quelle); err != nil { // #nosec G703 -- siehe darüber
		b.Skipf("%s: %v", quelle, err)
	}

	// **Ein eigenes Zuhause** (§3): `Starten` legt sein Beiwerk unter
	// `os.UserConfigDir()` an, und das ist auf darwin `$HOME/Library/…`. Ohne
	// diese Zeile schriebe ein Messlauf in den echten Bestand.
	b.Setenv("HOME", b.TempDir())

	anwendung := app.NeueApp(filepath.Join(b.TempDir(), "bestand.db"))
	if err := anwendung.Starten(context.Background()); err != nil {
		b.Fatalf("starten: %v", err)
	}

	b.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			b.Errorf("beenden: %v", fehler)
		}
	})

	// **Das Einfahren steht außerhalb der Messung.** Gemessen wird, was das
	// Fenster **bei jedem Öffnen** tut, nicht was es einmal tut.
	geoeffnet := anwendung.SetOeffnen(context.Background(), quelle)
	if geoeffnet.Fehler != nil {
		b.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	b.Logf("Set %s mit %d Karten", geoeffnet.SetID, geoeffnet.Karten)

	// **Einmal vorab, um die Größe zu nennen.** Sie gehört in die Ausgabe:
	// eine Zeitmessung ohne sie sagt nicht, woran es liegt.
	antwort := anwendung.Karten(context.Background(), geoeffnet.SetID)
	if antwort.Fehler != nil {
		b.Fatalf("Karten: %+v", antwort.Fehler)
	}

	roh, err := json.Marshal(antwort)
	if err != nil {
		b.Fatalf("umsetzen: %v", err)
	}

	b.Logf("alles: %d Karten, %d Bytes JSON", len(antwort.Karten), len(roh))

	// **Und dieselbe Frage über die Liste** (#891).
	zeilen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
	if zeilen.Fehler != nil {
		b.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	schlank, err := json.Marshal(zeilen)
	if err != nil {
		b.Fatalf("umsetzen: %v", err)
	}

	b.Logf("Liste: %d Zeilen, %d Bytes JSON — %.1f-mal kleiner",
		len(zeilen.Zeilen), len(schlank), float64(len(roh))/float64(len(schlank)))

	b.Run("alles", func(b *testing.B) {
		for b.Loop() {
			gelesen := anwendung.Karten(context.Background(), geoeffnet.SetID)
			if gelesen.Fehler != nil {
				b.Fatalf("Karten: %+v", gelesen.Fehler)
			}
		}
	})

	b.Run("liste", func(b *testing.B) {
		for b.Loop() {
			gelesen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
			if gelesen.Fehler != nil {
				b.Fatalf("Kartenzeilen: %+v", gelesen.Fehler)
			}
		}
	})
}
