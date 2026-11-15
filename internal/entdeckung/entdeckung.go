// Package entdeckung findet Gegenüber im selben Netz — über mDNS.
//
// **Ein Blatt** (§2): es kennt `internal/domain` und `internal/peer` und
// sonst nichts aus diesem Projekt. Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum das nicht in `internal/leitung` steht
//
// **Weil es eine andere Frage beantwortet.** Dort geht es darum, **wie**
// Bytes von hier nach dort kommen; hier darum, **wer** überhaupt da ist. Die
// zweite lässt sich ohne die erste beantworten — und der Weg „Adresse von
// Hand" beantwortet sie ohne dieses Paket.
//
// # Was hier bewusst nicht geschieht
//
// **Es meldet sich nichts von selbst.** Das Issue verlangt es: „Nur nach
// bewusster Auswahl Spieler/Kontakte in der Nähe sichtbar werden; keine
// unbemerkte Dauerwerbung aller Profile." Wer [Melde] nicht ruft, kündigt
// nichts an — und wer [Suche] nicht ruft, fragt nichts.
//
// **Und eine Ankündigung ist eine Preisgabe.** Was hinausgeht, ist die
// Peerkennung dieses Profils — also sein öffentlicher Schlüssel — und der
// selbstgewählte Name. Jeder im selben Netz kann beides lesen. Genau deshalb
// ist es ein eigener Schritt und nicht die Folge eines anderen.
package entdeckung

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Dienstart ist der Name, unter dem dieses Programm sich meldet.
//
// **Ein eigener und kein `_http._tcp`**: wer nach diesem Namen fragt, bekommt
// nur Antworten von diesem Programm — und wer nach anderen fragt, bekommt
// keine von ihm.
const Dienstart = "_mse4mac._tcp"

// Bereich ist die mDNS-Domäne. `local` ist die einzige, die es dort gibt.
const Bereich = "local"

// Hoechstfrist ist, wie lange eine Suche höchstens dauert.
//
// **„Discovery bleibt kein unbegrenzter Spinner"** steht in der Abnahme des
// Issues, und diese Zahl ist die Stelle, an der das eingelöst wird. Fünf
// Sekunden sind lang genug für ein Netz, das antwortet, und kurz genug, dass
// niemand glaubt, es sei etwas abgestürzt.
//
// **Sie gilt für die Wanduhr und nicht für einen Teilschritt** — der
// Unterschied steht bei [Suche], und er ist gemessen.
const Hoechstfrist = 5 * time.Second

// Nachlauf ist, was eine Suche über ihre Frist hinaus braucht.
//
// **Gemessen und nicht geschätzt** (2026-09-14, `lo0`): ab zwei Sekunden Frist
// kommt **eine Sekunde** hinzu — die Bibliothek fragt ein zweites Mal. Darunter
// kommt nichts hinzu. Anderthalb Sekunden lassen Luft für einen langsamen
// Rechner, ohne die Zusage weich zu machen.
//
// **Er steht hier und nicht in einem Kommentar**, weil ein Aufrufer ihn
// braucht: `Frist + Nachlauf` ist, was die Wanduhr höchstens zeigt.
const Nachlauf = 1500 * time.Millisecond

// Schlangenlaenge ist, wie viele Antworten zwischengelagert werden.
//
// # Warum die Antworten erst nach dem Fragen gelesen werden
//
// **Weil `hashicorp/mdns` in einen Eintrag weiterschreibt, nachdem es ihn
// gesendet hat.** Nachgelesen in `client.go` (v1.0.7): `query` hält die
// Einträge in einer Map `inprogress` und sendet einen, sobald er
// `complete()` ist — **behält den Zeiger aber** und trägt jedes weitere
// Paket für denselben Namen ein: `inp.InfoFields = rr.Txt` (Z. 340),
// `inp.Port = int(rr.Port)` (Z. 337), `inp.Addr = rr.A` (Z. 353). mDNS
// wiederholt seine Ankündigungen, und über zwei Adressfamilien kommen
// ohnehin mehrere.
//
// **Der erste Anlauf las nebenher**, in einer eigenen Goroutine — und der
// Race-Detektor hat es am 2026-09-14 angestrichen: `deute` las
// `InfoFields`, während `query` es neu zuwies. Es war kein Testartefakt:
// dieselbe Stelle läuft im Betrieb. Wahrscheinlich ist es auch die
// Erklärung für eine Antwort, die vorher **verworfen** wurde, obwohl die
// Bibliothek nur vollständige Einträge sendet.
//
// **Deshalb wird erst gelesen, wenn `QueryContext` zurück ist.** Dann
// schreibt niemand mehr, und was dasteht, ist der **letzte** Stand statt
// eines zufälligen Zwischenstands.
//
// # Was das kostet
//
// **Die Schlange ist jetzt eine Obergrenze.** Vorher hielt der nebenherlaufende
// Leser sie leer; jetzt ist sie der ganze Platz. Die Bibliothek sendet
// **nicht blockierend** (`select { case params.Entries <- inp: default: }`,
// Z. 380) — was nicht hineinpasst, fällt weg, und niemand erfährt davon.
//
// Deshalb 256 und nicht 64: es sind Zeiger, also zwei Kilobyte, und 256
// gleichzeitige MSE-Instanzen in einem Netz sind keine Zahl, die vorkommt.
const Schlangenlaenge = 256

