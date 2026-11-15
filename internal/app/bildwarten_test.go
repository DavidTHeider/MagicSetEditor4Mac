package app_test

// **Was geschieht, während ein Bild entsteht** — Werkstück A17 von Phase 5.3
// (#1532).
//
// # Warum es diese Läufe gibt
//
// **Kein einziger hat die Bilderzeugung je warten sehen.** In der Roadmap
// steht zu A12: „**Nicht gelaufen**: alles gegen den echten Dienst — ein
// Bildauftrag kostet Geld." Jedes Doppel antwortet **sofort**; das ist die
// Vorrichtung, die großzügiger ist als der Betrieb (§3).
//
// **Damit war ungeprüft, was in genau der Minute geschieht, in der das
// Programm nichts tun kann.** Lässt sich abbrechen? Endet der Auftrag dann
// wirklich? Bleibt ein halber Stand zurück?
//
// # Und warum kein echtes Bildmodell
//
// **Entscheidung des Auftraggebers vom 2026-09-12**: die KI-Modelle werden am
// Ende von 0.5.0 gelöscht und nicht vorgehalten. Ein lokales Stable Diffusion
// wären mehrere Gigabyte für einen Lauf, der danach nicht wiederholbar ist.
//
// **Was das kostet, gehört gesagt** (0007, dieselbe Logik): ob ein **echtes**
// Modellbild in Format, Größe und Farbraum so ankommt, wie dieses Programm es
// erwartet, bleibt unbelegt. Die Vorrichtung liefert ein echtes PNG — eines,
// das dieses Repositorium selbst geschrieben hat.

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// langsamebildquelle braucht Zeit und hört auf den Abbruch.
//
// **Beides gehört zusammen.** Eine Vorrichtung, die schläft und den Abbruch
// **nicht** hört, hielte die Teststufe fest — der Fehler steht in diesem
// Projekt schon als Merkposten.
type langsamebildquelle struct {
	*laufzeitdoppel

	dauer time.Duration
	bild  app.KIRohbild

	// begonnen und beendet zählen, was die Gegenstelle wirklich gesehen hat.
	begonnen atomic.Int32
	beendet  atomic.Int32

	// abgebrochen zählt die Läufe, die der Kontext beendet hat.
	abgebrochen atomic.Int32
}

func (l *langsamebildquelle) Bild(ctx context.Context, _ app.KIBildwunsch) (
	app.KIRohbild, error,
) {
	l.begonnen.Add(1)

	select {
	case <-time.After(l.dauer):
		l.beendet.Add(1)

		return l.bild, nil

	case <-ctx.Done():
		l.abgebrochen.Add(1)

		return app.KIRohbild{}, ctx.Err()
	}
}

// mitLangsamerQuelle baut eine Bildwelt, deren Anbieter Zeit braucht.
func mitLangsamerQuelle(t *testing.T, dauer time.Duration) (
	bildwelt, *langsamebildquelle,
) {
	t.Helper()

	welt := einebildwelt(t)

	langsam := &langsamebildquelle{
		laufzeitdoppel: welt.laufzeit, dauer: dauer,
		bild: app.KIRohbild{
			Inhalt: welt.bild, Medientyp: "image/png",
			Modell: "gpt-image-2.5-flare", Umschrift: "ein Blitz",
		},
	}

	welt.anwendung.MitKILaufzeit(
		func(string) app.KILaufzeitquelle { return langsam })

	return welt, langsam
}

