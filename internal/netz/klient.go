package netz

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Fristen sind die drei Zeiten, die eine Anfrage begrenzen.
//
// # Warum drei und nicht eine
//
// **`http.Client.Timeout` deckt den ganzen Vorgang ab, Rumpf eingeschlossen.**
// Eine Bulk-Datei von Scryfall wiegt 74 MB, ein Vorlagenpaket bis zu zwei
// Gigabyte; über eine langsame Leitung ist das legitim und dauert. Eine
// Gesamtfrist, die dafür reicht, fängt keinen hängenden Server mehr — und
// eine, die ihn fängt, bricht den langsamen Bezug ab.
//
// Deshalb: eine Frist für den **Verbindungsaufbau**, eine für den
// **Antwortkopf**, und eine für den **Stillstand** des Rumpfes.
//
// **Die dritte ist die, die man vergisst.** `ResponseHeaderTimeout` misst bis
// zur Statuszeile; wer sie schickt und dann schweigt, war danach unbeobachtet.
// Das war Audit-Befund E-37, und
// #916 hat ihn
// geschlossen.
//
// # Warum sie hier stehen und nicht als Veränderliche im Paket
//
// `internal/bezug` hält sie heute als `var`, damit seine Tests sie
// herunterdrehen können. Für **zwei** Verbraucher trägt das nicht: sie
// bräuchten verschiedene Werte, und eine gemeinsame Veränderliche wäre eine
// Absprache zwischen zwei Paketen, die einander nicht kennen.
type Fristen struct {
	// Verbindung ist die Zeit für Verbindungsaufbau und TLS.
	Verbindung time.Duration
	// Kopf ist die Zeit, die eine Gegenstelle für ihren Antwortkopf hat.
	Kopf time.Duration
	// Stillstand ist die Zeit, die der Rumpf ohne Fortschritt haben darf.
	Stillstand time.Duration
}

// UeblicheFristen sind die Werte, mit denen dieses Projekt arbeitet.
//
// **Sie stehen an einer Stelle**, damit nicht jeder Dienst sie neu erfindet —
// und sie sind trotzdem einstellbar, weil ein Test sie herunterdrehen muss.
func UeblicheFristen() Fristen {
	return Fristen{
		Verbindung: 15 * time.Second,
		Kopf:       30 * time.Second,
		Stillstand: 60 * time.Second,
	}
}

// LangeFristen sind die Werte für einen Abruf, von dem der Anbieter selbst
// sagt, dass er lange dauert.
//
// # Der Fund, für den es sie gibt
//
// **CardTrader empfiehlt für den vollen Bestandsexport 120 bis 180 Sekunden
// Lesefrist** (nachgelesen am 2026-09-19, [0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
// Das Paket schrieb es auf, die Naht nahm die üblichen dreißig — ein Export,
// dessen Kopf nach dreißig Sekunden noch nicht da ist, wurde abgebrochen
// (Befund 6 des Zweitblicks über Phase 5.6).
//
// **Die Verbindung bleibt kurz**: dass ein Rechner nicht antwortet, merkt
// man in fünfzehn Sekunden — lange dauert das **Rechnen** drüben, nicht das
// Verbinden.
func LangeFristen() Fristen {
	return Fristen{
		Verbindung: 15 * time.Second,
		Kopf:       180 * time.Second,
		Stillstand: 180 * time.Second,
	}
}

// Klient ist, womit ein Dienst ins Netz geht.
//
// # Warum ein eigener Typ und nicht `*http.Client`
//
// Damit **kein** Dienstpaket `net/http` einbinden muss. Der Wächter in
// `internal/schichten` führt, wer das Netz erreichen darf, und diese
// Aufzählung soll **nicht** mit jedem neuen Dienst wachsen — sonst stünde dort
// bald jeder Klient, und die Grenze wäre keine
// (#1005).
//
// **Aufgefallen ist das an `internal/scryfall`**: es nahm einen
// `*http.Client` entgegen, machte damit selbst nichts — und der Wächter
// wurde trotzdem rot. Er hatte recht: der Import ist da, und ein Wächter über
// den Importgraphen kann nicht sehen, dass niemand ihn benutzt.
type Klient struct {
	unten *http.Client

	// fristen nennt die Frist, die griff, wenn eine Anfrage an ihr endet (D02).
	fristen Fristen

	// proxy ist der **wirksame** Stand, mit dem dieser Klient gebaut wurde.
	//
	// **Er wird mitgeführt, damit ein Fehlschlag ihn nennen kann** (§8):
	// „connection refused" ohne den Zusatz, dass ein Proxy dazwischenstand,
	// schickt jemanden in die falsche Richtung — und #1533 führt genau das
	// als offenen Punkt.
	proxy Proxyquelle

	// erlaubnis wird vor **jeder** Anfrage gefragt — Phase 4.8
	// (#1321).
	//
	// # Warum hier und nicht dort, wo der Klient entsteht
	//
	// **Weil ein Schalter sich ändert.** Der erste Anlauf prüfte beim
	// Verdrahten: wer einen Dienst danach abschaltete, sah bis zum Neustart
	// keine Wirkung — und die Zusage „aus heißt null Anfragen" galt für den
	// Stand von vorhin.
	//
	// **Und weil hier jede Anfrage vorbeikommt.** Ein Klient hat viele
	// Wege; eine Prüfung je Weg wäre so oft dieselbe Zusage und einmal die
	// vergessene.
	//
	// `nil` heißt: es fragt niemand — dieselbe Bedeutung wie eine fehlende
	// Drossel.
	erlaubnis func(context.Context) error

	// ziele ist der Bereich, in den dieser Klient wählen darf (#2000).
	//
	// **Er steht auch hier und nicht nur im Wähler**, weil die Prüfung vor
	// der Anfrage ihn braucht ([Klient.darfFragen]) — die einzige, die auch
	// hinter einem Proxy greift.
	ziele Zielbereich
}

