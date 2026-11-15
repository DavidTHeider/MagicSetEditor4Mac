package service_test

// **Format und Vorlage eines Turniers** — DOM-04.AK-01 und AK-03
// (#1699).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitFormat gibt ein Turnierwerk samt Formaten — eines je Spiel.
func mitFormat(t *testing.T) (werk *turnierwerk, pauper, fremdes domain.Format) {
	t.Helper()

	werk = neuesTurnierwerk(t)
	formate := &formatdoppel{}
	werk.leitung.MitFormaten(formate)

	pauper, err := formate.Anlegen(t.Context(), domain.Format{Name: "Pauper", Spiel: "magic"})
	if err != nil {
		t.Fatalf("Pauper: %v", err)
	}

	fremdes, err = formate.Anlegen(t.Context(), domain.Format{Name: "Standard", Spiel: "pokemon"})
	if err != nil {
		t.Fatalf("Standard: %v", err)
	}

	return werk, pauper, fremdes
}

func TestEinTurnierNimmtNurEinFormatSeinesSpiels(t *testing.T) {
	werk, pauper, fremdes := mitFormat(t)

	angelegt, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: pauper.ID,
	})
	if err != nil || angelegt.FormatID != pauper.ID {
		t.Fatalf("mit Pauper: %+v, %v", angelegt, err)
	}

	// **Ein Format aus einem anderen Spiel ist ein Nein am Feld** — mit
	// eigenem Schlüssel, denn die Abhilfe ist ein anderes Format.
	_, err = werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Samstag", Spiel: "magic", FormatID: fremdes.ID,
	})

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Field != "formatID" ||
		feldfehler.Key != domain.SchluesselTurnierformatSpiel {
		t.Errorf("Fehler ist %v, erwartet ein Nein am Feld formatID", err)
	}

	// **Ein Format, das es nicht gibt, ist nicht gefunden** — und nichts
	// ist angelegt.
	if _, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Sonntag", Spiel: "magic", FormatID: 99,
	}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	if werk.turniere.angelegt != 1 {
		t.Errorf("%d Turniere angelegt, erwartet 1", werk.turniere.angelegt)
	}
}

// **Ohne Naht zu den Formaten wird keines angenommen** — statt es
// ungeprüft abzulegen.
func TestOhneFormatnahtNimmtEinTurnierKeinFormat(t *testing.T) {
	werk := neuesTurnierwerk(t)

	if _, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: 1,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet ErrValidation", err)
	}
}

// vorlage legt ein gespieltes Turnier an: Format, Einzelspiel, Notiz, zwei
// Teilnehmende mit Deckabzug, eine ausgestiegen, und eine Runde.
func vorlage(t *testing.T, werk *turnierwerk, format domain.Format) domain.Turnier {
	t.Helper()

	turnier, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: format.ID,
		Einzelspiel: true, Notiz: "im Hinterzimmer",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werk.decks.deck = domain.Deck{ID: 7, Spiel: "magic", Plaetze: []domain.Deckplatz{
		{Printing: "druck-1", Menge: 4, Bereich: domain.BereichHauptdeck},
	}}
	if _, err := werk.leitung.Anmelden(t.Context(), turnier.ID,
		domain.Teilnehmer{Name: "Anna", DeckID: 7}); err != nil {
		t.Fatalf("Anna: %v", err)
	}

	bert := werk.anmelden(t, turnier.ID, "Bert")[0]
	if err := werk.leitung.Aussteigen(t.Context(), turnier.ID, bert, true); err != nil {
		t.Fatalf("Bert steigt aus: %v", err)
	}

	werk.anmelden(t, turnier.ID, "Cleo")

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); err != nil {
		t.Fatalf("die Runde: %v", err)
	}

	return turnier
}

// **Eine Vorlage bringt ihre Einstellungen und Namen mit — und sonst nichts.**
// Kein Deck und kein Abzug: eine eingereichte Liste ist die Zusage eines
// Menschen für **ein** Turnier. Keine Runde, kein Ausstieg.
func TestEineVorlageBringtEinstellungenUndNamenMit(t *testing.T) {
	werk, pauper, _ := mitFormat(t)
	quelle := vorlage(t, werk, pauper)

	kopie, err := werk.leitung.Kopieren(t.Context(), quelle.ID, "Freitag II", true)
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	gelesen, err := werk.leitung.Turnier(t.Context(), kopie.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.ID == quelle.ID || gelesen.Name != "Freitag II" || gelesen.Spiel != "magic" ||
		gelesen.FormatID != pauper.ID || !gelesen.Einzelspiel || gelesen.Notiz != "im Hinterzimmer" {
		t.Errorf("die Einstellungen: %+v", gelesen)
	}

	if len(gelesen.Runden) != 0 {
		t.Errorf("die Kopie trägt %d Runden", len(gelesen.Runden))
	}

	namen := make([]string, 0, len(gelesen.Teilnehmer))
	for _, teilnehmer := range gelesen.Teilnehmer {
		namen = append(namen, teilnehmer.Name)

		if teilnehmer.DeckID != 0 || len(teilnehmer.Deckabzug) != 0 || teilnehmer.Ausgestiegen {
			t.Errorf("%s bringt mehr als den Namen mit: %+v", teilnehmer.Name, teilnehmer)
		}
	}

	if len(namen) != 3 || namen[0] != "Anna" || namen[1] != "Bert" || namen[2] != "Cleo" {
		t.Errorf("die Namen: %v", namen)
	}
}

// **Ohne Teilnehmende nur die Einstellungen.**
func TestEineVorlageGehtAuchOhneNamen(t *testing.T) {
	werk, pauper, _ := mitFormat(t)
	quelle := vorlage(t, werk, pauper)

	kopie, err := werk.leitung.Kopieren(t.Context(), quelle.ID, "Freitag II", false)
	if err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	if len(kopie.Teilnehmer) != 0 || kopie.FormatID != pauper.ID {
		t.Errorf("die Kopie: %+v", kopie)
	}
}

// **Eine Kopie braucht einen Namen, und die Vorlage muss es geben.**
func TestEineKopieBrauchtNamenUndVorlage(t *testing.T) {
	werk, pauper, _ := mitFormat(t)
	quelle := vorlage(t, werk, pauper)

	var feldfehler *domain.FieldError
	if _, err := werk.leitung.Kopieren(t.Context(), quelle.ID, " ", true); !errors.As(err, &feldfehler) ||
		feldfehler.Field != "name" {
		t.Errorf("ohne Namen: %v, erwartet ein Nein am Feld name", err)
	}

	if _, err := werk.leitung.Kopieren(t.Context(), 99, "Freitag II", true); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ohne Vorlage: %v, erwartet ErrNotFound", err)
	}
}
