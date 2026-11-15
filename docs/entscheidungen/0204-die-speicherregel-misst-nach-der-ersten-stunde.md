# 0204 — Die Speicherregel des Dauerlaufs misst nach der ersten Stunde und zählt das bekannte Restwachstum nicht als Trend

**Datum:** 2026-10-20
**Stand:** angenommen — Weg „Regel begründet ändern“, entschieden vom
Auftraggeber im Gespräch, **vor** dem Lauf, der sie belegt
**Anlass:** #2527
(Gesamtabnahme QA01, #2382)

## Anlass

Die Regel aus B04 (Planung 0.8.0, `docs/audits/0.8.0/ergebnis/PLAN-0.8.0.md:138`): nach der Aufwärmzeit von 20 Minuten **≤ 10 % Wachstum je Stunde und kein stetiger Anstieg über drei Abschnitte, „auch wenn er klein ist“** (`scripts/dauerbetrieb.test.mjs`). Der Plan verlangte: „Ziel vor Lauf bestätigen.“

Am 2026-10-20 hielt sie mit gleichbleibender Probe nicht (`589bdf8e5`, `docs/LEISTUNGSSCHNITTE.md` §9.1): 507 · 540 · 573 · 588 · 590 MiB, +12,3 %/h. Zwei Ursachen, je gemessen:

1. **Der Webinhalt wärmt eine Stunde lang auf** — in zwei unabhängigen Läufen: am 2026-10-12 (`b9a943091`) stieg der Anteil von WebKit in der ersten Stunde um 109 MB und stand danach; am 2026-10-20 um 84 MiB, danach 0,37 MB/min.
2. **Die Anwendung wächst um 0,166 MB/min** — dieselbe Rate wie nach #2312 (0,167), in den Bausteinen von macOS und der Go-Laufzeit, außerhalb des Go-Heaps; #2312 ist mit dieser Zuordnung geschlossen. Rund 3,3 MB je 20-Minuten-Abschnitt, bei 550 bis 600 MB zusammen **rund 0,6 % je Abschnitt** — und das genügt, damit drei Abschnitte hintereinander steigen.

## Entscheidung

1. **Aufwärmzeit 60 statt 20 Minuten.** Begründung: die beiden Läufe oben; eine Regel, die die Aufwärmung von WebKit als Trend zählt, misst den Zwischenspeicher und nicht ein Leck.
2. **Ein Abschnitt zählt als Anstieg erst ab +1 % über dem vorigen.** Begründung: das Restwachstum aus #2312 liegt bei rund 0,6 % je Abschnitt. Darunter liegt, was bekannt und zugeordnet ist; ab 1 % je Abschnitt — gut 3 % je Stunde — zählt jeder Trend weiter.
3. **≤ 10 % je Stunde bleibt**, gerechnet wie bisher vom ersten zum letzten ganzen Abschnitt.
4. **Neu: was wächst, muss freigebbar sein.** Der Zwei-Stunden-Lauf endet mit zwei Minuten Ruhe und einer Speicherwarnung an WebKit (`notifyutil -p org.WebKit.lowMemory`); danach liegt der Fußabdruck der WebKit-Prozesse der Anwendung **unter ihrem Median im ersten Abschnitt nach der Aufwärmzeit**. Gesendet ist nicht empfangen — reagiert WebKit nicht, ist der Lauf rot.

## Verworfen

- **Die Regel so lassen** — dann bräche sie an jedem Lauf allein am Restwachstum aus #2312, das zugeordnet ist und außerhalb dieses Codes liegt; sie sagte über ein neues Leck nichts mehr.
- **Die Aufwärmzeit länger als 60 Minuten** — ohne Messung dafür; eine längere wäre eine Zahl, die man passend wählt.
- **Eine Schwelle von 2 %** — sie ließe ein Wachstum von 5 % je Stunde als Trend unbemerkt; 1 % liegt knapp über dem Gemessenen und nicht großzügig darüber.

## Was das kostet

Ein Leck, das langsamer wächst als 1 % je Abschnitt, fällt dem Trend nicht mehr auf — nur noch der Rate. Und die erste Stunde eines Laufs sagt über Speicher nichts mehr.

## Beleg

Der Lauf mit dieser Regel steht in `docs/belege/huellenprobe/dauerbetrieb.md`; die Grenzen sind vor ihm festgelegt und in `scripts/dauerbewertung.mjs` und `scripts/dauerbetrieb.test.mjs` niedergelegt.

## Ergebnis vom 2026-10-21: die Regel hält nicht, und was wächst, gibt frei

Der Lauf unter dieser Regel (`fe167f897`, 125 Minuten, 1.496 Runden, Beleg
`docs/belege/huellenprobe/dauerbetrieb.md`): kein Absturz, nichts verloren,
Fußabdruck nach 120 Minuten höchstens 816 MiB (≤ 1.536). **Nach der
Aufwärmstunde 557 · 590 · 601 MiB, +11,8 % je Stunde, stetig — die Regel hält
nicht.** Die Grenzen bleiben, wie sie vor dem Lauf festgelegt waren.

**Was der Lauf zusätzlich zeigt** — und was den Anlass von Punkt 1
widerlegt: der Webinhalt wärmte diesmal **nicht** in der ersten Stunde auf.
Er stand dort bei 430 · 418 · 420 MiB und wuchs in der zweiten um 0,65 MB/min
(454 · 478 · 488). Die Anwendung wuchs wie zuvor um 0,15–0,17 MB/min (#2312).
**Und der Webinhalt gibt zurück**: nach zwei Minuten Ruhe 311 MiB, nach der
Speicherwarnung 242 MiB — unter dem Median seines ersten Abschnitts (455),
Punkt 4 hält. Was in der zweiten Stunde wächst, ist damit ein Zwischenspeicher,
den WebKit unter Druck freigibt, und kein Leck — aber wann er wächst, hängt am
Lauf und nicht an einer festen Aufwärmzeit.

**Folge**: D6 bleibt R3, die Zusage begrenzt (README, „Was diese Fassung
nicht belegt“). Eine Regel, die den Zwischenspeicher von einem Leck trennt,
müsste den Speicher nach einer Warnung messen statt im Lauf — das wäre eine
neue Entscheidung und kein Nachziehen dieser Grenzen.

