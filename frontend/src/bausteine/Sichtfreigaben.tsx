// **Was ein Gegenüber von der Sammlung sieht** — Werkstück T3 von Phase 5.6
// (#1468).
//
// # Warum der Kasten erst bei einem bestätigten Gegenüber erscheint
//
// **Weil eine Freigabe ohne Bestätigung nichts tut.** `domain.FreigabenFuer`
// fragt zuerst den Kontakt: wer nur gepaart oder gesperrt ist, sieht nichts,
// gleich welche Freigaben danebenliegen. Ein Formular, das etwas erteilt, das
// nicht wirkt, wäre eine Zusage ohne Einlösung.
//
// # Und warum die Vorschau dabeisteht
//
// **Weil niemand einer Zusage glauben muss.** „Deine Preise gehen nicht
// hinaus" ist ein Satz; die Vorschau ist der Blick darauf — und sie geht
// durch **denselben** Weg wie die Antwort an das Gegenüber.
//
// # Und warum zwei Umfänge ein Feld mitbringen
//
// **Weil die Domäne zwei Angaben zum Umfang verlangt** (UXR-F18,
// #2033;
// `Sammlungsfreigabe.pruefeUmfangsangaben`): „Ausgewählte Drucke" braucht
// mindestens einen Druck, „Dubletten" eine Behaltemenge. Das Fenster bot beide
// Umfänge an und ließ nichts davon setzen — die Auswahl scheiterte immer, und
// die Dubletten gingen mit 0 hinaus. **Jedes Feld steht nur bei seinem
// Umfang, und ein anderer Umfang schickt es nicht mit**: die Domäne weist
// eine Auswahl bei einem anderen Umfang ebenso ab wie eine Behaltemenge
// außerhalb der Dubletten.
import { useId, useState } from 'react'

import { Behaltemenge } from './Behaltemenge'
import { Druckwahl } from './Druckwahl'
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { zwecksatz } from '../zwecksatz'

import type { Sichtfreigabe, Sichtkarte } from '../apisichtfreigabe'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

// Freigabebefunde sind die Sätze, die der Dienst an einem Feld meldet.
export interface Freigabebefunde {
  auswahl?: string | undefined
  behalten?: string | undefined
}

// umfangAuswahl und umfangDubletten sind die beiden Umfänge, die eine Angabe
// mitbringen — ihre Schlüssel aus `domain.Freigabeumfang`.
const umfangAuswahl = 'auswahl'
const umfangDubletten = 'dubletten'

// vorgabeBehalten ist die Behaltemenge, die ein neues Formular vorschlägt —
// wie `mse dubletten` und die Ordnung der Sammlung (siehe [Behaltemenge]).
const vorgabeBehalten = 1

interface Sichtfreigabeneigenschaften {
  kennung: string
  freigaben: Sichtfreigabe[]
  umfaenge: string[]
  zwecke: string[]
  vorschau: Sichtkarte[] | undefined
  laeuft: boolean

  // befunde sind die Sätze, die der Dienst an einem Feld gemeldet hat, und
  // befundGeraeumt sagt, dass sich seither etwas geändert hat (§6, §7).
  befunde: Freigabebefunde
  befundGeraeumt: () => void
  erteilen: (freigabe: Sichtfreigabe) => void
  widerrufen: (zweck: string, umfang: string) => void
  zeigen: (zweck: string) => void
}

