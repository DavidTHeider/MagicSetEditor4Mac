# WebKit-Probe — play-beginn.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:28Z |

- ✓ Z15 `motor wk2`
- ✓ Z16 `lade /`
- ✓ Z17 `frist 20000`
- ✓ Z18 `warte nav`
- ✓ Z19 `wartetext Erste Schritte`
- ✓ Z20 `taste escape`
- ✓ Z21 `taste cmd+4`
- ✓ Z22 `wartetext Partie anfangen`
- ✓ Z23 `klicketext Partie anfangen`
- ✓ Z24 `wartetext Weiterspielen`
- ✓ Z25 `klicketext Weiterspielen`
- ✓ Z26 `klicketext Weiterspielen als Bo`
- ✓ Z27 `warte canvas`
- ✓ Z28 `wartetext Der Beginn der Partie`
- ✓ Z29 `notiz Die Tafel steht am eigenen Tisch unter h2 — der Tisch ist eine eigene Seite unter h1 (#1992); beim Gast wäre es h3`
- ✓ Z30 `pruefe (window.__beginn = () => [...document.querySelectorAll('section')].find((s) => s.querySelector('h2')?.textContent === 'Der Beginn der Partie'), !!window.__beginn())`
- ✓ Z31 `notiz Vor der Starthand steht genau sie da, und wer sonst noch nicht so weit ist`
- ✓ Z32 `gleich [...window.__beginn().querySelectorAll('button')].map((b) => b.innerText.trim()).join(', ') => Starthand ziehen`
- ✓ Z33 `pruefe document.body.innerText.includes('Ada hat noch keine Starthand gezogen.')`
- ✓ Z34 `klicketext Starthand ziehen`
- ✓ Z35 `wartetext Deine Starthand hat 7 Karten.`
- ✓ Z36 `notiz Ein Mulligan zieht neu und verlangt eine Karte unter die Bibliothek`
- ✓ Z37 `klicketext Mulligan`
- ✓ Z38 `wartetext Lege eine Karte unter deine Bibliothek.`
- ✓ Z39 `pruefe [...document.querySelectorAll('button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Unter die Bibliothek legen').disabled`
- ✓ Z40 `gleich document.querySelectorAll('ul[aria-label="Deine Hand"] input[type=checkbox]').length => 7`
- ✓ Z41 `klicke ul[aria-label="Deine Hand"] input[type=checkbox]`
- ✓ Z42 `wartetext 1 von 1 gewählt`
- ✓ Z43 `pruefe ![...document.querySelectorAll('button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Unter die Bibliothek legen').disabled`
- ✓ Z44 `abzug beginn-unterlegen`
  ```
  Bild: beginn-unterlegen.png
  ```
- ✓ Z45 `klicketext Unter die Bibliothek legen`
- ✓ Z46 `wartetext Deine Starthand hat 6 Karten.`
- ✓ Z47 `notiz Behalten beendet den eigenen Beginn; die Tafel nennt nur noch Ada`
- ✓ Z48 `klicketext Hand behalten`
- ✓ Z49 `bis !document.body.innerText.includes('Deine Starthand hat')`
- ✓ Z53 `notiz Der Vorhang geht zu, Ada ist dran — Bo bleibt (#2375)`
- ✓ Z54 `wartetext Gerät an Ada weitergeben`
- ✓ Z55 `klicketext Weiter als Bo`
- ✓ Z56 `bis !document.body.innerText.includes('Gerät an Ada weitergeben')`
- ✓ Z57 `pruefe (() => { const t = window.__beginn(); return !!t && t.querySelectorAll('button').length === 0 && t.innerText.includes('Ada hat noch keine Starthand gezogen.') })()`
- ✓ Z58 `notiz Und die Spielerleiste nennt den Mulligan`
- ✓ Z59 `pruefe document.body.innerText.includes('Mulligans: 1')`
- ✓ Z60 `abzug beginn-behalten`
  ```
  Bild: beginn-behalten.png
  ```
- ✓ Z61 `notiz Ada hat noch nicht gezogen — sie hat die Priorität des ersten Zugs und darf trotzdem nicht passen (CR 103.8)`
- ✓ Z62 `pruefe (window.__sicht = (wert) => { const f = [...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Sicht für')).querySelector('select'); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(f, wert); f.dispatchEvent(new Event('change', { bubbles: true })); return true }, window.__passen = () => [...document.querySelectorAll('button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Passen'), true)`
- ✓ Z63 `pruefe window.__sicht('ada')`
- ✓ Z64 `wartetext Der Zug wartet, bis jeder seine Starthand behalten hat.`
- ✓ Z65 `pruefe window.__passen().disabled`
- ✓ Z66 `abzug beginn-zug-wartet`
  ```
  Bild: beginn-zug-wartet.png
  ```
- ✓ Z67 `notiz Ada zieht und behält — danach ist der Beginn vorbei, und Passen geht wieder`
- ✓ Z68 `klicketext Starthand ziehen`
- ✓ Z69 `wartetext Deine Starthand hat 7 Karten.`
- ✓ Z70 `klicketext Hand behalten`
- ✓ Z71 `bis !document.body.innerText.includes('Der Zug wartet, bis jeder')`
- ✓ Z72 `pruefe !window.__passen().disabled`
- ✓ Z73 `abzug beginn-vorbei`
  ```
  Bild: beginn-vorbei.png
  ```
- ✓ Z74 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z75 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 58 Schritte.
