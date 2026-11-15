# Kompressionsbomben

**Selbst gebaut, nicht heruntergeladen** (§11). Was hier liegt, ist ein Archiv,
das sich beim Auspacken vervielfacht — die Gestalt, gegen die
`internal/bezug` seine Mengengrenzen hat
(#1163,
Befund Z7-01 des Audits für 0.3.0).

| Datei | Verdichtet | Ausgepackt | Verhältnis |
| --- | --- | --- | --- |
| `verdichtungsbombe.tar.gz` | 65.349 Byte | 67.108.864 Byte | **1026 : 1** |

**Nullbytes und sonst nichts.** Der Inhalt ist gleichgültig; was zählt, ist das
Verhältnis. Ein Archiv aus Nullbytes ist die einfachste Bombe, die es gibt, und
sie ist mit zehn Zeilen nachzubauen — das Skript dafür steht im Kommentar von
`internal/bezug/verdichtungsbombe_intern_test.go`.

**Sie ist absichtlich klein.** Eine, die die **Produktionsgrenzen** von zwei
und vier Gigabyte reißt, wöge hier vier Megabyte. Der Lauf dreht die Grenzen
stattdessen herunter und fährt denselben Weg — und was das belegt und was
nicht, steht dort.
