package app_test

// **„Das Umlegen wirkt sofort, ohne Neustart"** — die Zusage aus
// [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
// Entscheidung 1, für die **drei neuen** Dienste.
//
// # Warum es diesen Test gibt
//
// `Suchquellen` liest die Schalter **einmal**, beim Verdrahten, und baut die
// Karte der Quellen daraus. Für Scryfall macht das nichts: sein Klient trägt
// `netz.Klient.MitErlaubnis`, und der fragt **vor jeder Anfrage**. Die drei
// neuen bekamen ihn nicht.
//
// Im Fenster wird `Suchnaht` genau einmal gerufen — `scryfallstart.go`, beim
// Start. Wer danach einen Dienst abschaltet, ändert damit die Einstellung und
// nicht die Karte, die schon gebaut ist.
//
// **Der Kommentar über `Suchquellen` behauptete das Gegenteil**: „Der
// Transport hält sie ohnehin". Für Scryfall stimmte er, für die anderen drei
// nicht — dieselbe Gestalt wie zweimal zuvor in dieser Phase: eine Zusage im
// Kopf, die der Rumpf nicht einlöst (§10).

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schaltbareWahl lässt sich **nach** dem Verdrahten umlegen.
//
// **Das ist der ganze Unterschied zu `dienstwahlprobe`**: die feste Karte dort
// kann die Frage gar nicht stellen, weil sie sich nie ändert.
type schaltbareWahl struct{ an map[domain.Dienst]bool }

func (s *schaltbareWahl) IstAn(_ context.Context, dienst domain.Dienst) (
	bool, error,
) {
	return s.an[dienst], nil
}

