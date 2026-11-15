// **Keine Auszeichnungszeichen im Fenster**
// (#1100).
//
// # Der Befund
//
// Unter „Nach neuen Fassungen sehen" stand wörtlich `**aus**` — mit
// Sternchen. Auf der Lasche *Ausgeben* stand `` `.mse-set` `` — mit
// Schrägstrichen. Die Kataloge benutzen beides seit der Hilfeseite, und
// `bausteine/Satz.tsx` löst es auf; die beiden Stellen reichten den Text roh
// durch.
//
// **Gefunden bei der Sichtprüfung am laufenden Fenster** für die Freigabe
// 0.2.0, nicht von einem Test. Genau die Sorte Befund, für die es sie gibt
// (§3) — und genau die Sorte, die sich danach maschinell halten lässt.
//
// # Warum das ein eigener Wächter ist und keine zwei Korrekturen
//
// Weil die **Klasse** das Problem ist, nicht der Fall: jeder Katalogtext darf
// auszeichnen, und jede Stelle, die ihn zeigt, muss ihn auflösen. Zwei
// Korrekturen ließen die dritte Stelle offen — und niemand sähe sie, bis
// wieder jemand hinschaut.
//
// # Was er nicht prüft
//
//   - **Ob die Auszeichnung sinnvoll gesetzt ist.** Er sieht, dass kein
//     Zeichen stehen bleibt, nicht dass die richtigen Wörter betont sind.
//   - **Text, der aus einem MSE-Paket kommt.** Ein Kartentext darf einen Stern
//     tragen; geprüft werden die Seiten mit ihrer Vorrichtung, und die führt
//     das Probeset.
//   - **Seiten mit offenem Dialog.** Er zeichnet jede Seite ohne offenes
//     Fenster darüber.
//
// # Und seit dem 2026-09-06 schlägt er die Laschen auf
//
// **Weil er sonst die halbe Seite prüfte.** Hier stand „er zeichnet jede Seite
// in ihrem Ruhezustand" — und *Nutzer und Profile* hat drei Laschen, von denen
// er nur die erste sah. In den beiden anderen standen **zwölf** rohe
// Auszeichnungen, aufgebaut über die Werkstücke 8 bis 15 von Phase 4.12;
// gefunden hat sie der Auftraggeber am laufenden Fenster, nicht dieser Lauf.
//
// **Das ist die Klasse, für die es ihn gibt** (§10): jede Lasche ist eine
// Seite, die jemand sieht. Er klickt sie deshalb der Reihe nach an und prüft
// nach jedem Klick.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { seiten } from './navigation'
import { sprachen } from './sprache/kataloge'
import { belegeProbeset, zeigeSeite } from './test/seitenvorrichtung'
import { raeumeApi } from './test/api'
import { zurRuhe } from './test/geladeneseite'

afterEach(() => {
  raeumeApi()
})

// zeichen sind die Auszeichnungen, die `Satz` auflöst.
//
// **Der Schrägstrich einzeln und das Sternchen doppelt.** Ein einzelnes
// Sternchen steht in keinem Katalogtext, käme aber aus einem Kartentext; die
// Auszeichnung, die `Satz` kennt, ist das Paar.
const zeichen = /\*\*|`/

// rohe gibt die Textknoten, in denen eine Auszeichnung stehen geblieben ist.
function rohe(bereich: HTMLElement): string[] {
  return [...bereich.querySelectorAll('*')]
    .flatMap((knoten) => [...knoten.childNodes])
    .filter((kind) => kind.nodeType === Node.TEXT_NODE)
    .map((kind) => kind.textContent ?? '')
    .filter((text) => zeichen.test(text))
}

describe('Kein Katalogtext zeigt seine Auszeichnung', () => {
  for (const sprache of sprachen) {
    // **Über `seiten.map(…)` und nicht über `Object.keys(seiten)`.**
    // `seiten` ist ein **Array** von Einträgen; `Object.keys` gäbe „0", „1",
    // „2" — und `zeigeSeite('0', …)` zeichnet nichts. Der erste Anlauf tat
    // genau das: 34 grüne Fälle, die keine Seite ansahen. Ein Wächter, der
    // nichts prüft, ist schlimmer als keiner (§10).
    for (const { schluessel } of seiten) {
      it(`${schluessel} (${sprache}) löst sie auf`, async () => {
        belegeProbeset()

        zeigeSeite(schluessel, sprache)

        // **Auf den Hauptbereich gewartet**, damit die Seite wirklich steht:
        // eine Zusicherung über einen leeren Baum ist immer wahr (§3).
        await screen.findByRole('main')

        // **Geprüft wird das ganze Dokument, nicht nur `main`.** Seit UX-08
        // (#1680) stehen sieben Setseiten in einer Schublade, und die zeichnet
        // Mantine in ein Portal außerhalb des Hauptbereichs — ein Wächter über
        // `main` sähe von *Stil* nur die Karten darunter und hieße grün.
        const bereich = document.body

        // **Und auf die Antworten der Seite gewartet.**
        //
        // **`findByRole('main')` sagt nur, dass der Rahmen steht.** Die
        // Nutzerseite holt sechs Antworten und zeigt bis dahin „Wird geladen
        // …" — ein Wächter, der davor prüft, sieht drei Wörter und keine
        // Seite (§3: gewartet wird auf die **Wirkung**, nicht auf den Ruf).
        // Eine Runde reichte nicht (G02) — gewartet wird, bis sie steht.
        await zurRuhe(bereich, schluessel)

        expect(rohe(bereich)).toEqual([])

        // **Und jede Lasche einzeln** — sie sind Seiten, die jemand sieht.
        //
        // **Die erste ist schon offen**, deshalb beginnt die Schleife bei
        // eins; sie noch einmal anzuklicken prüfte dasselbe zweimal.
        const laschen = screen.queryAllByRole('tab')

        for (const lasche of laschen.slice(1)) {
          await userEvent.click(lasche)
          await zurRuhe(bereich, schluessel)

          expect(rohe(bereich)).toEqual([])
        }
      })
    }
  }
})
