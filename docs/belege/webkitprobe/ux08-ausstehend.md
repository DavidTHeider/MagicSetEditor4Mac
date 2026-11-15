# WebKit-Probe — ux08-ausstehend.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:50:25Z |

- ✓ Z10 `motor wk2`
- ✓ Z11 `lade /`
- ✓ Z12 `frist 20000`
- ✓ Z13 `warte nav`
- ✓ Z14 `wartetext Erste Schritte`
- ✓ Z15 `taste escape`
- ✓ Z16 `pruefe (window.__halte = (name, ms, fehler) => { const echt = window.__echtfetch || (window.__echtfetch = window.fetch.bind(window)); window.fetch = async (url, init) => { const inhalt = init && typeof init.body === 'string' ? init.body : ''; if (inhalt.includes('.' + name + '"')) { await new Promise((weiter) => setTimeout(weiter, ms)); if (fehler) { window.fetch = echt; return new Response(JSON.stringify({ kind: 'RuntimeError', message: fehler }), { status: 500, headers: { 'Content-Type': 'application/json' } }) } } return echt(url, init) }; return true }, window.__zur = (seite) => { const b = [...document.querySelectorAll('[role=dialog] button')].find((x) => x.getClientRects().length > 0 && x.innerText.trim() === seite); b.dataset.ziel = 'seite'; return true }, true)`
- ✓ Z17 `notiz Decks — die Liste wird vier Sekunden zurückgehalten`
- ✓ Z18 `pruefe window.__halte('Decks', 4000)`
- ✓ Z19 `taste cmd+k`
- ✓ Z20 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z21 `pruefe window.__zur('Decks')`
- ✓ Z22 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z23 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z24 `wartetext Wird geladen …`
- ✓ Z25 `pruefe !document.body.innerText.includes('Noch kein Deck.')`
- ✓ Z26 `abzug decks-ausstehend`
  ```
  Bild: decks-ausstehend.png
  ```
- ✓ Z27 `notiz Die Antwort kommt — das Deck steht da, ohne dass jemand die Seite neu öffnet (S15.NAT-03)`
- ✓ Z28 `wartetext Probe-Deck`
- ✓ Z29 `pruefe !document.body.innerText.includes('Noch kein Deck.')`
- ✓ Z30 `pruefe !document.body.innerText.includes('Wird geladen …')`
- ✓ Z31 `abzug decks-angekommen`
  ```
  Bild: decks-angekommen.png
  ```
- ✓ Z32 `pruefe (window.fetch = window.__echtfetch, true)`
- ✓ Z33 `notiz Turniere — die erste Antwort scheitert, der nächste Versuch trägt (S15.NAT-02)`
- ✓ Z34 `pruefe window.__halte('Turniere', 300, 'Probe: die Leitung ist weg')`
- ✓ Z35 `taste cmd+k`
- ✓ Z36 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z37 `pruefe window.__zur('Turniere')`
- ✓ Z38 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z39 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z40 `wartetext Erneut laden`
- ✓ Z41 `pruefe !document.body.innerText.includes('Noch kein Turnier.')`
- ✓ Z42 `abzug turniere-fehlschlag`
  ```
  Bild: turniere-fehlschlag.png
  ```
- ✓ Z43 `klicketext Erneut laden`
- ✓ Z44 `notiz Ein leeres Ergebnis sagt, was als Nächstes dran ist`
- ✓ Z45 `wartetext Noch kein Turnier.`
- ✓ Z46 `pruefe !document.body.innerText.includes('Erneut laden')`
- ✓ Z47 `abzug turniere-leer`
  ```
  Bild: turniere-leer.png
  ```
- ✓ Z48 `notiz Vorlagen — die Paketliste wird vier Sekunden zurückgehalten`
- ✓ Z49 `pruefe window.__halte('Pakete', 4000)`
- ✓ Z50 `taste cmd+k`
- ✓ Z51 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z52 `pruefe window.__zur('Vorlagen')`
- ✓ Z53 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z54 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z55 `wartetext Wird geladen …`
- ✓ Z56 `pruefe !document.body.innerText.includes('Noch kein Paket zum Bearbeiten')`
- ✓ Z57 `abzug vorlagen-ausstehend`
  ```
  Bild: vorlagen-ausstehend.png
  ```
- ✓ Z58 `bis !document.body.innerText.includes('Wird geladen …')`
- ✓ Z59 `pruefe !document.body.innerText.includes('Noch kein Paket zum Bearbeiten')`
- ✓ Z60 `abzug vorlagen-angekommen`
  ```
  Bild: vorlagen-angekommen.png
  ```
- ✓ Z61 `pruefe (window.fetch = window.__echtfetch, true)`
- ✓ Z62 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z63 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 54 Schritte.
