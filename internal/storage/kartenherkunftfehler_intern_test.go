package storage

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine zurückgebrachte Herkunft meldet ihren Fehlerweg** —
// #1036.
//
// # Warum das kein erfundener Fall ist
//
// `wert_herkunft` hängt über `ON DELETE CASCADE` an `cards`. Eine gelöschte
// Karte nimmt ihre Nennungen mit; zurück kommen sie nur, weil das Löschen sie
// vorher gelesen und in den Rückweg gelegt hat.
//
// Der Weg liegt **hinter** dem Feldwert: die Karte steht dann schon wieder da,
// ihr Wert auch. Ein stiller Fehler hier hinterließe genau den halben Stand,
// den Migration 0014 ausschließt — ein Wert ohne seine Herkunft.
func TestDasZurueckbringenDerKartenherkunftMeldetSeinenWeg(t *testing.T) {
	db, err := sql.Open("stolper-kartenherkunft-schreiben", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	tx, txfehler := db.BeginTx(probekontext(t), nil)
	if txfehler != nil {
		t.Fatalf("Transaktion: %v", txfehler)
	}

	stand := &domain.Kartenstand{GabEs: true, Felder: []domain.Feldstand{
		{Feld: "kunst", Wert: "x", Herkunft: "https://scryfall.com/card/lea/232"},
	}}

	fehler := ganzeKarteGehen(probekontext(t), tx, 1, 1, stand)
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), `die Herkunft von "kunst" der Karte 1 zurückbringen`) {
		t.Errorf("Fehler %v", fehler)
	}
}
