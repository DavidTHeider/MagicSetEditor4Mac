package app

// **Der Umzug von `com.wails.*` auf die neue Kennung** — R02
// (#2334),
// Vermerk 0191. Gegen ein Zuhause im Testordner; `launchctl` und `defaults`
// gehen durch eine aufzeichnende Naht und erreichen diesen Rechner nie.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// aufzeichnung merkt sich jeden Befehl und lässt ihn gelingen — oder nicht.
// `domaenen` sind die Vorgaben, die cfprefsd führt — eine Domäne oder
// „Domäne Schlüssel": `defaults read` gelingt nur für sie.
type aufzeichnung struct {
	befehle   [][]string
	scheitert string
	domaenen  map[string]bool

	// ausgaenge geben einem Befehl einen Ausgang wie ein beendetes Programm
	// — `launchctl print` endet für einen ungeladenen Dienst mit 113.
	ausgaenge map[string]int

	// inhalte sind die Vorgaben je Domäne als XML-Plist — `defaults export`
	// schreibt sie in die Datei, `defaults import` liest sie daraus (#2424).
	inhalte map[string]string
}

func (a *aufzeichnung) ausfuehren(_ context.Context, befehl []string) error {
	a.befehle = append(a.befehle, befehl)
	if code, da := a.ausgaenge[strings.Join(befehl, " ")]; da {
		return programmausgang(code)
	}

	// **Wie `defaults` es sagt** (gemessen am 2026-10-16, macOS 27.0.1:
	// `Domain '…' not found.`; ein fehlender Schlüssel: `… does not exist`).
	if len(befehl) >= 3 && befehl[0] == "defaults" && befehl[1] == "read" && !a.domaenen[strings.Join(befehl[2:], " ")] {
		return fmt.Errorf("Domain %s not found.", strings.Join(befehl[2:], " "))
	}

	if a.scheitert != "" && strings.HasPrefix(strings.Join(befehl, " "), a.scheitert) {
		return errors.New("abgelehnt")
	}

	return a.vorgabe(befehl)
}

// nur gibt die Befehle an ein Programm — die Läufe des Dienstes fragen nicht
// nach den Vorgaben, und die der Vorgaben nicht nach dem Dienst.
func (a *aufzeichnung) nur(programm string) [][]string {
	var gefiltert [][]string

	for _, befehl := range a.befehle {
		if befehl[0] == programm {
			gefiltert = append(gefiltert, befehl)
		}
	}

	return gefiltert
}

func umzugIn(t *testing.T) (kennungsumzug, *aufzeichnung) {
	t.Helper()

	buch := &aufzeichnung{}

	return kennungsumzug{heim: t.TempDir(), domaene: "gui/501", ausfuehren: buch.ausfuehren}, buch
}

func legeAn(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(pfad), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}
}

func inhalt(t *testing.T, pfad string) string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein Pfad im Testordner
	if err != nil {
		t.Fatalf("%s: %v", pfad, err)
	}

	return string(roh)
}

func TestDerWebspeicherZiehtUmUndDerAlteBleibt(t *testing.T) {
	t.Parallel()

	for _, alt := range altkennungen {
		t.Run(alt, func(t *testing.T) {
			t.Parallel()

			umzug, _ := umzugIn(t)
			alterOrt := filepath.Join(umzug.heim, "Library", "WebKit", alt, "WebsiteData", "LocalStorage", "a.sqlite3")
			legeAn(t, alterOrt, "weggelegt=0.8.0")

			if err := umzug.ziehen(t.Context()); err != nil {
				t.Fatal(err)
			}

			neu := filepath.Join(umzug.heim, "Library", "WebKit", Kennung, "WebsiteData", "LocalStorage", "a.sqlite3")
			if inhalt(t, neu) != "weggelegt=0.8.0" {
				t.Error("der Webspeicher ist nicht umgezogen")
			}

			// **Der alte bleibt** — er ist der Rückweg zum alten Bündel.
			if inhalt(t, alterOrt) != "weggelegt=0.8.0" {
				t.Error("der alte Webspeicher ist weg")
			}
		})
	}
}

// **Ein zweiter Start zieht nichts mehr um** — und überschreibt nichts, was
// die neue Fassung schon geschrieben hat.
func TestEinZweiterStartZiehtNichtsUm(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	legeAn(t, filepath.Join(umzug.heim, "Library", "WebKit", altkennungen[0], "x"), "alt")
	legeAn(t, filepath.Join(umzug.heim, "Library", "WebKit", Kennung, "x"), "neu")
	legeAn(t, filepath.Join(umzug.heim, "Library", "WebKit", Kennung, umzugsdatei), altkennungen[0])
	buch.domaenen = map[string]bool{Kennung + " " + umzugsmarke: true, altkennungen[0]: true}

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	if inhalt(t, filepath.Join(umzug.heim, "Library", "WebKit", Kennung, "x")) != "neu" {
		t.Error("der neue Webspeicher wurde überschrieben")
	}

	// **Gefragt und nichts geschrieben**: die Marken des Umzugs stehen schon da.
	for _, befehl := range buch.nur("defaults") {
		if befehl[1] != "read" {
			t.Errorf("ohne etwas umzuziehen gerufen: %v", befehl)
		}
	}

	if len(buch.nur("launchctl"))+len(buch.nur("plutil")) != 0 {
		t.Errorf("ohne etwas umzuziehen gerufen: %v", buch.befehle)
	}
}

