package app_test

// **Der zweite Teil des Vertragstests** — dieselbe Anwendung, dieselbe Frage.
//
// Eine eigene Datei, weil `bindungen_test.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort stehen die Bindungen, die eine Datei
// oder ein Paket berühren, hier die über den Bestand einer geöffneten Sammlung.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestJedeLesendeBindungErreichtIhreAnwendung(t *testing.T) {
	stand := bindungsprobe(t)
	anwendung, bindungen := stand.Anwendung, stand.Bindungen
	schreiber, massen := stand.Schreiber, stand.Massen

	t.Run("Sets", func(t *testing.T) {
		antwort := bindungen.Sets(t.Context())
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Sets) != 1 || antwort.Sets[0].Name != "Aus dem Bestand" {
			t.Errorf("%+v", antwort.Sets)
		}
	})

	// **Der Datenstand geht denselben Weg** (#156): das Frontend fragt ihn alle
	// zwei Sekunden, und ohne Bindung fragte es ins Leere.
	t.Run("Datenstand", func(t *testing.T) {
		if antwort := bindungen.Datenstand(t.Context()); antwort.Fehler != nil {
			t.Errorf("%+v", antwort.Fehler)
		}
	})

	t.Run("Karten", func(t *testing.T) {
		antwort := bindungen.Karten(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Karten) != 1 || antwort.Karten[0].ID != "7" {
			t.Errorf("%+v", antwort.Karten)
		}
	})

	// **Setangaben und SetwertSetzen** (#567): die Naht reicht durch, und ein
	// leerer Feldname ist dort ein Befund — geprüft wird, dass er ankommt.
	t.Run("Setangaben", func(t *testing.T) {
		anwendung.MitSetwerten(&setwerteprobe{})

		if antwort := bindungen.Setangaben(t.Context(), "1"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	t.Run("SetwertSetzen", func(t *testing.T) {
		anwendung.MitSetwerten(&setwerteprobe{})

		if antwort := bindungen.SetwertSetzen(t.Context(), "1", "set code", "PRB", nil); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	// **Bezugsquellen geht ohne jede Naht** (#263) — sie ist das, was jemand
	// vor seinem ersten Paket sieht.
	t.Run("Bezugsquellen", func(t *testing.T) {
		antwort := bindungen.Bezugsquellen(t.Context())
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Quellen) == 0 {
			t.Error("keine Quelle")
		}
	})

	// **Die Einstellungen gehen denselben Weg** (#245).
	t.Run("Vorbehalte", func(t *testing.T) {
		antwort := bindungen.Vorbehalte(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Vorbehalte) != 1 || antwort.Vorbehalte[0].Inhalt != "roh" {
			t.Errorf("%+v", antwort.Vorbehalte)
		}
	})

	t.Run("FeldSetzen", func(t *testing.T) {
		if antwort := bindungen.FeldSetzen(t.Context(), "1", "1", "name", "Feuerball", nil); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if schreiber.feld != "name" || schreiber.wert != "Feuerball" {
			t.Errorf("beim Schreiber kam %q = %q an", schreiber.feld, schreiber.wert)
		}
	})

	t.Run("Verlauf", func(t *testing.T) {
		anwendung.MitVerlauf(&verlaufsprobe{})

		antwort := bindungen.Verlauf(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Schritte == nil {
			t.Error("die Schritte fehlen — eine leere Liste, kein `null`")
		}
	})

	t.Run("FelderVorschau", func(t *testing.T) {
		antwort := bindungen.FelderVorschau(t.Context(), "1", []string{"1"}, "typ", "Artefakt")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if massen.feld != "typ" || len(antwort.Zeilen) == 0 {
			t.Errorf("%q, %d Zeilen", massen.feld, len(antwort.Zeilen))
		}
	})

	t.Run("FelderSetzenMitErwartung", func(t *testing.T) {
		antwort := bindungen.FelderSetzenMitErwartung(t.Context(), "1", []app.FeldvorwertDTO{{KarteID: "1"}}, "typ", "Artefakt")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Vorgang != "vorgang8" || len(massen.erwartet) != 1 {
			t.Errorf("Vorgang %q, Erwartung %v", antwort.Vorgang, massen.erwartet)
		}
	})

	t.Run("KarteAnlegen", func(t *testing.T) {
		antwort := bindungen.KarteAnlegen(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	// **Die Zwischenablage** (#1168) — drei Bindungen, ein Weg.
	t.Run("Ablage", func(t *testing.T) {
		var tafel string

		anwendung.MitKartenablage(&ablagequelle{inhalt: "game_name: magic\ncard:\n"})
		anwendung.MitZwischenablage(
			func(context.Context) (string, error) { return tafel, nil },
			func(_ context.Context, inhalt string) error { tafel = inhalt; return nil },
		)

		if antwort := bindungen.AblageKopieren(t.Context(), "1", []string{"7"}); antwort.Fehler != nil {
			t.Fatalf("kopieren: %+v", antwort.Fehler)
		}

		if antwort := bindungen.AblageAusschneiden(t.Context(), "1", []string{"7"}); antwort.Fehler != nil {
			t.Fatalf("ausschneiden: %+v", antwort.Fehler)
		}

		if antwort := bindungen.AblageEinfuegen(t.Context(), "1"); antwort.Fehler != nil {
			t.Fatalf("einfuegen: %+v", antwort.Fehler)
		}
	})

	// **Die Gestalt der Kartenliste** (#1168) — drei Bindungen, ein Weg.
	t.Run("Kartenliste", func(t *testing.T) {
		anwendung.MitKartenliste(&listenquelle{})

		if antwort := bindungen.Kartenlistenaufbau(t.Context(), "1"); antwort.Fehler != nil {
			t.Fatalf("aufbau: %+v", antwort.Fehler)
		}

		if antwort := bindungen.KartenFiltern(t.Context(), "1", "bear"); antwort.Fehler != nil {
			t.Fatalf("filtern: %+v", antwort.Fehler)
		}

		if antwort := bindungen.ListenspaltenSetzen(t.Context(), "1",
			[]app.ListenspalteDTO{{Feld: "name", Sichtbar: true}}); antwort.Fehler != nil {
			t.Fatalf("spalten: %+v", antwort.Fehler)
		}

		if antwort := bindungen.ListensortierungSetzen(t.Context(),
			"1", "name", true); antwort.Fehler != nil {
			t.Fatalf("sortierung: %+v", antwort.Fehler)
		}
	})

	// **Suchen und Ersetzen** (#1168) — zwei Bindungen, ein Weg.
	t.Run("Ersetzen", func(t *testing.T) {
		anwendung.MitErsetzung(&ersetzungsquelle{schau: eineStelle()})

		frage := app.ErsetzungsfrageDTO{Text: "Flieg", Ersatz: "Renn"}

		if antwort := bindungen.ErsetzungVorschau(t.Context(), "1", frage); antwort.Fehler != nil {
			t.Fatalf("vorschau: %+v", antwort.Fehler)
		}

		if antwort := bindungen.ErsetzungAusfuehren(t.Context(), "1", frage); antwort.Fehler != nil {
			t.Fatalf("ausfuehren: %+v", antwort.Fehler)
		}
	})

	// **Kopieren und Löschen** (#625, #621) — dieselbe Naht wie das Anlegen.
	t.Run("KarteKopieren", func(t *testing.T) {
		antwort := bindungen.KarteKopieren(t.Context(), "1", "7")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.KarteID != "11" {
			t.Errorf("KarteID %q, erwartet `11`", antwort.KarteID)
		}
	})

	t.Run("KarteLoeschen", func(t *testing.T) {
		antwort := bindungen.KarteLoeschen(t.Context(), "1", "7", false)
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	t.Run("KartenTauschen", func(t *testing.T) {
		if antwort := bindungen.KartenTauschen(t.Context(), "1", "7", "9"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	t.Run("Kartenfelder", func(t *testing.T) {
		antwort := bindungen.Kartenfelder(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	t.Run("BildSetzen", func(t *testing.T) {
		anwendung.MitBildablage(service.NeueBildablage(t.TempDir()))
		anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
			return "", nil
		})

		if antwort := bindungen.BildSetzen(t.Context(), "1", "kunst"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	// **BildAusDatei ohne Datei**: die Bindung reicht durch, und ein leerer
	// Pfad ist dort ein Befund — geprüft wird, dass er ankommt.
	t.Run("BildAusDatei", func(t *testing.T) {
		if antwort := bindungen.BildAusDatei(t.Context(), "1", "kunst", ""); antwort.Fehler == nil {
			t.Error("der leere Pfad kam durch")
		}
	})

	// **BildEinsetzen ohne Datei**: dasselbe eine Stufe später
	// (#576).
	t.Run("BildEinsetzen", func(t *testing.T) {
		if antwort := bindungen.BildEinsetzen(t.Context(),
			"1", "1", "kunst", "", app.AuswahlDTO{Breite: 10, Hoehe: 10},
		); antwort.Fehler == nil {
			t.Error("der leere Pfad kam durch")
		}
	})

	// **Die beiden Auskünfte ohne Zugang** (#888): die Bindung reicht durch,
	// und ohne eingerichteten Zugang ist das dort ein Befund.
	t.Run("ScryfallRegeln", func(t *testing.T) {
		if antwort := bindungen.ScryfallRegeln(t.Context(), "Prüfstein"); antwort.Fehler == nil {
			t.Error("der fehlende Zugang kam durch")
		}
	})

	t.Run("ScryfallManakosten", func(t *testing.T) {
		if antwort := bindungen.ScryfallManakosten(t.Context(), "{W}"); antwort.Fehler == nil {
			t.Error("der fehlende Zugang kam durch")
		}
	})

	// **Die Sammeldaten ohne Zuhause** (#888): beide Bindungen reichen durch.
	t.Run("ScryfallSammeldaten", func(t *testing.T) {
		if antwort := bindungen.ScryfallSammeldaten(t.Context()); antwort.Fehler == nil &&
			antwort.Art == "" {
			t.Error("es kam weder ein Stand noch ein Fehler")
		}
	})

	// **Umgelenkte Adresse und ein eigenes Zuhause** (#888).
	//
	// # Was hier zweimal schiefging
	//
	// Ohne die Adresse ging der Lauf gegen `api.scryfall.com` und **lud
	// wirklich 25 MB**. Mit der Adresse allein ging er immer noch: der
	// Zwischenspeicher im **echten** Zuhause hatte das Verzeichnis vom
	// vorigen Lauf, und die Adresse der Datei darin zeigt auf
	// `data.scryfall.io` — die `MitScryfalladresse` nicht umlenkt.
	//
	// Ein Testlauf, der einen fremden Rechner befragt, ist keiner (§3).
	// Aufgefallen ist es beide Male daran, dass er **grün** war, wo ein
	// Fehler erwartet wurde.
	t.Run("ScryfallSammeldatenHolen", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		vorher := app.MitScryfalladresse("http://unerreichbar.invalid")
		t.Cleanup(func() { app.MitScryfalladresse(vorher) })

		if antwort := bindungen.ScryfallSammeldatenHolen(t.Context()); antwort.Fehler == nil {
			t.Error("ohne erreichbare Gegenstelle blieb es ohne Fehler")
		}
	})

	// **Zuschnittgrenze rechnet und liest nichts** — die einzige Bindung ohne
	// Kontext und ohne Umschlag.
	// **Zuschnittziehen ebenso** — beide rechnen ohne Kontext und ohne Umschlag.
	t.Run("Zuschnittziehen", func(t *testing.T) {
		gab := bindungen.Zuschnittziehen(t.Context(), 400, 100, 100, 80,
			app.AuswahlDTO{Breite: 125, Hoehe: 100}, false, 10, 0, 0, 0)

		if gab != (app.AuswahlDTO{X: 10, Breite: 125, Hoehe: 100}) {
			t.Errorf("gab %+v", gab)
		}
	})

	t.Run("Zuschnittgrenze", func(t *testing.T) {
		gab := bindungen.Zuschnittgrenze(t.Context(), 400, 100, 100, 80,
			app.AuswahlDTO{Breite: 400, Hoehe: 100}, false, "")

		if gab != (app.AuswahlDTO{Breite: 125, Hoehe: 100}) {
			t.Errorf("gab %+v", gab)
		}
	})

	t.Run("Symbolpalette", func(t *testing.T) {
		antwort := bindungen.Symbolpalette(t.Context(), "1", "kosten")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})

	// **Die Auswahllisten** (#1074). Ohne Spielpaket gibt es keine, und das
	// ist kein Fehler: die allermeisten Felder haben keine.
	t.Run("Wortlisten", func(t *testing.T) {
		antwort := bindungen.Wortlisten(t.Context(), "1", "type")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Bereiche == nil {
			t.Error("nil statt einer leeren Liste")
		}
	})

	// **Und eine Wahl darin.** Sie scheitert hier — es gibt keinen Bereich —,
	// und genau das soll sie: eine Wahl ins Leere darf keinen Wert schreiben.
	t.Run("WortWaehlen", func(t *testing.T) {
		if antwort := bindungen.WortWaehlen(t.Context(), "1", "1", "type", 0, []int{0}, nil); antwort.Fehler == nil {
			t.Error("die Wahl ins Leere kam ohne Fehler durch")
		}
	})
}
