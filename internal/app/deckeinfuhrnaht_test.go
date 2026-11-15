// **Die Decklisteneinfuhrnaht** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Warum gegen echtes Storage und eine Gegenstelle
//
// **Dieselbe Überlegung wie bei `Listeneinfuhrnaht`**: die Naht ist der
// Composition Root dieser Fähigkeit — sie baut Klient, Zwischenspeicher,
// Drossel und Deckablage. Gegen Doubles geprüft bliebe genau das ungeprüft,
// was sie tut.
//
// **Und der Stand wird hier wirklich gemessen**: das Deck ändert sich
// zwischen Vorschau und Übernahme, weil ein echter Speicher dahintersteht.
//
// **Zwei Dinge stehen an jedem Lauf** (§3): ein eigenes `HOME` **und** eine
// umgelenkte Adresse.
package app_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dieDecklisteImLauf nennt eine eindeutige Zeile, eine offene und eine
// Merkliste.
const dieDecklisteImLauf = "Deck\n2 Erste Probe (probe) 1\n" +
	"3 Zweite Probe\n\nMaybeboard\n1 Erste Probe (probe) 1\n"

// listenprobekarte ist eine Kartenantwort, wie Scryfall sie schickt — dieselbe
// Vorlage wie in `listeneinfuhrnaht_test.go`, nur mit ihren Werten gefüllt.
func listenprobekarte(kennung int, name string, nummer int) string {
	return fmt.Sprintf(listenkarte, kennung, name, nummer)
}

// deckeinfuhrnaht baut die Naht über einem frischen Bestand und einer
// Gegenstelle — und gibt die Datenbank dazu.
func deckeinfuhrnaht(t *testing.T) (*app.Deckeinfuhrnaht, *sql.DB) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			// **Ein genauer Name samt Setangabe gibt genau einen** — so
			// verhält sich der echte Dienst (nachgemessen am 2026-08-30),
			// und dieselbe Bedingung steht in `listeneinfuhrnaht_test.go`.
			if strings.Contains(anfrage.URL.RawQuery, "%21%22") &&
				strings.Contains(anfrage.URL.RawQuery, "e%3A") {
				schreibeRoh(t, schreiben, `{"object":"list","has_more":false,`+
					`"total_cards":1,"data":[`+
					listenprobekarte(1, "Erste Probe", 1)+`]}`)

				return
			}

			schreibeRoh(t, schreiben, `{"object":"list","has_more":false,`+
				`"total_cards":2,"data":[`+
				listenprobekarte(1, "Erste Probe", 1)+`,`+
				listenprobekarte(2, "Erste Probe", 2)+`]}`)
		}))

	t.Cleanup(dienst.Close)

	vorher := app.MitScryfalladresse(dienst.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	db := frischerBestand(t)

	naht, err := app.DeckeinfuhrUeber(db)
	if err != nil {
		t.Fatalf("DeckeinfuhrUeber: %v", err)
	}

	return naht, db
}

// einProbedeck legt ein Deck an und gibt seine Kennung.
func einProbedeck(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	deck, fehler := storage.NeueDeckablage(db).Anlegen(probekontext(t),
		domain.Deck{Name: "Einfuhr", Spiel: "magic"})
	if fehler != nil {
		t.Fatalf("das Deck anlegen: %v", fehler)
	}

	return deck.ID
}

