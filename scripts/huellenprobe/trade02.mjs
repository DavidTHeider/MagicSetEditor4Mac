// **Derselbe Szenariosatz vor und nach dem Umbau** — TRADE-02.AK-04
// (#1716).
//
// Drei Handel in drei Zuständen — ein Entwurf, ein angebotener und ein
// angenommener, alle aus Sicht der Initiatorin —, und je Zustand wird am
// **nativen Fenster** gezählt, was im Handelsraum sichtbar ist: Knöpfe,
// Eingaben, Überschriften, die Höhe und **Schrittknöpfe, die hier gar nicht
// gehen** (ein Klick darauf wäre ein Fehlversuch).
//
// **Die Probe misst generisch** — über den rechten Bereich der Seite und die
// deutschen Schrittnamen —, damit sie auf dem Stand vor dem Umbau genauso
// läuft wie danach. Die Zahlen beider Läufe stehen im Vermerk.
//
// **Seit TRADE-02b (#1951) mit Preisen**: eine Gegenstelle auf 127.0.0.1
// nennt sie — dieselbe Gestalt wie `scryfallgegenstelle_test.go`, und kein
// Aufruf geht hinaus. Das Szenario bleibt dasselbe, damit die Zählung
// vergleichbar bleibt.
import { execFile, execFileSync } from 'node:child_process'
import { createServer } from 'node:http'
import { promisify } from 'node:util'
import { fileURLToPath } from 'node:url'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = fileURLToPath(new URL('../..', import.meta.url))

// eroeffne legt einen Tausch an und gibt seine Kennung.
function eroeffne(ort, gegenueber) {
  const positionen = JSON.stringify([
    { von: gegenueber, printing: 'bolt-m10', menge: 1, guete: 'nearMint', sprache: 'de', finish: 'nonfoil' },
  ])

  return execFileSync(`${wurzel}.go/bin/mse`, ['handel', 'eroeffnen', '', 'tausch', gegenueber], {
    cwd: wurzel,
    input: positionen,
    env: kindumgebung(ort),
  }).toString().split('\t')[0]
}

// Die Schritte, wie sie am Fenster heißen — und welche je Zustand gehen.
const schrittnamen = ['Anbieten', 'Gegenangebot', 'Zustimmen', 'Vorbereiten', 'Übergabe bestätigen', 'Ablehnen', 'Abbrechen', 'Abgelaufen', 'Zur Klärung']
const zulaessig = {
  entwurf: ['Anbieten', 'Abbrechen'],
  angeboten: ['Zustimmen', 'Abbrechen', 'Abgelaufen'],
  angenommen: ['Vorbereiten', 'Abbrechen'],
}

const messung = `(() => {
  const raum = document.querySelector('main .min-w-0.grow')
  if (!raum) return null
  const sichtbar = (e) => e.checkVisibility()
  const knoepfe = [...raum.querySelectorAll('button')].filter(sichtbar)
  const namen = ${JSON.stringify(schrittnamen)}
  return {
    knoepfe: knoepfe.length,
    eingaben: [...raum.querySelectorAll('input, select, textarea')].filter(sichtbar).length,
    ueberschriften: [...raum.querySelectorAll('h2, h3, h4')].filter(sichtbar).length,
    hoehe: Math.round(raum.scrollHeight),
    schritte: knoepfe.map((b) => b.innerText.trim()).filter((t) => namen.includes(t)),
  }
})()`

const zeile = (partner) => `[...document.querySelectorAll('li button')].find((b) => b.innerText.startsWith('${partner}'))`

// karte ist die Antwort der Gegenstelle auf `/cards/<id>`.
const karte = (id) => JSON.stringify({
  object: 'card', id, name: 'Erste Probe', lang: 'en', layout: 'normal',
  mana_cost: '{2}{W}', type_line: 'Creature — Testling', oracle_text: '{T}: Add one mana of any color.',
  power: '2', toughness: '3', rarity: 'rare', set: 'prf', set_name: 'Prüfsatz', collector_number: '1',
  artist: 'Selbst Gemalt', image_uris: { normal: 'http://127.0.0.1/bild.png', art_crop: 'http://127.0.0.1/bild.png' },
  scryfall_uri: 'https://beispiel.invalid/karte/1',
  prices: { eur: '1.50', eur_foil: '4.25', usd: '1.80', usd_foil: '5.10', tix: '0.03' },
})

