# VK-01 — Rechnet die Ausfuhr ein Kartenfeld im Stil der Karte?

**Diagnose vom 2026-10-07** zu #2161.
Ergebnis: **nein, und es hat eine Wirkung.** Die Umsetzung ist als eigener
PR geschnitten (AK-03) — siehe das Folge-Issue unten.

## Was MSE2 tut

MSE2 rechnet jedes Kartenfeld im Kontext **des Stils der Karte**:
`SetScriptContext::getContext(card)` holt `set.stylesheetForP(card)`, und der
Kontext eines Stils führt dessen `init script:` aus und setzt `stylesheet`,
`card_style` und `styling` (`src/script/script_manager.cpp`, Zeilen 30–75,
Fassung f9d9356 vom 2021-07-06). Ein Stil kann damit Funktionen des Spiels
überschreiben.

**Im Pack geschieht das**: `magic.mse-game` erklärt `card_shape := { "normal" }`
(`script`, Zeile 395) und das Feld `shape` als `stylesheet; card_shape()`
(`card_fields`) — „determined by the style, by override", und `shape` bestimmt
laut Pack die Stelle der Karte in der Setnummerierung. Stile überschreiben es,
darunter `magic-m15-mainframe-planes` (`card_shape := { "plane card" }`,
Zeile 106), `magic-agclassic-token` und `magic-DayNight`.

## Die Messung

**Eingabe**: das kanonische Set `Planechase.mse-set` (86 Karten, alle mit
`stylesheet: m15-mainframe-planes`, SHA-256 beginnend `112566f4b051bd5e`) und
die Pakete aus `full-magic-pack` — **als Kopien** in einer Wegwerfumgebung
(eigenes `HOME`, `MSE_BESTAND`, `MSE_SCHLUESSELBUND`), denn das Paketlesen
weist Verweise nach draußen ab.

**Orakel** — MSE2 aus `~/mse2-quelle` (f9d9356), isoliert mit `HOME` und
`CFFIXED_USER_HOME`:

```
to_text(for each c in set.cards do c.name + "=" + c.shape + ";")
```

**Diese Anwendung** — dieselbe Frage als eigene Exportvorlage
(`script: to_text(for each card in cards do card.name + "=" + card.shape + ";")`),
`mse vorlagen 1 ausfuehren magic-vk01form.mse-export-template …`.

| | `shape` = „plane card" | `shape` = „normal" |
| --- | --- | --- |
| MSE2 | **86** | 0 |
| diese Anwendung | 0 | **86** |

**Kosten** (AK-02): unsere Ausfuhr über die 86 Karten 0,62–0,71 s, MSE2 für
Laden und Frage 0,55–0,65 s — die Lücke ist keine Frage der Zeit.

## Was nicht gelang

**Ein selbst geschriebenes Set mit zwei Stilen** (AK-01 verlangt eines)
brachte MSE2 in eine Endlosschleife beim Lesen der `pack types` des Sets
(`Reader::handle<vector<PackType>>`, gesehen mit `sample`), auch mit den
Schlüsseln echter Dateien und als ZIP. Die Ursache ist nicht gefunden; der
Beleg stützt sich deshalb auf ein **echtes** Set und echte Stile, und das
Selbstgeschriebene bleibt für den Umsetzungs-PR (dort als Go-Test gegen
unsere Ausfuhr, ohne Orakel).

## Was folgt

Die Umsetzung: je Kartenstil eine Rechenumgebung aus Spiel **und** Stil
(dessen `init script:`, `stylesheet`, `card_style`), damit `shape` und jede
überschriebene Funktion so rechnen wie in MSE2 — und damit auch
`card_number` und die Reihenfolge, die davon abhängen. **Tiefere
Abhängigkeitsketten** (`updateAllDependend`) bleiben getrennt entschieden
(`docs/PARITAET.md`).

## Nachher — die Umsetzung (#2182)

**Gemessen am 2026-10-07** mit derselben Eingabe und derselben Vorlage,
nachdem jede Karte im Kontext ihres Stils rechnet
(#2182,
`internal/service/ausfuhrkontext.go`):

| | `shape` = „plane card" | `shape` = „normal" |
| --- | --- | --- |
| MSE2 | **86** | 0 |
| diese Anwendung | **86** | 0 |

**Alle 86 Zeilen `Name=shape` sind gleich** — SHA-256 beider Listen beginnt
mit `e967c427dfd66313`.

**Der erste Lauf nach dem Umbau gab noch 86 × „normal"**, und der Grund lag
eine Schicht höher: der Vorlagenweg, über den Kommandozeile, REST und MCP
ausführen (`app.VorlagenwegUeber`), setzte **keine Stilwahl**. Eine eigene
Probevorlage zeigte `set.stylesheet` dort als `nil` — auch die
Stilgegenstände aus #2039 kamen also nur im Fenster an. Der Go-Test
`TestDerVorlagenwegKenntDenStilDesSets` hält das jetzt fest.

**Kosten**: 1,09 s statt 0,62–0,71 s für die 86 Karten — das `init script`
des Stils und das Öffnen der Stilpakete. MSE2 braucht für Laden und Frage
0,55–0,65 s.

**Das selbst geschriebene Zwei-Stile-Set** steht als Go-Test gegen die
Ausfuhr (`internal/service/exportkartenstil_test.go`): zwei Stile, deren
`init script` dieselbe Funktion verschieden belegt, und eine Karte mit
eigenem Stil zwischen zweien mit dem des Sets. Gegen das Orakel läuft es
weiterhin nicht — siehe „Was nicht gelang".
