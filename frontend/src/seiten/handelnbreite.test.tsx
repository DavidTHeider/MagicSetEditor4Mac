// **Handeln teilt sich nach seinem Platz und nicht nach dem Fenster** — QA01
// (#2382),
// gefunden in der WebKit-Abnahme `ux14-groessen`: bei sehr großer Schrift
// galt bei 1024 Punkten `lg:`, die Liste nahm ihre 20rem — 400 Punkte —, und
// dem offenen Handel blieben 173. Die Fächer des Markttischs liefen über, und
// die Seite rollte waagerecht.
//
// **jsdom misst nichts**, deshalb hält dieser Test die Klassen; gemessen wird
// am gerenderten Element in `ux14-groessen`.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { handelsuebersicht, zehnGegenueber } from '../test/handelsbestand'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

it('stellt Liste und Handel erst nebeneinander, wenn ihr Platz reicht', async () => {
  belegeApi({
    Peerkontakte: () => Promise.resolve(zehnGegenueber()),
    HandelUebersicht: handelsuebersicht,
  })
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  const knopf = await screen.findByRole('button', { name: eigene.neuerTausch })
  const spalte = knopf.parentElement?.parentElement
  const reihe = spalte?.parentElement

  expect(spalte).toHaveClass('@2xl:w-80')
  expect(reihe).toHaveClass('@2xl:flex-row')
  expect(
    `${String(spalte?.className)} ${String(reihe?.className)}`,
  ).not.toMatch(/(^|\s)lg:/)
  expect(reihe?.parentElement).toHaveClass('@container')
})
