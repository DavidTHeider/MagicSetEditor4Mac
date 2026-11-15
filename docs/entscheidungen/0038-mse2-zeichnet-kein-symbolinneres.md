# 0038 — MSE2 zeichnet auf diesem Rechner kein Symbolinneres

**Stand:** angenommen
**Datum:** 2026-08-21
**Berührt:** [`internal/app/kanonsymbolfeld_intern_test.go`](../../internal/app/kanonsymbolfeld_intern_test.go),
[`internal/app/mse2kanon_intern_test.go`](../../internal/app/mse2kanon_intern_test.go),
[0026](0026-mse2-rendert-die-referenz-selbst.md),
#744,
#742

## Lage

Seit [0026](0026-mse2-rendert-die-referenz-selbst.md) rastert MSE2 die
Referenzbilder selbst — auf diesem Rechner, aus dem Quelltext gebaut. Das war
der Schritt von einem Verlaufsmaß zu einer Aussage über Parität.

**In einem Feld gilt sie nicht.** Jedes Setsymbol in `mse2-ausgabe/` steht
durchweg in seiner **Randfarbe**; die Füllung der Variante kommt nirgends vor.

## Der Befund ist von #458, nicht von diesem Vermerk

Das gehört an den Anfang und nicht in eine Fußnote. **Gemessen wurde es am
2026-08-19**, über die **3.761 Karten von `White`** und mit dem Rückgabewert des
Blits selbst:

| Rechenart | `Blit` gibt | ausgeführt |
| --- | --- | --- |
| `wxCOPY`, `wxOR`, `wxXOR` | true | ja |
| **`wxAND`**, **`wxAND_INVERT`** | **false** | **nein** |

Gemessen gegen **wxWidgets 3.3.3 / wxOSX-Cocoa**, den Bau aus
[0026](0026-mse2-rendert-die-referenz-selbst.md).

Am 2026-08-21 ist derselbe Befund beim Nachgehen von
#742 **ein
zweites Mal hergeleitet** worden — über einen Bisect an einem auf zwei Karten
gekürzten `White.mse-set`, sechs Seltenheiten, acht Karten. Er stimmt mit #458
überein und fügt ihm nichts hinzu.

**Das war vermeidbar**, und der Satz gehört hierher: #458 stand geschlossen da,
und in den geschlossenen Issues wurde nicht nachgesehen. §10 sagt es für den
umgekehrten Fall — „die Erinnerung eines Werkzeugs ist kein Prüfergebnis" —,
und er gilt in beide Richtungen: **wer eine Messung neu erhebt, sieht vorher
nach, ob sie schon dasteht.**

Was dieser Vermerk beiträgt, ist deshalb nicht der Befund, sondern die
**Entscheidung** darüber — #458 nannte zwei Wege und wählte keinen.

## Wie es 2026-08-21 noch einmal gemessen wurde

| Seltenheit | Füllung laut `set_fields` | Rand laut `set_fields` | MSE2 zeichnet |
| --- | --- | --- | --- |
| `common` | schwarz | weiß | **weiß** |
| `basic land` | schwarz | weiß | **weiß** |
| `uncommon` | Grauverlauf | schwarz | **schwarz** |
| `rare` | Goldverlauf | schwarz | **schwarz** |
| `mythic rare` | Orangeverlauf | schwarz | **schwarz** |
| `special` | Lilaverlauf | weiß | **weiß** |

Sechsmal die Randfarbe, nullmal die Füllung.

**Vorher ausgeschlossen**, damit die Erklärung nicht die erstbeste ist:

- Die Variante wird gefunden. Am Orakel gibt
  `symbol_variation(symbol: set.symbol, variation: "rare")` einen Bildwert;
  erst ein erfundener Name wirft `Variation of symbol not found`. Die
  Variantenliste des Stylesheets **ergänzt** die des Spiels, sie ersetzt sie
  nicht.
- `use_main_rarity()` ist falsch — `set.mainframe_rarity_name` ist leer.
- `use_alt_rarity()` ist falsch — `alt_rarity_color` steht nirgends im Set.
- Die abgelegten Bilder sind reproduzierbar: ein frischer Export von
  `Archenemy.mse-set` ist **70 von 70 bitgleich**.

## Wo es herkommt

```cpp
void combineBuffers(DC& dc, DC* borders, DC* interior) {
  if (borders)  dc.Blit(0,0,w,h, borders,  0,0, wxOR);
  if (interior) dc.Blit(0,0,w,h, interior, 0,0, wxAND_INVERT);
}
```

