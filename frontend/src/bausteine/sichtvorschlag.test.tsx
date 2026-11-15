// **Die Sicht des Gegenübers über die Leitung, und sein Vorschlag als
// Zeilen** — P11
// (#2379),
// Befund C-15: was zusammenpasst, sah nur, wer sich den Stand der Gegenseite
// als Datei oder Text besorgte, und übernehmen ließ es sich nicht.
import { cleanup, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Handelspositionen } from './Handelspositionen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { Handelsposition } from '../apihandelsleitung'
import type { Sichtkarte, Sichtvorschauantwort } from '../apisichtfreigabe'
import type {
  Tauschabgleichantwort,
  Tauschkandidat,
} from '../apitauschabgleich'

const texte = kataloge.de.seiten.handeln.sicht

afterEach(() => {
  raeumeApi()
})

const angebot: Sichtkarte = {
  referenz: 'r1',
  printing: 'p-sturm',
  menge: 2,
  zustand: 'excellent',
  sprache: 'en',
  finish: 'foil',
  bezeichnung: 'Sturmrufer',
}
const gesucht: Sichtkarte = {
  ...angebot,
  referenz: 'r2',
  printing: 'p-glut',
  bezeichnung: 'Glutläufer',
  wunsch: true,
}

const kandidat = (
  angebotDruck: string,
  name: string,
  variante: Pick<
    Tauschkandidat,
    'angebotZustand' | 'angebotSprache' | 'angebotFinish'
  > = {
    angebotZustand: 'nearMint',
    angebotSprache: 'de',
    angebotFinish: 'nonfoil',
  },
): Tauschkandidat => ({
  wunsch: name,
  angebot: name,
  angebotDruck,
  ...variante,
  menge: 1,
  fehlend: 0,
  rang: 'normal',
  art: 'genau',
})

const seite = (zeilen: Tauschkandidat[]) => ({
  zeilen,
  stueck: zeilen.length,
  gedeckt: zeilen.length,
  offen: 0,
  wert: 0,
  waehrung: '',
  bepreist: false,
})

function einAbgleich(): Tauschabgleichantwort {
  const gebe = [kandidat('p-glut', 'Glutläufer')]
  const bekomme = [
    kandidat('p-sturm', 'Sturmrufer', {
      angebotZustand: 'excellent',
      angebotSprache: 'en',
      angebotFinish: 'foil',
    }),
  ]

  return {
    kennung: 'du-1',
    gebe,
    bekomme,
    vorschlag: { gebe: seite(gebe), bekomme: seite(bekomme) },
    hatVorschlag: true,
    befunde: [],
    fehler: null,
  }
}

