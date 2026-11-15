# 0148 — Die Liste der Handel: abgeleitete Lage, Entwürfe bis zum Anbieten, ein Archiv als Ansicht

**Stand:** entschieden am 2026-09-30 · TRADE-01.AK-01 bis AK-04
(#1715)

## Die Frage

TRADE-01 verlangt, dass eigene laufende, wartende, abgeschlossene und
abgebrochene Vorgänge an allen vier Flächen abzufragen sind (AK-01), dass
Suche, Sortierung und Filter nach Partner, Status, Zeit, Karten und Tags
direkt zum Vorgang und seiner nächsten zulässigen Handlung führen (AK-02),
dass Entwürfe angelegt, gespeichert, bearbeitet, kopiert und gelöscht werden
— und abgeschlossene oder verbindliche Vorgänge eigene Regeln statt
„unkontrolliertem CRUD" bekommen (AK-03) —, und dass die Liste veraltete oder
unvollständig synchronisierte Daten als solche zeigt (AK-04).

Offen war, **woher** „du bist dran" kommt, **was** ein Entwurf ist und was
mit einem Handel geschieht, der zu Ende ist.

## Was MSE2 tut

**Es handelt nicht.** Gesucht nach `trade`, `offer` und `exchange` in
`/Users/david/mse2-quelle/src` (Fassung f9d9356 vom 2021-07-06); die Marke
ist **new**. Die Zustandsmatrix dieses Programms steht seit Phase 5.6
(`internal/domain/handelsmatrix.go`).

## Was gilt

1. **Die Lage wird abgeleitet und nirgends gespeichert.** Vier Lagen über
   dreizehn Zuständen: *dran*, wenn ein Schritt des üblichen Weges meiner
   ist — nach der Matrix, der Rolle und den schon abgegebenen Stimmen —,
   sonst *wartet*; *abgeschlossen* ist allein „beidseitig verbucht", die drei
   anderen Enden sind *abgebrochen*. **Eine Klärung ist immer dran.** Der
   nächste Schritt ist der **weiteste** des üblichen Weges: wer ein Angebot
   bekommt, soll „zustimmen" lesen, nicht „Gegenangebot". Die dreizehn
   Zustände bleiben darunter stehen (TRADE-02.AK-01).
2. **Ein Entwurf ist ein Vorgang, der nie hinausging** — Revision 1 im
   Zustand `entwurf`, ohne zweite Revision. Nur er wird ersetzt und gelöscht,
   und die Bedingung steht **in der Anweisung** der Ablage, nicht nur im
   Dienst: zwei Prozesse schreiben, und ein Löschen nach einem Lesen ließe
   ein Angebot verschwinden, das inzwischen hinausging.
3. **Was hinausging, wird kopiert oder archiviert.** Eine Kopie ist ein
   neuer Entwurf mit mir als Initiator, ohne Stimmen und ohne Vorgeschichte —
   sie wiederholt keine alte Zustimmung (TRADE-03.AK-04). Ins Archiv geht nur
   ein Handel, der zu Ende ist; **gelöscht wird dabei nichts**, denn ein
   abgeschlossener Handel ist ein Beleg (TRADE-03.AK-03). Das Archiv ist eine
   eigene Tafel (`handelsarchiv`, Wanderung 0093) und keine Spalte am
   unterschriebenen Vertrag.
4. **Unbestätigt heißt einseitig verbucht** — `lokalVerbucht` und
   `abgleichOffen`. Die Liste zeigt es an der Zeile, dazu die letzte
   Änderung.
5. **Die Namen der Karten kommen über die Scryfall-Kennung**, nicht über
   `DruckID`: die eines empfangenen Angebots ist eine Zeile im Bestand des
   Gegenübers und träfe hier einen anderen Druck.
6. **MCP liest nur** — `lesen.handelsuebersicht`. Entwürfe bearbeiten,
   kopieren, löschen und archivieren gibt es dort nicht, aus demselben Grund,
   aus dem schon das Eröffnen fehlt
   ([0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)).

## Verworfen

- **Die Lage speichern.** Sie hinge an jeder Stimme und jedem Wechsel und
  wäre die zweite Fassung dessen, was die Matrix schon sagt.
- **Einen angebotenen Handel bearbeitbar machen.** Eine Änderung nach dem
  Anbieten ist eine neue Revision (Vorschlag) — bearbeiten hieße, eine
  unterschriebene Revision zu überschreiben.
- **Abgeschlossene Handel löschen.** Das Journal und die Belege hängen daran;
  ein Archiv reicht, um die Liste frei zu halten.

## Was offen bleibt

- **Über Prozesse hinweg lädt die Liste nach dem Bestandsstand nach** —
  derselbe Zähler, mit dem die Anwendung alle zwei Sekunden fragt, ob ein
  anderer Prozess geschrieben hat (#156). Einen **Handel des Gegenübers**,
  der erst noch über die Leitung kommen muss, zeigt sie erst, wenn er da ist;
  wie alt das Gezeigte ist, sagt die letzte Änderung an der Zeile.
- **Die Ordnung nach Gegenüber folgt der Sprache des Fensters**; die
  Flächen ohne Fenster ordnen deutsch (0146).
