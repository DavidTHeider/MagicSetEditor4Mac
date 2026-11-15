# WebKit-Probe — ux02-palette.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:19Z |

- ✓ Z7 `motor wk2`
- ✓ Z8 `lade /`
- ✓ Z9 `frist 20000`
- ✓ Z10 `warte nav`
- ✓ Z11 `wartetext Erste Schritte`
- ✓ Z12 `taste escape`
- ✓ Z13 `notiz Ohne offenes Set bietet Tun das Set an, das Deck und die Sicherung`
- ✓ Z14 `taste cmd+k`
- ✓ Z15 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z16 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].length >= 3`
- ✓ Z17 `zeige [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].map((b) => b.innerText.replace(/\s+/g, ' ')).join(' | ')`
  ```
  "Ein Set anlegen oder öffnen Gestalten · Sets | Schnell spielen mit Probe-Deck Spielen · Partie beginnen | Karten tauschen mit Bo Handeln · Gegenüber | Sicherung erstellen Einstellungen · Daten und Sicherung"
  ```
- ✓ Z18 `gleich [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')][0].innerText.split('\n')[0] => Ein Set anlegen oder öffnen`
- ✓ Z19 `abzug tun`
  ```
  Bild: tun.png
  ```
- ✓ Z20 `notiz Finden: eine Karte beim Namen, mit der Pfeiltaste gewählt und mit Enter geöffnet`
- ✓ Z21 `tippe langt`
- ✓ Z22 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Finden"] button')].some((b) => b.innerText.includes('Langtext'))`
- ✓ Z23 `zeige [...document.querySelectorAll('[role=dialog] ul[aria-label="Finden"] button')].map((b) => b.innerText.replace(/\s+/g, ' ')).join(' | ')`
  ```
  "Langtext Karte · Werkstatt-Testkarten"
  ```
- ✓ Z24 `abzug finden`
  ```
  Bild: finden.png
  ```
- ✓ Z25 `taste down`
- ✓ Z26 `gleich document.activeElement.innerText.split('\n')[0] => Langtext`
- ✓ Z27 `taste enter`
- ✓ Z28 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z29 `bis [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].some((l) => l.textContent.includes('Langtext'))`
- ✓ Z30 `zeige [...document.querySelectorAll('nav[aria-label="Wo du bist"] li')].map((l) => l.textContent.replace('›', '').trim()).join(' › ')`
  ```
  "Gestalten › Werkstatt-Testkarten › Langtext"
  ```
- ✓ Z31 `abzug karte-offen`
  ```
  Bild: karte-offen.png
  ```
- ✓ Z32 `notiz Karte erstellen im offenen Set`
- ✓ Z34 `bis /\bvon 5\b/.test(document.querySelector('main').innerText)`
- ✓ Z35 `taste cmd+k`
- ✓ Z36 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z37 `tippe karte erst`
- ✓ Z38 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].some((b) => b.innerText.startsWith('Karte erstellen'))`
- ✓ Z39 `zeige [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].map((b) => b.innerText.replace(/\s+/g, ' ')).join(' | ')`
  ```
  "Karte erstellen Gestalten · Werkstatt-Testkarten"
  ```
- ✓ Z40 `pruefe ([...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].find((b) => b.innerText.startsWith('Karte erstellen')).dataset.ziel = 'tun', true)`
- ✓ Z41 `klicke [role=dialog] [data-ziel="tun"]`
- ✓ Z42 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z43 `bis /\bvon 6\b/.test(document.querySelector('main').innerText)`
- ✓ Z44 `abzug karte-erstellt`
  ```
  Bild: karte-erstellt.png
  ```
- ✓ Z45 `notiz Und die neue Karte wieder weg — über das Menü der Karte, denn die Abnahmen teilen sich einen Bestand`
- ✓ Z46 `klicketext Diese Karte`
- ✓ Z48 `bis [...document.querySelectorAll('[role=menuitem]')].some((e) => e.innerText.includes('Karte löschen'))`
- ✓ Z49 `pruefe ([...document.querySelectorAll('[role=menuitem]')].find((e) => e.innerText.includes('Karte löschen')).dataset.ziel = 'loeschen', true)`
- ✓ Z50 `klicke [data-ziel="loeschen"]`
- ✓ Z51 `bis [...document.querySelectorAll('button')].some((b) => b.innerText.replace(/\s+/g, ' ').trim().endsWith(' Löschen'))`
- ✓ Z52 `klicketext Löschen`
- ✓ Z53 `bis /\bvon 5\b/.test(document.querySelector('main').innerText)`
- ✓ Z54 `notiz Sicherung erstellen öffnet Daten und Sicherung`
- ✓ Z55 `taste cmd+k`
- ✓ Z56 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z57 `tippe sicherung`
- ✓ Z58 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].some((b) => b.innerText.startsWith('Sicherung erstellen'))`
- ✓ Z59 `pruefe ([...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].find((b) => b.innerText.startsWith('Sicherung erstellen')).dataset.ziel = 'sicherung', true)`
- ✓ Z60 `klicke [role=dialog] [data-ziel="sicherung"]`
- ✓ Z61 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z62 `bis [...document.querySelectorAll('[role=tabpanel] h2')].some((h) => h.textContent.startsWith('Daten und Sicherung'))`
- ✓ Z63 `abzug sicherung`
  ```
  Bild: sicherung.png
  ```
- ✓ Z64 `notiz Ein Deck beim Namen gefunden steht aufgeschlagen da`
- ✓ Z65 `taste cmd+k`
- ✓ Z66 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z67 `tippe probe-d`
- ✓ Z68 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Finden"] button')].some((b) => b.innerText.startsWith('Probe-Deck'))`
- ✓ Z69 `pruefe ([...document.querySelectorAll('[role=dialog] ul[aria-label="Finden"] button')].find((b) => b.innerText.startsWith('Probe-Deck')).dataset.ziel = 'deck', true)`
- ✓ Z70 `klicke [role=dialog] [data-ziel="deck"]`
- ✓ Z71 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z72 `wartetext Probe-Deck`
- ✓ Z73 `bis !document.querySelector('main').innerText.includes('Kein Deck aufgeschlagen')`
- ✓ Z74 `abzug deck-offen`
  ```
  Bild: deck-offen.png
  ```
- ✓ Z75 `notiz Schnell spielen steht in der Palette, mit dem Namen des Decks`
- ✓ Z81 `taste cmd+k`
- ✓ Z82 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z83 `tippe schnell spielen`
- ✓ Z84 `bis [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].some((b) => b.innerText.startsWith('Schnell spielen mit Probe-Deck'))`
- ✓ Z85 `zeige [...document.querySelectorAll('[role=dialog] ul[aria-label="Tun"] button')].map((b) => b.innerText.replace(/\s+/g, ' ')).join(' | ')`
  ```
  "Schnell spielen mit Probe-Deck Spielen · Partie beginnen"
  ```
- ✓ Z86 `abzug schnell-spielen`
  ```
  Bild: schnell-spielen.png
  ```
- ✓ Z87 `taste escape`
- ✓ Z88 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z89 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z90 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 77 Schritte.
