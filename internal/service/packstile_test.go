package service_test

// **Jede Karte bekommt das Initskript ihres eigenen Stilbogens** — Befund 1 des
// Zweitblicks über Phase 4.1
// (#1288).
//
// # Was der Befund war
//
// Der Packfilter las **einmal** `StilFuerSet` und baute daraus **einen**
// Umgebungsbauer für alle Karten. Eine Karte mit eigenem `stylesheet:` bekam
// damit das Initskript eines fremden Stils.
//
// MSE2 hält seine Kontexte in einer Abbildung über den Stilbogen und wählt für
// jede Karte zuerst ihren eigenen: `contexts.try_emplace(stylesheet.get())`
// und `stylesheetForP(card)` (`src/script/script_manager.cpp`, Zeilen 30 und
// 64, Fassung f9d9356 vom 2021-07-06).
//
// # Warum das nicht auffiel
//
// Weil es **vor** der Portierung genauso falsch war: der alte Weg reichte
// dieselbe Setvorlage in `umgebungAus`. Die Portierung hat den Fehler nicht
// gemacht, sondern sichtbar — ein Bauer je Stil ist eine Frage, die sich
// vorher niemand stellte.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stileJeKarte gibt je nach `stylesheet:`-Wert der Karte eine andere Vorlage.
type stileJeKarte struct {
	set    service.Vorlage
	eigene map[string]service.Vorlage
}

func (s stileJeKarte) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	return s.set, nil, true
}

func (s stileJeKarte) StilFuerKarte(_ context.Context, karte *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	wert, _ := karte.Value(service.Kartenstilfeld)
	if vorlage, da := s.eigene[wert]; da {
		return vorlage, nil, true
	}

	return s.set, nil, true
}

// karteMitEigenemStil gibt eine Karte mit eigenem Stil und einem Namen.
func karteMitEigenemStil(id int64, stil, name string) *domain.CardRow {
	return &domain.CardRow{
		ID: id, SetID: 1,
		Values: []domain.Value{
			{Field: service.Kartenstilfeld, Text: stil},
			{Field: "name", Text: name},
		},
	}
}

// packvorlageMitInitskript baut eine Vorlage, deren Initskript einen Namen setzt.
func packvorlageMitInitskript(paket, gemerkt string) service.Vorlage {
	return service.Vorlage{
		Stil: msepaket.Stil{
			Paket:      paket + ".mse-style",
			Initskript: `erkannt := "` + gemerkt + `"`,
		},
		Packs: []msepaket.Pack{
			{Name: "je Stil", Filter: `erkannt == card.name`},
		},
	}
}

// **Zwei Stilbögen im selben Set — jede Karte trifft ihren eigenen.**
//
// Der Filter vergleicht den Kartennamen mit dem, was das **Initskript des
// Stils** gesetzt hat. Läuft für beide Karten dasselbe Initskript, trifft nur
// eine; läuft je Karte das ihres Stils, treffen beide.
func TestJedeKarteBekommtDasInitskriptIhresStils(t *testing.T) {
	t.Parallel()

	eins := packvorlageMitInitskript("stil-eins", "eins")
	zwei := packvorlageMitInitskript("stil-zwei", "zwei")

	karten := []*domain.CardRow{
		karteMitEigenemStil(1, "stil-eins", "eins"),
		karteMitEigenemStil(2, "stil-zwei", "zwei"),
	}

	bereich := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(stileJeKarte{
			set:    eins,
			eigene: map[string]service.Vorlage{"stil-eins": eins, "stil-zwei": zwei},
		}))

	angebot, befunde, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("unerwartete Befunde: %+v", befunde)
	}

	if len(angebot) != 1 || angebot[0].Karten != len(karten) {
		t.Errorf("das Angebot ist %+v — erwartet waren beide Karten, jede mit "+
			"dem Initskript ihres eigenen Stilbogens", angebot)
	}
}

// **Ein Set ohne Karten baut keinen Bauer und meldet nichts.**
func TestEinSetOhneKartenGibtLeerePackmengen(t *testing.T) {
	t.Parallel()

	angebot, befunde, err := packbereichMit(nil, packvorlage(), true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %+v", befunde)
	}

	for _, art := range angebot {
		if art.Karten != 0 {
			t.Errorf("die Packart %q zählt %d Karten in einem leeren Set",
				art.Name, art.Karten)
		}
	}
}

// ohneStilFuerKarte findet für keine Karte einen eigenen Stil.
type ohneStilFuerKarte struct{ set service.Vorlage }

func (o ohneStilFuerKarte) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	return o.set, nil, true
}

func (o ohneStilFuerKarte) StilFuerKarte(context.Context, *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	return service.Vorlage{}, nil, false
}

// **Findet die Karte keinen eigenen Stil, gilt der des Sets.**
//
// **Und nicht ein leerer Verbund.** Ein Set, dessen Stilpaket sich nicht
// öffnen lässt, zöge sonst aus nichts: die Packarten stünden da, und keine
// Karte passte zu einer — obwohl der Lauf die Vorlage bereits gelesen hat
// (§8, „ein Set mit 300 lesbaren Karten wird geöffnet").
func TestOhneEigenenStilGiltImPackDerDesSets(t *testing.T) {
	t.Parallel()

	vorlage := packvorlageMitInitskript("stil-eins", "eins")

	karten := []*domain.CardRow{karteMitEigenemStil(1, "gibtsnicht", "eins")}

	bereich := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(ohneStilFuerKarte{set: vorlage}))

	angebot, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Karten != 1 {
		t.Errorf("das Angebot ist %+v — erwartet war die eine Karte mit dem "+
			"Initskript des Sets", angebot)
	}
}

// **Und ohne Stilquelle überhaupt gilt, was der Lauf schon hat.**
//
// Die eigenen Packarten eines Sets gelten auch ohne Vorlage (#1081) — dann
// steht ein Filter da und keine Stilquelle. Der Lauf darf daran nicht
// scheitern.
func TestOhneStilquelleGiltDieSetvorlageDesLaufs(t *testing.T) {
	t.Parallel()

	karten := []*domain.CardRow{karteMitEigenemStil(1, "", "eins")}

	bereich := service.NeuerPackbereich(nil,
		setkarten{
			karten: karten,
			eigene: []*domain.PackArt{{Name: "eigene", Filter: `card.name != ""`}},
		},
		service.NeuerPlanbauer())

	angebot, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Karten != 1 {
		t.Errorf("das Angebot ist %+v — erwartet war die eine Karte", angebot)
	}
}
