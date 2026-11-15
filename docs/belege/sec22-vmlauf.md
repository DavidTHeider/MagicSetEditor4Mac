# Beleg: zwei Maschinen im eigenen Netz und über DERP — SEC-22, ABN-01.AK-05

**Gefahren am 2026-10-07** mit der Freigabe des Auftraggebers für die zwei
Lume-Gäste aus dem Audit (`mse070-a`, `mse070-b`, macOS 15.3.1, je 4 Kerne
und 8 GB), beide **überbrückt** im Netz dieses Hauses (192.168.178.90 und
.94). Die Läufe stehen hinter dem Tag `vmlauf`, gebaut am Wirt für
darwin/arm64 und in den Gästen über `lume ssh` gestartet:

- `internal/app/zweimaschinen_vmlauf_test.go` — mDNS und TLS zwischen den
  Gästen, beide Richtungen,
- `internal/fernleitung/fernweg_vmlauf_test.go` — der Fernweg über tailcat,
  gewöhnlich und mit `TS_DEBUG_ALWAYS_USE_DERP=1`.

## Eigenes Netz: mDNS und TLS, beide Richtungen

**Bestanden.** A findet B und B findet A — Kennung, Name und Port 47120
passen, der `.local`-Name ist der Fingerabdruck der Gegenseite und nicht der
Rechnername. Beide wählen die andere Seite über TLS 1.3 an, festgenagelt auf
die gefundene Kennung: Echo da, Kanalbindung 64 Zeichen; und jede Seite hat
den Anruf der anderen bedient.

Für die Veröffentlichung ersetzt: die Rechnernamen der Gäste und die vollständigen
Einladungen; die Gestalt des Protokolls ist die des Laufs.

```text
# Gast A
    zweimaschinen_vmlauf_test.go:65: Rolle a, Rechner "gast-a.local", Kennung bafzaajaiaejcbhji7azmzz4gffocu3doytnsyhvr6vk26yeeywfnuq22ez3jlb6x, Fingerabdruck b73c-b5f1-82a8-c1d9-1c08-56b9-fdb6-faf0
    zweimaschinen_vmlauf_test.go:103: Runde 1: 1 Funde, 0 verworfen, VM-b noch nicht dabei
    zweimaschinen_vmlauf_test.go:105: gefunden: Kennung bafzaajaiaejcarxt73lzx3q2bzvtx2t6cr44lrvqsgrt7a7q7nrijkhbmxkoz34d, Name "VM-b", Wirt 192.168.178.94, Port 47120, Wirtname "7817-6b32-5da2-8ba5-13d0-31b0-4d68-d01a.local"
    zweimaschinen_vmlauf_test.go:141: gewählt: Echo "Gruß von a", Gegenstelle bafzaajaiaejcarxt73lzx3q2bzvtx2t6cr44lrvqsgrt7a7q7nrijkhbmxkoz34d, Kanal 64 Zeichen
    zweimaschinen_vmlauf_test.go:147: bedient: von bafzaajaiaejcarxt73lzx3q2bzvtx2t6cr44lrvqsgrt7a7q7nrijkhbmxkoz34d, Kanal 64 Zeichen
--- PASS: TestZweiMaschinenFindenEinanderUndSprechenGesichert (6.01s)
PASS

# Gast B
    zweimaschinen_vmlauf_test.go:65: Rolle b, Rechner "gast-b.fritz.box", Kennung bafzaajaiaejcarxt73lzx3q2bzvtx2t6cr44lrvqsgrt7a7q7nrijkhbmxkoz34d, Fingerabdruck 7817-6b32-5da2-8ba5-13d0-31b0-4d68-d01a
    zweimaschinen_vmlauf_test.go:105: gefunden: Kennung bafzaajaiaejcbhji7azmzz4gffocu3doytnsyhvr6vk26yeeywfnuq22ez3jlb6x, Name "VM-a", Wirt 192.168.178.90, Port 47120, Wirtname "b73c-b5f1-82a8-c1d9-1c08-56b9-fdb6-faf0.local"
    zweimaschinen_vmlauf_test.go:141: gewählt: Echo "Gruß von b", Gegenstelle bafzaajaiaejcbhji7azmzz4gffocu3doytnsyhvr6vk26yeeywfnuq22ez3jlb6x, Kanal 64 Zeichen
    zweimaschinen_vmlauf_test.go:147: bedient: von bafzaajaiaejcbhji7azmzz4gffocu3doytnsyhvr6vk26yeeywfnuq22ez3jlb6x, Kanal 64 Zeichen
--- PASS: TestZweiMaschinenFindenEinanderUndSprechenGesichert (4.17s)
PASS
```

## Fremdes Netz: der Fernweg über tailcat

**Bestanden, in beiden Fällen.** A horcht und lädt ein, die Einladung trägt
der Wirt zu B, B wählt an und bekommt sein Echo — realer Datentransfer über
die Netzstapel zweier Maschinen. Gewöhnlich nach 353 ms, mit erzwungenem
DERP nach 2,3 s. **Der Weg ist nicht messbar**: tailcat nennt bei offener
Verbindung keinen (SEC-21), und der Lauf sagt das, statt „direkt" oder
„Relais" zu raten.

