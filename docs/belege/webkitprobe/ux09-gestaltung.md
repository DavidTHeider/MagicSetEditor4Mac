# WebKit-Probe — ux09-gestaltung.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:58Z |

- ✓ Z28 `motor wk2`
- ✓ Z29 `lade /`
- ✓ Z30 `frist 20000`
- ✓ Z31 `warte nav`
- ✓ Z32 `wartetext Erste Schritte`
- ✓ Z33 `taste escape`
- ✓ Z34 `pruefe (window.__kontraste = () => { const rgb = (s) => { const m = s.match(/rgba?\(([^)]+)\)/); if (!m) return null; const t = m[1].split(/[ ,\/]+/).filter(Boolean).map(Number); return [t[0], t[1], t[2], t.length > 3 ? t[3] : 1] }; const lin = (c) => { c /= 255; return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4) }; const L = (c) => 0.2126 * lin(c[0]) + 0.7152 * lin(c[1]) + 0.0722 * lin(c[2]); const grund = (e) => { for (let x = e; x; x = x.parentElement) { const b = rgb(getComputedStyle(x).backgroundColor); if (b && b[3] > 0.5) return b } return [255, 255, 255, 1] }; const raus = []; for (const e of document.querySelectorAll('body *')) { if (!e.checkVisibility()) continue; if (![...e.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim().length > 1)) continue; const f = rgb(getComputedStyle(e).color); if (!f) continue; const g = grund(e); const v = f[3] < 1 ? [0, 1, 2].map((i) => f[i] * f[3] + g[i] * (1 - f[3])) : f; const a = L(v), b = L(g); raus.push([Math.round((Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05) * 100) / 100, e.textContent.trim().slice(0, 30)]) } return raus.sort((x, y) => x[0] - y[0]) }, window.__unter45 = () => window.__kontraste().filter((z) => z[0] < 4.5).map((z) => z[1]).slice(0, 5), window.__waehle = (text, wert) => { const s = [...document.querySelectorAll('select')].find((w) => w.checkVisibility() && w.labels && w.labels[0] && w.labels[0].textContent.includes(text)); if (s) { if (s.value !== wert) { s.value = wert; s.dispatchEvent(new Event('change', { bubbles: true })) } return true } const g = [...document.querySelectorAll('[role=radiogroup]')].find((r) => r.checkVisibility() && document.getElementById(r.getAttribute('aria-labelledby')).textContent.includes(text)); const f = g && [...g.querySelectorAll('input[type=radio]')].find((i) => i.value === wert); if (f && !f.checked) f.click(); return !!f }, window.__gebiet = () => { const b = [...document.querySelectorAll('main button')].find((k) => k.checkVisibility() && k.textContent.startsWith('Erscheinungsbild')); if (b) b.dataset.ziel = 'eb'; return true }, window.__grundbild = () => { const e = [...document.querySelectorAll('.bg-grund')].find((x) => x.checkVisibility()); return e ? getComputedStyle(e) : null }, true)`
- ✓ Z36 `notiz Die Kartenvorschau, bevor irgendetwas an der Schrift geschieht — ihre Adresse folgt aus dem Inhalt`
- ✓ Z37 `taste cmd+2`
- ✓ Z38 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z39 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.textContent.includes('Werkstatt-Testkarten')).dataset.ziel = 'set', true)`
- ✓ Z40 `klicke [data-ziel="set"]`
- ✓ Z41 `bis [...document.querySelectorAll('main img')].some((b) => b.checkVisibility() && b.src.includes('/vorschau/'))`
- ✓ Z42 `pruefe (window.__karte = [...document.querySelectorAll('main img')].find((b) => b.checkVisibility() && b.src.includes('/vorschau/')).src, true)`
- ✓ Z43 `zeige 'Kartenvorschau: ' + window.__karte`
  ```
  "Kartenvorschau: http://127.0.0.1:56918/vorschau/e56cabe9a3d71945/Aschenwaechter.png"
  ```
- ✓ Z45 `notiz AK-01 — die Farbwelten, über die Einstellungen`
- ✓ Z46 `taste cmd+,`
- ✓ Z47 `wartetext Erscheinungsbild`
- ✓ Z48 `pruefe window.__gebiet()`
- ✓ Z49 `klicke [data-ziel="eb"]`
- ✓ Z50 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Farbwelt'))`
- ✓ Z51 `bis (window.__waehle('Farbwelt', 'klar'), document.documentElement.dataset.farbwelt === 'klar')`
- ✓ Z52 `pause 300`
- ✓ Z53 `zeige 'klar, hell — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "klar, hell — kleinstes: [6.09,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z54 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z55 `abzug farbwelt-klar-hell`
  ```
  Bild: farbwelt-klar-hell.png
  ```
- ✓ Z56 `erscheinung dunkel`
- ✓ Z57 `pause 500`
- ✓ Z58 `zeige 'klar, dunkel — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "klar, dunkel — kleinstes: [5.81,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z59 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z60 `abzug farbwelt-klar-dunkel`
  ```
  Bild: farbwelt-klar-dunkel.png
  ```
