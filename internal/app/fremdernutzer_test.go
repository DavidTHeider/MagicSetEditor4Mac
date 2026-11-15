// **Über einen fremden Menschen schreibt nur, wer ihn werden kann** — ID-03
// (#1702).
//
// # Der Befund
//
// **Die Matrix sagt es seit 0078 für jede Fähigkeit am Träger Nutzer**:
// „wer arbeitet, darf über sich schreiben; fremde Nutzer verlangen deren
// zweiten Faktor" (0074). **Kein Dienst hielt es.** `FaktorEntfernen` nahm
// jedem Nutzer seinen Faktor, gleich wer arbeitete — und danach war der
// Wechsel zu ihm frei. Der Faktor schützte damit vor niemandem, der wusste,
// dass es diesen Knopf gibt.
//
// # Die Regel
//
// **Wer über einen anderen Menschen schreibt, muss ihn werden können, ohne
// dass dessen Faktor fragt** — sonst wechselt er zuerst zu ihm, und dort
// fragt der Faktor. Für einen Menschen ohne Faktor ändert sich nichts: ihn
// schützt nichts (0074, „Was das nicht löst"), und wer ihn verwalten will,
// könnte ihn ebenso gut werden.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// geschuetzterZweiter legt einen zweiten Nutzer an, der sich einen Faktor
// einrichtet, und kehrt zum ersten zurück.
func geschuetzterZweiter(t *testing.T, anwendung *app.App) *app.NutzerDTO {
	t.Helper()

	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	if ich.Fehler != nil {
		t.Fatalf("wer arbeitet: %+v", ich.Fehler)
	}

	bob := anwendung.NutzerAnlegen(ctx, "Bob", "")
	if bob.Fehler != nil {
		t.Fatalf("anlegen: %+v", bob.Fehler)
	}

	// **Bob richtet ihn selbst ein** — als Bob, denn über sich darf jeder.
	if zu := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); zu.Fehler != nil {
		t.Fatalf("zu Bob: %+v", zu.Fehler)
	}

	eingerichtet := anwendung.FaktorEinrichten(ctx, bob.Nutzer.ID)
	if eingerichtet.Fehler != nil {
		t.Fatalf("einrichten: %+v", eingerichtet.Fehler)
	}

	if bestaetigt := anwendung.FaktorBestaetigen(ctx, bob.Nutzer.ID,
		faktorcode(t, eingerichtet.Geheimnis)); bestaetigt.Fehler != nil {
		t.Fatalf("bestätigen: %+v", bestaetigt.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	return bob.Nutzer
}

func TestUeberEinenGeschuetztenNutzerSchreibtNurEr(t *testing.T) {
	anwendung := faktoranwendung(t)
	ctx := t.Context()
	bob := geschuetzterZweiter(t, anwendung)
	profil := profilAusUebersicht(t, anwendung, bob.Profile[0].ID)

	faelle := map[string]*app.Fehler{
		"Faktor entfernen":  anwendung.FaktorEntfernen(ctx, bob.ID, "totp").Fehler,
		"Faktor einrichten": anwendung.FaktorEinrichten(ctx, bob.ID).Fehler,
		"Nutzer löschen":    anwendung.NutzerLoeschen(ctx, bob.ID, bob.Name).Fehler,
		"Profil löschen":    anwendung.ProfilLoeschen(ctx, profil.ID, profil.Name).Fehler,
		"Nutzer umbenennen": anwendung.NutzerAendern(ctx, app.NutzerangabenDTO{
			ID: bob.ID, Name: "Übernommen", Fassung: bob.Fassung,
		}).Fehler,
		"Profiltext": anwendung.ProfilAendern(ctx, app.ProfilangabenDTO{
			ID: profil.ID, Name: profil.Name, Kurzebeschreibung: "Übernommen", Fassung: profil.Fassung,
		}).Fehler,
	}

	for name, fehler := range faelle {
		if fehler == nil {
			t.Errorf("%s: über Bob geschrieben, ohne Bob zu sein", name)

			continue
		}

		if fehler.Code != app.CodeErlaubnis {
			t.Errorf("%s: Code %q, erwartet %q — %s", name, fehler.Code, app.CodeErlaubnis, fehler.Meldung)
		}
	}

	// **An der Wirkung** (AK-04): der Faktor steht, und der Wechsel zu Bob
	// fragt weiter nach ihm.
	if faktoren := anwendung.Faktoren(ctx, bob.ID); len(faktoren.Faktoren) != 1 {
		t.Errorf("Bobs Faktoren: %+v", faktoren)
	}

	if zu := anwendung.NutzerWechseln(ctx, bob.ID, ""); zu.Fehler == nil {
		t.Error("der Wechsel zu Bob ging ohne seinen Faktor durch")
	}

	if nachher := profilAusUebersicht(t, anwendung, profil.ID); nachher.Kurzebeschreibung != profil.Kurzebeschreibung ||
		nachher.Fassung != profil.Fassung {
		t.Errorf("Bobs Profil geändert: %+v", nachher)
	}
}

// profilAusUebersicht liest ein Profil aus der Nutzerübersicht.
//
// **Und nicht aus der Antwort des Wechsels oder `GezeigtesProfil`**: beide
// tragen nur Kennung und Name. Ohne Fassung ließe sich der Profiltext gar nicht
// erst ändern, und ohne den Text prüfte der Vergleich danach einen leeren
// gegen einen leeren — eine Zusage, die nie scheitern kann (§10).
func profilAusUebersicht(t *testing.T, anwendung *app.App, profilID string) *app.ProfilDTO {
	t.Helper()

	uebersicht := anwendung.Nutzer(t.Context())
	if uebersicht.Fehler != nil {
		t.Fatalf("Übersicht: %+v", uebersicht.Fehler)
	}

	for _, nutzer := range uebersicht.Nutzer {
		for _, profil := range nutzer.Profile {
			if profil.ID == profilID {
				return &profil
			}
		}
	}

	t.Fatalf("das Profil %s steht nicht in der Übersicht", profilID)

	return nil
}
