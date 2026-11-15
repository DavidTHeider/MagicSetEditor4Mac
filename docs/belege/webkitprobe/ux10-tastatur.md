# WebKit-Probe — ux10-tastatur.probe

| | |
| --- | --- |
| Stand | `d33405833` |
| Motor | wk2 — WKWebView, eingerichtet wie in Wails 3 |
| Basis | http://127.0.0.1:56918/ |
| Fenster | 1440 × 900 Punkte, hell |
| macOS | Version 27.0.1 (Build 26A434) |
| WebKit | 22625.1.29.11.28 |
| Beginn | 2026-10-20T20:51:32Z |

- ✓ Z21 `motor wk2`
- ✓ Z22 `lade /`
- ✓ Z23 `frist 20000`
- ✓ Z24 `warte nav`
- ✓ Z25 `wartetext Erste Schritte`
- ✓ Z26 `taste escape`
- ✓ Z27 `pruefe (window.__ring = () => { const e = document.activeElement; const s = getComputedStyle(e); return (s.outlineStyle !== 'none' && parseFloat(s.outlineWidth) > 0) || s.boxShadow !== 'none' }, window.__halt = () => { const e = document.activeElement; return ((e.getAttribute('aria-label') || e.innerText || e.value || e.getAttribute('placeholder') || e.tagName) + '').trim().replace(/\s+/g, ' ').slice(0, 48) }, window.__halte = [], window.__ohneRing = [], window.__merke = () => { const e = document.activeElement; if (e === document.body) return true; const h = window.__halt(); window.__halte.push(h); if (!window.__ring()) window.__ohneRing.push(h); return true }, window.__imInhalt = () => document.activeElement.id === 'inhalt', true)`
- ✓ Z29 `notiz Der erste Halt ist der Sprung zum Inhalt — und Enter führt dorthin`
- ✓ Z30 `taste tab`
- ✓ Z31 `pruefe window.__merke()`
- ✓ Z32 `gleich window.__halt() => Zum Inhalt springen`
- ✓ Z33 `taste enter`
- ✓ Z34 `pause 300`
- ✓ Z35 `pruefe window.__imInhalt()`
- ✓ Z36 `taste tab`
- ✓ Z37 `pruefe window.__merke()`
- ✓ Z42 `zeige window.__halt()`
  ```
  "2 Sets 0 von 2 geprüft"
  ```
