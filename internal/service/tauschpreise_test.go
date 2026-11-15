package service_test

// **Was ein Tausch wiegt** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass nur die gewählte Reihe zählt.** Ein Stand aus einer anderen Quelle,
// einem anderen Markt oder der anderen Richtung ist kein Preis für diesen
// Tausch — er ist ein Preis für eine andere Frage (A3).
//
// **Dass ein Foil nicht den Preis des gewöhnlichen Stücks bekommt.**
//
// **Und die Fehlerwege**: wer einen Bestand nicht lesen kann, bekommt keinen
// halb gerechneten Abgleich.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// preisverzeichnisdoppel gibt die Stände, die der Abgleich filtern soll.
type preisverzeichnisdoppel struct {
	staende []domain.Preisstand
	fehler  error
}

func (p preisverzeichnisdoppel) NeuesteAlle(_ context.Context) ([]domain.Preisstand, error) {
	return p.staende, p.fehler
}

// einStand baut eine Preisbeobachtung der gewählten Reihe — für Shivan.
func einStand(wert int64, anders func(*domain.Preisstand)) domain.Preisstand {
	stand := domain.Preisstand{
		Printing: shivanDruck, Quelle: domain.PreisquelleScryfall,
		Markt: domain.MarktPapier, Art: domain.PreisartVerkauf,
		Waehrung: "EUR", Finish: domain.FinishNonfoil, Wert: wert,
	}

	if anders != nil {
		anders(&stand)
	}

	return stand
}

// dieReihe ist die ausdrücklich gewählte.
func dieReihe() domain.Preisstand {
	return domain.Preisstand{
		Quelle: domain.PreisquelleScryfall, Markt: domain.MarktPapier,
		Art: domain.PreisartVerkauf, Waehrung: "EUR",
	}
}

// einBepreistesWerk stellt einen Abgleich mit Preisen auf.
func einBepreistesWerk(staende []domain.Preisstand) *service.Tauschwerk {
	eigene := []*domain.Exemplar{
		meinExemplar(shivanDruck, 3, false),
		meinExemplar(boltDruck, 1, true),
	}

	drucke := []domain.Druck{
		einDruckMitKonzept(shivanDruck, "konzept-shivan"),
		einDruckMitKonzept(boltDruck, "konzept-bolt"),
	}

	return einTauschwerk(eigene, drucke).
		MitPreisreihe(preisverzeichnisdoppel{staende: staende}, dieReihe())
}

// einAbgleichMit fährt den Abgleich über beide Richtungen.
func einAbgleichMit(t *testing.T, werk *service.Tauschwerk) service.Tauschabgleich {
	t.Helper()

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung:  "bgegenueber",
		Angebot:  []domain.Sichtkarte{eineFremdzeile(boltDruck, 4)},
		Wuensche: []domain.Sichtkarte{eineFremdzeile(shivanDruck, 2)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	return abgleich
}

func TestNurDieGewaehlteReiheWiegt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		staende  []domain.Preisstand
		bepreist bool
		wert     int64
	}{
		"die gewählte Reihe": {
			[]domain.Preisstand{einStand(417, nil)}, true, 834,
		},
		"eine andere Quelle": {
			[]domain.Preisstand{einStand(417, func(s *domain.Preisstand) {
				s.Quelle = domain.Preisquelle("mtgjson")
			})}, false, 0,
		},
		"die andere Richtung": {
			[]domain.Preisstand{einStand(417, func(s *domain.Preisstand) {
				s.Art = domain.PreisartAnkauf
			})}, false, 0,
		},
		"ein anderer Markt": {
			[]domain.Preisstand{einStand(417, func(s *domain.Preisstand) {
				s.Markt = domain.MarktOnline
			})}, false, 0,
		},
		"eine andere Währung": {
			[]domain.Preisstand{einStand(417, func(s *domain.Preisstand) {
				s.Waehrung = "USD"
			})}, false, 0,
		},
		// **Ein Foil bekommt nicht den Preis des gewöhnlichen Stücks.**
		"nur das Foil ist bepreist": {
			[]domain.Preisstand{einStand(900, func(s *domain.Preisstand) {
				s.Finish = domain.FinishFoil
			})}, false, 0,
		},
		// **Ein Stand ohne Ausführung gilt als gewöhnlich** — dieselbe
		// Regel wie überall sonst (0119). Sie hier anders zu lesen hieße,
		// dass ein Preis, den eine Quelle nicht einordnet, gar keiner ist.
		"ein Stand ohne Ausführung": {
			[]domain.Preisstand{einStand(417, func(s *domain.Preisstand) {
				s.Finish = ""
			})}, true, 834,
		},
		"gar keine Stände": {nil, false, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			abgleich := einAbgleichMit(t, einBepreistesWerk(fall.staende))

			if !abgleich.HatVorschlag {
				t.Fatal("es kam kein Vorschlag")
			}

			if abgleich.Vorschlag.Gebe.Bepreist != fall.bepreist {
				t.Fatalf("bepreist ist %v", abgleich.Vorschlag.Gebe.Bepreist)
			}

			if abgleich.Vorschlag.Gebe.Wert != fall.wert {
				t.Errorf("der Wert ist %d", abgleich.Vorschlag.Gebe.Wert)
			}
		})
	}
}

