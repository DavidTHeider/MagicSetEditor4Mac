// **Eine Karte lässt sich löschen — nach einer Rückfrage**
// (#621).
//
// #562 hat das
// Anlegen gebaut und das Löschen vertagt: *„eine gelöschte Karte ohne ⌘Z ist
// Datenverlust mit einem Klick."* Seit der Verlauf eine eigene Art für ganze
// Karten trägt, ist der Weg zurück da — und die Rückfrage sagt es dazu.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneKartenhandlungen } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function karte(id: string, name: string): KarteDTO {
  return { id, nummer: Number(id) - 1, werte: [{ feld: 'name', wert: name }] }
}

function zeige(
  teile: Partial<Api>,
  bestandGeaendert = () => undefined,
  bestand: () => KarteDTO[] = () => [karte('1', 'Erste'), karte('2', 'Zweite')],
) {
  belegeApi({
    ...ausKarten(() => Promise.resolve({ karten: bestand() })),
    Karten: () => Promise.resolve({ karten: bestand() }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    // **Ein eigener `Karten`-Beleg wird mitprojiziert** (#891). Die Seite holt
    // `Kartenzeilen` und `Karte`; ohne diese Zeile bliebe die Vorgabe darüber
    // stehen, und der Test prüfte Karten, die er nicht beschrieben hat.
    ...(teile.Karten ? ausKarten(teile.Karten, '7') : {}),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="7"
        karteID="1"
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={bestandGeaendert}
      />
    </Sprachraum>,
  )
}

// **Eine Karte, die in einem Deck liegt, bekommt eine zweite Frage** —
// #1369.
//
// Der Dienst weist ab und nennt die Decks; erst dann steht das Trotzdem da.
describe('eine Karte in einem Deck', () => {
  it('nennt die Decks und bietet das Trotzdem an', async () => {
    const KarteLoeschen = vi
      .fn()
      .mockResolvedValueOnce({
        fehler: {
          code: 'konflikt',
          schluessel: 'karteInDeck',
          meldung: 'Diese Karte liegt in 1 Deck(s) — Aggro.',
          feld: 'karteID',
          werte: { anzahl: '1', decks: 'Aggro' },
        },
      })
      .mockResolvedValueOnce({ fehler: null })

    zeige({ KarteLoeschen })

    await oeffneKartenhandlungen()
    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    // **Der Satz aus dem Katalog, mit eingesetzten Namen.**
    expect(await screen.findByText(/Aggro/)).toBeInTheDocument()

    // **Und erst jetzt das Trotzdem.**
    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.loeschenTrotzdem,
      }),
    )

    expect(KarteLoeschen).toHaveBeenLastCalledWith('7', '1', true)
  })

  it('bietet nach einem gewöhnlichen Fehler kein Trotzdem an', async () => {
    const KarteLoeschen = vi.fn().mockResolvedValue({
      fehler: {
        code: 'unbekannt',
        schluessel: 'unbekannt',
        meldung: 'etwas ging schief',
        feld: '',
        werte: {},
      },
    })

    zeige({ KarteLoeschen })

    await oeffneKartenhandlungen()
    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    await screen.findByText(texte.fehler.unbekannt)

    expect(
      screen.queryByRole('button', {
        name: texte.seiten.karten.loeschenTrotzdem,
      }),
    ).not.toBeInTheDocument()
  })
})
