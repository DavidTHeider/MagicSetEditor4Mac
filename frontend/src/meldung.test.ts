// Die Meldung aus einer Ursache — **in der Sprache der Sitzung**.
//
// # Der Befund, um den es geht
//
// #133:
// `FehlerDTO` trägt einen `schluessel`, und niemand las ihn. Die sichtbare
// Meldung war der **deutsche** Satz aus Go — eine englische Sitzung zeigte
// deutsche Fehlermeldungen, und die englischen Kataloge waren produktiv
// nirgends verdrahtet.
//
// Der Vertragstest zwischen Go und den Katalogen prüft, dass jeder angemeldete
// Schlüssel in beiden Katalogen steht. Er prüft nicht, dass ihn jemand liest.
import { describe, expect, it } from 'vitest'

import { Bindungsfehler, type FehlerDTO } from './api'
import { Katalogfehler, meldungAus } from './meldung'
import { kataloge } from './sprache/kataloge'

// bindungsfehler baut einen Fehler, wie ihn `internal/app` schickt.
function bindungsfehler(teile: Partial<FehlerDTO> = {}): Bindungsfehler {
  return new Bindungsfehler({
    code: 'nichtGefunden',
    schluessel: 'nichtGefunden',
    meldung: 'Der gesuchte Eintrag wurde nicht gefunden.',
    ...teile,
  })
}

describe('meldungAus', () => {
  // **Der Schlüssel entscheidet, nicht der Satz aus Go.**
  it('übersetzt einen Bindungsfehler über seinen Schlüssel', () => {
    const ursache = bindungsfehler()

    expect(meldungAus(ursache, kataloge.en)).toBe(
      kataloge.en.fehler.nichtGefunden,
    )
    expect(meldungAus(ursache, kataloge.de)).toBe(
      kataloge.de.fehler.nichtGefunden,
    )
  })

  // **Der deutsche Satz bleibt der Rückfall** (§9). Ein Schlüssel, den der
  // Katalog nicht kennt, kostet eine unübersetzte Meldung — nicht eine leere.
  it('fällt auf den Satz aus Go zurück, wenn der Schlüssel unbekannt ist', () => {
    const ursache = bindungsfehler({
      schluessel: 'gibtEsNicht',
      meldung: 'Ein Satz, den nur Go kennt.',
    })

    expect(meldungAus(ursache, kataloge.en)).toBe('Ein Satz, den nur Go kennt.')
  })

  // **Ein Fehler aus dem Frontend trägt einen Schlüssel, keinen Satz.**
  it('übersetzt einen Katalogfehler', () => {
    const ursache = new Katalogfehler('vorschauLeer')

    expect(meldungAus(ursache, kataloge.en)).toBe(
      kataloge.en.allgemein.meldungen.vorschauLeer,
    )
  })

  // **Mit Platzhaltern**, wie die Ersatzmethode der Fassade sie braucht.
  it('setzt die Werte eines Katalogfehlers ein', () => {
    const ursache = new Katalogfehler('methodeFehlt', { name: 'Sets' })

    expect(meldungAus(ursache, kataloge.de)).toContain('Sets')
    expect(meldungAus(ursache, kataloge.en)).toContain('Sets')
    expect(meldungAus(ursache, kataloge.en)).not.toBe(
      meldungAus(ursache, kataloge.de),
    )
  })

  // **Ein gewöhnlicher Fehler kommt durch, wie er ist.**
  //
  // Er entsteht in Wails oder im Browser und trägt keinen Schlüssel. Ihn zu
  // übersetzen hieße zu raten, was er bedeutet.
  it('reicht einen gewöhnlichen Fehler durch', () => {
    expect(meldungAus(new Error('network down'), kataloge.en)).toBe(
      'network down',
    )
  })

  it('macht aus einer Ursache ohne Fehler ihren Text', () => {
    expect(meldungAus('nur ein Text', kataloge.en)).toBe('nur ein Text')
  })

  // **Eine Validierung ohne eigenen Schlüssel zeigt ihren Grund** —
  // #1601.
  //
  // # Der Befund, für den es diesen Lauf gibt
  //
  // **An der laufenden Anwendung gesehen**: eine Partie mit leerem Deck
  // angefangen, und es stand „Es ist ein unerwarteter Fehler aufgetreten. Was
  // genau, steht im Protokoll." Im Protokoll stand nichts — für eine
  // Validierung bleibt es nach eigener Entscheidung leer. Der Satz aus Go war
  // die ganze Zeit da und wurde weggeworfen.
  //
  // **Geprüft wird beides**: dass der Rahmen übersetzt ist und dass der Grund
  // darin steht. Eine Prüfung nur auf den Grund ginge auch durch, wenn der
  // Rahmen fehlte.
  it('zeigt den Grund einer Validierung im übersetzten Rahmen', () => {
    const ursache = new Bindungsfehler({
      code: 'validierung',
      schluessel: 'validierungsgrund',
      meldung: 'Validierung fehlgeschlagen: ein Deckabzug ohne Karten',
      feld: '',
      werte: { grund: 'ein Deckabzug ohne Karten' },
    })

    const deutsch = meldungAus(ursache, kataloge.de)
    const englisch = meldungAus(ursache, kataloge.en)

    expect(deutsch).toContain('ein Deckabzug ohne Karten')
    expect(englisch).toContain('ein Deckabzug ohne Karten')

    // **Der Rahmen ist übersetzt** — sonst wäre es dieselbe Zeile zweimal.
    expect(deutsch).not.toBe(englisch)

    // **Und die Marke steht nicht darin**: „Validierung fehlgeschlagen" ist
    // das Wort, an dem die Maschine erkennt, und für Menschen eines zuviel.
    expect(deutsch).not.toContain('Validierung fehlgeschlagen')
  })

  // **Und wer aufs Protokoll verweist, nennt den Griff dorthin** — derselbe
  // Issue, letzter Punkt.
  //
  // # Was dieser Lauf hält, das die Go-Seite nicht hält
  //
  // **Dass der Satz den Platzhalter wirklich trägt.** `internal/app` hält,
  // dass der Schlüssel in beiden Katalogen steht und dass Go `werte.vorgang`
  // füllt — ein Satz, aus dem jemand `{vorgang}` herausnimmt, bliebe dort
  // grün, und die Kennung fiele still weg (§10).
  it('nennt die Vorgangskennung im Verweis aufs Protokoll', () => {
    const ursache = new Bindungsfehler({
      code: 'unbekannt',
      schluessel: 'unbekanntMitVorgang',
      meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
      feld: '',
      werte: { vorgang: 'abc12345' },
    })

    expect(meldungAus(ursache, kataloge.de)).toContain('abc12345')
    expect(meldungAus(ursache, kataloge.en)).toContain('abc12345')

    // **Die Gegenprobe**: ohne Kennung bleibt der Satz ohne Griff — sonst
    // stünde dort irgendwann `{vorgang}` im Klartext (§10).
    const ohne = new Bindungsfehler({
      code: 'unbekannt',
      schluessel: 'unbekannt',
      meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
      feld: '',
      werte: {},
    })

    expect(meldungAus(ohne, kataloge.de)).not.toContain('{vorgang}')
  })
})