- ✓ Z61 `erscheinung hell`
- ✓ Z62 `bis (window.__waehle('Farbwelt', 'schiefer'), document.documentElement.dataset.farbwelt === 'schiefer')`
- ✓ Z63 `pause 300`
- ✓ Z64 `zeige 'schiefer, hell — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "schiefer, hell — kleinstes: [5.86,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z65 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z66 `abzug farbwelt-schiefer-hell`
  ```
  Bild: farbwelt-schiefer-hell.png
  ```
- ✓ Z67 `erscheinung dunkel`
- ✓ Z68 `pause 500`
- ✓ Z69 `zeige 'schiefer, dunkel — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "schiefer, dunkel — kleinstes: [6.45,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z70 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z71 `erscheinung hell`
- ✓ Z72 `bis (window.__waehle('Farbwelt', 'moos'), document.documentElement.dataset.farbwelt === 'moos')`
- ✓ Z73 `pause 300`
- ✓ Z74 `zeige 'moos, hell — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "moos, hell — kleinstes: [5.78,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z75 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z76 `abzug farbwelt-moos-hell`
  ```
  Bild: farbwelt-moos-hell.png
  ```
- ✓ Z77 `erscheinung dunkel`
- ✓ Z78 `pause 500`
- ✓ Z79 `zeige 'moos, dunkel — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "moos, dunkel — kleinstes: [6.36,\"Thema, Schrift, Fenster\"]"
  ```
- ✓ Z80 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z81 `abzug farbwelt-moos-dunkel`
  ```
  Bild: farbwelt-moos-dunkel.png
  ```
- ✓ Z82 `notiz — und auf einer zweiten Seite, dunkel`
- ✓ Z83 `taste cmd+4`
- ✓ Z84 `wartetext Probe-Deck`
- ✓ Z85 `pause 500`
- ✓ Z86 `zeige 'moos, dunkel, Decks — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "moos, dunkel, Decks — kleinstes: [7.59,\"Arbeitsbereiche\"]"
  ```
- ✓ Z87 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z88 `erscheinung hell`
- ✓ Z89 `pause 300`
- ✓ Z90 `zeige 'moos, hell, Decks — kleinstes: ' + JSON.stringify(window.__kontraste()[0])`
  ```
  "moos, hell, Decks — kleinstes: [6.36,\"Spielen\"]"
  ```
- ✓ Z91 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z93 `notiz AK-01 — das Fensterbild: Vorschau, Skalierung, Aus`
- ✓ Z94 `taste cmd+,`
- ✓ Z95 `wartetext Erscheinungsbild`
- ✓ Z96 `pruefe window.__gebiet()`
- ✓ Z97 `klicke [data-ziel="eb"]`
- ✓ Z98 `bis document.querySelector('[data-vorschau="fensterbild"]') !== null`
- ✓ Z99 `pruefe getComputedStyle(document.querySelector('[data-vorschau="fensterbild"]')).backgroundImage.includes('/fensterbild/')`
- ✓ Z100 `zeige 'vorher: ' + (document.documentElement.dataset.hintergrund || 'kein Fensterbild')`
  ```
  "vorher: kein Fensterbild"
  ```
- ✓ Z101 `pruefe document.documentElement.dataset.hintergrund === undefined`
- ✓ Z102 `bis (window.__waehle('Fensterbild', 'fuellen'), document.documentElement.dataset.hintergrund === 'fuellen')`
- ✓ Z103 `pause 500`
- ✓ Z104 `zeige 'füllen: ' + window.__grundbild().backgroundSize + ' | ' + window.__grundbild().backgroundRepeat`
  ```
  "füllen: auto, cover | no-repeat"
  ```
- ✓ Z105 `pruefe window.__grundbild().backgroundImage.includes('/fensterbild/') && window.__grundbild().backgroundSize.includes('cover')`
- ✓ Z106 `zeige 'Schleier: ' + window.__grundbild().backgroundImage.slice(0, 90)`
  ```
  "Schleier: linear-gradient(color(srgb 0.905882 0.933333 0.886275 / 0.9), color(srgb 0.905882 0.933333"
  ```
- ✓ Z107 `gleich JSON.stringify(window.__unter45()) => []`
- ✓ Z108 `abzug fensterbild-fuellen`
  ```
  Bild: fensterbild-fuellen.png
  ```
- ✓ Z109 `bis (window.__waehle('Fensterbild', 'kacheln'), document.documentElement.dataset.hintergrund === 'kacheln')`
- ✓ Z110 `pause 300`
- ✓ Z111 `zeige 'kacheln: ' + window.__grundbild().backgroundSize + ' | ' + window.__grundbild().backgroundRepeat`
  ```
  "kacheln: auto, auto | repeat"
  ```
- ✓ Z112 `pruefe window.__grundbild().backgroundRepeat.includes('repeat')`
- ✓ Z113 `bis (window.__waehle('Fensterbild', 'einpassen'), document.documentElement.dataset.hintergrund === 'einpassen')`
- ✓ Z114 `pause 300`
- ✓ Z115 `pruefe window.__grundbild().backgroundSize.includes('contain')`
- ✓ Z116 `abzug fensterbild-einpassen`
  ```
  Bild: fensterbild-einpassen.png
  ```
- ✓ Z117 `bis (window.__waehle('Fensterbild', 'aus'), document.documentElement.dataset.hintergrund === undefined)`
- ✓ Z118 `pruefe !window.__grundbild().backgroundImage.includes('/fensterbild/')`
- ✓ Z120 `notiz AK-02 — die ganze Oberfläche in einer Textschrift`
- ✓ Z121 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Anzeigeschrift gilt'))`
- ✓ Z122 `bis (window.__waehle('Anzeigeschrift', 'buch'), getComputedStyle(document.documentElement).getPropertyValue('--font-anzeige').includes('ui-serif'))`
- ✓ Z123 `bis (window.__waehle('Anzeigeschrift gilt', 'ganz'), getComputedStyle(document.body).fontFamily.includes('ui-serif'))`
- ✓ Z124 `zeige 'Fließtext: ' + getComputedStyle(document.body).fontFamily.slice(0, 60)`
  ```
  "Fließtext: ui-serif, New York, Iowan Old Style, Georgia, Hiragino Minch"
  ```
