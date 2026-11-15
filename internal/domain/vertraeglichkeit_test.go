package domain_test

// **Was eine Fassung öffnen kann** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dieseFassung ist ein Manifest, das zwei Wanderungen kennt.
func dieseFassung() domain.Fassungsmanifest {
	return domain.Fassungsmanifest{
		Fassung: "0.5.0", Plattform: "darwin/arm64", Artefakt: "abc",
		Einstellungen: 1, Protokoll: 1,
		Wanderungen: []domain.Wanderungsangabe{
			{Name: "0001_anfang.sql", Summe: "aaa"},
			{Name: "0002_karten.sql", Summe: "bbb"},
		},
	}
}

// mitGrundlinie ist [dieseFassung] mit einer Grundlinie bis `bis`.
func mitGrundlinie(bis string) domain.Fassungsmanifest {
	manifest := dieseFassung()
	manifest.Grundlinie = bis

	return manifest
}

func TestVertraeglichkeit(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		manifest  domain.Fassungsmanifest
		verbucht  map[string]string
		plattform string
		urteil    domain.Vertraeglichkeitsurteil
		arten     []domain.Hindernisart
	}{
		{
			name:      "derselbe Stand passt",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa", "0002_karten.sql": "bbb"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilVertraeglich,
		},
		{
			// **Weniger verbucht als bekannt ist kein Hindernis**: die
			// Fassung wandert den Bestand vorwärts, das ist ihr Beruf.
			name:      "ein Bestand hinter der Fassung passt",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilVertraeglich,
		},
		{
			// **Der Rückschritt**: der Bestand hat etwas, das die Fassung
			// nicht kennt.
			name:      "eine unbekannte Wanderung weist ab",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa", "0003_decks.sql": "ccc"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilUnvertraeglich,
			arten:     []domain.Hindernisart{domain.HindernisWanderungFehlt},
		},
		{
			// **Der Fall, den eine Schemanummer nicht sieht** — dieselbe
			// Migration, anderer Inhalt.
			name:      "dieselbe Wanderung mit anderer Summe weist ab",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0002_karten.sql": "anders"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilUnvertraeglich,
			arten:     []domain.Hindernisart{domain.HindernisSummeAnders},
		},
		{
			// **Ein Bestand von vor der Prüfsummenspalte** ist nicht
			// unverträglich, weil er alt ist.
			name:      "eine leere Summe im Ledger ist kein Hindernis",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0002_karten.sql": ""},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilVertraeglich,
		},
		{
			name:      "eine fremde Plattform weist ab",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "linux/amd64",
			urteil:    domain.UrteilUnvertraeglich,
			arten:     []domain.Hindernisart{domain.HindernisPlattform},
		},
		{
			// **Und beides zusammen wird auch zusammen genannt** — wer nur
			// das erste Hindernis behebt, soll nicht überrascht werden.
			name:      "zwei Hindernisse stehen beide da",
			manifest:  dieseFassung(),
			verbucht:  map[string]string{"0003_decks.sql": "ccc"},
			plattform: "linux/amd64",
			urteil:    domain.UrteilUnvertraeglich,
			arten: []domain.Hindernisart{
				domain.HindernisPlattform, domain.HindernisWanderungFehlt,
			},
		},
		{
			// **Eine Fassung mit Grundlinie wandert nur, was auf ihr steht**
			// (Vermerk 0189): ein Bestand, dem ein Schritt bis zu ihr fehlt,
			// öffnet sie nicht — „hinter der Fassung" reicht dann nicht.
			name:      "ein Bestand vor der Grundlinie weist ab",
			manifest:  mitGrundlinie("0002_karten.sql"),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilUnvertraeglich,
			arten:     []domain.Hindernisart{domain.HindernisVorDerGrundlinie},
		},
		{
			// **Was nach der Grundlinie kommt, wandert sie wie bisher.**
			name:      "ein Bestand auf der Grundlinie passt",
			manifest:  mitGrundlinie("0001_anfang.sql"),
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilVertraeglich,
		},
		{
			// **Ungeklärt ist nicht unverträglich** (Abnahme aus dem Issue).
			name:      "ohne Manifest bleibt es ungeklärt",
			manifest:  domain.Fassungsmanifest{},
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilUngeklaert,
			arten:     []domain.Hindernisart{domain.HindernisKeinManifest},
		},
		{
			// **Und ein Manifest ohne Wanderungen ist keins**: es behauptete
			// „diese Fassung kennt keine Migration", und das ist bei einer
			// Anwendung mit Bestand keine Aussage, sondern eine Lücke.
			name: "ein Manifest ohne Wanderungen bleibt ungeklärt",
			manifest: domain.Fassungsmanifest{
				Fassung: "0.5.0", Plattform: "darwin/arm64",
			},
			verbucht:  map[string]string{"0001_anfang.sql": "aaa"},
			plattform: "darwin/arm64",
			urteil:    domain.UrteilUngeklaert,
			arten:     []domain.Hindernisart{domain.HindernisKeinManifest},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			urteil, hindernisse := domain.Vertraeglichkeit(
				fall.manifest, fall.verbucht, fall.plattform)

			if urteil != fall.urteil {
				t.Errorf("Urteil = %q, erwartet %q", urteil, fall.urteil)
			}

			if len(hindernisse) != len(fall.arten) {
				t.Fatalf("%d Hindernisse, erwartet %d: %+v",
					len(hindernisse), len(fall.arten), hindernisse)
			}

			for i, art := range fall.arten {
				if hindernisse[i].Art != art {
					t.Errorf("Hindernis %d = %q, erwartet %q", i, hindernisse[i].Art, art)
				}
			}
		})
	}
}

