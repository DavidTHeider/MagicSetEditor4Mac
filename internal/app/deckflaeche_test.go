// **Die Deckfläche für die drei Flächen ohne Fenster** — Werkstück 7 von
// Phase 4.6
// (#1293).
//
// # Warum sie hier geprüft wird und nicht nur in `internal/flaechen`
//
// **Weil jene Läufe eigene Prozesse starten.** Was dort läuft, zählt für die
// Abdeckung dieses Pakets nicht — und ein Durchreicher ohne Lauf ist ein
// Zweig, den kein Test betritt (§3). Die **fachliche** Zusage steht weiter
// dort: derselbe Ablauf über alle drei Flächen.
package app_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Der ganze Weg über die Fläche.**
func TestDerDeckwegUeberDieFlaeche(t *testing.T) {
	flaeche := deckflaeche(t)

	angelegt := flaeche.DeckAnlegen(probekontext(t), "Aggro", "magic", "Notiz")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	deckID := angelegt.Deck.ID

	if liste := flaeche.Decks(probekontext(t), "magic"); len(liste.Decks) != 1 {
		t.Errorf("%d Decks, erwartet eines", len(liste.Decks))
	}

	if gelesen := flaeche.Deck(probekontext(t), deckID); gelesen.Deck.Name != "Aggro" {
		t.Errorf("%+v", gelesen.Deck)
	}

	umbenannt := flaeche.DeckUmbenennen(probekontext(t), deckID, "1", "Midrange", "")
	if umbenannt.Fehler != nil {
		t.Fatalf("umbenennen: %+v", umbenannt.Fehler)
	}

	// **Kopie und Wunschliste** — DOM-03.AK-02 (#1698); die Kopie ist leer,
	// und einem leeren Deck fehlt nichts.
	if kopie := flaeche.DeckKopieren(probekontext(t), deckID, "Midrange II"); kopie.Fehler != nil ||
		kopie.Deck.Name != "Midrange II" {
		t.Errorf("kopieren: %+v", kopie)
	}

	// **Aus einem Set, das es nicht gibt, entsteht kein Deck** — FMT-04
	// (#1693); der Weg mit Karten steht in `internal/flaechen/deckausset_test.go`.
	if ausSet := flaeche.DeckAusSet(probekontext(t), "999", "", 1); ausSet.Fehler == nil ||
		ausSet.Fehler.Code != "nichtGefunden" {
		t.Errorf("aus einem fehlenden Set: %+v", ausSet)
	}

	if wunsch := flaeche.DeckFehlkartenWuenschen(probekontext(t), deckID); wunsch.Fehler != nil ||
		len(wunsch.Neu) != 0 {
		t.Errorf("wünschen: %+v", wunsch)
	}

	gelegt := flaeche.DeckplatzLegen(probekontext(t), deckID, "",
		"a1b2c3d4-0000-4000-8000-00000000000a", 4, "hauptdeck")
	if gelegt.Fehler != nil {
		t.Fatalf("legen: %+v", gelegt.Fehler)
	}

	format := flaeche.FormatAusBauart(
		probekontext(t), "konstruiert", "Konstruiert", "magic", "standard")
	if format.Fehler != nil {
		t.Fatalf("das Format: %+v", format.Fehler)
	}

	if formate := flaeche.Formate(probekontext(t), ""); len(formate.Formate) != 1 {
		t.Errorf("%d Formate, erwartet eines", len(formate.Formate))
	}

	if bauarten := flaeche.Bauarten(); len(bauarten.Bauarten) != 3 {
		t.Errorf("%d Bauarten, erwartet 3", len(bauarten.Bauarten))
	}

	gestellt := flaeche.DeckFormatWaehlen(probekontext(t), deckID, format.Format.ID)
	if gestellt.Fehler != nil {
		t.Fatalf("stellen: %+v", gestellt.Fehler)
	}

	// **Vier Karten in einem 60-Karten-Format** — der Größenbefund steht ohne
	// jede Gegenstelle.
	urteil := flaeche.DeckPruefen(probekontext(t), deckID)
	if urteil.Fehler != nil {
		t.Fatalf("prüfen: %+v", urteil.Fehler)
	}

	if urteil.Urteil.Zulaessig {
		t.Error("vier Karten galten als zulässig")
	}

	abgleich := flaeche.DeckAbgleichen(probekontext(t), deckID)
	if abgleich.Fehler != nil || abgleich.Fehlend != 4 {
		t.Errorf("%+v", abgleich)
	}

	geaendert := flaeche.FormatSetzen(probekontext(t), app.FormatDTO{
		ID: format.Format.ID, Fassung: "1", Name: "Konstruiert", Spiel: "magic",
		Mindestgroesse: 40,
	})
	if geaendert.Fehler != nil {
		t.Fatalf("das Format ändern: %+v", geaendert.Fehler)
	}

	// **Der Bereichswechsel über dieselbe Fläche** — Befund 7 des Zweitblicks
	// über Phase 5.1.
	verschoben := flaeche.DeckplatzVerschieben(
		probekontext(t), deckID, gelegt.Platz.ID, "sideboard")
	if verschoben.Fehler != nil {
		t.Fatalf("verschieben: %+v", verschoben.Fehler)
	}

	if verschoben.Platz.Bereich != "sideboard" {
		t.Errorf("Bereich %q, erwartet sideboard", verschoben.Platz.Bereich)
	}

	if genommen := flaeche.DeckplatzNehmen(
		probekontext(t), deckID, verschoben.Platz.ID); genommen.Fehler != nil {
		t.Errorf("nehmen: %+v", genommen.Fehler)
	}

	if geloescht := flaeche.DeckLoeschen(probekontext(t), deckID); geloescht.Fehler != nil {
		t.Errorf("löschen: %+v", geloescht.Fehler)
	}

	if weg := flaeche.FormatLoeschen(
		probekontext(t), format.Format.ID); weg.Fehler != nil {
		t.Errorf("das Format löschen: %+v", weg.Fehler)
	}
}

