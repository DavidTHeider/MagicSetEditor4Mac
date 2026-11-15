# Hüllenprobe: budgets

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:40 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.

- Last vor dem Lauf 2.73 / 3.13 / 3.51 auf 10 Kernen

**Start bis bedienbar**

- ✓ Kaltstart 1: ⌘Q beendet
- ✓ Kaltstart 2: ⌘Q beendet
- ✓ Kaltstart 3: ⌘Q beendet
- ✓ Kaltstart 4: ⌘Q beendet
- ✓ Kaltstart 5: ⌘Q beendet
- ✓ Start 1: ⌘Q beendet
- ✓ Start 2: ⌘Q beendet
- ✓ Start 3: ⌘Q beendet
- ✓ Start 4: ⌘Q beendet
- ✓ Start 5: ⌘Q beendet
- ✓ Start 6: ⌘Q beendet

**Feld → dekodiertes Vorschaubild**

- 0 Tasten kamen nicht im Feld an und zählen nicht

**Seitenwechsel über die Befehlspalette**

- ✓ Start, Vorschau und Seitenwechsel: ⌘Q beendet

**Rollen durch »Black« (3.788 Karten)**

- das große Set heißt »Black.mse-set«
- ✓ Black: ⌘Q beendet

**Ergebnis**

- ✓ jede Vorschau kam an — 0 ohne Bild: 
- ✓ jeder Seitenwechsel kam an
- ✓ die Kartenliste von Black wurde gerollt

| Reihe | n | min | Median | P95 | max | Vorschlag des Plans |
| --- | --- | --- | --- | --- | --- | --- |
| Start kalt (ms) | 5 | 491 | 494 | — | 510 | ≤ 2000 |
| Start warm (ms) | 6 | 423 | 506 | — | 521 | ≤ 2000 |
| Feld → Vorschaubild (ms) | 24 | 22 | 36 | 47 | 49 | Median ≤ 150 |
| Seitenwechsel (ms) | 24 | 34 | 58 | 102 | 108 | — |
| Rollbild Black (ms) | 159 | 3 | 21 | 49 | 80 | ≤ 50 |

Speicher (MB, RSS): nach dem Start {"anwendung":163,"web":135,"gpu":83}, nach den Vorschauen {"anwendung":189,"web":234,"gpu":145}.

Rohwerte — kalt [510,494,496,491,492], warm [521,486,423,507,459,506], Vorschau [40,49,28,47,39,36,38,39,34,36,38,35,41,22,31,31,32,32,36,34,36,32,31,31], Wechsel [70,58,80,55,45,102,63,54,69,53,37,81,81,54,53,54,34,108,71,54,55,69,37,86], Rollen [6,17,44,3,47,60,27,22,37,49,51,50,33,35,34,36,43,43,37,36,80,18,49,42,30,34,31,32,31,32,34,33,37,37,42,43,58,34,54,18,22,21,20,21,20,24,21,20,26,22,21,29,32,17,16,18,17,17,16,18,30,20,20,21,18,19,21,19,23,18,19,16,17,17,18,17,22,24,21,20,28,23,24,22,21,33,36,27,21,26,21,23,24,23,32,28,24,28,30,20,19,25,30,30,35,38,19,22,27,21,44,17,16,17,33,15,18,17,34,17,15,34,17,17,15,17,18,16,17,17,17,19,17,18,18,19,18,18,23,23,18,19,20,22,21,20,24,20,20,21,24,47,17,17,17,33,16,17,18].
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht
