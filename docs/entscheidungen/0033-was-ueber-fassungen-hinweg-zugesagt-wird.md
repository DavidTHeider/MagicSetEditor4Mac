# 0033 — Was über Fassungen hinweg zugesagt wird, und was nicht

**Stand:** angenommen
**Datum:** 2026-08-20
**Ergänzt:** [0011](0011-die-quelldatei-wird-mitgefuehrt.md) — die Byte-Zusage
gilt der **Datei**, nicht dem Schema
**Eingeschränkt:** [0189](0189-nur-die-grundlinie.md) — Satz 1 gilt seit
0.8.0 für jeden Bestand **ab 0.7.0**; ein älterer steigt über 0.7.0 auf

## Lage

#242 fragt, was
mit einem Bestand geschieht, der aus einer anderen Fassung stammt. Der Stand
war nie ausgesprochen, und das ist der ganze Anlass: was niemand zusagt, kann
auch niemand brechen — und niemand sich darauf verlassen.

**Nachgesehen am 2026-08-20, nicht erinnert** — und die Zahl der Migrationen am
2026-08-23 nachgezählt:

> Hier stand *elf*. Die zwölfte kam am **selben** 2026-08-20 dazu
> (#633), und
> die Zahl war damit binnen Stunden falsch. Nachgezählt:
> `ls internal/storage/migrationen/ | wc -l` → **12**. Eine Zahl in Fließtext,
> die niemand nachzählt, veraltet still; angestrichen vom Audit über 0.1.0
> (#900).

| Frage | Stand heute |
| --- | --- |
| Vorwärts migrieren | `storage.Migriere` läuft bei jedem Start; **zwölf** Migrationen, lückenlos geprüft (`luecken`) |
| Eine angewendete Migration, die sich geändert hat | wird abgewiesen — Prüfsummenwächter (`migrator.go`, `pruefsumme`) |
| Ein Bestand aus einer **neueren** Fassung | **wird abgewiesen** — eine verbuchte, aber nicht vorhandene Migration ist der Beleg (`migrator.go`, Zeilen 170–174), und `TestMigriereWeistEineVerschwundeneMigrationAb` fährt genau den Fall |
| Rückwärts migrieren | **gibt es nicht** — der Weg zurück führt über die Sicherung, und die gibt es seit dem 2026-08-23 (#939, #940) |
| Sicherung vor einer Migration | **gibt es** seit dem 2026-08-23 (#939) — selbsttätig, und nur wenn etwas aussteht |
| Ein Test, der einen **halb** migrierten Bestand weiterführt | **braucht es nicht mehr**: die Folge läuft in **einer** Transaktion (#938), einen halben Stand gibt es nicht. Gefahren wird stattdessen ein **abbrechendes** Upgrade (#957) |

## Entscheidung

**Vier Sätze. Drei sind die Zusage, der vierte ist eine Absicht** — und der
Unterschied steht hier, statt zwei Absätze später zurückgenommen zu werden
(angestrichen vom Nachlauf zum Zweitblick über Phase K).

1. **Vorwärts, immer.** Ein Bestand einer älteren Fassung wird beim Start
   weitermigriert. Das ist der einzige Weg, und er ist der Normalfall.

2. **Rückwärts, nie** — und dieser Satz ist **seit dem 2026-08-23 eine
   Zusage.** Es gibt keine Abwärtsmigration und es wird keine geben; wer eine
   ältere Fassung auf einem neueren Bestand öffnet, bekommt eine **Abweisung**.
   Der Weg zurück führt über die Sicherung, und beide Hälften stehen: vor einer
   ausstehenden Migration schreibt die Anwendung selbsttätig
   `bestand-vor-<Fassung>.db` (#939),
   und zurückspielen lässt sie sich im Fenster
   (#940).

   **Was der Weg nicht kann**, steht in `docs/RELEASE.md`: startet die
   Anwendung wegen einer abgebrochenen Migration gar nicht erst, gibt es kein
   Fenster und damit keinen Knopf. Dafür bleibt die Handarbeit dort stehen.

3. **Die Byte-Zusage gilt der Datei, nicht dem Schema.** `#48` sagt zu, dass
   ein geöffnetes und wieder geschriebenes `.mse-set` byteweise dasselbe ist —
   das hängt an `raw_blocks` und der mitgeführten Quelldatei
   ([0011](0011-die-quelldatei-wird-mitgefuehrt.md)), nicht daran, wie die
   Tabellen heißen. Eine Migration darf das Schema ändern, solange die
   Rundlaufzusage steht.

4. **Der Paketkatalog darf verlorengehen.** Er ist abgeleitet und wird bei
   jedem Öffnen der Paketseite abgeglichen. Wer ihn wiederherstellen will,
   öffnet die Seite.

**Warum Satz 2 am 2026-08-20 nur eine Absicht war**: `storage.anwenden` legte
**jede einzelne** Migration in eine eigene Transaktion, nicht die Folge.
Scheiterte die dritte von fünf, waren die ersten beiden festgeschrieben — der
Bestand stand dann auf einem Stand, den es als Fassung nie gab. Angestrichen
vom Zweitblick über Phase K.

> **Behoben am 2026-08-23**
> (#938). Die
> Folge läuft in **einer** Transaktion, und die Bestandsprüfung läuft darin —
> eine Abweisung ist damit eine Rücknahme statt eines festgeschriebenen
> Schadens. Den halben Stand gibt es nicht mehr.
>
> **Der Preis ist der verlorene Fortschritt**: vorher blieb erhalten, was schon
> lief, und der nächste Start setzte dort auf; jetzt beginnt er von vorn. Das
> ist der bessere Handel — „von vorn" ist ein Zustand, den es gibt, „zur
> Hälfte" nicht.
>
> **Und die Sicherung steht seit demselben Tag**
> (#939): vor
> einer ausstehenden Migration schreibt die Anwendung `bestand-vor-<Fassung>.db`
> neben den Bestand — aber nur, wenn schon ein Stand da ist. Ein frischer
> Bestand hat nichts zu sichern.
>
> **Satz 2 ist damit eingelöst**, soweit der Weg zurück eine Datei braucht —
> und seit dem 2026-08-23 auch, soweit er eine Bedienung braucht: das
> **Zurückspielen in der Anwendung** steht unter *Einstellungen → Bestand*
> (#940).
>
> **Eine Lücke bleibt, und sie ist benannt**: startet die Anwendung wegen einer
> abgebrochenen Migration gar nicht erst, gibt es kein Fenster und damit keinen
> Knopf (#962).
> Für diesen Fall steht der Weg in `docs/RELEASE.md` und verlangt eine
> Kommandozeile.

**Das macht die Sicherung nicht wünschenswert, sondern notwendig.**

**Gebaut worden ist sie in Phase S**
(#885), nicht
in Phase N — der Satz „Was daraus als Bau folgt, gehört Phase N" stand hier bis
zum 2026-08-23 und nannte eine Phase, die längst durch war. Angestrichen vom
Zweitblick über Phase S: **ein Vermerk, der von oben nach unten gelesen zwei
gegensätzliche Stände nennt, ist keine Entscheidung, sondern eine Schichtung.**

## Fortschreibung vom 2026-08-23 — die Schnittstelle

**Vier Sätze regeln Schema und Datei. Über die öffentliche Schnittstelle stand
hier nichts** — und 0.2.0 fügt zwei Flächen hinzu: eine Kommandozeile
(#886) und
REST (#887).
Ohne Regeln sind beide beim Erscheinen instabil
(#941, Befund
E-21 des Audits).

### 5 — Innen oder außen, und das ist entschieden

| Fläche | Gilt als | Heißt |
| --- | --- | --- |
| Die Wails-Bindungen | **innen** | Zwei Hälften **derselben** Auslieferung. Sie dürfen sich mit jeder Fassung ändern; gehalten wird nur, dass Fenster und Go dasselbe meinen (`internal/app/bindungen_intern_test.go`) |
| MCP, Kommandozeile, REST | **außen** | Dort steht ein Gegenüber, das **nicht** mitgeliefert wird — es kann eine andere Fassung haben, und niemand kann es mitändern |

**Ohne diese Zeile wäre jede Änderung an einer Wails-Bindung eine brechende.**
Es sind heute 64; sie einzufrieren hieße, die Oberfläche einzufrieren.

### 6 — Jede Änderung außen wird eingeordnet

**Verträglich** (etwas kommt hinzu, Vorhandenes bleibt), **erweiternd** (ein
neues Feld, das ein alter Klient ignorieren darf) oder **brechend**. Wer eine
Änderung nicht einordnen kann, hat die Fläche nicht verstanden — und soll sie
nicht ändern.

### 7 — Eine brechende Änderung braucht eine Hauptfassung

**0.x bindet daran nach SemVer nicht.** Dieses Projekt bindet sich trotzdem,
und der Grund ist derselbe wie bei allem hier: ein Gegenüber braucht die
Zusage, nicht die Ausrede, dass man sie noch nicht geben musste.

### 8 — Ein abgekündigter Name lebt eine Nebenfassung weiter

Und das Gegenüber erfährt es — **nicht erst dadurch, dass es bricht**. Wo das
Protokoll es hergibt, steht die Abkündigung dort (bei MCP in der `description`
des Werkzeugs); sonst in `CHANGELOG.md` unter einer eigenen Überschrift.

### 9 — Die Fähigkeiten sind maschinenlesbar

Ein Gegenüber soll **fragen** können, was diese Fassung kann. MCP hat
`tools/list`; die Kommandozeile bekommt es, und REST ebenso. **Erraten ist
keine Schnittstelle** — und eine Liste in einer README ist es auch nicht: sie
war am 2026-08-23 falsch
(#900).

### 10 — Die Fassung steht an einer Stelle

`app.Fassung`, gehalten gegen `wails.json` in **jedem** Lauf
(`internal/werkzeuge/fassung_test.go`) und im Freigabelauf gegen drei weitere
Orte. Ein hartkodiertes `0.0.0` war keine Fassung, sondern eine Lüge mit
Ziffern (#896).

### Was diese sechs Regeln **nicht** sind

**Kein Wächter.** Sie sind Entscheidungen; ob eine Änderung verträglich oder
brechend ist, entscheidet ein Mensch beim Lesen. Was ein Wächter halten kann,
hält er schon: die Fassung an einer Stelle, die Namen in der README, die
Vollständigkeit der Schreibwege.

**Und sie gelten erst, wenn es die Flächen gibt.** Für MCP gelten sie ab
sofort; für Kommandozeile und REST ab der Phase, die sie baut.

## Fortschreibung vom 2026-10-14 — die Rückgabewerte und die Fehlerform

**F03** (#2342):
ein Skript verzweigt nach dem Rückgabewert der Kommandozeile und nach dem
Code eines Fehlers, nicht nach einem Satz. Beides wird hier zugesagt.

**Die Rückgabewerte von `mse`** — **ihre Bedeutung zu ändern ist brechend**
(Regel 7) und braucht einen eigenen Übergangsvermerk; ein Wächter hält die
Konstanten gegen diese Tafel (`cmd/mse/rueckgabewerte_test.go`):

| Wert | Name | Bedeutung |
| --- | --- | --- |
| 0 | `CodeGut` | getan |
| 1 | `CodeFachlich` | der Aufruf war richtig, die Antwort ist nein — nicht gefunden, Konflikt, Validierung, belegt, voll, keine Erlaubnis |
| 2 | `CodeBedienung` | der Aufruf selbst stimmt nicht — unbekannter Befehl, fehlendes Argument |
| 2 | `CodeUngeklaert` | nur `mse vertraeglichkeit`: es lässt sich nicht sagen |
| 130 | `CodeAbgebrochen` | ⌃C oder SIGTERM — ein Skript soll danach nicht weiterprobieren |

**Die Fehlerform** — **erweiternd** (Regel 6): mit `--json` gibt die
Kommandozeile bei einem fachlichen Fehler **zusätzlich** `{"fehler": …}` auf
die Ausgabe, MCP trägt dieselbe Form in `structuredContent.fehler`; der Dienst
nennt sie seit jeher im Rumpf. **Code, Schlüssel, Werte und Vorgangskennung**
kommen an allen drei aus derselben Übersetzung wie im Fenster
(`app.Uebersetze`). Der Satz auf der Fehlerausgabe und im Text des
MCP-Ergebnisses bleibt; wer ihn las, liest ihn weiter. Dass jede Fehlerklasse
an jeder Fläche ihren Code nennt, verlangt `abweisungGleich`
(`internal/flaechen`).

**Und welche Wege mit 1 enden, ohne ein Fehler zu sein** (Zweitblick über
Phase 1 von 0.9.0, Befund 4): ein KI-Batch, in dem ein Ziel scheiterte, eine
unverträgliche Fassung, eine Pflege mit Konflikt oder Befund, ein Deck mit
fehlenden Karten, ein Rücksprung mit Unklarem, eine verweigerte Regelfreigabe.
Sie geben **ihr Ergebnis** aus — mit `--json` als JSON — und tragen kein
`fehler`; der Wert 1 sagt nur, dass es keines zum Weiterlaufen ist. **Das
Zurücksetzen** gibt mit `--json` seine ganze Antwort aus, und ein Fehler steht
darin unter `fehler` — dieselbe Form. Jeder andere Weg mit 1 meldet seinen
Fehler samt Form; `cmd/mse/fehlerjsonweg_test.go` prüft dafür **jeden Weg**
zu einer Rückgabe mit 1 — jede Konstante mit diesem Wert, die Zahl selbst und
jede Zuordnung, an jeder Stelle einer Rückgabe —, ob davor als ganze
Anweisung gemeldet wurde (Nachlauf über Phase 1 von 0.9.0, Befund 3), und
nennt die Ergebniswege mit Grund. **Ein Textweg** hinter
`if an.alsJSON { …; return … }` braucht keine Form.

## Verworfen

**Abwärtsmigrationen schreiben.** Sie sind die naheliegende Antwort und die
teuerste: jede Migration bräuchte eine zweite, jede zweite einen Test, und der
Fall, für den sie da wäre — jemand hat zwei Fassungen und öffnet beide —, ist
mit einer Sicherung billiger gelöst. **Der Preis der Ablehnung**: wer ohne
Sicherung auf eine ältere Fassung zurückgeht, kommt an seinen Bestand nicht
mehr heran. Deshalb ist die Sicherung kein Nebenpunkt, sondern die Bedingung,
unter der dieser Satz vertretbar ist — gebaut in Phase S.

**Eine Fassungsnummer in die Datenbank schreiben** (`PRAGMA user_version` oder
eine Tabellenzeile). Sie klingt sauberer als „eine verbuchte Migration, die es
nicht gibt", und sie sagt weniger: eine Nummer bezeugt nur, was jemand
hineingeschrieben hat, während die Buchführung bezeugt, was **wirklich
gelaufen** ist. Der bestehende Weg bleibt.

**Nichts zusagen und es beim heutigen Verhalten belassen.** Das war der Stand,
und er ist der Grund für dieses Issue: das Verhalten stimmt, aber niemand
konnte es nachlesen — und ein Verhalten, das nur im Quelltext steht, ist keine
Zusage, sondern ein Zufall, den der nächste Umbau nicht kennt.

## Folgen

- **`docs/ARCHITEKTUR.md` sagt, was im Bestand liegt** und was auf der Platte
  ([0029](0029-bilder-liegen-neben-dem-set.md), seit dem 2026-08-20 auch dort).
- **Die Sicherung steht seit dem 2026-08-23**
  (#939) —
  Phase N hatte sie nicht gebaut, obwohl dieser Vermerk sie ihr zugewiesen
  hatte. Ohne sie war Satz 2 nicht eingelöst, sondern nur behauptet; das Audit
  über 0.1.0 hat es als kritischen Befund geführt (E-23).
- **Ein neuer Migrationsname wird nie wiederverwendet.** Die Buchführung ist
  die Fassungsangabe; ein zweimal vergebener Name machte sie unbrauchbar.
