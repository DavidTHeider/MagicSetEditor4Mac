package app

// **Der Speicher von Anwendung und Webprozess, getrennt** — OPS02
// (#2348), Abnahmekriterium 2: nicht der Go-Heap als Gesamtspeicher, und
// was fehlt, steht da.

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/prozessspeicher"
)

func TestDerSpeicherStehtJeProzessDa(t *testing.T) {
	t.Parallel()

	gemessen := func(int) (prozessspeicher.Stand, error) {
		return prozessspeicher.Stand{
			Anwendung: prozessspeicher.Prozess{PID: 1, Art: prozessspeicher.ArtAnwendung, FussabdruckBytes: 400 << 20},
			Web: []prozessspeicher.Prozess{
				{PID: 2, Art: prozessspeicher.ArtWebinhalt, FussabdruckBytes: 300 << 20},
				{PID: 3, Art: prozessspeicher.ArtGrafik, FussabdruckBytes: 80 << 20},
			},
		}, nil
	}

	var antwort MetrikenAntwort

	speicherAnhaengen(&antwort, gemessen)

	erwartet := []WebprozessDTO{{Art: "webinhalt", FussabdruckBytes: 300 << 20}, {Art: "gpu", FussabdruckBytes: 80 << 20}}
	if antwort.Prozess.FussabdruckBytes != 400<<20 || !slices.Equal(antwort.Prozess.Web, erwartet) || antwort.Prozess.WebUngemessen != "" {
		t.Errorf("Prozess %+v", antwort.Prozess)
	}

	ohneWeb := func(int) (prozessspeicher.Stand, error) {
		return prozessspeicher.Stand{Anwendung: prozessspeicher.Prozess{FussabdruckBytes: 1}, Ungemessen: "keine Zuordnung"}, nil
	}

	var zweite MetrikenAntwort

	speicherAnhaengen(&zweite, ohneWeb)

	if zweite.Prozess.WebUngemessen != "keine Zuordnung" || zweite.Prozess.Web == nil || len(zweite.Prozess.Web) != 0 {
		t.Errorf("ohne Zuordnung %+v", zweite.Prozess)
	}

	var dritte MetrikenAntwort

	speicherAnhaengen(&dritte, func(int) (prozessspeicher.Stand, error) { return prozessspeicher.Stand{}, errors.New("nichts") })

	if !slices.Contains(dritte.Ungemessen, UngemessenSpeicher) || dritte.Prozess.FussabdruckBytes != 0 || dritte.Prozess.WebUngemessen == "" {
		t.Errorf("ungemessen %+v, %v", dritte.Prozess, dritte.Ungemessen)
	}
}

// **Höchstens eine Messung je Frist** — ein Klient, der die Metriken in
// schneller Folge fragt, ließe sonst je Ruf jeden Prozess des Rechners
// aufzählen. Gefunden, als `TestDieAusfuehrlicheStufeKostetHoechstensFuenfMillisekunden`
// unter dem Race-Detektor tausend Rufe maß und die Messung den Abstand
// überdeckte, um den es ging.
func TestDerSpeicherWirdHoechstensEinmalJeFristGemessen(t *testing.T) {
	t.Parallel()

	gemessen := 0
	jetzt := time.Unix(1000, 0)
	messung := &gemerkteMessung{
		frist: time.Second,
		uhr:   func() time.Time { return jetzt },
		messen: func(int) (prozessspeicher.Stand, error) {
			gemessen++

			return prozessspeicher.Stand{Anwendung: prozessspeicher.Prozess{FussabdruckBytes: uint64(gemessen)}}, nil
		},
	}

	erster, err := messung.messe(1)
	zweiter, zweiterFehler := messung.messe(1)

	jetzt = jetzt.Add(time.Second)
	dritter, dritterFehler := messung.messe(1)

	if err != nil || zweiterFehler != nil || dritterFehler != nil {
		t.Fatal(err, zweiterFehler, dritterFehler)
	}

	if gemessen != 2 || erster.Anwendung.FussabdruckBytes != 1 || zweiter.Anwendung.FussabdruckBytes != 1 || dritter.Anwendung.FussabdruckBytes != 2 {
		t.Errorf("gemessen %d-mal: %d, %d, %d", gemessen, erster.Anwendung.FussabdruckBytes,
			zweiter.Anwendung.FussabdruckBytes, dritter.Anwendung.FussabdruckBytes)
	}
}
