// **Die Kontextleiste: wer wo als was arbeitet, jederzeit sichtbar** —
// Werkstück U4 von Phase 5.1
// (#1415).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Kontextleiste } from './Kontextleiste'
import { Sprachraum } from '../sprache/kontext'
import { einNutzer } from '../test/anmeldebelegung'
import { belegeApi, raeumeApi } from '../test/api'

import type { ArbeitskontextAntwort } from '../apikontext'
import type { NutzerDTO } from '../apinutzer'

afterEach(() => {
  raeumeApi()
})

function nutzer(
  id: string,
  name: string,
  profilID: string,
  profilname: string,
): NutzerDTO {
  const vorlage = einNutzer()

  return {
    ...vorlage,
    id,
    name,
    profile: [
      { ...vorlage.profile[0], id: profilID, nutzerID: id, name: profilname },
    ],
  } as NutzerDTO
}

function kontext(
  teile: Partial<ArbeitskontextAntwort> = {},
): ArbeitskontextAntwort {
  return {
    nutzerID: '1',
    profilID: '10',
    eignerID: '1',
    rolle: 'verwalter',
    nurLesend: false,
    modus: 'fenster',
    revision: 3,
    fehler: null,
    ...teile,
  }
}

function belege(teile: Partial<ArbeitskontextAntwort> = {}) {
  belegeApi({
    Arbeitskontext: vi.fn().mockResolvedValue(kontext(teile)),
    Nutzer: vi.fn().mockResolvedValue({
      nutzer: [
        nutzer('1', 'Anna', '10', 'Turnier'),
        nutzer('2', 'Ben', '20', 'Sammlung'),
      ],
      fehler: null,
    }),
    Nutzerbilder: vi
      .fn()
      .mockResolvedValue({ bilder: {}, badges: {}, fehler: null }),
  })
}

function zeige(waagerecht = false) {
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Kontextleiste wechseln={wechseln} waagerecht={waagerecht} />
    </Sprachraum>,
  )

  return wechseln
}

