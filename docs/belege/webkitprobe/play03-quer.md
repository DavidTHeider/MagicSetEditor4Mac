# WebKit-Probe — play03-quer.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:34Z |

- ✓ Z10 `motor wk2`
- ✓ Z11 `lade /`
- ✓ Z12 `frist 20000`
- ✓ Z13 `warte nav`
- ✓ Z14 `wartetext Erste Schritte`
- ✓ Z15 `taste escape`
- ✓ Z16 `taste cmd+4`
- ✓ Z17 `wartetext Partie anfangen`
- ✓ Z18 `klicketext Partie anfangen`
- ✓ Z19 `wartetext Weiterspielen`
- ✓ Z20 `klicketext Weiterspielen`
- ✓ Z21 `klicketext Weiterspielen als Ada`
- ✓ Z22 `warte canvas`
- ✓ Z23 `pause 1500`
- ✓ Z24 `notiz Der Beginn ist vorbei — `play-beginn` hat ihn in derselben Partie zu Ende gebracht; vorher nähme der Tisch nur ihn an (CR 103.8, #1851)`
- ✓ Z31 `notiz Vorbedingung: der Beginn ist vorbei — »Passen« steht für Ada offen`
- ✓ Z32 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Passen' && !b.disabled)`
- ✓ Z33 `notiz Der Zug zeigt als Balken, wie weit er ist — mit dem Abschnitt als Namen`
- ✓ Z34 `gleich (() => { const b = document.querySelector('[role=progressbar][aria-labelledby]'); return b.getAttribute('aria-valuenow') + '/' + b.getAttribute('aria-valuemax') + ' ' + document.getElementById(b.getAttribute('aria-labelledby')).textContent })() => 2/12 Anfangsphase · Versorgung`
- ✓ Z35 `zeige (() => { const b = document.querySelector('[role=progressbar][aria-labelledby]'); const farbe = (e) => getComputedStyle(e).backgroundColor; return 'Füllung ' + farbe(b.firstElementChild) + ' auf Bahn ' + farbe(b) + ' im Kasten ' + farbe(b.closest('section')) })()`
  ```
  "Füllung rgb(107, 68, 35) auf Bahn rgb(232, 220, 194) im Kasten rgb(253, 249, 240)"
  ```
- ✓ Z36 `notiz Unter den Handlungen steht, warum einer fehlt`
- ✓ Z37 `wartetext Angeboten wird nur, was gerade geht`
- ✓ Z38 `notiz Die oberste Karte aus Adas Bibliothek aufs Schlachtfeld und tappen`
- ✓ Z39 `pruefe (() => { const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => (x.getAttribute('aria-label') || x.textContent).includes('Bibliothek')); const unten = b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)); unten.dataset.ziel = 'bibliothek'; return true })()`
- ✓ Z40 `klicke [data-ziel="bibliothek"]`
- ✓ Z41 `wartetext Mit dieser Karte`
- ✓ Z42 `klicketext Schieben nach…`
- ✓ Z43 `bis document.querySelector('[role=group][aria-label="Schieben nach"]') !== null`
- ✓ Z44 `pruefe (() => { const s = document.querySelector('[role=group][aria-label="Schieben nach"] select'); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, 'tisch'); s.dispatchEvent(new Event('change', { bubbles: true })); return true })()`
- ✓ Z45 `klicketext Stellen`
- ✓ Z46 `bis [...document.querySelectorAll('[role=application] li button')].some((b) => (b.getAttribute('aria-label') || b.textContent).includes('· Schlachtfeld'))`
- ✓ Z47 `pruefe ([...document.querySelectorAll('[role=application] li button')].find((b) => (b.getAttribute('aria-label') || b.textContent).includes('· Schlachtfeld')).dataset.ziel = 'schlachtfeld', true)`
- ✓ Z51 `notiz Ein Klick auf die Karte tappt sie (#2372)`
- ✓ Z52 `klicke [data-ziel="schlachtfeld"]`
- ✓ Z53 `bis [...document.querySelectorAll('[role=application] li button')].some((b) => /Schlachtfeld.*getappt/.test(b.getAttribute('aria-label') || b.textContent))`
- ✓ Z54 `zeige [...document.querySelectorAll('[role=application] li button')].map((b) => b.getAttribute('aria-label') || b.textContent).filter((t) => t.includes('Schlachtfeld')).join(' | ')`
  ```
  "Langtext · Schlachtfeld · getappt"
  ```
- ✓ Z55 `pause 1500`
- ✓ Z56 `abzug quer`
  ```
  Bild: quer.png
  ```
- ✓ Z57 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z58 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 40 Schritte.
