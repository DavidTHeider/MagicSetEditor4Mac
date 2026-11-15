// **Die Tischleiste bricht in Gruppen um** — P04
// (#2167).
//
// Die vier Blickhandlungen und die zwei Zielhandlungen standen mit dem Strich
// dazwischen in einer einzigen umbrechenden Reihe; bei schmalem Fenster stand
// „Auf meinen Sitz" allein in der nächsten Zeile. Jetzt bilden beide je eine
// benannte Gruppe, die als Ganzes umbricht.
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'

import type { Tischmalerei } from '../tisch/maler'
import { tischplan } from '../tisch/tischplan'
import { einMaler } from '../test/malerprobe'

const eigene = kataloge.de.seiten.spielen

it('fasst Blick und Ziel je in eine benannte Gruppe', async () => {
  const maler = einMaler({ breite: 800, hoehe: 600 })

  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={tischplan(einePartie({ objekte: [] }))}
        bild={() => undefined}
        malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
      />
    </Sprachraum>,
  )

  const blick = await screen.findByRole('group', { name: eigene.tischBlick })
  expect(
    within(blick)
      .getAllByRole('button')
      .map((knopf: HTMLElement) => knopf.textContent),
  ).toEqual([
    eigene.tischEinpassen,
    eigene.tischNaeher,
    eigene.tischWeiter,
    eigene.tischDrehen,
  ])

  const ziel = screen.getByRole('group', { name: eigene.tischZiel })
  expect(
    within(ziel)
      .getAllByRole('button')
      .map((knopf: HTMLElement) => knopf.textContent),
  ).toEqual([eigene.tischAuswahl, eigene.tischEigenerSitz])
})