- ✓ Z125 `abzug schrift-ganz-buch`
  ```
  Bild: schrift-ganz-buch.png
  ```
- ✓ Z126 `notiz — eine Schmuckschrift bleibt an den Überschriften, und die Seite sagt es`
- ✓ Z127 `bis (window.__waehle('Anzeigeschrift', 'kalligrafie'), !getComputedStyle(document.body).fontFamily.includes('ui-serif'))`
- ✓ Z128 `pruefe !getComputedStyle(document.body).fontFamily.includes('Trattatello')`
- ✓ Z129 `wartetext bleibt an den Überschriften`
- ✓ Z130 `notiz — japanische und chinesische Zeichen fallen auf Hiragino und PingFang/Songti`
- ✓ Z131 `bis (window.__waehle('Anzeigeschrift', 'buch'), getComputedStyle(document.body).fontFamily.includes('ui-serif'))`
- ✓ Z135 `pruefe (window.__cjk = (familie) => { const h = document.createElement('h2'); h.textContent = '日本語の見出し中文标题'; h.style.position = 'absolute'; h.style.whiteSpace = 'nowrap'; document.body.appendChild(h); const s = document.createElement('h2'); s.textContent = h.textContent; s.style.cssText = 'position:absolute;white-space:nowrap;font-family:' + familie; document.body.appendChild(s); const w = [h.getBoundingClientRect().width, s.getBoundingClientRect().width]; h.remove(); s.remove(); return w }, true)`
- ✓ Z136 `zeige 'Überschrift gegen Hiragino Mincho: ' + JSON.stringify(window.__cjk("'Hiragino Mincho ProN', 'Songti SC'"))`
  ```
  "Überschrift gegen Hiragino Mincho: [176,176]"
  ```
