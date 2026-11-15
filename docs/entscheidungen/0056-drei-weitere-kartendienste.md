# 0056 — Drei weitere Kartendienste, und drei verschiedene Grade von Gewissheit

**Datum:** 2026-08-30
**Anlass:** Wunsch für 0.4.0 — die Magic-API, die MTGJSON-GraphQL-API und
Scrydex sollen genauso benutzbar sein wie Scryfall, und jeder Dienst soll
sich abschalten lassen
**Stand:** entschieden, noch nicht gebaut — Phase 4.8

## Die Frage

Bis heute gibt es **einen** fremden Kartendienst: Scryfall. Drei weitere sollen
dazukommen, und die Anwendung soll unter *Einstellungen* sagen lassen, ob und
wie jeder benutzt wird — samt Zugangsdaten, wo es welche braucht.

Naheliegend wäre, alle drei gleich zu behandeln: ein Klient je Dienst, ein
Schalter je Klient, fertig.

**Das wäre falsch**, und zwar an der Stelle, die am meisten kostet: die drei
sind **nicht gleich prüfbar**. Wer sie gleich behandelt, gibt am Ende für alle
drei dieselbe Zusage ab — und für zwei davon kann er sie nicht halten.

## Was nachgelesen wurde

Alle drei am **2026-08-30** abgerufen (§16, Regel 1: die Bedingungen sind
nachgelesen, nicht erinnert).

| Dienst | Adresse | Zugang | Grenze |
| --- | --- | --- | --- |
| **magicthegathering.io** | `https://api.magicthegathering.io/v1` | **keiner** — kein Schlüssel, kein Konto | **1.000 Anfragen je Stunde**, gemessen (siehe unten); darüber **403** mit „Rate Limit Exceeded". Die Antwort führt `Ratelimit-Limit` und `Ratelimit-Remaining` |
| **MTGJSON GraphQL** | `https://graphql.mtgjson.com/` | `authorization: Bearer <Token>` — **nur für Patreon-Unterstützer**, im Beta | **1.000** je IP und Stunde, **500** je Token und Stunde |
| **Scrydex** | `https://api.scrydex.com` | **zwei** Kopfzeilen: `X-Api-Key` **und** `X-Team-ID`. Konto, Team und Schlüssel nötig | **100 Anfragen je Sekunde**, darüber 429 (nachgelesen am 2026-09-02); daneben **Kontingente**, die Geld kosten. Ohne Kopfzeilen sagt die Doku „heavily reduced rate limits" — **gemessen ist 401** (siehe unten) |

Dazu:

- Das **Go-SDK** `github.com/MagicTheGathering/mtg-sdk-go` (MIT) ist ein
  Wrapper um dieselbe HTTP-Schnittstelle.
- Scrydex führt **sechs** Spiele; hier zählt eines. Pokémon, Lorcana, Gundam,
  One Piece und Riftbound sind ausdrücklich außerhalb.

### Nachgemessen am 2026-09-02: die Doku sagt 5.000, der Dienst sagt 1.000

Hier stand **5.000 Anfragen je Stunde** — so steht es auf
`https://docs.magicthegathering.io/`, abgerufen am 2026-08-30 und am
2026-09-02 unverändert. Der Dienst selbst schickt am 2026-09-02:

```
ratelimit-limit: 1000
ratelimit-remaining: 997
```

**Der Dienst hat recht und die Seite nicht.** Gefunden hat es der Lauf hinter
dem Build-Tag `netzlauf` — er war rot, als `internal/mtgapi` noch die
dokumentierte Zahl prüfte, und die Drossel war damit **fünffach zu schnell**.

**Das ist genau der Grad von Gewissheit, den Entscheidung 4 beschreibt.** Für
diesen einen Dienst lässt sich nachsehen, ob er antwortet wie angenommen — und
beim ersten Nachsehen stimmte es nicht. Für MTGJSON und Scrydex gibt es diesen
Lauf nicht, und was dort in der Tabelle steht, ist die **Auskunft des
Anbieters** und keine Messung.

### Und bei Scrydex sagt die Doku, es ginge ohne Zugang weiter — es geht nicht

`https://scrydex.com/docs/getting-started/authentication` sagt am 2026-09-02:

> „Requests made without authentication headers will proceed, but with heavily
> reduced rate limits."

**Gemessen am selben Tag** gegen `api.scrydex.com`, ohne Kopfzeilen:

```
HTTP/2 401
{"error":{"message":"Invalid credentials","code":"INVALID_CREDENTIALS"}}
```

**Es geht nicht weiter.** Damit ist eine Anfrage ohne Zugang eine, deren
Antwort feststeht — `internal/scrydex` fragt gar nicht erst (§16, Regel 4).

**Zwei von drei Diensten haben heute anders geantwortet, als ihre Doku sagt.**
Bei magicthegathering.io war es die Grenze (5.000 gegen 1.000), hier die Frage,
ob es ohne Zugang überhaupt geht. Das ist kein Vorwurf an die Anbieter — es ist
der Grund, warum §16 Regel 1 „nachgelesen, nicht erinnert" sagt und warum das
**Nachmessen** danach noch einmal etwas anderes ist als das Nachlesen.

## Die Entscheidung

### 1 · Drei Klienten, ein Transport — und kein neuer Eintrag in der Netzgrenze

Jeder Dienst wird ein **Blatt** nach §16: Transport, fremdes DTO, Zuordnung —
in dieser Reihenfolge und ohne Vermischung. Der Transport ist in allen drei
Fällen `internal/netz`, und das ist keine Bequemlichkeit: `netz_test.go` sagt
über diesen Eintrag ausdrücklich, er solle die Liste **kürzer** machen und
nicht länger — „jeder weitere fremde Dienst hängt an ihm und kommt **nicht** in
diese Aufzählung".

**Damit bleibt die Netzgrenze bei ihren fünf Einträgen.** Ein Zuwachs dort wäre
das Zeichen, dass jemand am Transport vorbeigebaut hat.

### 2 · Das SDK wird **nicht** aufgenommen

`mtg-sdk-go` ist MIT-lizenziert und täte, was es soll. Es kommt trotzdem nicht
ins `go.mod`, und zwar aus drei Gründen:

- **Es bringt seinen eigenen Transport mit.** Damit läge die Drossel, die Frist
  und der Umgang mit dem 403 in einer fremden Bibliothek — genau das, was
  `internal/netz` an **einer** Stelle hält (#1002).
- **Es wäre der erste Netzverbraucher außerhalb der Aufzählung.** Ein
  eingeführtes Paket, das `net/http` mitbringt, ist für `netz_test.go`
  unsichtbar: der Wächter liest `internal/` und `cmd/`, nicht `.go/mod`.
- **Die Schnittstelle ist klein.** Sieben Endpunkte, eine Kopfzeile für die
  Seitenzahl, ein Statuscode für die Grenze.

**Was das kostet, gehört dazu**: jede Feldzuordnung von MTG-API-Karten auf
`domain.Card` schreiben wir selbst, und wenn der Anbieter ein Feld umbenennt,
merken wir es an unserem Vertragstest statt an einer neuen SDK-Fassung.

### 3 · Zugangsdaten kommen **nicht** in den Bestand

Die Einstellungen liegen seit [0010](https://github.com/DavidTHeider/MagicSetEditor4Mac/blob/v0.7.0/internal/storage/migrationen/0010_einstellungen.sql)
in der SQLite-Datei, und das ist für Sprache, Schriftgröße und Spaltenwahl
richtig. **Ein API-Schlüssel gehört dort nicht hin**: die Datei ist die, die
Nutzende sichern und weitergeben — „wer seine Datei kopiert, nimmt seine
Einstellungen mit" steht als Vorteil in derselben Migration, und für ein
Geheimnis ist es der Nachteil.

**Geheimnisse gehören in den Schlüsselbund von macOS.** Was im Bestand steht,
ist die **Absicht** — Dienst an, Dienst aus, welcher Dienst für welche Frage —,
nicht der Schlüssel.

**Und die Anwendung muss ohne ihn laufen.** §16, Regel 5: ein Dienst, dessen
Ausfall das Bearbeiten anhält, ist falsch eingebaut. Ein fehlender Schlüssel
ist derselbe Fall wie ein fehlendes Netz.

### 4 · Drei Dienste, **drei** Grade von Gewissheit — und sie werden benannt

Das ist der eigentliche Inhalt dieses Vermerks.

| Dienst | Was sich hier prüfen lässt | Was **nicht** |
| --- | --- | --- |
| **magicthegathering.io** | alles — kein Konto nötig. Ein Lauf hinter einem Build-Tag geht gegen den echten Dienst, wie `netzlauf` es für Scryfall tut | nichts |
| **MTGJSON GraphQL** | die Zuordnung gegen **eingecheckte, selbst geschriebene** Beispiele; die Gestalt der Anfrage gegen einen eigenen Testserver | **dass der echte Dienst so antwortet.** Es gibt hier kein Patreon-Token, und der Dienst ist im Beta |
| **Scrydex** | dasselbe — und dazu, dass beide Kopfzeilen gesetzt werden | **alles Übrige.** Es gibt kein Konto — und damit lässt sich die veröffentlichte Grenze von **100 Anfragen je Sekunde** hier nicht nachmessen. Sie steht in der Tabelle oben, nachgelesen am 2026-09-02 |

**Was daraus folgt, ist eine Regel für die Oberfläche**: ein Dienst, der hier
nie gegen sein Original gelaufen ist, sagt das dort, wo man ihn einschaltet.
Nicht im Kleingedruckten — an der Stelle. „Ungeprüft gegen den echten Dienst"
ist eine Auskunft, die jemand braucht, bevor er seinen Schlüssel einträgt.

**Hier stand bis zum 2026-09-02 „die Zahlen für die Drossel sind nicht
veröffentlicht"** — und zwanzig Zeilen darüber, in der Tabelle dieses
Vermerks, stand die Zahl. Angestrichen vom Zweitblick über Phase 4.8 (Befund
7). Ein Vermerk, der sich selbst widerspricht, ist an der Stelle wertlos, an
der jemand ihn wirklich befragt.

Eine Abnahme ist beobachtbares Verhalten. Für zwei dieser drei Dienste ist das
beobachtbare Verhalten **unsere Seite der Grenze** — und mehr zu behaupten wäre
genau der Fehler, den die Phasen 4.0 bis 4.2 siebzehnmal gefunden haben.

## Verworfen

- **Alle drei über ein gemeinsames „Kartendienst"-Interface.** Verlockend, weil
  drei Dienste dieselbe Frage beantworten. Falsch, weil sie es **verschieden**
  tun: Scryfall und die MTG-API liefern Drucke, MTGJSON liefert Setdaten samt
  Preisen, Scrydex liefert Preise und Bestandsdaten. Ein Interface, das alle
  vier trägt, trägt am Ende den kleinsten gemeinsamen Nenner — und die Fragen,
  für die man den Dienst geholt hat, passen nicht hinein.
  **Gemeinsam wird, was wirklich gemeinsam ist**: Drossel, Frist, Abbruch,
  Zwischenspeicher — und das ist `internal/netz`, das es schon gibt.
- **Den Schlüssel in einer Datei neben dem Bestand.** Eine Datei mit `0600`
  wäre besser als eine Spalte in der Datenbank und schlechter als der
  Schlüsselbund — und sie wäre eine dritte Stelle, an der Zustand liegt.
- **Scrydex weglassen, weil es nicht zu prüfen ist.** Der Wunsch ist
  ausdrücklich, dass künftige Nutzende ihre Konten mitbringen können. Was nicht
  geht, ist es **so** zu bauen, als wäre es geprüft.

## Nachtrag vom 2026-09-09

**Die Tabelle oben steht seit Werkstück A1 von Phase 5.3 im Quelltext** —
`internal/domain/dienstvertragweitere.go`, als Vertrag je Dienst und Frage,
mit Abrufdatum und Beleg, und ein Wächter hält Adresse und Endpunkte gegen
die Pakete. Was hier steht, bleibt die Begründung; was gilt, liest jede Fläche
aus dem Katalog ([0082](0082-der-dienstkatalog.md)). Entscheidung 4 hat damit
ihre Form: kein Häkchen »geprüft«, sondern der letzte Lauf mit Datum und
Fassung — und für MTGJSON und Scrydex keiner.
