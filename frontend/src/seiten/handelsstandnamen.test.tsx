// **Der Stand eines Handels nennt Menschen und keine Kennungen** — UXR-15
// (#1923),
// nachgeschärft beim Abschluss von Phase 6.8.
//
// # Was vorher war
//
// **Unter „Was tauscht ihr" standen fünf Zeilen Stand**: Zustand, Revision
// und Abdruck in einer, die Parteien als rohe Peerkennungen („bafzaaja…"),
// die Art, die Stimmen — ebenfalls als Kennungen —, und daneben „Stand
// ansehen", das neben Lage und Spalten nichts mehr zeigte, was nicht schon
// dastand. Die Lage sagte „Wartet auf dein Gegenüber", obwohl der Name
// bekannt war, und der Gegenbeleg nannte den, der gezeichnet hatte, mit
// seiner Kennung.
//
// # Was jetzt gilt
//
// **Vorn steht, was ein Mensch liest**: Zustand, Art und, wer zugestimmt hat
// — „du" und der Name des Gegenübers. **Revision, Abdruck, die Parteien mit
// ihrer Kennung und „Stand ansehen" stehen in einem Kasten für Fachleute.**
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { eineUebersicht, eineZeile, einVorgang } from './handelslisteprobe'
import { oeffneHandel, zeigeHandel, zuPhase } from './tauschprobe'
import {
  bestaetigt,
  dieVorgaenge,
  einJournal,
  einStand,
  eineFreigabe,
  imAbschluss,
  trageEin,
  mitBelegnachlauf,
} from './buchungsprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgang } from '../apihandelsleitung'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

// derAusweis sagt, wer dieses Programm ist.
const derAusweis = () =>
  Promise.resolve({
    stand: 'bereit',
    kennung: 'bich',
    fingerabdruck: 'AB-CD',
    seit: '2026-09-19T08:00:00Z',
    vorher: '',
    vorherFingerabdruck: '',
    befunde: [],
    fehler: null,
  })

// mitHandel stellt einen Handel mit dem Gegenüber der Kontaktprobe bereit.
function mitHandel(vorgang: Handelsvorgang) {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Peeridentitaet: derAusweis,
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
      HandelUebersicht: () => eineUebersicht([eineZeile({}, vorgang)]),
    }),
  )
}

// imAssistenten sucht einen Satz im Assistenten — **und nicht in der
// Liste**, in der derselbe Satz ohnehin steht.
async function imAssistenten(satz: string) {
  const leiste = await screen.findByRole('navigation', {
    name: eigene.assistent.schrittleiste,
  })
  const assistent = leiste.closest('section')
  if (assistent === null) {
    throw new Error('der Assistent fehlt')
  }

  return within(assistent).findByText(satz)
}

// standImRaum gibt den Kasten „Was tauscht ihr" des Handels `h-1`.
async function standImRaum(hand: Awaited<ReturnType<typeof zeigeHandel>>) {
  await oeffneHandel(hand, 'h-1')
  await zuPhase(hand, 'angebot')

  return screen.findByRole('region', { name: eigene.vorgangTitel })
}

it('nennt das Gegenüber beim Namen, auf das gewartet wird', async () => {
  mitHandel(einVorgang({ lage: 'wartet', naechster: '', moegliche: [] }))

  const hand = await zeigeHandel()
  await standImRaum(hand)

  expect(
    await screen.findByText('Wartet auf Fremder Rechner'),
  ).toBeInTheDocument()
})

it('nimmt den Namen aus der Liste, wenn der Handel dort aufgeschlagen wird', async () => {
  mitHandel(einVorgang({ lage: 'wartet', naechster: '', moegliche: [] }))

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
  await userEvent
    .setup()
    .click(await screen.findByRole('button', { name: /^Turfer/ }))

  expect(await imAssistenten('Wartet auf Turfer')).toBeInTheDocument()
})

// **Ohne Namen bleibt es „dein Gegenüber"** — in der Liste steht dann der
// Anfang der Kennung, und die gehört nicht in einen Satz.
it('sagt „dein Gegenüber", wo kein Name bekannt ist', async () => {
  const vorgang = einVorgang({ lage: 'wartet', naechster: '', moegliche: [] })
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: () =>
        Promise.resolve({ vorgaenge: [vorgang], fehler: null }),
      HandelUebersicht: () =>
        eineUebersicht([
          eineZeile({ gegenueber: [{ kennung: 'bgegenueber', name: '' }] }),
        ]),
    }),
  )

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
  await userEvent
    .setup()
    .click(await screen.findByRole('button', { name: /^bgegenueber/ }))

  expect(await imAssistenten('Wartet auf dein Gegenüber')).toBeInTheDocument()
})

it('nennt in den Stimmen dich und das Gegenüber beim Namen', async () => {
  mitHandel(
    einVorgang({
      vorgang: 'h-1',
      beschluss: 'zustimmen',
      zustimmungen: ['bich', 'bgegenueber'],
    }),
  )

  const raum = await standImRaum(await zeigeHandel())

  expect(
    await within(raum).findByText(
      `${eigene.vorgangZustimmungen}: ${eigene.vorgangDu}, Fremder Rechner`,
    ),
  ).toBeInTheDocument()
  expect(within(raum).queryByText(/bgegenueber/)).not.toBeInTheDocument()
})

it('stellt Revision, Abdruck, Parteien und „Stand ansehen" unter Fachleute', async () => {
  mitHandel(einVorgang({ vorgang: 'h-1', revision: 3 }))

  const hand = await zeigeHandel()
  const raum = await standImRaum(hand)

  await within(raum).findByText(
    `${eigene.buchungZustand}: ${eigene.buchungStandAngeboten}`,
  )
  expect(
    within(raum).queryByText(new RegExp(`${eigene.vorgangRevision} 3`)),
  ).not.toBeInTheDocument()
  expect(
    within(raum).queryByRole('button', { name: eigene.vorgangAnsehen }),
  ).not.toBeInTheDocument()

  await hand.click(
    within(raum).getByRole('button', { name: eigene.vorgangFachleute }),
  )

  expect(
    within(raum).getByText(
      `${eigene.vorgangRevision} 3 · ${eigene.vorgangAbdruck} abcdef012345`,
    ),
  ).toBeInTheDocument()
  expect(
    within(raum).getByText(`${eigene.vorgangParteien}: bich, bgegenueber`),
  ).toBeInTheDocument()
  expect(
    within(raum).getByRole('button', { name: eigene.vorgangAnsehen }),
  ).toBeInTheDocument()
})

it('nennt den, der gegengezeichnet hat, beim Namen', async () => {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(imAbschluss()),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsverbuchung: () =>
        Promise.resolve(
          einStand({
            zustand: 'beidseitigVerbucht',
            beleg: {
              vorgang: 'v-1',
              abdruck: 'abdruck-1',
              unterschrift: 'deadbeef',
              beidseitig: true,
              gegenzeichner: 'bgegenueber',
              gegengezeichnet: '2026-09-19T12:00:00Z',
              gebucht: '2026-09-18T12:00:00Z',
              befund: null,
            },
          }),
        ),
      Handelsjournal: () => Promise.resolve(einJournal()),
    }),
  )

  await trageEin(await zeigeHandel(), eigene.buchungVerbuchen)

  expect(
    await screen.findByText(/Geprüfter Beleg von Fremder Rechner/),
  ).toBeInTheDocument()
})
