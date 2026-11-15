package app

// **Wiederverbindung und Zustandsabgleich, gemessen am Tisch** — D01
// (#2232).
//
// Zwei Anwendungen mit eigenem Bestand und eigener Identität, der Wirt horcht
// auf einem **echten Port**, der Gast wählt ihn über ein [relais], das ein Netz
// spielt, das stockt oder abreißt. Darüber TLS 1.3, Rahmung, Handschlag,
// Partiesitzung, Tischfunk — derselbe Weg wie im Betrieb.
//
// # Drei Fälle
//
//  1. **Kurzer Ausfall**: das Relais hält zwei Sekunden an; beide ziehen
//     währenddessen. Die Leitung überlebt.
//  2. **Langer Ausfall**: das Relais kappt; der Gast pausiert, ein Wunsch
//     währenddessen scheitert **laut**, der Wirt zieht weiter. Nach der
//     Freigabe kehrt der Gast zurück (`PartieBeitreten`).
//  3. **Veralteter Gaststand**: der Gast hat eine Karte des Wirts offen auf
//     dem Tisch gesehen; während er fort ist, nimmt der Wirt sie auf die
//     Hand und zieht mehr, als eine Sitzung nachreicht. Nach der Rückkehr
//     darf ihr Name beim Gast **nicht** mehr stehen.
//
// # Was gemessen wird
//
// **Startereignis** ist die Freigabe des Relais (Fall 1) beziehungsweise die
// Rückkehr des Gastes nach der Freigabe (Fall 2, 3); **Endereignis** ist das
// erste Tischbild beim Gast, das **gleich** der Sicht ist, die der Wirt für
// Berts Sitz rechnet — die ganze `PartieDTO`, mit Fassung. **Frist: 10 s**.
// Dazu je Fall: der Verlauf des Wirts wächst um genau die angenommenen
// Befehle, und kein geheimes Objekt trägt beim Gast Name, Karte oder Bild.
//
// # Was er nicht zeigt
//
// Zwei Rechner, ein echtes Netz mit Verlust und Umordnung unterhalb von TCP,
// Schlafen und Aufwachen. Die Zahlen gelten für die Rückschleife.

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// konvergenzfrist ist die Zusage aus D01.
const konvergenzfrist = 10 * time.Second

// gastsicht hält das letzte Tischbild des Gastes fest.
type gastsicht struct {
	schloss sync.Mutex
	letzte  *PartieDTO
}

func (g *gastsicht) melden(_ context.Context, ereignis string, daten ...any) {
	if ereignis != EreignisPartietischbild || len(daten) < 2 {
		return
	}

	antwort, ist := daten[1].(PartieAntwort)
	if !ist {
		return
	}

	g.schloss.Lock()
	defer g.schloss.Unlock()

	g.letzte = &antwort.Partie
}

func (g *gastsicht) jetzt() *PartieDTO {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	return g.letzte
}

// tischUeberRelais setzt den Gast über das Relais an den Tisch des Wirts.
type tischUeberRelais struct {
	wirt, gast *seite
	sicht      *gastsicht
	relais     *relais
	partieID   string
	stellen    befehlsweg
	bertsKarte string
	annasKarte string
}

