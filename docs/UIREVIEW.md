# Die Durchsicht der Oberfläche — das Prüfmaß

**Phase 6.13** (#1906),
angelegt mit UXR-01 (#1909).
Diese Datei hält **was gemessen wird** und **welche Bauteilregeln gelten** — das
Prüfmaß, auf das sich Code und Tests mit R1 bis R7 berufen.

**Die Ausgangslage, die Züge danach und die Entscheidung je Befund** — die
Messwerte von Phase 6.13 bis 0.9.0 — liegen seit 1.0.0 im
Entwicklungsarchiv (Vermerk
[0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). Was gilt, steht hier.

## Wie gemessen wird

**Am nativen Fenster, nicht im Browser**: die Hüllenprobe, Drehbuch
[`uirevue`](../scripts/huellenprobe/uirevue.mjs), startet das gebaute Bündel
mit eigenem Bestand und einem geöffneten Set, geht über die Palette jede
**Seite** an (25 Seitenschlüssel) und zählt im Inhaltsbereich — oder, wenn sie
offen ist, **in der Schublade**:

| Größe | Was |
| --- | --- |
| Bedienelemente | Knöpfe, Verweise, Laschen, Menüeinträge, Aufklapper — gesamt und **im Bild** (ohne Rollen sichtbar) |
| Hauptknöpfe | Knöpfe im Hauptgewicht (`bg-marke`) — gesamt und **höchstens je Region** (`section`, `form`, `aside`, Laschenfeld, Dialog) |
| Felder | Eingaben, Auswahlen, Textfelder, Schalter — gesamt und im Bild |
| Überschriften | h1–h4 |
| Rolltiefe | Höhe des Inhalts durch Höhe des Rollbereichs; 1 heißt: passt |
| waagerecht | muss die Seite waagerecht gerollt werden? |

**Zwei Größen**: 1280 × 800 („normal") und 1024 × 700 („eng" — die kleinste
Breite, die das Fenster annimmt). Gesetzt über „wie zuletzt" mit gemerkten
Maßen; ein `set_size` am maximierten Fenster ließ Wails' MCP-Server
verstummen (erster Lauf am 2026-10-02).

**Was das Maß nicht sieht**: ob eine Handlung die richtige ist, ob eine
Beschriftung verständlich ist, wie sich eine Aufgabe bedient. Das bleibt die
Sichtprüfung aus Nutzersicht (UXR-*.AK-04). Und es zählt, was **ohne
weiteres Klicken** dasteht.

**Seit dem 2026-10-02 zählen die Feldflächen über einer Karte nicht mit**
(UXR-06, #1914):
ein Knopf ohne Tabstelle hinter einem Bild ist der Weg der Maus in ein Feld,
ohne eigene Gestalt, und das Formular daneben trägt dieselben Felder. Mit ihnen
hätte eine reiche Vorlage R2 allein durch ihre Felder gerissen — `magic-m15`
zeichnet rund fünfzehn. **Die Zahlen davor** (Ausgangslage und die ersten zwei
Züge, im Entwicklungsarchiv) zählen sie; auf dem Probebestand sind es drei, und sie stehen nur auf
*Karten*.

## Das Prüfmaß — die Bauteilregeln

Abgeleitet aus den Artboards des Entwurfs der Zieloberfläche 0.6.0 (gezählt:
3–14 Knöpfe je Fläche, der Tisch 22; nie zwei Hauptknöpfe in einem Bereich)
und aus der Ausgangslage vom 2026-10-02 — beide im Entwicklungsarchiv.

| Regel | Grenze | Warum diese Zahl |
| --- | --- | --- |
| **R1 — eine führende Handlung** | höchstens **ein Hauptknopf je Region** — und seit AP-3 **je Blick höchstens ein gefüllter im Bild**, ein gewählter Umschalter **nie** gefüllt; der Tisch ausgenommen | das Artboard zeigt nie zwei nebeneinander; `Zustaende09` sagt „je Blick“, und drei Knöpfe in drei Regionen konkurrieren doch. **Der Tisch** ist eine Spielfläche, an der der Kern bestimmt, welche Antwort fällig ist; was dort führt, entscheidet `Tisch09` (P4) |
| **R2 — gleichrangige Bedienelemente** | höchstens **20 im Bild** bei 1280 × 800, der Tisch ausgenommen | das reichste Artboard ohne Tisch (Sets) zeigt 14; 20 lässt einer gefüllten Liste Raum |
| **R3 — keine Formularwand** | höchstens **12 Felder** gezeichnet auf einer Fläche, sonst gruppiert und **eingeklappt** bis auf eine Gruppe — **eine Segmentreihe zählt einmal** (sie ist eine Entscheidung wie eine Auswahlliste; seit UXR-F2) | das Artboard *Karten* zeigt 19 Felder in Gruppen, *Einstellungen* 7 |
| **R4 — Rolltiefe** | höchstens **2,5** bei 1280 × 800 | Start misst 1,9 und liest sich als eine Seite; darüber wird gesucht statt gesehen |
| **R5 — nie waagerecht** | kein waagerechtes Rollen, auch eng | §6 und UX-14 |
| **R6 — lange Listen** | über **200** gezeichnete Bedienelemente auf einer Fläche werden gefaltet oder nur im Bild gezeichnet | jedes weitere kostet Webprozess-Speicher (#1852) und Tabstopps |
| **R7 — das passende Bedienelement** | Schalter für Sofortwirkung, Segmente bis **4** Werte, Auswahl darüber, Auswahl **mit Suche ab 15** Werten | Mantine-Regel aus UX-07; für die Einstellungen gezählt (`wahlkatalog.test.tsx`, UXR-F2), sonst Durchsicht je Issue |

**R1 bis R6 hält heute jede Seite** bei 1280 × 800 (R5 auch eng) — **mit
dem Probeset der Saat**. Befüllt hält es nicht: im großen Set, im offenen
Deck und in der gefüllten Sammlung reißen R2, R3 und R4, und R1 je Blick
reißt fast überall (Messung „Befüllt“, AP-13, am 2026-10-13 — im
Entwicklungsarchiv). Das ist dieselbe Gestalt wie §10, »die Reichweite einer Messung«. Das
Drehbuch `uirevue` prüft alle sechs, und eine Seite, die darunter fällt,
macht den Lauf rot — belegt am 2026-10-04 in
[`belege/huellenprobe/uirevue.md`](belege/huellenprobe/uirevue.md), 50 von
50 Seitenaufrufen. **Bis zum Nachlauf über 6.13 stand hier „R1 und R5"**,
und an anderer Stelle schon „alle sechs", während der abgelegte Beleg nur zwei
Prüfzeilen trug (N5).