- ✓ Z43 `pruefe /^\d+ Sets?\b/.test(window.__halt())`
- ✓ Z45 `notiz karte — zu den Sets, das Set öffnen, den Namen ändern; ⇥ schreibt, und der Speicherstand sagt es`
- ✓ Z46 `taste enter`
- ✓ Z47 `wartetext Jedes offene Set liegt im Bestand`
- ✓ Z48 `pause 800`
- ✓ Z49 `pruefe window.__imInhalt()`
- ✓ Z50 `taste tab`
- ✓ Z51 `pruefe window.__merke()`
- ✓ Z52 `taste tab`
- ✓ Z53 `pruefe window.__merke()`
- ✓ Z58 `gleich window.__halt() => Set öffnen …`
- ✓ Z59 `taste tab`
- ✓ Z60 `pruefe window.__merke()`
- ✓ Z61 `gleich window.__halt() => + Neues Set`
- ✓ Z62 `taste tab`
- ✓ Z63 `pruefe window.__merke()`
- ✓ Z64 `notiz Seit UX-11 (#1683) steht über der Liste die Leiste der Listenansicht — Suche, Ordnung, Ansicht —, und jeder Halt trägt seinen Ring`
- ✓ Z65 `gleich window.__halt() => Name oder Spiel`
- ✓ Z66 `taste tab`
- ✓ Z67 `pruefe window.__merke()`
- ✓ Z68 `taste tab`
- ✓ Z69 `pruefe window.__merke()`
- ✓ Z73 `gleich document.activeElement.closest('[role=radiogroup]') !== null => true`
- ✓ Z74 `taste tab`
- ✓ Z75 `pruefe window.__merke()`
- ✓ Z76 `gleich window.__halt() => Werkstatt-Testkarten — Karten dieses Sets zeigen`
- ✓ Z77 `taste enter`
- ✓ Z78 `wartetext Wie viel du siehst`
- ✓ Z79 `pause 1500`
- ✓ Z80 `pruefe window.__imInhalt()`
- ✓ Z81 `taste tab`
- ✓ Z82 `pruefe window.__merke()`
- ✓ Z83 `taste tab`
- ✓ Z84 `pruefe window.__merke()`
- ✓ Z85 `taste tab`
- ✓ Z86 `pruefe window.__merke()`
- ✓ Z87 `taste tab`
- ✓ Z88 `pruefe window.__merke()`
- ✓ Z89 `taste tab`
- ✓ Z90 `pruefe window.__merke()`
- ✓ Z91 `taste tab`
- ✓ Z92 `pruefe window.__merke()`
- ✓ Z93 `taste tab`
- ✓ Z94 `pruefe window.__merke()`
- ✓ Z95 `taste tab`
- ✓ Z96 `pruefe window.__merke()`
- ✓ Z97 `taste tab`
- ✓ Z98 `pruefe window.__merke()`
- ✓ Z99 `taste tab`
- ✓ Z100 `pruefe window.__merke()`
- ✓ Z101 `taste tab`
- ✓ Z102 `pruefe window.__merke()`
- ✓ Z103 `taste tab`
- ✓ Z104 `pruefe window.__merke()`
- ✓ Z105 `taste tab`
- ✓ Z106 `pruefe window.__merke()`
- ✓ Z107 `taste tab`
- ✓ Z108 `pruefe window.__merke()`
- ✓ Z109 `taste tab`
- ✓ Z110 `pruefe window.__merke()`
- ✓ Z111 `taste tab`
- ✓ Z112 `pruefe window.__merke()`
- ✓ Z113 `taste tab`
- ✓ Z114 `pruefe window.__merke()`
- ✓ Z115 `taste tab`
- ✓ Z116 `pruefe window.__merke()`
- ✓ Z117 `taste tab`
- ✓ Z118 `pruefe window.__merke()`
- ✓ Z119 `taste tab`
- ✓ Z120 `pruefe window.__merke()`
- ✓ Z121 `taste tab`
- ✓ Z122 `pruefe window.__merke()`
- ✓ Z123 `taste tab`
- ✓ Z124 `pruefe window.__merke()`
- ✓ Z125 `taste tab`
- ✓ Z126 `pruefe window.__merke()`
- ✓ Z127 `taste tab`
- ✓ Z128 `pruefe window.__merke()`
- ✓ Z129 `taste tab`
- ✓ Z130 `pruefe window.__merke()`
- ✓ Z131 `taste tab`
- ✓ Z132 `pruefe window.__merke()`
- ✓ Z133 `taste tab`
- ✓ Z134 `pruefe window.__merke()`
- ✓ Z135 `taste tab`
- ✓ Z136 `pruefe window.__merke()`
- ✓ Z137 `taste tab`
- ✓ Z138 `pruefe window.__merke()`
- ✓ Z139 `taste tab`
- ✓ Z140 `pruefe window.__merke()`
- ✓ Z146 `taste tab`
- ✓ Z147 `pruefe window.__merke()`
- ✓ Z148 `taste tab`
- ✓ Z149 `pruefe window.__merke()`
- ✓ Z150 `taste tab`
- ✓ Z151 `pruefe window.__merke()`
- ✓ Z152 `taste tab`
- ✓ Z153 `pruefe window.__merke()`
- ✓ Z154 `zeige window.__halte.slice(-8).join(' → ')`
  ```
  "Manakosten prüfen → Breite von „Felder und Prüfung“ → Normal → Erweitert → Experte → Kunst — Bild wählen → kunst → name"
  ```
