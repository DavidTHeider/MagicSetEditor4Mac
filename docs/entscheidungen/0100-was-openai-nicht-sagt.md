# 0100 — Was OpenAI nicht sagt, und wie man es trotzdem weiß

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A9 von Phase 5.3
(#1448)

## Worum es geht

A7 hat den ersten Anbieter gebracht, A8 den ersten in der Cloud. Beide
**beschreiben ihre Modelle**: Ollama nennt `capabilities` unter `/api/show`,
Anthropic nennt `capabilities.structured_outputs.supported` an `/v1/models`.

OpenAI nennt **nichts**. Ein Modellobjekt trägt `id`, `object`, `created`,
`owned_by` und ein optionales `shutdown_date` — nachgelesen am 2026-09-12 unter
<https://developers.openai.com/api/docs/api-reference/models/list>. Kein
`capabilities`, kein Hinweis auf feste Form, Bilder oder Einbettung.

Damit steht die Regel aus
[0097](0097-was-ueber-ein-ki-modell-feststeht.md) gegen das Ziel des Issues:

> „Unbekannt bleibt unbekannt; Modelllisten garantieren keine Capability."

Ohne Beleg gehört ein Modell nicht in den Katalog. Ohne Katalogeintrag lässt
sich keines wählen. Der Anbieter wäre ein **Schalter, der nichts tut**.

Sieben Entscheidungen lösen das auf.

## 1 Aus der Modellliste entsteht kein Eintrag, sondern ein Befund

Je Modell einer, und er nennt den **Weg** und nicht bloß das Modell: wer liest
„gpt-probe: keine Fähigkeiten", sucht den Fehler bei sich; wer liest, dass
`/v1/models` keine nennt, weiß, dass es dort nichts zu suchen gibt.

**Aus dem Namen zu schließen wäre der Fehler, den §15 verbietet.** „`gpt-4o`
liest Bilder" ist eine Erinnerung, keine Auskunft — und ein Modellname ist ein
Bezeichner.

## 2 Eine Fähigkeit wird gemessen

**Der dritte Weg stand schon da, nur für etwas anderes.** A7 **misst** den
Einbettungsvertrag, statt ihn zu raten ([0098](0098-was-ollama-belegt-und-was-nicht.md)).
Hier ist es dasselbe: eine **Probe** schickt einen kleinen Auftrag mit festem
Schema; kommt etwas Schemakonformes zurück, sind **Textgenerierung** und
**strukturierte Ausgabe** belegt — mit dem Tag, an dem gemessen wurde, und mit
dem Modell, das wirklich geantwortet hat.

**Zwei Fähigkeiten und nicht eine**: eine schemakonforme Antwort zeigt beides.
Wer nur die zweite führte, ließe ein Modell ungeprüft, das gerade eben
geantwortet hat.

**Was eine Probe nicht belegt**, steht ausdrücklich daneben: nicht, dass das
Modell Bilder liest; nicht, dass es einbettet; und nicht, dass es morgen
dasselbe tut. `KIGeprueft.Datum` sagt, wie alt die Auskunft ist — nicht, dass
sie noch stimmt.

**Und eine abgeschnittene Antwort belegt gar nichts**, auch wenn das JSON darin
zufällig vollständig wäre: gemessen wird, dass das Modell **fertig wird**.

## 3 Eine Messung bleibt liegen — im Bestand, je Profil

**Weil sie Geld kostet.** Ein Ergebnis, das beim Schließen des Fensters
verschwindet, müsste beim nächsten Start bezahlt werden.

**Je Profil**, weil der Schlüssel je Profil liegt: wer mit einem anderen Zugang
fragt, fragt vielleicht ein anderes Konto mit anderen Modellen.

**Und je Profil, Anbieter, Alias und Fähigkeit genau eine.** Eine zweite ersetzt
die erste: was gestern galt, sagt über heute nichts, und zwei Daten
nebeneinander ließen offen, welches gilt.

## 4 Eine Messung geht durch dieselbe Sperre wie ein Vorschlag

Schalter, »nur lokal«, Journal, Freigabe. **Sie ist eine Anfrage und keine
Ausnahme** — wer sie daran vorbeiführte, hätte einen kostenpflichtigen Ruf ohne
Buchung, und genau das schließt A8 aus.

**Aber nicht durch das Auftragswerk.** `KIAuftragswerk.Anlegen` verlangt, dass
das Modell die Fähigkeit **schon belegt** hat — und genau das soll hier erst
gemessen werden. Wer die Probe dort hindurchführte, hätte einen Kreis: ohne
Beleg keine Messung, ohne Messung kein Beleg.

## 5 Der Ausgang steht an drei Stellen, nicht an einer

| Was der Dienst sagt | Ausgang |
| --- | --- |
| `status: completed`, ein `output_text` | fertig |
| `status: incomplete`, `incomplete_details.reason: max_output_tokens` | unvollständig |
| `status: incomplete`, `reason: content_filter` | abgelehnt |
| ein Inhaltsblock `{"type":"refusal"}` — **bei `status: completed`** | abgelehnt |
| `status: failed` mit `error` | ein Fehler, kein Ausgang |
| alles andere | unbekannt |

**Die vierte Zeile ist die, die man übersieht.** Eine Ablehnung steht weder in
`status` noch in `incomplete_details`; `status` ist dabei `completed`. Wer nur
ihn läse, hielte sie für ein Ergebnis — und die Schemaprüfung spräche danach
von fehlenden Feldern statt davon, dass das Modell nicht wollte.

## 6 Ein 429 heißt dreierlei, und nur eines geht vorüber

**Nachgelesen am 2026-09-12** unter
<https://developers.openai.com/api/docs/guides/error-codes>:

- `slow_down` — die Drossel. Sie trägt `Retry-After` und vergeht.
- `credit_balance_exhausted` — das Guthaben ist alle.
- `organization_spend_limit_exceeded` / `project_spend_limit_exceeded` — eine
  gesetzte Ausgabengrenze ist erreicht.

> „Retrying billing, spend, or quota errors won't restore API access."

Die letzten drei sind **Erlaubnisfragen** und keine Geduldsfragen (§8). Erkannt
werden sie an einer **Aufzählung** und nicht an einem Präfix: „enthält `spend`"
träfe auch einen künftigen Schlüssel, der etwas anderes heißt — und was dieses
Programm nicht kennt, soll als unbekannt durchkommen und nicht als erraten.

## 7 `store` steht ausdrücklich auf `false`

**Die Vorgabe der Schnittstelle ist `true`.** Ohne diese Zeile bliebe jede
Kartenanfrage beim Anbieter liegen — und `omitempty` machte aus `false` ein
Weglassen und damit ein `true`. Derselbe Fall wie `truncate` bei Ollama
([0098](0098-was-ollama-belegt-und-was-nicht.md), Entscheidung 5).

Die Hilfe sagt es dort, wo sie sagt, was hinausgeht.

## Was hier nicht steht

**Ob irgendetwas davon gegen den echten Dienst gelaufen ist.** Es ist nicht.
Ein Lauf brauchte ein Konto mit Guthaben, und jeder Ruf kostete Geld; der
Dienstvertrag trägt deshalb `Nachweis: nil`, und es gibt **keine** Datei hinter
dem Bautag `netzlauf` — eine leere wäre eine Andeutung. Was daraus folgt, steht
in [`../HAERTUNG.md`](../HAERTUNG.md).

## Verworfen

**Modelle aus `/v1/models` ungeprüft in den Katalog nehmen und im Fenster als
»ungeprüft« zeigen.** Der Katalog sagt seit A6: „hier stehen nur Modelle, die
geprüft wurden." Eine Liste mit einer Fußnote wäre die Modellliste eines
Anbieters, gegen die 0097 geschrieben wurde.

**Beim Vorschlag einen Namen tippen lassen.** A8 hat das Textfeld abgeschafft,
weil der Anbieter dazugehört und ein getippter Name erst beim Ruf zu einer
Meldung wird. Es zurückzuholen, hieße die Begründung von A8 zurückzunehmen,
ohne sie zu widerlegen.

**Die Probe bei jedem Aufschlagen des Katalogs fahren.** Sie kostet je Modell
Geld. Ein Katalog, der beim Ansehen abrechnet, ist eine Falle.

**Bilderzeugung in diesem Werkstück.** Die Responses-API kann sie über ein
Werkzeug; wohin ein Bild danach gehört, entscheidet A12
(#1451). Ein
Weg, dessen Ergebnis nirgends abgelegt werden kann, ist einer ohne Ziel (§10).
