package domain_test

// **Die Invarianten und die Rangfolge** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass die vier Invarianten eine Aufzählung sind und keine Prosa.** G1
// hatte den Satz im Kommentar; erst als Wert lässt er sich gegen eine
// Änderung halten.
//
// **Und dass die Leiter in der Reihenfolge des Issues steht** — eine
// Rangfolge, deren Reihenfolge jemand umdreht, kehrt alle Entscheidungen um.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Es sind vier, und ein erfundener Name ist keiner.**
func TestDieInvariantenSindVier(t *testing.T) {
	t.Parallel()

	if len(domain.Invarianten()) != 4 {
		t.Fatalf("%d Invarianten", len(domain.Invarianten()))
	}

	for _, gesetz := range domain.Invarianten() {
		if !domain.IstInvariante(gesetz) {
			t.Errorf("%q gilt nicht", gesetz)
		}
	}

	if domain.IstInvariante("bequemlichkeit") {
		t.Error("eine erfundene Invariante gilt")
	}
}

// **Die Leiter steht von stark nach schwach.**
//
// # Die Zusicherung, die wirklich zählt
//
// **Dass die Invariante ganz oben steht und die Hausregel ganz unten.** Das
// Issue gibt die Reihenfolge wörtlich vor; wer sie umdreht, kehrt jede
// Entscheidung um — und zwar still, denn beide Seiten sind Zahlen.
func TestDieRangfolgeStehtVonStarkNachSchwach(t *testing.T) {
	t.Parallel()

	raenge := domain.Regelraenge()

	if len(raenge) != 5 {
		t.Fatalf("%d Ränge", len(raenge))
	}

	erwartet := []domain.Regelrang{
		domain.RangInvariante, domain.RangBasis, domain.RangFormat,
		domain.RangTurnier, domain.RangHausregel,
	}

	for i, rang := range raenge {
		if rang != erwartet[i] {
			t.Errorf("an Stelle %d steht %d, erwartet %d", i, rang, erwartet[i])
		}

		if !domain.IstRegelrang(rang) {
			t.Errorf("%d gilt nicht", rang)
		}

		if name := rang.Name(); name == "" || strings.Contains(name, "Unbekannt") {
			t.Errorf("Rang %d heißt %q", rang, name)
		}
	}

	if domain.IstRegelrang(99) {
		t.Error("ein erfundener Rang gilt")
	}

	if name := domain.Regelrang(99).Name(); !strings.Contains(name, "Unbekannt") {
		t.Errorf("ein erfundener Rang heißt %q", name)
	}
}

// **Ein Konflikt sagt, wer wen nicht ändern kann — und woran es lag.**
//
// **„Das geht nicht" sagt niemandem, was zu tun ist.** Der Satz nennt beide
// Stufen und den Gegenstand; wer ihn liest, weiß, welche Angabe er ändern
// muss.
func TestEinKonfliktNenntBeideStufenUndDenGegenstand(t *testing.T) {
	t.Parallel()

	fehler := domain.Regelkonflikt{
		Staerker:   domain.RangInvariante,
		Schwaecher: domain.RangHausregel,
		Gegenstand: "sichtbarkeit",
	}.Fehler()

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v ist keine Validierung", fehler)
	}

	satz := fehler.Error()

	for _, teil := range []string{"Hausregel", "Invariante", "sichtbarkeit"} {
		if !strings.Contains(satz, teil) {
			t.Errorf("%q nennt %q nicht", satz, teil)
		}
	}
}
