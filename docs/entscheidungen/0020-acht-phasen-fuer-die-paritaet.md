# 0020 — Acht Phasen für die Parität zu MSE2, und die Messung zuerst

**Stand:** angenommen
**Datum:** 2026-08-16
**Ergänzt:** [0006](0006-review-am-phasenende.md) — der Zweitblick hängt an
Phasen, und seit dem Abschluss der neun gab es keine

## Lage

Die neun Phasen aus dem Entwurf vom 2026-08-08 sind abgeschlossen. Seither
läuft die Arbeit **ohne Phasenstruktur**, und das ist keine Formsache: der
Zweitblick durch Codex hängt am Phasenende (§10), das Nachschärfen der Issues
ebenso (§13), und die Fortschreibung von Roadmap und Chronik hat ohne
Phasengrenze keinen Anlass mehr.

**Der Preis ist am 2026-08-16 angefallen.** Vier Werkstücke — #274, #275, #277,
#278 — sind ohne Zweitblick gemerged worden. Genau davor hat der Auftrag von
#230 gewarnt:
„Wird nach dem Zweitblick weitergebaut, bekommt der neue Teil einen eigenen
Durchgang."

Zugleich hat sich das Ziel geändert. Bis zu
#240 lautete es
*„eine Karte zeichnet überhaupt"*; seit sie zeichnet, lautet es **Parität zu
MSE2**: dieselbe Vorlage, dieselben Kartendaten, **dasselbe Bild** — innerhalb
einer benannten Toleranz. Das ist die erste Zielsetzung dieses Projekts, die
sich messen lässt, statt sie zu behaupten.

### Der Stand, nachgemessen und nicht übernommen

Der Lauf hinter `fremdpakete` am 2026-08-16, über die 190 Stile von
`M15-Magic-Pack`:

```
190 Stile, 5950 Feldstile → 795 Bildebenen, 3119 Textebenen, 2729 Befunde
19 Exportvorlagen gelesen, 13 liefen durch
```

## Entscheidung

**Acht neue Phasen, A bis H**, geführt in
[`../ROADMAP.md`](../ROADMAP.md). Buchstaben statt Nummern, damit niemand sie
mit den neun alten verwechselt: die neun sind Geschichte und bleiben stehen,
diese acht sind Plan.

**Phase A steht vor allem anderen, und sie baut kein Merkmal.** Sie baut die
Messung: einen Goldbildvergleich über die Naht Stil → Plan → Bild. Ohne ihn
ist „Parität" eine Meinung; mit ihm ist jede der Phasen B bis F ein Schritt,
dessen Wirkung in einer Zahl steht — und eine Phase, die die Zahl nicht bewegt,
hat etwas anderes getan, als sie sollte.

**Jede Phase hat ein Abnahmekriterium, das eine Messung ist.** Die Tabelle in
der Roadmap führt sie einzeln.

### Drei Änderungen am Vorschlag der Zieldatei, jede gemessen

Die Zieldatei nennt ihren Schnitt ausdrücklich „einen Vorschlag, keine Vorgabe"
und verlangt für jede Änderung eine Begründung in der Roadmap. Es sind drei.

**Erstens: Phase E wiegt 639 Befunde, nicht 71.** Der Vorschlag führt für „Was
die Karte mitbringt" die 71 Meldungen „ein eigenes Bild". Der Messlauf zerlegt
die bis dahin unsortierte Klasse „weder eine Schrift noch ein Bild" (370) nach
Feldnamen:

| Anzahl | Feld |
| ---: | --- |
| 137 | `border color` |
| 125 | `image` |
| 87 | `mainframe image` |
| 13 | `image 2` |
| 5 | `name` |
| 1 | `mainframe image 2`, `name 2`, `rarity` (je 1) |

**226 davon sind Bildfelder der Karte.** Zusammen mit den 342 Wahlbildern des
Spiels und den 71 eigenen Bildern trägt Phase E damit **639** der 2.729
Befunde — der zweitgrößte Posten überhaupt. Sie rückt deshalb vor Phase D.

