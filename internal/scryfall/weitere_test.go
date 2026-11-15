package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// gegenstelleMit antwortet auf jeden Weg mit demselben Rumpf und zählt mit.
func gegenstelleMit(t *testing.T, rumpf string) (*httptest.Server, *[]time.Time) {
	t.Helper()

	var zeiten []time.Time

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			zeiten = append(zeiten, time.Now())

			if anfrage.Header.Get("User-Agent") == "" || anfrage.Header.Get("Accept") == "" {
				t.Error("die Pflichtkopfzeilen fehlen")
			}

			schreiben.Header().Set("Content-Type", "application/json")

			if _, fehler := schreiben.Write([]byte(rumpf)); fehler != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	return server, &zeiten
}

const rulingsrumpf = `{"object":"list","has_more":false,"data":[
	{"object":"ruling","oracle_id":"0000-1","source":"wotc",
	 "published_at":"2004-10-04","comment":"Der erste Hinweis."},
	{"object":"ruling","oracle_id":"0000-1","source":"scryfall",
	 "published_at":"2021-03-19","comment":"Der zweite."}
]}`

func TestKlientRulings(t *testing.T) {
	t.Parallel()

	t.Run("liest die Hinweise samt Quelle und Datum", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, rulingsrumpf)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		hinweise, _, befunde, fehler := klient.Rulings(t.Context(), "0000-1")
		if fehler != nil {
			t.Fatalf("Rulings: %v", fehler)
		}

		if len(befunde) != 0 {
			t.Errorf("Befunde = %v", befunde)
		}

		if len(hinweise) != 2 {
			t.Fatalf("%d Hinweise, erwartet 2", len(hinweise))
		}

		if hinweise[0].Quelle != scryfall.QuelleWotc ||
			hinweise[1].Quelle != scryfall.QuelleScryfall {
			t.Errorf("Quellen = %q / %q", hinweise[0].Quelle, hinweise[1].Quelle)
		}

		if hinweise[0].Text != "Der erste Hinweis." {
			t.Errorf("Text = %q", hinweise[0].Text)
		}

		// **Das Datum bleibt eine Zeichenkette**, wie Scryfall es schickt:
		// eine Umrechnung in eine Zeitzone wäre eine Entscheidung, die dieses
		// Paket nicht treffen kann (§3, „eine Eingabe, die der Test schon
		// aufgeräumt hat").
		if hinweise[0].Datum != "2004-10-04" {
			t.Errorf("Datum = %q", hinweise[0].Datum)
		}
	})

	t.Run("führt eine unbekannte Quelle als Befund", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"ruling","oracle_id":"x","source":"jemandanders",
			 "published_at":"2026-01-01","comment":"Neu."}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		hinweise, _, befunde, fehler := klient.Rulings(t.Context(), "x")
		if fehler != nil {
			t.Fatalf("Rulings: %v", fehler)
		}

		// **Nicht verschluckt und nicht abgewiesen** (§16, Driftkontrolle):
		// der Hinweis kommt, und der unbekannte Wert steht als Befund daneben.
		if len(hinweise) != 1 || hinweise[0].Text != "Neu." {
			t.Errorf("Hinweise = %+v", hinweise)
		}

		if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "jemandanders") {
			t.Errorf("Befunde = %+v", befunde)
		}
	})

	t.Run("weist ein fehlendes Pflichtfeld ab", func(t *testing.T) {
		t.Parallel()

		// **`comment` fehlt** — das ist der Text selbst. Ein Hinweis ohne ihn
		// ist keiner, und ihn leer durchzulassen hieße, eine leere Zeile als
		// Regelauskunft auszugeben.
		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"ruling","oracle_id":"x","source":"wotc","published_at":"2026-01-01"}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		if _, _, _, fehler := klient.Rulings(t.Context(), "x"); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}
	})

	t.Run("erträgt ein unbekanntes Zusatzfeld", func(t *testing.T) {
		t.Parallel()

		server, _ := gegenstelleMit(t, `{"object":"list","data":[
			{"object":"ruling","oracle_id":"x","source":"wotc","published_at":"2026-01-01",
			 "comment":"Da.","irgendwas_neues":42}]}`)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		hinweise, _, _, fehler := klient.Rulings(t.Context(), "x")
		if fehler != nil {
			t.Fatalf("das neue Feld hat gebrochen: %v", fehler)
		}

		if len(hinweise) != 1 {
			t.Fatalf("%d Hinweise", len(hinweise))
		}
	})
}

