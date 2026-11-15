package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Eine MTG-JSON-Datei landet als Set im Bestand.
//
// **Derselbe Weg wie eine `.mse-set`**: dasselbe `Ergebnis`, dieselbe Behandlung
// von Befunden (§8). Wer MTG JSON hat, soll es öffnen können, ohne es vorher in
// ein MSE-Paket zu verwandeln.
func TestMTGJSONLandetImBestand(t *testing.T) {
	speicher := &speicherDouble{}

	ergebnis, err := service.NeuerImporteur(speicher).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "probe.json"), "magic")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if ergebnis.SetID != 42 {
		t.Errorf("Kennung %d", ergebnis.SetID)
	}

	if ergebnis.Karten != 2 {
		t.Fatalf("%d Karten, erwartet waren 2", ergebnis.Karten)
	}

	if speicher.set.Name != "Probeset" {
		t.Errorf("Setname %q, erwartet %q", speicher.set.Name, "Probeset")
	}

	if speicher.set.Game != "magic" {
		t.Errorf("Spiel %q, erwartet %q", speicher.set.Game, "magic")
	}

	if name, _ := speicher.set.Cards[0].Value("name"); name != "Flammenball" {
		t.Errorf("Karte 1 heißt %q", name)
	}

	// Der mitgeführte Schlüssel steht an der Karte — nicht weggeworfen.
	if uuid, da := speicher.set.Cards[0].Value("uuid"); !da || uuid == "" {
		t.Error("»uuid« ist nicht mitgeführt worden")
	}

	// Ein Befund je Set und Feld, nicht je Karte — beide Karten führen »uuid«.
	if len(ergebnis.Befunde) != 1 {
		t.Errorf("Befunde: %v, erwartet einen", ergebnis.Befunde)
	}
}

// Eine AllSets-Datei mit mehreren Sets wird nicht stillschweigend halbiert.
//
// Sie führt mehr als ein Set, der Bestand kennt aber nur eines je Import. Das
// zu übergehen hieße, 400 Sets zu öffnen und eines zu zeigen — deshalb wird
// jedes angelegt und die Zahl gemeldet.
func TestAllSetsLegtJedesSetAn(t *testing.T) {
	speicher := &zaehlenderSpeicher{}

	ergebnis, err := service.NeuerImporteur(speicher).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "alle.json"), "magic")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.aufrufe != 2 {
		t.Errorf("%d Sets angelegt, erwartet 2", speicher.aufrufe)
	}

	if ergebnis.Karten != 2 {
		t.Errorf("Karten: %d, erwartet 2 — die Zahl gilt über alle Sets", ergebnis.Karten)
	}
}

// Ein Spiel ohne Zuordnung wird abgewiesen, nicht geraten.
//
// Eine erfundene Zuordnung ließe genau die Pakete nicht importieren, für die
// jemand sie braucht — und niemand sähe, dass sie erfunden ist.
func TestSpielOhneZuordnungWirdAbgewiesen(t *testing.T) {
	_, err := service.NeuerImporteur(&speicherDouble{}).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "probe.json"), "gibtsnicht")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "gibtsnicht") {
		t.Errorf("Fehler: %q — das Spiel ist nicht benannt", err)
	}
}

// zaehlenderSpeicher zählt, wie oft ein Set angelegt wurde.
type zaehlenderSpeicher struct {
	aufrufe int

	// scheitertAb lässt den n-ten Aufruf und alle danach scheitern. Null: nie.
	scheitertAb int
}

func (z *zaehlenderSpeicher) SetAnlegen(
	_ context.Context, _ *domain.Set, _ []domain.RawBlock,
) (int64, error) {
	z.aufrufe++

	if z.scheitertAb > 0 && z.aufrufe >= z.scheitertAb {
		return 0, errors.New("die Platte ist voll")
	}

	return int64(z.aufrufe), nil
}

// Eine Datei, die es nicht gibt, kommt als Fehler zurück.
func TestMTGJSONMeldetEineUnlesbareQuelle(t *testing.T) {
	_, err := service.NeuerImporteur(&speicherDouble{}).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "gibtsnicht.json"), "magic")
	if err == nil {
		t.Fatal("kein Fehler")
	}
}

