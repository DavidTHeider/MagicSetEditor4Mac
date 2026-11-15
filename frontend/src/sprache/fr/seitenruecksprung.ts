// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: 'Revenir en arrière',
  hinweis:
    'Quatre choses peuvent revenir en arrière, et elles ne coûtent pas ' +
    'la même chose. Deux d’entre elles, cette application ne peut pas les ' +
    'faire elle-même — elle indique alors ici qui s’en charge.',

  arten: {
    appfassung: 'L’application',
    schema: 'La structure de la bibliothèque',
    daten: 'Les données',
    pakete: 'Les paquets de modèles',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'Une personne remplace le bundle dans le Finder une fois que cette ' +
      'application ne s’exécute plus. Un programme qui se remplacerait lui-même ' +
      'pendant son exécution aurait besoin d’un second processus disposant ' +
      'du droit d’écriture sur « Applications ».',
    nurVorwaerts:
      'Il n’existe aucun chemin en arrière à travers les migrations. Le ' +
      'chemin du retour est une sauvegarde : restaure l’état d’alors.',
    unbekannteArt: 'L’application ne connaît pas ce type.',
  } as Record<string, string>,

  gehtHier: 'Cette application peut le faire elle-même.',

  vergleichTitel: 'Ce que coûte une restauration',
  datenstand: 'État restauré',
  spalteGegenstand: 'Élément',
  spalteSicherung: 'Dans la sauvegarde',
  spalteJetzt: 'Maintenant',
  spalteVerlust: 'Sera perdu',

  gegenstaende: {
    sets: 'Sets',
    karten: 'Cartes',
    decks: 'Decks',
    exemplare: 'Exemplaires',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'Ce qui est perdu ici, tu peux d’abord l’exporter : un set via « Enregistrer sous », la collection via « Exporter ».',

  unklarTitel: 'Incertain ensuite',
  unklarHinweis:
    'Ces exemplaires seront de nouveau dans ta collection après la ' +
    'restauration. Personne ne sait si tu les as encore — ils ont pu ' +
    'être cédés depuis. Ils restent hors des échanges jusqu’à ce que tu ' +
    'aies vérifié ; rien n’est déduit.',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'La comparaison des collections n’était pas disponible — cette ' +
    'sauvegarde ou ta bibliothèque ne la fournit pas. Passe ta ' +
    'collection en revue à la main après la restauration.',

  rueckkehrenTitel: 'Restaurations enregistrées',
  keineRueckkehren:
    'Aucune pour l’instant. Dès que tu restaures une sauvegarde, elle ' +
    'apparaît ici — avec l’état vers lequel elle est revenue et ce qu’elle ' +
    'a laissé incertain.',
  rueckkehrZeile: 'Le {zurueckgespielt}, retour à l’état du {datenstand}',
  ausDatei: 'depuis {quelle}',
  offen: 'En suspens',
  quittiert: 'Vérifié le {geklaertAm}',
  betrifft: 'concerne {menge} exemplaires',
  quittieren: 'Vérifié',
  quittierenLaeuft: 'Enregistrement…',
  quittierenHinweis:
    'Tu indiques ainsi avoir vérifié à la main les exemplaires ' +
    'concernés. Cela ne déduit rien.',
}
