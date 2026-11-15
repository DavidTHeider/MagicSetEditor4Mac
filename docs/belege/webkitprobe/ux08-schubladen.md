# WebKit-Probe — ux08-schubladen.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:55Z |

- ✓ Z11 `motor wk2`
- ✓ Z12 `lade /`
- ✓ Z13 `frist 20000`
- ✓ Z14 `warte nav`
- ✓ Z15 `wartetext Erste Schritte`
- ✓ Z16 `taste escape`
- ✓ Z17 `taste cmd+2`
- ✓ Z18 `wartetext Werkstatt-Testkarten`
- ✓ Z19 `klicketext Werkstatt-Testkarten`
- ✓ Z20 `wartetext Wie viel du siehst`
- ✓ Z21 `notiz Mit der Maus: Klick auf »Eigenschaften«`
- ✓ Z22 `klicketext Eigenschaften`
- ✓ Z23 `warte [role=dialog]`
- ✓ Z24 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z25 `zeige document.activeElement.getAttribute('role') + ': ' + document.activeElement.innerText.replace(/\s+/g, ' ').slice(0, 40)`
  ```
  "tab: Stil"
  ```
- ✓ Z26 `abzug schublade-eigenschaften`
  ```
  Bild: schublade-eigenschaften.png
  ```
- ✓ Z27 `taste tab`
- ✓ Z28 `taste tab`
- ✓ Z29 `taste tab`
- ✓ Z30 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z31 `taste escape`
- ✓ Z32 `pruefe ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z33 `gleich document.activeElement.innerText.trim() => Eigenschaften`
- ✓ Z34 `notiz Mit der Tastatur: Enter auf »Werkzeuge«`
- ✓ Z35 `taste tab`
- ✓ Z36 `gleich document.activeElement.innerText.trim() => Werkzeuge`
- ✓ Z37 `taste enter`
- ✓ Z38 `warte [role=dialog]`
- ✓ Z39 `pruefe [...document.querySelectorAll('[role=dialog]')].find((e) => e.getClientRects().length > 0).contains(document.activeElement)`
- ✓ Z40 `abzug schublade-werkzeuge`
  ```
  Bild: schublade-werkzeuge.png
  ```
- ✓ Z41 `taste escape`
- ✓ Z42 `pruefe ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z43 `gleich document.activeElement.innerText.trim() => Werkzeuge`

**Ergebnis: bestanden** — 33 Schritte.
