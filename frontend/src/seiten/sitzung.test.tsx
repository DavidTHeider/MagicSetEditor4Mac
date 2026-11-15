// **Die lokale Sitzung am Fenster** — ID-02
// (#1701).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sitzungsbereich } from './Sitzungsbereich'
import { Sitzungstor } from './Sitzungstor'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { Sitzungsraum } from '../haken/useSitzungsschritt'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de.seiten.nutzer.sitzung
const ok = { fehler: null }

const stand = (
  zustand: 'angemeldet' | 'gesperrt' | 'abgemeldet',
  mehr: Partial<{ kennwort: boolean; sensor: boolean }> = {},
) => ({
  zustand,
  nutzerID: '1',
  kennwort: true,
  sensor: false,
  fehler: null,
  ...mehr,
})

const fehlerAm = (feld: string, meldung: string, schluessel = '') => ({
  fehler: { code: 'validierung', meldung, schluessel, feld },
})

afterEach(() => {
  raeumeApi()
})

function tor(geoeffnet = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Sitzungstor geoeffnet={geoeffnet}>
        <p>die Anwendung</p>
      </Sitzungstor>
    </Sprachraum>,
  )

  return geoeffnet
}

describe('Das Tor der Sitzung', () => {
  it('zeigt angemeldet die Anwendung', async () => {
    belegeApi({
      Sitzungsstand: vi.fn().mockResolvedValue(stand('angemeldet')),
    })
    tor()
    expect(await screen.findByText('die Anwendung')).toBeInTheDocument()
  })

  // **Ein Fehler beim Fragen lässt die Anwendung stehen** — die Sperre hält Go.
  it('zeigt die Anwendung auch, wenn der Stand nicht zu lesen ist', async () => {
    belegeApi({
      Sitzungsstand: vi.fn().mockResolvedValue({
        ...stand('angemeldet'),
        fehler: { code: 'intern', meldung: 'kaputt', schluessel: '' },
      }),
    })
    tor()
    expect(await screen.findByText('die Anwendung')).toBeInTheDocument()
  })

  it('entsperrt mit dem Kennwort und zeigt den Fehler am Feld', async () => {
    const sitzungsstand = vi
      .fn()
      .mockResolvedValueOnce(stand('gesperrt'))
      .mockResolvedValue(stand('angemeldet'))
    const entsperren = vi
      .fn()
      .mockResolvedValueOnce(fehlerAm('kennwort', 'falsch', 'kennwortFalsch'))
      .mockResolvedValue(ok)
    belegeApi({ Sitzungsstand: sitzungsstand, Entsperren: entsperren })
    const geoeffnet = tor()

    const feld = await screen.findByLabelText(texte.kennwort)
    expect(screen.queryByText('die Anwendung')).toBeNull()

    await userEvent.type(feld, 'falsch')
    await userEvent.click(
      screen.getByRole('button', { name: texte.entsperren }),
    )
    expect(
      await screen.findByText(kataloge.de.fehler.kennwortFalsch),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.entsperren }),
    )
    expect(await screen.findByText('die Anwendung')).toBeInTheDocument()
    expect(entsperren).toHaveBeenLastCalledWith('falsch')
    expect(geoeffnet).toHaveBeenCalledTimes(1)
  })

  it('entsperrt lokal, stellt über den Zettel wieder her und meldet sich ab', async () => {
    const wiederherstellen = vi
      .fn()
      .mockResolvedValueOnce(fehlerAm('faktor', 'Dieser Code stimmt nicht.'))
      .mockResolvedValueOnce(fehlerAm('neu', 'Zu kurz.'))
      .mockResolvedValue(ok)
    const sensor = vi.fn().mockResolvedValue({
      fehler: { code: 'intern', meldung: 'abgebrochen', schluessel: '' },
    })
    const abmelden = vi.fn().mockResolvedValue(ok)
    const sitzungsstand = vi
      .fn()
      .mockResolvedValue(stand('gesperrt', { kennwort: false, sensor: true }))
    belegeApi({
      Sitzungsstand: sitzungsstand,
      EntsperrenMitSensor: sensor,
      KennwortWiederherstellen: wiederherstellen,
      Abmelden: abmelden,
    })
    tor()

    // **Ohne Kennwort kein Kennwortfeld**, und Touch ID heißt lokal.
    await userEvent.click(
      await screen.findByRole('button', { name: texte.lokalEntsperren }),
    )
    expect(screen.queryByLabelText(texte.kennwort)).toBeNull()
    expect(await screen.findByRole('alert')).toHaveTextContent('abgebrochen')
    // Der erste Satz steht da, der Rest hinter dem Fragezeichen (UX-15).
    expect(
      screen.getByText(
        texte.lokalHinweis.slice(0, texte.lokalHinweis.indexOf('.')),
        { exact: false },
      ),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.vergessen }))
    await userEvent.type(screen.getByLabelText(texte.code), 'zettel')
    await userEvent.type(screen.getByLabelText(texte.kennwortNeu), 'kurz')
    const setzen = screen.getByRole('button', { name: texte.neuSetzen })

    await userEvent.click(setzen)
    expect(
      await screen.findByText('Dieser Code stimmt nicht.'),
    ).toBeInTheDocument()
    await userEvent.click(setzen)
    expect(await screen.findByText('Zu kurz.')).toBeInTheDocument()
    await userEvent.click(setzen)
    await waitFor(() => {
      expect(wiederherstellen).toHaveBeenCalledTimes(3)
    })
    expect(wiederherstellen).toHaveBeenLastCalledWith('1', 'zettel', 'kurz')

    // **Nach der Wiederherstellung wird neu gelesen** — die Sperranzeige
    // steht frisch da, das Formular ist zu.
    await waitFor(() => {
      expect(sitzungsstand).toHaveBeenCalledTimes(2)
    })
    await screen.findByRole('button', { name: texte.vergessen })
    expect(screen.queryByLabelText(texte.code)).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: texte.abmelden }))
    await waitFor(() => {
      expect(abmelden).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
  })

  it('meldet abgemeldet an — mit dem zweiten Faktor, wenn er gefragt wird', async () => {
    const anmelden = vi
      .fn()
      .mockResolvedValueOnce({
        fehler: {
          code: 'validierung',
          meldung: 'Dieser Nutzer verlangt einen zweiten Faktor.',
          schluessel: 'zweiterFaktor',
          feld: 'faktor',
        },
      })
      .mockResolvedValue(ok)
    belegeApi({
      Sitzungsstand: vi
        .fn()
        .mockResolvedValueOnce(stand('abgemeldet'))
        .mockResolvedValue(stand('angemeldet')),
      Nutzer: vi.fn().mockResolvedValue({
        nutzer: [
          { id: '1', name: 'Ich' },
          { id: '2', name: 'Ada' },
        ],
        fehler: null,
      }),
      LokalAnmelden: anmelden,
    })
    tor()

    await screen.findByRole('option', { name: 'Ada' })
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.wer }),
      '2',
    )
    await userEvent.type(screen.getByLabelText(texte.kennwort), 'geheim')
    await userEvent.click(screen.getByRole('button', { name: texte.anmelden }))

    const nachweis = await screen.findByRole('textbox', {
      name: kataloge.de.bausteine.nachweis.beschriftung,
    })
    await userEvent.type(nachweis, '123456{Enter}')

    expect(await screen.findByText('die Anwendung')).toBeInTheDocument()
    expect(anmelden).toHaveBeenLastCalledWith('2', 'geheim', '123456')
  })

  it('sagt es, wenn die Nutzer sich nicht lesen lassen, und nennt einen fremden Fehler', async () => {
    belegeApi({
      Sitzungsstand: vi.fn().mockResolvedValue(stand('abgemeldet')),
      Nutzer: vi.fn().mockResolvedValue({
        nutzer: [],
        fehler: { code: 'intern', meldung: 'Bestand weg', schluessel: '' },
      }),
      LokalAnmelden: vi.fn().mockResolvedValue({
        fehler: { code: 'intern', meldung: 'klemmt', schluessel: '' },
      }),
    })
    tor()

    expect(await screen.findByRole('alert')).toHaveTextContent('Bestand weg')
    expect(screen.getByRole('button', { name: texte.anmelden })).toBeDisabled()
  })
})

