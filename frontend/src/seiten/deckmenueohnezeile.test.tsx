// **Das Menü „Deck“ steht, solange ein Deck offen ist** — Rest R6 aus dem
// Nachlauf über Phase 6 von 0.9.0
// (#2469).
// Liste und Zustand hingen an der Zeile im Navigator: traf die Suche das
// offene Deck nicht oder war der Navigator zu, fehlte das Menü, und jede
// Handlung daraus hatte keinen Ort.
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { ereignisObjekthandlung } from '../fenster'
import type { Objekthandlungen } from '../objekthandlungen'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { Decks } from './Decks'

const decktexte = kataloge.de.seiten.decks
const platz = kataloge.de.bausteine.arbeitsplatz

let laufzeit: Laufzeitprobe

beforeEach(() => {
  laufzeit = laufzeitprobe()
})

// **Erst abbauen, dann die Fassade räumen**: beim Abbau verwirft `Teilen`
// die vorbereitete Datei, und das braucht noch ihre Methode.
afterEach(() => {
  cleanup()
  laufzeit.raeume()
  raeumeApi()
})

async function zeige() {
  const melden = vi.fn<
    (meldung: Objekthandlungen) => Promise<{ fehler: null }>
  >(() => Promise.resolve({ fehler: null }))
  const vorbereiten = vi.fn(() =>
    Promise.resolve({
      teilgut: {
        kennung: 't1',
        art: 'deck' as const,
        titel: 'Aggro',
        dateiname: 'Aggro.txt',
        groesse: 12,
        text: '',
        mitgehendes: [],
        verluste: [],
        zuGrossFuerMail: false,
      },
      fehler: null,
    }),
  )
  belegeApi({
    ObjekthandlungenMelden: melden,
    Decks: () =>
      Promise.resolve({
        decks: [deckprobe(), deckprobe({ id: '2', name: 'Kontrolle' })],
        fehler: null,
      }),
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    TeilenVorbereiten: vorbereiten,
    TeilenVerwerfen: () => Promise.resolve({ teilgut: null, fehler: null }),
  })
  render(
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Decks wechseln={vi.fn()} />
      </Objekthandlungsraum>
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: `${decktexte.waehlen}: Aggro` }),
  )

  return { melden, vorbereiten }
}

// deckmeldung ist die letzte Meldung des Decks an die Menüleiste.
function deckmeldung(melden: ReturnType<typeof vi.fn>) {
  return (melden.mock.calls as [Objekthandlungen][])
    .map(([meldung]) => meldung)
    .filter((meldung) => meldung.art === 'deck')
    .at(-1)
}

// waehle wählt einen Eintrag des Menüs „Deck“ in der Menüleiste.
async function waehle(melden: ReturnType<typeof vi.fn>, schluessel: string) {
  const meldung = await waitFor(() => {
    const zuletzt = deckmeldung(melden)
    if (zuletzt === undefined || zuletzt.eintraege.length === 0) {
      throw new Error('das Menü „Deck“ ist nicht gemeldet')
    }

    return zuletzt
  })

  act(() => {
    laufzeit.melde(ereignisObjekthandlung, {
      art: 'deck',
      kennung: meldung.kennung,
      epoche: meldung.epoche,
      schluessel,
    })
  })
}

it('hält das offene Deck in der Liste, auch wenn die Suche es nicht trifft', async () => {
  const { melden } = await zeige()

  await userEvent.type(
    screen.getByLabelText(kataloge.de.bausteine.listenansicht.suche),
    'Kontrolle',
  )

  expect(screen.getByText(decktexte.offenMarke)).toBeVisible()
  await waitFor(() => {
    expect(deckmeldung(melden)?.titel).toBe(decktexte.deckmenue)
  })
})

it('meldet das Menü bei eingeklapptem Navigator, und „Umbenennen“ klappt ihn auf', async () => {
  const { melden } = await zeige()

  await userEvent.click(screen.getByRole('button', { name: platz.navigatorZu }))
  await waitFor(() => {
    expect(screen.queryByText(decktexte.offenMarke)).not.toBeInTheDocument()
  })

  await waehle(melden, 'umbenennen')

  expect(await screen.findByLabelText(decktexte.neuerName)).toHaveValue('Aggro')
})

it('teilt aus der Menüleiste auch, wenn der Navigator erst aufgehen muss', async () => {
  const { melden, vorbereiten } = await zeige()

  await userEvent.click(screen.getByRole('button', { name: platz.navigatorZu }))
  await waehle(melden, 'teilen')

  expect(
    await screen.findByRole('heading', {
      name: kataloge.de.bausteine.teilen.ueberschrift,
    }),
  ).toBeVisible()
  expect(vorbereiten).toHaveBeenCalledTimes(1)
})

it('lässt einen offenen Navigator, wie er ist, und zeigt die Handlung an der Zeile', async () => {
  const { melden } = await zeige()

  await waehle(melden, 'umbenennen')

  expect(await screen.findByLabelText(decktexte.neuerName)).toHaveValue('Aggro')
  expect(screen.getByRole('button', { name: platz.navigatorZu })).toBeVisible()
})
