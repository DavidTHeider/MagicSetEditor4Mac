# WebKit-Probe — play01-anfang.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:32Z |

- ✓ Z12 `motor wk2`
- ✓ Z13 `lade /`
- ✓ Z14 `frist 20000`
- ✓ Z15 `warte nav`
- ✓ Z16 `wartetext Erste Schritte`
- ✓ Z17 `taste escape`
- ✓ Z18 `taste cmd+4`
- ✓ Z19 `wartetext Partie anfangen`
- ✓ Z20 `klicketext Partie anfangen`
- ✓ Z21 `wartetext Wer sitzt am Tisch`
- ✓ Z22 `notiz Die Seite heißt, was sie tut, und ein einziges Deck genügt`
- ✓ Z23 `gleich [...document.querySelectorAll('main h1')].find((e) => e.getClientRects().length > 0).textContent => Eine Partie anfangen`
- ✓ Z24 `gleich [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ') => Spielen › Partie anfangen`
- ✓ Z25 `pruefe document.querySelectorAll('main select').length >= 2`
- ✓ Z26 `notiz Beide Sitze wählen dasselbe Deck`
- ✓ Z27 `pruefe (() => { const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; [...document.querySelectorAll('section[aria-label="Wer sitzt am Tisch"] select')].forEach((e) => { setze.call(e, '1'); e.dispatchEvent(new Event('change', { bubbles: true })) }); return true })()`
- ✓ Z28 `warte section[aria-label="Nach welchen Regeln"]`
- ✓ Z29 `pause 500`
- ✓ Z30 `notiz Die Regelgrenzen stehen eingeklappt, der Anfang steht davor im Blick`
- ✓ Z31 `pruefe [...document.querySelectorAll('main button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Regelgrenzen im Einzelnen').getAttribute('aria-expanded') === 'false'`
- ✓ Z32 `pruefe ![...document.querySelectorAll('main h3')].some((h) => h.textContent === 'Was durchgesetzt wird')`
- ✓ Z33 `zeige (() => { const k = [...document.querySelectorAll('main button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Partie anfangen'); return 'Seite ' + document.querySelector('main').scrollHeight + ' hoch, Startknopf bei ' + Math.round(k.getBoundingClientRect().top + scrollY) })()`
  ```
  "Seite 1205 hoch, Startknopf bei 1164"
  ```
- ✓ Z34 `pruefe [...document.querySelectorAll('main button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Partie anfangen').getBoundingClientRect().top + scrollY < 1400`
- ✓ Z35 `notiz Keine rohe Auszeichnung in den Regeltexten`
- ✓ Z36 `klicketext Regelgrenzen im Einzelnen`
- ✓ Z37 `bis [...document.querySelectorAll('main h3')].some((h) => h.textContent === 'Was durchgesetzt wird')`
- ✓ Z38 `pruefe !document.querySelector('main').innerText.includes('**')`
- ✓ Z39 `abzug anfang`
  ```
  Bild: anfang.png
  ```
- ✓ Z40 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z41 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 30 Schritte.