- ✓ Z137 `pruefe Math.abs(window.__cjk("'Hiragino Mincho ProN', 'Songti SC'")[0] - window.__cjk("'Hiragino Mincho ProN', 'Songti SC'")[1]) < 1`
- ✓ Z138 `pruefe document.fonts.check("16px 'Hiragino Mincho ProN'", '日本') && document.fonts.check("16px 'PingFang SC'", '中文')`
- ✓ Z139 `bis (window.__waehle('Anzeigeschrift gilt', 'ueberschriften'), !getComputedStyle(document.body).fontFamily.includes('ui-serif'))`
- ✓ Z140 `notiz — und die Karte bleibt, wie sie war: dieselbe Adresse, also derselbe Inhalt`
- ✓ Z141 `taste cmd+2`
- ✓ Z142 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z143 `pruefe ([...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.textContent.includes('Werkstatt-Testkarten')).dataset.ziel = 'set', true)`
- ✓ Z144 `klicke [data-ziel="set"]`
- ✓ Z145 `bis [...document.querySelectorAll('main img')].some((b) => b.checkVisibility() && b.src.includes('/vorschau/'))`
- ✓ Z146 `gleich [...document.querySelectorAll('main img')].find((b) => b.checkVisibility() && b.src.includes('/vorschau/')).src === window.__karte => true`
- ✓ Z147 `taste cmd+,`
- ✓ Z148 `wartetext Erscheinungsbild`
- ✓ Z149 `pruefe window.__gebiet()`
- ✓ Z150 `klicke [data-ziel="eb"]`
- ✓ Z151 `bis [...document.querySelectorAll('select, [role=radiogroup]')].some((w) => w.checkVisibility() && (w.labels?.[0] ?? document.getElementById(w.getAttribute('aria-labelledby')))?.textContent.includes('Anzeigeschrift gilt'))`
- ✓ Z152 `bis (window.__waehle('Anzeigeschrift', 'erben'), getComputedStyle(document.documentElement).getPropertyValue('--font-anzeige').includes('Hoefler'))`
- ✓ Z154 `notiz AK-03/04 — eigene Farben: die Form an ihrer Zeile, unlesbar gilt nicht, lesbar gilt`
- ✓ Z156 `bis [...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Eigene Farben')`
- ✓ Z157 `pruefe ([...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Eigene Farben').dataset.ziel = 'farbfalte', true)`
- ✓ Z158 `klicke [data-ziel="farbfalte"]`
- ✓ Z159 `bis document.querySelector('[data-ziel="farbfalte"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z160 `warte textarea[data-einstellung="gestaltung"]`
- ✓ Z164 `klicke textarea[data-einstellung="gestaltung"]`
- ✓ Z165 `tippe --mse-grund: #111111;`
- ✓ Z166 `taste enter`
- ✓ Z167 `tippe --mse-grund: url(x);`
- ✓ Z168 `pruefe (document.querySelector('textarea[data-einstellung="gestaltung"]').closest('section').querySelector('button').dataset.ziel = 'uebernehmen', true)`
- ✓ Z169 `klicke [data-ziel="uebernehmen"]`
- ✓ Z170 `wartetext Zeile 2:`
- ✓ Z171 `gleich document.querySelector('textarea[data-einstellung="gestaltung"]').getAttribute('aria-invalid') => true`
- ✓ Z172 `klicke textarea[data-einstellung="gestaltung"]`
- ✓ Z173 `taste cmd+a`
- ✓ Z174 `taste backspace`
- ✓ Z175 `tippe --mse-text: #efe5d0;`
- ✓ Z176 `pruefe (document.querySelector('textarea[data-einstellung="gestaltung"]').closest('section').querySelector('button').dataset.ziel = 'uebernehmen', true)`
- ✓ Z177 `klicke [data-ziel="uebernehmen"]`
- ✓ Z178 `wartetext Nicht angewandt`
- ✓ Z179 `pruefe document.documentElement.style.getPropertyValue('--mse-text') === ''`
- ✓ Z180 `zeige [...document.querySelectorAll('[role=status]')].map((e) => e.textContent.trim()).filter((t) => t.includes(': ')).join(' | ').slice(0, 200)`
  ```
  "Nicht angewandt — so wäre es nicht zu lesen:--mse-text auf --mse-grund: 1,06 : 1, verlangt 4,50 : 1--mse-text auf --mse-flaeche: 1,19 : 1, verlangt 4,50 : 1--mse-text auf --mse-betont: 1,04 : 1, verla"
  ```
