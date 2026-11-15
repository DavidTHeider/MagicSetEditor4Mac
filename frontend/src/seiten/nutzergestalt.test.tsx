// **Die Gestalt der Nutzerseite** — UXR-20
// (#1928),
// nach dem Artboard `Nutzer`: *+ Profil anlegen* führt im Kopf und öffnet
// eine Schublade.
//
// # Der Anlass
//
// **Gemessen am 2026-10-02**, am nativen Fenster mit einem Menschen und
// einem Profil: *Allgemeines* 13 Bedienelemente, 10 Felder, zwei
// Hauptknöpfe. Mit drei Menschen, sieben Profilen, zwei Freigaben, drei
// Token und fünf Vorgaben — gezählt wie das Drehbuch `uirevue` — zeichnete
// *Profile* **36 Felder** (R3 in `docs/UIREVIEW.md` lässt zwölf zu), und das
// Anlegen eines Profils stand ganz unten, unter Vorgaben, Freigaben und Token.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { probeprofil } from '../test/profil'
import {
  belegeSeite,
  belegeVielBestand,
  gezeichneteFelder,
  zeigeNutzerseite,
} from '../test/nutzerseite'

afterEach(() => {
  raeumeApi()
})

const t = kataloge.de.seiten.nutzer

async function oeffneAnlage(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', { name: t.profilAnlegenOeffnen }),
  )

  return screen.findByRole('dialog', { name: t.neuesProfil })
}

describe('Die Nutzerseite', () => {
  it('führt im Kopf mit „Profil anlegen …“ und legt in einer Schublade an', async () => {
    const anlegen = vi.fn().mockResolvedValue({
      profil: probeprofil({ id: '20', name: 'Turnier' }),
    })
    belegeSeite({ ProfilAnlegen: anlegen })
    zeigeNutzerseite()
    await effekteNachholen()

    const knopf = await screen.findByRole('button', {
      name: t.profilAnlegenOeffnen,
    })

    // **Der einzige Hauptknopf im Kopf** — wie im Artboard (R1).
    expect(knopf.className).toMatch(/(^|\s)bg-marke(\s|$)/)
    // **Und der einzige auf der Seite** — AP-3 (#2360): das Artboard
    // *Nutzer* füllt nur ihn, und je Blick führt einer (`Zustaende09`).
    // „Angaben sichern“ unter der Person steht im Nebengewicht.
    expect(
      screen.getByRole('button', { name: t.angabenSichern }),
    ).not.toHaveClass('bg-marke')
    expect(
      screen
        .getAllByRole('button')
        .filter((b) => /(^|\s)bg-marke(\s|$)/.test(b.className)),
    ).toEqual([knopf])
    expect(screen.queryByLabelText(new RegExp(t.neuesProfilName))).toBeNull()

    const schublade = await oeffneAnlage()

    await userEvent.type(
      within(schublade).getByLabelText(new RegExp(t.neuesProfilName)),
      'Turnier',
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: t.neuesProfilKnopf }),
    )

    expect(
      await within(schublade).findByText(/Angelegt: Turnier/),
    ).toBeVisible()
    expect(anlegen).toHaveBeenCalledWith('1', 'Turnier')
  })

  // **Unter der Schublade ist die Seite gesperrt** — `inert` wie bei den
  // Schubladen der Pakete und der Sammlung.
  it('sperrt die Seite, solange die Schublade steht', async () => {
    belegeSeite({})
    zeigeNutzerseite()

    const knopf = await screen.findByRole('button', {
      name: t.profilAnlegenOeffnen,
    })
    expect(knopf.closest('[inert]')).toBeNull()

    await oeffneAnlage()

    expect(knopf.closest('[inert]')).not.toBeNull()

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
    expect(knopf.closest('[inert]')).toBeNull()
  })

  // **Schließen verwirft nichts** — dieselbe Zusage wie bei den Paketen:
  // ein halb getippter Name übersteht das Zuziehen.
  it('behält den getippten Namen über das Schließen hinweg', async () => {
    belegeSeite({})
    zeigeNutzerseite()

    const schublade = await oeffneAnlage()
    await userEvent.type(
      within(schublade).getByLabelText(new RegExp(t.neuesProfilName)),
      'Liga',
    )
    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    const wieder = await oeffneAnlage()

    expect(
      within(wieder).getByLabelText(new RegExp(t.neuesProfilName)),
    ).toHaveValue('Liga')
  })

  // **Das Anlegen steht nicht mehr am Fuß der Lasche** — dort stand es unter
  // Vorgaben, Freigaben und Token.
  it('trägt in der Lasche „Profile“ kein Anlegen mehr', async () => {
    belegeSeite({})
    zeigeNutzerseite()

    await userEvent.click(
      await screen.findByRole('tab', { name: t.laschen.profil }),
    )

    expect(
      within(screen.getByRole('tabpanel')).queryByRole('button', {
        name: t.neuesProfilKnopf,
      }),
    ).toBeNull()
  })
})

