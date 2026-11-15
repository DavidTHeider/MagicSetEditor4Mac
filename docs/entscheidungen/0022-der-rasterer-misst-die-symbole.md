# 0022 — Der Rasterer misst die Symbole, der Plan nennt sie

**Stand:** 2026-08-17
**Betrifft:** #330

## Der Zwang

Ein Symbol ist bei MSE **nicht** quadratisch. Seine Größe ist die
**tatsächliche Bildgröße**, verhältnisgleich umgerechnet über
`font_size / image font size`, dazu je Symbol ein Abstand; innerhalb seines
Kastens wird das Bild ausgerichtet (`src/data/symbol_font.cpp`, Zeilen 50–52,
147–153, 279–294, Fassung f9d9356 vom 2021-07-06; ausführlich in
`docs/FORMATE.md`).

Bei `magic-mana-large` heißt das für ein Manasymbol der Größe 15: das Bild
`mana_r.png` misst 165 × 178 Bildpunkte, `image font size` ist 135, also
18,3 × 19,8 Punkte, plus `horizontal space: 2` ein Kasten von 20,3 × 19,8.
Dieses Programm rechnete mit 15 × 15.

**Die tatsächliche Bildgröße kennt nur, wer das Bild öffnet.** Und Pakete
öffnet nach §2 nicht der Planbauer: „der Planbauer öffnet keine Pakete — er
bekommt ein Modell und rechnet" (`feldangaben.go`). Damit steht die Rechnung
auf der falschen Seite der Naht.

## Die Entscheidung

**Der Plan nennt die Symbole und die Angaben ihrer Schrift. Gemessen und
angeordnet wird im Rasterer.**

Der Zeichenplan trägt je Symbolzug

- die Bilder samt ihrer eigenen `image font size` (`render.Symbol`),
- die Maße der Symbolschrift und des Feldstils (`render.Symbolmasse`):
  Schriftgröße, `image font size`, `horizontal space`, `vertical space` und die
  `alignment:` des Symbolschriftblocks.

Der Rasterer schlägt die Bildgröße bei seiner `Bildquelle` nach und rechnet
daraus den Kasten. Er tut das an **einer** Stelle
(`Symbolmasse.Kasten`) für beide Wege — die Manakosten wie die Symbole mitten
im Regeltext.

**Damit verschwindet die vorberechnete Symbolebene.** Die Manakosten waren bis
heute *n* Ebenen der Art `Symbol` mit ausgerechneten Rechtecken; sie sind jetzt
**eine** Ebene der Art `Symbolzug`, die der Rasterer setzt. Das ist zugleich
näher an MSE: dort gibt es diese Trennung gar nicht, ein Feld mit
`always symbol: true` wird vom **Textsatz** in `SymbolTextElement`e zerlegt
(`src/render/text/element.cpp`, Zeilen 244–262).

## Was verworfen wurde

**Ein Verhältnis im Plan statt einer Kantenlänge.** Der Planbauer schriebe
`Groesse / Bildgroesse` hin, der Rasterer multiplizierte die Bildgröße damit.
Das trüge die halbe Rechnung über die Naht und ließe die andere Hälfte —
Abstand, Ausrichtung im Kasten, Zeilenhöhe — dort, wo sie nicht hingehört. Zwei
Stellen für eine Regel laufen auseinander, sobald eine von ihnen etwas lernt.

**Den Planbauer Bilder öffnen lassen.** Das hieße, §2 aufzugeben: er wäre ein
zweiter Paketleser, und der Plan hinge an der Platte statt an einem Modell.

## Der Preis

**Der Plan ist nicht mehr allein aus sich heraus prüfbar.** Wo ein Rechteck
stand, das man nachrechnen konnte, steht jetzt eine Angabe, deren Wirkung erst
mit einer Bildquelle sichtbar wird. Ein Plantest über die Manakosten prüft ab
jetzt die **Absicht**, nicht das Maß; das Maß prüft ein Rastertest.

**`content_width` bleibt vorerst falsch.** `inhaltsmasse` in `service` rechnet
weiter mit dem Quadrat, denn sie darf keine Bilder öffnen. Für die Manakosten
heißt das: die Symbole haben ab jetzt das richtige Maß, der **Kasten**, in dem
sie sitzen, noch nicht. Das ist
#291 und
steht dort benannt, statt hier stillschweigend mitzulaufen.

**Ohne `image font size` wird weiter quadratisch gerechnet.** MSE setzt in
diesem Fall `img_size = 1` (`symbol_font.cpp`, Zeile 113) und zeichnete ein
165 Bildpunkte breites Symbol 2.475 Punkte breit. Das ist kein Verhalten,
dem man folgen will, und `msepaket` kann „nicht angegeben" von „als 0
geschrieben" ohnehin nicht unterscheiden. Die Abweichung steht hier, damit
niemand sie später für ein Versehen hält.

**Abgerundet wird nicht.** MSE rechnet in ganzen Bildpunkten des Geräts
(`wxSize(actual_size * (int)(size) / (int)(img_size))`); dieser Plan misst in
Vorlagenpunkten und kennt die Vergrößerung nicht. Bei Bildschirmgröße kostet
das bis zu einem Bildpunkt je Symbol, beim Druck wäre ein Abrunden auf
Vorlagenpunkte deutlich falscher als keines.
