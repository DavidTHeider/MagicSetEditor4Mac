# 0035 — Der Satz bleibt auflösungsunabhängig, MSEs ist es nicht

**Stand:** angenommen
**Datum:** 2026-08-20
**Berührt:** [`internal/render/zeilenraster.go`](../../internal/render/zeilenraster.go),
#612

## Lage

**MSEs Textraster hängt an der Vergrößerung, unseres nicht.** Angestrichen vom
Zweitblick über Phase K.

Beide Formeln stehen nachgelesen da, Fassung `f9d9356` vom 2021-07-06:

| | Rechnung | Fundstelle |
| --- | --- | --- |
| dieses Programm | `ceil(v · to_int(4·G)) / 4` | `zeilenraster.go` |
| MSE, allgemein | `ceil(v · to_int(4·z·G)) / (4·z)` | `rotation.cpp:326`, `rotation.hpp:63`, `rotation.cpp:349` |

**Unsere ist der Sonderfall `z = 1`.** Für den Bildexport ist das genau
richtig — dort setzt MSE selbst `zoom = 1` (`image.cpp:37–45`), und gegen
dessen Ausgabe wird der Abstand gemessen. Die Formel hat 21,553 → 20,692
gebracht (#546).

**Für den Druckbogen ist es eine Abweichung.** `bogen.go` rechnet aus Papier
und Auflösung eine Vergrößerung; bei einer 63-mm-Karte, 300 dpi und einem Plan
von 375 Punkten sind das **1,9843**. Dort rastert dieses Programm weiter auf
Viertelpunkte, MSE feiner.

## Was sie kostet — ausgerechnet

**2.528.805 Paare durchgerechnet**, Verhältnisse 0,600–2,000 und Größen
4,0–40,0, je Zeile:

| Bereich | 300 dpi | 600 dpi |
| --- | --- | --- |
| alle Verhältnisse und Größen | **0,513 pt** = 1,02 Bildpunkte | 0,480 pt = 1,90 Bildpunkte |
| übliche Größen 6–20 | 0,489 pt = 0,97 Bildpunkte | 0,469 pt = 1,86 Bildpunkte |
| Zeilenhöhe MPlantin (v = 1,1607) | **0,327 pt = 0,65 Bildpunkte** | 0,366 pt = 1,45 Bildpunkte |
| Grundlinie MPlantin (v = 0,7857) | 0,305 pt = 0,60 Bildpunkte | 0,277 pt = 1,10 Bildpunkte |

**Über einen sechszeiligen Regeltext summiert sich das auf knapp vier
Bildpunkte — 0,33 mm bei 300 dpi.** Es summiert sich, weil der Unterschied bei
fester Größe und festem Verhältnis deterministisch ist; er hat aber **kein
festes Vorzeichen** über Größen hinweg: er zittert, er verschiebt nicht.

### Eine Schranke von einem Viertelpunkt gilt nicht

**Das war die erste Annahme**, und sie klang zwingend: unser Raster ist das
gröbere, also müsste MSEs Wert darin liegen. Der Durchlauf hat sie widerlegt —
bis zu **0,51 pt**, gut zwei Viertelpunkte.

Der Grund: `to_int(4 · z · G)` rundet die **skalierte** Größe, und das Teilen
durch `4 · z` bringt den Fehler nicht auf unser Raster zurück. Die beiden
Rasterungen sind nicht ineinander geschachtelt.

### Was diese Zahlen nicht belegen

**Sie sind gerechnet, nicht gemessen** (§10). Sie folgen aus beiden Formeln;
MSEs steht mit drei Fundstellen da, aber seine **Ausgabe** bei `zoom ≠ 1` liegt
nicht vor. Wer sie hat, kann die Rechnung prüfen — und sollte es tun, bevor
jemand sie als Messung zitiert.

## Entscheidung

**Die Abweichung bleibt, und sie steht benannt in `zeilenraster.go`.**

Der Tausch ist beziffert: **0,33 mm auf einem gedruckten Regeltext** gegen die
Naht, die den ganzen Rasterpfad auflösungsunabhängig hält. §2 sagt es als
Bauform — der Plan steht in Vorlagenpunkten, die Auflösung entscheidet erst der
Rasterlauf; `SetzeAbschnitte` bekommt deshalb gar keine Vergrößerung. Diese
Naht für ein Drittelmillimeter aufzugeben wäre der schlechtere Tausch.

## Was die verworfenen Wege gekostet hätten

**Angleichen** — `SetzeAbschnitte` bekäme die Vergrößerung. Der Preis ist nicht
die Rechnung, sondern die Bauform: ein Plan, dessen Zeilenumbrüche von der
Zielauflösung abhängen, ist auf dem Bildschirm ein anderer als im Druck. Genau
das ist MSEs Eigenschaft, und sie ist keine, die man sich zulegt, ohne sie zu
brauchen.

**Nur dort angleichen, wo gegen MSE gemessen wird** — das ist kein dritter Weg,
sondern dieser: dort ist `zoom = 1`.

## Was das offen lässt

**Wenn eines Tages gegen eine MSE-Ausgabe bei `zoom ≠ 1` gemessen wird**, ist
diese Entscheidung neu zu treffen. Dann liegt eine Messung vor, und die schlägt
eine Rechnung.
