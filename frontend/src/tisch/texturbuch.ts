// **Welche Texturen ein Tisch freigibt** — B04
// (#2223).
//
// # Warum es das gibt
//
// **`Assets.load` legt jede Textur in einem globalen Speicher ab**, und der
// gibt sie erst mit `Assets.unload` frei (PixiJS 8.22, „Destroying textures",
// nachgelesen am 2026-10-09). Die Spielbilder sind inhaltsadressiert: eine
// bearbeitete Karte hat am nächsten Tisch eine neue Adresse, und die alte
// blieb liegen — samt ihrem Bild im Grafikspeicher von WebKit. Gemessen am
// nativen Fenster (`scripts/huellenprobe/dauerteile.mjs`): Bearbeiten und Tisch
// im Wechsel ließen den Webprozess um 45–60 MB je Minute wachsen, jeder Teil
// allein nicht.
//
// # Was hier entschieden wird
//
// **Was ein Tisch geladen hat, gibt er beim Abräumen frei — außer ein anderer
// Tisch hält es noch.** Der Zähler über alle Tische steht deshalb außerhalb
// des Buchs. Ein Bild, das erst nach dem Abräumen ankommt, wird gleich wieder
// freigegeben; sonst bliebe es liegen, weil niemand es mehr nennt.
//
// Den Ruf an PixiJS macht `pixitextur.ts`; hier steht nur die Rechnung.

export interface Texturbuch {
  // geladen vermerkt eine angekommene Adresse und sagt, ob sie bleibt.
  geladen: (adresse: string) => 'behalten' | 'freigeben'

  // raeume schließt das Buch und nennt, was jetzt freizugeben ist.
  raeume: () => string[]
}

// tischnutzer zählt je Adresse die Tische, die sie halten — über alle Tische
// eines Fensters, wie der Speicher von PixiJS.
const tischnutzer = new Map<string, number>()

export function neuesTexturbuch(
  nutzer: Map<string, number> = tischnutzer,
): Texturbuch {
  const eigene = new Set<string>()
  let geraeumt = false

  return {
    geladen(adresse) {
      if (geraeumt)
        return (nutzer.get(adresse) ?? 0) > 0 ? 'behalten' : 'freigeben'

      if (!eigene.has(adresse)) {
        eigene.add(adresse)
        nutzer.set(adresse, (nutzer.get(adresse) ?? 0) + 1)
      }

      return 'behalten'
    },

    raeume() {
      if (geraeumt) return []
      geraeumt = true

      const frei: string[] = []
      for (const adresse of eigene) {
        // **Jede eigene Adresse ist gezählt** — `geladen` hat sie eingetragen.
        const rest = Number(nutzer.get(adresse)) - 1
        if (rest > 0) {
          nutzer.set(adresse, rest)
        } else {
          nutzer.delete(adresse)
          frei.push(adresse)
        }
      }
      eigene.clear()

      return frei
    },
  }
}
