// Der Wächter über die dritte Zusage aus
// #54: **die
// Anwendung lädt von sich aus nichts herunter.**
//
// # Warum das strukturell geprüft wird und nicht am Verhalten
//
// „Sie lädt nichts" ist eine Aussage über **alle** Wege, nicht über die, die
// ein Test geht. Ein Verhaltenstest zeigte, dass ein bestimmter Aufruf nicht
// lädt; er sagt nichts über den nächsten, den jemand morgen schreibt. Der
// Importgraph sagt es über jeden: wer das Netz nicht erreichen **kann**, lädt
// nicht.
//
// # Wie diese Zusage heute lautet
//
// **Sie ist am 2026-08-23 enger geworden, nicht schwächer.** Sie hieß „die
// Anwendung lädt von sich aus nichts herunter" und deckte damit zwei Dinge
// zugleich ab: dass sie nichts holt, und dass sie nichts anbietet. Seit
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
// bietet sie etwas an — auf `127.0.0.1`, auf Aufruf, mit Marke.
//
// Was bleibt, ist die Zusage, die trägt: **von sich aus** geschieht nichts.
// Kein Abruf ohne Knopfdruck, kein Dienst ohne Start. Und die Aufzählung ist
// jetzt zwei Zeilen lang statt einer — das ist der sichtbare Preis, und er
// steht hier, damit ihn jeder sieht, der die Liste liest.
//
// # Was vorher hier stand
//
// **Phase M hat sie zur Frage gestellt** (#591):
// „MCP und REST vollständig … Bindung ausschließlich an `127.0.0.1`". Ein
// REST-Dienst auf der Rückschleife **ist** ein Netzwerkdienst, und dieser
// Wächter hätte ihm im Weg gestanden.
//
// **[0039](../../docs/entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md)
// hat am 2026-08-22 dagegen entschieden** — kein benannter Abnehmer, und die
// Lücke lag im Umfang der MCP-Werkzeuge, nicht im Protokoll.
//
// Daneben stand der Satz: **„Wer ihn künftig lockern will, schreibt 0039
// fort."** Genau das ist geschehen, und in dieser Reihenfolge: erst
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md),
// dann der Dienst, dann der rote Lauf dieses Wächters, dann der Eintrag
// unten. Ein vorab gelockerter Wächter wäre das stille Lockern gewesen, das
// 0039 ausgeschlossen hat.
//
// **Er fängt beide Richtungen.** `net` und `net/http` bringt mit, wer eine
// Verbindung aufmacht **und** wer auf eine wartet; ein Dienst kommt hier
// genauso wenig durch wie ein Abruf.
//
// **Und das stimmte bis zum 2026-08-23 nicht.** Der Satz stand hier, die Liste
// darunter kannte `net` aber nicht — `net.Listen` braucht nichts weiter, und
// ein Lauscher wäre unbemerkt durchgegangen. Aufgefallen ist es, als
// `internal/rest` einen aufmachte und dieser Wächter **grün** blieb
// (#997).
// Eine Zusage, die zwei Richtungen behauptet und eine hält, macht die andere
// unsichtbar (§10).
//
// # Was er nicht prüft
//
//   - **Dass `internal/bezug` erst auf Anweisung lädt.** Es *darf* laden — das
//     ist sein Zweck (#56). Dass der Weg dorthin mit einer ausdrücklichen
//     Zustimmung beginnt, hält `service.Paketbezieher` und die Paketseite; hier
//     steht nur, dass es keinen **zweiten** Weg gibt.
//
//   - **Kein Unterprozess.** Wer `os/exec` aufriefe und `curl` startete, käme
//     hier durch.
//
//     **Hier stand „`os/exec` kommt im Projekt nicht vor"** — und das war seit
//     dem 2026-08-25 falsch: `internal/app/protokoll_darwin.go` und
//     `startmeldung_darwin.go` importieren es beide. Angestrichen vom ersten
//     Zweitblick über die Freigabe 0.3.0
//     (#1170).
//
//     **Was heute wirklich gilt**: die beiden starten feste Programme mit
//     festen Argumenten — `/usr/bin/logger` und `/usr/bin/osascript` — und
//     eröffnen keinen Netzweg. Das hält [TestKeinUnterprozessNimmtEinArgumentAusDerFerne]
//     nach: ein Aufruf mit einem Programmnamen aus einer Veränderlichen macht
//     die Pipeline rot. Was er **nicht** hält, ist ein fest verdrahtetes
//     `curl` — dafür gehört die Frage strukturell gestellt (#227).
//
//   - **Keinen Weg unter den Bibliotheken durch.** `syscall.Socket` und
//     `syscall.Connect` kämen hier durch; die Liste unten kennt sie nicht.
//     `syscall` aufzunehmen ginge nicht — das Paket wird für **Signale**
//     gebraucht (`cmd/mse/main.go`), und ein Wächter, der jedes Programm mit
//     ⌃C-Behandlung anhält, wäre unbrauchbar. Wer wirklich so ins Netz will,
//     tut es absichtlich, und dagegen ist dieser Wächter nicht gebaut.
//     Angestrichen vom Nachlauf zum Zweitblick über Phase U.
//
//   - **Wohin.** Er sieht, wer ins Netz darf, nicht, welche Adresse. Das
//     hält seit SEC-P4 (#2000) die Zielprüfung in `internal/netz/ziel.go`,
//     und dass jeder Wähler und jeder Klient dort sie trägt,
//     `netzziel_test.go`.
//
//   - **Keine Testdatei.** `service/paketbezug_test.go` startet einen eigenen
//     `httptest`-Server und muss es: der Bezugsweg lässt sich sonst nicht
//     prüfen, ohne fremde Rechner zu befragen. Ausgeliefert wird davon nichts.
//     Die Zusage handelt von der **Anwendung**, und die besteht aus dem, was
//     gebaut wird.
package schichten

