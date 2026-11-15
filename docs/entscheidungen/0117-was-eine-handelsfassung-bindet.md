# 0117 — Was eine Handelsfassung bindet, und was ein neues Feld kostet

**Stand:** angenommen · **Datum:** 2026-09-18
**Berührt:** [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md),
[0075](0075-der-satz-und-sein-grund.md)

## Warum es diesen Vermerk gibt

Weil Werkstück `T1` von Phase 5.6
(#1466)
verlangt, „kompatible additive Erweiterungen und Migrationen … festzulegen;
offene Fragen nicht an nachgelagerte Agenten delegieren". Die Frage dahinter
ist eine einzige, und sie hat keine bequeme Antwort:

> **Was geschieht mit einer Unterschrift, wenn das Angebot ein Feld mehr
> bekommt?**

## Die Antwort: ein neues Feld ist eine neue Fassung

`domain.Handelsangebot.Unterschriftsform` fügt **jedes fachliche Feld** in
fester Reihenfolge zusammen, mit dem Nullbyte getrennt und dem Domänenwort
`handelsangebot` zuerst. Wer ein Feld hinzufügt, ändert damit die Form —
**und jeden Abdruck jeder Revision, die je geschrieben wurde.**

**Das ist kein Nachteil, den man abmildert, sondern der Sinn der Sache.** Eine
Unterschrift, die ein Feld nicht bindet, sagt über dieses Feld nichts; eine,
die sich beim Hinzufügen nicht ändert, hätte dasselbe Problem eine Ebene
tiefer.

Deshalb gilt:

1. **`Handelsvertragsfassung` steigt, sobald sich die Form ändert.** Ein Feld
   mehr, ein Feld weniger, eine andere Reihenfolge — jedes davon.
2. **Die Form einer erschienenen Fassung bleibt, wie sie ist.** Sie steht
   unter einem Goldlauf (`domain/handelsgold_test.go`); wer sie anfasst,
   bekommt einen roten Lauf und keine stille Änderung.
3. **Eine unbekannte Fassung wird abgewiesen**, nicht so gut wie möglich
   gedeutet. Wer eine Fassung nicht kennt, weiß nicht, was er nicht weiß —
   und unterschriebe womöglich etwas anderes, als er gelesen hat.
4. **Alte Revisionen behalten ihre Fassung.** Sie steht in ihrer Zeile
   (`handelsrevisionen.fassung`), und ein Vorgang, der unter Fassung 1
   begonnen hat, läuft unter Fassung 1 zu Ende — dieselbe Regel wie beim
   Regelprofil einer Partie.

## Was das kostet, und es gehört ausgesprochen

**Ein Feld hinzuzufügen ist teuer.** Es ist keine „additive Erweiterung" im
Sinn von „die alte Seite merkt nichts": die alte Seite **merkt es**, weil sie
die Fassung nicht kennt, und weist ab. Zwei Rechner mit verschiedenen Fassungen
handeln nicht miteinander, bis der ältere aktualisiert ist.

**Die bequeme Alternative ist geprüft und verworfen**: eine Form, die
unbekannte Felder überspringt, ließe zwei Seiten über **verschiedene** Inhalte
unterschreiben und beide glauben, es sei dasselbe. Genau das ist der Schaden,
gegen den eine Unterschrift steht.

## Und was daraus nicht folgt

**Nicht jede Änderung am Verbund ist eine Fassungsänderung.** `Zustand` und
`Zustimmungen` gehen **nicht** in die Form ein — eine Aussage über eine
Revision kann nicht Teil der Revision sein, der sie gilt. Wer dort ein Feld
hinzufügt, ändert keine Unterschrift.

**Und die Bewertung ist der benannte Sonderfall**: sie bindet nur, wenn
`BewertungGeteilt` es sagt. Der Schalter selbst bindet immer — sonst hätten
„geteilt, aber null" und „still" dieselbe Unterschrift, und sie sagte nicht,
worüber sie geschlossen wurde.
