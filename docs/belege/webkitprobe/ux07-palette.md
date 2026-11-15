# WebKit-Probe — ux07-palette.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:17Z |

- ✓ Z6 `motor wk2`
- ✓ Z7 `lade /`
- ✓ Z8 `frist 20000`
- ✓ Z9 `warte nav`
- ✓ Z10 `wartetext Erste Schritte`
- ✓ Z11 `taste escape`
- ✓ Z12 `taste cmd+3`
- ✓ Z13 `wartetext Sammlung`
- ✓ Z14 `klicketext Suchen`
- ✓ Z15 `gleich document.activeElement.getAttribute('placeholder') => Name, Set, Kennung oder Notiz`
- ✓ Z16 `notiz Die Palette öffnet über ⌘K und nimmt den Fokus`
- ✓ Z17 `taste cmd+k`
- ✓ Z18 `warte [role=dialog]`
- ✓ Z19 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z20 `zeige document.activeElement.getAttribute('aria-label')`
  ```
  "Wohin?"
  ```
- ✓ Z21 `tippe samm`
- ✓ Z22 `abzug palette`
  ```
  Bild: palette.png
  ```
- ✓ Z23 `notiz Tab bleibt im Dialog — zwölfmal`
- ✓ Z24 `taste tab`
- ✓ Z25 `taste tab`
- ✓ Z26 `taste tab`
- ✓ Z27 `taste tab`
- ✓ Z28 `taste tab`
- ✓ Z29 `taste tab`
- ✓ Z30 `taste tab`
- ✓ Z31 `taste tab`
- ✓ Z32 `taste tab`
- ✓ Z33 `taste tab`
- ✓ Z34 `taste tab`
- ✓ Z35 `taste tab`
- ✓ Z36 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z37 `taste shift+tab`
- ✓ Z38 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z39 `notiz Escape schließt, und der Fokus kehrt zurück`
- ✓ Z40 `taste escape`
- ✓ Z41 `pruefe ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z42 `gleich document.activeElement.getAttribute('placeholder') => Name, Set, Kennung oder Notiz`

**Ergebnis: bestanden** — 37 Schritte.
