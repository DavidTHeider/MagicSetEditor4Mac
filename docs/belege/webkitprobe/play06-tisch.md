# WebKit-Probe — play06-tisch.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:11Z |

- ✓ Z7 `motor wk2`
- ✓ Z8 `lade /`
- ✓ Z9 `frist 20000`
- ✓ Z10 `warte nav`
- ✓ Z11 `wartetext Erste Schritte`
- ✓ Z12 `notiz Wie lange die Oberfläche braucht (PLAY-06.AK-03)`
- ✓ Z16 `zeige 'Oberfläche bereit nach ' + Math.round(performance.now()) + ' ms'`
  ```
  "Oberfläche bereit nach 177 ms"
  ```
- ✓ Z17 `taste escape`
- ✓ Z18 `taste cmd+4`
- ✓ Z19 `wartetext Partie anfangen`
- ✓ Z20 `klicketext Partie anfangen`
- ✓ Z21 `wartetext Weiterspielen`
- ✓ Z22 `klicketext Weiterspielen`
- ✓ Z23 `pruefe (window.__tischstart = performance.now(), true)`
- ✓ Z24 `klicketext Weiterspielen als Ada`
- ✓ Z25 `warte canvas`
- ✓ Z26 `bis document.querySelectorAll('[role=application] li button').length > 0`
- ✓ Z27 `zeige 'Tisch mit seinen Knöpfen nach ' + Math.round(performance.now() - window.__tischstart) + ' ms'`
  ```
  "Tisch mit seinen Knöpfen nach 215 ms"
  ```
- ✓ Z28 `pause 1500`
- ✓ Z29 `notiz Der Tisch steht im Blick, der Fokus am neuen Arbeitsplatz (F09)`
- ✓ Z33 `gleich document.activeElement.tagName + ' ' + document.activeElement.textContent => H1 Am Tisch`
- ✓ Z34 `gleich window.scrollY => 0`
- ✓ Z35 `pruefe document.querySelector('[role=application]').getBoundingClientRect().bottom <= innerHeight`
- ✓ Z36 `notiz Die Brotkrume nennt die Partie (UX-03.AK-02, #1675)`
- ✓ Z41 `gleich [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ') => Spielen › Partie gegen Bo › Am Tisch`
- ✓ Z42 `gleich document.querySelectorAll('nav[aria-label="Wo du bist"] [aria-current=page]').length => 1`
- ✓ Z43 `notiz WebGL 2 über die GPU, der Kontext hält`
- ✓ Z44 `zeige (() => { const g = document.querySelector('canvas').getContext('webgl2'); const i = g && g.getExtension('WEBGL_debug_renderer_info'); return g ? g.getParameter(g.VERSION) + ' | ' + (i ? g.getParameter(i.UNMASKED_RENDERER_WEBGL) : '?') : 'kein webgl2' })()`
  ```
  "WebGL 2.0 | Apple GPU"
  ```
- ✓ Z45 `pruefe (() => { const g = document.querySelector('canvas').getContext('webgl2'); return !!g && !g.isContextLost() })()`
- ✓ Z46 `notiz Mindestens fünfzig Bilder je Sekunde, auch außerhalb des Bildschirms`
- ✓ Z47 `pruefe (window.__bilder = 0, (function zaehle() { window.__bilder++; requestAnimationFrame(zaehle) })(), true)`
- ✓ Z48 `pause 2000`
- ✓ Z49 `zeige window.__bilder + ' Bilder in 2 s'`
  ```
  "129 Bilder in 2 s"
  ```
- ✓ Z50 `pruefe window.__bilder >= 100`
- ✓ Z51 `notiz Der Tisch füllt seine Fläche, auch bei doppelter Dichte`
- ✓ Z57 `zeige (() => { const f = document.querySelector('[role=application]').getBoundingClientRect(); const k = [...document.querySelectorAll('[role=application] li')].map((l) => l.getBoundingClientRect()); return devicePixelRatio + '-fache Dichte, Fläche ' + Math.round(f.width) + ' × ' + Math.round(f.height) + ', Knöpfe bis ' + Math.round(Math.max(...k.map((b) => b.right)) - f.left) + ' × ' + Math.round(Math.max(...k.map((b) => b.bottom)) - f.top) })()`
  ```
  "2-fache Dichte, Fläche 759 × 660, Knöpfe bis 610 × 575"
  ```
- ✓ Z58 `pruefe (() => { const f = document.querySelector('[role=application]').getBoundingClientRect(); const k = [...document.querySelectorAll('[role=application] li')].map((l) => l.getBoundingClientRect()); return Math.max(...k.map((b) => b.right)) - f.left > 0.6 * f.width || Math.max(...k.map((b) => b.bottom)) - f.top > 0.6 * f.height })()`
- ✓ Z59 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z60 `gleich JSON.stringify(window.__probeFehler || []) => []`
- ✓ Z61 `abzug tisch`
  ```
  Bild: tisch.png
  ```
