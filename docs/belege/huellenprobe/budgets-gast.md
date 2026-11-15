# Beleg: B02 im Gast — Leistungsbudgets in der VM

**Commit:** `0138f053c` · **Datum:** 2026-10-09 13:30 UTC · **Befehl:** `node scripts/budgetsgast.mjs mse080-152` · **Rückgabe:** 0

**Maschine:** Lume-VM `mse080-152` — macOS 15.2 (24C101), VirtualMac2,1, 4 Kerne, 4096 MB; Last im Gast nach dem Start { 33.53 7.11 2.52 }, beim Messen { 1.47 5.10 3.70 }. Wirt: Apple M5, macOS 27.0.1.

**Bündel:** Abnahmebau, Baustand `d9fb877d0f09`, Programm `f0b8b2b912aa0c69…`.

Gemessen mit **denselben JavaScript-Stücken** wie am Wirt (`scripts/huellenprobe/budgets.mjs`, übergeben als `messungen.json`); getrieben von `scripts/huellenprobe/gast/budgets.pl`, weil der Gast kein Node hat. **Eine VM hat keine GPU des Wirts** — die Zahlen sind eine Obergrenze, kein zweiter Rechner.

| Reihe | n | min | Median | P95 | max | Vorschlag des Plans |
| --- | --- | --- | --- | --- | --- | --- |
| Start kalt (ms) | 5 | 1489.8 | 1531.7 | — | 2002.3 | ≤ 2000 |
| Start warm (ms) | 6 | 1494.9 | 1528.9 | — | 1621.9 | ≤ 2000 |
| Feld → Vorschaubild (ms) | 0 | — | — | — | — | Median ≤ 150 — **nicht gemessen** |
| Seitenwechsel (ms) | 0 | — | — | — | — | — — **nicht gemessen** |
| Rollbild Black (ms) | 0 | — | — | — | — | ≤ 50 — **nicht gemessen** |

- 0 Tasten kamen nicht im Feld an und zählen nicht; 0 Vorschauen ohne Bild; 0 Seitenwechsel ohne Ankunft.
- Speicher (MB, RSS): nach dem Start {"anwendung":126,"gpu":18,"web":107}.
- Die Seite zeichnet: 0 Bilder in 2200 ms (`requestAnimationFrame`).

**Grenzen dieses Laufs:**

- die Seite zeichnet nicht (0 Bilder in 2 s, Sichtbarkeit »hidden«) — Vorschau, Seitenwechsel und Rollen bleiben ungemessen
- Warum: der Gast steht hinter seinem Sperrbildschirm ([Bild](budgets-gast/sperrbildschirm.jpg)); entsperren hieße ein Kennwort eingeben, und das geschieht nicht (`AGENTS.md` §17).

## Rohwerte

```json
{"kalt":[2002.2900390625,1489.828125,1547.02319335938,1509.712890625,1531.69018554688],"warm":[1516.02392578125,1554.19018554688,1511.041015625,1528.94091796875,1494.89599609375,1621.89477539062]}
```
