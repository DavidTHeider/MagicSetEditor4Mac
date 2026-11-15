// **Die drei Werkzeuge, die den Arbeitsablauf schließen**
// (#886,
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
//
// Einfahren, sichern, zeichnen — bis heute konnte ein Sprachmodell keines
// davon, während die Kommandozeile alle drei kann.
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// mitAusfuhr sucht ein Werkzeug in der Liste mit Ausfuhr.
func mitAusfuhr(t *testing.T, probe *ausfuhrprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       &bestandsprobe{},
		Schreiber:   &bestandsprobe{},
		Ausfuhr:     probe,
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// rufeRoh ruft ein Werkzeug und gibt seine Antwort als Text.
func rufeRoh(t *testing.T, werkzeug mcp.Werkzeug, argumente string) string {
	t.Helper()

	roh, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
	if err != nil {
		t.Fatalf("rufen: %v", err)
	}

	return roh
}

// ausfuhrprobe merkt sich, was gerufen wurde.
type ausfuhrprobe struct {
	eingefahren string
	gesichert   string
	gezeichnet  string
	gelegt      string
	gejsont     string
	fehler      error

	// guete ist die Güte, die das Werkzeug durchgereicht hat (Ω.4, #1029).
	guete string
}

func (a *ausfuhrprobe) Einfahren(_ context.Context, pfad string) (mcp.Einfahrt, error) {
	a.eingefahren = pfad

	if a.fehler != nil {
		return mcp.Einfahrt{}, a.fehler
	}

	return mcp.Einfahrt{SetID: 7, Karten: 3, Befunde: []string{"eine Zeile war krumm"}}, nil
}

func (a *ausfuhrprobe) Sichern(_ context.Context, _ int64, ziel string) ([]string, error) {
	a.gesichert = ziel

	return []string{}, a.fehler
}

func (a *ausfuhrprobe) Zeichnen(
	_ context.Context, _ int64, ziel, guete string,
) (breite, hoehe int, befunde []string, fehler error) {
	a.gezeichnet, a.guete = ziel, guete

	if a.fehler != nil {
		return 0, 0, nil, a.fehler
	}

	return 375, 523, nil, nil
}

func (a *ausfuhrprobe) Bogen(
	_ context.Context, _ int64, ziel string,
) (seiten, karten int, befunde []string, fehler error) {
	a.gelegt = ziel

	if a.fehler != nil {
		return 0, 0, nil, a.fehler
	}

	return 1, 3, []string{}, nil
}

func (a *ausfuhrprobe) MTGJSONLesen(_ context.Context, pfad, _ string) (
	mcp.MTGJSONEinfahrt, error,
) {
	a.gejsont = pfad

	// **Auch im Fehlerfall mit Teilerfolg** (#1291, Befund 1 des Zweitblicks):
	// eine AllSets-Datei kann mitten im Lauf abbrechen, und was vorher
	// durchkam, steht im Bestand.
	if a.fehler != nil {
		return mcp.MTGJSONEinfahrt{Angelegt: []int64{7}, Befunde: []string{}}, a.fehler
	}

	return mcp.MTGJSONEinfahrt{
		Angelegt: []int64{7, 8}, Befunde: []string{},
	}, nil
}

func (a *ausfuhrprobe) MTGJSONSchreiben(_ context.Context, _ int64, _, ziel string) (
	karten int, befunde []string, fehler error,
) {
	a.gejsont = ziel

	if a.fehler != nil {
		return 0, nil, a.fehler
	}

	return 3, []string{}, nil
}

func (a *ausfuhrprobe) MTGJSONSpiele() []string { return []string{"magic"} }

// **Die drei stehen in der Liste.**
func TestDerArbeitsablaufStehtUeberMCP(t *testing.T) {
	namen := map[string]bool{}
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       &bestandsprobe{},
		Schreiber:   &bestandsprobe{},
		Ausfuhr:     &ausfuhrprobe{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		namen[werkzeug.Name] = true
	}

	for _, name := range []string{
		"schreiben.set.einfahren", "schreiben.set.sichern", "schreiben.karte.zeichnen",
	} {
		if !namen[name] {
			t.Errorf("%q fehlt", name)
		}
	}
}

// **Einfahren nennt Kennung, Kartenzahl und Befunde.**
//
// Die Befunde gehören dazu (§8): ein Set mit 300 lesbaren und zwei unlesbaren
// Karten wird eingefahren, nicht abgewiesen — und wer die zwei nicht sähe,
// hielte ein halbes Set für ein ganzes.
func TestEinfahrenNenntBefunde(t *testing.T) {
	probe := &ausfuhrprobe{}

	antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.set.einfahren"),
		`{"pfad":"/tmp/probe.mse-set"}`)

	if probe.eingefahren != "/tmp/probe.mse-set" {
		t.Errorf("eingefahren wurde %q", probe.eingefahren)
	}

	for _, teil := range []string{`"setID"`, `"karten"`, "krumm"} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("die Antwort nennt %s nicht: %s", teil, antwort)
		}
	}
}

