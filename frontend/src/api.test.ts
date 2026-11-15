import { afterEach, describe, expect, it, vi } from 'vitest'

import { api, auspacken, Bindungsfehler, fassade } from './api'
import { belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
})

describe('die Fassade', () => {
  // **Der Fall, für den sie da ist.** Eine Fassade, die stillschweigend nichts
  // zurückgibt, macht aus einem vergessenen Beleg eine leere Seite — und dann
  // sucht jemand im Zustand statt in der Vorrichtung.
  it('meldet eine Methode, die niemand belegt hat', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [] }) })

    await expect(api().Karten('1')).rejects.toMatchObject({
      schluessel: 'methodeFehlt',
      werte: { name: 'Karten' },
    })
    await expect(api().Vorbehalte('1')).rejects.toMatchObject({
      schluessel: 'methodeFehlt',
      werte: { name: 'Vorbehalte' },
    })
  })

  it('reicht eine belegte Methode durch', async () => {
    const sets = vi.fn().mockResolvedValue([
      {
        id: '1',
        name: 'a',
        spiel: 'magic',
        stil: 'm15',
        herkunft: '',
        karten: 0,
        herkunftFehlt: false,
      },
    ])

    belegeApi({ Sets: sets })

    await expect(api().Sets()).resolves.toHaveLength(1)
    expect(sets).toHaveBeenCalledOnce()
  })

  // Ohne Wails-Laufzeit meldet sich **jede** Methode — statt die Anwendung mit
  // einem `undefined is not a function` anzuhalten.
  it('meldet sich auch ganz ohne Laufzeit', async () => {
    raeumeApi()

    await expect(api().Sets()).rejects.toMatchObject({
      schluessel: 'methodeFehlt',
      werte: { name: 'Sets' },
    })
  })

  it('baut auch aus nichts eine vollständige Fassade', async () => {
    const gebaut = fassade(undefined)

    await expect(gebaut.Karten('1')).rejects.toMatchObject({
      schluessel: 'methodeFehlt',
      werte: { name: 'Karten' },
    })
  })

  // **Aufräumen ist Pflicht.** Ein Test, der grün ist, weil ein anderer vor ihm
  // lief, ist keiner.
  it('nimmt die Belegung wieder weg', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [] }) })
    await expect(api().Sets()).resolves.toEqual({ sets: [] })

    raeumeApi()
    await expect(api().Sets()).rejects.toMatchObject({
      schluessel: 'methodeFehlt',
      werte: { name: 'Sets' },
    })
  })
})

describe('auspacken', () => {
  // **Der Umschlag ist der Grund, warum es diese Funktion gibt.** Wails nimmt
  // einen zweiten Rückgabewert nur dann als Fehler, wenn er Gos `error`
  // erfüllt — ein strukturierter Fehler tut das nicht und wäre stillschweigend
  // verworfen worden.
  it('gibt den Inhalt, wenn kein Fehler dasteht', () => {
    expect(auspacken({}, () => [1, 2])).toEqual([1, 2])
  })

  it('lehnt ab und behält Code und Feldbezug', () => {
    const umschlag = {
      sets: [],
      fehler: {
        code: 'validierung' as const,
        schluessel: 'pflichtfeld',
        meldung: 'Ein Feld braucht einen Namen.',
        feld: 'feld',
      },
    }

    try {
      auspacken(umschlag, () => umschlag.sets)
      expect.unreachable('auspacken hätte ablehnen müssen')
    } catch (ursache: unknown) {
      expect(ursache).toBeInstanceOf(Bindungsfehler)

      const fehler = ursache as Bindungsfehler
      expect(fehler.code).toBe('validierung')
      expect(fehler.feld).toBe('feld')
      expect(fehler.message).toBe('Ein Feld braucht einen Namen.')
    }
  })

  it('lässt ein `null` als kein Fehler gelten', () => {
    expect(auspacken({ fehler: null }, () => [1])).toEqual([1])
  })
})