- ✓ Z62 `notiz Im schmalen Fenster steht der Zug unter dem Tisch — und trotzdem im Blick`
- ✓ Z74 `pruefe (window.__groesse = performance.now(), window.__passend = undefined, new ResizeObserver(() => { if (window.__passend === undefined) window.__passend = performance.now() }).observe(document.querySelector('[role=application] canvas')), true)`
- ✓ Z75 `groesse 1024 700`
  ```
  1024 × 700 Punkte
  ```
- ✓ Z76 `zeige window.__passend === undefined ? 'die Leinwand folgte nicht' : 'Leinwand folgte nach ' + Math.round(window.__passend - window.__groesse) + ' ms'`
  ```
  "Leinwand folgte nach 1 ms"
  ```
- ✓ Z77 `pruefe window.__passend !== undefined`
- ✓ Z78 `zeige (() => { const t = document.querySelector('[role=application]').getBoundingClientRect(); const z = document.querySelector('[aria-label="Der Zug"]').getBoundingClientRect(); return 'Tisch ' + Math.round(t.top) + '–' + Math.round(t.bottom) + ', Zug ' + Math.round(z.top) + '–' + Math.round(z.bottom) + ', Fenster ' + innerHeight })()`
  ```
  "Tisch 224–446, Zug 528–690, Fenster 700"
  ```
- ✓ Z79 `zeige (() => { const k = [...document.querySelectorAll('[aria-label="Ansicht des Tisches"], [aria-label="Springen zu"], [data-tischleiste="schmal"]')].flatMap((g) => [...g.querySelectorAll('button')]).filter((b) => b.checkVisibility() && b.getBoundingClientRect().width > 0); return k.length + ' Knöpfe in der Leiste, oben bei ' + [...new Set(k.map((b) => Math.round(b.getBoundingClientRect().top)))].join(', ') })()`
  ```
  "5 Knöpfe in der Leiste, oben bei 461"
  ```
- ✓ Z80 `zeige (() => { const l = document.querySelector('[data-tischleiste]')?.closest('.rounded-sm'); const b = l?.getBoundingClientRect(); return b ? 'Leiste ' + Math.round(b.top) + '–' + Math.round(b.bottom) : 'keine Leiste' })()`
  ```
  "Leiste 454–496"
  ```
- ✓ Z81 `abzug tisch-1024`
  ```
  Bild: tisch-1024.png
  ```
- ✓ Z82 `pruefe (() => { const t = document.querySelector('[role=application]').getBoundingClientRect(); const z = document.querySelector('[aria-label="Der Zug"]').getBoundingClientRect(); return t.bottom <= z.top && z.bottom <= innerHeight && t.height >= 150 })()`
- ✓ Z86 `pruefe (() => { const k = [...document.querySelectorAll('[aria-label="Ansicht des Tisches"], [aria-label="Springen zu"], [data-tischleiste="schmal"]')].flatMap((g) => [...g.querySelectorAll('button')]).filter((b) => b.checkVisibility() && b.getBoundingClientRect().width > 0); return k.length === 5 && new Set(k.map((b) => Math.round(b.getBoundingClientRect().top))).size === 1 })()`
- ✓ Z87 `notiz Und der Tisch passt sich an — kein Knopf liegt außerhalb der Leinwand`
- ✓ Z91 `pruefe (() => { const f = document.querySelector('[role=application]').getBoundingClientRect(); return [...document.querySelectorAll('[role=application] li')].every((l) => { const b = l.getBoundingClientRect(); return b.left >= f.left - 1 && b.right <= f.right + 1 && b.top >= f.top - 1 && b.bottom <= f.bottom + 1 }) })()`
- ✓ Z92 `abzug tisch-schmal`
  ```
  Bild: tisch-schmal.png
  ```
- ✓ Z93 `notiz Was hinter „Darstellung" steht, bleibt erreichbar`
- ✓ Z94 `klicke [data-tischleiste="schmal"] button`
- ✓ Z95 `wartetext Tischbild ausgeben`
- ✓ Z96 `pruefe (() => { const g = [...document.querySelectorAll('[role=group][aria-labelledby]')].find((x) => document.getElementById(x.getAttribute('aria-labelledby'))?.textContent === 'Darstellung'); return !!g && !!g.querySelector('select') && [...g.querySelectorAll('button')].some((b) => b.textContent.trim() === 'Tischbild ausgeben') })()`
- ✓ Z97 `abzug tisch-darstellung`
  ```
  Bild: tisch-darstellung.png
  ```
- ✓ Z98 `taste escape`
- ✓ Z99 `pause 400`
- ✓ Z100 `pruefe document.querySelector('[data-tischleiste="schmal"] button').getAttribute('aria-expanded') === 'false' && document.activeElement === document.querySelector('[data-tischleiste="schmal"] button')`

**Ergebnis: bestanden** — 62 Schritte.
