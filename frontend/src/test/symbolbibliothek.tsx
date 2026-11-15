// **Vorrichtungen für die Symbolbibliothek** — DOM-02
// (#1697).
// Geteilt von den beiden Testdateien, damit keine von ihnen über die
// Zeilengrenze aus §2 geht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import type { BibliothekssymbolDTO } from '../apisymbolbibliothek'
import { Symbolbibliothek } from '../seiten/Symbolbibliothek'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'

export const krone: BibliothekssymbolDTO = {
  id: '1',
  name: 'Krone',
  abdruck: 'sha256:krone',
  angelegt: '2026-09-27T10:00:00Z',
  geaendert: '2026-09-27T10:00:00Z',
  vorschau: 'data:image/png;base64,S1JPTkU=',
  verwendet: ['Kronenset', 'Leerset'],
  vorige: false,
}

export const zacke: BibliothekssymbolDTO = {
  id: '2',
  name: 'Zacke',
  abdruck: 'sha256:zacke',
  angelegt: '',
  geaendert: '',
  vorschau: '',
  verwendet: [],
  vorige: false,
}

// mitSymbolen ist die Antwort der Liste.
export const mitSymbolen = (...symbole: BibliothekssymbolDTO[]) => ({
  symbole,
  fehler: null,
})

// ohneMarken ist die Antwort, die der Markenfilter beim Aufschlagen holt.
export const ohneMarken = () => Promise.resolve({ marken: {}, fehler: null })

// symbolantwort ist eine geglückte Antwort eines Weges, der ein Symbol nennt.
export const symbolantwort = (
  angaben: Partial<{
    id: string
    name: string
    pfad: string
    abgebrochen: boolean
  }>,
) => ({
  id: '',
  name: '',
  pfad: '',
  abgebrochen: false,
  fehler: null,
  ...angaben,
})

// zeigeBibliothek setzt die Seite für das Set 7 auf.
export function zeigeBibliothek(eingesetzt = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Symbolbibliothek setID="7" eingesetzt={eingesetzt} />
    </Sprachraum>,
  )

  return eingesetzt
}

// bibliothekAufklappen öffnet die Bibliothek im Fach *Symbol*.
//
// **Sie steht dort seit UXR-07 eingeklappt** (#1915): die Lasche maß 4,1
// Bildschirme, und der Editor ist ihre führende Handlung. Wer im Fach etwas
// aus der Bibliothek braucht, klappt sie auf — dieser Lauf ebenso.
export async function bibliothekAufklappen() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.karten.symbolbibliothek.ueberschrift,
    }),
  )
}
