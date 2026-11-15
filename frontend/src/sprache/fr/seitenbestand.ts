// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: 'Base de données',
  hinweis:
    'Tout ce que cette application conserve : les sets, les cartes, les valeurs ' +
    'des champs, l’historique et les images que tu as insérées. Une ' +
    'sauvegarde contient les deux dans un seul fichier — la base de données ' +
    'et les images.',
  sichern: 'Sauvegarder la base de données …',
  waehlen: 'Restaurer une sauvegarde …',
  rueckfrage: 'Restaurer cette sauvegarde ?',
  datei: 'Fichier',
  angelegtAm: 'Base de données créée le',
  sets: 'Sets',
  karten: 'Cartes',
  unbekannt: '—',
  warnung:
    'La base de données actuelle sera remplacée. Elle n’est pas perdue : elle ' +
    'reste à côté sous le nom bestand-vor-rueckspielung.db — jusqu’à la ' +
    'prochaine restauration.',
  bilderLage: 'Images insérées',
  bilderErsetzt:
    'Cette sauvegarde contient des images. Elles remplacent les images ' +
    'actuelles ; celles-ci restent à côté sous le nom beiwerk-vorher.',
  bilderGeleert:
    'Cette sauvegarde ne contient aucune image — elle provient d’un état où ' +
    'aucune n’était insérée. Les images actuelles sont retirées ; elles ' +
    'restent à côté sous le nom beiwerk-vorher.',
  bilderUnberuehrt:
    'Cette sauvegarde est une base de données isolée de la version 0.1.0 ou ' +
    '0.2.0 et ne contient aucune image. Les images actuelles restent ' +
    'inchangées — sur un autre ordinateur, elles manqueraient complètement.',
  bestaetigen: 'Restaurer',
  abbrechen: 'Annuler',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: 'Mot de passe (facultatif)',
  kennwortWiederholen: 'Répéter le mot de passe',
  kennwortHinweis:
    'Avec un mot de passe, la sauvegarde est chiffrée et peut aussi être ouverte sur un autre ordinateur. Un mot de passe oublié ne peut pas être récupéré — la sauvegarde est alors perdue.',
  kennwortVerschieden: 'Les deux mots de passe ne correspondent pas.',
  klartext: 'Sauvegarder sans mot de passe — le fichier est alors en clair',
  klartextHinweis:
    'La collection est chiffrée. Sans mot de passe, la sauvegarde serait sans protection sur le support où elle est enregistrée.',
  gesperrt: 'Cette sauvegarde est chiffrée.',
  gesperrtKennwort: 'Mot de passe de la sauvegarde',
  entsperren: 'Ouvrir',
  verschluesselt: 'Chiffrée',
  ja: 'oui',
  nein: 'non',
  paketeEnthalten: 'Paquets personnels inclus',
  paketeFehlen: 'À télécharger de nouveau',
  keine: 'aucun',
  zurueckgespielt:
    'Restauration terminée. La base de données précédente est désormais {beiseite}.',
  rueckkehrUnveraendert: 'Ta base de données précédente est inchangée.',
  rueckkehrUnvollstaendig:
    'Le retour en arrière a également échoué : des parties de la base précédente se trouvent à côté sous les noms bestand-vor-rueckspielung.db et beiwerk-vorher. Restaure une sauvegarde ou remets-les en place à la main.',
  wiederherstellungTitel: 'La base de données n’a pas pu être ouverte',
  wiederherstellungHinweis:
    'L’application fonctionnera de nouveau dès qu’un fichier de base de ' +
    'données utilisable sera en place. Restaurer une sauvegarde est la voie ' +
    'à suivre — elle remplace le fichier qui pose problème en ce moment.',
  wiederherstellungGrund: 'Ce qui a été signalé',
  wiederherstellungHinweisHeil:
    'Le fichier de la bibliothèque est sans doute intact — le problème vient de ce qui l’entoure. Corrige-le comme décrit ci-dessous et réessaie ; il n’y a rien à remplacer.',
  erneutVersuchen: 'Réessayer',
  weiterGescheitert:
    'Nouvelle tentative — la bibliothèque ne s’ouvre toujours pas.',

  wegZurueckTitel: 'Le chemin du retour',
  wegZurueck:
    'Pour revenir en arrière, utilise « {datei} » dans le dossier de ta ' +
    'bibliothèque — {sets} set(s), {karten} cartes, étape de schéma {stand}. ' +
    'Choisis-la ci-dessous.',

  chiffreTitel: 'Chiffrement de la bibliothèque',
  chiffreAnHinweis:
    'La bibliothèque est stockée chiffrée sur le disque. Sans la clé, elle ' +
    'ne peut pas être ouverte — pas même par toi.',
  chiffreAusHinweis:
    'La bibliothèque est stockée en clair. Quiconque détient le fichier peut ' +
    'tout lire : noms, dates de naissance, adresses, photos et méthodes de ' +
    'connexion.',
  chiffreAnschalten: 'Activer le chiffrement …',
  chiffreAusschalten: 'Désactiver le chiffrement …',
  chiffreFrageAn: 'Chiffrer la bibliothèque maintenant ?',
  chiffreFrageAus: 'Désactiver de nouveau le chiffrement ?',
  chiffreWarnungAn:
    'La clé est créée maintenant et placée dans le trousseau de ce Mac. Elle ' +
    'ne peut pas être recalculée et n’existe nulle part ailleurs : si elle ' +
    'disparaît, les données disparaissent avec elle. Faites d’abord une ' +
    'sauvegarde.',
  chiffreWarnungAus:
    'Tout sera de nouveau stocké en clair, et la clé sera retirée du ' +
    'trousseau.',
  chiffreGrenzen:
    'Le chiffrement ne protège **pas** contre la falsification. Quiconque ' +
    'compare plusieurs copies du même fichier peut voir quelles parties ont ' +
    'changé. **Une sauvegarde reçoit son propre mot de passe** — ci-dessus, ' +
    'lors de la sauvegarde ; sans mot de passe, elle serait en clair, ce qui ' +
    'exige alors de choisir explicitement « Sauvegarder sans mot de passe ».',
  chiffreLaeuft: 'Réécriture de la bibliothèque …',

  platzUeberschrift: 'Espace disque',
  platzErklaerung:
    'Ce que cette application stocke et le poids que cela représente — le dossier entier, car tout ce qu’aucun poste ne réclame est compté sous **Autres**. Les **sauvegardes avant migration** sont créées à chaque démarrage qui reconstruit la bibliothèque ; ce qui reste, c’est le **chemin du retour** — la sauvegarde la plus récente que cette version peut ouvrir. Du **journal**, le fichier en cours est conservé.',
  platzDateien: '{anzahl} fichier(s)',
  platzFreigeben: 'Supprimer les plus anciens',
  platzFreigebenFuer: '{posten} : supprimer les plus anciens',
  platzFreigebenFrage: '{posten} : vraiment supprimer les plus anciens ?',
  platzFreigebenErklaerung:
    'Les fichiers plus anciens sont supprimés et ne peuvent pas être récupérés. Ce qui est nécessaire comme chemin du retour est conservé.',
  platzLoeschenJa: 'Supprimer',
  platzNamen: {
    bestand: 'Bibliothèque',
    bilder: 'Images intégrées',
    pakete: 'Paquets téléchargés',
    vorsicherungen: 'Sauvegardes avant migration',
    protokoll: 'Journal',
    sonstiges: 'Autres fichiers du dossier de l’application',
  },
  alter: {
    ueberschrift: 'Dernière sauvegarde',
    unbekannt: 'inconnue',
    unbekanntSatz:
      'Cette version n’a encore jamais sauvegardé la bibliothèque. Inconnu veut dire inconnu — ici ne s’affiche jamais « aujourd’hui » sans qu’un fichier ait été relu.',
    heute: 'aujourd’hui',
    gestern: 'il y a un jour',
    vorTagen: 'il y a {tage} jours',
    alt: '{alter} · plus de 14 jours',
    unstimmig: 'heure après l’horloge',
    satz: 'Dernière sauvegarde le {zeitpunkt} via {flaeche} dans {datei}, relue et vérifiée.',
    satzPlan:
      'Dernière sauvegarde le {zeitpunkt} par le plan dans {datei}, relue et vérifiée.',
    unstimmigSatz:
      'La dernière sauvegarde porte la date {zeitpunkt}, postérieure à l’horloge de ce Mac. Son âge ne peut pas être dit ; sauvegarde pour en être sûr.',
    flaechen: {
      fenster: 'la fenêtre',
      cli: 'la ligne de commande',
      mcp: 'MCP',
      rest: 'le service',
    },
    planSchalter: 'Sauvegarder chaque semaine dans un dossier, en garder sept',
    planHilfe:
      'Seulement après le choix d’un dossier. Seules ses propres sauvegardes sont supprimées, jamais d’autres fichiers. Il sauvegarde tant que la fenêtre est ouverte.',
    planOrdner: 'Dossier : {ordner}',
    planZuletzt: 'Dernière sauvegarde par le plan le {zeitpunkt}.',
    planNie: 'Le plan n’a encore rien sauvegardé.',
    planFehler: 'Le plan n’a pas pu sauvegarder la dernière fois : {grund}',
  },
} as const
