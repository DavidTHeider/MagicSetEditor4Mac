# WebKit-Probe — ux08-deckuebernahme.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:36Z |

- ✓ Z15 `motor wk2`
- ✓ Z16 `lade /`
- ✓ Z17 `frist 20000`
- ✓ Z18 `warte nav`
- ✓ Z19 `wartetext Erste Schritte`
- ✓ Z20 `taste escape`
- ✓ Z21 `taste cmd+4`
- ✓ Z22 `wartetext Probe-Deck`
- ✓ Z25 `klicketext Holen und weitergeben`
- ✓ Z26 `wartetext Ein Deck aus einem fremden Profil holen`
- ✓ Z27 `klicketext Ein Deck aus einem fremden Profil holen`
- ✓ Z28 `pruefe (window.__holen = () => [...document.querySelectorAll('section')].find((s) => s.querySelector('h2') && s.querySelector('h2').textContent === 'Ein Deck aus einem fremden Profil holen'), window.__knopf = () => [...window.__holen().querySelectorAll('button')].find((b) => b.textContent === 'Holen'), true)`
- ✓ Z29 `notiz Cleos Profil steht zur Wahl — das eigene nicht, und kein Feld für eine Kennung ist zu sehen`
- ✓ Z30 `bis [...window.__holen().querySelectorAll('option')].some((o) => o.textContent.includes('Cleo'))`
- ✓ Z31 `pruefe ![...window.__holen().querySelectorAll('label')].some((l) => l.checkVisibility() && /Kennung des/.test(l.textContent))`
- ✓ Z32 `zeige [...window.__holen().querySelectorAll('option')].map((o) => o.textContent.trim()).join(' | ')`
  ```
  "Ein Profil wählen … | Cleo · Mein Profil"
  ```
- ✓ Z33 `pruefe window.__knopf().disabled`
- ✓ Z34 `notiz Das Profil wählen — dann stehen seine Decks da`
- ✓ Z35 `pruefe (window.__profilwahl = [...window.__holen().querySelectorAll('select')][0], window.__profilwahl.dataset.ziel = 'profil', true)`
- ✓ Z36 `pruefe (window.__profilwahl.value = [...window.__profilwahl.options].find((o) => o.textContent.includes('Cleo')).value, window.__profilwahl.dispatchEvent(new Event('change', { bubbles: true })), true)`
- ✓ Z37 `bis [...window.__holen().querySelectorAll('option')].some((o) => o.textContent === 'Leihdeck')`
- ✓ Z38 `zeige [...window.__holen().querySelectorAll('select')].map((s) => [...s.options].map((o) => o.textContent.trim()).join(', ')).join(' ‖ ')`
  ```
  "Ein Profil wählen …, Cleo · Mein Profil ‖ Ein Deck wählen …, Leihdeck"
  ```
- ✓ Z39 `notiz Das Deck wählen — dann geht der Knopf`
- ✓ Z40 `pruefe (window.__deckwahl = [...window.__holen().querySelectorAll('select')][1], window.__deckwahl.value = [...window.__deckwahl.options].find((o) => o.textContent === 'Leihdeck').value, window.__deckwahl.dispatchEvent(new Event('change', { bubbles: true })), true)`
- ✓ Z41 `bis !window.__knopf().disabled`
- ✓ Z42 `pruefe (window.__holen().scrollIntoView(), true)`
- ✓ Z43 `abzug deckwahl`
  ```
  Bild: deckwahl.png
  ```
- ✓ Z44 `notiz Wer die Kennungen kennt, nennt sie selbst — der erste Satz sagt, wann`
- ✓ Z45 `klicketext Kennungen selbst nennen`
- ✓ Z46 `bis [...window.__holen().querySelectorAll('label')].some((l) => l.checkVisibility() && l.textContent.trim().startsWith('Kennung des Decks'))`
- ✓ Z47 `abzug deckwahl-fachleute`
  ```
  Bild: deckwahl-fachleute.png
  ```
- ✓ Z48 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 32 Schritte.
