package domain_test

// **Was ein Index über sich selbst weiß** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Der Satz, um den herum das gebaut ist
//
// **„Gleich langer Vektor bedeutet nicht gleicher Vektorraum."** Zwei Modelle
// mit 768 Dimensionen liefern Zahlen derselben Länge und nichts, was sich
// vergleichen ließe. Wer nur die Dimension prüfte, bekäme Treffer — falsche,
// und ohne Meldung.
//
// **Deshalb gibt es den Fingerabdruck.** Er nennt alles, was den Raum
// bestimmt: Modell, Revision, Artefakt, Zerteiler, Anweisungen, Pooling,
// Normalisierung, Dimension und Abstandsmaß.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einManifest ist ein vollständiges.
func einManifest() domain.Indexmanifest {
	return domain.Indexmanifest{
		Schemafassung:  1,
		Zerteilfassung: "abschnitte-1",
		Indexfassung:   1,
		Modell:         "nomic-embed-text",
		Modellrevision: "0fd5c1e",
		Artefakthash:   "sha256:1234",
		Laufzeit:       "ollama 0.12.3",
		Abstandsmass:   domain.AbstandCosinus,
		Vertrag: domain.KIEinbettungsvertrag{
			Dimension: 768, Pooling: domain.KIPoolingMittel,
			Normalisiert: true, Zerteiler: "nomic-bert",
		},
		Korpusstaende: []domain.Korpusstand{
			{Quelle: domain.QuelleOracle, Revision: "2026-09-01", Datum: "2026-09-01"},
		},
		Suchkonfiguration: "rrf-60",
	}
}

// **Ein vollständiges Manifest geht durch.**
func TestEinVollstaendigesIndexmanifestGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := einManifest().Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}

// **Und jedes unvollständige nicht.**
func TestEinUnvollstaendigesIndexmanifestWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.Indexmanifest){
		"ohne Schemafassung":  func(m *domain.Indexmanifest) { m.Schemafassung = 0 },
		"ohne Zerteilfassung": func(m *domain.Indexmanifest) { m.Zerteilfassung = "" },
		"ohne Indexfassung":   func(m *domain.Indexmanifest) { m.Indexfassung = 0 },
		"ohne Modell":         func(m *domain.Indexmanifest) { m.Modell = "" },
		"ohne Modellrevision": func(m *domain.Indexmanifest) { m.Modellrevision = "" },
		"ohne Laufzeit":       func(m *domain.Indexmanifest) { m.Laufzeit = "" },
		"ohne Suchkonfiguration": func(m *domain.Indexmanifest) {
			m.Suchkonfiguration = ""
		},
		"erfundenes Abstandsmaß": func(m *domain.Indexmanifest) {
			m.Abstandsmass = "gefuehlt"
		},
		"ohne Korpusstand": func(m *domain.Indexmanifest) { m.Korpusstaende = nil },
		"mit kaputtem Vertrag": func(m *domain.Indexmanifest) {
			m.Vertrag.Dimension = 0
		},
		"mit Korpusstand ohne Revision": func(m *domain.Indexmanifest) {
			m.Korpusstaende[0].Revision = ""
		},
		"mit erfundener Korpusquelle": func(m *domain.Indexmanifest) {
			m.Korpusstaende[0].Quelle = "geraten"
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			manifest := einManifest()
			kaputt(&manifest)

			if fehler := manifest.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}
		})
	}
}

