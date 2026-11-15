package service_test

// **Die Schleife über einer offenen Sitzung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass jedes der vier Urteile etwas anderes bewirkt.** Das Issue verlangt
// „Doppelte/geänderte IDs, Lücken, falsche Hashes und alte Epochs blockieren
// statt weiterzuspielen" — und das ist **nicht** viermal dasselbe: eine
// Wiederholung ist der Normalfall und fällt still weg, eine Lücke hält an.
// Wer beide gleich behandelte, hätte entweder eine Partie, die bei jeder
// wiederholten Zustellung stehenbleibt, oder eine, die über eine Lücke
// hinwegspielt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was neu ist, kommt beim Boten an.**
func TestEineNeueNachrichtKommtAn(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, einEreignis(1)), verpackt(t, einEreignis(2)),
	}}
	bote := &botendoppel{}

	fehler := netzUeber(zugang, bote).Bediene(context.Background())

	switch {
	case !errors.Is(fehler, errStromZuEnde):
		t.Fatalf("es kam %v", fehler)
	case len(bote.bekommen) != 2:
		t.Fatalf("%d Nachrichten, erwartet 2", len(bote.bekommen))
	case bote.bekommen[1].Folge != 2:
		t.Errorf("die zweite trägt Folge %d", bote.bekommen[1].Folge)
	}
}

// **Eine zweite Zustellung fällt still weg.**
//
// **Still und nicht als Befund**: auf einer Leitung, die wiederholt, ist sie
// der Normalfall. Ein Protokoll voller Meldungen über den Normalbetrieb
// verdeckt die Lücke, die wirklich eine ist (§8).
func TestEineZweiteZustellungWirktNichtNochEinmal(t *testing.T) {
	t.Parallel()

	erste := verpackt(t, einEreignis(1))

	zugang := &zugangsdoppel{zustellungen: [][]byte{erste, erste}}
	bote := &botendoppel{}

	gemeldet := 0
	netz := netzUeber(zugang, bote).MitMeldung(func(error) { gemeldet++ })

	_ = netz.Bediene(context.Background()) //nolint:errcheck // der Lauf endet am Strom

	switch {
	case len(bote.bekommen) != 1:
		t.Errorf("%d Nachrichten, erwartet 1", len(bote.bekommen))
	case gemeldet != 0:
		t.Errorf("%d Meldungen — eine Wiederholung ist kein Befund", gemeldet)
	}
}

// **Eine Lücke hält an.**
//
// Das Issue verlangt es wörtlich: „Lücken … blockieren statt weiterzuspielen".
// Wer über sie hinwegspielte, schriebe einen Stand fort, dem ein Schritt
// fehlt — und niemand könnte hinterher sagen, welcher.
func TestEineLueckeHaeltAn(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, einEreignis(1)),
		verpackt(t, einEreignis(3)),
		verpackt(t, einEreignis(4)),
	}}
	bote := &botendoppel{}

	fehler := netzUeber(zugang, bote).Bediene(context.Background())

	switch {
	case fehler == nil:
		t.Fatal("die Lücke kam durch")
	case errors.Is(fehler, errStromZuEnde):
		t.Fatal("der Lauf ist bis zum Ende gelaufen")
	case len(bote.bekommen) != 1:
		t.Errorf("%d Nachrichten beim Boten, erwartet 1", len(bote.bekommen))
	case zugang.gelesen != 2:
		t.Errorf("gelesen wurden %d — nach der Lücke wird nicht weitergelesen",
			zugang.gelesen)
	}
}

// **Eine alte Epoche fällt weg und wird gemeldet.**
//
// **Gemeldet und nicht still**: anders als eine Wiederholung ist sie kein
// Normalfall — sie heißt, dass am anderen Ende ein Stand liegt, den es hier
// nicht mehr gibt.
func TestEineAlteEpocheWirdGemeldetUndFaelltWeg(t *testing.T) {
	t.Parallel()

	// **Alt heißt: älter als der eigene Stand.** Der Empfänger steht in der
	// zweiten Epoche, die Nachricht kommt aus der ersten.
	alt := einEreignis(1)

	zugang := &zugangsdoppel{zustellungen: [][]byte{verpackt(t, alt)}}
	bote := &botendoppel{}

	var gemeldet []error

	netz := netzInEpoche(zugang, bote, 2).MitMeldung(func(f error) {
		gemeldet = append(gemeldet, f)
	})

	fehler := netz.Bediene(context.Background())

	switch {
	case !errors.Is(fehler, errStromZuEnde):
		t.Fatalf("der Lauf endete mit %v", fehler)
	case len(bote.bekommen) != 0:
		t.Errorf("die alte Epoche kam beim Boten an")
	case len(gemeldet) != 1:
		t.Errorf("%d Meldungen, erwartet 1", len(gemeldet))
	}
}