// darfFragen fragt die Erlaubnis, wenn eine gesetzt ist — und prüft das Ziel.
//
// **Das Ziel gehört seit #2000 dazu**: „darf ich fragen" heißt auch „darf ich
// **dorthin** fragen". Weil der Wächter in `erlaubnisvollstaendig_test.go`
// nachzählt, dass jeder `Do` hier vorbeikommt, hält er damit auch die
// Zielprüfung vor jeder Anfrage. Eine Adresse, die sich nicht zerlegen
// lässt, geht durch — sie scheitert beim Bau der Anfrage mit ihrer eigenen
// Meldung.
//
// **Sie steht hier und nicht in jedem Weg nach draußen**, damit „vor jeder
// Anfrage" eine Stelle hat statt einer Gewohnheit. Bis zum 2026-09-11 las
// [Strom] sie nicht — die Zusage am Feld galt für [Hole] und [SendeJSON], und
// niemand sah es, weil der Wächter drüben prüft, ob eine Erlaubnis **gesetzt**
// wird, nicht, ob sie **gelesen** wird. Seitdem zählt ein zweiter Wächter
// nach, dass jeder `Do` hier vorbeikommt.
func (k *Klient) darfFragen(ctx context.Context, adresse string) error {
	if k.erlaubnis != nil {
		if fehler := k.erlaubnis(ctx); fehler != nil {
			return fehler
		}
	}

	ziel, fehler := url.Parse(adresse)
	if fehler != nil {
		return nil //nolint:nilerr // der Bau der Anfrage meldet dieselbe Adresse mit seinem eigenen Satz
	}

	return pruefeZiel(ziel, k.ziele)
}

// MitErlaubnis gibt einen Klienten, der vor jeder Anfrage fragt.
//
// **Er gibt einen neuen zurück und ändert den alten nicht** — ein Klient, der
// unterwegs seine Erlaubnis wechselte, wäre ein Zustand, den zwei Aufrufer
// teilten.
//
// **Als Kopie des Ganzen und nicht Feld für Feld** (D02, #2233): die Liste
// der Felder vergaß die Fristen, und eine Meldung über die Kopffrist nannte
// „{sekunden}" statt einer Zahl — für jeden Klienten mit Erlaubnis, also für
// Scryfall und die Kartendienste. Ein neues Feld reist so von selbst mit.
func (k *Klient) MitErlaubnis(fragen func(context.Context) error) *Klient {
	kopie := *k
	kopie.erlaubnis = fragen

	return &kopie
}

// NeuerKlient baut einen Klienten mit den drei Fristen.
//
// **Nicht `http.DefaultClient`.** Dessen `Timeout` ist null, und ein Kontext
// aus dem Fenster trägt keine Frist: ein hängender Server hielte den Aufruf
// beliebig lange fest, und abbrechen könnte nur, wer die Anwendung beendet
// (#899).
func NeuerKlient(fristen Fristen) *Klient {
	return NeuerKlientMit(fristen, nil)
}

// NeuerKlientMit baut einen Klienten, dessen Proxy **eingestellt** sein darf.
//
// **Die Einstellung schlägt die Umgebung** — Werkstück N5 von Phase 5.4
// (#1571),
// und das ist die Antwort auf die Frage, die A18 offengelassen hat.
//
// **[NeuerKlient] daneben bleibt** und ruft hierher mit einer leeren Vorgabe:
// es gibt Stellen, die keine Einstellungen lesen können — `internal/bezug`
// kennt den Bestand nicht —, und ein Klient, der dort eine leere Struktur
// zusammenbauen müsste, sagte nichts über seine Absicht.
//
// **Er wählt jedes Ziel außer der Grundsperre** ([ZielUeberall]); wer eine
// Adresse wählt, die ein Rufer genannt hat, baut ihn über [NeuerKlientNach].
func NeuerKlientMit(fristen Fristen, quelle Proxyquelle) *Klient {
	return neuerKlient(fristen, quelle, ZielUeberall)
}

