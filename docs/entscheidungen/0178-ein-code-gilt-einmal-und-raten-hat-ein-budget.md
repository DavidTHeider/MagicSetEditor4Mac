# 0178 — Ein Code gilt einmal, und Raten hat ein Budget

**Stand:** angenommen am 2026-10-07 · SEC-08, SEC-09, SEC-10
(#2149,
#2150,
#2151).

## Die Frage

**Der zweite Faktor war zustandslos.** `totp.Stimmt` rechnete, ob ein Code
zu einem der drei Schritte im Fenster passt, und mehr nicht: derselbe Code
galt neunzig Sekunden lang beliebig oft, und wer das Kennwort hatte, konnte
Codes und Zettel raten, so schnell der Bestand antwortete — an der
Kommandozeile und am Fenster ohne jede Grenze, an REST und MCP nur hinter
dem allgemeinen Budget für Marken, das einen Nutzer nicht kennt.

Die Issues ließen die Entscheidung ausdrücklich offen: Schlüssel, Fenster,
Grenze, Rücksetzung, Recovery und Schema. **Sie sind hier entschieden; keine
davon ist aus MSE2** — MSE2 hat keine Anmeldung.

## Die Entscheidung

### Ein Schritt gilt einmal (SEC-09)

**Verbraucht wird der Schritt, zu dem der Code passt** — nicht die Uhr beim
Prüfen (`totp.PassenderSchritt`). `nutzerfaktoren.letzter_schritt` hält den
zuletzt angenommenen, und nur ein **späterer** gilt
(`UPDATE … WHERE letzter_schritt < ?`). Damit gilt auch ein älterer Schritt
aus dem Fenster nicht mehr, sobald ein neuerer angenommen ist. RFC 6238 §5.2
verlangt das für denselben Code; für den älteren ist es die strengere Lesart.

- **Zwei Prozesse, ein Code**: der Bestand entscheidet, der zweite trifft
  null Zeilen.
- **Der Code der Bestätigung** zählt ebenso — wer ihn beim Einrichten
  mitlas, meldet sich damit nicht an. Der Faktor entsteht mit diesem
  Schritt (`FaktorAnlegen` mit `LetzterSchritt`), nicht in einem zweiten
  Schreiben danach.
- **Neu eingerichtet** heißt neue Zeile — mit dem Schritt ihres
  Bestätigungscodes, nicht null (berichtigt nach dem Nachlauf über F).
- **Uhrsprung zurück**: Codes, deren Schritt nicht später ist, gelten nicht,
  bis die Uhr den angenommenen überholt hat. Der Zettel bleibt der Weg.
- **Ein verbrauchter Code** fällt weiter zum Zettel und scheitert dort — er
  ist ein Fehlversuch wie jeder andere.

### Ein Budget je Nutzer (SEC-10)

**Fünf Versuche frei, danach 30 Sekunden Sperre, mit jedem weiteren
Fehlversuch doppelt so lang, höchstens eine Stunde** (`domain.Faktorsperre`).
Ein gelungener Code löscht den Haushalt.

- **Je Nutzer, nicht je Faktor**: wer rät, rät an jedem Weg — Anmeldung,
  Profilwechsel, Entsperren und der Zettel beim vergessenen Kennwort
  (`CodeFinden`). Alle gehen durch `imBudget` im Dienst.
- **Vorab gezählt**: jeder Versuch meldet sich an, **bevor** geprüft wird,
  und zählt als Fehlversuch, bis er gelingt. Erst zu prüfen und dann zu
  zählen ließe gleichzeitige Versuche am vierten Fehlversuch alle durch;
  gemessen mit zwei Vorräten auf einer Datei: genau einer geht.
- **Während der Sperre wird nicht geprüft** — auch ein richtiger Code und
  ein richtiger Zettel bekommen `faktorGesperrt`. Sonst wäre der Zettel ein
  Weg um die Sperre herum.
- **Der Sensor geht nicht hindurch**: Touch ID lässt sich nicht raten, und
  wer von einem Ratenden ausgesperrt wurde, kommt am eigenen Rechner so
  herein (an REST und MCP gibt es ihn nicht, SEC-07).
- **Ein Fehler unterwegs zählt** — ein Schlüsselbund, der nicht antwortet,
  ist von außen von einem falschen Code nicht zu unterscheiden.
- **Die allgemeinen Budgets bleiben unabhängig**: das für falsche Marken am
  MCP-HTTP (SEC-12) und das Gedränge vor REST (SEC-13/14) zählen
  Verbindungen und Anfragen, dieses zählt einen Menschen.

**Die Zahlen sind eine Entscheidung dieses Projekts — und sie weichen von
NIST ab.** NIST SP 800-63B §5.2.2 verlangt, aufeinanderfolgende
Fehlversuche je Konto auf **höchstens hundert** zu begrenzen; die steigenden
Wartezeiten sind dort eine Ergänzung, kein Ersatz. Hier gibt es diese
Obergrenze nicht: nach jeder Sperre geht wieder ein Versuch (nachgemessen
vom Zweitblick über F, Befund F3: der 101. Folgeversuch geht).

**Warum, und was es kostet**: eine endgültige Sperre bräuchte einen Weg
zurück — einen Verwalter, eine zweite Person, einen Dienst —, und den hat
eine Anwendung auf einem Rechner ohne Konto nicht. Wer gesperrt bliebe,
käme nur über Touch ID am eigenen Rechner wieder herein, und ohne Touch ID
gar nicht. Nach der Sättigung bleiben dafür 24 Versuche am Tag; bei sechs
Ziffern und dem Fenster ±1 sind das drei passende von einer Million je
Versuch, rund 2,6 % im Jahr. Wer NIST-Konformität braucht, muss diese
Entscheidung neu treffen.

**Auch die Bestätigung beim Einrichten** geht durch das Budget, und ihr
Faktor entsteht mit seinem Schritt in **einem** Schreiben (Zweitblick über
F, Befunde F1 und F2): getrennt galt der Bestätigungscode zwischen Anlage
und Verbrauch noch einmal.

### Das Schema (SEC-08)

Wanderung **0097**: `letzter_schritt INTEGER NOT NULL DEFAULT 0` an
`nutzerfaktoren` und die Tafel `faktorhaushalt` (Nutzer, Fehlversuche,
`gesperrt_bis` als Unixzeit). **Ein alter Bestand verliert nichts**: jeder
Faktor beginnt bei Schritt null, und keine Haushaltszeile heißt kein
Fehlversuch — gemessen an einem Bestand vor 0097 mit TOTP-Faktor und Zettel.

## Was es nicht ist

- **Keine Sperre des Kontos.** Das Kennwort allein öffnet weiter, was es
  ohne zweiten Faktor öffnete; gesperrt ist nur das Prüfen von Codes.
- **Kein Schutz gegen einen, der den Bestand schreiben kann** — wer die
  Datei hat, löscht den Haushalt. Der Bestand ist die Grenze dieses
  Programms, nicht seine Verschlüsselung.
- **Ein Ratender kann den Nutzer aussperren** — nach dem Kennwort. Das ist
  der Preis jeder Sperre; die Stunde ist die Obergrenze, und Touch ID am
  eigenen Rechner bleibt.

## Verworfen

- **Ein festes Fenster** (fünf Versuche je Viertelstunde): 480 Versuche am
  Tag, zwanzigmal so viele wie mit Verdopplung.
- **Zählen nach dem Prüfen**: lässt gleichzeitige Versuche durch, siehe oben.
- **Der Haushalt im Prozess**: vier Prozesse auf einer Datei sähen vier
  Budgets.
