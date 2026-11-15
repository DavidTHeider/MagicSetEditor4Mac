package domain_test

// **Ein Turnier als Datei** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gespieltesTurnier trägt Bestandskennungen, die in keiner Datei stehen
// dürfen — 4711 für das Deck, 815 für den Druck, 9001 für die Teilnehmenden.
func gespieltesTurnier() domain.Turnier {
	return domain.Turnier{
		ID: 3, Fassung: 7, Name: "Freitag", Spiel: "magic", FormatID: 12,
		Einzelspiel: true, Notiz: "Hinterzimmer", Abgeschlossen: "2026-09-20T20:00:00Z",
		Teilnehmer: []domain.Teilnehmer{
			{ID: 9001, Name: "Anna", DeckID: 4711, Deckabzug: []domain.Abzugsplatz{
				{
					ID: 5, Bezeichnung: "Blitz", Printing: "abc-123", DruckID: 815,
					Menge: 4, Bereich: domain.BereichHauptdeck,
				},
				// **So zieht die Leitung einen echten Druck ab** — ohne
				// Bezeichnung, mit Kennung (Befund 1 des Zweitblicks über 6.6).
				{ID: 6, Printing: "def-456", Menge: 2, Bereich: domain.BereichSideboard},
			}},
			{ID: 9002, Name: "Bert", Ausgestiegen: true},
			{ID: 9003, Name: "Cleo"},
		},
		Runden: []domain.Runde{
			{ID: 40, Nummer: 1, Paarungen: []domain.Paarung{
				{ID: 41, A: 9001, B: 9002, Ergebnis: &domain.Ergebnis{SpieleA: 1}},
				{ID: 42, A: 9003},
			}},
			{ID: 50, Nummer: 2, Paarungen: []domain.Paarung{
				{ID: 51, A: 9001, B: 9003, Ergebnis: &domain.Ergebnis{SpieleB: 1}},
			}},
		},
	}
}

func TestEinTurnierReistOhneBestandskennungen(t *testing.T) {
	t.Parallel()

	roh := domain.FormeTurnier(domain.AusgabeVon(gespieltesTurnier(), "Pauper"))

	for _, kennung := range []string{"4711", "815", "9001", "\"id\"", "deckID", "druckID"} {
		if bytes.Contains(roh, []byte(kennung)) {
			t.Errorf("die Datei trägt %s:\n%s", kennung, roh)
		}
	}

	austausch, befunde, err := domain.ZerlegeTurnier(roh)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("zerlegen: %v, %v", befunde, err)
	}

	if austausch.Fassung != domain.TurnieraustauschFassung || austausch.Turnier.Format != "Pauper" {
		t.Errorf("der Umschlag: %+v", austausch)
	}

	gelesen := austausch.Turnier.AlsTurnier()
	if gelesen.ID != 0 || gelesen.FormatID != 0 || gelesen.Name != "Freitag" || !gelesen.Einzelspiel ||
		gelesen.Notiz != "Hinterzimmer" || gelesen.Abgeschlossen != "2026-09-20T20:00:00Z" {
		t.Errorf("der Kopf: %+v", gelesen)
	}

	anna := gelesen.Teilnehmer[0]
	if len(gelesen.Teilnehmer) != 3 || anna.DeckID != 0 || len(anna.Deckabzug) != 2 ||
		anna.Deckabzug[1].Printing != "def-456" || anna.Deckabzug[1].Bezeichnung != "" ||
		anna.Deckabzug[0].DruckID != 0 || anna.Deckabzug[0].Printing != "abc-123" ||
		!gelesen.Teilnehmer[1].Ausgestiegen {
		t.Fatalf("die Teilnehmenden: %+v", gelesen.Teilnehmer)
	}

	// **Die Paarungen verweisen über vorläufige Kennungen** — die Stelle in
	// der Datei, ab 1.
	erste := gelesen.Runden[0].Paarungen
	if erste[0].A != 1 || erste[0].B != 2 || erste[0].Ergebnis.SpieleA != 1 || !erste[1].IstBye() || erste[1].A != 3 {
		t.Errorf("die erste Runde: %+v", erste)
	}

	if zweite := gelesen.Runden[1]; zweite.Nummer != 2 || zweite.Paarungen[0].Ergebnis.SpieleB != 1 {
		t.Errorf("die zweite Runde: %+v", zweite)
	}
}

// **Ein Feld, das niemand kennt, ist ein Befund und kein Nein** (§8).
func TestEinFremdesFeldImTurnierIstEinBefund(t *testing.T) {
	t.Parallel()

	roh := `{"fassung": 1, "zukunft": true, "turnier": {"name": "T", "spiel": "magic"}}`

	austausch, befunde, err := domain.ZerlegeTurnier([]byte(roh))
	if err != nil || austausch.Turnier.Name != "T" || len(befunde) != 1 || befunde[0].Schluessel != "zukunft" {
		t.Errorf("ergab %+v, %+v, %v", austausch, befunde, err)
	}
}

