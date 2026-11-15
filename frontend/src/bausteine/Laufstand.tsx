// **Der Stand eines Datei-Laufs, wo er angestoßen wurde** — FMT-03.AK-02
// (#1692).
//
// Ein Druckbogen, MTG JSON, ein Installer: wer ihn anstößt, sieht an **der
// Seite, auf der er es tat**, wie weit er ist, und kann ihn anhalten — nicht
// erst auf der Startseite unter den Hintergrundaufgaben. Die Gestalt ist die
// des [Fortschritt]s; hier steht, woher der Stand kommt und wie der Satz in
// der Sprache des Fensters entsteht.
//
// **Angehalten wird zwischen zwei Stücken** (`service.Schritt`): der Knopf
// sagt deshalb erst „wird angehalten", und „abgebrochen" steht da, wenn Go
// es meldet — nicht, wenn der Klick zurückkommt.
import { useState } from 'react'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { Fortschritt } from './Fortschritt'
import { useLaufstand } from '../haken/useLaufstand'
import { laufsatz } from './laufsatz'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Laufstand({ art }: { art: string }) {
  const texte = useTexte().bausteine.laufstand
  const alleTexte = useTexte()
  const auftrag = useLaufstand(art)

  // **Je Auftrag, nicht je Seite** (§7, Regel 2): ein Anhalten gehört zu dem
  // Lauf, für den es verlangt wurde; ein neuer beginnt ohne.
  const [verlangt, setzeVerlangt] = useState<string>()
  const [fehler, setzeFehler] = useState<{ id: string; satz: string }>()

  if (
    auftrag === undefined ||
    auftrag.stand === 'fertig' ||
    auftrag.stand === 'gescheitert'
  ) {
    return null
  }

  // **Was bis zum Anhalten fertig wurde, steht dabei** — bei einer
  // Exportvorlage bleiben geschriebene Bilder liegen, und wie viele es sind,
  // sagt der letzte gezählte Stand.
  if (auftrag.stand === 'abgebrochen') {
    return (
      <p role="status" className="mt-3 text-leise">
        {auftrag.erledigt > 0
          ? `${texte.abgebrochen} ${einsetzen(texte.bisDahin, { stand: laufsatz(texte, auftrag) })}`
          : texte.abgebrochen}
      </p>
    )
  }

  const anhalten = async (id: string) => {
    setzeVerlangt(id)
    setzeFehler(undefined)

    try {
      const antwort = await api().AuftragAbbrechen(id)
      auspacken(antwort, () => undefined)
    } catch (ursache: unknown) {
      setzeVerlangt(undefined)
      setzeFehler({ id, satz: meldungAus(ursache, alleTexte) })
    }
  }

  const angehalten = verlangt === auftrag.id

  return (
    <div className="mt-3">
      <Fortschritt
        satz={laufsatz(texte, auftrag)}
        erledigt={auftrag.erledigt}
        gesamt={auftrag.gesamt}
        unbestimmt={auftrag.gesamt === 0}
        {...(auftrag.abbrechbar && {
          anhalten: {
            beschriftung: angehalten ? texte.wirdAngehalten : texte.abbrechen,
            angehalten,
            aufKlick: () => {
              void anhalten(auftrag.id)
            },
          },
        })}
      />
      {fehler?.id === auftrag.id && (
        <p role="alert" className="mt-2 text-fehler">
          {alleTexte.bausteine.zustaende.fehler} {fehler.satz}
        </p>
      )}
    </div>
  )
}
