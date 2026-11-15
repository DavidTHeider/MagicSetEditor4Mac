package storage_test

// **Ein Turnier trägt sein Format, und eine Vorlage ihre Namen** —
// DOM-04.AK-01 (#1699).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Das Format geht hinein und kommt wieder heraus** — beim Lesen eines
// Turniers und in der Liste.
func TestEinTurnierTraegtSeinFormat(t *testing.T) {
	db := frischeDatenbank(t)
	formate := storage.NeueFormatablage(db)
	ablage := storage.NeueTurnierablage(db)

	format, err := formate.Anlegen(probekontext(t), domain.Format{Name: "Pauper", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: format.ID,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen, err := ablage.Turnier(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.FormatID != format.ID {
		t.Errorf("gelesen trägt Format %d, erwartet %d", gelesen.FormatID, format.ID)
	}

	liste, err := ablage.Turniere(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("die Liste: %v", err)
	}

	if len(liste) != 1 || liste[0].FormatID != format.ID {
		t.Errorf("die Liste nennt %+v, erwartet Format %d", liste, format.ID)
	}
}

// **Ein gelöschtes Format lässt das Turnier stehen** — dieselbe Regel wie
// beim Deck: wer ein Format aufräumt, räumt kein Turnier mit ab.
func TestEinGeloeschtesFormatLaesstDasTurnierStehen(t *testing.T) {
	db := frischeDatenbank(t)
	formate := storage.NeueFormatablage(db)
	ablage := storage.NeueTurnierablage(db)

	format, err := formate.Anlegen(probekontext(t), domain.Format{Name: "Pauper", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das Format: %v", err)
	}

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Freitag", Spiel: "magic", FormatID: format.ID,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if loeschfehler := formate.Loeschen(probekontext(t), format.ID); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	geblieben, err := ablage.Turnier(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("das Turnier ist weg: %v", err)
	}

	if geblieben.FormatID != 0 {
		t.Errorf("das Turnier hält noch das Format %d", geblieben.FormatID)
	}
}

// **Eine Vorlage bringt ihre Namen mit — in einem Zug mit dem Turnier.**
// Ein Turnier ohne die Hälfte seiner Teilnehmenden wäre ein halber Stand;
// scheitert ein Name, entsteht gar nichts.
func TestEinTurnierEntstehtSamtSeinenTeilnehmenden(t *testing.T) {
	ablage := turnierablage(t)

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Freitag", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{{Name: "Anna"}, {Name: "Bert"}},
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen, err := ablage.Turnier(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Teilnehmer) != 2 || gelesen.Teilnehmer[0].Name != "Anna" ||
		gelesen.Teilnehmer[1].Name != "Bert" || gelesen.Teilnehmer[1].Stelle != 2 {
		t.Errorf("gelesen: %+v", gelesen.Teilnehmer)
	}

	// **Und scheitert ein Name, bleibt kein Turnier zurück.** Ein Deck, das
	// es im Profil nicht gibt, weist der Speicher ab — wie beim Anmelden.
	if _, abbruch := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Samstag", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{{Name: "Anna"}, {Name: "Bert", DeckID: 999}},
	}); !errors.Is(abbruch, domain.ErrNotFound) {
		t.Fatalf("Fehler ist %v, erwartet ErrNotFound", abbruch)
	}

	liste, err := ablage.Turniere(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("die Liste: %v", err)
	}

	if len(liste) != 1 {
		t.Errorf("nach dem Abbruch stehen %d Turniere da, erwartet 1", len(liste))
	}
}
