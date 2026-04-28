/* Listing V2 — top filter bar + 4col grid */

function ListingV2({ mobile }) {
  const products = [...PRODUCTS, ...PRODUCTS].slice(0, 12).map((p, i) => ({ ...p, _k: i, price: p.price + (i % 4) * 20 }));

  if (mobile) {
    return (
      <div className="frame-scroll">
        <div style={{ position: "sticky", top: 0, zIndex: 30, background: "rgba(251,251,253,0.9)", backdropFilter: "saturate(180%) blur(20px)", borderBottom: "1px solid var(--hairline)", padding: "10px 14px", display: "flex", alignItems: "center", gap: 10 }}>
          <IconChevL size={20} />
          <span style={{ flex: 1, fontSize: 15, fontWeight: 600 }}>Audio</span>
          <IconSearch size={18} />
          <IconBag size={18} />
        </div>
        <div style={{ padding: "14px 16px 6px" }}>
          <h1 style={{ fontSize: 28, fontWeight: 600, letterSpacing: "-0.02em", margin: 0 }}>Audio.</h1>
          <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 2 }}>148 products</div>
        </div>
        <div style={{ display: "flex", gap: 6, padding: "8px 16px", overflow: "auto", borderBottom: "1px solid var(--hairline)" }}>
          <span className="chip active" style={{ background: "var(--ink)", color: "white", borderColor: "var(--ink)" }}><IconFilter size={12} /> Filter · 2</span>
          <span className="chip">Sort: Popular <IconChevD size={10} /></span>
          <span className="chip" style={{ color: "var(--accent)", borderColor: "var(--accent)" }}>Wireless ✕</span>
          <span className="chip" style={{ color: "var(--accent)", borderColor: "var(--accent)" }}>$100-500 ✕</span>
        </div>
        <div style={{ padding: 16, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
          {products.slice(0, 8).map(p => (
            <div key={p._k}>
              <div style={{ position: "relative", background: p.hero, borderRadius: 12 }}>
                <ProductImage product={p} style={{ borderRadius: 12 }} />
                <span style={{ position: "absolute", top: 8, right: 8, background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 999, width: 28, height: 28, display: "flex", alignItems: "center", justifyContent: "center" }}><IconHeart size={14} /></span>
              </div>
              <div style={{ fontSize: 12, color: "var(--ink-2)", marginTop: 6 }}>{p.category}</div>
              <div style={{ fontSize: 14, fontWeight: 500 }}>{p.name}</div>
              <div style={{ fontSize: 12, color: "var(--ink-2)", marginTop: 2, display: "flex", gap: 6 }}>
                <Rating value={p.rating} size={11} /> · <span>{fmt(p.price)}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="frame-scroll">
      <Nav active="audio" />
      <div style={{ background: "var(--surface-2)", borderBottom: "1px solid var(--hairline)", padding: "10px 0", fontSize: 12, color: "var(--ink-2)" }}>
        <div style={{ maxWidth: 1280, margin: "0 auto", padding: "0 40px" }}>
          Home &nbsp;›&nbsp; <span style={{ color: "var(--ink)" }}>Audio</span>
        </div>
      </div>
      <section style={{ padding: "40px 40px 18px", maxWidth: 1280, margin: "0 auto" }}>
        <div style={{ display: "flex", alignItems: "baseline", gap: 14 }}>
          <h1 style={{ fontSize: 56, fontWeight: 600, letterSpacing: "-0.025em", margin: 0, lineHeight: 1 }}>Audio.</h1>
          <div style={{ fontSize: 19, color: "var(--ink-2)" }}>148 products, built to listen.</div>
        </div>
      </section>

      {/* Filter bar */}
      <div style={{ borderTop: "1px solid var(--hairline)", borderBottom: "1px solid var(--hairline)", background: "rgba(251,251,253,0.8)", backdropFilter: "blur(12px)", position: "sticky", top: 0, zIndex: 20 }}>
        <div style={{ maxWidth: 1280, margin: "0 auto", padding: "14px 40px", display: "flex", alignItems: "center", gap: 10 }}>
          {["Category","Brand","Price","Features","Color","Connectivity","Rating"].map((f, i) => (
            <span key={f} className="chip" style={i < 2 ? { borderColor: "var(--ink)", color: "var(--ink)" } : {}}>
              {f}{i < 2 ? " · 1" : ""} <IconChevD size={12} />
            </span>
          ))}
          <div style={{ flex: 1 }} />
          <span style={{ fontSize: 13, color: "var(--ink-2)", marginRight: 8 }}>Sort by</span>
          <span className="chip">Popular <IconChevD size={12} /></span>
        </div>
      </div>

      {/* Active filters strip */}
      <div style={{ maxWidth: 1280, margin: "0 auto", padding: "14px 40px", display: "flex", gap: 8, alignItems: "center", fontSize: 13 }}>
        <span style={{ color: "var(--ink-2)" }}>Showing</span>
        <span className="chip" style={{ color: "var(--accent)", borderColor: "rgba(0,113,227,0.3)", background: "rgba(0,113,227,0.06)" }}>Wireless ✕</span>
        <span className="chip" style={{ color: "var(--accent)", borderColor: "rgba(0,113,227,0.3)", background: "rgba(0,113,227,0.06)" }}>$100 – $500 ✕</span>
        <a style={{ color: "var(--accent)", marginLeft: 4, fontSize: 13 }}>Clear all</a>
      </div>

      {/* Grid */}
      <section style={{ maxWidth: 1280, margin: "0 auto", padding: "10px 40px 60px" }}>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 20 }}>
          {products.map(p => (
            <article key={p._k} style={{ cursor: "pointer" }}>
              <div style={{ position: "relative", background: p.hero, borderRadius: 18, overflow: "hidden" }}>
                <ProductImage product={p} style={{ borderRadius: 0 }} />
                <button style={{ position: "absolute", top: 12, right: 12, background: "rgba(255,255,255,0.85)", border: "none", backdropFilter: "blur(8px)", borderRadius: 999, width: 36, height: 36, display: "inline-flex", alignItems: "center", justifyContent: "center", cursor: "pointer" }}>
                  <IconHeart size={18} />
                </button>
                {p._k % 5 === 0 && <span style={{ position: "absolute", top: 12, left: 12, fontSize: 11, fontWeight: 600, padding: "3px 10px", background: "var(--ink)", color: "white", borderRadius: 999 }}>New</span>}
              </div>
              <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 12 }}>{p.category}</div>
              <div style={{ fontSize: 17, fontWeight: 500, letterSpacing: "-0.015em", marginTop: 2 }}>{p.name} {p.tagline && <span style={{ fontWeight: 400, color: "var(--ink-2)" }}> — {p.tagline}</span>}</div>
              <div style={{ fontSize: 14, color: "var(--ink-2)", marginTop: 6, display: "flex", gap: 10, alignItems: "center" }}>
                <Rating value={p.rating} count={p.reviews} />
              </div>
              <div style={{ fontSize: 17, fontWeight: 500, marginTop: 6 }}>From {fmt(p.price)}</div>
            </article>
          ))}
        </div>
        <div style={{ display: "flex", justifyContent: "center", marginTop: 32 }}>
          <Btn kind="outline" size="md">Show more ›</Btn>
        </div>
      </section>
    </div>
  );
}

window.ListingV2 = ListingV2;