export function Sichtfreigaben({
  kennung,
  freigaben,
  umfaenge,
  zwecke,
  vorschau,
  laeuft,
  befunde,
  befundGeraeumt,
  erteilen,
  widerrufen,
  zeigen,
}: Sichtfreigabeneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const worteDrucke = texte.bausteine.druckwahl
  const zweckfeld = useId()
  const umfangfeld = useId()

  // **Der Zweck steht einmal da und gilt für beides** — erteilen und
  // vorschauen. Zwei Wahlen für dieselbe Frage ließen jemanden eine Sicht
  // für „teilen" erteilen und die für „handel" ansehen.
  // **Der erste erlaubte Zweck ist die Vorgabe** — und wenn es keinen
  // gibt, steht das Formular gar nicht da.
  const [zweck, setzeZweck] = useState(erlaubteZwecke(zwecke)[0] ?? '')
  const [umfang, setzeUmfang] = useState(umfaenge[0] ?? '')

  // **Die Angaben zum Umfang bleiben, wenn der Umfang wechselt** — wer
  // zwischen zwei Umfängen hin- und herschaltet, verliert seine Wahl nicht.
  // Geschickt wird sie nur beim passenden Umfang (siehe `erteile`).
  const [auswahl, setzeAuswahl] = useState<string[]>([])
  const [behalten, setzeBehalten] = useState(vorgabeBehalten)

  // behaltenGilt sagt, ob der getippte Text der Behaltemenge eine Zahl ist,
  // die gelten darf — **sonst schickte der Knopf still den letzten gültigen
  // Wert** (siehe `Zahlenfeld.gilt`).
  const [behaltenGilt, setzeBehaltenGilt] = useState(true)

  // ohneAuswahl merkt, dass jemand erteilen wollte, ohne einen Druck gewählt
  // zu haben — **der Befund steht erst dann da**: eine frische Wahl ist leer,
  // und ein Fehler vor dem ersten Versuch wäre Lärm.
  const [ohneAuswahl, setzeOhneAuswahl] = useState(false)

  // aendert räumt die Befunde: wer etwas ändert, hat den Stand vor sich, auf
  // den sie sich bezogen (§7, Regel 1).
  const aendert = () => {
    setzeOhneAuswahl(false)
    befundGeraeumt()
  }

  // angabe nennt, was an einer erteilten Sicht zu lesen sein soll: wie viele
  // Drucke, oder was jemand behält. **Die anderen Umfänge haben nichts zu
  // ergänzen** — und eine Lücke an dieser Stelle läse sich wie ein Fehler.
  const angabe = (freigabe: Sichtfreigabe): string | undefined => {
    if (freigabe.umfang === umfangAuswahl) {
      return zaehlwort(
        freigabe.auswahl.length,
        texte.allgemein.zaehlwoerter.drucke,
      )
    }

    if (freigabe.umfang === umfangDubletten) {
      return einsetzen(worteDrucke.detailBehalten, {
        behalten: String(freigabe.behalten),
      })
    }

    return undefined
  }

  const erteile = () => {
    if (umfang === umfangAuswahl && auswahl.length === 0) {
      setzeOhneAuswahl(true)

      return
    }

    erteilen({
      empfaenger: kennung,
      umfang,
      zweck,
      // **Nur beim eigenen Umfang** — die Domäne weist beides sonst ab.
      auswahl: umfang === umfangAuswahl ? auswahl : [],
      behalten: umfang === umfangDubletten ? behalten : 0,
      ablauf: '',
      erteilt: '',
    })
  }

  return (
    // **Die Überschrift trägt der Aufklapper darüber** (UXR-16, #1924) —
    // zwei Namen für dieselbe Sache sind einer zu viel.
    <section aria-label={eigene.freigabenTitel} className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.freigabenErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {freigaben.length === 0 && (
        <p className="text-sm text-leise">{eigene.freigabenLeer}</p>
      )}

      {/* **Ohne erlaubten Zweck gibt es nichts zu erteilen** — eine Sicht
          für einen Zweck, den dieses Gegenüber nicht darf, gilt nie.
          **Und ohne Umfänge ebenso**: sie kommen aus Go, und eine leere
          Liste hieße, dass diese Fassung keinen kennt. Ein Formular mit
          leerer Auswahl erteilte dann etwas Namenloses. */}
      {(erlaubteZwecke(zwecke).length === 0 || umfaenge.length === 0) && (
        <p className="max-w-prose text-sm text-leise">
          {eigene.freigabenOhneZweck}
        </p>
      )}

      {freigaben.length > 0 && (
        <ul className="flex flex-col gap-1">
          {freigaben.map((freigabe) => (
            <li
              key={`${freigabe.zweck}/${freigabe.umfang}`}
              className="flex flex-wrap items-center gap-2 text-sm"
            >
              <span>
                {zwecksatz(eigene, freigabe.zweck)} ·{' '}
                {umfangsatz(eigene, freigabe.umfang)}
                {angabe(freigabe) !== undefined && <> · {angabe(freigabe)}</>}
              </span>

              <Knopf
                gewicht="leise"
                angehalten={laeuft}
                aufKlick={() => {
                  widerrufen(freigabe.zweck, freigabe.umfang)
                }}
              >
                {eigene.freigabenWiderrufen}
              </Knopf>
            </li>
          ))}
        </ul>
      )}

      {erlaubteZwecke(zwecke).length > 0 && umfaenge.length > 0 && (
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-sm" htmlFor={zweckfeld}>
              {eigene.freigabenZweck}
              <select
                id={zweckfeld}
                className={wahlfeld}
                value={zweck}
                onChange={(ereignis) => {
                  setzeZweck(ereignis.target.value)
                  aendert()
                }}
              >
                {erlaubteZwecke(zwecke).map((eintrag) => (
                  <option key={eintrag} value={eintrag}>
                    {zwecksatz(eigene, eintrag)}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1 text-sm" htmlFor={umfangfeld}>
              {eigene.freigabenUmfang}
              <select
                id={umfangfeld}
                className={wahlfeld}
                value={umfang}
                onChange={(ereignis) => {
                  setzeUmfang(ereignis.target.value)

                  // **Das Zahlenfeld steht beim Zurückkommen wieder mit dem
                  // letzten gültigen Wert da** — ein halb getippter Text
                  // überlebt den Wechsel nicht, und ein „gilt nicht" dürfte
                  // es auch nicht.
                  setzeBehaltenGilt(true)
                  aendert()
                }}
              >
                {umfaenge.map((eintrag) => (
                  <option key={eintrag} value={eintrag}>
                    {umfangsatz(eigene, eintrag)}
                  </option>
                ))}
              </select>
            </label>
          </div>

          {/* **Die Angabe zum Umfang steht nur bei ihrem Umfang** (UXR-F18,
              #2033) — und ihr Fehler an ihr (§6), nicht über dem Kasten. */}
          {umfang === umfangAuswahl && (
            <Druckwahl
              gewaehlt={auswahl}
              angehalten={laeuft}
              befund={ohneAuswahl ? worteDrucke.fehlt : befunde.auswahl}
              waehlen={(neu) => {
                setzeAuswahl(neu)
                aendert()
              }}
            />
          )}

          {umfang === umfangDubletten && (
            <Behaltemenge
              wert={behalten}
              angehalten={laeuft}
              befund={befunde.behalten}
              gilt={setzeBehaltenGilt}
              geaendert={(neu) => {
                setzeBehalten(neu)
                aendert()
              }}
            />
          )}

          <div className="flex flex-wrap items-end gap-2">
            <Knopf
              gewicht="leise"
              angehalten={
                laeuft || (umfang === umfangDubletten && !behaltenGilt)
              }
              aufKlick={erteile}
            >
              {eigene.freigabenErteilen}
            </Knopf>

            <Knopf
              gewicht="leise"
              angehalten={laeuft}
              aufKlick={() => {
                zeigen(zweck)
              }}
            >
              {eigene.freigabenVorschau}
            </Knopf>
          </div>
        </div>
      )}

      {vorschau !== undefined && (
        <ul
          aria-label={eigene.freigabenVorschau}
          className="flex flex-col gap-1"
        >
          {vorschau.map((zeile) => (
            <li key={zeile.referenz} className="text-sm tabular-nums">
              {zeile.menge} × {zeile.bezeichnung} · {zeile.zustand} ·{' '}
              {zeile.sprache} · {zeile.finish}
            </li>
          ))}
        </ul>
      )}

      {vorschau?.length === 0 && (
        <p className="text-sm text-leise">{eigene.freigabenVorschauLeer}</p>
      )}
    </section>
  )
}

// umfangsatz gibt den Namen eines Umfangs — UXR-16 (#1924): bis dahin stand
// hier der Schlüssel aus Go, „dubletten" und „digitales".
//
// **Der Rückfall ist der Schlüssel selbst**, wie bei `zwecksatz`: einen
// sechsten Umfang, den dieses Fenster noch nicht kennt, nennt es so, wie Go
// ihn schickt.
function umfangsatz(
  eigene: { gegenueber: { umfaenge: Record<string, string> } },
  umfang: string,
): string {
  return eigene.gegenueber.umfaenge[umfang] ?? umfang
}

// erlaubteZwecke sind die, für die sich etwas freigeben lässt.
//
// **Die Rechte dieses Gegenübers und nicht alle vier.** Eine Sicht für einen
// Zweck, den es gar nicht darf, gilt nie: `domain.FreigabenFuer` fragt zuerst
// den Kontakt. Wer sie anböte, ließe jemanden etwas erteilen, das nichts tut.
//
// **`paarung` gehört auch dann nicht dazu**, wenn sie dasteht — sie ist das
// Kennenlernen, und die Domäne weist eine Freigabe darauf ab.
function erlaubteZwecke(zwecke: string[]): string[] {
  return zwecke.filter((zweck) => zweck !== 'paarung')
}

const wahlfeld =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'
