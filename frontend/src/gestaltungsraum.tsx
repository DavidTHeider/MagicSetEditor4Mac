// **Farbwelt, Fensterbild, Schrift und eigene Merkmale gelten überall** —
// UX-09 (#1681).
//
// # Warum ein Raum
//
// **Dieselbe Begründung wie beim `Nutzereinstellungsraum`**: das
// Wurzelelement bekommt die Attribute, und die Einstellungsseite braucht
// denselben Stand — das Fensterbild, das sie in der Vorschau zeigt, und die
// Befunde, die sagen, warum eigene Farben nicht gelten.
//
// # Warum vor dem Zeichnen
//
// **Die Sperre aus AK-03 misst, was gesetzt ist** — unlesbare eigene Farben
// dürfen nie auf dem Bildschirm stehen, auch nicht einen Augenblick lang.
// Deshalb `useLayoutEffect`: setzen, messen und zurücknehmen geschieht,
// bevor das Fenster zeichnet. **Und Erscheinungsbild und Kontrast werden
// hier noch einmal gesetzt**, obwohl `App` und der Nutzereinstellungsraum es
// auch tun: deren Effekte laufen **nach** diesem (React ruft die Effekte
// eines Kindes vor denen seiner Eltern), und gemessen werden muss am Bild,
// das gleich gilt — nicht an dem von vorhin.
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useState,
  type ReactNode,
} from 'react'

import { api, auspacken } from './api'
import { ereignisDarstellungZuruecksetzen, horche } from './fenster'
import { meldungAus } from './meldung'
import { useTexte } from './sprache/kontext'

import {
  schriftbereichGilt,
  setzeErscheinungsbild,
  setzeKontrast,
  setzeSchriftbereich,
} from './erscheinungsbild'
import {
  schleierAnteil,
  setzeEigeneMerkmale,
  setzeFarbwelt,
  setzeHintergrund,
} from './gestaltung'
import { useFensterbild } from './haken/useFensterbild'
import { useMantineschema } from './mantine'
import { useMedienfrage } from './haken/useMedienfrage'
import { useNutzereinstellungsstand } from './nutzereinstellungsraum'

import type { Eigenbildstand } from './haken/useEigenesBild'
import type { Lesbarkeitsbefund } from './kontrast'

export interface Gestaltungsstand {
  fensterbild: Eigenbildstand

  // befunde sind die Paare, an denen die eigenen Merkmale scheiterten —
  // leer heißt: sie gelten, oder es gibt keine.
  befunde: Lesbarkeitsbefund[]

  // schmuckNurOben sagt, dass die ganze Oberfläche gewünscht war und eine
  // Schmuckschrift gewählt ist — sie bleibt an den Überschriften.
  schmuckNurOben: boolean

  // ruecksetzfehler ist, woran das Zurücksetzen aus der Menüleiste
  // scheiterte — das Meldungsband zeigt es und die Einstellungsseite auch.
  ruecksetzfehler: string | undefined

  // ruecksetzfehlerRaeumen nimmt ihn weg — das Kreuz am Meldungsband
  // (Befund G5 des Zweitblicks über Phase 6.3).
  ruecksetzfehlerRaeumen: () => void

  // kontrastHoch sagt, dass hoher Kontrast gilt — dann ruhen eigene Farben,
  // und die Seite sagt das statt „Angewandt." (Zweitblick über Phase 6.3,
  // G1).
  kontrastHoch: boolean
}

const Kontext = createContext<Gestaltungsstand | undefined>(undefined)

