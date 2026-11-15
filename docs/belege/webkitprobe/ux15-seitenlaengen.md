# WebKit-Probe — ux15-seitenlaengen.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:54:01Z |

- ✓ Z17 `motor wk2`
- ✓ Z18 `lade /`
- ✓ Z19 `frist 20000`
- ✓ Z20 `warte nav`
- ✓ Z21 `wartetext Erste Schritte`
- ✓ Z22 `taste escape`
- ✓ Z23 `pause 500`
- ✓ Z25 `notiz Start`
- ✓ Z26 `taste cmd+k`
- ✓ Z27 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z28 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Start').dataset.ziel = 'seite', true)`
- ✓ Z29 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z30 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z31 `pause 1200`
- ✓ Z32 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "MagicSetEditor4Mac · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z33 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z34 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z35 `abzug seite-01`
  ```
  Bild: seite-01.png
  ```
- ✓ Z36 `notiz Die Startseite zeigt fünf Hinweise und den Rest auf Wunsch`
- ✓ Z37 `pruefe (() => { const b = document.querySelector('section[aria-label="Was Aufmerksamkeit braucht"]'); if (!b) return true; const knoepfe = b.querySelectorAll('ul > li > button').length; return knoepfe <= 5 })()`
- ✓ Z39 `notiz Sets`
- ✓ Z40 `taste cmd+k`
- ✓ Z41 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z42 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Sets').dataset.ziel = 'seite', true)`
- ✓ Z43 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z44 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z45 `pause 1200`
- ✓ Z46 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Sets · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z47 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z48 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z49 `abzug seite-02`
  ```
  Bild: seite-02.png
  ```
