package service_test

// **Die offenen Plätze eines Tisches** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Satz von Anna bei Bert ankommt.** Ohne diese Stelle hatte eine
// Partie mit drei Sitzen drei Zwiegespräche mit dem Wirt und kein gemeinsames
// — der Wirt konnte nur antworten, nie von sich aus etwas schicken.
//
// **Und dass ein langsamer Platz den Tisch nicht anhält.** Wer klagt, wird
// benannt; die übrigen bekommen ihre Nachricht trotzdem.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// platzdoppel ist eine Leitung, die sich merkt, was hinausging.
type platzdoppel struct {
	schloss     sync.Mutex
	gesendet    []domain.Partienachricht
	geleert     int
	geschlossen int

	klage error
}

func (p *platzdoppel) Sende(
	_ context.Context, roh []byte, _ service.Zustellung,
) error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	if p.klage != nil {
		return p.klage
	}

	var nachricht domain.Partienachricht
	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		return fehler
	}

	p.gesendet = append(p.gesendet, nachricht)

	return nil
}

func (p *platzdoppel) Leere(context.Context) error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.geleert++

	return nil
}

// Schliesse zählt, wie oft die Leitung geschlossen wurde.
func (p *platzdoppel) Schliesse() error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.geschlossen++

	return nil
}

// Stand sagt, wie es um diese Leitung steht — die Probe nennt einen Kanal,
// damit ein Lauf sieht, dass er durchgereicht wird.
func (p *platzdoppel) Stand() service.Sitzungsstand {
	return service.Sitzungsstand{Kanal: "probe", Wartend: p.geleert}
}

func (p *platzdoppel) alle() []domain.Partienachricht {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	return append([]domain.Partienachricht(nil), p.gesendet...)
}

// einPlatz baut einen Gastplatz samt seiner Leitung.
func einPlatz(sitz string) (*service.Gastplatz, *platzdoppel) {
	doppel := &platzdoppel{}
	sitzung := service.NeuePartiesitzung(
		"p1", "wirt", doppel, service.NeueReihenfolge(1))

	return service.NeuerGastplatz(sitz, sitzung, doppel), doppel
}

// **Ein Satz geht an alle außer dem, der ihn gesagt hat.**
//
// **Außer ihm**, denn sein Fenster hat ihn beim Absenden angezeigt; ein Echo
// käme als zweiter Satz an und sähe aus wie ein doppeltes Wort.
func TestEinSatzGehtAnAlleAusserDemSprecher(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	anna, annasLeitung := einPlatz("anna")
	bert, bertsLeitung := einPlatz("bert")

	runde.Setze(anna)
	runde.Setze(bert)

	if fehler := runde.Sage(t.Context(), domain.Chat{
		Von: "anna", Text: "gut gezogen",
	}); fehler != nil {
		t.Fatalf("sagen: %v", fehler)
	}

	if bei := annasLeitung.alle(); len(bei) != 0 {
		t.Errorf("anna bekam ihren eigenen Satz zurück: %+v", bei)
	}

	bei := bertsLeitung.alle()
	if len(bei) != 1 {
		t.Fatalf("bert bekam %d Nachrichten statt einer", len(bei))
	}

	if bei[0].Art != domain.ArtChat {
		t.Errorf("Art %q statt Chat", bei[0].Art)
	}

	// **Der Umschlag trägt den Wirt und die Nutzlast die Sprecherin.** Beides
	// gleichzusetzen hieße entweder lügen oder die Herkunft verlieren.
	if bei[0].Von != "wirt" {
		t.Errorf("Absender %q — der Wirt hat diesen Sprung geschickt",
			bei[0].Von)
	}

	var gesagt domain.Chat
	if fehler := json.Unmarshal(bei[0].Inhalt, &gesagt); fehler != nil {
		t.Fatalf("kein Satz: %v", fehler)
	}

	if gesagt.Von != "anna" || gesagt.Text != "gut gezogen" {
		t.Errorf("gesagt hat %q: %q", gesagt.Von, gesagt.Text)
	}

	if bertsLeitung.geleert == 0 {
		t.Error("die Leitung wurde nicht geleert — gesagt und nicht gehört")
	}
}

// **Ein Zeiger bekommt seinen Stand von der Runde, und er steigt je Sprecher.**
//
// # Warum die Runde ihn vergibt und nicht der Absender
//
// **Weil sie den Abriss einer Gastsitzung überlebt.** Zählte der Gast, finge
// er nach einem Neuverbinden wieder bei eins an — und jeder andere Gast wiese
// von da an alles ab, was von ihm kommt.
func TestDieRundeVergibtDenZeigerstandJeSprecher(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	bert, bertsLeitung := einPlatz("bert")
	runde.Setze(bert)

	for _, sprecher := range []string{"wirt", "anna", "wirt"} {
		if fehler := runde.Zeige(t.Context(), domain.Praesenz{
			Von: sprecher, Art: domain.PraesenzZeiger, X: 10, Y: 20,
		}); fehler != nil {
			t.Fatalf("zeigen: %v", fehler)
		}
	}

	staende := map[string][]uint64{}

	for _, nachricht := range bertsLeitung.alle() {
		var gezeigt domain.Praesenz
		if fehler := json.Unmarshal(nachricht.Inhalt, &gezeigt); fehler != nil {
			t.Fatalf("keine Präsenz: %v", fehler)
		}

		staende[gezeigt.Von] = append(staende[gezeigt.Von], gezeigt.Stand)
	}

	if got := staende["wirt"]; len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("die Stände des Wirts sind %v, erwartet [1 2]", got)
	}

	if got := staende["anna"]; len(got) != 1 || got[0] != 1 {
		t.Errorf("annas Stände sind %v, erwartet [1] — jeder zählt für sich",
			got)
	}
}

