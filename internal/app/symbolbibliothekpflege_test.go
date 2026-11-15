package app_test

// **Ein Symbol der Bibliothek bearbeiten, an der Fassade** — DOM-02.AK-02
// (#1697).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSymbolWirdBearbeitetNachgezogenUndZurueckgenommen(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)
	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	teile := flaeche.BibliothekssymbolTeile(ctx, krone.ID)
	if teile.Fehler != nil || len(teile.Teile) != 1 || teile.Teile[0].Name != "Rechteck" || teile.Name != "Krone" {
		t.Fatalf("die Teile: %+v", teile)
	}

	kreis := app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Kreis"}).Teile

	gesichert := flaeche.BibliothekssymbolSichern(ctx, krone.ID, kreis)
	if gesichert.Fehler != nil || gesichert.ID != krone.ID || !slices.Equal(gesichert.Veraltet, []string{"Kronenset"}) {
		t.Fatalf("sichern: %+v", gesichert)
	}

	if liste := flaeche.Symbolbibliothek(ctx, ""); !liste.Symbole[0].Vorige || len(liste.Symbole[0].Verwendet) != 0 {
		t.Errorf("danach: %+v", liste.Symbole[0])
	}

	if nachgezogen := flaeche.BibliothekssymbolNachziehen(ctx, krone.ID); nachgezogen.Fehler != nil ||
		!slices.Equal(nachgezogen.Sets, []string{"Kronenset"}) {
		t.Fatalf("nachziehen: %+v", nachgezogen)
	}

	zurueck := flaeche.BibliothekssymbolZuruecknehmen(ctx, krone.ID)
	if zurueck.Fehler != nil || !slices.Equal(zurueck.Veraltet, []string{"Kronenset"}) {
		t.Fatalf("zurücknehmen: %+v", zurueck)
	}

	if wieder := flaeche.BibliothekssymbolTeile(ctx, krone.ID); wieder.Teile[0].Name != "Rechteck" {
		t.Errorf("nach dem Zurücknehmen: %+v", wieder.Teile)
	}
}

func TestDieBearbeitungSagtWasNichtGeht(t *testing.T) {
	flaeche, mitSymbol, _ := bibliotheksflaeche(t)
	ctx := probekontext(t)
	krone := flaeche.BibliothekssymbolAufnehmen(ctx, mitSymbol, "Krone")

	for name, fehler := range map[string]*app.Fehler{
		"Teile":        flaeche.BibliothekssymbolTeile(ctx, "x").Fehler,
		"Sichern":      flaeche.BibliothekssymbolSichern(ctx, "x", nil).Fehler,
		"Zurücknehmen": flaeche.BibliothekssymbolZuruecknehmen(ctx, "x").Fehler,
		"Nachziehen":   flaeche.BibliothekssymbolNachziehen(ctx, "x").Fehler,
	} {
		if fehler == nil || fehler.Feld != "id" {
			t.Errorf("%s ohne Kennung: %+v", name, fehler)
		}
	}

	for name, fehler := range map[string]*app.Fehler{
		"Teile":        flaeche.BibliothekssymbolTeile(ctx, "99").Fehler,
		"Sichern":      flaeche.BibliothekssymbolSichern(ctx, "99", nil).Fehler,
		"Zurücknehmen": flaeche.BibliothekssymbolZuruecknehmen(ctx, "99").Fehler,
		"Nachziehen":   flaeche.BibliothekssymbolNachziehen(ctx, "99").Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeNichtGefunden {
			t.Errorf("%s ohne Symbol: %+v", name, fehler)
		}
	}

	if fehler := flaeche.BibliothekssymbolZuruecknehmen(ctx, krone.ID).Fehler; fehler == nil ||
		fehler.Schluessel != domain.SchluesselSymbolOhneVorige {
		t.Errorf("ohne vorige Fassung: %+v", fehler)
	}

	if antwort := flaeche.BibliothekssymbolNachziehen(ctx, "x"); antwort.Sets == nil {
		t.Error("die Folgen sind nil")
	}
}
