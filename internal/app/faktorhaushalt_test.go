package app_test

// **Ein Code gilt einmal, und Fehlversuche haben ein Budget — durch die
// Wurzel** — SEC-09 und SEC-10
// (#2150,
// #2151,
// Vermerk 0178).
//
// Echter Bestand, eigener Probebund, kein Sensor. Die gesteuerte Uhr und zwei
// Vorräte auf einer Datei stehen eine Schicht tiefer
// (`internal/storage/faktorhaushalt_intern_test.go`); hier zählt, dass die
// Wurzel denselben Haushalt verdrahtet, den die Ablage hält.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDieWurzelVerbrauchtDenCodeUndZaehltDieFehlversuche(t *testing.T) {
	t.Setenv(schluesselbund.Bundort, probebundOrt(t))

	flaeche := app.NutzerflaecheOhneSensor(offeneProbedatenbank(t), &probesensor{}, keineBiometrie)
	ctx := probekontext(t)
	kennung := "haushalt@example.org"

	angelegt := flaeche.KontoEintragen(ctx, kennung, probekennwort, "Haushalt")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	// **In seinem Profil**, denn einrichten darf nur er selbst.
	if gewechselt := flaeche.ProfilWechseln(ctx, angelegt.Nutzer.Profile[0].ID, ""); gewechselt.Fehler != nil {
		t.Fatalf("wechseln: %+v", gewechselt.Fehler)
	}

	eingerichtet := flaeche.FaktorEinrichten(ctx, angelegt.Nutzer.ID)
	if eingerichtet.Fehler != nil {
		t.Fatalf("einrichten: %+v", eingerichtet.Fehler)
	}

	code := faktorcode(t, eingerichtet.Geheimnis)
	if bestaetigt := flaeche.FaktorBestaetigen(ctx, angelegt.Nutzer.ID, code); bestaetigt.Fehler != nil {
		t.Fatalf("bestätigen: %+v", bestaetigt.Fehler)
	}

	// **Der Code der Bestätigung gilt nicht noch einmal** (SEC-09) — und
	// das ist der erste Fehlversuch.
	if wieder := flaeche.KontoAnmelden(ctx, kennung, probekennwort, code); wieder.Fehler == nil {
		t.Fatal("der Code der Bestätigung galt ein zweites Mal")
	}

	for versuch := 2; versuch <= domain.Faktorgrenze; versuch++ {
		if falsch := flaeche.KontoAnmelden(ctx, kennung, probekennwort, "000000"); falsch.Fehler == nil ||
			falsch.Fehler.Schluessel == domain.SchluesselFaktorGesperrt {
			t.Fatalf("Versuch %d: %+v", versuch, falsch.Fehler)
		}
	}

	// **Gesperrt ist jeder Codeweg** (SEC-10) — auch der Zettel.
	gesperrt := flaeche.KontoAnmelden(ctx, kennung, probekennwort, eingerichtet.Codes[0])
	if gesperrt.Fehler == nil || gesperrt.Fehler.Schluessel != domain.SchluesselFaktorGesperrt {
		t.Errorf("mit dem Zettel: %+v statt der Sperre", gesperrt.Fehler)
	}
}
