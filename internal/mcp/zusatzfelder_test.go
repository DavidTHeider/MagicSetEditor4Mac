package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// zusatzprobe hält die Zusatzfelder einer Karte und schreibt mit.
type zusatzprobe struct {
	felder     []mcp.Zusatzfeld
	gesetzt    []string
	fehler     error
	setzfehler error

	// ohneBlock ist eine Karte, die `extra data` gar nicht trägt.
	ohneBlock bool
}

func (z *zusatzprobe) Zusatzfelder(context.Context, int64, int64) ([]mcp.Zusatzfeld, error) {
	return z.felder, z.fehler
}

// ZusatzwertSetzen gibt den Vorwert so, wie das Setzen ihn liest — aus den
// Feldern der Probe, dazu ein Block, an dem der Rückweg zu erkennen ist.
func (z *zusatzprobe) ZusatzwertSetzen(
	_ context.Context, _, _ int64, feld, wert string,
) (mcp.Zusatzvorwert, error) {
	if z.fehler != nil {
		return mcp.Zusatzvorwert{}, z.fehler
	}

	if z.setzfehler != nil {
		return mcp.Zusatzvorwert{}, z.setzfehler
	}

	z.gesetzt = append(z.gesetzt, feld+"="+wert)

	vorher := mcp.Zusatzvorwert{Block: "magic-probe:\n\tvermerk: alt", BlockDa: !z.ohneBlock}
	if z.ohneBlock {
		vorher.Block = ""
	}
	for _, f := range z.felder {
		if f.Name == feld && f.Gespeichert {
			vorher.Wert, vorher.Gespeichert = f.Wert, true
		}
	}

	return vorher, nil
}

func zusatzwerkzeug(t *testing.T, zusatz mcp.Zusatzquelle, name string) (mcp.Werkzeug, bool) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(), Ausfuhr: ruhendeAusfuhr{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Zusatz: zusatz,
	}) {
		if werkzeug.Name == name {
			return werkzeug, true
		}
	}

	return mcp.Werkzeug{}, false
}

// **Die Zusatzfelder einer Karte lassen sich lesen und setzen** (#2298) — und
// das Setzen nennt, was vorher galt, und wie es zurückgeht.
func TestDieZusatzfelderStehenAlsWerkzeug(t *testing.T) {
	probe := &zusatzprobe{felder: []mcp.Zusatzfeld{
		{Name: "vermerk", Art: "text", Wert: "alt", Gespeichert: true},
		{Name: "promo label", Art: "text"},
	}}

	lesen, da := zusatzwerkzeug(t, probe, "lesen.zusatzfelder")
	if !da || lesen.Art != mcp.Lesend {
		t.Fatalf("lesen.zusatzfelder fehlt oder ist nicht lesend: %+v", lesen)
	}

	antwort, err := lesen.Ruf(t.Context(), json.RawMessage(`{"setID": 7, "karteID": 1}`))
	if err != nil || !strings.Contains(antwort, `"vermerk"`) || !strings.Contains(antwort, `"gespeichert": true`) {
		t.Errorf("Antwort %s, Fehler %v", antwort, err)
	}

	setzen, da := zusatzwerkzeug(t, probe, "schreiben.zusatzwert")
	if !da || setzen.Art != mcp.Schreibend {
		t.Fatalf("schreiben.zusatzwert fehlt oder ist nicht schreibend: %+v", setzen)
	}

	for _, fall := range []struct{ feld, vorher string }{
		{feld: "vermerk", vorher: `"vorher": "alt"`},
		// **Nicht gesetzt: der Rückweg nennt den Block** (Befund 6 des
		// Nachlaufs über E/F) — ohne ihn war er nicht auszuführen.
		{feld: "promo label", vorher: `"vorherBlock": "magic-probe:\n\tvermerk: alt"`},
	} {
		antwort, err := setzen.Ruf(t.Context(), json.RawMessage(
			`{"setID": 7, "karteID": 1, "feld": "`+fall.feld+`", "wert": "neu"}`))
		if err != nil || !strings.Contains(antwort, fall.vorher) || !strings.Contains(antwort, "rueckweg") {
			t.Errorf("%s: Antwort %s, Fehler %v", fall.feld, antwort, err)
		}
	}

	if strings.Join(probe.gesetzt, ";") != "vermerk=neu;promo label=neu" {
		t.Errorf("gesetzt: %v", probe.gesetzt)
	}
}

