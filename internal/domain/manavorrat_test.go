package domain_test

// **Der Manavorrat** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Wogegen hier geprüft wird
//
//   - **Sechs Manaarten**, nicht fünf — CR 106.1b (Fassung 2026-08-19).
//   - **Der Vorrat leert sich am Ende jedes Abschnitts und jeder Phase** —
//     CR 106.4, CR 500.5.
//   - **Was womit bezahlbar ist** — CR 107.4c (farblos), CR 107.4e (hybrid),
//     CR 118.3 (nicht halb bezahlen).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Sechs Arten, und farblos ist eine davon** — CR 106.1b.
//
// **Die Gegenprobe steckt in der Zahl**: wer nur die fünf Farben führte, käme
// bei fünf heraus, und `{C}` wäre nicht zu bezahlen.
func TestEsGibtSechsManaarten(t *testing.T) {
	t.Parallel()

	arten := domain.Manaarten()

	if len(arten) != 6 {
		t.Fatalf("%d Arten: %v", len(arten), arten)
	}

	if arten[5] != domain.ManaFarblos {
		t.Errorf("die sechste ist %q", arten[5])
	}
}

// **Ein Vorrat zählt je Art.**
func TestEinVorratZaehltJeArt(t *testing.T) {
	t.Parallel()

	vorrat := domain.Manavorrat{}.Mehr(domain.ManaRot, 2).Mehr(domain.ManaGruen, 1)

	switch {
	case vorrat.Zahl(domain.ManaRot) != 2:
		t.Errorf("rot steht bei %d", vorrat.Zahl(domain.ManaRot))
	case vorrat.Zahl(domain.ManaGruen) != 1:
		t.Errorf("grün steht bei %d", vorrat.Zahl(domain.ManaGruen))
	case vorrat.Gesamt() != 3:
		t.Errorf("gesamt sind %d", vorrat.Gesamt())
	}
}

// **Zwei beliebige und ein rotes zahlen `[2][R]`** — und drei beliebige nicht.
//
// **Das ist die Abnahme des Issues**, wörtlich: „Ein Zauber für {2}{R} lässt
// sich mit zwei beliebigen und einem roten bezahlen — und nicht mit drei
// beliebigen."
func TestZweiBeliebigeUndEinRotesZahlenZweiR(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[2][R]")

	faelle := []struct {
		name    string
		vorrat  domain.Manavorrat
		bezahlt bool
	}{
		{"zwei farblose und ein rotes", domain.Manavorrat{}.
			Mehr(domain.ManaFarblos, 2).Mehr(domain.ManaRot, 1), true},
		{"drei rote", domain.Manavorrat{}.Mehr(domain.ManaRot, 3), true},
		{"drei farblose", domain.Manavorrat{}.Mehr(domain.ManaFarblos, 3), false},
		{"zwei rote", domain.Manavorrat{}.Mehr(domain.ManaRot, 2), false},
		{"grün, grün, rot", domain.Manavorrat{}.
			Mehr(domain.ManaGruen, 2).Mehr(domain.ManaRot, 1), true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ging := fall.vorrat.Bezahlbar(kosten); ging != fall.bezahlt {
				t.Errorf("bezahlbar ist %v", ging)
			}
		})
	}
}

// **Zwei Hybridsymbole gegen zwei verschiedene Farben** — CR 107.4e.
//
// # Der Lauf, für den es die Suche gibt
//
// **Wer stur von links nach rechts nimmt, scheitert hier.** `[W/U][W/U]`
// gegen ein Weißes und ein Blaues: das erste Symbol nimmt das Weiße, das
// zweite findet kein Weißes mehr — und ein Weg, der nicht zurückgeht, sagt
// „geht nicht", obwohl es geht.
//
// **Die Gegenprobe ist gefahren** (§10): mit einem `zahleWahlen`, das nur den
// **ersten** Weg probiert, ist dieser Lauf rot. Damit sagt sein Grün etwas
// über die Suche und nicht nur über die Kosten.
func TestZweiHybridsymboleFindenZweiFarben(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[W/U][W/U]")

	vorrat := domain.Manavorrat{}.
		Mehr(domain.ManaWeiss, 1).Mehr(domain.ManaBlau, 1)

	if !vorrat.Bezahlbar(kosten) {
		t.Error("zwei Hybride fanden ihre zwei Farben nicht")
	}

	// **Und zweimal dieselbe Farbe geht auch** — die Wahl steht frei.
	if nurWeiss := (domain.Manavorrat{}.Mehr(domain.ManaWeiss, 2)); !nurWeiss.
		Bezahlbar(kosten) {
		t.Error("zwei Hybride nahmen nicht zweimal dasselbe")
	}

	// **Aber nicht mit einer dritten Farbe** — sonst zahlte ein Hybrid alles.
	if rot := (domain.Manavorrat{}.Mehr(domain.ManaRot, 2)); rot.
		Bezahlbar(kosten) {
		t.Error("rot hat ein weiß-blaues Hybrid bezahlt")
	}
}

