# Hüllenprobe: panik09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:21 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**Menü, Finder, Ablegen und Schluss werfen**

- time=2026-10-20T13:21:09.398+02:00 level=ERROR msg="ein Rückruf des Fensters brach mit einer Panik ab" art="ein Befehl aus dem Menü" panik="Probepanik in ein Befehl aus dem Menü" kennung=d4d87eef
- ✓ ein Befehl aus dem Menü: die Panik steht mit Vorgangskennung im Protokoll
- ✓ ein Befehl aus dem Menü: die Anwendung antwortet danach
- time=2026-10-20T13:21:09.483+02:00 level=ERROR msg="ein Rückruf des Fensters brach mit einer Panik ab" art="das Öffnen aus dem Finder" panik="Probepanik in das Öffnen aus dem Finder" kennung=6863667b
- ✓ das Öffnen aus dem Finder: die Panik steht mit Vorgangskennung im Protokoll
- ✓ das Öffnen aus dem Finder: die Anwendung antwortet danach
- time=2026-10-20T13:21:09.667+02:00 level=ERROR msg="ein Rückruf des Fensters brach mit einer Panik ab" art="das Ablegen" panik="Probepanik in das Ablegen" kennung=e0b9c46f
- ✓ das Ablegen: die Panik steht mit Vorgangskennung im Protokoll
- ✓ das Ablegen: die Anwendung antwortet danach
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/panik09/f02-nach-drei-paniken.png`
- ✓ ⌘Q beendet trotz Panik im Schluss
- bis der Prozess weg war: 22 ms
- ✓ in der Frist des Schlusses (5 s) und einer Sekunde Spiel
- time=2026-10-20T13:21:10.068+02:00 level=ERROR msg="ein Rückruf des Fensters brach mit einer Panik ab" art="der Schluss" panik="Probepanik in der Schluss" kennung=d42c7b50
- ✓ der Schluss: die Panik steht mit Vorgangskennung im Protokoll
- ✓ der Bestand ist derselbe und lesbar
- ✓ der Schluss räumte trotzdem ab: keine Marke liegt

**Der Start wirft**

- time=2026-10-20T13:21:10.378+02:00 level=ERROR msg="ein Rückruf des Fensters brach mit einer Panik ab" art="der Start" panik="Probepanik in der Start" kennung=c859ae22
- ✓ der Start: die Panik steht mit Vorgangskennung im Protokoll
- ✓ der Start: die Anwendung antwortet danach
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/panik09/f02-nach-panik-im-start.png`
- ✓ ⌘Q beendet
- ✓ der Bestand ist auch danach derselbe
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
