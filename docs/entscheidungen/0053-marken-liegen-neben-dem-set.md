# 0053 — Marken liegen neben dem Set, nicht darin

**Datum:** 2026-08-27
**Anlass:** Werkstück 3 von Phase 3.5 (#1166) — „Tags und Kategorien, Notizen an einer Karte"
**Stand:** entschieden, gebaut wird es in #1208

## Die Frage

Eine **Marke** ist ein Wort, das jemand an eine Karte hängt, um sie
wiederzufinden: *fehlt noch Kunst*, *fertig*, *für den Draft*. Wo steht sie?

Zwei Antworten liegen nahe, und nur eine trägt.

## Was nachgelesen wurde

**MSE2 kennt nichts dergleichen.** Gesucht am 2026-08-27 in Fassung `f9d9356`:

| Wo | Womit | Ergebnis |
| --- | --- | --- |
| `src/data/*.hpp`, `*.cpp` | `tag`, `category`, `categories` | nur `statistics_categories` (`game.hpp`, Zeile 46) und die Kategorien einer `.mse-locale` (`locale.hpp`) — beides etwas anderes |
| `src/data/card.hpp`, die Klasse `Card` | gelesen | Feldwerte, **Notizen**, zwei Zeitstempel, ein abweichender Stil, Stildaten, Schlüsselwortstatistik — keine Marke |

**Damit ist die Marke `new`** im Sinne von §15, und das ist hier
aufgeschrieben, damit die Kennzeichnung keine unterlassene Suche verdeckt.

## Und das entscheidet die Frage nicht — das tut der Leser von MSE2

Die naheliegende Antwort wäre: als Kind eines `card:`-Blocks, so wie `notes`.
Unser Schreiber gibt jedes Kind wieder heraus, ein Rundlauf hier bliebe
byte-genau.

**Nur liest MSE2 die Datei anders.** `Reader::unknownKey()`:

```cpp
void Reader::unknownKey() {
  if (ignore_invalid) { do { moveNext(); } while (indent > expected_indent); return; }
  if (indent >= expected_indent) {
    warning(_("Unexpected key: '") + key + _("'"), 0, false);
    do { moveNext(); } while (indent > expected_indent);
  }
}
```

`src/util/io/reader.cpp`, Zeilen 231–246, Fassung `c9c708d` vom 2020-05-27.

**Ein unbekannter Schlüssel wird übersprungen, nicht aufgehoben.** Wer eine so
markierte Datei in MSE2 öffnet, bekommt eine Warnung — und wer sie dort
**speichert**, hat die Marken verloren. Still, denn beim Speichern warnt
niemand mehr.

Eine Datei, die in dem Programm, dessen Format sie trägt, beim Öffnen warnt und
beim Speichern Daten verliert, ist keine gute Datei.

## Die Entscheidung

**Marken stehen im Bestand, neben dem Set — nicht in der `.mse-set`.**

Es ist dieselbe Trennung wie bei der Sammlung
([0052](0052-die-sammlung-ist-nicht-das-set.md)), und aus einem verwandten
Grund: eine `.mse-set` gibt man weiter. Was jemand sich zum Wiederfinden
notiert, geht niemanden etwas an, dem er das Set schickt.

**Die Notizen sind der Gegenfall, und er zeigt, wo die Grenze liegt.** `notes`
steht **in** der Datei — weil MSE2 das Feld kennt, es mitschreibt und es beim
Öffnen zeigt. Es ist Teil des Werks, nicht des Arbeitsplatzes.

## Was das kostet

**Marken reisen nicht mit dem Set.** Wer sein Set auf einen zweiten Rechner
trägt, hat dort seine Karten und nicht seine Marken. Die Sicherung des Bestands
nimmt sie mit (`BestandSichern`), die `.mse-set` nicht.

Das ist derselbe Preis wie bei der Sammlung, und er ist derselbe Tausch:
lieber etwas, das nicht mitreist, als etwas, das jemand ungewollt mitschickt.

## Verworfen

| Weg | Warum nicht |
| --- | --- |
| **Als Kind von `card:`** in der `.mse-set` | MSE2 verwirft es beim Speichern — siehe oben. Eine Zusage, die nur gilt, solange niemand die Datei mit dem Programm öffnet, für das sie gemacht ist |
| **Als Kartenfeld des Spielpakets** (`card field: tags`) | Es wäre **je Spielpaket** zu erklären. Marken gäbe es dann nur, wenn die Vorlage daran gedacht hat — und ein bezogenes Paket hat es nicht |
| **In einer Datei neben der `.mse-set`** | Sie ginge beim Verschieben verloren, und zwei Dateien, von denen eine unsichtbar ist, verwechselt man einmal |
