# WebKit-Probe — ux05-erstesset.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:47Z |

- ✓ Z6 `motor wk2`
- ✓ Z7 `lade /`
- ✓ Z8 `frist 20000`
- ✓ Z9 `warte nav`
- ✓ Z10 `wartetext Erste Schritte`
- ✓ Z11 `taste escape`
- ✓ Z12 `taste cmd+2`
- ✓ Z15 `wartetext Neues Set`
- ✓ Z16 `klicketext Neues Set`
- ✓ Z17 `wartetext Ein neues Set`
- ✓ Z23 `notiz Vorbedingung: zwei Spiele, also drei Schritte (testdaten/assistent)`
- ✓ Z24 `bis document.querySelectorAll('ol[aria-label="Schritte zum neuen Set"] li').length === 3`
- ✓ Z28 `notiz Der Fokus steht im Namen, und darunter steht, was vorgewählt ist`
- ✓ Z29 `bis document.querySelector('ol[aria-label="Schritte zum neuen Set"] li[aria-current="step"]')?.lastChild.textContent === 'Name'`
- ✓ Z30 `bis document.activeElement.labels && document.activeElement.labels.length > 0 && document.activeElement.labels[0].textContent.includes('Name des Sets')`
- ✓ Z31 `zeige [...document.querySelectorAll('section p')].map((p) => p.textContent).find((t) => t.includes(' · ')) || 'ohne Zusammenfassung'`
  ```
  "probe · 5 Karten · probe-kunst"
  ```
- ✓ Z32 `tippe Nachtprobe`
- ✓ Z33 `pruefe [...document.querySelectorAll('section p')].some((p) => p.textContent === 'Wahlspiel · Mondlicht')`
- ✓ Z34 `klicketext Set anlegen`
- ✓ Z35 `wartetext Nachtprobe`
- ✓ Z36 `notiz Das Set steht mit seinem Namen da, und der Editor öffnet es`
- ✓ Z37 `bis [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].some((l) => l.textContent.includes('Nachtprobe'))`
- ✓ Z38 `zeige [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ')`
  ```
  "Gestalten › Nachtprobe › Karten"
  ```
- ✓ Z39 `abzug erstes-set`
  ```
  Bild: erstes-set.png
  ```
- ✓ Z40 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 25 Schritte.
