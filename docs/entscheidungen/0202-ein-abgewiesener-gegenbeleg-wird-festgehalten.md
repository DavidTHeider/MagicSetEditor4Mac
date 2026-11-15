# 0202 — Ein abgewiesener Gegenbeleg wird festgehalten, nicht gemeldet

**Datum:** 2026-10-20
**Stand:** angenommen
**Anlass:** #2500
(Rest aus #2492, Punkt 3)

## Der Befund

Passte ein Gegenbeleg nicht zum eigenen Beleg, endete
`Handelspost.gegenbeleg` mit einem Fehler. Die Aufnahmeschleife der Sitzung
(`app/handelsaufnahme.go`) beendet sich bei jedem Fehler — also endete die
ganze Sitzung, und mit jedem neuen Anlauf wieder. Beim Absender blieb die
Nachricht im Ausgang („hinausgegangen"), beim Empfänger stand nichts, und
beide Fenster sagten dauerhaft „Nur du hast unterschrieben".

## Die Entscheidung

1. **Der Empfänger hält den Befund fest** — am eigenen Beleg, einer je
   Vorgang (der jüngste gilt), mit Absender, behauptetem Abdruck, einem von
   fünf Mängeln und dem Satz der Prüfung (Wanderung
   `0101_gegenbelegbefunde.sql`). **Dann geht die Zustellung weiter**: die
   Sitzung endet nicht, was nach dem Beleg in ihr kommt, wird zugestellt, und
   ein neuer Anlauf mit derselben Urkunde hält denselben Befund fest — die
   Gegenzeichnung wird bei jedem Anlauf versucht (TRADE-03), das Festhalten
   ersetzt.
2. **Ein Fehler bleibt, was wiederholbar ist**: ohne eigenen Beleg (hier ist
   noch nicht verbucht, TRADE-03) und wenn der Befund sich nicht festhalten
   lässt.
3. **Der Absender erfährt es auf seiner eigenen Seite, nicht über eine neue
   Nachricht.** `Belegurkunde.PasstZu` vergleicht Vorgang, Revision,
   Parteien und Zeilen **gespiegelt** — die Prüfung ist symmetrisch. Weicht
   sein Beleg vom hiesigen ab, weicht der hiesige auch von seinem ab, und
   wenn er ihn bekommt, hält er denselben Mangel fest.

## Einordnung nach §16

**Verträglich am Draht**: keine neue Nachrichtenart, kein neues Feld in einer
Nachricht. **Erweiternd an den Flächen**: `befund` am Beleg bei
Kommandozeile, Dienst und MCP (0033, Regel 6).

## Verworfen

- **Eine Nachricht „abgewiesen" an den Absender.** Ein Empfänger mit älterer
  Fassung kennte die Art nicht; seine Post wiese sie ab, und genau das
  beendet die Sitzung — derselbe Fehler, nur auf der anderen Seite. Eine neue
  Art bräuchte eine neue Fassung des Kanals (`mse4mac/trade/1`).
- **Den Befund als Fehler weiterreichen, nur mit Namen.** Der Fehler ist es,
  der die Sitzung beendet; ein besserer Name hätte daran nichts geändert.
- **Jeden Befund anhängen.** Ein Gegenüber, das immer neue Belege schickt,
  füllte den Bestand; der Grund jedes einzelnen steht im Protokoll.

## Was das kostet

**Eine Unterschrift, die nur in eine Richtung nicht trägt**, ist nicht
symmetrisch: hält B die Unterschrift von A für falsch, A die von B aber für
echt, steht bei A „beide haben unterschrieben" und bei B der Befund. Über die
gesicherte Leitung (TLS 1.3 mit festgenagelter Peerkennung, oder WireGuard)
kommt ein Beleg nicht verändert an; dieser Fall setzt einen Fehler im
Programm einer Seite voraus. Er bleibt sichtbar — bei B — und wird nicht
still.
