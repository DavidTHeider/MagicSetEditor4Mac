package storage_test

// **Die eigenen Stile der Karten eines Sets** — #2423. Der Abdruck einer
// Problemprüfung (PERF02) faltet die Pakete jedes Stils, aus dem gezeichnet
// wird; ein Stil an der Karte (`stylesheet:`) gehört dazu.

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieKartenstileEinesSetsLassenSichLesen(t *testing.T) {
	db := frischeDatenbank(t)

	karte := func(stil string) *domain.Card {
		werte := []domain.Value{{Field: "name", Text: "Karte"}}
		if stil != "" {
			werte = append(werte, domain.Value{Field: "stylesheet", Text: stil})
		}

		return &domain.Card{Values: werte}
	}

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{karte("archenemy"), karte(""), karte("zweit"), karte("archenemy")},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	stile, err := storage.NeuerSetLeser(db).Kartenstile(probekontext(t), setID)
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(stile)

	if !slices.Equal(stile, []string{"archenemy", "zweit"}) {
		t.Errorf("die Kartenstile sind %v", stile)
	}

	// **Nur aus dem gezeigten Profil** (#1815) — ein fremdes Set gibt es nicht.
	if _, err := storage.NeuerSetLeser(db).Kartenstile(probekontext(t), setID+99); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fremdes Set: %v", err)
	}
}
