# WebKit-Probe — ux02-ziffern.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:25Z |

- ✓ Z6 `motor wk2`
- ✓ Z7 `lade /`
- ✓ Z8 `frist 20000`
- ✓ Z9 `warte nav`
- ✓ Z10 `wartetext Erste Schritte`
- ✓ Z11 `taste escape`
- ✓ Z12 `notiz Vorher: ⌘3 führt zur Sammlung, die Leiste zeigt ⌘3 bei Sammeln`
- ✓ Z13 `taste cmd+3`
- ✓ Z14 `wartetext Sammlung`
- ✓ Z15 `gleich [...document.querySelectorAll('nav[aria-label="Hauptnavigation"] button')].find((b) => b.innerText.includes('Sammeln')).innerText.includes('⌘3') => true`
- ✓ Z16 `notiz An der Kürzeltafel ⌘3 auf die Karten legen`
- ✓ Z17 `taste cmd+,`
- ✓ Z21 `bis [...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Tastenkürzel')`
- ✓ Z22 `pruefe ([...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Tastenkürzel').dataset.ziel = 'kuerzelfalte', true)`
- ✓ Z23 `klicke [data-ziel="kuerzelfalte"]`
- ✓ Z24 `bis document.querySelector('[data-ziel="kuerzelfalte"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z25 `warte select[aria-label="Seite für ⌘3"]`
- ✓ Z26 `pruefe (() => { const s = document.querySelector('select[aria-label="Seite für ⌘3"]'); s.scrollIntoView({ block: 'center' }); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, 'karten'); s.dispatchEvent(new Event('change', { bubbles: true })); return true })()`
- ✓ Z27 `bis document.querySelector('select[aria-label="Seite für ⌘3"]').value === 'karten'`
- ✓ Z28 `bis [...document.querySelectorAll('table tr')].some((z) => z.innerText.startsWith('⌘3') && z.innerText.includes('Karten'))`
- ✓ Z29 `abzug tafel`
  ```
  Bild: tafel.png
  ```
- ✓ Z30 `notiz Jetzt führt ⌘3 zu den Karten, und die Leiste hat bei Sammeln keine Ziffer`
- ✓ Z31 `taste cmd+3`
- ✓ Z32 `bis [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].some((l) => l.textContent.includes('Karten'))`
- ✓ Z33 `zeige [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ')`
  ```
  "Gestalten › Karten"
  ```
- ✓ Z34 `gleich [...document.querySelectorAll('nav[aria-label="Hauptnavigation"] button')].find((b) => b.innerText.includes('Sammeln')).innerText.includes('⌘') => false`
- ✓ Z35 `abzug karten`
  ```
  Bild: karten.png
  ```
- ✓ Z36 `notiz Zurück zur Vorgabe`
- ✓ Z37 `taste cmd+,`
- ✓ Z39 `bis document.querySelector('[data-ziel="kuerzelfalte"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z40 `wartetext Zurück zur Vorgabe`
- ✓ Z41 `klicketext Zurück zur Vorgabe`
- ✓ Z42 `bis document.querySelector('select[aria-label="Seite für ⌘3"]').value === 'sammlung'`
- ✓ Z43 `taste cmd+3`
- ✓ Z44 `wartetext Sammlung`
- ✓ Z45 `gleich [...document.querySelectorAll('nav[aria-label="Hauptnavigation"] button')].find((b) => b.innerText.includes('Sammeln')).innerText.includes('⌘3') => true`
- ✓ Z46 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z47 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 38 Schritte.