- ✓ Z155 `gleich window.__halt() => name`
- ✓ Z156 `pruefe (window.__vorher = document.activeElement.value, true)`
- ✓ Z157 `taste cmd+a`
- ✓ Z158 `tippe Ohne Maus`
- ✓ Z159 `taste tab`
- ✓ Z160 `pruefe window.__merke()`
- ✓ Z161 `bis [...document.querySelectorAll('[role=status]')].some((e) => e.textContent.includes('Jede Änderung steht im Bestand.'))`
- ✓ Z162 `abzug karte-ohne-maus`
  ```
  Bild: karte-ohne-maus.png
  ```
- ✓ Z163 `notiz Der Name kommt zurück — die Drehbücher danach laufen auf demselben Bestand`
- ✓ Z164 `taste shift+tab`
- ✓ Z165 `gleich document.activeElement.value => Ohne Maus`
- ✓ Z166 `pruefe (Object.getOwnPropertyDescriptor(Object.getPrototypeOf(document.activeElement), 'value').set.call(document.activeElement, window.__vorher), document.activeElement.dispatchEvent(new Event('input', { bubbles: true })), true)`
- ✓ Z167 `taste tab`
- ✓ Z168 `pause 1500`
- ✓ Z170 `notiz deck — ⌘4, das Probe-Deck über Tab und Enter aufschlagen`
- ✓ Z171 `taste cmd+4`
- ✓ Z172 `wartetext Probe-Deck`
- ✓ Z173 `pause 800`
- ✓ Z174 `pruefe window.__imInhalt()`
- ✓ Z175 `taste tab`
- ✓ Z176 `pruefe window.__merke()`
- ✓ Z177 `taste tab`
- ✓ Z178 `pruefe window.__merke()`
- ✓ Z179 `taste tab`
- ✓ Z180 `pruefe window.__merke()`
- ✓ Z181 `taste tab`
- ✓ Z182 `pruefe window.__merke()`
- ✓ Z183 `taste tab`
- ✓ Z184 `pruefe window.__merke()`
- ✓ Z185 `taste tab`
- ✓ Z186 `pruefe window.__merke()`
- ✓ Z187 `taste tab`
- ✓ Z188 `pruefe window.__merke()`
- ✓ Z189 `taste tab`
- ✓ Z190 `pruefe window.__merke()`
- ✓ Z191 `taste tab`
- ✓ Z192 `pruefe window.__merke()`
- ✓ Z196 `gleich window.__halt() => Deck anlegen`
- ✓ Z197 `taste tab`
- ✓ Z198 `pruefe window.__merke()`
- ✓ Z199 `gleich window.__halt() => Name oder Spiel`
- ✓ Z200 `taste tab`
- ✓ Z201 `pruefe window.__merke()`
- ✓ Z202 `taste tab`
- ✓ Z203 `pruefe window.__merke()`
- ✓ Z204 `gleich document.activeElement.closest('[role=radiogroup]') !== null => true`
- ✓ Z205 `taste tab`
- ✓ Z206 `pruefe window.__merke()`
- ✓ Z207 `gleich window.__halt() => Öffnen: Probe-Deck`
- ✓ Z208 `taste enter`
- ✓ Z209 `bis !document.querySelector('main').innerText.includes('Kein Deck aufgeschlagen')`
- ✓ Z211 `notiz spielen — über den Start zurück, „Partie anfangen" über Tab und Enter`
- ✓ Z212 `taste cmd+1`
- ✓ Z213 `wartetext Was Aufmerksamkeit braucht`
- ✓ Z214 `taste cmd+4`
- ✓ Z215 `wartetext Probe-Deck`
- ✓ Z216 `pause 800`
- ✓ Z217 `pruefe window.__imInhalt()`
- ✓ Z218 `taste tab`
- ✓ Z219 `pruefe window.__merke()`
- ✓ Z220 `taste tab`
- ✓ Z221 `pruefe window.__merke()`
- ✓ Z222 `taste tab`
- ✓ Z223 `pruefe window.__merke()`
- ✓ Z225 `taste tab`
- ✓ Z226 `pruefe window.__merke()`
- ✓ Z227 `gleich window.__halt() => Partie anfangen`
- ✓ Z228 `taste enter`
- ✓ Z230 `wartetext Bei jemandem mitspielen …`
- ✓ Z232 `notiz sammlung — ⌘3, suchen, und die Zahl steht als Status da`
- ✓ Z233 `notiz Seit UXR-10 (#1918) stehen im Kopf »Einlesen und ausgeben …« und »+ Aufnehmen …« — zwei Halte mehr vor dem Suchfeld`
- ✓ Z234 `taste cmd+3`
- ✓ Z235 `wartetext Einlesen und ausgeben …`
- ✓ Z236 `pause 800`
- ✓ Z237 `pruefe window.__imInhalt()`
- ✓ Z238 `taste tab`
- ✓ Z239 `pruefe window.__merke()`
- ✓ Z240 `taste tab`
- ✓ Z241 `pruefe window.__merke()`
- ✓ Z242 `taste tab`
- ✓ Z243 `pruefe window.__merke()`
- ✓ Z244 `taste tab`
- ✓ Z245 `pruefe window.__merke()`
- ✓ Z246 `taste tab`
- ✓ Z247 `pruefe window.__merke()`
- ✓ Z251 `taste tab`
- ✓ Z252 `pruefe window.__merke()`
- ✓ Z253 `taste tab`
- ✓ Z254 `pruefe window.__merke()`
- ✓ Z255 `gleich window.__halt() => Name, Set, Kennung oder Notiz`
- ✓ Z256 `tippe scryfall:a`
- ✓ Z257 `bis [...document.querySelectorAll('[role=status]')].some((e) => /^1 von \d+$/.test(e.textContent.trim()))`
- ✓ Z258 `zeige [...document.querySelectorAll('[role=status]')].map((e) => e.textContent.trim()).filter((t) => / von /.test(t)).join(' | ')`
  ```
  "2 von 2 | 1 von 1 | 1 von 5"
  ```
