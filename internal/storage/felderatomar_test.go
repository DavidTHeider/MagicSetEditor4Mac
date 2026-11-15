package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Mehrere Felder einer Karte gehen zusammen hinaus oder gar nicht** —
// #714.
//
// # Der Befund
//
// Aus dem Zweitblick durch Codex vom 2026-08-21: `Feldweg.teileSchreiben`
// schrieb Teil für Teil, und **jeder Aufruf eröffnete eine eigene
// Transaktion**. Scheiterte Teil zwei, war Teil eins schon bestätigt — eine
// halb geänderte Karte, und niemand konnte sagen, welche Hälfte.
//
// Bei `text` einer M15-Karte heißt das: der Regeltext steht neu da, der
// Flavortext alt, und die gezeichnete Karte ist eine, die niemand gemeint hat.
func TestMehrereFelderGehenZusammenHinaus(t *testing.T) {
	db := frischeDatenbank(t)

	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Feuerball"})
	setID := setVon(t, db, karteID)

	schreiber := storage.NeuerKartenschreiber(db)

	// **Ein Feldname, den die Domäne abweist**, mitten in der Reihe: der
	// billigste Weg, einen Fehler **nach** dem ersten Schreiben zu erzeugen.
	_, _, err := schreiber.FelderSetzenMitVorwert(probekontext(t), karteID, []domain.Value{
		{Field: "rule text", Text: "Neu"},
		{Field: "  ", Text: "kaputt"},
	})
	if err == nil {
		t.Fatal("der kaputte Feldname ging durch")
	}

	// **Und der erste Teil ist nicht da.** Genau das ist die Zusage.
	werte, fehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	for _, wert := range werte[0].Values {
		if wert.Field == "rule text" {
			t.Errorf("der erste Teil steht im Bestand: %q", wert.Text)
		}
	}
}

// **Und wenn alles geht, geht alles.**
func TestMehrereFelderKommenAlleAn(t *testing.T) {
	db := frischeDatenbank(t)

	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Feuerball"})
	setID := setVon(t, db, karteID)

	schreiber := storage.NeuerKartenschreiber(db)

	vorher, hatte, err := schreiber.FelderSetzenMitVorwert(probekontext(t), karteID,
		[]domain.Value{
			{Field: "rule text", Text: "Regel"},
			{Field: "flavor text", Text: "Flavor"},
		})
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Der Vorwert ist der des ersten Feldes** — er geht an die Rücknahme des
	// Aufrufers; die übrigen stehen im Verlauf.
	if vorher != "" || hatte {
		t.Errorf("Vorwert %q, hatte %v", vorher, hatte)
	}

	karten, fehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	gefunden := map[string]string{}
	for _, wert := range karten[0].Values {
		gefunden[wert.Field] = wert.Text
	}

	if gefunden["rule text"] != "Regel" || gefunden["flavor text"] != "Flavor" {
		t.Errorf("Werte = %v", gefunden)
	}

	// **Ein Aufruf, ein Schritt** (#2396): seit der Schreibweg nachzieht, wer
	// ein Feld liest, schreibt schon eine einzige Bearbeitung mehrere Felder —
	// und MSE nimmt sie als **eine** Aktion zurück
	// (`SetScriptManager::onAction`, `src/script/script_manager.cpp`,
	// Zeilen 141–147, Fassung f9d9356 vom 2021-07-06). Je Feld ein Schritt
	// hieße ⌘Z je nachgezogenem Feld, und dazwischen ein halber Stand.
	schritt, da, fehler := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if fehler != nil || !da {
		t.Fatalf("zurücknehmen: %v, da %v", fehler, da)
	}

	// **Der Schritt nennt das erste Feld und zählt Karten** — eine.
	if schritt.Feld != "rule text" || schritt.Umfang != 1 {
		t.Errorf("zurückgenommen: Feld %q, Umfang %d — erwartet rule text über eine Karte",
			schritt.Feld, schritt.Umfang)
	}

	karten, fehler = storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	for _, wert := range karten[0].Values {
		if wert.Field == "rule text" || wert.Field == "flavor text" {
			t.Errorf("nach einem ⌘Z steht noch %s=%q", wert.Field, wert.Text)
		}
	}
}

// **Eine leere Reihe ändert nichts und meldet nichts.**
func TestKeinFeldIstKeinFehler(t *testing.T) {
	db := frischeDatenbank(t)
	karteID := eineKarte(t, db, domain.Value{Field: "name", Text: "Feuerball"})

	if _, _, err := storage.NeuerKartenschreiber(db).FelderSetzenMitVorwert(
		probekontext(t), karteID, nil,
	); err != nil {
		t.Errorf("die leere Reihe meldete %v", err)
	}
}
