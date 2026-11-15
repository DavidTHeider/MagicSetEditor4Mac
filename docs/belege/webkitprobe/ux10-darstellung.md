# WebKit-Probe — ux10-darstellung.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:51:19Z |

- ✓ Z21 `motor wk2`
- ✓ Z22 `lade /`
- ✓ Z23 `frist 20000`
- ✓ Z24 `warte nav`
- ✓ Z25 `wartetext Erste Schritte`
- ✓ Z26 `taste escape`
- ✓ Z27 `pruefe (window.__kontraste = () => { const rgb = (s) => { const m = s.match(/rgba?\(([^)]+)\)/); if (!m) return null; const t = m[1].split(/[ ,\/]+/).filter(Boolean).map(Number); return [t[0], t[1], t[2], t.length > 3 ? t[3] : 1] }; const lin = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4) }; const L = (c) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]); const grund = (e) => { for (let x = e; x; x = x.parentElement) { const b = rgb(getComputedStyle(x).backgroundColor); if (b && b[3] > 0.5) return b } return [255, 255, 255, 1] }; const raus = []; for (const e of document.querySelectorAll('body *')) { if (!e.checkVisibility()) continue; if (![...e.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim().length > 1)) continue; const f = rgb(getComputedStyle(e).color); if (!f) continue; const g = grund(e); const v = f[3] < 1 ? [0, 1, 2].map((i) => f[i] * f[3] + g[i] * (1 - f[3])) : f; const a = L(v), b = L(g); raus.push([Math.round((Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05) * 100) / 100, e.textContent.trim().slice(0, 30)]) } return raus.sort((x, y) => x[0] - y[0]) }, window.__unter7 = () => window.__kontraste().filter((z) => z[0] < 7).map((z) => z[1]).slice(0, 5), window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s) { if (s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return true } const g = [...document.querySelectorAll('[role=radiogroup]')].find((r) => r.checkVisibility() && document.getElementById(r.getAttribute('aria-labelledby')).textContent.includes(text)); const f = g && [...g.querySelectorAll('input[type=radio]')].find((i) => i.value === wert); if (f && !f.checked) f.click(); return !!f }, window.__abgeschnitten = () => [...document.querySelectorAll('main button, header button')].filter((b) => b.checkVisibility() && (b.scrollWidth > b.clientWidth + 1 || b.getBoundingClientRect().right > document.documentElement.clientWidth + 1)).map((b) => (b.innerText || b.getAttribute('aria-label') || '').trim().slice(0, 30)), true)`
- ✓ Z28 `notiz Vorher: im hellen Bild liegen Texte unter 7 : 1`
- ✓ Z29 `zeige JSON.stringify(window.__kontraste().slice(0, 3))`
  ```
  "[[6.96,\"Karten für Magic und andere Sp\"],[6.96,\"Probleme: 2 von 2 Sets nicht g\"],[8.04,\"Arbeitsbereiche\"]]"
  ```
