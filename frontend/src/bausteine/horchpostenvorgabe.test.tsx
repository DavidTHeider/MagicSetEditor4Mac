// **Das Feld beginnt mit der Adresse, die der Hinweis darunter erklärt** —
// K02 (#2188).
//
// Gesehen am 2026-10-07 im Gast mit der gebauten Anwendung: `:41234` stand nur
// als Platzhalter da, sah aus wie ein Wert, und „Erreichbar werden" blieb ohne
// Erklärung angehalten. Eigene Datei, weil `horchposten.test.tsx` an der
// Zeilengrenze steht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Horchposten } from './Horchposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Horchantwort } from '../apipeerkontakt'

const texte = kataloge.de.seiten.technik

// einStand baut eine Antwort ohne Namen und Adressen.
function einStand(laeuft: boolean, adresse: string): Horchantwort {
  return {
    stand: {
      laeuft,
      adresse,
      sichtbar: false,
      erreichbarNamen: [],
      erreichbarAdressen: [],
    },
    fehler: null,
  }
}

describe('Die Adresse, unter der gehorcht wird', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('beginnt mit :41234, und der Knopf ist sofort bereit', async () => {
    const gesehen: string[] = []

    belegeApi({
      HorchStand: () => Promise.resolve(einStand(false, '')),
      HorchStarten: (adresse: string) => {
        gesehen.push(adresse)

        return Promise.resolve(einStand(true, adresse))
      },
    })

    render(
      <Sprachraum sprache="de">
        <Horchposten />
      </Sprachraum>,
    )
    const hand = userEvent.setup()

    expect(await screen.findByLabelText(texte.horchenAdresse)).toHaveValue(
      ':41234',
    )
    await hand.click(screen.getByRole('button', { name: texte.horchenStarten }))

    expect(gesehen).toEqual([':41234'])
  })
})
