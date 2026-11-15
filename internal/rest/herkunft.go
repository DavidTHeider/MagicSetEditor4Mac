package rest

import (
	"net"
	"net/http"
	"slices"
)

// eigeneRechner sind die Namen, unter denen dieser Dienst sich ansprechen lässt.
//
// **`localhost` steht dabei, obwohl gebunden wird an `127.0.0.1`.** Wer
// `http://localhost:7777/sets` aufruft, schickt `Host: localhost:7777`, und
// das ist der gewöhnliche Weg — ihn abzuweisen hieße, jedem Aufrufer eine
// Schreibweise vorzuschreiben, die ihm nichts nützt.
//
// **`::1` steht nicht dabei** — und hier stand, es werde nicht gebunden. Seit
// N6 durfte es das, und ein Dienst auf `::1` wies jede eigene Anfrage ab.
// **Seit G08 (#2272) bringt jedes Programm die Adressen mit, auf denen es
// lauscht** (`app.Rechnerteile` über [Dienst.MitEigenenNamen]); hier bleiben
// die zwei, die ohne Lauscher gelten.
var eigeneRechner = map[string]bool{"127.0.0.1": true, "localhost": true}

// HerkunftTaugt sagt, ob eine Anfrage von einem Aufrufer kommt, der gemeint ist.
//
// # Was sie prüft
//
//   - **Kein `Origin`, kein `Referer`.** Beide setzt ein Browser und kein
//     Skript. Eine beliebige Seite, die `fetch` auf diesen Port richtet, wird
//     damit abgewiesen, bevor die Marke überhaupt geprüft wird.
//   - **`Host` ist einer der eigenen Namen.** Das ist die Abwehr gegen
//     DNS-Rebinding: dort hält der Browser die Seite für gleichherkünftig und
//     schickt **kein** `Origin` — wohl aber den Namen des Angreifers als
//     `Host`.
//
// # Was sie nicht prüft
//
// **Welcher lokale Prozess fragt.** Das kann sie nicht: eine TCP-Verbindung
// von der Rückschleife trägt keinen Absender, dem zu trauen wäre. Dagegen
// steht die Marke, und nur sie.
//
// **Sie ersetzt die Marke nicht.** Ein Skript setzt weder `Origin` noch
// `Referer`; ein Angreifer mit einem lokalen Prozess kann sich also
// hindurchschreiben. Die beiden Prüfungen decken verschiedene Gegenüber ab und
// stehen deshalb nebeneinander, nicht alternativ.
func HerkunftTaugt(anfrage *http.Request) bool {
	return HerkunftTaugtUnter(anfrage, nil)
}

// HerkunftTaugtUnter ist [HerkunftTaugt] mit zusätzlichen eigenen Namen.
//
// # Warum es das seit N6 gibt
//
// **Weil `Host` sonst jede Bindung jenseits der Schleife abweist** —
// Werkstück N6 von Phase 5.4
// (#1572).
// Die Liste oben kennt `127.0.0.1` und `localhost`, und das war vollständig,
// solange der Dienst nirgendwo sonst hinband. Gemessen am 2026-09-15: an
// `0.0.0.0` gebunden und über die eigene LAN-Adresse gefragt, antwortete er
// `403 herkunft` — die Marke stimmte, das Zeugnis trug, und die Anfrage kam
// trotzdem nicht durch.
//
// # Was die Abwehr dabei **nicht** verliert
//
// **Der Name des Angreifers steht nicht in der Liste.** Beim DNS-Rebinding
// hält der Browser die Seite für gleichherkünftig und schickt kein `Origin`
// — wohl aber **seinen** Namen als `Host`. Solange hier nur steht, wohin
// dieser Dienst wirklich gebunden ist, trifft das keinen von ihnen.
//
// **Und die Liste kommt von außen**: welche Adressen dieser Rechner hat,
// weiß `internal/leitung` und nicht dieses Blatt (§2).
func HerkunftTaugtUnter(anfrage *http.Request, weitere []string) bool {
	if anfrage.Header.Get("Origin") != "" || anfrage.Header.Get("Referer") != "" {
		return false
	}

	rechner, _, fehler := net.SplitHostPort(anfrage.Host)
	if fehler != nil {
		// Ohne Port ist der ganze Wert der Rechnername.
		rechner = anfrage.Host
	}

	if eigeneRechner[rechner] {
		return true
	}

	return slices.Contains(weitere, rechner)
}
