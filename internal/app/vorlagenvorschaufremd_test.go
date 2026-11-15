//go:build fremdpakete

// **Die Vorlagenvorschau an einem echten Paket** —
// #1164.
//
// Sie beantwortet die eine Frage, die kein Doppel beantworten kann: **kommt
// die ungespeicherte Datei wirklich bis in das gezeichnete Bild?**
//
//	MSE_FREMDPAKETE=~/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/app/ -run VorlageAmEchtenPaket -v
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineVorlageAmEchtenPaket(t *testing.T) {
	quelle := os.Getenv("MSE_FREMDPAKETE")
	if quelle == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	// **Ein eigenes Zuhause mit einer Kopie des Stilpakets.** Der echte
	// Paketordner wird nicht angefasst (§3).
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	wurzel, err := app.Paketverzeichnis()
	if err != nil {
		t.Fatalf("Paketverzeichnis: %v", err)
	}

	for _, name := range []string{"magic-m15.mse-style", "magic.mse-game"} {
		if fehler := os.CopyFS(filepath.Join(wurzel, name),
			os.DirFS(filepath.Join(quelle, name))); fehler != nil {
			t.Fatalf("%s kopieren: %v", name, fehler)
		}
	}

	// **Der Bestand liegt dort, wo die Anwendung ihn sucht.** `Starten`
	// öffnet den Pfad aus dem Zuhause; ein Bestand daneben wäre ein zweiter,
	// und das Set stünde im falschen.
	pfad, pfadfehler := app.Datenbankpfad()
	if pfadfehler != nil {
		t.Fatalf("Datenbankpfad: %v", pfadfehler)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	db, dbfehler := storage.Oeffnen(probekontext(t), pfad)
	if dbfehler != nil {
		t.Fatalf("öffnen: %v", dbfehler)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	setID, setfehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "magic", Style: "m15"}, nil)
	if setfehler != nil {
		t.Fatalf("Set anlegen: %v", setfehler)
	}

	if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	anwendung := app.NeueApp(pfad)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("Starten: %v", fehler)
	}

	paket := filepath.Join(wurzel, "magic-m15.mse-style")

	roh, lesefehler := os.ReadFile(filepath.Join(paket, "style")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("style lesen: %v", lesefehler)
	}

	vorher := anwendung.Vorlagenvorschau(t.Context(), paket, "style", string(roh), 1, "")
	if vorher.Fehler != nil {
		t.Fatalf("%+v", vorher.Fehler)
	}

	// **Ein anderes Kartenmaß ist der sichtbarste Beleg.** Es steht als
	// `card width:` im Kopf der Datei und geht durch den ganzen Weg — Zerteiler,
	// Stilmodell, Planbauer, Rasterer.
	schmal := strings.Replace(string(roh), "card width: 375", "card width: 200", 1)
	if schmal == string(roh) {
		t.Fatal("die Zeile `card width: 375` steht nicht in dieser Vorlage")
	}

	nachher := anwendung.Vorlagenvorschau(t.Context(), paket, "style", schmal, 1, "")
	if nachher.Fehler != nil {
		t.Fatalf("%+v", nachher.Fehler)
	}

	if vorher.Vorschau.Breite == nachher.Vorschau.Breite {
		t.Errorf("beide Male %d Punkte breit — der Ersatz kam nicht an",
			nachher.Vorschau.Breite)
	}

	if nachher.Vorschau.Breite != 200 {
		t.Errorf("die Karte ist %d Punkte breit — erwartet waren 200",
			nachher.Vorschau.Breite)
	}
}
