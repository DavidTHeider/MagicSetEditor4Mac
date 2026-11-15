# Die Stiloptionen des Packs gegen ihre Bedienstelle

Was ein Stilpaket als `styling field:` erklärt, und wo es im Fenster bedient
wird. Die Datei entsteht aus einer **Zählung**, nicht aus einer Schätzung —
und ein Wächter (`internal/werkzeuge/stiloptionen_test.go`) hält sie gegen den
Quelltext.

**Grundlage:** #630,
Teil von #590.

---

## Worüber gezählt wurde

Am **2026-08-22** über `full-magic-pack/data`: **377** `style`-Dateien,
**4.067** `styling field:`-Blöcke. Gezählt am **Text** der Stildateien — je
Block die erste `type:`-Zeile.

**Hier standen 376 und 4.054**, gezählt am 2026-08-20. Der Zweitblick über
Phase L hat den Widerspruch angestrichen: `docs/FORMATE.md` führte längst
4.067. Nachgezählt sind es 377 Dateien — das Pack hat 376 `*.mse-style`
-Verzeichnisse, und eine `style`-Datei liegt woanders.

**Die Reichweite der Messung ist die Reichweite ihrer Aussage** (AGENTS.md §10).
Diese Zahlen gelten für dieses Pack an diesem Tag. Ein anderes Pack kann andere
Arten führen; die Zeile „was nicht bedient wird" bliebe dieselbe.

---

## Die Arten