// **Eine Reihe ohne Angaben nimmt jeden Stand** — wer nur die Quelle nennt,
// bekommt jede ihrer Reihen. Eine Prüfung gegen die Leere wiese jeden Stand ab.
func TestEineOffeneReiheNimmtJedenStand(t *testing.T) {
	t.Parallel()

	werk := einTauschwerk(
		[]*domain.Exemplar{
			meinExemplar(shivanDruck, 3, false),
			meinExemplar(boltDruck, 1, true),
		},
		[]domain.Druck{
			einDruckMitKonzept(shivanDruck, "konzept-shivan"),
			einDruckMitKonzept(boltDruck, "konzept-bolt"),
		},
	).MitPreisreihe(preisverzeichnisdoppel{
		staende: []domain.Preisstand{einStand(417, nil)},
	}, domain.Preisstand{})

	if abgleich := einAbgleichMit(t, werk); !abgleich.Vorschlag.Gebe.Bepreist {
		t.Error("die offene Reihe hat nichts genommen")
	}
}

// **Ein Fehler der Preisablage kommt durch.**
func TestEinFehlerDerPreisablageKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Preise sind zu")

	werk := einTauschwerk(nil, nil).
		MitPreisreihe(preisverzeichnisdoppel{fehler: gestolpert}, dieReihe())

	if _, fehler := werk.Abgleichen(context.Background(),
		service.Fremdstand{Kennung: "b"}, 1, domain.Tauschumfang{}); !errors.Is(fehler, gestolpert) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Fehler der Sammlung kommt durch** — und zwar aus beiden Sichten.
func TestEinFehlerDerSammlungKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Sammlung ist zu")

	for name, ab := range map[string]int{"aus den Dubletten": 1, "aus den Wünschen": 2} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werk := service.NeuesTauschwerk(
				service.NeueSammlungssichten(&stolperndeSammlung{abLauf: ab, fehler: gestolpert}),
				druckverzeichnisdoppel{},
			)

			if _, fehler := werk.Abgleichen(context.Background(),
				service.Fremdstand{Kennung: "b"}, 1, domain.Tauschumfang{}); !errors.Is(fehler, gestolpert) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Abbruch mitten in der Rechnung bricht ab** — der Kontext wird nach
// dem Lesen noch einmal gefragt, weil zwischen Lesen und Rechnen Zeit liegt.
func TestEinAbbruchNachDemLesenBrichtAb(t *testing.T) {
	t.Parallel()

	beendet, abbrechen := context.WithCancel(context.Background())

	werk := service.NeuesTauschwerk(
		service.NeueSammlungssichten(&sammlungsdoppel{}),
		druckverzeichnisdoppel{abbrechen: abbrechen},
	)

	if _, fehler := werk.Abgleichen(beendet,
		service.Fremdstand{Kennung: "b"}, 1, domain.Tauschumfang{}); !errors.Is(fehler, context.Canceled) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Druck ohne Konzeptkennung steht nicht im Verzeichnis** — und ein
// druckunabhängiger Wunsch auf ihn wird zum Befund statt zur Vermutung.
func TestEinDruckOhneKonzeptStehtNichtImVerzeichnis(t *testing.T) {
	t.Parallel()

	ohneKonzept := domain.Druck{
		ID: 7, Name: "Shivan Dragon",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck,
			Wert: shivanDruck,
		}},
	}

	wunsch := meinExemplar(shivanDruck, 1, true)
	wunsch.Wunschart = domain.WunschJederDruck

	werk := einTauschwerk([]*domain.Exemplar{wunsch}, []domain.Druck{ohneKonzept})

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung: "b", Angebot: []domain.Sichtkarte{eineFremdzeile(shivanNeu, 1)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(abgleich.Bekomme) != 0 {
		t.Errorf("es kam ein geratener Treffer: %+v", abgleich.Bekomme)
	}

	if len(abgleich.Befunde) == 0 {
		t.Error("es kam kein Befund")
	}
}

// **Eine lokale Druckzeile steht auch ohne Alias im Verzeichnis** — sonst
// fiele jeder vorläufige Druck aus dem druckunabhängigen Abgleich (A2).
func TestEineLokaleDruckzeileStehtImVerzeichnis(t *testing.T) {
	t.Parallel()

	lokal := domain.Druck{
		ID: 7, Name: "Shivan Dragon",
		Kennungen: []domain.Druckkennung{{
			System: domain.SystemScryfall, Semantik: domain.SemantikKonzept,
			Wert: "konzept-shivan",
		}},
	}

	wunsch := &domain.Exemplar{
		ID: 1, DruckID: 7, Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Wunsch: true,
		Wunschart: domain.WunschJederDruck, Name: "Shivan Dragon",
	}

	werk := einTauschwerk([]*domain.Exemplar{wunsch},
		[]domain.Druck{lokal, einDruckMitKonzept(shivanNeu, "konzept-shivan")})

	abgleich, fehler := werk.Abgleichen(context.Background(), service.Fremdstand{
		Kennung: "b", Angebot: []domain.Sichtkarte{eineFremdzeile(shivanNeu, 1)},
	}, 1, domain.Tauschumfang{})
	if fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if len(abgleich.Bekomme) != 1 {
		t.Errorf("es kamen %d Treffer, Befunde: %v", len(abgleich.Bekomme), abgleich.Befunde)
	}
}
