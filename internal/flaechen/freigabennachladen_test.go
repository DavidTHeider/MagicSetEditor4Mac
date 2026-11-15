package flaechen_test

// **Wer sich selbst herabstuft, hat sich herabgestuft — an jeder Fläche** —
// §7, Regel 3 (#1970).
//
// # Was hier gemessen wird
//
// Ein Verwalter an einem fremden Profil stuft sich selbst zum Leser herab
// oder löscht seine Freigabe. **Die Änderung gilt, und die Liste danach darf
// er nicht mehr lesen.** Kommandozeile, Dienst und MCP sagen dann dasselbe:
// geschehen, und warum nichts nachkommt — Rückgabe 0 samt Hinweis, 200 samt
// `nachladefehler`, dasselbe Feld am Werkzeug.

import (
	"net/http"
	"strings"
	"testing"
)

// zumVerwalter macht `ich` zum Verwalter am Profil des Fremden — als er.
func zumVerwalter(t *testing.T, umgebung []string, fremder, fremdes, ich string) {
	t.Helper()

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "verwalter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
}

func TestEineSelbstHerabgestufteFreigabeGiltAnJederFlaeche(t *testing.T) {
	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	ich := eignerkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	fremder, fremdes := fremdesProfilAnDerZeile(t, umgebung)

	// **Die Kommandozeile: 0 und ein Satz auf dem Fehlerstrom.**
	zumVerwalter(t, umgebung, fremder, fremdes, ich)

	if aus, fehlerstrom, code := kommandozeile(t, umgebung,
		"freigabe-setzen", fremdes, ich, "leser"); code != 0 ||
		!strings.Contains(fehlerstrom, "geschrieben") {
		t.Fatalf("Kommandozeile: %d, %q, %q", code, aus, fehlerstrom)
	}

	// **Mit `--json` steht er im Rumpf.**
	zumVerwalter(t, umgebung, fremder, fremdes, ich)

	aus, _, code := kommandozeile(t, umgebung,
		"freigabe-loeschen", fremdes, ich, "--json")
	zeile := nachgeladenAus(t, aus)

	if code != 0 || zeile.Nachladefehler == nil || zeile.Nachladefehler.Code != "erlaubnis" {
		t.Fatalf("Kommandozeile mit --json: %d, %s", code, aus)
	}

	// **Der Dienst: 200, und der Rumpf sagt dasselbe.**
	zumVerwalter(t, umgebung, fremder, fremdes, ich)

	status, rumpf := an.ruf(t, http.MethodPut,
		"/profile/"+fremdes+"/freigaben/"+ich, `{"rolle":"leser"}`)
	if dienst := nachgeladenAus(t, rumpf); status != http.StatusOK ||
		dienst.Nachladefehler == nil || *dienst.Nachladefehler != *zeile.Nachladefehler {
		t.Fatalf("Dienst: %d, %s", status, rumpf)
	}

	// **Und das Werkzeug.**
	zumVerwalter(t, umgebung, fremder, fremdes, ich)

	if werkzeug := nachgeladenAus(t, mussMCP(t, umgebung,
		"schreiben.freigabe.loeschen", map[string]any{"profilID": fremdes, "nutzerID": ich},
	)); werkzeug.Nachladefehler == nil || *werkzeug.Nachladefehler != *zeile.Nachladefehler {
		t.Fatalf("MCP: %+v", werkzeug)
	}

	// **Geschehen ist alles** — der Fremde sieht keine Freigabe mehr.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	if liste := mussKommandozeile(t, umgebung, "freigaben", fremdes); strings.Contains(liste, ich+"\t") {
		t.Errorf("die gelöschte Freigabe steht noch da: %q", liste)
	}
}
