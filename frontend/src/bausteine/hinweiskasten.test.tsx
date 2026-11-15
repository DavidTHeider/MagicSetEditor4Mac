// **Ein Kasten für Hinweis, Erfolg und Befund** — Werkstück U3 von Phase 5.1
// (#1414).
//
// Geprüft wird, was der Baustein **zusagt**: die Rolle je Art, der Satz, die
// Kennung für ein `aria-describedby` und der Schließknopf. Die Farben prüft
// `docs/KONTRASTE.md` und ein Mensch — ein DOM-Test sähe sie nicht (§3).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Hinweiskasten } from './Hinweiskasten'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine

// **Ein Fehler ist ein Alarm**: er unterbricht, was gerade vorgelesen wird.
it('meldet einen Fehler als Alarm', () => {
  render(
    <Sprachraum sprache="de">
      <Hinweiskasten art="fehler" meldung="Das Muster fehlt." />
    </Sprachraum>,
  )

  expect(screen.getByRole('alert')).toHaveTextContent('Das Muster fehlt.')
  expect(screen.queryByRole('status')).toBeNull()
})

// **Erfolg und Hinweis sind kein Alarm.** Sie unterbrechen nicht — wer sie
// als `alert` ausgäbe, machte aus einer guten Nachricht eine Störung.
it.each(['erfolg', 'hinweis'] as const)(
  'meldet %s als Zustand und nicht als Alarm',
  (art) => {
    render(
      <Sprachraum sprache="de">
        <Hinweiskasten art={art} meldung="Zwei Karten geschrieben." />
      </Sprachraum>,
    )

    expect(screen.getByRole('status')).toHaveTextContent(
      'Zwei Karten geschrieben.',
    )
    expect(screen.queryByRole('alert')).toBeNull()
  },
)

// **Die Kennung ist der halbe Zweck** (§6): ein Feld zeigt mit
// `aria-describedby` darauf, und dann steht der Befund **am** Feld.
it('trägt die Kennung, mit der ein Feld auf ihn zeigt', () => {
  render(
    <Sprachraum sprache="de">
      <Hinweiskasten
        art="fehler"
        kennung="ergebnis-befund"
        meldung="Ohne ein Spiel ist es kein Ergebnis."
      />
    </Sprachraum>,
  )

  expect(screen.getByRole('alert')).toHaveAttribute('id', 'ergebnis-befund')
})

// **Ohne Rückruf kein Knopf.** Ein Schließknopf, der nichts schließt, ist eine
// Zusage ohne Rumpf (§10).
it('zeigt keinen Schließknopf, solange niemand einen bekommt', () => {
  render(
    <Sprachraum sprache="de">
      <Hinweiskasten art="hinweis" meldung="Nichts gefunden." />
    </Sprachraum>,
  )

  expect(screen.queryByRole('button')).toBeNull()
})

it('schließt auf Verlangen und nennt den Knopf', async () => {
  const schliessen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Hinweiskasten
        art="erfolg"
        meldung="Gespeichert."
        schliessen={schliessen}
      />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.aktionen.schliessen }),
  )

  expect(schliessen).toHaveBeenCalledTimes(1)
})

// **Eine Meldung bietet den nächsten Schritt an, wenn sie einen hat** —
// UX-08.AK-02 (#1680). Ohne Rückruf kein Knopf, wie beim Schließen.
it('bietet eine Handlung an und ruft sie', async () => {
  const aufKlick = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Hinweiskasten
        art="fehler"
        meldung="Der Eintrag wurde zwischenzeitlich geändert."
        handlung={{ beschriftung: texte.zustaende.neuLaden, aufKlick }}
      />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.zustaende.neuLaden }),
  )

  expect(aufKlick).toHaveBeenCalledOnce()
})