`SymbolViewer::draw` und `combineBuffers`, `src/render/symbol/viewer.cpp`,
Zeilen 85–118, Fassung f9d9356 vom 2021-07-06.

Der Randpuffer trägt das **ganze** Vieleck: `border->SetBrush(…)` steht vor
`DrawPolygon`, nicht nur ein Strich (Zeilen 265–270). Erst der zweite Blit nimmt
das Innere wieder heraus. **Bleibt der zweite ohne Wirkung**, steht das ganze
Symbol in der Randfarbe und außen bleibt es durchsichtig — genau das ist zu
sehen, sechsmal von sechs.

**Dass er ohne Wirkung bleibt, ist gemessen** — der Blit gibt `false` zurück
und führt nichts aus (Tabelle oben, #458).

**Warum wxOSX-Cocoa ihn abweist, ist es nicht** (uncertain, §15). wxWidgets
liegt auf diesem Rechner nur als Homebrew-Binärdatei vor. Der Zweitblick durch
Codex hat am 2026-08-21 drei Lesarten nebeneinandergestellt, die alle zu den
Beobachtungen passen — der Cocoa-`wxMemoryDC`-Pfad kennt die Rechenart nicht,
`Blit` scheitert an der Tiefenkombination mit dem ausdrücklichen 24-Bit-Puffer,
oder die Umwandlung zwischen Bitmap- und DC-Modus behandelt sie anders als den
erwarteten monochromen Maskenweg. **Keine davon ist belegt, und keine ändert
etwas an der Wirkung.** Gebaut ist gegen wx **3.3.3** statt der in MSE2s
Anleitung geprüften 3.0.5 — ein Unterschied, den 0026 als offenen Punkt führt.

## Entscheidung

**Der Maßstab wird nicht angepasst, und unsere Zeichnung auch nicht.**

1. **Das Symbolfeld ist als unerreichbar gekennzeichnet**, nicht ausgeschnitten.
   Ein Wächter hält den Befund fest — `TestMSE2ZeichnetKeinSymbolinneres` prüft,
   dass MSE2s Symbol **einfarbig** ist. Ein richtig gezeichnetes hat zwei
   Farben; wird MSE2 neu gebaut und zeichnet wieder ein Inneres, wird der Lauf
   rot und dieser Vermerk fällig.
2. **Herausgerechnet wird nichts.** Der Kanonabstand bliebe sonst eine Zahl mit
   einer Ausnahme, die niemand nachvollzieht — und die Lage kann sich mit dem
   nächsten wx-Bau ändern.
3. **Was es kostet, steht bei der Zahl.** Über acht Karten aus `White`, mit der
   Metrik dieses Projekts:

   | | Abstand |
   | --- | ---: |
   | ganze Karte | 22,202 |
   | ohne das Symbolfeld | 21,423 |
   | im Symbolfeld allein | 158,557 |

   Rund **0,78** sind dort nicht einzuholen.

## Folgen

**#742 ist
damit erklärt.** Die Symbolauflösung wurde am 2026-08-21 auf MSEs Regel
umgestellt, und der Kanonabstand stieg um 0,032 — eine schärfere Zeichnung einer
Füllung, die im Vergleichsbild gar nicht vorkommt, liegt weiter von der reinen
Randfarbe weg. Außerhalb des Feldes steht die Zahl bei beiden Fassungen auf
**derselben** 21,423.

**Und ein Satz, der allgemeiner gilt als dieser Fall**: der Vergleichsmaßstab
ist selbst ein Programm auf einem Rechner. Wo unsere Zahl nicht besser wird,
ist die erste Frage nicht mehr nur „was machen wir falsch", sondern auch „misst
der Maßstab hier überhaupt". Die Antwort gehört belegt, nicht vermutet — hier
hat sie einen Bisect über sechs Varianten gekostet.

## Was offen bleibt

**Unsere Symbolzeichnung ist gegen kein fremdes Bild belegt.** Sie ist gegen den
C++-Quelltext portiert und durch Einheitstests abgedeckt
(`internal/render/symbolfilter_test.go`), und #458 hat ihre Rechnung an einem
Punkt nachgerechnet — der gemessene Wert (172,172,172) ergibt mit
`t = |x+y−1| = 0,41` und `lerp(224, 84, 0,41) = 167` Silber, wie ein Uncommon
sein soll. **Das ist eine Stichprobe von einem Punkt und kein Bildvergleich.** Ein Weg dorthin wäre MSE2 auf einer Plattform,
auf der der zweite Blit trägt — er ist nicht gegangen und steht als
#744.
