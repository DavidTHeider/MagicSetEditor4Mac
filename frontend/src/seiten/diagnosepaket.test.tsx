// **Das Diagnosepaket im Fenster** — Werkstück O10 von Phase 5.2
// (#1434).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Diagnosepaket } from './Diagnosepaket'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { DiagnosepaketAntwort } from '../apidiagnosepaket'

const texte = kataloge.de.seiten.technik.paket

afterEach(raeumeApi)

// vorschau ist eine vollständige Antwort mit zwei Teilen.
function vorschau(
  teil: Partial<DiagnosepaketAntwort> = {},
): DiagnosepaketAntwort {
  return {
    teile: [
      {
        teil: 'fassung',
        datei: 'fassung.txt',
        zeilen: 3,
        bytes: 64,
        grund: '',
      },
      {
        teil: 'protokoll',
        datei: 'protokoll.txt',
        zeilen: 0,
        bytes: 0,
        grund: 'Die Protokollzeilen stehen hier nicht.',
      },
    ],
    bytes: 64,
    grenze: 8388608,
    von: '2026-09-10T05:00:00+02:00',
    bis: '2026-09-10T06:00:00+02:00',
    ohne: [],
    fehler: null,
    ...teil,
  }
}

function zeige(antwort = vorschau(), weiteres = {}) {
  belegeApi({
    Diagnosepaket: vi.fn(() => Promise.resolve(antwort)),
    ...weiteres,
  })

  render(
    <Sprachraum sprache="de">
      <Diagnosepaket />
    </Sprachraum>,
  )
}

describe('Das Diagnosepaket', () => {
  // **Vor dem Zusammenstellen steht keine Vorschau** — und kein Sichern.
  it('zeigt zuerst nur den Knopf zum Zusammenstellen', () => {
    zeige()

    expect(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.sichern }),
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('nennt jeden Teil mit Größe, den Zeitraum und die Grenze', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )

    expect(await screen.findByRole('table')).toBeInTheDocument()
    expect(screen.getByText('fassung.txt')).toBeInTheDocument()

    // **Der Teil steht unter seinem Namen und nicht unter seinem Bezeichner**
    // — die Sichtprüfung fand die Spalte sonst als Wiederholung der Datei
    // daneben.
    expect(screen.getByText(texte.teile.fassung)).toBeInTheDocument()

    // **Und die Größe in einer Einheit, die jemand einordnen kann** — mit
    // dem Komma der eingestellten Sprache: „64 B" und „8,0 MB", nicht
    // „8388608".
    expect(screen.getByText('64 B')).toBeInTheDocument()

    expect(
      screen.getByText(
        texte.groesse.replace('{bytes}', '64 B').replace('{grenze}', '8,0 MB'),
      ),
    ).toBeInTheDocument()

    expect(
      screen.getByText(
        texte.zeitraum
          .replace('{von}', '2026-09-10T05:00:00+02:00')
          .replace('{bis}', '2026-09-10T06:00:00+02:00'),
      ),
    ).toBeInTheDocument()
  })

  // **Ein ausgelassener Teil wird genannt** — sonst vermisst ihn niemand.
  it('nennt einen ausgelassenen Teil samt Grund', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )

    expect(
      await screen.findByText(/Die Protokollzeilen stehen hier nicht/),
    ).toBeInTheDocument()
  })

  // **Ohne Protokollzeilen steht kein Zeitraum da** — „von leer bis leer"
  // wäre eine Zeile ohne Aussage.
  it('lässt den Zeitraum weg, wenn keine Zeilen dabei sind', async () => {
    zeige(vorschau({ von: '', bis: '' }))

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )

    await screen.findByRole('table')
    expect(screen.queryByText(/reichen von/)).not.toBeInTheDocument()
  })

  // **Der Löschknopf kommt erst nach dem Sichern** — vorher gibt es nichts
  // zu löschen.
  it('bietet das Löschen erst nach dem Sichern an', async () => {
    const sichern = vi.fn(() => Promise.resolve({ fehler: null }))
    const loeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige(vorschau(), {
      DiagnosepaketSichern: sichern,
      DiagnosepaketLoeschen: loeschen,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )
    await screen.findByRole('table')

    expect(
      screen.queryByRole('button', { name: texte.loeschen }),
    ).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

    await userEvent.click(
      await screen.findByRole('button', { name: texte.loeschen }),
    )

    expect(sichern).toHaveBeenCalledTimes(1)
    expect(loeschen).toHaveBeenCalledTimes(1)

    // **Und danach ist er wieder weg** — es gibt nichts mehr zu löschen.
    expect(
      screen.queryByRole('button', { name: texte.loeschen }),
    ).not.toBeInTheDocument()
  })

  // **Ein Teil, den dieses Fenster nicht kennt, steht im Wortlaut da** — er
  // ist immer noch eine Auskunft, eine leere Zelle wäre keine. Und seine
  // Größe kommt in Megabyte, sobald sie so groß ist.
  it('zeigt einen unbekannten Teil im Wortlaut und große Zahlen in MB', async () => {
    zeige(
      vorschau({
        teile: [
          {
            teil: 'gibtsnicht',
            datei: 'gibtsnicht.txt',
            zeilen: 1,
            bytes: 3 * 1024 * 1024,
            grund: 'Diese Quelle ließ sich nicht lesen.',
          },
          {
            teil: 'pakete',
            datei: 'pakete.txt',
            zeilen: 40,
            bytes: 2048,
            grund: '',
          },
        ],
        bytes: 3 * 1024 * 1024,
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )

    await screen.findByRole('table')
    expect(screen.getByText('gibtsnicht')).toBeInTheDocument()
    expect(screen.getByText('3,0 MB')).toBeInTheDocument()
    expect(screen.getByText('2,0 KB')).toBeInTheDocument()

    // **Und der ausgelassene Teil trägt denselben Wortlaut.**
    expect(screen.getByText(/Ausgelassen — gibtsnicht/)).toBeInTheDocument()
  })

  it('zeigt einen Fehler als Meldung', async () => {
    zeige(
      vorschau({
        fehler: {
          code: 'validierung',
          schluessel: 'fehler.validierung',
          meldung: 'Das Diagnosepaket wäre zu groß.',
          werte: {},
        },
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.zusammenstellen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })
})
