// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: '外部任务的费用上限',
  hinweis:
    '发送给外部服务的任务可能产生费用。这里显示已批准的额度——' +
    '以及其中已被占用的部分。',

  kontoWaehlen: '账户',
  zeitraumWaehlen: '时间段',

  zeitraeume: {
    tag: '今天',
    monat: '本月',
  } as Record<string, string>,

  grenze: '已批准',
  reserviert: '已占用',
  verbraucht: '已花费',
  unklar: '未结清',
  verfuegbar: '可用',

  unklarHinweis:
    '“未结清”是指任务在提供方可能已经接受之后遇到了超时。' +
    '在有人核实之前，它按已花费计算。',

  ohneGrenze: '尚未为此账户批准任何额度。没有批准，就不会发出任何付费任务。',

  lokalHinweis:
    '这是本机上的限额。在别处使用同一凭据的人所消耗的额度，' +
    '此账本并不知情。',

  grenzeSetzen: '更改批准额度',
  betragTippen: '金额（以最小单位计，欧元为分）',
  waehrungTippen: '货币',
  setzen: '批准',
  setztGerade: '正在设置…',
  nullNimmtWeg: '填 0 即撤销批准。',

  postenTitel: '最近的任务',
  keinePosten: '尚无记账。',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: '已占用',
    abgerechnet: '已结算',
    freigegeben: '未扣费',
    unklar: '未结清',
  } as Record<string, string>,
}
