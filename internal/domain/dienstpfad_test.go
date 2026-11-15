package domain_test

// **Was ein offener Server an Pfaden annimmt** — OPS-02.AK-04
// (#1719):
// „Ein offener Server darf über Pfadangaben keine beliebigen fremden
// Hostdateien offenlegen oder überschreiben."

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestPruefeDienstpfad(t *testing.T) {
	t.Parallel()

	const heim = "/Users/ich"

	weitere := []string{"/Volumes/Karten/Ausfuhr", "/var/folders/xy/T"}

	faelle := []struct {
		name string
		pfad string
		gut  bool
	}{
		{"unter dem Zuhause", "/Users/ich/Documents/Sets/probe.mse-set", true},
		{"das Zuhause selbst", "/Users/ich", true},
		{"ein eingestellter Ordner auf einem Laufwerk", "/Volumes/Karten/Ausfuhr/bogen.pdf", true},
		{"der eigene Zwischenablageordner", "/var/folders/xy/T/lauf/probe.mse-set", true},

		// **Ein verborgener Winkel ist überall einer** — auch in einem
		// erlaubten Ordner. Gefunden in der Pipeline: dort liegt das Zuhause
		// selbst unter `$TMPDIR`, und `~/.ssh` kam über die Erlaubnis für den
		// Zwischenablageordner durch.
		{"ein verborgener Ordner in einem erlaubten", "/var/folders/xy/T/heim/.ssh/id_ed25519", false},
		{"ein verborgener Ordner auf einem Laufwerk", "/Volumes/Karten/Ausfuhr/.git/config", false},
		{"ein verborgener Ordner im Zuhause", "/Users/ich/.ssh/id_ed25519", false},
		{"eine verborgene Datei im Zuhause", "/Users/ich/.zshrc", false},
		{"die Bibliothek mit Schlüsselbund und Anwendungsdaten", "/Users/ich/Library/Keychains/login.keychain-db", false},
		// **Und in jeder Schreibweise** — Befund 1 des Zweitblicks über Phase
		// 6.4: APFS unterscheidet Groß- und Kleinschreibung in der Vorgabe
		// nicht, und `library` ist dieselbe Bibliothek.
		{"die Bibliothek klein geschrieben", "/Users/ich/library/LaunchAgents/x.plist", false},
		{"die Bibliothek gemischt geschrieben", "/Users/ich/LIBRARY/Keychains/x", false},
		{"ein fremdes Zuhause", "/Users/anna/Documents/x.mse-set", false},
		{"das System", "/etc/passwd", false},
		{"neben einem eingestellten Ordner", "/Volumes/Karten/Privat/x", false},
		{"ein Präfix ist kein Ordner", "/Volumes/Karten/Ausfuhr2/x", false},
		{"relativ — relativ wozu?", "Documents/x.mse-set", false},
		{"nach oben hinaus", "/Users/ich/Documents/../../anna/x", false},
		{"leer", "", false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeDienstpfad(fall.pfad, heim, weitere)
			if fall.gut && fehler != nil {
				t.Fatalf("%q abgewiesen: %v", fall.pfad, fehler)
			}

			if !fall.gut && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q angenommen oder falsch abgewiesen: %v", fall.pfad, fehler)
			}
		})
	}
}

// **Die Namen, unter denen ein Pfad reist**, sind eine Aufzählung — ein
// Wächter hält sie gegen die Felder von REST und MCP.
func TestPfadfelderSindBenannt(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"pfad", "ziel", "datei", "verzeichnis"} {
		if !domain.IstPfadfeld(name) {
			t.Errorf("%s ist kein Pfadfeld", name)
		}
	}

	if domain.IstPfadfeld("name") {
		t.Error("name ist ein Pfadfeld")
	}
}

// **Die Pfade eines Rumpfs, in der Ordnung ihrer Namen** — dieselbe Lesart
// für Dienst und Werkzeugserver, damit zwei falsche Pfade immer dieselbe
// Meldung geben.
func TestPfadeInEinemRumpf(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		`{"ziel":"/b","name":"x","pfad":"/a","datei":7}`: "/a|/b",
		`["kein","objekt"]`: "",
		`kaputt`:            "",
		`{}`:                "",
	} {
		if gelesen := strings.Join(domain.PfadeIn([]byte(roh)), "|"); gelesen != erwartet {
			t.Errorf("%s: %q, erwartet %q", roh, gelesen, erwartet)
		}
	}
}

// **Welche Ordner ein offener Server überhaupt als seine gelten lässt** —
// Befund 2 des Zweitblicks über Phase 6.4: ein eingestellter Ordner weitete
// den Vertrag, und über den Dienst ließ sich jeder einstellen.
func TestPruefeDienstordner(t *testing.T) {
	t.Parallel()

	const heim = "/Users/ich"

	for ordner, gut := range map[string]bool{
		"/Users/ich/Documents/Ausfuhr":    true,
		"/Volumes/Karten/Ausfuhr":         true,
		"/Users/ich/Library/LaunchAgents": false,
		"/Users/ich/library/LaunchAgents": false,
		"/Users/ich/.ssh":                 false,
		"/":                               false,
		"/etc":                            false,
		"/Users/anna/Documents":           false,
		"/Volumes":                        false,
		"relativ":                         false,
	} {
		if fehler := domain.PruefeDienstordner(ordner, heim); (fehler == nil) != gut {
			t.Errorf("%q: %v", ordner, fehler)
		}
	}
}
