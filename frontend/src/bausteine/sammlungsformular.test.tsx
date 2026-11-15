// **Das Formular und die Zeile der Sammlung** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum jedes Feld einen Lauf hat
//
// Ein Formular mit acht Feldern hat acht Stellen, an denen ein Wert im
// falschen Feld landet — und keine davon fällt beim Zeichnen auf. Was hier
// geprüft wird, ist, dass jedes getippte Zeichen dort ankommt, wo es hingehört.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sammlungsformular } from './Sammlungsformular'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.sammlung

// stillePreise belegt die Bindungen, die die Zeile nebenbei ruft.

describe('Das Sammlungsformular', () => {
  afterEach(raeumeApi)

  it('reicht jedes Feld an das Exemplar durch', async () => {
    const aufgenommen = vi.fn(() => Promise.resolve(true))

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
      </Sprachraum>,
    )

    await userEvent.type(await druckfeld(), 'abc')
    await userEvent.clear(screen.getByLabelText(texte.menge))
    await userEvent.type(screen.getByLabelText(texte.menge), '4')
    await userEvent.selectOptions(
      screen.getByLabelText(texte.zustand),
      'played',
    )
    await userEvent.type(screen.getByLabelText(texte.sprache), 'de')
    await userEvent.selectOptions(
      screen.getByLabelText(texte.ausfuehrung),
      'foil',
    )
    await userEvent.type(screen.getByLabelText(texte.erwerbspreis), '4.17')
    await userEvent.type(screen.getByLabelText(texte.erwerbswaehrung), 'EUR')
    await userEvent.type(
      screen.getByLabelText(texte.erwerbsdatum),
      '2026-08-01',
    )
    await userEvent.type(screen.getByLabelText(texte.notiz), 'gekauft')

    // **Der lesbare Bezug** (#1200) — wahlfrei, aber wer ihn tippt, soll ihn
    // wiederfinden.
    await userEvent.type(screen.getByLabelText(texte.name), 'Lightning Bolt')
    await userEvent.type(screen.getByLabelText(texte.setcode), 'm11')
    await userEvent.type(screen.getByLabelText(texte.nummer), '149')

    // **Die Ordnung** (#1292, Vermerk 0057) — Lagerort als Freitext, und die
    // Absicht als Schalter.
    await userEvent.type(screen.getByLabelText(texte.lagerort), 'Kiste 4')
    await userEvent.click(screen.getByLabelText(texte.wunsch))

    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(aufgenommen).toHaveBeenCalledWith({
      id: 0,
      fassung: 0,
      printing: 'abc',
      druckID: '',
      vorlaeufig: false,
      menge: 4,
      zustand: 'played',
      sprache: 'de',
      finish: 'foil',
      erwerbsdatum: '2026-08-01',
      erwerbspreis: '4.17',
      erwerbswaehrung: 'EUR',
      notiz: 'gekauft',
      lagerort: 'Kiste 4',
      wunsch: true,
      name: 'Lightning Bolt',
      setcode: 'm11',
      nummer: '149',

      // **Die Bezeichnung bleibt leer** (#1200): sie ist gerechnet und kommt
      // aus Go — ein Wert von hier wäre eine zweite Wahrheit (§16).
      bezeichnung: '',
      herkunft: keineHerkunft,
    })
  })

  // **Ohne Druck geht der Knopf nicht.** Ein Exemplar ohne Druck ist keines,
  // und ein Absenden, das sicher scheitert, ist ein Weg ins Leere.
  it('lässt den Knopf ohne Druck ruhen', () => {
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular
          aufAufnehmen={() => Promise.resolve(true)}
          laeuft={false}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('button', { name: texte.aufnehmen })).toBeDisabled()
  })

  // **Geräumt wird erst nach dem Zug** (§7, Regel 3).
  //
  // Wer bei einem Fehler ein leeres Formular vorfände, hätte seine Eingabe
  // verloren.
  it('räumt das Formular erst nach dem Aufnehmen', async () => {
    let weiter: (() => void) | undefined

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular
          aufAufnehmen={() =>
            new Promise<boolean>((los) => {
              weiter = () => {
                los(true)
              }
            })
          }
          laeuft={false}
        />
      </Sprachraum>,
    )

    const druck = await druckfeld()

    await userEvent.type(druck, 'abc')
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(druck).toHaveValue('abc')

    weiter?.()

    await waitFor(() => {
      expect(druck).toHaveValue('')
    })
  })
})
