# Hüllenprobe: bk07a

**bestanden** — Commit `1c0b74221`,
2026-10-09 19:51 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**1 — regulär: ein Horchposten läuft, eigene Größe, ⌘Q**

- ✓ eine Identität steht
- ✓ der Horchposten läuft — 127.0.0.1:54195
- ✓ ⌘Q beendet
- bis der Prozess weg war: 27 ms
- ✓ in der Frist des Schlusses (5 s) und einer Sekunde Spiel
- gemerkt: 1300 × 841
- ✓ die Größe steht im Bestand

**2 — nach dem Neustart gilt die gemerkte Größe**

- das Fenster ist 1301 pt breit
- ✓ so breit wie gemerkt

**3 — mit gesperrtem Bestand: das Merken geht verloren, das Programm endet trotzdem**

- ✓ ⌘Q beendet
- bis der Prozess weg war: 1037 ms
- ✓ in der Frist des Schlusses (5 s) und einer Sekunde Spiel
- gemerkt: 1300 × 841
- ✓ die alte Größe blieb — ein Bestenfalls-Verlust
- im Protokoll: time=2026-10-09T21:51:36.200+02:00 level=WARN msg="der Schluss kam nicht in seiner Frist zu Ende" grund="was beim Schließen gemerkt werden sollte"
- ✓ das Protokoll nennt den Verlust
- ✓ die eigenen Läufe kamen trotzdem zu Ende
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
