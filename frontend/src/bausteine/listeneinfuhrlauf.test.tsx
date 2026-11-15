// **Was eine überholte Antwort an der Listeneinfuhr anrichtet** — Nachlauf
// über Phase 4.5, Befunde 2 und 3
// (#1292).
//
// **Eine eigene Datei**, weil `listeneinfuhrbefunde.test.tsx` an der
// Zeilengrenze aus §2 stand. Der Schnitt folgt der Frage: dort steht, *was der
// Kasten über Fehlschläge sagt*, hier, *was er mit einer Antwort tut, die zu
// einem Stand gehört, den es nicht mehr gibt* (§7, Regel 2).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Listeneinfuhr } from './Listeneinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ohneDublette } from '../test/einfuhr'

import type { Listendeutungszeile } from '../antwortenlisteneinfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = {
  feld: 'Kartenliste',
  deuten: 'Zuordnen',
  einlesen: 'Aufnehmen',
}

// eindeutige gibt eine Zeile mit genau einem Druck.
const eindeutige: Listendeutungszeile = {
  roh: '4 Lightning Bolt (M11) 149',
  zeilennummer: 1,
  menge: 4,
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-1',
      name: 'Lightning Bolt',
      setcode: 'm11',
      setname: 'Magic 2011',
      nummer: '149',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 1,
  eindeutig: true,
  befund: '',
  weitere: false,
  fehler: null,
}

// aufgenommen ist der Rückruf, den der Kasten nach dem Schreiben ruft.
function aufgenommen(): void {
  // absichtlich leer
}

// tippeUndDeute füllt das Feld und drückt den Knopf.
async function tippeUndDeute(text: string) {
  await userEvent.type(screen.getByLabelText(texte.feld), text)
  await userEvent.click(screen.getByRole('button', { name: texte.deuten }))
}

// **Was beim Aufnehmen nicht durchkam, steht da** — Befund 4.
// **Eine überholte Deutung setzt sich nicht mehr ein** — Nachlauf über Phase
// 4.5, Befund 2.
//
// # Was der erste Anlauf löste und was er offenließ
//
// Behoben war: eine **abgeschlossene** Deutung wird verworfen, sobald der Text
// sich ändert. Offen blieb der Fall davor — die Deutung von A ist noch
// **unterwegs**, der Text wird auf B geändert, und dann trifft die Antwort
// ein. Der Antwortweg schrieb ungeprüft in den Zustand: A stand wieder da,
// während sichtbar B im Feld steht.
//
// `useVorgang` verhindert **Parallelstarts**, keine überholten Antworten; das
// ist §7 Regel 2, und dafür gibt es `useLauf`.
//
// **Der alte Lauf konnte es nicht sehen**: er benutzt eine sofort aufgelöste
// Antwort, und die ist da, bevor jemand tippen kann. Hier wird sie
// festgehalten.
it('setzt eine überholte Deutung nicht mehr ein', async () => {
  let loese: (() => void) | undefined

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      new Promise((erfuellen) => {
        loese = () => {
          erfuellen({
            deutung: { zeilen: [eindeutige], befunde: [], alterSekunden: 0 },
            fehler: null,
          })
        }
      }),
    ListeEinlesen: vi.fn(),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')

  // **Der Text wechselt, während die Deutung unterwegs ist.**
  await userEvent.clear(screen.getByLabelText(texte.feld))
  await userEvent.type(screen.getByLabelText(texte.feld), '2 Counterspell')

  loese?.()

  // **Und die Antwort von vorhin trägt sich nicht mehr ein.**
  //
  // Gewartet wird auf ein Abschlusssignal und nicht auf eine Verneinung, die
  // schon vorher zutraf: der Knopf wird erst wieder bedienbar, wenn der
  // Vorgang durch ist.
  await waitFor(() => {
    expect(screen.getByRole('button', { name: texte.deuten })).toBeEnabled()
  })

  expect(
    screen.queryByText('Lightning Bolt · Magic 2011 #149'),
  ).not.toBeInTheDocument()
})

