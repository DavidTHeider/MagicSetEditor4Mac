package service_test

// **Die Initskripte laufen einmal je Set, nicht je Karte** —
// #1304.
//
// # Warum das über den Packbereich geprüft wird
//
// Weil `umgebungsbauer` paketintern ist und der Packbereich der erste
// Aufrufer, der viele Karten durch **einen** Bauer schickt. Was hier steht,
// ist deshalb die Wirkung und nicht die Buchführung: **woran man sieht**, dass
// die Initskripte einmal gelaufen sind.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// benannteKarte gibt eine Karte mit einem Namen.
func benannteKarte(id int64, name string) *domain.CardRow {
	return &domain.CardRow{
		ID: id, SetID: 1,
		Values: []domain.Value{{Field: "name", Text: name}},
	}
}

// **Ein Name, den das Initskript **eifrig** aus der Karte liest, friert auf
// der ersten Karte ein — und genau so macht MSE2 es.**
//
// `SetScriptContext::getContext` setzt vor den Initskripten
// `set.cards.front()` als Attrappe und lässt sie **einmal** laufen
// (`src/script/script_manager.cpp`, Zeilen 30–75, Fassung f9d9356 vom
// 2021-07-06). Ein `x := card.name` im Initskript sieht dort dieselbe erste
// Karte, gleich welche gerade gezeichnet wird.
//
// **Der Filter liest genau diesen Namen.** Träfe er je Karte deren eigenen
// Namen, liefe das Initskript je Karte — und dann wäre der Posten von 1.905 ms
// wieder da.
func TestDasInitskriptSiehtDieErsteKarte(t *testing.T) {
	t.Parallel()

	vorlage := service.Vorlage{
		Spielskript: `gemerkt := card.name`,
		Packs: []msepaket.Pack{
			{Name: "wie die erste", Filter: `card.name == gemerkt`},
		},
	}

	karten := []*domain.CardRow{
		benannteKarte(1, "erste"), benannteKarte(2, "zweite"), benannteKarte(3, "dritte"),
	}

	angebot, befunde, err := packbereichMit(karten, vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("unerwartete Befunde: %+v", befunde)
	}

	if len(angebot) != 1 || angebot[0].Karten != 1 {
		t.Errorf("das Angebot ist %+v — erwartet war **eine** Karte: die, "+
			"deren Name das Initskript gesehen hat", angebot)
	}
}

// **Und ein Name, den es als Block erklärt, sieht die Karte des Aufrufs.**
//
// Das ist die andere Hälfte derselben Zusage, und ohne sie wäre der Bauer
// falsch: „eine geschweifte Klammer schließt nichts ein — sie bekommt ihre
// Umgebung gereicht" (`msescript.Umgebung`). Liefe die Auflösung bei der
// **Erklärung**, sähen alle drei Karten den Namen der ersten, und jedes
// `mana_sort` des Packs rechnete mit der falschen Karte.
func TestEinBlockAusDemInitskriptSiehtDieEigeneKarte(t *testing.T) {
	t.Parallel()

	vorlage := service.Vorlage{
		Spielskript: `eigener := { card.name }`,
		Packs: []msepaket.Pack{
			{Name: "jede", Filter: `eigener() != ""`},
		},
	}

	karten := []*domain.CardRow{
		benannteKarte(1, "erste"), benannteKarte(2, "zweite"), benannteKarte(3, "dritte"),
	}

	angebot, befunde, err := packbereichMit(karten, vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("unerwartete Befunde: %+v", befunde)
	}

	if len(angebot) != 1 || angebot[0].Karten != len(karten) {
		t.Errorf("das Angebot ist %+v — erwartet waren alle %d Karten",
			angebot, len(karten))
	}
}

// **Ein Befund aus dem Initskript geht an jede Karte mit.**
//
// Er entsteht einmal — aber er gilt für jede Karte, die mit dieser Vorlage
// gezeichnet wird. Wer ihn nur an der ersten meldete, verlöre ihn beim
// Blättern, und die Karte danach wäre still kaputt (§8).
func TestEinBefundAusDemInitskriptGiltFuerJedeKarte(t *testing.T) {
	t.Parallel()

	vorlage := service.Vorlage{Spielskript: `x := gibt_es_nicht("a")`}

	for _, karte := range []*domain.CardRow{benannteKarte(1, "erste"), benannteKarte(2, "zweite")} {
		plan, befunde, err := service.NeuerPlanbauer().
			MitStilen(spielwoerter{vorlage, true}).Kartenplan(t.Context(), karte)
		if err != nil {
			t.Fatalf("Kartenplan: %v", err)
		}

		_ = plan

		if len(befunde) == 0 {
			t.Errorf("die Karte %d bekam keinen Befund aus dem Initskript", karte.ID)
		}
	}
}
