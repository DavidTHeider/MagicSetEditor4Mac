// **Die Herkunft in einer Zeile, als Marke, wo sie abweicht** — AP-11
// (#2368),
// Befund B-18, Entwurf `Ruhe09`.
//
// Bis dahin standen unter jedem Wert zwei Sätze — „Wirksam ist … — gesetzt
// von … Ebene …“ und „Wirkt sofort.“ —, und im ersten Blick passten vier
// Einstellungen. Jetzt: nichts, wo der Wert die Vorgabe ist und sofort wirkt;
// eine Marke, wo jemand ihn gesetzt hat; „Es gilt …“, wo die Wahl erbt; der
// Neustart, wo er nötig ist. **Der ganze Satz bleibt** — als Titel der Marke
// und für das Vorleseprogramm.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import {
  Einstellungsherkunft,
  type Herkunftsangaben,
} from './Einstellungsherkunft'

const eigene = kataloge.de.seiten.einstellungen

function zeige(eintrag: Herkunftsangaben, ohneWert = false) {
  return render(
    <Sprachraum sprache="de">
      <Einstellungsherkunft eintrag={eintrag} ohneWert={ohneWert} />
    </Sprachraum>,
  )
}

const vorgabe: Herkunftsangaben = {
  schluessel: 'schriftgroesse',
  wert: 'normal',
  wirksam: 'normal',
  ebene: 'nutzer',
  herkunft: 'vorgabe',
  wirkung: 'sofort',
}

describe('die Herkunft einer Einstellung', () => {
  it('schweigt bei einem Wert der Vorgabe, der sofort wirkt', () => {
    const { container } = zeige(vorgabe)

    expect(container.querySelector('p')).toBeNull()
    expect(container.querySelector('[data-herkunftsmarke]')).toBeNull()
    expect(screen.queryByText(/Wirkt sofort/)).toBeNull()
  })

  it('setzt eine Marke, wo jemand den Wert gesetzt hat, mit dem ganzen Satz', () => {
    const { container } = zeige({
      ...vorgabe,
      herkunft: 'profil',
      ebene: 'profil',
    })

    const marke = container.querySelector('[data-herkunftsmarke]')
    expect(marke).toHaveTextContent('von diesem Profil')
    expect(marke).toHaveAttribute(
      'title',
      'Gesetzt von diesem Profil, Ebene Profil.',
    )
    // **Für das Vorleseprogramm der ganze Satz**, die Marke selbst stumm.
    expect(marke).toHaveAttribute('aria-hidden', 'true')
    expect(
      screen.getByText('Gesetzt von diesem Profil, Ebene Profil.'),
    ).toHaveClass('sr-only')
    expect(screen.queryByText(/Wirkt sofort/)).toBeNull()
  })

  it('nennt den wirksamen Wert nur, wo die Wahl erbt', () => {
    zeige({ ...vorgabe, wert: 'erben', wirksam: 'gross', herkunft: 'nutzer' })

    expect(screen.getByText('Es gilt „gross“.')).toBeInTheDocument()
  })

  // **Eine Schranke hält die Wahl** — „nur lokal“ schaltet einen Dienst ab,
  // den das Profil eingeschaltet hat (`domain.mitNurLokal`), eine strengere
  // Richtlinie gewinnt über die eigene (`domain.richtliniengeltung`). Der
  // Schalter steht dann auf „an“, und ohne diesen Satz hielte man den Dienst
  // für eingeschaltet.
  it('sagt, wenn eine Richtlinie die Wahl hält, und was dann gilt', () => {
    const { container } = zeige({
      ...vorgabe,
      schluessel: 'dienst.scryfall',
      wert: 'an',
      wirksam: 'aus',
      herkunft: 'nutzer',
      ebene: 'profil',
    })

    expect(
      screen.getByText('Es gilt „aus“ — eine Richtlinie hält es.'),
    ).toBeInTheDocument()
    expect(container.querySelector('[data-herkunftsmarke]')).toHaveTextContent(
      'von dir',
    )
  })

  it('sagt nichts über den Wert, wo die Wahl gilt', () => {
    zeige({ ...vorgabe, wert: 'gross', wirksam: 'gross', herkunft: 'profil' })

    expect(screen.queryByText(/Es gilt/)).toBeNull()
  })

  it('sagt, wenn eine geerbte Wahl beim Rechner endet', () => {
    zeige({ ...vorgabe, wert: 'erben', wirksam: '' })

    expect(screen.getByText(eigene.wirksamOffenKurz)).toBeInTheDocument()
  })

  it('nennt keinen Wert, wo er im Feld daneben steht', () => {
    zeige(
      { ...vorgabe, wert: 'erben', wirksam: 'gross', herkunft: 'nutzer' },
      true,
    )

    expect(screen.queryByText(/Es gilt/)).toBeNull()
  })

  it('sagt den Neustart in derselben Zeile', () => {
    const { container } = zeige({
      ...vorgabe,
      herkunft: 'geraet',
      ebene: 'geraet',
      wirkung: 'neustart',
    })

    const zeile = container.querySelector(
      '[data-herkunftsmarke]',
    )?.parentElement
    expect(zeile).toHaveTextContent(eigene.wirktNachNeustart)
  })
})