- ✓ Z51 `notiz Pakete`
- ✓ Z52 `taste cmd+k`
- ✓ Z53 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z54 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Pakete').dataset.ziel = 'seite', true)`
- ✓ Z55 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z56 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z57 `pause 1200`
- ✓ Z58 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Pakete · 1111 hoch, 81 % im ersten Blick"
  ```
- ✓ Z59 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z60 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z61 `abzug seite-03`
  ```
  Bild: seite-03.png
  ```
- ✓ Z63 `notiz Vorlagen`
- ✓ Z64 `taste cmd+k`
- ✓ Z65 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z66 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Vorlagen').dataset.ziel = 'seite', true)`
- ✓ Z67 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z68 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z69 `pause 1200`
- ✓ Z70 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Vorlagen · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z71 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: [\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\",\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\",\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\",\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\",\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\",\"Mitgeliefert — nur lesbar. Eine eigene Kopie lässt sich bearbeiten.\"]"
  ```
- ✓ Z72 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z73 `abzug seite-04`
  ```
  Bild: seite-04.png
  ```
- ✓ Z74 `notiz Die Werkstatt reicht bis an den Fensterrand, der Baum rollt in seiner Spalte`
- ✓ Z75 `gleich document.scrollingElement.scrollHeight <= innerHeight + 1 => true`
- ✓ Z76 `zeige (() => { const a = document.querySelector('aside[aria-label="Pakete und Dateien"]'); return 'Baum: ' + a.clientHeight + ' hoch sichtbar, ' + a.scrollHeight + ' mit Rollen' })()`
  ```
  "Baum: 658 hoch sichtbar, 1279 mit Rollen"
  ```
- ✓ Z77 `gleich (() => { const t = 'Wähle links im Baum'; return [...document.querySelectorAll('main p')].filter((p) => p.textContent.startsWith(t)).length })() => 1`
- ✓ Z79 `notiz Sammlung`
- ✓ Z80 `taste cmd+k`
- ✓ Z81 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z82 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Sammlung').dataset.ziel = 'seite', true)`
- ✓ Z83 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z84 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z85 `pause 1200`
- ✓ Z86 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Sammlung · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z87 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z88 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z89 `abzug seite-05`
  ```
  Bild: seite-05.png
  ```
- ✓ Z91 `notiz Decks`
- ✓ Z92 `taste cmd+k`
- ✓ Z93 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z94 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Decks').dataset.ziel = 'seite', true)`
- ✓ Z95 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z96 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z97 `pause 1200`
- ✓ Z98 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Decks · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z99 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z100 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z101 `abzug seite-06`
  ```
  Bild: seite-06.png
  ```
- ✓ Z102 `notiz Ohne offenes Deck steht der Satz einmal, im Zentrum`
- ✓ Z103 `gleich [...document.querySelectorAll('main p')].filter((p) => p.textContent.startsWith('Kein Deck aufgeschlagen')).length => 1`
- ✓ Z105 `notiz Turniere`
- ✓ Z106 `taste cmd+k`
- ✓ Z107 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z108 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Turniere').dataset.ziel = 'seite', true)`
- ✓ Z109 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z110 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z111 `pause 1200`
- ✓ Z112 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Turniere · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z113 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z114 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z115 `abzug seite-07`
  ```
  Bild: seite-07.png
  ```
- ✓ Z117 `notiz Spielen`
- ✓ Z118 `taste cmd+k`
- ✓ Z119 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z120 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Spielen').dataset.ziel = 'seite', true)`
- ✓ Z121 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z122 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z123 `pause 1200`
- ✓ Z124 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Eine Partie anfangen · 1237 hoch, 73 % im ersten Blick"
  ```
- ✓ Z125 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z126 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z127 `abzug seite-08`
  ```
  Bild: seite-08.png
  ```
- ✓ Z129 `notiz Handeln`
- ✓ Z130 `taste cmd+k`
- ✓ Z131 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z132 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Handeln').dataset.ziel = 'seite', true)`
- ✓ Z133 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z134 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z135 `pause 1200`
- ✓ Z136 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Handeln · 1256 hoch, 72 % im ersten Blick"
  ```
- ✓ Z137 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z138 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z139 `abzug seite-09`
  ```
  Bild: seite-09.png
  ```
- ✓ Z141 `notiz Hilfe`
- ✓ Z142 `taste cmd+k`
- ✓ Z143 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z144 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Hilfe').dataset.ziel = 'seite', true)`
- ✓ Z145 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z146 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z147 `pause 1200`
- ✓ Z148 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Hilfe · 1222 hoch, 74 % im ersten Blick"
  ```
- ✓ Z149 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z150 `abzug seite-10`
  ```
  Bild: seite-10.png
  ```
- ✓ Z152 `notiz Einstellungen`
- ✓ Z153 `taste cmd+k`
- ✓ Z154 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z155 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Einstellungen').dataset.ziel = 'seite', true)`
- ✓ Z156 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z157 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z158 `pause 1200`
- ✓ Z159 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Einstellungen · 901 hoch, 100 % im ersten Blick"
  ```
- ✓ Z160 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z161 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z162 `abzug seite-11`
  ```
  Bild: seite-11.png
  ```
- ✓ Z164 `notiz Nutzer und Profile`
- ✓ Z165 `taste cmd+k`
- ✓ Z166 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z167 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Nutzer und Profile').dataset.ziel = 'seite', true)`
- ✓ Z168 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z169 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z170 `pause 1200`
- ✓ Z171 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Nutzer und Profile · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z172 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z173 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z174 `abzug seite-12`
  ```
  Bild: seite-12.png
  ```
- ✓ Z176 `notiz Ohne offenes Set: die Karten sagen, wohin, und der Knopf führt hin`
- ✓ Z177 `taste cmd+k`
- ✓ Z178 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z179 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Karten').dataset.ziel = 'seite', true)`
- ✓ Z180 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z181 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z182 `wartetext Kein Set offen`
- ✓ Z183 `abzug ohne-set`
  ```
  Bild: ohne-set.png
  ```
- ✓ Z184 `klicketext Zu den Sets`
- ✓ Z185 `wartetext Werkstatt-Testkarten`
- ✓ Z186 `klicketext Werkstatt-Testkarten`
- ✓ Z187 `wartetext Wie viel du siehst`
- ✓ Z188 `pause 1500`
- ✓ Z190 `notiz Karten`
- ✓ Z191 `taste cmd+k`
- ✓ Z192 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z193 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Karten').dataset.ziel = 'seite', true)`
- ✓ Z194 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z195 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z196 `pause 1200`
- ✓ Z197 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z198 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z199 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z200 `abzug seite-13`
  ```
  Bild: seite-13.png
  ```
- ✓ Z201 `notiz Bei 1.440 × 900 stehen alle drei Spalten, und die Kartenliste rollt in sich`
- ✓ Z202 `gleich !!document.querySelector('aside[aria-label="Karten des Sets"]') => true`
- ✓ Z203 `zeige (() => { const r = document.querySelector('[role=region][aria-label="Kartenarbeitsplatz"]'); const a = document.querySelector('aside[aria-label="Karten des Sets"]'); return 'Arbeitsplatz ' + Math.round(r.getBoundingClientRect().width) + ' breit; Navigator ' + a.clientHeight + ' hoch, rollt selbst ' + (a.scrollHeight - a.clientHeight) + ' Punkte' })()`
  ```
  "Arbeitsplatz 1095 breit; Navigator 525 hoch, rollt selbst 0 Punkte"
  ```
- ✓ Z204 `gleich (() => { const a = document.querySelector('aside[aria-label="Karten des Sets"]'); return a.scrollHeight <= a.clientHeight + 1 })() => true`
- ✓ Z206 `notiz Bilder`
- ✓ Z207 `taste cmd+k`
- ✓ Z208 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z209 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Bilder').dataset.ziel = 'seite', true)`
- ✓ Z210 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z211 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z212 `pause 1200`
- ✓ Z213 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 1209 hoch, 74 % im ersten Blick"
  ```
- ✓ Z214 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z215 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z216 `abzug seite-14`
  ```
  Bild: seite-14.png
  ```
- ✓ Z218 `notiz Suche`
- ✓ Z219 `taste cmd+k`
- ✓ Z220 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z221 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Suche').dataset.ziel = 'seite', true)`
- ✓ Z222 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z223 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z224 `pause 1200`
- ✓ Z225 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z226 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z227 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z228 `abzug seite-15`
  ```
  Bild: seite-15.png
  ```
- ✓ Z230 `notiz Ersetzen`
- ✓ Z231 `taste cmd+k`
- ✓ Z232 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z233 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Ersetzen').dataset.ziel = 'seite', true)`
- ✓ Z234 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z235 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z236 `pause 1200`
- ✓ Z237 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 900 hoch, 100 % im ersten Blick"
  ```
- ✓ Z238 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z239 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z240 `abzug seite-16`
  ```
  Bild: seite-16.png
  ```
- ✓ Z242 `notiz Stil`
- ✓ Z243 `taste cmd+k`
- ✓ Z244 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z245 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Stil').dataset.ziel = 'seite', true)`
- ✓ Z246 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z247 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z248 `pause 1200`
- ✓ Z249 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Eigenschaften des Sets · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z250 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z251 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z252 `abzug seite-17`
  ```
  Bild: seite-17.png
  ```
- ✓ Z253 `taste escape`
- ✓ Z254 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z256 `notiz Setangaben`
- ✓ Z257 `taste cmd+k`
- ✓ Z258 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z259 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Setangaben').dataset.ziel = 'seite', true)`
- ✓ Z260 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z261 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z262 `pause 1200`
- ✓ Z263 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Eigenschaften des Sets · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z264 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z265 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z266 `abzug seite-18`
  ```
  Bild: seite-18.png
  ```
- ✓ Z267 `taste escape`
- ✓ Z268 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z270 `notiz Symbol`
- ✓ Z271 `taste cmd+k`
- ✓ Z272 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z273 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Symbol').dataset.ziel = 'seite', true)`
- ✓ Z274 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z275 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z276 `pause 1200`
- ✓ Z277 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Eigenschaften des Sets · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z278 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z279 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z280 `abzug seite-19`
  ```
  Bild: seite-19.png
  ```
- ✓ Z281 `taste escape`
- ✓ Z282 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z284 `notiz Schlüsselwörter`
- ✓ Z285 `taste cmd+k`
- ✓ Z286 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z287 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Schlüsselwörter').dataset.ziel = 'seite', true)`
- ✓ Z288 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z289 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z290 `pause 1200`
- ✓ Z291 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Eigenschaften des Sets · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z292 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z293 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z294 `abzug seite-20`
  ```
  Bild: seite-20.png
  ```
- ✓ Z295 `taste escape`
- ✓ Z296 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z298 `notiz Statistik`
- ✓ Z299 `taste cmd+k`
- ✓ Z300 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z301 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Statistik').dataset.ziel = 'seite', true)`
- ✓ Z302 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z303 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z304 `pause 1200`
- ✓ Z305 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Werkzeuge · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z306 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z307 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z308 `abzug seite-21`
  ```
  Bild: seite-21.png
  ```
- ✓ Z309 `taste escape`
- ✓ Z310 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z312 `notiz Zufallspack`
- ✓ Z313 `taste cmd+k`
- ✓ Z314 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z315 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Zufallspack').dataset.ziel = 'seite', true)`
- ✓ Z316 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z317 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z318 `pause 1200`
- ✓ Z319 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Werkzeuge · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z320 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z321 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z322 `abzug seite-22`
  ```
  Bild: seite-22.png
  ```
- ✓ Z323 `taste escape`
- ✓ Z324 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z326 `notiz Konsole`
- ✓ Z327 `taste cmd+k`
- ✓ Z328 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z329 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Konsole').dataset.ziel = 'seite', true)`
- ✓ Z330 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z331 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z332 `pause 1200`
- ✓ Z333 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · Schublade: Werkzeuge · 1234 hoch, 73 % im ersten Blick"
  ```
- ✓ Z334 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z335 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z336 `abzug seite-23`
  ```
  Bild: seite-23.png
  ```
- ✓ Z337 `taste escape`
- ✓ Z338 `bis ![...document.querySelectorAll('[role=dialog]')].some((e) => e.getClientRects().length > 0)`
- ✓ Z340 `notiz Ausgeben`
- ✓ Z341 `taste cmd+k`
- ✓ Z342 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z343 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Ausgeben').dataset.ziel = 'seite', true)`
- ✓ Z344 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z345 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z346 `pause 1200`
- ✓ Z347 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 1209 hoch, 74 % im ersten Blick"
  ```
- ✓ Z348 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z349 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z350 `abzug seite-24`
  ```
  Bild: seite-24.png
  ```
- ✓ Z352 `notiz Druckbogen`
- ✓ Z353 `taste cmd+k`
- ✓ Z354 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z355 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Druckbogen').dataset.ziel = 'seite', true)`
- ✓ Z356 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z357 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z358 `pause 1200`
- ✓ Z359 `zeige (() => { const m = document.querySelector('main'); const h = [...m.querySelectorAll('h1')].find((e) => e.getClientRects().length > 0); const s = [...document.querySelectorAll('[role=dialog] h2, [role=dialog] h1')].find((e) => e.getClientRects().length > 0); return (h ? h.textContent.trim() : '—') + (s ? ' · Schublade: ' + s.textContent.trim() : '') + ' · ' + document.scrollingElement.scrollHeight + ' hoch, ' + Math.round(Math.min(1, innerHeight / document.scrollingElement.scrollHeight) * 100) + ' % im ersten Blick' })()`
  ```
  "Werkstatt-Testkarten · 1158 hoch, 78 % im ersten Blick"
  ```
- ✓ Z360 `zeige 'lang: ' + JSON.stringify((() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().map((x) => x.slice(0, 60) + '… (' + x.length + ')')) + ' · doppelt: ' + JSON.stringify((() => { const t = [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null).map((p) => p.innerText.replace(/\s+/g, ' ').trim()); return t.filter((x, i) => x.length > 30 && t.indexOf(x) !== i) })())`
  ```
  "lang: [] · doppelt: []"
  ```
- ✓ Z361 `gleich (() => { const beschrieben = new Set([...document.querySelectorAll('[aria-describedby]')].flatMap((e) => e.getAttribute('aria-describedby').split(/\s+/))); return [...document.querySelectorAll('main p, [role=dialog] p')].filter((p) => p.offsetParent !== null && !beschrieben.has(p.id)).map((p) => p.innerText.replace(/\s+/g, ' ').trim()).filter((x) => x.length > 140) })().length => 0`
- ✓ Z362 `abzug seite-25`
  ```
  Bild: seite-25.png
  ```
- ✓ Z364 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z365 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 322 Schritte.