- ✓ Z30 `pruefe window.__unter7().length > 0`
- ✓ Z31 `notiz Kontrast „Immer hoch" — Nutzer und Profile, Lasche Persönliches`
- ✓ Z32 `taste cmd+'`
- ✓ Z33 `wartetext Persönliches`
- ✓ Z34 `klicketext Persönliches`
- ✓ Z35 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Kontrast'))`
- ✓ Z36 `bis (window.__waehle('Kontrast', 'hoch'), document.documentElement.dataset.kontrast === 'hoch')`
- ✓ Z37 `notiz Jeder Text hält 7 : 1 — hell, auf vier Seiten`
- ✓ Z41 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z42 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [9.92,\"Hierhin wechseln\"]"
  ```
- ✓ Z43 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z44 `taste cmd+1`
- ✓ Z45 `wartetext Was Aufmerksamkeit braucht`
- ✓ Z46 `pause 500`
- ✓ Z47 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z48 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [11.47,\"Weiter an Set „Werkstatt-Testk\"]"
  ```
- ✓ Z49 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z50 `taste cmd+4`
- ✓ Z51 `wartetext Probe-Deck`
- ✓ Z52 `pause 500`
- ✓ Z53 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z54 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [9.92,\"Navigator ausblenden\"]"
  ```
- ✓ Z55 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z56 `taste cmd+3`
- ✓ Z57 `wartetext Einlesen und ausgeben …`
- ✓ Z58 `pause 500`
- ✓ Z59 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z60 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [9.92,\"Navigator ausblenden\"]"
  ```
- ✓ Z61 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z62 `abzug kontrast-hoch-hell`
  ```
  Bild: kontrast-hoch-hell.png
  ```
- ✓ Z63 `notiz — und dunkel`
- ✓ Z64 `erscheinung dunkel`
- ✓ Z65 `pause 500`
- ✓ Z66 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z67 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [8.06,\"Aufnehmen …\"]"
  ```
- ✓ Z68 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z69 `taste cmd+4`
- ✓ Z70 `wartetext Probe-Deck`
- ✓ Z71 `pause 500`
- ✓ Z72 `bis document.getAnimations().filter((a) => a.transitionProperty !== undefined && a.playState === 'running').length === 0`
- ✓ Z73 `zeige 'kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "kleinstes: [8.06,\"Deck anlegen\"]"
  ```
- ✓ Z74 `gleich JSON.stringify(window.__unter7()) => []`
- ✓ Z75 `abzug kontrast-hoch-dunkel`
  ```
  Bild: kontrast-hoch-dunkel.png
  ```
- ✓ Z76 `erscheinung hell`
- ✓ Z77 `pause 300`
- ✓ Z78 `notiz Weniger Bewegung „Immer wenig" — das Wurzelelement trägt es; die Oberfläche hat keine Übergänge, und der Tisch folgt ihm (tisch/bewegung.test.ts)`
- ✓ Z79 `zeige 'vorher: ' + (document.documentElement.dataset.bewegung || 'keine Angabe')`
  ```
  "vorher: keine Angabe"
  ```
- ✓ Z80 `pruefe document.documentElement.dataset.bewegung === undefined`
- ✓ Z81 `taste cmd+'`
- ✓ Z82 `wartetext Persönliches`
- ✓ Z83 `klicketext Persönliches`
- ✓ Z84 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Bewegung'))`
- ✓ Z85 `bis (window.__waehle('Bewegung', 'reduziert'), document.documentElement.dataset.bewegung === 'reduziert')`
- ✓ Z86 `taste cmd+4`
- ✓ Z87 `wartetext Probe-Deck`
- ✓ Z88 `pause 300`
- ✓ Z89 `zeige 'nachher: ' + document.documentElement.dataset.bewegung`
  ```
  "nachher: reduziert"
  ```
- ✓ Z90 `notiz Schriftgröße „Sehr groß" — Einstellungen, Erscheinungsbild`
- ✓ Z91 `taste cmd+,`
- ✓ Z92 `wartetext Erscheinungsbild`
- ✓ Z93 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.textContent.startsWith('Erscheinungsbild')).dataset.ziel = 'eb', true)`
- ✓ Z94 `klicke [data-ziel="eb"]`
- ✓ Z95 `bis [...document.querySelectorAll('select')].some((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes('Schriftgröße'))`
- ✓ Z96 `bis (window.__waehle('Schriftgröße', 'sehrgross'), getComputedStyle(document.documentElement).fontSize === '20px')`
- ✓ Z97 `notiz Kein Knopf abgeschnitten — auf fünf Seiten`
- ✓ Z98 `gleich JSON.stringify(window.__abgeschnitten()) => []`
- ✓ Z99 `taste cmd+2`
- ✓ Z100 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z101 `pause 500`
- ✓ Z102 `gleich JSON.stringify(window.__abgeschnitten()) => []`
- ✓ Z103 `taste cmd+3`
- ✓ Z104 `wartetext Einlesen und ausgeben …`
- ✓ Z105 `pause 500`
- ✓ Z106 `gleich JSON.stringify(window.__abgeschnitten()) => []`
- ✓ Z107 `taste cmd+4`
- ✓ Z108 `wartetext Probe-Deck`
- ✓ Z109 `pause 500`
- ✓ Z110 `gleich JSON.stringify(window.__abgeschnitten()) => []`
- ✓ Z111 `abzug schrift-sehrgross`
  ```
  Bild: schrift-sehrgross.png
  ```
- ✓ Z112 `taste cmd+5`
- ✓ Z113 `pause 800`
- ✓ Z114 `gleich JSON.stringify(window.__abgeschnitten()) => []`
- ✓ Z115 `notiz Zurücksetzen — die Drehbücher danach laufen auf demselben Bestand`
- ✓ Z116 `taste cmd+,`
- ✓ Z117 `wartetext Erscheinungsbild`
- ✓ Z118 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.textContent.startsWith('Erscheinungsbild')).dataset.ziel = 'eb', true)`
- ✓ Z119 `klicke [data-ziel="eb"]`
- ✓ Z120 `bis [...document.querySelectorAll('select')].some((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes('Schriftgröße'))`
- ✓ Z121 `bis (window.__waehle('Schriftgröße', 'erben'), getComputedStyle(document.documentElement).fontSize === '16px')`
- ✓ Z122 `taste cmd+'`
- ✓ Z123 `wartetext Persönliches`
- ✓ Z124 `klicketext Persönliches`
- ✓ Z125 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Kontrast'))`
- ✓ Z126 `pause 1000`
- ✓ Z127 `pruefe window.__waehle('Kontrast', 'system')`
- ✓ Z128 `pause 500`
- ✓ Z129 `pruefe window.__waehle('Bewegung', 'system')`
- ✓ Z130 `bis document.documentElement.dataset.kontrast === undefined && document.documentElement.dataset.bewegung === undefined`
- ✓ Z131 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 108 Schritte.
