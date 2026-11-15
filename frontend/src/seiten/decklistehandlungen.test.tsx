// **Die Handlungen eines Decks stehen am aufgeschlagenen** — UXR-11
// (#1919).
//
// Jede Zeile der Deckliste trug sechs Bedienelemente — Öffnen, Umbenennen,
// Duplizieren, Teilen, Löschen und die Marken —, und in einem 240 Punkte
// breiten Navigator brach jede in drei, vier Zeilen um. Bei acht Decks waren
// das 48 Knöpfe, ehe die erste Karte zu sehen war (R2 in `docs/UIREVIEW.md`).
// Jetzt trägt eine Zeile Öffnen und Marken; was man an **einem** Deck tut,
// steht an dem, das offen ist.
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe } from './deckprobe'
import {
  commanderdeck,
  deckformate,
  deckuebersicht,
  standarddeck,
} from '../test/deckbestand'
import { Anschlussraum } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { deckhandlung, deckmenue } from '../test/deckhandlung'

import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.decks
const teilen = kataloge.de.bausteine.teilen

afterEach(() => {
  raeumeApi()
})

function zeige(
  teilen = vi.fn(() => Promise.resolve({ teilgut: null, fehler: null })),
  wechseln: (seite: Seitenschluessel) => void = vi.fn(),
) {
  const decks = new Map([
    [standarddeck.id, standarddeck],
    [commanderdeck.id, commanderdeck],
  ])

  belegeApi({
    Decks: () => Promise.resolve({ decks: deckuebersicht, fehler: null }),
    Formate: () => Promise.resolve({ formate: deckformate, fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: (deckID: string) =>
      Promise.resolve({ deck: decks.get(deckID) ?? null, fehler: null }),
    DeckPruefen: () =>
      Promise.resolve({ urteil: null, befunde: [], fehler: null }),
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    MarkenJeObjekt: () => Promise.resolve({ marken: {}, fehler: null }),
    TeilenVorbereiten: teilen,
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="decks" wechseln={wechseln}>
        <Decks wechseln={vi.fn()} />
      </Anschlussraum>
    </Sprachraum>,
  )
}

// handlungen zählt die Knöpfe, die an einem Deck etwas tun.
function handlungen(): HTMLElement[] {
  const liste = screen.getByRole('region', { name: texte.titel })

  return within(liste)
    .queryAllByRole('button')
    .filter((knopf) =>
      [texte.umbenennen, texte.duplizieren, texte.loeschen, teilen.knopf].some(
        (wort) => knopf.textContent.includes(wort),
      ),
    )
}

describe('die Deckliste mit acht Decks', () => {
  it('trägt je Deck nur Öffnen, solange keines offen ist', async () => {
    zeige()

    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` })

    expect(
      screen.getAllByRole('button', {
        name: new RegExp(`^${texte.waehlen}: `),
      }),
    ).toHaveLength(8)
    expect(handlungen()).toHaveLength(0)
  })

  // **Hinter einem Knopf und im Rechtsklick** — AP-6 (#2363), Befund B-03:
  // vier Knöpfe am offenen Deck waren vier Bedienelemente mehr im Bild, für
  // Handlungen, die man an einem Deck selten tut (`Bereiche09`).
  it('legt Umbenennen, Duplizieren, Teilen und Löschen hinter »…«', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )
    await screen.findByRole('list', { name: texte.plaetze })

    expect(handlungen()).toHaveLength(0)
    await userEvent.click(await deckmenue('Aschenrat'))
    for (const name of [
      texte.umbenennen,
      texte.duplizieren,
      teilen.knopf,
      texte.loeschen,
    ]) {
      expect(
        screen.getByRole('menuitem', { name: new RegExp(`^${name}`) }),
      ).toBeInTheDocument()
    }
    expect(
      screen.queryByRole('button', {
        name: einsetzen(texte.handlungenAn, { name: 'Bernsteinflut' }),
      }),
    ).toBeNull()
  })

  it('öffnet mit dem Rechtsklick ein anderes Deck und dann sein Menü', async () => {
    zeige()
    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` })

    fireEvent.contextMenu(screen.getByText('Bernsteinflut'), {
      clientX: 10,
      clientY: 10,
    })

    const menue = await screen.findByRole('menu', {
      name: einsetzen(texte.handlungenAn, { name: 'Bernsteinflut' }),
    })
    expect(
      within(menue).getByRole('menuitem', {
        name: new RegExp(`^${texte.loeschen}`),
      }),
    ).toBeInTheDocument()
  })

  // **Escape gibt den Fokus an »…« zurück** — UX-08.AK-01, auch mit einem
  // Klick wie in WebKit: seit AP-6 (#2363) öffnet ein Menüeintrag das
  // Popover, und der Eintrag ist beim Schließen mit dem Menü fort. Am
  // Fenster hält es `ux08-popover` (QA01, #2382).
  it('gibt den Fokus nach Escape im Umbenennen an das Menü zurück', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )
    await userEvent.click(await deckmenue('Aschenrat'))

    // **Ein Klick wie in WebKit** (`fokus.ts`): der Druck merkt den Eintrag,
    // der Fokus geht an den nächsten fokussierbaren Vorfahren — in der
    // Anwendung der Inhaltsbereich —, und der Klick selbst fokussiert nicht.
    const eintrag = screen.getByRole('menuitem', {
      name: new RegExp(`^${texte.umbenennen}`),
    })
    const vorfahr = screen.getByRole('region', { name: texte.titel })
    vorfahr.tabIndex = -1
    fireEvent.pointerDown(eintrag)
    vorfahr.focus()
    fireEvent.click(eintrag)

    await screen.findByRole('group', {
      name: einsetzen(texte.umbenennenTitel, { name: 'Aschenrat' }),
    })
    await userEvent.keyboard('{Escape}')

    await waitFor(async () => {
      expect(await deckmenue('Aschenrat')).toHaveFocus()
    })
  })

  it('räumt die Frage vor dem Löschen, wenn ein anderes Deck aufgeht', async () => {
    zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )
    await deckhandlung('Aschenrat', texte.loeschen)
    expect(screen.getByText(texte.loeschenFrage)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Bernsteinflut` }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )

    await deckmenue('Aschenrat')
    expect(screen.queryByText(texte.loeschenFrage)).toBeNull()
  })

  it('bereitet über »Teilen …« im Menü das Teilen des Decks vor', async () => {
    const vorbereiten = vi.fn(() =>
      Promise.resolve({ teilgut: null, fehler: null }),
    )
    zeige(vorbereiten)

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )
    await deckhandlung('Aschenrat', teilen.knopf)

    await waitFor(() => {
      expect(vorbereiten).toHaveBeenCalledWith(
        'deck',
        expect.any(String) as string,
      )
    })
  })

  // **Schnell spielen steht vorn** — P6 (#2374), Artboard `PartieStart09`:
  // aus dem Menü des Decks gleich an den Tisch (S1 aus dem Deck in drei).
  it('führt mit »Schnell spielen« zum Spielen', async () => {
    const wechseln = vi.fn()
    zeige(undefined, wechseln)

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Aschenrat`,
      }),
    )
    await userEvent.click(await deckmenue('Aschenrat'))
    const eintraege = screen.getAllByRole('menuitem')
    expect(eintraege[0]).toHaveTextContent(
      kataloge.de.seiten.spielen.schnellSpielen,
    )

    await userEvent.click(eintraege[0] ?? document.body)

    expect(wechseln).toHaveBeenCalledWith('spielen')
  })
})
