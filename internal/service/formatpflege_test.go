// **Die Formatpflege** — Werkstück 4 von Phase 4.6
// (#1293).
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Format entsteht aus einer Bauart und lässt sich danach ändern.**
//
// **Die Bauart ist ein Vorschlag**: sie bringt nachgelesene Zahlen mit, und
// wer sie ändert, darf das.
func TestEinFormatEntstehtAusEinerBauart(t *testing.T) {
	pflege := service.NeueFormatpflege(&formatdoppel{})

	angelegt, err := pflege.AusBauart(t.Context(),
		domain.BauartKonstruiert, "  Standard  ", " magic ", "STANDARD")
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.Mindestgroesse != 60 || angelegt.Sideboardgroesse != 15 ||
		angelegt.HoechstzahlGleich != 4 {
		t.Errorf("%+v", angelegt)
	}

	// **Geglättet und kleingeschrieben**, bevor es in den Bestand geht.
	if angelegt.Name != "Standard" || angelegt.Spiel != "magic" ||
		angelegt.Scryfallname != "standard" {
		t.Errorf("%+v", angelegt)
	}
}

// **Eine Bauart, die es nicht gibt, ist ein Nein.**
func TestEineUnbekannteBauartWirdAbgewiesen(t *testing.T) {
	pflege := service.NeueFormatpflege(&formatdoppel{})

	if _, err := pflege.AusBauart(t.Context(),
		domain.Bauart("turnier"), "X", "magic", ""); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ein Format ohne Namen kommt nicht durch** — auf beiden Wegen.
func TestEinFormatOhneNamenWirdAbgewiesen(t *testing.T) {
	pflege := service.NeueFormatpflege(&formatdoppel{})

	if _, err := pflege.Anlegen(t.Context(), domain.Format{
		Spiel: "magic",
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("anlegen: Fehler %v, erwartet ErrValidation", err)
	}

	if _, err := pflege.Aendern(t.Context(), domain.Format{
		ID: 1, Fassung: 1, Spiel: "magic",
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ändern: Fehler %v, erwartet ErrValidation", err)
	}

	if _, err := pflege.AusBauart(t.Context(),
		domain.BauartCommander, "  ", "magic", ""); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("aus Bauart: Fehler %v, erwartet ErrValidation", err)
	}
}

// **Lesen, auflisten und löschen reichen durch.**
func TestDieDurchreichendenWegeDerFormatpflege(t *testing.T) {
	speicher := &formatdoppel{}
	pflege := service.NeueFormatpflege(speicher)

	format, err := pflege.Anlegen(t.Context(), domain.Format{
		Name: "Standard", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := pflege.Format(t.Context(), format.ID); err != nil {
		t.Errorf("lesen: %v", err)
	}

	formate, listenfehler := pflege.Formate(t.Context(), "magic")
	if listenfehler != nil || len(formate) != 1 {
		t.Errorf("die Liste: %v, %d Formate", listenfehler, len(formate))
	}

	if err := pflege.Loeschen(t.Context(), format.ID); err != nil {
		t.Errorf("löschen: %v", err)
	}
}

// **Und jeder Weg meldet den Fehler seines Speichers.**
func TestJederFormatwegMeldetSeinenSpeicherfehler(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")
	pflege := service.NeueFormatpflege(&formatdoppel{fehler: kaputt})

	if _, err := pflege.Anlegen(t.Context(), domain.Format{
		Name: "F", Spiel: "magic",
	}); !errors.Is(err, kaputt) {
		t.Error("anlegen schwieg")
	}

	if _, err := pflege.AusBauart(t.Context(),
		domain.BauartLimitiert, "F", "magic", ""); !errors.Is(err, kaputt) {
		t.Error("aus Bauart schwieg")
	}

	if _, err := pflege.Aendern(t.Context(), domain.Format{
		ID: 1, Fassung: 1, Name: "F", Spiel: "magic",
	}); !errors.Is(err, kaputt) {
		t.Error("ändern schwieg")
	}

	if _, err := pflege.Format(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Error("lesen schwieg")
	}

	if _, err := pflege.Formate(t.Context(), ""); !errors.Is(err, kaputt) {
		t.Error("die Liste schwieg")
	}

	if err := pflege.Loeschen(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Error("löschen schwieg")
	}
}

// **Die Bauarten kommen mit ihren Zahlen heraus** — eine Fläche, die zur Wahl
// stellt, braucht beides.
func TestDieBauartenKommenMitIhrenZahlen(t *testing.T) {
	vorschlaege := service.NeueFormatpflege(&formatdoppel{}).Bauarten()

	if len(vorschlaege) != 3 {
		t.Fatalf("%d Bauarten, erwartet 3", len(vorschlaege))
	}

	for _, vorschlag := range vorschlaege {
		if vorschlag.Bauart == "" {
			t.Error("eine Bauart ohne Namen")
		}

		if vorschlag.Format.Mindestgroesse == 0 {
			t.Errorf("%q bringt keine Zahlen mit", vorschlag.Bauart)
		}
	}
}
