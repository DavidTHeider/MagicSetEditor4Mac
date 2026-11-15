# 0164 — Der Wirt sitzt im gezeigten Profil, wenn niemand anderes gewählt ist

**Stand:** angenommen am 2026-10-03 · UXR-F21
(#2041),
Folgebefund aus UXR-F16 (#2031).
Marke `new`: MSE2 spielt nicht (§11, gemessen für G1), und eine Spielregel ist
es auch nicht — die *Comprehensive Rules* kennen keine Profile. Die Frage
gehört dem Produkt.

## Die Frage

**Ob man in einer Partie dran ist, hängt am eigenen Sitz**
(`service.eigenerSitz`), und einen eigenen Sitz hatte nur, wer einen Sitz an
sein Profil gebunden hat. Binden ließ sich am Fenster nur mit **zwei**
Profilen — mit einem, dem Normalfall, erschien die Marke „dran" nie. An
Kommandozeile, MCP und REST kam die Bindung erst mit #2041 hinzu, und auch
dort nur, wenn jemand sie nennt.

Offen war, ob der Start den **ersten Sitz** von selbst binden soll. Dagegen
sprach die Zuschauerpartie — zwei Menschen am Küchentisch, keiner ist „ich";
dafür, dass der Tisch einem ungebundenen Profil ohnehin den Wirt zeigt.

## Die Entscheidung

**Bindet kein Sitz ein Profil, sitzt das gezeigte Profil am Sitz des
Wirts.** Der Wirt ist der erste Sitz, wenn der Aufrufer keinen nennt. Die
Regel steht im Dienst (`service.WirtImGezeigtenProfil`) und wirkt damit an
allen vier Flächen gleich (§16).

**Sie tut nichts**, sobald irgendein Sitz gebunden ist, ohne gezeigtes
Profil, und wenn der Wirt nicht am Tisch sitzt.

## Warum sie keine Sicht ändert

**Der Tisch zeigt einem ungebundenen Profil den Wirt**
(`domain.Partie.SichtFuerProfil`, G8). Wer den Wirt an das gezeigte Profil
bindet, gibt ihm den Sitz, den es schon sah — und einem zweiten Profil weiter
den Wirt. **Neu ist nur, dass die Partienliste ihn als den eigenen kennt.**
Die Zuschauerpartie verliert dadurch nichts: am Küchentisch sieht man den
Tisch wie zuvor, und „dran" nennt den Sitz, dessen Sicht man hat.

## Was sichtbar wird

Am Fenster mit **einem** Profil gibt es keine Wahl, und die Vorgabe greift
still — sie zeigt sich an der Marke „dran". Mit **zwei** Profilen steht die
Wahl am Sitz, und die Erklärung darunter sagt jetzt, was ohne Wahl gilt
(„Wählst du für keinen Sitz ein Profil, sitzt dein gezeigtes Profil am ersten
Sitz") — sonst stünde dort „kein Profil", und die Liste nennte danach einen
eigenen Sitz.

## Einordnung nach 0033

**Verträglich** an Kommandozeile, MCP und REST: Eingaben und Felder bleiben;
eine Partie, die ohne Bindung angelegt wird, trägt danach eine — die Liste
nennt `eigenerSitz`, wo sie vorher leer war. Wer „zusehen" will, bindet einen
Sitz an ein anderes Profil.

## Verworfen

- **Den ersten Sitz statt des Wirts.** Am Fenster ist das dasselbe; über eine
  Leitung angelegt hält der Wirt die Partie auf diesem Rechner, und der erste
  Sitz könnte ein Gast sein.
- **Nur bei genau einem Profil.** Der Dienst müsste die Profile des Nutzers
  zählen; und mit zwei Profilen ohne Wahl stünde wieder niemand am Tisch —
  dieselbe Lücke, nur seltener.
- **„Eigen" ohne Bindung aus der Sicht ableiten** (`SichtFuerProfil` statt
  `SitzeVonProfil`). Für eine Partie, die über eine Leitung kam, ist der Wirt
  das Gegenüber; die Liste nennte dann dessen Sitz „dein Sitz".
