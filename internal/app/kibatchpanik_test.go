package app_test

// **Ein Ziel, das zusammenbricht, ist ein gescheitertes Ziel** — BK-05b-KI
// (#2112).
//
// Ein Batch fährt seine Ziele in eigenen Goroutinen (`app.batchFahren`).
// Eine Panik dort fing niemand: der Prozess endete, mitsamt der Freigabe,
// die schon gebucht war, und den Zielen, die schon geantwortet hatten.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinZusammenbrechendesZielWirdAlsGescheitertGefuehrt(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)
	welt.karten.panikBei = 5

	antwort := welt.batch(t, dieBatchziele())
	if antwort.Fehler != nil {
		t.Fatalf("der Batch: %+v", antwort.Fehler)
	}

	if len(antwort.Ziele) != len(dieBatchziele()) {
		t.Fatalf("es kamen %d Lagen zurück: %+v", len(antwort.Ziele), antwort.Ziele)
	}

	for _, lage := range antwort.Ziele {
		gebrochen := lage.Zielkarte == 5

		switch {
		// **Unbekannt, mit dem Griff ins Protokoll** — dort steht der Grund
		// unter derselben Vorgangskennung (#1601).
		case gebrochen && (lage.Fehler == nil || lage.Fehler.Code != app.CodeUnbekannt ||
			lage.Fehler.Vorgang == ""):
			t.Errorf("das zusammengebrochene Ziel: %+v", lage.Fehler)
		case gebrochen && (len(lage.Felder) != 0 || lage.Kennung != ""):
			t.Errorf("das zusammengebrochene Ziel trägt einen Vorschlag: %+v", lage)
		case !gebrochen && lage.Fehler != nil:
			t.Errorf("%d kam nicht durch: %+v", lage.Zielkarte, lage.Fehler)
		}
	}

	// **Keine zweite Buchung**: der Batch bleibt **ein** Posten (0105).
	posten := welt.anwendung.Kostenposten(probekontext(t), derCloudanbieter,
		string(domain.ZeitraumTag))
	if posten.Fehler != nil || len(posten.Posten) != 1 {
		t.Errorf("Kostenposten: %+v", posten)
	}

	// **Und die Schranke ist wieder frei** — ein zweiter Batch ohne Panik
	// kommt ganz durch.
	welt.karten.panikBei = 0

	danach := welt.batch(t, dieBatchziele())
	for _, lage := range danach.Ziele {
		if lage.Fehler != nil {
			t.Errorf("danach: %d: %+v", lage.Zielkarte, lage.Fehler)
		}
	}

	if danach.Fehler != nil {
		t.Errorf("der Batch danach: %+v", danach.Fehler)
	}
}
