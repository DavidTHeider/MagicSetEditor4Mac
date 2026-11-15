// **War der Bildschirm aus, während ein Drehbuch lief?** — C03
// (#2228).
//
// Der zweite volle Lauf vom 2026-10-10 lief ganz bei ausgeschaltetem
// Bildschirm, und zwanzig Drehbücher wurden rot — fast alle an Fristen. Ein
// fehlendes Bild vermerkte die Probe schon als Lücke (`huellenbeleg.mjs`); was
// sie nicht sah: **eine Webansicht, die niemand sieht, zeichnet nicht**, und
// wer auf ein gezeichnetes Bild wartet, wartet bis zur Frist. Der Fehler sieht
// dann aus wie einer der Anwendung.
//
// **Die Quelle ist das Energieprotokoll** (`pmset -g log`) und nicht ein
// Blick vor und nach jedem Drehbuch: der fände eine Phase nicht, die mitten
// darin beginnt und endet. Gelesen wird einmal je Lauf — das Protokoll ist
// über 20 MB groß und braucht vier Sekunden.
import { execFileSync } from 'node:child_process'

const zeile = /^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) ([+-]\d{2})(\d{2})\s+Notification\s+Display is turned (on|off)/
const zeitzeile = /^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}) ([+-]\d{2})(\d{2})\s/

// ausphasen liest die Zeiträume, in denen der Bildschirm aus war, als
// ISO-Zeiten in UTC. `von: null` heißt aus seit vor dem Protokoll, `bis: null`
// aus bis jetzt.
export function ausphasen(text) {
  const phasen = []
  let offen
  for (const roh of text.split('\n')) {
    const treffer = zeile.exec(roh)
    if (!treffer) continue
    const [, tag, uhr, stunden, minuten, art] = treffer
    const zeit = new Date(`${tag}T${uhr}${stunden}:${minuten}`).toISOString()
    if (art === 'off') {
      offen ??= { von: zeit, bis: null }
    } else {
      phasen.push({ von: offen?.von ?? null, bis: zeit })
      offen = undefined
    }
  }
  if (offen) phasen.push(offen)

  return phasen
}

// ueberschneidungen nennt die Phasen, die den Zeitraum von `beginn` bis
// `ende` berühren.
export function ueberschneidungen(phasen, beginn, ende) {
  const anfang = Date.parse(beginn)
  const schluss = Date.parse(ende)

  return phasen.filter(({ von, bis }) =>
    (von === null || Date.parse(von) < schluss) && (bis === null || Date.parse(bis) > anfang))
}

// bildschirmAus liest das Energieprotokoll und nennt die Ausphasen im
// Zeitraum — **oder `undefined`, wenn es das nicht weiß**: wenn es sich nicht
// lesen ließ, und auch, wenn **vor dem Beginn kein Wechsel** darin steht
// (Zweitblick über Phase C, Befund 7). Ein leeres Protokoll ergab sonst
// „keine Ausphase", und wer es nicht weiß, behauptet nicht, der Bildschirm
// sei an gewesen.
export function bildschirmAus(beginn, ende, lesen = protokoll) {
  let text
  try {
    text = lesen()
  } catch {
    return undefined
  }

  const vorher = text.split('\n').some((roh) => {
    const treffer = zeile.exec(roh)

    return treffer !== null && Date.parse(`${treffer[1]}T${treffer[2]}${treffer[3]}:${treffer[4]}`) <= Date.parse(beginn)
  })
  if (!vorher) return undefined

  // **Und es reicht bis zum Ende des Laufs** (Nachlauf über Phase C, #2284):
  // ein abgeschnittenes Protokoll mit einer alten Einschaltzeile galt sonst
  // als „nie aus" — was nach seiner letzten Zeile geschah, weiß es nicht.
  // Jede Zeile trägt eine Zeit, nicht nur die des Bildschirms — und
  // `pmset -g log` schließt mit einer Zeile zur Zeit des Lesens („Showing all
  // currently held IOKit power assertions", nachgesehen am 2026-10-10); ein
  // vollständiges Protokoll reicht damit immer bis jetzt.
  const zuletzt = Math.max(...text.split('\n').map((roh) => {
    const treffer = zeitzeile.exec(roh)

    return treffer === null ? -Infinity : Date.parse(`${treffer[1]}T${treffer[2]}${treffer[3]}:${treffer[4]}`)
  }))
  if (zuletzt < Date.parse(ende)) return undefined

  return ueberschneidungen(ausphasen(text), beginn, ende)
}

function protokoll() {
  return execFileSync('pmset', ['-g', 'log'], { maxBuffer: 512 * 1024 * 1024 }).toString()
}

// vermerkeImBeleg setzt die Ausphasen ans Ende des Kopfes — dorthin, wo der
// Beleg auch die fehlenden Bilder nennt, damit niemand den Vermerk zwischen
// den Zeilen sucht.
export function vermerkeImBeleg(beleg, phasen) {
  const zeitraum = phasen.map(({ von, bis }) => `${von ?? 'unbekannt'} bis ${bis ?? 'andauernd'}`).join('; ')
  const vermerk = `**Der Bildschirm war aus** (${zeitraum}), während das Drehbuch lief — ein Rot hier ist zuerst die Umgebung: eine Webansicht, die niemand sieht, zeichnet nicht.`
  const kopfAnfang = beleg.indexOf('\n\n') + 2
  const kopfEnde = beleg.indexOf('\n\n', kopfAnfang)

  return `${beleg.slice(0, kopfEnde)}\n${vermerk}${beleg.slice(kopfEnde)}`
}

// vermerkeLauf setzt den Vermerk in jeden Beleg eines Drehbuchs, das während
// einer Ausphase lief, trägt die Phasen an seinem Fall ein und nennt die
// Drehbücher.
export function vermerkeLauf(faelle, phasen, { lesen, schreiben }) {
  const imDunkeln = []
  for (const fall of faelle) {
    const getroffen = ueberschneidungen(phasen, fall.beginn, fall.ende)
    if (getroffen.length === 0) continue
    fall.bildschirmAus = getroffen
    imDunkeln.push(fall.name)
    for (const pfad of fall.belege) schreiben(pfad, vermerkeImBeleg(lesen(pfad), getroffen))
  }

  return imDunkeln
}
