# WebKit-Probe — ux08-handelswahl.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:38Z |

- ✓ Z12 `motor wk2`
- ✓ Z13 `lade /`
- ✓ Z14 `frist 20000`
- ✓ Z15 `warte nav`
- ✓ Z16 `wartetext Erste Schritte`
- ✓ Z17 `taste escape`
- ✓ Z18 `taste cmd+5`
- ✓ Z19 `wartetext Mit wem`
- ✓ Z20 `notiz Seit UXR-15 (#1923) steht die Wahl des Gegenübers in der Schublade hinter „Neuer Tausch …"`
- ✓ Z21 `klicketext Neuer Tausch …`
- ✓ Z22 `wartetext Bo`
- ✓ Z23 `pruefe ([...document.querySelectorAll('button')].find((b) => (b.getAttribute('aria-label') || '').includes('Bo')).dataset.ziel = 'bo', true)`
- ✓ Z24 `klicke [data-ziel="bo"]`
- ✓ Z25 `notiz Seit TRADE-02 (#1716) steht die Wahl in „Mit wem"; ein neuer Handel öffnet in „Was tauscht ihr"`
- ✓ Z26 `klicke [data-handelsassistent] nav li:nth-child(1) button`
- ✓ Z27 `wartetext Welcher Handel`
- ✓ Z28 `pruefe (window.__wahl = () => [...document.querySelectorAll('fieldset')].find((f) => f.querySelector('legend') && f.querySelector('legend').textContent === 'Welcher Handel'), window.__vorgang = () => [...document.querySelectorAll('section')].find((s) => s.querySelector('h3') && s.querySelector('h3').textContent === 'Ein Handel von der Eröffnung bis zur Übergabe'), true)`
- ✓ Z29 `notiz Ein neuer Handel ist vorgewählt — mit Bo gibt es noch keinen, und nirgends steht ein Feld für eine Kennung`
- ✓ Z30 `bis document.body.innerText.includes('Mit diesem Gegenüber gibt es noch keinen Handel.')`
- ✓ Z31 `pruefe window.__wahl().querySelector('input[type=radio]:checked').value === ''`
- ✓ Z32 `pruefe ![...document.querySelectorAll('main label')].some((l) => l.checkVisibility() && /Welcher Handel|Kennung des Handels/.test(l.textContent))`
- ✓ Z33 `notiz Eine Zeile von dir, dann eröffnen`
- ✓ Z34 `klicke [data-handelsassistent] nav li:nth-child(2) button`
- ✓ Z38 `bis !!document.querySelector('button[aria-label="Zeile hinzufügen: Von dir"]')`
- ✓ Z39 `klicke button[aria-label="Zeile hinzufügen: Von dir"]`
- ✓ Z42 `bis [...window.__vorgang().querySelectorAll('button[aria-expanded]')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Kennung von Hand')`
- ✓ Z43 `pruefe ([...window.__vorgang().querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Kennung von Hand').dataset.ziel = 'kennungskasten', true)`
- ✓ Z44 `klicke [data-ziel="kennungskasten"]`
- ✓ Z45 `bis document.querySelector('[data-ziel="kennungskasten"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z46 `bis [...window.__vorgang().querySelectorAll('label')].some((l) => l.checkVisibility() && l.textContent.trim().startsWith('Druck'))`
- ✓ Z47 `pruefe ([...window.__vorgang().querySelectorAll('label')].find((l) => l.checkVisibility() && l.textContent.trim().startsWith('Druck')).control.dataset.ziel = 'druck', true)`
- ✓ Z48 `klicke [data-ziel="druck"]`
- ✓ Z49 `tippe scryfall:aaa-1`
- ✓ Z50 `klicketext Zeile hinzufügen`
- ✓ Z51 `notiz Die Zeile steht in der eigenen Spalte und nicht in der des Gegenübers`
- ✓ Z52 `pruefe (window.__spalte = (titel) => [...document.querySelectorAll('section[aria-labelledby]')].find((s) => document.getElementById(s.getAttribute('aria-labelledby'))?.textContent === titel), window.__spalte('Von dir').innerText.includes('1 × scryfall:aaa-1') && !window.__spalte('Vom Gegenüber').innerText.includes('scryfall:aaa-1'))`
- ✓ Z53 `klicketext Eröffnen`
- ✓ Z54 `notiz Der neue Handel steht in der Wahl und ist gewählt — seine Kennung hat die Leitung vergeben`
- ✓ Z55 `bis window.__wahl().querySelectorAll('input[type=radio]').length === 2`
- ✓ Z56 `pruefe window.__wahl().querySelector('input[type=radio]:checked').value.startsWith('h-')`
- ✓ Z57 `zeige [...window.__wahl().querySelectorAll('label')].map((l) => l.innerText.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "Ein neuer HandelSeine Kennung vergibt das Programm beim Eröffnen. | Entwurf · Revision 1 · 1 Position"
  ```
- ✓ Z58 `notiz Die Zusagen bieten die eigenen Exemplare dieses Drucks an`
- ✓ Z59 `klicke [data-handelsassistent] nav li:nth-child(3) button`
- ✓ Z60 `bis [...document.querySelectorAll('fieldset legend')].some((l) => /^1 × \S/.test(l.textContent))`
- ✓ Z61 `zeige [...document.querySelectorAll('fieldset')].filter((f) => f.querySelector('legend') && /^1 × \S/.test(f.querySelector('legend').textContent)).map((f) => f.innerText.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "1 × scryfall:aaa-1 Keine Zeile deiner Sammlung passt zu Zustand, Sprache und Ausführung dieser Zeile. Ergänze die Sammlung, oder lege die Zeile im Tausch mit passenden Angaben neu an. Festhalten"
  ```
- ✓ Z62 `abzug handelswahl`
  ```
  Bild: handelswahl.png
  ```
- ✓ Z63 `notiz Für Fachleute steht die Kennung mit einer Erklärung da`
- ✓ Z64 `klicke [data-handelsassistent] nav li:nth-child(1) button`
- ✓ Z65 `klicketext Einen Handel über seine Kennung öffnen`
- ✓ Z66 `bis document.body.innerText.includes('Gewählt ist der Handel h-')`
- ✓ Z67 `abzug handelswahl-fachleute`
  ```
  Bild: handelswahl-fachleute.png
  ```
- ✓ Z68 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 52 Schritte.
