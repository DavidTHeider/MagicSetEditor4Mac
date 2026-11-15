// **Die Ebenen erklärt, die Verwalterrechte gemessen, und was nur Fachleute
// brauchen, eingeklappt** — UX-04.AK-02, AK-03 und AK-04
// (#1676).
//
// # Was offen war
//
// `docs/PROTOTYP-0.6.0.md` nannte es beim Namen: die Suche zeigte die Ebene
// am Feld, **erklärt** war sie nirgends; ob eine Einstellung Verwalterrechte
// braucht, war **nicht gemessen**; und technische Expertenoptionen waren
// **nicht einklappbar**.
//
// # Die Messung
//
// **Keine Einstellung verlangt Verwalterrechte.** `App.EinstellungSetzen`
// ruft `Einstellungsverwalter.Setzen`, und der prüft den Wert
// (`domain.PruefeEinstellung`) und sonst nichts — nachgelesen am 2026-09-24.
// „Keine" ist damit eine Antwort und keine Lücke; sie steht in der
// Erklärung der Ebenen, wo jemand nach Rechten fragt.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import {
  eintraege,
  zeigeEinstellungen as zeige,
  zumGebiet,
} from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

const frist: EinstellungDTO = {
  schluessel: 'netz.frist.verbindung',
  wert: '',
  wirksam: '',
  werte: [],
  bereich: 'netz',
  ebene: 'geraet',
  herkunft: 'vorgabe',
  vorgabe: '10s',
  wirkung: 'sofort',
}

const proxy: EinstellungDTO = {
  ...frist,
  schluessel: 'netz.proxy.http',
  vorgabe: '',
}

const alleEinstellungen = [...eintraege, proxy, frist]

describe('die Ebenen an der Suche', () => {
  it('nennt die Ebenen in einem Satz und erklärt sie hinter dem Hilfesymbol', async () => {
    zeige()

    expect(
      screen.getByText(kurzform(eigene.ebenenErklaerung)),
    ).toBeInTheDocument()

    await userEvent.hover(
      screen.getByRole('button', { name: /Mehr zu „Jede Einstellung gilt/ }),
    )

    const erklaerung = await screen.findByRole('tooltip')
    expect(erklaerung).toHaveTextContent(/Profil/)
    expect(erklaerung).toHaveTextContent(/Verwalterrechte/)
  })
})

describe('was Fachleute hier sonst noch einstellen', () => {
  it('klappt die Fristen ein und zeigt sie auf Wunsch', async () => {
    zeige({ alleEinstellungen })

    await zumGebiet(eigene.gruppen.verbindungen)
    const gebiet = screen.getByRole('tabpanel')

    // **Der Proxy bleibt draußen** — ihn suchte man (Befund F14), und er
    // ist keine Frage nur für Fachleute.
    expect(
      within(gebiet).getByText(eigene.namen['netz.proxy.http']),
    ).toBeInTheDocument()
    expect(
      within(gebiet).queryByText(eigene.namen['netz.frist.verbindung']),
    ).toBeNull()

    const kasten = within(gebiet).getByRole('button', {
      name: eigene.fachleute,
    })
    expect(kasten).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(kasten)

    expect(kasten).toHaveAttribute('aria-expanded', 'true')
    expect(
      within(gebiet).getByText(eigene.namen['netz.frist.verbindung']),
    ).toBeInTheDocument()
  })

  // **Die Suche führt direkt zum Feld** (AK-03) — auch in einen Kasten, der
  // zu ist. Sonst fände „frist" einen Treffer, den niemand sieht.
  it('öffnet den Kasten, wenn die Suche darin trifft', async () => {
    zeige({ alleEinstellungen })

    await userEvent.type(screen.getByLabelText(eigene.suche), 'frist')

    const gebiet = screen.getByRole('tabpanel')
    expect(
      within(gebiet).getByText(eigene.namen['netz.frist.verbindung']),
    ).toBeInTheDocument()
    expect(
      within(gebiet).getByRole('button', { name: eigene.fachleute }),
    ).toHaveAttribute('aria-expanded', 'true')
  })
})
