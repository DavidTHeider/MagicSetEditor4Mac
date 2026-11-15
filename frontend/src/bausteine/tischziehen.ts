// **Wie eine Karte am Zeiger hängt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Eine eigene Datei**, seit `Tischflaeche.tsx` über die Zeilengrenze aus §2
// lief. Der Schnitt folgt der Frage: dort steht, **was die Fläche ist** —
// Leinwand, Kamera, Tischbild, Knöpfe; hier, **was beim Ziehen geschieht**.
//
// **Der Zustand bleibt drüben.** Diese Datei rechnet und entscheidet; wer den
// Zustand hält, ist die Fläche. Damit ist jede Regel hier ohne einen Baum
// prüfbar.
import type { Kamera } from '../tisch/kamera'
import type { Tischplan } from '../tisch/tischplan'

// Ziehen ist ein laufender Zug samt dem Punkt, an dem er begann.
export interface Ziehen {
  kennung: string
  vonX: number
  vonY: number
  dx: number
  dy: number
}

// Zeigerort ist ein Punkt in Weltkoordinaten.
export interface Zeigerort {
  x: number
  y: number
}

// angefasst gibt den Zug, der mit diesem Druck beginnt — oder keinen.
//
// **Nur die Hauptzeigertaste zieht.** Die rechte öffnet anderswo ein Menü,
// und ein Zug daran wäre eine Überraschung.
export function angefasst(
  taste: number,
  kennung: string | undefined,
  ort: Zeigerort,
): Ziehen | undefined {
  if (taste !== 0 || kennung === undefined) {
    return undefined
  }

  return { kennung, vonX: ort.x, vonY: ort.y, dx: 0, dy: 0 }
}

// geschoben gibt den Zug mit neuem Versatz.
export function geschoben(laufend: Ziehen, ort: Zeigerort): Ziehen {
  return { ...laufend, dx: ort.x - laufend.vonX, dy: ort.y - laufend.vonY }
}

// Zugziel ist, was ein losgelassener Zug meldet — oder nichts.
//
// **Drei Arten und ein unterscheidender Verbund.** Wer `art === 'haufen'`
// liest, hat eine Karte, auf die gelegt wird, und keine Zone; der Übersetzer
// hält das, und der Aufrufer prüft es nicht noch einmal.
export type Zugziel =
  // **`frei` ist ⌥ beim Loslassen** — P5 (#2373): ausdrücklich schieben
  // statt deuten lassen, wo das Profil es erlaubt.
  | { art: 'zone'; kennung: string; zone: string; sitz: string; frei?: true }
  | { art: 'haufen'; kennung: string; worauf: string }
  | { art: 'enthaufen'; kennung: string }

  // **`ablegen` ist kein Spielbefehl**, und das ist der Unterschied zu den
  // dreien darüber: eine Lage in Millimetern ist Darstellung, keine Regel.
  // CR 400.5 lässt die Anordnung frei — der Wirt hat dazu nichts zu sagen,
  // und ein Gegenüber bekommt davon nichts zu sehen. PLAY-04.AK-03
  // (#1712).
  | { art: 'ablegen'; kennung: string; x: number; y: number }

// losgelassen sagt, was beim Loslassen zu melden ist.
//
// # Der Zug in eine andere Zone
//
// **Das war bis zum 2026-09-21 die einzige Meldung**, und ein Zug innerhalb
// derselben Zone tat nichts: CR 400.5 lässt die Anordnung frei, und der Wirt
// hatte nichts zu entscheiden.
//
// # Und die Anordnung, die er seitdem trägt
//
// **Eine Karte auf eine andere zu legen ist die Geste für einen Haufen**
// (#1756).
// Sie liegt dort, wo vorher nichts geschah — und die Gegenbewegung, das
// Ziehen auf freien Platz, nimmt sie wieder heraus.
//
// **Nichts, wo das Ziel keine Zone ist**: neben den Feldern liegt Tisch,
// und ein Wurf dorthin ist kein Zug.
export function losgelassen(
  plan: Tischplan,
  laufend: Ziehen,
  ziel: { zone: string; sitz: string } | undefined,
  worauf?: string,

  // ablegen sagt, dass der Tisch **nicht** selbst ordnet — PLAY-04.AK-03
  // (#1712).
  //
  // **Ein Schalter und keine Stelle.** Wo die Karte landet, weiß diese
  // Funktion besser als ihr Aufrufer: sie hat die bisherige Lage und den
  // gezogenen Weg. Wer die Stelle hereinreichte, rechnete sie ein zweites
  // Mal — und die zweite Rechnung liefe irgendwann auseinander (§10).
  ablegen?: boolean,
): Zugziel | undefined {
  const bisher = plan.lagen.find((eine) => eine.kennung === laufend.kennung)

  if (!ziel || !bisher) {
    return undefined
  }

  if (ziel.zone !== bisher.zone || ziel.sitz !== bisher.sitz) {
    return {
      art: 'zone',
      kennung: laufend.kennung,
      zone: ziel.zone,
      sitz: ziel.sitz,
    }
  }

  return inDerselbenZone(
    plan,
    bisher,
    laufend,
    worauf,
    ablegen === true
      ? { x: bisher.x + laufend.dx, y: bisher.y + laufend.dy }
      : undefined,
  )
}

// inDerselbenZone sagt, was ein Zug bedeutet, der die Zone nicht verlässt.
//
// **Drei Lagen und ein Ablegen**: auf eine andere Karte legen, aus dem
// Haufen nehmen, oder — wo beides nicht zutrifft — die Lage festhalten.
//
// # Wann die Lage festgehalten wird
//
// **Wenn der Tisch nicht selbst ordnet** — die Anordnung ist abschaltbar
// (PLAY-04.AK-03,
// #1712).
// Ordnet er selbst, wäre eine festgehaltene Lage beim nächsten Zug
// überschrieben; `stelle` sagt, welcher der beiden Fälle gerade gilt.
//
// **Aus dem Haufen genommen wird weiterhin zuerst.** Das ist eine Aussage
// über die Anordnung im Sinne von CR 400.5 und keine über Millimeter; sie
// gilt, ob der Tisch selbst ordnet oder nicht.
function inDerselbenZone(
  plan: Tischplan,
  bisher: { haufen: string },
  laufend: Ziehen,
  worauf: string | undefined,
  stelle: { x: number; y: number } | undefined,
): Zugziel | undefined {
  const kennung = laufend.kennung

  if (worauf === undefined) {
    // **Neben den Haufen gezogen heißt: heraus damit.**
    if (bisher.haufen !== '') {
      return { art: 'enthaufen', kennung }
    }

    return stelle === undefined
      ? undefined
      : { art: 'ablegen', kennung, x: stelle.x, y: stelle.y }
  }

  // **Auf sich selbst legt niemand etwas.** Beim Loslassen steht der Zeiger
  // oft noch über dem Platz, von dem die Karte kam.
  if (worauf === kennung) {
    return undefined
  }

  const ziel = plan.lagen.find((eine) => eine.kennung === worauf)

  // **Und nichts, wo sie schon auf demselben Haufen liegt** — ein Befehl
  // dafür ginge gegen einen Stand, der schon gilt (§7, Regel 4).
  if (!ziel || (bisher.haufen !== '' && bisher.haufen === ziel.haufen)) {
    return undefined
  }

  return { art: 'haufen', kennung, worauf }
}

export type { Kamera }
