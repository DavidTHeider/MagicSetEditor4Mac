// **Der Hinweis zur eigenen Kopie über dem Editor der Werkstatt** —
// FMT-02.AK-04 und S10.NAT-01
// (#1691).
//
// Nach dem Artboard `Werkstatt`: **„Das Original bleibt"** steht über einer
// Kopie. Über einem fremden Paket steht, warum eine Kopie der bessere Weg
// ist, und der Knopf dazu — für die Pakete, deren Dateien sich öffnen lassen;
// eine mitgelieferte bietet der Baum an. Was geschieht, entscheidet
// `usePaketkopie`.
//
// **In der Zeile über dem Editor, nicht als Kasten** — UXR-18
// (#1926):
// sichtbar steht der erste Satz und der Knopf, die Begründung hinter dem
// Fragezeichen (`Erklaerungsabsatz`, derselbe Schnitt wie UX-15). Ein
// umrandeter Kasten mit drei Zeilen schob den Editor nach unten, und das
// Artboard zeichnet den Hinweis gar nicht über ihm.
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Paketkopie } from '../haken/usePaketkopie'

interface Hinweiseigenschaften {
  kopie: Paketkopie

  // paket ist der Pfad des gewählten Pakets.
  paket: string
}

export function Paketkopiehinweis({ kopie, paket }: Hinweiseigenschaften) {
  const eigene = useTexte().seiten.vorlagen

  if (kopie.original !== undefined) {
    return (
      <p role="status" className="text-sm text-leise">
        {einsetzen(eigene.kopieOriginal, { original: kopie.original })}
      </p>
    )
  }

  if (kopie.herkunftsart === undefined || kopie.herkunftsart === 'eigen') {
    return null
  }

  const fehler = kopie.fehler(paket)

  return (
    <section
      aria-label={eigene.kopieHinweis}
      className="flex flex-wrap items-center gap-2"
    >
      <Erklaerungsabsatz
        text={
          kopie.herkunftsart === 'mitgeliefert'
            ? eigene.kopieMitgeliefert
            : eigene.kopieFremd
        }
        klasse="text-sm"
      />
      <Aktionsknopf
        aktion="kopieren"
        klein
        angehalten={kopie.laeuft(paket)}
        aufKlick={() => void kopie.kopieren(paket)}
      >
        {eigene.kopieren}
      </Aktionsknopf>
      {fehler !== undefined && (
        <p role="alert" className="basis-full text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
