// **Der Markttisch misst an seinem eigenen Platz, nicht am Fenster** — QA01
// (#2382),
// gefunden in der WebKit-Abnahme `ux14-groessen`: bei 1024 Punkten
// Fensterbreite galt `lg:`, und drei Spalten standen in 343 Punkten neben der
// Liste der Handel. Die Mitte schrumpft nicht unter ihren Inhalt, und die
// Seite rollte waagerecht — „Zeile hinzufügen“ stand halb außerhalb.
//
// **jsdom misst nichts**, deshalb hält dieser Test die Klassen; gemessen wird
// am gerenderten Element in `ux14-groessen`.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Markttisch } from './Markttisch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einAbgleich } from '../test/tauschabgleich'

const tisch = kataloge.de.seiten.handeln.markttisch

afterEach(() => {
  raeumeApi()
})

it('legt die drei Spalten erst nebeneinander, wenn sein eigener Platz reicht', async () => {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
    HandelSichtAnfragen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Tauschabgleich: () => Promise.resolve(einAbgleich([])),
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  render(
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Markttisch
          ich="ich-1"
          gegenueber="du-1"
          name="Bert"
          positionen={[]}
          setzePositionen={() => {
            /* hier wird nichts gezogen */
          }}
        />
      </Anschlussraum>
    </Sprachraum>,
  )

  const gitter = (await screen.findByRole('region', { name: tisch.meinFach }))
    .parentElement

  expect(gitter).toHaveClass(
    '@2xl:grid-cols-[minmax(0,1fr)_minmax(0,2fr)_minmax(0,1fr)]',
  )
  expect(gitter?.className).not.toMatch(/(^|\s)lg:/)
  expect(gitter?.parentElement).toHaveClass('@container')
})
