# WebKit-Probe — ux07-baum.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:54Z |

- ✓ Z8 `motor wk2`
- ✓ Z9 `lade /`
- ✓ Z10 `frist 20000`
- ✓ Z11 `warte nav`
- ✓ Z12 `wartetext Erste Schritte`
- ✓ Z13 `taste escape`
- ✓ Z14 `taste cmd+k`
- ✓ Z15 `tippe Vorlagen`
- ✓ Z16 `pause 400`
- ✓ Z19 `taste down`
- ✓ Z20 `taste enter`
- ✓ Z23 `pruefe (window.__baum = () => [...document.querySelectorAll('main ul')].find((u) => u.checkVisibility() && u.querySelector('button[aria-expanded]') && u.closest('section, div') && [...u.querySelectorAll('button')].some((k) => k.tabIndex === 0)), true)`
- ✓ Z24 `bis window.__baum() !== undefined`
- ✓ Z30 `pruefe (window.__knoten = () => [...window.__baum().querySelectorAll('[data-baumknoten]')], window.__knoten().length > 1)`
- ✓ Z31 `notiz Eine Tabstelle für alle Knoten — die Kopieaktionen bleiben je für sich erreichbar`
- ✓ Z32 `zeige window.__knoten().filter((k) => k.tabIndex === 0).length + ' Knoten mit tabIndex 0 von ' + window.__knoten().length + ', Kopieaktionen: ' + window.__baum().querySelectorAll('button:not([data-baumknoten])').length`
  ```
  "1 Knoten mit tabIndex 0 von 15, Kopieaktionen: 7"
  ```
- ✓ Z33 `gleich window.__knoten().filter((k) => k.tabIndex === 0).length => 1`
- ✓ Z34 `pruefe [...window.__baum().querySelectorAll('button:not([data-baumknoten])')].every((k) => k.tabIndex === 0)`
- ✓ Z35 `notiz Alles beginnt aufgeklappt, und die Leertaste klappt`
- ✓ Z36 `pruefe (window.__baum().querySelector('button[aria-expanded]').dataset.ziel = 'zweig', true)`
- ✓ Z37 `klicke [data-ziel="zweig"]`
- ✓ Z38 `gleich document.querySelector('[data-ziel="zweig"]').getAttribute('aria-expanded') => false`
- ✓ Z39 `taste space`
- ✓ Z40 `gleich document.querySelector('[data-ziel="zweig"]').getAttribute('aria-expanded') => true`
- ✓ Z41 `notiz Pfeil ab läuft zum nächsten Knopf, Ende ans Ende, Pos1 an den Anfang`
- ✓ Z42 `taste down`
- ✓ Z43 `pruefe document.activeElement !== document.querySelector('[data-ziel="zweig"]') && window.__baum().contains(document.activeElement)`
- ✓ Z44 `taste end`
- ✓ Z45 `pruefe (() => { const k = window.__knoten(); return document.activeElement === k[k.length - 1] })()`
- ✓ Z46 `taste home`
- ✓ Z47 `pruefe (() => { const k = window.__knoten(); return document.activeElement === k[0] })()`
- ✓ Z48 `pruefe getComputedStyle(document.activeElement).outlineStyle !== 'none'`
- ✓ Z49 `abzug baum`
  ```
  Bild: baum.png
  ```
- ✓ Z50 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 34 Schritte.
