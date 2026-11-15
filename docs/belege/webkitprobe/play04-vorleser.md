# WebKit-Probe — play04-vorleser.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk1 — WebCore im eigenen Prozess |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:40Z |

- ✓ Z10 `motor wk1`
- ✓ Z11 `lade /`
- ✓ Z12 `frist 20000`
- ✓ Z13 `warte nav`
- ✓ Z14 `wartetext Erste Schritte`
- ✓ Z15 `fokus`
  ```
  Ansage: (nichts fokussiert) — DOM: "body "
  ```
- ✓ Z16 `taste escape`
- ✓ Z17 `fokus`
  ```
  Ansage: (nichts fokussiert) — DOM: "body "
  ```
- ✓ Z18 `taste cmd+4`
- ✓ Z19 `wartetext Partie anfangen`
- ✓ Z20 `klicketext Partie anfangen`
- ✓ Z21 `wartetext Weiterspielen`
- ✓ Z22 `klicketext Weiterspielen`
- ✓ Z23 `klicketext Weiterspielen als Ada`
- ✓ Z24 `wartetext Zuletzt geschehen`
- ✓ Z25 `pause 1500`
- ✓ Z26 `notiz Der Beginn ist vorbei — `play-beginn` hat ihn in derselben Partie zu Ende gebracht; vorher nähme der Tisch nur ihn an (CR 103.8, #1851)`
- ✓ Z33 `notiz Vorbedingung: der Beginn ist vorbei — »Passen« steht für Ada offen`
- ✓ Z34 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Passen' && !b.disabled)`
- ✓ Z35 `notiz Ein Ereignis — Ada wird Monarch`
- ✓ Z38 `klicketext Monarch werden`
- ✓ Z39 `bis (() => { const h = [...document.querySelectorAll('h2')].find((x) => x.textContent === 'Zuletzt geschehen'); return !!h && h.parentElement.querySelectorAll('li').length > 0 })()`
- ✓ Z40 `notiz Die Ereignisliste im Zugänglichkeitsbaum`
- ✓ Z43 `pruefe ([...document.querySelectorAll('h2')].find((x) => x.textContent === 'Zuletzt geschehen').parentElement.querySelector('ul').dataset.ziel = 'ereignisse', true)`
- ✓ Z44 `baum [data-ziel="ereignisse"]`
  ```
  AXStaticText «» · text · Wert «Ada: Monarch werden»
  ```
- ✓ Z45 `notiz Und die Karten am Tisch tragen ihren Stand in Worten`
- ✓ Z46 `zeige [...document.querySelectorAll('[role=application] li button')].slice(0, 4).map((b) => b.getAttribute('aria-label') || b.textContent).join(' | ')`
  ```
  "verdeckt · Hand | verdeckt · Hand | verdeckt · Hand | verdeckt · Hand"
  ```
- ✓ Z47 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 28 Schritte.
