# WebKit-Probe — einstellungen-dienstfelder.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:48:25Z |

- ✓ Z6 `motor wk2`
- ✓ Z7 `lade /`
- ✓ Z8 `frist 20000`
- ✓ Z9 `warte nav`
- ✓ Z10 `wartetext Erste Schritte`
- ✓ Z11 `taste escape`
- ✓ Z12 `taste cmd+,`
- ✓ Z13 `wartetext Verbindungen`
- ✓ Z14 `klicke [role=tab][data-lasche="verbindungen"]`
- ✓ Z15 `wartetext Was Fachleute hier sonst noch einstellen`
- ✓ Z16 `klicketext Was Fachleute hier sonst noch einstellen`
- ✓ Z17 `wartetext Adresse der Skriptschnittstelle`
- ✓ Z18 `zeige [...document.querySelectorAll('[role=tabpanel] label')].filter((l) => l.getClientRects().length > 0 && l.querySelector('input')).map((l) => l.innerText.trim().split('\n')[0]).join(' | ')`
  ```
  "Proxy für http | Proxy für https | Ohne Proxy erreichen | Adresse, unter der gehorcht wird | Frist: Verbindungsaufbau (s) | Frist: Antwortkopf (s) | Frist: Stillstand (s) | Schnittstelle für die Suche im eigenen Netz | Wo die eigenen Dienste lauschen | Eigenes Zeugnis der Dienste (Zertifikat) | Schlüssel zum eigenen Zeugnis | Zwischenzertifikate des Zeugnisses | Vertrauensanker des Zeugnisses | Kontingente der Dienste"
  ```
- ✓ Z19 `gleich ['Adresse der Skriptschnittstelle', 'Port der Skriptschnittstelle', 'Adresse des Werkzeugservers', 'Port des Werkzeugservers'].every((name) => [...document.querySelectorAll('[role=tabpanel] label')].some((l) => l.getClientRects().length > 0 && l.innerText.trim().startsWith(name))) => true`
- ✓ Z20 `gleich (() => { const namen = [...document.querySelectorAll('[role=tabpanel] input, [role=tabpanel] select, [role=tabpanel] textarea')].filter((e) => e.getClientRects().length > 0).map((e) => e.getAttribute('aria-label') || (e.labels && e.labels[0] ? e.labels[0].innerText.trim().split('\n')[0] : '')).filter((n) => n !== ''); return namen.filter((n, i) => namen.indexOf(n) !== i) })().length => 0`
- ✓ Z21 `abzug dienstfelder`
  ```
  Bild: dienstfelder.png
  ```
- ✓ Z22 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z23 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 18 Schritte.
