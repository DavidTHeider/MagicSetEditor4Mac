// **Die Autoersetzungsregeln eines Spiels** — Werkstück E von Phase 4.11
// (#1341,
// Vermerk [0060](../../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).
//
// # Was MSE2 hier tut
//
// Ein eigenes Fenster (`src/gui/auto_replace_window.cpp`, Zeilen 157–210,
// Fassung f9d9356 vom 2021-07-06). Hier ist es ein Kasten auf der
// Kartenseite: §6 verbietet verschachtelte Dialoge, und die Regeln greifen
// beim Tippen in genau den Feldern, die daneben stehen.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Autoersetzungen } from './Autoersetzungen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.autoersetzung

// zwei sind eine Regel des Pakets und eine eigene.
const zwei = [
  {
    muster: '(W)',
    ersatz: '<sym>W</sym>',
    ganzesWort: true,
    aktiv: true,
    eigen: false,
  },
  { muster: '!!', ersatz: '…', ganzesWort: false, aktiv: false, eigen: true },
]

function zeige(zusatz = {}) {
  belegeApi({
    Autoersetzungen: () =>
      Promise.resolve({ regeln: zwei, an: true, fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Autoersetzungen setID="1" />
    </Sprachraum>,
  )
}

// aufklappen wartet, bis der Inhalt des Kastens dasteht.
//
// **Seit dem 2026-09-03 ist er zu** — Werkstück C von Phase 4.10: ein
// Erklärabsatz und zwei Felder standen über der Kartenliste, und die Regeln
// stellt man einmal je Spiel ein. **Seit UXR-05 (#1913) öffnet ihn das Menü
// „Mehr" der Kartenseite** (`kartenwerkzeugzeile.test.tsx`); dieses Bauteil
// ist sein Inhalt und steht hier ohne Aufklapper da. Der Name bleibt, damit
// die Fälle darunter lesen, was sie schon lasen.
async function aufklappen() {
  await screen.findByRole('checkbox', { name: texte.schalter })
}

// **Beide Regeln stehen da — und nur die eigene lässt sich löschen.**
//
// MSEs `remove->Enable(ar && ar->custom)`: eine Regel des Pakets gehört dem
// Paket. Was geht, ist sie abzuschalten.
it('zeigt beide Regeln und nur bei der eigenen einen Löschknopf', async () => {
  zeige()
  await aufklappen()

  expect(await screen.findByText('(W)')).toBeInTheDocument()
  expect(screen.getByText('!!')).toBeInTheDocument()

  expect(
    screen.getByRole('button', { name: `${texte.loeschen} — !!` }),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: `${texte.loeschen} — (W)` }),
  ).not.toBeInTheDocument()

  // **Und bei der Paketregel steht, woher sie kommt.**
  expect(screen.getByText(texte.ausDemPaket)).toBeInTheDocument()
})

// **Der Ersatz einer Paketregel lässt sich ändern.**
it('schreibt den Ersatz einer Regel des Pakets fort', async () => {
  const AutoersetzungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSetzen })
  await aufklappen()

  const feld = await screen.findByRole('textbox', {
    name: `${texte.ersatz} — (W)`,
  })

  await userEvent.clear(feld)
  await userEvent.type(feld, 'WEISS')
  await userEvent.tab()

  await waitFor(() => {
    expect(AutoersetzungSetzen).toHaveBeenCalledWith(
      '1',
      '(W)',
      'WEISS',
      true,
      true,
    )
  })
})

// **Eine neue Regel wird angelegt — und der Knopf bleibt ohne Muster tot.**
//
// Eine Regel ohne Muster greift nie, und der Dienst wiese sie ab. Ein Knopf,
// der nur eine Fehlermeldung auslöst, ist keine Hilfe (§6).
it('legt eine Regel an und verlangt dafür ein Muster', async () => {
  const AutoersetzungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSetzen })
  await aufklappen()

  const anlegen = await screen.findByRole('button', { name: texte.anlegen })

  expect(anlegen).toBeDisabled()

  await userEvent.type(
    screen.getByRole('textbox', { name: texte.muster }),
    '(G)',
  )
  await userEvent.type(
    screen.getByRole('textbox', { name: texte.ersatz }),
    '<sym>G</sym>',
  )

  await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))

  await waitFor(() => {
    expect(AutoersetzungSetzen).toHaveBeenCalledWith(
      '1',
      '(G)',
      '<sym>G</sym>',
      true,
      true,
    )
  })
})

