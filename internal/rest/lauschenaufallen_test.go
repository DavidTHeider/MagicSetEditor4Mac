package rest_test

// **Ein Dienst, mehrere Adressen** — G08
// (#2272).
//
// Jede Adresse der Lauschwahl bekommt ihren Lauscher, und der Dienst sieht
// **einen**: `Vereint` reicht die Verbindungen aller weiter. Eine Adresse,
// die es nicht (mehr) gibt, wird benannt, und die übrigen laufen.
//
// **`192.0.2.1` ist auf keinem Rechner vergeben** (TEST-NET-1, RFC 5737) —
// der Fall „verschwundene Schnittstelle" hängt so an keiner Eigenheit dieses
// Rechners (C06).

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func TestJedeGewaehlteAdresseBekommtIhrenLauscher(t *testing.T) {
	t.Parallel()

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,::1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	vereint, befunde, fehler := rest.LauschenAufAllen(t.Context(), wahl.Bindungen(0), nil)
	if fehler != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", fehler, befunde)
	}

	t.Cleanup(func() { _ = vereint.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	adressen := rest.Adressen(vereint)
	if len(adressen) != 2 || !strings.HasPrefix(adressen[0], "127.0.0.1:") || !strings.HasPrefix(adressen[1], "[::1]:") {
		t.Fatalf("die Lauscher liegen auf %v", adressen)
	}

	if vereint.Addr().String() != adressen[0] {
		t.Errorf("Addr nennt %s statt der ersten", vereint.Addr())
	}

	// **Über jede Adresse kommt eine Verbindung an.**
	for _, adresse := range adressen {
		go func() {
			verbindung, wahlfehler := (&net.Dialer{}).DialContext(context.Background(), "tcp", adresse)
			if wahlfehler == nil {
				_ = verbindung.Close() //nolint:errcheck // sie sollte nur ankommen
			}
		}()
	}

	gesehen := map[string]bool{}

	for range adressen {
		verbindung, annahmefehler := vereint.Accept()
		if annahmefehler != nil {
			t.Fatalf("Accept: %v", annahmefehler)
		}

		gesehen[verbindung.LocalAddr().String()] = true
		_ = verbindung.Close() //nolint:errcheck // angenommen ist angenommen
	}

	if len(gesehen) != 2 {
		t.Errorf("angenommen auf %v", gesehen)
	}

	// **Und geschlossen ist geschlossen** — an jeder Adresse.
	if fehler := vereint.Close(); fehler != nil {
		t.Fatalf("Close: %v", fehler)
	}

	if _, annahmefehler := vereint.Accept(); !errors.Is(annahmefehler, net.ErrClosed) {
		t.Errorf("nach dem Schließen: %v", annahmefehler)
	}

	for _, adresse := range adressen {
		if verbindung, wahlfehler := (&net.Dialer{}).DialContext(t.Context(), "tcp", adresse); wahlfehler == nil {
			_ = verbindung.Close() //nolint:errcheck // es sollte gar keine geben
			t.Errorf("auf %s nimmt nach dem Schließen noch jemand an", adresse)
		}
	}
}

func TestEineVerschwundeneAdresseWirdBenanntUndDieAnderenLaufen(t *testing.T) {
	t.Parallel()

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,192.0.2.1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	vereint, befunde, fehler := rest.LauschenAufAllen(t.Context(), wahl.Bindungen(0), &tls.Config{MinVersion: tls.VersionTLS13})
	if fehler != nil {
		t.Fatalf("eine fehlende Adresse nahm den ganzen Dienst mit: %v", fehler)
	}

	t.Cleanup(func() { _ = vereint.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	if adressen := rest.Adressen(vereint); len(adressen) != 1 || !strings.HasPrefix(adressen[0], "127.0.0.1:") {
		t.Errorf("es lauscht auf %v", adressen)
	}

	if len(befunde) != 1 || befunde[0].Adresse != "192.0.2.1" || befunde[0].Grund == "" {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

func TestOhneEineEinzigeAdresseStartetNichts(t *testing.T) {
	t.Parallel()

	wahl, fehler := domain.LauschwahlAus("192.0.2.1,192.0.2.2")
	if fehler != nil {
		t.Fatal(fehler)
	}

	vereint, _, fehler := rest.LauschenAufAllen(t.Context(), wahl.Bindungen(0), &tls.Config{MinVersion: tls.VersionTLS13})
	if fehler == nil {
		_ = vereint.Close() //nolint:errcheck // es hätte keinen geben dürfen

		t.Fatal("ohne eine einzige Adresse lief der Dienst trotzdem")
	}

	if !strings.Contains(fehler.Error(), "192.0.2.1") || !strings.Contains(fehler.Error(), "192.0.2.2") {
		t.Errorf("der Fehler nennt nicht jede Adresse: %v", fehler)
	}
}

// **Jenseits der Schleife ohne TLS startet nichts** — auch nicht die
// Schleife daneben: wer eine Öffnung nach außen gewählt hat, soll nicht
// still nur die halbe bekommen.
func TestEineOeffnungOhneTLSNimmtDieGanzeWahlZurueck(t *testing.T) {
	t.Parallel()

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,192.0.2.1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	vereint, _, fehler := rest.LauschenAufAllen(t.Context(), wahl.Bindungen(0), nil)
	if !errors.Is(fehler, domain.ErrValidation) {
		if vereint != nil {
			_ = vereint.Close() //nolint:errcheck // es hätte keinen geben dürfen
		}

		t.Fatalf("ohne TLS: %v", fehler)
	}
}

// **Ein belegter Port nimmt nur seine Adresse mit** — derselbe Port auf
// `::1` ist frei, und dort läuft der Dienst.
func TestEinBelegterPortWirdBenanntUndDieAnderenLaufen(t *testing.T) {
	t.Parallel()

	belegt, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = belegt.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	adresse, istTCP := belegt.Addr().(*net.TCPAddr)
	if !istTCP {
		t.Fatalf("kein TCP-Lauscher: %T", belegt.Addr())
	}

	port := adresse.Port

	wahl, fehler := domain.LauschwahlAus("127.0.0.1,::1")
	if fehler != nil {
		t.Fatal(fehler)
	}

	vereint, befunde, fehler := rest.LauschenAufAllen(t.Context(), wahl.Bindungen(port), nil)
	if fehler != nil {
		t.Fatalf("ein belegter Port nahm den ganzen Dienst mit: %v", fehler)
	}

	t.Cleanup(func() { _ = vereint.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	if adressen := rest.Adressen(vereint); len(adressen) != 1 || !strings.HasPrefix(adressen[0], "[::1]:") {
		t.Errorf("es lauscht auf %v", adressen)
	}

	if len(befunde) != 1 || befunde[0].Adresse != "127.0.0.1" || !strings.Contains(befunde[0].Grund, "in use") {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}
