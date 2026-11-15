# 0112 — tailcat statt eines eigenen Dienstes

**Datum:** 2026-09-14 · **Stand:** entschieden ·
**Berührt:** [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) ·
**Anlass:** Werkstück N3 von Phase 5.4
(#1456),
auf Nachfrage am 2026-09-14

## Die Frage

**Wie verbinden sich zwei Macs, die nicht im selben Netz stehen?**

N2 löst das LAN: mDNS findet das Gegenüber, ein Strom trägt die Bytes, und
niemand braucht eine Adresse. Für „einer zu Hause, einer irgendwo" half davon
nichts — und der geplante Weg war **Pion** plus ein selbst betriebener Dienst
`mse-connect` für Signaling, dazu TURN als Relais.

> Hier stand „QUIC trägt die Bytes". **Welcher Transport es im LAN ist,
> entscheidet seit dem 2026-09-14 [0113](0113-kein-libp2p-im-lan.md)** — und es
> ist TCP, weil nach N1 und diesem Vermerk von libp2p nur mDNS und QUIC übrig
> blieben und beides einzeln billiger zu haben ist.

**Der schwere Teil war nie die Bibliothek.** Er stand im Issue selbst: Betreiber,
Repositorium, Endpunkt, Budget, Quoten, TLS, Aufbewahrung, Monitoring,
Missbrauchsverfahren. Ein Dienst ist eine Dauerverpflichtung, und dieses
Projekt hat niemanden, der sie trägt.

**Und libp2p nimmt sie nicht ab.** Nachgelesen am 2026-09-14 in `libp2p/specs`
(`connections/hole-punching.md`): zwei **private** Knoten verbinden sich über
`identify` und AutoNAT auf beiden Seiten, **eine Circuit-Relay-v2-Reservierung
beim Ziel** und ein per DCUtR koordiniertes Durchstoßen. libp2p **kann** den
Internetfall — aber nur mit einem öffentlich erreichbaren Relais. Die
Infrastrukturfrage verschwindet dort nicht, sie wandert von „Signaling und
TURN" zu „Rendezvous und Relais".

## Die Entscheidung

**N3 läuft über [`tailscale/tailcat`](https://github.com/tailscale/tailcat)**,
nicht über Pion mit eigenem Dienst.

tailcat ist Tailscales **Datenebene ohne deren Kontrollebene**: WireGuard
verschlüsselt, magicsock durchstößt NATs, DERP ist Rendezvous **und**
Relais-letzter-Instanz. Nachgelesen am 2026-09-14 im README:

- **Kein Tailscale-Konto, kein Auth-Key, keine Kontrollebene.** Die
  Verbindungsangaben werden *„exchanged out of band, however you want"*.
- Die `Addr` trägt die WireGuard- und Pfadfindungsschlüssel des Servers, ein
  optionales PSK und die DERP-Region. **Sie ist damit die Einladung**, die
  #1456 ohnehin verlangt — Code und QR sind zwei Gestalten derselben Sache.
- Die Go-Fläche ist `Server`, `Client`, `Addr`; `Server.OnTCP` reicht
  **`net.Conn`**, `Client.DialTCPPort(ctx, port)` wählt.
- **BSD-3-Clause** — geht in ein GPL-2.0-or-later-Projekt hinein (§11).

**Das Letzte ist der Grund, warum dieser Tausch billig ist.** `net.Conn` ist
genau die Form, die [service.Verbindung] erwartet. Der Adapter ist ein Blatt
hinter [service.Waehler]; N1 hat diese Naht gebaut, damit ein Transport
austauschbar ist, und hier wird sie zum ersten Mal dafür benutzt.

**N2 bleibt, wie es geschnitten war**, und bekommt einen Zusatz: einen Weg
**„Adresse von Hand"**. mDNS läuft nicht über jedes Netz, das zwei Rechner
verbindet — nicht über ein Tailnet, nicht über ein VPN, nicht über eine
Portweiterleitung. Dort ist das Gegenüber **erreichbar** und nur nicht
**auffindbar**; ein direktes Wählen genügt, und der Handschlag aus N1 bindet
die Identität ohnehin an den Kanal und nicht an die Route. Damit werden
Tailscale, ZeroTier, WireGuard und Portweiterleitung zu **unterstütztem
Betrieb** statt zu eingebauten Transporten — ohne eine einzige neue
Abhängigkeit.

## Was der Tausch kostet

**Er steht hier und nicht in einer Fußnote**, weil eine Entscheidung ohne ihren
Preis keine ist.

**1. Keine Stabilitätszusage.** Wörtlich: *„the Go API, the CLI flags and
output, and the wire format may all change."* Ein Projekt, das jedes Werkzeug
auf eine Fassung pinnt (§1), bekommt eine Abhängigkeit, deren **Drahtformat**
sich ändern darf — zwei Fassungen von MSE4Mac könnten aufhören, miteinander zu
reden. Das ist der schwerste der drei Posten, und deshalb steht die Antwort
darauf als **erster** Punkt im Umfang von #1456: die Fassung wird gepinnt, und
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) bekommt einen Eintrag
darüber, welche Fassungen miteinander sprechen, wie eine Abweichung erkannt
wird und was Nutzende dann sehen. **Ein Adapter vor dieser Antwort wäre eine
Zusage, die niemand einlösen kann.**

**2. Öffentliches DERP ohne Zusage.** *„no uptime SLAs or throughput targets"*,
dazu *„rate-limited throughput if you're using our public hosted DERP relays"*,
wenn die direkte Route scheitert; Übertragungen sind *„not compressed"*. Für
den Objektaustausch (N4,
#1457) ist der
Durchsatz eine reale Größe.

**3. Die Betriebsverantwortung wandert, sie verschwindet nicht.** Aus „wir
betreiben `mse-connect`" wird „wir hängen an `tailcat.dev/derpmap.json` und an
Relais, die wir nicht steuern". Das ist günstiger — kein Budget, kein
Missbrauchsverfahren, keine TLS-Verlängerung —, aber es ist eine Abhängigkeit,
und sie gehört nach `docs/NETZEIGENSCHAFTEN.md` und in die Release-Notiz.

## Was nicht entschieden ist

**Ob tailcat eine Datagrammfläche hat.** Belegt ist die TCP-Seite (`OnTCP`,
`DialTCPPort`); alles Weitere stand nur in einer Suchzusammenfassung, und eine
solche ist kein Prüfergebnis (§10). Davon hängt ab, ob `Verwerfbar` — die
Präsenz aus N1 — **unter** oder **über** dem Transport modelliert wird.
Nachzulesen am Quelltext der gepinnten Fassung, bevor der Adapter entsteht.

**Und ob es überhaupt dazu kommt.** #1456 trägt weiter seinen
Entscheidungspunkt: fällt die Fassungsfrage negativ aus, ist ein negatives
Ergebnis **keine Implementierung**, und der verbleibende Umfang wird am
Sammelissue festgehalten.

## Verworfen

**Pion mit eigenem `mse-connect`** — nicht wegen der Bibliothek, sondern wegen
des Dienstes. Es ist **zurückgestellt und nicht verworfen**: fällt Kosten 1
negativ aus, ist es die Rückfallebene, und dann gilt #1456 in seiner Fassung
vom 2026-09-07.

**`tsnet` als eingebettetes Tailscale** — es täte dasselbe und wäre der
naheliegende erste Gedanke, **braucht aber einen `AuthKey`**: ein
Tailscale-Konto je Nutzer und beide im selben Tailnet. Für zwei Menschen, die
sich auf einem Turnier getroffen haben und Karten tauschen wollen, ist das eine
härtere Voraussetzung als eine Einladung. tailcat ist genau dieselbe Technik
**ohne** diese Voraussetzung.

**Ein drittes Underlay neben libp2p und tailcat** — §16 verlangt an jeder
Fläche dieselbe Antwort auf dieselbe Fähigkeit. Jeder weitere Transport
verdreifacht die Fläche, über die das zu belegen ist.
