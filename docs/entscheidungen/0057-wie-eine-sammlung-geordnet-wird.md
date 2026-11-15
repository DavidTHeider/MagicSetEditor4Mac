# 0057 — Wie eine Sammlung geordnet wird

**Stand:** entschieden · **Datum:** 2026-08-30 ·
**Betrifft:** #1292 ·
**Ergänzt** [0052](0052-die-sammlung-ist-nicht-das-set.md)

## Die Frage

Phase 4.5 verlangt **Ordnung**: „Lagerort, Ordner, Box, Wunschliste,
Dubletten — **neue Entscheidung**, in einem Vermerk." Das Issue sagt
ausdrücklich, dass es sich hier nicht um eine Portierung handelt.

Zu entscheiden ist, **welche dieser fünf Dinge Daten sind** und welche eine
Frage an vorhandene Daten.

## Erst gesucht, dann entschieden

**Es gibt dafür keine Vorlage in MSE2**, und §15 verlangt, dass gesagt wird,
wo gesucht wurde. Gesucht am 2026-08-30 in der Fassung `f9d9356` nach
`storage`, `location`, `binder`, `box`, `wishlist`, `duplicate` und `shelf`:
die Treffer sind Kommentare über Zeigerbesitz und ein `Storage`-Typ des
Skriptsystems — nichts zur Sache. `magic-set-generator` (`1b8fb7f`) kennt
weder Sammlung noch Ordnung.

**Und keine bei Scryfall.** Wo etwas liegt, weiß kein Kartendienst; das ist
eine Aussage über die Wohnung der nutzenden Person.

## Die Entscheidung

**Zwei Spalten, keine dritte Tabelle, und zwei Fragen ohne Spalte.**

| Was | Wie | Warum |
| --- | --- | --- |
| **Lagerort** | eine Spalte `lagerort` als **Freitext** | „Kiste 4", „Ordner Rot, Seite 12", „bei Anna" — jede Aufzählung, die wir vorgeben, ist bei jemandem falsch |
| **Ordner, Box** | **keine eigene Spalte** | sie sind Lagerorte. Drei Felder für dieselbe Sache zwingen zu einer Entscheidung, die niemand treffen will („ist das ein Ordner oder eine Box?") |
| **Wunschliste** | eine Spalte `wunsch` — 0 oder 1 | siehe unten |
| **Dubletten** | **keine Spalte**, eine Abfrage | eine Dublette ist kein Zustand, sondern ein Verhältnis: *mehr als ich behalten will* |

### Warum der Lagerort Freitext ist und keine Aufzählung

Weil eine Sammlung physisch ist und jede physische Ordnung anders. Ein
`enum` aus `ordner | box | regal` müsste raten, was Menschen tun, und läge bei
jedem Zweiten daneben. **Freitext kostet die Sortierbarkeit nicht**: gleiche
Zeichenketten gruppieren sich von selbst, und die Sicht bietet an, was
vorkommt.

**Was das nicht kann**: eine Hierarchie („Ordner Rot → Seite 12"). Wer sie
will, schreibt sie in den Text; wer sie als Baum braucht, braucht eine zweite
Tabelle — und das ist eine Erweiterung, keine Berichtigung.

### Warum die Wunschliste eine Spalte ist und keine zweite Tabelle

**Eine Wunschzeile ist derselbe Verbund wie eine Besitzzeile**: ein Druck, eine
Menge, ein Zustand, eine Sprache, eine Ausführung, eine Notiz, ein Preis, den
man zu zahlen bereit ist. Eine zweite Tabelle wäre eine Kopie derselben zwölf
Spalten — mit demselben Leser, demselben Schreiber, derselben Ausfuhr und
demselben Rundlauf, doppelt.

**Was sie bedeutet, steht dagegen fest**: eine Zeile mit `wunsch = 1` ist
**kein Besitz**. Daraus folgen drei Dinge, und sie sind nicht verhandelbar:

1. **Das Portfolio zählt sie nicht.** Was man nicht hat, ist nichts wert.
2. **„Besitze ich diese Karte?" sagt Nein.** Ein Wunsch ist die Antwort auf
   dieselbe Frage — mit dem anderen Vorzeichen.
3. **Die Eindeutigkeit nimmt sie auf.** Der Schlüssel lautet ab jetzt
   `(printing, zustand, sprache, foil, wunsch)`. Ohne das könnte man denselben
   Druck nicht zugleich besitzen und sich wünschen — und genau das tut, wer
   eine gespielte Fassung hat und eine schöne will.

**Der Preis dieser Entscheidung, ausgesprochen**: jede Abfrage über die
Sammlung muss sich entscheiden, ob sie Wünsche meint. Wer das vergisst, zählt
sie mit. Dagegen steht ein Wächter — `TestJedeSammlungsabfrageEntscheidetUeber
Wuensche` — und nicht die Aufmerksamkeit des Nächsten.

### Warum Dubletten keine Spalte sind

Weil „Dublette" keine Eigenschaft einer Karte ist. Es ist die Frage: *wovon
habe ich mehr, als ich behalten will?* Die Antwort hängt an einer Zahl, die
die fragende Person mitbringt — meist 1, bei Ländern 4, bei einem Deckbau
anders.

**Deshalb ist es eine Abfrage mit einem Parameter** (`behalten`), keine Spalte
mit einem Wert. Eine Spalte müsste bei jeder Mengenänderung nachgeführt
werden, und sie wäre falsch, sobald jemand die Zahl anders meint.

## Was verworfen wurde

| Verworfen | Warum |
| --- | --- |
| **Eine Tabelle `lagerorte` mit Fremdschlüssel** | eine Normalisierung, die nichts trägt: ein Lagerort hat außer seinem Namen keine Eigenschaft. Sie kostet einen Verbund bei jeder Abfrage und eine Pflegefläche, die niemand will |
| **Eine eigene Tabelle `wunschliste`** | zwölf Spalten doppelt, samt Leser, Schreiber, Ausfuhr und Rundlauf. Der einzige Unterschied wäre ein Wort |
| **`menge = 0` als Wunsch** | die Domäne sagt seit 0052: „Null ist kein Exemplar, sondern seine Abwesenheit." Zwei Bedeutungen für dieselbe Null wären eine Falle |
| **Eine berechnete Spalte `dublette`** | siehe oben: die Zahl kommt von der fragenden Person |
| **Ordner und Box als eigene Felder** | drei Felder für dieselbe Sache. Wer „Ordner Rot" tippt, hat einen Lagerort genannt |

## Wie es rückgängig zu machen wäre

Beide Spalten haben eine Vorgabe und sind damit **verträglich** im Sinne von
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md): ein älteres Programm
liest den Bestand weiter. Wer die Wunschliste verwerfen will, löscht die Zeilen
mit `wunsch = 1` und die Spalte; die Eindeutigkeit fällt auf ihre vier Spalten
zurück.
