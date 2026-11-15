package app_test

// **Die Sitzung an der Kommandozeile** — ID-02
// (#1701):
// derselbe festgehaltene Stand, dieselbe Antwort.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieFlaecheSiehtDieFestgehalteneSitzung(t *testing.T) {
	db := offeneProbedatenbank(t)
	flaeche := app.NutzerflaecheOhneSensor(db, &probesensor{}, keineBiometrie)
	ctx := probekontext(t)

	if stand := flaeche.Sitzungsstand(ctx); stand.Fehler != nil || stand.Zustand != "angemeldet" || stand.NutzerID != "1" {
		t.Fatalf("Stand: %+v", stand)
	}

	if antwort := flaeche.KennwortWechseln(ctx, "", probekennwort); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	for name, antwort := range map[string]app.SchreibAntwort{
		"ohne das bisherige":        flaeche.KennwortWechseln(ctx, "falsch", "noch-ein-langes-kennwort"),
		"keine Kennung":             flaeche.KennwortWiederherstellen(ctx, "x", "code", "noch-ein-langes-kennwort"),
		"angemeldet ist nichts weg": flaeche.KennwortWiederherstellen(ctx, "1", "code", "noch-ein-langes-kennwort"),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s: angenommen", name)
		}
	}

	// **Ohne lesbaren Bestand gibt es keinen Stand.**
	if fehler := db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if stand := flaeche.Sitzungsstand(ctx); stand.Fehler == nil {
		t.Errorf("ohne Bestand: %+v", stand)
	}

	if antwort := flaeche.KennwortWechseln(ctx, probekennwort, "noch-ein-langes-kennwort"); antwort.Fehler == nil {
		t.Error("ohne Bestand gewechselt")
	}
}
