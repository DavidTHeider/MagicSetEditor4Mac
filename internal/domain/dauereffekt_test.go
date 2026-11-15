package domain_test

// **Schichten, Zeitstempel und die Dauer eines Effekts** — Werkstück M10 von
// Phase 5.5
// (#1612).
//
// # Die drei Abnahmen des Issues, und die zweite steht hier nicht
//
//  1. Eine Kreatur unter zwei Effekten hat die Werte, die die
//     **Schichtenreihenfolge** ergibt — nicht die der Anwendungsreihenfolge.
//  2. Ein abhängiger Effekt wird nach dem angewandt, von dem er abhängt
//     (CR 613.8). **Dazu steht hier kein Lauf.** Eine Abhängigkeit entsteht
//     nur **innerhalb** einer Schicht (CR 613.8a), und keine der beiden
//     wirkenden Schichten kann eine hervorbringen — der Grund steht im Kopf
//     von `dauereffekt.go`, die Grenze in der Regelmatrix. **Aufgezählt
//     bleibt sie trotzdem**: ein Kopf, der eine Abnahme weglässt, lässt
//     niemanden mehr nachfragen, warum sie fehlt (§10).
//  3. Ein Effekt bis zum Ende des Zuges ist nach dem Aufräumschritt weg.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei Effekte in derselben Schicht gelten nach ihrem Zeitstempel** —
// CR 613.7.
//
// **Der Riese ist 3/3.** Der erste Effekt setzt ihn auf 5/5, der zweite gibt
// +1/+1 — 6/6, und zwar in dieser Reihenfolge. Andersherum wäre er 5/5, weil
// das Setzen die Erhöhung überschriebe.
func TestZweiEffekteGeltenNachIhremZeitstempel(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t)

	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7b, Objekt: "sitz-1/riese/0",
		Staerke: 5, Widerstand: 5,
	})
	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0",
		Staerke: 1, Widerstand: 1,
	})

	staerke, widerstand := partie.Werte(objektVon(t, partie, "sitz-1/riese/0"))

	if staerke.Zahl != 6 || widerstand.Zahl != 6 {
		t.Errorf("%d/%d", staerke.Zahl, widerstand.Zahl)
	}
}

// **Und die Schicht schlägt den Zeitstempel** — CR 613.1.
//
// **Derselbe Tisch, umgekehrt gestellt**: erst +1/+1, dann auf 5/5 gesetzt.
// Nach der Reihenfolge der Anwendung wäre er 5/5; nach den Schichten ist er
// 6/6, weil 7b vor 7c gilt.
func TestDieSchichtSchlaegtDenZeitstempel(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t)

	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0",
		Staerke: 1, Widerstand: 1,
	})
	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7b, Objekt: "sitz-1/riese/0",
		Staerke: 5, Widerstand: 5,
	})

	staerke, widerstand := partie.Werte(objektVon(t, partie, "sitz-1/riese/0"))

	if staerke.Zahl != 6 || widerstand.Zahl != 6 {
		t.Errorf("%d/%d", staerke.Zahl, widerstand.Zahl)
	}
}

// **Ein Effekt der sechsten Schicht verleiht eine Fähigkeit** — CR 613.1f.
//
// **Die Grenze aus `M14` fällt damit zur Hälfte**: „gelesen wird der
// gedruckte Text — ein Effekt, der ein Schlüsselwort verleiht, wirkt nicht".
func TestEinEffektDerSechstenSchichtVerleihtEineFaehigkeit(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht6, Objekt: "sitz-1/riese/0",
		Faehigkeit: domain.SchluesselFlug,
	})

	// **Und ein Effekt der siebten Schicht daneben verleiht nichts** — die
	// Frage gilt der sechsten (CR 613.1f).
	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0", Staerke: 1,
	})

	if !partie.HatFaehigkeit("sitz-1/riese/0", domain.SchluesselFlug) {
		t.Error("der Riese fliegt nicht")
	}

	if partie.HatFaehigkeit("sitz-1/riese/0", domain.SchluesselEile) {
		t.Error("der Riese hat Eile")
	}

	// **Und der gedruckte Text zählt weiter** — CR 613.1f **fügt hinzu**,
	// es ersetzt nicht. Ohne diese Zeile fände die Frage nur noch, was ein
	// Effekt verliehen hat.
	gedruckt := einSchluesseltisch(t, map[string]string{"grosser": "Trample"})

	if !gedruckt.HatFaehigkeit(
		"sitz-1/riese/0", domain.SchluesselTrampelschaden) {
		t.Error("der gedruckte Trampelschaden zählt nicht")
	}

	// **Die Gegenprobe**: ohne Effekt fliegt er nicht.
	if eineSchichtenpartie(t).HatFaehigkeit(
		"sitz-1/riese/0", domain.SchluesselFlug) {
		t.Error("der Riese fliegt ohne Effekt")
	}
}

// **Ein Effekt bis zum Ende des Zuges ist nach dem Aufräumschritt weg** —
// CR 514.2, zweite Hälfte, und CR 611.2b.
//
// **`G9` hat sie ausdrücklich offen gelassen**, und die Grenze stand seitdem
// in der Regelmatrix.
func TestEinEffektBisZumZugendeEndetImAufraeumschritt(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0",
		Staerke: 3, Widerstand: 3, Dauer: domain.DauerZugende,
	})

	staerke, _ := partie.Werte(objektVon(t, partie, "sitz-1/riese/0"))
	if staerke.Zahl != 6 {
		t.Fatalf("Stärke %d", staerke.Zahl)
	}

	danach := bisZumSchritt(t, partie, domain.SchrittVersorgung)

	if anzahl := len(danach.Dauereffekte); anzahl != 0 {
		t.Errorf("%d Effekte stehen noch", anzahl)
	}

	staerke, _ = danach.Werte(objektVon(t, danach, "sitz-1/riese/0"))
	if staerke.Zahl != 3 {
		t.Errorf("Stärke %d", staerke.Zahl)
	}
}

