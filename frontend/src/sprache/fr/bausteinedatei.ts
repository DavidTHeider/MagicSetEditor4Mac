export const dateiansicht = {
  ueberschrift: 'Le fichier',
  laedt: 'Examen du fichier …',
  arten: {
    paket: 'Paquet MSE',
    bild: 'Image',
    vektor: 'Image vectorielle',
    pdf: 'PDF',
    text: 'Texte',
    archiv: 'Archive',
    unbekannt: 'Type inconnu',
  },
  name: 'Nom',
  art: 'Type',
  groesse: 'Taille',
  geaendert: 'Modifié',
  ort: 'Emplacement',
  masse: '{breite} × {hoehe} points',
  seiten: '{anzahl} pages',
  eintraege: '{anzahl} éléments',
  gekuerzt: 'Seul le début est affiché — le fichier se poursuit.',
  unsicher:
    '{anzahl} éléments mènent hors du fichier ou sont des liens. Ils sont comptés, mais jamais ouverts.',
  unsicherEintrag: 'non ouvert',
  verschluesselt:
    'Une sauvegarde protégée par mot de passe. Son contenu n’apparaît qu’à la restauration.',
  zuGross: 'Trop volumineux pour un aperçu. Ouvre-le avec son application.',
  unlesbar:
    'Le contenu est illisible. Le nom, la taille et la date sont corrects.',
  keineVorschau:
    'Il n’existe pas d’aperçu pour ce type : il pourrait s’agir d’un programme. Affiche-le dans le Finder si tu en as besoin.',
  imFinder: 'Afficher dans le Finder',
  oeffnen: 'Ouvrir',
  imFinderFuer: 'Afficher {name} dans le Finder',
  oeffnenFuer: 'Ouvrir {name} avec son application',
  ansehen: 'Voir',
  ansehenFuer: 'Voir {name}',
  zuklappen: 'Fermer l’aperçu',
} as const
