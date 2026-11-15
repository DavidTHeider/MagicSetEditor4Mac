// **Das Kontextfenster je Modell** — AI-01.AK-02
// (#1742):
// „Kontext-/Ressourcenlimits … sind nachvollziehbar". Bis dahin stand das
// Fenster nur am Kontextpaket der Recherche, und dort angenommen. Jetzt
// steht es an jedem Modell — mit der Herkunft der Zahl, und wo niemand es
// sagt, steht das da.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { KIModelle } from './KIModelle'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { KIModellDTO } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

function modell(teil: Partial<KIModellDTO>): KIModellDTO {
  return {
    anbieter: 'ollama',
    alias: 'beispiel:3b',
    laufzeit: 'ollama 0.12.3',
    ort: 'lokal',
    faehigkeiten: [],
    herkunft: [],
    bindung: 'abcdef',
    kontextfenster: 0,
    kontextbeleg: '',
    ...teil,
  }
}

function zeige(modelle: KIModellDTO[]) {
  belegeApi({
    KIModelle: () =>
      Promise.resolve({
        nurLokal: true,
        vorrang: [],
        vorrangbefunde: [],
        einrichtung: 'bereit',
        anbieter: [],
        cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
        befunde: [],
        modelle,
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <KIModelle />
    </Sprachraum>,
  )
}

it('nennt das Fenster und woher die Zahl stammt', async () => {
  zeige([modell({ kontextfenster: 8192, kontextbeleg: 'Ollama, num_ctx' })])

  expect(
    await screen.findByText(
      einsetzen(eigene.kiKontextfenster, {
        tokens: '8.192',
        beleg: 'Ollama, num_ctx',
      }),
    ),
  ).toBeInTheDocument()
})

// **Wo niemand es sagt, steht das da** — und nicht „0 Tokens".
it('sagt, wenn kein Fenster angegeben ist', async () => {
  zeige([modell({})])

  expect(
    await screen.findByText(eigene.kiKontextfensterUnbekannt),
  ).toBeInTheDocument()
  expect(screen.queryByText(/0 Tokens/)).not.toBeInTheDocument()
})
