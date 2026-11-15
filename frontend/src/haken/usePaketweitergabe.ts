// **Ein Paket als `.mse-installer` weitergeben** — FMT-02.AK-04
// (#1691).
//
// Der Dialog von macOS fragt nach dem Ort; hier steht, was danach am Paket
// zu lesen ist.
//
// # Die vier Regeln aus §7
//
// **Regel 1 und 2**: eine Meldung gehört zum Paket, an dem sie entstand, und
// steht nur dort — wer inzwischen ein anderes im Inspector hat, sieht sie
// nicht, und eine späte Antwort trägt sich bei ihrem Paket ein. **Regel 3**:
// eine Weitergabe ändert nichts am Bestand; es gibt nichts nachzuladen.
// **Regel 4**: ein Vorgang je Paket.
import { useState } from 'react'

import { api } from '../fassade'
import { istAbgebrochen, meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useVorgang } from './useVorgang'

import type { PaketDTO } from '../pakete'

export interface Weitergabemeldung {
  text: string
  fehler: boolean
}

export interface Paketweitergabezustand {
  weitergeben: (paket: PaketDTO) => Promise<void>
  laeuft: (paket: PaketDTO) => boolean
  meldung: (paket: PaketDTO) => Weitergabemeldung | undefined
}

// dateiname ist der Name, unter dem die Anwendung ein Paket kennt — mit
// Endung, wie in `depends on:`.
function dateiname(paket: PaketDTO): string {
  return `${paket.name}.${paket.art}`
}

export function usePaketweitergabe(): Paketweitergabezustand {
  const texte = useTexte()
  const vorgang = useVorgang()

  const [meldungen, setzeMeldungen] = useState<
    Readonly<Record<string, Weitergabemeldung | undefined>>
  >({})

  const melde = (schluessel: string, meldung?: Weitergabemeldung) => {
    setzeMeldungen((alt) => ({ ...alt, [schluessel]: meldung }))
  }

  return {
    weitergeben: async (paket) => {
      const schluessel = dateiname(paket)

      await vorgang.starte(schluessel, async () => {
        melde(schluessel)

        try {
          const antwort = await api().PaketWeitergeben(schluessel)
          const gelesen = auspacken(antwort, () => antwort)

          if (!gelesen.abgebrochen) {
            melde(schluessel, {
              text: einsetzen(texte.seiten.pakete.plan.weitergegeben, {
                datei: gelesen.datei,
              }),
              fehler: false,
            })
          }
        } catch (ursache: unknown) {
          // **Ein Anhalten ist keine Fehlermeldung** (FMT-03): der Installer
          // entsteht erst am Ende, es liegt keiner da.
          if (!istAbgebrochen(ursache)) {
            melde(schluessel, {
              text: meldungAus(ursache, texte),
              fehler: true,
            })
          }
        }
      })
    },
    laeuft: (paket) => vorgang.laeuft(dateiname(paket)),
    meldung: (paket) => meldungen[dateiname(paket)],
  }
}
