// **Die fünf Formatangaben stehen bei den Setangaben** —
// #1325.
//
// # Der Befund
//
// Sie standen im Kasten „Nicht gedeutet" — in **jedem** der zehn kanonischen
// Sets dieselben fünf. Das ist keine unverstandene Zeile, sondern ein Feld,
// das MSE2 kennt; was jede bedeutet, steht in `internal/domain/setkopf.go`.
//
// # Warum bei den Setangaben und nicht auf der Kartenseite
//
// Sie handeln vom **Set**, und der Kasten stand auf der Kartenlasche. Wer
// Karten baut, sucht dort keine Formatfassung.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Setangaben } from './Setangaben'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { SetkopfDTO } from '../api'

const texte = kataloge.de.seiten.setangaben

afterEach(() => {
  raeumeApi()
})

const voll: SetkopfDTO = {
  mseFassung: '2.1.2',
  spielfassung: '2020-04-25',
  stilfassung: '2020-09-04',
  versionskontrolle: 'none',
  apprenticecode: '',
}

const leer: SetkopfDTO = {
  mseFassung: '',
  spielfassung: '',
  stilfassung: '',
  versionskontrolle: '',
  apprenticecode: '',
}

function zeige(kopf: SetkopfDTO) {
  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    Setangaben: () =>
      Promise.resolve({ felder: [], werte: [], kopf, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Setangaben setID="1" verlaufsstand={0} />
    </Sprachraum>,
  )
}

describe('die Formatangaben eines Sets', () => {
  it('nennt jede Angabe mit ihrer Aufschrift', async () => {
    zeige(voll)

    const kasten = await screen.findByRole('region', { name: texte.kopf })

    const paare: [string, string][] = [
      [texte.mseFassung, '2.1.2'],
      [texte.spielfassung, '2020-04-25'],
      [texte.stilfassung, '2020-09-04'],
      [texte.versionskontrolle, 'none'],
    ]

    for (const [aufschrift, wert] of paare) {
      expect(kasten).toHaveTextContent(aufschrift)
      expect(kasten).toHaveTextContent(wert)
    }
  })

  // **Eine leere Angabe wird weggelassen und nicht als leer gezeigt.** Sie
  // stand nicht in der Datei; eine Zeile mit einem Gedankenstrich behauptete,
  // es gäbe sie und sie sei leer.
  it('lässt weg, was nicht in der Datei stand', async () => {
    zeige(voll)

    await screen.findByRole('region', { name: texte.kopf })

    expect(screen.queryByText(texte.apprenticecode)).toBeNull()
  })

  // **Ohne eine einzige Angabe steht der Kasten gar nicht da.** Ein Set, das
  // in dieser Anwendung entsteht, hat keine — und ein leerer Kasten wäre eine
  // Überschrift ohne Inhalt.
  it('bleibt weg, wo keine Angabe dasteht', async () => {
    zeige(leer)

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3): der
    // Leerzustand steht erst da, wenn die Antwort angekommen ist — und erst
    // dann sagt die Verneinung darunter etwas.
    expect(await screen.findByText(texte.leerTitel)).toBeInTheDocument()

    expect(screen.queryByRole('region', { name: texte.kopf })).toBeNull()
  })
})

// **Ohne eigenen Rahmen** — AP-12
// (#2369),
// Befund B-28: am Fenster stand „Angaben der Datei“ als einziger gerahmter
// Kasten in der Schublade, zwischen ungerahmten Feldern. Ein Abschnitt mit
// Überschrift wie die Feldgruppen daneben.
describe('die Formatangaben in der Schublade', () => {
  it('stehen als Abschnitt mit Überschrift und ohne Rahmen', async () => {
    zeige(voll)

    const abschnitt = await screen.findByRole('region', { name: texte.kopf })
    expect(abschnitt).not.toHaveClass('border')
    expect(
      screen.getByRole('heading', { level: 3, name: texte.kopf }),
    ).toBeInTheDocument()
  })
})
