package app

// **Die Fehlerwege des Umzugs** — R02
// (#2334).
// Jeder wird gemeldet, und keiner hält die übrigen Schritte auf.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Ein Testlauf zieht im echten Zuhause nichts um** — und ohne Zuhause gar
// nichts.
func TestKennungUmziehenNurImEigenenZuhause(t *testing.T) {
	t.Setenv("HOME", "")

	if err := KennungUmziehen(t.Context()); err == nil {
		t.Error("ohne Zuhause umgezogen")
	}

	t.Setenv("HOME", "/Users/niemand")

	if err := KennungUmziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "Testlauf") {
		t.Errorf("im echten Zuhause: %v", err)
	}

	heim := t.TempDir()
	t.Setenv("HOME", heim)
	legeAn(t, filepath.Join(heim, "Library", "WebKit", altkennungen[0], "x"), "alt")

	if err := KennungUmziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	if inhalt(t, filepath.Join(heim, "Library", "WebKit", Kennung, "x")) != "alt" {
		t.Error("im eigenen Zuhause nichts umgezogen")
	}
}

func TestEinUnkopierbarerWebspeicherWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	alt := filepath.Join(umzug.heim, "Library", "WebKit", altkennungen[0])
	legeAn(t, filepath.Join(alt, "x"), "alt")

	// **Eine unlesbare Datei lässt sich nicht kopieren** — und der halbe Stand
	// bleibt am Zwischenort, nicht unter der neuen Kennung.
	legeAn(t, filepath.Join(alt, "gesperrt"), "geheim")

	if err := os.Chmod(filepath.Join(alt, "gesperrt"), 0o000); err != nil {
		t.Fatal(err)
	}

	if err := umzug.ziehen(t.Context()); err == nil {
		t.Fatal("der Fehler ging verloren")
	}

	if _, err := os.Stat(filepath.Join(umzug.heim, "Library", "WebKit", Kennung)); err == nil {
		t.Error("ein halber Umzug steht unter der neuen Kennung")
	}
}

func TestEinUnraeumbarerZwischenortWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	unter := filepath.Join(umzug.heim, "Library", "WebKit")
	legeAn(t, filepath.Join(unter, altkennungen[0], "x"), "alt")
	legeAn(t, filepath.Join(unter, Kennung+zwischenendung, "halb"), "halb")

	//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
	if err := os.Chmod(filepath.Join(unter, Kennung+zwischenendung), 0o500); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(filepath.Join(unter, Kennung+zwischenendung), 0o700) }) //nolint:errcheck,gosec // nur fürs Aufräumen

	if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "räumen") {
		t.Errorf("Fehler %v", err)
	}
}

func TestScheiterndeVorgabenWerdenGemeldet(t *testing.T) {
	t.Parallel()

	for _, schritt := range []string{"export", "-convert", "import"} {
		t.Run(schritt, func(t *testing.T) {
			t.Parallel()

			umzug, buch := umzugIn(t)
			buch.domaenen = map[string]bool{altkennungen[0]: true}

			ausfuehren := buch.ausfuehren
			umzug.ausfuehren = func(ctx context.Context, befehl []string) error {
				if befehl[1] == schritt {
					return errors.New("abgelehnt")
				}

				return ausfuehren(ctx, befehl)
			}

			if err := umzug.ziehen(t.Context()); err == nil {
				t.Errorf("%s: der Fehler ging verloren", schritt)
			}
		})
	}
}

