// **Was die Mengenentscheidung tut** — DOM-03, S04.NAT-01 bis 03
// (#1698).
// Wie sie aussieht, steht in `bausteine/Mengenentscheidung.tsx` (§2).
import {
  exemplarschluessel,
  type Dublette,
  type Sammlungsstand,
} from './useSammlung'

export interface Mengenschritte {
  laeuft: boolean
  erhoehen: () => void
  zurZeile: () => void
  abbrechen: () => void
}

type Sammlungsteil = Pick<
  Sammlungsstand,
  'aendern' | 'waehlen' | 'dubletteSchliessen' | 'vorgang'
>

// mengenschritte gibt die drei Antworten auf eine Dublette.
//
// **Erhöhen geht gegen die gelesene Fassung** (§7, Regel 2): hat jemand die
// Zeile inzwischen geändert, sagt das Ändern es, und die Menge bleibt.
// **Abbrechen ändert nichts** — die Eingabe im Formular bleibt stehen
// (S04.NAT-03).
export function mengenschritte(
  dublette: Dublette,
  sammlung: Sammlungsteil,
  aufZusammengefuehrt: () => void,
): Mengenschritte {
  const { neu, vorhanden } = dublette

  return {
    laeuft: sammlung.vorgang.laeuft(exemplarschluessel(vorhanden.id)),
    erhoehen: () => {
      const summe = vorhanden.menge + neu.menge

      void sammlung
        .aendern({ ...vorhanden, menge: summe })
        .then((geglueckt) => {
          sammlung.dubletteSchliessen()

          if (geglueckt) {
            aufZusammengefuehrt()

            // **Die Fassung, die der Bestand jetzt führt**: er zählt sie
            // beim Ändern um eins hoch, und nur, wenn die gelesene stimmte
            // (`storage.Sammlungsspeicher.Aendern`).
            sammlung.waehlen(vorhanden.id, {
              ...vorhanden,
              menge: summe,
              fassung: vorhanden.fassung + 1,
            })
          }
        })
    },
    zurZeile: () => {
      sammlung.waehlen(vorhanden.id, vorhanden)
      sammlung.dubletteSchliessen()
    },
    abbrechen: sammlung.dubletteSchliessen,
  }
}

// zurZeileSchliesst lässt „Zur vorhandenen Zeile" die Schublade schließen —
// **sonst wählte der Knopf eine Zeile im Inspector, den die Schublade
// verdeckt**. Erhöhen und Abbrechen lassen sie stehen: wer zehn Karten
// aufnimmt, nimmt danach die elfte auf.
export function zurZeileSchliesst(
  schritte: ReturnType<typeof mengenschritte>,
  schliessen: () => void,
): ReturnType<typeof mengenschritte> {
  return {
    ...schritte,
    zurZeile: () => {
      schritte.zurZeile()
      schliessen()
    },
  }
}
