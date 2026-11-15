// **Der Tausch als Tisch, der Assistent als Liste, und Anbieten in einem
// Schritt** — P12
// (#2380),
// Befund C-14: H1 ≤ 5 — dafür gehen Eröffnen und Anbieten in einen Knopf,
// und der bisherige Assistent bleibt als Liste erreichbar.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelsvorgang } from './Handelsvorgang'
import { Handelszusammenstellung } from './Handelszusammenstellung'
import { Zugmeldung } from './Zugmeldung'
import { useHandelszug } from '../haken/useHandelszug'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { mitStandnachlauf } from '../test/standnachlauf'
import { einAbgleich } from '../test/tauschabgleich'

import type {
  Handelsvorgang as Vorgangsstand,
  Handelsvorgangantwort,
} from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
  try {
    localStorage.clear()
  } catch {
    // ohne Speicher gilt die Vorgabe
  }
})

function Buehne({ geaendert }: { geaendert: (neu: Vorgangsstand) => void }) {
  const lauf = useVorgang()
  const zug = useHandelszug('du-1', lauf)

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Handelsvorgang
          kennung="du-1"
          name="Bert"
          vorgang=""
          stand={undefined}
          geaendert={geaendert}
          gesehen={() => {
            /* nur lesen */
          }}
          zug={zug}
          lauf={lauf}
        />
        {/* **Wie im Raum**: was der Zug sagt, steht unter dem Kasten. */}
        <Zugmeldung
          ausgang={zug.ausgang}
          name="Bert"
          angehalten={zug.laeuft}
          nochmal={(bisher, legen) => {
            void zug.nochmal(bisher, legen)
          }}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

const einStand = (zustand: string): Vorgangsstand =>
  ({
    vorgang: 'v-1',
    revision: 1,
    abdruck: `a-${zustand}`,
    zustand,
    positionen: [],
    parteien: [],
    zustimmungen: [],
    beschluss: '',
    lage: 'wartet',
    naechster: '',
    moegliche: [],
  }) as unknown as Vorgangsstand

function zeige(
  schritt: () => Promise<Handelsvorgangantwort>,
  eroeffnet: () => Promise<Handelsvorgangantwort> = () =>
    Promise.resolve({ vorgang: einStand('entwurf'), fehler: null }),
) {
  const eroeffnen = vi.fn(eroeffnet)
  const schreiten = vi.fn(schritt)
  belegeApi(
    mitStandnachlauf({
      Peeridentitaet: () =>
        Promise.resolve({ kennung: 'ich-1', fingerabdruck: '', fehler: null }),
      Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
      HandelSichtAnfragen: () => Promise.resolve({ zeilen: [], fehler: null }),
      Tauschabgleich: () => Promise.resolve(einAbgleich()),
      HandelPreisanhalt: () =>
        Promise.resolve({
          anhalt: {
            zeilen: [],
            seiten: [],
            differenzen: [],
            hauptwaehrung: '',
          },
          fehler: null,
        }),
      HandelEroeffnen: eroeffnen,
      HandelSchritt: schreiten,
      HandelSchicken: () =>
        Promise.resolve({ nachrichtId: 'n-1', fehler: null }),
      HandelZustellen: () =>
        Promise.resolve({ gegenueber: 'du-1', getragen: 1, fehler: null }),
    } as never),
  )
  const geaendert = vi.fn()
  render(<Buehne geaendert={geaendert} />)

  return { eroeffnen, schreiten, geaendert }
}

describe('ein neuer Tausch', () => {
  it('steht als Tisch da, und die Liste bleibt gemerkt erreichbar', async () => {
    zeige(() =>
      Promise.resolve({ vorgang: einStand('angeboten'), fehler: null }),
    )
    const hand = userEvent.setup()

    expect(
      await screen.findByRole('region', { name: texte.markttisch.meinFach }),
    ).toBeVisible()

    await hand.click(
      screen.getByRole('button', { name: texte.markttisch.alsListe }),
    )
    expect(
      screen.queryByRole('region', { name: texte.markttisch.meinFach }),
    ).toBeNull()
    expect(
      screen.getByRole('button', { name: texte.markttisch.alsListe }),
    ).toHaveAttribute('aria-pressed', 'true')

    // **Gemerkt**: ein neuer Tausch beginnt in der gewählten Ansicht.
    cleanup()
    zeige(() =>
      Promise.resolve({ vorgang: einStand('angeboten'), fehler: null }),
    )
    expect(
      await screen.findByRole('button', { name: texte.markttisch.alsListe }),
    ).toHaveAttribute('aria-pressed', 'true')
    expect(
      screen.queryByRole('region', { name: texte.markttisch.meinFach }),
    ).toBeNull()
  })

  it('eröffnet und bietet in einem Schritt an', async () => {
    const { eroeffnen, schreiten, geaendert } = zeige(() =>
      Promise.resolve({ vorgang: einStand('angeboten'), fehler: null }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zug.anbietenZustimmen }),
    )

    await vi.waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(
        expect.objectContaining({ zustand: 'angeboten' }),
      )
    })
    expect(eroeffnen).toHaveBeenCalledTimes(1)
    // **Mit dem Abdruck des eben eröffneten Entwurfs** (#2496).
    expect(schreiten).toHaveBeenCalledWith('v-1', 'anbieten', '', 'a-entwurf')
  })

  it('stimmt beim Anbieten zu und bringt das Angebot hinüber', async () => {
    const angeboten = {
      ...einStand('angeboten'),
      parteien: ['ich-1', 'du-1'],
      beschluss: 'zustimmen',
    }
    const { schreiten, geaendert } = zeige(((
      _vorgang: string,
      schritt: string,
    ) =>
      Promise.resolve({
        vorgang:
          schritt === 'anbieten'
            ? { ...angeboten, lage: 'dran', naechster: 'zustimmen' }
            : { ...angeboten, lage: 'wartet', zustimmungen: ['ich-1'] },
        fehler: null,
      })) as never)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zug.anbietenZustimmen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.zug.hinaus, { name: 'Bert' }),
    )
    expect(
      schreiten.mock.calls.map((ruf) => (ruf as unknown as string[])[1]),
    ).toEqual(['anbieten', 'zustimmen'])
    expect(geaendert).toHaveBeenLastCalledWith(
      expect.objectContaining({ zustimmungen: ['ich-1'] }),
    )
  })

  it('behält den Entwurf, wenn nur das Anbieten scheitert', async () => {
    const { geaendert } = zeige(() =>
      Promise.resolve({
        vorgang: einStand('entwurf'),
        fehler: {
          code: 'konflikt',
          meldung: 'Das Gegenüber ist nicht zu erreichen.',
          schluessel: '',
          feld: '',
        },
      }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zug.anbietenZustimmen }),
    )

    // **Geschrieben ist geschrieben** (§7, Regel 3): der Entwurf steht da.
    await vi.waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(
        expect.objectContaining({ zustand: 'entwurf' }),
      )
    })
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das Gegenüber ist nicht zu erreichen.',
    )
  })

  it('bietet nichts an, wenn schon das Eröffnen scheitert', async () => {
    const { schreiten, geaendert } = zeige(
      () => Promise.resolve({ vorgang: einStand('angeboten'), fehler: null }),
      () =>
        Promise.resolve({
          vorgang: einStand('entwurf'),
          fehler: {
            code: 'validierung',
            meldung: 'Ein Tausch braucht eine Zeile.',
            schluessel: '',
            feld: '',
          },
        }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.zug.anbietenZustimmen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Ein Tausch braucht eine Zeile.',
    )
    expect(schreiten).not.toHaveBeenCalled()
    expect(geaendert).not.toHaveBeenCalled()
  })

  it('kehrt zum Tisch zurück und kommt ohne Speicher aus', async () => {
    const lesen = vi
      .spyOn(Storage.prototype, 'getItem')
      .mockImplementation(() => {
        throw new Error('kein Speicher')
      })
    const schreiben = vi
      .spyOn(Storage.prototype, 'setItem')
      .mockImplementation(() => {
        throw new Error('kein Speicher')
      })
    zeige(() =>
      Promise.resolve({ vorgang: einStand('angeboten'), fehler: null }),
    )
    const hand = userEvent.setup()

    await hand.click(
      await screen.findByRole('button', { name: texte.markttisch.alsListe }),
    )
    await hand.click(
      screen.getByRole('button', { name: texte.markttisch.alsTisch }),
    )
    expect(
      await screen.findByRole('region', { name: texte.markttisch.meinFach }),
    ).toBeVisible()

    lesen.mockRestore()
    schreiben.mockRestore()
  })
})

describe('ein geeinigter Tausch', () => {
  it('zeigt die Zeilen ohne Tisch und ohne Wahl', () => {
    belegeApi(
      mitStandnachlauf({
        HandelPreisanhalt: () =>
          Promise.resolve({
            anhalt: {
              zeilen: [],
              seiten: [],
              differenzen: [],
              hauptwaehrung: '',
            },
            fehler: null,
          }),
      }),
    )

    render(
      <Sprachraum sprache="de">
        <Anschlussraum
          aktuell="handeln"
          wechseln={() => {
            /* hier wird nicht gewechselt */
          }}
        >
          <Handelszusammenstellung
            ich="ich-1"
            kennung="du-1"
            name="Bert"
            positionen={[]}
            setzePositionen={() => {
              /* nur lesen */
            }}
            bearbeitbar={false}
          />
        </Anschlussraum>
      </Sprachraum>,
    )

    expect(
      screen.queryByRole('button', { name: texte.markttisch.alsListe }),
    ).toBeNull()
    expect(
      screen.queryByRole('region', { name: texte.markttisch.meinFach }),
    ).toBeNull()
  })
})
