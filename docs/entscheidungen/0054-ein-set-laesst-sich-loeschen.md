# 0054 — Ein Set lässt sich löschen, und zwar endgültig

**Datum:** 2026-08-27
**Status:** angenommen
**Anlass:** #1218

## Der Anlass

Beim Aufräumen nach der Abnahme von Phase 3.5 lag ein Testset im Bestand, und
es ließ sich nur über `sqlite3` von Hand entfernen. **Es gab keinen Weg
hinaus** — an keiner der vier Flächen und nicht im Fenster.

Ein Vorgang, den die Anwendung selbst nicht kann, ist kein Betriebsweg. Er
verlangt von jedem, der aufräumen will, den Umgang mit einer Datenbank, deren
Schema er nicht kennt, und einen falschen Griff darin nimmt mehr weg als
gewollt.

## Was MSE2 tut

**Nichts.** Nachgelesen und nicht erschlossen: das Dateimenü führt `new`,
`open`, `save`, `save_as`, `save_as_directory`, `check_updates`,
`show_profiler`, `reload_data`, `print_preview`, `print` und `exit`
(`src/gui/set/window.cpp`, Zeilen 58–78, Fassung f9d9356 vom 2021-07-06).

Der Grund ist keine Nachlässigkeit, sondern ein anderes Datenmodell: **ein Set
ist dort eine Datei.** Sie wegzuwerfen ist Sache des Dateimanagers, und das
Programm hat damit nichts zu tun.

Hier liegt das Set im Bestand. Damit ist die Fähigkeit **`adapted`** und nicht
`new`: sie entsteht aus einer Plattformentscheidung dieses Projekts, nicht aus
einer neuen Idee — und §15 verlangt, dass `new` keine unterlassene Suche
verdeckt. Gesucht wurde im Dateimenü, in den Menübefehlen des Setfensters und
in `src/data/set.cpp`.

## Die Entscheidung

**Endgültig, bestätigt durch den Namen.**

`SetLoeschen(setID, name)` nimmt das Set samt allem, was an ihm hängt. Der
**Name muss der heutige sein**; stimmt er nicht, geschieht nichts.

### Warum der Name und nicht ein „Ja"

Er tut **zweierlei zugleich**, und das ist der Grund für ihn:

1. **Bestätigung.** Ein „Ja" lässt sich wegklicken, ohne gelesen zu haben —
   und der zweite Klick sitzt an der Stelle, an der eben noch etwas Harmloses
   stand. Ein Name, den man abtippt, verlangt einen Blick auf das, was gemeint
   ist. Dieselbe Überlegung steht hinter GitHubs Rückfrage beim Löschen eines
   Repositoriums.

2. **Wächter über den Stand** (§7, Regel 2). `sets` führt keine `fassung`.
   Wurde das Set zwischen Lesen und Löschen umbenannt — von der
   Kommandozeile, vom MCP-Server, von einem zweiten Fenster —, stimmt der Name
   nicht mehr, und es bleibt stehen. Die Meldung sagt dann genau das: „Der
   Eintrag wurde zwischenzeitlich geändert."

**Und er trägt an allen vier Flächen**, MCP eingeschlossen. Gerade dort: ein
Sprachmodell müsste `lesen.sets` gerufen haben, um den Namen nennen zu können.
Eine Kennung ließe sich raten.

## Was verworfen wurde

### Ein Papierkorb

Er müsste an **jedem** Leseweg filtern — Setliste, Volltextsuche, Kartenzahlen,
die Regel „eine Quelle, ein Set" beim Import. Einer, der nur die Liste
filtert, wäre eine Zusage, die nicht hält: die Suche fände Karten eines Sets,
das der Papierkorb als weggeworfen führt.

Dazu kommt, dass er eine zweite Frage aufwirft, die niemand gestellt hat: wann
wird er geleert? Ein Papierkorb, der nie geleert wird, ist ein Bestand, der nur
noch wächst.

**Wenn er gewünscht ist, ist er ein eigenes Werkstück** — und dann eines, das
alle Lesewege anfasst.

### Eine Rettungsdatei vor dem Löschen

Der erste Entwurf sah vor, das Set vorher als `.mse-set` wegzuschreiben. Das
scheitert an dem Fall, für den man es am ehesten bräuchte:
`Setschreiber.Schreibe` verlangt eine **Quelldatei** und weist ein Set ab, das
in der Anwendung entstanden ist. Genau die sind es, die man löscht.

Ein Ausweichen auf MTG JSON — das von Grund auf schreibt — verlöre alles, was
nicht Magic ist. Eine Rettung, die die Hälfte rettet, ist schlechter als eine
klare Ansage.

**Die Ansage steht deshalb in der Warnung**: es gibt kein Zurück, und was in
einer Datei liegt, bleibt dort.

## Was mitgeht und was nicht

| | |
| --- | --- |
| **Mit** | Karten, Kartenwerte, Setwerte, Rohblöcke, Quelldatei, Schlüsselwörter, Packarten — über die Kaskaden des Schemas |
| **Mit** | der Volltextindex. Sein Auslöser hängt an `card_values` und feuert auch bei einer **kaskadierten** Löschung; gemessen am 2026-08-27 bei `recursive_triggers = 0` |
| **Mit** | die eingesetzten Bilder auf der Platte — sie liegen unter `<wurzel>/<setID>/` und stehen in keiner Tabelle. Sie räumt `service.Setloeschung` |
| **Bleibt** | die **Sammlung** ([0052](0052-die-sammlung-ist-nicht-das-set.md)). Ein Exemplar hängt an einem Druck, nicht an einem Set. Wer sein Set löscht, hat die Karten nicht verkauft |
| **Bleibt** | jede Datei, aus der das Set einmal gelesen wurde. Gelöscht wird der Stand im Bestand |

**Die Bilder sind ein Befund und kein Fehler** (§7, Regel 3). Lassen sie sich
nicht wegräumen, ist das Set trotzdem weg — ein Fehler an dieser Stelle ließe
die Oberfläche ein Set anzeigen, das es nicht mehr gibt.

## Was diese Entscheidung nicht sagt

- **Ob jemand das Löschen bereuen wird.** Sie sagt, dass er es nicht aus
  Versehen tut.
- **Dass der Bestand danach kleiner ist.** SQLite gibt Seiten nicht von selbst
  zurück; ein `VACUUM` steht hier nicht an und wäre eine eigene Entscheidung.
