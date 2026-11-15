// **Ob in diesem Profil geschrieben werden darf** — G10
// (#2276).
//
// B03 hat jede Seite mit einem Leser geöffnet: auf sechzehn standen
// Schreibhandlungen bedienbar da, und ein Klick scheiterte erst am Tor der
// Bindung. **Die Regel kommt jetzt aus einer Stelle**: die Anwendung fragt den
// Arbeitskontext einmal je Profilstand, legt die Antwort in diesen Raum, und
// jeder `Knopf`, der `schreibt`, liest sie — statt sechzehn Seiten, die je für
// sich fragen.
//
// **Welcher Knopf schreibt, sagt sein Aufrufer** und nicht der
// Aktionskatalog: was ein Leser darf, hängt am Träger der Sache
// (`docs/AUTORISIERUNG.md`) — ein Paket gehört dem Rechner, ein Deck dem
// Profil.
//
// **Ein Fehlschlag sperrt nichts**: wer trotzdem schreibt, bekommt die Absage
// vom Tor — sie ist die verlässliche Auskunft, und dieser Raum nur die frühere
// (derselbe Grund wie beim Streifen, `Nurlesehinweis`).
import {
  createContext,
  use,
  useEffect,
  useId,
  useState,
  type ReactNode,
} from 'react'

import { api } from '../fassade'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLauf } from './useLauf'

const NurLesend = createContext(false)

export function Schreibrechtraum({
  nurLesend,
  children,
}: {
  nurLesend: boolean
  children: ReactNode
}) {
  return <NurLesend value={nurLesend}>{children}</NurLesend>
}

// useNurLesend sagt, ob hier nur gelesen werden darf.
export function useNurLesend(): boolean {
  return use(NurLesend)
}

// Schreibsperre ist, was ein Knopf braucht, der schreibt: ob er gesperrt
// ist, die Eigenschaften dafür, und der verborgene Satz, der den Grund nennt.
export interface Schreibsperre {
  gesperrt: boolean

  // knopf gehört an das `<button>`: `aria-disabled` und die Beschreibung
  // durch den Satz — **fokussierbar bleibt er**, denn ein Vorleseprogramm
  // soll lesen können, warum er nicht geht; `disabled` nähme ihn aus der
  // Reihe.
  knopf: {
    'aria-disabled': true | undefined
    'aria-describedby': string | undefined
  }

  // titel ist derselbe Satz für den Zeiger — wo der Knopf keinen eigenen
  // Tooltip hat; zwei übereinander sagten dasselbe zweimal.
  titel: string | undefined

  // satz gehört **in** den Knopf: verborgen zählt er nicht zum Namen, wohl
  // aber zur Beschreibung, die ihn nennt.
  satz: ReactNode
}

// useSchreibsperre sagt einem Knopf, der `schreibt`, ob er hier gesperrt ist
// — für die Stellen, die keinen `Knopf` nehmen können, etwa den Absendeknopf
// eines Formulars. `Knopf` selbst nimmt denselben Haken.
export function useSchreibsperre(schreibt: boolean): Schreibsperre {
  const texte = useTexte()
  const gesperrt = useNurLesend() && schreibt
  const grund = useId()

  return {
    gesperrt,
    knopf: {
      'aria-disabled': gesperrt ? true : undefined,
      'aria-describedby': gesperrt ? grund : undefined,
    },
    titel: gesperrt ? texte.bausteine.nurlesen.knopf : undefined,
    satz: gesperrt ? (
      <span id={grund} hidden>
        {texte.bausteine.nurlesen.knopf}
      </span>
    ) : null,
  }
}

// Lesegrund ist, was eine **Eingabe** braucht, die für einen Leser gesperrt
// ist — dasselbe wie [Schreibsperre] für einen Knopf, nur ohne die Griffe des
// Knopfs: ob sie gesperrt ist, ihre Beschreibung samt dem Satz, und der
// verborgene Satz selbst.
//
// **Warum nicht der Streifen oben genügt** (H01,
// #2255):
// ein Vorleser sagt am Feld »schreibgeschützt« oder »abgeblendet«, aber nicht,
// was über der Seite steht. Der Kernweg je Rolle am nativen Fenster fand das
// Kartenfeld eines Lesers ohne Grund, während jeder gesperrte Knopf ihn nennt.
export interface Lesegrund {
  gesperrt: boolean

  // beschreibe hängt die Kennung des Satzes an eine eigene Beschreibung — und
  // gibt sie unverändert zurück, wenn nichts gesperrt ist.
  beschreibe: (eigene?: string) => string | undefined

  // satz gehört **neben** die Eingabe, einmal je Baustein: verborgen, aber
  // durch `aria-describedby` erreichbar.
  satz: ReactNode
}

// useLesegrund gibt einer Eingabe, die für einen Leser gesperrt ist, den
// Grund mit — denselben Satz wie [useSchreibsperre] einem Knopf.
export function useLesegrund(): Lesegrund {
  const sperre = useSchreibsperre(true)
  const grund = sperre.knopf['aria-describedby']

  return {
    gesperrt: sperre.gesperrt,
    beschreibe: (eigene) =>
      [eigene, grund]
        .filter((teil) => teil !== undefined && teil !== '')
        .join(' ') || undefined,
    satz: sperre.satz,
  }
}

// useNurLesendLaden fragt den Arbeitskontext — einmal je `stand`, also je
// Profilwechsel.
//
// **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2): mit
// einer geteilten Referenz, die jeder Effekt wieder auf „gilt" setzt, kam die
// späte Antwort des verlassenen Profils doch noch an und sperrte das neue.
export function useNurLesendLaden(stand: number): boolean {
  const [nurLesend, setzeNurLesend] = useState(false)
  const lauf = useLauf()

  useEffect(() => {
    const nummer = lauf.naechster()

    void (async () => {
      // **Der Ruf steht für sich**, und die Kette hängt an seiner Zusage: der
      // Wächter über `docs/SCHREIBWEGE.md` liest die Fassade und den Namen
      // dahinter in einem Stück.
      const gefragt = api().Arbeitskontext()

      const wert = await gefragt
        .then((antwort) => auspacken(antwort, () => antwort.nurLesend))
        .catch(() => false)

      if (lauf.istAktuell(nummer)) {
        setzeNurLesend(wert)
      }
    })()

    return () => {
      lauf.naechster()
    }
  }, [lauf, stand])

  return nurLesend
}

// Schreibrecht füllt den Raum — innerhalb der offenen Sitzung, und je
// Profilstand neu gefragt, **ohne** den Baum darunter neu aufzubauen.
export function Schreibrecht({
  stand,
  children,
}: {
  stand: number
  children: ReactNode
}) {
  return (
    <Schreibrechtraum nurLesend={useNurLesendLaden(stand)}>
      {children}
    </Schreibrechtraum>
  )
}
