package service_test

// **Die Verwaltung** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Der Satz, für den es sie gibt**: gesucht wird mit dem Modell, mit dem
// gebaut wurde — nicht mit dem, das gerade eingerichtet ist.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einbetterquellendoppel gibt Einbetter zu Namen.
type einbetterquellendoppel struct {
	nach    map[string]service.Einbetter
	befunde []domain.Befund
	fehler  error

	// gefragt zählt die Rufe — für die Frage, ob gemerkt wird.
	gefragt int
}

func (e *einbetterquellendoppel) Einbetter(_ context.Context, modell string) (
	service.Einbetter, []domain.Befund, error,
) {
	e.gefragt++

	if e.fehler != nil {
		return nil, nil, e.fehler
	}

	if len(e.befunde) > 0 {
		return nil, e.befunde, nil
	}

	einbetter, bekannt := e.nach[modell]
	if !bekannt {
		return nil, []domain.Befund{{
			Datei: modell, Art: domain.ArtUmgebung,
			Meldung: modell + " ist nicht installiert",
		}}, nil
	}

	return einbetter, nil, nil
}

// korpusoeffnerdoppel gibt einen Korpus und zählt das Schließen.
type korpusoeffnerdoppel struct {
	fehler      error
	geschlossen int
}

func (k *korpusoeffnerdoppel) Oeffnen(_ context.Context, _ service.Aufbauauftrag) (
	service.Korpusquelle, func(), error,
) {
	if k.fehler != nil {
		return nil, nil, k.fehler
	}

	return einKorpus(), func() { k.geschlossen++ }, nil
}

// eineVerwaltung baut sie über den Doppeln.
func eineVerwaltung(
	ablage *retrievalablagedoppel, schreiber *schreiberdoppel,
	quelle *einbetterquellendoppel, oeffner *korpusoeffnerdoppel,
) *service.Retrievalverwaltung {
	return service.NeueRetrievalverwaltung(ablage, schreiber, quelle, oeffner,
		service.NeuerHaushalt(domain.Vorgabegrenzen()))
}

// dieQuelleMit gibt eine Quelle, die genau dieses Modell kennt.
func dieQuelleMit(modell string) *einbetterquellendoppel {
	return &einbetterquellendoppel{nach: map[string]service.Einbetter{
		modell: &retrievaleinbetterdoppel{
			raum: einRetrievalmanifest().Raum(), vektor: []float32{1, 0, 0},
		},
	}}
}

// **Gesucht wird mit dem Modell des Index.**
func TestGesuchtWirdMitDemModellDesIndex(t *testing.T) {
	t.Parallel()

	ablage := &retrievalablagedoppel{
		manifest:    einRetrievalmanifest(),
		lexikalisch: []domain.Retrievaltreffer{einRetrievaltreffer("eines", 1, domain.QuelleOracle)},
		semantisch:  []domain.Retrievaltreffer{einRetrievaltreffer("zweites", 1, domain.QuelleOracle)},
	}

	quelle := dieQuelleMit(einRetrievalmanifest().Modell)

	verwaltung := eineVerwaltung(ablage, &schreiberdoppel{}, quelle,
		&korpusoeffnerdoppel{})

	ergebnis, fehler := verwaltung.Suchen(t.Context(), service.Suchwunsch{
		Frage: "lightning", Grenze: 10,
	})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusHybrid {
		t.Fatalf("der Modus ist %q (%s)", ergebnis.Modus, ergebnis.Grund)
	}

	// **Und der Einbetter wird gemerkt** — ihn zu bauen ist ein Lauf am
	// echten Modell.
	if _, fehler := verwaltung.Suchen(t.Context(), service.Suchwunsch{
		Frage: "shock", Grenze: 10,
	}); fehler != nil {
		t.Fatalf("die zweite Suche: %v", fehler)
	}

	if quelle.gefragt != 1 {
		t.Errorf("das Modell wurde %dmal vermessen", quelle.gefragt)
	}
}

