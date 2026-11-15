// **Die Grenze an der Wurzel des Fensters** — F01
// (#2340).
//
// # Wofür
//
// **Null leere Fenster nach einem Fehler.** Was die Grenze eines Bereichs
// nicht fängt — einen Wurf im Rahmen, in den Räumen der Anwendung, beim
// Anmelden —, nahm bis hierher den ganzen Baum mit, und das Fenster stand
// weiß da. Hier steht dann ein Satz mit Kennung und ein Knopf, der das
// Fenster neu lädt; der Bestand bleibt, wie er ist.
//
// # Ohne die Räume der Anwendung
//
// **Die Sprache kommt vom System**, nach derselben Regel wie im Fenster
// (`spracheAus`), aber ohne die Einstellungen: gerade sie könnten es sein,
// die nicht zeichnen.
import { useEffect, useId, useRef, type ReactNode } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { spracheAus } from '../sprache/einstellung'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum, useTexte } from '../sprache/kontext'
import { Fehlergrenze, type Fehlerstand } from './Fehlergrenze'

interface Eigenschaften {
  children: ReactNode

  // neuLaden lädt das Fenster neu — eingesetzt, damit ein Test es sieht.
  neuLaden?: () => void

  // systemsprache ist die Sprache des Rechners — eingesetzt aus demselben Grund.
  systemsprache?: string
}

export function Wurzelgrenze({
  children,
  neuLaden = () => {
    location.reload()
  },
  systemsprache = navigator.language,
}: Eigenschaften) {
  return (
    <Fehlergrenze
      ebene="wurzel"
      bereich="fenster"
      zeige={(stand) => (
        <Sprachraum sprache={spracheAus('', systemsprache)}>
          <Wurzelfehler {...stand} neuLaden={neuLaden} />
        </Sprachraum>
      )}
    >
      {children}
    </Fehlergrenze>
  )
}

function Wurzelfehler({
  kennung,
  neuLaden,
}: Fehlerstand & { neuLaden: () => void }) {
  const eigene = useTexte().allgemein.fehlergrenze
  const titel = useId()
  const knopf = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    knopf.current?.focus()
  }, [])

  return (
    <main className="flex min-h-screen items-center justify-center bg-grund p-8 text-text">
      <section
        role="alert"
        aria-labelledby={titel}
        className="flex max-w-prose flex-col gap-3"
      >
        <h1 id={titel} className="text-lg font-semibold">
          {eigene.wurzelTitel}
        </h1>

        <p>{einsetzen(eigene.wurzelSatz, { kennung })}</p>

        <div>
          <Knopf gewicht="haupt" griff={knopf} aufKlick={neuLaden}>
            {eigene.fensterNeuLaden}
          </Knopf>
        </div>
      </section>
    </main>
  )
}
