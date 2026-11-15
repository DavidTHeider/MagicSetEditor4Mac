// **Die Worte des Symboleditors — je nach Ziel** — DOM-02.AK-02
// (#1697).
//
// Überschrift, Erklärung und Sichern heißen beim Setsymbol anders als bei
// einem Symbol der Bibliothek: dort wird ins Set gelegt, hier eine neue
// Fassung gesichert. **Eine Stelle für die Wahl**, damit der Editor sie nicht
// an vier Stellen trifft.
import type { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import type { Bibliotheksbearbeitung } from './symbolablegen'

type Kartentexte = ReturnType<typeof useTexte>['seiten']['karten']

export interface Symbolworte {
  ueberschrift: string
  erklaerung: string
  sichern: string

  // gesichert ist die Vorlage der Meldung — mit `{name}`.
  gesichert: string
}

export function symbolworte(
  karten: Kartentexte,
  bibliothek: Bibliotheksbearbeitung | undefined,
): Symbolworte {
  if (bibliothek === undefined) {
    return {
      ueberschrift: karten.symbol.ueberschrift,
      erklaerung: karten.symbol.erklaerung,
      sichern: karten.symbol.sichern,
      gesichert: karten.symbol.gesichert,
    }
  }

  const eigene = karten.symbolbibliothek

  return {
    ueberschrift: einsetzen(eigene.bearbeitungsueberschrift, {
      name: bibliothek.name,
    }),
    erklaerung: eigene.bearbeitungserklaerung,
    sichern: eigene.inBibliothekSichern,
    gesichert: eigene.inBibliothekGesichert,
  }
}
