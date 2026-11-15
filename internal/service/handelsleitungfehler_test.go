package service_test

// **Was die Handelsleitung abweist** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Warum die Fehlerwege eine eigene Datei haben
//
// **Weil sie eine andere Frage beantworten** (§2): nebenan steht, **was** die
// Leitung tut; hier steht, **was sie nicht tut** — und vor allem, dass sie
// dabei nichts schreibt. Die Zeilengrenze macht den Schnitt nur sichtbar.
//
// # Und weil hier der Unterschied liegt, an dem es hängt
//
// **„Gibt es nicht" ist beim Eröffnen die gute Antwort.** Wer jeden Lesefehler
// so läse, legte über einem kaputten Bestand einen zweiten Vorgang an.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Vorgang wird mit der ersten Revision eröffnet und mit keiner anderen.**
//
// **Sonst stünde eine Revision 7 ohne die sechs davor** — und ihr
// Elternabdruck zeigte auf etwas, das dieser Rechner nie gesehen hat.
func TestEinVorgangWirdMitDerErstenRevisionEroeffnet(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	spaet := einHandelsangebot()
	spaet.Revision = 2
	spaet.Eltern = "irgendwas"

	if _, fehler := leitung.Eroeffnen(
		context.Background(), spaet); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Was der Bestand klagt, kommt zurück — und wird nicht für „gibt es nicht"
// gehalten.**
//
// **Das ist der Unterschied, an dem es hängt**: „gibt es nicht" ist beim
// Eröffnen die gute Antwort. Wer jeden Lesefehler so läse, legte über einem
// kaputten Bestand einen zweiten Vorgang an.
func TestWasDerBestandKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Platte ist voll")
	ablage := &handelsprobe{
		staende: map[string]domain.Handelsangebot{}, klagt: kaputt,
	}
	leitung := service.NeueHandelsleitung(ablage, &zusagenhaltprobe{})
	ctx := context.Background()

	if _, fehler := leitung.Eroeffnen(
		ctx, einHandelsangebot()); !errors.Is(fehler, kaputt) {
		t.Errorf("beim Eröffnen kam %v", fehler)
	}

	if _, fehler := leitung.Vorschlagen(
		ctx, einHandelsangebot(), "bert"); !errors.Is(fehler, kaputt) {
		t.Errorf("beim Vorschlagen kam %v", fehler)
	}

	if _, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittAnbieten, "anna",
	); !errors.Is(fehler, kaputt) {
		t.Errorf("beim Schreiten kam %v", fehler)
	}

	if ablage.geschrieben != 0 {
		t.Errorf("geschrieben wurde %dmal", ablage.geschrieben)
	}
}

// **Und was das Schreiben klagt, kommt ebenfalls zurück.**
//
// **Ohne halben Stand**: die Leitung gibt in diesem Fall **kein** Angebot
// zurück, sondern den Fehler. Wer den neuen Stand zurückgäbe, hätte ihn im
// Fenster und nicht im Bestand (§7, Regel 3).
func TestWasDasSchreibenKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	voll := errors.New("die Platte ist voll")
	ablage := &handelsprobe{staende: map[string]domain.Handelsangebot{}}
	leitung := service.NeueHandelsleitung(ablage, &zusagenhaltprobe{})
	ctx := context.Background()

	ablage.schreibklage = voll

	if _, fehler := leitung.Eroeffnen(
		ctx, einHandelsangebot()); !errors.Is(fehler, voll) {
		t.Fatalf("beim Eröffnen kam %v", fehler)
	}

	// **Jetzt liegt ein angebotener Vorgang da** — geschrieben von Hand,
	// damit die beiden anderen Wege ihren Schreibfehler überhaupt erreichen.
	// Angeboten und nicht im Entwurf: aus einem Entwurf führt kein
	// Vorschlag (QA-06b, #1967).
	ablage.schreibklage = nil

	erste := angebotenerHandel(t, leitung)

	ablage.schreibklage = voll

	if _, fehler := leitung.Vorschlagen(
		ctx, erste.Fortgeschrieben(mitMenge(erste, 3)), "bert",
	); !errors.Is(fehler, voll) {
		t.Errorf("beim Vorschlagen kam %v", fehler)
	}

	if _, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittZustimmen, "anna",
	); !errors.Is(fehler, voll) {
		t.Errorf("beim Schreiten kam %v", fehler)
	}
}

// **Einen Vorgang, den es nicht gibt, findet auch ein Vorschlag nicht.**
func TestEinUnbekannterVorgangWirdNichtFortgeschrieben(t *testing.T) {
	t.Parallel()

	leitung, _ := eineHandelsleitung()

	if _, fehler := leitung.Vorschlagen(
		context.Background(), einHandelsangebot(), "bert",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine wiederholte Zustimmung schreibt nicht.**
//
// **Sie ändert nichts** — und ein Schreiben, das nichts ändert, kostet eine
// Fassung, an der ein anderer gerade rechnet (§7, Regel 4).
func TestEineWiederholteZustimmungSchreibtNicht(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	ctx := context.Background()

	if _, fehler := leitung.Eroeffnen(ctx, einHandelsangebot()); fehler != nil {
		t.Fatalf("Eroeffnen: %v", fehler)
	}

	if _, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittAnbieten, "anna"); fehler != nil {
		t.Fatalf("Anbieten: %v", fehler)
	}

	if _, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittZustimmen, "bert"); fehler != nil {
		t.Fatalf("Zustimmen: %v", fehler)
	}

	vorher := ablage.geschrieben

	nochmal, fehler := leitung.Schreiten(
		ctx, "h1", domain.SchrittZustimmen, "bert")
	if fehler != nil {
		t.Fatalf("noch einmal zustimmen: %v", fehler)
	}

	if len(nochmal.Zustimmungen) != 1 {
		t.Errorf("es stehen %d Zustimmungen da", len(nochmal.Zustimmungen))
	}

	if ablage.geschrieben != vorher {
		t.Errorf("die Wiederholung hat geschrieben: %d statt %d",
			ablage.geschrieben, vorher)
	}
}

// **Und der Stand ist die letzte Revision.**
func TestDerStandIstDieLetzteRevision(t *testing.T) {
	t.Parallel()

	leitung, _ := eineHandelsleitung()
	ctx := context.Background()

	erste := angebotenerHandel(t, leitung)

	if _, klage := leitung.Vorschlagen(
		ctx, erste.Fortgeschrieben(mitMenge(erste, 3)), "bert"); klage != nil {
		t.Fatalf("Vorschlagen: %v", klage)
	}

	stand, fehler := leitung.Stand(ctx, "h1")
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if stand.Revision != 2 || stand.Positionen[0].Menge != 3 {
		t.Errorf("der Stand ist %+v", stand)
	}
}
