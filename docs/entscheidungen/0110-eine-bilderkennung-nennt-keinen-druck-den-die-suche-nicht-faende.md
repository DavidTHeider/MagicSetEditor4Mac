# 0110 — Eine Bilderkennung nennt keinen Druck, den die Suche nicht fände

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Verwandt:** [0108](0108-eine-balanceanalyse-hat-keine-zahl.md) ·
**Anlass:** Werkstück A13.3 von Phase 5.3
(#1452)

## Die Frage

**Wer entscheidet, welche Karte auf einem Foto zu sehen ist?**

Die naheliegende Antwort ist: das Modell. Man gibt ihm das Bild und fragt
„welche Karte ist das?", und es antwortet „Lightning Bolt aus Alpha". Das
funktioniert in der Vorführung und ist der Grund, warum dieser Vermerk
existiert.

## Die Entscheidung

**Das Modell liest. Die Suche findet. Ein Mensch wählt.**

Was das Modell beisteuert, sind **Beobachtungen** — gelesener Name, gelesener
Setcode, gelesene Sammlernummer, gelesene Typzeile, erkannte Bauart,
beschriebene Illustration. Daraus baut `internal/domain` einen
**Suchbegriff**, und die Kandidaten kommen aus **derselben** Drucksuche, die
eine getippte Kartenliste geht (A2).

Damit kann eine Bilderkennung keinen Druck nennen, den die Suche nicht auch
gefunden hätte.

## Warum

**Weil „welche Karte ist das?" zwei Fragen sind und nur eine davon ein Modell
beantworten kann.**

Die erste ist: *was steht da?* Ein Sichtmodell liest Text von einem Foto; das
ist eine Wahrnehmung, und sie lässt sich am Bild nachprüfen. Die zweite ist:
*welches Papier ist das?* Das ist eine Frage an einen **Bestand** — an die
Liste aller Drucke, die es gibt. Ein Modell, das sie beantwortet, antwortet
aus dem, was in seinen Gewichten steht, und das ist ein Stand von irgendwann.

**Gemessen:** „Counterspell" ohne Setangabe passt bei Scryfall auf **72**
Drucke (nachgelesen am 2026-08-30, dieselbe Zahl, für die es die
Listendeutung gibt). Ein Modell, das einen davon nennt, hat 71-mal geraten und
sieht dabei jedes Mal gleich sicher aus.

## Die Unsicherheit ist eine gezählte Zahl und kein Prozentwert

**Dieselbe Überlegung wie bei der Balanceanalyse
([0108](0108-eine-balanceanalyse-hat-keine-zahl.md)), und hier trägt sie
weiter.** Dort fehlt die Zahl, weil keine zu haben ist. Hier **gibt es** eine:
die Zahl der passenden Drucke. Sie kommt aus der Suche und nicht aus dem
Modell, sie ist nachzählbar, und sie sagt etwas Wahres:

> Die Karte ist erkannt. Der **Druck** ist offen — es passen 72.

„87 % sicher" sagte dasselbe nicht, sondern etwas, das wie eine Messung
aussieht und keine ist. `weitere` steht daneben, weil die Suche eine Seite
gibt und nicht alles: ohne dieses Feld hieße „72" manchmal „mindestens 72",
und niemand sähe, wann.

## Die Evidenz bleibt getrennt

Das Issue verlangt es wörtlich: „OCR/Layout/Face/Set/Kunstwerk als **getrennte
Evidenz** behandeln." Der Grund ist derselbe wie oben — ein gelesener Name und
eine beschriebene Illustration sind zwei Auskünfte mit verschiedener
Verlässlichkeit, und wer sie zu einer Zahl verrührt, kann hinterher nicht
sagen, woran es lag.

Daraus folgen zwei Dinge, die leicht zu übersehen sind:

1. **„Nicht lesbar" ist ein eigener Zustand neben dem leeren Wert.** „Die
   Fußzeile ist abgeschnitten" erklärt, **warum** ein Kandidat offenbleibt;
   „dort steht nichts" erklärt es nicht. Und ein Wert an einer als unlesbar
   gemeldeten Stelle ist ein **Befund**: wer sagt, er sehe die Nummer nicht,
   und trotzdem eine nennt, hat sie geraten.
2. **Bauart und Illustration stützen keinen Kandidaten.** Eine Bauart passt
   auf Tausende, eine Bildbeschreibung lässt sich gegen keinen Druck halten.
   Sie stehen in der Antwort, damit jemand sieht, was das Modell gesehen hat —
   nicht, damit sie mitzählen.

## Und deshalb wird nichts angelegt

Das Issue nennt es als eigene Abnahme: „Menge/Exemplar wird **nicht
automatisch** angelegt", und „Bild einer mehrdeutigen/falschen/keiner Karte
erzeugt keine falsche Printing-Bindung oder Sammlungsmengen."

**Es gibt deshalb keine Bindung, die einen Kandidaten in die Sammlung trüge.**
`eindeutig` heißt „genau ein Druck passt, und dahinter liegt nichts" — es
heißt nicht „nimm ihn". Wer ein Exemplar eintragen will, tut das im
Sammlungsformular, und dann steht fest, welcher Druck gemeint ist.

## Was verworfen wurde

**Scrydex als Bildquelle.** Das Issue erlaubt es ausdrücklich — „Scrydex nur
bei belegter Magic-Vision-Capability" —, und genau diese Bedingung ist nicht
erfüllt: `internal/scrydex` liest Karten über vier Wege und belegt **keine**
Sichtfähigkeit ([0056](0056-drei-weitere-kartendienste.md),
[0095](0095-was-cardtrader-sagt-und-was-davon-gilt.md) für die Form der
Positivliste). Es sieht keine Bilder; es kennt Karten. Ein Weg dorthin wäre
eine Fähigkeit, die niemand nachgewiesen hat — und eine Ablehnung ohne Deckung
ist eine Ausrede, eine Zusage ohne Deckung wäre schlimmer (§10).

**Ein eigener Bildindex zur Kandidatensuche.** Das Issue lässt ihn
ausdrücklich offen — „Vision **darf später** einen separat versionierten
Bildindex verwenden, wenn Text-/Bildraumkompatibilität, Runtime und
Datennutzung belegt sind; für 0.5.0 kein zusätzlicher Pflichtumfang und kein
A17." Nichts davon ist heute belegt, und der Textencoder aus A15 erkennt keine
Bilder: ein gemeinsamer Raum für Text und Bild ist eine Eigenschaft des
Modells und keine Annahme, die man trifft.

**Ein Modell, das die Karte nennt.** Siehe oben. Wer es ausprobieren will,
findet den Unterschied am schnellsten an einer Karte mit vielen Drucken.

## Was es kostet

**Eine abgeschnittene Karte bleibt offen, wo ein ratendes Modell geantwortet
hätte.** Wer nur die Illustration fotografiert, bekommt hier eine
Beobachtung und keinen Kandidaten — der Suchbegriff ist leer, und die Suche
wird gar nicht erst gefragt. Das ist gewollt: eine Suche über die leere
Zeichenkette fände alles.

**Und die Erkennung ist nur so gut wie die Drucksuche.** Fällt Scryfall aus,
fällt sie aus; was das Modell gelesen hat, steht trotzdem da. §16, Regel 5
gilt: der letzte gute Stand bleibt, und er sagt, wie alt er ist.
