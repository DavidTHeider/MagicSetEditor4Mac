package app

// **Ein Schritt nennt die Revision, die er meint** — #2496, an der Bindung.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSchrittMitFremdemAbdruckIstEinKonflikt(t *testing.T) {
	anwendung, _ := eineLeitungsflaeche(t)
	ctx := riegelkontext(t)

	angelegt := anwendung.HandelEroeffnen(ctx, einLeitungswunsch())
	if angelegt.Fehler != nil {
		t.Fatalf("Eröffnen: %+v", angelegt.Fehler)
	}

	abgewiesen := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", "ein-anderer-abdruck")
	if abgewiesen.Fehler == nil || abgewiesen.Fehler.Code != CodeKonflikt {
		t.Fatalf("mit fremdem Abdruck kam %+v", abgewiesen.Fehler)
	}

	// **Mit dem gesehenen geht es.**
	if angeboten := anwendung.HandelSchritt(ctx, derLeitungsvorgang,
		string(domain.SchrittAnbieten), "", angelegt.Vorgang.Abdruck); angeboten.Fehler != nil {
		t.Errorf("mit dem gesehenen Abdruck: %+v", angeboten.Fehler)
	}
}
