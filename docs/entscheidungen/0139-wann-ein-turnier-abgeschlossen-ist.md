# 0139 — Wann ein Turnier abgeschlossen ist

**Stand:** entschieden am 2026-09-27 · DOM-04.AK-04
(#1699)

## Die Frage

AK-04 verlangt für **laufende und abgeschlossene** Turniere *„definierte
Änderungs- und Löschregeln, Protokollierung und Rechteprüfung über die vier
Flächen"*. Bis hierher kannte ein Turnier keinen Abschluss: Vermerk 0062,
Entscheidung 10, hielt fest, dass es **kein Zustandsfeld** gibt, weil ein
`läuft`-Flag bei einem Absturz falsch stehen bliebe.

## Die Entscheidung

**Ein Turnier ist abgeschlossen, wenn ein Mensch es abschließt** — und nicht,
wenn ein Zustand im Arbeitsspeicher das meint. Das ist kein Widerspruch zu
0062/10: dort ging es um ein Flag, das ein laufender Vorgang setzt und ein
Absturz falsch hinterlässt. Hier ist es eine **Handlung mit Datum**, die im
Bestand steht, bevor sie gilt, und die niemand „mitten drin" unterbricht.

| | Laufend | Abgeschlossen |
| --- | --- | --- |
| Anmelden, Aus- und Einsteigen | ja | **nein** |
| Runde legen | ja | **nein** |
| Ergebnis eintragen, berichtigen, zurücknehmen (auch aus einer Partie) | ja | **nein** |
| Nach einer Vorlage kopieren | ja | ja — die Kopie ist **offen** |
| Wiedereröffnen | — | ja, mit Rückfrage; steht im Protokoll |
| Löschen | mit Rückfrage | mit Rückfrage — **das Protokoll geht mit** |

**Abschließen setzt voraus, dass es nichts Offenes gibt**: mindestens eine
Runde, und jede Paarung hat ein Ergebnis. Ein Abschluss mit offenen Tischen
schlösse einen Stand ab, den es nie gab.

**Die Regel steht zweimal**: im Dienst (`Turnierleitung`, `Matchleitung`)
als fachliches Nein mit eigenem Schlüssel, und im Speicher als zweite Reihe —
das Fortschreiben der Fassung trifft kein abgeschlossenes Turnier. Ein Weg,
der am Dienst vorbeiführte, schriebe trotzdem nicht.

### Das Protokoll

**Jede Änderung schreibt eine Zeile — im selben Zug wie die Änderung.**
Angelegt, angemeldet, aus- und eingestiegen, Runde gelegt, Ergebnis
eingetragen oder zurückgenommen, abgeschlossen, wiedereröffnet. Die Stelle
dafür ist das Fortschreiben der Fassung, durch das jede Änderung schon geht;
eine zweite Stelle daneben liefe auseinander.

**Es ist ein Verlauf und keine Aufsicht.** Es sagt, **was** wann geschah, und
nicht, wer es tat: ein Turnier gehört einem Profil, und wer in diesem Profil
arbeitet, weiß der Anmeldebund, nicht das Turnier.

**Es geht mit dem Turnier.** Wer ein Turnier löscht, löscht sein Protokoll
(`ON DELETE CASCADE`) — ein Protokoll zu einem Turnier, das es nicht mehr
gibt, gehörte keinem Profil mehr.

### Die Rechte

Nichts Neues, und das ist die Antwort: jede Turnierfähigkeit steht in
`docs/AUTORISIERUNG.md` am Träger **Profil** — lesen braucht `leser`,
schreiben `mitarbeiter`, an allen vier Flächen durch dieselbe Naht. Die drei
neuen Fähigkeiten stehen dort mit derselben Einordnung, und ein Lauf ohne
Rolle weist sie ab wie jede andere.

## Verworfen

| Verworfen | Weil |
| --- | --- |
| **Abgeschlossen, sobald die empfohlene Rundenzahl gespielt ist** | Die Rundenzahl ist ein Vorschlag (0062, Entscheidung 4); ein Turnier, das sich selbst schlösse, nähme dem Turnierleiter eine Entscheidung ab, die ihm gehört |
| **Ein abgeschlossenes Turnier lässt sich nicht löschen** | Wer ein Turnier aus Versehen angelegt und abgeschlossen hat, käme nie davon los. Die Rückfrage genügt |
| **Wer es tat, steht im Protokoll** | Das Profil ist der Träger (0066); ein Name im Protokoll hielte eine Person fest, die das Turnier nicht kennt |
