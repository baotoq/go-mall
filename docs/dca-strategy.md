# BTC Recursive Buy Strategy

Research-backed design for the trading bot's recurring BTC accumulation logic. Feeds Phase 3 ("Smart Multipliers").

## Goal

Maximize BTC accumulated per unit of fiat deployed over a multi-year horizon while staying disciplined (no manual timing, no full-cash-out).

## TL;DR

**Tiered Smart-DCA**: weekly base buy + sentiment multiplier (Fear & Greed Index) + on-chain multiplier (Mayer Multiple, MVRV-Z). Backtests show this beats fixed DCA by 22–35% and beats buy-and-hold by ~99 percentage points over 7 years.

## Backtest evidence

| Strategy | Period | Return | Source |
|----------|--------|--------|--------|
| Fixed weekly $10 DCA | 5y | +202% | SpotedCrypto |
| Fixed monthly $100 DCA | 12y (2014–2026) | +6,712% | SpotedCrypto |
| Fear-multiplier DCA (2× <25 F&G, 3× <15) | 7y | +1,145% (HODL +99pp) | SpotedCrypto |
| Value DCA vs Fixed DCA | 2020–2025 | +22–35% extra | SpotedCrypto |
| DCA started at F&G<20 | 18mo | avg +180% | SpotedCrypto |
| DCA started 2022 fear | 3y | +192% (lump-sum +33pp) | SpotedCrypto |

## Cadence

**Weekly on Monday.** dcaBTC 7-year backtest (2018–2025) shows Monday weekly buys accumulated **+14.36% more BTC** than other weekdays. Weekly is the best fee/return tradeoff vs. daily (fee drag) and monthly (worse averaging granularity).

## Multiplier stack

```
buy_amount = base * sentiment_mult * onchain_mult
clamp buy_amount to [0.25 * base, 4.0 * base]
```

### Sentiment multiplier (Fear & Greed Index)

| F&G value | Multiplier | Rationale |
|-----------|-----------|-----------|
| < 10 | 3.0 | Extreme fear, historic bottom zone |
| 10–25 | 2.0 | Fear, accumulate aggressively |
| 25–45 | 1.25 | Mild fear |
| 45–60 | 1.0 | Neutral baseline |
| 60–75 | 0.75 | Mild greed |
| 75–90 | 0.5 | Greed, slow buys |
| > 90 | 0.25 | Euphoria, near-skip |

### On-chain multiplier (confluence)

| Indicator | Threshold | Multiplier |
|-----------|-----------|-----------|
| Mayer Multiple (price / 200dMA) | < 1.0 | 2.0 |
| Mayer Multiple | 1.0–1.5 | 1.5 |
| Mayer Multiple | 1.5–2.4 | 1.0 |
| Mayer Multiple | > 2.4 | 0.5 |
| MVRV-Z Score | < 0 | × 1.5 |
| MVRV-Z Score | 0–7 | × 1.0 |
| MVRV-Z Score | > 7 | × 0.25 |

Combine Mayer × MVRV-Z modifiers, then clamp.

## Principles

1. **Confluence > single signal.** Multiplier stack only fires max when sentiment AND on-chain agree.
2. **Never sell, only modulate buy size.** Keep base buy floor (0.25×) even at extreme greed to avoid missing breakouts.
3. **Cash reserve required.** Hold 30–50% dry powder so 3× extreme-fear triggers do not exhaust budget.
4. **Discipline > timing.** Bot auto-executes, no manual override path. Backtest winners ran multi-year uninterrupted.
5. **Strongest historical buy zone**: Mayer < 1.5 AND F&G < 20.

## Implementation map

### Data sources

| Indicator | Source | Cost | Latency |
|-----------|--------|------|---------|
| Fear & Greed Index | `alternative.me/crypto/fear-and-greed-index/` (REST, free) | free | ~daily |
| Mayer Multiple | Compute from OHLC: `price / SMA(close, 200)` | free (price feed) | live |
| MVRV-Z Score | Glassnode / CoinMetrics API | paid | ~1 day |
| Price feed | Hyperliquid spot ticker | free | live |

