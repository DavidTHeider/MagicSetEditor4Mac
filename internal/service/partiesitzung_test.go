package service_test

// **Eine Partie über eine Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass die vier Nachrichtenarten getrennt bleiben** — bis dorthin, wo sie
// die Leitung verlassen. Ein Ereignis wird nummeriert und gekettet, ein
// Mauszeiger nicht; wer beides über denselben Weg schickte, ohne es zu
// unterscheiden, hätte am Ende einen Mauszeiger in der Hashkette.
//
// **Und dass ein doppelt zugestelltes Ereignis nichts anrichtet.** Das Issue
// nennt den Schaden: „erzeugen keinen doppelten Draw".

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nahtdoppel merkt sich, was gesendet wurde.
type nahtdoppel struct {
	schloss  sync.Mutex
	gesendet [][]byte
	arten    []service.Zustellung
	klage    error
}

func (n *nahtdoppel) Sende(
	_ context.Context, nachricht []byte, art service.Zustellung,
) error {
	n.schloss.Lock()
	defer n.schloss.Unlock()

	if n.klage != nil {
		return n.klage
	}

	n.gesendet = append(n.gesendet, nachricht)
	n.arten = append(n.arten, art)

	return nil
}

func (n *nahtdoppel) letzte(t *testing.T) domain.Partienachricht {
	t.Helper()

	n.schloss.Lock()
	defer n.schloss.Unlock()

	if len(n.gesendet) == 0 {
		t.Fatal("es wurde nichts gesendet")
	}

	var nachricht domain.Partienachricht
	if fehler := json.Unmarshal(n.gesendet[len(n.gesendet)-1], &nachricht); fehler != nil {
		t.Fatalf("die letzte Nachricht ist kein JSON: %v", fehler)
	}

	return nachricht
}

const langeKette = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// gepackt macht aus einer Nachricht das, was über die Leitung geht.
func gepackt(t *testing.T, nachricht domain.Partienachricht) []byte {
	t.Helper()

	roh, fehler := json.Marshal(nachricht)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	return roh
}

// sendeOderScheitere schickt und bricht ab, wo es klemmt — die Läufe darunter
// prüfen, **was** gesendet wurde, nicht **ob**.
func sendeOderScheitere(
	t *testing.T, sitzung *service.Partiesitzung,
	ctx context.Context, art domain.Nachrichtenart, inhalt []byte,
) {
	t.Helper()

	if fehler := sitzung.Sende(ctx, art, inhalt); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}
}

func eineSitzung(naht service.Partienaht) *service.Partiesitzung {
	return service.NeuePartiesitzung("p1", "sitz-1", naht, service.NeueReihenfolge(1))
}

// **Ein Ereignis bekommt eine Nummer und ein Kettenglied.**
func TestEinEreignisWirdNummeriertUndGekettet(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := eineSitzung(naht)

	if fehler := sitzung.Sende(
		context.Background(), domain.ArtEreignis, []byte(`{"was":"ziehen"}`),
	); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}

	nachricht := naht.letzte(t)
	switch {
	case nachricht.Folge != 1:
		t.Errorf("Folge %d, erwartet 1", nachricht.Folge)
	case len(nachricht.Kette) != domain.Kettengliedlaenge:
		t.Errorf("Kettenglied %q", nachricht.Kette)
	case nachricht.Von != "sitz-1":
		t.Errorf("von %q, erwartet sitz-1", nachricht.Von)
	}
}

// **Die Kette bindet jedes Ereignis an das vorige.**
//
// Dieselbe Zusage wie in G1: wer ein Ereignis in der Mitte austauscht, ändert
// jedes Glied danach.
func TestJedesEreignisHaengtAmVorigen(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := eineSitzung(naht)

	sendeOderScheitere(t, sitzung, context.Background(), domain.ArtEreignis, []byte(`{"a":1}`))
	erstes := naht.letzte(t).Kette

	sendeOderScheitere(t, sitzung, context.Background(), domain.ArtEreignis, []byte(`{"a":1}`))
	zweites := naht.letzte(t).Kette

	if erstes == zweites {
		t.Error("zwei gleiche Inhalte gaben dasselbe Kettenglied — die Kette " +
			"bindet nicht ans vorige")
	}
}

