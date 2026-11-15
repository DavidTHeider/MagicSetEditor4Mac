// **Die Vorbelegungen der Dienste und Grenzen** — ausgelagert aus `api.ts`
// (Zeilengrenze, §2): Zugänge, Anbietergeheimnisse, Dienstkatalog, Grenzen.
// vorgabegrenzen sind die Grenzen, wie `domain.Vorgabegrenzen` sie nennt
// (#1429) — die Hilfe zeigt sie, und ein Test, der andere meint, belegt sie.
export function vorgabegrenzen() {
  return Promise.resolve({
    grenzen: {
      parallel: { rendern: 2, einfuhr: 1, ausfuhr: 1, bild: 4, masse: 1 },
      antwortBytes: 33554432,
      principal: { burst: 120, nachfuellen: 120, fensterSekunden: 60 },
      global: { burst: 600, nachfuellen: 600, fensterSekunden: 60 },
      reserve: 120,
      kosten: {
        lesen: 1,
        schreiben: 2,
        bild: 3,
        rendern: 5,
        ausfuhr: 8,
        einfuhr: 8,
        masse: 10,
        abbruch: 0,
      },
      kontingente: 'an',
      verbindungen: 16,
      auftraege: { laufend: 4, behalten: 16, aufbewahrungSekunden: 3600 },
      kostentafel: [
        {
          faehigkeit: 'Kartenbilder schreiben',
          klasse: 'rendern',
          kosten: 5,
          rest: ['POST /karten/{}/bild'],
          mcp: ['schreiben.karte.zeichnen'],
          cli: ['mse zeichnen'],
        },
      ],
    },
    fehler: null,
  })
}

// ohneZugaenge ist die Vorbelegung für `Zugaenge` — Phase 4.8
// (#1321).
//
// # Warum auch sie vorbelegt ist
//
// **Dieselbe Überlegung wie bei `ohneErsetzung`**: die Einstellungsseite ruft
// sie beim Aufschlagen — nicht, weil ein Test sie meint, sondern weil die
// Seite sie fragt. Ohne Vorbelegung meldete jeder Test auf dieser Seite eine
// Methode, die mit seinem Fall nichts zu tun hat.
//
// **Sie gibt eine leere Liste** — der ehrliche Fall „auf diesem Rechner liegt
// kein Zugang", und derselbe, den die Go-Seite mit leerem Schlüsselbund gibt.
// Ein Test, der die Zugänge **meint**, belegt sie selbst.
export function ohneZugaenge() {
  return Promise.resolve({ zugaenge: [], fehler: null })
}

// ohneAnbietergeheimnisse ist dieselbe Vorbelegung für die Anmeldeanbieter —
// #1401.
//
// **Aus demselben Grund wie [ohneZugaenge]**: die Einstellungsseite fragt sie
// beim Aufschlagen, und ein Test, der die Anmeldung nicht meint, soll darüber
// nicht stolpern.
export function ohneAnbietergeheimnisse() {
  return Promise.resolve({ anbieter: [], fehler: null })
}
