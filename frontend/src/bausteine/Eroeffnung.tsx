// **Der Beginn einer Partie** — Starthand, Mulligan, Unterlegen und
// Behalten, CR 103.5
// (#1839).
//
// # Warum eine Tafel über dem Tisch und nicht vier Knöpfe in der Spalte
//
// **Weil der Beginn das Spiel anhält**, wie die Legendenwahl daneben: bis
// alle behalten haben, gibt es nichts anderes zu tun. Und weil das
// Unterlegen Karten aus der **eigenen Hand** wählt — eine Liste, die die
// Handlungsspalte nicht kennt.
//
// # Was sie nicht rechnet
//
// **Ob ein Mulligan noch geht**, weiß der Kern: im Spiel zu mehreren zählt
// der erste nicht (CR 103.5c), und nach der leeren Starthand ist Schluss.
// Die Tafel zeigt, was `erlaubt` nennt — eine zweite Rechnung hier liefe mit
// der ersten auseinander (§2).
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { handlungswort } from '../seiten/partiewoerter'
import { kartenwort } from '../spiel/feldwerte'
import { useTexte } from '../sprache/kontext'

import type { Partie, Sitz, Spielobjekt } from '../apipartie'
import { Abschnittstitel } from './Abschnittstitel'

// Beginnwege ist, was die Tafel vom Tisch braucht.
interface Beginnwege {
  erlaubt: readonly string[]
  angehalten: boolean
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
}

export function Eroeffnung({
  partie,
  ...wege
}: { partie: Partie } & Beginnwege) {
  const eigene = useTexte().seiten.spielen
  const titel = useId()

  const imBeginn = partie.sitze.filter(
    (sitz) => sitz.eroeffnung === 'ausstehend' || sitz.eroeffnung === 'offen',
  )

  if (imBeginn.length === 0) {
    return undefined
  }

  const eigener = imBeginn.find((sitz) => sitz.kennung === partie.fuer)

  return (
    <section aria-labelledby={titel} className="flex flex-col gap-2">
      <Abschnittstitel id={titel} className="text-base font-semibold">
        {eigene.beginnTitel}
      </Abschnittstitel>

      {eigener !== undefined && (
        <EigenerBeginn
          // **Ein neuer Mulligan räumt die Wahl ab** (§7, Regel 1): die
          // Karten von vorhin liegen nicht mehr auf der Hand.
          key={`${partie.id}:${eigener.kennung}:${String(eigener.mulligans)}`}
          partie={partie}
          sitz={eigener}
          {...wege}
        />
      )}

      {imBeginn
        .filter((sitz) => sitz !== eigener)
        .map((sitz) => (
          <p key={sitz.kennung} className="text-sm text-leise">
            {einsetzen(wartewort(eigene, sitz), { wer: sitz.name })}
          </p>
        ))}
    </section>
  )
}

// EigenerBeginn ist der Schritt, der für den eigenen Sitz ansteht.
function EigenerBeginn({
  partie,
  sitz,
  erlaubt,
  angehalten,
  befehlen,
}: { partie: Partie; sitz: Sitz } & Beginnwege) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const [gewaehlt, setzeGewaehlt] = useState<string[]>([])

  // knopf stellt einen Schritt — **nur, wenn der Kern ihn annähme**.
  const knopf = (art: string, nutzlast = {}, bereit = true) =>
    erlaubt.includes(art) && (
      <Knopf
        key={art}
        gewicht={art === 'spieler.mulligan' ? 'neben' : 'haupt'}
        angehalten={angehalten || !bereit}
        aufKlick={() => {
          befehlen(art, nutzlast)
        }}
      >
        {handlungswort(texte.seiten.handlungen.handlungsname, art)}
      </Knopf>
    )

  if (sitz.eroeffnung === 'ausstehend') {
    return (
      <>
        <p className="text-sm">{eigene.beginnStarthand}</p>
        <div>{knopf('spieler.starthand')}</div>
      </>
    )
  }

  const hand = handkarten(partie, sitz.kennung)

  if (sitz.unterlegen > 0) {
    // **In der Reihenfolge des Wählens** — CR 103.5 lässt sie offen, und
    // wer sie bestimmt, ist der Spieler.
    const umschalten = (kennung: string) => {
      setzeGewaehlt((vorher) =>
        vorher.includes(kennung)
          ? vorher.filter((eine) => eine !== kennung)
          : [...vorher, kennung],
      )
    }

    return (
      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm">
          {sitz.unterlegen === 1
            ? eigene.beginnUnterlegenEine
            : einsetzen(eigene.beginnUnterlegen, {
                zahl: String(sitz.unterlegen),
              })}
        </legend>

        <p className="text-sm text-leise">{eigene.beginnReihenfolge}</p>

        <ul
          aria-label={eigene.beginnHandkarten}
          className="flex flex-col gap-1"
        >
          {hand.map((objekt) => (
            <li key={objekt.kennung}>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  className={kaestchen}
                  checked={gewaehlt.includes(objekt.kennung)}
                  onChange={() => {
                    umschalten(objekt.kennung)
                  }}
                />
                {kartenwort(objekt)}
              </label>
            </li>
          ))}
        </ul>

        <p className="text-sm text-leise" aria-live="polite">
          {einsetzen(eigene.beginnGewaehlt, {
            zahl: String(gewaehlt.length),
            soll: String(sitz.unterlegen),
          })}
        </p>

        <div>
          {knopf(
            'spieler.unterlegen',
            { objekte: gewaehlt },
            gewaehlt.length === sitz.unterlegen,
          )}
        </div>
      </fieldset>
    )
  }

  return (
    <>
      <p className="text-sm">
        {hand.length === 1
          ? eigene.beginnHandEine
          : einsetzen(eigene.beginnHand, { zahl: String(hand.length) })}
      </p>

      <p className="text-sm text-leise">{eigene.beginnMulligan}</p>

      <div className="flex flex-wrap gap-2">
        {knopf('spieler.behalten')}
        {knopf('spieler.mulligan')}
      </div>
    </>
  )
}

// handkarten sind die Karten auf der Hand eines Sitzes — ungeordnet wie die
// Hand selbst, in der Reihenfolge der Partie.
function handkarten(partie: Partie, sitz: string): Spielobjekt[] {
  return partie.objekte.filter(
    (objekt) => objekt.zone === 'hand' && objekt.beherrscher === sitz,
  )
}

// wartewort sagt, wo ein anderer Sitz im Beginn steht.
function wartewort(
  eigene: ReturnType<typeof useTexte>['seiten']['spielen'],
  sitz: Sitz,
): string {
  if (sitz.eroeffnung === 'ausstehend') {
    return eigene.beginnWartetAusstehend
  }

  return sitz.unterlegen > 0
    ? eigene.beginnWartetUnterlegen
    : eigene.beginnWartetOffen
}

// kaestchen sind die Klassen eines Kästchens — der sichtbare Fokusring (§6),
// am Element und nicht an der Beschriftung.
const kaestchen =
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring '
