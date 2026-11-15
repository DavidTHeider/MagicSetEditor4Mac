// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: 'Consulter la base de connaissances',
  rechercheErklaerung:
    'Interroge les textes des cartes, les rulings et les règles sur ce Mac — texte intégral et vecteurs combinés. Le modèle d’embedding participe, en local ; aucun modèle qui écrit des textes n’est sollicité. Rien ne quitte ce Mac, et rien n’est modifié.',

  rechercheFall: 'Sur quoi',
  rechercheFallKarten: 'Cartes',
  rechercheFallRegeln: 'Règles',
  rechercheFallRulings: 'Rulings',

  rechercheFrage: 'Question',
  rechercheSuchen: 'Consulter',
  rechercheSprache: 'Langue',
  rechercheAlleSprachen: 'Toutes',

  rechercheModus: 'Recherche effectuée',
  rechercheModusHybrid: 'texte intégral et vecteurs combinés',
  rechercheModusLexikalisch:
    'texte intégral uniquement — pas de modèle d’embedding',
  rechercheModusOhneIndex: 'pas du tout — il n’y a pas d’index de recherche',

  rechercheLeereNichtsGefunden: 'Essayer un autre mot ?',
  rechercheLeereOhneIndex:
    'Il peut être construit sous « Réglages › Diagnostic ».',
  rechercheLeereOhneModell:
    'Sans modèle d’embedding, la recherche trouve des mots, pas des similarités.',
  rechercheLeereOhneQuelle:
    'Cette question nécessite une source qui ne figure pas dans l’index.',
  rechercheLeereGesperrt: 'L’accès à cette source n’a pas été autorisé.',
  rechercheLeereBudget: 'Cette exécution a atteint sa limite.',

  rechercheHerkunft:
    'Révision {revision}, valable {datum}, récupérée {abgerufen}',
  rechercheWeitere: '{anzahl} autres sections de la même carte',
  rechercheOeffnen: 'Ouvrir la carte',
  rechercheEinschraenkung: 'Non effectué',
  rechercheStaende: 'Révisions',
  rechercheIndex: 'Index de recherche',
  rechercheMehr: 'Charger plus',

  aehnlichTitel: 'Cartes similaires',
  aehnlichErklaerung:
    'Recherche avec le texte de cette carte, pas avec son nom. La carte elle-même ne figure pas dans la liste.',
  aehnlichSuchen: 'Trouver des similaires',
  aehnlichLeer: 'Rien n’a encore été recherché.',

  freigabeTitel: 'Ce qu’un modèle recevrait',
  freigabeErklaerung:
    'Assemble ce qu’un modèle verrait — sans l’envoyer. Quiconque génère ensuite a vu au préalable avec quoi.',
  freigabeAufgabe: 'Tâche',
  freigabeAnbieter: 'Fournisseur',
  freigabeModell: 'Modèle',
  freigabeFenster: 'Fenêtre de contexte',
  freigabeReserve: 'Réserve de sortie',
  freigabeZeigen: 'Assembler',
  freigabeOrtLokal: 'Calculé sur ce Mac.',
  freigabeOrtCloud: 'Calculé chez un fournisseur externe.',
  freigabeBudget: '{verbraucht} tokens sur {grenze}, {reserve} mis de côté',
  freigabeZaehlweise: 'estimé d’après {zaehlweise}',
  freigabeZieladresse: 'Adresse cible',
  freigabeKategorien: 'Catégories de données',
  freigabePrivat: 'Du contenu personnel est inclus.',
  freigabeOeffentlich: 'Sources publiques uniquement.',
  freigabeBelege: 'Citations',
  freigabeKeineBelege:
    'Aucune citation — assemblé ainsi, un modèle ne verrait rien.',
  freigabeVeraltet: 'Ce n’est pas la révision active',
  freigabeLaeuftAb: 'Valable jusqu’au {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Rechercher aussi dans mes propres cartes',
  freigabeBestandErklaerung:
    'Sans cette case cochée, tes propres cartes restent à l’écart — la recherche ne porte alors que sur des sources publiques, et la réponse l’indique. Ce qui est transmis quitte ce Mac si le modèle s’exécute ailleurs.',
}
