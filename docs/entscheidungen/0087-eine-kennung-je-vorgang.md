# 0087 — Eine Kennung je Vorgang, und was sie nicht ist

**Stand:** angenommen · **Datum:** 2026-09-10 · **Werkstück:** O9 von Phase 5.2
(#1433)

## Der Anlass

Ein Fehler, den jemand sieht, und die Zeile, die ihn erklärt, sind zwei Dinge
an zwei Orten. §8 verlangt genau das: der Satz ist für Menschen, der **Grund**
fürs Protokoll (`domain.MitGrund`, [0075](0075-der-satz-und-sein-grund.md)) —
und damit gibt es keinen Weg vom einen zum anderen, außer über die Uhrzeit. Die
reicht nicht, sobald zwei Vorgänge nebeneinander laufen.

## Die Entscheidung

**Es gibt eine Kennung, nicht drei.** Acht Hexziffern, gewürfelt an der Grenze,
an der ein Ruf in diese Anwendung kommt. Sie liegt im `context`, steht in jeder
Protokollzeile dieses Rufs und reist mit seiner Antwort zurück.

Vier Grenzen, und nur dort wird gewürfelt:

| Fläche | Stelle |
| --- | --- |
| Fenster | `App.imProfil` — die erste Zeile jeder Bindung, und ein Wächter hält das — sowie `App.kontext` für die Bindungen ohne Kontextparameter |
| Kommandozeile | `cmd/mse/lauf.go`, vor der Verteilung |
| MCP | `cmd/mse-mcp/server.go`, in der Ausstattung **je Ruf** |
| Dienst | `cmd/mse-rest/dienst.go`, in `imProfil` **je Anfrage** |

**Zurück kommt sie je Fläche in ihrer Form**: am Fenster als Feld `vorgang` am
Fehler-DTO, an der Kommandozeile als Zeile `Vorgang: …` bei einem fachlichen
Fehlschlag, über MCP angehängt an den Fehlertext, am Dienst als Kopfzeile
`X-Vorgang` an **jeder** Antwort.

## Was verworfen wurde

**Ein Trace über Spans.** Das Issue verbietet „drei bedeutungsgleiche
Pflicht-IDs" ausdrücklich, und der Bedarf ist nicht belegt: dieser Prozess
ruft keine zweite Anwendung, deren Zeitscheiben jemand nebeneinanderlegen
müsste.

**Eine Kennung je Fehler statt je Vorgang.** Sie wäre billiger gewesen —
`Uebersetze` hätte sie selbst würfeln können, ohne einen Kontext. Sie hätte
aber nur den einen Fehler benannt und nicht den Ruf, in dem er entstand: zwei
Protokollzeilen desselben Vorgangs trügen verschiedene Kennungen.

**Ein `X-Request-Id`, das der Aufrufer setzen darf.** Ein Aufrufer, der zwei
Anfragen dieselbe Kennung gibt, macht das Protokoll unlesbar, und ein
Aufrufer, der eine fremde einträgt, mischt sich in eine andere Sitzung. Wer
eine eigene Korrelation braucht, hat seine eigene Zeile daneben.

## Der Preis, und er ist bezahlt

`app.Uebersetze` hat einen Kontext bekommen — **446 Rufstellen**. Das ist der
Grund, aus dem eine Kennung je Fehler verlockend war. Der Compiler hat jede
Stelle genannt; an **36** war kein Kontext in Reichweite, und die haben eine
zweite Tür bekommen: `app.Nutzerfehler` für einen Fehler, dessen ganze
Erklärung in seinem Satz steht.

**Diese zweite Tür ist die Stelle, an der diese Entscheidung erodieren kann**,
und sie steht deshalb unter einem Wächter: `Nutzerfehler` darf nur dort stehen,
wo **gar kein** `context.Context` in Reichweite ist
(`internal/werkzeuge/vorgangskennung_test.go`).

## Eine Stelle ohne Kennung, benannt

`app.Beispielsetliste` — die Liste der mitgelieferten Sets. Die drei Nähte, die
sie rufen, führen absichtlich keinen Kontext (sie lesen ein Verzeichnis im
Bündel), und an der Kommandozeile gibt es dafür keinen Vorgang: der Befehl
öffnet keinen Bestand. Eine Kennung wäre dort ein Griff ins Leere.

## Und was gezählt wird

Dazu gehört die zweite Hälfte des Werkstücks: **Metriken mit endlich vielen
Reihen.** Gezählt wird jede **geprüfte Wirkung** (O1,
[0078](0078-der-wirksame-akteur.md)) nach Fläche, Wirkungsklasse und Ausgang —
drei Aufzählungen, also höchstens **vierzig** Reihen, gleich wie viele Karten,
Sets oder Profile es gibt. Das ist keine Absicht des Aufrufers, sondern eine
Eigenschaft des Schlüsseltyps `service.Messpunkt`.

**Nicht auf der Kommandozeile.** Dort lebt ein Prozess genau einen Befehl lang;
ein Zähler über seine Lebenszeit wäre immer leer, und eine Fähigkeit, die nie
etwas sagt, ist keine. Der Grund steht in `docs/FAEHIGKEITEN.md`.

**Und kein Prometheus-Format.** Ein Textformat mit `# HELP`-Zeilen wäre eine
zweite Gestalt derselben Zahlen, und die Verpflichtung aus
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) gälte dann für beide.

## Und wer antwortet

Das Issue nennt „capabilities/meta" in einem Atemzug, und die beiden Hälften
gehören nicht zusammen. **Meta ist eine Auskunft**: Fassung, Plattform,
Go-Fassung, und die Fläche, über die der Ruf kam — sie steht in der Binärdatei
und braucht keinen Bestand, damit „welche Fassung bist du" auch dann eine
Antwort hat, wenn der Bestand nicht aufgeht.

**Capabilities sind eine Aushandlung** und gehören ins Kompatibilitätsmanifest
von O6 (#1430).
Eine fünfte Liste dessen, was eine Fläche kann, wäre eine, die veraltet — jede
sagt es heute schon selbst (`tools/list`, `mse hilfe`, `GET /grenzen`,
`GET /dienstkatalog`).

**Nicht am Fenster.** Dort steht dasselbe im Diagnosetext; eine zweite Bindung
daneben wäre eine ohne Aufrufer (§10).
