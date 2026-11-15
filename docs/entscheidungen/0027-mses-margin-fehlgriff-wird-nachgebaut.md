# 0027 — MSEs `margin`-Fehlgriff wird nachgebaut, nicht behoben

**Stand:** gültig
**Datum:** 2026-08-18
**Betrifft:** #433

## Die Lage

`<margin:links:rechts:oben>` setzt bei MSE die Ränder des laufenden Absatzes.
Die drei Zahlen werden so abgegriffen:

```cpp
text.substr(colon  + 1, colon2 - colon  - 2).ToDouble(&m.left);
text.substr(colon2 + 1, colon3 - colon2 - 2).ToDouble(&m.right);
text.substr(colon3 + 1, pos    - colon3 - 2).ToDouble(&m.top);
```

`src/render/text/element.cpp`, Zeilen 181–183, Fassung f9d9356 vom 2021-07-06.

**Die Länge zieht 2 ab, wo für ein durch `:` beendetes Feld 1 richtig wäre.**
Für `<margin:130:0:0>` durchgezählt — `colon` 7, `colon2` 11, `colon3` 13,
`pos` 16:

| Feld | Ausdruck | Ergebnis |
| --- | --- | --- |
| links | `substr(8, 2)` | `"13"` → **13**, nicht 130 |
| rechts | `substr(12, 0)` | `""` → `ToDouble` scheitert, bleibt **0** |
| oben | `substr(14, 1)` | `"0"` → **0** |

Das dritte Feld stimmt: es wird von `>` beendet, und `pos` steht schon dahinter.
Die ersten beiden verlieren ihr letztes Zeichen — aus `20` wird 2, aus `130`
wird 13, aus `9` wird nichts.

## Die Entscheidung

**Nachgebaut, nicht behoben.**

Das Ziel dieses Projekts ist **Parität**: eine Karte aus diesem Programm soll
von einer aus MSE2 nicht zu unterscheiden sein. Ein „richtig" gelesener Rand
rückte den Absatz um das **Zehnfache** ein und stünde damit weiter von MSEs
Bild entfernt als das bisherige Wegwerfen der Auszeichnung.

§15.4 verlangt, dass eine Änderung bestehender Semantik ausgesprochen und
begründet wird. Hier wird sie nicht geändert — und *dass* sie nicht geändert
wird, ist die Aussage, die aufgeschrieben gehört: ohne diesen Vermerk liest der
nächste Leser `gekuerzteZahl` als Fehler dieses Projekts und behebt ihn.

**Nachgestellt wird die Rechnung, nicht ihr Ergebnis.** `service.gekuerzteZahl`
schneidet das letzte Zeichen ab, weil MSE die Länge falsch rechnet. Ein
`/ 10` an derselben Stelle träfe für `130` denselben Wert und für `9` nicht —
und verlöre den Grund.

## Was das kostet

**Wer die Vorlagen schreibt, schreibt gegen den Fehler an.** `<margin:130:0:0>`
im Pack meint sichtbar 13; die Zehn davor ist Kompensation, nicht Absicht. Ein
späteres MSE, das den Fehler behöbe, veränderte damit jede Vorlage, die sich
darauf eingestellt hat — und dieses Projekt müsste dann **beide** Lesarten
kennen, unterschieden an der Fassung.

## Am Bild bestätigt

**Die Prüfung ist gelaufen.** `scripts/mse2-referenz.sh Green` hat alle 3.689
Karten gerastert; gemessen wurde an `A-Tyvar Kell`, der linke Rand der
Fähigkeitszeilen als linkeste dunkle Spalte je Zeilenband:

| Zeilen | MSE2 | **verkürzt** (→13) | ungekürzt (→130) |
| --- | ---: | ---: | ---: |
| 336–352 | 62 | **67** | **184** |
| 376–392 | 62 | 68 | 62 |
| 414–430 | 67 | 62 | 62 |

**122 Punkte Unterschied** in der Zeile, die die Auszeichnung trägt — genau die
Einrückung, die `130` gegenüber `13` erzeugt. Die verkürzte Lesart liegt auf
fünf Punkte an MSE2, die ungekürzte um mehr als das Doppelte der Kastenbreite
daneben.

Die Abstände sagen dasselbe: ganze Karte 49,740 gegen 50,092, Textkasten
48,594 gegen 49,326.

**Der erste Anlauf trug noch nicht.** Er ergab 1,3 Einheiten Unterschied, weil
in derselben Zone die Treuemarken falsch standen
(#449,
#451). Eine
Messung, die von einem größeren Fehler beherrscht wird, ist keine — sie wurde
nach dessen Behebung wiederholt.