- ✓ Z181 `abzug eigene-farben-unlesbar`
  ```
  Bild: eigene-farben-unlesbar.png
  ```
- ✓ Z182 `klicke textarea[data-einstellung="gestaltung"]`
- ✓ Z183 `taste cmd+a`
- ✓ Z184 `taste backspace`
- ✓ Z185 `tippe --mse-marke: #4e2f16;`
- ✓ Z186 `pruefe (document.querySelector('textarea[data-einstellung="gestaltung"]').closest('section').querySelector('button').dataset.ziel = 'uebernehmen', true)`
- ✓ Z187 `klicke [data-ziel="uebernehmen"]`
- ✓ Z188 `wartetext Angewandt.`
- ✓ Z189 `gleich document.documentElement.style.getPropertyValue('--mse-marke') => #4e2f16`
- ✓ Z190 `klicke textarea[data-einstellung="gestaltung"]`
- ✓ Z191 `taste cmd+a`
- ✓ Z192 `taste backspace`
- ✓ Z193 `pruefe (document.querySelector('textarea[data-einstellung="gestaltung"]').closest('section').querySelector('button').dataset.ziel = 'uebernehmen', true)`
- ✓ Z194 `klicke [data-ziel="uebernehmen"]`
- ✓ Z195 `bis document.documentElement.style.getPropertyValue('--mse-marke') === ''`
- ✓ Z197 `notiz AK-01 — das Bild wieder weg, und die Farbwelt zurück`
- ✓ Z198 `klicketext Bild entfernen`
- ✓ Z199 `wartetext Noch kein eigenes Bild`
- ✓ Z200 `bis (window.__waehle('Farbwelt', 'pergament'), document.documentElement.dataset.farbwelt === undefined)`
- ✓ Z201 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 161 Schritte.
