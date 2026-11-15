package app

// **Der freie Wert einer Panik bleibt am Rechner** — #2399, Vermerk 0193:
// im Protokoll dieses Rechners steht er ganz, zur Diagnose (F02, F03); was
// davon in die Diagnose und das Diagnosepaket geht, hat die Gestalt, die das
// Laufzeitsystem schreibt — dieselbe Zusage wie für die Absturzspur.

import (
	"strings"
	"testing"
)

func TestPaniktexteVerlassenDenRechnerNurInDerGestaltDerLaufzeit(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		zeile, bleibt, geht string
	}{
		"Merkmal, frei": {
			`level=ERROR msg="ein Schritt brach mit einer Panik ab" panik="Kennwort hunter2 falsch" kennung=k1`,
			`kennung=k1`, `hunter2`,
		},
		"Merkmal, ohne Anführung": {`level=ERROR msg=x panik=geheim kennung=k2`, `kennung=k2`, `geheim`},
		"Merkmal, Laufzeit": {
			`level=ERROR msg=x panik="runtime error: index out of range [3] with length 2" art=ruf`,
			`runtime error: index out of range [3] with length 2`, ``,
		},
		"Wails": {
			`level=ERROR msg="ein Ruf scheiterte" fehlerkennung=f1 grund="Bindungen.Sets: panic: Kunde Meier, Konto 4711" art=ruf`,
			`art=ruf`, `Meier`,
		},
		"Zusammenbruch": {
			`level=ERROR msg=y grund="ein Nebenlauf ist zusammengebrochen: Passphrase blau" vorgang=v`,
			`vorgang=v`, `blau`,
		},
		"panikFehler": {`level=ERROR msg=z grund="eine Panik: das Token xyz"`, `msg=z`, `xyz`},
		"Marke bleibt": {
			`level=ERROR panik="geheim [recovered]"`,
			Geschwaerzt + ` [recovered]`, `geheim`,
		},
		"ohne Panik": {`level=INFO msg="alles gut" grund="die Platte ist voll"`, `die Platte ist voll`, ``},
		// **Und im Format JSON** (Zweitblick über Phase 5, Befund 1): die
		// Weiche schreibt es, wenn `protokoll.format` es sagt, und dort heißt
		// das Merkmal `"panik":"…"`.
		"JSON, frei": {
			`{"time":"t","level":"ERROR","msg":"ein Schritt brach mit einer Panik ab","panik":"Kennwort hunter2 falsch","kennung":"k1"}`,
			`"kennung":"k1"`, `hunter2`,
		},
		"JSON, mit Anführung im Wert": {
			`{"level":"ERROR","msg":"x","panik":"das \"Geheimnis\" blau","art":"ruf"}`,
			`"art":"ruf"`, `blau`,
		},
		"JSON, Laufzeit": {
			`{"level":"ERROR","msg":"x","panik":"runtime error: integer divide by zero"}`,
			`runtime error: integer divide by zero`, ``,
		},
		"JSON, Wails": {
			`{"level":"ERROR","msg":"ein Ruf scheiterte","grund":"Bindungen.Sets: panic: Kunde Meier","art":"ruf"}`,
			`"art":"ruf"`, `Meier`,
		},
	} {
		sauber := paniktexteSchwaerzen(fall.zeile)
		if !strings.Contains(sauber, fall.bleibt) || fall.geht != "" && strings.Contains(sauber, fall.geht) {
			t.Errorf("%s: %s", name, sauber)
		}
	}
}
