package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein abgebrochener Satz wird nicht halb angelegt** —
// #1042,
// Befund 7 des Nachlaufs über Phase V.
//
// # Der Zustand, den es zu verhindern gilt
//
// `Setkarten` blättert bis zu einer Grenze und gibt danach die **Teilmenge**
// samt Befund zurück — ohne Fehler. `Einlesen` legte sie als gewöhnliches Set
// ab, und ein zweiter Aufruf wurde wegen derselben Herkunft als „schon offen"
// abgewiesen: **derselbe Weg konnte das Set nie vervollständigen.**
//
// Das Änderungsprotokoll sagte „der ganze Satz … mit allen Karten".
func TestEinAbgebrochenerSatzWirdNichtHalbAngelegt(t *testing.T) {
	t.Parallel()

	speicher := &speicherprobe{}
	fremd := &satzprobe{satz: scryfall.Satz{
		Karten:      []*scryfall.Karte{{Name: "Eine", Kennung: "a"}},
		Gesamt:      5000,
		Abgebrochen: true,
	}}

	einlesen := service.NeuerSatzimport(fremd, speicher, nil)

	_, fehler := einlesen.Einlesen(t.Context(), "plst", "magic")
	if fehler == nil {
		t.Fatal("der abgebrochene Satz kam ohne Fehler durch")
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v — kein Validierungsbefund", fehler)
	}

	// **Die Zahlen stehen in der Meldung.** „Es hat nicht geklappt" ohne das
	// Wieviel ließe niemanden entscheiden, was zu tun ist.
	for _, teil := range []string{"5000", "plst"} {
		if !strings.Contains(fehler.Error(), teil) {
			t.Errorf("Fehler %v — %q fehlt", fehler, teil)
		}
	}

	// **Und es liegt kein Set da.** Das ist der Punkt: ein halbes Set, das als
	// vollständig gilt, ist schlimmer als keines.
	if speicher.set != nil {
		t.Errorf("es wurde ein Set angelegt: %v", speicher.set.Name)
	}
}

// **Ein vollständiger Satz kommt weiterhin durch.**
//
// Ohne diesen Fall bestünde den Lauf darüber auch eine Fassung, die **jeden**
// Satz abwiese.
func TestEinVollstaendigerSatzKommtDurch(t *testing.T) {
	t.Parallel()

	speicher := &speicherprobe{}
	fremd := &satzprobe{satz: scryfall.Satz{
		Karten: []*scryfall.Karte{{Name: "Eine", Kennung: "a"}},
		Gesamt: 1,
	}}

	einlesen := service.NeuerSatzimport(fremd, speicher, nil)

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if speicher.set == nil {
		t.Fatal("es wurde kein Set angelegt")
	}
}

// **Wer das Rennen verliert, bekommt das Set des Gewinners** —
// #1042,
// Befund 8.
//
// # Warum der Konflikt allein nicht genügt
//
// Die Bedingung im Schema (Migration 0015) sorgt dafür, dass **ein** Set
// entsteht. Sie sagt dem Verlierer aber nur „Konflikt" — und der wollte
// dasselbe, was der Gewinner jetzt hat.
//
// **Es ist dieselbe Antwort wie beim zweiten Aufruf hintereinander**:
// `SchonOffen` mit der Kennung des vorhandenen Sets. Ein Rennen soll sich
// nicht anders anfühlen als eine Wiederholung.
func TestWerDasRennenVerliertBekommtDasVorhandeneSet(t *testing.T) {
	t.Parallel()

	// **Der Bestand ist beim Nachsehen leer und beim zweiten Blick nicht** —
	// genau der Verlauf eines verlorenen Rennens.
	koepfe := &wachsendeKoepfe{spaeter: []domain.SetHead{
		{ID: 7, Cards: 2, Origin: service.Satzherkunft("lea")},
	}}

	einlesen := service.NeuerSatzimport(
		&satzprobe{satz: einSatz()},
		&speicherprobe{fehler: domain.ErrConflict},
		koepfe)

	ergebnis, fehler := einlesen.Einlesen(t.Context(), "lea", "magic")
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if !ergebnis.SchonOffen || ergebnis.SetID != 7 {
		t.Errorf("Ergebnis %+v — erwartet war das vorhandene Set 7", ergebnis)
	}
}

// **Ein Konflikt ohne vorhandenes Set bleibt ein Fehler.**
//
// Ohne diesen Fall verschluckte der Weg darüber jeden Konflikt — auch einen,
// der nichts mit dem Rennen zu tun hat.
func TestEinKonfliktOhneVorhandenesSetBleibtEinFehler(t *testing.T) {
	t.Parallel()

	einlesen := service.NeuerSatzimport(
		&satzprobe{satz: einSatz()},
		&speicherprobe{fehler: domain.ErrConflict},
		koepfeprobe{})

	if _, fehler := einlesen.Einlesen(t.Context(), "lea", "magic"); fehler == nil {
		t.Fatal("der Konflikt wurde verschluckt")
	}
}

// **Und ein anderer Fehler wird nicht zum vorhandenen Set gedeutet.**
//
// # Warum dieser Fall dazugehört
//
// Die Prüfung besteht aus zwei Bedingungen: „ein Konflikt" **und** „es steht
// wirklich eines da". Die Gegenprobe auf die erste blieb ohne diesen Lauf
// grün — im Test darüber war die zweite ohnehin nicht erfüllt, und damit
// unterschied nichts, ob die erste geprüft wird.
//
// Hier ist die zweite erfüllt: der Bestand führt den Satz. Ein Schreibfehler,
// der **kein** Konflikt ist, darf trotzdem nicht als „schon offen" durchgehen
// — er sagt etwas anderes, und ein Erfolg an seiner Stelle verschwiege es.
func TestEinAndererFehlerWirdNichtZumVorhandenenSetGedeutet(t *testing.T) {
	t.Parallel()

	// **Über einen Bestand, der den Satz beim ersten Blick noch nicht führt.**
	// Führte er ihn schon, käme der Lauf gar nicht bis zum Schreiben — und
	// prüfte einen anderen Zweig als seinen.
	koepfe := &wachsendeKoepfe{spaeter: []domain.SetHead{
		{ID: 7, Cards: 2, Origin: service.Satzherkunft("lea")},
	}}

	einlesen := service.NeuerSatzimport(
		&satzprobe{satz: einSatz()},
		&speicherprobe{fehler: errors.New("die Platte ist voll")},
		koepfe)

	_, fehler := einlesen.Einlesen(t.Context(), "lea", "magic")
	if fehler == nil {
		t.Fatal("die volle Platte wurde als vorhandenes Set ausgegeben")
	}

	if !strings.Contains(fehler.Error(), "die Platte ist voll") {
		t.Errorf("Fehler %v — der des Schreibers war es nicht", fehler)
	}
}

// wachsendeKoepfe ist beim ersten Blick leer und danach nicht mehr.
type wachsendeKoepfe struct {
	gefragt bool
	spaeter []domain.SetHead
}

func (k *wachsendeKoepfe) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	if !k.gefragt {
		k.gefragt = true

		return nil, nil
	}

	return k.spaeter, nil
}
