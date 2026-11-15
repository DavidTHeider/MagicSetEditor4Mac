# 0058 — Eine außerplanmäßige Paritätsphase, direkt nach 4.5

**Datum:** 2026-08-31
**Zustand:** angenommen

## Der Anlass

Der Zweitblick über Phase 4.4 hat sechs Befunde gebracht, die keine Defekte
waren, sondern **Lücken zur Referenzimplementierung** — jeder von ihnen eine
Einheit, die MSE2 hat und dieses Programm nicht:

| Issue | Was fehlt | Wo es in MSE2 steht |
| --- | --- | --- |
| #1337 | Karten aus einem Skript anlegen — `add cards script` | `src/data/game.cpp` |
| #1338 | die erklärten Auszählungsdimensionen eines Spielpakets | `src/data/statistics.cpp` |
| #1339 | die Feldart `symbol` **je Karte** | `src/data/field/symbol.cpp` |
| #1340 | Rücknahme je Handgriff im Symboleditor, ein Bild nachzeichnen | `src/gui/symbol/` |
| #1341 | Autoersetzungsregeln verwalten | `src/gui/preferences_window.cpp` |
| #1342 | die Parameterhilfen der Schlüsselwörter | `src/data/keyword.cpp` |

Sie standen ohne Meilenstein da — nach §10 als Issues eröffnet, „weil sie den
Rahmen der Phase sprengten", und danach ohne Ort.

## Die Entscheidung

**Sie werden als eigene Phase abgewickelt, unmittelbar nach 4.5.**

Ihre Nummer ist **4.11**, und die sagt — wie bei Phase 9 gegenüber Phase 8 —
**wann sie geschnitten wurde, nicht wann sie gebaut wird** (§13). Gebaut wird
sie vor 4.6.

## Warum nicht anders

**Nicht in 4.6 hineinnehmen.** §13 sagt es ausdrücklich: „eine Phase wird nicht
stillschweigend größer. Kommt Umfang hinzu, bekommt er ein eigenes Issue und
eine sichtbare Begründung in der Roadmap." Sechs Einheiten mit Domäne,
Paketleser, Rasterung und Oberfläche sind zusammen mehr Arbeit als Phase 4.5
im Ganzen; sie in „Decks und Formatprüfung" zu schieben hieße, eine Phase zu
haben, deren Name nicht mehr sagt, was in ihr geschieht.

**Nicht hinten anstellen.** Der Anspruch aus der Roadmap lautet:
*MagicSetEditor4Mac enthält die Fähigkeiten von MSE2 als echtes Subset und geht
darüber hinaus.* Solange sechs benannte Einheiten fehlen, ist der Satz falsch —
und jede Phase, die davor gebaut wird, baut auf einem Fundament, dessen Lücken
bekannt sind. Drei der sechs stehen zudem in `PARITAET.md` mit der Planung
„0.3.0", und die Fassung ist am 2026-08-28 erschienen.

**Nicht ohne Nummer als „Nacharbeit" führen.** Eine Arbeit ohne Phase hat
keinen Abschluss: keinen Zweitblick, keinen Nachlauf, keine Fortschreibung.
Genau das ist der Zustand, aus dem diese sechs Issues kommen.

## Was sie kostet

**Die Reihenfolge der übrigen Phasen verschiebt sich um eine.** 4.6 bis 4.10
rücken hinter 4.11 — die Ziele bleiben, der Zeitpunkt nicht.

**Und eine Aussage über den Umfang steht nicht hier**, weil sie niemand
belegen kann: was #1340 wirklich kostet, weiß erst, wer `src/gui/symbol/`
gelesen hat (§15, Schritt 1).
