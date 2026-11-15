// **„Neuer Tausch …" und die Schublade dahinter** — UXR-15
// (#1923).
//
// Die Wahl des Gegenübers, das Aufnehmen und die Kontakte standen offen in
// der Spalte links, unter der Liste der Handel. Jetzt führt „Neuer Tausch …"
// die Spalte, und alle drei stehen in der Schublade dahinter — **erreichbar
// bleibt jedes**, nur einen Klick weiter.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineUebersicht } from './handelslisteprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { zehnGegenueber } from '../test/handelsbestand'

const texte = kataloge.de
const eigene = texte.seiten.handeln

// zeige stellt *Handeln* ohne einen Handel auf.
function zeige() {
  const Peerkontakte = vi.fn(() => Promise.resolve(zehnGegenueber()))

  belegeApi({
    Peerkontakte,
    HandelUebersicht: () => eineUebersicht([]),
    Kontakte: () => Promise.resolve({ kontakte: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )

  return { hand: userEvent.setup(), Peerkontakte }
}

// oeffne drückt „Neuer Tausch …" und gibt die Schublade.
async function oeffne(hand: ReturnType<typeof userEvent.setup>) {
  await hand.click(
    await screen.findByRole('button', { name: eigene.neuerTausch }),
  )

  return screen.findByRole('dialog', { name: eigene.neuerTauschTitel })
}

describe('Neuer Tausch', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('führt die Spalte und nennt den Weg, solange nichts läuft', async () => {
    zeige()

    expect(
      await screen.findByRole('button', { name: eigene.neuerTausch }),
    ).toHaveClass('bg-marke')
    expect(await screen.findByText(eigene.liste.leer)).toBeInTheDocument()
    expect(eigene.liste.leer).toContain(eigene.neuerTausch)
    expect(eigene.partnerKeiner).toContain(eigene.neuerTausch)
  })

  it('öffnet die Wahl, das Aufnehmen und die Kontakte in einer Schublade', async () => {
    const { hand } = zeige()

    const schublade = await oeffne(hand)

    expect(
      within(schublade).getByRole('heading', { name: eigene.partnerTitel }),
    ).toBeInTheDocument()
    expect(
      within(schublade).getByRole('button', {
        name: texte.bausteine.gegenueberTitel,
      }),
    ).toBeInTheDocument()
    expect(
      within(schublade).getByRole('button', {
        name: texte.seiten.adressbuch.titel,
      }),
    ).toBeInTheDocument()
    expect(
      await within(schublade).findByRole('button', {
        name: einsetzen(eigene.partnerWaehlenName, { name: 'Ines' }),
      }),
    ).toBeInTheDocument()
  })

  it('schlägt mit der Wahl einen neuen Handel auf und schließt die Schublade', async () => {
    const { hand } = zeige()

    const schublade = await oeffne(hand)
    await hand.click(
      await within(schublade).findByRole('button', {
        name: einsetzen(eigene.partnerWaehlenName, { name: 'Anna' }),
      }),
    )

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
    expect(await screen.findByText(eigene.assistent.neu)).toBeInTheDocument()
    expect(screen.queryByText(eigene.partnerKeiner)).not.toBeInTheDocument()

    // **Beim nächsten Öffnen nennt der Knopf, was er tut** (#2391): eine
    // Wahl beginnt immer einen neuen Entwurf. Hier stand das gewählte
    // gedrückt da, als »Aufgeschlagen« — und ein zweiter Druck blieb im
    // Handel, den der Raum inzwischen gewählt hatte.
    const wieder = await oeffne(hand)
    expect(
      within(wieder).getByRole('button', {
        name: einsetzen(eigene.partnerWaehlenName, { name: 'Anna' }),
      }),
    ).toHaveAttribute('aria-pressed', 'false')
  })

  // **Gezeichnet beim ersten Öffnen, danach nur verborgen** — wie die
  // Schubladen der Sammlung: was darin aufgeklappt war, steht beim nächsten
  // Öffnen noch so da.
  it('behält beim Schließen, was in ihr aufgeklappt war', async () => {
    const { hand } = zeige()

    const schublade = await oeffne(hand)
    const kontakte = within(schublade).getByRole('button', {
      name: texte.seiten.adressbuch.titel,
    })
    await hand.click(kontakte)
    expect(kontakte).toHaveAttribute('aria-expanded', 'true')

    await hand.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    const wieder = await oeffne(hand)
    expect(
      within(wieder).getByRole('button', {
        name: texte.seiten.adressbuch.titel,
      }),
    ).toHaveAttribute('aria-expanded', 'true')
  })
})

// **Die Seite darunter ist gesperrt, und der Fokus kehrt zurück** — wie bei
// den Schubladen von Sammlung und Paketen: `inert` nimmt Fokus, Klick und
// Vorleseprogramm, und nach dem Schließen steht der Fokus wieder auf dem
// Knopf, der sie geöffnet hat (in WebKit gibt ein Klick keinen Fokus).
describe('Neuer Tausch über der Seite', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sperrt die Seite, solange die Schublade steht, und gibt den Fokus zurück', async () => {
    const { hand } = zeige()
    const kopf = await screen.findByRole('heading', {
      level: 1,
      name: eigene.titel,
    })

    expect(kopf.closest('[inert]')).toBeNull()

    await oeffne(hand)
    expect(kopf.closest('[inert]')).not.toBeNull()

    await hand.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    expect(kopf.closest('[inert]')).toBeNull()
    expect(
      screen.getByRole('button', { name: eigene.neuerTausch }),
    ).toHaveFocus()
  })
})
