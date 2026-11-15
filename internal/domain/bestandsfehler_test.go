package domain_test

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Drei Arten von Konflikt, und nur eine heißt „lade neu"** — ENG-05.AK-03
// (#1730),
// nativer Befund F08.
func TestDieKonfliktartenBleibenKonflikte(t *testing.T) {
	t.Parallel()

	for name, art := range map[string]error{
		"doppelt":  domain.ErrDoppelt,
		"überholt": domain.ErrUeberholt,
	} {
		if !errors.Is(art, domain.ErrConflict) {
			t.Errorf("%s ist kein Konflikt mehr — REST antwortete nicht mehr mit 409", name)
		}
	}

	if errors.Is(domain.ErrDoppelt, domain.ErrUeberholt) || errors.Is(domain.ErrUeberholt, domain.ErrDoppelt) {
		t.Error("doppelt und überholt sind dieselbe Art geworden")
	}

	if errors.Is(domain.ErrVoll, domain.ErrConflict) || errors.Is(domain.ErrVoll, domain.ErrValidation) {
		t.Error("ein voller Datenträger ist weder Konflikt noch falsche Eingabe")
	}
}

// **Ein Feldbefund über etwas Doppeltes ist ein doppelter Eintrag** — und
// trägt den Treibertext als verdeckten Grund.
func TestEinBefundUeberEtwasDoppeltesIstDoppelt(t *testing.T) {
	t.Parallel()

	treiber := errors.New("UNIQUE constraint failed: decks.profil_id, decks.spiel, decks.name")

	for _, schluessel := range []domain.Schluessel{
		domain.SchluesselDoppelt,
		domain.SchluesselDoppelterName,
		domain.SchluesselDoppeltesExemplar,
	} {
		befund := domain.NewFieldError("name", schluessel, "gibt es schon").MitGrund(treiber)

		if !errors.Is(befund, domain.ErrDoppelt) || !errors.Is(befund, domain.ErrConflict) {
			t.Errorf("%s: kein doppelter Eintrag", schluessel)
		}

		if !errors.Is(befund, treiber) {
			t.Errorf("%s: der Grund ging aus der Kette verloren", schluessel)
		}

		if grund := domain.GrundVon(fmt.Errorf("das Deck anlegen: %w", befund)); !errors.Is(grund, treiber) {
			t.Errorf("%s: GrundVon gibt %v, erwartet den Treibertext", schluessel, grund)
		}

		if befund.Error() != "gibt es schon" {
			t.Errorf("%s: der Satz trägt den Grund mit: %q", schluessel, befund.Error())
		}
	}
}

// **Ohne Grund bleibt ein Befund, was er war.**
func TestEinBefundOhneGrundHatKeinen(t *testing.T) {
	t.Parallel()

	befund := domain.NewFieldError("name", domain.SchluesselPflichtfeld, "leer")

	if domain.GrundVon(befund) != nil {
		t.Error("ein Befund ohne Grund meldet einen")
	}

	if !errors.Is(befund, domain.ErrValidation) {
		t.Error("ein Pflichtfeld ist keine Validierung mehr")
	}
}

// **ENOSPC ist ein voller Datenträger** (#1812) — gleich, wo geschrieben
// wurde.
func TestPlatzmangelErkenntEinenVollenDatentraeger(t *testing.T) {
	t.Parallel()

	voll := &fs.PathError{Op: "write", Path: "/geheim/pfad/set.mse-set", Err: syscall.ENOSPC}

	eingeordnet := domain.Platzmangel(fmt.Errorf("das Set sichern: %w", voll))
	if !errors.Is(eingeordnet, domain.ErrVoll) {
		t.Fatalf("%v ist kein voller Datenträger", eingeordnet)
	}

	if grund := domain.GrundVon(eingeordnet); !errors.Is(grund, syscall.ENOSPC) {
		t.Errorf("der Grund ist %v, erwartet ENOSPC", grund)
	}

	// **Der Pfad bleibt im Grund und nicht im Satz** (§8).
	if satz := eingeordnet.Error(); satz == "" || strings.Contains(satz, "/geheim/pfad") {
		t.Errorf("der Satz trägt den Pfad: %q", satz)
	}

	andere := errors.New("etwas anderes")
	if umgedeutet := domain.Platzmangel(andere); errors.Is(umgedeutet, domain.ErrVoll) ||
		umgedeutet.Error() != andere.Error() {
		t.Errorf("ein anderer Fehler wurde umgedeutet: %v", umgedeutet)
	}

	if domain.Platzmangel(nil) != nil {
		t.Error("kein Fehler wurde zu einem")
	}

	schon := fmt.Errorf("außen: %w", domain.ErrVoll)
	if zweimal := domain.Platzmangel(schon); zweimal.Error() != schon.Error() {
		t.Errorf("ein schon eingeordneter voller Datenträger wurde zweimal eingewickelt: %v", zweimal)
	}
}

