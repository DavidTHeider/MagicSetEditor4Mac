# Dateiansichten

**Was eine Datei in dieser Anwendung zeigt, bevor und nachdem sie gelesen
oder geschrieben wird** — UX-13
(#1685),
Marke `new`: MSE2 zeigt eine Datei nicht an, es öffnet oder schreibt sie.
Gesucht wurde in `src/gui/set/window.cpp` (Ablagemenü) und
`src/gui/card_select_window.cpp`; eine Vorschau der geschriebenen Datei gibt
es dort nicht.

## 1 Welche Vorschau für welche Art sicher ist (AK-01, AK-03)

**Die Regel steht an einer Stelle**, `domain.AnsichtsregelFuer`
(`internal/domain/dateiart.go`), und gilt an jeder Fläche. Gelesen wird sie
von `service.DateiAnsehen`; das Fenster, die Kommandozeile (`mse
datei-ansehen`), der Dienst (`GET /dateien/ansicht`) und der Werkzeugserver
(`lesen.datei.ansicht`) geben dieselbe Auskunft — **nur das Fenster bekommt
Bilddaten**, die anderen haben die Datei selbst.

| Art | Endungen | Vorschau | Gelesen höchstens | Öffnen mit dem Standardprogramm |
| --- | --- | --- | --- | --- |
| **MSE-Paket** | `mse-set`, `mse-game`, `mse-style`, `mse-symbol-font`, `mse-include`, `mse-locale`, `mse-export-template` | Liste der Einträge — als Ordner über `WalkDir`, als Archiv über das ZIP-Verzeichnis | 2.000 Einträge | nein — es gehört in diese Anwendung |
| **Bild** | `png`, `jpg`, `jpeg`, `gif`, `webp`, `bmp`, `tif`, `tiff` | Bild samt Maßen | 20 MiB | ja |
| **Vektorbild** | `svg` | Bild über `<img>` | 1 MiB | ja |
| **PDF** | `pdf` | die Zahl der Seiten | 20 MiB | ja |
| **Text** | `txt`, `json`, `csv`, `tsv`, `md`, `dec`, `dck`, `mwdeck`, `xml`, `log`, `html`, `htm` | Text in `<pre>` | 256 KiB | ja — **außer** `html`, `htm`, `xml` |
| **Archiv** | `zip`, `msesicherung`, `msepackage` | Liste der Einträge | 2.000 Einträge | nein |
| **Unbekannt** | alles andere, auch ohne Endung | nur Name, Größe, Zeit, Ort | nichts | nein |

**Die sieben Paketarten stehen hier ein zweites Mal**, weil `domain` nichts
aus dem Projekt kennt (§2) — ein Wächter hält sie gegen `msepaket.Arten()`
(`internal/werkzeuge/dateiart_test.go`).

### Was dabei nie geschieht (AK-03)

| Gefahr | Was die Ansicht tut | Beleg |
| --- | --- | --- |
| **Ein Archiv packt sich aus** | Das ZIP-Verzeichnis wird gelesen, kein Eintrag geöffnet; das Verzeichnis daneben bleibt leer | `TestEinArchivWirdGelesenUndNichtAusgepackt` |
| **Ein Archiv mit Millionen Einträgen** | Das Verzeichnis wird **Kopf um Kopf** gelesen und nach der Grenze **nicht weiter** — `archive/zip` läse es vorher ganz (Go 1.27.1, `reader.go`, `(*Reader).init`). Ein Name wird nach 1 KiB gekürzt | `TestHinterDerGrenzeWirdNichtGelesen`, `TestEinUeberlangerNameWirdGekuerzt` (`internal/service/zipverzeichnis_test.go`) |
| **Ein Name führt hinaus** (`../`, `/absolut`) | Er wird **gezählt** und als *nicht betreten* markiert | dieselbe Probe, `Unsicher = 2` |
| **Ein Verweis im Paketordner** | `WalkDir` folgt keinem Verweis; er steht als *nicht betreten* da | `TestEinPaketErscheintAlsListeUndNichtsWirdGefolgt` |
| **Ein Paketordner mit sehr vielen Dateien** | Der Gang **endet** an der Grenze (`fs.SkipAll`) und steigt in keinen weiteren Unterordner — nicht nur die Liste ist begrenzt | `TestDerOrdnergangEndetAnDerGrenze` |
| **Ein SVG mit Skript** | Es erscheint über `<img>` — in WebKit läuft dort kein Skript, und keine fremde Adresse wird geladen. **Die Quelle** ist W3C *SVG Integration* (Entwurf vom 2014-04-17, §2 und §3.4: `<img>` heißt *secure animated mode*, ohne Skript und ohne externe Verweise), die HTML aber ausdrücklich nicht bindet — **deshalb gemessen**: an der WKWebView der Anwendung lädt ein SVG mit `<script>`, `onload`, `@import` und `<image>` auf einen Dienst an `127.0.0.1` über `<img>` **nichts**; als Dokument im `<iframe>` fragt dasselbe SVG alle vier Adressen an (Gegenprobe) | `scripts/huellenprobe/svgbild.mjs`, Beleg `docs/belege/huellenprobe/svgbild.md`; Bytes und Typ `TestEinBildErscheintMitSeinenMassen` |
| **HTML mit Skript** | Es erscheint als **Text**; React setzt ihn als Textknoten, und `Öffnen` fehlt, weil ein Browser das Skript ausführte | `TestEinTextErscheintAlsTextUndHoechstensBisZurGrenze`, `dateiansicht.test.tsx` |
| **Eine sehr große Datei** | Gelesen wird höchstens die Grenze der Art; ein Bild darüber wird **gar nicht** gelesen und sagt *zu groß* | `TestZuGrossZeigtNurWasBekanntIst` |
| **Eine Datei, die ein Programm sein kann** (`.command`, `.app`, ohne Endung) | Sie zeigt nur ihre Angaben; `Öffnen` fehlt, und Go weist es auch ohne den fehlenden Knopf ab | `TestEineUnbekannteDateiZeigtNurIhreAngaben`, `TestImFinderZeigenUndOeffnenFolgenDerRegel` |
| **Ein Pfad, den `open` als Option läse** | Nur ein ganzer Pfad geht an macOS; `-a Rechner` wird abgewiesen | `TestImFinderZeigenUndOeffnenFolgenDerRegel` |
| **Ein Ordner zum Öffnen** | Abgewiesen — ein Programmbündel ist einer | dieselbe Probe |
| **Eine Sicherung mit Kennwort** | Sie sagt, dass sie es ist; ohne Kennwort liegt in ihr kein lesbares Verzeichnis (OPS-06) | `TestEinArchivWirdGelesenUndNichtAusgepackt` |
| **Ein Winkel des Zuhauses über den Dienst** | `GET /dateien/ansicht` folgt dem Pfadvertrag: `~/.ssh` bleibt zu, auch zum Ansehen | `TestEineDateiAnsehenUeberJedeFlaeche` |

**Was die Ansicht nicht prüft**: ob ein PDF mehr Seiten hat, als die Zählung
der Seitenobjekte sagt — ein PDF mit Objektströmen nennt weniger. Das ist
eine Zählung und kein Lesen, und sie steht so an `pdfseiten`.

## 2 Jeder Ein- und Ausgabeweg mit seiner Vorschau (AK-02, AK-04)

**Vorher** heißt: was zu sehen ist, bevor etwas geschrieben wird. **Nachher**:
was die Anwendung über das Ergebnis sagt. *Ansehen* ist die Dateiansicht aus
§1, aufgeklappt auf Wunsch (`bausteine/Dateiverweis.tsx`) — gelesen wird die
Datei erst dann.

### Ausgabe

| Weg | Wo | Vorher: Ziel, Umfang, Zuschnitt | Nachher |
| --- | --- | --- | --- |
| **Set speichern** (`.mse-set`) | *Sets* | Ziel über den Dialog von macOS; Umfang ist das ganze Set — ein Teilset wäre ein anderes Set (MSE2 ebenso) | Pfad, Befunde, **Ansehen** |
| **Bilder** (PNG je Karte) | Lasche *Bilder* | Ordner, Kartenauswahl (`Ausgabeauswahl`), Auflösung | Zahl, Übersprungene, Befunde, **Ansehen** des Ordners |
| **Druckbogen** (PDF) | Lasche *Druckbogen* | Kartenauswahl, Papier, Anschnitt, Schnittmarken — der **Zuschnitt** | Karten und Seiten, Befunde, **Ansehen** mit Seitenzahl |
| **Exportvorlage** (HTML, Text) | *Ausgeben* | Vorlage, ihre Optionen, Kartenauswahl | Pfad, Befunde, **Ansehen** als Text |
| **MTG JSON** | *Ausgeben* | Ziel über den Dialog | Karten, Verluste als Befunde, **Ansehen** als Text |
| **Sammlung als Liste oder CSV** | *Sammlung* | Der Text steht im Feld, bevor er kopiert wird | — |
| **Bestand sichern** | *Bestandspflege* | Kennwort, Umfang nach `docs/SICHERUNG.md` | Meldung mit dem Ort |
| **Diagnosepaket** | *Diagnosepaket* | Jeder Teil mit Größe und Inhalt, **vor** der Zustimmung | Pfad |
| **Einstellungen ausführen** | *Einstellungen* | Welche Abschnitte hinausgehen, der Arbeitsplatz wählbar | — |

### Einfuhr

| Weg | Wo | Vorher: Objekte, Konflikte, Abhängigkeiten, Änderungen | Nachher |
| --- | --- | --- | --- |
| **Liste einlesen** | *Sammlung* | Jede Zeile mit ihrem Druck, mehrdeutige zur Wahl, unbekannte als übergangen — **ohne zu schreiben** (`ListeDeuten`) | neu, erhöht, Stück, Befunde |
| **Deckliste** | *Decks* | Welche Zeile welchen Druck meint und welche offen bleibt; der Stand des Decks geht mit (§7, Regel 2) | gelegte Karten |
| **Einstellungen einlesen** | *Einstellungen* | Jede Einstellung mit altem und neuem Wert, einzeln abwählbar | Bericht |
| **Sicherung zurückspielen** | *Bestandspflege* | Sicherungsschau: Umfang, Pakete enthalten oder fehlend, verschlüsselt; Rücksprunglage mit Unterschieden zum jetzigen Bestand | Bericht, Rücksprung |
| **Paket aus dem Puffer prüfen** | *Vorlagenwerkstatt* | Befund mit Zeile und Feld; die letzte gültige Vorschau bleibt als veraltet stehen | — |
| **Handelsumschlag** | *Handeln* | Objekte, Abhängigkeiten, Konflikte und wer unterschrieben hat — vor der Übernahmefreigabe (#1457) | Übernahme |
| **Sammlung aus CSV** | *Sammlung* | Der Text im Feld — **keine Trockenprobe** (Lücke, unten) | neu, erhöht, Stück, Befunde |
| **Paket installieren** | *Pakete* | Die Vorschau des Dateidialogs von macOS — **keine eigene** (Lücke, unten) | Paket samt Abhängigkeiten (`braucht`) |
| **Set aus Datei** | *Sets*, Finder | Die Vorschau des Dateidialogs — **keine eigene**; ein Set wird **neu** angelegt, nichts Vorhandenes ändert sich | Set mit Befunden |
| **Partie einführen** | *Spielen* | keine; eine Partie wird **neu** angelegt | Meldung |

### Rückkehr zum Fachobjekt

**Jede Ausgabe liegt dort, wo das Fachobjekt offen ist**: Bilder, Druckbogen
und Ausgeben sind Laschen am Set; die Erfolgsmeldung steht unter dem Knopf,
der sie ausgelöst hat, und die Karten sind eine Lasche weiter. Eine eigene
„Zurück"-Handlung wäre derselbe Weg zweimal.

### Was macOS dazu tut (AK-04)

| Handlung | Stand | Grund |
| --- | --- | --- |
| **Im Finder zeigen** | für jede Datei, die es gibt | `open -R` startet nichts |
| **Öffnen** | nur, wo §1 es erlaubt | `open` startete sonst ein Programm |
| **Öffnen mit …** (Programmwahl) | **nicht gebaut** | Die Wahl eines Programms ist die Wahl dessen, was ausgeführt wird; wer sie braucht, nimmt *Im Finder zeigen* und das Kontextmenü des Finders, das genau das bietet |
| **Teilen** | **nicht gebaut** | Das Teilenblatt (`NSSharingServicePicker`) braucht einen Anker im nativen Fenster; Wails 3 reicht keinen durch. Der Finder bietet *Teilen* an derselben Datei |

**Beides sind Plattformausnahmen im Sinne des Issues** („wo Plattform und
Berechtigung dies erlauben") und keine stillen Lücken.

## 3 Was offen bleibt

- **Sammlung aus CSV ohne Trockenprobe** — **erledigt mit DOM-03.AK-01**
  (#1698):
  `SammlungCSVVorschauen` sagt je Zeile, was das Zurücklesen täte, und
  schreibt nichts.
- **Paket installieren und Set aus Datei ohne eigene Vorschau davor** —
  **erledigt**: vor einer Paketänderung steht seit DOM-01.AK-02 der Plan
  (#1696),
  vor dem Einlesen eines Sets seit FMT-03 die Vorschau — Befunde und fehlende
  Pakete **davor**, und bei einer sauberen Datei ohne Klick dazwischen.

**Offen bleibt der Ausgabeassistent** des Artboards *Ausgeben*
(`docs/PROTOTYP-0.6.0.md`): er ist eine Neugestaltung der Oberfläche und
gehört zu UXR-09
(#1917);
die Aufträge mit Fortschritt und Abbruch, auf denen er steht, bringt FMT-03
(#1692).