// **R3 auf der Nutzerseite** — höchstens zwölf Felder gezeichnet, solange
// niemand aufklappt; gezählt mit einem Bestand, wie ihn ein Haushalt hat.
//
// **Was er nicht hält**: die Rolltiefe (R4) — jsdom rechnet kein Layout, sie
// misst das Drehbuch `uirevue` am Fenster. **Und *Persönliches* zählt hier
// ohne Einstellungen**: sie kommen aus einem Raum, den diese Vorrichtung
// nicht aufspannt; dort stehen heute fünf Auswahlen (`Persoenliches.tsx`).
describe('R3 auf der Nutzerseite', () => {
  it.each([
    t.laschen.allgemein,
    t.laschen.konto,
    t.laschen.profil,
    t.laschen.persoenlich,
  ])('%s zeichnet höchstens zwölf Felder', async (lasche) => {
    belegeVielBestand()
    zeigeNutzerseite()
    await effekteNachholen()

    await userEvent.click(await screen.findByRole('tab', { name: lasche }))
    await effekteNachholen()
    await antwortAbwarten()

    expect(gezeichneteFelder()).toBeLessThanOrEqual(12)
  })

  // **Zu heißt: nicht gezeichnet, und nicht gefragt** — der Kasten fragt den
  // Bestand erst, wenn jemand ihn öffnet.
  it('klappt Vorgaben, Freigaben und Token in „Profile“ ein', async () => {
    const token = vi.fn(() => Promise.resolve({ token: [] }))
    belegeVielBestand({ Token: token })
    zeigeNutzerseite()

    await userEvent.click(
      await screen.findByRole('tab', { name: t.laschen.profil }),
    )

    for (const titel of [t.vorgaben, t.freigaben, t.token]) {
      expect(screen.getByRole('button', { name: titel })).toHaveAttribute(
        'aria-expanded',
        'false',
      )
    }
    expect(token).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: t.token }))

    expect(
      await screen.findByRole('region', { name: t.token }),
    ).toBeInTheDocument()
    expect(token).toHaveBeenCalledWith('10')
  })
})

