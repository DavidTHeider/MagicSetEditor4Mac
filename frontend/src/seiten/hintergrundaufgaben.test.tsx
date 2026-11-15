// **Hintergrundaufgaben** — Werkstück U11 von Phase 5.1 (#1422).
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { Hintergrundaufgaben } from './Hintergrundaufgaben'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { antwortAbwarten, belegeApi } from '../test/api'
import type { AuftraegeAntwort, AuftragDTO } from '../apiaufmerksamkeit'

const texte = kataloge.de.seiten.start.auftraege

function auftrag(
  id: string,
  stand: AuftragDTO['stand'],
  fortschritt = '',
  grund = '',
): AuftragDTO {
  return {
    id,
    art: 'preise',
    was: `Auftrag ${id}`,
    fortschritt,
    stand,
    grund,
    begonnen: '2026-09-09T10:00:00Z',
    sekunden: 12,
    erledigt: 0,
    gesamt: 0,
    einheit: '',
    abbrechbar: false,
  }
}

function antwort(auftraege: AuftragDTO[]): AuftraegeAntwort {
  return {
    auftraege,
    laufend: auftraege.filter((eines) => eines.stand === 'laeuft').length,
    fehler: null,
  }
}

function hinstellen(antworten: AuftraegeAntwort[]) {
  const Auftraege = vi.fn()
  for (const eine of antworten) {
    Auftraege.mockResolvedValueOnce(eine)
  }
  Auftraege.mockResolvedValue(antworten[antworten.length - 1])
  belegeApi({ Auftraege })

  render(
    <Sprachraum sprache="de">
      <Hintergrundaufgaben stand={0} abstand={20} />
    </Sprachraum>,
  )

  return { Auftraege }
}

