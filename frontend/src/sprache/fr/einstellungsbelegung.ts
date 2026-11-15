// The mapping of ⌘0 to ⌘9 — UX-02.AK-03 (#1674), split out like the German file.
export const einstellungsbelegung = {
  belegungUeberschrift: 'Attribution de ⌘0 à ⌘9',
  belegungHinweis:
    'La page qu’ouvre un chiffre avec ⌘. Par défaut, ce sont les espaces de la barre latérale ; « Comme avant 0.6.0 » rétablit l’attribution jusqu’à 0.5.0, dont le tableau figure dans le journal des modifications.',
  belegungFuer: 'Page pour {kuerzel}',
  belegungFrei: '—',
  belegungVor06: 'Comme avant 0.6.0',
  belegungVorgabe: 'Revenir à la valeur par défaut',
  kuerzel: {
    palette: 'Ouvrir la palette de commandes',
    suche: 'Rechercher dans le set',
    zurueck: 'Annuler une modification',
    vor: 'Rétablir une modification',
    sichern: 'Enregistrer le set',
    sichernUnter: 'Enregistrer le set sous …',
    oeffnen: 'Ouvrir un set …',
    neu: 'Créer un nouveau set',
    schliessen: 'Fermer le set',
    drucken: 'Aller à la feuille d’impression',
    navigator: 'Afficher ou masquer le navigateur',
    inspector: 'Afficher ou masquer l’inspecteur',

    // Back and forward between pages — UX-03.AK-02 (#1675).
    seiteZurueck: 'Page précédente',
    seiteVor: 'Page suivante',
  },
} as const
