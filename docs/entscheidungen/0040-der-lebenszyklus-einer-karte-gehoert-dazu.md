# 0040 — Der Lebenszyklus einer Karte gehört zu den MCP-Werkzeugen

**Stand:** angenommen
**Datum:** 2026-08-22
**Schreibt fort:** [0010](0010-mcp-als-eigener-prozess.md)
**Berührt:** [0039](0039-kein-dienst-auf-der-rueckschleife.md) — er stützt sich
auf die Lücke, die dieser Vermerk schließt

## Lage

[0010](0010-mcp-als-eigener-prozess.md) hat den Umfang der MCP-Werkzeuge
ausdrücklich beschränkt:

> …keines, das eine Karte löscht, und keines, das mehrere Karten auf einmal
> ändert — obwohl der Bestand das könnte. Der Grund steht in der Frage selbst:
> „ein Werkzeug, das ein ganzes Set überschreibt, gehört nicht in dieselbe
> Reihe wie eines, das ein Feld setzt."

Ein Wächter hielt das: `TestEsGibtGenauEinSchreibendesWerkzeug`.

**Und [0039](0039-kein-dienst-auf-der-rueckschleife.md) hat sich am 2026-08-22
genau darauf gestützt**, um gegen einen zweiten Zugang zu entscheiden: fünf
Werkzeuge, davon eines schreibendes; die Lücke liege im **Umfang** und nicht im
Protokoll. Dieser Vermerk schließt sie.

## Entscheidung

**Der Lebenszyklus einer Karte kommt dazu** — anlegen, kopieren, löschen,
umordnen. Vier Werkzeuge, jedes für **eine** benannte Karte (`tauschen` für die
zwei, die es nennt).

**Die andere Hälfte von 0010 bleibt:** kein Werkzeug überschreibt ein Set, und
keines ändert mehrere Karten auf einmal. Der `Massenschreiber` des Bestands
bleibt außen vor.

## Was sich geändert hat, und es ist nicht die Meinung

0010 nennt zwei Gründe. **Der erste steht unverändert**, der zweite nicht mehr.

### Der Maßstab — unverändert

„Ein ganzes Set" gegen „ein Feld": eine **Karte** liegt dazwischen, und
`AGENTS.md` §2 sagt, wofür es dieses Paket gibt — „Werkzeuge, mit denen ein
Sprachmodell **Karten** liest und ändert". Nicht Felder, nicht Sets.

Die Zeile aus 0010 schließt das Set aus. Sie schließt die Karte nicht ein und
nicht aus; sie war zu dem Zeitpunkt einfach nicht gebaut.

### Die Rücknahme — hier hat sich der Boden verschoben

0010 sagt: „jede Änderung gibt den vorigen Wert zurück, und wer ihn hat, kann
sie umkehren." Für ein Feld geht das. **Für eine gelöschte Karte ging es am
2026-08-15 nicht** — es gab keinen Weg, sie zurückzubringen.

**Am 2026-08-20 gibt es ihn**
(#621): das
Löschen schreibt einen eigenen Verlaufsschritt, und die Rücknahme bringt die
Karte **mit derselben Kennung** zurück. Fünf Tage nach 0010, und deshalb steht
es dort nicht.

## Was das kostet, und es gehört ausgesprochen

**Die Rücknahme des Löschens ist nicht mehr in der Antwort enthalten.** Bei
einem Feld gibt das Werkzeug den Vorwert zurück, und der Aufrufer kann die
Änderung allein umkehren. Bei einer gelöschten Karte kann er das nicht: der
Rückweg führt über den Verlauf, und den bedient das **Fenster**.

Das ist eine schwächere Zusage als die von 0010, und sie steht deshalb in der
Antwort des Werkzeugs im Klartext:

> über den Verlauf des Fensters (⌘Z) — die Karte kommt mit derselben Kennung
> zurück

**Wer das nicht will, benutzt das Werkzeug nicht.** Ein Gegenüber, das vor
zerstörenden Werkzeugen nachfragt, erkennt es an drei Stellen — am `schreiben.`
im Namen, am `ÄNDERT` im Text und an `destructiveHint`.

**Die drei anderen sind selbst umkehrbar**: eine angelegte oder kopierte Karte
nimmt `schreiben.karte.loeschen` wieder weg, und ein Tausch kehrt sich durch
denselben Aufruf um. Ihre Antwort nennt den Weg.

## Was der Wächter danach hält

`TestEsGibtGenauEinSchreibendesWerkzeug` prüfte eine Zahl. Er prüft jetzt die
**Zusagen**, die bleiben:

- kein Werkzeug ändert mehr als die Karten, die es **nennt**
- jedes schreibende Werkzeug sagt in seiner Antwort, **wie man es zurücknimmt**
- die Reihen bleiben getrennt — `lesen.` und `schreiben.` im Namen, `ÄNDERT` im
  Text, `destructiveHint` gesetzt

**Eine Zahl ist keine Zusage.** Sie war eine, solange die Zahl eins war und der
Grund dafür in 0010 stand; sobald der Grund sich verschiebt, prüft eine Zahl nur
noch sich selbst.
