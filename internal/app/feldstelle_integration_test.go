package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Entfernen und zurücknehmen lässt die Felder stehen, wo sie standen** —
// durch den ganzen Stapel, von der Paketdatei bis zurück in den Bestand.
//
// # Warum dieser Test hier steht und nicht in `storage`
//
// Er hält die **Naht** fest: `card_values.reihenfolge` entscheidet, in welcher
// Reihenfolge der Bestand die Felder einer Karte hergibt, und diese Reihenfolge
// ist es, die Editor, Vorschau und MCP-Server zu sehen bekommen. Ein Test in
// `storage` prüft die Spalte; dieser prüft, was am anderen Ende ankommt.
//
// # Der Byte-Vergleich steht daneben
//
// Bis #191
// wäre er grün gewesen, ohne seinen Fall auszuführen: der Schreiber entfernte
// keine Zeilen, also war die Reihenfolge für die Datei folgenlos. Seit #191
// entfernt und ergänzt er sie, und damit hat der Vergleich etwas zu prüfen —
// er steht in `TestEineZurueckgenommeneEntfernungLaesstDieDateiWieSieWar`.
//
// # Warum `farbe` und nicht `name`
//
// `farbe` steht in jeder Karte des Prüfstücks an **zweiter** Stelle. Ein Feld,
// das ohnehin am Ende steht, käme auch mit `max + 1` an die richtige Stelle
// zurück.
func TestEineZurueckgenommeneEntfernungLaesstDieFelderStehen(t *testing.T) {
	db := frischeAppdatenbank(t)

	if _, err := service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		Set(probekontext(t), probeset); err != nil {
		t.Fatalf("Import: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	koepfe, err := leser.Setkoepfe(probekontext(t))
	if err != nil || len(koepfe) != 1 {
		t.Fatalf("Setköpfe: %v (%d)", err, len(koepfe))
	}

	setID := koepfe[0].ID

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten: %v", err)
	}

	schreiber := storage.NeuerKartenschreiber(db)

	// **Ein drittes Feld, und das ist keine Zutat, sondern die Bedingung.**
	//
	// Die Karten des Prüfstücks führen `name` und `farbe`. Nimmt man dort das
	// zweite fort, ist `max + 1` **dieselbe** Stelle, die es hatte — der Test
	// wäre grün, ohne seinen Fall auszuführen, und die Gegenprobe hat ihn genau
	// so erwischt. Erst mit einem Feld dahinter wandert es sichtbar ans Ende.
	if _, _, err := schreiber.
		FeldSetzenMitVorwert(probekontext(t), karten[0].ID, "text", "Drei Schaden"); err != nil {
		t.Fatalf("ergänzen: %v", err)
	}

	vorher := feldnamen(t, leser, setID)
	if len(vorher) != 3 || vorher[1] != "farbe" {
		t.Fatalf("Ausgangslage %v", vorher)
	}

	if _, _, _, err := schreiber.
		FeldEntfernen(probekontext(t), karten[0].ID, "farbe"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if _, gab, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v (gab=%v)", err, gab)
	}

	nachher := feldnamen(t, leser, setID)
	if len(nachher) != len(vorher) {
		t.Fatalf("%d Felder, erwartet %d", len(nachher), len(vorher))
	}

	for i := range vorher {
		if nachher[i] != vorher[i] {
			t.Errorf("Reihenfolge %v, erwartet %v", nachher, vorher)

			break
		}
	}
}

// feldnamen gibt die Feldnamen der ersten Karte in ihrer Reihenfolge.
func feldnamen(t *testing.T, leser *storage.SetLeser, setID int64) []string {
	t.Helper()

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil || len(karten) == 0 {
		t.Fatalf("Karten: %v", err)
	}

	namen := make([]string, 0, len(karten[0].Values))
	for _, wert := range karten[0].Values {
		namen = append(namen, wert.Field)
	}

	return namen
}

// **Entfernen, zurücknehmen, schreiben — und die Datei ist wieder dieselbe.**
//
// Die Zusage aus #160,
// nachweisbar seit #191:
// vorher entfernte der Schreiber keine Zeilen, und der Vergleich wäre grün
// gewesen, ohne seinen Fall auszuführen.
//
// **Beide Hälften müssen stimmen**, und das ist der Grund, warum dieser Test
// hier steht und nicht in einer der beiden Schichten: der Schreiber muss die
// Zeile entfernen **und** an der richtigen Stelle wieder einsetzen. Fiele eine
// von beiden aus, wäre die Datei eine andere.
func TestEineZurueckgenommeneEntfernungLaesstDieDateiWieSieWar(t *testing.T) {
	db := frischeAppdatenbank(t)

	if _, err := service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		Set(probekontext(t), probeset); err != nil {
		t.Fatalf("Import: %v", err)
	}

	urschrift, err := os.ReadFile(filepath.Join(probeset, "set"))
	if err != nil {
		t.Fatalf("Urschrift lesen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	koepfe, err := leser.Setkoepfe(probekontext(t))
	if err != nil || len(koepfe) != 1 {
		t.Fatalf("Setköpfe: %v", err)
	}

	setID := koepfe[0].ID

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten: %v", err)
	}

	// **Das mittlere Feld einer Karte mit dreien** — sonst wäre `max + 1`
	// dieselbe Stelle, und der Test bliebe grün, ohne seinen Fall auszuführen.
	schreiber := storage.NeuerKartenschreiber(db)
	if _, _, _, err := schreiber.
		FeldEntfernen(probekontext(t), karten[1].ID, "name"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	// **Dazwischen wird geschrieben, und das ist der Punkt.**
	//
	// Ohne diesen Schritt führte der Test die Entfernung nie aus: nach der
	// Rücknahme ist im Bestand alles wieder da, und der Schreiber hätte nichts
	// wegzunehmen. Die Gegenprobe hat es erwischt — der Lauf blieb grün, als
	// `zeilenEntfernen` versuchsweise ausgebaut wurde.
	ohneFeld := geschriebenesSetVon(t, leser, setID)
	if strings.Contains(ohneFeld, "name: Eisball") {
		t.Errorf("die entfernte Zeile steht noch da:\n%s", ohneFeld)
	}

	if _, gab, err := storage.NeuerVerlaufsleser(db).
		Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v (gab=%v)", err, gab)
	}

	if wieder := geschriebenesSetVon(t, leser, setID); wieder != string(urschrift) {
		t.Errorf("die Datei hat sich verändert:\n%q\nerwartet\n%q", wieder, urschrift)
	}
}

// geschriebenesSetVon schreibt ein Set an einen frischen Ort und liest es.
func geschriebenesSetVon(t *testing.T, leser *storage.SetLeser, setID int64) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")
	if _, err := service.NeuerSetschreiber(leser, leser).Schreibe(probekontext(t), setID, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return string(gelesen)
}
