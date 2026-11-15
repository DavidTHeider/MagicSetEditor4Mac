// **Was zu einem Gegenüber gehört** — aus `Peerkontaktzeile.tsx`
// herausgelöst, als die Zeile mit UXR-16
// (#1924)
// einen Kopf bekam. **Der Schnitt folgt der Frage**: dort steht, **wer** es
// ist und in welchem Stand; hier, was man mit ihm tut — vergleichen,
// freigeben, sperren.
//
// # Und warum jeder Zweck seine eigene Wahl hat
//
// **Weil jeder eine eigene Entscheidung ist.** „Darf Karten tauschen" und
// „darf eine Partie spielen" sind nicht dasselbe, und eine Sammelfreigabe
// wäre die Gestalt, in der jemand das eine will und das andere gibt.
import { useId } from 'react'

import { Abdruckvergleich } from './Abdruckvergleich'
import { Aufklappkasten } from './Aufklappkasten'
import { Codevergleich } from './Codevergleich'
import { Knopf } from './Knopf'
import { Kontaktadresse } from './Kontaktadresse'
import { Kontaktfreigaben } from './Kontaktfreigaben'
import { Kontaktnotiz } from './Kontaktnotiz'
import { einsetzen } from '../sprache/einsetzen'
import { zwecksatz } from '../zwecksatz'
import { useTexte } from '../sprache/kontext'

import type { Peerkontakt } from '../apipeerkontakt'

// Kontaktwege sind die schreibenden Wege einer Zeile — die Seite führt sie.
export interface Kontaktwege {
  bestaetigen: (kennung: string, verglichen: string) => void
  sperren: (kennung: string) => void
  entsperren: (kennung: string) => void
  erlauben: (kennung: string, zweck: string) => void
  entziehen: (kennung: string, zweck: string) => void
  adresseSetzen: (kennung: string, adresse: string) => void
  notizSetzen: (kennung: string, notiz: string) => void
}

interface Einzelheiteneigenschaften extends Kontaktwege {
  kontakt: Peerkontakt

  // name ist, wie die Zeile darüber das Gegenüber nennt.
  name: string
  zwecke: string[]
  laeuft: boolean
}