describe('Die Anmeldung, im Einzelnen', () => {
  it('nennt einen Fehler ohne Feld über dem Formular', async () => {
    belegeApi({
      Sitzungsstand: vi.fn().mockResolvedValue(stand('abgemeldet')),
      Nutzer: vi.fn().mockResolvedValue({
        nutzer: [{ id: '1', name: 'Ich' }],
        fehler: null,
      }),
      LokalAnmelden: vi.fn().mockResolvedValue({
        fehler: { code: 'intern', meldung: 'klemmt', schluessel: '' },
      }),
    })
    tor()

    await screen.findByRole('option', { name: 'Ich' })
    await userEvent.click(screen.getByRole('button', { name: texte.anmelden }))
    expect(await screen.findByRole('alert')).toHaveTextContent('klemmt')
  })
})

describe('Die Sitzung auf der Nutzerseite', () => {
  it('sperrt, meldet ab und setzt ein Kennwort', async () => {
    const wechseln = vi
      .fn()
      .mockResolvedValueOnce(
        fehlerAm('kennwort', 'Dieses Kennwort stimmt nicht.'),
      )
      .mockResolvedValue(ok)
    const sperren = vi.fn().mockResolvedValue({
      fehler: {
        code: 'validierung',
        meldung:
          'Ohne Kennwort und ohne lokale Entsperrung ließe sich die Sperre nicht lösen.',
        schluessel: 'validierung',
        feld: 'kennwort',
      },
    })
    const abmelden = vi.fn().mockResolvedValue(ok)
    belegeApi({
      Sperren: sperren,
      Abmelden: abmelden,
      KennwortWechseln: wechseln,
    })
    const neuLesen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Sitzungsraum.Provider value={neuLesen}>
          <Sitzungsbereich />
        </Sitzungsraum.Provider>
      </Sprachraum>,
    )

    // **Ohne Kennwort sperrt nichts** — die Meldung steht da, und nichts
    // wird neu gelesen.
    await userEvent.click(screen.getByRole('button', { name: texte.sperren }))
    await waitFor(() => {
      expect(sperren).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
    expect(neuLesen).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: texte.abmelden }))
    await waitFor(() => {
      expect(neuLesen).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()

    // **Das Kennwort ändern steht eingeklappt** — UXR-20 (#1928): es ist
    // kein Alltag wie Sperren und Abmelden.
    await userEvent.click(
      screen.getByRole('button', { name: texte.kennwortAendern }),
    )
    await userEvent.type(screen.getByLabelText(texte.kennwortAlt), 'alt')
    await userEvent.type(
      screen.getByLabelText(texte.kennwortNeu),
      'neu-und-lang',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.kennwortSetzen }),
    )
    expect(
      await screen.findByText('Dieses Kennwort stimmt nicht.'),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.kennwortSetzen }),
    )
    expect(await screen.findByRole('status')).toHaveTextContent(
      texte.kennwortGeaendert,
    )
    expect(wechseln).toHaveBeenLastCalledWith('alt', 'neu-und-lang')
    expect(screen.getByLabelText(texte.kennwortNeu)).toHaveValue('')

    // **Wer wieder tippt, sieht die alte Erfolgsmeldung nicht mehr.**
    await userEvent.type(screen.getByLabelText(texte.kennwortNeu), 'x')
    expect(screen.queryByRole('status')).toBeNull()
  })

  it('liest ohne Tor nichts neu und zeigt einen Fehler ohne Feld oben', async () => {
    const abmelden = vi.fn().mockResolvedValue(ok)
    belegeApi({
      Abmelden: abmelden,
      Sperren: vi.fn().mockResolvedValue({
        fehler: { code: 'intern', meldung: 'klemmt', schluessel: '' },
      }),
      KennwortWechseln: vi.fn().mockResolvedValue({
        fehler: { code: 'intern', meldung: 'Platte voll', schluessel: '' },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Sitzungsbereich />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: texte.sperren }))
    expect(await screen.findByRole('alert')).toHaveTextContent('klemmt')

    // **Ohne Tor liest ein Abmelden nichts neu** — und scheitert nicht daran.
    await userEvent.click(screen.getByRole('button', { name: texte.abmelden }))
    await waitFor(() => {
      expect(abmelden).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()

    await userEvent.click(
      screen.getByRole('button', { name: texte.kennwortAendern }),
    )
    await userEvent.type(
      screen.getByLabelText(texte.kennwortNeu),
      'neu-und-lang',
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.kennwortSetzen }),
    )
    expect(await screen.findByText(/Platte voll/)).toBeInTheDocument()
  })
})
