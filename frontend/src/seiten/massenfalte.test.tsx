// **Die Massenänderung zeigt zehn Karten und den Rest stückweise** — UXR-06
// (#1914).
//
// # Der Befund
//
// Am Probebestand ist nichts angehakt, und die Messung sah die
// Massenänderung gar nicht. Wer aber „Alle Karten anhaken" wählt — für die
// Seltenheit eines ganzen Sets, für einen Copyright-Satz —, bekommt unter dem
// Arbeitsplatz **je Karte eine Zeile**: erst die Liste der Auswahl, nach der
// Vorschau eine Tabelle mit Bisher und Danach. Bei einem Grundset mit 300
// Karten sind das 300 Zeilen untereinander — gerechnet, nicht gemessen: bei
// rund 30 Punkten je Zeile etwa 9.000 Punkte unter dem Arbeitsplatz, eine
// Rolltiefe weit über den 2,5 aus R4 (`docs/UIREVIEW.md`).
//
// # Was hier gehalten wird
//
// - **Zehn vorn, „10 von 300" und „Weitere 50 zeigen"** — dieselbe Falte wie
//   der Verlauf und die Statistik (`Faltfuss`).
// - **Gerufen und angewendet wird für alle**: die Falte ist eine Ansicht, und
//   die Vorschau gilt für die ganze Auswahl.
// - **Was inzwischen anders dasteht, steht vorn** — sonst läge eine Karte im
//   Konflikt hinter der Falte, und niemand sähe, welche es ist.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO, MassenvorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { feldwerte, fuelleAus, texte, vorschauen, zeige } from './massenprobe'

const worte = kataloge.de.bausteine.listenansicht

const dreihundert: KartenzeileDTO[] = Array.from(
  { length: 300 },
  (_, stelle) => ({
    id: String(stelle + 1),
    nummer: stelle,
    name: `Karte ${String(stelle + 1)}`,
    felder: 1,
  }),
)

function vorschauFuerAlle(): MassenvorschauAntwort {
  return {
    feld: 'typ',
    wert: 'Artefakt',
    zeilen: dreihundert.map((karte) => ({
      karteID: karte.id,
      name: karte.name,
      alt: '',
      gabAlt: false,
      neu: 'Artefakt',
      ausschluss: '',
    })),
  }
}

afterEach(() => {
  raeumeApi()
})

describe('Die Massenänderung über ein ganzes Set', () => {
  it('nennt von 300 angehakten Karten zehn und den Rest stückweise', async () => {
    belegeApi({ ...feldwerte })
    zeige(dreihundert)

    // **Die Auswahl ist die einzige Liste** der Massenänderung.
    expect(await screen.findAllByRole('listitem')).toHaveLength(10)
    expect(
      screen.getByText(
        worte.zahl.replace('{gezeigt}', '10').replace('{gesamt}', '300'),
      ),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: worte.weitere.replace('{anzahl}', '50'),
      }),
    )
    expect(screen.getAllByRole('listitem')).toHaveLength(60)
  })

  it('holt die Vorschau für alle und zeigt davon zehn', async () => {
    const FelderVorschau = vi.fn(() => Promise.resolve(vorschauFuerAlle()))
    belegeApi({ ...feldwerte, FelderVorschau })
    zeige(dreihundert)

    await vorschauen()

    expect(FelderVorschau).toHaveBeenCalledWith(
      '1',
      dreihundert.map((karte) => karte.id),
      'typ',
      'Artefakt',
    )

    // Kopfzeile und zehn Karten.
    expect(within(screen.getByRole('table')).getAllByRole('row')).toHaveLength(
      11,
    )
    expect(
      screen.getByText(
        worte.zahl.replace('{gezeigt}', '10').replace('{gesamt}', '300'),
      ),
    ).toBeInTheDocument()
    expect(
      screen.getByText('300 Karte(n) werden geändert, 0 bleiben.'),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: worte.weitere.replace('{anzahl}', '50'),
      }),
    )
    expect(within(screen.getByRole('table')).getAllByRole('row')).toHaveLength(
      61,
    )
  })

  it('stellt Karten im Konflikt vor die Falte', async () => {
    belegeApi({
      ...feldwerte,
      FelderVorschau: () => Promise.resolve(vorschauFuerAlle()),
      FelderSetzenMitErwartung: () =>
        Promise.resolve({
          betroffen: ['250'],
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung:
              'Das Feld steht nicht mehr auf dem Wert, den die Vorschau gezeigt hat.',
          },
        }),
    })
    zeige(dreihundert)

    await fuelleAus()
    expect(await screen.findByText(texte.konflikt)).toBeInTheDocument()

    const zeilen = within(screen.getByRole('table')).getAllByRole('row')
    expect(zeilen).toHaveLength(11)
    expect(zeilen[1]).toHaveTextContent('Karte 250')
    expect(zeilen[1]).toHaveTextContent(texte.inzwischenGeaendert)
    expect(zeilen[2]).toHaveTextContent('Karte 1')
  })
})
