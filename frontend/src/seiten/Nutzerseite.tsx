// **Der Ladeweg der Nutzerseite** — Phase 4.12
// (#1351).
//
// **Was sie tut, steht hier; wie sie aussieht, in `Nutzer.tsx`** (§2).
//
// # Der Kopf führt
//
// **Seit UXR-20** (#1928)
// steht *+ Profil anlegen …* als einziger Hauptknopf im Kopf, wie im Artboard
// `Nutzer`, und öffnet eine Schublade. Vorher stand das Anlegen am Fuß der
// Lasche *Profile* — unter Vorgaben, Freigaben und Token, mit realistischem
// Bestand die 36 Felder einer Formularwand weiter unten.
//
// **Die Schublade wird beim ersten Öffnen gezeichnet und danach nur
// verborgen** — wie die der Pakete: ein halb getippter Name übersteht das
// Schließen.
import { Activity, useRef, useState } from 'react'

import { Nutzer } from './Nutzer'
import { Profilanlage } from './Profilanlage'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Schublade } from '../bausteine/Schublade'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { useNutzer } from '../haken/useNutzer'
import { useTexte } from '../sprache/kontext'
import { Verdeckung, useVerdeckt } from '../verdeckung'

const sichtbar = (ja: boolean) => (ja ? 'visible' : 'hidden')

// bestandsstand zählt fremde Schreibvorgänge (#156) — Befund 4 des Nachlaufs
// zum Gesamtreview über 0.6.0.
export function Nutzerseite({ bestandsstand = 0 }: { bestandsstand?: number }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer
  const {
    nutzer,
    gezeigt,
    anbieter,
    anbieterstaende,
    konten,
    eigner,
    arbeitet,
    bilder,
    badges,
    laedt,
    fehler,
    neuLaden,
  } = useNutzer(bestandsstand)

  const [offen, setzeOffen] = useState(false)
  const [gezeichnet, setzeGezeichnet] = useState(false)
  const ausloeser = useRef<HTMLButtonElement>(null)

  // **Eine Seite unter einer Schublade ist verdeckt** (`verdeckung.ts`) —
  // und liegt sie schon unter einer, bleibt sie es.
  const verdeckt = useVerdeckt() || offen

  // **Nur beim ersten Laden.** Ein Nachladen zeigt weiter, was dasteht —
  // sonst blinkt die Seite bei jeder Änderung.
  const bereit = (!laedt || nutzer.length > 0) && fehler === undefined

  // **Die Überschrift steht über dem Ladezustand und nicht darunter.**
  //
  // Eine Seite, die beim Laden nur „lädt …" zeigt, hat keinen Titel — und wer
  // sie mit einer Vorlesehilfe öffnet, weiß nicht, wo er ist. Angestrichen
  // vom Wächter über die Überschriftenebenen (#834): er verlangt auf **jeder**
  // Seite genau eine `h1`, und der Ladezustand hatte keine.
  return (
    <>
      <Verdeckung value={verdeckt}>
        <section className="flex flex-col gap-4" inert={offen}>
          {/* **Ein Satz, der Rest hinter dem Hilfesymbol** (UX-15, #1771) —
              der Vorbehalt über der Seite ist ihre Kopferklärung. */}
          <Seitenkopf titel={eigene.titel} erklaerung={eigene.nochNicht}>
            {/* **Der Fokus geht vorher auf den Knopf**, damit er nach dem
                Schließen dorthin zurückkehrt — in WebKit gibt ein Klick
                keinen Fokus. */}
            {bereit && (
              <Aktionsknopf
                aktion="anlegen"
                gewicht="haupt"
                griff={ausloeser}
                aufKlick={() => {
                  ausloeser.current?.focus()
                  setzeGezeichnet(true)
                  setzeOffen(true)
                }}
              >
                {eigene.profilAnlegenOeffnen}
              </Aktionsknopf>
            )}
          </Seitenkopf>

          {laedt && nutzer.length === 0 && (
            <p>{texte.bausteine.zustaende.laedt}</p>
          )}

          {fehler !== undefined && (
            <p role="alert" className="text-fehler">
              {texte.bausteine.zustaende.fehler} {fehler}
            </p>
          )}

          {bereit && (
            <Nutzer
              nutzer={nutzer}
              gezeigt={gezeigt}
              anbieter={anbieter}
              anbieterstaende={anbieterstaende}
              konten={konten}
              eigner={eigner}
              arbeitet={arbeitet}
              bilder={bilder}
              badges={badges}
              neuLaden={neuLaden}
            />
          )}
        </section>
      </Verdeckung>

      {gezeichnet && (
        <Activity mode={sichtbar(offen)}>
          <Schublade
            titel={eigene.neuesProfil}
            schliessenBezeichnung={
              texte.allgemein.arbeitsplatz.schubladen.schliessen
            }
            schliessen={() => {
              setzeOffen(false)
            }}
          >
            <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
              <Profilanlage nutzer={nutzer} angelegt={neuLaden} />
            </div>
          </Schublade>
        </Activity>
      )}
    </>
  )
}
