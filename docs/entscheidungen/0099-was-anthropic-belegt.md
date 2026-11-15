# 0099 — Was Anthropic belegt, und was es kostet

**Stand:** angenommen · **Datum:** 2026-09-12 · **Werkstück:** A8 von Phase 5.3
(#1447)

## Worum es geht

A7 ([0098](0098-was-ollama-belegt-und-was-nicht.md)) hat den ersten Anbieter
gebracht, und er rechnet auf **diesem** Rechner. A8 bringt den ersten, der es
nicht tut. Damit stellen sich drei Fragen, die bei Ollama gar nicht auftraten:

- **Was belegt ein fremder Katalog?** Anthropic nennt Fähigkeiten je Modell —
  und eine Angabe eines Anbieters ist etwas anderes als eine Messung.
- **Was kostet ein Ruf, und wer hat ihn erlaubt?** Ein lokales Modell kostet
  Strom. Dieses kostet Geld, je Zeichen.
- **Was heißt „zwei Anbieter" für die Auskunft?** A7 trug Einrichtung, Fassung
  und Adresse **flach** am Katalog. Mit zweien trägt das nicht mehr.

Acht Entscheidungen halten das auseinander.

## 1 Die Positivliste hat zwei Wege, und beide lesen nur

`internal/anthropic` kennt `/v1/models` und `/v1/messages`. Es gibt kein
`/v1/messages/batches` (das steht in A13), keine Dateien-API, kein Konto, das
dieses Programm anlegt.

**Ein Leser über den Quelltext hält sie**, wie bei Ollama, samt Gegenprobe.

## 2 Fähigkeiten kommen aus `capabilities`, und `null` ist ein Befund

Die Modellliste nennt je Modell `capabilities.structured_outputs.supported` und
`capabilities.image_input.supported` (nachgelesen am 2026-09-11). Was dort
`true` sagt, wird als [domain.KIGeprueft] geführt — mit dem Feldnamen im Beleg,
damit nachlesbar ist, **woher** die Aussage stammt.

**Fehlt `capabilities` ganz, entsteht kein Eintrag, sondern ein Befund.** Das
ist dieselbe Regel wie in 0097: „unbekannt bleibt unbekannt". Ein Modell ohne
Auskunft ist nicht eines, das nichts kann.

**Text ist die Ausnahme, und sie ist begründet**: dass ein Modell der
Messages-API Text erzeugt, ist die Schnittstelle selbst und steht deshalb
nicht unter `capabilities`. Der Beleg sagt genau das.

## 3 Ohne lesbares Datum kein Eintrag — und der Befund sagt es

[domain.KIGeprueft] verlangt neben dem Beleg ein **Datum**. Anthropic liefert
es als `created_at`; ist es unlesbar, bleibt es leer.

**Und dann entsteht kein Modell, sondern ein Befund.** Bis zum 2026-09-12 stand
das nur als Kommentar da — „dann fällt das Modell an `Validate` durch". Es fiel
durch: eine Schicht weiter oben, in [service.NeuerModellkatalog], und der weist
bei **einem** unvollständigen Eintrag den **ganzen** Katalog ab — auch die
Modelle der anderen Anbieter. Das ist genau der Fall, den §8 für fremde Daten
ausschließt: von 300 lesbaren Karten und zwei unlesbaren werden 300 geöffnet
und zwei benannt.

## 4 Die Herkunftskette bleibt leer, und das heißt „unbekannt"

Anthropic nennt weder Artefakt noch Revision noch Prüfsumme. Ein Modellname ist
kein Digest: was hinter `claude-opus-5` läuft, kann sich ändern, ohne dass der
Name es sagt.

**Also bleibt [domain.KIHerkunft] leer**, und das Fenster schreibt es aus:
„Herkunft unbekannt — es ist nicht belegt, welche Datei hier wirklich läuft."
Eine erfundene Kette wäre schlimmer als keine.

## 5 Vier Ausgänge, und drei davon sind kein Ergebnis

Das Issue verlangt: „unvollständig, Refusal und unbekannt getrennt behandeln."

| `stop_reason` | Ausgang | Was daraus folgt |
| --- | --- | --- |
| `end_turn`, `stop_sequence` | fertig | die Antwort gilt |
| `max_tokens`, `model_context_window_exceeded` | unvollständig | die Antwort gilt **und** sagt, dass sie abgeschnitten ist |
| `refusal` | abgelehnt | ein Fehler, und die Meldung sagt, dass das Modell nicht wollte |
| alles andere | unbekannt | ein Fehler, und die Meldung nennt den Grund wörtlich |

**Der letzte ist der wichtige.** Ein Grund, den dieses Programm nicht kennt, ist
kein Grund, eine Antwort für fertig zu halten — und ein neuer kommt bei jedem
Anbieter irgendwann.

## 6 Ein 429 sind zwei verschiedene Dinge, und ein 400 auch

**Nachgelesen am 2026-09-11 und am 2026-09-12** unter
<https://platform.claude.com/docs/en/api/rate-limits>:

- Eine **Drossel** antwortet 429 mit `retry-after`. Sie geht vorüber.
- Der **Monatsdeckel des Tarifs** antwortet ebenfalls 429, mit demselben
  `error.type` — und **ohne** `retry-after`. Er geht bis zum Monatsanfang nicht
  vorüber; erkennbar ist er allein an
  `error.details.error_code == "enforced_spend_limit_reached"`.
- Eine **selbst gesetzte** Ausgabengrenze antwortet dagegen **400** mit
  `invalid_request_error`, und ihre Meldung beginnt mit „You have reached your
  specified …". Derselbe Status wie ein Tippfehler im Schema, und eine völlig
  andere Abhilfe.

**Ein blindes Wiederholen auf 429 liefe bei der zweiten ins Leere**, und ein
`ErrValidation` bei der dritten schickte jemanden auf die Suche nach einem
Fehler, den es nicht gibt. Beide sind hier **Erlaubnisfragen** (§8).

**Dafür musste `internal/netz` etwas ändern**: es verwarf bei einem 429 den
Rumpf. Ohne ihn war der Unterschied nicht zu sehen.

## 7 Kein Ruf ohne Freigabe und ohne Journal

Ein Modell, das anderswo rechnet, kostet Geld. Deshalb:

- **Vor** dem Ruf reserviert das Kostenjournal aus O12
  (#1439) —
  wer erst nach der Antwort bucht, hat bei zwei gleichzeitigen Aufträgen
  zweimal die Grenze gelesen und keinmal gehalten.
- **Ohne den Nachweis eines Menschen** geht gar nichts. Das ist keine Erfindung
  dieses Werkstücks: `Kostenwache.Reservieren` verlangt ihn ohnehin.
- **Ohne Journal** geht gar nichts. Ein kostenpflichtiger Auftrag ohne Stelle,
  die mitzählt, ist der Fall, den das Issue ausschließt: „Unter Local-only,
  fehlendem Scope oder Budget entstehen null Anbieteranfragen."

**Und der Riegel prüft »nur lokal« vor dem Schalter.** Die Richtlinie schaltet
jeden `dienst.*` mit ab; wer nur den Schalter läse, meldete „Anthropic ist
abgeschaltet" und schickte jemanden an den falschen Schalter.

## 8 Der Katalog führt Anbieter als Liste — eine brechende Änderung

`lesen.ki.modelle` verliert `laufzeit` und `adresse` und bekommt `anbieter`:
eine Zeile je Anbieter mit Einrichtung, Fassung, Adresse und **seinen**
Befunden. `einrichtung` bleibt am Katalog und nennt den **besten** Stand über
alle — ein eingerichteter Anbieter macht die Anwendung brauchbar, auch wenn ein
zweiter aus ist.

**Brechend, und zwar absichtlich.** Die flachen Felder sprächen von einem von
zweien und sähen aus, als sprächen sie von beiden. Die Einordnung nach
[0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md) steht im PR.

## Was hier nicht steht

**Ob irgendetwas davon gegen den echten Dienst gelaufen ist.** Es ist nicht.
Ein Lauf brauchte ein Konto mit Guthaben, und jeder Ruf kostete Geld; der
Dienstvertrag trägt deshalb `Nachweis: nil`, und die Einstellungsseite sagt
es. Was daraus folgt, steht in [`../HAERTUNG.md`](../HAERTUNG.md) — dort ist es
die fünfte Ebene, und sie ist bei Anthropic leer.

## Verworfen

**Ein eigener Schalter `ki.anthropic`.** Es gibt `dienst.anthropic`, und er
beantwortet dieselbe Frage: darf er gefragt werden, welche Zugangsdaten
braucht er, wo liegen sie. Ein zweiter daneben wäre die Aufzählung, vor der der
Kopf von `einstellung.go` warnt — und die Einstellungsseite, der
Schlüsselbund, der Dienstkatalog und ihre Wächter hängen ohnehin an dieser
Familie.

**Den Tarif je Modell im Quelltext führen.** Er ändert sich, und eine Zahl, die
niemand nachzählt, veraltet still (§9). Der Vertrag nennt die **Spanne** mit
dem Tag, an dem sie gelesen wurde, und die Adresse, unter der die heutige
steht.

**Eine zweite Seite der Modellliste holen.** `has_more` wird als **Befund**
gemeldet und nicht durchgeblättert. Tausend Modelle gibt es heute nicht; käme
der Tag, stünde dort eine Schleife — und bis dahin sagt der Befund, dass sie
fehlt.
