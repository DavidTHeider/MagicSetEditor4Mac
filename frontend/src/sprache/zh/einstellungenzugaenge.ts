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
    marke: '访问令牌',
    schluessel: 'API 密钥',
    team: '团队 ID',
  },

  // And per service, where the kind is not enough (#1534). Looked up, not
  // inferred: MTGJSON’s token exists only for Patreon supporters (0056),
  // CardTrader’s comes from its own account’s profile settings (0095).
  dienstangaben: {
    'mtgjson.marke': 'Patreon 令牌',
    'cardtrader.marke': '来自个人资料设置的令牌',
    'anthropic.schluessel': 'API 密钥',
  },

  zugangLiegt: '已保存',
  zugangFehlt: '缺失',
  zugangSetzen: '保存',
  zugangLoeschen: '删除',
  zugangLoeschenFrage: '要删除 {name} 的{angabe}吗？',
  zugangLoeschenErklaerung:
    '该值将从这台 Mac 的钥匙串中移除。只有其签发者才能再次提供。',
  zugangLoeschenJa: '删除',
}
