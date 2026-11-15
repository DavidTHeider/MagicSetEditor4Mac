// **Das eine Bild der Statistik** — UXR-08
// (#1916).
//
// Unter der Wahl „Worüber" steht genau eines: die Belegung, die Verteilung
// eines Felds oder die Auszählung nach einer Dimension. Alle drei haben
// dieselbe Gestalt — eine Überschrift, die sagt, wonach gezählt ist, und
// Balken —, und deshalb steht sie hier und nicht dreimal.
import type { ReactNode } from 'react'

import type { ZahlDTO } from '../api'
import { Balken } from './Balken'
import { Faltliste } from '../bausteine/Faltliste'
import { useTexte } from '../sprache/kontext'

// vorn ist, wie viele Balken zu Anfang dastehen — die größten, denn Belegung
// und Verteilung kommen absteigend; eine Auszählung in der Reihenfolge, die
// das Spielpaket erklärt.
//
// **Zehn, weil sie mit Kennzahlen und Wahl in eine Schublade von 800 Punkten
// Höhe passen** (gut 36 Punkte je Balken). Gemessen an `White.mse-set`
// (3.761 Karten, aus der Setdatei gezählt am 2026-10-02) trägt das Feld
// `illustrator` 586 verschiedene Werte und `name` 3.756 — untereinander war
// das eine Rolltiefe, die R4 aus `docs/UIREVIEW.md` um das Vielfache bricht.
const vorn = 10

interface BildEigenschaften {
  // ueberschrift sagt, wonach gezählt ist — und benennt die Region.
  ueberschrift: string
  children: ReactNode
}

export function Statistikbild({ ueberschrift, children }: BildEigenschaften) {
  return (
    <section aria-label={ueberschrift} className="mt-6">
      <h3 className="text-lg font-medium">{ueberschrift}</h3>
      {children}
    </section>
  )
}

// Balkenwahl macht die Balken einer Liste wählbar — UXR-F6
// (#2013).
export interface Balkenwahl<T extends ZahlDTO> {
  // gewaehlt ist der Name des gewählten Balkens — oder keiner.
  gewaehlt: string | undefined
  waehlen: (eintrag: T) => void
}

interface ListeEigenschaften<T extends ZahlDTO> {
  eintraege: readonly T[]

  // ganzes ist, worauf sich jeder Anteil bezieht.
  ganzes: number

  // feldwert sagt, ob die Namen Feldwerte sind (mit Auszeichnung) oder
  // Feldnamen — der Grund steht bei `Balken`.
  feldwert?: boolean

  // wahl macht die Balken wählbar. **Ohne sie sind es Zahlen zum Lesen**: die
  // Belegung zählt Karten je Feld, und eine Karte steht in vielen davon — das
  // sind keine Gruppen, in die man hineinwählt.
  wahl?: Balkenwahl<T>
}

export function Balkenliste<T extends ZahlDTO>({
  eintraege,
  ganzes,
  feldwert = false,
  wahl,
}: ListeEigenschaften<T>) {
  const eigene = useTexte().seiten.statistik
  const anteil = eigene.anteil

  return (
    <Faltliste
      eintraege={eintraege}
      vorn={vorn}
      schritt={100}
      klasse="mt-3 max-w-xl"
      zeile={(eintrag) => (
        <Balken
          key={eintrag.name}
          name={eintrag.name}
          karten={eintrag.karten}
          ganzes={ganzes}
          anteil={anteil}
          feldwert={feldwert}
          waehlbar={
            wahl && {
              gewaehlt: wahl.gewaehlt === eintrag.name,
              waehlen: () => {
                wahl.waehlen(eintrag)
              },
              leerTitel: eigene.gruppeLeer,
            }
          }
        />
      )}
    />
  )
}
