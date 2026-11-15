// **Wohin ⌘Z zielt, wenn eine Zeichnung offen ist** —
// #1340,
// Werkstück F von Phase 4.11.
//
// # Warum es diese Stelle gibt
//
// MSE2 hat für das Symbol ein **eigenes Fenster** mit eigenem `ActionStack`
// (`src/gui/symbol/window.cpp`, Fassung `f9d9356` vom 2021-07-06). Das ⌘Z des
// Setfensters erreicht es nie — die Frage stellt sich dort gar nicht.
//
// Hier ist der Editor eine **Seite** im selben Fenster, und beide Verläufe
// hängen an demselben Menüeintrag. **Das ist eine Abweichung mit Grund** (§15
// Satz 4): ein zweites Fenster für die Zeichnung wäre unter macOS ein zweites
// Dokumentfenster mit eigener Menüleiste, und §6 sagt, dass Bearbeitung im
// Seitenbereich stattfindet.
//
// # Warum ein Kontext und keine Eigenschaft
//
// Dieselbe Begründung wie bei `ablage.tsx` daneben: der Weg von `Anwendung`
// bis zum Symboleditor führt über Bauteile, die mit einer Rücknahme nichts zu
// tun haben. Eine Eigenschaft durch alle hindurchzureichen hieße, sie alle zu
// ändern, sobald jemand eine zweite Seite mit eigenem Verlauf baut.

import {
  createContext,
  useContext,
  useLayoutEffect,
  useState,
  type ReactNode,
} from 'react'

// Ruecknahmeweg ist, was eine offene Zeichnung beisteuert.
export interface Ruecknahmeweg {
  zuruecknehmen: () => void
  wiederherstellen: () => void
}

// Ruecknahmeziel ist die Stelle, an der ein Weg gemeldet wird.
export interface Ruecknahmeziel {
  // melden trägt den Weg ein — `undefined` nimmt ihn zurück.
  //
  // **Beim Zeichnen aufgerufen und nicht in einem Effekt.** Ein Effekt liefe
  // **nach** dem Zeichnen, und ein Menübefehl in derselben Aufgabe träfe noch
  // den vorigen Stand; dieselbe Falle wie bei den Horchern in §3.
  melden: (weg: Ruecknahmeweg | undefined) => void

  // gemeldetes gibt den Weg, den die offene Zeichnung eingetragen hat.
  gemeldetes: () => Ruecknahmeweg | undefined
}

// neuesRuecknahmeziel baut ein Ziel über einem eigenen, veränderlichen Stand.
export function neuesRuecknahmeziel(): Ruecknahmeziel {
  let weg: Ruecknahmeweg | undefined

  return {
    melden: (neuer) => {
      weg = neuer
    },
    gemeldetes: () => weg,
  }
}

const Kontext = createContext<Ruecknahmeziel | undefined>(undefined)

// Ruecknahmeraum stellt das Ziel bereit.
export function Ruecknahmeraum({ children }: { children: ReactNode }) {
  const [ziel] = useState(neuesRuecknahmeziel)

  return <Kontext.Provider value={ziel}>{children}</Kontext.Provider>
}

// useRuecknahmeziel gibt das Ziel.
//
// **Ohne Raum ein eigenes** — dieselbe Zusage wie bei `useAblage`: ein
// Bauteil, das außerhalb geprüft wird, soll nicht daran scheitern.
export function useRuecknahmeziel(): Ruecknahmeziel {
  const [ersatz] = useState(neuesRuecknahmeziel)

  return useContext(Kontext) ?? ersatz
}

// useRuecknahmeweg meldet den Weg an, solange das Bauteil sichtbar dasteht.
//
// **Und räumt ihn ab, wenn es geht** — §7 Regel 1. Das ist der Grund, warum
// es diesen Haken gibt und nicht nur `melden`: das Abräumen ist die Hälfte,
// die man vergisst. Wer sie dem Aufrufer überlässt, hat sie beim zweiten
// Aufrufer nicht mehr, und ⌘Z nähme auf der Kartenseite einen Handgriff an
// einer Zeichnung zurück, die niemand mehr sieht.
//
// **Gemeldet und abgeräumt im Layout-Effekt** — P8 (#2376). Hier stand
// „gemeldet beim Zeichnen“, damit ein Menübefehl nicht ins Leere trifft. Doch
// `<Activity>` zeichnet verborgene Seiten neu, und eine solche meldete sich
// dabei wieder an: ⌘Z auf einer anderen Seite traf dann sie. Ein
// Layout-Effekt läuft noch im selben Schritt wie das Zeichnen, bevor ein
// Menübefehl als nächste Aufgabe ankommt — und eine verborgene Seite hat
// keinen.
export function useRuecknahmeweg(weg: Ruecknahmeweg): void {
  const ziel = useRuecknahmeziel()

  useLayoutEffect(() => {
    ziel.melden(weg)

    return () => {
      ziel.melden(undefined)
    }
  }, [ziel, weg])
}
