# Hüllenprobe: syncpanik09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:26 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**1 — eine Panik im Rückruf eines InvokeSync, drei Sekunden nach dem Start**

- time=2026-10-20T13:26:19.601+02:00 level=INFO msg="der synchrone Ruf kam nach der Panik zurück"
- ✓ der Wartende kam nach der Panik frei — time=2026-10-20T13:26:19.601+02:00 level=INFO msg="der synchrone Ruf kam nach der Panik zurück"

**2 — die Anwendung läuft weiter, und das Protokoll nennt die Panik**

- time=2026-10-20T13:26:19.600+02:00 level=ERROR msg="Wails fing eine Panik" panik="Probepanik im Hauptfaden" stelle=github.«geschwärzt».probeSyncpanik.func1.1 kennung=2a3e5259
- ✓ die Panik steht mit Vorgangskennung da — time=2026-10-20T13:26:19.600+02:00 level=ERROR msg="Wails fing eine Panik" panik="Probepanik im Hauptfaden" stelle=github.«geschwärzt».probeSyncpanik.func1.1 kennung=2a3e5259
- ✓ kein Ende ist angesagt
- ✓ die Seite antwortet nach der Panik
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/syncpanik09/nach-der-panik.png`

**3 — ⌘Q beendet, und der Folgestart nennt keinen Abbruch**

- ✓ ⌘Q beendet
- ✓ kein Hinweis auf einen Abbruch
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
