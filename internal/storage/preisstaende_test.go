// **Die Preisstände im Bestand** — Werkstück 3 von Phase 3.4
// (#1165).
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.
package storage_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// amTag gibt einen Zeitpunkt am genannten Tag im August 2026.
func amTag(tag int) time.Time {
	return time.Date(2026, time.August, tag, 12, 0, 0, 0, time.UTC)
}

// probestand gibt einen Preisstand am genannten Tag.
func probestand(tag int, wert int64) domain.Preisstand {
	return domain.Preisstand{
		Printing: probedruck,
		Quelle:   domain.PreisquelleScryfall,

		// **Markt, Art und Lage gehören seit A3 zur Reihe** (#1442) — ohne
		// sie steht die Zeile in einer anderen als der, nach der gefragt wird.
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Lage:     domain.PreisGenannt,
		Waehrung: "EUR",
		Finish:   domain.FinishNonfoil,
		Wert:     wert,
		Erhoben:  amTag(tag),
	}
}

// **Was eingetragen wurde, kommt zurück — samt Zeitpunkt.**
func TestEinPreisstandKommtZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	if err := speicher.Eintragen(probekontext(t), probestand(20, 417)); err != nil {
		t.Fatalf("Eintragen: %v", err)
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != 1 {
		t.Fatalf("es sind %d Stände", len(gelesen))
	}

	if gelesen[0].Wert != 417 {
		t.Errorf("der Wert ist %d", gelesen[0].Wert)
	}

	// **Der Zeitpunkt überlebt den Rundlauf.** Ohne ihn ist ein Preis keine
	// Momentaufnahme, sondern eine Behauptung ohne Alter.
	if !gelesen[0].Erhoben.Equal(amTag(20)) {
		t.Errorf("der Zeitpunkt ist %v statt %v", gelesen[0].Erhoben, amTag(20))
	}
}

// **Zwei Abrufe an zwei Tagen sind zwei Zeilen — und der jüngere gilt.**
func TestDerJuengereStandGilt(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	// **Der ältere wird zuletzt eingetragen.** Andersherum stimmte das
	// Ergebnis auch dann, wenn „der neueste" in Wahrheit „der zuletzt
	// geschriebene" wäre.
	for _, stand := range []domain.Preisstand{probestand(25, 500), probestand(20, 417)} {
		if err := speicher.Eintragen(probekontext(t), stand); err != nil {
			t.Fatalf("Eintragen: %v", err)
		}
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != 1 {
		t.Fatalf("es sind %d Stände: %+v", len(gelesen), gelesen)
	}

	if gelesen[0].Wert != 500 {
		t.Errorf("es gilt %d statt 500", gelesen[0].Wert)
	}
}

// **Derselbe Abruf zweimal ist einmal.**
//
// Ein `INSERT` in derselben Sekunde ist kein neuer Stand, sondern derselbe —
// und er darf den Lauf nicht abbrechen.
func TestDerselbeAbrufZweimalIstEineZeile(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	for range 2 {
		if err := speicher.Eintragen(probekontext(t), probestand(20, 417)); err != nil {
			t.Fatalf("Eintragen: %v", err)
		}
	}

	verlauf, err := speicher.Verlauf(probekontext(t), probedruck,
		reihevonScryfall("EUR", domain.FinishNonfoil))
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 1 {
		t.Errorf("es sind %d Zeilen", len(verlauf))
	}
}

// **Währung und Ausführung sind getrennte Reihen.**
//
// Der Euro-Preis eines Drucks sagt nichts über seinen Foil-Preis in Dollar;
// wer beides in eine Reihe legte, hätte einen Verlauf, der zwischen zwei
// Größen springt.
func TestWaehrungUndAusfuehrungSindGetrennteReihen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	for _, stand := range []domain.Preisstand{
		probestand(20, 417),
		{
			Printing: probedruck, Quelle: domain.PreisquelleScryfall,
			Waehrung: "USD", Finish: domain.FinishNonfoil,
			Wert: 499, Erhoben: amTag(20),
		},
		{
			Printing: probedruck, Quelle: domain.PreisquelleScryfall,
			Waehrung: "EUR", Finish: domain.FinishFoil,
			Wert: 1200, Erhoben: amTag(20),
		},
	} {
		if err := speicher.Eintragen(probekontext(t), stand); err != nil {
			t.Fatalf("Eintragen: %v", err)
		}
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(gelesen) != 3 {
		t.Fatalf("es sind %d Reihen: %+v", len(gelesen), gelesen)
	}
}

// **Ein Verlauf steht in der Zeit — ältester zuerst.**
func TestEinVerlaufStehtInDerZeit(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	for _, tag := range []int{25, 20, 23} {
		if err := speicher.Eintragen(probekontext(t), probestand(tag, int64(tag))); err != nil {
			t.Fatalf("Eintragen: %v", err)
		}
	}

	verlauf, err := speicher.Verlauf(probekontext(t), probedruck,
		reihevonScryfall("EUR", domain.FinishNonfoil))
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	for i, erwartet := range []int64{20, 23, 25} {
		if verlauf[i].Wert != erwartet {
			t.Errorf("an Stelle %d steht %d statt %d", i, verlauf[i].Wert, erwartet)
		}
	}
}

// **Ohne Preise gibt es leere Listen und kein `nil`** (§3).
func TestOhnePreiseGibtEsLeereListen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	neueste, err := speicher.Neueste(probekontext(t), "gibtesnicht")
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	verlauf, verlauffehler := speicher.Verlauf(probekontext(t), "gibtesnicht",
		reihevonScryfall("EUR", domain.FinishNonfoil))
	if verlauffehler != nil {
		t.Fatalf("Verlauf: %v", verlauffehler)
	}

	if neueste == nil || verlauf == nil {
		t.Errorf("neueste=%v verlauf=%v", neueste, verlauf)
	}
}

// reihevonScryfall benennt eine Reihe dieser Quelle (A3, #1442).
//
// **Als Helfer, weil sie sieben Teile hat**: sieben Zeilen je Aufruf sagten
// nichts über den Lauf, in dem sie stehen.
func reihevonScryfall(waehrung string, finish domain.Finish) domain.Preisstand {
	return domain.Preisstand{
		Quelle:   domain.PreisquelleScryfall,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Waehrung: waehrung,
		Finish:   finish,
	}
}
