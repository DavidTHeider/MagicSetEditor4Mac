# WebKit-Probe — ux07-einstellungen.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:57Z |

- ✓ Z6 `motor wk2`
- ✓ Z7 `lade /`
- ✓ Z8 `frist 20000`
- ✓ Z9 `warte nav`
- ✓ Z10 `wartetext Erste Schritte`
- ✓ Z11 `taste escape`
- ✓ Z12 `taste cmd+,`
- ✓ Z13 `warte [role=tablist][aria-orientation=vertical]`
- ✓ Z14 `notiz Tab erreicht nach dem Hilfesymbol der Suche die gewählte Gruppe, die Pfeile wandern`
- ✓ Z15 `klicketext Einstellung suchen`
- ✓ Z16 `taste tab`
- ✓ Z17 `pruefe document.activeElement.tagName === 'BUTTON' && document.activeElement.getAttribute('aria-label').startsWith('Mehr zu')`
- ✓ Z18 `taste tab`
- ✓ Z19 `pruefe document.activeElement.getAttribute('role') === 'tab' && document.activeElement.getAttribute('aria-selected') === 'true'`
- ✓ Z20 `zeige document.activeElement.innerText.replace(/\s+/g, ' ')`
  ```
  "Allgemein Sprache, Start, Kürzel"
  ```
- ✓ Z21 `taste down`
- ✓ Z22 `pruefe document.activeElement.getAttribute('role') === 'tab' && document.activeElement.getAttribute('aria-selected') === 'true'`
- ✓ Z23 `zeige document.activeElement.innerText.replace(/\s+/g, ' ')`
  ```
  "Bereiche und Funktionen Arbeitsbereiche, Funktionen, Reife"
  ```
- ✓ Z24 `taste down`
- ✓ Z25 `taste down`
- ✓ Z26 `zeige document.activeElement.innerText.replace(/\s+/g, ' ')`
  ```
  "Nutzer und Profil Anmeldung, Entsperren"
  ```
- ✓ Z27 `taste up`
- ✓ Z28 `zeige document.activeElement.innerText.replace(/\s+/g, ' ')`
  ```
  "Erscheinungsbild Thema, Schrift, Fenster"
  ```
- ✓ Z29 `abzug einstellungen-hell`
  ```
  Bild: einstellungen-hell.png
  ```
- ✓ Z30 `notiz Die Suche springt auf das Feld`
- ✓ Z31 `klicketext Einstellung suchen`
- ✓ Z32 `tippe proxy`
- ✓ Z33 `pause 400`
- ✓ Z34 `zeige document.querySelector('[role=tablist][aria-orientation=vertical]').innerText.replace(/\s+/g, ' ')`
  ```
  "Allgemein Sprache, Start, Kürzel Bereiche und Funktionen Arbeitsbereiche, Funktionen, Reife Erscheinungsbild Thema, Schrift, Fenster Nutzer und Profil Anmeldung, Entsperren Fachfunktionen Karten, Sammlung, Spiel KI Anbieter, Modelle, Kosten Verbindungen 3 Netz, Gegenüber, CLI, MCP, REST Dienste Kartendienste, Preise, Markt Daten und Sicherung Bestand, Pfade, Sicherung Diagnose Protokolle, Messung, Technik"
  ```
- ✓ Z35 `taste enter`
- ✓ Z36 `pause 300`
- ✓ Z37 `zeige (document.activeElement.closest('[data-einstellung]') || {}).dataset ? document.activeElement.closest('[data-einstellung]').dataset.einstellung : 'kein Feld: ' + document.activeElement.tagName`
  ```
  "netz.proxy.http"
  ```
- ✓ Z38 `abzug einstellungen-proxy`
  ```
  Bild: einstellungen-proxy.png
  ```
- ✓ Z39 `erscheinung dunkel`
- ✓ Z40 `abzug einstellungen-dunkel`
  ```
  Bild: einstellungen-dunkel.png
  ```

**Ergebnis: bestanden** — 35 Schritte.
