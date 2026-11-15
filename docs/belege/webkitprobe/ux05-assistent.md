# WebKit-Probe — ux05-assistent.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:49:41Z |

- ✓ Z18 `motor wk2`
- ✓ Z19 `lade /`
- ✓ Z20 `frist 20000`
- ✓ Z21 `warte nav`
- ✓ Z22 `wartetext Erste Schritte`
- ✓ Z23 `taste escape`
- ✓ Z24 `taste cmd+2`
- ✓ Z27 `wartetext Neues Set`
- ✓ Z28 `klicketext Neues Set`
- ✓ Z29 `wartetext Ein neues Set`
- ✓ Z32 `bis document.querySelector('ol[aria-label="Schritte zum neuen Set"] li[aria-current="step"]')?.lastChild.textContent === 'Name'`
- ✓ Z33 `klicketext Zurück`
- ✓ Z34 `bis document.querySelector('ol[aria-label="Schritte zum neuen Set"] li[aria-current="step"]')?.lastChild.textContent === 'Aussehen'`
- ✓ Z35 `klicketext Zurück`
- ✓ Z36 `notiz Vorbedingung: das Spiel der Vorrichtung steht zur Wahl (testdaten/assistent)`
- ✓ Z37 `bis !!document.querySelector('fieldset input[type=radio][value="wahlspiel"]')`
- ✓ Z38 `bis document.querySelectorAll('ol[aria-label="Schritte zum neuen Set"] li').length === 3`
- ✓ Z39 `pruefe (window.__aktuell = () => document.querySelector('ol[aria-label="Schritte zum neuen Set"] li[aria-current="step"]').lastChild.textContent, window.__sichtbar = () => [...document.querySelectorAll('fieldset input[type=radio]')].map((r) => r.value), window.__gewaehlt = () => [...document.querySelectorAll('fieldset input[type=radio]')].filter((r) => r.checked).map((r) => r.value).join(), window.__knopf = (text) => { document.querySelectorAll('[data-ziel]').forEach((x) => { delete x.dataset.ziel }); const kopf = [...document.querySelectorAll('h2')].find((h) => h.textContent === 'Ein neues Set'); const knopf = [...kopf.closest('section').querySelectorAll('button')].find((b) => b.innerText.trim() === text); knopf.dataset.ziel = 'knopf'; return true }, true)`
- ✓ Z40 `notiz Schritt 1 von 3 — jedes nutzbare Spiel beim Namen, mit der Zahl seiner Stile, das Wahlspiel vorgewählt`
- ✓ Z41 `gleich window.__aktuell() => Art`
- ✓ Z42 `zeige [...document.querySelectorAll('fieldset label')].map((l) => l.innerText.replace(/\s+/g, ' ').trim()).join(' | ')`
  ```
  "Wahlspiel wahlspiel · 3 Stile verfügbar | Probe probe · 5 Stile verfügbar"
  ```
- ✓ Z43 `gleich window.__gewaehlt() => wahlspiel`
- ✓ Z44 `abzug schritt-art`
  ```
  Bild: schritt-art.png
  ```
- ✓ Z45 `notiz Ein anderes Spiel, vor und zurück — die Wahl bleibt (S07.NAT-03)`
- ✓ Z46 `klicke fieldset input[type=radio][value="probe"]`
- ✓ Z47 `gleich window.__gewaehlt() => probe`
- ✓ Z48 `pruefe window.__knopf('Weiter')`
- ✓ Z49 `klicke [data-ziel="knopf"]`
- ✓ Z50 `notiz Schritt 2 — der Fokus steht auf dem Kopf des Schritts, die Stile sind die des Spiels`
- ✓ Z51 `bis window.__aktuell() === 'Aussehen'`
- ✓ Z52 `gleich document.activeElement.tagName + ' ' + document.activeElement.textContent => H3 Aussehen`
- ✓ Z57 `pruefe (window.__stilfeld = () => [...document.querySelectorAll('input[role=combobox]')].find((e) => e.checkVisibility() && e.labels?.[0]?.textContent === 'Stil des neuen Sets'), window.__zeilen = () => [...document.querySelectorAll('[role=option]')].filter((o) => o.checkVisibility()).map((o) => o.innerText.replace(/\s+/g, ' ').trim()), window.__stilfeldZiel = () => { document.querySelectorAll('[data-ziel]').forEach((x) => { delete x.dataset.ziel }); window.__stilfeld().dataset.ziel = 'stilfeld'; return true }, !!window.__stilfeld())`
- ✓ Z58 `pruefe window.__stilfeldZiel()`
- ✓ Z59 `klicke [data-ziel="stilfeld"]`
- ✓ Z60 `bis window.__zeilen().length > 0`
- ✓ Z61 `zeige window.__zeilen().join(' | ')`
  ```
  "Mit Rahmen probe-bild | Einzeiler probe-einzeiler | Mit Kunstfenster probe-kunst | Schlicht probe-schlicht | Spielskript probe-spielskript"
  ```
