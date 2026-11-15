package domain_test

// **Welche Kanäle das Protokoll hat und wie ausführlich es ist** — OPS-04
// (#1721):
// Datei, Standardausgabe und Systemprotokoll unabhängig wählbar, Stufe,
// Komponentenfilter, Ort, Drehung, Aufbewahrung und Format einstellbar.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieProtokolleinstellungenGehoerenDemGeraet(t *testing.T) {
	t.Parallel()

	for _, schluessel := range domain.Protokollschluessel() {
		ebene, _ := domain.EbeneVon(schluessel)
		if !slices.Contains(domain.Einstellungsreihe(), schluessel) || ebene != domain.EbeneGeraet ||
			domain.EinstellungsWirkung(schluessel) != domain.WirkungNeustart ||
			domain.EinstellungsBereich(schluessel) != domain.BereichAnwendung {
			t.Errorf("%s: Ebene %q, Wirkung %q, Bereich %q", schluessel, ebene,
				domain.EinstellungsWirkung(schluessel), domain.EinstellungsBereich(schluessel))
		}
	}

	if len(domain.Protokollschluessel()) != 9 {
		t.Errorf("%d Schlüssel", len(domain.Protokollschluessel()))
	}
}

func TestDieProtokollangabenWerdenGeprueft(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		schluessel string
		wert       string
		gut        bool
	}{
		{domain.ProtokollGrenze, "", true},
		{domain.ProtokollGrenze, "1", true},
		{domain.ProtokollGrenze, "64", true},
		{domain.ProtokollGrenze, "0", false},
		{domain.ProtokollGrenze, "65", false},
		{domain.ProtokollGrenze, "4MB", false},
		{domain.ProtokollAufbewahrung, "", true},
		{domain.ProtokollAufbewahrung, "9", true},
		{domain.ProtokollAufbewahrung, "0", false},
		{domain.ProtokollAufbewahrung, "10", false},
		{domain.ProtokollKomponenten, "", true},
		{domain.ProtokollKomponenten, "service,storage", true},
		{domain.ProtokollKomponenten, "service, storage", false},
		{domain.ProtokollKomponenten, "Service", false},
		{domain.ProtokollKomponenten, "service,,storage", false},
		{domain.ProtokollOrt, "~/Protokolle", true},
		{domain.ProtokollOrt, "relativ", false},
		{domain.ProtokollStufe, "ausfuehrlich", true},
		{domain.ProtokollStufe, "debug", false},
		{domain.ProtokollFormat, "json", true},
	}

	for _, fall := range faelle {
		fehler := domain.PruefeEinstellung(fall.schluessel, fall.wert)
		if fall.gut != (fehler == nil) || (!fall.gut && !errors.Is(fehler, domain.ErrValidation)) {
			t.Errorf("%s=%q: %v", fall.schluessel, fall.wert, fehler)
		}
	}
}

func TestDieProtokollregelAusDenEinstellungen(t *testing.T) {
	t.Parallel()

	vorgabe := domain.ProtokollregelAus(map[string]string{})
	if vorgabe.Stufe != domain.ProtokollstufeInfo || !vorgabe.Datei || !vorgabe.System || vorgabe.Ausgabe ||
		vorgabe.Format != domain.ProtokollformatText || vorgabe.GrenzeBytes != 4<<20 ||
		vorgabe.Aufbewahrung != 1 || len(vorgabe.Komponenten) != 0 || vorgabe.Ort != "" || vorgabe.Ausfuehrlich() {
		t.Errorf("Vorgabe: %+v", vorgabe)
	}

	gewaehlt := domain.ProtokollregelAus(map[string]string{
		domain.ProtokollStufe: domain.ProtokollstufeWarnung, domain.ProtokollDatei: domain.ProtokollAus,
		domain.ProtokollAusgabe: domain.ProtokollAn, domain.ProtokollFormat: domain.ProtokollformatJSON,
		domain.ProtokollGrenze: "16", domain.ProtokollAufbewahrung: "3",
		domain.ProtokollKomponenten: "service,storage", domain.ProtokollOrt: "~/Protokolle",
	})
	if gewaehlt.Stufe != domain.ProtokollstufeWarnung || gewaehlt.Datei || !gewaehlt.Ausgabe ||
		gewaehlt.GrenzeBytes != 16<<20 || gewaehlt.Aufbewahrung != 3 ||
		!slices.Equal(gewaehlt.Komponenten, []string{"service", "storage"}) || gewaehlt.Ort != "~/Protokolle" {
		t.Errorf("gewählt: %+v", gewaehlt)
	}

	// **Ausführlich hat eine begrenzte Aufbewahrung** (AK-03): höchstens
	// 1 MB je Datei und eine Vorfassung, gleich was eingestellt ist — ein
	// Protokoll, das jeden Schritt kennt, soll nicht wochenlang liegen.
	ausfuehrlich := domain.ProtokollregelAus(map[string]string{
		domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich, domain.ProtokollGrenze: "32",
		domain.ProtokollAufbewahrung: "5",
	})
	if !ausfuehrlich.Ausfuehrlich() || ausfuehrlich.GrenzeBytes != 1<<20 || ausfuehrlich.Aufbewahrung != 1 {
		t.Errorf("ausführlich: %+v", ausfuehrlich)
	}

	// **Ein unlesbarer Wert fällt auf die Vorgabe** — ein von Hand geänderter
	// Bestand darf das Protokoll nicht abschalten.
	kaputt := domain.ProtokollregelAus(map[string]string{
		domain.ProtokollGrenze: "viel", domain.ProtokollAufbewahrung: "-1", domain.ProtokollStufe: "laut",
	})
	if kaputt.GrenzeBytes != 4<<20 || kaputt.Aufbewahrung != 1 || kaputt.Stufe != domain.ProtokollstufeInfo {
		t.Errorf("kaputt: %+v", kaputt)
	}
}
