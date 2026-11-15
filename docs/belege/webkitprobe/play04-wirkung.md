# WebKit-Probe — play04-wirkung.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:43Z |

- ✓ Z13 `motor wk2`
- ✓ Z14 `lade /`
- ✓ Z15 `frist 20000`
- ✓ Z16 `warte nav`
- ✓ Z17 `wartetext Erste Schritte`
- ✓ Z18 `taste escape`
- ✓ Z19 `taste cmd+4`
- ✓ Z20 `wartetext Partie anfangen`
- ✓ Z21 `klicketext Partie anfangen`
- ✓ Z22 `wartetext Weiterspielen`
- ✓ Z23 `klicketext Weiterspielen`
- ✓ Z24 `klicketext Weiterspielen als Ada`
- ✓ Z25 `warte canvas`
- ✓ Z26 `pause 1500`
- ✓ Z27 `notiz Der Beginn ist vorbei — `play-beginn` hat ihn in derselben Partie zu Ende gebracht; vorher nähme der Tisch nur ihn an (CR 103.8, #1851)`
- ✓ Z34 `notiz Vorbedingung: der Beginn ist vorbei — »Passen« steht für Ada offen`
- ✓ Z35 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Passen' && !b.disabled)`
- ✓ Z36 `notiz Die oberste Karte aus Adas Bibliothek aufs Schlachtfeld — ein Eintritt`
- ✓ Z37 `pruefe (() => { const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => (x.getAttribute('aria-label') || x.textContent).includes('Bibliothek')); const unten = b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)); unten.dataset.ziel = 'bibliothek'; return true })()`
- ✓ Z38 `klicke [data-ziel="bibliothek"]`
- ✓ Z39 `wartetext Mit dieser Karte`
- ✓ Z40 `klicketext Schieben nach…`
- ✓ Z41 `bis document.querySelector('[role=group][aria-label="Schieben nach"]') !== null`
- ✓ Z42 `pruefe (() => { const s = document.querySelector('[role=group][aria-label="Schieben nach"] select'); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, 'tisch'); s.dispatchEvent(new Event('change', { bubbles: true })); return true })()`
- ✓ Z43 `klicketext Stellen`
- ✓ Z44 `abzug wirkung-eintritt`
  ```
  Bild: wirkung-eintritt.png
  ```
- ✓ Z47 `bis [...document.querySelectorAll('[role=application] li button')].some((b) => (b.getAttribute('aria-label') || b.textContent).includes('· Schlachtfeld'))`
- ✓ Z48 `notiz Dann getappt — eine Drehung`
- ✓ Z50 `pruefe ([...document.querySelectorAll('[role=application] li button')].find((b) => { const t = b.getAttribute('aria-label') || b.textContent; return t.includes('· Schlachtfeld') && !t.includes('getappt') }).dataset.ziel = 'schlachtfeld', true)`
- ✓ Z53 `klicke [data-ziel="schlachtfeld"]`
- ✓ Z54 `abzug wirkung-drehung`
  ```
  Bild: wirkung-drehung.png
  ```
- ✓ Z55 `bis [...document.querySelectorAll('[role=application] li button')].some((b) => /Schlachtfeld.*getappt/.test(b.getAttribute('aria-label') || b.textContent))`
- ✓ Z56 `notiz In der Liste trägt die Zone dieselbe Wirkung — als Rahmen und als Zeichen`
- ✓ Z60 `klicketext Als Liste`
- ✓ Z61 `bis [...document.querySelectorAll('button')].some((b) => /getappt/.test(b.getAttribute('aria-label') || b.textContent))`
- ✓ Z62 `pruefe (window.__wirkungen = [], new MutationObserver(() => { for (const e of document.querySelectorAll('[data-wirkung]')) { const w = e.getAttribute('data-wirkung'); if (!window.__wirkungen.includes(w)) window.__wirkungen.push(w + ' ' + (e.textContent || '').trim()) } }).observe(document.body, { subtree: true, childList: true, attributes: true }), true)`
- ✓ Z63 `pruefe ([...document.querySelectorAll('button')].find((b) => /getappt/.test(b.getAttribute('aria-label') || b.textContent)).dataset.ziel = 'liste', true)`
- ✓ Z64 `klicke [data-ziel="liste"]`
- ✓ Z65 `bis window.__wirkungen.length > 0`
- ✓ Z66 `zeige JSON.stringify(window.__wirkungen)`
  ```
  "[\"drehung ↻Langtext (getappt)Ada\",\"drehung ↻Langtext (getappt)Ada\",\"drehung ↻Langtext (getappt)Ada\",\"drehung ↻LangtextAda\"]"
  ```
- ✓ Z67 `pruefe window.__wirkungen.some((w) => w.startsWith('drehung'))`
- ✓ Z68 `abzug wirkung-liste`
  ```
  Bild: wirkung-liste.png
  ```
- ✓ Z69 `klicketext Als Tisch`
- ✓ Z70 `notiz Die Ereignisliste sagt es in Worten, und das Zeichen bleibt stumm`
- ✓ Z73 `zeige (() => { const h = [...document.querySelectorAll('h2')].find((x) => x.textContent === 'Zuletzt geschehen'); return [...h.parentElement.querySelectorAll('li')].slice(0, 3).map((l) => l.innerText.replace(/\s+/g, ' ').trim()).join(' | ') })()`
  ```
  "Ada: Enttappen — Langtext"
  ```
- ✓ Z74 `pruefe (() => { const h = [...document.querySelectorAll('h2')].find((x) => x.textContent === 'Zuletzt geschehen'); return [...h.parentElement.querySelectorAll('li')].some((l) => l.innerText.includes('Enttappen')) })()`
- ✓ Z75 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z76 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 48 Schritte.
