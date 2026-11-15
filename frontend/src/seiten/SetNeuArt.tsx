// **Der erste Schritt zum neuen Set: die Art** — nach dem Artboard *Sets*
// (#1677):
// „Wofür sind die Karten? Den Paketnamen suchen wir für dich."
//
// # Was hier steht
//
// **Jedes nutzbare Spiel, beim Namen und mit der Zahl seiner Stile** —
// nutzbar heißt: es gibt einen Stil dazu (UX-05.AK-04). Gibt es nur eines,
// steht dieser Schritt gar nicht da; das entscheidet `SetNeu`.
//
// **Und „etwas ganz Eigenes"** führt zu den Paketen, wo ein leeres
// Spielpaket entsteht — dieselbe Stelle, zu der ein fehlendes Spiel führt
// (S07.NAT-04).
import { useId } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { zaehlwort } from '../sprache/zaehlwort'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { PaketDTO } from '../api'

interface ArtEigenschaften {
  spiele: readonly PaketDTO[]
  stileJeSpiel: (spiel: string) => number
  gewaehlt: string
  waehlen: (spiel: string) => void
  zuDenPaketen: () => void
}

// wahlstil ist das Optionsfeld — **mit sichtbarem Fokusring** (§6), wie die
// Wahl des Regelmodus: ein Feld, das man mit der Tastatur erreicht und dabei
// nicht sieht, ist nicht bedienbar (`internal/werkzeuge/fokusring_test.go`).
const wahlstil =
  'mt-1 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

export function SetNeuArt({
  spiele,
  stileJeSpiel,
  gewaehlt,
  waehlen,
  zuDenPaketen,
}: ArtEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.sets
  const gruppe = useId()

  return (
    <div className="flex flex-col gap-2">
      <fieldset className="flex flex-col gap-1">
        <legend className="mb-1 text-sm">{worte.neuArtFrage}</legend>

        {spiele.map((spiel) => (
          <label
            key={spiel.name}
            className="flex cursor-pointer items-start gap-2 rounded-sm border border-kante px-2 py-1.5 has-[:checked]:border-marke"
          >
            <input
              type="radio"
              name={gruppe}
              value={spiel.name}
              checked={spiel.name === gewaehlt}
              onChange={() => {
                waehlen(spiel.name)
              }}
              className={wahlstil}
            />
            {/* **Der Name steht unmittelbar im Kasten**, nicht in einer
                eigenen Hülle: eine Ebene tiefer fände ihn
                `label-has-associated-control` nicht mehr. */}
            <span className="flex flex-col font-medium">
              {spiel.anzeigename}
              <span className="text-xs font-normal text-leise">
                <span className="font-mono">{spiel.name}</span>{' '}
                {einsetzen(worte.neuArtZusatz, {
                  stile: zaehlwort(
                    stileJeSpiel(spiel.name),
                    texte.allgemein.zaehlwoerter.stile,
                  ),
                })}
              </span>
            </span>
          </label>
        ))}
      </fieldset>

      <div>
        <Knopf gewicht="leise" aufKlick={zuDenPaketen}>
          {worte.neuEigenes}
        </Knopf>
      </div>
    </div>
  )
}
