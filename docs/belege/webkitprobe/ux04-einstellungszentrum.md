# WebKit-Probe — ux04-einstellungszentrum.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:36Z |

- ✓ Z8 `motor wk2`
- ✓ Z9 `lade /`
- ✓ Z10 `frist 20000`
- ✓ Z11 `warte nav`
- ✓ Z12 `wartetext Erste Schritte`
- ✓ Z13 `taste escape`
- ✓ Z14 `taste cmd+,`
- ✓ Z15 `warte [role=tablist][aria-orientation=vertical]`
- ✓ Z16 `notiz Unter der Suche ein Satz über die Ebenen — mit der Tastatur erreichbar, die Erklärung hinter dem Hilfesymbol`
- ✓ Z17 `wartetext Jede Einstellung gilt auf einer Ebene`
- ✓ Z18 `klicketext Einstellung suchen`
- ✓ Z19 `taste tab`
- ✓ Z20 `zeige document.activeElement.tagName + ': ' + (document.activeElement.getAttribute('aria-label') || document.activeElement.innerText)`
  ```
  "BUTTON: Mehr zu „Jede Einstellung gilt auf einer Ebene: Gerät, Nutzer, Profil oder Arbeitsplatz.“"
  ```
- ✓ Z21 `bis !!document.querySelector('[role=tooltip]')`
- ✓ Z22 `zeige document.querySelector('[role=tooltip]').innerText.replace(/\s+/g, ' ')`
  ```
  "Gerät beschreibt diesen Rechner — einen Pfad, ein Fenstermaß — und reist bei einem Austausch nicht mit. Nutzer gilt für dich in jedem deiner Profile. Profil gilt nur hier und übersteuert, was du als Nutzer gesetzt hast. Arbeitsplatz ist, wie die Spalten dieses Fensters stehen. Wirksam ist, was im Profil steht, sonst was du als Nutzer gesetzt hast, sonst die Vorgabe — bei einer Richtlinie der strengste Wert. Verwalterrechte verlangt keine Einstellung."
  ```
- ✓ Z23 `gleich document.querySelector('[role=tooltip]').innerText.includes('Verwalterrechte verlangt keine Einstellung') => true`
- ✓ Z24 `abzug ebenen`
  ```
  Bild: ebenen.png
  ```
- ✓ Z25 `taste escape`
- ✓ Z26 `notiz Die Einstellungen des Arbeitsplatzes nennen ihre Ebene beim Namen`
- ✓ Z27 `pruefe ([...document.querySelectorAll('[role=tab]')].find((t) => t.innerText.startsWith('Erscheinungsbild')).dataset.ziel = 'gruppe', true)`
- ✓ Z28 `klicke [data-ziel="gruppe"]`
- ✓ Z31 `bis [...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Spalten und Ansichten')`
- ✓ Z32 `pruefe ([...document.querySelectorAll('[role=tabpanel] button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Spalten und Ansichten').dataset.ziel = 'spaltenfalte', true)`
- ✓ Z33 `klicke [data-ziel="spaltenfalte"]`
- ✓ Z34 `bis document.querySelector('[data-ziel="spaltenfalte"]').getAttribute('aria-expanded') === 'true'`
- ✓ Z38 `bis document.querySelectorAll('[role=tabpanel] [data-einstellungszeile]').length > 0`
- ✓ Z39 `gleich [...document.querySelectorAll('[role=tabpanel] [title]')].some((e) => /Ebene [a-zäöü]/.test(e.title)) || document.querySelector('[role=tabpanel]').innerText.includes('Ebene arbeitsplatz') => false`
- ✓ Z40 `notiz Unter Verbindungen: Identität, Gegenüber und Proxy offen, die Fristen im Kasten für Fachleute`
- ✓ Z41 `pruefe (document.querySelector('[data-ziel="gruppe"]').removeAttribute('data-ziel'), [...document.querySelectorAll('[role=tab]')].find((t) => t.innerText.startsWith('Verbindungen')).dataset.ziel = 'gruppe', true)`
- ✓ Z42 `klicke [data-ziel="gruppe"]`
- ✓ Z43 `wartetext Proxy für http`
- ✓ Z44 `gleich document.querySelector('[role=tabpanel]').innerText.includes('Frist: Verbindungsaufbau') => false`
- ✓ Z45 `pruefe ([...document.querySelectorAll('[role=tabpanel] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Was Fachleute hier sonst noch einstellen').dataset.ziel = 'kasten', true)`
- ✓ Z46 `gleich document.querySelector('[data-ziel="kasten"]').getAttribute('aria-expanded') => false`
- ✓ Z47 `abzug verbindungen-zu`
  ```
  Bild: verbindungen-zu.png
  ```
- ✓ Z48 `klicke [data-ziel="kasten"]`
- ✓ Z49 `wartetext Frist: Verbindungsaufbau`
- ✓ Z50 `gleich document.querySelector('[data-ziel="kasten"]').getAttribute('aria-expanded') => true`
- ✓ Z51 `abzug verbindungen-auf`
  ```
  Bild: verbindungen-auf.png
  ```
- ✓ Z52 `klicke [data-ziel="kasten"]`
- ✓ Z53 `bis !document.querySelector('[role=tabpanel]').innerText.includes('Frist: Verbindungsaufbau')`
- ✓ Z54 `notiz Die Suche öffnet den Kasten, wenn der Treffer darin liegt, und Enter springt auf das Feld`
- ✓ Z55 `klicketext Einstellung suchen`
- ✓ Z56 `tippe frist`
- ✓ Z57 `pause 400`
- ✓ Z58 `wartetext Frist: Verbindungsaufbau`
- ✓ Z59 `gleich [...document.querySelectorAll('[role=tabpanel] button')].find((b) => b.getClientRects().length > 0 && b.innerText.trim() === 'Was Fachleute hier sonst noch einstellen').getAttribute('aria-expanded') => true`
- ✓ Z60 `taste enter`
- ✓ Z61 `pause 300`
- ✓ Z62 `zeige (document.activeElement.closest('[data-einstellung]') || {}).dataset ? document.activeElement.closest('[data-einstellung]').dataset.einstellung : 'kein Feld: ' + document.activeElement.tagName`
  ```
  "netz.frist.verbindung"
  ```
- ✓ Z63 `abzug suche-frist`
  ```
  Bild: suche-frist.png
  ```
- ✓ Z64 `notiz Und die Seite hat keinen Fehler geworfen`
- ✓ Z65 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 53 Schritte.