describe('Kontextleiste', () => {
  it('nennt den Menschen, sein Profil und seine Rolle', async () => {
    belege()
    zeige()

    // **Gewartet wird auf die Wirkung, nicht auf den Kasten** (§3): die
    // Region steht schon, während geladen wird.
    await screen.findByText('Anna · Turnier')
    const leiste = screen.getByRole('region', { name: 'Arbeitskontext' })

    expect(leiste).toHaveTextContent('Anna')
    expect(leiste).toHaveTextContent('Verwalter')
    expect(leiste).not.toHaveTextContent('nur lesen')
  })

  // **Eine Zeile, und sie ist der Knopf** — AP-11 (#2368), B-31: „Bild +
  // Name, Rolle im Tooltip“. Neben einem eigenen Knopf blieb in der Leiste
  // von „Anna · Turnier“ nur „Anna · T…“; jetzt hat der Name die Breite,
  // und der Weg zu den Nutzern ist die Zeile selbst.
  it('steht als eine Zeile, die zu den Nutzern führt, die Rolle für den Vorleser', async () => {
    belege()
    zeige()

    const zeile = await screen.findByText('Anna · Turnier')
    expect(zeile).toHaveAttribute('title', 'Anna · Turnier')
    const knopf = screen.getByRole('button', { name: /^Anna · Turnier/ })
    expect(knopf).toContainElement(zeile)
    expect(knopf).toHaveAttribute('title', expect.stringContaining('Verwalter'))
    expect(knopf).toHaveAccessibleDescription('Profil wechseln …')
    expect(screen.getByText('Verwalter')).toHaveClass('sr-only')
    // **Das Bild spricht nicht mit** — der Name steht schon im Text.
    expect(knopf).not.toHaveAccessibleName(/Profilbild/)
  })

  // **Person A im Profil von B bleibt A** — und B steht als Eigner dabei.
  it('zeigt im fremden Profil den Gast als arbeitend und den Eigner dazu', async () => {
    belege({
      nutzerID: '1',
      profilID: '20',
      eignerID: '2',
      rolle: 'leser',
      nurLesend: true,
    })
    zeige()

    await screen.findByText('Anna · Sammlung')
    const leiste = screen.getByRole('region', { name: 'Arbeitskontext' })

    expect(leiste).toHaveTextContent('Anna')
    expect(leiste).toHaveTextContent('von Ben')
    expect(leiste).toHaveTextContent('Leser')
    expect(leiste).toHaveTextContent('nur lesen')
  })

  // **Wessen Profil es ist, steht sichtbar da** — und zwar auch, wo man
  // schreiben darf: „Anna · Sammlung“ allein läse sich wie das eigene
  // Profil, und dann fehlte auch „nur lesen“ als Zeichen. Einmal, nicht
  // noch einmal für den Vorleser.
  it('zeigt den Eigner eines fremden Profils sichtbar, auch beim Mitarbeiter', async () => {
    belege({
      nutzerID: '1',
      profilID: '20',
      eignerID: '2',
      rolle: 'mitarbeiter',
      nurLesend: false,
    })
    zeige()

    await screen.findByText('Anna · Sammlung')

    expect(screen.getAllByText('von Ben')).toHaveLength(1)
    expect(screen.getByText('von Ben')).not.toHaveClass('sr-only')
    expect(screen.getByText('von Ben')).toHaveAttribute('title', 'von Ben')
  })

  // **Oben in der Zeile** — UX-14.AK-02 (#1686): die Angaben stehen für den
  // Vorleser, sichtbar bleibt das Bild — und „nur lesen" für sich, denn das
  // ändert, was man hier tun kann. Einmal angesagt, nicht zweimal.
  it('zeigt oben „nur lesen" für sich und nennt es dem Vorleser einmal', async () => {
    belege({
      nutzerID: '1',
      profilID: '20',
      eignerID: '2',
      rolle: 'leser',
      nurLesend: true,
    })
    zeige(true)

    await screen.findByText('Sammlung')
    const leiste = screen.getByRole('region', { name: 'Arbeitskontext' })

    expect(leiste).toHaveTextContent('Anna')
    expect(within(leiste).getByText('nur lesen')).toHaveAttribute(
      'aria-hidden',
      'true',
    )
  })

  // **Links steht es sichtbar für sich** — AP-11 (#2368): im schmalen Block
  // bleibt „nur lesen“ zu sehen, denn es ändert, was man tun kann.
  // Angesagt wird es einmal: die Rolle für den Vorleser nennt es nicht noch
  // einmal.
  it('zeigt „nur lesen" links sichtbar, einmal', async () => {
    belege({ nurLesend: true, rolle: 'leser' })
    zeige()

    await screen.findByText('Anna · Turnier')

    expect(screen.getAllByText('nur lesen')).toHaveLength(1)
    expect(screen.getByText('nur lesen')).not.toHaveClass('sr-only')
  })

  it('führt zu den Nutzern und Profilen', async () => {
    belege()
    const wechseln = zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: /^Anna · Turnier/ }),
    )

    expect(wechseln).toHaveBeenCalledExactlyOnceWith('nutzer')
  })

  // **Ohne Kontext bleibt der Weg** — wie vor AP-11 steht dann der Knopf
  // mit seinem ausgeschriebenen Ziel da.
  it('führt auch ohne lesbaren Kontext zu den Nutzern', async () => {
    belegeApi({
      Arbeitskontext: vi.fn().mockResolvedValue({
        ...kontext(),
        fehler: { code: 'unbekannt', meldung: 'kein Bestand', feld: '' },
      }),
      Nutzer: vi.fn().mockResolvedValue({ nutzer: [], fehler: null }),
      Nutzerbilder: vi
        .fn()
        .mockResolvedValue({ bilder: {}, badges: {}, fehler: null }),
    })
    const wechseln = zeige()

    await screen.findByRole('alert')
    await userEvent.click(
      screen.getByRole('button', { name: 'Profil wechseln …' }),
    )

    expect(wechseln).toHaveBeenCalledExactlyOnceWith('nutzer')
  })

  // **Oben führt derselbe Weg dorthin** — die Zeile hat seit AP-11 ihren
  // eigenen Knopf, mit dem ausgeschriebenen Ziel.
  it('führt auch oben zu den Nutzern und Profilen', async () => {
    belege()
    const wechseln = zeige(true)

    await userEvent.click(
      await screen.findByRole('button', { name: 'Profil wechseln …' }),
    )

    expect(wechseln).toHaveBeenCalledExactlyOnceWith('nutzer')
  })

  // **Ohne Namen bleiben die Kennungen** — die Leiste behauptet keinen
  // Menschen, den sie nicht kennt.
  it('nennt die Kennung, wenn die Liste den Menschen nicht führt', async () => {
    belege({ nutzerID: '9', profilID: '90', eignerID: '9' })
    zeige()

    await screen.findByText('9 · 90')
    const leiste = screen.getByRole('region', { name: 'Arbeitskontext' })

    expect(leiste).toHaveTextContent('9')
  })

  // **Links wie oben** — die Zeile oben hat seit AP-11 ihren eigenen Zweig.
  it.each([
    ['links', false],
    ['oben', true],
  ])('sagt %s, wenn der Kontext nicht zu lesen ist', async (_lage, oben) => {
    belegeApi({
      Arbeitskontext: vi.fn().mockResolvedValue({
        ...kontext(),
        fehler: { code: 'unbekannt', meldung: 'kein Bestand', feld: '' },
      }),
      Nutzer: vi.fn().mockResolvedValue({ nutzer: [], fehler: null }),
      Nutzerbilder: vi
        .fn()
        .mockResolvedValue({ bilder: {}, badges: {}, fehler: null }),
    })
    zeige(oben)

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Bestand')
  })

  // **Bild und Abzeichen kommen aus derselben Antwort wie die Namen.**
  it('zeigt das Bild des Menschen und das Abzeichen des Profils', async () => {
    belegeApi({
      Arbeitskontext: vi.fn().mockResolvedValue(kontext()),
      Nutzer: vi.fn().mockResolvedValue({
        nutzer: [nutzer('1', 'Anna', '10', 'Turnier')],
        fehler: null,
      }),
      Nutzerbilder: vi.fn().mockResolvedValue({
        bilder: { '1': 'data:image/png;base64,AA==' },
        badges: { '10': 'data:image/png;base64,BB==' },
        fehler: null,
      }),
    })
    zeige()

    // **In der Zeile stumm** (AP-11, #2368): sie ist ein Knopf, und Name und
    // Profil stehen darin als Text.
    expect(
      await screen.findByRole('img', {
        name: 'Profilbild von Anna',
        hidden: true,
      }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('img', {
        name: 'Abzeichen des Profils von Anna',
        hidden: true,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('img')).toBeNull()
  })

  // **Eine Rolle, die der Katalog nicht kennt, bleibt, wie sie kam** — sie
  // stammt aus dem Bestand, und ein leeres Wort verschwiege sie.
  it('nennt eine unbekannte Rolle beim Namen', async () => {
    belege({ rolle: 'kurator' })
    zeige()

    await screen.findByText('kurator')

    expect(
      screen.getByRole('region', { name: 'Arbeitskontext' }),
    ).toHaveTextContent('Anna')
  })

  it('zeigt, dass geladen wird, bis die Antwort da ist', () => {
    belegeApi({
      Arbeitskontext: vi.fn(
        () => new Promise<ArbeitskontextAntwort>(() => undefined),
      ),
      Nutzer: vi.fn(() => new Promise<never>(() => undefined)),
      Nutzerbilder: vi.fn(() => new Promise<never>(() => undefined)),
    })
    zeige()

    expect(
      screen.getByRole('region', { name: 'Arbeitskontext' }),
    ).toHaveTextContent('Wird geladen')
  })
})
