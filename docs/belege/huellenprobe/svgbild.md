# Hüllenprobe: svgbild

**bestanden** — Commit `1c0b74221`,
2026-10-09 22:11 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**Gegenprobe A — die Seite erreicht den Dienst**

- ✓ der Dienst hört zu — ["/gegenprobe"]

**Das SVG über <img>, wie die Dateiansicht es zeigt**

- ✓ das SVG ist gezeichnet — Breite 40
- ✓ kein Skript, kein onload, kein @import, kein <image> von außen — []

**Gegenprobe B — dasselbe SVG als Dokument lädt**

- ✓ als Dokument fragt dasselbe SVG an — ["/gegenprobe","/stil","/skript","/bild","/onload"]
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
