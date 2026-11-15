package domain_test

// **Ersetzung und Verhinderung** — Werkstück M11 von Phase 5.5
// (#1613).
//
// # Die drei Abnahmen des Issues
//
//  1. Wird eine Kreatur stattdessen ins Exil geschickt, erreicht sie den
//     Friedhof **nie**.
//  2. Zwei Ersetzungen auf dasselbe Ereignis: die Wahl der Reihenfolge steht
//     im Protokoll.
//  3. Verhinderter Schaden wird **nicht** markiert.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Statt in den Friedhof ins Exil** — CR 614.1a mit CR 614.6.
func TestStattInDenFriedhofInsExil(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
		Objekt: "sitz-1/riese/0",
	})

	danach, fehler := partie.Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if zone := objektVon(t, danach, "sitz-1/riese/0").Zone; zone !=
		domain.ZoneExil {
		t.Errorf("der Riese liegt in %q", zone)
	}

	// **Die Gegenprobe**: ohne Ersetzung geht er in den Friedhof.
	ohne, fehler := eineSchichtenpartie(t).Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if zone := objektVon(t, ohne, "sitz-1/riese/0").Zone; zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Riese liegt in %q", zone)
	}
}

// **Und ein Effekt für ein anderes Objekt greift nicht** — CR 614.1: er
// wartet auf **ein** Ereignis.
func TestEinErsatzFuerEinAnderesObjektGreiftNicht(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
		Objekt: "sitz-2/baer/0",
	})

	danach, fehler := partie.Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if zone := objektVon(t, danach, "sitz-1/riese/0").Zone; zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Riese liegt in %q", zone)
	}
}

// **Ein Effekt ohne Objekt gilt für jedes** — CR 614.1 lässt beides zu.
func TestEinErsatzOhneObjektGiltFuerJedes(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
	})

	danach, fehler := partie.Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if zone := objektVon(t, danach, "sitz-1/riese/0").Zone; zone !=
		domain.ZoneExil {
		t.Errorf("der Riese liegt in %q", zone)
	}
}

// **Die Gelegenheit gilt je Ereignis und nicht je Zug** — CR 614.5.
//
// # Der Befund des Zweitblicks über Phase 5.5
//
// **Hier stand das Gegenteil**: der Effekt galt als verbraucht, bis der Zug
// endete, und die zweite Kreatur ging in den Friedhof. CR 614.5 sagt aber
// „ein Ereignis **oder die veränderten Ereignisse, die es ersetzen**" — die
// Gelegenheit hängt am **Ereignis**, nicht an der Uhr. Ein Effekt „wenn
// etwas in deinen Friedhof käme, schicke es stattdessen ins Exil" greift bei
// jeder Kreatur.
func TestDieGelegenheitGiltJeEreignis(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
	})

	partie, fehler := partie.Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	danach, fehler := partie.Zerstoeren("sitz-2/baer/0")
	if fehler != nil {
		t.Fatalf("zweites zerstören: %v", fehler)
	}

	for _, kennung := range []string{"sitz-1/riese/0", "sitz-2/baer/0"} {
		if zone := objektVon(t, danach, kennung).Zone; zone !=
			domain.ZoneExil {
			t.Errorf("%q liegt in %q", kennung, zone)
		}
	}
}

// **Und nach der ersten Ersetzung geht es mit den übrigen weiter** —
// CR 616.1, zweiter Satz.
//
// **Zwei Effekte über je eins**: drei Schaden werden zu eins, nicht zu
// zwei. Der erste greift, das **veränderte** Ereignis wird noch einmal
// gegen die übrigen gehalten, und der zweite greift auch.
func TestNachDerErstenErsetzungGehtEsWeiter(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t)

	for range 2 {
		partie = partie.MitErsetzung(domain.Ersetzungseffekt{
			Ereignis: domain.EreignisSchaden,
			Ersatz:   domain.ErsetzungVerhindern,
			Objekt:   "sitz-1/riese/0", Menge: 1,
		})
	}

	danach, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 3,
	})
	if !ging {
		t.Fatal("der Schaden ging nicht")
	}

	if schaden := objektVon(t, danach, "sitz-1/riese/0").Schaden; schaden != 1 {
		t.Errorf("%d Schaden", schaden)
	}
}