// **Was zu einem Stand gehört, den es nicht mehr gibt, fällt still weg.**
func TestEineVeralteteSitzungReichtNichtsWeiter(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{
		zustellungen: [][]byte{verpackt(t, einEreignis(1))},
		veraltet:     true,
	}
	bote := &botendoppel{}

	gemeldet := 0
	netz := netzUeber(zugang, bote).MitMeldung(func(error) { gemeldet++ })

	_ = netz.Bediene(context.Background()) //nolint:errcheck // der Lauf endet am Strom

	if len(bote.bekommen) != 0 || gemeldet != 0 {
		t.Errorf("%d Nachrichten, %d Meldungen — beides erwartet 0",
			len(bote.bekommen), gemeldet)
	}
}

// **Eine unlesbare Zustellung hält die Sitzung nicht an** (§8).
//
// Sie ist ein Befund über **eine** Nachricht. Wer dafür die Leitung
// niederlegte, ließe ein fremdes Programm die Partie beenden — mit einem
// einzigen kaputten Byte.
func TestEineUnlesbareZustellungWirdGemeldetUndUebersprungen(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		[]byte("{kein json"), verpackt(t, einEreignis(1)),
	}}
	bote := &botendoppel{}

	var gemeldet []error

	netz := netzUeber(zugang, bote).MitMeldung(func(f error) {
		gemeldet = append(gemeldet, f)
	})

	fehler := netz.Bediene(context.Background())

	switch {
	case !errors.Is(fehler, errStromZuEnde):
		t.Fatalf("der Lauf endete mit %v", fehler)
	case len(gemeldet) != 1:
		t.Fatalf("%d Meldungen, erwartet 1", len(gemeldet))
	case len(bote.bekommen) != 1:
		t.Errorf("%d Nachrichten — die danach muss ankommen",
			len(bote.bekommen))
	}
}

// **Ohne Melder geht ein Befund verloren, und nichts bricht.**
//
// **Die Gegenprobe zum Lauf darüber** (§10): dort ist der Melder gesetzt, hier
// nicht — und die Zusage ist, dass dieser Unterschied nur die Meldung betrifft
// und nicht den Lauf.
func TestOhneMelderLaeuftDieSchleifeWeiter(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		[]byte("{kein json"), verpackt(t, einEreignis(1)),
	}}
	bote := &botendoppel{}

	fehler := netzUeber(zugang, bote).Bediene(context.Background())

	if !errors.Is(fehler, errStromZuEnde) || len(bote.bekommen) != 1 {
		t.Errorf("es kam %v bei %d Nachrichten", fehler, len(bote.bekommen))
	}
}

// **Was der Bote klagt, hält an.**
//
// **Weil er die Fachseite ist**: kann die Partieleitung ein Ereignis nicht
// anwenden, ist der Stand hier nicht mehr der drüben — und weiterzulesen
// hieße, die Abweichung zu vergrößern.
func TestWasDerBoteKlagtHaeltAn(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, einEreignis(1)), verpackt(t, einEreignis(2)),
	}}

	fehler := netzUeber(zugang, &botendoppel{klagt: kaputt}).Bediene(
		context.Background())

	switch {
	case !errors.Is(fehler, kaputt):
		t.Fatalf("es kam %v", fehler)
	case zugang.gelesen != 1:
		t.Errorf("gelesen wurden %d — nach der Klage wird nicht weitergelesen",
			zugang.gelesen)
	}
}

