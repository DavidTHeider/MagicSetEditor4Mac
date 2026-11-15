# 0029 — Bilder liegen neben dem Set, nicht darin

**Stand:** angenommen
**Datum:** 2026-08-19

## Lage

#247 stellt drei
Fragen, und #246
wartet auf die dritte: ein Kartenfeld der Art `image` lässt sich nicht füllen,
weil niemand gesagt hat, wohin das Bild dann gehört.

Der heutige Stand: die Hauptdatei eines Sets liegt Byte für Byte im Bestand
([0011](0011-die-quelldatei-wird-mitgefuehrt.md)), die Bilder liegen im
**Herkunftspaket** auf der Platte, und `service.beiwerk` kopiert sie beim
Speichern mit. Ein Bild, das jemand in der Anwendung einsetzt, hat kein
Herkunftspaket.

## Entscheidung

### 1 MTG JSON bleibt ein Austauschformat

**Nachgemessen, nicht wiederholt.** Über `Multicolor.mse-set` — 3.008 Karten,
77.463 Feldwerte, 62 verschiedene Feldnamen; weiter reicht die Zahl nicht:

| | |
| --- | --- |
| Namen, die `internal/mtgjson/zuordnung.go` kennt | **11** von 62 |
| Feldwerte, die davon gedeckt sind | **30.047** von 77.463 — 38,8 % |

Was es an **jeder** der 3.008 Karten nicht kennt: `stylesheet`,
`stylesheet version`, `has styling`, `styling data` — also welche Vorlage die
Karte zeichnet und mit welchen Einstellungen —, dazu `image`, `image 2`,
`mainframe image`, `border color`, `notes`, `time created`, `time modified`.

Ein Modell, das an jeder Karte nicht weiß, **womit sie gezeichnet wird**, kann
kein Modell für einen Karteneditor sein. Die Byte-Zusage aus
#48 wäre damit
ohnehin nicht zu halten.

### 2 Die Grafiken bleiben auf der Platte

**Nachgemessen über die zehn kanonischen Sets** — 23.168 Karten; weiter reicht
die Zahl nicht:

| | |
| --- | --- |
| Bilder | **1.996,0 MB** |
| Hauptdateien | **22,9 MB** |
| Anteil der Bilder am Gewicht | **98,9 %** |

Kein Set liegt unter 98,3 %. Sie in den Bestand zu nehmen hieße, eine
SQLite-Datei von zwei Gigabyte zu führen, damit die 23 MB daneben nicht allein
stehen.

### 3 Ein eingesetztes Bild liegt neben dem Set

**In einem Beiwerkverzeichnis je Set** unter
`~/Library/Application Support/MagicSetEditor4Mac/beiwerk/<setID>/`. Beim
Speichern kommt es mit — durch dieselbe Naht, die heute das Herkunftspaket
ausliest (`service.beiwerk`), und **vor** ihm: was jemand eingesetzt hat,
schlägt, was im Herkunftspaket unter demselben Namen lag.

**Der Name folgt MSE.**

    LocalFileName new_image_file = getLocalPackage().newFileName(field().name, "");
    img.SaveFile(getLocalPackage().nameOut(new_image_file), wxBITMAP_TYPE_PNG);

`ImageValueEditor::sliceImage`, `src/gui/value/image.cpp`, Zeilen 49–51,
Fassung f9d9356 vom 2021-07-06, mit

    name = prefix; name << ++infix; name += suffix;
    name = normalize_internal_filename(name);

`Package::newFileName`, `src/util/io/package.cpp`, Zeilen 248–267. Also:
**Feldname, eine laufende Zahl, kleingeschrieben, ohne Endung** — und der
Inhalt ist **immer PNG**, gleich was hereinkam. Nachgeprüft an
`Archenemy.mse-set`: die 70 Bilder darin heißen `image0` bis `image69`.

## Verworfen

**Die Bilder in den Bestand.** Ein Set wäre dann **eine** Datei, und das ist
ein echter Vorteil — er kostet nach der Messung oben das Neunzigfache an
Datenbank. Dagegen steht außerdem [0002](0002-sqlite-als-arbeitsformat.md): die
Byte-Zusage gilt der Paketdatei, nicht einer Kopie davon.

**Das Bild in das Herkunftspaket schreiben**, so wie MSE es tut. MSE **ist** das
Paket; dieses Programm hat den Bestand dazwischen. Ein Herkunftspaket darf
schreibgeschützt sein, auf einem abgezogenen Datenträger liegen oder gar nicht
existieren — ein Set, das hier entstanden ist, hat keines. Und wer ein Set
öffnet, ein Bild einsetzt und **nicht** speichert, hätte die fremde Datei
trotzdem schon verändert.

**Das Bild an seinem ursprünglichen Ort lassen und den Pfad merken.** Das kostet
nichts und ist die schlechteste Antwort: das Set verweist dann auf einen Ort auf
**diesem** Rechner. Weitergegeben öffnet es mit leeren Kunstfenstern, und der
Grund steht in keiner Datei.

**Den Zuschnitt weglassen und das Bild roh übernehmen.** MSE schneidet vor dem
Ablegen zu (`ImageSliceWindow`, Zeilen 44–48) und legt **das zugeschnittene**
Bild ab. Wer roh übernimmt, legt für ein 40-Kilobyte-Kunstfenster eine
12-Megapixel-Aufnahme ins Set. Der Zuschnitt wird deshalb gebaut — in einem
eigenen Werkstück, aber vor dem Speichern.

## Was das kostet

**Ein zweiter Ort, an dem etwas zu einem Set gehört.** Bisher waren es zwei
(Bestand, Herkunftspaket), jetzt sind es drei. Wer ein Set aus dem Bestand
löscht, muss das Verzeichnis mitlöschen; wer es vergisst, sammelt Bilder ohne
Set. **Das ist eine Aufgabe und keine Nebenbemerkung** — sie steht in
#246 als
Abnahmepunkt.

**Ein Set ist ohne sein Beiwerkverzeichnis unvollständig**, solange es nicht
gespeichert wurde. Das gilt heute schon für das Herkunftspaket; neu ist, dass
es nun auch für etwas gilt, das **diese** Anwendung angelegt hat.

**Eine Endung fehlt.** `image1` ohne `.png` sieht nach einer kaputten Datei aus,
wenn jemand das Paket auspackt. Es ist MSEs Form, und ein Set mit `image1.png`
darin öffnete MSE nicht — der Wert im `set` nennt den Namen ohne Endung.
