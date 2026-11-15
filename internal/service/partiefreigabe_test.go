package service_test

// **Welcher Sitz einem Gegenüber gehört** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die Frage, die vorher niemand beantwortet hat
//
// **Eine Sitzung belegt einen Rechner, keinen Sitz.** Der Handschlag prüft die
// Peerkennung; welcher Sitz das ist, steht nirgends — und bis dahin nahm der
// Wirt es aus der Nachricht, also aus der Behauptung des Gastes.
//
// **Hier wird es entschieden, und zwar vom Wirt**: er gibt eine Partie frei
// und nennt dabei die Sitze, die zu haben sind. Wer sich meldet, bekommt den
// nächsten — **und beim zweiten Mal denselben**.
//
// # Warum „denselben" die eigentliche Zusage ist
//
// **Weil eine Leitung abreißt.** Wer nach einem Abriss wiederkommt, ist
// derselbe Spieler und kein neuer; bekäme er einen zweiten Sitz, säße er
// plötzlich am anderen Ende des Tisches und die Partie hätte einen
// Teilnehmer zu viel.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einGegenueber gibt eine Peerkennung, die sich von anderen unterscheidet.
func einGegenueber(t *testing.T) peer.Kennung {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich.Kennung()
}

// mussSitz holt einen Sitz und bricht ab, wenn keiner kommt.
//
// **Und der Fehler wird gelesen, nicht verworfen**: ein `_` machte aus einem
// leeren Sitz eine Zusicherung über nichts (§10).
func mussSitz(
	t *testing.T, freigabe *service.Partiefreigabe, gast peer.Kennung,
) string {
	t.Helper()

	sitz, fehler := freigabe.Sitz(gast)
	if fehler != nil {
		t.Fatalf("Sitz: %v", fehler)
	}

	return sitz
}

// **Wer sich zuerst meldet, bekommt den ersten Sitz.**
func TestDerErsteGastBekommtDenErstenSitz(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bert", "cora"})

	sitz, fehler := freigabe.Sitz(einGegenueber(t))
	if fehler != nil {
		t.Fatalf("Sitz: %v", fehler)
	}

	if sitz != "bert" {
		t.Errorf("Sitz %q, erwartet den ersten", sitz)
	}
}

// **Zwei Gäste bekommen zwei Sitze.**
func TestZweiGaesteBekommenZweiSitze(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bert", "cora"})

	erster := mussSitz(t, freigabe, einGegenueber(t))

	zweiter, fehler := freigabe.Sitz(einGegenueber(t))
	if fehler != nil {
		t.Fatalf("Sitz: %v", fehler)
	}

	if erster == zweiter {
		t.Errorf("beide bekamen %q", erster)
	}
}

// **Derselbe Gast bekommt denselben Sitz — auch nach einem Abriss.**
//
// **Die Zusage, für die es diese Rechnung gibt.** Wer wiederkommt, ist
// derselbe Spieler; ein zweiter Sitz machte aus ihm einen zweiten Teilnehmer.
func TestDerselbeGastBekommtDenselbenSitz(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bert", "cora"})
	gast := einGegenueber(t)

	erster := mussSitz(t, freigabe, gast)

	wieder, fehler := freigabe.Sitz(gast)
	if fehler != nil {
		t.Fatalf("Sitz beim zweiten Mal: %v", fehler)
	}

	if wieder != erster {
		t.Errorf("beim zweiten Mal %q statt %q", wieder, erster)
	}

	// **Und der zweite Sitz ist noch zu haben.**
	dritter, fehler := freigabe.Sitz(einGegenueber(t))
	if fehler != nil || dritter == erster {
		t.Errorf("der nächste bekam %q (%v)", dritter, fehler)
	}
}

// **Ist kein Sitz mehr frei, wird abgewiesen.**
//
// **Und zwar als Konflikt und nicht als Validierungsfehler**: der Gast hat
// nichts falsch gemacht, der Tisch ist voll.
func TestEinVollerTischWeistAb(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bert"})

	if _, fehler := freigabe.Sitz(einGegenueber(t)); fehler != nil {
		t.Fatalf("der erste: %v", fehler)
	}

	if _, fehler := freigabe.Sitz(
		einGegenueber(t)); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Freigabe ohne Sitz ist keine.**
//
// **Sie wird beim Bauen nicht abgewiesen, sondern beim Fragen** — wer sie
// anlegt, hat vielleicht gerade noch keinen freien; wer fragt, braucht einen.
func TestEineFreigabeOhneSitzeWeistJedenAb(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", nil)

	if _, fehler := freigabe.Sitz(
		einGegenueber(t)); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Die Partie steht dabei** — ohne sie wüsste der Wirt nicht, welche er führt.
func TestEineFreigabeNenntIhrePartie(t *testing.T) {
	t.Parallel()

	if partie := service.NeuePartiefreigabe("p1", nil).Partie(); partie != "p1" {
		t.Errorf("Partie %q", partie)
	}
}

// **Und der vergebene Stand ist ablesbar** — das Fenster zeigt, wer sitzt.
func TestDerVergebeneStandIstAblesbar(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bert", "cora"})
	gast := einGegenueber(t)

	sitz := mussSitz(t, freigabe, gast)

	vergeben := freigabe.Vergeben()

	switch {
	case len(vergeben) != 1:
		t.Fatalf("%d vergeben, erwartet 1", len(vergeben))
	case vergeben[0].Sitz != sitz:
		t.Errorf("vergeben ist %q, bekommen hat er %q", vergeben[0].Sitz, sitz)
	case vergeben[0].Gegenueber != gast:
		t.Errorf("das Gegenüber stimmt nicht")
	}
}
