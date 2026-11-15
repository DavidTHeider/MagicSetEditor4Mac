package service

// **Was eine Datei auf der Platte zugesagt bekommt, hängt an ihrer Rolle** —
// BK-06b und SEC-20
// (#2131,
// #2133,
// Vermerk 0174).
//
// - Jeder Fehler **vor** dem Veröffentlichen — Schreiben, Festschreiben,
//   Schließen, Umbenennen — lässt die alte Datei Byte für Byte stehen und
//   keine Nachbardatei zurück.
// - Eine Ausgabe wird mit `fsync` festgeschrieben, eine Marke für den
//   Wiederanlauf mit `F_FULLFSYNC`, ein Zwischenspeicher gar nicht — je Datei
//   **und** Verzeichnis.
// - Eine vertrauliche Ausgabe endet mit 0600, auch über einer Datei mit 0644;
//   eine gewöhnliche behält den Modus der Datei, die sie ersetzt.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var errEingespielt = errors.New("planmäßig eingespielt")

// stolpernachbar ist eine echte Nachbardatei, die an einer Stelle scheitert.
type stolpernachbar struct {
	*os.File
	stelle string
}

func (s stolpernachbar) Write(daten []byte) (int, error) {
	if s.stelle == "schreiben" {
		return 0, errEingespielt
	}

	return s.File.Write(daten)
}

func (s stolpernachbar) Close() error {
	return errors.Join(s.File.Close(), map[bool]error{true: errEingespielt}[s.stelle == "schliessen"])
}

// mitStelle setzt die Nähte so, dass an `stelle` gescheitert wird.
func mitStelle(t *testing.T, stelle string) {
	t.Helper()

	anlegen, nennen, leicht := nachbarAnlegen, umbenennen, leichtFestschreiben
	t.Cleanup(func() { nachbarAnlegen, umbenennen, leichtFestschreiben = anlegen, nennen, leicht })

	nachbarAnlegen = func(verzeichnis, muster string) (nachbardatei, error) {
		datei, err := os.CreateTemp(verzeichnis, muster)
		if err != nil {
			return nil, err
		}

		return stolpernachbar{File: datei, stelle: stelle}, nil
	}

	umbenennen = func(alt, neu string) error {
		if stelle == "umbenennen" {
			return errEingespielt
		}

		return os.Rename(alt, neu)
	}

	leichtFestschreiben = func(d nachbardatei) error {
		if stelle == "festschreiben" {
			return errEingespielt
		}

		return leicht(d)
	}
}

func TestJederFehlerVorDemVeroeffentlichenLaesstDieAlteDateiStehen(t *testing.T) {
	for _, stelle := range []string{"schreiben", "festschreiben", "schliessen", "umbenennen"} {
		t.Run(stelle, func(t *testing.T) {
			ordner := t.TempDir()
			ziel := filepath.Join(ordner, "set.mse-set")
			alt := []byte("der alte Stand")

			if err := os.WriteFile(ziel, alt, 0o600); err != nil {
				t.Fatal(err)
			}

			mitStelle(t, stelle)

			if err := schreibeAtomar(ziel, []byte("der neue Stand")); !errors.Is(err, errEingespielt) {
				t.Fatalf("erwartet war der eingespielte Fehler, gekommen ist %v", err)
			}

			// #nosec G304 -- das Prüfverzeichnis.
			if danach, err := os.ReadFile(ziel); err != nil || !bytes.Equal(danach, alt) {
				t.Errorf("die alte Datei ist nicht mehr dieselbe: %q (%v)", danach, err)
			}

			if reste, err := os.ReadDir(ordner); err != nil || len(reste) != 1 {
				t.Errorf("neben der Datei liegt noch etwas: %v (%v)", reste, err)
			}
		})
	}
}

// festschreibungen zeichnet auf, was festgeschrieben wird — `leicht` für
// `fsync`, `voll` für `F_FULLFSYNC`.
func festschreibungen(t *testing.T) *[]string {
	t.Helper()

	leicht, voll := leichtFestschreiben, vollFestschreiben
	t.Cleanup(func() { leichtFestschreiben, vollFestschreiben = leicht, voll })

	gesehen := []string{}
	leichtFestschreiben = func(d nachbardatei) error {
		gesehen = append(gesehen, "leicht "+art(d))

		return leicht(d)
	}
	vollFestschreiben = func(d nachbardatei) error {
		gesehen = append(gesehen, "voll "+art(d))

		return voll(d)
	}

	return &gesehen
}

func art(d nachbardatei) string {
	if angabe, err := os.Stat(d.Name()); err == nil && angabe.IsDir() {
		return "Verzeichnis"
	}

	return "Datei"
}

