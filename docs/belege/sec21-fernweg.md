# SEC-21 — Ein offener Fernstrom und was der Stand darüber sagt

**Gemessen am 2026-10-07** zu #2162,
an Commit `f946e2113` plus der Umsetzung auf `feat/2086-phase-h`, mit
`tailscale/tailcat v0.6.0` und `tailscale.com v1.101.0-pre.0.20260720143344-246c82a658b3`.

## Die Frage

Der Kasten „Über fremde Netze erreichbar" sagte bei offener Verbindung
„Noch hat niemand diese Einladung benutzt." Warum — und was stattdessen
gilt?

## Was tailcat sagt, nachgelesen

- `Server.Status()` gibt `lb.Status()` (`tailcat.go`, Zeilen 2418–2420).
- `locoBackend.Status()` baut einen `ipnstate.StatusBuilder` mit
  Nullwerten — **ohne `WantPeers`** — und reicht ihn an magicsock und die
  Engine (Zeilen 1653–1659).
- `magicsock.Conn.UpdateStatus` trägt Gegenstellen nur mit `WantPeers` ein
  (`wgengine/magicsock/magicsock.go`, Zeilen 3973–3978).

**Damit ist `Status().Peer` immer leer**, und die öffentliche API bietet
keine zweite Quelle: das Backend und seine magicsock sind nicht exportiert.
**Entscheidung**: der Weg je Gegenüber ist mit dieser Fassung **nicht zu
ermitteln**; ob jemand angebunden ist, zählt der Horcher selbst
(`fernleitung.Horcher.Angebunden`), und der Kasten unterscheidet
„niemand", „angebunden, Weg unbestimmt" und — sollte eine spätere Fassung
Gegenstellen nennen — den Weg.

## Die Läufe

`internal/fernleitung/fernweg_netzlauf_test.go` (Bauschalter `netzlauf`):
Horcher und Wählender **in einem Prozess auf diesem Mac**, verbunden über
tailcat und ein DERP-Relais von Tailscale; eine Nachricht hin, eine zurück.

| Lauf | Echo | `Wege()` | `Angebunden()` | Ergebnis |
| --- | --- | --- | --- | --- |
| gewöhnlich | nach 1,05 s, 19 Bytes | 0 | 1 | PASS |
| `TS_DEBUG_ALWAYS_USE_DERP=1` (Protokoll: `disabled udp4/udp6`) | nach 4,30 s, 19 Bytes | 0 | 1 | PASS |
| gewöhnlich, **Zählung auf 0 gesetzt** (Gegenprobe) | nach 1,05 s | 0 | 0 | **FAIL** `0 angebunden, erwartet 1` |

Die dritte Zeile ist der rote Lauf am echten Draht: so stand es, bevor der
Horcher zählte — eine offene Verbindung, und niemand sagte es.

## Seit tailcat v0.7.0: der Weg — K01

**Am 2026-10-07**, K01 (#2187). tailcat v0.7.0 setzt in
`locoBackend.Status` den `StatusBuilder` mit `WantPeers` (in der Quelle
nachgelesen); derselbe Lauf verlangt jetzt **einen** Weg — gewöhnlich einen
direkten, mit erzwungenem DERP das Relais und keinen direkten — und fragt bis
15 s nach, weil der direkte Weg erst mit Verkehr im Tunnel entsteht
(`TestEinOffenerFernwegNenntSeinenWeg`).

| Lauf | tailcat | Echo | `Wege()` | Ergebnis |
| --- | --- | --- | --- | --- |
| gewöhnlich, **vorher** | v0.6.0 | nach 1,06 s | `[]` — 15 s lang | **FAIL** `tailcat nennt 0 Wege, erwartet einen` |
| gewöhnlich | v0.7.0 | nach 1,05 s | `[{Direkt:192.168.178.74:57160 Relais:fra}]` | PASS, 1,80 s |
| `TS_DEBUG_ALWAYS_USE_DERP=1` | v0.7.0 | nach 4,14 s | `[{Direkt: Relais:fra}]` | PASS, 4,87 s |

**Die Schwärzung hält auch unter v0.7.0**: die `NetworkMap`-Zeilen tragen
`nodekey:…` und `discokey:…`. Zwischen v0.6.0 und v0.7.0 kam genau eine
Logzeile hinzu (`exec %v: %v`, für einen Dienst, den diese Anwendung nicht
nutzt), ohne Schlüsselmaterial.

## Keine Einladung im Protokoll

Eine Einladung besteht aus dem öffentlichen Knotenschlüssel, dem
Disco-Schlüssel („the unguessable part of the server's tailcat address") und
dem Pre-Shared Key (`ConnInfo`, `tailcat.go`, Zeilen 155–169); `Server.Start`
erzeugt den Pre-Shared Key, wenn keiner gesetzt ist (Zeilen 558–562).

| Protokoll des Laufs | Pre-Shared Key, Adresse, `mse4mac:1:` | Knoten-/Disco-Schlüssel im Klartext |
| --- | --- | --- |
| vor der Schwärzung | 0 | ja (`NetworkMap`, `disco pub key`) |
| nach der Schwärzung (`protokoll.go`) | 0 | **0** — 15 Stellen als `nodekey:…`/`discokey:…` |

## Was nicht belegt ist

- **Zwei Rechner.** Beide Enden stehen auf einem Mac; der „direkte" Weg geht
  über die Schleife. Zwei physische Macs sind außerhalb des Auftrags.
- **Die VM-Probe** aus dem Audit (zwei Gäste, normal und mit abgeschaltetem
  UDP) ist hier nicht wiederholt; ihr Befund — leerer Stand bei offenem
  Echo — ist derselbe, den dieser Lauf zeigt.
- **Flächen**: der Fernstand ist nur am Fenster (`docs/FAEHIGKEITEN.md`,
  „Über fremde Netze erreichbar sein"); Kommandozeile, REST und MCP haben
  ihn nicht, also auch keine zweite Anzeige.
