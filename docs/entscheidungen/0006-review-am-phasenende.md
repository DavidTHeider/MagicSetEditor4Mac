# 0006 — Zweitblick am Phasenende statt je Pull Request

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

Im Schwesterprojekt Projektkompass wird **jeder** PR vor dem Merge von einem
zweiten Werkzeug gelesen (dort: Codex). Das hat sich bewährt und ist teuer: die
Nachmessung dort zeigt, dass ab dem zweiten Durchgang zunehmend die *Antwort auf
den Zweitblick* geprüft wird, nicht mehr die Änderung.

Hier kommt hinzu, dass die Arbeit von Subagenten in Serie erledigt wird. Ein
Zweitblick je PR hieße: ein Werkzeugwechsel je Unteraufgabe.

## Entscheidung

**Zwei Ebenen, verschieden geschnitten:**

| Ebene | Wann | Wer | Gegenstand |
| --- | --- | --- | --- |
| **Inline-Review** | vor jedem Merge | derselbe Zweig, der gearbeitet hat | Testabdeckung, Schichtenregel, Zeilengrenze, Doku, die Regeln aus §7 |
| **Zweitblick** | am **Phasenende** | Codex | der gesamte Beitrag der Phase — vor allem die Nähte zwischen den PRs |

Der Zweitblick am Phasenende läuft **genau zweimal**: einmal über die fertige
Phase, und nach der Korrekturrunde einmal darüber. Danach ist die Phase
abgeschlossen. Befunde werden geprüft, nicht blind umgesetzt; ein falscher
Befund wird begründet abgelehnt statt stillschweigend übergangen.

**Am Ende aller acht Phasen steht ein eigenständiges Gesamtreview** über das
fertige Werk und die Nähte zwischen den Phasen.

Zum Phasenabschluss gehört danach zwingend das Nachschärfen der Issues der
folgenden Phasen (`AGENTS.md` §13).

## Verworfen

**Zweitblick je PR, wie im Projektkompass.** Verworfen für dieses Projekt, nicht
als Verfahren. Er kostete einen Werkzeugwechsel je Unteraufgabe, und die Befunde,
auf die es hier ankommt, entstehen ohnehin zwischen den PRs: eine Naht zwischen
`msescript` und `render`, ein Vertrag zwischen `app` und dem Frontend, eine
Rundlaufzusage, die drei PRs gemeinsam halten. Genau die sieht ein PR-Review
nicht.

**Gar kein Zweitblick, nur Inline.** Verworfen. Ein Zweig, der seine eigene
Arbeit liest, findet Handwerkliches und übersieht die Annahme, unter der er
gearbeitet hat. Die teuersten Befunde im Schwesterprojekt waren durchweg von der
zweiten Art.

**Ein Durchgang am Phasenende.** Verworfen. Nachgemessen in Phase 1 dieses
Projekts: der erste Durchgang fand 17 Befunde, der zweite bewertete fünf der
sieben Korrekturen als nur *teilweise* geschlossen und fand zwei neue. Eine
Korrektur, die den *genannten* Fall behebt und die Klasse dahinter stehen lässt,
sieht wie eine Lösung aus und ist keine.

**So lange, bis einer nichts mehr findet.** Verworfen, und das ist eine
Entscheidung mit Preis. Ein dritter Durchgang fände etwas — aber ab dem zweiten
prüft der Zweitblick zunehmend die Antwort auf den Zweitblick, und das hat kein
natürliches Ende. Was dabei liegen bleibt, fängt das Gesamtreview am Ende auf;
bis dahin ist es liegen geblieben, und das steht hier, damit es niemand für
geprüft hält.

## Folgen

- Eine Phase ist abgeschlossen, wenn zwei Durchgänge gelaufen sind und ihre
  materiellen Befunde behoben oder als Issue eröffnet wurden. Beide Durchgänge
  gehören in die Planung, nicht in den Puffer.
- Der Inline-Review muss dafür wirklich stattfinden. Wo er zur Formsache wird,
  bekommt der Zweitblick am Phasenende Handwerksbefunde statt Nahtbefunde — und
  genau die Arbeit, für die er da ist, kommt nicht dran.
- Ein Befund über etwas ganz anderes wird nicht weggeworfen, sondern als Issue
  eröffnet — gleich in welchem Durchgang er auftaucht.
