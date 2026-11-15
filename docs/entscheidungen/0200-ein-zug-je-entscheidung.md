# 0200 — Ein Zug je Entscheidung: der Abschluss eines Tauschs in drei Klicks je Seite

**Stand:** angenommen am 2026-10-19 · #2381 (P13, Befund C-16). Baut auf [0151](0151-der-handel-in-vier-phasen.md) und [0152](0152-was-geschrieben-ist-bleibt.md) auf; der Vertrag aus T6 und T10 bleibt. Umgesetzt in `frontend/src/handelszug.ts`, `frontend/src/haken/useHandelszug.ts`, `frontend/src/bausteine/Zugmeldung.tsx`, `Handelsraum.tsx`, `Handelsschritte.tsx`, `Handelsbuchung.tsx`, `Handelsvorgang.tsx`, `Handelsliste.tsx`.

## Lage

Vom Angebot bis „beidseitig verbucht“ brauchte die Initiatorin am
2026-10-14 **35 Handlungen**, das Gegenüber **30** (P0,
`ausgangswerte09handeln`). Jeder Schritt kostete seinen Knopf, dazu „Schritt
4“, „In den Ausgang legen“ und „Jetzt übertragen“; das Exemplar wurde in
einer eigenen Phase gewählt, obwohl es oft nur eines gibt, und die
Anbietende stimmte ihrem eigenen Angebot noch einmal zu. Das Abnahmekriterium
verlangt **höchstens drei Bestätigungen je Seite**, die Stimmen beider Seiten
weiterhin getrennt, und jeden Teilschritt mit seinem eigenen Fehlerweg.

**MSE2 handelt nicht** (siehe [0149](0149-was-eine-handelsnachricht-behaupten-darf.md));
Marke **new**.

## Entscheidung

1. **Ein Zug geht die Schritte seiner Entscheidung, die der Vertrag als
   nächste nennt, bis das Gegenüber dran ist.** Zwei Entscheidungen gibt es:
   die **Einigung** (anbieten, zustimmen) und den **Abschluss**
   (vorbereiten, übergeben, verbuchen, abgleichen, die eigene Stimme zum
   Abschluss). Was als Nächstes kommt, rechnet das Fenster nicht aus — es
   liest nach jedem Schritt `naechster` und `lage` aus dem Dienst (0151 §2).
   **Eine Einigung zieht keinen Abschluss nach sich**: wer einem
   Gegenangebot zustimmt, hat damit nichts übergeben.
2. **Anbieten trägt die eigene Zustimmung** — als zweiter Schritt im selben
   Zug, mit eigener Zeile im Protokoll, und **der Knopf sagt es**: „Anbieten
   und zustimmen“, wie im Artboard `Markttisch09`. Eine Zustimmung, die
   niemand sieht, wäre keine. Die Matrix bleibt, wie sie ist: wer das Angebot
   ändert, wird neu gefragt, denn eine Zustimmung bindet **diese** Revision.
3. **Die Initiatorin bereitet vor und bestätigt die Übergabe in einem Zug**
   („Vorbereiten und übergeben“). Zwischen beiden geschieht nichts, worauf
   das Gegenüber antworten könnte; vorbereitet ist nur, damit die Exemplare
   festgehalten sind, bevor sie weggehen.
4. **Verbuchen geht bis zur eigenen Abschlussstimme** — verbuchen, den
   Abgleich melden, den eigenen Abschluss bestätigen —, und **es legt die
   Übergabe zuerst in den Ausgang**. Der Grund ist am nativen Fenster
   gefunden worden: eine Stimme zählt drüben nur im selben Stand
   (TRADE-03, `NimmtStand`). Die Stimme der Initiatorin zur Übergabe reist
   mit dem Schritt nach „vorbereitet“ und zählt beim Gegenüber deshalb
   nicht; es bestätigt die Übergabe selbst, und bei der Initiatorin schließt
   sie, wenn diese Stimme ankommt — ohne dass das Gegenüber es erfährt. Ein
   Abschlussschritt wird drüben nicht nachgespielt; ohne die Übergabe im
   Ausgang bliebe das Gegenüber in „vorbereitet“ stehen. **Wo dieses Fenster
   die Stimme des Gegenübers schon kennt**, verbucht die eigene Bestätigung
   der Übergabe im selben Zug („Übergeben und verbuchen“). **Ein Zug wartet
   nie auf das Gegenüber**: er endet, wo es dran ist.
