# Schalter

**Welche Fähigkeit an welchem Schalter hängt** — OPS-10
(#1908,
Vermerk [0158](entscheidungen/0158-ein-schalter-verbirgt-er-sichert-nicht.md)).

Ein Schalter verbirgt einen Arbeitsbereich oder eine Funktion an allen vier
Flächen. **Er sichert nichts**: wer einen Bereich abschaltet, hat ihn aus dem
Blick genommen, nicht vor jemandem verschlossen — die Rollen aus
[`AUTORISIERUNG.md`](AUTORISIERUNG.md) gelten unverändert.

## Die Schalter

Vorgabemäßig ist **jeder an**. Gespeichert wird je Schalter eine Einstellung
`schalter.<kennung>` mit `an` oder `aus`, Träger **Rechner**; was nicht
eingetragen ist, gilt mit seiner Vorgabe. Das Verzeichnis steht in
`internal/domain/schalter.go`, und ein Wächter hält es gegen diese Tabelle.

| Kennung | Art | Reife | Vorgabe | Was er abschaltet |
| --- | --- | --- | --- | --- |
| `sammeln` | Bereich | stabil | an | den Arbeitsbereich **Sammeln** — Sammlung, Besitz, Wunschlisten |
| `spielen` | Bereich | stabil | an | den Arbeitsbereich **Spielen** — Decks, Turniere, Partien und den Tisch |
| `handeln` | Bereich | stabil | an | den Arbeitsbereich **Handeln** — Handel, Freigaben, Umschläge |
| `ressourcen` | Bereich | stabil | an | den Arbeitsbereich **Ressourcen** — Vorlagenpakete und Vorlagen |
| `ki` | Funktion | stabil | an | die **KI** — Vorschläge, Bilder, Suchindex, Konten und Kosten |
| `netz` | Funktion | stabil | an | **Netz und Gegenüber** — Identität, Kontakte, Einladungen, Erreichbarkeit |
| `markt` | Funktion | stabil | an | **Marktplätze und Preise** |
| `turniere` | Funktion | stabil | an | die **Turniere** |
| `dienst` | Funktion | stabil | an | den **Dienst für Skripte** (REST), wenn das Fenster ihn startet |
| `werkzeugserver` | Funktion | stabil | an | den **Werkzeugserver über HTTP** (MCP), wenn das Fenster ihn startet |

**Gestalten und das Programm sind der Kern** und haben keinen Schalter:
ohne sie gäbe es kein Fenster, in dem man einen Schalter wieder einschaltet.

## Was mit einem Schalter endet