// kennungsmarke ist der Name des TXT-Eintrags mit der Peerkennung.
const kennungsmarke = "kennung="

// namensmarke ist der Name des TXT-Eintrags mit dem selbstgewählten Namen.
const namensmarke = "name="

// Was dieses Paket von `hashicorp/mdns` und vom Netz braucht.
//
// # Warum das Veränderliche sind und keine Rufe
//
// **Weil die Fehlerwege sonst unbetretbar wären.** Keiner der vier tritt auf
// diesem Betriebssystem auf: `Addrs()` gibt für eine erfundene Schnittstelle
// eine **leere Liste** und keinen Fehler (gemessen am 2026-09-14),
// `NewMDNSService` prüft nur Eingaben, die hier vorher geprüft werden, und
// `Shutdown` ist ausdrücklich mehrfach aufrufbar.
//
// **Ein `_ =` daneben wäre die stille Absenkung, die §3 verbietet**, und eine
// Ausnahme in `.covignore` auch: dort steht nur, was ohne echtes Fenster,
// echten Prozess oder echte Plattform nicht läuft. Dieselbe Überlegung wie bei
// `leitung.zufallsquelle`, und derselbe Preis: sie werden **ausschließlich von
// Läufen** getauscht, ohne `t.Parallel()` und mit `t.Cleanup`.
var (
	adresslisteVon = eigeneAdressliste
	dienstBauen    = mdns.NewMDNSService
	serverBauen    = mdns.NewServer
	fragen         = mdns.QueryContext
)

// Netzschnittstelle ist `net.Interface` unter einem Namen, den ein Aufrufer
// nennen kann, ohne `net` zu kennen.
//
// **Für `internal/app`** (§2): der Wächter in
// `internal/schichten/netz_test.go` zählt nach, wer `net` importiert, und die
// Bindungsfassade gehört nicht dazu.
type Netzschnittstelle = net.Interface

// niederlegbar ist, was eine Anmeldung zurücknehmen kann.
//
// **Eine Schnittstelle und kein `*mdns.Server`**, aus demselben Grund wie
// darüber: `Shutdown` klagt auf diesem Rechner nicht, und der Zweig daneben
// will trotzdem betreten werden.
type niederlegbar interface{ Shutdown() error }

// Angebot ist, was eine Anmeldung ankündigt.
type Angebot struct {
	// Kennung ist die Peerkennung dieses Profils — sie **ist** der
	// öffentliche Schlüssel, und sie geht hier an jeden im Netz.
	Kennung peer.Kennung

	// Name ist, wie dieses Profil sich nennt — **kein Nachweis**.
	Name string

	// Port ist, wo der Horchposten steht.
	Port int

	// Schnittstelle ist, worüber angekündigt wird — `nil` heißt: die des
	// Betriebssystems. Ein Lauf gibt `lo0` an und bleibt damit auf diesem
	// Rechner.
	Schnittstelle *net.Interface
}

// Anmeldung ist eine laufende Ankündigung.
// Anmeldung ist eine laufende Ankündigung — **einer je Schnittstelle**.
//
// **Weil `net.ListenMulticastUDP("udp4", nil, …)` nur auf der vom System
// zugewiesenen horcht.** Wer eine Frage vom WLAN nicht hört, beantwortet sie
// auch nicht (NET-02.NAT-01,
// #1706).
type Anmeldung struct {
	dienste  []niederlegbar
	schnitte []*net.Interface
}

