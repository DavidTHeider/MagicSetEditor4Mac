package flaechen_test

// **Ein Nachladefehler ist an keiner Fläche ein gescheiterter Schritt** — §7,
// Regel 3 (#1960).
//
// # Was hier gemessen wird
//
// **Dass Kommandozeile, Dienst und MCP dasselbe sagen**, an einem echten
// Bestand: eine Zeile mit krummem Zeitpunkt lässt jede Änderung an einem
// anderen Kontakt glücken und das Lesen der Liste danach scheitern. Die
// Kommandozeile endet dann mit 0 und sagt auf dem Fehlerstrom, was nicht zu
// lesen war; der Dienst antwortet 200 und trägt `nachladefehler` im Rumpf,
// ebenso das Werkzeug, das an MCP sperren darf. **Keine meldet die Änderung
// als gescheitert** — ein Skript, das auf den Fehler hin wiederholte, liefe
// gegen den geschriebenen Stand.
//
// # Was hier **nicht** gemessen wird
//
// **Das Fenster** — es liest die Liste nach jedem geglückten Schritt selbst
// nach und hält sich nur an `fehler`; das hält `peerkontakte.test.tsx`.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kontaktzeileVerbiegen legt eine Zeile an, die sich nicht lesen lässt — oder
// nimmt sie wieder weg.
func kontaktzeileVerbiegen(t *testing.T, umgebung []string, anlegen bool) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	profil, profilfehler := storage.ProfilAus(ctx)
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	befehl := `DELETE FROM peerkontakte WHERE profil_id = ? AND kennung = 'bkrumm'`
	if anlegen {
		befehl = `INSERT INTO peerkontakte (profil_id, kennung, stand, gesehen)
VALUES (?, 'bkrumm', 'unbestaetigt', 'gestern')`
	}

	if _, fehler := db.ExecContext(ctx, befehl, profil); fehler != nil {
		t.Fatalf("die krumme Zeile: %v", fehler)
	}
}

// nachgeladen ist, was beide Flächen über einen Nachladefehler sagen.
type nachgeladen struct {
	Kontakte       []json.RawMessage `json:"kontakte"`
	Nachladefehler *struct {
		Code    string `json:"code"`
		Meldung string `json:"meldung"`
	} `json:"nachladefehler"`
	Fehler json.RawMessage `json:"fehler"`
}

// nachgeladenAus liest eine Antwort der Kommandozeile oder des Dienstes.
func nachgeladenAus(t *testing.T, roh string) nachgeladen {
	t.Helper()

	var gelesen nachgeladen
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("keine Kontaktantwort: %v\n%s", fehler, roh)
	}

	return gelesen
}

func TestEinNachladefehlerIstAnKeinerFlaecheEinGescheiterterSchritt(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Annas Mac")
	kontaktzeileVerbiegen(t, umgebung, true)

	// **Die Kommandozeile: getan, und gesagt, was fehlt.**
	if aus, fehlerstrom, code := kommandozeile(t, umgebung,
		"kontakt-notiz", kennung, "vom Stammtisch"); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") || strings.TrimSpace(aus) != "" {
		t.Fatalf("Kommandozeile: %d, %q, %q", code, aus, fehlerstrom)
	}

	// **Mit `--json` steht er im Rumpf** — und der Dienst sagt dasselbe.
	aus, _, code := kommandozeile(t, umgebung,
		"kontakt-adresse", kennung, "192.168.1.7:41234", "--json")
	zeile := nachgeladenAus(t, aus)

	if code != 0 || zeile.Nachladefehler == nil || string(zeile.Fehler) != "null" {
		t.Fatalf("Kommandozeile mit --json: %d, %s", code, aus)
	}

	// **Der Dienst: 200, und der Rumpf trägt ihn neben der leeren Liste.**
	status, rumpf := an.ruf(t, http.MethodPut, "/peer/kontakte/"+kennung+"/sperre", "")
	dienst := nachgeladenAus(t, rumpf)

	if status != http.StatusOK || dienst.Kontakte == nil || len(dienst.Kontakte) != 0 ||
		dienst.Nachladefehler == nil {
		t.Fatalf("Dienst: %d, %s", status, rumpf)
	}

	if *dienst.Nachladefehler != *zeile.Nachladefehler ||
		dienst.Nachladefehler.Code != "validierung" {
		t.Errorf("zwei Flächen, zwei Auskünfte: %+v gegen %+v",
			*dienst.Nachladefehler, *zeile.Nachladefehler)
	}

	// **Und das Werkzeug, das sperren darf** — das einzige schreibende über
	// die Kontakte an MCP. Eine leere Liste ohne Wort hielte ein Modell für
	// „niemand mehr bekannt".
	werkzeug := nachgeladenAus(t, mussMCP(t, umgebung,
		"schreiben.peerkontakt.sperren", map[string]any{"kennung": kennung}))

	if werkzeug.Nachladefehler == nil || *werkzeug.Nachladefehler != *zeile.Nachladefehler {
		t.Errorf("MCP sagt etwas anderes: %+v", werkzeug)
	}

	// **Und geschrieben ist alles drei** — gelesen, sobald die Liste es wieder
	// hergibt.
	kontaktzeileVerbiegen(t, umgebung, false)

	aus = mussKommandozeile(t, umgebung, "kontakte")
	for _, erwartet := range []string{
		"gesperrt\t" + kennung, "unter\t192.168.1.7:41234", "notiz\tvom Stammtisch",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("nach dem Nachladefehler fehlt %q in %q", erwartet, aus)
		}
	}
}