// preiseAus holt den Preis eines Drucks aus einer Gegenstelle auf 127.0.0.1.
//
// **Asynchron und nicht über `p.mse`**: ein `execFileSync` hielte die
// Ereignisschleife an, und die Gegenstelle im selben Prozess antwortete nie.
async function preiseAus(ort, druck) {
  const gegenstelle = createServer((anfrage, antwort) => {
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(karte(anfrage.url.split('?')[0].split('/').pop()))
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })
  try {
    await promisify(execFile)(`${wurzel}.go/bin/mse`, ['preise-holen', druck], {
      cwd: wurzel,
      env: {
        ...kindumgebung(ort),
        MSE_SCRYFALL_ADRESSE: `http://127.0.0.1:${String(gegenstelle.address().port)}`,
      },
    })
  } finally {
    gegenstelle.close()
  }
}

export default async function trade02(p) {
  const ort = p.ort()

  p.notiz('1 — drei Tausche: Entwurf, angeboten, angenommen')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'peer-anlegen')
  await preiseAus(ort, 'bolt-m10')
  eroeffne(ort, 'bentwurf')
  p.mse(ort, 'handel', 'anbieten', eroeffne(ort, 'bangeboten'))
  const angenommen = eroeffne(ort, 'bangenommen')
  p.mse(ort, 'handel', 'anbieten', angenommen)
  p.mse(ort, 'handel', 'zustimmen', angenommen)
  p.mse(ort, 'handel', 'zustimmen', angenommen, 'bangenommen')

  await p.starte({ ort })
  await p.taste('5', 'befehl')

  const summe = { knoepfe: 0, eingaben: 0, ueberschriften: 0, unzulaessig: 0 }

  for (const [partner, zustand] of [['bentwurf', 'entwurf'], ['bangeboten', 'angeboten'], ['bangenommen', 'angenommen']]) {
    p.notiz(`2 — ${zustand}`)
    await p.klickeAuf(zeile(partner), `die Zeile von ${partner}`)
    await p.warte('der Raum steht', async () => ((await p.seite(`return ${messung}`))?.knoepfe ?? 0) > 3, 20000)
    await new Promise((fertig) => { setTimeout(fertig, 800) })
    const gemessen = await p.seite(`return ${messung}`)
    const falsch = gemessen.schritte.filter((s) => !zulaessig[zustand].includes(s))
    p.zeige(`${zustand}: ${gemessen.knoepfe} Knöpfe, ${gemessen.eingaben} Eingaben, ${gemessen.ueberschriften} Überschriften, ${gemessen.hoehe} pt hoch; Schritte ${gemessen.schritte.join(', ')}; unzulässig ${falsch.length} (${falsch.join(', ')})`)
    summe.knoepfe += gemessen.knoepfe
    summe.eingaben += gemessen.eingaben
    summe.ueberschriften += gemessen.ueberschriften
    summe.unzulaessig += falsch.length
    // **Das Bild zeigt den Assistenten von oben** — die Messung oben hängt
    // daran nicht.
    await p.seite("document.querySelector('[data-handelsassistent]')?.scrollIntoView({ block: 'start', behavior: 'instant' })")
    if (zustand === 'entwurf') {
      const hoehen = await p.seite("return [...document.querySelectorAll('[data-tafel=angebot] input, [data-tafel=angebot] select')].filter((e) => e.checkVisibility()).map((e) => (e.labels?.[0]?.innerText ?? e.id).split('\\n')[0] + ' ' + Math.round(e.getBoundingClientRect().height))")
      p.zeige(`Höhen der Positionsfelder: ${hoehen.join(', ')}`)
    }
    await p.bild(zustand)
  }

  p.zeige(`zusammen: ${summe.knoepfe} Knöpfe, ${summe.eingaben} Eingaben, ${summe.ueberschriften} Überschriften, ${summe.unzulaessig} unzulässige Schrittknöpfe`)
  p.pruefe('⌘Q beendet', await p.beende())
}