// **Aber derselbe Effekt greift im selben Ereignis nur einmal** —
// CR 614.5: er ruft sich nicht selbst auf.
func TestDerselbeEffektGreiftImSelbenEreignisNurEinmal(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
		Objekt: "sitz-1/riese/0", Menge: 1,
	})

	danach, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 3,
	})
	if !ging {
		t.Fatal("der Schaden ging nicht")
	}

	// **Zwei und nicht null** — eines verhindert, und dann ist Schluss.
	if schaden := objektVon(t, danach, "sitz-1/riese/0").Schaden; schaden != 2 {
		t.Errorf("%d Schaden", schaden)
	}

	// **Und wo die Zahl genau reicht, bleibt nichts übrig** — eines gegen
	// eines. Der Fall trennt „alles verhindert" von „genug verhindert".
	genau, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 1,
	})
	if !ging {
		t.Fatal("der Schaden ging nicht")
	}

	if schaden := objektVon(t, genau, "sitz-1/riese/0").Schaden; schaden != 0 {
		t.Errorf("%d Schaden", schaden)
	}
}

// **Zwei Ersetzungen auf dasselbe Ereignis geben in jeder Reihenfolge
// dasselbe** — CR 616.1 hat heute keine sichtbare Folge.
//
// # Was hier stand und warum es falsch war
//
// **„Der ältere verhindert eins, der jüngere hätte alles verhindert."** Das
// stimmte nur, solange ein Effekt nach der ersten Ersetzung schwieg — und
// genau das hat der Zweitblick über Phase 5.5 als Verstoß gegen CR 616.1
// angestrichen. Seitdem kommen **beide** an das veränderte Ereignis, und
// die Reihenfolge ändert die Zahl nicht mehr.
//
// **Der Lauf bleibt trotzdem**: er hält fest, dass die Abweichung aus
// CR 616.1 heute **folgenlos** ist. Wer eine zweite Ersetzungsart für
// dieselbe Ereignisart baut, sieht ihn rot werden.
func TestZweiErsetzungenGebenInJederReihenfolgeDasselbe(t *testing.T) {
	t.Parallel()

	klein := domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
		Objekt: "sitz-1/riese/0", Menge: 1,
	}
	alles := domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
		Objekt: "sitz-1/riese/0",
	}

	for _, reihe := range [][]domain.Ersetzungseffekt{
		{klein, alles}, {alles, klein},
	} {
		partie := eineSchichtenpartie(t)
		for _, effekt := range reihe {
			partie = partie.MitErsetzung(effekt)
		}

		danach, ging := partie.Anwenden(domain.Spielwirkung{
			Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 2,
		})
		if !ging {
			t.Fatal("der Schaden ging nicht")
		}

		if schaden := objektVon(t, danach,
			"sitz-1/riese/0").Schaden; schaden != 0 {
			t.Errorf("%d Schaden bei %+v", schaden, reihe)
		}
	}
}

// **Verhinderter Schaden wird nicht markiert** — CR 615.1, die dritte
// Abnahme.
func TestVerhinderterSchadenWirdNichtMarkiert(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
		Objekt: "sitz-1/riese/0",
	})

	danach, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 3,
	})
	if !ging {
		t.Fatal("der Schaden ging nicht")
	}

	if schaden := objektVon(t, danach, "sitz-1/riese/0").Schaden; schaden != 0 {
		t.Errorf("%d Schaden", schaden)
	}
}

