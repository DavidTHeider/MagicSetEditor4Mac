// **Was eine Seite über ihren Weg sagt** — UX-03.AK-02
// (#1675).
//
// # Wozu
//
// **Die Brotkrume steht im Rahmen, und der Rahmen kennt keine
// Bezugsobjekte** (#52) — das offene Set reicht ihm das Fenster. Die Partie
// über dem Tisch kennt nur die Spielen-Seite: ob ein Tisch offensteht und
// wer daran sitzt. Das Artboard *Tisch* sagt *Spielen › Partie gegen Bo › Am
// Tisch*.
//
// # Wie
//
// **Die Seite sagt ihren Weg, und der Rahmen setzt ihn an die Stelle ihres
// Namens** (`brotkrumen`). Wer die Seite verlässt, nimmt den Weg mit — der
// Aufräumer im Effekt; ein Weg, der eine Seite überlebte, stünde über der
// nächsten.
import { createContext, use, useEffect, type ReactNode } from 'react'

import type { Krume } from './brotkrumen'

// Wegsetzer nimmt den Weg einer Seite entgegen — `undefined` heißt: keiner.
type Wegsetzer = (weg: readonly Krume[] | undefined) => void

// **Ohne Rahmen tut der Setzer nichts**: eine Seite, die ein Lauf allein
// zeichnet, hat keine Brotkrume, der sie etwas sagen könnte.
const Kontext = createContext<Wegsetzer>(() => undefined)

// Seitenweggeber reicht den Setzer des Rahmens an die Seite darunter.
export function Seitenweggeber({
  setze,
  children,
}: {
  setze: Wegsetzer
  children: ReactNode
}) {
  return <Kontext.Provider value={setze}>{children}</Kontext.Provider>
}

// useSeitenweg sagt dem Rahmen, wie der Weg hinter dem Bereich weitergeht.
//
// **Verglichen wird der Inhalt und nicht der Verbund**: eine Seite baut ihren
// Weg bei jedem Zeichnen neu, und ein Effekt über den Verbund liefe jedes
// Mal — setzen, aufräumen, setzen, für denselben Weg.
export function useSeitenweg(weg: readonly Krume[] | undefined) {
  const setze = use(Kontext)

  const inhalt = weg === undefined ? '' : JSON.stringify(weg)

  useEffect(() => {
    setze(inhalt === '' ? undefined : (JSON.parse(inhalt) as Krume[]))

    return () => {
      setze(undefined)
    }
  }, [setze, inhalt])
}
