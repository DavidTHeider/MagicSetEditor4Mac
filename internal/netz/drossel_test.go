package netz_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die ausgehende Drossel wartet, sie weist nicht ab**
// (#1002).
//
// # Warum sie das Gegenteil der in `internal/rest` tut
//
// Dort kommt eine Anfrage **herein**, und wer zu viele schickt, soll ein 429
// bekommen — abweisen ist die Antwort. Hier geht eine Anfrage **hinaus**, und
// der Aufrufer ist das eigene Programm: es soll langsamer werden, nicht
// scheitern. Ein Abweisen machte aus einer Grenze der Gegenstelle einen
// Fehler bei uns.
//
// # Warum die Uhr eingesetzt wird
//
// §16: Klient, Uhr und Zwischenspeicher werden eingesetzt, nicht erzeugt.
// Sonst wartete jeder Test wirklich, und ein Lauf, der auf Uhren wartet,
// flackert unter Last (§3).
func TestDieDrosselWartetStattAbzuweisen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	var gewartet []time.Duration

	drossel := netz.NeueDrossel(500*time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) { gewartet = append(gewartet, dauer); jetzt = jetzt.Add(dauer) })

	// Die erste Anfrage wartet nicht.
	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	if len(gewartet) != 0 {
		t.Errorf("die erste Anfrage hat gewartet: %v", gewartet)
	}

	// Die zweite unmittelbar danach wartet den vollen Abstand.
	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die zweite Anfrage: %v", fehler)
	}

	if len(gewartet) != 1 || gewartet[0] != 500*time.Millisecond {
		t.Errorf("gewartet wurde %v, erwartet war einmal 500ms", gewartet)
	}
}

// **Wer lange genug gewartet hat, wartet nicht noch einmal.**
//
// Ohne diesen Fall wäre auch eine Drossel grün, die vor **jeder** Anfrage den
// vollen Abstand wartet — und die wäre bei zwanzig Aufrufen zehn Sekunden
// langsamer als nötig.
func TestWerLangeGenugWarteteWartetNicht(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	gewartet := 0

	drossel := netz.NeueDrossel(500*time.Millisecond,
		func() time.Time { return jetzt },
		func(time.Duration) { gewartet++ })

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	jetzt = jetzt.Add(2 * time.Second)

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die zweite Anfrage: %v", fehler)
	}

	if gewartet != 0 {
		t.Errorf("es wurde %dmal gewartet, obwohl der Abstand längst um war", gewartet)
	}
}

// **Und ein Abbruch bricht das Warten ab.**
//
// Eine Drossel, die auch nach einem `⌃C` noch eine halbe Sekunde stillsteht,
// hält den ganzen Vorgang auf — und der Aufrufer hat gesagt, dass er nicht
// mehr will.
func TestEinAbbruchBrichtDasWartenAb(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	drossel := netz.NeueDrossel(time.Minute,
		func() time.Time { return jetzt }, nil)

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	// **Mit einer echten Uhr**, denn geprüft wird, dass der Kontext gewinnt.
	echte := netz.NeueDrossel(time.Minute, time.Now, nil)
	if fehler := echte.Warte(t.Context()); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	abgebrochen, abbrechen := contextMitAbbruch(t)
	abbrechen()

	if fehler := echte.Warte(abgebrochen); fehler == nil {
		t.Error("das Warten lief trotz Abbruch durch")
	}
}

// **Und ohne eingesetzten Schlaf wird wirklich gewartet.**
//
// # Warum dieser Fall dabeisteht
//
// Sonst liefe der Betriebsweg — der mit einer echten Uhr — durch keinen Test.
// Die Zeit ist so kurz gewählt, dass der Lauf nichts kostet, und der Beleg ist
// **dass** gewartet wurde, nicht wie lange: eine Messung auf die Millisekunde
// flackert unter Last (§3).
func TestOhneEingesetztenSchlafWirdWirklichGewartet(t *testing.T) {
	t.Parallel()

	drossel := netz.NeueDrossel(20*time.Millisecond, time.Now, nil)

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	los := time.Now()

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("die zweite Anfrage: %v", fehler)
	}

	if gedauert := time.Since(los); gedauert < 10*time.Millisecond {
		t.Errorf("es wurde nur %v gewartet — die Drossel greift nicht", gedauert)
	}
}
