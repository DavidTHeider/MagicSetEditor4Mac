# 0043 — Neun Phasen bis 0.2.0, und drei Entscheidungen, die sie geschnitten haben

**Stand:** angenommen
**Datum:** 2026-08-23
**Ergänzt:** [0031](0031-sieben-phasen-bis-zur-freigabe.md) — dort wurden sieben
Phasen bis zur Freigabe 0.1 geschnitten; diese neun schneiden den Weg zu 0.2.0
**Berührt:** [0039](0039-kein-dienst-auf-der-rueckschleife.md) — wird von Phase U
fortgeschrieben; [0003](0003-keine-notarisierung.md) — bleibt, ausdrücklich
geprüft; [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) — wird von
Phase S vervollständigt

## Lage

0.1.0 ist am 2026-08-22 veröffentlicht. Alle 375 Issues sind geschlossen; die
Liste ist leer. Am 2026-08-23 ist ein **mehrdimensionales Audit** über den
freigegebenen Stand gelaufen — elf Dimensionen, zwei unabhängige Durchgänge,
danach zusammengeführt (`docs/AUDIT-0.2.0.md` (Entwicklungsarchiv)).

**Das Zielbild hat sich geändert, und das ist der eigentliche Anlass.** 0.1
war eine Portierung, die benutzbar sein sollte. 0.2 soll eine **Magic-Zentrale
für macOS** sein: normale Mac-Nutzende kommen über die Oberfläche an, und
Fortgeschrittene bekommen dieselben wesentlichen Fähigkeiten über
Schnittstellen.

**Der Abstand dorthin ist vermessen, nicht geschätzt.** Neun gegen 63 — so
viele MCP-Werkzeuge stehen so vielen Wails-Bindungen gegenüber. Keine
Kommandozeile. Keine einzige Netzintegration für Kartendaten. Keine
Dateizuordnung im Finder. Keine automatische Sicherung vor einer Migration,
obwohl [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) sie
„notwendig" nennt. Und im veröffentlichten `.dmg` steht „Lizenz: MIT".

## Entscheidung

**Neun Phasen, R bis Z**, und die Reihenfolge ist die Rangfolge des Audits,
gruppiert nach der Naht, die eine Phase baut:

| Phase | Worum es geht | Hängt an |
| --- | --- | --- |
| **R** | Was ausgeliefert ist, stimmt | — |
| **S** | Der Bestand hat einen Rückweg | R |
| **T** | Ein Fachkern, drei Flächen — CLI und MCP auf Parität | S |
| **U** | Der Dienst auf der Rückschleife — REST | T |
| **V** | Fremde Dienste, gekapselt — Scryfall | T |
| **W** | Was MSE2 kann — Schlüsselwörter, Konsole, Zufallspack | T |
| **X** | Der Symboleditor | W |
| **Y** | macOS, wie man es erwartet | S |
| **Z** | Doku, Aktualisierung, Freigabe 0.2.0 | alle |

**R steht zuerst, und das ist der Kern dieses Vermerks.** Es behebt nichts
Neues — es macht wahr, was schon behauptet wird. Ein veröffentlichtes Artefakt
mit falscher Lizenzangabe, ein MCP-Server, der Idempotenz zusagt, wo er keine
hat, eine README, die eine Schnittstelle beschreibt, die es nicht mehr gibt:
das sind keine Lücken im Umfang, das sind **Zusagen ohne Deckung**. Auf einer
solchen Grundlage misst keine spätere Phase etwas Verlässliches.

**U steht nach T und nicht neben ihm.** Punkt 3 von
[0039](0039-kein-dienst-auf-der-rueckschleife.md) gilt unverändert: die Lücke
liegt im Umfang, nicht im Protokoll. Ein REST-Dienst neben einem
unvollständigen MCP wäre eine zweite unvollständige Fläche.

**Y hängt nur an S.** Es kann parallel zu T bis X laufen, sobald der Bestand
seinen Rückweg hat.

## Die drei Entscheidungen, die den Schnitt bestimmt haben

Sie sind am 2026-08-23 **vorgelegt und vom Auftraggeber entschieden** worden,
nicht vom Arbeitszweig. Das steht hier, weil zwei von ihnen einen bestehenden
Vermerk berühren und die dritte den Umfang der Iteration bestimmt.

### 1. REST wird gebaut

**Der Zielsatz für 0.2.0 nennt REST ausdrücklich.**
[0039](0039-kein-dienst-auf-der-rueckschleife.md) hat ihn am 2026-08-22
abgelehnt — mit vier Gründen, und der erste lautete: *es gibt keinen benannten
Abnehmer.* **Den gibt es jetzt.**

Der Vermerk hat diesen Fall selbst vorgesehen: „Käme ein benannter Abnehmer …
wäre die Rechnung neu aufzumachen. **Dieser Vermerk ist dann fortzuschreiben,
nicht zu umgehen.**" Genau das tut Phase U — in einem eigenen Vermerk, nicht
durch stilles Lockern eines Wächters.

**Der unabhängige Zweitblick war dagegen**, und sein Argument steht: CLI und
MCP liefern dieselbe Fähigkeit ohne lauschenden Port. Es ist nicht widerlegt
worden, es ist **überstimmt** — und der Preis, den es benennt, gilt weiter und
ist in Phase U einzulösen: Marken, Herkunftsprüfung, Ratenbegrenzung, und je
Endpunkt vier Testklassen bei **jeder künftigen Änderung**, nicht nur der
ersten.

