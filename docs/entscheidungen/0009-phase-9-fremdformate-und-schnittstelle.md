# 0009 — Fremdformate und Schnittstelle als Phase 9

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

Der Entwurf vom 2026-08-09
(`superpowers/specs/2026-08-09-fremdvorlagen-und-json-entwurf.md` (Entwicklungsarchiv))
hat fünf Issues eröffnet und für drei davon ausdrücklich **keine** Zuordnung
getroffen:

> #141 bis #143 sind neuer Umfang, und wo sie hingehören, ist eine Entscheidung,
> die in der Roadmap sichtbar werden muss (§13) — nicht eine, die dieser Entwurf
> trifft.

Seither stehen sie in `docs/ROADMAP.md` unter „Noch ohne Phase". Das ist kein
Zustand, den man lange hält: was ohne Phase geführt wird, wird nicht gebaut.

Am **2026-08-15** hat der Auftraggeber angeordnet, den Entwurf umzusetzen. Damit
sind zwei Fragen fällig, nicht eine:

1. Welche Phase trägt #141, #142 und #143?
2. In welcher Reihenfolge wird gebaut? Phase 6 ist offen, Phase 7 auch.

## Entscheidung

**#141 bis #143 bilden Phase 9 — Fremdformate und Schnittstelle.** Eine neunte
Phase, keine stillschweigende Vergrößerung einer bestehenden.

| Issue | Was |
| --- | --- |
| #141 | MTG JSON lesen |
| #142 | MTG JSON schreiben, mit Rundlauf |
| #143 | MCP-Server: Karten lesen und ändern |

**Die Nummer sagt, wann die Phase geschnitten wurde, nicht wann sie gebaut
wird.** Gebaut wird sie **jetzt**, zusammen mit #139 und #140 aus Phase 7 — weil
der Auftraggeber den Entwurf als Ganzes angeordnet hat und weil #139/#140 und
#141–#143 dieselbe Frage von zwei Seiten beantworten: wie fremdes Material
hereinkommt.

**Phase 8 bleibt die letzte.** Freigegeben wird, wenn alle **neun** Phasen durch
sind — die Zusage aus `AGENTS.md` §12 ändert nur ihre Zahl, nicht ihren Sinn.

## Verworfen

**Phase 6 — Schreiben und Export.** Die naheliegendste, und deshalb die, gegen
die am meisten spricht. Phase 6 hat ein Versprechen, das sie ausmacht: was
gelesen wurde, kommt **Byte für Byte** wieder heraus
(#48). MTG JSON
kann das nicht halten und soll es nicht — #142 sagt es selbst: es beschreibt
Karten, nicht ihre Darstellung, und Stil, Styling und Rohblöcke haben dort
keinen Platz. Zwei Rundlaufversprechen unterschiedlicher Stärke in einer Phase
machen das starke unkenntlich, und das starke ist das, an dem ein Set hängt, das
jemand in MSE2 zurückbringt.

**Phase 7 — Paketverwalter und Saat.** Trägt den Bezug von Paketen. Ein
Kartenformat ist kein Paket, und ein MCP-Server erst recht nicht. Dass #139 und
#140 aus demselben Entwurf stammen, macht #141 bis #143 nicht zu Paketarbeit.

**Phase 8 — Freigabe.** Ein Server, der dort einzöge, wäre Umfang, den die
Freigabe nicht mehr prüfen kann: Phase 8 ist die Phase, in der nichts Neues mehr
entsteht.

**Weiter ohne Phase führen.** Das ist die Möglichkeit, die es schon gab, und sie
hat sechs Tage lang nichts entschieden. Eine Zeile „offen" ist keine
Entscheidung, sondern ihr Aufschub.

**Alles vor Phase 6 einsortieren und Phase 6 zuerst fertig bauen.** Wäre die
ordentliche Reihenfolge, und sie steht dem Auftrag entgegen. Der Preis steht
unten; er wird bezahlt, nicht verschwiegen.

## Was das kostet

- **`AGENTS.md` und `CLAUDE.md` sagen an drei Stellen „acht Phasen"** (§10 §12
  §13). Sie ändern sich mit — über `make leitlinien`, damit nicht eine der
  beiden Dateien eine andere Zahl führt als die andere.
- **Phase 6 bleibt liegen.** Zwölf Unteraufgaben, darunter der `.mse-set`-Rückschreibweg
  (#48) und der Zeichenplan aus einem echten Stil (#120). Die Freigabe rückt um
  den Umfang von Phase 9 nach hinten.
- **#143 entsteht auf einem Bestand, dessen Exportweg fehlt.** Was ein Modell
  über den Server ändert, verlässt die Anwendung heute nur als MTG JSON (#142)
  oder gar nicht. Das ist auszuhalten, aber es gehört gesagt: der Server ist
  vollständig, sein Umfeld ist es nicht.
- **Phase 9 wird vor Phase 6 abgeschlossen und bekommt ihren Zweitblick vor
  deren.** Damit liest der Zweitblick eine Naht, deren eine Seite noch wächst.

## Woran man merkt, dass die Entscheidung falsch war

- Wenn beim Bau von #141 die Zuordnung MTG JSON → Kartenfelder ohne ein echtes
  Spielpaket nicht zu treffen ist. Dann steht Phase 7 sachlich davor, und die
  Reihenfolge ist zu tauschen — nicht die Zuordnung zu erfinden.
- Wenn Phase 6 nach Phase 9 nicht wieder aufgenommen wird, sondern die nächste
  Anordnung wieder woanders hinzeigt. Dann ist nicht diese Entscheidung falsch,
  sondern das Phasenmodell — und das gehört dann als eigener Vermerk verhandelt.
- Wenn #143 mehr über den Bestand aussagt als über das Protokoll: sollte sich
  zeigen, dass zwei Schreiber auf einer SQLite-Datei eine eigene Phase Arbeit
  sind, ist Phase 9 zu klein geschnitten und #143 herauszulösen.