function Buehne({
  gesetzt,
  anfang = [],
}: {
  gesetzt: (positionen: Handelsposition[]) => void
  anfang?: Handelsposition[]
}) {
  const [positionen, setzePositionen] = useState<Handelsposition[]>(anfang)

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Handelspositionen
          ich="ich-1"
          gegenueber="du-1"
          positionen={positionen}
          bearbeitbar
          setzePositionen={(neu) => {
            gesetzt(neu)
            setzePositionen(neu)
          }}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

function zeige(
  sicht: () => Promise<Sichtvorschauantwort>,
  abgleichen: () => Tauschabgleichantwort = einAbgleich,
  anfang: Handelsposition[] = [],
) {
  const anfragen = vi.fn(sicht)
  const abgleich = vi.fn(() => Promise.resolve(abgleichen()))
  belegeApi({
    HandelSichtAnfragen: anfragen,
    Tauschabgleich: abgleich,
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  const gesetzt = vi.fn()
  render(<Buehne gesetzt={gesetzt} anfang={anfang} />)

  return { anfragen, abgleich, gesetzt }
}

const zweiZeilen = () =>
  Promise.resolve({ zeilen: [angebot, gesucht], fehler: null })

describe('die Sicht des Gegenübers', () => {
  it('fragt über die Leitung, gleicht ab und übernimmt den Vorschlag als Zeilen', async () => {
    const { anfragen, abgleich, gesetzt } = zeige(() =>
      Promise.resolve({ zeilen: [angebot, gesucht], fehler: null }),
    )
    const hand = userEvent.setup()

    await hand.click(screen.getByRole('button', { name: texte.anfragen }))

    // **Angebot und Wunsch getrennt** — eine Wunschzeile ist kein Angebot.
    expect(anfragen).toHaveBeenCalledWith('du-1')
    await vi.waitFor(() => {
      expect(abgleich).toHaveBeenCalledWith(
        'du-1',
        { angebot: [angebot], wuensche: [gesucht] },
        1,
        0,
      )
    })

    const gruppe = await screen.findByRole('group', { name: texte.titel })
    expect(
      within(gruppe).getByText(
        texte.zeilen.replace('{zahl}', '2').replace('{wuensche}', '1'),
      ),
    ).toBeVisible()

    const uebernehmen = within(gruppe).getByRole('button', {
      name: texte.uebernehmen,
    })
    await hand.click(uebernehmen)

    // **Jede Zeile kommt mit ihrer Variante** — auch die eigene (Befund 2
    // des Zweitblicks über Phase 8): ohne Sprache scheiterte drüben der
    // Zugang beim Verbuchen. Welche Besitzzeile weggeht, wählt weiter
    // „Zusagen" oder der Zug.
    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({
        von: 'ich-1',
        printing: 'p-glut',
        menge: 1,
        guete: 'nearMint',
        sprache: 'de',
        finish: 'nonfoil',
      }),
      expect.objectContaining({
        von: 'du-1',
        printing: 'p-sturm',
        menge: 1,
        guete: 'excellent',
        sprache: 'en',
        finish: 'foil',
      }),
    ])

    // **Kein zweites Mal**: übernommen ist übernommen.
    expect(uebernehmen).toBeDisabled()
    expect(gesetzt).toHaveBeenCalledTimes(1)
  })

  it('nennt eine Absage am Ort und legt nichts an', async () => {
    const { abgleich, gesetzt } = zeige(() =>
      Promise.resolve({
        zeilen: [],
        fehler: {
          code: 'erlaubnis',
          meldung: 'Das Gegenüber gibt dir zum Handeln nichts frei.',
          schluessel: '',
          feld: '',
        },
      }),
    )
    const hand = userEvent.setup()

    await hand.click(screen.getByRole('button', { name: texte.anfragen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das Gegenüber gibt dir zum Handeln nichts frei.',
    )
    expect(abgleich).not.toHaveBeenCalled()
    expect(gesetzt).not.toHaveBeenCalled()
  })

  it('legt eine Zeile, die schon dasteht, nicht doppelt an', async () => {
    // **Dieselbe Zeile heißt: derselbe Druck in derselben Variante.**
    const schonDa: Handelsposition = {
      von: 'du-1',
      printing: 'p-sturm',
      druckId: 0,
      menge: 1,
      guete: 'excellent',
      sprache: 'en',
      finish: 'foil',
      abbild: '',
      bewertung: 0,
    }
    const { gesetzt } = zeige(zweiZeilen, einAbgleich, [schonDa])
    const hand = userEvent.setup()

    await hand.click(screen.getByRole('button', { name: texte.anfragen }))
    await hand.click(
      await screen.findByRole('button', { name: texte.uebernehmen }),
    )

    expect(gesetzt).toHaveBeenLastCalledWith([
      schonDa,
      expect.objectContaining({ von: 'ich-1', printing: 'p-glut' }),
    ])
    expect(
      screen.getByText(texte.uebernommen.replace('{zahl}', '1')),
    ).toBeVisible()
  })

  it('nimmt ohne Gegenleistung, was du bekommen könntest — und sagt, wenn nichts passt', async () => {
    const ohneVorschlag = (): Tauschabgleichantwort => ({
      ...einAbgleich(),
      gebe: [],
      hatVorschlag: false,
    })
    const { gesetzt } = zeige(zweiZeilen, ohneVorschlag)
    const hand = userEvent.setup()

    await hand.click(screen.getByRole('button', { name: texte.anfragen }))
    await hand.click(
      await screen.findByRole('button', { name: texte.uebernehmen }),
    )

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ von: 'du-1', printing: 'p-sturm' }),
    ])
    expect(screen.queryByText(texte.gibst)).toBeNull()
    expect(screen.getByText(texte.bekommst)).toBeVisible()

    cleanup()
    zeige(zweiZeilen, () => ({ ...ohneVorschlag(), bekomme: [] }))
    await hand.click(screen.getByRole('button', { name: texte.anfragen }))

    expect(await screen.findByText(texte.keinVorschlag)).toBeVisible()
    expect(screen.queryByRole('button', { name: texte.uebernehmen })).toBeNull()
  })

  it('schreibt nach dem Verlassen nichts mehr hierher', async () => {
    let antworten: ((antwort: Sichtvorschauantwort) => void) | undefined
    const { abgleich } = zeige(
      () =>
        new Promise((aufloesen) => {
          antworten = aufloesen
        }),
    )

    await userEvent.click(screen.getByRole('button', { name: texte.anfragen }))
    expect(screen.getByText(texte.laeuft)).toBeVisible()

    cleanup()
    antworten?.({ zeilen: [angebot], fehler: null })
    await vi.waitFor(() => {
      expect(abgleich).toHaveBeenCalledTimes(1)
    })
    expect(screen.queryByRole('group')).toBeNull()
  })
})