func TestEineFalscheTurnierdateiIstEinNein(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein JSON":        `{"fassung": x}`,
		"unvollständig":    `{"fassung":`,
		"kein Objekt":      `[1]`,
		"zu tief":          `{"turnier": ` + strings.Repeat(`[`, 20) + strings.Repeat(`]`, 20) + `}`,
		"Fassung ist Text": `{"fassung": "eins"}`,
		"Fassung null":     `{"fassung": 0, "turnier": {}}`,
		"Fassung zwei":     `{"fassung": 2, "turnier": {}}`,
		"Turnier ist Text": `{"fassung": 1, "turnier": "Freitag"}`,
		"ohne Namen":       `{"fassung": 1, "turnier": {"spiel": "magic"}}`,
		"ohne Spiel":       `{"fassung": 1, "turnier": {"name": "T"}}`,
		"namenlos dabei":   `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic", "teilnehmer": [{"name": " "}]}}`,
		"leere Zeile": `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic", "teilnehmer": [
			{"name": "A", "deckabzug": [{"bezeichnung": "Blitz", "menge": 0, "bereich": "hauptdeck"}]}]}}`,
		"fremder Bereich": `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic", "teilnehmer": [
			{"name": "A", "deckabzug": [{"bezeichnung": "Blitz", "menge": 1, "bereich": "keller"}]}]}}`,
		"Zeile ohne Bezeichnung und Druck": `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic", "teilnehmer": [
			{"name": "A", "deckabzug": [{"menge": 1, "bereich": "hauptdeck"}]}]}}`,
		"Lücke in den Runden": mitRunden(`{"nummer": 2, "paarungen": [{"a": 1, "b": 2}]}`),
		// **Was die Leitung nie legt** — eine Runde ohne Paarung, zwei Byes
		// (Befund 2 des Zweitblicks über 6.6).
		"leere Runde":        mitRunden(`{"nummer": 1, "paarungen": []}`),
		"zwei Byes in einer": mitRunden(`{"nummer": 1, "paarungen": [{"a": 1}, {"a": 2}]}`),
		"Seite A fehlt":      mitRunden(`{"nummer": 1, "paarungen": [{"a": 3, "b": 2}]}`),
		"Seite B fehlt":      mitRunden(`{"nummer": 1, "paarungen": [{"a": 1, "b": -1}]}`),
		"gegen sich selbst":  mitRunden(`{"nummer": 1, "paarungen": [{"a": 1, "b": 1}]}`),
		"zweimal in der Runde": mitRunden(`{"nummer": 1, "paarungen": [{"a": 1, "b": 2},
			{"a": 2}]}`),
		"unmögliches Ergebnis": mitRunden(`{"nummer": 1, "paarungen": [
			{"a": 1, "b": 2, "ergebnis": {"spieleA": -1}}]}`),
		"Bye mit Ergebnis": mitRunden(`{"nummer": 1, "paarungen": [
			{"a": 1, "ergebnis": {"spieleA": 2}}]}`),
		"offener Tisch vor der nächsten Runde": mitRunden(
			`{"nummer": 1, "paarungen": [{"a": 1, "b": 2}]}`,
			`{"nummer": 2, "paarungen": [{"a": 1, "b": 2}]}`),
		"abgeschlossen mit offenem Tisch": `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic",
			"abgeschlossen": "2026-09-20T20:00:00Z", "teilnehmer": [{"name": "A"}, {"name": "B"}],
			"runden": [{"nummer": 1, "paarungen": [{"a": 1, "b": 2}]}]}}`,
		"abgeschlossen ohne Datum": `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic",
			"abgeschlossen": "gestern", "teilnehmer": [{"name": "A"}],
			"runden": [{"nummer": 1, "paarungen": [{"a": 1}]}]}}`,
	}

	for name, roh := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, err := domain.ZerlegeTurnier([]byte(roh)); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler ist %v, erwartet ErrValidation", err)
			}
		})
	}

	t.Run("zu groß", func(t *testing.T) {
		t.Parallel()

		riesig := bytes.Repeat([]byte(" "), domain.TurnieraustauschHoechstens+1)
		if _, _, err := domain.ZerlegeTurnier(riesig); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler ist %v, erwartet ErrValidation", err)
		}
	})
}

// mitRunden baut eine Datei mit zwei Teilnehmenden und den genannten Runden.
func mitRunden(runden ...string) string {
	return `{"fassung": 1, "turnier": {"name": "T", "spiel": "magic",
		"teilnehmer": [{"name": "A"}, {"name": "B"}], "runden": [` + strings.Join(runden, ",") + `]}}`
}
