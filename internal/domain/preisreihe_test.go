package domain_test

// **Eine Preisreihe ist mehr als Quelle, Währung und Ausführung** — Werkstück
// A3 von Phase 5.3 (#1442).
//
// # Warum
//
// Bis hierher war Scryfall die einzige Quelle, und sie nennt **einen** Preis je
// Währung und Ausführung. MTGJSON führt dieselbe Karte bei vier Anbietern, auf
// zwei Märkten und in **zwei Richtungen**: was ein Händler verlangt, und was er
// zahlt. Wer das in einer Reihe zusammenlegte, verglich einen Einkaufspreis mit
// einem Verkaufspreis — und niemand sähe es.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Aufzählungen sagen, was sie zulassen.**
func TestDieReihenangabenSindAufzaehlungen(t *testing.T) {
	t.Parallel()

	t.Run("Markt", func(t *testing.T) {
		t.Parallel()

		for _, markt := range []domain.Preismarkt{
			domain.MarktPapier, domain.MarktOnline,
		} {
			if !domain.IstPreismarkt(markt) {
				t.Errorf("%q gilt nicht als Markt", string(markt))
			}
		}

		if domain.IstPreismarkt("mond") {
			t.Error("„mond“ kam durch")
		}
	})

	t.Run("Preisart", func(t *testing.T) {
		t.Parallel()

		for _, art := range []domain.Preisart{
			domain.PreisartVerkauf, domain.PreisartAnkauf,
		} {
			if !domain.IstPreisart(art) {
				t.Errorf("%q gilt nicht als Preisart", string(art))
			}
		}

		if domain.IstPreisart("tausch") {
			t.Error("„tausch“ kam durch")
		}
	})

	t.Run("Lage", func(t *testing.T) {
		t.Parallel()

		for _, lage := range []domain.Preislage{
			domain.PreisGenannt, domain.PreisWeggefallen, domain.PreisUngeprueft,
		} {
			if !domain.IstPreislage(lage) {
				t.Errorf("%q gilt nicht als Lage", string(lage))
			}
		}

		if domain.IstPreislage("vielleicht") {
			t.Error("„vielleicht“ kam durch")
		}
	})
}

// **Die Werte sind die des Anbieters** — nicht übersetzt.
//
// **Der Grund steht in §9**: `paper`, `mtgo`, `retail` und `buylist` sind
// MTGJSONs Vokabular, und eine deutsche Entsprechung in der Tafel machte jede
// Zeile, die von dort kommt, unlesbar für den, der beides vergleicht.
func TestDieWerteSindDieDesAnbieters(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		string(domain.MarktPapier):     "paper",
		string(domain.MarktOnline):     "mtgo",
		string(domain.PreisartVerkauf): "retail",
		string(domain.PreisartAnkauf):  "buylist",
	}

	for wert, erwartet := range faelle {
		if wert != erwartet {
			t.Errorf("%q, erwartet %q", wert, erwartet)
		}
	}
}
