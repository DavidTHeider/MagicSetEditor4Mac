// **Wer wo als was arbeitet — aus einem Kontext, mit Namen dazu** —
// Werkstück U4 von Phase 5.1
// (#1415).
//
// # Warum `Arbeitskontext` und nicht `profil.nutzerID`
//
// **Weil der Eigner nicht der ist, der arbeitet**, sobald jemand in ein
// fremdes Profil gewechselt ist ([0074](../../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md)).
// Der Kontext kommt aus **einem** Ruf ([0078](../../../docs/entscheidungen/0078-der-wirksame-akteur.md))
// — Nutzer, Profil, Eigner, Rolle, ob nur gelesen wird, und eine Revision.
// Die Namen dazu holt dieser Haken aus der Nutzerliste und den Bildern; sie
// kommen im selben Zug, und die Anwendung baut die Leiste beim Profilwechsel
// neu auf (`key={profilstand}`), sodass keine alte Liste zu einem neuen
// Kontext gehört (§7, Regel 1 und 2).
//
// **Wen die Liste nicht führt, den nennt die Leiste bei der Kennung** — sie
// behauptet keinen Menschen, den sie nicht kennt.
import { useCallback } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { ArbeitskontextAntwort } from '../apikontext'
import type { NutzerDTO } from '../apinutzer'

// Arbeitskontextstand ist, was die Leiste zeigt.
export interface Arbeitskontextstand {
  kontext: ArbeitskontextAntwort | undefined

  nutzername: string
  profilname: string
  eignername: string

  bild: string
  badge: string

  laedt: boolean
  fehler: string | undefined
}

// nameVon nennt den Menschen mit dieser Kennung — oder die Kennung selbst.
function nameVon(nutzer: readonly NutzerDTO[], kennung: string): string {
  return nutzer.find((einer) => einer.id === kennung)?.name ?? kennung
}

// profilnameVon nennt das Profil mit dieser Kennung — oder die Kennung selbst.
function profilnameVon(nutzer: readonly NutzerDTO[], kennung: string): string {
  for (const einer of nutzer) {
    const eines = einer.profile.find((profil) => profil.id === kennung)
    if (eines !== undefined) {
      return eines.name
    }
  }

  return kennung
}

// **Ohne eigenes Nachladen**: die Anwendung baut die Leiste beim Wechsel neu
// auf (`key={profilstand}`), und ein zweiter Weg daneben wäre ein zweiter
// Stand (§7, Regel 2).
export function useArbeitskontext(): Arbeitskontextstand {
  const holen = useCallback(async () => {
    const [kontext, liste, bilder] = await Promise.all([
      api().Arbeitskontext(),
      api().Nutzer(),
      api().Nutzerbilder(),
    ])

    const gelesen = auspacken(kontext, () => kontext)
    const nutzer = auspacken(liste, () => liste.nutzer)
    const abbild = auspacken(bilder, () => bilder)

    return {
      kontext: gelesen,
      nutzername: nameVon(nutzer, gelesen.nutzerID),
      profilname: profilnameVon(nutzer, gelesen.profilID),
      eignername: nameVon(nutzer, gelesen.eignerID),
      bild: abbild.bilder[gelesen.nutzerID] ?? '',
      badge: abbild.badges[gelesen.profilID] ?? '',
    }
  }, [])

  const { daten, laedt, fehler } = useLaden(holen, [], true)

  return {
    kontext: daten?.kontext,
    nutzername: daten?.nutzername ?? '',
    profilname: daten?.profilname ?? '',
    eignername: daten?.eignername ?? '',
    bild: daten?.bild ?? '',
    badge: daten?.badge ?? '',
    laedt,
    fehler,
  }
}
