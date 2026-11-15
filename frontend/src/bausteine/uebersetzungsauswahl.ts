// **Welche Sätze der Editor einer eigenen Übersetzung zeigt** — LOC-02
// (#1688).
//
// Über viertausend Sätze stehen im Katalog; gezeigt wird eine Seite davon,
// ausgewählt nach Stand und Suche. **Fehlend und veraltet sagt Go** — hier
// wird nur gefiltert und geblättert.

import type { UebersetzungstextDTO } from '../apisprache'

export type Uebersetzungsfilter = 'fehlend' | 'veraltet' | 'uebersetzt' | 'alle'

// zeilenJeSeite ist, was eine Seite zeigt — genug, um zu arbeiten, wenig
// genug, dass jedes Feld ein eigenes Bauteil sein darf.
export const zeilenJeSeite = 25

export interface Uebersetzungsstand {
  texte: Record<string, UebersetzungstextDTO>
  fehlendePfade: readonly string[]
  veraltetePfade: readonly string[]
}

export interface Uebersetzungsseite {
  pfade: string[]
  seite: number
  seiten: number
}

// auswahl gibt die Pfade einer Seite — geordnet, gefiltert und gesucht.
export function auswahl(
  vorlage: Readonly<Record<string, string>>,
  stand: Uebersetzungsstand,
  filter: Uebersetzungsfilter,
  suche: string,
  seite: number,
): Uebersetzungsseite {
  const nachStand: Record<Uebersetzungsfilter, readonly string[]> = {
    fehlend: stand.fehlendePfade,
    veraltet: stand.veraltetePfade,
    uebersetzt: Object.keys(stand.texte),
    alle: Object.keys(vorlage),
  }

  const frage = suche.trim().toLowerCase()
  const pfade = [...nachStand[filter]]
    .filter((pfad) =>
      [pfad, vorlage[pfad] ?? '', stand.texte[pfad]?.text ?? ''].some((text) =>
        text.toLowerCase().includes(frage),
      ),
    )
    .sort()

  const seiten = Math.max(1, Math.ceil(pfade.length / zeilenJeSeite))
  const geklemmt = Math.min(Math.max(1, seite), seiten)

  return {
    pfade: pfade.slice(
      (geklemmt - 1) * zeilenJeSeite,
      geklemmt * zeilenJeSeite,
    ),
    seite: geklemmt,
    seiten,
  }
}
