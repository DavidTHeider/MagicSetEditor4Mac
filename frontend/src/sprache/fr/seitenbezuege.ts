// Les textes des **états récupérés** sur la page des paquets — G04 (#2251).
// Leur source allemande est `de/seitenbezuege.ts`.

export const bezuege = {
  ueberschrift: 'États récupérés',
  erklaerung:
    'Un paquet récupéré reste à un commit et ne vieillit pas de lui-même. Demande aux sources s’il en existe un plus récent — ce n’est qu’alors que quelque chose passe par le réseau.',
  fragen: 'Interroger les sources',
  leer: 'Aucun paquet récupéré pour l’instant. Récupères-en un avec « Récupérer un paquet … » — tu verras ensuite ici où en est sa source.',
  bezogen: 'Récupéré : {stand}',
  geprueft: 'Vérifié : {stand}',
  ohnePruefung: 'Absent de la liste des sources vérifiées',
  pakete: '{zahl} paquets',
  paketEins: 'un paquet',
  keinPaket: 'aucun paquet',
  lage: {
    ungefragt: 'Pas encore interrogée.',
    aktuell: 'À l’état le plus récent de la source.',
    neuerEins: 'État plus récent {spitze} du {datum} — un commit d’avance.',
    neuer:
      'État plus récent {spitze} du {datum} — {vorsprung} commits d’avance.',
    abweichend:
      'La source n’est pas en avance — réinitialisée ou divergente (tête {spitze}).',
    unbekannt: 'La source n’a pas pu être interrogée : {grund}',
  },
  ansehenBezeichnung: 'Examiner la mise à jour … ({name})',
  zurueckBezeichnung: 'Revenir à l’état conservé ({name}, {stand})',
  ansehen: 'Examiner la mise à jour …',
  aufgehoben: 'Conservé : {stand}',
  zurueck: 'Revenir à l’état conservé',
  vorschauTitel: 'Mise à jour de {name}',
  vonZu: '{von} → {zu} : {vorsprung} d’avance, {rueckstand} de retard',
  paketzeile:
    '{paket} : {neu} nouveaux, {geaendert} modifiés, {entfernt} supprimés',
  keinePakete: 'Aucun paquet de modèles ne change.',
  uebrige: '{zahl} autres fichiers n’appartiennent à aucun paquet.',
  sets: 'Ces sets nomment un paquet concerné : {sets}',
  keineSets: 'Aucun set ne nomme un paquet concerné.',
  aufheben: 'L’état actuel est conservé et peut être restauré.',
  vorigeWeicht: 'L’état déjà conservé {stand} est remplacé.',
  gruende: {
    setVerliert: 'Un paquet nommé par un set perd des fichiers.',
    nichtVoraus:
      'La source n’est pas en avance — la mise à jour serait un pas en arrière ou de côté.',
    gekuerzt:
      'La source indique plus de 300 fichiers modifiés ; la liste ci-dessus est tronquée.',
  },
  bestaetigen: 'J’ai vu les raisons et je veux quand même mettre à jour.',
  ausfuehren: 'Mettre à jour vers {zu}',
  verwerfen: 'Ne pas mettre à jour',
}
