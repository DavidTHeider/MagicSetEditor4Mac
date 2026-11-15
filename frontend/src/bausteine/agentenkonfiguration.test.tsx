// **Fertige Konfiguration für einen Agenten** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Was hier geprüft wird
//
// **Dass dasteht, was jemand einfügt** — und dass daneben steht, wohin es
// gehört und woher die Gestalt stammt. Ohne Bündel gibt es kein `mse-mcp`,
// und dann steht der Kasten gar nicht da (§6).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Agentenkonfiguration } from './Agentenkonfiguration'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'

import type { AgentenbruchstueckDTO } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.technik

// einBruchstueck baut eines für Claude Code.
function einBruchstueck(): AgentenbruchstueckDTO {
  return {
    agent: 'claude',
    datei: '~/.claude.json',
    form: 'json',
    inhalt: '{\n  "mcpServers": {}\n}',
    befehl:
      'claude mcp add --transport stdio magicseteditor4mac -- "/x/mse-mcp"',
    nachgelesen: 'code.claude.com/docs/en/mcp, abgerufen am 2026-09-13',
  }
}

// zeige stellt den Kasten hin.
function zeige(...stuecke: AgentenbruchstueckDTO[]) {
  render(
    <Sprachraum sprache="de">
      <Agentenkonfiguration agenten={stuecke} />
    </Sprachraum>,
  )
}

// **Das Bruchstück steht da — samt Ort, Befehl und Fundstelle.**
it('zeigt Bruchstück, Ort, Befehl und Fundstelle', () => {
  const stueck = einBruchstueck()

  zeige(stueck)

  // **Über den Textinhalt und nicht über `getByText`**: Testing Library
  // normalisiert Leerraum, und ein eingerücktes Bruchstück steht danach in
  // einer Zeile — die Zusicherung prüfte dann etwas anderes als das, was
  // jemand einfügt.
  expect(
    screen.getByRole('region', { name: eigene.agenten }),
  ).toHaveTextContent('"mcpServers"')

  expect(screen.getByText(stueck.befehl)).toBeInTheDocument()

  expect(
    screen.getByText(einsetzen(eigene.agentenDatei, { datei: stueck.datei })),
  ).toBeInTheDocument()

  // **Die Fundstelle ist keine Zierde** (§10): eine Konvention aus dem
  // Gedächtnis ist kein Prüfergebnis, und wer sie morgen nachsieht, soll
  // wissen, wo und wann.
  expect(
    screen.getByText(
      einsetzen(eigene.agentenNachgelesen, { quelle: stueck.nachgelesen }),
    ),
  ).toBeInTheDocument()
})

// **Ein Agent ohne Namen steht mit seiner Marke da.**
//
// **Die letzte Reihe und nicht der Normalfall** (§6): lieber `neuling` lesen
// als eine Überschrift ohne Text. Dass die beiden gebauten einen Namen haben,
// hält `internal/app/agentenkonfiguration_test.go`.
it('fällt ohne Namen auf die Marke zurück', () => {
  zeige({ ...einBruchstueck(), agent: 'neuling' })

  expect(screen.getByRole('heading', { name: 'neuling' })).toBeInTheDocument()
})

// **Und der Satz, dass kein Schlüssel darin steht.**
//
// **Er ist die Zusage, die den Kasten erlaubt**: ein Zugangsschlüssel in
// einer Konfigurationsdatei läge dort im Klartext und wanderte mit ihr in
// jedes Repositorium.
it('sagt, dass kein Zugangsschlüssel darin steht', () => {
  zeige(einBruchstueck())

  expect(screen.getByText(eigene.agentenOhneGeheimnis)).toBeInTheDocument()
})

// **Ohne Bruchstück steht der Kasten gar nicht da.**
//
// **Kein Kasten ohne Inhalt** (§6): wer die Anwendung aus dem Quelltext
// startet, hat kein `mse-mcp` — eine Überschrift über nichts verspräche
// etwas, was nicht kommt.
it('steht ohne Bruchstück gar nicht da', () => {
  zeige()

  expect(screen.queryByText(eigene.agenten)).not.toBeInTheDocument()
})

// **Kopiert wird auf Verlangen — und das Nein steht auch da.**
it('kopiert das Bruchstück und meldet ein Nein', async () => {
  const stueck = einBruchstueck()

  zeige(stueck)

  // **Erst `setup`, dann die Ablehnung** — und das ist keine Feinheit:
  // `userEvent.setup()` legt eine **eigene** Zwischenablage an und
  // überschreibt dabei, was vorher dastand. Wer sie davor setzt, prüft den
  // Stub von Testing Library, und der gelingt immer. Beim ersten Anlauf
  // stand deshalb „Kopiert." da, wo ein Nein erwartet wurde.
  const hand = userEvent.setup()

  setzeZwischenablage(vi.fn(() => Promise.reject(new Error('keine Erlaubnis'))))

  // **Ohne Erlaubnis schlägt das Kopieren fehl** — und genau das soll
  // dastehen, statt still nichts zu tun.
  await hand.click(
    screen.getByRole('button', { name: eigene.agentenKopierenDatei }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    eigene.kopierfehler,
  )
})

// **Mit Erlaubnis steht »Kopiert« am Knopf.**
it('meldet am Knopf, dass kopiert wurde', async () => {
  zeige(einBruchstueck())

  const hand = userEvent.setup()

  const schreiben = vi.fn(() => Promise.resolve())

  setzeZwischenablage(schreiben)

  await hand.click(
    screen.getByRole('button', { name: eigene.agentenKopierenBefehl }),
  )

  expect(await screen.findByText(eigene.kopiert)).toBeInTheDocument()

  // **Und zwar der Befehl und nicht das Bruchstück** — zwei Knöpfe, zwei
  // Inhalte, und eine Verwechslung fiele niemandem auf.
  expect(schreiben).toHaveBeenCalledWith(einBruchstueck().befehl)

  // **Der zweite Knopf meldet für sich** — „Kopiert" gehört an den, der
  // gedrückt wurde, und nicht an beide.
  await hand.click(
    screen.getByRole('button', { name: eigene.agentenKopierenDatei }),
  )

  expect(await screen.findByText(eigene.kopiert)).toBeInTheDocument()
  expect(schreiben).toHaveBeenLastCalledWith(einBruchstueck().inhalt)
})

// setzeZwischenablage legt eine eigene an.
//
// **Über `defineProperty` und nicht `Object.assign`**: nach
// `userEvent.setup()` ist `navigator.clipboard` ein **Getter**, und eine
// Zuweisung darauf wirft „Cannot set property clipboard of #<Navigator> which
// has only a getter". Beim ersten Anlauf sah das aus wie ein stiller
// Fehlschlag des Doppels.
function setzeZwischenablage(schreiben: () => Promise<void>) {
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: schreiben },
  })
}
