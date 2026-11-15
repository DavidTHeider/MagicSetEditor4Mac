package service_test

// **Wo ein Handschlag scheitern kann** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum jeder Schritt einen Lauf bekommt
//
// **Weil ein Handschlag sechs Wege über die Leitung hat** — drei hin, drei
// her —, und jeder von ihnen kann abreißen. Ein verschluckter Fehler wäre hier
// eine Sitzung, die eine Seite für offen hält und die andere nicht; §10
// verlangt, dass jeder Fehlerweg betreten wird, und hier ist er die halbe
// Fachlichkeit.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stolperleitung reicht durch — bis zum gezählten Schritt.
type stolperleitung struct {
	schloss sync.Mutex
	innen   *leitung.Leitung

	// schliessfehler lässt das Niederlegen klagen.
	schliessfehler error

	// sendefehler und empfangsfehler zählen ab eins: 1 heißt „der erste".
	sendefehler    int
	empfangsfehler int

	// unsinnAb ersetzt die n-te **empfangene** Nachricht durch etwas, das
	// keine ist — die Leitung trägt Bytes, nicht Bedeutung.
	unsinnAb int

	gesendet  int
	empfangen int
}

func (s *stolperleitung) Kanal() string { return s.innen.Kanal() }

func (s *stolperleitung) Gegenstelle() peer.Kennung { return s.innen.Gegenstelle() }
func (s *stolperleitung) Schliesse() error {
	if fehler := s.innen.Schliesse(); fehler != nil {
		return fehler
	}

	return s.schliessfehler
}

func (s *stolperleitung) Sende(ctx context.Context, nachricht []byte) error {
	s.schloss.Lock()
	s.gesendet++
	stolpert := s.gesendet == s.sendefehler
	s.schloss.Unlock()

	if stolpert {
		return errors.New("die Leitung riss beim Senden")
	}

	return s.innen.Sende(ctx, nachricht)
}

func (s *stolperleitung) Empfange(ctx context.Context) ([]byte, error) {
	s.schloss.Lock()
	s.empfangen++
	stolpert := s.empfangen == s.empfangsfehler
	s.schloss.Unlock()

	if stolpert {
		return nil, errors.New("die Leitung riss beim Empfangen")
	}

	gelesen, fehler := s.innen.Empfange(ctx)
	if fehler != nil {
		return nil, fehler
	}

	s.schloss.Lock()
	verdreht := s.empfangen == s.unsinnAb
	s.schloss.Unlock()

	if verdreht {
		return []byte("kein JSON"), nil
	}

	return gelesen, fehler
}

// stolperwaehler gibt eine stolpernde Leitung.
type stolperwaehler struct{ leitung *stolperleitung }

func (s *stolperwaehler) Waehle(
	context.Context, peer.Kennung, string,
) (service.Verbindung, error) {
	return s.leitung, nil
}

