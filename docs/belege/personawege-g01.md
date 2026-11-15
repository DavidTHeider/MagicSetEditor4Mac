# Beleg: Vier Personas, jeder Weg in einem Zug — G01

**G01** (#2248),
Abnahmekriterium: »Entwickler: Diagnose/API/Build; Admin: Rolle/Sicherung/Update;
erfahrener Fan: Import/Deck/Netz; Fan ohne IT: Erstkarte/erste Liste/Export.
Aufgabenbeginn, Entscheidungen, Fehler und Ziel dokumentieren; Klicks/Zeit
vorher/nachher aus B02/B03. Menschliche Erfolgsquote nur nach tatsächlicher
Studie berichten.« Erhoben am 2026-10-10/11 am Zweig `fix/2290-leser-eingaben`.

## Methode und Grenzen

**Je Persona ein Drehbuch, ein Start, ein Zuhause** — `g01einsteiger`,
`g01fan`, `g01admin`, `g01entwickler` unter `scripts/huellenprobe/`. B03 hatte
die Teilaufgaben aus getrennten Drehbüchern zugeordnet und selbst gesagt, was
das nicht belegt: einen durchgehenden Weg
(`docs/belege/arbeitswege-phase-b.md`, »Was dieser Beleg nicht sagt«). Hier
geht jede Persona ihre drei Wege hintereinander am nativen Fenster.

**Gezählt wird wie in `ausgangswerte`** (B02, UX-01.AK-03): eine Handlung ist,
was ein Mensch einmal tut — Klick, Kürzel, getippter Wert, Wahl in einer
Liste, Wahl im Dateidialog; Rollen zählt nicht. Der Zähler steht seit diesem
Werkstück in `scripts/huellenzaehler.mjs` und wird von beiden benutzt — eine
zweite Zählung daneben wäre eine zweite Wahrheit. **Ein provozierter
Fehlgriff zählt mit**: auch ihn tut ein Mensch.

**Die Dauer ist Maschinenzeit**, vom ersten Handgriff bis die Wirkung im
Bestand steht — keine Bearbeitungszeit eines Menschen. Neben den Läufen lief
die Fuzzkampagne von C02 mit fünf Zielen zugleich; die Zeiten tragen diese
Last und sind nur untereinander vergleichbar.

**Je Abschnitt stehen Beginn, Entscheidungen, ein Fehler und das Ziel.** Der
Fehler ist **provoziert** — der, der einem Menschen dort begegnet —, und
notiert ist, was die Anwendung dazu sagt. Das Ziel wird **am Bestand**
gelesen (`mse`, eine HTTP-Antwort, eine Datei), nicht am Bildschirm.

**Keine Nutzerstudie, keine Erfolgsquote.** Die Drehbücher kennen den Weg.
Ob ein Mensch ihn findet, misst keines.

**Ausgangsbestand je Persona**: der frische Bestand mit der Saat; der
erfahrene Fan bringt zwei Drucke (über die Liste und Scryfall eingelesen) und
den Einladungscode eines Freundes mit (eine zweite Identität auf diesem Mac);
der Admin ein Paket in alter Fassung und ein Set darauf. Was so entsteht,
läuft über `mse` und zählt nicht.

## Die Wege

Spalte **vorher** ist B02/B03 (`ausgangswerte`, Freigabestand und Schlusslauf
von Phase B) oder der erste Lauf dieses Werkstücks vor seinen Korrekturen;
**nachher** ist der Schlusslauf am Commit `c0988636d` (Lauf
`2026-10-10T23-41-32Z-5848`; der Admin am Commit `d7ca57bbe`, Lauf
`2026-10-10T23-44-04Z-7895`, nach einer Berichtigung seines Drehbuchs).

### Fan ohne IT — Erstkarte, erste Liste, Ausdruck (`g01einsteiger`)

| Abschnitt | Beginn | Entscheidungen | Fehler und was die Anwendung sagt | Ziel | vorher | nachher |
| --- | --- | --- | --- | --- | --- | --- |
| Erstkarte | erster Start, die Einführung steht | Stil übernehmen oder wählen; wie Set und Karte heißen | keiner provoziert (ein leerer Setname wird am Feld abgewiesen, `uxr04`) | eine Karte mit Name und Text im Bestand | 4 Handlungen **im offenen Set** (B02); von der leeren Anwendung 7 laut `docs/UX-ARCHITEKTUR.md` (2026-09-03, nicht neu gemessen) | **10 Handlungen, 3,7 s** — ab der Einführung, samt Set und Text |
| Erste Liste | die leere Sammlung | welcher Druck bei einem Namen ohne Set | »Gibtesnichtxyz« — vorher der rohe Fehler »nicht gefunden: Scryfall kennt eine Karte zu dieser Suche nicht — und es gibt keinen Stand im Speicher«, deutsch in jeder Sprache; jetzt »Keine Karte heißt so — prüfe die Schreibweise.« und »Eine Zeile findet keine Karte und wird übergangen.« | die Exemplare in der Sammlung (»2 neu, 0 erhöht, 5 Stück«) | nicht gemessen (B03: kein Drehbuch) | **7 Handlungen, 4,4 s** |
| Ausdruck | das offene Set | Papier und Anschnitt — die Vorgaben bleiben | keiner provoziert (ein gesperrter Ort nennt Ursache und Schritt, `fix-f02`) | ein Druckbogen als PDF | nicht gezählt (B03: ⌘P öffnet den Dialog) | **3 Handlungen, 0,7 s** |

### Erfahrener Fan — Import, Deck, Netz (`g01fan`)

| Abschnitt | Beginn | Entscheidungen | Fehler und was die Anwendung sagt | Ziel | vorher | nachher |
| --- | --- | --- | --- | --- | --- | --- |
| Import | die Setseite | Datei oder Verzeichnis; welches Spiel zuordnet | eine JSON-Datei ohne Set — vorher »… an Byte 22 steht string, erwartet war mtgjson.rohSet. Berichtige die Angabe …«; jetzt »notizen.json ist JSON, enthält aber kein Set im Format MTG JSON (an Byte 22 weicht sie ab). Wähle eine Datei mit einem Set — oder ein ganzes Verzeichnis.« | die Sets des Verzeichnisses im Bestand | 4 Handlungen ohne Fehlgriff (B02) | **6 Handlungen, 1,4 s** — davon 2 für den Fehlgriff |
| Deck | die Sammlung mit zwei Drucken | welche Exemplare; Name und Spiel | keiner provoziert | beide Drucke als Plätze im Deck | **nicht gefahren** (B03: A07.2, »ein Deck aus Exemplaren der Sammlung baut die Oberfläche in keinem Drehbuch«); Deckbau aus einem Set 10 (B02) | **10 Handlungen, 3,0 s** — über das Auswahlband |
| Netz | Handeln, der Code des Freundes liegt bereit | stimmen die dreißig Ziffern; darf er handeln | ein verstümmelter Code: »die Unterschrift passt nicht — die Einladung wurde verändert …«; **und dem Profil fehlt die eigene Identität** | der Freund bestätigt, mit Erlaubnis zu handeln | **nicht erreichbar**: der erste Lauf endete an »Lege sie im Kasten ‚Womit sich dieses Profil ausweist‘ an« — der Kasten steht nicht in der Schublade | **11 Handlungen, 2,6 s** — davon 2 für den Fehlgriff, 1 für die Identität am Fehler |

### Admin — Rolle, Sicherung, Update (`g01admin`)

| Abschnitt | Beginn | Entscheidungen | Fehler und was die Anwendung sagt | Ziel | vorher | nachher |
| --- | --- | --- | --- | --- | --- | --- |
| Rolle | Nutzer und Profile | wer; Leser, Mitarbeiter oder Verwalter | ohne Namen: der Knopf ist gesperrt, solange das Feld leer ist — verhindert statt gemeldet | Mia liest das eigene Profil mit (`mse freigaben`) | nicht gemessen (B03: Rolle über die Kommandozeile vergeben) | **9 Handlungen, 2,4 s** — samt dem eingeklappten Kasten der Freigaben |
| Sicherung | Daten und Sicherung | mit Kennwort; wohin; zurückspielen | beim Zurückspielen das falsche Kennwort: »Das Kennwort passt nicht zu dieser Sicherung.« | der Stand der Sicherung ist zurück, das spätere Set fort | Wiederherstellung **ohne** Kennwort 5 Handlungen (B02); mit Kennwort **nicht gefahren** (B03, Hürde 2) | **17 Handlungen, 8,1 s** — Sichern mit Kennwort, Fehlgriff, Zurückspielen |
| Update | Ressourcen › Pakete | den Plan lesen: Fassungen, Sets, Quelle | »Nichts ändern« beim ersten Plan | probe-fremd in der neuen Fassung | nicht gezählt (`dom01` fährt den Plan) | **7 Handlungen, 1,3 s** |

### Entwickler — Diagnose, API, Bau (`g01entwickler`)

| Abschnitt | Beginn | Entscheidungen | Fehler und was die Anwendung sagt | Ziel | vorher | nachher |
| --- | --- | --- | --- | --- | --- | --- |
| Diagnose | Einstellungen › Diagnose | welche Teile; wohin | ein Ort ohne Schreibrecht: »Auf diesen Ort darf dieses Programm nicht zugreifen — er ist gesperrt. Wähle einen anderen oder gib ihn frei …« | ein Diagnosepaket auf der Platte | nicht gezählt (`ops04`) | **7 Handlungen, 1,0 s** |
| API | Einstellungen › Verbindungen | wo der Dienst lauscht — die Vorgabe »Nur dieser Mac« | eine Anfrage ohne Zugang: 401 | 200 auf `/sets` | **teilweise** (B03: REST am Fenster fuhr kein Drehbuch; `g08` seit Phase G) | **3 Handlungen, 1,0 s** |
| Bau | das Menü Hilfe | — | ein Bau mit ungesicherten Änderungen sagt es | die Hilfe nennt den Commit des Bündels (`c098863`) | **nicht geprüft** (B03) | **2 Handlungen, 0,6 s** |

### Und die fünf Kernwege, zur Gegenprobe des Zählers

Derselbe Lauf wie beim Admin fuhr `ausgangswerte` mit dem gezogenen Zähler:
Kartenanlage **4**, Import **4**, Deckbau **10**, Spielstart **9**,
Wiederherstellung **5** Handlungen — genau die Zahlen von B02 vor dem Umzug
(`docs/belege/arbeitswege-phase-b.md`, »Vier Personas«). Der Zähler zählt
also wie vorher; die Maschinenzeiten (0,9 / 0,3 / 1,5 / 1,8 / 0,6 s) tragen
die Last der Fuzzkampagne.

## Die Hürden

### Behoben, je mit rotem Test

| Hürde | Persona | Commit |
| --- | --- | --- |
| Die Hilfe beschrieb den Stand vor der Ein-Klick-Einrichtung und der Lauschwahl: MCP »kein Port«, »installiert nichts … legt den Verweis selbst«, REST »auf 127.0.0.1 und auf keiner anderen Adresse«, »startet nicht von selbst«, »führt keine Befehle für dich aus« | Entwickler | `bc0d8fa04` |
| Eine Listenzeile ohne Karte zeigte den rohen Fehler samt Klasse, deutsch auch im englischen Fenster; die Bilanz zählte sie zu den offenen Entscheidungen; die Auswahl des Drucks ragte über die Schublade | Fan ohne IT | `8287170df` |
| Ein Profil ohne Identität kam im Tausch nicht weiter: die Meldung nannte einen Kasten, der nicht in der Schublade steht | erfahrener Fan | `33c448211` |
| Eine MTG-JSON-Datei ohne Set nannte den Typ des Dekodierers | erfahrener Fan | `34cd448ae` |
| »0 Pakete durchgesehen« neben sieben Paketen unter Ressourcen | Fan ohne IT | `ec69d6528` |
| Der Verlauf nannte die erste Karte »angelegt oder gelöscht«, mit »kein Wert« vorher und nachher | Fan ohne IT | `46038f5ba` |

### Gesehen und nicht geändert — mit Grund

| Beobachtung | Warum nicht hier |
| --- | --- |
| Unter *Pakete* steht zwischen der Liste und den Kästen darunter freier Raum | Absicht: der Arbeitsplatz nimmt die Fensterhöhe und rollt in sich (DOM-01, Artboard); bei hunderten Paketen ist er voll |
| Ein Exemplar, das nur mit seiner Scryfall-Kennung angelegt wurde, zeigt die Kennung statt des Namens | Nur über `mse exemplar` oder ein CSV ohne Namen; die Wege im Fenster holen den Namen. Ein Nachladen der Namen wäre eine neue Fähigkeit |
| »Für andere erreichbar sein« (Handel und Tisch) nimmt eine Adresse als Freitext, die Dienste dagegen die Lauschwahl | G08 hat die Lauschwahl für REST und MCP gebaut; der Horchposten des Netzes ist ein eigener Weg mit eigener Prüfung (`seitentechnik.ts`, Zeilen 306–313). Eine gemeinsame Wahl gehört in ein eigenes Paket |
| Das Update der **Anwendung** fährt kein Persona-Weg | braucht eine Freigabe auf GitHub und ein ausgeliefertes Bündel — I03 (#2261); das Update eines **bezogenen Paketbaums** belegt `g04bezuege` gegen GitHub selbst |
| Ein neuer Nutzer bekommt ein Profil »Mein Profil« — in der Profilwahl des Admins stehen dann zwei gleichnamige, nur am Eigner darunter zu unterscheiden | Der Name stammt aus der Anlage (`nutzer-neu`: »einen Nutzer samt erstem Profil«) und ist aus seiner Sicht richtig; wer ihn ändert, ändert die Anlage für alle Flächen |

## Was dieser Beleg nicht sagt

- **Nichts über Menschen**: keine Erfolgsquote, keine Bearbeitungszeit, keine
  Verständlichkeit.
- **Nichts über zwei Geräte**: der Freund des erfahrenen Fans ist eine zweite
  Identität auf diesem Mac.
- **Nichts über echte Magic-Stile**: gestaltet wird im Saatstil; ein echter
  Pack wäre ein Download von über einem Gigabyte.
- **Nichts über das Bündel der Auslieferung**: gefahren ist das Abnahmebündel
  aus `scripts/bauen.sh abnahme`.