func TestDieRolleBestimmtWieWeitFestgeschriebenWird(t *testing.T) {
	faelle := map[string]struct {
		schreibe func(pfad string) error
		erwartet []string
	}{
		"Ausgabe": {
			schreibe: func(pfad string) error { return schreibeAtomar(pfad, []byte("x")) },
			erwartet: []string{"leicht Datei", "leicht Verzeichnis"},
		},
		"Marke für den Wiederanlauf": {
			schreibe: func(pfad string) error { return schreibeMarke(pfad, []byte("x")) },
			erwartet: []string{"voll Datei", "voll Verzeichnis"},
		},
		"Zwischenspeicher": {
			schreibe: func(pfad string) error { return schreibeZwischenspeicher(pfad, []byte("x")) },
			erwartet: []string{},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gesehen := festschreibungen(t)

			if err := fall.schreibe(filepath.Join(t.TempDir(), "datei")); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(*gesehen, fall.erwartet) {
				t.Errorf("festgeschrieben wurde %v, erwartet war %v", *gesehen, fall.erwartet)
			}
		})
	}
}

// **Ein Verzeichnis, das sich nicht festschreiben lässt, ist keine
// Erfolgsmeldung** — die Datei steht dann da, aber ihr Name übersteht keinen
// Stromausfall.
func TestEinNichtFestgeschriebenesVerzeichnisWirdGemeldet(t *testing.T) {
	leicht := leichtFestschreiben
	t.Cleanup(func() { leichtFestschreiben = leicht })

	leichtFestschreiben = func(d nachbardatei) error {
		if art(d) == "Verzeichnis" {
			return errEingespielt
		}

		return leicht(d)
	}

	if err := schreibeAtomar(filepath.Join(t.TempDir(), "datei"), []byte("x")); !errors.Is(err, errEingespielt) {
		t.Errorf("erwartet war der eingespielte Fehler, gekommen ist %v", err)
	}
}

func TestEineVertraulicheAusgabeEndetMit0600(t *testing.T) {
	faelle := map[string]struct {
		schreibe func(pfad string, inhalt []byte) error
		vorher   fs.FileMode
		erwartet fs.FileMode
	}{
		"vertraulich über 0644": {SchreibeAtomarVertraulich, 0o644, 0o600},
		"vertraulich über 0660": {SchreibeAtomarVertraulich, 0o660, 0o600},
		"vertraulich, neu":      {SchreibeAtomarVertraulich, 0, 0o600},
		"gewöhnlich über 0644":  {SchreibeAtomar, 0o644, 0o644},
		"gewöhnlich, neu":       {SchreibeAtomar, 0, 0o600},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			ziel := filepath.Join(t.TempDir(), "auskunft.json")

			if fall.vorher != 0 {
				if err := os.WriteFile(ziel, []byte("alt"), 0o600); err != nil {
					t.Fatal(err)
				}

				// **Ausdrücklich gesetzt** — `WriteFile` ginge durch die umask.
				if err := os.Chmod(ziel, fall.vorher); err != nil {
					t.Fatal(err)
				}
			}

			if err := fall.schreibe(ziel, []byte("neu")); err != nil {
				t.Fatal(err)
			}

			angabe, err := os.Stat(ziel)
			if err != nil {
				t.Fatal(err)
			}

			if modus := angabe.Mode().Perm(); modus != fall.erwartet {
				t.Errorf("Modus %o, erwartet %o", modus, fall.erwartet)
			}
		})
	}
}

func TestEinFehlendesVerzeichnisLaesstSichNichtFestschreiben(t *testing.T) {
	if err := verzeichnisFestschreiben(filepath.Join(t.TempDir(), "fehlt"), alsAusgabe); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("erwartet war ErrNotExist, gekommen ist %v", err)
	}
}

// **Ein Verzeichnis ohne Leserecht nimmt die Datei trotzdem** — es lässt
// sich nur nicht öffnen, um es festzuschreiben. Hier schreiben ging vorher,
// und ein Fehler wäre ein Rückschritt; die Grenze steht im Vermerk 0174.
func TestEinVerzeichnisOhneLeserechtNimmtDieDateiTrotzdem(t *testing.T) {
	ordner := filepath.Join(t.TempDir(), "einwurf")
	if err := os.Mkdir(ordner, 0o300); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(ordner, 0o700); err != nil { //nolint:gosec // sonst räumt t.TempDir ihn nicht weg
			t.Error(err)
		}
	})

	if err := schreibeAtomar(filepath.Join(ordner, "datei"), []byte("x")); err != nil {
		t.Errorf("ins Verzeichnis ohne Leserecht: %v", err)
	}
}
