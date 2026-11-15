package app

// **Unter welcher Adresse jemand anderes ankommt** — der Befund aus der
// Sichtprüfung vom 2026-09-14.
//
// # Was hier gemessen wird
//
// **Dass der Horchstand sagt, wohin ein Gegenüber wählen soll** — und nicht
// bloß, wo gebunden wurde. Das Fenster sagte „Erreichbar unter
// 127.0.0.1:53123", und darunter erreichte niemand diesen Rechner.

import (
	"context"
	"net"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// weiterlauscher horcht angeblich auf jeder Schnittstelle.
type weiterlauscher struct{ probelauscher }

func (w *weiterlauscher) Adresse() string { return "[::]:41234" }

func TestDerHorchstandNenntEineAdresseZumWeitergeben(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &weiterlauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	// **Über eine eingesetzte Naht und nicht über dieses Netz** (§3): was
	// `ErreichbarUnter` aus einer Schnittstellenliste macht, steht in
	// `internal/leitung` und wird dort geprüft — samt Gegenprobe. Hier geht
	// es darum, **dass** die Antwort ankommt.
	anwendung.MitErreichbarkeit(func(
		_ context.Context, horchadresse string,
	) (leitung.Erreichbarkeit, error) {
		if horchadresse != "[::]:41234" {
			t.Errorf("gefragt wurde nach %q", horchadresse)
		}

		return leitung.Erreichbarkeit{
			Namen:    []string{"mein-mac.fritz.box:41234"},
			Adressen: []string{"192.168.178.51:41234"},
		}, nil
	})

	if antwort := anwendung.HorchStarten(ctx, ":41234"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	stand := anwendung.HorchStand(ctx).Stand

	if stand.Adresse != "[::]:41234" {
		t.Errorf("die gebundene Adresse fehlt: %q", stand.Adresse)
	}

	// **Die Gegenprobe steckt in der Erwartung selbst**: `[::]:41234` darf
	// hier nicht stehen, denn genau das war der Fehler.
	if len(stand.ErreichbarAdressen) == 0 {
		t.Fatal("keine Adresse zum Weitergeben")
	}

	// **Der Name steht vor der Adresse, weil er länger gilt.** Eine
	// IP-Adresse aus einem DHCP-Lease ist eine Momentaufnahme.
	if len(stand.ErreichbarNamen) != 1 ||
		stand.ErreichbarNamen[0] != "mein-mac.fritz.box:41234" {
		t.Errorf("der Name fehlt: %v", stand.ErreichbarNamen)
	}

	for _, adresse := range stand.ErreichbarAdressen {
		wirt, _, fehler := net.SplitHostPort(adresse)
		if fehler != nil {
			t.Fatalf("%q ist keine Adresse: %v", adresse, fehler)
		}

		ip := net.ParseIP(wirt)
		if ip == nil {
			t.Errorf("%q trägt keine IP-Adresse", adresse)

			continue
		}

		if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			t.Errorf("%q nützt einem Gegenüber nichts", adresse)
		}
	}
}

// **Und wer auf der Rückschleife horcht, bekommt keine erfundene Adresse.**
func TestAufDerRueckschleifeGibtEsNichtsWeiterzugeben(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &probelauscher{}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	stand := anwendung.HorchStand(ctx).Stand

	if len(stand.ErreichbarAdressen) != 0 || len(stand.ErreichbarNamen) != 0 {
		t.Errorf("erwartet war nichts, kam: %v / %v",
			stand.ErreichbarNamen, stand.ErreichbarAdressen)
	}
}

// **Und ein Lauscher mit krummer Adresse hält nichts auf.**
//
// **Ein Befund und kein Abbruch**: der Port steht, die Verbindung trägt —
// nur die Liste zum Weitergeben bleibt leer. Wer hier abbräche, nähme
// jemandem einen laufenden Horchposten weg, weil eine Anzeige nicht ging.
func TestEineUndeutbareLauscheradresseKostetNurDieListe(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &krummerLauscher{adresse: "krumm"}, nil)

	t.Cleanup(anwendung.horchenBeenden)

	if antwort := anwendung.HorchStarten(
		ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	stand := anwendung.HorchStand(ctx).Stand

	if !stand.Laeuft {
		t.Error("der Horchposten läuft nicht mehr")
	}

	if len(stand.ErreichbarAdressen) != 0 || len(stand.ErreichbarNamen) != 0 {
		t.Errorf("erwartet war nichts, kam: %v / %v",
			stand.ErreichbarNamen, stand.ErreichbarAdressen)
	}
}

var _ service.Hereinkommend = (*weiterlauscher)(nil)
