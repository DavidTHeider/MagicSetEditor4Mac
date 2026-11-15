# WebKit-Probe — play06-kontextverlust.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:02Z |

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
- ✓ Z33 `notiz Der Kontext geht verloren`
- ✓ Z34 `pruefe (window.__bilder = 0, (() => { const vorher = window.requestAnimationFrame.bind(window); window.requestAnimationFrame = (f) => { window.__bilder++; return vorher(f) } })(), true)`
- ✓ Z35 `pruefe (() => { const g = document.querySelector('[role=application] canvas').getContext('webgl2'); g.getExtension('WEBGL_lose_context').loseContext(); return true })()`
- ✓ Z36 `wartetext Die Grafikverbindung ist abgerissen.`
- ✓ Z37 `abzug kontext-verloren`
  ```
  Bild: kontext-verloren.png
  ```
- ✓ Z38 `notiz Keine Schleife danach — in zwei Sekunden fast keine Bildanforderung`
- ✓ Z39 `pruefe (window.__bilder = 0, true)`
- ✓ Z40 `pause 2000`
- ✓ Z41 `zeige window.__bilder + ' Bildanforderungen in 2 s (eine Schleife: etwa 120)'`
  ```
  "0 Bildanforderungen in 2 s (eine Schleife: etwa 120)"
  ```
- ✓ Z42 `pruefe window.__bilder <= 6`
- ✓ Z43 `notiz Die Liste darunter bleibt bedienbar`
- ✓ Z44 `pruefe (() => { const b = [...document.querySelectorAll('button')].filter((x) => /Bibliothek/.test(x.getAttribute('aria-label') || x.textContent)); b[b.length - 1].dataset.ziel = 'liste'; return b.length > 0 })()`
- ✓ Z45 `klicke [data-ziel="liste"]`
- ✓ Z46 `wartetext Mit dieser Karte`
- ✓ Z47 `notiz Die Diagnose nennt den Verlust`
- ✓ Z48 `zeige [...document.querySelectorAll('[role=alert], [role=status]')].map((e) => e.innerText.replace(/\s+/g, ' ').trim()).filter((t) => t.includes('Grafik')).join(' | ')`
  ```
  "Die Grafikverbindung ist abgerissen. Der Tisch erscheint wieder, sobald du ihn verlässt und erneut aufrufst; bis dahin führt die Liste darunter jede Karte."
  ```
- ✓ Z49 `notiz Wiederherstellung — so, wie die Meldung es sagt: das Fenster neu öffnen`
- ✓ Z54 `lade /`
- ✓ Z55 `warte nav`
- ✓ Z58 `pause 1000`
- ✓ Z59 `pruefe ![...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility() && d.innerText.includes('Erste Schritte'))`
- ✓ Z60 `taste cmd+4`
- ✓ Z61 `wartetext Partie anfangen`
- ✓ Z62 `klicketext Partie anfangen`
- ✓ Z63 `wartetext Weiterspielen`
- ✓ Z64 `klicketext Weiterspielen`
- ✓ Z65 `klicketext Weiterspielen als Ada`
- ✓ Z66 `bis document.querySelector('[role=application] canvas') !== null && !document.querySelector('[role=application] canvas').getContext('webgl2').isContextLost()`
- ✓ Z67 `pause 1000`
- ✓ Z68 `pruefe ![...document.querySelectorAll('*')].some((e) => e.children.length === 0 && (e.textContent || '').includes('Die Grafikverbindung ist abgerissen.'))`
- ✓ Z69 `abzug kontext-zurueck`
  ```
  Bild: kontext-zurueck.png
  ```
- ✓ Z70 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 49 Schritte.
