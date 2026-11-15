// **Die Sammlung im Bestand** — Werkstück 2 von Phase 3.4
// (#1165).
//
// **Kein `t.Parallel()` hier**, und das ist keine Nachlässigkeit: die Datenbank
// im Arbeitsspeicher trägt `cache=shared` und ist damit **eine** je Prozess.
// Der Grund steht ausführlich bei `frischeDatenbank`; wer es doch tut, bekommt
// „die Migration schlug fehl: database schema is locked" — hier gesehen, nicht
// vermutet.
package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probedruck ist die Scryfall-Kennung, die diese Läufe benutzen.
const probedruck = "8a1b0d5e-1e0f-4c0f-8b39-3d3a2f9a1c2d"

// probeexemplar gibt ein Exemplar, das jeder Fall abwandelt.
func probeexemplar() *domain.Exemplar {
	return &domain.Exemplar{
		Printing: probedruck,
		Menge:    3,
		Zustand:  domain.ZustandNearMint,
		Sprache:  "en",
	}
}

// **Was abgelegt wurde, kommt zurück — Feld für Feld.**
func TestEinExemplarKommtZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	exemplar := probeexemplar()
	exemplar.Finish = domain.FinishFoil
	exemplar.Erwerbsdatum = "2026-08-01"
	exemplar.Erwerbspreis = 417
	exemplar.Erwerbswaehrung = "EUR"
	exemplar.Notiz = "aus dem Laden an der Ecke"

	id, err := speicher.Anlegen(probekontext(t), exemplar)
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesen, lesefehler := speicher.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if len(gelesen) != 1 {
		t.Fatalf("es sind %d Exemplare", len(gelesen))
	}

	exemplar.ID = id

	// **Die Fassung kommt vom Bestand.** Ein frisch angelegtes Exemplar hat
	// die erste; wer sie hier selbst setzte, prüfte seine eigene Annahme.
	exemplar.Fassung = 1

	// **Und die Zeile des Drucks ebenso** (A2, #1441): der Alias hat sie
	// angelegt; welche Nummer sie trägt, entscheidet der Bestand.
	if gelesen[0].DruckID == 0 {
		t.Error("das Exemplar hat keine Druckzeile bekommen")
	}

	exemplar.DruckID = gelesen[0].DruckID

	if *gelesen[0] != *exemplar {
		t.Errorf("zurück kam %+v statt %+v", *gelesen[0], *exemplar)
	}
}

// **Dieselbe Karte in demselben Zustand ist eine Zeile, keine zweite.**
//
// # Warum das ein Konflikt ist und keine Addition
//
// Wer zweimal anlegt, hat entweder zweimal geklickt oder meinte eine
// **Änderung**. Beides still zu addieren machte aus einem Versehen einen
// Bestand, den niemand nachvollziehen kann — und §7, Regel 4 sagt, dass ein
// zweiter Aufruf gegen einen Stand geht, den es so nicht gibt.
func TestZweimalDasselbeExemplarIstEinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	if _, err := speicher.Anlegen(probekontext(t), probeexemplar()); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	_, err := speicher.Anlegen(probekontext(t), probeexemplar())

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Derselbe Druck in einem anderen Zustand ist eine zweite Zeile.**
func TestDerselbeDruckInAnderemZustandIstEineZweiteZeile(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	for _, zustand := range []domain.Zustand{
		domain.ZustandNearMint, domain.ZustandPlayed,
	} {
		exemplar := probeexemplar()
		exemplar.Zustand = zustand

		if _, err := speicher.Anlegen(probekontext(t), exemplar); err != nil {
			t.Fatalf("Anlegen (%s): %v", zustand, err)
		}
	}

	gelesen, err := speicher.Fuer(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Fuer: %v", err)
	}

	if len(gelesen) != 2 {
		t.Errorf("es sind %d Zeilen", len(gelesen))
	}
}

