// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// **Am Bauteil und nicht an der Seite**: dass das Gebiet „Bestand" es zeigt,
// prüft `appruecksetzung.test.tsx`; hier steht, **was** es zeigt und was eine
// Art an der Gestalt ändert.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Ruecksetzung } from './Ruecksetzung'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { RuecksetzungsvorschauAntwort } from '../apiruecksetzung'

const worte = kataloge.de.seiten.ruecksetzung

afterEach(() => {
  raeumeApi()
})

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Ruecksetzung gewechselt={vi.fn()} veraendert={vi.fn()} />
    </Sprachraum>,
  )
}

// satz holt einen Eintrag aus einer Marken-Tafel — und scheitert laut, wenn es
// ihn nicht gibt (§10: ein `?? ''` träfe immer zu).
function satz(tafel: Record<string, string>, schluessel: string): string {
  const wert = tafel[schluessel]

  if (wert === undefined) {
    throw new Error(`der Katalog kennt ${schluessel} nicht`)
  }

  return wert
}

function vorschau(
  teil: Partial<RuecksetzungsvorschauAntwort> = {},
): RuecksetzungsvorschauAntwort {
  return {
    art: 'profil',
    betroffen: [
      { was: 'sets', anzahl: 3, namen: ['Eins', 'Zwei'] },
      { was: 'karten', anzahl: 107, namen: [] },
    ],
    hindernisse: [],
    bereit: true,
    nameNoetig: true,
    nachweisNoetig: false,
    sicherungEmpfohlen: true,
    revision: 'marke-der-vorschau',
    fehler: null,
    ...teil,
  }
}