// **Und im Kampf ebenso** — CR 615.1 an der Stelle, an der der Schaden
// wirklich entsteht (CR 510.2).
func TestVerhinderterKampfschadenWirdNichtMarkiert(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, nil).MitErsetzung(
		domain.Ersetzungseffekt{
			Ereignis: domain.EreignisSchaden,
			Ersatz:   domain.ErsetzungVerhindern,
			Objekt:   "sitz-2/baer/0",
		})

	partie, fehler := partie.AngreiferErklaeren("sitz-1", []domain.Angriff{
		{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
	})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	partie = bisZumSchritt(t, partie, domain.SchrittBlocker)

	partie, fehler = partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
	})
	if fehler != nil {
		t.Fatalf("blocken: %v", fehler)
	}

	danach := bisZumSchritt(t, partie, domain.SchrittKampfende)

	// **Der Bär lebt** — drei verhinderte Schadenspunkte sind keine.
	if zone := objektVon(t, danach, "sitz-2/baer/0").Zone; zone !=
		domain.ZoneTisch {
		t.Errorf("der Bär liegt in %q", zone)
	}
}

// **Ein Effekt bis zum Zugende endet im Aufräumschritt** — CR 514.2.
func TestErsetzungenBisZugendeEndenImAufraeumschritt(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t)

	partie = partie.MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
		Dauer: domain.DauerZugende,
	})
	partie = partie.MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
	})

	danach := bisZumSchritt(t, partie, domain.SchrittVersorgung)

	if anzahl := len(danach.Ersetzungen); anzahl != 1 {
		t.Fatalf("%d Ersetzungen stehen noch: %+v",
			anzahl, danach.Ersetzungen)
	}

	if danach.Ersetzungen[0].Ereignis != domain.EreignisSchaden {
		t.Errorf("%+v", danach.Ersetzungen[0])
	}
}

// **Ein neuer Effekt bekommt einen Zeitstempel, den kein lebender trägt** —
// CR 616.1 ordnet nach ihm, und die Gelegenheit aus CR 614.5 merkt sich
// Effekte an ihm. Bis zur Zufallsmutation (B05, Nachlauf über Phase B) war
// der Stempel die Länge der Liste plus eins: endete ein Effekt im
// Aufräumschritt, bekam der nächste den Stempel eines noch lebenden, und
// einer der beiden verhinderte nichts.
func TestEinNeuerEffektNachDemZugendeTraegtEinenEigenenStempel(t *testing.T) {
	t.Parallel()

	verhindern := domain.Ersetzungseffekt{
		Ereignis: domain.EreignisSchaden, Ersatz: domain.ErsetzungVerhindern,
		Objekt: "sitz-1/riese/0", Menge: 1,
	}

	partie := eineSchichtenpartie(t)
	partie = partie.MitErsetzung(domain.Ersetzungseffekt{
		Ereignis: domain.EreignisFriedhof, Ersatz: domain.ErsetzungExil,
		Dauer: domain.DauerZugende,
	})
	partie = partie.MitErsetzung(verhindern)

	partie = bisZumSchritt(t, partie, domain.SchrittVersorgung).MitErsetzung(verhindern)

	stempel := map[int]bool{}
	for _, effekt := range partie.Ersetzungen {
		if stempel[effekt.Zeitstempel] {
			t.Fatalf("zwei Effekte mit dem Stempel %d: %+v", effekt.Zeitstempel, partie.Ersetzungen)
		}

		stempel[effekt.Zeitstempel] = true
	}

	danach, ging := partie.Anwenden(domain.Spielwirkung{
		Art: domain.WirkungSchaden, Objekt: "sitz-1/riese/0", Menge: 2,
	})
	if !ging {
		t.Fatal("der Schaden ging nicht")
	}

	if schaden := objektVon(t, danach, "sitz-1/riese/0").Schaden; schaden != 0 {
		t.Errorf("%d Schaden — beide Effekte hätten je einen verhindert", schaden)
	}
}