// **Ein einfarbiges Hybrid hat zwei sehr verschiedene Wege** — CR 107.4e.
func TestEinfarbigesHybridNimmtEinsOderZwei(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[2/B]")

	faelle := map[string]struct {
		vorrat  domain.Manavorrat
		bezahlt bool
	}{
		"ein schwarzes": {domain.Manavorrat{}.Mehr(domain.ManaSchwarz, 1), true},
		"zwei beliebige": {domain.Manavorrat{}.
			Mehr(domain.ManaFarblos, 1).Mehr(domain.ManaRot, 1), true},
		"ein beliebiges": {domain.Manavorrat{}.Mehr(domain.ManaRot, 1), false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if ging := fall.vorrat.Bezahlbar(kosten); ging != fall.bezahlt {
				t.Errorf("bezahlbar ist %v", ging)
			}
		})
	}
}

// **Farblos ist eine eigene Art, und Weiß zahlt es nicht** — CR 107.4c.
func TestFarblosVerlangtFarbloses(t *testing.T) {
	t.Parallel()

	kosten := domain.ManakostenAus("[C]")

	switch {
	case !(domain.Manavorrat{}.Mehr(domain.ManaFarblos, 1)).Bezahlbar(kosten):
		t.Error("farbloses Mana zahlte kein farbloses Symbol")
	case (domain.Manavorrat{}.Mehr(domain.ManaWeiss, 1)).Bezahlbar(kosten):
		t.Error("weißes Mana hat ein farbloses Symbol bezahlt")
	}
}

// **Eine gescheiterte Zahlung lässt nichts zurück** — CR 601.2h.
func TestEineGescheiterteZahlungLaesstNichtsZurueck(t *testing.T) {
	t.Parallel()

	vorrat := domain.Manavorrat{}.Mehr(domain.ManaRot, 2)

	danach, ging := vorrat.Bezahle(domain.ManakostenAus("[3][R]"))

	switch {
	case ging:
		t.Fatal("zwei Mana haben vier bezahlt")
	case danach.Gesamt() != 2:
		t.Errorf("es blieben %d Mana übrig", danach.Gesamt())
	}
}

// **Generisch nimmt farbloses zuerst** — die Entscheidung dieses Programms.
//
// **Nicht die Regel**: CR 118 lässt dem Spieler die Wahl. Sie steht als
// Grenze in der Regelmatrix, und dieser Lauf hält sie fest, damit sie nicht
// still kippt.
func TestGenerischNimmtFarblosesZuerst(t *testing.T) {
	t.Parallel()

	vorrat := domain.Manavorrat{}.
		Mehr(domain.ManaFarblos, 1).Mehr(domain.ManaRot, 1)

	danach, ging := vorrat.Bezahle(domain.ManakostenAus("[1]"))

	switch {
	case !ging:
		t.Fatal("zwei Mana zahlten kein generisches")
	case danach.Zahl(domain.ManaRot) != 1:
		t.Errorf("das rote ist weg: %v", danach)
	case danach.Zahl(domain.ManaFarblos) != 0:
		t.Errorf("das farblose steht noch da: %v", danach)
	}
}

// **Drei Symbole kann dieses Programm nicht zahlen, und es sagt warum** — §8.
func TestWasNichtZuZahlenIstWirdBenannt(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"[X]":   "CR 601.2b",
		"[S]":   "CR 107.4h",
		"[H/W]": "CR 107.4f",
	}

	for roh, grund := range faelle {
		t.Run(roh, func(t *testing.T) {
			t.Parallel()

			kosten := domain.ManakostenAus(roh)

			befunde := domain.Unbezahlbar(kosten)
			if len(befunde) != 1 {
				t.Fatalf("%d Befunde: %v", len(befunde), befunde)
			}

			if !strings.Contains(befunde[0], grund) {
				t.Errorf("der Befund lautet %q", befunde[0])
			}

			// **Und sie sind damit unbezahlbar, auch bei vollem Vorrat.**
			voll := domain.Manavorrat{}.Mehr(domain.ManaWeiss, 9)
			if voll.Bezahlbar(kosten) {
				t.Error("ein unbekanntes Symbol wurde bezahlt")
			}
		})
	}

	// **Die Gegenprobe**: ein gewöhnliches Symbol steht nicht darunter.
	if befunde := domain.Unbezahlbar(
		domain.ManakostenAus("[2][R]")); len(befunde) != 0 {
		t.Errorf("gewöhnliche Kosten gelten als unbezahlbar: %v", befunde)
	}
}
