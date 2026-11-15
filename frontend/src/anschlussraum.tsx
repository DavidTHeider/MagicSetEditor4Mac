// **Der Weg, den ein Objekt zwischen zwei Flächen nimmt** — UX-03.AK-01
// (#1675).
//
// # Warum ein Kontext und keine Eigenschaft
//
// **Weil die zwei Enden fünf Bauteile auseinanderliegen.** Der
// Sammlungsinspector gibt ab, die Kartenwahl des Decks nimmt auf — dazwischen
// liegen `Sammlung`, `Anwendungsinhalt`, `Seiteninhalt`, `Decks` und
// `deckbaubereiche`, und keines davon hat mit einem Druck zu tun. Dieselbe
// Überlegung wie bei [Ablageraum](./ablage.tsx), und `Seiteninhalt.tsx` steht
// bei 360 Zeilen: zwei weitere Eigenschaften durch die Kette gingen an der
// Zeilengrenze aus §2 vorbei.
//
// # Regel 1 steht hier und nicht bei den Flächen
//
// **Ein Wechsel auf eine dritte Fläche räumt, was unterwegs war** (§7,
// Regel 1). Wer von der Sammlung zum Deck schickt und dann in die
// Einstellungen geht, hat es sich anders überlegt — käme der Druck beim
// nächsten Blick aufs Deck doch noch an, wäre das eine Handlung, die niemand
// angeordnet hat.
//
// **Beim Zeichnen und nicht in einem Effekt**, wie in `useSymbolziel`: ein
// Effekt zeichnete erst einmal mit dem alten Stand und räumte danach.
import { createContext, useContext, useState, type ReactNode } from 'react'

import { anschlussseite, type Anschluss } from './anschluss'

import type { Seitenschluessel } from './navigation'

// Anschlussstelle ist, was eine Fläche mit einem Anschluss tun kann.
export interface Anschlussstelle {
  // offen ist, was gerade unterwegs ist — `undefined`, wenn nichts reist.
  offen: Anschluss | undefined

  // uebergib legt das Objekt hin **und geht dorthin, wo es hingehört**.
  //
  // **Beides in einem Zug**, weil es ein Vorgang ist: „Ins Deck" heißt nicht
  // „merke dir das", sondern „nimm es mit". Wohin, sagt [anschlussseite] —
  // und damit steht das Ziel an einer Stelle statt an jedem Knopf.
  uebergib: (anschluss: Anschluss) => void

  // verbrauche räumt weg, was die Zielfläche aufgenommen hat.
  //
  // **Aus einem Ereignis und nicht beim Zeichnen.** Die Fläche füllt ihr
  // Formular, sobald der Anschluss da ist; *verbraucht* ist er erst, wenn
  // jemand die Karte wirklich gelegt hat.
  verbrauche: () => void
}

const Kontext = createContext<Anschlussstelle | undefined>(undefined)

// Anschlussraum hält, was zwischen zwei Flächen unterwegs ist.
export function Anschlussraum({
  aktuell,
  wechseln,
  children,
}: {
  aktuell: Seitenschluessel
  wechseln: (seite: Seitenschluessel) => void
  children: ReactNode
}) {
  const [offen, setzeOffen] = useState<Anschluss | undefined>(undefined)
  const [vorige, setzeVorige] = useState(aktuell)

  // **Der Vergleich steht neben dem Zustand** — die Form, die React für
  // „Zustand an eine geänderte Eigenschaft anpassen" vorsieht.
  if (vorige !== aktuell) {
    setzeVorige(aktuell)

    if (offen !== undefined && anschlussseite[offen.art] !== aktuell) {
      setzeOffen(undefined)
    }
  }

  const stelle: Anschlussstelle = {
    offen,

    uebergib: (neuer) => {
      setzeOffen(neuer)
      wechseln(anschlussseite[neuer.art])
    },

    verbrauche: () => {
      setzeOffen(undefined)
    },
  }

  return <Kontext.Provider value={stelle}>{children}</Kontext.Provider>
}

// taubeStelle nimmt entgegen und tut nichts.
const taubeStelle: Anschlussstelle = {
  offen: undefined,
  uebergib: () => undefined,
  verbrauche: () => undefined,
}

// useAnschluss gibt die Anschlussstelle.
//
// **Ohne Raum eine taube.** Ein Bauteil, das außerhalb geprüft wird, soll
// nicht daran scheitern — es übergibt dann eben dorthin, wo niemand hinsieht.
// Dieselbe Überlegung wie bei `useAblage`.
export function useAnschluss(): Anschlussstelle {
  return useContext(Kontext) ?? taubeStelle
}

// useAufgenommen ruft `aufnehmen` **genau einmal** je mitgereistem Anschluss.
//
// # Warum beim Zeichnen und nicht in einem Effekt
//
// **Weil ein Effekt eine zweite Anzeige kostet**: das Formular stünde erst
// leer da und füllte sich danach. React nennt das eine kaskadierende Anzeige,
// und der Prüfer streicht es an. Dies hier ist die Form daneben — der vorige
// Wert steht als Zustand daneben, und der Vergleich entscheidet.
//
// **`aufnehmen` setzt dabei nur den eigenen Zustand der rufenden Fläche.**
// Einen fremden zu setzen wäre das, was der Prüfer zu Recht abweist.
//
// # Warum es an der Art hängt
//
// **Gebunden an die Art und nicht an die Union.** Die Fläche hat vorher
// verengt — `offen?.art === 'deck' ? offen : undefined` —, und ohne diese
// Schranke käme der Anschluss als Union zurück: der Aufrufer läse `printing`
// auf einem Verbund, der eine Deckkennung trägt, und der Übersetzer wiese es
// zu Recht ab.
export function useAufgenommen<A extends Anschluss>(
  mitgereist: A | undefined,
  aufnehmen: (anschluss: A) => void,
): void {
  const [aufgenommen, setzeAufgenommen] = useState<A | undefined>(undefined)

  if (mitgereist !== undefined && mitgereist !== aufgenommen) {
    setzeAufgenommen(mitgereist)
    aufnehmen(mitgereist)
  }
}
