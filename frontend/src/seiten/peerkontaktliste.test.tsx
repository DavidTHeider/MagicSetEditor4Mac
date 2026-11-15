// **Wen dieses Profil kennt, als Liste und nicht als Formularwand** — UXR-16
// (#1924).
//
// # Der Befund
//
// **Jede Zeile war ein offenes Formular**: Stand, Fingerabdruck zum
// Aufklappen, Vergleich, Adresse mit „Merken", Notiz mit „Merken", je Zweck
// ein Knopf, die Sichten mit zwei Auswahlen und vier Knöpfen, Sperren. Mit
// acht Gegenübern (`test/gegenueberbestand.ts`) gezählt am 2026-10-02:
// **85 Bedienelemente und 23 Felder** im Kasten, und **vier Hauptknöpfe in
// einer Region** — jedes wartende Gegenüber trug sein „Stimmt überein"
// (R1, R2, R3 in `docs/UIREVIEW.md`). Und die Zeile nannte nicht, **wen** sie
// beschreibt: vorn stand „Unbestätigt", der Name nur als „Nennt sich: …".
//
// # Was jetzt gilt
//
// **Eine Zeile ist ein Name mit seinem Stand**, und **eine** ist offen: zu
// Anfang die erste, die auf einen Vergleich wartet — wie das Artboard
// *Gegenüber*, das neben der Liste genau einen bestätigen lässt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Peerkontakte } from './Peerkontakte'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { zaehle } from '../test/bedienzaehlung'
import { alsZeitpunkt } from '../zeitpunkt'
import {
  achtGegenueber,
  gegenueber,
  umfaenge,
  zweiSichten,
  zwecke,
} from '../test/gegenueberbestand'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik
const worte = eigene.gegenueber

afterEach(() => {
  raeumeApi()
})

function liste(kontakte: Peerkontakt[]) {
  return Promise.resolve({ kontakte, zwecke, fehler: null })
}

async function zeige(
  kontakte: Peerkontakt[] = achtGegenueber(),
  mehr: Parameters<typeof belegeApi>[0] = {},
) {
  belegeApi({
    Peerkontakte: () => liste(kontakte),
    Sammlungsfreigaben: (kennung: string) =>
      Promise.resolve({
        freigaben: zweiSichten(kennung),
        umfaenge,
        fehler: null,
      }),
    ...mehr,
  })

  render(
    <Sprachraum sprache="de">
      <Peerkontakte eingebettet />
    </Sprachraum>,
  )

  const hand = userEvent.setup()

  // **Die Liste faltet seit #2035 nach fünf** — diese Läufe prüfen jede
  // Zeile, also klappen sie auf, wo eine Falte steht
  // (`gegenueberfalte.test.tsx` hält die Falte selbst).
  await screen.findAllByRole('listitem')
  const weitere = screen.queryByRole('button', {
    name: (name) => /^Weitere \d+ zeigen$/.test(name),
  })
  if (weitere !== null) {
    await hand.click(weitere)
  }

  return hand
}

// zeile ist das Listenelement, das einen Namen trägt.
function zeile(name: string): HTMLElement {
  return screen
    .getByText(name, { selector: 'li span' })
    .closest('li') as HTMLElement
}

