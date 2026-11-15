# 0089 — Ein Befund ist kein Auftrag

**Stand:** angenommen · **Datum:** 2026-09-10 · **Werkstück:** O8 von Phase 5.2
(#1432)

## Der Anlass

`storage.Waisen` zählt seit #1220
drei Stellen im Bestand und sagt eine Zahl. Wer sie größer als null sieht, hat
einen Befund und **keinen Weg** — und der naheliegende Weg wäre der falsche:
löschen, bis die Zahl null ist.

**Das verdeckt echte Korruption.** Eine Zeile, die auf eine fehlende Karte
zeigt, ist ein Befund über den Bestand; sie wegzuräumen macht die Zahl hübsch
und den Schaden unsichtbar. Das Issue sagt es beim Namen: *echte Korruption
nicht durch kosmetisches Löschen verdecken.*

## Die Entscheidung

**Drei Klassen, und die Grenze zwischen ihnen ist die Arbeit.**

| Klasse | Heißt | Beispiel |
| --- | --- | --- |
| `sicher` | die Zeile zeigt auf nichts und hält nichts fest | ein Feldwert ohne seine Karte |
| `entscheidung` | es sieht nach einem Rest aus und **kann** ein Nachweis sein | ein abgelaufener Freigabenachweis |
| `nurInformieren` | auffällig, und es wird **nicht** angefasst | was nach Dubletten aussieht; eine offene Rückkehr |

**`nurInformieren` hat keine Löschabfrage** — nicht „noch keine", sondern
keine. Das ist die Zusage und keine Vergesslichkeit: wer sie später ergänzte,
müsste diesen Vermerk ändern.

### Gleiche Printingdaten sind keine Dubletten

Zwei Sammlungszeilen mit derselben Kartennummer sehen aus wie ein Fehler und
sind meistens keiner: **Zustand, Sprache, Ausführung, Erwerb und Lagerort**
unterscheiden sie, und die Sammlung führt genau diese Felder
([0057](0057-wie-eine-sammlung-geordnet-wird.md)). Was hier steht, ist deshalb
eine Auskunft und kein Auftrag — und die Sammlung wird von der Pflege
**überhaupt nicht** angefasst.

### Und eine offene Rückkehr wird nie weggeräumt

Sie ist der Nachweis darüber, dass etwas unklar ist
([0088](0088-die-epoche-einer-rueckkehr.md)). Ein Nachweis, der mit dem Alter
verschwindet, ist keiner. Das Issue verlangt es ausdrücklich für T2, T5 und
T6: Tombstones, offene Übergaben und bestätigte Reservierungen werden nicht
allein wegen Alter oder Ablauf entfernt. **Heute gibt es davon eine Sorte —
die Epoche —, und sie steht als `nurInformieren` da, damit sich die anderen
danebenstellen können.**

## Die Revision, und warum sie zweimal geprüft wird

**Zwischen „zeig mir, was los ist" und „mach es" liegt die Entscheidung eines
Menschen** — und in dieser Zeit kann sich der Bestand ändern. Eine Vorschau,
deren Ergebnis man aufhebt und später ausführt, beschreibt die Vergangenheit.

Die Revision ist die Prüfsumme **über die betroffenen Zeilen** und nicht über
ihre Zahl: zwei Zeilen weg und zwei dazu ergäben dieselbe Zahl und einen
anderen Bestand.

Geprüft wird sie an zwei Stellen, und das sind zwei verschiedene Fragen:

1. **Gegen die Vorschau, die ein Mensch gesehen hat** — „hat sich seit dem
   Anschauen etwas geändert?"
2. **In derselben Transaktion wie das Löschen** — „ändert sich gerade etwas?"

**Stimmt eine nicht, stoppt die betroffene Maßnahme — und nur sie.** Ein
Konflikt an einer Stelle nimmt den anderen nichts; eine gemeinsame Transaktion
hieße, dass eine geänderte Zeile fünf erledigte Maßnahmen zurückdreht.

## Die Reihenfolge ist die Zusage

Analysieren (rein lesend) → zeigen → abgleichen → **sichern** → ausführen →
**nachsehen**.

**Gesichert wird vor dem ersten Schreiben.** Nicht danach und nicht „falls
etwas schiefgeht": eine Sicherung, die erst entsteht, wenn schon gelöscht ist,
sichert den Schaden. **Ist nichts zu tun, wird nicht gesichert** — eine
Sicherung ohne Schreibvorgang wäre eine Datei, die niemand erklärt.

**Nachgesehen wird danach noch einmal**, und die Bilanz nennt **beide Stände**.
„Drei erledigt" ist keine Auskunft über den Bestand; „vorher sieben Befunde,
nachher vier" ist eine — und ohne den zweiten Stand wäre „erledigt" eine
Behauptung.

## Was das kostet

**Die Analyse geht über jedes Profil.** Sie zählt gerade die Zeilen, die zu
**keinem** mehr gehören, weil ihr Elternteil fehlt; eine `profil_id` in der
Klausel schlösse sie aus. Der Wächter über die Profilklausel führt sie deshalb
als Ausnahme — mit demselben Grund wie den Waisenzähler daneben.

**Und ausgeführt wird nicht über MCP.** `entscheidung` heißt „ein Mensch
entscheidet"; ein Werkzeug, das sie ausführen könnte, machte aus der Klasse
eine Beschriftung ([0044](0044-was-ein-werkzeug-schreiben-darf.md)). Dazu ist
die Vorschau die **Zustimmung**, und zustimmen kann nur jemand, der da ist.
Nachsehen darf ein Modell überall — die Analyse ändert nichts.

## Verworfen

**Alles wegräumen, was auf nichts zeigt.** Das ist der Weg, gegen den es diese
Entscheidung gibt.

**Eine Klasse „vermutlich sicher".** Vier Klassen sind eine mehr, als jemand
im Kopf behält — und die vierte wäre die, bei der niemand mehr weiß, ob sie
gefragt hat.

**Die Vorschau serverseitig merken.** Das Fenster könnte sie holen und beim
Ausführen wiederfinden — die Kommandozeile nicht: dort ist jeder Befehl ein
eigener Prozess. Was stattdessen reist, ist die **Revision**; sie ist die
Bindung, und sie funktioniert an jeder Fläche gleich.