- ✓ Z259 `taste cmd+a`
- ✓ Z260 `taste backspace`
- ✓ Z261 `bis document.activeElement.value === ''`
- ✓ Z263 `notiz handeln — ⌘5, „Neuer Tausch …" und darin das Gegenüber über Tab und Enter aufschlagen`
- ✓ Z269 `taste cmd+5`
- ✓ Z270 `wartetext Mit wem`
- ✓ Z271 `pause 800`
- ✓ Z272 `pruefe window.__imInhalt()`
- ✓ Z273 `taste tab`
- ✓ Z274 `pruefe window.__merke()`
- ✓ Z275 `taste tab`
- ✓ Z276 `pruefe window.__merke()`
- ✓ Z277 `taste tab`
- ✓ Z278 `pruefe window.__merke()`
- ✓ Z279 `taste tab`
- ✓ Z280 `pruefe window.__merke()`
- ✓ Z281 `gleich window.__halt() => Neuer Tausch …`
- ✓ Z282 `taste enter`
- ✓ Z283 `bis [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())`
- ✓ Z284 `pause 300`
- ✓ Z285 `pruefe window.__merke()`
- ✓ Z286 `taste tab`
- ✓ Z287 `pruefe window.__merke()`
- ✓ Z288 `taste tab`
- ✓ Z289 `pruefe window.__merke()`
- ✓ Z290 `gleich window.__halt() => Bo aufschlagen`
- ✓ Z291 `taste enter`
- ✓ Z292 `bis ![...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())`
- ✓ Z293 `pause 300`
- ✓ Z294 `gleich window.__halt() => Neuer Tausch …`
- ✓ Z295 `pruefe window.__merke()`
- ✓ Z297 `notiz paket — ⌘6, im Katalog suchen, die Zahl als Status`
- ✓ Z298 `taste cmd+6`
- ✓ Z300 `wartetext Installierte Pakete`
- ✓ Z301 `pause 800`
- ✓ Z302 `zeige document.activeElement.tagName + '#' + document.activeElement.id + ' ' + (document.activeElement.getAttribute('aria-label') || document.activeElement.innerText || '').trim().slice(0, 60)`
  ```
  "MAIN#inhalt Pakete"
  ```
