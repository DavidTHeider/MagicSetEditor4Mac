package domain_test

// **Das kanonische Schema, gegen das ein KI-Vorschlag geprüft wird** —
// Werkstück A6 von Phase 5.3
// (#1445).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einSchema baut ein Schema aus drei Feldern.
func einSchema() domain.KISchema {
	return domain.KISchema{
		Paket:   "magic-m15",
		Fassung: "2026-09-11",
		Felder: []domain.KISchemafeld{
			{Name: "name", Typ: domain.KIText, Schreibbar: true},
			{
				Name: "rarity", Typ: domain.KIAuswahl, Schreibbar: true,
				Werte: []string{"common", "uncommon", "rare"},
			},
			{Name: "casting cost", Typ: domain.KIText, Schreibbar: false},
			{
				Name: "keyword", Typ: domain.KIText, Schreibbar: true,
				Mehrfach: true,
			},
		},
	}
}

// **Ein Vorschlag, der ins Schema passt, kommt durch.**
func TestEinPassenderVorschlagKommtDurch(t *testing.T) {
	t.Parallel()

	befunde := einSchema().Pruefe(map[string][]string{
		"name":    {"Blitzschlag"},
		"rarity":  {"rare"},
		"keyword": {"Fliegend", "Eile"},
	})

	if len(befunde) != 0 {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Ein Feld, das es nicht gibt, wird abgewiesen.**
//
// Das Issue: „Unbekanntes Modell/**unbekanntes Feld**/Schema-Escape … erzeugt
// keine Kartenänderung."
func TestEinUnbekanntesFeldWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	befunde := einSchema().Pruefe(map[string][]string{"erfunden": {"x"}})

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Ein Feld, das nicht beschreibbar ist, ebenso.**
//
// **Nicht dasselbe wie ein unbekanntes**: `casting cost` gibt es, und trotzdem
// darf eine KI es nicht setzen. Wer beides gleich meldete, ließe den Unterschied
// zwischen „kenne ich nicht" und „darfst du nicht" verschwinden.
func TestEinNichtBeschreibbaresFeldWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	befunde := einSchema().Pruefe(map[string][]string{"casting cost": {"{R}"}})

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Ein Wert außerhalb einer Auswahl wird abgewiesen.**
func TestEinWertAusserhalbDerAuswahlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	befunde := einSchema().Pruefe(map[string][]string{"rarity": {"mythisch"}})

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}
}

// **Mehrere Werte gibt es nur, wo das Feld sie trägt.**
func TestMehrereWerteGibtEsNurWoDasFeldSieTraegt(t *testing.T) {
	t.Parallel()

	if befunde := einSchema().Pruefe(map[string][]string{
		"keyword": {"Fliegend", "Eile"},
	}); len(befunde) != 0 {
		t.Errorf("ein Mehrfachfeld wurde abgewiesen: %+v", befunde)
	}

	if befunde := einSchema().Pruefe(map[string][]string{
		"name": {"eins", "zwei"},
	}); len(befunde) != 1 {
		t.Errorf("ein Einfachfeld nahm zwei Werte: %+v", befunde)
	}
}

// **Ein leeres Schema weist alles ab — und ist kein Freibrief.**
//
// # Der Satz, um den es geht
//
// Das Issue: „Fehlendes/unlesbares Paket erzeugt für KI **keinen beliebigen
// Feldfreibrief**." Ein Schema, das nichts weiß, weiß nicht, dass alles
// erlaubt ist — es weiß nichts, und das ist das Gegenteil.
func TestEinLeeresSchemaIstKeinFreibrief(t *testing.T) {
	t.Parallel()

	leer := domain.KISchema{}

	befunde := leer.Pruefe(map[string][]string{"name": {"Blitzschlag"}})

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}

	if !leer.Unbekannt() {
		t.Error("das leere Schema gilt nicht als unbekannt")
	}

	if einSchema().Unbekannt() {
		t.Error("ein volles Schema gilt als unbekannt")
	}
}

// **Und ein leerer Vorschlag ändert nichts — das ist kein Befund.**
//
// **Ein Modell darf sich verweigern.** Das Issue nennt „Refusal" ausdrücklich
// neben den Fehlern: eine Antwort ohne Vorschlag ist eine Antwort, und sie
// führt zu keiner Kartenänderung — genau wie gewollt.
func TestEinLeererVorschlagIstKeinBefund(t *testing.T) {
	t.Parallel()

	if befunde := einSchema().Pruefe(nil); len(befunde) != 0 {
		t.Errorf("ein leerer Vorschlag gab Befunde: %+v", befunde)
	}
}

// **Die Fassung gehört zum Schema und wird mitgeführt.**
//
// Sie bindet einen Auftrag an den Stand, gegen den er geprüft wurde — ändert
// sich das Paket, passt der Vorschlag vielleicht nicht mehr.
func TestDieFassungGehoertZumSchema(t *testing.T) {
	t.Parallel()

	schema := einSchema()

	if schema.Kennung() == "" {
		t.Fatal("das Schema hat keine Kennung")
	}

	andere := einSchema()
	andere.Fassung = "2026-10-01"

	if andere.Kennung() == schema.Kennung() {
		t.Error("eine andere Fassung gibt dieselbe Kennung")
	}

	nochAndere := einSchema()
	nochAndere.Paket = "magic-m11"

	if nochAndere.Kennung() == schema.Kennung() {
		t.Error("ein anderes Paket gibt dieselbe Kennung")
	}
}