**Ein Lauf endet, sobald sein Beenden gesperrt ist** (OPS-10b,
#1978).
Sonst wiese das Tor das Beenden ab, und der Lauf hielte bis zum
Programmende, was er hält. Die Liste steht in
`internal/domain/schalterlaeufe.go`; welcher Schalter einen Lauf beendet,
sagt die Tafel über seine Beenden-Bindung, und
`internal/werkzeuge/schalterlaeufe_test.go` hält diese Tabelle gegen beide.

| Lauf | Beenden | endet mit | was er hält |
| --- | --- | --- | --- |
| Sitz an einem fremden Tisch | `PartieBeitrittBeenden` | `spielen`, `netz` | eine Sitzung beim Wirt |
| Partiefreigabe | `PartieFreigabeBeenden` | `spielen`, `netz` | die Zusage eines Gastsitzes und die Sitzungen der Gäste |
| Ankündigung im Netz | `SichtbarkeitBeenden` | `netz` | Peerkennung und Namen im mDNS |
| Horchposten | `HorchBeenden` | `netz` | einen Port |
| Fernhorchen | `FernHorchBeenden` | `netz` | einen Tunnel |
| Indexaufbau | `IndexAbbrechen` | `ki` | Rechenzeit |
| Bildaufträge, alle | `KIBildAbbrechen` | `ki` | Anfragen, die Geld kosten |
| Dienst für Skripte | `RESTBeenden` | `dienst` | einen Port und einen Prozess |
| Werkzeugserver über HTTP | `MCPBeenden` | `werkzeugserver` | einen Port und einen Prozess |

- **Erst die Sitzungen, dann die Ankündigung, dann die Ports** — ein
  Gegenüber erfährt vom Aufstehen, bevor der Weg zu ihm zugeht.
- **Legt ein anderer Prozess den Schalter um**, endet der Lauf beim nächsten
  Weg des Fensters, der den Stand liest.
- **Ein Start, der in das Abschalten fällt, fragt nach dem Eintragen noch
  einmal** und endet dann selbst (Zweitblick über Phase 6.12, Z2).
  - **Auch, wenn der Schalter inzwischen wieder an ist.** Je Lauf wird das
    Abschalten mitgezählt; ein Start vergleicht mit dem Zählerstand vor
    seinem Beginn.
  - Er beendet nur seinen **eigenen** Lauf, nicht einen, der danach anlief
    (Nachlauf über Phase 6.12, N2).
- **Abschalten entlässt die Gäste**; das Ende der Freigabe allein tut es
  nicht (Z3). Ein Gast, der sich erst danach einträgt, geht ebenso (N1).
- **Dienst und Werkzeugserver, abgeschaltet gestartet, enden mit Rückgabe
  0** und nennen den Grund auf der Fehlerausgabe. Sonst startete ihr
  LaunchAgent sie immer wieder (N4).
- **Eingeschaltet startet keiner von selbst.**
- **Eine Partie ist kein Lauf**: `PartieBeenden` hängt an `spielen` und
  bleibt, wie sie war.

## Wie zu lesen

- **Bereich** ist der Arbeitsbereich, in dem eine Fähigkeit steht — `kern`,
  wenn sie zu Gestalten oder zum Programm gehört.
- **Funktion** ist eine Funktion quer durch die Bereiche — `—`, wenn keine.
- **Eine Fähigkeit ist aus, sobald einer ihrer beiden Schalter aus ist.**
  Freigaben an ein Gegenüber hängen am Handeln **und** am Netz.
- **Ein Weg weist erst ab, wenn jede Fähigkeit, die ihn nennt, aus ist.** Ein
  Weg, den auch eine Fähigkeit des Kerns nimmt, weist nie ab.
- Die Wege selbst stehen in [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md); erzeugt wird
  daraus `internal/domain/schalterwege.json`, und
  `internal/werkzeuge/schalterwege_test.go` wird rot, sobald die beiden
  auseinanderlaufen — `SCHALTERWEGE_SCHREIBEN=1` schreibt sie neu.


## Sets

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Sets auflisten | kern | — |
| Ein Set öffnen | kern | — |
| Ein Set vorab ansehen | kern | — |
| Ein leeres Set anlegen | kern | — |
| Ein Set löschen | kern | — |
| Zuletzt benutzt merken | kern | — |
| Die Menüleiste nach dem Stand des Fensters führen — Titel und Aktivierung von Seitenleiste, Navigator und Inspector | kern | — |
| Die Handlungen des gewählten Objekts in die Menüleiste bringen — dieselbe Liste wie im Rechtsklick und im Überlauf | kern | — |
| Weiterarbeiten — die Übersicht mit Angeheftetem | kern | — |
| Was Aufmerksamkeit braucht — Sets mit Vorbehalten, Decks mit Fehlkarten, Pakete mit Skriptbefunden, und was der Betrieb meldet | kern | — |
| Hintergrundaufgaben — was gerade geholt wird, mit Stand, Fortschritt und Dauer | kern | — |
| Eine Ausgabe des Dienstes abholen | kern | — |
| Mitgelieferte Sets nennen | kern | — |
| Ein Set sichern | kern | — |
| Setangaben lesen und setzen | kern | — |
| Vorbehalte eines Sets lesen | kern | — |
| Die Karten eines Sets auf Probleme prüfen | kern | — |

## Karten

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Karten auflisten | kern | — |
| Die Gestalt der Kartenliste — Spalten und Ordnung | kern | — |
| Der Schnellfilter der Kartenliste | kern | — |
| Eine Diagnose für einen Fehlerbericht holen | kern | — |
| Ein Diagnosepaket zusammenstellen, sichern und wieder löschen | kern | — |
| Einen Fehler, den das Fenster beim Zeichnen gefangen hat, mit seiner Kennung ins Protokoll schreiben | kern | — |
| Sehen, dass die letzte Sitzung nicht sauber endete, und ob eine Absturzspur vorliegt | kern | — |
| Die Karten mit ihrem Bild sehen | kern | — |
| Marken an einer Karte führen | kern | — |
| Marken an jedem Fachobjekt führen | kern | — |
| Nach Marken filtern | kern | — |
| Marken pflegen — umbenennen, zusammenführen, überall entfernen | kern | — |
| Eine Karte auf einen echten Druck beziehen | kern | — |
| Karten bei Scryfall suchen | kern | — |
| Regelhinweise von Scryfall lesen | kern | — |
| Einen Manakosten gegen Scryfall prüfen | kern | — |
| Sammeldaten von Scryfall vorhalten | kern | — |
| Kartendaten von Scryfall übernehmen | kern | — |
| Kartenfelder beschreiben | kern | — |
| Eine Karte anlegen | kern | — |
| Die Kartenskripte eines Spielpakets zeigen | kern | — |
| Ein Kartenskript ausführen | kern | — |
| Wonach ein Spielpaket auszählen lässt | kern | — |
| Ein Set nach einer Dimension auszählen | kern | — |
| Die Parameterhilfen eines Schlüsselworts | kern | — |
| Die Decks sehen | spielen | — |
| Ein Deck samt seinen Plätzen sehen | spielen | — |
| Ein Deck anlegen | spielen | — |
| Ein Deck umbenennen | spielen | — |
| Ein Deck kopieren | spielen | — |
| Aus einem Set ein Deck anlegen | spielen | — |
| Ein Deck löschen | spielen | — |
| Eine Karte ins Deck legen | spielen | — |
| Einen Platz aus dem Deck nehmen | spielen | — |
| Einen Platz in den anderen Bereich legen | spielen | — |
| Ein Deck auf ein Format stellen | spielen | — |
| Ein Deck gegen sein Format prüfen | spielen | — |
| Nachsehen, was von einem Deck da ist | spielen | — |
| Sehen, welche Exemplare für ein Deck vorgesehen sind | spielen | — |
| Exemplare einem Deckplatz zuordnen | spielen | — |
| Eine Zuordnung wieder lösen | spielen | — |
| Die Fehlkarten eines Decks als Liste bekommen | spielen | — |
| Ein Deck als Liste ausgeben | spielen | — |
| Etwas über die Ziele des Systems teilen — eine Einladung, ein Kartenbild, ein Set, ein Deck, die Sammlung | kern | — |
| Die Fehlkarten eines Decks auf die Wunschliste setzen | spielen | — |
| Sehen, welche Decks ein Exemplar benutzen | spielen | — |
| Die Formate sehen | spielen | — |
| Die nachgelesenen Deckbauarten sehen | spielen | — |
| Ein Format aus einer Bauart anlegen | spielen | — |
| Ein Format anlegen oder fortschreiben | spielen | — |
| Ein Format löschen | spielen | — |
| Die Turniere sehen | spielen | turniere |
| Ein Turnier samt Verlauf sehen | spielen | turniere |
| Ein Turnier anlegen | spielen | turniere |
| Ein Turnier nach einer Vorlage anlegen | spielen | turniere |
| Ein Turnier löschen | spielen | turniere |
| Einen Teilnehmer anmelden | spielen | turniere |
| Einen Teilnehmer abmelden | spielen | turniere |
| Sehen, wie die nächste Runde ausfiele | spielen | turniere |
| Die nächste Runde legen | spielen | turniere |
| Ein Ergebnis eintragen | spielen | turniere |
| Die Rangliste sehen | spielen | turniere |
| Sehen, was eine Berichtigung verschöbe | spielen | turniere |
| Ein Turnier abschließen | spielen | turniere |
| Ein Turnier wiedereröffnen | spielen | turniere |
| Das Protokoll eines Turniers sehen | spielen | turniere |
| Ein Turnier als Datei ausgeben | spielen | turniere |
| Ein Turnier aus einer Datei einlesen | spielen | turniere |
| Den Bericht eines Turniers sichern | spielen | turniere |
| Die Autoersetzungen eines Spiels sehen | kern | — |
| Eine Autoersetzungsregel setzen | kern | — |
| Eine eigene Autoersetzungsregel löschen | kern | — |
| Die eigenen Autoersetzungen vergessen | kern | — |
| Die Autoersetzung ganz schalten | kern | — |
| Eine Karte kopieren | kern | — |
| Karten über die Zwischenablage bewegen | kern | — |
| Eine Karte löschen | kern | — |
| Karten umordnen | kern | — |
| Ein Feld setzen | kern | — |
| Ein Feld auf vielen Karten setzen | kern | — |
| Eine Massenänderung vorher sehen | kern | — |
| Ein Symbol zeichnen | kern | — |
| Ein Bild als Symbol nachzeichnen | kern | — |
| Ein gezeichnetes Symbol ins Set legen | kern | — |
| Einen Bézierhandgriff ziehen | kern | — |
| Einem Punkt seine Art geben | kern | — |
| Einen Punkt einer Form einfügen oder löschen | kern | — |
| Das Symbol eines Sets lesen | kern | — |
| Die Packarten eines Spiels zeigen | kern | — |
| Einen Booster ziehen | kern | — |
| Die Schlüsselwörter eines Sets zeigen | kern | — |
| Ein Schlüsselwort anlegen | kern | — |
| Ein Schlüsselwort ändern | kern | — |
| Ein Schlüsselwort löschen | kern | — |
| Die Auswahl eines Feldes zeigen | kern | — |
| Aus der Auswahl wählen | kern | — |
| Beim Tippen ersetzen | kern | — |

## Suchen und zählen

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Im Volltext suchen | kern | — |
| Im Text ersetzen | kern | — |
| Ein Set auszählen | kern | — |
| Einen Ausdruck auswerten | kern | — |
| Die Namen der Umgebung lesen | kern | — |
| Einen Ausdruck einfärben | kern | — |
| Eine Paketdatei einfärben | ressourcen | — |
| Die Dateien einer Vorlage aufzählen | ressourcen | — |
| Die Skripte einer Vorlage prüfen | ressourcen | — |
| Eine Vorlage lesen | ressourcen | — |
| Eine Vorlage schreiben | ressourcen | — |
| Eine Vorlage im Paket anlegen | ressourcen | — |
| Ein Bild oder eine Schrift ins Paket legen | ressourcen | — |
| Eine Datei im Paket umbenennen | ressourcen | — |
| Eine Datei aus dem Paket entfernen | ressourcen | — |
| Eine ungespeicherte Vorlage sehen | ressourcen | — |
| Die Testkarten zu einer Vorlage nennen | ressourcen | — |
| Die Symbole einer Vorlage nennen | ressourcen | — |
| Die Verwendungen eines Namens nennen | kern | — |
| Die eigene Sammlung aufzählen | sammeln | — |
| Nachsehen, was von einem Druck da ist | sammeln | — |
| Ein Exemplar aufnehmen | sammeln | — |
| Ein Exemplar ändern | sammeln | — |
| Ein Exemplar herausnehmen | sammeln | — |
| Die abgelegten Preise eines Drucks sehen | kern | markt |
| Den Preisverlauf sehen | kern | markt |
| Preise holen | kern | markt |
| Preise einer ganzen Quelle holen | kern | markt |
| Die Angebote eines Marktplatzes sehen | kern | markt |
| Den Preisverlauf einer anderen Quelle holen | kern | markt |
| Das Kontingent an einem Anbieterkonto sehen | kern | markt |
| Sehen, woher ein Kartenwert stammt | kern | ki |
| Sehen, was über KI-Modelle feststeht | kern | ki |
| Ein Modell Felder einer Karte vorschlagen lassen | kern | ki |
| Mehrere Karten auf einmal vorschlagen lassen | kern | ki |
| Mehrere Vorschläge als eine Gruppe übernehmen | kern | ki |
| Ein Modell ein Kartenbild erzeugen oder bearbeiten lassen | kern | ki |
| Eine Karte beurteilen lassen — mit Belegen | kern | ki |
| Eine Karte auf einem Foto erkennen | kern | ki |
| Messen, was ein Modell kann | kern | ki |
| Apples Gerätemodell für eine Sitzung abschalten | kern | ki |
| Eine Kartenliste deuten | sammeln | — |
| Eine Kartenliste einlesen | sammeln | — |
| Eine Kartenliste vorab ansehen | sammeln | — |
| Einen verbindlichen Handelsschritt freigeben | handeln | netz |
| Eine Deckliste deuten | spielen | — |
| Eine Deckliste ins Deck legen | spielen | — |
| Viele Exemplare auf einmal ändern | sammeln | — |
| Viele Exemplare auf einmal löschen | sammeln | — |
| Preise für die ganze Sammlung | sammeln | markt |
| Besitze ich diese Karte? | sammeln | — |
| Die Lagerorte zeigen | sammeln | — |
| Dubletten finden | sammeln | — |
| Die Wunschliste zeigen | sammeln | — |
| Die Sammlung ausgeben | sammeln | — |
| Eine CSV-Sammlung zurücklesen | sammeln | — |
| Eine CSV-Sammlung vorab ansehen | sammeln | — |
| Die Sammlung zusammenrechnen | sammeln | markt |
| Die Werkzeuge ohne Fenster finden | kern | — |
| Die Kommandozeile einrichten — `mse` und `mse-mcp` in `~/.local/bin` | kern | — |
| Nachsehen, was Kommandozeile, Codex, Claude und Dienst von dieser Anwendung wissen | kern | — |
| Kommandozeile, Codex, Claude und Dienst einrichten und widerrufen — je Ziel oder alle mit einem Ruf, und wählen, wo der Dienst lauscht | kern | — |

## Aussehen

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Stile nennen und wechseln | kern | — |
| Stiloptionen lesen und setzen | kern | — |
| Die Felder des Stils an einer Karte lesen und setzen | kern | — |
| Die Wahlen einer Stiloption zeigen | kern | — |
| Eine Karte zeichnen | kern | — |

## Bilder

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Ein Kartenbild einsetzen | kern | — |
| Ein Bild zuschneiden | kern | — |
| Ein Kartenbild von Scryfall holen | kern | — |
| Eine Karte aus einem Scryfall-Namen anlegen | kern | — |
| Ein ganzes Set von Scryfall einlesen | kern | — |
| Teile gruppieren, auflösen und umordnen | kern | — |
| Eine Symmetrie bedienen | kern | — |
| Symbole zur Auswahl stellen | kern | — |
| Die Symbolbibliothek sehen und durchsuchen | ressourcen | — |
| Ein Setsymbol in die Bibliothek aufnehmen | ressourcen | — |
| Ein Symbol der Bibliothek duplizieren und umbenennen | ressourcen | — |
| Ein Symbol der Bibliothek löschen | ressourcen | — |
| Ein Symbol der Bibliothek in ein Set einsetzen | ressourcen | — |
| Ein Symbol ein- und ausgeben | ressourcen | — |
| Ein Symbol der Bibliothek bearbeiten und zurücknehmen | ressourcen | — |
| Die Sets mit der vorigen Fassung nachziehen | ressourcen | — |
| Ein Symbol in den Fassungen eines Sets sehen | ressourcen | — |
| Die eigenen Übersetzungen sehen | kern | — |
| Eine eigene Übersetzung anlegen | kern | — |
| Eine eigene Übersetzung bearbeiten | kern | — |
| Eine eigene Übersetzung löschen | kern | — |
| Eine eigene Übersetzung ein- und ausgeben | kern | — |
| Kartenbilder schreiben | kern | — |
| Ein Profilbild setzen und wegnehmen | kern | — |
| Die Einstellungen eines Menschen lesen und setzen | kern | — |
| Die eigenen Daten ausgeben | kern | — |
| Ein Profil-Abzeichen setzen und wegnehmen | kern | — |
| Die Nutzer mit ihrem Bild sehen | kern | — |

## Ausgeben

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Einen Druckbogen legen | kern | — |
| Über eine Exportvorlage ausgeben | kern | — |
| MTG JSON lesen | kern | — |
| MTG JSON schreiben | kern | — |

## Vorlagenpakete

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Pakete auflisten | ressourcen | — |
| Ein Paket beziehen | ressourcen | — |
| Nutzer und Profile führen | kern | — |
| Das Profil wechseln | kern | — |
| Über einen Anbieter anmelden | kern | — |
| Anbieterkonten führen | kern | — |
| Nutzer und Profile bearbeiten | kern | — |
| Ein eigenes Konto anlegen und sich damit anmelden | kern | — |
| Einen zweiten Faktor einrichten und verlangen | kern | — |
| Das Fenster sperren, entsperren, abmelden und lokal anmelden | kern | — |
| Das eigene Kennwort wechseln oder über den Zettel wiederherstellen | kern | — |
| Zugriffstoken für ein technisches Profil führen | kern | — |
| Vorgaben eines Profils, die auf Zeilen zeigen | kern | — |
| Wer außer dem Eigner an einem Profil etwas darf | kern | — |
| Ein Deck aus einem fremden Profil holen | spielen | — |
| Die Decks eines fremden Profils sehen | spielen | — |
| Ein Paket aus einem Gerüst anlegen | ressourcen | — |
| Ein Paket als eigene Kopie anpassen | ressourcen | — |
| Ein Paket als `.mse-installer` weitergeben und einlesen | ressourcen | — |
| Ein Paket einbinden oder entfernen | ressourcen | — |
| Vor einer Paketänderung sehen, was sie bewirkt | ressourcen | — |
| Eine neue Fassung über ein Paket legen | ressourcen | — |
| Die vorige Fassung eines Pakets zurückholen | ressourcen | — |
| Die Stände der bezogenen Pakete sehen und eine Aktualisierung prüfen | ressourcen | — |
| Ein bezogenes Paket aktualisieren und zurückholen | ressourcen | — |
| Sehen, welche Pakete einem Set fehlen | ressourcen | — |
| Aus einer MSE2-Installation übernehmen | ressourcen | — |

## Der Bestand

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Den Bestand sichern | kern | — |
| Sehen, wann der Bestand zuletzt gesichert wurde — und den freiwilligen Plan | kern | — |
| Den freiwilligen Plan einrichten und beenden — wöchentlich in einen Ordner, sieben eigene | kern | — |
| Eine Datei ansehen | kern | — |
| Eine Sicherung zurückspielen | kern | — |
| Einen gescheiterten Start erneut versuchen | kern | — |
| Wer wo als was arbeitet — der Arbeitskontext | kern | — |
| Wer arbeitet, ansehen | kern | — |
| Ein anderer Nutzer werden | kern | — |
| Den Bestand verschlüsseln | kern | — |
| Eine Änderung zurücknehmen | kern | — |
| Den Verlauf eines Sets sehen | kern | — |
| Den Stand des Bestands lesen | kern | — |

## Das Programm

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Sehen, was die fremden Dienste hier können — der Dienstkatalog | kern | — |
| Sehen, was dieser Prozess an Rufen und Parallelität zulässt — die Grenzen | kern | — |
| Sehen, ob dieser Prozess lebt und arbeiten kann — der Betriebszustand | kern | — |
| Sehen, was dieser Prozess gewirkt hat — Zahlen mit endlich vielen Reihen | kern | — |
| Ein Leistungsprofil dieses Fensters aufzeichnen | kern | — |
| Die Zwischenspeicher sehen und leeren | kern | — |
| Betrieb und Vertrauen sehen — Fassung, Datenstand, Signatur, Aktualisierung, Orte und was geprüft ist | kern | — |
| Sehen, welches Programm antwortet — Fassung, Plattform, Fläche | kern | — |
| Die Kommandozentrale — einen Fachbefehl strukturiert wählen, mit Vorschau ausführen, abbrechen | kern | — |
| Die Skriptkonsole — einen laufenden Ausdruck abbrechen | kern | — |
| Zugangsdaten fremder Dienste verwalten | kern | — |
| Das Geheimnis eines Anmeldeanbieters hinterlegen | kern | — |
| Kontakte aus einem Adressbuch übernehmen | kern | netz |
| Sehen, was Platz belegt — und alte Sicherungen wegräumen | kern | — |
| Sehen, welche Bereiche und Funktionen an sind — die Schalter | kern | — |
| Einstellungen lesen und setzen | kern | — |
| Einstellungen ausführen | kern | — |
| Eine Einfuhr von Einstellungen vorschauen | kern | — |
| Eine Einfuhr von Einstellungen übernehmen | kern | — |
| Nach einer neueren Fassung sehen und sie holen | kern | — |
| Sehen, ob die neueste Freigabe diesen Bestand öffnen kann | kern | — |
| Nachsehen, was im Bestand herumliegt | kern | — |
| Aufräumen, was im Bestand herumliegt | kern | — |
| Sehen, was zurückgehen kann und was davon diese Anwendung selbst tut | kern | — |
| Sehen, was ein Zurückspielen dieser Sicherung kostet | kern | — |
| Sehen, worüber nach einer Rückkehr nichts feststeht | kern | — |
| Eine Rückkehr für nachgesehen erklären | kern | — |
| Sehen, wie viel Geld an einem Konto freigegeben ist | kern | ki |
| Die letzten Aufträge eines Kontos ansehen | kern | ki |
| Die Kostengrenze eines Kontos setzen | kern | ki |
| Einen bezahlten KI-Auftrag freigeben | kern | ki |
| Die Arten, zurückzusetzen, ansehen | kern | — |
| Sehen, was eine Rücksetzung anfassen würde | kern | — |
| Zurücksetzen | kern | — |
| Den Neuanfang freigeben | kern | — |
| Einen Fassungswechsel vorbereiten — Platz, laufende Arbeit, eine zurückgelesene Sicherung | kern | — |
| Einen laufenden Bildauftrag abbrechen | kern | ki |
| Sehen, wogegen der lokale Suchindex sucht | kern | ki |
| Lokal suchen — Volltext und Vektoren zusammengeführt | kern | ki |
| Den Suchindex neu aufbauen | kern | ki |
| Den Wissensbestand fragen — Karten, Regeln, Rulings | kern | ki |
| Karten finden, die einer eigenen ähneln | kern | ki |
| Sehen, was an ein Modell ginge | kern | ki |
| Einen laufenden Indexaufbau abbrechen | kern | ki |

## Peer und Austausch

| Fähigkeit | Bereich | Funktion |
| --- | --- | --- |
| Sehen, womit sich dieses Profil gegenüber anderen ausweist | kern | netz |
| Die erste Identität anlegen | kern | netz |
| Die Identität ablösen | kern | netz |
| Sehen, wen dieses Profil kennt | kern | netz |
| Ein Gegenüber von Hand eintragen | kern | netz |
| Ein Gegenüber bestätigen | kern | netz |
| Eine Einladung ausstellen | kern | netz |
| Eine Einladung prüfen | kern | netz |
| Eine Einladung annehmen | kern | netz |
| Ein Gegenüber sperren | kern | netz |
| Eine Sperre zurücknehmen | kern | netz |
| Einen Zweck freigeben | kern | netz |
| Einen Zweck zurücknehmen | kern | netz |
| Eintragen, wo ein Gegenüber erreichbar ist | kern | netz |
| Sich zu einem Gegenüber etwas notieren | kern | netz |
| Sehen, was ein Gegenüber sehen darf | handeln | netz |
| Einem Gegenüber etwas freigeben | handeln | netz |
| Eine Freigabe zurücknehmen | handeln | netz |
| Sehen, was ein Gegenüber wirklich sähe | handeln | netz |
| Fragen, was ein Gegenüber zum Handeln freigibt — über die Leitung | handeln | netz |
| Sehen, was zwei Sammlungen einander zu bieten haben | handeln | netz |
| Sehen, was zu einem Gegenüber noch aussteht | handeln | netz |
| Einen Handel eröffnen | handeln | — |
| Eine neue Revision vorschlagen | handeln | — |
| Einen Handel bis zur Übergabe führen | handeln | — |
| Sehen, wo ein Handel steht | handeln | — |
| Die Handel mit einem Gegenüber sehen | handeln | — |
| Alle eigenen Handel sehen | handeln | — |
| Sehen, wie ein Handel dorthin kam | handeln | — |
| Sehen, was die Seiten eines Tauschs ungefähr wert sind | handeln | markt |
| Einen Entwurf neu schreiben | handeln | — |
| Einen Handel als neuen Entwurf kopieren | handeln | — |
| Einen Entwurf löschen | handeln | — |
| Einen beendeten Handel archivieren | handeln | — |
| Den Stand eines Handels an ein Gegenüber schicken | handeln | netz |
| Welche Besitzzeilen für einen Handel weggehen | handeln | — |
| Den Ausgang über eine Leitung zu einem Gegenüber tragen | handeln | netz |
| Eine angekommene Nachricht wegräumen | handeln | — |
| Sagen, was nach einem Abriss noch einmal hinausmuss | handeln | netz |
| Einen übergebenen Handel im eigenen Bestand verbuchen | handeln | — |
| Der Gegenseite melden, dass hier fertig ist | handeln | — |
| Die zweite Unterschrift entgegennehmen | handeln | — |
| Sehen, was sich im eigenen Bestand bewegt hat | handeln | — |
| Sehen, wer einen Handel unterschrieben hat | handeln | — |
| Ein Deck aus einem Umschlag annehmen | handeln | — |
| Eine Karte aus einem Umschlag annehmen | handeln | — |
| Eine Vorlage aus einem Umschlag annehmen | handeln | — |
| Eine empfangene Sammlungssicht ansehen | handeln | — |
| Sehen, wer an einem Tisch saß | spielen | — |
| Das eigene Konto bei einem Marktplatz ansehen | kern | markt |
| Sehen, wo es eine Karte außerhalb gibt | kern | markt |
| Einen dieser Wege im Browser öffnen | kern | markt |
| Für andere erreichbar sein | kern | netz |
| Eine Partie für Gäste öffnen | spielen | netz |
| An einem fremden Tisch mitspielen | spielen | netz |
| Am Tisch reden und zeigen | spielen | — |
| Was von einer Partie bleibt | spielen | — |
| Wie die Leitung steht | spielen | — |
| Ein Deck verleihen und leihen | spielen | netz |
| Eine Paarung an eine Partie binden | spielen | turniere |
| Über fremde Netze erreichbar sein | kern | netz |
| In der Nähe gefunden werden | kern | netz |
| Sehen, wer in der Nähe ist | kern | netz |
| Einen Gegenstand als Umschlag ablegen | handeln | — |
| Sehen, was in einem Umschlag steht | handeln | — |
| Einen Umschlag übernehmen | handeln | — |
| Eine Partie anfangen | spielen | — |
| Sehen, wie eine Partie für einen Sitz aussieht | spielen | — |
| Sehen, welche Partien es gibt | spielen | — |
| Sehen, was in einer Partie geschehen ist | spielen | — |
| Einen Spielbefehl stellen | spielen | — |
| Eine Partie anhalten, fortsetzen oder beenden | spielen | — |
| Eine Partie löschen oder kopieren | spielen | — |
| Eine Partie ausführen und einlesen | spielen | — |
| Sehen, was von den Spielregeln durchgesetzt wird | spielen | — |
| Einen eigenen Tischhintergrund mitbringen | spielen | — |
| Ein eigenes Fensterbild mitbringen | kern | — |
| Einen Abzug des Tisches sichern | spielen | — |
| Eine Karte am Spieltisch zeichnen | spielen | — |
| Sehen, wie dieses Programm hinausgeht | kern | — |
| Den Dienst für Skripte schalten | kern | dienst |
| Den Werkzeugserver über HTTP schalten | kern | werkzeugserver |