import (
	"slices"
	"strings"
	"testing"
)

// netzpakete sind die Standardbibliotheken, die eine Verbindung aufmachen.
//
// **`net/url` steht nicht dabei**, und das ist kein Versehen: es zerlegt
// Adressen und öffnet nichts. Wer es aufnähme, machte den Wächter unbrauchbar,
// sobald jemand eine Adresse prüfen will.
var netzpakete = []string{"net/http", "net/rpc", "net/smtp"}

// netzgenau sind die Pakete, die **nur genau so** zählen.
//
// **`net` steht hier und nicht oben**, und das ist der ganze Grund für die
// Trennung: die Liste darüber trifft auch die Unterpakete, und `net/url`
// zerlegt Adressen, ohne etwas zu öffnen. Wer `net` dort einträgt, nimmt
// `net/url` mit und macht den Wächter unbrauchbar, sobald jemand eine Adresse
// prüfen will.
//
// **Warum `net` überhaupt dazugehört**: `net.Listen` braucht nichts weiter.
// Bis zum 2026-08-23 stand hier nichts, und ein Lauscher in `internal/rest`
// ging durch, während der Kommentar am Kopf zusagte, der Wächter fange „wer
// auf eine wartet" (#997).
var netzgenau = []string{"net"}

// netzfrei sind die Pakete, die das Netz nicht erreichen dürfen.
//
// **Eine Ausnahmeliste und keine Aufzählung der Erlaubten.** Ein neues Paket
// unter `internal/` ist damit von sich aus netzfrei — und wer es nicht sein
// will, trägt sich hier ein und sieht dabei, was er tut.
var mitNetz = map[string]bool{
	// `bezug` holt Vorlagenpakete — auf Knopfdruck, an einem festen Stand und
	// mit Prüfsumme (#56). Das ist sein Zweck.
	"bezug": true,

	// `rest` **wartet** auf Verbindungen, und das ist die neue Art von Eintrag
	// in dieser Liste.
	//
	// **Er steht hier, weil 0046 ihn erlaubt hat** — und 0046 steht, weil 0039
	// den Fall vorgesehen hat. Was die Lockerung **nicht** aufgibt, hält
	// `internal/rest/lauscher_test.go`: gebunden wird an `127.0.0.1` und an
	// nichts sonst, und jede andere Adresse wird abgewiesen. Dieser Wächter
	// sagt nur, **dass** das Paket das Netz erreichen darf; **wohin**, sagt
	// jener.
	//
	// **Zwei Wächter für eine Zusage**, und das ist Absicht: der Importgraph
	// kennt keine Adressen, und ein Test, der eine Adresse prüft, kennt nicht
	// jeden Weg.
	"rest": true,

	// **Und sein Composition Root**, der den Lauscher an einen `http.Server`
	// gibt. Dass er `cmd/…` heißt, ändert nichts: dieser Wächter liest `cmd/`
	// mit, seit es einen zweiten Prozess gibt.
	//
	// **`cmd/mse` steht ausdrücklich nicht hier.** Der Dienst hätte auch ein
	// Unterbefehl sein können; dann dürfte die **ganze** Kommandozeile ins
	// Netz — auch `feld` und `karten`, die damit nichts zu tun haben. Ein
	// eigenes Programm hält die Aufzählung so klein, wie die Sache es zulässt.
	"cmd/mse-rest": true,

	// **Und `mcp`, seit es einen zweiten Transport hat** — Werkstück N6 von
	// Phase 5.4
	// (#1572).
	//
	// **Es wartet, es holt nicht.** Der Satz im Kopf dieses Wächters lautet
	// „die Anwendung soll von sich aus nichts laden"; dieser Transport lädt
	// nichts — er beantwortet, was hereinkommt, und das ist dieselbe Art von
	// Eintrag wie `rest` darüber.
	//
	// **Was der Wächter davon nicht sieht**: dass daraus kein Klient wird.
	// `net/http` kann beides, und ein `http.Get` in diesem Paket fiele ihm
	// nicht auf. Dagegen hält der Riegel an `netz.Klient` — jeder Weg nach
	// draußen geht über ihn, und dieses Paket kennt ihn nicht.
	"mcp": true,

	// **Und `cmd/mse-mcp`**, aus demselben Grund wie `cmd/mse-rest`: der
	// Composition Root gibt den Lauscher an einen `http.Server`.
	"cmd/mse-mcp": true,

	// `vorschauweg` **erreicht das Netz nicht** — und steht trotzdem hier.
	//
	// Es hat keinen Klienten, keinen Lauscher und keine Adresse. Es formt
	// gerasterte Kartenbilder zu einer HTTP-Antwort für die **Assetleitung**,
	// die Wails für das eigene Fenster ohnehin betreibt — dieselbe, die
	// `index.html` ausliefert. Der Eintrag steht hier, weil dieser Wächter
	// **strukturell** prüft, und das mit Absicht (siehe oben): wer `net/http`
	// einführt, **kann** ins Netz, und ein `http.Handler` ist ohne `net/http`
	// nicht zu schreiben.
	//
	// **Die Alternative war `internal/app`, und sie wäre teurer gewesen**: die
	// ganze Bindungsfassade hätte damit den Weg nach draußen gehabt, für einen
	// Typnamen. Der Weg liegt deshalb in einem eigenen Blatt, und der
	// Rückgabetyp ist `*vorschauweg.Weg` statt `http.Handler` — so führen
	// weder `app` noch `main.go` ein `net/http` ein
	// (#1270).
	//
	// **Dass die Begründung stimmt, ist nachgeprüft** und nicht behauptet:
	// `internal/vorschauweg/keineverbindung_test.go` macht die Pipeline rot,
	// sobald dort ein Klient, ein Lauscher oder eine Adresse auftaucht (§10).
	//
	// **Seit SEC-17 steht dort auch die Schutzweiche** (#2164): sie setzt die
	// Schutzrichtlinie an jede Antwort derselben Assetleitung und nimmt die
	// Berichte an — aus demselben Grund hier und nicht in `internal/fenster`.
	"vorschauweg": true,

	// **`anmeldung` erreicht zwei fremde Endpunkte und horcht auf einem
	// eigenen** — Werkstück 4 von Phase 4.12
	// (#1351).
	//
	// # Warum es ein eigener Eintrag ist und nicht über `netz` läuft
	//
	// **Weil der Horcher kein Klient ist.** `internal/netz` gibt einen
	// Klienten mit Frist und Drossel; hier steht daneben ein `net.Listener`
	// auf `127.0.0.1:0` und ein `http.Server` darauf — das ist die Gegenseite
	// und keine Anfrage.
	//
	// **Und Google verlangt genau das** (nachgelesen am 2026-09-05): „start an
	// HTTP listener on a random available port"; eigene Schemata sind
	// abgeschafft, „no longer supported due to the risk of app
	// impersonation".
	//
	// **Was der Eintrag kostet, gehört gesagt**: dieses Paket kann von sich
	// aus ins Netz. Dass es das nur auf Anweisung tut, hängt an seinen
	// Aufrufern — der Klient wird **eingesetzt** (§16), und ohne Kennung
	// beginnt gar kein Lauf.
	"anmeldung": true,

	// `netz` **ist** der Transport, und das ist der Unterschied zu den beiden
	// darüber.
	//
	// Die kennen das Netz, weil sie eine Aufgabe haben, die es braucht. Dieses
	// Paket hat keine Aufgabe außer dem Netz: es ist die **eine** Stelle, an
	// der Fristen, Kopfzeilen, Drossel und der Umgang mit einem 429 stehen
	// (#1002).
	//
	// **Ein Eintrag, der die Liste kürzer machen soll, nicht länger.** Jeder
	// weitere fremde Dienst hängt an ihm und kommt **nicht** in diese
	// Aufzählung — sonst stünde hier bald jeder Klient, und die Grenze wäre
	// keine.
	//
	// **Was er nicht aufgibt**: `netz` fragt nichts von selbst. Es hat keinen
	// Zeitgeber, keinen Faden und keine Adresse; wer es benutzt, sagt wann und
	// wohin. Die Zusage „von sich aus geschieht nichts" hängt damit
	// unverändert an seinen Aufrufern.
	//
	// **Und seit SEC-P4 prüft es dieses Wohin** (#2000): die Grundsperre für
	// jeden Klienten, nur öffentlich gerichtete Ziele für eine Adresse, die
	// ein Rufer nennt — nach der Auflösung und bei jeder Weiterleitung. Dass
	// kein Weg des Transports daran vorbeigeht, hält `netzziel_test.go`.
	"netz": true,

	// **`leitung` kennt `net` und lädt trotzdem nichts** — Werkstück N2 von
	// Phase 5.4
	// (#1455).
	//
	// Es trägt Nachrichten über einen **fertigen** `net.Conn`: es wählt
	// nicht, es horcht nicht, es kennt keine Adresse und keinen Namen. Wer
	// den Strom aufbaut, ist der Composition Root — und der entscheidet,
	// ob überhaupt etwas nach draußen geht.
	//
	// **Die Zusage „von sich aus geschieht nichts" bleibt damit unberührt**,
	// genau wie bei `netz` darüber: `net` steht hier für die Typen
	// `net.Conn` und `net.Pipe`, nicht für einen Verbindungsaufbau.
	"leitung": true,

	// **`fernleitung` geht ins Internet, und das ist sein Zweck** —
	// Werkstück N3 von Phase 5.4
	// (#1456),
	// entschieden in [0114](../../docs/entscheidungen/0114-tailcat-gepinnt.md).
	//
	// **Der weitreichendste Eintrag dieser Liste, und er gehört benannt.**
	// Über tailcat spricht dieses Programm mit einem **DERP-Relais von
	// Tailscale** — einem fremden Dienst, den wir nicht betreiben. Was
	// dorthin geht, ist verschlüsselter WireGuard-Verkehr; was dort anfällt,
	// sind Verbindungsdaten.
	//
	// **Und trotzdem gilt „von sich aus geschieht nichts".** `Horche` und
	// `Waehle` entstehen nur auf Aufruf; wer keines von beiden ruft, baut
	// keine Verbindung und holt keine DERP-Karte. Dieselbe Linie wie beim
	// Horchposten aus N2 — und sie ist hier wichtiger, weil das Gegenüber
	// nicht mehr im selben Raum steht.
	"fernleitung": true,

	// **`entdeckung` ruft ins Netz, und das ist sein Zweck** — Werkstück N2
	// von Phase 5.4
	// (#1455).
	//
	// # Warum das die Zusage aus #54 nicht aufhebt
	//
	// **Weil von selbst nichts geschieht.** Die Zusage lautet „**von sich
	// aus** geschieht nichts" — kein Abruf ohne Knopfdruck, kein Dienst ohne
	// Start. mDNS ist beides: `Melde` kündigt an, `Suche` fragt, und keines
	// von beiden läuft, ohne dass ein Mensch es ausgelöst hat.
	//
	// **Und was hinausgeht, ist benannt**: die Peerkennung dieses Profils und
	// sein selbstgewählter Name. Der Rechnername geht ausdrücklich **nicht**
	// hinaus — die Bibliothek nähme ihn ohne Zutun, und das Paket setzt ihn
	// deshalb selbst.
	"entdeckung": true,
}

