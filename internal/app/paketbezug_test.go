package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bezugsprobe ist ein Double des Paketbeziehers.
type bezugsprobe struct {
	ergebnis *service.Bezugsergebnis
	fehler   error
	adresse  string
	stand    string
	name     string
}

func (b *bezugsprobe) Herkunft(adresse, stand string) (quelle, gestand string) {
	return adresse, stand
}

func (b *bezugsprobe) Beziehen(
	_ context.Context, adresse, stand, _, name string,
) (*service.Bezugsergebnis, error) {
	b.adresse, b.stand, b.name = adresse, stand, name

	return b.ergebnis, b.fehler
}

// mitBezieher baut eine Anwendung mit Bezieher **und** Katalog.
//
// **Beide Nähte**, seit der Bezug vermerkt, was ankam (#56). Eine Vorrichtung,
// die eine Naht auslässt, prüft sie nicht — sie stürzt an ihr ab.
func mitBezieher(probe *bezugsprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitPaketbezieher(probe)
	anwendung.MitPaketen(&paketfuehrer{})

	return anwendung
}

// mitKatalog baut dieselbe Anwendung und gibt den Katalog dazu.
func mitKatalog(probe *bezugsprobe, katalog *paketfuehrer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitPaketbezieher(probe)
	anwendung.MitPaketen(katalog)

	return anwendung
}

// Ein Paket wird bezogen, und das Ergebnis nennt Stand und Prüfsumme.
func TestPaketBeziehenNenntStandUndPruefsumme(t *testing.T) {
	probe := &bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort:         "/pakete/M15",
		Stand:       "3.1.0",
		Pruefsumme:  "sha256:abc",
		Uebergangen: []string{"mse.exe"},
	}}

	antwort := mitBezieher(probe).PaketBeziehen(
		probekontext(t), "https://github.com/x/y", "3.1.0", "data", "M15")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if probe.adresse != "https://github.com/x/y" || probe.stand != "3.1.0" || probe.name != "M15" {
		t.Errorf("beim Bezieher kam %q, %q, %q an", probe.adresse, probe.stand, probe.name)
	}

	if antwort.Pruefsumme != "sha256:abc" || antwort.Stand != "3.1.0" {
		t.Errorf("%+v", antwort)
	}

	if len(antwort.Uebergangen) != 1 {
		t.Errorf("Übergangen: %v — was fehlt, wird genannt", antwort.Uebergangen)
	}
}

// **Die Herkunft steht vor dem Bezug** (#139) — als Angaben, nicht als Satz.
func TestPaketHerkunftNenntQuelleUndStand(t *testing.T) {
	antwort := mitBezieher(&bezugsprobe{}).
		PaketHerkunft("https://github.com/MagicSetEditorPacks/M15-Magic-Pack", "3.1.0")

	if antwort.Quelle == "" || antwort.Stand != "3.1.0" {
		t.Errorf("Herkunft: %+v", antwort)
	}
}

// Ein Fehler kommt als übersetzter Befund an, und die Liste bleibt eine Liste.
func TestPaketBeziehenMeldetBefunde(t *testing.T) {
	probe := &bezugsprobe{fehler: errors.New("die Gegenstelle antwortet nicht")}

	antwort := mitBezieher(probe).PaketBeziehen(probekontext(t), "https://x", "1.0", "", "P")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler ging verloren")
	}

	// **Eine leere Liste ist `[]`, nicht `null`** — die Oberfläche darf darüber
	// laufen, ohne vorher zu prüfen.
	if antwort.Uebergangen == nil {
		t.Error("Übergangen ist null statt einer leeren Liste")
	}
}

// Das Paketverzeichnis liegt neben der Datenbank und wird angelegt.
func TestDasPaketverzeichnisLiegtNebenDerDatenbank(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, err := app.Paketverzeichnis()
	if err != nil {
		t.Fatalf("Paketverzeichnis: %v", err)
	}

	if filepath.Base(ort) != "pakete" {
		t.Errorf("Ort: %q", ort)
	}

	angaben, err := os.Stat(ort)
	if err != nil || !angaben.IsDir() {
		t.Errorf("das Verzeichnis ist nicht angelegt worden: %v", err)
	}
}

// Ohne Benutzerverzeichnis wird gemeldet statt geraten.
func TestOhneBenutzerverzeichnisWirdGemeldet(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.Paketverzeichnis(); err == nil {
		t.Error("kein Fehler")
	}
}

