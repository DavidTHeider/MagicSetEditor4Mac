package storage_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei Karten tauschen ihre Stellen**
// (#628).
//
// **ported — die Absicht, nicht der Ausfall.** MSE2 ordnet durch Ziehen um und
// löst dabei je überfahrenem Element einen `ReorderCardsAction` aus, der zwei
// Stellen **tauscht** (`src/gui/control/card_list.cpp:370–388`,
// `src/data/action/set.hpp:56–66`, Fassung f9d9356 vom 2021-07-06).
//
// **Ausgeführt wird der Tausch dort nie.** `ReorderCardsAction::perform`
// prüft `card_id1 >= size || card_id2 < size` und kehrt bei jeder gültigen
// Auswahl zurück, ohne zu tauschen (`src/data/action/set.cpp:54–63`); der
// Kommentar daneben — *„TODO : Too lazy to fix this right now."* — sagt, dass
// der Verfasser es wusste. Portiert ist deshalb die Absicht (§15.4).
func TestZweiKartenTauschenIhreStellen(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Dritte"}}},
	)

	if fehler := storage.NeuerKartenschreiber(db).
		KartenTauschen(probekontext(t), karten[0].ID, karten[1].ID); fehler != nil {
		t.Fatalf("tauschen: %v", fehler)
	}

	danach, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if got := kartennamen(danach); got != "Zweite,Erste,Dritte" {
		t.Errorf("Reihenfolge %q, erwartet `Zweite,Erste,Dritte`", got)
	}

	// **Und die Kennungen sind mitgegangen**, nicht die Werte umgeschrieben.
	if danach[0].ID != karten[1].ID || danach[1].ID != karten[0].ID {
		t.Errorf("Kennungen %d,%d — erwartet %d,%d",
			danach[0].ID, danach[1].ID, karten[1].ID, karten[0].ID)
	}
}

// kartennamen gibt die Namen in ihrer Reihenfolge, durch Komma getrennt.
func kartennamen(karten []*domain.CardRow) string {
	teile := make([]string, 0, len(karten))

	for _, karte := range karten {
		wert, _ := karte.Value("name")
		teile = append(teile, wert)
	}

	return strings.Join(teile, ",")
}

// **Der Tausch ist ein Schritt** — ⌘Z stellt beide Stellen zurück.
func TestEinTauschLaesstSichZuruecknehmen(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
	)

	if fehler := storage.NeuerKartenschreiber(db).
		KartenTauschen(probekontext(t), karten[0].ID, karten[1].ID); fehler != nil {
		t.Fatalf("tauschen: %v", fehler)
	}

	verlauf := storage.NeuerVerlaufsleser(db)
	leser := storage.NeuerSetLeser(db)

	if _, gab, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⌘Z: %v, gab=%v", fehler, gab)
	}

	danach, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if got := kartennamen(danach); got != "Erste,Zweite" {
		t.Errorf("nach ⌘Z %q, erwartet `Erste,Zweite`", got)
	}

	if _, gab, fehler := verlauf.Wiederherstellen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⇧⌘Z: %v, gab=%v", fehler, gab)
	}

	wieder, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if got := kartennamen(wieder); got != "Zweite,Erste" {
		t.Errorf("nach ⇧⌘Z %q, erwartet `Zweite,Erste`", got)
	}
}

// **Karten aus verschiedenen Sets zu tauschen ist ein Befund.**
//
// Ihre Stellen zählen je Set; ein Tausch über die Grenze setzte beide auf eine
// Stelle, die im anderen Set schon besetzt sein kann — und hinterließe zwei
// Sets, deren Ordnung niemand angeordnet hat.
func TestEinTauschUeberSetgrenzenIstEinBefund(t *testing.T) {
	db, _, erste := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
	)

	zweiteID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Zweites", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Fremde"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	fremde, err := storage.NeuerSetLeser(db).Karten(probekontext(t), zweiteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	fehler := storage.NeuerKartenschreiber(db).
		KartenTauschen(probekontext(t), erste[0].ID, fremde[0].ID)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v, erwartet ErrValidation", fehler)
	}
}

// **Eine Karte, die es nicht gibt, ist ein `ErrNotFound`.**
func TestEinTauschMitEinerFehlendenKarteIstEinBefund(t *testing.T) {
	db, _, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
	)

	fehler := storage.NeuerKartenschreiber(db).KartenTauschen(probekontext(t), karten[0].ID, 4711)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("%v, erwartet ErrNotFound", fehler)
	}
}

// **Eine Karte mit sich selbst zu tauschen ist keine Bewegung.**
//
// Sie liefe durch — `IN (?, ?)` gibt eine Zeile, beide Nachschläge finden sie
// —, und heraus käme ein Verlaufsschritt, der nichts tut, den aber ⌘Z
// verbraucht.
func TestEineKarteLaesstSichNichtMitSichSelbstTauschen(t *testing.T) {
	db, _, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
	)

	fehler := storage.NeuerKartenschreiber(db).
		KartenTauschen(probekontext(t), karten[0].ID, karten[0].ID)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v, erwartet ErrValidation", fehler)
	}
}
