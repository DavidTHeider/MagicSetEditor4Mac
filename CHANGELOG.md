# Änderungsprotokoll

Alle für **Anwendende** sichtbaren Änderungen. Innerer Umbau steht nicht hier,
sondern in [`docs/ROADMAP.md`](docs/ROADMAP.md) und
[`docs/CHRONIK.md`](docs/CHRONIK.md).

Das Format folgt [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
die Versionierung [Semantic Versioning](https://semver.org/lang/de/).

`## [Unveröffentlicht]

## [1.0.0] — 2026-11-15

**Die erste Freigabe für Anwendende.** Sie trägt den Programmstand von 0.9.0 —
was dort gebaut war, kann 1.0.0, und was dort nicht belegt war, ist es hier
auch nicht. Die Grenzen stehen im README unter „Was diese Fassung nicht
belegt" und in [`docs/REIFEPROFIL.md`](docs/REIFEPROFIL.md).

### Was 1.0.0 kann

- **Gestalten** — Sets aus Magic Set Editor 2 öffnen, Karten anlegen und
  gestalten, mit den Vorlagen, die MSE2 benutzt, und ohne dass die Dateien
  dabei etwas verlieren. Eine Änderung am Feld ist sofort in der gezeichneten
  Karte zu sehen; ein Feld mit Wertskript speichert, was es rechnet, und ⌘Z
  nimmt eine Bearbeitung ganz zurück. Dazu der Symboleditor, Suchen und
  Ersetzen über ein Set, die Prüfung eines Sets auf Befunde, Druckbogen,
  Bildausfuhr als PNG und SVG und MTG JSON in beide Richtungen.
- **Sammeln** — eine eigene Sammlung aus Drucken mit Zustand, Sprache,
  Ausführung, Menge und Ort, Preise aus MTGJSON und Angebote von CardTrader,
  Decks, die auf Karten verweisen statt sie zu kopieren, mit Formaten und
  Turnieren.
- **Spielen** — ein gezeichneter Spieltisch mit einer Regelmaschine — wie weit
  sie reicht, zählt [`docs/REGELABDECKUNG.md`](docs/REGELABDECKUNG.md) —,
  eine Partie in drei Klicks oder „Schnell spielen" aus einem Deck, ⌘Z für den
  letzten Zug, ein Vorhang zwischen zwei Menschen an einem Gerät, und
  mitspielen über das eigene Netz oder über fremde Netze.
- **Handeln** — Tauschen, Verleihen und Teilen mit einem bestätigten
  Gegenüber, am Markttisch: ziehen statt tippen, Preise am Tausch, und ein
  Tausch prüft vor jedem Schritt, ob noch gilt, was du siehst.
- **Ressourcen** — Vorlagenpakete beziehen und in der Paketwerkstatt
  bearbeiten, Bilder, Marken und Schlüsselwörter; mitgeliefert sind die
  Vorlagen für M15-Karten und ein Beispielset, damit nach dem ersten Start
  sofort eine Karte dasteht.
- **Programm** — Nutzer und Profile mit Rollen, Anmeldung über fremde Anbieter,
  Touch ID und ein zweiter Faktor, Einstellungen in Ebenen, das Fenster in acht
  Sprachen, eine Startseite mit dem, was Aufmerksamkeit braucht, Kennzahlleisten
  und Rechtsklick an jedem Objekt.
- **Ohne Fenster** — dieselben Fähigkeiten über die Kommandozeile `mse`, den
  MCP-Server `mse-mcp` und den Dienst `mse-rest`, mit maschinenlesbaren
  Fehlern; eingerichtet und wieder herausgenommen auf Knopfdruck.
- **Betrieb** — Sicherungen mit Kennwort im Format von `age`, wöchentlich auf
  Wunsch, ein Hinweis nach einem Abbruch, Meldungen, bevor etwas kippt, und
  der Speicher von Anwendung und Webprozess getrennt sichtbar.
- **Lokale und fremde KI** — Ollama, Apples Gerätemodell, Claude, OpenAI und
  Grok für Kartennamen, Flavortexte und die Suche im eigenen Bestand; jeder
  Dienst ist aus, bis du ihn einschaltest.

### Geändert gegenüber 0.9.0

- **Go 1.27.4 und `klauspost/compress` v1.20.2** — zwei
  Sicherheitsaktualisierungen: GO-2026-6859 (ein HTTP/1-Server konnte über
  einen Wettlauf im `MaxBytesReader` abstürzen, erreicht über den Paketbezug
  und den REST-Dienst) und GO-2026-6865 (eine Panik bei Zwei-Block-Frames).
- **Ein eigener Freigabeschlüssel** — Freigaben signiert `44C59428DE1F766E`,
  getrennt vom Schlüssel der Commits, und die Anwendung prüft eine
  Aktualisierung nur noch gegen ihn. **Eine installierte Entwicklungsfassung
  nimmt 1.0.0 deshalb nicht über die Aktualisierung an** — sie kennt nur den
  alten Schlüssel; 1.0.0 kommt dann als Abbild von Hand.
- **Der Prüfstand belegt VoiceOver mit der Tafel der Handprüfungen** statt mit
  einem Issue, das außerhalb nicht zu öffnen ist.

## Entwicklungsfassungen 0.1.0 bis 0.9.0

Nicht mehr veröffentlicht. Ihr Protokoll im Einzelnen, ihre Freigaben und ihre
Issues liegen im Entwicklungsarchiv
([0205](docs/entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)); ein Verweis
`#N` in Code und Doku nennt ein Issue oder einen PR von dort.

| Fassung | Erschienen | Schwerpunkt |
| --- | --- | --- |
| 0.9.0 | 2026-10-21 | Produktionsreife: Markttisch, direkte Hand, gespeichert wird, was gerechnet ist, Kennzahlleisten, ein Betrieb, der sich meldet, jeder Kernweg mit Budget |
| 0.8.0 | 2026-10-12 | Einrichtung aus dem Fenster, Fehlerfälle und Belege belastbarer |
| 0.7.0 | 2026-10-08 | Verfeinerung: Entwürfe, die überstehen, Schließen in fünf Sekunden, strengere Dienste, 70 MB weniger |
| 0.6.0 | 2026-10-05 | Arbeitsbereiche auf Wails 3, der gezeichnete Tisch, Handel zwischen zwei Anwendungen, Einstellungen in Ebenen, acht Sprachen |
| 0.5.0 | 2026-09-20 | Spielen, Handeln, KI, die semantische Recherche, Nutzer und Profile, das Netz |
| 0.4.0 | 2026-09-07 | Nutzer und Profile, der Lebenszyklus, eine Messung für jedes Set |
| 0.3.0 | 2026-08-28 | Paketwerkstatt, Sammlung, Marken, Suchen und Ersetzen über ein Set, eine Vorschau unter 100 ms |
| 0.2.0 | 2026-08-26 | Symboleditor, Schlüsselwörter, der Dienst auf der Rückschleife, Scryfall |
| 0.1.0 | 2026-08-22 | die erste Fassung: Sets aus MSE2 öffnen, Karten bearbeiten und zeichnen, Bildausfuhr, der Bestand und seine Sicherung |
