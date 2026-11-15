package service_test

// **Ein schweigender Klient hält den nächsten nicht auf** — SEC-04
// (#2139).
//
// Der Horchposten nahm der Reihe nach an, und `Nimm` führte den
// TLS-Handschlag: wer eine Verbindung öffnete und schwieg, hielt jeden
// folgenden bis zur Handschlagfrist auf — zehn Sekunden, und wer sie immer
// wieder öffnete, für immer. Jetzt nehmen mehrere zugleich an; ein Schweiger
// hält einen von ihnen.

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tlslauscher hängt einen echten Horcher an [service.Hereinkommend].
type tlslauscher struct{ *leitung.Horcher }

func (l tlslauscher) Nimm(ctx context.Context) (service.Verbindung, error) {
	return l.Horcher.Nimm(ctx)
}

// tlswaehler wählt über TCP mit TLS.
type tlswaehler struct {
	adresse string
	ich     peer.Identitaet
}

func (w tlswaehler) Waehle(ctx context.Context, kennung peer.Kennung, _ string) (service.Verbindung, error) {
	return leitung.Waehle(ctx, w.adresse, w.ich, kennung)
}

func TestEinSchweigenderKlientHaeltDenNaechstenNichtAuf(t *testing.T) {
	t.Parallel()

	hiesige, dortige := erzeugt(t), erzeugt(t)

	horcher, fehler := leitung.Horche(t.Context(), "127.0.0.1:0", dortige)
	if fehler != nil {
		t.Fatal(fehler)
	}

	dort := service.NeuePeersitzungen(&rechteprobe{darf: true}, nil, ausweisprobe{ich: dortige}, einePeeruhr, 2)
	hier := service.NeuePeersitzungen(&rechteprobe{darf: true},
		tlswaehler{adresse: horcher.Adresse(), ich: hiesige}, ausweisprobe{ich: hiesige}, einePeeruhr, 2).MitZufall(nil)

	ctx, aufhoeren := context.WithCancel(t.Context())
	fertig := make(chan struct{})

	go func() {
		service.NeuerHorchposten(dort, nil).Bediene(ctx, tlslauscher{horcher})
		close(fertig)
	}()

	t.Cleanup(func() {
		aufhoeren()
		<-fertig
	})

	// **Der Schweiger** — eine Verbindung, die nie etwas sagt.
	schweiger, fehler := (&net.Dialer{}).DialContext(t.Context(), "tcp", horcher.Adresse())
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = schweiger.Close() }) //nolint:errcheck // nur fürs Aufräumen — womöglich schon zu

	// **Und einer, der etwas will** — er ist fertig, lange bevor die Frist
	// des Schweigers abläuft.
	anfang := time.Now()
	griff, fehler := hier.Oeffne(t.Context(), dortige.Kennung(), domain.ZweckTeilen)
	dauer := time.Since(anfang)

	if fehler != nil {
		t.Fatalf("der zweite Klient: %v (nach %v)", fehler, dauer)
	}

	_ = griff.Schliesse() //nolint:errcheck // nur ein Lauf

	if dauer > leitung.Handschlagfrist/2 {
		t.Errorf("der zweite Klient wartete %v — der Schweiger hielt ihn auf", dauer)
	}
}
