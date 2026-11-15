// The **data hygiene** texts — work item O8 of phase 5.2
// (#1432).

export const datenpflege = {
  titel: 'Rangement',
  hinweis:
    'Au fil du temps, une bibliothèque accumule des lignes qui ne pointent ' +
    'plus vers rien. Voici de quoi il s’agit — et lesquelles peuvent être ' +
    'supprimées sans risque.',

  nachsehen: 'Vérifier',
  nachsehenLaeuft: 'Vérification …',
  sauber: 'Rien trouvé. La bibliothèque est en ordre.',
  gueltigbis: 'Cette vue d’ensemble est valable jusqu’à {zeit}.',

  arten: {
    kartenwerteOhneKarte: 'Valeurs de champ sans leur carte',
    setangabenOhneSet: 'Informations de set sans leur set',
    suchzeilenOhneKarte: 'Lignes de recherche sans leur carte',
    deckplaetzeOhneDeck: 'Emplacements de deck sans leur deck',
    zuordnungenOhneExemplar: 'Affectations de deck sans leur exemplaire',
    verlaufOhneSet: 'Entrées d’historique de sets qui n’existent plus',
    abgelaufeneNachweise: 'Justificatifs d’approbation expirés',
    abgelaufeneToken: 'Jetons d’accès expirés',
    moeglicheDubletten: 'Ressemble à des doublons',
    offeneRueckkehren: 'Restaurations ouvertes',
  } as Record<string, string>,

  klassen: {
    sicher: 'Peut être supprimé sans risque',
    entscheidung: 'Tu décides',
    nurInformieren: 'Ne sera pas touché',
  } as Record<string, string>,

  klassenhinweise: {
    sicher:
      'Ces lignes ne pointent vers rien et ne retiennent rien. Elles sont ' +
      'présélectionnées.',
    entscheidung:
      'Cela ressemble à un reste — et peut être un justificatif. Regarde ' +
      'les exemples avant de le choisir.',
    nurInformieren:
      'Ce n’est pas supprimé. Des numéros de carte identiques ne sont pas ' +
      'des doublons — l’état, la langue, la finition, l’acquisition et ' +
      'l’emplacement de stockage les distinguent — et une restauration ' +
      'ouverte est un justificatif.',
  } as Record<string, string>,

  anzahl: '{anzahl} endroits dans la bibliothèque',
  beispieleTitel: 'Par exemple',

  aufraeumen: 'Supprimer la sélection',
  aufraeumenLaeuft: 'Suppression en cours …',
  aufraeumenHinweis:
    'La bibliothèque est d’abord sauvegardée. Ensuite, elle est vérifiée de nouveau.',

  bilanzTitel: 'Ce qui s’est passé',
  gesichert: 'Sauvegardé dans {ort}',

  ausgaenge: {
    erledigt: 'Supprimé',
    konflikt: 'Modifié depuis la vue d’ensemble — non touché',
    schonWeg: 'Déjà supprimé',
    uebersprungen: 'Ignoré',
    gescheitert: 'Échec',
    bereit: 'Prêt',
  } as Record<string, string>,

  bilanzZeile: '{art} : {ausgang} ({anzahl})',
  bilanzStaende: 'Avant : {vorher} constats ; après : {nachher}.',
}