// **Ein Ort ohne Zugriffsrecht ist gesperrt** — FIX-F02 (#1746), dieselbe
// Gestalt wie der volle Datenträger: EACCES und EPERM, gleich wo zugegriffen
// wurde, und der Pfad bleibt im Grund. **Auch beim Lesen**: macOS verweigert
// es in geschützten Ordnern mit `EPERM`, und der Satz darf dann nicht vom
// Schreiben reden (N8 des Nachlaufs über Phase 6.3).
func TestZugriffsmangelErkenntEinenGesperrtenOrt(t *testing.T) {
	t.Parallel()

	for _, zahl := range []syscall.Errno{syscall.EACCES, syscall.EPERM} {
		for _, art := range []string{"mkdir", "open"} {
			gesperrt := &fs.PathError{Op: art, Path: "/geheim/pfad/set.mse-set", Err: zahl}

			eingeordnet := domain.Zugriffsmangel(fmt.Errorf("das Set: %w", gesperrt))
			if !errors.Is(eingeordnet, domain.ErrGesperrt) {
				t.Fatalf("%v ist kein gesperrter Ort", eingeordnet)
			}

			if satz := eingeordnet.Error(); strings.Contains(satz, "schreib") || strings.Contains(satz, "schrieb") {
				t.Errorf("%s: der Satz redet vom Schreiben: %q", art, eingeordnet.Error())
			}
		}

		gesperrt := &fs.PathError{Op: "mkdir", Path: "/geheim/pfad/set.mse-set", Err: zahl}

		eingeordnet := domain.Zugriffsmangel(fmt.Errorf("das Set sichern: %w", gesperrt))

		if grund := domain.GrundVon(eingeordnet); !errors.Is(grund, zahl) {
			t.Errorf("der Grund ist %v, erwartet %v", grund, zahl)
		}

		if strings.Contains(eingeordnet.Error(), "/geheim/pfad") {
			t.Errorf("der Satz trägt den Pfad: %q", eingeordnet.Error())
		}
	}

	andere := errors.New("etwas anderes")
	if umgedeutet := domain.Zugriffsmangel(andere); errors.Is(umgedeutet, domain.ErrGesperrt) {
		t.Errorf("ein anderer Fehler wurde umgedeutet: %v", umgedeutet)
	}

	if domain.Zugriffsmangel(nil) != nil {
		t.Error("kein Fehler wurde zu einem")
	}

	schon := fmt.Errorf("außen: %w", domain.ErrGesperrt)
	if zweimal := domain.Zugriffsmangel(schon); zweimal.Error() != schon.Error() {
		t.Errorf("ein schon eingeordneter Ort wurde zweimal eingewickelt: %v", zweimal)
	}
}

// **Die Schlüssel sind angemeldet.**
func TestDieSchluesselDesBestandsSindAngemeldet(t *testing.T) {
	t.Parallel()

	for _, schluessel := range []domain.Schluessel{
		domain.SchluesselDoppelt,
		domain.SchluesselDoppelterName,
		domain.SchluesselDoppeltesExemplar,
		domain.SchluesselKonfliktgrund,
		domain.SchluesselVoll,
		domain.SchluesselGesperrt,
	} {
		if !domain.IstAngemeldet(schluessel) {
			t.Errorf("%s ist nicht angemeldet", schluessel)
		}
	}
}
