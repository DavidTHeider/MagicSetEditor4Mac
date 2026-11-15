# 0126 — Der Tisch zeichnet nicht optimistisch

**Stand:** angenommen · **Datum:** 2026-09-22
**Berührt:** `AGENTS.md` §6 („Optimistische Aktualisierung mit Rücknahme bei
Fehlern"), §7 (die vier Regeln für nebenläufige Änderungen),
#1712
(PLAY-04.AK-04),
#1460 (G3,
der gezeichnete Tisch)

## Warum es diesen Vermerk gibt

**Weil §6 das Gegenteil verlangt** — „optimistische Aktualisierung mit
Rücknahme bei Fehlern" — und der Tisch es seit G3 nicht tut. Eine Abweichung,
die niemand aufschreibt, ist keine Entscheidung, sondern ein Versäumnis, das
wie eine aussieht (§10).

**Und weil PLAY-04.AK-04 sie zur Frage macht:** *„Ungültige Aktionen werden
erklärt und visuell zurückgenommen. Optimistische Darstellung muss mit dem
bestätigten Partiestand wieder übereinstimmen, auch nach Verbindungsabbruch."*
Wer das prüft, ohne zu wissen, dass hier nichts vorweggenommen wird, prüft
eine Rücknahme, die es nicht geben kann — und hält ihr Ausbleiben für einen
Befund.

## Die Entscheidung

**Was der Tisch zeichnet, ist der bestätigte Stand und nichts sonst.** Ein
Befehl geht hinaus, der Kern antwortet mit einer neuen Sicht, und erst
**diese** wird gezeichnet. Zwischen Griff und Bild liegt die Antwort.

## Warum

**Weil der Kern die Regeln hält und nicht das Fenster.** Ob eine Karte
gespielt werden darf, entscheiden `service` und `domain` — CR 601 für das
Wirken, CR 506 für den Kampf, CR 704 für die zustandsbasierten Aktionen. Ein
Fenster, das die Karte schon einmal hinlegte, müsste **dieselben Regeln ein
zweites Mal kennen**, um zu wissen, ob es sie gleich wieder wegnehmen muss.
Zwei Regelausleger sind einer zu viel: der zweite ist immer der veraltete.

**Weil ein halber Stand schlimmer ist als ein später.** Ein Zug am Tisch ist
selten eine Änderung: „lege sie getappt aufs Feld" ist ein Ereignis mit drei
Wirkungen, und ein Schieben in eine geordnete Zone verschiebt alles darunter.
Wer das vorwegnähme und zurückrollen müsste, hätte zwischendurch ein Brett,
das die Karte nie gemeint hat — sichtbar, und bei einem Gast auch fotografiert.

