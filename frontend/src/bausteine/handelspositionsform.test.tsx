// **Eine Zeile über den Namen der Karte** — P10
// (#2378),
// Befund C-13: wer eine Karte der Gegenseite anlegte, musste ihre
// Druckkennung kennen.
//
// # Wogegen hier geprüft wird
//
// Der Hauptweg sucht nach Namen wie die Sammlung (`Drucksuche`, AP-10), der
// Fokus steht beim Öffnen im Suchfeld, und **die Kennung steht nur noch im
// Kasten** „Kennung von Hand". Ohne gewählten Druck legt der Knopf nichts an.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { kennungsfeld } from '../seiten/tauschprobe'
import { belegeApi, raeumeApi } from '../test/api'

import type { Listendeutungsantwort } from '../antwortenlisteneinfuhr'
import type { Handelsposition } from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln
const suche = kataloge.de.seiten.sammlung.drucksuche

afterEach(() => {
  raeumeApi()
})

const sturm = '00000000-0000-4000-8000-0000000000b1'

function eineDeutung(): Listendeutungsantwort {
  return {
    deutung: {
      zeilen: [
        {
          roh: '1 Sturmrufer',
          zeilennummer: 1,
          menge: 1,
          name: 'Sturmrufer',
          setcode: '',
          nummer: '',
          finish: 'nonfoil',
          treffer: [
            {
              printing: sturm,
              name: 'Sturmrufer',
              setcode: 'zwe',
              setname: 'Zweite',
              nummer: '3',
              seltenheit: 'rare',
              bild: '',
            },
          ],
          drucke: 1,
          eindeutig: true,
          befund: '',
          weitere: false,
          fehler: null,
        },
      ],
      befunde: [],
      alterSekunden: 0,
    },
    fehler: null,
  }
}

function Buehne({
  gesetzt,
}: {
  gesetzt: (positionen: Handelsposition[]) => void
}) {
  const [positionen, setzePositionen] = useState<Handelsposition[]>([])

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Handelspositionen
          ich="ich-1"
          gegenueber="du-1"
          positionen={positionen}
          bearbeitbar
          setzePositionen={(neu) => {
            gesetzt(neu)
            setzePositionen(neu)
          }}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

function zeige() {
  const deuten = vi.fn(() => Promise.resolve(eineDeutung()))
  belegeApi({
    ListeDeuten: deuten,
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  const gesetzt = vi.fn()
  render(<Buehne gesetzt={gesetzt} />)

  return { deuten, gesetzt }
}

describe('eine Zeile im Tausch anlegen', () => {
  it('sucht die Karte über ihren Namen, ohne Kennung im Hauptweg', async () => {
    const { deuten, gesetzt } = zeige()
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonGegenueber}`,
      }),
    )

    const seine = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })

    // **Der Fokus steht im Suchfeld**, und die Kennung steht nicht da.
    const feld = within(seine).getByRole('searchbox', { name: suche.karte })
    expect(feld).toHaveFocus()
    expect(within(seine).queryByLabelText(texte.positionPrinting)).toBeNull()

    // **Ohne Druck legt der Knopf nichts an.**
    const anlegen = within(seine).getByRole('button', {
      name: texte.positionHinzufuegen,
    })
    expect(anlegen).toBeDisabled()

    await hand.type(feld, 'Sturmrufer{Enter}')
    expect(deuten).toHaveBeenCalledWith('1 Sturmrufer')

    // **Ein einziger Treffer steht gewählt da** — aufgenommen wird erst mit
    // dem Knopf (§16, Regel 6).
    expect(
      await within(seine).findByRole('radio', { checked: true }),
    ).toBeInTheDocument()
    expect(gesetzt).not.toHaveBeenCalled()

    await hand.click(anlegen)

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ von: 'du-1', printing: sturm, menge: 1 }),
    ])
  })

  it('nimmt eine Kennung von Hand im Kasten an', async () => {
    const { gesetzt } = zeige()
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonIch}`,
      }),
    )
    await hand.click(screen.getByRole('button', { name: texte.kennungVonHand }))
    await hand.type(screen.getByLabelText(texte.positionPrinting), 'eigen-7')
    await hand.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ von: 'ich-1', printing: 'eigen-7' }),
    ])
  })

  it('nimmt den Druck wieder heraus, wenn eine neue Suche beginnt — außer er ist von Hand geändert', async () => {
    const { deuten } = zeige()
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', {
        name: `Zeile hinzufügen: ${texte.positionVonIch}`,
      }),
    )

    const feld = screen.getByRole('searchbox', { name: suche.karte })

    // **Ab der zweiten Suche ohne Treffer** — sonst wählte die neue Antwort
    // gleich wieder, und man sähe nicht, was die Suche aufhob.
    const ohneTreffer = eineDeutung()
    const [ersteZeile] = ohneTreffer.deutung?.zeilen ?? []
    if (ersteZeile !== undefined) {
      ersteZeile.treffer = []
      ersteZeile.drucke = 0
    }
    const anlegen = screen.getByRole('button', {
      name: texte.positionHinzufuegen,
    })

    await hand.type(feld, 'Sturmrufer{Enter}')
    await screen.findByRole('radio', { checked: true })
    expect(anlegen).toBeEnabled()

    // **Eine neue Suche hebt die Wahl auf** — der Knopf ruht wieder.
    deuten.mockResolvedValue(ohneTreffer)
    await hand.type(feld, '{Enter}')
    expect(await screen.findByText(suche.keinTreffer)).toBeVisible()
    expect(anlegen).toBeDisabled()

    // **Was von Hand geändert ist, bleibt.**
    await hand.clear(await kennungsfeld(hand))
    await hand.type(screen.getByLabelText(texte.positionPrinting), 'eigen-8')
    await hand.type(feld, '{Enter}')
    expect(screen.getByLabelText(texte.positionPrinting)).toHaveValue('eigen-8')
    expect(anlegen).toBeEnabled()
  })
})