// **Präsenz trägt weder Nummer noch Kette** — und geht **verwerfbar**.
//
// Die Begründung steht schon in `peerverbindung.go`: „ich sehe gerade Karte
// 12 an" ist in dem Moment wertlos, in dem „ich sehe Karte 13 an" vorliegt.
func TestPraesenzGehtFluechtigUndVerwerfbar(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := eineSitzung(naht)

	if fehler := sitzung.Sende(
		context.Background(), domain.ArtPraesenz, []byte(`{"x":1,"y":2}`),
	); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}

	nachricht := naht.letzte(t)
	switch {
	case nachricht.Folge != 0:
		t.Errorf("Folge %d, erwartet 0", nachricht.Folge)
	case nachricht.Kette != "":
		t.Errorf("Kettenglied %q, erwartet keines", nachricht.Kette)
	case naht.arten[0] != service.Verwerfbar:
		t.Errorf("Zustellung %v, erwartet verwerfbar", naht.arten[0])
	}
}

// **Chat und Kontrolle bekommen eine Nummer, aber kein Kettenglied.**
func TestChatUndKontrolleWerdenNummeriertAberNichtGekettet(t *testing.T) {
	t.Parallel()

	for _, art := range []domain.Nachrichtenart{domain.ArtChat, domain.ArtKontrolle} {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			naht := &nahtdoppel{}
			sitzung := eineSitzung(naht)

			if fehler := sitzung.Sende(
				context.Background(), art, []byte(`{"text":"hallo"}`),
			); fehler != nil {
				t.Fatalf("senden: %v", fehler)
			}

			nachricht := naht.letzte(t)
			switch {
			case nachricht.Folge == 0:
				t.Error("keine Folge")
			case nachricht.Kette != "":
				t.Errorf("Kettenglied %q, erwartet keines", nachricht.Kette)
			}
		})
	}
}

// **Eine Präsenz verbraucht keine Folgenummer.**
//
// Sonst entstünde bei jedem Mauszeiger eine Lücke im Strom der Ereignisse —
// und die blockiert beim Empfänger.
func TestPraesenzVerbrauchtKeineFolge(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := eineSitzung(naht)
	hintergrund := context.Background()

	sendeOderScheitere(t, sitzung, hintergrund, domain.ArtEreignis, []byte(`{"a":1}`))
	sendeOderScheitere(t, sitzung, hintergrund, domain.ArtPraesenz, []byte(`{"x":1}`))
	sendeOderScheitere(t, sitzung, hintergrund, domain.ArtEreignis, []byte(`{"a":2}`))

	if folge := naht.letzte(t).Folge; folge != 2 {
		t.Errorf("Folge %d, erwartet 2", folge)
	}
}

// **Was die Naht nicht annimmt, wird gemeldet.**
func TestEineKlagendeNahtWirdGemeldet(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Leitung ist zu")
	naht := &nahtdoppel{klage: kaputt}

	fehler := eineSitzung(naht).Sende(
		context.Background(), domain.ArtEreignis, []byte(`{}`))

	if !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine fremde Art wird gar nicht erst gesendet.**
func TestEineFremdeArtWirdNichtGesendet(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}

	fehler := eineSitzung(naht).Sende(
		context.Background(), "geraune", []byte(`{}`))

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Errorf("es kam %v", fehler)
	case len(naht.gesendet) != 0:
		t.Error("es wurde trotzdem gesendet")
	}
}

// **Die Kette ist ablesbar** — der Wirt braucht sie für seinen Stand, und
// ohne sie könnte niemand eine Wiederaufnahme beantworten.
func TestDieKetteDerSitzungIstAblesbar(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := eineSitzung(naht)

	if sitzung.Kette() != "" {
		t.Error("eine frische Sitzung hat schon ein Kettenglied")
	}

	sendeOderScheitere(t, sitzung, context.Background(), domain.ArtEreignis, []byte(`{"a":1}`))

	if sitzung.Kette() != naht.letzte(t).Kette {
		t.Error("die abgelesene Kette ist eine andere als die gesendete")
	}
}