// **Eine Änderung trifft die Zeile, die gemeint ist.**
func TestEinExemplarLaesstSichAendern(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	geaendert := probeexemplar()
	geaendert.ID = id
	geaendert.Fassung = 1
	geaendert.Menge = 7
	geaendert.Notiz = "zwei dazugekauft"

	if fehler := speicher.Aendern(probekontext(t), geaendert); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	gelesen, lesefehler := speicher.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if gelesen[0].Menge != 7 || gelesen[0].Notiz != "zwei dazugekauft" {
		t.Errorf("zurück kam %+v", *gelesen[0])
	}
}

// **Eine Änderung an einer Zeile, die es nicht gibt, ist ein Fehlschlag.**
//
// Sonst sähe ein Aufruf gegen eine gelöschte Zeile wie ein Erfolg aus, und die
// Oberfläche zeigte einen Stand, den der Bestand nicht führt.
func TestEineAenderungOhneZeileScheitert(t *testing.T) {
	db := frischeDatenbank(t)

	geaendert := probeexemplar()
	geaendert.ID = 4711
	geaendert.Fassung = 1

	err := storage.NeuerSammlungsspeicher(db).Aendern(probekontext(t), geaendert)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Gelöscht ist gelöscht.**
func TestEinExemplarLaesstSichLoeschen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesenVorher, lesefehlerVorher := speicher.Fuer(probekontext(t), probeexemplar().Printing)
	if lesefehlerVorher != nil || len(gelesenVorher) != 1 {
		t.Fatalf("Fuer: %v, %d Zeilen", lesefehlerVorher, len(gelesenVorher))
	}

	if fehler := speicher.Loeschen(probekontext(t), id,
		gelesenVorher[0].Fassung); fehler != nil {
		t.Fatalf("Loeschen: %v", fehler)
	}

	gelesen, lesefehler := speicher.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if len(gelesen) != 0 {
		t.Errorf("es sind noch %d Exemplare da", len(gelesen))
	}
}

// **Und ein Löschen ins Leere sagt es.**
func TestEinLoeschenOhneZeileScheitert(t *testing.T) {
	err := storage.NeuerSammlungsspeicher(frischeDatenbank(t)).
		Loeschen(probekontext(t), 4711, 1)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Eine leere Sammlung gibt eine leere Liste und kein `nil`.**
//
// Sonst käme über die Bindungen ein `null` heraus, wo die Oberfläche eine
// Liste erwartet — genau der Fall aus §3, den keine der beiden Prüfebenen
// sieht.
func TestEineLeereSammlungGibtEineLeereListe(t *testing.T) {
	gelesen, err := storage.NeuerSammlungsspeicher(frischeDatenbank(t)).
		Liste(probekontext(t))
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if gelesen == nil {
		t.Error("die Liste ist nil")
	}
}

// **Die Liste steht in einer Reihenfolge, die zweimal dieselbe ist.**
//
// Ohne `ORDER BY` gibt SQLite zurück, was ihm gerade passt; eine Sammlung, die
// bei jedem Aufschlagen anders sortiert ist, lässt niemanden etwas
// wiederfinden.
func TestDieSammlungHatEineFesteReihenfolge(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	for _, druck := range []string{"zzz-druck", "aaa-druck", "mmm-druck"} {
		exemplar := probeexemplar()
		exemplar.Printing = druck

		if _, err := speicher.Anlegen(probekontext(t), exemplar); err != nil {
			t.Fatalf("Anlegen: %v", err)
		}
	}

	gelesen, err := speicher.Liste(probekontext(t))
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	for i, erwartet := range []string{"aaa-druck", "mmm-druck", "zzz-druck"} {
		if gelesen[i].Printing != erwartet {
			t.Errorf("an Stelle %d steht %q statt %q",
				i, gelesen[i].Printing, erwartet)
		}
	}
}

// **Eine veraltete Fassung ist ein Konflikt und kein Überschreiben.**
//
// # Warum das der Kern dieser Zusage ist
//
// Ohne sie ersetzte jedes Schreiben alle Spalten allein nach `id`: zwei
// Fenster bearbeiten dieselbe Zeile, und wer später speichert, überschreibt
// die frühere Änderung — unbemerkt. Angestrichen vom Zweitblick über
// Phase 3.4.
func TestEineVeralteteFassungIstEinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	// **Der erste schreibt.**
	erster := probeexemplar()
	erster.ID, erster.Fassung, erster.Menge = id, 1, 7

	if fehler := speicher.Aendern(probekontext(t), erster); fehler != nil {
		t.Fatalf("der erste Schreibvorgang: %v", fehler)
	}

	// **Der zweite hat noch die alte Fassung gelesen.**
	zweiter := probeexemplar()
	zweiter.ID, zweiter.Fassung, zweiter.Menge = id, 1, 99

	if fehler := speicher.Aendern(probekontext(t), zweiter); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", fehler)
	}

	// **Und die frühere Änderung steht noch da.**
	gelesen, lesefehler := speicher.Liste(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("Liste: %v", lesefehler)
	}

	if gelesen[0].Menge != 7 {
		t.Errorf("die Menge ist %d — die frühere Änderung wurde überschrieben",
			gelesen[0].Menge)
	}

	// **Und die Fassung ist gestiegen.** Ohne das griffe der zweite Versuch
	// beim nächsten Mal wieder daneben.
	if gelesen[0].Fassung != 2 {
		t.Errorf("die Fassung ist %d", gelesen[0].Fassung)
	}
}

