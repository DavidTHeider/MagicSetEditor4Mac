package domain_test

// **Objekte, Zonen und was zwischen ihnen gilt** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Wonach hier gefragt wird
//
// Danach, dass **zwei Karten gleichen Namens zwei Objekte bleiben** — das ist
// die vierte Abnahme des Issues und der Satz, an dem sich ein Spielkern
// entscheidet —, dass eine Zone weiß, ob man in sie hineinsieht, und dass
// eine Beziehung nicht an sich selbst hängt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Gleicher Name, gleiche Textur — zwei Objekte.**
func TestZweiKartenGleichenNamensSindZweiObjekte(t *testing.T) {
	t.Parallel()

	erste := domain.Spielobjekt{
		Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
		Zone: domain.ZoneBibliothek, Karte: "wald",
	}

	zweite := erste
	zweite.Kennung = "o2"

	if erste.Kennung == zweite.Kennung {
		t.Fatal("die Vorrichtung vergleicht dasselbe Objekt mit sich")
	}

	if erste.Gleich(zweite) {
		t.Error("zwei Objekte mit derselben Karte gelten als eines")
	}

	// **Und dasselbe Objekt bleibt dasselbe** — die Gegenprobe, ohne die
	// oben nur stünde, dass irgendetwas verschieden ist.
	if !erste.Gleich(erste) {
		t.Error("ein Objekt ist sich selbst nicht gleich")
	}
}

// **Was kein Spielobjekt ist.**
func TestWasKeinSpielobjektIst(t *testing.T) {
	t.Parallel()

	gut := domain.Spielobjekt{
		Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
		Zone: domain.ZoneHand, Karte: "wald",
	}

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("das kleinste Objekt: %v", fehler)
	}

	faelle := map[string]func(*domain.Spielobjekt){
		"ohne Kennung":     func(o *domain.Spielobjekt) { o.Kennung = " " },
		"ohne Beherrscher": func(o *domain.Spielobjekt) { o.Beherrscher = "" },
		"mit erfundener Art": func(o *domain.Spielobjekt) {
			o.Art = domain.Objektart("gespenst")
		},
		"mit erfundener Zone": func(o *domain.Spielobjekt) {
			o.Zone = domain.Zone("keller")
		},
		// **Eine Karte ohne Karte.** Ein Spielstein hat keine, eine
		// Karteninstanz schon — sonst wüsste niemand, was dort liegt.
		"als Karte ohne Karte": func(o *domain.Spielobjekt) { o.Karte = "" },
		"an sich selbst hängend": func(o *domain.Spielobjekt) {
			o.HaengtAn = o.Kennung
		},
		"mit negativem Zähler": func(o *domain.Spielobjekt) {
			o.Zaehler = map[string]int{"+1/+1": -1}
		},
		"mit namenlosem Zähler": func(o *domain.Spielobjekt) {
			o.Zaehler = map[string]int{"": 1}
		},
	}

	for name, verderben := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			objekt := gut
			verderben(&objekt)

			if fehler := objekt.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Spielstein ist eine Karte, die keine ist.**
func TestEinSpielsteinBrauchtKeineKarte(t *testing.T) {
	t.Parallel()

	stein := domain.Spielobjekt{
		Kennung: "o1", Art: domain.ObjektSpielstein, Beherrscher: "sitz-1",
		Zone: domain.ZoneTisch, Name: "Bär",
	}

	if fehler := stein.Pruefe(); fehler != nil {
		t.Fatalf("der Spielstein: %v", fehler)
	}

	// **Aber einen Namen braucht er** — sonst steht auf dem Tisch nichts.
	stein.Name = ""

	if fehler := stein.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein namenloser Spielstein: %v", fehler)
	}
}

// **In welche Zone sieht wer hinein.**
//
// **Das ist die Tabelle, aus der die Projektion später rechnet** — sie steht
// hier und nicht dort, weil sie eine Regel ist und keine Übersetzung.
func TestWerInWelcheZoneSieht(t *testing.T) {
	t.Parallel()

	faelle := map[domain.Zone]domain.Zonensicht{
		domain.ZoneBibliothek: domain.SichtVerdeckt,
		domain.ZoneHand:       domain.SichtBeherrscher,
		domain.ZoneTisch:      domain.SichtOffen,
		domain.ZoneFriedhof:   domain.SichtOffen,
		domain.ZoneExil:       domain.SichtOffen,
		domain.ZoneStapel:     domain.SichtOffen,
		domain.ZoneKommando:   domain.SichtOffen,
	}

	for zone, erwartet := range faelle {
		t.Run(string(zone), func(t *testing.T) {
			t.Parallel()

			if gilt := zone.Sicht(); gilt != erwartet {
				t.Errorf("%s ist %q statt %q", zone, gilt, erwartet)
			}
		})
	}

	if len(faelle) != len(domain.Zonen()) {
		t.Errorf("es gibt %d Zonen, geprüft sind %d",
			len(domain.Zonen()), len(faelle))
	}

	// **Eine erfundene Zone ist verdeckt** — der sichere Ausgang. Wer hier
	// „offen" zurückgäbe, zeigte bei jedem Tippfehler alles.
	if gilt := domain.Zone("keller").Sicht(); gilt != domain.SichtVerdeckt {
		t.Errorf("eine unbekannte Zone ist %q", gilt)
	}
}
