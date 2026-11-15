// **Was das Fenster mit einer gezogenen Datei tut** —
// #891,
// Werkstück 4 von Phase Y.
//
// # Warum ein Kontext und keine Eigenschaft
//
// Wails meldet ein Ablegen **einmal für das ganze Fenster**
// (`runtime.OnFileDrop`), und `OnFileDropOff` nimmt **jede** Anmeldung
// zurück. Es kann also genau einen Horcher geben.
//
// Bis zum 2026-08-25 saß er in `useBildfeld` — auf der **Kartenseite**. Wer
// dort ein Set hinzog, bekam es als **Bild** ins Feld darunter; wer es
// irgendwo sonst hinzog, bekam gar nichts. Beides war falsch, und beides
// dieselbe Ursache: der Horcher stand eine Ebene zu tief.
//
// Jetzt steht er in `Anwendung`, und die Kartenseite trägt hier ein, was sie
// mit einem Bild vorhat. Ein Kontext, weil der Weg von `Anwendung` bis zur
// Kartenvorschau über fünf Bauteile führt, die damit nichts zu tun haben.
//
// # Warum Methoden und keine Verweise
//
// Der erste Anlauf legte zwei `useRef` in den Kontext. Der React-Übersetzer
// weist das ab — ein Verweis aus einem Kontext ist für ihn unveränderlich, und
// ihn beim Zeichnen zu lesen ist ein Fehler. Er hat recht: was hier geteilt
// wird, ist **Verhalten**, kein Zustand des Zeichnens.
import { createContext, useContext, useState, type ReactNode } from 'react'

// Ablagestelle ist, was die Kartenseite beisteuert.
export interface Ablagestelle {
  // merkeFeld hält fest, über welchem Feld der Zeiger steht.
  //
  // **Kein Zustand**: es ändert sich bei jeder Mausbewegung über der Karte,
  // und ein Zustand ließe die Seite dabei neu zeichnen.
  merkeFeld: (feld: string | undefined) => void

  // nimmFeld gibt das gemerkte Feld **und vergisst es**.
  //
  // **Beides in einem Zug**, weil es zusammengehört: das gemerkte Feld gilt
  // genau einmal. Wer danach neben die Karte zieht, setzte sonst ein Bild in
  // das Feld von vorhin.
  nimmFeld: () => string | undefined

  // setzeBildweg trägt ein, was mit einem Bild geschieht.
  setzeBildweg: (
    weg: ((feld: string, pfad: string) => void) | undefined,
  ) => void

  // bildweg gibt den eingetragenen Weg — undefined, wenn keine Karte offen ist.
  bildweg: () => ((feld: string, pfad: string) => void) | undefined
}

// neueAblagestelle baut eine Stelle über einem eigenen, veränderlichen Stand.
export function neueAblagestelle(): Ablagestelle {
  let feld: string | undefined
  let weg: ((feld: string, pfad: string) => void) | undefined

  return {
    merkeFeld: (neues) => {
      feld = neues
    },
    nimmFeld: () => {
      const gemerktes = feld
      feld = undefined

      return gemerktes
    },
    setzeBildweg: (neuer) => {
      weg = neuer
    },
    bildweg: () => weg,
  }
}

const Kontext = createContext<Ablagestelle | undefined>(undefined)

// Ablageraum stellt die Ablagestelle bereit.
export function Ablageraum({ children }: { children: ReactNode }) {
  // **Über `useState` und nicht `useRef`.** Der Wert entsteht einmal und
  // ändert sich nie; ein Verweis wäre dasselbe, aber der React-Übersetzer
  // liest ihn beim Zeichnen nicht gern — zu Recht, denn hier steht kein
  // Zwischenergebnis des Zeichnens.
  const [stelle] = useState(neueAblagestelle)

  return <Kontext.Provider value={stelle}>{children}</Kontext.Provider>
}

// useAblage gibt die Ablagestelle.
//
// **Ohne Raum eine eigene.** Ein Bauteil, das außerhalb geprüft wird, soll
// nicht daran scheitern — es legt dann eben ab, wo niemand hinsieht.
export function useAblage(): Ablagestelle {
  const [ersatz] = useState(neueAblagestelle)

  return useContext(Kontext) ?? ersatz
}
