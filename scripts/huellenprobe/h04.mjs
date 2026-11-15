// **Je Sprache Stand, Herkunft und Redaktion unter der Sprachwahl** — H04
// (#2258),
// am nativen Fenster:
//
//   1. auf Deutsch: *Einstellungen* öffnen, „Wie weit jede Sprache ist"
//      aufklappen — acht Zeilen, jede mit Stand, Herkunft und „von keinem
//      Menschen redaktionell geprüft", der Tag der Übersetzung in der
//      Darstellung des Fensters,
//   2. auf Japanisch dasselbe — der Stand spricht die Sprache des Fensters.
const texte = {
  de: { titel: 'Wie weit jede Sprache ist', ungeprueft: 'von keinem Menschen redaktionell geprüft', maschinell: 'maschinell übersetzt' },
  ja: { titel: '各言語の状況', ungeprueft: '人による校閲なし', maschinell: '機械翻訳' },
}

const kasten = (titel) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(titel)})`

const zeilen = "return [...document.querySelectorAll('table tbody tr')].filter((z) => z.checkVisibility()).map((z) => [...z.children].map((z) => z.innerText.trim()))"

export default async function h04(p) {
  const ort = p.ort()

  for (const sprache of ['de', 'ja']) {
    const t = texte[sprache]
    p.notiz(`Das Fenster auf ${sprache}`)
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    await p.starte({ ort })
    await p.warte('das Dokument spricht die Sprache', async () => (await p.seite('return document.documentElement.lang')) === sprache, 20000)

    await p.taste(',', 'befehl')
    await p.warte('der Kasten steht unter der Sprachwahl', () => p.seite(`return !!(${kasten(t.titel)})`), 20000)
    await p.seite(`(${kasten(t.titel)}).scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(kasten(t.titel), `der Kasten »${t.titel}«`)
    const tafel = await p.warte('die Tafel steht da', async () => {
      const gelesen = await p.seite(zeilen)
      return gelesen.length >= 8 ? gelesen : null
    }, 10000)

    p.zeige(tafel.map((z) => z.join(' | ')).join('\n'))
    p.pruefe('acht Sprachen, je vier Spalten', tafel.length === 8 && tafel.every((z) => z.length === 4))
    p.pruefe('keine Redaktion behauptet', tafel.every((z) => z[3] === t.ungeprueft))
    p.pruefe('sieben maschinell übersetzt, in der Sprache des Fensters', tafel.filter((z) => z[1] === t.maschinell).length === 7)
    p.pruefe('der Tag der Übersetzung steht in der Herkunft', tafel.filter((z) => /2026/.test(z[2])).length === 6)
    await p.bild(`h04-sprachstaende-${sprache}`)
    p.pruefe('⌘Q beendet', await p.beende())
  }
}