// **Ein Teilerfolg wird gezeigt, auch wenn ein Fehler danebensteht** —
// Nachlauf über Phase 4.5, Befund 3.
//
// # Was der erste Anlauf löste und was er offenließ
//
// Behoben war die **Go-Seite**: `App.ListeEinlesen` schickt Bericht und Fehler
// zugleich, weil geschrieben geschrieben ist (§7, Regel 3). Im Fenster kam der
// Bericht trotzdem nie an — `auspacken` wirft beim Fehler, **bevor** jemand
// `bericht` liest, und der `catch` setzte nur die Meldung.
//
// Wer also vierzig Zeilen aufnimmt, von denen dreißig durchgehen und zehn
// nicht, sah nur „es ging schief" — und ein zweiter Versuch schriebe die
// dreißig noch einmal.
it('zeigt den Teilbericht neben dem Fehler', async () => {
  const aufgenommenSpion = vi.fn()

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [eindeutige], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeEinlesen: () =>
      Promise.resolve({
        bericht: {
          angelegt: [1],
          erhoeht: 0,
          ausgelassen: 0,
          stueck: 4,
          befunde: [],
        },
        fehler: {
          code: 'unbekannt' as const,
          schluessel: '',
          meldung: 'Zeile 2 ließ sich nicht schreiben',
        },
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommenSpion} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  // **Beides steht da.**
  expect(
    await screen.findByText(/Zeile 2 ließ sich nicht schreiben/),
  ).toBeInTheDocument()

  expect(screen.getByText(/1 neu, 0 erhöht, 4 Stück/)).toBeInTheDocument()

  // **Und die Sammlung daneben wird nachgeladen**: es ist etwas geschrieben
  // worden, und ohne das Nachladen zeigte sie einen überholten Stand.
  await waitFor(() => {
    expect(aufgenommenSpion).toHaveBeenCalled()
  })
})

// **Und wenn gar nichts geschrieben wurde, wird auch nicht nachgeladen.**
//
// Die Gegenprobe zur Zeile darüber — und §7 Regel 3 in ihrer zweiten Hälfte:
// nach einem gescheiterten Schreiben überschriebe ein Lesefehler die Meldung,
// die sagt, was zu tun ist.
it('lädt nach einem ganz gescheiterten Einlesen nicht nach', async () => {
  const aufgenommenSpion = vi.fn()

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [eindeutige], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeEinlesen: () =>
      Promise.resolve({
        bericht: null,
        fehler: {
          code: 'unbekannt' as const,
          schluessel: '',
          meldung: 'kein Bestand',
        },
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommenSpion} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  // **Gewartet wird auf ein Abschlusssignal** und nicht auf die Verneinung
  // selbst: die Meldung erscheint erst, wenn der Vorgang durch ist.
  await screen.findByText(/kein Bestand/)

  expect(aufgenommenSpion).not.toHaveBeenCalled()
})

// **Auch ein überholter Fehler trägt sich nicht ein.**
//
// Die unauffälligere Hälfte von §7 Regel 2: eine Meldung zu einer Liste, die
// niemand mehr im Feld hat, sieht aus wie ein Fehler der jetzigen — und der
// Benutzer sucht ihn in dem Text, den er gerade getippt hat.
it('zeigt einen überholten Deutungsfehler nicht mehr an', async () => {
  let loese: (() => void) | undefined

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      new Promise((erfuellen) => {
        loese = () => {
          erfuellen({
            deutung: null,
            fehler: {
              code: 'validierung' as const,
              schluessel: '',
              meldung: 'diese Liste ist unlesbar',
            },
          })
        }
      }),
    ListeEinlesen: vi.fn(),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  await tippeUndDeute('völliger Unsinn')

  await userEvent.clear(screen.getByLabelText(texte.feld))
  await userEvent.type(screen.getByLabelText(texte.feld), '2 Counterspell')

  loese?.()

  await waitFor(() => {
    expect(screen.getByRole('button', { name: texte.deuten })).toBeEnabled()
  })

  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})

// **Ein überholtes Einlesen zeigt weder Bericht noch Fehler** — schreibt aber
// weiter und lädt nach.
//
// Beide Hälften in einem Lauf, weil sie zusammengehören: §7 Regel 3 sagt,
// geschrieben ist geschrieben (also läuft `aufAufgenommen`), und Regel 2 sagt,
// eine Antwort gehört zu ihrem Stand (also steht der Bericht nicht da).
it('trägt ein überholtes Einlesen nicht ein, lädt aber nach', async () => {
  const aufgenommenSpion = vi.fn()

  let loese: (() => void) | undefined

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [eindeutige], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeEinlesen: () =>
      new Promise((erfuellen) => {
        loese = () => {
          erfuellen({
            bericht: {
              angelegt: [1],
              erhoeht: 0,
              ausgelassen: 0,
              stueck: 4,
              befunde: [],
            },
            fehler: null,
          })
        }
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommenSpion} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await userEvent.type(screen.getByLabelText(texte.feld), ' und mehr')

  loese?.()

  // **Nachgeladen wird trotzdem**: geschrieben ist geschrieben.
  await waitFor(() => {
    expect(aufgenommenSpion).toHaveBeenCalled()
  })

  expect(screen.queryByText(/1 neu, 0 erhöht/)).not.toBeInTheDocument()
})

// **Und ein überholter Einlesefehler ebenso.**
//
// # Warum die Naht hier **wirft** statt einen Fehler zurückzugeben
//
// Ein Fehler *im Umschlag* kommt gar nicht so weit: die Prüfung auf den Stand
// steht vor `auspacken`, und der überholte Lauf kehrt dort um. Die einzige
// Gestalt, die den `catch` erreicht, ist eine **abgewiesene** Zusage — eine
// Wails-Bindung, die nicht antwortet, weil die Brücke weg ist.
it('zeigt einen überholten Einlesefehler nicht mehr an', async () => {
  let loese: (() => void) | undefined

  belegeApi({
    ListeVorschauen: ohneDublette,
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [eindeutige], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeEinlesen: () =>
      new Promise((_, abweisen) => {
        loese = () => {
          abweisen(new Error('die Brücke ist weg'))
        }
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={aufgenommen} />
    </Sprachraum>,
  )

  await tippeUndDeute('4 Lightning Bolt (M11) 149')
  await screen.findByText('Lightning Bolt · Magic 2011 #149')
  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await userEvent.type(screen.getByLabelText(texte.feld), ' und mehr')

  loese?.()

  await waitFor(() => {
    expect(screen.getByRole('button', { name: texte.deuten })).toBeEnabled()
  })

  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})
