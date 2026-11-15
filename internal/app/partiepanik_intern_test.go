package app

// **Ein Tisch, der zusammenbricht, nimmt das Programm nicht mit** — BK-05a
// (#2111).
//
// Beide Seiten eines Tisches laufen in eigenen Goroutinen: der Gast in
// `amTischSitzen`, der Wirt je Gast in `partieBedienen`. Eine Panik dort fing
// niemand — der Prozess endete, mitsamt jeder anderen Partie, jedem Handel
// und der offenen Bearbeitung. Ausgelöst wird sie hier über den Meldeweg ans
// Fenster, durch den jedes Tischbild geht; er wird gesetzt, **bevor** etwas
// läuft (siehe [seite]).

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zerbrichtBei lässt den Meldeweg einer Seite bei diesen Ereignissen
// zusammenbrechen, sobald `scharf` steht — und zählt, wie oft.
func zerbrichtBei(s *seite, scharf *atomic.Bool, gebrochen *atomic.Int32, ereignisse ...string) {
	s.anwendung.melden = func(ctx context.Context, ereignis string, daten ...any) {
		if scharf.Load() {
			for _, gemeint := range ereignisse {
				if ereignis == gemeint {
					gebrochen.Add(1)
					panic("das Fenster bricht beim Melden zusammen")
				}
			}
		}

		s.sammeln(ctx, ereignis, daten...)
	}
}

// amTischSetzen bringt den Gast an den Tisch des Wirts.
func amTischSetzen(t *testing.T, wirt, gast *seite) (partieID, objekt string) {
	t.Helper()

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, objekt = einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	if antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), partieID, PartiemitbringDTO{},
	); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	return partieID, objekt
}

// wartetBisDerGastNichtMehrSitzt wartet, bis der Beitritt nicht mehr läuft.
func wartetBisDerGastNichtMehrSitzt(t *testing.T, gast *seite) PartiebeitrittAntwort {
	t.Helper()

	frist := time.Now().Add(5 * time.Second)

	for {
		stand := gast.anwendung.PartieBeitrittStand(gast.ctx)
		if !stand.Laeuft || time.Now().After(frist) {
			return stand
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// **Beim Gast**: das erste Tischbild des Wirts bricht den Meldeweg — der Tisch
// endet, der Gast sitzt danach nicht mehr, und das Programm antwortet weiter.
func TestEinZusammenbruchAmTischDesGastesBeendetNurDiesenTisch(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	var (
		scharf    atomic.Bool
		gebrochen atomic.Int32
	)

	zerbrichtBei(gast, &scharf, &gebrochen, EreignisPartietischbild)
	scharf.Store(true)

	amTischSetzen(t, wirt, gast)

	stand := wartetBisDerGastNichtMehrSitzt(t, gast)
	if stand.Laeuft || gebrochen.Load() == 0 {
		t.Fatalf("nach %d Zusammenbrüchen: %+v", gebrochen.Load(), stand)
	}

	// **Wie bei einer abgerissenen Leitung**: der Platz gehört dem Gast
	// weiter, die Partie pausiert (`App.tischVerloren`).
	if !stand.Pausiert {
		t.Errorf("der Tisch endete ohne Pause: %+v", stand)
	}

	// **Und das Programm antwortet weiter.**
	if antwort := gast.anwendung.PartieBeitrittBeenden(gast.ctx); antwort.Fehler != nil {
		t.Errorf("aufstehen danach: %+v", antwort.Fehler)
	}
}

// **Beim Wirt**: der Meldeweg an sein eigenes Fenster bricht, während er
// einen Gast bedient — der Gast verliert diesen Tisch, der Wirt nichts sonst.
func TestEinZusammenbruchBeimWirtBeendetNurDieseSitzung(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	var (
		scharf    atomic.Bool
		gebrochen atomic.Int32
	)

	zerbrichtBei(wirt, &scharf, &gebrochen,
		EreignisEigenesTischbild, EreignisEigenerZug, EreignisPartietischbild, EreignisPartiezug)

	_, objekt := amTischSetzen(t, wirt, gast)

	// **Der Gast zieht** — und was er tut, meldet der Wirt seinem eigenen
	// Fenster (Z1 aus dem Zweitblick über Phase 6.11). Genau dort bricht es.
	scharf.Store(true)

	gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	})

	stand := wartetBisDerGastNichtMehrSitzt(t, gast)
	if stand.Laeuft || gebrochen.Load() == 0 {
		t.Fatalf("nach %d Zusammenbrüchen: %+v", gebrochen.Load(), stand)
	}

	// **Der Wirt antwortet weiter** — und sein Tisch steht noch offen.
	if antwort := wirt.anwendung.PartieFreigabeStand(wirt.ctx); antwort.Fehler != nil || !antwort.Laeuft {
		t.Errorf("der Wirt danach: %+v", antwort)
	}
}

// **Auch ein Zusammenbruch beim Aufräumen bleibt bei dieser Sitzung** —
// Zweitblick über A und B, Befund 1: die Besetzungsmeldung lief in einem
// `defer`, der **vor** dem Fang stand, und ihre Panik erreichte ihn nicht.
func TestEinZusammenbruchBeimAufraeumenDerSitzungBleibtBeiIhr(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	var (
		scharf    atomic.Bool
		gebrochen atomic.Int32
	)

	zerbrichtBei(wirt, &scharf, &gebrochen, EreignisPartiefreigabe)

	amTischSetzen(t, wirt, gast)

	// **Der Gast steht auf** — beim Wirt endet die Sitzung, und ihr
	// Aufräumen meldet die neue Besetzung.
	scharf.Store(true)

	if antwort := gast.anwendung.PartieBeitrittBeenden(gast.ctx); antwort.Fehler != nil {
		t.Fatalf("aufstehen: %+v", antwort.Fehler)
	}

	frist := time.Now().Add(5 * time.Second)
	for gebrochen.Load() == 0 && time.Now().Before(frist) {
		time.Sleep(10 * time.Millisecond)
	}

	if gebrochen.Load() == 0 {
		t.Fatal("die Besetzungsmeldung kam nicht — der Lauf prüft nichts")
	}

	scharf.Store(false)

	// **Der Wirt antwortet weiter.**
	if antwort := wirt.anwendung.PartieFreigabeStand(wirt.ctx); antwort.Fehler != nil || !antwort.Laeuft {
		t.Errorf("der Wirt danach: %+v", antwort)
	}
}
