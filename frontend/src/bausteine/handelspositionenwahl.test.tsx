// **Güte, Ausführung und Sprache einer Handelsposition** — UX-06.AK-01
// (#1678).
//
// # Der Befund
//
// **Die Güte stand als Freitext mit der Vorgabe `NM` da.** Verbucht wird
// eine Position aber gegen die sieben Zustände der Sammlung
// (`service.Handelsverbuchung` prüft `Journalzeile.AlsExemplar().Validate()`,
// und dort gilt `domain.IstZustand`) — `NM` ist keiner. Eine Position, die
// jemand im Fenster mit der Vorgabe anlegte, scheiterte also erst beim
// Verbuchen, am Ende eines Handels. Was aus der Sammlung mitreist, trug schon
// immer das richtige Wort (`lightPlayed`); nur die leere Zeile nicht.
//
// **Die Ausführung ebenso**: drei Werte, und ein Tippfehler darin wäre erst
// beim Verbuchen aufgefallen. **Die Sprache bleibt frei**, bekommt aber die
// Codes vorgeschlagen, unter denen Scryfall führt — dieselbe Entscheidung wie
// in der Sammlung.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState, type ReactElement } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'

import type { Handelsposition } from '../apihandelsleitung'
import { kennungsfeld } from '../seiten/tauschprobe'
import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.handeln
const sammlung = kataloge.de.seiten.sammlung

function Buehne({
  setzePositionen,
}: {
  setzePositionen: (positionen: Handelsposition[]) => void
}) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('handeln')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        <Handelspositionen
          ich="ich-1"
          gegenueber="du-1"
          positionen={[]}
          setzePositionen={setzePositionen}
          bearbeitbar
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

// oeffne zeigt die Bühne und öffnet das Formular in der eigenen Spalte —
// seit TRADE-02b (#1951) steht es erst auf Wunsch da.
async function oeffne(buehne: ReactElement) {
  render(buehne)
  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(texte.assistent.zeileIn, { seite: texte.positionVonIch }),
    }),
  )
}

describe('eine Handelsposition', () => {
  it('wählt die Güte aus den sieben Zuständen der Sammlung', async () => {
    await oeffne(<Buehne setzePositionen={vi.fn()} />)

    const guete = screen.getByLabelText(texte.positionGuete)

    expect(guete.tagName).toBe('SELECT')
    expect(
      within(guete)
        .getAllByRole('option')
        .map((eine) => eine.getAttribute('value')),
    ).toEqual([
      'mint',
      'nearMint',
      'excellent',
      'good',
      'lightPlayed',
      'played',
      'poor',
    ])
    expect(
      within(guete).getByRole('option', { name: sammlung.zustaende.nearMint }),
    ).toBeInTheDocument()
  })

  it('wählt die Ausführung aus dreien', async () => {
    await oeffne(<Buehne setzePositionen={vi.fn()} />)

    const ausfuehrung = screen.getByLabelText(texte.positionFinish)

    expect(ausfuehrung.tagName).toBe('SELECT')
    expect(
      within(ausfuehrung)
        .getAllByRole('option')
        .map((eine) => eine.getAttribute('value')),
    ).toEqual(['nonfoil', 'foil', 'etched'])
  })

  it('schlägt Sprachen vor und nimmt trotzdem jeden Code', async () => {
    await oeffne(<Buehne setzePositionen={vi.fn()} />)

    const sprache = screen.getByLabelText(texte.positionSprache)

    expect(sprache.getAttribute('list')).not.toBeNull()

    await userEvent.clear(sprache)
    await userEvent.type(sprache, 'tlh')

    expect(sprache).toHaveValue('tlh')
  })

  it('legt mit der Vorgabe eine Position an, die sich verbuchen lässt', async () => {
    const setzePositionen = vi.fn()

    await oeffne(<Buehne setzePositionen={setzePositionen} />)

    await userEvent.type(await kennungsfeld(userEvent), 'p-1')
    await userEvent.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(setzePositionen).toHaveBeenCalledWith([
      expect.objectContaining({
        printing: 'p-1',
        menge: 1,
        guete: 'nearMint',
        sprache: 'de',
        finish: 'nonfoil',
      }),
    ])
  })

  it('trägt die gewählte Güte und Ausführung in die Position', async () => {
    const setzePositionen = vi.fn()

    await oeffne(<Buehne setzePositionen={setzePositionen} />)

    await userEvent.type(await kennungsfeld(userEvent), 'p-2')
    await userEvent.selectOptions(
      screen.getByLabelText(texte.positionGuete),
      'played',
    )
    await userEvent.selectOptions(
      screen.getByLabelText(texte.positionFinish),
      'foil',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.positionHinzufuegen }),
    )

    expect(setzePositionen).toHaveBeenCalledWith([
      expect.objectContaining({ guete: 'played', finish: 'foil' }),
    ])
  })
})
