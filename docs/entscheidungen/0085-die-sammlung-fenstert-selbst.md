# 0085 — Die Sammlung fenstert selbst: TanStack Table und Virtual getrennt bewertet

**Stand:** angenommen
**Datum:** 2026-09-09
**Anlass:** Werkstück U8 von Phase 5.1
(#1419):
10.000 Exemplare als Tabelle, Galerie oder Gruppen, mit Auswahl über nicht
gezeichnete Zeilen hinweg — und das Issue verlangt, TanStack Table und
TanStack Virtual **getrennt** zu bewerten, bevor eines davon hereinkommt.

## Lage

**Was am 2026-09-09 galt, ist nachgelesen** (Stand `ba83c411`):

- `seiten/Sammlung.tsx` lädt alle Exemplare in einem Ruf (`Sammlung`), rechnet
  Suche, Ausführung und Sortierung in `sammlungssicht.ts` als reine Funktionen
  und zeichnet die ersten 200 Zeilen mit „mehr zeigen" (#1292). Bei 10.000
  Exemplaren steht die Liste nach 250 bis 530 ms, die Suche antwortet in 3 bis
  8 ms (`arbeitswegmass.mjs`, 2026-09-09).
- Die Auswahl hängt als `Set<id>` an der Liste; eine Zeile ist über ihre `id`
  gemeint, nicht über ihre Stelle.
- Das Frontend verlangt 100 % Abdeckung unter jsdom (§3), und jsdom rechnet
  **kein Layout**: `getBoundingClientRect` gibt Nullen, `ResizeObserver` gibt
  es nicht.

**Nachgeschlagen über Context7 am 2026-09-09:**

- **TanStack Virtual** (`/tanstack/virtual`, v3): `useVirtualizer` nimmt
  `estimateSize` als Schätzung und misst danach jedes Element über
  `measureElement`, standardmäßig mit `getBoundingClientRect`; das
  Scroll-Element liefert seine Größe über einen `ResizeObserver`. Ohne Layout
  ist jedes Maß null — unter jsdom zeichnet der Virtualizer nichts, und jeder
  Test bräuchte eine Attrappe für Rechtecke und Beobachter, die den Betrieb
  nicht abbildet (§3: eine Vorrichtung, die anders ist als der Betrieb, prüft
  die Vorrichtung).
- **TanStack Table** (`/tanstack/table`): kopflos; Gruppierung über
  `getGroupedRowModel` (v8) beziehungsweise `columnGroupingFeature` mit
  `createGroupedRowModel` (Hauptzweig, v9 im Umbau); Aggregation je Spalte
  über `aggregationFn`. Der Rat der Anleitung „with-tanstack-virtual": nicht
  die Rohdaten, sondern `table.getRowModel().rows` virtualisieren.

## Entscheidung

### 1 · Keine der beiden Bibliotheken — jetzt

**Die Sammlung fenstert selbst**: `bausteine/Fensterliste.tsx` zeichnet bei
fester Zeilenhöhe nur die Zeilen im Sichtfenster (Stand aus `scrollTop` und
`clientHeight`, Überhang zehn Zeilen), führt `aria-rowcount` und
`aria-rowindex`, und lässt den Fokus am **Datensatz** (`id`), nicht an der
Stelle. Unter jsdom sind `scrollTop` und `clientHeight` gewöhnliche
Eigenschaften, die ein Test setzt; kein Rechteck, kein Beobachter.

**Sortierung, Filter und Gruppierung bleiben reine Funktionen** in
`sammlungssicht.ts` — der Vertrag von #1200 und #1292 —, und `gruppiert()`
liefert Abschnitte mit Kopf (Name, Anzahl, Summe der Mengen). Das ist, was
Table an Gruppierung böte, in dreißig Zeilen, ohne Feature-Registrierung und
ohne eine API, die zwischen v8 und v9 gerade umgebaut wird.

### 2 · Die Naht bleibt

`Fensterliste` bekommt Zeilen, Zeilenhöhe und einen Zeichner. Wer später
variable Höhen braucht (Galerie mit Bildern verschiedener Formate), tauscht
den Baustein gegen `useVirtualizer` — und trägt dann die Attrappen für die
Tests. Heute hat jede Zeile dieselbe Höhe, und die Galerie hat feste Kacheln.

### 3 · Was vertagt ist, steht mit Grund

Gruppieren nach **Farbe** und **Typ** braucht Angaben, die ein Exemplar heute
nicht kennt: es trägt Set, Nummer, Zustand, Sprache und Lagerort. Gebaut wird
gegen die heutige Druckidentität (Entscheidung vom 2026-09-09).

**Hier stand „kommen mit A2", und das war falsch** — angestrichen vom
Zweitblick über Phase 5.1 (Befund 2, 2026-09-10). A2
(#1441) ist
seit dem 2026-09-09 gebaut, und seine Druckzeile führt Name, Setcode, Nummer
und Sprache — **keine Farbe und keinen Typ**. Sie kommen mit **A3**
(#1442):
dort wird die MTGJSON-Sammelausgabe angebunden, und die trägt beides je Druck.
Erst dann sind es zwei weitere Gruppierungen an derselben Naht — vorher wäre
es geraten.

**Was das über die Abnahme von U8 sagt**:
#1419 nennt
sechs Gruppierungen und ist mit vieren geschlossen worden. Das ist im Issue
selbst so vorgesehen — es nennt A3 als Voraussetzung —, aber der Vermerk hat
den falschen Vorgänger genannt und damit den Eindruck erweckt, die Sache sei
erledigt. Die zwei fehlenden stehen jetzt in der Abnahme von A3.

## Verworfen

- **TanStack Virtual jetzt.** Für eine Liste gleicher Zeilen ist der
  Messapparat Ballast, und unter jsdom prüfte jede Abdeckung eine Attrappe.
- **TanStack Table jetzt.** Die drei Rechnungen sind da und getestet; die
  Bibliothek brächte eine Zustandsmaschine für Spalten, die das Fenster nicht
  braucht — und eine API im Umbau.
- **Weiter 200 Zeilen und „mehr zeigen".** Eine Auswahl über nicht gezeichnete
  Zeilen und eine Tastatur, die 10.000 Zeilen erreicht, gehen damit nicht;
  genau das ist die Abnahme.

## Folgen

- Keine neue Abhängigkeit; `DRITTANBIETER.md` bleibt.
- `arbeitswegmass.mjs` prüft die Sammlung mit 10.000 Exemplaren weiter, und
  der Lauf bekommt eine Probe, die eine Zeile jenseits des Sichtfensters
  auswählt und nach einem Filterwechsel wiederfindet.
