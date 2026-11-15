// **Eine Vorlage aus einem Umschlag annehmen** —
// #1639.
//
// # Warum sie hier steht
//
// **Weil eine Vorlage ein Paket ist und Pakete hier verwaltet werden.** Der
// Umschlag trägt die **Dateien** und nicht den Namen, aus dem sie stammen;
// wie das Paket hier heißen soll, entscheidet ein Mensch.
//
// # Und warum der Name seine Endung trägt
//
// **Weil sie sagt, was das Paket ist** — `meinspiel.mse-game` ist ein Spiel,
// `meinstil.mse-style` ein Stil. Eine Auswahl daneben wäre dieselbe Angabe
// zweimal, und zwei Angaben laufen auseinander.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

// feldstil ist der Rahmen des Namensfeldes — mit sichtbarem Fokusring (§6).
const feldstil =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Vorlagenumschlageigenschaften {
  // angenommen läuft, nachdem ein Paket hereingekommen ist.
  angenommen: () => void
}

export function Vorlagenumschlag({
  angenommen,
}: Vorlagenumschlageigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.pakete
  const lauf = useVorgang()

  const feld = useId()

  const [name, setzeName] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [satz, setzeSatz] = useState<string>()

  const kennzeichen = 'vorlagenumschlag'

  const annehmen = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeSatz(undefined)

      if (name.trim() === '') {
        setzeFehler(eigene.umschlagOhneNamen)

        return
      }

      try {
        const antwort = await api().UmschlagVorlage(name)

        const paket = auspacken(antwort, () => antwort.paket)

        // **Ein abgebrochener Dialog ist keine Meldung wert** (§6).
        if (paket !== null) {
          setzeSatz(eigene.umschlagAngenommen)
          angenommen()
        }
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  // **Ohne eigene Überschrift und ohne Trennlinie** — seit UXR-17 (#1925)
  // steht der Umschlag in einem Kasten der Schublade *Paket beziehen*, und
  // dessen Schalter trägt den Namen. Bis dahin stand er als eigener Abschnitt
  // unter der Paketliste, mit Linie, damit er sich nicht als ihre letzte Zeile
  // las (Sichtprüfung am 2026-09-19).
  return (
    <section className="flex flex-col gap-2">
      {/* **Wie der Name aussehen muss, steht im ersten Satz** — er ist
          nötig, um das Feld darunter zu füllen, und gehört deshalb nicht
          hinter das Hilfesymbol (UX-15.AK-02). */}
      <Erklaerungsabsatz text={eigene.umschlagErklaerung} />

      <label className="flex flex-col gap-1 text-sm" htmlFor={feld}>
        {eigene.umschlagName}
        <input
          id={feld}
          type="text"
          // Der Name wird zum Verzeichnis — keine Rechtschreibkorrektur.
          spellCheck={false}
          className={feldstil}
          value={name}
          onChange={(ereignis) => {
            setzeName(ereignis.target.value)
          }}
        />
      </label>

      <div>
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void annehmen()
          }}
        >
          {eigene.umschlagAnnehmen}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {satz !== undefined && (
        <p role="status" className="text-sm text-leise">
          {satz}
        </p>
      )}
    </section>
  )
}
