# 0063 — Was „Dienst an" heißt, und warum es keinen Schalter je Dienst gibt

**Datum:** 2026-09-01
**Anlass:** Phase 4.8 (#1321)
— drei weitere Kartendienste sollen dazukommen, und alle vier sollen sich
unter *Einstellungen* an- und abschalten lassen
**Stand:** entschieden, wird gebaut
**Baut auf:** [0056](0056-drei-weitere-kartendienste.md), das Transport,
SDK-Verzicht und Schlüsselablage entschieden hat

## Die Frage

[0056](0056-drei-weitere-kartendienste.md) sagt, **wie** ein Dienst gebaut
wird. Es sagt nicht, **was es bedeutet, ihn einzuschalten** — und genau das
muss vor der ersten Einstellung feststehen, sonst entsteht ein Schalter, dessen
Wirkung niemand beschreiben kann.

Naheliegend wäre: ein Häkchen je Dienst, vier Häkchen, fertig.

**Das beantwortet die Frage nicht, sondern verschiebt sie.** Wer zwei Dienste
anhakt, die beide eine Karte liefern können — wer antwortet dann?

## Was die Anwendung heute fragt

**Gemessen am 2026-09-01** an `internal/app/` (ohne Testdateien): elf Bindungen
tragen `Scryfall` im Namen, und sie stellen **sechs** verschiedene Fragen. Dazu
kommen zwei Wege, die Scryfall benutzen, ohne es im Namen zu führen.

| Frage | Wer sie heute stellt |
| --- | --- |
| **Karte suchen** — Text hinein, Treffer heraus | `ScryfallSuche` |
| **Einen Druck holen** und in ein Kartenfeld übernehmen | `ScryfallVorschlag`, `ScryfallUebernehmen`, `ScryfallKarteAnlegen` |
| **Ein Bild holen** | `ScryfallBild`, `ScryfallBildEinsetzen` |
| **Nachschlagen** — Regeltext, Manakosten | `ScryfallRegeln`, `ScryfallManakosten` |
| **Ein ganzes Set einlesen** | `ScryfallSatzEinlesen` |
| **Sammeldaten** — die Bulk-Datei | `ScryfallSammeldaten`, `ScryfallSammeldatenHolen` |
| **Preise** | `service/preispflege.go` — legt sie unter `domain.PreisquelleScryfall` ab |
| **Formatzulässigkeit** | `service/druckschub.go` — seit dem 2026-09-01 in Schüben zu 75 |

**Acht Fragen, nicht eine.** Und die vier Dienste beantworten nicht dieselben:

| | Suchen | Druck | Bild | Nachschlagen | Set | Sammeldaten | Preise | Format |
| --- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| **Scryfall** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **magicthegathering.io** | ✓ | ✓ | ✓ | ✓ | ✓ | — | — | ✓ (`legalities`) |
| **MTGJSON GraphQL** | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Scrydex** | ✓ | ✓ | ✓ | — | ✓ | — | ✓ | — |

**Die Zeilen sind aus den Endpunkten der Anbieter gelesen** (0056, abgerufen am
2026-08-30) und **nicht** gegen die echten Dienste gemessen — für zwei von
dreien geht das hier gar nicht. Ein Häkchen bedeutet „der Anbieter nennt diesen
Endpunkt", nicht „wir haben es gesehen". Das ist derselbe Unterschied, den 0056
für die Gewissheit macht, und er gilt auch für diese Tabelle.

## Die Entscheidung

### 0 · Die Sperre sitzt an der **Anfrage**, nicht am Bauen

**Nachgetragen am 2026-09-02**, weil der erste Anlauf es anders machte und
zwei Fehler hatte.

Naheliegend ist, beim **Verdrahten** zu prüfen: entsteht kein Klient, geht
auch keine Anfrage hinaus. Das ist wahr und trotzdem falsch:

- **Ein Schalter ändert sich.** Wer einen Dienst abschaltet, sah bis zum
  nächsten Start keine Wirkung — die Zusage galt für den Stand von vorhin.
- **Beim Start ist die Datenbank noch nicht migriert.** `scryfallStarten`
  läuft vor `storage.Migriere`; die Einstellung war dort gar nicht zu lesen,
  und alle vier Scryfall-Nähte fielen mit „no such table: settings" aus.

**Sie sitzt deshalb am Transport** (`netz.Klient.MitErlaubnis`) und wird vor
**jeder** Anfrage gefragt — vor der Drossel, damit niemand erst wartet, um
dann ein Nein zu bekommen.

**Und der Klient entsteht an genau einer Stelle.** `app.Scryfallklient` war
schon da; **acht** Stellen im Paket bauten ihn daneben wörtlich gleich noch
einmal. Solange das nur Adresse, Frist und Drossel setzte, war es Verdopplung
— mit einer Sperre wären es sieben Löcher gewesen.

### 1 · Eingeschaltet wird ein Dienst, gewählt wird je Frage

**Zwei Einstellungen und nicht eine**, weil es zwei verschiedene Aussagen sind:

- **„Dieser Dienst darf gefragt werden."** Ein Schalter je Dienst. Aus heißt:
  **kein einziger Aufruf geht hinaus** — auch nicht als Ausweichweg, auch nicht
  im Hintergrund. Das ist die Aussage, die jemand trifft, der einem Anbieter
  seine Anfragen nicht geben will, und sie muss ohne Einschränkung gelten,
  sonst ist sie nichts wert.
- **„Für diese Frage nimm jenen Dienst."** Eine Wahl je Frage, aus den
  Diensten, die eingeschaltet **und** für die Frage geeignet sind.

**Warum nicht eine Rangfolge, die von allein weiterrückt:** weil ein
Ausweichweg still ist. Wer „Bilder von Scrydex" einstellt und ein Bild von
Scryfall bekommt, weil Scrydex schwieg, hat eine Antwort aus einer Quelle, die
er nicht gewählt hat — und sieht es nicht. Was bei einem Bild harmlos aussieht,
ist bei einem **Preis** eine falsche Zahl mit einem falschen Namen daneben.

**Ein Ausweichweg ist damit nicht verboten, sondern eine Handlung**: der
gewählte Dienst schweigt, die Anwendung sagt es und bietet den anderen an. §16,
Regel 6 — nichts wird still übernommen.

### 2 · Was die Anwendung ohne jeden Dienst noch kann, ist die Probe

**Alle vier aus, und die Anwendung bleibt vollständig bedienbar** (§16, Regel
5). Karten entwerfen, Vorlagen laden, rendern, drucken, ausgeben, Decks und
Turniere — nichts davon fragt einen Dienst.

Was ohne Dienst **nicht** geht, ist genau das, was von draußen kommt: ein Druck
holen, ein Bild holen, ein Preis, eine Formatauskunft. Das ist keine Einbuße,
die man wegdefiniert — es ist die Grenze, die diese Anwendung ohnehin hat, und
sie wird **benannt**: was zuletzt bekannt war, bleibt stehen und sagt, wie alt
es ist.

**Das ist zugleich die Abnahme.** Ein Lauf mit allen vier Schaltern aus und
einem Testserver, der zählt, muss **null** Anfragen sehen.

### 3 · Der Schalter steht im Bestand, der Schlüssel nicht

Aus [0056](0056-drei-weitere-kartendienste.md), Entscheidung 3, und hier nur
zugespitzt: im Bestand steht die **Absicht** — welcher Dienst an ist, welcher
Dienst welche Frage beantwortet. Der Schlüssel steht im Schlüsselbund von
macOS.

**Und daraus folgt ein dritter Zustand, den es bisher nicht gab.** Ein Dienst
kann *an* sein und trotzdem nicht antworten können, weil sein Schlüssel fehlt.
Das ist **nicht** dasselbe wie „aus", und die Oberfläche muss es unterscheiden:

| Zustand | Heißt | Was die Anwendung tut |
| --- | --- | --- |
| **aus** | „frag ihn nicht" | fragt nicht, bietet ihn nirgends an |
| **an, Schlüssel fehlt** | „ich will, aber es geht nicht" | fragt nicht, **sagt warum**, und wo der Schlüssel hingehört |
| **an** | „frag ihn" | fragt |

Ein Dienst ohne Schlüssel, der trotzdem fragt und einen 401 bekommt, wäre die
schlechteste der drei Antworten: eine Anfrage, die niemandem nützt, und eine
Fehlermeldung, die nach einem Ausfall aussieht statt nach einer fehlenden
Angabe.

### 4 · Wo der Dienst herkommt, steht an dem, was er geliefert hat

Heute trägt ein Preis seine Quelle (`domain.Preisquelle`, und es gibt genau
einen Wert: `scryfall`). **Das wird die Regel für alles Bezogene**, nicht die
Ausnahme für Preise: ein übernommener Regeltext, ein eingesetztes Bild, eine
Formatauskunft sagen, von wem sie stammen.

**Warum das mehr ist als Ordnungsliebe:** sobald es vier Quellen gibt,
widersprechen sie einander. Zwei Preise für denselben Druck sind kein Fehler,
sondern zwei Aussagen — und ohne den Namen daneben ist es eine Zahl, deren
Herkunft niemand mehr feststellen kann. §16, Regel 6 sagt es schon: was von
außen in ein Kartenfeld will, trägt danach seine Herkunft.

### 5 · Die Namen der Dienste werden **nicht** übersetzt

„Scryfall", „magicthegathering.io", „MTGJSON", „Scrydex" sind Eigennamen (§9).
Was übersetzt wird, sind die **Fragen** — „Bilder", „Preise", „Formatprüfung" —
und die Zustände aus Entscheidung 3.

## Verworfen

- **Ein Häkchen je Dienst und sonst nichts.** Es beantwortet die Frage nicht,
  wer antwortet, wenn zwei können — es verschiebt sie auf die erste
  Fehlermeldung.
- **Eine Rangfolge mit stillem Ausweichen.** Der bequemste Entwurf und der
  gefährlichste: eine Antwort aus einer nicht gewählten Quelle sieht genauso
  aus wie eine aus der gewählten. Bei einem Preis ist das eine falsche Zahl mit
  einem falschen Namen.
- **Je Frage mehrere Dienste gleichzeitig fragen und vergleichen.** Reizvoll —
  und es vervierfacht die Anfragen an Anbieter, deren Grenzen wir einhalten
  wollen (§16, Regel 4). Wer zwei Preise sehen will, holt sie nacheinander und
  sieht beide Quellen; das leistet Entscheidung 4 schon.
- **Den Dienst je Frage aus der Verfügbarkeit erraten.** „Nimm den, der kann"
  ist eine Rangfolge ohne den Namen — mit demselben Fehler und ohne die
  Einstellung, an der man ihn sähe.
- **Alle vier von Anfang an eingeschaltet.** Ein Dienst, dessen Bedingungen
  jemand nicht gelesen hat, soll nicht in seinem Namen fragen. Scryfall bleibt
  an, weil es heute schon an ist und die Anwendung ohne es weniger könnte als
  gestern; die drei neuen fangen **aus** an.

## Nachtrag vom 2026-09-09

**Die Fragen kommen seit Werkstück A1 von Phase 5.3 aus dem Vertragskatalog**
([0082](0082-der-dienstkatalog.md)): eine Frage steht zur Wahl, sobald zwei
Dienste einen Vertrag dafür haben — Entscheidung 1 unverändert, nur die
Quelle ist eine. Die sieben Fragen, die heute nur Scryfall beantwortet,
führt der Katalog als Fähigkeiten mit Status, ohne Einstellung.
