package service_test

// **Hinaus geht es in der Reihenfolge, in der der Kern annahm** — #2481
// (Zweitblick über Phase 7 von 0.9.0, Befund 8): das Schloss `zug` ordnete
// die Züge, aber nicht, was danach an den Tisch ging. Ein späterer Zug konnte
// sein Tischbild vor dem früheren verschicken, und der Gast sah danach den
// älteren Stand.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// meldendeProbe sagt nach jedem Schreiben Bescheid.
type meldendeProbe struct {
	*partieprobe

	geschrieben chan struct{}
}

func (m *meldendeProbe) Anwenden(
	ctx context.Context, erwartet int64, befehlID string,
	zustand domain.Partie, ereignisse []domain.Ereignis,
) error {
	fehler := m.partieprobe.Anwenden(ctx, erwartet, befehlID, zustand, ereignisse)
	if fehler == nil {
		m.geschrieben <- struct{}{}
	}

	return fehler
}

func einZug(kennung, art string) domain.Spielbefehl {
	return domain.Spielbefehl{
		PartieID: "p1", Kennung: kennung, Absender: "sitz-1",
		Art: art, Nutzlast: map[string]any{"objekt": "o1"},
	}
}

func TestEinSpaetererZugGehtNichtVorDemFrueherenHinaus(t *testing.T) {
	t.Parallel()

	probe := &meldendeProbe{
		partieprobe: &partieprobe{partien: map[string]domain.Partie{}},
		geschrieben: make(chan struct{}, 2),
	}
	leitung := service.NeuePartieleitung(probe, service.NeuerSandkasten(service.NeuerZufall()))
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	hinaus := make(chan int64, 2)
	angefangen := make(chan struct{})
	weiter := make(chan struct{})

	erster := make(chan error, 1)
	go func() {
		_, fehler := leitung.BefehlVerkuendet(ctx, einZug("b1", "objekt.tappen"),
			func(ausgang service.Befehlsausgang) {
				close(angefangen)
				<-weiter
				hinaus <- ausgang.Fassung
			})
		erster <- fehler
	}()

	<-angefangen

	zweiter := make(chan error, 1)
	go func() {
		_, fehler := leitung.BefehlVerkuendet(ctx, einZug("b2", "objekt.enttappen"),
			func(ausgang service.Befehlsausgang) { hinaus <- ausgang.Fassung })
		zweiter <- fehler
	}()

	// **Beide sind angenommen** — das ist der Beleg, dass der zweite
	// unterwegs ist, bevor gefragt wird, ob er überholt.
	<-probe.geschrieben
	<-probe.geschrieben

	// **Ginge er vor dem ersten hinaus, wäre er in dieser Frist durch.** Er
	// wartet, also läuft sie ab — und erst dann darf der erste weiter.
	select {
	case fehler := <-zweiter:
		zweiter <- fehler
	case <-time.After(50 * time.Millisecond):
	}

	close(weiter)

	for name, kanal := range map[string]chan error{"der erste": erster, "der zweite": zweiter} {
		if fehler := <-kanal; fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}
	}

	if zuerst, danach := <-hinaus, <-hinaus; zuerst > danach {
		t.Errorf("hinaus ging zuerst die Fassung %d, dann %d", zuerst, danach)
	}
}

// **Eine Antwort ohne Zug wartet, bis die laufende Verkündung durch ist** —
// sonst überholte ihr Bild das des Zuges.
func TestEineAntwortOhneZugWartetAufDieVerkuendung(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	hinaus := make(chan string, 2)
	angefangen := make(chan struct{})
	weiter := make(chan struct{})

	zug := make(chan error, 1)
	go func() {
		_, fehler := leitung.BefehlVerkuendet(ctx, einZug("b1", "objekt.tappen"),
			func(service.Befehlsausgang) {
				close(angefangen)
				<-weiter
				hinaus <- "der Zug"
			})
		zug <- fehler
	}()

	<-angefangen

	antwort := make(chan error, 1)
	go func() {
		antwort <- leitung.InReihe(func() error {
			hinaus <- "die Antwort"

			return nil
		})
	}()

	// **Ginge sie vorbei, wäre sie in dieser Frist durch.**
	select {
	case fehler := <-antwort:
		antwort <- fehler
	case <-time.After(50 * time.Millisecond):
	}

	close(weiter)

	for name, kanal := range map[string]chan error{"der Zug": zug, "die Antwort": antwort} {
		if fehler := <-kanal; fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}
	}

	if zuerst := <-hinaus; zuerst != "der Zug" {
		t.Errorf("zuerst ging %s hinaus", zuerst)
	}
}

// **Auch wer liest und schickt, ohne zu schreiben, geht in der Reihe** — die
// Absage, der Gruß und die Rückkehr. Ihr Bild darf kein neueres überholen.
func TestAntwortenOhneZugGehenInDerReihe(t *testing.T) {
	t.Parallel()

	stand := domain.Wirtsstand{Epoche: domain.Partieepoche, Folge: 5, Kette: "abc", AelteteFolge: 1}

	for name, fall := range map[string]struct {
		nehmer *nehmerdoppel
		tun    func(*testing.T, *service.Partiewirt) error
	}{
		"die Absage": {
			nehmer: &nehmerdoppel{klagt: fmt.Errorf("%w: geht nicht", domain.ErrValidation)},
			tun: func(t *testing.T, wirt *service.Partiewirt) error {
				t.Helper()

				return wirt.Angekommen(t.Context(), einSpielwunsch(t,
					domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))
			},
		},
		"der Gruß": {
			nehmer: &nehmerdoppel{},
			tun: func(t *testing.T, wirt *service.Partiewirt) error {
				t.Helper()

				return wirt.Begruesse(t.Context())
			},
		},
		"die Rückkehr": {
			nehmer: &nehmerdoppel{stand: stand},
			tun: func(t *testing.T, wirt *service.Partiewirt) error {
				t.Helper()

				return wirt.Angekommen(t.Context(), einRueckkehrwunsch(t, 5, "abc"))
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wirt, naht := einWirt(fall.nehmer)

			if fehler := fall.tun(t, wirt); fehler != nil {
				t.Fatalf("es kam %v", fehler)
			}

			if fall.nehmer.reihe != 1 || len(naht.gesendet) == 0 {
				t.Errorf("%d-mal in der Reihe, %d Nachrichten hinaus",
					fall.nehmer.reihe, len(naht.gesendet))
			}
		})
	}
}
