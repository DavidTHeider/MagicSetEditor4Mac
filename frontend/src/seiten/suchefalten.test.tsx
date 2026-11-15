// **Die Suche mit der Menge, die sie wirklich liefert** — UXR-05
// (#1913).
//
// # Was hier gemessen wird
//
// **Dass zweihundert Treffer nicht zweihundert Knöpfe auf einmal sind.** Die
// Volltextsuche gibt bis zu `domain.SearchLimit` = 200 Treffer zurück, und
// jeder ist ein Knopf, der die Karte aufschlägt. Mit den dreizehn
// Bedienelementen der Seite standen damit 213 da — R6 aus
// `docs/UIREVIEW.md` lässt 200 zu, und jeder ist ein Tabstopp mehr. Gefaltet
// wird wie in der `Listenansicht`: sechzig, dann „Weitere … zeigen".
//
// **Dass eine neue Frage wieder vorn beginnt** (§7, Regel 1) — die
// aufgeklappten Seiten gehören zur vorigen Antwort.
//
// **Und dass das Zusammenstellen für ein Modell nicht offen dasteht.** Es ist
// der seltenste Weg der Seite und trug fünf der elf Felder (R3: keine
// Formularwand); jetzt steht er hinter seinem Titel.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { seitengroesse } from '../bausteine/Listenansicht'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Suche } from './Suche'

const texte = kataloge.de.seiten.suche
const liste = kataloge.de.bausteine.listenansicht
const recherche = kataloge.de.seiten.recherche

afterEach(() => {
  raeumeApi()
})

// zweihundert sind so viele Treffer, wie der Dienst höchstens schickt.
function zweihundert(vorsatz: string) {
  return Array.from({ length: 200 }, (_, stelle) => ({
    karteID: `${vorsatz}${String(stelle)}`,
    setID: '7',
    nummer: stelle,
    kartenname: `${vorsatz} ${String(stelle + 1)}`,
    feld: 'text',
    wert: 'Kreatur',
  }))
}

function zeige() {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [] }),
    Suchfelder: () => Promise.resolve({ felder: [] }),
    Suchen: (text: string) =>
      Promise.resolve({ treffer: zweihundert(text), gekuerzt: true }),
  })

  render(
    <Sprachraum sprache="de">
      <Suche oeffnen={vi.fn()} />
    </Sprachraum>,
  )
}

// treffer zählt die Knöpfe, die eine Karte aufschlagen — erkannt am Namen.
function trefferknoepfe(vorsatz: string) {
  return screen
    .queryAllByRole('button')
    .filter((knopf) => knopf.textContent.startsWith(`${vorsatz} `))
}

describe('Die Treffer der Suche', () => {
  it('zeigt von 200 Treffern erst sechzig und den Rest auf Nachfrage', async () => {
    zeige()

    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.eingabe }),
      'Glut',
    )

    await waitFor(() => {
      expect(trefferknoepfe('Glut')).toHaveLength(seitengroesse)
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(liste.weitere, { anzahl: String(seitengroesse) }),
      }),
    )

    expect(trefferknoepfe('Glut')).toHaveLength(2 * seitengroesse)
  })

  it('beginnt mit einer neuen Frage wieder bei den ersten sechzig', async () => {
    zeige()
    const feld = screen.getByRole('searchbox', { name: texte.eingabe })

    await userEvent.type(feld, 'Glut')
    await waitFor(() => {
      expect(trefferknoepfe('Glut')).toHaveLength(seitengroesse)
    })
    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(liste.weitere, { anzahl: String(seitengroesse) }),
      }),
    )

    await userEvent.clear(feld)
    await userEvent.type(feld, 'Eis')

    await waitFor(() => {
      expect(trefferknoepfe('Eis')).toHaveLength(seitengroesse)
    })
  })

  it('legt das Zusammenstellen für ein Modell hinter seinen Titel', async () => {
    zeige()

    // **Die Recherche selbst steht hinter „Recherche …“** (AP-11, #2368).
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.seiten.suche.recherche,
      }),
    )

    expect(
      screen.queryByRole('textbox', { name: recherche.freigabeAufgabe }),
    ).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: recherche.freigabeTitel }),
    )

    expect(
      screen.getByRole('textbox', { name: recherche.freigabeAufgabe }),
    ).toBeInTheDocument()

    // **Der Titel steht einmal da** — als Knopf, nicht noch einmal als
    // Überschrift darunter.
    expect(screen.getAllByText(recherche.freigabeTitel)).toHaveLength(1)
  })
})