func einTischUeberEinRelais(t *testing.T) *tischUeberRelais {
	t.Helper()

	wirt, gast := zweiUeberTCP(t)
	sicht := &gastsicht{}
	gast.anwendung.melden = sicht.melden

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, bertsKarte := einTischBeimWirt(t, wirt)
	stellen := ueberDenWirt(wirt)

	// **Eine Karte aus Annas Hand liegt offen auf dem Tisch** — der Gast
	// sieht ihren Namen, und Fall 3 nimmt sie ihm wieder.
	annasKarte := inZone(t, wirt.anwendung.Partie(wirt.ctx, partieID, "anna").Partie.Objekte, "anna", "hand")
	befehl(t, stellen, partieID, "vorher/auf-den-tisch", "objekt.schieben",
		map[string]any{"objekt": annasKarte, "zone": "tisch"})

	if antwort := wirt.anwendung.PartieFreigeben(wirt.ctx, partieID, "anna", []string{"bert"}); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	gestartet := wirt.anwendung.HorchStarten(wirt.ctx, "127.0.0.1:0")
	if gestartet.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", gestartet.Fehler)
	}

	t.Cleanup(wirt.anwendung.horchenBeenden)

	zwischen := neuesRelais(t, gestartet.Stand.Adresse)
	if antwort := gast.anwendung.PeerkontaktAdresse(gast.ctx, string(wirt.kennung), zwischen.adresse()); antwort.Fehler != nil {
		t.Fatalf("PeerkontaktAdresse: %+v", antwort.Fehler)
	}

	tisch := &tischUeberRelais{
		wirt: wirt, gast: gast, sicht: sicht, relais: zwischen, partieID: partieID,
		stellen: stellen, bertsKarte: bertsKarte, annasKarte: annasKarte,
	}

	beginn := time.Now()
	tisch.zurueck(t)

	if dauer := tisch.konvergiert(t, "der erste Beitritt", beginn, func(PartieDTO) bool { return true }); dauer > konvergenzfrist {
		t.Fatalf("schon der erste Beitritt brauchte %v", dauer)
	}

	return tisch
}

// zurueck setzt den Gast (wieder) an den Tisch.
func (tisch *tischUeberRelais) zurueck(t *testing.T) {
	t.Helper()

	if antwort := tisch.gast.anwendung.PartieBeitreten(
		tisch.gast.ctx, string(tisch.wirt.kennung), tisch.partieID, PartiemitbringDTO{},
	); antwort.Fehler != nil {
		t.Fatalf("PartieBeitreten: %+v", antwort.Fehler)
	}
}

// wirtssicht ist, was der Wirt für Berts Sitz rechnet.
func (tisch *tischUeberRelais) wirtssicht(t *testing.T) PartieDTO {
	t.Helper()

	antwort := tisch.wirt.anwendung.Partie(tisch.wirt.ctx, tisch.partieID, "bert")
	if antwort.Fehler != nil {
		t.Fatalf("Partie für bert: %+v", antwort.Fehler)
	}

	return antwort.Partie
}

// konvergiert wartet, bis die Sicht des Gastes der des Wirts gleicht **und**
// das erwartete Ende zeigt, und gibt die Dauer seit `beginn` — oder bricht
// nach der Frist mit dem Unterschied ab.
//
// **Die Bedingung ist nicht entbehrlich**: zwischen zwei Zügen können beide
// Sichten kurz gleich sein, bevor der zweite verarbeitet ist.
func (tisch *tischUeberRelais) konvergiert(
	t *testing.T, fall string, beginn time.Time, ende func(PartieDTO) bool,
) time.Duration {
	t.Helper()

	for {
		soll := tisch.wirtssicht(t)
		ist := tisch.sicht.jetzt()

		if ist != nil && ende(*ist) && reflect.DeepEqual(*ist, soll) {
			dauer := time.Since(beginn)
			t.Logf("gemessen: %s — gleich nach %v (Fassung %s)", fall, dauer.Round(time.Millisecond), soll.Fassung)
			nichtsVerdecktesSichtbar(t, fall, *ist)

			return dauer
		}

		if time.Since(beginn) > konvergenzfrist {
			roh, _ := json.Marshal(ist)      //nolint:errcheck,errchkjson // nur für die Meldung
			sollroh, _ := json.Marshal(soll) //nolint:errcheck,errchkjson // nur für die Meldung
			t.Fatalf("%s: nach %v nicht gleich\nGast: %.600s\nWirt: %.600s", fall, konvergenzfrist, roh, sollroh)
		}

		time.Sleep(20 * time.Millisecond)
	}
}

// verlaufslaenge zählt die Ereignisse des Wirts.
func (tisch *tischUeberRelais) verlaufslaenge(t *testing.T) int {
	t.Helper()

	antwort := tisch.wirt.anwendung.PartieVerlauf(tisch.wirt.ctx, tisch.partieID, "anna")
	if antwort.Fehler != nil {
		t.Fatalf("PartieVerlauf: %+v", antwort.Fehler)
	}

	return len(antwort.Verlauf)
}

