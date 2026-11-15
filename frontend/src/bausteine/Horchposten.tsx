// **Für andere erreichbar sein** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum hier ein Knopf steht und kein Schalter
//
// **Weil Anfangen und Aufhören verschiedene Angaben brauchen.** Anfangen
// braucht eine Adresse, Aufhören nichts. Ein Schalter nähme die Adresse auch
// beim Ausschalten entgegen — und niemand wüsste, welche gilt.
//
// # Und warum die Adresse danach noch einmal dasteht
//
// **Weil sie dann eine andere ist.** Wer `:41234` eingibt, bekommt genau den
// Port; wer `:0` eingibt, bekommt einen freien vom Betriebssystem, und nur die
// Antwort sagt, welchen. Ohne sie hätte der Weg „Adresse von Hand" an einem
// Ende keine Hand.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Diagnoseverweis } from './Diagnoseverweis'

import type { Horchantwort } from '../apipeerkontakt'
import { kennungseingabe } from './kennungseingabe'

// vorgabeadresse ist die Adresse, die der Hinweis unter dem Feld nennt: der
// Port allein, auf jeder Schnittstelle dieses Rechners.
const vorgabeadresse = ':41234'

export function Horchposten() {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()
  const feld = useId()

  // **Das Feld beginnt mit der Adresse, die der Hinweis darunter erklärt**
  // (K02, #2188): als Platzhalter sah `:41234` aus wie ein Wert, und der
  // Knopf blieb ohne Erklärung angehalten — gesehen im Gast am 2026-10-07.
  const [adresse, setzeAdresse] = useState(vorgabeadresse)
  const [geaendert, setzeGeaendert] = useState(0)
  const [fehler, setzeFehler] = useState<string>()

  // ursache bleibt für „Diagnose öffnen" stehen (NET-02, S13.NAT-04, #1706).
  const [ursache, setzeUrsache] = useState<unknown>()

  const stand = useLaden<Horchantwort>(async () => {
    const antwort = await api().HorchStand()

    return auspacken(antwort, () => antwort)
  }, [geaendert])

  const laeuft = vorgang.laeuft('horchposten')

  // handeln fährt einen der beiden schreibenden Wege.
  //
  // **Ein Vorgang für beide** (§7, Regel 4): sie gehen an denselben Port, und
  // zwei Riegel ließen ein Beenden und ein Starten zugleich laufen.
  const handeln = (weg: () => Promise<Horchantwort>): Promise<void> =>
    vorgang.starte('horchposten', async () => {
      setzeFehler(undefined)
      setzeUrsache(undefined)

      try {
        auspacken(await weg(), () => undefined)

        // **Geschrieben ist geschrieben** (§7, Regel 3): der Stand wird erst
        // nach dem geglückten Ruf nachgeladen.
        setzeGeaendert((runde) => runde + 1)
      } catch (gescheitert) {
        setzeFehler(meldungAus(gescheitert, texte))
        setzeUrsache(gescheitert)
      }
    })

  // **Einmal festgehalten und nicht zweimal gefragt.** Ein zweiter Griff
  // (`stand.daten?.stand.adresse ?? ''`) trüge einen Rückfall, den es nicht
  // gibt: wer horcht, hat einen Stand — und ein Zweig, den kein Lauf betreten
  // kann, ist ein Zweig, den niemand prüft (§10).
  const daten = stand.daten
  const laufend = daten?.stand.laeuft === true ? daten.stand : undefined

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-3">
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {eigene.horchenTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.horchenErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      <Diagnoseverweis ursache={ursache} />

      {laufend !== undefined ? (
        <div className="flex flex-col gap-2">
          <p className="text-sm font-medium">
            {einsetzen(eigene.horchenLaeuft, { adresse: laufend.adresse })}
          </p>

          {/* **Wohin ein Gegenüber wählt — und nicht, wo gebunden wurde.**
              Der Befund aus der Sichtprüfung vom 2026-09-14: hier stand
              „Erreichbar unter `[::]:41234`", und darunter erreichte niemand
              diesen Rechner.

              **Der Name steht vor der Adresse**, weil er länger gilt: eine
              IP-Adresse kommt über DHCP, und eine Lease endet. */}
          {laufend.erreichbarNamen.length === 0 &&
          laufend.erreichbarAdressen.length === 0 ? (
            <p className="max-w-prose text-sm text-leise">
              {eigene.horchenNurHier}
            </p>
          ) : (
            <>
              <p className="max-w-prose text-sm text-leise">
                {eigene.horchenWeitergeben}
              </p>

              <Weitergeben
                bezeichnung={eigene.horchenNamen}
                eintraege={laufend.erreichbarNamen}
              />

              <Weitergeben
                bezeichnung={eigene.horchenAdressen}
                eintraege={laufend.erreichbarAdressen}
              />

              <p className="max-w-prose text-sm text-leise">
                {eigene.horchenNurDiesesNetz}
              </p>
            </>
          )}

          <Knopf
            gewicht="leise"
            angehalten={laeuft}
            aufKlick={() => {
              void handeln(() => api().HorchBeenden())
            }}
          >
            {eigene.horchenBeenden}
          </Knopf>

          {/* **Die zweite Entscheidung, und erst hier** — eine Ankündigung
              ohne Port wäre eine Einladung an eine Tür, hinter der niemand
              steht. */}
          <Ankuendigung
            sichtbar={laufend.sichtbar}
            laeuft={laeuft}
            handeln={handeln}
          />
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          <p className="text-sm text-leise">{eigene.horchenStill}</p>

          <p className="max-w-prose text-sm text-leise">
            {eigene.sichtbarBrauchtHorchen}
          </p>

          <div className="flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-sm" htmlFor={feld}>
              {eigene.horchenAdresse}
              <input
                {...kennungseingabe}
                id={feld}
                className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                placeholder={eigene.horchenAdresseBeispiel}
                value={adresse}
                onChange={(ereignis) => {
                  setzeAdresse(ereignis.target.value)
                }}
              />
            </label>

            <Knopf
              gewicht="haupt"
              angehalten={laeuft || adresse === ''}
              aufKlick={() => {
                void handeln(() => api().HorchStarten(adresse))
              }}
            >
              {eigene.horchenStarten}
            </Knopf>
          </div>

          <Erklaerungsabsatz
            text={eigene.horchenAdresseHinweis}
            klasse="max-w-prose text-sm text-leise"
          />
        </div>
      )}
    </section>
  )
}

