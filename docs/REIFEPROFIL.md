# Reifeprofil — erhoben zur Freigabe 0.9.0, gültig für 1.0.0

**1.0.0 trägt den Programmstand von 0.9.0**: dieselbe Fachlogik, dieselbe
Oberfläche, dieselben Grenzen, neu gebaut mit der Fassung 1.0.0, mit zwei
Sicherheitsaktualisierungen (Go 1.27.4, `klauspost/compress` v1.20.2) und einem
eigenen Freigabeschlüssel (Vermerk
[0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). **Das Profil darunter
ist deshalb das der Freigabe 0.9.0**, unverändert übernommen; es gilt, bis es
neu erhoben wird. Die Kriterien, die Zuordnung der Befunde und die Zweitblicke,
auf die es sich beruft, liegen mit den Audits im Entwicklungsarchiv.

**Was das heißt**: das Tor hält auch 1.0.0 nicht ganz — D1, D4 und D6 an der
Sache, D10, D12 und D13 an Mensch oder Hardware. 1.0.0 ist die erste Freigabe
für Anwendende **mit dieser benannten Grenze**, nicht ihre Aufhebung.

---

## Das Profil zur Freigabe 0.9.0

**Stand 2026-10-20**, Zweig `qa/2382-reifeprofil` vom `main` nach dem letzten Merge der Iteration (`2b2204ffa`, #2525), QA01 #2382. Maßstab: `KRITERIEN.md` (Entwicklungsarchiv) mit dem Zielprofil für 0.9.0, die Stufen R0–R5 und die Teilfragen K aus `../0.8.0/REIFEKRITERIEN.md` (Entwicklungsarchiv). Maschine: Mac17,3, zehn Kerne, macOS 27.0.1.

**An diesem Stand neu gelaufen**: `make fremdpakete`, `make dienstdrift`, 73 Drehbücher der Hüllenprobe, der Dauerlauf, die ganze WebKit-Abnahme (40 von 41 Drehbüchern, `4d2bdaae2`), der Lastlauf (`f381125b`) und `make ci`. **Nach den neun Korrekturen an der Oberfläche (§4) liefen erneut**, was sie berühren: `h01rollen`, `h02flaechen`, `h03sprachen`, die vier Personawege, `ap6decks09`, `ap6sammlung09`, `ap8kacheln09` und `ap12feinschliff09` am Stand `61ded082c`, `p12markttisch09` an `189bdd180` — alle zwölf bestanden. Was nicht neu lief, steht mit Commit und Datum da und liegt in seiner Frist (`scripts/belegplan.mjs`).

**Dieselben drei Auslegungen wie bei 0.8.0 (J04):**
1. **R4** heißt „Ziel gemessen, Abweichung benannt". Vergeben wird R4 nur, wenn das Ziel **gehalten** ist oder die Abweichung in der Außenzusage begrenzt steht; sonst R3 mit Vermerk.
2. **Eine Kennzahl ohne Produktlauf zählt R2.**
3. **Ein Beleg gilt für seinen Commit.** Wo seitdem Code dazukam, der das Gemessene berührt, steht es dabei.

**Und eine vierte, aus QA01**: was nur ein Mensch oder zweite Hardware belegen kann, ist **offen** und hält die Stufe, an der es hängt. Kein Agent attestiert es.

## 1 Das Profil

| Dim. | Ziel | Stufe | schwächstes tragendes Glied | Ist · Ziel · n | Beleg |
| --- | --- | --- | --- | --- | --- |
| D1 Umfang, Flächen | ≥ R3 | **R2** | K1.4: über MCP, Kommandozeile und REST gelaufen je Klasse, nicht je Fähigkeit | 379 Fähigkeiten, Lücken 4/118/63/72 begründet; am Fenster **alle 554 Bindungen** gerufen | `FAEHIGKEITEN.md`, `FLAECHENPARITAET.md`, `huellenprobe/vertrag.md` `0a3ea930c` |
| D2 MSE2 | ≥ R3 | **R3** | kein absoluter Zielwert (0188 setzt eine Grundlinie) | fest **9,319** (n 4.266) · Zufall **9,293** [9,186 · 9,399] (n 4.582); gegen 0188 keine Karte schlechter | `belege/fremdpakete.json` `2b2204ffa`, 2026-10-20 |
| D3 Korrektheit | R4 | **R4** | K3.2 Mutation vom 2026-10-10, vor 0.9.0-Phase 6–8 | Mutation 99,1 % (111/112) · ≥ 80 %; Naht 518 von 554 gelungen (93,5 %) · ≥ 90 %, alle 554 am Fenster gerufen; `make ci` vor dem Push, für genau diesen Stand (§4) | `zufallsmutation.md` `80ccb6c49`; `huellenprobe/vertrag.md` `0a3ea930c` |
| D4 Bedienbarkeit | R3 | **R2** | K4.1: A01–A18 in Teilaufgaben, nicht als Ganzes an dieser Fassung | vier Personawege als Ganzes bestanden (2026-10-20); A01–A18 „teilweise" (2026-10-10) | `huellenprobe/g01*.md` `61ded082c`; `belege/arbeitswege-phase-b.md` |
| D5 Barrierefreiheit | R3 | **R3** | K5.4 gesprochenes VoiceOver offen (Mensch) | Tastatur je Rolle 14/14, Flächen 46/46 | `huellenprobe/h01rollen.md`, `h02flaechen.md` `61ded082c`; `webkitprobe/ux07-baum.md` `4d2bdaae2` |
| D6 Stabilität | R4 | **R3** | K6.2 Speicher: auch unter der vorab geänderten Regel (Vermerk 0204) stetig, +11,8 %/h; was wächst, gibt WebKit nach der Warnung frei | 125 min, 1.496 Runden: kein Absturz, nichts verloren, Fußabdruck max 816 MiB · ≤ 1.536, +11,8 %/h · ≤ 10 %, nach der Warnung WebKit 242 MiB · < 455; Wiederanlauf 32/32, Panik und Synchronpanik bestanden | `huellenprobe/dauerbetrieb.md`, `wiederanlauf09.md`, `panik09.md`, `syncpanik09.md` |
| D7 Leistung | R4 | **R4** | — (seit #2532 hat jeder Kernweg ein Budget; #2538 behoben) | Start kalt P95 580 ms · ≤ 2.000; Feld → Bild P95 111 ms · ≤ 300; Set 3.788 Karten P95 249 ms · ≤ 1.000; Leerlauf ≤ 0,16 % CPU · ≤ 1 %; Fußabdruck max 1.425 MiB · ≤ 1.536 | `huellenprobe/budgets09.md`, `leerlauf09.md`, `kennzahlen09.md` `718bb0ca9`; `budgetswege09.md` `d9109cd67` |
| D8 Integrität | R4 | **R4** (bedingt) | K8.3 am Bündel der Weg, kein Vergleich je Objektart | Sicherung 25/25, Aufstieg 30/30, Dienstdrift 9/9; Archiv bytegleich | `huellenprobe/sicherung09.md`, `aufstieg09.md` `718bb0ca9`; `belege/dienstdrift.json` `d8e19ab` |
| D9 Sicherheit | R4 | **R4** | nicht gelaufen: zwei Geräte, fremde Netze, ein Angreifer mit eigenem Zeugnis (Hardware) | Pentest 2026-10-02, Nachtest SEC01 2026-10-16, offen 0; Fuzzen 60 Ziele (2026-10-11), drei Funde am selben Tag behoben; Lastlauf 2026-10-20, 25 Fälle, kein Befund | `PENTEST.md`, `fuzzkampagne.md`, `lastlauf.md` `f381125b` |
| D10 Netz | ≥ R3 | **R2** | zwei Geräte, Router, VPN offen (Hardware) | Netzpartie 32/32, Netzbild 18/18 auf einem Rechner | `huellenprobe/netzpartie09.md`, `netzbild09.md` `718bb0ca9` |
| D11 Sprachen | ≥ R3 | **R4** | K11.1: Übersetzungen nicht redigiert | acht Sprachen am Fenster, 51/51 | `huellenprobe/h03sprachen.md` `61ded082c` |
| D12 Plattform | ≥ R3 | **R2** | Handprüfung 6 und 7 offen (Mensch) | 15.2 im Gast (2026-10-11) | `huellenprobe/untergrenze-15.2.md` `2ca9a871c` |
| D13 Betrieb | R4 | **R3** | K13.1 Konsole von Hand offen | Betriebsband 16/16, OPS 16/16, Fehlerform 7/7 | `huellenprobe/betriebsband09.md`, `ops09b.md`, `fehlerform09.md` `718bb0ca9` |
| D14 Pflege | ≥ R3 | **R3** | Pipelineziel 20 min verfehlt (PERF03) | 1.656 bis 2.161 s je Lauf (2026-10-20) · ≤ 1.200 s | `PIPELINE.md` |

## 2 Die Dimensionen

**D1.** K1.1 `FAEHIGKEITEN.md`: 379 Fähigkeiten, am Fenster 375, über MCP 261, an der Kommandozeile 316, über REST 307; jede Lücke hat ihren Grund, ein Wächter zählt nach. K1.4: am **Fenster** ruft der Vertragsrundgang heute **alle 554 Bindungen** über die echte Brücke (`vertrag`, 518 mit gelungener Antwort in der Form der Typen, 36 fachliche Absagen in der erwarteten Form, 0 Abweichungen). Über MCP, Kommandozeile und REST laufen die Flächenläufe **je Klasse** (`FLAECHENPARITAET.md`, 32 Klassen mit mehr als einer Fläche), nicht je Fähigkeit. **Dafür R2** — dieselbe Lücke wie bei 0.8.0.

**D2.** K2.1/K2.2 `fremdpakete.json` am Stand `2b2204ffa` (2026-10-20 10:58–12:11, sauberer Baum): PASS, 83 von 92 Fällen gemessen; übersprungen wie bei 0.8.0 die Vorrichtungen, `MSE_BUENDEL` und `MSE_MTGJSON`. Feste Kohorte **9,319**, Median 8,577, P99 33,945; Zufallsstichprobe **9,293** mit Band [9,186 · 9,399], je Satz 99,90 %, über alle zehn 99,25 %. Gegen die Grundlinie aus 0188 ist seit #2396 **keine Karte schlechter** (Warlock Class 10,59, Cleric Class 10,99). Ein absolutes Ziel gibt es bewusst nicht (0188) — **R3**.

**D3.** K3.2 Mutation 99,1 % (111 von 112, Wilson 95,1–99,8 %) am 2026-10-10 (`80ccb6c49`), in der Frist von 60 Tagen, aber vor den Phasen 6–8. K3.3/K3.4: der Vertragsrundgang ruft am nativen Fenster alle 554 Bindungen (Stand `0a3ea930c`); **dafür mussten zwei Rufe ihren neuen Argumenten folgen** (`HandelSchritt` seit #2496, `PartieBeitreten` seit #2377) — Wails wies sie mit falscher Argumentzahl ab. **Seit #2528 hält ein Wächter die Argumentzahl jedes Rufs gegen die Bindungen** (`scripts/vertragsargumente.test.mjs`). `make ci` läuft vor dem Push für genau diesen Stand; der `pre-push`-Hook weist einen anderen ab (§4). **R4.**

**D4.** K4.1: die Bilanz A01–A18 ist vom 2026-10-10 (`arbeitswege-phase-b.md`, „zugeordnet, nicht neu gefahren"), jeder Weg „teilweise". Am 2026-10-20 liefen die **vier Personawege** aus G01 als Ganzes (`g01einsteiger`, `g01fan`, `g01admin`, `g01entwickler`, je bestanden) und die Arbeitspakete der Iteration (`ap5`–`ap12`, Tisch, Markt, Decks — alle bestanden). Das Ziel „die Arbeitswege als Ganzes am Fenster" ist für A01–A18 **nicht** neu belegt. **R2.**

**D5.** K5.3 Tastatur: `h01rollen` (ein Kernweg je Rolle nur mit der Tastatur, 14/14) und `h02flaechen` (46/46). K5.1 axe über jede Seite in der Pipeline. K5.4 **gesprochenes VoiceOver**: offen, Handprüfung. K5.2 Zugänglichkeitsbaum je Seite am nativen WebKit: `ux07-baum` (WebKit-Abnahme an diesem Stand, bestanden). **R3** — R4 verlangt einen Kernweg mit gesprochenem VoiceOver.

**D6.** Zwei Dauerläufe am 2026-10-20, je 125 Minuten aus Bearbeiten, Navigation, Sammlung, Tisch und Ausgabe: dieselbe Kennung am Ende, jede Runde gelang, keine getippte Änderung ging verloren, der Fußabdruck nach 120 Minuten blieb bei höchstens 790 und 789 MiB (≤ 1.536). **Die Speicherregel hält in keinem** (#2527). Der erste (`34ef8a1a4`) nahm die Probe mit einer Uhr alle 30 s, und weil eine Runde rund fünf Sekunden dauert, traf sie über lange Strecken denselben Teil der Runde: nach Teil lagen die Mediane zwischen 676 MiB (Bearbeiten) und 893 MiB (Tisch), und seine Abschnitte 826 · 806 · 832 · 873 · 739 verglichen Teile statt Zeiten. `bewertung` erklärt so einen Lauf seither für ungültig, und das Drehbuch nimmt die Probe zwischen zwei Runden. Der zweite (`589bdf8e5`) misst so: **507 · 540 · 573 · 588 · 590 MiB, +12,3 % je Stunde, stetig.** Der Webinhalt wächst in der ersten Stunde um rund 84 MiB und danach langsamer (0,37 MB/min ab Minute 60); die Anwendung wächst gleichmäßig um 0,166 MB/min — dieselbe Rate wie das Restwachstum aus #2312 in den Bausteinen von macOS und der Go-Laufzeit. **Der erste Lauf erlaubt wegen seiner Abtastung keine belastbare Aussage über ein Wachstum; der zweite zeigt es** — ein kontrollierter Vergleich ist es nicht, zwischen beiden Ständen änderte sich auch Code der Oberfläche. Ob der Webinhalt ihn wieder freigibt, belegt erst ein Lauf mit Speicherwarnung am Ende (`MSE_DAUER_WARNUNG=1`). **Danach ein dritter Lauf unter Vermerk 0204** (`fe167f897`, Regel vor dem Lauf geändert, Entscheidung des Auftraggebers): kein Absturz, nichts verloren, Fußabdruck höchstens 816 MiB; nach der Aufwärmstunde **557 · 590 · 601 MiB, +11,8 % je Stunde, stetig — die Regel hält wieder nicht.** Der Webinhalt wuchs diesmal erst in der zweiten Stunde (0,65 MB/min) und **gab zurück**: nach zwei Minuten Ruhe 311 MiB, nach der Speicherwarnung 242 MiB, unter seinem ersten Abschnitt (455). Was wächst, ist damit ein Zwischenspeicher, kein Leck; die Grenzen bleiben, wie sie vor dem Lauf standen, und die Zusage ist begrenzt (README, „Was diese Fassung nicht belegt“). **R3.** Dazu am selben Stand: `wiederanlauf09` 32/32, `panik09`, `syncpanik09` (nach einer Panik in einem synchronen Ruf kommt der Wartende frei, seit Wails beta.30), `fehlergrenze09`.

**D7.** `budgets09` (`718bb0ca9`): Start kalt P95 580 ms (≤ 2.000), warm 522 ms, Set mit 3.788 Karten P95 249 ms (≤ 1.000), Feld → Vorschaubild P95 111 ms (≤ 300), Sammlung 25.000 Zeilen P95 153 ms (≤ 500), Ausgabe 60 Karten 5,96 s, Druckbogen 7,97 s, Zurückspielen 25.000 Zeilen 0,56 s (je ≤ 60 s), Fußabdruck nach jedem Weg max 1.425 MiB (≤ 1.536), Bündel 257,6 MiB (≤ 500). `leerlauf09`: sichtbar 0,16 % CPU, ausgeblendet und minimiert 0,01–0,02 % (je ≤ 1 %). `kennzahlen09`: jede Kennzahl P95 ≤ 500 ms. **Alle Budgets des Messvertrags gehalten**, seit #2532 auch für die Kernwege, die nach ihm kamen (`budgetswege09`, `d9109cd67`, 25.000 Exemplare, Bert als zweiter Prozess): Markttisch aufschlagen P95 371 ms (≤ 1.000), Anbieten bis zum neuen Stand P95 362 ms (≤ 2.000), Kennzahlleisten Set 225, Sammlung 62, Deck 219, Pakete 336 ms (je ≤ 500), Fußabdruck höchstens 1.005 MiB (≤ 1.536). **Der Messlauf fand einen Fehler**: das eigene Fach des Markttischs baute jedes Exemplar als Knopf auf, und das Aufschlagen dauerte 23 bis 45 s (#2538, behoben — über 200 Karten zeigt ein Fach die ersten, eine Suche und „Mehr zeigen“). **R4.**

**D8.** `sicherung09` 25/25 (Alter, Plan, zwei Prozesse, Zurückspielen als ganzer Weg, volle Platte), `a02` (nach dem Zurückspielen benutzbar), `aufstieg09` 30/30 (Inhalte je Tabelle beim Aufstieg), `dienstdrift` 9/9 gegen die echten Dienste (`d8e19ab`). K8.3: die Sicherung ist der **ganze Bestand** in einem Archiv, und es gibt heraus, was hineinging (`internal/sicherung/archiv_test.go`) — damit kommt jede Objektart zurück; K8.4 jede freie Eingabe geprüft oder mit Grund ungeprüft (`FELDPRUEFUNG.md`, Wächter). **Ein Vergleich je Objektart nach dem Zurückspielen am Bündel** fehlt; belegt ist dort der Weg. **R4 (bedingt).**

**D9.** `PENTEST.md`: Lauf am 2026-10-02, Nachtest SEC01 am 2026-10-16, offen 0; benannt und nicht gelaufen: zwei Geräte, fremde Netze, ein Angreifer mit eigenem Zeugnis, Last über die Drosselgrenzen. Fuzzkampagne 2026-10-11 (`a397d98dc`): 60 Ziele, drei Funde, alle am selben Tag behoben (#2227) — ein Nachlauf der ganzen Kampagne steht aus. Lastlauf am 2026-10-20 (`f381125b`, go1.27.2): 25 Fälle über REST und MCP, mit und ohne Kontingente, kein Befund; ohne Kontingente REST lesend 6.139 Rufe je Sekunde bei p95 2,1 ms, MCP über stdio 12.094 je Sekunde, und nach der Überlast antwortet der Dienst wie vorher. GO-2026-5932 ist mit Datum entschieden (`ABHAENGIGKEITEN.md`). **Damit sind die drei Lücken geschlossen, die das Tor von 0.8.0 für D9 nannte** — die Fälle gegen G08, G09 und das Netz (SEC01), ein Lastlauf mit go1.27.2 und die Entscheidung zu GO-2026-5932. **R4** — R5 verlangte Fuzzkampagne und Pentest wiederkehrend; ein Nachlauf der ganzen Kampagne steht aus.

**D10.** `netzpartie09` 32/32 und `netzbild09` 18/18 — zwei Anwendungen auf **einem** Rechner. Zwei Geräte, Router-NAT, VPN, Schlaf und Netzwechsel: offen, Handprüfung und Hardware (#1998). **R2.**

**D11.** `h03sprachen` 51/51: acht Sprachen, jede registrierte Seite im kleinsten Fenster, keine waagerechte Rolle, keine abgeschnittene Beschriftung ohne erreichbaren vollen Text. Übersetzungen nicht redigiert (`SPRACHINVENTAR.md`). **R4** für das Fenster, die Redaktion bleibt Menschenwerk.

**D12.** macOS 15.2 im Gast gestartet (`untergrenze-15.2.md`, `2ca9a871c`, 2026-10-11); Aufstieg im Gast ebenso. **Handprüfung 6** (Erststart aus dem Abbild über Gatekeeper) und **7** (Bestand danach) von einem Menschen: offen. **R2.**

**D13.** `betriebsband09` 16/16, `ops09b` 16/16, `fehlerform09` 7/7, `ops03` (alles widerrufen) bestanden. K13.1 Unified Log in der **Konsole**: Handprüfung offen. **R3.**

**D14.** Pipeline am 2026-10-20 je Lauf 1.656 bis 2.161 s (die Läufe der Hebungen aus #1784) gegen das Ziel 1.200 s (PERF03) — **verfehlt**, zum Teil mit kaltem Zwischenspeicher nach Werkzeughebungen. Zeilengrenze, Wächter und Doku halten; Abhängigkeiten gehoben (#1784, alles außer npm 11.21/12 und tailscale, je mit Grund). **R3.**

## 3 Das Freigabetor

Tor nach `KRITERIEN.md` §1: D3, D6, D7, D8, D9, D13 **R4**; D4, D5 **R3**; D1, D2, D10, D11, D12, D14 **≥ R3**.

| Dim. | Tor | Ergebnis | Was bis zum Tor fehlt — sonst die Zusage begrenzen |
| --- | --- | --- | --- |
| D1 | R3 | **nicht erfüllt** | Läufe je Fähigkeit über MCP, Kommandozeile und REST; sonst „gelaufen je Klasse" |
| D2 | R3 | erfüllt | — |
| D3 | R4 | erfüllt | — (der Wächter für die Argumentzahl steht seit #2528) |
| D4 | R3 | **nicht erfüllt** | A01–A18 als Ganzes an dieser Fassung; sonst „in Teilaufgaben geprüft, vier Personawege als Ganzes" |
| D5 | R3 | erfüllt | gesprochenes VoiceOver (Mensch) für R4 |
| D6 | R4 | **nicht erfüllt** | ein Dauerlauf, der die Regel hält — auch unter Vermerk 0204 nicht; belegt ist, dass der Webinhalt nach einer Warnung freigibt (#2527) |
| D7 | R4 | erfüllt | — (Budgets für Markttisch, Handelszug und Kennzahlleisten seit #2532, gehalten) |
| D8 | R4 | erfüllt, **bedingt** | ein Vergleich je Objektart nach dem Zurückspielen am Bündel |
| D9 | R4 | erfüllt | — (zwei Geräte und fremde Netze bleiben benannt) |
| D10 | R3 | **nicht erfüllt** | zwei Geräte, Router, VPN — Hardware; bereits begrenzt |
| D11 | R3 | erfüllt | — |
| D12 | R3 | **nicht erfüllt** | Handprüfung 6 und 7 (Mensch) |
| D13 | R4 | **nicht erfüllt** | die Konsole von Hand (Mensch) |
| D14 | R3 | erfüllt (Ziel der Dauer verfehlt) | die Teststufe schneiden oder das Ziel begründet setzen |

## 4 Was diese Erhebung nicht prüft

- **Menschenprüfungen** (`HANDPRUEFUNGEN.md`): Konsole und Instruments, zweiter Bildschirm, zwei Geräte im WLAN samt VPN und Schlaf, gesprochenes VoiceOver je Rolle, Bewegung und Kontrast aus den Systemeinstellungen, der Touch-ID-Sensor, **Handprüfung 5** (Druckbogen auf Papier), **6** (Erststart aus dem Abbild) und **7** (Bestand danach), zwei Rechner hinter verschiedenen Routern.
- **Eine Maschine** (Mac17,3, macOS 27.0.1); der Gast 15.2 nur aus dem Lauf vom 2026-10-11.
- **Die Ausgangsmessungen** `ausgangswerte09handeln`, `-netz` und `-spielen` messen die Oberfläche vom 13./14. Oktober. Sie sind **nicht** neu gefahren: Phase 7 und 8 haben genau diese Wege umgebaut, und ihre Belege bleiben das „vorher".
- **`ap12ziehquelle09`** braucht eine Hand, die zieht; sein Beleg vom 2026-10-17 bleibt.
- **Nicht neu gefahren, in der Frist**: Mutation (2026-10-10), Reichweite der nativen Bindungen (2026-10-09), Fuzzkampagne (2026-10-11), Leistungsvergleich mit MSE2 (2026-10-05).

**Was die Gesamtabnahme fand.** Zwei Arten, und sie stehen getrennt:

*Fehler des Produkts*, gefunden von der WebKit-Abnahme, die seit dem 2026-10-06 nicht mehr ganz gelaufen war. Behoben auf diesem Zweig, jeder mit einem vorher roten Test oder Wächter:

1. **Die Seite warf beim Öffnen einen Fehler** — `useFensterrest` zeichnete im Rückruf des `ResizeObserver` synchron neu, und WebKit meldete „ResizeObserver loop completed with undelivered notifications“ auf jeder Seite mit Arbeitsplatz. Jetzt folgt die Höhe im nächsten Bild.
2. **Die Knopfzeile im Seitenkopf brach bei sehr großer Schrift nicht um** — „Bei Bo mitspielen“ stand halb außerhalb, die Seite rollte waagerecht.
3. **Der Markttisch legte drei Spalten nach dem Fenster an** (`lg:`) und nicht nach seinem Platz — bei 1024 Punkten standen sie in 343, und die Seite *Handeln* rollte waagerecht. Jetzt misst er am eigenen Platz (`@container`).
4. **Die Seite *Handeln* teilte sich ebenso nach dem Fenster** — bei sehr großer Schrift nahm die Liste 400 Punkte, und dem Handel blieben 173.
5. **Nach Escape auf der Löschfrage im Sammlungsinspector ging der Fokus verloren** (UX-10.AK-04) — der Knopf war während der Frage abgebaut.
6. **Ein Klick ins Ansichtssegment nahm den Fokus aus der Liste** (UX-11) — seit die Ansicht ein Segment ist (#2363); vorher waren es Knöpfe, und die nehmen in WebKit keinen Fokus.
7. **Der Leertext der Turniere hatte 179 Zeichen** (UX-15, Grenze 140) — und die Klasse dahinter drei weitere, die die Abnahme nicht sah, weil der Prüfbestand dort nicht leer ist: *Decks* 144, *Sammlung* 145, Symbolbibliothek 142. Gekürzt; `leertexte.test.ts` hält jeden Leertext des deutschen Katalogs darunter.
8. **Die Erklärung des Vertrags stand mit 189 Zeichen im ersten Blick** (UX-15) — seit *Handeln* einen Handel von selbst aufschlägt (P13). Jetzt erster Satz sichtbar, Rest hinter dem Fragezeichen.
9. **Das Popover gab den Fokus nicht seinem Auslöser zurück, wenn er beim Aufgehen auf `body` stand** (UX-08.AK-01) — gefunden beim Nachgehen von `ux08-popover`; dessen Rot war am Ende ein Wettlauf im Drehbuch, der Fehler im Popover ist trotzdem einer und rot gesehen.

**Im Fenster 1024 × 700 lag der aktuelle Zug unter der Falz** (F09, #2526) — Rücknahmeknopf, Tischleiste und Kopf fraßen die Reserve. **Entschieden und behoben am 2026-10-20**: die Rücknahme unter dem Zug, die Tischleiste dort in einer Zeile mit „Darstellung“, und der Tisch lässt dem Zug 13rem. **Danach lief die ganze WebKit-Abnahme erneut: 41 von 41** (`d33405833`).

*Drehbücher, die einer gewollten Änderung nicht gefolgt waren* — kein Fehler des Produkts, nachgezogen: in der Hüllenprobe `druckfeld09`, `geste09`, `tauschneu09` (#2378), `p11sicht09` (#2494), `vertrag` (#2496, #2377), `a04`, `g07palette` und `p12markttisch09` (die gemerkte Ansicht des Handels überlebte das Drehbuch davor); in der WebKit-Abnahme 20 Drehbücher — der Vorhang (#2375), der Klick tappt (#2372), Menü »…« an Deck und Platz und die Ansicht als Segment (#2363), die Schublade des Setassistenten (#2367), die Herkunft als Marke (#2368), Falten (#2465), der Kasten »Kennung von Hand« (#2378), die Rolle des Protokolls (#2372), die Einführung bleibt zu (#2428), »Schnell spielen« (#2374); `ux10-aktionen` und `ux10-vorleser` messen den Aktionskatalog jetzt am Sammlungsinspector, weil am Deck kein Löschknopf mehr steht. **Ein abgebrochenes Drehbuch ließ Zustand im gemeinsamen Bestand liegen** (Farbwelt, Schriftgröße) und färbte die folgenden — zwei Kontrast- und Längenbefunde waren solche Folgen. Zwei blieben als Befund und sind behoben: der Vertragsrundgang prüft die Argumentzahl jetzt gegen die Bindungen (#2528), und die Vorwahl der Palette springt nicht mehr, während die Daten für „Tun“ nachkommen (#2529).

**Und die WebKit-Abnahme schrieb in die echten Vorgaben dieses Rechners**: `defaults` folgt keinem umgelenkten `HOME`, und der Kennungsumzug legte die echten Vorgaben in eine neue Domäne. Behoben mit `MSE_DIENST_GESPERRT=an` in Abnahme und Sichtprüfung; ein Wächter hält jeden Fensterstart mit eigenem Bestand dagegen (`dienstsperre_test.go`). Die echte Datei liegt beiseitegelegt, nicht gelöscht.

**Zuordnung und Gesamtreview** (zweites Abnahmekriterium von QA01). `ZUORDNUNG.md` (Entwicklungsarchiv) ordnet jeden der 88 Befunde aus den drei Audits und dem Zweitblick seinen Paketen zu und jedes der 58 Pakete seinem Vermerk und seinen Nachtests — lückenlos, und ein Test hält es so (`scripts/befundzuordnung.test.mjs`); Ausnahmen stehen mit Grund darin. **Das Gesamtreview über die Nähte** (Codex `gpt-6.1-sol`, Denkstufe mittel, nur lesend; `zweitblicke/9-gesamtreview.md` (Entwicklungsarchiv)) fand acht Befunde, sechs S2 und zwei S3. **Der schwerste war eine Regression aus #2396**: eine Bearbeitung, deren Wertskript ein anderes gespeichertes Feld nachzieht, galt im Fenster als fremde Änderung — in einem M15-Set jede Änderung der Manakosten. Behoben sind 1 bis 4 und 8; 5 ist entschieden — die Bilder der Handelszeilen bleiben, wie sie sind (#2531, 2026-10-20) —, 6 steht begrenzt in der Außenzusage (#2532), 7 ist mit Lebensdauer und Grenze benannt. **Der Nachlauf** (`9-gesamtreview-nachlauf.md` (Entwicklungsarchiv)) fand sechs weitere, davon einen Fehler in einer Behebung (behoben) und drei Grenzen derselben Klasse, die als #2533, #2534 und #2535 offen sind. **Damit ist das zweite Abnahmekriterium erfüllt**; was die Klasse dahinter offen lässt, steht in diesen Issues.

## 5 Die Entscheidung

**Sie liegt beim Auftraggeber** (REL #2385). `KRITERIEN.md` lässt ein nicht erfülltes Tor die Freigabe anhalten **oder** zu einer dokumentiert begrenzten Zusage führen — wie bei 0.8.0 (§5 dort).

**Was diese Erhebung dazu sagt, ohne zu entscheiden:**

- **Das volle Tor hält 0.9.0 nicht.** Drei Dimensionen fehlen an der Sache — D1 (Flächenläufe je Klasse statt je Fähigkeit), D4 (A01–A18 nicht als Ganzes an dieser Fassung) und D6 (die Speicherregel des Dauerlaufs, #2527) —, drei an Mensch oder Hardware: D10 (zwei Geräte, Router, VPN), D12 (Handprüfung 6 und 7) und D13 (die Konsole von Hand).
- **Was ein Agent tun konnte, ist getan**: #2526 behoben (WebKit-Abnahme 41 von 41); #2532 gemessen — alle Budgets gehalten, nachdem #2538 (Markttisch 45 s) behoben war; #2527 nach der Entscheidung des Auftraggebers mit geänderter Regel gefahren — sie hält nicht, der Webinhalt gibt nach einer Warnung frei. Die Entscheidung zur Freigabe ist gefallen: **begrenzt**, wie bei 0.8.0 (#2385).
- **Was nur ein Mensch tun kann**: Handprüfung 5 (Druckbogen auf Papier), 6 (Erststart aus dem Abbild mit „Dennoch öffnen“) und 7 (Bestand danach), gesprochenes VoiceOver je Rolle, die Konsole, zwei Geräte.
- **Wer trotzdem freigibt**, gibt mit begrenzter Zusage frei wie bei 0.8.0 — jede Begrenzung steht in §3 in der rechten Spalte. Ein Release oder ein Tag ist in dieser Erhebung nicht entstanden.
