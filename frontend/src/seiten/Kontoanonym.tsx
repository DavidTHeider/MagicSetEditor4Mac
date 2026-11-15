// **Ein Konto ohne Angaben — und der Weg ohne jedes** — Werkstück 6 von
// Phase 4.12
// (#1351).
//
// # Warum das eine eigene Fläche ist
//
// **Weil es keine Felder hat.** Ein Knopf ohne Formular zwischen zwei Feldern
// sähe aus, als gehörte er zu ihnen — und wer ihn drückte, hätte seine
// Eingaben verworfen, ohne es zu wollen.
//
// **Und weil daneben der vierte Weg steht**, der gar keiner ist: ohne Konto
// loslegen heißt, dass niemand fragt. Der Nutzer vom ersten Start ist da; das
// ist kein Knopf, sondern eine Auskunft (§6).
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Katalogfehler } from '../katalogfehler'
import { Knopf } from '../bausteine/Knopf'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// Der Schlüssel des Vorgangs (§7, Regel 4).
const ANONYM = 'anonym'

export function Kontoanonym({ geaendert }: { geaendert: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.konto

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)
  const [erfolg, setzeErfolg] = useState<string | undefined>(undefined)

  const vorgang = useVorgang()

  const anlegen = async () => {
    setzeFehler(undefined)
    setzeErfolg(undefined)

    await vorgang.starte(ANONYM, async () => {
      try {
        const antwort = await api().KontoAnonym()
        const angelegt = auspacken(antwort, () => antwort.nutzer)

        // **Eine Antwort ohne Nutzer und ohne Fehler ist ein Fehler** (§8) —
        // dieselbe Zeile wie in `Kontoanlage.tsx`, aus demselben Grund.
        if (angelegt === null) {
          throw new Katalogfehler('kontoLeer')
        }

        setzeErfolg(einsetzen(eigene.anonymAngelegt, { name: angelegt.name }))

        // **Geschrieben ist geschrieben** (§7, Regel 3).
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  return (
    <section aria-label={eigene.ohneAngaben} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß. */}
      <Erklaerungsabsatz
        text={eigene.ohneAngabenErklaerung}
        klasse="max-w-prose text-leise"
      />

      <div>
        <Knopf
          gewicht="leise"
          angehalten={vorgang.laeuft(ANONYM)}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.anonym}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {erfolg !== undefined && (
        <p role="status" className="text-leise">
          {erfolg}
        </p>
      )}
    </section>
  )
}
