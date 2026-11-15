package service_test

// **Welche Partie der Wirt meint — aus seinem ersten Wort** — UX-08.AK-04
// (#1680).
//
// # Was hier gemessen wird
//
// **Dass ein Gast die Partie nicht nennen muss.** Der Wirt spricht zuerst
// und schickt das Tischbild für den vergebenen Sitz; sein Umschlag trägt die
// Partie. Der Gast liest, bis etwas Gültiges kommt, nimmt die Partie daraus —
// **und verliert die Nachricht dabei nicht**: die Schleife danach bekommt sie
// als erste.

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// angenommenedoppel ist ein Zugang mit Auskunft und Ende.
type angenommenedoppel struct {
	*zugangsdoppel
}

func (angenommenedoppel) Stand() service.Sitzungsstand { return service.Sitzungsstand{} }

func (angenommenedoppel) Schliesse() error { return nil }

// **Der Wirt begrüßt mit dem Tischbild des Sitzes — und nennt so die Partie.**
func TestDerWirtBegruesstMitDemTischbildDesSitzes(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, naht := einWirt(nehmer)

	if fehler := wirt.Begruesse(context.Background()); fehler != nil {
		t.Fatalf("Begruesse: %v", fehler)
	}

	gesendet := naht.letzte(t)

	art, inhalt, fehler := domain.LiesKontrolle(
		gesendet.Inhalt, gesendet.Von, domain.KontrolleTischbild)

	switch {
	case fehler != nil:
		t.Fatalf("die Begrüßung ist kein Tischbild: %v", fehler)
	case art != domain.KontrolleTischbild || len(inhalt) == 0:
		t.Errorf("es kam %q", art)
	case gesendet.Partie != "p1":
		t.Errorf("der Umschlag nennt %q statt p1", gesendet.Partie)
	case !slices.Equal(nehmer.gefragt, []string{"gast"}):
		t.Errorf("gerechnet wurde für %v", nehmer.gefragt)
	}
}

// **Was beim Rechnen des Bildes klagt, kommt zurück — und nichts geht hinaus.**
func TestEineBegruessungOhneBildKlagt(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Partie ließ sich nicht lesen")

	wirt, naht := einWirt(&nehmerdoppel{bildklage: kaputt})

	if fehler := wirt.Begruesse(context.Background()); !errors.Is(fehler, kaputt) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(naht.gesendet) != 0 {
		t.Errorf("%d Nachrichten hinaus", len(naht.gesendet))
	}
}

// **Die erste gültige Nachricht nennt die Partie — und sie kommt noch einmal.**
//
// Davor stehen eine, die nicht mehr gilt, und eine, die sich nicht lesen
// lässt: **beide werden übergangen**, wie in [service.Partienetz.Bediene].
func TestDieErsteGueltigeNachrichtNenntDiePartie(t *testing.T) {
	t.Parallel()

	bild := eineGruessnachricht(t, "p7")
	danach := verpackt(t, einEreignis(2))

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		[]byte(`kein JSON`),
		verpackt(t, domain.Partienachricht{Art: domain.ArtKontrolle, Von: "wirt"}),
		bild,
		danach,
	}}

	partie, weiter, fehler := service.Partieerfragen(
		context.Background(), angenommenedoppel{zugang})
	if fehler != nil {
		t.Fatalf("Partieerfragen: %v", fehler)
	}

	if partie != "p7" {
		t.Errorf("genannt wurde %q", partie)
	}

	// **Zuerst die vorgelesene, dann die Leitung** — und die Auskunft des
	// Griffs bleibt die des Griffs.
	for _, erwartet := range [][]byte{bild, danach} {
		roh, lesefehler := weiter.Empfange(context.Background())
		if lesefehler != nil || !bytes.Equal(roh, erwartet) {
			t.Fatalf("gelesen %q (%v), erwartet %q", roh, lesefehler, erwartet)
		}
	}

	if _, lesefehler := weiter.Empfange(
		context.Background()); !errors.Is(lesefehler, errStromZuEnde) {
		t.Errorf("danach kam %v", lesefehler)
	}
}

// **Eine Nachricht, die nicht mehr gilt, nennt keine Partie.**
//
// Sie gehört zu einem Stand, den es nicht mehr gibt (§7, Regel 2) — wer ihre
// Partie nähme, säße an einem Tisch von vorhin.
func TestEineVeralteteNachrichtNenntKeinePartie(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{
		zustellungen: [][]byte{eineGruessnachricht(t, "p7")},
		veraltet:     true,
	}

	if _, _, fehler := service.Partieerfragen(
		context.Background(), angenommenedoppel{zugang},
	); !errors.Is(fehler, errStromZuEnde) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Reißt die Leitung vorher ab, kommt ihr Grund zurück.**
func TestOhneErstesWortKommtDerGrund(t *testing.T) {
	t.Parallel()

	abgerissen := errors.New("die Leitung ist weg")

	if _, weiter, fehler := service.Partieerfragen(context.Background(),
		angenommenedoppel{&zugangsdoppel{lesefehler: abgerissen}},
	); !errors.Is(fehler, abgerissen) || weiter != nil {
		t.Errorf("es kam %v und %v", weiter, fehler)
	}
}

// eineGruessnachricht ist das Tischbild, mit dem ein Wirt begrüßt.
func eineGruessnachricht(t *testing.T, partie string) []byte {
	t.Helper()

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleTischbild,
		domain.Tischbild{Sicht: domain.Partiesicht{PartieID: partie, Fuer: "bert"}})
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	return verpackt(t, domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: partie, Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	})
}