// **Fehlt das Modell des Index, bleibt die Suche lexikalisch und sagt es.**
//
// **Das ist der Modellwechsel aus dem Issue**: „Bei Modellwechsel alten Index
// nur mit seinem kompatiblen Query-Embeddingweg benutzen; sonst lexikalischer
// Modus mit sichtbarem Grund."
func TestOhneDasModellDesIndexBleibtDieSucheLexikalisch(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]*einbetterquellendoppel{
		"es ist ein anderes installiert": dieQuelleMit("ein-anderes"),
		"die Laufzeit antwortet nicht": {
			fehler: domain.MitGrund(domain.ErrConflict, nil, "niemand da"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// **Je Lauf eine eigene Ablage.** Eine geteilte zählte die
			// Grenzen aus zwei Goroutinen in dieselbe Liste.
			ablage := &retrievalablagedoppel{
				manifest: einRetrievalmanifest(),
				lexikalisch: []domain.Retrievaltreffer{
					einRetrievaltreffer("eines", 1, domain.QuelleOracle),
				},
				semantisch: []domain.Retrievaltreffer{
					einRetrievaltreffer("zweites", 1, domain.QuelleOracle),
				},
			}

			ergebnis, fehler := eineVerwaltung(ablage, &schreiberdoppel{},
				quelle, &korpusoeffnerdoppel{}).Suchen(t.Context(),
				service.Suchwunsch{Frage: "lightning", Grenze: 10})
			if fehler != nil {
				t.Fatalf("Suchen: %v", fehler)
			}

			if ergebnis.Modus != service.ModusLexikalisch ||
				ergebnis.Grund == "" {
				t.Fatalf("das Ergebnis ist %+v", ergebnis)
			}

			if len(ergebnis.Treffer) != 1 {
				t.Errorf("die Treffer sind %+v", ergebnis.Treffer)
			}
		})
	}
}

// **Der Stand sagt, ob semantisch gesucht werden kann — und warum nicht.**
func TestDerStandSagtObSemantischGesuchtWerdenKann(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		ablage  *retrievalablagedoppel
		quelle  *einbetterquellendoppel
		aktiv   bool
		suchbar bool
	}{
		"alles da": {
			&retrievalablagedoppel{manifest: einRetrievalmanifest()},
			dieQuelleMit(einRetrievalmanifest().Modell), true, true,
		},
		"kein Index": {
			&retrievalablagedoppel{kein: true},
			dieQuelleMit("egal"), false, false,
		},
		"Index ohne sein Modell": {
			&retrievalablagedoppel{manifest: einRetrievalmanifest()},
			dieQuelleMit("ein-anderes"), true, false,
		},
		"Bestand unlesbar": {
			&retrievalablagedoppel{manifestfehler: domain.MitGrund(
				domain.ErrConflict, nil, "die Datei ist zu")},
			dieQuelleMit("egal"), false, false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand := eineVerwaltung(fall.ablage, &schreiberdoppel{},
				fall.quelle, &korpusoeffnerdoppel{}).Stand(t.Context())

			if stand.Aktiv != fall.aktiv || stand.Suchbar != fall.suchbar {
				t.Fatalf("der Stand ist %+v", stand)
			}

			if !stand.Suchbar && stand.Grund == "" {
				t.Error("der Stand nennt keinen Grund")
			}
		})
	}
}

// **Ein Aufbau schließt seinen Korpus — auch wenn er scheitert.**
func TestEinAufbauSchliesstSeinenKorpus(t *testing.T) {
	t.Parallel()

	oeffner := &korpusoeffnerdoppel{}

	verwaltung := eineVerwaltung(&retrievalablagedoppel{}, &schreiberdoppel{},
		dieQuelleMit("all-minilm"), oeffner)

	if _, fehler := verwaltung.Aufbauen(t.Context(), service.Aufbauauftrag{
		Modell: "all-minilm", Quelle: "oracle.jsonl.gz",
	}); fehler != nil {
		t.Fatalf("Aufbauen: %v", fehler)
	}

	if oeffner.geschlossen != 1 {
		t.Errorf("geschlossen wurde %dmal", oeffner.geschlossen)
	}
}

// **Ein Aufbau ohne taugliches Modell findet nicht statt.**
func TestEinAufbauOhneTauglichesModellFindetNichtStatt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		quelle  *einbetterquellendoppel
		oeffner *korpusoeffnerdoppel
	}{
		"das Modell fehlt": {
			dieQuelleMit("ein-anderes"), &korpusoeffnerdoppel{},
		},
		"die Laufzeit antwortet nicht": {
			&einbetterquellendoppel{fehler: domain.MitGrund(
				domain.ErrConflict, nil, "niemand da")}, &korpusoeffnerdoppel{},
		},
		"der Korpus geht nicht auf": {
			dieQuelleMit("all-minilm"), &korpusoeffnerdoppel{
				fehler: domain.MitGrund(domain.ErrNotFound, nil,
					"die Sammeldatei fehlt"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := eineVerwaltung(&retrievalablagedoppel{},
				&schreiberdoppel{}, fall.quelle, fall.oeffner).Aufbauen(
				t.Context(), service.Aufbauauftrag{Modell: "all-minilm"})
			if fehler == nil {
				t.Fatal("der Aufbau kam durch")
			}

			if fall.oeffner.geschlossen != 0 {
				t.Errorf("geschlossen wurde %dmal", fall.oeffner.geschlossen)
			}
		})
	}
}

