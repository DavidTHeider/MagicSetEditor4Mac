package service_test

// **Und zwei Menschen spielen über die Leitung** — die zweite Hälfte der
// Phasenabnahme
// (#1593).
//
// # Was hier dazukommt
//
// **Nebenan spielt eine Partie, aber beide Sitze rufen dieselbe Naht.** Das
// beantwortet die Frage nach den Regeln und nicht die nach der Entfernung.
// Hier sitzt **Bert am anderen Ende**: sein Befehl geht als **Wunsch** durch
// [service.Partiegast] hinaus, wird verpackt, kommt beim
// [service.Partiewirt] an, wird von der **echten** [service.Partieleitung]
// ausgelegt — und was daraus wurde, kommt als **Tatsache** zurück.
//
// **Die Läufe von `G5` prüfen dieselbe Naht gegen ein Doppel.** Das ist
// richtig für die Frage, die sie stellen (wer darf was, und was steht im
// Umschlag); es lässt aber die Stelle offen, an der aus einem Wunsch ein
// **Spielzustand** wird. Die steht hier.
//
// # Was er nicht hält
//
//   - **Den Draht.** Zwischen den beiden Enden liegt kein Socket, sondern
//     [draht]: er packt die Bytes aus, die eine Sitzung geschrieben hat, und
//     gibt sie dem anderen Ende. Dass dieselben Bytes über TCP mit TLS 1.3
//     gehen, hält `internal/leitung` (N2,
//     #1455).
//   - **Epoche, Folge und Wiederholung.** Darüber urteilt `Partienetz`,
//     bevor eine Nachricht hier ankommt — `Angekommen` sagt das selbst.
//
// # Und was er gefunden hat
//
// **Ein Gast erfährt heute nur von seinen eigenen Zügen**
// (#1627).
// Das steht unten als eigener Lauf — gemessen, nicht vermutet, und mit der
// Zusage, die es einlösen müsste. **Ein Lauf hat es gezeigt und kein Lesen**;
// dieselbe Gestalt wie bei
// #1623.

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// draht ist das eine Ende einer Leitung: was hineingeht, kommt drüben an.
//
// **Er packt aus und urteilt nicht.** Genau das tut `Partienetz` im Betrieb
// auch — die Reihenfolge entscheidet es, den Inhalt nicht.
type draht struct {
	drueben func(context.Context, domain.Partienachricht) error
}

func (d *draht) Sende(
	ctx context.Context, roh []byte, _ service.Zustellung,
) error {
	var nachricht domain.Partienachricht

	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		return fehler
	}

	return d.drueben(ctx, nachricht)
}

// Leere und Stand machen ihn zu einem [service.Partieplatz] — damit er in
// einer [service.Partierunde] sitzen kann und der Funk ihn erreicht (#1627).
//
// **Er hält nichts zurück**: was `Sende` bekommt, ist drüben schon da.
func (d *draht) Leere(context.Context) error { return nil }

// Schliesse legt ihn nieder — hier gibt es nichts zu schließen.
func (d *draht) Schliesse() error { return nil }

func (d *draht) Stand() service.Sitzungsstand {
	return service.Sitzungsstand{Kanal: "draht"}
}

// mitschrift hält fest, was beim Gast ankommt.
type mitschrift struct {
	ereignisse []domain.Ereignis
	bilder     []domain.Tischbild
}

func (m *mitschrift) Ereignis(_ context.Context, ereignis domain.Ereignis) error {
	m.ereignisse = append(m.ereignisse, ereignis)

	return nil
}

func (m *mitschrift) Tischbild(_ context.Context, bild domain.Tischbild) error {
	m.bilder = append(m.bilder, bild)

	return nil
}

// leitungslauf ist der Abnahmelauf, bei dem **Berts** Befehle über die
// Leitung gehen und Annas nicht.
//
// **Die Asymmetrie ist der Punkt**: Anna sitzt am Wirt und ruft die Naht
// unmittelbar, Bert sitzt am anderen Ende und **wünscht**. Ein Lauf, in dem
// beide dasselbe täten, prüfte die Leitung gar nicht.
type leitungslauf struct {
	*abnahmelauf

	gast    *service.Partiegast
	gehoert *mitschrift
	wunsch  int
}