// **Ein laufender Bildauftrag lässt sich abbrechen.**
//
// # Der Befund
//
// **Es gab keinen Weg dazu.** `KIBild.tsx` kennt einen Zustand `laeuft`, der
// Knöpfe abschaltet — und nichts, was den Lauf beendet. Wer einen Auftrag
// anstößt, den er nicht wollte, wartet ihn ab und bezahlt ihn.
//
// **A15 hat den Weg längst**: `IndexAbbrechen`. Der Bildweg hatte ihn nicht.
func TestEinLaufenderBildauftragLaesstSichAbbrechen(t *testing.T) {
	t.Parallel()

	welt, langsam := mitLangsamerQuelle(t, time.Minute)

	fertig := make(chan app.KIBildvorschlagAntwort, 1)

	go func() { fertig <- welt.vorschlagen(t) }()

	// **Warten, bis die Gegenstelle wirklich angefangen hat** — ein Abbruch
	// vor dem Ruf bräche nichts ab und wäre grün.
	warteAufBeginn(t, langsam)

	if !welt.anwendung.KIBildAbbrechen(welt.karte) {
		t.Fatal("der Abbruch fand keinen laufenden Auftrag")
	}

	antwort := <-fertig

	if antwort.Fehler == nil {
		t.Fatal("der abgebrochene Auftrag kam mit einem Vorschlag zurück")
	}

	// **Die Gegenstelle hat den Abbruch gesehen** — und nicht bloß der
	// Aufrufer aufgehört zuzuhören.
	if langsam.abgebrochen.Load() != 1 {
		t.Errorf("die Gegenstelle sah %d Abbrüche", langsam.abgebrochen.Load())
	}

	if langsam.beendet.Load() != 0 {
		t.Errorf("der Auftrag lief trotz Abbruch zu Ende")
	}
}

// **Und ein Abbruch ohne laufenden Auftrag ist kein Fehler.**
//
// **Sondern eine Auskunft** — dieselbe Gestalt wie `IndexAbbrechen`: „es lief
// nichts" ist eine Antwort, kein Fehlschlag.
func TestEinAbbruchOhneAuftragIstEineAuskunft(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	if welt.anwendung.KIBildAbbrechen(welt.karte) {
		t.Fatal("der Abbruch fand einen Auftrag, den es nicht gibt")
	}
}

// **Ein Abbruch trifft nur die Karte, die gemeint ist.**
//
// **Die Gegenprobe** (§10): ein Abbruch, der jeden Auftrag beendete, hielte
// den Lauf ganz oben — und nähme jemandem den Auftrag weg, den er gerade an
// einer anderen Karte angestoßen hat.
func TestEinAbbruchTrifftNurSeineKarte(t *testing.T) {
	t.Parallel()

	welt, langsam := mitLangsamerQuelle(t, 300*time.Millisecond)

	fertig := make(chan app.KIBildvorschlagAntwort, 1)

	go func() { fertig <- welt.vorschlagen(t) }()

	warteAufBeginn(t, langsam)

	// **Eine andere Karte** — dieselbe Anwendung, derselbe Anbieter.
	if welt.anwendung.KIBildAbbrechen(welt.karte + 1) {
		t.Fatal("der Abbruch fand einen Auftrag an einer fremden Karte")
	}

	if antwort := <-fertig; antwort.Fehler != nil {
		t.Fatalf("der Auftrag wurde mit abgebrochen: %+v", antwort.Fehler)
	}
}

// **Nach einem Abbruch bleibt kein halber Stand zurück.**
//
// **Kein Vorschlag, kein Bild im Vorrat** (§8). Ein halber Stand wäre ein
// Bild, das niemand angefordert hat und das trotzdem Platz belegt.
func TestNachEinemAbbruchBleibtKeinHalberStand(t *testing.T) {
	t.Parallel()

	welt, langsam := mitLangsamerQuelle(t, time.Minute)

	fertig := make(chan app.KIBildvorschlagAntwort, 1)

	go func() { fertig <- welt.vorschlagen(t) }()

	warteAufBeginn(t, langsam)

	welt.anwendung.KIBildAbbrechen(welt.karte)

	antwort := <-fertig

	if antwort.Kennung != "" {
		t.Fatalf("der Abbruch gab eine Vorschlagskennung: %q", antwort.Kennung)
	}

	herkunft := welt.anwendung.KIBildHerkunft(probekontext(t), welt.karte, "kunst")
	if len(herkunft.Angaben) != 0 {
		t.Errorf("nach dem Abbruch steht Herkunft da: %+v", herkunft.Angaben)
	}
}

// warteAufBeginn hält an, bis die Gegenstelle den Auftrag gesehen hat.
//
// **Auf ein Signal und nicht auf eine Frist** (§3): ein `Sleep` wäre unter
// Last zu kurz und sonst zu lang.
func warteAufBeginn(t *testing.T, langsam *langsamebildquelle) {
	t.Helper()

	frist := time.After(10 * time.Second)

	for langsam.begonnen.Load() == 0 {
		select {
		case <-frist:
			t.Fatal("die Gegenstelle hat den Auftrag nie gesehen")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
