// **Der Deckbauplatz — drei Bereiche, und jede Änderung geht durch die
// Naht** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Eine eigene Datei**, damit `decks.test.tsx` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, dass ein Deck **entsteht**,
// hier, dass man an einem offenen **arbeitet**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import { platzeintrag } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

const platz = platzprobe({
  id: '3',
  bezeichnung: 'Blitz',
  menge: 4,
  typzeile: 'Instant',
  manakosten: '{R}',
})

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () =>
      Promise.resolve({
        deck: deckprobe({ plaetze: [platz], hauptdeck: 4 }),
        fehler: null,
      }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// aufschlagen öffnet das Deck über den Navigator.
async function aufschlagen() {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )

  return await screen.findByRole('list', { name: texte.plaetze })
}

afterEach(() => {
  raeumeApi()
})

describe('der Deckbauplatz', () => {
  // **Einmal, im Zentrum** — UX-15 (#1771). Hier stand derselbe Satz in
  // Zentrum und Inspector, in der Sichtprüfung vom 2026-09-24 nebeneinander;
  // der Inspector sagt jetzt, was **er** zeigen wird.
  it('sagt im Zentrum, dass kein Deck offen ist, und im Inspector, was dort erscheint', async () => {
    zeige()

    expect(await screen.findAllByText(texte.keinDeckOffen)).toHaveLength(1)
    expect(
      screen.getByRole('complementary', { name: texte.analyse }),
    ).toHaveTextContent(texte.analyseLeer)
  })

  it('zählt eine Karte herauf — über die Naht, mit der neuen Menge', async () => {
    const legen = vi
      .fn()
      .mockResolvedValue({ platz: { ...platz, menge: 5 }, fehler: null })

    zeige({ DeckplatzLegen: legen })

    await aufschlagen()

    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))

    await waitFor(() => {
      expect(legen).toHaveBeenCalledWith(
        '1',
        '',
        platz.printing,
        5,
        'hauptdeck',
      )
    })
  })

  // **Der Wechsel ist ein Ruf** — Befund 7 des Zweitblicks über Phase 5.1.
  // Hier standen zwei (legen, dann nehmen), und ein gescheiterter zweiter
  // ließ dieselbe Karte in beiden Bereichen liegen.
  it('wechselt den Bereich in einem Ruf', async () => {
    const verschieben = vi.fn().mockResolvedValue({
      platz: { ...platz, bereich: 'sideboard' },
      fehler: null,
    })
    const legen = vi.fn().mockResolvedValue({ platz, fehler: null })
    const nehmen = vi.fn().mockResolvedValue({ fehler: null })

    zeige({
      DeckplatzVerschieben: verschieben,
      DeckplatzLegen: legen,
      DeckplatzNehmen: nehmen,
    })

    await aufschlagen()

    await userEvent.click(await platzeintrag('Blitz', texte.insSideboard))

    await waitFor(() => {
      expect(verschieben).toHaveBeenCalledWith('1', platz.id, 'sideboard')
    })

    // **Und die zwei alten Wege bleiben unberührt** — der Beleg, dass es
    // wirklich einer ist und nicht einer mehr.
    expect(legen).not.toHaveBeenCalled()
    expect(nehmen).not.toHaveBeenCalled()
  })

  it('nimmt einen Platz heraus', async () => {
    const nehmen = vi.fn().mockResolvedValue({ fehler: null })

    zeige({ DeckplatzNehmen: nehmen })

    await aufschlagen()

    await userEvent.click(await platzeintrag('Blitz', texte.nehmen))

    await waitFor(() => {
      expect(nehmen).toHaveBeenCalledWith('1', '3')
    })
  })

  it('zeigt die Analyse im Inspector — samt dem nachgelesenen Manawert', async () => {
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({
            plaetze: [
              platzprobe({
                id: '4',
                bezeichnung: 'Werwolf',
                typzeile: 'Creature — Werewolf',
                manakosten: '',
                manawert: 3,
                menge: 1,
              }),
            ],
            hauptdeck: 1,
          }),
          fehler: null,
        }),
    })

    await aufschlagen()

    await userEvent.click(
      screen.getByRole('tab', { name: texte.gebietAuswertung }),
    )

    // **3,00 im Schnitt** — aus dem nachgelesenen Wert, nicht aus den Kosten:
    // die sind leer, und gerechnet käme null heraus.
    expect(await screen.findByText('3.00')).toBeInTheDocument()
  })
})