// Ein Ort, an dem sich nichts anlegen lässt, wird gemeldet.
func TestEinUnschreibbaresPaketverzeichnisWirdGemeldet(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root lässt sich überall anlegen")
	}

	zu := filepath.Join(t.TempDir(), "zu")
	if err := os.Mkdir(zu, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(zu, 0o700); err != nil { // #nosec G302 -- Verzeichnis
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	// **`HOME`, nicht `XDG_CONFIG_HOME`**: `os.UserConfigDir` liest unter macOS
	// nur ersteres und legt `Library/Application Support` darunter.
	t.Setenv("HOME", zu)

	if _, err := app.Paketverzeichnis(); err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein Start ohne Paketverzeichnis meldet sich beim ersten Bezug.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9. Der Composition Root verwarf den Fehler von
// `Paketverzeichnis()` und baute den Bezieher mit **leerer** Wurzel — daraus
// wurde ein relativer Pfad, den ein späterer Bezug geräumt hätte.
//
// Der Start hält deswegen trotzdem nicht an: wer nie ein Paket bezieht, braucht
// keines. Der Fehler kommt dort ans Licht, wo er zählt.
func TestEinStartOhnePaketverzeichnisMeldetSichBeimBezug(t *testing.T) {
	// Eine Datei dort, wo das Verzeichnis hingehört: `MkdirAll` scheitert daran.
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	stelle := filepath.Join(zuhause, "Library", "Application Support", "MagicSetEditor4Mac")
	if err := os.MkdirAll(filepath.Dir(stelle), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(stelle, []byte("kein Verzeichnis"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	// **Der Start geht durch** — das ist die Zusage.
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	antwort := anwendung.PaketBeziehen(t.Context(), "https://example.test", "1.0", "data", "P")
	if antwort.Fehler == nil {
		t.Error("der Bezug lief los, obwohl es kein Paketverzeichnis gibt")
	}
}

// **Was ankam, wird vermerkt** — die zweite Hälfte der Zusage aus #56.
//
// Angezeigt hat die Paketseite die Prüfsumme seit #139; nach dem nächsten Start
// war sie fort. Jetzt steht sie am Paket, zusammen mit der Herkunft.
func TestEinBezugVermerktHerkunftUndPruefsumme(t *testing.T) {
	katalog := &paketfuehrer{}
	anwendung := mitKatalog(&bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort: "/pakete/M15", Stand: "3.1.0", Pruefsumme: "sha256:abc",
	}}, katalog)

	if antwort := anwendung.PaketBeziehen(
		probekontext(t), "https://github.com/x/y", "3.1.0", "data", "M15"); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if katalog.vermerkt.Ort != "/pakete/M15" ||
		katalog.vermerkt.Herkunft != "https://github.com/x/y@3.1.0" ||
		katalog.vermerkt.Pruefsumme != "sha256:abc" {
		t.Errorf("vermerkt: %+v", katalog.vermerkt)
	}
}

// **Ohne Stand steht die Adresse allein** — und kein „@" ohne etwas dahinter.
func TestEineHerkunftOhneStandBleibtDieAdresse(t *testing.T) {
	katalog := &paketfuehrer{}
	anwendung := mitKatalog(&bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort: "/pakete/M15", Pruefsumme: "sha256:abc",
	}}, katalog)

	if antwort := anwendung.PaketBeziehen(
		probekontext(t), "https://github.com/x/y", "", "", "M15"); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if katalog.vermerkt.Herkunft != "https://github.com/x/y" {
		t.Errorf("Herkunft: %q", katalog.vermerkt.Herkunft)
	}
}

// **Ein Fehler beim Vermerken nimmt den Bezug nicht zurück** (§7, Regel 3).
//
// Das Paket liegt auf der Platte; gemeldet wird trotzdem, denn der Katalog ist
// danach nicht auf dem Stand, den er behauptet.
func TestEinFehlerBeimVermerkenMeldetUndBehaeltDenBezug(t *testing.T) {
	katalog := &paketfuehrer{fehler: errors.New("der Katalog schweigt")}
	anwendung := mitKatalog(&bezugsprobe{ergebnis: &service.Bezugsergebnis{
		Ort: "/pakete/M15", Stand: "3.1.0", Pruefsumme: "sha256:abc",
	}}, katalog)

	antwort := anwendung.PaketBeziehen(t.Context(), "https://github.com/x/y", "3.1.0", "", "M15")
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if antwort.Ort != "/pakete/M15" || antwort.Pruefsumme != "sha256:abc" {
		t.Errorf("der Bezug ging verloren: %+v", antwort)
	}
}
