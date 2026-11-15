// **Die Kartenbilder am fremden Tisch** — P9
// (#2377),
// Befunde C-11 und C-12, Vermerk 0198.
//
// Zwei Anwendungen auf diesem Mac, je mit eigenem Bestand — wie in `nat04`.
// **Und die Bestände kollidieren mit Absicht**: der Wirt liest
// `tisch/TSP.json`, der Gast `probe.json`, und unter derselben Zeile liegt
// bei beiden eine **andere** Karte (Zeile 2: »Bergpfad« beim Wirt,
// »Eiserner Wächter« beim Gast). Bo spielt das Deck des Wirts mit sechzehn
// Bergpfaden.
//
// - **Der Gast zeigt die Bilder des Wirts** — erfragt über die Leitung, und
//   byteweise dasselbe Bild, das der Wirt selbst zeichnet.
// - **Er schlägt keine Zeile im eigenen Bestand nach**: im Rufbuch steht kein
//   `Spieltextur`. Hätte er es getan, sähe er den Eisernen Wächter.
// - **Der Wirt zeichnet nur, was der Gast sehen darf** — auch wer die
//   Bildkennung einer verdeckten Karte kennt, bekommt einen Grund.
//
// **Was sie nicht belegt**: zwei Geräte in zwei Netzen (#1998), und ein Deck
// aus dem Bestand des Gastes — das ist P9b.
import { bindung } from '../huellenhandel.mjs'
import { rufbuchHolen } from '../huellenrufbuch.mjs'

const leinwand = "return !!document.querySelector('canvas')?.checkVisibility()"

// koppeln macht zwei Anwendungen einander bekannt — bestätigt, mit Zweck.
async function koppeln(hier, dort, name) {
  await bindung(hier, `b.PeerkontaktAufnehmen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(name)})`)
  const bestaetigt = await bindung(hier, `b.PeerkontaktBestaetigen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(dort.fingerabdruck)})`)
  hier.pruefe(`${name} ist bestätigt`, !bestaetigt?.fehler, JSON.stringify(bestaetigt?.fehler ?? ''))
  const erlaubt = await bindung(hier, `b.PeerzweckErlauben(${JSON.stringify(dort.kennung)}, 'partie')`)
  hier.pruefe(`${name} darf eine Partie spielen`, !erlaubt?.fehler, JSON.stringify(erlaubt?.fehler ?? ''))
}

// abdruck liest ein Bild über seine Adresse und gibt Länge und einen
// FNV-1a-Abdruck der Bytes — ohne `crypto.subtle`, das ein eigenes Schema
// nicht als sicheren Ursprung kennen muss.
const abdruck = (p, adresse) => p.seite(`const r = await fetch(${JSON.stringify(adresse)}); const b = new Uint8Array(await r.arrayBuffer()); let h = 0x811c9dc5; for (const x of b) { h ^= x; h = Math.imul(h, 0x01000193) >>> 0 } return r.ok ? b.length + ':' + h.toString(16) : 'HTTP ' + r.status`)

// rufe zählt, wie oft eine Bindung im Rufbuch steht — und wie oft gelungen.
function rufe(buch, name) {
  const eintrag = buch[name] ?? { anzahl: 0, ausgaenge: {} }

  return { anzahl: eintrag.anzahl, gelungen: eintrag.ausgaenge.gelungen ?? 0 }
}

export default async function netzbild09(p) {
  const gast = new p.constructor('netzbild09-gast')
  gast.zeilen = p.zeilen

  try {
    await zweiFenster(p, gast)
  } finally {
    p.fehler += gast.fehler
    gast.aufraeumen()
  }
}