// **Gleich langer Vektor heißt nicht gleicher Vektorraum.**
//
// # Warum dieser Lauf der wichtigste dieser Datei ist
//
// **Weil sein Gegenteil still ausfällt.** Ein Index aus Modell A, befragt mit
// Modell B derselben Dimension, gibt Treffer — sie sind nur falsch, und nichts
// sagt es. Das Issue: „Der neue Encoder darf nicht im alten Raum suchen."
func TestGleicheDimensionIstNichtDerselbeRaum(t *testing.T) {
	t.Parallel()

	alt := einManifest()

	for name, aendern := range map[string]func(*domain.Indexmanifest){
		"ein anderes Modell":      func(m *domain.Indexmanifest) { m.Modell = "all-minilm" },
		"eine andere Revision":    func(m *domain.Indexmanifest) { m.Modellrevision = "anders" },
		"ein anderes Artefakt":    func(m *domain.Indexmanifest) { m.Artefakthash = "sha256:anders" },
		"ein anderer Zerteiler":   func(m *domain.Indexmanifest) { m.Vertrag.Zerteiler = "anders" },
		"ein anderes Pooling":     func(m *domain.Indexmanifest) { m.Vertrag.Pooling = domain.KIPoolingErster },
		"ohne Normalisierung":     func(m *domain.Indexmanifest) { m.Vertrag.Normalisiert = false },
		"eine andere Anweisung":   func(m *domain.Indexmanifest) { m.Vertrag.Frageanweisung = "Query: " },
		"ein anderes Abstandsmaß": func(m *domain.Indexmanifest) { m.Abstandsmass = domain.AbstandPunkt },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			neu := alt
			neu.Vertrag = alt.Vertrag
			aendern(&neu)

			// **Die Dimension bleibt gleich** — genau darum geht es.
			if neu.Vertrag.Dimension != alt.Vertrag.Dimension {
				t.Fatalf("dieser Fall ändert die Dimension: %d", neu.Vertrag.Dimension)
			}

			if neu.Fingerabdruck() == alt.Fingerabdruck() {
				t.Fatalf("%q ergibt denselben Fingerabdruck", name)
			}

			if neu.PasstZu(alt) {
				t.Fatalf("%q gilt als derselbe Raum", name)
			}
		})
	}
}

// **Derselbe Raum bleibt derselbe — auch bei anderem Korpus.**
//
// **Der Korpusstand gehört nicht zum Raum.** Ein Dokument mehr im Index
// ändert nicht, wie Vektoren verglichen werden; wer ihn mitzählte, baute bei
// jedem Nachladen neu.
func TestEinAndererKorpusIstDerselbeRaum(t *testing.T) {
	t.Parallel()

	alt := einManifest()

	neu := einManifest()
	neu.Korpusstaende = append(neu.Korpusstaende, domain.Korpusstand{
		Quelle: domain.QuelleRuling, Revision: "2026-09-05", Datum: "2026-09-05",
	})

	if !neu.PasstZu(alt) {
		t.Fatal("ein zweiter Korpusstand machte einen anderen Vektorraum daraus")
	}

	if neu.Fingerabdruck() != alt.Fingerabdruck() {
		t.Fatal("der Fingerabdruck zählt den Korpus mit")
	}
}

// **Und ein Reranker hat seinen eigenen Abdruck.**
//
// Das Issue: „Reranker hat eigenen Fingerprint." Er ordnet um, was gefunden
// wurde — er bestimmt nicht, wie gesucht wird, und darf den Index deshalb
// nicht ungültig machen.
func TestDerRerankerHatSeinenEigenenAbdruck(t *testing.T) {
	t.Parallel()

	ohne := einManifest()

	mit := einManifest()
	mit.Rerankerabdruck = "bge-reranker-v2-m3@abc"

	if !mit.PasstZu(ohne) {
		t.Fatal("ein Reranker machte einen anderen Vektorraum daraus")
	}

	if mit.Fingerabdruck() == ohne.Fingerabdruck() {
		t.Fatal("der Fingerabdruck des Index zählt den Reranker mit")
	}
}

// **Die Abstandsmaße stehen als Aufzählung da.**
func TestDieAbstandsmasseStehenAlsAufzaehlungDa(t *testing.T) {
	t.Parallel()

	masse := domain.Abstandsmasse()
	if len(masse) != 2 {
		t.Fatalf("die Maße sind %v", masse)
	}

	domain.Abstandsmasse()[0] = "verbogen"

	if domain.Abstandsmasse()[0] == "verbogen" {
		t.Error("die Aufzählung ließ sich von außen ändern")
	}
}
