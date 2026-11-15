// **Die Setangaben in Gruppen** — UXR-07
// (#1915).
//
// # Woher die Gruppen kommen
//
// **Aus dem Spiel, nicht von uns.** `magic.mse-game` setzt zwischen seine 89
// Setfelder zehn Auskünfte (`type: info`) — „Set Information“, „Language“,
// „Card Numbering“ … (gezählt am 2026-10-02 in `set_fields` des
// Full-Magic-Pack). MSE zeichnet jede als **Kopfzeile** über die Breite des
// Editors, mit ihrer Aufschrift als Text (`InfoValueViewer::draw` in
// `src/render/value/information.cpp`, `InfoValue::update` in
// `src/data/field/information.cpp`, Fassung f9d9356 vom 2021-07-06). Eine
// Kopfzeile überschreibt, was ihr folgt — das ist die Gruppe.
//
// # Was hier anders ist
//
// **Eingeklappt wird bei MSE nichts**; hier stehen alle Gruppen außer der
// ersten hinter ihrer Überschrift — **adapted**, nach R3 aus
// `docs/UIREVIEW.md`: höchstens zwölf gezeichnete Felder, sonst gruppiert und
// eingeklappt bis auf eine Gruppe.
import type { FeldbeschreibungDTO } from '../api'

const auskunft = 'info'

// Setangabengruppe sind die Felder unter einer Überschrift.
export interface Setangabengruppe {
  kopf: FeldbeschreibungDTO
  felder: FeldbeschreibungDTO[]
}

interface Gegliedert {
  // erste steht offen — mit ihrer Überschrift, wenn die Felder mit einer
  // beginnen, und ohne, wenn vor der ersten Überschrift schon Felder stehen.
  erste: {
    kopf: FeldbeschreibungDTO | undefined
    felder: FeldbeschreibungDTO[]
  }

  // weitere stehen eingeklappt hinter ihrer Überschrift.
  weitere: Setangabengruppe[]
}

// gliedere teilt die Setfelder an ihren Auskünften.
export function gliedere(felder: readonly FeldbeschreibungDTO[]): Gegliedert {
  const erste: Gegliedert['erste'] = { kopf: undefined, felder: [] }
  const weitere: Setangabengruppe[] = []
  let aktuell: { felder: FeldbeschreibungDTO[] } = erste

  for (const form of felder) {
    if (form.art !== auskunft) {
      aktuell.felder.push(form)
      continue
    }

    // **Eine Überschrift ganz vorn überschreibt die erste Gruppe** — jede
    // andere beginnt eine neue.
    if (erste.kopf === undefined && erste.felder.length === 0) {
      erste.kopf = form
      continue
    }

    const gruppe: Setangabengruppe = { kopf: form, felder: [] }
    weitere.push(gruppe)
    aktuell = gruppe
  }

  return { erste, weitere }
}
