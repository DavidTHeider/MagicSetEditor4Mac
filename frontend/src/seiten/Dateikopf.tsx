// **Die Zeile über dem Editor der Werkstatt** — UXR-18
// (#1926).
//
// # Wofür es sie gibt
//
// Über dem Editor standen untereinander ein umrandeter Kasten zur eigenen
// Kopie mit drei Zeilen Begründung, eine Reihe von vier Handgriffen und die
// Warnung zum Geltungsbereich. Gemessen an der Hüllenprobe `fmt02b` begann
// der Editor bei 900 × 820 Punkten erst bei rund 400 (Kommentar am Issue).
// Das Issue fragt, ob diese Hinweise **zu einer Zeile zusammenfinden** — hier
// tun sie es: das Menü »Datei«, der Satz zum Geltungsbereich und, wo es eine
// gibt, die Kopie mit ihrem ersten Satz. Bei schmalem Fenster bricht die
// Zeile um, statt waagerecht zu rollen.
//
// **Der Konflikt steht nicht hier**, sondern im Fuß unter dem Editor bei
// Sichern — dort, wo er entsteht (FMT-02, `Vorlagendatei.tsx`).
import { useTexte } from '../sprache/kontext'
import { Dateiaktionen } from './Dateiaktionen'
import { Paketkopiehinweis } from './Paketkopiehinweis'

import type { Dateiaktionen as Handgriffe } from '../haken/useDateiaktionen'
import type { Paketkopie } from '../haken/usePaketkopie'

interface Kopfeigenschaften {
  aktionen: Handgriffe
  kopie: Paketkopie

  // paket und datei sind die gewählten; hauptdatei sagt, ob es die des
  // Pakets ist, und gesperrt, ob sie ungesichert ist.
  paket: string
  datei: string
  hauptdatei: boolean
  gesperrt: boolean
}

export function Dateikopf({
  aktionen,
  kopie,
  paket,
  datei,
  hauptdatei,
  gesperrt,
}: Kopfeigenschaften) {
  const texte = useTexte().seiten.vorlagen

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
      <Dateiaktionen
        aktionen={aktionen}
        datei={datei}
        hauptdatei={hauptdatei}
        gesperrt={gesperrt}
      />

      {/* **Die Warnung steht vor dem Feld.** Danach hätte sie niemand
          gelesen, der schon tippt. */}
      <p className="min-w-0 flex-1 basis-60 text-sm text-hinweis">
        {texte.warnung}
      </p>

      <Paketkopiehinweis kopie={kopie} paket={paket} />
    </div>
  )
}
