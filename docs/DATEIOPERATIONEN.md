# Dateioperationen

**Was jede Fläche mit einer MSE-Datei tun kann — Eingabe, Rechte, Grenzen,
Ergebnis, Fehler und Nebenwirkung je Operation.** FMT-03.AK-01
(#1692).

Die **Fähigkeiten** stehen in [`FAEHIGKEITEN.md`](FAEHIGKEITEN.md), die
**Rollen** in [`AUTORISIERUNG.md`](AUTORISIERUNG.md), die Operationen je
Paketart in der Tabelle darunter (FMT-01). Hier steht, was ein
Aufrufer über den **Vertrag** einer Dateioperation wissen muss, bevor er sie in
ein Skript schreibt. Ein Wächter
(`internal/werkzeuge/dateioperationen_test.go`) hält jeden Weg dieser Seite
gegen die eingetragenen Wege des Dienstes und jeden Auftragsweg gegen diese
Seite; **was er nicht hält**, sind die Sätze darunter.

## 1 Was für alle gilt

### Pfade sind kein Freibrief

| Fläche | Wessen Pfad | Wer urteilt |
| --- | --- | --- |
| Kommandozeile | ein Pfad **dieses Rechners**, mit den Rechten dessen, der tippt | niemand außer dem Betriebssystem |
| MCP über stdio | derselbe — der Werkzeugserver läuft als Kindprozess dessen, der ihn startet | niemand außer dem Betriebssystem |
| MCP über HTTP | ein Pfad **des Rechners, auf dem der Server läuft** | der **Pfadvertrag** (OPS-02.AK-04): das Zuhause ohne seine verborgenen Winkel, die eingestellten Ordner, der Ort der Pakete und des Bestands, der Zwischenordner und `MSE_DIENSTORDNER`; aufgelöst über Verknüpfungen |
| Dienst (REST) | dito — **oder gar keiner**: die Datei selbst im Rumpf, die Ausgabe zum Abholen (§3) | dito — **vor** dem Weg, an jedem Feld, das einen Pfad trägt (`domain.IstPfadfeld`) |
| Fenster | kein Pfad: der Dialog von macOS wählt | der Mensch |

**Welche Felder einen Pfad tragen** — seit
#2073, der
Analyse nach 0.6.0: die Namen aus `domain.IstPfadfeld` (`pfad`, `ziel`,
`datei`, `verzeichnis`, `bildpfad`) **in jeder Schreibweise**, denn
`encoding/json` legt `"Ziel"` in dasselbe Feld wie `"ziel"`, auch mit
`DisallowUnknownFields`. Dazu `quelle` und `ort` an den Wegen und Werkzeugen,
die darunter einen Pfad nehmen — Pakete einbinden, ersetzen, planen und
übernehmen, beim Dienst auch die Sammeldatei des Indexaufbaus. Bis dahin kam
ein Pfad unter `"Ziel"` oder `quelle` ungeprüft an den Weg.

**Geprüft wird die oberste Ebene eines Rumpfs**, weil darunter freie Daten
stehen — `ziel` eines Angriffs ist ein Spieler. Dass kein Pfadfeld tiefer
steht und jedes `quelle`, `ort` und `bild` entschieden ist, hält
`internal/werkzeuge/pfadstrukturen_test.go`.

**MCP über stdio ist kein offener Dienst**: wer ihn startet, hat die Datei
ohnehin. Über HTTP gilt, was für den Dienst gilt — auch auf der Schleife, denn
dort fragt jeder lokale Prozess.

### Schreiben überschreibt nicht, wo niemand hinsieht

Über **MCP und REST** schreibt eine Ausgabe nur dorthin, **wo nichts liegt**
([0044](entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)); die
Kommandozeile und das Fenster ersetzen, denn dort sieht jemand seinen Aufruf.
Geschrieben wird **atomar** — neben dem Ziel und dann umbenannt; eine halbe
Datei steht nie unter dem Zielnamen.

### Lange Läufe: Fortschritt, Abbruch, Auftrag

| Fläche | Fortschritt | Abbruch | Wiederholen |
| --- | --- | --- | --- |
| Kommandozeile | eine Zeile am Terminal (nur dort) | ⌃C — Rückgabewert 130 | derselbe Befehl |
| MCP | `notifications/progress` auf `_meta.progressToken`; über HTTP als Ereignisstrom der Antwort | stdio: `notifications/cancelled` mit `requestId`, ohne Antwort; HTTP: den Strom schließen | derselbe Ruf |
| Dienst | am Auftrag, wenn mit `Prefer: respond-async` gefragt (`GET /auftraege/{auftragID}`) | `DELETE /auftraege/{auftragID}` | `POST /auftraege/{auftragID}/wiederholung` — durch den Pfadvertrag von **jetzt** |
| Fenster | an der Seite, die den Lauf anstieß, und unter den Hintergrundaufgaben (`auftrag:stand`) | der Knopf *Abbrechen* daneben (`AuftragAbbrechen`) | derselbe Knopf wie zuvor |

**Ein Abbruch lässt nur Ganzes stehen** (AK-04): der Fachkern fragt zwischen
zwei Stücken (`service.Schritt`) — vor der nächsten Karte eines Bogens, vor dem
nächsten Set einer MTG-JSON-Datei, vor dem nächsten Paket eines Installers,
vor dem nächsten Bild einer Exportvorlage. Was dann gilt, steht je Operation
unter *Nebenwirkung*.

### Grenzen

| Was | Grenze | Wo |
| --- | --- | --- |
| Rumpf einer Anfrage am Dienst | 1 MiB (JSON) | `internal/rest`, `rumpfgrenze` |
| eine übertragene Datei am Dienst | 1 GiB, 30 Minuten | `rest.Hochladegrenze`, `rest.Hochladefrist` |
| eine Ausgabe zum Abholen | eine Stunde, 30 Minuten zum Abholen | `app/dienstablage.go`, `rest.Abholfrist` |
| Antwort des Dienstes | 32 MiB | `domain.Vorgabegrenzen` |
| eine Datei in einem Paket, ausgepackt | 256 MiB | `msepaket.GrenzePaketdatei` |
| ein Bild | 32 768 Punkte je Seite, 100 Millionen Punkte | `msepaket.GrenzeBildseite`, `GrenzeBildpunkte` |
| ein Installer | 4 GiB ausgepackt, 131 072 Einträge | `service/paketeinlesen.go` |
| eine Nachricht an MCP | 8 MiB | `mcp.Hoechstenachricht` |
| Aufträge am Dienst | vier laufende, sechzehn beendete für eine Stunde | `internal/rest/auftraege.go` |

### Fehler

| Klasse | Dienst | MCP | Kommandozeile |
| --- | --- | --- | --- |
| der Aufruf selbst ist falsch — ein fehlendes Argument, ein unbekannter Befehl | `400 validierung` | `isError` mit dem Satz | **2** |
| die Datei ist unlesbar, kein Set, oder die Eingabe verletzt eine Regel | `400 validierung` | `isError` mit dem Satz | **1** — „die Antwort ist nein" |
| gibt es nicht — auch eine Datei, die fehlt | `404 nichtGefunden` | `isError` | 1 |
| läuft noch, Konflikt | `409 laeuft` / `konflikt` | `isError` | 1 |
| Rolle fehlt, nur lesend | `403 erlaubnis` / `nurLesend` | `isError` | 1 |
| Pfad außerhalb des Vertrags | `400 validierung` | `isError` (nur über HTTP) | — |
| Rumpf zu groß | `413 validierung` | stdio: die Leitung endet; HTTP: `400` | — |
| zu viele Anfragen, zu viele Aufträge — auch ein Kontingent, das der Fachkern meldet (Nachlauf über Phase 1 von 0.9.0, Befund 4) | `429 drossel` / `auftraege` mit `Retry-After` | `isError` mit Wartezeit, `structuredContent.fehler` mit Code `drossel` und `sekunden` | — |
| Bestand belegt | `503 belegt` mit `Retry-After` | `isError` | 1 |
| Platte voll | `507 voll` | `isError` | 1 |
| abgebrochen | `499 abgebrochen` | keine Antwort | 130 |
| alles andere | `500 unbekannt` — **ohne** Pfad und ohne SQLite-Satz (§8) | `isError` | 1 |

**Eine Datei, die fehlt, keine Leserechte hat oder deren Inhalt kaputt ist,
ist nie „unbekannt"**: bis FMT-03 gingen eine fehlende MTG-JSON-Datei, jede
Datei ohne Leserecht und abgeschnittenes MTG JSON als `500` hinaus.

## 2 Die Operationen

Je Operation: die drei Wege ohne Fenster, dann Eingabe, Rolle, Ergebnis,
Nebenwirkung. **Auftrag** heißt: am Dienst mit `Prefer: respond-async`
möglich.

### Die Karten eines Sets prüfen

`mse probleme <set>` · `lesen.set.probleme.pruefen {setID}` · `POST /sets/{setID}/probleme` — **Auftrag**

- **Keine Datei**: geprüft wird, was im Bestand steht, gezeichnet wie die
  Vorschau. **Sie steht hier, weil die Marke Auftrag hier geführt wird** —
  sie zeichnet jede Karte und kann die Schreibfrist des Dienstes überdauern
  (PERF02, #2352;
  Zweitblick über Phase 3, Befund 11).
- **Ergebnis**: der Stand mit Urteil (`aktuell`, `veraltet`,
  `unvollstaendig`, `ungeprueft`); nur bei `sicher` ist die Zahl eine, die
  gilt. **Ein zweiter Lauf über dasselbe Set löst den ersten ab**, und der
  erste endet mit `konflikt`, ohne seinen Stand abzulegen.
- **Nebenwirkung**: keine im Bestand; der Stand bleibt im Prozess, höchstens
  64 Sets (`service.ProblemstaendeHoechstens`).

### Ein Set vorab ansehen

`mse set-vorschau <pfad>` · `lesen.set.vorschau {pfad}` · `POST /sets/vorschau {"pfad"}`

- **Rolle**: leser (`lokalLesen`).
- **Ergebnis**: `pfad`, `name`, `spiel`, `stil`, `karten`, `befunde`,
  `schonOffen` (Kennung oder leer), `bedarf` (fehlende oder zu alte Pakete).
- **Fehler**: dieselben wie beim Einlesen, mit **demselben** Satz.
- **Nebenwirkung**: keine.

```bash
curl -s -X POST -H "Authorization: Bearer $MARKE" \
  -d '{"pfad":"/Users/ich/Sets/probe.mse-set"}' http://127.0.0.1:$PORT/sets/vorschau
```

### Ein Set einlesen

`mse oeffnen <pfad>` · `schreiben.set.einfahren {pfad}` · `POST /sets {"pfad"}` — **Auftrag**

- **Rolle**: mitarbeiter (`lokalSchreiben`).
- **Ergebnis**: `setID`, `karten`, `befunde`; `201`. Liegt dieselbe Quelle schon
  im Bestand, kommt **dieses** Set zurück und kein zweites.
- **Nebenwirkung**: ein Set im Bestand, in **einer** Transaktion — ein Abbruch
  davor legt nichts an.

### Ein Set sichern

`mse sichern <set> <ziel>` · `schreiben.set.sichern {setID, ziel}` · `POST /sets/{setID}/sicherung {"ziel"}` — **Auftrag**

- **Rolle**: mitarbeiter.
- **Ergebnis**: `befunde`; `201`.
- **Nebenwirkung**: eine `.mse-set`-Datei, atomar; über MCP und REST nur, wo
  nichts liegt.
- **Nicht darin**: Decks, Exemplare, Preise, Marken und die Herkunft einzelner
  Werte — sie bleiben im Bestand, denn MSE2 verwirft fremde Angaben beim
  Speichern ([0145](entscheidungen/0145-zusatzdaten-werden-nicht-eingebettet.md),
  [`ZUSATZDATEN.md`](ZUSATZDATEN.md)).

### Ein Kartenbild schreiben

`mse zeichnen <karte> <ziel>` · `schreiben.karte.zeichnen {karteID, ziel, guete}` · `POST /karten/{karteID}/bild {"ziel","guete"}` — **Auftrag**

- **Rolle**: mitarbeiter. **Ergebnis**: `breite`, `hoehe`, `befunde`.
- **Nebenwirkung**: ein PNG, JPEG oder SVG nach der Endung des Ziels.

### Einen Druckbogen legen

`mse bogen <set> <ziel>` · `schreiben.set.bogen {setID, ziel}` · `POST /sets/{setID}/bogen {"ziel"}` — **Auftrag**

- **Rolle**: mitarbeiter. **Ergebnis**: `seiten`, `karten`, `befunde`.
- **Fortschritt**: je Karte. **Nebenwirkung**: ein PDF, erst am Ende
  geschrieben — ein Abbruch hinterlässt keine Datei.
- **Der Probelauf davor** (UXR-F7, #2015):
  `mse bogen <set> [<ziel>] --probe` · `lesen.set.bogenprobelauf {setID, ziel?}` · `POST /sets/{setID}/bogen/probelauf {"ziel"?}`.
  **Rolle**: leser — er fragt den Schreibriegel nicht. **Ergebnis**: `karten`,
  `seiten`, `jeSeite`, `spalten`, `zeilen`, `aufLetzterSeite`, `bildbreite`,
  die Maße und `probelauf: true`. **Nebenwirkung**: keine. Dieselbe Prüfung von
  Maßen, Set und Auswahl wie der Lauf (ein Plan, den beide durchlaufen); ein
  genanntes Ziel wird über MCP und REST **nur angesehen**: liegt dort schon
  etwas, antwortet er wie der Lauf. **Nicht darin**: was erst das Zeichnen
  zeigt — ein Plan, der sich nicht bauen lässt, eine verzerrte Vorlage.

### Über eine Exportvorlage ausgeben

`mse vorlagen <set> ausfuehren …` · `schreiben.vorlage.ausfuehren` · `POST /sets/{setID}/vorlagen` — **Auftrag**

- **Rolle**: mitarbeiter. **Fortschritt**: je geschriebenem Bild, ohne
  Gesamtzahl — wie viele es werden, weiß nur das Skript.
- **Nebenwirkung**: die Ausgabe der Vorlage; **geschriebene Bilder bleiben bei
  einem Abbruch liegen**, wie bei MSE2, und der letzte gemeldete Stand nennt
  ihre Zahl — am Fenster als „Bis dahin: 3 Bilder".
- **Der Probelauf davor** (UXR-F7, #2015):
  `mse vorlagen <set> ausfuehren <paket> [<ziel>] --probe` · `lesen.vorlage.probelauf {setID, paket, ziel?, wahl?}` · `POST /sets/{setID}/vorlagen/probelauf {"paket","ziel"?,"wahl"?}`.
  **Rolle**: leser. **Ergebnis**: `karten`, `dateiname`, `bytes`, `beigaben`,
  `beigabenBytes`, `verzeichnis`, `geschaetzt`, `befunde`. **Nebenwirkung**:
  keine — die Vorlage wird durch denselben Weg **ausgewertet**, ihre Beigaben
  (`write_image_file`, `copy_file`, `write_text_file`, `write_set_file`) tragen
  sich in eine Ablage ein statt auf die Platte, und ein `-files`-Ordner wird
  nicht angelegt. **Die Größe der Hauptdatei ist genau**; die der Beigaben eine
  Untergrenze (`geschaetzt`), wo ein Kartenbild erst beim Rastern entsteht.
  **Nicht darin**: das Rastern der Kartenbilder, ein Ziel, das sich nicht
  anlegen ließe (Platte), und die Maße der Bilder, die eine Vorlage liest.

### MTG JSON lesen und schreiben

`mse mtgjson-lesen` · `schreiben.mtgjson.lesen` · `POST /mtgjson {"pfad","spiel"}` — **Auftrag**<br>
`mse mtgjson-schreiben` · `schreiben.mtgjson.schreiben` · `POST /sets/{setID}/mtgjson {"spiel","ziel"}` — **Auftrag**

- **Rolle**: mitarbeiter.
- **Nebenwirkung beim Lesen**: je Set eine Transaktion; ein Abbruch fällt
  zwischen zwei Sets, und das Ergebnis nennt, was schon angelegt ist.

### Pakete als Installer weitergeben und einlesen

`mse paket-weitergeben` · `schreiben.paket.weitergeben` · `POST /pakete/weitergabe {"namen","ziel"}` — **Auftrag**<br>
`mse installer-einlesen` · `schreiben.paket.einlesen` · `POST /pakete/eingelesen {"datei"}` — **Auftrag**

- **Rolle**: Pakete gehören dem Rechner (`lokalSchreiben`).
- **Nebenwirkung beim Weitergeben**: eine `.mse-installer`-Datei, erst am Ende
  geschrieben. **Beim Einlesen**: je Paket ein Verzeichnis im verwalteten Ort;
  was schon daliegt, wird übergangen; bei einem Abbruch sind die Pakete davor
  ganz **und vermerkt**.

### Eine Datei prüfen und ansehen

`mse paketdatei-pruefen` · `lesen.skriptbefund` · `POST /paketdateien/pruefung`<br>
`mse datei-ansehen <pfad>` · `lesen.datei.ansicht {pfad}` · `GET /dateien/ansicht?pfad=…`

- **Rolle**: leser. **Nebenwirkung**: keine — ausgepackt, gefolgt oder
  ausgeführt wird nichts.

## 3 Ohne Pfad: die Datei selbst

**Ein Pfad ist eine Auskunft über den Rechner, auf dem der Dienst läuft.** Wer
ihn jenseits der Schleife ruft (N6, mit TLS und Zugang), hat dort keinen. Er
schickt deshalb die **Datei selbst** und holt eine Ausgabe ab, statt sie
irgendwohin schreiben zu lassen. FMT-03.AK-03.

### Hin

| Weg | Rumpf | Geht weiter als |
| --- | --- | --- |
| `POST /sets` | die `.mse-set`-Datei | `{"pfad"}` |
| `POST /sets/vorschau` | dito | `{"pfad"}` |
| `POST /pakete/eingelesen` | die `.mse-installer`-Datei | `{"datei"}` |

- **Art**: `Content-Type: application/zip` oder `application/octet-stream`;
  jede andere Art ist die JSON-Anfrage von oben.
- **Wohin**: in den **Eingang** neben dem Bestand (`dienst/eingang`), **nach
  dem Inhalt benannt und dauerhaft** — ein Set liest seine Bilder später aus
  seiner Quelle, und dieselbe Datei ein zweites Mal hat denselben Ort und gibt
  **dasselbe** Set zurück. **Die Vorschau legt nur zwischen**
  (`dienst/vorschau`) und nimmt die Datei nach der Antwort wieder weg — nicht
  im Eingang, denn dort nähme sie womöglich einem Set seine Quelle.
- **Aufgeräumt** wird der Eingang höchstens stündlich: was älter als eine
  Stunde ist und worauf **kein Set irgendeines Profils** mehr verweist, geht
  (#1944).
  Die Stunde schützt ein laufendes Einfahren und die Wiederholung eines
  Auftrags; wer nicht sicher sagen kann, ob verwiesen wird, behält die Datei.
- **Danach** geht die Anfrage mit diesem Pfad weiter, durch den Pfadvertrag wie
  jede andere, und als Auftrag, wenn darum gebeten.
- **Grenze**: 1 GiB, sonst `413`; eine abgerissene Übertragung lässt nichts im
  Eingang liegen.
- **Frist**: der Kopf in zehn Sekunden, der Rumpf **hinter der Marke** in
  dreißig Minuten — gemessen: mit der Frist des Kopfes brach ein Set über eine
  langsame Leitung nach zehn Sekunden ab (`internal/flaechen/uebertragung_test.go`).

```bash
curl -s -X POST -H "Authorization: Bearer $MARKE" -H "Content-Type: application/zip" \
  --data-binary @probe.mse-set https://dienst.example:$PORT/sets
```

### Zurück

| Weg | Ausgabe |
| --- | --- |
| `POST /sets/{setID}/sicherung` | `.mse-set` |
| `POST /karten/{karteID}/bild` | `.png` |
| `POST /sets/{setID}/bogen` | `.pdf` |
| `POST /sets/{setID}/mtgjson` | `.json` |
| `POST /pakete/weitergabe` | `.mse-installer` |

- **Ohne `ziel`** schreibt der Weg in den **Ausgang** (`dienst/ausgang`) und
  antwortet mit `Location: /ausgang/{name}`; mit `ziel` bleibt alles wie oben.
  Als Auftrag steht der Ort am Ergebnis, und eine Wiederholung bekommt einen
  **frischen** — der alte ist womöglich schon abgeholt.
- `GET /ausgang/{name}` holt die Datei, mit ihrer Art und als Anhang.
  **Keine Grenze der Antwort**: ein Bogen über dreihundert Karten ist größer
  als jede JSON-Antwort.
- `DELETE /ausgang/{name}` vergisst sie; **nach einer Stunde** geht sie von
  selbst. Danach, und für jeden Namen, der kein schlichter Dateiname ist,
  `404`.
- **Rolle**: keine eigene, aber **ein Eigner** — die Ausgabe gehört dem
  Principal, der sie anstieß, wie ein Auftrag (SEC-P6,
  #2002).
  Ein anderer — ein Token eines anderen Profils, ein zweites Token desselben,
  auch die Prozessmarke — bekommt `404` wie für einen Namen, den es nicht
  gibt. Der Name bleibt 80 Bit gewürfelt, nur in der Antwort genannt, eine
  Stunde gültig; **er ist nicht mehr das einzige Recht**. Abholen braucht eine
  Marke mit `lesen`, Vergessen eine mit `schreiben`; beide gehen durch die
  Wachen wie jeder Weg. **Ein neu gestarteter Dienst kennt keinen Eigner
  mehr** — eine Ausgabe von davor holt niemand ab.

**Nicht über MCP und nicht an der Kommandozeile**: dort läuft der Aufrufer auf
demselben Rechner (stdio) oder nennt, wie am Dienst, einen Pfad des Servers
(HTTP). Ein MCP-Werkzeug, das Megabyte als Base64 in eine Nachricht von
höchstens 8 MiB packte, wäre ein zweiter, schlechterer Weg.

