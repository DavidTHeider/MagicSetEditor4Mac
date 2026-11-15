// **Die Griffe des Karteneditors ins Formular** — Werkstück U5 von Phase 5.1
// (#1416)
// hat sie aus `Karteneditor.tsx` herausgelöst, damit die Datei unter der
// Zeilengrenze bleibt (§2). Beide greifen über das Formular auf einen Knoten
// zu — **über das Formular und nicht über eine Kennung**: zwei Karten
// desselben Sets nebeneinander gäbe es zweimal `#feld-name`; das Formular
// ist das dieser Karte.
import { symbolEinsetzen } from './symboleinfuegen'

// feldFokussieren setzt den Fokus in das Eingabefeld eines Kartenfeldes.
//
// **Ein Feld ohne Eingabe tut nichts.** Ein Feldstil kann eine Ebene für
// etwas zeichnen, das die Karte nicht führt; dann gibt es nichts zu
// fokussieren, und ein Fehler wäre die falsche Antwort.
export function feldFokussieren(
  formular: HTMLFormElement | null,
  feld: string,
) {
  formular
    ?.querySelector<HTMLElement>(`[aria-label="${CSS.escape(feld)}"]`)
    ?.focus()
}

// symbolEinfuegenIn setzt ein Symbol an der Schreibmarke des Feldes ein und
// gibt den neuen Wert — oder nichts, wenn es das Feld nicht gibt.
//
// **Über den DOM-Knoten**, denn die Eingaben sind unkontrolliert: der
// Entwurf steht im Feld, nicht im Zustand. Wer nur schriebe, hätte den neuen
// Wert im Bestand und den alten vor Augen.
//
// **`roh` kommt von der Aufrufstelle.** Dort steht die geladene Palette
// ohnehin fest; hier noch einmal danach zu fragen ergäbe einen Rückfall, den
// kein Aufruf betreten kann — eine Zusage ohne Gegenprobe (§10).
export function symbolEinfuegenIn(
  formular: HTMLFormElement | null,
  feld: string,
  wertRoh: string,
  code: string,
  roh: boolean,
): string | undefined {
  return symbolEinsetzen(
    formular?.querySelector(`[aria-label="${CSS.escape(feld)}"]`),
    wertRoh,
    code,
    roh,
  )
}
