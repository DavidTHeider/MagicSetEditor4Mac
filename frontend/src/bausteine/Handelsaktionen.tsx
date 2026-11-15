// **Mit diesem Handel: kopieren, Entwurf löschen, archivieren** —
// TRADE-01.AK-03 (#1715).
//
// **Was hinausgegangen ist, wird nicht gelöscht, sondern kopiert oder
// archiviert.** Löschen gibt es nur für einen Entwurf — und dann mit einer
// Rückfrage, denn es gibt keine Rücknahme. Ins Archiv geht nur, was zu Ende
// ist; der Dienst sagt es auch, der Eintrag fehlt vorher schon.
//
// # Warum ein Menü und keine Knopfreihe
//
// **Weil es Handgriffe sind und kein nächster Schritt** (UXR-15,
// #1923).
// Die Reihe stand gleich unter den Schritten und ihrem Hauptknopf — bei
// einem Entwurf zwei Knöpfe, bei einem beendeten Handel zwei und ein Satz.
// Jetzt steht dort ein Menü „Mit diesem Handel"; die Rückfrage zum Löschen
// steht wie bisher an seiner Stelle, nicht in einem Dialog darüber.
import { useState } from 'react'

import { api } from '../fassade'
import type { Vorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Entscheidung } from './Entscheidung'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Menue, type Menueeintrag } from './Menue'
import { useHandlungsliste } from '../haken/useHandlungsliste'

import type { Handelsvorgang } from '../apihandelsleitung'

interface Handelsaktioneneigenschaften {
  vorgang: string
  stand: Handelsvorgang | undefined

  // lauf und kennzeichen sind der Riegel des Raums — **derselbe wie für
  // Schritte und Vorschläge** (§7, Regel 4; Befund Z4 des Zweitblicks über
  // Phase 6.8): zwei Schreibaufrufe an denselben Handel gingen sonst
  // gleichzeitig.
  lauf: Vorgang
  kennzeichen: string

  // archiviert sagt, ob der Vorgang im Archiv liegt — bekannt, wenn er aus
  // der Liste aufgeschlagen wurde.
  archiviert: boolean

  // hinweis sagt, was eben geschah — der Raum hält ihn, denn nach dem
  // Kopieren ist ein anderer Vorgang gewählt und diese Leiste eine neue.
  hinweis?: string | undefined

  // kopiert wählt die Kopie; verworfen räumt die Wahl; archiv lädt nach.
  kopiert: (kopie: Handelsvorgang) => void
  verworfen: () => void
  archiv: () => void
}

export function Handelsaktionen({
  vorgang,
  stand,
  lauf,
  kennzeichen,
  archiviert,
  hinweis,
  kopiert,
  verworfen,
  archiv,
}: Handelsaktioneneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln.liste
  const [fragt, setzeFragt] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  const fahre = (ruf: () => Promise<void>) =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      try {
        await ruf()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const laeuft = lauf.laeuft(kennzeichen)

  // **Was der Zustand nicht erlaubt, steht nicht im Menü** — ein Entwurf
  // lässt sich löschen, ein beendeter Handel archivieren.
  const eintraege: Menueeintrag[] = [
    {
      schluessel: 'kopieren',
      schreibt: true,
      name: eigene.kopieren,
      aktion: 'kopieren',
      angehalten: laeuft,
      aufWahl: () => {
        void fahre(async () => {
          const antwort = await api().HandelKopieren(vorgang)

          kopiert(auspacken(antwort, () => antwort.vorgang))
        })
      },
    },
  ]

  if (stand?.zustand === 'entwurf' && !fragt) {
    eintraege.push({
      schluessel: 'loeschen',
      schreibt: true,
      name: eigene.entwurfLoeschen,
      aktion: 'loeschen',
      angehalten: laeuft,
      aufWahl: () => {
        setzeFragt(true)
      },
    })
  }

  if (stand?.ende === true) {
    eintraege.push({
      schluessel: 'archiv',
      schreibt: true,
      name: archiviert ? eigene.hervorholen : eigene.archivieren,
      angehalten: laeuft,
      aufWahl: () => {
        void fahre(async () => {
          const antwort = await api().HandelArchivieren(vorgang, !archiviert)

          auspacken(antwort, () => undefined)
          archiv()
        })
      },
    })
  }

  // **Dieselbe Liste an drei Orten** (GESTE, #2358): das Menü hier, der
  // Rechtsklick auf die Schritte des Handels (`Handelsraum`) und das Menü
  // „Handel" in der Menüleiste.
  const handlungen = useHandlungsliste({
    titel: eigene.handelsmenue,
    art: 'handel',
    kennung: vorgang,
    eintraege,
  })

  return (
    <div className="flex flex-col gap-2">
      <div>
        <Menue bezeichnung={eigene.aktionen} eintraege={handlungen} />
      </div>

      {stand?.ende === true && (
        <Erklaerungsabsatz
          text={eigene.archivErklaerung}
          klasse="text-sm text-leise"
        />
      )}

      {fragt && (
        <Entscheidung
          frage={eigene.entwurfLoeschenFrage}
          bestaetigen={eigene.entwurfLoeschenJa}
          aktion="loeschen"
          abbrechen={eigene.abbrechen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeFragt(false)
            void fahre(async () => {
              const antwort = await api().HandelEntwurfLoeschen(vorgang)

              auspacken(antwort, () => undefined)
              verworfen()
            })
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      )}

      {hinweis !== undefined && (
        <p role="status" className="text-sm text-leise">
          {hinweis}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
