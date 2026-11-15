# 0044 — Was ein Werkzeug schreiben darf: nichts, was es überschreibt

**Stand:** angenommen
**Datum:** 2026-08-23
**Schreibt fort:** [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md),
und über sie [0010](0010-mcp-als-eigener-prozess.md)
**Berührt:** [0043](0043-neun-phasen-bis-0-2-0.md) — Phase T verlangt Parität
der Flächen, und diese Frage steht ihr im Weg

## Lage

Phase T (#886)
verlangt, dass **dieselbe Fähigkeit an jeder Fläche** dieselbe Antwort gibt.
Der Fähigkeitskatalog zählt es: **37** Fähigkeiten, **9** über MCP, **6** über
die Kommandozeile (Stand 2026-08-23).

Zu den fehlenden gehören die drei, die den Arbeitsablauf schließen: **ein Set
einfahren**, **ein Set sichern**, **eine Karte zeichnen**. Die Kommandozeile
kann sie seit heute.

**Über MCP dürfen sie es nach 0040 nicht.** Der Wächter sagt es wörtlich:

> Kein Werkzeug überschreibt ein Set, und keines ändert mehrere Karten auf
> einmal.

Ein `einfahren` legt Hunderte Karten auf einmal an. Ein `sichern` schreibt ein
ganzes Set. Beides fällt unter den Satz.

## Die Frage, die 0040 nicht gestellt hat

**0040 spricht vom Bestand.** Alle fünf Werkzeuge, die es erlaubt, ändern
Zeilen in `bestand.db`, und der Maßstab — „eine Karte je Aufruf" — ist auf
diesen Gegenstand gemünzt.

`sichern` und `zeichnen` ändern **den Bestand nicht**. Sie schreiben eine
Datei. Der Maßstab passt nicht, und ihn gedehnt anzuwenden hieße, eine Regel
auf einen Fall auszudehnen, für den sie nicht gemacht ist — das ist die
Ausrede, vor der §10 warnt.

## Entscheidung

**Der Maßstab ist nicht die Menge, sondern die Umkehrbarkeit.**

> **Ein Werkzeug schreibt nur, was es nicht überschreibt.**

Daraus folgt für jede der drei Fähigkeiten etwas anderes:

| Fähigkeit | Was sie anfasst | Warum sie erlaubt ist |
| --- | --- | --- |
| **einfahren** | legt ein **neues** Set an | Sie ändert nichts Bestehendes. Und die Regel „eine Quelle, ein Set" (#886) macht den zweiten Ruf zu einem Lesevorgang: er gibt das Set zurück, das schon da ist |
| **sichern** | schreibt eine Datei | **Nur an einen Ort, an dem nichts liegt.** Ein vorhandenes Ziel wird abgewiesen, wie bei `storage.Sichern` |
| **zeichnen** | schreibt eine Datei | dasselbe |

**Die andere Hälfte von 0040 bleibt unangetastet:** kein Werkzeug ändert
mehrere **vorhandene** Karten auf einmal. Der `Massenschreiber` bleibt außen
vor, und `schreiben.set.ueberschreiben` gibt es nicht.

## Warum die Kommandozeile es anders darf

`mse zeichnen` **überschreibt** ein vorhandenes Ziel, `mse sichern` auch. Das
ist kein Widerspruch, sondern der Kern von 0010:

> Eine Rücknahme, die nicht sagt, wessen Änderung sie umkehrt, ist im
> Mehr-Schreiber-Betrieb eine Vermutung.

**Wer eine Kommandozeile tippt, sieht seinen Aufruf.** Er nennt das Ziel, er
drückt Eingabe, und er kann `ls` davor setzen. Ein Sprachmodell tut beides
nicht: es ruft, während niemand hinsieht, und es wiederholt nach einer
Zeitüberschreitung.

**Der Unterschied ist die Aufsicht, nicht die Fähigkeit.** Deshalb bekommt
dieselbe Fähigkeit an zwei Flächen dieselbe Wirkung und **eine andere Wache** —
und der Fähigkeitskatalog führt sie trotzdem als eine.

## Was das kostet

**Ein Sprachmodell kann nicht zweimal an denselben Ort zeichnen.** Wer eine
Karte neu rastern will, muss einen anderen Namen nennen oder die Datei
wegräumen — und Letzteres kann es über MCP nicht.

Das ist unbequem und es ist die richtige Seite des Irrtums: eine Datei, die ein
Modell versehentlich überschreibt, ist fort; eine Datei, die es nicht schreiben
kann, kostet einen zweiten Aufruf.

## Verworfen

**Einen Schalter `ueberschreiben: true` anbieten.** Er verlagert die
Entscheidung auf das Modell — und ein Modell, das nach einer
Zeitüberschreitung wiederholt, setzt ihn beim zweiten Mal genauso wie beim
ersten. Ein Schalter, den der Gefährdete selbst umlegt, ist keine Wache.

**Den Umfang gar nicht erweitern und MCP hinter der Kommandozeile lassen.**
Das war der Stand, und es ist genau die Drift, für die es Phase T gibt: die
Flächen laufen auseinander, ohne dass es jemand merkt. Der Katalog zählt es
jetzt — und eine Zahl, die man sieht und nicht ändert, ist eine Entscheidung,
die niemand aufgeschrieben hat.

**Nur `zeichnen` erlauben und `sichern` weglassen**, weil ein Set „größer" ist.
Das wäre wieder die Menge als Maßstab, und sie trägt hier nicht: beide
schreiben genau eine Datei an einen Ort, an dem nichts liegt.

---

## Nachtrag vom 2026-08-31 — die Zusage wird jetzt eingelöst, nicht nur geprüft

Die Wache hieß `freierOrt` und war ein `os.Stat`. Ihr eigener Kommentar räumte
die Lücke ein: „was die Wache verhindert, ist das **Wahrscheinliche** — ein
Modell, das nach einer Zeitüberschreitung denselben Aufruf wiederholt."

**Der Vermerk hier sagte das nicht**, und die Werkzeugbeschreibung auch nicht:
oben steht „an einen Ort, an dem nichts liegt", ohne Einschränkung. Der
Zweitblick über Phase 4.4 hat den Abstand zwischen beidem angestrichen
(#1343).
Zwischen Nachsehen und Schreiben liegt ein Zeitraum; zwei gleichzeitig
arbeitende Aufrufer bekamen beide **frei** zur Antwort, und der spätere
ersetzte den früheren.

Seitdem wird das Ziel **reserviert** statt befragt, und der Kern entscheidet
es:

| Was entsteht | Wie reserviert | Was der Verlierer bekommt |
| --- | --- | --- |
| eine Datei — `zeichnen`, `bogen`, `mtgjson-schreiben`, `vorlagen ausfuehren` | `os.OpenFile` mit `O_CREATE\|O_EXCL` | `EEXIST` → `ErrConflict` |
| ein Verzeichnis — `sichern` | `os.Mkdir` | `EEXIST` → `ErrConflict` |

**Zwei Reservierungen, nicht eine**, und der Grund ist gemessen: eine leere
**Datei** am Ziel von `sichern` ließ `service.keinArchiv` mit „ein Archiv wird
nicht zurückgeschrieben" abweisen — die Reservierung verhinderte das Schreiben,
das sie schützen sollte. Gefunden hat es der Ablauf über alle Flächen; die
Testläufe der beiden Kommandos für sich waren grün, weil sie nur Fehlerwege
prüften.

**Und was scheitert, räumt seine Reservierung wieder ab.** Eine Null-Byte-Datei
sähe aus wie ein geglücktes Schreiben — §7 Regel 3 andersherum: was nicht
geschrieben ist, soll auch nicht dastehen.
