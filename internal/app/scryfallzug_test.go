// **Bild, Herkunft und Künstler sind ein Zug** — Befund Z7-02 des
// Audits für 0.3.0 (Entwicklungsarchiv), Phase 3.2
// (#1163).
//
// # Der Befund
//
// `ScryfallBildEinsetzen` schrieb **zweimal**: erst Bildfeld und Herkunft,
// dann `illustrator`. Der Kommentar an der Stelle sagte es selbst — „das Bild
// liegt schon. Ein Fehler hier wird gemeldet, macht es aber nicht rückgängig."
//
// Dann trägt die Karte fremde Kunst **ohne ihre Nennung**, und genau das
// schließt die Bedingung von Scryfall aus, für die es diesen ganzen Weg gibt.
//
// **§7, Regel 3 galt hier nicht.** „Geschrieben ist geschrieben" spricht vom
// **Nachladen** nach einer geglückten Änderung — nicht davon, eine Änderung in
// zwei zu zerlegen, die einzeln falsch sind.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// **Beide Felder gehen in einem Zug hinaus.**
func TestScryfallBildSchreibtBeideFelderInEinemZug(t *testing.T) {
	holer := &bildholer{holung: einBildfund(t, "Christopher Rush")}
	anwendung, schreiber, _ := scryfallbildaufbau(t, holer)

	vorbereitet := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Lightning Bolt", "normal")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	if antwort := anwendung.ScryfallBildEinsetzen(t.Context(), "1", "1", "kunst",
		vorbereitet.Pfad, "Christopher Rush", vorbereitet.Adresse,
		vorbereitet.Auswahl); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(schreiber.zuege) != 1 {
		t.Fatalf("es waren %d Züge — zwischen zweien liegt ein Stand, in dem "+
			"das Bild an der Karte hängt und der Name nicht: %v",
			len(schreiber.zuege), schreiber.zuege)
	}

	if len(schreiber.zuege[0]) != 2 {
		t.Errorf("der Zug trug %d Werte: %v", len(schreiber.zuege[0]), schreiber.zuege[0])
	}
}

// **Scheitert der Zug, bleibt kein halber Stand — auch nicht auf der Platte.**
//
// Das Bild wird **vor** dem Schreiben abgelegt; es hat zu dem Zeitpunkt noch
// keinen Verweis. Bliebe es liegen, wüchse das Bilderverzeichnis bei jedem
// Fehlversuch um eine Datei, die niemand je wieder findet.
func TestEinGescheiterterZugLaesstKeinBildLiegen(t *testing.T) {
	holer := &bildholer{holung: einBildfund(t, "Christopher Rush")}
	anwendung, schreiber, wurzel := scryfallbildaufbau(t, holer)

	schreiber.fehler = errors.New("der Bestand nimmt nichts an")

	vorbereitet := anwendung.ScryfallBild(t.Context(), "1", "kunst", "Lightning Bolt", "normal")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	if antwort := anwendung.ScryfallBildEinsetzen(t.Context(), "1", "1", "kunst",
		vorbereitet.Pfad, "Christopher Rush", vorbereitet.Adresse,
		vorbereitet.Auswahl); antwort.Fehler == nil {
		t.Fatal("der gescheiterte Zug meldete Erfolg")
	}

	eintraege, err := os.ReadDir(filepath.Join(wurzel, "7"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 0 {
		t.Errorf("es blieb ein Bild ohne Verweis liegen: %v", eintraege)
	}
}
