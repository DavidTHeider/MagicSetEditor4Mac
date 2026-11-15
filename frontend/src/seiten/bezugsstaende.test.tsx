// **Die bezogenen Stände** — G04
// (#2251).
//
// Drei Stände und ein Fehler, unterscheidbar; gefragt wird erst auf
// Knopfdruck; eine späte Antwort landet nirgends. Was eine Aktualisierung
// vorher zeigt und wie der Rückweg geht, steht in `bezugsvorschau.test.tsx`.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import {
  aufschlagen,
  bezogen,
  bezug,
  spitze,
  t,
  vorschau,
} from '../test/bezuege'
import { texte } from '../test/paketseite'

import type { BezugsvorschauAntwort } from '../paketbezuege'

afterEach(() => {
  raeumeApi()
})

describe('Die bezogenen Stände', () => {
  it('nennt ohne Bezug den Weg zum Beziehen', async () => {
    await aufschlagen({})

    const kasten = (await screen.findByText(t.leer)).closest(
      '[data-leerzustand]',
    )
    if (!(kasten instanceof HTMLElement)) {
      throw new Error('der Leerzustand fehlt')
    }

    await userEvent.click(
      within(kasten).getByRole('button', {
        name: texte.seiten.pakete.plan.beziehen,
      }),
    )

    expect(
      await screen.findByRole('dialog', {
        name: texte.seiten.pakete.plan.beziehenTitel,
      }),
    ).toBeInTheDocument()
  })

  it('zeigt bezogenen und geprüften Stand, ohne zu fragen', async () => {
    const bezuege = vi.fn(() =>
      Promise.resolve({
        bezuege: [bezug(), bezug({ name: 'Eigen', geprueft: '' })],
      }),
    )
    await aufschlagen({ PaketBezuege: bezuege })

    expect(await screen.findAllByText(t.lage.ungefragt)).toHaveLength(2)
    expect(screen.getByText(new RegExp(t.ohnePruefung))).toBeInTheDocument()
    expect(bezuege).toHaveBeenCalledWith(false)
    expect(bezuege).not.toHaveBeenCalledWith(true)
  })

  // **Ein Baum ohne Paket steht da** — der erste Commit einer Quelle kann nur
  // ihre README tragen (g04bezuege). Und ein Paket ist keine Mehrzahl.
  it('zählt kein, ein und viele Pakete', async () => {
    await aufschlagen({
      PaketBezuege: () =>
        Promise.resolve({
          bezuege: [
            bezug({ name: 'Leer', pakete: 0 }),
            bezug({ name: 'Eins', pakete: 1 }),
            bezug({ name: 'Viele', pakete: 324 }),
          ],
        }),
    })

    const zeile = async (name: string) =>
      (await screen.findByText(name)).closest('li')?.textContent ?? ''

    expect(await zeile('Leer')).toContain(t.keinPaket)
    expect(await zeile('Eins')).toContain(t.paketEins)
    expect(await zeile('Viele')).toContain(t.pakete.replace('{zahl}', '324'))
  })

  it('fragt die Quellen erst auf Knopfdruck, und ein zweites Mal neu', async () => {
    const bezuege = vi.fn((fragen: boolean) =>
      Promise.resolve({
        bezuege: [
          fragen
            ? bezug({
                lage: 'neuer',
                spitze,
                spitzeZeit: '2026-10-07T02:02:04Z',
                vorsprung: 63,
              })
            : bezug(),
        ],
      }),
    )
    await aufschlagen({ PaketBezuege: bezuege })
    await screen.findByText(t.lage.ungefragt)

    await userEvent.click(screen.getByRole('button', { name: t.fragen }))
    expect(
      await screen.findByText(
        /c8e38c9 vom 7\. Oktober 2026 — 63 Commits voraus/,
      ),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: t.fragen }))
    await vi.waitFor(() => {
      expect(bezuege.mock.calls.filter(([fragen]) => fragen)).toHaveLength(2)
    })
  })

  it('sagt einen Commit voraus in der Einzahl', async () => {
    await aufschlagen({
      PaketBezuege: () =>
        Promise.resolve({
          bezuege: [
            bezug({
              lage: 'neuer',
              spitze,
              spitzeZeit: '2021-10-07T04:47:46Z',
              vorsprung: 1,
            }),
          ],
        }),
    })

    expect(
      await screen.findByText(
        'Neuer Stand c8e38c9 vom 7. Oktober 2021 — einen Commit voraus.',
      ),
    ).toBeInTheDocument()
  })

  it('unterscheidet aktuell, abweichend und unbekannt', async () => {
    await aufschlagen({
      PaketBezuege: () =>
        Promise.resolve({
          bezuege: [
            bezug({ name: 'A', lage: 'aktuell' }),
            bezug({ name: 'B', lage: 'abweichend', spitze }),
            bezug({
              name: 'C',
              lage: 'unbekannt',
              fehler: {
                code: 'unbekannt',
                schluessel: 'unbekannt',
                meldung: 'weg',
              },
            }),
          ],
        }),
    })

    expect(await screen.findByText(t.lage.aktuell)).toBeInTheDocument()
    expect(
      screen.getByText(/zurückgesetzt oder verzweigt \(Spitze c8e38c9\)/),
    ).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent(
      /Die Quelle ließ sich nicht fragen: .+/,
    )
    expect(
      screen.getAllByRole('button', { name: /Aktualisierung ansehen/ }),
    ).toHaveLength(1)
  })

  it('zeigt eine Liste, die nicht kam, als Meldung', async () => {
    await aufschlagen({ PaketBezuege: () => Promise.reject(new Error('weg')) })

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

// **Eine Antwort gehört zu ihrem Stand** (§7, Regel 2): was spät kommt,
// nachdem jemand etwas anderes angefangen hat, landet nirgends.
describe('Späte Antworten', () => {
  function aufgeschoben<T>() {
    let erfuellen: (wert: T) => void = () => undefined
    let ablehnen: (grund: unknown) => void = () => undefined
    const versprechen = new Promise<T>((ja, nein) => {
      erfuellen = ja
      ablehnen = nein
    })

    return { versprechen, erfuellen, ablehnen }
  }

  const zwei = () =>
    Promise.resolve({
      bezuege: [
        bezug({
          name: 'A',
          lage: 'neuer',
          spitze,
          spitzeZeit: '2026-10-07T02:02:04Z',
          vorsprung: 2,
        }),
        bezug({ name: 'B', vorige: true, vorigerStand: spitze }),
      ],
    })

  const zurueckGut = () =>
    Promise.resolve({
      ort: '/x',
      stand: bezogen,
      pruefsumme: '',
      uebergangen: [],
    })

  // **Befund 7 des Zweitblicks über Phase G, zweiter Teil**: ein geglücktes
  // Aktualisieren von A, das ankommt, nachdem B schon angesehen wird, räumt
  // nicht die Vorschau von B. Geschrieben ist geschrieben — die Liste lädt
  // trotzdem nach.
  it('lässt die Vorschau eines anderen Baums stehen, wenn ein älteres Aktualisieren glückt', async () => {
    const spaet = aufgeschoben<{
      ort: string
      stand: string
      pruefsumme: string
      uebergangen: string[]
    }>()
    const bezuege = vi.fn(() =>
      Promise.resolve({
        bezuege: ['A', 'B'].map((name) =>
          bezug({
            name,
            lage: 'neuer',
            spitze,
            spitzeZeit: '2026-10-07T02:02:04Z',
            vorsprung: 2,
          }),
        ),
      }),
    )
    await aufschlagen({
      PaketBezuege: bezuege,
      PaketBezugVorschau: (name: string) =>
        Promise.resolve(vorschau({ name, sets: [] })),
      PaketBezugAktualisieren: () => spaet.versprechen,
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Aktualisierung ansehen … \(A\)/,
      }),
    )
    const kastenA = await screen.findByRole('region', {
      name: 'Aktualisierung von A',
    })
    await userEvent.click(
      within(kastenA).getByRole('button', {
        name: 'Auf c8e38c9 aktualisieren',
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: /Aktualisierung ansehen … \(B\)/ }),
    )
    await screen.findByRole('region', { name: 'Aktualisierung von B' })

    spaet.erfuellen({
      ort: '/x',
      stand: spitze,
      pruefsumme: '',
      uebergangen: [],
    })

    await vi.waitFor(() => {
      expect(bezuege).toHaveBeenCalledTimes(2)
    })
    expect(
      screen.getByRole('region', { name: 'Aktualisierung von B' }),
    ).toBeInTheDocument()
  })

  it('verwirft eine Vorschau, die nach einem Zurück kommt', async () => {
    const spaet = aufgeschoben<BezugsvorschauAntwort>()
    await aufschlagen({
      PaketBezuege: zwei,
      PaketBezugVorschau: () => spaet.versprechen,
      PaketBezugZurueck: zurueckGut,
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Aktualisierung ansehen … \(A\)/,
      }),
    )
    await userEvent.click(screen.getByRole('button', { name: /zurück \(B,/ }))
    spaet.erfuellen(vorschau({ name: 'A' }))

    await vi.waitFor(() => {
      expect(
        screen.queryByRole('region', { name: 'Aktualisierung von A' }),
      ).not.toBeInTheDocument()
    })
  })

  it('verwirft einen Fehler, der nach einem Zurück kommt', async () => {
    const spaet = aufgeschoben<BezugsvorschauAntwort>()
    await aufschlagen({
      PaketBezuege: zwei,
      PaketBezugVorschau: () => spaet.versprechen,
      PaketBezugZurueck: zurueckGut,
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Aktualisierung ansehen … \(A\)/,
      }),
    )
    await userEvent.click(screen.getByRole('button', { name: /zurück \(B,/ }))
    spaet.ablehnen(new Error('weg'))
    await new Promise((weiter) => setTimeout(weiter, 0))

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('verwirft den Fehler einer Aktualisierung, die schon verworfen ist', async () => {
    const spaet = aufgeschoben<{
      ort: string
      stand: string
      pruefsumme: string
      uebergangen: string[]
    }>()
    await aufschlagen({
      PaketBezuege: zwei,
      PaketBezugVorschau: () => Promise.resolve(vorschau({ name: 'A' })),
      PaketBezugAktualisieren: () => spaet.versprechen,
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: /Aktualisierung ansehen … \(A\)/,
      }),
    )
    const kasten = await screen.findByRole('region', {
      name: 'Aktualisierung von A',
    })
    await userEvent.click(
      within(kasten).getByRole('button', { name: 'Auf c8e38c9 aktualisieren' }),
    )
    await userEvent.click(
      within(kasten).getByRole('button', { name: t.verwerfen }),
    )
    spaet.ablehnen(new Error('fort'))
    await new Promise((weiter) => setTimeout(weiter, 0))

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('verwirft den Fehler eines Zurück, nach dem schon etwas anderes begann', async () => {
    const spaet = aufgeschoben<{
      ort: string
      stand: string
      pruefsumme: string
      uebergangen: string[]
    }>()
    await aufschlagen({
      PaketBezuege: zwei,
      PaketBezugVorschau: () => Promise.resolve(vorschau({ name: 'A' })),
      PaketBezugZurueck: () => spaet.versprechen,
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /zurück \(B,/ }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: /Aktualisierung ansehen … \(A\)/ }),
    )
    spaet.ablehnen(new Error('weg'))

    expect(
      await screen.findByRole('region', { name: 'Aktualisierung von A' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('nennt eine unbekannte Lage auch ohne Grund', async () => {
    await aufschlagen({
      PaketBezuege: () =>
        Promise.resolve({ bezuege: [bezug({ lage: 'unbekannt' })] }),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Quelle ließ sich nicht fragen:',
    )
  })
})
