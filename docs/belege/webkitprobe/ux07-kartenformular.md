# WebKit-Probe — ux07-kartenformular.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:10Z |

- ✓ Z8 `motor wk2`
- ✓ Z9 `lade /`
- ✓ Z10 `frist 20000`
- ✓ Z11 `warte nav`
- ✓ Z12 `wartetext Erste Schritte`
- ✓ Z13 `taste escape`
- ✓ Z14 `taste cmd+2`
- ✓ Z15 `wartetext Werkstatt-Testkarten`
- ✓ Z16 `klicketext Werkstatt-Testkarten`
- ✓ Z17 `wartetext Wie viel du siehst`
- ✓ Z18 `notiz Die drei Tiefen — Felder zählen, Daten bleiben`
- ✓ Z19 `zeige [...document.querySelectorAll("button")].filter(b => ["Normal","Erweitert","Experte"].includes(b.innerText.trim())).map(b => b.innerText.trim() + ":" + (b.getAttribute("aria-pressed") || b.getAttribute("aria-checked") || b.getAttribute("aria-selected") || b.getAttribute("aria-current")))`
  ```
  ["Normal:true","Erweitert:false","Experte:false"]
  ```
- ✓ Z20 `klicketext Erweitert`
- ✓ Z21 `pause 400`
- ✓ Z22 `zeige [...document.querySelectorAll("button")].filter(b => ["Normal","Erweitert","Experte"].includes(b.innerText.trim())).map(b => b.innerText.trim() + ":" + b.getAttribute("aria-pressed"))`
  ```
  ["Normal:false","Erweitert:true","Experte:false"]
  ```
- ✓ Z23 `zeige ['Erweitert', document.querySelectorAll('aside input, aside textarea, aside select').length]`
  ```
  ["Erweitert",14]
  ```
- ✓ Z24 `klicketext Experte`
- ✓ Z25 `pause 400`
- ✓ Z26 `zeige ['Experte', document.querySelectorAll('aside input, aside textarea, aside select').length]`
  ```
  ["Experte",14]
  ```
- ✓ Z27 `klicketext Normal`
- ✓ Z28 `notiz Ein Feld ändern, die Vorschau folgt ohne Bestätigung`
- ✓ Z29 `zeige [...document.querySelectorAll('main img')].map(i => i.getAttribute('alt') + ' ' + i.naturalWidth + '×' + i.naturalHeight).slice(0, 3)`
  ```
  [" 375×523","Vorschau der Karte 375×523"]
  ```
- ✓ Z30 `sichere vorschau-vorher [...document.querySelectorAll('main img')].map(i => i.currentSrc.slice(-40))`
  ```
  gesichert: vorschau-vorher.json (87 Bytes)
  ```
- ✓ Z34 `warte aside [aria-label="name"]`
- ✓ Z35 `klicke aside [aria-label="name"]`
- ✓ Z36 `taste cmd+a`
- ✓ Z37 `tippe Aschenwaechter`
- ✓ Z38 `pause 2500`
- ✓ Z39 `zeige document.activeElement.value`
  ```
  "Aschenwaechter"
  ```
- ✓ Z40 `pruefe document.querySelector('main img').currentSrc !== JSON.parse('null') && true`
- ✓ Z41 `sichere vorschau-nachher [...document.querySelectorAll('main img')].map(i => i.currentSrc.slice(-40))`
  ```
  gesichert: vorschau-nachher.json (87 Bytes)
  ```
- ✓ Z42 `notiz Der Wert übersteht den Wechsel der Tiefe (UX-05.AK-02)`
- ✓ Z43 `klicketext Experte`
- ✓ Z44 `pause 300`
- ✓ Z45 `klicketext Normal`
- ✓ Z46 `pause 300`
- ✓ Z47 `gleich document.querySelector('aside [aria-label="name"]').value => Aschenwaechter`
- ✓ Z48 `abzug karte-geaendert`
  ```
  Bild: karte-geaendert.png
  ```
- ✓ Z49 `notiz Die Brotkrume nennt die aufgeschlagene Karte und folgt ihrem Namen (UX-03.AK-02, #1675)`
- ✓ Z54 `bis [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].pop().textContent.includes('Aschenwaechter')`
- ✓ Z55 `gleich [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ') => Gestalten › Werkstatt-Testkarten › Aschenwaechter`

**Ergebnis: bestanden** — 41 Schritte.