// **Und eine Zeile, die es nicht gibt, sagt das und nicht „Konflikt".**
//
// Der Unterschied ist der halbe Zweck: „gibt es nicht" heißt aufräumen,
// „jemand anderes hat geschrieben" heißt nachladen (§8).
func TestEineFehlendeZeileIstKeinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)

	geaendert := probeexemplar()
	geaendert.ID, geaendert.Fassung = 4711, 1

	err := storage.NeuerSammlungsspeicher(db).Aendern(probekontext(t), geaendert)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("der Fehler ist %v — will ErrNotFound", err)
	}
}

// **Ein Löschen gegen einen veralteten Stand wird abgewiesen.**
//
// # Warum Löschen dieselbe Bedingung braucht wie Ändern
//
// Seit #1196 schützt `Aendern` gegen eine fremde Änderung; `Loeschen` nahm nur
// die Kennung. Damit blieb die **Klasse** offen, die die Korrektur schließen
// sollte: ein schreibender Zug gegen einen veralteten Stand entfernt fremde
// Arbeit.
//
// Der Fall dazu ist alltäglich. Jemand liest „ein Stück", ein zweites Fenster
// trägt drei nach, und das erste löscht seinen alten Stand — weg sind vier.
// Angestrichen vom Nachlauf über Phase 3.4 als Befund 10.
func TestEinLoeschenGegenEinenAltenStandWirdAbgewiesen(t *testing.T) {
	speicher := storage.NeuerSammlungsspeicher(frischeDatenbank(t))

	id, anlegefehler := speicher.Anlegen(probekontext(t), probeexemplar())
	if anlegefehler != nil {
		t.Fatalf("Anlegen: %v", anlegefehler)
	}

	gelesen, err := speicher.Fuer(probekontext(t), probeexemplar().Printing)
	if err != nil || len(gelesen) != 1 {
		t.Fatalf("Fuer: %v, %d Zeilen", err, len(gelesen))
	}

	alt := gelesen[0].Fassung

	// **Jemand anderes schreibt.** Danach ist die gelesene Fassung überholt.
	geaendert := *gelesen[0]
	geaendert.Menge = 4

	if fehler := speicher.Aendern(probekontext(t), &geaendert); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if fehler := speicher.Loeschen(probekontext(t), id, alt); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", fehler)
	}

	// **Und die Zeile steht noch da.** Ein Konflikt, der trotzdem löscht, wäre
	// die schlimmere Hälfte.
	danach, lesefehler := speicher.Fuer(probekontext(t), probeexemplar().Printing)
	if lesefehler != nil || len(danach) != 1 {
		t.Fatalf("nach dem Konflikt: %v, %d Zeilen", lesefehler, len(danach))
	}

	if danach[0].Menge != 4 {
		t.Errorf("die fremde Änderung ist weg: Menge %d", danach[0].Menge)
	}
}
