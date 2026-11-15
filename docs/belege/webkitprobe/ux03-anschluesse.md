# WebKit-Probe — ux03-anschluesse.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:28Z |

- ✓ Z8 `motor wk2`
- ✓ Z9 `lade /`
- ✓ Z10 `frist 20000`
- ✓ Z11 `warte nav`
- ✓ Z12 `wartetext Erste Schritte`
- ✓ Z13 `taste escape`
- ✓ Z15 `notiz AK-02 — auf den Decks einen Namen anfangen`
- ✓ Z16 `taste cmd+k`
- ✓ Z17 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z18 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Decks').dataset.ziel = 'seite', true)`
- ✓ Z19 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z20 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z21 `klicketext Deck anlegen`
- ✓ Z22 `warte section[aria-label="Deck anlegen"] input`
- ✓ Z23 `klicke section[aria-label="Deck anlegen"] input`
- ✓ Z24 `tippe Mono-Rot`
- ✓ Z25 `gleich document.querySelector('section[aria-label="Deck anlegen"] input').value => Mono-Rot`
- ✓ Z26 `abzug decks-angefangen`
  ```
  Bild: decks-angefangen.png
  ```
- ✓ Z28 `notiz In die Sammlung und mit ⌘[ zurück — der Name steht noch da`
- ✓ Z29 `taste cmd+k`
- ✓ Z30 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z31 `pruefe ([...document.querySelectorAll('[role=dialog] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Sammlung').dataset.ziel = 'seite', true)`
- ✓ Z32 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z33 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z34 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Sammlung')`
- ✓ Z35 `taste cmd+[`
- ✓ Z36 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Decks')`
- ✓ Z37 `gleich [...document.querySelectorAll('section[aria-label="Deck anlegen"] input')].find((e) => e.offsetParent !== null).value => Mono-Rot`
- ✓ Z38 `abzug decks-zurueck`
  ```
  Bild: decks-zurueck.png
  ```
- ✓ Z39 `notiz Und mit ⌘] wieder vor`
- ✓ Z40 `taste cmd+]`
- ✓ Z41 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Sammlung')`
- ✓ Z42 `zeige [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].filter((l) => l.offsetParent !== null).map((l) => l.textContent.replace('›', '').trim()).join(' › ')`
  ```
  "Sammeln › Sammlung"
  ```
- ✓ Z44 `notiz AK-01 — die zwei Exemplare mit Scryfall-Kennung wählen, das Band trägt sie zum Deck`
- ✓ Z45 `bis [...document.querySelectorAll('main input[type=checkbox]')].filter((e) => e.offsetParent !== null && /^[0-9a-f]{8}-/.test(e.getAttribute('aria-label'))).length === 2`
- ✓ Z46 `pruefe ([...document.querySelectorAll('main input[type=checkbox]')].filter((e) => e.offsetParent !== null && /^[0-9a-f]{8}-/.test(e.getAttribute('aria-label'))).forEach((e) => e.click()), true)`
- ✓ Z47 `bis [...document.querySelectorAll('main button')].some((b) => b.offsetParent !== null && b.innerText.trim() === '3 Exemplare → Deck')`
- ✓ Z48 `abzug band`
  ```
  Bild: band.png
  ```
- ✓ Z49 `klicketext 3 Exemplare → Deck`
- ✓ Z50 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Decks')`
- ✓ Z52 `notiz Das Deck aufschlagen — der Stapel wartet darauf und zeigt, was ankam`
- ✓ Z53 `klicke button[aria-label="Öffnen: Probe-Deck"]`
- ✓ Z54 `warte [role=group][aria-label="Aus der Sammlung"]`
- ✓ Z55 `zeige document.querySelector('[role=group][aria-label="Aus der Sammlung"]').innerText.replace(/\s+/g, ' ')`
  ```
  "3 Exemplare aus der Sammlung, noch in keinem Deck: 4c1a8f6e-2b3d-4e5f-8a9b-0c1d2e3f4a5b × 2, 7d2e9a1b-3c4d-4e5f-9a0b-1c2d3e4f5a6b × 1. Alle ins Deck legen"
  ```
- ✓ Z56 `zeige document.querySelector('main').innerText.match(/\d+ im Hauptdeck/)[0]`
  ```
  "40 im Hauptdeck"
  ```
- ✓ Z57 `abzug stapel`
  ```
  Bild: stapel.png
  ```
- ✓ Z58 `klicketext Alle ins Deck legen`
- ✓ Z59 `bis !document.querySelector('[role=group][aria-label="Aus der Sammlung"]')`
- ✓ Z60 `bis /43 im Hauptdeck/.test(document.querySelector('main').innerText)`
- ✓ Z61 `zeige document.querySelector('main').innerText.match(/\d+ im Hauptdeck/)[0]`
  ```
  "43 im Hauptdeck"
  ```
- ✓ Z62 `abzug stapel-gelegt`
  ```
  Bild: stapel-gelegt.png
  ```