// nichtsVerdecktesSichtbar prüft, dass kein geheimes Objekt beim Gast etwas verrät.
func nichtsVerdecktesSichtbar(t *testing.T, fall string, sicht PartieDTO) {
	t.Helper()

	for _, eines := range sicht.Objekte {
		if eines.Geheim && (eines.Name != "" || eines.Karte != "" || eines.KarteID != "" || eines.Eigenschaften != nil) {
			t.Errorf("%s: das geheime Objekt %s verrät %q/%q", fall, eines.Kennung, eines.Name, eines.Karte)
		}
	}
}

func befehl(t *testing.T, stellen befehlsweg, partieID, kennung, art string, nutzlast map[string]any) {
	t.Helper()

	if antwort := stellen(domain.Spielbefehl{
		PartieID: partieID, Kennung: kennung, Absender: "anna", Art: art, Nutzlast: nutzlast,
	}); antwort.Fehler != nil {
		t.Fatalf("%s: %+v", kennung, antwort.Fehler)
	}
}

func inZone(t *testing.T, objekte []ObjektDTO, wer, zone string) string {
	t.Helper()

	for _, eines := range objekte {
		if eines.Eigner == wer && eines.Zone == zone {
			return eines.Kennung
		}
	}

	t.Fatalf("%s hat nichts in %s", wer, zone)

	return ""
}

func TestWiederverbindungNachKurzemAusfall(t *testing.T) {
	t.Parallel()

	tisch := einTischUeberEinRelais(t)
	vorher := tisch.verlaufslaenge(t)

	tisch.relais.anhalten()

	befehl(t, tisch.stellen, tisch.partieID, "kurz/wirt", "objekt.tappen", map[string]any{"objekt": tisch.annasKarte})

	gewuenscht := make(chan *Fehler, 1)

	go func() {
		gewuenscht <- tisch.gast.anwendung.PartieWuenschen(tisch.gast.ctx, domain.Spielbefehl{
			Kennung: "kurz/gast", Art: "objekt.tappen", Nutzlast: map[string]any{"objekt": tisch.bertsKarte},
		}).Fehler
	}()

	time.Sleep(2 * time.Second)

	// **Das Relais hat wirklich zurückgehalten** — sonst stünde der Zug des
	// Wirts schon vor der Freigabe beim Gast, und die Messung danach
	// mäße nichts.
	if anna, _ := objektIn(tisch.sicht.jetzt(), tisch.annasKarte); anna.Getappt {
		t.Fatal("der Zug des Wirts kam durch ein angehaltenes Relais")
	}

	beginn := time.Now()
	tisch.relais.freigeben()

	if dauer := tisch.konvergiert(t, "kurzer Ausfall (2 s angehalten)", beginn, func(sicht PartieDTO) bool {
		anna, _ := objektIn(&sicht, tisch.annasKarte)
		bert, _ := objektIn(&sicht, tisch.bertsKarte)

		return anna.Getappt && bert.Getappt
	}); dauer > konvergenzfrist {
		t.Errorf("kurzer Ausfall: %v", dauer)
	}

	if fehler := <-gewuenscht; fehler != nil {
		t.Fatalf("der Wunsch während des Anhaltens: %+v", fehler)
	}

	if danach := tisch.verlaufslaenge(t); danach-vorher != 2 {
		t.Errorf("der Verlauf wuchs um %d Ereignisse, angenommen waren 2 Befehle", danach-vorher)
	}
}

