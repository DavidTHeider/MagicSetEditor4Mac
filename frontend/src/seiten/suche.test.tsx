// Die Suchseite.
//
// Geprüft wird, was die Seite **zusagt**: dass die Filter kombiniert ankommen,
// dass die Filterwerte aus dem Bestand stammen, dass eine Kürzung zu sehen ist
// und dass ein Treffer die Karte aufschlägt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from '../App'
import type { SuchAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from '../test/api'
import { Suche } from './Suche'
import { ausKarten } from '../test/kartenproben'
import { leisteneintrag } from '../test/leistenweg'

const texte = kataloge.de.seiten.suche

// treffer baut einen Treffer mit sprechenden Vorgaben.
function treffer(teile: Partial<SuchAntwort['treffer'][number]> = {}) {
  return {
    karteID: '1',
    setID: '7',
    nummer: 0,
    kartenname: 'Feuerball',
    feld: 'text',
    wert: 'Feuerball verursacht Schaden.',
    ...teile,
  }
}

function zeige(oeffnen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Suche oeffnen={oeffnen} />
    </Sprachraum>,
  )

  return oeffnen
}

// tippen füllt das Suchfeld **der Volltextsuche**.
//
// **Mit Namen und nicht über die Rolle allein** (A16, #1515): seit die
// Recherche im selben Arbeitsplatz steht, gibt es zwei Suchfelder. Beide
// tragen ein Etikett — für eine Vorlesehilfe wie für diesen Lauf.
async function tippe(text: string) {
  await userEvent.type(
    screen.getByRole('searchbox', { name: texte.eingabe }),
    text,
  )
}

describe('Suche', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('sucht erst, wenn etwas dasteht', async () => {
    const Suchen = vi.fn(() =>
      Promise.resolve({ treffer: [], gekuerzt: false }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen,
    })

    zeige()

    // Ein leerer Zustand erklärt den nächsten Schritt (§6).
    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
    expect(Suchen).not.toHaveBeenCalled()
  })

  // **Die Filter sind kombinierbar** — und sie kommen zusammen an.
  it('reicht Text, Feld und Set gemeinsam durch', async () => {
    const Suchen = vi.fn(() =>
      Promise.resolve({ treffer: [], gekuerzt: false }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Erstes Set',
              spiel: 'magic',
              stil: 'm15',
              herkunft: '',
              karten: 1,
              herkunftFehlt: false,
            },
          ],
        }),
      Suchfelder: () => Promise.resolve({ felder: ['name', 'typ'] }),
      Suchen,
    })

    zeige()

    await tippe('feuer')
    await userEvent.selectOptions(
      await screen.findByLabelText(texte.set),
      'Erstes Set',
    )
    await userEvent.selectOptions(
      await screen.findByLabelText(texte.feld),
      'typ',
    )

    await waitFor(() => {
      expect(Suchen).toHaveBeenLastCalledWith('feuer', 'typ', '7')
    })
  })

  // **Die Filterwerte kommen aus dem Bestand**, nicht aus einer Liste bei uns.
  it('bietet die Felder an, die der Bestand führt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: ['casting cost', 'name'] }),
      Suchen: () => Promise.resolve({ treffer: [], gekuerzt: false }),
    })

    zeige()

    expect(
      await screen.findByRole('option', { name: 'casting cost' }),
    ).toBeInTheDocument()
  })

  // **Ein Wechsel des Sets räumt das Feld mit** (§7, Regel 1): ein Feld, das
  // nur das vorige Set führt, filterte im neuen alles weg.
  it('räumt das Feld, wenn das Set wechselt', async () => {
    const Suchen = vi.fn(() =>
      Promise.resolve({ treffer: [], gekuerzt: false }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Erstes Set',
              spiel: 'magic',
              stil: 'm15',
              herkunft: '',
              karten: 1,
              herkunftFehlt: false,
            },
          ],
        }),
      Suchfelder: () => Promise.resolve({ felder: ['typ'] }),
      Suchen,
    })

    zeige()

    await tippe('feuer')
    await userEvent.selectOptions(
      await screen.findByLabelText(texte.feld),
      'typ',
    )
    await userEvent.selectOptions(
      await screen.findByLabelText(texte.set),
      'Erstes Set',
    )

    await waitFor(() => {
      expect(Suchen).toHaveBeenLastCalledWith('feuer', '', '7')
    })
  })

  it('zeigt Treffer mit Kartenname, Feld und Wert', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () => Promise.resolve({ treffer: [treffer()], gekuerzt: false }),
    })

    zeige()
    await tippe('feuer')

    expect(
      await screen.findByRole('button', { name: 'Feuerball' }),
    ).toBeInTheDocument()
    expect(
      screen.getByText('Feuerball verursacht Schaden.'),
    ).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveTextContent(`1 ${texte.treffer}`)
  })

  // Eine Karte ohne Namen bleibt anklickbar — sie kommt aus einem fremden Set.
  it('nennt eine Karte ohne Namen beim Ersatznamen', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () =>
        Promise.resolve({
          treffer: [treffer({ kartenname: '' })],
          gekuerzt: false,
        }),
    })

    zeige()
    await tippe('feuer')

    expect(
      await screen.findByRole('button', {
        name: kataloge.de.seiten.karten.ohneName,
      }),
    ).toBeInTheDocument()
  })

  // **Die Kürzung ist zu sehen.** Eine stille sähe aus wie ein vollständiges
  // Ergebnis.
  it('sagt, dass es mehr Treffer gibt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () => Promise.resolve({ treffer: [treffer()], gekuerzt: true }),
    })

    zeige()
    await tippe('feuer')

    expect(await screen.findByRole('status')).toHaveTextContent(texte.gekuerzt)
  })

  it('erklärt eine leere Trefferliste, statt nichts zu zeigen', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () => Promise.resolve({ treffer: [], gekuerzt: false }),
    })

    zeige()
    await tippe('feuer')

    expect(await screen.findByText(texte.nichts)).toBeInTheDocument()
  })

  it('meldet einen Fehler aus der Bindung am Ort', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () =>
        Promise.resolve({
          treffer: [],
          gekuerzt: false,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Die Suche ging schief.',
          },
        }),
    })

    zeige()
    await tippe('feuer')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Suche ging schief.',
    )
  })

  it('gibt einen Treffer nach außen weiter', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () => Promise.resolve({ treffer: [treffer()], gekuerzt: false }),
    })

    const oeffnen = zeige()
    await tippe('feuer')

    await userEvent.click(
      await screen.findByRole('button', { name: 'Feuerball' }),
    )

    expect(oeffnen).toHaveBeenCalledWith('7', '1')
  })
})