- ✓ Z303 `pruefe window.__imInhalt()`
- ✓ Z304 `taste tab`
- ✓ Z305 `pruefe window.__merke()`
- ✓ Z306 `taste tab`
- ✓ Z307 `pruefe window.__merke()`
- ✓ Z308 `taste tab`
- ✓ Z309 `pruefe window.__merke()`
- ✓ Z310 `taste tab`
- ✓ Z311 `pruefe window.__merke()`
- ✓ Z312 `taste tab`
- ✓ Z313 `pruefe window.__merke()`
- ✓ Z314 `taste tab`
- ✓ Z315 `pruefe window.__merke()`
- ✓ Z318 `taste tab`
- ✓ Z319 `pruefe window.__merke()`
- ✓ Z320 `zeige window.__halte.slice(-8).join(' → ')`
  ```
  "Neuer Tausch … → Vorlagen → Paket beziehen … → Installiert 11 → Paket einbinden … → Navigator ausblenden → Inspector ausblenden → Name, Art, Herkunft oder Gruppe"
  ```
- ✓ Z321 `gleich window.__halt() => Name, Art, Herkunft oder Gruppe`
- ✓ Z322 `tippe magic`
- ✓ Z323 `pause 800`
- ✓ Z324 `zeige [...document.querySelectorAll('[role=status]')].map((e) => e.textContent.trim()).filter((t) => / von /.test(t)).join(' | ')`
  ```
  "2 von 2 | 1 von 1 | 5 von 5 | 0 von 11 Paketen"
  ```
- ✓ Z325 `taste cmd+a`
- ✓ Z326 `taste backspace`
- ✓ Z327 `bis document.activeElement.value === ''`
- ✓ Z329 `notiz profil — ⌘', die Laschen mit den Pfeiltasten`
- ✓ Z330 `taste cmd+'`
- ✓ Z331 `wartetext Persönliches`
- ✓ Z332 `pause 800`
- ✓ Z333 `pruefe window.__imInhalt()`
- ✓ Z334 `taste tab`
- ✓ Z335 `pruefe window.__merke()`
- ✓ Z336 `taste tab`
- ✓ Z337 `pruefe window.__merke()`
- ✓ Z339 `gleich window.__halt() => + Profil anlegen …`
- ✓ Z340 `taste tab`
- ✓ Z341 `pruefe window.__merke()`
- ✓ Z343 `pruefe window.__halt().startsWith('Mehr zu')`
- ✓ Z346 `taste tab`
- ✓ Z347 `pruefe window.__merke()`
- ✓ Z348 `gleich window.__halt() => Zu Cleo · Mein Profil wechseln`
- ✓ Z349 `taste tab`
- ✓ Z350 `pruefe window.__merke()`
- ✓ Z351 `pruefe window.__halt().startsWith('Mehr zu')`
- ✓ Z352 `taste tab`
- ✓ Z353 `pruefe window.__merke()`
- ✓ Z354 `gleich window.__halt() => Cleo werden`
- ✓ Z355 `taste tab`
- ✓ Z356 `pruefe window.__merke()`
- ✓ Z357 `gleich window.__halt() => Allgemeines`
- ✓ Z358 `taste right`
- ✓ Z359 `pause 500`
- ✓ Z360 `zeige 'Lasche: ' + window.__halt() + ' ausgewählt=' + document.activeElement.getAttribute('aria-selected')`
  ```
  "Lasche: Konto ausgewählt=true"
  ```
