# UX-Architektur — Objekt, Aufgabe, Muster, Zielzustand

**Werkstück U1 von Phase 5.1**
(#1412),
erhoben am **2026-09-08** am Stand `v0.4.0` — aus dem Quelltext
(`frontend/src/`, `internal/app/`) und an der laufenden Anwendung bei
1.280 × 800 gegen den Messbestand, der unten beschrieben ist.

**Wofür diese Datei da ist.** Vor 0.5.0 bekommen vier Seiten einen
Arbeitsplatz (U5, U7, U8, U10), zwei Bereiche entstehen neu (Spielen, Handeln),
und die Übersicht wird umgebaut (U11). Wer so etwas tut, verliert leicht eine
Funktion, die auf einer alten Seite stand und auf der neuen keinen Platz
bekommt. Diese Datei hält dagegen: **jede Seite und jede Lasche steht hier mit
ihrem Objekt, ihren Aufgaben, ihrem Muster und dem Zustand, den sie nach dem
Umbau haben soll** — und ein Wächter (`internal/werkzeuge/uxarchitektur_test.go`)
hält die Liste gegen `navigation.ts`, in beide Richtungen.

**Was hier nicht steht.** Das Urteil über jede Seite — was sie jemandem sagt,
der sie zum ersten Mal sieht — steht in `SEITENKRITIK.md` (Entwicklungsarchiv);
die Zahlen je Seite in [`SERIENARBEIT.md`](SERIENARBEIT.md). Hier steht die
**Ordnung**: welche Seite zu welchem Weg gehört, und was sie am Ende der
Iteration sein soll.

**Und dieselbe Ordnung als Datei** — [`inventur.json`](inventur.json),
UX-01.AK-01 (#1673).
Je Seitenschlüssel: Komponente, Aufgaben (aus der Tabelle unten), Einstiege
(Seitenleiste, Laschenleiste, Befehlspalette, Kürzel und Menü — abgelesen aus
`navigation.ts`, `bereiche.ts` und `menue.go`), wem die Seite gehört und
welche Rolle zählt, **fünf Zustände** — leer, laden, Fehler, offline, nur
lesen — und das Prüfergebnis. **Ein Zustand ist belegt, anwendungsweit, nicht
vorhanden oder offen**, und jeder dieser Stände trägt seine Deckung: einen
Test mit Namen oder einen Grund. `internal/werkzeuge/inventur_test.go` hält
beides gegen den Baum.

**Am 2026-09-23 sind 48 der 125 Zustände offen**, und das ist die Auskunft
und nicht ihr Mangel: kein sichtbarer Ladezustand ist auf einer Seite belegt
(S12 nennt es) — das sind 25 —, acht Seiten gehen ins Netz, ohne dass ein
Lauf ohne Netz sie zeigt, elf haben keinen Test, der ihren Leerzustand
benennt, und vier keinen für einen Fehlerzustand.

**Die native Prüfung ist seit dem 2026-09-26 bei jeder Seite bestanden**
(UX-01.AK-02): die Hüllenprobe `inventur` öffnet alle 25 Seiten am gepinnten
Abnahmebau über die Befehlspalette, in einem frischen Bestand mit eigenem
Schlüsselbund und unter eigener Kennung, liest Überschrift, Meldungen und
Überlauf und legt je Seite einen Abzug nach
[`belege/huellenprobe/inventur/`](belege/huellenprobe/inventur/) — die
Inventur nennt ihn als Beleg. **Keine Seite zeigte eine Fehlermeldung, keine
rollte waagerecht.** Die Zustände Laden, Offline und Nur-Lesen je Seite
bleiben, wie sie oben stehen.

**Die Ausgangswerte der Kernwege** (UX-01.AK-03), am selben Fenster
erhoben von der Hüllenprobe `ausgangswerte` — **Maschinenzeit, keine
Bearbeitungszeit eines Menschen**, und ohne Fehlversuche, weil die Probe
den Weg kennt:

| Kernweg | Handlungen | Fehlversuche (Probe) | Maschinenzeit | Was gezählt ist |
| --- | ---: | ---: | ---: | --- |
| Kartenanlage | 4 | 0 | 0,7 s | im offenen Set: *Karte anlegen*, ins Namensfeld, tippen, ⇥ |
| Import | 4 | 0 | 0,4 s | ⌘2, *Ein Set von woanders einlesen*, die Wahl im Dateidialog, *Verzeichnis öffnen …* |
| Deckbau | 10 | 0 | 1,0 s | ⌘4, *Decks*, *Deck anlegen*, Name, Spiel, *Anlegen*, Set, Karte, *Ins Deck legen* |
| Spielstart | 9 | 0 | 1,4 s | ⌘4, *Partie anfangen*, je Sitz Name und Deck, *Partie anfangen* |
| Wiederherstellung | 5 | 0 | 0,4 s | ⌘,, *Daten*, die Wahl im Dateidialog, *Sicherung zurückspielen …*, *Zurückspielen* |

**Pairing und Handel stehen nicht darin**: sie brauchen ein Gegenüber mit
eigenem Bestand, eigener Identität und eigenem Horchposten, und die zwei
getrennten Instanzen am Fenster trägt S12.NAT-04 (#1672). Ihre Ausgangswerte
erheben NET-01 (#1705) und TRADE-02 (#1716), **bevor** sie diese Wege
umbauen. **Nach ENG-02 gemessen, nicht davor**: die Ausgangslage unter
Wails 2 lässt sich nicht nachholen — die Hüllenprobe gibt es erst mit Wails 3.

**Und die achtzehn Arbeitswege des nativen Audits stehen darin** — unter
`arbeitswege`, UX-01.NAT-01: je Weg Aufgabe und Status **wörtlich** aus
`pruefaufgaben.csv` (Stand 0.5.0), die Seiten, **was offen ist** — aus der
Spalte »grenze« — und wer es trägt. Darunter die sieben, die NAT-01 beim
Namen nennt: Vorlagenkopie mit Karteneinsatz, sammlungsbasierter Deckbau,
Undo, Token, Counter und Anhänge, Mindestgröße, reduzierte Bewegung und der
vollständige Fokuszyklus. **Übernommen heißt nicht bestanden**: ob ein Weg
heute geht, zeigt die native Prüfung, und die trägt S12 in Phase 6.11. Hier
fehlten sie bis zum Zweitblick über Phase 6.0; `inventurwege_test.go` hält sie
seitdem gegen das Audit.

---

## Die sieben Arbeitswege

Jede Seite dient mindestens einem. Alle sieben gibt es heute; reserviert ist keiner mehr.

| Weg | Was jemand tut | Seiten |
| --- | --- | --- |
| `karte` | ein Set öffnen, eine Karte anlegen und füllen, die nächste — und das Set hinausschreiben | Start, Sets, Karten, Setangaben, Schlüsselwörter, Konsole, Zufallspack, Symbol, Stil, Suche, Ersetzen, Statistik, Bilder, Druckbogen, Ausgeben |
| `paket` | ein Vorlagenpaket beziehen, einbinden, korrigieren | Pakete, Vorlagen |
| `sammlung` | eintragen, was man besitzt, es suchen, bepreisen, ordnen | Sammlung |
| `deck` | ein Deck bauen, gegen ein Format prüfen, bei einem Turnier einreichen | Decks, Turniere |
| `profil` | wer arbeitet, in welchem Profil, mit welchen Rechten — und was die Anwendung dabei tut | Nutzer und Profile, Einstellungen, Hilfe |
| `spielen` | eine Partie mit den eigenen Karten: Decks an den Tisch, Karten bewegen, durchsehen, mischen | Spielen (seit G1; ⌘4 über *Spielen ▸ Decks*, seit UX-02) |
| `handeln` | tauschen und handeln mit einem Partner: abgleichen, was zusammenpasst, sehen, was unterwegs ist, und verbuchen, was den Besitzer gewechselt hat | Handeln (seit T9; ⌘5, seit UX-02) |

**Die Kürzel gehören seit UX-02 dem Bereich und nicht der Seite**
(#1674).
⌘1 bis ⌘6 führen auf die sechs Einträge der Leiste — *Start*, *Gestalten*,
*Sammeln*, *Spielen*, *Handeln*, *Ressourcen* —, und zwar je auf die **erste
Seite** des Bereichs. ⌘? und ⌘, sind macOS, ⌘' sind Nutzer und Profile;
⌥⌘1 bis ⌥⌘0 sind die Laschen des offenen Sets und bleiben unverändert.

**Was hier vorher stand, und warum es nicht mehr gilt.** Bis zum 2026-09-21
lautete der Satz „⌘1 bis ⌘0 führen auf die Leiste", und jede Seite der Leiste
hatte ihre eigene Ziffer — vierzehn Einträge, zehn Ziffern, vier ohne. Der
Entwurf der Zieloberfläche ordnet sie zu **sechs Bereichen**, und UX-02.AK-01
verlangt genau das. Die Migration steht in
[`CHANGELOG.md`](../CHANGELOG.md), wie AK-03 es verlangt.

**⌘7 bis ⌘0 sind damit frei**, und das ist eine Aussage über heute und keine
Reservierung: wer sie vergibt, vergibt sie an einen Bereich. `reservierteKuerzel`
in `frontend/src/navigation.ts` bleibt leer — eine Taste zugleich frei und
vergeben zu nennen, war schon einmal der Fehler.

**Und die Null steht damit nicht mehr am Ende der Leiste.** Die Überlegung,
dass sie auf der Tastatur rechts neben der Neun liegt, galt einer Ordnung
nach Ziffern; die Ordnung kommt jetzt aus `frontend/src/bereiche.ts`, und die
Leiste liest sie.

---

## Jede Seite und jede Lasche

**Objekt** ist, worüber die Seite spricht. **Aufgabe** ist, was jemand darauf
tut — die Schreibwege stehen vollständig in
[`SCHREIBWEGE.md`](SCHREIBWEGE.md). **Muster** ist der heutige Aufbau.
**Zielzustand** ist, was die Seite nach der Iteration ist — und welches
Werkstück sie dorthin bringt. Wo „bleibt" steht, ändert sich die Seite nicht;
das ist eine Aussage, keine Auslassung.

### Die Anwendungsseiten

| Schlüssel | Seite | Objekt | Aufgabe | Muster heute | Arbeitsweg | Zielzustand |
| --- | --- | --- | --- | --- | --- | --- |
| `start` | Start | das Ganze: Sets, Karten, Pakete, Exemplare | vier Kacheln zur Navigation, „Zuletzt benutzt" (neun Sets) — keine Schreibbindung | Seitenkopf, Kachelraster, Liste | `karte`, `profil` | **gebaut (U11, 2026-09-09)**: Überblick, Weiterarbeiten mit Angeheftetem und dem jüngsten Schritt als Handlung, Aufmerksamkeit aus Vorbehalten, Fehlkarten und Skriptbefunden, Hintergrundaufgaben aus dem Auftragsbuch; leere Bereiche mit ausführbarer Handlung. **Umgebaut mit UXR-03 (#1911)**: die vier Zahlen als eine Zeile (Sets, Karten, Exemplare, Decks), Weiterarbeiten als eine Reihe neben der Aufmerksamkeit, die Hintergrundaufgaben nur, solange das Buch etwas führt |
| `sets` | Sets | Set | öffnen, anlegen, sichern, sichern unter, über Exportvorlage, als MTG JSON, löschen (Name bestätigt), Beispielsets, MTG JSON und Scryfall einlesen | Liste aus Karten, Anlage darunter, Aufklappkasten mit drei Einfuhrwegen | `karte` | bleibt; gewinnt Field/Callout aus U2/U3 und die Konfliktdarstellung aus U6 |
| `pakete` | Pakete | Paket, Bezugsquelle, Sammeldaten | beziehen (mit Zustimmung an Ort und Stelle), einbinden, anlegen, entfernen, suchen, filtern, MSE2 übernehmen, Sammeldaten holen | Arbeitsplatz aus Navigator, Baumgitter und Inspector; **seit UXR-17** (#1925) Bezug, Übernahme und Umschlag in einer Schublade hinter *Paket beziehen …*, darunter Anlegen, Abhängigkeiten und Sammeldaten eingeklappt | `paket` | bleibt; Paketdiagnose aus U7 zeigt Befunde je Paket am Eintrag |
| `vorlagen` | Vorlagen | Paket, seine Textdateien und sein Beiwerk | Datei im Baum wählen, ändern, Änderungen sehen, speichern, Dateien anlegen, hinzufügen, umbenennen und löschen (FMT-02), bei Konflikt neu laden, Testkarte wählen, zur Stelle eines Befunds oder zur Definition eines Namens springen | Arbeitsplatz: Paketbaum mit Suche, Editor mit Speichern und Diff, Inspector mit Testkarte, Vorschau, Diagnose, Herkunft | `paket` | **gebaut (U7, 2026-09-09)**: Paketbaum als Navigator, Editor mit Vervollständigung und Befund an der Stelle im Zentrum, Vorschau mit der Marke „veraltet“, Diagnose und Herkunft eines Namens im Inspector |
| `sammlung` | Sammlung | Exemplar | aufnehmen, ändern, löschen, Liste einlesen, CSV hinaus und zurück, Preise holen, Masse ändern/löschen, suchen, sortieren, filtern | Werkzeuge im Aufklappkasten, Sicht, Liste (gefenstert, #1500), Portfolio | `sammlung` | **Sammlungsarbeitsplatz** (U8): Tabelle, Galerie oder Gruppen im Zentrum, virtualisiert über 10.000 Zeilen; Menge, Zustand, Erwerb und Preisalter im Inspector |
| `decks` | Decks | Deck, Platz, Format | anlegen, öffnen, umbenennen, löschen, Karte hineinlegen und herausnehmen, Format wählen und prüfen, Formate verwalten, aus fremdem Profil holen | **gebaut (U10, 2026-09-09)**: Arbeitsplatz mit drei Bereichen — Navigator (Deckliste, Kartensuche, Anlage, Übernahme), Zentrum (Format, Plätze mit +/− — „−" ruht bei eins seit DOM-03 — und Bereichswechsel), Inspector (Auswertung samt Kennzahlen, Prüfung, Zuordnung); bei Enge weicht der Inspector. Die Prüfung läuft bei jeder Änderung mit. **Seit UXR-11 (2026-10-02, #1919)**: der Navigator führt mit „Deck anlegen“ und trägt je Deck nur Öffnen und die Marken — Umbenennen, Duplizieren, Teilen und Löschen stehen am offenen; Kartenwahl und „Eine Deckliste einlesen“ stehen in der Mitte über den Plätzen, und die Knöpfe eines Platzes an der Zeile mit der Tabstelle (Pfeile wandern) und unter dem Zeiger; die Zuordnung zeigt oben, was fehlt, die Zeilen eingeklappt; ein Format ist eine Zeile mit „Bearbeiten“ | `deck` | **erfüllt** (U10): Decks links, Deck in der Mitte — seit UXR-11 samt der Kartenwahl, vorher links —, Analyse rechts; der Regelbefund steht **am Posten** (Marke an der Zeile) **und im Kopf** (Lasche »Prüfung«); Bedarf gegen Bestand aus U9 |
| `turniere` | Turniere | Turnier, Teilnehmer, Runde, Paarung | anlegen (Assistent in vier Schritten: Turnier, Regeln, Teilnehmende, erste Runde — auch nach einer Vorlage), öffnen, löschen, an- und abmelden, Runde ansehen und legen, Ergebnis eintragen und zurücknehmen | Assistent in einer Schublade hinter *Turnier anlegen …*; das offene Turnier als Werkbank mit fünf Laschen (Runden, Rangliste, Teilnehmende, Decklisten, Protokoll) über der Liste — eine Paarung ist eine Zeile, ihr Eintragen klappt an ihr auf (UXR-12) | `deck` | bleibt; Deckabzug wird in 5.5 (G7) an eine Spielsession gebunden |
| `spielen` | Spielen | Partie, Sitz, Spielobjekt | eine Partie aus Decks anfangen, die Sicht eines Sitzes ansehen, Karten bewegen, tappen, umdrehen, die Bibliothek durchsehen und mischen, anhalten und beenden | Anfang in zwei Tafeln — wer sitzt, nach welchen Regeln, die Grenzen eingeklappt; **seit UXR-13** (#1921) *Deine Partien* als Spalte daneben (ab `@4xl`, sonst darunter), je Partie eine Zeile mit den Menüs *Weiterspielen* (Sitze und Zuschauen) und *…* (Duplizieren, Ausführen, Löschen, Marken), vier vorn und der Rest stückweise; *Bei jemandem mitspielen …* und *Verleihen oder leihen …* im Kopf öffnen je eine Schublade samt Kasten für ein Gegenüber, und wer an einem fremden Tisch sitzt, hat ihn oben auf der Seite; im freien Modus die Regelgruppen eingeklappt; gezeichneter Tisch mit einer Spalte daneben — Zug, Spieler, gewählte Karte und ihre Handlungen als Knöpfe, deren Felder erst auf Zuruf aufklappen; Anhalten, Verlauf und Freigabe unter dem Tisch | `spielen` | **gebaut (G1, 2026-09-15)**, als gezeichneter Tisch seit G3 (#1460); nach dem Artboard *Tisch* angeordnet seit dem 2026-09-24 (F09, UX-08 #1680) |
| `nutzer` | Nutzer und Profile | Nutzer, Profil, Konto, zweiter Faktor, Freigabe, Token | anlegen, ändern, löschen (in der Zeile bestätigt), Bild und Abzeichen, wechseln (Profil und Mensch), Konto und Anmeldung, zweiter Faktor, Freigaben, Token, Vorgaben, Daten sichern | vier Laschen, je Lasche Formulare untereinander | `profil` | bleibt als Verwaltung; **wer arbeitet** wandert in die globale Leiste (U4) und ist von dort aus erreichbar |
| `hilfe` | Hilfe | die Texte | lesen, springen, „Über" | acht Laschen aus Beschreibungslisten | `profil` | **gebaut (U11, 2026-09-09)**: zehn Laschen — dazu Führungen (die ersten Schritte lassen sich erneut zeigen), die Kürzeltafel und die drei Bereiche ohne Seite (KI mit Weg zu den Schnittstellen, Spielen und Handeln als reserviert) |
| `einstellungen` | Einstellungen | Einstellung, Dienst, Zugang, Bestand | Werte setzen, Dienste schalten, Zugänge und Geheimnisse, Bestand sichern und zurückspielen (Rückfrage im Seitenbereich), Verschlüsselung, Diagnose, Kürzeltafel | bis zu acht Laschen, Formularraster | `profil` | **gebaut (U11, 2026-09-09)**: Suche über alle Gebiete mit Sprung zur Einstellung, je Wert Herkunft, Ebene und wirksamer Wert, Zurücksetzen auf die Vorgabe am Feld |

### Die Laschen eines offenen Sets

Alle dreizehn stehen im Setarbeitsplatz (`rahmen/Setarbeitsplatz.tsx`): Kopf
mit Setname und Kennzahlen, darunter eine Laschenleiste in fünf Gruppen, ohne
offenes Set der Leerzustand „Kein Set offen".

| Schlüssel | Lasche | Objekt | Aufgabe | Muster heute | Arbeitsweg | Zielzustand |
| --- | --- | --- | --- | --- | --- | --- |
| `karten` | Karten | Karte im Set; Marke, Notiz, Verweis, Ablage | anlegen (auch aus Scryfall und aus einem Skript), duplizieren, löschen (Rückfrage, „Trotzdem löschen" bei Deckbezug), ordnen, kopieren/ausschneiden/einfügen, blättern (⇞/⇟), Felder füllen, Bild wählen und zuschneiden (im Editor, kein Modal), Symbol, Verweis, Marken, Notizen, Spalten, Massenänderung, Scryfall-Abgleich | Werkzeugleiste, Liste links, Formular und Vorschau rechts, Massenänderung darunter; unter rund 1.000 Punkten gestapelt | `karte` | **Kartenarbeitsplatz** — **gebaut in U5** (#1416): Navigator mit Liste, Schnellfilter (MSE2-Sprache, an allen vier Flächen) und Marken; Vorschau mit Zoomleiste (Anpassen, 50–200 %, Druckgröße aus `card dpi`) und den Überlagerungen Bildgrenzen, Diagnose, Beschnitt im Zentrum; Inspector mit Problemliste, die zum Feld führt, und dem Formular. Miniaturen im Navigator und Verlauf/Vergleich kommen mit U6. **Seit UXR-06** (#1914) über der Karte **eine** Zeile — ←, Stelle, →, das Menü *Diese Karte* mit Zwischenablage, Bezug zu einem Druck und Teilen — und darunter die Ansicht: der Maßstab als Auswahl, die drei Überlagerungen daneben; *Anpassen* passt die Karte in Breite und Höhe des Zentrums; Verlauf und Massenänderung zeigen zehn Zeilen und den Rest stückweise |
| `setangaben` | Setangaben | die `set field:`-Werte und die Dateiangaben | Werte ändern (beim Verlassen geschrieben), Vorbehalte ansehen | Fläche, Dateiangaben, zweispaltiges Formular; Fehler am Feld | `karte` | bleibt; Field aus U2/U3 |
| `schluesselwoerter` | Schlüsselwörter | Schlüsselwort des Sets und des Spiels | anlegen, ändern, löschen, die des Spiels zeigen, Parameter einsetzen | Knopfreihe, Liste aus Formularen, eingeklappter Block | `karte` | bleibt |
| `konsole` | Konsole | ein MSE-Script-Ausdruck über der aufgeschlagenen Karte | auswerten, Namen zeigen und eingrenzen, Beispiele einsetzen, Vervollständigung | Quelltextfeld mit drei Zeilen, Ergebnisblock | `karte` | **Skriptkonsole** (U13): Editor aus U7, Verlauf, Abbruch, Ergebnisbaum; aus der Kommandozentrale (U12) erreichbar |
| `pack` | Zufallspack | Packart und die gezogenen Karten | Packart wählen, Same setzen, ziehen | Steuerleiste, Ergebnisliste | `karte` | bleibt |
| `symbol` | Symbol | das Symbol des Sets oder eines Felds | Grundformen, nachzeichnen, Punkte, Gruppen, Symmetrie, Verrechnung, ins Set legen | Werkzeugleiste, Zeichenfläche, Symmetriefeld, Teilliste | `karte` | bleibt |
| `stil` | Stil | Stilpaket und Stiloptionen | Stil wählen (Rückfrage mit Folgen), Optionen setzen und zurücksetzen | Auswahlfeld, Rückfragekasten, Optionen, Vorschau | `karte` | bleibt; die Stiloptionen wandern in den Inspector des Kartenarbeitsplatzes (U5), die Lasche bleibt der Ort des Wechsels |
| `suche` | Suche | Treffer über alle Sets | suchen, auf Feld und Set eingrenzen, Treffer aufschlagen | Filterzeile, Trefferliste | `karte` | bleibt — **die Filterzeile im Navigator, die das Artboard meint, ist der Schnellfilter des offenen Sets (U5)**; diese Seite sucht über alle Sets und trägt die Recherche (A16). Seit UXR-05 (#1913) gefaltet: sechzig Treffer, dann „Weitere … zeigen“ (bis zu 200 waren 200 Knöpfe, R6), und das Zusammenstellen für ein Modell steht hinter seinem Aufklapper (R3) |
| `ersetzen` | Ersetzen | Textfelder eines Sets | vorschauen, dann ersetzen (Zweischritt statt Dialog), ⌘Z nimmt den Zug zurück | Formularleiste, Ergebnisliste vorher/nachher | `karte` | bleibt; die Vorschau ist das Vorbild für die Massenänderung aus U6 |
| `statistik` | Statistik | das Set als Ganzes | Feld und Dimension wählen | Kennzahlen, Balkenlisten | `karte` | bleibt |
| `bilder` | Bilder | die Kartenbilder | Ordner wählen (nativer Dialog), Vorlage, Konflikt, Format, Güte, Breite, Auswahl, ausgeben, anhalten | Formular, Knopfpaar, Laufblock | `karte` | bleibt; Fortschritt und Abbruch mit Live-Region (U3) |
| `druckbogen` | Druckbogen | der PDF-Bogen | Papier, Anschnitt, Auflösung, Schnittmarken, Auswahl, legen (nativer Dialog) | Formular, ein Knopf, Ergebnisblock | `karte` | bleibt. **Die Handprüfung mit dem Lineal bleibt getrennt** — sie steht in der Roadmap als offene Handprüfung 5 und ist kein Teil dieser Messung |
| `ausgeben` | Ausgeben | die fünf Wege hinaus | Setdatei, Bilder, PDF, Exportvorlage, MTG JSON | fünf Flächen untereinander | `karte` | bleibt |
| `handeln` | Handeln | Gegenüber, Abgleich, Vorgang | ein bestätigtes Gegenüber aufschlagen, den Abgleich fahren, sehen, was unterwegs ist, und verbuchen, was den Besitzer gewechselt hat | zwei Laschen: *Mit wem* („Was gerade läuft" links, darüber „Neuer Tausch …" mit Wahl, Aufnahme und Kontakten in einer Schublade — seit UXR-15, #1923; der Handelsraum rechts) und *Was es wert ist* | `handeln` | **gebaut (T9, 2026-09-19)** — was hier steht, stand vorher in einer Kontaktzeile unter *Technik*, mit dem Satz „bis der Bereich kommt". **Eine Liste der laufenden Handel fehlt**: die Schritte von der Eröffnung bis zur Übergabe haben bis heute keinen Aufrufer (#1637); solange ist der Vorgang ein Feld und keine Auswahl, und der Satz darüber sagt es |

### Die Kontextleiste

**Seit Werkstück U4** (#1415)
steht in der Seitenleiste unter dem Anwendungsnamen, **wer wo als was
arbeitet**: Bild und Name des Menschen, sein Profil, dessen Eigner, wenn es
ein fremdes ist, die wirksame Rolle und „nur lesen" als Wort. Sie liest den
`Arbeitskontext` ([0078](entscheidungen/0078-der-wirksame-akteur.md)) — einen
Ruf, einen Stand — und wird beim Profilwechsel mit der Anwendung neu
aufgebaut, damit kein Wechsel zwischen zwei Rufe fällt. Sie ist
`role="region"` mit Namen und damit anspringbar; ihr Inhalt ist
`role="status"`. Was sie **nicht** tut: Profile wechseln oder Zugänge
verwalten — dafür führt ein Knopf zur Seite *Nutzer und Profile*, denn zwei
Wege zu derselben Handlung wären zwei Stände.

**Und der Arbeitsplatz hat seit demselben Werkstück ein Gedächtnis.** Ob
Navigator und Inspector stehen und wie breit sie sind, liegt in den vier
Schlüsseln der Ebene `arbeitsplatz` aus
[0079](entscheidungen/0079-einstellungsebenen-und-portabler-import.md) —
je Profil gesetzt, vom Menschen geerbt, aufgelöst wie jede Einstellung. Der
Baustein `Arbeitsplatz` bekommt es als Eigenschaft (`useArbeitsplatzgedaechtnis`)
und meldet jedes Klappen dorthin, statt selbst zu klappen; ein Wechsel ist
sofort zu sehen und kommt nur dann zurück, wenn das Schreiben scheitert (§7).
Zu sehen ist das mit dem ersten Verbraucher des Bausteins, dem
Kartenarbeitsplatz (U5) — bis dahin ist es eine Form mit Tests und ohne
Fenster, und das steht hier, damit es niemand für eine Fähigkeit hält.

---

## Dialoge und Rückfragen

**Modal ist genau dreierlei**: die Befehlspalette (⌘K), die Einführung beim
ersten Start — und seit U12 die Kommandozentrale über einem Befehl, die aus der
Palette aufgeht, nachdem diese zu ist
([0084](entscheidungen/0084-die-kommandozentrale-ruft-die-werkzeuge-des-mcp-servers.md)). Jede Bestätigung im Datenbereich ist eine **Rückfrage an der
Klickstelle** — §6 verbietet verschachtelte Dialoge, und der Bestand hält
sich daran:

| Rückfrage | Wo | Muster | Zielzustand |
| --- | --- | --- | --- |
| Set löschen | Sets | Name abtippen, dann „Endgültig löschen" | bleibt — das Muster für alles, was nicht zurückkommt |
| Karte löschen | Karten | „Löschen" / „Abbrechen"; bei Deckbezug „Trotzdem löschen" | bleibt; U3 gibt ihr den Entscheidungsdialog-Baustein aus U2 |
| Deck, Turnier löschen | Decks, Turniere | Frage am offenen Deck oder Turnier (UXR-11, UXR-12) | dito |
| Nutzer, Profil löschen | Nutzer und Profile | in der Zeile, Name aus der Zeile | dito |
| Stil wechseln | Stil | Frage mit Folgen (Spielwechsel, ungedeutete Felder, verlorene Optionen) | bleibt |
| Ein Turnier wiedereröffnen | Turniere, Kopf der Werkbank | Frage mit Folgen: danach lassen sich Teilnehmende, Runden und Ergebnisse wieder ändern | **seit dem 2026-09-27** (DOM-04.AK-04) — das Abschließen fragt nicht: es verliert nichts, und das Wiedereröffnen holt es zurück |
| Ein Turnierergebnis berichtigen | Turniere, an der Paarung | Frage mit Folgen: wer Platz oder Punkte ändert, welche späteren Runden nach dem alten Stand gepaart wurden und bleiben — nur, wenn sich etwas verschiebt; das erste Ergebnis der laufenden Runde fragt nie | **seit dem 2026-09-27** (DOM-04.AK-02) |
| Ersetzen | Ersetzen | Vorschau, dann „Jetzt ersetzen" mit dem Stand der Vorschau | bleibt — Vorbild für die Massenänderung (U6) |
| Sicherung zurückspielen | Einstellungen | Kasten im Seitenbereich mit Datei, Datum, Zählung, Warnung | bleibt (O4, O7 erweitern ihn um den Rückweg) |
| Verschlüsselung | Einstellungen | Frage mit Grenzen | bleibt |
| Paket beziehen | Pakete | Herkunft und Lizenz zeigen, dann „Beziehen" | bleibt |
| Zweiter Faktor | Profil- und Nutzerwechsel, Konto | Feld klappt auf, wenn verlangt | bleibt |
| Zugang, Anbietergeheimnis, Token, zweiter Faktor, Nutzer, Profil, Partieprotokoll, alle eigenen Autoersetzungen | Einstellungen, Konto, Nutzer, Spielen, Karten | Rückfrage an der Klickstelle | **seit dem 2026-09-25** — bis dahin genügte ein Klick (UX-10.AK-04, unten) |

**Dass es bei drei Modalen bleibt, hält seit dem 2026-09-25 ein Lauf**
(`frontend/src/modalregel.test.ts`): nur Einführung, Palette und
Kommandozentrale benutzen den `Dialog`, nur `Dialog.tsx` und `Schublade.tsx`
fassen Mantines Modal und Drawer an, und niemand setzt `role="dialog"` von
Hand.

**Achtzehn native Dialoge** (Öffnen, Sichern, Verzeichnis) liegen in Go hinter
`internal/app/dateiwahl.go`; `MessageDialog` benutzt niemand. Sie gehören dem
Fenster — Kommandozeile, MCP und Dienst nehmen den Pfad als Argument (§16).

**UX-08.AK-03 für diese Dialoge** (#1680) —
*„Fokusführung, Escape, Rückkehr zum Auslöser und Layer-Reihenfolge
funktionieren auch bei nativen Dateidialogen; verschachtelte Modalketten
werden vermieden."* Was davon diese Anwendung tut und was macOS:

| Zusage | Wer sie hält | Beleg |
| --- | --- | --- |
| Fokus und Escape **im** Dialog | macOS — `NSOpenPanel` und `NSSavePanel` führen den Fokus selbst, Escape bricht ab | kein eigener Code; der Abbruch kommt als leerer Pfad zurück, und **ein leerer Pfad ist kein Fehler** (`EinstellungenSichern` und die übrigen) |
| **Rückkehr zum Auslöser** danach | `bausteine/fokus.ts`: gibt das Fenster den Fokus beim Schließen zurück, geht er an den Knopf, der gedrückt wurde — der hatte ihn nie, denn ein Klick lässt ihn in WebKit auf `body` oder gibt ihn dem Behälter darum. Ein Feld, in das danach geklickt wurde, und eine Taste heben den Druck auf | `fokus.test.ts`, „der Fokus nach einem nativen Dialog" — auch aus dem Behälter |
| **Layer-Reihenfolge** | ein nativer Dialog liegt über dem ganzen Fenster, als Blatt am Fenster | — |
| **Keine Modalkette** | die Befehlspalette schließt, **bevor** sie ausführt (`rahmen/Palette.tsx`); aus Einführung und Kommandozentrale öffnet keiner einen Dateidialog, und die Seiten, die es tun, sind keine Modale (`modalregel.test.ts`) | `palettetunfinden.test.tsx`, „schließt die Palette und führt zum Spielen" — ein Eintrag schließt sie im selben Klick, der ihn ausführt |

**Was davon Handabnahme bleibt**: dass WKWebView beim Schließen wirklich ein
`focus` am Fenster meldet. Die WebKit-Probe hat keinen nativen Dialog, sie
zeigt die Seite ohne Hülle.

---

## Muster — wann welches

**UX-08.AK-01** (#1680),
nach der linken Tafel des Artboards *BelegMuster*. Die Frage ist nicht, wie
etwas aussieht, sondern **wie viel es unterbricht** — von „gar nicht" bis
„alles".

| Muster | Wofür | Baustein | Hier zu sehen |
| --- | --- | --- | --- |
| Tooltip | Zusatzhilfe, die nichts verlangt | `bausteine/Tooltip.tsx` | jeder Knopf des Aktionskatalogs, die Fragezeichen der Einstellungen |
| Aufklappen | Angaben, die niemand zuerst braucht — **zu heißt verborgen, nicht fort**: was darin eingegeben war, steht beim Wiederaufklappen noch da (S11.NAT-03) | `bausteine/Aufklappkasten.tsx` | die Einfuhrwege auf der Setliste, Gegenüber finden, die Regelgrenzen im Partieanfang, die Diagnose eines Dienstes |
| Popover | eine kurze Eingabe dort, wo sie hingehört | `bausteine/Popover.tsx` | ein Deck umbenennen |
| Rückfrage | eine Bestätigung an der Klickstelle | `bausteine/Entscheidung.tsx` | jede Löschung, die fragt (Aktionen, oben) |
| Inspector | die Felder dessen, was gerade gewählt ist | `bausteine/Arbeitsplatz.tsx` | Kartenarbeitsplatz, Sammlung, Vorlagenwerkstatt |
| Schublade | Details zum Objekt, das im Blick bleibt | `bausteine/Schublade.tsx` | Eigenschaften und Werkzeuge des Sets, der Turnierassistent |
| Assistent | eine Aufgabe, mehrere Schritte, ein Ergebnis | `bausteine/Schrittfolge.tsx` | ein neues Set, ein neues Turnier (DOM-04.AK-01) |
| Modal | nur bei begründeter Unterbrechung | `bausteine/Dialog.tsx` | Befehlspalette, Einführung, Kommandozentrale |

**Die Reihenfolge ist die Regel**: wer ein Muster wählt, nimmt das oberste,
das die Aufgabe trägt. Eine Eingabe, die in ein Popover passt, bekommt
keine Schublade; ein Detail, das in den Inspector passt, keinen Assistenten.

**`frontend/src/musterkatalog.test.ts` hält die Tafel gegen den Quelltext**:
jeder genannte Baustein liegt da und wird von mindestens einer anderen Datei
benutzt. **Was er nicht hält**: ob eine Seite das richtige Muster gewählt hat
— das ist eine Frage an die Durchsicht und an diese Tafel.

### Kennungen, die niemand tippt

**UX-08.AK-04**: *„Normalnutzende müssen keine Netzwerk- oder
Transaktionskennungen eingeben, soweit diese technisch ableitbar sind;
Expertendetails bleiben mit verständlicher Erläuterung erreichbar."* Jede
Stelle, an der ein Mensch eine Kennung von einem anderen Bildschirm abtippte,
mit dem, was heute an ihrer Stelle steht:

| Stelle | Getippt wurde | Jetzt | Für Fachleute |
| --- | --- | --- | --- |
| Handel — Vorgang, Zusagen, Ausgang, Buchung | die Kennung des Handels, viermal | eine Wahl der Handel mit diesem Gegenüber; ein neuer bekommt seine Kennung von der Leitung | „Einen Handel über seine Kennung öffnen" |
| Handel — Zusagen | Kennung und Fassung einer Besitzzeile | die eigenen Exemplare des Drucks zur Wahl | die Felder, mit Erklärung |
| Handel — Ausgang | die Kennung einer Nachricht | ein Knopf je ausstehender Nachricht | die Felder, mit Erklärung |
| An einem fremden Tisch mitspielen | die Peerkennung des Wirts **und** die Kennung seiner Partie | eine Wahl der Gegenüber, bei denen mitgespielt werden darf; **die Partie nennt der Wirt mit seinem ersten Wort** (`service.Partieerfragen`) | „Die Partie selbst nennen" — für einen Wirt vor 0.6.0 |
| Ein Gegenüber aufnehmen | die Peerkennung | „Wer in der Nähe ist" nennt, wer sich im selben Netz ankündigt, mit „Als Kontakt aufnehmen" je Eintrag | die Felder für Kennung und Adresse, darunter |
| Ein Deck verleihen | die Peerkennung des Entleihers, die Kennung des Decks und die einer Partie | eine Wahl der bestätigten Gegenüber und eine der eigenen Decks | „Nur für eine bestimmte Partie verleihen" — ohne Angabe gilt die nächste |
| Ein Deck aus einem anderen Profil holen | die Kennungen von Profil und Deck | eine Wahl der fremden Profile und ihrer Decks — **hinter derselben Erlaubnis wie das Holen**, geprüft, bevor gelesen wird (`service.Deckuebernahme.DeckeZurUebernahme`, [0071](entscheidungen/0071-woher-ein-deck-kommt.md)) | „Kennungen selbst nennen" |

**Was keine Kennung ist, bleibt ein Feld**: ein Name, eine Menge, eine
Frist, ein Satz. **Und ein Pfad** gehört in einen nativen Dialog (UX-08.AK-03),
nicht in diese Tafel.

---

## Zustände — was jeder zeigt

**UX-08.AK-02** (#1680):
*„Ladezustand, Fortschritt, Abbruch, Wiederholen, Konflikt, Erfolg und
Hintergrundauftrag werden sichtbar und konsistent dargestellt; ein Toast ist
nicht der einzige Nachweis einer wichtigen Zustandsänderung."* **Die ersten
sieben Zeilen folgen dem Artboard *BelegZustaende*** („Sieben Zustände je
Seite"), die übrigen fünf nennt das Kriterium dazu. Je Zustand ein Aussehen
und ein Baustein, der es trägt:

| Zustand | Wie er aussieht | Baustein | Hier zu sehen |
| --- | --- | --- | --- |
| Laden | „Wird geladen …" als `status` — **nie** ein leerer Bestand, solange die Antwort aussteht (UX-08.NAT-01) | `bausteine/Ladehinweise.tsx` | Decks, Turniere, Spielen, Handeln, Vorlagen, Symbol |
| Leer | ein Satz, **was** fehlt, und der nächste Schritt als Knopf | `bausteine/Leerzustand.tsx` | die Karten eines neuen Sets, die Nutzerliste, die Sammlung (seit UXR-10: *Ein Exemplar aufnehmen*, *Eine Liste einfügen*), *Eine Partie anfangen* ohne Deck (seit UXR-13: *Zu den Decks*) |
| Gefiltert leer | etwas ist da, nur nicht in dieser Sicht — und der Knopf nimmt die Auswahl zurück | `bausteine/Partienliste.tsx` | die Partienliste, die Sammlung |
| Fehler | der Satz für Menschen und „Erneut laden" — **kein** Leerhinweis (S15.NAT-02) | `bausteine/Ladehinweise.tsx` | wie beim Laden |
| Ohne Netz | der letzte gute Stand, und wie alt er ist; das Bearbeiten läuft weiter (§16, Regel 5) | `bausteine/Portfolio.tsx` | die Preise der Sammlung |
| Nur lesen | wessen Sicht es ist; was sich nicht ändern lässt, sagt es — **und jeder Knopf, der ins Profil schreibt, steht gesperrt da** (`aria-disabled`, fokussierbar, der Grund als Beschreibung), statt erst am Tor zu scheitern (G10, #2276) — **auch eine Menüzeile** (`schreibt` am `Menueeintrag`, #2288) **und eine Eingabe, die ohne Knopf schreibt**: ein Textfeld ist dann schreibgeschützt, eine Auswahl gesperrt (#2290). Welcher schreibt, entscheidet der Träger in `docs/AUTORISIERUNG.md`: Pakete gehören dem Rechner, Nutzer dem Menschen; die Sperre trägt `haken/schreibrecht.tsx` | `bausteine/Nurlesehinweis.tsx` | ein Profil, das man nur ansieht |
| Teilerfolg | was geschrieben ist, und getrennt, was nicht (S08.NAT-03, Tafel unten) | `bausteine/Teilerfolg.tsx` | ein Deck aus einem anderen Profil |
| Fortschritt | ein Satz, wie weit es ist, und ein Balken mit `aria-valuenow` — als Live-Region, damit ein Vorleser mitkommt | `bausteine/Fortschritt.tsx` | die Bildausgabe |
| Abbruch | ein Knopf am laufenden Vorgang; was bis dahin geschah, bleibt und wird genannt | `bausteine/Fortschritt.tsx` | die Bildausgabe („Anhalten"), der Suchindex, ein KI-Bild |
| Konflikt | „Der Eintrag wurde zwischenzeitlich geändert" am Datensatz, und **„Neu laden"** holt den heutigen Stand; die eigene Eingabe steht bis dahin da | `bausteine/Hinweiskasten.tsx` | Formate, Decks, Sammlung, Vorlagen |
| Erfolg | am Objekt, das sich geändert hat — die Liste zeigt den neuen Stand; wo es kein Objekt gibt, das Meldungsband | `rahmen/Meldungsband.tsx` | ⌘Z und ⌘S im ganzen Fenster, das Sichern eines Sets |
| Hintergrundauftrag | Stand, Fortschritt und Dauer, solange er läuft, dazu die letzten fünf beendeten | `seiten/Hintergrundaufgaben.tsx` | der Start |

**Es gibt keinen Toast.** Nichts verschwindet von selbst: das Meldungsband
bleibt, bis es jemand schließt, und eine Meldung am Objekt, bis der nächste
Vorgang sie ablöst. **Sieben Dateien stellen eine Uhr**, und keine räumt
damit eine Meldung weg — der Wächter nennt sie mit ihrem Zweck.

**`frontend/src/zustandskatalog.test.ts` hält die Tafel gegen den
Quelltext**: jeder genannte Baustein liegt da und wird benutzt, keine
Bibliothek für Toasts steht in den Abhängigkeiten, und eine neue Uhr
verlangt eine Entscheidung. **Was er nicht hält**: dass jede der rund
hundertsechzig Stellen, die einen Fehler zeigen, schon den Baustein nimmt —
ein Konflikt bietet „Neu laden" an den vier genannten Stellen an, und eine
fünfte bekommt es, wenn sie ihn meldet.

### Ein Teilerfolg

**S08.NAT-03**: *„Teilerfolg benennt geschriebene Daten und verbleibenden
Fehler getrennt."* Jede Stelle, an der ein Vorgang zum Teil gelingt — **was
geschrieben ist, steht als `status` da, was offen bleibt, darunter für
sich**:

| Vorgang | Geschrieben | Was offen bleibt | Wo |
| --- | --- | --- | --- |
| Eine Setdatei einführen, abgebrochen | die Sets, die schon im Bestand stehen — ihre Zahl | der Fehler, an dem es abbrach, als `alert` | `seiten/Sets.tsx` |
| Ein Set öffnen | das Set, mit der Zahl seiner Karten | die Stellen, die sich nicht deuten ließen | `seiten/Sets.tsx` |
| Ein Set löschen | das Löschen | was dabei liegen blieb | `seiten/Setloeschen.tsx` |
| Eine Liste in die Sammlung einlesen | die angelegten Exemplare — ihre Zahl | die Zeilen mit Befund | `bausteine/Listeneinfuhr.tsx` |
| Ein Deck aus einem anderen Profil holen | das Deck, mit Namen und Kartenzahl | was nicht mitkam | `bausteine/Deckuebernahme.tsx` über `bausteine/Teilerfolg.tsx` |

**Die letzte Zeile stand bis zum 2026-09-25 nur zur Hälfte da**: gezeigt
wurde, was nicht mitkam, und ob das Deck angekommen war, sagte erst die
Liste darüber.

---

## Aktionen — was jede überall bedeutet

**UX-10.AK-01** (#1682),
nach dem Artboard *BelegMuster*. Der Katalog steht in
`frontend/src/aktionen.ts` (Symbol, Kürzel, Bestätigung) und unter
`bausteine.aktionskatalog` in beiden Sprachen (Kurzlabel, Bedeutung). Jeder
Knopf einer dieser Aktionen ist ein `Aktionsknopf`: das Symbol stumm vor der
Aufschrift, die Bedeutung als Tooltip, das Kürzel nur, wo es wirkt.

| Aktion | Symbol | Kurzlabel | Bedeutung (Tooltip) | Kürzel | Wo das Kürzel angesagt wird | Bestätigung |
| --- | --- | --- | --- | --- | --- | --- |
| anlegen | + | Anlegen | Legt etwas Neues an. | ⌘N | nirgends — ⌘N öffnet das Formular für ein neues Set und legt keines an | keine |
| bearbeiten | ✎ | Bearbeiten | Ändert es an Ort und Stelle. | ↵ | nirgends | keine |
| sichern | ⤓ | Sichern | Hält den Stand fest. | ⌘S | nirgends — ⌘S sichert das **offene** Set, nicht die Zeile der Settabelle | keine |
| zuruecknehmen | ⎌ | Zurücknehmen | Nimmt den letzten Zug zurück. | ⌘Z | Symboleditor | keine |
| duplizieren | ⧉ | Duplizieren | Legt ein zweites daneben. | ⌘D | nirgends | keine |
| kopieren | ⎘ | Kopieren | Legt es in die Zwischenablage. | ⌘C | Menü *Diese Karte* (bis UXR-06 der Knopf in der Kartenablage) | keine |
| entfernen | − | Entfernen | Nimmt es aus dieser Ansicht; anderswo bleibt es. | ⌫ | nirgends | Rücknahme |
| loeschen | ✕ | Löschen | Löscht es endgültig, überall. | ⇧⌫ | nirgends | Rückfrage |

**„Wichtige oder mehrdeutige Aktionen bleiben beschriftet"** — hier bleibt
**jede** beschriftet: das Symbol steht zusätzlich, nie allein. Wo die
Aufschrift die Sache nicht nennt und der Knopf mehrfach dasteht, nennt sie
der zugängliche Name (`bezeichnung`, etwa „Sichern — Mein Probeset").

**Drei Wörter sind dabei gerade gezogen worden**: „Partie kopieren" legte
ein zweites daneben und heißt jetzt **Duplizieren** — „Kopieren" ist die
Zwischenablage, seit Phase 3.7 schon bei der Karte. **„Speichern" heißt
Sichern**, wie in der Menüleiste: bei Formaten, Vorlagendateien und den
Angaben eines Exemplars.

### Was vor einer Löschung steht

**42 Bindungen** löschen, entfernen, widerrufen, vergessen oder setzen
zurück (gezählt am 2026-10-10, mit den Kontakten aus einem Adressbuch, den Symbolen der Bibliothek, den eigenen Übersetzungen, dem Widerrufen einer Schnittstelle und dem aller auf einmal). Jede steht in `frontend/src/test/aktionstafel.ts` mit ihrer
Bestätigung, und `frontend/src/aktionskatalog.test.ts` prüft sie am Ort —
und die Zahlen hier gegen die Tafel:

| Bestätigung | Zahl | Beleg im Lauf |
| --- | --- | --- |
| Rückfrage an der Klickstelle | 22 | eine `Entscheidung` oder ein Ja in der Zeile, **deren Bestätigung nach dem Löschen fragt** |
| getippter Name | 2 | Set löschen, Bestand zurücksetzen — das Feld dafür steht am Ort |
| Rücknahme mit ⌘Z | 2 | Schlüsselwort (Verlauf des Sets), Punkt einer Zeichnung (Zeichenverlauf) |
| keine, mit Grund | 16 | ein einzelner kurzer Eintrag, den dieselbe Stelle neu anlegt; ein erzeugtes Ding; ein Entwurf |

**Bis zum 2026-09-25 fragten acht nicht, die es hätten tun sollen** —
Zugang, Anbietergeheimnis, Token, zweiter Faktor, Nutzer, Profil,
Partieprotokoll und das Zurücksetzen der eigenen Autoersetzungen. Der
Durchgang vom 2026-09-21 hatte die übrigen „geprüft, nicht angenommen";
geprüft war eine Liste, die kein Lauf hielt. Beim Token stand sogar eine
Rückfrage in der Zeile — vor dem **Erneuern**. Deshalb verlangt der Lauf,
dass die Frage nach dem Löschen fragt.

**Was der Lauf nicht hält**: ob das Wort zur Handlung passt, und ob die
Frage vor genau diesem Ruf steht, wenn eine Datei zwei Löschwege hat. Das
sehen die Tests der Seiten und ein Mensch.

### Was eine Rücknahme umfasst

**S08.NAT-04** (#1680):
*„Rücknahme umfasst den fachlichen Zug und ist per Tastatur erreichbar."*

**Zurückgenommen wird, was ein Verlauf hält — und das ist das Set.** Ein
Zug ist, was **ein** Befehl geschrieben hat: ein verlassenes Feld, eine
duplizierte Karte samt aller Felder, eine Änderung an mehreren Karten
zugleich (`apikarten.ts`: „ein ⌘Z holt alle zusammen zurück"). ⌘Z nimmt ihn
ganz zurück, ⇧⌘Z stellt ihn wieder her — ebenso über „Bearbeiten" in der
Menüleiste und je Schritt auf der Lasche *Verlauf*. **Im Textfeld gehört ⌘Z
dem Feld** (`rahmen/tastatur.ts`): was gerade getippt wird, steht noch in
keinem Verlauf, und eine Rücknahme träfe einen früheren Zug. Eine Zeichnung,
an der gearbeitet wird, hat ihren eigenen Verlauf und geht vor
(`ruecknahmeweg.ts`).

**Außerhalb eines Sets gibt es keinen Verlauf** — Decks, Turniere, Pakete,
Nutzer und Sammlung schreiben in den Bestand, und was dort weggenommen
wird, **fragt vorher** (die Tafel darüber). Belegt ist der Weg über
`anwendungruecknahme.test.tsx` und die Tastenlage in `rahmen/tastatur.ts`
samt ihrem Lauf; **ein Drehbuch in WKWebView hing beim ⌘Z** — die Probe
nimmt die Taste in ihrem eigenen Fenster, und das ist eine Frage an die
Probe, nicht an diese Anwendung.

### Sammlung und Deck — ein Muster je Vorgang

**DOM-03.AK-03** (#1698):
*einheitliche Rücknahme- und Bestätigungsmuster und kontrollierte
Mengenänderungen* für Massenänderung, Bereichswechsel, Filter, Sortierung
und Ausfuhr. **Die Regel dahinter ist die der Löschtafel**: gefragt wird,
wo sich ein Verlust nicht in einem Schritt an derselben Stelle beheben
lässt; wo er es lässt, steht der Schritt zurück **dort, wo der Vorgang
stand**.

| Vorgang | Muster | Wo | Warum so |
| --- | --- | --- | --- |
| Lagerort für viele Zeilen | *Zurücknehmen* im Bericht | `bausteine/Massenvorgaenge.tsx` | Ein Lagerort überschreibt viele, und die Sammlung hat keinen Verlauf. Der Bericht nennt je Zeile den Stand **vorher** samt der Fassung danach (`service.Massenbericht.Vorher`); die Rücknahme schreibt jeden alten Lagerort **gegen diese Fassung** zurück — was seither jemand anderes schrieb, bleibt und steht als Befund da (§7, Regel 2) |
| Viele Zeilen löschen | Rückfrage | dito | nicht in einem Schritt zu beheben (Löschtafel oben) |
| Hauptdeck ↔ Sideboard | keine Bestätigung | `bausteine/Deckplaetze.tsx` | derselbe Knopf nimmt den Wechsel zurück, und nichts geht verloren |
| Menge im Deck | ± in Einerschritten, **„−" ruht bei eins** | dito | bis zum 2026-09-27 nahm „eine weniger" bei eins den Platz still heraus — eine Löschung in der Gestalt einer Zählung. Herausnehmen ist der eigene Knopf daneben, mit dem Namen der Karte |
| Menge in der Sammlung, derselbe Druck zweimal | Entscheidung *Auf N erhöhen* / *Zur vorhandenen Zeile* / *Abbrechen* | `bausteine/Mengenentscheidung.tsx` | S04.NAT-01 bis -03: Abbrechen ändert keine Menge |
| Einlesen über Vorhandenes | Vorschau *vorher → nachher*, *addieren* oder *auslassen* | `bausteine/Einfuhrvorschau.tsx` | AK-01: vor dem Schreiben, nicht danach |
| Suche, Ausführung, Marken | *Suche und Filter zurücksetzen*, sobald die Sicht etwas wegnimmt — **mit Treffern und ohne** | `bausteine/Sammlungssicht.tsx`, `seiten/Sammlung.tsx` | „N von M" sagt, dass etwas fehlt; der Knopf nimmt es in einem Schritt zurück, statt dass jemand erst ausprobiert, welches Feld er leeren muss |
| Sortierung | keine | Leiste über der Liste (`bausteine/Sammlungsleiste.tsx`, seit UXR-10) | sie nimmt nichts weg und schreibt nichts; dieselbe Wahl stellt sie um |
| Ausfuhr als CSV | keine | Schublade *Einlesen und ausgeben* (`bausteine/Sammlungswerkzeuge.tsx`, seit UXR-10) | sie schreibt nicht in den Bestand. Was **zurück**gelesen wird, geht durch die Vorschau |

**An jeder Fläche dieselbe Rücknahme**: `vorher` steht im Bericht von
`mse masse-aendern --json`, `schreiben.sammlung.masse` und
`POST /sammlung/masse`; zurückgenommen wird mit **derselben** Änderung und
den Angaben daraus (`internal/flaechen/massenruecknahme_test.go`). Das
Fenster bietet von den Massenfeldern nur den Lagerort an und nimmt nur ihn
zurück; wer über die anderen Flächen Menge, Zustand oder Wunsch setzt,
findet deren alten Stand im selben Bericht.

**Was die Tafel nicht hält**: ob ein neuer Vorgang das richtige Muster
bekommt. Es gibt keinen Lauf über sie; die Tests stehen je Zeile am Ort.

---

## Ohne Maus — UX-10.AK-02

**UX-10.AK-02** (#1682) —
*„Alle Workflows sind ohne Maus bedienbar; Fokus sichtbar, Reihenfolge
sinnvoll, Statusänderungen zugänglich und Farben nicht allein
bedeutungstragend."*

| Teil | Wie | Beleg |
| --- | --- | --- |
| **Alle Arbeitswege** | Jeder der sieben Wege hat einen Einstieg über die Tastatur — ⌘1 bis ⌘6, ⌘', ⌘, und ⌘K —, und was darin geschieht, sind Knöpfe, Felder und Laschen. **Keine Maushandlung an einem Element ohne Rolle**: `eslint-plugin-jsx-a11y` weist sie ab, und die zwei Ausnahmen (`Entscheidung`, `Tischflaeche`) horchen auf Tasten. **Wo gezogen wird, gibt es eine Taste**: Bildzuschnitt (Links, Oben, Breite, Höhe als Zahlenfelder — wie MSEs `ImageSliceWindow`), Symbolfläche (Pfeiltasten), Tisch (Pfeiltasten und Befehlsfeld) | WebKit-Drehbuch `ux10-tastatur` — **kein einziger Klick**: Set öffnen und ein Feld schreiben, Deck aufschlagen, Partie anfangen, in Sammlung und Paketen suchen, ein Gegenüber aufschlagen, die Laschen der Nutzerseite; dazu `ux07-palette`, `ux07-handelsassistent`, `ux07-einstellungen`, `ux07-baum` |
| **Fokus sichtbar** | an **jedem** Halt gemessen: Umriss oder Schatten am fokussierten Element, berechnet und nicht angenommen | `ux10-tastatur`: die Liste „ohne Ring" ist leer |
| **Reihenfolge sinnvoll** | Sprunglink, Leistenknopf, Profil, die Arbeitsbereiche als **ein** Halt (Pfeiltasten darin), die Palette — dann der Inhalt in Lesereihenfolge, im Arbeitsplatz spaltenweise: Navigator, Mitte, Inspector | die Folge der Halte steht im Bericht des Drehbuchs |
| **Statusänderungen zugänglich** | jede Meldung aus dem Zustandskatalog steht in `role="status"` oder `role="alert"`; nach ⇥ aus einem Kartenfeld sagt der Speicherstand, dass geschrieben ist; die Zahl der Treffer steht als Status da — in Paketliste, Kartenliste, Sammlung und Textsuche | `ux10-tastatur` („1 von 5"), `sammlungssichtregler.test.tsx`, `textsuche.test.tsx` |
| **Farbe trägt nie allein** | Es gibt **eine** Zustandsfarbe, `warnung`, und sie steht nur an Text, der die Meldung selbst sagt — gezählt am 2026-09-25, keine an einem Rand oder einer Fläche. Die aktive Seite trägt `aria-current` **und** eine Leiste; ein gedrückter Knopf `aria-pressed` **und** eine andere Aufschrift; am Tisch stehen neben den gezeichneten Marken die Ereignisliste und der zugängliche Spiegel (PLAY-04) | diese Zählung; `leistenmarkierung.test.tsx`, `partnerwahl.test.tsx` |

### Was die Messung gefunden hat

Das Drehbuch ist kein Nachweis über einen fertigen Stand gewesen, sondern hat
**vier Fehler** gefunden, bei grüner Pipeline und voller Abdeckung:

1. **Nach ⌘2 und ⌘6 bewegte Tab den Fokus gar nicht.** Der Fokus stand auf
   `body`, und WebKit suchte von einem Knoten der vorigen Seite aus weiter,
   den der Wechsel abgeräumt oder — bei einer aufbewahrten Seite — verborgen
   hatte. Jetzt setzt `haken/useFokusNachSeitenwechsel.ts` ihn auf den Inhalt,
   wenn er nirgends mehr steht, und `main` nimmt ihn (`tabIndex={-1}`). **Der
   Sprunglink führt seitdem wirklich dorthin**: WebKit bewegt den Fokus bei
   einem Sprung zu einem Ziel, das keinen nehmen kann, nicht.
2. **Nach „Bo aufschlagen" stand der Fokus auf `body`** — aus dem Knopf wurde
   ein Text. Jetzt bleibt der Knopf, gedrückt (`Partnerwahl.tsx`).
3. **Die Zahl der Treffer in der Sammlung und in der Textsuche wurde nicht
   angesagt** — die Paketliste und die Kartenliste taten es schon.
4. **Der Tisch ließ Karten fliegen, obwohl „Bewegung: Immer wenig" gewählt
   war** — er fragte nur das System (`tisch/bewegung.ts`, unter AK-03).

## Größer, kontrastreicher, ruhiger — UX-10.AK-03

**UX-10.AK-03** — *„Vergrößerte Schrift, hoher Kontrast, reduzierte Bewegung
sowie eine tatsächliche VoiceOver-Stichprobe im nativen Fenster sind
dokumentiert; automatisierte Accessibility-Checks allein zählen nicht als
vollständige Barrierefreiheit."*

| Einstellung | Was sie tut | Beleg |
| --- | --- | --- |
| **Schriftgröße** | vier Stufen, 87,5 % bis 125 % der Wurzelschrift; alles in `rem` skaliert mit, Abstände und Fokusringe eingeschlossen. Wird es eng, weicht der Inspector aus („Kein Platz für den Inspector") | `erscheinungsbild.test.ts`; WebKit-Drehbuch `ux10-darstellung`: bei „Sehr groß" ist auf fünf Seiten kein Knopf abgeschnitten |
| **Hoher Kontrast** | nimmt die Faser **und** hebt die Farben: `leise` und `kante` werden Tinte, im hellen Bild `marke` und `warnung` dunkler. **Jeder Text in den Merkmalen hält 7 : 1** (WCAG 1.4.6); die farbigen Meldungen halten 4,5 : 1 und nicht überall 7 : 1 (Zweitblick über Phase 6.3, M7). **Bis zum 2026-09-25 war es nur die Faser**, und vier Paare lagen knapp darunter | `ux10-darstellung` misst jedes sichtbare Textelement gegen seinen wirksamen Grund, hell und dunkel; die Werte in [`KONTRASTE.md`](KONTRASTE.md#hoher-kontrast--2026-09-25) |
| **Weniger Bewegung** | Die Oberfläche hat keine Übergänge — Mantines stehen auf `duration: 0` —, und `barrierefreiheit.css` schaltet jeden verbleibenden ab. **Der Tisch folgt ihr seit dem 2026-09-25**: vorher fragte er nur `prefers-reduced-motion` | `ux10-darstellung` (das Attribut am Wurzelelement), `tisch/bewegung.test.ts`, `tischaussehen.test.tsx` |
| **VoiceOver** | **Handprüfung** — die Probe liest den Baum, aus dem VoiceOver liest (`ux06-vorleser`, `ux10-vorleser`, `play04-vorleser`), aber nicht, was VoiceOver daraus sagt | [`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md) |

**Was davon eine Systemeinstellung auslöst**: `prefers-reduced-motion`
dieselbe Regel wie die Einstellung, `prefers-contrast: more` dieselben Farben.
Gemessen ist der Weg über die Einstellung; der über das System hat dieselben
Werte und ist nicht eigens gemessen — die Probe kann die Medienabfrage nicht
setzen.

## Leere Zustände

**Fünf Seiten benutzen den Baustein `Leerzustand`**, alle ohne Handlungsknopf
im Kasten: der Setarbeitsplatz ohne Set, die Kartenlasche ohne Karte (zweimal:
keine Karte im Set, keine Karte gewählt), die Setangaben ohne Felder, die
Nutzerliste ohne zweiten Nutzer. **Rund dreißig weitere** stehen als freier
Absatz, teils mit einem Knopf daneben — die Liste je Seite steht in der
Inventur, die diese Datei speist, und jeder Absatz sagt, was zu tun ist (§6),
mit einer Ausnahme:

> **Setangaben ohne Felder** nennt die Handlung („wähle ein Spiel, das welche
> mitbringt") und bietet sie nicht an. **Befund**, behoben in U3.

**Zielzustand**: der Baustein bekommt in U2 seine Handlung als Kind, und U11
verlangt für jeden leeren Bereich der Übersicht eine ausführbare Handlung.

---

## Fehlerwege

Drei Muster, und jedes hat seinen Ort:

| Muster | Ort | Wo es heute steht | Zielzustand |
| --- | --- | --- | --- |
| **Ladefeld** — laden, Fehler, leer | jede Seite, die liest | 14 Seiten | bleibt; bekommt den Callout aus U2 |
| **Feldfehler am Feld** — `role="alert"` unter dem Feld, `aria-describedby` | Setangaben, Karteneditor, Paketanlage, Nutzeranlage, Profilanlage; `Pflichtfeld`, `Zahlenfeld` | fünf Seiten und zwei Bausteine | **Field** (U3): ein Baustein statt fünf Formen, angeschlossen an Fehlercode, Schlüssel und Feldbezug; die offenen Zeilen aus `FELDPRUEFUNG.md` entschieden |
| **Meldungsband** — `role="status"`, `aria-live="polite"` | der Rahmen, für alles, was zu keiner Seite gehört; Reihenfolge in `bandmeldung.ts` | ein Band, ein Fassungsband | bleibt; Fortschritt, Abbruch und Erfolg mit Live-Region (U3) |

**Ein Fehler erreicht die Oberfläche nie roh** (§8): `meldungAus` übersetzt
Code und Schlüssel, `verursachendesFeld` ordnet ihn dem Feld zu. Was die
Maschinenclients lesen, ändert U3 nicht — die Pflichtfelder der Antwort
bleiben.

---

## Der Messbestand

**Reproduzierbar, nicht der Bestand dieses Rechners.** Ein verlegtes Zuhause,
gelegt von einem Skript, gemessen von den Skripten daneben:

```bash
node scripts/sammlungssaat.mjs /tmp/mse51heim 10000   # einmal
HOME=/tmp/mse51heim make dev                          # in einem eigenen Fenster
node scripts/arbeitswegmass.mjs                       # die vier Wege unten
node scripts/kernwegmass.mjs http://localhost:34115 probe probe-schlicht
node scripts/serienmass.mjs                           # jede Seite
node scripts/serienabzug.mjs <verzeichnis>            # die Vorherbilder
```

| | |
| --- | --- |
| **Exemplare** | **10.000** in der Sammlung — der Messbestand, den U8 verlangt; zehn Kartennamen, fünfzig Setcodes, jede vierte Zeile mit Menge 4, jede siebte Foil |
| **Sets** | eines: „Messset 5.1", Spiel `probe`, Stil `probe-schlicht`, **fünf** Karten mit Name und zweizeiligem Text |
| **Pakete** | die **sieben** der Saat (`saat/`): ein Spiel, fünf Stile, eine Exportvorlage |
| **Nutzer** | einer mit einem Profil; die Wege legen an, was sie darüber hinaus brauchen, und räumen es weg |
| **Fenster** | 1.280 × 800, ein Chrome ohne Fenster über CDP an `wails dev` (seit ENG-02 `make dev-browser`) — echte Tastendrücke durch die laufende Brücke |
| **Was fehlt** | ein Beispielset: der Entwicklungsbau bringt keines mit. Der erste Weg („von der leeren Anwendung zur ersten Karte") ist deshalb weiter die Zahl aus `SERIENARBEIT.md` vom 2026-09-03, gegen die Saat der Auslieferung |

**Was eine Nutzerhandlung ist**: ein Klick, ein Tastendruck, ein Tippen —
gleich wie lang. **Tabstelle**: die wievielte erreichbare Stelle das erste
Bedienelement des Weges auf seiner Seite ist; die Seitenleiste mit Sprunglink
und zehn Einträgen liegt davor, also beginnt jede Seite bei **12**. **Ladezeit**:
vom Kürzel bis die Seite dasteht, auf diesem Rechner, in Schritten von 50 ms —
eine Aussage über diesen Rechner (0007), keine Zusage.

**Seit U14 (#1490)
lässt sich die Seitenleiste einklappen** — Knopf an ihrem Rand, ⌥⌘S wie im
Finder, Eintrag unter *Darstellung*; der Zustand ist eine Einstellung auf
Geräteebene und bleibt über den Neustart. Eingeklappt liegen vor dem Inhalt
nur noch der Sprunglink und der Knopf, und die Breite der Leiste gehört dem
Arbeitsplatz. `MSE_SEITENLEISTE=zu` fährt `kernwegmass.mjs` und
`arbeitswegmass.mjs` so und nennt die Breite des Hauptbereichs vorher und
nachher; die Tabstellen in den Tabellen oben gelten für die **stehende**
Leiste. **Der Knopf ist eine Tabstelle mehr** vor dem Inhalt — gemessen am
2026-09-09: `paket` beginnt bei der 23. statt der 22. —, bis
#1488 die
elf Einträge der Leiste zu einer Tabstelle macht.
**Eingeklappt gemessen am 2026-09-09** (`MSE_SEITENLEISTE=zu`, 1280 × 800):
Hauptbereich **999 → 1239** Punkte; `paket` beginnt bei der **11.** Tabstelle,
`sammlung` bei der **13.**, `profil` bei der **10.** — die Ziele „≤ 14" sind
mit eingeklappter Leiste erreicht, mit stehender erst durch #1488.
**Der Kernweg `karte` braucht das Paket `magic-m15` im Zuhause**: der Lauf
legt sein Messset damit an, und im Messzuhause `/tmp/mse51heim` fehlt es —
die Karte hat dann keine Felder, und der Lauf wartet vergebens auf das
Namensfeld. Das sah am 2026-09-09 wie ein Fehler des Arbeitsplatzes aus
(#1497,
entkräftet: die Sonde gegen das Messset zeigt die neue Karte offen, Zeiger
„8 von 8"). `MSE_ABZUG_BEI_FEHLER=<pfad>` legt seitdem beim Timeout ein Bild
ab. Der Ausgangswert oben stammt aus einem Zuhause mit dem Paket.

---

## Die Kernwege, gemessen

**Sieben Wege, sieben Zeilen, jede mit Ausgangswert und Ziel** — die Abnahme
von #1412 verlangt beides, bevor ein Arbeitsplatz umgebaut wird. Der Ausgangswert
ist vom **2026-09-08**; das Ziel gilt für das Werkstück, das den Weg umbaut, und
der Messlauf senkt seine Erwartung mit.

| Kernweg | Ausgangsbestand | Ausgangswert · Handlungen · Tabstelle · Ladezeit | Messgerät | Ziel |
| --- | --- | --- | --- | --- |
| `karte` | Messset mit fünf Karten, Kartenlasche offen | **Mit AP-9 (2026-10-17, #2366) eine Tabstelle mehr**: die Leiste am Setkopf ist eine `toolbar` — „Karte anlegen“ die 18., das Namensfeld die 39., gerechnet und nicht nachgemessen · **4** Handlungen für „anlegen, füllen, weiter", kein Seitenwechsel · „Karte anlegen" ist die **17.** Tabstelle, das Namensfeld war die **57.** und ist mit U5 die **38.** (2026-09-09, 1.440 × 900, Navigator offen; die Rechnung steht unten) · Ladezeit der Kartenlasche siehe Serienmaß · **7** Handlungen von der leeren Anwendung zur ersten Karte (2026-09-03, gegen die Saat der Auslieferung) · Vorschau nach dem Verlassen des Felds **286 bis 398 ms** (2026-09-09, ruhige Maschine, drei Läufe à drei Schreibvorgänge: Text einsetzen, ⇥, warten, bis ein neues Bild steht); davon Rastern bei 100 % **207 bis 337 ms** (`Vorschau` über die Brücke), Schreiben **30 bis 90 ms** — vorher 1,2 bis 1,5 s je Änderung nach der ersten, siehe `0049_volltext_zeilennummer.sql` | `kernwegmass.mjs`, `erstlaufmass.mjs` | ≤ 4 Handlungen bleibt (**gehalten**); Namensfeld ≤ 14 war gegen eine Basis von 12 gerechnet, die es auf einer Setseite nicht gibt — **≤ 38 bei offenem Navigator** ist der gemessene Stand von U5, und wer ihn senken will, nimmt die Werkzeuge der Lasche in eine Werkzeugleiste (eine Stelle); Vorschau nach dem Verlassen des Felds ≤ 300 ms — **nicht erreicht**: der Rasterer allein liegt mit 207 bis 337 ms an der Grenze, das Fenster hat daran keinen Anteil mehr; wer es senken will, setzt am Rasterer an (Phase 5.5, Texturen und Cache) |
| `paket` | die sieben Saatpakete; der Weg legt ein Stilpaket an | **6** Handlungen (⌘5, Paket wählen, Datei, in den Text, tippen, Speichern) · **19.** Tabstelle · **54 ms** — **mit U7 (2026-09-09): 5 Handlungen** (⌘5, Datei im Baum, in den Text, tippen, Speichern) · **22.** Tabstelle · **60 ms** (unter Last, U6-Pipeline daneben). Der Baum ist **eine** Tabstelle (#1290); vor ihm liegen elf Einträge der Seitenleiste, der Profilwechsel, zwei Regionsknöpfe und das Suchfeld — beim ersten Umbau war die Datei die **30.**, weil jeder Paketknopf ein Halt war — **mit #1488 (2026-09-09): 9. Tabstelle** (die Seitenleiste ist eine Tabstelle; Ziel ≤ 14 erreicht)** — **nachgemessen am 2026-09-10: 5 Handlungen · 11. Tabstelle · 54 ms**. Der Ausdruck des Wächters traf dabei zuerst den Halt der **Seitenleiste** statt den des Paketbaums und gab −1, weil eine eingeklappte Leiste `hidden` ist: eine Aussage über die Leiste, nicht über den Kernweg (§10). Er zeigt jetzt auf `nav[data-baum]` — **nachgemessen am 2026-09-10: 3 Handlungen · 10. Tabstelle · 212 ms** | `arbeitswegmass.mjs` | ≤ 5 Handlungen **erreicht**; ≤ 14. Tabstelle **erreicht** — die 9. Stelle mit eingeklappter Seitenleiste (#1488), die **11.** mit ausgeklappter (nachgemessen 2026-09-10). Hier stand bis dahin „nicht erreicht“, während die Spalte links die 9. Stelle schon maß; der Zweitblick über Phase 5.1 hat den Widerspruch angestrichen; ≤ 200 ms erreicht (54 ms) |
| `sammlung` | 10.000 Exemplare | **3** Handlungen (⌘6, ins Suchfeld, tippen) · **24.** Tabstelle · **185 bis 361 ms** bis die Liste steht (drei Läufe), **2 bis 9 ms** von der Eingabe bis zum Ergebnis — **mit #1488 (2026-09-09): 16. Tabstelle** (zwölf Werkzeugknöpfe liegen weiter vor dem Suchfeld)** — **mit U8 (2026-09-09): 3 Handlungen · 10. Tabstelle · 209 bis 371 ms** bis das Gitter steht (drei Läufe, 10.000 Exemplare, 24 Zeilen im DOM), **1 ms** von der Eingabe bis zum Ergebnis — **nachgemessen am 2026-09-10: 3 Handlungen · 10. Tabstelle · 212 ms** — **mit #1500 (2026-09-10): 3 Handlungen · 12. Tabstelle · 175 bis 318 ms** (fünf Läufe, je ein frisch gestarteter Server), **110 ms** von der Eingabe bis zum Ergebnis | `arbeitswegmass.mjs` | 3 **erreicht**; ≤ 13. Tabstelle **erreicht** (12.); ≤ 200 ms bei 10.000 Zeilen **im Entwicklungsmodus nicht durchgängig** — **175 · 175 · 266 · 267 · 318 ms** nach #1500 gegen **265 · 270 · 332 ms** davor, am selben Tag auf demselben Rechner gemessen. Die Zahlen zerfallen in zwei Häufungen im Abstand eines Rufs über die Naht: unter `StrictMode` läuft jeder Ladeeffekt zweimal, und die Zählzeile wartet mal auf die erste Antwort (175 ms), mal auf die zweite (266 ms) — im Bau, der ausgeliefert wird, gibt es den zweiten Ruf nicht. Dazu übersetzt `wails dev` mit `all=-N -l`: derselbe Ruf kostet so 80 ms und optimiert 30 bis 50 ms. Die Herleitung samt Spur steht in [0092](entscheidungen/0092-eine-sicht-rechnet-dort-wo-die-liste-liegt.md) |
| `deck` | Messset mit fünf Karten; der Weg legt ein Deck an | **11** Handlungen (⌘7, Name, tippen, Spiel, ⌘A, tippen, Anlegen, Öffnen, Set, Karte, Ins Deck legen) · **21.** Tabstelle · **2 ms** — die Anlage steht, bevor die Liste geladen ist — **mit #1488 (2026-09-09): 13. Tabstelle** — **nach dem Zweitblick über Phase 5.1 (2026-09-10): 9 Handlungen · 13. Tabstelle · 5 ms**. Der Zweitblick hat angestrichen, dass der Wächter nach U10 noch den **alten** Weg fuhr und bei elf stand (Befund 6). Drei Handlungen sind seitdem weg — das Spiel wird **gewählt** statt angeklickt, markiert und neu getippt (zwei), und das neue Deck **geht von selbst auf** (eine) —, eine ist dazugekommen: der Kasten »Deck anlegen« wird aufgeklappt. Er steht über der Liste, weil das Namensfeld darunter auf der 16. Stelle lag | `arbeitswegmass.mjs` | ≤ 6 Handlungen **nicht erreicht** (9), und die drei übrigen sind benannt: **der Kasten** (er kostet eine Handlung und spart 625 Punkte Seitenhöhe — ohne ihn stand die Kartensuche unter der Falz), **die Wahl des Spiels** (sie fiele weg, wenn das Profil nur eines hätte; vorbelegt ist das des ersten Sets) und **die Wahl des Sets** (sie fiele weg, wenn die Kartenwahl das Set des Arbeitskontexts vorbelegte — den kennt das Fenster heute nicht: `useArbeitskontext` trägt Nutzer, Profil und Rolle, keinen offenen Bestand). ≤ 14. Tabstelle **erreicht** (13); ≤ 200 ms erreicht |
| `profil` | ein Nutzer; der Weg legt ein zweites Profil an | **3** Handlungen (⌘', Reiter „Profile", „Hierhin wechseln") · **21.** Tabstelle · **52 ms** bis zur Seite, **54 bis 85 ms** für den Wechsel samt Neuaufbau — **mit #1488 (2026-09-09): 13. Tabstelle** — **nachgemessen am 2026-09-10: 3 Handlungen · 10. Tabstelle · 57 ms, Wechsel 160 ms** | `arbeitswegmass.mjs` | ≤ 2 Handlungen **nicht erreicht und nicht mehr angestrebt**: U4 hat ausdrücklich entschieden, dass die Kontextleiste den Profilwechsel **nicht** wiederholt, sondern auf die Seite führt („die Leiste wiederholt sie nicht", `Kontextleiste.tsx`) — zwei Knöpfe mit demselben Namen wären für ein Vorleseprogramm einer. Das Ziel stammt aus U1 und ist von U4 überholt; angestrichen vom Zweitblick über Phase 5.1 (Befund 6), der einen Wächter fand, der die alte Erwartung bestätigte. ≤ 3. Tabstelle **nicht erreicht** (10): vor dem Reiter liegen die Seitenleiste als eine Stelle, der Profilwechselknopf der Leiste und die Kopfzeile der Seite; Wechsel ≤ 300 ms erreicht (160 ms) |
| `spielen` | `testdaten/mtgjson/tisch/TSP.json` (sechs erfundene Karten), zwei Decks; für S12/S13 zwei Anwendungen auf diesem Mac | **gemessen am 2026-10-14 (P0, #2336), Bündel `25408dc82266`, 1.440 × 900**: S1 **9** Handlungen mit der Maus, **25** mit der Tastatur (18 Tabstellen) · S2 **6** / **94** · S3 **2** (»Als Land spielen«), **1** gezogen, **54** mit der Tastatur · S4 **7** · S5 **2** / **84** · S6 **4**, **1** gezogen, **111** mit der Tastatur · S7 **4** · S8 **6** · S9 **4** · S10 **1** / **63** · S11 **fehlt** · S12 **10** · S13 **5** und **2** für die Eröffnung des Gastes. Jede Zeile mit Sichtwechseln, Maschinenzeit und beobachteter Wirkung in `belege/ausgangswerte09-spielen.json` und `-netz.json`; Einordnung in `audits/0.9.0/umsetzung/P0-2336.md` | `huellenprobe/ausgangswerte09spielen.mjs`, `ausgangswerte09netz.mjs`; Vertrag `scripts/ausgangswege.mjs` | **Die Ziele von 0.9.0** stehen in Audit C, Abschnitt 3.1, und in den Paketen P4 bis P9: S1 ≤ 3, S2 ≤ 2, S3 eine Geste mit Regelweg, S6 per Tastatur ≤ 3, S12 ≤ 3, S13 ≤ 2. Das alte Ziel dieser Zeile (⌘9, ≤ 14. Tabstelle) ist überholt: ⌘4 öffnet heute *Decks*, und am Tisch liegt das nächste Ziel 50 bis 110 Tabstellen weit |
| `handeln` | Anna und Bert als zwei Bestände auf diesem Mac, die Drucke `09-blitz` und `09-schock` erfunden, Preise von einer Gegenstelle im Prozess | **gemessen am 2026-10-14 (P0), dasselbe Bündel**: H1 **13** Handlungen mit der Maus, **90** mit der Tastatur (79 Tabstellen) · H2 **6** · H3 **3** · H4 **3** · H5 **35** für die Anbietende und **33** für das Gegenüber (H4 darin), **jedes Übertragen mitgezählt** — je einer davon **bedingt**, und der Datensatz nennt ihn · H6 **4**. Datensatz `belege/ausgangswerte09-handeln.json` | `huellenprobe/ausgangswerte09handeln.mjs`; Vertrag `scripts/ausgangswege.mjs` | **Die Ziele von 0.9.0** stehen in Audit C, Abschnitt 3.2, und in P10 bis P13: H1 ≤ 5, H2 ≤ 2, H3 ohne eigenen Weg, H4 ≤ 2, H5 ≤ 3 je Seite, H6 0–1. Audit C schätzte H5 am Code auf 8/7 — gezählt ohne das Übertragen; gemessen sind es mehr als viermal so viele |

**Drei Befunde aus dem Messen**, die kein Test gefunden hatte — jeder ein
Fall für das Werkstück, das die Seite umbaut:

1. **Das Spielfeld der Deckanlage ist mit `magic` vorbelegt**, gleich welches
   Spiel installiert ist. Gegen die Saat heißt das ⌘A und überschreiben — die
   elfte Handlung. Ein Deck ohne bekanntes Spiel („magicprobe") wird dabei
   **angelegt** und weist erst beim Legen einer Karte ab. U10.
2. **Nach einer geglückten Anlage meldet das Namensfeld „Name braucht einen
   Wert."** — das Formular leert sich, das Feld gilt als berührt, und der
   Pflichtfeld-Baustein liest daraus einen Fehler. U3 (Field).
3. **Die Deckliste lädt nach dem Legen einer Karte nicht nach**: sie zeigt
   „0 im Hauptdeck", während das offene Deck darunter vier Messkarten führt.
   U10.

**Warum die Tabstellen so hoch lagen — und seit #1488 nicht mehr.** Die
Seitenleiste war mit Sprunglink elf Stellen, und auf jeder Seite kommen
Seitenkopf und Werkzeuge davor: das Suchfeld der Sammlung war die 24., weil
zwölf Knöpfe der Werkzeuge vor ihm stehen. **Seit dem 2026-09-09 ist die
Leiste eine Tabstelle** (#1488):
ein Tab landet auf der aufgeschlagenen Seite, ↑/↓ wandern, Home/End springen —
derselbe Vertrag wie Paketliste und Vorlagenbaum. Gemessen bei 1280 × 800 mit
stehender Leiste: `paket` **9.**, `deck` **13.**, `profil` **13.**,
`sammlung` **16.** Tabstelle (vorher 23, 23, 23, 26). Das Ziel „≤ 14" heißt:
**das erste Bedienelement des Weges ist das erste im Inhalt** — genau das, was
ein Arbeitsplatz mit Navigator, Zentrum und Inspector leistet, wenn der
Navigator vorn steht — seit U8 auch die Sammlung: ihr Suchfeld ist die
**10.** Tabstelle.

**Die Sammlung fenstert seit U8** (#1419,
Vermerk 0085): bei 10.000 Exemplaren stehen **24** Zeilen im DOM statt
10.001 — gefunden in der Sichtprüfung am 2026-09-09, weil das Fenster der
Liste zuerst keine feste Höhe hatte und `clientHeight` die Höhe aller Zeilen
war (1,3 s). Mit fester Höhe steht das Gitter nach **209 bis 371 ms** (drei
Läufe); das Ziel „≤ 200 ms" ist damit **nicht** erreicht, und was bleibt, ist
nicht das Zeichnen, sondern die Naht: rund 3 MB JSON für 10.000 Zeilen in
einem Ruf. Ein Fenster auf der Go-Seite änderte den Vertrag von `Sammlung`;
das ist #1500
und kein Teil von U8.

**#1500 ist seit dem 2026-09-10 gebaut**
([0092](entscheidungen/0092-eine-sicht-rechnet-dort-wo-die-liste-liegt.md)):
die Sicht rechnet in Go, über die Naht kommen zweihundert Zeilen statt
zehntausend, und die Zeile in der Tabelle oben trägt die neuen Zahlen samt der
Frage, die sie offen lassen — was im Entwicklungsmodus gemessen wird, ist
nicht, was ausgeliefert wird.

**Die Basis ist nachgezählt, und sie ist nicht 12** (2026-09-09, Sichtprüfung
von U5, Stelle für Stelle mit `tabstellen.mjs`): Sprunglink, seit U4 der
Profilwechsel, elf Einträge und die Befehlspalette sind **14**; auf einer
Setseite kommen die Laschenleiste (eine Stelle, Pfeiltasten dazwischen) und
das Laschenfeld dazu — **16**. Die Ziele oben sind gegen 12 gerechnet und
gelten **relativ zur Basis**: „≤ 14" heißt zwei Stellen nach ihr. Für `karte`
reicht das nicht — der Inspector ist nach 0076 die **dritte** Region, hinter
Navigator und Zentrum, und beide tragen Bedienelemente. Das Namensfeld war
die **57.** Stelle; U5 nimmt die neunzehn Feldflächen der Vorschau aus der
Tabfolge (die Maus behält sie) und macht aus der Zoomleiste **eine** Stelle
statt sechs — danach ist es die **38.** bei offenem Navigator. Was davor
bleibt, sind die Werkzeuge der Lasche (4), die zwei Klappknöpfe, Filter und
Gitter des Navigators (2), Spaltenwahl, Blättern, die zwei Handlungsmenüs,
Zoom, drei Überlagerungen und drei Scryfall-Knöpfe.

---

## Referenzgerät, Fenster und die zwei Layoutvarianten

**Entschieden in [Vermerk 0076](entscheidungen/0076-die-ordnung-der-arbeitsplaetze.md)**,
mit beiden Varianten, ihren Kosten und dem, was verworfen wurde. Hier das
Ergebnis:

| | |
| --- | --- |
| **Referenzgerät** | MacBook Air 13″ (M2/M3), 1.470 × 956 Punkte im Standardmaßstab |
| **Messfenster** | 1.280 × 800 — unverändert, damit jede Zahl aus `SERIENARBEIT.md` vergleichbar bleibt |
| **Mindestfenster** | **1.024 × 700** (`internal/fenster/ereignisse.go`, bis ENG-02 `main.go`), vorher 960 × 640: bei eingeklapptem Navigator bleiben einer M15-Karte neben dem Inspector **334** Punkte — gemessen am 2026-09-09 (U5): der Arbeitsplatz misst dort 670, weil Rahmen und Setkopf 114 Punkte Rand nehmen; 0076 rechnete 464 ohne sie. Die Vorschau passt sich an, und die Seite rollt nicht waagerecht. In der Höhe 647 |
| **Gewählte Variante** | **B — der Rahmen bleibt, der Arbeitsplatz bekommt drei Regionen** aus einem gemeinsamen Baustein (U2): Navigator 240 links, Inspector 320 rechts, Zentrum nie unter 480, solange der Navigator steht; unter 1.200 Punkten Inhaltsbreite klappt er ein — **am Arbeitsplatz selbst gemessen sind das 1.072** (240 + 480 + 320 + zwei Abstände), siehe den Nachtrag in 0076 |
| **Verworfen** | A — drei Spalten im Rahmen für alle 23 Seiten: vier Spalten auf 1.280 Punkten, sechs Seiten ohne Navigator, ein Umbau in einem Zug |

---

## Listen, Kacheln und gemerkte Ansichten — UX-11

**UX-11** (#1683).
Die Sammlung hatte ihre drei Ansichten seit U8 (#1419); Sets und Decks
bekommen jetzt dasselbe Muster über **ein** Bauteil
([`Listenansicht`](../frontend/src/bausteine/Listenansicht.tsx)).

| Kriterium | Wie | Beleg |
| --- | --- | --- |
| **AK-01** — ein Ansichtswechsel erhält Auswahl, Suche, Filter, Sortierung, das sichtbare Objekt und die Stelle | Suche, Ordnung und die gezeigten Seiten stehen **über** der Gestalt und bleiben beim Wechsel stehen; **der Eintrag, an dem man war, hat danach wieder den Fokus** und steht im Blick. Die Sammlung behält ihre Sicht ohnehin (`useSammlung`) | `listenansicht.test.tsx`; WebKit-Drehbuch `ux11-ansichten` |
| **AK-02** — Detail und Mehrfachaktionen mit denselben Daten und Rechten; keine eigene Fachlogik in Galerien | **Ein Aufruf zeichnet beide Gestalten** (`eintrag(e, gestalt)`): `Setkarte` und der Eintrag der Deckliste sind Kachel und Zeile zugleich, mit denselben Knöpfen, denselben Vorgängen und derselben Sperre — die Gestalt ändert die Anordnung und sonst nichts | `listenansicht.test.tsx`, `deckliste`-Läufe |
| **AK-03** — große Bestände seitenweise oder virtualisiert; fehlende Vorschaubilder blockieren nicht | **Sets und Decks seitenweise** zu sechzig, die Sammlung virtualisiert (seit U8, Vermerk 0085). **Keine der drei Listen lädt Vorschaubilder** — eine Kachel zeigt Name, Spiel und Zahlen; das Bild einer Karte steht erst im Arbeitsplatz der Karten, und dort hat es seit #1270 seinen Ersatz | `listenansicht.test.tsx` (61 Einträge) |
| **AK-04** — Ansicht je Bereich und Profil speicherbar; leer und gefiltert leer unterschieden | **Drei Einstellungen am Arbeitsplatz** (`arbeitsplatz.ansicht.sets`, `.decks`, `.sammlung`): das Profil erbt vom Menschen, wie Navigator und Inspector (`internal/domain/ansicht.go`). **„Noch keine …" und „Nichts passt zu …"** sind zwei Sätze mit zwei Folgeschritten — anlegen oder die Suche leeren | `ansicht_test.go`, `ansicht.test.tsx`, `listenansicht.test.tsx`; `ux11-ansichten` |

---

## Farbwelten, Schrift und Fensterbild — UX-09

**UX-09** (#1681),
entschieden in [0134](entscheidungen/0134-eigene-gestaltung-sind-farben.md).

| Kriterium | Wie | Beleg |
| --- | --- | --- |
| **AK-01** — klar unterscheidbare, gut lesbare Standards; Nutzerhintergründe mit Vorschau, Skalierung und sofort erreichbarer Rücksetzung | **Drei Farbwelten neben Pergament** — Klar, Schiefer, Moos —, je hell und dunkel, gerechnet wie Pergament (`themen.css`, `docs/KONTRASTE.md`). **Ein eigenes Fensterbild**: gewählt über den Dialog wie das Tischbild (`FensterbildWaehlen`), gezeigt über einen eigenen Bildweg (`/fensterbild/`), unter einem Schleier von 90 %; *Einstellungen › Erscheinungsbild* zeigt es in einer Vorschau mit der gewählten Skalierung — Füllen, Einpassen, Kacheln —, und **Aus** und **Bild entfernen** stehen daneben | `themen.test.ts`, `gestaltung.test.ts`, `fensterbild_test.go`, `fensterbildweg_test.go`, `fensterbild_intern_test.go`; WebKit-Drehbuch `ux09-gestaltung` |
| **AK-02** — Schrift wahlweise für Überschriften oder die ganze Oberfläche; japanische und chinesische Rückfälle; Kartenschriften unabhängig | **„Anzeigeschrift gilt für"**: nur die Überschriften oder die ganze Oberfläche — dort nur eine der **vier Textschriften** (Pergament, Buch, Rund, Serifenlos); eine Schmuckschrift bleibt an den Überschriften, und die Seite sagt es. **Jede Kette trägt Hiragino vor PingFang/Songti** vor der allgemeinen Familie. Die Karte rastert Go mit den Schriften ihrer Vorlage | `schriftbereich.test.ts`; `ux09-gestaltung` misst die Breite japanischer und chinesischer Zeichen gegen Hiragino Mincho und zeigt, dass die Kartenvorschau dieselbe Adresse behält |
| **AK-03** — Prüfung auf Lesbarkeit, Kontrast und Überläufe; sicherer Standard auch bei unlesbarer Konfiguration | **Eigene Farben gelten nur, wenn jedes Paar hält** — gemessen am Wurzelelement nach Farbwelt und Erscheinungsbild, **bevor gezeichnet wird**; sonst gelten sie nicht, und die Seite nennt das Paar mit Zahl und Schwelle. **Schmuckschriften nie im Fließtext.** Überläufe hält `ux14-groessen` für alle Seiten. **„Darstellung zurücksetzen" in der Menüleiste** — macOS zeichnet sie, sie bleibt lesbar — führt die Layoutrücksetzung aus, die seit UX-09 auch Farbwelt, Schriftbereich, Fensterbild und eigene Farben umfasst; **hoher Kontrast sticht alles davon** | `kontrast.test.ts`, `gestaltung.test.ts`, `menuedarstellung_test.go`, `gestaltungseinstellungen_test.go`; `ux09-gestaltung` |
| **AK-04** — deklarativ und eingeschränkt; keine Skripte, keine unbemerkten Netzbezüge, keine verteilten Schriftdateien | **Die Form ist die Grenze**: ein bekanntes Merkmal und eine Hexzahl je Zeile (`domain.PruefeGestaltung`), im Fenster noch einmal gelesen und einzeln über `style.setProperty` gesetzt — kein `url(`, kein `@import`, kein `@font-face`, keine Maskierung. **Das Fensterbild nimmt nur die eigene Adresse** (`/fensterbild/<n>.png`). Schriften sind Namen von Systemschriften, keine Dateien | `gestaltung_test.go` (vierzehn abgewiesene Formen), `gestaltung.test.ts` |

**Was nur ein Mensch prüft**: der Dateidialog des Fensterbilds und der
Menüeintrag ([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)).

---

## Anordnung und Enge — UX-14

**UX-14** (#1686).
AK-01 (wie das Fenster aufgeht) steht in `PROTOTYP-0.6.0.md`; hier der Rest.

| Kriterium | Wie | Beleg |
| --- | --- | --- |
| **AK-02** — die Leiste einklappbar, **und** oben möglich, mit denselben Aktionen, Kürzeln, Fokuswegen und Markierungen | Einklappen gab es (⌥⌘S). **Neu die Geräteeinstellung „Lage der Leiste: Links / Oben"** (`leistenlage`): oben stehen dieselben Teile in derselben Reihenfolge als Zeile, die aufgeschlagene Seite trägt den Balken unten statt links, ← und → wandern, ⌘1–⌘6 bleiben. **Damit die Zeile schmal bleibt**: die Kürzel stehen im Titel, der Name der Anwendung und die Angaben zu Mensch und Profil für den Vorleser (sichtbar bleibt „nur lesen"), die Palette zeigt „⌘K"; bricht sie um, statt zu rollen — 49 Punkte hoch bei 1.440, 87 bei 1.024 | `leistenlage.test.tsx`, `kontextleiste.test.tsx`, `seitenleiste.test.tsx` (Haken); WebKit-Drehbuch `ux14-leistenlage` |
| **AK-03** — bei Mindestgröße, Maximieren, Rückkehr, großer Schrift, Retina und zweitem Bildschirm bleiben primäre Aktionen erreichbar; Navigator und Inspector weichen aus | **Bis zum 2026-09-25 wichen sie nicht aus, sie wurden gesperrt**: bei 1.024 × 700 war „Navigator einblenden" auf dem Kartenarbeitsplatz angehalten — die Kartenliste war nicht zu erreichen —, bei Decks und Sammlung der Inspector. **Jetzt geht die weichende Spalte über dem Zentrum auf** (Kante statt Schatten), ⌥⌘[ und ⌥⌘] tun dasselbe, Escape schließt und gibt den Fokus zurück. **Und eine Knopfzeile brach bei großer Schrift nicht um** (`Setkarte`: „Über eine Exportvorlage …" ragte über den Rand) | `arbeitsplatzenge.test.tsx`; WebKit-Drehbuch `ux14-groessen`; Retina ist das Gerät der Probe (Verhältnis 2) |
| **AK-04** — gespeicherte Positionen beim Monitorwechsel plausibilisiert; jede registrierte Seite mit einer Größenabnahme | **Gemerkt wird nur die Größe**; sie wird beim Start auf die Arbeitsfläche des Bildschirms geklemmt, auf dem das Fenster aufgeht (`internal/fenster/bildschirm.go`, erst nach dem Start — vorher kennt Wails die Bildschirme nicht). **Jede der 25 registrierten Seiten** öffnet `ux14-groessen` bei einem **Inhalt** von 1.024 × 700 (nicht dem Rahmen des Fensters), in normaler und in sehr großer Schrift, über dem befüllten Bestand der Abnahme: keine rollt waagerecht, keine Spalte ist gesperrt, und **kein Knopf außerhalb eines Rollbehälters** ragt über den Rand. **Was der Lauf nicht fährt** (Zweitblick über Phase 6.3, M5 und M14): Knöpfe in rollenden Spalten — also in den Spalten der Arbeitsplätze —, „Leiste oben" zusammen mit „Sehr groß" und dem Mindestfenster, und eine eigene Schrift; er nennt auch nicht, wie viele Knöpfe er prüfte. Das trägt Phase 6.11 (#1672). **Die Klemmung selbst belegt die Hüllenprobe** am nativen Fenster (`start`) | `bildschirm_test.go`; `ux14-groessen`; Hüllenprobe `start` |
| **S06.NAT-01** — bei Mindestgröße alle Werkzeuge mit Tastatur | keine Spalte ist mehr gesperrt, und die Überlagerung öffnet über die Taste | `arbeitsplatzenge.test.tsx`, `ux14-groessen` |
| **S06.NAT-02** — die Schublade übernimmt den Fokus und gibt ihn zurück | seit #1810 | `ux08-schubladen` |
| **S06.NAT-03** — Maximieren und Rückkehr erhalten Auswahl, Entwurf und Spalten | die gewählte Karte, ein halb getippter Name und die offene Überlagerung überstehen 1.440 × 900 und den Weg zurück; **die Überlagerung ruht, solange Platz ist**, statt verworfen zu werden | `ux14-groessen`, `arbeitsplatzenge.test.tsx` |
| **S06.NAT-04** — große Schrift schneidet keinen primären Befehl ab | der zweite Durchgang von `ux14-groessen`, bei „Sehr groß" — **mit derselben Grenze wie oben**: Knöpfe in rollenden Spalten sieht er nicht | `ux14-groessen` |

**Was nur ein Mensch prüft**: ein zweiter Bildschirm und das native Maximieren
([`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md)).

---

## Die Vorherbilder

**Alle 23 Seiten, hell und dunkel, bei 1.280 × 800**, gegen den Messbestand,
erzeugt mit `node scripts/serienabzug.mjs <verzeichnis>` am 2026-09-08. Sie
liegen **nicht im Baum** — sechsundvierzig Bilder sind ein Messmittel, kein
Dokumentationsmaterial —, sondern unter
`~/Library/Application Support/MagicSetEditor4Mac/vorherbilder-0.5.0/`, wo
schon die kanonischen Sets und die Abstandsbilder liegen. Wer ein Nachherbild
neben ein Vorherbild legen will, fährt denselben Lauf noch einmal in ein
zweites Verzeichnis.

**Was die Bilder zeigen und was nicht**: den Aufbau jeder Seite mit dem
Messbestand — nicht die Karten eines echten Sets, denn die Saat hat keines.
Die Handbuchbilder in `docs/bilder/` bleiben davon unberührt.
