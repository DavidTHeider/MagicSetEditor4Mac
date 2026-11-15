# 0036 — `card_regions` ist außer Reichweite, und MSE2 bleibt bei f9d9356

**Stand:** angenommen
**Datum:** 2026-08-20
**Berührt:** [`internal/service/exportfremd_test.go`](../../internal/service/exportfremd_test.go),
#659,
#609

## Lage

Zwei Exportvorlagen des vollen Packs lesen ein Glied, das der gepinnte
MSE2-Quelltext nicht kennt:

```
sts := card.stylesheet
if sts == nil then sts := set.stylesheet
if length(sts.card_regions) > 0
```

`magic-lackeyccg.mse-export-template`, Zeilen 745 und 758. Dazu
`magic-lackeybot.mse-export-template/script`, `magic-cockatrice-v2` und
`magic.mse-game/script` 491–493.

**`card_regions` kommt in `f9d9356` nirgends vor.** Nachgesehen mit
`grep -rn card_regions /Users/david/mse2-quelle/src` — kein Treffer, auch nicht
unter `regions` in `stylesheet.hpp` oder `stylesheet.cpp`. Fünf `.mse-locale`
tragen eine Beschriftung `card regions:`; das Pack zielt auf eine **neuere**
MSE-Fassung als die hier gepinnte.

## Entscheidung

**`Lackey` bleibt außer Reichweite.** Die MSE2-Fassung wird **nicht**
nachgezogen.

**Und `Cockatrice` steht ausdrücklich nicht auf dieser Liste.** Es nennt
`card_regions` ebenfalls, läuft aber durch: der Zweig wird mit den
Vorgabewerten seiner Optionen nicht betreten. Ein Name in einer Datei ist
kein Fehlschlag — gezählt wird, was ein Lauf antrifft (§10).

## Warum

**Ein Fassungswechsel berührt jede Herkunftsangabe im Baum.** §11 nennt
`f9d9356` als *die* Fassung, und
[0028](0028-die-fassung-in-einer-herkunftsangabe.md) macht einen Verweis ohne
sie rot. Nachzuziehen hieße: jede portierte Regel gegen den neuen Stand
nachlesen, sonst steht neben ihr eine Fassung, gegen die sie nie geprüft
wurde. Das ist die teuerste Art, eine Zusage zu verlieren — sie sieht danach
aus, als gälte sie.

**Der Nutzen ist eine Vorlage.** Über `M15-Magic-Pack` gemessen laufen 14 von
19 durch, eine schweigt zulässig, drei scheitern an etwas anderem — und
`Lackey` ist die eine, die an dieser Fassung nicht laufen **kann**. Und
`card.stylesheet` allein zu
portieren brächte `Lackey` **einen Schritt weiter und nicht durch**: der
nächste Ausdruck ist `sts.card_regions`, und dafür gäbe es weiterhin keine
Referenz. §15 verbietet, das Verhalten zu erfinden.

## Was sie kostet

**Eine Vorlage läuft nicht**, und niemand kann sie zum Laufen bringen, ohne
diese Entscheidung umzudrehen. Der Messlauf nennt sie namentlich und mit
eigenem Zeichen (`⊘`), damit sie nicht wie ein unerklärter Rest aussieht.

**Und `Cockatrice` kann jederzeit dazukommen.** Es liest dasselbe Glied in
einem Zweig, den seine Vorgabewerte nicht betreten; wer die Optionen anders
stellt, trifft es. Dann gilt dieser Vermerk auch dort.

**Und die Grenze wandert.** Je neuer ein bezogenes Pack, desto mehr Vorlagen
können auf Sprachmittel zeigen, die es 2021 nicht gab. Das ist kein Mangel
dieses Vermerks, sondern seine Voraussetzung: er entscheidet über **diesen**
Fall und nennt die Regel, nach der der nächste zu entscheiden ist — nachlesen,
ob es die Stelle in `f9d9356` gibt, und wenn nicht, den Grund aufschreiben
statt das Verhalten zu erraten.

## Wer sie umdreht

Wer sie umdreht, zieht **zuerst** die Fassung in §11 und
[0028](0028-die-fassung-in-einer-herkunftsangabe.md) nach, dann jede
Herkunftsangabe — und erst danach `card.stylesheet`.
