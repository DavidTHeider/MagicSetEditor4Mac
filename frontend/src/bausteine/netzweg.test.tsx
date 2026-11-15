// **Der Kasten über den Netzweg** — Werkstück N5 von Phase 5.4
// (#1571),
// umgebaut für UX-15.AK-04 und AK-05
// (#1771).
//
// # Was er seit dem 2026-09-22 **nicht** mehr zeigt
//
// **Die Werte selbst.** Er stand in der Gruppe *Verbindungen*, und unter ihm
// standen dieselben sechs Werte als **Einstellungen** — dieselbe Sache
// zweimal, einmal als Auskunft und einmal zum Ändern. Das ist der Fall, für
// den es #1771
// gibt.
//
// # Was er zeigt
//
// **Nur, was kein Feld sagen kann**: dass ein Wert aus der **Umgebung dieses
// Prozesses** kommt und das Feld darunter übersteuert. Kommt keiner von
// dort, steht hier nichts — und die Felder sagen alles.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Netzweg } from './Netzweg'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.technik

const leer = {
  proxyHTTP: { wert: '', herkunft: 'ohne' },
  proxyHTTPS: { wert: '', herkunft: 'ohne' },
  proxyAus: { wert: '', herkunft: 'ohne' },
  fristVerbindung: 15,
  fristKopf: 30,
  fristStillstand: 60,
  proxyBenutzt: false,
  fehler: null,
}

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Netzweg />
    </Sprachraum>,
  )
}

// **Ohne Umgebung steht hier nichts** — die Felder darunter sagen alles, und
// ein Kasten, der sie wiederholt, macht die Seite lang.
it('zeigt nichts, wo nichts aus der Umgebung kommt', async () => {
  belegeApi({ Netzstand: vi.fn().mockResolvedValue(leer) })

  zeige()

  await antwortAbwarten()

  // **Kein Kopf und keine Liste** — der Sprachraum legt eigene Stilknoten in
  // den Rahmen, und „leer" wäre eine Aussage über die, nicht über uns.
  expect(screen.queryByRole('heading')).not.toBeInTheDocument()
  expect(screen.queryByRole('list')).not.toBeInTheDocument()
})

// **Ein Wert aus der Umgebung steht da, samt seinem Namen.**
it('nennt jeden Wert, der aus der Umgebung kommt', async () => {
  belegeApi({
    Netzstand: vi.fn().mockResolvedValue({
      ...leer,
      proxyHTTPS: { wert: 'http://proxy.firma:3128', herkunft: 'umgebung' },
      proxyBenutzt: true,
    }),
  })

  zeige()

  expect(
    await screen.findByText(
      einsetzen(texte.netzwegAusUmgebung, {
        name: texte.netzwegProxyHTTPS,
        wert: 'http://proxy.firma:3128',
      }),
    ),
  ).toBeInTheDocument()
})

// **Und ein Wert aus der Einstellung steht nicht da** — er steht im Feld
// darunter, und zweimal wäre einmal zu viel (AK-04).
it('wiederholt keinen Wert, den ein Feld schon trägt', async () => {
  belegeApi({
    Netzstand: vi.fn().mockResolvedValue({
      ...leer,
      proxyAus: { wert: 'gpu.intern', herkunft: 'einstellung' },
      proxyBenutzt: true,
    }),
  })

  zeige()

  await antwortAbwarten()

  expect(screen.queryByRole('heading')).not.toBeInTheDocument()
  expect(screen.queryByText(/gpu\.intern/)).not.toBeInTheDocument()
})

// **Ein Lesefehler steht da und bleibt stehen** — er ist eine Auskunft, die
// kein Feld hat.
it('meldet einen Lesefehler', async () => {
  belegeApi({
    Netzstand: vi.fn().mockResolvedValue({
      ...leer,
      fehler: {
        code: 'unbekannt',
        schluessel: '',
        meldung: 'Der Netzstand ist nicht zu lesen.',
        feld: '',
      },
    }),
  })

  zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Der Netzstand ist nicht zu lesen.',
  )
})
