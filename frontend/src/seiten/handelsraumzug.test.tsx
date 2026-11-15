// **Der Zug gehört dem Raum** — P13
// (#2381),
// Vermerk 0200: seine Meldung steht unter den Schritten, ihr Ausweg geht
// dort, und wer einen anderen Handel wählt, sieht sie nicht mehr (§7,
// Regel 1).
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Handeln } from './Handeln'
import { einVorgang } from './handelslisteprobe'
import { eineListe } from './peerkontaktprobe'
import { bestaetigt, oeffneHandel, waehleGegenueber } from './tauschprobe'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einDienst } from '../test/handelszugprobe'

const eigene = kataloge.de.seiten.handeln
const zug = eigene.zug
const name = 'Fremder Rechner'

afterEach(() => {
  raeumeApi()
})

const vorbereitet = einVorgang({
  zustand: 'vorbereitet',
  beschluss: 'uebergeben',
  zustimmungen: ['bgegenueber'],
  lage: 'dran',
  naechster: 'uebergeben',
  moegliche: ['uebergeben', 'abbrechen'],
  phase: 'zusagen',
})
const anderer = einVorgang({
  vorgang: 'h-2',
  zustimmungen: ['bich'],
  lage: 'wartet',
  naechster: '',
  moegliche: ['abbrechen'],
})
const folge = {
  uebergeben: einVorgang({
    zustand: 'uebergabeOffen',
    naechster: 'verbuchen',
    moegliche: ['verbuchen', 'klaeren'],
    phase: 'verbuchen',
  }),
  verbuchen: einVorgang({
    zustand: 'lokalVerbucht',
    naechster: 'abgleichen',
    phase: 'verbuchen',
  }),
  abgleichen: einVorgang({
    zustand: 'abgleichOffen',
    beschluss: 'bestaetigen',
    naechster: 'bestaetigen',
    phase: 'verbuchen',
  }),
  'bestaetigen:bich': einVorgang({
    zustand: 'abgleichOffen',
    beschluss: 'bestaetigen',
    zustimmungen: ['bich'],
    lage: 'wartet',
    naechster: '',
    phase: 'verbuchen',
  }),
}

// stelle belegt die Seite samt einem Dienst, der den Zug beantwortet.
function stelle(dienst: ReturnType<typeof einDienst>) {
  belegeApi({
    ...dienst.api,
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [vorbereitet, anderer], fehler: null }),
  } as never)
}

it('zeigt unter den Schritten, wo der Zug hängt, und bringt es von dort hinaus', async () => {
  stelle(
    einDienst(
      folge,
      { zustellen: `${name} ist nicht zu erreichen.` },
      vorbereitet,
    ),
  )
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
  const hand = userEvent.setup()

  await waehleGegenueber(hand, name)
  await oeffneHandel(hand, 'h-1')
  await hand.click(
    await screen.findByRole('button', { name: zug.uebergebenVerbuchen }),
  )

  expect(
    await screen.findByText(
      einsetzen(zug.unterwegs, {
        name,
        meldung: `${name} ist nicht zu erreichen.`,
      }),
    ),
  ).toBeVisible()

  const dann = einDienst(folge)
  stelle(dann)
  await hand.click(
    screen.getByRole('button', { name: eigene.ausgangZustellen }),
  )
  expect(await screen.findByText(einsetzen(zug.hinaus, { name }))).toBeVisible()
  expect(dann.rufe).toEqual(['zustellen'])

  // **Ein anderer Handel, und die Meldung gehört nicht zu ihm.**
  // **Erst der Beleg, dass gewechselt ist**: der Zug von h-1 steht nicht
  // mehr da.
  await oeffneHandel(hand, 'h-2')
  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: zug.uebergebenVerbuchen }),
    ).toBeNull()
  })
  expect(screen.queryByText(einsetzen(zug.hinaus, { name }))).toBeNull()
})

it('zeigt keine Meldung eines Zugs, der nach einem Wechsel ankommt', async () => {
  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
  const dienst = einDienst(folge, {}, vorbereitet)
  let weiter: () => void = () => undefined
  const gehalten = new Promise<void>((los) => {
    weiter = los
  })
  belegeApi({
    ...dienst.api,
    HandelSchritt: async (
      vorgang: string,
      schritt: string,
      wer: string,
      gesehen: string,
    ) => {
      await gehalten

      return dienst.api.HandelSchritt(vorgang, schritt, wer, gesehen)
    },
    Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
    HandelVorgaenge: () =>
      Promise.resolve({ vorgaenge: [vorbereitet, anderer], fehler: null }),
  } as never)
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
  const hand = userEvent.setup()

  await waehleGegenueber(hand, name)
  await oeffneHandel(hand, 'h-1')
  await hand.click(
    await screen.findByRole('button', { name: zug.uebergebenVerbuchen }),
  )
  await oeffneHandel(hand, 'h-2')
  weiter()

  // **Erst der Beleg, dass der Zug zu Ende ging**: er hat übertragen.
  await waitFor(() => {
    expect(dienst.rufe).toContain('zustellen')
  })
  await act(async () => {
    await Promise.resolve()
  })
  expect(screen.queryByText(einsetzen(zug.hinaus, { name }))).toBeNull()
})

it('zeigt auch die Antwort von „Jetzt übertragen“ nach einem Wechsel nicht', async () => {
  stelle(
    einDienst(
      folge,
      { zustellen: `${name} ist nicht zu erreichen.` },
      vorbereitet,
    ),
  )
  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
  const hand = userEvent.setup()

  await waehleGegenueber(hand, name)
  await oeffneHandel(hand, 'h-1')
  await hand.click(
    await screen.findByRole('button', { name: zug.uebergebenVerbuchen }),
  )
  const nochmal = await screen.findByRole('button', {
    name: eigene.ausgangZustellen,
  })

  const dann = einDienst(folge)
  let weiter: () => void = () => undefined
  const gehalten = new Promise<void>((los) => {
    weiter = los
  })
  stelle({
    ...dann,
    api: {
      ...dann.api,
      HandelZustellen: async (kennung: string) => {
        await gehalten

        return dann.api.HandelZustellen(kennung)
      },
    },
  })
  await hand.click(nochmal)
  await oeffneHandel(hand, 'h-2')
  weiter()

  await waitFor(() => {
    expect(dann.rufe).toContain('zustellen')
  })
  await act(async () => {
    await Promise.resolve()
  })
  expect(screen.queryByText(einsetzen(zug.hinaus, { name }))).toBeNull()
})
