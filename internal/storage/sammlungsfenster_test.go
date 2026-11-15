// **Ordnen und Zeigen sind zwei Abfragen** — Werkstück
// #1500.
//
// **Kein `t.Parallel()`**, aus demselben Grund wie in `sammlung_test.go`: die
// Datenbank im Arbeitsspeicher ist eine je Prozess.
package storage_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Ordnungsliste trägt, was eine Sicht braucht — und sonst nichts.**
//
// Das ist die Zusage, für die es sie gibt: wer sie zeigte, zeigte eine Karte
// ohne Zustand, ohne Preis und ohne Herkunft. Der Lauf hält beide Hälften —
// die neun Felder stehen da, die anderen sind leer.
func TestDieOrdnungslisteTraegtNurWasEineSichtBraucht(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	exemplar := probeexemplar()
	exemplar.Name = "Ätherfluss"
	exemplar.Setcode = "m11"
	exemplar.Nummer = "42"
	exemplar.Notiz = "getauscht"
	exemplar.Finish = domain.FinishFoil
	exemplar.Erwerbsdatum = "2026-08-01"
	exemplar.Erwerbspreis = 417
	exemplar.Erwerbswaehrung = "EUR"
	exemplar.Lagerort = "Kiste 4"

	if _, err := speicher.Anlegen(probekontext(t), exemplar); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	zeilen, fehler := speicher.Ordnungsliste(probekontext(t))
	if fehler != nil {
		t.Fatalf("Ordnungsliste: %v", fehler)
	}

	if len(zeilen) != 1 {
		t.Fatalf("es sind %d Zeilen", len(zeilen))
	}

	zeile := zeilen[0]
	if zeile.Name != "Ätherfluss" || zeile.Setcode != "m11" || zeile.Nummer != "42" ||
		zeile.Printing != probedruck || zeile.Notiz != "getauscht" ||
		zeile.Menge != 3 || zeile.Erwerbsdatum != "2026-08-01" ||
		zeile.Ausfuehrung() != domain.FinishFoil {
		t.Errorf("die Zeile ist %+v", zeile)
	}

	// **Und die anderen bleiben leer** — sonst wäre die schmale Abfrage keine.
	if zeile.Zustand != "" || zeile.Sprache != "" || zeile.Erwerbspreis != 0 ||
		zeile.Erwerbswaehrung != "" || zeile.Lagerort != "" || zeile.Fassung != 0 {
		t.Errorf("die Zeile trägt mehr, als sie soll: %+v", zeile)
	}
}

// **Die Reihenfolge kommt vom Rufenden.**
func TestNachKennungenHaeltDieUebergebeneReihenfolge(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	kennungen := make([]int64, 0, 3)

	for zaehler, name := range []string{"A", "B", "C"} {
		// **Ein eigener Druck je Zeile**: dieselbe Karte in demselben Zustand,
		// derselben Sprache und derselben Ausführung gibt es **einmal**.
		exemplar := probeexemplar()
		exemplar.Printing = probedruck[:len(probedruck)-1] + strconv.Itoa(zaehler)
		exemplar.Name = name
		exemplar.Notiz = name

		id, err := speicher.Anlegen(probekontext(t), exemplar)
		if err != nil {
			t.Fatalf("Anlegen %s: %v", name, err)
		}

		kennungen = append(kennungen, id)
	}

	// Rückwärts, und eine, die es nicht gibt: sie fehlt still.
	gewuenscht := []int64{kennungen[2], kennungen[0], 999999}

	gelesen, fehler := speicher.NachKennungen(probekontext(t), gewuenscht)
	if fehler != nil {
		t.Fatalf("NachKennungen: %v", fehler)
	}

	if len(gelesen) != 2 || gelesen[0].Name != "C" || gelesen[1].Name != "A" {
		t.Fatalf("gelesen wurde %d Zeilen in der Reihenfolge %v", len(gelesen), namen(gelesen))
	}

	// **Und sie sind vollständig** — der Unterschied zur Ordnungsliste.
	if gelesen[0].Zustand != domain.ZustandNearMint {
		t.Errorf("der Zustand fehlt: %+v", gelesen[0])
	}
}

// **Mehr als ein Stapel geht auch** — die Abfrage hat nicht beliebig viele
// Platzhalter, und die Grenze gehört diesem Programm und nicht dem Treiber.
func TestNachKennungenTraegtMehrAlsEinenStapel(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	const wieViele = 620

	kennungen := make([]int64, 0, wieViele)

	for zaehler := range wieViele {
		exemplar := probeexemplar()
		exemplar.Printing = fmt.Sprintf("00000000-0000-4000-8000-%012d", zaehler)
		exemplar.Notiz = strconv.Itoa(zaehler)

		id, err := speicher.Anlegen(probekontext(t), exemplar)
		if err != nil {
			t.Fatalf("Anlegen %d: %v", zaehler, err)
		}

		kennungen = append(kennungen, id)
	}

	gelesen, fehler := speicher.NachKennungen(probekontext(t), kennungen)
	if fehler != nil {
		t.Fatalf("NachKennungen: %v", fehler)
	}

	if len(gelesen) != wieViele {
		t.Fatalf("gelesen wurden %d von %d", len(gelesen), wieViele)
	}

	for stelle, eines := range gelesen {
		if eines.ID != kennungen[stelle] {
			t.Fatalf("an Stelle %d steht %d statt %d", stelle, eines.ID, kennungen[stelle])
		}
	}
}

// **Ohne Kennungen wird nicht gefragt** — und es kommt eine leere Liste, kein
// `nil` (§3).
func TestNachKennungenOhneKennungenGibtEineLeereListe(t *testing.T) {
	db := frischeDatenbank(t)

	gelesen, fehler := storage.NeuerSammlungsspeicher(db).
		NachKennungen(probekontext(t), nil)
	if fehler != nil {
		t.Fatalf("NachKennungen: %v", fehler)
	}

	if gelesen == nil || len(gelesen) != 0 {
		t.Errorf("gelesen wurde %v", gelesen)
	}
}

// **Ohne Profil geht keiner der beiden Wege.**
func TestDieFensterwegeBrauchenEinProfil(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if _, fehler := speicher.Ordnungsliste(t.Context()); fehler == nil {
		t.Error("die Ordnungsliste kam ohne Profil zurück")
	}

	if _, fehler := speicher.NachKennungen(t.Context(), []int64{1}); fehler == nil {
		t.Error("NachKennungen kam ohne Profil zurück")
	}
}

// **Ein geschlossener Bestand meldet sich.**
func TestDieOrdnungslisteMeldetEinenGeschlossenenBestand(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := speicher.Ordnungsliste(probekontext(t)); fehler == nil {
		t.Error("der geschlossene Bestand gab eine Ordnungsliste her")
	}

	if _, fehler := speicher.NachKennungen(probekontext(t), []int64{1}); fehler == nil {
		t.Error("der geschlossene Bestand gab Exemplare her")
	}
}

// namen zieht die Namen heraus — für eine Meldung, die man lesen kann.
func namen(exemplare []*domain.Exemplar) []string {
	gelesen := make([]string, 0, len(exemplare))
	for _, eines := range exemplare {
		gelesen = append(gelesen, eines.Name)
	}

	return gelesen
}
