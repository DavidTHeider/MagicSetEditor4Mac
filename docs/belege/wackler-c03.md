# Wackler aus vollen nativen Läufen — C03

**C03** (#2228):
drei vollständige Läufe der Hüllenprobe am **identischen Bündel**, jeder rote
Fall einer Ursache zugeordnet — Produkt, Vorrichtung oder Umgebung.

## Die Läufe

Alle am Bauabdruck `4232d51b3deb`, Mac17,3, macOS 27.0.1, `MSE_DAUER_MIN=20`
(der Dauerbetrieb auf 20 Minuten gekürzt, siehe unten). Zwischen den Commits
der Läufe ändern sich nur Skripte außerhalb der Baueingänge
(`scripts/baustand.mjs`, `baueingaenge`): die Fuzzkampagne und, ab Lauf 4,
die Bildschirmerkennung. **Die Commits stammen aus dem Stand vor dem
Neuaufsetzen von Phase C auf das gemergte Phase B**; das Bündel des
Endstands trägt deren Nachlaufkorrekturen und hat einen anderen Bauabdruck.
Die Wacklerfrage gilt dem Bündel, an dem gemessen ist.

| Lauf | Beginn | Ende | HEAD | Bauabdruck | Fälle | rot | Bilder fehlen |
|---|---|---|---|---|---|---|---|
| `2026-10-10T01-43-50Z-46468` | 2026-10-10T01:43:50.917Z | 2026-10-10T02:35:34.697Z | `05e7fc019` | `4232d51b3deb` | 107 | 2 | 0 |
| `2026-10-10T03-50-42Z-69067` | 2026-10-10T03:50:42.360Z | 2026-10-10T04:41:49.352Z | `0326d3d7e` | `4232d51b3deb` | 107 | 20 | 330 |
| `2026-10-10T04-45-17Z-7002` | 2026-10-10T04:45:17.801Z | 2026-10-10T05:34:49.252Z | `0326d3d7e` | `4232d51b3deb` | 107 | 2 | 0 |
| `2026-10-10T05-38-54Z-51548` | 2026-10-10T05:38:54.546Z | 2026-10-10T06:28:03.407Z | `24c0a07bb` | `4232d51b3deb` | 107 | 2 | 0 |

**Ein fünfter Anlauf zählt nicht und ist nicht ausgewertet**: der um 03:31 UTC
(`2026-10-10T03-31-02Z-62448`) lief neben einem versehentlich gestarteten
Fuzzlauf — ein `node -e "import(…fuzzkampagne.mjs)"` führte das Skript aus,
statt es nur zu laden. Er wurde abgebrochen, der Fuzzlauf beendet und sein
Korpus entfernt, bevor Lauf 2 begann.

**Gezählt werden Lauf 1, 3 und 4** — Lauf 4 schon mit der Bildschirmerkennung, die das Energieprotokoll las und keine Ausphase fand. Lauf 2 lief bis auf seine letzten
zehn Minuten bei ausgeschaltetem Bildschirm und ist kein Lauf der Anwendung,
sondern einer der Umgebung — unten, wie das belegt ist.

## Je rotes Drehbuch

Nur Drehbücher, die in mindestens einem Lauf rot waren; die übrigen 87 sind in allen vier grün. L2 ist der Lauf bei ausgeschaltetem Bildschirm.

| Drehbuch | L1 | L2 | L3 | L4 |
|---|---|---|---|---|
| ai02 | ✗ 1 | ✗ 1 | ✗ 1 | ✗ 1 |
| bildrate | ✓ | ✗ 1 | ✓ | ✓ |
| bk07a | ✓ | ✗ 1 | ✓ | ✓ |
| budgets | ✓ | ✗ 1 | ✓ | ✓ |
| c05fremd | ✓ | ✗ 2 | ✓ | ✓ |
| dauerbetrieb | ✗ 1 | ✗ 1 | ✗ 1 | ✗ 1 |
| dom03 | ✓ | ✗ 2 | ✓ | ✓ |
| dom03ak02 | ✓ | ✗ 2 | ✓ | ✓ |
| eintraege | ✓ | ✗ 2 | ✓ | ✓ |
| fix-f02 | ✓ | ✗ 2 | ✓ | ✓ |
| fmt02f | ✓ | ✗ 2 | ✓ | ✓ |
| fp02 | ✓ | ✗ 2 | ✓ | ✓ |
| leiste | ✓ | ✗ 2 | ✓ | ✓ |
| loc01 | ✓ | ✗ 1 | ✓ | ✓ |
| nat04handel | ✓ | ✗ 4 | ✓ | ✓ |
| phasec | ✓ | ✗ 2 | ✓ | ✓ |
| readmebilder | ✓ | ✗ 64 | ✓ | ✓ |
| start | ✓ | ✗ 1 | ✓ | ✓ |
| u02b | ✓ | ✗ 2 | ✓ | ✓ |
| u13 | ✓ | ✗ 2 | ✓ | ✓ |

**In den gezählten Läufen 1, 3 und 4 ist kein Drehbuch einmal grün und einmal rot.** Rot sind in allen dreien dieselben zwei, aus denselben Gründen:

| Drehbuch | Ursache | Klasse | Beleg |
| --- | --- | --- | --- |
| `ai02` | Ollama läuft nicht, kein Textmodell auf dem Rechner (`ECONNREFUSED` an 127.0.0.1:11434). Die Modelle von AI-02 sind gelöscht, ein neuer Download ist anzukündigen | **Plattformgrenze** — die ausdrücklich getrennte KI-Grenze aus der Definition of Ready | jeder Lauf, dieselbe Zeile |
| `dauerbetrieb` | In 20 Minuten gibt es nach der Aufwärmzeit keinen ganzen Abschnitt; die Auswertung sagt „nur 0 vollständige Abschnitte" und urteilt nicht | **Vorrichtung, eingestellt** — die Kürzung ist gewollt, damit drei Läufe in einen Vormittag passen. Der Speicheranstieg selbst ist ein **Produktbefund** und steht in E06 (#2278), gemessen über zwei Stunden | jeder Lauf |

## Lauf 2: der Bildschirm war aus

Das Energieprotokoll (`pmset -g log`) nennt „Display is turned off" um
**02:48:35 UTC** und „Display is turned on" um **04:31:33 UTC**. Lauf 2 lief
von 03:50:42 bis 04:41:49 UTC. Die Warteschleife vor dem Lauf lief ohne
`caffeinate`, und `caffeinate -d` hält einen Bildschirm nur an — es schaltet
keinen ein.

- **20 rote Drehbücher, 97 Befunde, 330 fehlende Bilder.** Fast alle Befunde
  sind abgelaufene Fristen: `js_eval: keine Antwort nach 30000 ms`, „die
  Vorschau ist gezeichnet (nach 20000 ms) — null", „Liste und Inspector zeigen
  drei (nach 15000 ms) — false". Eine Webansicht, die niemand sieht, zeichnet
  nicht.
- **Das letzte Rot steht um 04:29 UTC im Beleg** (`u13`), der Bildschirm ging
  um 04:31:33 an, und **jedes Drehbuch, dessen Beleg ab 04:31 geschrieben ist, ist grün** — 21 von 21 (gezählt an der Minute im Kopf des Belegs).
- Lauf 1 davor, Lauf 3 und Lauf 4 danach: dieselben 18 Drehbücher grün.

**Was daraus wurde** (Vorrichtung): `huellenprobe.sh` hält den Bildschirm mit
`caffeinate -dimsu` wach — `-u` schaltet ihn ein —, und schaltet ihn trotzdem
jemand aus, liest der Lauf das Energieprotokoll und vermerkt die Ausphase im
Kopf jedes betroffenen Belegs und im Manifest (`scripts/bildschirmzeit.mjs`).
**Das Urteil bleibt**: ob die Anwendung bei hellem Bildschirm bestanden
hätte, weiß niemand. Gegengeprüft am echten Protokoll: der Zeitraum von
Lauf 2 trifft die Ausphase, der von Lauf 1 nicht.

## Was das nicht belegt

- **Wackler unter Last.** Gemessen ist ein ruhiger Rechner (Last unter 3 beim
  Start); ein Drehbuch, das erst unter Last wackelt, sieht man hier nicht.
- **Den Dauerbetrieb über zwei Stunden** — das ist E06.
- **Einen zweiten Rechner** — ein Mac, ein Bündel.

## Wo die Rohdaten liegen

Je Lauf unter `.huellenprobe/laeufe/<kennung>/`: `manifest.json` (HEAD,
offene Dateien, Bauabdruck, Programm- und Oberflächenabdrücke, Umwelt, je
Drehbuch Befunde, fehlende Bilder, Bindungen und — ab Lauf 4 — Beginn, Ende
und Ausphasen), die Belege und die Rufbücher. Das Verzeichnis ist nicht im
Repositorium, und es bleiben die jüngsten zehn Läufe. Die Tabellen oben sind
aus den vier Manifesten erzeugt.
