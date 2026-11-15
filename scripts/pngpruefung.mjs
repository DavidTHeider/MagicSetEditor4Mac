// **Ein PNG vollständig prüfen** — Z2 des Zweitblicks über Phase 6.11.
//
// # Warum mehr als der Kopf
//
// Bis Z2 las der Leistungsvergleich Breite und Höhe aus den ersten 24 Byte.
// Das reicht für „welche Größe", aber nicht für „ist das Bild unversehrt": ein
// Bild, das auf halber Strecke abbricht, hat denselben Anfang. Der Leistungsvergleich sagt zu, dass je Bild Signatur, Breite und
// Höhe stimmen, und der Lastlauf, dass die zugesagte Datei ein Bild ist —
// beides setzt voraus, dass „Bild" mehr heißt als „fängt an wie eines".
//
// # Was geprüft wird
//
// PNG, ISO/IEC 15948 (Abschnitte 5.2 Signatur, 5.3 Blöcke, 11.2.2 IHDR, 7.3
// Filter, 8.2 Adam7):
//
//   1. die **Signatur**;
//   2. jeder **Block** hat die Bytes, die seine Länge verspricht, und **seine
//      Prüfsumme stimmt** (CRC-32 über Art und Daten). Die Prüfsumme kostet bei
//      einer Karte Mikrosekunden und fängt, was die Länge nicht fängt: ein
//      verdrehtes Bit mitten in den Daten;
//   3. der erste Block ist **IHDR**, mit zulässigen Angaben (Farbart und
//      Tiefe passen zusammen, Verfahren 0, Filterverfahren 0, Verschachtelung
//      0 oder 1) und einer Größe, die entpackt werden darf;
//   4. bei Farbart 3 ein **PLTE**, mindestens ein **IDAT**, und am Ende ein
//      **IEND**;
//   5. die IDAT-Blöcke, zusammengesetzt und entpackt, haben **genau** die
//      Länge `Höhe × (1 + Zeilenbytes)` — bei Adam7 die Summe der sieben
//      Durchgänge —, und jede Zeile beginnt mit einem Filterbyte von 0 bis 4.
//
// # Was nicht
//
// **Die Pixel.** Entfiltert und verglichen wird nicht: ob ein Bild so aussieht
// wie das der Referenz, misst der Kanonlauf (`docs/MESSMETHODIK.md` §5b).
// Auch Hilfsblöcke (`tEXt`, `pHYs`, …) werden nur auf Länge und Prüfsumme
// gelesen, und **Bytes hinter IEND sind kein Fehler**: das Bild ist vollständig.

import { crc32, inflateSync } from 'node:zlib'

const signatur = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])

/** Die Zahl der Kanäle einer Farbart (PNG, Tabelle 11.1). */
const kanaele = { 0: 1, 2: 3, 3: 1, 4: 2, 6: 4 }

/** zeilenlaenge gibt die Byte einer Zeile ohne ihr Filterbyte. */
export const zeilenlaenge = (breite, farbart, tiefe) => Math.ceil((breite * kanaele[farbart] * tiefe) / 8)

/**
 * durchgaenge gibt die Teilbilder als [Breite, Höhe] — eines, wenn das Bild
 * nicht verschachtelt ist, sonst die bis zu sieben von Adam7 (PNG, 8.2).
 */
export function durchgaenge(breite, hoehe, verschachtelt) {
  if (!verschachtelt) return [[breite, hoehe]]
  return [[0, 0, 8, 8], [4, 0, 8, 8], [0, 4, 4, 8], [2, 0, 4, 4], [0, 2, 2, 4], [1, 0, 2, 2], [0, 1, 1, 2]]
    .map(([x0, y0, dx, dy]) => [breite > x0 ? Math.ceil((breite - x0) / dx) : 0, hoehe > y0 ? Math.ceil((hoehe - y0) / dy) : 0])
    .filter(([b, h]) => b > 0 && h > 0)
}

/** Was eine Farbart an Tiefen zulässt (PNG, Tabelle 11.1). */
const tiefen = { 0: [1, 2, 4, 8, 16], 2: [8, 16], 3: [1, 2, 4, 8], 4: [8, 16], 6: [8, 16] }

/** Höchstens so viele Byte ungepackter Bilddaten werden erwartet: ein Kopf, der mehr verlangt, ist ein Angriff oder ein Schaden. */
const hoechstens = 512 * 2 ** 20

const gross = 0x7fffffff

const nein = (fehler, masse = {}) => ({ ok: false, fehler, ...masse })

