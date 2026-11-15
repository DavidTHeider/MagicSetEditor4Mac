// **Ein Setwechsel räumt die setgebundenen Seiten** —
// #749.
//
// # Der Befund
//
// Über die 24 Schreibwege des Frontends setzen **16** einen Fehlerzustand im
// `catch`, ohne den Stand zu prüfen. Ob das je Weg ein Fehler ist, hängt daran,
// ob die Komponente den Wechsel ihres Bezugsobjekts überlebt — und die
// Meldungen dieser Seiten tragen **keinen Bezug**, an dem man es sähe.
//
// Durchgesehen, mit Ergebnis je Weg:
//
//   - **Räumen selbst**: `Kartenkopieren`, `Kartenloeschen` und `Kartenordnen`
//     setzen ihren Zustand zurück, sobald `karteID` wechselt.
//   - **Sind geschlüsselt**: `Karteneditor` über `inBearbeitung.id`,
//     `SetVorlagenausfuhr` und `MTGJSONSichern` über `satz.id`, und seit hier
//     `Setangaben`, `Stil`, `Statistik`, `Bilder` und `Druckbogen` über `setID`.
//   - **Prüfen den Stand**: `Karten` über `useAktuellesSet` (#716), `App` über
//     die Kennung an der Speichermeldung.
//   - **Hängen an keinem Set**: `Sets`, `Pakete`, `Paketliste`, `Uebernahme`.
//
// # Was dieser Wächter prüft, und was nicht
//
// Er fährt **einen** dieser Wege ganz — `Setangaben`, weil dort ein
// gescheitertes Schreiben mit einer Meldung ohne Bezug endet.
//
// **Die übrigen vier geschlüsselten Seiten prüft er nicht.** Sie hängen an
// derselben Zeile in `Seiteninhalt.tsx`; wer sie einzeln nachstellte, bräuchte
// je eine eigene Vorrichtung — Bilder, einen Bogen, eine Statistik — und
// prüfte am Ende dieselbe eine Zeile fünfmal. Das steht hier, damit niemand
// den Wächter für mehr hält, als er ist (§10).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { leererSetkopf } from '../test/setkopfprobe'
import { Setangaben } from './Setangaben'

afterEach(() => {
  raeumeApi()
})

function angaben() {
  return Promise.resolve({
    felder: [
      {
        name: 'title',
        aufschrift: 'Title',
        art: 'text',
        mehrzeilig: false,
        wahlen: [],
        beschreibung: '',
        vorgabe: '',
        teilfelder: [],
        reihe: '',
        stelle: 0,
        editierbar: true,
        gespeichert: true,
      },
    ],
    werte: [{ feld: 'title', wert: 'Erstes' }],
    kopf: leererSetkopf,
    fehler: null,
  })
}

// zeige baut die Seite so, wie `Seiteninhalt` es tut — samt Schlüssel.
function zeige(setID: string) {
  return (
    <Sprachraum sprache="de">
      <Setangaben key={`${setID}/0`} setID={setID} verlaufsstand={0} />
    </Sprachraum>
  )
}

describe('Ein Setwechsel', () => {
  it('räumt die Fehlermeldung der Setangaben', async () => {
    belegeApi({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
      Setangaben: angaben,
      SetwertSetzen: () =>
        Promise.reject(new Error('Der Bestand ist gesperrt.')),
    })

    const { rerender } = render(zeige('4'))

    const feld = await screen.findByLabelText('title')

    await userEvent.type(feld, 'X')
    await userEvent.tab()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand ist gesperrt.',
    )

    rerender(zeige('9'))

    await waitFor(() => {
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })

    // **Und die Seite steht wirklich da** — ohne diese Zeile wäre „kein
    // Fehler" auch dann wahr, wenn gar nichts mehr gezeichnet würde (§3).
    expect(await screen.findByLabelText('title')).toBeInTheDocument()
  })
})
