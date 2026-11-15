// **Die zwei Leerzustände des Paketkatalogs** — G02
// (#2249):
// ein leerer Katalog bietet das Einbinden an, eine Suche ohne Treffer führt
// zurück. Aus `Paketliste.tsx` herausgelöst, als G02 es über 400 Zeilen trug
// (§2).
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import { useTexte } from '../sprache/kontext'

// Katalogleer steht da, wenn kein Paket eingebunden ist — **mit der Handlung
// daneben**: wer auf eine leere Liste schaut, schaut nicht zum Knopf oben.
export function Katalogleer({
  laeuft,
  einbinden,
}: {
  laeuft: boolean
  einbinden: () => void
}) {
  const eigene = useTexte().seiten.pakete

  return (
    <Leerzustand
      titel={eigene.katalogLeerTitel}
      erklaerung={eigene.katalogLeer}
    >
      <Knopf angehalten={laeuft} aufKlick={einbinden}>
        {eigene.plan.einbinden}
      </Knopf>
    </Leerzustand>
  )
}

// Katalogohnetreffer steht da, wenn Suche und Filter nichts übrig lassen.
export function Katalogohnetreffer({
  zuruecksetzen,
}: {
  zuruecksetzen: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.pakete

  return (
    <Leerzustand
      titel={eigene.katalogNichtsTitel}
      erklaerung={eigene.katalogNichtsGefunden}
    >
      <Knopf aufKlick={zuruecksetzen}>
        {texte.bausteine.listenansicht.filterZuruecksetzen}
      </Knopf>
    </Leerzustand>
  )
}
