// **Ein Exemplar über seinen Namen aufnehmen** — AP-10
// (#2367),
// Entwurf `Wege09`: der Name führt, die Druckkennung steht unter „Weitere
// Angaben“ — und der gewählte Druck trägt Name, Set und Nummer mit.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { effekteNachholen } from '../test/effekte'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlungsformular } from './Sammlungsformular'

const texte = kataloge.de.seiten.sammlung

afterEach(raeumeApi)

const boltM11 = {
  printing: 'a',
  name: 'Lightning Bolt',
  setcode: 'm11',
  setname: 'Magic 2011',
  nummer: '149',
  seltenheit: 'common',
  bild: '',
}

function deutung(treffer: (typeof boltM11)[]) {
  return {
    deutung: {
      zeilen: [
        {
          roh: '1 Lightning Bolt',
          zeilennummer: 1,
          menge: 1,
          name: 'Lightning Bolt',
          setcode: '',
          nummer: '',
          finish: 'nonfoil',
          treffer,
          drucke: treffer.length,
          eindeutig: treffer.length === 1,
          befund: '',
          weitere: false,
          fehler: null,
        },
      ],
      befunde: [],
      alterSekunden: 0,
    },
    fehler: null,
  }
}

describe('das Aufnehmen über den Namen', () => {
  it('nimmt den gewählten Druck mit Name, Set und Nummer auf', async () => {
    belegeApi({
      ListeDeuten: () =>
        Promise.resolve({
          deutung: {
            zeilen: [
              {
                roh: '1 Lightning Bolt',
                zeilennummer: 1,
                menge: 1,
                name: 'Lightning Bolt',
                setcode: '',
                nummer: '',
                finish: 'nonfoil',
                treffer: [
                  {
                    printing: 'a',
                    name: 'Lightning Bolt',
                    setcode: 'm11',
                    setname: 'Magic 2011',
                    nummer: '149',
                    seltenheit: 'common',
                    bild: '',
                  },
                  {
                    printing: 'b',
                    name: 'Lightning Bolt',
                    setcode: 'm10',
                    setname: 'Magic 2010',
                    nummer: '146',
                    seltenheit: 'common',
                    bild: '',
                  },
                ],
                drucke: 2,
                eindeutig: false,
                befund: '',
                weitere: false,
                fehler: null,
              },
            ],
            befunde: [],
            alterSekunden: 0,
          },
          fehler: null,
        }),
    })
    const aufgenommen = vi.fn(() => Promise.resolve(true))
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
      </Sprachraum>,
    )

    // **Die Druckkennung steht nicht mehr im Hauptweg.**
    expect(screen.queryByLabelText(texte.druck)).toBeNull()

    await userEvent.type(
      screen.getByLabelText(texte.drucksuche.karte),
      'Lightning Bolt{Enter}',
    )
    await userEvent.click(
      await screen.findByRole('radio', {
        name: 'Lightning Bolt · Magic 2010 (M10) · Nr. 146',
      }),
    )
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(aufgenommen).toHaveBeenCalledWith(
      expect.objectContaining({
        printing: 'b',
        name: 'Lightning Bolt',
        setcode: 'm10',
        nummer: '146',
      }),
    )
  })

  // **Was die Suche nicht mehr zeigt, nimmt das Formular nicht auf**
  // (Zweitblick 6b, Befund 1): A gewählt, dann B ohne Treffer gesucht — die
  // Kennung von A stand noch hinter „Weitere Angaben“, und „Aufnehmen“
  // schrieb sie.
  it('räumt den übernommenen Druck, wenn eine neue Suche nichts findet', async () => {
    belegeApi({
      ListeDeuten: vi
        .fn()
        .mockResolvedValueOnce(deutung([boltM11]))
        .mockResolvedValueOnce(deutung([])),
    })
    const aufgenommen = vi.fn(() => Promise.resolve(true))
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
      </Sprachraum>,
    )
    const suche = screen.getByLabelText(texte.drucksuche.karte)
    const aufnehmen = screen.getByRole('button', { name: texte.aufnehmen })

    await userEvent.type(suche, 'Lightning Bolt{Enter}')
    await screen.findByRole('radio', { checked: true })
    expect(aufnehmen).toBeEnabled()

    await userEvent.clear(suche)
    await userEvent.type(suche, 'Gibtesnicht{Enter}')
    expect(
      await screen.findByText(texte.drucksuche.keinTreffer),
    ).toBeInTheDocument()

    expect(aufnehmen).toBeDisabled()
    await userEvent.click(screen.getByText(texte.weitereAngaben))
    expect(screen.getByLabelText(texte.druck)).toHaveValue('')
    expect(screen.getByLabelText(texte.name)).toHaveValue('')
    expect(screen.getByLabelText(texte.setcode)).toHaveValue('')
    expect(screen.getByLabelText(texte.nummer)).toHaveValue('')
  })

  // **Von Hand Eingetragenes bleibt** — geräumt wird, was die Suche
  // übernommen hat, und nicht, was jemand selbst geschrieben hat.
  it('lässt eine von Hand geänderte Angabe stehen', async () => {
    belegeApi({
      ListeDeuten: vi
        .fn()
        .mockResolvedValueOnce(deutung([boltM11]))
        .mockResolvedValueOnce(deutung([])),
    })
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular
          aufAufnehmen={() => Promise.resolve(true)}
          laeuft={false}
        />
      </Sprachraum>,
    )
    const suche = screen.getByLabelText(texte.drucksuche.karte)

    await userEvent.type(suche, 'Lightning Bolt{Enter}')
    await screen.findByRole('radio', { checked: true })
    await userEvent.click(screen.getByText(texte.weitereAngaben))
    await userEvent.clear(screen.getByLabelText(texte.nummer))
    await userEvent.type(screen.getByLabelText(texte.nummer), '150')

    await userEvent.clear(suche)
    await userEvent.type(suche, 'Gibtesnicht{Enter}')
    await screen.findByText(texte.drucksuche.keinTreffer)

    expect(screen.getByLabelText(texte.druck)).toHaveValue('')
    expect(screen.getByLabelText(texte.nummer)).toHaveValue('150')
  })

  // **Nach dem Aufnehmen beginnt die Suche neu** — sonst stünde ihr Druck
  // gewählt da, während das Formular leer ist, und ein zweiter Klick auf das
  // schon gewählte Optionsfeld meldete nichts.
  it('räumt nach einem geglückten Aufnehmen auch die Suche', async () => {
    belegeApi({
      ListeDeuten: vi
        .fn()
        .mockResolvedValueOnce(deutung([boltM11]))
        .mockResolvedValueOnce(deutung([boltM11])),
    })
    const aufgenommen = vi.fn(() => Promise.resolve(true))
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
      </Sprachraum>,
    )
    const aufnehmen = screen.getByRole('button', { name: texte.aufnehmen })

    await userEvent.type(
      screen.getByLabelText(texte.drucksuche.karte),
      'Lightning Bolt{Enter}',
    )
    await screen.findByRole('radio', { checked: true })
    await userEvent.click(aufnehmen)

    await waitFor(() => {
      expect(screen.queryByRole('radio')).toBeNull()
    })
    expect(screen.getByLabelText(texte.drucksuche.karte)).toHaveValue('')
    expect(aufnehmen).toBeDisabled()

    // **Und dieselbe Karte ein zweites Mal** geht wieder über die Suche.
    await userEvent.type(
      screen.getByLabelText(texte.drucksuche.karte),
      'Lightning Bolt{Enter}',
    )
    await screen.findByRole('radio', { checked: true })
    await userEvent.click(aufnehmen)
    expect(aufgenommen).toHaveBeenCalledTimes(2)
  })

  // **Eine Antwort aus der vorigen Runde fällt nicht in die neue** — Nachlauf
  // zum Zweitblick über Phase 6, zweiter Teil, R1: eine Kennung von Hand,
  // eine Suche unterwegs, dann „Aufnehmen“. Die Suche entsteht neu, und ihre
  // späte Antwort setzte den Druck still in das leere Formular.
  it('lässt eine Suchantwort aus der vorigen Runde fallen', async () => {
    let spaet: (wert: ReturnType<typeof deutung>) => void = () => undefined
    belegeApi({
      ListeDeuten: vi.fn().mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            spaet = fertig
          }),
      ),
    })
    const aufgenommen = vi.fn(() => Promise.resolve(true))
    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={aufgenommen} laeuft={false} />
      </Sprachraum>,
    )
    await userEvent.click(screen.getByText(texte.weitereAngaben))
    await userEvent.type(screen.getByLabelText(texte.druck), 'von-hand')
    await userEvent.type(
      screen.getByLabelText(texte.drucksuche.karte),
      'Lightning Bolt{Enter}',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))
    await waitFor(() => {
      expect(screen.getByLabelText(texte.druck)).toHaveValue('')
    })

    spaet(deutung([boltM11]))
    await effekteNachholen()

    expect(screen.getByLabelText(texte.druck)).toHaveValue('')
    expect(screen.getByRole('button', { name: texte.aufnehmen })).toBeDisabled()
    expect(aufgenommen).toHaveBeenCalledWith(
      expect.objectContaining({ printing: 'von-hand' }),
    )
  })
})
