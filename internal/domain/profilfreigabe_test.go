package domain_test

// **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Drei Rollen, und jede schließt die darunter ein.**
func TestEineRolleSchliesstDieDarunterEin(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		hat      domain.Rolle
		verlangt domain.Rolle
		reicht   bool
	}{
		"Leser liest":              {domain.RolleLeser, domain.RolleLeser, true},
		"Leser arbeitet nicht mit": {domain.RolleLeser, domain.RolleMitarbeiter, false},
		"Leser verwaltet nicht":    {domain.RolleLeser, domain.RolleVerwalter, false},
		"Mitarbeiter liest auch":   {domain.RolleMitarbeiter, domain.RolleLeser, true},
		"Mitarbeiter arbeitet mit": {domain.RolleMitarbeiter, domain.RolleMitarbeiter, true},
		"Mitarbeiter verwaltet nicht": {
			domain.RolleMitarbeiter, domain.RolleVerwalter, false,
		},
		"Verwalter darf alles":   {domain.RolleVerwalter, domain.RolleVerwalter, true},
		"Verwalter liest auch":   {domain.RolleVerwalter, domain.RolleLeser, true},
		"keine Rolle reicht nie": {"", domain.RolleLeser, false},

		// **Eine erfundene Rolle reicht für nichts** — auch nicht für eine
		// erfundene: `rangfolge` gibt für beide null, und `0 >= 0` wäre wahr.
		"erfundene Rolle":      {"gott", domain.RolleLeser, false},
		"erfundenes Verlangen": {domain.RolleVerwalter, "gott", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fall.hat.Reicht(fall.verlangt) != fall.reicht {
				t.Errorf("%q reicht für %q = %v, erwartet %v",
					fall.hat, fall.verlangt, !fall.reicht, fall.reicht)
			}
		})
	}
}

// **Und es sind genau drei.**
func TestEsSindDreiRollen(t *testing.T) {
	t.Parallel()

	if rollen := domain.Rollen(); len(rollen) != 3 {
		t.Fatalf("es sind %d Rollen: %v", len(rollen), rollen)
	}

	for _, rolle := range domain.Rollen() {
		if fehler := rolle.Pruefe(); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", rolle, fehler)
		}
	}

	if fehler := domain.Rolle("gott").Pruefe(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die erfundene Rolle kam durch: %v", fehler)
	}
}

// **Zwei Sichtbarkeiten, und leer heißt privat.**
func TestEineSichtbarkeitWirdGeprueft(t *testing.T) {
	t.Parallel()

	for _, sichtbar := range domain.Sichtbarkeiten() {
		if fehler := sichtbar.Pruefe(); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", sichtbar, fehler)
		}
	}

	// **Leer gilt** — ein Aufruf, der sie wegläßt, meint nicht „egal", sondern
	// die vorsichtige Antwort.
	if fehler := domain.Sichtbarkeit("").Pruefe(); fehler != nil {
		t.Errorf("die leere Sichtbarkeit wurde abgewiesen: %v", fehler)
	}

	if fehler := domain.Sichtbarkeit("halb").Pruefe(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die erfundene Sichtbarkeit kam durch: %v", fehler)
	}
}

// **Eine Freigabe braucht Profil, Nutzer und eine echte Rolle.**
func TestEineFreigabeWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		freigabe domain.Profilfreigabe
		glueckt  bool
		feld     string
	}{
		"vollständig": {
			domain.Profilfreigabe{
				ProfilID: 1, NutzerID: 2, Rolle: domain.RolleLeser,
			}, true, "",
		},
		"ohne Profil": {
			domain.Profilfreigabe{NutzerID: 2, Rolle: domain.RolleLeser},
			false, "profilID",
		},
		"ohne Nutzer": {
			domain.Profilfreigabe{ProfilID: 1, Rolle: domain.RolleLeser},
			false, "nutzerID",
		},
		"erfundene Rolle": {
			domain.Profilfreigabe{ProfilID: 1, NutzerID: 2, Rolle: "gott"},
			false, "rolle",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.freigabe.Validate()
			if fall.glueckt != (fehler == nil) {
				t.Fatalf("glückt = %v, Fehler = %v", fall.glueckt, fehler)
			}

			if fall.glueckt {
				return
			}

			// **Der Befund nennt das Feld** (§8) — sonst stünde die Meldung
			// über dem Formular statt an der Stelle, die sie meint.
			var feld *domain.FieldError
			if !errors.As(fehler, &feld) || feld.Field != fall.feld {
				t.Errorf("der Befund nennt %v statt %q", fehler, fall.feld)
			}
		})
	}
}

// **Die Absage nennt, was verlangt war.**
//
// **Die einzige Ablehnung, die mehr sagt als „nein"** — dieselbe Überlegung
// wie beim Scope eines Zugriffstokens: wer bis hierher kommt, ist kein
// Fremder, und ohne diesen Satz suchte er den Fehler bei sich.
func TestDieFreigabeabsageNenntWasVerlangtWar(t *testing.T) {
	t.Parallel()

	fehler := domain.Freigabefehler(domain.RolleMitarbeiter)

	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Fehler = %v, erwartet ErrErlaubnis", fehler)
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		// **Und ausdrücklich kein [domain.ErrValidation]** — wären beide
		// dasselbe, unterschiede eine Fläche den Berechtigten nicht vom
		// Fremden.
		return
	}

	t.Error("die Absage ist zugleich eine Validierung")
}
