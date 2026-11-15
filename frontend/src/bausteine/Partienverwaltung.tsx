// **Die Verwaltung einer Partie als Ganzem** — PLAY-03.AK-05
// (#1711):
// „Löschen betrifft eigene Entwürfe nach definierter Regel, Kopien werden
// neue Übungszustände ohne alte Peer-Zustimmungen."
//
// # Warum die Wege in einem Menü stehen und nicht als Knöpfe
//
// **Weil sie sich auf die Zeile beziehen und selten gebraucht werden** —
// UXR-13 (#1921).
// Bis dahin standen sie als drei Knöpfe an jeder Zeile, dazu „+ Marke": bei
// neunzehn Partien 76 Bedienelemente für Wege, die man alle paar Wochen
// geht, neben dem Anfang einer neuen Partie. **Das Menü sagt, welche Partie
// es meint** — sein Name nennt sie, und die Zeile, an der es steht, auch.
//
// # Und warum keine Rückfrage vor dem Löschen
//
// **Weil die Regel die Rückfrage ist.** Was gelöscht werden darf, ist ein
// Entwurf: keine Einwilligung, nicht beendet. Alles, woran jemand hängt,
// weist die Domäne ohnehin ab und sagt, warum ([domain.Loeschbar]). Eine
// Rückfrage obendrauf wäre dieselbe Frage zweimal — und wer oft genug ohne
// Grund gefragt wird, klickt irgendwann alles weg.
import { Menue, type Menueeintrag } from './Menue'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'
import { useTexte } from '../sprache/kontext'

import type { Verwaltungsstand } from '../haken/usePartieverwaltung'

interface Verwaltungseigenschaften {
  kennung: string

  // name ist, wie die Partie in der Zeile heißt — für den Namen des Menüs.
  name: string

  lauf: Verwaltungsstand

  // marken öffnet oder schließt die Marken der Zeile — **sie schreiben über
  // einen eigenen Vorgang** (`Objektmarken`) und ruhen deshalb nicht mit.
  marken: () => void
}

// verwaltungseintraege sind die Wege der Verwaltung — für das Menü hier
// **und** den Rechtsklick an der Zeile (AP-5, #2362): eine Liste an einer
// Stelle, damit keiner der beiden Wege einen Eintrag hat, den der andere
// nicht kennt.
export function verwaltungseintraege(
  kennung: string,
  lauf: Verwaltungsstand,
  marken: () => void,
  texte: Katalog,
): Menueeintrag[] {
  const eigene = texte.seiten.spielen
  const laeuft = lauf.laeuft(kennung)

  return [
    {
      schluessel: 'duplizieren',
      schreibt: true,
      name: eigene.partienDuplizieren,
      aktion: 'duplizieren',
      angehalten: laeuft,
      aufWahl: () => {
        lauf.kopieren(kennung)
      },
    },
    {
      schluessel: 'ausfuehren',
      name: eigene.partienAusfuhr,
      angehalten: laeuft,
      aufWahl: () => {
        lauf.ausfuehren(kennung)
      },
    },
    {
      schluessel: 'loeschen',
      schreibt: true,
      name: eigene.partienLoeschen,
      aktion: 'loeschen',
      angehalten: laeuft,
      aufWahl: () => {
        lauf.loeschen(kennung)
      },
    },
    {
      schluessel: 'marken',
      schreibt: true,
      name: texte.bausteine.objektmarken.ueberschrift,
      aufWahl: marken,
    },
  ]
}

export function Partienverwaltung({
  kennung,
  name,
  lauf,
  marken,
}: Verwaltungseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  return (
    <span className="flex flex-wrap items-center gap-2">
      <Menue
        bezeichnung={einsetzen(eigene.partienHandlungen, { name })}
        gewicht="leise"
        eintraege={verwaltungseintraege(kennung, lauf, marken, texte)}
      >
        {eigene.partienHandlungenKurz}
      </Menue>

      <Verwaltungsmeldung kennung={kennung} lauf={lauf} />
    </span>
  )
}

// Verwaltungsmeldung zeigt, was aus dem letzten Weg wurde — **an der Zeile,
// die ihn ausgelöst hat**.
//
// **Und nur dort** (§6): eine Meldung über das Löschen von A an der Zeile
// von B ließe jemanden am falschen Ende suchen.
export function Verwaltungsmeldung({
  kennung,
  lauf,
}: {
  kennung: string
  lauf: Verwaltungsstand
}) {
  const meldung = lauf.meldung

  if (meldung?.kennung !== kennung) {
    return null
  }

  return meldung.istFehler ? (
    <span role="alert" className="text-sm text-fehler">
      {meldung.satz}
    </span>
  ) : (
    <span role="status" className="text-sm text-leise">
      {meldung.satz}
    </span>
  )
}
