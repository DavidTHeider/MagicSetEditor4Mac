// **Wie sich etwas bewegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum eine Karte fliegt
//
// **Damit man sieht, woher sie kam.** #1460 nennt „Kartenflug/Tap/
// Counter-Feedback": eine Karte, die am Zielort erscheint, hinterlässt die
// Frage, welche es war; eine, die hinfliegt, beantwortet sie unterwegs.
//
// # Und warum es abschaltbar ist
//
// **Weil Bewegung nicht jedem hilft.** `prefers-reduced-motion` ist die
// Ansage des Betriebssystems, und wer sie setzt, hat einen Grund — Schwindel,
// Migräne, Aufmerksamkeit. #1460 verlangt „abschaltbare Animationen", und die
// Systemeinstellung ist die erste Abschaltung, nicht die zweite.
//
// **Der fachliche Ereignisstrom bleibt derselbe** (#1460: „Animation aus/ein
// erzeugt denselben fachlichen Eventstrom"), und das ist hier keine Zusage,
// sondern eine Folge: eine Bewegung ist eine Sache der Darstellung, und die
// Darstellung schickt nichts an den Wirt.

// Bewegungsdauer ist, wie lange eine Karte unterwegs ist — in Millisekunden.
//
// **220 ms**: kurz genug, dass niemand wartet, lang genug, dass ein Auge
// folgen kann. Die Zahl ist gewählt und nicht gemessen; wer sie ändert, ändert
// eine Gestaltung und keine Zusage.
export const Bewegungsdauer = 220

// Beweglich ist, was sich bewegen kann.
export interface Beweglich {
  x: number
  y: number
  drehung: number
}

// weich beschleunigt und bremst.
//
// **Eine Karte, die mit voller Geschwindigkeit anfängt und abrupt hält, sieht
// aus wie ein Sprung mit Zwischenbildern.** Die Kurve ist das gewöhnliche
// „ease in out" als Kosinus — sie braucht keine Bibliothek und keine
// Magiezahlen.
export function weich(anteil: number): number {
  const gedeckelt = Math.min(1, Math.max(0, anteil))

  return (1 - Math.cos(gedeckelt * Math.PI)) / 2
}

// zwischen gibt den Stand einer Bewegung.
//
// **Die Drehung nimmt den kürzeren Weg.** Von 350° nach 10° sind es zwanzig
// Grad und nicht dreihundertvierzig; ohne diese Rechnung drehte sich eine
// enttappte Karte einmal ganz herum.
export function zwischen(
  von: Beweglich,
  nach: Beweglich,
  anteil: number,
): Beweglich {
  const t = weich(anteil)
  const unterschied = (((nach.drehung - von.drehung) % 360) + 540) % 360

  return {
    x: von.x + (nach.x - von.x) * t,
    y: von.y + (nach.y - von.y) * t,
    drehung: von.drehung + (unterschied - 180) * t,
  }
}

// lohntBewegung sagt, ob sich das Bewegen überhaupt lohnt.
//
// **Ein halber Millimeter ist keine Bewegung**, sondern ein Rundungsfehler —
// und ein Ticker, der für ihn anspringt, kostet sechzig Bilder für nichts.
export function lohntBewegung(von: Beweglich, nach: Beweglich): boolean {
  return (
    Math.abs(nach.x - von.x) > 1 ||
    Math.abs(nach.y - von.y) > 1 ||
    Math.abs(nach.drehung - von.drehung) > 1
  )
}

// bewegungErwuenscht sagt, ob animiert werden soll.
//
// **Zwei Stimmen, und die Systemeinstellung hat die stärkere.** Wer
// `prefers-reduced-motion` setzt, bekommt keine Bewegung — auch dann nicht,
// wenn die Anwendung sie anböte. Umgekehrt darf abschalten, wer will.
//
// **Und die Einstellung der Anwendung** — UX-10.AK-03
// (#1682):
// „Bewegung: Immer wenig" steht als `data-bewegung` am Wurzelelement, und
// dieselbe Angabe schaltet in `barrierefreiheit.css` jeden Übergang ab. Bis zum
// 2026-09-25 flogen die Karten trotzdem, solange das System nichts sagte —
// die Einstellung versprach etwas, das nur für die Oberfläche galt.
export function bewegungErwuenscht(
  fenster:
    | {
        matchMedia?: (frage: string) => { matches: boolean }
        document?: { documentElement: { dataset: DOMStringMap } }
      }
    | undefined,
  eingeschaltet: boolean,
): boolean {
  if (!eingeschaltet) return false

  if (fenster?.document?.documentElement.dataset.bewegung === 'reduziert') {
    return false
  }

  const antwort = fenster?.matchMedia?.('(prefers-reduced-motion: reduce)')

  return antwort ? !antwort.matches : true
}
