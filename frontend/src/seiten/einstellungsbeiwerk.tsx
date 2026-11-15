// **Was außer Einstellungen in einer Gruppe steht** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei seit dem 2026-09-21**, als `Einstellungen.tsx` mit der
// Trefferanzeige über die Zeilengrenze aus §2 ging. Der Schnitt folgt der
// Frage: dort steht, **wie die Seite aussieht**; nebenan in
// `einstellungsgebiete.tsx`, **wie eine Gruppe aussieht**; hier, **was in
// welcher liegt**.
//
// **Ein vollständiger Verbund**: wer eine Gruppe aufnimmt, entscheidet hier,
// was darin liegt — `null` heißt „nur die Einstellungen", und das ist eine
// Aussage und kein Vergessen.
import { Anmeldegebiet } from './einstellungsgebiete'
import { Bestandseingriffe, Bestandsgebiet } from '../bausteine/Bestandsgebiet'
import { Bereichsschalter } from '../bausteine/Bereichsschalter'
import {
  Diensteeinstellungen,
  KIeinstellungen,
} from '../bausteine/Diensteeinstellungen'
import { Eigenuebersetzungen } from '../bausteine/Eigenuebersetzungen'
import { Einstellungsaustausch } from '../bausteine/Einstellungsaustausch'
import {
  Eigenfarben,
  Gestaltungsbeiwerk,
} from '../bausteine/Gestaltungsbeiwerk'
import { Dienstkatalog } from '../bausteine/Dienstkatalog'
import { KIModelle } from '../bausteine/KIModelle'
import { Wirkungszaehler } from '../bausteine/Wirkungszaehler'
import { Kosten } from './Kosten'
import { Kuerzeltafel } from './Kuerzeltafel'
import { belegungAus } from '../rahmen/ziffernbelegung'
import { Netzbereich, Netzfachbereich, Technik, Werkzeuge } from './Technik'
import { Leistungsprofil } from './Leistungsprofil'
import { Suchindex } from './Suchindex'
import { Zwischenspeicher } from './Zwischenspeicher'
import { Zugaenge } from '../bausteine/Zugaenge'

import type { ReactNode } from 'react'

import { einstellungsgruppen } from './einstellungsgruppen'
import { abschnitteJeGruppe } from './einstellungsabschnitte'

import type { EinstellungDTO } from '../api'
import type { Einstellungsgruppe as Gruppe } from './einstellungsgruppen'
import type { Abschnittsname } from './einstellungsabschnitte'
import type { Einstellungsabschnitt } from './einstellungsgebiete'
import type { Bestandspflege as Bestandspflegestand } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'

// Beiwerksteile ist, was das Beiwerk der Gruppen und Abschnitte braucht.
export interface Beiwerksteile {
  kimodelle: EinstellungDTO[]
  ki: EinstellungDTO[]
  kartendienste: EinstellungDTO[]
  fragen: EinstellungDTO[]
  laeuft: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void
  bestandspflege: Bestandspflegestand
  ruecksprung: Ruecksprungstand
  pflege: Pflegelauf
  setID: string | undefined

  // eintraege sind alle gezeigten Einstellungen — die Gestaltung (UX-09,
  // #1681) liest daraus ihr Feld und die Stellung des Fensterbilds, die
  // Kürzeltafel ihre Belegung (UX-02.AK-03, #1674).
  eintraege?: readonly EinstellungDTO[]
}

export function beiwerkJeGruppe({
  kimodelle,
  ki,
  kartendienste,
  fragen,
  laeuft,
  setze,
  bestandspflege,
  eintraege = [],
}: Beiwerksteile): Record<Gruppe, ReactNode> {
  return {
    // **Allgemein zeichnet nur seine Einstellungen offen** — eigene
    // Übersetzungen und Kürzel stehen seit UXR-19 (#1927) in Abschnitten.
    allgemein: null,
    // **Die Schalter zeichnet ihr eigenes Bauteil** — OPS-10 (#1908).
    bereiche: (
      <Bereichsschalter
        eintraege={eintraege}
        angehalten={laeuft}
        setze={setze}
      />
    ),
    erscheinungsbild: (
      <Gestaltungsbeiwerk
        hintergrund={
          eintraege.find((eintrag) => eintrag.schluessel === 'hintergrund')
            ?.wert ?? 'aus'
        }
      />
    ),
    nutzer: <Anmeldegebiet />,
    fach: null,
    ki: (
      <KIeinstellungen
        modelle={kimodelle}
        ki={ki}
        angehalten={laeuft}
        setze={setze}
      />
    ),
    verbindungen: <Netzbereich />,
    // **Katalog und Zugangsdaten stehen seit #2465 in Abschnitten**, wie
    // in *Daten und Sicherung* die Eingriffe, die Kosten und das Mitnehmen.
    dienste: (
      <Diensteeinstellungen
        dienste={kartendienste}
        fragen={fragen}
        angehalten={laeuft}
        setze={setze}
      />
    ),
    daten: <Bestandsgebiet bestandspflege={bestandspflege} />,
    diagnose: <Technik />,
  }
}