// **Der ganze Weg: ansehen, dann legen.**
func TestDieDeckeinfuhrnahtFuehrtDenGanzenWeg(t *testing.T) {
	naht, db := deckeinfuhrnaht(t)
	deckID := einProbedeck(t, db)

	vorschau, err := naht.Vorschau(probekontext(t), deckID, dieDecklisteImLauf)
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	if vorschau.DeckID != deckID || vorschau.Stand == "" {
		t.Fatalf("die Vorschau kam als %+v an", vorschau)
	}

	if len(vorschau.Zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet 3: %+v", len(vorschau.Zeilen), vorschau.Zeilen)
	}

	if vorschau.Hauptdeck != 2 || vorschau.Sideboard != 0 {
		t.Errorf("gezählt sind %d/%d, erwartet 2/0",
			vorschau.Hauptdeck, vorschau.Sideboard)
	}

	// **Die Zeile mit Setangabe ist eindeutig, die ohne nicht — und die
	// Merkliste gehört gar nicht ins Deck.**
	if !vorschau.Zeilen[0].Uebernehmbar || vorschau.Zeilen[0].Printing == "" {
		t.Errorf("die eindeutige Zeile kam als %+v an", vorschau.Zeilen[0])
	}

	if vorschau.Zeilen[1].Uebernehmbar || len(vorschau.Zeilen[1].Treffer) != 2 {
		t.Errorf("die offene Zeile kam als %+v an", vorschau.Zeilen[1])
	}

	// **Der lesbare Bezug ist mitgekommen**, nicht nur die Kennung — ohne ihn
	// könnte niemand zwischen den zweien wählen.
	if vorschau.Zeilen[1].Treffer[0].Setname != "Probesatz" {
		t.Errorf("Setname %q", vorschau.Zeilen[1].Treffer[0].Setname)
	}

	if vorschau.Zeilen[2].Uebernehmbar || vorschau.Zeilen[2].Befund == "" {
		t.Errorf("die Merkliste kam als %+v an", vorschau.Zeilen[2])
	}

	bericht, err := naht.Uebernehmen(probekontext(t), deckID, vorschau.Stand,
		[]app.Deckeinfuhrzeile{{
			Printing: vorschau.Zeilen[0].Printing,
			Menge:    vorschau.Zeilen[0].Menge,
			Bereich:  vorschau.Zeilen[0].Bereich,
		}})
	if err != nil {
		t.Fatalf("Uebernehmen: %v", err)
	}

	if bericht.Gelegt != 1 || bericht.Stueck != 2 {
		t.Errorf("der Bericht kam als %+v an", bericht)
	}

	// **Der neue Stand ist ein anderer** — im Deck liegt jetzt etwas.
	if bericht.Stand == vorschau.Stand || bericht.Stand == "" {
		t.Errorf("der Stand blieb %q", bericht.Stand)
	}
}

// **Ein Stand von vorhin legt nichts.**
//
// **Gemessen und nicht behauptet**: zwischen Vorschau und Übernahme legt
// dieser Lauf selbst eine Karte ins Deck — genau das, was am Fenster ein
// zweites Fenster täte.
func TestDieDeckeinfuhrnahtWeistEinenAltenStandAb(t *testing.T) {
	naht, db := deckeinfuhrnaht(t)
	deckID := einProbedeck(t, db)

	vorschau, err := naht.Vorschau(probekontext(t), deckID, dieDecklisteImLauf)
	if err != nil {
		t.Fatalf("Vorschau: %v", err)
	}

	if _, fehler := storage.NeueDeckablage(db).PlatzSetzen(probekontext(t),
		deckID, domain.Deckplatz{
			Printing: vorschau.Zeilen[0].Printing, Menge: 1,
			Bereich: domain.BereichHauptdeck,
		}); fehler != nil {
		t.Fatalf("dazwischenlegen: %v", fehler)
	}

	_, fehler := naht.Uebernehmen(probekontext(t), deckID, vorschau.Stand,
		[]app.Deckeinfuhrzeile{{
			Printing: vorschau.Zeilen[0].Printing, Menge: 2,
			Bereich: "hauptdeck",
		}})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der Fehler ist kein Konflikt: %v", fehler)
	}
}

// **Ein Deck, das es nicht gibt, gibt keine Vorschau und nimmt nichts an.**
func TestDieDeckeinfuhrnahtOhneDeck(t *testing.T) {
	naht, _ := deckeinfuhrnaht(t)

	if _, err := naht.Vorschau(probekontext(t), 4711, "2 Erste Probe"); err == nil {
		t.Error("die Vorschau kam ohne Deck zurück")
	}

	if _, err := naht.Uebernehmen(probekontext(t), 4711, "a",
		[]app.Deckeinfuhrzeile{{
			Printing: "00000000-0000-4000-8000-000000000001",
			Menge:    1, Bereich: "hauptdeck",
		}}); err == nil {
		t.Error("das Übernehmen kam ohne Deck zurück")
	}
}

