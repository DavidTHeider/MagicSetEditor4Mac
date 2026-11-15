# Hüllenprobe: aufstieg09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 10:20 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**0.8.0 aus dem Tag: Kommandozeile und Abnahmebündel**

- 0.8.0: Stand 52e1ee3e8, Bündelabdruck 365d0c853daf, fassung	0.8.0
- 50 Sets eingelesen in 4 s
- ✓ 0.8.0 führt 50 Sets — 50
- ✓ und 25.000 Karten — 25000
- ✓ auf der Grundlinie: 97 Wanderungen — 0097_faktorhaushalt.sql 5e1f77d437b5c3153629c91e23d16b1a3a4803a93e209e85066669c3f72d5d94
- ✓ jeder Bildverweis findet seine Datei — 1000 von 1000

**Der Aufstieg: das heutige mse öffnet den Bestand**

- Aufstieg in 316 ms
- ✓ die 4 Wanderungen nach der Grundlinie kommen dazu — 0101_gegenbelegbefunde.sql 1c671c3af30241ccc225c4b8dd5ebfdb4a527f3de0c3bc4ce5a9f13e33032535
- ✓ jede Tafel trägt denselben Inhalt
- ✓ jedes Set hat seinen Änderungszähler
- ✓ die Bilder sind dieselben — 1000 von 1000
- ✓ die Tafel der Vermerke ist leer — unbekannt heißt unbekannt
- ✓ vor der Wanderung entstand genau eine Sicherung — vorher keine, neu bestand-vor-0.8.0.db
- ✓ sie hält den Stand von 0.8.0
- ✓ und sie ist heil

**Die alte Kommandozeile weist den neuen Stand ab**

- der Bestand der Migrationen ist nicht in Ordnung: dieser Bestand führt 0098_sicherungsvermerke.sql, 0099_setstand.sql, 0100_setstand_genauer.sql, 0101_gegenbelegbefunde.sql, und diese Fassung kennt nur bis 0097_faktorhaushalt.sql — er stammt aus einer neueren Fassung von MagicSetEditor4Mac und wird von dieser nicht geöffnet, damit kein halber Stand entsteht
- ✓ 0.8.0 öffnet ihn nicht und sagt, dass er neuer ist
- ✓ und lässt ihn, wie er ist
- **Befund**: die Kommandozeile von 0.8.0 spielt nicht zurück — der Bestand der Migrationen ist nicht in Ordnung: dieser Bestand führt 0098_sicherungsvermerke.sql, 0099_setstand.sql, 0100_setstand_genauer.sql, 0101_gegenbelegbefunde.sql, und diese Fassung kennt nur bis 0097_faktorhaushalt.sql — er stammt aus einer neueren Fassung von MagicSetEditor4Mac und wird von dieser nicht geöffnet, damit kein halber Stand entsteht
- ✓ auch dabei bleibt der Bestand, wie er ist

**Der Rückweg durch das Fenster von 0.8.0**

- ✓ 0.8.0 zeigt die Wiederherstellung — Der Bestand ließ sich nicht öffnen ⏎  ⏎ Die Anwendung arbeitet erst wieder, wenn eine brauchbare Bestandsdatei dasteht.
- ✓ und nennt den Weg zurück über die Sicherung von vor der Wanderung
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/aufstieg09/m01-wiederherstellung-080.png`
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/aufstieg09/m01-zurueckgespielt-080.png`
- ✓ mit allen Karten
- ✓ und den 97 Wanderungen von 0.8.0
- ✓ und denselben Bildern
- ✓ ⌘Q beendet 0.8.0

**Dieselbe Buchführung: frisch, aufgestiegen, zwei zugleich**

- ✓ beide Starter kamen durch — 0/0
- ✓ frisch und aufgestiegen führen dieselbe Buchführung
- ✓ jede Wanderung steht genau einmal da
- ✓ zwei gleichzeitige Starter legen eine Sicherung an — bestand-vor-0.8.0.db
- ✓ und sie hält die 97 Wanderungen von 0.8.0
- ✓ und der Inhalt ist der von 0.8.0
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
