// **Ein verborgener Tisch verpasst nichts** — M9 (b) aus dem Zweitblick über
// Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// Die Seite *Spielen* steht in einem `<Activity>`; verborgen baut React ihre
// Effekte ab, und damit die Horcher. Was in der Zeit kam — ein Tischbild,
// ein Zug, ein Satz im Gespräch —, war fort. Gemessen wird am echten
// `<Activity>` und an der Laufzeitprobe, über die auch die Wails-Ereignisse
// kommen.
import { act, render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { Activity } from 'react'
import { afterEach, expect, it } from 'vitest'

import { ereignisTischbild, ereignisZug, usePartiegast } from './usePartiegast'
import { Nachholgrenze } from '../nachholen'
import { ereignisNebenbei, usePartienebenbei } from './usePartienebenbei'
import { Sprachraum } from '../sprache/kontext'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { belegeApi, raeumeApi } from '../test/api'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function Tisch() {
  const gast = usePartiegast()
  const gespraech = usePartienebenbei('sitz-2', 'probe-gast', 'p1')

  return (
    <ul>
      <li>
        <button
          type="button"
          onClick={() => {
            gast.beitreten('ab:cd', 'p1')
          }}
        >
          Beitreten
        </button>
      </li>
      <li>{`Sitzt: ${String(gast.stand?.laeuft ?? false)}`}</li>
      <li>{`Partie: ${gast.tisch?.partie.id ?? '—'}`}</li>
      <li>{`Lücke: ${String(gast.luecke)}`}</li>
      <li>{`Züge: ${String(gast.zuege.length)}`}</li>
      <li>{`Sätze: ${gespraech.gespraech.map((zeile) => zeile.text).join(', ')}`}</li>
    </ul>
  )
}

function zeige(sichtbar: boolean) {
  return (
    <Sprachraum sprache="de">
      <Activity mode={sichtbar ? 'visible' : 'hidden'}>
        <Tisch />
      </Activity>
    </Sprachraum>
  )
}

it('holt Tischbild, Züge und Gespräch nach, die kamen, während er verborgen war', async () => {
  probe = laufzeitprobe()
  const laufzeit = probe

  belegeApi({
    PartieBeitreten: () =>
      Promise.resolve({
        laeuft: true,
        partie: 'p1',
        gegenueber: 'ab:cd',
        pausiert: false,
        absage: '',
        fehler: null,
      }),
  })

  const { rerender } = render(zeige(true))

  // **Erst am Tisch** (Z1, Zweitblick über 6.11): ohne Platz nimmt der Gast
  // nichts.
  await nutzerereignis
    .setup()
    .click(screen.getByRole('button', { name: 'Beitreten' }))
  expect(await screen.findByText('Sitzt: true')).toBeInTheDocument()

  act(() => {
    laufzeit.melde(ereignisZug, 'p1', { art: 'gezogen' })
  })
  expect(await screen.findByText('Züge: 1')).toBeInTheDocument()

  rerender(zeige(false))

  act(() => {
    laufzeit.melde(ereignisTischbild, 'p1', {
      partie: { id: 'p1' },
      erlaubt: [],
    })
    laufzeit.melde(ereignisZug, 'p1', { art: 'gespielt' })
    laufzeit.melde(ereignisNebenbei, 'p1', {
      art: 'chat',
      von: 'sitz-1',
      text: 'Hallo',
    })
  })

  rerender(zeige(true))

  expect(await screen.findByText('Partie: p1')).toBeInTheDocument()
  expect(screen.getByText('Züge: 2')).toBeInTheDocument()
  expect(screen.getByText('Sätze: Hallo')).toBeInTheDocument()
})

// **Und was über der Grenze wegfiel, wird genannt** — Z3 aus dem Zweitblick
// über Phase 6.11 (#1672).
it('meldet eine Lücke, wenn verborgen mehr kam, als nachgeholt wird', async () => {
  probe = laufzeitprobe()
  const laufzeit = probe
  belegeApi({
    PartieBeitreten: () =>
      Promise.resolve({
        laeuft: true,
        partie: 'p1',
        gegenueber: 'ab:cd',
        pausiert: false,
        absage: '',
        fehler: null,
      }),
  })

  const { rerender } = render(zeige(true))

  await nutzerereignis
    .setup()
    .click(screen.getByRole('button', { name: 'Beitreten' }))
  expect(await screen.findByText('Sitzt: true')).toBeInTheDocument()
  expect(screen.getByText('Lücke: false')).toBeInTheDocument()

  rerender(zeige(false))

  act(() => {
    for (let nr = 0; nr < Nachholgrenze + 1; nr++) {
      laufzeit.melde(ereignisZug, 'p1', { art: 'gezogen' })
    }
  })

  rerender(zeige(true))

  expect(await screen.findByText('Lücke: true')).toBeInTheDocument()
  expect(screen.getByText(`Züge: ${String(Nachholgrenze)}`)).toBeInTheDocument()
})
