package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fehler werden mit errors.Is geprüft, nie per Zeichenkettenvergleich
// (AGENTS.md §8). Diese Tabelle hält fest, dass jede Fehlerart sich auch
// wirklich so einordnen lässt — ein Fehler, den man nur an seinem Text erkennt,
// ist einer, den die Fassade nicht übersetzen kann.
func TestFehlerartenLassenSichEinordnen(t *testing.T) {
	faelle := []struct {
		name   string
		fehler error
		art    error
		fremde error
	}{
		{
			name:   "ein Feldbefund ist eine Validierung",
			fehler: domain.NewFieldError("name", domain.SchluesselPflichtfeld, "%s darf nicht leer sein", "Name"),
			art:    domain.ErrValidation,
			fremde: domain.ErrNotFound,
		},
		{
			name:   "ein umschlossener Nichtgefunden bleibt einer",
			fehler: fmt.Errorf("Set laden: %w", domain.ErrNotFound),
			art:    domain.ErrNotFound,
			fremde: domain.ErrConflict,
		},
		{
			name:   "ein umschlossener Konflikt bleibt einer",
			fehler: fmt.Errorf("Karte anlegen: %w", domain.ErrConflict),
			art:    domain.ErrConflict,
			fremde: domain.ErrValidation,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if !errors.Is(fall.fehler, fall.art) {
				t.Errorf("%v gilt nicht als %v", fall.fehler, fall.art)
			}
			if errors.Is(fall.fehler, fall.fremde) {
				t.Errorf("%v gilt fälschlich als %v", fall.fehler, fall.fremde)
			}
		})
	}
}

func TestFeldbefundTraegtFeldSchluesselUndSatz(t *testing.T) {
	fehler := domain.NewFieldError("name", domain.SchluesselPflichtfeld,
		"%s darf nicht leer sein", "Name")

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) {
		t.Fatalf("%v ist kein *domain.FieldError", fehler)
	}

	if befund.Field != "name" {
		t.Errorf("Feld %q, erwartet war %q", befund.Field, "name")
	}
	if befund.Key != domain.SchluesselPflichtfeld {
		t.Errorf("Schlüssel %q, erwartet war %q", befund.Key, domain.SchluesselPflichtfeld)
	}
	if befund.Error() != "Name darf nicht leer sein" {
		t.Errorf("Satz %q, erwartet war %q", befund.Error(), "Name darf nicht leer sein")
	}
}

// Ein nicht angemeldeter Schlüssel wird zu SchluesselUnbekannt — der Satz bleibt.
//
// Hier stand vorher, ein leerer Schlüssel sei zulässig, solange der Satz
// mitkommt. Der Zweitblick am Ende von Phase 1 hat gezeigt, was das öffnet:
// `Schluessel` ist ein offener Zeichenkettentyp, und der Vertragstest prüft nur
// die **angemeldeten**. Ein selbst erfundener Schlüssel wäre also durch beide
// Wächter gekommen und hätte in einer englischen Sitzung einen deutschen Satz
// stehen lassen.
func TestEinNichtAngemeldeterSchluesselWirdZuUnbekannt(t *testing.T) {
	faelle := []struct {
		name       string
		schluessel domain.Schluessel
	}{
		{"ein leerer Schlüssel", ""},
		{"ein erfundener Schlüssel", "gibtsNicht"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			fehler := domain.NewFieldError("name", fall.schluessel, "etwas stimmt nicht")

			if fehler.Key != domain.SchluesselUnbekannt {
				t.Errorf("Schlüssel %q, erwartet war %q", fehler.Key, domain.SchluesselUnbekannt)
			}
			if fehler.Error() != "etwas stimmt nicht" {
				t.Errorf("Satz %q, erwartet war %q", fehler.Error(), "etwas stimmt nicht")
			}
		})
	}
}

func TestIstAngemeldetKenntDieAngemeldeten(t *testing.T) {
	if !domain.IstAngemeldet(domain.SchluesselPflichtfeld) {
		t.Error("SchluesselPflichtfeld gilt als nicht angemeldet")
	}
	if domain.IstAngemeldet("gibtsNicht") {
		t.Error("ein erfundener Schlüssel gilt als angemeldet")
	}
}

// Jeder Schlüssel, den Go setzt, muss angemeldet sein.
//
// Ohne diese Anmeldung könnte die Fassade ihn nicht gegen die Kataloge halten —
// und ein Schlüssel, den niemand kennt, erscheint in der Oberfläche als
// deutscher Satz in einer englischen Sitzung.
func TestJederSchluesselIstAngemeldet(t *testing.T) {
	angemeldet := domain.Fehlerschluessel()

	if len(angemeldet) == 0 {
		t.Fatal("es ist kein Fehlerschlüssel angemeldet")
	}

	for _, schluessel := range angemeldet {
		t.Run(string(schluessel), func(t *testing.T) {
			if schluessel == "" {
				t.Error("ein leerer Schlüssel ist angemeldet")
			}
		})
	}
}

func TestAngemeldeteSchluesselSindEindeutig(t *testing.T) {
	gesehen := map[domain.Schluessel]bool{}

	for _, schluessel := range domain.Fehlerschluessel() {
		if gesehen[schluessel] {
			t.Errorf("der Schlüssel %q ist doppelt angemeldet", schluessel)
		}
		gesehen[schluessel] = true
	}
}

// Die Fehlerart folgt dem Schlüssel.
func TestDieFehlerartFolgtDemSchluessel(t *testing.T) {
	faelle := []struct {
		schluessel domain.Schluessel
		art        error
	}{
		{domain.SchluesselKonflikt, domain.ErrConflict},
		{domain.SchluesselPflichtfeld, domain.ErrValidation},
		{domain.SchluesselValidierung, domain.ErrValidation},
	}

	for _, fall := range faelle {
		t.Run(string(fall.schluessel), func(t *testing.T) {
			fehler := domain.NewFieldError("feld", fall.schluessel, "etwas")

			if !errors.Is(fehler, fall.art) {
				t.Errorf("%v gilt nicht als %v", fehler, fall.art)
			}
		})
	}
}

// **Die Abweisung einer Sicherung aus einer neueren Fassung** (#940).
func TestNeuereFassungNenntWasSieNichtKennt(t *testing.T) {
	t.Parallel()

	abweisung := &domain.NeuereFassung{
		Unbekannt: []string{"0013_neu.sql", "0014_noch_neuer.sql"},
		Stand:     "0012_paketkurzname.sql",
	}

	meldung := abweisung.Error()

	// **Alle drei Angaben**, denn jede einzelne allein hilft nicht weiter.
	for _, teil := range []string{"0013_neu.sql", "0014_noch_neuer.sql", "0012_paketkurzname.sql"} {
		if !strings.Contains(meldung, teil) {
			t.Errorf("die Meldung nennt %q nicht: %q", teil, meldung)
		}
	}

	// **Und sie ordnet sich den Sicherungsfehlern zu**, damit `internal/app`
	// sie erkennt, ohne `internal/storage` zu kennen (§2).
	if !errors.Is(abweisung, domain.ErrSicherung) {
		t.Error("die Abweisung gilt nicht als Sicherungsfehler")
	}
}
