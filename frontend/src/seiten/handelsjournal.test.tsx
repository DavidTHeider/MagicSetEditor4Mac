// **Das Journal eines Handels, getrennt vom laufenden Vorgang** — UXR-15
// (#1923).
//
// # Was vorher war
//
// **Das Journal stand nur nach einem Knopfdruck da**: wer „Verbuchen",
// „Abgleich melden" oder „Gegenseite bestätigt" drückte, sah darunter, was
// sich bewegt hatte. Wer nur nachsehen wollte, musste einen Schritt gehen —
// und jede Bewegung stand untereinander, bei einem Tausch gegen eine
// Sammelmappe achtzig Zeilen (R4 in `docs/UIREVIEW.md`).
//
// # Was jetzt gilt
//
// **Ein eigener Kasten in „Verbuchen"**, der beim Öffnen fragt und nach jedem
// Schritt der Buchung nachlädt; zehn Bewegungen vorn, der Rest stückweise.
// Ein Lesefehler dort nimmt keine Buchung zurück (§7, Regel 3).
import { screen, within } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zeigeHandel, zuPhase } from './tauschprobe'
import {
  bestaetigt,
  buchung,
  dieVorgaenge,
  einJournal,
  einStand,
  eineBewegung,
  eineFreigabe,
  imAbschluss,
  mitBelegnachlauf,
  standSagt,
} from './buchungsprobe'
import { einLangesJournal } from '../test/handelsbestand'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
  vi.restoreAllMocks()
})

// mitJournal belegt den Raum mit einem Handel nach der Übergabe.
function mitJournal(Handelsjournal: Api['Handelsjournal']) {
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      HandelVorgaenge: dieVorgaenge(imAbschluss()),
      HandelFreigabeErteilen: eineFreigabe,
      Handelsverbuchung: () => Promise.resolve(einStand()),
      Handelsjournal,
    }),
  )
}

// oeffneJournal klappt den Kasten auf und gibt ihn.
async function oeffneJournal(
  hand: Awaited<ReturnType<typeof zeigeHandel>>,
): Promise<HTMLElement> {
  const schalter = await screen.findByRole('button', {
    name: eigene.buchungJournal,
  })
  await hand.click(schalter)

  // **Der Kasten samt Schalter** — die Zeilen darin und nicht die der
  // Schrittleiste.
  const kasten = schalter.closest('section')
  if (kasten === null) {
    throw new Error('der Kasten des Journals fehlt')
  }

  return kasten
}

it('fragt erst, wenn jemand den Kasten öffnet', async () => {
  const gelesen = vi.fn(() =>
    Promise.resolve(
      einJournal([
        eineBewegung(),
        eineBewegung({
          id: 2,
          art: 'zugang',
          printing: 'shivan-lea',
          // **Ein Druck, den dieser Rechner nicht kennt, behält seine
          // Kennung** — einen Namen zu erfinden wäre schlimmer.
          bezeichnung: 'shivan-lea',
          menge: 1,
          lagerort: '',
        }),
        eineBewegung({
          id: 3,
          art: 'zugang',
          anlass: 'korrektur',
          menge: 2,
          lagerort: '',
          bezug: 1,
          grund: 'verzählt',
        }),
      ]),
    ),
  )
  mitJournal(gelesen)

  const hand = await zeigeHandel()
  await buchung(hand)
  expect(gelesen).not.toHaveBeenCalled()

  const kasten = await oeffneJournal(hand)
  const liste = await within(kasten).findByRole('list', {
    name: eigene.buchungJournal,
  })

  expect(gelesen).toHaveBeenCalledWith('v-1')

  // **Die Richtung steht dran, und der Lagerort auch** — das Journal ist
  // das eigene Buch und nicht der Beleg, der hinausgeht.
  expect(
    within(liste).getByText(/−2 × Lightning Bolt · M10 #146/),
  ).toBeInTheDocument()
  expect(within(liste).getByText(/Kiste 4/)).toBeInTheDocument()
  expect(within(liste).getByText(/\+1 × shivan-lea/)).toBeInTheDocument()

  // **Eine Korrektur nennt ihren Grund** und sonst niemand.
  expect(within(liste).getByText(/verzählt/)).toBeInTheDocument()
})

it('zeigt zehn Bewegungen und den Rest stückweise', async () => {
  mitJournal(() => Promise.resolve(einJournal(einLangesJournal('v-1'))))

  const hand = await zeigeHandel()
  await buchung(hand)
  const kasten = await oeffneJournal(hand)

  await within(kasten).findByText('10 von 80')
  expect(within(kasten).getAllByRole('listitem')).toHaveLength(10)

  await hand.click(
    within(kasten).getByRole('button', { name: 'Weitere 20 zeigen' }),
  )
  expect(within(kasten).getAllByRole('listitem')).toHaveLength(30)
})

it('lädt nach einer Buchung nach, wenn er offen ist', async () => {
  mitJournal(
    vi
      .fn()
      .mockResolvedValueOnce(einJournal())
      .mockResolvedValue(einJournal([eineBewegung()])),
  )

  const hand = await zeigeHandel()
  const kasten = await buchung(hand)
  const journal = await oeffneJournal(hand)

  // **Das leere Journal sagt es**, statt eine leere Liste zu zeigen.
  expect(
    await within(journal).findByText(eigene.buchungJournalLeer),
  ).toBeInTheDocument()

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.buchungVerbuchen }),
  )

  expect(
    await within(journal).findByText(/−2 × Lightning Bolt/),
  ).toBeInTheDocument()
})

// **Geschrieben ist geschrieben** (§7, Regel 3): ein Lesefehler beim
// Nachladen steht im Kasten, nimmt die Buchung aber nicht zurück.
it('meldet einen Lesefehler im Kasten, und die Buchung bleibt', async () => {
  mitJournal(
    vi
      .fn()
      .mockResolvedValueOnce(einJournal())
      .mockResolvedValue({
        ...einJournal(),
        fehler: {
          code: 'unbekannt' as const,
          schluessel: 'journal',
          meldung: 'das Journal klagt',
        },
      }),
  )

  const hand = await zeigeHandel()
  const kasten = await buchung(hand)
  const journal = await oeffneJournal(hand)
  await within(journal).findByText(eigene.buchungJournalLeer)

  await hand.click(
    within(kasten).getByRole('button', { name: eigene.buchungVerbuchen }),
  )

  expect(
    await within(journal).findByText('das Journal klagt'),
  ).toBeInTheDocument()
  await standSagt(eigene.buchungStandLokalVerbucht)
})

it('fragt ohne Handel nicht', async () => {
  const gelesen = vi.fn()
  belegeApi(
    mitBelegnachlauf({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Handelsjournal: gelesen,
    }),
  )

  const hand = await zeigeHandel()
  await zuPhase(hand, 'verbuchen')
  const journal = await oeffneJournal(hand)

  expect(
    await within(journal).findByText(eigene.buchungJournalLeer),
  ).toBeInTheDocument()
  expect(gelesen).not.toHaveBeenCalled()
})