describe('Hintergrundaufgaben', () => {
  // **Ein leeres Buch steht nicht da** — UXR-03
  // (#1911).
  // Hier stand ein Bereich mit Überschrift und einem Satz von 160 Zeichen, der
  // erklärte, was erschiene, wenn etwas liefe — auf jeder Startseite, auf der
  // nichts lief. Das Artboard *Main* kennt den Bereich nicht. **Was läuft und
  // was beendet ist, steht weiter da**, sobald das Buch es führt.
  it('steht nicht da, solange das Buch leer ist', async () => {
    const Auftraege = vi
      .fn()
      .mockResolvedValueOnce(antwort([]))
      .mockResolvedValue(antwort([auftrag('1', 'fertig')]))
    belegeApi({ Auftraege })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Hintergrundaufgaben stand={0} abstand={20} />
      </Sprachraum>,
    )
    await antwortAbwarten()

    expect(
      screen.queryByRole('region', { name: texte.titel }),
    ).not.toBeInTheDocument()

    // **Ohne laufenden Auftrag wird nicht gefragt** — einmal, beim Aufschlagen.
    await new Promise((weiter) => setTimeout(weiter, 100))
    expect(Auftraege).toHaveBeenCalledTimes(1)

    // **Und jemand hört zu** (§3): ein neuer Stand fragt wieder, und das Buch,
    // das jetzt etwas führt, steht da.
    rerender(
      <Sprachraum sprache="de">
        <Hintergrundaufgaben stand={1} abstand={20} />
      </Sprachraum>,
    )
    expect(
      await screen.findByRole('region', { name: texte.titel }),
    ).toBeInTheDocument()
  })

  it('fragt nach, solange etwas läuft, und hört danach auf', async () => {
    const { Auftraege } = hinstellen([
      antwort([auftrag('1', 'laeuft', '3 von 9')]),
      antwort([auftrag('1', 'laeuft', '7 von 9')]),
      antwort([auftrag('1', 'fertig', '9 von 9')]),
    ])

    const bereich = await screen.findByRole('region', { name: texte.titel })
    expect(within(bereich).getByRole('status')).toHaveTextContent(
      '1 Aufgabe läuft',
    )
    expect(within(bereich).getByText('3 von 9')).toBeInTheDocument()
    expect(within(bereich).getByText('12 s')).toBeInTheDocument()

    // Die Wirkung der Nachfrage: der Stand wechselt auf fertig — im **selben**
    // Bereich, denn die Liste bleibt beim Nachfragen stehen.
    expect(await within(bereich).findByText('fertig')).toBeInTheDocument()
    expect(within(bereich).queryByRole('status')).not.toBeInTheDocument()
    expect(screen.getByRole('region', { name: texte.titel })).toBe(bereich)

    // **Erst zur Ruhe kommen lassen, dann zählen.** Ein Tick der Uhr, der vor
    // dem Aufräumen des Effekts schon gefeuert hatte, fragt noch einmal — das
    // ist kein Nachfragen mehr, sondern die Antwort auf den letzten Tick. Bis
    // zum 2026-09-09 zählte der Test ab dem Erscheinen von »fertig« und war
    // unter Last rot (4 statt 3), bei freier Maschine grün.
    await act(async () => {
      await new Promise((weiter) => setTimeout(weiter, 100))
    })
    const gefragt = Auftraege.mock.calls.length
    await act(async () => {
      await new Promise((weiter) => setTimeout(weiter, 100))
    })
    expect(Auftraege).toHaveBeenCalledTimes(gefragt)
  })

  it('nennt den Grund eines gescheiterten Auftrags', async () => {
    hinstellen([antwort([auftrag('2', 'gescheitert', '', 'kein Netz')])])

    const bereich = await screen.findByRole('region', { name: texte.titel })
    expect(within(bereich).getByText('gescheitert')).toBeInTheDocument()
    expect(within(bereich).getByText('kein Netz')).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Bindung', async () => {
    hinstellen([
      {
        auftraege: [],
        laufend: 0,
        fehler: {
          code: 'unbekannt',
          meldung: 'Buch fehlt',
          feld: '',
          schluessel: '',
        },
      },
    ])

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Buch fehlt')
    })
  })

  // **Die Art in der Sprache des Fensters, der Gegenstand daneben** —
  // FMT-03 (#1692). Bis dahin stand hier ein deutscher Satz aus Go, auch im
  // englischen Fenster.
  it('nennt die Art in der Sprache des Fensters und den Gegenstand', async () => {
    hinstellen([
      antwort([
        { ...auftrag('1', 'fertig'), art: 'bogen', was: 'Mein Bogen.pdf' },
        { ...auftrag('2', 'fertig'), was: '' },
        { ...auftrag('3', 'fertig'), art: 'unbekannt', was: '' },
      ]),
    ])

    const bereich = await screen.findByRole('region', { name: texte.titel })
    expect(
      within(bereich).getByText(`${texte.arten.bogen} · Mein Bogen.pdf`),
    ).toBeInTheDocument()
    expect(within(bereich).getByText(texte.arten.preise)).toBeInTheDocument()
    expect(within(bereich).getByText('unbekannt')).toBeInTheDocument()
  })

  it('hält einen Datei-Lauf an und zeigt ihn danach als abgebrochen', async () => {
    const laufend = {
      ...auftrag('1', 'laeuft'),
      art: 'bogen',
      erledigt: 2,
      gesamt: 5,
      einheit: 'karte',
      abbrechbar: true,
    }
    const AuftragAbbrechen = vi.fn(() => Promise.resolve(antwort([laufend])))
    const { Auftraege } = hinstellen([
      antwort([laufend]),
      antwort([{ ...laufend, stand: 'abgebrochen', abbrechbar: false }]),
    ])
    belegeApi({ Auftraege, AuftragAbbrechen })

    const bereich = await screen.findByRole('region', { name: texte.titel })
    expect(within(bereich).getByText('2 von 5 Karten')).toBeInTheDocument()

    fireEvent.click(
      within(bereich).getByRole('button', {
        name: `${texte.arten.bogen} abbrechen`,
      }),
    )
    expect(AuftragAbbrechen).toHaveBeenCalledWith('1')

    expect(
      await within(bereich).findByText(texte.staende.abgebrochen),
    ).toBeInTheDocument()
    expect(within(bereich).queryByRole('button')).not.toBeInTheDocument()
  })

  it('meldet ein gescheitertes Anhalten', async () => {
    const laufend = { ...auftrag('1', 'laeuft'), abbrechbar: true }
    const { Auftraege } = hinstellen([antwort([laufend])])
    belegeApi({
      Auftraege,
      AuftragAbbrechen: () =>
        Promise.resolve({
          auftraege: [],
          laufend: 0,
          fehler: {
            code: 'konflikt',
            meldung: 'läuft zu Ende',
            feld: '',
            schluessel: '',
          },
        }),
    })

    const bereich = await screen.findByRole('region', { name: texte.titel })
    fireEvent.click(within(bereich).getByRole('button'))

    expect(await screen.findByRole('alert')).toHaveTextContent('läuft zu Ende')
  })
})
