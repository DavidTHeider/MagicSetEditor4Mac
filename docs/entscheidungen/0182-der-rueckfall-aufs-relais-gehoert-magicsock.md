# 0182 — Der Rückfall aufs Relais gehört magicsock, und die Lücke ist seine Frist

**Stand:** angenommen am 2026-10-10 · D03
(#2234),
mit Blick auf D04 (#2235).

## Die Frage

D03 verlangt drei Dinge: den **Direktweg** belegen, **UDP gezielt
blockieren** und den **automatischen Rückfall** aufs Relais samt **Rückkehr**
nachweisen — mit Weg, Latenz und Ausfallzeit. Trägt die Architektur das
nicht, ist die Zusage zu begrenzen.

## Wer den Rückfall macht

**Nicht diese Anwendung.** Der Fernweg ist tailcat (`internal/fernleitung`,
[0112](0112-tailcat-statt-eines-eigenen-dienstes.md),
[0114](0114-tailcat-gepinnt.md)), und darunter wählt Tailscales `magicsock`
je Paket den Weg. Nachgelesen am 2026-10-10 in tailscale.com
v1.103.0-pre.0.20260916030321-a2263542f260:

- Ein direkter Weg gilt **`trustUDPAddrDuration` = 6,5 s** nach seiner
  letzten Bestätigung (`wgengine/magicsock/magicsock.go`, Zeile 4014);
  bestätigt wird er über Disco-Pings, der Herzschlag ist 3 s (Zeile 4010).
- Ist die Frist um, schickt magicsock **jedes Paket an beide**, den alten
  direkten Weg und DERP (`addrForSendLocked`, `endpoint.go` ab Zeile 632).
  Das ist der Rückfall: DERP trägt, sobald der direkte Weg nicht mehr
  bestätigt ist.
- Kommt eine Bestätigung wieder, gilt der direkte Weg wieder allein.

## Was gemessen ist

Am 2026-10-10 auf einem Mac (Mac17,3, macOS 27.0.1), beide Enden in einem
Prozess, über das öffentliche DERP von Tailscale, Tag `netzlauf`
(`internal/fernleitung/`):

| Weg | gemessen | Lauf |
| --- | --- | --- |
| **direkt** (über die Schleife) | Echo unter 1 ms; der Stand nennt `direkt` samt Relaisregion | `fernweg_netzlauf_test.go` |
| **nur Relais** (`TS_DEBUG_ALWAYS_USE_DERP`, vor dem Aufbau) | DERP Frankfurt, 20 Echos: Median 26–41 ms, längstes 266–286 ms (das erste, mit Wegaufbau) | `rueckfall_netzlauf_test.go` |
| **UDP einer Seite gesperrt**, mitten in der Verbindung | **Lücke 6,456–6,457 s** in drei Läufen — das längste Echo; danach über das Relais im Median 33–57 ms; kein Echo verloren | `udpsperre_netzlauf_test.go` |
| **UDP wieder offen** | direkt nach 299–345 ms, ein Echo wieder unter 1 ms | dito |
| **nur die Pfadsuche gesperrt** (`TS_DEBUG_NEVER_DIRECT_UDP`) | der Stand nennt das Relais nach 6,55–6,58 s; die Daten liefen weiter direkt | `rueckfall_netzlauf_test.go` |

**Wie UDP gesperrt wird, ohne Systemeinstellung**: magicsock bindet seine
Sockets neu (`Conn.Rebind`), und ist dabei `TS_DEBUG_ALWAYS_USE_DERP`
gesetzt, nimmt es statt eines Sockets seinen **Verwerfer** — Schreiben geht
still verloren, Lesen bringt nichts (`newBlockForeverConn`,
`magicsock.go` Zeile 3706; `blockforever_conn.go` Zeile 33). Das ist das
Verhalten einer Firewall, die UDP verwirft; DERP spricht TCP und bleibt.
An `Rebind` kommt der Lauf über die unexportierten Felder des gepinnten
tailcat (`Server.lb.sys.MagicSock`) und scheitert laut, wenn sich der Weg
ändert. **Gegenprobe**: ohne das Neubinden bleibt ein Echo „gesperrt" unter
1 ms, und der Lauf wird rot.

**Wer es weiß**: `Rebind` setzt die Wege **der gesperrten Seite** zurück
(`resetEndpointStates`, Zeile 3829) — sie schickt sofort über DERP. **Die
andere Seite weiß nichts** und schickt an den toten Port, bis ihr direkter
Weg nach 6,5 s seine Geltung verliert; daher die Lücke, an ihr gemessen.
Das ist die Lage eines Menschen, dessen Gegenüber hinter eine Firewall
gerät. Bei einer Firewall auf der **eigenen** Seite wüsste auch diese Seite
nichts und schickte bis zum Ende ihrer eigenen Frist an den toten Weg — die
Frist ist dieselbe, die Lücke deshalb von derselben Größe.

## Was nicht gemessen ist

- **Zwei unabhängige NAT-Grenzen.** Beide Enden stehen auf einem Mac; zwei
  physische Macs sind außerhalb des Auftrags. Das betrifft D04 ebenso.
- **Eine Firewall außerhalb von magicsock.** Gesperrt wird in magicsock,
  nicht im Netz; ein Netz, das Pakete erst verzögert und dann verwirft,
  verhält sich zwischen den Zuständen anders.

## Die Entscheidung

**Kein eigener Rückfall.** Diese Anwendung baut keinen zweiten Mechanismus
über magicsock; einer daneben stritte mit ihm um denselben Weg.

**Gemessen** (drei Läufe, UDP einer Seite gesperrt): der Verkehr stand
6,46 s, danach trug das Relais, ohne dass eine Nachricht verloren ging; war
UDP wieder offen, ging es binnen einer Sekunde wieder direkt. Die Anzeige im
Fenster folgte nach derselben Frist.

**Zugesagt ist weniger, und zwar das, was der Netzlauf prüft**
(`internal/fernleitung/udpsperre_netzlauf_test.go`): eine Lücke unter
30 s und die Rückkehr zum direkten Weg binnen 90 s. Die **6,5 s** sind die
Vertrauensfrist, nach der magicsock einen Weg neu wählt
(`wgengine/magicsock/endpoint.go`, Zeilen 632–648, in der gepinnten
Fassung von tailcat) — **keine Zustellfrist**: DERP-Verfügbarkeit,
TCP-Wiederholung und eine erneute Pfadsuche kommen hinzu. Hier stand bis zum
Zweitblick über Phase D (Befund 4) „höchstens 6,5 s" als Zusage — mehr, als
Quelle und Wächter trugen. **Nicht zugesagt** ist außerdem ein Verhalten über
zwei NAT-Grenzen, das niemand gemessen hat.

## Verworfen

- **UDP mit `pf` sperren.** Eine Systemeinstellung, die einem Menschen
  gehört (AGENTS.md §17); im Gast einer VM bräuchte es `sudo` und ein
  Kennwort.
- **`TS_DEBUG_NEVER_DIRECT_UDP` als Nachweis einer Lücke.** Der Knopf sperrt
  die Pfadsuche, nicht die Daten: nach der Frist geht jedes Paket an beide
  Wege, und über die offene Schleife kam keine Lücke. Der erste Entwurf
  hielt 9 ms für den Rückfall — ein Echo „über das Relais" in 1 ms ist über
  ein Relais im Internet unmöglich.
