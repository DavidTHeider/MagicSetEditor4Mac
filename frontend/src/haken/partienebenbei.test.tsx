// **Was neben dem Spiel läuft, im Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Gegen die zwei Hälften, die das Fenster hält.** Die Go-Seite hält den
// **Stand** — eine überholte Präsenz kommt gar nicht erst an. Hier steht der
// **Ablauf**: ein Zeiger, zu dem nichts mehr nachkommt, verschwindet. Und
// hier steht, dass ein Zeiger den vorigen desselben Sitzes **ersetzt**, statt
// sich neben ihn zu legen — sonst wüchse bei zwanzig Meldungen je Sekunde
// eine Liste, die niemand liest.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitbelegen } from '../seiten/partiegastprobe'
import {
  Gespraechslaenge,
  Praesenzfrist,
  Zeigertakt,
  ereignisNebenbei,
  usePartienebenbei,
} from './usePartienebenbei'

import type { Nebenbei } from '../apinebenbei'
import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
  vi.useRealTimers()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

// einSatz baut, was Go meldet.
function einSatz(von: string, text: string): Nebenbei {
  return { art: 'chat', von, text, zeigerart: '', objekt: '', x: 0, y: 0 }
}

// einZeiger baut, was Go bei Präsenz meldet.
function einZeiger(von: string, x: number): Nebenbei {
  return {
    art: 'praesenz',
    von,
    text: '',
    zeigerart: 'zeiger',
    objekt: 'k1',
    x,
    y: 0,
  }
}

// **Ein Satz wächst an, ein Zeiger ersetzt den vorigen desselben Sitzes.**
it('sammelt Sätze und hält je Sitz einen Zeiger', async () => {
  const horcher = laufzeitbelegen()

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  const melde = horcher.get(ereignisNebenbei)
  expect(melde).toBeDefined()

  act(() => {
    melde?.('p1', einSatz('bert', 'hallo'))
    melde?.('p1', einSatz('bert', 'und tschüss'))
    melde?.('p1', einZeiger('bert', 10))
    melde?.('p1', einZeiger('bert', 20))
    melde?.('p1', einZeiger('carla', 30))
  })

  await waitFor(() => {
    expect(result.current.gespraech).toHaveLength(2)
  })

  expect(result.current.gespraech[1]?.text).toBe('und tschüss')

  // **Zwei Sitze, zwei Zeiger** — und berts zweiter hat seinen ersten
  // abgelöst, statt sich danebenzulegen.
  expect(result.current.zeiger).toHaveLength(2)
  expect(result.current.zeiger.find((einer) => einer.von === 'bert')?.x).toBe(
    20,
  )
})

// **Ein Zeiger, zu dem nichts nachkommt, verschwindet.**
//
// **Mit einer gestellten Uhr**, denn eine echte Frist von fünfzehn Sekunden
// in einem Lauf wäre fünfzehn Sekunden Wartezeit je Durchgang.
it('lässt einen Zeiger nach der Frist fallen', () => {
  vi.useFakeTimers()

  const horcher = laufzeitbelegen()

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  act(() => {
    horcher.get(ereignisNebenbei)?.('p1', einZeiger('bert', 10))
  })

  expect(result.current.zeiger).toHaveLength(1)

  // **Ein Takt vor der Frist ändert nichts** — die Gegenprobe dazu, dass der
  // Zeiger nicht einfach beim ersten Schlag verschwindet (§10).
  act(() => {
    vi.advanceTimersByTime(Zeigertakt)
  })

  expect(result.current.zeiger).toHaveLength(1)

  act(() => {
    vi.advanceTimersByTime(Praesenzfrist)
  })

  expect(result.current.zeiger).toHaveLength(0)
})

