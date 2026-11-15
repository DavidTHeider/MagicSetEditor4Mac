// **Den Abgleich fahren und seinen Stand halten** — Werkstück T4 von Phase
// 5.6 (#1469).
//
// # Warum der eingesetzte Text hier liegt und nicht in der Anzeige
//
// **Weil er den Wechsel des Gegenübers überleben darf und nicht soll.** Der
// Kasten hängt an **einem** Kontakt; wer den Kontakt wechselt, hat einen
// anderen Stand vor sich (§7, Regel 1). Der `key` an der Kontaktzeile räumt
// beides zugleich — den Text und das Ergebnis.
//
// # Und warum ein unlesbarer Text kein Ruf ist
//
// **Weil ein leerer Abgleich wie ein Ergebnis aussieht.** „Dazu passt nichts"
// und „das war kein Stand" sind zwei Auskünfte; wer den Text ungeprüft
// hinausschickte, bekäme für beides dieselbe.
import { useState } from 'react'

import { Tauschabgleich } from './Tauschabgleich'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { Fremdstand, Tauschabgleichantwort } from '../apitauschabgleich'

interface Kontakttauscheigenschaften {
  kennung: string
}

export function Kontakttausch({ kennung }: Kontakttauscheigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [stand, setzeStand] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [abgleich, setzeAbgleich] = useState<Tauschabgleichantwort>()

  const kennzeichen = `tauschabgleich/${kennung}`

  // **Der Stand kann aus einem Umschlag kommen** (T7, #1472) — der Weg
  // **ohne Leitung**. Über die Leitung fragt seit P11 (#2379) „Freigegebene
  // Sicht anfragen“ an den Zeilen des Tauschs (`Sichtvorschlag`); dieser
  // Kasten bleibt für Datei und Text. **Gelesen und nicht gebucht**: eine
  // Sammlungssicht sagt, was ein Gegenüber hat.
  const ausUmschlag = (): Promise<void> =>
    vorgang.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().UmschlagSicht()

        const zeilen = auspacken(antwort, () => antwort.zeilen)

        setzeStand(JSON.stringify({ angebot: zeilen, wuensche: [] }, null, 2))
        setzeAbgleich(undefined)
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const abgleichen = (): Promise<void> =>
    vorgang.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      const gelesen = fremdstandAus(stand)
      if (gelesen === undefined) {
        setzeAbgleich(undefined)
        setzeFehler(texte.seiten.handeln.tauschUnlesbar)

        return
      }

      try {
        const antwort = await api().Tauschabgleich(kennung, gelesen, 1, 0)

        setzeAbgleich(auspacken(antwort, () => antwort))
      } catch (ursache) {
        setzeAbgleich(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <>
      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      <Tauschabgleich
        stand={stand}
        abgleich={abgleich}
        laeuft={vorgang.laeuft(kennzeichen)}
        setzeStand={(neuer) => {
          setzeStand(neuer)
        }}
        abgleichen={() => {
          void abgleichen()
        }}
        ausUmschlag={() => {
          void ausUmschlag()
        }}
      />
    </>
  )
}

// fremdstandAus liest, was jemand eingesetzt hat — oder `undefined`.
//
// **Zwei Listen und beide wahlfrei.** Wer nur die Wünsche seines Gegenübers
// hat, bekommt eine Richtung; wer gar nichts Lesbares einsetzt, bekommt eine
// Meldung statt eines leeren Ergebnisses.
function fremdstandAus(text: string): Fremdstand | undefined {
  try {
    const gelesen: unknown = JSON.parse(text)

    if (typeof gelesen !== 'object' || gelesen === null) {
      return undefined
    }

    const roh = gelesen as Partial<Fremdstand>

    if (!Array.isArray(roh.angebot) && !Array.isArray(roh.wuensche)) {
      return undefined
    }

    return {
      angebot: Array.isArray(roh.angebot) ? roh.angebot : [],
      wuensche: Array.isArray(roh.wuensche) ? roh.wuensche : [],
    }
  } catch {
    return undefined
  }
}
