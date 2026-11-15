// **Laden und Scheitern einer Seite, an einer Stelle** — G11
// (#2277).
//
// B03 fand sechs Seiten ohne Ladehinweis und drei ohne Meldung, wenn eine
// Bindung, die sie beim Öffnen rufen, langsam ist oder scheitert. Der Baustein
// nimmt die Ladezustände einer Seite und sagt, **dass** sie lädt, und nach
// einem Scheitern, **was** scheiterte — mit einem Rückweg.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import type { Ladezustand } from '../haken/useLaden'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Seitenladung } from './Seitenladung'

const zustaende = kataloge.de.bausteine.zustaende

function zustand(
  teil: Partial<Ladezustand<unknown>>,
  nachgeladen?: string[],
  name = '',
): Ladezustand<unknown> {
  return {
    daten: undefined,
    laedt: false,
    fehler: undefined,
    nachladen: () => {
      nachgeladen?.push(name)
    },
    ...teil,
  }
}

function zeige(...liste: Ladezustand<unknown>[]) {
  return render(
    <Sprachraum sprache="de">
      <Seitenladung zustaende={liste} />
    </Sprachraum>,
  )
}

describe('Seitenladung', () => {
  it('sagt, dass die Seite lädt, solange eine Bindung aussteht', () => {
    zeige(zustand({ daten: [] }), zustand({ laedt: true }))

    const hinweis = screen.getByRole('status')
    expect(hinweis).toHaveTextContent(zustaende.laedt)
    expect(hinweis.closest('[aria-busy]')).toHaveAttribute('aria-busy', 'true')
  })

  it('nennt das Scheitern und lädt nur nach, was scheiterte', async () => {
    const nachgeladen: string[] = []

    zeige(
      zustand({ daten: [] }, nachgeladen, 'gut'),
      zustand(
        { fehler: 'Die Felder ließen sich nicht lesen.' },
        nachgeladen,
        'felder',
      ),
      zustand({ laedt: true }, nachgeladen, 'langsam'),
    )

    // **Das Scheitern geht vor dem Laden**: wer wartet, soll nicht auf etwas
    // warten, das schon gescheitert ist.
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Die Felder ließen sich nicht lesen.',
    )
    expect(screen.queryByRole('status')).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: zustaende.wiederholen }),
    )
    expect(nachgeladen).toEqual(['felder'])
  })

  it('schweigt, wenn alles da ist', () => {
    zeige(zustand({ daten: [] }), zustand({ daten: 'x' }))

    expect(screen.queryByRole('status')).toBeNull()
    expect(screen.queryByRole('alert')).toBeNull()
    expect(document.querySelector('[aria-busy]')).toBeNull()
  })
})