// neuerKlient ist der eine Bauweg — mit Fristen, Proxy und Zielbereich.
func neuerKlient(fristen Fristen, quelle Proxyquelle, ziele Zielbereich) *Klient {
	return &Klient{fristen: fristen, unten: &http.Client{
		CheckRedirect: weiterleitungFuer(ziele),
		Transport: &http.Transport{
			Proxy:                  proxyAus(quelle),
			OnProxyConnectResponse: proxyanmeldung,
			DialContext:            waehlerMit(fristen, ziele, quelle),
			TLSHandshakeTimeout:    fristen.Verbindung,
			ResponseHeaderTimeout:  fristen.Kopf,

			// **HTTP/2 ausdrücklich** — Werkstück A18 von Phase 5.3
			// (#1533).
			//
			// **Go schaltet es sonst ab**, und zwar still: ein Transport mit
			// eigenem `DialContext` bekommt h2 nur mit dieser Zeile. Gemessen
			// am 2026-09-12 gegen einen echten h2-Server — `HTTP/1.1`.
			//
			// **Was es bringt**: Multiplexing über **eine** Verbindung bei
			// vielen kleinen Anfragen — Rulings, Preise, Einbettungen.
			// **Was nicht**: eine einzelne große Bulk-Datei wird davon nicht
			// schneller.
			//
			// **Und der alte Server bleibt erreichbar.** `ForceAttemptHTTP2`
			// erzwingt nichts gegen eine Gegenstelle, die es nicht kann — sie
			// bekommt weiter HTTP/1.1, und ein Lauf hält das fest.
			ForceAttemptHTTP2: true,
		},
	}, proxy: quelle, ziele: ziele}
}

// KlientUeber baut einen Klienten über einem fertigen `*http.Client`.
//
// **Für Tests, und nur für sie.** `httptest.Server` bringt einen eigenen
// Klienten mit, der sein Zertifikat kennt; ohne diesen Weg bräuchte jeder Test
// ein echtes.
func KlientUeber(unten *http.Client) *Klient { return &Klient{unten: unten} }

// waehlerMit baut einen Wähler, dessen Verbindungen Stillstand melden.
//
// **An der Verbindung und nicht am Antwortrumpf** (#916). Der erste Anlauf
// legte eine Frist um jedes `Read` des Rumpfes und startete dafür eine
// Goroutine je Lesen. Das trug — und brachte zwei Fragen mit, die keine
// Antwort brauchen sollten: ob die Goroutine nach einem Abbruch wirklich
// endet, und ob sie noch in einen Puffer schreibt, den der Aufrufer aufgegeben
// hat.
//
// **`SetReadDeadline` gibt es an der Verbindung**, und sie deckt damit den
// ganzen Vorgang — Kopf wie Rumpf —, ohne eine einzige Goroutine.
//
// **Und er prüft, wohin er verbindet** (#2000): nach der Auflösung, bevor
// ein Paket hinausgeht. Die Leitung zum Proxy steht dabei unter der
// Grundsperre allein ([istProxyleitung]).
func waehlerMit(
	fristen Fristen, ziele Zielbereich, quelle Proxyquelle,
) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, art, adresse string) (net.Conn, error) {
		bereich := ziele
		if istProxyleitung(ctx, adresse, quelle) {
			bereich = ZielUeberall
		}

		verbindung, err := (&net.Dialer{
			Timeout: fristen.Verbindung,
			Control: kontrolleFuer(bereich),
		}).DialContext(ctx, art, adresse)
		if err != nil {
			return nil, err //nolint:wrapcheck // der Aufrufer erwartet den Fehler des Wählers
		}

		return &fristverbindung{Conn: verbindung, frist: fristen.Stillstand}, nil
	}
}

// fristverbindung stellt die Stillstandsfrist vor jedem Lesen neu.
//
// **Vor jedem Lesen und nicht einmal**: die Frist gilt dem **Stillstand**,
// nicht der Dauer. Was langsam aber stetig ankommt, läuft durch.
type fristverbindung struct {
	net.Conn

	frist time.Duration
}

func (f *fristverbindung) Read(puffer []byte) (int, error) {
	if err := f.SetReadDeadline(time.Now().Add(f.frist)); err != nil {
		return 0, err //nolint:wrapcheck // der Aufrufer erwartet den Fehler der Verbindung
	}

	return f.Conn.Read(puffer) //nolint:wrapcheck // dito
}
