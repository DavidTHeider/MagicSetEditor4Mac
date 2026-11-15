// **Die Handgriffe am Dateibaum** — FMT-02.AK-01
// (#1691):
// eine Vorlage anlegen, ein Bild oder eine Schrift einfügen, die gewählte
// Datei umbenennen oder entfernen.
//
// # Die vier Regeln aus §7
//
//  1. **Ein Wechsel räumt den offenen Zustand** — Modus und Fehler tragen den
//     Schlüssel der Datei, bei der sie entstanden; eine andere sieht sie nicht.
//  2. **Eine späte Antwort wählt nichts um** — wer inzwischen eine andere
//     Datei aufgeschlagen hat, bleibt dort.
//  3. **Geschrieben ist geschrieben** — nach einem geglückten Handgriff lädt
//     der Baum nach; scheitert das, steht es am Baum und nimmt nichts zurück.
//  4. **Ein Vorgang je Paket** — ein zweiter Handgriff, während der erste
//     läuft, ginge gegen einen Stand, den der Dienst noch nicht kennt.
import { useEffect, useRef, useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

import type { Vorlagenwerkstatt } from './useVorlagenwerkstatt'

export type Dateimodus = 'ruhe' | 'anlegen' | 'umbenennen' | 'entfernen'

export interface Dateiaktionen {
  modus: Dateimodus
  setzeModus: (modus: Dateimodus) => void
  fehler: string | undefined
  laeuft: boolean
  anlegen: (name: string) => Promise<void>
  einfuegen: () => Promise<void>
  umbenennen: (neu: string) => Promise<void>
  entfernen: () => Promise<void>
}

type Werkstattteil = Pick<
  Vorlagenwerkstatt,
  'wahl' | 'waehlen' | 'hauptdatei' | 'baumNeuLaden' | 'puffer'
>

// ordnerVon gibt den Ordner einer Datei im Paket — leer, wenn sie oben liegt.
function ordnerVon(datei: string): string {
  return datei.slice(0, Math.max(0, datei.lastIndexOf('/')))
}

export function useDateiaktionen(werkstatt: Werkstattteil): Dateiaktionen {
  const texte = useTexte()
  const vorgang = useVorgang()
  const { wahl, puffer } = werkstatt
  const paket = wahl?.paket ?? ''
  const datei = wahl?.datei ?? ''
  const beiwerk = wahl?.beiwerk === true
  const schluessel = `${paket}\n${datei}`

  const aktuell = useRef(schluessel)
  useEffect(() => {
    aktuell.current = schluessel
  }, [schluessel])

  const [zustand, setzeZustand] = useState<{
    schluessel: string
    modus: Dateimodus
    fehler: string | undefined
  }>({ schluessel: '', modus: 'ruhe', fehler: undefined })

  const hier = zustand.schluessel === schluessel
  const modus = hier ? zustand.modus : 'ruhe'
  const fehler = hier ? zustand.fehler : undefined

  // ausfuehren startet einen Handgriff und wählt danach, was er ergab.
  const ausfuehren = async (
    arbeit: () => Promise<{ datei: string; beiwerk: boolean } | null>,
  ) => {
    const eigener = schluessel

    await vorgang.starte(paket, async () => {
      try {
        const ziel = await arbeit()

        if (ziel === null) {
          return
        }

        werkstatt.baumNeuLaden()

        if (aktuell.current === eigener) {
          setzeZustand({
            schluessel: eigener,
            modus: 'ruhe',
            fehler: undefined,
          })
          werkstatt.waehlen(paket, ziel.datei, ziel.beiwerk)
        }
      } catch (ursache: unknown) {
        if (aktuell.current === eigener) {
          setzeZustand((war) => ({
            ...war,
            schluessel: eigener,
            fehler: meldungAus(ursache, texte),
          }))
        }
      }
    })
  }

  // **Der Stand einer Vorlage geht mit**; Beiwerk hat keinen (service).
  const stand = beiwerk ? '' : puffer.stand

  return {
    modus,
    setzeModus: (neu) => {
      setzeZustand({ schluessel, modus: neu, fehler: undefined })
    },
    fehler,
    laeuft: vorgang.laeuft(paket),
    anlegen: (name) =>
      ausfuehren(async () => {
        const antwort = await api().PaketdateiAnlegen(paket, name, '')
        auspacken(antwort, () => antwort.stand)

        return { datei: name, beiwerk: false }
      }),
    einfuegen: () =>
      ausfuehren(async () => {
        const antwort = await api().PaketdateiEinfuegen(paket, ordnerVon(datei))
        const gelesen = auspacken(antwort, () => antwort)

        return gelesen.abgebrochen
          ? null
          : { datei: gelesen.datei, beiwerk: true }
      }),
    umbenennen: (neu) =>
      ausfuehren(async () => {
        const antwort = await api().PaketdateiUmbenennen(
          paket,
          datei,
          neu,
          stand,
        )
        auspacken(antwort, () => null)

        return { datei: neu, beiwerk }
      }),
    entfernen: () =>
      ausfuehren(async () => {
        const antwort = await api().PaketdateiEntfernen(paket, datei, stand)
        auspacken(antwort, () => null)

        return { datei: werkstatt.hauptdatei, beiwerk: false }
      }),
  }
}
