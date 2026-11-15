// **Nutzer und Profile bearbeiten** — Werkstück 5 von Phase 4.12
// (#1351).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'

import type { NutzerDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
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
  profile: [
    probeprofil({ id: '10', nutzerID: '1', name: 'Mein Profil', fassung: '1' }),
  ],
}

const gezeigtesProfil = () =>
  Promise.resolve({ profil: einer.profile[0] ?? null })

function belege(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    ...belegeAnmeldung(),
    GezeigtesProfil: vi.fn(gezeigtesProfil),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// oeffne klappt die Bearbeitung einer Zeile auf.
// oeffne klappt eine Zeile auf — **und seit UXR-20 (#1928) zuerst den
// Kasten, wenn er zu ist**: das Bearbeiten steht eingeklappt unter den
// eigenen Angaben. Ein zweiter Druck schlösse ihn wieder.
async function oeffne(name: string) {
  const kasten = await screen.findByRole('button', {
    name: 'Nutzer und Profile bearbeiten',
  })

  if (kasten.getAttribute('aria-expanded') === 'false') {
    await userEvent.click(kasten)
  }

  await userEvent.click(
    await screen.findByRole('button', { name: `${name} bearbeiten` }),
  )
}

// frage ist die Rückfrage vor dem Löschen, an ihrem Satz erkannt.
function frage(satz: string): HTMLElement {
  return screen.getByRole('group', { name: satz })
}

describe('Nutzer und Profile bearbeiten', () => {
  // **Jede Zeile lässt sich aufklappen** — und zugeklappt steht nur der Name.
  it('klappt eine Zeile auf und wieder zu', async () => {
    belege()
    await oeffne('Ich')

    expect(await screen.findByLabelText('Name')).toHaveValue('Ich')

    await userEvent.click(screen.getByRole('button', { name: 'Abbrechen' }))

    expect(screen.queryByLabelText('Name')).not.toBeInTheDocument()
  })

  // **Ein neuer Name geht an die Anwendung — und die Seite liest neu.**
  it('benennt einen Nutzer um', async () => {
    const aendern = vi.fn().mockResolvedValue({ nutzer: einer })

    belege({ NutzerAendern: aendern })
    await oeffne('Ich')

    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(screen.getByLabelText('Name'), 'Wer')
    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

    // **Der ganze Verbund und nicht nur der Name** — Werkstück 8 (#1351):
    // `NutzerAendern` **setzt**, es verschmilzt nicht. Ein Aufruf mit leeren
    // Feldern löschte Vorname, Slogan und Biografie, und niemand hätte das
    // verlangt.
    await vi.waitFor(() => {
      expect(aendern).toHaveBeenCalledWith({
        id: '1',
        name: 'Wer',
        bild: '',
        anrede: '',
        namenszusatz: '',
        vorname: '',
        weiterenamen: '',
        nachname: '',
        slogan: '',
        kurzebio: '',
        langebio: '',
        fassung: '1',
      })
    })
  })

  // **Und ein Profil ebenso** — über seine eigene Bindung.
  it('benennt ein Profil um', async () => {
    const aendern = vi.fn().mockResolvedValue({ profil: einer.profile[0] })

    belege({ ProfilAendern: aendern })
    await oeffne('Mein Profil')

    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(screen.getByLabelText('Name'), 'Turnier')
    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

    await vi.waitFor(() => {
      // **Mit dem ganzen Verbund seit Werkstück 9** (#1351): `ProfilAendern`
      // setzt und verschmilzt nicht, und die übrigen Angaben kommen aus
      // derselben Zeile, die dasteht.
      expect(aendern).toHaveBeenCalledWith(
        expect.objectContaining({ id: '10', name: 'Turnier', fassung: '1' }),
      )
    })
  })

  // **Löschen geht nur, solange der Name steht, wie er stand.**
  //
  // # Das ist die Bestätigung, und sie ist sichtbar
  //
  // Wer den Namen ändert, will umbenennen; wer ihn stehen lässt und auf
  // „Löschen" drückt, hat gesehen, wen er meint. Ein Dialog über einem Dialog
  // wäre die Alternative — und §6 verbietet sie.
  it('löscht nur bei stehendem Namen', async () => {
    const loeschen = vi.fn().mockResolvedValue({ fehler: null })

    belege({ NutzerLoeschen: loeschen })
    await oeffne('Ich')

    const knopf = await screen.findByRole('button', {
      name: 'Ich endgültig löschen',
    })

    // **Ein geänderter Name hält den Knopf an.**
    await userEvent.type(screen.getByLabelText('Name'), 'x')
    expect(knopf).toBeDisabled()

    await userEvent.clear(screen.getByLabelText('Name'))
    await userEvent.type(screen.getByLabelText('Name'), 'Ich')

    // **Und gelöscht wird erst nach der Frage** (UX-10.AK-04, #1682) —
    // Abbrechen löscht nichts.
    await userEvent.click(knopf)
    await userEvent.click(
      within(frage('Ich endgültig löschen?')).getByRole('button', {
        name: 'Abbrechen',
      }),
    )
    expect(loeschen).not.toHaveBeenCalled()

    await userEvent.click(knopf)
    await userEvent.click(
      within(frage('Ich endgültig löschen?')).getByRole('button', {
        name: 'Endgültig löschen',
      }),
    )

    await vi.waitFor(() => {
      expect(loeschen).toHaveBeenCalledWith('1', 'Ich')
    })
  })

  // **Und ein Profil ebenso.**
  it('löscht ein Profil', async () => {
    const loeschen = vi.fn().mockResolvedValue({ fehler: null })

    belege({ ProfilLoeschen: loeschen })
    await oeffne('Mein Profil')

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Mein Profil endgültig löschen',
      }),
    )
    await userEvent.click(
      within(frage('Mein Profil endgültig löschen?')).getByRole('button', {
        name: 'Endgültig löschen',
      }),
    )

    await vi.waitFor(() => {
      expect(loeschen).toHaveBeenCalledWith('10', 'Mein Profil')
    })
  })

  // **Ein gescheiterter Ruf meldet sich und lässt die Zeile offen.**
  //
  // §7, Regel 3 von der anderen Seite: nach einem gescheiterten Schreiben wird
  // nicht nachgeladen, und was eingetippt war, bleibt stehen.
  it('meldet einen gescheiterten Ruf und lässt die Zeile offen', async () => {
    belege({
      NutzerAendern: vi.fn(() =>
        Promise.resolve({
          nutzer: null,
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung: 'den Namen gibt es schon',
          },
        }),
      ),
    })
    await oeffne('Ich')

    await userEvent.click(
      await screen.findByRole('button', { name: 'Umbenennen' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )

    expect(screen.getByLabelText('Name')).toBeInTheDocument()
  })

  // **Und der Hinweis sagt, was ein Löschen bedeutet** (§6).
  it('sagt, was ein Löschen mitnimmt', async () => {
    belege()
    await oeffne('Ich')

    expect(
      await screen.findByText(/nimmt seine Profile mit/),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await oeffne('Mein Profil')

    expect(
      await screen.findByText(/letzte Profil eines Nutzers/),
    ).toBeInTheDocument()
  })

  // **Ein gescheitertes Löschen meldet sich ebenso.**
  it('meldet ein gescheitertes Löschen', async () => {
    belege({
      ProfilLoeschen: vi.fn(() =>
        Promise.resolve({
          fehler: {
            code: 'konflikt' as const,
            schluessel: 'konflikt',
            meldung: 'das letzte Profil bleibt',
          },
        }),
      ),
    })
    await oeffne('Mein Profil')

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Mein Profil endgültig löschen',
      }),
    )
    await userEvent.click(
      within(frage('Mein Profil endgültig löschen?')).getByRole('button', {
        name: 'Endgültig löschen',
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
  })

  // **Eine geschlossene Zeile räumt ihre Frage mit** (§7, Regel 1).
  it('stellt die Frage nach dem Schließen nicht wieder', async () => {
    belege({ ProfilLoeschen: vi.fn() })
    await oeffne('Mein Profil')

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Mein Profil endgültig löschen',
      }),
    )
    expect(frage('Mein Profil endgültig löschen?')).toBeInTheDocument()

    // **Das Abbrechen der Zeile**, nicht das der Frage — das eine, das in
    // keiner Gruppe steht.
    const [zeilenabbruch, ...weitere] = screen
      .getAllByRole('button', { name: 'Abbrechen' })
      .filter((knopf) => knopf.closest('[role="group"]') === null)
    expect(weitere).toEqual([])
    if (zeilenabbruch === undefined) {
      throw new Error('Die Zeile hat kein Abbrechen.')
    }
    await userEvent.click(zeilenabbruch)
    await oeffne('Mein Profil')

    expect(
      screen.queryByRole('group', { name: 'Mein Profil endgültig löschen?' }),
    ).toBeNull()
  })
})
