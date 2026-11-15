# WebKit-Probe — ux07-handelsassistent.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:02Z |

- ✓ Z10 `motor wk2`
- ✓ Z11 `lade /`
- ✓ Z12 `frist 20000`
- ✓ Z13 `warte nav`
- ✓ Z14 `wartetext Erste Schritte`
- ✓ Z15 `taste escape`
- ✓ Z16 `taste cmd+5`
- ✓ Z17 `wartetext Mit wem`
- ✓ Z18 `notiz Die zwei Laschen, mit der Tastatur`
- ✓ Z19 `pruefe [...document.querySelectorAll('[role=tab]')].map((l) => l.textContent.trim()).join('|').includes('Mit wem|Was es wert ist')`
- ✓ Z20 `pruefe (document.querySelector('[role=tab][aria-selected=true]').dataset.ziel = 'lasche', true)`
- ✓ Z21 `klicke [data-ziel="lasche"]`
- ✓ Z22 `taste right`
- ✓ Z23 `gleich document.activeElement.getAttribute('role') + ' ' + document.activeElement.textContent.trim() => tab Was es wert ist`
- ✓ Z24 `gleich document.querySelector('[role=tab][aria-selected=true]').textContent.trim() => Was es wert ist`
- ✓ Z25 `taste left`
- ✓ Z26 `gleich document.querySelector('[role=tab][aria-selected=true]').textContent.trim() => Mit wem`
- ✓ Z27 `notiz Das Gegenüber aufschlagen — der Knopf nennt es im Namen; seit UXR-15 (#1923) in der Schublade hinter „Neuer Tausch …"`
- ✓ Z28 `klicketext Neuer Tausch …`
- ✓ Z29 `wartetext Bo`
- ✓ Z30 `pruefe ([...document.querySelectorAll('button')].find((b) => (b.getAttribute('aria-label') || '').includes('Bo')).dataset.ziel = 'bo', true)`
- ✓ Z31 `zeige document.querySelector('[data-ziel="bo"]').getAttribute('aria-label') + ' / sichtbar: ' + document.querySelector('[data-ziel="bo"]').innerText`
  ```
  "Bo aufschlagen / sichtbar: Aufschlagen"
  ```
- ✓ Z32 `klicke [data-ziel="bo"]`
- ✓ Z33 `bis document.querySelectorAll('main h3, main h2').length > 2`
- ✓ Z34 `zeige [...document.querySelectorAll('main h2, main h3')].map((h) => h.textContent.trim()).join(' | ')`
  ```
  "Weiterarbeiten | Was Aufmerksamkeit braucht 1 | Was gerade läuft0 | Ein Handel von der Eröffnung bis zur Übergabe | Welche Exemplare weggehen | Was zum Gegenüber hinausgeht | Einen Handel verbuchen"
  ```
- ✓ Z35 `abzug handel-hell`
  ```
  Bild: handel-hell.png
  ```
- ✓ Z36 `notiz Jeder Knopf und jedes Feld im Handelsraum ist mit Tab erreichbar, der Fokus sichtbar`
- ✓ Z37 `pruefe (window.__erreicht = new Set(), document.addEventListener('focusin', (e) => window.__erreicht.add(e.target)), true)`
- ✓ Z38 `taste tab`
- ✓ Z39 `taste tab`
- ✓ Z40 `taste tab`
- ✓ Z41 `taste tab`
- ✓ Z42 `taste tab`
- ✓ Z43 `taste tab`
- ✓ Z44 `taste tab`
- ✓ Z45 `taste tab`
- ✓ Z46 `taste tab`
- ✓ Z47 `taste tab`
- ✓ Z48 `taste tab`
- ✓ Z49 `taste tab`
- ✓ Z50 `taste tab`
- ✓ Z51 `taste tab`
- ✓ Z52 `taste tab`
- ✓ Z53 `taste tab`
- ✓ Z54 `taste tab`
- ✓ Z55 `taste tab`
- ✓ Z56 `taste tab`
- ✓ Z57 `taste tab`
- ✓ Z58 `zeige window.__erreicht.size + ' verschiedene Ziele mit Tab erreicht'`
  ```
  "20 verschiedene Ziele mit Tab erreicht"
  ```
- ✓ Z59 `pruefe window.__erreicht.size >= 10`
- ✓ Z60 `pruefe [...window.__erreicht].every((e) => e.matches('button, input, select, textarea, a, [tabindex], [role=tab]'))`
- ✓ Z61 `pruefe getComputedStyle(document.activeElement).outlineStyle !== 'none'`
- ✓ Z62 `notiz Dasselbe im dunklen Erscheinungsbild`
- ✓ Z63 `erscheinung dunkel`
- ✓ Z64 `pause 500`
- ✓ Z65 `abzug handel-dunkel`
  ```
  Bild: handel-dunkel.png
  ```
- ✓ Z66 `erscheinung hell`
- ✓ Z67 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 58 Schritte.
