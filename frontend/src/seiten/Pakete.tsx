// Die Paketseite — Ressourcen › Pakete.
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Die Gestalt seit UXR-17
//
// #1925,
// nach dem Artboard `Pakete`: **der Arbeitsplatz ist die Seite**, und im
// Kopf führt *Paket beziehen …* — es öffnet eine Schublade mit dem Bezug,
// darunter eingeklappt die Übernahme aus MSE2 und der Umschlag. Was selten
// gebraucht wird — anlegen, die Abhängigkeiten, die Sammeldaten —, steht
// eingeklappt unter dem Arbeitsplatz.
//
// **Vorher** standen sechs Abschnitte offen untereinander, mit elf Feldern;
// gemessen am 2026-10-02 an der Hüllenprobe `uirevue` war die Seite 3,3
// Bildschirme hoch (R4 in `docs/UIREVIEW.md` lässt 2,5 zu).
import { Activity, useRef, useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Knopf } from '../bausteine/Knopf'
import { Sammeldaten } from '../bausteine/Sammeldaten'
import { Schublade } from '../bausteine/Schublade'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Vorlagenumschlag } from '../bausteine/Vorlagenumschlag'
import { useTexte } from '../sprache/kontext'
import { Verdeckung, useVerdeckt } from '../verdeckung'
import { useAbhaengigkeiten } from '../haken/useAbhaengigkeiten'
import { Abhaengigkeiten } from './Abhaengigkeiten'
import { Bezugsstaende } from './Bezugsstaende'
import { Paketbezug } from './Paketbezug'
import { Paketliste } from './Paketliste'
import { Paketkennzahlen } from './Paketkennzahlen'
import { usePakete } from '../haken/usePakete'
import { Uebernahme } from './Uebernahme'

import type { Seitenschluessel } from '../navigation'

// **Der Anschluss im Kopf** — UX-02
// (#1674).
//
// Seit die Seitenleiste **Bereiche** führt, hat *Ressourcen* einen Eintrag und
// zwei Seiten. Der Weg zur Werkstatt steht hier, so wie der Entwurf ihn
// zeichnet: `<a class="knopf" href="Werkstatt.dc.html">` im Kopf des Artboards
// *Pakete*.
interface Paketeeigenschaften {
  wechseln: (seite: Seitenschluessel) => void

  // bestandsstand steigt, wenn ein anderer Prozess geschrieben hat (#2011) —
  // ein Paket, das die Kommandozeile bezieht, steht dann auch hier.
  bestandsstand?: number
}

// sichtbar macht aus einem Wahrheitswert die Art einer `<Activity>`.
const sichtbar = (ja: boolean) => (ja ? 'visible' : 'hidden')

