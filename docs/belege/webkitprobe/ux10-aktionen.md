# WebKit-Probe — ux10-aktionen.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:51:14Z |

- ✓ Z12 `motor wk2`
- ✓ Z13 `lade /`
- ✓ Z14 `frist 20000`
- ✓ Z15 `warte nav`
- ✓ Z16 `wartetext Erste Schritte`
- ✓ Z17 `taste escape`
- ✓ Z18 `pruefe (window.__halte = (name, ms, fehler) => { const echt = window.__echtfetch || (window.__echtfetch = window.fetch.bind(window)); window.fetch = async (url, init) => { const inhalt = init && typeof init.body === 'string' ? init.body : ''; if (inhalt.includes('.' + name + '"')) { await new Promise((weiter) => setTimeout(weiter, ms)); if (fehler) { window.fetch = echt; return new Response(JSON.stringify({ kind: 'RuntimeError', message: fehler }), { status: 500, headers: { 'Content-Type': 'application/json' } }) } } return echt(url, init) }; return true }, window.__zur = (seite) => { const b = [...document.querySelectorAll('[role=dialog] button')].find((x) => x.getClientRects().length > 0 && x.innerText.trim() === seite); b.dataset.ziel = 'seite'; return true }, window.__band = () => document.querySelector('[role=status][aria-label="Meldungen des Fensters"]').innerText.trim(), true)`
- ✓ Z19 `notiz Sammlung — der Löschknopf trägt sein Zeichen, und das Zeichen ist stumm`
- ✓ Z23 `taste cmd+k`
- ✓ Z24 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z25 `pruefe window.__zur('Sammlung')`
- ✓ Z26 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z27 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z28 `bis [...document.querySelectorAll('main [role=grid][aria-label="Exemplare"] [role=gridcell][title]')].some((z) => z.checkVisibility())`
- ✓ Z29 `pruefe ([...document.querySelectorAll('main [role=grid][aria-label="Exemplare"] [role=gridcell][title]')].find((z) => z.checkVisibility()).dataset.ziel = 'zeile', true)`
- ✓ Z30 `klicke [data-ziel="zeile"]`
- ✓ Z31 `bis [...document.querySelectorAll('aside[aria-label="Exemplar"] button[data-aktion="loeschen"]')].some((b) => b.checkVisibility())`
- ✓ Z32 `pruefe ([...document.querySelectorAll('aside[aria-label="Exemplar"] button[data-aktion="loeschen"]')].find((b) => b.checkVisibility()).dataset.ziel = 'loeschen', true)`
- ✓ Z33 `gleich document.querySelector('[data-ziel="loeschen"] > [aria-hidden="true"]').textContent => ✕`
- ✓ Z34 `gleich document.querySelector('[data-ziel="loeschen"] > span:not([aria-hidden])').textContent => Löschen`
- ✓ Z35 `gleich document.getElementById(document.querySelector('[data-ziel="loeschen"]').getAttribute('aria-describedby')).textContent => Löscht es endgültig, überall.`
- ✓ Z36 `zeige 'Anlegen: ' + [...document.querySelectorAll('button[data-aktion="anlegen"]')].filter((b) => b.checkVisibility()).map((b) => b.innerText.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "Anlegen: + Aufnehmen …"
  ```
- ✓ Z37 `notiz Die Frage steht an der Klickstelle, der Fokus auf dem Abbrechen — Escape schließt sie und gibt den Fokus zurück`
- ✓ Z41 `klicke [data-ziel="loeschen"]`
- ✓ Z42 `warte [role=group]`
- ✓ Z43 `gleich document.activeElement.innerText.trim() => Abbrechen`
- ✓ Z44 `abzug sammlung-rueckfrage`
  ```
  Bild: sammlung-rueckfrage.png
  ```
- ✓ Z45 `taste escape`
- ✓ Z46 `pruefe !document.querySelector('[role=group] button') || ![...document.querySelectorAll('[role=group]')].some((g) => g.innerText.includes('Abbrechen'))`
- ✓ Z47 `bis document.activeElement.dataset.aktion === 'loeschen'`
- ✓ Z48 `notiz Mit der Tastatur: der Tooltip erscheint beim Fokus`
- ✓ Z49 `taste shift+tab`
- ✓ Z50 `taste tab`
- ✓ Z51 `bis document.activeElement.dataset.aktion === 'loeschen'`
- ✓ Z52 `bis document.getElementById(document.activeElement.getAttribute('aria-describedby')).checkVisibility()`
- ✓ Z53 `abzug sammlung-tooltip`
  ```
  Bild: sammlung-tooltip.png
  ```
- ✓ Z54 `notiz Ein Set öffnen, das Sichern scheitern lassen — die Meldung nennt Vorgang und Set`
- ✓ Z55 `taste cmd+2`
- ✓ Z56 `wartetext Werkstatt-Testkarten`
- ✓ Z57 `klicketext Werkstatt-Testkarten`
- ✓ Z58 `wartetext Wie viel du siehst`
- ✓ Z59 `pruefe window.__halte('SetSpeichern', 0, 'Eintrag nicht gefunden.')`
- ✓ Z60 `taste cmd+s`
- ✓ Z61 `bis window.__band().includes('wurde nicht gesichert')`
- ✓ Z62 `zeige window.__band()`
  ```
  "Das Set „Werkstatt-Testkarten“ wurde nicht gesichert — Es ist ein unerwarteter Fehler aufgetreten. Was genau, steht im Protokoll — unter „Einstellungen › Diagnose“ siehst du sein Ende und kannst es sichern. Suche dort nach f3c9-5ad8: unter dieser Kennung stehen alle Zeilen dieses Vorgangs.\n×"
  ```
- ✓ Z63 `pruefe window.__band().includes('Werkstatt-Testkarten')`
- ✓ Z64 `abzug band-set`
  ```
  Bild: band-set.png
  ```
- ✓ Z65 `notiz Auf einer fremden Seite steht sie nicht (S08.NAT-02)`
- ✓ Z66 `taste cmd+k`
- ✓ Z67 `warte [role=dialog] input[aria-label="Wohin?"]`
- ✓ Z68 `pruefe window.__zur('Hilfe')`
- ✓ Z69 `klicke [role=dialog] [data-ziel="seite"]`
- ✓ Z70 `bis !document.querySelector('input[aria-label="Wohin?"]')`
- ✓ Z71 `bis window.__band() === ''`
- ✓ Z72 `abzug band-hilfe`
  ```
  Bild: band-hilfe.png
  ```
- ✓ Z73 `notiz Zurück beim Set steht sie wieder — der Vorgang ist gescheitert, und das gilt weiter`
- ✓ Z74 `taste cmd+[`
- ✓ Z75 `bis window.__band().includes('wurde nicht gesichert')`
- ✓ Z76 `abzug band-zurueck`
  ```
  Bild: band-zurueck.png
  ```

**Ergebnis: bestanden** — 59 Schritte.
