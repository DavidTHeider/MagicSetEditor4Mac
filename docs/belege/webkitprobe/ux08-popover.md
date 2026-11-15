# WebKit-Probe — ux08-popover.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:52Z |

- ✓ Z13 `motor wk2`
- ✓ Z14 `lade /`
- ✓ Z15 `frist 20000`
- ✓ Z16 `warte nav`
- ✓ Z17 `wartetext Erste Schritte`
- ✓ Z18 `taste escape`
- ✓ Z19 `pruefe (window.__zur = (seite) => { const b = [...document.querySelectorAll('[role=dialog] button')].find((x) => x.getClientRects().length > 0 && x.innerText.trim() === seite); b.dataset.ziel = 'seite'; return true }, window.__popover = () => [...document.querySelectorAll('[role=group][aria-labelledby]')].find((g) => document.getElementById(g.getAttribute('aria-labelledby')).textContent === 'Probe-Deck umbenennen'), window.__umbenennen = () => { const b = [...document.querySelectorAll('button')].find((x) => x.checkVisibility() && x.getAttribute('aria-label') === 'Handlungen — Probe-Deck'); b.dataset.ziel = 'umbenennen'; return true }, window.__eintrag = (wort) => { const e = [...document.querySelectorAll('[role=menuitem]')].find((x) => x.checkVisibility() && x.innerText.includes(wort)); if (!e) return false; e.dataset.ziel = 'eintrag'; return true }, true)`
- ✓ Z20 `taste cmd+k`
- ✓ Z21 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z22 `pruefe window.__zur('Decks')`
- ✓ Z23 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z24 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z25 `wartetext Probe-Deck`
- ✓ Z29 `klicke button[aria-label="Öffnen: Probe-Deck"]`
- ✓ Z30 `warte button[aria-label="Handlungen — Probe-Deck"]`
- ✓ Z31 `notiz Umbenennen — über »…«; das Popover steht an ihm, der Fokus im Namen`
- ✓ Z32 `pruefe window.__umbenennen()`
- ✓ Z33 `klicke [data-ziel="umbenennen"]`
- ✓ Z34 `bis window.__eintrag('Umbenennen')`
- ✓ Z35 `klicke [data-ziel="eintrag"]`
- ✓ Z36 `bis window.__popover() !== undefined`
- ✓ Z37 `gleich document.activeElement.value => Probe-Deck`
- ✓ Z38 `gleich document.activeElement.labels[0].textContent.trim() => Neuer Name`
- ✓ Z41 `pruefe (() => { const p = window.__popover().getBoundingClientRect(); const k = document.querySelector('[data-ziel="umbenennen"]').getBoundingClientRect(); return p.top >= k.bottom - 1 || p.bottom <= k.top + 1 })()`
- ✓ Z42 `pruefe (() => { const g = window.__popover(); const r = g.getBoundingClientRect(); return r.right <= window.innerWidth && g.contains(document.elementFromPoint(r.right - 4, r.top + r.height / 2)) && g.contains(document.elementFromPoint(r.left + r.width / 2, r.bottom - 4)) })()`
- ✓ Z43 `zeige (() => { const r = window.__popover().getBoundingClientRect(); return Math.round(r.left) + ',' + Math.round(r.top) + ' ' + Math.round(r.width) + '×' + Math.round(r.height) })()`
  ```
  "385,472 303×190"
  ```
- ✓ Z44 `notiz Der Titel trägt die Textfarbe des Inhaltsbereichs — im Portal erbt er sie nicht von selbst`
- ✓ Z45 `gleich getComputedStyle(document.getElementById(window.__popover().getAttribute('aria-labelledby'))).color === getComputedStyle(document.getElementById('inhalt')).color => true`
- ✓ Z46 `abzug popover`
  ```
  Bild: popover.png
  ```
- ✓ Z47 `notiz Escape schließt, der Fokus kehrt zum Auslöser zurück`
- ✓ Z48 `taste escape`
- ✓ Z49 `bis window.__popover() === undefined`
- ✓ Z53 `bis document.activeElement.getAttribute('aria-label') === 'Handlungen — Probe-Deck'`
- ✓ Z54 `pruefe (window.__nachEscape = (() => { const e = document.activeElement; return e.tagName + (e.id ? '#' + e.id : '') + ' ' + (e.getAttribute('aria-label') || e.innerText || '').trim().slice(0, 50) })(), true)`
- ✓ Z55 `zeige window.__nachEscape`
  ```
  "BUTTON Handlungen — Probe-Deck"
  ```
- ✓ Z57 `gleich document.activeElement.getAttribute('aria-label') => Handlungen — Probe-Deck`
- ✓ Z58 `notiz Ein Klick daneben schließt es auch — und der Fokus bleibt, wo geklickt wurde`
- ✓ Z59 `pruefe window.__umbenennen()`
- ✓ Z60 `klicke [data-ziel="umbenennen"]`
- ✓ Z61 `bis window.__eintrag('Umbenennen')`
- ✓ Z62 `klicke [data-ziel="eintrag"]`
- ✓ Z63 `bis window.__popover() !== undefined`
- ✓ Z64 `pruefe ([...document.querySelectorAll('main#inhalt h1, main#inhalt h2')].find((h) => h.checkVisibility()).dataset.ziel = 'daneben', true)`
- ✓ Z65 `klicke [data-ziel="daneben"]`
- ✓ Z66 `bis window.__popover() === undefined`
- ✓ Z67 `pruefe document.activeElement.getAttribute('aria-label') !== 'Handlungen — Probe-Deck'`
- ✓ Z68 `wartetext Probe-Deck`
- ✓ Z69 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 48 Schritte.
