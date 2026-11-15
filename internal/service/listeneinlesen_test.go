package service_test

// **Was beim Übernehmen einer Liste geschieht** — Werkstück 2 von Phase 4.5
// (#1292).

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probeprinting ist eine Druckkennung in der Gestalt, die Scryfall führt.
const probeprinting = "e768c957-3a1f-42f5-853a-96942f645df5"

// listenzeile gibt eine Übernahme mit den genannten Abweichungen.
func listenzeile(teile ...func(*service.Listenzeile)) service.Listenzeile {
	zeile := service.Listenzeile{
		Printing: probeprinting,
		Menge:    4,
		Zustand:  domain.ZustandNearMint,
		Sprache:  domain.Vorgabesprache,
		Name:     "Lightning Bolt",
		Setcode:  "m11",
		Nummer:   "149",
	}

	for _, teil := range teile {
		teil(&zeile)
	}

	return zeile
}

// einfuhr baut eine Einfuhr über einem leeren Bestand.
func einfuhr(ablage *sammlungsdoppel) *service.Listeneinfuhr {
	return service.NeueListeneinfuhr(&listenquelle{}, ablage)
}

// **Was übernommen wird, steht danach im Bestand.**
func TestUebernehmenLegtAn(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	bericht, fehler := einfuhr(ablage).Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile()})
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Fatalf("%d angelegt, erwartet 1", len(bericht.Angelegt))
	}

	if bericht.Stueck != 4 {
		t.Errorf("%d Stück, erwartet 4", bericht.Stueck)
	}

	zeilen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	if len(zeilen) != 1 || zeilen[0].Menge != 4 {
		t.Errorf("im Bestand steht %+v", zeilen)
	}
}

// **Dieselbe Zeile ein zweites Mal erhöht die Menge — sie legt nichts an.**
//
// Der Bestand hat eine Eindeutigkeit über Druck, Zustand, Sprache und
// Ausführung, und `Anlegen` gibt darauf einen Konflikt. Beim Einlesen einer
// zweiten Liste ist das nicht der gemeinte Fall: wer zweimal dieselbe Karte
// einliest, hat mehr davon.
func TestEinZweitesMalErhoehtStattAbzuweisen(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	weg := einfuhr(ablage)

	if _, fehler := weg.Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile()}); fehler != nil {
		t.Fatalf("erste Einfuhr: %v", fehler)
	}

	bericht, fehler := weg.Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile(func(z *service.Listenzeile) { z.Menge = 3 })})
	if fehler != nil {
		t.Fatalf("zweite Einfuhr: %v", fehler)
	}

	if bericht.Erhoeht != 1 || len(bericht.Angelegt) != 0 {
		t.Errorf("erhöht %d, angelegt %d — erwartet 1 und 0",
			bericht.Erhoeht, len(bericht.Angelegt))
	}

	zeilen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	if len(zeilen) != 1 {
		t.Fatalf("%d Zeilen im Bestand, erwartet 1", len(zeilen))
	}

	if zeilen[0].Menge != 7 {
		t.Errorf("Menge %d, erwartet 7", zeilen[0].Menge)
	}
}

// **Ein anderer Zustand ist eine andere Zeile.**
//
// Die Gegenprobe zur Regel darüber: verglichen werden die **vier** Spalten der
// Eindeutigkeit. Ohne diesen Test wäre ein Vergleich, der nur den Druck
// ansieht, grün — und ein Foil landete auf dem Nonfoil-Stapel.
func TestEinAndererZustandIstEineAndereZeile(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	weg := einfuhr(ablage)

	for _, teil := range []func(*service.Listenzeile){
		func(_ *service.Listenzeile) {},
		func(z *service.Listenzeile) { z.Zustand = domain.ZustandPlayed },
		func(z *service.Listenzeile) { z.Finish = domain.FinishFoil },
		func(z *service.Listenzeile) { z.Sprache = "de" },
	} {
		if _, fehler := weg.Einlesen(t.Context(),
			[]service.Listenzeile{listenzeile(teil)}); fehler != nil {
			t.Fatalf("Uebernehmen: %v", fehler)
		}
	}

	zeilen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	if len(zeilen) != 4 {
		t.Errorf("%d Zeilen im Bestand, erwartet 4", len(zeilen))
	}
}

