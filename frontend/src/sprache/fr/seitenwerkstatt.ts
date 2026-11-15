// The card workspace: workspace, quick filter, zoom, overlays and the problem list —
// part of the `karten` catalog, split off because `seitenkarten.ts` sits at the line limit.
export const werkstatt = {
  arbeitsplatz: 'Espace de travail des cartes',
  inspectorBezeichnung: 'Champs et vérifications',
  // **Marks on cards** (#1208).
  schnellfilter: {
    bezeichnung: 'Filtrer les cartes',
    platzhalter: 'Filtrer — p. ex. type:creature -goblin',
    leeren: 'Effacer le filtre',
    hilfe:
      'Un espace sépare les parties, "…" les garde ensemble, champ:valeur lie à un champ, un - initial nie. La casse n’a pas d’importance.',
    treffer: '{gezeigt} sur {gesamt} cartes',
  },
  zoom: {
    bezeichnung: 'Taille de l’aperçu',
    anpassen: 'Ajuster',
    druck: 'Taille d’impression',
    ohneDruck:
      'Le modèle n’indique aucune résolution, il n’y a donc pas de taille d’impression.',
  },
  // **Das Kartenbild in den Finder ziehen** — AP-12 (#2369).
  ziehen: {
    beschriftung: 'Faire glisser l’image',
    hinweis:
      'Fais glisser l’image dans le Finder pour enregistrer la carte en PNG. {weg} fait la même chose.',
  },
  overlays: {
    bezeichnung: 'Superpositions',
    bildgrenzen: 'Limites d’image',
    beschnitt: 'Fond perdu',
    diagnose: 'Diagnostic',
    beschnittHinweis:
      'Ce qui se trouve à l’intérieur du trait survit à une coupe de 3 mm.',
  },
  probleme: {
    ueberschrift: 'Problèmes',
    keine: 'Aucun problème sur cette carte.',
    zumFeld: 'Aller au champ {feld}',
    ohneFeld: 'sans champ',
  },
  // **Paging** (#1290) — ported from MSE's Card menu.
  blaettern: {
    vorige: 'Carte précédente',
    naechste: 'Carte suivante',
    stelle: '{stelle} sur {gesamt}',
    kuerzelVorige: '⇞',
    kuerzelNaechste: '⇟',
  },
}
