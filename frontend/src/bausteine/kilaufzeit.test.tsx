// **Die lokale KI-Laufzeit einrichten** — Werkstück A7 von Phase 5.3
// (#1446).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { KILaufzeit } from './KILaufzeit'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

// dieZeilen sind die zwei, die es gibt.
const dieZeilen: EinstellungDTO[] = [
  {
    schluessel: 'ki.laufzeit',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
  {
    schluessel: 'ki.adresse',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'dienste',
  },
]

// zeige rendert das Bauteil und gibt den Spion auf »setze«.
function zeige(zeilen: EinstellungDTO[] = dieZeilen, angehalten = false) {
  const setze = vi.fn()

  render(
    <Sprachraum sprache="de">
      <KILaufzeit zeilen={zeilen} angehalten={() => angehalten} setze={setze} />
    </Sprachraum>,
  )

  return setze
}

// **Beide Bedienelemente stehen da, mit ihrem Hinweis.**
it('zeigt Schalter und Adresse samt Hinweis', () => {
  zeige()

  expect(screen.getByLabelText(eigene.namen['ki.laufzeit'])).toBeInTheDocument()
  expect(screen.getByLabelText(eigene.namen['ki.adresse'])).toBeInTheDocument()

  // **Der Hinweis sagt, was »aus« bedeutet** — kein einziger Aufruf.
  expect(screen.getByText(/kein einziger Aufruf/)).toBeInTheDocument()
})

// **Der Schalter schreibt seinen Wert.**
it('meldet einen umgelegten Schalter', async () => {
  const setze = zeige()

  // **Ein Schalter und keine Auswahl** — UXR-F2
  // (#2008).
  await userEvent.click(
    screen.getByRole('switch', { name: eigene.namen['ki.laufzeit'] }),
  )

  expect(setze).toHaveBeenCalledWith('ki.laufzeit', 'an')
})

// **Und die Adresse ebenso.**
it('meldet eine getippte Adresse', async () => {
  const setze = zeige()

  // **Beim Verlassen des Feldes** — seit OPS-02 (#1719) hält ein
  // Freitextfeld einen Entwurf; eine halbe Adresse wird nicht geschrieben.
  await userEvent.type(
    screen.getByLabelText(eigene.namen['ki.adresse']),
    'http://127.0.0.1:11434',
  )
  expect(setze).not.toHaveBeenCalled()

  await userEvent.tab()

  expect(setze).toHaveBeenCalledWith('ki.adresse', 'http://127.0.0.1:11434')
})

// **Solange etwas läuft, ist nichts zu bedienen.**
//
// **§7, Regel 4**: ein zweiter Ruf gegen einen Stand, den das Backend noch
// nicht kennt, gewinnt — und was der erste wollte, ist weg.
it('hält beide Felder an, solange etwas läuft', () => {
  zeige(dieZeilen, true)

  expect(screen.getByLabelText(eigene.namen['ki.laufzeit'])).toBeDisabled()
  expect(screen.getByLabelText(eigene.namen['ki.adresse'])).toBeDisabled()
})

// **Fehlt eine Zeile, fehlt ihr Bedienelement — und nicht die Seite.**
//
// **Die Gegenprobe zur Suche**: ein Bauteil, das eine fehlende Zeile als
// leeren Wert läse, zeigte ein Feld, das nichts schreibt.
it('zeigt nur, was dasteht', () => {
  zeige(dieZeilen.slice(0, 1))

  expect(screen.getByLabelText(eigene.namen['ki.laufzeit'])).toBeInTheDocument()
  expect(
    screen.queryByLabelText(eigene.namen['ki.adresse']),
  ).not.toBeInTheDocument()
})
