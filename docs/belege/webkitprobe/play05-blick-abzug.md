# WebKit-Probe — play05-blick-abzug.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:49Z |

- ✓ Z9 `motor wk2`
- ✓ Z10 `lade /`
- ✓ Z11 `frist 20000`
- ✓ Z12 `warte nav`
- ✓ Z13 `wartetext Erste Schritte`
- ✓ Z14 `taste escape`
- ✓ Z15 `taste cmd+4`
- ✓ Z16 `wartetext Partie anfangen`
- ✓ Z17 `klicketext Partie anfangen`
- ✓ Z18 `wartetext Weiterspielen`
- ✓ Z19 `klicketext Weiterspielen`
- ✓ Z20 `klicketext Weiterspielen als Ada`
- ✓ Z21 `warte canvas`
- ✓ Z22 `pause 1500`
- ✓ Z25 `notiz Vorbedingung: der Beginn ist vorbei — »Passen« steht für Ada offen`
- ✓ Z26 `bis [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Passen' && !b.disabled)`
- ✓ Z27 `notiz Die Steuerung nennt beide Ziele als Knöpfe`
- ✓ Z28 `pruefe [...document.querySelectorAll('button')].some((b) => (b.getAttribute('aria-label') || b.textContent).trim() === 'Auf meinen Sitz')`
- ✓ Z29 `pruefe [...document.querySelectorAll('button')].some((b) => (b.getAttribute('aria-label') || b.textContent).trim() === 'Auf die Auswahl')`
- ✓ Z30 `notiz Auf meinen Sitz — Adas Seite füllt den Blick, Bos rückt hinaus`
- ✓ Z35 `pruefe (window.__oben = () => { const f = document.querySelector('[role=application]').getBoundingClientRect(); const b = [...document.querySelectorAll('[role=application] li button')].map((x) => x.getBoundingClientRect().top - f.top); return { hoch: Math.min(...b), tief: Math.max(...b), hoehe: f.height } }, true)`
- ✓ Z36 `zeige 'vorher ' + JSON.stringify(window.__oben())`
  ```
  "vorher {\"hoch\":99.203125,\"tief\":522.125,\"hoehe\":660}"
  ```
- ✓ Z37 `pruefe window.__oben().hoch >= 0`
- ✓ Z38 `pruefe ([...document.querySelectorAll('button')].find((b) => (b.getAttribute('aria-label') || b.textContent).trim() === 'Auf meinen Sitz').dataset.ziel = 'sitz', true)`
- ✓ Z39 `klicke [data-ziel="sitz"]`
- ✓ Z40 `pause 800`
- ✓ Z41 `zeige 'nachher ' + JSON.stringify(window.__oben())`
  ```
  "nachher {\"hoch\":-57.875,\"tief\":365.03125,\"hoehe\":660}"
  ```
- ✓ Z42 `pruefe window.__oben().hoch < 0`
- ✓ Z43 `pruefe window.__oben().tief < window.__oben().hoehe`
- ✓ Z44 `abzug blick-sitz`
  ```
  Bild: blick-sitz.png
  ```
- ✓ Z45 `notiz Alles zeigen, dann eine Karte wählen und auf die Auswahl`
- ✓ Z46 `klicketext Alles zeigen`
- ✓ Z47 `pause 800`
- ✓ Z48 `pruefe (() => { const b = [...document.querySelectorAll('[role=application] li button')].filter((x) => (x.getAttribute('aria-label') || x.textContent).includes('Bibliothek')); const unten = b.reduce((a, c) => (c.getBoundingClientRect().top > a.getBoundingClientRect().top ? c : a)); unten.dataset.ziel = 'karte'; window.__kartevorher = unten.getBoundingClientRect().width; return true })()`
- ✓ Z49 `klicke [data-ziel="karte"]`
- ✓ Z50 `wartetext Mit dieser Karte`
- ✓ Z51 `pruefe ([...document.querySelectorAll('button')].find((b) => (b.getAttribute('aria-label') || b.textContent).trim() === 'Auf die Auswahl').dataset.ziel = 'auswahl', true)`
- ✓ Z52 `klicke [data-ziel="auswahl"]`
- ✓ Z53 `pause 800`
- ✓ Z54 `zeige 'Karte vorher ' + Math.round(window.__kartevorher) + ', nachher ' + Math.round(document.querySelector('[data-ziel="karte"]').getBoundingClientRect().width) + ' Punkte'`
  ```
  "Karte vorher 38, nachher 112 Punkte"
  ```
- ✓ Z55 `pruefe document.querySelector('[data-ziel="karte"]').getBoundingClientRect().width > window.__kartevorher * 1.5`
- ✓ Z56 `abzug blick-auswahl`
  ```
  Bild: blick-auswahl.png
  ```
