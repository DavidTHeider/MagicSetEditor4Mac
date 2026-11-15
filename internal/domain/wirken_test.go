package domain_test

// **Einen Zauber wirken** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Wogegen hier geprüft wird
//
//   - **Der Vorgang hat Schritte, und erst am Ende ist etwas gewirkt** —
//     CR 601.2a–i (Fassung 2026-08-19).
//   - **Eine abgebrochene Ankündigung hinterlässt keinen halben Stand** —
//     CR 733: weder Mana weg noch Karte auf dem Stapel.
//   - **Ziele werden beim Ankündigen genannt** (CR 115.1) und **beim
//     Verrechnen noch einmal geprüft** (CR 608.2b).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineWirkpartie gibt eine Partie mit einem Blitz auf der Hand.
func eineWirkpartie(t *testing.T, kosten string) domain.Partie {
	t.Helper()

	partie := einePartie().ZugBeginnen("sitz-1")

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "blitz", Name: "Blitz", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Instant", "casting cost": kosten,
			}}},
		}},
	}
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/blitz/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "blitz",
	})

	return partie
}

// **Ein Zauber wandert auf den Stapel, und sein Mana geht weg** — CR 601.2a,
// CR 601.2h.
func TestEinZauberWandertAufDenStapelUndKostet(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0", nil)
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	oben := danach.StapelOben()

	switch {
	case oben == nil:
		t.Fatal("der Stapel ist leer")
	case oben.Kennung != "sitz-1/blitz/0":
		t.Errorf("oben liegt %q", oben.Kennung)
	case !danach.Sitze[0].Manavorrat.Leer():
		t.Errorf("der Vorrat steht bei %v", danach.Sitze[0].Manavorrat)
	}
}

// **Wer nicht bezahlen kann, wirkt nicht — und verliert nichts** — CR 601.2h,
// CR 733.
//
// **Das ist die Abnahme des Issues**, wörtlich: „Eine abgebrochene
// Ankündigung hinterlässt **keinen** halben Stand — nachgewiesen an Mana,
// Hand und Stapel."
func TestEineAbgebrocheneAnkuendigungLaesstNichtsZurueck(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[2][R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0", nil)

	if fehler == nil {
		t.Fatal("ein Zauber wurde ohne Mana gewirkt")
	}

	auf := danach.Objekte[len(danach.Objekte)-1]

	switch {
	case danach.StapelOben() != nil:
		t.Error("auf dem Stapel liegt etwas")
	case auf.Zone != domain.ZoneHand:
		t.Errorf("die Karte liegt in %q statt auf der Hand", auf.Zone)
	case danach.Sitze[0].Manavorrat.Zahl(domain.ManaRot) != 1:
		t.Errorf("das Mana ist weg: %v", danach.Sitze[0].Manavorrat)
	}
}

// **Wer nicht an der Reihe ist, wirkt nicht** — CR 117.1.
func TestOhnePrioritaetWirdNichtGewirkt(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	if _, fehler = partie.Wirken("sitz-2", "sitz-1/blitz/0", nil); fehler == nil {
		t.Error("ein Sitz ohne Priorität hat gewirkt")
	}
}

// **Ein Ziel wird beim Ankündigen genannt** — CR 115.1, CR 601.2c.
func TestEinZielWirdBeimAnkuendigenGenannt(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	// **Ein Ziel, das es gibt** — der Sitz gegenüber.
	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2"})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	oben := danach.StapelOben()

	switch {
	case oben == nil || len(oben.Ziele) != 1:
		t.Fatalf("die Ziele sind %v", oben)
	case oben.Ziele[0].Kennung != "sitz-2":
		t.Errorf("das Ziel heißt %q", oben.Ziele[0].Kennung)

	// **Ein Sitz hat keine Zone** — der Grund steht bei [domain.Ziel].
	case oben.Ziele[0].Zone != "":
		t.Errorf("der Sitz trägt die Zone %q", oben.Ziele[0].Zone)
	}
}

// **Ein Ziel, das es nicht gibt, kommt nicht durch** — CR 115.1.
func TestEinZielDasEsNichtGibtKommtNichtDurch(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	_, fehler = partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"gibtsnicht"})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 115") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Dasselbe Ziel zweimal geht nicht** — CR 115.3.
func TestDasselbeZielZweimalGehtNicht(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	_, fehler = partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2", "sitz-2"})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 115.3") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Zauber, dessen Ziele alle unzulässig wurden, verpufft** — CR 608.2b.
//
// **Und er verpufft in den Friedhof**, nicht ins Nichts: CR 608.2b sagt, dass
// er nicht verrechnet wird **und** dass er in den Friedhof seines Eigners
// geht — beides steht in derselben Regel.
func TestEinZauberOhneZulaessigesZielVerpufft(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	// **Ein Objekt als Ziel, damit es verschwinden kann** — ein Sitz
	// verschwindet nicht so leicht.
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2/baer/0"})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Das Ziel wechselt die Zone** — damit ist es unzulässig (CR 608.2b).
	for stelle := range danach.Objekte {
		if danach.Objekte[stelle].Kennung == "sitz-2/baer/0" {
			danach.Objekte[stelle].Zone = domain.ZoneFriedhof
		}
	}

	nachher := danach.Verrechnen()

	blitz := nachher.Objekte[len(nachher.Objekte)-2]

	switch {
	case nachher.StapelOben() != nil:
		t.Error("der Zauber liegt noch auf dem Stapel")
	case blitz.Kennung != "sitz-1/blitz/0":
		t.Fatalf("an der Stelle liegt %q", blitz.Kennung)
	case blitz.Zone != domain.ZoneFriedhof:
		t.Errorf("der verpuffte Zauber liegt in %q", blitz.Zone)
	}
}

