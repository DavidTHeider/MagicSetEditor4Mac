// **Am Feld steht, wen der Vorrang fragt** — #2054, aus AI-02 (#1743): ohne
// Wahl nimmt der Vorrang bei Ollama das zuletzt gezogene Modell, und das soll
// am Formular stehen, bevor jemand fragt.
//
// **Geladen erst, wenn jemand die Zeile berührt**: der Katalog fragt auch
// eingeschaltete Cloudanbieter nach ihren Modellen, und das soll nicht bei
// jeder Karte geschehen. **Und einmal für beide Felder**.
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIFeldvorschlag } from './KIFeldvorschlag'
import { vergissVorrangreihe } from '../haken/useVorrangreihe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { KIKatalogantwort } from '../antwortenmarkt'

afterEach(() => {
  raeumeApi()
  vergissVorrangreihe()
})

const eigene = kataloge.de.seiten.einstellungen

function katalog(
  vorrang: string[],
  vorrangbefunde: string[] = [],
): KIKatalogantwort {
  return {
    nurLokal: false,
    einrichtung: 'bereit',
    anbieter: [],
    cloud: { stand: 'ungewiss', beleg: '' },
    befunde: [],
    modelle: [],
    vorrang,
    vorrangbefunde,
    fehler: null,
  }
}

function zeige(feld: string) {
  render(
    <Sprachraum sprache="de">
      <KIFeldvorschlag
        karteID="4"
        feld={feld}
        aufschrift={feld}
        werte={[]}
        formen={new Map()}
        geaendert={vi.fn()}
      />
    </Sprachraum>,
  )
}

const knopf = (feld: string) =>
  screen.getByRole('button', {
    name: einsetzen(eigene.kiFeldVorschlagenFuer, { feld }),
  })

it('nennt die Reihe des Vorrangs, sobald jemand die Zeile berührt — einmal geladen', async () => {
  const gefragt = vi.fn(() =>
    Promise.resolve(katalog(['apple/system', 'ollama/gemma3:1b'])),
  )
  belegeApi({ KIModelle: gefragt })
  zeige('name')
  zeige('flavor text')

  expect(gefragt).not.toHaveBeenCalled()

  await userEvent.hover(knopf('name'))
  const satz = einsetzen(eigene.kiFeldVorrangFragt, {
    reihe: 'apple/system, ollama/gemma3:1b',
  })

  expect(await screen.findByText(satz)).toBeInTheDocument()
  expect(knopf('name')).toHaveAccessibleDescription(satz)

  await userEvent.hover(knopf('flavor text'))
  expect(await screen.findAllByText(satz)).toHaveLength(2)
  expect(gefragt).toHaveBeenCalledOnce()
})

it('zeigt nichts, wenn der Vorrang niemanden fragt', async () => {
  const gefragt = vi.fn(() => Promise.resolve(katalog([])))
  belegeApi({ KIModelle: gefragt })
  zeige('name')

  await userEvent.hover(knopf('name'))

  // **Erst die Antwort, dann die Verneinung** (§3): gewartet wird auf das,
  // was die Laufzeit zurückgab, und auf die Effekte danach — sonst träfe
  // „keine Beschreibung" schon vor der Antwort zu.
  await vi.waitFor(() => {
    expect(gefragt).toHaveBeenCalledOnce()
  })
  await gefragt.mock.results[0]?.value
  await effekteNachholen()

  expect(knopf('name')).not.toHaveAccessibleDescription()
})

// **Eine Wahl, die nicht greift, steht am Feld** — Z4 aus dem Zweitblick über
// Phase 6.11 (#1672): Kommandozeile, MCP und REST nannten den Befund, das
// Formular schwieg.
it('nennt den Befund, wenn das gewählte Modell fehlt', async () => {
  const befund =
    'Das gewählte Vorrangmodell „mistral:7b“ nennt die Laufzeit nicht.'
  belegeApi({
    KIModelle: () => Promise.resolve(katalog(['apple/system'], [befund])),
  })
  zeige('name')

  await userEvent.hover(knopf('name'))

  expect(await screen.findByText(befund, { exact: false })).toBeInTheDocument()
  expect(knopf('name')).toHaveAccessibleDescription(
    `${einsetzen(eigene.kiFeldVorrangFragt, { reihe: 'apple/system' })} ${befund}`,
  )
})

// **Ein stehendes Feld folgt der Einstellung** — Z5: der Speicher wurde
// geleert, aber ein Feld, das die Reihe schon zeigte, fragte nicht neu.
it('frischt ein stehendes Feld auf, wenn sich die Wahl ändert', async () => {
  let reihe = ['ollama/gemma3:1b']
  belegeApi({ KIModelle: () => Promise.resolve(katalog(reihe)) })
  zeige('name')

  await userEvent.hover(knopf('name'))
  expect(
    await screen.findByText(
      einsetzen(eigene.kiFeldVorrangFragt, { reihe: 'ollama/gemma3:1b' }),
    ),
  ).toBeInTheDocument()

  reihe = ['ollama/qwen3:0.6b']
  act(() => {
    vergissVorrangreihe()
  })

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFeldVorrangFragt, { reihe: 'ollama/qwen3:0.6b' }),
    ),
  ).toBeInTheDocument()
})

// **Und nach der Frist beim nächsten Berühren** — Z5: die Minute lief ab,
// und das Feld fragte trotzdem nie wieder.
it('fragt nach der Frist beim nächsten Berühren neu', async () => {
  const jetzt = vi.spyOn(Date, 'now').mockReturnValue(1_000_000)
  const gefragt = vi.fn(() => Promise.resolve(katalog(['apple/system'])))
  belegeApi({ KIModelle: gefragt })
  zeige('name')

  await userEvent.hover(knopf('name'))
  await vi.waitFor(() => {
    expect(gefragt).toHaveBeenCalledOnce()
  })

  await userEvent.unhover(knopf('name'))
  await userEvent.hover(knopf('name'))
  expect(gefragt).toHaveBeenCalledOnce()

  jetzt.mockReturnValue(1_000_000 + 61_000)
  await userEvent.unhover(knopf('name'))
  await userEvent.hover(knopf('name'))

  await vi.waitFor(() => {
    expect(gefragt).toHaveBeenCalledTimes(2)
  })
  jetzt.mockRestore()
})
