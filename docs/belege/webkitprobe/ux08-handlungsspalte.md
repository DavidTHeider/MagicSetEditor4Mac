# WebKit-Probe — ux08-handlungsspalte.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:42Z |

- ✓ Z9 `motor wk2`
- ✓ Z10 `lade /`
- ✓ Z11 `frist 20000`
- ✓ Z12 `warte nav`
- ✓ Z13 `wartetext Erste Schritte`
- ✓ Z14 `taste escape`
- ✓ Z15 `taste cmd+4`
- ✓ Z16 `wartetext Partie anfangen`
- ✓ Z17 `klicketext Partie anfangen`
- ✓ Z18 `wartetext Weiterspielen`
- ✓ Z19 `klicketext Weiterspielen`
- ✓ Z20 `klicketext Weiterspielen als Ada`
- ✓ Z21 `warte canvas`
- ✓ Z22 `pause 1500`
- ✓ Z23 `notiz Der Beginn ist vorbei — `play-beginn` hat ihn in derselben Partie zu Ende gebracht; vorher nähme der Tisch nur ihn an (CR 103.8, #1851)`
- ✓ Z30 `notiz Vorbedingung: der Beginn ist vorbei — »Passen« steht für Ada offen`
- ✓ Z31 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Passen' && !b.disabled)`
- ✓ Z32 `notiz Eine Karte wählen — die oberste von Adas Bibliothek`
- ✓ Z33 `pruefe (() => { const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => (x.getAttribute('aria-label') || x.textContent).includes('Bibliothek')); const unten = b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)); unten.dataset.ziel = 'bibliothek'; return true })()`
- ✓ Z34 `klicke [data-ziel="bibliothek"]`
- ✓ Z35 `wartetext Mit dieser Karte`
- ✓ Z36 `pause 500`
- ✓ Z37 `notiz Keine Formularwand: kein offenes Feld, „Mit dieser Karte" im Blick`
- ✓ Z39 `zeige (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement; const k = [...s.querySelectorAll('h3')].find((h) => h.textContent === 'Mit dieser Karte'); return 'Spalte ' + s.scrollHeight + ' hoch, sichtbar ' + s.clientHeight + ' · Eingabefelder ' + s.querySelectorAll('input,select,textarea').length + ' · „Mit dieser Karte" bei ' + Math.round(k.getBoundingClientRect().top - s.getBoundingClientRect().top) })()`
  ```
  "Spalte 1205 hoch, sichtbar 748 · Eingabefelder 0 · „Mit dieser Karte\" bei 573"
  ```
- ✓ Z40 `pruefe document.querySelector('[aria-label="Der Zug"]').parentElement.querySelectorAll('input,select,textarea').length === 0`
- ✓ Z41 `pruefe (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement; const k = [...s.querySelectorAll('h3')].find((h) => h.textContent === 'Mit dieser Karte'); return k.getBoundingClientRect().bottom <= s.getBoundingClientRect().bottom })()`
- ✓ Z42 `notiz Anhalten, Beenden und der Verlauf stehen unter dem Tisch und nicht in der Spalte`
- ✓ Z43 `pruefe (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement; const knopf = [...document.querySelectorAll('button')].find((b) => b.textContent === 'Anhalten'); return !!knopf && !s.contains(knopf) && !!(document.querySelector('[role=application]').compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING) })()`
- ✓ Z44 `abzug spalte`
  ```
  Bild: spalte.png
  ```
- ✓ Z49 `notiz Eine Karte auf dem Schlachtfeld wählen — über ihr Menü, »Weitere Handlungen …«`
- ✓ Z50 `pruefe (() => { const b = [...document.querySelectorAll('[role=application] li button')].find((x) => (x.getAttribute('aria-label') || x.textContent).includes('· Schlachtfeld')); const r = b.getBoundingClientRect(); b.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true, clientX: r.left + r.width / 2, clientY: r.top + r.height / 2 })); return true })()`
- ✓ Z51 `bis [...document.querySelectorAll('[role=menuitem]')].some((e) => e.checkVisibility() && e.innerText.includes('Weitere Handlungen'))`
- ✓ Z52 `pruefe ([...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes('Weitere Handlungen')).dataset.ziel = 'weitere', true)`
- ✓ Z53 `klicke [data-ziel="weitere"]`
- ✓ Z54 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Marken…')`
- ✓ Z55 `notiz „Marken…" klappt seine Felder auf und sagt es`
- ✓ Z56 `klicketext Marken…`
- ✓ Z57 `bis document.querySelector('[role=group][aria-label="Marken"]') !== null`
- ✓ Z58 `gleich [...document.querySelectorAll('button')].find((b) => b.innerText === 'Marken…').getAttribute('aria-expanded') => true`
- ✓ Z62 `zeige (() => { const k = [...document.querySelectorAll('button')].find((b) => b.innerText === 'Marken…').getBoundingClientRect(); const f = document.querySelector('[role=group][aria-label="Marken"]').getBoundingClientRect(); return 'Knopf ' + Math.round(k.top) + '–' + Math.round(k.bottom) + ', Felder ab ' + Math.round(f.top) })()`
  ```
  "Knopf 435–465, Felder ab 473"
  ```
