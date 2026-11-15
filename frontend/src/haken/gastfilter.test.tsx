// **Der Gast nimmt nur, was seinem Tisch gehört** — Z1 aus dem Zweitblick
// über Phase 6.11 (#1672).
//
// # Der Befund
//
// Go nennt bei jedem Tischbild und jedem Zug die Partie; der Haken warf die
// Kennung weg. Wer von Tisch A zu Tisch B wechselte, bekam verspätete und
// nachgeholte Züge von A in die Liste von B — und das Fenster des **Wirts**
// bekommt dieselben Ereignisse für seine eigene Partie (`MitFenster`, #1627),
// die dann als fremder Tisch im Gastplatz standen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { ereignisTischbild, ereignisZug, usePartiegast } from './usePartiegast'
import { einePartie } from '../seiten/partieprobe'

import type { ReactNode } from 'react'
import type { PartiebeitrittAntwort, Spielereignis } from '../apipartie'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function stand(teile: Partial<PartiebeitrittAntwort>): PartiebeitrittAntwort {
  return {
    laeuft: true,
    partie: '',
    gegenueber: 'ab:cd',
    pausiert: false,
    absage: '',
    fehler: null,
    ...teile,
  }
}

function bild(partie: string, fassung: string) {
  return { partie: einePartie({ id: partie, fassung }), erlaubt: [] }
}

function zug(folge: string): Spielereignis {
  return {
    folge,
    art: 'objekt.getappt',
    oeffentlich: {},
    privat: {},
    kette: '',
  }
}

function gast(antwort: PartiebeitrittAntwort) {
  probe = laufzeitprobe()
  belegeApi({
    PartieBeitreten: () => Promise.resolve(antwort),
    PartieBeitrittBeenden: () => Promise.resolve(stand({ laeuft: false })),
  })

  return {
    lauf: renderHook(() => usePartiegast(), { wrapper: raum }),
    melde: probe.melde,
  }
}

it('nimmt nichts, solange er an keinem Tisch sitzt', () => {
  const { lauf, melde } = gast(stand({ partie: 'p9' }))

  act(() => {
    melde(ereignisTischbild, 'p1', bild('p1', '3'))
    melde(ereignisZug, 'p1', zug('1'))
  })

  expect(lauf.result.current.tisch).toBeUndefined()
  expect(lauf.result.current.zuege).toEqual([])

  // **Und es hört jemand zu** (§3): wer beitritt, bekommt sein Bild.
  act(() => {
    lauf.result.current.beitreten('ab:cd', 'p9')
  })
  act(() => {
    melde(ereignisTischbild, 'p9', bild('p9', '1'))
  })

  expect(lauf.result.current.tisch?.partie.id).toBe('p9')
})

it('nimmt nur die Partie, an der er sitzt', async () => {
  const { lauf, melde } = gast(stand({ partie: 'pA' }))

  act(() => {
    lauf.result.current.beitreten('ab:cd', 'pA')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('pA')
  })

  act(() => {
    melde(ereignisZug, 'pB', zug('7'))
    melde(ereignisTischbild, 'pB', bild('pB', '7'))
    melde(ereignisZug, 'pA', zug('1'))
  })

  expect(lauf.result.current.zuege).toEqual([zug('1')])
  expect(lauf.result.current.tisch).toBeUndefined()
})

it('ohne genannte Partie nimmt er die des ersten Bildes', () => {
  const { lauf, melde } = gast(stand({ partie: 'pX' }))

  act(() => {
    lauf.result.current.beitreten('ab:cd', '')
  })
  act(() => {
    melde(ereignisTischbild, 'pX', bild('pX', '1'))
    melde(ereignisTischbild, 'pY', bild('pY', '2'))
  })

  expect(lauf.result.current.tisch?.partie.id).toBe('pX')
})

it('nimmt nach dem Aufstehen nichts mehr', async () => {
  const { lauf, melde } = gast(stand({ partie: 'pA' }))

  act(() => {
    lauf.result.current.beitreten('ab:cd', 'pA')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.laeuft).toBe(true)
  })

  act(() => {
    lauf.result.current.aufstehen()
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.laeuft).toBe(false)
  })

  act(() => {
    melde(ereignisZug, 'pA', zug('2'))
  })

  expect(lauf.result.current.zuege).toEqual([])
})

// **Zwei Beitritte nacheinander** — N2 aus dem Nachlauf über Phase 6.11:
// wer von A zu B wechselt, ohne B zu nennen, bekam womöglich zuerst eine
// verspätete Meldung von A und merkte sich A.
it('lernt die Partie nicht von dem Tisch, den er gerade verlassen hat', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  const antworten = [stand({ partie: 'pA' }), stand({ partie: '' })]
  belegeApi({
    PartieBeitreten: () =>
      Promise.resolve(antworten.shift() ?? stand({ laeuft: false })),
  })
  const lauf = renderHook(() => usePartiegast(), { wrapper: raum })

  act(() => {
    lauf.result.current.beitreten('ab:cd', 'pA')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('pA')
  })

  act(() => {
    lauf.result.current.beitreten('ef:01', '')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('')
  })

  act(() => {
    melde(ereignisZug, 'pA', zug('9'))
    melde(ereignisTischbild, 'pA', bild('pA', '9'))
    melde(ereignisTischbild, 'pB', bild('pB', '1'))
    melde(ereignisZug, 'pB', zug('1'))
  })

  expect(lauf.result.current.tisch?.partie.id).toBe('pB')
  expect(lauf.result.current.zuege).toEqual([zug('1')])
})

// **Und ein Zug lehrt keine Partie** — der Wirt grüßt mit einem Bild.
it('lernt die Partie nur aus einem Bild, nicht aus einem Zug', () => {
  const { lauf, melde } = gast(stand({ partie: '' }))

  act(() => {
    lauf.result.current.beitreten('ab:cd', '')
  })
  act(() => {
    melde(ereignisZug, 'pX', zug('1'))
    melde(ereignisTischbild, 'pY', bild('pY', '1'))
  })

  expect(lauf.result.current.zuege).toEqual([])
  expect(lauf.result.current.tisch?.partie.id).toBe('pY')
})

// **Wer sich beim selben Wirt wieder hinsetzt, sieht dessen Tisch** — P9c
// (#2377), gefunden am nativen Fenster (`netzpartie09`): nach dem Aufstehen
// stand die Partie unter „verlassen“, und jedes Bild desselben Tisches fiel
// beim Wiederhinsetzen weg. Der Schutz gilt den späten Meldungen eines
// **anderen** Tisches.
it('lernt die Partie wieder, wenn er sich beim selben Wirt hinsetzt', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  const antworten = [stand({ partie: 'pA' }), stand({ partie: '' })]
  belegeApi({
    PartieBeitreten: () =>
      Promise.resolve(antworten.shift() ?? stand({ laeuft: false })),
  })
  const lauf = renderHook(() => usePartiegast(), { wrapper: raum })

  act(() => {
    lauf.result.current.beitreten('ab:cd', 'pA')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('pA')
  })

  act(() => {
    lauf.result.current.beitreten('ab:cd', '')
  })
  await waitFor(() => {
    expect(lauf.result.current.stand?.partie).toBe('')
  })

  act(() => {
    melde(ereignisTischbild, 'pA', bild('pA', '9'))
  })

  expect(lauf.result.current.tisch?.partie.id).toBe('pA')
})
