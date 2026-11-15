// Die **Namensliste** der Skriptkonsole
// (#1164).
//
// # Warum sie eine eigene Datei ist
//
// Weil sie eine eigene Frage beantwortet (§2): die Konsole fragt „was ergibt
// dieser Ausdruck?", diese Liste „welche Namen darf ich überhaupt nennen?".
//
// # Warum sie eingrenzt und deckelt
//
// Am laufenden Fenster gezählt, an einer Karte aus `Vanguard.mse-set` mit
// eingebundenem `m15-altered`: **1.644 Namen**. Ohne Vorlage sind es gut
// neunzig. Beide Male gilt derselbe Satz — niemand liest eine Reihe von
// Knöpfen dieser Länge, und für den, der MSE-Script nicht kennt, ist sie so
// wenig eine Einladung wie ein leeres Feld.
//
// **Gedeckelt wird sichtbar** (§10): eine Liste, die einfach aufhört, liest
// sich wie eine vollständige.
import { useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

// hoechstens ist, wie viele Knöpfe zugleich dastehen.
//
// **Sechzig und nicht alle.** Die Zahl ist gegriffen und soll es sein: sie
// füllt gut zwei Reihen im Fenster und lässt erkennen, dass eingegrenzt werden
// muss. Was sie **nicht** ist, ist eine Aussage über die Umgebung — deren
// Größe steht daneben.
const hoechstens = 60

interface NamenslisteEigenschaften {
  // namen sind alle Namen der Umgebung — sortiert und ohne Doppel.
  namen: string[]

  // einsetzen hängt einen angeklickten Namen an den Ausdruck.
  einsetzen: (name: string) => void
}

export function Namensliste({
  namen,
  einsetzen: uebernehmen,
}: NamenslisteEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.konsole

  const [eingrenzung, setzeEingrenzung] = useState('')

  // **Ohne Rücksicht auf Groß- und Kleinschreibung.** MSE-Bezeichner sind
  // durchweg klein, aber wer `Card` tippt, meint `card`.
  const passend = namen.filter((name) =>
    name.toLowerCase().includes(eingrenzung.toLowerCase()),
  )
  const gezeigt = passend.slice(0, hoechstens)

  return (
    <>
      <label className="flex max-w-sm flex-col gap-1">
        <span className="text-sm font-medium">{eigene.namenFiltern}</span>
        <input
          type="search"
          value={eingrenzung}
          onChange={(ereignis) => {
            setzeEingrenzung(ereignis.target.value)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      <p className="text-sm text-leise">
        {einsetzen(eigene.namenGezeigt, {
          gezeigt: String(gezeigt.length),
          ganz: String(namen.length),
        })}
      </p>

      {gezeigt.length === 0 ? (
        <p className="text-sm text-leise">{eigene.namenLeer}</p>
      ) : (
        <ul className="flex flex-wrap gap-1">
          {gezeigt.map((name) => (
            <li key={name}>
              <Knopf
                bezeichnung={name}
                aufKlick={() => {
                  uebernehmen(name)
                }}
              >
                <span className="font-mono text-xs">{name}</span>
              </Knopf>
            </li>
          ))}
        </ul>
      )}
    </>
  )
}
