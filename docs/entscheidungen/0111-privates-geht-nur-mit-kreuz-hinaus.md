# 0111 — Privates geht nur mit Kreuz hinaus

**Datum:** 2026-09-13 · **Stand:** entschieden ·
**Schreibt fort:** [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) ·
**Anlass:** Werkstück A13.4 von Phase 5.3
(#1452)

## Die Frage

**Wann darf der eigene Kartenbestand an ein Modell gehen?**

Bis A13.4 lautete die Antwort: *wann immer der Recherchefall ihn liest.*
`domain.Recherchefall.Quellen()` nennt für `karten` und `aehnlich` die beiden
Korpora `oracle` **und** `eigenekarte`, und der Index suchte in beiden. Wer
eine Karte mit einem Modell verglich, schickte dabei seine eigenen mit — ohne
es zu sagen und ohne es zu sehen.

## Die Entscheidung

**Ein privater Korpus wird nur gesucht, wenn ein Mensch ihn ausdrücklich
ausgewählt hat.** Leer heißt: keiner.

Zwei Korpora sind privat — `eigenekarte` und `eigenertext` —, und **eine
unbekannte Quelle gilt als privat**. Dieselbe Fail-closed-Regel wie bei der
Datenkategorie in `service`: eine neue Quelle stillschweigend als öffentlich
zu führen wäre die Gestalt, in der sie unbemerkt hinausginge.

**Die Weglassung ist sichtbar.** Was ohne Freigabe außen bleibt, steht als
Einschränkung in der Antwort: *„Der Korpus »eigenekarte« ist nicht
freigegeben; in ihm wurde nicht gesucht."* Ohne diesen Satz sähe „nichts
gefunden" wie eine Aussage über den Bestand aus, und sie wäre eine über die
Freigabe.

## Warum das Kreuz und keine Einstellung

**Weil das Issue „aufgabenbezogen" sagt und nicht „einmal eingerichtet".**

Eine Einstellung unter *Einstellungen → KI* wäre bequemer und falsch: sie gilt
dann für jede Frage, auch für die, bei der niemand daran gedacht hat. Eine
Freigabe, an die man sich nicht erinnert, ist keine Entscheidung — sie ist
eine Vorgabe mit einem Häkchen davor.

**Und deshalb ist das Kästchen auch nicht vorbelegt.** Ein Schalter, der
aufgeräumt auf „an" stünde, hätte dieselbe Eigenschaft.

## Warum eine wirkungslose Freigabe ein Fehler ist

**Weil ein Kreuz ohne Wirkung eine Zusage ist, die niemand einlöst** (§10).

Wer „meine Karten sind freigegeben" ankreuzt und eine **Regel**recherche
fährt, hat etwas anderes entschieden, als er glaubt: `regeln` sucht nur im
Regelwerk. Eine stille Weglassung gäbe ihm eine Antwort, die aussieht, als sei
sein Kreuz berücksichtigt worden. Deshalb weist `Rechercheanfrage.Pruefe` eine
Freigabe ab, die der Fall gar nicht lesen kann — und ebenso eine für einen
**öffentlichen** Korpus, der ohnehin gesucht wird.

Am Fenster steht das Kästchen deshalb **nur dort, wo es wirkt**: beim
Kartenfall, bei der Ähnlichkeit, bei einer Balanceanalyse und bei einem
Vorschlag mit Belegfrage. Wer es woanders sähe, bekäme einen Fehler für eine
Entscheidung, die die Oberfläche ihm angeboten hat.

## Was es kostet

**Die Ähnlichkeitssuche findet ohne Kreuz keine einzige eigene Karte.** Das
ist der spürbarste Preis: „ähnliche Karten" hieß bisher „ähnliche Karten,
auch aus deinem Bestand", und heißt jetzt „ähnliche Karten bei Scryfall",
solange niemand das Kästchen setzt. Die Einschränkung in der Antwort sagt es,
und das Kästchen steht direkt über dem Knopf — aber es ist ein Klick mehr für
etwas, das vorher von selbst geschah.

**Das ist beabsichtigt.** Der Klick ist der Punkt: er ist die Stelle, an der
jemand weiß, dass seine eigenen Karten mitgehen — und bei einem Modell
außerhalb dieses Rechners heißt das, dass sie ihn verlassen.

**Und die MCP-Werkzeugbeschreibung musste berichtigt werden.**
`lesen.aehnlichekarten` sagte „findet Karten, die einer Karte AUS DIESEM
BESTAND ähneln" — die **Ausgangs**karte ist aus dem Bestand, die Treffer
waren es auch. Jetzt sagt sie, dass die eigenen Karten nur mit `freigegeben`
dazukommen. Ein Kopf, der mehr verspricht als sein Rumpf hält, ist schlimmer
als gar keiner (§10).

## Und was A11 damit bekommt

**Dieselben Belege wie eine Einschätzung — über denselben Dienst.**

Das Issue verlangt es: „A11 erhält dieselben Manifest-/Belegdaten wie A16;
keine zweite Prompt-/Retrievalpipeline." Ein Vorschlag bekam bis hierher den
getippten Auftrag und das Schema, sonst nichts. Jetzt darf er eine
**Belegfrage** tragen; wo sie steht, stellt [service.Kontextdienst] ein Paket
zusammen — dasselbe, das die Einschätzung bekommt, mit demselben Manifest,
derselben Zählweise und derselben Freigabe —, und die Belege stehen **an der
Antwort**.

**Leer heißt: gar nicht fragen.** „Schreibe einen Regeltext" ist eine Aufgabe
an das Modell und keine Frage an einen Index; wer trotzdem suchte, zahlte Zeit
für nichts.

**Und die Belege stehen neben den Feldern, nicht unter ihnen.** Das ist die
zweite Hälfte des Issues — „Beobachtungen und KI-Vermutungen auseinanderhalten":
die Belege sind, was **dasteht**, mit Quelle, Revision und Abrufdatum; die
Feldwerte sind, was ein Modell daraus **macht**, und sie werden bei der
Übernahme als `ki` verbucht (A11). Kein Wort der Antwort behauptet, dass ein
Wert aus einem Beleg folgt.

## Was verworfen wurde

**Eine Voreinstellung je Profil.** Sie wäre die bequeme Lösung und genau die,
die das Issue ausschließt — siehe oben.

**Die Freigabe am Recherchefall.** Dann bräuchte man zwei Fälle für dieselbe
Frage — „ähnlich" und „ähnlich mit eigenen" —, und die Zahl verdoppelte sich
mit jedem weiteren privaten Korpus.

**Eine stille Weglassung statt eines Fehlers.** Siehe oben: sie macht aus
einer falschen Entscheidung eine unsichtbare.
