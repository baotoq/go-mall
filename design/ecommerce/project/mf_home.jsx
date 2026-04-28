/* Home V2 — Editorial single-hero, Apple-style */

function HomeV2({ mobile }) {
  const xr7 = PRODUCTS[0];
  if (mobile) {
    return (
      <div className="frame-scroll">
        {/* Mobile nav */}
        <div style={{ position: "sticky", top: 0, zIndex: 30, background: "rgba(251,251,253,0.85)", backdropFilter: "saturate(180%) blur(20px)", borderBottom: "1px solid var(--hairline)", padding: "10px 16px", display: "flex", alignItems: "center" }}>
          <IconMenu size={20} />
          <span style={{ flex: 1, textAlign: "center", fontSize: 14, fontWeight: 500 }}>◯</span>
          <IconBag size={20} />
        </div>
        {/* Hero */}
        <div style={{ background: xr7.hero, padding: "32px 20px 24px", textAlign: "center" }}>
          <div style={{ fontSize: 11, fontWeight: 600, color: "var(--accent)", letterSpacing: "0.02em", textTransform: "uppercase" }}>New</div>
          <h1 style={{ fontSize: 44, fontWeight: 600, letterSpacing: "-0.025em", lineHeight: 1.02, margin: "8px 0 6px" }}>XR-7.</h1>
          <div style={{ fontSize: 19, color: "var(--ink-2)", letterSpacing: "-0.015em" }}>Sound you can feel.</div>
          <div style={{ marginTop: 10, display: "flex", gap: 14, justifyContent: "center", fontSize: 14 }}>
            <a style={{ color: "var(--accent)" }}>Buy ›</a>
            <a style={{ color: "var(--accent)" }}>Learn more ›</a>
          </div>
          <div style={{ marginTop: 18 }}>
            <ProductImage product={xr7} style={{ width: "80%", margin: "0 auto" }} />
          </div>
        </div>
        {/* Secondary tile */}
        <div style={{ background: PRODUCTS[2].hero, padding: "28px 20px", textAlign: "center", marginTop: 10 }}>
          <h2 style={{ fontSize: 32, fontWeight: 600, letterSpacing: "-0.02em", margin: 0 }}>Luma 16.</h2>
          <div style={{ fontSize: 17, color: "var(--ink-2)" }}>Pro power, pro battery.</div>
          <div style={{ marginTop: 8, fontSize: 14, display: "flex", gap: 14, justifyContent: "center" }}>
            <a style={{ color: "var(--accent)" }}>Buy ›</a>
            <a style={{ color: "var(--accent)" }}>Learn more ›</a>
          </div>
          <ProductImage product={PRODUCTS[2]} style={{ width: "100%", marginTop: 14 }} />
        </div>
        <div style={{ padding: 16 }}>
          <div style={{ fontSize: 22, fontWeight: 600, letterSpacing: "-0.015em" }}>Shop the latest.</div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, marginTop: 12 }}>
            {PRODUCTS.slice(3,7).map(p => (
              <div key={p.id} style={{ background: p.hero, borderRadius: 16, padding: 10, textAlign: "center" }}>
                <ProductImage product={p} style={{ width: "100%" }} />
                <div style={{ fontSize: 13, fontWeight: 500, marginTop: 6 }}>{p.name}</div>
                <div style={{ fontSize: 11, color: "var(--ink-2)" }}>From {fmt(p.price)}</div>
              </div>
            ))}
          </div>
        </div>
        <div style={{ padding: "16px", color: "var(--ink-3)", fontSize: 11, textAlign: "center", borderTop: "1px solid var(--hairline)", marginTop: 10 }}>Copyright © 2026 brand. All rights reserved.</div>
      </div>
    );
  }
  return (
    <div className="frame-scroll">
      <Nav active="store" />
      {/* Sub-nav (product ribbon) */}
      <div style={{ height: 44, borderBottom: "1px solid var(--hairline)", background: "var(--surface-2)", display: "flex", alignItems: "center", justifyContent: "center", gap: 30, fontSize: 12, color: "var(--ink-2)" }}>
        <span style={{ color: "var(--ink)", fontWeight: 500 }}>Shop the Latest</span>
        <span>New Arrivals</span>
        <span>Deals</span>
        <span>Apple Trade In</span>
        <span>Financing</span>
        <span>Help Me Choose</span>
      </div>

      {/* Hero — editorial full bleed */}
      <section style={{ background: xr7.hero, padding: "56px 20px 40px", textAlign: "center", position: "relative" }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: "var(--accent)", letterSpacing: "0.02em", textTransform: "uppercase" }}>New</div>
        <h1 style={{ fontSize: 80, fontWeight: 600, letterSpacing: "-0.03em", lineHeight: 1, margin: "10px 0 6px" }}>XR-7.</h1>
        <div style={{ fontFamily: "var(--font-serif)", fontStyle: "italic", fontSize: 26, color: "var(--ink-2)", letterSpacing: "-0.01em" }}>Sound you can feel.</div>
        <div style={{ marginTop: 16, display: "flex", gap: 22, justifyContent: "center", fontSize: 17 }}>
          <a style={{ color: "var(--accent)" }}>Buy ›</a>
          <a style={{ color: "var(--accent)" }}>Learn more ›</a>
        </div>
        <div style={{ marginTop: 30, maxWidth: 420, margin: "30px auto 0" }}>
          <ProductImage product={xr7} style={{ borderRadius: 0 }} />
        </div>
      </section>

      {/* Second hero tile */}
      <section style={{ background: PRODUCTS[2].hero, padding: "48px 20px 0", textAlign: "center", marginTop: 10 }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: "var(--ink-2)", letterSpacing: "0.02em", textTransform: "uppercase" }}>Now with M3 Pro</div>
        <h2 style={{ fontSize: 64, fontWeight: 600, letterSpacing: "-0.025em", lineHeight: 1.02, margin: "10px 0 6px" }}>Luma 16.</h2>
        <div style={{ fontSize: 23, color: "var(--ink-2)" }}>Pro power. Pro battery. Pro everything.</div>
        <div style={{ marginTop: 14, display: "flex", gap: 22, justifyContent: "center", fontSize: 17 }}>
          <a style={{ color: "var(--accent)" }}>Buy from {fmt(PRODUCTS[2].price)} ›</a>
          <a style={{ color: "var(--accent)" }}>Learn more ›</a>
        </div>
        <div style={{ maxWidth: 720, margin: "30px auto 0" }}>
          <ProductImage product={PRODUCTS[2]} size="wide" style={{ borderRadius: 0 }} />
        </div>
      </section>

      {/* Two-up split */}
      <section style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, marginTop: 10 }}>
        {[PRODUCTS[3], PRODUCTS[5]].map(p => (
          <div key={p.id} style={{ background: p.hero, padding: "40px 20px 24px", textAlign: "center" }}>
            <h3 style={{ fontSize: 42, fontWeight: 600, letterSpacing: "-0.025em", margin: 0 }}>{p.name}.</h3>
            <div style={{ fontSize: 18, color: "var(--ink-2)" }}>{p.tagline}</div>
            <div style={{ marginTop: 10, display: "flex", gap: 18, justifyContent: "center", fontSize: 14 }}>
              <a style={{ color: "var(--accent)" }}>Buy ›</a>
              <a style={{ color: "var(--accent)" }}>Learn more ›</a>
            </div>
            <div style={{ maxWidth: 260, margin: "20px auto 0" }}>
              <ProductImage product={p} />
            </div>
          </div>
        ))}
      </section>

      {/* Shop by section */}
      <section style={{ padding: "48px 40px", background: "var(--surface)" }}>
        <h2 style={{ fontSize: 32, fontWeight: 600, letterSpacing: "-0.02em", margin: "0 0 20px" }}>Shop the latest.</h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 16 }}>
          {PRODUCTS.slice(0, 4).map(p => (
            <div key={p.id} style={{ background: p.hero, borderRadius: 18, padding: 18, color: "var(--ink)", cursor: "pointer" }}>
              <div style={{ fontSize: 19, fontWeight: 500 }}>{p.name}</div>
              <div style={{ fontSize: 14, color: "var(--ink-2)", marginTop: 2 }}>From {fmt(p.price)}</div>
              <div style={{ marginTop: 4, display: "flex", gap: 12, fontSize: 13 }}>
                <span style={{ color: "var(--accent)" }}>Buy ›</span>
                <span style={{ color: "var(--accent)" }}>Learn more ›</span>
              </div>
              <div style={{ marginTop: 12 }}><ProductImage product={p} /></div>
            </div>
          ))}
        </div>
      </section>

      {/* Footer */}
      <footer style={{ background: "var(--surface-3)", padding: "24px 40px", fontSize: 12, color: "var(--ink-3)", borderTop: "1px solid var(--hairline)" }}>
        <div style={{ maxWidth: 1024, margin: "0 auto" }}>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 24, paddingBottom: 16, borderBottom: "1px solid var(--hairline)" }}>
            {["Shop and Learn","Account","About","Support"].map(h => (
              <div key={h}>
                <div style={{ color: "var(--ink)", fontWeight: 500, fontSize: 12, marginBottom: 6 }}>{h}</div>
                {["Store","Mac","iPad","Accessories"].map(l => <div key={l} style={{ marginBottom: 4 }}>{l}</div>)}
              </div>
            ))}
          </div>
          <div style={{ paddingTop: 14 }}>Copyright © 2026 brand Inc. All rights reserved. Privacy · Terms · Legal · Site Map</div>
        </div>
      </footer>
    </div>
  );
}

window.HomeV2 = HomeV2;
