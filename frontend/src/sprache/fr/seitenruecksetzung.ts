// The **staged reset** — work item O4 of phase 5.2
// (#1428).

export const ruecksetzung = {
  titel: 'Réinitialisation',
  hinweis:
    'Six niveaux, du plus léger au plus lourd. Ce que chacun touche est ' +
    'indiqué au préalable — avec des chiffres et des noms.',

  artWaehlen: 'Que faut-il réinitialiser ?',

  arten: {
    layout: 'Uniquement l’apparence de la fenêtre',
    profileinstellungen: 'Les réglages de ce profil',
    nutzereinstellungen: 'Mes propres réglages',
    profil: 'Ce profil et tout ce qu’il possède',
    nutzer: 'Moi et tous mes profils',
    anwendung: 'L’application entière',
  } as Record<string, string>,

  nachsehen: 'Voir ce qui serait concerné',
  nachsehenLaeuft: 'Vérification …',

  nichtsBetroffen: 'Il n’y a ici rien qui puisse être réinitialisé.',
  gruppe: '{was} : {anzahl}',

  gruppen: {
    einstellungen: 'Réglages',
    profile: 'Profils',
    sets: 'Sets',
    karten: 'Cartes',
    decks: 'Decks',
    exemplare: 'Exemplaires de la collection',
    token: 'Jetons d’accès',
    freigaben: 'Justificatifs d’autorisation',
    zugaenge: 'Accès aux profils',
    bilder: 'Images',
  } as Record<string, string>,

  sicherung:
    'Des données seront perdues ici. Sauvegarde d’abord la base de données — dans la zone ci-dessus.',

  hindernisse: {
    offeneRueckkehr:
      'Les quantités de ce profil sont indéterminées après un retour en arrière. ' +
      'Vérifie-les et valide ce retour avant de supprimer.',
    fremdesZiel: 'Cela ne t’appartient pas.',
    letzterNutzer:
      'C’est la dernière personne de cette base de données — la supprimer ' +
      'ne laisserait plus personne.',
    laufenderAuftrag:
      'Quelque chose est encore en cours. Attends qu’il soit terminé.',
    nachweisFehlt:
      'Réinitialiser toute l’application exige une autorisation explicite ' +
      'sur ce Mac.',
  } as Record<string, string>,

  nameTippen: 'Saisis le nom pour confirmer',
  freigabeSatz:
    'Réinitialiser l’application entière nécessite ton autorisation sur cet ordinateur. Elle est valable cinq minutes, pour un seul nouveau départ.',
  freigeben: 'Accorder l’autorisation',
  freigebenLaeuft: 'Autorisation en cours …',
  freigegeben: 'Autorisé jusqu’au {zeit}.',
  kennungFuerKommandozeile:
    'La même autorisation vaut en ligne de commande (mse zuruecksetzen anwendung) :',

  tun: 'Réinitialiser',
  tunLaeuft: 'Réinitialisation …',

  bilanz: '{anzahl} réglages annulés.',
  beiseite: 'La base de données précédente se trouve à l’emplacement {ort}.',
}
