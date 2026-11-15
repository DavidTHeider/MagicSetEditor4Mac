# Freigabe

Releases entstehen **lokal** über `scripts/release.sh`. Einen entfernten Lauf
gibt es nicht ([`entscheidungen/0007`](entscheidungen/0007-kein-entfernter-lauf.md)).
Ziel ist ausschließlich macOS auf Apple Silicon.

**Freigegeben wird erst, wenn alle Phasen durch sind** (`AGENTS.md` §12) —
nicht am Ende jeder Phase. **1.0.0 ist am 2026-11-15 freigegeben worden** —
die erste Freigabe für Anwendende: der Programmstand von 0.9.0, neu gebaut mit
zwei Sicherheitsaktualisierungen und einem eigenen Freigabeschlüssel, in einem
neu angelegten Repositorium unter demselben Namen
([0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). **Davor entstanden
neun Entwicklungsfassungen**, 0.1.0 am 2026-08-22 bis 0.9.0 am 2026-10-21;
sie sind nicht mehr veröffentlicht, ihre Freigaben liegen im
Entwicklungsarchiv. Die nächste Freigabe ist **1.1.0**; ihr Zuschnitt ist
**noch nicht gemacht**.

**Diese zwei Zahlen hält seit dem 2026-08-29 ein Wächter** gegen die Roadmap
und das Änderungsprotokoll (`internal/werkzeuge/standderfreigabe_test.go`).
Hier stand „Die nächste Freigabe ist 0.3.0", während 0.3.0 seit einem Tag
draußen war — und in der Roadmap stand es richtig. **Zwei Dateien, zwei
Antworten**, angestrichen vom Zweitblick über Phase 4.0 (Befund 5).

**Hier stand bis zum 2026-08-21 „Phase 8 bleibt die letzte in der
Bauabfolge"** — das galt für die alte Neunernummerierung und war überholt,
seit die Phasen Buchstaben tragen. Angestrichen vom Zweitblick durch Codex.

---

## Vor jeder Freigabe

Diese Liste ist nicht optional. Jeder Punkt hat einen Grund, und die Gründe
stehen dabei.

| Schritt                                                         | Warum                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Der MSE2-Bestand**                                            | **Bleibt die Obermenge vollständig?** `docs/MSE2BESTAND.md` führt die Menübefehle von MSE2 — die Liste, gegen die „echtes Subset" überhaupt geprüft werden kann (#1291).<br>`MSE2_QUELLE=/…/mse2-quelle go test -tags fremdpakete ./internal/werkzeuge/ -run Mse2Bestand -v`<br>**Er läuft hinter `fremdpakete`**, weil er den fremden Quelltext braucht.<br>**Was er nicht hält**: ob die Einordnung stimmt — ob ein ✓ verdient ist, liest ein Mensch. Und nichts jenseits des Menüs: Kontextmenüs, Panels, Ziehen und Ablegen stehen in keiner der beiden Listen. |
| Die Läufe gegen **bezogene** Daten                              | Sie sind der einzige Beleg, dass die Leser echte Dateien lesen — unsere Testdaten prüfen, ob sie tun, was wir für richtig halten. Drei Befunde der Phase 9 kamen von hier und von nirgends sonst.<br>`MSE_FREMDPAKETE=… go test -tags fremdpakete ./internal/msepaket/ -run Fremd -v`<br>`MSE_MTGJSON=… go test -tags fremdpakete ./internal/mtgjson/ -run Fremd -v` — **einmal mit einer Datei, einmal mit einem Verzeichnis**.<br>**`MSE_FREMDPAKETE` zeigt auf das Verzeichnis, das die Pakete _enthält_** — bei `M15-Magic-Pack` also auf `…/M15-Magic-Pack/data` und nicht auf die Sammlung selbst. Daneben gezeigt meldet der Lauf „der Lauf greift ins Leere" statt grün zu werden; die Schranke greift also, aber sie kostet einen Anlauf                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **Der Kontrastlauf**                                             | **Kontraste am gerenderten Element, in beiden Erscheinungsbildern** (#1290). §3 sagt, warum es ihn braucht: unter jsdom meldet axe `color-contrast` als **unentschieden**.<br>`node scripts/kontrastlauf.mjs`<br>**Gemessen wird über einen Canvas**, nicht über ein Textmuster — Tailwind 4 gibt `oklch()`, und wer daraus Ziffern zieht, liest Unsinn. Der Hintergrund kommt vom nächsten Vorfahren, der einen hat.<br>**Erhoben am 2026-08-30**: **4.526** Textstellen über sechs Seiten, hell und dunkel — **null** unter der Schranke. Die Gegenprobe (Überschriften auf `#a8b0bb`) macht 28 davon rot, und zwar **nur im hellen** Erscheinungsbild.<br>**Was er nicht misst**: ob die Farbe die richtige ist, Ränder ohne Fläche (WCAG 1.4.11 — die führt `docs/KONTRASTE.md` von Hand) und Text auf einem Bild. |
| **Der Pergamentlauf**                                            | **Was die Textur mit den Kontrasten macht** (#1325, Werkstück J). `KONTRASTE.md` führt Paare aus zwei **Farben**, und `kontrastlauf.mjs` liest über `getComputedStyle` ebenfalls eine Farbe. Seit über Grund und Fläche eine Faser liegt, hat der Hintergrund keine *eine* Farbe mehr — er hat eine hellste und eine dunkelste Stelle.<br>`node scripts/pergamentmass.mjs`<br>Er rastert die Kachel aus `frontend/src/pergament-hell.svg` beziehungsweise `-dunkel.svg`, legt sie über jede Grundfarbe und rechnet den Kontrast am **ungünstigsten** der 810.000 Bildpunkte. Die Kachel kommt aus der Datei, die `stil.css` nennt — nicht aus einem Pfad im Skript.<br>**Erhoben am 2026-09-03**: **32 Paare**, keines unter der Schranke; der knappste ist `border-kante` auf `bg-grund` mit 3,99 gegen 3. Der Verlust durch die Textur liegt zwischen 0,45 und 1,58 Punkten.<br>**Was er nicht misst**: ob die Textur schön ist, Text auf einem Bild, die Kartenvorschau (sie trägt keine Faser) und eine Vignette — es gibt keine, sie ist an genau dieser Messung gescheitert. |
| **Der Kernwegslauf**                                            | **Was der Weg von Karte zu Karte kostet** (#1290). `serienmass.mjs` zaehlt, was auf einer Seite **steht**; dieser zaehlt, was jemand **tut**.<br>`node scripts/kernwegmass.mjs`<br>**Er legt sich sein eigenes Set an und raeumt es weg** — er schreibt, und ein Messlauf, der schreibt, fasst keinen fremden Bestand an. Misslingt das Aufraeumen, sagt er es.<br>**Erhoben am 2026-08-30**: **4 Nutzerhandlungen**, kein Seitenwechsel, und beide Karten tragen ihren Namen.<br>**Was er nicht misst**: die Zeit, und die uebrigen Kernwege — Suchen/Ersetzen, Massenaenderung, Bild setzen, Ausgeben, Sammlung, Preis. |
| **Der Kürzellauf**                                               | **Ob ein Tastenkürzel im Fenster wirklich wirkt** (#1291). `tastatur.test.ts` prüft `deuten()` — eine reine Funktion; ob ⌘4 auf die Vorlagen führt, sagt sie nicht.<br>`node scripts/kuerzellauf.mjs`<br>**Drei Messfehler hat er gekostet**: erst `window.location.pathname` (die Anwendung führt ihre Navigation im **Zustand**, nicht in der Adresse), dann die Überschrift (sechs Seiten brauchen ein offenes Set und zeigen sonst „Kein Set offen"), dann der `textContent` des Navigationseintrags — seit die Leiste ihre Tasten zeigt, lautet er „Vorlagen⌘4". Jedes Mal hat erst eine **Kontrolle** es entschieden.<br>**Erhoben am 2026-09-03**: ⌘4 → *Vorlagen*, ⌘5 → *Sammlung*, ⌘6 → *Decks*, ⌘7 → *Turniere*, alle vier ✓.<br>**Was er nicht misst**: die zweite Reihe. Sie führt seit dem 2026-09-03 ausschließlich auf **Laschen eines Sets**, und die verlangen ein offenes — ein Messlauf, der eines öffnet, fasste einen fremden Bestand an. |
| **Der Serienlauf**                                               | **Was eine Seite kostet, gemessen statt geschätzt** (#1290). Er fährt **alle zweiundzwanzig** Seiten bei **1280 × 800** durch die laufende Brücke — neun Anwendungsseiten und die dreizehn Laschen eines offenen Sets — und zählt Tabstellen, Knöpfe, Höhe und was über der Falz steht.<br>`node scripts/serienmass.mjs`<br>**Er läuft nicht in `make ci`** — er baut die Anwendung und öffnet ein Fenster, wie der Nahtlauf. Und er urteilt nicht: was er gibt, sind Zahlen für `docs/SERIENARBEIT.md`.<br>**Was er nicht misst**: ob eine Seite gut aussieht, wie lange etwas dauert, und ob eine Seite vollständig ist. |
| **Die Arbeitswege**                                              | **Was die vier übrigen Kernwege kosten** — Paket korrigieren, Sammlung suchen, Deck bauen, Profil wechseln (#1412, Werkstück U1 von Phase 5.1). Je Weg drei Zahlen: **Handlungen**, **Tabstelle** des ersten Bedienelements und **Ladezeit** auf diesem Rechner.<br>`node scripts/sammlungssaat.mjs /tmp/mse51heim 10000` · `HOME=/tmp/mse51heim make dev` · `node scripts/arbeitswegmass.mjs`<br>Er fährt gegen den Messbestand aus [`UX-ARCHITEKTUR.md`](UX-ARCHITEKTUR.md) — 10.000 Exemplare, ein Set mit fünf Karten, die Saat — und **weist ab**, sobald ein Weg mehr Handlungen zählt als zugesagt; wer einen Arbeitsplatz umbaut, senkt die Zusage in `arbeitswege.mjs` mit. Gemessen am 2026-09-08: **6 · 3 · 11 · 3** Handlungen. |
| **Der Tischlauf**                                                | **Was der Spieltisch kostet** (#1460, Werkstück G3 von Phase 5.5). **Zwei Läufe, zwei Fragen**: was die Rechnung kostet und was die Grafikkarte daraus macht.<br>`node scripts/tischmass.mjs` · `node scripts/tischframezeit.mjs`<br>Der erste fährt 200 Läufe über die Leistungsszene `extreme-board-state` — vier Spieler, 150 Karten, 100 Marken, 50 Spielsteine, 30 Anhängsel, und diese vier Zahlen werden **gezählt** statt geschätzt. Der zweite zeichnet 240 Bilder in einem eigenen Chrome mit `--use-angle=metal` und trennt **Bildabstand** von **Zeichenzeit**: der eine sagt, ob 60 Hz gehalten werden, der andere, wie viel Luft bleibt.<br>**Erhoben am 2026-09-16**: Rechnung 0,171 ms im Median; Zeichnen 0,40 ms, kein Bild von 240 unter 30 Hz, erstes Bild 17,70 ms (Kaltcache). Renderer: **ANGLE Metal Renderer, Apple M5** — er wird ausgelesen, weil headless Chrome sonst SwiftShader nimmt.<br>**Was sie nicht messen**: externes Display, Vollbild und einen aktiven Drag — den gibt es noch nicht. Die Zahlen stehen in [`TISCHLEISTUNG.md`](TISCHLEISTUNG.md). |
| **Der Erstlauf**                                                 | **Was der erste Weg kostet** (#1325). Von der Anwendung bis zur ersten Karte, die ihren Namen trägt.<br>`node scripts/erstlaufmass.mjs`<br>**Er legt sein Set an und räumt es weg** — wie das Kernwegmaß. Misslingt das Aufräumen, endet er rot.<br>**Erhoben am 2026-09-03**: **7 Nutzerhandlungen**.<br>**Was er nicht misst**: den ersten Eindruck, einen Bestand ohne Saat, und die Zeit. |
| **Der Serienabzug**                                              | **Dieselben zweiundzwanzig Seiten als Bild** (#1325), hell und dunkel, bei 1280 × 800.<br>`node scripts/serienabzug.mjs <zielverzeichnis>`<br>**Er ist ein Messmittel und kein Dokumentationsmaterial** — die vierundvierzig Bilder kommen nicht in den Baum; die elf des Handbuchs macht `abzuege.mjs`. Er fährt dieselbe Liste wie der Serienlauf (`serienmasslisten.mjs`): eine Seite, die gemessen und nicht abgelichtet wird, gäbe es sonst.<br>**Was er nicht tut**: urteilen. Das steht in `SEITENKRITIK.md` (Entwicklungsarchiv) und geht nur von Hand. |
| **Die Fuzzläufe**                                                | **Sechzig Ziele in siebenundzwanzig Paketen**; welche Fläche jedes liest und was es zusagt, steht im Register [`FUZZZIELE.md`](FUZZZIELE.md). **Alle auf einmal, mit Budget und Bericht**: `node scripts/fuzzkampagne.mjs` (`make fuzzkampagne`; `ZEIT`, `PARALLEL`, `SPEICHER`), Bericht in [`belege/fuzzkampagne.md`](belege/fuzzkampagne.md). **Die volle Kampagne vom 2026-10-11** (C02, #2227): alle 60 Ziele, **zehn Minuten je Ziel**, fünf Arbeiter, zusammen rund 1,6 Milliarden Ausführungen. **Gesättigt sind acht** — im letzten Viertel kam keine neue Eingabe —, 49 wachsen noch, und drei sind nicht bewertbar, weil ein Fund sie beendete; Sättigung ist ein festgelegtes Maß und kein Beweis, dass nichts mehr zu finden wäre. **Drei Funde, jeder reproduziert, behoben und als Testfall erhalten**: ein Wert mit offener Klammer verschluckte beim Wiederlesen seine Nachbarn (`FuzzSymbolzeichnung`; dieselbe Klasse in der Setdatei, und auf der Leseseite in echten Paketen, #2293), eine Übersetzung mit einem Satz ohne Pfad ging durch (`FuzzSprachpaketEinlesen`), und ein Ordnereintrag sprengte die Vorrichtung von `FuzzInstaller` — ein Fund an ihr, nicht am Leser. Sie lesen alle Daten, die von außen kommen — `msepaket` (Dateien von 2007, halb geschrieben), `bezug` (Archive aus dem Netz, also **feindlich beeinflussbar**), `msescript` (eine Sprache mit Zerteiler und Auswerter), dazu `mtgjson` und `scryfall` (fremde Schemata) und `pdf` (entartete Maße). **Was ein Fund ist, steht in jedem Ziel im Kopf**, festgelegt vor dem ersten Lauf (#1289).<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzZerteile -fuzztime 2m`<br>`go test ./internal/msescript -run Fuzz -fuzz FuzzAuswerten -fuzztime 2m`<br>`go test ./internal/msescript -run Fuzz -fuzz FuzzZerteileSprache -fuzztime 2m`<br>`go test ./internal/bezug -run Fuzz -fuzz FuzzAuspacken -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/scryfall -run Fuzz -fuzz FuzzLiesKarte -fuzztime 2m`<br>`go test ./internal/mtgjson -run Fuzz -fuzz FuzzLiesInhalt -fuzztime 2m`<br>`go test ./internal/pdf -run Fuzz -fuzz FuzzDokument -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzBeschaedigtesPaket -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeAustausch -fuzztime 2m`<br>`go test ./internal/aktualisierung -run Fuzz -fuzz FuzzManifestAus -fuzztime 2m`<br>`go test ./internal/anmeldung -run Fuzz -fuzz FuzzPruefeRueckruf -fuzztime 2m`<br>`go test ./internal/sammelliste -run Fuzz -fuzz FuzzLiesListe -fuzztime 2m`<br>`go test ./internal/sammelliste -run Fuzz -fuzz FuzzLiesCSV -fuzztime 2m`<br>`go test ./internal/rest -run Fuzz -fuzz FuzzDienstruf -fuzztime 2m`<br>`go test ./internal/service -run Fuzz -fuzz FuzzTokenPruefen -fuzztime 2m`<br>`go test ./internal/sicherung -run Fuzz -fuzz FuzzEntpacken -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzLoeseAuf -fuzztime 2m`<br>`go test ./internal/mtgapi -run Fuzz -fuzz FuzzLiesKarten -fuzztime 2m`<br>`go test ./internal/mtggraphql -run Fuzz -fuzz FuzzLies -fuzztime 2m`<br>`go test ./internal/scrydex -run Fuzz -fuzz FuzzLiesKarten -fuzztime 2m`<br>`go test ./internal/mtgjsonpreise -run Fuzz -fuzz FuzzLies -fuzztime 2m`<br>`go test ./internal/mtgjsonpreise -run Fuzz -fuzz FuzzLiesGepackt -fuzztime 2m`<br>`go test ./internal/anthropic -run Fuzz -fuzz FuzzModelle -fuzztime 2m`<br>`go test ./internal/openai -run Fuzz -fuzz FuzzModelle -fuzztime 2m`<br>`go test ./internal/xai -run Fuzz -fuzz FuzzModelle -fuzztime 2m`<br>`go test ./internal/ollama -run Fuzz -fuzz FuzzModelle -fuzztime 2m`<br>`go test ./internal/cardtrader -run Fuzz -fuzz FuzzAngebote -fuzztime 2m`<br>`go test ./internal/leitung -run Fuzz -fuzz FuzzRahmen -fuzztime 2m`<br>`go test ./internal/peer -run Fuzz -fuzz FuzzAuspacke -fuzztime 2m`<br>`go test ./internal/peer -run Fuzz -fuzz FuzzFingerabdruck -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeKontakteinladung -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeEinloesung -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeEinloeseantwort -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeEinladung -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzZerlegeAdresse -fuzztime 2m`<br>`go test ./internal/entdeckung -run Fuzz -fuzz FuzzDeute -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzLiesKontrolle -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzPartienachricht -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzEmpfaengerstand -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzChat -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzWiederaufnahme -fuzztime 2m`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzHandelsumschlag -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzInstaller -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzPaketpfad -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzEinbinden -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzSpielmodell -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzStilmodell -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzDekodiereBild -fuzztime 2m`<br>`go test ./internal/msepaket -run Fuzz -fuzz FuzzSymbolzeichnung -fuzztime 2m`<br>`go test ./internal/austausch -run Fuzz -fuzz FuzzUmschlag -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/service -run Fuzz -fuzz FuzzSprachpaketEinlesen -fuzztime 2m`<br>`go test ./internal/service -run Fuzz -fuzz FuzzZugfolge -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/domain -run Fuzz -fuzz FuzzHandelsfolge -fuzztime 2m`<br>`go test ./internal/service -run Fuzz -fuzz FuzzBildangabe -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/msescript -run Fuzz -fuzz FuzzTextskript -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/service -run Fuzz -fuzz FuzzKartenskript -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/sicherung -run Fuzz -fuzz FuzzEntpackenVerschluesselt -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/app -run Fuzz -fuzz FuzzUmschlagdeck -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/peer -run Fuzz -fuzz FuzzHandschlag -fuzztime 2m -fuzzminimizetime 5s`<br>`go test ./internal/service -run Fuzz -fuzz FuzzSpielbefehl -fuzztime 2m -fuzzminimizetime 5s`<br>**Die sieben letzten kamen mit QA-06c dazu** (#1976, 2026-10-01) und schließen die hohen Flächen der **Paketseite** und den **Zustand**: Bildquelle, Text- und Kartenskripte, die verschlüsselte Sicherung, den Umschlaginhalt, den Handschlag als Folge und den Spielbefehl als Ganzes. Die Kampagne lief **30 Sekunden je Ziel mit zwei Arbeitern** ([`belege/fuzzkampagne-qa06c.md`](belege/fuzzkampagne-qa06c.md)): von 906.967 Ausführungen (`FuzzTextskript`, im Nachtest) bis 14.175 (`FuzzUmschlagdeck`, das je Ausführung einen Bestand anlegt). **Seitdem nennt jeder Bericht je Ziel einen Anteil** — der Anweisungen seines Pakets, den das Korpus betritt (QA-06.AK-04, [`FUZZZIELE.md`](FUZZZIELE.md#der-anteil-je-ziel--qa-06ak-04)). **Fünf Fehler in der Anwendung, alle behoben und mit benanntem Lauf, und keinen fand ein Lauf — alle fanden sich beim Festlegen der Zusagen**: eine fremde Sicherung bestimmte selbst, wie lange gerechnet wird (age nimmt ohne Deckel bis 2^22, rund vier GiB vor der Kennwortprüfung); ein Kartenskript legte jede Abbildung als Karte an, auch ohne `new_card` und mit einem Bildfeld, das `new_card` verbietet (MSE2 weist sie ab, `add_cards_script.cpp` Zeile 33); es lief ohne Frist und ohne Grenze der Kartenzahl; und ein Stil konnte mit `setpaket:` einen beliebigen Ort auf der Platte als Bildquelle nennen. **Der fünfte kam beim Festlegen der Zusage für den Umschlaginhalt** (#1995): ein Deck verlor nach dem ersten Platz, der sich nicht legen ließ, alle weiteren still. **Dazu eine Zusage, die zu eng war**: `FuzzTextskript` verlangte, dass ein Text ohne Ausdruck unverändert zurückkommt, und `"\xfa"` kam als Ersatzzeichen — eine Paketdatei mit ungültigem UTF-8 weist aber schon `msepaket` mit Stelle ab, wie MSE2; die Zusage gilt jetzt für gültiges UTF-8, die Eingabe bleibt als Saat<br>**Die elf letzten kamen mit QA-06b dazu** (#1967, 2026-10-01) und schließen die **Paketseite**: Installer, Paketpfade, Einbinden, Spiel- und Stilmodell, Bilder, Symbole, Umschlag samt Manifest und Sprachpaket, dazu zwei **Folgen mit Zustand** — Priorität und Stapel (`FuzzZugfolge`) und Handelsangebot (`FuzzHandelsfolge`). Die Kampagne lief **30 Sekunden je Ziel mit zwei Arbeitern** ([`belege/fuzzkampagne-paketseite.md`](belege/fuzzkampagne-paketseite.md), dazu der Erstlauf und der Nachtest daneben): von 1.202.627 Ausführungen (`FuzzEinbinden`) bis **493** (`FuzzUmschlag`, das je Ausführung unterschreibt, packt und zweimal auspackt) und 11.683 (`FuzzInstaller`, ein Archiv auf der Platte). **Acht Fehler in der Anwendung, alle behoben und mit benanntem Lauf.** **Fünf fand ein Lauf** — die Kampagne einen Paketverweis ohne gültiges UTF-8, acht Byte TIFF, die zehn Megabyte zuteilten, weil der Kodierer keinen wahlfreien Zugriff bekam, und einen Bildwert, der beim Zerlegen seinen Leerraum am Ende verlor; der Saatlauf ein Objekt namens `.`, das an die Stelle des Verzeichnisses ausgepackt wurde, und einen Handelsvorschlag nach der Einigung. **Drei fanden sich beim Festlegen der Zusagen**, und kein Lauf von 30 Sekunden hätte sie erreicht: `include_file:` band nicht ein, obwohl MSE2 es so liest; ein Symbol, das geschrieben über 1 MiB käme, ließ sich ablegen und nicht wieder öffnen; und ein unterschriebener Umschlag mit falscher Größe trug beliebig viel durch die Quarantäne. **Dazu zwei Zusagen, die zu eng waren** und berichtigt sind: 4 Byte je Bildpunkt (die Grenze zählt Bildpunkte, ein PNG mit 16 Bit je Kanal braucht 8, ein progressives JPEG bis zu 24) und `ErrNotFound` als Klasse einer Ablehnung. **Offen**: ob die Bildgrenze in Byte zählen soll (#1974). Eine Mehrspielerpartie, in der ein Gehender die Priorität behielt, ist seit #1971 behoben (CR 800.4a und 800.4j), und `FuzzZugfolge` fährt seitdem drei Sitze<br>**Die fünfzehn davor kamen mit QA-06 dazu** (#1737, 2026-10-01) und schließen die Lücke am **Draht**: Rahmen, Hülle, Kennung, Einladungen, Einlösung, Adresse, Entdeckung, Kontrolle, Partienachricht, Folgen, Satz am Tisch, Wiederaufnahme und Handelsumschlag hatten keines. Ihre Kampagne steht in [`belege/fuzzkampagne.md`](belege/fuzzkampagne.md).<br>**Die zehn letzten kamen mit A19 dazu** (#1550, 2026-09-13): sie lesen, was die **neun Blätter** von A1–A14 von fremden Diensten entgegennehmen — vier über ihren exportierten Zerteiler, fünf über eine `httptest`-Gegenstelle, weil sie keinen haben. **Der zehnte ist der gzip-Weg von `mtgjsonpreise`**: eine Packbombe und ein abgeschnittener Strom sind der Missbrauchsfall dieses Blatts, und der Deckel liegt auf dem **ausgepackten** Strom. Erhoben am 2026-09-13, je Ziel zwei Minuten: 15.746.826 Ausführungen (`FuzzLiesKarten`/mtgapi), 13.231.379 (`FuzzLies`/mtggraphql), 12.603.169 (`FuzzLiesKarten`/scrydex), 11.876.243 (`FuzzLiesGepackt`), 5.924.064 (`FuzzLies`/mtgjsonpreise), 1.150.175 (`FuzzAngebote`), 1.062.582 (`FuzzModelle`/anthropic), 455.785 (`FuzzModelle`/ollama), 271.328 (`FuzzModelle`/xai), 206.390 (`FuzzModelle`/openai).<br>**Die fünf Transportziele liegen ein bis zwei Größenordnungen tiefer, und das ist keine Nachlässigkeit**: sie fahren je Ausführung eine echte HTTP-Anfrage gegen eine Gegenstelle im selben Prozess. **Einen Server je Ausführung gab es zuerst, und er war der Fehler**: bei zehn Arbeitern gingen die Ports aus (`dial tcp: resource temporarily unavailable`), und der Lauf meldete daraufhin »zweimal dieselbe Antwort, zwei Ergebnisse« — ein Fund über das Betriebssystem und keiner über das Blatt. Jetzt gibt es **einen** Server je Fuzzprozess.<br>**Zwei Funde in der Anwendung.** Der erste ist eine **Klasse über vier Blätter**: ein Modelleintrag ohne Kennung gab in `anthropic`, `openai`, `xai` und `ollama` einen Befund, der sich nicht benennen konnte — zwei solche Einträge, zwei **wortgleiche** Sätze, und §8 verlangt für jeden Fehler die Stelle. Jeder Befund trägt jetzt `data[i]` bzw. `models[i]`. Der zweite steht in `cardtrader`: ein **Angebot ohne `blueprint_id`** kam als vollwertiges Angebot durch und wäre über `AlsStand` zu einer Preisreihe geworden — für einen Druck, nach dem niemand gefragt hat; geprüft wurde bis dahin allein die Währung. Beide behoben, beide mit benanntem Lauf und Gegenprobe.<br>**Dazu zwei Fehlgriffe in den Zielen selbst**, die derselbe Lauf angestrichen hat und die **nicht** als Fund an der Anwendung zählen: eine Gegenstelle, die für 204 und 101 einen Rumpf schreiben wollte (HTTP erlaubt dort keinen), und der Server je Ausführung von oben<br>**Die drei letzten kamen im Nachgang zu O11 dazu** (Befund 3 des Zweitblicks über Phase 5.2, 2026-09-11): das Issue nennt **Token**, **Sicherungsmanifest** und **Referenzen** ausdrücklich, und sie fehlten. Erhoben am selben Tag, je zwei Minuten: 33.680.341 Ausführungen (`FuzzLoeseAuf`) und 10.575.440 (`FuzzTokenPruefen`); `FuzzEntpacken` steht im nächsten Satz.<br>**Das letzte lag drei Größenordnungen tiefer**: es schreibt je Ausführung ein Archiv auf die Platte und packt es **zweimal** aus — dieselbe Lage wie bei `FuzzBeschaedigtesPaket`, und **dieselbe Falle beim Minimieren** (unten). Seine 58.145 vom selben Tag sind deshalb ersetzt: am 2026-09-23 mit der Vorgabe 2.786 Ausführungen und 12 neue Eingaben, mit `-fuzzminimizetime 5s` **23.854 und 168**.<br>**Was sie zusagen, steht in ihren Köpfen**, und zwei davon sind enger als bloß „abgewiesen“: bei `FuzzTokenPruefen` muss **jede** Abweisung außer der am Scope gleich aussehen (jede Unterscheidung wäre ein Orakel), und bei `FuzzEntpacken` darf **keine** Datei außerhalb des genannten Verzeichnisses landen.<br>**Die sechs letzten kamen mit O11 dazu** (#1435, 2026-09-11): sie lesen, was die Phasen 5.1 und 5.2 an neuen Flächen aufgemacht haben — den **Einstellungsaustausch** (O3), das **Kompatibilitätsmanifest** (O6), den **OAuth-Rückruf** auf der Schleifenadresse, die **Kartenliste** und das **Sammlungs-CSV** sowie den **Dienst als Ganzes** (Marke, Herkunft, Drossel, Entzifferung und Antwortschreiber in einem Ruf). Erhoben am 2026-09-11, je Ziel zwei Minuten: 18.246.583 Ausführungen (`FuzzLiesListe`), 15.251.277 (`FuzzPruefeRueckruf`), 15.180.242 (`FuzzLiesCSV`), 14.030.957 (`FuzzManifestAus`), 11.898.370 (`FuzzZerlegeAustausch`), 1.173.777 (`FuzzDienstruf`).<br>**Das letzte liegt eine Größenordnung tiefer, und das ist keine Nachlässigkeit**: es baut je Ausführung einen ganzen Dienst samt Doubles und Haushalt auf — ein geteilter gäbe nach ein paar hundert Rufen nur noch 429, und der Lauf prüfte dann die Drossel und nichts sonst.<br>**Zwei Funde in der Anwendung**, beide behoben und als Testfall unter `testdata/fuzz/` erhalten: ein Befund über ein **Feld ohne Namen**, der weder Ort noch Inhalt nannte (»Das Feld »« kennt dieses Programm nicht«), und ein **CSV-Rundlauf, der einen Wagenrücklauf verlor** — `encoding/csv` liest `\r\n` im Feld als `\n`, ein einzelnes `\r` bleibt stehen und wird beim nächsten Lesen zum Umbruch; dieselbe Notiz gab damit beim ersten und beim zweiten Rundlauf zwei Werte, obwohl `FAEHIGKEITEN.md` zusagt, dass sich ein CSV **vollständig** zurücklesen lässt.<br>**In `make ci` läuft nur der Saatkorpus** — `go test` fährt jedes `FuzzX` mit seinen Seeds als gewöhnlichen Lauf, und das ist die Regressionsprobe. Das eigentliche Fuzzen braucht Zeit und gehört auf diese Liste.<br>**Jeder Fund bleibt als Testfall**: `go test -fuzz` legt ihn unter `testdata/fuzz/` ab, und der Ordner gehört ins Repositorium.<br>**Erhoben am 2026-08-30**, je **Ziel** und mit **demselben Budget** — zwei Minuten, so wie die Aufrufe darüber es sagen. Ein Paket kann zwei Ziele haben, und sie laufen in verschiedenen Größenordnungen: 32.432.820 Ausführungen (`FuzzLiesInhalt`), 27.713.138 (`FuzzZerteile`), 23.578.162 (`FuzzZerteileSprache`), 13.474.978 (`FuzzLiesKarte`), 10.760.791 (`FuzzAuswerten`), 2.249.703 (`FuzzDokument`).<br>**Zwei Ziele fehlen in dieser Liste, und die Erklärung für sie war nur zur Hälfte richtig.** `FuzzBeschaedigtesPaket` und `FuzzAuspacken` lagen mit 10.673 und 2.787 drei Größenordnungen tiefer, und hier stand: sie legen je Ausführung ein Paket auf die Platte. Das kostet rund **3,5 ms** (1.300 Saatläufe von `FuzzAuspacken` in 4,5 s) und erklärt eine Größenordnung. **Den Rest kostete das Minimieren**: `go test -fuzz` verkleinert jede neue Eingabe bis zu 60 Sekunden, bevor sie zählt, und bei Millisekunden je Versuch hält das einen Arbeiter die ganze Laufzeit fest — der Koordinator wartet dann in `workerClient.minimize`. **Nachgemessen am 2026-09-23** (FMT-06, #1695): `FuzzAuspacken` fand mit der Vorgabe in einer Minute **keine einzige** neue Eingabe (42 Ausführungen), mit `-fuzzminimizetime 5s` in zwei Minuten **165** (2.527); `FuzzBeschaedigtesPaket` in einer Minute 8 (2.561), mit 5 s in zweien **192** (50.411). **Deshalb tragen die drei langsamen Ziele diesen Schalter.** Bei den schnelleren wirkt er schwächer, aber in dieselbe Richtung — je eine Minute, derselbe Tag: `FuzzModelle` (openai) fand mit der Vorgabe 212 neue Eingaben, mit 5 s 454; `FuzzDienstruf` 53 und 72. **Ihre Aufrufe bleiben, wie sie sind**, weil ihre Zahlen oben mit der Vorgabe erhoben sind; ob alle den Schalter tragen, entscheidet der nächste volle Lauf. **Und eine Minimierung, die bis zum Ende der Laufzeit nicht fertig ist, zählt gar nicht** — weder als Ausführung noch als neue Eingabe (Go 1.27.1, `internal/fuzz`: `coordinator` zählt eine Eingabe, die die Abdeckung erweitert, erst nach ihrer Minimierung, und `worker.minimize` gibt beim Abbruch keine Zahl zurück). Daher die kleinen Zahlen; was ein Lauf erkundet hat, sagt die Zahl der neuen Eingaben — eine Zahl reicht so weit wie ihre Erhebung (§10).<br>**Zahlen werden hier ersetzt und nicht fortgeschrieben.** Die von 2026-08-29 stammten aus verschieden langen Läufen und waren weder reproduzierbar noch untereinander vergleichbar; die erste Fassung von heute galt für Zusicherungen, die der Nachlauf über Phase 4.2 an sieben Zielen noch einmal verschärft hat. Eine Zahl gilt für den Lauf, der sie erzeugt hat — und für den Stand, den er gefahren ist.<br>**Vier Funde in der Anwendung**, alle behoben und als benannter Lauf erhalten: ein Befund ohne Ort in `mtgjson`, `NaN`/`Inf` im PDF-Inhaltsstrom, ein Befund über ein Set und ein Feld, deren Namen nur aus Leerraum bestehen, und eine Meldung von Scryfall, die bei derselben Antwort zweimal verschieden ausfiel.<br>**Dazu zwei Fehlgriffe in den Zusicherungen selbst** (`sort_text` und `0/0`), die derselbe Lauf angestrichen hat — sie zählen nicht als Fund an der Anwendung und stehen hier, weil sie sonst als einer gelesen würden |
| **Der Mutationslauf**                                            | **Ob die Tests einen echten Fehler sehen** (#1736, QA-05): zwölf Wachen aus [`RISIKOKATALOG.md`](RISIKOKATALOG.md) werden einzeln abgeschaltet, und jedes Mal muss ein benannter Lauf rot werden.<br>`node scripts/mutationslauf.mjs` (`make mutationslauf`)<br>**Am 2026-10-01: 12 von 12 gefangen** ([`belege/mutationen.md`](belege/mutationen.md)); der erste Lauf meldete 11, weil ein zu enger Filter den fangenden Test ausschloss |
| **Der Lastlauf**                                                | **Was Dienst und Werkzeugserver unter Last tun — und was der Schutz daraus macht** (#1739, QA-08): Lesen, Schreiben, Rendern, Einfuhr, lange Aufträge, MCP über stdio und über HTTP getrennt, gemischte Klienten, zwei Identitäten, Überlast mit Erholung und langsame Klienten — **mit und ohne** Kontingente, je Profil gegen einen frisch gestarteten Dienst über einem eigenen Bestand.<br>`node scripts/lastlauf.mjs` (`make lastlauf`; `DAUER`, `KLIENTEN`, `VARIANTEN`), Beleg in [`belege/lastlauf.md`](belege/lastlauf.md), die Auswertung in [`LASTMESSUNG.md`](LASTMESSUNG.md). **Ein Ruf zählt erst als gelungen, wenn seine Wirkung an einer Stichprobe da ist** (Bild am Ziel, Feld im Bestand, Set nach der Einfuhr); ein Lauf mit fehlender Wirkung endet mit Exitcode 1 |
| **Der Leistungsvergleich**                                       | **Dieselben Aufgaben mit denselben Eingaben in MSE2, MSE4Mac und 0.5.0** (#1735, QA-04): Start, Öffnen, Speichern, Export, Skript, eine Karte in einer laufenden Sitzung, Batchrendern, große Sets, Dauerbetrieb und die Last auf `mse-rest` — je Wiederholung in gedrehter Reihenfolge, unter `/usr/bin/time -l`, mit Kreuzprobe der gesicherten Dateien.<br>`node scripts/leistungsvergleich.mjs` (rund 16 Minuten; wartet, solange eine Pipeline läuft; **gegen heute ohne Guthaben, außer in der Lastgrenze**; ein Lauf mit Befund endet mit Exitcode 1)<br>**Am 2026-10-03**, nach #2005: 159 Fälle, 366 Schritte, null Befunde ([`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md) §9, Beleg [`belege/leistungsvergleich.json`](belege/leistungsvergleich.json)). Öffnen und Speichern um ein Mehrfaches schneller als MSE2, eine warme Karte über `mse-mcp` 105 ms (MSE2 43 ms, das Fenster 43 bis 61 ms, 0.5.0 1.195 ms), 70 Bilder in 2,4 s (MSE2 3,0 s) — **widerlegt**: gezeichnet waren 24 der 70 (Z1 des Zweitblicks über die QA-Arbeit), zu wiederholen im Freigabelauf. Der Lauf vom 2026-10-02 davor fand die Karte rund vierzigmal langsamer als MSE2 — die Ursache war #2005.<br>**Was er nicht misst**: das Fenster, das Bearbeiten und einen kalten Seitenspeicher |
| **Die Hüllenprobe**                                             | **Die Kernwege am nativen Fenster** — ENG-02.AK-04 (#1727), die Menüleiste aus #255, FIX-F02 bis FIX-F04.<br>`make huellenprobe` — baut das Abnahmebündel (`scripts/bauen.sh abnahme`) und fährt `node scripts/huellenprobe.mjs` über jedes Drehbuch in `scripts/huellenprobe/`; der Beleg je Drehbuch steht in [`belege/huellenprobe/`](belege/huellenprobe/).<br>**Über jedes Drehbuch und nicht nur über die des letzten Werkstücks**: zur Freigabe 0.6.0 waren 7 von 83 seit einem Umbau veraltet, ihr Beleg stand grün im Baum; ein rotes fand einen Fehler (#2068), und ein wackelndes einen zweiten (#2069) — **ein wackelndes Drehbuch wird gemessen, nicht wiederholt** — `ABNAHME-0.6.0.md` (Entwicklungsarchiv). Ein abgebrochenes Drehbuch hinterlässt sein Fensterbild `abbruch.png`.<br>**Zwei Helfer laufen mit und nicht allein**: `node scripts/baustand.mjs` gibt den Bauabdruck, den `bauen.sh abnahme` ins Bündel schreibt und gegen den die Probe ein veraltetes Bündel abweist; `node scripts/huellenorte.mjs` gibt den Abdruck der echten Orte, den `scripts/webkitabnahme.sh` vor und nach seinem Lauf vergleicht (T01, T03).<br>**Er läuft nicht in `make ci`** (§4): er startet die Anwendung über LaunchServices, je Drehbuch mit frischem Bestand und eigenem Schlüsselbund, unter einer eigenen Kennung.<br>**Was er nicht belegt**, steht in [`WAILS3.md`](WAILS3.md): die Hand vor AppKit, die Wahl im Dateidialog, `NSPasteboard` — und was nur ein Mensch sieht ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)). |
| **Der Nahtlauf durch die Wails-Brücke**                          | Der einzige Lauf, der die Naht zwischen Go und Fenster **betritt**. Go-Tests sehen die Struct-Marken nicht, Frontend-Tests gehen über `src/test/api.ts` und nie durch die JSON-Serialisierung (§3) — **zwei grüne Seiten einer Grenze sind kein Nachweis für die Grenze**.<br>`node scripts/nahtlauf.mjs`<br>Er startet die Anwendung im Browsermodus (`scripts/entwickeln.sh browser`, seit ENG-02) mit **eigenem Zuhause**, schickt einen Chrome ohne Fenster nach `34115` und ruft `window.go.app.Bindungen.*`. Antwort**formen** und keine Bindungen; wie viele, steht nur in `scripts/nahtvertrag.mjs`.<br>**Er läuft nicht in `make ci`** (§4), weil er die Anwendung baut; ein Fenster öffnet er seit Wails 3 nicht mehr.<br>**Was er nicht deckt**, steht in seinem Kopf: nicht alle Bindungen, nichts am Aussehen, und die dreizehn Fähigkeiten am nativen Dateidialog nicht — sie öffnen ein modales Fenster und kämen nie zurück.<br>**Und nach dem ersten Bruch trägt eine grüne Zeile weniger**: die Fälle tragen gemeinsamen Zustand, und wer nach einem gebrochenen zustandsbildenden Fall weiterliest, liest Zahlen ohne Isolation. Der Lauf sagt es an der Stelle (Nachlauf über Phase 4.3, Befund 6).<br>**Belegt am 2026-08-29**, dass er trägt: eine Antwort, die auf dem **Erfolgsweg** eine `nil`-Liste gibt, machte ihn rot — damals drei von elf Formen —, während `go test ./…` und `vitest` grün blieben (#1288). Die Zahl drei gehört zu jener Messung und nicht zum heutigen Stand |
| **Der Kanonlauf gegen MSE2**                                    | **Gefahren über `make fremdpakete`, und er hinterlässt einen Beleg** — `docs/belege/fremdpakete.json`; `make belege` sagt, ob er zum Stand der Freigabe passt (QA-01, #1733). Ein Beleg, der nicht PASS sagt oder nicht aktuell ist, gehört in die Release-Notiz.<br>**Wie gemessen wird, steht in [`MESSMETHODIK.md`](MESSMETHODIK.md)** — Maß, Stichprobe, geschichteter Schätzer, mit Gleichungen. Eingabe **und** Ausgabe stammen von MSE2 ([0026](entscheidungen/0026-mse2-rendert-die-referenz-selbst.md)): die kanonischen Sets sind seine eigenen, die Bilder daneben hat es selbst gerastert.<br>`MSE_FREMDPAKETE=/…/full-magic-pack/data MSE_SCHRIFTWURZEL=/…/full-magic-pack MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets MSE_MSE2AUSGABE=~/Library/Application\ Support/MagicSetEditor4Mac/mse2-ausgabe go test -tags fremdpakete ./internal/app/ -run TestDerAbstandZuMSE2AmKanon -v`<br>**`MSE_SCHRIFTWURZEL` zeigt auf die Wurzel des Packs und nicht auf sein `data`.** Die Schriften liegen **neben** den Paketen, unter `Magic - Fonts/`; `msepaket.Paketwurzeln` gibt die Verzeichnisse, die Pakete _enthalten_, und wirft sie damit heraus. Ohne die Angabe wird **jede Karte ohne Text gezeichnet**, und der Lauf misst das: 28,362 statt 18,853, in der Zerlegung 0 % `Text` statt 56 %. Seit dem 2026-08-28 wird er dann rot statt grün.<br>**Die drei Angaben sind Pflicht, und das ist teuer bezahlt.** `TestMain` gibt dem ganzen Lauf seit dem 2026-08-26 ein eigenes Zuhause unter `/tmp`, damit kein Test den Stand dieses Rechners sieht — und beide Wege bauten ihren Pfad aus `HOME`. Der Lauf **sprang ab**, und ein Sprung sieht aus wie ein grüner Lauf. Gefunden beim Vorbereiten von 0.3.0; `MSE_KANON` gab es damals schon, es war nur nie hierher gezogen worden<br>**Die Wurzel muss das _volle_ Pack sein.** `Archenemy` verlangt `magic-archenemy.mse-style`, und `M15-Magic-Pack` führt es nicht; ohne dieses Paket zeichnet dieses Programm die Schemakarten im Kartenmaß des Sets — 375×523 gegen 435×621 —, der Vergleich weist sie ab, und das **beste** Set fällt still aus. Genau das hat den Deckel einmal rot gemacht, ohne dass sich am Code etwas geändert hätte (#862).<br>`MSE_STICHPROBE=alle` fährt alles und dauert Stunden; ohne die Angabe zieht der Lauf so viele Karten, dass ein Fehler in einem Prozent der Karten mit 99,9 % Wahrscheinlichkeit auffällt.<br>**Er läuft nicht in `make ci`** (§4) — er gehört auf diese Liste, weil ihn sonst niemand fährt |
| **Der Abstand zu MSE2**                                         | Der einzige Vergleich dieses Projekts gegen einen **fremden** Maßstab; alle anderen Goldbilder prüfen uns gegen uns selbst ([0021](entscheidungen/0021-das-referenzbild-kommt-vom-nutzenden.md)).<br>`MSE_FREMDPAKETE=… MSE_REFERENZBILDER=… go test -tags fremdpakete ./internal/app/ -run TestDerAbstandZuMSE2 -v`<br>**Er ist rot, und er soll es sein** — er beziffert den Abstand, statt ihn wegzudefinieren. Was zählt, ist die Bewegung der Zahl gegenüber dem letzten Lauf; sie steht in [`ROADMAP.md`](ROADMAP.md).<br>`MSE_REFERENZBILDER` zeigt auf ein Verzeichnis mit je `<name>.jpg` und `<name>.karte`. **Es liegt nicht im Repository** (§11): die Bilder zeigen Rahmen und Symbole von Wizards of the Coast, die Kartendaten die schöpferische Leistung eines fremden Urhebers.<br>Mit `MSE_AUSGABE=…` legt der Lauf Differenzbild und eigene Ausgabe dort ab, statt in ein Temporärverzeichnis                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **Der Freigabebeleg**                                           | `release.sh` legt `freigabebeleg-<Fassung>.md` neben das Abbild. Darin stehen **Commit**, **Artefakthash** und die Prüfsummen, dazu die sieben Handprüfungen als **offene** Haken.<br>**Und die Sicherheitslage** (ENG-04): je Scanner einer von vier Zuständen — `ohne-befund`, `befund`, `nicht-erreichbar`, `gescheitert` — samt Kennungen oder Zahlen je Schwere. Ein Befund und ein gescheiterter Scanner beenden die Freigabe (`freigabeurteil` in `releaseschluss.sh`; dass es **wirkt**, hält `releaseschluss.test.sh`, indem er den Schluss mit einer gescheiterten Lage fährt, #1830), ebenso eine Lage, die nicht beide Scanner nennt — keine Antwort ist kein grüner Lauf. Ein Scanner, der nicht antwortete, ebenfalls — **es sei denn, jemand entscheidet es** mit `MSE_FREIGABE_OHNE_SCANNER=1`; dann steht die Entscheidung als eigene Zeile im Beleg, und der Scanner steht als nicht erreichbar darin und nicht als grün. Bis zum Zweitblick über Phase 6.2 hielt allein ein Befund an, und ein erneuter Scan, der scheiterte, lief bis „Fertig“ durch.<br>**Gebunden ist er an den Artefakthash, nicht an den Tag** — ein Tag lässt sich verschieben, ein Hash nicht. Wer neu baut, bekommt ein neues Abbild und braucht einen neuen Beleg, auch bei gleichem Quelltext: das `.dmg` trägt Zeitstempel. Das ist der Preis dieser Bindung und zugleich ihr Sinn — geprüft wird, was ausgeliefert wird.<br>Eine erledigte Prüfung wird eingetragen; ohne den passenden Hash weist der Befehl ab:<br>`scripts/releasebeleg.sh haken <beleg> <artefakthash> <1..7> "<wer>"`<br>**Was er nicht kann**: er glaubt, was jemand einträgt. Ein Haken sagt, dass eine Person zu einem Zeitpunkt behauptet hat, die Prüfung gemacht zu haben — mehr kann ein Beleg über eine Handprüfung nicht sagen, und weniger wäre gar nichts.<br>Der Beleg gehört mit ins Release, neben `pruefsummen.txt`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **Die Messwege**                                                | Im ganzen Projekt gab es null Benchmarks; was schnell genug ist, war damit eine Meinung (#891, Werkstück 6). Drei Wege sind gemessen — ein Set einfahren, die Liste laden, eine Karte rastern —, und jeder hat eine **benannte Schwelle**, an der der Lauf rot wird.<br>`MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets MSE_FREMDPAKETE=/…/full-magic-pack/data go test -tags fremdpakete ./internal/app/ -run XXX -bench Weg -benchtime 3x -v`<br>**Die Schwelle gilt für eine Maschine, nicht für eine Plattform** ([0007](entscheidungen/0007-kein-entfernter-lauf.md)): eine langsamere reißt sie, ohne dass am Code etwas falsch wäre. Sie ist ein Wächter gegen **Rückschritt**. Gemessen am 2026-08-25 auf einem Apple M5: einfahren 43 ms (107 Karten) und 1,78 s (3.788), Liste 10 ms, Rastern 71 ms.<br>**Und sie misst den Go-Weg, nicht das Warten am Fenster** — dazwischen liegen die Wails-Naht und der JSON-Zerteiler: dieselbe Liste braucht dort 529 ms. Wer wissen will, was jemand wartet, misst im Fenster, und das ist Handarbeit                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **Die Sammlungsseite im Fenster**                                | Seit Phase 4.5 (#1292). Er misst, was der Go-Weg nicht sieht: die Wails-Naht und das Zeichnen.<br>**Er läuft gegen ein eigenes Zuhause**, nicht gegen den Bestand der nutzenden Person — das Fenster nimmt `MSE_BESTAND` bewusst nicht entgegen, also wird `HOME` verlegt:<br>`node scripts/sammlungssaat.mjs` · `HOME=/tmp/mse45heim make dev` · `node scripts/sammlungsmass.mjs`<br>Gemessen am 2026-08-30 auf einem Apple M5 bei 1280 × 800 gegen 5.000 Exemplare: **Aufbau 321–358 ms, Suche 67–69 ms**, 200 Zeilen im DOM. Beide Schranken hält der Lauf selbst (600 ms und 120 ms) und wird rot, wenn sie reißen.<br>**Die Wails-Naht allein kostet 167 ms** für 1,5 MB JSON — der Go-Weg darunter 4,5 ms. Ohne die Zeilengrenze aus `sammlungssicht.ts` steht dieselbe Seite bei 680–791 ms und 145–150 ms; so ist die Gegenprobe gefahren worden |
| **Die Bildschirmabzüge der Dokumentation**                      | Seit #892; **seit 0.6.0 am nativen Fenster** — `scripts/bauen.sh abnahme`, `node scripts/huellenprobe.mjs readmebilder`, dann `python3 scripts/abzuegeschrumpfen.py docs/bilder`. Die Wege stehen in `scripts/abzuegeliste.mjs`; bis 0.5.0 lichtete `scripts/abzuege.mjs` einen Browser ab.<br>**Er stand bis zum 2026-08-30 in keiner Liste**, und der Wächter daneben zählte ihn nicht mit: er hielt vier Namen von Hand. Seit #1292 sucht er die Läufer, statt sie aufzuzählen — und hat diesen und `kuerzellauf.mjs` sofort angestrichen |
| **Die Leistungsschnitte**                                       | Seit PERF-01 (#2168) — der Messvertrag in benannten Schnitten, mit Umgebung und Pause je Schnitt; darunter `BenchmarkWegAusfuhr` mit Schwelle und die Kartenpaare aus PERF-04.<br>`MSE_FREMDPAKETE=/…/full-magic-pack/data MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets node scripts/leistungsschnitte.mjs`<br>**Er schreibt `docs/belege/leistungsschnitte.json`** und endet mit 1, wenn ein Schnitt nicht gemessen ist. Was die Zahlen sagen und was nicht, steht in [`LEISTUNGSSCHNITTE.md`](LEISTUNGSSCHNITTE.md); für jede Optimierung dazu `scripts/leistungssemantik.sh` (gleiche Ausgabe Byte für Byte) |
| **Die Sammlung bei fünftausend Exemplaren**                      | Seit Phase 4.5 (#1292) — `BenchmarkWegSammlungsliste`, `BenchmarkWegSammlungFuerEinenDruck`, `BenchmarkWegPortfolio`. Dasselbe Kommando, `-bench Weg` fährt sie mit; sie brauchen **kein** `MSE_FREMDPAKETE`, weil eine Sammlung an keinem Stilpaket hängt.<br>**Gemessen wird an 5.000 selbst gelegten Exemplaren**, nicht am Bestand dieses Rechners — der stand am 2026-08-30 auf null, und jeder Lauf über ihn kam in unter 10 ms zurück. Eine Zusage daraus wäre eine Aussage über eine leere Tabelle.<br>Gemessen am 2026-08-30 auf einem Apple M5 mit `-benchtime=10x`: **Liste 4,5 ms, ein Druck 17 µs, Portfolio 4,3 ms.**<br>**Der erste Lauf fand einen Befund**: das Portfolio brauchte **89 ms** über eine Preistabelle **ohne eine einzige Zeile** — es fragte einmal je Exemplar. Behoben mit `Preisspeicher.NeuesteAlle`; die Zusage wird als **Zahl der Abfragen** gehalten (`TestPortfolioFragtNichtJeExemplar`), weil die gilt, wo eine Millisekundenschwelle nur diese Maschine beschreibt |
| **Wo die Zeit einer Vorschau hingeht**                          | Seit #1270 ist der Vorschauweg **zerlegt** — an `Vanguard.mse-set` mit `magic-m15-altered`, also an dem Stil, an dem im Fenster 0,4–1,0 s gemessen wurden.<br>Dasselbe Kommando, `-bench Weg` fährt ihn mit. Gemessen am 2026-08-28 mit `-benchtime 10x`: **Kartenplan 27–28 ms, Zeichnen 13 ms, `png.Encode` 6,5 ms, Base64 0,07 ms, der ganze Go-Weg 53 ms.**<br>**Die Auskunft war, dass der Go-Weg die 100 ms hält** und das Fenster trotzdem 0,4–1,0 s brauchte: eine winzige Antwort (`Datenstand`) kostete dort 0,3 ms, eine 242-KB-Data-URI aber 550 ms.<br>**Behoben am 2026-08-28**: das Bild geht seither über die Assetleitung statt durch die JSON-Naht (`internal/vorschauweg`). Im Fenster nachgemessen: **72–87 ms**, davon 2–4 ms für das Bild; die Antwort daneben ist von 242.000 auf 560–824 Byte geschrumpft |
| **Der Kaltstart eines Stils**                                   | Seit #1270, `BenchmarkWegVorschaukaltstart` — dasselbe Kommando, `-bench Weg` fährt ihn mit, und er braucht `-benchtime 1x`.<br>**Er misst, was der warme Weg nicht sieht.** Gemessen am 2026-08-28 an `Vanguard.mse-set`: **Kartenplan kalt 620–683 ms, erstes Zeichnen 652–678 ms** — dieselben Schritte warm 29 ms und 13 ms.<br>**Dass es zwei Hälften sind, ist die Auskunft.** Die erste war zu erwarten (ein `.mse-style` wird ausgepackt, seine Skripte zerteilt); die zweite nicht — das erste **Zeichnen** kostet das Fünfzigfache jedes weiteren, weil dabei Schriften gesucht und Rahmenbilder entschlüsselt werden. Wer nur den Planbau beschleunigte, hätte die Hälfte gewonnen.<br>**Was sie nicht hält**: eine Schwelle über einen **einzelnen** Durchlauf mittelt nichts; sie steht bei 2,5 s und fängt einen Rückschritt, keine Streuung |
| **Planbau und Rastern, je Güte**                                | Seit Ω.5 (#1029) sind die beiden Hälften **getrennt** gemessen — `BenchmarkWegRastern` mischt sie, und wer eine Beschleunigung gegen diese Summe belegt, belegt sie gegen eine Zahl, in der die Hälfte gar nicht bewegt werden kann.<br>Dasselbe Kommando, `-bench Weg` fährt sie mit. Gemessen am 2026-08-25 mit `-benchtime 5x` über die **erste Karte** aus `Black.mse-set`: Planbau **41–47 ms**, `fast` **10 ms**, `medium` **18 ms**, `high` **20 ms**, `ultra` **33 ms**.<br>**Jede Schwelle liegt unter den Kosten der nächstlangsameren Güte** — das ist die Rückschrittsklasse, gegen die es sie gibt: eine Güte, die still die Abtastung ihrer Nachbarin nimmt, ist keine eigene mehr. Der erste Anlauf setzte `fast` auf 30 ms und hätte den Rückfall auf `medium` nicht gesehen; nachgestellt reißt sie bei 15 ms mit 20 ms.<br>**Was sie nicht halten**: `medium` fängt einen Rückfall auf `ultra`, nicht einen auf `high` — dafür läge die Schwelle 5 Prozent über dem Messwert und flackerte.<br>**Und es sind andere Zahlen als in Ω.4** (dort `fast` 17,5 ms): hier dieselbe Karte fünfmal, dort zwanzig verschiedene. Der Ebenenspeicher hält Rahmen und Illustration — beim Tippen ist die Wiederholung der richtige Fall, für einen Stapellauf die andere Zahl                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| **Der Weg der Flächen ohne Fenster**                            | Seit #2005, `BenchmarkWegKopflos` — dasselbe Kommando, `-bench Weg` fährt ihn mit. Er klont das Pack samt `Magic - Fonts` in ein eigenes Zuhause, wie der Leistungsvergleich, und misst an `Archenemy.mse-set` **eine Karte und einen Ausdruck über die Naht**, wie `mse-mcp` und `mse-rest` sie rufen — nach einem ersten Ruf —, dazu **die erste Karte über einem frischen Werk** und denselben Weg mit gehaltenem Werk; je Durchlauf Wand- und Rechenzeit.<br>**Drei Schwellen, und jede fängt einen bestimmten Rückfall**: 250 ms je Karte und 100 ms je Ausdruck fangen ein Werk, das wieder je Ruf entsteht (gemessen 65 und 16 ms, vorher 1,8 s und 0,5 s); 1,2 s für die erste Karte fangen eine zweite Wurzel über demselben Baum der Schriftsuche (gemessen 623 ms, 0.5.0 1,45 s). Zahlen und Bedingungen in [`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md) §8 |
| **Die Belege der Reife**                                        | Seit 0.8.0, Phase B (#2206) — jeder mit Frist im **Belegplan**, und `node scripts/belegplan.mjs` ist rot, wenn einer fehlt, überfällig ist, keine Herkunft nennt oder seine Schwelle reißt.<br>**Am Fenster**: `node scripts/huellenprobe.mjs vertrag zustaende budgets dauerbetrieb` — Vertragsrundgang über jede Bindung, 25 Seiten in sechs Zuständen, die Bedienbudgets, zwei Stunden Dauerbetrieb; danach `node scripts/bindungsreichweite.mjs <lauf> --beleg docs/belege/bindungsreichweite.md` über einen **vollen** Lauf.<br>**Die Antwortformen** liest `node scripts/antwortformen.mjs` aus `frontend/src/api.ts`; der Vertragsrundgang prüft gegen sie.<br>**Im Gast**: `node scripts/budgetsgast.mjs [<vm>]` — dieselben Budgets in einer Lume-VM; **was er nicht misst**, steht in seinem Kopf (hinter dem Sperrbildschirm zeichnet der Gast nicht). `node scripts/untergrenzegast.mjs [<vm>|--wirt]` fährt den Kernweg auf der Untergrenze (I02, Vermerk 0184), `node scripts/aufstieggast.mjs <0.7.0.dmg> [<vm>|--wirt]` Quarantäne und Aufstieg mit befülltem Bestand (I03), `node scripts/erststartgast.mjs <Abbild> [<vm>]` den Erststart aus dem Abbild des Kandidaten — Quarantäne, Abweisung, erster Start über einem frischen Bestand (HAND6, #2384, **nie am Wirt**; „Dennoch öffnen“ bleibt beim Menschen), `node scripts/bindunggast.mjs [<vm>]` die Bindung im Schlüsselbund mit eigenem Herausgeber (I01, Vermerk 0185, **nie am Wirt**). `node scripts/quicklookgast.mjs [<vm>]` misst, ob macOS eine ad hoc signierte Quick-Look-Erweiterung lädt und was sie aus der Sandbox sieht (G06, **nie am Wirt**). `node scripts/kennunggast.mjs <altes Bündel> <neues Bündel> [<vm>]` fährt den Umzug der Kennung mit beiden Schreibweisen der alten Domäne (R02, Vermerk 0191, **nie am Wirt**), `node scripts/m02gast.mjs <0.8.0> <Kandidat> [<vm>]` den Aufstieg mit einem Dienst unter launchd (M02, **nie am Wirt**), `node scripts/schriftordnergast.mjs [<vm>]` ob macOS Schriften in Unterordnern von `~/Library/Fonts` aktiviert (#2411, **nie am Wirt**), `node scripts/ops03gast.mjs [<vm>]` „Alles widerrufen“ unter launchd, wenn das Programm schon fehlt (OPS03, **nie am Wirt**, mit dem Abnahmebau), `node scripts/betriebsband09gast.mjs [<vm>]` ein Dienst unter launchd, der endet, im Betriebsband (OPS02, **nie am Wirt**, mit dem Abnahmebau), `node scripts/sec02gast.mjs [<vm>]` zehn Neustarts unter launchd mit je einer Marke in der Ansage (SEC02, Vermerk 0192, **nie am Wirt**, mit dem Abnahmebau) — die übrigen mit Freigabebauten aus `scripts/bauen.sh` und `scripts/werkzeuge.sh`.<br>**Die Zufallsmutation**: `node scripts/zufallsmutation.mjs` (Samen 2224, je Paket 40, in einem eigenen Arbeitsbaum), dann `node scripts/zufallsmutationbericht.mjs` mit der Durchsicht der Überlebenden.<br>**Die Zuordnung der Audit-Befunde** (QA01, #2382): `node scripts/befundzuordnung.mjs --schreiben` erzeugt `audits/0.9.0/ZUORDNUNG.md` (Entwicklungsarchiv) aus Audits, Planung und Umsetzungsvermerken; die Pipeline hält sie aktuell und lückenlos (`befundzuordnung.test.sh`).<br>**Was sie nicht sagen**: eine Maschine, ein Lauf je Beleg — die Grenzen stehen in jedem. |
| `make ci` grün                                                  | selbstverständlich, aber **nicht hinreichend** — es gibt keinen entfernten Lauf ([0007](entscheidungen/0007-kein-entfernter-lauf.md)), geprüft ist damit eine Maschine, nicht eine Plattform                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `CHANGELOG.md` datiert                                          | `## [Unveröffentlicht]` wird zu `## [x.y.z] — JJJJ-MM-TT`. Eine Fassung mit Nummer und ohne Datum ist ein vergessenes Datum                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| `docs/ROADMAP.md` und `docs/CHRONIK.md` fortgeschrieben         | Offenes und Erledigtes stehen getrennt                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| **Die gebaute Anwendung gestartet**                             | siehe unten                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| **Die Saat liegt im gebauten Bündel**                           | `ls build/bin/MagicSetEditor4Mac.app/Contents/Resources/saat` — die Wächter `werkzeuge.TestDasBauskriptLegtDieSaatInsBuendel` und `TestJederWegZumBuendelGehtUeberDasBauskript` halten den **Aufruf** in `scripts/bauen.sh`, nicht den Bau. Ob er die Saat wirklich hingelegt hat, sagt nur ein Bau, und ein Bau läuft nicht in der Pipeline ([0007](entscheidungen/0007-kein-entfernter-lauf.md))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Sichtprüfung hell und dunkel                                    | Farbkontraste sieht kein DOM-Test                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Rundlauf über echte MSE-Pakete                                  | lesen, schreiben, vergleichen — die Zusage aus `ARCHITEKTUR.md`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **Einen Druckbogen wirklich drucken und nachmessen**            | Die Tests prüfen die Zahlen im PDF, und Quartz liest es. Ob eine gedruckte Karte 63 mm breit ist, sagt nur ein Lineal: ein Drucker skaliert in der Vorgabe „an Seite anpassen", und dann stimmt jede Zahl im PDF und keine auf dem Papier. Zu drucken ist **ohne** Skalierung („Originalgröße", „100 %")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| **Die neue Fassung starten und den Bestand nachsehen**          | Sets, Karten und Einstellungen liegen unter `~/Library/Application Support/MagicSetEditor4Mac/`, **außerhalb** des Bündels — ein Wächter hält fest, dass kein Schreibort darin liegt (`internal/app/orteueberleben_test.go`). **Was er nicht zeigt, ist der Austausch selbst**: hier wird eine Fassung durch eine andere ersetzt, und danach muss dasselbe dastehen. Wer freigibt, schlägt ein Set auf, das vor dem Tausch da war (#1100, Werkstück 6)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| **Die Schritte 5 und 6 des Erststart-Wegs gehen und eintragen** | Was nach „Dennoch öffnen" erscheint, ist als einziges nicht abgelichtet — es braucht eine Hand an der Tastatur (siehe „Nachgestellt, nicht angenommen"). Wer freigibt, hat sie                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| **Eine Karte aus einem _bezogenen_ Paket ist zu sehen**         | Der Rundlauf oben prüft das **Format**, nicht das Bild: er liest 324 Pakete und schreibt sie zurück, und dabei bleibt eine Karte, die leer rendert, unauffällig. Der Kanonlauf misst inzwischen das Bild — 25.148 Karten, null Befunde der Art `Portierung` —, aber er misst gegen sich selbst. Zu prüfen bleibt mit den Augen: Set aus einem bezogenen Spiel öffnen, eine Karte anwählen, **hinsehen**. Der Ausgangsbefund #240 ist geschlossen                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **Ein Doppelklick im Finder auf eine `.mse-set`**               | Die Zuordnung ist am **gebauten** Bündel gemessen — `plutil -p …/Info.plist` zeigt `CFBundleDocumentTypes`, und LaunchServices führt nach `lsregister -f` die Bindung `.mse-set` samt `Contents/Resources/dokument.icns`. Was **kein** Lauf hier zeigen kann, ist der Doppelklick selbst: er geht über LaunchServices, und die weckt beim ersten Mal Gatekeeper — die Anwendung ist ad hoc signiert und nicht notarisiert (§12). Wer freigibt, klickt einmal: Symbol im Finder ansehen, doppelklicken, Set aufschlagen                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| **Die Menüleiste in beiden Sprachen**                           | Sie ist nativ, und `keintextimjsx.test.ts` liest JSX — kein Test sieht sie. Zu prüfen sind die Kürzel `⌘S`, `⌘,`, `⌘Z`, `⌘1`…`⌘4` und `⌘K` **und die Sprache**: seit #255 sind alle Menüs in der Sprache des Fensters, auch Anwendung, Bearbeiten und Fenster; „Über" und _Einstellungen_ stehen im Anwendungsmenü. **Am Fenster zu sehen ist**, ob _Alle einblenden_, _Alle nach vorne bringen_ und die _Dienste_ wirken und ob ⌘Z im Textfeld widerruft und außerhalb den fachlichen Verlauf trifft (Vermerk bei `Menuhandlungen.Rueckgaengig`). Unter Wails 2 war die Leiste zur Hälfte englisch (#822, [0030](entscheidungen/0030-die-menueleiste-bleibt-zur-haelfte-englisch.md)) |

### Die gebaute Anwendung starten

**Nicht die Entwicklungsfassung.** `make dev` lädt das Frontend aus Vite und
geht nicht durch dieselben Nähte wie der fertige Bau.

Der Grund ist eine Fehlerklasse, die zwischen den Prüfebenen liegt: Frontend-Tests
gehen über eine Vorrichtung und nie durch die JSON-Serialisierung, Go-Tests sehen
die Struct-Marken nicht. Ein `json:",omitempty"` auf einem Slice lässt eine
**leere** Liste aus der Antwort verschwinden — die Seite bleibt leer, und die
Pipeline ist grün.

Also: `.dmg` einhängen, App wirklich öffnen, jede Seite einmal ansehen.

---

## Der Bau

```bash
source scripts/activate.sh
./scripts/release.sh 0.1.0
```

Das Skript tut, in dieser Reihenfolge:

1. **Versionen prüfen**: `wails.json`, `CHANGELOG.md` und `frontend/package.json`
   nennen dieselbe Version — und den Tag `v<version>` gibt es **noch nicht**.
2. Arbeitsbaum sauber? Sonst Abbruch.
3. `make ci`.
4. `scripts/bauen.sh freigabe` — Oberfläche, Programm mit dem Bauschalter
   `production` und sonst keinem, Bündel samt Symbolen und Saat. Danach liest
   `release.sh` die Schalter am fertigen Programm (`go version -m`) und bricht
   ab, wenn einer der Zugänge `mcp` oder `server` darin steckt (ENG-02,
   [0132](entscheidungen/0132-wails-3-als-huelle.md)).
5. **Vorlagen und Beispielset ins Bündel** (`scripts/mitliefern.sh`) — 167 MB
   Pakete aus dem **bezogenen** Pack und ein kanonisches Set, damit die
   Anwendung nach dem ersten Start sofort eine Karte zeigt
   ([0037](entscheidungen/0037-die-vorlagen-gehen-ins-buendel.md)). **Ohne das
   bezogene Pack bricht die Freigabe ab**: ein Bündel ohne Vorlagen sähe aus
   wie eines mit und zeichnete nichts.
6. **Kommandozeile, MCP-Server und Dienst ins Bündel** (`scripts/werkzeuge.sh`)
   — `mse`, `mse-mcp` und `mse-rest` werden für `darwin/arm64` gebaut und nach
   `Contents/MacOS/` gelegt
   (#992).
   **Vor der Signatur**, damit `codesign --deep` sie mitnimmt: auf arm64
   startet ein unsigniertes Binärprogramm nicht.

   **Der Schritt prüft, dass sie laufen**, nicht nur, dass sie dastehen —
   `mse hilfe`, `mse-mcp` mit geschlossener Eingabe, `mse-rest` mit einem
   unmöglichen Port. Was er **nicht** prüft: dass sie bedienen. Das tun die
   Flächenläufe in `internal/flaechen`, gegen dieselben Programme.

   **Im `PATH` liegen sie nicht.** Diese Anwendung installiert nichts
   außerhalb ihres eigenen Bündels; ein `mse` in `/usr/local/bin` verlangte
   Rechte, die ein Doppelklick nicht hat, und bliebe liegen, wenn jemand die
   Anwendung in den Papierkorb zieht. Der Verweis ist eine Zeile, und sie steht
   in der Erststart-Notiz und im `README.md`.

6a. **Die zwei Läufe am Bündel** — reicht, was die Freigabe hineinlegt
   (#702)?

   ```bash
   MSE_BUENDEL=build/bin/MagicSetEditor4Mac.app go test -count=1 -tags fremdpakete \
     ./internal/app/ -run '^(TestNurAusDemBuendel|TestRundlaufUeberDasMitgelieferteSet)$' -v
   ```

   **Ein eigener Schritt, weil kein anderer Lauf sie fährt**: der Messlauf
   `make fremdpakete` führt sie als optional, denn ein Bündel entsteht erst
   hier. Hier stand bis QA-03
   (#1656),
   sie liefen »bei jedem Lauf mit gebautem Bündel« — gefahren hat sie kein
   Schritt, und der erste Beleg führte sie als übersprungen.

7. **Ad-hoc signieren, mit Hardened Runtime**: `codesign --force --deep --options runtime --sign - <App>`, ohne Entitlements ([0195](entscheidungen/0195-hardened-runtime-ohne-entitlements.md)); der Lauf prüft die Flagge `(adhoc,runtime)` am Siegel des Bündels **und an jedem ausführbaren Mach-O darin**, gleich unter welchem Ordner von `Contents` (Nachlauf über Phase 5, Befund 4).
8. Signatur nachsehen und `spctl` befragen — es **muss** ablehnen.
9. `.dmg` bauen, mit der Erststart-Notiz darin.
10. Beide Stücklisten erzeugen, **beide als CycloneDX in JSON**
    (`scripts/releasestueckliste.sh`). Ein Fehler dabei beendet die
    Freigabe, und eine Stückliste in einer anderen Form auch: bis 0.8.0 lag
    unter `stueckliste-go.json` XML, weil `cyclonedx-gomod` ohne `-json` XML
    schreibt (SEC04,
    #2357).
11. **Die Herkunft schreiben** (I06,
    #2264):
    `herkunft.json` nennt Commit und ob der Baum sauber war, die Fassungen der
    Werkzeuge, die gebaut haben — Go, Wails, Node, npm, cyclonedx-gomod,
    macOS, clang —, den SHA-256 des Abbilds und **jedes Programms im Bündel**
    und jeder Stückliste. Fehlt Bündel oder Abbild, endet die Freigabe
    (`scripts/releaseherkunft.sh`, gehalten von `releaseherkunft.test.sh`).
12. Prüfsummen schreiben — sie gehen über `*.dmg` und `*.json`, also auch
    über die Herkunft.

**Ob zwei Bauläufe dasselbe ergeben**, prüft `scripts/baugleich.sh zweimal
<raum>` — zwei Arbeitsbäume, je mit eigenem Bauzwischenspeicher, gebaut mit
`MSE_FREIGABE_NUR_BAU=1`, dann Bündel und eingehängtes Abbild Eintrag für
Eintrag. **Am Freigabestand von 0.8.0** (`e48f6d905`, 2026-10-12): 4.473 und
4.475 Einträge, kein Unterschied, das Programm bytegleich; die zwei
`.dmg`-Container unterscheiden sich in Kennung und Zeitstempel des
Datenträgers ([`belege/baugleich.md`](belege/baugleich.md)). Nicht gemessen
ist eine zweite Maschine.

**Das Bündel wiegt damit rund 437 MB und 301 MB als `.dmg`** — gemessen am
2026-10-07 (WART-02,
#2173) an
einem Bündel, das wie oben entsteht, nur ohne die Pipeline davor und ohne die
Signatur der Prüfsummen: Commit `687c72201` mit dem neuen Werkzeugschritt,
go1.27.1, darwin/arm64, macOS 27.0.1. **Drei Größen, und sie sind nicht
dieselbe**:

| | vorher | nachher |
| --- | --- | --- |
| logisch — die Summe der Dateigrößen | 519,8 MB | 437,1 MB |
| auf der Platte — APFS-Belegung (`du -sk`) | 530,6 MB | 447,9 MB |
| als `.dmg` (UDZO) | 370,7 MB | 301,4 MB |
| davon Programme | 315,9 MB | 233,2 MB |
| davon `Resources` — Vorlagen, Schriften, Beispielset, Saat | 202,4 MB | 202,4 MB |

**Der Unterschied liegt allein an den drei Werkzeugen.** Das Fenster baut
`bauen.sh freigabe` seit jeher mit `-trimpath -ldflags "-s -w"`; `mse`,
`mse-mcp` und `mse-rest` baute `werkzeuge.sh` ohne — je rund 80 MB statt
53 MB, und **jedes trug rund 3.300 Pfade des bauenden Rechners mit aus**,
samt seinem Kontonamen. Seit WART-02 dieselben Schalter; `werkzeuge.test.sh`
hält sie am Ergebnis fest. **Was gleich blieb**, am selben Commit gegeneinander
gefahren: die Ausgabe von `mse hilfe` Byte für Byte, die Exitcodes (unbekannter
Befehl 2, `mse-mcp` ohne Eingabe 0, `mse-rest` mit unmöglichem Port 1),
SIGTERM endet mit 0, und der Abzug nach SIGQUIT nennt dieselben Goroutinen mit
Zeilennummern — als `github.com/DavidTHeider/…` statt als Pfad hier.

**Ein Programm für alle drei** (ein Verteiler über `argv[0]`) ist **nicht
geplant**. Seine Obergrenze ist abgeleitet und nicht gemessen: was die drei
Werkzeuge teilen, also höchstens die Summe der zwei kleineren, rund 105 MB
logisch. Dafür änderte es, wie jedes der drei aufgerufen wird, startet und
Signale nimmt — vier Verträge, die je eine eigene Abnahme bräuchten.

Hier stand bis dahin „rund 506 MB auf der Platte und 367 MB als `.dmg`",
gemessen am 2026-10-05 am ausgelieferten Abbild von 0.6.0, und davor „rund
207 MB" vom 2026-08-21 — vor den Werkzeugen. **Am ausgelieferten Abbild von
0.7.0 nachgemessen** (2026-10-08, Commit `ad53f5e3c`): 438,2 MB logisch, 448,9
MB auf der Platte, 303,3 MB als `.dmg` — die Programme 234,3 MB und damit 1,1
MB mehr als bei der Messung oben, die Ressourcen gleich. Die Zahl ist an der
nächsten Freigabe wieder neu zu messen und nicht zu schätzen. Für einen Karteneditor
ist das üblich; es ist der Preis dafür, dass nach dem ersten Start sofort eine
Karte dasteht statt eines Downloads.

**Der Tag entsteht danach, von Hand** — und das ist Absicht, keine Lücke.

Hier stand bis zum Nachlauf des Zweitblicks über Phase 1, das Skript prüfe die
Version _gegen_ den Tag und setze ihn am Ende selbst. Beides stimmte nicht. Die
Beschreibung war älter als die Umsetzung, und niemand hat sie nachgelesen — sie
sah ja aus wie eine Beschreibung.

Die Reihenfolge folgt daraus, dass der Bau den Tag nicht braucht und der Tag den
Bau schon: getaggt wird der Stand, der **geprüft und gebaut** ist, nicht der, den
man freizugeben beabsichtigt.

Was das kostet, gehört dabei: `cyclonedx-gomod` liest die Modulversion aus Git,
und ohne passenden Tag steht in der Go-Stückliste eine **Pseudoversion**. Das
Skript sagt es an der Stelle. Die npm-Stückliste und das Bündel tragen dagegen
die richtige Nummer.

Danach von Hand:

**Nicht `build/freigabe/*`.** Darin liegt auch `dmg/` — das Verzeichnis, aus
dem `hdiutil` das Abbild gebaut hat —, und `gh release create` bricht an einem
Verzeichnis ab
(#878).

```bash
git tag -s v0.1.0 -m "Fassung 0.1.0" && git push origin v0.1.0
gh release create v0.1.0 \
  build/freigabe/*.dmg build/freigabe/*.json build/freigabe/*.tar.gz \
  build/freigabe/pruefsummen.txt build/freigabe/pruefsummen.txt.asc \
  --notes-file <notiz>
```

**Die `.asc` gehört dazu, und sie ist der Grund, warum die Liste hier steht.**
Ohne sie sind Abbild und Prüfsummen zwei Dateien aus dem Netz, und wer beide
austauscht, hat nichts gebrochen.

**Und das Quelltextpaket** (`*.tar.gz`, seit D01): ohne es gäbe die Freigabe
ein Programm unter GPL weiter und den vollständigen Quelltext nicht — der
Abschnitt „Bezug und Quelltext" sagt, was darin liegt.

### Die Signatur — was sie sagt und was nicht

`scripts/release.sh` signiert **`pruefsummen.txt`** mit dem Schlüssel aus
`git config user.signingkey` — demselben, der unter jedem Commit dieses
Repositoriums steht (§5). Die Prüfsummen decken das `.dmg`, beide
Stücklisten und `herkunft.json`; die Signatur deckt die Prüfsummen.

**Eine Signatur, nicht drei.** Wer sie prüft und danach `shasum -c` fährt, hat
die ganze Freigabe geprüft und muss dafür genau **einen** Schlüssel kennen.
Drei einzelne Signaturen wären nicht mehr wert: eine gültige Signatur über ein
`.dmg`, dessen Prüfsumme in einer unsignierten Liste steht, sagt nichts über
die Stücklisten daneben.

**Sie ist etwas anderes als `codesign`.** Beides beantwortet eine eigene Frage
([`entscheidungen/0003`](entscheidungen/0003-keine-notarisierung.md)):

|                     | Frage                                    | Antwort                                                                                                                               |
| ------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `codesign --sign -` | Ist das Bündel seit dem Bau unverändert? | ja — und **wer** es gebaut hat, sagt sie nicht. Eine Ad-hoc-Signatur trägt keine Identität, und genau deshalb weist Gatekeeper sie ab |
| `gpg --detach-sign` | Wer hat es gebaut?                       | der Freigabeschlüssel `44C59428DE1F766E`, den die Anwendung eingebettet trägt — seit 1.0.0 getrennt vom Commitschlüssel ([0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)) |

**Der Bauweg prüft seine eigene Signatur im selben Lauf.** Eine Signatur, die
niemand prüft, ist eine Datei (§10); scheitert die Prüfung, endet die Freigabe.
Der Fingerabdruck steht danach im Freigabebeleg — ein Beleg, der „signiert"
sagt und nicht **womit**, lässt die eine Frage offen, für die es ihn gibt.

**Was sie nicht leistet**: sie macht die Anwendung nicht notariell beglaubigt.
Gatekeeper sieht sie nicht an; der Erststart-Weg über _Dennoch öffnen_ bleibt,
wie er ist.

### Der Freigabeschlüssel wechselt über benannte Anker

**Vermerk [0194](entscheidungen/0194-der-freigabeschluessel-wechselt-ueber-benannte-anker.md)**
(SEC03, #2356).
Die Anwendung nimmt eine Freigabe nur an, wenn ein **benannter Anker** sie
signiert hat (`freigabeanker` in `internal/aktualisierung/signatur.go`), und
nur, wenn das Abbild den Namen der Fassung trägt, die der Tag behauptet
(`MagicSetEditor4Mac-<fassung>.dmg`) — so steht er in der signierten Liste.

**Ein Wechsel geht über zwei Freigaben**, und den neuen Schlüssel erzeugt und
verwahrt der Auftraggeber, mit vorab erzeugtem Widerrufszertifikat:

1. **Freigabe N** führt den alten und den neuen Anker im eingebetteten Bund
   (`freigabeschluessel.asc`) und in `freigabeanker`, und sie ist mit dem
   **alten** signiert. Jede installierte Fassung nimmt sie an.
2. **Freigabe N+1** ist mit dem **neuen** signiert (`Signaturschluessel`)
   und führt nur ihn — den alten nur noch mit seinem Widerruf. N nimmt sie
   an. Wer N übersprungen hat, braucht N oder ein Abbild von Hand; das
   steht in der Notiz von N+1.

**Zu 1.0.0 ist er gewechselt — ohne die zwei Freigaben.** Bis 0.9.0 signierte
der Commitschlüssel; seit 1.0.0 signiert ein eigener, `44C59428DE1F766E`, und
der Bund führt nur ihn. Den Übergang brauchte niemand: es gab keinen
öffentlich installierten Bestand, der den alten kannte
([0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). Wer eine
Entwicklungsfassung installiert hat, nimmt 1.0.0 nicht über die Aktualisierung
an, sondern als Abbild von Hand.

**Ein kompromittierter Schlüssel** kommt sofort mit Widerruf in den Bund der
nächsten Freigabe. Eine Fassung von vor dem Widerruf nimmt seine Signatur
weiter an — das steht in der Notiz an die Nutzenden, und nicht nur hier.

### Die Anwendung tauscht sich nicht selbst aus

**Stufe 2 der Aktualisierung lädt und bittet um die Installation** — sie
ersetzt das Bündel nicht ([0050](entscheidungen/0050-die-anwendung-ersetzt-sich-nicht-selbst.md)).
Der Grund steht dort und hängt an [0003](entscheidungen/0003-keine-notarisierung.md):
ein Selbsttausch braucht Schreibrecht auf das eigene Bündel und damit meistens
eine Rechteabfrage — und ein Programm ohne Notarisierung sollte nicht nach dem
Administratorkennwort fragen, um sich selbst zu überschreiben.

**Gemessen am 2026-08-26**: eine Datei, die diese Anwendung selbst über HTTP
lädt, trägt **kein** `com.apple.quarantine` — `xattr -l` zeigt nur
`com.apple.macl` und `com.apple.provenance`. Ein Go-Programm setzt das
Quarantäneattribut nicht; Safari und der Finder bekommen es über
LaunchServices. Was das **nicht** belegt, ist der Start aus dem echten `.dmg`
— dafür gibt es die Zeile in der Handprüfungstabelle oben.

**Geprüft ist eine Maschine, nicht eine Plattform.** Es gibt keinen entfernten
Lauf ([`entscheidungen/0007`](entscheidungen/0007-kein-entfernter-lauf.md)).

### Was in jede Release-Notiz gehört — im Wortlaut

Hier stand bis zum 2026-08-22 „die dortige Liste gehört in jede
Release-Notiz". Ein Verweis ist keine Liste: wer die Notiz schreibt, müsste ihn
aufschlagen, und wer sie liest, sieht ihn nie
(#874).

Der Block steht deshalb hier zum Übernehmen. **Die Zahlen werden vor jeder
Freigabe neu erhoben**; die eingesetzten sind die vom **2026-09-07**.

**Und er steht wortgleich im `README.md`**, unter „Was diese Fassung nicht
belegt" — auf Anweisung des Auftraggebers vom 2026-09-07. Ein Wächter hält
beide zusammen (`internal/werkzeuge/freigabeblock_test.go`): wer hier eine
Zahl ändert und dort nicht, macht die Pipeline rot. Zwei Fassungen desselben
Absatzes sind dieselbe Falle wie zwei Fassungen einer Leitlinie (§1).

> ## Was diese Fassung nicht belegt
>
> **Geprüft ist eine Maschine, nicht eine Plattform.** Diese Fassung ist auf
> einem Apple-Silicon-Mac gebaut und geprüft; einen entfernten Lauf gibt es
> nicht. Dauerhaft unsichtbar bleiben damit
>
> - alles Plattformabhängige jenseits der bloßen Übersetzung — für Windows und
>   Linux wird gebaut, aber nichts ausgeführt,
> - der Stand aus der Sperrdatei, weil lokal genommen wird, was daliegt, statt
>   `npm ci` zu fahren,
> - jede Eigenheit dieses Rechners, die in jedes Ergebnis eingeht,
> - ein frischer Klon auf einer zweiten Maschine.
>
> **Sie ist nicht notarisiert.** Es gibt keinen Apple-Developer-Account und es
> soll keinen geben. Beim ersten Start weist Gatekeeper sie ab; der Weg steht
> unten und im `.dmg`.
>
> **Zwei Rechner und fremde Konten liegen außerhalb des Rahmens.** Handel,
> Tisch und Paarung über eine Leitung sind mit zwei Anwendungen auf **einem**
> Mac belegt — nicht über zwei Geräte, zwei Netze oder einen Router
> dazwischen. **Die Wege der Leitung selbst** — mDNS, TLS und der Fernweg über
> tailcat, auch mit erzwungenem Relais — sind zusätzlich zwischen **zwei
> VM-Gästen** auf diesem Mac gelaufen, und die gebaute Anwendung hat dort unter
> der Erlaubnis *Lokales Netzwerk* gefunden und ist gefunden worden — im selben
> Netz, nicht hinter verschiedenen Routern. Die Dienste,
> die einen Zugang verlangen — MTGJSON über GraphQL, Scrydex, CardTrader,
> Anthropic, OpenAI, xAI und die Anmeldeanbieter —, sind gegen eingecheckte
> Beispiele geprüft; mit einem Konto gegen das Original gelaufen ist keiner.
>
> **Ein Geheimnis im Schlüsselbund liest jedes Programm desselben Nutzers.**
> Die Anwendung legt Zugänge, Anmeldegeheimnisse, den zweiten Faktor und den
> Schlüssel ihrer Identität über `security` in den Anmeldebund; ein solcher
> Eintrag vertraut diesem Werkzeug, und wer es startet, liest ohne Rückfrage.
> Eine Bindung an die eigenen Programme überlebte als Ad-hoc-Signatur keine
> Fassung und ist deshalb nicht gebaut — `docs/entscheidungen/0181`. Ein
> gesperrter Anmeldebund schützt auch diese Einträge.
>
> **Touch ID ist am echten Sensor zur Hälfte gelaufen.** Eine Frist und ein
> Abbruch mitten im Dialog enden ohne späten Erfolg; ob ein „Abbrechen" und ein
> Finger auf dem Sensor das Richtige tun, braucht einen Menschen und ist nicht
> gefahren.
>
> **Belegt ist je Klasse, nicht jeder Weg** (Reifeprofil zur Freigabe,
> `docs/REIFEPROFIL.md`). Jede Fähigkeit ist mit ihren
> Wegen über Fenster, Kommandozeile, MCP und Dienst benannt; gelaufen ist
> jede Klasse von Fähigkeiten, nicht jede Fähigkeit an jeder Fläche. Die
> achtzehn Arbeitswege sind in Teilaufgaben geprüft, nicht als Ganzes an
> dieser Fassung; als Ganzes liefen die vier Personawege.
>
> **Barrierefreiheit ist mit Tastatur und Zugänglichkeitsbaum belegt, nicht
> mit Ton.** Die gesprochene Ausgabe von VoiceOver und die Bedienhilfen des
> Systems (reduzierte Bewegung, erhöhter Kontrast) braucht einen Menschen; sie
> sind nicht abgenommen.
>
> **Ein Bestand von vor 0.7.0 öffnet sich nicht.** Das Schema bis 0.7.0 ist
> eine Grundlinie und keine Kette von Schritten mehr; einen älteren Bestand
> oder eine ältere Sicherung weist diese Fassung ab, ohne sie anzufassen, und
> sagt, was zu tun ist (Vermerk 0189) — den Weg über eine Fassung von 0.7.0
> bis 0.9.0, und die Entwicklungsfassungen sind nicht mehr veröffentlicht.
>
> **Im Dauerbetrieb hält eine Speicherregel nicht.** Zwei Stunden gemischter
> Arbeit laufen ohne Absturz und ohne Verlust. Gemessen am physischen
> Fußabdruck, je Probe zwischen zwei Runden, steigen Anwendung und Webansicht
> nach der ersten Stunde im Median je Abschnitt von 557 auf 601 MB (+12 % je
> Stunde; einzelne Proben bis 816 MB): der Webinhalt als Zwischenspeicher, der
> wieder freigibt — nach zwei Minuten Ruhe von 488 auf 311 MB, nach einer
> Speicherwarnung auf 242 MB —, die Anwendung um rund 10 MB je Stunde in den
> Bausteinen von macOS und der Go-Laufzeit (#2312).
>
> **Der letzte gezielte Angriffslauf ist vom 2026-10-16** — der Nachtest
> SEC01 gegen die eigene Netzadresse dieses Rechners, mit TLS und Marke
> jenseits der Schleife und dem Rollenwechsel im Betrieb. Nicht gelaufen sind
> zwei Geräte, fremde Netze, ein Angreifer mit eigenem Zeugnis im selben Netz
> und Last über die Drosselgrenzen hinaus (`docs/PENTEST.md`).
>
> **Der Erststart über „Dennoch öffnen" ist nachgestellt.** Im Gast mit macOS
> 15.2 weist Gatekeeper das Abbild ab, wie zugesagt; den Weg durch die
> Systemeinstellungen ist dort kein Mensch gegangen.
>
> **Die Untergrenze ist macOS 15.2, weil nur sie belegt ist.** Der Kernweg
> lief im Gast mit 15.2 und am Wirt mit 27; ein System mit 13 oder 14 gab es
> hier nicht (Vermerk 0184).
>
> **Das Bildmodell der lokalen KI läuft nicht.** Ollama 0.35.1 führt
> `x/flux2-klein:4b` nicht aus; Text und Einbettung sind gemessen.
>
> **Der Abstand zu Magic Set Editor 2 ist beziffert, nicht geschlossen.**
> Gemessen gegen MSE2s **eigene** Ausgabe der zehn kanonischen Sets mit
> **25.148** Karten, am **2026-10-21**, an der Rasterung dieser Fassung:
>
> | | feste Kohorte | Zufallsstichprobe |
> | --- | ---: | ---: |
> | Karten | **4.266** | **4.582** |
> | Kanonmittel | **9,319** | **9,293** |
> | 95-%-Band | — | [9,186 · 9,399] |
> | Median | 8,577 | 8,617 |
> | Q1 · Q3 | 7,460 · 9,593 | 7,451 · 9,610 |
> | P99 | **33,945** | 33,272 |
> | Minimum · Maximum | 1,691 · 49,040 | 1,691 · 72,914 |
>
> **Auf der Skala 0 bis 255**, und die Toleranz, bis zu der zwei Bilder hier
> gleich heißen, ist **1,0**. Ein Abstand von 9,3 ist also kein Rundungsfehler,
> sondern ein sichtbarer Unterschied — er sitzt zu **15 Prozent in der
> Schrift**, auf **4,4 Prozent der Fläche**. Zu 0.5.0 stand er bei 20,1, zu
> 0.7.0 bei 13,8, zu 0.8.0 bei 9,4.
>
> **Das Band gehört zur Zufallsstichprobe**: einmal zufällig gezogen, je Set
> so groß, dass sie einen Fehler auf einem Prozent seiner Karten mit 99,9
> Prozent trifft. Die feste Kohorte ist der Verlauf von Fassung zu Fassung und
> trägt kein Band. Wie gemessen wird — das Maß, die Stichproben und der
> Schätzer, jeweils mit Gleichungen —, steht in `docs/MESSMETHODIK.md`.
>
> **Die Zahl gilt für diese Menge** und sagt nichts über Vorlagen, die darin
> nicht vorkommen. Sie sagt auch nicht, dass zwei Karten **gleich aussehen** —
> sie sagt, wie weit sie in Bildpunkten auseinanderliegen. Was ein Mensch davon
> sieht, beantwortet der Blick.
>
> **Inhalte Dritter gehen mit.** Damit nach dem ersten Start sofort eine Karte
> dasteht, liegen im Bündel Vorlagenpakete und ein Beispielset — Rahmen,
> Manasymbole, Stilbögen und die Schriften der Vorlagen für M15-Karten. Für
> sie gilt die Lizenz dieses Programms **nicht**: sie gehören ihren
> Rechteinhabern, und die Schriften haben je eigene Lizenzen. Wem was gehört und
> unter welchen Bedingungen, steht in `DRITTANBIETER.md`.
>
> Alles Weitere bezieht die Anwendung auf dem Rechner der Nutzenden.

---

## Bezug und Quelltext

**Entschieden am 2026-10-13** ([0190](entscheidungen/0190-das-repositorium-wird-oeffentlich.md),
#2331):
**Kanal ist die Freigabeseite dieses Repositoriums**,
<https://github.com/DavidTHeider/MagicSetEditor4Mac/releases> — dieselbe,
die die eingebaute Aktualisierungsprüfung fragt
(`internal/aktualisierung/klient.go`). Herausgeber und Verantwortlicher ist
der Inhaber des Repositoriums; signiert wird mit dem Schlüssel aus dem
Abschnitt darunter.

| Je Fassung `v<fassung>` liegt dort | Was es ist |
| --- | --- |
| `MagicSetEditor4Mac-<fassung>.dmg` | das Abbild |
| `MagicSetEditor4Mac-<fassung>-quelltext.tar.gz` | **der Quelltext** nach GPL-2.0 §3: der Baum am Commit vollständig, `vendor/` mit den Go-Modulen, `frontend/node_modules/` mit den npm-Paketen des Bündels, `QUELLEN.md` mit Lizenzdatei und Ursprung jeder Abhängigkeit, der Tafel der erzeugt eingebetteten Formen und der Bearbeitungsformen, `quellen/` mit den Bauquellen (SQLite-Modul, Wails-Laufzeit) und unter `quellen/bearbeitungsformen/` das Repositorium jedes npm-Pakets am Commit seiner Veröffentlichung samt der SQLite-Quelle (#2393, rund 77 MB), `QUELLTEXT.md` (`scripts/releasequelltext.sh`). **Wer eine Abhängigkeit hebt, fährt `node scripts/bearbeitungsformen.mjs`** — sonst ist `bearbeitungsformen.test.sh` rot (Vermerk 0190, Nachtrag vom 2026-10-16) |
| `stueckliste-*.json`, `herkunft.json`, Manifest | Stücklisten, Herkunft, Verträglichkeit |
| `pruefsummen.txt`, `pruefsummen.txt.asc` | die Prüfsummen **aller** Beigaben, auch des Quelltexts, und ihre Signatur |

Adressierbar ist eine Fassung unter
<https://github.com/DavidTHeider/MagicSetEditor4Mac/releases/tag/v1.0.0> —
dieselbe Seite für Abbild und Quelltext; §3 der GPL lässt gleichwertigen
Zugang an derselben Stelle als Weitergabe des Quelltexts gelten.

**Seit dem 2026-11-15 ist das Repositorium öffentlich** — auf Auftrag des
Auftraggebers am selben Tag, nach der Freigabe 1.0.0. Befund A-08 ist damit
erledigt: `releases/latest` antwortet **ohne Anmeldung** mit `v1.0.0` und
seinen zehn Beigaben, die Signatur der heruntergeladenen Prüfsummen hält, und
die Anwendung selbst meldet aus einem Wegwerfzuhause „laufend 1.0.0,
freigegeben 1.0.0" (Vermerk 0190, Nachtrag vom 2026-11-15). **Vor dem Öffnen**
ist der Baum auf Inhalte geprüft worden, die nicht öffentlich gehören —
Geheimnisse, fremde Vorlagen (§11), persönliche Angaben in Belegen
([`belege/sensibledaten-1.0.0.md`](belege/sensibledaten-1.0.0.md)).

`internal/werkzeuge/quelltextangebot_test.go` hält diesen Abschnitt, den
Vermerk und `scripts/release.sh` gegeneinander.

## Signatur: ad hoc, nicht notarisiert

Es gibt **keinen Apple-Developer-Account** und es soll keinen geben. Die
Entscheidung und ihre Folgen stehen in
[`entscheidungen/0003-keine-notarisierung.md`](entscheidungen/0003-keine-notarisierung.md).

Was das praktisch heißt:

|                                           |                                                                                                                                                                                                                  |
| ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Ad-hoc-Signatur** (`codesign --sign -`) | **Pflicht.** Auf arm64 startet ein Binärprogramm ohne Signatur überhaupt nicht. Go und Wails signieren ad hoc schon beim Bau; der ausdrückliche Schritt stellt sicher, dass es auch nach dem Bündeln noch stimmt |
| **Developer-ID-Zertifikat**               | nicht vorhanden                                                                                                                                                                                                  |
| **Notarisierung**                         | findet nicht statt                                                                                                                                                                                               |
| **Folge**                                 | Gatekeeper weist die App beim ersten Start ab. Nutzende geben sie einmalig frei                                                                                                                                  |

Die Signatur lässt sich nachsehen:

```bash
codesign -dv --verbose=4 build/bin/MagicSetEditor4Mac.app
spctl --assess --type execute -vv build/bin/MagicSetEditor4Mac.app
```

`spctl` **wird** hier ablehnen. Der Ausgabetext ist auf macOS 26.6.1 genau eine
Zeile, und der Rückgabewert ist 3:

```
build/bin/MagicSetEditor4Mac.app: rejected
```

Hier stand bis #62, es komme `rejected (the code is valid but does not seem to
be an app)` oder `source=no usable signature`. Das war einmal so und ist es
nicht mehr; nachgemessen kommt weder das eine noch das andere — auch nicht mit
`-vvvv`. Wer eine Ausgabe zitiert, hat sie gesehen.

**Das `-vv` ist trotzdem nicht sinnlos**: bei einer Annahme nennt es die Quelle.
Es sagt nur nichts über die Ablehnung.

Warum abgelehnt wird, sagt ein anderes Werkzeug — und zwar vollständig:

```bash
syspolicy_check distribution build/bin/MagicSetEditor4Mac.app
```

Es nennt beides ausdrücklich: _Adhoc Signed App_ als **Warning** und _Notary
Ticket Missing_ als **Fatal**. Genau das ist die Entscheidung aus
[`0003`](entscheidungen/0003-keine-notarisierung.md), von Apples eigenem
Werkzeug ausgesprochen. Wer den erwarteten Zustand belegen will, belegt ihn
damit und nicht mit einem `rejected` ohne Begründung.

Der erwartete Zustand ist kein Fehler im Bau. Wer ihn wegdiskutiert, hat die
Entscheidung rückgängig gemacht, ohne sie zu ändern.

**Die Quarantänemarke ändert an `spctl` nichts.** Sie entscheidet nicht über das
Urteil, sondern darüber, ob das System beim Start überhaupt fragt: `spctl` lehnt
das frisch gebaute `.app` genauso ab wie die heruntergeladene Fassung.

---

## Was Nutzende beim ersten Start tun

Dieser Weg gehört in jede Release-Notiz, in `README.md` und als Textdatei ins
`.dmg`. Er ist kein Mangel, den man verschweigt, sondern eine Eigenschaft, die
man erklärt.

> **Der Zettel in der `.dmg` von 0.1.0 trägt zwei falsche Angaben.** Er sagt
> „Lizenz: MIT" — seit dem 2026-08-16 gilt GPL-2.0-or-later
> ([0019](entscheidungen/0019-gpl-2-0-statt-mit.md)) — und er behauptet, Inhalte
> Dritter würden nicht mitgeliefert, während derselbe Lauf 167 MB davon einlegt.
> Berichtigt am 2026-08-23 mit
> #894; drei
> Wächter halten den Zettel seither.
>
> **Für die bereits veröffentlichte `.dmg` ändert das nichts.** Wer sie
> weitergibt, gibt den falschen Text mit. `LICENSE`, `README.md` und die Angabe
> **in** der Anwendung waren immer richtig — wer die Lizenz wissen will, findet
> sie dort. Die nächste Freigabe trägt den berichtigten Zettel; eine
> nachträgliche Berichtigung der veröffentlichten Datei ist **nicht** möglich,
> ohne ihre Prüfsumme zu brechen.

1. `MagicSetEditor4Mac.app` in **Programme** ziehen.
2. App starten. Es erscheint das Warnfenster:

   ![Das Warnfenster: „MagicSetEditor4Mac" nicht geöffnet](bilder/gatekeeper-blockiert.png)

   Titel: **„MagicSetEditor4Mac" nicht geöffnet**. Darunter: „Apple konnte
   nicht überprüfen, ob „MagicSetEditor4Mac" frei von Schadsoftware ist, die
   deinen Mac beschädigen oder deine Privatsphäre verletzen kann."

   Zwei Knöpfe, und der **obere** ist der gefährliche: **„In den Papierkorb
   legen"**. Der richtige ist der untere, **„Fertig"**.

3. **Systemeinstellungen** öffnen → **Datenschutz & Sicherheit**. Direkt dorthin
   führt `open "x-apple.systempreferences:com.apple.settings.PrivacySecurity.extension?Security"`.
4. Nach unten zum Abschnitt _Sicherheit_ scrollen. Dort steht:

   ![Der Abschnitt Sicherheit mit dem Knopf „Dennoch öffnen"](bilder/gatekeeper-dennoch-oeffnen.png)

   „„MagicSetEditor4Mac" wurde blockiert, um deinen Mac zu schützen." — und
   daneben der Knopf **„Dennoch öffnen"**.

5. **„Dennoch öffnen"** anklicken, mit Kennwort oder Touch ID bestätigen.
6. App erneut starten, Rückfrage mit **„Öffnen"** bestätigen.

Danach startet sie ohne weitere Nachfrage.

**Ab macOS 15 ist das der einzige Weg.** Das frühere Öffnen über Rechtsklick →
_Öffnen_ funktioniert nicht mehr; wer es in einer Anleitung findet, liest eine
veraltete.

### Von 0.8.0 oder älter: eine neue Kennung

**Betrifft nur, wer eine Entwicklungsfassung bis 0.8.0 installiert hat.**
**Seit 0.9.0 heißt das Bündel `io.github.davidtheider.magicseteditor4mac`**
und nicht mehr `com.wails.MagicSetEditor4Mac`
([0191](entscheidungen/0191-die-kennung-ist-io-github-davidtheider.md),
#2334).
Beim ersten Start zieht die Anwendung selbst um, was unter der alten Kennung
lag: die Vorgaben, den Speicher des Fensters und einen eingerichteten Dienst,
der danach unter dem neuen Etikett läuft — genau einer. **Erlaubnisse gehen
nicht mit**: macOS bindet *Lokales Netzwerk* und *Kontakte* an die Kennung und
fragt neu, wenn die Anwendung sie wieder braucht. Das gehört in die
Release-Notiz von 0.9.0.

**Zurück zu 0.8.0** geht, wie jeder Rückweg einer Fassung („Die Appfassung
tauschen", `ADMINISTRATION.md`): das alte Bündel findet seine Vorgaben, denn
der Umzug kopiert und lässt das Alte liegen. **Seinen Dienst findet es nicht**
— der läuft unter dem neuen Etikett weiter, mit demselben Programm. Wer ihn am
alten Fenster führen will, **widerruft ihn zuerst am Fenster der neuen
Fassung** und richtet ihn dann am alten ein; sonst lauschen zwei Dienste auf
demselben Port.

### Die Kette nach der Freigabe: was ein Neuling wirklich vor sich hat

**Der wunde Punkt war Schritt 4** (#676):
ohne bezogene Vorlagen zeichnet keine Karte, und ohne Set gibt es nichts zu
öffnen. Seit [0037](entscheidungen/0037-die-vorlagen-gehen-ins-buendel.md)
bringt das Bündel beides mit; die Kette ist damit kürzer:

| Schritt                                      | Was zu sehen sein muss                                                       |
| -------------------------------------------- | ---------------------------------------------------------------------------- |
| 1. `.dmg` laden, App nach `Programme` ziehen | —                                                                            |
| 2. Erster Start: Gatekeeper weist ab         | das Warnfenster oben, dann _Datenschutz & Sicherheit_ → **„Dennoch öffnen"** |
| 3. Die Anwendung startet                     | Einstiegsseite mit **drei** Kacheln                                          |
| 4. **Sets** aufschlagen                      | Abschnitt **„Zum Ausprobieren mitgeliefert"** mit dem Knopf _Vanguard_       |
| 5. _Vanguard_ anklicken                      | „107 Karten geöffnet", das Set steht in der Liste                            |
| 6. Auf das Set klicken, Lasche **Karten**    | die erste Karte **gezeichnet** — Rahmen, Kunst, Typzeile, Regeltext          |
| 7. Ein Feld ändern                           | die Vorschau folgt, ohne dass jemand etwas bestätigt                         |
| 8. **Sichern**                               | die Datei liegt neu da; was niemand angefasst hat, steht unverändert darin   |

**Kein Paketbezug dazwischen.** Er bleibt der Weg zu allem, was **nicht**
mitkommt — **316 der 324** Pakete von `M15-Magic-Pack`, nachgezählt am
2026-08-28 —, aber die Kette oben braucht ihn nicht.

**Was davon maschinell geprüft ist**, und zwar in Schritt 6a der Freigabe: `TestNurAusDemBuendel` startet die Anwendung mit einem **leeren**
Paketverzeichnis und allein den mitgelieferten Vorlagen — Schritt 4 bis 6, mit
107 Karten und null Befunden. `TestRundlaufUeberDasMitgelieferteSet` geht
Schritt 7 und 8 über dieselben 107 Karten.

**Was ein Mensch tun muss**: Schritt 1 bis 3. Sie hängen an Gatekeeper und an
einem Fenster, und beides prüft kein Test — der Weg dafür steht unten.

**Und was auf dieser Maschine grundsätzlich nicht zu prüfen ist**: ein
**fremder** Mac. [0007](entscheidungen/0007-kein-entfernter-lauf.md) führt das
als dauerhaft unsichtbar, und es bleibt es. Was hier steht, gilt für einen Mac
mit demselben Betriebssystem und ohne die Daten dieses Projekts — nicht für
einen, den niemand gesehen hat.

### Nachgestellt, nicht angenommen

Die eigene Maschine hat die App schon oft gestartet; was ein **fremder**
Benutzer sieht, sieht man dort nie. Der Weg oben ist deshalb am **2026-08-16 auf
macOS 26.6.1 (Build 25G76)** nachgestellt worden. So:

```bash
# 1. Das .dmg wie ein Download markieren, einhängen, die App herauskopieren.
xattr -w com.apple.quarantine "0001;$(printf %x $(date +%s));Safari;$(uuidgen)" MagicSetEditor4Mac-0.1.0.dmg
hdiutil attach MagicSetEditor4Mac-0.1.0.dmg -nobrowse
cp -R "/Volumes/MagicSetEditor4Mac 0.1.0/MagicSetEditor4Mac.app" ~/Schreibtisch/

# 2. Die Marke am .app selbst setzen — siehe die zwei Fallen darunter.
xattr -w com.apple.quarantine "0001;$(printf %x $(date +%s));Safari;$(uuidgen)" ~/Schreibtisch/MagicSetEditor4Mac.app

# 3. Über den **Finder** starten, nicht über `open`.
osascript -e 'tell application "Finder" to open (POSIX file "…/MagicSetEditor4Mac.app" as alias)'
```

**Zwei Fallen, beide beim ersten Anlauf hineingetreten:**

| Falle                                  | Was dann passiert                                                                                                                                                              |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `open …` aus dem Terminal              | Die App startet **ohne jede Rückfrage**. Ein Terminal gilt als Entwicklerwerkzeug, und was es startet, prüft Gatekeeper nicht. Wer den Weg so nachstellt, hält ihn für unnötig |
| `hdiutil attach` auf der Kommandozeile | Setzt am Abbild die Freigabemarke (`0x0080`), und die **erbt** die herauskopierte App: `0281;…`. Ein Doppelklick im Finder tut das nicht. Deshalb Schritt 2                    |

**Was dabei sonst noch herauskam:** die Marke allein entscheidet nicht. Setzt
man von Hand `0081` — den Wert, den „Dennoch öffnen" hinterlässt —, erscheint
das Warnfenster **trotzdem**. Die Freigabe liegt in der Systemdatenbank von
`syspolicyd`, nicht im erweiterten Attribut. Wer sie durch ein `xattr -w`
nachzubilden versucht, prüft nichts.

**Ungeprüft sind die Schritte 5 und 6.** Was nach dem Klick auf „Dennoch
öffnen" erscheint, steht hier aus Apples Dokumentation und nicht von einem
Bildschirm: der Klick und die Authentisierung brauchen eine Hand an der
Tastatur, und der Lauf, der die Schritte 1 bis 4 abgelichtet hat, hatte keine.
**Wer die Freigabe fährt, geht die Schritte und trägt das Ergebnis hier ein** —
sie stehen in der Checkliste oben.

### Warum kein zweites Benutzerkonto

#62 verlangt ein
**frisches Konto**, und der Grund dafür ist gut: die eigene Maschine hat die App
schon oft gestartet, und eine einmal erteilte Freigabe sieht man danach nie
wieder. Der Zweitblick über Phase 8 hat angestrichen, dass der Lauf oben auf dem
eigenen Konto stattfand.

Er hat trotzdem geleistet, was das Issue meint, und der Beleg steht im Ergebnis:
**das Warnfenster ist erschienen.** Wäre die Freigabe für diese Signatur schon
erteilt gewesen, wäre die App gestartet — genau das ist auf demselben Konto
zweimal zu sehen gewesen, einmal über `open` aus dem Terminal und einmal, als
die Marke `0281` vom Abbild geerbt war. Der Zustand, den ein frisches Konto
herstellen soll, lag also vor.

**Was ein zweites Konto zusätzlich zeigte**, und was hier deshalb offen bleibt:
ob der Weg über _Datenschutz & Sicherheit_ auch einem Benutzer **ohne
Administratorrechte** offensteht. Der Knopf „Dennoch öffnen" verlangt eine
Authentisierung; ob dafür ein eigenes Kennwort genügt oder das eines
Administrators nötig ist, ist hier nicht nachgemessen. Für eine Anwendung, die
Einzelne herunterladen, ist das die weniger dringende Frage — aufgeschrieben
gehört sie trotzdem.

Ein Wächter (`internal/werkzeuge/erststartweg_test.go`) hält die drei Orte
gegeneinander, an denen dieser Weg steht: hier, in `README.md` und als Zettel im
`.dmg`. Was er **nicht** kann, ist nachsehen, ob Apple die Sätze inzwischen
geändert hat — dagegen hilft nur, den Weg wieder nachzustellen.

Für Fortgeschrittene bleibt die Quarantänemarke selbst:

```bash
xattr -d com.apple.quarantine /Applications/MagicSetEditor4Mac.app
```

Das steht in der Dokumentation als **Erwähnung**, nicht als Empfehlung. Ein
Kommando, das eine Sicherheitsprüfung abschaltet, gehört nicht in eine
Schritt-für-Schritt-Anleitung für alle.

---

## Sicherung und Rückspielung des Bestands

Der Bestand liegt unter
`~/Library/Application Support/MagicSetEditor4Mac/bestand.db`.

**Nicht kopieren.** Die Datenbank läuft im WAL-Modus; neben `bestand.db` liegen
`bestand.db-wal` und `bestand.db-shm`, und alles seit dem letzten Checkpoint steht im
`-wal`. Wer nur die erste Datei wegträgt, bekommt einen Bestand, der sich
anstandslos öffnen lässt und stillschweigend älter ist, als er aussieht — bei
frisch geschriebenen Daten ist er **leer**, samt Tabellen. Nachgemessen in
`internal/storage/sicherung_test.go`.

### Sichern

**In der Anwendung: _Ablage → Bestand sichern …_** Der native Speichern-Dialog
fragt nach einem Ort, und die Meldung oben nennt danach den Pfad. Geschrieben
wird über `VACUUM INTO`: **eine** Datei, auf der offenen Datenbank, während
gearbeitet wird, und ohne `-wal` und `-shm` daneben.

Von Hand, bei geschlossener Anwendung, geht dasselbe über die SQLite-Konsole:

```bash
sqlite3 ~/Library/Application\ Support/MagicSetEditor4Mac/bestand.db "VACUUM INTO '$HOME/Desktop/mse-sicherung.db'"
```

Das Ziel darf nicht dastehen — weder `Sichern` noch SQLite überschreibt etwas.

**Ohne Kennwort ist die Sicherung Klartext**, gleich wie der Rechner
eingestellt ist. FileVault schützt das Laufwerk _dieses_ Rechners; eine Kopie
auf einem USB-Stick oder in einer Wolke ist davon nicht berührt. **Mit
Kennwort** (seit OPS-06, `MSE_SICHERUNGSKENNWORT` an der Kommandozeile) ist sie
mit age verschlüsselt — was gilt, steht in [`SICHERUNG.md`](SICHERUNG.md).

**Der Bestand daneben kann verschlüsselt sein, das Archiv wird es dadurch
nicht** ([0072](entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md),
Werkstück 14 von Phase 4.12). Hier stand bis zum 2026-09-07 „der Bestand selbst
wird nicht verschlüsselt, und warum, steht in
[0042](entscheidungen/0042-keine-verschluesselung-des-bestands.md)" — der
Vermerk ist seit dem 2026-09-06 abgelöst, und es war die **dritte** Stelle mit
demselben überholten Satz.

Wer die Sicherung aus dem Haus gibt, gibt ihr ein Kennwort.

### Zurückspielen

**In der Anwendung: _Einstellungen_ → _Bestand_ → „Sicherung zurückspielen …"**
(#940). Der
Weg hat drei Schritte, und die Reihenfolge ist die Zusage:

1. Der native Dateidialog fragt nach der Sicherung.
2. Sie wird **gelesen, nicht übernommen** — nur lesend geöffnet, damit die
   Prüfung ihren Gegenstand nicht verändert. Was dabei herauskommt, steht
   danach auf der Seite: wann der Bestand angelegt wurde, wie viele Sets und
   Karten darin liegen. Erst dann kommt die Frage.
3. Ein Klick auf „Zurückspielen" **prüft noch einmal** — zwischen Auswahl und
   Bestätigung liegt die Entscheidung eines Menschen, und in dieser Zeit kann
   die Datei eine andere geworden sein —, schließt dann die Datenbank, tauscht
   die Datei samt `-wal` und `-shm` und öffnet wieder.

**Und vorher wird gefragt, ob jemand sonst die Datei hält**
(#961). Läuft
`mse-mcp` oder ein zweites Fenster auf demselben Bestand, wird **nicht**
getauscht: eine Datei umzubenennen, die ein anderer Prozess geöffnet hält,
nennt SQLite undefiniert und korruptionsgefährlich. Die Meldung sagt, dass ein
anderes Programm sie geöffnet hat.

**Drei Abweisungen, und sie sind nicht dieselbe:** eine Datei, die keine
lesbare Datenbank ist; eine, die kein Bestand dieser Anwendung ist; und eine
aus einer **neueren** Fassung — die wird **benannt** abgewiesen, mit dem
Schemastand, den diese Fassung nicht kennt, und dem, bis zu dem sie reicht.
Alle drei geschehen, **bevor** getauscht wird.

**Der ersetzte Bestand ist nicht fort.** Er bleibt als
`bestand-vor-rueckspielung.db` neben dem neuen liegen — bis zum nächsten
Zurückspielen, das ihn überschreibt. Wer zwei Stände aufheben will, sichert
vorher selbst.

**Und lässt sich der zurückgespielte Stand nicht öffnen, geht der Weg zurück:**
der alte kommt an seinen Platz, die Anwendung läuft mit ihm weiter, und die
Meldung sagt, woran es lag. Ein Zurückspielen, das scheitert, endet nicht
schlimmer als eines, das gar nicht stattfand.

**Was der Weg im Fenster nicht kann:** eine Sicherung einspielen, wenn die
Anwendung gar nicht startet. Dafür bleibt die Handarbeit unten stehen — sie ist
kein Überbleibsel, sondern der Rückfall.

---

#### Von Hand, bei geschlossener Anwendung

**Am 2026-08-22 einmal gegangen**, nicht nur gelesen — auf einem Abzug des
echten Bestands, damit ein Fehler im Weg nicht den echten trifft. Der Lauf hat
sofort etwas gefunden: **die Datei heißt `bestand.db` und nicht `mse.db`.** Der
ganze Abschnitt nannte sechsmal einen Namen, den es nicht gibt; wer ihn befolgt
hätte, hätte `sqlite3` mit einer leeren Datei gefüttert und eine Sicherung
weggetragen, die nichts enthält
(#872).

Gemessen am Abzug: 3 Sets vorher, `VACUUM INTO` gibt 716.800 Byte, 3 Sets nach
dem Zurückspielen, `PRAGMA integrity_check` sagt `ok`, und die Migrationstabelle
steht auf `0012_paketkurzname.sql`.

Ein Handgriff, und die Reihenfolge zählt:

1. **MagicSetEditor4Mac beenden** (`⌘Q`). Bei laufender Anwendung ist der
   Bestand offen, und eine Datei unter einer offenen Datenbank auszutauschen
   hinterlässt einen Stand, den niemand vorhergesagt hat.
2. Die drei Dateien am alten Ort **wegräumen**, nicht überschreiben:

   ```bash
   cd ~/Library/Application\ Support/MagicSetEditor4Mac && mkdir -p alt && mv bestand.db bestand.db-wal bestand.db-shm alt/ 2>/dev/null
   ```

   `-wal` und `-shm` müssen mit. Bleiben sie liegen, legt SQLite den alten
   Vorschreib über die zurückgespielte Datei — und das Ergebnis ist keins von
   beiden.

3. Die Sicherung an ihren Platz legen:

   ```bash
   cp ~/Desktop/mse-sicherung.db ~/Library/Application\ Support/MagicSetEditor4Mac/bestand.db
   ```

4. Anwendung starten. Sie migriert den zurückgespielten Bestand auf den Stand
   dieser Fassung und prüft ihn (`PruefeBestand`). Stammt die Sicherung aus
   einer **neueren** Fassung, sagt sie das und öffnet ihn nicht — siehe
   #795.
5. Erst wenn die Sets dastehen: `alt/` löschen.

## Nach der Freigabe

- Release auf GitHub mit `.dmg`, Prüfsummen und Stückliste.
- Release-Notiz aus `CHANGELOG.md`, ergänzt um den Erststart-Weg oben.
- `docs/CHRONIK.md` fortschreiben.
- Nächste Phase eröffnen: Issues nachschärfen (`AGENTS.md` §13).
