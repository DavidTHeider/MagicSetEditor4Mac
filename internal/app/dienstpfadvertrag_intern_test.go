package app

// **Der Pfadvertrag eines offenen Servers, an einer echten Platte** —
// OPS-02.AK-04 (#1719).
// Ohne Zwischenablageordner, damit das Zuhause unter `t.TempDir()` nicht
// schon durch ihn erlaubt ist.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerPfadvertragFolgtVerknuepfungenUndEingestelltenOrdnern(t *testing.T) {
	wurzel := t.TempDir()
	heim := filepath.Join(wurzel, "heim")
	laufwerk := filepath.Join(wurzel, "laufwerk")

	for _, ordner := range []string{filepath.Join(heim, ".ssh"), filepath.Join(heim, "Sets"), laufwerk} {
		if fehler := os.MkdirAll(ordner, 0o750); fehler != nil {
			t.Fatal(fehler)
		}
	}

	// **Eine Verknüpfung im Zuhause, die in einen verborgenen Winkel zeigt.**
	if fehler := os.Symlink(filepath.Join(heim, ".ssh"), filepath.Join(heim, "Karten")); fehler != nil {
		t.Fatal(fehler)
	}

	werte := map[string]string{domain.VorgabeAusfuhrpfad: laufwerk}
	einstellungen := func(context.Context) (map[string]string, error) { return werte, nil }
	vertrag := dienstpfadvertragMit(einstellungen, func() (string, error) { return heim, nil }, nil, wurzel)
	ctx := t.Context()

	for pfad, gut := range map[string]bool{
		filepath.Join(heim, "Sets", "probe.mse-set"):     true,
		filepath.Join(heim, "Sets", "neu", "bogen.pdf"):  true,
		filepath.Join(laufwerk, "bogen.pdf"):             true,
		filepath.Join(heim, ".ssh", "id_ed25519"):        false,
		filepath.Join(heim, "Karten", "id_ed25519"):      false,
		filepath.Join(wurzel, "anderswo", "x.mse-set"):   false,
		filepath.Join(heim, "Library", "Keychains", "x"): false,
	} {
		fehler := vertrag(ctx, pfad)
		if gut != (fehler == nil) || (!gut && !errors.Is(fehler, domain.ErrValidation)) {
			t.Errorf("%s: %v", pfad, fehler)
		}
	}

	// **Ohne lesbare Einstellungen fällt nur der eingestellte Ordner weg.**
	ohne := dienstpfadvertragMit(func(context.Context) (map[string]string, error) {
		return nil, errors.New("die Platte schweigt")
	}, func() (string, error) { return heim, nil }, nil, wurzel)

	if ohne(ctx, filepath.Join(laufwerk, "bogen.pdf")) == nil || ohne(ctx, filepath.Join(heim, "Sets", "x")) != nil {
		t.Error("ein Lesefehler hat den Vertrag verändert, wie er es nicht soll")
	}

	// **Ohne Zuhause gibt es keinen Vertrag, der etwas erlaubt.**
	kein := dienstpfadvertragMit(einstellungen, func() (string, error) { return "", errors.New("kein Zuhause") }, nil, wurzel)
	if !errors.Is(kein(ctx, filepath.Join(laufwerk, "x")), domain.ErrValidation) {
		t.Error("ohne Zuhause wurde etwas angenommen")
	}
}