// **Ein Platz, der klagt, hält die übrigen nicht auf — und wird benannt.**
func TestEinKlagenderPlatzHaeltDenTischNichtAn(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	anna, annasLeitung := einPlatz("anna")
	bert, bertsLeitung := einPlatz("bert")

	bertsLeitung.klage = errors.New("die Leitung ist zu")

	runde.Setze(anna)
	runde.Setze(bert)

	fehler := runde.Sage(t.Context(), domain.Chat{Von: "wirt", Text: "hallo"})
	if fehler == nil {
		t.Fatal("die Klage von bert wurde verschluckt")
	}

	if !strings.Contains(fehler.Error(), "bert") {
		t.Errorf("die Klage nennt nicht, wer sie erhoben hat: %v", fehler)
	}

	if len(annasLeitung.alle()) != 1 {
		t.Error("anna bekam nichts, weil bert klagte")
	}
}

// **Eine abgelöste Sitzung nimmt die neue nicht mit.**
//
// # Warum das kein Sonderfall ist
//
// **Weil eine abgerissene Sitzung endet, nachdem die neue schon steht.** Der
// Gast verbindet neu, sein frischer Platz wird eingetragen, und erst danach
// läuft die alte Schleife aus. Wer dann blind löschte, nähme den lebenden
// Platz weg — und der Gast säße am Tisch, ohne etwas zu hören.
func TestEineAbgeloesteSitzungNimmtDieNeueNichtMit(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	alt, _ := einPlatz("anna")
	neu, neueLeitung := einPlatz("anna")

	runde.Setze(alt)
	runde.Setze(neu)

	runde.Geht(alt)

	if sitze := runde.Sitze(); len(sitze) != 1 || sitze[0] != "anna" {
		t.Fatalf("besetzt sind %v — anna ist weg, obwohl sie gerade kam", sitze)
	}

	if fehler := runde.Sage(t.Context(), domain.Chat{
		Von: "wirt", Text: "willkommen zurück",
	}); fehler != nil {
		t.Fatalf("sagen: %v", fehler)
	}

	if len(neueLeitung.alle()) != 1 {
		t.Error("die frische Sitzung bekam nichts")
	}

	runde.Geht(neu)

	if sitze := runde.Sitze(); len(sitze) != 0 {
		t.Errorf("nach dem Aufstehen sind noch %v besetzt", sitze)
	}
}

// **Ein Rundruf an einen leeren Tisch geht ins Leere und klagt nicht.**
func TestEinLeererTischNimmtEinenSatzOhneKlage(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	if fehler := runde.Sage(t.Context(), domain.Chat{
		Von: "wirt", Text: "jemand da?",
	}); fehler != nil {
		t.Errorf("ein leerer Tisch klagte: %v", fehler)
	}

	if sitze := runde.Sitze(); len(sitze) != 0 {
		t.Errorf("ein leerer Tisch nennt Sitze: %v", sitze)
	}
}

// **Was keine Gestalt hat, geht gar nicht erst hinaus.**
func TestDieRundeWeistUngueltigesVorDemSendenAb(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	bert, bertsLeitung := einPlatz("bert")
	runde.Setze(bert)

	if fehler := runde.Sage(t.Context(), domain.Chat{Von: "wirt"}); fehler == nil {
		t.Error("ein leerer Satz ging hinaus")
	}

	if fehler := runde.Zeige(t.Context(), domain.Praesenz{
		Von: "wirt", Art: domain.Praesenzart("winken"),
	}); fehler == nil {
		t.Error("eine erfundene Präsenzart ging hinaus")
	}

	if bei := bertsLeitung.alle(); len(bei) != 0 {
		t.Errorf("bert bekam %d Nachrichten, die nie hätten hinausgehen "+
			"dürfen", len(bei))
	}
}

// **Die Runde sagt, wie es um jede Leitung steht** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// **Nach Sitz geordnet**, denn die Antwort steht in einem Fenster: eine Liste,
// deren Reihenfolge bei jedem Ruf eine andere ist, lässt sich nicht lesen.
func TestDieRundeNenntJedeLeitung(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	for _, sitz := range []string{"cem", "anna", "bert"} {
		platz, _ := einPlatz(sitz)
		runde.Setze(platz)
	}

	leitungen := runde.Leitungen()

	if len(leitungen) != 3 {
		t.Fatalf("erwartet drei Leitungen, bekam %v", leitungen)
	}

	for stelle, sitz := range []string{"anna", "bert", "cem"} {
		if leitungen[stelle].Kennung != sitz {
			t.Fatalf("an Stelle %d steht %q", stelle, leitungen[stelle].Kennung)
		}
	}
}

// **Und eine leere Runde gibt eine leere Liste.**
func TestEineLeereRundeNenntKeineLeitung(t *testing.T) {
	t.Parallel()

	var runde service.Partierunde

	if leitungen := runde.Leitungen(); leitungen == nil || len(leitungen) != 0 {
		t.Fatalf("erwartet eine leere Liste, bekam %v", leitungen)
	}
}
