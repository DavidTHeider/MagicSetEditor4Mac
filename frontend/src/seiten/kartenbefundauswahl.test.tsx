// **Die Karten mit Befund, gefiltert** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`: die Zeile eines Sets in der Aufmerksamkeit öffnet seine
// Kartenliste mit genau den Karten mit Befund. Der Filter nennt sich und
// lässt sich lösen — und gilt nur für das Set, dem er gehört.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { effekteNachholen } from '../test/effekte'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
  karteGewaehlt.mockClear()
})

const texte = kataloge.de.seiten.karten

function karte(id: string, nummer: number, name: string): KarteDTO {
  return { id, nummer, werte: [{ feld: 'name', wert: name }] }
}

const karteGewaehlt = vi.fn()

function seite(setID: string, anschluss: Anschluss) {
  return (
    <Sprachraum sprache="de">
      <Anschlussbuehne anschluss={anschluss}>
        <Karten
          key={setID}
          karteGewaehlt={karteGewaehlt}
          setID={setID}
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Anschlussbuehne>
    </Sprachraum>
  )
}

function zeige(anschluss: Anschluss) {
  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    ...ausKarten(
      (setID) =>
        Promise.resolve({
          karten: [
            karte('8', 0, setID === '7' ? 'Erste' : 'Fremde'),
            karte('9', 1, 'Zweite'),
            karte('10', 2, 'Dritte'),
          ],
        }),
      '7',
    ),
  })
  return render(seite('7', anschluss))
}

const liste = () => within(screen.getByRole('grid', { name: texte.titel }))

describe('die Karten mit Befund', () => {
  it('stehen gefiltert da, und der Filter lässt sich lösen', async () => {
    zeige({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '7',
      karteIDs: ['9', '10'],
    })
    await screen.findByRole('grid', { name: texte.titel })
    expect(await liste().findByText('Erste')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(await screen.findByText('2 Karten mit Befund')).toBeInTheDocument()
    expect(liste().queryByText('Erste')).toBeNull()
    // **Aufgeschlagen ist die erste Karte mit Befund** — nicht die vorige,
    // die keinen trägt: unter dem Satz „2 Karten mit Befund“ stünde sonst
    // „Keine Probleme an dieser Karte“.
    await vi.waitFor(() => {
      expect(karteGewaehlt).toHaveBeenLastCalledWith('9')
    })
    expect(liste().getByText('Zweite')).toBeInTheDocument()
    expect(liste().getByText('Dritte')).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.auswahlLoesen }),
    )

    expect(await liste().findByText('Erste')).toBeInTheDocument()
    expect(screen.queryByText('2 Karten mit Befund')).toBeNull()
  })

  it('gelten nicht für ein anderes Set', async () => {
    zeige({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '8',
      karteIDs: ['9', '10'],
    })
    await screen.findByRole('grid', { name: texte.titel })
    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(await liste().findByText('Erste')).toBeInTheDocument()
    expect(screen.queryByText('2 Karten mit Befund')).toBeNull()
  })

  it('kommen nach einem Setwechsel nicht wieder', async () => {
    const auswahl: Anschluss = {
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '7',
      karteIDs: ['9', '10'],
    }
    const { rerender } = zeige(auswahl)
    await screen.findByRole('grid', { name: texte.titel })
    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
    expect(await screen.findByText('2 Karten mit Befund')).toBeInTheDocument()

    rerender(seite('8', auswahl))
    await screen.findByRole('grid', { name: texte.titel })
    expect(await liste().findByText('Fremde')).toBeInTheDocument()
    rerender(seite('7', auswahl))
    await screen.findByRole('grid', { name: texte.titel })

    // **Der Beleg, dass die Liste wieder steht** — erst dann sagt das Fehlen
    // des Satzes etwas.
    expect(await liste().findByText('Erste')).toBeInTheDocument()
    expect(screen.queryByText('2 Karten mit Befund')).toBeNull()
  })

  it('lassen eine offene Karte mit Befund offen', async () => {
    zeige({
      art: 'kartenauswahl',
      name: '2 Karten mit Befund',
      setID: '7',
      karteIDs: ['8', '10'],
    })
    await screen.findByRole('grid', { name: texte.titel })
    await vi.waitFor(() => {
      expect(karteGewaehlt).toHaveBeenLastCalledWith('8')
    })
    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(await liste().findByText('Dritte')).toBeInTheDocument()
    expect(liste().queryByText('Zweite')).toBeNull()
    await effekteNachholen()
    expect(karteGewaehlt).toHaveBeenLastCalledWith('8')
  })

  // **Kennungen, die es nicht mehr gibt**, wählen nichts — die Liste ist
  // leer, und der Satz sagt, warum.
  it('wählen nichts, wenn keine der Karten mehr da ist', async () => {
    zeige({
      art: 'kartenauswahl',
      name: '1 Karte mit Befund',
      setID: '7',
      karteIDs: ['99'],
    })
    await screen.findByRole('grid', { name: texte.titel })
    await vi.waitFor(() => {
      expect(karteGewaehlt).toHaveBeenLastCalledWith('8')
    })
    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(await screen.findByText('1 Karte mit Befund')).toBeInTheDocument()
    await effekteNachholen()
    expect(karteGewaehlt).toHaveBeenLastCalledWith('8')
  })
})