// **Fehlte der Block ganz, nimmt der Rückweg ihn weg** (J05, #2269, Befund 7
// des nachgeholten Zweitblicks über die Folgepakete): einen leeren Block
// zurückzuschreiben ließe ein neues, leeres Feld stehen, und der Kartenstand
// unterscheidet „fehlt" von „leer".
func TestEinFehlenderBlockGehtUeberEntfernenZurueck(t *testing.T) {
	probe := &zusatzprobe{felder: []mcp.Zusatzfeld{{Name: "promo label", Art: "text"}}, ohneBlock: true}

	setzen, _ := zusatzwerkzeug(t, probe, "schreiben.zusatzwert")

	antwort, err := setzen.Ruf(t.Context(), json.RawMessage(
		`{"setID": 7, "karteID": 1, "feld": "promo label", "wert": "neu"}`))
	if err != nil || !strings.Contains(antwort, "entfernen") || strings.Contains(antwort, "vorherBlock") {
		t.Errorf("Antwort %s, Fehler %v", antwort, err)
	}
}

// **Was nicht geht, wird benannt, und ohne Naht gibt es die Werkzeuge nicht.**
func TestDieZusatzwerkzeugeWeisenAbWasNichtGeht(t *testing.T) {
	setzen, _ := zusatzwerkzeug(t, &zusatzprobe{}, "schreiben.zusatzwert")
	lesen, _ := zusatzwerkzeug(t, &zusatzprobe{}, "lesen.zusatzfelder")

	for _, fall := range []struct {
		name      string
		werkzeug  mcp.Werkzeug
		argumente string
		art       error
	}{
		{name: "kein JSON", werkzeug: setzen, argumente: `{`},
		{name: "kein JSON beim Lesen", werkzeug: lesen, argumente: `{`},
		{name: "ohne Karte", werkzeug: setzen, argumente: `{"setID": 7, "feld": "vermerk", "wert": "x"}`},
		{name: "ohne Karte beim Lesen", werkzeug: lesen, argumente: `{"setID": 7}`},
		{name: "ohne Feld", werkzeug: setzen, argumente: `{"setID": 7, "karteID": 1, "wert": "x"}`, art: domain.ErrValidation},
		{name: "ohne Wert", werkzeug: setzen, argumente: `{"setID": 7, "karteID": 1, "feld": "vermerk"}`, art: domain.ErrValidation},
	} {
		t.Run(fall.name, func(t *testing.T) {
			_, err := fall.werkzeug.Ruf(t.Context(), json.RawMessage(fall.argumente))
			if err == nil || (fall.art != nil && !errors.Is(err, fall.art)) {
				t.Errorf("Fehler %v, erwartet %v", err, fall.art)
			}
		})
	}

	boese := &zusatzprobe{fehler: domain.ErrNotFound}

	for _, name := range []string{"lesen.zusatzfelder", "schreiben.zusatzwert"} {
		werkzeug, _ := zusatzwerkzeug(t, boese, name)
		if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(
			`{"setID": 7, "karteID": 1, "feld": "vermerk", "wert": "x"}`)); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("%s: Fehler %v", name, err)
		}
	}

	// **Lesen gelingt, Setzen scheitert** — der Fehler kommt an.
	nurLesen, _ := zusatzwerkzeug(t, &zusatzprobe{setzfehler: domain.ErrConflict}, "schreiben.zusatzwert")
	if _, err := nurLesen.Ruf(t.Context(), json.RawMessage(
		`{"setID": 7, "karteID": 1, "feld": "vermerk", "wert": "x"}`)); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler beim Setzen: %v", err)
	}

	if _, da := zusatzwerkzeug(t, nil, "lesen.zusatzfelder"); da {
		t.Error("ohne Naht gibt es lesen.zusatzfelder trotzdem")
	}
}
