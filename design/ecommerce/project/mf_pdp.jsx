/* PDP V1 — classic gallery + buy box */

function PDPV1({ mobile }) {
  const p = PRODUCTS[0];
  const [selected, setSelected] = useState({ color: 0, storage: 1, care: 0 });
  const colors = [
    { name: "Midnight", hex: "#1d1d1f" },
    { name: "Silver", hex: "#e0dcd1" },
    { name: "Ocean", hex: "#2a5a8a" },
    { name: "Crimson", hex: "#c8382f" },
  ];
  const storage = ["128GB", "256GB", "512GB"];

  if (mobile) {
    return (
      <div className="frame-scroll" style={{ paddingBottom: 80 }}>
        <div style={{ position: "sticky", top: 0, zIndex: 30, background: "rgba(251,251,253,0.9)", backdropFilter: "blur(20px)", padding: "10px 14px", display: "flex", alignItems: "center", borderBottom: "1px solid var(--hairline)" }}>
          <IconChevL size={20} /><span style={{ flex: 1 }} /><IconHeart size={18} /><div style={{ width: 12 }} /><IconBag size={18} />
        </div>
        <div style={{ background: p.hero, padding: "20px 0" }}>
          <ProductImage product={p} style={{ width: "80%", margin: "0 auto", borderRadius: 0 }} />
          <div style={{ display: "flex", justifyContent: "center", gap: 5, marginTop: 10 }}>
            {[0,1,2,3].map(i => <span key={i} style={{ width: 6, height: 6, borderRadius: 999, background: i === 0 ? "var(--ink)" : "var(--ink-4)" }} />)}
          </div>
        </div>
        <div style={{ padding: 18 }}>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>{p.tagline}</div>
          <h1 style={{ fontSize: 28, fontWeight: 600, letterSpacing: "-0.02em", margin: "4px 0 6px" }}>{p.name}</h1>
          <Rating value={p.rating} count={p.reviews} />
          <div style={{ fontSize: 24, fontWeight: 500, marginTop: 10 }}>{fmt(p.price)}</div>
          <div style={{ fontSize: 12, color: "var(--ink-2)", marginTop: 2 }}>or $29/mo for 12 mo ›</div>

          <div style={{ marginTop: 20 }}>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 6 }}>Color — <span style={{ color: "var(--ink)" }}>{colors[selected.color].name}</span></div>
            <div style={{ display: "flex", gap: 8 }}>
              {colors.map((c, i) => (
                <button key={c.name} onClick={() => setSelected(s => ({ ...s, color: i }))} style={{ width: 32, height: 32, borderRadius: 999, background: c.hex, border: selected.color === i ? "2px solid var(--accent)" : "1px solid var(--hairline)", outline: "2px solid var(--surface)", outlineOffset: -4, cursor: "pointer" }} />
              ))}
            </div>
          </div>
          <div style={{ marginTop: 14 }}>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 6 }}>Storage</div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 6 }}>
              {storage.map((s, i) => (
                <button key={s} onClick={() => setSelected(sel => ({ ...sel, storage: i }))} style={{ padding: 10, border: selected.storage === i ? "2px solid var(--accent)" : "1px solid var(--hairline)", borderRadius: 10, background: "var(--surface)", fontSize: 13, cursor: "pointer" }}>{s}</button>
              ))}
            </div>
          </div>
          <div style={{ marginTop: 16, padding: 12, background: "var(--surface-3)", borderRadius: 12, display: "flex", alignItems: "center", gap: 10 }}>
            <IconTruck size={18} />
            <div>
              <div style={{ fontSize: 13, fontWeight: 500 }}>Free delivery</div>
              <div style={{ fontSize: 11, color: "var(--ink-2)" }}>Arrives Thu, Apr 23</div>
            </div>
          </div>
        </div>
        <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, padding: 14, background: "rgba(255,255,255,0.95)", backdropFilter: "blur(20px)", borderTop: "1px solid var(--hairline)", display: "flex", gap: 8 }}>
          <Btn kind="secondary" full>Add to Bag</Btn>
          <Btn kind="primary" full>Buy now</Btn>
        </div>
      </div>
    );
  }

  return (
    <div className="frame-scroll">
      <Nav active="audio" />
      {/* sticky secondary nav */}
      <div style={{ position: "sticky", top: 0, zIndex: 20, background: "rgba(251,251,253,0.85)", backdropFilter: "blur(20px)", borderBottom: "1px solid var(--hairline)" }}>
        <div style={{ maxWidth: 1280, margin: "0 auto", padding: "14px 40px", display: "flex", alignItems: "center", gap: 24, fontSize: 12 }}>
          <div style={{ fontSize: 17, fontWeight: 500, flex: 1 }}>{p.name}</div>
          <a style={{ color: "var(--ink-2)" }}>Overview</a>
          <a style={{ color: "var(--ink)" }}>Specs</a>
          <a style={{ color: "var(--ink-2)" }}>Compare</a>
          <a style={{ color: "var(--ink-2)" }}>Reviews</a>
          <div style={{ fontSize: 13, color: "var(--ink-2)" }}>From {fmt(p.price)}</div>
          <Btn kind="primary" size="sm">Buy</Btn>
        </div>
      </div>

      <div style={{ fontSize: 12, color: "var(--ink-2)", maxWidth: 1280, margin: "0 auto", padding: "14px 40px 0" }}>
        Audio &nbsp;›&nbsp; Headphones &nbsp;›&nbsp; <span style={{ color: "var(--ink)" }}>{p.name}</span>
      </div>

      <section style={{ maxWidth: 1280, margin: "0 auto", padding: "24px 40px 40px", display: "grid", gridTemplateColumns: "90px 1fr 380px", gap: 28, alignItems: "start" }}>
        {/* Thumbnails */}
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {[0,1,2,3,4].map(i => (
            <div key={i} style={{ width: 78, height: 78, background: p.hero, borderRadius: 12, border: i === 0 ? "2px solid var(--ink)" : "1px solid var(--hairline)", padding: 8, cursor: "pointer" }}>
              <ProductImage product={p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
            </div>
          ))}
        </div>
        {/* Main gallery */}
        <div style={{ background: p.hero, borderRadius: 22, padding: 40, minHeight: 460, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <ProductImage product={p} style={{ width: "72%", background: "transparent" }} />
        </div>
        {/* Buy box */}
        <aside style={{ padding: "8px 0", position: "sticky", top: 72 }}>
          <div style={{ fontSize: 13, color: "var(--ink-2)" }}>{p.tagline}</div>
          <h1 style={{ fontSize: 36, fontWeight: 600, letterSpacing: "-0.02em", margin: "4px 0 8px", lineHeight: 1.05 }}>{p.name} Wireless Headphones</h1>
          <Rating value={p.rating} count={p.reviews} />
          <div style={{ fontSize: 28, fontWeight: 500, marginTop: 16 }}>{fmt(p.price)}</div>
          <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 2 }}>or <b style={{ color: "var(--ink)", fontWeight: 500 }}>$29.08</b>/mo. for 12 mo.*</div>

          <div style={{ marginTop: 20 }}>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 8 }}>Color — <span style={{ color: "var(--ink)" }}>{colors[selected.color].name}</span></div>
            <div style={{ display: "flex", gap: 10 }}>
              {colors.map((c, i) => (
                <button key={c.name} onClick={() => setSelected(s => ({ ...s, color: i }))} title={c.name} style={{ width: 34, height: 34, borderRadius: 999, background: c.hex, border: "2px solid var(--surface)", outline: selected.color === i ? "2px solid var(--ink)" : "1px solid var(--hairline)", outlineOffset: 1, cursor: "pointer" }} />
              ))}
            </div>
          </div>

          <div style={{ marginTop: 18 }}>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 8 }}>Storage</div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 8 }}>
              {storage.map((s, i) => (
                <button key={s} onClick={() => setSelected(sel => ({ ...sel, storage: i }))} style={{ padding: "12px 8px", border: selected.storage === i ? "2px solid var(--ink)" : "1px solid var(--hairline)", borderRadius: 12, background: "var(--surface)", fontSize: 14, cursor: "pointer", textAlign: "center" }}>
                  <div style={{ fontWeight: 500 }}>{s}</div>
                  <div style={{ fontSize: 11, color: "var(--ink-2)", marginTop: 2 }}>+${i * 100}</div>
                </button>
              ))}
            </div>
          </div>

          <div style={{ marginTop: 18 }}>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 8 }}>Protection</div>
            {[{l:"None",s:""},{l:"Care+ 2 years",s:"+$49"}].map((o, i) => (
              <button key={o.l} onClick={() => setSelected(sel => ({ ...sel, care: i }))} style={{ display: "flex", alignItems: "center", width: "100%", padding: "10px 12px", border: selected.care === i ? "2px solid var(--ink)" : "1px solid var(--hairline)", borderRadius: 12, background: "var(--surface)", fontSize: 13, marginBottom: 6, cursor: "pointer", gap: 10 }}>
                <span style={{ width: 14, height: 14, borderRadius: 999, border: "1px solid var(--ink-3)", background: selected.care === i ? "var(--ink)" : "transparent", flexShrink: 0 }} />
                <span style={{ flex: 1, textAlign: "left" }}>{o.l}</span>
                {o.s && <span style={{ color: "var(--ink-2)" }}>{o.s}</span>}
              </button>
            ))}
          </div>

          <div style={{ marginTop: 20, display: "flex", flexDirection: "column", gap: 8 }}>
            <Btn kind="primary" size="lg" full>Add to Bag</Btn>
            <Btn kind="outline" size="lg" full>Buy now</Btn>
          </div>
          <div style={{ marginTop: 16, display: "flex", flexDirection: "column", gap: 10, fontSize: 13, color: "var(--ink-2)" }}>
            <div style={{ display: "flex", gap: 10, alignItems: "center" }}><IconTruck size={18} /><span><b style={{ color: "var(--ink)", fontWeight: 500 }}>Free delivery</b> — arrives Thu, Apr 23</span></div>
            <div style={{ display: "flex", gap: 10, alignItems: "center" }}><IconReturn size={18} /><span><b style={{ color: "var(--ink)", fontWeight: 500 }}>Returns within 30 days</b> — free and easy</span></div>
            <div style={{ display: "flex", gap: 10, alignItems: "center" }}><IconShield size={18} /><span><b style={{ color: "var(--ink)", fontWeight: 500 }}>2-year warranty</b> — included</span></div>
          </div>
        </aside>
      </section>

      {/* Highlights */}
      <section style={{ background: "var(--surface-3)", padding: "60px 40px" }}>
        <div style={{ maxWidth: 1280, margin: "0 auto" }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: "var(--accent)", textTransform: "uppercase", letterSpacing: "0.02em" }}>Why you'll love it</div>
          <h2 style={{ fontSize: 44, fontWeight: 600, letterSpacing: "-0.025em", margin: "8px 0 28px", maxWidth: 780, lineHeight: 1.05 }}>Every note. Every detail. Every time.</h2>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16 }}>
            {[
              { t: "Adaptive EQ", d: "Tunes sound to the shape of your ear in real time." },
              { t: "40-hour battery", d: "Up to 40 hours of playback. 5 min charge = 3 hours." },
              { t: "Spatial Audio", d: "Dynamic head-tracking puts you inside the sound." },
            ].map(f => (
              <div key={f.t} style={{ background: "var(--surface)", borderRadius: 22, padding: 28 }}>
                <div style={{ fontSize: 22, fontWeight: 600, letterSpacing: "-0.015em" }}>{f.t}</div>
                <div style={{ fontSize: 15, color: "var(--ink-2)", marginTop: 6, lineHeight: 1.47 }}>{f.d}</div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}

window.PDPV1 = PDPV1;
