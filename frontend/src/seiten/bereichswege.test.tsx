// **Jede Seite eines Bereichs ist aus seinem Kopf erreichbar** — UX-02
// (#1674).
//
// # Wofür es diesen Lauf gibt
//
// **Weil die Seitenleiste seit UX-02 Bereiche führt und keine Seiten.** Sie
// hatte vierzehn Einträge, sie hat neun — und *Turniere*, *Spielen* und
// *Vorlagen* standen darunter. Ohne einen Weg zu ihnen wäre der Umbau genau
// das, was **UX-01.AK-04** verbietet: *„Kein bisheriger Anwendungsfall
// verliert beim Umbau seinen Zugang."*
//
// **Der Entwurf zeichnet den Weg**: die Artboards *Decks* und *Pakete* tragen
// ihn im Kopf — `<a class="knopf" href="Turniere.dc.html">` und
// `<a class="knopf" href="Werkstatt.dc.html">`
// (`prototyp/teile/` (Entwicklungsarchiv)).
//
// # Und was er nicht hält
//
// **Den Weg zu den Karten.** Er führt nicht über den Kopf, sondern über das
// **Öffnen eines Sets** aus der Setliste — mit einem Set, das es geben muss.
// Geprüft ist er in `App.test.tsx` („führt zu den Karten des gewählten
// Sets"); die Aufzählung unten nimmt ihn deshalb ausdrücklich aus, statt ihn
// stillschweigend zu übergehen.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { bereiche } from '../bereiche'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { belegeProbeset, zeigeSeite } from '../test/seitenvorrichtung'

import type { Seitenschluessel } from '../navigation'

afterEach(() => {
  raeumeApi()
})

// wege sind die Anschlüsse, die der Kopf einer Seite anbietet.
//
// **Die Aufschrift ist nicht immer der Seitentitel.** Der Eintrag der Leiste
// trägt für *Spielen* denselben Namen wie die Seite; zwei Knöpfe „Spielen"
// nebeneinander sind für ein Vorleseprogramm nicht zu unterscheiden. Der Kopf
// sagt deshalb, **was geschieht** — „Partie anfangen", wie im Entwurf.
const wege = [
  { von: 'decks', zu: 'turniere', wort: kataloge.de.seiten.turniere.titel },
  { von: 'decks', zu: 'spielen', wort: kataloge.de.seiten.spielen.anfangen },
  { von: 'pakete', zu: 'vorlagen', wort: kataloge.de.seiten.vorlagen.titel },
] as const

// ueberDieSetliste sind die Seiten, die man anders erreicht.
const ueberDieSetliste: readonly Seitenschluessel[] = ['karten']

describe('Der Kopf einer Bereichsseite', () => {
  it.each(wege)('führt von $von nach $zu', async ({ von, zu, wort }) => {
    belegeProbeset()

    const wechseln = vi.fn()

    zeigeSeite(von, 'de', wechseln)

    await userEvent.click(await screen.findByRole('button', { name: wort }))

    expect(wechseln).toHaveBeenCalledWith(zu)
  })
})

describe('Die Aufzählung der Wege', () => {
  // **Sie hält sich selbst vollständig.** Ein neuer Bereichseintrag ohne Weg
  // fiele sonst nicht auf — er stünde einfach nicht in der Leiste, und
  // niemand sähe es. Das ist dieselbe Regel wie bei der Seitenliste (§3): eine
  // zu kurze Aufzählung ist grün.
  it('nennt jede Seite, die nicht die erste ihres Bereichs ist', () => {
    const ohneEintrag = bereiche
      .filter((bereich) => !bereich.dienst)
      .flatMap((bereich) => bereich.seiten.slice(1))
      .filter((seite) => !ueberDieSetliste.includes(seite))

    const erreichbar = wege.map((weg) => weg.zu)

    expect([...ohneEintrag].sort()).toEqual([...erreichbar].sort())
  })
})
