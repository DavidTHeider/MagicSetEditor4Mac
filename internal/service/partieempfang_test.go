package service_test

// **Was von der Leitung hereinkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partiesitzung_test.go` herausgelöst**, als sie über die Zeilengrenze
// aus §2 ging. Der Schnitt folgt der Frage und derselben Naht wie im
// Produktivcode: dort steht, was **hinausgeht**, hier, was **hereinkommt** —
// und das ist die Seite mit den Fällen, die eine Leitung von selbst erzeugt:
// eine zweite Zustellung, eine Lücke, eine alte Epoche.
//
// # Die Zusage
//
// Das Issue nennt den Schaden beim Namen: Wiederholungen „erzeugen keinen
// doppelten Draw". Eine zweite Zustellung ist auf einer Leitung, die
// wiederholt, der **Normalfall** — sie darf nur nicht zweimal wirken.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Was ankommt, wird geprüft, bevor es gilt.**
//
// Das Issue nennt den Schaden beim Namen: „erzeugen keinen doppelten Draw".
// Eine zweite Zustellung ist auf einer Leitung, die wiederholt, der
// Normalfall — sie darf nur nicht zweimal wirken.
func TestEinDoppeltZugestelltesEreignisWirktNurEinmal(t *testing.T) {
	t.Parallel()

	empfang := service.NeuerPartieempfang(1)
	roh := gepackt(t, domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "sitz-2", Inhalt: []byte(`{"a":1}`), Kette: langeKette,
	})

	erste, fehler := empfang.Nimm(roh)
	if fehler != nil {
		t.Fatalf("die erste Zustellung: %v", fehler)
	}

	if erste.Urteil != domain.FolgeNeu {
		t.Fatalf("%v, erwartet %v", erste.Urteil, domain.FolgeNeu)
	}

	zweite, nochmal := empfang.Nimm(roh)
	switch {
	case nochmal != nil:
		t.Errorf("die zweite Zustellung klagte: %v", nochmal)
	case zweite.Urteil != domain.FolgeSchonDa:
		t.Errorf("%v, erwartet %v", zweite.Urteil, domain.FolgeSchonDa)
	}
}

// **Eine Lücke blockiert und wird gemeldet** — sie ist kein Fall, den man
// still übergeht.
func TestEineLueckeWirdGemeldet(t *testing.T) {
	t.Parallel()

	empfang := service.NeuerPartieempfang(1)

	ausgang, fehler := empfang.Nimm(gepackt(t, domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 5,
		Von: "sitz-2", Inhalt: []byte(`{}`), Kette: langeKette,
	}))

	switch {
	case fehler != nil:
		t.Fatalf("nehmen: %v", fehler)
	case ausgang.Urteil != domain.FolgeLuecke:
		t.Errorf("%v, erwartet %v", ausgang.Urteil, domain.FolgeLuecke)
	}
}

// **Präsenz geht an der Reihenfolge vorbei** — sie hat keine.
func TestPraesenzGehtAnDerReihenfolgeVorbei(t *testing.T) {
	t.Parallel()

	empfang := service.NeuerPartieempfang(1)

	for i := 0; i < 3; i++ {
		ausgang, fehler := empfang.Nimm(gepackt(t, domain.Partienachricht{
			Art: domain.ArtPraesenz, Partie: "p1", Epoche: 1,
			Von: "sitz-2", Inhalt: []byte(`{"x":1}`),
		}))

		switch {
		case fehler != nil:
			t.Fatalf("nehmen: %v", fehler)
		case ausgang.Urteil != domain.FolgeNeu:
			t.Errorf("%v, erwartet %v", ausgang.Urteil, domain.FolgeNeu)
		}
	}
}

// **Was keine Nachricht ist, kommt nicht durch** (§8).
func TestWasKeineNachrichtIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string][]byte{
		"kein JSON":       []byte("{nicht"),
		"ohne Art":        gepackt(t, domain.Partienachricht{Partie: "p1", Epoche: 1}),
		"leere Nachricht": {},
	}

	for name, roh := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := service.NeuerPartieempfang(1).Nimm(roh); fehler == nil {
				t.Error("es kam durch")
			}
		})
	}
}

// **Eine alte Epoche wird gemeldet und nicht genommen.**
func TestEineAlteEpocheWirdGemeldet(t *testing.T) {
	t.Parallel()

	empfang := service.NeuerPartieempfang(3)

	ausgang, fehler := empfang.Nimm(gepackt(t, domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "sitz-2", Inhalt: []byte(`{}`), Kette: langeKette,
	}))

	switch {
	case fehler != nil:
		t.Fatalf("nehmen: %v", fehler)
	case ausgang.Urteil != domain.FolgeAlteEpoche:
		t.Errorf("%v, erwartet %v", ausgang.Urteil, domain.FolgeAlteEpoche)
	}
}

// **Der Stand ist ablesbar** — ein Gast braucht ihn für die Wiederaufnahme.
func TestDerEmpfangsstandIstAblesbar(t *testing.T) {
	t.Parallel()

	empfang := service.NeuerPartieempfang(1)
	if _, fehler := empfang.Nimm(gepackt(t, domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "sitz-2", Inhalt: []byte(`{}`), Kette: langeKette,
	})); fehler != nil {
		t.Fatalf("nehmen: %v", fehler)
	}

	wunsch := empfang.Wunsch("p1")
	switch {
	case wunsch.Folge != 1:
		t.Errorf("Folge %d, erwartet 1", wunsch.Folge)
	case wunsch.Kette != langeKette:
		t.Errorf("Kette %q", wunsch.Kette)
	case wunsch.Partie != "p1":
		t.Errorf("Partie %q", wunsch.Partie)
	}
}
