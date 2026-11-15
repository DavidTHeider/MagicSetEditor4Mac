// **Alle Pakete mit Befund in einer Zeile** — AP-8
// (#2365),
// Befund B-14. Die Zeile öffnet die Paketliste, gefiltert auf genau diese
// Pakete; die Namen kommen aus derselben Antwort wie die Zahl.
//
// # Die Gründe auf Wunsch
//
// Bis dahin stand je Paket sein erster Skriptfehler da, in der Sprache des
// Fensters (#1963). Die Paketliste nennt Skriptfehler nicht — eine Zahl
// allein nähme die Auskunft weg. **Sie steht deshalb darunter, zugeklappt.**
import { useId, useState } from 'react'

import { useAnschluss } from '../anschlussraum'
import type { HinweisDTO } from '../apiaufmerksamkeit'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { grundsatz } from './aufmerksamkeitszeilen'

export function Paketsammelzeile({
  hinweise,
  klasse,
}: {
  hinweise: readonly HinweisDTO[]

  // klasse ist die der übrigen Zeilen, damit die Liste eine bleibt.
  klasse: string
}) {
  const texte = useTexte()
  const eigene = texte.seiten.start.aufmerksamkeit
  const anschluss = useAnschluss()
  const [gruende, setzeGruende] = useState(false)
  const liste = useId()

  const satz = einsetzen(eigene.paketeMitBefund, {
    pakete: zaehlwort(
      hinweise.length,
      texte.allgemein.zaehlwoerter.eigenePakete,
    ),
  })

  return (
    <div className="flex flex-col">
      <button
        type="button"
        className={klasse}
        onClick={() => {
          anschluss.uebergib({
            art: 'paketauswahl',
            name: satz,
            namen: hinweise.map((hinweis) => hinweis.objekt.kennung),
          })
        }}
      >
        <span className="font-medium">{satz}</span>
      </button>

      <button
        type="button"
        aria-expanded={gruende}
        aria-controls={liste}
        className="self-start rounded-sm px-2 py-0.5 text-xs underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        onClick={() => {
          setzeGruende((bisher) => !bisher)
        }}
      >
        {gruende ? eigene.gruendeVerbergen : eigene.gruendeZeigen}
      </button>

      {gruende && (
        <ul id={liste} className="flex flex-col gap-1 px-2 py-1 text-xs">
          {hinweise.map((hinweis) => (
            <li key={hinweis.objekt.kennung}>
              <span className="font-medium">{hinweis.objekt.name}</span>{' '}
              <span className="text-leise">
                {einsetzen(eigene.arten.paketBefund, {
                  anzahl: String(hinweis.anzahl),
                })}
              </span>
              {hinweis.grund !== '' && (
                <span className="block text-leise">
                  {grundsatz(hinweis, texte)}
                </span>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
