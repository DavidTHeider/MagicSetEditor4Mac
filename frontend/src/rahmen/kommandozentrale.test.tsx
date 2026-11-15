// **Die Kommandozentrale** — Werkstück U12 von Phase 5.1 (#1423).
import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Kommandozentrale } from './Kommandozentrale'
import { leere, merke } from '../befehlsverlauf'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import type { BefehlDTO } from '../apibefehle'

const texte = kataloge.de.bausteine.kommandozentrale

const lesend: BefehlDTO = {
  name: 'lesen.karten',
  faehigkeit: 'Karten auflisten',
  cli: 'mse karten <set>',
  rest: 'GET /sets/{setID}/karten',
  erklaerung: 'Zählt die Karten eines Sets auf.',
  schreibend: false,
  idempotent: true,
  gefaehrlich: false,
  schema: {
    type: 'object',
    properties: {
      set: { type: 'string', description: 'die Kennung des Sets' },
      hoechstens: { type: 'integer' },
      token: { type: 'string' },
    },
    required: ['set'],
  },
}

const gefaehrlich: BefehlDTO = {
  name: 'schreiben.karte.loeschen',
  faehigkeit: 'Eine Karte löschen',
  cli: 'mse karte <set> loeschen <karte>',
  rest: '',
  erklaerung: 'Löscht eine Karte.',
  schreibend: true,
  idempotent: true,
  gefaehrlich: true,
  schema: {
    type: 'object',
    properties: { karte: { type: 'string' } },
    required: ['karte'],
  },
}

afterEach(() => {
  leere()
})

function zeige(befehl: BefehlDTO, teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi(teile)
  const schliessen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Kommandozentrale befehl={befehl} schliessen={schliessen} />
    </Sprachraum>,
  )

  return { schliessen }
}

