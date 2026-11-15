package service_test

// **Der Dienst unter dem alten Etikett** — OPS02/OPS03, Zweitblick über
// Phase 4, Befund 7: blieb der Kennungsumzug unvollständig oder lief er nie
// — die Kommandozeile führt ihn nicht aus —, stand ein LaunchAgent unter
// `com.wails.MagicSetEditor4Mac.dienst`, und „Alles widerrufen" endete
// erfolgreich, während er blieb.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitAltlast ist ein Agentenziel samt einer plist unter dem alten Etikett.
func mitAltlast(t *testing.T, inhalt string) (ziel service.Schnittstellenziel, gerufen *[][]string, alt string) {
	t.Helper()

	ziel, gerufen = einAgentenziel(t)
	alt = filepath.Join(filepath.Dir(ziel.Datei), "com.probe.alt.plist")
	ziel.Altlasten = []service.Altagent{{
		Datei: alt, Schluessel: "com.probe.alt", Programm: "mse-rest",
		Widerrufbefehle: [][]string{{"launchctl", "bootout", "gui/501/com.probe.alt"}},
	}}

	if err := os.MkdirAll(filepath.Dir(alt), 0o700); err != nil {
		t.Fatal(err)
	}

	schreibeDatei(t, alt, inhalt)

	return ziel, gerufen, alt
}

// eigeneAltlast ist eine plist, wie 0.8.0 sie schrieb — Etikett unter
// `Label`, das Programm als erstes Argument, in einem Bündel, das es nicht
// mehr gibt.
const eigeneAltlast = `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.probe.alt</string>
	<key>ProgramArguments</key>
	<array>
		<string>/Programme/Alt.app/Contents/MacOS/mse-rest</string>
	</array>
</dict>
</plist>
`

// **Nur die alte plist: veraltet, und der Widerruf nimmt sie mit.**
func TestEinAlterDienstWirdMitWiderrufen(t *testing.T) {
	t.Parallel()

	ziel, gerufen, alt := mitAltlast(t, eigeneAltlast)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleVeraltet {
		t.Fatalf("Stand %q — ein alter eigener Dienst ist nicht offen", stand)
	}

	if aus := ausgang(t)(service.SchnittstelleWiderrufen(t.Context(), ziel)); aus.Stand != service.SchnittstelleOffen {
		t.Fatalf("Widerrufen: %+v", aus)
	}

	if _, fehler := os.Stat(alt); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die alte plist steht noch: %v", fehler)
	}

	if !slices.ContainsFunc(*gerufen, func(b []string) bool {
		return slices.Equal(b, []string{"launchctl", "bootout", "gui/501/com.probe.alt"})
	}) {
		t.Errorf("der alte Dienst wurde nicht entladen: %v", *gerufen)
	}
}

// **Einrichten über einer alten plist räumt sie und richtet neu ein.**
func TestEinrichtenUeberEinemAltenDienstRaeumtIhn(t *testing.T) {
	t.Parallel()

	ziel, _, alt := mitAltlast(t, eigeneAltlast)

	if aus := ausgang(t)(service.SchnittstelleEinrichten(t.Context(), ziel)); aus.Stand != service.SchnittstelleEingerichtet {
		t.Fatalf("Einrichten: %+v", aus)
	}

	if _, fehler := os.Stat(alt); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("die alte plist steht noch: %v", fehler)
	}
}

// **Eine fremde plist unter dem alten Etikett bleibt** — ohne `mse-rest`
// ist sie nicht von hier, auch wenn sie so heißt.
func TestEineFremdeAltePlistBleibt(t *testing.T) {
	t.Parallel()

	fremd := strings.Replace(eigeneAltlast, "/Programme/Alt.app/Contents/MacOS/mse-rest", "/Andere.app/Contents/MacOS/server", 1)
	ziel, gerufen, alt := mitAltlast(t, fremd)

	if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
		t.Errorf("Stand %q", stand)
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if liesDatei(t, alt) != fremd || len(*gerufen) != 0 {
		t.Errorf("die fremde plist wurde angefasst: %v", *gerufen)
	}
}

// **Lässt sich der alte Dienst nicht entladen, bleibt seine plist** — und
// der Fehler wird genannt.
func TestEinAlterDienstDerNichtGehtWirdGenannt(t *testing.T) {
	t.Parallel()

	ziel, _, alt := mitAltlast(t, eigeneAltlast)
	kaputt := errors.New("Boot-out failed: 5")
	ziel.Ausfuehren = func(_ context.Context, befehl []string) error {
		if befehl[2] == "gui/501/com.probe.alt" {
			return kaputt
		}

		return nil
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler %v", fehler)
	}

	if liesDatei(t, alt) != eigeneAltlast {
		t.Error("die alte plist ging trotzdem")
	}
}