func TestScheiternderDienstumzugWirdGemeldet(t *testing.T) {
	t.Parallel()

	t.Run("die alte plist ist unlesbar", func(t *testing.T) {
		t.Parallel()

		umzug, _ := umzugIn(t)
		if err := os.MkdirAll(filepath.Join(umzug.heim, "Library", "LaunchAgents", altdienstlabel+".plist"), 0o700); err != nil {
			t.Fatal(err)
		}

		if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "lesen") {
			t.Errorf("Fehler %v", err)
		}
	})

	t.Run("die neue plist lässt sich nicht schreiben", func(t *testing.T) {
		t.Parallel()

		umzug, buch := umzugIn(t)
		agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
		legeAn(t, filepath.Join(agenten, altdienstlabel+".plist"), "alt")

		//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
		if err := os.Chmod(agenten, 0o500); err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() { _ = os.Chmod(agenten, 0o700) }) //nolint:errcheck,gosec // nur fürs Aufräumen

		if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "schreiben") || len(buch.nur("launchctl")) != 0 {
			t.Errorf("Fehler %v, Befehle %v", err, buch.befehle)
		}
	})

	t.Run("beide stehen da, und der alte lässt sich nicht entladen", func(t *testing.T) {
		t.Parallel()

		umzug, buch := umzugIn(t)
		agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
		legeAn(t, filepath.Join(agenten, altdienstlabel+".plist"), "alt")
		legeAn(t, filepath.Join(agenten, dienstlabel+".plist"), "neu")
		buch.scheitert = "launchctl bootout gui/501/" + altdienstlabel

		if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "entladen") {
			t.Errorf("Fehler %v", err)
		}

		// **Die alte bleibt**, solange sie geladen sein kann.
		if _, err := os.Stat(filepath.Join(agenten, altdienstlabel+".plist")); err != nil {
			t.Error("die alte plist ist weg, obwohl der Dienst nicht entladen wurde")
		}
	})

	t.Run("der alte Dienst lässt sich nicht entladen", func(t *testing.T) {
		t.Parallel()

		umzug, buch := umzugIn(t)
		agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
		legeAn(t, filepath.Join(agenten, altdienstlabel+".plist"), "alt")
		buch.scheitert = "launchctl bootout gui/501/" + altdienstlabel

		if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "entladen") {
			t.Errorf("Fehler %v", err)
		}

		// **Die neue plist geht wieder weg**, die alte bleibt.
		if _, err := os.Stat(filepath.Join(agenten, dienstlabel+".plist")); err == nil {
			t.Error("die neue plist steht da")
		}
	})
}

// **Lässt sich der frische Ordner nicht beiseitelegen, zieht nichts um** —
// und der Fehler sagt es (Zweitblick über Phase 3, Befund 10).
func TestEinUnverrueckbarerNeuerOrdnerWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	unter := filepath.Join(umzug.heim, "Library", "WebKit")
	legeAn(t, filepath.Join(unter, altkennungen[0], "x"), "alt")
	legeAn(t, filepath.Join(unter, Kennung, "x"), "frisch")

	//#nosec G302 -- absichtlich zugesperrt, um den Fehlerweg zu betreten
	if err := os.Chmod(unter, 0o500); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(unter, 0o700) }) //nolint:errcheck,gosec // nur fürs Aufräumen

	if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "beiseitelegen") {
		t.Errorf("Fehler %v", err)
	}
}

// **Lässt sich die Marke nicht schreiben, bekommt der Umzug seinen Namen
// nicht** — sonst stünde ein Ordner da, der beim nächsten Start beiseiteginge.
func TestEineUnschreibbareMarkeWirdGemeldet(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	unter := filepath.Join(umzug.heim, "Library", "WebKit")
	legeAn(t, filepath.Join(unter, altkennungen[0], umzugsdatei, "ein Ordner"), "an der Stelle der Marke")

	if err := umzug.ziehen(t.Context()); err == nil || !strings.Contains(err.Error(), "vermerken") {
		t.Errorf("Fehler %v", err)
	}

	if _, err := os.Stat(filepath.Join(unter, Kennung)); err == nil {
		t.Error("der Umzug steht ohne Marke unter der neuen Kennung")
	}
}

// **Was `defaults export` schreibt, liest der Umzug** — eine binäre Plist,
// mit dem echten `plutil` nach XML gebracht und vereint; und was dabei
// entsteht, nimmt `plutil -lint` an. Ohne Sperre, denn es fasst nichts dieses
// Rechners an.
func TestEineBinaerePlistWirdVereint(t *testing.T) {
	t.Parallel()

	datei := filepath.Join(t.TempDir(), "vorgaben.plist")
	legeAn(t, datei, `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>a</key><data>AAEC</data><key>b</key><dict><key>c</key><true/></dict></dict></plist>
`)

	for _, befehl := range [][]string{{"plutil", "-convert", "binary1", datei}, {"plutil", "-convert", "xml1", datei}} {
		if err := befehlAusfuehren(t.Context(), befehl); err != nil {
			t.Fatal(err)
		}
	}

	vereint, err := vorgabenVereinen([]byte(inhalt(t, datei)), []byte(leereVorgaben), altkennungen[0])
	if err != nil {
		t.Fatal(err)
	}

	ziel := filepath.Join(t.TempDir(), "vereint.plist")
	legeAn(t, ziel, string(vereint))

	if err := befehlAusfuehren(t.Context(), []string{"plutil", "-lint", ziel}); err != nil {
		t.Fatalf("%v:\n%s", err, vereint)
	}

	for _, erwartet := range []string{"<data>", "AAEC", "<key>c</key>", "<key>" + umzugsmarke + "</key>"} {
		if !strings.Contains(string(vereint), erwartet) {
			t.Errorf("es fehlt %q:\n%s", erwartet, vereint)
		}
	}
}
