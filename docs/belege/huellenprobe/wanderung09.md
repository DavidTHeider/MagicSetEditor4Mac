# Hüllenprobe: wanderung09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:31 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**0.8.0 aus dem Tag legt den Bestand an**

- ✓ 0.8.0 führt 97 Wanderungen und 5.000 Karten

**mse mit dem Halt des Abnahmebaus**


**SIGKILL mitten in der Wanderung**

- ✓ vor dem Halt entstand die Sicherung — bestand-vor-0.8.0.db
- ✓ der Prozess endete am Signal — {"code":null,"signal":"SIGKILL"}
- ✓ der Bestand ist heil
- ✓ er führt die 97 Wanderungen von vorher — keine halbe
- ✓ jede Tafel trägt den Inhalt von vorher
- ✓ die neue Tafel gibt es nicht

**Der nächste Start wandert zu Ende**

- ✓ jetzt 101 Wanderungen — 0101_gegenbelegbefunde.sql 1c671c3af30241ccc225c4b8dd5ebfdb4a527f3de0c3bc4ce5a9f13e33032535
- ✓ mit demselben Inhalt
- ✓ und denselben Bildern
- ✓ die Sicherung von vorher wurde wieder genommen, keine zweite — bestand-vor-0.8.0.db

**Die Platte ist voll**

- Bestand 4064 KB, frei 2032 KB
- den Bestand vor der Migration sichern: geschrieben werden konnte nicht, und es ist kaum Platz frei: schaffe Platz und versuche es noch einmal; hilft das nicht, prüfe den Datenträger und das Schreibrecht
- ✓ der Start scheitert und sagt, dass kaum Platz frei ist und der Datenträger zu prüfen — den Bestand vor der Migration sichern: geschrieben werden konnte nicht, und es ist kaum Platz frei: schaffe Platz und versuche es noch einmal; hilft das nicht, prüfe den Datenträger und das Schreibrecht
- ✓ ohne den Text des Treibers
- ✓ der Bestand ist bitgleich
- ✓ keine Sicherung bleibt liegen, auch keine halbe

**Platz frei, und noch einmal**

- ✓ jetzt 101 Wanderungen — 0101_gegenbelegbefunde.sql 1c671c3af30241ccc225c4b8dd5ebfdb4a527f3de0c3bc4ce5a9f13e33032535
- ✓ mit demselben Inhalt
- ✓ und genau einer Sicherung davor
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
