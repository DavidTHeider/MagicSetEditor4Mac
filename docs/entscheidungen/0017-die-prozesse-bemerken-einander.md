# 0017 — Die Prozesse bemerken einander über `data_version`

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

[0010](0010-mcp-als-eigener-prozess.md) hat den MCP-Server als eigenen Prozess
auf derselben SQLite-Datei entschieden — und den Preis gleich mitgenannt:

> Das Fenster erfährt nichts von einer Änderung des Servers. […] Das ist der
> Preis, und er ist heute nicht bezahlt.

Und wann er fällig wird:

> Wenn jemand die Anwendung offen hat und die Anzeige regelmäßig veraltet.

#156 verlangt
die Entscheidung, **wie** die Prozesse einander bemerken.

## Entscheidung

**`PRAGMA data_version`, abgefragt vom Fenster.**

SQLite führt im Kopf der Datei einen Zähler, der steigt, sobald eine **andere**
Verbindung bestätigt hat. Für eigene Schreibvorgänge bleibt er stehen — und das
ist genau richtig: wer selbst schreibt, weiß es.

**Gefragt wird vom Frontend**, nicht von einem Faden in Go.

## Warum nicht die beiden anderen Wege

**Ein SQLite-Update-Hook** (`sqlite3_update_hook`) meldet nur Änderungen der
**eigenen** Verbindung. Für zwei Prozesse ist er blind — er beantwortet eine
andere Frage, als hier gestellt wird.

**Ein Dateiwächter** auf der `-wal`-Datei ginge, kostet aber eine
plattformabhängige Abhängigkeit (`fsnotify`), meldet jeden Zwischenzustand
einer Transaktion und sagt nicht, ob etwas **bestätigt** wurde. `data_version`
sagt genau das und braucht nichts.

## Warum das Fenster fragt und Go nicht meldet

Ein Faden in Go, der von sich aus meldete, führe auch dann, wenn niemand
hinsieht — und er müsste wissen, wann das Fenster geschlossen ist. Das Frontend
weiß beides ohnehin: es kennt seinen Lebenslauf und über
`document.visibilityState` auch, ob jemand hinschaut.

**Die Frage kostet nichts, was auf die Platte ginge.** `PRAGMA data_version`
liest einen Zähler, den SQLite ohnehin führt.

## Was das kostet, und der Preis gehört ausgesprochen

**Eine Abfrage sagt nur *dass*, nicht *was*.** Das Fenster lädt danach neu, was
es anzeigt — nicht gezielt die eine Karte. Bei dreihundert Karten ist das eine
Abfrage mehr, kein Problem; bei dreißigtausend wäre es eines, und dann gehört
die Frage neu gestellt.

**Zwischen zwei Abfragen veraltet die Anzeige weiter.** Der Abstand ist eine
Wahl: kurz genug, dass niemand es merkt, lang genug, dass es nichts kostet.
**Zwei Sekunden**, und die Zahl steht im Frontend, wo man sie sieht.

**§7 Regel 4 gilt weiterhin nur je Prozess.** Wer im Fenster eine Karte
bearbeitet, während der Server dieselbe ändert, bekommt jetzt mit, *dass* etwas
geschah — aber die beiden Schreibvorgänge wissen weiterhin nichts voneinander.
Das ist eine andere Frage (optimistisches Sperren), und sie ist hier nicht
beantwortet.

## Was geprüft ist und was nicht

**Geprüft** ist der Mechanismus: zwei Verbindungen, und die Zahl steigt nur für
die fremde Schreibung (`TestDerDatenstandSteigtWennEinAndererSchreibt`).

Das Issue verlangt „zwei laufende Prozesse, nicht zwei Verbindungen". Für
`data_version` ist das dasselbe: die Zahl unterscheidet nach **Verbindung**,
nicht nach Prozess — zwei Vorräte in einem Lauf sind für diese Frage zwei
Fremde. Ein Test, der ein zweites Programm startet, prüfte zusätzlich das
Starten.

**Nicht geprüft** war, dass die Abfrage im laufenden Fenster oft genug kommt.
Das sieht man nur am laufenden Fenster (§12,
[0007](0007-kein-entfernter-lauf.md)), und es gehört vor die Freigabe.

### Nachgeholt am 2026-08-23