// mitStolper baut zwei Seiten, deren Leitungen an einer Stelle reißen.
func mitStolper(
	t *testing.T, hier, dort *stolperleitung,
) (diesseits, jenseits *service.Peersitzungen, jenseitige peer.Kennung) {
	t.Helper()

	hiesige, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	dortige, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	hiesigeLeitung, dortigeLeitung := leitung.Schleife("kanal-schleife")

	t.Cleanup(func() {
		_ = hiesigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
		_ = dortigeLeitung.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	hier.innen, dort.innen = hiesigeLeitung, dortigeLeitung

	return service.NeuePeersitzungen(&rechteprobe{darf: true},
			&stolperwaehler{leitung: hier}, ausweisprobe{ich: hiesige},
			einePeeruhr, 2).MitZufall(nil),
		// **Ohne Wähler** — sie bekommt ihre Leitung hereingereicht.
		service.NeuePeersitzungen(&rechteprobe{darf: true}, nil,
			ausweisprobe{ich: dortige}, einePeeruhr, 2),
		dortige.Kennung()
}

// **Jeder der sechs Schritte lässt die Sitzung scheitern.**
//
// **Und auf beiden Seiten**: eine Sitzung, die eine Seite für offen hält und
// die andere nicht, ist schlimmer als gar keine.
func TestJederSchrittDesHandschlagsKannScheitern(t *testing.T) {
	t.Parallel()

	// **Zeiger und keine Werte** — eine `stolperleitung` trägt ein Schloss,
	// und wer sie aus einer Karte kopierte, kopierte es mit.
	faelle := map[string]struct{ hier, dort *stolperleitung }{
		"das Angebot geht nicht hinaus": {
			hier: &stolperleitung{sendefehler: 1}, dort: &stolperleitung{},
		},
		"das Angebot kommt nicht an": {
			hier: &stolperleitung{}, dort: &stolperleitung{empfangsfehler: 1},
		},
		"die Annahme geht nicht hinaus": {
			hier: &stolperleitung{}, dort: &stolperleitung{sendefehler: 1},
		},
		"die Annahme kommt nicht an": {
			hier: &stolperleitung{empfangsfehler: 1}, dort: &stolperleitung{},
		},
		"der Abschluss geht nicht hinaus": {
			hier: &stolperleitung{sendefehler: 2}, dort: &stolperleitung{},
		},
		"der Abschluss kommt nicht an": {
			hier: &stolperleitung{}, dort: &stolperleitung{empfangsfehler: 2},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hier, dort, kennung := mitStolper(t, fall.hier, fall.dort)

			var (
				warten     sync.WaitGroup
				dortFehler error
			)

			warten.Add(1)

			go func() {
				defer warten.Done()

				_, dortFehler = dort.Annehmen(t.Context(), fall.dort)
			}()

			_, hierFehler := hier.Oeffne(
				t.Context(), kennung, domain.ZweckTeilen)

			warten.Wait()

			if hierFehler == nil {
				t.Error("die anbietende Seite kam durch")
			}

			if dortFehler == nil {
				t.Error("die annehmende Seite kam durch")
			}

			if len(hier.Alle()) != 0 || len(dort.Alle()) != 0 {
				t.Fatalf("es steht etwas in den Übersichten: %d und %d",
					len(hier.Alle()), len(dort.Alle()))
			}
		})
	}
}

// **Ein Fehler der Rechteauskunft auf der annehmenden Seite ist kein Nein.**
//
// **Der Unterschied zählt auch hier**: „darf nicht" ist eine Antwort, „ich
// weiß es nicht" ist keine — und eine Sitzung auf einer unbeantworteten Frage
// wäre eine ohne Freigabe.
func TestEinFehlerBeimAnnehmenIstKeinNein(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	paar.dortigeRechte.fehler = errors.New("kein Bestand")

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		_, dortFehler = paar.dort.Annehmen(t.Context(), paar.dortigeLeitung)
	}()

	if _, fehler := paar.hier.Oeffne(
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen,
	); fehler == nil {
		t.Error("die Sitzung kam zustande")
	}

	warten.Wait()

	if dortFehler == nil {
		t.Fatal("die annehmende Seite schwieg")
	}
}

// **Ein Wechsel während des Annehmens lässt keine Leitung offen.**
//
// **Dieselbe Lücke wie beim Wählen, nur auf der anderen Seite**: der
// Handschlag geht dreimal über das Netz, und wer in dieser Zeit das Profil
// wechselt, bekäme eine Sitzung, die niemand mehr führt.
func TestEinWechselWaehrendDesAnnehmensLaesstKeineLeitungOffen(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var (
		warten     sync.WaitGroup
		dortFehler error
	)

	warten.Add(1)

	go func() {
		defer warten.Done()

		_, dortFehler = paar.dort.Annehmen(t.Context(), paar.dortigeLeitung)
	}()

	// **Der Wechsel geschieht, während der Handschlag läuft** — das Angebot
	// ist hinaus, die Annahme unterwegs.
	paar.dortigeRechte.dazwischen = func() {
		if fehler := paar.dort.Pausiere(); fehler != nil {
			t.Errorf("Pausiere: %v", fehler)
		}
	}

	_, _ = paar.hier.Oeffne( //nolint:errcheck // die anbietende Seite darf hier scheitern
		t.Context(), paar.dortige.Kennung(), domain.ZweckTeilen)

	warten.Wait()

	if !errors.Is(dortFehler, domain.ErrConflict) {
		t.Fatalf("die annehmende Seite sagte: %v", dortFehler)
	}

	if len(paar.dort.Alle()) != 0 {
		t.Fatal("die Sitzung steht in der Übersicht")
	}
}
