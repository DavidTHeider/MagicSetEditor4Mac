// **Der Abschluss eines Turniers in der Werkbank** — DOM-04.AK-04
// (#1699),
// Vermerk 0139.
//
// **Offen**: ein Knopf *Abschließen* — ob es geht, entscheidet der Dienst
// (mindestens eine Runde, kein offener Tisch), und sein Nein steht über der
// Seite. **Abgeschlossen**: das Datum und *Wiedereröffnen* mit Rückfrage an
// der Klickstelle (Muster *Rückfrage*, `docs/UX-ARCHITEKTUR.md`).
import { useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import type { TurnierDTO } from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'

export function Turnierabschluss({
  stand,
  turnier,
}: {
  stand: Turnierstand
  turnier: TurnierDTO
}) {
  const eigene = useTexte().seiten.turniere
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  const [fragt, setzeFragt] = useState(false)

  const marke = `turnier:${turnier.id}`
  const laeuft = stand.vorgang.laeuft(marke)

  const abschliessen = async () => {
    await stand.fuehre(marke, async () => {
      auspacken(
        await api().TurnierAbschliessen(turnier.id, turnier.fassung),
        () => undefined,
      )
    })
  }

  const wiedereroeffnen = async () => {
    setzeFragt(false)

    await stand.fuehre(marke, async () => {
      auspacken(
        await api().TurnierWiedereroeffnen(turnier.id, turnier.fassung),
        () => undefined,
      )
    })
  }

  if (turnier.abgeschlossen === '') {
    return (
      <div>
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            void abschliessen()
          }}
        >
          {eigene.abschliessen}
        </Knopf>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm">
        {einsetzen(eigene.abgeschlossenAm, {
          datum: alsZeitpunkt(turnier.abgeschlossen, darstellung, zeitzone),
        })}
      </p>

      {fragt ? (
        <Entscheidung
          frage={eigene.wiedereroeffnenFrage}
          erklaerung={eigene.wiedereroeffnenErklaerung}
          bestaetigen={eigene.wiedereroeffnen}
          abbrechen={eigene.nichtWiedereroeffnen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            void wiedereroeffnen()
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      ) : (
        <div>
          <Knopf
            angehalten={laeuft}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.wiedereroeffnen}
          </Knopf>
        </div>
      )}
    </div>
  )
}
