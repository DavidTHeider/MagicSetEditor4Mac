// **Jede Einstellung sagt, wann sie wirkt** — UX-04.AK-02
// (#1676).
//
// # Der Anlass
//
// **Das Kriterium verlangt es am Feld**: „Jede Einstellung zeigt
// Geltungsbereich, wirksamen Wert, Herkunft, Standard und **Zeitpunkt der
// Wirksamkeit**". Die ersten vier stehen da; der fünfte stand am 2026-09-21
// an **einer** Stelle im ganzen Programm — als Satz unter dem Netzweg:
// „Eine geänderte Frist gilt ab dem nächsten Aufruf; ein geänderter Proxy
// sofort." Er war richtig und unsichtbar.
//
// # Woher die Auskunft kommt
//
// **Aus der Domäne und nicht aus dem Fenster** — `domain.einstellung.Wirkung`
// und `internal/domain/wirkzeitpunkt_test.go`. Ob der Proxy sofort wirkt,
// hängt daran, wie `internal/netz` ihn liest; eine Antwort je Fläche wären
// vier Antworten auf dieselbe Frage (§16).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Wahl } from './Wahl'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

function zeige(wirkung: string | undefined) {
  const eintrag: EinstellungDTO = {
    schluessel: 'netz.proxy.http',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'netz',
    ebene: 'geraet',
    herkunft: 'vorgabe',
    vorgabe: '',
    ...(wirkung === undefined ? {} : { wirkung }),
  }

  render(
    <Sprachraum sprache="de">
      <Wahl eintrag={eintrag} angehalten={false} aendern={() => undefined} />
    </Sprachraum>,
  )
}

describe('der Zeitpunkt der Wirksamkeit', () => {
  // **Was sofort wirkt, sagt es nicht mehr** — AP-11 (#2368), B-18: der
  // Satz stand unter jeder Einstellung und sagte das Selbstverständliche;
  // die Ausnahme ist der Neustart.
  it('schweigt, wenn das Umlegen genügt', () => {
    zeige('sofort')

    expect(screen.queryByText(/Wirkt sofort/)).not.toBeInTheDocument()
    expect(screen.queryByText(eigene.wirktNachNeustart)).not.toBeInTheDocument()
  })

  it('sagt, dass es einen Neustart braucht', () => {
    zeige('neustart')

    expect(screen.getByText(eigene.wirktNachNeustart)).toBeInTheDocument()
  })

  // **Ohne Angabe steht nichts da.** Ein Bestand von morgen kann einen
  // Schlüssel tragen, den diese Anwendung nicht kennt — `sofort`
  // anzunehmen wäre eine Behauptung über etwas, das sie nicht weiß (§10).
  it('schweigt, wo nichts gesagt ist', () => {
    zeige(undefined)

    expect(screen.queryByText(/Wirkt sofort/)).not.toBeInTheDocument()
    expect(screen.queryByText(eigene.wirktNachNeustart)).not.toBeInTheDocument()
  })

  // **Und ein fremdes Wort ebenso.** Es kann nur aus einer neueren Go-Seite
  // kommen; einen erfundenen Satz dafür zu zeigen wäre schlimmer als keiner.
  it('schweigt bei einem Wort, das es nicht kennt', () => {
    zeige('irgendwann')

    expect(screen.queryByText(/Wirkt sofort/)).not.toBeInTheDocument()
    expect(screen.queryByText(eigene.wirktNachNeustart)).not.toBeInTheDocument()
    expect(screen.queryByText('irgendwann')).not.toBeInTheDocument()
  })
})
