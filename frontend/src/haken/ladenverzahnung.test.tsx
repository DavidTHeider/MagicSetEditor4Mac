// **Fragen und Antworten verzahnt, mit Abbrüchen und Fehlern dazwischen**
// (C05, #2230).
//
// `ladenreihenfolge.test.tsx` stellt erst alle Fragen und lässt dann alle
// Antworten kommen. Im Betrieb wechselt sich beides ab: ein Rückspielen, ein
// Zurücksetzen oder ein fremder Schreiber erhöht den Nachlesestand, während
// die vorige Antwort noch unterwegs ist — und die Vorschau bricht ihren
// Vorgänger ab (`abgebrochen`). Hier bestimmt ein Samen, wann gefragt und
// wann geantwortet wird und wie jede Frage endet; ein kleines Modell sagt
// nach jedem Schritt, was dastehen muss:
//
// - eine Antwort zählt nur, wenn ihre Frage die jüngste ist (§7, Regel 2),
// - ein Fehler lässt die bestätigten Daten stehen (U07, #2123),
// - ein Abbruch lässt alles stehen und holt **einmal** nach, wenn noch
//   nichts dasteht (#566).
//
// **Was er nicht prüft**: den Bezugswechsel — der räumt, und das hält der
// Lauf daneben — und den Profilwechsel, der die Anwendung über ihren
// Schlüssel neu baut (`App.tsx`), sodass keine alte Instanz mehr zeichnet.
import { act, renderHook } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Bindungsfehler } from '../api'
import { useLaden } from './useLaden'

type Ausgang = 'gut' | 'fehler' | 'abbruch'

interface Frage {
  wert: string
  erfuellen: (wert: string) => void
  ablehnen: (ursache: Bindungsfehler) => void
  offen: boolean
}

// zufall ist derselbe gesäte Zahlengeber wie im Lauf daneben.
function zufall(samen: number): () => number {
  let s = samen >>> 0

  return () => {
    s = (s + 0x6d2b79f5) >>> 0
    let t = s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)

    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const fehlerMit = (code: 'abgebrochen' | 'unbekannt', meldung: string) =>
  new Bindungsfehler({
    code,
    schluessel: 'unbekannt',
    meldung,
    feld: '',
    werte: {},
  })

it.each(Array.from({ length: 40 }, (_, samen) => samen))(
  'Samen %i: Nachlesen, Abbruch und Fehler zeigen nie einen überholten Stand',
  async (samen) => {
    const naechste = zufall(2230 + samen)
    const fragen: Frage[] = []

    const { result, rerender } = renderHook(
      ({ stand }: { stand: number }) =>
        useLaden(
          () =>
            new Promise<string>((erfuellen, ablehnen) => {
              fragen.push({
                wert: `eins@${String(fragen.length)}`,
                erfuellen,
                ablehnen,
                offen: true,
              })
            }),
          ['eins'],
          false,
          stand,
        ),
      { initialProps: { stand: 0 } },
    )

    // **Das Modell**: was dastehen muss, und ob schon einmal nachgeholt wurde.
    let daten: string | undefined
    let fehler = false
    let nachgeholt = false
    let stand = 0

    for (
      let schritt = 0;
      schritt < 12 || fragen.some((f) => f.offen);
      schritt++
    ) {
      const offene = fragen.filter((f) => f.offen)

      if (schritt < 12 && (offene.length === 0 || naechste() < 0.4)) {
        stand++
        rerender({ stand })

        continue
      }

      const frage = offene[Math.floor(naechste() * offene.length)]
      if (frage === undefined) break

      const zahl = naechste()
      const ausgang: Ausgang =
        zahl < 0.2 ? 'abbruch' : zahl < 0.4 ? 'fehler' : 'gut'
      const juengste = frage === fragen[fragen.length - 1]
      frage.offen = false

      await act(async () => {
        if (ausgang === 'gut') frage.erfuellen(frage.wert)
        else if (ausgang === 'fehler')
          frage.ablehnen(fehlerMit('unbekannt', 'der Bestand klemmt'))
        else frage.ablehnen(fehlerMit('abgebrochen', 'abgebrochen'))
        await Promise.resolve()
        await Promise.resolve()
      })

      if (juengste && ausgang === 'gut') {
        daten = frage.wert
        fehler = false
      } else if (juengste && ausgang === 'fehler') {
        fehler = true
      } else if (juengste && daten === undefined && !nachgeholt) {
        nachgeholt = true
      }

      const jetzt = result.current
      expect(jetzt.daten).toBe(daten)

      // **Ein Fehler steht nur da, wenn nicht mehr geladen wird** — und dann
      // ist es der der jüngsten beantworteten Frage oder der, den ein Abbruch
      // stehen ließ.
      if (!jetzt.laedt) expect(jetzt.fehler !== undefined).toBe(fehler)
    }

    // **Am Ende**: nichts mehr offen, nichts lädt, und es steht der Stand der
    // jüngsten Frage — oder, wenn sie scheiterte, der zuletzt bestätigte.
    expect(fragen.every((f) => !f.offen)).toBe(true)
    expect(result.current.daten).toBe(daten)
  },
)
