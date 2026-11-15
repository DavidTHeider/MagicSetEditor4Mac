//go:build darwin

// **Der Arbeitskontext am gestarteten Fenster** — Werkstück O1 von Phase 5.2
// (#1425).
//
// **Darwin, weil `alsLeserHinein` den Schlüsselbund braucht** — derselbe
// Grund wie in `nurlesen_test.go`.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Person A im Profil von B bleibt A** — und der Kontext sagt alles in
// einem Zug: wer, wo, wessen Profil, welche Rolle, ob nur gelesen wird.
func TestDerArbeitskontextSagtWerWoAlsWas(t *testing.T) {
	anwendung := alsLeserHinein(t)

	kontext := anwendung.Arbeitskontext(t.Context())
	if kontext.Fehler != nil {
		t.Fatalf("Arbeitskontext: %+v", kontext.Fehler)
	}

	gezeigt := anwendung.GezeigtesProfil(t.Context())
	if gezeigt.Fehler != nil {
		t.Fatalf("gezeigtes Profil: %+v", gezeigt.Fehler)
	}

	if kontext.ProfilID != gezeigt.Profil.ID || kontext.EignerID != gezeigt.Profil.NutzerID {
		t.Errorf("Profil %s (Eigner %s), gezeigt wird %s (Eigner %s)",
			kontext.ProfilID, kontext.EignerID, gezeigt.Profil.ID, gezeigt.Profil.NutzerID)
	}

	if kontext.NutzerID == kontext.EignerID {
		t.Error("der Gast wurde zum Eigner")
	}

	if kontext.Rolle != string(domain.RolleLeser) || !kontext.NurLesend {
		t.Errorf("Rolle %q, nur lesend %v — erwartet leser, nur lesend", kontext.Rolle, kontext.NurLesend)
	}

	if kontext.Modus != string(domain.ModusFenster) {
		t.Errorf("Modus %q, erwartet %q", kontext.Modus, domain.ModusFenster)
	}

	// **Und der Weg hinaus ändert die Revision** — sonst könnte das Fenster
	// nicht sehen, dass sein Kontext ein anderer ist.
	eigenes := anwendung.Nutzer(t.Context())
	if eigenes.Fehler != nil {
		t.Fatalf("Nutzer: %+v", eigenes.Fehler)
	}

	var meinProfil string

	for _, nutzer := range eigenes.Nutzer {
		if nutzer.ID == kontext.NutzerID {
			meinProfil = nutzer.Profile[0].ID
		}
	}

	if zurueck := anwendung.ProfilWechseln(t.Context(), meinProfil, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück ins eigene Profil: %+v", zurueck.Fehler)
	}

	danach := anwendung.Arbeitskontext(t.Context())
	if danach.Fehler != nil {
		t.Fatalf("Arbeitskontext danach: %+v", danach.Fehler)
	}

	if danach.Revision <= kontext.Revision {
		t.Errorf("Revision %d nach dem Wechsel, vorher %d", danach.Revision, kontext.Revision)
	}

	if danach.NurLesend || danach.Rolle != string(domain.RolleVerwalter) {
		t.Errorf("im eigenen Profil: %+v — erwartet der Verwalter", danach)
	}

	// **Und die Bindung nimmt denselben Weg.**
	if gebunden := app.NeueBindungen(anwendung).Arbeitskontext(t.Context()); gebunden.ProfilID != danach.ProfilID {
		t.Errorf("die Bindung sagt %+v", gebunden)
	}
}
