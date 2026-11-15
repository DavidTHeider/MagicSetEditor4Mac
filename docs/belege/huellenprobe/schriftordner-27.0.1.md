# Beleg: #2411 — Schriften in Unterordnern von ~/Library/Fonts, macOS 27.0.1

**Stand:** Commit `b01d125d4` · **Datum:** 2026-10-15 10:09 UTC · **Befehl:** `node scripts/schriftordnergast.mjs mse090-r02` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse090-r02` — macOS 27.0.1 (26A434).

**Ergebnis:** **macOS aktiviert eine Schrift im Unterordner**; die Gegenprobe flach gelang.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Schriftorte | VanillaWhale-Regular: vorher nicht da | gelungen | 0 | nicht da |
| Schriftorte | VanillaWhale-Regular: im Unterordner | gelungen | 0 | macOS aktiviert sie |
| Schriftorte | MagicMedieval: vorher nicht da | gelungen | 0 | nicht da |
| Schriftorte | MagicMedieval: flach (Gegenprobe) | gelungen | 0 | macOS aktiviert sie |

**Gefragt** wurde AppKit (`NSFontManager.availableFonts`) über JXA, bis zu 20 s je Lage. **Nicht gefragt**: Font Book und CoreText unmittelbar; eine Schrift, die AppKit nicht nennt, steht in keinem Schriftmenü.
