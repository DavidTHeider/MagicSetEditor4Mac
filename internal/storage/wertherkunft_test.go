package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Herkunft des Werts, der jetzt dasteht** —
// #1025.
//
// # Warum der Verlauf dafür nicht reicht
//
// Der Verlauf ist ein **Protokoll**: er sagt, dass irgendwann einmal ein Wert
// von Scryfall kam. Er sagt nicht, ob er noch dasteht. Wer die Nennung eines
// Bildes daraus läse, müsste den jüngsten Eintrag je Feld suchen und darauf
// hoffen, dass niemand den Verlauf gekürzt hat — und läge falsch, sobald ein
// eigenes Foto darüber geschrieben wurde.
//
// **Die Zusage lautet deshalb: was in `card_values` steht, und was in
// `wert_herkunft` steht, gehören zusammen.** Beides wird in derselben
// Transaktion geführt; auseinanderlaufen können sie nicht.
func TestDerAktuelleWertTraegtSeineHerkunft(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)

	if _, _, fehler := schreiber.FeldSetzenVonAussen(probekontext(t), karteID,
		"kunst", "lotus.png", "https://scryfall.com/card/lea/232"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Selbst getippt"); fehler != nil {
		t.Fatalf("FeldSetzen: %v", fehler)
	}

	herkuenfte, fehler := storage.NeuerSetLeser(db).Herkuenfte(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("Herkuenfte: %v", fehler)
	}

	if herkuenfte["kunst"] != "https://scryfall.com/card/lea/232" {
		t.Errorf("Herkunft von »kunst« = %q, erwartet die Kartenseite bei Scryfall — "+
			"ohne sie weiß niemand mehr, dass zu diesem Bild eine Nennung gehört",
			herkuenfte["kunst"])
	}

	if _, da := herkuenfte["name"]; da {
		t.Errorf("»name« trägt eine Herkunft (%q), obwohl der Wert von hier kam",
			herkuenfte["name"])
	}
}

// **Ein eigener Wert löscht die fremde Herkunft** — und das ist die Hälfte,
// die man vergisst.
//
// Bliebe die Zeile stehen, behauptete die Anwendung über ein selbst gewähltes
// Foto, es sei ein Bildausschnitt von Scryfall, samt fremdem Copyright. Eine
// falsche Nennung ist schlimmer als keine.
func TestEinEigenerWertNimmtDieHerkunftMit(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Probe"})

	schreiber := storage.NeuerKartenschreiber(db).Als(domain.UrheberFenster)

	if _, _, fehler := schreiber.FeldSetzenVonAussen(probekontext(t), karteID,
		"kunst", "lotus.png", "https://scryfall.com/card/lea/232"); fehler != nil {
		t.Fatalf("FeldSetzenVonAussen: %v", fehler)
	}

	if _, _, fehler := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "kunst", "eigenes-foto.png"); fehler != nil {
		t.Fatalf("FeldSetzen: %v", fehler)
	}

	herkuenfte, fehler := storage.NeuerSetLeser(db).Herkuenfte(probekontext(t), karteID)
	if fehler != nil {
		t.Fatalf("Herkuenfte: %v", fehler)
	}

	if _, da := herkuenfte["kunst"]; da {
		t.Errorf("»kunst« trägt weiter %q, obwohl darüber ein eigenes Bild "+
			"geschrieben wurde", herkuenfte["kunst"])
	}
}
