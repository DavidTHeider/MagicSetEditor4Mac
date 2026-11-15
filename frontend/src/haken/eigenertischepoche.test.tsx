// **Ein Ereignis des eigenen Tischs gehört zu seinem Bestand** — #2307,
// Folge aus dem Nachlauf über J05 (#2269), Befund 4.
//
// Nach dem Zurückspielen einer **älteren** Sicherung derselben Partie passen
// Kennung und Sicht weiter. Ohne Epoche übermalte ein verspätetes Tischbild
// den wiederhergestellten Stand, und ein alter Zug geriet in seine Liste —
// und umgekehrt wies der Empfänger das **richtige** Bild der neuen Epoche ab,
// weil er sich die höhere Fassung von vorher gemerkt hatte.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { epocheGesehen, epocheVergessen } from '../bestandsepoche'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { usePartie } from './usePartie'
import {
  ereignisEigenerZug as ereignisZug,
  ereignisEigenesTischbild as ereignisTischbild,
} from './useEigenerTischfunk'
import { eineAntwort, einePartie } from '../seiten/partieprobe'

import type { ReactNode } from 'react'
import type { Spielereignis } from '../apipartie'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
  epocheVergessen()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function einEreignis(folge: string): Spielereignis {
  return {
    folge,
    art: 'objekt.getappt',
    oeffentlich: {},
    privat: {},
    kette: '',
  }
}

function einBild(fassung: string) {
  return { partie: einePartie({ fassung, fuer: 'sitz1' }), erlaubt: [] }
}

async function geoeffnet(fassung = '1') {
  probe = laufzeitprobe()
  belegeApi({
    Partie: () => Promise.resolve(eineAntwort({ fuer: 'sitz1', fassung })),
  })

  const lauf = renderHook(() => usePartie(), { wrapper: raum })

  act(() => {
    lauf.result.current.ansehen('p1', 'sitz1')
  })

  await waitFor(() => {
    expect(lauf.result.current.partie?.fuer).toBe('sitz1')
  })

  return { lauf, melde: probe.melde }
}

it('verwirft ein Tischbild aus einem älteren Bestand', async () => {
  const { lauf, melde } = await geoeffnet()

  epocheGesehen(2)

  act(() => {
    melde(ereignisTischbild, 'p1', einBild('5'), 1)
  })

  expect(lauf.result.current.partie?.fassung).toBe('1')

  // **Die Gegenprobe**: dasselbe Bild aus dem jetzigen Bestand wirkt.
  act(() => {
    melde(ereignisTischbild, 'p1', einBild('6'), 2)
  })

  expect(lauf.result.current.partie?.fassung).toBe('6')
})

it('verwirft einen Zug aus einem älteren Bestand', async () => {
  const { lauf, melde } = await geoeffnet()

  epocheGesehen(2)

  act(() => {
    melde(ereignisZug, 'p1', einEreignis('1'), 1)
    melde(ereignisZug, 'p1', einEreignis('2'), 2)
  })

  expect(lauf.result.current.ereignisse).toEqual([einEreignis('2')])
})

it('nimmt nach einem Tausch die niedrigere Fassung des neuen Bestands', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisTischbild, 'p1', einBild('9'), 1)
  })

  expect(lauf.result.current.partie?.fassung).toBe('9')

  // **Im selben Bestand bleibt eine niedrigere Fassung liegen** — so war es
  // gewollt: sie kam nur später an.
  act(() => {
    melde(ereignisTischbild, 'p1', einBild('4'), 1)
  })

  expect(lauf.result.current.partie?.fassung).toBe('9')

  // **Aus dem neuen Bestand gilt sie** — die Sicherung war älter.
  act(() => {
    melde(ereignisTischbild, 'p1', einBild('3'), 2)
  })

  expect(lauf.result.current.partie?.fassung).toBe('3')
})

// **Das erste Ereignis nach dem Öffnen ist kein neuer Bestand** — der Haken
// beginnt mit der Epoche, die das Fenster kennt. Sonst überschriebe ein
// verspätetes Bild derselben Epoche die eben geladene Partie.
it('lässt ein verspätetes Bild derselben Epoche nach dem Öffnen liegen', async () => {
  epocheGesehen(1)
  const { lauf, melde } = await geoeffnet('6')

  act(() => {
    melde(ereignisTischbild, 'p1', einBild('4'), 1)
  })

  expect(lauf.result.current.partie?.fassung).toBe('6')
})

// **Mehrere Bilder des neuen Bestands in einem Zug** (Zweitblick über die
// Folgen vor 0.8.0, Befund 7): der Nachholpuffer liefert synchron, und die
// Untergrenze war bis zum nächsten Effekt noch die Fassung des alten Bestands
// — 3 kam an, 4 wurde gegen 9 abgewiesen.
it('nimmt nach einem Tausch mehrere Bilder des neuen Bestands nacheinander', async () => {
  const { lauf, melde } = await geoeffnet()

  act(() => {
    melde(ereignisTischbild, 'p1', einBild('9'), 1)
  })

  act(() => {
    melde(ereignisTischbild, 'p1', einBild('3'), 2)
    melde(ereignisTischbild, 'p1', einBild('4'), 2)
  })

  expect(lauf.result.current.partie?.fassung).toBe('4')
})