// **Die Prüfung läuft bei jeder Änderung mit** — Werkstück U10 von Phase 5.1
// (#1421).
describe('die laufende Regelprüfung', () => {
  it('prüft beim Aufschlagen und noch einmal nach einer Änderung', async () => {
    const pruefen = vi.fn().mockResolvedValue({
      urteil: {
        formatname: 'Konstruiert',
        zulaessig: false,
        hauptdeck: 4,
        sideboard: 0,
        ungeprueft: 0,
        urteile: [],
        groessen: [],
      },
      befunde: [],
      fehler: null,
    })

    const legen = vi
      .fn()
      .mockResolvedValue({ platz: { ...platz, menge: 5 }, fehler: null })

    let gelesen = 0

    zeige({
      DeckPruefen: pruefen,
      DeckplatzLegen: legen,
      Deck: () => {
        gelesen += 1

        return Promise.resolve({
          deck: deckprobe({
            formatID: '5',
            plaetze: [{ ...platz, menge: gelesen === 1 ? 4 : 5 }],
            hauptdeck: gelesen === 1 ? 4 : 5,
          }),
          fehler: null,
        })
      },
    })

    await aufschlagen()

    // **Ohne Knopfdruck** — der Inspector prüft, sobald das Deck steht.
    await userEvent.click(
      screen.getByRole('tab', { name: texte.gebietPruefung }),
    )

    // **Ohne Knopfdruck geprüft** — und zwar von beiden Bereichen: das
    // Zentrum braucht das Urteil für die Marke am Posten, der Inspector für
    // den Kopf. Gezählt wird deshalb **relativ**, nicht absolut.
    await waitFor(() => {
      expect(pruefen).toHaveBeenCalled()
    })

    const vorher = pruefen.mock.calls.length

    // **Frisch abgefragt und nicht über den alten Knoten** (§3): der
    // Laschenwechsel baut den Arbeitsplatz neu auf, und ein Klick auf einen
    // abgehängten Knoten erreicht React gar nicht — er tut still nichts.
    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))

    await waitFor(() => {
      expect(legen).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(pruefen.mock.calls.length).toBeGreaterThan(vorher)
    })
  })

  it('fragt ohne Format nach keinem Urteil und sagt, warum', async () => {
    const pruefen = vi.fn().mockResolvedValue({
      urteil: null,
      befunde: [],
      fehler: null,
    })

    // **Der Beleg, dass jemand zuhört** (§3): dieselbe Naht, die das Urteil
    // holte, liest auch das Deck. Ohne ihn hieße das `not.toHaveBeenCalled`
    // nur, dass der Bereich gar nicht stand.
    const lesen = vi.fn().mockResolvedValue({
      deck: deckprobe({ plaetze: [platz], hauptdeck: 4 }),
      fehler: null,
    })

    zeige({ DeckPruefen: pruefen, Deck: lesen })

    await aufschlagen()

    await userEvent.click(
      screen.getByRole('tab', { name: texte.gebietPruefung }),
    )

    expect(
      await screen.findByText(texte.pruefungOhneFormat),
    ).toBeInTheDocument()

    expect(lesen).toHaveBeenCalled()
    expect(pruefen).not.toHaveBeenCalled()
  })

  // **Die zwei Kästen des Navigators** — Werkstück U10, Befund der
  // Sichtprüfung vom 2026-09-09: mit Anlage und Übernahme ausgeklappt stand
  // die Seite 1.847 Punkte hoch (1440 × 900), und die Kartensuche — das
  // Werkzeug, das man beim Bauen **immer** braucht — begann bei 607 und
  // endete unter der Falz.
  it('hält Anlage und Übernahme zu, bis jemand sie öffnet', async () => {
    zeige()

    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` })

    expect(
      screen.queryByRole('region', { name: texte.neu }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('region', { name: texte.uebernahme }),
    ).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.neu }))

    expect(screen.getByRole('region', { name: texte.neu })).toBeInTheDocument()

    // **Der zweite bleibt zu** — sie sind zwei Kästen und nicht einer.
    expect(
      screen.queryByRole('region', { name: texte.uebernahme }),
    ).not.toBeInTheDocument()
  })
})
