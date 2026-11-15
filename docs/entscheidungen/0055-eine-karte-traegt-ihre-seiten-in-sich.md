# 0055 — Eine Karte trägt ihre Seiten in sich, und wir erfinden keine zweite

**Datum:** 2026-08-27
**Anlass:** Werkstück 3 von Phase 3.6 (#1167) — „Varianten, Tokens und DFCs sind keine Fachobjekte" (Audit Z1-03)
**Stand:** entschieden und gebaut

## Die Frage

Das Audit schreibt: „**Keine Beziehung Hauptkarte ↔ Rückseite**, keine
Variantenkennung, keine Exportauswahl entlang solcher Beziehungen." Das Issue
verlangt „explizite Beziehungen, die den MSE- und MTG-JSON-Rundlauf
**überleben**".

Naheliegend wäre: eine Kante `karte → rückseite` im Bestand, mit einer Spalte
`rueckseite_id` und einem Verweis in beide Richtungen.

**Sie wäre falsch.** Und zwar nicht, weil sie unbequem ist, sondern weil MSE
Rückseiten anders führt — und weil eine erfundene Kante genau das nicht täte,
was das Issue verlangt: den Rundlauf überleben.

## Was nachgelesen wurde

Drei Quellen (§11), alle am 2026-08-27:

| Quelle | Gefunden |
| --- | --- |
| **MSE2**, `src/data/card.hpp`, Fassung `f9d9356` vom 2021-07-06 | `Card` führt `data`, `notes`, `time_created`, `time_modified`, `stylesheet`, `styling_data`, `extra_data`, `keyword_usage`. **Keinen Bezug auf eine andere Karte.** Gesucht wurde zusätzlich über `src/` nach `back`, `backside`, `double.faced`, `flip`, `transform` — keine Fundstelle, die zwei Karten verbindet |
| **Die Datenpakete**, `magic.mse-game` | 249 Kartenfelder, davon **43 mit numerierten Geschwistern** (`name`/`name 2`/`name 3`/`name 4`, `casting cost`/`casting cost 2`, …), höchste Stelle **8**. Dazu **40** Felder mit Zahl **ohne** Grundfeld — `loyalty cost 1` bis `loyalty cost 16` sind die Zeilen eines Planeswalkers, keine sechzehn Seiten |
| **[`fenhl/magic-set-generator`](https://github.com/fenhl/magic-set-generator)** (MIT), `src/mse.rs` | Sein Makro `push_alt!` hängt bei der zweiten Seite ` 2` an **jeden** Schlüssel und legt beide Seiten in **dieselbe** Karte — für `Layout::DoubleFaced`, `Flip`, `Split`, `Meld` und `Adventure` |

**Die Antwort ist damit gemessen und nicht erschlossen:** eine mehrseitige
Karte ist in MSE **eine** Karte mit zwei Feldsätzen.

## Und die erfundene Kante überlebte den Rundlauf nicht

`Reader::unknownKey()` überspringt einen unbekannten Schlüssel und verliert ihn
beim Speichern still ([0053](0053-marken-liegen-neben-dem-set.md)). Eine
`rueckseite:`-Zeile in einem `card:`-Block wäre nach einem Umweg durch MSE2
fort — und niemand sähe, wann. Die Beziehung, die den Rundlauf überlebt, ist
die **im** Kartensatz.

**Und `magic-set-generator` baut sein Set nicht aus Scryfall.** Hier stand es
zuerst anders; nachgelesen am 2026-08-27 (Befund 10 des Zweitblicks): die
Kartendaten kommen aus einer **MTG-JSON-Datenbank** (`src/lib.rs`,
`Db::from_mtg_json` bzw. `Db::download`), und Scryfall liefert dort **nur**
Künstler und `art_crop` über `/cards/named` (`src/art.rs`, Zeilen 238–256).
Am Befund ändert das nichts — die Karte trägt ihre Seiten dort ebenso in sich —,
an der Begründung schon: vom Bildweg auf den ganzen Erzeugerpfad zu schließen
ist genau die Klasse, vor der §10 warnt.

**Für MTG JSON gilt dasselbe von der anderen Seite**: dort ist je Seite ein
eigenes Kartenobjekt, verbunden über `otherFaceIds`. `internal/mtgjson` legt
sie beim Lesen zusammen — seit dem Zweitblick über Phase V, und aus demselben
Grund.

## Die Entscheidung

**Der Fachbegriff ist die Feldreihe, nicht die Rückseite.**

**Und die Regel dafür ist `adapted`, nicht `ported`** (Befund 8 des
Zweitblicks): die **Gestalt** steht in allen drei Quellen, der
**Erkennungsalgorithmus** in keiner. MSE2 braucht ihn nicht — dort zeichnet das
Stilpaket, was es kennt, und niemand fragt, ob `loyalty cost 16` eine
sechzehnte Seite ist. Ein Formular, das hundert Felder auflistet, muss es
fragen.

`domain.Game.Feldreihen()` leitet aus den Feldern eines Spiels ab, welche eine
Reihe bilden. Die Regel ist ein Satz:

> Ein Feld `F N` gehört zur Reihe von `F`, **wenn es das Feld `F` gibt.**

Ohne die zweite Hälfte wäre `loyalty cost 16` eine sechzehnte Seite.

**Und sie heißt „Reihe" und nicht „Seite", weil die Zahl nicht immer eine Seite
meint.** `additional credit 2` bis `additional credit 5` sind fünf Zeilen einer
Danksagung; `magic-set-generator` benutzt `rule text 3` für das Schlüsselwort
*Fuse* (`src/mse.rs`, Zeile 293). **Was eine Stelle bedeutet, entscheidet das
Stilpaket**: `magic-m15-mainframe-dfc.mse-style` zeichnet `name` und `name 2`
als Vorder- und Rückseite, `magic-m15-flip.mse-style` als oben und unten.

Damit ist auch die **Variantenkennung** beantwortet: sie ist der **eigene Stil
der Karte** — `Card::stylesheet` bei MSE2 (`src/data/set.cpp`, Zeilen 90–93),
hier der Feldwert `stylesheet`, und `service.StilFuerKarte` liest ihn seit
#395. Eine
zweite Kennung daneben wäre eine zweite Wahrheit über dieselbe Karte.

## Und Tokens?

**Befund 7 des Zweitblicks über Phase 3.6**: die Überschrift des Issues nennt
„Varianten, **Tokens** und DFCs", und untersucht waren nur die ersten und die
letzten. Nachgeholt am 2026-08-27:

| Wo | Gefunden |
| --- | --- |
| **MSE2**, `src/data/` | **Kein Treffer** auf `token` in `.hpp` oder `.cpp`. Ein Token ist dort kein Fachbegriff |
| **Die Datenpakete** | **14** Pakete mit `token` im Namen — `magic-m15-token.mse-style`, `magic-m15-mainframe-tokens.mse-style` und weitere. Alles **Stilpakete** |
| **`magic-set-generator`** | Schließt Tokens beim `!all`-Weg ausdrücklich aus (`README.md`), und sein Layoutzweig kennt nur `Normal`, `Split`, `Flip`, `DoubleFaced`, `Meld`, `Adventure` |

**Damit ist ein Token dasselbe wie eine Variante: eine Karte mit einem eigenen
Stil.** Es ist keine eigene Fachklasse, es hat keine eigene Beziehung, und es
braucht keinen eigenen Datensatz — `service.StilFuerKarte` liest den Stil einer
Karte seit #395.

**Was daraus folgt und hier nicht steht**: ob die *Oberfläche* das Wechseln des
Kartenstils bequem genug macht. Das ist eine Frage an Phase 3.7 und keine an
das Datenmodell.

## Was das kostet

**Es gibt keine Liste „alle Rückseiten".** Wer sie will, fragt nach Karten mit
einem Wert an Stelle 2 — das ist dieselbe Auskunft, nur ohne die Behauptung,
dass Stelle 2 immer eine Rückseite ist.

**Und die Deutung bleibt beim Menschen.** Ob Stelle 2 eine Rückseite, eine
untere Hälfte oder eine zweite Danksagungszeile ist, sagt der Stil — und den
sieht man an der gezeichneten Karte.

**Eine Zählung „diese Karte belegt zwei Stellen" gibt es (noch) nicht.** Hier
stand sie als Zusage, und `domain.Stellenzahl` war gebaut — **ohne Aufrufer**
(Befund 6 des Zweitblicks). Sie ist wieder entfernt: ein Paket ohne Aufrufer
ist keine Anwendungsfunktion (§2), und einen Aufrufer zu erfinden, damit eine
Funktion bleiben darf, ist die falsche Richtung. **Was ausgeliefert ist, ist
die Regel** — `reihe` und `stelle` stehen an jeder Feldbeschreibung, und die
Massenänderung gruppiert danach. Die Zählung entsteht, wenn eine Fläche sie
braucht; die nächstliegende ist die **wählbare Spalte** der Kartenliste
(`card list name: Name 2`), und die steht in Phase 3.7.

## Verworfen

| Vorschlag | Warum nicht |
| --- | --- |
| Spalte `rueckseite_id` an `cards` | Überlebt den Rundlauf durch MSE2 nicht (oben). Und sie widerspräche beiden Referenzen: dort ist es **eine** Karte |
| Eine feste Liste „diese Felder sind Seiten" | Wäre die Verfestigung der Konvention `name 2`, die das Issue ausdrücklich ausschließt — und ginge an jedem Spiel vorbei, das nicht Magic ist. Die Reihen kommen aus dem Paket |
| Stelle 2 als „Rückseite" benennen | Eine Bedeutung, die die Quellen nicht hergeben: `additional credit 2` ist keine Rückseite. §15 verbietet, Fachlogik aus Feldnamen abzuleiten |
