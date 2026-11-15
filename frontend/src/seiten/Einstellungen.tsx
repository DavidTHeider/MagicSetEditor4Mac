import { useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'
// Die Einstellungen: Sprache, Erscheinungsbild, Schriftgröße — und die
// Tastatur (#245).
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Was sich mit #245 geändert hat
//
// Vorher führte die Seite die **Sprache** und sonst nichts; darunter stand ein
// Absatz, der erklärte, dass das Erscheinungsbild der Systemeinstellung folgt.
// Das war keine Einstellung, sondern eine Entschuldigung.
//
// **„Eine Einstellung, die es nicht gibt, ist eine Entscheidung, die jemand
// anderes getroffen hat."** Wer die Anwendung hell benutzen will, während sein
// System dunkel steht, kann es jetzt; wer größere Schrift braucht, auch.
//
// # Die Sprachwahl wirkt auch auf die native Menüleiste
//
// Sie liegt in Go und weiß von React nichts; ohne den Aufruf bliebe sie bei
// der Sprache, mit der die Anwendung gestartet ist — eine Anwendung in zwei
// Sprachen.
import {
  dienstlisten,
  einstellungsgruppen,
  gruppeFuer,
} from './einstellungsgruppen'

import type { Einstellungsgruppe as Gruppe } from './einstellungsgruppen'

import type { EinstellungDTO } from '../api'
import type { Bestandspflege as Bestandspflegestand } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'
import { meldungAus } from '../meldung'
import { Ohnetreffer } from '../bausteine/Ohnetreffer'
import { useTexte } from '../sprache/kontext'
import { Gruppenliste } from '../bausteine/Gruppenliste'
import { Suchfeld } from '../bausteine/Einstellungssuche'
import {
  abschnitteJeGruppeAus,
  beiwerkJeGruppe,
  fachbeiwerkJeGruppe,
  selbstgezeichneteSchluessel,
} from './einstellungsbeiwerk'
import { Einstellungsgruppe } from './einstellungsgebiete'
import {
  aufschlagAus,
  ausserAbschnitten,
  mitAufschlag,
  type Aufschlag,
} from './einstellungsabschnitte'
import { einsetzen } from '../sprache/einsetzen'
import { useEinstellungssuche } from '../haken/useEinstellungssuche'

interface EinstellungenEigenschaften {
  eintraege: EinstellungDTO[]

  // dienste sind die Schalter der fremden Kartendienste — Phase 4.8
  // (#1321).
  //
  // **Getrennt und nicht in `eintraege`**: sie stehen in einem eigenen
  // Bereich mit einer eigenen Auskunft (Vermerk 0063).
  dienste: EinstellungDTO[]

  // fragen sind die Wahlen je Frage — Werkstück 8 von Phase 4.8.
  fragen: EinstellungDTO[]

  // ki sind die zwei Zeilen der lokalen KI-Laufzeit — A7, #1446.
  ki: EinstellungDTO[]

  // vorgaben sind die, mit denen ein neuer Datensatz anfängt — Werkstück 9
  // von Phase 4.12 (#1351).
  //
  // **Getrennt und nicht in `eintraege`**: sie stehen in einem eigenen
  // Gebiet, weil sie eine andere Frage beantworten.
  vorgaben: EinstellungDTO[]

  // anmeldung sind die Kennungen, unter denen sich diese Anwendung bei einem
  // Anbieter ausweist — Nacharbeit zu Werkstück 4 von Phase 4.12 (#1351).
  //
  // **Getrennt und nicht in `eintraege`**: sie stehen in einem eigenen
  // Gebiet, und der Satz daneben erklärt, was man beim Anbieter tun muss,
  // bevor eine Kennung überhaupt entsteht.
  anmeldung: EinstellungDTO[]

  // alleEinstellungen ist, was die **Suche** durchsucht — UX-04.AK-03
  // (#1676),
  // zugleich NET-02.NAT-03 (#1706).
  //
  // # Warum das nicht dieselbe Liste ist wie das, was die Seite zeigt
  //
  // **Weil Zeichnen und Suchen zwei Fragen sind.** Gezeigt wird jede
  // Einstellung **genau einmal** — die sechs Listen darüber sind dafür zu
  // Recht eng, und eine siebte Zeile wäre eine Dopplung (#1768). Gesucht
  // wird in **allem, was dieses Fenster kennt**: `netz.proxy.http` zeichnet
  // `Netzweg` selbst, und die Suche fand ihn deshalb nie — „Keine
  // Einstellung passt zu „proxy"", gemessen am 2026-09-21.
  //
  // **Wahlweise, und der Rückfall ist die gezeigte Liste**: ein Fenster,
  // dem niemand die volle Liste gibt, hat dann keine leere Suche, sondern
  // die alte.
  alleEinstellungen?: readonly EinstellungDTO[] | undefined

  // ursache ist ein gescheiterter Zugriff auf den Bestand — **roh**.
  //
  // Der Satz dazu entsteht hier und nicht im Haken: der läuft **über** dem
  // `Sprachraum` (`App` stellt ihn her) und bekäme dort den Vorgabekatalog —
  // ein Fehler stünde in der englischen Oberfläche auf Deutsch.
  ursache: unknown

  // meldefehler ist ein gescheiterter Aufruf an die native Menüleiste.
  //
  // **Er kommt von außen**, weil auch der Aufruf von außen kommt: gemeldet wird
  // beim Start und bei jedem Wechsel, und beides führt die Anwendung, nicht
  // diese Seite. Angezeigt gehört er trotzdem hier — an der Stelle, an der
  // jemand die Sprache wählt.
  meldefehler: string | undefined

  setze: (schluessel: string, wert: string) => void
  laeuft: (schluessel: string) => boolean

  // bestandspflege ist das Sichern und Zurückspielen des Bestands (#940).
  //
  // **Hier und nicht auf einer eigenen Seite**: es ist eine Einstellung im
  // Wortsinn nicht, aber es gehört zu dem, was für die ganze Anwendung gilt —
  // und eine zehnte Seite für zwei Knöpfe wäre eine Navigation, die niemand
  // durchsucht.
  bestandspflege: Bestandspflegestand

  // ruecksprung sind die vier Rückkehrarten und die verbuchten Rückkehren
  // (O7, #1431) — **neben** der Bestandspflege, weil es eine andere Frage
  // ist: jene sichert und spielt zurück, diese sagt, was zurückgehen kann
  // und was eine Rückkehr hinterlassen hat.
  ruecksprung: Ruecksprungstand

  // pflege ist das Aufräumen verwaister Zeilen (O8, #1432) — **neben** der
  // Rückkehr, weil es eine andere Frage ist: jene geht zurück, diese räumt
  // auf, was ohnehin auf nichts mehr zeigt.
  pflege: Pflegelauf

  // setID ist das offene Set — es steht in den Befehlszeilen des technischen
  // Bereichs, damit niemand eine Kennung abtippen muss (#1164).
  setID: string | undefined
}

export function Einstellungen({
  eintraege,
  dienste,
  fragen,
  ki,
  vorgaben,
  anmeldung,
  ursache,
  meldefehler,
  setze,
  laeuft,
  bestandspflege,
  ruecksprung,
  pflege,
  setID,
  alleEinstellungen,
}: EinstellungenEigenschaften) {
  const texte = useTexte()

  // **Ein Wunsch, ein Gebiet zu öffnen** (#1422): die Suche nennt das Gebiet
  // eines Treffers, und ein Knopf führt hin — mit Nummer, damit derselbe Weg
  // zweimal wirkt (siehe `Laschen.oeffne`).
  const [oeffne, setzeOeffne] = useState<string | undefined>(undefined)
  const [wuensche, setzeWuensche] = useState(0)

  // **Die Frage steht hier und nicht im Suchfeld** — sie wird an zwei
  // Stellen gebraucht: das Feld steht in der Spalte, die Treffer in der
  // Fläche (UX-04.AK-01).
  const [frage, setzeFrage] = useState('')

  // **Welche Gruppe offen ist**, gemeldet von der [Gruppenliste] — die Seite
  // braucht es, um zu entscheiden, ob sie bei einer neuen Frage umschaltet.
  const [offeneGruppe, setzeOffeneGruppe] = useState<Gruppe>(
    einstellungsgruppen[0],
  )
  // **Eine Liste und sechs Quellen.** Die Seite bekommt die Einstellungen
  // nach Bereich getrennt; der Entwurf ordnet sie nach **Sache**. Wer sie
  // nicht zusammenlegte, könnte `sprache` und `erscheinungsbild` nicht in
  // zwei Gruppen stellen — sie liegen im selben Bereich.
  //
  // **Und `alleEinstellungen` geht vor.** Am 2026-09-21 am laufenden
  // Fenster gemessen: von 51 Einstellungen standen **32** da. Die sechs
  // `netz.*`, vier `arbeitsplatz.*`, sieben `richtlinie.*`, `konto.schau`
  // und `bildkonflikt` erreichten die Seite nie — gebaut, geprüft, mit
  // Ebene, Herkunft, Vorgabe und Wirkzeitpunkt, und **nirgends zu
  // ändern**. `Netzweg` sagt sogar wörtlich, man ändere den Proxy hier.
  const gezeigte = alleEinstellungen ?? [
    ...eintraege,
    ...dienste,
    ...fragen,
    ...ki,
    ...vorgaben,
    ...anmeldung,
  ]

  // **Die Kartendienste und die Sprachmodelle sind zwei Listen** — beide
  // heißen `dienst.`; geteilt wird in `einstellungsgruppen.ts`.
  // **Ohne die Adressen** — sie zeichnet ein Abschnitt (UXR-21, #1929).
  const { kimodelle, kartendienste } = dienstlisten(ausserAbschnitten(dienste))
  const kiOffen = ausserAbschnitten(ki)

  // **Und ohne das, was ein Bauteil der Gruppe selbst zeichnet.**
  //
  // # Der Anlass
  //
  // **Gefunden in der Sichtprüfung am 2026-09-21**: nach dem Umbau auf neun
  // Gruppen standen die fünf Kartendienste **zweimal** auf der Seite —
  // einmal als `dienst.scryfall` mit rohem Schlüssel als Aufschrift, einmal
  // richtig benannt als *Scryfall*. In *KI* standen so die Sprachmodelle
  // und vier Adressen. Ein `Dienstschalter` weiß, wie ein Dienst heißt
  // (`bausteine/dienstnamen.ts`); die gewöhnliche Wahl schlägt den Namen im
  // Katalog nach und findet keinen.
  //
  // # Warum die Liste hier entsteht und nicht danebensteht
  //
  // **Weil sie sonst von Hand geführt wäre.** Ein erster Anlauf zählte die
  // Schlüssel in `einstellungsgruppen.ts` auf — und vergaß die vier
  // Adressen, die ein `Dienstschalter` mitzeichnet. Es sind **dieselben
  // Listen**, die zwei Zeilen weiter unten übergeben werden; wer eine
  // ändert, ändert beide.
  //
  // **Und die Belegung der Ziffern** (UX-02.AK-03, #1674): sie steht als
  // Tafel bei den Kürzeln, nicht als Textfeld.
  const selbstgezeichnet = new Set([
    ...[...kimodelle, ...kartendienste, ...fragen, ...kiOffen].map(
      (eintrag) => eintrag.schluessel,
    ),
    ...selbstgezeichneteSchluessel,
  ])

  // **Und die Schalter zeichnet `Bereichsschalter`** (OPS-10, #1908).
  const inGruppe = (gruppe: Gruppe) =>
    gezeigte.filter(
      (eintrag) =>
        gruppe !== 'bereiche' &&
        gruppeFuer(eintrag.schluessel) === gruppe &&
        !selbstgezeichnet.has(eintrag.schluessel),
    )

  // **Die Treffer je Gruppe** (UX-04.AK-03, #1676) — einmal gerechnet in
  // `useEinstellungssuche`, gelesen von Spalte und Fläche.
  //
  // zumGebiet schlägt ein Gebiet auf — die Hilfe und die Suche gehen
  // darüber.
  const zumGebiet = (schluessel: string) => {
    setzeWuensche(wuensche + 1)
    setzeOeffne(`${schluessel}#${String(wuensche + 1)}`)
  }

  // **Und die Befehlspalette** — „Sicherung erstellen" öffnet *Daten und
  // Sicherung* (UX-02.AK-02, #1674); mit Abschnitt schlägt sie ihn auf, bis
  // eine andere Gruppe aufgeht (AP-10, #2367).
  const [aufschlag, setzeAufschlag] = useState<Aufschlag>()
  const anschluss = useAnschluss()
  useAufgenommen(
    anschluss.offen?.art === 'einstellungsgruppe' ? anschluss.offen : undefined,
    (gewuenscht) => {
      zumGebiet(gewuenscht.gruppe)
      setzeAufschlag(aufschlagAus(gewuenscht, wuensche + 1))
    },
  )

  const { trefferIn, ersteMitTreffer, aufSprung } = useEinstellungssuche({
    gezeigte,
    frage,
    texte,
    gruppen: einstellungsgruppen,
    gruppeFuer,
    offeneGruppe,
    zumGebiet,
  })

  const teile = {
    kimodelle,
    ki: kiOffen,
    kartendienste,
    fragen,
    laeuft,
    setze,
    bestandspflege,
    ruecksprung,
    pflege,
    setID,
    eintraege: gezeigte,
  }
  const beiwerk = beiwerkJeGruppe(teile)
  const abschnitte = abschnitteJeGruppeAus(teile)

  return (
    <>
      <h1 className="text-2xl font-semibold tracking-tight">
        {texte.seiten.einstellungen.titel}
      </h1>

      {ursache !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {meldungAus(ursache, texte)}
        </p>
      )}

      {/* **Neun Gruppen als Liste, wie der Entwurf sie zeichnet** —
          UX-04.AK-01
          (#1676).
          Welcher Schlüssel in welche gehört, steht in
          `einstellungsgruppen.ts`; hier steht, **was** außer Einstellungen
          noch in einer Gruppe liegt.

          **Die Suche steht in derselben Spalte über der Liste**, nicht über
          der ganzen Seite: der Entwurf setzt sie dorthin, und sie führt in
          eine Gruppe — also gehört sie neben die Gruppen. */}
      <Gruppenliste
        bezeichnung={texte.seiten.hilfe.aufDieserSeite}
        oeffne={oeffne}
        aufWechsel={(schluessel) => {
          setzeOffeneGruppe(schluessel as Gruppe)
          if (schluessel !== aufschlag?.gruppe) setzeAufschlag(undefined)
        }}
        darueber={
          <Suchfeld
            frage={frage}
            setzeFrage={setzeFrage}
            aufSprung={aufSprung}
          />
        }
        gebiete={einstellungsgruppen.map((gruppe) => ({
          schluessel: gruppe,
          name: texte.seiten.einstellungen.gruppen[gruppe],
          unter: texte.seiten.einstellungen.untertitel[gruppe],
          marke:
            trefferIn(gruppe).size > 0 ? (
              <span
                className="rounded-full border border-marke px-1.5 text-xs text-marke"
                aria-hidden="true"
              >
                {String(trefferIn(gruppe).size)}
              </span>
            ) : undefined,
          inhalt: (
            <Einstellungsgruppe
              name={texte.seiten.einstellungen.gruppen[gruppe]}
              treffer={trefferIn(gruppe)}
              /* **Die Zahl steht im Kopf der Fläche**, wie der Entwurf sie
                 zeichnet — und nur, wenn es welche gibt. */
              kopfmarke={
                trefferIn(gruppe).size > 0 ? (
                  <span className="rounded-full border border-marke px-2 py-0.5 text-xs font-normal text-marke">
                    {einsetzen(texte.seiten.einstellungen.trefferInGruppe, {
                      zahl: String(trefferIn(gruppe).size),
                      frage: frage.trim(),
                    })}
                  </span>
                ) : undefined
              }
              /* **Und wo gar nichts passt, steht der Satz** — eine leere
                 Fläche wäre keine Antwort. */
              vorne={
                frage.trim() !== '' && ersteMitTreffer === undefined ? (
                  <Ohnetreffer
                    satz={einsetzen(texte.seiten.einstellungen.sucheKeine, {
                      frage: frage.trim(),
                    })}
                    leeren={() => {
                      setzeFrage('')
                    }}
                  />
                ) : undefined
              }
              eintraege={inGruppe(gruppe)}
              laeuft={laeuft}
              setze={setze}
              fach={fachbeiwerkJeGruppe[gruppe]}
              abschnitte={mitAufschlag(abschnitte[gruppe], gruppe, aufschlag)}
            >
              {beiwerk[gruppe]}
            </Einstellungsgruppe>
          ),
        }))}
      />

      {/* **Ein gescheiterter Aufruf ist zu sehen.** Die Oberfläche steht dann
          in der neuen Sprache und die Menüleiste in der alten — das ist ein
          Zustand, den man erklärt bekommt, statt ihn zu suchen.

          **Er steht unter der Liste und nicht in einem Gebiet**: er kann
          jedes von ihnen betreffen. */}
      {meldefehler !== undefined && (
        <p role="alert" className="mt-4 text-fehler">
          {texte.bausteine.zustaende.fehler} {meldefehler}
        </p>
      )}
    </>
  )
}