// **Die zweite Endpunktklasse hat ihre eigene Drossel** —
// Abnahmekriterium aus #888:
// „Drossel je Endpunktklasse; **8/s als Reserve** unter der Grenze von 10."
//
// # Warum sie eine eigene ist und nicht die der Kartenwege
//
// Weil die Grenzen verschieden sind: `/cards/named` und Geschwister dürfen
// **2/s**, alles Übrige **10/s** (nachgelesen am 2026-08-23). Eine gemeinsame
// Drossel müsste die strengere nehmen und bremste damit die Wege, die es nicht
// nötig haben — auf ein Fünftel.
//
// # Was dieser Test misst und was nicht
//
// Er misst **den Abstand zwischen zwei Anfragen an der Gegenstelle**, in
// echter Zeit. Er sagt damit nichts darüber, ob Scryfall die Zahl akzeptiert —
// nur, dass diese Anwendung sie einhält.
func TestDieZweiteKlasseHatIhreEigeneDrossel(t *testing.T) {
	t.Parallel()

	server, zeiten := gegenstelleMit(t, rulingsrumpf)

	// **Eine echte Uhr, kein fester Zeitpunkt.** Mit `func() time.Time` auf
	// einen Wert gäbe die Drossel jede Wartezeit sofort frei, und der Test
	// wäre grün, ohne etwas gemessen zu haben.
	klient := scryfall.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "prüfung")

	for i := range 3 {
		if _, _, _, fehler := klient.Rulings(t.Context(), "karte-"+string(rune('a'+i))); fehler != nil {
			t.Fatalf("Rulings: %v", fehler)
		}
	}

	if len(*zeiten) != 3 {
		t.Fatalf("%d Anfragen, erwartet 3", len(*zeiten))
	}

	for i := 1; i < len(*zeiten); i++ {
		abstand := (*zeiten)[i].Sub((*zeiten)[i-1])

		if abstand < 100*time.Millisecond {
			t.Errorf("Abstand %d = %v — die Drossel greift nicht", i, abstand)
		}

		// **Und eine Obergrenze**, denn ohne sie prüft dieser Test nur, dass
		// **irgendeine** Drossel greift. Die erste Gegenprobe — beide Drosseln
		// auf die strengere Zahl gelegt — blieb damit grün, und die Zusage
		// „je Endpunktklasse eine" war unbelegt.
		//
		// **400 ms liegen zwischen beiden Zahlen**: erwartet sind 125,
		// zusammengelegt wären es 600. Die Spanne ist weit genug für eine
		// belastete Maschine und eng genug, um den Fall zu treffen.
		if abstand > 400*time.Millisecond {
			t.Errorf("Abstand %d = %v — hier greift die Kartendrossel, "+
				"nicht die allgemeine", i, abstand)
		}
	}
}

// **Und die Kartenwege bleiben bei ihrer strengeren Zahl.**
//
// Der Test daneben belegte sonst nur, dass **eine** Drossel greift — nicht,
// dass es zwei sind. Wer beide zusammenlegte, machte diesen hier rot.
func TestDieKartenwegeBleibenLangsamer(t *testing.T) {
	t.Parallel()

	server, zeiten := gegenstelleMit(t,
		`{"object":"card","id":"a","name":"Prüfstein"}`)

	klient := scryfall.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "prüfung")

	for _, name := range []string{"Eins", "Zwei"} {
		if _, _, _, fehler := klient.KarteNachName(t.Context(), name); fehler != nil {
			t.Fatalf("KarteNachName: %v", fehler)
		}
	}

	if len(*zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(*zeiten))
	}

	if abstand := (*zeiten)[1].Sub((*zeiten)[0]); abstand < 500*time.Millisecond {
		t.Errorf("Abstand = %v — die Kartendrossel ist zu schnell geworden", abstand)
	}
}