// **Der Vertrag über einem echten Bestand** — mit dem Zwischenablageordner
// dieses Prozesses, wie `mse-rest` ihn bekommt.
func TestDerPfadvertragUeberEinemBestand(t *testing.T) {
	heim := t.TempDir()
	t.Setenv("HOME", heim)

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Error(fehler)
		}
	})

	vertrag := Dienstpfadvertrag(anwendung.db)
	if fehler := vertrag(t.Context(), "/etc/passwd"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("/etc/passwd: %v", fehler)
	}

	if fehler := vertrag(t.Context(), filepath.Join(os.TempDir(), "probe.mse-set")); fehler != nil {
		t.Errorf("der Zwischenablageordner: %v", fehler)
	}

	if fehler := vertrag(t.Context(), "relativ/x"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein relativer Pfad: %v", fehler)
	}

	// **Die eigenen Orte gelten** — Befund 4 des Zweitblicks über Phase 6.4:
	// Pakete und Vorsicherungen liegen unter `~/Library/Application
	// Support`, und der Vertrag wies Pfade ab, die der Dienst selbst ausgibt.
	pakete, fehler := Paketverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := vertrag(t.Context(), filepath.Join(pakete, "probe.mse-style")); fehler != nil {
		t.Errorf("ein Paket: %v", fehler)
	}

	if fehler := vertrag(t.Context(), filepath.Join(filepath.Dir(anwendung.datenbankpfad), "bestand-vor-0.5.0.db")); fehler != nil {
		t.Errorf("eine Vorsicherung: %v", fehler)
	}

	// **Und ein eingestellter Ordner in der Bibliothek weitet ihn nicht** —
	// Befund 2: er wäre über den Dienst einzustellen gewesen.
	agenten := filepath.Join(heim, "Library", "LaunchAgents")
	if fehler := os.MkdirAll(agenten, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if antwort := anwendung.EinstellungSetzen(t.Context(), domain.VorgabeAusfuhrpfad, agenten); antwort.Fehler != nil {
		t.Fatalf("am Fenster: %+v", antwort.Fehler)
	}

	if fehler := vertrag(t.Context(), filepath.Join(agenten, "boese.plist")); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein Ordner in der Bibliothek: %v", fehler)
	}

	// **Über den Dienst lässt er sich gar nicht erst einstellen** — und die
	// Antwort verrät nicht, was dort liegt: „ist eine Datei" hieße „gibt es".
	ssh := filepath.Join(heim, ".ssh")
	if fehler := os.MkdirAll(ssh, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(ssh, "id_ed25519"), []byte("geheim"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ueberDienst := MitDienstmodus(t.Context(), domain.ModusDienst)
	for _, angabe := range []string{"~/.ssh/id_ed25519", "~/Library/LaunchAgents", "/"} {
		antwort := anwendung.EinstellungSetzen(ueberDienst, domain.VorgabeAusfuhrpfad, angabe)
		if antwort.Fehler == nil || strings.Contains(antwort.Fehler.Meldung, "Datei") {
			t.Errorf("%s über den Dienst: %+v", angabe, antwort.Fehler)
		}
	}
}

// **`MSE_DIENSTORDNER` nennt weitere Ordner** — getrennt wie `PATH`.
func TestDienstordnerAusDerUmgebung(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(Dienstordnerumgebung, "/srv/karten"+string(filepath.ListSeparator)+"/Volumes/Ablage")

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Error(fehler)
		}
	})

	vertrag := Dienstpfadvertrag(anwendung.db)
	for pfad, gut := range map[string]bool{
		"/srv/karten/probe.mse-set": true,
		"/Volumes/Ablage/x.pdf":     true,
		"/srv/anderes/x":            false,
	} {
		if gut != (vertrag(t.Context(), pfad) == nil) {
			t.Errorf("%s: %v", pfad, vertrag(t.Context(), pfad))
		}
	}
}

func TestAufgeloestFolgtDemLaengstenAnfang(t *testing.T) {
	wurzel := t.TempDir()
	echt, fehler := filepath.EvalSymlinks(wurzel)
	if fehler != nil {
		t.Fatal(fehler)
	}
	datei := filepath.Join(wurzel, "datei")

	if fehler := os.WriteFile(datei, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	for pfad, erwartet := range map[string]string{
		"":                                      "",
		filepath.Join(wurzel, "gibts", "nicht"): filepath.Join(echt, "gibts", "nicht"),
		wurzel:                                  echt,
		"relativ/gibts/nicht":                   "relativ/gibts/nicht",

		// **Hinter einer Datei geht es nicht weiter** — kein »gibt es
		// nicht«, sondern »ist kein Ordner«; dann bleibt der Pfad, wie er
		// genannt ist.
		filepath.Join(datei, "innen") + "/": filepath.Join(datei, "innen"),
	} {
		if gelesen := aufgeloest(pfad); gelesen != erwartet {
			t.Errorf("aufgeloest(%q) = %q, erwartet %q", pfad, gelesen, erwartet)
		}
	}
}