// **Der globale Schalter legt sich um.**
it('schaltet die Autoersetzung für das ganze Spiel', async () => {
  const AutoersetzungSchalten = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSchalten })
  await aufklappen()

  await userEvent.click(
    await screen.findByRole('checkbox', { name: texte.schalter }),
  )

  await waitFor(() => {
    expect(AutoersetzungSchalten).toHaveBeenCalledWith('1', false)
  })
})

// **Löschen und Zurücksetzen gehen durch.**
it('löscht eine eigene Regel und setzt zurück', async () => {
  const AutoersetzungLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
  const AutoersetzungenZuruecksetzen = vi.fn(() =>
    Promise.resolve({ fehler: null }),
  )

  zeige({ AutoersetzungLoeschen, AutoersetzungenZuruecksetzen })
  await aufklappen()

  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.loeschen} — !!` }),
  )

  await waitFor(() => {
    expect(AutoersetzungLoeschen).toHaveBeenCalledWith('1', '!!')
  })

  // **Auf den bedienbaren Knopf gewartet, nicht auf den Ruf** (#1809):
  // solange das Löschen läuft, ruht er.
  await waitFor(() => {
    expect(
      screen.getByRole('button', { name: texte.zuruecksetzen }),
    ).toBeEnabled()
  })

  await userEvent.click(
    screen.getByRole('button', { name: texte.zuruecksetzen }),
  )

  // **Alle eigenen Regeln erst nach der Frage** (UX-10.AK-04, #1682) —
  // und Abbrechen vergisst nichts.
  expect(AutoersetzungenZuruecksetzen).not.toHaveBeenCalled()
  await userEvent.click(
    within(
      screen.getByRole('group', { name: texte.zuruecksetzenFrage }),
    ).getByRole('button', { name: kataloge.de.bausteine.aktionen.abbrechen }),
  )
  expect(AutoersetzungenZuruecksetzen).not.toHaveBeenCalled()
  await userEvent.click(
    screen.getByRole('button', { name: texte.zuruecksetzen }),
  )
  await userEvent.click(
    within(
      screen.getByRole('group', { name: texte.zuruecksetzenFrage }),
    ).getByRole('button', { name: texte.zuruecksetzenJa }),
  )

  await waitFor(() => {
    expect(AutoersetzungenZuruecksetzen).toHaveBeenCalledWith('1')
  })
})

// **Ein Fehler steht am Kasten und hält ihn nicht an.**
it('zeigt einen Fehler des Schreibens', async () => {
  zeige({
    AutoersetzungSetzen: () =>
      Promise.resolve({
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'eine Regel ohne Muster greift nie',
        },
      }),
  })
  await aufklappen()

  await userEvent.type(
    await screen.findByRole('textbox', { name: texte.muster }),
    '(G)',
  )

  await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))

  expect(await screen.findByText(/ohne Muster greift nie/)).toBeInTheDocument()

  // **Und der Kasten steht weiter da** — ein Fehler beim Schreiben nimmt die
  // Liste nicht mit (§7, Regel 3).
  expect(screen.getByText('(W)')).toBeInTheDocument()
})

// **Die beiden Kästchen einer Zeile schreiben fort.**
//
// `aktiv` gibt es an **jeder** Regel — auch an einer des Pakets: sie
// abzuschalten ist das, was MSE statt des Löschens erlaubt
// (`AutoReplaceWindow::refreshItem`, `src/gui/auto_replace_window.cpp`,
// Zeilen 258–264, Fassung f9d9356 vom 2021-07-06).
it('schaltet eine Regel ab und ändert ihre Wortgrenze', async () => {
  const AutoersetzungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSetzen })
  await aufklappen()

  // **Das erste Kästchen gehört der ersten Regel** — `(W)` aus dem Paket.
  const [erstesKaestchen] = await screen.findAllByRole('checkbox', {
    name: texte.aktiv,
  })

  if (erstesKaestchen === undefined) {
    throw new Error('keine Zeile trägt ein Kästchen')
  }

  await userEvent.click(erstesKaestchen)

  await waitFor(() => {
    expect(AutoersetzungSetzen).toHaveBeenCalledWith(
      '1',
      '(W)',
      '<sym>W</sym>',
      true,
      false,
    )
  })

  // **Auf die bedienbare Wortgrenze gewartet** (#1809) — solange das Setzen
  // läuft, ruhen die Kästchen der Zeile.
  await waitFor(() => {
    expect(
      screen.getAllByRole('checkbox', { name: texte.ganzesWort })[0],
    ).toBeEnabled()
  })

  const [ersteWortgrenze] = screen.getAllByRole('checkbox', {
    name: texte.ganzesWort,
  })

  if (ersteWortgrenze === undefined) {
    throw new Error('keine Zeile trägt eine Wortgrenze')
  }

  await userEvent.click(ersteWortgrenze)

  await waitFor(() => {
    expect(AutoersetzungSetzen).toHaveBeenCalledWith(
      '1',
      '(W)',
      '<sym>W</sym>',
      false,
      true,
    )
  })
})

// **Und ein unveränderter Ersatz schreibt nicht.**
//
// Das Feld schreibt beim Verlassen; ohne diesen Zweig löste jedes Wegklicken
// einen Schreibvorgang aus — und damit einen Verlaufsschritt für nichts.
it('schreibt nicht, wenn der Ersatz unverändert bleibt', async () => {
  const AutoersetzungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSetzen })
  await aufklappen()

  const feld = await screen.findByRole('textbox', {
    name: `${texte.ersatz} — (W)`,
  })

  await userEvent.click(feld)
  await userEvent.tab()

  // **Der Beleg, dass jemand zuhört**, steht im Fall darüber: dieselbe
  // Bedienung mit geändertem Text ruft `AutoersetzungSetzen`. Ohne ihn wäre
  // diese Verneinung auch dann grün, wenn gar kein Horcher hinge (§3).
  expect(AutoersetzungSetzen).not.toHaveBeenCalled()
})

// **Das Muster sagt am Feld, dass es eines braucht** — Werkstück U3 von Phase
// 5.1 (#1414).
//
// Die Regel steht in `domain.PruefeAutoersetzung`: ein leeres Muster wird
// abgewiesen („eine Regel ohne Muster greift nie"). Bis heute sagte das nur
// der ausgegraute Knopf — und wer ihn sieht, weiß nicht, **welches** der
// beiden Felder fehlt (§6).
//
// **Gemeldet wird, wer das Feld anfasst und leer verlässt**: ein leeres
// Formular ist kein Fehler, sondern ein leeres Formular.
it('meldet das fehlende Muster am Feld, nicht nur am Knopf', async () => {
  zeige()
  await aufklappen()

  const feld = await screen.findByRole('textbox', { name: texte.muster })

  expect(feld).toHaveAttribute('aria-invalid', 'false')

  await userEvent.click(feld)
  await userEvent.tab()

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toBeInTheDocument()
})

// **Der Ersatz darf leer bleiben** — `PruefeAutoersetzung` prüft ihn nicht,
// und eine Regel, die einen Text **entfernt**, ist eine gültige Regel.
it('verlangt keinen Ersatz', async () => {
  const AutoersetzungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ AutoersetzungSetzen })
  await aufklappen()

  await userEvent.type(
    await screen.findByRole('textbox', { name: texte.muster }),
    '(G)',
  )

  await userEvent.click(screen.getByRole('button', { name: texte.anlegen }))

  await waitFor(() => {
    expect(AutoersetzungSetzen).toHaveBeenCalledWith('1', '(G)', '', true, true)
  })
})
