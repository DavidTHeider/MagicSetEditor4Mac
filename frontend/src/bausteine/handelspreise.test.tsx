// **Preise holen, wo getauscht wird** — P10
// (#2378),
// Befund C-17: ohne geholte Preise stand keine Bilanz da, nur der Verweis
// auf die Sammlung.
//
// # Wogegen hier geprüft wird
//
// **Derselbe Weg wie in der Sammlung** (`PreiseHolen`, je Druck einmal,
// gedrosselt im Klienten), ein eigener Druck ohne Scryfall-Kennung wird nicht
// gefragt. **Ein Fehler steht am Knopf und nimmt nichts weg**, ein Abbruch
// hält vor dem nächsten Druck an, und danach liest die Bilanz nach.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { Preisantwort } from '../api'
import type { Handelsposition, Preisanhalt } from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln
const preise = texte.preise

afterEach(() => {
  raeumeApi()
})

const blitz = '00000000-0000-4000-8000-0000000000a1'
const sturm = '00000000-0000-4000-8000-0000000000b1'

function zeile(von: string, printing: string, druckId = 0): Handelsposition {
  return {
    von,
    printing,
    druckId,
    menge: 1,
    guete: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    abbild: '',
    bewertung: 0,
  }
}

const positionen = [
  zeile('ich-1', blitz),
  zeile('du-1', sturm),
  zeile('ich-1', blitz),
  zeile('ich-1', '', 3),
]

const leererAnhalt: Preisanhalt = {
  zeilen: [],
  seiten: [],
  differenzen: [],
  hauptwaehrung: '',
}

const geholt: Preisantwort = { staende: [], befunde: [], fehler: null }

function zeige(PreiseHolen: (printing: string) => Promise<Preisantwort>) {
  const anhalt = vi.fn(() =>
    Promise.resolve({ anhalt: leererAnhalt, fehler: null }),
  )
  const holen = vi.fn(PreiseHolen)
  const gesetzt = vi.fn()
  belegeApi({ HandelPreisanhalt: anhalt, PreiseHolen: holen })

  render(
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
          setzePositionen={gesetzt}
        />
      </Anschlussraum>
    </Sprachraum>,
  )

  return { anhalt, holen, gesetzt }
}

describe('Preise holen am Tisch', () => {
  it('fragt je Druck einmal, übergeht den eigenen und liest die Bilanz nach', async () => {
    const { anhalt, holen, gesetzt } = zeige(() => Promise.resolve(geholt))
    const hand = userEvent.setup()

    // **Die Bilanz sagt, was fehlt, und schickt nirgends hin.**
    expect(await screen.findByText(texte.assistent.keinePreise)).toBeVisible()
    await vi.waitFor(() => {
      expect(anhalt).toHaveBeenCalledTimes(1)
    })

    await hand.click(screen.getByRole('button', { name: preise.holen }))

    expect(
      await screen.findByText(preise.fertig.replace('{zahl}', '2')),
    ).toBeVisible()
    expect(holen.mock.calls).toEqual([[blitz], [sturm]])
    await vi.waitFor(() => {
      expect(anhalt).toHaveBeenCalledTimes(2)
    })
    expect(gesetzt).not.toHaveBeenCalled()
  })

  it('nennt einen Fehler am Knopf und holt die übrigen trotzdem', async () => {
    const fehler = (meldung: string): Preisantwort => ({
      staende: [],
      befunde: [],
      fehler: { code: 'unbekannt', meldung, schluessel: '', feld: '' },
    })
    const { anhalt, holen, gesetzt } = zeige((printing) =>
      Promise.resolve(
        fehler(
          printing === blitz
            ? 'Scryfall drosselt gerade.'
            : 'Keine Verbindung zu Scryfall.',
        ),
      ),
    )
    const hand = userEvent.setup()

    await hand.click(await screen.findByRole('button', { name: preise.holen }))

    // **Der erste Grund steht da**, und gefragt wurde trotzdem jeder Druck.
    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(
      preise.teils
        .replace('{zahl}', '2')
        .replace('{gesamt}', '2')
        .replace('{meldung}', 'Scryfall drosselt gerade.'),
    )
    expect(holen).toHaveBeenCalledTimes(2)
    await vi.waitFor(() => {
      expect(anhalt).toHaveBeenCalledTimes(2)
    })

    // **Die Verhandlung bleibt**: keine Zeile geht verloren.
    expect(gesetzt).not.toHaveBeenCalled()
    expect(screen.getAllByRole('listitem')).toHaveLength(positionen.length)
  })

  it('fragt nach dem Verlassen des Tauschs keinen weiteren Druck', async () => {
    let erster: ((antwort: Preisantwort) => void) | undefined
    const { holen, anhalt } = zeige(
      () =>
        new Promise((aufloesen) => {
          erster = aufloesen
        }),
    )
    const hand = userEvent.setup()

    await hand.click(await screen.findByRole('button', { name: preise.holen }))
    await vi.waitFor(() => {
      expect(anhalt).toHaveBeenCalledTimes(1)
    })

    // **Gegenprobe, dass eine Antwort dort ankäme**: der Lauf steht.
    expect(holen).toHaveBeenCalledTimes(1)

    cleanup()
    erster?.(geholt)
    await new Promise((weiter) => setTimeout(weiter, 0))

    expect(holen).toHaveBeenCalledTimes(1)
    expect(anhalt).toHaveBeenCalledTimes(1)
  })

  it('hält nach einem Abbruch vor dem nächsten Druck an', async () => {
    let erster: ((antwort: Preisantwort) => void) | undefined
    const { holen } = zeige(
      () =>
        new Promise((aufloesen) => {
          erster = aufloesen
        }),
    )
    const hand = userEvent.setup()

    await hand.click(await screen.findByRole('button', { name: preise.holen }))

    // **Ein Lauf zur Zeit** (§7, Regel 4): der Knopf ruht, solange er läuft.
    expect(screen.getByRole('button', { name: preise.holen })).toBeDisabled()
    expect(
      screen.getByText(
        preise.laeuft.replace('{fertig}', '0').replace('{gesamt}', '2'),
      ),
    ).toBeVisible()

    await hand.click(screen.getByRole('button', { name: preise.abbrechen }))
    erster?.(geholt)

    expect(
      await screen.findByText(
        preise.abgebrochen.replace('{fertig}', '1').replace('{gesamt}', '2'),
      ),
    ).toBeVisible()
    expect(holen).toHaveBeenCalledTimes(1)
    expect(screen.getByRole('button', { name: preise.holen })).toBeEnabled()
  })
})