- ✓ Z64 `notiz Einem der zwei Drucke eine Karte mehr geben — dann fehlt eine, und die Zuordnung bietet sie zum Tausch an`
- ✓ Z67 `pruefe ([...document.querySelectorAll('main ul[aria-label="Plätze"] > li')].find((l) => l.innerText.includes('4c1a8f6e')).focus(), true)`
- ✓ Z70 `bis [...document.querySelectorAll('main ul[aria-label="Plätze"] > li')].find((l) => l.innerText.includes('4c1a8f6e'))?.querySelector('button[aria-label^="Plätze: "]')?.checkVisibility() === true`
- ✓ Z71 `pruefe ([...document.querySelectorAll('main ul[aria-label="Plätze"] > li')].find((l) => l.innerText.includes('4c1a8f6e')).querySelector('button[aria-label^="Plätze: "]').dataset.ziel = 'platzmenue', true)`
- ✓ Z72 `klicke [data-ziel="platzmenue"]`
- ✓ Z73 `bis [...document.querySelectorAll('[role=menuitem]')].some((e) => e.checkVisibility() && e.innerText.includes('Eine mehr'))`
- ✓ Z74 `pruefe ([...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes('Eine mehr')).dataset.ziel = 'eintrag', true)`
- ✓ Z75 `klicke [data-ziel="eintrag"]`
- ✓ Z76 `bis /44 im Hauptdeck/.test(document.querySelector('main').innerText)`
- ✓ Z77 `klicke [role=tab][data-lasche="zuordnung"]`
- ✓ Z78 `bis [...document.querySelectorAll('button')].some((b) => b.offsetParent !== null && b.innerText.trim() === 'Die 1 eintauschen')`
- ✓ Z79 `abzug eintauschen`
  ```
  Bild: eintauschen.png
  ```
- ✓ Z80 `klicketext Die 1 eintauschen`
- ✓ Z81 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Handeln')`
- ✓ Z82 `bis [...document.querySelectorAll('main p')].some((p) => p.offsetParent !== null && p.textContent.startsWith('„1 Karte“ wartet'))`
- ✓ Z83 `zeige [...document.querySelectorAll('main p')].find((p) => p.offsetParent !== null && p.textContent.startsWith('„1 Karte“ wartet')).textContent`
  ```
  "„1 Karte“ wartet — wähl unter „Neuer Tausch …“ ein Gegenüber, dann steht sie in seiner Zeile."
  ```
- ✓ Z84 `abzug wunsch-ohne-gegenueber`
  ```
  Bild: wunsch-ohne-gegenueber.png
  ```
- ✓ Z86 `notiz AK-03 — das Gegenüber am Ort aufnehmen, ohne die Einstellungen; seit UXR-15 (#1923) in der Schublade hinter „Neuer Tausch …"`
- ✓ Z87 `klicketext Neuer Tausch …`
- ✓ Z88 `klicketext Gegenüber aufnehmen oder bestätigen`
- ✓ Z89 `notiz seit NET-01 (#1705) steht vorn der Einladungstausch, die Kennung von Hand dahinter`
- ✓ Z90 `wartetext Meinen Code zeigen`
- ✓ Z91 `gleich [...document.querySelectorAll('[role=dialog] button')].some((b) => b.offsetParent !== null && b.innerText.trim() === 'Code eingeben') => true`
- ✓ Z92 `klicketext Von Hand mit einer Kennung eintragen`
- ✓ Z93 `wartetext Kennung des Gegenübers`
- ✓ Z94 `gleich [...document.querySelectorAll('[role=dialog] label')].some((l) => l.offsetParent !== null && l.innerText.trim().startsWith('Kennung des Gegenübers') && l.querySelector('input') !== null) => true`
- ✓ Z95 `gleich [...document.querySelectorAll('[role=dialog] button')].some((b) => b.offsetParent !== null && b.innerText.trim() === 'Kontakt aufnehmen') => true`
- ✓ Z96 `abzug gegenueber-kasten`
  ```
  Bild: gegenueber-kasten.png
  ```
- ✓ Z97 `taste escape`
- ✓ Z101 `notiz Escape schließt die Schublade sichtbar und gibt den Fokus aus ihr heraus`
- ✓ Z102 `bis ![...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())`
- ✓ Z103 `pruefe document.activeElement !== document.body && !document.activeElement.closest('[role=dialog]')`
- ✓ Z104 `zeige 'Fokus danach: ' + document.activeElement.tagName + ' ' + (document.activeElement.getAttribute('aria-label') || document.activeElement.innerText || '').trim().slice(0, 60)`
  ```
  "Fokus danach: BUTTON Neuer Tausch …"
  ```
- ✓ Z105 `notiz AK-04 — ein Druck ohne Scryfall-Kennung scheitert, und der Grund steht am Stapel`
- ✓ Z106 `taste cmd+3`
- ✓ Z107 `bis [...document.querySelectorAll('main h1')].some((h) => h.offsetParent !== null && h.textContent.trim() === 'Sammlung')`
- ✓ Z108 `pruefe ([...document.querySelectorAll('main input[type=checkbox]')].filter((e) => e.offsetParent !== null && e.checked).forEach((e) => e.click()), true)`
- ✓ Z109 `pruefe ([...document.querySelectorAll('main input[type=checkbox]')].find((e) => e.offsetParent !== null && e.getAttribute('aria-label') === 'scryfall:bbb-2').click(), true)`
- ✓ Z110 `bis [...document.querySelectorAll('main button')].some((b) => b.offsetParent !== null && b.innerText.trim() === '1 Exemplar → Deck')`
- ✓ Z111 `klicketext 1 Exemplar → Deck`
- ✓ Z112 `warte [role=group][aria-label="Aus der Sammlung"]`
- ✓ Z113 `klicketext Alle ins Deck legen`
- ✓ Z114 `warte [role=group][aria-label="Aus der Sammlung"] [role=alert]`
- ✓ Z115 `zeige document.querySelector('[role=group][aria-label="Aus der Sammlung"]').innerText.replace(/\s+/g, ' ')`
  ```
  "1 Exemplar aus der Sammlung, noch in keinem Deck: scryfall:bbb-2 × 1. 0 von 1 liegen im Deck, der Rest nicht: „scryfall:bbb-2“ hat keine Scryfall-Kennung — ein Deck nimmt nur Drucke mit einer auf. Alle ins Deck legen"
  ```
- ✓ Z116 `abzug stapel-grund`
  ```
  Bild: stapel-grund.png
  ```
- ✓ Z117 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z118 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 98 Schritte.
