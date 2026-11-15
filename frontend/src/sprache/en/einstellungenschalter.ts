// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'When switched off, an area disappears from the sidebar and the palette, and its routes are rejected on every surface. Its data stays — switching it back on restores everything.',
    hinweis:
      'A switch hides, it does not secure: to protect something, use roles and shares.',
    reife: {
      stabil: 'stable',
      vorschau: 'Preview',
      auslaufend: 'deprecated',
    },
    saetze: {
      sammeln: 'Collection, holdings and wishlists',
      spielen: 'Decks, tournaments, games and the table',
      handeln: 'Trades, shares and envelopes',
      ressourcen: 'Template packages and templates',
      ki: 'Suggestions, images, search index, accounts and cost',
      netz: 'Identity, contacts, invitations and reachability',
      markt: 'Offers, prices and accounts of the marketplaces',
      turniere: 'Tournaments with rounds, pairings and results',
      dienst: 'REST for scripts — a running service stops when switched off',
      werkzeugserver:
        'MCP over HTTP — a running server stops when switched off',
    },
    abgeschaltet: {
      titel: '{seite} is switched off',
      erklaerung:
        'Turn the area back on under Settings › Areas and functions. Its data is still there, unchanged.',
      oeffnen: 'Open areas and functions',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Collect',
  'schalter.spielen': 'Play',
  'schalter.handeln': 'Trade',
  'schalter.ressourcen': 'Resources',
  'schalter.ki': 'AI',
  'schalter.netz': 'Network and counterparts',
  'schalter.markt': 'Marketplaces and prices',
  'schalter.turniere': 'Tournaments',
  'schalter.dienst': 'Scripting service',
  'schalter.werkzeugserver': 'Tool server over HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'on',
  'schalter.sammeln.aus': 'off',
  'schalter.spielen.an': 'on',
  'schalter.spielen.aus': 'off',
  'schalter.handeln.an': 'on',
  'schalter.handeln.aus': 'off',
  'schalter.ressourcen.an': 'on',
  'schalter.ressourcen.aus': 'off',
  'schalter.ki.an': 'on',
  'schalter.ki.aus': 'off',
  'schalter.netz.an': 'on',
  'schalter.netz.aus': 'off',
  'schalter.markt.an': 'on',
  'schalter.markt.aus': 'off',
  'schalter.turniere.an': 'on',
  'schalter.turniere.aus': 'off',
  'schalter.dienst.an': 'on',
  'schalter.dienst.aus': 'off',
  'schalter.werkzeugserver.an': 'on',
  'schalter.werkzeugserver.aus': 'off',
}
