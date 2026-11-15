# 0116 — Eine Partie zeigt eine Sicht und nie den Zustand

**Stand:** angenommen · **Datum:** 2026-09-15
**Berührt:** [0044](0044-was-ein-werkzeug-schreiben-darf.md),
[0061](0061-was-ein-deck-hier-ist.md),
[0106](0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)

## Warum es diesen Vermerk gibt

Weil Werkstück G1 von Phase 5.5
(#1458)
vier Dinge festlegt, die man später hinterfragen wird — und jedes davon hat
einen Preis, der ausgesprochen gehört:

1. **Wer wie viel sieht** — und dass der Wirt mehr sieht als alle.
2. **Dass eine Suche im eigenen Deck überhaupt möglich ist**, obwohl sie ein
   Geheimnis zerstört.
3. **Dass die Mischpflicht daraus nicht erzwungen wird.**
4. **Dass hier nicht fair gemischt wird, sondern nur zufällig.**

## 1. Die Projektion ist der einzige Weg zur Oberfläche

`domain.Partie.Sicht(fuer)` rechnet aus dem Zustand, was **ein** Sitz sehen
darf. `Vollzustand` existiert daneben und hat an **keiner** der vier Flächen
einen Weg — kein DTO, kein MCP-Werkzeug, kein Endpunkt, kein Befehl.

**Warum das eine Entscheidung ist und keine Selbstverständlichkeit:** ein DTO
aus dem Vollzustand wäre bequemer und sähe **vollständiger** aus, nicht
falscher. Genau das ist die gefährliche Gestalt: ein Fehler, der wie ein
Fortschritt aussieht.

**Was es kostet:** wer den ganzen Tisch braucht — ein Beobachtungswerkzeug,
eine Wiedergabe im Nachhinein — bekommt ihn nicht über diese Flächen. Für die
Wiedergabe gibt es den Ereignisstrom; wer daraus mehr rechnen will, als seine
Sicht hergibt, muss es begründen und einen eigenen Weg bekommen.

**Eine Bibliothek ist auch für ihren Eigner verdeckt.** `ZoneBibliothek` gibt
`SichtVerdeckt`, und das gilt ohne Ansehen der Person. Wer seine eigene
Bibliothek läse, hätte sie sortiert.

## 2. Der Wirt hält den Vollzustand — und es steht auf der Seite

In einer lokalen Partie ist der Wirt dieses Programm; über ein Netz (G5,
#1462) ist
es der Rechner dessen, der serialisiert. Er **kann** technisch alles sehen.

Das Issue verlangt: „diese Vertrauensgrenze wird vor Beitritt offengelegt."
Eingelöst wird das nicht durch eine Zeile in einem Dokument, sondern **auf
der Seite**: der Wirt steht im Kopf des Tisches, und darunter steht ein Satz,
der sagt, was das heißt. Wer es erst danach erführe, hätte die Wahl nicht
gehabt.

## 3. Ein Blick ist erlaubt, und er kostet eine Mischpflicht

**Ohne Blick gäbe es kein Suchen** — und ohne Suchen kein Magic. Also gibt es
`zone.blick`: ein Sitz sieht in seine **eigene** verborgene Zone, und der
Inhalt kommt als empfängergebundene Ergänzung, die in keiner zweiten Sicht
auftaucht.

**Zwei Bedingungen sind Invarianten und keine Regeln**, und der Sandkasten
lockert sie ausdrücklich nicht:

- **Nur die eigene Zone.** „Offenbare deinem Gegner deine Hand" ist etwas
  anderes — es gibt jemand anderem eine Sicht, die er sonst nicht hätte. Das
  ist eine Regel und gehört in die Matrix von G4
  (#1461).
- **Nur eine verborgene Zone.** Ein Blick auf den Tisch ist keiner.

**Und er erzeugt eine öffentliche Mischpflicht.** Sie steht im **Zustand**
und nicht nur im Strom: dass jemand gesehen hat, ist ein Ereignis; dass er
seitdem **nicht** gemischt hat, ist eine Tatsache über den Tisch — und ein
Zuschauer, der später dazukommt, hat den Strom gar nicht.

## 4. Erzwungen wird sie nicht — und das ist der Sandkasten

Der Sandkasten gehorcht dem Menschen. Zwei Leute an einem Tisch mischen, weil
sie es schulden, nicht weil ein Programm sie zwingt.

**Was G1 deshalb schuldet, ist der Platz, an dem G4 es ablesen kann** — und
die **Sichtbarkeit**, die einen Verstoß auffällig macht, statt ihn zu
verhindern. Die Mischpflicht steht jedem am Tisch vor Augen; wer sie
übergeht, tut es vor Zeugen.

**Was das kostet:** in einer Partie unter `StrengeSandkasten` kann jemand
suchen und weiterspielen, ohne zu mischen. Das ist gewollt, solange die
Strenge es sagt — und es ist der Grund, warum es die drei Strengen gibt.

## 5. Zufällig gemischt, nicht beweisbar fair

Das Issue schließt es aus: „kein kryptografisch beweisbares faires Shuffle in
v1". Also:

- `service.Zufallsquelle` ist eine **Naht**, keine Funktion — im Betrieb
  Fisher–Yates über `crypto/rand` mit Ablehnungsverfahren, in Läufen eine
  eingesetzte Permutation. Ohne die Naht käme ein Lauf nie zweimal auf
  dasselbe Brett, und die Zusage „derselbe Startsnapshot plus dieselbe Folge
  ergibt denselben Zustand" wäre unprüfbar.
- **Der Wirt mischt und kennt die Reihenfolge danach.** Das ist dieselbe
  Vertrauensgrenze wie oben und keine zweite.

**Was das kostet:** gegen einen unehrlichen Wirt schützt hier nichts. Wer
kompetitiv spielen will, braucht ein Protokoll mit beidseitigem Commitment —
das ist kein Werkstück dieser Runde, und es als „später" zu führen wäre
ehrlicher als es als „vorhanden" zu zeigen.

## Was verworfen wurde

| Verworfen | Warum |
| --- | --- |
| **Ein Vollzustands-DTO für den Wirt** | Es sähe vollständiger aus, nicht falscher — und wäre damit der Fehler, den niemand bemerkt. Der Wirt braucht ihn nicht, um zu **zeigen**; er braucht ihn, um zu **rechnen**, und das tut er innen |
| **Die Mischpflicht nur im Ereignisstrom** | Ein Zuschauer, der später dazukommt, hat den Strom nicht — und jeder Empfänger müsste sie nachrechnen |
| **Ein Zwang zum Mischen schon in G1** | Er wäre eine Regel, und Regeln sind G4. Ein Sandkasten, der eine Regel erzwingt, ist keiner |
| **`math/rand/v2` global statt `crypto/rand`** | Es wäre kürzer und „sicher genug". Aber das Issue sagt „sicher gezogen", und eine Naht, die man ohnehin braucht, kostet dieselbe Zeile |
