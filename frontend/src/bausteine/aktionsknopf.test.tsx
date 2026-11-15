// **Ein Knopf aus dem Aktionskatalog** — UX-10.AK-01
// (#1682).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { aktionen, aktionsarten } from '../aktionen'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge, sprachen } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'

describe('Aktionsknopf', () => {
  it('trägt Symbol, Aufschrift und Bedeutung aus dem Katalog', async () => {
    const aufKlick = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Aktionsknopf aktion="loeschen" aufKlick={aufKlick}>
          Deck löschen
        </Aktionsknopf>
      </Sprachraum>,
    )

    // **Das Symbol ist stumm**: der Name ist die Aufschrift allein.
    const knopf = screen.getByRole('button', { name: 'Deck löschen' })
    expect(knopf).toHaveTextContent(aktionsarten.loeschen.symbol)
    expect(knopf).toHaveAttribute('data-aktion', 'loeschen')

    // **Die Bedeutung beschreibt ihn** — als Tooltip, der auch ohne Zeiger im
    // Baum steht.
    expect(knopf).toHaveAccessibleDescription(
      kataloge.de.bausteine.aktionskatalog.aktionen.loeschen.heisst,
    )

    // **Und ohne `mitKuerzel` kündigt er keines an.**
    expect(knopf).not.toHaveAttribute('aria-keyshortcuts')

    await userEvent.click(knopf)
    expect(aufKlick).toHaveBeenCalledTimes(1)
  })

  it('kündigt das Kürzel an, wo sein Aufrufer es verdrahtet hat', () => {
    render(
      <Sprachraum sprache="de">
        <Aktionsknopf aktion="zuruecknehmen" mitKuerzel aufKlick={vi.fn()} />
      </Sprachraum>,
    )

    // **Ohne Aufschrift steht das Kurzlabel des Katalogs da.**
    const eintrag = kataloge.de.bausteine.aktionskatalog.aktionen.zuruecknehmen
    const knopf = screen.getByRole('button', { name: eintrag.kurz })

    expect(knopf).toHaveAttribute('aria-keyshortcuts', 'Meta+Z')
    expect(knopf).toHaveAccessibleDescription(`${eintrag.heisst} ⌘Z`)
  })

  it('nennt die Sache, wo die Aufschrift sie nicht nennt', () => {
    render(
      <Sprachraum sprache="de">
        <Aktionsknopf
          aktion="entfernen"
          bezeichnung="Bild entfernen: Bo"
          aufKlick={vi.fn()}
        >
          Entfernen
        </Aktionsknopf>
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: 'Bild entfernen: Bo' }),
    ).toBeInTheDocument()
  })

  it('ruht, solange sein Vorgang läuft', () => {
    render(
      <Sprachraum sprache="de">
        <Aktionsknopf aktion="sichern" angehalten aufKlick={vi.fn()} />
      </Sprachraum>,
    )

    expect(screen.getByRole('button')).toBeDisabled()
  })

  // **Schreiben ist eine Angabe am Knopf und keine des Katalogs** — G10
  // (#2276): „Entfernen" nimmt am Paket etwas vom Rechner und an einem Deck
  // etwas aus dem Profil, und nur das zweite verbietet die Rolle.
  it('sperrt für einen Leser nur, was sein Aufrufer schreibend nennt', async () => {
    const geklickt: string[] = []

    render(
      <Sprachraum sprache="de">
        <Schreibrechtraum nurLesend>
          <Aktionsknopf
            aktion="loeschen"
            schreibt
            aufKlick={() => {
              geklickt.push('deck')
            }}
          >
            Deck löschen
          </Aktionsknopf>
          <Aktionsknopf
            aktion="entfernen"
            aufKlick={() => {
              geklickt.push('paket')
            }}
          >
            Paket entfernen
          </Aktionsknopf>
        </Schreibrechtraum>
      </Sprachraum>,
    )

    const deck = screen.getByRole('button', { name: 'Deck löschen' })
    expect(deck).toHaveAttribute('aria-disabled', 'true')

    // **Die Bedeutung bleibt, und der Grund kommt dazu.**
    expect(deck).toHaveAccessibleDescription(
      `${kataloge.de.bausteine.aktionskatalog.aktionen.loeschen.heisst} ${kataloge.de.bausteine.nurlesen.knopf}`,
    )

    await userEvent.click(deck)
    await userEvent.click(
      screen.getByRole('button', { name: 'Paket entfernen' }),
    )
    expect(geklickt).toEqual(['paket'])
  })
})

describe('Der Katalog', () => {
  // **Ein Symbol je Aktion und ein Kürzel je Aktion** — sonst hieße dasselbe
  // Zeichen zweierlei, und genau das soll der Katalog verhindern.
  it('vergibt kein Symbol und kein Kürzel zweimal', () => {
    const symbole = aktionen.map((aktion) => aktionsarten[aktion].symbol)
    const tasten = aktionen.map((aktion) => aktionsarten[aktion].tasten)

    expect(new Set(symbole).size).toBe(aktionen.length)
    expect(new Set(tasten).size).toBe(aktionen.length)
  })

  it.each(sprachen)(
    'hat in %s für jede Aktion Kurzlabel und Bedeutung',
    (sprache) => {
      for (const aktion of aktionen) {
        const eintrag =
          kataloge[sprache].bausteine.aktionskatalog.aktionen[aktion]

        expect(eintrag.kurz.length).toBeGreaterThan(0)
        expect(eintrag.heisst.length).toBeGreaterThan(eintrag.kurz.length)
      }
    },
  )
})
