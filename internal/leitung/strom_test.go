package leitung_test

// **Eine Leitung über einen Strom, den jemand anderes gebaut hat** —
// Werkstück N3 von Phase 5.4
// (#1456).
//
// # Wofür
//
// **Weil tailcat einen `net.Conn` gibt und keine Adresse.** [Waehle] und
// [Horche] machen ihre Verbindung selbst; über ein fremdes Netz kommt sie von
// woanders. Alles darüber — TLS 1.3, das Festnageln auf die Kennung, die
// Kanalbindung, die Rahmung — ist dasselbe und soll es bleiben.

import (
	"errors"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestZweiSeitenSprechenUeberEinenFertigenStrom(t *testing.T) {
	t.Parallel()

	hier, dort := zweiIdentitaeten(t)

	hiesig, dortig := net.Pipe()

	fertig := make(chan error, 1)

	go func() {
		genommen, fehler := leitung.NimmStrom(t.Context(), dortig, dort)
		if fehler != nil {
			fertig <- fehler

			return
		}

		nachricht, fehler := genommen.Empfange(t.Context())
		if fehler != nil {
			fertig <- fehler

			return
		}

		if string(nachricht) != "Mausbraten" {
			fertig <- errors.New("falsche Nachricht: " + string(nachricht))

			return
		}

		fertig <- nil
	}()

	gewaehlt, fehler := leitung.UeberStrom(
		t.Context(), hiesig, hier, dort.Kennung())
	if fehler != nil {
		t.Fatalf("UeberStrom: %v", fehler)
	}

	if fehler := gewaehlt.Sende(t.Context(), []byte("Mausbraten")); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die andere Seite: %v", fehler)
	}

	// **Und der Kanal ist derselbe** — die Bindung kommt aus dem
	// Schlüsselmaterial und nicht aus der Adresse, die es hier gar nicht gibt.
	if gewaehlt.Kanal() == "" {
		t.Error("die Leitung hat keinen Kanal")
	}
}

// **Und ein falsches Gegenüber kommt auch über einen fremden Strom nicht
// durch.**
func TestEinFremderAmFertigenStromWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	hier, dort := zweiIdentitaeten(t)
	_, dritter := zweiIdentitaeten(t)

	hiesig, dortig := net.Pipe()

	go func() {
		//nolint:errcheck // die wählende Seite ist die, die hier prüft
		_, _ = leitung.NimmStrom(t.Context(), dortig, dort)
	}()

	_, fehler := leitung.UeberStrom(
		t.Context(), hiesig, hier, dritter.Kennung())
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet war eine Abweisung, kam: %v", fehler)
	}
}