// mitNetzDatei nennt **einzelne Dateien**, die das Netz kennen dürfen, wo ihr
// Paket es nicht darf.
//
// **Die Abnahmeleitung horcht auf einem Unix-Socket** (#1727): eine Datei mit
// den Rechten `0600`, keine Adresse und kein Port, gebaut nur mit dem
// Schalter `abnahme`, den `scripts/releaseschalter.sh` aus jedem Freigabebau
// fernhält. Ein Eintrag für ganz `fenster` ließe jede andere Datei mit durch.
//
// **Das Rufbuch, seine Einordnung und seine Störungen** (B01, B03 — #2220,
// #2222) sind Hüllen um
// den Asset-Server: sie nehmen einen `http.Handler` und geben einen zurück.
// Ohne Schalter gebaut sind sie, damit ihre Tests in der Teststufe laufen —
// wie `abnahme.go`; benutzt werden sie nur von `abnahme_darwin.go`. Dass sie
// aus `net/http` nichts als die Typen nehmen, hält `netzhuellen_test.go`.
//
// **Und der Riegel um dieselben Hüllen** (F02, #2341): er nimmt den Handler
// der Weiche und fängt eine Panik darin, weil Wails jede Anfrage der
// Webansicht in einer eigenen Goroutine ohne `recover` beantwortet.
var mitNetzDatei = map[string]bool{
	"fenster/abnahme_darwin.go": true,
	"fenster/rufbuch.go":        true,
	"fenster/rufausgang.go":     true,
	"fenster/rufstoerung.go":    true,
	"fenster/riegelweiche.go":   true,
}