// **Und ein Effekt ohne Frist bleibt** — CR 611.1: „für einen festen oder
// unbestimmten Zeitraum".
func TestEinEffektOhneFristBleibt(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0",
		Staerke: 3, Widerstand: 3,
	})

	danach := bisZumSchritt(t, partie, domain.SchrittVersorgung)

	if anzahl := len(danach.Dauereffekte); anzahl != 1 {
		t.Errorf("%d Effekte stehen noch", anzahl)
	}
}

// **Ein Effekt an einem Objekt, das die Zone wechselt, fällt weg** —
// CR 400.7: es ist ein neues Objekt, und der Effekt galt dem alten.
func TestEinEffektAmZonenwechslerFaelltWeg(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0",
		Staerke: 3, Widerstand: 3,
	})

	// **Und ein Effekt an einem anderen Objekt bleibt** — die Gegenprobe:
	// CR 400.7 spricht von **diesem** Objekt (§10).
	partie = partie.MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht7c, Objekt: "sitz-2/baer/0", Staerke: 1,
	})

	danach, fehler := partie.Zerstoeren("sitz-1/riese/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if anzahl := len(danach.Dauereffekte); anzahl != 1 {
		t.Fatalf("%d Effekte stehen noch", anzahl)
	}

	if danach.Dauereffekte[0].Objekt != "sitz-2/baer/0" {
		t.Errorf("%+v", danach.Dauereffekte[0])
	}
}

// **Eine Schicht, die es nicht gibt, wird abgewiesen** — CR 613.1 (§8).
func TestEineUnbekannteSchichtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	partie := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: "achtzehn", Objekt: "sitz-1/riese/0",
	})

	if fehler := partie.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die Schicht ging durch: %v", fehler)
	}

	// **Und die zehn gehen durch** — die Gegenprobe.
	for _, schicht := range domain.Schichten() {
		eine := eineSchichtenpartie(t).MitDauereffekt(domain.Dauereffekt{
			Schicht: schicht, Objekt: "sitz-1/riese/0",
		})

		if fehler := eine.Pruefe(); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", schicht, fehler)
		}
	}
}

// eineSchichtenpartie gibt eine Partie mit einem Riesen 3/3 im Endschritt.
func eineSchichtenpartie(t *testing.T) domain.Partie {
	t.Helper()

	return bisZumSchritt(t, einKampftisch(t), domain.SchrittEnde)
}

// **Eine verliehene Flugfähigkeit wirkt im Kampf** — CR 613.1f mit
// CR 702.9b.
//
// # Warum das der Lauf ist, der zählt
//
// **Weil eine Schicht, die keine Regel fragt, nichts tut** (§10). Dass
// [Partie.HatFaehigkeit] den Effekt findet, steht oben; dass **CR 509.1a**
// ihn fragt, steht nur hier — und genau diese Naht hat in dieser Phase schon
// zweimal gefehlt.
func TestEineVerlieheneFlugfaehigkeitWirktImKampf(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t).MitDauereffekt(domain.Dauereffekt{
		Schicht: domain.Schicht6, Objekt: "sitz-1/riese/0",
		Faehigkeit: domain.SchluesselFlug,
	})

	partie, fehler := partie.AngreiferErklaeren("sitz-1", []domain.Angriff{
		{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
	})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	partie = bisZumSchritt(t, partie, domain.SchrittBlocker)

	_, fehler = partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Bär blockte den Flieger: %v", fehler)
	}

	if !contains(fehler.Error(), "702.9b") {
		t.Errorf("die Meldung nennt die Regel nicht: %v", fehler)
	}
}

// **Ein neuer Effekt bekommt einen Zeitstempel, den kein lebender trägt** —
// CR 613.7 ordnet innerhalb einer Schicht nach ihm. Dieselbe Klasse wie bei
// den Ersetzungen (`ersetzungseffekt_test.go`): mit der Länge der Liste plus
// eins bekam nach dem Aufräumschritt ein neuer Effekt den Stempel eines noch
// lebenden.
func TestEinNeuerDauereffektNachDemZugendeTraegtEinenEigenenStempel(t *testing.T) {
	t.Parallel()

	staerker := domain.Dauereffekt{Schicht: domain.Schicht7c, Objekt: "sitz-1/riese/0", Staerke: 1, Widerstand: 1}

	bisZugende := staerker
	bisZugende.Dauer = domain.DauerZugende

	partie := eineSchichtenpartie(t).MitDauereffekt(bisZugende).MitDauereffekt(staerker)
	partie = bisZumSchritt(t, partie, domain.SchrittVersorgung).MitDauereffekt(staerker)

	stempel := map[int]bool{}
	for _, effekt := range partie.Dauereffekte {
		if stempel[effekt.Zeitstempel] {
			t.Fatalf("zwei Effekte mit dem Stempel %d: %+v", effekt.Zeitstempel, partie.Dauereffekte)
		}

		stempel[effekt.Zeitstempel] = true
	}
}