func TestEinAbgeschalteterDienstSchicktNichtsMehr(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			if _, err := schreiber.Write([]byte(`{"cards":[{
				"id":"aaaa","name":"Blitzschlag","set":"PRB","setName":"Probeset",
				"manaCost":"{R}","type":"Instant","rarity":"Common","number":"141"
			}]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	// **Beides zusammen** (§3): kein Zuhause dieses Rechners und eine
	// umgelenkte Adresse. Einzeln reicht keines von beiden.
	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	wahl := &schaltbareWahl{an: map[domain.Dienst]bool{domain.DienstMTGAPI: true}}

	quelle := app.Suchnaht(wahl, bundprobe{},
		fragewahlprobe{dienst: domain.DienstMTGAPI})

	// **Erst der Beleg, dass jemand zuhört** (§3): ohne ihn sagte die
	// Verneinung darunter nur, dass gar nichts verdrahtet war.
	if _, suchfehler := quelle.Suche(probekontext(t), "Blitz"); suchfehler != nil {
		t.Fatalf("die erste Suche: %v", suchfehler)
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Fatalf("der Dienst sah %d Anfragen, erwartet genau eine", gezaehlt)
	}

	// **Und jetzt aus** — ohne Neustart, ohne neue Naht.
	wahl.an[domain.DienstMTGAPI] = false

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if gezaehlt := anfragen.Load(); gezaehlt != 1 {
		t.Errorf("der abgeschaltete Dienst sah %d Anfragen statt keiner — "+
			"»aus« heißt: kein einziger Aufruf geht hinaus", gezaehlt-1)
	}

	if suchfehler == nil {
		t.Fatal("die Suche über einen abgeschalteten Dienst kam durch")
	}

	// **Und die Meldung nennt den Dienst.** „Nichts gefunden" wäre die
	// falsche Auskunft: gefunden wurde nichts, weil nicht gefragt wurde (§8).
	if !domain.IstDienstAus(suchfehler) {
		t.Errorf("die Meldung ist %v — erwartet wird ErrDienstAus", suchfehler)
	}

	if !strings.Contains(suchfehler.Error(), string(domain.DienstMTGAPI)) {
		t.Errorf("die Meldung %q nennt den abgeschalteten Dienst nicht",
			suchfehler)
	}
}

// **Und die andere Richtung: `aus` → `an` wirkt genauso sofort.**
//
// **Befund 1 des Zweitblicks über Phase 4.8.** Die erste Korrektur hängte die
// Sperre an den Transport und löste damit `an` → `aus`: ein gebauter Klient
// wird vor jeder Anfrage erneut gefragt. Für `aus` → `an` gibt es den Klienten
// gar nicht — `Suchquellen` nahm einen Dienst nur auf, wenn er **beim
// Verdrahten** eingeschaltet war, und das Fenster verdrahtet einmal, beim
// Start.
//
// Damit war der Fall behoben und die Klasse nicht (§10). Dasselbe galt für
// Zugangsdaten, die jemand nach dem Start einträgt: MTGJSON und Scrydex kamen
// ohne sie nicht in die Karte und blieben bis zum Neustart draußen.
func TestEinEingeschalteterDienstAntwortetOhneNeustart(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			if _, err := schreiber.Write([]byte(`{"cards":[{
				"id":"aaaa","name":"Blitzschlag","set":"PRB","setName":"Probeset",
				"manaCost":"{R}","type":"Instant","rarity":"Common","number":"141"
			}]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	// **Aus, beim Verdrahten.**
	wahl := &schaltbareWahl{an: map[domain.Dienst]bool{}}

	quelle := app.Suchnaht(wahl, bundprobe{},
		fragewahlprobe{dienst: domain.DienstMTGAPI})

	// **Der Beleg, dass jemand zuhört** (§3): abgeschaltet kommt nichts durch.
	if _, suchfehler := quelle.Suche(probekontext(t), "Blitz"); suchfehler == nil {
		t.Fatal("die Suche über einen abgeschalteten Dienst kam durch")
	}

	if gezaehlt := anfragen.Load(); gezaehlt != 0 {
		t.Fatalf("der Dienst sah %d Anfragen, erwartet keine", gezaehlt)
	}

	// **Und jetzt an** — ohne Neustart, ohne neue Naht.
	wahl.an[domain.DienstMTGAPI] = true

	auskunft, suchfehler := quelle.Suche(probekontext(t), "Blitz")
	if suchfehler != nil {
		t.Fatalf("nach dem Einschalten: %v", suchfehler)
	}

	if len(auskunft.Treffer) != 1 {
		t.Fatalf("%d Treffer, erwartet 1", len(auskunft.Treffer))
	}
}

// **Ein Fehler an einem Dienst reißt die anderen nicht mit** — Zusatzbefund 1
// des Nachlaufs über Phase 4.8
// (#1321).
//
// # Die Regression, die die Korrektur mitgebracht hat
//
// Seit die Quellenkarte **je Anfrage** entsteht, entsteht auch jeder
// Schlüsselbundzugriff je Anfrage. `Suchquellen` brach beim ersten Fehler ab
// und gab eine leere Karte zurück — damit scheiterte eine Suche über
// magicthegathering.io, weil Scrydex danebenstand und sein Schlüsselbund
// klemmte. Ein Dienst, den niemand gewählt hat, hielt den gewählten an.
//
// **Vorher fiel das nicht auf**, weil der Fehler einmal beim Verdrahten kam
// und die Naht gar nicht erst zustande brachte. Die Korrektur hat den Fehler
// von einem Startproblem zu einem Dauerproblem gemacht.
func TestEinKaputterBundEinesAnderenDienstesHaeltDieSucheNichtAn(t *testing.T) {
	var anfragen atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			anfragen.Add(1)

			if _, err := schreiber.Write([]byte(`{"cards":[{
				"id":"aaaa","name":"Blitzschlag","set":"PRB","setName":"Probeset",
				"manaCost":"{R}","type":"Instant","rarity":"Common","number":"141"
			}]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	// **Beide an, und der Schlüsselbund klemmt.** Gewählt ist der, der ohne
	// Zugang auskommt.
	quelle := app.Suchnaht(
		&schaltbareWahl{an: map[domain.Dienst]bool{
			domain.DienstMTGAPI:  true,
			domain.DienstScrydex: true,
		}},
		kaputterbund{},
		fragewahlprobe{dienst: domain.DienstMTGAPI})

	auskunft, suchfehler := quelle.Suche(probekontext(t), "Blitz")
	if suchfehler != nil {
		t.Fatalf("die Suche über magicthegathering.io: %v", suchfehler)
	}

	if len(auskunft.Treffer) != 1 {
		t.Fatalf("%d Treffer, erwartet 1", len(auskunft.Treffer))
	}
}

// **Und wer den kaputten Dienst wählt, erfährt, was ihm fehlt.**
//
// Der Grund reist in der Karte mit; ohne ihn sagte die Meldung „fehlt sein
// Zugang, oder ist er abgeschaltet?" — eine Frage statt einer Auskunft.
func TestWerDenKaputtenDienstWaehltErfaehrtDenGrund(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		&schaltbareWahl{an: map[domain.Dienst]bool{domain.DienstScrydex: true}},
		kaputterbund{},
		fragewahlprobe{dienst: domain.DienstScrydex})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if !errors.Is(suchfehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v, erwartet ErrConflict aus dem Bund", suchfehler)
	}
}

// **Ein fehlender Zugang ist ein eigener Grund**, kein Schweigen.
func TestEinFehlenderZugangNenntSichSelbst(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		&schaltbareWahl{an: map[domain.Dienst]bool{domain.DienstScrydex: true}},
		bundprobe{},
		fragewahlprobe{dienst: domain.DienstScrydex})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if suchfehler == nil {
		t.Fatal("die Suche glückte ohne Zugang")
	}

	if strings.Contains(suchfehler.Error(), "oder ist er abgeschaltet") {
		t.Errorf("die Meldung rät zwischen zwei Gründen: %v", suchfehler)
	}

	if !strings.Contains(suchfehler.Error(), "scrydex") {
		t.Errorf("die Meldung %q nennt den Dienst nicht", suchfehler)
	}
}

// **Und dasselbe für MTGJSON, dem die Marke fehlt.**
//
// Der Weg ist ein anderer — eine Angabe statt zweier —, und ein Zweig, den
// kein Test betritt, ist keiner (§3).
func TestEineFehlendeMarkeNenntSichSelbst(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		&schaltbareWahl{an: map[domain.Dienst]bool{domain.DienstMTGJSON: true}},
		bundprobe{},
		fragewahlprobe{dienst: domain.DienstMTGJSON})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if suchfehler == nil {
		t.Fatal("die Suche glückte ohne Marke")
	}

	for _, teil := range []string{"mtgjson", domain.ZugangMarke} {
		if !strings.Contains(suchfehler.Error(), teil) {
			t.Errorf("die Meldung %q nennt %q nicht", suchfehler, teil)
		}
	}
}
