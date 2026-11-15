//go:build darwin

// **Ein `leser` liest und schreibt nicht** — Werkstück 16 von Phase 4.12
// (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// # Warum das an der **gestarteten** Anwendung geprüft wird
//
// **Weil die Zusage an der Verdrahtung hängt.** Ein Doppel prüfte die
// Entscheidung — hier wird geprüft, dass die Repositories wirklich an der
// zweiten Verbindung hängen und dass der Weg zurück offen bleibt (§10: zwei
// grüne Seiten einer Naht sind kein Beleg für die Naht).

package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// alsLeserHinein legt einen zweiten Menschen samt Profil an, gibt sich selbst
// darauf die Rolle `leser` und wechselt hinein.
//
// **Über die Bindungen und nicht über den Bestand**: der Weg dorthin gehört zu
// dem, was hier geprüft wird.
func alsLeserHinein(t *testing.T) *app.App {
	t.Helper()

	anwendung, _ := gestarteteAnwendung(t)

	ich := anwendung.WerArbeitet(t.Context())
	if ich.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", ich.Fehler)
	}

	zweiter := anwendung.NutzerAnlegen(t.Context(), "Zweiter", "")
	if zweiter.Fehler != nil {
		t.Fatalf("anlegen: %+v", zweiter.Fehler)
	}

	fremdesProfil := zweiter.Nutzer.Profile[0].ID

	// **Die Freigabe vergibt der Eigner** — dafür wird man er.
	if wechsel := anwendung.NutzerWechseln(
		t.Context(), zweiter.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("Nutzer wechseln: %+v", wechsel.Fehler)
	}

	if gesetzt := anwendung.FreigabeSetzen(t.Context(), app.FreigabeangabenDTO{
		ProfilID: fremdesProfil,
		NutzerID: ich.NutzerID,
		Rolle:    string(domain.RolleLeser),
	}); gesetzt.Fehler != nil {
		t.Fatalf("Freigabe setzen: %+v", gesetzt.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(
		t.Context(), ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	// **Und jetzt als Leser hinein.**
	if hinein := anwendung.ProfilWechseln(
		t.Context(), fremdesProfil, ""); hinein.Fehler != nil {
		t.Fatalf("als Leser hinein: %+v", hinein.Fehler)
	}

	return anwendung
}

// **Der Leser liest — und schreibt nicht.**
//
// **Beides in einem Lauf**, denn einzeln sagt keines etwas: eine Anwendung,
// die gar nichts kann, erfüllte die zweite Hälfte.
func TestEinLeserLiestUndSchreibtNicht(t *testing.T) {
	anwendung := alsLeserHinein(t)

	// **Lesen geht** — die Setliste kommt, wenn auch leer.
	if gelesen := anwendung.Sets(t.Context()); gelesen.Fehler != nil {
		t.Fatalf("der Leser konnte nicht lesen: %+v", gelesen.Fehler)
	}

	// **Und Schreiben nicht** — im Treiber, nicht an dieser Bindung.
	//
	// **Über eine Einstellung und nicht über ein Set**: `SetAnlegen` weist
	// vorher ab, weil ein Set einen Stil braucht — dann prüfte der Lauf die
	// Validierung statt der Verbindung. Eine Einstellung schreibt ohne
	// Vorbedingung.
	angelegt := anwendung.EinstellungSetzen(t.Context(), "sprache", "en")

	if angelegt.Fehler == nil {
		t.Fatal("der Leser schrieb eine Einstellung")
	}

	// **Und die Meldung sagt, was los ist** (§8): nicht „unerwarteter Fehler"
	// und nicht der rohe Treibertext.
	if angelegt.Fehler.Code != app.CodeNurLesend {
		t.Errorf("Code %q, erwartet war %q", angelegt.Fehler.Code, app.CodeNurLesend)
	}
}

// **Und der Weg hinaus bleibt offen.**
//
// # Warum das die Hälfte der Zusage ist
//
// **Weil der Wechsel den Zeiger schreibt.** Hinge die Nutzerverwaltung an der
// nur lesenden Verbindung, säße der Leser fest — in einem fremden Profil, ohne
// Weg zurück. Das ist die eine benannte Ausnahme, und hier wird sie eingelöst.
func TestEinLeserKommtWiederHinaus(t *testing.T) {
	anwendung := alsLeserHinein(t)

	eigenes := anwendung.Nutzer(t.Context())
	if eigenes.Fehler != nil {
		t.Fatalf("Nutzer: %+v", eigenes.Fehler)
	}

	meines := eigenes.Nutzer[0].Profile[0].ID

	if hinaus := anwendung.ProfilWechseln(
		t.Context(), meines, ""); hinaus.Fehler != nil {
		t.Fatalf("der Leser kam nicht hinaus: %+v", hinaus.Fehler)
	}

	// **Und danach schreibt er wieder** — die zweite Verbindung ist fort.
	//
	// **Ohne diese Zeile wäre der Lauf oben auch dann grün, wenn gar nichts
	// mehr schriebe** (§10: eine Gegenprobe, die nichts ändert, ist keine).
	if angelegt := anwendung.EinstellungSetzen(
		t.Context(), "sprache", "en"); angelegt.Fehler != nil {
		t.Errorf("im eigenen Profil ging das Schreiben nicht: %+v", angelegt.Fehler)
	}
}

// **Und ein Fehler beim Fragen hält den Start nicht auf.**
//
// **Was das kostet, gehört gesagt**: dann wird geschrieben, wo vielleicht
// nicht geschrieben werden sollte. Der Fehler steht im Protokoll — und eine
// Anwendung, die deswegen gar nicht aufgeht, tauscht einen kleinen Mangel
// gegen einen großen (§8).
func TestEinFehlerBeimFragenHaeltDenStartNichtAuf(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	anwendung.MitProfilwechsel(&wechseldoppel{fehler: errors.New("klemmt")})

	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("der Start scheiterte: %v", fehler)
	}
}
