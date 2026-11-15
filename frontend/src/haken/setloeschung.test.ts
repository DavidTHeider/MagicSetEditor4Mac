// **Der Haken hinter dem Löschen eines Sets** —
// #1218.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { useSetloeschung } from './useSetloeschung'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

it('meldet das gelöschte Set und lädt nach', async () => {
  const geloescht = vi.fn()
  const nachladen = vi.fn()
  belegeApi({
    SetLoeschen: () => Promise.resolve({ befunde: [], fehler: null }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(texte, nachladen, geloescht),
  )

  await act(async () => {
    await result.current.loeschen('7', 'Mein Set')
  })

  await waitFor(() => {
    expect(geloescht).toHaveBeenCalledWith('7')
  })
  expect(nachladen).toHaveBeenCalled()
})

it('reicht den Namen durch, den jemand getippt hat', async () => {
  const gerufen = vi.fn(() => Promise.resolve({ befunde: [], fehler: null }))
  belegeApi({ SetLoeschen: gerufen })

  const { result } = renderHook(() => useSetloeschung(texte, vi.fn(), vi.fn()))

  await act(async () => {
    await result.current.loeschen('7', 'Getippt')
  })

  expect(gerufen).toHaveBeenCalledWith('7', 'Getippt')
})

it('hält die Befunde eines geglückten Löschens am Set fest', async () => {
  belegeApi({
    SetLoeschen: () => Promise.resolve({ befunde: ['Reste'], fehler: null }),
  })

  const { result } = renderHook(() => useSetloeschung(texte, vi.fn(), vi.fn()))

  await act(async () => {
    await result.current.loeschen('7', 'Mein Set')
  })

  expect(result.current.befunde('7')).toEqual(['Reste'])
  expect(result.current.befunde('8')).toEqual([])
})

it('meldet den Fehler am Set und lädt nicht nach', async () => {
  const geloescht = vi.fn()
  const nachladen = vi.fn()
  belegeApi({
    SetLoeschen: () =>
      Promise.resolve({
        befunde: [],
        fehler: {
          code: 'konflikt',
          meldung: 'Das Set heißt anders.',
          schluessel: 'konflikt',
        },
      }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(texte, nachladen, geloescht),
  )

  await act(async () => {
    await result.current.loeschen('7', 'Falsch')
  })

  // **Der Satz kommt aus dem Katalog und nicht aus Go** — so wie bei jedem
  // anderen Fehler (§9). Und er passt: am Fenster ist der Knopf gesperrt,
  // solange der getippte Name nicht stimmt, also kann ein Konflikt hier nur
  // heißen, dass jemand anderes das Set umbenannt hat. Genau das sagt er.
  expect(result.current.fehler('7')).toBe(texte.fehler.konflikt)
  expect(geloescht).not.toHaveBeenCalled()

  // **Nach einem gescheiterten Schreiben wird nicht nachgeladen** (§7,
  // Regel 3): ein Lesefehler überschriebe sonst die Meldung, die sagt, was zu
  // tun ist.
  expect(nachladen).not.toHaveBeenCalled()
})

it('lässt keinen zweiten Lauf gegen dasselbe Set zu', async () => {
  let loesen: (() => void) | undefined
  const gerufen = vi.fn(
    () =>
      new Promise<{ befunde: string[]; fehler: null }>((fertig) => {
        loesen = () => {
          fertig({ befunde: [], fehler: null })
        }
      }),
  )
  belegeApi({ SetLoeschen: gerufen })

  const { result } = renderHook(() => useSetloeschung(texte, vi.fn(), vi.fn()))

  let erster: Promise<void> | undefined
  act(() => {
    erster = result.current.loeschen('7', 'Mein Set')
  })

  expect(result.current.laeuft('7')).toBe(true)

  await act(async () => {
    await result.current.loeschen('7', 'Mein Set')
  })

  expect(gerufen).toHaveBeenCalledTimes(1)

  await act(async () => {
    loesen?.()
    await erster
  })
})

it('räumt den Fehler weg, wenn es beim zweiten Mal glückt', async () => {
  let scheitern = true
  belegeApi({
    SetLoeschen: () =>
      Promise.resolve(
        scheitern
          ? {
              befunde: [],
              fehler: {
                code: 'konflikt',
                meldung: 'Das Set heißt anders.',
                schluessel: 'konflikt',
              },
            }
          : { befunde: [], fehler: null },
      ),
  })

  const { result } = renderHook(() => useSetloeschung(texte, vi.fn(), vi.fn()))

  await act(async () => {
    await result.current.loeschen('7', 'Falsch')
  })

  // **Der Beleg, dass überhaupt etwas dastand.** Ohne ihn sagte die
  // Verneinung darunter nichts — sie träfe auch auf einen Haken zu, der nie
  // einen Fehler einträgt.
  expect(result.current.fehler('7')).toBe(texte.fehler.konflikt)

  scheitern = false

  await act(async () => {
    await result.current.loeschen('7', 'Richtig')
  })

  expect(result.current.fehler('7')).toBeUndefined()
})
