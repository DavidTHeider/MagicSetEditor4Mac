// **Die Felder des Stils am nativen Fenster** —
// #2298.
//
// Der Weg in Schritten: die Sets aufschlagen (⌘2), das Set im Saatstil
// `probe-schlicht` wählen, die Karte wählen — unter dem Formular steht
// „Felder des Stils" mit `vermerk`. Ein Wert wird getippt und mit Tab
// verlassen; danach steht er im Bestand (gelesen über `mse zusatzfeld` im Bündel, also
// über einen **anderen** Weg als den, der schrieb), und die Vorschau zeigt
// ihn unten rechts.
//
// **Vorher gab es diesen Weg nicht**: ein Zusatzwert ließ sich nur als roher
// Block im Feld `extra data` der tiefsten Formularstufe setzen.
//
// **Was er nicht belegt**: andere Feldarten als Text, und Stile des Packs.
import { execFileSync } from 'node:child_process'
import { join, resolve } from 'node:path'
import { kindumgebung } from '../huellenorte.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')

// mseImBuendel liest über das `mse` **im Bündel** — `scripts/bauen.sh
// abnahme` legt es neben das Programm, wie die Einrichtung es verlinkt. Nur
// dort findet es die mitgelieferte Saat unter `Contents/Resources/saat`
// (`app.Saatverzeichnis`); `.go/bin/mse` neben dem Repo kennt
// `probe-schlicht` ohne sie gar nicht und meldete, der Stil habe keine
// Zusatzfelder.
function mseImBuendel(ort, ...argumente) {
  const imBuendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse')

  return execFileSync(imBuendel, argumente, { encoding: 'utf8', env: kindumgebung(ort) })
}

const setname = 'Stilfeldprobe'
const kartenname = 'Vermerkprobe'
const kasten = "document.querySelector('section[aria-label=\"Felder des Stils\"]')"
const vermerk = "[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'vermerk')"
const zeile = (name) => `[...document.querySelectorAll('[role=row]')].find((e) => (e.getAttribute('aria-label') ?? '').startsWith(${JSON.stringify(name)}))`

// abdruck liest die Pixel des Vorschaubilds — die Quelle allein sagt nichts,
// sie ist bei jedem Rendern eine neue.
const abdruck = `return (async () => {
  const b = document.querySelector('img[alt="Vorschau der Karte"]')
  if (!b) return null
  await b.decode()
  const c = document.createElement('canvas')
  c.width = b.naturalWidth; c.height = b.naturalHeight
  const x = c.getContext('2d'); x.drawImage(b, 0, 0)
  const d = x.getImageData(0, 0, c.width, c.height).data
  let h = 2166136261
  for (let i = 0; i < d.length; i++) h = Math.imul(h ^ d[i], 16777619) >>> 0
  return h.toString(16)
})()`

const kennung = (roh) => String(JSON.parse(roh).id)

export default async function stilfelder(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  const set = kennung(p.mse(ort, 'set-neu', setname, 'probe', 'probe-schlicht', '--json'))
  // **Eine Karte antwortet mit `karteID`**, ein Set mit `id` — der Vertrag
  // der Kommandozeile (`mse karte … anlegen --json`).
  const karte = String(JSON.parse(p.mse(ort, 'karte', set, 'anlegen', '--json')).karteID)
  p.mse(ort, 'feld', set, karte, 'name', kartenname)

  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.wartetext(setname, 30000)
  await p.klickeText(setname)
  await p.warte('die Kartenliste steht', () => p.seite(`return !!(${zeile(kartenname)})`), 30000)
  await p.klickeAuf(zeile(kartenname), `die Zeile »${kartenname}«`)

  const steht = await p.warte('„Felder des Stils" steht unter dem Formular', () => p.seite(`return !!(${kasten}) && !!(${vermerk})`), 15000).then(() => true, () => false)
  p.pruefe('der Kasten „Felder des Stils" steht mit dem Feld vermerk', steht)
  await p.warte('die Vorschau steht', () => p.seite(abdruck), 15000)
  const vorher = await p.seite(abdruck)
  await p.bild('stilfelder-vorher')

  await p.klickeAuf(vermerk, 'das Feld vermerk')
  await p.tippe('Vorabdruck')
  await p.taste('Tab')

  const gespeichert = await p.warte('der Wert steht im Bestand', () => mseImBuendel(ort, 'zusatzfeld', set, karte).includes('Vorabdruck'), 15000).then(() => true, () => false)
  // **Im roten Fall steht da, was die Kommandozeile las** — ein bloßes
  // `false` kostet einen zweiten Lauf.
  let gelesen = ''
  if (!gespeichert) {
    try {
      gelesen = mseImBuendel(ort, 'zusatzfeld', set, karte).replace(/\s+/g, ' ').trim()
    } catch (fehler) {
      gelesen = String(fehler).replace(/\s+/g, ' ').slice(0, 300)
    }
  }
  p.pruefe('der Wert steht im Bestand — über die Kommandozeile gelesen', gespeichert, gelesen)

  const geaendert = await p.warte('die Vorschau zeigt den Vermerk', async () => {
    const jetzt = await p.seite(abdruck)

    return jetzt !== null && jetzt !== vorher
  }, 15000).then(() => true, () => false)
  p.pruefe('die Vorschau hat sich mit dem Wert geändert', geaendert)
  await p.bild('stilfelder-nachher')

  // **Funktionszugang über die Tastatur**: das Feld ist ein gewöhnliches
  // Eingabefeld im Fokusweg — Tab vom Feld davor erreicht es.
  const fokus = await p.seite(`return document.activeElement !== (${vermerk})`)
  p.pruefe('nach dem Verlassen liegt der Fokus nicht mehr im Feld', fokus)

  p.pruefe('⌘Q beendet', await p.beende())
}
