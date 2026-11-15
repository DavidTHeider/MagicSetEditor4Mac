// **Ein Zug, der unterwegs hängen bleibt** — P13
// (#2381),
// Abnahmekriterium 2 am Knopf: jede Stelle hat ihren eigenen Satz und, wo es
// einen gibt, ihren eigenen Ausweg — und was gegangen ist, bleibt.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { einDienst, einFehler, ich } from '../test/handelszugprobe'
import { folge, vorbereitet, zeige } from '../test/zugbuehne'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de.seiten.handeln
const zug = texte.zug

// mitGlut ist der vorbereitete Tausch, in dem Anna den Glutläufer gibt —
// ohne Absprache über Zustand, Sprache und Ausführung.
const mitGlut = {
  ...vorbereitet,
  positionen: [
    {
      von: ich,
      printing: 'p-glut',
      druckId: 0,
      menge: 1,
      guete: '',
      sprache: '',
      finish: '',
      abbild: '',
      bewertung: 0,
    },
  ],
}

afterEach(() => {
  raeumeApi()
})

describe('ein Zug, der unterwegs hängen bleibt', () => {
  it('nennt den Schritt, der nicht ging', async () => {
    belegeApi(
      einDienst(folge, { verbuchen: 'Der Bestand reicht nicht.' }, vorbereitet)
        .api as never,
    )
    const { geaendert, gebucht } = zeige(vorbereitet)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(zug.gescheitert, {
        schritt: texte.buchungVerbuchen,
        meldung: 'Der Bestand reicht nicht.',
      }),
    )
    // **Die Übergabe bleibt** — und der Raum trägt ihren Stand ein.
    expect(geaendert).toHaveBeenCalledWith(folge.uebergeben)
    expect(gebucht).not.toHaveBeenCalled()
  })

  it('sagt, wo das Festhalten an einem Menschen hängt', async () => {
    belegeApi({
      ...einDienst(folge, {}, mitGlut).api,
      Handelszusagen: () =>
        Promise.resolve({ vorgang: 'v-1', zusagen: [], fehler: null }),
      SammlungBesitz: () => Promise.resolve({ besitz: null, fehler: null }),
    } as never)
    const { geaendert, gebucht } = zeige(mitGlut)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      zug.festhaltenOffen,
    )
    expect(geaendert).not.toHaveBeenCalled()
    expect(gebucht).not.toHaveBeenCalled()
  })

  it('bringt hinaus, was im Ausgang liegt, wenn das Übertragen erst nicht ging', async () => {
    const zuerst = einDienst(
      folge,
      {
        zustellen: 'Bert ist nicht zu erreichen.',
      },
      vorbereitet,
    )
    belegeApi(zuerst.api as never)
    zeige(vorbereitet)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(zug.unterwegs, {
        name: 'Bert',
        meldung: 'Bert ist nicht zu erreichen.',
      }),
    )

    const dann = einDienst(folge, {}, vorbereitet)
    belegeApi(dann.api as never)
    await hand.click(
      screen.getByRole('button', { name: texte.ausgangZustellen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(zug.hinaus, { name: 'Bert' }),
    )
    expect(screen.queryByRole('alert')).toBeNull()
    expect(dann.rufe).toEqual(['zustellen'])
  })

  it('legt in den Ausgang, was nicht hineinkam, und bringt es hinaus', async () => {
    const zuerst = einDienst(
      folge,
      {
        ausgang: 'Der Schlüsselbund hat abgelehnt.',
      },
      vorbereitet,
    )
    belegeApi(zuerst.api as never)
    zeige(vorbereitet)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )
    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(zug.ausgang, { meldung: 'Der Schlüsselbund hat abgelehnt.' }),
    )

    const dann = einDienst(folge, {}, vorbereitet)
    belegeApi(dann.api as never)
    await hand.click(
      screen.getByRole('button', { name: texte.ausgangSchicken }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(zug.hinaus, { name: 'Bert' }),
    )
    expect(dann.rufe).toEqual(['ausgang', 'zustellen'])
  })

  it('nennt es, wenn auch der zweite Versuch nicht hinausgeht', async () => {
    belegeApi(
      einDienst(
        folge,
        { zustellen: 'Bert ist nicht zu erreichen.' },
        vorbereitet,
      ).api as never,
    )
    zeige(vorbereitet)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )
    await hand.click(
      await screen.findByRole('button', { name: texte.ausgangZustellen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Bert ist nicht zu erreichen.',
    )
    expect(
      screen.getByRole('button', { name: texte.ausgangZustellen }),
    ).toBeVisible()
  })

  it('sagt, wenn der Stand danach nicht zu lesen war', async () => {
    belegeApi(
      einDienst(
        folge,
        { 'stand-danach': 'Der Bestand ist belegt.' },
        vorbereitet,
      ).api as never,
    )
    const { gebucht } = zeige(vorbereitet)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(zug.lesen, { meldung: 'Der Bestand ist belegt.' }),
    )
    expect(gebucht).toHaveBeenCalledTimes(1)
  })
})