| Art | Vorkommen | in Paketen | Bedienstelle | Trägt |
| --- | ---: | ---: | --- | --- |
| `boolean` | 1.489 | 364 | Kästchen — oder zwei Bildkacheln, wo die Vorlage welche erklärt (#1225) | ✅ |
| `choice` | 1.051 | 362 | Auswahlliste — mit Bildern, wo die Vorlage welche erklärt (#1225) | ✅ |
| `text` | 738 | 180 | Textfeld | ✅ |
| `package choice` | 629 | 349 | Aufklappliste der passenden Pakete | ✅ |
| `color` | 105 | 57 | Auswahlliste der benannten Farben und ein Farbwähler | ✅ |
| `multiple choice` | 33 | 21 | Kästchen, eines je Wahl | ✅ |
| `info` | 13 | 2 | Überschrift — sie trägt keinen Wert | ✅ |
| `slider` | 9 | 3 | Textfeld | ❌ |
| `image` | 0 | 0 | Textfeld | ❌ |
| `symbol` | 0 | 0 | Textfeld | ❌ |

**`symbol` kommt als `styling field:` ebenso wenig vor** (gezählt am
2026-08-31, dieselbe Messung wie oben). Die Art gibt es seit
#1339, weil
**Karten- und Setfelder** sie führen: `type: symbol` steht im Pack neunmal in
sieben Dateien, jedes Mal als `set field:`.

**`image` kommt als `styling field:` in diesem Pack nicht vor** — `msepaket`
kennt die Art trotzdem, weil **Kartenfelder** sie führen. Sie steht hier mit
ihrer Null, statt zu fehlen: eine Art, die man nicht sieht, hält man leicht für
eine, die es nicht gibt.

**Bedient sind 4.058 von 4.067 — 99,8 %.** Übrig bleiben die **9** `slider`,
und die kennt MSE selbst nicht.

**Und „bedient" heißt hier: es gibt ein Bedienelement — nicht, dass es
aussieht wie bei MSE2.** Am 2026-08-27 nachgezählt, im Rahmen von Werkstück 5
der Phase 3.6 (#1167):
im `styling style:`-Abschnitt stehen **1.096** `choice images:`-Blöcke in
**303** von **376** Paket-Hauptdateien, darunter **400**
`symbol_variation`-Aufrufe.

**Zwei Grundgesamtheiten, und sie sind zu unterscheiden** (Befund 11 des
Zweitblicks über Phase 3.6): der Abschnitt oben zählt **377** `style`-Dateien,
weil eine davon nicht die Hauptdatei eines Pakets ist
(`magic-m15-future.mse-style/new/style`). Über alle 377 sind es **1.100**
Blöcke in **304** Dateien und **402** Aufrufe. Hier stand „303 von 377" — ein
Zähler aus der einen und ein Nenner aus der anderen Menge.

**Und gezählt wird ohne Kommentare** (Befund 7 des Nachlaufs über Phase 3.6).
Hier standen 1.097 und 1.101; die Differenz ist genau **eine** auskommentierte
Zeile — `#		choice images:` in `magic-genevensis-00-main.mse-style/style`,
Zeile 1417. Ein `#` am Anfang der eingerückten Zeile ist für MSE kein Block,
und eine Textzählung, die das übersieht, zählt eine Bedienstelle, die es nicht
gibt. Die 400 beziehungsweise 402 `symbol_variation`-Aufrufe sind davon nicht
betroffen. MSE2 zeigt daraus **je Wahl ein Bild**; die
Auswahlliste hier zeigte die **Namen**. Wer zwischen `modern`, `old` und
`diagonal T` wählen sollte, sah drei Wörter statt dreier Tapsymbole.

**Seit dem 2026-08-28 zeigt sie die Bilder**
(#1225) —
und zwar als **Reihe von Kacheln** statt als nachgebauter Auswahlliste: ein
`<option>` trägt in HTML keinen Bildinhalt, und eine Reihe von Optionsfeldern
ist dieselbe Sache mit fertiger Tastaturbedienung. Dass das nicht ausufert,
sagt das Material: **kein** `choice images:`-Block führt mehr als **fünf**
Wahlen — 591 führen zwei, 500 drei, fünf führen fünf (gezählt am 2026-08-28).

**Die Spalte „Bedienstelle" sagt seitdem „Auswahlliste, mit Bildern, wo die
Vorlage welche erklärt".** Es hier zu vermerken war der Punkt: eine Tabelle,
die „✅" sagt und eine halbe Zusage meint, ist genau der Wächter, der die
andere Hälfte unsichtbar macht (§10).

**Bis #632
waren es 80,6 %**, bis #634
96,1 %, bis #637
98,6 %, bis #639
99,5 %.

**`info` war kein Bedienelement, sondern eine Frage.** Die Art ist in MSE
**nicht bearbeitbar** (`InfoField() { editable = false; }`), und alle 13
Vorkommen sind **Zwischenüberschriften** in langen Optionslisten — „Advanced
customization", „Other stuff". Ein beschreibbares Feld dafür war nicht zu
wenig, sondern zu viel: es schrieb einen Wert in den Bestand, den kein Skript
liest.

**`slider` kennt `msepaket.feldarten` nicht** (`internal/msepaket/spiel.go`):
sie wird `KindUnknown` und fällt als Text durch — 9 Vorkommen.

---

## Was MSE2 dafür anzeigt

Alle Fundstellen Fassung `f9d9356` vom 2021-07-06.

| Art | MSE2 |
| --- | --- |
| `color` | Aufklappliste der `choice:`-Blöcke mit Farbkachel; dazu ein **Vorgabeeintrag**, wenn `default_script` dasteht, und ein **eigener** Eintrag, wenn `allow custom` gilt — der öffnet einen Farbwähler (`src/gui/value/color.cpp:37–95`) |
| `package choice` | Aufklappliste der **installierten Pakete**, die auf `match:` passen — `package_manager.findMatching`, sortiert nach `position_hint`, dann nach Namen. Gespeichert wird der **Dateiname** des Pakets, angezeigt sein `short_name` mit seinem 16×16-Symbol (`src/render/value/package_choice.cpp:28–43`). Ist `required: false`, steht ein erster Eintrag mit `empty_name` (Vorgabe `none`) für den leeren Wert (`src/gui/value/package_choice.cpp:44–59`) |
| `multiple choice` | Kästchen je Wahl; der Wert steht in einer **Normalform** — der Reihenfolge, in der die Wahlen erklärt sind (`src/data/field/multiple_choice.cpp:82–110`) |
| `info` | **Nicht bearbeitbar**: `InfoField() { editable = false; }` (`src/data/field/information.hpp:26–34`) |
| `symbol` | Das Feld trägt **keine eigenen Angaben** — im Quelltext steht wörtlich „no extra data" (`src/data/field/symbol.hpp:26–31`). Was es aussieht, entscheidet der **Stil**: `min_aspect_ratio` und `max_aspect_ratio` (beide Vorgabe 1) klemmen das Seitenverhältnis, und `variations` führt die Spielarten derselben Zeichnung — jede mit `name`, einem `filter` und `border_radius` (Vorgabe **0,05**), gezeichnet nebeneinander mit zwei Stellen Abstand (`src/data/field/symbol.cpp:24–46`, `src/render/value/symbol.cpp:22–52`). Der **Wert** ist ein Dateiname im Paket, keine eingebettete Zeichnung |
| `slider` | **Gibt es dort nicht.** `read_field` wirft auf einen unbekannten `type:` einen `ParseError` und bricht das Lesen ab (`src/data/field.cpp:75–93`) — MSE2 dieser Fassung öffnet die drei betroffenen Pakete **nicht**. Dass wir sie als Text durchreichen, ist großzügiger als die Referenz und richtig so (§8) |

---

## Was `package choice` verlangt

Von den 627 Vorkommen: **331** ohne `required:` (Vorgabe `true`), **296** mit
`required: false`.

| Vorkommen | `match:` |
| ---: | --- |
| 308 | `magic-mana-*.mse-symbol-font` |
| 296 | `magic-overlay-*.mse-include` |
| 3 | `magic-pt-symbols-*.mse-symbol-font` |
| 1 | `*.mse-symbol-font` |

Der Paketkatalog beantwortet das seit #632 über `PaketeNachMuster`: die Endung
ist die Spalte `art`, der Rest ein `GLOB` auf `name`, sortiert nach `rang` und
dann Namen — genau MSEs Ordnung.

**Das Symbol fehlt.** MSE zeigt neben jedem Eintrag das 16×16-Bild des Pakets;
ein `<select>` trägt keine Bilder. Eine eigene Aufklappliste wäre ein eigenes
Werkstück, und bis dahin steht der Name allein da.

## Was `color` verlangt

Von den 105: **89** mit `choice:`-Blöcken, **16** ohne — die brauchen den
freien Farbwähler, nicht die Liste.

**Keine einzige verbietet die eigene Farbe**: `allow custom` ist ohne Angabe
`true` (`src/data/field/color.cpp:15–18`), und im ganzen Pack steht die Angabe
nirgends. Der Farbwähler ist damit nicht die Ausnahme, sondern der Normalfall.

**Beides steht seit dem 2026-08-21.** Hier standen bis zum 2026-08-22 zwei
Sätze nebeneinander — der eine sagte, der Vorgabeeintrag fehle, der andere, er
stehe. Der Zweitblick über Phase L hat den Widerspruch angestrichen.

**Gerechnet** wird `default:`: es ist ein Skript, es schlägt `initial:`, und was
dabei herauskommt, steht als Wert der Option da (`service.vorgabeAusAusdruck`,
`docs/FORMATE.md`). Es sind **10** der 105, alle vom Typ `color`.

**Der Eintrag** steht seit demselben Tag. MSE stellt ihn einer Farboption voran
— er sagt „nimm, was das Skript rechnet" — und er bleibt gewählt, denn
`change(Defaultable<T>&, …)` ruft `assignDontChangeDefault`
(`src/script/scriptable.hpp:40`): der Wert bleibt nach dem Skript **Vorgabe**.

Dafür führt eine Option jetzt **zwei** Angaben statt einer: `gesetzt` sagt, ob
der Wert aus dem Bestand kommt, `vorgabe` trägt das gerechnete `default:` und
ist leer, wenn die Vorlage keins nennt. Damit gibt es auch die Handlung, die
vorher fehlte: **zurück auf die Vorgabe** — sie entfernt den gespeicherten
Wert, so wie `select(0)` einen leeren `Defaultable` einsetzt.

**Beschriftet ist der Eintrag immer gleich**: `default name:` steht im ganzen
Pack **kein einziges Mal**, und ohne Angabe ist es das übersetzte Wort
„Default" (`src/data/field/color.cpp:17`).

**Und er steht auch dann, wenn das Skript nicht aufgeht.** `hasDefault()` fragt
nach dem **Skript**, nicht nach seinem Ergebnis (`color.cpp:38`).
`magic-m15-bigtext` schreibt `default: black` ohne Anführungszeichen — ein
Skriptfehler, der Wert fällt auf `initial:` zurück, und bis zum 2026-08-22 ließ
das den Eintrag bei uns verschwinden. Genau bei der einen Option des Packs, die
ihn gebraucht hätte; gefunden vom Zweitblick über Phase L.

---

## Was der Wächter nicht prüft

**Die Zahlen.** Er kann sie nicht nachzählen: das Pack liegt nicht im
Repositorium ([0005](entscheidungen/0005-keine-fremden-vorlagen.md)), und ein
Lauf dagegen liefe nicht in der Pipeline. Er hält die **Namen** — jede Art, die
`Stiloptionen.tsx` bedient, steht hier, und jede, die hier steht, kommt dort
vor.

**Ob die Bedienstelle die richtige ist.** Dass `color` ein Textfeld bekommt,
sieht er als „bedient von der Rückfallstelle" und nicht als Mangel. Der Mangel
steht in der Spalte „Trägt", und die liest ein Mensch.

**Ob das Pack noch dieselben Arten führt.** Ein neues Paket mit einer zehnten
Art fiele hier durch und stünde als Text im Fenster — bemerkt würde es beim
nächsten Zählen, nicht von der Pipeline.

---

## Was `multiple choice` verlangt

Gezählt am 2026-08-20 über die 33 Vorkommen:

| Angabe | Vorkommen |
| --- | ---: |
| `initial:` | 18 |
| `empty choice:` | **0** |
| `minimum selection:` | **0** |
| `maximum selection:` | **0** |

**Der Wert steht in einer Normalform**: `"wahl, wahl"` in der Reihenfolge, in
der die Wahlen **erklärt** sind, nicht in der des Anklickens
(`src/data/field/multiple_choice.cpp:82–113`). `service.normalform` gibt es
dafür seit #556;
seit #637
läuft auch eine **Stiloption** hindurch — beim Lesen wie beim Schreiben, wie
MSE `normalForm` in `update` ruft.

**Die Grenzen werden gelesen und nicht durchgesetzt.** `minimum selection` und
`maximum selection` kommen im ganzen Pack nicht vor; ein Bedienelement dafür
wäre ungeprüfte Mechanik.
