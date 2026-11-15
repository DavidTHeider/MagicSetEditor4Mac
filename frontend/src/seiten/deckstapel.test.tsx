// **Ein Stapel aus der Sammlung landet im offenen Deck** — UX-03.AK-01
// (#1675),
// durch die echte Deckseite: `mehrfachanschluss.test.tsx` belegt die
// Aufnahme mit einem nachgebildeten Legen, dieser Lauf das Legen selbst.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'

import type { Anschluss } from '../anschluss'
import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

const stapel: Anschluss = {
  art: 'deckstapel',
  name: '5 Exemplare',
  stuecke: [
    { printing: 'druck-a', name: 'Aschenwächter', menge: 4 },
    { printing: 'druck-b', name: 'Glut', menge: 1 },
  ],
}

async function zeige(
  DeckplatzLegen: Api['DeckplatzLegen'],
  anschluss: Anschluss = stapel,
) {
  const Deck = vi.fn(() => Promise.resolve({ deck: deckprobe(), fehler: null }))
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Deck,
    DeckBenutzt: () => Promise.resolve({ fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    DeckplatzLegen,
  })

  render(
    <Anschlussbuehne anschluss={anschluss}>
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>
    </Anschlussbuehne>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )
  await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

  return Deck
}

describe('der Stapel im Deck', () => {
  it('legt jeden Druck ins Hauptdeck und lädt danach neu', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: platzprobe(), fehler: null }),
    )
    const Deck = await zeige(DeckplatzLegen)
    const vorher = Deck.mock.calls.length

    await userEvent.click(
      await screen.findByRole('button', { name: texte.stapelLegen }),
    )

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledTimes(2)
    })
    expect(DeckplatzLegen).toHaveBeenNthCalledWith(
      1,
      deckprobe().id,
      '',
      'druck-a',
      4,
      'hauptdeck',
    )
    expect(DeckplatzLegen).toHaveBeenNthCalledWith(
      2,
      deckprobe().id,
      '',
      'druck-b',
      1,
      'hauptdeck',
    )
    await waitFor(() => {
      expect(Deck.mock.calls.length).toBeGreaterThan(vorher)
    })

    // **Verbraucht**: der Kasten ist fort.
    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: texte.stapelLegen }),
      ).toBeNull()
    })
  })

  // **Geschrieben ist geschrieben** (§7, Regel 3): was lag, bleibt liegen,
  // und das Deck zeigt es. **Warum der Rest nicht liegt, steht am Stapel** —
  // die WebKit-Probe fand die erste Fassung, deren Nachladen die Meldung
  // verschluckte: sie galt danach als Antwort eines veralteten Laufs.
  it('zeigt am Stapel, warum der Rest nicht liegt, und legt ihn nach', async () => {
    const DeckplatzLegen = vi
      .fn<Api['DeckplatzLegen']>()
      .mockResolvedValueOnce({ platz: platzprobe(), fehler: null })
      .mockResolvedValueOnce({
        platz: null,
        fehler: {
          code: 'nichtGefunden',
          meldung: 'nicht gefunden',
          feld: '',
          schluessel: 'nichtGefunden',
        },
      })
      .mockResolvedValue({ platz: platzprobe(), fehler: null })
    const Deck = await zeige(DeckplatzLegen)
    const vorher = Deck.mock.calls.length

    await userEvent.click(
      await screen.findByRole('button', { name: texte.stapelLegen }),
    )

    const kasten = screen.getByRole('group', { name: texte.stapelTitel })
    expect(await within(kasten).findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.nichtGefunden,
    )
    expect(kasten).toHaveTextContent(
      einsetzen(texte.stapelTeilweise, { gelegt: '1', gesamt: '2' }),
    )

    // **Was lag, zeigt das Deck** — nachgeladen, obwohl der Rest scheiterte.
    await waitFor(() => {
      expect(Deck.mock.calls.length).toBeGreaterThan(vorher)
    })

    // **Und der zweite Klick legt nur den Rest.**
    await userEvent.click(
      within(kasten).getByRole('button', { name: texte.stapelLegen }),
    )
    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledTimes(3)
    })
    expect(DeckplatzLegen).toHaveBeenLastCalledWith(
      deckprobe().id,
      '',
      'druck-b',
      1,
      'hauptdeck',
    )
  })

  // **Ein abgewiesenes Stück hält die übrigen nicht auf** — Befund G16 des
  // Zweitblicks über Phase 6.3 (#1666). Hier brach der Stapel beim ersten
  // Fehler ab, und was danach kam, wurde gar nicht versucht: dieselbe
  // Gestalt, die #1995 beim Umschlag schloss und die Vermerk 0127 für die
  // Sammlung verwirft — **eine Zeile scheitert, nicht die Liste**.
  it('legt nach einem abgewiesenen Stück die übrigen und nennt jedes', async () => {
    const drei: Anschluss = {
      art: 'deckstapel',
      name: '3 Exemplare',
      stuecke: [
        { printing: 'bbb-2', name: 'Ohne Kennung', menge: 1 },
        { printing: 'druck-b', name: 'Glut', menge: 1 },
        { printing: 'druck-c', name: 'Zuviel', menge: 1 },
      ],
    }
    const abgewiesen = (feld: string) => ({
      platz: null,
      fehler: {
        code: 'validierung' as const,
        meldung: 'roh',
        feld,
        schluessel: 'validierung',
      },
    })
    const DeckplatzLegen = vi
      .fn<Api['DeckplatzLegen']>()
      .mockResolvedValueOnce(abgewiesen('printing'))
      .mockResolvedValueOnce({ platz: platzprobe(), fehler: null })
      .mockResolvedValueOnce(abgewiesen('menge'))
      .mockResolvedValue({ platz: platzprobe(), fehler: null })
    await zeige(DeckplatzLegen, drei)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.stapelLegen }),
    )

    const kasten = screen.getByRole('group', { name: texte.stapelTitel })
    const grund = await within(kasten).findByRole('alert')
    expect(grund).toHaveTextContent(
      einsetzen(texte.stapelOhneDruck, { name: 'Ohne Kennung' }),
    )
    expect(grund).toHaveTextContent(
      einsetzen(texte.stapelGrund, {
        name: 'Zuviel',
        grund: kataloge.de.fehler.validierung,
      }),
    )
    expect(kasten).toHaveTextContent(
      einsetzen(texte.stapelTeilweise, { gelegt: '1', gesamt: '3' }),
    )
    expect(DeckplatzLegen).toHaveBeenCalledTimes(3)

    // **Und der zweite Klick legt nur die beiden abgewiesenen.**
    await userEvent.click(
      within(kasten).getByRole('button', { name: texte.stapelLegen }),
    )
    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: texte.stapelLegen }),
      ).toBeNull()
    })
    expect(DeckplatzLegen.mock.calls.slice(3).map((ruf) => ruf[2])).toEqual([
      'bbb-2',
      'druck-c',
    ])
  })

  // **Ein Grund, der nicht am Stück liegt, hält den Stapel an** — derselbe
  // Schnitt wie bei der Listeneinfuhr in Go: ein Eingabefehler betrifft
  // eine Zeile, alles andere jede. Ein fehlendes Deck nähme kein einziges
  // Stück, und dreimal derselbe Satz sagte nichts Neues.
  it('hält an einem Grund an, der nicht am Stück liegt', async () => {
    const DeckplatzLegen = vi.fn<Api['DeckplatzLegen']>().mockResolvedValue({
      platz: null,
      fehler: {
        code: 'nichtGefunden',
        meldung: 'nicht gefunden',
        feld: '',
        schluessel: 'nichtGefunden',
      },
    })
    await zeige(DeckplatzLegen)

    await userEvent.click(
      await screen.findByRole('button', { name: texte.stapelLegen }),
    )

    const kasten = screen.getByRole('group', { name: texte.stapelTitel })
    expect(await within(kasten).findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.nichtGefunden,
    )
    expect(kasten).toHaveTextContent(
      einsetzen(texte.stapelTeilweise, { gelegt: '0', gesamt: '2' }),
    )
    expect(DeckplatzLegen).toHaveBeenCalledTimes(1)
  })
})
