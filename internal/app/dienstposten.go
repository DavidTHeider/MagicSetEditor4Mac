package app

// **REST und MCP schalten — auf Aufruf** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Was ein Posten hält
//
// **Dieselbe Gestalt wie der Horchposten aus N2**, bis auf eine Zeile: dort
// nimmt ein Handschlag die Verbindung an, hier läuft ein Programm daneben.
// Der Rest ist derselbe — eigener Kontext, `fertig`-Kanal, und Beenden wartet
// wirklich.
//
// **Warum ein Programm daneben und kein Lauscher im Fenster**, steht in
// `dienstwerk.go`; dort steht auch, was es kostet und was dagegen hält.
//
// # Und warum nichts von selbst startet
//
// **Weil [0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)
// es verlangt**: eine sichtbare Zustimmung **je Öffnung** und nicht einmal im
// Leben. Wer keinen Dienst will, hat keinen — auch nach einem Neustart.

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// dienstposten ist ein laufender Lauscher samt allem, was ihn beendet.
type dienstposten struct {
	schloss sync.Mutex

	// adresse ist, wo wirklich gelauscht wird — leer heißt: nirgends.
	//
	// **Die wirkliche und nicht die gewünschte.** Wer Port 0 angibt, bekommt
	// einen vom Betriebssystem, und nur so erfährt ein Mensch, welchen er
	// nennen muss.
	adresse string

	// adressen und fehlende hat der Dienst angesagt (G08).
	adressen []string
	fehlende []domain.Lauschbefund

	// marke ist der Zugang, den dieser Posten verlangt.
	//
	// **Eine je Öffnung und keine je Leben.** Sie entsteht im Kind beim
	// Starten und stirbt mit ihm — wer sie hat, hat sie für diese eine
	// Öffnung.
	marke string

	// erreichbar ist, was ein Gegenüber eintippen kann — leer bei einer
	// Bindung, die den Rechner nicht verlässt.
	erreichbar leitung.Erreichbarkeit

	// fingerabdruck ist der des Zeugnisses — leer ohne TLS.
	fingerabdruck string

	// aussteller und gueltigBis kennt nur ein eigenes Zeugnis (NET-03).
	aussteller string
	gueltigBis time.Time

	// startet sagt, dass gerade einer entsteht.
	//
	// **Ohne dieses Feld ist „läuft schon" eine Prüfung auf einen Stand von
	// vorhin** — Befund 3 des Zweitblicks über Phase 5.4. Das Starten
	// braucht einen Prozess und eine Ansage; solange das läuft, steht
	// `beenden` noch auf `nil`, und ein zweiter Ruf sähe den Posten als
	// frei. Zwei Prozesse gingen auf, und der zweite überschriebe den
	// einzigen Abbruchgriff des ersten — der liefe weiter, auch über das
	// Beenden und den Profilwechsel hinaus (§7, Regel 4).
	startet bool

	beenden func()
	fertig  chan struct{}
}

// dienstfelder hält beide Posten.
type dienstfelder struct {
	rest dienstposten
	mcp  dienstposten
}

// posten gibt den Posten einer Art.
func (d *dienstfelder) posten(art domain.Dienstart) *dienstposten {
	if art == domain.DienstMCP {
		return &d.mcp
	}

	return &d.rest
}

// Dienstzustand ist, was eine Fläche über einen Posten erfährt.
type Dienstzustand struct {
	Art string `json:"art"`

	// Laeuft sagt, ob gerade gelauscht wird.
	Laeuft bool `json:"laeuft"`

	// Adresse ist die **wirkliche** Adresse — leer, solange nichts läuft.
	Adresse string `json:"adresse"`

	// TLS sagt, ob der Lauscher verschlüsselt ist.
	//
	// **Jenseits der Schleife immer**, und das ist keine Wahl (0115).
	//
	// **Gelesen und nicht gemerkt**: die Adresse, die der Dienst über sich
	// sagt, trägt ihr Schema — und zwei Merker über dieselbe Frage liefen
	// auseinander.
	TLS bool `json:"tls"`

	// Marke ist der Zugang, den dieser Posten verlangt.
	//
	// **Sie steht an der Fläche**, weil sie sonst niemand hätte: sie entsteht
	// beim Starten und stirbt mit dem Posten.
	Marke string `json:"marke"`

	// Namen und Adressen sind, was ein Gegenüber eintippen kann.
	Namen    []string `json:"namen"`
	Adressen []string `json:"adressen"`

	// Fingerabdruck ist der des Zeugnisses — genauer: der der Peeridentität,
	// die es trägt.
	//
	// **Er ist der einzige Nachweis, der zählt.** Das Zeugnis ist selbst
	// unterschrieben; wer sich verbindet, vergleicht ihn mit dem, was hier
	// steht (0115).
	Fingerabdruck string `json:"fingerabdruck"`

	// ZeugnisAussteller und ZeugnisBis nennt nur ein **eigenes** Zeugnis
	// (NET-03, #1707) — beim Zeugnis auf der Peeridentität bleiben sie leer.
	// Das Ende steht in RFC 3339, damit das Fenster es in seiner Sprache
	// schreibt.
	ZeugnisAussteller string `json:"zeugnisAussteller"`
	ZeugnisBis        string `json:"zeugnisBis"`

	// Lauscher sind alle Adressen, auf denen der Dienst lauscht — die erste
	// ist [Dienstzustand.Adresse]; Fehlende die gewählten, auf denen kein
	// Lauscher entstand (G08, #2272).
	Lauscher []string          `json:"lauscher"`
	Fehlende []LauschbefundDTO `json:"fehlende"`
}

