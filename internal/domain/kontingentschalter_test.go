package domain_test

// **Die Kontingente lassen sich abschalten** — QA-08.AK-03
// (#1739).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerKontingentschalterIstEineGeraeteeinstellung(t *testing.T) {
	t.Parallel()

	schluessel := domain.GrenzenKontingente

	if ebene, _ := domain.EbeneVon(schluessel); ebene != domain.EbeneGeraet {
		t.Errorf("Ebene %q", ebene)
	}

	if wirkung := domain.EinstellungsWirkung(schluessel); wirkung != domain.WirkungNeustart {
		t.Errorf("Wirkung %q", wirkung)
	}

	if bereich := domain.EinstellungsBereich(schluessel); bereich != domain.BereichNetz {
		t.Errorf("Bereich %q", bereich)
	}

	if vorgabe := domain.EinstellungsVorgabe(schluessel); vorgabe != domain.KontingenteAn {
		t.Errorf("Vorgabe %q", vorgabe)
	}

	if werte := domain.ErlaubteEinstellungswerte(schluessel); !slices.Equal(werte,
		[]string{domain.KontingenteAn, domain.KontingenteAus}) {
		t.Errorf("Werte %v", werte)
	}
}

func TestNurEinAusdruecklichesAusSchaltetDieKontingenteAb(t *testing.T) {
	t.Parallel()

	// **Eine Liste und keine Map**: ein Schlüssel mit Leerraum vorn ist hier
	// gerade der Fall, den die Prüfung ablehnen soll.
	for _, fall := range []struct {
		wert string
		ohne bool
	}{
		{"", false}, {"an", false}, {"aus", true}, {"AUS", false}, {"nein", false}, {" aus", false},
	} {
		wert, ohne := fall.wert, fall.ohne
		werte := map[string]string{domain.GrenzenKontingente: wert}

		if ist := domain.OhneKontingente(werte); ist != ohne {
			t.Errorf("%q: ohne=%v", wert, ist)
		}

		if ist := domain.Vorgabegrenzen().NachEinstellungen(werte).OhneKontingente; ist != ohne {
			t.Errorf("%q in den Grenzen: ohne=%v", wert, ist)
		}
	}

	// **Ohne Einstellung gelten sie** — die Vorgabe schützt.
	if domain.Vorgabegrenzen().OhneKontingente || domain.OhneKontingente(nil) {
		t.Error("die Vorgabe schaltet die Kontingente ab")
	}
}

func TestDerKontingentschalterAendertNurDieKontingente(t *testing.T) {
	t.Parallel()

	vorher := domain.Vorgabegrenzen()
	nachher := vorher.NachEinstellungen(map[string]string{domain.GrenzenKontingente: domain.KontingenteAus})

	// **Alles andere bleibt, wie es war** — Plätze, Antwortgrenze, Zahlen.
	nachher.OhneKontingente = false
	if nachher.Antwort != vorher.Antwort || nachher.Principal != vorher.Principal ||
		nachher.Global != vorher.Global || nachher.Reserve != vorher.Reserve ||
		len(nachher.Parallel) != len(vorher.Parallel) {
		t.Errorf("geändert: %+v", nachher)
	}

	for klasse, anzahl := range vorher.Parallel {
		if nachher.Parallel[klasse] != anzahl {
			t.Errorf("Plätze %s: %d statt %d", klasse, nachher.Parallel[klasse], anzahl)
		}
	}

	if stand := vorher.Kontingentstand(); stand != domain.KontingenteAn {
		t.Errorf("Stand mit Kontingenten: %q", stand)
	}

	if stand := vorher.NachEinstellungen(map[string]string{
		domain.GrenzenKontingente: domain.KontingenteAus,
	}).Kontingentstand(); stand != domain.KontingenteAus {
		t.Errorf("Stand ohne Kontingente: %q", stand)
	}
}

// **„Nur lokal" öffnet keine Schranke.** Die Richtlinie schaltet jeden
// Schlüssel auf `dienst.` ab; hieße der Schalter so, schaltete wer „nur
// lokal" wählt, die Kontingente gleich mit ab.
func TestNurLokalLaesstDieKontingenteStehen(t *testing.T) {
	t.Parallel()

	stand := domain.WirksamerStand(map[string]string{domain.RichtlinieNurLokal: domain.NurLokalAn}, nil)

	if geltung := stand[domain.GrenzenKontingente]; geltung.Wirksam != domain.KontingenteAn {
		t.Errorf("unter „nur lokal“: %+v", geltung)
	}
}
