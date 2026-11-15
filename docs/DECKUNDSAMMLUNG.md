# Set, Deck und Sammlung

**Drei Dinge mit drei Bedeutungen — und was zwischen ihnen hin und her geht.**
FMT-04 (#1693).

Die Entscheidungen dahinter stehen in [0061](entscheidungen/0061-was-ein-deck-hier-ist.md)
(was ein Deck ist), [0052](entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)
(die Sammlung ist nicht das Set) und [0144](entscheidungen/0144-keine-neue-endung-fuer-deck-und-sammlung.md)
(keine neue Endung). Hier steht die **Abbildung**: wer was trägt, und was auf
welchem Weg verloren geht.

## 1 Die drei Dinge

| | Set — **Kartendefinition** | Deck — **Zusammenstellung** | Sammlung — **Besitz** |
| --- | --- | --- | --- |
| Frage | *Wie sieht die Karte aus?* | *Womit spiele ich?* | *Was habe ich in der Hand?* |
| Einheit | Karte mit Feldwerten, Stil, Bildern | **Platz**: eine eigene Karte **oder** ein Druck, mit **Menge** und **Bereich** (Hauptdeck, Sideboard) | **Exemplar**: ein Druck mit **Menge**, **Zustand**, Sprache, Ausführung, Erwerb (Datum, Preis, Währung), Lagerort, Notiz |
| Kennung | Kartenkennung im Bestand | Deck- und Platzkennung | Exemplarkennung |
| gehört | einem Profil | einem Profil | einem Profil |
| MSE2 kennt es | **ja** (`.mse-set`) | **nein** (0061) | **nein** (0052) |

**Ein Platz verweist, er kopiert nicht.** Wer eine Karte im Set berichtigt,
berichtigt sie in jedem Deck. **Ein Exemplar ist ein Druck**, keine eigene
Karte: eine selbst entworfene Karte hat keinen Druck, den jemand besitzen
könnte.

## 2 Die Abbildung

| Von → Nach | Weg | Umkehrbar? | Was dabei bleibt, was nicht |
| --- | --- | --- | --- |
| Set → Deck | **Deck aus einem Set** (`DeckAusSet`, `mse deck aus-set`, `schreiben.deck.ausset`, `POST /decks/aus-set`) — jede Karte ein Platz im Hauptdeck, dieselbe Menge | nein — ein Deck macht kein Set | Die Karten bleiben im Set; das Deck verweist. Die Herkunft des Decks nennt das Set |
| Karte → Deckplatz | ein Platz auf die eigene Karte | ja: der Platz nennt die Karte | nur **im selben Bestand** — die Kartenkennung gilt nirgends sonst |
| Druck → Deckplatz | ein Platz auf die Scryfall-Kennung | ja | überall — eine Druckkennung ist weltweit dieselbe |
| Exemplar ↔ Deckplatz | **Zuordnung** (U9) — welche Stücke für welchen Platz vorgesehen sind | ja | Zugeordnet wird nie mehr, als da ist; ein Exemplar kann in mehreren Decks vorgesehen sein |
| Sammlung → Deck | „Aus der Sammlung" legt Stücke ins Deck (UX-03) | über die Zuordnung | nur Drucke |
| Deck → Sammlung | Fehlkarten auf die **Wunschliste** (DOM-03) | nein | Wünsche sind Zeilen der Sammlung, keine Exemplare |

**Mehrfach vorhandene Karten** sind zwei Plätze, wenn sie in zwei Bereichen
liegen, und eine höhere Menge, wenn im selben (0061). **Getrennte Exemplare**
desselben Drucks — anderer Zustand, anderer Erwerb — bleiben getrennt; sie zu
einer Menge zusammenzuziehen hieße, Zustand oder Preis zu verlieren.

## 3 Was welcher Behälter trägt

| Behälter | Trägt | Trägt nicht | Bezeichnung an der Oberfläche |
| --- | --- | --- | --- |
| **`.mse-set`** — MSE2-kompatibel | Kartendefinitionen, Stil, Bilder, Setangaben | Decks, Exemplare, Marken, Herkunft der Werte, Verweise auf Drucke — siehe §4 | Set |
| **Umschlag mit Deck** (`.msepackage`) | Name, Spiel, Notiz, jeder Platz auf einen **Druck** mit Menge und Bereich | **Plätze auf eigene Karten** (ihre Kennung gilt im Bestand des Absenders — sie kommen als Befund mit Namen, #1693), Zuordnungen, Format | Deckdatei |
| **Umschlag mit Sammlungssicht** | was freigegeben ist, zum **Ansehen** | es wird nichts gebucht (T7) | Sammlungsdatei |
| **CSV der Sammlung** | jedes Exemplar mit jeder Angabe, auch getrennte Stücke desselben Drucks | Zuordnungen zu Decks, Wünsche | Sammlungsdatei |
| **Deckliste** (Text) | Menge, Name, Set und Nummer, Sideboard | Zuordnungen, eigene Karten ohne Druck, Format | Deckliste |

**Die Roundtrips dazu sind gelaufen** (AK-04): ein Deck mit Karten aus zwei
Sets und demselben Druck in Haupt- und Nebendeck durch den Umschlag
(`internal/app/deckumschlagrundlauf_intern_test.go`), eine eigene Karte des
Absenders, deren Kennung hier eine andere Karte träfe
(`internal/app/deckfremdkarte_intern_test.go`), zwei Exemplare desselben
Drucks durch die CSV (`internal/flaechen/getrennteexemplare_test.go`) und ein
Deck aus einem Set an allen drei Flächen ohne Fenster
(`internal/flaechen/deckausset_test.go`). **Fehlende Pakete** berühren ein
Deck nicht: ein Platz verweist auf eine Karte, nicht auf ihren Stil; ein Set,
dem ein Paket fehlt, zeigt die Vorschau vor dem Einlesen (FMT-03).

## 4 MSE2-kompatibel heißt: das Set, und nur das Set

**Ein Deck oder eine Sammlung „als MSE2" gibt es nicht** — MSE2 hat für beides
keine Gestalt. Deshalb schreibt dieses Programm **nichts** unter einer
MSE-Endung, was MSE2 nicht lesen könnte; eine Datei umzubenennen macht sie
nicht kompatibel (AK-03).

Was ein Set hier **mehr** trägt als in seiner `.mse-set`-Datei, nennt die
Sicherung **vorher** (FMT-05, #1694):
Decks, die auf seine Karten verweisen, Marken, die Herkunft einzelner Werte.
Die Datei selbst bleibt, was MSE2 erwartet.
