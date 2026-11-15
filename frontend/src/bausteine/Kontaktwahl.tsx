// **Wen es betrifft — gewählt und nicht getippt** — UX-08.AK-04
// (#1680):
// *„Normalnutzende müssen keine Netzwerk- oder Transaktionskennungen
// eingeben, soweit diese technisch ableitbar sind."*
//
// # Was vorher war
//
// **Ein Feld für eine Peerkennung** — ein Schlüssel von über fünfzig
// Zeichen, abgelesen von einem anderen Bildschirm —, am fremden Tisch und bei
// der Deckleihe. Wer gemeint ist, steht aber in den Kontakten: hier steht er
// zur Wahl.
//
// # Wer dasteht
//
// **Bestätigte Gegenüber**, und wo ein Zweck verlangt ist, nur die, denen er
// freigegeben ist — am fremden Tisch „Eine Partie spielen", genau die Frage,
// die `service.Peersitzungen.Oeffne` stellt, bevor es wählt. Ein Eintrag ohne
// das Recht wäre ein Weg, der mit „darf nicht" endet.
import { useId } from 'react'

import { Ladefehler, Wirdgeladen } from './Ladehinweise'
import { Suchauswahl } from './Suchauswahl'
import { suchschwelle } from './suchwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

import type { Peerkontakt } from '../apipeerkontakt'

// derBestaetigteStand ist der Stand, mit dem überhaupt etwas geht.
const derBestaetigteStand = 'bestaetigt'

// wahlstil ist ein Optionsfeld — **mit sichtbarem Fokusring** (§6).
const wahlstil =
  'mt-1 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

// eintragstext ist die Zeile eines Gegenübers: sein Name — der eigene, sonst
// der, den es sich selbst gibt — und sein Fingerabdruck, die Form, die zwei
// Menschen vergleichen.
function eintragstext(vorlage: string, kontakt: Peerkontakt): string {
  return einsetzen(vorlage, {
    name: kontakt.eigenerName || kontakt.selbstgewaehlterName,
    fingerabdruck: kontakt.fingerabdruck,
  })
}

interface Kontaktwahleigenschaften {
  // titel ist die Frage über der Wahl, leer der Satz, wenn niemand dasteht —
  // er sagt, wie jemand dazukommt.
  titel: string
  leer: string

  // recht ist der Zweck, der freigegeben sein muss — ohne: keiner.
  recht?: string | undefined

  // gewaehlt ist die Peerkennung des Gewählten — leer heißt keiner.
  gewaehlt: string
  waehlen: (kennung: string) => void

  // stand steigt, wenn daneben ein Gegenüber aufgenommen, bestätigt oder
  // freigegeben wurde — dann lädt die Wahl nach (wie `Partnerwahl`).
  stand?: number | undefined
}

export function Kontaktwahl({
  titel,
  leer,
  recht,
  gewaehlt,
  waehlen,
  stand = 0,
}: Kontaktwahleigenschaften) {
  const eigene = useTexte().bausteine
  const gruppe = useId()
  const suchfeld = useId()

  const liste = useLaden(async () => {
    const antwort = await api().Peerkontakte()

    return auspacken(antwort, () => antwort.kontakte).filter(
      (kontakt) =>
        kontakt.stand === derBestaetigteStand &&
        (recht === undefined || kontakt.rechte.includes(recht)),
    )
  }, [stand, recht])

  // **Ab fünfzehn Gegenübern eine Auswahl mit Suche** — UXR-F1
  // (#2007),
  // R7. Bis vierzehn sieht man alle samt Fingerabdruck; darüber überblickt sie
  // niemand mehr, und man tippt den Namen. Die Beschriftung steht neben dem
  // Feld statt in einer `legend`: das Feld ist keine Gruppe.
  if (liste.daten !== undefined && liste.daten.length >= suchschwelle) {
    return (
      <div className="flex flex-col gap-1">
        <label htmlFor={suchfeld} className="mb-1 text-sm font-medium">
          {titel}
        </label>
        <Suchauswahl
          id={suchfeld}
          bezeichnung={titel}
          eintraege={liste.daten.map((kontakt) => ({
            wert: kontakt.kennung,
            name: eintragstext(eigene.kontaktwahlEintrag, kontakt),
          }))}
          wert={gewaehlt}
          aendern={waehlen}
          platzhalter={eigene.kontaktwahlSuchen}
        />
      </div>
    )
  }

  return (
    <fieldset className="flex flex-col gap-1">
      <legend className="mb-1 text-sm font-medium">{titel}</legend>

      {/* **Warten, Fehlschlag, leer — drei Sätze** (UX-08.NAT-01): solange
          die Antwort aussteht, steht das da, und ein Fehlschlag bietet den
          nächsten Versuch an. */}
      {liste.daten === undefined ? (
        liste.fehler === undefined ? (
          <Wirdgeladen />
        ) : (
          <Ladefehler fehler={liste.fehler} wiederholen={liste.nachladen} />
        )
      ) : liste.daten.length === 0 ? (
        <p className="max-w-prose text-sm text-leise">{leer}</p>
      ) : (
        liste.daten.map((kontakt) => (
          <label
            key={kontakt.kennung}
            className="flex cursor-pointer items-start gap-2 rounded-sm border border-kante px-2 py-1.5 has-[:checked]:border-marke"
          >
            <input
              type="radio"
              name={gruppe}
              value={kontakt.kennung}
              checked={gewaehlt === kontakt.kennung}
              onChange={() => {
                waehlen(kontakt.kennung)
              }}
              className={wahlstil}
            />
            <span>{eintragstext(eigene.kontaktwahlEintrag, kontakt)}</span>
          </label>
        ))
      )}
    </fieldset>
  )
}
