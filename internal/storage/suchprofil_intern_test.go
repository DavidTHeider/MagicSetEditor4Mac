package storage

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Suche bleibt im Profil** — gefunden von der Messung zu ENG-05.AK-02
// (#1730).
//
// # Was die Messung sah
//
// Die Markensuche las `karten_marken` ganz — über **jedes** Profil. Der Grund
// war nicht ein fehlender Index, sondern eine fehlende Bedingung: ohne Set
// („über alle Sets") klemmten weder die Volltextsuche noch die Markensuche
// noch die Feldnamen auf das Profil. Nachgestellt an der Kommandozeile am
// 2026-09-23: `mse suche Feuer` fand im eigenen Profil die Karte eines
// anderen Nutzers.
//
// # Warum hier und nicht an den Flächen
//
// **Weil es eine Abfrage ist, die zu weit greift.** Jede Fläche ruft dieselbe
// (Fenster, Kommandozeile, MCP, Dienst); an der Abfrage behoben, gilt es für
// alle vier.
func TestDieSucheBleibtImProfil(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	for _, anweisung := range []string{
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Fremd')`,
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES
		   (1, 'Eigen', 'magic', '2026-09-23', 1), (2, 'Fremd', 'magic', '2026-09-23', 2)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (1, 1, 0), (2, 2, 0)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES
		   (1, 'name', 'Feuerball', 0), (2, 'name', 'Feuerregen', 0),
		   (2, 'geheimfeld', 'Feuerfest', 1)`,
		`INSERT INTO karten_marken (card_id, marke) VALUES (1, 'feuerprobe'), (2, 'feuerwerk')`,
	} {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
			t.Fatalf("%s: %v", anweisung, fehler)
		}
	}

	leser := NeuerSetLeser(db)

	for profil, erwartet := range map[int64]int64{1: 1, 2: 2} {
		imProfil := MitProfil(t.Context(), profil)

		gefunden, fehler := leser.Suche(imProfil, domain.SearchQuery{Text: "Feuer"})
		if fehler != nil {
			t.Fatalf("Profil %d: %v", profil, fehler)
		}

		if len(gefunden.Hits) == 0 {
			t.Errorf("Profil %d findet seine eigene Karte nicht", profil)
		}

		for _, fund := range gefunden.Hits {
			if fund.CardID != erwartet {
				t.Errorf("Profil %d findet die Karte %d aus einem fremden Profil (%s: %s)",
					profil, fund.CardID, fund.Field, fund.Value)
			}
		}

		// **Auch eine fremde Setkennung trifft nichts.**
		fremdesSet := 3 - erwartet
		if gezielt, fehler := leser.Suche(imProfil, domain.SearchQuery{
			Text: "Feuer", SetID: fremdesSet,
		}); fehler != nil || len(gezielt.Hits) != 0 {
			t.Errorf("Profil %d findet im fremden Set %d: %+v (%v)", profil, fremdesSet, gezielt.Hits, fehler)
		}
	}

	felder, fehler := leser.Felder(ctx, 0)
	if fehler != nil {
		t.Fatalf("Felder: %v", fehler)
	}

	if slices.Contains(felder, "geheimfeld") {
		t.Errorf("die Feldnamen nennen ein Feld, das nur ein fremdes Profil führt: %v", felder)
	}
}

// **Ohne Profil sucht niemand** — dieselbe Antwort wie jede Wurzelabfrage:
// ein Ruf, der kein Profil nennt, sagt nicht, wessen Karten gemeint sind.
func TestOhneProfilSuchtNiemand(t *testing.T) {
	t.Parallel()

	leser := NeuerSetLeser(migrierteDatenbank(t))
	ohne := t.Context()
	frage := domain.SearchQuery{Text: "Feuer"}

	_, suchfehler := leser.Suche(ohne, frage)
	_, markenfehler := leser.Markentreffer(ohne, frage)
	_, felderfehler := leser.Felder(ohne, 0)

	for name, fehler := range map[string]error{
		"Suche": suchfehler, "Marken": markenfehler, "Felder": felderfehler,
	} {
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s ohne Profil: %v, erwartet eine Validierung", name, fehler)
		}
	}
}