// einLeitungslauf steckt Wirt und Gast über einen [draht] zusammen.
func einLeitungslauf(t *testing.T) *leitungslauf {
	t.Helper()

	lauf := einAbnahmelauf(t)

	zumGast := &draht{}

	// **Bert sitzt in einer Runde**, so wie in der Anwendung: `App` trägt
	// jeden Gast als [service.Gastplatz] ein, und darüber erreicht ihn der
	// Funk (#1627).
	runde := &service.Partierunde{}
	bertsSitzung := service.NeuePartiesitzung(
		"abnahme", "sitz-1", zumGast, service.NeueReihenfolge(1))
	runde.Setze(service.NeuerGastplatz("sitz-2", bertsSitzung, zumGast))

	wirt := service.NeuerPartiewirt(
		"abnahme", "sitz-2", lauf.leitung, bertsSitzung).
		MitRunde(runde)

	gehoert := &mitschrift{}
	gast := service.NeuerPartiegast("abnahme",
		service.NeuePartiesitzung("abnahme", "sitz-2",
			&draht{drueben: wirt.Angekommen}, service.NeueReihenfolge(1)),
		gehoert)

	zumGast.drueben = gast.Angekommen

	// **Und Annas Züge gehen denselben Weg wie in der Anwendung**:
	// `App.PartieBefehl` trägt das Ergebnis an den Tisch.
	funk := service.NeuerTischfunk("abnahme", lauf.leitung, runde)
	lauf.nachDemBefehl = func(
		befehl domain.Spielbefehl, ausgang service.Befehlsausgang,
	) {
		if fehler := funk.Verkuende(
			context.Background(), befehl.Absender, ausgang); fehler != nil {
			t.Fatalf("der Tisch erfuhr von %q nichts: %v", befehl.Art, fehler)
		}
	}

	return &leitungslauf{abnahmelauf: lauf, gast: gast, gehoert: gehoert}
}

// wuensche schickt einen Befehl Berts über die Leitung.
//
// **Auch hier wird vorher gefragt, ob er angeboten ist**: ein Gast sieht
// dieselbe Liste wie das Fenster des Wirts — sonst hätte er Knöpfe, die es
// nicht gibt (§16).
func (l *leitungslauf) wuensche(art string, nutzlast map[string]any) {
	l.t.Helper()

	ctx := context.Background()

	erlaubt, fehler := erlaubtVon(l.leitung, ctx, "abnahme", "sitz-2")
	if fehler != nil {
		l.t.Fatalf("Erlaubnis: %v", fehler)
	}

	if !slices.Contains(erlaubt, art) {
		l.t.Fatalf("%q ist für Bert nicht angeboten — angeboten ist %v",
			art, erlaubt)
	}

	l.wunsch++

	if !slices.Contains(l.benutzt, art) {
		l.benutzt = append(l.benutzt, art)
	}

	if fehler := l.gast.Wuensche(ctx, domain.Spielbefehl{
		Kennung:  fmt.Sprintf("w%d", l.wunsch),
		PartieID: "abnahme",
		Absender: "sitz-2",
		Art:      art,
		Nutzlast: nutzlast,
	}); fehler != nil {
		l.t.Fatalf("%q über die Leitung: %v", art, fehler)
	}
}

// sieht gibt den Namen, unter dem Bert ein Objekt sieht — **und bricht ab,
// wenn es in seinem letzten Tischbild nicht so heißt** (#2480): ein Gast
// nennt Objekte bei ihrer Sichtkennung, und die muss er kennen können.
func (l *leitungslauf) sieht(kennung string) string {
	l.t.Helper()

	partie := l.partie()
	name := ""

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			name = partie.Sichtkennung(objekt)
		}
	}

	if len(l.gehoert.bilder) == 0 {
		l.t.Fatal("Bert hat noch kein Tischbild")
	}

	for _, gesehen := range l.gehoert.bilder[len(l.gehoert.bilder)-1].Sicht.Objekte {
		if gesehen.Kennung == name {
			return name
		}
	}

	l.t.Fatalf("Bert sieht %q nicht unter %q", kennung, name)

	return ""
}

// passeBis lässt passen — Anna unmittelbar, Bert über die Leitung.
func (l *leitungslauf) passeBis(wohin string, fertig func(domain.Partie) bool) {
	l.t.Helper()

	for range 200 {
		partie := l.partie()

		if fertig(partie) {
			return
		}

		wer := partie.Zug.Prioritaet
		if wer == "" {
			l.t.Fatalf("auf dem Weg zu %s hat niemand Priorität: "+
				"Zug %d, %s/%s", wohin, partie.Zug.Nummer,
				partie.Zug.Abschnitt.Phase, partie.Zug.Abschnitt.Schritt)
		}

		if wer == "sitz-2" {
			l.wuensche(service.BefehlPassen, nil)

			continue
		}

		l.stelle(wer, service.BefehlPassen, nil)
	}

	l.t.Fatalf("nach 200 Schritten steht %s immer noch aus", wohin)
}