```text
# gewoehnlich, Gast A
    fernweg_vmlauf_test.go:51: Rolle a, TS_DEBUG_ALWAYS_USE_DERP=""
    fernweg_vmlauf_test.go:56: Einladung abgelegt: mse4mac:1:… (gekürzt: die Einladung trägt Schlüsselmaterial)
    fernweg_vmlauf_test.go:56: bedient: "Mausbraten von b" von bafzaajaiaejcaaem6l3k5rv3cq4ygt5mhju7t6scok7w44uf7i5dy23bndj4eolk, 1 angebunden
    fernweg_vmlauf_test.go:56: Weg: nicht messbar — tailcat nennt bei offener Verbindung keinen (SEC-21)
--- PASS: TestDerFernwegZwischenZweiMaschinen (13.71s)
PASS
# gewoehnlich, Gast B
    fernweg_vmlauf_test.go:51: Rolle b, TS_DEBUG_ALWAYS_USE_DERP=""
    fernweg_vmlauf_test.go:61: Echo nach 353ms: "zurück: Mausbraten von b" (25 Bytes), Gegenstelle bafzaajaiaejcbe7d5mv5lm4e74kuhpjnilfwyj72z3ubd7f6jmtt2rubvwpyzhsa
--- PASS: TestDerFernwegZwischenZweiMaschinen (0.36s)
PASS
# derp, Gast A
    fernweg_vmlauf_test.go:51: Rolle a, TS_DEBUG_ALWAYS_USE_DERP="1"
    fernweg_vmlauf_test.go:56: Einladung abgelegt: mse4mac:1:… (gekürzt: die Einladung trägt Schlüsselmaterial)
    fernweg_vmlauf_test.go:56: bedient: "Mausbraten von b" von bafzaajaiaejcabe7hpxzh4r4dhon5gcvcik7gunzik6sqpfkzdzc3naohxrnql3t, 1 angebunden
    fernweg_vmlauf_test.go:56: Weg: nicht messbar — tailcat nennt bei offener Verbindung keinen (SEC-21)
--- PASS: TestDerFernwegZwischenZweiMaschinen (15.65s)
PASS
# derp, Gast B
    fernweg_vmlauf_test.go:51: Rolle b, TS_DEBUG_ALWAYS_USE_DERP="1"
    fernweg_vmlauf_test.go:61: Echo nach 2.318s: "zurück: Mausbraten von b" (25 Bytes), Gegenstelle bafzaajaiaejcb7lrmp5tqtamvaqbipghcbr4jnx6q4cdypyl2ws3qne25wvy2xjx
--- PASS: TestDerFernwegZwischenZweiMaschinen (2.32s)
PASS
```

## Seit tailcat v0.7.0: der Weg, und alte gegen neue Fassung — K01

**Am 2026-10-07 nachts**, zwei frische Gäste aus `macos-sequoia-vanilla`
(macOS 15.2, A `.96`, B `.95`), die Einladung über SSH von A nach B
getragen — der geteilte Ordner zeigte B eine neue Datei nicht. Der Lauf
verlangt jetzt den Weg (K01, #2187).

| Fall | A (horcht) | B (wählt) | Echo | Weg an A |
| --- | --- | --- | --- | --- |
| gewöhnlich | v0.7.0 | v0.7.0 | 119 ms | `Direkt 192.168.178.95:53656`, Relais `fra` — PASS |
| erzwungenes DERP | v0.7.0 | v0.7.0 | 3,16 s | nur Relais `fra` — PASS |
| **alt gegen neu** | v0.7.0 | **v0.6.0** | 95 ms | `Direkt [2003:…]:57865`, Relais `fra` — PASS |
| **neu gegen alt** | **v0.6.0** | v0.7.0 | 94 ms | „nicht messbar" — die alte Fassung nennt keinen, wie bisher — PASS |

**Damit ist gemessen, was [0114](../entscheidungen/0114-tailcat-gepinnt.md)
verlangt**: eine Einladung derselben Fassung trägt zwischen tailcat v0.6.0
und v0.7.0 in beide Richtungen; die Einladungsfassung bleibt bei 1.

## Was dieser Beleg nicht sagt

- **Die Erlaubnis *Lokales Netzwerk* der Anwendung im Gast.** Gestartet ist
  über SSH; macOS fragte dabei nichts, und die Pakete gingen. Der Weg der
  Anwendung selbst — über LaunchServices, mit der Rückfrage in der
  Oberfläche des Gasts — ist damit nicht gefahren. An ihm endete der Audit
  mit `no route` trotz Erlaubnis.
- **Zwei physische Macs in zwei Netzen** hinter verschiedenen Routern —
  außerhalb des Auftrags (Ausgangslage §9).
- **Den Weg**, den tailcat wählt — er wird nicht genannt.

**Aufgeräumt**: die Testprogramme und die Einladungsdatei in beiden Gästen
gelöscht; danach sind die Gäste auf Wunsch des Auftraggebers entfernt.
