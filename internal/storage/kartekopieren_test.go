package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probeset legt ein Set mit den gegebenen Karten an.
func probeset(t *testing.T, karten ...*domain.Card) (db *sql.DB, setID int64, gelesen []*domain.CardRow) {
	t.Helper()

	db = frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic", Cards: karten,
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	gelesen, err = storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return db, setID, gelesen
}

// **Die Kopie trägt alle Werte und steht am Ende**
// (#625).
//
// **ported.** MSE2 fügt eine kopierte Karte über `AddCardAction(ADD, …)` ein,
// und das nimmt `pos = container.size()` — das Ende, nicht die Stelle neben
// dem Original (`src/data/action/generic.hpp:54–56`, Fassung f9d9356 vom
// 2021-07-06). Mitgereist ist dabei die **ganze** Karte: die Zwischenablage
// schreibt sie über dieselbe Reflexion wie eine `.mse-set`
// (`src/data/format/clipboard.cpp:44–79`).
func TestEineKopieTraegtAlleWerteUndStehtAmEnde(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		&domain.Card{Values: []domain.Value{
			{Field: "name", Text: "Zweite"},
			{Field: "rule text", Text: "Füge drei Schaden zu."},
			{Field: "has styling", Text: "true"},
		}},
	)

	kennung, fehler := storage.NeuerKartenschreiber(db).KarteKopieren(probekontext(t), karten[1].ID)
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if kennung == karten[1].ID {
		t.Fatalf("die Kopie trägt die Kennung des Originals: %d", kennung)
	}

	danach, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(danach) != 3 {
		t.Fatalf("%d Karten, erwartet 3", len(danach))
	}

	kopie := danach[2]

	if kopie.ID != kennung {
		t.Errorf("am Ende steht %d, kopiert wurde nach %d", kopie.ID, kennung)
	}

	if len(kopie.Values) != len(karten[1].Values) {
		t.Fatalf("%d Werte, erwartet %d: %+v", len(kopie.Values), len(karten[1].Values), kopie.Values)
	}

	for stelle, wert := range karten[1].Values {
		if kopie.Values[stelle] != wert {
			t.Errorf("Wert %d: %+v, erwartet %+v", stelle, kopie.Values[stelle], wert)
		}
	}
}

// **Und sie lässt sich zurücknehmen** — ein Schritt, nicht zwei.
func TestEineKopieLaesstSichZuruecknehmen(t *testing.T) {
	db, setID, karten := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
	)

	kennung, fehler := storage.NeuerKartenschreiber(db).KarteKopieren(probekontext(t), karten[0].ID)
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	verlauf := storage.NeuerVerlaufsleser(db)
	leser := storage.NeuerSetLeser(db)

	if _, gab, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⌘Z: %v, gab=%v", fehler, gab)
	}

	if danach, err := leser.Karten(probekontext(t), setID); err != nil || len(danach) != 1 {
		t.Fatalf("nach ⌘Z: %d Karten, %v — erwartet eine", len(danach), err)
	}

	// **⇧⌘Z bringt sie mit derselben Kennung zurück**, samt ihrem Wert.
	if _, gab, fehler := verlauf.Wiederherstellen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⇧⌘Z: %v, gab=%v", fehler, gab)
	}

	danach, err := leser.Karten(probekontext(t), setID)
	if err != nil || len(danach) != 2 {
		t.Fatalf("nach ⇧⌘Z: %d Karten, %v", len(danach), err)
	}

	if danach[1].ID != kennung {
		t.Errorf("Kennung %d, erwartet %d", danach[1].ID, kennung)
	}

	if danach[1].Values[0].Text != "Erste" {
		t.Errorf("Wert %q, erwartet `Erste`", danach[1].Values[0].Text)
	}
}

// **Eine Karte, die es nicht gibt, ist ein Befund** und kein stiller Erfolg.
func TestEineKarteDieEsNichtGibtLaesstSichNichtKopieren(t *testing.T) {
	db := frischeDatenbank(t)

	_, fehler := storage.NeuerKartenschreiber(db).KarteKopieren(probekontext(t), 4711)
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("%v, erwartet ErrNotFound", fehler)
	}
}

// **Auch das Anlegen schreibt einen Verlaufsschritt**
// (#626).
//
// **ported.** In MSE2 ist Anlegen dieselbe Handlung wie Einfügen —
// `AddCardAction(Set&)` legt eine frische Karte an und reiht sie in den
// Handlungsstapel ein (`src/data/action/set.cpp:20–22`, dieselbe Fassung). Es
// gibt dort keinen Weg, eine Karte anzulegen, ohne dass es im Verlauf steht.
//
// Hier gab es ihn: `KarteAnlegen` setzte ein `INSERT` ab und sonst nichts, und
// ⌘Z antwortete mit „nichts zurückzunehmen".
func TestEineAngelegteKarteLaesstSichZuruecknehmen(t *testing.T) {
	db, setID, _ := probeset(t,
		&domain.Card{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
	)

	kennung, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	verlauf := storage.NeuerVerlaufsleser(db)
	leser := storage.NeuerSetLeser(db)

	if _, gab, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⌘Z: %v, gab=%v", fehler, gab)
	}

	if danach, err := leser.Karten(probekontext(t), setID); err != nil || len(danach) != 1 {
		t.Fatalf("nach ⌘Z: %d Karten, %v — erwartet eine", len(danach), err)
	}

	if _, gab, fehler := verlauf.Wiederherstellen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("⇧⌘Z: %v, gab=%v", fehler, gab)
	}

	danach, err := leser.Karten(probekontext(t), setID)
	if err != nil || len(danach) != 2 {
		t.Fatalf("nach ⇧⌘Z: %d Karten, %v", len(danach), err)
	}

	if danach[1].ID != kennung {
		t.Errorf("Kennung %d, erwartet %d", danach[1].ID, kennung)
	}
}
