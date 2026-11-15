// **Was die Diagnose über Schema und Pakete sagt** — Befund 2 des Zweitblicks
// über Phase 3.5
// (#1211).
//
// **Eine eigene Datei**, weil `diagnose_test.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, was die Diagnose **ist** und
// wie sie gesichert wird, hier, was sie über den **Bestand** sagt.
package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schemaprobe gibt eine feste Migrationsliste.
type schemaprobe struct {
	namen  []string
	fehler error

	// waisen ist die Zeile über verwaiste Zeilen — leer heißt „keine" (#1220).
	waisen       string
	waisenfehler error
}

func (s schemaprobe) AngewendeteMigrationen(context.Context) ([]string, error) {
	return s.namen, s.fehler
}

func (s schemaprobe) Waisen(context.Context) (string, error) {
	if s.waisen == "" {
		return "keine", s.waisenfehler
	}

	return s.waisen, s.waisenfehler
}

// **Der Schemastand nennt Zahl und jüngste Migration.**
func TestDerSchemastandNenntZahlUndJuengste(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitSchemastand(schemaprobe{namen: []string{"0001_a", "0002_b"}})

	if text := diagnosetext(t, anwendung); !strings.Contains(
		text, "2 Migrationen, zuletzt 0002_b") {
		t.Errorf("der Schemastand fehlt:\n%s", text)
	}
}

// **Ein leerer Bestand sagt es.**
func TestEinLeererSchemastandSagtEs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitSchemastand(schemaprobe{})

	if text := diagnosetext(t, anwendung); !strings.Contains(
		text, "keine Migration verbucht") {
		t.Errorf("die Auskunft fehlt:\n%s", text)
	}
}

// **Und ein Lesefehler ebenso** — genau dann ist die Diagnose am nötigsten.
func TestEinUnlesbarerSchemastandSagtEs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitSchemastand(schemaprobe{fehler: errSchema})

	if text := diagnosetext(t, anwendung); !strings.Contains(text, "nicht zu lesen") {
		t.Errorf("der Grund fehlt:\n%s", text)
	}
}

// **Die Paketstände stehen mit Namen und Fassung da.**
func TestDiePaketstaendeStehenDa(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketfuehrer{pakete: []domain.Paket{
		{Name: "magic", Version: "2020-01-01"},
		{Name: "magic-m15", Version: "2019-05-05"},
	}})

	text := diagnosetext(t, anwendung)

	for _, erwartet := range []string{"Pakete: 2", "magic 2020-01-01"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Diagnose nennt %q nicht:\n%s", erwartet, text)
		}
	}
}

// **Und über der Schranke steht, wie viele fehlen.**
//
// Ein Bestand mit dem vollen Pack hat 331 Pakete; sie alle aufzuzählen wären
// 331 Zeilen vor dem Protokoll, und das Protokoll ist der Teil, den jemand
// liest. **Die Zahl steht immer davor** — sie ist die Auskunft, die zählt.
func TestUeberDerSchrankeStehtWieVieleFehlen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	viele := make([]domain.Paket, 0, app.Paketzeilen+3)
	for i := range app.Paketzeilen + 3 {
		viele = append(viele, domain.Paket{
			Name: fmt.Sprintf("paket-%d", i), Version: "1",
		})
	}

	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketfuehrer{pakete: viele})

	text := diagnosetext(t, anwendung)

	if !strings.Contains(text, "… und 3 weitere") {
		t.Errorf("die Schranke sagt nicht, wie viele fehlen:\n%s", text)
	}
}

// **Ein Lesefehler der Pakete hält nichts auf.**
func TestEinPaketfehlerHaeltDieDiagnoseNichtAuf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitPaketen(&paketfuehrer{fehler: errSchema})

	text := diagnosetext(t, anwendung)

	if !strings.Contains(text, "Pakete: nicht zu lesen") {
		t.Errorf("der Grund fehlt:\n%s", text)
	}

	if !strings.Contains(text, app.Fassung) {
		t.Errorf("der Rest fehlt:\n%s", text)
	}
}

// errSchema ist der Lesefehler dieser Läufe.
var errSchema = errors.New("der Bestand ist nicht zu lesen")

// **Die Diagnose spricht die Sprache der Oberfläche.**
//
// # Der Befund
//
// Der Text kommt aus Go und geht durch **keinen** der fünf Sprachkataloge —
// `keintextimjsx` sieht ihn nicht. Die englische Oberfläche übersetzte
// Überschrift, Erklärung und Knöpfe und zeigte dann `Bestand`, `Migrationen`,
// `Pakete` und `Kein Protokoll` auf Deutsch: zweisprachig **innerhalb**
// derselben Fläche, und zwar gerade in dem Bereich, der für wenig technische
// Menschen gebaut ist.
//
// Angestrichen vom Nachlauf über Phase 3.5 als Befund 4.
func TestDieDiagnoseSprichtDieSpracheDerOberflaeche(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitSchemastand(schemaprobe{namen: []string{"0001_a"}})
	anwendung.SpracheSetzen("en")

	text := diagnosetext(t, anwendung)

	for _, erwartet := range []string{"Database:", "Schema:", "Packages:", "No log"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die englische Diagnose nennt %q nicht:\n%s", erwartet, text)
		}
	}

	// **Und kein deutsches Wort steht darin.** Ohne diese Hälfte wäre auch ein
	// Text grün, der beide Sprachen mischt.
	for _, deutsch := range []string{"Bestand:", "Pakete:", "Kein Protokoll"} {
		if strings.Contains(text, deutsch) {
			t.Errorf("die englische Diagnose nennt %q:\n%s", deutsch, text)
		}
	}
}

// **Ein Fehler beim Zählen der Waisen hält die Diagnose nicht auf** (§8).
//
// Sie ist der Bericht für jemanden, dessen Anwendung klemmt; ein Weg, der
// daran abbräche, ließe genau den Fall unberichtet, für den es ihn gibt
// (#1220).
func TestEinFehlerBeimWaisenzaehlenHaeltDieDiagnoseNichtAuf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitSchemastand(schemaprobe{
		namen:        []string{"0001_grundlage.sql"},
		waisenfehler: errors.New("der Bestand schweigt"),
	})

	text := diagnosetext(t, anwendung)
	if !strings.Contains(text, "der Bestand schweigt") {
		t.Errorf("die Diagnose verschweigt den Zählfehler:\n%s", text)
	}
}

// **Und ohne Naht sagt sie, dass sie es nicht lesen konnte.**
func TestOhneSchemanahtStehtNichtLesbarDa(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")

	if text := diagnosetext(t, anwendung); !strings.Contains(text, "Verwaiste Zeilen") {
		t.Errorf("die Zeile über verwaiste Zeilen fehlt:\n%s", text)
	}
}
