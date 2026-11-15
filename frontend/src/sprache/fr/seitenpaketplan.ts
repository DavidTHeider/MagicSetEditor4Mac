// The texts of the **package inspector and the plan before a package
// change** — DOM-01 (#1696). Their German source is `de/seitenpaketplan.ts`.

export const paketplan = {
  arbeitsplatz: 'Espace de travail des paquets',
  navigator: 'Recherche et filtres',
  inspector: 'Paquet',
  zentrum: 'Paquets',

  herkunft: 'Origine',
  herkunftAlle: 'Toute origine',
  herkunftsarten: {
    mitgeliefert: 'fournis',
    bezogen: 'récupérés',
    uebernommen: 'importés',
    eigen: 'personnels',
  },

  spalteStand: 'État',
  standEingebunden: 'installé',
  standGeaendert: 'modifié localement',

  inspectorLeer:
    'Choisis un paquet dans la liste — son origine, sa version, ses dépendances et ce qu’une modification provoquerait s’affichent alors ici.',
  kopf: '{art} · version {version}',
  kopfOhneFassung: '{art} · sans version',
  eingebunden: 'Installé',
  keinSet: 'Aucun set de ce profil ne l’utilise.',
  einSet: 'Un set l’utilise.',
  vieleSets: '{anzahl} sets l’utilisent.',
  setsUnbekannt:
    'Les sets qui en ont besoin ne figurent dans aucun en-tête de set.',
  geaendert:
    'Modifié localement depuis son installation. Le remplacer demande d’abord confirmation ; la modification est conservée comme version précédente.',
  woher: 'Origine',
  woherMitgeliefert: 'Fourni avec l’application.',
  woherBezogen: 'Récupéré depuis {herkunft}.',
  woherUebernommen: 'Importé depuis {herkunft}.',
  woherEigen: 'Créé ici ou placé à la main dans le dossier des paquets.',
  pruefsumme: 'Somme de contrôle',
  lizenz: 'Licence',
  lizenzUnbekannt:
    'non indiquée — un paquet MSE ne porte pas de champ de licence ; les contenus appartiennent à leurs ayants droit',
  ort: 'Emplacement',
  haengtAb: 'Dépend de',
  gebrauchtVon: 'Nécessaire à',
  nichts: 'rien',

  einbinden: 'Installer un paquet…',
  beziehen: 'Récupérer un paquet…',
  beziehenTitel: 'Récupérer un paquet',
  loeschen: 'Supprimer…',
  zurueck: 'Restaurer la version précédente…',
  weitergeben: 'Partager comme installateur…',
  weitergegeben: 'Enregistré comme installateur : {datei}',

  titel: {
    loeschen: 'Supprimer {paket}',
    einbinden: 'Installer {paket}',
    ersetzen: 'Remplacer {paket}',
    zurueck: 'Rétablir {paket} à sa version précédente',
  },
  fassung: 'Version {von} → {zu}',
  ohneFassung: 'sans version',
  sets: 'Sets concernés',
  keineSets: 'Aucun set de ce profil n’est concerné.',
  konflikte: 'Non satisfait ensuite',
  keineKonflikte: 'Aucune dépendance ne casse.',
  planGeaendert: 'Le paquet a été modifié localement depuis son installation.',
  ueberschreiben:
    'Remplacer la modification locale — elle est conservée comme version précédente',
  nichtMoeglich: {
    loeschen:
      'Ce paquet ne se trouve pas dans le dossier géré. Seul ce que cette application a installé est supprimé ici.',
    einbinden: 'Ce paquet ne peut pas être installé ici.',
    ersetzen: 'Ce paquet ne peut pas être remplacé ici.',
    zurueck:
      'Il n’y a pas de version précédente. Il en existe une dès que tu installes une version plus récente par-dessus ce paquet.',
  },
  bestaetigen: {
    loeschen: 'Supprimer',
    einbinden: 'Installer',
    ersetzen: 'Remplacer',
    zurueck: 'Rétablir',
  },
  verwerfen: 'Ne rien modifier',
  bedarf: {
    ueberschrift: 'Paquets manquant à ce set',
    erklaerung:
      'Sans eux, l’aperçu dessine un plan de secours ou omet ce que le paquet apporte. Installe le paquet depuis un dossier — ou cherche-le parmi les paquets pour le récupérer ou l’importer depuis MSE2.',
    suchen: 'Chercher parmi les paquets',
  },
  quelle: 'Depuis {quelle}',
  keineRechte:
    'Un paquet apporte des modèles et du MSE-Script. Le script n’a accès ni aux fichiers, ni au réseau, ni aux programmes, et ce dont le paquet a besoin n’est pas récupéré avec lui — cela figure plus haut comme conflit.',
}
