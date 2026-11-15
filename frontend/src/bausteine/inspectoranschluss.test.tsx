// **Der Inspector gibt weiter, was er zeigt** — UX-03.AK-01
// (#1675).
//
// **Eine eigene Datei**, damit `sammlungsinspector.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, was
// man am Exemplar **tut**, hier, wohin es **weiterreist**.
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import { kataloge } from '../sprache/kataloge'
import { Sammlungsinspector } from './Sammlungsinspector'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'

import type { ExemplarDTO } from '../api'

const texte = kataloge.de.seiten.sammlung

const dasExemplar: ExemplarDTO = {
  id: 1,
  fassung: 1,
  printing: 'p-4',
  druckID: '',
  vorlaeufig: false,
  menge: 3,
  zustand: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
  erwerbsdatum: '',
  erwerbspreis: '',
  erwerbswaehrung: '',
  notiz: '',
  name: 'Aschenwächter',
  setcode: '',
  nummer: '',
  lagerort: '',
  wunsch: false,
  bezeichnung: 'p-4',
  herkunft: keineHerkunft,
}

function zeige(gewaehlt: ExemplarDTO | undefined) {
  belegeApi({
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsinspector
        exemplar={gewaehlt}
        inSicht
        aufAendern={vi.fn(() => Promise.resolve(true))}
        aufLoeschen={vi.fn(() => Promise.resolve(true))}
        aufGeholt={vi.fn()}
        laeuft={false}
      />
    </Sprachraum>,
  )
}

describe('die Anschlüsse am Sammlungsinspector', () => {
  it('bietet beide Wege am gewählten Exemplar an', () => {
    zeige(dasExemplar)

    expect(
      screen.getByRole('button', { name: texte.zumTausch }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: texte.insDeck }),
    ).toBeInTheDocument()
  })

  // **Ohne Wahl gibt es nichts weiterzugeben.** Ein Knopf, der ein Objekt
  // mitnimmt, das es nicht gibt, wäre ein Weg ins Leere.
  it('bietet ohne Wahl keinen an', () => {
    zeige(undefined)

    expect(
      screen.queryByRole('button', { name: texte.zumTausch }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.insDeck }),
    ).not.toBeInTheDocument()
  })
})
