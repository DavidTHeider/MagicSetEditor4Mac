# 0101 — Was xAI belegt, und warum dieses Paket nichts abschreibt

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A10 von Phase 5.3
(#1449)

## Worum es geht

Der dritte Cloudanbieter. Das Issue stellt eine Bedingung, die kein anderes
gestellt hat:

> „ohne Annahmen aus OpenAI-API-Ähnlichkeiten"

Das klingt nach Vorsicht und ist eine **Warnung**. xAIs Schnittstelle sieht der
von OpenAI ähnlich genug, dass man sie für dieselbe hält — dieselben
`choices`, dasselbe `finish_reason`, dasselbe `usage`. Wer daraus schließt,
baut vier Fehler ein, und **zwei davon fallen erst beim Anbieter auf**, also
nachdem sie Geld gekostet haben.

Alles Folgende ist am **2026-09-12** unter <https://docs.x.ai/developers>
nachgelesen.

## 1 Die Modellliste heißt anders und sagt mehr

`GET /v1/language-models` — nicht `/v1/models`. Beide gibt es; nur der erste
nennt die **Modalitäten**.

```json
{"models": [{"id": "grok-4", "fingerprint": "fp_…", "version": "1.0.0",
  "aliases": ["grok-4-latest"], "input_modalities": ["text", "image"],
  "output_modalities": ["text"], "created": 1743724800}]}
```

Die Wurzel heißt **`models`** und nicht `data`. Der Tag ist eine
**Sekundenzahl** und kein RFC 3339. Und je Modell stehen Preise dabei.

**Damit belegt xAI zwei von drei Fähigkeiten** aus
[0097](0097-was-ueber-ein-ki-modell-feststeht.md): `output_modalities: text`
belegt **Textgenerierung**, `input_modalities: image` belegt **Bildlesen**.
Damit sagt es mehr als OpenAI, das gar nichts sagt
([0100](0100-was-openai-nicht-sagt.md)), und weniger als Anthropic.

**Die feste Form steht nicht dabei.** Die Schnittstelle kann sie
(`response_format.json_schema`), die Liste sagt über das einzelne Modell
nichts — und was nicht dasteht, ist nicht belegt. Dafür gibt es die **Messung**
aus A9; sie war für diesen Fall gebaut und brauchte hier keine Zeile.

**Ein Zweitname wird kein Eintrag.** `grok-4-latest` zeigt heute hierhin und
morgen woandershin; als eigener Eintrag stünde dieselbe Sache zweimal da, und
wer den Zweitnamen wählte, bekäme ohne Ankündigung ein anderes Modell. Er
reist im **Beleg** mit, zusammen mit dem `fingerprint` — das ist das Nächste,
was bisher ein Anbieter zu „was läuft hier wirklich" gesagt hat.

## 2 Der Ausgang kennt ein viertes Wort

`finish_reason` ist `stop`, `length`, **`end_turn`** oder `null`.

`end_turn` kennt Anthropic, OpenAI nicht. Wer die Liste von OpenAI übernähme,
hielte eine **fertige** Antwort für einen unbekannten Ausgang — und das sähe
aus wie ein Modell, das sich verweigert. Der Fehler fiele niemandem auf, weil
er wie eine Ablehnung aussieht, die es zu geben scheint.

**Und die Ablehnung wird vor dem Grund gelesen**: `refusal` steht neben
`finish_reason: stop`. Wer den Grund zuerst läse, hielte sie für ein Ergebnis.

## 3 Derselbe Anbieter hat zwei Bildformen

Das ist der Fund, um den sich das Issue dreht.

```
/v1/responses         {"type":"input_image","image_url":"data:…"}
/v1/chat/completions  {"type":"image_url","image_url":{"url":"data:…","detail":"high"}}
```

**Dieselbe Fähigkeit, derselbe Anbieter, zwei Gestalten.** Die erste ist die
von OpenAI. Wer sie an den zweiten Weg schickte, bekäme einen 400 — nach dem
Ruf.

**Und die Medientypen sind zwei, nicht vier**: „jpg/jpeg or png". Webp und GIF
nehmen OpenAI und Anthropic; hier wären sie ein Fehler, der Geld kostet.

## 4 Und `/v1/responses` bleibt draußen

**Weil es Antworten 30 Tage hält**: „New responses will be stored for 30 days
and then permanently deleted."

Bei OpenAI war das eine Zeile in der Anfrage (`store: false`). Hier ist es
**eine Wegentscheidung**: `/v1/chat/completions` legt nichts ab, und die feste
Form ist dort dokumentiert. Ein Weg, den man nicht geht, braucht keine
Abschaltung, die jemand vergessen kann.

**Die Positivliste führt deshalb zwei Adressen**, beide lesend.

## 5 Die Antwort nennt ihren Preis — und das tut sonst keiner

`usage.cost_in_usd_ticks`, und der Tick ist definiert: zehn Milliarden je
Dollar. Anthropic und OpenAI nennen Zeichen; wer daraus Geld machen will,
braucht einen Tarif und rechnet ihn nach — und rechnet falsch, sobald sich der
Tarif ändert.

**Der Wert bleibt ein Tick und wird kein Betrag.** Ein Zehnmilliardstel Dollar
ist kleiner als jede Währungseinheit; wer hier rundete, verlöre genau das,
wofür die Zahl da ist. Was das Kostenjournal daraus macht, ist eine Frage für
den Tag, an dem es diese Zahl benutzt — heute steht sie da, mit ihrer Einheit.

## 6 Die Fehlerart kommt aus dem Status, weil der Rumpf nicht dokumentiert ist

<https://docs.x.ai/developers/debugging> nennt die Statuscodes und ihre
Bedeutung — und **kein einziges Feld** des Fehlerrumpfs. Weder `error` noch
`code` noch `message`.

**Also wird tolerant gelesen und streng gedeutet**: was an bekannten Feldern
ankommt, reist als Grund mit; wo nichts ankommt, reist der **rohe Rumpf** mit.
Die Art kommt aus dem Status, denn der ist dokumentiert.

Eine angenommene Feldgestalt wäre erfunden (§16, Regel 1) — und der Schaden
wäre leise: eine leere Meldung im Protokoll, genau dann, wenn jemand sie
braucht.

**Zwei Statuscodes hat sonst keiner**: 415 („falscher Inhaltstyp") und 422
(„ein Feld hat ein ungültiges Format"). Beide sind Eingabefehler.

**Und ein 429 heißt hier genau eines**: zu schnell gefragt. Bei Anthropic sind
es zwei Dinge, bei OpenAI drei — dort kann derselbe Status leeres Guthaben
heißen, und das geht nicht vorüber. xAI nennt einen Grund, also steht hier ein
Konflikt und keine Erlaubnisfrage.

## 7 Gedrosselt wird, obwohl keine Zahl dasteht

Die Dokumentation nennt **keine** Grenze und **keine** Kopfzeilen über das
Kontingent. Sie sagt, dass ein 429 kommt, und rät zu wachsendem Warten.

**Eine geratene Grenze wäre schlimmer als eine vorsichtige** (§16, Regel 4).
Eine Anfrage je Sekunde, wie bei den anderen beiden.

**Und deshalb gibt es hier kein `grenzen.go`.** Bei OpenAI liest eines die
`x-ratelimit-*`-Kopfzeilen; hier gäbe es nichts zu lesen, und eine Datei, die
leere Werte zurückgibt, sähe aus wie eine Auskunft.

## 8 Der Bildvertrag ist gelesen und **nicht gebaut**

Das Issue verlangt beides, und es verlangt es in zwei Sätzen, die auseinander
zeigen:

> „Generation und Bearbeitung getrennt an A12 übergeben."

> „Aktuellen Bildedit-Vertrag prüfen: dokumentiertes JSON statt Übernahme
> eines multipart-Vertrags aus anderem Provider."

**Gebaut wird hier nichts davon** — der erste Satz ist der engere, und A12
(#1451) hat
den Umfang. **Gelesen ist es trotzdem**, denn genau darum geht es: A12 soll
eine **Auskunft** vorfinden und keine Vermutung.

Nachgelesen am 2026-09-12 unter
<https://docs.x.ai/developers/model-capabilities/images/editing>:

| | xAI | OpenAI |
| --- | --- | --- |
| Weg | `POST /v1/images/edits` | `POST /v1/images/edits` |
| Rumpf | **`application/json`** | `multipart/form-data` |
| Bild | `image: {type: "image_url", url: …}` | Dateifeld |
| Quelle | öffentliche Adresse, Daten-URI **oder** Datei-Kennung | Datei |

**Der Weg heißt gleich und der Rumpf ist ein anderer.** Wer den Vertrag von
OpenAI übernähme — dieselbe Adresse, dasselbe Verb —, bekäme einen Fehler, der
nach einem kaputten Bild aussieht und keiner ist. Das ist derselbe Fund wie in
Abschnitt 3, eine Ebene höher.

**Und eine Regel gehört mit übergeben**: was zurückkommt, ist eine **Adresse**.
Sie gehört nicht in ein Kartenfeld — weder als Adresse noch als Dateipfad. Ein
Kartenfeld, das auf einen fremden Wirt zeigt, ist ein Set, das ohne Netz
unvollständig ist; das Bild gehört in die Ablage, die es schon gibt
(`internal/app/bilddateinaht.go`), und die Karte zeigt dorthin. Das Issue sagt
es als Abnahme, und es steht hier, damit es A12 erreicht.

## Was das kostet

**Nichts davon ist gegen den echten Dienst gelaufen.** Ein Lauf brauchte ein
Konto mit Guthaben, und jeder Ruf kostete Geld — derselbe Grund wie bei
Anthropic und OpenAI. Es gibt deshalb **keine** Datei hinter `netzlauf`; eine
leere wäre eine Andeutung. Was daraus folgt, steht in `docs/HAERTUNG.md`.

**Und die Gestalt des Fehlerrumpfs bleibt ungeprüft**, weil sie nirgends
steht. Ändert sie sich, ändert sich hier nichts: der rohe Rumpf reist weiter
mit, und der Status bleibt der Status.

## Verworfen

**Ein gemeinsames Paket für OpenAI und xAI.** Die Gestalten sehen sich ähnlich,
und ein `internal/openaiartig` mit einem Schalter wäre kürzer gewesen. Er
hätte vier Unterschiede verwischt, von denen zwei erst beim Anbieter aufgefallen
wären — und beim vierten Anbieter wäre aus dem Schalter eine Tabelle geworden,
die niemand mehr liest. §2 sagt es: zwei Anbieter derselben Fähigkeit sind
zwei Blätter.

**`/v1/responses` mit einer Aufräumanfrage hinterher.** Es gibt sie; sie wäre
ein zweiter Ruf, der scheitern kann, und ein Fehler dabei ließe die Antwort
liegen, während die Oberfläche „nichts bleibt dort" sagt. Der Weg, den man
nicht geht, ist der einzige, der nichts hinterlässt.