// **Ein Abbruch ohne laufenden Aufbau sagt es.**
func TestEinAbbruchOhneAufbauSagtEs(t *testing.T) {
	t.Parallel()

	if eineVerwaltung(&retrievalablagedoppel{}, &schreiberdoppel{},
		dieQuelleMit("x"), &korpusoeffnerdoppel{}).Abbrechen() {
		t.Error("es wurde etwas abgebrochen, das nicht lief")
	}
}

// **Ein laufender Aufbau lässt sich abbrechen.**
func TestEinLaufenderAufbauLaesstSichAbbrechen(t *testing.T) {
	t.Parallel()

	laeuft := make(chan struct{})
	schreiber := &schreiberdoppel{}

	verwaltung := eineVerwaltung(&retrievalablagedoppel{}, schreiber,
		dieQuelleMit("all-minilm"), &korpusoeffnerdoppel{})

	abgebrochen := make(chan struct{})

	var einmal sync.Once

	// **Der Lauf hält an, bis wirklich abgebrochen wurde.** Ohne das wäre
	// der Ausgang ein Rennen: wer nach dem letzten Schreiben abbricht,
	// bricht nichts ab — und der Lauf wäre manchmal grün geworden, ohne den
	// Zweig je betreten zu haben.
	schreiber.vorSchreiben = func() {
		einmal.Do(func() { close(laeuft); <-abgebrochen })
	}

	fertig := make(chan error, 1)

	go func() {
		_, fehler := verwaltung.Aufbauen(context.WithoutCancel(t.Context()),
			service.Aufbauauftrag{Modell: "all-minilm"})
		fertig <- fehler
	}()

	<-laeuft

	if !verwaltung.Abbrechen() {
		t.Error("der laufende Aufbau ließ sich nicht abbrechen")
	}

	close(abgebrochen)

	if fehler := <-fertig; !errors.Is(fehler, context.Canceled) {
		t.Fatalf("der Aufbau: %v", fehler)
	}

	if stand := verwaltung.Stand(t.Context()); stand.Aufbau.Laeuft {
		t.Errorf("der Stand ist %+v", stand.Aufbau)
	}
}

// **Eine Suche ohne aktiven Index sagt es und rechnet nichts.**
func TestEineSucheOhneAktivenIndexSagtEs(t *testing.T) {
	t.Parallel()

	ergebnis, fehler := eineVerwaltung(&retrievalablagedoppel{kein: true},
		&schreiberdoppel{}, dieQuelleMit("x"), &korpusoeffnerdoppel{}).Suchen(
		t.Context(), service.Suchwunsch{Frage: "lightning", Grenze: 10})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	if ergebnis.Modus != service.ModusOhneIndex ||
		!strings.Contains(ergebnis.Grund, "kein Suchindex") {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Ein unlesbarer Bestand ist ein Fehler und kein leeres Ergebnis.**
func TestEinUnlesbarerBestandIstBeiDerSucheEinFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := eineVerwaltung(&retrievalablagedoppel{
		manifestfehler: domain.MitGrund(domain.ErrConflict, nil, "zu"),
	}, &schreiberdoppel{}, dieQuelleMit("x"), &korpusoeffnerdoppel{}).Suchen(
		t.Context(), service.Suchwunsch{Frage: "x", Grenze: 10},
	); fehler == nil {
		t.Fatal("der Bestandsfehler kam als Ergebnis zurück")
	}
}

// **Ein besetzter Haushalt hält auch die Suche auf.**
func TestEinBesetzterHaushaltHaeltAuchDieSucheAuf(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel = map[domain.Ressourcenklasse]int{domain.RessourceSuche: 1}

	haushalt := service.NeuerHaushalt(grenzen)

	freigeben, genommen := haushalt.Versuchen(domain.RessourceSuche)
	if !genommen {
		t.Fatal("der einzige Platz war schon weg")
	}

	defer freigeben()

	kontext, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := service.NeueRetrievalverwaltung(
		&retrievalablagedoppel{}, &schreiberdoppel{}, dieQuelleMit("x"),
		&korpusoeffnerdoppel{}, haushalt).Suchen(kontext,
		service.Suchwunsch{Frage: "x", Grenze: 10},
	); !errors.Is(fehler, context.Canceled) {
		t.Fatalf("Suchen: %v", fehler)
	}
}
