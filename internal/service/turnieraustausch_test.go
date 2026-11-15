package service_test

// **Ein Turnier ausgeben und einlesen** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitPauper gibt ein Turnierwerk samt Formatdoppel und dem Format Pauper.
func mitPauper(t *testing.T) (*turnierwerk, *formatdoppel, domain.Format) {
	t.Helper()

	werk := neuesTurnierwerk(t)
	formate := &formatdoppel{}
	werk.leitung.MitFormaten(formate)

	pauper, err := formate.Anlegen(t.Context(), domain.Format{Name: "Pauper", Spiel: "magic"})
	if err != nil {
		t.Fatalf("Pauper: %v", err)
	}

	return werk, formate, pauper
}

func TestEinTurnierGehtAlsDateiHinUndZurueck(t *testing.T) {
	werk, _, pauper := mitPauper(t)

	angelegt, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: pauper.ID,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werk.anmelden(t, angelegt.ID, "Anna", "Bert")

	roh, err := werk.leitung.Ausgeben(t.Context(), angelegt.ID)
	if err != nil {
		t.Fatalf("ausgeben: %v", err)
	}

	austausch, _, err := domain.ZerlegeTurnier(roh)
	if err != nil || austausch.Turnier.Format != "Pauper" || len(austausch.Turnier.Teilnehmer) != 2 {
		t.Fatalf("die Datei: %+v, %v", austausch, err)
	}

	// **Das Format kommt über seinen Namen zurück** — mit der Kennung dieses
	// Profils.
	austausch.Turnier.Name = "Freitag II"

	eingelesen, befunde, err := werk.leitung.Einlesen(t.Context(), domain.FormeTurnier(austausch.Turnier))
	if err != nil || len(befunde) != 0 || eingelesen.FormatID != pauper.ID || eingelesen.Name != "Freitag II" ||
		len(eingelesen.Teilnehmer) != 2 {
		t.Errorf("eingelesen: %+v, %+v, %v", eingelesen, befunde, err)
	}
}

// **Ein Format, das es hier nicht gibt, ist ein Befund und kein Nein** —
// auch eines, das nur ein anderes Spiel kennt.
func TestEinFremdesFormatKommtAlsBefund(t *testing.T) {
	werk, formate, _ := mitPauper(t)
	if _, err := formate.Anlegen(t.Context(), domain.Format{Name: "Standard", Spiel: "pokemon"}); err != nil {
		t.Fatalf("Standard: %v", err)
	}

	for _, format := range []string{"Vintage", "Standard"} {
		roh := domain.FormeTurnier(domain.Turnierausgabe{Name: format, Spiel: "magic", Format: format})

		eingelesen, befunde, err := werk.leitung.Einlesen(t.Context(), roh)
		if err != nil || eingelesen.FormatID != 0 || len(befunde) != 1 || befunde[0].Schluessel != "format" {
			t.Errorf("%s: %+v, %+v, %v", format, eingelesen, befunde, err)
		}
	}
}

func TestAusgebenUndEinlesenReichenIhreFehlerDurch(t *testing.T) {
	klemmt := errors.New("klemmt")
	mitFormat := domain.FormeTurnier(domain.Turnierausgabe{Name: "T", Spiel: "magic", Format: "Pauper"})

	t.Run("ein Turnier, das es nicht gibt", func(t *testing.T) {
		werk, _, _ := mitPauper(t)
		if _, err := werk.leitung.Ausgeben(t.Context(), 99); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Fehler ist %v", err)
		}
	})

	t.Run("das Format lässt sich nicht nachschlagen", func(t *testing.T) {
		werk, formate, pauper := mitPauper(t)
		angelegt, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{Name: "T", Spiel: "magic", FormatID: pauper.ID})
		if err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		formate.fehler = klemmt

		if _, err := werk.leitung.Ausgeben(t.Context(), angelegt.ID); !errors.Is(err, klemmt) {
			t.Errorf("ausgeben: %v", err)
		}

		if _, _, err := werk.leitung.Einlesen(t.Context(), mitFormat); !errors.Is(err, klemmt) {
			t.Errorf("einlesen: %v", err)
		}
	})

	t.Run("ohne Naht zu den Formaten", func(t *testing.T) {
		werk := neuesTurnierwerk(t)
		angelegt := werk.turnierAnlegen(t)
		werk.turniere.turniere[angelegt.ID].FormatID = 5

		if _, err := werk.leitung.Ausgeben(t.Context(), angelegt.ID); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("ausgeben: %v", err)
		}

		if _, _, err := werk.leitung.Einlesen(t.Context(), mitFormat); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("einlesen: %v", err)
		}
	})

	t.Run("eine kaputte Datei legt nichts an", func(t *testing.T) {
		werk, _, _ := mitPauper(t)
		if _, _, err := werk.leitung.Einlesen(t.Context(), []byte("{")); !errors.Is(err, domain.ErrValidation) ||
			werk.turniere.angelegt != 0 {
			t.Errorf("Fehler ist %v, angelegt %d", err, werk.turniere.angelegt)
		}
	})

	t.Run("der Speicher klemmt", func(t *testing.T) {
		werk, _, _ := mitPauper(t)
		werk.turniere.fehler = klemmt

		if _, _, err := werk.leitung.Einlesen(t.Context(), mitFormat); !errors.Is(err, klemmt) {
			t.Errorf("Fehler ist %v", err)
		}
	})
}

// **Der Bericht nennt das Format beim Namen** — und reicht Fehler durch wie
// die Ausgabe.
func TestDerBerichtKommtVomDienst(t *testing.T) {
	werk, formate, pauper := mitPauper(t)

	angelegt, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: pauper.ID,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	bericht, err := werk.leitung.Bericht(t.Context(), angelegt.ID, domain.Berichttexte{Format: "FORMAT"})
	if err != nil || !strings.Contains(bericht, "# Freitag") || !strings.Contains(bericht, "- FORMAT: Pauper") {
		t.Errorf("der Bericht: %q, %v", bericht, err)
	}

	if _, err := werk.leitung.Bericht(t.Context(), 99, domain.Berichttexte{}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Turnier: %v", err)
	}

	formate.fehler = errors.New("klemmt")
	if _, err := werk.leitung.Bericht(t.Context(), angelegt.ID, domain.Berichttexte{}); err == nil {
		t.Error("der Fehler beim Nachschlagen ging verloren")
	}
}
