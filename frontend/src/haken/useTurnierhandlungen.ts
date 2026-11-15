// **Was man am ganzen Turnier tut, außer es abzuschließen** — AP-6
// (#2363),
// Befund B-04: Ausgeben, Bericht sichern und Löschen standen gleichrangig
// neben „Abschließen“ über den Laschen. Sie stehen jetzt hinter „…“ — und im
// Rechtsklick auf den Kopf dieselbe Liste (Entwurf `Kontext09`).
//
// # Ausgeben und Bericht sind ein Lesen, Löschen ein Schreiben
//
// **Ausgeben und Bericht** fragen nach einem Ort im Dialog von macOS und
// sagen danach, wo die Datei liegt (DOM-04.AK-03, #1699). Ein weggeklickter
// Dialog sagt nichts: er ist eine Entscheidung und kein Fehler. Sie ändern
// nichts und laden nichts nach. **Löschen fragt erst** (UX-08.AK-03), und das
// gelöschte Turnier bleibt nicht aufgeschlagen (§7, Regel 1) — eine
// unterwegs befindliche Leseantwort schlägt es nicht wieder auf (Regel 2).
import { useState } from 'react'

import { api } from '../api'
import type { TurnierDTO, TurnierdateiAntwort } from '../apiturniere'
import type { Menueeintrag } from '../bausteine/Menue'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import type { Turnierstand } from './useTurniere'
import { useHandlungsliste } from './useHandlungsliste'

export interface Turnierhandlungen {
  eintraege: Menueeintrag[]
  // meldung nennt, wo eine Datei liegt — leer, solange keine gesichert ist.
  meldung: string
  gefragt: boolean
  laeuft: boolean
  loeschen: () => Promise<void>
  abbrechen: () => void
}

export function useTurnierhandlungen(
  stand: Turnierstand,
  turnier: TurnierDTO,
): Turnierhandlungen {
  const eigene = useTexte().seiten.turniere
  const [meldung, setzeMeldung] = useState('')
  const [gefragt, setzeGefragt] = useState(false)

  const marke = `turnier:${turnier.id}`
  const laeuft = stand.vorgang.laeuft(marke)

  const sichere = (weg: () => Promise<TurnierdateiAntwort>) => {
    void stand.lies(marke, async () => {
      const antwort = await weg()
      auspacken(antwort, () => undefined)

      setzeMeldung(
        antwort.abgebrochen
          ? ''
          : einsetzen(eigene.gesichertUnter, { pfad: antwort.pfad }),
      )
    })
  }

  const loeschen = async () => {
    await stand.fuehre(marke, async () => {
      auspacken(await api().TurnierLoeschen(turnier.id), () => undefined)
      stand.vergiss(turnier.id)
    })
  }

  const eintraege: Menueeintrag[] = [
    {
      schluessel: 'ausgeben',
      name: eigene.ausgeben,
      angehalten: laeuft,
      aufWahl: () => {
        sichere(() => api().TurnierAusgabeSpeichern(turnier.id))
      },
    },
    {
      schluessel: 'bericht',
      name: eigene.berichtSichern,
      aktion: 'sichern',
      angehalten: laeuft,
      aufWahl: () => {
        sichere(() => api().TurnierBerichtSpeichern(turnier.id))
      },
    },
    {
      schluessel: 'loeschen',
      name: eigene.loeschen,
      aktion: 'loeschen',
      schreibt: true,
      angehalten: laeuft || gefragt,
      aufWahl: () => {
        setzeGefragt(true)
      },
    },
  ]

  // **Dieselbe Liste in der Menüleiste** — Zweitblick über Phase 6 von
  // 0.9.0, zweiter Teil, Befund 6.
  useHandlungsliste({
    titel: eigene.turniermenue,
    art: 'turnier',
    kennung: turnier.id,
    eintraege,
  })

  return {
    eintraege,
    meldung,
    gefragt,
    laeuft,
    loeschen,
    abbrechen: () => {
      setzeGefragt(false)
    },
  }
}
