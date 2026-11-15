// **Mit wem gehandelt wird** — Werkstück T9 von Phase 5.6
// (#1474).
//
// # Warum nur bestätigte Gegenüber dastehen
//
// **Weil an einem unbestätigten nichts wirkt.** Eine Freigabe fragt zuerst
// den Kontakt (`domain.FreigabenFuer`), und ein Abgleich ohne Freigabe
// bekommt nichts — ein Eintrag hier wäre ein Weg, der ins Leere führt.
// Aufnehmen und Bestätigen stehen unter Technik, und der Text sagt es.
//
// # Und warum sie geladen und nicht gereicht wird
//
// **Weil dieser Bereich nicht an der Technikseite hängt.** Wer ihn über ⌘0
// aufschlägt, hat keine Kontaktliste in der Hand; sie hier zu laden ist ein
// Ruf und keine Kopplung.
import { Ladefehler, Wirdgeladen } from './Ladehinweise'
import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

import type { Peerkontakt } from '../apipeerkontakt'

// derBestaetigteStand ist der Stand, mit dem gehandelt werden darf.
const derBestaetigteStand = 'bestaetigt'

interface Partnerwahleigenschaften {
  // offen ist die Kennung des aufgeschlagenen Gegenübers — leer heißt keines.
  offen: string | undefined

  // waehlen nennt das Gegenüber mit seinem Namen — der Raum spricht von ihm
  // (UXR-15, #1923) und nicht von „deinem Gegenüber".
  waehlen: (kennung: string, name: string) => void

  // stand steigt, wenn daneben ein Gegenüber aufgenommen oder bestätigt wurde
  // (UX-03.AK-03) — dann lädt die Wahl nach.
  stand?: number | undefined
}

export function Partnerwahl({
  offen,
  waehlen,
  stand = 0,
}: Partnerwahleigenschaften) {
  const eigene = useTexte().seiten.handeln

  const liste = useLaden(async () => {
    const antwort = await api().Peerkontakte()

    return auspacken(antwort, () => antwort.kontakte).filter(
      (kontakt) => kontakt.stand === derBestaetigteStand,
    )
  }, [stand])

  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-base font-semibold">{eigene.partnerTitel}</h2>

      {/* **Ein Satz, der Rest hinter dem Hilfesymbol** (UX-15, #1771: lang und ganz sichtbar, gemessen in WebKit am 2026-09-24). */}
      <Erklaerungsabsatz text={eigene.partnerErklaerung} />

      {/* **Warten, Fehlschlag, leer — drei Sätze** (UX-08.NAT-01, S15.NAT-02):
          solange die Antwort aussteht, steht das da, und ein Fehlschlag
          bietet den nächsten Versuch an. */}
      {liste.daten === undefined && liste.fehler === undefined && (
        <Wirdgeladen />
      )}

      {liste.fehler !== undefined && (
        <Ladefehler fehler={liste.fehler} wiederholen={liste.nachladen} />
      )}

      {liste.daten?.length === 0 && (
        <p className="text-sm text-leise">{eigene.partnerLeer}</p>
      )}

      {liste.daten !== undefined && liste.daten.length > 0 && (
        <ul className="flex flex-col gap-1">
          {liste.daten.map((kontakt) => (
            <Partnerzeile
              key={kontakt.kennung}
              kontakt={kontakt}
              aufgeschlagen={kontakt.kennung === offen}
              waehlen={waehlen}
            />
          ))}
        </ul>
      )}
    </section>
  )
}

// Partnerzeile ist ein Gegenüber zur Wahl.
//
// **Der Knopf nennt das Gegenüber im `aria-label`** — „Aufschlagen" stünde
// sonst zehnmal gleich da, und wer die Liste hört, hörte zehnmal dasselbe
// (§6).
function Partnerzeile({
  kontakt,
  aufgeschlagen,
  waehlen,
}: {
  kontakt: Peerkontakt
  aufgeschlagen: boolean
  waehlen: (kennung: string, name: string) => void
}) {
  const eigene = useTexte().seiten.handeln

  const name = kontakt.eigenerName || kontakt.selbstgewaehlterName

  return (
    <li className="flex items-center justify-between gap-2">
      <span className="min-w-0 truncate text-sm">
        {name}
        <span className="ml-2 text-xs text-leise tabular-nums">
          {kontakt.fingerabdruck}
        </span>
      </span>

      {/* **Ein Knopf in beiden Zuständen** (UX-10.AK-02, #1682): er bleibt
          nach dem Aufschlagen, gedrückt, und sein Name folgt der Aufschrift
          (WCAG 2.5.3). Verschwände er, stünde der Fokus danach auf `body`,
          sichtbar nirgends. */}
      <Knopf
        gewicht="leise"
        gedrueckt={aufgeschlagen}
        bezeichnung={einsetzen(
          aufgeschlagen ? eigene.partnerOffenName : eigene.partnerWaehlenName,
          { name },
        )}
        aufKlick={() => {
          waehlen(kontakt.kennung, name)
        }}
      >
        {aufgeschlagen ? eigene.partnerOffen : eigene.partnerWaehlen}
      </Knopf>
    </li>
  )
}
