// **Ein Pflichtfeld sagt am Feld, dass es eines ist** — Werkstück F von
// Phase 4.10
// (#1325).
//
// # Was vorher war
//
// Sechs Formulare haben ein Pflichtfeld — Deck, Format, Turnier, Teilnehmer,
// Autoersetzungsregel, Set. Bleibt es leer, **ruht der Knopf**: die
// Rückmeldung ist also da, nur nicht **am Feld**, wie §6 es verlangt. Wer
// den ausgegrauten Knopf sieht, weiß nicht, welches der drei Felder fehlt.
//
// # Und warum erst nach der Berührung
//
// **Ein leeres Formular ist kein Fehler, sondern ein leeres Formular.** Ein
// Hinweis, der beim Aufschlagen schon dasteht, ist eine Anklage für etwas,
// das niemand getan hat — und er stumpft ab, bis ihn keiner mehr liest.
import { useState } from 'react'

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Pflichtfeld } from './Pflichtfeld'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const worte = kataloge.de.bausteine

function zeige(zusatz: Record<string, unknown> = {}) {
  const geaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Pflichtfeld
        beschriftung="Name"
        wert=""
        geaendert={geaendert}
        {...zusatz}
      />
    </Sprachraum>,
  )

  return geaendert
}

// **Eine Kennung korrigiert macOS nicht** — gemessen am nativen Fenster
// (Hüllenprobe `fmt02d`): aus dem Kurznamen „namen“ machte die
// Rechtschreibkorrektur beim Verlassen des Felds „Namen“, und das Paket hieß
// danach anders als getippt. Ein Name bleibt dagegen korrigierbar.
it('schaltet die Korrektur nur für eine Kennung ab', () => {
  zeige({ kennung: true })
  expect(screen.getByLabelText('Name')).toHaveAttribute('spellcheck', 'false')
})

it('lässt einen gewöhnlichen Namen korrigierbar', () => {
  zeige()
  expect(screen.getByLabelText('Name')).not.toHaveAttribute('spellcheck')
})

it('meldet ein leeres Feld nicht, solange niemand es angefasst hat', () => {
  zeige()

  const feld = screen.getByLabelText('Name')

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByRole('alert')).toBeNull()
})

it('meldet es, sobald jemand es leer verlässt', async () => {
  zeige()

  const feld = screen.getByLabelText('Name')

  await userEvent.click(feld)
  await userEvent.tab()

  expect(feld).toHaveAttribute('aria-invalid', 'true')

  const hinweis = screen.getByRole('alert')
  expect(hinweis.textContent).toContain('Name')
  // **Das Ziel liegt im Alarm, nicht auf ihm** (#1989): zeigt
  // `aria-describedby` auf ein Element mit `role="alert"`, liest VoiceOver es
  // nicht vor.
  expect(feld.getAttribute('aria-describedby')).toBe(
    hinweis.firstElementChild?.id,
  )
})

// **Über einen Wrapper mit Zustand.** Das Feld ist **gesteuert**: ohne
// nachgeführten Wert käme jeder Tastendruck einzeln an, und der Hinweis
// hinge weiter am leeren `wert` — der Test prüfte dann eine Komponente, die
// es so nicht gibt (§3: eine Vorrichtung, die großzügiger ist als der
// Betrieb).
function MitZustand() {
  const [wert, setzeWert] = useState('')

  return <Pflichtfeld beschriftung="Name" wert={wert} geaendert={setzeWert} />
}

it('nimmt den Hinweis zurück, sobald etwas dasteht', async () => {
  render(
    <Sprachraum sprache="de">
      <MitZustand />
    </Sprachraum>,
  )

  const feld = screen.getByLabelText('Name')

  await userEvent.click(feld)
  await userEvent.tab()

  expect(screen.getByRole('alert')).toBeInTheDocument()

  await userEvent.type(feld, 'Weiß-Blau')

  expect(feld).toHaveValue('Weiß-Blau')
  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByRole('alert')).toBeNull()
})

// **Ein gefülltes Feld meldet nie** — auch nicht, wenn man es verlässt.
it('meldet ein gefülltes Feld nicht', async () => {
  zeige({ wert: 'Weiß-Blau' })

  const feld = screen.getByLabelText('Name')

  await userEvent.click(feld)
  await userEvent.tab()

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByRole('alert')).toBeNull()
})

// **Der Satz nennt das Feld**, damit er auch dann trägt, wenn drei
// nebeneinanderstehen.
it('nennt das Feld in seinem Satz', async () => {
  zeige({ beschriftung: 'Scryfall-Name' })

  await userEvent.click(screen.getByLabelText('Scryfall-Name'))
  await userEvent.tab()

  expect(screen.getByRole('alert').textContent).toContain('Scryfall-Name')
  expect(screen.getByRole('alert').textContent).toContain(
    worte.pflichtfeldLeer.slice(worte.pflichtfeldLeer.indexOf('}') + 1).trim(),
  )
})

// **Ein `ref` reicht durch bis ans Feld** — Werkstück F von Phase 4.10.
//
// ⌘N und der Menüeintrag „Neues Set" setzen den Fokus in das Namensfeld
// (`SetNeu.tsx`). Ohne durchgereichten `ref` griffe der Befehl ins Leere —
// dieselbe Klasse wie beim ersten Zuschnitt der Setseite, wo das Feld hinter
// einem Aufklapp lag.
it('reicht einen ref bis ans Feld durch', () => {
  const griff = { current: null as HTMLInputElement | null }

  render(
    <Sprachraum sprache="de">
      <Pflichtfeld
        beschriftung="Name"
        wert=""
        griff={griff}
        geaendert={vi.fn()}
      />
    </Sprachraum>,
  )

  expect(griff.current).toBe(screen.getByLabelText('Name'))
})
