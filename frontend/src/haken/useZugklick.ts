// **Ein Zug ist kein Klick** — P5
// (#2373),
// am nativen Fenster gefunden (`tischhand09`): WebKit schickt nach einem Zug,
// der auf einem Kartenknopf begann, noch einen `click` an ihn — und der löst
// seit P4 die Primärhandlung aus. Wer eine Handkarte auf den Friedhof zog,
// wirkte sie zugleich.
//
// **Ein Zug über die Schwelle verschluckt den einen Klick danach**; der
// nächste Druck hebt es auf, damit ein ausgebliebener Klick keinen späteren
// schluckt.
//
// **Und ohne Bewegung gibt es keinen Zug** (ebenda gefunden): kam zwischen
// Drücken und Loslassen ein neues Bild, in dem die Karte schon woanders lag,
// sah der Tisch sie dort und den Zeiger am alten Ort — und schob sie zurück.
import { useRef } from 'react'

// klickschwelle ist, wie weit (in Punkten) der Zeiger zwischen Drücken und
// Loslassen wandern darf, und es bleibt ein Klick — eine Hand zittert.
const klickschwelle = 4

interface Zeigerpunkt {
  clientX: number
  clientY: number
}

export interface Zugklick {
  gedrueckt: (ereignis: Zeigerpunkt) => void

  // losgelassen sagt, ob der Zeiger über die Schwelle gewandert ist — nur
  // dann ist es ein Zug.
  losgelassen: (ereignis: Zeigerpunkt) => boolean

  // nimmKlick sagt, ob ein Klick gilt — und verbraucht einen verschluckten.
  nimmKlick: () => boolean
}

export function useZugklick(): Zugklick {
  const druck = useRef({ x: 0, y: 0 })
  const zugVorbei = useRef(false)

  return {
    gedrueckt: (ereignis) => {
      druck.current = { x: ereignis.clientX, y: ereignis.clientY }
      zugVorbei.current = false
    },
    losgelassen: (ereignis) => {
      zugVorbei.current =
        Math.hypot(
          ereignis.clientX - druck.current.x,
          ereignis.clientY - druck.current.y,
        ) > klickschwelle

      return zugVorbei.current
    },
    nimmKlick: () => {
      const gilt = !zugVorbei.current
      zugVorbei.current = false

      return gilt
    },
  }
}
