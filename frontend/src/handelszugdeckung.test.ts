// **Was weggeht, ist vorbelegt — wenn es eindeutig ist** — P13
// (#2381),
// Befund C-16: das Exemplar wurde nach der Einigung in einer eigenen Phase
// gewählt, obwohl es oft nur eines gibt. Der Zug hält es fest, wo genau eine
// Zeile des Bestands reicht, und hält sonst an.
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ziehe } from './handelszug'
import { du, einDienst, einFehler, einStand, ich } from './test/handelszugprobe'
import { belegeApi, raeumeApi } from './test/api'

import type { ExemplarDTO } from './api'
import type {
  Handelsvorgang as Vorgangsstand,
  Handelszusage,
} from './apihandelsleitung'

afterEach(() => {
  raeumeApi()
})

const zeile = (id: number, menge: number, teil: Partial<ExemplarDTO> = {}) =>
  ({
    id,
    menge,
    fassung: 3,
    wunsch: false,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    ...teil,
  }) as ExemplarDTO

const eigeneZeile = {
  von: ich,
  printing: 'p-glut',
  druckId: 0,
  menge: 1,
  guete: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
  abbild: '',
  bewertung: 0,
}

const fremdeZeile = {
  ...eigeneZeile,
  von: du,
  printing: 'p-sturm',
  guete: '',
  sprache: '',
  finish: '',
}

const angenommen = einStand({
  zustand: 'angenommen',
  naechster: 'vorbereiten',
  positionen: [eigeneZeile, fremdeZeile],
})

// gesehen ist der Stand, den der Zug zuletzt vorgelegt bekam — der Dienst
// bestätigt ihn beim Nachlesen vor dem ersten Schritt.
let gesehen = angenommen

// zieheVorbereiten geht den Zug „Vorbereiten und übergeben" von `stand` aus.
function zieheVorbereiten(stand: Vorgangsstand) {
  gesehen = stand

  return ziehe({
    stand,
    schritt: 'vorbereiten',
    bis: 'uebergeben',
    gegenueber: du,
  })
}

function stelle(zeilen: ExemplarDTO[], zusagen: Handelszusage[] = []) {
  const { api, rufe } = einDienst({
    vorbereiten: einStand({ zustand: 'vorbereitet', lage: 'wartet' }),
  })
  const besitz = vi.fn((printing: string) =>
    Promise.resolve({
      besitz: {
        printing,
        stueck: 1,
        gewuenscht: 0,
        besitzt: true,
        zeilen,
      },
      fehler: null,
    }),
  )
  const festhalten = vi.fn(() =>
    Promise.resolve({ vorgang: 'v-1', zusagen: [], fehler: null }),
  )
  belegeApi({
    ...api,
    SammlungBesitz: besitz,
    Handelszusagen: () =>
      Promise.resolve({ vorgang: 'v-1', zusagen, fehler: null }),
    HandelFesthalten: festhalten,
    HandelStand: () => Promise.resolve({ vorgang: gesehen, fehler: null }),
  } as never)

  return { rufe, besitz, festhalten }
}