- ✓ Z57 `notiz Das Tischbild — Vorschau für den ganzen Tisch`
- ✓ Z58 `klicketext Alles zeigen`
- ✓ Z59 `pause 500`
- ✓ Z60 `klicketext Tischbild ausgeben`
- ✓ Z61 `wartetext Vorschau erzeugen`
- ✓ Z62 `pruefe (window.__waehle = (beschriftung, wert) => { const f = [...document.querySelectorAll('label')].find((l) => l.textContent.startsWith(beschriftung)).querySelector('select'); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(f, wert); f.dispatchEvent(new Event('change', { bubbles: true })); return true }, window.__vorschau = () => { const i = document.querySelector('img[alt="Vorschau des Tischbilds"]'); return i && i.naturalWidth > 0 ? i.naturalWidth + ' × ' + i.naturalHeight : '' }, true)`
- ✓ Z63 `zeige [...[...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Ausschnitt')).querySelectorAll('option')].map((o) => o.textContent).join(' | ')`
  ```
  "Sichtbarer Bereich | Ganzer Tisch | Eine Zone"
  ```
- ✓ Z64 `zeige 'Auflösungen: ' + [...[...document.querySelectorAll('label')].find((l) => l.textContent.startsWith('Auflösung')).querySelectorAll('option')].map((o) => o.textContent).join(', ')`
  ```
  "Auflösungen: 150, 300, 600"
  ```
- ✓ Z65 `pruefe window.__waehle('Ausschnitt', 'tisch')`
- ✓ Z66 `klicketext Vorschau erzeugen`
- ✓ Z67 `bis window.__vorschau() !== ''`
- ✓ Z68 `zeige 'ganzer Tisch: ' + window.__vorschau() + ' Bildpunkte'`
  ```
  "ganzer Tisch: 7134 × 5362 Bildpunkte"
  ```
- ✓ Z69 `pruefe (window.__ganz = window.__vorschau(), true)`
- ✓ Z70 `pruefe [...document.querySelectorAll('select')].some((s) => [...s.options].some((o) => o.textContent === 'A4') && [...s.options].some((o) => o.textContent === 'Letter'))`
- ✓ Z71 `abzug abzug-tisch`
  ```
  Bild: abzug-tisch.png
  ```
- ✓ Z72 `notiz Eine Zone — die Wahl der Zone erscheint, und die Vorschau wird kleiner`
- ✓ Z73 `pruefe window.__waehle('Ausschnitt', 'zone')`
- ✓ Z74 `wartetext Welche Zone`
- ✓ Z75 `klicketext Vorschau erzeugen`
- ✓ Z76 `bis window.__vorschau() !== '' && window.__vorschau() !== window.__ganz`
- ✓ Z77 `zeige 'eine Zone: ' + window.__vorschau() + ' Bildpunkte'`
  ```
  "eine Zone: 6992 × 1122 Bildpunkte"
  ```
- ✓ Z78 `pruefe (() => { const [b, h] = window.__vorschau().split(' × ').map(Number); const [gb, gh] = window.__ganz.split(' × ').map(Number); return b * h < gb * gh })()`
- ✓ Z79 `abzug abzug-zone`
  ```
  Bild: abzug-zone.png
  ```
- ✓ Z80 `notiz Der sichtbare Bereich — nach „Alles zeigen" ist er der ganze Tisch`
- ✓ Z81 `pruefe (window.__zone = window.__vorschau(), window.__waehle('Ausschnitt', 'sichtbar'))`
- ✓ Z82 `klicketext Vorschau erzeugen`
- ✓ Z83 `bis window.__vorschau() !== window.__zone`
- ✓ Z84 `zeige 'sichtbarer Bereich: ' + window.__vorschau() + ' Bildpunkte'`
  ```
  "sichtbarer Bereich: 7394 × 2929 Bildpunkte"
  ```
- ✓ Z85 `notiz Die Sicht wechselt bei offener Vorschau — Adas Tischbild verschwindet, und sichern lässt es sich nicht mehr`
- ✓ Z89 `pruefe (window.__aufschrift = (b) => [...b.childNodes].filter((n) => !(n.nodeType === 1 && n.getAttribute('aria-hidden') === 'true')).map((n) => n.textContent).join('').trim(), true)`
- ✓ Z90 `pruefe [...document.querySelectorAll('button')].some((b) => window.__aufschrift(b) === 'Als PDF sichern')`
- ✓ Z91 `pruefe window.__waehle('Sicht für', '')`
- ✓ Z92 `bis window.__vorschau() === ''`
- ✓ Z93 `pruefe ![...document.querySelectorAll('button')].some((b) => ['Bild sichern', 'Als PDF sichern'].includes(window.__aufschrift(b)))`
- ✓ Z94 `abzug abzug-sichtwechsel`
  ```
  Bild: abzug-sichtwechsel.png
  ```
- ✓ Z95 `notiz Und in der Zuschauersicht entsteht eine eigene Vorschau`
- ✓ Z100 `wartetext Tischbild ausgeben`
- ✓ Z101 `warte canvas`
- ✓ Z102 `pause 1500`
- ✓ Z106 `klicketext Alles zeigen`
- ✓ Z107 `pause 800`
- ✓ Z108 `klicketext Tischbild ausgeben`
- ✓ Z109 `wartetext Vorschau erzeugen`
- ✓ Z110 `klicketext Vorschau erzeugen`
- ✓ Z111 `bis window.__vorschau() !== ''`
- ✓ Z112 `zeige 'Zuschauersicht: ' + window.__vorschau() + ' Bildpunkte'`
  ```
  "Zuschauersicht: 7394 × 4961 Bildpunkte"
  ```
- ✓ Z113 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 89 Schritte.
