// **Eine Seite, deren Bereich abgeschaltet ist** — OPS-10
// (#1908).
//
// **Ein leerer Zustand und keine Weiterleitung**: wer über ⌘4 oder den
// Verlauf hierher kommt, erfährt, warum nichts dasteht, und hat den Weg
// zurück in der Hand (§6: leere Zustände erklären den nächsten Schritt). Eine
// stille Weiterleitung ließe ihn rätseln, warum die Taste nichts tut.
import { useAnschluss } from '../anschlussraum'
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import { beschriftung, type Seitenschluessel } from '../navigation'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Abgeschaltet({ seite }: { seite: Seitenschluessel }) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.schalter.abgeschaltet
  const anschluss = useAnschluss()

  return (
    <Leerzustand
      titel={einsetzen(eigene.titel, { seite: beschriftung(texte, seite) })}
      erklaerung={eigene.erklaerung}
    >
      <Knopf
        aufKlick={() => {
          anschluss.uebergib({
            art: 'einstellungsgruppe',
            name: texte.seiten.einstellungen.gruppen.bereiche,
            gruppe: 'bereiche',
          })
        }}
      >
        {eigene.oeffnen}
      </Knopf>
    </Leerzustand>
  )
}
