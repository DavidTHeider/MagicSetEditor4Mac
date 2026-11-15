// **Was bisher geschah** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum das neben „Zuletzt geschehen" steht und es nicht ersetzt
//
// **Weil die beiden verschiedene Fragen beantworten.** Die Ereignisliste in
// `Partiebefehle` zeigt, was der **letzte Befehl** bewirkt hat — sie ist die
// Quittung auf einen Klick. Hier steht der **ganze** Strom, und er kommt aus
// dem Bestand statt aus der Antwort von eben. Wer die Seite neu lädt oder den
// Sitz wechselt, hat die Quittung verloren und den Verlauf nicht.
//
// # Die Grenze steht dabei und wird nicht bloß eingehalten
//
// **Weil man ihr sonst nicht ansieht, dass es sie gibt.** Ein Verlauf sieht
// aus wie eine vollständige Buchführung; dass er gekürzt ist, ist eine
// Zusage, und eine Zusage, die niemand liest, wirkt wie ein Mangel (§6:
// leere Zustände erklären den nächsten Schritt).
import { useId } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { usePartieverlauf } from '../haken/usePartieverlauf'
import { useTexte } from '../sprache/kontext'

import type { Spielereignis } from '../apipartie'
import { Abschnittstitel } from '../bausteine/Abschnittstitel'

interface Verlaufseigenschaften {
  kennung: string
  fuer: string
}

export function Partieverlauf({ kennung, fuer }: Verlaufseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = usePartieverlauf(kennung, fuer)

  // **Der Abschnitt trägt seinen Namen** — ohne `aria-labelledby` ist ein
  // `section` für ein Vorleseprogramm kein Bereich, sondern ein Kasten ohne
  // Aufschrift, und niemand kann ihn anspringen. Die Überschrift steht
  // ohnehin da; sie hier zu verknüpfen kostet eine Kennung.
  const ueberschrift = useId()

  return (
    <section className="flex flex-col gap-2" aria-labelledby={ueberschrift}>
      <div className="flex items-center gap-3">
        <Abschnittstitel id={ueberschrift} className="text-base font-semibold">
          {eigene.verlauf}
        </Abschnittstitel>

        <Knopf gewicht="leise" aufKlick={lauf.ansehen} angehalten={lauf.laeuft}>
          {lauf.verlauf === undefined
            ? eigene.verlaufAnsehen
            : eigene.verlaufNachladen}
        </Knopf>
      </div>

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      <Verlaufsstand lauf={lauf} />
    </section>
  )
}

// Verlaufsstand zeigt genau einen der vier Zustände.
//
// **Vier und nicht drei**: „läuft", „noch nicht geholt", „geholt und leer"
// und „geholt und voll". Die mittleren beiden sehen gleich aus und sind es
// nicht — das eine sagt, dass jemand klicken muss, das andere, dass es
// nichts zu sehen gibt (§6).
function Verlaufsstand({
  lauf,
}: {
  lauf: ReturnType<typeof usePartieverlauf>
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (lauf.laeuft) {
    return <p className="text-sm text-leise">{eigene.verlaufLaeuft}</p>
  }

  if (lauf.verlauf === undefined) {
    return <p className="text-sm text-leise">{eigene.verlaufUngeholt}</p>
  }

  if (lauf.verlauf.length === 0) {
    return <p className="text-sm text-leise">{eigene.verlaufLeer}</p>
  }

  return (
    <>
      <ol className="flex flex-col gap-1 text-sm">
        {lauf.verlauf.map((ereignis) => (
          <Verlaufszeile key={ereignis.folge} ereignis={ereignis} />
        ))}
      </ol>

      <p className="text-xs text-leise">{eigene.verlaufGrenze}</p>
    </>
  )
}

// Verlaufszeile ist ein Zug — mit seiner Folge, damit die Reihenfolge
// ablesbar bleibt und eine Lücke auffiele.
function Verlaufszeile({ ereignis }: { ereignis: Spielereignis }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  return (
    <li className="flex flex-col">
      <span>
        {ereignis.folge}. {ereignis.art}
      </span>

      {Object.keys(ereignis.privat).length > 0 && (
        <span className="font-mono text-xs text-leise">
          {eigene.ereignisPrivat}: {JSON.stringify(ereignis.privat)}
        </span>
      )}
    </li>
  )
}
