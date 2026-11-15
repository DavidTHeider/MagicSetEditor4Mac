// **Der Fuß der Seitenleiste in jeder Sprache** — #2032
// (#2032).
//
// Im Bildabzug zu UXR-02 stand im Englischen „Open command pale…": der Text
// war breiter als die Leiste. Dieses Drehbuch startet die Anwendung in jeder
// der acht Sprachen im kleinsten Fenster und misst, ob der Text des Fußes
// abgeschnitten ist (`scrollWidth` über `clientWidth`).
const sprachen = ['de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh']

const fuss = `
  const knopf = [...document.querySelectorAll('#seitenleiste button')].filter((b) => b.checkVisibility()).at(-1)
  const text = knopf?.querySelector('span')
  if (!text) return null
  return { text: text.textContent.trim(), breite: text.clientWidth, noetig: text.scrollWidth, abgeschnitten: text.scrollWidth > text.clientWidth + 1 }
`

export default async function leistenfuss(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1024')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '700')

  const abgeschnitten = []
  for (const sprache of sprachen) {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    await p.starte({ ort })
    await p.warte('die Leiste steht', () => p.seite("return !!document.querySelector('#seitenleiste')"), 20000)
    const wert = await p.seite(fuss)
    p.zeige(`${sprache}: ${JSON.stringify(wert)}`)
    if (!wert || wert.abgeschnitten) abgeschnitten.push(sprache)
    p.pruefe(`⌘Q beendet die Anwendung (${sprache})`, await p.beende())
  }

  p.pruefe('der Fuß steht in jeder Sprache ganz da', abgeschnitten.length === 0, abgeschnitten.join(', '))
}
