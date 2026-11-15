package storage_test

// **Ein Absturz beim Sichern hinterlässt keine unbrauchbare Sicherung** —
// BK-06a (#2130).
//
// # Der Befund
//
// `Sichern` reservierte den **endgültigen** Namen und ließ `VACUUM INTO`
// hinein schreiben. Endete der Prozess dazwischen — ein Absturz, ein
// Stromausfall, ein `kill` —, stand dort eine leere oder halbe Datei. Der
// nächste Start sah sie als „schon da", prüfte sie, fand sie untauglich und
// hielt an: „lege sie beiseite und starte erneut". Von Hand.
//
// # Der Lauf
//
// Ein **zweiter Prozess** — dieses Testprogramm in einer Helferrolle — sichert
// einen echten Bestand von 32 MB und endet an drei Stellen: nach der
// Reservierung, **während** des Kopierens und vor dem Veröffentlichen. Danach
// gilt: am endgültigen Namen steht nichts oder eine taugliche Sicherung, und
// der nächste Versuch gelingt.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// abbruchschalter sagt dem zweiten Prozess, wo er sichern und wann er enden
// soll.
const abbruchschalter = "MSE_SICHERUNGSABBRUCH"

type abbruchangabe struct {
	Bestand  string `json:"bestand"`
	Ziel     string `json:"ziel"`
	Schritt  string `json:"schritt"`
	Waehrend bool   `json:"waehrend"`
}

// sichereUndBrichAb ist die Helferrolle: sichern und an der genannten Stelle
// enden — mit 3; ohne Abbruch mit 4.
func sichereUndBrichAb(angabe string) {
	var wahl abbruchangabe
	if json.Unmarshal([]byte(angabe), &wahl) != nil {
		os.Exit(2)
	}

	db, err := storage.Oeffnen(context.Background(), wahl.Bestand)
	if err != nil {
		os.Exit(2)
	}

	storage.SicherungsschrittSetzen(func(schritt, pfad string) {
		if schritt != wahl.Schritt {
			return
		}

		if !wahl.Waehrend {
			os.Exit(3)
		}

		// **Während**: sobald die Datei wächst, endet der Prozess — mitten im
		// Kopieren.
		go func() {
			for {
				if stand, fehler := os.Stat(pfad); fehler == nil && stand.Size() > 0 {
					os.Exit(3)
				}

				time.Sleep(100 * time.Microsecond)
			}
		}()
	})

	_ = storage.Sichern(context.Background(), db, wahl.Ziel) //nolint:errcheck // der Ausgang ist der Rückgabewert des Prozesses
	os.Exit(4)
}

// grosserBestand legt einen migrierten Bestand mit 32 MB an — groß genug,
// dass das Kopieren dauert.
func grosserBestand(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "mse.db")

	db, err := storage.Oeffnen(t.Context(), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := storage.Migriere(t.Context(), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	for _, anweisung := range []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (1, 'Groß', 'magic', '2026-10-07', 1)`,
		`INSERT INTO set_quellen (set_id, datei, inhalt) VALUES (1, '/gross.mse-set', randomblob(33554432))`,
	} {
		if _, err := db.ExecContext(t.Context(), anweisung); err != nil {
			t.Fatalf("%s: %v", anweisung, err)
		}
	}

	return pfad
}

func TestEinAbsturzBeimSichernHinterlaesstKeineUnbrauchbareSicherung(t *testing.T) {
	faelle := map[string]abbruchangabe{
		"nach der Reservierung":   {Schritt: "vor dem Kopieren"},
		"während des Kopierens":   {Schritt: "vor dem Kopieren", Waehrend: true},
		"vor dem Veröffentlichen": {Schritt: "vor dem Veröffentlichen"},
	}

	for name, wahl := range faelle {
		t.Run(name, func(t *testing.T) {
			wahl.Bestand = grosserBestand(t)
			wahl.Ziel = filepath.Join(t.TempDir(), "vor-migration.sqlite")

			angabe, kodierfehler := json.Marshal(wahl)
			if kodierfehler != nil {
				t.Fatal(kodierfehler)
			}

			helfer := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^$") // #nosec G204,G702 -- der Testlauf selbst
			helfer.Env = append(os.Environ(), abbruchschalter+"="+string(angabe))

			var ende *exec.ExitError
			if err := helfer.Run(); !errors.As(err, &ende) || ende.ExitCode() != 3 {
				t.Fatalf("der Helfer brach nicht an der Stelle ab: %v", err)
			}

			// **Am endgültigen Namen steht nichts oder eine taugliche
			// Sicherung** — nie etwas, das der nächste Start abweist.
			if _, fehlt := os.Stat(wahl.Ziel); fehlt == nil {
				if _, untauglich := storage.PruefeSicherung(t.Context(), wahl.Ziel); untauglich != nil {
					t.Fatalf("nach dem Absturz steht eine untaugliche Sicherung da: %v", untauglich)
				}

				return
			}

			// **Und der nächste Versuch gelingt.**
			db, oeffnungsfehler := storage.Oeffnen(t.Context(), wahl.Bestand)
			if oeffnungsfehler != nil {
				t.Fatal(oeffnungsfehler)
			}
			defer func() { _ = db.Close() }()

			if err := storage.Sichern(t.Context(), db, wahl.Ziel); err != nil {
				t.Fatalf("der nächste Versuch: %v", err)
			}

			if _, untauglich := storage.PruefeSicherung(t.Context(), wahl.Ziel); untauglich != nil {
				t.Errorf("die neue Sicherung taugt nicht: %v", untauglich)
			}
		})
	}
}

// **Eine fertige Sicherung bleibt Byte für Byte, wie sie ist** — auch wenn
// sie erst entsteht, während eine zweite gerade kopiert.
func TestEineZweiteSicherungUeberschreibtDieErsteNicht(t *testing.T) {
	pfad := grosserBestand(t)
	ziel := filepath.Join(t.TempDir(), "vor-migration.sqlite")

	db, err := storage.Oeffnen(t.Context(), pfad)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	vorher := []byte("die erste, die schneller war")

	storage.SicherungsschrittSetzen(func(schritt, _ string) {
		if schritt == "vor dem Veröffentlichen" {
			if err := os.WriteFile(ziel, vorher, 0o600); err != nil {
				t.Errorf("die erste Sicherung: %v", err)
			}
		}
	})
	t.Cleanup(func() { storage.SicherungsschrittSetzen(func(_, _ string) {}) })

	if err := storage.Sichern(t.Context(), db, ziel); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("die zweite Sicherung meldet %v statt eines Konflikts", err)
	}

	// #nosec G304 -- der Pfad liegt im Prüfverzeichnis.
	if danach, err := os.ReadFile(ziel); err != nil || !bytes.Equal(danach, vorher) {
		t.Errorf("die erste Sicherung ist nicht mehr dieselbe: %q (%v)", danach, err)
	}

	// **Auch die versteckte Nachbardatei ist fort** — `ReadDir` sieht sie,
	// ein `*` nicht.
	if reste, err := os.ReadDir(filepath.Dir(ziel)); err != nil || len(reste) != 1 {
		t.Errorf("neben der Sicherung liegt noch etwas: %v (%v)", reste, err)
	}
}
