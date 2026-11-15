// **Nutzer und Profile ändern und löschen** — Werkstück 5 von Phase 4.12
// (#1351).
//
// # Der Name bestätigt das Löschen
//
// **Dieselbe Form wie beim Set** (#1218), und hier wiegt sie schwerer: ein
// gelöschter Nutzer nimmt seine Profile mit. Der Name kommt aus der Zeile, die
// dasteht — er ist keine zweite Eingabe, sondern die Bestätigung dessen, was
// gemeint war.
//
// # Was hier **nicht** steht
//
// **Keine verschachtelte Rückfrage** (§6): die Bestätigung erscheint **in** der
// Zeile, nicht in einem Dialog über einem Dialog.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Knopf } from '../bausteine/Knopf'
import { Nutzerbild } from '../bausteine/Nutzerbild'
import { Nutzerbildwahl } from './Nutzerbildwahl'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type {
  NutzerDTO,
  NutzerangabenDTO,
  ProfilangabenDTO,
  ProfilDTO,
} from '../api'

interface Eigenschaften {
  nutzer: readonly NutzerDTO[]

  // bilder sind die Profilbilder als Daten-URI, je Nutzerkennung.
  bilder: Record<string, string>

  // geaendert meldet nach oben, dass neu zu lesen ist.
  geaendert: () => void
}

export function Nutzerpflege({ nutzer, bilder, geaendert }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)
  const [offen, setzeOffen] = useState('')
  const [name, setzeName] = useState('')

  // **Ein Vorgang je Datensatz** (§7, Regel 4).
  const vorgang = useVorgang()

  // fuehreAus ist der gemeinsame Rumpf: laufen lassen, Fehler zeigen, und nur
  // nach einem **geglückten** Ruf neu laden (§7, Regel 3).
  const fuehreAus = async (schluessel: string, arbeit: () => Promise<void>) => {
    setzeFehler(undefined)

    await vorgang.starte(schluessel, async () => {
      try {
        await arbeit()
        setzeOffen('')
        setzeName('')
        geaendert()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })
  }

  // umbenennenNutzer setzt **nur** den Namen — und behält alles andere.
  //
  // # Warum die übrigen Felder mitgehen
  //
  // **Weil `NutzerAendern` setzt und nicht verschmilzt.** Ein Aufruf mit
  // leeren Feldern löschte Vorname, Slogan und Biografie — und niemand hätte
  // das verlangt.
  //
  // **Und weil die Zeile sie ohnehin hat**: sie kommt aus derselben Liste.
  // Ein Nachschlagen hier wäre ein zweiter Stand und ein Zweig, den kein
  // Klick betreten kann (§3, §7 Regel 2).
  const umbenennenNutzer = (einer: NutzerDTO, neuerName: string) =>
    fuehreAus('n' + einer.id, async () => {
      const antwort = await api().NutzerAendern({
        ...angabenAus(einer),
        name: neuerName,
      })

      auspacken(antwort, () => antwort.nutzer)
    })

  // umbenennenProfil setzt **nur** den Namen — und behält alles andere.
  //
  // **Dieselbe Überlegung wie beim Nutzer** (§7, Regel 2 und der Satz „setzt
  // und verschmilzt nicht"): seit Werkstück 9 trägt ein Profil Subslogan,
  // zwei Beschreibungen und ein Abzeichen. Ein Aufruf mit leeren Feldern
  // löschte sie — und niemand hätte das verlangt. Die Werte kommen aus
  // derselben Zeile, die dasteht.
  const umbenennenProfil = (eines: ProfilDTO, neuerName: string) =>
    fuehreAus('p' + eines.id, async () => {
      const antwort = await api().ProfilAendern({
        ...profilangabenAus(eines),
        name: neuerName,
      })

      auspacken(antwort, () => antwort.profil)
    })

  const loeschen = (schluessel: string, bestaetigt: string) =>
    fuehreAus(schluessel, async () => {
      const antwort = schluessel.startsWith('n')
        ? await api().NutzerLoeschen(schluessel.slice(1), bestaetigt)
        : await api().ProfilLoeschen(schluessel.slice(1), bestaetigt)

      auspacken(antwort, () => undefined)
    })

  return (
    <section aria-label={eigene.pflege} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß. */}
      <Erklaerungsabsatz
        text={eigene.pflegeErklaerung}
        klasse="max-w-prose text-leise"
      />

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}

      <ul className="flex max-w-xl flex-col gap-3">
        {nutzer.flatMap((einer) => [
          <Zeile
            key={'n' + einer.id}
            schluessel={'n' + einer.id}
            beschriftung={einer.name}
            offen={offen}
            name={name}
            setzeOffen={setzeOffen}
            setzeName={setzeName}
            umbenennen={(neuerName) => umbenennenNutzer(einer, neuerName)}
            loeschen={loeschen}
            laeuft={vorgang.laeuft('n' + einer.id)}
            bild={bilder[einer.id] ?? ''}
            /* **Das Bild hängt am Nutzer und nicht am Profil** — ein Mensch
               sieht in jeder Arbeitsumgebung gleich aus. */
            bildwahl={<Nutzerbildwahl nutzer={einer} geaendert={geaendert} />}
          />,
          ...einer.profile.map((eines) => (
            <Zeile
              key={'p' + eines.id}
              schluessel={'p' + eines.id}
              beschriftung={eines.name}
              eingerueckt
              offen={offen}
              name={name}
              setzeOffen={setzeOffen}
              setzeName={setzeName}
              umbenennen={(neuerName) => umbenennenProfil(eines, neuerName)}
              loeschen={loeschen}
              laeuft={vorgang.laeuft('p' + eines.id)}
            />
          )),
        ])}
      </ul>
    </section>
  )
}

