// **Die Angaben eines Profils im Fenster** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 10 bis 13.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'
import { probeprofil } from '../test/profil'

import type { NutzerDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const einBild =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='

const eigene = kataloge.de.seiten.nutzer

// profil ist das gezeigte — mit Subslogan und Abzeichen.
const profil = probeprofil({
  id: '10',
  nutzerID: '1',
  name: 'Turnier',
  fassung: '2',
})

const einer: NutzerDTO = {
  id: '1',
  name: 'Ich',
  bild: '',
  email: '',
  telefon: '',
  anonym: false,
  anrede: '',
  namenszusatz: '',
  vorname: '',
  weiterenamen: '',
  nachname: '',
  slogan: '',
  kurzebio: '',
  langebio: '',
  fassung: '1',
  profile: [{ ...profil, subslogan: 'Nur alte Formate' }],
}

function zeige(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    ...belegeAnmeldung(),
    GezeigtesProfil: vi.fn(() => Promise.resolve({ profil })),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),

    // **Die Zugriffstoken stehen unter den Angaben** — Werkstück 10 (#1351).
    // Sie werden hier belegt, weil die Seite sie lädt; **geprüft** werden sie
    // in `zugriffstoken.test.tsx`.
    Token: vi.fn(() => Promise.resolve({ token: [] })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// imFormular grenzt die Suche auf den Abschnitt ein — auf der Seite stehen
// mehrere Formulare mit ähnlich beschrifteten Feldern.
async function imFormular() {
  return within(
    await screen.findByRole('region', { name: eigene.profilangaben }),
  )
}

// zumProfil klappt die Lasche *Profile* auf.
async function zumProfil() {
  await userEvent.click(
    await screen.findByRole('tab', { name: eigene.laschen.profil }),
  )
}

describe('Die Angaben eines Profils', () => {
  it('zeigt jedes Feld mit seinem Wert', async () => {
    zeige()
    await zumProfil()

    const formular = await imFormular()

    expect(formular.getByLabelText(eigene.profilfelder.subslogan)).toHaveValue(
      'Nur alte Formate',
    )

    for (const beschriftung of Object.values(eigene.profilfelder)) {
      expect(formular.getByLabelText(beschriftung)).toBeInTheDocument()
    }
  })

  // **Alle Felder gehen zusammen hinaus — samt der Fassung.**
  it('schickt jedes Feld und die Fassung mit', async () => {
    const aendern = vi.fn(() => Promise.resolve({ profil }))

    zeige({ ProfilAendern: aendern })
    await zumProfil()

    const formular = await imFormular()

    await userEvent.type(
      formular.getByLabelText(eigene.profilfelder.kurzebeschreibung),
      'Nur Legacy',
    )

    await userEvent.click(
      formular.getByRole('button', { name: eigene.profilangabenSichern }),
    )

    await screen.findByDisplayValue('Nur Legacy')

    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '10',
        name: 'Turnier',
        subslogan: 'Nur alte Formate',
        kurzebeschreibung: 'Nur Legacy',
        fassung: '2',
      }),
    )
  })

  // **Ein Profil lässt sich zum technischen machen** — Werkstück 10 (#1351),
  // Punkt 12.
  //
  // **Beides in einem Schreiben und nicht in zweien**: sie stehen im selben
  // Formular, weil sie dieselbe Fassung teilen (§7, Regel 4).
  it('macht ein Profil technisch und gibt ihm einen Ablauf', async () => {
    const aendern = vi.fn(() => Promise.resolve({ profil }))

    zeige({ ProfilAendern: aendern })
    await zumProfil()

    const formular = await imFormular()

    await userEvent.click(formular.getByLabelText(eigene.technisch))
    await userEvent.type(
      formular.getByLabelText(eigene.ablauf),
      '2027-01-01T00:00:00Z',
    )

    await userEvent.click(
      formular.getByRole('button', { name: eigene.profilangabenSichern }),
    )

    expect(aendern).toHaveBeenCalledWith(
      expect.objectContaining({
        id: '10',
        technisch: true,
        ablauf: '2027-01-01T00:00:00Z',
        fassung: '2',
      }),
    )
  })

  // **Das Abzeichen liegt über dem Gesicht** — Punkt 13.
  it('legt das Abzeichen über das Nutzerfoto', async () => {
    zeige({
      Nutzerbilder: vi.fn(() =>
        Promise.resolve({ bilder: {}, badges: { '10': einBild } }),
      ),
    })

    const abzeichen = await screen.findAllByRole('img', {
      name: 'Abzeichen des Profils von Ich',
    })

    expect(abzeichen[0]).toHaveAttribute('src', einBild)
  })

  it('lässt das Abzeichen wählen und wieder wegnehmen', async () => {
    const waehlen = vi.fn(() => Promise.resolve({ profil }))
    const entfernen = vi.fn(() => Promise.resolve({ profil }))

    zeige({
      ProfilbadgeWaehlen: waehlen,
      ProfilbadgeEntfernen: entfernen,
      Nutzer: vi.fn(() =>
        Promise.resolve({
          nutzer: [
            { ...einer, profile: [{ ...profil, badge: 'profil-10.png' }] },
          ],
        }),
      ),
    })

    await zumProfil()

    const formular = await imFormular()

    await userEvent.click(
      formular.getByRole('button', { name: eigene.badgeWaehlen }),
    )

    await userEvent.click(
      formular.getByRole('button', { name: eigene.badgeEntfernen }),
    )

    expect(waehlen).toHaveBeenCalledWith('10', '2')
    expect(entfernen).toHaveBeenCalledWith('10', '2')
  })

  // **Ohne Profilnamen ruht der Knopf** — und das Feld sagt, warum.
  it('hält den Knopf an, solange der Profilname fehlt', async () => {
    zeige()
    await zumProfil()

    const formular = await imFormular()

    await userEvent.clear(formular.getByLabelText(eigene.profilfelder.name))

    expect(
      formular.getByRole('button', { name: eigene.profilangabenSichern }),
    ).toBeDisabled()
  })

  // **Ohne Abzeichen gibt es nichts wegzunehmen** — und keinen Knopf dafür.
  //
  // **Der Beleg, dass jemand zuhört, steht daneben**: der Knopf zum Wählen
  // ist da.
  it('zeigt keinen Knopf zum Wegnehmen, wenn keines da ist', async () => {
    zeige()
    await zumProfil()

    const formular = await imFormular()

    expect(
      formular.getByRole('button', { name: eigene.badgeWaehlen }),
    ).toBeInTheDocument()

    expect(
      formular.queryByRole('button', { name: eigene.badgeEntfernen }),
    ).toBeNull()
  })

  it('meldet, was beim Sichern schiefgeht', async () => {
    zeige({
      ProfilAendern: vi.fn(() =>
        Promise.resolve({
          profil: null,
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung: 'Inzwischen geändert.',
          },
        }),
      ),
    })

    await zumProfil()

    const formular = await imFormular()

    await userEvent.click(
      formular.getByRole('button', { name: eigene.profilangabenSichern }),
    )

    expect(await formular.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })
})