### 2. Keine Notarisierung — 0003 bleibt

Beide Auditspalten führen den rauen Erststart und den fehlenden
Aktualisierungsweg unter den größten Hindernissen für normale Mac-Nutzende.
Eine Developer ID kostet 99 USD im Jahr.

**Die Entscheidung ist geprüft und bestätigt worden, nicht übergangen.**
Was folgt: der Aktualisierungsweg kommt in Phase Z trotzdem — aber als
**Hinweis auf eine neue Fassung mit Prüfsummenabgleich**, nicht als
selbsttätige Installation. Eine Updatekette ohne Signaturkette wäre selbst ein
Risiko, und ein Programm, das sich unsigniert selbst ersetzt, ist schlimmer
als eines, das sagt, dass es etwas Neues gibt.

### 3. Der Symboleditor kommt in 0.2.0

Er ist der einzige XL-Posten des Audits und bekommt deshalb eine eigene Phase.
`.mse-symbol` wird heute nur **gelesen**; mit Phase X bekommt das Format
erstmals einen Schreibweg.

**Was das kostet, gehört ausgesprochen:** 0.2.0 wird damit die umfangreichste
Iteration des Projekts. Das ist eine Entscheidung und kein schleichender
Zuwachs (§13).

## Verworfen

**0.2.0 als breite Sammlung neuer Integrationen beginnen.** Das war der
naheliegende Zuschnitt — Scryfall zuerst, weil es der sichtbarste Abstand zum
Zielbild ist. Beide Durchgänge des Audits sind unabhängig darauf gekommen,
dass es der falsche Anfang wäre: der einzige HTTP-Klient des Projekts hat kein
Zeitlimit, nennt sich nicht, kennt keine 429 und kopiert ein 2-GiB-Archiv im
Arbeitsspeicher. Eine erste Integration darauf erzeugte Schuld statt Wert.

**Den Symboleditor auf 0.3 vertagen**, wie
[0034](0034-paketerstellung-ist-keine-luecke.md) es mit der Paketerstellung
getan hat. Verworfen vom Auftraggeber. **Der Preis der Aufnahme**: eine Phase
mehr und eine spürbar längere Iteration.

**REST ablehnen und 0039 bestätigen.** Es war der Vorschlag des Zweitblicks
und die Empfehlung dieses Zweigs. Verworfen vom Auftraggeber. **Der Preis**:
eine dauerhaft zu verteidigende Fläche, und ein Wächter, der bislang eine
Zusage hielt, hält danach eine andere.

**Eine Phase je Befund.** Das Audit führt 38 eigene und 10 fremde Befunde;
achtundvierzig Phasen wären eine Liste, kein Plan. Gruppiert wird nach der
**Naht**, die eine Phase baut — deshalb steht der Speicherfehler des
Paketbezugs in R und nicht in V, obwohl er dort gebraucht wird.

## Folgen

- **`docs/ROADMAP.md` führt die neun Phasen**, und wie bisher **nur Offenes**.
- **Jede Phase endet mit dem Zweitblick durch Codex** — zwei Durchgänge, nicht
  mehr (§10) — und mit dem Nachschärfen der folgenden Phasen (§13).
- **Phase U schreibt [0039](0039-kein-dienst-auf-der-rueckschleife.md) fort**,
  bevor die erste Zeile entsteht. Ohne diesen Vermerk keine Lockerung von
  `internal/schichten/netz_test.go`.
- **Phase S vervollständigt [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)**
  um das, was heute fehlt: Regeln für die öffentliche Schnittstelle, für
  Abkündigung und für die Fassung, die eine Schnittstelle trägt.
- **[0003](0003-keine-notarisierung.md) bleibt unverändert** — geprüft am
  2026-08-23, mit dem Ergebnis, dass sie steht.

---

## Nachtrag vom 2026-08-23: der `README.md` ist vorgezogen

Er stand in Phase Z, bei den Handbüchern. Vorgezogen worden ist er, weil er die
**Eingangstür** ist: was jemand über diese Anwendung zuerst liest, beschrieb
Bauwege und Lizenz und kein einziges Mal, wie sie aussieht oder wie man mit ihr
arbeitet. Ein Audit, das elf Dimensionen misst, und daneben ein `README.md`, der
die Anwendung nicht zeigt, ist ein Widerspruch mit Publikum.

**Was das an Phase Z ändert:** sie schreibt ihn fort, statt ihn zu beginnen —
CLI, REST und Scryfall kommen hinzu, die Bildschirmabzüge werden nachgezogen,
wo sich die Oberfläche geändert hat. Die Handbücher bleiben dort, wo sie
standen.

**Und was es gekostet hat, gehört dazu:** die Bildschirmabzüge sind aus der
laufenden Anwendung, und beim Ablichten sind zwei Befunde aufgefallen, die
keine Testebene sieht —
#951 (die
Stilauswahl findet den installierten Stil nicht) und
#952 (die
Suche zeigt MSE-Auszeichnung roh). Das ist die Sichtprüfung aus §3, und sie
findet, was die Pipeline durchlässt.
