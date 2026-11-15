package service_test

// **Die Zusammenführung zweier Listen** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Was hier gemessen wird
//
// **Dass Ränge zusammengeführt werden und keine Punktwerte.** Das Issue
// verbietet das andere ausdrücklich: „BM25, Cosinus und Rerankwerte nicht roh
// addieren und nicht als Wahrscheinlichkeit ausgeben." Was hier passiert, ist
// Reciprocal Rank Fusion — ein dokumentiertes Verfahren mit einer Konstanten,
// die dasteht.
//
// **Und dass ein Filter vor der Grenze greift.** „Filterfall darf nicht allein
// durch ungefiltertes Top-k seine Treffer verlieren": wer nach Regelstellen
// sucht, bekommt Regelstellen — auch wenn zwanzig Kartentexte besser passen.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// retrievalablagedoppel hält Treffer, die ein Lauf vorgibt.
type retrievalablagedoppel struct {
	lexikalisch []domain.Retrievaltreffer
	semantisch  []domain.Retrievaltreffer

	manifest       domain.Indexmanifest
	kein           bool
	manifestfehler error

	lexfehler error
	semfehler error

	// gefragt hält die Grenzen, mit denen gesucht wurde.
	gefragt []int
}

func (a *retrievalablagedoppel) Lexikalisch(_ context.Context, _ string, grenze int) (
	[]domain.Retrievaltreffer, error,
) {
	a.gefragt = append(a.gefragt, grenze)

	return a.lexikalisch, a.lexfehler
}

func (a *retrievalablagedoppel) Semantisch(_ context.Context, _ []float32, grenze int) (
	[]domain.Retrievaltreffer, error,
) {
	a.gefragt = append(a.gefragt, grenze)

	return a.semantisch, a.semfehler
}

func (a *retrievalablagedoppel) Manifest(context.Context) (domain.Indexmanifest, error) {
	if a.kein {
		return domain.Indexmanifest{}, domain.MitGrund(domain.ErrNotFound, nil,
			"es ist kein Suchindex aktiv")
	}

	return a.manifest, a.manifestfehler
}

// retrievaleinbetterdoppel gibt feste Vektoren.
type retrievaleinbetterdoppel struct {
	raum    domain.Vektorraum
	vektor  []float32
	fehler  error
	zweimal bool

	// groesster ist der größte Stapel, der je gefragt wurde.
	groesster int

	// gesehen hält die Texte, die wirklich eingebettet wurden.
	gesehen []string
}

func (e *retrievaleinbetterdoppel) Raum() domain.Vektorraum { return e.raum }

func (e *retrievaleinbetterdoppel) Einbetten(_ context.Context, texte []string) (
	[][]float32, error,
) {
	e.gesehen = append(e.gesehen, texte...)
	e.groesster = max(e.groesster, len(texte))

	if e.fehler != nil {
		return nil, e.fehler
	}

	vektoren := make([][]float32, 0, len(texte))
	for range texte {
		vektoren = append(vektoren, e.vektor)
	}

	if e.zweimal {
		vektoren = append(vektoren, e.vektor)
	}

	return vektoren, nil
}

// einRetrievaltreffer baut einen.
func einRetrievaltreffer(kennung string, rang int, quelle domain.Retrievalquelle) domain.Retrievaltreffer {
	return domain.Retrievaltreffer{
		Dokument: domain.Retrievaldokument{
			Kennung: kennung, Quelle: quelle, Spiel: "magic", Sprache: "en",
			Titel: kennung, Text: "Text " + kennung,
			Quellrevision: "1", Abgerufen: time.Now(),
			Inhaltshash:  domain.Retrievalhash(kennung),
			Sichtbarkeit: domain.SichtbarOeffentlich,
		},
		Rang: rang,
	}
}

