# WebKit-Probe — ux11-ansichten.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:52:09Z |

- ✓ Z11 `motor wk2`
- ✓ Z12 `lade /`
- ✓ Z13 `frist 20000`
- ✓ Z14 `warte nav`
- ✓ Z15 `wartetext Erste Schritte`
- ✓ Z16 `taste escape`
- ✓ Z17 `pruefe (window.__form = () => { const l = [...document.querySelectorAll('ul[data-form]')].find((x) => x.checkVisibility()); return l ? l.dataset.form : 'keine' }, window.__knopf = (text) => { document.querySelectorAll('[data-ziel="ziel"]').forEach((x) => delete x.dataset.ziel); const b = [...document.querySelectorAll('main button, main [role=radiogroup] label')].find((k) => k.checkVisibility() && k.textContent.trim() === text); if (b) b.dataset.ziel = 'ziel'; return b !== undefined }, window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s && s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return s !== undefined }, true)`
- ✓ Z19 `notiz Sets — Kacheln sind die Vorgabe, die Liste wird gemerkt`
- ✓ Z22 `taste cmd+2`
- ✓ Z23 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z24 `bis window.__form() === 'kacheln'`
- ✓ Z25 `abzug sets-kacheln`
  ```
  Bild: sets-kacheln.png
  ```
- ✓ Z26 `pruefe (document.querySelector('ul[data-form] button').focus(), true)`
- ✓ Z27 `pruefe window.__knopf('Liste')`
- ✓ Z28 `klicke [data-ziel="ziel"]`
- ✓ Z29 `bis window.__form() === 'liste'`
- ✓ Z30 `notiz — der Eintrag, an dem man war, hat wieder den Fokus`
- ✓ Z31 `zeige (() => { const e = document.activeElement; return e.tagName + ' ' + (e.getAttribute('aria-label') || e.innerText || e.value || '').trim().slice(0, 60) + (e.closest('[data-eintrag]') ? ' · im Eintrag' : '') })()`
  ```
  "BUTTON Werkstatt-Testkarten — Karten dieses Sets zeigen · im Eintrag"
  ```
- ✓ Z32 `gleich document.activeElement.closest('[data-eintrag]') !== null => true`
- ✓ Z33 `abzug sets-liste`
  ```
  Bild: sets-liste.png
  ```
- ✓ Z34 `notiz — gefiltert leer ist nicht leer`
- ✓ Z35 `pruefe (document.querySelector('input[type=search]').focus(), true)`
- ✓ Z36 `tippe xyz`
- ✓ Z37 `wartetext Nichts passt zu »xyz«.`
- ✓ Z38 `pruefe window.__knopf('Suche leeren')`
- ✓ Z39 `klicke [data-ziel="ziel"]`
- ✓ Z40 `bis window.__form() === 'liste'`
- ✓ Z41 `notiz — neu geladen: die Liste steht noch da`
- ✓ Z42 `pause 1500`
- ✓ Z43 `lade /`
- ✓ Z44 `warte nav`
- ✓ Z45 `taste escape`
- ✓ Z46 `pruefe (window.__form = () => { const l = [...document.querySelectorAll('ul[data-form]')].find((x) => x.checkVisibility()); return l ? l.dataset.form : 'keine' }, window.__knopf = (text) => { document.querySelectorAll('[data-ziel="ziel"]').forEach((x) => delete x.dataset.ziel); const b = [...document.querySelectorAll('main button, main [role=radiogroup] label')].find((k) => k.checkVisibility() && k.textContent.trim() === text); if (b) b.dataset.ziel = 'ziel'; return b !== undefined }, window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s && s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return s !== undefined }, true)`
- ✓ Z47 `taste cmd+2`
- ✓ Z48 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z49 `bis window.__form() === 'liste'`
- ✓ Z51 `notiz Decks — die Liste ist die Vorgabe, Kacheln werden gemerkt`
- ✓ Z52 `taste cmd+4`
- ✓ Z53 `wartetext Probe-Deck`
- ✓ Z54 `bis window.__form() === 'liste'`
- ✓ Z55 `pruefe window.__knopf('Kacheln')`
- ✓ Z56 `klicke [data-ziel="ziel"]`
- ✓ Z57 `bis window.__form() === 'kacheln'`
- ✓ Z58 `abzug decks-kacheln`
  ```
  Bild: decks-kacheln.png
  ```