export function Pakete({ wechseln, bestandsstand = 0 }: Paketeeigenschaften) {
  const texte = useTexte()
  const t = texte.seiten.pakete

  // **Ein Zähler für die Katalogliste.** Übernahme und Umschlag legen Pakete
  // ins Verzeichnis; ohne diesen Anstoß zeigte die Liste daneben den Stand
  // von vorhin.
  const [runde, setzeRunde] = useState(0)

  // **Die Abhängigkeiten liest die Seite einmal** (U14, #2159) — für die
  // Zahlen an den Zeilen und für den Abschnitt darunter.
  const abhaengigkeiten = useAbhaengigkeiten(runde + bestandsstand)
  // **Die Pakete ebenso** (AP-9) — für die Leiste und die Liste.
  const pakete = usePakete(runde + bestandsstand)
  const nachladen = () => {
    setzeRunde((zahl) => zahl + 1)
  }

  // **Die Schublade wird beim ersten Öffnen gezeichnet und danach nur
  // verborgen** — wie die des Sets (`Setbereich.tsx`). Eine halb getippte
  // Adresse, ein Ergebnis der Übernahme überstehen so das Schließen.
  const [offen, setzeOffen] = useState(false)
  const [gezeichnet, setzeGezeichnet] = useState(false)
  const ausloeser = useRef<HTMLButtonElement>(null)

  // **Eine Seite unter einer Schublade ist verdeckt** (`verdeckung.ts`):
  // ⌥⌘[ gehörte sonst einem Navigator, den niemand sieht. Liegt sie schon
  // unter einer, bleibt sie es.
  const verdeckt = useVerdeckt() || offen

  // beziehen öffnet die Schublade — aus dem Kopf und aus dem Leerzustand der
  // bezogenen Stände (G04). **Der Fokus geht vorher auf den Auslöser**,
  // damit er nach dem Schließen dorthin zurückkehrt — in WebKit gibt ein
  // Klick keinen Fokus.
  const beziehen = () => {
    ausloeser.current?.focus()
    setzeGezeichnet(true)
    setzeOffen(true)
  }

  return (
    <>
      {/* **Gesperrt, solange die Schublade steht**: `inert` nimmt Fokus,
          Klick und Vorleseprogramm, `Verdeckung` die Horcher am Fenster.
          `contents` nimmt den Umschlag aus dem Layout. */}
      <Verdeckung value={verdeckt}>
        <div className="contents" inert={offen}>
          <Seitenkopf titel={t.titel}>
            <Knopf
              aufKlick={() => {
                wechseln('vorlagen')
              }}
            >
              {texte.seiten.vorlagen.titel}
            </Knopf>

            {/* **Die führende Handlung der Seite** — das Artboard zeichnet
                „+ Paket beziehen" als einzigen Hauptknopf im Kopf. Der Fokus
                geht vorher auf ihn, damit er nach dem Schließen dorthin
                zurückkehrt — in WebKit gibt ein Klick keinen Fokus. */}
            <Knopf gewicht="haupt" griff={ausloeser} aufKlick={beziehen}>
              {t.plan.beziehen}
            </Knopf>
          </Seitenkopf>

          {/* **Die Leiste über den Paketen** — AP-9 (#2366). */}
          <div className="mb-4">
            <Paketkennzahlen pakete={pakete} befunde={abhaengigkeiten} />
          </div>

          <Paketliste pakete={pakete} befunde={abhaengigkeiten.daten} />

          {/* **Das Seltene eingeklappt unter dem Arbeitsplatz** — eine
              Zeile je Kasten, bis jemand ihn öffnet. Die Anlage steht in
              `Paketliste`, weil sie deren Spiele braucht. */}
          <div className="mt-3 flex flex-col gap-3">
            <Aufklappkasten titel={t.abhaengigUeberschrift}>
              <Abhaengigkeiten
                daten={abhaengigkeiten.daten}
                fehler={abhaengigkeiten.fehler}
              />
            </Aufklappkasten>

            {/* **Die bezogenen Stände** (G04, #2251): wie weit die Quelle
                eines bezogenen Baums ist — gefragt wird erst auf Knopfdruck. */}
            <Aufklappkasten titel={t.bezuege.ueberschrift}>
              <Bezugsstaende geaendert={nachladen} beziehen={beziehen} />
            </Aufklappkasten>

            {/* **Die Sammeldaten stehen hier** (#888): sie tun dasselbe wie
                die Pakete — etwas Großes von außen holen und ablegen. */}
            <Aufklappkasten titel={t.sammeldaten.ueberschrift}>
              <Sammeldaten />
            </Aufklappkasten>
          </div>
        </div>
      </Verdeckung>

      {gezeichnet && (
        <Activity mode={sichtbar(offen)}>
          <Schublade
            titel={t.plan.beziehenTitel}
            schliessenBezeichnung={
              texte.allgemein.arbeitsplatz.schubladen.schliessen
            }
            schliessen={() => {
              setzeOffen(false)
            }}
          >
            <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
              <Paketbezug />

              {/* **Zwei andere Wege zum selben Ziel, eingeklappt** — ein
                  Verzeichnis von MSE2, ein `.mse-installer`, ein Umschlag.
                  Offen ist eine Gruppe, nicht drei (R3). */}
              <Aufklappkasten titel={t.mse2Ueberschrift}>
                <Uebernahme uebernommen={nachladen} />
              </Aufklappkasten>

              {/* **Eine Vorlage aus einem Umschlag** (#1639): eine Vorlage
                  ist ein **Paket**, und der Umschlag trägt die Dateien und
                  nicht den Namen. */}
              <Aufklappkasten titel={t.umschlagTitel}>
                <Vorlagenumschlag angenommen={nachladen} />
              </Aufklappkasten>
            </div>
          </Schublade>
        </Activity>
      )}
    </>
  )
}
