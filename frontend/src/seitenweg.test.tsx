// **Was eine Seite über ihren Weg sagt** — UX-03.AK-02
// (#1675).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Weg mit seiner Seite geht.** Wer den Tisch offen hat und auf
// eine andere Seite wechselt, soll dort nicht „Partie gegen Bo" über dem
// Kopf lesen. Der Aufräumer im Effekt ist die einzige Stelle, die das hält —
// die Seite selbst ist dann schon weg.
import { render, screen, within } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Rahmen } from './rahmen/Rahmen'
import { Sprachraum } from './sprache/kontext'
import { kataloge } from './sprache/kataloge'
import { useSeitenweg } from './seitenweg'

const texte = kataloge.de

// MitWeg ist eine Seite, die ihren Weg sagt.
function MitWeg() {
  useSeitenweg([{ beschriftung: 'Partie gegen Bo' }])

  return undefined
}

// weg liest die Stationen der Brotkrume ohne das Zeichen dazwischen.
function weg(): string[] {
  return within(
    screen.getByRole('navigation', { name: texte.allgemein.navigation.weg }),
  )
    .getAllByRole('listitem')
    .map((station) => station.textContent.replace('›', '').trim())
}

it('nimmt den Weg mit, wenn die Seite geht', () => {
  const rahmen = (kind: React.ReactNode) => (
    <Sprachraum sprache="de">
      <Rahmen aktuell="spielen" wechseln={vi.fn()}>
        {kind}
      </Rahmen>
    </Sprachraum>
  )

  const { rerender } = render(rahmen(<MitWeg />))

  expect(weg()).toEqual([texte.allgemein.bereiche.spielen, 'Partie gegen Bo'])

  rerender(rahmen(undefined))

  expect(weg()).toEqual([
    texte.allgemein.bereiche.spielen,
    texte.seiten.spielen.titel,
  ])
})
