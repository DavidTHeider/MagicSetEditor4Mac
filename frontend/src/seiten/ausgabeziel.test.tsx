// **Der Assistent merkt sich je Set, wohin es zuletzt ging** — AP-10
// (#2367),
// Befund B-20: „Das letzte Ziel je Set merken." Über einen Neustart trägt
// es die Einstellung `ausgabe.ziele`; hier steht, dass das Fenster sie liest
// und schreibt — ob sie den Neustart wirklich übersteht, zeigt der Lauf am
// nativen Fenster.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import type { Ausgabeseite } from '../ausgabe'
import { Ausgeben } from './Ausgeben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

function satz(id: string) {
  return {
    id,
    name: `Set ${id}`,
    spiel: 'magic',
    stil: 'm15',
    herkunft: `/pfad/${id}.mse-set`,
    karten: 2,
    herkunftFehlt: false,
  }
}

function zeige(
  gemerkt: string,
  belegung: Record<string, unknown> = {},
  seite: Ausgabeseite = 'ausgeben',
) {
  const Einstellungen = vi.fn(() =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'ausgabe.ziele',
          wert: gemerkt,
          wirksam: gemerkt,
          werte: [],
          bereich: 'arbeitsplatz',
        },
      ],
    }),
  )

  belegeApi({
    Sets: () => Promise.resolve({ sets: [satz('4'), satz('7')] }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Einstellungen,
    EinstellungSetzen: () => Promise.resolve({}),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Ausgeben setID="4" seite={seite} wechseln={vi.fn()} />
    </Sprachraum>,
  )

  return Einstellungen
}

// geladen wartet, bis die Einstellungen da sind und ihre Effekte gelaufen
// — **vor** einer Verneinung: „nichts vorgewählt" trifft auch zu, solange
// noch gar nichts geladen ist (§3).
async function geladen(Einstellungen: ReturnType<typeof vi.fn>) {
  await waitFor(() => {
    expect(Einstellungen).toHaveBeenCalled()
  })
  await effekteNachholen()
}

// geschrieben gibt die Liste, die ein Ruf von `EinstellungSetzen` trug.
function geschrieben(ruf: unknown[] | undefined) {
  expect(ruf?.[0]).toBe('ausgabe.ziele')

  return JSON.parse(String(ruf?.[1])) as unknown
}

const beide = JSON.stringify([
  { set: '4', ziel: 'druckbogen' },
  { set: '7', ziel: 'mtgjson' },
])

describe('Das gemerkte Ziel', () => {
  it('ist für dieses Set vorgewählt, und seine Fläche steht da', async () => {
    // **Erst geladen** — die Wahl steht vor den Einstellungen schon da,
    // nur leer; unter Last fand der Test sie zu früh.
    await geladen(zeige(beide))

    expect(
      screen.getByRole('radio', { name: texte.druckbogen.titel }),
    ).toBeChecked()
    expect(
      await screen.findByRole('region', { name: texte.druckbogen.titel }),
    ).toBeInTheDocument()
  })

  it('ist das eines anderen Sets nicht', async () => {
    await geladen(zeige(JSON.stringify([{ set: '7', ziel: 'mtgjson' }])))

    expect(await screen.findByText(texte.ausgeben.keinZiel)).toBeVisible()
  })

  it('wird bei einer Wahl fortgeschrieben, hinten und nur einmal je Set', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))

    zeige(beide, { EinstellungSetzen })

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.bilder.titel }),
    )

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    expect(geschrieben(EinstellungSetzen.mock.calls[0])).toEqual([
      { set: '7', ziel: 'mtgjson' },
      { set: '4', ziel: 'bilder' },
    ])
  })

  // **Ein Wunsch der Seite ist auch ein Ziel** — wer mit ⌥⌘1 kam, sah die
  // Bilder zuletzt.
  it('folgt einem Wunsch der Seite und merkt ihn', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))

    zeige(beide, { EinstellungSetzen }, 'bilder')

    expect(
      await screen.findByRole('radio', { name: texte.bilder.titel }),
    ).toBeChecked()
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    expect(geschrieben(EinstellungSetzen.mock.calls[0])).toEqual([
      { set: '7', ziel: 'mtgjson' },
      { set: '4', ziel: 'bilder' },
    ])
  })

  // **Eine zweite Wahl, während die erste geschrieben wird, geht nicht
  // verloren** — der Vorgang je Einstellung wiese sie sonst still ab (§7,
  // Regel 4), und nach dem Neustart stünde die erste da.
  it('schreibt die letzte von zwei schnellen Wahlen', async () => {
    let erste: ((wert: object) => void) | undefined
    const EinstellungSetzen = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            erste = fertig
          }),
      )
      .mockResolvedValue({})

    zeige(beide, { EinstellungSetzen })

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.bilder.titel }),
    )
    await waitFor(() => {
      expect(erste).toBeDefined()
    })
    await userEvent.click(
      screen.getByRole('radio', { name: texte.sets.vorlageWaehlen }),
    )

    erste?.({})

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(2)
    })
    expect(geschrieben(EinstellungSetzen.mock.calls[1])).toEqual([
      { set: '7', ziel: 'mtgjson' },
      { set: '4', ziel: 'vorlage' },
    ])
  })

  it('hält höchstens zweihundert und lässt die ältesten fallen', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))
    const voll = Array.from({ length: 200 }, (_, stelle) => ({
      set: String(stelle + 10),
      ziel: 'bilder',
    }))

    zeige(JSON.stringify(voll), { EinstellungSetzen })

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.druckbogen.titel }),
    )

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    const liste = geschrieben(EinstellungSetzen.mock.calls[0]) as unknown[]
    expect(liste).toHaveLength(200)
    expect(liste[0]).toEqual({ set: '11', ziel: 'bilder' })
    expect(liste[199]).toEqual({ set: '4', ziel: 'druckbogen' })
  })

  it.each([
    ['kein JSON', '[{"set":"4","ziel":"pdf"}, kaputt'],
    ['ein Objekt statt einer Liste', '{"4":"druckbogen"}'],
    [
      'Einträge ohne Gestalt',
      '[null, {"set":4,"ziel":"druckbogen"}, {"set":"4"}]',
    ],
  ])(
    'wählt aus einem unlesbaren Gedächtnis nichts vor — %s',
    async (_, wert) => {
      const EinstellungSetzen = vi.fn(() => Promise.resolve({}))

      await geladen(zeige(wert, { EinstellungSetzen }))

      expect(await screen.findByText(texte.ausgeben.keinZiel)).toBeVisible()

      // **Und gelesen wurde es** — die nächste Wahl ersetzt, was nicht zu
      // lesen war, statt es mitzuschleppen.
      await userEvent.click(
        screen.getByRole('radio', { name: texte.bilder.titel }),
      )
      await waitFor(() => {
        expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
      })
      expect(geschrieben(EinstellungSetzen.mock.calls[0])).toEqual([
        { set: '4', ziel: 'bilder' },
      ])
    },
  )

  it('wählt kein Ziel vor, das es nicht gibt', async () => {
    await geladen(zeige(JSON.stringify([{ set: '4', ziel: 'pdf' }])))

    expect(await screen.findByText(texte.ausgeben.keinZiel)).toBeVisible()
  })

  // **Ein Merken, das scheitert, sagt es** — die Wahl gilt, beim nächsten
  // Start stünde sie aber nicht da.
  it('sagt, wenn sich das Ziel nicht merken ließ', async () => {
    zeige(beide, {
      EinstellungSetzen: () =>
        Promise.resolve({
          fehler: { code: 'validierung', meldung: 'Der Bestand klemmt.' },
        }),
    })

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.bilder.titel }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.ausgeben.nichtGemerkt,
    )
    expect(
      screen.getByRole('radio', { name: texte.bilder.titel }),
    ).toBeChecked()
  })
})
