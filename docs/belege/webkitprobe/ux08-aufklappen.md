# WebKit-Probe — ux08-aufklappen.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:21Z |

- ✓ Z13 `motor wk2`
- ✓ Z14 `lade /`
- ✓ Z15 `frist 20000`
- ✓ Z16 `warte nav`
- ✓ Z17 `wartetext Erste Schritte`
- ✓ Z18 `taste escape`
- ✓ Z19 `pruefe (window.__zur = (seite) => { const b = [...document.querySelectorAll('[role=dialog] button')].find((x) => x.getClientRects().length > 0 && x.innerText.trim() === seite); b.dataset.ziel = 'seite'; return true }, window.__schalter = () => [...document.querySelectorAll('main button[aria-expanded]')].find((b) => b.innerText.trim() === 'Deck anlegen'), window.__feld = () => document.querySelector('section[aria-label="Deck anlegen"] input'), true)`
- ✓ Z20 `taste cmd+k`
- ✓ Z21 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z22 `pruefe window.__zur('Decks')`
- ✓ Z23 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z24 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z25 `wartetext Probe-Deck`
- ✓ Z26 `notiz Der Kasten ist zu und sagt es — dann auf, und einen Namen tippen`
- ✓ Z27 `gleich window.__schalter().getAttribute('aria-expanded') => false`
- ✓ Z28 `gleich window.__feld() => null`
- ✓ Z29 `pruefe (window.__schalter().dataset.ziel = 'schalter', true)`
- ✓ Z30 `klicke [data-ziel="schalter"]`
- ✓ Z31 `bis window.__schalter().getAttribute('aria-expanded') === 'true'`
- ✓ Z32 `bis window.__feld() !== null && window.__feld().checkVisibility()`
- ✓ Z33 `klicke section[aria-label="Deck anlegen"] input`
- ✓ Z34 `tippe Halbfertig`
- ✓ Z35 `notiz Zu — verborgen, nicht fort`
- ✓ Z36 `klicke [data-ziel="schalter"]`
- ✓ Z37 `bis window.__schalter().getAttribute('aria-expanded') === 'false'`
- ✓ Z38 `pruefe !window.__feld().checkVisibility()`
- ✓ Z39 `pruefe !window.__schalter().hasAttribute('aria-controls')`
- ✓ Z40 `notiz Wieder auf — der Name steht noch da (S11.NAT-03)`
- ✓ Z41 `klicke [data-ziel="schalter"]`
- ✓ Z42 `bis window.__feld().checkVisibility()`
- ✓ Z43 `gleich window.__feld().value => Halbfertig`
- ✓ Z44 `abzug aufklappen-behaelt`
  ```
  Bild: aufklappen-behaelt.png
  ```
- ✓ Z45 `notiz Einstellungen › Fremde Kartendienste — der Pfad des Laufs steht unter „Diagnose" (S11.NAT-02)`
- ✓ Z46 `taste cmd+,`
- ✓ Z47 `wartetext Verbindungen`
- ✓ Z48 `klicke [role=tab][data-lasche="dienste"]`
- ✓ Z49 `wartetext Was die Dienste hier können`
- ✓ Z51 `pruefe ([...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Was die Dienste hier können').dataset.ziel = 'katalog', true)`
- ✓ Z52 `klicke [data-ziel="katalog"]`
- ✓ Z53 `bis document.querySelector('[data-ziel="katalog"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z54 `bis document.querySelector('article[aria-label="Scryfall"]') !== null`
- ✓ Z55 `pruefe !document.querySelector('main').innerText.includes('netzlauf_test.go')`
- ✓ Z56 `pruefe document.querySelector('article[aria-label="Scryfall"]').innerText.includes('nicht mit dieser Fassung.')`
- ✓ Z57 `pruefe ([...document.querySelectorAll('article[aria-label="Scryfall"] button[aria-expanded]')].find((b) => b.innerText.trim() === 'Diagnose').dataset.ziel = 'diagnose', true)`
- ✓ Z58 `klicke [data-ziel="diagnose"]`
- ✓ Z59 `bis document.querySelector('article[aria-label="Scryfall"]').innerText.includes('internal/scryfall/netzlauf_test.go')`
- ✓ Z60 `abzug diagnose`
  ```
  Bild: diagnose.png
  ```
- ✓ Z61 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z62 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 49 Schritte.
