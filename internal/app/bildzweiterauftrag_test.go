package app_test

// **Ein zweiter Bildauftrag an derselben Karte wird abgewiesen** — Befund 3
// des Nachlaufs über Phase 5.3.
//
// # Warum es eine eigene Datei ist
//
// `bildwarten_test.go` steht bei 225 Zeilen und trägt die Vorrichtung; der
// Schnitt folgt der Frage: dort, wie ein laufender Auftrag **angehalten**
// wird, hier, dass es nur **einen** gibt.
//
// # Der Befund
//
// **Die Identität am Eintrag schützte nur den jüngeren Lauf.** Ein zweiter
// Ruf an derselben Karte übernahm den Abbruchweg, und der erste lief ohne
// einen zu Ende — bezahlt und nicht anzuhalten. §7, Regel 4 verlangt einen
// Vorgang je Datensatz; `useVorgang` hält das am Fenster, die Bindung ist
// aber auch direkt zu rufen.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinZweiterBildauftragAnDerselbenKarteWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt, langsam := mitLangsamerQuelle(t, 300*time.Millisecond)

	fertig := make(chan app.KIBildvorschlagAntwort, 1)

	go func() { fertig <- welt.vorschlagen(t) }()

	warteAufBeginn(t, langsam)

	// **Derselbe Ruf ein zweites Mal**, während der erste läuft.
	zweiter := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "ein Blitz",
		domain.KIBildgroesseHoch, domain.KIBildgueteMittel, false, "")

	if zweiter.Fehler == nil {
		t.Fatal("der zweite Auftrag kam durch")
	}

	if zweiter.Fehler.Code != "konflikt" {
		t.Errorf("der Fehler ist %+v", zweiter.Fehler)
	}

	// **Und der erste läuft zu Ende** — abgewiesen wird der zweite, nicht
	// der, der schon arbeitet.
	if antwort := <-fertig; antwort.Fehler != nil {
		t.Fatalf("der erste Auftrag scheiterte: %+v", antwort.Fehler)
	}

	// **Die Gegenstelle wurde einmal gefragt und nicht zweimal.** Ohne diese
	// Zeile wäre ein zweiter Ruf, der Geld kostet und dann abgewiesen wird,
	// von einem abgewiesenen nicht zu unterscheiden.
	if langsam.begonnen.Load() != 1 {
		t.Errorf("die Gegenstelle sah %d Aufträge", langsam.begonnen.Load())
	}
}