5. **Die Stimme des Gegenübers bleibt ein eigener Klick**
   („Gegenseite bestätigt“) — die Zwei-Unterschriften-Regel aus T6 gilt
   unverändert, und über die Leitung reist sie weiterhin nicht (#1637).
6. **Festgehalten wird vorbelegt, wo es eindeutig ist**: derselbe Druck
   steht nur in einer eigenen Zeile, und im Bestand gibt es genau eine Zeile
   dieses Drucks, die kein Wunsch ist und reicht. Sonst hält der Zug vor dem
   Schritt an und sagt, wo gewählt wird.
7. **In den Ausgang gelegt wird, wenn ein Stand verlassen wird, und am
   Ende; übertragen wird einmal, am Ende — auf den Klick hin.** Drüben spielt
   die Leitung höchstens einen Schritt nach und keinen Abschlussschritt: die
   Übergabe muss deshalb im Ausgang liegen, **bevor** gebucht wird. Und eine
   Stimme reist mit ihrem Stand — eine zweite Nachricht derselben Art zur
   selben Revision hieße drüben „bekannt“. **Eine Wiederholung ohne Zutun gibt
   es weiterhin nicht** (0152 §4): scheitert das Übertragen, bleibt der Stand,
   die Nachricht liegt im Ausgang, und „Jetzt übertragen“ steht daneben.
8. **Jeder Teilschritt hat seinen eigenen Fehlerweg.** Was geglückt ist,
   bleibt (§7, Regel 3); der Zug hält an der Stelle, die scheiterte, und
   nennt sie — Festhalten, Schritt, Lesen, Ausgang oder Übertragen. Nach
   einem gescheiterten **Schritt** legt er nichts mehr in den Ausgang: der
   Stand trüge die Stimme nicht, um die es ging. Ein zweiter Klick beginnt
   bei dem Schritt, den der Vertrag dann nennt, und geht keinen doppelt.
9. **Jede Freigabe bleibt einzeln**: für verbuchen und für die eigene Stimme
   holt der Zug je einen Nachweis unmittelbar vor dem Schritt, gebunden an
   den Stand von dann (T10). Der Klick ist die Zustimmung zu jedem Schritt,
   den sein Knopf nennt — deshalb nennt er sie.
10. **Der eine Handel, an dem du dran bist, steht beim Öffnen offen** (H4),
    solange nichts aufgeschlagen ist. Bei mehreren wählt ein Mensch.
11. **Ein Zug geht höchstens bis zu dem Schritt, den sein Knopf nennt**
    (`bis`; Befund 1 des Zweitblicks über Phase 8): „Übergabe bestätigen“
    verbucht nicht, auch wenn die Stimme des Gegenübers inzwischen da ist und
    die Übergabe schließt; „Zustimmen“ bereitet nichts vor.
12. **Ein Zug geht nur von dem Stand aus, den der Mensch sah.** Vor dem
    ersten Schritt liest er ihn nach — Revision, Abdruck, Zustand, Stimmen,
    Lage und nächster Schritt; weicht etwas ab, geht er nichts, sagt es und
    zeigt den neuen Stand. **Und der Dienst prüft dasselbe noch einmal**
    (#2496, am 2026-10-19): jeder Schritt nennt den Abdruck des Stands, den
    der Zug zuletzt kannte, und `SchreitenAls` weist ihn als Konflikt ab,
    wenn der Vorgang inzwischen eine andere Revision ist
    (`domain.Handelsangebot.PruefeGesehen`) — an Fenster, Kommandozeile
    (`--gesehen=`) und Dienst (`gesehen`), erweiternd. **Die
    Abschlussschritte brauchen ihn nicht**: ein Gegenangebot geht nur aus
    „angeboten“ und „in Verhandlung“ (`handelsmatrix.go`), nach der Einigung
    ändert sich die Revision also nicht mehr, und einen geänderten Zustand
    weist die Matrix ohnehin ab. Der Dienst weist einen Abdruck dort ab,
    statt ihn still zu übergehen. **Was der Abdruck nicht fasst**: eine
    Stimme des Gegenübers zur selben Revision — sie verschiebt nichts, worüber
    entschieden wird; das Fenster vergleicht sie vor dem Zug.

## Verworfen

- **„Anbieten zählt als Zustimmung“ im Vertrag**: eine Kernänderung an der
  Matrix und an `PruefeErsteRevision`, erweiternd an allen vier Flächen —
  für dieselbe Wirkung, die ein Zug ohne sie hat.
- **Verbuchen von selbst, sobald die Übergabe des Gegenübers ankommt**: der
  Nachweis gälte für einen Stand, den kein Mensch gesehen hat (T10), und
  Bestand bewegte sich ohne Klick.
- **„Gegenseite bestätigt“ aus der Nachricht des Gegenübers**: der Draht
  trägt die Abschlussstimme nicht (#1637), und ein Beleg dafür liegt im
  Moment des Klicks nicht vor.
- **Nach jedem Schritt übertragen**: eine Stimme, die nach ihrem Stand
  hinausginge, hieße drüben „bekannt“ und ginge verloren.

## Was offen bleibt

- **„Gegenseite bestätigt“ hat im Moment des Klicks keinen Beleg** — der
  Mensch trägt ein, was er vom Gegenüber erfahren hat. Das war vor P13 so
  und bleibt so; ein Nachweis dafür wäre eine Änderung am Drahtformat.
- **Zwei Geräte** (#1998) und **VoiceOver** — Handtore.
- **Die Kommandozeile, MCP und REST gehen weiter Schritt für Schritt**: der
  Zug ist eine Bedienung des Fensters, keine neue Fähigkeit; jeder seiner
  Schritte steht dort wie bisher (§16, verträglich).