func TestWiederverbindungNachLangemAusfall(t *testing.T) {
	t.Parallel()

	tisch := einTischUeberEinRelais(t)
	vorher := tisch.verlaufslaenge(t)

	tisch.relais.kappen()
	warteAufLeerenTisch(t, tisch.gast)

	if stand := tisch.gast.anwendung.PartieBeitrittStand(tisch.gast.ctx); !stand.Pausiert {
		t.Fatalf("nach dem Abriss nicht pausiert: %+v", stand)
	}

	// **Ein Wunsch ins Leere scheitert laut** — und wird nicht aufgehoben,
	// um nach der Rückkehr doppelt zu wirken.
	if antwort := tisch.gast.anwendung.PartieWuenschen(tisch.gast.ctx, domain.Spielbefehl{
		Kennung: "lang/verloren", Art: "objekt.tappen", Nutzlast: map[string]any{"objekt": tisch.bertsKarte},
	}); antwort.Fehler == nil {
		t.Error("ein Wunsch ohne Tisch kam ohne Fehler zurück")
	}

	befehl(t, tisch.stellen, tisch.partieID, "lang/wirt", "objekt.tappen", map[string]any{"objekt": tisch.annasKarte})
	time.Sleep(3 * time.Second)

	beginn := time.Now()
	tisch.relais.freigeben()
	tisch.zurueck(t)

	if dauer := tisch.konvergiert(t, "langer Ausfall (gekappt, Rückkehr)", beginn, func(sicht PartieDTO) bool {
		anna, _ := objektIn(&sicht, tisch.annasKarte)

		return anna.Getappt
	}); dauer > konvergenzfrist {
		t.Errorf("langer Ausfall: %v", dauer)
	}

	if danach := tisch.verlaufslaenge(t); danach-vorher != 1 {
		t.Errorf("der Verlauf wuchs um %d Ereignisse, angenommen war 1 Befehl", danach-vorher)
	}

	if karte, da := objektIn(tisch.sicht.jetzt(), tisch.bertsKarte); !da || karte.Getappt {
		t.Error("der Wunsch ins Leere hat nach der Rückkehr gewirkt")
	}
}

func TestWiederverbindungMitVeraltetemGaststand(t *testing.T) {
	t.Parallel()

	tisch := einTischUeberEinRelais(t)

	if karte, da := objektIn(tisch.sicht.jetzt(), tisch.annasKarte); !da || karte.Karte == "" || karte.Geheim {
		t.Fatalf("der Gast sah Annas Karte auf dem Tisch nicht offen: %+v", karte)
	}

	hand, _ := verborgenIn(tisch.sicht.jetzt(), "anna", "hand")

	tisch.relais.kappen()
	warteAufLeerenTisch(t, tisch.gast)

	vorher := tisch.verlaufslaenge(t)

	befehl(t, tisch.stellen, tisch.partieID, "alt/auf-die-hand", "objekt.schieben",
		map[string]any{"objekt": tisch.annasKarte, "zone": "hand"})

	andere := inZone(t, tisch.wirt.anwendung.Partie(tisch.wirt.ctx, tisch.partieID, "anna").Partie.Objekte, "anna", "bibliothek")
	befehl(t, tisch.stellen, tisch.partieID, "alt/auf-den-tisch", "objekt.schieben",
		map[string]any{"objekt": andere, "zone": "tisch"})

	zuege := 2 * Sitzungstiefe
	for zug := range zuege {
		art := "objekt.tappen"
		if zug%2 == 1 {
			art = "objekt.enttappen"
		}

		befehl(t, tisch.stellen, tisch.partieID, "alt/zug-"+string(rune('a'+zug)), art, map[string]any{"objekt": andere})
	}

	beginn := time.Now()
	tisch.relais.freigeben()
	tisch.zurueck(t)

	// **Unter neuem Namen** (#2480): verborgen ist sie ein neues Objekt.
	if dauer := tisch.konvergiert(t, "veralteter Gaststand", beginn, func(sicht PartieDTO) bool {
		_, noch := objektIn(&sicht, tisch.annasKarte)
		jetzt, _ := verborgenIn(&sicht, "anna", "hand")

		return !noch && jetzt == hand+1
	}); dauer > konvergenzfrist {
		t.Errorf("veralteter Gaststand: %v", dauer)
	}

	if _, offen := verborgenIn(tisch.sicht.jetzt(), "anna", "hand"); offen {
		t.Error("Annas Hand steht beim Gast offen")
	}

	if danach := tisch.verlaufslaenge(t); danach-vorher != 2+zuege {
		t.Errorf("der Verlauf wuchs um %d Ereignisse, angenommen waren %d Befehle", danach-vorher, 2+zuege)
	}
}
