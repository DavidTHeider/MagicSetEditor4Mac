# 0002 — SQLite als Arbeitsformat, MSE-Dateien als Austauschformat

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

MSE2 hält ein Set im Arbeitsspeicher und schreibt es als ZIP-Archiv. Das ist
einfach und hat drei Grenzen, die erst bei großen Sets auffallen: keine Suche
über Kartentexte, keine Massenbearbeitung mit Rücknahme, und ein Rückgängig,
das mit dem Fenster endet.

Zugleich gilt die Zusage aus dem Zielvorschlag: vorhandene MSE2-Daten bleiben
nutzbar. Was hineingeht, muss wieder herauskommen.

## Entscheidung

**SQLite ist das Arbeitsformat, die MSE-Datei das Austauschformat.**

```
.mse-set  ──import──►  SQLite  ──export──►  .mse-set
```

Import beim Öffnen, Export beim Speichern. Suche, Massenänderung und Verlauf
sind Datenbankarbeit.

Als Treiber `modernc.org/sqlite` — SQLite ohne cgo. Das hält den Bau frei von
einem C-Compiler und macht die Kreuzkompilierung für Windows und Linux möglich,
ohne dass dort jemand etwas einrichtet (`docs/PORTIERUNG.md`).

Dazu gehört eine **Rundlaufzusage**: was gelesen wurde, wird unverändert
zurückgeschrieben — auch was wir nicht verstehen. Unbekannte Schlüssel und
Bereiche werden roh mitgeführt (`raw_blocks`) und beim Schreiben an ihrer Stelle
wieder ausgegeben.

**Zwei Zeitpunkte, die nicht verwechselt werden dürfen** — der Zweitblick über
Phase 2 hat sie hier und in `docs/ARCHITEKTUR.md` widersprüchlich vorgefunden:

| Was | Wann | Warum |
| --- | --- | --- |
| Die **Daten** verlustfrei mitführen | **ab Phase 2**, nicht nachrüstbar | wer unbekannte Bereiche erst wegwirft, hat sie in jeder bereits importierten Datenbank verloren |
| Der **Rundlauftest** über echte Pakete — lesen, schreiben, Byte für Byte | **Phase 6** (#48) | es gibt vor Phase 6 keinen Schreiber, gegen den er laufen könnte |

Was Phase 2 hält, ist deshalb die schwächere und wichtigere Hälfte: der Knoten
merkt sich seine **Rohzeile** samt Zeilenende, und `raw_blocks` bekommt sie
wörtlich.

Was Phase 2 **nicht** hält, ist mehr, als hier zunächst stand. Der erste Anlauf
nannte nur „Kommentare und Leerzeilen innerhalb verstandener Blöcke"; der
Nachlauf des Zweitblicks hat vier weitere Fälle aufgezählt, und alle haben
dieselbe Ursache: ein Schlüssel, den der Leser kennt, schließt seinen Bereich
vom Rohweg aus — auch das Vorkommen, das der Leser gar nicht verwertet.

Zweites `game:`, zweites `stylesheet:`, zweites `set info:`, Kommentare auf der
Wurzelebene, der skalare Rest eines `card: irgendwas`. Aufgezählt in
#88, fällig
gemacht von #48:
der Byte-Vergleich ist der Test, der diese Liste entweder bestätigt oder als
unvollständig erweist.

## Verworfen

**Direkt auf der Datei arbeiten, wie MSE2.** Verworfen — dann gäbe es keinen der
drei Vorteile, und der Neubau hätte an dieser Stelle keinen Grund.

**SQLite als *Speicherformat*, die MSE-Datei nur als Import.** Verworfen, und
das ist die wichtigere Ablehnung: es hätte den Rundlauf erspart und damit die
teuerste Zusage des Projekts. Aber ein Set, das nach einer Bearbeitung nicht mehr
in MSE2 zu öffnen ist, macht den Wechsel unumkehrbar — und ein unumkehrbarer
Wechsel ist für ein Werkzeug, an dem Leute jahrelang gearbeitet haben, keine
Zumutung, die man ihnen zumuten darf.

**Eine Spalte je Kartenfeld.** Verworfen. MSE-Spiele definieren ihre Felder
selbst; eine Spalte je Feld hieße, für jedes fremde Spiel zu migrieren. Statt
dessen `card_values` mit Feldname und Wert als Text. Der Preis: Typisierung steht
nicht in der Datenbank, sondern im Spielpaket — dort steht sie ohnehin.

**`mattn/go-sqlite3`** (der verbreitete cgo-Treiber). Verworfen wegen cgo:
er brächte einen C-Compiler in jeden Bau und beendete die Kreuzkompilierung.

## Folgen

- Öffnen eines Sets kostet einen Import. Bei großen Sets muss das sichtbar
  fortschreiten statt zu blockieren.
- Die Rundlaufzusage ist ab Phase 2 mitzubauen und **nicht** nachzurüsten: wer
  unbekannte Felder erst wegwirft und später mitführen will, hat sie in jeder
  bereits importierten Datenbank verloren.
- Volltextsuche über FTS5, per Trigger nachgeführt.
- Die Datenbank liegt neben dem Set, nicht darin. Wo genau, entscheidet Phase 2.
