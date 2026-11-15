// **Die Skriptkonsole mit Abbruch, Dauer, Verlauf und Ergebnisbaum** —
// Werkstück U13 von Phase 5.1 (#1424).
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KonsolenAntwort } from '../antworten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { leere } from '../skriptverlauf'
import { belegeApi, raeumeApi } from '../test/api'
import { Skriptkonsole } from './Skriptkonsole'

const texte = kataloge.de.seiten.karten.konsole

function antwort(teile: Partial<KonsolenAntwort> = {}): KonsolenAntwort {
  return {
    quelltext: '2',
    text: '2',
    art: 'Zahl',
    befunde: [],
    stelle: null,
    dauerMs: 12,
    abgebrochen: false,
    eintraege: [],
    fehler: null,
    ...teile,
  }
}

function zeige(teile: Parameters<typeof belegeApi>[0]) {
  belegeApi(teile)

  render(
    <Sprachraum sprache="de">
      <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
  leere()
})

describe('Ein Lauf der Skriptkonsole', () => {
  it('wertet mit ⌘↩ aus, zeigt die Dauer und trägt den Lauf in den Verlauf', async () => {
    const Skript = vi.fn(() => Promise.resolve(antwort()))
    zeige({ Skript })

    const feld = screen.getByLabelText(texte.ausdruck)
    await userEvent.type(feld, '1 + 1')
    fireEvent.keyDown(feld, { key: 'Enter', metaKey: true })

    expect(await screen.findByText('2')).toBeInTheDocument()
    expect(Skript).toHaveBeenCalledWith('7', '1 + 1')
    expect(screen.getByText('12 ms')).toBeInTheDocument()

    const verlauf = within(screen.getByRole('region', { name: texte.verlauf }))
    expect(verlauf.getByRole('button', { name: /1 \+ 1/ })).toBeInTheDocument()
  })

  // **Eingefügt ist nicht ausgeführt.** Ein Schnipsel aus fremder Quelle
  // landet im Feld und läuft erst, wenn jemand es ausdrücklich will.
  it('wertet ein eingefügtes Schnipsel nicht von selbst aus', async () => {
    const Skript = vi.fn(() => Promise.resolve(antwort()))
    zeige({ Skript })

    const feld = screen.getByLabelText(texte.ausdruck)
    await userEvent.click(feld)
    await userEvent.paste('card.name')

    expect(feld).toHaveValue('card.name')
    expect(Skript).not.toHaveBeenCalled()

    // Der Beleg, dass der Weg offen ist (§3): erst der Knopf ruft.
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))
    await waitFor(() => {
      expect(Skript).toHaveBeenCalledWith('7', 'card.name')
    })
  })

  it('lässt sich abbrechen, und die Antwort sagt es', async () => {
    let freigeben: (() => void) | undefined
    const angehalten = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })
    const Skript = vi.fn(async () => {
      await angehalten

      return antwort({ quelltext: '', text: '', art: '', abgebrochen: true })
    })
    const SkriptAbbrechen = vi.fn(() => Promise.resolve({ fehler: null }))
    zeige({ Skript, SkriptAbbrechen })

    await userEvent.type(screen.getByLabelText(texte.ausdruck), 'endlos')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    // Solange es läuft: der Status und der Knopf zum Abbrechen.
    expect(await screen.findByText(texte.laeuft)).toBeInTheDocument()

    // **Ein Lauf je Karte** (§7, Regel 4): ⌘↩ während des Laufs startet
    // keinen zweiten.
    fireEvent.keyDown(screen.getByLabelText(texte.ausdruck), {
      key: 'Enter',
      metaKey: true,
    })
    expect(Skript).toHaveBeenCalledTimes(1)
    await userEvent.click(screen.getByRole('button', { name: texte.abbrechen }))
    expect(SkriptAbbrechen).toHaveBeenCalledWith('7')

    freigeben?.()

    expect(await screen.findByText(texte.abgebrochen)).toBeInTheDocument()
    expect(screen.queryByText(texte.laeuft)).not.toBeInTheDocument()

    const verlauf = within(screen.getByRole('region', { name: texte.verlauf }))
    expect(verlauf.getByText(texte.ausgangAbgebrochen)).toBeInTheDocument()
  })

  // **Abbrechen ist Bestenfalls-Arbeit** (§7, Regel 3): scheitert der Ruf
  // selbst, endet der Lauf trotzdem — und die Seite bleibt stehen.
  it('schluckt einen Fehler beim Abbrechen', async () => {
    let freigeben: (() => void) | undefined
    const angehalten = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })
    const Skript = vi.fn(async () => {
      await angehalten

      return antwort({ abgebrochen: true })
    })
    const SkriptAbbrechen = vi.fn(() => Promise.reject(new Error('weg')))
    zeige({ Skript, SkriptAbbrechen })

    await userEvent.type(screen.getByLabelText(texte.ausdruck), 'endlos')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.abbrechen }),
    )
    expect(SkriptAbbrechen).toHaveBeenCalledWith('7')

    freigeben?.()
    expect(await screen.findByText(texte.abgebrochen)).toBeInTheDocument()
  })

  it('zeigt eine Abbildung als Baum', async () => {
    const Skript = vi.fn(() =>
      Promise.resolve(
        antwort({
          quelltext: '[a: 1, b: [c: 2]]',
          art: 'Abbildung',
          eintraege: [
            { name: 'a', quelltext: '1', art: 'Zahl', eintraege: [] },
            {
              name: 'b',
              quelltext: '[c: 2]',
              art: 'Abbildung',
              eintraege: [
                { name: 'c', quelltext: '2', art: 'Zahl', eintraege: [] },
              ],
            },
          ],
        }),
      ),
    )
    zeige({ Skript })

    await userEvent.type(screen.getByLabelText(texte.ausdruck), 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    const baum = await screen.findByRole('tree', { name: texte.baum })
    expect(within(baum).getByText('a')).toBeInTheDocument()
    expect(within(baum).getByText('c')).toBeInTheDocument()
  })
})

