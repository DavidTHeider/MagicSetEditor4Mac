# Beleg: Arbeitswege A01–A18 und vier Personas am Schlussstand von Phase B — B03

**B03** (#2222),
Abnahmekriterium: »A01–A18 mit Teilaufgaben statt grünem Sammelhaken; vier
Personas und ihre Hürden. Keine Übernahme alter bestanden-Werte«
(`docs/audits/0.8.0/ergebnis/PLAN-0.8.0.md:128`). Erhoben am 2026-10-10 am
Zweig `feat/2206-phase-b`.

## Methode und Grenzen

**Zugeordnet, nicht neu gefahren.** Dieser Beleg liest die Drehbücher
(`scripts/huellenprobe/<name>.mjs`) und ihre Belege
(`docs/belege/huellenprobe/<name>.md`) des Schlusslaufs und ordnet sie den
Teilaufgaben der achtzehn Arbeitswege zu. Er startet nichts. Der Lauf ist
`2026-10-09T19-47-35Z-46912` am Commit `1c0b74221` (Mac17,3, macOS 27.0.1, Bündel
aus `scripts/bauen.sh abnahme`, Bauabdruck `1b6e1cccf72a`): **106 Drehbücher
gefahren, 102 bestanden, 4 rot** — `ai02`, `dauerbetrieb`, `vertrag`, `vollbild`
(Manifest `.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/manifest.json`, Feld
`faelle`; liegt lokal und nicht im Git, `.gitignore:119`). `vollbild` und
`dauerteile` sind **Einzelläufe** am Commit `256cebec9`
(`2026-10-09T22-29-24Z-35842`, `2026-10-09T22-30-28Z-36573`): der Beleg von
`vollbild` hat den roten des Schlusslaufs ersetzt, `dauerteile` fährt nur, wer es
nennt, und lief im Schlusslauf nicht.

**Was »geprüft« heißt:** ein Drehbuch fährt die Teilaufgabe **am nativen
Fenster**, und sein Beleg ist grün. Die vier Stände:

| Stand | Heißt |
| --- | --- |
| geprüft | ein grünes Drehbuch fährt die Teilaufgabe am Fenster |
| teilweise | ein Drehbuch fährt einen Teil — oder fährt sie nur über die Kommandozeile oder eine Bindung statt über die Bedienung |
| nicht geprüft | kein Drehbuch fährt sie |
| rot | ein Drehbuch fährt sie und zeigt einen Befund oder scheitert |

**Keine Nutzerstudie.** Die Drehbücher sind Maschinen; die Daten sind künstlich
(Saat `probe-werkstatt`, kanonisches Set »Black« nur zum Rollen). **Es gibt keine
menschliche Erfolgsquote**, und das Audit hält es ebenso fest: Beobachtungen sind
Experteninspektionen unter angenommenen Rollen
(`docs/audits/0.8.0/ergebnis/AUDIT.md:16`, `docs/audits/0.8.0/ergebnis/README.md:12`).

**Der Zuschnitt der Teilaufgaben ist eine Entscheidung dieses Belegs.**
`arbeitswege.json` führt je Weg einen Satz Beleg und keine nummerierten
Teilaufgaben, obwohl `AUDIT.md:27` (im selben Ordner) auf sie verweist. Abgeleitet sind sie aus
`docs/inventur.json` (`arbeitswege.wege[].offen`), aus `arbeitswege.json`
(`beleg`) und aus dem, was die Drehbücher tatsächlich fahren. Ein anderer
Zuschnitt änderte die Zählung. **Ein Weg ist hier die Summe seiner Teilaufgaben
aus getrennten Drehbüchern, kein durchgehender Lauf eines Menschen.**

**Zeilenangaben** ohne Pfad meinen den Beleg des genannten Drehbuchs unter
`docs/belege/huellenprobe/`. »Auditstand« ist `arbeitswege.json` vom 2026-10-08.

**Seit dem Lauf hat sich das Repositorium bewegt.** Der Zweig steht beim Schreiben
bei `80ccb6c49`, sechzehn Commits über `1c0b74221`. An Produktcode kamen hinzu:
`internal/fenster/menue.go` (`b78a2fe6e`), `internal/fenster/rufbuch.go` und
`rufausgang.go` (`75c90a394`), `internal/app/paketkatalog.go` und
`internal/service/paketuebersicht.go` (`e72161b86`). **Nur `vollbild`,
`dauerteile` und `zustaende` sind nach diesen Änderungen gefahren worden**; die
übrigen Belege, auch `leiste`, `darstellung` und `eintraege`, gelten für den
Stand davor. **`zustaende` ist ein zweiter Lauf**: der Schlusslauf lief mit einer
Fassung der Probe, die der Zweitblick über Phase B beanstandete (`9cdd66a39`);
die 150 Zellen sind am Commit `fa888de59` neu gemessen (Lauf
`2026-10-10T00-11-39Z-60703`, Bauabdruck `b2ea4168becd`, Beleg in `80ccb6c49`) —
dazu der Abschnitt »Die 150 Zellen«. Die Zahlen von `dauerbetrieb` hat die
Auswertung danach korrigiert (`docs/LEISTUNGSSCHNITTE.md` §9.1); der Beleg selbst
trägt noch die erste.

## Übersicht

| Weg | Aufgabe | Auditstand | geprüft | teilweise | nicht geprüft | rot | Stand jetzt |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| A01 | Set und Karte mit Bild/Text anlegen | teilweise | 5 | 0 | 1 | 0 | teilweise |
| A02 | Speichern, exportieren, Neustart, erneut öffnen | teilweise | 3 | 1 | 1 | 0 | teilweise |
| A03 | Rücknahme, unbekannte Felder, Normal-/Expertenansicht | teilweise | 4 | 0 | 2 | 0 | teilweise |
| A04 | Vorlage kopieren, validieren, korrigieren, verwenden | **nicht geprüft** | 5 | 0 | 1 | 0 | **teilweise** |
| A05 | Sammlungsimport und Dubletten | teilweise | 4 | 0 | 2 | 0 | teilweise |
| A06 | Sammlung suchen/filtern/Ansichten/Mengen | teilweise | 4 | 1 | 1 | 0 | teilweise |
| A07 | Deck aus Sammlung und Testpartie | teilweise | 5 | 1 | 0 | 0 | teilweise |
| A08 | Turnier mit Ergebnissen | teilweise | 5 | 1 | 0 | 0 | teilweise |
| A09 | Zweite lokale Instanz, Einladung, Trennung | teilweise | 4 | 1 | 1 | 0 | teilweise |
| A10 | Testhandel und Zustandsfolge | teilweise | 4 | 1 | 1 | 0 | teilweise |
| A11 | Spieltisch Karten, Tokens, Counter, Anhänge | teilweise | 4 | 0 | 2 | 0 | teilweise |
| A12 | Gegnerperspektive und verdeckte Daten | teilweise | 3 | 0 | 1 | 0 | teilweise |
| A13 | Nutzer/Profile/Nur-Lesen/Rechteentzug | teilweise | 4 | 0 | 1 | 1 | teilweise, **mit roter Teilaufgabe** |
| A14 | Einstellungen, KI, CLI, Diagnose | teilweise | 3 | 1 | 1 | 1 | teilweise, **mit roter Teilaufgabe** |
| A15 | Testbestand sichern und restaurieren | **fehlgeschlagen** | 4 | 1 | 1 | 0 | **teilweise** |
| A16 | Hilfe, Lizenz, Datenschutz, Sprachinfo | teilweise | 4 | 1 | 0 | 0 | teilweise |
| A17 | Fenstergrößen, Sprachen, Schrift, Bewegung | teilweise | 4 | 1 | 1 | 0 | teilweise |
| A18 | Tastatur/Fokus/Navigation | teilweise | 4 | 0 | 2 | 0 | teilweise |
| | **zusammen (104 Teilaufgaben)** | | **73** | **10** | **19** | **2** | |

**Keiner der achtzehn Wege steht als Ganzes auf geprüft.** Verändert gegenüber dem
Audit hat sich das Wort bei zwei Wegen: A04 (von »nicht geprüft«) und A15 (von
»fehlgeschlagen«) stehen jetzt auf teilweise. Bei den übrigen sechzehn ist das
Wort gleich geblieben; geändert hat sich, **welche** Teilaufgaben offen sind. Die
Zahl 73 von 104 ist eine Zeilenzählung und kein Reifemaß: die Zeilen wiegen nicht
gleich. Die Zahlen der Übersicht sind aus den Tabellen dieses Belegs gezählt, nicht
aus den Zeilen »Gesamt« abgeschrieben.

## Die Wege

### A01 — Set und Karte mit Bild/Text anlegen

Auditstand 2026-10-08: teilweise (`arbeitswege.json:8`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A01.1 Neues Set anlegen, den Stil aus der Auswahl wählen | `uxr04`, `fix-f02` | »Neues Set« springt zum Namen, der Stil ist ein Kombinationsfeld mit gewähltem Namen; Stil »Mit Kunstfenster« (probe-kunst) | geprüft |
| A01.2 Karte anlegen, Name und Regeltext füllen, die Vorschau folgt | `ausgangswerte`, `budgets`, `fp02` | Kartenanlage 4 Handlungen, 0 Fehlversuche, 0,8 s Maschinenzeit (`ausgangswerte.md:11`); Feld bis Vorschaubild Median 27 ms, 0 ohne Bild (`budgets.md:47`); Text mit Leerzeile steht nach Neustart wieder da (`fp02`) | geprüft |
| A01.3 Ein Bild einsetzen (Dialog, Ziehen auf das Feld) | `fix-f02`, `kernwege` | »Bild wählen« öffnet den Dialog von macOS, das Set trägt danach das Bild (kunst1); ein Bild auf dem Feld »kunst« führt zum Zuschnitt | geprüft |
| A01.4 Ein Setsymbol zeichnen und in die Bibliothek aufnehmen | `fix-f02`, `dom02`, `dom02ak02` | das Set trägt symbol1.mse-symbol; Bibliothek: aufnehmen, duplizieren, einsetzen; Löschen nennt die Sets; Zeichnung bearbeiten, sichern, vorige Fassung zurückholen | geprüft |
| A01.5 Die Setdatei schreiben und im frischen Bestand öffnen | `fix-f02` | ⌘S fragt wohin (Vorschlag »set.mse-set«), Abbruch schreibt nichts; ⌘O in frischem Bestand: das Set öffnet mit seiner Karte | geprüft |
| A01.6 Eine Karte in einem echten Magic-Stil gestalten | — | alle Drehbücher legen Karten im Saatstil `probe-*` an; »Black« wird nur gerollt (`budgets`), und `readmebilder` bildet Flächen mit einer Kopie eines kanonischen Sets ab — gestaltet wird dort keine Karte | nicht geprüft |

**Gesamt:** 5 geprüft, 1 nicht geprüft — **kein Sammelhaken**. Vorher: teilweise;
jetzt: teilweise. Zu ist der Setdatei-Rundlauf (F02), den `docs/inventur.json` als
»gebaut, nativ nicht abgenommen« führte. Offen: ein echter Magic-Stil.

### A02 — Speichern, exportieren, Neustart und erneut öffnen

Auditstand: teilweise (`arbeitswege.json:17`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A02.1 Die Setdatei schreiben: ⌘S, Sichern unter, Abbruch, Schreibfehler | `fix-f02` | beide Wege gehen denselben Schreibweg; Abbruch schreibt nichts; ein gesperrter Ort nennt Ursache und nächsten Schritt, die alte Datei steht unverändert | geprüft |
| A02.2 Beenden, neu starten, erneut öffnen — Inhalt und unbekannte Felder bleiben | `fix-f02`, `fp02`, `kernwege` | unbekannte Zeile und unbekannter Block bleiben; Text samt Absätzen und der getippten Zeile steht nach ⌘Q und Neustart da; Doppelklick im Finder öffnet das Set (Weg 4) | geprüft |
| A02.3 Druckbogen und Exportvorlage schreiben; ein langer Lauf lässt sich anhalten und wiederholen | `uxrf7`, `fmt03b`, `a04` | Bogen mit genau der Seitenzahl, die der Kasten nannte (4); Exportdatei 272 Bytes gleich Kasten 272; 400 Karten auf 45 Seiten (3.393.501 Bytes), angehalten bei 7 von 400 ohne halben Bogen; ⌘P öffnet »Druckbogen sichern« | geprüft |
| A02.4 Ausgeben als Bilder | `eintraege`, `dauerbetrieb` | der Menüeintrag führt auf die Seite »Bilder«; ein Kartenbild schreibt nur die Bindung ohne Dialog, in jeder der 1.712 Runden gelungen (`dauerbetrieb` ist im Ganzen rot, wegen des Speichers, nicht wegen dieser Zeile) | teilweise |
| A02.5 Die Gegenprobe außerhalb: MSE2 öffnet die Setdatei, ein Drucker druckt den Bogen | — | Dialog und Datei sind belegt; MSE2-Gegenöffnung und Drucker hat kein Drehbuch (so auch `arbeitswege.json:18`) | nicht geprüft |

**Gesamt:** 3 geprüft, 1 teilweise, 1 nicht geprüft — fünf Teilaufgaben, **kein
Sammelhaken**. Vorher: teilweise; jetzt: teilweise. Zu sind ⌘S und »Sichern
unter« als Setdatei (F02) und der Neustart, die `docs/inventur.json` (A02) als
offen führte — **am Abnahmebau und nicht am ausgelieferten Bündel**. Offen: die
Gegenprobe außerhalb und das Schreiben von Bildern über die Bedienung.

### A03 — Rücknahme, unbekannte Felder, Normal-/Expertenansicht

Auditstand: teilweise (`arbeitswege.json:26`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A03.1 Eine fachliche Änderung zurücknehmen: Menü, Sammlung, Verlauf | `eintraege`, `dom03ak03`, `u13` | »Rückgängig« nimmt den Kartennamen im Bestand zurück, »Wiederherstellen« bringt ihn wieder; zwei Exemplare zurück auf »Kiste 4«; nach ⌘Z nennt der Weg weiter den Namen, der Kopf folgt dem Verlauf (`u13`, geprüft ist der Name, nicht das Verschwinden der Karte) | geprüft |
| A03.2 ⌘Z im Textfeld nimmt die Eingabe zurück | `leiste` | »Widerrufen« nimmt die Eingabe im Feld zurück, nicht eine fachliche Änderung; die Seite lässt ⌘Z durch | geprüft |
| A03.3 Rücknahme nach einem Fokuswechsel im Kartenformular | — | `eintraege` ändert **ein** Feld ohne Fokuswechsel; der Zyklus ist G05 (#2252) | nicht geprüft |
| A03.4 Unbekannte Felder überstehen Öffnen und Sichern | `fix-f02` | die unbekannte Zeile in der Karte und der unbekannte Block bleiben | geprüft |
| A03.5 Ein Entwurf übersteht eine fremde Änderung; der Konflikt steht am Feld | `phasec`, `u02b` | Knoten, Entwurf und Fokus bleiben in sechs Fällen (fünf Formulare); vor der Annahme bleibt der fremde Wert im Bestand, »behalten« schreibt den Entwurf (drei Runden) | geprüft |
| A03.6 Normal-, Erweitert- und Expertenmodus des Kartenformulars | — | kein Drehbuch nennt die Modi | nicht geprüft |

**Gesamt:** 4 geprüft, 2 nicht geprüft. Vorher: teilweise; jetzt: teilweise. Das
Menü-Widerrufen wirkte schon im Audit; offen bleiben der Fokuszyklus und die Modi.

### A04 — Vorlage kopieren, validieren, korrigieren, verwenden

Auditstand: **nicht geprüft** (`arbeitswege.json:35–36`: »Vorlagenseite gesehen;
vollständiger Kopie→Korrektur→Karteneinsatz-Weg nicht gefahren«).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A04.1 Eine eigene Kopie einer mitgelieferten Vorlage anlegen | `fmt02` | eigenes Paket mit eigenem Kurznamen, die Werkstatt öffnet ihren Stil; das Original ist Byte für Byte, wie es war | geprüft |
| A04.2 In der Werkstatt Dateien anlegen, umbenennen, löschen, ein Bild hinzufügen; Include und Exportvorlage aus einem Gerüst | `fmt02c`, `fmt02d`, `fmt02e` | neue Datei liegt auf der Platte, alter Name fort, gelöschte Datei in der Sicherung; vier Arten im Anlegeformular; Installer hinaus und wieder herein | geprüft |
| A04.3 Ein Befund erscheint an der Stelle und verschwindet nach der Korrektur | `fmt02b`, `fmt02f` | unbekannte Feldart: Art und bekannte Arten, im sichtbaren Bereich; Tippfehler: Vorschau als veraltet markiert, Korrektur nimmt beides weg — **an zwei Befundarten** | geprüft |
| A04.4 Jemand ändert die Datei auf der Platte: der Konflikt bietet beide Wege | `fmt02b` | »Meine behalten« sichert den eigenen Text, auf der Platte steht er | geprüft |
| A04.5 Die korrigierte Vorlage ist auf der Karte angekommen | `fmt02f` | Vorschau folgt dem Editor, bevor gesichert ist ([200,30,30]); nach dem Sichern ist die Karte im Set rot | geprüft |
| A04.6 Dieselbe Kopie an einem bezogenen Stil des Packs | — | alle Drehbücher kopieren den Saatstil `probe-schlicht` | nicht geprüft |

**Gesamt:** 5 geprüft, 1 nicht geprüft — von »nicht geprüft« auf **teilweise**.
Der Kopierweg, den das Audit nicht fand, ist da; offen ist ein echter Stil aus
dem Pack.

### A05 — Sammlungsimport und Dubletten

Auditstand: teilweise (`arbeitswege.json:44`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A05.1 Ein CSV mit neuer Kennung und Menge 3 legt genau drei Stück an | `fix-f03` | genau drei Stück des neuen Drucks | geprüft |
| A05.2 Dublette: die Zeile erhöht statt eines zweiten Eintrags; ein zweites Aufnehmen fragt nach | `fix-f03`, `dom03` | ein Eintrag mit fünf Stück; die Frage nennt die vorhandene Menge, Escape ändert nichts, »Auf N erhöhen« erhöht, »Zur vorhandenen Zeile« wählt sie | geprüft |
| A05.3 Die eigene Ausfuhr zurücklesen, vorher sehen, was sie trifft | `dom03ak01`, `fix-f03` | Vorschau »2 → 4« ohne Schreiben; »Vorhandene auslassen« lässt die Menge bei zwei; im frischen Bestand steht dieselbe Sammlung | geprüft |
| A05.4 Eine kaputte CSV-Zeile gibt einen fachlichen Bericht | `fix-f03` | der Bericht nennt die Zeile (Zeile 2: »drei« ist keine Stückzahl), ohne technischen Wortlaut und ohne unerklärte Teiländerung | geprüft |
| A05.5 Eine Liste einfügen (Namensliste, ein Name wird nicht gefunden) | — | kein Drehbuch; die Produktionsprobe des Audits brach nach F-01 ab | nicht geprüft |
| A05.6 Import aus Scryfall (Satz, Suche, UUID) | — | die Scryfall-Bindungen sind vom Rundgang ausgenommen (`bindungsreichweite.md`, Ausnahmen) | nicht geprüft |

**Gesamt:** 4 geprüft, 2 nicht geprüft. Vorher: teilweise. Von den drei Fehlern,
die `docs/inventur.json` führte, laufen CSV und Doppelaufnahme; der Listenimport
hat kein Drehbuch.

### A06 — Sammlung suchen/filtern/Ansichten/Mengen

Auditstand: teilweise (`arbeitswege.json:53`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A06.1 Eine Suche blendet Zeilen aus; »Suche und Filter zurücksetzen« holt sie zurück | `dom03ak03` | der Knopf steht neben »1 von 2« und holt beide zurück | geprüft |
| A06.2 Mehrere Zeilen wählen, den Lagerort ändern, zurücknehmen | `dom03ak03` | die Kommandozeile liest »Ordner Rot«, nach »Zurücknehmen« wieder »Kiste 4« | geprüft |
| A06.3 Die Menge einer Zeile ändern | `dom03` | auf drei erhöht: Liste und Inspector zeigen drei; Escape lässt die Menge bei zwei | geprüft |
| A06.4 Eine große Sammlung: ans Ende springen, die Summe lesen | `phaseg2` | 260 Exemplare: der Überblick nennt die Summe der Mengen (519), das nachgeholte Fenster kommt an | geprüft |
| A06.5 Nach Marken filtern | `ux12` | »Nach Marken« lässt nur das rote Deck übrig — **an Decks, nicht an Sammlungszeilen** | teilweise |
| A06.6 Gruppen und Ansichten der Sammlung | — | kein Drehbuch | nicht geprüft |

**Gesamt:** 4 geprüft, 1 teilweise, 1 nicht geprüft. Vorher: teilweise. Die
Audit-Lücken »Suchfilter« und »Mengenänderung« sind zu; »Gruppen« ist offen.

### A07 — Deck aus Sammlung und Testpartie

Auditstand: teilweise (`arbeitswege.json:62`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A07.1 Ein Deck anlegen und eine Karte hineinlegen | `ausgangswerte` | Deckbau 10 Handlungen, 0 Fehlversuche, 1,2 s Maschinenzeit (`ausgangswerte.md:21`); »Ausgangsdeck« steht im Bestand | geprüft |
| A07.2 Ein Deck aus Exemplaren der Sammlung bauen | `dom03ak02` | Deck, Platz und Exemplar legt die Kommandozeile an; im Fenster geht Kopieren und die Wunschliste, nicht das Hineinlegen eines Exemplars | teilweise |
| A07.3 Ein Deck kopieren; was fehlt, auf die Wunschliste | `dom03ak02` | Vorschlag »Aschenrat (Kopie)« mit denselben vier Karten; drei fehlende neu auf der Wunschliste, ein zweiter Klick legt nichts doppelt an | geprüft |
| A07.4 Die Plätze eines Decks bearbeiten | `a04`, `dom03ak03` | Kontextmenü: Eine mehr, Eine weniger, Ins Sideboard, Herausnehmen (rechte Taste und ⇧F10); »−« ruht bei einer Karte | geprüft |
| A07.5 Eine Partie mit dem Deck anfangen und später weiterspielen | `ausgangswerte`, `tischwiedereinstieg`, `abn01tisch` | Spielstart 9 Handlungen, 1,5 s (`ausgangswerte.md:25`); je Erscheinungsbild zehn Zyklen und hundert Besuche in einer Sitzung, jedes Mal eine neue Leinwand mit denselben Karten | geprüft |
| A07.6 Die Testpartie spielen: Karten ziehen | `a04`, `fix-f04` | Hand → Spielfeld, Spielfeld → Friedhof mit dem Zeiger, Escape bricht ab; zwei Auslöser zugleich ordnen (M9) | geprüft |

**Gesamt:** 5 geprüft, 1 teilweise. Vorher: teilweise. Offen bleibt, was das
Audit nannte: der sammlungsbasierte Deckbau über die Oberfläche.

### A08 — Turnier mit Ergebnissen

Auditstand: teilweise (`arbeitswege.json:71`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A08.1 Ein Turnier mit Format anlegen bis zur ersten Auslosung | `dom04ak01` | vier Schritte; »Weiter« ruht mit einer Person; die Kommandozeile liest Format, zwei Namen, eine Runde | geprüft |
| A08.2 Ein Turnier als Vorlage kopieren | `dom04ak01` | die Kopie trägt Namen und Format und keine Runde | geprüft |
| A08.3 Das Ergebnis einer gespielten Paarung berichtigen | `dom04ak02` | die Frage nennt, wer den Platz wechselt, und dass Runde 2 bleibt; »Nicht eintragen« lässt die Rangliste; »Trotzdem eintragen« ändert sie (Platz 1 wechselt, OMW 0.4150) | geprüft |
| A08.4 Ergebnisse mehrerer Runden im Fenster eintragen und weiter auslosen | `dom04ak02`, `uxr12` | die gespielten Runden legt die Kommandozeile an; das Fenster berichtigt und lost Runde 1; `uxr12` misst das Layout von Runden, Rangliste, Protokoll | teilweise |
| A08.5 Ein Turnier abschließen | `dom04ak03` | danach weist die Kommandozeile ein Anmelden ab, das Protokoll nennt den Abschluss | geprüft |
| A08.6 Ausgeben, Bericht sichern, wieder einlesen | `dom04ak03` | Dialog, Abbruch sagt nichts; Datei trägt Namen und Verlauf ohne Kennung des Bestands; Bericht nennt Kopf, Rangliste, Runde | geprüft |

**Gesamt:** 5 geprüft, 1 teilweise. Vorher: teilweise. Der Tiebreaker steht in der
Rangliste (OMW, GW, OGW); mehrere Runden **im Fenster** bis zum Ende sind es nicht.

### A09 — Zweite lokale Instanz, Einladung, Trennung

Auditstand: teilweise (`arbeitswege.json:80`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A09.1 Zwei Programme mit eigenem Bestand und eigener Identität verbinden | `nat04`, `nat04handel` | zwei verschiedene Identitäten, beide bestätigt, der Wirt horcht auf 127.0.0.1; **eingerichtet über die Kommandozeile, und der Horchposten über die Bindung `HorchStarten`**, nicht über die Bedienung | geprüft |
| A09.2 Einladen: den Code zeigen und eingeben, dreißig Ziffern vergleichen; eine Einladung genügt | `net01`, `net01b` | Anna sieht dieselben dreißig Ziffern wie Bert, »stimmt überein« gibt den Handel frei, dieselbe Einladung ein zweites Mal bietet keinen Vergleich an; bei Bert steht Anna unbestätigt unter ihrem Namen | geprüft |
| A09.3 Teilen: erst sehen, was mitgeht | `net04` | Adresse und Satz über die Kopie stehen vorher da, genau eine Kopie im Teilordner der Probe, nur für den Eigentümer; »Fertig« räumt sie weg (gewartet, bis sie fort ist). Der Teilen-Dialog von macOS wird gedrückt und mit Escape geschlossen — sein Erscheinen hat **kein ✓** im Beleg, nur ein Bild | geprüft |
| A09.4 Kontakte aus Apple Kontakte übernehmen | `id04` | Vorschau zeigt beide, übernommen wird die Auswahl; ein zweites Einlesen legt nichts doppelt an | geprüft |
| A09.5 Trennung und Wiederverbindung | `nat04` | der Gast steht auf, der Beitritt endet beim zuständigen Stop — **die Wiederverbindung hat kein Drehbuch** | teilweise |
| A09.6 Entdeckung im Netz (mDNS), Einladung über fremde Netze, zwei Geräte oder Netze | — | `SichtbarWerden` und `FernHorchStarten` ruft kein Drehbuch (sie stehen als Ausnahme neben der Quote); `HorchStarten` ruft `nat04` nur mit 127.0.0.1; beide Programme laufen auf einem Mac | nicht geprüft |

**Gesamt:** 4 geprüft, 1 teilweise, 1 nicht geprüft. Vorher: teilweise. **Alles
auf demselben Mac, über 127.0.0.1** — die Grenze, die das Audit nannte
(»zwei unabhängige NAT-Grenzen … fehlen«), ist unverändert.

### A10 — Testhandel und Zustandsfolge

Auditstand: teilweise (`arbeitswege.json:89`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A10.1 Die Liste »Was gerade läuft«: Entwurf, angeboten; ein Klick schlägt auf; den Entwurf löschen | `trade01` | der Entwurf »Du bist dran: anbieten«, der angebotene »Wartet auf …«; nach der Rückfrage bleibt der angebotene | geprüft |
| A10.2 Die Zustandsfolge in zwei Fenstern bis »beidseitig verbucht · abgeschlossen« | `nat04handel` | jeder Schritt (zustimmen, vorbereiten, übergeben, verbuchen, abgleichen, bestätigen) im Fenster dessen, der ihn geht, und beim anderen von selbst in der Liste; beide Seiten »abgeschlossen« | geprüft |
| A10.3 Je Zustand und Art nur Schritte, die gehen | `trade02`, `trade03` | Entwurf, angeboten, angenommen: 35 Knöpfe, 0 unzulässige Schrittknöpfe (gezählt, nicht bewertet); der Tausch steht als Vereinbarung da, die Deckleihe sagt, dass sie keinen Bestand bewegt, der Verlauf nennt jeden Wechsel | geprüft |
| A10.4 Genau eine Buchung, auch bei doppeltem Druck und Neustart | `trade03nat` | nach der Buchung ist »Verbuchen« fort, nach dem Neustart noch immer eine Bewegung und genau ein Wechsel | geprüft |
| A10.5 Abbruch, Pause und Fortsetzen eines laufenden Handels | `trade03`, `trade03nat` | geprüft ist, dass »abbrechen« angeboten wird, ein Abbruch ist kein ✓; Fortsetzen nur nach einem Neustart | teilweise |
| A10.6 Ein vollständiger eigener Mausweg über zwei Netze mit Netzwechsel | — | beide Fenster laufen auf einem Mac; die Einrichtung geht über die Kommandozeile | nicht geprüft |

**Gesamt:** 4 geprüft, 1 teilweise, 1 nicht geprüft. Vorher: teilweise; der
»abgeschlossene und beidseitig verbuchte Handel«, den die Inventur offen führte,
ist zu.

### A11 — Spieltisch: Karten, Tokens, Counter, Anhänge

Auditstand: teilweise (`arbeitswege.json:98`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A11.1 Der Tisch steht da und der Zug ist mit der Tastatur erreichbar | `fix-f04` | der Tisch lässt sich fokussieren, Zug »2/12 Anfangsphase · Versorgung« | geprüft |
| A11.2 Eine Karte mit dem Zeiger ziehen: Hand, Spielfeld, Friedhof; Abbruch | `a04` | Zeigerereignisse über AppKit; Hand → Spielfeld, Spielfeld → Friedhof; Escape und Loslassen außerhalb ändern keine Zone | geprüft |
| A11.3 Eine Karte zonengültig bewegen (Bibliothek mit Stelle, Hand ohne) | `fix-f04` | die Bibliothek bietet »oben, unten, zufällig« an, die Hand keine; Adas Hand 7 → 8 ohne Fehler | geprüft |
| A11.4 Den Tisch wieder und wieder betreten; die Bildzeit | `tischwiedereinstieg`, `abn01tisch`, `bildrate` | 2 × 10 Zyklen und 100 Besuche, jedes Mal neue Leinwand; Besuche 91–100 im Median 405 ms gegen 327 ms (2–11); Bildzeit Median 17,0 ms in Ruhe und in Bewegung | geprüft |
| A11.5 Karten tappen | — | kein Drehbuch | nicht geprüft |
| A11.6 Token, Counter, Anhänge; ein Abzug des Tisches als Bild und PDF | — | kein Drehbuch; der Vertragsrundgang ruft `PartieBefehl` nur mit einem Münzwurf | nicht geprüft |

**Gesamt:** 4 geprüft, 2 nicht geprüft. Vorher: teilweise. Die Handbewegung (F04),
die die Inventur offen führte, ist zu; offen ist alles, was über Ziehen und
Münzwurf hinausgeht.

### A12 — Gegnerperspektive und verdeckte Daten

Auditstand: teilweise (`arbeitswege.json:107`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A12.1 Jedes Fenster sieht die eigene Hand offen, die fremde verdeckt | `nat04` | Wirt: 7 offen, 7 verdeckt; Gast: 7 offen, 7 verdeckt | geprüft |
| A12.2 Eine Karte geht auf eine Hand, und das Gegenüber sieht sie nicht mit Namen | `fix-f04` | aus Bos Sicht liegt Adas Hand verdeckt da (7 offen, 8 verdeckt), keine Karte mit Namen | geprüft |
| A12.3 Der Wirt sieht, was der Gast tut, ohne selbst etwas zu tun | `nat04` | Münzwurf und Satz des Gastes erscheinen beim Wirt, keine Fehlermeldung | geprüft |
| A12.4 Verdeckte Daten über zwei Geräte oder Netze | — | beide Fenster auf einem Mac | nicht geprüft |

**Gesamt:** 3 geprüft, 1 nicht geprüft. Vorher: teilweise. Der Bewegungsfehler, der
im Audit die Handkarte in der Gegnersicht verhinderte, ist im Schlusslauf nicht
mehr da.

### A13 — Nutzer/Profile/Nur-Lesen/Rechteentzug

Auditstand: teilweise (`arbeitswege.json:116`; die Quelle von 0.5.0 sagte
»fehlgeschlagen in mehreren Schreibwegen«).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A13.1 Das Profil wechseln, mit zweitem Faktor | `profilwechsel` | ohne Code fragt das Fenster nach ihm, ein falscher Code wird abgewiesen, der richtige wechselt in Bobs Profil »nur lesen« | geprüft |
| A13.2 Die lokale Sitzung: Kennwort setzen, sperren, entsperren, abmelden | `id02` | nach ⌘Q und Neustart bleibt es gesperrt; das falsche Kennwort bleibt gesperrt, das richtige entsperrt | geprüft |
| A13.3 Anmeldeanbieter nennen ihren Prüfstand | `id05` | alle vier »nur gegen eine Nachbildung geprüft«, Microsoft nach dem Eintragen »eingerichtet« | geprüft |
| A13.4 Das Fenster nennt die Rolle »Leser« | `zustaende`, `profilwechsel` | »nur lesen: das Fenster nennt die Leserrolle — Leser« (`zustaende.md:141`); im Profil wechselt der Kopf zu »Mein Profil von Bob · nur lesen« | geprüft |
| A13.5 Als Leser lassen sich auf keiner Seite Schreibwege bedienen | `zustaende` | **16 von 25 Seiten zeigen aktive Schreibknöpfe**, 9 keinen (`zustaende.md:311–335`, in beiden Läufen gleich); der Knopf wurde nicht gedrückt; G10, #2276 | rot |
| A13.6 Ein Recht wird entzogen, während das Fenster offen ist | — | kein Drehbuch | nicht geprüft |

**Gesamt:** 4 geprüft, 1 rot, 1 nicht geprüft. Vorher: teilweise. Die Zelle
belegt, dass der Knopf **aktiv** ist; ob Go den Ruf abweist, hat kein Drehbuch
ausgelöst. Rot ist die Oberfläche, nicht das Backend.

### A14 — Einstellungen, KI, CLI, Diagnose

Auditstand: teilweise (`arbeitswege.json:125`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A14.1 Einstellungen setzen und im Bestand sehen; Bereiche abschalten | `ops02`, `uxrf1`, `uxrf2`, `ops10` | ein Ordner, den es nicht gibt, wird mit einem Satz abgewiesen; ⏎ schreibt; Schalter per Leertaste, Segmente per Pfeil; »Spielen« aus: die übrigen Bereiche stehen weiter da, ⌘4 führt zu einer Seite mit dem Weg zurück, die Kommandozeile weist den Bereich ab | geprüft |
| A14.2 Diagnose: Protokoll, Laufzeiten, Leistungsprofil, Diagnosepaket, Zwischenspeicher | `ops04`, `ops05`, `ops08` | Datei im eingestellten Ordner; Paket ohne den abgewählten Teil (sechs Dateien); von zwei Leistungsprofilen bleibt eines; acht Zwischenspeicher mit Ort und Grenze, »Alle leeren« steht da und wurde nicht gedrückt | geprüft |
| A14.3 Ein KI-Vorschlag vom Gerätemodell | `ai05` | Knopf am Feld Name, »von apple/system · Vorrang«, übernommen; der Sitzungsschalter aus beendet `fm serve`, ⌘Q beendet es auch | geprüft |
| A14.4 Ein KI-Vorschlag von einem lokalen Ollama-Modell | `ai02` | Vorbedingung scheitert: Ollama antwortet nicht unter 127.0.0.1:11434 (`ECONNREFUSED`, `ai02.md:11`); die Modelle sind gelöscht (Commitnachricht von `256cebec9`) — **kein Produktbefund**, aber auch kein Beleg | rot |
| A14.5 Eine KI-Anfrage an einen Anbieter in der Wolke | `ops07` | der Prüfstand nennt »Anthropic nur vorhanden«; die KI-Bindungen sind vom Rundgang ausgenommen | nicht geprüft |
| A14.6 Die Kommandozeile und den Werkzeugserver einrichten und widerrufen | `ops09b`, `phase64` | Claude Desktop: Einrichten, der fremde Eintrag bleibt, Widerrufen gibt die Datei Byte für Byte zurück; Claude Code, der Dienst (LaunchAgent) und der Kopierknopf des Prüfbefehls wurden nicht gedrückt | teilweise |

**Gesamt:** 3 geprüft, 1 teilweise, 1 rot, 1 nicht geprüft. Vorher: teilweise;
zu ist das erzeugte Diagnosepaket. `ai02` ist rot, weil der Lauf rot sein **muss**,
wenn seine Vorbedingung fehlt (`ai02.mjs`, Kopf).

### A15 — Testbestand sichern und restaurieren

Auditstand: **fehlgeschlagen** (`arbeitswege.json:134`; F-01: nach dem
Zurückspielen blieb die Anwendung gesperrt).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A15.1 Den Bestand sichern: Dialog, Vorschlag, Datei | `eintraege`, `ops02` | »Bestand sichern …« öffnet den Dialog (Vorschlag »mse-sicherung.msesicherung«) und schreibt 274016 Bytes; der Sicherungsordner wird gesetzt oder abgewiesen | geprüft |
| A15.2 Zurückspielen und ohne Neustart weiterarbeiten | `a02`, `ausgangswerte` | nach dem Zurückspielen: Set öffnen, Karte anlegen, Diagnose lesen — ohne unerwarteten Fehler; Wiederherstellung 5 Handlungen, 0,5 s (`ausgangswerte.md:29`) | geprüft |
| A15.3 Die Verschlüsselung ein- und ausschalten und danach schreiben | `a02` | »Nach dem Verschlüsseln« und »Nach dem Entschlüsseln« angelegt, ohne unerwarteten Fehler | geprüft |
| A15.4 Eine Sicherung mit Kennwort | `ops06` | zwei verschiedene Kennwörter halten den Knopf an, zwei gleiche geben ihn frei, der Dialog öffnet (»Bestand sichern«) und wird **abgebrochen**; keine Datei, kein Zurückspielen | teilweise |
| A15.5 Ein unbrauchbarer Bestand: der Start zeigt den Weg zur Sicherung | `kernwege` | »Der Bestand ließ sich nicht öffnen«, nicht still | geprüft |
| A15.6 Alle Objektarten kommen identisch zurück; eine unterbrochene Wiederherstellung | — | kein Drehbuch; `arbeitswege.json:135` nennt »Alle Objektarten offen« | nicht geprüft |

**Gesamt:** 4 geprüft, 1 teilweise, 1 nicht geprüft. Von **fehlgeschlagen auf
teilweise**: kein Schritt des Weges ist im Schlusslauf mehr rot. `a02` hält F-01
(Phase A, #2216).

### A16 — Hilfe, Lizenz, Datenschutz, Sprachinfo

Auditstand: teilweise (`arbeitswege.json:143`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A16.1 Die Hilfe erreichen; »Über« nennt Lizenz und Inhalte Dritter | `eintraege` | »Hilfe › MagicSetEditor4Mac-Hilfe« führt auf »Hilfe«; »Über« rollt auf den Abschnitt mit Lizenz und Rechten Dritter | geprüft |
| A16.2 Fassung, Datenstand, Signatur, Aktualisierung und Orte stehen da; der Prüfstand trennt Geprüftes von Vorhandenem | `ops07` | »ad-hoc und nicht notarisiert«; Scryfall geprüft mit Beleg, Anthropic nur vorhanden, VoiceOver wartet auf einen Menschen | geprüft |
| A16.3 Der Datenschutztext nennt Wege, die es gibt | `ops07` | kein Verweis auf eine erfundene Gruppe | geprüft |
| A16.4 Das Netzverhalten gegen das, was die Texte sagen | `sec17`, `sec17bilder`, `svgbild` | einzelne Wege gemessen: die Schutzrichtlinie weist Inline-Skript, `eval` und fremdes Skript ab (3 Verstöße, keine anderen); fremde Vorschaubilder laden; die SVG-Ansicht ruft nichts. **Der Abgleich Text gegen Netzausgang ist keiner dieser Läufe** | teilweise |
| A16.5 Die Sprachinfo in den Einstellungen | `loc01` | Japanisch gilt nach dem Neustart; eine eigene Übersetzung steht mit ihren Zahlen da | geprüft |

**Gesamt:** 4 geprüft, 1 teilweise. Vorher: teilweise. Die Frage des Audits —
»das tatsächliche Netzverhalten gegen das, was die Texte sagen« — ist nur an
einzelnen Wegen beantwortet.

### A17 — Fenstergrößen, Sprachen, Schrift, Bewegung

Auditstand: teilweise (`arbeitswege.json:152`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A17.1 Das Fenster startet maximiert, »wie zuletzt«, eine zu große Größe wird geklemmt | `start` | 1470 × 838 in der Arbeitsfläche 1470 × 841, nicht im Vollbild; 1100 × 760 kommt wieder; zu große Größe passt auf den Bildschirm | geprüft |
| A17.2 Vollbild hinein und heraus | `vollbild` | **Schlusslauf rot** (macOS hängte keinen Vollbild-Eintrag an, `git show 256cebec9:docs/belege/huellenprobe/vollbild.md`); Einzellauf am Commit `256cebec9` grün: der Eintrag ändert sich und heißt danach wie vorher, Größe 1470 × 838 gleich | geprüft |
| A17.3 Layout bei schmalem Fenster, großer Schrift und weniger Bewegung | `phase611`, `p03`, `politur`, `uirevue` | Leiste oben, »Sehr groß«, Schrift »Buch«, 1.024 × 700: kein Knopf abgeschnitten, nichts rollt waagerecht; 900 × 700 in Deutsch hell und Englisch dunkel; der Übergang schrumpft von 0.2 s auf 0.00001 s | geprüft |
| A17.4 Die Mindestgröße, die das Fenster zulässt | — | kleinste gemessene Größe 900 × 700 (`p03`); die Grenze des Fensters selbst ist nicht gemessen (H02) | nicht geprüft |
| A17.5 Acht Sprachen | `leistenfuss`, `loc01`, `g18`, `leisteenglisch` | der Fuß der Seitenleiste steht in allen acht ganz da (de en es fr it ja ru zh); Japanisch ganz in Leiste und Seite; Glyphen de/zh/ja; englisches System ohne deutschen Titel — **die Seiten in es, fr, it, ru, zh nicht** | teilweise |
| A17.6 Kontrast hell und dunkel, auch im älteren System | `a03`, `n13`, `politur`, `a03-gast-15-2` | Wirt: kleinster Text 8,04 : 1 hell, 6,99 : 1 dunkel (`a03.md`); Gast macOS 15.2: 1,27 : 1 ohne die Regeln, 6,99 : 1 mit ihnen — **DOM-Messung ohne Bild, eigener Lauf, nicht Teil des Schlusslaufs** | geprüft |

**Gesamt:** 4 geprüft, 1 teilweise, 1 nicht geprüft. Vorher: teilweise. F-03 (der
schwarze Titel im Gast) ist am Wirt und im Gast gemessen und behoben.

### A18 — Tastatur/Fokus/Navigation

Auditstand: teilweise (`arbeitswege.json:161`).

| Teilaufgabe | Drehbuch (Beleg) | Ergebnis im Schlusslauf | Stand |
| --- | --- | --- | --- |
| A18.1 Eine Seite über die Befehlspalette erreichen: ⌘K, tippen, den Treffer anklicken | `inventur`, `budgets`, `uirevue` | alle 25 Seiten öffnen sich ohne Fehlermeldung und ohne waagerechtes Rollen; Seitenwechsel Median 49 ms (`budgets.md:48`) | geprüft |
| A18.2 Die Befehlspalette allein mit der Tastatur bedienen (Pfeile, Return) | — | jedes Drehbuch klickt die Zeile; der Audit nannte »Return ohne Auswahl navigiert nicht« (`docs/inventur.json`, A18) | nicht geprüft |
| A18.3 Menüeinträge und Kürzel wirken | `leiste`, `eintraege`, `darstellung` | Kürzel und Rollen der Einträge stimmen; Zurück/Vorwärts, Seitenleiste ⌥⌘S; Navigator ⌥⌘[ und Inspector ⌥⌘] klappen genau einmal | geprüft |
| A18.4 Tasten in Auswahl, Schalter, Segmenten und Liste | `uxrf1`, `uxrf2`, `uxrf3` | die Leertaste legt den Schalter um und der Fokus bleibt; Pfeile wählen in den Segmenten; ↓ und ↑ nehmen den Fokus mit | geprüft |
| A18.5 Der Fokus bei Überlagerungen und Menüs | `phase611`, `a04`, `p03` | dreißig Mal Tab bleibt in der Schublade; ⇧F10 setzt den Fokus ins Kontextmenü; »Mehr« gibt den Fokus zurück | geprüft |
| A18.6 Ein vollständiger Fokuszyklus durch eine Seite; VoiceOver spricht | — | `ops07`: »VoiceOver wartet auf einen Menschen« | nicht geprüft |

**Gesamt:** 4 geprüft, 2 nicht geprüft. Vorher: teilweise.

## Die 150 Zellen (Querverweis)

Die 25 Seiten in sechs Zuständen stehen in `zustaende`
(`docs/belege/huellenprobe/zustaende.md`, 150 Zellen, 146 Bilder, davon 4 Zellen
»nicht anwendbar« mit Grund). Sie speisen A13 und die Hürden unten. **Der Beleg
ist nicht der des Schlusslaufs**: Commit `5d34989ff`, Lauf
`2026-10-10T02-36-43Z-88701`, Bauabdruck `f6301743d8d5` — die Probe in der
Fassung nach Zweitblick und Nachlauf (Ruhe im Rufbuch, Set gilt erst mit
einer Setseite als offen, jede gestörte Zelle weist ihren gestörten Ruf
nach; keine Zelle blieb »nicht erreicht«). Der Schlusslauf hatte die Probe in einer Fassung, die der Zweitblick
über Phase B beanstandete — angekommen war, wer die Palette schloss; beobachtet
wurde auch im Meldungsband und in der Leiste; die Bindungen der Palette zählten
zu jeder Seite; feste Pausen statt Wirkung; und das Urteil »offline« stand für
jede Seite fest auf »steht« (`9cdd66a39`). Die Zahlen unten sind die der neuen
Fassung. Gegenüber dem Schlusslauf unterscheiden sich nur zwei Zustände: bei
»laden« sind aus 8 Seiten ohne Ladehinweis 6 geworden (`hilfe` und
`einstellungen` stehen jetzt »nicht anwendbar«), bei »Fehler« aus 5 Seiten ohne
Meldung 3 (dieselben zwei); die übrigen vier Zustände sind gleich.

| Zustand | Urteile | Zeilen |
| --- | --- | --- |
| befüllt | 25 × »steht«, davon 13 mit dem Zusatz »leer« | `zustaende.md:186–210` |
| laden | 17 × »sagt, dass sie lädt«, 6 × »ohne Ladehinweis« (start, sammlung, suche, ersetzen, bilder, ausgeben), 2 × »nicht anwendbar« (hilfe, einstellungen) | `zustaende.md:211–235` |
| Fehler | 20 × »meldet«, 3 × »ohne Meldung« (suche, ersetzen, ausgeben), 2 × »nicht anwendbar« (hilfe, einstellungen) | `zustaende.md:236–260` |
| offline | 25 × »steht«, jede mit »keine Meldung« | `zustaende.md:261–285` |
| leer | 19 × »Leerzustand« — darunter die 13 Setseiten, die ein früherer Lauf ungemessen »nicht anwendbar« nannte —, 6 × »ohne Leerzustand« (start, pakete, vorlagen, nutzer, hilfe, einstellungen) | `zustaende.md:286–310` |
| nur lesen | 16 × »Schreibknöpfe aktiv«, 9 × »ohne aktiven Schreibknopf« | `zustaende.md:311–335` |

**Was die Zellen nicht tragen**, aus der Beleglage:

- **Die Störung antwortet mit dem Code `unbekannt`**
  (`internal/fenster/rufstoerung.go:34`). Die 20 Meldungen »Es ist ein
  unerwarteter Fehler aufgetreten …« sind der Rückfalltext dazu; sie belegen,
  **dass** eine Meldung kommt, nicht, was echte Fehler sagen.
- **Der Saatbestand hat kein Exemplar, kein Turnier, keinen Tausch**
  (`scripts/huellenprobe/zustaende.mjs` Z. 170–179; der Kopf nennt ein Exemplar,
  Z. 11, der Aufbau legt keines an). Die Zellen »befüllt« für `sammlung`,
  `turniere` und `handeln` sind deshalb leer, und 13 »befüllt«-Zellen tragen ein
  »leer« — bei Urteil »steht«.
- **Die Spalte »wie« unterscheidet acht Setseiten nicht**: `karten`, `setangaben`,
  `schluesselwoerter`, `konsole`, `pack`, `symbol`, `stil`, `statistik` zeigen
  alle »Langtext«, den Kartennamen (`zustaende.md:186–210`).
- **Ein früherer Widerspruch ist behoben**: `zustaende` führte die 13
  Setseiten im Zustand »leer« ungemessen als »nicht anwendbar — ohne Set gibt
  es keine Setseite«, `uxr23` erreichte sie und maß je einen Leerzustand.
  Seit `8bac6e808` misst auch `zustaende` sie, und alle 13 zeigen einen
  (`zustaende.md:298–310`). **Offen bleibt das Maß**: `uxr23` zählt
  `data-leerzustand`, `zustaende` urteilt auch über einen Satz im Text (G02,
  #2249).
- **»nur lesen« misst, ob ein Knopf aktiv ist**, nicht, ob der Schreibruf
  gelingt (`zustaende.mjs`, die Zeile mit `aktiv`); der Befund G10
  (#2276)
  steht in `docs/CHRONIK.md` (B03).
- **Die Bilder liegen unter `.huellenprobe/`**, das nicht im Git steht
  (`.gitignore:119`); auf diesem Mac sind alle 133 vorhanden.

## Vier Personas

Die Namen und Wege kommen aus G01
(`docs/audits/0.8.0/ergebnis/PLAN-0.8.0.md:446–452`); die Perspektiven aus
`docs/audits/0.8.0/ergebnis/AUDIT.md:22–25`. **Es gibt keine menschliche Erfolgsquote**: das Audit ist, in
seinen Worten, »rollenbasierte Experteninspektion; kein Test mit vier realen
Nutzergruppen« (G01, Anlass). Klickzahlen stehen nur dort, wo ein Beleg sie
nennt, sonst steht »nicht gemessen«.

**Die einzigen Zahlen zu Handlungen** sind die der Probe `ausgangswerte`. Eine
Handlung ist, was ein Mensch einmal tut (Klick, Kürzel, getippter Wert samt ⇥,
Wahl in einer Liste); die Dauer ist **Maschinenzeit und keine Bearbeitungszeit
eines Menschen**, Fehlversuche hat die Probe keine, weil sie den Weg kennt
(`ausgangswerte.mjs`, Kopf). Vorher ist der Freigabestand 0.7.0
(`367bc8b4c`, 2026-10-08), nachher der Schlusslauf:

| Kernweg | Handlungen vorher | Handlungen jetzt | Maschinenzeit vorher | Maschinenzeit jetzt |
| --- | ---: | ---: | ---: | ---: |
| Kartenanlage | 4 | 4 | 0,7 s | 0,8 s |
| Import (MTG JSON) | 4 | 4 | 0,2 s | 0,4 s |
| Deckbau | 10 | 10 | 1,3 s | 1,2 s |
| Spielstart | 9 | 9 | 1,4 s | 1,5 s |
| Wiederherstellung | 5 | 5 | 0,4 s | 0,5 s |

Quelle »vorher«: `git show 367bc8b4c:docs/belege/huellenprobe/ausgangswerte.md`;
»jetzt«: `ausgangswerte.md:11–29`. **Die Handlungen sind in allen fünf Wegen
gleich geblieben**; die Zeiten liegen im Rauschen eines Laufs. Die Audit-Zeit für
einen warmen Start (Median 2131 ms, `docs/audits/0.8.0/ergebnis/AUDIT.md`, Zeile »warme Prozessstarts«) und die von B02 (514 ms) sind
**nicht vergleichbar**: anderes Endereignis und eine halbe Sekunde Pause weniger
(`docs/LEISTUNGSSCHNITTE.md` §9).

**Die Budgets aus B02** (`budgets.md:45–49`, Lauf vom 2026-10-09, Last vor dem
Lauf 6,75 auf zehn Kernen): Start kalt Median 500 ms (n = 5, max. 514), warm 514 ms
(n = 6), Feld bis dekodiertes Vorschaubild Median 27 ms (n = 24, max. 97, Ziel
150), Seitenwechsel Median 49 ms (n = 24), Rollbild in »Black« Median 22 ms,
**P95 55 ms gegen höchstens 50** (n = 159; Folgepaket E07,
#2279).

### Entwickler — Diagnose, API, Build

| Schritt | Drehbuch (Beleg) | Ergebnis | Stand |
| --- | --- | --- | --- |
| Diagnose einstellen und ein Paket erzeugen | `ops04` | Protokoll am eingestellten Ort; Kanäle und Stufe stehen in der Diagnose; im Paket fehlt der abgewählte Teil | geprüft |
| Betriebszahlen lesen, Leistungsprofil ziehen, Zwischenspeicher ansehen | `ops05`, `ops08` | Arbeitsspeicher mit Einheit, Verbindungen, Zeitraum und Aufwand; was nicht gemessen ist, steht da; acht Zwischenspeicher | geprüft |
| Jede Bindung am Fenster rufen, die Antwortform prüfen | `vertrag` | 454 von 495 gezählten Bindungen gelungen, dazu 37 Ausnahmen neben der Quote (`bindungsreichweite.md:10`, Lauf 2026-10-09); lesende Bindungen mit geprüfter Antwortform: 171 von 239, 71,5 % gegen verlangte 90 % (`vertrag.md:109`); eine Antwort weicht bei fachlicher Ablehnung ab, `Fremdpreisverlauf` gibt `befunde: null` statt einer Liste (`vertrag.md:108`) | rot |
| Werkzeugserver und Kommandozeile einrichten | `ops09b`, `phase64` | Claude Desktop einrichten und widerrufen; Claude Code, Dienst und Kopierknopf nicht gedrückt | teilweise |
| `mse`, REST und MCP als Gegenstellen des Fensters | `dom03ak01`, `dom03ak02`, `ux12`, `ops10` | die Kommandozeile liest am selben Bestand dasselbe, was das Fenster schrieb, und weist abgeschaltete Bereiche ab; REST und MCP als Dienst fährt am Fenster kein Drehbuch | teilweise |
| Das Bündel bauen und starten | — | die Belege tragen Bauabdruck und Commit (`1b6e1cccf72a`); der Bau selbst ist kein Drehbuch, und die vollständige Pipeline war im Audit rot (`docs/audits/0.8.0/ergebnis/AUDIT.md:49`, F-04) — ihr Stand am Schlussstand ist hier **nicht erhoben** | nicht geprüft |
| Zwei Stunden Betrieb | `dauerbetrieb` | 125 Minuten, 1.712 Runden, dieselbe Kennung, nichts verloren; das Speicherziel (höchstens 10 % je Stunde, kein stetiger Anstieg) ist **nicht gehalten**. Der Beleg trägt noch die erste Auswertung (+10,6 %, `dauerbetrieb.md:14`); `docs/LEISTUNGSSCHNITTE.md` §9.1 nennt sie einen Rechenfehler und rechnet über ganze Abschnitte nach: −0,3 % vom ersten zum letzten ganzen Abschnitt, aber **+25 % von Minute 40 bis 110**, stetig über vier Abschnitte | rot |

**Hürden** (die fünf größten, jede mit Beleg):

1. **Die Antwortform ist bei 28,5 % der lesenden Bindungen nicht geprüft**
   (`vertrag.md:109`): wer die API baut, verlässt sich dort auf die Gestalt, die
   ihm die Brücke gerade gibt. Phase C (C01).
2. **Die Bedienung erreicht nur gut die Hälfte der Bindungen**: ohne den
   Vertragsrundgang 254 von 495 (51,3 %), schreibend 104 von 268 (38,8 %)
   (`docs/belege/bindungsreichweite.md:15–21`, Lauf vom 2026-10-09); der Rest
   ist nur über den Rundgang erreicht, der keine Bedienung ist.
3. **37 Bindungen stehen mit Grund neben der Quote** — Scryfall, KI, Netzfreigabe,
   LaunchAgent, Aktualisierung (`bindungsreichweite.md`, Ausnahmen) —, und zwei
   Namen wurden am Tor abgewiesen (`Datenstand` 4, `Einstellungen` 1,
   `bindungsreichweite.md:24`).
4. **Der Speicher steigt im Dauerbetrieb stetig**: nachgerechnet +25 % von
   Minute 40 bis 110 (`docs/LEISTUNGSSCHNITTE.md` §9.1; der Beleg nennt noch
   +10,6 %); was bleibt, liegt im Webprozess (E06, #2278).
5. **Die Pipeline ist als Beleg nicht erhoben**: F-04 stand im Audit rot; ob sie
   heute grün ist, sagt dieser Beleg nicht — und `rufbuch.go` hat sich seit der
   Reichweitenmessung geändert (`75c90a394`), die Zahlen oben gelten für den
   Stand davor.

Klickzahlen für den Weg: **nicht gemessen**.

### Admin — Rolle, Sicherung, Update

| Schritt | Drehbuch (Beleg) | Ergebnis | Stand |
| --- | --- | --- | --- |
| Das Profil wechseln, mit zweitem Faktor | `profilwechsel` | siehe A13.1 | geprüft |
| Eine Rolle »Leser« (über die Kommandozeile vergeben) am Fenster wirken sehen | `zustaende` | die Leserrolle steht in der Leiste; 16 von 25 Seiten zeigen aktive Schreibknöpfe | rot |
| Die Sitzung sperren und entsperren | `id02` | siehe A13.2 | geprüft |
| Bereiche abschalten, der Weg zurück | `ops10` | siehe A14.1 | geprüft |
| Den Bestand sichern, den Ordner setzen | `eintraege`, `ops02` | Dialog, 274016 Bytes; ein abgezogener Ordner wird mit dem Laufwerk als Grund abgewiesen | geprüft |
| Eine Sicherung mit Kennwort | `ops06` | Felder, Hinweis ohne Rückweg, Dialog geöffnet und abgebrochen | teilweise |
| Zurückspielen und weiterarbeiten | `a02`, `ausgangswerte` | 5 Handlungen, 0,5 s; danach schreibt die Anwendung ohne Neustart | geprüft |
| Ein unbrauchbarer Bestand | `kernwege` | der Start nennt den Weg zur Sicherung | geprüft |
| Update: ein neuer Paketstand mit Plan | `dom01` | erst der Plan (Fassungen, betroffenes Set, Quelle), dann das Ersetzen; »Vorige Fassung zurückholen« und »Nichts ändern« | geprüft |
| Update: eine neue Fassung der Anwendung sehen und laden | — | `NachNeuerFassungSehen` und `NeueFassungLaden` sind ausgenommen (I03, #2261); `ops07` zeigt nur, dass die Aktualisierung in der Hilfe steht | nicht geprüft |

**Hürden:**

1. **Die Rolle wirkt an der Oberfläche nicht auf 16 von 25 Seiten**: ein Leser sieht
   aktive Schreibknöpfe (`zustaende.md:311–335`, G10, #2276). Ob Go den Ruf
   abweist, ist ungeprüft.
2. **Wer das Kennwort einer Sicherung vergisst, hat keinen Rückweg** — der
   Hinweis sagt es (`ops06.md`) —, und zurückgespielt wird eine Sicherung mit
   Kennwort am Fenster in keinem Drehbuch.
3. **Wiederherstellen ist an einem identischen Teststand gefahren**, nicht an allen
   Objektarten und nicht unterbrochen (`arbeitswege.json:135`).
4. **Der Update der Anwendung ist nicht gefahren**: kein Drehbuch ruft
   `NachNeuerFassungSehen` oder `NeueFassungLaden` (Suche in
   `scripts/huellenprobe/`); beide stehen neben der Quote.
5. **Der Webprozess wächst ab Minute 40 stetig** (+25 % bis Minute 110): ein Betrieb
   über Tage braucht einen Neustart, den kein Beleg beziffert. Ohne Absturz und
   ohne verlorene Änderung in 125 Minuten (`dauerbetrieb.md:11–13`).

Ausgangszahlen: Wiederherstellung **5 Handlungen, 0,5 s Maschinenzeit**
(vorher 5, 0,4 s); Start kalt Median 500 ms.

### Erfahrener Fan — Import, Deck, Netz

| Schritt | Drehbuch (Beleg) | Ergebnis | Stand |
| --- | --- | --- | --- |
| Ein Set aus Datei, Finder oder Verzeichnis einlesen; fehlt ein Paket, steht die Vorschau da | `fmt03a`, `kernwege`, `eintraege`, `ausgangswerte` | Import 4 Handlungen, 0,4 s; ein Verzeichnis MTG JSON bringt Sets von 1 auf 3 Zeilen; die Vorschau nennt das fehlende Spiel | geprüft |
| Eine Sammlung aus CSV, Dubletten, Rückfrage | `fix-f03`, `dom03`, `dom03ak01` | siehe A05.1 bis A05.4 | geprüft |
| Eine Namensliste einfügen; Scryfall | — | kein Drehbuch | nicht geprüft |
| Ein Deck bauen, kopieren, Fehlkarten auf die Wunschliste | `ausgangswerte`, `dom03ak02`, `ux12` | Deckbau 10 Handlungen, 1,2 s | geprüft |
| Eine Partie anfangen und Karten ziehen | `ausgangswerte`, `a04`, `fix-f04` | Spielstart 9 Handlungen, 1,5 s; Karte mit dem Zeiger Hand → Spielfeld → Friedhof | geprüft |
| Einladen und den Code vergleichen | `net01`, `net01b` | dieselben dreißig Ziffern auf beiden Seiten | geprüft |
| Zwei Fenster, ein Tisch | `nat04` | verdeckte Hand, Münzwurf und Satz beim Wirt — **auf einem Mac** | geprüft |
| Einen Tausch bis beidseitig verbucht | `nat04handel`, `trade03nat` | beide Seiten »abgeschlossen«, genau eine Buchung — **auf einem Mac** | geprüft |
| Zwei Geräte, zwei Netze, mDNS, Fernweg | — | ausgenommen | nicht geprüft |

**Hürden:**

1. **Das Netz ist zwischen zwei Programmen auf einem Mac über 127.0.0.1 gefahren**
   (`nat04.md`, »der Wirt horcht — 127.0.0.1«); zwei Geräte oder NATs gibt es
   nicht.
2. **Entdeckung und Fernweg fehlen**: `SichtbarWerden` und
   `FernHorchStarten` ruft kein Drehbuch; `HorchStarten` rufen `nat04` und
   `nat04handel` nur mit 127.0.0.1 und über die Bindung (macOS fragt sonst nach der
   Netzerlaubnis).
3. **Eine Namensliste und Scryfall haben kein Drehbuch**; der Import, der am
   Fenster läuft, ist MTG JSON und CSV.
4. **Tappen, Token, Counter, Anhänge und der Abzug des Tisches** sind nicht gefahren
   (A11.5, A11.6).
5. **Der Deckbau ist der längste der fünf gemessenen Wege** (10 Handlungen, vor
   dem Spielstart mit 9), und ein Deck aus Exemplaren der Sammlung baut die
   Oberfläche in keinem Drehbuch (A07.2).

Dazu aus dem Zustandsbeleg: die Seite »Sammlung« zeigt bei einer um 4 s
verzögerten `Portfolio`-Antwort keinen Ladehinweis (`zustaende.md:215`) — bei einer
Sammlung ohne Exemplare im Saatbestand, also ohne die Last, die ein Mensch mit
vielen Karten hätte.

### Fan ohne IT — Erstkarte, erste Liste, Export

| Schritt | Drehbuch (Beleg) | Ergebnis | Stand |
| --- | --- | --- | --- |
| Die ersten Schritte sehen | `a03`, `start` | die ersten Schritte: 13 Texte gemessen, kleinstes Verhältnis 8,04 : 1 (hell) und 6,99 : 1 (dunkel); das Fenster startet maximiert — **ob sie verständlich sind, misst kein Drehbuch** | teilweise |
| Ein neues Set anlegen, den Stil wählen | `uxr04` | »Neues Set« führt zum Namen, der Stil ist eine Auswahl mit Suche | geprüft |
| Die erste Karte: Name, Text, Bild | `ausgangswerte`, `fix-f02`, `fp02` | Kartenanlage 4 Handlungen, 0,8 s; Bild, Text und Symbol im Set | geprüft |
| Die Vorschau folgt | `budgets` | Feld bis Vorschaubild Median 27 ms, 0 ohne Bild | geprüft |
| Die erste Liste: eine Namensliste einfügen und aufnehmen | — | kein Drehbuch; der Leerzustand der Sammlung steht da (`zustaende.md:290`, `uxr23.md`) | nicht geprüft |
| Den Export: Set sichern, Druckbogen, PDF | `fix-f02`, `uxrf7`, `fmt03b`, `a04` | Dialog und Datei; »Wie es wird: 9 je Seite · 3 Seiten« vor dem Schreiben; ⌘P öffnet »Druckbogen sichern« — **kein Drucker** | geprüft |
| Zurück, wenn etwas schiefging | `eintraege`, `leiste` | »Rückgängig« nimmt den Kartennamen zurück, ⌘Z das Feld; der Zyklus über Fokuswechsel ist ungefahren (A03.3) | teilweise |

**Hürden:**

1. **Im Fehlerfall zeigt die Oberfläche den Rückfalltext** »Es ist ein
   unerwarteter Fehler aufgetreten. Was genau, steht im Protokoll« (20 Seiten,
   `zustaende.md:236–260`) — für einen Menschen ohne IT ein Verweis auf ein
   Werkzeug, das er nicht kennt. Der Text kommt vom Code `unbekannt` der Störung;
   echte Fehler tragen eigene Sätze, **und welche, belegt dieser Lauf nicht**.
2. **Sechs Seiten tragen im frischen Bestand keine Leerzustands-Marke** —
   die 13 Setseiten ohne Set zeigen inzwischen in beiden Drehbüchern einen
   Leerzustand (`start`, `pakete`, `vorlagen`, `nutzer`, `hilfe`, `einstellungen`,
   `zustaende.md:286–297`). Ob das ein Mangel ist, entscheidet G02.
3. **Sechs Seiten ohne Ladehinweis** bei langsamer Antwort (`start`, `sammlung`,
   `suche`, `ersetzen`, `bilder`, `ausgeben`, `zustaende.md:211–235`), und drei
   ohne Meldung bei einem Fehler (`suche`, `ersetzen`, `ausgeben`,
   `zustaende.md:236–260`).
4. **Die erste Liste ist ungefahren**: weder die Namensliste noch Scryfall hat ein
   Drehbuch; was die Seite Sammlung zeigt, ist der Leerzustand.
5. **Englische Vorlagenfelder und Hinweise trotz deutscher Oberfläche** nannte das
   Audit (`docs/audits/0.8.0/ergebnis/AUDIT.md:22`); im Schlusslauf fährt kein
   Drehbuch das nach.

Ausgangszahlen: Kartenanlage **4 Handlungen** im offenen Set (vorher ebenfalls 4);
**von der leeren Anwendung zur ersten Karte** nannte `docs/UX-ARCHITEKTUR.md:634`
sieben Handlungen (2026-09-03, gegen die Saat der Auslieferung) — **im Schlusslauf
nicht neu gemessen**.

## Was dieser Beleg nicht sagt

- **Nichts über Menschen.** Keine Erfolgsquote, keine Zeit eines Menschen, keine
  Verständlichkeit: die Drehbücher bewegen eine Maschine, die den Weg kennt.
  »Geprüft« heißt, dass ein Weg **gegangen werden kann**, nicht, dass jemand ihn
  findet.
- **Nichts über den Stand nach dem Lauf.** Sechzehn Commits liegen über
  `1c0b74221`; das Menü, das Rufbuch und der Paketkatalog haben sich geändert,
  die Zustandsprobe und die Auswertung des Dauerbetriebs ebenso. Ein neuer Lauf
  kann Stände verschieben — in beide Richtungen.
- **Nichts über die 150 Zellen als Urteil.** Sie sind die Ausgabe einer Probe,
  die an künstlichen Störungen misst (Code `unbekannt`, Saat ohne Exemplar,
  Turnier und Tausch); sie zeigen, **wo man hinsehen muss**, nicht was falsch ist.
- **Nichts über einen durchgehenden Weg.** Die Teilaufgaben eines Weges stammen
  aus verschiedenen Drehbüchern mit je eigenem Zuhause; dass sie hintereinander
  gehen, ist nicht gefahren.
- **Nichts über ein zweites Gerät, ein zweites Netz, macOS 13 oder 14.** Alles
  auf einem Mac (Mac17,3, macOS 27.0.1); der Gast unter macOS 15.2 hat nur die
  Kontrastmessung zu A17.6 gesehen.
- **Nichts über echte Bestände.** Die Daten sind künstlich; »Black« wurde gerollt,
  nicht bearbeitet.
- **Nichts über die Gegenprobe außerhalb der Anwendung**: MSE2 öffnet die Setdatei
  nicht in diesem Beleg, kein Drucker druckt.
- **Nichts über die Pipeline.** Ob `make ci` am Schlussstand grün ist, steht hier
  nicht.
- **Nichts darüber, ob ein Zuschnitt der richtige ist.** 104 Teilaufgaben sind
  eine Zählung dieses Belegs; mit einem anderen Zuschnitt wäre die Zahl eine
  andere.
