package service_test

// **Der Bereichswechsel im Dienst** — Befund 7 des Zweitblicks über Phase 5.1
// (#1405).
//
// **Eine eigene Datei**, damit `deckbau_test.go` unter der Zeilengrenze aus §2
// bleibt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Dienst gibt den Wechsel durch — und den Fehler ebenso.**
//
// **Ohne Spielprüfung**, und der Lauf sagt es: der Platz liegt schon im Deck,
// sein Bezug ändert sich nicht. Ein Kartendoppel, das jede Karte abwiese,
// darf den Wechsel deshalb nicht anhalten.
func TestDerDienstVerschiebtEinenPlatz(t *testing.T) {
	speicher := &deckdoppel{}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	deck := deckAnlegenImDoppel(t, bau)

	gelegt, err := bau.PlatzLegen(t.Context(), deck.ID, domain.Deckplatz{
		Printing: gueltigerDruck, Menge: 4, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	verschoben, err := bau.PlatzVerschieben(
		t.Context(), deck.ID, gelegt.ID, domain.BereichSideboard)
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.Bereich != domain.BereichSideboard || verschoben.Menge != 4 {
		t.Errorf("%+v", verschoben)
	}
}

// **Ein klemmender Bestand kommt als Fehler durch.**
func TestEinKlemmenderBestandHaeltDenWechselAn(t *testing.T) {
	speicher := &deckdoppel{fehler: errors.New("zu")}
	bau := service.NeuerDeckbau(speicher, &kartenspieldoppel{}, &formatdoppel{})

	if _, err := bau.PlatzVerschieben(
		t.Context(), 1, 1, domain.BereichSideboard); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
