package domain_test

// **Die beiden Gegnerquoten** — Werkstück 2 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**, damit `rangliste_test.go` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, was **eine** Bilanz
// ergibt, hier, was sich aus **mehreren** ergibt.
//
// # Was hier nachgelesen ist und was nicht
//
// **Nachgelesen** sind die Gegnerbilanzen: Appendix C nennt für sein Beispiel
// wörtlich „4–4–0, 7–1–0, 1–3–1, 3–3–1, 6–2–0, 5–2–1, 4–3–1, and 6–1–1".
//
// **Nicht nachgelesen** ist das Ergebnis: das Dokument zeigt es als **Bild**
// einer Formel, und aus einem Bild lässt sich keine Zahl lesen. Die erwarteten
// Werte hier sind deshalb **aus der dokumentierten Formel gerechnet** — mit
// den Zahlen des Dokuments, aber von diesem Test. Das ist ein Unterschied, und
// er gehört genannt (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bilanzAus baut eine Bilanz aus „Siege–Niederlagen–Unentschieden".
//
// **Über Matches und nicht über Spiele**: Appendix C nennt für dieses Beispiel
// nur Matchbilanzen, und die Matchquote braucht nichts anderes.
func bilanzAus(siege, niederlagen, unentschieden int) domain.Bilanz {
	return domain.Bilanz{
		Siege: siege, Niederlagen: niederlagen, Unentschieden: unentschieden,
	}
}

// **Die Gegnermatchquote ist das Mittel der Matchquoten aller Gegner.**
//
// Wörtlich aus Appendix C: „A player's opponents' match-win percentage is the
// average match-win percentage of each opponent that player faced (ignoring
// those rounds for which the player received a bye)."
func TestDieGegnermatchquoteMitteltDieGegner(t *testing.T) {
	t.Parallel()

	// Die acht Gegner aus dem Beispiel des Dokuments.
	gegner := []domain.Bilanz{
		bilanzAus(4, 4, 0), // 12/24 = 0,5
		bilanzAus(7, 1, 0), // 21/24 = 0,875
		bilanzAus(1, 3, 1), // 4/15 = 0,267 → 0,33
		bilanzAus(3, 3, 1), // 10/21 = 0,476…
		bilanzAus(6, 2, 0), // 18/24 = 0,75
		bilanzAus(5, 2, 1), // 16/24 = 0,667…
		bilanzAus(4, 3, 1), // 13/24 = 0,541…
		bilanzAus(6, 1, 1), // 19/24 = 0,791…
	}

	// **Gerechnet, nicht abgelesen** — siehe der Kopf dieser Datei.
	nahe(t, "Gegnermatchquote", domain.Gegnerquote(gegner, domain.Bilanz.Matchquote),
		0.616399)
}

// **Eine Runde mit Bye zählt nicht als Gegner.**
//
// Dasselbe Beispiel im Dokument, zweiter Spieler: „His opponents' records
// were: bye, 7–1–0, …" — sieben Gegner statt acht, und der Erste fällt weg.
func TestEinByeIstKeinGegner(t *testing.T) {
	t.Parallel()

	ohneDenErsten := []domain.Bilanz{
		bilanzAus(7, 1, 0),
		bilanzAus(1, 3, 1),
		bilanzAus(3, 3, 1),
		bilanzAus(6, 2, 0),
		bilanzAus(5, 2, 1),
		bilanzAus(4, 3, 1),
		bilanzAus(6, 1, 1),
	}

	nahe(t, "Gegnermatchquote ohne den Bye-Gegner",
		domain.Gegnerquote(ohneDenErsten, domain.Bilanz.Matchquote), 0.633027)
}

// **Die Untergrenze gilt je Gegner, nicht am Ende.**
//
// # Der Fehler, den dieser Test fängt
//
// Wer erst mittelt und dann die Untergrenze anwendet, bekommt etwas anderes:
// zwei Gegner mit 0,25 und 0,75 ergeben gemittelt 0,50 — nach Appendix C
// aber (0,33 + 0,75)/2 = **0,54**. Die Grenze steht in der Definition der
// **einzelnen** Quote, nicht in der des Mittels.
func TestDieUntergrenzeGiltJeGegner(t *testing.T) {
	t.Parallel()

	gegner := []domain.Bilanz{
		bilanzAus(1, 3, 0), // 3/12 = 0,25 → 0,33
		bilanzAus(3, 1, 0), // 9/12 = 0,75
	}

	nahe(t, "Gegnermatchquote", domain.Gegnerquote(gegner, domain.Bilanz.Matchquote),
		(domain.Quotenuntergrenze+0.75)/2)
}

// **Ohne Gegner gibt es keine Gegnerquote — und nicht 0,33.**
//
// Dieselbe eigene Entscheidung wie bei der leeren Bilanz: die MTR kennt den
// Fall nicht, weil dort jeder spielt.
func TestOhneGegnerGibtEsKeineGegnerquote(t *testing.T) {
	t.Parallel()

	if ist := domain.Gegnerquote(nil, domain.Bilanz.Matchquote); ist != 0 {
		t.Errorf("Gegnerquote ohne Gegner = %v, erwartet 0", ist)
	}
}

// **Und sie gilt für die Spielquote genauso.**
//
// Wörtlich: „Similar to opponents' match-win percentage, a player's opponents'
// game-win percentage is simply the average game-win percentage of all that
// player's opponents. And, as with opponents' match-win percentage, each
// opponent has a minimum game-win percentage of 0.33."
func TestDieGegnerspielquoteMitteltDieSpielquoten(t *testing.T) {
	t.Parallel()

	var stark, schwach domain.Bilanz

	stark.Spiele(2, 0, 0)   // 6 von 6
	schwach.Spiele(0, 2, 0) // 0 von 6 → 0,33

	nahe(t, "Gegnerspielquote",
		domain.Gegnerquote([]domain.Bilanz{stark, schwach}, domain.Bilanz.Spielquote),
		(1.0+domain.Quotenuntergrenze)/2)
}
