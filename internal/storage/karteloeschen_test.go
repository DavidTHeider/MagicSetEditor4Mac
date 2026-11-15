package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine gelöschte Karte kommt mit derselben Kennung zurück**
// (#621).
//
// # Warum die Kennung zählt
//
// **Die Verlaufszeilen zeigen über `karte_id` auf die Karte.** Käme sie mit
// einer neuen zurück, hingen alle früheren Feldänderungen ins Leere — und ein
// zweites ⌘Z führte auf eine Karte, die es nicht mehr gibt.
//
// Das ist der Grund, aus dem
// #562 das
// Löschen vertagt hat: *„eine gelöschte Karte ohne ⌘Z ist Datenverlust mit
// einem Klick."*
//
// # Warum das Löschen eine Lücke lässt
//
// `cards` trägt `UNIQUE (set_id, reihenfolge)`. Rückten die übrigen Karten
// nach, müsste die Rücknahme sie alle zurückschieben — und zwischen Löschen
// und Rücknahme könnte jemand eine Karte anlegen, die den Platz einnimmt.
//
// **Die Reihenfolge ist ein Sortierschlüssel, keine dichte Zählung.**
func TestEineGeloeschteKarteKommtMitDerselbenKennungZurueck(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
			{Values: []domain.Value{
				{Field: "name", Text: "Zweite"},
				{Field: "rule text", Text: "Füge drei Schaden zu."},
			}},
			{Values: []domain.Value{{Field: "name", Text: "Dritte"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	mitte := karten[1]

	if fehler := storage.NeuerKartenschreiber(db).KarteLoeschen(probekontext(t), mitte.ID); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	// **Weg, und die Nachbarn stehen, wo sie standen.**
	nachher, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(nachher) != 2 {
		t.Fatalf("%d Karten, erwartet 2", len(nachher))
	}

	if nachher[0].ID != karten[0].ID || nachher[1].ID != karten[2].ID {
		t.Errorf("übrig sind %d und %d, erwartet %d und %d",
			nachher[0].ID, nachher[1].ID, karten[0].ID, karten[2].ID)
	}

	// **Und ⌘Z bringt sie zurück — dieselbe Kennung, dieselbe Stelle, alle
	// Felder.**
	if _, gab, fehler := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("zurücknehmen: %v, gab=%v", fehler, gab)
	}

	zurueck, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(zurueck) != 3 {
		t.Fatalf("%d Karten, erwartet 3", len(zurueck))
	}

	if zurueck[1].ID != mitte.ID {
		t.Errorf("die Karte kam als %d zurück, erwartet %d", zurueck[1].ID, mitte.ID)
	}

	if len(zurueck[1].Values) != 2 {
		t.Fatalf("%d Feldwerte, erwartet 2", len(zurueck[1].Values))
	}

	if zurueck[1].Values[0].Field != "name" || zurueck[1].Values[0].Text != "Zweite" {
		t.Errorf("erstes Feld %+v, erwartet name/Zweite", zurueck[1].Values[0])
	}

	if zurueck[1].Values[1].Field != "rule text" {
		t.Errorf("zweites Feld %q, erwartet `rule text` — die Stellen sind nicht zurück",
			zurueck[1].Values[1].Field)
	}
}

// **Eine Karte, die es nicht gibt, ist ein Befund.**
func TestEineKarteDieEsNichtGibtLaesstSichNichtLoeschen(t *testing.T) {
	db := frischeDatenbank(t)

	err := storage.NeuerKartenschreiber(db).KarteLoeschen(probekontext(t), 4711)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **⇧⌘Z löscht sie wieder, und ein zweites ⌘Z geht auf die Änderung davor.**
//
// # Warum beides in einem Test steht
//
// Weil es dieselbe Kette ist. Ein Verlauf, der nach der Rücknahme des Löschens
// nicht weiterläuft, wäre am gefährlichsten genau dort: die Karte ist zurück,
// und der Schritt davor gehört ihr — wenn die Kennung stimmt.
func TestNachDerRuecknahmeGehtDerVerlaufWeiter(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Alt"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	erste, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	karteID := erste[0].ID

	schreiber := storage.NeuerKartenschreiber(db)

	// Erst eine Feldänderung, dann das Löschen.
	// **`FeldSetzenMitVorwert` und nicht `FeldSetzen`**: nur der erste schreibt
	// einen Verlaufsschritt, und ohne den prüfte dieser Test nichts.
	if _, _, fehler := schreiber.FeldSetzenMitVorwert(probekontext(t), karteID, "name", "Neu"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if fehler := schreiber.KarteLoeschen(probekontext(t), karteID); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	verlauf := storage.NeuerVerlaufsleser(db)
	leser := storage.NeuerSetLeser(db)

	// ⌘Z bringt die Karte zurück — mit dem **neuen** Wert, denn die
	// Feldänderung gilt noch.
	if _, gab, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("erstes Zurücknehmen: %v, gab=%v", fehler, gab)
	}

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil || len(karten) != 1 {
		t.Fatalf("nach dem ersten ⌘Z: %d Karten, %v", len(karten), err)
	}

	if karten[0].Values[0].Text != "Neu" {
		t.Errorf("Wert %q, erwartet `Neu` — die Feldänderung gilt noch",
			karten[0].Values[0].Text)
	}

	// **Ein zweites ⌘Z geht auf die Feldänderung**, nicht ins Leere.
	if _, gab, fehler := verlauf.Zuruecknehmen(probekontext(t), setID); fehler != nil || !gab {
		t.Fatalf("zweites Zurücknehmen: %v, gab=%v", fehler, gab)
	}

	karten, err = leser.Karten(probekontext(t), setID)
	if err != nil || len(karten) != 1 {
		t.Fatalf("nach dem zweiten ⌘Z: %d Karten, %v", len(karten), err)
	}

	if karten[0].Values[0].Text != "Alt" {
		t.Errorf("Wert %q, erwartet `Alt`", karten[0].Values[0].Text)
	}

	// **Und ⇧⌘Z zweimal löscht sie wieder.**
	for i := range 2 {
		if _, gab, fehler := verlauf.Wiederherstellen(probekontext(t), setID); fehler != nil || !gab {
			t.Fatalf("Wiederherstellen %d: %v, gab=%v", i+1, fehler, gab)
		}
	}

	if karten, err := leser.Karten(probekontext(t), setID); err != nil || len(karten) != 0 {
		t.Errorf("nach zweimal ⇧⌘Z: %d Karten, %v — erwartet keine", len(karten), err)
	}
}

// **Ist die Stelle besetzt, scheitert die Rücknahme — laut.**
//
// # Warum das der richtige Ausgang ist
//
// `cards` trägt `UNIQUE (set_id, reihenfolge)`. Die Lücke, die das Löschen
// lässt, ist normalerweise frei — `KarteAnlegen` hängt hinten an. Wer sie
// trotzdem besetzt (ein Import, ein zweiter Schreiber), macht die Rücknahme
// unmöglich.
//
// **Dann ist ein Fehler die freundlichste Antwort.** Die Karte woanders
// hinzusetzen wäre keine Rücknahme, sondern eine neue Änderung unter falschem
// Namen — und niemand sähe, dass die Ordnung eine andere ist als vorher.
func TestEineBesetzteStelleMachtDieRuecknahmeZuEinemFehler(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if fehler := storage.NeuerKartenschreiber(db).
		KarteLoeschen(probekontext(t), karten[0].ID); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	// **Die Lücke besetzen** — so, wie es ein zweiter Schreiber täte.
	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (?, 0)`, setID); fehler != nil {
		t.Fatalf("die Lücke besetzen: %v", fehler)
	}

	if _, _, fehler := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); fehler == nil {
		t.Error("die Rücknahme meldete keinen Fehler, obwohl die Stelle besetzt ist")
	}
}
