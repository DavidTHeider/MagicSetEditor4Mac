// **Die Wahl je Frage auf der Einstellungsseite** — Werkstück 8 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Einstellungen } from './Einstellungen'
import { Sprachraum } from '../sprache/kontext'
import { zuGebiet } from '../test/gebiete'
import { kataloge } from '../sprache/kataloge'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from '../test/bestandspflegeprobe'
import type { EinstellungDTO } from '../api'

const texte = kataloge.de.seiten.einstellungen

const suchfrage: EinstellungDTO = {
  schluessel: 'frage.suche',
  wert: 'scryfall',
  wirksam: 'scryfall',
  werte: ['scryfall', 'mtgapi', 'mtgjson', 'scrydex'],
  bereich: 'dienste',
}

const fragen: EinstellungDTO[] = [suchfrage]

// **Alle vier an** — der Stand, in dem jede Wahl offensteht.
const alleAn: EinstellungDTO[] = (
  ['scryfall', 'mtgapi', 'mtgjson', 'scrydex'] as const
).map((kennung) => ({
  schluessel: `dienst.${kennung}`,
  wert: 'an',
  wirksam: 'an',
  werte: ['aus', 'an'],
  bereich: 'dienste',
}))

async function hinstellen(
  setze = vi.fn(),
  eintraege: EinstellungDTO[] = fragen,
  dienste: EinstellungDTO[] = alleAn,
) {
  render(
    <Sprachraum sprache="de">
      <Einstellungen
        eintraege={[]}
        dienste={dienste}
        fragen={eintraege}
        ki={[]}
        vorgaben={[]}
        anmeldung={[]}
        ursache={undefined}
        meldefehler={undefined}
        setze={setze}
        laeuft={() => false}
        bestandspflege={stillePflege()}
        ruecksprung={stillerRuecksprung()}
        pflege={stilleDatenpflege()}
        setID={undefined}
      />
    </Sprachraum>,
  )

  // **Die Dienste liegen in ihrem Gebiet** — Werkstück I von Phase 4.10
  // (#1325).
  await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.dienste)
}

describe('die Wahl je Frage', () => {
  it('steht in einem eigenen Bereich', async () => {
    await hinstellen()

    const bereich = screen.getByRole('region', { name: texte.fragen })

    expect(
      within(bereich).getByLabelText(texte['frage.suche']),
    ).toBeInTheDocument()
  })

  // **Die Dienste stehen bei ihrem Eigennamen** (0063, Entscheidung 5) —
  // übersetzt wird die **Frage**, nicht der Anbieter.
  it('nennt die Dienste bei ihrem Eigennamen', async () => {
    await hinstellen()

    // **Vier Dienste sind seit UXR-F2 (#2008) vier Segmente.**
    const wahl = within(
      screen.getByRole('region', { name: texte.fragen }),
    ).getByLabelText(texte['frage.suche'])

    for (const name of [
      'Scryfall',
      'magicthegathering.io',
      'MTGJSON',
      'Scrydex',
    ]) {
      expect(within(wahl).getByRole('radio', { name })).toBeInTheDocument()
    }
  })

  it('reicht die Wahl durch', async () => {
    const setze = vi.fn()

    await hinstellen(setze)

    await userEvent.click(
      within(
        within(
          screen.getByRole('region', { name: texte.fragen }),
        ).getByLabelText(texte['frage.suche']),
      ).getByRole('radio', { name: 'magicthegathering.io' }),
    )

    expect(setze).toHaveBeenCalledWith('frage.suche', 'mtgapi')
  })

  // **Eine Frage ohne Namen fällt auf ihren Schlüssel zurück.**
  //
  // Sie kann nur entstehen, wenn Go eine zweite Frage kennt und das Fenster
  // noch nicht — dann steht dort `frage.preise` statt einer leeren Zeile.
  it('nennt eine unbekannte Frage bei ihrem Schlüssel', async () => {
    await hinstellen(vi.fn(), [
      {
        schluessel: 'frage.preise',
        wert: 'scryfall',
        wirksam: 'scryfall',
        werte: ['scryfall', 'scrydex'],
        bereich: 'dienste',
      },
    ])

    expect(
      within(screen.getByRole('region', { name: texte.fragen })).getByLabelText(
        'frage.preise',
      ),
    ).toBeInTheDocument()
  })

  it('sagt nichts, wo keine Frage eine Wahl hat', async () => {
    await hinstellen(vi.fn(), [])

    expect(screen.queryByRole('region', { name: texte.fragen })).toBeNull()
  })
  // **Befund 4 des Zweitblicks über Phase 4.8.**
  //
  // [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md) verlangt
  // eine Wahl „aus den Diensten, die eingeschaltet **und** für die Frage
  // geeignet sind". Angeboten wurden alle geeigneten — auch die
  // abgeschalteten, und zwar unmittelbar unter den Schaltern, die sie
  // ausschalten.
  it('bietet keinen abgeschalteten Dienst an', async () => {
    await hinstellen(
      vi.fn(),
      fragen,
      alleAn.map((schalter) =>
        schalter.schluessel === 'dienst.scrydex'
          ? { ...schalter, wert: 'aus' }
          : schalter,
      ),
    )

    const bereich = screen.getByRole('region', { name: texte.fragen })
    const liste = within(bereich).getByLabelText(texte['frage.suche'])

    expect(
      within(liste)
        .getAllByRole('radio')
        .map((eines) => eines.getAttribute('value')),
    ).toEqual(['scryfall', 'mtgapi', 'mtgjson'])

    await Promise.resolve()
  })

  // **Der eingestellte bleibt darin, auch wenn er aus ist.**
  //
  // Sonst zeigte die Auswahl einen anderen Dienst als den gespeicherten — sie
  // behauptete eine Wahl, die niemand getroffen hat. Dass er nicht antwortet,
  // sagt die Suche mit Namen und Weg.
  it('behält den eingestellten Dienst, auch wenn er abgeschaltet ist', async () => {
    await hinstellen(
      vi.fn(),
      [{ ...suchfrage, wert: 'scrydex' }],
      alleAn.map((schalter) =>
        schalter.schluessel === 'dienst.scrydex'
          ? { ...schalter, wert: 'aus' }
          : schalter,
      ),
    )

    const bereich = screen.getByRole('region', { name: texte.fragen })
    const liste = within(bereich).getByLabelText(texte['frage.suche'])

    expect(within(liste).getByRole('radio', { name: 'Scrydex' })).toBeChecked()
    expect(
      within(liste)
        .getAllByRole('radio')
        .map((eines) => eines.getAttribute('value')),
    ).toEqual(['scryfall', 'mtgapi', 'mtgjson', 'scrydex'])
  })
})
