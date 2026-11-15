// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'Abgeschaltet verschwindet ein Bereich aus Leiste und Palette, und seine Wege weisen an jeder Fläche ab. Seine Daten bleiben — einschalten stellt alles wieder her.',
    hinweis:
      'Ein Schalter verbirgt, er sichert nicht: wer etwas schützen will, nimmt Rollen und Freigaben.',
    reife: {
      stabil: 'stabil',
      vorschau: 'Vorschau',
      auslaufend: 'auslaufend',
    },
    saetze: {
      sammeln: 'Sammlung, Besitz und Wunschlisten',
      spielen: 'Decks, Turniere, Partien und der Tisch',
      handeln: 'Handel, Freigaben und Umschläge',
      ressourcen: 'Vorlagenpakete und Vorlagen',
      ki: 'Vorschläge, Bilder, Suchindex, Konten und Kosten',
      netz: 'Identität, Kontakte, Einladungen und Erreichbarkeit',
      markt: 'Angebote, Preise und Konten der Marktplätze',
      turniere: 'Turniere mit Runden, Paarungen und Ergebnissen',
      dienst: 'REST für Skripte — ein laufender Dienst endet beim Abschalten',
      werkzeugserver:
        'MCP über HTTP — ein laufender Server endet beim Abschalten',
    },
    abgeschaltet: {
      titel: '{seite} ist abgeschaltet',
      erklaerung:
        'Unter Einstellungen › Bereiche und Funktionen lässt sich der Bereich wieder einschalten. Seine Daten sind unverändert da.',
      oeffnen: 'Bereiche und Funktionen öffnen',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Sammeln',
  'schalter.spielen': 'Spielen',
  'schalter.handeln': 'Handeln',
  'schalter.ressourcen': 'Ressourcen',
  'schalter.ki': 'KI',
  'schalter.netz': 'Netz und Gegenüber',
  'schalter.markt': 'Marktplätze und Preise',
  'schalter.turniere': 'Turniere',
  'schalter.dienst': 'Dienst für Skripte',
  'schalter.werkzeugserver': 'Werkzeugserver über HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'an',
  'schalter.sammeln.aus': 'aus',
  'schalter.spielen.an': 'an',
  'schalter.spielen.aus': 'aus',
  'schalter.handeln.an': 'an',
  'schalter.handeln.aus': 'aus',
  'schalter.ressourcen.an': 'an',
  'schalter.ressourcen.aus': 'aus',
  'schalter.ki.an': 'an',
  'schalter.ki.aus': 'aus',
  'schalter.netz.an': 'an',
  'schalter.netz.aus': 'aus',
  'schalter.markt.an': 'an',
  'schalter.markt.aus': 'aus',
  'schalter.turniere.an': 'an',
  'schalter.turniere.aus': 'aus',
  'schalter.dienst.an': 'an',
  'schalter.dienst.aus': 'aus',
  'schalter.werkzeugserver.an': 'an',
  'schalter.werkzeugserver.aus': 'aus',
}
