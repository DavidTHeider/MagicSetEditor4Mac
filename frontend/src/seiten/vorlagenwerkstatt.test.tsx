// **Die Läufe der Vorlagenwerkstatt** — Werkstück U7 von Phase 5.1
// (#1418).
//
// `vorlagen.test.tsx` fährt den Weg „Datei, ändern, speichern"; hier stehen
// die Läufe des Arbeitsplatzes: die Rückfrage vor dem Verwerfen, der Sprung
// zur Definition in eine andere Datei, die Testkarte je Paket, der Befund
// mit seiner Stelle, die Herkunft des Namens unter dem Zeiger und der Diff.
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { waehleZeile } from '../test/suchwahl'
import { paketDTO } from '../test/vorlagenprobe'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

import type { Api } from '../api'

const texte = kataloge.de.seiten.vorlagen

function zeige(eigene: Partial<Api> = {}) {
  belegeWerkstatt({
    Paketdateien: () =>
      Promise.resolve({
        dateien: ['game', 'style'],
        beiwerk: [],
        fehler: null,
      }),
    PaketdateiLesen: (_paket: string, datei: string) =>
      Promise.resolve({
        inhalt:
          datei === 'game'
            ? 'card field:\n\tname: title\nscript:\n\tx := card.title\n'
            : 'card style:\n\ttitle:\n\t\tleft: 1\n',
        stand: 'alt-' + datei,
        fehler: null,
      }),
    Paketsymbole: () =>
      Promise.resolve({
        symbole: [
          {
            name: 'title',
            anzeigename: 'title',
            art: 'set field',
            datei: 'game',
            zeile: 2,
            spalte: 2,
          },
        ],
        fehler: null,
      }),
    ...eigene,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )
}

// zweig gibt den Listeneintrag eines Pakets im Baum — für Dateien, die es in
// mehreren Paketen gibt.
async function zweig(paket: string): Promise<HTMLElement> {
  const knopf = await screen.findByRole('button', { name: paket })
  const eintrag = knopf.closest('li')

  if (eintrag === null) {
    throw new Error(`kein Zweig für ${paket}`)
  }

  return eintrag
}

async function dateiOeffnen(name: string) {
  await userEvent.click(await screen.findByRole('button', { name }))

  return screen.findByLabelText<HTMLTextAreaElement>(texte.inhalt)
}

describe('Die Vorlagenwerkstatt', () => {
  afterEach(raeumeApi)

  it('fragt vor dem Verwerfen und bleibt auf Wunsch', async () => {
    zeige()

    const feld = await dateiOeffnen('game')
    await userEvent.type(feld, 'x')

    await userEvent.click(screen.getByRole('button', { name: 'style' }))

    const frage = await screen.findByRole('group', { name: texte.verwerfen })
    expect(frage).toHaveTextContent(
      'Die offene Datei hat ungesicherte Änderungen. Sollen sie verworfen und style geöffnet werden?',
    )

    await userEvent.click(
      within(frage).getByRole('button', { name: texte.hierBleiben }),
    )

    expect(
      screen.queryByRole('group', { name: texte.verwerfen }),
    ).not.toBeInTheDocument()
    expect(screen.getByLabelText(texte.inhalt)).toHaveValue(
      'card field:\n\tname: title\nscript:\n\tx := card.title\nx',
    )

    await userEvent.click(screen.getByRole('button', { name: 'style' }))
    await userEvent.click(
      within(
        await screen.findByRole('group', { name: texte.verwerfen }),
      ).getByRole('button', {
        name: texte.verwerfen,
      }),
    )

    await waitFor(() => {
      expect(screen.getByLabelText(texte.inhalt)).toHaveValue(
        'card style:\n\ttitle:\n\t\tleft: 1\n',
      )
    })
  })

  it('sagt zum Namen unter dem Zeiger, woher er kommt, und springt zur Definition in die andere Datei', async () => {
    const Symbolverwendungen = vi.fn(() =>
      Promise.resolve({
        verwendungen: [{ datei: 'game', zeile: 4, text: 'x := card.title' }],
        fehler: null,
      }),
    )
    zeige({ Symbolverwendungen })

    const feld = await dateiOeffnen('style')

    // **Der Zeiger auf `title`** in der Stildatei — Zeile 2, im Namen.
    const stelle = 'card style:\n\tt'.length
    feld.setSelectionRange(stelle, stelle)
    fireEvent.select(feld)

    expect(
      await screen.findByText('Setfeld in game, Zeile 2'),
    ).toBeInTheDocument()

    await waitFor(() => {
      expect(Symbolverwendungen).toHaveBeenCalledWith(
        '/pakete/probe.mse-style',
        'title',
      )
    })
    expect(
      await screen.findByRole('button', { name: 'game:4 x := card.title' }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.zurDefinition }),
    )

    // **Die andere Datei geht auf, und der Zeiger steht an der Definition.**
    await waitFor(() => {
      expect(screen.getByLabelText(texte.inhalt)).toHaveValue(
        'card field:\n\tname: title\nscript:\n\tx := card.title\n',
      )
    })

    const neuesFeld = screen.getByLabelText<HTMLTextAreaElement>(texte.inhalt)
    await waitFor(() => {
      expect(neuesFeld.selectionStart).toBe('card field:\n\t'.length)
    })
    expect(neuesFeld).toHaveFocus()

    // **Eine Verwendung in derselben Datei** springt ohne Dateiwechsel.
    //
    // **Gewartet wird auf die Antwort, nicht auf den Ruf** (§3): die Liste der
    // vorigen Suche steht noch, bis die neue lädt — ein Klick in diesem
    // Augenblick träfe einen Knopf, der gerade abgehängt wird.
    fireEvent.select(neuesFeld)
    await waitFor(() => {
      expect(screen.queryByText(texte.verwendungLaden)).toBeNull()
      expect(
        screen.getByRole('button', { name: 'game:4 x := card.title' }),
      ).toBeInTheDocument()
    })
    await userEvent.click(
      screen.getByRole('button', { name: 'game:4 x := card.title' }),
    )

    await waitFor(() => {
      expect(neuesFeld.selectionStart).toBe(
        'card field:\n\tname: title\nscript:\n'.length,
      )
    })
  })

  it('springt zur Stelle eines Befunds und markiert die Vorschau als veraltet', async () => {
    zeige({
      PaketdateiPruefen: (inhalt: string) =>
        Promise.resolve(
          inhalt.includes('(')
            ? {
                stelle: { zeile: 4, spalte: 2 },
                // **Wie Go antwortet** — Schlüssel und beide Werte, gemessen
                // an `service.PaketSkriptfehler("script: f((1)")`.
                meldung: 'diese `(` wird nicht geschlossen — `)` fehlt',
                schluessel: 'klammerOffen',
                werte: ['`(`', '`)`'],
              }
            : { stelle: null, meldung: '', schluessel: '', werte: [] },
        ),
    })

    const feld = await dateiOeffnen('game')
    expect(await screen.findByText(texte.diagnoseFrei)).toBeInTheDocument()

    await userEvent.type(feld, '(')

    expect(
      await screen.findByText('diese `(` wird nicht geschlossen — `)` fehlt'),
    ).toBeInTheDocument()
    expect(screen.getByText('Zeile 4, Spalte 2')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.zurStelle }))

    expect(feld.selectionStart).toBe(
      'card field:\n\tname: title\nscript:\n\t'.length,
    )
  })

  it('hält die Testkarte je Paket und wechselt sie beim Paketwechsel', async () => {
    const Vorlagenvorschau = vi.fn<Api['Vorlagenvorschau']>(() =>
      Promise.resolve({
        vorschau: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'keinSetMitStil',
          meldung: 'kein Set',
        },
      }),
    )
    zeige({
      Pakete: () =>
        Promise.resolve({
          pakete: [paketDTO('probe.mse-style'), paketDTO('zweit.mse-style')],
          befunde: [],
          fehler: null,
        }),
      Probekarten: (paket: string) =>
        Promise.resolve({
          karten:
            paket === '/pakete/probe.mse-style'
              ? [{ id: '7', name: 'Langtext', nummer: 1 }]
              : [{ id: '9', name: 'Fremd', nummer: 1 }],
          fehler: null,
        }),
      Skriptnamen: () => Promise.resolve({ namen: ['to_upper'], fehler: null }),
      Vorlagenvorschau,
    })

    await userEvent.click(
      within(await zweig('probe.mse-style')).getByRole('button', {
        name: 'game',
      }),
    )
    await screen.findByLabelText(texte.inhalt)

    // **Über die Auswahl mit Suche** (UXR-F1, #2007): öffnen, Zeile klicken.
    await waehleZeile(await screen.findByLabelText(texte.testkarte), 'Langtext')

    await waitFor(() => {
      expect(Vorlagenvorschau).toHaveBeenLastCalledWith(
        '/pakete/probe.mse-style',
        'game',
        'card field:\n\tname: title\nscript:\n\tx := card.title\n',
        1,
        '7',
      )
    })

    // **Das andere Paket hat seine eigene Wahl** — die Vorgabe.
    await userEvent.click(
      within(await zweig('zweit.mse-style')).getByRole('button', {
        name: 'game',
      }),
    )

    // **Die Vorgabe steht im Feld** — der leere Wert hat einen Namen.
    await waitFor(() => {
      expect(screen.getByLabelText(texte.testkarte)).toHaveValue(
        texte.testkarteVorgabe,
      )
    })

    // **Und die Zeilen sind die des anderen Pakets**, sobald das Feld offen
    // ist.
    await userEvent.click(screen.getByLabelText(texte.testkarte))
    expect(screen.getByRole('option', { name: 'Fremd' })).toBeInTheDocument()
    expect(screen.queryByRole('option', { name: 'Langtext' })).toBeNull()
  })

  it('zeigt den Diff vor dem Speichern und meldet einen Fehler der Paketliste', async () => {
    zeige()

    const feld = await dateiOeffnen('game')
    expect(
      screen.getByRole('button', { name: texte.aenderungenZeigen }),
    ).toBeDisabled()

    await userEvent.type(feld, 'neu')
    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungenZeigen }),
    )

    expect(screen.getByText('1 Zeile(n) neu, 0 entfernt.')).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.aenderungenVerbergen }),
    )

    expect(
      screen.queryByText('1 Zeile(n) neu, 0 entfernt.'),
    ).not.toBeInTheDocument()
  })

  it('lädt nach einem Konflikt auf Wunsch neu', async () => {
    let stand = 'alt'
    zeige({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: 'von ' + stand + '\n', stand, fehler: null }),
      PaketdateiSchreiben: () =>
        Promise.resolve({
          stand: '',
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'geändert',
          },
        }),
    })

    const feld = await dateiOeffnen('game')
    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    expect(await screen.findByRole('alert')).toHaveTextContent(texte.konflikt)

    stand = 'fremd'
    await userEvent.click(screen.getByRole('button', { name: texte.neuLaden }))

    await waitFor(() => {
      expect(screen.getByLabelText(texte.inhalt)).toHaveValue('von fremd\n')
    })
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('nennt den Fehler der Paketliste', async () => {
    zeige({
      Pakete: () =>
        Promise.resolve({
          pakete: [],
          befunde: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'kein Verzeichnis',
          },
        }),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })
})
