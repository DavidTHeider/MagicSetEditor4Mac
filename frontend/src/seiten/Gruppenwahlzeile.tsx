// **Die Zeile über den Balken: was gewählt ist und wie es weitergeht** —
// UXR-F6 (#2013),
// das Feld „Von hier weiter" des Artboards *Werkzeuge*.
//
// # Zwei Zustände
//
// **Ohne Wahl ein Satz**, der sagt, was ein Klick tut — ein Balken, der wie
// Text aussieht, lädt nicht zum Klicken ein (§6: leere Zustände erklären den
// nächsten Schritt). **Mit Wahl der Balken in Worten** und zwei Handlungen:
// die Karten zeigen — sie führt auf die Kartenliste, die Schublade geht zu —
// und die Wahl aufheben.
//
// **Die Wahl steht schon auf der Kartenliste, bevor jemand dorthin geht**:
// sie reist mit (siehe `statistikfilter.tsx`). Der Knopf ist der Weg, nicht
// das Absenden.
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useStatistikfilter } from '../statistikfilter'
import { ohneAuszeichnung } from './ohneauszeichnung'

interface ZeilenEigenschaften {
  // gewaehlt ist der gewählte Balken samt seiner Kartenzahl — oder keiner.
  gewaehlt: { name: string; karten: number } | undefined
}

export function Gruppenwahlzeile({ gewaehlt }: ZeilenEigenschaften) {
  const eigene = useTexte().seiten.statistik
  const filter = useStatistikfilter()

  if (gewaehlt === undefined) {
    return (
      <Erklaerungsabsatz
        text={eigene.gruppeHinweis}
        klasse="mt-2 max-w-prose text-sm text-leise"
      />
    )
  }

  return (
    <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
      <p role="status" className="text-sm">
        {einsetzen(eigene.gruppeGewaehlt, {
          gruppe: ohneAuszeichnung(gewaehlt.name),
          karten: String(gewaehlt.karten),
        })}
      </p>

      <Knopf gewicht="haupt" klein aufKlick={filter.zurKartenliste}>
        {eigene.gruppeZeigen}
      </Knopf>

      <Knopf gewicht="leise" klein aufKlick={filter.aufheben}>
        {eigene.gruppeAufheben}
      </Knopf>
    </div>
  )
}
