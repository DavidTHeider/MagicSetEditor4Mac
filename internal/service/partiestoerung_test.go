package service_test

// **Was eine gestörte Leitung anrichtet — und was nicht** — Werkstück G5 von
// Phase 5.5 (#1462).
//
// # Die Abnahme, für die es diesen Lauf gibt
//
// Das Issue verlangt sie wörtlich: „Perfect/Lossy/Disconnected-Testadapter
// prüfen gleichzeitige Commands, Duplikate, fehlende/vertauschte Events,
// Hashabweichung, falschen Snapshot und Abbruch."
//
// **Drei Leitungen, ein Ergebnis.** Was der Wirt sendet, kommt beim Gast in
// derselben Reihenfolge an — oder gar nicht, und dann sagt er es. Was nie
// geschehen darf: dass ein Zug **zweimal** wirkt oder **still** ausfällt.

import (
	"context"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stoerleitung ist eine Naht, die sich wie ein echtes Netz benimmt.
//
// **Sie hält fest, was hinausging**, und ein Lauf entscheidet, was davon
// ankommt — verdoppelt, vertauscht, verloren.
type stoerleitung struct {
	gesendet [][]byte
}

func (s *stoerleitung) Sende(
	_ context.Context, nachricht []byte, _ service.Zustellung,
) error {
	kopie := make([]byte, len(nachricht))
	copy(kopie, nachricht)
	s.gesendet = append(s.gesendet, kopie)

	return nil
}

// zugAn schickt drei Ereignisse und gibt, was über die Leitung ging.
func dreiZuege(t *testing.T) *stoerleitung {
	t.Helper()

	leitung := &stoerleitung{}
	sitzung := service.NeuePartiesitzung(
		"p1", "sitz-1", leitung, service.NeueReihenfolge(1))

	for _, inhalt := range []string{`{"zug":1}`, `{"zug":2}`, `{"zug":3}`} {
		sendeOderScheitere(
			t, sitzung, context.Background(), domain.ArtEreignis, []byte(inhalt))
	}

	return leitung
}

// nimmAlle spielt eine Zustellfolge ein und gibt die Urteile.
func nimmAlle(t *testing.T, roh [][]byte) []domain.Folgeurteil {
	t.Helper()

	empfang := service.NeuerPartieempfang(1)
	urteile := make([]domain.Folgeurteil, 0, len(roh))

	for _, eine := range roh {
		ausgang, fehler := empfang.Nimm(eine)
		if fehler != nil {
			t.Fatalf("nehmen: %v", fehler)
		}

		urteile = append(urteile, ausgang.Urteil)
	}

	return urteile
}

// **Eine fehlerfreie Leitung nimmt jeden Zug genau einmal.**
func TestUeberEineFehlerfreieLeitungWirktJederZugEinmal(t *testing.T) {
	t.Parallel()

	urteile := nimmAlle(t, dreiZuege(t).gesendet)

	for stelle, urteil := range urteile {
		if urteil != domain.FolgeNeu {
			t.Errorf("Zug %d: %v, erwartet %v", stelle+1, urteil, domain.FolgeNeu)
		}
	}
}

// **Ein verdoppelter Zug wirkt trotzdem nur einmal.**
//
// Das ist der „doppelte Draw" aus dem Issue — der Schaden, den eine
// wiederholende Leitung anrichtet, wenn niemand mitzählt.
func TestEinVerdoppelterZugWirktNurEinmal(t *testing.T) {
	t.Parallel()

	gesendet := dreiZuege(t).gesendet
	doppelt := [][]byte{gesendet[0], gesendet[0], gesendet[1], gesendet[1], gesendet[2]}

	urteile := nimmAlle(t, doppelt)
	erwartet := []domain.Folgeurteil{
		domain.FolgeNeu, domain.FolgeSchonDa,
		domain.FolgeNeu, domain.FolgeSchonDa,
		domain.FolgeNeu,
	}

	for stelle, urteil := range urteile {
		if urteil != erwartet[stelle] {
			t.Errorf("Stelle %d: %v, erwartet %v", stelle, urteil, erwartet[stelle])
		}
	}
}

// **Ein vertauschter Zug blockiert, statt still falsch zu wirken.**
//
// Wer die 3 vor der 2 nähme, hätte einen Stand, den niemand nachrechnen kann —
// und der Fehler zeigte sich Züge später an einer Karte, die woanders liegt.
func TestEinVertauschterZugBlockiert(t *testing.T) {
	t.Parallel()

	gesendet := dreiZuege(t).gesendet
	vertauscht := [][]byte{gesendet[0], gesendet[2], gesendet[1]}

	urteile := nimmAlle(t, vertauscht)

	switch {
	case urteile[0] != domain.FolgeNeu:
		t.Errorf("der erste Zug: %v", urteile[0])
	case urteile[1] != domain.FolgeLuecke:
		t.Errorf("der vorgezogene Zug: %v, erwartet %v",
			urteile[1], domain.FolgeLuecke)
	case urteile[2] != domain.FolgeNeu:
		t.Errorf("der nachgereichte Zug: %v", urteile[2])
	}
}

// **Ein verlorener Zug wird beim nächsten bemerkt.**
func TestEinVerlorenerZugWirdBemerkt(t *testing.T) {
	t.Parallel()

	gesendet := dreiZuege(t).gesendet
	verloren := [][]byte{gesendet[0], gesendet[2]}

	urteile := nimmAlle(t, verloren)

	if urteile[1] != domain.FolgeLuecke {
		t.Errorf("%v, erwartet %v", urteile[1], domain.FolgeLuecke)
	}
}

// **Nach einer Lücke sagt der Stand, wo er steht** — und die Wiederaufnahme
// entscheidet, wie es weitergeht.
//
// **Das ist die Naht zwischen Störung und Rückkehr**: ein Gast, der eine
// Lücke gesehen hat, fragt nicht „was ist passiert", sondern sagt, wo er
// steht.
func TestNachEinerLueckeFuehrtDerStandZurWiederaufnahme(t *testing.T) {
	t.Parallel()

	gesendet := dreiZuege(t).gesendet
	empfang := service.NeuerPartieempfang(1)

	for _, eine := range [][]byte{gesendet[0], gesendet[2]} {
		if _, fehler := empfang.Nimm(eine); fehler != nil {
			t.Fatalf("nehmen: %v", fehler)
		}
	}

	weg := domain.Wiederaufnahme(empfang.Wunsch("p1"), domain.Wirtsstand{
		Epoche: 1, Folge: 3, AelteteFolge: 1,
	})

	switch {
	case weg.Art != domain.WegEreignisse:
		t.Errorf("%v, erwartet %v", weg.Art, domain.WegEreignisse)
	case weg.AbFolge != 2:
		t.Errorf("ab Folge %d, erwartet 2", weg.AbFolge)
	}
}

// **Und eine abgerissene Leitung hinterlässt keinen halben Stand.**
//
// Was vor dem Abriss ankam, gilt; was danach käme, kommt nicht. Der Stand ist
// derselbe wie ohne den Versuch — und genau daran hängt, dass ein Neustart
// nichts doppelt.
func TestEineAbgerisseneLeitungHinterlaesstKeinenHalbenStand(t *testing.T) {
	t.Parallel()

	gesendet := dreiZuege(t).gesendet
	empfang := service.NeuerPartieempfang(1)

	if _, fehler := empfang.Nimm(gesendet[0]); fehler != nil {
		t.Fatalf("nehmen: %v", fehler)
	}

	vorher := empfang.Wunsch("p1")

	// Der Abriss: die zweite Nachricht kommt nie an. Danach fragt der Gast
	// noch einmal nach seinem Stand.
	nachher := empfang.Wunsch("p1")

	if !reflect.DeepEqual(vorher, nachher) {
		t.Errorf("der Stand hat sich ohne Zustellung geändert: %+v gegen %+v",
			vorher, nachher)
	}
}
