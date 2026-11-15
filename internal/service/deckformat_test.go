package service_test

// **Ein Deck auf ein Format stellen** — Werkstück 4 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `deckbau_test.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, wie ein Deck entsteht und
// was hineinkommt; hier, wogegen es geprüft werden soll.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Deck lässt sich auf ein Format desselben Spiels stellen.**
func TestEinDeckLaesstSichAufEinFormatSeinesSpielsStellen(t *testing.T) {
	formate, bau := deckbauMitFormaten(t)

	deck := deckAnlegenImDoppel(t, bau)

	format, err := formate.Anlegen(t.Context(), domain.Format{
		Name: "Standard", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	gestellt, err := bau.FormatWaehlen(t.Context(), deck.ID, format.ID)
	if err != nil {
		t.Fatalf("stellen: %v", err)
	}

	if gestellt.FormatID != format.ID {
		t.Errorf("FormatID %d, erwartet %d", gestellt.FormatID, format.ID)
	}

	if gestellt.Fassung != 2 {
		t.Errorf("Fassung %d, erwartet 2", gestellt.Fassung)
	}
}

// **Ein Format eines anderen Spiels wird abgewiesen.**
//
// **Ohne diese Prüfung ließe sich ein Magic-Deck gegen ein VS-Format
// stellen** — und die Antwort der Prüfung wäre Unsinn, nicht falsch.
func TestEinFormatEinesAnderenSpielsWirdAbgewiesen(t *testing.T) {
	formate, bau := deckbauMitFormaten(t)

	deck := deckAnlegenImDoppel(t, bau)

	fremd, err := formate.Anlegen(t.Context(), domain.Format{
		Name: "Golden Age", Spiel: "vs",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	_, err = bau.FormatWaehlen(t.Context(), deck.ID, fremd.ID)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet ErrValidation", err)
	}

	var feld *domain.FieldError
	if !errors.As(err, &feld) || feld.Field != "formatID" {
		t.Errorf("der Befund nennt das Feld nicht: %v", err)
	}
}

// **Null nimmt das Format wieder herunter** — und fragt keinen Speicher.
//
// Ein Deck ohne Format ist **ungeprüft**, nicht falsch.
func TestNullNimmtDasFormatHerunter(t *testing.T) {
	formate, bau := deckbauMitFormaten(t)

	deck := deckAnlegenImDoppel(t, bau)

	format, err := formate.Anlegen(t.Context(), domain.Format{
		Name: "Standard", Spiel: "magic",
	})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	if _, err := bau.FormatWaehlen(t.Context(), deck.ID, format.ID); err != nil {
		t.Fatalf("stellen: %v", err)
	}

	// **Der Speicher darf jetzt schweigen**: null fragt ihn gar nicht erst.
	formate.fehler = errors.New("der Formatspeicher schweigt")

	herunter, ablegefehler := bau.FormatWaehlen(t.Context(), deck.ID, 0)
	if ablegefehler != nil {
		t.Fatalf("herunterstellen: %v", ablegefehler)
	}

	if herunter.FormatID != 0 {
		t.Errorf("das Format blieb: %d", herunter.FormatID)
	}
}

// **Ein Deck, das es nicht gibt, lässt sich auf nichts stellen.**
func TestEinFehlendesDeckLaesstSichNichtStellen(t *testing.T) {
	_, bau := deckbauMitFormaten(t)

	if _, err := bau.FormatWaehlen(t.Context(), 999, 0); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Format, das es nicht gibt, ebenso.**
func TestEinFehlendesFormatLaesstSichNichtWaehlen(t *testing.T) {
	_, bau := deckbauMitFormaten(t)

	deck := deckAnlegenImDoppel(t, bau)

	if _, err := bau.FormatWaehlen(t.Context(), deck.ID, 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// deckbauMitFormaten baut einen Deckbau, der Formate kennt.
func deckbauMitFormaten(t *testing.T) (*formatdoppel, *service.Deckbau) {
	t.Helper()

	formate := &formatdoppel{}

	return formate, service.NeuerDeckbau(&deckdoppel{}, &kartenspieldoppel{}, formate)
}
