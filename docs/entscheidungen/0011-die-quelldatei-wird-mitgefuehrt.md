# 0011 — Die Quelldatei wird mitgeführt

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

#48 sagt zu: ein
unverändert durchgereichtes Set — öffnen, speichern, vergleichen — ist **Byte
für Byte** dasselbe. Das Issue nennt die Frage, die dabei offen steht, selbst:

> Die verstandenen Teile kommen aus `card_values` zurück — **ohne** ihre Rohform.
> Einrückung und Abstände stehen nirgends. Entweder ist der Byte-Vergleich so
> nicht zu halten, oder der Bestand muss die Quelldatei mitführen.

Der Schreiber für den **Baum** ist gebaut und byte-genau (#128): er setzt nur
zusammen, was niemand gelesen hat, und gibt sonst `Knoten.Roh` aus. Was fehlt,
ist der Weg **vom Bestand zum Baum**.

## Entscheidung

**Der Bestand führt die Quelldatei mit.** Ein Set trägt die Bytes der Datei, aus
der es gelesen wurde; geschrieben wird, indem diese Bytes wieder zerteilt, die
geänderten Werte darin ersetzt und der Baum ausgegeben wird.

Ist nichts geändert, kommt die Datei unverändert heraus — nicht, weil wir sie
gut genug nachbauen, sondern weil es dieselbe ist.

## Verworfen

**Die Rohzeile je Wert mitführen.** Eine Spalte `roh` an `card_values` und
`set_values`, gefüllt aus `Knoten.Roh`. Das rettet Einrückung und Abstände einer
Zeile — und **nicht** das, was zwischen den Zeilen steht. Genau davon handelt
#88: vier
benannte Klassen gehen heute verloren, darunter Kommentare und Leerzeilen auf
der Wurzelebene und ein zweites `set info:`. Eine Lösung, die vier bekannte
Verluste stehen lässt, löst das Problem nicht, sondern verkleinert es.

**Die Datei beim Schreiben von ihrem Pfad neu lesen.** `sets.herkunft` steht da,
die Datei liegt dort. Das kostet keinen Speicher und **widerspricht
[0002](0002-sqlite-als-arbeitsformat.md)**: SQLite ist das Arbeitsformat, und
ein Arbeitsformat, das ohne eine Datei daneben unvollständig ist, ist keines.
Wer ein Set öffnet, den Datenträger abzieht und dann speichert, bekäme einen
Fehler über eine Datei, die er nie angefasst hat.

**Die Zusage abschwächen.** „Semantisch gleich statt byte-gleich" wäre ehrlich
formuliert und träfe den Zweck: 0002 gibt die Byte-Zusage, weil ein Set, das
nach einer Bearbeitung in MSE2 nicht mehr öffnet, den Wechsel unumkehrbar macht.
Wer nicht Byte für Byte vergleichen kann, merkt eine solche Änderung nicht.

## Was das kostet

**Speicher.** Die `set`-Datei eines großen Sets wiegt einige hundert Kilobyte;
`Full-Magic-Pack` kennt größere. Bei dreißig geöffneten Sets sind das einige
Megabyte in der Datenbank — messbar, aber nicht in der Größenordnung, in der
jemand es bemerkt. **Die Bilder eines Sets liegen ohnehin nicht darin**, und sie
sind es, die ein `.mse-set` groß machen.

**Zwei Wahrheiten, solange nichts geschrieben ist.** Nach einer Änderung stehen
der Wert in `card_values` und die Zeile in der Quelldatei verschieden da. Das
ist kein Widerspruch, sondern die Arbeitsteilung: die Quelle ist die **Herkunft**,
der Bestand ist der **Stand**. Wer schreibt, bringt beides zusammen — und danach
ist die neue Ausgabe die neue Quelle.

**Ein Set, das in der Anwendung entsteht, hat keine.** Dann wird
zusammengesetzt, wie `msepaket.Schreibe` es für Knoten ohne Rohzeile tut. Die
Byte-Zusage gilt dort nicht — es gibt nichts, womit zu vergleichen wäre.

> **Dieser Absatz stand hier ab dem ersten Tag und war dreizehn Monate lang
> nicht umgesetzt.** `Setschreiber.Schreibe` wies ein Set ohne Quelldatei mit
> `ErrNotFound` ab, und ein Lauf hielt die Abweisung fest — begründet mit
> derselben Byte-Zusage, von der hier steht, dass sie nicht gilt. Damit ließ
> sich **kein** in dieser Anwendung angelegtes Set sichern; gefunden hat es
> erst die native Prüfung von v0.5.0 als Befund **F02**
> (#1746).
>
> **Seit dem 2026-09-20 tut der Schreiber, was hier steht**
> (`internal/service/setneubau.go`). Die Lehre ist nicht, dass der Vermerk
> falsch war — er hatte recht —, sondern dass ein Vermerk keine Zusage hält:
> hier stand die Entscheidung, und der Lauf daneben hielt das Gegenteil fest.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn die Datenbank durch die Quellen spürbar wächst. Dann gehört die Quelle
  komprimiert oder ausgelagert — nicht weggelassen.
- Wenn das Zusammenbringen von Quelle und Stand beim Schreiben regelmäßig
  Abweichungen erzeugt, die niemand angeordnet hat. Dann ist nicht die
  Mitführung falsch, sondern das Ersetzen — und der Weg führt über eine feinere
  Zuordnung Wert → Knoten, nicht über eine schwächere Zusage.
- Wenn sich zeigt, dass MSE2 Dateien akzeptiert, die byte-verschieden sind, und
  die Byte-Zusage damit strenger ist als nötig. Dann ist sie trotzdem die
  einzige, die sich **prüfen** lässt — und das ist ihr Wert.
