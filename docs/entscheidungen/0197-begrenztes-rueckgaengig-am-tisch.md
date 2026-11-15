# 0197 — Am Tisch nimmt ein Sitz seinen letzten Zug zurück, wenn er nichts verrät und nichts entschieden hat

**Stand:** angenommen am 2026-10-18 · #2376 (P8, Befund C-07). Umgesetzt in `internal/service/zuruecknahme.go`; Lauf über alle Flächen in `internal/flaechen`.

## Lage

Am Tisch gab es kein Rückgängig: wer sich vertappte oder das Leben um
eins zu viel senkte, stellte den Stand mit einem zweiten Befehl von Hand
zurück — und am gemeinsamen Gerät sah der andere Sitz nicht, dass es eine
Berichtigung war. ⌘Z stand im Menü und tat am Tisch nichts.

**Die Comprehensive Rules geben dafür keine Erlaubnis.** CR 733.1
(Fassung 2026-09-25) beschreibt, wie eine **regelwidrige** Handlung
zurückgedreht wird; ein regelgerechter Zug, den man sich anders überlegt,
kommt dort nicht vor. Was hier entschieden wird, ist eine **Tischabsprache**
dieses Programms, keine Regel des Spiels — und es behauptet keine.

## Entscheidung

1. **Ein Befehl, `spieler.zuruecknehmen`, ohne Nutzlast.** Er meint den
   **letzten** Befehl der Partie und nichts davor — kein Verlauf, kein
   Wiederherstellen.
2. **Nur der eigene.** Der letzte Befehl muss vom selben Sitz stammen. Hat
   seither ein anderer Sitz gehandelt, wird abgewiesen: sein Zug baute auf
   dem Stand auf, den die Rücknahme änderte.
3. **Nur, was auf der Liste steht** — Befehle, deren Wirkung vollständig im
   öffentlichen Ereignis steht und sich ohne Rest umkehren lässt:

   | Befehl | Umkehr |
   | --- | --- |
   | `objekt.tappen` | `objekt.enttappen` am selben Objekt |
   | `objekt.enttappen` | `objekt.tappen` am selben Objekt |
   | `spieler.leben` | `spieler.leben` mit dem Gegenwert |
   | `spieler.vergiftung` | `spieler.vergiftung` mit dem Gegenwert |

4. **Nicht auf der Liste, und mit Grund abgewiesen:**
   - **Zufall** — Münzwurf, Würfeln, Mischen, Starthand, Mulligan, Ziehen:
     ein zweiter Versuch wäre ein neuer Wurf.
   - **Aufgedecktes** — Umdrehen, Spielen, Wirken, Offenlegen, jeder
     Zonenwechsel: was gesehen wurde, bleibt gesehen.
   - **Abschluss** — Passen, Abkürzen, Zugende, Behalten, Aufgeben, eine
     erklärte Kampfform: sie geben das Wort weiter oder beenden etwas.
   - **Eine Rücknahme selbst** — kein Wiederherstellen über die Hintertür.
   - **Ein Befehl mit mehr als einem Ereignis** — etwa ein Leben, das die
     Partie beendet (CR 704.5a): mehr als eine Folge kehrt die Liste nicht
     um.
   - **Ein Zähler** (`objekt.zaehler`): sein Ereignis nennt den neuen Stand,
     nicht den Unterschied, und unter null klemmt er — die Umkehr ließe sich
     nicht sicher rechnen.
5. **Der Wirt entscheidet**, an jeder Fläche gleich: die Rücknahme ist ein
   Befehl wie jeder andere, sie geht durch denselben Kern und wird dort
   angenommen oder abgewiesen — vom Fenster, von `mse partie-befehl`, über
   MCP und über `POST /partien/{id}/befehle`, und von einem Gast über den
   Wirt.
6. **Angehängt, nicht gelöscht.** Der Strom bleibt append-only: die
   Rücknahme schreibt ein Ereignis `spieler.zuruecknehmen` mit der Folge,
   die sie zurücknimmt, und danach das Ereignis der Umkehr. Die Fassung
   steigt um eins.
7. **Eine abgewiesene Rücknahme ändert nichts** — weder Strom noch Fassung.
   **Dieselbe Kennung zweimal** ist derselbe Wunsch und bekommt dieselbe
   Antwort; eine neue Kennung nach einer Rücknahme trifft auf Punkt 4.
8. **Die Umkehr geht durch dieselbe Auslegung** wie ein Befehl von Hand:
   nimmt der Kern sie dort nicht an, wird die Rücknahme mit seinem Grund
   abgewiesen.

## Verworfen

- **Ein Verlauf mit mehreren Schritten**: er müsste fremde Züge dazwischen
  überspringen oder mit zurücknehmen — beides ändert, worauf ein anderer
  Sitz gebaut hat.
- **Den vorigen Stand speichern und zurückschreiben**: das wäre eine
  Schemaänderung (ein eigener Zug nach §5) und löschte das Wissen, dass etwas
  zurückgenommen wurde.
- **CR 733.1 als Grundlage**: sie gilt für Regelverstöße, und eine
  Rücknahme regelgerechter Züge darauf zu stützen, wäre ein Beleg, der
  keiner ist (§10).

## Rückweg

Ein Befehl mehr auf der Liste ist eine neue Fassung dieses Vermerks und
ein Fall mehr in `zuruecknahme_test.go`. Ohne die Rücknahme fällt nur der
Befehl weg; der Strom bleibt lesbar, denn ihre Ereignisse sind gewöhnliche
Ereignisse.
