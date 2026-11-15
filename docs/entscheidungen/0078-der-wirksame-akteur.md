# 0078 — Der wirksame Akteur: Rechte als Schnittmenge, Wirkung als Klasse

**Stand:** angenommen
**Datum:** 2026-09-08
**Anlass:** Werkstück O1 von Phase 5.2
(#1425) —
jede Fachoperation soll denselben aktuellen Nutzer-, Profil- und
Maschinenkontext prüfen, Rechte sollen eine **Schnittmenge** sein, und was
nach draußen wirkt, soll mehr brauchen als eine Rolle.

## Lage

**Was am 2026-09-08 galt, ist nachgelesen, nicht erinnert:**

| Frage | Antwort, am Stand `a5ce76c6` |
| --- | --- |
| Wer arbeitet | ein Zeiger im Bestand, getrennt vom gezeigten Profil ([0074](0074-wer-arbeitet-steht-fest.md)) |
| Was jemand darf | drei Rollen je Profil, `leser` < `mitarbeiter` < `verwalter`; der Eigner ist Verwalter ohne Zeile ([0070](0070-sichtbarkeit-und-rollen.md)) |
| Wo geprüft wird | an **zwei** Stellen: beim Wechsel in ein Profil und beim Setzen einer Freigabe. Sonst nirgends |
| Wie ein Leser gehalten wird | über eine zweite Verbindung mit `query_only` ([0073](0073-ein-leser-darf-wirklich-lesen.md)) — der **Treiber** weist das Schreiben ab |
| Was ein Token darf | `lesen` oder `schreiben`, nach HTTP-Methode ([0068](0068-technische-profile-und-zugriffstoken.md)) |
| Was die Flächen wissen | nichts voneinander: kein Aufruf trägt, ob er vom Fenster, von der Kommandozeile, über MCP oder über den Dienst kam |

**Und was das nicht hielt, ist gemessen** — am gestarteten Fenster und an
der gebauten Kommandozeile, bevor dieses Werkstück etwas änderte:

- **Ein Leser exportierte.** `SetSpeichern`, `BogenLegen`, `BildSchreiben`,
  `MTGJSONSchreiben` und `VorlageAusfuehren` schreiben **Dateien**, und die
  Platte fragt den Bestand nicht. Jeder dieser fünf Wege ging für einen
  Leser durch; beim Sichern lag die Kopie schon da, als der Treiber die
  Herkunftszeile abwies. `query_only` ist eine Eigenschaft der Verbindung,
  keine Autorisierung.
- **Drei Rufe, drei Stände.** Wer wissen wollte, wer wo als was arbeitet,
  fragte `WerArbeitet`, `GezeigtesProfil` und die Freigaben nacheinander —
  und konnte zwischen dem ersten und dem dritten einen Wechsel erwischen.
- **Ein selbstgesetztes Feld war der einzige Nachweis.** Für nichts, was
  heute gebaut ist, braucht es einen; für die KI-Aufträge mit Kostengrenze
  (Phase 5.3) und den Handel (Phase 5.6) gäbe es sonst nur ein `bestaetigt`,
  das der Aufrufer selbst setzt — und das beweist, dass er das Feld kennt.

## Entscheidung

### Fünf Wirkungsklassen, vier Dienstmodi, ein Akteur

Jede Fähigkeit hat **eine** Wirkungsklasse ([domain.Wirkung]):

| Klasse | Heißt | Mindestrolle im Profil |
| --- | --- | --- |
| `lokalLesen` | liest aus dem Bestand | `leser` |
| `lokalSchreiben` | schreibt in den Bestand **oder auf die Platte** | `mitarbeiter` |
| `externSenden` | schickt Daten an einen fremden Dienst | `mitarbeiter` |
| `externSchreiben` | ändert etwas bei einem fremden Dienst | `verwalter` **und ein Nachweis** |
| `verbindlich` | schließt etwas ab, das nicht zurückgeht — ein Kauf, ein Handel | `verwalter` **und ein Nachweis** |

Jeder Aufruf hat einen **Dienstmodus** ([domain.Dienstmodus]): `fenster`,
`kommandozeile`, `mcp` oder `dienst`. Die drei Flächen ohne Fenster tragen
ihn in ihren Wurzelkontext ein; das Fenster ist der Vorgabewert.

Aus beidem und dem Bestand entsteht je Aufruf ein **Akteur**
([domain.Akteur]): Nutzer, Profil, Eigner, wirksame Rolle, Token, Modus und
die **Revision** der Rechte. Ob er eine Wirkung haben darf, sagt
`Akteur.Darf` — als **Schnittmenge**:

- die **Rolle** reicht bis zu ihrer Klasse (oben), und `leser` heißt zudem
  `NurLesend`;
- ein **Token** reicht nie weiter als seine Rolle, und **nie nach draußen**
  — `lesen` bis `lokalLesen`, `schreiben` bis `lokalSchreiben`;
- was einen **Nachweis** braucht, bekommt keiner durch die Rolle allein.

**Das ist die Matrix aus dem Issue**, und sie steht ausgeschrieben in
[`docs/AUTORISIERUNG.md`](../AUTORISIERUNG.md): **jede** Fähigkeit mit Träger,
Klasse, Mindestrolle und der Stelle, an der das durchgesetzt wird. **Hier stand
„164"** — eine Zahl vom Tag dieses Vermerks, die mit jeder neuen Fähigkeit
falscher wurde (Hinweis 14 des Zweitblicks über Phase 5.2). Gezählt wird sie in
der Tabelle dort, und ein Wächter hält sie.
Ein Wächter hält sie gegen `FAEHIGKEITEN.md` und gegen die Klassen im Code.

### Ein Riegel vor jedem Weg, der Dateien schreibt

`query_only` bleibt — es hält den Bestand, und es hält ihn auch dann, wenn
jemand einen Riegel vergisst. **Daneben** fragt jeder Dienst mit Dateiwirkung
einen Prüfer ([service.Pruefer]), bevor er etwas anlegt: der Setschreiber,
die Bildausfuhr, der Bogenleger, der Exporteur — und das Fenster bei der
Exportvorlage, die es selbst schreibt. Die Wurzeln reichen die
[service.Autorisierung] hinein, für das Fenster wie für die drei Flächen
ohne Fenster.

**Das Nein kommt vor jeder anderen Antwort**, und es hinterlässt nichts:
belegt am gestarteten Fenster (`internal/app/leserausfuhr_test.go`) und an
der gebauten Kommandozeile (`internal/flaechen/leserausfuhr_test.go`).

### Der Arbeitskontext: ein Zug, ein Stand

`Arbeitskontext` sagt an allen vier Flächen, wer wo als was arbeitet —
Nutzer, Profil, Eigner, Rolle, `nurLesend`, Modus und Revision, aus **einem**
Kontext. Er steht an der Stelle von `Leseart`: die sagte nur, ob nur gelesen
wird, und nur am Fenster.

**Die Revision steigt in der Ablage, nicht im Aufrufer.** Auslöser an
`bestand` (die beiden Zeiger) und `profilfreigaben` zählen sie in derselben
Transaktion hoch, in der sich etwas ändert — auch bei der Kaskade, die ein
gelöschtes Profil hinterlässt und die kein Aufrufer sieht. Zwei Antworten mit
derselben Revision beschreiben denselben Stand; das Fenster (Werkstück U4)
liest daran ab, ob es neu laden muss.

**Laufende Arbeit behält ihren Ursprung, neue Arbeit sieht den Entzug**: der
Akteur entsteht **je Aufruf** aus dem Bestand, nicht einmal beim Start. Eine
Rolle, die zwischen zwei Rufen entzogen wird, ist beim zweiten weg.

### Der Nachweis: vorab, eng, einmal

Ein [domain.Nachweis] bindet Nutzer, Profil, Zweck, Ziel, Revision und — wo
es um Geld geht — eine Kostengrenze; er hat eine Frist von höchstens
24 Stunden und wird **genau einmal** eingelöst. Ausgestellt wird er nur am
**Fenster**, von einem Menschen mit mindestens `mitarbeiter` und ohne Token.
Eingelöst wird er atomar in der Tafel `freigabenachweise`: `UPDATE … WHERE
eingeloest = 0` trifft einmal, gleich wie viele zugleich kommen.

**Ein wiederverwendeter, abgelaufener oder auf ein anderes Ziel, einen
anderen Zweck, eine andere Revision oder einen anderen Menschen gebundener
Nachweis erlaubt nichts** — jeder dieser Fälle ist ein Test in
`internal/domain/nachweis_test.go` und `internal/service/nachweise_test.go`.
Wer ihn braucht, sind die Phasen 5.3 und 5.6; hier steht die Form, damit
sie dort nicht erst entsteht, wenn der erste Kauf schon läuft.

## Verworfen

- **Eine zweite, universelle Akteurarchitektur** mit eigener Sitzung,
  eigenen Rechten je Ressource und eigenem Speicher. Das Issue warnt selbst
  davor („keine zweite universelle Actorarchitektur ohne Bedarf"), und der
  Bedarf ist konkret: fünf Klassen, drei Rollen, zwei Tokenumfänge. Was es
  gibt — Zeiger, Freigaben, Token, `query_only` — bleibt und wird zu einem
  Akteur **zusammengesetzt**, nicht ersetzt.
- **Die Revision im Aufrufer zählen.** Vier Wege schreiben die Zeiger, zwei
  die Freigaben, und die Kaskade ist ein siebter; ein Zähler, den jeder Weg
  selbst erhöht, wäre an dem Tag falsch, an dem einer es vergisst — und ein
  Nachweis gälte dann still weiter.
- **Den Nachweis im Arbeitsspeicher halten.** Ein Handel wartet auf einen
  Partner, und dazwischen darf der Prozess enden.
- **Eine Tabelle je Bindung, in der jede Bindung ihre Klasse trägt, und ein
  Rahmen, der sie vor dem Ruf prüft.** Am 2026-09-08 waren es 244 Bindungen,
  191 Werkzeuge, 137 Befehle, 191 Endpunkte — vier Tabellen, und keine davon liegt an der
  Stelle, an der die Wirkung entsteht. Der Riegel sitzt im **Dienst**, weil
  alle vier Flächen dieselben Dienste rufen (§16).
- **`Leseart` neben `Arbeitskontext` stehen lassen.** Zwei Bindungen für
  dieselbe Frage sind zwei Stände.

## Folgen

- **Ein Leser exportiert nicht mehr** — an keiner Fläche, in keine Datei.
  Das ist eine Verhaltensänderung und steht im Änderungsprotokoll.
- **Was `externSenden` ist, steht in der Matrix**, und die Fähigkeiten
  dieser Klasse tragen denselben Riegel: ein Leser sucht nicht bei Scryfall
  und holt keine Preise. Was er sähe, wäre öffentlich — was er **schickte**,
  wäre der Name einer Karte aus einem Profil, das ihm nicht gehört.
- **Nichts braucht heute einen Nachweis.** Die Klassen `externSchreiben` und
  `verbindlich` sind besetzt, aber leer; die erste Fähigkeit darin kommt mit
  Phase 5.3. Bis dahin belegen Tests die Form, nicht ein Betrieb.
- **Die Wails-Naht ändert sich**: `Leseart` fällt, `Arbeitskontext` kommt.
  Sie ist von den Zusagen über Fassungen hinweg ausgenommen ([0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md));
  MCP, Kommandozeile und Dienst bekommen je eine **erweiternde** Änderung.
- **Zwei neue Tafeln** — `freigabenachweise` und `freigabenrevision` — in
  Wanderung 0047.
