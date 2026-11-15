# 0127 — Ein Betrag steht in der kleinsten Einheit seiner Währung

**Stand:** entschieden am 2026-09-22 · LOC-03.AK-02
(#1689)

## Was galt

**„Hundertstel der Währung".** So stand es an `domain.Exemplar.Erwerbspreis`,
an `domain.Preisstand.Wert` und in `domain.PreisAusText`: 4,17 € sind 417.

Für den Euro und den Dollar ist das richtig. Für zwei Gruppen ist es falsch,
und beides ist **gemessen** (2026-09-22, `Intl.NumberFormat` und
`strconv.ParseFloat`):

| Fall | vorher | was daran falsch ist |
| --- | --- | --- |
| 15.050 „Hundertstel JPY" | »151 ¥« | ¥150,50 gibt es nicht — gerundet wurde stillschweigend |
| 1 „Hundertstel BHD" | »0,010 BHD« | ein Fils lässt sich gar nicht ausdrücken |
| »1,23« aus einer deutschen Liste | `0` | `ParseFloat` scheitert, und der Aufrufer machte eine Null daraus |

Der dritte Fall ist der teuerste: wer eine deutsch formatierte Sammlung
einlas, verlor **jeden** Preis, ohne ein Wort zu lesen.

## Was gilt

**Ein Betrag ist eine ganze Zahl in der kleinsten Einheit seiner Währung** —
Cent beim Euro, **Yen beim Yen**, Fils beim Dinar. Wie viele Stellen das
sind, sagt die Währung und nicht der Aufrufer: `domain.Waehrungsstellen`
führt die null- und die dreistelligen Währungen nach ISO 4217, alles andere
hat zwei.

**Die kanonische Form hat einen Punkt und kein Komma.** Sie ist keine
Sprache: in ihr wird gespeichert und über die Flächen geredet. Wie eine Zahl
**aussieht**, entscheidet die Sprache des Fensters — `Intl` im Frontend, mit
den Stellen derselben Währung.

**Was nicht kanonisch ist, wird erklärt** und nicht still zu null gemacht.
Beim Einlesen einer Sammlung bleibt die Zeile trotzdem: der alte Grund
(„eine Sammlung scheitert nicht an einer Spalte, die niemand braucht") gilt
weiter — was nicht mehr gilt, ist das Schweigen. Preis **und Währung** der
Zeile bleiben dann leer, denn eine Währung ohne Betrag sagt nichts.

**Und dasselbe gilt für einen Preis, der sich liest, aber nicht gilt.** Hier
stand der Satz nur über die Form, und der Code hielt ihn nur dort: »-1.00«,
»1.00« ohne Währung, »EURO« oder eine Währung ohne Betrag kamen durch
`GeldAus` — und die Prüfung des Exemplars verwarf danach die **ganze** Zeile.
Befund G16 des Zweitblicks über Phase 6.3
(#1666).
Seither leert das Einlesen das Paar, wenn `Exemplar.Preisbefunde` es
abweist, und nennt den Grund; **den Tag des Erwerbs ebenso**
(`Exemplar.Datumsbefunde`), denn auch er beschreibt den Kauf und nicht die
Karte (`service.erwerbGeprueft`).

## Was das kostet

**Eine Migration** (`0078_geldskala.sql`): vorhandene Zeilen in
null- und dreistelligen Währungen werden umgerechnet, alle übrigen bleiben
unberührt. Gerundet wird zur nächsten ganzen Einheit — ein halber Yen war
schon vorher keine Angabe.

## Und seit dem 2026-09-22 gilt er auch für die Marktpreise

**Hier stand eine Grenze**: „der Erwerbspreis folgt dem Vertrag, der
Marktpreis noch nicht" — `PreisAusText` las jeden Preis mit zwei
Nachkommastellen und kannte die Währung nicht. Sie ist am selben Tag
gefallen: `MarktpreisAus` nimmt die Währung entgegen und rechnet über
`GeldAus`; die drei Leser bei Scryfall, Scrydex und MTGJSON geben sie mit,
und `0079_marktpreisskala.sql` rechnet den Bestand um.

**Zwei Dinge bleiben dabei anders als beim Erwerbspreis**, und beide stehen
an `MarktpreisAus`:

  * **Ein leerer Text ist hier ein Fehler.** In einer Sammlung heißt „kein
    Preis" *kein Preis erfasst*; bei einer Quelle heißt es, sie hat einen
    angekündigt und nichts geschickt — und eine Null daraus sähe aus wie
    „geschenkt".
  * **Ein negativer Marktpreis ist keiner** — was eine Quelle darunter
    verstünde, weiß niemand. **Ein Erwerbspreis ist es auch nicht**
    (`Exemplar.Preisbefunde`); hier stand bis zum 2026-09-26, er dürfe
    negativ sein, und der Code sagte seit jeher das Gegenteil (Zweitblick
    über Phase 6.3, G19).

**Was damit vollständig ist**: LOC-03.AK-02.

## Und die Skala hat eine Quelle, nicht zwei

**Hier stand, das Fenster nehme die Stellen „aus `Intl` und nicht aus
einer zweiten Liste".** Das waren zwei Listen: Go rechnet nach ISO 4217,
`Intl` nach CLDR — und die beiden sind nicht dieselbe. Gemessen am
2026-09-22 über alle 162 Codes, die `Intl` kennt: **17 Abweichungen**, und
in jeder nennt `Intl` weniger Stellen als ISO.

| Code | Go (ISO 4217) | `Intl` (CLDR) | was im Fenster stand |
| --- | --- | --- | --- |
| HUF | 2 | 0 | 1.999 Fillér als »1.999 Ft« — hundertfach |
| IQD | 3 | 0 | 1.234 Fils als »1.234 IQD« — tausendfach |
| AFN, ALL, COP, IDR, IRR, KPW, LAK, LBP, MGA, MMK, PKR, SLL, SOS, SYP, YER | 2 | 0 | hundertfach |

**Die Skala gehört dem, der speichert.** Das Fenster teilt jetzt durch
dieselbe Stellenzahl wie Go (`frontend/src/waehrungsstellen.ts`) und lässt
`Intl` nur noch **zeigen** — beim Forint ohne Nachkommastellen, wie es dort
üblich ist. Ein Wächter
(`internal/werkzeuge/waehrungsstellen_test.go`) fragt Go für jeden der
17.576 möglichen Codes und vergleicht; **was er nicht hält**, ist ISO 4217
selbst.

**Gefunden hat es der Rundlauf aus AK-01**, nicht ein Lauf über Geld: wer
eine Eingabe in die Stellen der Währung zurückliest, muss fragen, woher
diese Stellen kommen — und bekam zwei Antworten.

**Was offen bleibt**, und es gehört nicht mehr zu diesem Vermerk: AK-03s
„die volle Locale bleibt erhalten" — das Fenster kürzt `de-CH` noch auf
`de` — und der Rest von AK-04.
