// **Marken vergeben und pflegen** — UX-12.AK-02
// (#1684).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Markenpflege, Markenvergabe } from './Markenpflege'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.objektmarken

afterEach(() => {
  raeumeApi()
})

describe('Die Vergabe an die gewählten', () => {
  it('vergibt eine Marke an alle und meldet einen Fehler', async () => {
    const vergeben = vi
      .fn()
      .mockResolvedValueOnce({ fehler: null })
      .mockResolvedValueOnce({
        fehler: {
          code: 'nichtGefunden',
          meldung: 'deck 99 gibt es nicht',
          schluessel: '',
        },
      })
    const geaendert = vi.fn()
    belegeApi({ MarkeVergeben: vergeben })

    render(
      <Sprachraum sprache="de">
        <Markenvergabe
          art="exemplar"
          kennungen={['31', '32']}
          geaendert={geaendert}
        />
      </Sprachraum>,
    )

    const feld = screen.getByRole('textbox', { name: 'Marke für 2 gewählte' })
    await userEvent.type(feld, 'Binder 2')
    await userEvent.click(screen.getByRole('button', { name: texte.vergeben }))

    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
    expect(vergeben).toHaveBeenCalledWith('exemplar', ['31', '32'], 'Binder 2')

    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.vergeben }))
    expect(await screen.findByRole('alert')).toHaveTextContent('gibt es nicht')
  })
})

describe('Die Pflege der Marken', () => {
  it('sagt ohne Marke, dass es nichts zu pflegen gibt', async () => {
    belegeApi({
      Markenkatalog: vi.fn().mockResolvedValue({ marken: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Markenpflege runde={0} geaendert={vi.fn()} />
      </Sprachraum>,
    )

    expect(await screen.findByText(texte.pflegeKeine)).toBeInTheDocument()
  })

  it('benennt um, führt zusammen und entfernt überall nach Rückfrage', async () => {
    const umbenennen = vi
      .fn()
      .mockResolvedValueOnce({
        anzahl: 0,
        fehler: {
          code: 'konflikt',
          meldung: '»handel« gibt es schon',
          schluessel: '',
        },
      })
      .mockResolvedValueOnce({ anzahl: 3, fehler: null })
    const entfernen = vi.fn().mockResolvedValue({ anzahl: 2, fehler: null })
    const geaendert = vi.fn()
    belegeApi({
      Markenkatalog: vi.fn().mockResolvedValue({
        marken: [
          { marke: 'handel', anzahl: 1 },
          { marke: 'tausch', anzahl: 2 },
        ],
        fehler: null,
      }),
      MarkeUmbenennen: umbenennen,
      MarkeUeberallEntfernen: entfernen,
    })

    render(
      <Sprachraum sprache="de">
        <Markenpflege runde={0} geaendert={geaendert} />
      </Sprachraum>,
    )

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: texte.pflegeWahl }),
      'tausch',
    )
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.umbenennenName }),
      'handel',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.umbenennen }),
    )

    // **Die Kollision kommt als Meldung** — ohne ausdrückliches Zusammenführen.
    expect(await screen.findByRole('alert')).toHaveTextContent('gibt es schon')
    expect(umbenennen).toHaveBeenLastCalledWith('tausch', 'handel', false)

    await userEvent.click(
      screen.getByRole('checkbox', { name: texte.zusammenfuehren }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.umbenennen }),
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      '3 Objekte geändert.',
    )
    expect(umbenennen).toHaveBeenLastCalledWith('tausch', 'handel', true)

    // **Überall entfernen fragt** — und Abbrechen tut nichts.
    await userEvent.click(
      screen.getByRole('button', { name: texte.ueberallEntfernen }),
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )
    expect(entfernen).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: texte.ueberallEntfernen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.ueberallLoeschenJa }),
    )

    await waitFor(() => {
      expect(entfernen).toHaveBeenCalledWith('tausch')
    })
    await antwortAbwarten()
    expect(geaendert).toHaveBeenCalledTimes(2)
  })
})

describe('Die Pflege, im Einzelnen', () => {
  // **Ein Wechsel der Marke schließt Rückfrage und Formular** (§7, Regel 1) —
  // sonst bestätigte, wer die Frage für »handel« las, das Entfernen von
  // »tausch« (Befund 2 des Zweitblicks über UX-12).
  it('schließt beim Wechsel der Marke die Rückfrage und leert das Formular', async () => {
    const entfernen = vi.fn().mockResolvedValue({ anzahl: 1, fehler: null })
    belegeApi({
      Markenkatalog: vi.fn().mockResolvedValue({
        marken: [
          { marke: 'handel', anzahl: 1 },
          { marke: 'tausch', anzahl: 2 },
        ],
        fehler: null,
      }),
      MarkeUeberallEntfernen: entfernen,
    })

    render(
      <Sprachraum sprache="de">
        <Markenpflege runde={0} geaendert={vi.fn()} />
      </Sprachraum>,
    )

    const wahl = await screen.findByRole('combobox', { name: texte.pflegeWahl })
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.umbenennenName }),
      'neu',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.ueberallEntfernen }),
    )
    expect(
      await screen.findByRole('button', { name: texte.ueberallLoeschenJa }),
    ).toBeInTheDocument()

    await userEvent.selectOptions(wahl, 'tausch')

    expect(
      screen.queryByRole('button', { name: texte.ueberallLoeschenJa }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('textbox', { name: texte.umbenennenName }),
    ).toHaveValue('')

    await userEvent.click(
      screen.getByRole('button', { name: texte.ueberallEntfernen }),
    )
    await userEvent.click(
      await screen.findByRole('button', { name: texte.ueberallLoeschenJa }),
    )
    await waitFor(() => {
      expect(entfernen).toHaveBeenCalledWith('tausch')
    })
    await antwortAbwarten()
    expect(entfernen).toHaveBeenCalledTimes(1)
  })

  it('benennt mit leerem Namen nicht um', async () => {
    const umbenennen = vi.fn()
    belegeApi({
      Markenkatalog: vi.fn().mockResolvedValue({
        marken: [{ marke: 'rot', anzahl: 1 }],
        fehler: null,
      }),
      MarkeUmbenennen: umbenennen,
    })

    render(
      <Sprachraum sprache="de">
        <Markenpflege runde={0} geaendert={vi.fn()} />
      </Sprachraum>,
    )

    // **Ein leeres Feld benennt nicht um** — auch nicht über das Absenden
    // des Formulars, das die Eingabetaste im Fenster auslöst.
    const feld = await screen.findByRole('textbox', {
      name: texte.umbenennenName,
    })
    fireEvent.submit(feld)
    expect(umbenennen).not.toHaveBeenCalled()
  })
})