// **Eine Zeile, die nicht durchgeht, hält die übrigen nicht auf** (§8).
func TestEineAbgewieseneZeileHaeltDenRestNichtAuf(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	bericht, fehler := einfuhr(ablage).Einlesen(t.Context(), []service.Listenzeile{
		listenzeile(func(z *service.Listenzeile) { z.Printing = "" }),
		listenzeile(),
	})
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 {
		t.Errorf("%d angelegt, erwartet 1", len(bericht.Angelegt))
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}

	// **Der Befund zeigt auf die Zeile**, sonst weiß niemand, welche es war.
	if bericht.Befunde[0].Ort != "Zeile 1" {
		t.Errorf("der Befund zeigt auf %q", bericht.Befunde[0].Ort)
	}
}

// **Ohne eine einzige Zeile ist es ein fachliches Nein.**
func TestUebernehmenOhneZeilen(t *testing.T) {
	t.Parallel()

	_, fehler := einfuhr(&sammlungsdoppel{}).Einlesen(t.Context(), nil)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
	}
}

// **Ein Lesefehler des Bestands kommt als Fehler heraus.**
//
// Nicht als Befund: ein Plattenfehler sagt über jede Zeile dasselbe, und
// dreihundert gleichlautende Befunde sind keine Auskunft (§8).
func TestEinLesefehlerBeimUebernehmen(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	bericht, fehler := einfuhr(&sammlungsdoppel{fehler: gestolpert}).
		Einlesen(t.Context(), []service.Listenzeile{listenzeile()})

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}

	if len(bericht.Angelegt) != 0 {
		t.Errorf("%d angelegt, erwartet 0", len(bericht.Angelegt))
	}
}

// **Ein Schreibfehler beim Erhöhen ebenso.**
//
// Der Zweig, der ohne diesen Test unbetreten bliebe: die vorhandene Zeile ist
// gefunden, und erst das Schreiben scheitert.
func TestEinSchreibfehlerBeimErhoehen(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}
	weg := einfuhr(ablage)

	if _, fehler := weg.Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile()}); fehler != nil {
		t.Fatalf("erste Einfuhr: %v", fehler)
	}

	gestolpert := errors.New("die Platte klemmt")
	ablage.fehler = gestolpert

	bericht, fehler := weg.Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile()})

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}

	if bericht.Erhoeht != 0 {
		t.Errorf("%d erhöht, erwartet 0", bericht.Erhoeht)
	}
}

// nurSchreibfehler liest willig und scheitert beim Schreiben.
//
// **Ein eigenes Doppel und kein Schalter am vorhandenen.** `sammlungsdoppel`
// hat **einen** Fehler für jeden Weg; damit scheitert schon `Fuer`, und
// `Anlegen` und `Aendern` werden nie gerufen. Der erste Anlauf dieser beiden
// Tests hat genau das getan — sie waren grün und haben den Zweig, den sie
// meinten, **nicht betreten** (§10: eine Gegenprobe, die nichts ändert, ist
// keine).
type nurSchreibfehler struct {
	*sammlungsdoppel

	beimSchreiben error
}

func (n *nurSchreibfehler) Anlegen(_ context.Context, _ *domain.Exemplar) (int64, error) {
	return 0, n.beimSchreiben
}

func (n *nurSchreibfehler) Aendern(_ context.Context, _ *domain.Exemplar) error {
	return n.beimSchreiben
}

// **Ein Konflikt beim Anlegen wird zum Befund** — und der Rest läuft weiter.
//
// Ein Konflikt ist eine Aussage über **diese** Zeile: sie gibt es schon. Das
// ist der Fall, für den es Befunde gibt (§8) — anders als ein Plattenfehler,
// der über jede Zeile dasselbe sagt.
func TestEinKonfliktBeimAnlegen(t *testing.T) {
	t.Parallel()

	ablage := &nurSchreibfehler{
		sammlungsdoppel: &sammlungsdoppel{},
		beimSchreiben:   fmt.Errorf("%w: die gibt es schon", domain.ErrConflict),
	}

	bericht, fehler := service.NeueListeneinfuhr(&listenquelle{}, ablage).
		Einlesen(t.Context(), []service.Listenzeile{listenzeile()})
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if len(bericht.Angelegt) != 0 {
		t.Errorf("%d angelegt, erwartet 0", len(bericht.Angelegt))
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}
}

