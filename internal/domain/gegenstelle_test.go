package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeGegenstellenklasseIstEineGegenstelle(t *testing.T) {
	t.Parallel()

	for _, klasse := range []error{
		domain.ErrNameUnaufloesbar, domain.ErrNichtErreichbar, domain.ErrKeineAntwort,
		domain.ErrAntwortStockt, domain.ErrProxyAnmeldung, domain.ErrVerbindungAbgerissen,
	} {
		if !errors.Is(klasse, domain.ErrGegenstelle) {
			t.Errorf("%v ist keine Gegenstelle", klasse)
		}
	}
}

func TestEinGegenstellenfehlerTraegtWirtFristUndGrund(t *testing.T) {
	t.Parallel()

	grund := errors.New("read tcp: i/o timeout")

	for _, fall := range []struct {
		name  string
		frist time.Duration
		satz  string
	}{
		{"ohne Frist nennt er nur den Wirt", 0, "api.scryfall.com:443"},
		{"mit Frist nennt er sie", 300 * time.Millisecond, "api.scryfall.com:443, nach 300ms"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, grund, "api.scryfall.com:443", fall.frist)

			var gegenstelle *domain.Gegenstellenfehler
			if !errors.As(fehler, &gegenstelle) || gegenstelle.Wirt != "api.scryfall.com:443" || gegenstelle.Frist != fall.frist {
				t.Fatalf("Wirt und Frist fehlen: %#v", gegenstelle)
			}

			if !errors.Is(fehler, domain.ErrAntwortStockt) || !strings.HasSuffix(fehler.Error(), fall.satz) {
				t.Errorf("der Satz %q endet nicht auf %q", fehler, fall.satz)
			}

			if !errors.Is(domain.GrundVon(fehler), grund) || strings.Contains(fehler.Error(), grund.Error()) {
				t.Errorf("der Grund gehört ins Protokoll und nicht in den Satz: %v", fehler)
			}
		})
	}
}

// **Eine Drosselung mit Uhrzeit** — R01 nach dem Zweitblick über Phase 0
// (#2333):
// die Uhrzeit stand nur im Satz, und das Fenster übersetzte sie weg.
func TestEineDrosselungTraegtIhrEnde(t *testing.T) {
	t.Parallel()

	bis := time.Date(2026, 10, 13, 21, 37, 0, 0, time.UTC)
	grund := errors.New("403, x-ratelimit-remaining: 0")

	fehler := domain.NeueDrosselung(grund, "api.github.com", bis)

	if !errors.Is(fehler, domain.ErrGedrosselt) {
		t.Errorf("keine Drosselung: %v", fehler)
	}

	var genauer *domain.Gegenstellenfehler
	if !errors.As(fehler, &genauer) || genauer.Wirt != "api.github.com" || !genauer.Bis.Equal(bis) {
		t.Fatalf("Wirt und Ende kommen nicht an: %+v", genauer)
	}

	if !strings.Contains(fehler.Error(), "api.github.com") || strings.Contains(fehler.Error(), "remaining") {
		t.Errorf("der Satz nennt den Wirt nicht oder zeigt den Grund: %q", fehler.Error())
	}

	if !errors.Is(domain.GrundVon(fehler), grund) {
		t.Errorf("der Grund ist fort: %v", domain.GrundVon(fehler))
	}
}
