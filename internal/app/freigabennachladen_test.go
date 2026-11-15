//go:build darwin

// **Ein Nachladefehler ist keine gescheiterte Freigabe** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ohne Kunstgriff, an einem echten Bestand**: wer Verwalter an einem
// fremden Profil ist und sich selbst zum Leser herabstuft oder seine
// Freigabe löscht, darf das — und darf die Liste danach nicht mehr lesen
// (`lesbarMit`). Bis #1970 kam das Nein des Nachlesens als `fehler` zurück:
// „dafür braucht es die Rolle", während die Änderung längst galt.
package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// alsBob fährt einen Ruf als Bob und kehrt danach zurück.
func alsBob(
	t *testing.T, anwendung *app.App, bob, ich string, ruf func(context.Context),
) {
	t.Helper()

	ctx := t.Context()

	if zu := anwendung.NutzerWechseln(ctx, bob, ""); zu.Fehler != nil {
		t.Fatalf("zu Bob: %+v", zu.Fehler)
	}

	ruf(ctx)

	if zurueck := anwendung.NutzerWechseln(ctx, ich, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}
}

// freigabeNachgeladen prüft eine Antwort, deren Änderung galt und deren
// Liste sich danach nicht lesen ließ.
func freigabeNachgeladen(t *testing.T, schritt string, antwort app.FreigabenAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil || antwort.Nachladefehler.Code != app.CodeErlaubnis:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Freigaben == nil:
		t.Fatalf("%s: die Liste ist nil", schritt)
	}
}

func TestEinNachladefehlerIstKeineGescheiterteFreigabe(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	bob := anwendung.NutzerAnlegen(ctx, "Bob", "")

	if ich.Fehler != nil || bob.Fehler != nil {
		t.Fatalf("anlegen: %+v %+v", ich.Fehler, bob.Fehler)
	}

	bobsProfil := bob.Nutzer.Profile[0].ID
	verwalter := app.FreigabeangabenDTO{
		ProfilID: bobsProfil, NutzerID: ich.NutzerID, Rolle: string(domain.RolleVerwalter),
	}

	alsBob(t, anwendung, bob.Nutzer.ID, ich.NutzerID, func(ctx context.Context) {
		if gesetzt := anwendung.FreigabeSetzen(ctx, verwalter); gesetzt.Fehler != nil {
			t.Fatalf("Bob macht mich zum Verwalter: %+v", gesetzt.Fehler)
		}
	})

	// **Setzen**: herabgestuft ist herabgestuft.
	leser := verwalter
	leser.Rolle = string(domain.RolleLeser)

	freigabeNachgeladen(t, "setzen", anwendung.FreigabeSetzen(ctx, leser))

	alsBob(t, anwendung, bob.Nutzer.ID, ich.NutzerID, func(ctx context.Context) {
		stand := anwendung.Freigaben(ctx, bobsProfil)
		if stand.Fehler != nil || len(stand.Freigaben) != 1 ||
			stand.Freigaben[0].Rolle != string(domain.RolleLeser) {
			t.Fatalf("nach dem Herabstufen: %+v", stand)
		}

		if gesetzt := anwendung.FreigabeSetzen(ctx, verwalter); gesetzt.Fehler != nil {
			t.Fatalf("wieder Verwalter: %+v", gesetzt.Fehler)
		}
	})

	// **Löschen**: wer seine eigene Freigabe löscht, ist sie los.
	freigabeNachgeladen(t, "löschen",
		anwendung.FreigabeLoeschen(ctx, bobsProfil, ich.NutzerID))

	alsBob(t, anwendung, bob.Nutzer.ID, ich.NutzerID, func(ctx context.Context) {
		if stand := anwendung.Freigaben(ctx, bobsProfil); stand.Fehler != nil ||
			len(stand.Freigaben) != 0 {
			t.Fatalf("nach dem Löschen: %+v", stand)
		}
	})

	// **Ein gescheiterter Schritt bleibt einer**: ohne Rolle setzt niemand
	// etwas, und das Nein steht in `fehler`.
	if antwort := anwendung.FreigabeSetzen(ctx, verwalter); antwort.Fehler == nil ||
		antwort.Nachladefehler != nil {
		t.Fatalf("ohne Rolle gesetzt: %+v", antwort)
	}
}
