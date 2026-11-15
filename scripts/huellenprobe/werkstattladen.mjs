// **Wie lange die Werkstatt braucht, bis ihr Baum steht** — #2010, Folge aus
// UXR-18 (#1926).
//
// Der Baum holt die Dateien **aller** Pakete auf einmal — mit Absicht, denn
// der Filter sucht über Dateinamen auch zugeklappter Pakete (U7, #1418). Seit
// UXR-18 zeichnet er nur noch, was zu sehen ist; ob das Laden dabei im Weg
// steht, misst dieses Drehbuch: die Zeit vom Aufschlagen bis zur ersten
// bedienbaren Zeile, mit den Paketen, die auf diesem Rechner liegen.
//
// **Was es nicht misst**: einen kalten Seitenspeicher des Systems. Es fährt
// zweimal und nennt beide Zahlen; die zweite ist die warme.
import { zurSeite } from './uirevue.mjs'

const zeilen = "return [...document.querySelectorAll('main ul button[aria-expanded]')].filter((b) => b.checkVisibility()).length"

export default async function werkstattladen(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  for (const lauf of ['erster', 'zweiter']) {
    await p.starte({ ort })
    await p.warte('das Fenster steht', () => p.seite("return !!document.querySelector('main')"), 20000)
    const beginn = Date.now()
    await zurSeite(p, 'Vorlagen')
    await p.warte('die erste Paketzeile steht', async () => (await p.seite(zeilen)) > 0, 60000)
    // **`zurSeite` wartet nach dem Klick 1,5 Sekunden fest** — abgezogen,
    // damit die Zahl das Laden misst und nicht das Drehbuch.
    const dauer = Date.now() - beginn - 1500
    const pakete = p.mse(ort, 'pakete').trim().split('\n').length
    p.zeige(`${lauf} Lauf: ${String(dauer)} ms bis zur ersten Paketzeile, ${String(pakete)} Pakete`)
    p.pruefe(`${lauf} Lauf: unter zwei Sekunden`, dauer < 2000, `${String(dauer)} ms`)
    p.pruefe(`⌘Q beendet (${lauf})`, await p.beende())
  }
}
