package domain_test

// **Nutzer und Profil** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk 0066.
//
// # Was hier steht und was nicht
//
// **Die Begriffe, nicht die Trennung.** Ein Nutzer ist die Person, ein Profil
// eine Arbeitsumgebung dieser Person. Dass ein Profil die Daten eines anderen
// nicht sieht, ist Werkstück 2 — und es ist eine Frage an den Bestand, nicht
// an die Domäne.
//
// **Kein Kennwort und kein Geheimnis.** Was ein zweiter Faktor ist, entscheidet
// Werkstück 5; was hier liegt, ist nichts, das man hüten müsste.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Nutzer braucht einen Namen.**
func TestEinNutzerBrauchtEinenNamen(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"", "   ", "\t\n"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var feld *domain.FieldError
			if err := (domain.Nutzer{Name: name}).Validate(); !errors.As(err, &feld) ||
				feld.Field != "name" {
				t.Errorf("Fehler ist %v", err)
			}
		})
	}
}

// **Ein Name mit Rändern wird nicht heimlich beschnitten.**
//
// Er wird **abgewiesen**. Zwei Nutzer „David" und „David " wären zwei Zeilen
// mit demselben sichtbaren Namen, und wer sie unterscheiden soll, sieht
// nichts. Stillschweigend zu beschneiden hieße, eine Eingabe zu ändern, ohne
// es zu sagen (§8).
func TestEinNameMitRaendernWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	var feld *domain.FieldError
	if err := (domain.Nutzer{Name: " David "}).Validate(); !errors.As(err, &feld) {
		t.Errorf("Fehler ist %v", err)
	}
}

// **Ein Profil braucht einen Namen und einen Nutzer.**
func TestEinProfilBrauchtNamenUndNutzer(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		profil domain.Profil
		feld   string
	}{
		"ohne Namen":  {domain.Profil{NutzerID: 1}, "name"},
		"ohne Nutzer": {domain.Profil{Name: "Arbeit"}, "nutzerID"},

		// **Und die Sichtbarkeit steht vor beiden** — Werkstück 11
		// (#1351). Sie ist die einzige, die aus einem **Bestand** kommen
		// kann: Name und Nutzer setzt immer ein Aufrufer, eine
		// unbekannte Sichtbarkeit steht in einer Zeile, die eine ältere
		// Fassung geschrieben hat.
		"mit einer Sichtbarkeit, die es nicht gibt": {
			domain.Profil{NutzerID: 1, Name: "Arbeit", Sichtbarkeit: "halb"},
			"sichtbarkeit",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var feld *domain.FieldError
			if err := fall.profil.Validate(); !errors.As(err, &feld) ||
				feld.Field != fall.feld {
				t.Errorf("Fehler ist %v, erwartet am Feld %q", err, fall.feld)
			}
		})
	}
}

// **Ein vollständiger Nutzer und ein vollständiges Profil gehen durch.**
//
// Die Gegenprobe zu den drei Läufen darüber: ohne sie wäre eine Validierung,
// die **alles** abweist, ebenso grün.
func TestEinVollstaendigerNutzerGehtDurch(t *testing.T) {
	t.Parallel()

	if err := (domain.Nutzer{Name: "David"}).Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}

	if err := (domain.Profil{NutzerID: 1, Name: "Arbeit"}).Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Ein Bild ist wahlfrei** — und ein Nutzer ohne Bild ist kein halber Nutzer.
func TestEinNutzerOhneBildGehtDurch(t *testing.T) {
	t.Parallel()

	if err := (domain.Nutzer{Name: "David", Bild: ""}).Validate(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Ein Bildname, der aus dem Bildverzeichnis führte, wird abgewiesen.**
//
// Dieselbe Frage wie beim Paketnamen (#1350): der Wert wird an ein Verzeichnis
// gehängt. Ein `../` darin läse eine Datei außerhalb.
func TestEinBildnameFuehrtNichtHinaus(t *testing.T) {
	t.Parallel()

	for _, bild := range []string{"../heimlich.png", "unter/ordner.png", ".."} {
		t.Run(bild, func(t *testing.T) {
			t.Parallel()

			var feld *domain.FieldError
			if err := (domain.Nutzer{Name: "David", Bild: bild}).Validate(); !errors.As(
				err, &feld) || feld.Field != "bild" {
				t.Errorf("Fehler ist %v", err)
			}
		})
	}
}