### Bot wiring (kratos layers)

| Layer | Responsibility |
|-------|---------------|
| `internal/biz/dca` | `DCAUsecase`: compute multiplier, decide buy size, expose `RepoMetrics` + `RepoOrders` interfaces |
| `internal/data/metrics` | Fetch F&G, MVRV from external APIs; compute Mayer from local OHLC store |
| `internal/data/exchange` | Hyperliquid spot client; place market buys |
| `internal/server` | Cron trigger (Monday 00:00 UTC) → call usecase |
| `internal/conf` | Multiplier brackets, base amount, reserve cap, exchange creds |

### Decision log (mandatory)

Each weekly run must persist a structured record for replay/backtest:

```
{
  ts, fg_index, mayer, mvrv_z,
  sentiment_mult, onchain_mult, final_mult,
  base_amount, computed_amount, clamped_amount,
  reserve_remaining, btc_spot_price,
  order_id, fill_price, fee
}
```

### Guardrails

- Max single buy = 4 × base.
- Pause week if 7-day budget exhausted (avoid overdraft via 3× then 3× back-to-back).
- Hard stop if exchange returns >N consecutive errors; alert via configured channel.
- Pre-flight check: account balance ≥ computed buy + fee buffer.

## Backtest first

Before live capital:

1. Replay 2018–2026 weekly Mondays through the multiplier table using historic OHLC + F&G + Mayer (+ MVRV if available).
2. Compare to: (a) fixed weekly DCA same total fiat, (b) lump sum at start, (c) buy-and-hold variants.
3. Tune brackets to your risk tolerance — the table above is a starting point, not optimum.
4. Track max drawdown of the BTC stack value, not just terminal return.

## Risks and caveats

- **Backtest bias**: 2018–2026 is bull-dominant. Extreme multi-year bear (e.g. 2014–2015) less kind if reserve exhausts early.
- **F&G persistence**: index can stay <20 for months. Sequencing matters; consider geometric reserve drawdown rule.
- **On-chain lag**: MVRV-Z lags ~1 day; fine for weekly cadence, unfit for intraday.
- **Cycle compression**: halving cycles flattening as adoption matures — historical multipliers may overfit past behavior.
- **Hyperliquid spot liquidity**: confirm BTC spot pair has depth for your buy size; otherwise use post-only limit ladders to avoid slippage.
- **Past performance ≠ future.**

## Sources

- [SpotedCrypto — Crypto DCA Strategy Guide 2026 (202% backtest)](https://www.spotedcrypto.com/crypto-dca-dollar-cost-averaging-strategy-backtesting-guide-2026/)
- [SpotedCrypto — Bitcoin DCA 12-Year +6,712%](https://www.spotedcrypto.com/bitcoin-dca-strategy-guide-2026/)
- [SpotedCrypto — Bitcoin DCA 7-Year +1,145% Fear-Multiplier](https://www.spotedcrypto.com/bitcoin-dca-strategy-guide/)
- [SpotedCrypto — Crypto DCA Bear Market Guide](https://www.spotedcrypto.com/bitcoin-dca-strategy-bear-market/)
- [Nakamoto Notes — Mayer Multiple explained](https://nakamotonotes.com/mayer-multiple-explanation)
- [Bitbo — Mayer Multiple live chart](https://charts.bitbo.io/mayermultiple/)
- [Bitbo — MVRV-Z Score live chart](https://charts.bitbo.io/mvrv-z-score/)
- [BM Pro — Smart Bitcoin Investment Strategy](https://www.bitcoinmagazinepro.com/blog/how-to-develop-a-smart-bitcoin-investment-strategy-for-maximum-gains/)
- [checkonchain — On-chain analysis charts](https://charts.checkonchain.com/)
- [alternative.me — Fear & Greed Index API](https://alternative.me/crypto/fear-and-greed-index/)
