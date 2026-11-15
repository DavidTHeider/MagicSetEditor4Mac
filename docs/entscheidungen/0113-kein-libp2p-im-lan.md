# 0113 — kein libp2p im LAN

**Datum:** 2026-09-14 · **Stand:** entschieden ·
**Berührt:** [0112](0112-tailcat-statt-eines-eigenen-dienstes.md) ·
**Anlass:** Werkstück N2 von Phase 5.4
(#1455),
auf Nachfrage am 2026-09-14

## Die Frage

**#1455 schreibt seit seiner Anlage am 2026-09-07 einen go-libp2p-Adapter vor**
— „mit gepinnter Version, mDNS-Discovery und QUIC-Verbindung auf N1". Zu diesem
Zeitpunkt gab es N1 noch nicht, und N3 sollte über Pion mit eigenem Dienst
laufen.

**Beides hat sich geändert, und zwar in die Richtung, die libp2p überflüssig
macht.**

## Was N1 bereits tut

Es ist genau die Liste, für die man libp2p sonst nimmt:

| Zusage | Wo sie schon steht |
| --- | --- |
| Identität eines Knotens, als Kennung, die den Schlüssel trägt | `internal/peer` — Ed25519, die Kennung ist die eines libp2p-Peers |
| Nachrichten, die niemand fälschen kann | `internal/peer` — Hülle mit kanonischer Unterschriftsform und Bereichstrenner |
| Ein Verbindungsaufbau, der beide Seiten bindet | `internal/peer` — Dreischritt-Handschlag, Kanalbindung, Wiedereinspielschutz |
| Nachrichtengrenzen auf einem Strom | `internal/leitung` — Längenpräfix, Grenze **vor** der Speicherstellung |

**libp2p brächte davon nichts Neues, sondern ein zweites Mal dasselbe.** Seine
Sicherheitsschicht (Noise oder TLS) und sein Identitätsbegriff lägen **neben**
denen aus N1, nicht an ihrer Stelle — und zwei Handschläge übereinander sind
nicht sicherer, sondern schwerer zu prüfen.

## Was von libp2p übrig bliebe

**Drei Dinge**, und keines davon wiegt 136 Module:

- **mDNS** — dafür gibt es eine Bibliothek von drei Modulen.
- **QUIC** — dafür gibt es `quic-go`, und es kostet 16.
- **NAT-Durchdringung über Relay und DCUtR** — und **dafür ist seit
  [0112](0112-tailcat-statt-eines-eigenen-dienstes.md) tailcat zuständig.** Das
  ist der Posten, der libp2p überhaupt gerechtfertigt hätte; er ist weg.

## Die Entscheidung

**N2 setzt auf TCP aus der Standardbibliothek und eine mDNS-Bibliothek.** Kein
libp2p, und vorerst auch kein QUIC.

**Gemessen am 2026-09-14**, in einem eigenen `GOMODCACHE`, gegen einen Graphen,
der heute 147 Module und neun direkte Abhängigkeiten führt:

| Weg | neue Module | Modul-Cache |
| --- | --- | --- |
| `go-libp2p v0.49.0` | **136** | 487 MB |
| `quic-go v0.62.0` + mDNS | ~20 | — |
| **TCP (Standardbibliothek) + `hashicorp/mdns v1.0.7`** | **3** | — |

Die drei sind `hashicorp/mdns`, `miekg/dns` und `golang.org/x/telemetry`; alles
Übrige aus dessen Graphen steht hier schon.

**Die mDNS-Bibliothek ist `hashicorp/mdns`**, und die Wahl ist eine Frage der
Pflege, nicht der Fläche. Abgerufen am 2026-09-14 über den Modulproxy:

| Bibliothek | jüngste Fassung | veröffentlicht |
| --- | --- | --- |
| `github.com/hashicorp/mdns` | v1.0.7 | **2026-06-11** |
| `github.com/brutella/dnssd` | v1.2.14 | 2024-10-22 |
| `github.com/libp2p/zeroconf/v2` | v2.2.0 | 2022-08-07 |
| `github.com/grandcat/zeroconf` | v1.0.0 | 2020-01-29 |

## Was das kostet

**0. Die Vertraulichkeit ist keine Zugabe mehr, sondern eigene Arbeit.**

**Dieser Posten stand hier nicht, als der Vermerk am 2026-09-14 geschrieben
wurde**, und er ist der schwerste. Er ist eine Stunde später beim Bauen
aufgefallen, und §10 verlangt, dass er hier steht statt in einer Fußnote: eine
Kostenliste, die den teuersten Posten auslässt, ist schlimmer als keine.

**Der Befund.** `internal/peer` **unterschreibt und verschlüsselt nicht**.
`Huelle` trägt eine `Unterschrift`, `Identitaet` hat `Signiere`, und nirgends
im Stapel steht ein Siegel. libp2p hätte Noise oder TLS **mitgebracht**; mit
blankem TCP gingen Kartendaten, Namen und Bestände im Klartext über das Netz —
über das LAN, und über eine „Adresse von Hand" auch über jede Strecke
dazwischen.

**Die Antwort ist TLS aus der Standardbibliothek**, und sie kostet **null
Module**: ein selbst unterschriebenes Zeugnis, dessen Schlüssel die
Peeridentität **ist**, und eine Prüfung, die gegen die erwartete Kennung
festnagelt statt gegen eine Zertifizierungsstelle. Dasselbe, was `libp2p-tls`
tut.

**Der wirkliche Tausch lautet damit: 136 Module gegen rund zweihundert Zeilen
sicherheitsnahen Code, den dieses Projekt selbst prüfen muss.** Das ist etwas
anderes als „drei Module gegen 136", und wer die Entscheidung später liest,
soll es so lesen.

**Ein Nebenertrag gehört dazu**: TLS liefert die **Kanalbindung** als Ausfuhr
von Schlüsselmaterial (RFC 5705, `ExportKeyingMaterial`) — eindeutig je
Verbindung und **geheim**. Ohne TLS wäre der Kanal aus dem Vierertupel
abzuleiten: auf beiden Seiten gleich und je Verbindung eindeutig, aber für
jeden auf der Strecke lesbar. Der Handschlag aus N1 bindet damit an etwas
Stärkeres, als er es ohne diese Entscheidung täte.

**1. Keine Multiplexierung über eine Verbindung.** QUIC trüge mehrere Ströme
über einen Aufbau; TCP braucht je Strom einen. Der Preis ist hier klein, weil
`Oeffne(kennung, zweck)` ohnehin **eine Verbindung je Zweck** führt — eine
Partie und eine Übertragung sind zwei Sitzungen, nicht zwei Ströme einer. Teuer
würde es erst, wenn ein Zweck viele nebenläufige Ströme brauchte; N4 ist
dahingehend zu prüfen, bevor es gebaut wird.

**2. Kein 0-RTT und keine Wiederaufnahme über einen Netzwechsel hinweg.** QUIC
überlebt einen Wechsel der Adresse, TCP nicht. Für #1455 ist das kein Verlust,
sondern eine benannte Zusage: „Netzwechsel, Sleep/Wake … Wiederverbindung
meldet Status an fachliche Resume-Schicht" — die Wiederaufnahme ist **ohnehin**
fachlich zu bauen und darf nicht am Transport hängen.

**3. Kein Weg zurück zu Relay und DCUtR ohne Umbau.** Bewährte sich tailcat
nicht und wäre libp2p die Rückfallebene, käme der Adapter dann erst. Das ist
die Rückseite von [0112](0112-tailcat-statt-eines-eigenen-dienstes.md) und wird
dort schon getragen: Pion ist zurückgestellt, nicht verworfen.

## Warum der Tausch billig ist

**Weil die Naht dafür gebaut ist.** [service.Waehler] gibt eine
[service.Verbindung], und `internal/leitung` macht aus **jedem** `net.Conn`
eine. TCP gibt einen, ein QUIC-Strom gibt einen, tailcats `OnTCP` gibt einen.
Wer später QUIC will, tauscht ein Blatt.

**Und die Gegenprobe dazu steht schon da**: `internal/leitung.Schleife` ist
derselbe Typ ohne Netz, und der Lauf über zwei Anwendungen in
`internal/app/peerpaarung_intern_test.go` fährt jede Schicht darüber.

## Verworfen

**`go-libp2p` wie im Issue** — nicht, weil es schlecht wäre, sondern weil nach
N1 und [0112](0112-tailcat-statt-eines-eigenen-dienstes.md) nur noch mDNS und
QUIC von ihm übrig sind und beides einzeln billiger zu haben ist. Die
Entscheidung ist an N1 und an tailcat gebunden; fällt eines von beiden, ist sie
neu zu stellen.

**`quic-go` sofort** — es wäre die teurere Hälfte eines Nutzens, den heute
niemand einlöst. Es kommt, wenn ein Werkstück es **braucht**, und dann mit der
Messung, die das zeigt. **Für die Vertraulichkeit braucht es ihn nicht**: QUIC
brächte TLS 1.3 mit, aber `crypto/tls` bringt es auch, und über TCP ebenso.

**Blankes TCP ohne TLS** — es wäre der kürzeste Weg zu einem laufenden
Werkstück und der einzige, der wirklich nicht ginge. Unterschrieben heißt
nicht verschlossen: ein Mitleser auf der Strecke kann keine Nachricht fälschen,
aber jede lesen, und was hier hinausgeht, sind Kartenbestände, Namen und
Lagerorte.

**`grandcat/zeroconf`** — die naheliegendste Wahl und die am schlechtesten
gepflegte: letzte Fassung Januar 2020, und ihr `miekg/dns` ist von 2019.