/** kopfLesen deutet die 13 Byte von IHDR — oder gibt den Grund, warum nicht. */
function kopfLesen(daten) {
  const kopf = {
    breite: daten.readUInt32BE(0), hoehe: daten.readUInt32BE(4), tiefe: daten[8], farbart: daten[9],
    verfahren: daten[10], filter: daten[11], verschachtelt: daten[12],
  }
  const masse = { breite: kopf.breite, hoehe: kopf.hoehe }
  if (kopf.breite < 1 || kopf.hoehe < 1 || kopf.breite > gross || kopf.hoehe > gross) {
    return { fehler: nein(`IHDR: unmögliche Größe ${kopf.breite}x${kopf.hoehe}`, masse) }
  }
  if (!tiefen[kopf.farbart]?.includes(kopf.tiefe)) {
    return { fehler: nein(`IHDR: Farbart ${kopf.farbart} gibt es nicht mit Tiefe ${kopf.tiefe}`, masse) }
  }
  if (kopf.verfahren !== 0 || kopf.filter !== 0) return { fehler: nein('IHDR: unbekanntes Verfahren oder Filterverfahren', masse) }
  if (kopf.verschachtelt > 1) return { fehler: nein(`IHDR: unbekannte Verschachtelung ${kopf.verschachtelt}`, masse) }
  return { kopf: { ...kopf, verschachtelt: kopf.verschachtelt === 1 } }
}

/** bloecke zerlegt die Datei in ihre Blöcke und prüft Länge und Prüfsumme jedes einzelnen. */
function bloecke(daten) {
  const liste = []
  let ort = 8
  while (ort < daten.length) {
    if (ort + 12 > daten.length) return { fehler: `abgeschnitten im Kopf eines Blocks (Byte ${ort} von ${daten.length})` }
    const laenge = daten.readUInt32BE(ort)
    const art = daten.toString('latin1', ort + 4, ort + 8)
    if (laenge > gross) return { fehler: `Block ${art} mit unmöglicher Länge ${laenge}` }
    if (ort + 12 + laenge > daten.length) return { fehler: `Block ${art} abgeschnitten (${daten.length - ort - 12} von ${laenge} Byte)` }
    const summe = daten.readUInt32BE(ort + 8 + laenge)
    if (crc32(daten.subarray(ort + 4, ort + 8 + laenge)) !== summe) return { fehler: `Prüfsumme von ${art} stimmt nicht` }
    liste.push({ art, daten: daten.subarray(ort + 8, ort + 8 + laenge) })
    ort += 12 + laenge
    if (art === 'IEND') break
  }
  return { liste }
}

/**
 * pngPruefen prüft ein PNG vollständig und gibt `{ ok: true, breite, hoehe,
 * farbart, tiefe, verschachtelt }` oder `{ ok: false, fehler }` — mit den
 * Maßen aus dem Kopf, wo dieser sich lesen ließ.
 */
export function pngPruefen(daten) {
  if (daten.length < 8 || !daten.subarray(0, 8).equals(signatur)) return nein('keine PNG-Signatur')

  const zerlegt = bloecke(daten)
  if (zerlegt.fehler) return nein(zerlegt.fehler)
  const { liste } = zerlegt
  if (liste[0]?.art !== 'IHDR' || liste[0].daten.length !== 13) return nein('IHDR fehlt am Anfang')

  const gelesen = kopfLesen(liste[0].daten)
  if (gelesen.fehler) return gelesen.fehler
  const { kopf } = gelesen
  const masse = { breite: kopf.breite, hoehe: kopf.hoehe }

  if (liste.at(-1).art !== 'IEND') return nein('IEND fehlt (die Datei ist abgeschnitten)', masse)
  if (kopf.farbart === 3 && !liste.some((b) => b.art === 'PLTE')) return nein('PLTE fehlt bei Farbart 3', masse)
  const gepackt = liste.filter((b) => b.art === 'IDAT').map((b) => b.daten)
  if (gepackt.length === 0) return nein('IDAT fehlt', masse)

  // **Die Länge, die der Kopf verspricht**, durchgangsweise gerechnet — ohne
  // je Zeile etwas anzulegen: der Kopf kann Milliarden Zeilen versprechen.
  const teile = durchgaenge(kopf.breite, kopf.hoehe, kopf.verschachtelt)
    .map(([b, h]) => ({ hoehe: h, zeile: 1 + zeilenlaenge(b, kopf.farbart, kopf.tiefe) }))
  const erwartet = teile.reduce((summe, t) => summe + t.hoehe * t.zeile, 0)
  if (erwartet > hoechstens) return nein(`Größe ${kopf.breite}x${kopf.hoehe} verlangt ${erwartet} Byte — mehr als erlaubt`, masse)

  let roh
  try {
    roh = inflateSync(Buffer.concat(gepackt), { maxOutputLength: erwartet })
  } catch (fehler) {
    if (fehler.code === 'ERR_BUFFER_TOO_LARGE') return nein(`mehr Bilddaten als die erwarteten ${erwartet} Byte`, masse)
    return nein(`Bilddaten lassen sich nicht entpacken: ${fehler.message}`, masse)
  }
  if (roh.length !== erwartet) return nein(`Bilddaten haben ${roh.length} Byte, erwartet ${erwartet} (${kopf.breite}x${kopf.hoehe})`, masse)

  let ort = 0
  for (const t of teile) {
    for (let y = 0; y < t.hoehe; y++, ort += t.zeile) {
      if (roh[ort] > 4) return nein(`Filterbyte ${roh[ort]} in der Zeile ab Byte ${ort}`, masse)
    }
  }
  return { ok: true, ...masse, farbart: kopf.farbart, tiefe: kopf.tiefe, verschachtelt: kopf.verschachtelt }
}