// **Was die Leitung klagt, beendet den Ruf.**
//
// **Und zwar ohne Deutung**: ob ein Abriss den Wirt kostet, entscheidet
// der Gast (`app.App.tischVerloren`) und nicht der Transport.
func TestEinLesefehlerBeendetDenRuf(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Leitung ist weg")

	fehler := netzUeber(
		&zugangsdoppel{lesefehler: kaputt}, &botendoppel{},
	).Bediene(context.Background())

	if !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein beendeter Kontext beendet den Ruf, bevor gelesen wird.**
func TestEinBeendeterKontextBeendetDenRuf(t *testing.T) {
	t.Parallel()

	ctx, beenden := context.WithCancel(context.Background())
	beenden()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, einEreignis(1)),
	}}

	fehler := netzUeber(zugang, &botendoppel{}).Bediene(ctx)

	switch {
	case !errors.Is(fehler, context.Canceled):
		t.Fatalf("es kam %v", fehler)
	case zugang.gelesen != 0:
		t.Errorf("es wurde %d-mal gelesen", zugang.gelesen)
	}
}

// **Präsenz geht an der Reihenfolge vorbei — auch in der Schleife.**
//
// **Ein eigener Lauf, obwohl `Partieempfang` es schon hält**: dort ist es die
// Rechnung, hier der **Weg**. Eine Schleife, die Flüchtiges durch den Stand
// schickte, bekäme beim zweiten Mauszeiger ein „schon da" und ließe ihn fallen
// — und die Zusage „verlorene Präsenz darf verschwinden" wäre in ihr Gegenteil
// verkehrt.
func TestPraesenzKommtAuchMehrfachAn(t *testing.T) {
	t.Parallel()

	// **Eine Epoche trägt auch Flüchtiges** — sie sagt, zu welchem Lauf es
	// gehört. Was ihm fehlt, ist die **Folge**, und genau darum geht es hier.
	zeiger := domain.Partienachricht{
		Art: domain.ArtPraesenz, Partie: "p1", Epoche: 1, Von: "sitz-2",
		Inhalt: []byte(`{"x":1,"y":2}`),
	}

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, zeiger), verpackt(t, zeiger), verpackt(t, zeiger),
	}}
	bote := &botendoppel{}

	_ = netzUeber(zugang, bote).Bediene(context.Background()) //nolint:errcheck // der Lauf endet am Strom

	if len(bote.bekommen) != 3 {
		t.Errorf("%d Zeiger angekommen, erwartet 3", len(bote.bekommen))
	}
}

// **Was beim Leeren klagt, beendet den Ruf.**
//
// # Warum das anhält und nicht bloß gemeldet wird
//
// **Weil danach etwas fehlt, das der Gast erwartet.** In der Schlange liegen
// die Ereignisse und das Tischbild zu dem Zug, der gerade angekommen ist;
// geht das nicht hinaus, sieht die Gegenseite einen Tisch, der nicht
// nachgezogen ist — und die nächste Nachricht baute darauf auf. Ein Befund im
// Protokoll änderte daran nichts.
func TestWasBeimLeerenKlagtBeendetDenRuf(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Leitung ist weg")

	zugang := &zugangsdoppel{
		zustellungen: [][]byte{
			verpackt(t, einEreignis(1)), verpackt(t, einEreignis(2)),
		},
		leerfehler: kaputt,
	}
	bote := &botendoppel{}

	fehler := netzUeber(zugang, bote).Bediene(context.Background())

	switch {
	case !errors.Is(fehler, kaputt):
		t.Fatalf("es kam %v", fehler)
	case zugang.gelesen != 1:
		t.Errorf("gelesen wurden %d — nach der Klage wird nicht weitergelesen",
			zugang.gelesen)
	case len(bote.bekommen) != 1:
		t.Errorf("%d beim Boten — die erste Nachricht gilt trotzdem",
			len(bote.bekommen))
	}
}

// **Und im Normalbetrieb wird nach jeder Nachricht geleert.**
//
// **Nach jeder und nicht am Ende**: was zu einem Zug gehört, geht hinaus,
// bevor der nächste gelesen wird — sonst sammelte sich eine Partie in einer
// Schlange, und die Gegenseite bekäme alles auf einmal oder gar nicht.
func TestNachJederNachrichtWirdGeleert(t *testing.T) {
	t.Parallel()

	zugang := &zugangsdoppel{zustellungen: [][]byte{
		verpackt(t, einEreignis(1)), verpackt(t, einEreignis(2)),
	}}

	_ = netzUeber(zugang, &botendoppel{}).Bediene(context.Background()) //nolint:errcheck // der Lauf endet am Strom

	if zugang.geleert != 2 {
		t.Errorf("%dmal geleert, erwartet 2", zugang.geleert)
	}
}
