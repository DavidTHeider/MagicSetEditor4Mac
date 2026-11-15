// **Was die Liste der eigenen Übersetzungen tut** — LOC-02
// (#1688).
//
// Anlegen, Ein- und Ausgeben, Löschen. **Ein Vorgang je Paket** (§7, Regel
// 4), und nach einem geglückten Schreiben wird die Liste nachgeladen —
// Bestenfalls-Arbeit, die das Geschriebene nicht zurücknimmt (Regel 3).
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { Bindungsfehler } from '../bindungsfehler'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken, type Umschlag } from '../umschlag'
import { useLaden } from './useLaden'
import { useVorgang } from './useVorgang'

import type { SprachpaketAntwort, SprachpaketDTO } from '../apisprache'

// Sprachbefund ist ein Fehler — samt dem Feld, an dem er steht.
export interface Sprachbefund {
  feld: string
  text: string
}

// neuesPaket ist der Schlüssel, unter dem das Anlegen läuft — es hat noch
// keine Kennung, die es sperren könnte.
const neuesPaket = '+neu'

export function useSprachpakete(bearbeite: (kennung: string) => void) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.eigenuebersetzung
  const vorgang = useVorgang()
  const [befund, setzeBefund] = useState<Sprachbefund>()
  const [meldung, setzeMeldung] = useState<string>()

  const liste = useLaden(
    useCallback(async () => {
      const antwort = await api().Sprachpakete()

      return auspacken(antwort, () => antwort.pakete)
    }, []),
    [],
  )

  // fuehre führt einen Schreibweg aus — **ein weggeklickter Dialog ist
  // kein Fehler** und lädt nichts nach.
  const fuehre = async <T extends Umschlag>(
    kennung: string,
    arbeit: () => Promise<T>,
    danach: (antwort: T) => string | undefined,
  ) => {
    setzeBefund(undefined)
    setzeMeldung(undefined)

    await vorgang.starte(kennung, async () => {
      let antwort: T

      try {
        antwort = await arbeit()
        auspacken(antwort, () => undefined)
      } catch (ursache) {
        setzeBefund({
          feld: ursache instanceof Bindungsfehler ? (ursache.feld ?? '') : '',
          text: meldungAus(ursache, texte),
        })

        return
      }

      if ((antwort as Partial<SprachpaketAntwort>).abgebrochen === true) {
        return
      }

      setzeMeldung(danach(antwort))
      liste.nachladen()
    })
  }

  return {
    liste,
    befund,
    meldung,
    laeuft: vorgang.laeuft,
    anlegen: (kennung: string, name: string, basis: string) =>
      fuehre(
        neuesPaket,
        () => api().SprachpaketAnlegen(kennung, name, basis),
        (antwort) => {
          bearbeite(antwort.paket.kennung)

          return undefined
        },
      ),
    einlesen: (ersetzen: boolean) =>
      fuehre(
        neuesPaket,
        () => api().SprachpaketEinlesenOeffnen(ersetzen),
        (antwort) => einsetzen(eigene.eingelesen, { name: antwort.paket.name }),
      ),
    ausgeben: (paket: SprachpaketDTO) =>
      fuehre(
        paket.kennung,
        () => api().SprachpaketAusgabeSpeichern(paket.kennung),
        (antwort) => einsetzen(eigene.geschrieben, { pfad: antwort.pfad }),
      ),
    loeschen: (paket: SprachpaketDTO) =>
      fuehre(
        paket.kennung,
        () => api().SprachpaketLoeschen(paket.kennung),
        () => undefined,
      ),
  }
}
