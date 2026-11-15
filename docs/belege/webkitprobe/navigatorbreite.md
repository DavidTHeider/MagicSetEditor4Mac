# WebKit-Probe — navigatorbreite.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:26Z |

- ✓ Z7 `motor wk2`
- ✓ Z8 `lade /`
- ✓ Z9 `frist 20000`
- ✓ Z10 `warte nav`
- ✓ Z11 `wartetext Erste Schritte`
- ✓ Z12 `taste escape`
- ✓ Z13 `taste cmd+3`
- ✓ Z14 `wartetext Sammlung`
- ✓ Z15 `warte aside`
- ✓ Z16 `zeige [...document.querySelectorAll('aside input, aside select, aside textarea')].map(e => { const a = e.closest('aside').getBoundingClientRect(); const r = e.getBoundingClientRect(); return (e.placeholder || e.name || e.tagName) + ': ' + Math.round(r.right) + ' / ' + Math.round(a.right) })`
  ```
  ["Name, Set, Kennung oder Notiz: 540 / 553","SELECT: 483 / 553"]
  ```
- ✓ Z17 `pruefe [...document.querySelectorAll('aside input, aside select, aside textarea')].every(e => e.getBoundingClientRect().right <= e.closest('aside').getBoundingClientRect().right + 0.5)`
- ✓ Z18 `abzug sammlung-navigator`
  ```
  Bild: sammlung-navigator.png
  ```

**Ergebnis: bestanden** — 12 Schritte.