// aufstellen stellt die Horchposten auf — [Melde] prüft danach, dass die
// Ankündigung hinausgeht.
func aufstellen(angebot Angebot) (*Anmeldung, error) {
	// **Der Fingerabdruck prüft die Kennung mit** — er ist leer, wenn aus
	// ihr kein Schlüssel wird, und er ist zugleich der Name der Anmeldung.
	// Zwei Prüfungen für zwei Fragen wären hier zwei Antworten auf dieselbe.
	abdruck := peer.Fingerabdruck(angebot.Kennung)
	if abdruck == "" {
		return nil, fmt.Errorf(
			"%w: %q ist keine Peerkennung — ohne sie gibt es nichts "+
				"anzukündigen", domain.ErrValidation, angebot.Kennung)
	}

	if angebot.Port < 1 || angebot.Port > 65535 {
		return nil, fmt.Errorf(
			"%w: %d ist kein Port — eine Ankündigung ohne ihn schickte "+
				"niemanden irgendwohin", domain.ErrValidation, angebot.Port)
	}

	// **Eine genannte Schnittstelle schlägt die Wahl** —
	// `MSE_ENTDECKUNGSSCHNITTSTELLE` ist für Läufe da und für den Menschen,
	// der weiß, was er tut.
	if angebot.Schnittstelle != nil {
		return aufEiner(angebot, angebot.Schnittstelle)
	}

	// **Sonst auf jeder tauglichen** — NET-02.NAT-01.
	taugliche, fehler := tauglicheLesen()
	if fehler != nil {
		return nil, fehler
	}

	// **Ohne taugliche bleibt es beim Betriebssystem** — das ist, was vorher
	// immer galt.
	if len(taugliche) == 0 {
		return aufEiner(angebot, nil)
	}

	return aufAllen(angebot, taugliche)
}

// aufAllen stellt auf jeder Schnittstelle einen Horchposten auf.
//
// **Ein gescheiterter nimmt die anderen nicht mit**: eine Schnittstelle, auf
// der kein Listener aufgeht, ist eine weniger — gescheitert ist die
// Ankündigung erst, wenn **keine** steht.
//
// # Nacheinander, anders als die Suche
//
// **Weil ein Horchposten nur Sockets bindet.** Die Suche wartet je
// Schnittstelle bis zu 6,5 Sekunden — dort kostet die Reihe nach eine halbe
// Minute. Hier kostet sie nichts, und nebenläufig wäre sie nur ein Schloss
// mehr.
func aufAllen(angebot Angebot, taugliche []net.Interface) (*Anmeldung, error) {
	anmeldung := &Anmeldung{}

	var erster error

	for stelle := range taugliche {
		eine, fehler := aufEiner(angebot, &taugliche[stelle])
		if fehler != nil {
			if erster == nil {
				erster = fehler
			}

			continue
		}

		anmeldung.dienste = append(anmeldung.dienste, eine.dienste...)
		anmeldung.schnitte = append(anmeldung.schnitte, eine.schnitte...)
	}

	if len(anmeldung.dienste) == 0 {
		return nil, erster
	}

	return anmeldung, nil
}

// aufEiner stellt einen Horchposten auf.
func aufEiner(angebot Angebot, schnitt *net.Interface) (*Anmeldung, error) {
	// **Der Fingerabdruck ist zugleich der Name der Instanz** — die Prüfung
	// darüber steht in [Melde]; hier wird er nur noch gebraucht.
	abdruck := peer.Fingerabdruck(angebot.Kennung)

	adressen, fehler := adressenVon(schnitt)
	if fehler != nil {
		return nil, fehler
	}

	// **Der Fingerabdruck als Name der Instanz** und die Kennung im TXT.
	//
	// **Weil ein DNS-Etikett höchstens 63 Byte trägt.** Eine Peerkennung ist
	// 52 Zeichen lang und ginge gerade noch hinein — der Abdruck ist die
	// Form, die ein Mensch vergleicht, und er ist eindeutig genug, um zwei
	// Anmeldungen zu trennen.
	//
	// # Warum Rechnername und Adressen selbst gesetzt werden
	//
	// **Weil die Bibliothek sonst zwei Dinge tut, die hier nicht gewollt
	// sind.** Nachgelesen am 2026-09-14 in `zone.go`: ohne Rechnernamen ruft
	// sie `os.Hostname()`, und ohne Adressen ruft sie **`net.LookupIP`**.
	//
	// **Das erste ist eine Preisgabe**, um die niemand gebeten hat: „Davids
	// MacBook Pro" ginge an jeden im Netz, und wer eine Ankündigung erlaubt,
	// erlaubt seine **Peerkennung** — nicht den Namen seines Rechners.
	//
	// **Das zweite ist ein Griff nach draußen** (§3): eine Namensauflösung
	// mitten in der Ankündigung, die auf einem Rechner ohne auflösbaren
	// Namen scheitert — und in einem Lauf das Netz befragte.
	//
	// **Beides fällt weg**, wenn der Name aus dem Abdruck kommt und die
	// Adressen von der Schnittstelle.
	dienst, fehler := dienstBauen(
		abdruck, Dienstart, Bereich+".", abdruck+"."+Bereich+".",
		angebot.Port, adressen,
		[]string{
			kennungsmarke + string(angebot.Kennung),
			namensmarke + angebot.Name,
		})
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Ankündigung zusammenstellen")
	}

	server, fehler := serverBauen(&mdns.Config{
		Zone: dienst, Iface: schnitt, Logger: stilleLogger(),
	})
	if fehler != nil {
		return nil, deuteNetzfehler(fehler, "im Netz ankündigen")
	}

	return &Anmeldung{dienste: []niederlegbar{server}, schnitte: []*net.Interface{schnitt}}, nil
}

