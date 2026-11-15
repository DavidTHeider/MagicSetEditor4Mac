//go:build darwin

// **Ein Entzug wirkt beim nächsten Aufruf** — ID-03.AK-02
// (#1702).
//
// # Was hier gecacht war
//
// **Die Verbindung.** Ob das Fenster über die schreibende oder die nur
// lesende arbeitet, entschied sich beim Wechsel (`profilUmstellen`) und
// danach nie wieder. Zieht jemand die Freigabe zurück, während man im Profil
// steht — aus einem zweiten Prozess, über die Kommandozeile oder den Dienst —,
// schrieb das Fenster mit der Rolle von vorher weiter.
//
// # Wie der Entzug hier geschieht
//
// **Über eine zweite Verbindung auf dieselbe Datei**, wie ein zweiter
// Prozess sie hätte. Der Zeiger bleibt dabei, wo er ist: wer über einen
// zweiten Prozess **als jemand anderes** arbeitete, bewegte den Zeiger im
// Bestand, und dann wechselte das Fenster ohnehin.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinEntzugWirktBeimNaechstenAufruf(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	bob := anwendung.NutzerAnlegen(ctx, "Bob", "")

	if ich.Fehler != nil || bob.Fehler != nil {
		t.Fatalf("anlegen: %+v %+v", ich.Fehler, bob.Fehler)
	}

	bobsProfil := bob.Nutzer.Profile[0].ID

	// **Bob gibt mir die Rolle Mitarbeiter** — als Bob, denn nur der Eigner
	// verwaltet seine Freigaben.
	if zu := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); zu.Fehler != nil {
		t.Fatalf("zu Bob: %+v", zu.Fehler)
	}

	if gesetzt := anwendung.FreigabeSetzen(ctx, app.FreigabeangabenDTO{
		ProfilID: bobsProfil, NutzerID: ich.NutzerID, Rolle: string(domain.RolleMitarbeiter),
	}); gesetzt.Fehler != nil {
		t.Fatalf("Freigabe: %+v", gesetzt.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	if hinein := anwendung.ProfilWechseln(ctx, bobsProfil, ""); hinein.Fehler != nil {
		t.Fatalf("hinein: %+v", hinein.Fehler)
	}

	// **Vorher darf ich** — sonst bewiese der Lauf nichts über den Entzug.
	if vorher := anwendung.DeckAnlegen(ctx, "Vorher", "magic", ""); vorher.Fehler != nil {
		t.Fatalf("als Mitarbeiter: %+v", vorher.Fehler)
	}

	entziehen(t, bestand, bobsProfil, ich.NutzerID)

	nachher := anwendung.DeckAnlegen(ctx, "Nachher", "magic", "")
	if nachher.Fehler == nil {
		t.Error("nach dem Entzug angelegt — mit der Rolle von vorher")
	} else if nachher.Fehler.Code != app.CodeNurLesend {
		t.Errorf("Code %q, erwartet %q — %s", nachher.Fehler.Code, app.CodeNurLesend, nachher.Fehler.Meldung)
	}

	// **An der Wirkung** (AK-04): genau das Deck von vorher steht da.
	if decks := anwendung.Decks(ctx, ""); decks.Fehler != nil || len(decks.Decks) != 1 {
		t.Errorf("Decks nach dem Entzug: %+v", decks)
	}

	// **Und das Banner sagt es auch.**
	if kontext := anwendung.Arbeitskontext(ctx); !kontext.NurLesend {
		t.Errorf("Arbeitskontext nach dem Entzug: %+v", kontext)
	}
}

// entziehen nimmt eine Freigabe über eine zweite Verbindung zurück.
func entziehen(t *testing.T, bestand, profilID, nutzerID string) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestand)
	if fehler != nil {
		t.Fatalf("zweite Verbindung: %v", fehler)
	}

	defer func() {
		if schliessfehler := db.Close(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	}()

	profil, nutzer := zahlAus(t, profilID), zahlAus(t, nutzerID)
	if loeschfehler := storage.NeueFreigabenablage(db).FreigabeLoeschen(t.Context(), profil, nutzer); loeschfehler != nil {
		t.Fatalf("entziehen: %v", loeschfehler)
	}
}

// **Ohne lesbare Revision arbeitet das Fenster weiter** — beim Start und bei
// jedem Ruf. Der Bestand, der klemmt, sagt es an seiner eigenen Stelle
// deutlicher als eine Anwendung, die deswegen gar nicht erst aufgeht.
//
// **Umbenannt und nicht gelöscht**: die Auslöser an Zeigern und Freigaben
// schreiben in die Tafel, und SQLite zieht ihren Namen dort mit. Gelöscht
// scheiterte danach jeder Wechsel — ein anderer Fall als dieser.
func TestOhneLesbareRevisionArbeitetDasFensterWeiter(t *testing.T) {
	erste, bestand := gestarteteAnwendung(t)

	db, fehler := storage.Oeffnen(t.Context(), bestand)
	if fehler != nil {
		t.Fatalf("zweite Verbindung: %v", fehler)
	}

	if _, umbenennfehler := db.ExecContext(t.Context(),
		`ALTER TABLE freigabenrevision RENAME TO freigabenrevision_weg`); umbenennfehler != nil {
		t.Fatalf("umbenennen: %v", umbenennfehler)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	// **Beim Ruf**: die erste Anwendung liest die Revision und findet sie nicht.
	if decks := erste.Decks(t.Context(), ""); decks.Fehler != nil {
		t.Errorf("ein Ruf ohne lesbare Revision: %+v", decks.Fehler)
	}

	// **Beim Start**: eine zweite richtet ihre Verbindung ohne sie.
	zweite := app.NeueApp(bestand)
	if startfehler := zweite.Starten(t.Context()); startfehler != nil {
		t.Fatalf("starten ohne lesbare Revision: %v", startfehler)
	}

	t.Cleanup(func() {
		if beendefehler := zweite.Beenden(); beendefehler != nil {
			t.Errorf("beenden: %v", beendefehler)
		}
	})

	if decks := zweite.Decks(t.Context(), ""); decks.Fehler != nil {
		t.Errorf("nach dem Start ohne lesbare Revision: %+v", decks.Fehler)
	}
}
