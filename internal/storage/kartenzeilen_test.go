package storage_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Liste ohne die Feldwerte**
// (#891,
// Werkstück 5).
//
// Gemessen an `Black.mse-set` mit 3.788 Karten: `Karten` gibt 4.761.182 Bytes
// JSON, `Kartenzeilen` 245.641 — 19,4-mal weniger, bei 35,4 ms gegen 10,0 ms
// auf der Go-Seite (`BenchmarkKartenliste`, 2026-08-25).
func TestKartenzeilen(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{
			{Field: "name", Text: "Erste"},
			{Field: "type", Text: "Kreatur"},
		}},
		&domain.Card{Values: []domain.Value{{Field: "type", Text: "Land"}}},
		&domain.Card{},
	)

	zeilen, err := storage.NeuerSetLeser(db).Kartenzeilen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Jede Karte kommt vor, auch die ohne Felder.** Genau sie ist der Fall,
	// für den die Zahl mitkommt: „heißt nicht" ist etwas anderes als „ist leer".
	if len(zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet 3", len(zeilen))
	}

	erwartet := []domain.CardLine{
		{ID: karten[0].ID, Number: 0, Name: "Erste", Fields: 2},
		{ID: karten[1].ID, Number: 1, Name: "", Fields: 1},
		{ID: karten[2].ID, Number: 2, Name: "", Fields: 0},
	}

	for i, will := range erwartet {
		if zeilen[i] != will {
			t.Errorf("Zeile %d: %+v, erwartet %+v", i, zeilen[i], will)
		}
	}
}

// **Ein Feld über das ganze Set**, für die Massenänderung.
func TestKartenfeldwerte(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		&domain.Card{Values: []domain.Value{{Field: "type", Text: "Land"}}},
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Dritte"}}},
	)

	werte, err := storage.NeuerSetLeser(db).
		Kartenfeldwerte(probekontext(t), setID, "name")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Nur die Karten, die das Feld haben.** Die zweite steht nicht dabei —
	// eine Zeile mit leerem Wert sagte, das Feld sei leer, und das ist etwas
	// anderes.
	if got := feldwerte(werte); got != fmt.Sprintf(
		"%d=Erste,%d=Dritte", karten[0].ID, karten[2].ID,
	) {
		t.Errorf("Werte %q", got)
	}
}

// **Ein anderes Set bleibt draußen.** Ohne die Bedingung auf `set_id` gäbe die
// Abfrage die Werte jedes Sets im Bestand.
func TestKartenfeldwerteBleibenBeimSet(t *testing.T) {
	db, _, _ := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Fremde"}}},
	)

	zweites, err := storage.NeuerSetSpeicher(db).
		SetAnlegen(probekontext(t), &domain.Set{Name: "Zweites", Game: "magic"}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werte, err := storage.NeuerSetLeser(db).
		Kartenfeldwerte(probekontext(t), zweites, "name")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(werte) != 0 {
		t.Errorf("%d Werte, erwartet keinen: %+v", len(werte), werte)
	}
}

// feldwerte gibt die Werte als `kennung=wert`, durch Komma getrennt.
func feldwerte(werte []domain.CardField) string {
	teile := make([]string, 0, len(werte))

	for _, wert := range werte {
		teile = append(teile, fmt.Sprintf("%d=%s", wert.ID, wert.Text))
	}

	return strings.Join(teile, ",")
}
