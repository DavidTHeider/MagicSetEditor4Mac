package entdeckung

// **Eine Ankündigung muss hinausgehen** — K02
// (#2188).
//
// # Warum ein Probepaket
//
// **Weil das Binden nichts beweist.** Gemessen am 2026-10-07 in einem
// macOS-15.2-Gast mit der gebauten Anwendung: verwehrt macOS die Erlaubnis
// *Lokales Netzwerk*, bindet der mDNS-Server seine Sockets trotzdem. Gesendet
// wird erst als Antwort auf eine Frage — und deren Fehler verschluckt
// `hashicorp/mdns`. Die Anwendung zeigte „kündigt sich an", und niemand im
// Netz sah sie.
//
// **Apple rät selbst dazu**: „connect a UDP socket to a local network address"
// löst die Rückfrage aus, und ohne Erlaubnis endet ein Senden an eine lokale
// Adresse mit einem Fehler (TN3179, nachgelesen am 2026-10-07). Ein Paket an
// die mDNS-Gruppe ist genau so ein Senden — und es ist eine gewöhnliche Frage
// nach der eigenen Dienstart, die jeder Teilnehmer ohnehin versteht.
//
// # Eine Schnittstelle genügt
//
// **Weil die Erlaubnis dem Prozess gilt, nicht der Schnittstelle.** Geht es
// über eine hinaus, ist die Ankündigung in der Welt; scheitert es überall,
// sagt der erste Fehler, warum — gedeutet wie bei der Suche.

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

// Melde kündigt dieses Profil im Netz an — **und erst, wenn es hinausgeht.**
func Melde(angebot Angebot) (*Anmeldung, error) {
	anmeldung, fehler := aufstellen(angebot)
	if fehler != nil {
		return nil, fehler
	}

	if fehler := hinaus(anmeldung.schnitte, probeSenden); fehler != nil {
		_ = anmeldung.Beende() //nolint:errcheck // der Fehler davor ist der, der zählt

		return nil, fehler
	}

	return anmeldung, nil
}

// mdnsGruppe ist die Multicastadresse von mDNS über IPv4 (RFC 6762, 3).
var mdnsGruppe = &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}

// probeSenden schickt die Probe — eine Naht für Läufe, wie die übrigen.
var probeSenden = mdnsProbe

// hinaus prüft, ob die Probe über eine der Schnittstellen hinausgeht.
func hinaus(schnitte []*net.Interface, senden func(*net.Interface) error) error {
	var erster error

	for _, schnitt := range schnitte {
		fehler := senden(schnitt)
		if fehler == nil {
			return nil
		}

		if erster == nil {
			erster = fehler
		}
	}

	if erster == nil {
		return nil
	}

	return deuteNetzfehler(erster, "die Ankündigung hinausschicken")
}

// probefrist begrenzt das Öffnen der Steckdose — gesendet wird ohne Warten.
const probefrist = 2 * time.Second

// probeSteckdose öffnet die Steckdose der Probe — eine Naht, damit der Weg,
// auf dem das Betriebssystem keine hergibt, betreten werden kann.
var probeSteckdose = func(ctx context.Context) (net.PacketConn, error) {
	return (&net.ListenConfig{}).ListenPacket(ctx, "udp4", "0.0.0.0:0")
}

// mdnsProbe schickt eine Frage nach der eigenen Dienstart an die mDNS-Gruppe —
// über die genannte Schnittstelle, ohne sie über die des Systems.
func mdnsProbe(schnitt *net.Interface) error {
	ctx, abbrechen := context.WithTimeout(context.Background(), probefrist)
	defer abbrechen()

	verbindung, fehler := probeSteckdose(ctx)
	if fehler != nil {
		return fehler
	}

	defer func() { _ = verbindung.Close() }() //nolint:errcheck // eine Probe, nichts zu retten

	paket := ipv4.NewPacketConn(verbindung)
	if schnitt != nil {
		if wahlfehler := paket.SetMulticastInterface(schnitt); wahlfehler != nil {
			return wahlfehler
		}
	}

	_, fehler = paket.WriteTo(probefrage(), nil, mdnsGruppe)

	return fehler
}

// probename ist `_mse4mac._tcp.local.` als DNS-Etiketten — fest geschrieben
// statt aus [Dienstart] gebaut, damit keine Länge in ein Byte gezwängt wird;
// dass beide dasselbe sagen, hält ein Lauf.
const probename = "\x08_mse4mac\x04_tcp\x05local\x00"

// probefrage ist eine DNS-Frage nach PTR für [probename] — ohne Kennung, ohne
// Antworten, wie RFC 6762 eine Multicastfrage beschreibt.
func probefrage() []byte {
	return append([]byte{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}, probename+"\x00\x0c\x00\x01"...)
}