- ✓ Z362 `notiz Die Folge der Halte — und keiner ohne sichtbaren Ring`
- ✓ Z363 `zeige JSON.stringify(window.__halte)`
  ```
  "[\"Zum Inhalt springen\",\"2 Sets 0 von 2 geprüft\",\"Erklärung zu Sets\",\"Set öffnen …\",\"+ Neues Set\",\"Name oder Spiel\",\"Wie im Bestand Nach Name Nach Kartenzahl\",\"kacheln\",\"Werkstatt-Testkarten — Karten dieses Sets zeigen\",\"Eigenschaften\",\"Werkzeuge\",\"Ausgeben\",\"Karten 5\",\"Karten\",\"+ Karte anlegen Mehr Navigator ausblenden Inspec\",\"+ Karte anlegen\",\"Mehr\",\"Navigator ausblenden\",\"Inspector ausblenden\",\"Filtern — z. B. type:creature -goblin\",\"Erklärung zu Karten filtern\",\"Mehrere wählen\",\"Aschenwaechter\",\"Arbeitsstand dieser Karte\",\"Spalten der Liste\",\"Breite von „Karten des Sets“\",\"Nächste Karte (⇟)\",\"Diese Karte\",\"Größe der Vorschau\",\"on\",\"on\",\"on\",\"Mehr zu „Zeigt, was sich ändern würde.“\",\"Vorschlag holen\",\"Hinweise holen\",\"Manakosten prüfen\",\"Breite von „Felder und Prüfung“\",\"Normal\",\"Erweitert\",\"Experte\",\"Kunst — Bild wählen\",\"kunst\",\"name\",\"Name — Vorschlag\",\"Erklärung zu Decks\",\"Turniere\",\"Formate …\",\"Partie anfangen\",\"Bo einladen\",\"Bei Bo mitspielen\",\"Navigator ausblenden\",\"Inspector ausblenden\",\"Deck anlegen\",\"Name oder Spiel\",\"Zuletzt benutzt zuerst Nach Name Nach Kartenzahl\",\"liste\",\"Öffnen: Probe-Deck\",\"Erklärung zu Decks\",\"Turniere\",\"Formate …\",\"Partie anfangen\",\"Erklärung zu Sammlung\",\"Einlesen und ausgeben …\",\"+ Aufnehmen …\",\"Exemplare 11 5 Einträge\",\"Navigator ausblenden\",\"Inspector ausblenden\",\"Name, Set, Kennung oder Notiz\",\"Erklärung zu Handeln\",\"Mit wem\",\"Neuer Tausch … Was gerade läuft 1 Jeder Handel d\",\"Neuer Tausch …\",\"Schließen\",\"Mehr zu „Gehandelt wird nur mit einem Gegenüber,\",\"Bo aufschlagen\",\"Neuer Tausch …\",\"Vorlagen\",\"Paket beziehen …\",\"Installiert 11\",\"Paket einbinden …\",\"Navigator ausblenden\",\"Inspector ausblenden\",\"Name, Art, Herkunft oder Gruppe\",\"Erklärung zu Nutzer und Profile\",\"+ Profil anlegen …\",\"Mehr zu „Sets, Decks, Sammlung, Turniere und Ein\",\"Zu Cleo · Mein Profil wechseln\",\"Mehr zu „In ein fremdes Profil zu sehen ändert n\",\"Cleo werden\",\"Allgemeines\"]"
  ```
- ✓ Z364 `gleich JSON.stringify(window.__ohneRing) => []`
- ✓ Z365 `gleich JSON.stringify(window.__probeFehler || []) => []`

**Ergebnis: bestanden** — 300 Schritte.