// einRetrievalmanifest passt zum Einbetter darunter.
func einRetrievalmanifest() domain.Indexmanifest {
	return domain.Indexmanifest{
		Schemafassung: 1, Zerteilfassung: "abschnitte-1", Indexfassung: 1,
		Modell: "all-minilm", Modellrevision: "1b226e2802db",
		Laufzeit: "ollama 0.33.3", Abstandsmass: domain.AbstandCosinus,
		Vertrag: domain.KIEinbettungsvertrag{
			Dimension: 3, Pooling: domain.KIPoolingMittel,
			Normalisiert: true, Zerteiler: "bert",
		},
		Korpusstaende: []domain.Korpusstand{
			{Quelle: domain.QuelleOracle, Revision: "1"},
		},
		Suchkonfiguration: service.Rangfusionsname,
	}
}

// einRetrievaldienst baut ihn über den beiden Doppeln.
func einRetrievaldienst(ablage *retrievalablagedoppel, einbetter service.Einbetter) *service.Retrievaldienst {
	return service.NeuerRetrievaldienst(ablage, einbetter)
}

// **Zwei Listen werden über ihre Ränge zusammengeführt.**
//
// # Die Rechnung, und warum sie dasteht
//
// **Reciprocal Rank Fusion**: je Liste zählt `1 / (k + Rang)` mit k = 60.
// Ein Dokument, das in **beiden** Listen steht, sammelt zweimal — und steht
// deshalb vor einem, das nur eine Liste kennt, auch wenn es dort weiter hinten
// lag. Genau das ist der Gewinn der Hybridsuche.
func TestZweiListenWerdenUeberIhreRaengeZusammengefuehrt(t *testing.T) {
	t.Parallel()

	ablage := &retrievalablagedoppel{
		manifest: einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{
			einRetrievaltreffer("nurLexikalisch", 1, domain.QuelleOracle),
			einRetrievaltreffer("beide", 2, domain.QuelleOracle),
		},
		semantisch: []domain.Retrievaltreffer{
			einRetrievaltreffer("nurSemantisch", 1, domain.QuelleOracle),
			einRetrievaltreffer("beide", 2, domain.QuelleOracle),
		},
	}

	dienst := einRetrievaldienst(ablage, &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	})

	ergebnis, fehler := dienst.Suchen(t.Context(), service.Suchwunsch{
		Frage: "lightning", Grenze: 10,
	})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusHybrid {
		t.Fatalf("der Modus ist %q", ergebnis.Modus)
	}

	if len(ergebnis.Treffer) != 3 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}

	// **»beide« steht vorn**, obwohl es in jeder Liste nur Zweiter war.
	if ergebnis.Treffer[0].Dokument.Kennung != "beide" {
		t.Fatalf("die Ordnung ist %q, %q, %q",
			ergebnis.Treffer[0].Dokument.Kennung,
			ergebnis.Treffer[1].Dokument.Kennung,
			ergebnis.Treffer[2].Dokument.Kennung)
	}

	// **Und die Ränge zählen von eins durch** — sie sind die Ausgabe, nicht
	// die Eingabe.
	for stelle, treffer := range ergebnis.Treffer {
		if treffer.Rang != stelle+1 {
			t.Fatalf("der %d. Treffer trägt Rang %d", stelle+1, treffer.Rang)
		}
	}
}

// **Ein Filter greift vor der Grenze.**
//
// # Der Fall, für den das gebaut ist
//
// **Zwanzig Kartentexte und eine Regelstelle.** Wer nach Regelstellen filtert
// und die Grenze bei zehn liegt, bekäme ohne diese Zeile nichts: die Regel
// stünde auf Platz einundzwanzig, und das Top-k hätte sie weggeschnitten.
// Das Issue: „Filterfall darf nicht allein durch ungefiltertes Top-k seine
// Treffer verlieren."
func TestEinFilterGreiftVorDerGrenze(t *testing.T) {
	t.Parallel()

	viele := make([]domain.Retrievaltreffer, 0, 21)
	for i := range 20 {
		viele = append(viele, einRetrievaltreffer(
			"karte"+string(rune('a'+i)), i+1, domain.QuelleOracle))
	}

	viele = append(viele, einRetrievaltreffer("regel", 21, domain.QuelleRegelwerk))

	ablage := &retrievalablagedoppel{manifest: einRetrievalmanifest(), lexikalisch: viele}

	ergebnis, fehler := einRetrievaldienst(ablage, nil).Suchen(t.Context(),
		service.Suchwunsch{
			Frage:  "lightning",
			Grenze: 10,
			Filter: service.Retrievalfilter{Quellen: []domain.Retrievalquelle{
				domain.QuelleRegelwerk,
			}},
		})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if len(ergebnis.Treffer) != 1 ||
		ergebnis.Treffer[0].Dokument.Kennung != "regel" {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}

	// **Gefragt wurde weiter** als die Grenze — sonst hätte die Ablage die
	// Regelstelle nie herausgegeben.
	for _, grenze := range ablage.gefragt {
		if grenze <= 10 {
			t.Fatalf("gefragt wurde mit Grenze %d", grenze)
		}
	}
}

