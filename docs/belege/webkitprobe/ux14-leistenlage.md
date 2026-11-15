# WebKit-Probe — ux14-leistenlage.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:53:54Z |

- ✓ Z11 `motor wk2`
- ✓ Z12 `lade /`
- ✓ Z13 `frist 20000`
- ✓ Z14 `warte nav`
- ✓ Z15 `wartetext Erste Schritte`
- ✓ Z16 `taste escape`
- ✓ Z17 `pruefe (window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s) { if (s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return true } const g = [...document.querySelectorAll('[role=radiogroup]')].find((r) => r.checkVisibility() && document.getElementById(r.getAttribute('aria-labelledby')).textContent.includes(text)); const f = g && [...g.querySelectorAll('input[type=radio]')].find((i) => i.value === wert); if (f && !f.checked) f.click(); return !!f }, window.__eintraege = () => [...document.querySelectorAll('nav[aria-label="Hauptnavigation"] button')], true)`
- ✓ Z18 `notiz Vorher: links, mit Knopf zum Einklappen`
- ✓ Z19 `gleich document.querySelector('[data-lage]').dataset.lage => senkrecht`
- ✓ Z20 `notiz „Lage der Leiste: Oben" — Einstellungen, Erscheinungsbild`
- ✓ Z21 `taste cmd+,`
- ✓ Z22 `wartetext Erscheinungsbild`
- ✓ Z23 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.textContent.startsWith('Erscheinungsbild')).dataset.ziel = 'eb', true)`
- ✓ Z24 `klicke [data-ziel="eb"]`
- ✓ Z26 `klicketext Fenster und Leiste`
- ✓ Z27 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Lage der Leiste'))`
- ✓ Z28 `bis (window.__waehle('Lage der Leiste', 'waagerecht'), document.querySelector('[data-lage]').dataset.lage === 'waagerecht')`
- ✓ Z29 `notiz Die sechs Arbeitsbereiche stehen in einer Zeile, von links nach rechts, und nichts ist einzuklappen`
- ✓ Z30 `pruefe (() => { const r = window.__eintraege().slice(0, 6).map((b) => b.getBoundingClientRect()); return r.every((x) => Math.round(x.top) === Math.round(r[0].top)) && r.every((x, i) => i === 0 || x.left > r[i - 1].left) })()`
- ✓ Z31 `pruefe !document.querySelector('button[aria-controls="seitenleiste"]')`
- ✓ Z32 `abzug leiste-oben`
  ```
  Bild: leiste-oben.png
  ```
- ✓ Z33 `notiz Kein Eintrag über dem Rand — bei 1.440 × 900 und bei 1.024 × 700, der kleinsten erklärten Größe`
- ✓ Z34 `pruefe window.__eintraege().every((b) => { const r = b.getBoundingClientRect(); return r.width > 0 && r.left >= 0 && r.right <= document.documentElement.clientWidth + 1 })`
- ✓ Z35 `zeige 'Zeile bei 1440: ' + Math.round(document.querySelector('[data-lage]').getBoundingClientRect().height) + ' Punkte hoch'`
  ```
  "Zeile bei 1440: 49 Punkte hoch"
  ```
- ✓ Z36 `groesse 1024 700`
  ```
  1024 × 700 Punkte
  ```
- ✓ Z37 `pause 500`
- ✓ Z38 `pruefe window.__eintraege().every((b) => { const r = b.getBoundingClientRect(); return r.width > 0 && r.left >= 0 && r.right <= document.documentElement.clientWidth + 1 })`
- ✓ Z39 `zeige 'Zeile bei 1024: ' + Math.round(document.querySelector('[data-lage]').getBoundingClientRect().height) + ' Punkte hoch'`
  ```
  "Zeile bei 1024: 87 Punkte hoch"
  ```
- ✓ Z40 `abzug leiste-oben-1024`
  ```
  Bild: leiste-oben-1024.png
  ```
- ✓ Z41 `groesse 1440 900`
  ```
  1440 × 900 Punkte
  ```
- ✓ Z42 `pause 300`
- ✓ Z43 `notiz Dieselben Kürzel und dieselbe Markierung`
- ✓ Z44 `taste cmd+3`
- ✓ Z45 `wartetext Einlesen und ausgeben …`
- ✓ Z46 `gleich window.__eintraege().find((b) => b.getAttribute('aria-current') === 'page').innerText.replace(/\s+/g, ' ').trim() => ▤ Sammeln`
- ✓ Z47 `gleich window.__eintraege().find((b) => b.getAttribute('aria-current') === 'page').title => ⌘3`
- ✓ Z48 `gleich getComputedStyle(window.__eintraege().find((b) => b.getAttribute('aria-current') === 'page')).borderBottomWidth => 2px`
- ✓ Z49 `notiz Derselbe Fokusweg: die Leiste ist ein Halt, → wandert`
- ✓ Z50 `lade /`
- ✓ Z51 `warte nav`
- ✓ Z52 `pause 800`
- ✓ Z53 `taste escape`
- ✓ Z54 `pruefe (window.__eintraege = () => [...document.querySelectorAll('nav[aria-label="Hauptnavigation"] button')], window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s) { if (s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return true } const g = [...document.querySelectorAll('[role=radiogroup]')].find((r) => r.checkVisibility() && document.getElementById(r.getAttribute('aria-labelledby')).textContent.includes(text)); const f = g && [...g.querySelectorAll('input[type=radio]')].find((i) => i.value === wert); if (f && !f.checked) f.click(); return !!f }, true)`
- ✓ Z55 `taste tab`
- ✓ Z56 `gleich document.activeElement.textContent.trim() => Zum Inhalt springen`
- ✓ Z57 `taste tab`
- ✓ Z58 `gleich document.activeElement.textContent.trim() => Profil wechseln …`
- ✓ Z59 `taste tab`
- ✓ Z60 `pruefe window.__eintraege().includes(document.activeElement)`
- ✓ Z61 `pruefe (window.__erster = document.activeElement, true)`
- ✓ Z62 `taste right`
- ✓ Z63 `pruefe window.__eintraege().includes(document.activeElement) && document.activeElement !== window.__erster`
- ✓ Z64 `taste tab`
- ✓ Z65 `pruefe document.activeElement.textContent.includes('Befehlspalette öffnen') && document.activeElement.textContent.includes('⌘K')`
- ✓ Z66 `abzug leiste-oben-fokus`
  ```
  Bild: leiste-oben-fokus.png
  ```
- ✓ Z67 `notiz Zurücksetzen`
- ✓ Z68 `taste cmd+,`
- ✓ Z69 `wartetext Erscheinungsbild`
- ✓ Z70 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.textContent.startsWith('Erscheinungsbild')).dataset.ziel = 'eb', true)`
- ✓ Z71 `klicke [data-ziel="eb"]`
- ✓ Z73 `klicketext Fenster und Leiste`
- ✓ Z74 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Lage der Leiste'))`
- ✓ Z75 `bis (window.__waehle('Lage der Leiste', 'senkrecht'), document.querySelector('[data-lage]').dataset.lage === 'senkrecht')`
- ✓ Z76 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 64 Schritte.
