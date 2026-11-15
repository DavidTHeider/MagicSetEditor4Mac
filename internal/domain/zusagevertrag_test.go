package domain_test

// **Was ein Handel von diesem Exemplar nimmt** — #2492.
//
// Unter „Zusagen" ließ sich jede Besitzzeile des Drucks festhalten, auch
// eine in anderem Zustand. Verbucht ging dann das gewählte Exemplar weg, und
// drüben kam die abgesprochene Zeile an — die Belege passten nicht, und die
// Gegenzeichnung fiel still aus.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var glutNM = domain.Exemplar{
	Printing: "p-glut", Zustand: domain.ZustandNearMint, Sprache: "de", Finish: domain.FinishNonfoil,
}

func TestEinVertragNimmtNurSeineVariante(t *testing.T) {
	t.Parallel()

	vertrag := domain.Zusagevertrag{
		{Von: "anna", Printing: "p-glut", Menge: 2, Guete: "nearMint", Sprache: "de", Finish: "nonfoil"},
		{Von: "anna", Printing: "p-glut", Menge: 1, Guete: "excellent", Sprache: "de", Finish: "nonfoil"},
		{Von: "anna", Printing: "p-glut", Menge: 3, Guete: "nearMint", Sprache: "de", Finish: "nonfoil"},
	}

	excellent := glutNM
	excellent.Zustand = "excellent"

	englisch := glutNM
	englisch.Sprache = "en"

	// **Leer heißt nonfoil** ([domain.Exemplar.Ausfuehrung]) — auf beiden
	// Seiten, sonst passte ein Exemplar ohne Angabe zu keiner Zeile.
	ohneAusfuehrung := glutNM
	ohneAusfuehrung.Finish = ""

	eigenerDruck := domain.Exemplar{DruckID: 7, Zustand: domain.ZustandNearMint, Sprache: "de"}

	for _, fall := range []struct {
		name     string
		exemplar domain.Exemplar
		vertrag  domain.Zusagevertrag
		grenze   int
	}{
		{"zwei Zeilen derselben Variante zählen zusammen", glutNM, vertrag, 5},
		{"ein anderer Zustand ist eine andere Zeile", excellent, vertrag, 1},
		{"eine andere Sprache passt zu keiner", englisch, vertrag, 0},
		{"ohne Ausführung ist es nonfoil", ohneAusfuehrung, vertrag, 5},
		{
			"eine eigene Druckzeile über ihre Kennung", eigenerDruck,
			domain.Zusagevertrag{{DruckID: 7, Menge: 1, Guete: "nearMint", Sprache: "de"}},
			1,
		},
		{"ohne Vertrag nichts", glutNM, nil, 0},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if grenze := fall.vertrag.GrenzeFuer(fall.exemplar); grenze != fall.grenze {
				t.Errorf("Grenze %d, erwartet %d", grenze, fall.grenze)
			}
		})
	}
}

func TestGleicheVarianteIstDieDesVertrags(t *testing.T) {
	t.Parallel()

	ohneAusfuehrung := glutNM
	ohneAusfuehrung.Finish = ""

	anderer := glutNM
	anderer.Printing = "p-sturm"

	if !glutNM.GleicheVariante(ohneAusfuehrung) || glutNM.GleicheVariante(anderer) {
		t.Error("die Gleichheit der Variante weicht von der des Vertrags ab")
	}
}
