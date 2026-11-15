package leitung_test

// **Was ein Horcher tut, wenn nichts klappt** — Werkstück N2 von Phase 5.4
// (#1455).
//
// **Eine dritte Datei** aus demselben Grund wie die zweite. Der Schnitt folgt
// der Frage: hier die Wege, auf denen **keine** Verbindung entsteht.
//
// # Der Satz, für den es diese Läufe gibt
//
// **„Endet mit einem benannten Fehler und nicht mit Warten."** Die Abnahme von
// #1455 sagt es wörtlich, und sie sagt es zweimal — für die unerreichbare
// Adresse und für die abgelehnte Berechtigung. Ein unbegrenzter Spinner ist
// hier der Fehler, den es zu verhindern gilt.

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// **Eine unerreichbare Adresse endet mit einem Fehler und nicht mit Warten.**
//
// **Die Abnahme verlangt es wörtlich**: „eine unerreichbare Adresse endet mit
// einem benannten Fehler und nicht mit Warten." Ein unbegrenzter Spinner ist
// hier der Fehler, den es zu verhindern gilt.
func TestEineUnerreichbareAdresseEndetMitEinemFehler(t *testing.T) {
	t.Parallel()

	hiesige, dortige := zweiIdentitaeten(t)

	ctx, abbrechen := context.WithTimeout(t.Context(), 3*time.Second)
	defer abbrechen()

	// **Ein Horcher, der sofort wieder zumacht** — damit der Port belegt
	// war und jetzt keiner mehr ist; eine erfundene Adresse führte je nach
	// Netz zu einem anderen Fehler.
	lauscher, fehler := (&net.ListenConfig{}).Listen(
		ctx, "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("Listen: %v", fehler)
	}

	adresse := lauscher.Addr().String()

	if zu := lauscher.Close(); zu != nil {
		t.Fatalf("Close: %v", zu)
	}

	_, fehler = leitung.Waehle(ctx, adresse, hiesige, dortige.Kennung())
	if fehler == nil {
		t.Fatal("die Wahl kam an einer geschlossenen Tür durch")
	}

	// **Und er heißt, was er ist** (#2491): „nicht gefunden" schickte
	// jemanden zum Neuladen einer Liste, obwohl nichts gelöscht war. Die
	// Klasse bleibt `ErrNotFound` — der Horcher liest sie als „da ist
	// niemand" (siehe [tlsfehler]).
	if !errors.Is(fehler, domain.ErrPeerNichtErreichbar) ||
		!errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("die Wahl endete als %v", fehler)
	}
}

// **Ein zweites Niederlegen wird gemeldet und nicht verschwiegen.**
func TestEinZweitesNiederlegenWirdGemeldet(t *testing.T) {
	t.Parallel()

	_, dortige := zweiIdentitaeten(t)

	horcher, fehler := leitung.Horche(t.Context(), "127.0.0.1:0", dortige)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	if fehler := horcher.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if fehler := horcher.Schliesse(); fehler == nil {
		t.Fatal("das zweite Niederlegen schwieg")
	}
}

// **Ein geschlossener Horcher gibt keine Verbindung mehr.**
func TestEinGeschlossenerHorcherGibtNichtsMehr(t *testing.T) {
	t.Parallel()

	_, dortige := zweiIdentitaeten(t)

	horcher, fehler := leitung.Horche(t.Context(), "127.0.0.1:0", dortige)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	if fehler := horcher.Schliesse(); fehler != nil {
		t.Fatalf("Schliesse: %v", fehler)
	}

	if _, fehler := horcher.Nimm(t.Context()); fehler == nil {
		t.Fatal("der geschlossene Horcher gab eine Verbindung")
	}
}

// **Eine Adresse, an der nichts horchen kann, wird gemeldet.**
func TestEineAdresseAnDerNichtsHorchenKannWirdGemeldet(t *testing.T) {
	t.Parallel()

	_, dortige := zweiIdentitaeten(t)

	if _, fehler := leitung.Horche(
		t.Context(), "kein-host:kein-port", dortige); fehler == nil {
		t.Fatal("der Horcher stand an einer unmöglichen Adresse")
	}
}

// **Ein Gegenüber, das schweigt, hält den Horcher nicht fest.**
//
// # Was hier gemessen wird und wie
//
// **Dass der Handschlag am Kontext hängt.** Es genügte sonst, eine Verbindung
// zu öffnen und nichts zu schicken; der Horcher stünde still, und der nächste
// käme nicht dran. Im Betrieb begrenzt das [leitung.Handschlagfrist] — hier
// steht ein schon abgebrochener Kontext an ihrer Stelle, denn zwanzig Sekunden
// Wartezeit prüfen dieselbe Naht und kosten zwanzig Sekunden.
func TestEinSchweigendesGegenueberHaeltDenHorcherNichtFest(t *testing.T) {
	t.Parallel()

	_, dortige := zweiIdentitaeten(t)

	horcher, fehler := leitung.Horche(t.Context(), "127.0.0.1:0", dortige)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	t.Cleanup(func() {
		_ = horcher.Schliesse() //nolint:errcheck // nur ein Lauf
	})

	// **Öffnen und schweigen** — kein einziges Byte geht hinaus.
	stumm, fehler := (&net.Dialer{}).DialContext(
		t.Context(), "tcp", horcher.Adresse())
	if fehler != nil {
		t.Fatalf("Dial: %v", fehler)
	}

	t.Cleanup(func() {
		_ = stumm.Close() //nolint:errcheck // nur ein Lauf
	})

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	fertig := make(chan error, 1)

	go func() {
		_, fehler := horcher.Nimm(ctx)
		fertig <- fehler
	}()

	select {
	case fehler := <-fertig:
		if fehler == nil {
			t.Fatal("der Handschlag kam ohne ein Byte zustande")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Horcher wartete auf ein Gegenüber, das schweigt")
	}
}