// **Was sich nicht verpacken lässt, geht nicht hinaus.**
//
// `json.Marshal` scheitert an einer Nachricht nie — ihre Felder sind alle
// darstellbar. Der Zweig steht trotzdem da, und er wird hier betreten, indem
// der Inhalt kein gültiges JSON ist: `RawMessage` gibt ihn durch, und das
// Verpacken bricht.
func TestEinUnverpackbarerInhaltGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}

	fehler := eineSitzung(naht).Sende(
		context.Background(), domain.ArtEreignis, []byte("{kein json"))

	switch {
	case fehler == nil:
		t.Error("es kam durch")
	case len(naht.gesendet) != 0:
		t.Error("es wurde trotzdem gesendet")
	}
}

// **Und eine Nachricht, die ihre eigene Prüfung nicht besteht, auch nicht.**
//
// Sie entsteht, wenn ein Sitz keine Kennung hat — dann ist der Absender leer,
// und `domain.Partienachricht.Pruefe` weist sie ab, bevor die Leitung sie
// sieht (§8).
func TestEineUngueltigeNachrichtGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	ohneSitz := service.NeuePartiesitzung(
		"p1", "  ", naht, service.NeueReihenfolge(1))

	fehler := ohneSitz.Sende(
		context.Background(), domain.ArtEreignis, []byte(`{}`))

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Errorf("es kam %v", fehler)
	case len(naht.gesendet) != 0:
		t.Error("es wurde trotzdem gesendet")
	}
}

// **Zwei Absender teilen sich eine Sitzung, und die Reihenfolge hält.**
//
// # Was hier wirklich schiefgeht, wenn das Schloss fehlt
//
// **Nicht die Vergabe — die Auslieferung.** [Reihenfolge] hat ein eigenes
// Schloss, jede Nummer wird also genau einmal vergeben. Zwischen dem Vergeben
// und dem Absenden liegt aber ein Fenster: wer die Fünf schon hat und
// unterbrochen wird, während ein anderer die Sechs holt **und losschickt**,
// legt sie **nach** der Sechs auf die Leitung.
//
// **Und das ist kein Schönheitsfehler.** Beim Empfänger urteilt
// [domain.Empfaengerstand] über Folgenummern; die Sechs vor der Fünf ist eine
// **Lücke**, und eine Lücke lässt `Partienetz.Bediene` mit einem Fehler
// zurückkehren — die Leitung ist zu. Ein Satz im Gespräch reißt damit den
// Tisch ab.
//
// # Warum es das vorher nicht gab
//
// **Weil vorher nur einer schickte.** Die Schleife einer Gastsitzung war der
// einzige Absender. Mit dem Gespräch am Tisch schickt auch das Fenster — aus
// einer anderen Goroutine, über dieselbe Sitzung.
func TestZweiAbsenderTeilenSichEineSitzung(t *testing.T) {
	t.Parallel()

	naht := &nahtdoppel{}
	sitzung := service.NeuePartiesitzung("p1", "anna", naht,
		service.NeueReihenfolge(1))

	const je = 200

	var warte sync.WaitGroup

	for range 2 {
		warte.Add(1)

		go func() {
			defer warte.Done()

			for range je {
				if fehler := sitzung.Sende(t.Context(), domain.ArtChat,
					[]byte(`{"von":"anna","text":"hallo"}`)); fehler != nil {
					t.Errorf("senden: %v", fehler)

					return
				}
			}
		}()
	}

	warte.Wait()

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	if len(naht.gesendet) != 2*je {
		t.Fatalf("%d Nachrichten statt %d", len(naht.gesendet), 2*je)
	}

	vorige := uint64(0)

	for stelle, roh := range naht.gesendet {
		var nachricht domain.Partienachricht
		if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
			t.Fatalf("keine Nachricht: %v", fehler)
		}

		if nachricht.Folge <= vorige {
			t.Fatalf("an Stelle %d liegt die Folge %d hinter %d — der "+
				"Empfänger sähe eine Lücke und legte die Leitung nieder",
				stelle, nachricht.Folge, vorige)
		}

		vorige = nachricht.Folge
	}
}