// **Bert spielt seinen ersten Zug über die Leitung.**
//
// Anna sitzt am Wirt; Bert wünscht — auch sein Passen. Am Ende steht seine
// Kreatur auf dem Tisch, und **gestellt hat sie niemand**: sie ist aus einem
// Wunsch entstanden, den eine Sitzung verpackt und ein Wirt ausgelegt hat.
func TestEinWunschUeberDieLeitungWirdZumSpielzustand(t *testing.T) {
	t.Parallel()

	lauf := einLeitungslauf(t)

	lauf.passeBis("Berts erste Hauptphase", inHauptphase("sitz-2", 2))

	vorher := len(lauf.gehoert.ereignisse)

	// **Bei den Namen, unter denen Bert sie sieht** (#2480).
	lauf.wuensche(service.BefehlLandspielen,
		map[string]any{"objekt": lauf.sieht("sitz-2/gebirge/0")})
	lauf.wuensche(service.BefehlLandmana,
		map[string]any{"objekt": lauf.sieht("sitz-2/gebirge/0")})
	lauf.wuensche(service.BefehlWirken,
		map[string]any{"objekt": lauf.sieht("sitz-2/wesen/0")})

	lauf.passeBis("den Bären auf dem Tisch",
		inZone("sitz-2/wesen/0", domain.ZoneTisch))

	// **Und Bert hat davon gehört.** Die Tatsachen sind bei ihm angekommen
	// und nicht nur im Speicher des Wirts — sonst spielte er blind.
	if len(lauf.gehoert.ereignisse) <= vorher {
		t.Errorf("beim Gast kamen %d Ereignisse an, vorher waren es %d",
			len(lauf.gehoert.ereignisse), vorher)
	}

	// **Und kein Handgriff, auch nicht über die Leitung.**
	for _, art := range lauf.benutzt {
		if slices.Contains(handgriffe, art) {
			t.Errorf("%q ist ein Handgriff — gestellt wurde %v",
				art, lauf.benutzt)
		}
	}
}

// **Und er hört jeden Zug am Tisch** — die Zusage aus
// #1627.
//
// # Hier stand der Gegenbeweis
//
// **Bis zum 2026-09-19 maß dieser Lauf die Lücke**: Anna spielte einen
// ganzen Zug, und bei Bert kamen **null** Ereignisse und **null**
// Tischbilder an. `Partiewirt.verkuenden` schickte über **seine** Sitzung,
// und ein Wirt gehört zu **einem** Gastsitz; für die Züge des Wirts gab es
// keinen Weg hinaus, und `Partierunde.anJeden` trug nur Chat und Präsenz.
//
// **Sein Vorgänger hat seine eigene Ablösung verlangt** — „wird dieser Lauf
// rot, ist der Befund behoben, und er gehört ersetzt". Er wurde rot, sobald
// [service.Tischfunk] stand, und hier steht der Lauf, der die **Zusage**
// prüft.
func TestEinZugDesWirtsErreichtDenGast(t *testing.T) {
	t.Parallel()

	lauf := einLeitungslauf(t)

	lauf.passeBis("Annas erste Hauptphase", inHauptphase("sitz-1", 1))

	ereignisse := len(lauf.gehoert.ereignisse)
	bilder := len(lauf.gehoert.bilder)

	// **Anna spielt einen ganzen Zug** — drei Befehle, und Bert sieht jeden.
	lauf.stelle("sitz-1", service.BefehlLandspielen,
		map[string]any{"objekt": "sitz-1/gebirge/0"})
	lauf.stelle("sitz-1", service.BefehlLandmana,
		map[string]any{"objekt": "sitz-1/gebirge/0"})
	lauf.stelle("sitz-1", service.BefehlWirken,
		map[string]any{"objekt": "sitz-1/wesen/0"})

	if gekommen := len(lauf.gehoert.ereignisse) - ereignisse; gekommen < 3 {
		t.Errorf("beim Gast kamen %d Ereignisse an — erwartet ist "+
			"mindestens eines je Befehl", gekommen)
	}

	// **Und je Befehl ein Bild**: der Tisch sieht nach jedem Zug anders aus,
	// und der Gast rechnet nichts selbst.
	if gekommen := len(lauf.gehoert.bilder) - bilder; gekommen != 3 {
		t.Errorf("beim Gast kamen %d Tischbilder an statt drei", gekommen)
	}

	// **Das letzte Bild gehört ihm und nicht Anna.** Eine Sicht ist
	// sitzbezogen; dieselben Bytes an alle wären die zweite Lücke gewesen.
	letztes := lauf.gehoert.bilder[len(lauf.gehoert.bilder)-1]
	if letztes.Sicht.Fuer != "sitz-2" {
		t.Errorf("das Bild ist für %q gerechnet", letztes.Sicht.Fuer)
	}
}
