// **Bereiche und Funktionen** — OPS-10
// (#1908).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge; die Namen und Werte werden dort hineingestreut.
// Ein Bereich heißt hier wie sein Eintrag in der Leiste.
export const seitenschalter = {
  schalter: {
    erklaerung:
      'Désactivé, un domaine disparaît de la barre latérale et de la palette, et ses accès sont refusés sur chaque surface. Ses données restent — le réactiver rétablit tout.',
    hinweis:
      'Un interrupteur masque, il ne protège pas : pour protéger quelque chose, utilise les rôles et les partages.',
    reife: {
      stabil: 'stable',
      vorschau: 'Aperçu',
      auslaufend: 'en fin de vie',
    },
    saetze: {
      sammeln: 'Collection, possessions et listes de souhaits',
      spielen: 'Decks, tournois, parties et la table',
      handeln: 'Échanges, partages et enveloppes',
      ressourcen: 'Paquets de modèles et modèles',
      ki: 'Suggestions, images, index de recherche, comptes et coûts',
      netz: 'Identité, contacts, invitations et joignabilité',
      markt: 'Offres, prix et comptes des places de marché',
      turniere: 'Tournois avec rondes, appariements et résultats',
      dienst:
        'REST pour les scripts — un service en cours s’arrête à la désactivation',
      werkzeugserver:
        'MCP via HTTP — un serveur en cours s’arrête à la désactivation',
    },
    abgeschaltet: {
      titel: '{seite} est désactivé',
      erklaerung:
        'Réactive le domaine dans Réglages › Domaines et fonctions. Ses données sont toujours là, inchangées.',
      oeffnen: 'Ouvrir domaines et fonctions',
    },
  },
} as const

// schalternamen sind die Beschriftungen der zehn Schalter.
export const schalternamen = {
  'schalter.sammeln': 'Collectionner',
  'schalter.spielen': 'Jouer',
  'schalter.handeln': 'Échanger',
  'schalter.ressourcen': 'Ressources',
  'schalter.ki': 'IA',
  'schalter.netz': 'Réseau et correspondants',
  'schalter.markt': 'Places de marché et prix',
  'schalter.turniere': 'Tournois',
  'schalter.dienst': 'Service pour scripts',
  'schalter.werkzeugserver': 'Serveur d’outils via HTTP',
}

// schalterwerte sind an und aus, je Schalter.
export const schalterwerte = {
  'schalter.sammeln.an': 'activé',
  'schalter.sammeln.aus': 'désactivé',
  'schalter.spielen.an': 'activé',
  'schalter.spielen.aus': 'désactivé',
  'schalter.handeln.an': 'activé',
  'schalter.handeln.aus': 'désactivé',
  'schalter.ressourcen.an': 'activé',
  'schalter.ressourcen.aus': 'désactivé',
  'schalter.ki.an': 'activé',
  'schalter.ki.aus': 'désactivé',
  'schalter.netz.an': 'activé',
  'schalter.netz.aus': 'désactivé',
  'schalter.markt.an': 'activé',
  'schalter.markt.aus': 'désactivé',
  'schalter.turniere.an': 'activé',
  'schalter.turniere.aus': 'désactivé',
  'schalter.dienst.an': 'activé',
  'schalter.dienst.aus': 'désactivé',
  'schalter.werkzeugserver.an': 'activé',
  'schalter.werkzeugserver.aus': 'désactivé',
}
