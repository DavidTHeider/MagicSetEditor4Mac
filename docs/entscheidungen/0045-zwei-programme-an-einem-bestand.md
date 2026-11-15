# 0045 — Zwei Programme an einem Bestand

**Stand:** angenommen · **Datum:** 2026-08-23 ·
**Anlass:** #886,
Werkstück 2 — „der gleichzeitige Zugriff auf den Bestand ist bedacht"

## Worum es geht

Bis zum 2026-08-23 gab es **ein** Programm, das den Bestand öffnete: das
Fenster. Mit der Kommandozeile gibt es ein zweites, mit dem MCP-Server ein
drittes, und mit Phase U kommt ein viertes dazu. Sie greifen auf **dieselbe
Datei** zu, und keines weiß vom anderen.

Die Frage ist nicht, ob das vorkommt. Es ist der Normalfall: wer eine
Kommandozeile baut, damit sie neben dem Fenster läuft, hat sie gebaut, damit
sie neben dem Fenster läuft.

## Die Entscheidung

> **SQLite regelt es, und wir sagen, was es dabei tut.**

Kein eigener Sperrmechanismus, keine Lockdatei, kein Vorrang für das Fenster.
Was schon dasteht, trägt: WAL, `busy_timeout(5000)`, `_txlock=immediate`.

## Was dabei wirklich passiert

**Gemessen am 2026-08-23**, ein offenes Fenster (eine `storage.Oeffnen`-Verbindung,
wie `app.go` sie hält) gegen die gebaute `mse`-Kommandozeile, beide auf
demselben Bestand:

| | Gemessen |
| --- | --- |
| CLI schreibt, Fenster offen | gelingt nach **15 ms**, kein Fehler |
| Sieht das offene Fenster es? | **ja** — die nächste Abfrage liest den neuen Wert |
| CLI gegen eine gehaltene Schreibsperre | wartet **1.580 ms**, dann gelingt sie |
| Reihenfolge | der Zweite liest den Stand des Ersten — „Vorher: vom Fenster" |
| Sperre länger als `busy_timeout` | scheitert nach **5,1 s** |

**Die letzte Zeile war der Befund.** Sie lautete

    Migration beginnen: database is locked (5) (SQLITE_BUSY)

und daran ist alles falsch, was falsch sein kann: eine rohe SQLite-Meldung an
einer Oberfläche, die §8 nirgends zulässt; ein Zusammenhang, der auf eine
kaputte Migration zeigt, obwohl keine kaputt ist; und kein Wort darüber, dass
der Aufrufer **nichts** reparieren muss, sondern warten.

Sie lautet jetzt

    Der Bestand ist gerade belegt — vermutlich hat das Fenster ihn offen
    oder ein zweiter Befehl läuft. Bitte gleich noch einmal versuchen.

Der Weg dahin ist [`domain.ErrBelegt`](../../internal/domain/fehler.go): die
Abbildung in `storage/fehler.go` erkennt `SQLITE_BUSY` am **Basiscode** (so
gelten `_SNAPSHOT`, `_RECOVERY` und `_TIMEOUT` mit), `internal/app` übersetzt
ihn in `CodeBelegt`, beide Sprachkataloge kennen ihn, und die Kommandozeile
hängt den Rat an.

**Und der Anlass hat eine Klasse.** Nicht die eine Stelle wurde geheilt:
sechzehn Umschließungen in `internal/storage` gaben ihren SQLite-Fehler roh
weiter, und vierzehn Meldungen der Kommandozeile schrieben die ganze Kette mit
`%v`. Beides läuft jetzt durch je **eine** Stelle.

## Was das kostet

**Die letzte Schreibung gewinnt.** Wer im Fenster einen Namen ändert und
gleichzeitig `mse feld … name …` ruft, behält den zweiten Wert. Es gibt keine
Zusammenführung und keine Rückfrage — SQLite reiht, es vergleicht nicht.

**Das Fenster merkt es — und dieser Absatz hat zuerst das Gegenteil behauptet.**

Hier stand, die Oberfläche bemerke eine fremde Änderung nicht, und dazu ein
Issue, das die Abhilfe verlangte. Beides war falsch: sie ist seit
#156 gebaut
und in [0017](0017-die-prozesse-bemerken-einander.md) begründet. Das Fenster
fragt alle zwei Sekunden `PRAGMA data_version` und lädt nach, wenn die Zahl
gestiegen ist (`frontend/src/haken/useBestandsstand.ts`,
`frontend/src/App.tsx`).

**Woran der Irrtum lag, ist der Punkt.** Gemessen war die *Verbindung* des
Fensters; behauptet war etwas über die *Oberfläche*. Das ist genau der Fehler,
den §10 „die Reichweite einer Messung ist die Reichweite ihrer Aussage" nennt
— und er ist hier im Vermerk über das Messen passiert.

**Für den dritten Prozess ändert sich daran nichts.** `data_version`
unterscheidet nach **Verbindung**, nicht nach Prozess; ob die fremde Schreibung
aus `mse-mcp` oder aus `mse` kommt, sieht die Zahl gleich.

Was 0017 als ungeprüft auswies — „dass die Abfrage im laufenden Fenster oft
genug kommt" —, ist am 2026-08-23 nachgeholt worden: die Kommandozeile setzte
einen Kartennamen, und die offene Kartenseite zeigte ihn **ohne jede
Berührung**. Was diese Messung nicht sagt, ist *wie schnell* — der Grund steht
in [0017](0017-die-prozesse-bemerken-einander.md).

**Fünf Sekunden Geduld und dann ein Fehler.** `busy_timeout` ist keine
Warteschlange; wer länger sperrt, wird abgewiesen. Für ein Skript heißt das:
den Aufruf wiederholen. Der Satz sagt es.

## Was ausdrücklich anders ist

**Der Alleinzugriff bleibt die Ausnahme, die er ist.**
[`storage.AlleinAmBestand`](../../internal/storage/alleinzugriff.go) fordert die
Datei exklusiv und weist ab, wenn jemand anders sie hält. Das gilt für das
Zurückspielen einer Sicherung — dort wird die Datei **ersetzt**, und ein
zweiter Leser läse danach in ein Nichts. Für jeden gewöhnlichen Schreibweg
wäre dieselbe Forderung falsch: sie machte das zweite Programm unbenutzbar,
solange das erste läuft.

## Verworfen

**Ein Vorrang für das Fenster.** Klingt richtig — ein Mensch wartet ungern,
ein Skript schon. Es scheitert daran, dass keines der Programme vom anderen
weiß: Vorrang setzt eine gemeinsame Instanz voraus, und die gäbe es nur als
eigenen Dienst. Das ist [0010](0010-mcp-als-eigener-prozess.md) noch einmal,
und die Antwort dort gilt hier weiter.

**Eine Lockdatei neben dem Bestand.** Sie sagt dasselbe wie SQLites eigene
Sperre, nur schlechter: sie überlebt einen Absturz und muss dann von Hand
weggeräumt werden.
