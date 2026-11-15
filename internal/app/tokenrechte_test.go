//go:build darwin

// **Token und Vorgaben eines Profils führt nur, wer dort mitarbeitet** —
// #1817.
//
// # Der Befund
//
// Nachgestellt am 2026-09-23: aus dem **eigenen** Profil heraus legte ein
// Nutzer ein schreibendes Zugriffstoken für das Profil eines anderen an und
// bekam den Klartext; Rotieren und Sammelwiderruf gingen ebenso. Die Matrix
// verlangt „Mitarbeiter" — durchgesetzt über `query_only`, und das gehört zum
// **gezeigten** Profil. Die Bindungen nehmen aber die Kennung des Zielprofils
// vom Rufer.
//
// # Was er zeigt
//
// Jede Bindung, die eine Profil- oder Tokenkennung annimmt, ist gegen ein
// fremdes Profil abgewiesen — **an der Wirkung** geprüft: Bobs Token und
// Vorgaben stehen danach wie vorher. Und im eigenen Profil geht es weiter.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestTokenUndVorgabenBrauchenDieRolleAmZielprofil(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	eigenes := anwendung.GezeigtesProfil(ctx)

	bob := anwendung.NutzerAnlegen(ctx, "Audit Bob", "")
	if bob.Fehler != nil || ich.Fehler != nil || eigenes.Fehler != nil {
		t.Fatalf("vorbereiten: %+v %+v %+v", bob.Fehler, ich.Fehler, eigenes.Fehler)
	}

	fremd := bob.Nutzer.Profile[0].ID

	// **Bob legt selbst ein Token an** — damit es etwas zu rotieren, zu
	// ändern und zu löschen gibt.
	if wechsel := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("zu Bob: %+v", wechsel.Fehler)
	}

	bobs := anwendung.TokenAnlegen(ctx, app.TokenangabenDTO{
		ProfilID: fremd, Bezeichnung: "Bobs Dienst", Principal: "bob", Scope: "lesen",
	})
	if bobs.Fehler != nil {
		t.Fatalf("Bobs Token: %+v", bobs.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	faelle := map[string]*app.Fehler{
		"anlegen": anwendung.TokenAnlegen(ctx, app.TokenangabenDTO{
			ProfilID: fremd, Bezeichnung: "Fremdzugang", Principal: "angreifer", Scope: "schreiben",
		}).Fehler,
		"lesen": anwendung.Token(ctx, fremd).Fehler,
		"ändern": anwendung.TokenAendern(ctx, app.TokenangabenDTO{
			ID: bobs.Token.ID, Bezeichnung: "Bobs Dienst", Principal: "bob", Scope: "schreiben",
			Fassung: bobs.Token.Fassung,
		}).Fehler,
		"rotieren":                     anwendung.TokenRotieren(ctx, bobs.Token.ID).Fehler,
		"ein Token, das es nicht gibt": anwendung.TokenRotieren(ctx, "999").Fehler,
		"löschen":                      anwendung.TokenLoeschen(ctx, bobs.Token.ID).Fehler,
		"alle widerrufen":              anwendung.TokenAlleWiderrufen(ctx, fremd).Fehler,
		"Vorgabe setzen": anwendung.ProfilvorgabeSetzen(ctx, app.VorgabeangabenDTO{
			ProfilID: fremd, Art: "lagerort", Werte: []string{"Fremdes Regal"},
		}).Fehler,
		"Vorgaben lesen":  anwendung.Profilvorgaben(ctx, fremd).Fehler,
		"Freigaben lesen": anwendung.Freigaben(ctx, fremd).Fehler,
	}

	for name, fehler := range faelle {
		if fehler == nil {
			t.Errorf("%s: an einem fremden Profil angenommen", name)
		}
	}

	// **An der Wirkung**: bei Bob steht sein eines Token wie vorher.
	if wechsel := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("zu Bob: %+v", wechsel.Fehler)
	}

	liste := anwendung.Token(ctx, fremd)
	if liste.Fehler != nil || len(liste.Token) != 1 || liste.Token[0].Scope != "lesen" ||
		liste.Token[0].Fassung != bobs.Token.Fassung {
		t.Errorf("Bobs Token danach: %+v", liste)
	}

	if vorgaben := anwendung.Profilvorgaben(ctx, fremd); vorgaben.Fehler != nil {
		t.Errorf("Bob liest seine Vorgaben nicht: %+v", vorgaben.Fehler)
	} else {
		for _, stand := range vorgaben.Vorgaben {
			if stand.Art == "lagerort" && len(stand.Werte) != 0 {
				t.Errorf("Bobs Lagerorte danach: %v", stand.Werte)
			}
		}
	}

	// **Und im eigenen Profil geht es weiter** — sonst bewiese der Lauf eine
	// Anwendung, die gar kein Token mehr anlegt.
	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	if eigenes := anwendung.TokenAnlegen(ctx, app.TokenangabenDTO{
		ProfilID: eigenes.Profil.ID, Bezeichnung: "Eigener Dienst", Principal: "ich", Scope: "lesen",
	}); eigenes.Fehler != nil {
		t.Errorf("im eigenen Profil: %+v", eigenes.Fehler)
	}
}
