// **Was sich mit einem Symbol der Bibliothek tun lässt** — DOM-02
// (#1697):
// Einsetzen, Duplizieren, Umbenennen, Ausgeben und Löschen mit seinen Folgen
// (AK-02), dazu die Marken des Symbols (AK-01).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import {
  krone,
  mitSymbolen,
  ohneMarken,
  symbolantwort,
  zacke,
  zeigeBibliothek,
} from '../test/symbolbibliothek'

const texte = kataloge.de.seiten.karten.symbolbibliothek

afterEach(() => {
  raeumeApi()
})

// eintrag wartet auf die Zeile eines Symbols.
const eintrag = (name: string) => screen.findByRole('listitem', { name })

describe('Ein Symbol der Bibliothek', () => {
  it('wird eingesetzt, und das Fach erfährt es', async () => {
    const einsetzenAufruf = vi
      .fn()
      .mockResolvedValue({ datei: 'symbol2.mse-symbol', fehler: null })
    const liste = vi.fn().mockResolvedValue(mitSymbolen(krone))
    belegeApi({
      Symbolbibliothek: liste,
      BibliothekssymbolEinsetzen: einsetzenAufruf,
      MarkenJeObjekt: ohneMarken,
    })

    const eingesetzt = zeigeBibliothek()
    await userEvent.click(
      within(await eintrag('Krone')).getByRole('button', {
        name: texte.einsetzen,
      }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.eingesetzt, { name: 'Krone' }),
    )
    expect(einsetzenAufruf).toHaveBeenCalledWith('1', '7')
    expect(eingesetzt).toHaveBeenCalledTimes(1)
    await waitFor(() => {
      expect(liste).toHaveBeenCalledTimes(2)
    })
  })

  it('wird dupliziert und umbenannt — ein doppelter Name steht am Feld', async () => {
    const kopieren = vi
      .fn()
      .mockResolvedValue(symbolantwort({ id: '3', name: 'Krone (Kopie)' }))
    const umbenennen = vi
      .fn()
      .mockResolvedValueOnce({
        ...symbolantwort({}),
        fehler: {
          code: 'konflikt',
          schluessel: 'doppelterName',
          meldung: 'Diesen Namen gibt es hier schon.',
          feld: 'name',
          werte: {},
        },
      })
      .mockResolvedValueOnce(symbolantwort({ id: '1', name: 'Reif' }))
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone, zacke)),
      BibliothekssymbolKopieren: kopieren,
      BibliothekssymbolUmbenennen: umbenennen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Krone')

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.duplizieren }),
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.aufgenommen, { name: 'Krone (Kopie)' }),
    )
    expect(kopieren).toHaveBeenCalledWith('1', 'Krone (Kopie)')

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.umbenennen }),
    )
    const feld = within(zeile).getByRole('textbox', {
      name: einsetzen(texte.neuerName, { name: 'Krone' }),
    })
    await userEvent.clear(feld)
    await userEvent.type(feld, 'Zacke')
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.namenUebernehmen }),
    )

    await waitFor(() => {
      expect(feld).toHaveAccessibleDescription(kataloge.de.fehler.doppelterName)
    })
    expect(umbenennen).toHaveBeenLastCalledWith('1', 'Zacke')

    await userEvent.clear(feld)
    await userEvent.type(feld, 'Reif')
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.namenUebernehmen }),
    )

    await waitFor(() => {
      expect(feld).not.toBeInTheDocument()
    })
    expect(umbenennen).toHaveBeenLastCalledWith('1', 'Reif')
  })

  it('schließt das Umbenennen ohne zu schreiben', async () => {
    const umbenennen = vi.fn()
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      BibliothekssymbolUmbenennen: umbenennen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Krone')

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.umbenennen }),
    )
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.umbenennenAbbrechen }),
    )

    expect(within(zeile).queryByRole('textbox')).not.toBeInTheDocument()
    expect(umbenennen).not.toHaveBeenCalled()
  })

  it('wird ausgegeben — ein weggeklickter Dialog sagt nichts', async () => {
    const ausgeben = vi
      .fn()
      .mockResolvedValueOnce(symbolantwort({ abgebrochen: true }))
      .mockResolvedValueOnce(
        symbolantwort({ id: '1', pfad: '/tmp/k.mse-symbol' }),
      )
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      BibliothekssymbolAusgeben: ausgeben,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    const knopf = within(await eintrag('Krone')).getByRole('button', {
      name: texte.ausgeben,
    })

    await userEvent.click(knopf)
    await antwortAbwarten()
    expect(ausgeben).toHaveBeenCalledWith('1', 'Krone')
    expect(screen.queryByRole('status')).not.toBeInTheDocument()

    await userEvent.click(knopf)
    expect(await screen.findByRole('status')).toHaveTextContent(
      einsetzen(texte.ausgegeben, { pfad: '/tmp/k.mse-symbol' }),
    )
  })

  it('nennt vor dem Löschen, wer es trägt, und löscht erst auf Ja', async () => {
    const loeschen = vi
      .fn()
      .mockResolvedValue({ sets: ['Kronenset', 'Leerset'], fehler: null })
    const liste = vi
      .fn()
      .mockResolvedValueOnce(mitSymbolen(krone, zacke))
      .mockResolvedValue(mitSymbolen(zacke))
    belegeApi({
      Symbolbibliothek: liste,
      BibliothekssymbolLoeschen: loeschen,
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    const loeschknopf = () =>
      within(screen.getByRole('listitem', { name: 'Krone' })).getByRole(
        'button',
        { name: texte.loeschen },
      )

    await eintrag('Krone')
    await userEvent.click(loeschknopf())

    const frage = screen.getByRole('group', {
      name: einsetzen(texte.loeschenFrage, { name: 'Krone' }),
    })
    expect(frage).toHaveAccessibleDescription(
      einsetzen(texte.loeschenFolgen, { sets: 'Kronenset, Leerset' }),
    )

    await userEvent.click(
      within(frage).getByRole('button', { name: texte.loeschenNein }),
    )
    expect(screen.queryByRole('group')).not.toBeInTheDocument()
    expect(loeschen).not.toHaveBeenCalled()

    await userEvent.click(loeschknopf())
    await userEvent.click(
      within(screen.getByRole('group')).getByRole('button', {
        name: texte.loeschenJa,
      }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('listitem', { name: 'Krone' }),
      ).not.toBeInTheDocument()
    })
    expect(loeschen).toHaveBeenCalledWith('1', true)
    expect(screen.queryByRole('group')).not.toBeInTheDocument()
  })

  it('sagt bei einem unbenutzten, dass kein Set es trägt', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(zacke)),
      MarkenJeObjekt: ohneMarken,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Zacke')

    expect(within(zeile).getByText(texte.unverwendet)).toBeInTheDocument()
    expect(within(zeile).queryByRole('img')).not.toBeInTheDocument()

    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.loeschen }),
    )
    expect(
      screen.getByRole('group', {
        name: einsetzen(texte.loeschenFrage, { name: 'Zacke' }),
      }),
    ).toHaveAccessibleDescription(texte.loeschenOhneFolgen)
  })

  it('zeigt einen Fehler an seiner Zeile und liest dann nicht nach', async () => {
    const liste = vi.fn().mockResolvedValue(mitSymbolen(krone))
    belegeApi({
      Symbolbibliothek: liste,
      BibliothekssymbolEinsetzen: vi.fn().mockResolvedValue({
        datei: '',
        fehler: {
          code: 'nichtGefunden',
          schluessel: '',
          meldung: 'Das Set gibt es nicht mehr.',
          werte: {},
        },
      }),
      MarkenJeObjekt: ohneMarken,
    })

    const eingesetzt = zeigeBibliothek()
    const zeile = await eintrag('Krone')
    await userEvent.click(
      within(zeile).getByRole('button', { name: texte.einsetzen }),
    )

    expect(await within(zeile).findByRole('alert')).toHaveTextContent(
      'Das Set gibt es nicht mehr.',
    )
    expect(eingesetzt).not.toHaveBeenCalled()
    await antwortAbwarten()
    expect(liste).toHaveBeenCalledTimes(1)
  })

  it('trägt Marken und lässt sich nach ihnen filtern', async () => {
    const filtern = vi
      .fn()
      .mockResolvedValue({ kennungen: ['1'], fehler: null })
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone, zacke)),
      MarkenJeObjekt: vi
        .fn()
        .mockResolvedValue({ marken: { '1': ['rund'] }, fehler: null }),
      MarkenFiltern: filtern,
    })

    zeigeBibliothek()
    const zeile = await eintrag('Krone')
    expect(await within(zeile).findByText('rund')).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.objektmarken.werkzeugeUeberschrift,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Nach »rund« filtern' }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('listitem', { name: 'Zacke' }),
      ).not.toBeInTheDocument()
    })
    expect(filtern).toHaveBeenCalledWith('symbol', ['rund'], 'und')
    expect(screen.getByRole('listitem', { name: 'Krone' })).toBeInTheDocument()
  })

  it('sagt, wenn der Markenfilter nichts trifft', async () => {
    belegeApi({
      Symbolbibliothek: vi.fn().mockResolvedValue(mitSymbolen(krone)),
      MarkenJeObjekt: vi
        .fn()
        .mockResolvedValue({ marken: { '1': ['rund'] }, fehler: null }),
      MarkenFiltern: vi.fn().mockResolvedValue({ kennungen: [], fehler: null }),
    })

    zeigeBibliothek()
    await eintrag('Krone')

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.objektmarken.werkzeugeUeberschrift,
      }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: 'Nach »rund« filtern' }),
    )

    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
  })
})