describe('Der Verlauf in der Konsole', () => {
  it('lädt einen Ausdruck zurück, redigiert Geheimnisse und lässt sich löschen', async () => {
    const Skript = vi.fn(() => Promise.resolve(antwort()))
    zeige({ Skript })

    const feld = screen.getByLabelText(texte.ausdruck)
    await userEvent.type(feld, 'set.token')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))
    await screen.findByText('12 ms')

    await userEvent.clear(feld)
    await userEvent.type(feld, '1 + 1')
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))
    await waitFor(() => {
      expect(Skript).toHaveBeenCalledTimes(2)
    })

    // **Auf den zweiten Eintrag gewartet, nicht auf den Ruf** (#1809): er
    // steht erst, wenn die Antwort da ist — und der Griff darunter gilt ihm.
    // Im Bereich gesucht, denn „1 + 1" steht auch unter den Beispielen.
    const bereich = await screen.findByRole('region', { name: texte.verlauf })
    await within(bereich).findByRole('button', { name: /1 \+ 1/ })

    const verlauf = within(bereich)

    // Der redigierte Eintrag steht da — und lässt sich nicht zurückladen.
    expect(verlauf.getByText('•••')).toBeInTheDocument()
    expect(verlauf.queryByText('set.token')).not.toBeInTheDocument()
    expect(verlauf.getByRole('button', { name: /•••/ })).toBeDisabled()

    await userEvent.clear(feld)
    await userEvent.click(verlauf.getByRole('button', { name: /1 \+ 1/ }))
    expect(feld).toHaveValue('1 + 1')

    await userEvent.click(
      verlauf.getByRole('button', { name: texte.verlaufLoeschen }),
    )
    expect(screen.getByText(texte.verlaufLeer)).toBeInTheDocument()
  })
})
