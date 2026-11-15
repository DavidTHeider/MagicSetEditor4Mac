package storage_test

// **Löschen und Kopieren einer Partie** — PLAY-03.AK-05
// (#1711).
//
// **Gegen eine echte Datenbank** (§3): die Zusagen sind welche über
// Fremdschlüssel und Kaskaden, und die hält kein Doppel im Arbeitsspeicher.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinePartieLaesstSichLoeschen(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.Loeschen(ctx, "p1"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if _, fehler := ablage.Partie(ctx, "p1"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("die Partie steht noch da: %v", fehler)
	}
}

// **Eine Partie, die es nicht gibt, ist ein Befund und kein stiller
// Erfolg** — wer zweimal löscht, hat beim zweiten Mal etwas anderes gemeint.
func TestEinePartieOhneZeileLoeschenIstEinBefund(t *testing.T) {
	ablage := partieablage(t)

	fehler := ablage.Loeschen(probekontext(t), "gibtsnicht")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Und der Ereignisstrom geht mit** — `ON DELETE CASCADE`. Ein Strom ohne
// seine Partie wäre eine Kette, an deren Anfang nichts steht.
func TestDasLoeschenNimmtDieEreignisseMit(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.Freigeben(ctx, "p1",
		domain.Offenlegung{Von: "sitz-1", An: []string{"sitz-2"}}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	if fehler := ablage.Loeschen(ctx, "p1"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	// **An einer neuen Partie mit derselben Kennung** — eine gelöschte ist
	// seit #1815 „nicht gefunden", und die Kaskade zeigt sich daran, dass die
	// neue nichts erbt.
	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("neu anlegen: %v", fehler)
	}

	zahl, fehler := ablage.Offenlegungszahl(ctx, "p1")
	if fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	if zahl != 0 {
		t.Errorf("es stehen noch %d Freigaben da", zahl)
	}
}

func TestOffenlegungszahlZaehltDieFreigaben(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	if _, fehler := ablage.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	zahl, fehler := ablage.Offenlegungszahl(ctx, "p1")
	if fehler != nil || zahl != 0 {
		t.Fatalf("frisch sind es %d (%v)", zahl, fehler)
	}

	if gabefehler := ablage.Freigeben(ctx, "p1",
		domain.Offenlegung{Von: "sitz-1", An: []string{"sitz-2"}}); gabefehler != nil {
		t.Fatalf("freigeben: %v", gabefehler)
	}

	zahl, fehler = ablage.Offenlegungszahl(ctx, "p1")
	if fehler != nil || zahl != 1 {
		t.Errorf("danach sind es %d (%v)", zahl, fehler)
	}
}

func TestEinePartieLaesstSichKopieren(t *testing.T) {
	ablage := partieablage(t)
	ctx := probekontext(t)

	vorlage := einePartie()
	vorlage.Stand = domain.StandBeendet

	if _, fehler := ablage.Anlegen(ctx, vorlage); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.Freigeben(ctx, "p1",
		domain.Offenlegung{Von: "sitz-1", An: []string{"sitz-2"}}); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	kopie, fehler := ablage.Kopieren(ctx, "p1", "p2")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if kopie.ID != "p2" || kopie.Stand != domain.StandLaeuft {
		t.Errorf("die Kopie ist %q im Stand %q", kopie.ID, kopie.Stand)
	}

	// **Ohne alte Peer-Zustimmungen**, wörtlich aus dem Kriterium.
	zahl, fehler := ablage.Offenlegungszahl(ctx, "p2")
	if fehler != nil || zahl != 0 {
		t.Errorf("die Kopie trägt %d Freigaben (%v)", zahl, fehler)
	}

	// **Und die Vorlage bleibt, wie sie war.**
	alt, fehler := ablage.Partie(ctx, "p1")
	if fehler != nil || alt.Stand != domain.StandBeendet {
		t.Errorf("die Vorlage ist %q (%v)", alt.Stand, fehler)
	}
}

// **Eine Vorlage, die es nicht gibt, gibt keine Kopie.**
func TestEinePartieOhneZeileKopierenIstEinBefund(t *testing.T) {
	ablage := partieablage(t)

	_, fehler := ablage.Kopieren(probekontext(t), "gibtsnicht", "p2")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}
