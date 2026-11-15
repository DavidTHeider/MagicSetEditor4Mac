package domain_test

// **Wie die Ablaufspur und das Leistungsprofil eingestellt sind** — OPS-05.AK-03
// (#1722):
// „Sampling, Speicherbudget und Aufbewahrung sind einstellbar;
// Profiling/Tracing ist gezielt aktivierbar und verursacht keinen
// unbegrenzten lokalen Datenbestand."

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieMesseinstellungenGehoerenDemGeraet(t *testing.T) {
	t.Parallel()

	// **Die Spur gilt ab dem nächsten Start, die Zahl der Profile sofort** —
	// die Spur richtet jeder Prozess beim Start ein, die Profile zählt der
	// Ruf, der eines schreibt.
	for _, schluessel := range domain.Messschluessel() {
		wirkung := domain.WirkungNeustart
		if schluessel == domain.MessungProfile {
			wirkung = domain.WirkungSofort
		}

		ebene, _ := domain.EbeneVon(schluessel)
		if !slices.Contains(domain.Einstellungsreihe(), schluessel) || ebene != domain.EbeneGeraet ||
			domain.EinstellungsWirkung(schluessel) != wirkung ||
			domain.EinstellungsBereich(schluessel) != domain.BereichAnwendung {
			t.Errorf("%s: Ebene %q, Wirkung %q, Bereich %q", schluessel, ebene,
				domain.EinstellungsWirkung(schluessel), domain.EinstellungsBereich(schluessel))
		}
	}

	if len(domain.Messschluessel()) != 5 {
		t.Errorf("%d Schlüssel", len(domain.Messschluessel()))
	}
}

func TestDieMessangabenWerdenGeprueft(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		schluessel string
		wert       string
		gut        bool
	}{
		{domain.MessungSpur, "an", true},
		{domain.MessungSpur, "ja", false},
		{domain.MessungAbtastung, "", true},
		{domain.MessungAbtastung, "1", true},
		{domain.MessungAbtastung, "100", true},
		{domain.MessungAbtastung, "0", false},
		{domain.MessungAbtastung, "101", false},
		{domain.MessungSpurgrenze, "10", true},
		{domain.MessungSpurgrenze, "10000", true},
		{domain.MessungSpurgrenze, "9", false},
		{domain.MessungSpurgrenze, "10001", false},
		{domain.MessungAufbewahrung, "1440", true},
		{domain.MessungAufbewahrung, "1441", false},
		{domain.MessungProfile, "20", true},
		{domain.MessungProfile, "21", false},
	}

	for _, fall := range faelle {
		fehler := domain.PruefeEinstellung(fall.schluessel, fall.wert)
		if fall.gut != (fehler == nil) || (!fall.gut && !errors.Is(fehler, domain.ErrValidation)) {
			t.Errorf("%s=%q: %v", fall.schluessel, fall.wert, fehler)
		}
	}
}

func TestDieMessregelAusDenEinstellungen(t *testing.T) {
	t.Parallel()

	// **Aus ist die Vorgabe** — eine Spur, die niemand bestellt hat, liegt
	// nicht im Arbeitsspeicher.
	vorgabe := domain.MessregelAus(map[string]string{})
	if vorgabe.Spur || vorgabe.Abtastung != 100 || vorgabe.Spurgrenze != 1000 ||
		vorgabe.Aufbewahrung != time.Hour || vorgabe.Profile != 3 {
		t.Errorf("Vorgabe: %+v", vorgabe)
	}

	gewaehlt := domain.MessregelAus(map[string]string{
		domain.MessungSpur: "an", domain.MessungAbtastung: "10", domain.MessungSpurgrenze: "50",
		domain.MessungAufbewahrung: "5", domain.MessungProfile: "1",
	})
	if !gewaehlt.Spur || gewaehlt.Abtastung != 10 || gewaehlt.Spurgrenze != 50 ||
		gewaehlt.Aufbewahrung != 5*time.Minute || gewaehlt.Profile != 1 {
		t.Errorf("gewählt: %+v", gewaehlt)
	}

	// **Ein unlesbarer Wert fällt auf die Vorgabe** und nie auf „unbegrenzt".
	kaputt := domain.MessregelAus(map[string]string{
		domain.MessungSpurgrenze: "999999", domain.MessungAbtastung: "-3", domain.MessungProfile: "viele",
	})
	if kaputt.Spurgrenze != 1000 || kaputt.Abtastung != 100 || kaputt.Profile != 3 {
		t.Errorf("kaputt: %+v", kaputt)
	}
}