// Beende nimmt die Ankündigung zurück.
func (a *Anmeldung) Beende() error {
	// **Jeden Posten, auch wenn einer klagt.** Wer beim ersten Fehler
	// aufhörte, ließe die übrigen stehen — und die strahlten weiter einen
	// Port aus, den niemand mehr hält.
	var erster error

	for _, dienst := range a.dienste {
		if fehler := dienst.Shutdown(); fehler != nil && erster == nil {
			erster = fehler
		}
	}

	if erster != nil {
		return domain.MitGrund(domain.ErrValidation, erster,
			"die Ankündigung zurücknehmen")
	}

	return nil
}

// adressenVon sammelt, unter welchen Adressen angekündigt wird.
//
// **Ohne Schnittstelle: die aller.** Wer keine nennt, meint „wo auch immer
// dieser Rechner erreichbar ist" — und das ist die Vorgabe des
// Betriebssystems, nur eben ohne Namensauflösung.
//
// **Ohne eine einzige Adresse gibt es nichts anzukündigen**, und das ist ein
// Befund: eine Ankündigung ohne Adresse schickte jeden ins Leere.
func adressenVon(schnitt *net.Interface) ([]net.IP, error) {
	rohe, fehler := adresslisteVon(schnitt)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Adressen dieses Rechners lesen")
	}

	adressen := make([]net.IP, 0, len(rohe))

	for _, roh := range rohe {
		if netz, taugt := roh.(*net.IPNet); taugt {
			adressen = append(adressen, netz.IP)
		}
	}

	if len(adressen) == 0 {
		return nil, fmt.Errorf(
			"%w: dieser Rechner hat keine Adresse, unter der er sich "+
				"ankündigen könnte", domain.ErrValidation)
	}

	return adressen, nil
}

// SchnittstelleNamens findet eine Netzwerkschnittstelle.
//
// # Warum dieser Weg hier steht
//
// **Weil `internal/app` `net` nicht kennen darf** (§2) — der Wächter in
// `internal/schichten/netz_test.go` zählt nach, und `net.InterfaceByName`
// brächte den ganzen Weg ins Netz in die Bindungsfassade. Hier ist er ohnehin
// schon.
//
// **Ein unbekannter Name gibt `nil` und einen Fehler** — was daraus wird,
// entscheidet der Aufrufer; ein Programm, das wegen eines Tippfehlers gar
// nicht suchte, wäre die schlechtere Antwort.
func SchnittstelleNamens(name string) (*net.Interface, error) {
	schnitt, fehler := net.InterfaceByName(name)
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die genannte Netzwerkschnittstelle finden")
	}

	return schnitt, nil
}

// eigeneAdressliste fragt eine Schnittstelle — oder alle.
//
// **`&net.Interface{}` ist nicht „keine"**, sondern „alle": ein Index von
// null gab am 2026-09-14 neunzehn Adressen. Deshalb entscheidet hier `nil`
// und nicht ein leerer Verbund.
func eigeneAdressliste(schnitt *net.Interface) ([]net.Addr, error) {
	if schnitt != nil {
		return schnitt.Addrs()
	}

	return net.InterfaceAddrs()
}
