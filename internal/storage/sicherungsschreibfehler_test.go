//go:build darwin || linux

package storage_test

// **Ein Schreibfehler beim Sichern lässt nichts liegen und heißt, was er ist**
// — M02 (#2346),
// Befund A-12 des Audits 0.9.0.
//
// # Der Befund
//
// Gefunden von der Hüllenprobe `wanderung09` am 2026-10-14: auf einem vollen
// Datenträger scheiterte die Sicherung vor der Wanderung, `Sichern` räumte
// die reservierte Datei — und das Journal, das `VACUUM INTO` neben ihr
// angelegt hatte, blieb liegen (`.bestand-vor-0.8.0.db.entsteht-…-journal`).
// An der Kommandozeile stand dazu der Treibertext, `sqlite3: disk I/O error`.
//
// # Wie der Schreibfehler entsteht, ohne einen Datenträger zu füllen
//
// **Ein zweiter Prozess mit einer Grenze für die Dateigröße** (`RLIMIT_FSIZE`)
// und ohne das Signal, das sie sonst schickt: jedes Schreiben über die Grenze
// scheitert dann mit `EFBIG` — wie auf einem vollen Datenträger mitten im
// Kopieren. Den Bestand legt dieser Prozess an, bevor die Grenze gilt.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// grenzschalter sagt dem zweiten Prozess, wo er unter der Grenze sichern soll.
const grenzschalter = "MSE_SICHERUNGSGRENZE"

type grenzangabe struct {
	Bestand string `json:"bestand"`
	Ziel    string `json:"ziel"`
	Grenze  uint64 `json:"grenze"`
}

// grenzergebnis ist, was der zweite Prozess über den Fehler sagt.
type grenzergebnis struct {
	Fehler string `json:"fehler"`
	EA     bool   `json:"ea"`
	Voll   bool   `json:"voll"`
}

// sichereUnterGrenze ist die Helferrolle: sichern, während keine Datei über
// die Grenze wachsen darf, und den Fehler als JSON nennen.
func sichereUnterGrenze(angabe string) {
	var wahl grenzangabe
	if json.Unmarshal([]byte(angabe), &wahl) != nil {
		os.Exit(2)
	}

	db, err := storage.Oeffnen(context.Background(), wahl.Bestand)
	if err != nil {
		os.Exit(2)
	}

	var grenze syscall.Rlimit
	if syscall.Getrlimit(syscall.RLIMIT_FSIZE, &grenze) != nil {
		os.Exit(2)
	}

	signal.Ignore(syscall.SIGXFSZ)

	grenze.Cur = wahl.Grenze
	if syscall.Setrlimit(syscall.RLIMIT_FSIZE, &grenze) != nil {
		os.Exit(2)
	}

	fehler := storage.Sichern(context.Background(), db, wahl.Ziel)

	ergebnis := grenzergebnis{EA: errors.Is(fehler, domain.ErrEA), Voll: errors.Is(fehler, domain.ErrVoll)}
	if fehler != nil {
		ergebnis.Fehler = fehler.Error()
	}

	roh, _ := json.Marshal(ergebnis) //nolint:errcheck // drei einfache Felder

	_, _ = fmt.Fprintln(os.Stdout, string(roh)) //nolint:errcheck // der Ausgang ist die Ausgabe
	os.Exit(0)
}

func TestEinSchreibfehlerBeimSichernLaesstNichtsLiegen(t *testing.T) {
	ablage := t.TempDir()
	angabe, kodierfehler := json.Marshal(grenzangabe{
		Bestand: grosserBestand(t), Ziel: filepath.Join(ablage, "bestand-vor-0.8.0.db"), Grenze: 1 << 20,
	})

	if kodierfehler != nil {
		t.Fatal(kodierfehler)
	}

	helfer := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^$") // #nosec G204,G702 -- der Testlauf selbst
	helfer.Env = append(os.Environ(), grenzschalter+"="+string(angabe))

	ausgabe, laufehler := helfer.Output()
	if laufehler != nil {
		t.Fatalf("der Helfer: %v", laufehler)
	}

	var ergebnis grenzergebnis
	if fehler := json.Unmarshal(ausgabe, &ergebnis); fehler != nil {
		t.Fatalf("die Antwort des Helfers %q: %v", ausgabe, fehler)
	}

	if ergebnis.Fehler == "" {
		t.Fatal("die Sicherung gelang über die Grenze hinaus")
	}

	// **Und heißt, was er ist** — ein Fehler des Datenträgers, kein
	// Treibertext.
	if !ergebnis.EA && !ergebnis.Voll {
		t.Errorf("der Fehler ist nicht eingeordnet: %s", ergebnis.Fehler)
	}

	eintraege, lesefehler := os.ReadDir(ablage)
	if lesefehler != nil {
		t.Fatal(lesefehler)
	}

	for _, eintrag := range eintraege {
		t.Errorf("liegen geblieben: %s", eintrag.Name())
	}
}
