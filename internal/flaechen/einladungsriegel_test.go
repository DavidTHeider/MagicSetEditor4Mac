// **Ein Token nimmt keine Einladung an, die hinauswählt** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Der Angriff
//
// Eine Einladung ist **selbst unterschrieben**: wer einen Schlüssel hat,
// stellt eine aus, und die Adresse darin wählt er selbst. Nimmt die
// Anwendung sie an, **meldet** sie sich dort (NET-01b) — sie baut eine
// TCP-Verbindung und einen TLS-Handschlag zu der Adresse auf, die in der
// Einladung steht. Ein Automat mit einem Token `schreiben` konnte das
// auslösen, weil die Annahme nur `lokalSchreiben` fragte.
//
// **Die Meldung schickt den eigenen Namen und die eigene Adresse hinaus** —
// das ist `externSenden`, und ein Token reicht nie dorthin
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).
//
// # Woran gemessen
//
// **An einem Lauscher, der Verbindungen zählt**, und daran, dass kein
// Kontakt entstand — die Annahme ist ganz oder gar nicht. Die Gegenprobe:
// derselbe Mensch nimmt dieselbe Einladung ohne Token an, und der Lauscher
// sieht die Meldung.
package flaechen_test

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestEinTokenNimmtKeineEinladungAnDieHinauswaehlt(t *testing.T) {
	var gewaehlt atomic.Int64

	lauscher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = lauscher.Close() }) //nolint:errcheck // der Lauf endet

	go func() {
		for {
			verbindung, angenommen := lauscher.Accept()
			if angenommen != nil {
				return
			}

			gewaehlt.Add(1)
			_ = verbindung.Close() //nolint:errcheck // gezählt ist gezählt
		}
	}()

	fremd, bert := stelleAuf(t), stelleAuf(t)

	code := zeilenwert(t, mussKommandozeile(t, fremd.umgebung,
		"einladung-erstellen", "Fremd", "handel", "--adresse", lauscher.Addr().String()), "code")

	gezeigt := profilkennungAus(t, mussKommandozeile(t, bert.umgebung, "profil", "--json"))
	automat := strings.TrimSpace(mitEingabe(t, bert.umgebung,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"schreiben"}`,
		"token-anlegen", gezeigt))

	an := starteDienst(t, bert.umgebung)

	annahme, fehler := json.Marshal(map[string]any{
		"code": code, "name": "Fremd", "meinName": "Bert",
	})
	if fehler != nil {
		t.Fatal(fehler)
	}

	status, antwort := an.mitNachweis(automat).ruf(t, http.MethodPost,
		"/peer/einladungen/annahme", string(annahme))
	if status != http.StatusForbidden {
		t.Errorf("Status %d statt 403: %s", status, antwort)
	}

	if n := gewaehlt.Load(); n != 0 {
		t.Errorf("die Annahme mit Token wählte %d-mal die Adresse aus der Einladung", n)
	}

	if kontakte := mussKommandozeile(t, bert.umgebung, "kontakte"); strings.Contains(kontakte, fremd.kennung) {
		t.Errorf("die abgewiesene Annahme legte trotzdem einen Kontakt an:\n%s", kontakte)
	}

	// **Die Gegenprobe**: ohne Token — der Mensch über die Prozessmarke —
	// meldet sich die Annahme dort, und der Lauscher sieht es.
	mussREST(t, an, http.MethodPost, "/peer/einladungen/annahme", string(annahme))

	if gewaehlt.Load() == 0 {
		t.Error("auch ohne Token wählte niemand — der Lauscher misst nichts")
	}
}