export function Gestaltungsraum({
  einstellungen,
  neuLesen,
  children,
}: {
  einstellungen: { wert: (schluessel: string) => string }

  // neuLesen lässt die Anwendung ihre Einstellungen neu lesen — nach dem
  // Zurücksetzen aus der Menüleiste.
  neuLesen: () => void
  children: ReactNode
}) {
  const texte = useTexte()
  const [ruecksetzfehler, setzeRuecksetzfehler] = useState<string>()
  const ruecksetzfehlerRaeumen = useCallback(() => {
    setzeRuecksetzfehler(undefined)
  }, [])
  const nutzer = useNutzereinstellungsstand()
  const kontrastwahl = nutzer.wert('barrierefreiheit.kontrast')
  const systemHoch = useMedienfrage('(prefers-contrast: more)')
  const systemDunkel = useMedienfrage('(prefers-color-scheme: dark)')
  const kontrastHoch = kontrastwahl === 'hoch' || systemHoch

  const fensterbild = useFensterbild(nutzer.nutzerID)
  const [befunde, setzeBefunde] = useState<Lesbarkeitsbefund[]>([])

  const erscheinungsbild = einstellungen.wert('erscheinungsbild')
  const farbwelt = einstellungen.wert('farbwelt')
  const hintergrund = einstellungen.wert('hintergrund')
  const schriftbereich = einstellungen.wert('schriftbereich')
  const schriftart = einstellungen.wert('schriftart')
  const gestaltung = einstellungen.wert('gestaltung')
  const schmuckNurOben = !schriftbereichGilt(schriftbereich, schriftart)
  const adresse =
    fensterbild.adresse === ''
      ? ''
      : `${fensterbild.adresse}?stand=${String(fensterbild.stand)}`

  const dunkel =
    erscheinungsbild === 'dunkel' ||
    (erscheinungsbild !== 'hell' && systemDunkel)

  // **Mantine folgt dem Erscheinungsbild** (A03, #2217) — `mantine.ts`.
  useMantineschema(dunkel)

  useLayoutEffect(() => {
    const wurzel = document.documentElement

    setzeErscheinungsbild(wurzel, erscheinungsbild)
    setzeKontrast(wurzel, kontrastwahl)
    setzeFarbwelt(wurzel, farbwelt, kontrastHoch)
    setzeHintergrund(wurzel, hintergrund, adresse, kontrastHoch)
    setzeSchriftbereich(wurzel, schriftbereich, schriftart)
    // **Gemessen wird am gesetzten Bild**, und das gibt es erst hier: die
    // Befunde lassen sich beim Zeichnen nicht ausrechnen, weil die Farben
    // aus den Stilblättern kommen. Das ist der Fall, für den React
    // `useLayoutEffect` vorsieht — messen, bevor das Fenster zeichnet.
    //
    // **Gemessen wird im Erscheinungsbild, das gilt, und unter dem Schleier,
    // wenn ein Fensterbild darunter liegt** (Zweitblick über Phase 6.3, S4).
    const lage = {
      dunkel,
      ...(wurzel.style.getPropertyValue('--mse-fensterbild') === ''
        ? {}
        : { schleier: schleierAnteil(wurzel) }),
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- Messung nach dem Setzen, siehe oben
    setzeBefunde(setzeEigeneMerkmale(wurzel, gestaltung, kontrastHoch, lage))
  }, [
    erscheinungsbild,
    kontrastwahl,
    kontrastHoch,
    dunkel,
    farbwelt,
    hintergrund,
    adresse,
    schriftbereich,
    schriftart,
    gestaltung,
  ])

  // **Darstellung zurücksetzen aus der Menüleiste** (AK-03): die
  // Layoutrücksetzung, ohne Rückfrage — sie nimmt nichts weg außer dem
  // Aussehen (`domain.Layouteinstellungen`), und eine Rückfrage wäre gerade
  // dann nicht zu lesen, wenn es sie braucht.
  useEffect(
    () =>
      horche(ereignisDarstellungZuruecksetzen, () => {
        void (async () => {
          try {
            const vorschau = await api().Ruecksetzungsvorschau('layout', '', '')
            const revision = auspacken(vorschau, () => vorschau.revision)
            const bilanz = await api().Zuruecksetzen(
              'layout',
              '',
              '',
              '',
              '',
              revision,
            )
            auspacken(bilanz, () => bilanz.art)
            setzeRuecksetzfehler(undefined)
            neuLesen()
          } catch (grund) {
            setzeRuecksetzfehler(meldungAus(grund, texte))
          }
        })()
      }),
    [neuLesen, texte],
  )

  return (
    <Kontext.Provider
      value={{
        fensterbild,
        befunde,
        schmuckNurOben,
        ruecksetzfehler,
        ruecksetzfehlerRaeumen,
        kontrastHoch,
      }}
    >
      {children}
    </Kontext.Provider>
  )
}

// useGestaltung gibt den Stand aus dem Raum — **ohne Raum `undefined`**:
// die Einstellungsseite zeigt dann weder Vorschau noch Befunde, statt mit
// einem erfundenen Stand zu arbeiten.
export function useGestaltung(): Gestaltungsstand | undefined {
  return useContext(Kontext)
}
