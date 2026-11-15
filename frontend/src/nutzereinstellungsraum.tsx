// **Was dem Menschen gehört, gilt überall** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkte 9 und 11.
//
// # Warum ein Raum und keine Eigenschaft
//
// **Weil zwei Stellen sie brauchen und keine über der anderen liegt.** Das
// Wurzelelement bekommt `data-bewegung` und `data-kontrast` — sie gelten für
// **alles**, auch für den Rollbalken, den die Anwendung nicht zeichnet. Die
// Einstellungsseite zeigt dieselben Werte und ändert sie. Ein Weg dorthin
// über Eigenschaften ginge durch fünf Bauteile, die damit nichts zu tun
// haben — dieselbe Überlegung wie beim `Ablageraum`.
//
// # Und warum er unter dem Sprachraum steht
//
// **Weil er einen Bestand voraussetzt.** `App` zeichnet die
// Wiederherstellungsseite, solange keiner offen ist; ein Ruf darüber liefe in
// einen Nullzeiger, und Wails riefe den Rückweg nie.
import { createContext, useContext, useEffect, type ReactNode } from 'react'

import { setzeBewegung, setzeKontrast } from './erscheinungsbild'
import {
  useNutzereinstellungen,
  type Nutzereinstellungsstand,
} from './haken/useNutzereinstellungen'

const Kontext = createContext<Nutzereinstellungsstand | undefined>(undefined)

// Nutzereinstellungsraum lädt sie und legt sie ans Wurzelelement.
export function Nutzereinstellungsraum({
  stand = 0,
  children,
}: {
  stand?: number
  children: ReactNode
}) {
  const einstellungen = useNutzereinstellungen(stand)

  const bewegung = einstellungen.wert('barrierefreiheit.bewegung')
  const kontrast = einstellungen.wert('barrierefreiheit.kontrast')

  // **Am Wurzelelement und nicht in React** — dieselbe Begründung wie beim
  // Erscheinungsbild: beides gilt für alles.
  useEffect(() => {
    setzeBewegung(document.documentElement, bewegung)
    setzeKontrast(document.documentElement, kontrast)
  }, [bewegung, kontrast])

  return <Kontext.Provider value={einstellungen}>{children}</Kontext.Provider>
}

// useNutzereinstellungsstand gibt den Stand aus dem Raum.
//
// **Ohne Raum ein leerer Stand und kein Absturz.** Ein Bauteil, das ihn
// außerhalb ruft, ist ein Fehler — aber einer, der eine leere Auswahlliste
// zeigt und nicht das Fenster weiß macht.
export function useNutzereinstellungsstand(): Nutzereinstellungsstand {
  return (
    useContext(Kontext) ?? {
      eintraege: [],
      wert: () => '',
      nutzerID: '',
      fehler: undefined,
      setze: () => undefined,
      laeuft: () => false,
    }
  )
}
