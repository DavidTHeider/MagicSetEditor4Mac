# 0108 — Eine Balanceanalyse hat keine Zahl

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Anlass:** Werkstück A13.2b von Phase 5.3
(#1452)

## Die Frage

Das Issue verlangt eine Balanceanalyse — „wie steht diese Karte im Vergleich
da?" — und verbietet im selben Atemzug:

> kein objektiver Balancewert aus einem Ähnlichkeitsscore

Beides zusammen ist die Frage. Eine Ähnlichkeitssuche liefert für jede
Vergleichskarte eine Zahl; sie liegt da, sie sieht aus wie ein Maß, und sie in
eine Note zu überführen ist **eine Zeile Code**. Wer die Zusage als *Prüfung*
baut, hat das Feld und einen Wächter darüber — und der nächste Beitrag hebt den
Wächter auf.

## Was gilt

**Es gibt kein Feld für eine Zahl.** `domain.Balanceangaben` trägt
`Vergleiche` und `NichtBetrachtet`, sonst nichts. Der Verzicht ist nicht
geprüft, sondern **gestaltet**: wer eine Note wollte, müsste zuerst das Feld
schaffen, und dabei stößt er auf diese Entscheidung.

**Aufzählbar ist er trotzdem.** `domain.Balancefelder()` nennt die Felder, die
eine Balanceanalyse trägt, und ein Lauf geht darüber und sucht nach „wert",
„punkt", „rang" und „score". Er prüft damit **nicht**, dass keine Zahl
existiert — das kann er nicht —, sondern dass niemand eine einführt, ohne die
Liste anzufassen. **Was er nicht hält**: eine Zahl, die in einem der
vorhandenen Felder als Text steht. Dagegen hilft kein Wächter, sondern die
Anweisung an das Modell, die ausdrücklich nach *Unterschieden* fragt und
„stärker" verbietet.

**Ein Vergleich ohne konkreten Unterschied ist ein Befund.** Das ist die
positive Hälfte derselben Entscheidung. „Ähnlich" sagt der Rang schon, der die
Karte gefunden hat — und was der bedeutet, steht in
[0103](0103-wie-zwei-trefferlisten-zusammenkommen.md): **er ordnet, er bewertet
nicht.** Eine Aussage entsteht erst, wenn jemand sagt, *was* anders ist:
Manawert, Effekt, Zusatzfähigkeit.

**Strategie, Format und Stand kommen vom Programm.** Ein Modell bekommt
Ausschnitte und weiß nicht, wie sie gefunden wurden; eines, das seine eigene
Suche beschriebe, beschriebe eine erfundene. Die Strategie nennt deshalb den
Recherchefall, den **tatsächlich gelaufenen** Modus, die Grenzen — und dass
**kein Reranker** im Spiel ist, was A15 gemessen hat
([RETRIEVALMESSUNG.md](../RETRIEVALMESSUNG.md)).

**Der Modus wird gelesen und nicht gesetzt.** Ein Index ohne Einbettungsmodell
sucht **lexikalisch**; eine Strategie, die „hybrid" behauptete, wäre dann
falsch — und zwar genau dort, wo jemand wissen will, warum gerade diese Karten
dastehen.

## Was das kostet

**Zwei Karten lassen sich nicht sortieren.** Wer wissen will, welche von
dreißig Karten die stärkste ist, bekommt hier keine Reihenfolge, sondern
dreißig Vergleiche zum Lesen. Das ist die Zumutung, und sie ist gewollt: eine
Reihenfolge aus Ähnlichkeitsrängen sähe aus wie ein Urteil und wäre eine
Rechenausgabe.

**Und ein Vergleich ohne Format gilt für keines besonders.** Er wird trotzdem
erzeugt — als **Einschränkung** in der Antwort und nicht als Fehler am Feld.
Eine Werteliste über „Commander, Pauper, …" gäbe es hier nicht: dieses Programm
kennt kein Turnierformat, und eine gepflegte Liste fremder Regelwerke wäre eine
Zusage, die niemand einlöst.

## Wie man das Gegenteil erkennt

**An einem Zahlenfeld in `Balanceangaben` oder `BalanceDTO`** — oder an einer
Sortierung über Vergleichskarten. Beides hieße, dass aus dem Rang doch ein Maß
geworden ist.

## Verworfen

**„Eine Note von 1 bis 10, das versteht jeder."** Genau das ist der Einwand:
sie wird verstanden, und zwar als Messung. Hinter ihr stünde ein Kosinusabstand
zwischen zwei Einbettungen, den ein Modell erzeugt hat, das über Kartenstärke
nichts weiß.

**„Dann eben nur intern, zum Sortieren."** Eine interne Zahl, die die
Reihenfolge der Anzeige bestimmt, ist eine Bewertung mit unsichtbarer Herkunft
— schlechter als eine sichtbare, nicht besser.

**„Ein Wächter verbietet das Feld."** Er verbietet es, solange ihn niemand
ändert. Ein fehlendes Feld muss erst geschaffen werden, und das ist die
größere Hürde — dieselbe Überlegung wie in
[0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md) über den
Listener, den es nicht gibt.