interface Zeileneigenschaften {
  schluessel: string
  beschriftung: string

  // bild ist der Daten-URI des Profilbildes — leer heißt: keines. Nur ein
  // Nutzer hat eines.
  bild?: string

  // bildwahl sind die beiden Handgriffe daran — nur beim Nutzer.
  bildwahl?: React.ReactNode
  eingerueckt?: boolean
  offen: string
  name: string
  setzeOffen: (schluessel: string) => void
  setzeName: (name: string) => void
  umbenennen: (name: string) => Promise<void>
  loeschen: (schluessel: string, bestaetigt: string) => Promise<void>
  laeuft: boolean
}

// Zeile ist ein Nutzer oder ein Profil samt seinen beiden Handlungen.
function Zeile({
  schluessel,
  beschriftung,
  bild = '',
  bildwahl,
  eingerueckt = false,
  offen,
  name,
  setzeOffen,
  setzeName,
  umbenennen,
  loeschen,
  laeuft,
}: Zeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const istNutzer = schluessel.startsWith('n')

  // **Gelöscht wird erst nach einer Frage** — UX-10.AK-04 (#1682). Der Name
  // im Feld hält den Knopf an, wenn er nicht mehr stimmt; das war bis zum
  // 2026-09-25 die einzige Sperre vor dem Löschen eines Nutzers samt seiner
  // Profile.
  const [fragt, setzeFragt] = useState(false)

  // **Eine geschlossene Zeile räumt ihre Frage mit** (§7, Regel 1) — sonst
  // stünde sie beim nächsten Öffnen schon da.
  if (fragt && offen !== schluessel) {
    setzeFragt(false)
  }

  if (offen === schluessel) {
    return (
      <li className={eingerueckt ? 'ml-6' : ''}>
        <div className="flex flex-wrap items-end gap-3">
          <Pflichtfeld
            beschriftung={eigene.neuerNameFuer}
            wert={name}
            geaendert={setzeName}
          />

          <Aktionsknopf
            aktion="bearbeiten"
            gewicht="haupt"
            angehalten={laeuft || name.trim() === ''}
            aufKlick={() => {
              void umbenennen(name)
            }}
          >
            {eigene.umbenennen}
          </Aktionsknopf>

          <Aktionsknopf
            aktion="loeschen"
            angehalten={laeuft || name !== beschriftung}
            bezeichnung={einsetzen(eigene.loeschenBestaetigt, {
              name: beschriftung,
            })}
            aufKlick={() => {
              setzeFragt(true)
            }}
          >
            {eigene.loeschen}
          </Aktionsknopf>

          <Knopf
            aufKlick={() => {
              setzeOffen('')
              setzeName('')
            }}
          >
            {eigene.abbrechen}
          </Knopf>
        </div>

        {bildwahl !== undefined && (
          <div className="mt-2 flex items-center gap-3">
            <Nutzerbild name={beschriftung} datenURI={bild} />

            {bildwahl}
          </div>
        )}

        {/* **Der Hinweis steht in der Frage, sobald sie gestellt ist** —
            zweimal derselbe Satz untereinander wäre einer zu viel. */}
        {fragt && name === beschriftung ? (
          <div className="mt-2">
            <Entscheidung
              frage={einsetzen(eigene.loeschenFrage, { name: beschriftung })}
              erklaerung={
                istNutzer
                  ? eigene.nutzerLoeschenHinweis
                  : eigene.profilLoeschenHinweis
              }
              bestaetigen={eigene.loeschenJa}
              aktion="loeschen"
              abbrechen={eigene.abbrechen}
              laeuft={laeuft}
              aufBestaetigen={() => {
                setzeFragt(false)
                void loeschen(schluessel, name)
              }}
              aufAbbrechen={() => {
                setzeFragt(false)
              }}
            />
          </div>
        ) : (
          <p className="mt-1 max-w-prose text-sm text-leise">
            {istNutzer
              ? eigene.nutzerLoeschenHinweis
              : eigene.profilLoeschenHinweis}
          </p>
        )}
      </li>
    )
  }

  return (
    <li className={eingerueckt ? 'ml-6' : ''}>
      <div className="flex items-center gap-3">
        <span className="min-w-40">{beschriftung}</span>

        <Knopf
          gewicht="leise"
          bezeichnung={einsetzen(eigene.bearbeiten, { name: beschriftung })}
          aufKlick={() => {
            setzeOffen(schluessel)
            setzeName(beschriftung)
          }}
        >
          {eigene.bearbeitenKurz}
        </Knopf>
      </div>
    </li>
  )
}

// profilangabenAus liest die Angaben aus einem Profil der Liste.
//
// **Dieselbe Bauform wie [angabenAus] daneben** — und aus demselben Grund:
// `ProfilAendern` setzt und verschmilzt nicht.
export function profilangabenAus(eines: ProfilDTO): ProfilangabenDTO {
  return {
    id: eines.id,
    name: eines.name,
    subslogan: eines.subslogan,
    kurzebeschreibung: eines.kurzebeschreibung,
    langebeschreibung: eines.langebeschreibung,
    badge: eines.badge,
    technisch: eines.technisch,
    ablauf: eines.ablauf,
    fassung: eines.fassung,
  }
}

// angabenAus liest die Angaben aus einem Nutzer der Liste.
export function angabenAus(einer: NutzerDTO): NutzerangabenDTO {
  return {
    id: einer.id,
    name: einer.name,
    bild: einer.bild,
    anrede: einer.anrede,
    namenszusatz: einer.namenszusatz,
    vorname: einer.vorname,
    weiterenamen: einer.weiterenamen,
    nachname: einer.nachname,
    slogan: einer.slogan,
    kurzebio: einer.kurzebio,
    langebio: einer.langebio,
    fassung: einer.fassung,
  }
}
