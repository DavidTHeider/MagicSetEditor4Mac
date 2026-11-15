// **Bereiche und Funktionen auf der Einstellungsseite** — OPS-10
// (#1908).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Bereichsschalter } from './Bereichsschalter'
import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

function schalter(kennung: string, wert: string): EinstellungDTO {
  return {
    schluessel: `schalter.${kennung}`,
    wert,
    wirksam: wert,
    werte: ['an', 'aus'],
    bereich: 'schalter',
    ebene: 'geraet',
    herkunft: 'vorgabe',
    vorgabe: 'an',
    wirkung: 'sofort',
  }
}

function zeige(reife: Record<string, string>, setze = vi.fn()) {
  belegeApi({
    Schalter: () =>
      Promise.resolve({
        schalter: Object.entries(reife).map(([kennung, stufe]) => ({
          kennung,
          art: 'funktion',
          reife: stufe,
          vorgabe: 'an',
          an: true,
        })),
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Bereichsschalter
        eintraege={[
          schalter('spielen', 'an'),
          // **Abgeschaltet heißt gesetzt** — und gesetzt trägt eine Marke
          // (AP-11, #2368).
          { ...schalter('ki', 'aus'), herkunft: 'geraet' },
          // **Ein Schalter, den dieses Fenster nicht kennt**, steht mit
          // seinem Schlüssel da und ohne Satz.
          schalter('neu', 'an'),
          { ...schalter('x', 'an'), schluessel: 'sprache' },
        ]}
        angehalten={() => false}
        setze={setze}
      />
    </Sprachraum>,
  )

  return setze
}

describe('die Schalter der Bereiche und Funktionen', () => {
  // **Ein Schalter und keine Auswahl** — UXR-19
  // (#1927),
  // Regel R7 aus `docs/UIREVIEW.md`: was sofort wirkt und zwei Werte hat, ist
  // ein Schalter. Hier standen zehn Auswahllisten mit „An" und „Aus".
  it('zeigt je Schalter Namen, Stand und Satz — und sonst nichts', async () => {
    zeige({ spielen: 'stabil', ki: 'stabil' })

    expect(
      await screen.findByRole('switch', {
        name: eigene.namen['schalter.spielen'],
      }),
    ).toBeChecked()
    expect(
      screen.getByRole('switch', { name: eigene.namen['schalter.ki'] }),
    ).not.toBeChecked()
    expect(screen.getByText(eigene.schalter.saetze.spielen)).toBeVisible()
    expect(screen.getByRole('switch', { name: 'schalter.neu' })).toBeChecked()
    expect(screen.queryByLabelText(eigene.namen.sprache)).toBeNull()
    expect(screen.queryByRole('combobox')).toBeNull()
  })

  it('schreibt, was umgelegt wird — aus und wieder an', async () => {
    const setze = zeige({})

    await userEvent.click(
      screen.getByRole('switch', { name: eigene.namen['schalter.spielen'] }),
    )
    expect(setze).toHaveBeenCalledWith('schalter.spielen', 'aus')

    await userEvent.click(
      screen.getByRole('switch', { name: eigene.namen['schalter.ki'] }),
    )
    expect(setze).toHaveBeenCalledWith('schalter.ki', 'an')
  })

  // **Mit der Tastatur wie jeder Schalter** — die Leertaste legt um (§6).
  it('legt mit der Leertaste um und sagt, woher der Wert kommt', async () => {
    const setze = zeige({})

    screen
      .getByRole('switch', { name: eigene.namen['schalter.spielen'] })
      .focus()
    await userEvent.keyboard(' ')

    expect(setze).toHaveBeenCalledWith('schalter.spielen', 'aus')
    // **Woher der Wert kommt, sagt die Marke** (AP-11, #2368) — am
    // gesetzten, nicht an denen der Vorgabe.
    expect(document.querySelectorAll('[data-herkunftsmarke]')).toHaveLength(1)
    expect(screen.getByText('von diesem Gerät')).toBeInTheDocument()
  })

  it('nimmt keinen zweiten Wunsch, solange geschrieben wird', () => {
    render(
      <Sprachraum sprache="de">
        <Bereichsschalter
          eintraege={[schalter('spielen', 'an')]}
          angehalten={() => true}
          setze={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('switch', { name: eigene.namen['schalter.spielen'] }),
    ).toBeDisabled()
  })

  it('zeigt eine Reife, die nicht stabil ist, als Marke', async () => {
    zeige({ spielen: 'stabil', ki: 'vorschau' })

    expect(
      await screen.findByText(eigene.schalter.reife.vorschau),
    ).toBeVisible()
    expect(screen.queryByText(eigene.schalter.reife.stabil)).toBeNull()
  })

  it('zeigt eine unbekannte Reife beim Namen', async () => {
    zeige({ ki: 'experimentell' })

    expect(await screen.findByText('experimentell')).toBeVisible()
  })
})