// **Zeichnen nennt die Maße.**
func TestZeichnenNenntDieMasse(t *testing.T) {
	probe := &ausfuhrprobe{}

	antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.karte.zeichnen"),
		`{"karteID":1,"ziel":"/tmp/karte.png"}`)

	if probe.gezeichnet != "/tmp/karte.png" {
		t.Errorf("gezeichnet wurde nach %q", probe.gezeichnet)
	}

	if !strings.Contains(antwort, "375") || !strings.Contains(antwort, "523") {
		t.Errorf("die Antwort nennt die Maße nicht: %s", antwort)
	}
}

// **Die Güte reist mit, und ohne sie bleibt sie leer.**
//
// Werkstück Ω.4 aus
// #1029: das
// Werkzeug nimmt `guete` entgegen und reicht sie durch. **Leer bleibt leer**
// — die Übersetzung auf die Vorgabe steht in `app.GueteAusName` und nicht
// hier; sie zweimal zu treffen hieße, zwei Stellen zu haben, die
// auseinanderlaufen können.
func TestDieGueteReistDurchDasWerkzeug(t *testing.T) {
	for name, gewollt := range map[string]string{
		`{"karteID":1,"ziel":"/tmp/a.png","guete":"ultra"}`: "ultra",
		`{"karteID":1,"ziel":"/tmp/a.png","guete":"fast"}`:  "fast",
		`{"karteID":1,"ziel":"/tmp/a.png"}`:                 "",
	} {
		probe := &ausfuhrprobe{}

		rufeRoh(t, mitAusfuhr(t, probe, "schreiben.karte.zeichnen"), name)

		if probe.guete != gewollt {
			t.Errorf("%s gab die Güte %q statt %q", name, probe.guete, gewollt)
		}
	}
}

// **Ein Fehler der Ausfuhr kommt an.**
func TestEinFehlerDerAusfuhrKommtAn(t *testing.T) {
	probe := &ausfuhrprobe{fehler: errors.New("die Platte ist voll")}

	for _, fall := range []struct{ name, argumente string }{
		{"schreiben.set.einfahren", `{"pfad":"/tmp/x.mse-set"}`},
		{"schreiben.set.sichern", `{"setID":1,"ziel":"/tmp/x.mse-set"}`},
		{"schreiben.karte.zeichnen", `{"karteID":1,"ziel":"/tmp/x.png"}`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			werkzeug := mitAusfuhr(t, probe, fall.name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(fall.argumente)); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Keines der drei sagt Idempotenz zu.**
//
// Ein zweiter Ruf legt eine Datei an, die es noch nicht gab, oder er wird
// abgewiesen, weil sie schon da ist — beides ist eine zusätzliche Wirkung
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
func TestDieDreiSagenKeineIdempotenzZu(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       &bestandsprobe{},
		Schreiber:   &bestandsprobe{},
		Ausfuhr:     &ausfuhrprobe{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if !strings.HasPrefix(werkzeug.Name, "schreiben.set.") &&
			werkzeug.Name != "schreiben.karte.zeichnen" {
			continue
		}

		if werkzeug.Idempotent {
			t.Errorf("%q sagt Idempotenz zu", werkzeug.Name)
		}
	}
}

// **Kaputte Argumente werden gemeldet, nicht verschluckt.**
//
// Ein Sprachmodell schickt manchmal eine Zahl als Text; die Antwort muss sagen,
// was nicht ging, statt still nichts zu tun.
func TestKaputteArgumenteDerDreiWerdenGemeldet(t *testing.T) {
	probe := &ausfuhrprobe{}

	for _, name := range []string{
		"schreiben.set.einfahren", "schreiben.set.sichern", "schreiben.karte.zeichnen",
	} {
		t.Run(name, func(t *testing.T) {
			werkzeug := mitAusfuhr(t, probe, name)

			if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{`)); err == nil {
				t.Error("kaputte Argumente kamen durch")
			}
		})
	}
}

// **Sichern nennt den Pfad zurück.**
func TestSichernNenntDenPfad(t *testing.T) {
	probe := &ausfuhrprobe{}

	antwort := rufeRoh(t, mitAusfuhr(t, probe, "schreiben.set.sichern"),
		`{"setID":1,"ziel":"/tmp/raus.mse-set"}`)

	if probe.gesichert != "/tmp/raus.mse-set" {
		t.Errorf("gesichert wurde nach %q", probe.gesichert)
	}

	if !strings.Contains(antwort, "/tmp/raus.mse-set") {
		t.Errorf("die Antwort nennt den Pfad nicht: %s", antwort)
	}
}