describe('Wen dieses Profil kennt', () => {
  it('nennt jedes Gegenüber beim Namen und mit seinem Stand', async () => {
    await zeige()

    for (const name of ['Anna', 'Bert', 'Cleo', 'Dora', 'Emil', 'Hugo']) {
      expect(zeile(name)).toBeInTheDocument()
    }

    expect(
      within(zeile('Emil')).getByText(eigene.kontakteStandBestaetigt),
    ).toBeInTheDocument()
    expect(
      within(zeile('Hugo')).getByText(eigene.kontakteStandGesperrt),
    ).toBeInTheDocument()
    expect(
      within(zeile('Anna')).getByText(eigene.kontakteStandUnbestaetigt),
    ).toBeInTheDocument()
  })

  it('öffnet zu Anfang nur das erste wartende Gegenüber', async () => {
    await zeige()

    expect(
      screen.getAllByRole('button', { name: eigene.einladung.zeileStimmt }),
    ).toHaveLength(1)
    expect(
      within(zeile('Anna')).getByRole('button', {
        name: eigene.einladung.zeileStimmt,
      }),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: `${worte.vergleichen}: Anna` }),
    ).toHaveAttribute('aria-expanded', 'true')
    expect(
      screen.getByRole('button', { name: `${worte.vergleichen}: Bert` }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.getByRole('button', { name: `${worte.verwalten}: Emil` }),
    ).toHaveAttribute('aria-expanded', 'false')
  })

  it('hält R1 und R3 mit acht Gegenübern', async () => {
    await zeige()

    const kasten = screen.getByRole('region', { name: eigene.kontakteTitel })
    const zahl = zaehle(kasten)

    expect(zahl.hauptJeRegion).toBeLessThanOrEqual(1)
    expect(zahl.felder).toBeLessThanOrEqual(12)
  })

  it('öffnet immer nur eine Zeile', async () => {
    const hand = await zeige()

    await hand.click(
      screen.getByRole('button', { name: `${worte.verwalten}: Emil` }),
    )

    expect(
      screen.getByRole('button', { name: `${worte.vergleichen}: Anna` }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: eigene.einladung.zeileStimmt }),
    ).toBeNull()
    expect(
      within(zeile('Emil')).getByRole('button', {
        name: eigene.kontakteSperren,
      }),
    ).toBeInTheDocument()

    // **Und zu heißt zu** — ein zweiter Klick schließt sie, und keine
    // andere springt auf.
    await hand.click(
      screen.getByRole('button', { name: `${worte.verwalten}: Emil` }),
    )

    expect(
      screen.queryByRole('button', { name: eigene.kontakteSperren }),
    ).toBeNull()
  })

  it('lässt die Zeile offen, an der eben bestätigt wurde', async () => {
    let stand = achtGegenueber()
    const bestaetigen = vi.fn((kennung: string) => {
      stand = stand.map((kontakt) =>
        kontakt.kennung === kennung
          ? { ...kontakt, stand: 'bestaetigt', lage: 'bestaetigt' }
          : kontakt,
      )

      return liste(stand)
    })

    const hand = await zeige(stand, {
      Peerkontakte: () => liste(stand),
      PeerkontaktBestaetigen: bestaetigen,
    })

    await hand.click(
      within(zeile('Anna')).getByRole('button', {
        name: eigene.einladung.zeileStimmt,
      }),
    )

    await waitFor(() => {
      expect(
        within(zeile('Anna')).getByText(eigene.kontakteStandBestaetigt),
      ).toBeInTheDocument()
    })

    // **Anna bleibt offen**, mit dem, was ein Bestätigter darf — und Bert,
    // jetzt der erste Wartende, springt nicht an ihre Stelle.
    expect(
      screen.getByRole('button', { name: `${worte.verwalten}: Anna` }),
    ).toHaveAttribute('aria-expanded', 'true')
    expect(
      screen.getByRole('button', { name: `${worte.vergleichen}: Bert` }),
    ).toHaveAttribute('aria-expanded', 'false')
  })

  it('nennt ein Gegenüber ohne eigenen Namen, wie es sich nennt — und sonst beim Abdruck', async () => {
    await zeige([
      gegenueber('Ida', { eigenerName: '' }),
      gegenueber('Jan', {
        eigenerName: '',
        selbstgewaehlterName: '',
        fingerabdruck: 'abcd-ef01',
      }),
      gegenueber('Hugo', { stand: 'gesperrt' }),
    ])

    expect(zeile('Idas Mac')).toBeInTheDocument()
    expect(zeile('abcd-ef01')).toBeInTheDocument()

    // **Der Vorbehalt steht dabei**, und der Name nicht ein zweites Mal.
    expect(screen.getAllByText(/Idas Mac/)).toHaveLength(1)
    expect(
      within(zeile('Idas Mac')).getByText(eigene.kontakteNenntHinweis),
    ).toBeInTheDocument()
  })

  it('schreibt den Zeitpunkt der letzten Begegnung lesbar', async () => {
    await zeige()

    expect(screen.queryByText(/2026-09-30T18:00:00Z/)).toBeNull()
    expect(
      within(zeile('Emil')).getByText(
        new RegExp(alsZeitpunkt('2026-09-30T18:00:00Z', 'de')),
      ),
    ).toBeInTheDocument()
  })
})

// **Was man selten setzt, steht hinter einem Aufklapper** — Adresse und
// Notiz kommen mit der Einladung oder dem Fund im Netz, und erteilte Sichten
// sind selten. Offen trug jede Zeile zwei Felder, zwei „Merken" und zwei
// Hilfen dafür, und jede bestätigte fragte beim Zeichnen den Bestand.
describe('eine offene Zeile', () => {
  it('legt Adresse und Notiz hinter einen Aufklapper', async () => {
    const hand = await zeige()

    await hand.click(
      screen.getByRole('button', { name: `${worte.verwalten}: Emil` }),
    )

    expect(screen.queryByLabelText(eigene.kontakteAdresse)).toBeNull()
    expect(screen.queryByLabelText(eigene.kontakteNotiz)).toBeNull()

    await hand.click(
      within(zeile('Emil')).getByRole('button', {
        name: worte.adresseUndNotiz,
      }),
    )

    expect(screen.getByLabelText(eigene.kontakteAdresse)).toHaveValue(
      'emils-mac.local:41234',
    )
    expect(screen.getByLabelText(eigene.kontakteNotiz)).toHaveValue(
      'tauscht fair',
    )
  })

  it('fragt die Sichten erst, wenn jemand sie sehen will', async () => {
    const gefragt = vi.fn((kennung: string) =>
      Promise.resolve({
        freigaben: zweiSichten(kennung),
        umfaenge,
        fehler: null,
      }),
    )
    const hand = await zeige(achtGegenueber(), {
      Sammlungsfreigaben: gefragt,
    })

    await hand.click(
      screen.getByRole('button', { name: `${worte.verwalten}: Emil` }),
    )

    expect(gefragt).not.toHaveBeenCalled()

    await hand.click(
      within(zeile('Emil')).getByRole('button', {
        name: eigene.freigabenTitel,
      }),
    )

    expect(
      await within(zeile('Emil')).findAllByRole('button', {
        name: eigene.freigabenWiderrufen,
      }),
    ).toHaveLength(2)
    expect(gefragt).toHaveBeenCalledWith('bemil')
  })
})