func TestNurBezugKenntDasNetz(t *testing.T) {
	for paket, einfuhren := range importeLesen(t) {
		if mitNetz[paket] {
			continue
		}

		for _, einfuhr := range einfuhren {
			if strings.HasSuffix(einfuhr.Datei, "_test.go") || mitNetzDatei[paket+"/"+einfuhr.Datei] {
				continue
			}

			if istNetz(einfuhr.Pfad) {
				t.Errorf("internal/%s (%s) importiert %s — die Anwendung soll von sich aus "+
					"nichts laden (#54); der Bezug steht in internal/bezug",
					paket, einfuhr.Datei, einfuhr.Pfad)
			}
		}
	}
}

// TestDerNetzwaechterWirdRot ist die Gegenprobe — als Testfall (§10).
//
// Ohne sie bliebe offen, ob die Prüfung überhaupt etwas erkennt: eine, die
// `net/http` durchgehen ließe, sähe von außen genauso aus.
func TestDerNetzwaechterWirdRot(t *testing.T) {
	faelle := map[string]bool{
		"net/http":          true,
		"net/http/httptest": true,
		"net/rpc":           true,

		// **`net` selbst**, und das war eine Lücke bis zum 2026-08-23. Der
		// Kommentar oben sagt seit jeher, der Wächter fange „wer auf eine
		// wartet" — ein `net.Listen` braucht aber **nur** `net`, und das stand
		// nicht in der Liste. Aufgefallen ist es, als `internal/rest` einen
		// Lauscher aufmachte und der Wächter grün blieb (#997).
		"net": true,

		// **Und `net/url` bleibt draußen.** Es zerlegt Adressen und öffnet
		// nichts. Der Fall steht hier, weil die Aufnahme von `net` ihn
		// beinahe mitgenommen hätte: `HasPrefix("net/url", "net/")` trifft.
		"net/url": false,

		"os":                            false,
		"github.com/…/internal/bezug":   false,
		"internal/nethttp/nichtdasnetz": false,
	}

	for pfad, erwartet := range faelle {
		t.Run(pfad, func(t *testing.T) {
			if istNetz(pfad) != erwartet {
				t.Errorf("istNetz(%q) = %v, erwartet %v", pfad, !erwartet, erwartet)
			}
		})
	}
}

// istNetz sagt, ob ein Importpfad eine Netzbibliothek benennt.
//
// **Auch die Unterpakete.** `net/http/httptest` bringt einen Server mit; wer
// nur den genauen Pfad prüfte, ließe ihn durch.
func istNetz(pfad string) bool {
	if slices.Contains(netzgenau, pfad) {
		return true
	}

	for _, netz := range netzpakete {
		if pfad == netz || strings.HasPrefix(pfad, netz+"/") {
			return true
		}
	}

	return false
}