**Weil ein Gast gar nicht rechnet.** Er hat keinen Vollzustand und keinen
Regelausleger; er bekommt nach jedem Zug das **Tischbild**, das der Wirt für
ihn gerechnet hat (G5,
#1462). Eine
optimistische Darstellung wäre damit am Wirtsfenster eine und am Gastfenster
keine — zwei Tische, zwei Wahrheiten, und die Regel aus §16 („dieselbe
Fähigkeit, überall dieselbe Antwort") gilt hier zwischen zwei Sitzen
desselben Spiels.

**Und weil §6 dorthin zielt, wo es trägt.** Der Satz steht dort wegen der
Listen und Formulare: ein Name, der sich ändert, ein Häkchen, das umspringt —
Dinge, deren Gültigkeit das Fenster selbst kennt, und deren Rücknahme genau
einen Datensatz betrifft (§7, letzter Absatz). Am Tisch betrifft sie das
ganze Brett.

## Was daraus folgt

| | |
| --- | --- |
| **Erklärt wird** | am Tisch, nicht am Kopf der Seite. `Partietisch.tsx` zeigt `lauf.fehler` als `alert`; `Spielen.tsx` schweigt, solange ein Tisch offensteht — zwei gleiche Meldungen ließen jemanden zwei Fehler suchen |
| **Zurückgenommen wird** | nichts, weil nichts vorweggenommen wurde |
| **Nach einem Verbindungsabbruch** | sieht ein Gast **nichts mehr**, auch nicht, was zuletzt galt — das steht so in [`FAEHIGKEITEN.md`](../FAEHIGKEITEN.md) und ist die Folge derselben Entscheidung: er zeigt keinen Stand, den ihm niemand bestätigt hat |
| **Der Kartenflug** | ist keine Ausnahme. Er zeichnet die Bewegung **zum bestätigten Ziel**, nachdem die Antwort da ist (`tisch/bewegung.ts`); er nimmt sie nicht vorweg |
| **Die eigene Lage** | ist eine: wer bei abgeschalteter Autoanordnung eine Karte hinlegt, sieht sie sofort dort (PLAY-04.AK-03, `useEigeneLagen`). **Sie ist aber kein Spielzustand** — sie geht an keinen Kern, kann von keinem abgewiesen werden und liegt im Browserspeicher dieses Fensters. Es gibt nichts zurückzunehmen, weil nichts gefragt wurde |

## Was verworfen wurde

**Optimistisch nur für das, was sicher gilt** — Tappen etwa. Verworfen, weil
„sicher" eine Regelauskunft ist: CR 302.6 kennt Kreaturen, die nicht tappen
können, und eine Fähigkeit auf dem Brett kann es morgen ändern. Die Liste der
sicheren Fälle wäre der zweite Regelausleger aus dem Absatz oben, nur kleiner
— und damit unauffälliger falsch.

## Was dieser Vermerk nicht entscheidet

**Ob ein Warten sichtbar ist.** Dass zwischen Griff und Bild eine Antwort
liegt, heißt nicht, dass man es merken soll; was dabei angezeigt wird —
angehaltene Knöpfe, ein Takt —, ist eine Frage der Gestaltung und steht in
den Bauteilen.

**Und ob das immer so bleibt.** Über eine Leitung mit spürbarer Laufzeit kann
die Frage neu gestellt werden. Dann ist sie neu zu beantworten, mit einer
Messung — nicht mit diesem Vermerk.

## Nachtrag 2026-10-18 — ein wartender Wunsch (P5)

**Fortgeschrieben mit P5**
(#2373),
Befunde C-21 und C-24 aus dem Audit zu 0.9.0. Bis dahin verwarf das Fenster
einen zweiten Befehl, solange der erste unterwegs war, und drei schnelle
Gesten hinterließen eine — still.

| | |
| --- | --- |
| **Ein Wunsch wartet** | statt verworfen zu werden (`usePartiebefehle`, `useSchlange`): einer ist unterwegs, die übrigen stehen in Reihenfolge an. Der nächste rechnet gegen die Fassung, die der vorige **hinterlassen** hat, nicht gegen die beim Einreihen gezeigte; was der Kern dann abweist, kommt mit Grund zurück |
| **Ein wartender Wunsch ist kein Spielstand** | gezeichnet wird weiter nur, was der Kern bestätigt hat. Die Entscheidung oben bleibt unverändert |
| **Aber er ist zu sehen** | „Wartet auf den Tisch: n“ in einer Live-Region am Tisch — damit ist die erste offene Frage unter „Was dieser Vermerk nicht entscheidet“ für den Fall vieler Wünsche beantwortet |
| **Eine Ablehnung bleibt stehen** | bis eine neue Folge von Wünschen beginnt — sonst überschriebe der nächste Wunsch der Folge den Grund des vorigen |
| **Ein Wechsel räumt** | Partie oder Sitz gewechselt: was aus dem alten Stand wartet, geht nicht mehr hinaus (§7, Regel 1 und 2) |
| **Die Geste deutet der Kern** | `objekt.ziehen` (P3, #2371) — das Fenster legt fest, wohin gezogen wurde, nicht, was es heißt |
