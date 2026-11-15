// The access-credential texts — split out with
// #1534.
//
// A catalogue may spread over several files as long as it arrives as one
// (§9) — the same shape as `seitenki.ts`.
export const zugaenge = {
  // The per-kind fallback — it applies where `dienstangaben` says nothing.
  //
  // “Patreon token” used to stand here and was wrong for CardTrader (#1534):
  // `domain.ZugangMarke` serves **two** services, and one string cannot be
  // right for both. A provider’s name therefore belongs in `dienstangaben`;
  // what stands here must hold for **every** service that uses the kind.
  angaben: {
    marke: 'Access token',
    schluessel: 'API key',
    team: 'Team ID',
  },

  // And per service, where the kind is not enough (#1534). Looked up, not
  // inferred: MTGJSON’s token exists only for Patreon supporters (0056),
  // CardTrader’s comes from its own account’s profile settings (0095).
  dienstangaben: {
    'mtgjson.marke': 'Patreon token',
    'cardtrader.marke': 'Token from the profile settings',
    'anthropic.schluessel': 'API key',
  },

  zugangLiegt: 'stored',
  zugangFehlt: 'missing',
  zugangSetzen: 'Store',
  zugangLoeschen: 'Delete',
  zugangLoeschenFrage: 'Delete {angabe} for {name}?',
  zugangLoeschenErklaerung:
    'The value leaves the keychain of this Mac. Only whoever issued it can give it back.',
  zugangLoeschenJa: 'Delete',
}