// **Und die Wege bauen sich über einem Bestand.**
//
// `DeckwegeUeber` ist der Verbund, den die drei Flächen bekommen — vier
// Nähte, und alle über dieselben Ablagen.
func TestDieDeckwegeBauenSichUeberEinemBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	wege, fehler := app.DeckwegeUeber(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("DeckwegeUeber: %v", fehler)
	}

	if wege.Bau == nil || wege.Formate == nil ||
		wege.Abgleich == nil || wege.Pruefung == nil {
		t.Errorf("%+v", wege)
	}
}

// deckflaeche baut die Fläche über einem frischen Bestand.
//
// **Mit eigenem Zuhause und umgelenkter Adresse** (§3): die Prüfung fragt
// Scryfall nach jedem Druck, und beides einzeln reicht nicht.
func deckflaeche(t *testing.T) *app.Deckflaeche {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	flaeche, fehler := app.DeckflaecheUeber(frischerBestand(t))
	if fehler != nil {
		t.Fatalf("DeckflaecheUeber: %v", fehler)
	}

	return flaeche
}

// stilleGegenstelle antwortet auf jede Anfrage mit 404.
//
// **Damit kein Aufruf hinausgeht** und die Prüfung trotzdem läuft: ein Druck,
// den Scryfall nicht kennt, ist `ungeprüft` — genau die Antwort, die dieser
// Lauf sehen will.
func stilleGegenstelle(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))

	t.Cleanup(server.Close)

	return server.URL
}

// **Ohne Scryfall-Zugang bleibt die Prüfung leer — und sagt es.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil die Fläche trotzdem entsteht.** `DeckflaecheUeber` gibt den Fehler
// mit hinaus und stellt alles andere bereit; wer das Zuhause nicht findet,
// soll seine Decks trotzdem sehen. Erst `DeckPruefen` sagt, dass es nicht
// geht — und zwar **an der Stelle**, an der es auffällt (§8).
//
// # Wie das Zuhause unbrauchbar wird
//
// `HOME` zeigt auf eine **Datei**. `app.Scryfallverzeichnis` legt darunter
// ein Verzeichnis an, und das scheitert — dieselbe Lage wie auf einem Rechner
// ohne Schreibrecht.
func TestOhneScryfallzugangBleibtDiePruefungLeer(t *testing.T) {
	zuhause := filepath.Join(t.TempDir(), "keinverzeichnis")

	if err := os.WriteFile(zuhause, []byte("x"), 0o600); err != nil {
		t.Fatalf("die Sperrdatei anlegen: %v", err)
	}

	t.Setenv("HOME", zuhause)

	flaeche, fehler := app.DeckflaecheUeber(frischerBestand(t))
	if fehler == nil {
		t.Fatal("der fehlende Zugang wurde nicht gemeldet")
	}

	// **Und die Fläche steht trotzdem.**
	if liste := flaeche.Decks(probekontext(t), ""); liste.Fehler != nil {
		t.Errorf("die Deckliste ging mit unter: %+v", liste.Fehler)
	}

	antwort := flaeche.DeckPruefen(probekontext(t), "1")
	if antwort.Fehler == nil {
		t.Error("die Prüfung schwieg über den fehlenden Zugang")
	}

	if antwort.Befunde == nil {
		t.Error("die Befunde sind null statt leer")
	}
}