interface Ankuendigungseigenschaften {
  sichtbar: boolean
  laeuft: boolean
  handeln: (weg: () => Promise<Horchantwort>) => Promise<void>
}

// Ankuendigung ist die zweite Entscheidung: gefunden werden.
//
// **Ein eigener Baustein**, weil `Horchposten` sonst über die Zeilengrenze aus
// §2 ginge — und weil er eine andere Frage beantwortet: dort, **ob** dieses
// Profil erreichbar ist, hier, ob es **jeder im Netz** erfährt.
function Ankuendigung({
  sichtbar,
  laeuft,
  handeln,
}: Ankuendigungseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const feld = useId()

  const [name, setzeName] = useState('')

  return (
    <div className="flex flex-col gap-2 border-t border-kante pt-3">
      <h4 className="text-sm font-semibold">{eigene.sichtbarTitel}</h4>

      <Erklaerungsabsatz
        text={eigene.sichtbarErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {sichtbar ? (
        <>
          <p className="text-sm font-medium">{eigene.sichtbarLaeuft}</p>

          <Knopf
            gewicht="leise"
            angehalten={laeuft}
            aufKlick={() => {
              void handeln(() => api().SichtbarkeitBeenden())
            }}
          >
            {eigene.sichtbarBeenden}
          </Knopf>
        </>
      ) : (
        <>
          <p className="text-sm text-leise">{eigene.sichtbarStill}</p>

          <label className="flex flex-col gap-1 text-sm" htmlFor={feld}>
            {eigene.sichtbarName}
            <input
              id={feld}
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              value={name}
              onChange={(ereignis) => {
                setzeName(ereignis.target.value)
              }}
            />
          </label>

          <Erklaerungsabsatz
            text={eigene.sichtbarNameHinweis}
            klasse="max-w-prose text-sm text-leise"
          />

          <Knopf
            gewicht="haupt"
            angehalten={laeuft}
            aufKlick={() => {
              void handeln(() => api().SichtbarWerden(name))
            }}
          >
            {eigene.sichtbarWerden}
          </Knopf>
        </>
      )}
    </div>
  )
}

// Weitergeben zeigt eine beschriftete Liste zum Abschreiben.
//
// **Eine Liste und kein Satz mit Platzhalter**: es sind mehrere, und wer sie
// in einen Satz zwänge, gäbe eine Zeichenkette heraus, aus der niemand die
// eine herausliest, die er braucht. Eine leere Liste zeigt gar nichts —
// warum das so ist, sagt der Kasten darüber.
function Weitergeben({
  bezeichnung,
  eintraege,
}: {
  bezeichnung: string
  eintraege: string[]
}) {
  if (eintraege.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-1">
      <p className="text-sm text-leise">{bezeichnung}</p>

      <ul className="flex flex-col gap-1">
        {eintraege.map((eintrag) => (
          <li key={eintrag} className="font-mono text-sm select-all">
            {eintrag}
          </li>
        ))}
      </ul>
    </div>
  )
}
