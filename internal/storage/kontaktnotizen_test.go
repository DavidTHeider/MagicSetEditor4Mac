package storage_test

// **Was dieses Profil sich zu einem Gegenüber notiert** — Werkstück `T3` von
// Phase 5.6 (#1468).
//
// # Warum die erste Begegnung eigens dasteht
//
// **Weil „seit wann kenne ich den?" eine andere Frage ist als „wann zuletzt?".**
// Das Issue verlangt beides — first und last seen —, und `gesehen` beantwortet
// nur die zweite: es wird bei jeder Begegnung überschrieben.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die erste Begegnung bleibt die erste.**
func TestDieErsteBegegnungBleibtDieErste(t *testing.T) {
	ablage := storage.NeueKontaktablage(frischeDatenbank(t))
	ctx := probekontext(t)

	frueh := freigabezeit
	spaet := freigabezeit.Add(48 * time.Hour)

	for _, wann := range []time.Time{frueh, spaet} {
		if fehler := ablage.Begegnet(
			ctx, "peer-1", "ab:cd", "Anna", wann); fehler != nil {
			t.Fatalf("begegnet: %v", fehler)
		}
	}

	kontakt, fehler := ablage.Lesen(ctx, "peer-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if !kontakt.Erstmals.Equal(frueh) {
		t.Errorf("die erste Begegnung ist %v, erwartet war %v",
			kontakt.Erstmals, frueh)
	}

	// **Und die letzte ist die letzte** — sonst prüfte der Lauf nur, dass
	// überhaupt etwas stehen bleibt (§10).
	if !kontakt.Gesehen.Equal(spaet) {
		t.Errorf("die letzte Begegnung ist %v", kontakt.Gesehen)
	}
}

// **Eine Notiz gehört diesem Profil und geht nie hinaus.**
func TestEineNotizWirdGehaltenUndGeaendert(t *testing.T) {
	ablage := storage.NeueKontaktablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.Begegnet(
		ctx, "peer-1", "ab:cd", "Anna", freigabezeit); fehler != nil {
		t.Fatalf("begegnet: %v", fehler)
	}

	for _, notiz := range []string{"tauscht fair", ""} {
		if fehler := ablage.NotizSetzen(ctx, "peer-1", notiz); fehler != nil {
			t.Fatalf("notieren: %v", fehler)
		}

		kontakt, fehler := ablage.Lesen(ctx, "peer-1")
		if fehler != nil {
			t.Fatalf("lesen: %v", fehler)
		}

		if kontakt.Notiz != notiz {
			t.Errorf("die Notiz ist %q, erwartet war %q", kontakt.Notiz, notiz)
		}
	}
}

// **Und eine Begegnung überschreibt sie nicht.**
//
// **Sonst wäre jede Notiz nach dem nächsten Hallo fort** — derselbe Grund,
// aus dem der Stand nicht im `DO UPDATE` steht.
func TestEineBegegnungUeberschreibtDieNotizNicht(t *testing.T) {
	ablage := storage.NeueKontaktablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.Begegnet(
		ctx, "peer-1", "ab:cd", "Anna", freigabezeit); fehler != nil {
		t.Fatalf("begegnet: %v", fehler)
	}

	if fehler := ablage.NotizSetzen(ctx, "peer-1", "tauscht fair"); fehler != nil {
		t.Fatalf("notieren: %v", fehler)
	}

	if fehler := ablage.Begegnet(ctx, "peer-1", "ab:cd", "Anna",
		freigabezeit.Add(time.Hour)); fehler != nil {
		t.Fatalf("noch einmal begegnet: %v", fehler)
	}

	kontakt, fehler := ablage.Lesen(ctx, "peer-1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if kontakt.Notiz != "tauscht fair" {
		t.Errorf("die Notiz ist %q", kontakt.Notiz)
	}
}

// **Eine Notiz für einen Unbekannten ist ein Nein.**
func TestEineNotizFuerEinenUnbekanntenIstEinNein(t *testing.T) {
	if fehler := storage.NeueKontaktablage(frischeDatenbank(t)).NotizSetzen(
		probekontext(t), "peer-9", "hm",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}
