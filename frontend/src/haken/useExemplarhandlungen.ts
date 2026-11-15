// **Was an einem Exemplar zu tun ist, als eine Liste** — AP-6, Teil 2
// (#2363),
// Abnahme „Kontext an der Sammlungszeile mit Ins Deck, Zum Tausch, Menge,
// Löschen“. Der Inspector meldet sie dem Objekthandlungsraum; daraus werden
// das Menü „Exemplar“ der Leiste und der Rechtsklick an der Zeile
// (`Objektkontextmenue`). **Sichtbar bleiben sie im Inspector** — Bearbeiten
// und Löschen oben, Zum Tausch und Ins Deck am Fuß —, denn eine Handlung, die
// es nur im Kontextmenü gibt, gibt es nicht (U2, #1413).
//
// # Die Menge steht unter Bearbeiten
//
// **Kein eigener Eintrag „Menge“**: im Inspector ändert man sie im Formular
// hinter „Bearbeiten“, als erstes Feld (UXR-10, #1918), und der Rechtsklick
// zeigt dieselbe Liste wie die sichtbaren Knöpfe. Ein Eintrag „Eins mehr“
// wäre ein Schreibweg, den der Inspector nicht hat.
//
// # Löschen fragt im Inspector
//
// Der Eintrag stellt die Frage, wie sein Knopf — gelöscht wird erst nach der
// Antwort (UX-10.AK-04, #1682).
import type { ExemplarDTO } from '../api'
import { useAnschluss } from '../anschlussraum'
import { insDeck, zumTausch } from '../bausteine/exemplaruebergabe'
import { useTexte } from '../sprache/kontext'
import { useHandlungsliste } from './useHandlungsliste'

interface Exemplarhandlungen {
  exemplar: ExemplarDTO
  laeuft: boolean

  // offen sagt, dass das Formular schon offen ist; fragt, dass die
  // Rückfrage vor dem Löschen schon steht — dann ruht der Eintrag, wie
  // dort der Knopf fehlt.
  offen: boolean
  fragt: boolean
  bearbeiten: () => void
  loeschen: () => void
}

export function useExemplarhandlungen({
  exemplar,
  laeuft,
  offen,
  fragt,
  bearbeiten,
  loeschen,
}: Exemplarhandlungen): void {
  const eigene = useTexte().seiten.sammlung
  const anschluss = useAnschluss()

  useHandlungsliste({
    titel: eigene.exemplarmenue,
    art: 'exemplar',
    kennung: String(exemplar.id),
    eintraege: [
      {
        schluessel: 'bearbeiten',
        name: eigene.bearbeiten,
        aktion: 'bearbeiten',
        schreibt: true,
        angehalten: laeuft || offen,
        aufWahl: bearbeiten,
      },
      {
        schluessel: 'loeschen',
        name: eigene.loeschen,
        aktion: 'loeschen',
        schreibt: true,
        angehalten: laeuft || fragt,
        aufWahl: loeschen,
      },
      {
        schluessel: 'zumTausch',
        name: eigene.zumTausch,
        aufWahl: () => {
          anschluss.uebergib(zumTausch(exemplar))
        },
      },
      {
        schluessel: 'insDeck',
        name: eigene.insDeck,
        aufWahl: () => {
          anschluss.uebergib(insDeck(exemplar))
        },
      },
    ],
  })
}
