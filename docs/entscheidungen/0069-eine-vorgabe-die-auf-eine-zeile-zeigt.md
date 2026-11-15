# 0069 — Eine Vorgabe, die auf eine Zeile zeigt

**Entschieden am 2026-09-05**, Werkstück 15 von Phase 4.12
(#1351),
Punkte 4 bis 7 der Kommentierung.

## Der Anlass

Die Kommentierung nennt dreizehn Profil-Einstellungen. Neun davon trägt die
vorhandene Mechanik: `internal/domain` kennt die erlaubte Liste,
`PruefeEinstellung` urteilt danach, das Fenster baut daraus seine Auswahl. Ein
Kartenzustand, eine Turnierbauart, ein Pfad.

**Die Punkte 4 bis 7 sind etwas anderes.** Ein Vorgabeformat, ein Vorgabespiel,
eine Vorgabevorlage, Präferenzdecks, zugeordnete Sammlungen — sie zeigen auf
**Zeilen im Bestand**. Ihre erlaubten Werte stehen nicht in der Domäne, sie
ändern sich mit dem Bestand, und zwei davon sind **mehrwertig**.

Der Auftraggeber hat sie deshalb zu einem eigenen Werkstück gemacht (§13:
eine Phase wird nicht stillschweigend größer).

## 1. Eine zweite Mechanik neben den Einstellungen

**Und nicht `PruefeEinstellung` erweitert.** Das hieße, zwei Sachen unter einem
Namen zu führen: eine Prüfung gegen eine feste Liste und eine gegen den
Bestand. Danach könnte die eine die andere nicht mehr halten — und der Wächter
über die Einstellungswerte prüfte die Hälfte dessen, was er zu prüfen scheint.

**Der Preis ist eine zweite Fläche an jedem Zugang**: `Profilvorgaben` neben
`Einstellungen`, `lesen.profilvorgaben` neben `lesen.einstellungen`. Er ist
tragbar, weil der Unterschied echt ist und beim Lesen sofort auffällt.

## 2. Dieselbe Liste prüft und bietet an

**`Moeglich` ist beides.** Jede Antwort trägt, was ein Profil wählen kann, und
gegen genau diese Liste wird beim Setzen geurteilt.

**Zwei Listen liefen auseinander**, sobald jemand ein Format löscht: die
Auswahl zeigte es nicht mehr, die Prüfung ließe es noch durch — und der
Fehlschlag käme erst da, wo jemand die Vorgabe benutzt.

## 3. Eine Vorgabe überlebt das Löschen dessen, worauf sie zeigt

**Kein Fremdschlüssel auf `formate`, `decks` oder `sammlung`.** Drei Gründe,
und der dritte ist der eigentliche:

1. Zwei der fünf Arten zeigen gar nicht auf eine Tabelle — `spiel` und
   `vorlage` nennen ein **Paket** auf der Platte.
2. Ein Fremdschlüssel bräuchte je Art eine eigene Spalte, und damit fünf
   Spalten, von denen immer vier leer sind.
3. **Ein `ON DELETE CASCADE` nähme die Vorgabe still weg, ein `RESTRICT`
   verböte das Löschen des Decks.** Beides ist falsch: was ins Leere zeigt,
   wird **gemeldet** (§8), und wer es sieht, entscheidet.

**`verwaist` ist deshalb ein Befund und kein zweiter Stand**: der Wert steht
weiter in `werte`, und daneben steht, dass es ihn nicht mehr gibt.

## 4. Die Reihenfolge ist bei den Präferenzdecks die Aussage

**Deshalb eine Liste und keine Menge**, und deshalb hält die Tabelle eine
`ordnung`. Wer ein Deck dazunimmt, meint es als das **nächste** — angehängt und
nicht einsortiert.

**In der Deckliste wird sortiert und nicht gefiltert.** Wer ein Deck bevorzugt,
will die anderen nicht verlieren; eine Liste, die sie verschwinden ließe, wäre
eine Falle für den, der einmal etwas anderes sucht.

## 5. Jede Vorgabe wirkt dort, wo man sie sieht

**Eine Vorgabe ohne Wirkung wäre eine Zeile im Bestand.** Fünf Arten, fünf
Stellen:

| Art | wirkt |
| --- | --- |
| `spiel`, `vorlage` | das Formular für ein neues Set wählt sie vor — wie MSE2, das dafür `settings.default_game` und `gameSettingsFor(*game).default_stylesheet` liest (`src/gui/new_window.cpp`, Zeilen 53 und 66, Fassung f9d9356 vom 2021-07-06). Dort ist es das zuletzt benutzte, hier eine ausdrückliche Wahl |
| `format` | ein neu angelegtes Deck bekommt es gleich gesetzt — in einem **zweiten** Ruf, weil `DeckAnlegen` keines entgegennimmt und ihm eines beizubringen eine Fläche änderte, die seit Phase 4.6 steht (§16) |
| `lagerort` | eine neue Sammlungszeile beginnt damit — **der erste** von mehreren: eine Zeile liegt an **einem** Ort |
| `deck` | die Präferenzdecks stehen in der Liste zuerst |

**Eine Vorgabe, die ins Leere zeigt, wird beim Vorwählen übergangen.** Sonst
stünde das Formular auf einem Wert, den seine Liste nicht kennt — und der Knopf
wäre tot, ohne dass jemand sähe, warum.

## Was verworfen wurde

**Eine kommagetrennte Liste in einer Spalte.** In SQLite ist das keine Liste:
wer sie abfragen will, zerlegt Text; wer einen Wert löschen will, schreibt die
ganze Zeile neu; und ein Wert mit Komma darin bricht alles.

**Ein `Wirksam`, das verwaiste Werte verschluckt.** Wer eine Vorgabe benutzt,
bekommt vom Bestand ein Nein mit der richtigen Meldung; sie im Dienst
wegzufiltern machte aus einer falschen Vorgabe eine unsichtbare.

**Ein Fehler, wenn die Vorwahl nicht zu holen ist.** Eine Vorgabe ist
Bequemlichkeit: wer ein Set anlegen will, soll das können, auch wenn der Ruf
scheiterte. Eine Meldung darüber stünde im Weg und sagte nichts, was jemand
täte.