// **Eine alte plist, die sich nicht entfernen lässt, wird genannt.**
func TestEineAltePlistDieBleibtWirdGenannt(t *testing.T) {
	t.Parallel()

	ziel, _, alt := mitAltlast(t, eigeneAltlast)
	ordner := filepath.Dir(alt)

	if err := os.Chmod(ordner, 0o500); err != nil { //nolint:gosec // nur lesbar, für diesen Lauf
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(ordner, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); fehler == nil {
		t.Error("der Widerruf schwieg")
	}
}

// **Etikett und Programm an ihrer Stelle** (Nachlauf über Phase 4, Befund 6):
// zwei Zeichenfolgen irgendwo in der Datei genügten — ein Etikett unter einem
// anderen Schlüssel, `/mse-rest` als Ausgabepfad, ein fremdes Programm dieses
// Namens in einem Bündel, das es gibt.
func TestEineAltlastBrauchtEtikettUndProgrammAnIhrerStelle(t *testing.T) {
	t.Parallel()

	fremdesBuendel := filepath.Join(t.TempDir(), "Fremd.app")
	if err := os.MkdirAll(filepath.Join(fremdesBuendel, "Contents", "MacOS"), 0o700); err != nil {
		t.Fatal(err)
	}

	schreibeDatei(t, filepath.Join(fremdesBuendel, "Contents", "Info.plist"),
		"<plist><dict><key>CFBundleIdentifier</key><string>com.fremd.x</string></dict></plist>")

	for name, inhalt := range map[string]string{
		"Etikett unter anderem Schlüssel": strings.Replace(eigeneAltlast, "<key>Label</key>\n\t<string>com.probe.alt</string>",
			"<key>Label</key>\n\t<string>com.fremd</string>\n\t<key>Notiz</key>\n\t<string>com.probe.alt</string>", 1),
		"mse-rest nur als Ausgabe": strings.Replace(eigeneAltlast, "/Programme/Alt.app/Contents/MacOS/mse-rest",
			"/usr/bin/true</string>\n\t\t<string>/tmp/mse-rest", 1),
		"mse-rest in einem fremden Bündel": strings.Replace(eigeneAltlast, "/Programme/Alt.app",
			fremdesBuendel, 1),
	} {
		ziel, _, _ := mitAltlast(t, inhalt)

		if stand := service.SchnittstelleLesen(ziel); stand != service.SchnittstelleOffen {
			t.Errorf("%s: Stand %q", name, stand)
		}
	}
}

// **Die Ansage geht erst, wenn jeder eigene Dienst entladen ist** (Nachlauf
// über Phase 4, Befund 7): stand nur der alte da und ließ er sich nicht
// entladen, lief er weiter — ohne seine Ansage.
func TestDieAnsageBleibtWennDerAlteDienstNichtGeht(t *testing.T) {
	t.Parallel()

	ziel, _, alt := mitAltlast(t, eigeneAltlast)

	if err := os.MkdirAll(filepath.Dir(ziel.Ansage), 0o700); err != nil {
		t.Fatal(err)
	}

	schreibeDatei(t, ziel.Ansage, "Marke alt\n")

	kaputt := errors.New("Boot-out failed: 5")
	ziel.Ausfuehren = func(_ context.Context, befehl []string) error {
		if befehl[2] == "gui/501/com.probe.alt" {
			return kaputt
		}

		return nil
	}

	if _, fehler := service.SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler %v", fehler)
	}

	if liesDatei(t, ziel.Ansage) != "Marke alt\n" || liesDatei(t, alt) != eigeneAltlast {
		t.Error("Ansage oder alte plist gingen, obwohl der alte Dienst weiterläuft")
	}
}

// **Eine alte plist, die sich nicht lesen lässt, ist nicht von hier** — an
// ihrer Stelle steht ein Ordner.
func TestEineUnlesbareAltlastIstNichtEigen(t *testing.T) {
	t.Parallel()

	ziel, _ := einAgentenziel(t)
	alt := filepath.Join(filepath.Dir(ziel.Datei), "com.probe.alt.plist")

	if err := os.MkdirAll(filepath.Join(alt, "darin"), 0o700); err != nil {
		t.Fatal(err)
	}

	if service.AltlastEigen(service.Altagent{Datei: alt, Schluessel: "com.probe.alt", Programm: "mse-rest"}) {
		t.Error("ein Ordner galt als eigene plist")
	}
}