- ✓ Z62 `pruefe window.__zeilen().every((z) => z.split(' ').pop().startsWith('probe-'))`
- ✓ Z63 `taste escape`
- ✓ Z64 `pruefe window.__knopf('Zurück')`
- ✓ Z65 `klicke [data-ziel="knopf"]`
- ✓ Z66 `bis window.__aktuell() === 'Art'`
- ✓ Z67 `gleich window.__gewaehlt() => probe`
- ✓ Z68 `notiz Zum Wahlspiel — der Stil kommt mit dem Spiel und ist wieder vorgewählt`
- ✓ Z69 `klicke fieldset input[type=radio][value="wahlspiel"]`
- ✓ Z70 `pruefe window.__knopf('Weiter')`
- ✓ Z71 `klicke [data-ziel="knopf"]`
- ✓ Z72 `bis window.__aktuell() === 'Aussehen'`
- ✓ Z73 `gleich window.__stilfeld().value => Mondlicht`
- ✓ Z74 `pruefe window.__stilfeldZiel()`
- ✓ Z75 `klicke [data-ziel="stilfeld"]`
- ✓ Z76 `bis window.__zeilen().length === 3`
- ✓ Z77 `zeige window.__zeilen().join(' | ')`
  ```
  "Mondlicht wahlspiel-laterne | Herbstrand wahlspiel-eichel | Flusslauf wahlspiel-kiesel"
  ```
- ✓ Z78 `notiz Suche über den sichtbaren Namen (S07.NAT-02)`
- ✓ Z79 `tippe fluss`
- ✓ Z80 `bis window.__zeilen().length === 1`
- ✓ Z81 `gleich window.__zeilen()[0] => Flusslauf wahlspiel-kiesel`
- ✓ Z82 `abzug suche-name`
  ```
  Bild: suche-name.png
  ```
- ✓ Z83 `notiz Und über die technische Kennung — der Name trägt sie nicht`
- ✓ Z84 `taste cmd+a`
- ✓ Z85 `tippe kiesel`
- ✓ Z86 `bis window.__zeilen().length === 1`
- ✓ Z87 `gleich window.__zeilen()[0] => Flusslauf wahlspiel-kiesel`
- ✓ Z88 `taste enter`
- ✓ Z89 `bis window.__stilfeld().value === 'Flusslauf'`
- ✓ Z90 `abzug suche-kennung`
  ```
  Bild: suche-kennung.png
  ```
- ✓ Z91 `notiz Schritt 3 — der Fokus steht im Namen, und die Wahl steht beim Namen`
- ✓ Z92 `pruefe window.__knopf('Weiter')`
- ✓ Z93 `klicke [data-ziel="knopf"]`
- ✓ Z94 `bis window.__aktuell() === 'Name'`
- ✓ Z95 `bis document.activeElement.labels && document.activeElement.labels.length > 0 && document.activeElement.labels[0].textContent.includes('Name des Sets')`
- ✓ Z96 `tippe Assistentprobe`
- ✓ Z97 `zeige [...document.querySelectorAll('section p')].map((p) => p.textContent).find((t) => t.includes(' · ')) || 'ohne Zusammenfassung'`
  ```
  "probe · 5 Karten · probe-kunst"
  ```
- ✓ Z98 `pruefe [...document.querySelectorAll('section p')].some((p) => p.textContent === 'Wahlspiel · Flusslauf')`
- ✓ Z99 `notiz Zurück behält den Stil (S07.NAT-03)`
- ✓ Z100 `pruefe window.__knopf('Zurück')`
- ✓ Z101 `klicke [data-ziel="knopf"]`
- ✓ Z102 `bis window.__aktuell() === 'Aussehen'`
- ✓ Z103 `gleich window.__stilfeld().value => Flusslauf`
- ✓ Z104 `abzug zurueck-aussehen`
  ```
  Bild: zurueck-aussehen.png
  ```
- ✓ Z105 `notiz Noch einmal zurück — das Spiel steht, wie es gewählt war`
- ✓ Z106 `pruefe window.__knopf('Zurück')`
- ✓ Z107 `klicke [data-ziel="knopf"]`
- ✓ Z108 `bis window.__aktuell() === 'Art'`
- ✓ Z109 `gleich window.__gewaehlt() => wahlspiel`
- ✓ Z110 `notiz Überspringen führt zum Namen, und der Name ist noch da`
- ✓ Z111 `pruefe window.__knopf('Überspringen')`
- ✓ Z112 `klicke [data-ziel="knopf"]`
- ✓ Z113 `bis window.__aktuell() === 'Name'`
- ✓ Z114 `gleich document.activeElement.value => Assistentprobe`
- ✓ Z115 `abzug uebersprungen`
  ```
  Bild: uebersprungen.png
  ```
- ✓ Z116 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 91 Schritte.