describe('die Kommandozentrale', () => {
  it('führt einen lesenden Befehl aus und zeigt sein Ergebnis als JSON und Text', async () => {
    const BefehlAusfuehren = vi.fn().mockResolvedValue({
      name: 'lesen.karten',
      text: '{"karten":[{"name":"Bolt"}]}',
      dauerMs: 12,
      fehler: null,
    })
    zeige(lesend, { BefehlAusfuehren })

    // Kopf: Art, Fähigkeit, die Namen auf den anderen Flächen.
    expect(screen.getByText('mse karten <set>')).toBeInTheDocument()
    expect(screen.getByText('GET /sets/{setID}/karten')).toBeInTheDocument()

    // Ohne Pflichtwert ist der Knopf angehalten, und die Vorschau ist leer.
    const ausfuehren = screen.getByRole('button', { name: texte.ausfuehren })
    expect(ausfuehren).toBeDisabled()
    expect(screen.getByLabelText('set (Pflicht)')).toHaveAccessibleDescription(
      /die Kennung des Sets/,
    )

    await userEvent.type(screen.getByLabelText('set (Pflicht)'), '7"; rm -rf /')
    await userEvent.type(screen.getByLabelText('hoechstens'), '3')
    await userEvent.type(screen.getByLabelText('token'), 'mse_geheim')

    // **Die Vorschau ist, was geschickt wird** — Shellzeichen sind Daten.
    const vorschau = within(
      screen.getByRole('region', { name: texte.vorschau }),
    )
    expect(vorschau.getByText(/"set": "7\\"; rm -rf \/"/)).toBeInTheDocument()
    expect(vorschau.getByText(/"hoechstens": 3/)).toBeInTheDocument()
    // **Sie rollt, also erreicht sie die Tastatur** (Zweitblick über Phase
    // 6.3, G10).
    expect(
      vorschau.getByText(/"hoechstens": 3/).closest('pre'),
    ).toHaveAttribute('tabindex', '0')

    await userEvent.click(ausfuehren)

    const ergebnisbereich = await screen.findByRole('region', {
      name: texte.ergebnis,
    })
    expect(within(ergebnisbereich).getByText(/gelungen/)).toBeInTheDocument()
    expect(BefehlAusfuehren).toHaveBeenCalledWith(
      'lesen.karten',
      '{"set":"7\\"; rm -rf /","hoechstens":3,"token":"mse_geheim"}',
      false,
    )

    const ergebnis = within(
      screen.getByRole('region', { name: texte.ergebnis }),
    )
    expect(ergebnis.getByText(/"name": "Bolt"/)).toBeInTheDocument()
    expect(ergebnis.getByText(/12 ms/)).toBeInTheDocument()

    // Umschalten auf den rohen Text.
    await userEvent.click(ergebnis.getByRole('button', { name: texte.alsText }))
    expect(
      ergebnis.getByText('{"karten":[{"name":"Bolt"}]}'),
    ).toBeInTheDocument()

    // **Der Verlauf nennt den Lauf — ohne das Token.**
    const verlauf = within(screen.getByRole('region', { name: texte.verlauf }))
    expect(verlauf.getByText(/"token":"•••"/)).toBeInTheDocument()
    expect(verlauf.queryByText(/mse_geheim/)).not.toBeInTheDocument()
  })

  it('verlangt bei Gefahr die Bestätigung und nennt einen Fehler am Ort', async () => {
    const BefehlAusfuehren = vi.fn().mockResolvedValue({
      name: 'schreiben.karte.loeschen',
      text: '',
      dauerMs: 1,
      fehler: {
        code: 'nichtGefunden',
        meldung: 'die Karte gibt es nicht',
        feld: '',
        schluessel: 'nichtGefunden',
      },
    })
    zeige(gefaehrlich, { BefehlAusfuehren })

    expect(
      screen.getByText(/nicht ohne Weiteres zurückzunehmen/),
    ).toBeInTheDocument()

    await userEvent.type(screen.getByLabelText('karte (Pflicht)'), '9')
    const ausfuehren = screen.getByRole('button', { name: texte.ausfuehren })
    expect(ausfuehren).toBeDisabled()

    await userEvent.click(
      screen.getByRole('checkbox', { name: texte.bestaetigen }),
    )
    expect(ausfuehren).toBeEnabled()

    await userEvent.click(ausfuehren)

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Karte gibt es nicht',
    )
    expect(BefehlAusfuehren).toHaveBeenCalledWith(
      'schreiben.karte.loeschen',
      '{"karte":"9"}',
      true,
    )
    expect(
      within(screen.getByRole('region', { name: texte.verlauf })).getByText(
        /gescheitert/,
      ),
    ).toBeInTheDocument()
  })

  it('bricht einen laufenden Befehl ab und lädt ihn aus dem Verlauf erneut', async () => {
    let antworten: ((antwort: unknown) => void) | undefined
    const BefehlAusfuehren = vi.fn().mockImplementation(
      () =>
        new Promise((fertig) => {
          antworten = fertig
        }),
    )
    const BefehlAbbrechen = vi.fn().mockResolvedValue({ fehler: null })
    zeige(lesend, { BefehlAusfuehren, BefehlAbbrechen })

    await userEvent.type(screen.getByLabelText('set (Pflicht)'), '7')
    await userEvent.click(
      screen.getByRole('button', { name: texte.ausfuehren }),
    )

    // **Während des Laufs**: Status, Abbrechen, kein zweiter Start.
    expect(await screen.findByRole('status')).toHaveTextContent(texte.laeuft)
    expect(
      screen.getByRole('button', { name: texte.ausfuehren }),
    ).toBeDisabled()

    await userEvent.click(screen.getByRole('button', { name: texte.abbrechen }))
    expect(BefehlAbbrechen).toHaveBeenCalledTimes(1)

    antworten?.({
      name: 'lesen.karten',
      text: '',
      dauerMs: 5,
      fehler: {
        code: 'abgebrochen',
        meldung: 'abgebrochen',
        feld: '',
        schluessel: 'abgebrochen',
      },
    })

    expect(await screen.findByText(/abgebrochen · 5 ms/)).toBeInTheDocument()

    // **Ein Eintrag eines anderen Befehls füllt nichts** — er gehört nicht hierher.
    act(() => {
      merke({
        name: 'lesen.anderes',
        argumente: { set: 'fremd' },
        ausgang: 'gelungen',
        dauerMs: 1,
        zeitpunkt: '2026-09-09T10:00:00Z',
      })
    })
    await userEvent.click(
      screen.getByRole('button', { name: 'lesen.anderes erneut laden' }),
    )
    expect(screen.getByLabelText('set (Pflicht)')).toHaveValue('7')

    // Aus dem Verlauf zurück in die Felder — nach dem Leeren.
    await userEvent.clear(screen.getByLabelText('set (Pflicht)'))
    await userEvent.click(
      screen.getByRole('button', { name: 'lesen.karten erneut laden' }),
    )
    expect(screen.getByLabelText('set (Pflicht)')).toHaveValue('7')
  })

  it('zeigt Auswahl, JSON-Felder und Haken aus dem Schema', async () => {
    const bunt: BefehlDTO = {
      ...lesend,
      name: 'lesen.bunt',
      cli: '',
      rest: '',
      faehigkeit: '',
      schema: {
        type: 'object',
        properties: {
          art: { type: 'string', enum: ['png', 'svg'] },
          zeilen: { type: 'array' },
          trotzdem: { type: 'boolean', description: 'auch dann' },
        },
      },
    }
    zeige(bunt)

    await userEvent.selectOptions(screen.getByLabelText('art'), 'svg')
    await userEvent.type(screen.getByLabelText('zeilen'), '{{')
    expect(screen.getByLabelText('zeilen')).toHaveAccessibleDescription(
      /kein JSON/,
    )
    expect(
      screen.getByRole('button', { name: texte.ausfuehren }),
    ).toBeDisabled()

    await userEvent.clear(screen.getByLabelText('zeilen'))
    await userEvent.type(screen.getByLabelText('zeilen'), '[[[[]]')
    await userEvent.click(screen.getByRole('checkbox', { name: /trotzdem/ }))

    const vorschau = within(
      screen.getByRole('region', { name: texte.vorschau }),
    )
    expect(vorschau.getByText(/"art": "svg"/)).toBeInTheDocument()
    expect(vorschau.getByText(/"trotzdem": true/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: texte.ausfuehren })).toBeEnabled()
  })

  it('sagt, wenn ein Befehl keine Parameter hat, und schließt', async () => {
    const { schliessen } = zeige({ ...lesend, schema: {} })

    expect(screen.getByText(texte.keineParameter)).toBeInTheDocument()
    await userEvent.click(
      screen.getByRole('button', { name: texte.schliessen }),
    )
    expect(schliessen).toHaveBeenCalled()
  })

  // **Der Dialog hält WCAG 2.2 AA** — wie jede Seite in `barrierefreiheit.test.tsx`;
  // er steht nicht in der Seitenliste, deshalb hier.
  it('hält die Barrierefreiheit ein', async () => {
    belegeApi({})
    const { container } = render(
      <Sprachraum sprache="de">
        <Kommandozentrale befehl={gefaehrlich} schliessen={vi.fn()} />
      </Sprachraum>,
    )

    expect(await verstoesse(container)).toEqual([])
  })
})