// **Und die drei Flächen ohne Fenster nennen die Karten beim Namen** (#1376).
//
// Der Beleg ist der Weg selbst und nicht der Dienst darunter: `DeckwegeUeber`
// ist eine **zweite** Composition Root neben `deckstart.go` (§2), und beim
// ersten Anlauf war nur die erste verdrahtet. Der Dienst war grün, das Fenster
// zeigte Namen — und Kommandozeile, MCP und REST weiter UUIDs. Zwei grüne
// Seiten einer Naht belegen die Naht nicht.
func TestDieFlaechenOhneFensterNennenDieKartenBeimNamen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Scryfallumgebung, stilleGegenstelle(t))

	db := frischerBestand(t)
	karteID := eigeneKarteMit(t, db, "Lightning Bolt", "R")

	flaeche, fehler := app.DeckflaecheUeber(db)
	if fehler != nil {
		t.Fatalf("DeckflaecheUeber: %v", fehler)
	}

	angelegt := flaeche.DeckAnlegen(probekontext(t), "Burn", "magic", "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	gelegt := flaeche.DeckplatzLegen(probekontext(t), angelegt.Deck.ID,
		strconv.FormatInt(karteID, 10), "", 4, "hauptdeck")
	if gelegt.Fehler != nil {
		t.Fatalf("legen: %+v", gelegt.Fehler)
	}

	gelesen := flaeche.Deck(probekontext(t), angelegt.Deck.ID)
	if gelesen.Fehler != nil {
		t.Fatalf("lesen: %+v", gelesen.Fehler)
	}

	if len(gelesen.Deck.Plaetze) != 1 {
		t.Fatalf("%d Plätze, erwartet war einer", len(gelesen.Deck.Plaetze))
	}

	platz := gelesen.Deck.Plaetze[0]
	if platz.Bezeichnung != "Lightning Bolt" || platz.Manakosten != "R" {
		t.Errorf("der Platz heißt %q und kostet %q",
			platz.Bezeichnung, platz.Manakosten)
	}
}

// eigeneKarteMit legt eine Karte mit Namen und Manakosten in den Bestand.
func eigeneKarteMit(t *testing.T, db *sql.DB, name, kosten string) int64 {
	t.Helper()

	set, err := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, stil, geoeffnet, profil_id)
		 VALUES ('Probe', 'magic', 'm15', datetime('now'), 1)`)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	setID, kennungsfehler := set.LastInsertId()
	if kennungsfehler != nil {
		t.Fatalf("die Setkennung: %v", kennungsfehler)
	}

	karte, err := db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (?, 0)`, setID)
	if err != nil {
		t.Fatalf("die Karte anlegen: %v", err)
	}

	karteID, kennungsfehler := karte.LastInsertId()
	if kennungsfehler != nil {
		t.Fatalf("die Kartenkennung: %v", kennungsfehler)
	}

	for stelle, feld := range []struct{ name, wert string }{
		{"name", name}, {"casting cost", kosten},
	} {
		if _, err := db.ExecContext(probekontext(t),
			`INSERT INTO card_values (card_id, feld, wert, reihenfolge)
			 VALUES (?, ?, ?, ?)`, karteID, feld.name, feld.wert, stelle); err != nil {
			t.Fatalf("das Feld %q setzen: %v", feld.name, err)
		}
	}

	return karteID
}