describe('das Festhalten im Zug', () => {
  it('hält die eine Zeile fest, die reicht — nur für die eigenen Zeilen', async () => {
    const { rufe, besitz, festhalten } = stelle([
      zeile(5, 2),
      zeile(6, 1, { wunsch: true }),
    ])

    await zieheVorbereiten(angenommen)

    expect(besitz).toHaveBeenCalledTimes(1)
    expect(besitz).toHaveBeenCalledWith('p-glut')
    expect(festhalten).toHaveBeenCalledWith('v-1', 5, 1, 3)
    expect(rufe[0]).toBe('vorbereiten')
  })

  it('nimmt nur eine Zeile, die zur Absprache passt', async () => {
    // **Gefunden mit `nat04handel`**: eine Zeile in anderem Zustand ging
    // weg, drüben kam die abgesprochene an, und die Belege passten nicht
    // zusammen — die Gegenzeichnung blieb still aus.
    const { festhalten } = stelle([
      zeile(5, 1, { zustand: 'excellent' }),
      zeile(6, 1),
    ])

    await zieheVorbereiten(angenommen)

    expect(festhalten).toHaveBeenCalledWith('v-1', 6, 1, 3)
  })

  it.each<[string, Partial<ExemplarDTO>]>([
    ['Zustand', { zustand: 'excellent' }],
    ['Sprache', { sprache: 'en' }],
    ['Ausführung', { finish: 'foil' }],
  ])('hält an, wenn die einzige Zeile im %s abweicht', async (_fall, teil) => {
    const { rufe, festhalten } = stelle([zeile(5, 2, teil)])

    const ausgang = await zieheVorbereiten(angenommen)

    expect(festhalten).not.toHaveBeenCalled()
    expect(rufe).toEqual([])
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
  })

  it('nimmt keine Zeile, wo die Absprache Zustand oder Sprache offen lässt', async () => {
    // **#2492**: der Dienst hält nur fest, was eine Zeile erfüllt — eine
    // offene Angabe erfüllt kein Exemplar. Bis dahin nahm der Zug „jede
    // Zeile des Drucks", und drüben kam etwas anderes an als hier wegging.
    const { festhalten } = stelle([zeile(5, 1, { zustand: 'excellent' })])

    const ausgang = await zieheVorbereiten(
      einStand({
        ...angenommen,
        positionen: [{ ...eigeneZeile, guete: '', sprache: '' }],
      }),
    )

    expect(festhalten).not.toHaveBeenCalled()
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
  })

  it('nimmt eine Zeile ohne Ausführung als nonfoil — wie die Domäne', async () => {
    const { festhalten } = stelle([
      zeile(5, 1, { finish: '' as ExemplarDTO['finish'] }),
    ])

    await zieheVorbereiten(angenommen)

    expect(festhalten).toHaveBeenCalledWith('v-1', 5, 1, 3)
  })

  it('hält nichts ein zweites Mal fest', async () => {
    const { rufe, festhalten } = stelle(
      [zeile(5, 2)],
      [
        {
          exemplarId: 5,
          menge: 1,
          fassung: 3,
          revision: 1,
          uebergeben: false,
        },
      ],
    )

    await zieheVorbereiten(angenommen)

    expect(festhalten).not.toHaveBeenCalled()
    expect(rufe[0]).toBe('vorbereiten')
  })

  it.each([
    ['zwei Zeilen kämen in Frage', [zeile(5, 1), zeile(7, 1)]],
    ['keine Zeile reicht', [zeile(5, 0)]],
    ['es gibt keine', []],
  ])('hält an und geht keinen Schritt, wenn %s', async (_fall, zeilen) => {
    const { rufe, festhalten } = stelle(zeilen)

    const ausgang = await zieheVorbereiten(angenommen)

    expect(festhalten).not.toHaveBeenCalled()
    expect(rufe).toEqual([])
    expect(ausgang.gescheitert).toEqual({
      teil: 'festhalten',
      schritt: 'vorbereiten',
      ursache: null,
    })
  })

  it('lässt eine Zusage auf einer unpassenden Zeile nicht als Deckung gelten', async () => {
    // **Befund 4 des Zweitblicks über Phase 8**: die Prüfung der Merkmale
    // kam erst nach dem Zählen — eine schon festgehaltene Excellent-Zeile
    // deckte eine abgesprochene Near-Mint-Zeile.
    const { rufe, festhalten } = stelle(
      [zeile(5, 1, { zustand: 'excellent' }), zeile(6, 1)],
      [{ exemplarId: 5, menge: 1, fassung: 3, revision: 1, uebergeben: false }],
    )

    const ausgang = await zieheVorbereiten(angenommen)

    expect(festhalten).not.toHaveBeenCalled()
    expect(rufe).toEqual([])
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
  })

  it('hält an, wenn schon ein Teil festgehalten ist — wie viel noch fehlt, sagt ein Mensch', async () => {
    const { rufe, festhalten } = stelle(
      [zeile(5, 3)],
      [
        {
          exemplarId: 5,
          menge: 1,
          fassung: 3,
          revision: 1,
          uebergeben: false,
        },
      ],
    )

    const ausgang = await zieheVorbereiten(
      einStand({
        ...angenommen,
        positionen: [{ ...eigeneZeile, menge: 2 }],
      }),
    )

    expect(festhalten).not.toHaveBeenCalled()
    expect(rufe).toEqual([])
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
  })

  it('hält an, wenn derselbe Druck in zwei eigenen Zeilen steht', async () => {
    const { rufe, besitz } = stelle([zeile(5, 3)])

    const ausgang = await zieheVorbereiten(
      einStand({
        ...angenommen,
        positionen: [eigeneZeile, { ...eigeneZeile, guete: 'excellent' }],
      }),
    )

    expect(besitz).not.toHaveBeenCalled()
    expect(rufe).toEqual([])
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
  })

  it('nennt den Fehler, wenn das Festhalten selbst scheitert', async () => {
    const { rufe, festhalten } = stelle([zeile(5, 2)])
    festhalten.mockImplementation(
      () =>
        Promise.resolve({
          vorgang: 'v-1',
          zusagen: [],
          fehler: einFehler('Die Zeile hat sich geändert.'),
        }) as never,
    )

    const ausgang = await zieheVorbereiten(angenommen)

    expect(rufe).toEqual([])
    expect(ausgang.gescheitert?.teil).toBe('festhalten')
    expect(String(ausgang.gescheitert?.ursache)).toContain(
      'Die Zeile hat sich geändert.',
    )
  })

  it('braucht nichts festzuhalten, wo nichts weggeht', async () => {
    const { rufe, besitz } = stelle([])

    await zieheVorbereiten(
      einStand({
        ...angenommen,
        positionen: [fremdeZeile],
      }),
    )

    expect(besitz).not.toHaveBeenCalled()
    expect(rufe[0]).toBe('vorbereiten')
  })
})