export function Peerkontakteinzelheiten({
  kontakt,
  name,
  zwecke,
  laeuft,
  bestaetigen,
  sperren,
  entsperren,
  erlauben,
  entziehen,
  adresseSetzen,
  notizSetzen,
}: Einzelheiteneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik

  return (
    <>
      {/* **Der Vorbehalt zum Namen** — und der Name selbst nur, wenn die
          Zeile einen anderen trägt: steht oben schon, wie es sich nennt,
          wäre er hier eine Wiederholung. */}
      {kontakt.selbstgewaehlterName !== '' && (
        <p className="text-sm text-leise">
          {kontakt.selbstgewaehlterName !== name && (
            <>
              {einsetzen(eigene.kontakteNennt, {
                name: kontakt.selbstgewaehlterName,
              })}{' '}
            </>
          )}
          {eigene.kontakteNenntHinweis}
        </p>
      )}

      {kontakt.stand === 'unbestaetigt' && kontakt.vergleichscode !== '' && (
        <Codevergleich
          kontakt={kontakt}
          laeuft={laeuft}
          bestaetigen={bestaetigen}
        />
      )}

      {kontakt.stand === 'unbestaetigt' && kontakt.vergleichscode === '' && (
        <Abdruckvergleich
          kennung={kontakt.kennung}
          laeuft={laeuft}
          bestaetigen={bestaetigen}
        />
      )}

      {/* **Der Abdruck ist eine technische Angabe** (NET-01.AK-01, #1705):
          verglichen wird der Code, und wer den Abdruck braucht, klappt auf. */}
      <Aufklappkasten titel={eigene.einladung.technik}>
        <dl className="flex flex-col gap-1 text-sm">
          <dt className="font-medium">{eigene.kontakteAbdruck}</dt>
          <dd className="font-mono break-all">{kontakt.fingerabdruck}</dd>
        </dl>
      </Aufklappkasten>

      {/* **Adresse und Notiz hinter einem Aufklapper** (UXR-16, #1924): man
          setzt sie selten — die Adresse kommt mit der Einladung oder dem
          Fund im Netz —, und offen trugen sie je Zeile zwei Felder, zwei
          Knöpfe und zwei Hilfen.

          **Auch bei einem gesperrten Gegenüber.** Eine Sperre nimmt ihm,
          was es darf — nicht, wo es wohnt, und nicht, was man über es denkt;
          und wer sie zurücknimmt, will nicht die Adresse noch einmal tippen.

          **Der Schlüssel trägt seinen Namen vorweg** (T4, #1469): Adresse und
          Notiz sind Geschwister, und leer waren beide Schlüssel derselbe.
          React nennt das ausdrücklich nicht unterstützt — Kinder dürfen dabei
          verdoppelt oder ausgelassen werden, und hier hängt §7 Regel 1 daran:
          der `key` räumt das Bauteil beim Wechsel, und zwei gleiche räumen
          möglicherweise das falsche. */}
      <Aufklappkasten titel={eigene.gegenueber.adresseUndNotiz}>
        <div className="flex flex-col gap-3">
          <Kontaktadresse
            key={`adresse:${kontakt.adresse}`}
            kennung={kontakt.kennung}
            adresse={kontakt.adresse}
            laeuft={laeuft}
            setzen={adresseSetzen}
          />
          <Kontaktnotiz
            key={`notiz:${kontakt.notiz}`}
            kennung={kontakt.kennung}
            notiz={kontakt.notiz}
            laeuft={laeuft}
            setzen={notizSetzen}
          />
        </div>
      </Aufklappkasten>

      {kontakt.stand === 'bestaetigt' && (
        <Zweckliste
          kontakt={kontakt}
          zwecke={zwecke}
          laeuft={laeuft}
          erlauben={erlauben}
          entziehen={entziehen}
        />
      )}

      {/* **Erst bei einem bestätigten Gegenüber** (T3, #1468): eine Freigabe
          an einen unbestätigten tut nichts — `domain.FreigabenFuer` fragt
          zuerst den Kontakt. Ein Formular, das etwas erteilt, das nicht
          wirkt, wäre eine Zusage ohne Einlösung.

          **Und hinter einem Aufklapper** (UXR-16): ohne Freigabe sieht es
          nichts, und erteilt wird selten. Gefragt wird der Bestand erst, wenn
          jemand ihn öffnet — offen fragte jede Zeile schon beim Zeichnen.

          **Abgleich, Post und Verbuchung stehen seit T9 (#1474) im Bereich
          „Handeln"** und nicht hier: was hier bleibt, ist die **Identität** —
          aufnehmen, bestätigen, sperren, freigeben. */}
      {kontakt.stand === 'bestaetigt' && (
        <Aufklappkasten titel={eigene.freigabenTitel}>
          <Kontaktfreigaben kennung={kontakt.kennung} zwecke={kontakt.rechte} />
        </Aufklappkasten>
      )}

      <div className="flex flex-wrap gap-2">
        {kontakt.stand === 'gesperrt' ? (
          <Knopf
            gewicht="leise"
            angehalten={laeuft}
            aufKlick={() => {
              entsperren(kontakt.kennung)
            }}
          >
            {eigene.kontakteEntsperren}
          </Knopf>
        ) : (
          <Knopf
            gewicht="leise"
            angehalten={laeuft}
            aufKlick={() => {
              sperren(kontakt.kennung)
            }}
          >
            {eigene.kontakteSperren}
          </Knopf>
        )}
      </div>
    </>
  )
}

interface Zwecklisteneigenschaften {
  kontakt: Peerkontakt
  zwecke: string[]
  laeuft: boolean
  erlauben: (kennung: string, zweck: string) => void
  entziehen: (kennung: string, zweck: string) => void
}

// Zweckliste zeigt jeden Zweck mit seinem Schalter.
//
// **Alle Zwecke und nicht nur die freigegebenen.** Wer nur die gegebenen
// zeigte, ließe die Frage „was könnte ich noch freigeben" unbeantwortet — und
// eine Freigabe, die man nicht findet, gibt niemand zurück.
//
// **Ein Schalter und kein Knopf** — UXR-16 (#1924), Regel R7 aus
// `docs/UIREVIEW.md`: was sofort wirkt und zwei Werte hat, ist ein Schalter.
// Hier stand je Zweck ein Knopf „Freigeben" oder „Zurücknehmen", und ob ein
// Zweck gilt, las man aus dem Wort, das ihn **umkehrt**.
function Zweckliste({
  kontakt,
  zwecke,
  laeuft,
  erlauben,
  entziehen,
}: Zwecklisteneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const titel = useId()

  return (
    <div className="flex flex-col gap-1">
      <p id={titel} className="text-sm font-medium">
        {eigene.kontakteRechte}
      </p>

      <ul aria-labelledby={titel} className="flex flex-col gap-1">
        {zwecke
          .filter((zweck) => zweck !== 'paarung')
          .map((zweck) => {
            const hat = kontakt.rechte.includes(zweck)

            return (
              <li key={zweck}>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    role="switch"
                    checked={hat}
                    disabled={laeuft}
                    onChange={() => {
                      if (hat) {
                        entziehen(kontakt.kennung, zweck)

                        return
                      }

                      erlauben(kontakt.kennung, zweck)
                    }}
                    className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
                  />
                  {zwecksatz(eigene, zweck)}
                </label>
              </li>
            )
          })}
      </ul>
    </div>
  )
}
