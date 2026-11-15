package storage_test

// **Der Tag, den die Quelle nennt, kommt zurück** — Werkstück A3 von Phase 5.3
// (#1442).
//
// MTGJSON führt seine Preise unter einem ISO-8601-**Datum** je Anbieter; ein
// Abruf von heute kann einen Preis von vorgestern bringen. Scryfall nennt
// keinen — dort bleibt das Feld leer, und das ist die ehrliche Angabe.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDasBeobachtungsdatumKommtZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	beobachtet := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)

	mit := probestand(20, 417)
	mit.Quelle = "mtgjson"
	mit.Beobachtet = beobachtet

	if err := speicher.Eintragen(probekontext(t), mit); err != nil {
		t.Fatalf("Eintragen: %v", err)
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != 1 {
		t.Fatalf("%d Stände", len(gelesen))
	}

	if !gelesen[0].Beobachtet.Equal(beobachtet) {
		t.Errorf("beobachtet ist %v, erwartet %v",
			gelesen[0].Beobachtet, beobachtet)
	}
}

// **Und wo keiner genannt ist, steht auch keiner** — die Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber nur, dass irgendein Datum zurückkommt** —
// auch eines, das aus dem Abrufzeitpunkt gemacht wäre (§10).
func TestOhneBeobachtungsdatumBleibtEsLeer(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	if err := speicher.Eintragen(probekontext(t), probestand(20, 417)); err != nil {
		t.Fatalf("Eintragen: %v", err)
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != 1 || !gelesen[0].Beobachtet.IsZero() {
		t.Errorf("beobachtet ist %v, erwartet keines", gelesen[0].Beobachtet)
	}
}

// **Zwei Quellen zum selben Druck sind zwei Reihen.**
//
// Die Zusage von A3: Markt, Preisart, Zustand und Sprache gehören zum
// Schlüssel. Ohne sie liefe ein Ankaufspreis mit einem Verkaufspreis in einer
// Reihe zusammen, und „der neueste Preis" wäre der neueste von beiden.
func TestJedeAngabeDerReiheTrenntImBestand(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	aenderungen := []func(*domain.Preisstand){
		func(p *domain.Preisstand) { p.Quelle = "mtgjson" },
		func(p *domain.Preisstand) { p.Markt = domain.MarktOnline },
		func(p *domain.Preisstand) { p.Art = domain.PreisartAnkauf },
		func(p *domain.Preisstand) { p.Zustand = "NM" },
		func(p *domain.Preisstand) { p.Sprache = "de" },
	}

	staende := make([]domain.Preisstand, 0, len(aenderungen)+1)
	staende = append(staende, probestand(20, 417))

	for _, aendere := range aenderungen {
		anders := probestand(20, 500)
		aendere(&anders)
		staende = append(staende, anders)
	}

	if err := speicher.Eintragen(probekontext(t), staende...); err != nil {
		t.Fatalf("Eintragen: %v", err)
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != len(staende) {
		t.Errorf("%d Reihen statt %d — zwei liefen zusammen: %+v",
			len(gelesen), len(staende), gelesen)
	}
}