// **Der Weg von einem Treffer zur Karte** — durch die ganze Anwendung, weil
// zwischen Suchseite und Editor die Zustände von `App` liegen.
describe('Vom Treffer zur Karte', () => {
  afterEach(() => {
    raeumeApi()
    localStorage.clear()
  })

  it('schlägt die getroffene Karte im Editor auf', async () => {
    belegeApi({
      ...belegeFenster('de'),
      // **Ein Set muss offen sein.** Die Suche geht über ein Set — seit
      // #288
      // steht sie als Lasche im Arbeitsplatz und nicht mehr in der Leiste, wo
      // sie ohne Set anklickbar war und nichts finden konnte.
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Probeset',
              spiel: 'magic',
              stil: 'magic-m15',
              karten: 1,
              herkunft: '/a.mse-set',
              herkunftFehlt: false,
            },
          ],
        }),
      Suchfelder: () => Promise.resolve({ felder: [] }),
      Suchen: () =>
        Promise.resolve({
          treffer: [treffer({ karteID: '3', setID: '7' })],
          gekuerzt: false,
        }),
      ...ausKarten(() =>
        Promise.resolve({
          karten: [
            {
              id: '3',
              nummer: 2,
              werte: [{ feld: 'name', wert: 'Feuerball' }],
            },
          ],
        }),
      ),
      Karten: () =>
        Promise.resolve({
          karten: [
            {
              id: '3',
              nummer: 2,
              werte: [{ feld: 'name', wert: 'Feuerball' }],
            },
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    render(<App />)

    // Erst das Set aufschlagen, dann die Lasche.
    await userEvent.click(await leisteneintrag('sets'))
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Probeset — ${kataloge.de.seiten.sets.oeffnen}`,
      }),
    )

    await userEvent.click(
      // **Eine Lasche**, seit die Setseiten im Arbeitsplatz stehen (#288).
      await screen.findByRole('tab', { name: kataloge.de.seiten.suche.titel }),
    )

    await tippe('feuer')

    await userEvent.click(
      await screen.findByRole('button', { name: 'Feuerball' }),
    )

    // Der Editor steht offen — mit dem Feld der getroffenen Karte, nicht mit
    // dem Hinweis, dass keine gewählt ist.
    expect(await screen.findByLabelText('name')).toHaveValue('Feuerball')
  })
})