// **Und die Filter sind deterministisch — kein erfundener Treffer.**
func TestDieFilterSindDeterministisch(t *testing.T) {
	t.Parallel()

	treffer := []domain.Retrievaltreffer{
		einRetrievaltreffer("englisch", 1, domain.QuelleOracle),
		einRetrievaltreffer("deutsch", 2, domain.QuelleOracle),
	}
	treffer[1].Dokument.Sprache = "de"

	for name, fall := range map[string]struct {
		filter   service.Retrievalfilter
		erwartet []string
	}{
		"nach Sprache": {
			service.Retrievalfilter{Sprachen: []string{"de"}}, []string{"deutsch"},
		},
		"nach einer Sprache, die niemand hat": {
			service.Retrievalfilter{Sprachen: []string{"fr"}}, nil,
		},
		"nach Quelle": {
			service.Retrievalfilter{Quellen: []domain.Retrievalquelle{
				domain.QuelleRuling,
			}}, nil,
		},
		"ohne Filter": {
			service.Retrievalfilter{}, []string{"englisch", "deutsch"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
				manifest: einRetrievalmanifest(), lexikalisch: treffer,
			}, nil).Suchen(t.Context(), service.Suchwunsch{
				Frage: "text", Grenze: 10, Filter: fall.filter,
			})
			if fehler != nil {
				t.Fatalf("Suchen: %v", fehler)
			}

			if len(ergebnis.Treffer) != len(fall.erwartet) {
				t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
			}

			for stelle, kennung := range fall.erwartet {
				if ergebnis.Treffer[stelle].Dokument.Kennung != kennung {
					t.Fatalf("der %d. Treffer ist %q, erwartet %q", stelle+1,
						ergebnis.Treffer[stelle].Dokument.Kennung, kennung)
				}
			}
		})
	}
}

// **Ohne aktiven Index findet die Suche nichts — und sagt, woran es liegt.**
func TestOhneAktivenIndexSagtDieSucheEs(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{kein: true}, nil).Suchen(
		t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if len(ergebnis.Treffer) != 0 {
		t.Fatalf("die Treffer sind %+v", ergebnis.Treffer)
	}

	if ergebnis.Modus != service.ModusOhneIndex || ergebnis.Grund == "" {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Eine leere Frage wird abgewiesen, bevor irgendetwas gerechnet wird.**
func TestEineLeereSuchfrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	einbetter := &retrievaleinbetterdoppel{
		raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
	}

	for name, frage := range map[string]string{
		"leer":         "",
		"nur Leerraum": "   ",
		"zu lang":      strings.Repeat("x", service.FrageHoechstens+1),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := einRetrievaldienst(&retrievalablagedoppel{
				manifest: einRetrievalmanifest(),
			}, einbetter).Suchen(t.Context(), service.Suchwunsch{
				Frage: frage, Grenze: 10,
			}); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}
		})
	}

	if len(einbetter.gesehen) != 0 {
		t.Errorf("es wurde trotzdem eingebettet: %+v", einbetter.gesehen)
	}
}

// **Eine Suche nach null Treffern beantwortet sich von selbst.**
func TestEineSucheOhneGrenzeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, grenze := range map[string]int{"null": 0, "negativ": -3} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := einRetrievaldienst(&retrievalablagedoppel{
				manifest: einRetrievalmanifest(),
			}, nil).Suchen(t.Context(), service.Suchwunsch{
				Frage: "lightning", Grenze: grenze,
			}); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("Grenze %d kam durch: %v", grenze, fehler)
			}
		})
	}
}
