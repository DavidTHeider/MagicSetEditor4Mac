// **Ein Konflikt ist auflösbar** — FMT-02.AK-03
// (#1691).
//
// Hat jemand die Datei geändert, seit sie aufgeschlagen wurde, gibt es zwei
// Wege: **ihre Fassung laden** (der eigene Text geht verloren) oder **die
// eigene behalten** — dann ist ihre Fassung die neue Basis, die Änderungen
// zeigen, was gegenüber ihr überschrieben wird, und Sichern geht gegen ihren
// Stand.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

const texte = kataloge.de.seiten.vorlagen

const konflikt = {
  stand: '',
  fehler: {
    code: 'konflikt' as const,
    schluessel: '',
    meldung: 'Der Eintrag wurde zwischenzeitlich geändert.',
    werte: {},
  },
}

function zeige() {
  let fremd = false
  const lesen = vi.fn(() =>
    Promise.resolve(
      fremd
        ? {
            inhalt: 'mse version: 2.0.0\n# fremd\n',
            stand: 'fremd',
            fehler: null,
          }
        : { inhalt: 'mse version: 2.0.0\n', stand: 'alt', fehler: null },
    ),
  )
  const schreiben = vi.fn(
    (_paket: string, _datei: string, _inhalt: string, stand: string) => {
      fremd = true

      return Promise.resolve(
        stand === 'fremd' ? { stand: 'neu', fehler: null } : konflikt,
      )
    },
  )

  belegeWerkstatt({
    Paketdateien: () =>
      Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
    PaketdateiLesen: lesen,
    PaketdateiSchreiben: schreiben,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  return { lesen, schreiben }
}

async function tippeUndSichere() {
  await userEvent.click(await screen.findByRole('button', { name: 'style' }))
  const feld = await screen.findByLabelText<HTMLTextAreaElement>(texte.inhalt)
  await userEvent.type(feld, '# meins')
  await userEvent.click(screen.getByRole('button', { name: texte.speichern }))
  await screen.findByText(texte.konflikt)

  return feld
}

describe('Ein Konflikt in der Werkstatt', () => {
  afterEach(raeumeApi)

  it('behält die eigene Fassung auf Wunsch und sichert sie gegen die fremde', async () => {
    const { schreiben } = zeige()
    const feld = await tippeUndSichere()

    await userEvent.click(screen.getByRole('button', { name: texte.behalten }))

    await waitFor(() => {
      expect(screen.queryByText(texte.konflikt)).not.toBeInTheDocument()
    })
    expect(feld.value).toBe('mse version: 2.0.0\n# meins')

    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    expect(await screen.findByText(texte.gespeichert)).toBeInTheDocument()
    expect(schreiben).toHaveBeenLastCalledWith(
      expect.any(String),
      'style',
      'mse version: 2.0.0\n# meins',
      'fremd',
    )
  })

  it('lädt auf Wunsch die fremde Fassung', async () => {
    zeige()
    await tippeUndSichere()

    await userEvent.click(screen.getByRole('button', { name: texte.neuLaden }))

    await waitFor(() => {
      expect(
        screen.getByLabelText<HTMLTextAreaElement>(texte.inhalt).value,
      ).toBe('mse version: 2.0.0\n# fremd\n')
    })
  })
})

describe('Meine behalten, wenn es klemmt', () => {
  afterEach(raeumeApi)

  // zeigeMitLesen belegt die Werkstatt mit zwei Dateien und einem Lesen, das
  // nach dem Konflikt **einmal** der Test in der Hand hat — das von „Meine
  // behalten"; danach liest es wieder die Datei.
  function zeigeMitLesen(spaeter: () => Promise<unknown>) {
    let abfangen = false
    const lesen = vi.fn((_paket: string, datei: string) => {
      if (abfangen && datei === 'style') {
        abfangen = false

        return spaeter()
      }

      return Promise.resolve({
        inhalt: datei === 'style' ? 'mse version: 2.0.0\n' : 'card field:\n',
        stand: 'alt-' + datei,
        fehler: null,
      })
    })

    belegeWerkstatt({
      Paketdateien: () =>
        Promise.resolve({
          dateien: ['style', 'card_fields'],
          beiwerk: [],
          fehler: null,
        }),
      PaketdateiLesen: lesen as never,
      PaketdateiSchreiben: () => {
        abfangen = true

        return Promise.resolve(konflikt)
      },
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )
  }

  it('zeigt einen Lesefehler und lässt den Text stehen', async () => {
    zeigeMitLesen(() =>
      Promise.resolve({
        inhalt: '',
        stand: '',
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: '',
          meldung: 'Die Datei ist fort.',
          werte: {},
        },
      }),
    )

    const feld = await tippeUndSichere()
    await userEvent.click(screen.getByRole('button', { name: texte.behalten }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Datei ist fort.',
    )
    expect(feld.value).toBe('mse version: 2.0.0\n# meins')
  })

  // **Eine späte Antwort trägt sich nicht mehr ein** (§7, Regel 2) — für
  // Erfolg und Fehler. Wer die Datei verlassen und den Entwurf verworfen hat,
  // bekommt ihn nicht zurück, wenn er wiederkommt, und keinen Fehler zu einem
  // Wunsch, den er aufgegeben hat.
  it.each([
    ['eine Fassung', true],
    ['ein Fehler', false],
  ])('trägt %s nicht bei einer anderen Datei ein', async (_art, glueckt) => {
    let fertig: () => void = () => undefined
    zeigeMitLesen(
      () =>
        new Promise((erfuellen, ablehnen) => {
          fertig = () => {
            if (glueckt) {
              erfuellen({ inhalt: '# fremd\n', stand: 'fremd', fehler: null })
            } else {
              ablehnen(new Error('klemmt'))
            }
          }
        }),
    )

    await tippeUndSichere()
    await userEvent.click(screen.getByRole('button', { name: texte.behalten }))
    await userEvent.click(screen.getByRole('button', { name: 'card_fields' }))
    await userEvent.click(
      await screen.findByRole('button', { name: texte.verwerfen }),
    )
    await waitFor(() => {
      expect(
        screen.getByLabelText<HTMLTextAreaElement>(texte.inhalt).value,
      ).toBe('card field:\n')
    })

    fertig()
    await antwortAbwarten()

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByLabelText<HTMLTextAreaElement>(texte.inhalt).value).toBe(
      'card field:\n',
    )

    await userEvent.click(screen.getByRole('button', { name: 'style' }))
    await waitFor(() => {
      expect(
        screen.getByLabelText<HTMLTextAreaElement>(texte.inhalt).value,
      ).toBe('mse version: 2.0.0\n')
    })
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
