// **Zwei Karten nebeneinander** — Werkstück U6 von Phase 5.1
// (#1417):
// die Karte mit dem Wert von vorher und die mit dem Wert von danach, beide
// über die Probevorschau gerastert, die der Editor beim Tippen nimmt. **Der
// Bildvergleich ergänzt den Text, er ersetzt ihn nicht** (Issue): die Tafel
// darüber bleibt stehen.
import { useCallback } from 'react'

import {
  api,
  auspacken,
  type VerlaufsschrittDTO,
  type VorschauDTO,
} from '../api'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'

interface VerlaufsvergleichEigenschaften {
  schritt: VerlaufsschrittDTO
  kartenname: string
  schliessen: () => void
}

// Figur zeigt eine Seite des Vergleichs — das Bild, wenn es da ist, und
// immer die Bildunterschrift, damit die leere Seite benannt bleibt.
function Figur({
  bild,
  bezeichnung,
}: {
  bild: VorschauDTO | null | undefined
  bezeichnung: string
}) {
  return (
    <figure className="min-w-0 flex-1">
      {bild != null && (
        <img
          src={bild.bild}
          alt={bezeichnung}
          width={bild.breite}
          height={bild.hoehe}
          className="block w-full max-w-full rounded-sm border border-kante"
        />
      )}
      <figcaption className="mt-1 text-sm text-leise">{bezeichnung}</figcaption>
    </figure>
  )
}

export function Verlaufsvergleich({
  schritt,
  kartenname,
  schliessen,
}: VerlaufsvergleichEigenschaften) {
  const worte = useTexte().seiten.karten.verlauf

  // **Nacheinander, nicht nebeneinander.** `App.Probevorschau` teilt sich den
  // Lauf mit der Kartenvorschau, und ein neuer Ruf bricht den vorigen ab
  // (§7) — zwei Bilder zugleich angefordert, kam eines nie an. Gefunden in
  // der Sichtprüfung von U6: die Seite „danach“ blieb leer, ohne Meldung,
  // denn ein abgebrochener Lauf ist kein Fehler. Ein Lauf für beide Seiten
  // stellt die Reihenfolge her und lässt sich als Ganzes räumen.
  const holen = useCallback(async () => {
    const eine = async (wert: string) => {
      const antwort = await api().Probevorschau(
        schritt.karteID,
        schritt.feld,
        wert,
        1,
      )

      return auspacken(antwort, () => antwort.vorschau)
    }

    const vorher = await eine(schritt.gabVorher ? schritt.vorher : '')
    const nachher = await eine(schritt.gabNachher ? schritt.nachher : '')

    return { vorher, nachher }
  }, [schritt])
  const bilder = useLaden(holen, [holen])

  return (
    <div className="mt-3">
      <Ladefeld laedt={bilder.laedt} fehler={bilder.fehler} leer={false}>
        <span />
      </Ladefeld>
      <div className="flex gap-4">
        <Figur
          bild={bilder.daten?.vorher}
          bezeichnung={einsetzen(worte.vergleichBisher, { karte: kartenname })}
        />
        <Figur
          bild={bilder.daten?.nachher}
          bezeichnung={einsetzen(worte.vergleichDanach, { karte: kartenname })}
        />
      </div>
      <button
        type="button"
        onClick={schliessen}
        className="mt-2 rounded-sm border border-kante px-2 py-0.5 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        {worte.vergleichSchliessen}
      </button>
    </div>
  )
}
