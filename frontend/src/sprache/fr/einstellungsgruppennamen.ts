// **Die neun Gruppen der Einstellungsseite** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge. Ein Katalog darf sich auf mehrere Dateien
// verteilen, solange er als **einer** ankommt (§9).
export const einstellungsgruppennamen = {
  gruppen: {
    allgemein: 'Général',
    bereiche: 'Domaines et fonctions',
    erscheinungsbild: 'Apparence',
    nutzer: 'Utilisateur et profil',
    fach: 'Fonctions métier',
    ki: 'IA',
    verbindungen: 'Connexions',
    dienste: 'Services',
    daten: 'Données et sauvegarde',
    diagnose: 'Diagnostic',
  },

  untertitel: {
    allgemein: 'Langue, démarrage, raccourcis',
    bereiche: 'Espaces de travail, fonctions, maturité',
    erscheinungsbild: 'Thème, police, fenêtre',
    nutzer: 'Connexion, déverrouillage',
    fach: 'Cartes, collection, jeu',
    ki: 'Fournisseurs, modèles, coûts',
    verbindungen: 'Réseau, correspondants, CLI, MCP, REST',
    dienste: 'Services de cartes, prix, marché',
    daten: 'Stockage, chemins, sauvegarde',
    diagnose: 'Journaux, mesure, technique',
  },

  abschnitte: {
    fenster: 'Fenêtre et barre latérale',
    spalten: 'Colonnes et vues',
    handel: 'Limites des échanges',
    protokoll: 'Journal',
    messung: 'Mesure',
    speicher: 'Caches et index de recherche',
    adressen: 'Adresses',
    ordner: 'Dossiers et noms de fichier',
    eingriffe: 'Revenir en arrière, ranger, réinitialiser',
    mitnehmen: 'Emporter les réglages',
  },
} as const