// **Die Stelle benennt, was im Weg steht** — ohne sie wüsste niemand, welche
// Wanderung gemeint ist.
func TestEinHindernisNenntSeineStelle(t *testing.T) {
	t.Parallel()

	_, hindernisse := domain.Vertraeglichkeit(dieseFassung(),
		map[string]string{"0003_decks.sql": "ccc"}, "darwin/arm64")

	if len(hindernisse) != 1 || hindernisse[0].Stelle != "0003_decks.sql" {
		t.Fatalf("Hindernisse = %+v", hindernisse)
	}
}

// **Bereit ist nur, was verträglich und ohne Hindernis ist** — und
// „ungeklärt" ist kein Nein zur Fassung, sondern eines zum automatischen Weg.
func TestBereitschaft(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		urteil      domain.Vertraeglichkeitsurteil
		hindernisse []domain.Hindernis
		bereit      bool
	}{
		{"verträglich und ohne Hindernis", domain.UrteilVertraeglich, nil, true},
		{
			"verträglich, aber zu wenig Platz", domain.UrteilVertraeglich,
			[]domain.Hindernis{{Art: domain.HindernisPlatz}},
			false,
		},
		{"ungeklärt", domain.UrteilUngeklaert, nil, false},
		{"unverträglich", domain.UrteilUnvertraeglich, nil, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befund := domain.Bereitschaft(fall.urteil, fall.hindernisse)

			if befund.Bereit != fall.bereit {
				t.Errorf("Bereit = %v, erwartet %v", befund.Bereit, fall.bereit)
			}

			// **Nie `null`** — eine leere Liste, die als `null` ankäme,
			// ließe die Seite über `undefined` stolpern (§3).
			if befund.Hindernisse == nil {
				t.Error("die Hindernisse kamen als nil")
			}
		})
	}
}

// **Ein Manifest ohne Formatfassungen ist ungeklärt — und kein Ja.**
//
// Befund 6 des Zweitblicks über Phase 5.2: das Manifest führte sie, und die
// Entscheidung sah sie nicht an. Ein Katalog, der sie verschweigt, sagt nicht,
// ob Einstellungen und Protokoll den Wechsel überstehen — und darauf ist
// `ungeklaert` die Antwort, kein erfundener Wert.
func TestEinManifestOhneFormatfassungenIstUngeklaert(t *testing.T) {
	t.Parallel()

	verbucht := map[string]string{"0001_start.sql": "aa"}

	mit := func(einstellungen, protokoll int) domain.Fassungsmanifest {
		return domain.Fassungsmanifest{
			Fassung: "0.5.0", Plattform: "darwin/arm64",
			Einstellungen: einstellungen, Protokoll: protokoll,
			Wanderungen: []domain.Wanderungsangabe{{Name: "0001_start.sql", Summe: "aa"}},
		}
	}

	for name, fall := range map[string]struct {
		manifest domain.Fassungsmanifest
		erwartet domain.Vertraeglichkeitsurteil
	}{
		"beide genannt":        {mit(1, 1), domain.UrteilVertraeglich},
		"neuere":               {mit(9, 9), domain.UrteilVertraeglich},
		"ohne Einstellungen":   {mit(0, 1), domain.UrteilUngeklaert},
		"ohne Protokoll":       {mit(1, 0), domain.UrteilUngeklaert},
		"mit negativer Angabe": {mit(1, -1), domain.UrteilUngeklaert},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			urteil, hindernisse := domain.Vertraeglichkeit(
				fall.manifest, verbucht, "darwin/arm64")

			if urteil != fall.erwartet {
				t.Fatalf("das Urteil ist %q mit %+v", urteil, hindernisse)
			}

			if urteil == domain.UrteilUngeklaert && len(hindernisse) != 1 {
				t.Errorf("ungeklärt mit %d Hindernissen", len(hindernisse))
			}
		})
	}
}