// **Das Alltägliche vorn, das Seltene dahinter** — die Frage, die UXR-20
// an diese Lasche stellt. Vorher begann sie mit dem Eintragen eines Kontos
// (drei Felder und ein Hauptknopf) und einem Konto ohne Angaben; Sperren und
// Abmelden standen an vierter Stelle, die Anbieter zuletzt.
describe('Die Lasche „Konto“', () => {
  it('stellt Sitzung und zweiten Faktor vor die seltenen Wege', async () => {
    belegeSeite({})
    zeigeNutzerseite()

    await userEvent.click(
      await screen.findByRole('tab', { name: t.laschen.konto }),
    )
    const feld = within(screen.getByRole('tabpanel'))

    const sitzung = await feld.findByRole('region', {
      name: t.sitzung.ueberschrift,
    })
    const faktor = feld.getByRole('region', { name: t.faktor.titel })
    const selten = [t.konto.eigenes, t.konto.ohneAngaben, t.anmeldung].map(
      (titel) => feld.getByRole('button', { name: titel }),
    )

    expect(sitzung.compareDocumentPosition(faktor)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    )
    for (const knopf of selten) {
      expect(knopf).toHaveAttribute('aria-expanded', 'false')
      expect(faktor.compareDocumentPosition(knopf)).toBe(
        Node.DOCUMENT_POSITION_FOLLOWING,
      )
    }

    // **Kein Feld eines seltenen Wegs steht da** — weder Kennung noch
    // Kennwort; das Ändern des Kennworts klappt in der Sitzung auf.
    expect(feld.queryByLabelText(t.konto.kennung)).toBeNull()
    expect(feld.queryByLabelText(t.sitzung.kennwortNeu)).toBeNull()
    expect(
      feld.getByRole('button', { name: t.sitzung.kennwortAendern }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(feld.getByRole('button', { name: t.sitzung.sperren })).toBeVisible()
  })
})

// **Verwalten ist kein Alltag** — umbenennen, löschen und einen Menschen
// anlegen standen offen unter den eigenen Angaben, mit drei Menschen und
// sieben Profilen zehn Zeilen mit „Bearbeiten" und ein zweiter Hauptknopf.
describe('Die Lasche „Allgemeines“', () => {
  it('klappt Bearbeiten und Anlegen ein', async () => {
    belegeVielBestand()
    zeigeNutzerseite()

    const feld = within(await screen.findByRole('tabpanel'))
    await feld.findByRole('region', { name: t.angaben })

    for (const titel of [t.pflege, t.neuerNutzer]) {
      expect(feld.getByRole('button', { name: titel })).toHaveAttribute(
        'aria-expanded',
        'false',
      )
    }
    expect(feld.queryByRole('button', { name: 'Anna bearbeiten' })).toBeNull()
    expect(feld.queryByLabelText(t.neuerNutzerName)).toBeNull()

    await userEvent.click(feld.getByRole('button', { name: t.pflege }))

    expect(
      feld.getByRole('button', { name: 'Anna bearbeiten' }),
    ).toBeInTheDocument()
  })
})

// **Wer hier arbeitet** — die linke Spalte des Artboards `Nutzer`. Vorher
// stand die Profilwahl in der Lasche *Profile*, die Nutzerwahl in
// *Allgemeines* — darüber dieselben Menschen noch einmal als Liste mit Bild
// und ohne Handlung.
describe('Wer hier arbeitet', () => {
  it('wechselt das Profil, ohne eine Lasche zu öffnen', async () => {
    belegeVielBestand()
    zeigeNutzerseite()

    // **Ein Profil unter den ersten vier** — seit #2022 faltet die Wahl ab
    // dem fünften; Annas Sicht stünde hinter „Weitere".
    expect(
      await screen.findByRole('button', {
        name: 'Zu Ich · Werkstatt wechseln',
      }),
    ).toBeInTheDocument()

    // **Neben den Laschen und nicht in einer** — sonst hinge der häufigste
    // Handgriff der Seite davon ab, welche gerade offen ist.
    expect(
      within(screen.getByRole('tabpanel')).queryByRole('button', {
        name: /wechseln$/,
      }),
    ).toBeNull()
  })

  it('nennt jeden Menschen einmal, mit Bild und dem Weg, er zu werden', async () => {
    belegeVielBestand()
    zeigeNutzerseite()

    const bild = await screen.findByLabelText('Anna hat kein Profilbild')

    expect(screen.getAllByLabelText('Anna hat kein Profilbild')).toHaveLength(1)
    expect(
      within(bild.closest('li') as HTMLElement).getByRole('button', {
        name: 'Anna werden',
      }),
    ).toBeInTheDocument()
  })
})
