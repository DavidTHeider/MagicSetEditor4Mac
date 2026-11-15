# 0120 — Eine Zusage bindet den Schreibweg

**Stand:** angenommen
**Datum:** 2026-09-18
**Werkstück:** `T2` von Phase 5.6
(#1467)

## Lage

Seit `T2` hält eine **Reservierung** Stücke eines Exemplars für einen
Handelsvorgang fest. Sie prüft beim Eintragen, ob genug frei ist — und danach
prüfte sie **niemand mehr**.

Der Fall ist gewöhnlich und nicht konstruiert: jemand sagt zwei von drei
Stücken zu, und pflegt zwei Tage später seine Sammlung. Er setzt die Menge auf
eins, weil er eine Karte verschenkt hat. Der Bestand nimmt es an; der Partner
wartet auf zwei Karten, die es nicht gibt.

**Das Issue nennt die ganze Klasse**: Einzelbearbeitung, Massenpflege,
Import/Merge, Löschen, Profilübernahme, Wiederherstellung. Gezählt am
2026-09-18 sind es **drei** SQL-Anweisungen, die Zeilen der Sammlung schreiben,
und alle drei stehen in `internal/storage/sammlung.go`. Massenpflege und
Listeneinfuhr haben keinen eigenen Weg — sie gehen durch dieselbe Tür.

## Entscheidung

**Drei Sätze.**

### 1 — Was zugesagt ist, darf nicht wegfallen

Ein Schreiben, das weniger Stücke stehen ließe, als aktive Zusagen
festhalten, wird mit [domain.ErrConflict] abgewiesen — und schreibt **nichts**.
Eine Zeile mit einer aktiven Zusage lässt sich gar nicht löschen: an ihr hängt
`reservierungen.exemplar_id` mit `ON DELETE CASCADE`, und die Kaskade nähme die
Zusage stillschweigend mit.

**Aktiv heißt**: ohne Frist, oder die Frist läuft noch, **oder** eine Übergabe
ist bestätigt. Eine abgelaufene Zusage hält nichts fest; eine übergebene hält,
auch wenn die Frist vorbei ist. Das ist derselbe Satz wie in
`domain.Reservierung.Aktiv`, und er steht dort nur einmal.

### 2 — Aber nicht jede Änderung ist verboten

**Zustand, Sprache, Ausführung, Notiz und Lagerort darf jemand ändern, solange
die Stückzahl reicht.** Die Zusage trägt die **Fassung**, gegen die sie
gerechnet wurde; jede Änderung hebt sie, und der Handel liest an seiner eigenen
Prüfung ab, dass neu zuzustimmen ist.

Das ist die Antwort des Issues auf „vor Vorbereitung darf relevante Änderung
Revision invalidieren". **Ohne diesen Satz wäre eine Sammlung mit einem offenen
Handel nicht mehr zu pflegen** — und wer sie trotzdem pflegen will, gäbe den
Handel frei, statt ihn zu prüfen.

### 3 — Die Regel steht in der Domäne, der Augenblick im Bestand

`domain.AenderungPasst` und `domain.LoeschenPasst` sagen, **was** gilt.
Gefragt werden sie in `internal/storage/sammlung.go` — **innerhalb derselben
Transaktion**, die schreibt.

**Warum nicht in `service`**: dazwischen läge ein Fenster. Ein Dienst, der erst
die Zusagen liest und dann schreibt, lässt in dem Augenblick dazwischen eine
zweite Zusage zu — und die rechnet gegen einen Besitz, den der erste Zug gerade
wegschreibt. Dieselbe Überlegung trägt schon `Reservierungsablage.Festhalten`;
sie ist keine neue Erfindung, sondern dieselbe Stelle noch einmal.

**Das ist keine Fachlogik in `storage` (§2)**: die Regel steht in `domain`, und
`storage` ruft sie an dem einzigen Punkt, an dem sie wirken kann.

## Der Preis, und er wird ausgesprochen

**Ein Bestand mit offenen Zusagen lässt sich nicht beliebig kleinrechnen.** Wer
eine Karte verliert, während sie zugesagt ist, muss erst die Zusage freigeben.
Das ist gewollt — der Alternativfall ist ein Partner, der auf nichts wartet —
aber es ist eine Hürde, die es vorher nicht gab, und sie erscheint an einer
Stelle (der Sammlungspflege), an der niemand an einen Handel denkt.

**Die Meldung nennt deshalb beide Zahlen und den Ausweg.**

**Und die Wiederherstellung geht daran vorbei.** Sie tauscht die **Datei** und
kennt kein SQL dieses Programms; was danach mit offenen Zusagen ist, bleibt
eine fachliche Frage und ist kein Schreibweg. Sie steht als Grenze im Kopf des
Wächters.

## Verworfen

**Jede Änderung an einer zugesagten Zeile verbieten.** Das wäre die einfachere
Regel und die falsche: eine Sammlung mit einem offenen Handel wäre nicht mehr
zu pflegen, und der erste Reflex wäre, die Zusage wegzuräumen — womit die Regel
genau das bewirkte, was sie verhindern soll.

**Die Zusage beim Schreiben still anpassen** — die festgehaltene Menge auf das
senken, was noch dasteht. Sie sieht freundlich aus und nimmt dem Partner die
Auskunft: sein Angebot stünde weiter da und meinte etwas anderes. Eine Zusage,
die sich unter der Hand ändert, ist keine.

**Den Wächter über die Schreibwege weglassen** und sich auf die drei Stellen
verlassen. Sie stimmen heute; die vierte kommt, wenn jemand einen Massenweg
schneller machen will, und dann ist die Zusage wieder offen — ohne dass es
jemand merkt.