- ✓ Z63 `pruefe (() => { const k = [...document.querySelectorAll('button')].find((b) => b.innerText === 'Marken…').getBoundingClientRect(); const f = document.querySelector('[role=group][aria-label="Marken"]').getBoundingClientRect(); return f.top >= k.bottom && f.top - k.bottom < 24 })()`
- ✓ Z64 `abzug marken-offen`
  ```
  Bild: marken-offen.png
  ```
- ✓ Z65 `notiz Nach dem Stellen klappt es zu, und der Fokus steht wieder auf „Marken…"`
- ✓ Z66 `klicketext Stellen`
- ✓ Z67 `bis document.querySelector('[role=group][aria-label="Marken"]') === null`
- ✓ Z68 `gleich document.activeElement.innerText + ' ' + document.activeElement.getAttribute('aria-expanded') => Marken… false`
- ✓ Z72 `bis [...document.querySelectorAll('[role=log] li')].some((l) => l.textContent.includes('Marken'))`
- ✓ Z73 `notiz „Schieben nach…" legt die Karte auf die Hand — gestellt mit der Vorgabe`
- ✓ Z74 `klicketext Schieben nach…`
- ✓ Z75 `bis document.querySelector('[role=group][aria-label="Schieben nach"]') !== null`
- ✓ Z76 `klicketext Stellen`
- ✓ Z77 `bis [...document.querySelectorAll('[role=application] li button')].some((b) => (b.getAttribute('aria-label') || b.textContent).includes('· Hand'))`
- ✓ Z78 `zeige [...document.querySelectorAll('[role=application] li button')].map((b) => b.getAttribute('aria-label') || b.textContent).filter((t) => t.includes('· Hand')).join(' | ')`
  ```
  "verdeckt · Hand | verdeckt · Hand | verdeckt · Hand | verdeckt · Hand | verdeckt · Hand | verdeckt · Hand | Langtext · Hand | Langtext · Hand | Langtext · Hand | Sonderlayout — ein Name, der so lang ist, dass er das Namensfeld sprengt · Hand | Sonderlayout — ein Name, der so lang ist, dass er das Namensfeld sprengt · Hand | Zweiseitig · Hand | Ohne Bild · Hand | Sonderzeichen ÄÖÜ ß «» ✦ ☙ 漢字 🃏 · Hand"
  ```
- ✓ Z79 `abzug nach-schieben`
  ```
  Bild: nach-schieben.png
  ```
- ✓ Z80 `notiz „Hineinsehen…" nennt die Karten der Bibliothek beim Namen, der Reihe nach (#1836)`
- ✓ Z84 `klicketext Hineinsehen…`
- ✓ Z85 `bis document.querySelector('[role=group][aria-label="Hineinsehen"]') !== null`
- ✓ Z86 `klicketext Stellen`
- ✓ Z87 `bis document.querySelector('ol[aria-labelledby]') !== null`
- ✓ Z88 `zeige (() => { const e = [...document.querySelectorAll('ol[aria-labelledby] li')].map((l) => l.textContent); return e.length + ' Karten: ' + e.slice(0, 3).join(' | ') + ' …' })()`
  ```
  "31 Karten: Ohne Bild | Ohne Bild | Ohne Bild …"
  ```
- ✓ Z89 `pruefe (() => { const e = [...document.querySelectorAll('ol[aria-labelledby] li')].map((l) => l.textContent); return e.length > 0 && e.every((t) => t !== '' && !/^p[0-9]+-[0-9]+$/.test(t)) })()`
- ✓ Z90 `pruefe ![...document.querySelectorAll('li')].some((l) => l.textContent.includes('"karten"'))`
- ✓ Z93 `pause 800`
- ✓ Z94 `zeige (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement.getBoundingClientRect(); const e = document.querySelector('ol[aria-labelledby] li').getBoundingClientRect(); const t = document.querySelector('[role=application]').getBoundingClientRect(); return 'Spalte ' + Math.round(s.top) + '–' + Math.round(s.bottom) + ', erste Karte ' + Math.round(e.top) + '–' + Math.round(e.bottom) + ', Tisch ' + Math.round(t.top) + '–' + Math.round(t.bottom) + ', Fenster ' + innerHeight })()`
  ```
  "Spalte 76–824, erste Karte 148–168, Tisch 76–736, Fenster 900"
  ```
- ✓ Z95 `pruefe (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement.getBoundingClientRect(); const e = document.querySelector('ol[aria-labelledby] li').getBoundingClientRect(); return e.top >= s.top - 1 && e.bottom <= s.bottom + 1 && e.bottom <= innerHeight })()`
- ✓ Z96 `pruefe (() => { const t = document.querySelector('[role=application]').getBoundingClientRect(); return t.top < innerHeight / 2 && t.bottom > innerHeight / 2 })()`
- ✓ Z99 `pruefe (() => { const s = document.querySelector('[aria-label="Der Zug"]').parentElement.getBoundingClientRect(); const e = document.querySelector('ol[aria-labelledby] li').getBoundingClientRect(); return e.left - s.left >= 32 })()`
- ✓ Z100 `abzug hineinsehen`
  ```
  Bild: hineinsehen.png
  ```
- ✓ Z101 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z102 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 70 Schritte.
