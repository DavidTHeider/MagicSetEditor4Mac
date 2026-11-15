// **Ein Set löschen über die Naht** —
// #1218.

package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// setloeschdoppel steht für die Löschung.
type setloeschdoppel struct {
	setID   int64
	name    string
	gerufen int
	befunde []domain.Befund
	fehler  error
}

func (d *setloeschdoppel) Loeschen(_ context.Context, setID int64, name string) (
	[]domain.Befund, error,
) {
	d.setID, d.name, d.gerufen = setID, name, d.gerufen+1

	return d.befunde, d.fehler
}

func TestSetLoeschenReichtDurch(t *testing.T) {
	doppel := &setloeschdoppel{}
	anwendung := app.NeueApp("")
	anwendung.MitSetloeschung(doppel)

	antwort := anwendung.SetLoeschen(t.Context(), "7", "Mein Set")

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}
	if doppel.setID != 7 || doppel.name != "Mein Set" {
		t.Errorf("die Löschung bekam (%d, %q)", doppel.setID, doppel.name)
	}
	if len(antwort.Befunde) != 0 {
		t.Errorf("Befunde, wo keine sein sollten: %v", antwort.Befunde)
	}
}

// TestSetLoeschenGibtDieBefundeWeiter ist §7 Regel 3 an der Naht.
//
// **Kein Fehler und trotzdem etwas zu sagen.** Das Set ist weg; dass die
// Bilder liegen blieben, erfährt die Oberfläche, ohne dass sie das Set weiter
// anzeigt.
func TestSetLoeschenGibtDieBefundeWeiter(t *testing.T) {
	doppel := &setloeschdoppel{befunde: []domain.Befund{
		{
			Datei: "/wo/auch/immer/7", Meldung: "ließen sich nicht wegräumen",
			Art: domain.ArtUmgebung,
		},
	}}
	anwendung := app.NeueApp("")
	anwendung.MitSetloeschung(doppel)

	antwort := anwendung.SetLoeschen(t.Context(), "7", "Mein Set")

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}
	if len(antwort.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer", len(antwort.Befunde))
	}
}

func TestSetLoeschenUebersetztDenFehler(t *testing.T) {
	doppel := &setloeschdoppel{fehler: domain.NewFieldError(
		"name", domain.SchluesselKonflikt, "das Set heißt anders")}
	anwendung := app.NeueApp("")
	anwendung.MitSetloeschung(doppel)

	antwort := anwendung.SetLoeschen(t.Context(), "7", "Falsch")

	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}
	if antwort.Fehler.Code != app.CodeKonflikt {
		t.Errorf("Code war %q", antwort.Fehler.Code)
	}
}

// TestSetLoeschenWeistEineUnlesbareKennungAb prüft den Weg vor der Naht.
func TestSetLoeschenWeistEineUnlesbareKennungAb(t *testing.T) {
	doppel := &setloeschdoppel{}
	anwendung := app.NeueApp("")
	anwendung.MitSetloeschung(doppel)

	antwort := anwendung.SetLoeschen(t.Context(), "keine Zahl", "Mein Set")

	if antwort.Fehler == nil {
		t.Fatal("die Kennung ging durch")
	}
	if doppel.gerufen != 0 {
		t.Error("die Löschung wurde trotzdem gerufen")
	}
}

// TestDieBindungReichtDasLoeschenDurch fährt den Weg, den das Fenster nimmt.
//
// **Über [app.NeueBindungen] und nicht über [app.App]**: Wails ruft die
// Bindung, und eine Bindung, die es gibt, aber nichts weiterreicht, bestünde
// jeden Test darüber.
func TestDieBindungReichtDasLoeschenDurch(t *testing.T) {
	doppel := &setloeschdoppel{}
	anwendung := app.NeueApp("")
	anwendung.MitSetloeschung(doppel)

	antwort := app.NeueBindungen(anwendung).SetLoeschen(t.Context(), "7", "Mein Set")

	if antwort.Fehler != nil {
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	}
	if doppel.setID != 7 || doppel.name != "Mein Set" {
		t.Errorf("die Löschung bekam (%d, %q)", doppel.setID, doppel.name)
	}
}

// TestSetLoeschenOhneNaht meldet die fehlende Einrichtung.
func TestSetLoeschenOhneNaht(t *testing.T) {
	antwort := app.NeueApp("").SetLoeschen(probekontext(t), "7", "Mein Set")

	if antwort.Fehler == nil {
		t.Fatal("ohne Naht kam kein Fehler")
	}
	if antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("Code war %q", antwort.Fehler.Code)
	}
}