// **Das eigene Wort steht erst da, wenn es hinausging.**
it('trägt einen eigenen Satz nach dem Senden ein', async () => {
  laufzeitbelegen()

  const sagen = vi.fn().mockResolvedValue({ fehler: null })
  belegeApi({ PartieSagen: sagen })

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  act(() => {
    result.current.sagen('gut gezogen')
  })

  await waitFor(() => {
    expect(result.current.gespraech).toHaveLength(1)
  })

  expect(sagen).toHaveBeenCalledWith('gut gezogen')
  expect(result.current.gespraech[0]?.von).toBe('anna')
})

// **Was nicht hinausging, steht auch nicht da.**
it('meldet einen gescheiterten Satz und trägt ihn nicht ein', async () => {
  laufzeitbelegen()

  belegeApi({
    PartieSagen: vi.fn().mockResolvedValue({
      fehler: { code: 'konflikt', meldung: 'kein Tisch' },
    }),
  })

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  act(() => {
    result.current.sagen('hallo')
  })

  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })

  expect(result.current.gespraech).toHaveLength(0)
})

// **Ein Zeiger geht hinaus, mit Art, Objekt und Ort.**
//
// **Die Weltkoordinate geht durch, wie sie kommt.** Sie ist in Millimetern der
// Tischwelt und nicht in Bildpunkten — beide Enden rechnen sie aus derselben
// Sicht, und wer hier umrechnete, rechnete für ein Fenster.
it('schickt einen Zeiger hinaus', async () => {
  laufzeitbelegen()

  const zeigen = vi.fn().mockResolvedValue({ fehler: null })
  belegeApi({ PartieZeigen: zeigen })

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  act(() => {
    result.current.zeigen('greift', 'k7', 612.5, -140)
  })

  await waitFor(() => {
    expect(zeigen).toHaveBeenCalledWith('greift', 'k7', 612.5, -140)
  })

  // **Der eigene Zeiger steht nicht im eigenen Fenster.** Es weiß, wo die
  // eigene Maus ist; was hier ankommt, sind die der anderen.
  expect(result.current.zeiger).toHaveLength(0)
  expect(result.current.fehler).toBeUndefined()
})

// **Ein verlorener Zeiger ist kein Befund.**
//
// **Die Gegenprobe zum Satz darüber** (§10): derselbe Fehler, und diesmal
// bleibt das Fenster still — er ist flüchtig, der nächste kommt in
// Millisekunden, und eine Meldung dafür stünde dauerhaft da.
it('schweigt über einen gescheiterten Zeiger', async () => {
  laufzeitbelegen()

  const zeigen = vi.fn().mockRejectedValue(new Error('die Leitung ist weg'))
  belegeApi({ PartieZeigen: zeigen })

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  act(() => {
    result.current.zeigen('zeiger', 'k1', 10, 20)
  })

  await waitFor(() => {
    expect(zeigen).toHaveBeenCalledWith('zeiger', 'k1', 10, 20)
  })

  expect(result.current.fehler).toBeUndefined()
})

// **Das Gespräch wächst nicht über seine Grenze.**
it('behält höchstens hundert Sätze', async () => {
  const horcher = laufzeitbelegen()

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  const melde = horcher.get(ereignisNebenbei)

  act(() => {
    for (let zahl = 0; zahl <= Gespraechslaenge; zahl += 1) {
      melde?.('p1', einSatz('bert', `satz ${String(zahl)}`))
    }
  })

  await waitFor(() => {
    expect(result.current.gespraech).toHaveLength(Gespraechslaenge)
  })

  // **Vorn fällt heraus, hinten kommt an** — der erste ist weg, der letzte da.
  expect(result.current.gespraech[0]?.text).toBe('satz 1')
  expect(result.current.gespraech[Gespraechslaenge - 1]?.text).toBe(
    `satz ${String(Gespraechslaenge)}`,
  )
})

// **Ohne Laufzeit hängt kein Horcher, und das bricht nichts.**
it('kommt ohne Meldeweg aus', () => {
  ;(globalThis as { runtime?: unknown }).runtime = undefined

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'probe', 'p1'),
    {
      wrapper: raum,
    },
  )

  expect(result.current.gespraech).toHaveLength(0)
  expect(result.current.zeiger).toHaveLength(0)
})
