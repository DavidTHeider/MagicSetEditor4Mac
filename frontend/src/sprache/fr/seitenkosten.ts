// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: 'Ce qu’une tâche externe peut coûter',
  hinweis:
    'Les tâches envoyées à des services externes peuvent coûter de l’argent. Ce qui a été ' +
    'autorisé est indiqué ici — ainsi que la part déjà engagée.',

  kontoWaehlen: 'Compte',
  zeitraumWaehlen: 'Période',

  zeitraeume: {
    tag: 'Aujourd’hui',
    monat: 'Ce mois-ci',
  } as Record<string, string>,

  grenze: 'Autorisé',
  reserviert: 'Engagé',
  verbraucht: 'Dépensé',
  unklar: 'Non réglé',
  verfuegbar: 'Disponible',

  unklarHinweis:
    'Non réglé signifie que la tâche a dépassé un délai alors que le fournisseur ' +
    'pouvait déjà l’avoir acceptée. Elle compte comme dépensée jusqu’à ce que ' +
    'quelqu’un ait vérifié.',

  ohneGrenze:
    'Rien n’est autorisé pour ce compte. Sans autorisation, aucune tâche payante ' +
    'ne part.',

  lokalHinweis:
    'Il s’agit d’une limite sur ce Mac. Quiconque utilise les mêmes identifiants ' +
    'ailleurs consomme un quota dont ce journal ne sait rien.',

  grenzeSetzen: 'Modifier l’autorisation',
  betragTippen: 'Montant dans la plus petite unité (centimes pour l’EUR)',
  waehrungTippen: 'Devise',
  setzen: 'Autoriser',
  setztGerade: 'Définition en cours …',
  nullNimmtWeg: 'Zéro retire l’autorisation.',

  postenTitel: 'Les dernières tâches',
  keinePosten: 'Rien n’est encore comptabilisé.',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: 'engagé',
    abgerechnet: 'réglé',
    freigegeben: 'rien de facturé',
    unklar: 'non réglé',
  } as Record<string, string>,
}
