# 0076 — Die Ordnung der Arbeitsplätze: der Rahmen bleibt, der Arbeitsplatz bekommt drei Regionen

**Stand:** angenommen
**Datum:** 2026-09-08
**Anlass:** Werkstück U1 von Phase 5.1
(#1412) —
die Abnahme verlangt, dass **zwei Layoutvarianten mit Vor- und Nachteilen
bewertet** sind und die gewählte samt Mindestfenstermaß dokumentiert ist,
**bevor** ein Arbeitsplatz umgebaut wird (U5, U7, U8, U10).

## Lage

**Was heute ist, ist gemessen** (2026-09-08, `frontend/src/`, Stand
`v0.4.0`):

| | |
| --- | ---: |
| Seiten in `navigation.ts` | **23** — 10 Anwendungsseiten, 13 Laschen eines offenen Sets |
| Seitenleiste | fest, **240 Punkte** (`w-60`), Kürzel ⌘1 bis ⌘8, ⌘?, ⌘,, ⌘' |
| Inhaltsbreite | `max-w-6xl` (1.152) für Anwendungsseiten, `max-w-[100rem]` (1.600) für Laschen |
| Kartenlasche | zwei Spalten — Liste links, Formular und Vorschau rechts; **unter rund 1.000 Punkten untereinander** (`seiten/kartenraster.ts`) |
| Sammlung | Werkzeuge in einem Aufklappkasten, darunter eine Liste, gekappt bei 200 Zeilen |
| Decks | zwei Gebiete als Laschen; Anlage, Übernahme, Liste und Werkbank **untereinander** |
| Vorlagen | Auswahlfeld, darunter Dateiliste, Textfläche, Vorschau |
| Fenster | Vorgabe 1.280 × 860, **Minimum 960 × 640** (`main.go`) |

Der Plan für 0.5.0 verlangt für vier dieser Seiten einen **Arbeitsplatz**:
Karten (U5, #1416)
mit Navigator, Vorschau, Inspector und Problemliste; Vorlagen (U7) mit
Paketbaum, Editor und Vorschau; Sammlung (U8) mit Tabelle, Galerie und
Detailinspector; Decks (U10) mit Suche, Deck und Analyse. Jede dieser Seiten
hat heute ihre eigene Ordnung, und keine trägt den Namen der anderen.

**Die Frage ist nicht, ob es drei Regionen gibt** — das sagen die vier
Issues. Die Frage ist, **wo sie liegen**: im Rahmen, für alle Seiten gleich,
oder in der Seite, für jeden Arbeitsplatz einzeln.

## Die zwei Varianten

### A — Drei Spalten im Rahmen

Der Rahmen bekommt neben der Seitenleiste **drei feste Spalten**: Navigator,
Zentrum, Inspector. Jede Seite füllt sie; eine Seite ohne Navigator lässt die
Spalte leer oder klappt sie ein.

| Für | Gegen |
| --- | --- |
| **Ein** Modell für die ganze Anwendung: Navigator ist immer links, Inspector immer rechts, die Tasten zum Einklappen gelten überall | **Vier Spalten auf 1.280 Punkten**: 240 Leiste + 240 Navigator + 320 Inspector lassen dem Zentrum **480** — die M15-Karte (375 × 523) passt hinein, ein Formular daneben nicht mehr |
| Der Fokusweg ist auf jeder Seite derselbe (Leiste → Navigator → Zentrum → Inspector) | Start, Hilfe, Einstellungen, Pakete, Turniere und Nutzer **haben keinen Navigator** — sechs von zehn Anwendungsseiten müssten die Spalte wegerklären |
| Ein Inspector, der Seiten überlebt, könnte dieselbe Karte in Sammlung **und** Deck zeigen | **Ein Umbau in einem Zug**: der Rahmen ändert sich für alle 23 Seiten zugleich, und die 13 Laschen des Setarbeitsplatzes mit ihrer `role="tablist"` müssten in die Spaltenordnung eingepasst werden — das ist kein Werkstück, das ist eine Phase |
| | Jede Messung aus `SERIENARBEIT.md` — Tabstellen je Seite, was über der Falz steht — verlöre ihren Bezug, ohne dass ein einzelner Arbeitsplatz besser geworden wäre |

### B — Der Rahmen bleibt, der Arbeitsplatz bekommt seine Regionen

Seitenleiste und `<main>` bleiben, wie sie sind. Ein **gemeinsamer Baustein
`Arbeitsplatz`** (aus U2, #1413)
legt **innerhalb** einer Seite drei Regionen an — Navigator, Zentrum,
Inspector —, und die vier Arbeitsplätze benutzen ihn. Die übrigen Seiten
bleiben, was sie sind.

| Für | Gegen |
| --- | --- |
| **Ein Werkstück, eine Seite**: U5, U7, U8 und U10 bauen je einen Arbeitsplatz um, und jeder ist für sich abnehmbar — genau der Zuschnitt der Issues | **Zwei Ordnungen nebeneinander**: eine Seite mit drei Regionen neben einer mit einer Spalte. Der Baustein muss die Regionen **erzwingen** (Lage, Breiten, Tasten), sonst hat jede Seite wieder ihre eigene |
| Nichts an Seitenleiste, Kürzeln, Palette, Menüleiste und Setarbeitsplatz ändert sich; jede Zahl aus `SERIENARBEIT.md` bleibt vergleichbar | Ein Inspector gehört seiner Seite: die Karte aus der Sammlung im Deck zu zeigen, braucht eine Übergabe statt eines gemeinsamen Zustands |
| Auf 1.280 Punkten bleiben dem Arbeitsplatz **1.040**: Navigator 240, Inspector 320, Zentrum **480** — dieselbe Zahl wie in A, aber **nur dort, wo ein Zentrum gebraucht wird**; Start und Hilfe behalten ihre 1.040 | |
| Navigator und Inspector lassen sich je Arbeitsplatz einklappen (Tasten aus U2), und der Baustein merkt sich das je Seite im Settingsmodell (U4, O3) | |

## Entscheidung

**Variante B**, mit drei Zusagen, die den Nachteil von B — zwei Ordnungen —
klein halten:

1. **Ein Baustein, vier Verbraucher.** `Arbeitsplatz` entsteht in U2 und wird
   in U5 am Kartenarbeitsplatz und in U8 an der Sammlung eingesetzt, **bevor**
   Vorlagen und Decks folgen — das ist die Abnahme von U2 („ein
   Kartenarbeitsplatz und eine dichte Verwaltungsansicht verwenden dieselben
   neuen Primitive; danach erst breitere Umstellung").
2. **Dieselbe Lage, dieselben Tasten, dieselben Breiten.** Navigator links
   mit **240** Punkten, Inspector rechts mit **320**, Zentrum dazwischen; die
   Tasten zum Einklappen legt U2 fest und trägt sie in die Kürzeltafel ein.
   Ein Arbeitsplatz, der davon abweicht, ist ein Befund.
3. **Unter 1.200 Punkten Inhaltsbreite klappt der Navigator von selbst ein**,
   damit das Zentrum nie unter **480** fällt — die Breite, die eine M15-Karte
   samt Rand braucht.

### Referenzgerät und Fenstermaße

| | Maß | Warum |
| --- | ---: | --- |
| **Referenzgerät** | MacBook Air 13″ (M2/M3): **1.470 × 956** Punkte im Standardmaßstab | Das kleinste Gerät, auf dem die Anwendung ausgeliefert wird (§12: darwin/arm64 ab macOS 13). Nach Menüleiste (37) und Dock (rund 70) bleiben rund **1.470 × 850** |
| **Messfenster** | **1.280 × 800** | Unverändert seit Phase 4.3 (`SERIENARBEIT.md`): passt auf das Referenzgerät mit Rand, und jede bisherige Zahl ist dort erhoben |
| **Vorgabefenster** | 1.280 × 860 | Unverändert (`main.go`) |
| **Mindestfenster** | **1.024 × 700**, vorher 960 × 640 | Bei eingeklapptem Navigator: 240 Leiste + 320 Inspector lassen **464** für das Zentrum — eine M15-Karte (375) mit Rand. In der Höhe: Kopf 60 + Werkzeugzeile 40 + Karte 523 + Ränder 24 = **647**. Bei 960 × 640 blieben 400 und 640 — die Karte wäre unten abgeschnitten |

**Die Mindestgröße gilt ab jetzt**, nicht erst mit U5: ein Fenster, das
kleiner ist als der erste Arbeitsplatz braucht, würde beim Umbau größer
gemacht, und ein Nutzer, dessen gespeichertes Fenstermaß plötzlich nicht mehr
gilt, hat einen Grund zu fragen, warum. Jetzt gibt es keinen Arbeitsplatz,
der 960 braucht — die Kartenlasche stapelt unter 1.000 ohnehin.

## Verworfen

- **Variante A**, aus den Gründen in der Tabelle: vier Spalten auf 1.280
  Punkten, sechs Seiten ohne Navigator, ein Umbau in einem Zug.
- **Ein Mindestfenster von 1.280 × 800.** Das Messfenster als Minimum hieße,
  dass niemand zwei Fenster nebeneinander stellen kann; das Referenzgerät
  hat 1.470 Punkte Breite.
- **Das Minimum bei 960 × 640 lassen und in U5 erhöhen.** Siehe oben — die
  Änderung gehört dorthin, wo sie entschieden wird, nicht dorthin, wo sie
  auffällt.
- **Ein gemeinsamer Inspector-Zustand über Seiten hinweg.** Er ist der
  eine Vorteil von A, den B nicht hat, und er ist keine Anforderung: U9
  („Rückweg Sammlung → verwendende Decks") ist ein Verweis, kein geteilter
  Zustand.

## Folgen

- `main.go` trägt `MinWidth: 1024`, `MinHeight: 700`; `docs/UX-ARCHITEKTUR.md`
  nennt beide Zahlen und diesen Vermerk, und ein Wächter hält die Zahlen
  gegeneinander (`internal/werkzeuge/uxarchitektur_test.go`).
- U2 baut `Arbeitsplatz` mit den drei Zusagen oben; U5 und U8 sind seine
  ersten Verbraucher.
- `scripts/arbeitswegmass.mjs` misst die Kernwege weiter bei 1.280 × 800 —
  eine Änderung des Messfensters machte jede frühere Zahl unvergleichbar.

## Nachtrag vom 2026-09-09 — die Zahlen, am Arbeitsplatz gemessen (U5)

Die Sichtprüfung von U5 (#1416)
hat den ersten Verbraucher dieses Bausteins an der laufenden Anwendung
vermessen, und zwei Zahlen oben waren **ohne Ränder gerechnet**:

| Fenster | Inhaltsbereich | Arbeitsplatz | Zentrum neben dem Inspector |
| ---: | ---: | ---: | ---: |
| 1.440 | 1.200 | **1.086** | 494 (Navigator steht) |
| 1.280 (Messfenster) | 1.040 | **926** | 590 (Navigator zu) |
| 1.024 (Mindestfenster) | 784 | **670** | **334** (Navigator zu) |

Zwischen Inhaltsbereich und Arbeitsplatz liegen **114** Punkte: die Ränder des
Rahmens und des Setkopfs. Daraus folgen zwei Korrekturen, und die Entscheidung
bleibt, wie sie ist:

1. **Die Grenze von 1.200 gilt für den Inhaltsbereich; der Baustein misst
   sich selbst.** Mit 1.200 an ihm gemessen stand der Navigator auf keinem
   Fenster, das es gibt — auch nicht auf dem Referenzgerät (1.470, Arbeitsplatz
   1.116). Er klappt deshalb unter **1.072** ein, der Summe seiner Teile
   (240 + 480 + 320 + zweimal 16 Abstand); am Inhaltsbereich sind das 1.186.
2. **Die 480 sind eine Grenze, kein CSS-Maß.** Als Mindestbreite des Zentrums
   ließen sie die Seite im Mindestfenster waagerecht rollen — 480 + 320 passen
   nicht in 670. Das Zentrum nimmt dort, was bleibt (**334**, nicht die 464 der
   Tabelle oben), und die Vorschau passt sich an; die 480 hält die Grenze zum
   Einklappen, solange der Navigator steht.

**Und der Knopf sagt, wenn kein Platz ist.** Ein Klick auf „Navigator
einblenden" unter der Grenze schrieb den Wunsch ins Gedächtnis (U4), und zu
sehen war nichts; seit U5 ist er angehalten, und daneben steht, warum.

## Nachtrag vom 2026-09-09 — wer bei Enge weicht, entscheidet die Seite

Punkt 3 oben sagt: unter der Grenze klappt der **Navigator** von selbst ein.
Für die Kartenseite stimmt das — die Liste ist auch über Kürzel zu ersetzen,
der Inspector mit den Feldern nicht. Für die **Vorlagenwerkstatt** (U7,
#1418) stimmt
das Gegenteil: ihr Navigator ist der Baum aus Paketen und Dateien, und ohne
ihn führt kein Weg zu einer Datei. Der Messlauf `arbeitswegmass.mjs` bei
1.280 × 800 hat es gezeigt: „Kein Platz für den Navigator", der Knopf
angehalten, die Werkstatt leer.

Der Baustein bekommt deshalb die Angabe `beiEnge` — `navigator` (Vorgabe,
U5) oder `inspector`. Die Werkstatt gibt den Inspector her; er kommt zurück,
sobald Platz ist, und sein Knopf sagt in der Zwischenzeit, warum er ruht.
Die Grenze selbst (1.072) bleibt.