// Ein Fehler des Bestands hält den Import an.
func TestMTGJSONMeldetEinenScheiterndenBestand(t *testing.T) {
	speicher := &speicherDouble{fehler: errors.New("die Platte ist voll")}

	_, err := service.NeuerImporteur(speicher).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "probe.json"), "magic")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "die Platte ist voll") {
		t.Errorf("Fehler: %q — der Grund fehlt", err)
	}
}

// **Eine Zuordnung, die zwei JSON-Felder auf dasselbe Kartenfeld legt, wird
// abgewiesen — von der Domäne, nicht vom Bestand.**
//
// Sie ist ein Fehler in der Zuordnung, kein Fehler in der Datei. Ohne die
// Prüfung liefe sie bis in `card_values` und scheiterte dort am Primärschlüssel
// `(card_id, feld)` — mit einer Meldung über die Datenbank statt über die
// Zuordnung, und je nach Reihenfolge mit halb angelegtem Bestand.
func TestEineZuordnungMitZweiWegenAufEinFeldWirdAbgewiesen(t *testing.T) {
	mtgjson.Aufnehmen(mtgjson.Zuordnung{
		Spiel: "doppelspiel",
		Felder: []mtgjson.Feldpaar{
			{JSON: []string{"name"}, Feld: "name"},
			{JSON: []string{"artist"}, Feld: "name"},
		},
	})

	_, err := service.NeuerImporteur(&speicherDouble{}).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "probe.json"), "doppelspiel")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler: %v — erwartet war ein Konflikt", err)
	}
}

// **Ein halb geglückter AllSets-Import verschweigt nicht, was schon dasteht.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9. Jedes Set einer AllSets-Datei wird in einer
// **eigenen** Transaktion abgelegt — anders geht es nicht, `SetAnlegen` legt
// eines an. Scheitert das zweite, bleibt das erste im Bestand, und der Aufrufer
// bekam nur einen Fehler: einen Bestand mit einem Set, von dem niemand wusste.
//
// **Der Weg ist nicht, alles zurückzunehmen** — das hieße, eine Transaktion über
// die ganze Datei zu spannen, und eine AllSets-Datei führt Hunderte Sets. Der
// Weg ist, zu sagen, was dasteht.
func TestEinHalberAllSetsImportNenntWasDasteht(t *testing.T) {
	speicher := &zaehlenderSpeicher{scheitertAb: 2}

	ergebnis, err := service.NeuerImporteur(speicher).
		MTGJSON(t.Context(), filepath.Join(testdaten, "mtgjson", "alle.json"), "magic")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if ergebnis == nil {
		t.Fatal("kein Ergebnis — was schon dasteht, bleibt damit unsichtbar")
	}

	if len(ergebnis.Angelegt) != 1 {
		t.Errorf("Angelegt: %v, erwartet das eine Set, das durchkam", ergebnis.Angelegt)
	}

	if ergebnis.Karten != 1 {
		t.Errorf("Karten: %d, erwartet die des angelegten Sets", ergebnis.Karten)
	}
}

// **Die Sets einer Datei werden gezählt, und ein Abbruch fällt zwischen
// zwei** (FMT-03.AK-02 und AK-04). Jedes Set ist eine eigene Transaktion;
// was vor dem Abbruch dastand, ist ganz — und benannt, wie beim Fehler oben.
func TestEinAbgebrochenerAllSetsImportNenntWasDasteht(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	defer abbrechen()

	var erledigt []int

	ctx = service.MitFortschritt(ctx, func(stand domain.Fortschritt) {
		if stand.Gesamt != 2 || stand.Einheit != domain.EinheitSet {
			t.Errorf("Stand: %+v", stand)
		}

		erledigt = append(erledigt, stand.Erledigt)

		// **Der Abbruch kommt, als das erste Set steht** — das zweite
		// entsteht nicht mehr.
		if stand.Erledigt == 1 {
			abbrechen()
		}
	})

	speicher := &zaehlenderSpeicher{}

	ergebnis, err := service.NeuerImporteur(speicher).
		MTGJSON(ctx, filepath.Join(testdaten, "mtgjson", "alle.json"), "magic")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fehler: %v", err)
	}

	if speicher.aufrufe != 1 || ergebnis == nil || len(ergebnis.Angelegt) != 1 {
		t.Errorf("angelegt: %d, Ergebnis %+v", speicher.aufrufe, ergebnis)
	}

	if len(erledigt) != 2 || erledigt[0] != 0 || erledigt[1] != 1 {
		t.Errorf("erledigt: %v", erledigt)
	}
}