async function zweiFenster(p, gast) {
  const wirtOrt = p.ort()
  const gastOrt = gast.ort()
  for (const [probe, ort] of [[p, wirtOrt], [gast, gastOrt]]) {
    probe.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    probe.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  }
  p.mse(wirtOrt, 'mtgjson-lesen', 'testdaten/mtgjson/tisch/TSP.json', 'magic')
  gast.mse(gastOrt, 'mtgjson-lesen', 'testdaten/mtgjson/probe.json', 'magic')
  p.mse(wirtOrt, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const k of ['1', '3', '4', '5', '6']) p.mse(wirtOrt, 'deck', 'legen', '1', `--karte=${k}`, '8')
  p.mse(wirtOrt, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [['2', 16], ['4', 4], ['6', 4]]) p.mse(wirtOrt, 'deck', 'legen', '2', `--karte=${k}`, String(n))
  const angelegt = p.mse(wirtOrt, 'partie-anlegen', 'magic', 'ada:Ada:1', 'bo:Bo:2')
  const partie = /Partie ([0-9a-f]+)/.exec(angelegt)?.[1] ?? ''
  p.pruefe('die Partie ist angelegt', partie !== '', angelegt.split('\n')[0])
  for (const sitz of ['ada', 'bo']) {
    for (const art of ['spieler.starthand', 'spieler.behalten']) p.mse(wirtOrt, 'partie-befehl', partie, sitz, `${art}/${sitz}`, art)
  }

  p.notiz('Zwei Anwendungen, zwei Bestände, dieselben Zeilen — andere Karten')
  await p.starte({ ort: wirtOrt })
  await p.groesse(1440, 900)
  await gast.starte({ ort: gastOrt })
  await gast.groesse(1440, 900)
  const wirt = await bindung(p, 'b.PeeridentitaetSicherstellen()')
  const bo = await bindung(gast, 'b.PeeridentitaetSicherstellen()')
  await koppeln(p, bo, 'Gast')
  await koppeln(gast, wirt, 'Wirt')
  const horch = await bindung(p, "b.HorchStarten('127.0.0.1:0')")
  const port = horch?.stand?.adresse?.split(':').pop()
  p.pruefe('der Wirt horcht', Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
  await bindung(gast, `b.PeerkontaktAdresse(${JSON.stringify(wirt.kennung)}, '127.0.0.1:${port}')`)

  p.notiz('Der Wirt öffnet seinen Tisch')
  await p.aktiviere()
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  await p.wartetext('Weiterspielen')
  await p.klickeText('Weiterspielen')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('die Leinwand des Wirts zeichnet', () => p.seite(leinwand), 20000)
  await p.klickeText('Tisch öffnen')
  await p.wartetext('Dieser Tisch steht offen.', 10000)

  p.notiz('Der Gast setzt sich')
  // **Gezählt wird ab hier**: was die Seiten davor riefen, gehört nicht zum Tisch.
  const vorher = await rufbuchHolen(gast)
  const vorFremd = rufe(vorher, 'Fremdkartenbild')
  const vorEigen = rufe(vorher, 'Spieltextur')
  await gast.aktiviere()
  await gast.taste('4', 'befehl')
  await gast.wartetext('Partie anfangen', 20000)
  await gast.klickeText('Partie anfangen')
  await gast.wartetext('Bei jemandem mitspielen', 20000)
  await gast.klickeText('Bei jemandem mitspielen')
  await gast.klickeAuf("[...document.querySelectorAll('label, [role=radio]')].find((e) => e.checkVisibility() && /^Wirt · /.test((e.innerText ?? '').trim()))", 'den Wirt')
  await gast.klickeText('Mitspielen')
  await gast.wartetext('Du sitzt am Tisch von', 20000)
  await gast.warte('die Leinwand des Gastes zeichnet', () => gast.seite(leinwand), 20000)

  p.notiz('Der Gast zeigt die Bilder des Wirts — und fragt den eigenen Bestand nach keiner Zeile')
  // **Gleiche Karten teilen ein Bild**: die Bildkennung hängt am Platz im
  // Abzug, nicht am Objekt — sieben Handkarten aus drei Plätzen sind drei
  // Bilder. Wie viele es sein müssen, sagt die Sicht des Wirts für Bo.
  const sicht = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'bo')`)
  const hand = (sicht?.partie?.objekte ?? []).filter((o) => o.eigner === 'bo' && o.zone === 'hand' && o.bild)
  const bilderSoll = new Set(hand.map((o) => o.bild)).size
  p.zeige(`Bos Hand beim Wirt: ${String(hand.length)} Karten, ${String(bilderSoll)} Bilder, Zeilen ${JSON.stringify(hand.map((o) => o.karteID))}`)
  await gast.warte('die Bilder vom Wirt sind da', async () => rufe(await rufbuchHolen(gast), 'Fremdkartenbild').gelungen - vorFremd.gelungen >= bilderSoll, 30000).catch(() => false)
  const buch = gast.rufbuch
  const nachher = rufe(buch, 'Fremdkartenbild')
  const fremd = { anzahl: nachher.anzahl - vorFremd.anzahl, gelungen: nachher.gelungen - vorFremd.gelungen }
  const eigen = { anzahl: rufe(buch, 'Spieltextur').anzahl - vorEigen.anzahl }
  gast.zeige(`Rufbuch des Gastes: Fremdkartenbild ${JSON.stringify(fremd)}, Spieltextur ${JSON.stringify(eigen)}, Rueckentextur ${JSON.stringify(rufe(buch, 'Rueckentextur'))}`)
  gast.pruefe('der Gast hat jedes Bild seiner Hand beim Wirt erfragt, je einmal', hand.length === 7 && fremd.gelungen === bilderSoll && fremd.anzahl === bilderSoll, `${JSON.stringify(fremd)} für ${String(bilderSoll)} Bilder`)
  gast.pruefe('und keine Zeile im eigenen Bestand nachgeschlagen', eigen.anzahl === 0, JSON.stringify(eigen))
  // **Nah an der eigenen Hand**, damit das Bild die Karten zeigt und nicht
  // nur, dass da welche liegen.
  await gast.klickeText('Auf meinen Sitz')
  await gast.seite("document.querySelector('[role=application]')?.scrollIntoView({ block: 'start', behavior: 'instant' })")
  await gast.bild('netzbild09-gast-sieht-die-bilder')

  p.notiz('Byteweise dasselbe Bild — und der eigene Bestand hätte ein anderes gezeigt')
  // **Zeile 2 ist beim Wirt der Bergpfad** — die Sicht am Wirt rechnet für
  // den eigenen Bestand und nennt sie.
  const bergpfad = hand.find((o) => o.karteID === '2')
  p.pruefe('in Bos Hand liegt ein Bergpfad mit Bildkennung', Boolean(bergpfad), JSON.stringify(hand.map((o) => o.karteID)))
  if (bergpfad) {
    const beimWirt = await bindung(p, `b.Spieltextur(${JSON.stringify(bergpfad.karteID)}, 'vorne', 'tisch')`)
    const vomWirt = await bindung(gast, `b.Fremdkartenbild(${JSON.stringify(partie)}, ${JSON.stringify(bergpfad.kennung)}, ${JSON.stringify(bergpfad.bild)}, 'vorne', 'tisch')`)
    const gleichnamig = await bindung(gast, `b.Spieltextur(${JSON.stringify(bergpfad.karteID)}, 'vorne', 'tisch')`)
    const a = await abdruck(p, beimWirt?.textur?.adresse ?? '')
    const b = await abdruck(gast, vomWirt?.textur?.adresse ?? '')
    const c = await abdruck(gast, gleichnamig?.textur?.adresse ?? '')
    p.zeige(`Zeile ${bergpfad.karteID}: beim Wirt ${a}, beim Gast vom Wirt ${b}, beim Gast aus dem eigenen Bestand ${c}`)
    gast.pruefe('das Bild am Gast ist das Bild des Wirts', a === b && !a.startsWith('HTTP'), `${a} · ${b}`)
    gast.pruefe('dieselbe Zeile im eigenen Bestand wäre eine andere Karte', c !== a, `${c} · ${a}`)
  }

  p.notiz('Wer die Bildkennung einer verdeckten Karte kennt, bekommt trotzdem keine')
  const adaSicht = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'ada')`)
  const verdeckt = (adaSicht?.partie?.objekte ?? []).find((o) => o.eigner === 'ada' && o.zone === 'hand' && o.bild)
  if (verdeckt) {
    const abgewiesen = await bindung(gast, `b.Fremdkartenbild(${JSON.stringify(partie)}, ${JSON.stringify(verdeckt.kennung)}, ${JSON.stringify(verdeckt.bild)}, 'vorne', 'tisch')`)
    gast.zeige(`Antwort auf Adas Handkarte: ${JSON.stringify({ adresse: abgewiesen?.textur?.adresse, befunde: abgewiesen?.textur?.befunde, fehler: abgewiesen?.fehler })}`)
    gast.pruefe('der Wirt zeichnet Adas Hand nicht für Bo', !abgewiesen?.textur?.adresse && (abgewiesen?.textur?.befunde ?? []).some((s) => /nicht zu sehen/.test(s)), JSON.stringify(abgewiesen))
  } else {
    p.pruefe('Ada hat eine Handkarte mit Bildkennung', false, JSON.stringify(adaSicht?.fehler ?? ''))
  }

  const meldungen = await gast.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")
  gast.pruefe('am fremden Tisch steht keine Fehlermeldung', meldungen.length === 0, JSON.stringify(meldungen))
  await bindung(gast, 'b.PartieBeitrittBeenden()')
  gast.pruefe('⌘Q beendet den Gast', await gast.beende())
  p.pruefe('⌘Q beendet den Wirt', await p.beende())
}
