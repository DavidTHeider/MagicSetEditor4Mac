package app_test

// **Die Problemprüfung am Fenster** — AP-8, Teil 2
// (#2365),
// mit dem Vertrag aus PERF02 (#2352): **gelesen wird ohne zu prüfen**, und
// geprüft wird auf Anstoß, als Auftrag mit Fortschritt. Ein Set, das niemand
// prüfte, heißt ungeprüft — nicht „0 Probleme“.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// problemprobe gibt eine Prüfung, die je Set eine Karte mit einem Befund
// findet, und zählt, wie oft sie zeichnete.
func problemprobe(gezeichnet *int) *service.Problempruefung {
	return service.NeueProblempruefung(
		func(_ context.Context, setID int64) ([]*domain.CardRow, error) {
			return []*domain.CardRow{{ID: setID * 10, SetID: setID}}, nil
		},
		func(context.Context, *domain.CardRow) ([]string, error) {
			*gezeichnet++

			return []string{"Das Bild fehlt."}, nil
		},
		func(context.Context, int64) (string, error) { return "abdruck", nil },
		time.Now,
	)
}

func mitProblemen(gezeichnet *int, fehler error) *app.App {
	anwendung := mitLeser(leseprobe{
		koepfe: []domain.SetHead{{ID: 1, Name: "Eins"}, {ID: 2, Name: "Zwei"}},
		fehler: fehler,
	})
	anwendung.MitProblemweg(app.ProblemwegAus(problemprobe(gezeichnet)))

	return anwendung
}

func TestSetproblemeLiestJedenStandOhneZuPruefen(t *testing.T) {
	gezeichnet := 0
	antwort := mitProblemen(&gezeichnet, nil).Setprobleme(t.Context())

	if antwort.Fehler != nil || len(antwort.Staende) != 2 {
		t.Fatalf("%+v", antwort)
	}

	for _, stand := range antwort.Staende {
		if stand.Urteil != "ungeprueft" || stand.Sicher {
			t.Errorf("%s: %+v — ein ungeprüftes Set gilt als ungeprüft, nicht als sauber", stand.Name, stand)
		}
	}

	if antwort.Staende[0].SetID != "1" || antwort.Staende[1].Name != "Zwei" {
		t.Errorf("Kennung und Name: %+v", antwort.Staende)
	}

	if gezeichnet != 0 {
		t.Errorf("Lesen zeichnete %d Karten — es soll nicht prüfen", gezeichnet)
	}
}

func TestSetproblemePruefenPrueftJedesSet(t *testing.T) {
	gezeichnet := 0
	anwendung := mitProblemen(&gezeichnet, nil)

	antwort := anwendung.SetproblemePruefen(t.Context())

	if antwort.Fehler != nil || len(antwort.Staende) != 2 || gezeichnet != 2 {
		t.Fatalf("%d gezeichnet, %+v", gezeichnet, antwort)
	}

	for _, stand := range antwort.Staende {
		if stand.Urteil != "aktuell" || !stand.Sicher || stand.Zahl != 1 {
			t.Errorf("%s: %+v", stand.Name, stand)
		}
	}

	// **Danach liest das Lesen denselben Stand** — ohne neu zu zeichnen.
	gelesen := anwendung.Setprobleme(t.Context())
	if gelesen.Staende[1].Zahl != 1 || !gelesen.Staende[1].Sicher || gezeichnet != 2 {
		t.Errorf("%d gezeichnet, %+v", gezeichnet, gelesen.Staende)
	}
}

// **Ein Abbruch lässt den Stand unvollständig** — und kein Set gilt dann als
// geprüft, das nicht geprüft wurde.
func TestSetproblemePruefenNachAbbruch(t *testing.T) {
	gezeichnet := 0
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	antwort := mitProblemen(&gezeichnet, nil).SetproblemePruefen(ctx)

	if antwort.Fehler == nil {
		t.Fatal("ein abgebrochener Lauf meldet keinen Fehler")
	}

	for _, stand := range antwort.Staende {
		if stand.Sicher {
			t.Errorf("%s gilt nach dem Abbruch als sicher: %+v", stand.Name, stand)
		}
	}
}

func TestSetproblemeOhneWegOderLeser(t *testing.T) {
	ohneWeg := mitLeser(leseprobe{koepfe: []domain.SetHead{{ID: 1, Name: "Eins"}}})
	if antwort := ohneWeg.Setprobleme(t.Context()); antwort.Fehler != nil || len(antwort.Staende) != 0 {
		t.Errorf("ohne Weg: %+v", antwort)
	}

	if antwort := ohneWeg.SetproblemePruefen(t.Context()); antwort.Fehler != nil || len(antwort.Staende) != 0 {
		t.Errorf("ohne Weg geprüft: %+v", antwort)
	}

	gezeichnet := 0
	kaputt := mitProblemen(&gezeichnet, errors.New("die Tafel fehlt"))

	if antwort := kaputt.Setprobleme(t.Context()); antwort.Fehler == nil {
		t.Errorf("ein Lesefehler kommt nicht an: %+v", antwort)
	}

	if antwort := kaputt.SetproblemePruefen(t.Context()); antwort.Fehler == nil {
		t.Errorf("ein Lesefehler beim Prüfen kommt nicht an: %+v", antwort)
	}
}

// **Jedes Set steht in der Antwort, auch wenn sein Stand nicht zu lesen war**
// — und der erste Fehler sagt, warum.
func TestSetproblemeNennenJedesSetBeiLesefehlern(t *testing.T) {
	anwendung := mitLeser(leseprobe{koepfe: []domain.SetHead{{ID: 1, Name: "Eins"}, {ID: 2, Name: "Zwei"}}})
	anwendung.MitProblemweg(app.ProblemwegAus(service.NeueProblempruefung(
		func(context.Context, int64) ([]*domain.CardRow, error) { return nil, nil },
		func(context.Context, *domain.CardRow) ([]string, error) { return nil, nil },
		func(context.Context, int64) (string, error) { return "", domain.ErrNotFound },
		time.Now,
	)))

	antwort := anwendung.Setprobleme(t.Context())

	if antwort.Fehler == nil || len(antwort.Staende) != 2 {
		t.Fatalf("%+v", antwort)
	}

	if antwort.Staende[1].Name != "Zwei" || antwort.Staende[1].Sicher {
		t.Errorf("%+v", antwort.Staende[1])
	}
}
