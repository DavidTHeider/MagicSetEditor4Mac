# 0014 — Eine MSE-Datei ist nicht notwendig ein Baum

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

#172 fragt es
gerade heraus: „Ob eine Datei, deren Inhalt Skript ist, überhaupt wie ein
Paketbaum gelesen werden soll."

Der Zerteiler liest `schlüssel: wert` und Einrückung. Seit
#170 ist eine
**eingerückte** Zeile ohne Doppelpunkt eine Fortsetzung, und innerhalb einer
offenen Klammer gilt das auch ohne Einrückung. Was blieb, war die Wurzelebene
außerhalb jeder Klammer — dort war eine Zeile ohne Doppelpunkt ein **Fehler**,
und der Rundlauf über ein bezogenes Pack war deswegen rot.

## Was gezählt wurde

`M15-Magic-Pack`: **325 Pakete, 555 Textdateien.** Zwölf scheitern, und sie
sind es, um die es geht:

| Datei | Was drinsteht |
| --- | --- |
| `magic.mse-game/script` | ein Skriptrumpf — `crop_safe :=` und darunter `{` auf der Wurzelebene |
| `magic.mse-game/language` | dasselbe Muster, `complete_sub_type_list :=` |
| `magic.mse-game/statistics_script` | dasselbe Muster |
| `magic.mse-game/unit_tests` | reine Anweisungen: `assert(card_color_color_count("") == 0)` |
| `magic-modules.mse-include/loyalty/init_script`, `…_dfc` | Skriptrümpfe |
| `magic-m15-invocation…`, `magic-m15-sketch.mse-style/masterpiece_script` | Skriptrümpfe |
| `magic-blends.mse-include/new-blends` | Skriptrumpf |
| `magic-lackeybot.mse-export-template/script` | ein Programm |
| `magic-spoiler.mse-export-template/script js` | **JavaScript** |
| `magic-spoiler.mse-export-template/style css` | **CSS** |

**Keine einzige davon ist eine Hauptdatei.** Nicht `game`, nicht `style`, nicht
`export-template` — nachgezählt, nicht angenommen. Die Hauptdatei eines Pakets
ist ein Baum; was daneben liegt, ist, was es ist.

Bei `style css` erübrigt sich jede Feinheit: CSS ist kein `schlüssel:
wert`-Baum, und es sieht auch nicht so aus.

## Entscheidung

**Eine Zeile ohne Doppelpunkt ist Text — auch auf der Wurzelebene.** Der
Zerteiler weist sie nicht mehr ab; sie wird `ArtFortsetzung` und behält ihre
Bytes in `Roh`.

Damit liest der Zerteiler **jede** Datei eines Pakets, und der Rundlauf ist
grün, ohne dass irgendwo eine Ausnahme steht.

## Was das kostet, und der Preis gehört ausgesprochen

**Der Zerteiler kann eine Datei jetzt fast nicht mehr abweisen.** Die
Wurzelzeile ohne Doppelpunkt war einer seiner zwei Gründe dafür; übrig bleibt
der leere Schlüssel (`: wert`).

Eine `game`-Datei, in der eine Zeile ihren Doppelpunkt verloren hat, wird ab
jetzt gelesen: die Zeile wird Text, und was darunter eingerückt stand, hängt
danach eine Ebene zu hoch. **Das fällt nicht mehr beim Öffnen auf, sondern erst
an einem Feld, das fehlt.**

Vertretbar ist das aus zwei Gründen. Erstens sagt `AGENTS.md` §8, dass fremde
Daten der Normalfall sind und so weit gelesen wird, wie es verantwortbar ist —
eine abgewiesene Datei ist ein unlesbares Paket, und das ist der teurere
Fehler. Zweitens ging nichts verloren, was jemand mitbekam: der Fehler traf
zwölf gesunde Dateien und keine kaputte.

## Verworfen: eine Liste von Skriptdateinamen

`script`, `unit_tests`, `init_script` — das sähe nach einer Regel aus und wäre
eine Aufzählung. `new-blends`, `masterpiece_script` und `style css` zeigen, dass
die Namen frei sind. Eine solche Liste wäre beim nächsten Pack still zu kurz,
und **still zu kurz** ist die schlechteste Eigenschaft, die ein Wächter haben
kann.

## Verworfen: am Inhalt erkennen, ob eine Datei ein Baum ist

„Mehr als die Hälfte der Wurzelzeilen trägt einen Doppelpunkt, also ist es ein
Baum." Das ist eine Schwelle, und eine Schwelle ist eine Vermutung mit einer
Zahl davor. `magic.mse-game/language` fiele darunter oder darüber, je nachdem,
wo man sie legt — und niemand könnte sagen, warum die eine Zahl richtiger ist
als die andere.

## Verworfen: nur die Hauptdatei als Baum lesen

Das ist die **wahre** Regel, und sie ist trotzdem nicht die brauchbare.
`include file:` liest Nebendateien, und der Einbinder braucht ihre Knoten, um
sie einzusetzen. Wer Nebendateien nur als Text führte, müsste den Einbinder
zweigleisig machen — für einen Nutzen, den `Roh` schon liefert.

Die Regel bleibt trotzdem wahr, und sie steht in `docs/FORMATE.md`: **ein Baum
ist eine Datei erst, wenn jemand sie als einen liest.**

## Wer das ändern will

ändert zuerst diesen Vermerk. Der wahrscheinlichste Anlass wäre ein Fall, in
dem eine kaputte Hauptdatei still gelesen wird und jemand lange sucht — dann
gehört nicht die Strenge zurück, sondern ein **Befund** an der Stelle: der
Zerteiler kennt heute keinen Weg, einen zu melden, ohne abzubrechen.