describe('ein Zug, dessen Ausweg auch hängt', () => {
  it('nennt das Festhalten, wenn es selbst scheitert', async () => {
    belegeApi({
      ...einDienst(folge, {}, mitGlut).api,
      Handelszusagen: () =>
        Promise.resolve({
          vorgang: 'v-1',
          zusagen: [],
          fehler: einFehler('Die Zusagen sind gerade nicht zu lesen.'),
        }),
    } as never)
    zeige(mitGlut)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(zug.gescheitert, {
        schritt: texte.zusagenFesthalten,
        meldung: 'Die Zusagen sind gerade nicht zu lesen.',
      }),
    )
  })

  it('bleibt beim Ausgang, wenn auch das zweite Legen nicht geht', async () => {
    belegeApi(
      einDienst(
        folge,
        { ausgang: 'Der Schlüsselbund hat abgelehnt.' },
        vorbereitet,
      ).api as never,
    )
    zeige(vorbereitet)
    const hand = userEvent.setup()

    await hand.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )
    await hand.click(
      await screen.findByRole('button', { name: texte.ausgangSchicken }),
    )

    // **Derselbe Ausweg steht wieder da** — und kein Übertragen, denn im
    // Ausgang liegt nichts.
    expect(
      await screen.findByRole('button', { name: texte.ausgangSchicken }),
    ).toBeVisible()
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Der Schlüsselbund hat abgelehnt.',
    )
    expect(
      screen.queryByRole('button', { name: texte.ausgangZustellen }),
    ).toBeNull()
  })

  it('meldet einen Nebenweg, der nicht ging, wie bisher am Schritt', async () => {
    const { geaendert } = zeige(vorbereitet)
    belegeApi(
      einDienst(
        {},
        { abbrechen: 'Abbrechen geht hier nicht mehr.' },
        vorbereitet,
      ).api as never,
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.schrittAbbrechen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Abbrechen geht hier nicht mehr.',
    )
    expect(geaendert).not.toHaveBeenCalled()
  })
})

describe('ein Zug von einem Stand, den es nicht mehr gibt', () => {
  it('geht nichts, sagt es und zeigt den neuen Stand', async () => {
    // **Befund 1 des Zweitblicks über Phase 8**: zwischen Anzeige und Klick
    // kam ein Gegenangebot.
    const neu = { ...vorbereitet, revision: 2, abdruck: 'neu' }
    const dienst = einDienst(folge, {}, neu)
    belegeApi(dienst.api as never)
    const { geaendert, gebucht } = zeige(vorbereitet)

    await userEvent.click(
      screen.getByRole('button', { name: zug.uebergebenVerbuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(zug.veraltet)
    expect(geaendert).toHaveBeenCalledWith(neu)
    expect(gebucht).not.toHaveBeenCalled()
    expect(dienst.rufe).toEqual(['stand'])
  })
})