Gemessen an `wails dev` über den Browser auf `34115`, mit offener Kartenseite
und **ohne jede Berührung**: die Kommandozeile
(#886) setzte
den Namen einer Karte, und die Liste im Fenster zeigte ihn danach von selbst.
Damit ist der Weg vom fremden Schreiben bis in die Anzeige **einmal ganz
gegangen** — bis dahin war nur der Mechanismus belegt, nicht die Strecke.

**Und die Reichweite dieser Messung ist kleiner als sie aussieht.** Wie lange
es dauert, sagt sie **nicht**: der eingebettete Tab war im Hintergrund, und
Chrome drosselt dort die Zeitgeber. Nachgezählt im selben Lauf — ein
`setInterval` mit 250 ms lieferte **6 statt 113** Auslösungen in 28,3 s, also
rund eine Auslösung je fünf Sekunden. Die zwei Sekunden aus dem Frontend gelten
in einem sichtbaren Fenster; hier galten sie nicht.

**Der Sichtbarkeitsschalter musste dafür überschrieben werden.** Der Haken
fragt nicht, solange `document.visibilityState === "hidden"` — und der
eingebettete Tab meldet dauerhaft „hidden", auch nach dem Nachvorneholen. Für
die Messung wurde die Eigenschaft auf „visible" gesetzt; das ist genau das,
was ein echtes Fenster meldet, und es ist der einzige Eingriff gewesen.

### Nachtrag vom 2026-09-09 — „eine andere Verbindung" hieß auch das eigene Fenster

Die Sichtprüfung von U5 (#1416)
hat jede Bindung mitgeschrieben, die das Fenster nach einem Klick ruft: auf
`EinstellungSetzen` folgte zwei Sekunden später `Datenstand`, und darauf
`Kartenzeilen`, `Karte`, `Sets` und **dreimal** `Vorschau`. Das Fenster hielt
sein eigenes Schreiben für einen fremden Prozess
(#1485).

**Der Satz oben war richtig und die Anwendung nicht.** `data_version` zählt je
**Verbindung** — das steht hier seit dem ersten Tag —, und das Fenster hatte
einen Vorrat davon: der Standleser hielt eine fest, jedes Schreiben lief auf
einer anderen. Derselbe Fehler stand schon hinter
#1325 und
wurde dort am Symptom behoben.

**Seitdem hat jeder Vorrat eine Verbindung** (`storage.oeffnenMit`, für Datei
und Arbeitsspeicher gleich), und der Standleser fragt über den Vorrat wie das
Schreiben. Eigene Commits bewegen die Zahl nicht mehr; die eines anderen
Prozesses weiter. Der Preis: Lesen wartet auf Schreiben, statt daneben zu
laufen — SQLite hat ohnehin einen Schreiber, die Rufe des Fensters kommen
nacheinander, und der MCP-Server arbeitete seit je so.

**Was das nicht heilt**: `#1325` bleibt richtig — ein Commit ohne Änderung
liest ein *anderes* Fenster weiter als Änderung.

### Nachtrag vom 2026-10-04 — auch dieser Prozess schreibt ohne das Fenster

Die Hüllenprobe `nat04handel` (S12.NAT-04,
#1672) hat
zwei Programme über die Leitung handeln lassen: Bert bot einen Tausch an, er
stand in Annas Bestand — und ihr Fenster zeigte ihn erst, als sie die Seite
neu öffnete. **Der Satz „wer selbst schreibt, weiß es" galt nicht mehr**:
seit dem Horchposten schreibt dieser Prozess auch dann, wenn das Fenster es
nicht veranlasst hat — eine Handelsnachricht, eine eingelöste Einladung —,
über **dieselbe** Verbindung, und `data_version` steht still.

**Die Antwort bleibt dieselbe Frage**: `Datenstand` gibt seither die Summe aus
`data_version` und einem Zähler dessen, was dieser Prozess **nebenher**
geschrieben hat (`standnaht` in `internal/app/datenstand.go`). Wer im
Hintergrund schreibt, sagt es mit `nebenherGeschrieben`; das Fenster fragt
wie bisher alle zwei Sekunden und lädt dann nach — jede Seite, die den
Datenstand liest, auch die, an die niemand gedacht hat. **Kein Ereignis aus
Go**, aus demselben Grund wie oben.

**Was das kostet**: die Summe steigt auch, wenn `data_version` nach einem
Tausch der Verbindung neu beginnt und der Zähler zugleich steigt — dann lädt
das Fenster einmal zu viel, derselbe Preis wie bisher. Und **wer einen neuen
Hintergrundweg baut, muss es wissen**: ein Weg, der den Zähler nicht bewegt,
fällt wieder still aus.

**Nachgetragen am 2026-10-05** (Nachlauf zum Gesamtreview über 0.6.0, #1660):
hier stand, eine Partie melde ihre Züge über eigene Ereignisse. **Das reichte
für den Tisch und nicht für die Liste** — die Partienliste des Wirts hing am
Datenstand, und der stand still, während ein Gast zog. Seither sagen es **alle
drei Zwecke des Horchpostens**: Handel und Paarung in ihrer Schleife, die
Partie am Nehmer des Wirts (`nebenherNehmer`), denn dort geht jeder Zug durch
und nur ein angenommener schreibt. **Die Gastseite schreibt nichts** in den
Bestand; ihr Tisch gehört dem Wirt.

**Und der Stand erreicht seither auch das Aufgeschlagene**: ein offenes Deck
und ein offenes Turnier lesen am fremden Stand nach (`useFremdNachlesen`),
leise — ohne Wechsel, ohne Vermerk als benutzt und ohne eine stehende Meldung
wegzuräumen. **Was er nicht erreicht, sind die Dateien eines Pakets**: sie
liegen nicht im Bestand, und eine Datei, die ein anderer umschreibt, bewegt
`data_version` nicht. Dort hält der Inhaltsstand des Puffers den Konflikt
beim Speichern.

## Wer das ändern will

ändert zuerst diesen Vermerk. Der wahrscheinlichste Anlass ist ein Bestand, der
groß genug wird, dass „alles neu laden" spürbar wird — dann ist nicht der
Melder falsch, sondern das, was auf ihn folgt.
