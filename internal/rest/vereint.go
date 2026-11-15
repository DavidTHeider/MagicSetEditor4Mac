package rest

// **Ein Dienst, mehrere Adressen** — G08
// (#2272).
//
// Jede Adresse der Lauschwahl bekommt ihren eigenen Lauscher, und der Dienst
// dahinter sieht **einen** ([Vereint]): Verbindungsgrenze, Zustände, Fristen
// und das Ende bleiben, wie sie waren — sie fragen einen `net.Listener`, und
// woher die Verbindung kam, ist für sie gleich.
//
// # Was eine fehlende Adresse tut
//
// **Sie wird benannt, und die übrigen laufen** — eine Schnittstelle, die seit
// der Wahl verschwand (ein abgezogenes Kabel, eine beendete VM), soll den
// Dienst auf der Schleife nicht mitnehmen. **Ohne eine einzige** startet
// nichts, und der Fehler nennt jede.
//
// **Eine Öffnung ohne TLS nimmt dagegen die ganze Wahl zurück**: wer nach
// außen gewählt hat, soll nicht still nur die Schleife bekommen — das ist
// eine Eingabe, die nicht passt, und kein Ausfall.

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// LauschenAufAllen öffnet je Bindung einen Lauscher und gibt sie vereint.
func LauschenAufAllen(
	ctx context.Context, bindungen []domain.Bindung, zeugnis *tls.Config,
) (net.Listener, []domain.Lauschbefund, error) {
	for _, bindung := range bindungen {
		if fehler := domain.PruefeBindung(bindung); fehler != nil {
			return nil, nil, fehler
		}

		if bindung.VerlaesstDenRechner() && zeugnis == nil {
			return nil, nil, fmt.Errorf(
				"%w: %s verlässt diesen Rechner und braucht deshalb TLS — ein "+
					"Zugangstoken ginge sonst im Klartext über das Netz",
				domain.ErrValidation, bindung.Wirkadresse())
		}
	}

	var (
		offen   []net.Listener
		befunde []domain.Lauschbefund
		gruende []error
	)

	for _, bindung := range bindungen {
		lauscher, fehler := LauschenMit(ctx, bindung, zeugnis)
		if fehler != nil {
			befunde = append(befunde, domain.Lauschbefund{Adresse: bindung.Wirkadresse(), Grund: fehler.Error()})
			gruende = append(gruende, fmt.Errorf("%s: %w", bindung.Wirkadresse(), fehler))

			continue
		}

		offen = append(offen, lauscher)
	}

	if len(offen) == 0 {
		return nil, nil, fmt.Errorf("auf keiner der gewählten Adressen ließ sich lauschen: %w",
			errors.Join(gruende...))
	}

	return Vereint(offen...), befunde, nil
}

// Vereint reicht die Verbindungen mehrerer Lauscher als einer weiter.
//
// **`Addr` nennt den ersten** — wer nur eine Adresse kennt, bekommt die, die
// zuerst gewählt wurde; [Adressen] nennt alle.
func Vereint(lauscher ...net.Listener) net.Listener {
	v := &vereint{lauscher: lauscher, herein: make(chan angenommen), zu: make(chan struct{})}

	for _, einer := range lauscher {
		go v.annehmen(einer)
	}

	return v
}

type angenommen struct {
	verbindung net.Conn
	fehler     error
}

type vereint struct {
	lauscher []net.Listener
	herein   chan angenommen
	zu       chan struct{}
	einmal   sync.Once
}

func (v *vereint) annehmen(einer net.Listener) {
	for {
		verbindung, fehler := einer.Accept()

		select {
		case v.herein <- angenommen{verbindung, fehler}:
		case <-v.zu:
			if verbindung != nil {
				_ = verbindung.Close() //nolint:errcheck // niemand nimmt sie mehr an
			}

			return
		}

		if fehler != nil && errors.Is(fehler, net.ErrClosed) {
			return
		}
	}
}

func (v *vereint) Accept() (net.Conn, error) {
	select {
	case eines := <-v.herein:
		return eines.verbindung, eines.fehler
	case <-v.zu:
		return nil, net.ErrClosed
	}
}

func (v *vereint) Close() error {
	var gruende []error

	v.einmal.Do(func() {
		close(v.zu)

		for _, einer := range v.lauscher {
			gruende = append(gruende, einer.Close())
		}
	})

	return errors.Join(gruende...)
}

func (v *vereint) Addr() net.Addr { return v.lauscher[0].Addr() }

// Adressen nennt die Adressen eines Lauschers — aller, wenn er vereint ist.
func Adressen(lauscher net.Listener) []string {
	v, istVereint := lauscher.(*vereint)
	if !istVereint {
		return []string{lauscher.Addr().String()}
	}

	adressen := make([]string, 0, len(v.lauscher))
	for _, einer := range v.lauscher {
		adressen = append(adressen, einer.Addr().String())
	}

	return adressen
}
