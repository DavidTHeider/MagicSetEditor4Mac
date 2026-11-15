// **Ein Gegenüber aufnehmen und bestätigen, wo gehandelt und gespielt wird** —
// UX-03.AK-03 (#1675):
// *„Anlegen oder Bestätigen eines Kontakts ist aus dem Handels- und
// Spielworkflow möglich, ohne vorher eine technische Verwaltungsseite
// aufsuchen zu müssen."*
//
// # Der Anlass
//
// **Beide Wege endeten in den Einstellungen.** Handeln sagte ohne bestätigtes
// Gegenüber „Nimm eines unter ‚Einstellungen › Verbindungen' auf", und Spielen
// hatte für einen fremden Tisch keinen Ort, an dem man den Wirt kennenlernt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Handeln } from './Handeln'
import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { eineListe, einKontakt } from './peerkontaktprobe'
import { mitDecks } from './spielenprobe'

const texte = kataloge.de
const technik = texte.seiten.technik

afterEach(() => {
  raeumeApi()
})

describe('Handeln', () => {
  it('nimmt ein Gegenüber auf, ohne die Seite zu verlassen, und zeigt es danach zur Wahl', async () => {
    const fremd = einKontakt({
      kennung: 'peer-2',
      selbstgewaehlterName: 'Turfer',
    })

    // **Der Bestand antwortet nach seinem Stand**, nicht nach der Zahl der
    // Rufe: die Partnerwahl und die Liste im Kasten fragen beide, und in
    // welcher Reihenfolge, ist nicht Sache dieses Tests.
    let stand = fremd
    const Peerkontakte = vi.fn(() => Promise.resolve(eineListe([stand])))
    const PeerkontaktBestaetigen = vi.fn(() => {
      stand = { ...fremd, stand: 'bestaetigt' }

      return Promise.resolve(eineListe([stand]))
    })
    belegeApi({
      Peerkontakte,
      PeerkontaktBestaetigen,
      Sammlungsfreigaben: () =>
        Promise.resolve({ freigaben: [], umfaenge: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Handeln />
      </Sprachraum>,
    )

    // **Seit UXR-15 hinter „Neuer Tausch …"** (#1923): Wahl, Aufnehmen und
    // Kontakte stehen in einer Schublade, nicht mehr offen in der Spalte.
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.handeln.neuerTausch,
      }),
    )

    // **Vorher**: keiner zur Wahl, und der Satz verweist nicht mehr in die
    // Einstellungen, sondern auf den Kasten darunter.
    expect(
      await screen.findByText(texte.seiten.handeln.partnerLeer),
    ).toBeInTheDocument()
    expect(texte.seiten.handeln.partnerLeer).not.toMatch(/Einstellungen/)
    expect(texte.seiten.handeln.partnerErklaerung).not.toMatch(/Einstellungen/)

    await userEvent.click(
      screen.getByRole('button', { name: texte.bausteine.gegenueberTitel }),
    )

    // **Einladen steht vorn** (NET-01.AK-01): Code zeigen und Code
    // eingeben sind die direkten Handlungen — und das unbestätigte
    // Gegenüber steht mit seinem Vergleich darunter.
    expect(
      await screen.findByRole('button', { name: technik.einladung.zeigen }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: technik.einladung.eingeben }),
    ).toBeInTheDocument()
    const zeile = await screen.findByText(/Turfer/, { selector: 'li *' })
    expect(zeile).toBeInTheDocument()

    // **Bestätigt wird mit dem verglichenen Abdruck** — wie in den
    // Einstellungen, denn es ist dasselbe Bauteil.
    const aufrufeVorher = Peerkontakte.mock.calls.length
    const inZeile = within(zeile.closest('li') as HTMLElement)
    await userEvent.type(
      inZeile.getByLabelText(technik.kontakteVergleichen),
      fremd.fingerabdruck,
    )
    await userEvent.click(
      inZeile.getByRole('button', { name: technik.kontakteBestaetigen }),
    )

    await waitFor(() => {
      expect(PeerkontaktBestaetigen).toHaveBeenCalled()
    })

    // **Danach steht es zur Wahl**: die Partnerwahl lädt nach.
    //
    // **Beim vollen Namen des Knopfs gesucht** (UXR-16, #1924): die Zeile im
    // Kasten nennt Turfer jetzt auch in ihrem Knopf („Verwalten: Turfer"),
    // und das bloße „Turfer" fand zwei.
    await waitFor(() => {
      expect(Peerkontakte.mock.calls.length).toBeGreaterThan(aufrufeVorher)
    })
    expect(
      await screen.findByRole('button', {
        name: einsetzen(texte.seiten.handeln.partnerWaehlenName, {
          name: 'Turfer',
        }),
      }),
    ).toBeInTheDocument()
  })
})

// **Auf der Spielenseite steht der Kasten in beiden Schubladen** — seit
// UXR-13 (#1921) liegen Beitritt und Leihe hinter je einer, und unter einer
// offenen Schublade ist die Seite gesperrt.
describe('Spielen', () => {
  it('bietet beim fremden Tisch denselben Kasten an', async () => {
    mitDecks({
      Peerkontakte: () => Promise.resolve(eineListe([])),
    })

    render(
      <Sprachraum sprache="de">
        <Spielen wechseln={() => undefined} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.spielen.gastOeffnen,
      }),
    )
    await userEvent.click(
      within(
        await screen.findByRole('dialog', { name: texte.seiten.spielen.gast }),
      ).getByRole('button', {
        name: texte.bausteine.gegenueberTitel,
      }),
    )

    expect(
      await screen.findByRole('button', { name: technik.einladung.eingeben }),
    ).toBeInTheDocument()
  })

  // **Und ein hier bestätigtes Gegenüber steht gleich zur Wahl** — bei der
  // Leihe darunter, ohne die Seite neu zu öffnen (UX-08.AK-04, #1680). Ohne
  // eigenen Namen heißt es, wie es sich selbst nennt.
  it('zeigt ein hier bestätigtes Gegenüber gleich bei der Leihe', async () => {
    const fremd = einKontakt({
      kennung: 'peer-2',
      selbstgewaehlterName: 'Turfer',
    })

    let stand = fremd
    const Peerkontakte = vi.fn(() => Promise.resolve(eineListe([stand])))
    const PeerkontaktBestaetigen = vi.fn(() => {
      stand = { ...fremd, stand: 'bestaetigt' }

      return Promise.resolve(eineListe([stand]))
    })
    mitDecks({
      Peerkontakte,
      PeerkontaktBestaetigen,
      Sammlungsfreigaben: () =>
        Promise.resolve({ freigaben: [], umfaenge: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Spielen wechseln={() => undefined} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.spielen.leiheOeffnen,
      }),
    )
    const schublade = await screen.findByRole('dialog', {
      name: texte.seiten.spielen.leihe,
    })

    expect(
      await within(schublade).findByText(
        texte.seiten.spielen.leiheEntleiherLeer,
      ),
    ).toBeInTheDocument()

    await userEvent.click(
      within(schublade).getByRole('button', {
        name: texte.bausteine.gegenueberTitel,
      }),
    )

    const zeile = await screen.findByText(/Turfer/, { selector: 'li *' })
    const inZeile = within(zeile.closest('li') as HTMLElement)
    await userEvent.type(
      inZeile.getByLabelText(technik.kontakteVergleichen),
      fremd.fingerabdruck,
    )
    await userEvent.click(
      inZeile.getByRole('button', { name: technik.kontakteBestaetigen }),
    )

    expect(
      await screen.findByRole('radio', { name: /^Turfer · 1111-2222/ }),
    ).toBeInTheDocument()
  })
})
