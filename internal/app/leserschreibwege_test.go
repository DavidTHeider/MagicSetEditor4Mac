//go:build darwin

// **Was ein `leser` auch über die Nutzerverwaltung nicht schreiben darf** —
// der Nachlauf zum Zweitblick über Phase 4.12
// (#1351).
//
// # Der Befund
//
// **Die benannte Ausnahme war zu breit.** Werkstück 16 lässt Nutzer, Profile,
// Konten, Faktoren, Token, Freigaben und die Sicherung an der **schreibenden**
// Verbindung — mit der Begründung, sonst fände niemand aus einem fremden
// Profil wieder hinaus.
//
// **Diese Begründung trägt nicht so weit.** Der Weg hinaus schreibt den
// **Zeiger**; ein Zugriffstoken und eine Profilvorgabe schreiben in ein
// **Profil**, und wer dort nur lesen darf, darf beides nicht. Angestrichen vom
// Nachlauf; der Befund nannte MCP und den Dienst, und dieselbe Naht steht am
// Fenster.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein Leser legt kein Zugriffstoken im fremden Profil an.**
//
// **Sonst wäre der Nur-Lese-Bestand mit einem Handgriff zu umgehen**: ein
// Token trägt Rechte, und wer es anlegt, gibt sie sich selbst.
func TestEinLeserLegtKeinTokenAn(t *testing.T) {
	anwendung := alsLeserHinein(t)

	gezeigt := anwendung.GezeigtesProfil(t.Context())
	if gezeigt.Fehler != nil {
		t.Fatalf("gezeigtes Profil: %+v", gezeigt.Fehler)
	}

	angelegt := anwendung.TokenAnlegen(t.Context(), app.TokenangabenDTO{
		ProfilID: gezeigt.Profil.ID, Bezeichnung: "Probe", Principal: "Probe", Scope: "lesen",
	})

	if angelegt.Fehler == nil {
		t.Fatal("der Leser legte ein Token im fremden Profil an")
	}

	if angelegt.Fehler.Code != app.CodeNurLesend {
		t.Errorf("Code %q, erwartet war %q", angelegt.Fehler.Code, app.CodeNurLesend)
	}
}

// **Und er setzt keine Profilvorgabe.**
//
// **Sie gehört dem Profil und nicht dem Menschen**: sie sagt, womit ein neuer
// Datensatz **dort** anfängt.
func TestEinLeserSetztKeineProfilvorgabe(t *testing.T) {
	anwendung := alsLeserHinein(t)

	gezeigt := anwendung.GezeigtesProfil(t.Context())
	if gezeigt.Fehler != nil {
		t.Fatalf("gezeigtes Profil: %+v", gezeigt.Fehler)
	}

	gesetzt := anwendung.ProfilvorgabeSetzen(t.Context(), app.VorgabeangabenDTO{
		ProfilID: gezeigt.Profil.ID, Art: "format", Werte: []string{},
	})

	if gesetzt.Fehler == nil {
		t.Fatal("der Leser setzte eine Vorgabe im fremden Profil")
	}

	if gesetzt.Fehler.Code != app.CodeNurLesend {
		t.Errorf("Code %q, erwartet war %q", gesetzt.Fehler.Code, app.CodeNurLesend)
	}
}

// **Und die Wege, die es geben MUSS, bleiben offen** — die Gegenprobe.
//
// **Ohne sie sagte der Lauf oben nur, dass irgendetwas abweist** (§10): eine
// Anwendung, die einem Leser gar nichts mehr erlaubt, erfüllte ihn auch.
func TestEinLeserVerwaltetSichSelbstWeiter(t *testing.T) {
	anwendung := alsLeserHinein(t)

	ich := anwendung.WerArbeitet(t.Context())
	if ich.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", ich.Fehler)
	}

	// **Seine eigenen Einstellungen** — sie gehören dem Menschen und nicht
	// dem Profil, das er gerade ansieht.
	if gesetzt := anwendung.NutzereinstellungSetzen(
		t.Context(), ich.NutzerID, "barrierefreiheit.bewegung", "reduziert",
	); gesetzt.Fehler != nil {
		t.Errorf("der Leser konnte seine eigene Einstellung nicht ändern: %+v",
			gesetzt.Fehler)
	}

	// **Und der Weg hinaus** — sonst säße er fest.
	eigenes := anwendung.Nutzer(t.Context())
	if eigenes.Fehler != nil {
		t.Fatalf("Nutzer: %+v", eigenes.Fehler)
	}

	if hinaus := anwendung.ProfilWechseln(
		t.Context(), eigenes.Nutzer[0].Profile[0].ID, ""); hinaus.Fehler != nil {
		t.Errorf("der Leser kam nicht hinaus: %+v", hinaus.Fehler)
	}
}