**Zweitens: `exists_in_package` (126) wandert von F nach C.** Der Vorschlag
zählt es zur Sprache. Der Messlauf sagt, wo es auftritt:
`image: ausrechnen → der Name "exists_in_package" ist hier nicht bekannt` — es
steht auf dem **Bildweg** und blockiert Bildebenen, nicht Textwerte. Eine Phase,
deren Abnahme an Bildebenen hängt, muss es mitnehmen.

**Drittens: Phase B bewegt die Befundzahl fast nicht, und das steht jetzt
dabei.** `content_width` und `content_height` kommen in genau **6** Befunden
vor; die verwandte Klasse „ergab keine Zahl" in 100. Der Vorschlag begründet
Phase B mit dem zweistufigen Satz, und das bleibt richtig — aber ihre Wirkung
zeigt sich **allein** im Goldbild aus Phase A, nicht im Zähler. Wer sie nach
dem Zähler bewertete, hielte sie für vergeudet.

### Was der Zähler nicht sieht, und warum das der Grund für Phase A ist

**Phase D bewegt die Befundzahl um null.** `<i>` im Flavortext, ein Manasymbol
mitten im Regeltext — dafür meldet der Planbauer nichts, weil nichts scheitert:
der Text wird gesetzt, nur falsch. Drei der letzten fünf Fehler dieses Projekts
waren von dieser Art — Schriftnamen, Verankerung, Symbolschriften —, am Bild zu
sehen und in keinem Test.

## Verworfen

| Möglichkeit | Warum nicht |
| --- | --- |
| **Weitermachen ohne Phasen** | Das ist der Zustand, der die vier Merges ohne Zweitblick erzeugt hat. Die Regelprozesse haben keinen anderen Anlass als eine Phasengrenze |
| **Die neun Phasen weiterzählen — Phase 10, 11, …** | Die neun tragen eine abgeschlossene Bauabfolge vom Gerüst zur Freigabe, und Phase 8 ist ausdrücklich die letzte (§12). Eine Phase 10 dahinter machte den Satz „nach Phase 8 wird freigegeben" unlesbar |
| **Eine einzige große Phase „Parität"** | Sie hätte ein einziges Abnahmekriterium, und das wäre das Ziel selbst. Eine Phase, deren Abnahme das Projektziel ist, endet nie und schärft nichts nach |
| **Phase A ans Ende, wenn es etwas zu vergleichen gibt** | Genau die Reihenfolge hat #240 zwei Sitzungen gekostet. Ohne Messung wird jede der Fragen aus #276 einzeln nach Gefühl entschieden |
| **Nach Befundgewicht ordnen — C, E, F, B, D** | Es ordnet die Arbeit nach dem, was der Zähler sieht, und stellt damit Phase D (null Befunde) ans Ende, obwohl sie am Bild sofort auffällt. Der Zähler ist ein Hilfsmittel, kein Maßstab; der Maßstab entsteht in Phase A |
| **Ein Referenzbild von MSE2 ins Repository legen** | §11. Es zeigte Rahmen, Symbole und Illustrationen von Wizards of the Coast. Woher das Vergleichsbild stattdessen kommt, entscheidet Phase A in einem eigenen Vermerk |

## Folgen

- **Der ausstehende Zweitblick über #274, #275, #277 und #278 gehört nicht in
  eine der acht Phasen.** Er ist Nachholarbeit an der Lücke, die dieser Vermerk
  schließt, und läuft vor Phase A.
- Acht Meilensteine und acht Labels kommen hinzu, `phase-a` bis `phase-h`. Die
  neun alten bleiben bestehen; sie tragen geschlossene Arbeit.
- **#248 wird
  keine Phase.** Es ist die Verbesserungsiteration selbst, also der Rahmen um
  A bis H, und ein Rahmen, der zugleich ein Teil ist, führt sich selbst.
- Die Zwei-Durchgänge-Regel aus [0006](0006-review-am-phasenende.md) gilt für
  jede der acht unverändert: ein Zweitblick, höchstens ein Nachlauf.
- **Freigegeben wird weiterhin nicht** (§12). Was §12 „alle neun Phasen" nennt,
  liest sich ab jetzt als „alle Phasen" — die acht eingeschlossen.
