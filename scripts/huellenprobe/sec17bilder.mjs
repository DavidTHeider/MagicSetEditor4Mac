// **Fremde Vorschaubilder unter der Schutzrichtlinie** — SEC-17
// (#2164),
// angestrichen vom Zweitblick über Phase H.
//
// Die Kartensuche zeigt je Treffer ein kleines Bild, und dessen Adresse
// kommt **vom Dienst**: Scryfall nennt `cards.scryfall.io`, magicthegathering.io
// `gatherer.wizards.com`, das auf `gatherer-static.wizards.com` weiterleitet
// (gemessen am 2026-10-07). Die WebView holt sie selbst. Kein anderes
// Drehbuch betritt die Suche mit Bildern — `trade02` nennt nur `normal`, und
// das holt Go.
//
// **Eine Gegenstelle auf 127.0.0.1 antwortet für Scryfall**; die beiden
// Bildadressen darin sind **echte** — die eine, wie Scryfall sie nennt, die
// andere in der Gestalt, die `mtgapi` daraus macht. Für die Richtlinie zählt
// der Ursprung und nicht, wer ihn nannte. **Die WebView holt die zwei Bilder
// wirklich**; sonst geht nichts hinaus.
//
// **Was sie nicht belegt**: dass Scryfall und Gatherer morgen dieselben
// Ursprünge nennen. Eine neue Herkunft fiele als `CSP-Verstoß` im Protokoll
// auf — dafür liest jede Hüllenprobe ihr Protokoll.
import { createServer } from 'node:http'

import { zurSeite } from './uirevue.mjs'

const bilder = [
  'https://cards.scryfall.io/small/front/1/1/11bf83bb-c95b-4b4f-9a56-ce7a1816307a.jpg?1783940984',
  'https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=527297&type=card',
]

// treffer ist eine Karte der Suchantwort mit diesem Vorschaubild.
const treffer = (nummer, bild) => ({
  object: 'card', id: `00000000-0000-4000-8000-00000000000${String(nummer)}`,
  name: `Bildprobe ${String(nummer)}`, lang: 'en', layout: 'normal',
  mana_cost: '{U}', type_line: 'Creature — Testling', oracle_text: '',
  rarity: 'common', set: 'prf', set_name: 'Prüfsatz', collector_number: String(nummer),
  image_uris: { small: bild, normal: bild },
  scryfall_uri: 'https://beispiel.invalid/karte/1',
})

export default async function sec17bilder(p) {
  const ort = p.ort()
  const gegenstelle = createServer((anfrage, antwort) => {
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(JSON.stringify({
      object: 'list', total_cards: bilder.length, has_more: false,
      data: bilder.map((bild, i) => treffer(i + 1, bild)),
    }))
    p.zeige(`die Gegenstelle bekam ${anfrage.url.split('?')[0]}`)
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })

  try {
    p.notiz('1 — die Kartensuche mit zwei fremden Vorschaubildern')
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    await p.starte({
      ort,
      umgebung: { MSE_SCRYFALL_ADRESSE: `http://127.0.0.1:${String(gegenstelle.address().port)}` },
    })
    await p.taste('2', 'befehl')
    await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
    await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
    await zurSeite(p, 'Karten')
    await p.klickeAuf("[...document.querySelectorAll('main [aria-haspopup=menu]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Mehr')", 'das Menü »Mehr«')
    await p.klickeText('Aus Scryfall anlegen …')
    await p.klickeAuf("[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.startsWith('Bei Scryfall suchen'))", 'das Suchfeld')
    await p.tippeInsFeld('Bildprobe')
    await p.klickeText('Suchen')
    await p.wartetext('Bildprobe 2', 20000)

    const geladen = await p.warte('beide Vorschaubilder sind entschieden', async () => {
      const stand = await p.seite("return [...document.querySelectorAll('main img')].filter((i) => /scryfall\\.io|wizards\\.com/.test(i.src)).map((i) => ({ ursprung: new URL(i.src).host, fertig: i.complete, breite: i.naturalWidth }))")
      return stand.length === bilder.length && stand.every((b) => b.fertig) ? stand : null
    }, 30000)
    for (const bild of geladen) {
      p.pruefe(`das Bild von ${bild.ursprung} ist geladen`, bild.breite > 0, `${String(bild.breite)} px breit`)
    }
    await p.bild('sec17-suchbilder')
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    gegenstelle.close()
  }
}
