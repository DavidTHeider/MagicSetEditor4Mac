// Eine Massenänderung ist **ein** Schritt im Verlauf.
//
// # Warum das kein Beiwerk ist
//
// Ohne diesen Weg schriebe die Massenänderung gar keinen Verlauf — und ⌘Z nähme
// danach die **vorige** Änderung zurück, während die dreihundert geänderten
// Karten stehen blieben. Das ist schlimmer als ein ⌘Z, das nichts tut: es sieht
// aus, als hätte es gewirkt (#52).
//
// Und ein Eintrag **je Karte** wäre die andere falsche Antwort: wer dreihundert
// Karten in einem Zug ändert, will sie in einem Zug zurücknehmen, nicht
// dreihundertmal — und bei `GrenzeVerlauf` von 200 käme er nicht einmal
// vollständig zurück.
package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandMitDreiKarten legt ein Set mit drei benannten Karten an.
func bestandMitDreiKarten(t *testing.T) (db *sql.DB, setID int64, karten []int64) {
	t.Helper()

	datenbank := frischeDatenbank(t)

	nummer, err := storage.NeuerSetSpeicher(datenbank).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}},
			{Values: []domain.Value{{Field: "name", Text: "Eisball"}}},
			{Values: []domain.Value{{Field: "name", Text: "Steinwand"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	zeilen, err := storage.NeuerSetLeser(datenbank).Karten(probekontext(t), nummer)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	kennungen := make([]int64, 0, len(zeilen))
	for _, zeile := range zeilen {
		kennungen = append(kennungen, zeile.ID)
	}

	return datenbank, nummer, kennungen
}

// werte liest ein Feld aller Karten eines Sets, in ihrer Reihenfolge.
func werte(t *testing.T, db *sql.DB, setID int64, feld string) []string {
	t.Helper()

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	gelesen := make([]string, 0, len(karten))

	for _, karte := range karten {
		wert, gabEs := karte.Value(feld)
		if !gabEs {
			wert = "«fehlt»"
		}

		gelesen = append(gelesen, wert)
	}

	return gelesen
}

func TestEineMassenaenderungIstEinSchritt(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	if _, err := storage.NeuerKartenschreiber(db).
		FelderSetzen(probekontext(t), karten, "typ", "Artefakt"); err != nil {
		t.Fatalf("Massenänderung: %v", err)
	}

	if gelesen := werte(t, db, setID, "typ"); len(gelesen) != 3 ||
		gelesen[0] != "Artefakt" || gelesen[2] != "Artefakt" {
		t.Fatalf("nach der Änderung %v", gelesen)
	}

	leser := storage.NeuerVerlaufsleser(db)

	// **Drei Zeilen, ein Schritt.** Jede Karte trägt ihren eigenen Weg zurück —
	// die drei standen vorher auf nichts, und „gab es nicht" ist nicht „war
	// leer". Zusammengehalten werden sie von der Gruppe.
	verlauf, err := leser.Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 3 {
		t.Fatalf("%d Einträge, erwartet drei", len(verlauf))
	}

	gruppe := verlauf[0].Gruppe
	if gruppe == "" {
		t.Fatal("der erste Eintrag trägt keine Gruppe")
	}

	for _, eintrag := range verlauf {
		if eintrag.Gruppe != gruppe {
			t.Errorf("%+v gehört zu einer anderen Gruppe", eintrag)
		}
	}

	// **Ein ⌘Z nimmt alle drei zurück.**
	bewegt, gab, err := leser.Zuruecknehmen(probekontext(t), setID)
	if err != nil || !gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	if bewegt.Umfang != 3 {
		t.Errorf("Umfang %d, erwartet drei", bewegt.Umfang)
	}

	for i, wert := range werte(t, db, setID, "typ") {
		if wert != "«fehlt»" {
			t.Errorf("Karte %d steht auf %q — das Feld gab es vorher nicht", i, wert)
		}
	}

	// **Und ein ⇧⌘Z bringt alle drei zurück.**
	if _, gab, err := leser.Wiederherstellen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("wiederherstellen: %v / %v", gab, err)
	}

	for i, wert := range werte(t, db, setID, "typ") {
		if wert != "Artefakt" {
			t.Errorf("Karte %d steht auf %q", i, wert)
		}
	}
}

// **Eine Einzeländerung nach einer Massenänderung kommt zuerst zurück.**
//
// Die Reihenfolge ist eine Kette, keine Menge: wer die Gruppe zuerst nähme,
// stellte einen Stand her, den es nie gab.
func TestDieKetteHaeltUeberSchritteHinweg(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	schreiber := storage.NeuerKartenschreiber(db)

	if _, err := schreiber.FelderSetzen(probekontext(t), karten, "typ", "Artefakt"); err != nil {
		t.Fatalf("Massenänderung: %v", err)
	}

	if _, _, err := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karten[0], "name", "Meteor"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	// Erst die Einzeländerung …
	bewegt, _, err := leser.Zuruecknehmen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if bewegt.Umfang != 1 || bewegt.Feld != "name" {
		t.Errorf("%+v, erwartet war die Einzeländerung", bewegt)
	}

	if gelesen := werte(t, db, setID, "name"); gelesen[0] != "Feuerball" {
		t.Errorf("Name %q", gelesen[0])
	}

	if gelesen := werte(t, db, setID, "typ"); gelesen[0] != "Artefakt" {
		t.Errorf("die Gruppe wurde mit zurückgenommen: %v", gelesen)
	}

	// … dann die Gruppe.
	bewegt, _, err = leser.Zuruecknehmen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if bewegt.Umfang != 3 {
		t.Errorf("Umfang %d, erwartet drei", bewegt.Umfang)
	}
}

// **Die Grenze zählt Schritte, nicht Zeilen.**
//
// Sonst zerschnitte sie eine Massenänderung: die Hälfte der Karten käme zurück,
// die andere bliebe stehen — ein Stand, den es nie gab. Und eine einzige
// Massenänderung über dreihundert Karten löschte den ganzen übrigen Verlauf.
func TestDieGrenzeZaehltSchritte(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	schreiber := storage.NeuerKartenschreiber(db)

	for lauf := range storage.GrenzeVerlauf + 5 {
		if _, err := schreiber.FelderSetzen(
			probekontext(t), karten, "typ", string(rune('a'+lauf%26))); err != nil {
			t.Fatalf("Massenänderung: %v", err)
		}
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	// Zweihundert Schritte zu je drei Zeilen.
	if len(verlauf) != storage.GrenzeVerlauf*3 {
		t.Fatalf("%d Zeilen, erwartet %d", len(verlauf), storage.GrenzeVerlauf*3)
	}

	gruppen := map[string]int{}
	for _, eintrag := range verlauf {
		gruppen[eintrag.Gruppe]++
	}

	if len(gruppen) != storage.GrenzeVerlauf {
		t.Errorf("%d Gruppen, erwartet %d", len(gruppen), storage.GrenzeVerlauf)
	}

	for gruppe, anzahl := range gruppen {
		if anzahl != 3 {
			t.Errorf("Gruppe %s trägt %d Zeilen — sie wurde zerschnitten", gruppe, anzahl)
		}
	}
}
