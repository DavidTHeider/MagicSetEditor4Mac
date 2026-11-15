# 0010 — Der MCP-Server ist ein eigener Prozess über stdio

**Stand:** angenommen
**Datum:** 2026-08-15
**Fortgeschrieben durch:** [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md)
und [0044](0044-was-ein-werkzeug-schreiben-darf.md) — beide am Umfang der
Werkzeuge; [0045](0045-zwei-programme-an-einem-bestand.md) — an der Zahl der
Prozesse

## Lage

#143 stellt drei
Fragen und verlangt für die dritte einen Vermerk: **wo läuft der Server?**

> In der Anwendung (dann nur, solange sie offen ist) oder als eigener Prozess auf
> derselben Datenbank (dann gleichzeitig, mit allem, was SQLite dazu sagt)?

## Entscheidung

**Ein eigener Prozess, gesprochen über stdio, auf derselben SQLite-Datei.**

Das MCP-Gegenüber — ein Sprachmodell oder ein Werkzeug — startet
`mse-mcp` als Unterprozess und redet über Standardein- und -ausgabe. Es gibt
keinen Port, keine Adresse und keinen Dienst, der auf etwas wartet.

## Verworfen: in der Anwendung

Die naheliegende Möglichkeit, und sie scheitert an zwei Dingen.

**Sie machte die Anwendung zu einem Netzwerkdienst.** Ein Server im Fenster kann
nicht über stdio sprechen — die Standardeingabe des Fensters gehört dem Fenster.
Er bräuchte einen Port auf `localhost`, und damit stünde in `docs/ROADMAP.md`
unter „Was ausdrücklich nicht kommt" eine Zeile, die nicht mehr stimmt:

> Kein Netzwerkdienst, keine Anmeldung, keine Synchronisierung.

Diese Zeile ist keine Formsache. Ein lauschender Port ist eine Angriffsfläche,
die es zu verteidigen gilt — Herkunftsprüfung, Zugangsmarken, Bindung an die
Rückschleife —, und all das für einen Nutzen, den stdio ohne eine Zeile davon
liefert.

**Und sie liefe nur, solange jemand das Fenster offen hat.** Ein Werkzeug, das
sein Gegenüber nicht selbst starten kann, ist eines, das man vergisst
einzuschalten.

## Verworfen: ein eigener Prozess mit eigenem Bestand

Ein zweiter Prozess mit einer eigenen Kopie der Daten und einem Abgleich
dazwischen. Das ist eine Synchronisierung, und dieselbe Zeile der Roadmap sagt
nein — zu Recht: zwei Bestände, die auseinanderlaufen können, sind schlimmer als
einer, um den zwei Prozesse sich streiten müssen.

## Was das kostet

**Zwei Schreiber auf einer SQLite-Datei.** Bisher schrieb genau ein Fenster; ab
jetzt können Fenster und Server gleichzeitig schreiben. Das ist keine
theoretische Lage, sondern die normale:

