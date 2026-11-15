# Hüllenprobe: wiederanlauf09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:31 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**Zweimal sauber: kein Hinweis**

- ✓ während der Sitzung liegt eine Marke — absturz-1792495884983107000-32074.txt, fenster-32074.json, fenster-32074.sperre, sitzungen.sperre
- ✓ ⌘Q beendet
- ✓ nach dem sauberen Schluss liegt keine Marke — sitzungen.sperre
- ✓ nach einem sauberen Schluss: kein Hinweis
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/wiederanlauf09/f04-sauber-ohne-hinweis.png`
- ✓ ⌘Q beendet

**Absturz: eine Panik, die niemand fängt**

- ✓ der Prozess endet von selbst
- ✓ die Marke bleibt liegen — absturz-1792495887751022000-32126.txt, fenster-32126.json, fenster-32126.sperre, sitzungen.sperre
- panic: Probeabsturz nach 3s ⏎  ⏎ goroutine 697 [running]:
- ✓ die Absturzspur nennt die Panik

**Folgestart: Hinweis mit Spur, Weg in die Diagnose**

- Die letzte Sitzung, begonnen am 20.10.2026, 13:31, endete unerwartet. Eine Absturzspur liegt vor; das Diagnosepaket nimmt sie mit, wenn du es anlegst. Diagnose öffnen Ausblenden
- ✓ der Hinweis nennt den Abbruch
- ✓ der Hinweis sagt, dass eine Spur vorliegt
- ✓ die gelesene Marke ist weg — absturz-1792495887751022000-32126.txt, absturz-1792495891027904000-32147.txt, fenster-32147.json, fenster-32147.sperre, sitzungen.sperre
- ✓ nur die laufende Sitzung hält eine Sperre — absturz-1792495887751022000-32126.txt, absturz-1792495891027904000-32147.txt, fenster-32147.json, fenster-32147.sperre, sitzungen.sperre
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/wiederanlauf09/f04-hinweis-mit-spur-hell.png`
- ✓ »Diagnose öffnen« nimmt den Fokus
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/wiederanlauf09/f04-diagnose-offen.png`
- ✓ »Zusammenstellen« im Abschnitt »Diagnosepaket« nimmt den Fokus
- ✓ »Paket sichern« im Abschnitt »Diagnosepaket« nimmt den Fokus
- ✓ das Diagnosepaket trägt die Absturzspur
- ✓ der freie Text der Panik bleibt draußen
- ✓ ⌘Q beendet

**Wieder sauber: kein Hinweis mehr**

- ✓ ein gelesener Abbruch kommt nicht wieder

**Abschuss: SIGKILL**

- ✓ der Prozess ist weg
- ✓ die Marke bleibt liegen — absturz-1792495887751022000-32126.txt, absturz-1792495893167808000-32200.txt, fenster-32200.json, fenster-32200.sperre, sitzungen.sperre

**Folgestart dunkel: Hinweis ohne Spur**

- ✓ das Fenster ist dunkel
- Die letzte Sitzung, begonnen am 20.10.2026, 13:31, endete unerwartet. Eine Absturzspur gibt es nicht — wahrscheinlich wurde die Anwendung von außen beendet. Diagnose öffnen Ausblenden
- ✓ der Hinweis nennt den Abbruch
- ✓ der Hinweis sagt, dass es keine Spur gibt
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/wiederanlauf09/f04-hinweis-ohne-spur-dunkel.png`
- ✓ »Ausblenden« nimmt den Fokus
- ✓ ⌘Q beendet

**Zum Schluss sauber: kein Hinweis**

- ✓ nach dem Ausblenden und einem sauberen Schluss: kein Hinweis
- ✓ ⌘Q beendet
- ✓ höchstens eine Spur mit Inhalt liegt — absturz-1792495887751022000-32126.txt
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
