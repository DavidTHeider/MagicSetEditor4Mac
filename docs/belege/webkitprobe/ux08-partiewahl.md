# WebKit-Probe — ux08-partiewahl.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:49Z |

- ✓ Z15 `motor wk2`
- ✓ Z16 `lade /`
- ✓ Z17 `frist 20000`
- ✓ Z18 `warte nav`
- ✓ Z19 `wartetext Erste Schritte`
- ✓ Z20 `taste escape`
- ✓ Z21 `taste cmd+4`
- ✓ Z22 `wartetext Partie anfangen`
- ✓ Z23 `klicketext Partie anfangen`
- ✓ Z27 `wartetext Bei jemandem mitspielen …`
- ✓ Z28 `klicketext Bei jemandem mitspielen …`
- ✓ Z29 `pruefe (window.__gast = () => [...document.querySelectorAll('[role=dialog]')].find((d) => d.getAttribute('aria-label') === 'An einem fremden Tisch mitspielen'), window.__mitspielen = () => [...window.__gast().querySelectorAll('button')].find((b) => b.textContent === 'Mitspielen'), true)`
- ✓ Z30 `bis window.__gast() !== undefined`
- ✓ Z31 `notiz Bo steht zur Wahl — und nirgends ein Feld für eine Peerkennung oder eine Partie`
- ✓ Z32 `bis window.__gast().querySelectorAll('input[type=radio]').length === 1`
- ✓ Z33 `pruefe ![...window.__gast().querySelectorAll('label')].some((l) => l.checkVisibility() && /Peerkennung|Kennung der Partie/.test(l.textContent))`
- ✓ Z34 `zeige [...window.__gast().querySelectorAll('fieldset label')].map((l) => l.innerText.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "Bo · 580b-ba47-8782-72c2-9d87-8312-0ece-48b1"
  ```
- ✓ Z35 `pruefe window.__mitspielen().disabled`
- ✓ Z36 `notiz Bo wählen — dann geht der Knopf`
- ✓ Z37 `pruefe (window.__gast().querySelector('input[type=radio]').dataset.ziel = 'bo', true)`
- ✓ Z38 `klicke [data-ziel="bo"]`
- ✓ Z39 `bis !window.__mitspielen().disabled`
- ✓ Z40 `pruefe window.__gast().querySelector('input[type=radio]').checked`
- ✓ Z41 `abzug partiewahl`
  ```
  Bild: partiewahl.png
  ```
- ✓ Z42 `notiz Für einen Wirt vor 0.6.0 lässt sich die Partie nennen — der erste Satz sagt, warum es sonst nicht nötig ist, der Rest steht hinter dem Hilfesymbol`
- ✓ Z43 `klicketext Die Partie selbst nennen`
- ✓ Z44 `bis window.__gast().innerText.includes('Der Wirt nennt seine Partie mit seinem ersten Wort.')`
- ✓ Z45 `pruefe [...window.__gast().querySelectorAll('label')].some((l) => l.checkVisibility() && l.textContent.trim().startsWith('Kennung der Partie'))`
- ✓ Z46 `abzug partiewahl-fachleute`
  ```
  Bild: partiewahl-fachleute.png
  ```
- ✓ Z47 `notiz Bei der Leihe dasselbe — Bo und das Probe-Deck stehen zur Wahl, keine Kennung wird getippt`
- ✓ Z48 `taste escape`
- ✓ Z49 `bis !document.querySelector('[role=dialog]')?.checkVisibility()`
- ✓ Z50 `klicketext Verleihen oder leihen …`
- ✓ Z51 `pruefe (window.__leihe = () => [...document.querySelectorAll('[role=dialog]')].find((d) => d.getAttribute('aria-label') === 'Ein Deck verleihen oder leihen' && d.checkVisibility()), true)`
- ✓ Z52 `bis window.__leihe() !== undefined`
- ✓ Z53 `bis window.__leihe().querySelectorAll('input[type=radio]').length === 1`
- ✓ Z54 `bis [...window.__leihe().querySelectorAll('option')].some((o) => o.textContent === 'Probe-Deck')`
- ✓ Z55 `pruefe ![...window.__leihe().querySelectorAll('label')].some((l) => l.checkVisibility() && /Peerkennung|Kennung des Decks|Kennung der Partie/.test(l.textContent))`
- ✓ Z56 `zeige [...window.__leihe().querySelectorAll('fieldset label, option')].map((l) => l.textContent.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "Bo · 580b-ba47-8782-72c2-9d87-8312-0ece-48b1 | Ein Deck wählen … | Probe-Deck"
  ```
- ✓ Z57 `pruefe (window.__leihe().scrollIntoView(), true)`
- ✓ Z58 `abzug leihwahl`
  ```
  Bild: leihwahl.png
  ```
- ✓ Z59 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 42 Schritte.