// **Ein abgebrochener Umzug hinterlässt keinen halben Stand** — der
// Zwischenort zählt nicht als umgezogen.
func TestEinAbgebrochenerUmzugWirdNachgeholt(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	legeAn(t, filepath.Join(umzug.heim, "Library", "WebKit", altkennungen[0], "ganz"), "alt")
	legeAn(t, filepath.Join(umzug.heim, "Library", "WebKit", Kennung+zwischenendung, "halb"), "halb")

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	neu := filepath.Join(umzug.heim, "Library", "WebKit", Kennung)
	if inhalt(t, filepath.Join(neu, "ganz")) != "alt" {
		t.Error("der Umzug wurde nicht nachgeholt")
	}

	if _, err := os.Stat(filepath.Join(neu, "halb")); err == nil {
		t.Error("der halbe Stand ist mitgezogen")
	}
}

// **Ein Webordner, den WebKit nach einem gescheiterten Umzug selbst anlegte,
// verdeckt den alten nicht** (Zweitblick über Phase 3, Befund 10): ohne Marke
// geht er beiseite — nicht weg —, der alte zieht um, und die Marke steht da.
func TestEinNeuerWebordnerOhneMarkeWirdBeiseiteGelegt(t *testing.T) {
	t.Parallel()

	umzug, _ := umzugIn(t)
	webkit := filepath.Join(umzug.heim, "Library", "WebKit")
	legeAn(t, filepath.Join(webkit, altkennungen[0], "x"), "alt")
	legeAn(t, filepath.Join(webkit, Kennung, "x"), "frisch")

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	if inhalt(t, filepath.Join(webkit, Kennung, "x")) != "alt" {
		t.Error("der alte Webspeicher zog nicht um")
	}

	beiseite, _ := filepath.Glob(filepath.Join(webkit, Kennung+".vor-umzug-*", "x")) //nolint:errcheck // das Muster ist fest
	if len(beiseite) != 1 || inhalt(t, beiseite[0]) != "frisch" {
		t.Errorf("der frische Ordner ist nicht beiseite: %v", beiseite)
	}

	if inhalt(t, filepath.Join(webkit, Kennung, umzugsdatei)) != altkennungen[0]+"\n" {
		t.Error("die Marke des Webordners fehlt")
	}

	// **Und ein zweiter Start lässt ihn, wie er ist.**
	legeAn(t, filepath.Join(webkit, Kennung, "x"), "seither")

	if err := umzug.ziehen(t.Context()); err != nil || inhalt(t, filepath.Join(webkit, Kennung, "x")) != "seither" {
		t.Errorf("der umgezogene Ordner wurde noch einmal angefasst (%v)", err)
	}
}

func TestDerDienstZiehtAufDasNeueEtikettUm(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
	alt := filepath.Join(agenten, altdienstlabel+".plist")
	legeAn(t, alt, strings.ReplaceAll(dienstplist("/Programme/X.app/Contents/MacOS/mse-rest", "/tmp/dienst.txt", "/tmp/dienst.log"), dienstlabel, altdienstlabel))

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	neu := filepath.Join(agenten, dienstlabel+".plist")
	if !strings.Contains(inhalt(t, neu), "<string>"+dienstlabel+"</string>") || strings.Contains(inhalt(t, neu), altdienstlabel) {
		t.Errorf("die neue plist trägt nicht das neue Etikett:\n%s", inhalt(t, neu))
	}

	if !strings.Contains(inhalt(t, neu), "/Programme/X.app/Contents/MacOS/mse-rest") {
		t.Error("die neue plist startet ein anderes Programm als die alte")
	}

	if _, err := os.Stat(alt); err == nil {
		t.Error("die alte plist steht noch da")
	}

	// **Erst entladen, dann laden** — sonst liefen zwei an einem Port.
	erwartet := [][]string{
		{"launchctl", "bootout", "gui/501/" + altdienstlabel},
		{"launchctl", "bootstrap", "gui/501", neu},
	}
	if !slices.EqualFunc(buch.nur("launchctl"), erwartet, slices.Equal) {
		t.Errorf("Befehle %v, erwartet %v", buch.nur("launchctl"), erwartet)
	}
}