describe('Zurücksetzen', () => {
  it('zeigt die sechs Arten in der Reihe ihrer Schwere', async () => {
    belegeApi({})

    zeige()

    // **Gewartet wird auf die Wirkung, nicht auf den Kasten** (§3): das
    // Auswahlfeld steht sofort da, seine Einträge kommen mit der Antwort.
    await screen.findByRole('option', { name: satz(worte.arten, 'layout') })

    const wahl = screen.getByRole('combobox', { name: worte.artWaehlen })
    const arten = [...wahl.querySelectorAll('option')].map((o) => o.textContent)

    expect(arten).toEqual([
      satz(worte.arten, 'layout'),
      satz(worte.arten, 'profileinstellungen'),
      satz(worte.arten, 'nutzereinstellungen'),
      satz(worte.arten, 'profil'),
      satz(worte.arten, 'nutzer'),
      satz(worte.arten, 'anwendung'),
    ])
  })

  it('zählt und benennt, was betroffen wäre', async () => {
    const nachsehen = vi.fn().mockResolvedValue(vorschau())
    belegeApi({ Ruecksetzungsvorschau: nachsehen })

    const nutzer = userEvent.setup()

    zeige()

    // **Erst die Einträge, dann die Wahl** (§3): die Reihe kommt von der
    // Go-Seite, und das Auswahlfeld steht schon davor da.
    await screen.findByRole('option', { name: satz(worte.arten, 'profil') })

    await nutzer.selectOptions(
      screen.getByRole('combobox', { name: worte.artWaehlen }),
      'profil',
    )
    await nutzer.click(screen.getByRole('button', { name: worte.nachsehen }))

    expect(
      await screen.findByText(/Sets: 3/, { exact: false }),
    ).toBeInTheDocument()
    expect(screen.getByText(/Eins, Zwei/, { exact: false })).toBeInTheDocument()
    expect(screen.getByText(worte.sicherung)).toBeInTheDocument()

    // **Die Art reist mit** — die Vorschau gehört zu der, die dasteht.
    expect(nachsehen).toHaveBeenCalledWith('profil', '', '')
  })

  it('verlangt den Namen, bevor der Knopf etwas tut', async () => {
    const tun = vi.fn().mockResolvedValue({
      art: 'profil',
      zurueckgesetzt: 0,
      beiseite: '',
      fehler: null,
    })

    belegeApi({
      Ruecksetzungsvorschau: () => Promise.resolve(vorschau()),
      Zuruecksetzen: tun,
    })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    const knopf = await screen.findByRole('button', { name: worte.tun })
    expect(knopf).toBeDisabled()

    await nutzer.type(
      screen.getByRole('textbox', { name: worte.nameTippen }),
      'Meins',
    )

    await waitFor(() => {
      expect(knopf).toBeEnabled()
    })

    await nutzer.click(knopf)

    await waitFor(() => {
      // **Die Revision der Vorschau geht mit** (Befund 12 des Zweitblicks
      // über Phase 5.2): ausgeführt wird, was jemand gesehen hat.
      expect(tun).toHaveBeenCalledWith(
        'profil',
        '',
        '',
        'Meins',
        '',
        'marke-der-vorschau',
      )
    })
  })

  it('nennt das Hindernis und bietet dann gar nichts an', async () => {
    belegeApi({
      Ruecksetzungsvorschau: () =>
        Promise.resolve(
          vorschau({
            bereit: false,
            hindernisse: [{ art: 'laufenderAuftrag', stelle: 'die Ausfuhr' }],
          }),
        ),
    })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    expect(
      await screen.findByText(satz(worte.hindernisse, 'laufenderAuftrag'), {
        exact: false,
      }),
    ).toBeInTheDocument()
    expect(screen.getByText(/die Ausfuhr/)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: worte.tun })).toBeNull()
  })

  it('räumt die Vorschau ab, wenn die Art wechselt', async () => {
    belegeApi({ Ruecksetzungsvorschau: () => Promise.resolve(vorschau()) })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )
    expect(
      await screen.findByText(/Sets: 3/, { exact: false }),
    ).toBeInTheDocument()

    await nutzer.selectOptions(
      screen.getByRole('combobox', { name: worte.artWaehlen }),
      'layout',
    )

    // **§7, Regel 1**: die Vorschau beschreibt eine andere Rücksetzung.
    await waitFor(() => {
      expect(screen.queryByText(/Sets: 3/, { exact: false })).toBeNull()
    })
  })

  it('meldet einen Fehler statt einer leeren Vorschau', async () => {
    belegeApi({
      Ruecksetzungsvorschau: () =>
        Promise.reject(new Error('der Bestand ist zu')),
    })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Bestand ist zu',
    )
  })

  it('meldet, wenn schon die Reihe der Arten nicht kommt', async () => {
    belegeApi({
      Ruecksetzungsarten: () => Promise.reject(new Error('keine Verbindung')),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'keine Verbindung',
    )
  })

  it('zeigt eine unbekannte Marke als Marke — statt nichts', async () => {
    belegeApi({
      Ruecksetzungsvorschau: () =>
        Promise.resolve(
          vorschau({
            betroffen: [{ was: 'kisten', anzahl: 2, namen: [] }],
            bereit: false,
            hindernisse: [{ art: 'mondphase', stelle: '' }],
          }),
        ),
    })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )

    // **Der Katalog kennt sie nicht, und trotzdem steht etwas da** (§8): eine
    // Fläche, die eine neue Marke verschluckt, zeigt eine leere Zeile.
    expect(await screen.findByText(/kisten: 2/)).toBeInTheDocument()
    expect(screen.getByText('mondphase')).toBeInTheDocument()
  })

  it('zeigt eine neue Art als Marke, statt einen leeren Eintrag', async () => {
    // **Die Reihe kommt von der Go-Seite** — wächst sie, bevor der Katalog
    // nachzieht, steht die Marke da. Ein leerer Eintrag wäre eine Wahl, die
    // niemand treffen kann.
    belegeApi({ Ruecksetzungsarten: () => Promise.resolve(['mondphase']) })

    zeige()

    expect(
      await screen.findByRole('option', { name: 'mondphase' }),
    ).toBeInTheDocument()
  })

  it('bietet nichts an, solange keine Art dasteht', async () => {
    belegeApi({ Ruecksetzungsarten: () => Promise.resolve([]) })

    zeige()

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: worte.nachsehen }),
      ).toBeDisabled()
    })
  })

  it('meldet einen Fehler beim Zurücksetzen und lässt die Vorschau stehen', async () => {
    belegeApi({
      Ruecksetzungsvorschau: () => Promise.resolve(vorschau()),
      Zuruecksetzen: () => Promise.reject(new Error('das ging nicht')),
    })

    const nutzer = userEvent.setup()

    zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: worte.nachsehen }),
    )
    await nutzer.type(
      await screen.findByRole('textbox', { name: worte.nameTippen }),
      'Meins',
    )
    await nutzer.click(screen.getByRole('button', { name: worte.tun }))

    expect(await screen.findByRole('alert')).toHaveTextContent('das ging nicht')

    // **§7, Regel 3**: eine gescheiterte Rücksetzung räumt die Vorschau nicht
    // ab — sonst müsste jemand noch einmal nachsehen, um es zu wiederholen.
    expect(screen.getByText(/Sets: 3/, { exact: false })).toBeInTheDocument()
  })
})