// abschnitteJeGruppeAus gibt je Gruppe ihre Abschnitte samt Beiwerk —
// UXR-19 (#1927).
// **Welche Schlüssel** in einem stehen, sagt `einstellungsabschnitte.ts`;
// hier steht, **was außer ihnen** darin liegt.
export function abschnitteJeGruppeAus({
  kartendienste,
  fragen,
  laeuft,
  setze,
  bestandspflege,
  ruecksprung,
  pflege,
  setID,
  eintraege = [],
}: Beiwerksteile): Record<Gruppe, Einstellungsabschnitt[]> {
  const ziffern = eintraege.find(
    (eintrag) => eintrag.schluessel === ziffernschluessel,
  )

  const beiwerk: Partial<Record<Abschnittsname, ReactNode>> = {
    // **Die eigenen Übersetzungen** — LOC-02 (#1688).
    eigeneUebersetzungen: (
      <Eigenuebersetzungen
        eingebettet
        eintrag={eintraege.find(
          (eintrag) => eintrag.schluessel === eigenesprachschluessel,
        )}
        angehalten={laeuft(eigenesprachschluessel)}
        setze={(wert) => {
          setze(eigenesprachschluessel, wert)
        }}
      />
    ),
    // **Fehlt die Belegung**, wie in Läufen mit einer kurzen Liste, bleibt
    // die Tafel eine Anzeige.
    kuerzel:
      ziffern === undefined ? (
        <Kuerzeltafel eingebettet />
      ) : (
        <Kuerzeltafel
          eingebettet
          belegung={belegungAus(ziffern.wert)}
          eintrag={ziffern}
          angehalten={laeuft(ziffernschluessel)}
          setze={(wert) => {
            setze(ziffernschluessel, wert)
          }}
        />
      ),
    // **Die Diagnose** — Leistungsprofil (OPS-05, #1722), Zwischenspeicher
    // (OPS-08, #1725), Suchindex (A15, #1514) und die Werkzeuge ohne
    // Fenster (#1164, OPS-03, OPS-09) neben ihren Einstellungen.
    messung: <Leistungsprofil />,
    speicher: (
      <>
        <Zwischenspeicher />
        <Suchindex />
      </>
    ),
    werkzeuge: <Werkzeuge setID={setID} />,
    // **Was #2465 eingeklappt hat.** Der Katalog lädt an einem Stand aus
    // Schaltern und Wahlen (#1440): legt jemand einen um, lädt er nach —
    // eine Bereitschaft, die den Schalter nicht sieht, wäre eine Behauptung.
    dienstkatalog: (
      <Dienstkatalog
        stand={[...kartendienste, ...fragen]
          .map((eintrag) => eintrag.schluessel + eintrag.wirksam)
          .join()}
      />
    ),
    zugaenge: <Zugaenge />,
    eingriffe: (
      <Bestandseingriffe
        bestandspflege={bestandspflege}
        ruecksprung={ruecksprung}
        pflege={pflege}
      />
    ),
    // **Was ein Auftrag draußen kosten darf** (O12, #1439) steht bei den
    // Daten und nicht bei den Diensten: gefragt ist nicht, **welcher** Dienst
    // an ist, sondern was er **kosten** darf — eine Buchführung über diesen
    // Bestand, wie die Sicherung.
    kosten: <Kosten />,
    mitnehmen: <Einstellungsaustausch />,
    kimodelle: <KIModelle />,
    eigeneFarben: (
      <Eigenfarben
        gestaltung={eintraege.find(
          (eintrag) => eintrag.schluessel === gestaltungsschluessel,
        )}
        angehalten={laeuft(gestaltungsschluessel)}
        setze={(wert) => {
          setze(gestaltungsschluessel, wert)
        }}
      />
    ),
    wirkung: <Wirkungszaehler />,
  }

  return Object.fromEntries(
    einstellungsgruppen.map((gruppe) => [
      gruppe,
      (abschnitteJeGruppe[gruppe] ?? []).map((plan) => ({
        ...plan,
        beiwerk: beiwerk[plan.name],
      })),
    ]),
  ) as Record<Gruppe, Einstellungsabschnitt[]>
}

// fachbeiwerkJeGruppe ist, was außer den Einstellungen im eingeklappten
// Kasten für Fachleute steht (UX-04.AK-04, #1676).
export const fachbeiwerkJeGruppe: Readonly<Partial<Record<Gruppe, ReactNode>>> =
  { verbindungen: <Netzfachbereich /> }

// ziffernschluessel ist der Schlüssel der Belegung — aus der Domäne
// (`domain.EinstellungZiffernbelegung`).
export const ziffernschluessel = 'kuerzel.ziffern'

// gestaltungsschluessel sind die eigenen Merkmale — ein Feld mit mehreren
// Zeilen, das das Beiwerk zeichnet (`domain.EinstellungGestaltung`).
export const gestaltungsschluessel = 'gestaltung'

// eigenesprachschluessel ist die eigene Übersetzung, die gilt — die Wahl
// steht über ihrer Liste (`domain.EinstellungEigeneSprache`, LOC-02).
export const eigenesprachschluessel = 'sprache.eigene'

// planordnerschluessel ist der Ordner des freiwilligen Plans — gezeichnet von
// der Fläche „Letzte Sicherung" mit Kästchen und Dialog
// (`domain.SicherungsplanOrdner`, OPS01, #2347).
export const planordnerschluessel = 'sicherung.planordner'

// selbstgezeichnet sind die Schlüssel, die ein Bauteil zeichnet und nicht die
// gewöhnliche Liste — eine Liste, damit die Seite sie nicht einzeln aufzählt.
export const selbstgezeichneteSchluessel = [
  ziffernschluessel,
  gestaltungsschluessel,
  eigenesprachschluessel,
  planordnerschluessel,
]
