// **Die Paketwerkstatt am Fenster** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Was hier geprüft wird
//
// Der Weg: Paket wählen, Datei aufschlagen, ändern, speichern. Und die zwei
// Zusagen, die ihn tragen — der **Stand** geht mit, und ein **Konflikt** sagt,
// was zu tun ist, statt die fremde Änderung zu überschreiben.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'

import type { Feldapi } from '../apifeld'
import { Vorlagen } from './Vorlagen'

const texte = kataloge.de.seiten.vorlagen

// Schreibweg ist die Gestalt der Bindung, die diese Seite zum Speichern ruft.
type Schreibweg = Feldapi['PaketdateiSchreiben']

// werkstatt belegt die Bindungen der Seite — das Schreiben eigens.
function werkstatt(schreiben: Schreibweg) {
  belegeWerkstatt({ PaketdateiSchreiben: schreiben })
}

// bisZurDatei fährt den Weg bis zur aufgeschlagenen Datei.
async function bisZurDatei() {
  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  await userEvent.click(await screen.findByRole('button', { name: 'style' }))

  return screen.findByLabelText(texte.inhalt)
}

describe('Die Paketwerkstatt', () => {
  afterEach(raeumeApi)

  it('schlägt eine Datei auf und speichert sie mit ihrem Stand', async () => {
    const geschrieben = vi.fn<Schreibweg>(() =>
      Promise.resolve({ stand: 'neu', fehler: null }),
    )

    werkstatt(geschrieben)

    const feld = await bisZurDatei()

    expect(feld).toHaveValue('mse version: 2.0.0\n')

    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    await waitFor(() => {
      expect(geschrieben).toHaveBeenCalledWith(
        '/pakete/probe.mse-style',
        'style',
        'mse version: 2.0.0\nx',
        'alt',
      )
    })

    expect(await screen.findByText(texte.gespeichert)).toBeInTheDocument()
  })

  // **Der zurückgegebene Stand trägt den nächsten Zug.** Ohne diese Zusicherung
  // wäre er eine Zahl, die niemand benutzt — und das zweite Speichern liefe
  // gegen den veralteten Stand.
  it('schreibt beim zweiten Mal mit dem neuen Stand', async () => {
    const geschrieben = vi.fn<Schreibweg>(() =>
      Promise.resolve({ stand: 'neu', fehler: null }),
    )

    werkstatt(geschrieben)

    const feld = await bisZurDatei()

    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    await waitFor(() => {
      expect(geschrieben).toHaveBeenCalledTimes(1)
    })

    // **Auf die Wirkung gewartet, nicht auf den Ruf** (#1809): den neuen
    // Stand kennt die Seite erst mit der Antwort — und die sagt es.
    await screen.findByText(texte.gespeichert)

    await userEvent.type(feld, 'y')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    await waitFor(() => {
      expect(geschrieben).toHaveBeenCalledTimes(2)
    })

    // **Der vierte Wert des zweiten Aufrufs ist der Stand.** `waitFor` oben
    // hat schon zugesichert, dass es zwei gibt.
    expect(geschrieben.mock.calls[1]).toEqual([
      '/pakete/probe.mse-style',
      'style',
      'mse version: 2.0.0\nxy',
      'neu',
    ])
  })

  // **Ein Konflikt bekommt seinen eigenen Satz.** „Es ist ein unerwarteter
  // Fehler aufgetreten" sagte nicht, dass die Datei noch dasteht und was zu
  // tun ist.
  it('sagt bei einem Konflikt, was zu tun ist', async () => {
    werkstatt(() =>
      Promise.resolve({
        stand: '',
        fehler: {
          code: 'konflikt',
          schluessel: 'konflikt',
          meldung: 'geändert',
        },
      }),
    )

    const feld = await bisZurDatei()

    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    expect(await screen.findByRole('alert')).toHaveTextContent(texte.konflikt)

    // **Und der Text bleibt stehen** (§7, Regel 3): wer gerade zehn Minuten
    // getippt hat, verliert sie nicht an eine Meldung.
    expect(feld).toHaveValue('mse version: 2.0.0\nx')
  })

  // **Und ein Fehler, der kein Bindungsfehler ist, kommt auch an.**
  //
  // Die Gegenprobe zum Lauf darüber: er prüft, dass ein **Konflikt** seinen
  // eigenen Satz bekommt. Ohne diesen hier sagte er auch dann ja, wenn **jeder**
  // Fehler als Konflikt gälte — und dann stünde „zuerst neu laden" über einer
  // abgerissenen Verbindung.
  it('meldet auch einen Fehler, der keiner der Bindung ist', async () => {
    werkstatt(() => Promise.reject(new Error('die Leitung ist weg')))

    const feld = await bisZurDatei()

    await userEvent.type(feld, 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.speichern }))

    const meldung = await screen.findByRole('alert')

    expect(meldung).toHaveTextContent('die Leitung ist weg')
    expect(meldung).not.toHaveTextContent(texte.konflikt)
  })

  // **Ohne Änderung ist nichts zu speichern.**
  it('lässt den Knopf ruhen, solange nichts geändert ist', async () => {
    werkstatt(() => Promise.resolve({ stand: 'neu', fehler: null }))

    await bisZurDatei()

    expect(screen.getByRole('button', { name: texte.speichern })).toBeDisabled()
  })
})

// **Ohne gewähltes Paket sagt die Seite, was zu tun ist** — Werkstück C von
// Phase 4.10 (#1325).
//
// Der Abzug vom 2026-09-03 zeigte eine Seite, die zu 88 % leer war; das
// einzige, was darauf stand, nannte, was **nicht** ist. §6 verlangt das
// Gegenteil.
it('nennt ohne gewähltes Paket den nächsten Schritt', async () => {
  werkstatt(vi.fn())

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  // **Einmal, im Zentrum** — UX-15 (#1771). Hier stand „zweimal: im Zentrum
  // und im Inspector, denn beide warten auf die Datei"; in der Sichtprüfung
  // vom 2026-09-24 standen damit derselbe Satz nebeneinander. Der Inspector
  // sagt, was **er** zeigen wird.
  expect(await screen.findAllByText(texte.waehleImBaum)).toHaveLength(1)
  expect(
    screen.getByRole('complementary', { name: texte.inspectorBezeichnung }),
  ).toHaveTextContent(texte.inspectorLeer)

  // **Und er verschwindet, sobald eines dasteht.** Ein Hinweis, der neben der
  // Sache stehen bleibt, auf die er zeigt, ist keiner mehr.
  await userEvent.click(await screen.findByRole('button', { name: 'style' }))
  await screen.findByLabelText(texte.inhalt)

  expect(screen.queryByText(texte.waehleImBaum)).toBeNull()
  expect(screen.queryByText(texte.inspectorLeer)).toBeNull()
})