// **Ein Konflikt beim Erhöhen ebenso.**
//
// Der Unterschied zum Test darüber ist der Zweig: hier steht die Zeile schon
// im Bestand, und erst das Zurückschreiben scheitert.
func TestEinKonfliktBeimErhoehen(t *testing.T) {
	t.Parallel()

	darunter := &sammlungsdoppel{}

	if _, fehler := darunter.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
	}); fehler != nil {
		t.Fatalf("vorbereiten: %v", fehler)
	}

	ablage := &nurSchreibfehler{
		sammlungsdoppel: darunter,
		beimSchreiben:   fmt.Errorf("%w: die Fassung stimmt nicht", domain.ErrConflict),
	}

	bericht, fehler := service.NeueListeneinfuhr(&listenquelle{}, ablage).
		Einlesen(t.Context(), []service.Listenzeile{listenzeile()})
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if bericht.Erhoeht != 0 {
		t.Errorf("%d erhöht, erwartet 0", bericht.Erhoeht)
	}

	if len(bericht.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(bericht.Befunde), bericht.Befunde)
	}
}

// **Ein Systemfehler bricht ab — ein Eingabefehler nicht.**
//
// Der Unterschied ist der Punkt: bei dreihundert Zeilen und einer klemmenden
// Platte bekäme der Aufrufer sonst dreihundert gleichlautende Befunde und
// **kein** `error`, obwohl gar nichts geschrieben wurde. Ein Plattenfehler ist
// keine Aussage über eine Zeile (§8).
func TestEinSystemfehlerBrichtAb(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Platte klemmt")

	ablage := &nurSchreibfehler{
		sammlungsdoppel: &sammlungsdoppel{},
		beimSchreiben:   gestolpert,
	}

	bericht, fehler := service.NeueListeneinfuhr(&listenquelle{}, ablage).
		Einlesen(t.Context(), []service.Listenzeile{listenzeile(), listenzeile()})

	if !errors.Is(fehler, gestolpert) {
		t.Errorf("Fehler %v, erwartet %v", fehler, gestolpert)
	}

	// **Und der Bericht kommt trotzdem** (§7, Regel 3): was geschrieben ist,
	// ist geschrieben — hier nichts, und das steht auch so da.
	if len(bericht.Angelegt) != 0 {
		t.Errorf("%d angelegt, erwartet 0", len(bericht.Angelegt))
	}

	// **Ein Befund je Zeile wäre falsch.** Abgebrochen wird bei der ersten.
	if len(bericht.Befunde) != 0 {
		t.Errorf("%d Befunde, erwartet 0: %v", len(bericht.Befunde), bericht.Befunde)
	}
}

// **Eine Listeneinfuhr erhöht keinen Wunsch.**
//
// Die zweite Hälfte von Befund 1 des Zweitblicks über Phase 4.5: wer eine
// Karte auf der Wunschliste hat und sie dann **kauft**, legt Besitz an — er
// erhöht nicht seinen Wunsch.
func TestEineListeneinfuhrErhoehtKeinenWunsch(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{}

	if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 2,
		Zustand: domain.ZustandNearMint, Sprache: domain.Vorgabesprache,
		Wunsch: true,
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	bericht, fehler := einfuhr(ablage).Einlesen(t.Context(),
		[]service.Listenzeile{listenzeile()})
	if fehler != nil {
		t.Fatalf("Einlesen: %v", fehler)
	}

	if len(bericht.Angelegt) != 1 || bericht.Erhoeht != 0 {
		t.Errorf("%d angelegt, %d erhöht — erwartet 1 und 0",
			len(bericht.Angelegt), bericht.Erhoeht)
	}

	zeilen, lesefehler := ablage.Liste(t.Context())
	if lesefehler != nil {
		t.Fatalf("die Sammlung lesen: %v", lesefehler)
	}

	if len(zeilen) != 2 {
		t.Errorf("%d Zeilen, erwartet 2", len(zeilen))
	}
}