- ✓ Z59 `pause 1500`
- ✓ Z60 `lade /`
- ✓ Z61 `warte nav`
- ✓ Z62 `taste escape`
- ✓ Z63 `pruefe (window.__form = () => { const l = [...document.querySelectorAll('ul[data-form]')].find((x) => x.checkVisibility()); return l ? l.dataset.form : 'keine' }, window.__knopf = (text) => { document.querySelectorAll('[data-ziel="ziel"]').forEach((x) => delete x.dataset.ziel); const b = [...document.querySelectorAll('main button, main [role=radiogroup] label')].find((k) => k.checkVisibility() && k.textContent.trim() === text); if (b) b.dataset.ziel = 'ziel'; return b !== undefined }, window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s && s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return s !== undefined }, true)`
- ✓ Z64 `taste cmd+4`
- ✓ Z65 `wartetext Probe-Deck`
- ✓ Z66 `bis window.__form() === 'kacheln'`
- ✓ Z68 `notiz Sammlung — die Galerie wird gemerkt; seit UXR-10 (#1918) ein Segment über der Liste statt einer Auswahlliste im Navigator`
- ✓ Z69 `taste cmd+3`
- ✓ Z70 `wartetext Einlesen und ausgeben …`
- ✓ Z71 `bis window.__knopf('Galerie')`
- ✓ Z72 `klicke [data-ziel="ziel"]`
- ✓ Z73 `bis [...document.querySelectorAll('main [role=group][aria-label="Ansicht"] button')].some((b) => b.getAttribute('aria-pressed') === 'true' && b.textContent.trim() === 'Galerie')`
- ✓ Z74 `pause 1500`
- ✓ Z75 `lade /`
- ✓ Z76 `warte nav`
- ✓ Z77 `taste escape`
- ✓ Z78 `pruefe (window.__form = () => { const l = [...document.querySelectorAll('ul[data-form]')].find((x) => x.checkVisibility()); return l ? l.dataset.form : 'keine' }, window.__knopf = (text) => { document.querySelectorAll('[data-ziel="ziel"]').forEach((x) => delete x.dataset.ziel); const b = [...document.querySelectorAll('main button, main [role=radiogroup] label')].find((k) => k.checkVisibility() && k.textContent.trim() === text); if (b) b.dataset.ziel = 'ziel'; return b !== undefined }, window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s && s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return s !== undefined }, true)`
- ✓ Z79 `taste cmd+3`
- ✓ Z80 `wartetext Einlesen und ausgeben …`
- ✓ Z81 `bis [...document.querySelectorAll('main [role=group][aria-label="Ansicht"] button')].some((b) => b.checkVisibility() && b.getAttribute('aria-pressed') === 'true' && b.textContent.trim() === 'Galerie')`
- ✓ Z82 `abzug sammlung-galerie`
  ```
  Bild: sammlung-galerie.png
  ```
- ✓ Z84 `notiz Zurücksetzen — die Drehbücher danach laufen auf demselben Bestand`
- ✓ Z85 `pruefe window.__knopf('Tabelle')`
- ✓ Z86 `klicke [data-ziel="ziel"]`
- ✓ Z87 `bis [...document.querySelectorAll('main [role=group][aria-label="Ansicht"] button')].some((b) => b.getAttribute('aria-pressed') === 'true' && b.textContent.trim() === 'Tabelle')`
- ✓ Z88 `taste cmd+4`
- ✓ Z89 `wartetext Probe-Deck`
- ✓ Z90 `pruefe window.__knopf('Liste')`
- ✓ Z91 `klicke [data-ziel="ziel"]`
- ✓ Z92 `bis [...document.querySelectorAll('ul[data-form]')].some((x) => x.checkVisibility() && x.dataset.form === 'liste')`
- ✓ Z93 `taste cmd+2`
- ✓ Z94 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z95 `pruefe window.__knopf('Kacheln')`
- ✓ Z96 `klicke [data-ziel="ziel"]`
- ✓ Z97 `bis [...document.querySelectorAll('ul[data-form]')].some((x) => x.checkVisibility() && x.dataset.form === 'kacheln')`
- ✓ Z98 `pause 1500`
- ✓ Z99 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 83 Schritte.