// LauschbefundDTO ist eine gewählte Adresse ohne Lauscher.
type LauschbefundDTO struct {
	Adresse string `json:"adresse"`
	Grund   string `json:"grund"`
}

func lauschbefundeAus(befunde []domain.Lauschbefund) []LauschbefundDTO {
	dtos := make([]LauschbefundDTO, 0, len(befunde))
	for _, befund := range befunde {
		dtos = append(dtos, LauschbefundDTO{Adresse: befund.Adresse, Grund: befund.Grund})
	}

	return dtos
}

// ersteNachAussen gibt die angesagte Adresse zur ersten Bindung, die den
// Rechner verlässt — dort fragt die Erreichbarkeit.
//
// **Gepaart über die gelungenen Bindungen** (Befund 3 des Zweitblicks über
// Phase G): angesagt sind nur die, auf denen ein Lauscher entstand.
func ersteNachAussen(bindungen []domain.Bindung, lauf dienstlauf) (string, bool) {
	for stelle, bindung := range domain.Gelungene(bindungen, lauf.fehlende) {
		if !bindung.VerlaesstDenRechner() {
			continue
		}

		if stelle < len(lauf.adressen) {
			return lauf.adressen[stelle], true
		}

		return lauf.adresse, true
	}

	return "", false
}

// dienstStarten startet den Dienst und wartet auf seine Ansage.
//
// **Die Rolle wird gefragt, bevor der Rechner erreichbar wird** — dieselbe
// Wirkungsklasse wie beim Horchposten (N7, #1568).
func (a *App) dienstStarten(
	ctx context.Context, art domain.Dienstart, bindung domain.Bindung,
) (Dienstzustand, error) {
	return a.dienstStartenAuf(ctx, art, []domain.Bindung{bindung})
}

// dienstStartenAuf startet den Dienst auf jeder Bindung der Wahl — G08
// (#2272). Was das Programm ansagt, jede Adresse und jede fehlende, steht
// danach im Stand.
func (a *App) dienstStartenAuf(
	ctx context.Context, art domain.Dienstart, bindungen []domain.Bindung,
) (Dienstzustand, error) {
	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(laufDes(art))

	if fehler := a.darf(ctx, domain.WirkungExternEmpfangen); fehler != nil {
		return Dienstzustand{}, fehler
	}

	for _, bindung := range bindungen {
		if fehler := domain.PruefeBindung(bindung); fehler != nil {
			return Dienstzustand{}, fehler
		}
	}

	posten := a.dienste.posten(art)

	// **Der Platz wird belegt, bevor er gebraucht wird.** Wer nur läse und
	// das Schloss wieder löste, prüfte einen Stand von vorhin — der Grund
	// steht bei [dienstposten.startet].
	posten.schloss.Lock()
	laeuftSchon := posten.beenden != nil || posten.startet
	if !laeuftSchon {
		posten.startet = true
	}
	posten.schloss.Unlock()

	if laeuftSchon {
		return Dienstzustand{}, fmt.Errorf(
			"%w: dieser Dienst läuft schon — erst beenden, dann neu anfangen",
			domain.ErrConflict)
	}

	// **Und er wird auf jedem Weg wieder frei.** Ein Fehlschlag, der die
	// Belegung stehen ließe, machte aus einem Befund eine Sperre bis zum
	// nächsten Neustart.
	defer func() {
		posten.schloss.Lock()
		posten.startet = false
		posten.schloss.Unlock()
	}()

	lauf, fehler := a.dienststarten()(ctx, art, bindungen)
	if fehler != nil {
		return Dienstzustand{}, fehler
	}

	// **Ein Befund hier hält nichts auf** — derselbe Satz wie beim
	// Horchposten: der Port steht, nur die Liste zum Weitergeben bleibt leer.
	var erreichbar leitung.Erreichbarkeit

	if draussen, gibt := ersteNachAussen(bindungen, lauf); gibt {
		gefunden, fehler := a.erreichbarkeit()(ctx, ohneSchema(draussen))
		if fehler != nil {
			meldeDienstbefund(fehler)
		} else {
			erreichbar = gefunden
		}
	}

	posten.schloss.Lock()
	posten.adresse = lauf.adresse
	posten.adressen, posten.fehlende = lauf.adressen, lauf.fehlende
	posten.marke = lauf.marke
	posten.fingerabdruck = lauf.fingerabdruck
	posten.aussteller, posten.gueltigBis = lauf.aussteller, lauf.gueltigBis
	posten.erreichbar = erreichbar
	posten.beenden = lauf.beenden
	posten.fertig = lauf.fertig
	posten.schloss.Unlock()

	a.nachDemStart(ctx, laufDes(art), marke, func(context.Context) { a.dienstBeenden(art) })

	return a.dienststand(art), nil
}