// **Ohne Zuhause gibt es keine Naht.**
//
// Der Zweig, der ohne diesen Lauf unbetreten bliebe: `Scryfallverzeichnis`
// scheitert, und die Naht wird gar nicht erst gebaut — statt eine halbe
// herauszugeben (§8). Derselbe Lauf steht bei der Listeneinfuhr daneben.
func TestDieDeckeinfuhrnahtOhneZuhause(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.DeckeinfuhrUeber(frischerBestand(t)); err == nil {
		t.Error("ohne Zuhause kam eine Naht heraus")
	}
}

// **Ohne Naht sagt es die Bindung selbst** — und sie stürzt nicht ab.
//
// **Sie kann fehlen**: die Naht braucht ein Zuhause für den Zwischenspeicher
// von Scryfall, und ohne das gibt es sie nicht.
func TestOhneDeckeinfuhrSagtEsDieBindung(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if anwendung.DecklisteVorschau(t.Context(), 7, "4 Probe").Fehler == nil {
		t.Error("DecklisteVorschau ohne Naht gab keinen Fehler")
	}

	if anwendung.DecklisteUebernehmen(t.Context(), 7, "a", nil).Fehler == nil {
		t.Error("DecklisteUebernehmen ohne Naht gab keinen Fehler")
	}
}

// **Beide Bindungen gehen denselben Weg wie die App-Methoden.**
//
// Ohne diesen Lauf blieben `Bindungen.DecklisteVorschau` und
// `.DecklisteUebernehmen` ungerufen — und eine Bindung, die kein Lauf
// betritt, ist Code mit voller Abdeckung, den niemand benutzt (§10, erste
// Dimension).
func TestDieBindungenDerDecklisteneinfuhr(t *testing.T) {
	naht, db := deckeinfuhrnaht(t)
	deckID := einProbedeck(t, db)

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitDeckeinfuhr(naht)

	// **Der Fensterkontext trägt das Profil** — [Bindungen] baut ihren
	// Kontext daraus, und ohne Profil weiß niemand, wessen Daten gemeint
	// sind. Im Betrieb setzt ihn `Bereit`, wenn das Fenster da ist.
	anwendung.Bereit(probekontext(t))

	bindungen := app.NeueBindungen(anwendung)

	gesehen := bindungen.DecklisteVorschau(t.Context(), deckID, dieDecklisteImLauf)
	if gesehen.Fehler != nil {
		t.Fatalf("Bindungen.DecklisteVorschau: %+v", gesehen.Fehler)
	}

	gelegt := bindungen.DecklisteUebernehmen(t.Context(), deckID, gesehen.Vorschau.Stand,
		[]app.Deckeinfuhrzeile{{
			Printing: gesehen.Vorschau.Zeilen[0].Printing,
			Menge:    2, Bereich: "hauptdeck",
		}})
	if gelegt.Fehler != nil {
		t.Fatalf("Bindungen.DecklisteUebernehmen: %+v", gelegt.Fehler)
	}

	if gelegt.Bericht.Gelegt != 1 {
		t.Errorf("der Bericht kam als %+v an", gelegt.Bericht)
	}

	// **Und ein Fehler kommt als Fehler an** — derselbe Stand ein zweites
	// Mal geht gegen ein Deck, das es so nicht mehr gibt.
	wieder := bindungen.DecklisteUebernehmen(t.Context(), deckID, gesehen.Vorschau.Stand,
		[]app.Deckeinfuhrzeile{{
			Printing: gesehen.Vorschau.Zeilen[0].Printing,
			Menge:    2, Bereich: "hauptdeck",
		}})
	if wieder.Fehler == nil {
		t.Error("der zweite Lauf kam durch")
	}
}