> **Und seit dem 2026-08-23 sind es nicht mehr zwei.** Mit `mse`
> (#886) gibt
> es einen **dritten** Prozess an derselben Datei, mit Phase U kommt ein
> vierter. An dieser Entscheidung ändert das nichts — sie war nie eine über
> die Zahl zwei, sondern eine darüber, dass **ein** Bestand geteilt wird statt
> mehrerer abgeglichen. Was dabei wirklich passiert, ist inzwischen gemessen
> und steht in [0045](0045-zwei-programme-an-einem-bestand.md). Der letzte
> Absatz hier — „das Fenster erfährt nichts" — ist seit
> [0017](0017-die-prozesse-bemerken-einander.md) **eingelöst** und nicht
> etwa gewachsen: das Fenster fragt `data_version` und lädt nach.

- **Eine Transaktion beginnt sofort schreibend** (`_txlock=immediate`). Ohne das
  beginnt sie aufschiebend und will beim ersten Schreiben auf die Schreibschranke
  wechseln — im WAL-Modus geht das nicht mehr, wenn inzwischen jemand anders
  geschrieben hat, und SQLite meldet `SQLITE_BUSY_SNAPSHOT`. Dagegen hilft
  `busy_timeout` **nicht**.

  Gefunden hat das der Nachlauf zum Zweitblick: der Zwei-Schreiber-Test lief
  gegen eine Datenbank im Arbeitsspeicher mit **einer** Verbindung und prüfte
  die Zusage damit nicht. Auf einer Datei mit zwei Verbindungen scheiterte er —
  und mit `-race` blieb er grün, weil der Wettlaufprüfer die Läufe so
  verlangsamt, dass sie einander nicht treffen.
- SQLite läuft hier im **WAL-Modus** mit `busy_timeout = 5000`
  ([`internal/storage/datenbank.go`](../../internal/storage/datenbank.go)). WAL
  erlaubt beliebig viele Leser neben **einem** Schreiber; ein zweiter Schreiber
  wartet bis zu fünf Sekunden und bekommt danach `SQLITE_BUSY`.

  **Diese Zusage stimmte nicht, als sie hier zuerst stand.** Die Pragmas wurden
  über `db.ExecContext` gesetzt und trafen damit **eine** Verbindung aus dem
  Vorrat; jede weitere, die `database/sql` später öffnete, kam ohne sie — ohne
  Wartezeit und **ohne Fremdschlüssel**. Gesehen hat das erst der Test mit zwei
  Schreibern auf einer Datei: 17 von 20 Schreibvorgängen scheiterten. Seither
  stehen sie in der Datenquelle und gelten je Verbindung. Das ist der Grund,
  warum #143 „geprüft, nicht behauptet" verlangt.
- Damit sind die vier Regeln aus `AGENTS.md` §7 nicht mehr eine Frage des
  Frontends. „Ein Vorgang je Datensatz" gilt für **zwei Aufrufer**, und keiner
  von beiden sieht den anderen.
- **Das Fenster erfährt nichts von einer Änderung des Servers.** Wer im Modell
  ein Feld setzt, während die Karte offen ist, sieht es dort nicht — bis jemand
  neu lädt. Das ist der Preis, und er ist heute nicht bezahlt: eine
  Benachrichtigung zwischen den Prozessen gibt es nicht.

**Was dagegen steht, ist gebaut und geprüft:** der Server öffnet die Datenbank
mit denselben Pragmas, und ein Test fährt zwei Schreiber gleichzeitig gegen
**eine** Datei — unter `-race`, mit echten Verbindungen, nicht mit Doubles.

## Was ein Modell hier tun darf

Frage 1 aus #143, und sie gehört hierher, weil sie dieselbe Wurzel hat.

**Lesen ist frei. Schreiben ist ein einzelnes Feld.** Es gibt kein Werkzeug, das
ein Set überschreibt, keines, das eine Karte löscht, und keines, das mehrere
Karten auf einmal ändert — obwohl der Bestand das könnte
(`Massenschreiber`).

> **Dieser Absatz gilt nicht mehr, und das ist der Punkt.** Er stand hier als
> Antwort auf Frage 1 aus #143 und ist zweimal fortgeschrieben worden:
> [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md) hat den
> Lebenszyklus einer Karte hinzugenommen — Anlegen, Kopieren, **Löschen**,
> Tauschen —, [0044](0044-was-ein-werkzeug-schreiben-darf.md) das Einfahren,
> Sichern und Zeichnen.
>
> **Der Maßstab ist ein anderer geworden, nicht bloß weiter.** Hier war er die
> **Menge** — ein Feld ja, ein Set nein. In 0044 ist er die
> **Umkehrbarkeit**: ein Werkzeug darf schreiben, was es nicht überschreibt.
> Danach ist `schreiben.set.einfahren` erlaubt (es legt ein neues Set an) und
> ein Ausfahren, das eine vorhandene Datei ersetzte, nicht.
>
> Was heute wirklich dasteht, führt
> [`docs/FAEHIGKEITEN.md`](../FAEHIGKEITEN.md), und ein Wächter hält die
> Liste gegen den Quelltext. Der Satz oben bleibt stehen, weil er den Stand
> beschreibt, von dem aus entschieden wurde. Der Grund steht in der Frage selbst: „ein Werkzeug, das ein
ganzes Set überschreibt, gehört nicht in dieselbe Reihe wie eines, das ein Feld
setzt."

**Und die Reihen sind getrennt, nicht nur beschriftet.** Jedes Werkzeug trägt
seine Art im Namen (`lesen.` beziehungsweise `schreiben.`), im Text und in den
Angaben, die MCP dafür vorsieht (`readOnlyHint`, `destructiveHint`). Ein
Gegenüber, das nur lesen lassen will, erkennt das an drei Stellen.

**`destructiveHint` steht am schreibenden Werkzeug auf `true`.** Der erste
Anlauf sagte `false` und begründete es damit, dass nichts ein Set löscht — das
ist wahr und beantwortet die Frage nicht: `schreiben.feld` überschreibt einen
Wert und kann ein Feld wegnehmen. Ein Gegenüber, das vor zerstörenden Werkzeugen
nachfragt, soll hier nachfragen. Der Zweitblick über Phase 9 hat es angestoßen.

**Rücknahme statt Bestätigung.** #143 verlangt eines von beiden und nennt es
ausdrücklich eine Entscheidung. Eine Bestätigung können wir nicht zusagen: ob
ein Gegenüber nachfragt, entscheidet es selbst, und ein Server, der auf eine
Zusage baut, die er nicht durchsetzt, hat keine. Eine **Rücknahme** ist etwas,
das dieser Prozess herstellen kann — jede Änderung gibt den vorigen Wert zurück,
und wer ihn hat, kann sie umkehren.

**Und sie ist ganz oder gar nicht.** Der erste Anlauf las den Vorwert und schrieb
danach, in einem getrennten Aufruf; im ausdrücklich normalen Zwei-Schreiber-
Betrieb konnte das Fenster dazwischen denselben Wert ändern, und die Rücknahme
führte auf einen Stand, den es nie gab. Und ein Feld, das es vorher **nicht**
gab, ließ sich gar nicht wiederherstellen: `FeldSetzen` konnte es nur leeren,
und eine leere Zeile ist nicht dasselbe wie keine Zeile.

Beides ist behoben: `FeldSetzenMitVorwert` liest und schreibt in **einer**
Transaktion, und `entfernen` nimmt ein Feld weg. Der Zweitblick über Phase 9 hat
beides gefunden — und dass er es fand, ist der Grund, warum #143
„geprüft, nicht behauptet" verlangt.

**Und sie ist immer noch nicht ganz.** Der Nachlauf hat nachgesetzt: ein wieder
eingesetztes Feld landet **hinten**, weil `card_values` eine Reihenfolge führt
und ein neuer Wert `max + 1` bekommt. Für ein Set aus einer `.mse-set` heißt
das, dass die Zeile wandert. Das steht als
#160 und
gehört zu Phase 6, wo die Historie gebaut wird.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn `SQLITE_BUSY` im Betrieb auftritt, statt eine Möglichkeit zu bleiben.
  Dann ist die Zeitgrenze zu knapp oder die Schreiblast zu hoch, und der Weg
  führt über eine Warteschlange im Fenster — nicht über einen zweiten Bestand.

  **Nachgemessen am 2026-08-23**, Fenster gegen Kommandozeile: ein Schreiben
  neben einem offenen Fenster gelingt in 15 ms, gegen eine gehaltene Sperre
  nach 1,6 s. Erst wer **länger als fünf Sekunden** sperrt, wird abgewiesen.
  Die Zeitgrenze ist damit nicht zu knapp — und seit demselben Tag heißt die
  Abweisung `domain.ErrBelegt` und sagt, was zu tun ist, statt eine
  Treibermeldung zu zeigen.
- Wenn jemand die Anwendung offen hat und die Anzeige regelmäßig veraltet. Dann
  ist die fehlende Benachrichtigung zwischen den Prozessen kein Preis mehr,
  sondern ein Fehler, und sie gehört gebaut.
- Wenn ein Gegenüber die getrennten Reihen ignoriert und schreibende Werkzeuge
  ungefragt aufruft. Dann ist die Rücknahme das Einzige, was bleibt — und sie
  muss dann auch bedienbar sein, nicht nur vorhanden.
