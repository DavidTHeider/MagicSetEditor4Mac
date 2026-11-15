// **PNG-Dateien für Prüfungen bauen** — Hilfe der Tests von
// `scripts/pngpruefung.mjs` und der Läufer, die Bilder prüfen
// (Z2 des Zweitblicks über Phase 6.11).
//
// Keine Prüfung und kein Lauf: ein Test braucht Bilder, die **vollständig und
// richtig** sind, und solche, die es an genau einer Stelle nicht sind. Mit dem
// Kopf allein — 24 Byte — kam hier jedes Bild durch.

import { crc32, deflateSync } from 'node:zlib'

import { durchgaenge, zeilenlaenge } from './pngpruefung.mjs'

const signatur = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])

/** chunk baut einen Block: Länge, Art, Daten, Prüfsumme über Art und Daten. */
export function chunk(art, daten = Buffer.alloc(0)) {
  const kopf = Buffer.alloc(8)
  kopf.writeUInt32BE(daten.length, 0)
  kopf.write(art, 4, 'latin1')
  const summe = Buffer.alloc(4)
  summe.writeUInt32BE(crc32(Buffer.concat([kopf.subarray(4), daten])), 0)
  return Buffer.concat([kopf, daten, summe])
}

/** rohdaten baut die ungepackten Bilddaten: je Zeile ein Filterbyte und dann Inhalt. */
export function rohdaten(breite, hoehe, farbart, tiefe, verschachtelt = false, fuellung = 0x7f) {
  const zeilen = []
  for (const [b, h] of durchgaenge(breite, hoehe, verschachtelt)) {
    for (let y = 0; y < h; y++) zeilen.push(Buffer.concat([Buffer.from([0]), Buffer.alloc(zeilenlaenge(b, farbart, tiefe), fuellung)]))
  }
  return Buffer.concat(zeilen)
}

/**
 * png baut ein vollständiges PNG. Mit `roh` kommen die ungepackten Bilddaten
 * von außen — für ein Bild, das an einer Stelle nicht stimmt.
 */
export function png({
  breite = 4, hoehe = 3, farbart = 6, tiefe = 8, verschachtelt = false, roh, teile = 1, palette = farbart === 3,
  kopfAendern,
} = {}) {
  const kopf = Buffer.alloc(13)
  kopf.writeUInt32BE(breite, 0)
  kopf.writeUInt32BE(hoehe, 4)
  kopf[8] = tiefe
  kopf[9] = farbart
  kopf[12] = verschachtelt ? 1 : 0
  // **Nach der Prüfsumme geändert wäre es ein anderer Fehler**: wer einen Kopf
  // mit unmöglichen Angaben braucht, ändert ihn hier, vor dem Block.
  kopfAendern?.(kopf)
  const gepackt = deflateSync(roh ?? rohdaten(breite, hoehe, farbart, tiefe, verschachtelt))
  const stuecke = []
  const schritt = Math.ceil(gepackt.length / teile)
  for (let i = 0; i < gepackt.length; i += schritt) stuecke.push(chunk('IDAT', gepackt.subarray(i, i + schritt)))
  return Buffer.concat([
    signatur, chunk('IHDR', kopf),
    ...(palette ? [chunk('PLTE', Buffer.from([0, 0, 0, 255, 255, 255]))] : []),
    ...stuecke, chunk('IEND'),
  ])
}