// **Lädt der neue nicht, bleibt der alte stehen** — ein Umzug, der den Dienst
// verliert, ist schlimmer als keiner. Die neue plist geht wieder weg, und ein
// nächster Start versucht es noch einmal.
func TestScheitertDasLadenBleibtDerAlteDienst(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
	alt := filepath.Join(agenten, altdienstlabel+".plist")
	legeAn(t, alt, strings.ReplaceAll(dienstplist("/x/mse-rest", "/tmp/d.txt", "/tmp/d.log"), dienstlabel, altdienstlabel))
	buch.scheitert = "launchctl bootstrap gui/501 " + filepath.Join(agenten, dienstlabel+".plist")

	if err := umzug.ziehen(t.Context()); err == nil {
		t.Fatal("der Umzug meldet den Fehler nicht")
	}

	if _, err := os.Stat(alt); err != nil {
		t.Error("die alte plist ist weg, obwohl der neue Dienst nicht lädt")
	}

	if _, err := os.Stat(filepath.Join(agenten, dienstlabel+".plist")); err == nil {
		t.Error("die neue plist steht da, obwohl sie nicht lädt")
	}

	if dienst := buch.nur("launchctl"); dienst[len(dienst)-1][1] != "bootstrap" || dienst[len(dienst)-1][3] != alt {
		t.Errorf("der alte Dienst wurde nicht wieder geladen: %v", buch.befehle)
	}
}

// **Stehen beide da, bleibt der neue** — wer zurückging, den Dienst am alten
// Fenster neu einrichtete und wieder aufstieg, hat einen geladenen unter dem
// neuen Etikett. Der alte wird entladen und geht, der neue bleibt, wie er ist.
func TestStehenBeideDienstplistsDaBleibtDieNeue(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
	alt := filepath.Join(agenten, altdienstlabel+".plist")
	neu := filepath.Join(agenten, dienstlabel+".plist")
	legeAn(t, alt, "alt")
	legeAn(t, neu, "neu")

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	if inhalt(t, neu) != "neu" {
		t.Error("die neue plist wurde überschrieben")
	}

	if _, err := os.Stat(alt); err == nil {
		t.Error("die alte plist steht noch da")
	}

	erwartet := [][]string{
		{"launchctl", "print", "gui/501/" + dienstlabel},
		{"launchctl", "bootout", "gui/501/" + altdienstlabel},
	}
	if !slices.EqualFunc(buch.nur("launchctl"), erwartet, slices.Equal) {
		t.Errorf("Befehle %v, erwartet %v", buch.nur("launchctl"), erwartet)
	}
}

// **Stehen beide da und ist der neue nicht geladen, wird er geladen** —
// Zweitblick über Phase 3, Befund 7: ein Umzug, der zwischen dem Schreiben
// der neuen plist und dem Laden abbrach, ließ beide stehen, und der nächste
// Start entlud den alten und lud keinen.
func TestStehenBeideDaUndDerNeueFehltWirdErGeladen(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	agenten := filepath.Join(umzug.heim, "Library", "LaunchAgents")
	alt := filepath.Join(agenten, altdienstlabel+".plist")
	neu := filepath.Join(agenten, dienstlabel+".plist")
	legeAn(t, alt, "alt")
	legeAn(t, neu, "neu")
	buch.ausgaenge = map[string]int{"launchctl print gui/501/" + dienstlabel: nichtGeladenCode}

	if err := umzug.ziehen(t.Context()); err != nil {
		t.Fatal(err)
	}

	erwartet := [][]string{
		{"launchctl", "print", "gui/501/" + dienstlabel},
		{"launchctl", "bootout", "gui/501/" + altdienstlabel},
		{"launchctl", "bootstrap", "gui/501", neu},
	}
	if !slices.EqualFunc(buch.nur("launchctl"), erwartet, slices.Equal) {
		t.Errorf("Befehle %v, erwartet %v", buch.nur("launchctl"), erwartet)
	}

	if _, err := os.Stat(alt); err == nil || inhalt(t, neu) != "neu" {
		t.Error("die alte plist steht noch da, oder die neue ist nicht mehr dieselbe")
	}
}

// **Steht nur der neue da, ist nichts zu tun** — der Umzug lief schon, oder
// der Dienst wurde unter der neuen Fassung eingerichtet.
func TestStehtNurDieNeueDienstplistIstNichtsZuTun(t *testing.T) {
	t.Parallel()

	umzug, buch := umzugIn(t)
	legeAn(t, filepath.Join(umzug.heim, "Library", "LaunchAgents", dienstlabel+".plist"), "neu")

	if err := umzug.ziehen(t.Context()); err != nil || len(buch.nur("launchctl")) != 0 {
		t.Errorf("Fehler %v, Befehle %v", err, buch.nur("launchctl"))
	}
}