// **Der zweite Zauber liegt über dem ersten** — CR 405.2.
//
// # Der Befund, für den dieser Lauf dasteht
//
// **Gefunden beim Bauen von M4, 2026-09-17**: `Wirken` legte die Karte auf
// den Stapel, ohne ihre **Ordnung** zu setzen. Zwei Zauber lagen damit beide
// bei null, und [Partie.StapelOben] nahm den, der zufällig zuerst in der
// Objektreihe stand — bei zwei nacheinander gewirkten Zaubern also den
// **ersten**. Der Stapel ist aber „last in, first out" (CR 405.2), und genau
// das ist sein Sinn: wer antwortet, wird zuerst verrechnet.
//
// **Der Lauf zu M1 fand es nicht**, weil er den Stapel von Hand gefüllt hat.
// Erst der Weg über das Wirken zeigt es (§3).
func TestDerZweiteZauberLiegtUeberDemErsten(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 2)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	// **Eine zweite Karte derselben Art** — sie liegt auf derselben Hand.
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/blitz/1", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "blitz",
	})

	partie, fehler = partie.Wirken("sitz-1", "sitz-1/blitz/0", nil)
	if fehler != nil {
		t.Fatalf("erster Zauber: %v", fehler)
	}

	partie, fehler = partie.Wirken("sitz-1", "sitz-1/blitz/1", nil)
	if fehler != nil {
		t.Fatalf("zweiter Zauber: %v", fehler)
	}

	if oben := partie.StapelOben(); oben == nil ||
		oben.Kennung != "sitz-1/blitz/1" {
		t.Fatalf("oben liegt %v", oben)
	}

	// **Und nach dem Verrechnen kommt der erste** — die Gegenprobe (§10):
	// ohne sie sagte der Lauf nur, dass **irgendeiner** oben liegt.
	danach := partie.Verrechnen()

	if oben := danach.StapelOben(); oben == nil ||
		oben.Kennung != "sitz-1/blitz/0" {
		t.Errorf("danach liegt oben %v", oben)
	}
}