// ohneSchema nimmt `https://` vorn weg.
//
// **Weil [leitung.ErreichbarUnter] eine Horchadresse erwartet** und kein URL.
// Der Dienst sagt sich mit Schema an, damit ein Mensch die Zeile einfügen
// kann; die Rechnung darunter braucht nur Wirt und Port.
func ohneSchema(adresse string) string {
	for _, schema := range []string{"https://", "http://"} {
		if ohne, ja := strings.CutPrefix(adresse, schema); ja {
			return ohne
		}
	}

	return adresse
}

// dienstBeenden hält den Posten an und wartet, bis er steht.
//
// **Und es wartet wirklich** — sonst sagte der Stand „nicht mehr", während
// der Lauscher den Port noch hält, und das nächste Starten scheiterte an
// einer Adresse, die schon belegt ist.
func (a *App) dienstBeenden(art domain.Dienstart) {
	posten := a.dienste.posten(art)

	posten.schloss.Lock()
	beenden, fertig := posten.beenden, posten.fertig
	posten.beenden, posten.fertig = nil, nil
	posten.adresse, posten.marke = "", ""
	posten.adressen, posten.fehlende = nil, nil
	posten.fingerabdruck = ""
	posten.aussteller, posten.gueltigBis = "", time.Time{}
	posten.erreichbar = leitung.Erreichbarkeit{}
	posten.schloss.Unlock()

	if beenden == nil {
		return
	}

	beenden()
	<-fertig
}

// dienststand sagt, was gerade läuft.
func (a *App) dienststand(art domain.Dienstart) Dienstzustand {
	posten := a.dienste.posten(art)

	posten.schloss.Lock()
	defer posten.schloss.Unlock()

	stand := Dienstzustand{
		Art:      string(art),
		Laeuft:   posten.beenden != nil,
		Adresse:  posten.adresse,
		Marke:    posten.marke,
		TLS:      strings.HasPrefix(posten.adresse, "https://"),
		Namen:    posten.erreichbar.Namen,
		Adressen: posten.erreichbar.Adressen,

		Fingerabdruck: posten.fingerabdruck,

		ZeugnisAussteller: posten.aussteller,

		Lauscher: append([]string{}, posten.adressen...),
		Fehlende: lauschbefundeAus(posten.fehlende),
	}

	if !posten.gueltigBis.IsZero() {
		stand.ZeugnisBis = posten.gueltigBis.UTC().Format(time.RFC3339)
	}

	// **Eine leere Liste und kein `nil`** — sonst steht im Fenster
	// `namen: null`, und jede Schleife darüber fällt um.
	if stand.Namen == nil {
		stand.Namen = []string{}
	}

	if stand.Adressen == nil {
		stand.Adressen = []string{}
	}

	return stand
}

// diensteBeenden hält beide Posten an.
//
// **Für den Profilwechsel und fürs Schließen des Fensters.** Ein Lauscher
// unter der Identität eines Profils, das nicht mehr gezeigt wird, bediente
// Anfragen für einen Bestandsausschnitt, den niemand mehr sieht.
func (a *App) diensteBeenden() {
	for _, art := range domain.Dienstarten() {
		a.dienstBeenden(art)
	}
}

// meldeDienstbefund schreibt einen Befund ins Protokoll.
//
// **Ins Protokoll und nicht ans Fenster**: was hier auffällt, geschieht
// **nach** dem Ruf, der den Posten gestartet hat — es gibt keine Antwort
// mehr, an die es sich hängen ließe.
func meldeDienstbefund(fehler error) {
	slog.Default().Error("ein Dienstposten meldet einen Befund", "grund", fehler)
}
