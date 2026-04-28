/* PLP + Search/filters wireframes */

/* V1 — Sidebar filters (classic) */
function PLPV1Desktop() {
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>◯ brand</span>
        <Box style={{ width: 300, height: 28, borderRadius: 999, display: "flex", alignItems: "center", padding: "0 10px", color: "var(--ink-3)", fontSize: 13 }}>⌕ "headphones"</Box>
        <span style={{ flex: 1, textAlign: "right" }}>♡ 🛒 2</span>
      </Row>
      <div style={{ fontSize: 13, color: "var(--ink-2)", margin: "10px 0" }}>Home › Audio › Headphones</div>
      <Row style={{ alignItems: "baseline" }}>
        <div className="hand-title" style={{ fontSize: 26, flex: 1 }}>Headphones <span style={{ color: "var(--ink-3)", fontSize: 16 }}>(148)</span></div>
        <Row gap={8}>
          <Box style={{ padding: "4px 10px", fontSize: 13, borderRadius: 999 }}>Sort: Featured ▾</Box>
          <Box style={{ padding: "4px 10px", fontSize: 13, borderRadius: 999 }}>▦ ▤</Box>
        </Row>
      </Row>
      <Row gap={16} style={{ marginTop: 14 }}>
        {/* sidebar */}
        <Col gap={14} style={{ width: 180, flexShrink: 0 }}>
          {["Category","Brand","Price","Features","Rating","Color"].map((g, i) => (
            <div key={g}>
              <div className="hand-title" style={{ fontSize: 14, marginBottom: 6 }}>{g} {i===2?"—":"+"}</div>
              {i === 2 ? (
                <Col gap={4}>
                  <Box style={{ height: 6, borderRadius: 999, background: "rgba(0,0,0,0.05)" }} />
                  <Row style={{ fontSize: 11, color: "var(--ink-2)", justifyContent: "space-between" }}>
                    <span>$0</span><span>$999+</span>
                  </Row>
                </Col>
              ) : (
                <Col gap={2}>
                  {[0,1,2].map(j => <Row key={j} gap={6} style={{ fontSize: 12, color: "var(--ink-2)" }}><span>☐</span><span>Option {j+1}</span><span style={{ marginLeft: "auto" }}>({10+j*3})</span></Row>)}
                </Col>
              )}
            </div>
          ))}
        </Col>
        {/* grid */}
        <div style={{ flex: 1, display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
          {Array.from({length:6}).map((_,i) => (
            <Col key={i} gap={4}>
              <div style={{ position: "relative" }}>
                <ImgPh w="100%" h={130} />
                <span className="sb-tag" style={{ position: "absolute", top: 6, left: 6, background: "var(--paper)", fontSize: 11 }}>New</span>
                <span style={{ position: "absolute", top: 6, right: 8, fontSize: 14 }}>♡</span>
              </div>
              <div style={{ fontSize: 13, color: "var(--ink-2)" }}>Brand</div>
              <div style={{ fontSize: 14 }}>Product Name Here</div>
              <Row style={{ fontSize: 12, color: "var(--ink-2)", justifyContent: "space-between" }}>
                <span>★ 4.{5+i%3} ({120+i*30})</span>
                <span style={{ color: "var(--ink)" }}>$299</span>
              </Row>
            </Col>
          ))}
        </div>
      </Row>
    </div>
  );
}
function PLPV1Mobile() {
  return (
    <div style={{ padding: "0 10px", height: "100%", position: "relative" }}>
      <Row style={{ alignItems: "center", padding: "8px 0", gap: 8 }}>
        <span>←</span>
        <Box style={{ flex: 1, height: 26, borderRadius: 999, display: "flex", alignItems: "center", padding: "0 10px", color: "var(--ink-3)", fontSize: 12 }}>⌕ headphones</Box>
        <span>🛒</span>
      </Row>
      <div className="hand-title" style={{ fontSize: 18, margin: "4px 0" }}>Headphones (148)</div>
      <Row gap={6} style={{ overflow: "hidden" }}>
        <Box style={{ padding: "3px 8px", fontSize: 11, borderRadius: 999, whiteSpace: "nowrap" }}>Filter ⚙</Box>
        <Box style={{ padding: "3px 8px", fontSize: 11, borderRadius: 999, whiteSpace: "nowrap" }}>Sort ▾</Box>
        <Box style={{ padding: "3px 8px", fontSize: 11, borderRadius: 999, whiteSpace: "nowrap", color: "var(--accent)", borderColor: "var(--accent)" }}>$100-300 ✕</Box>
        <Box style={{ padding: "3px 8px", fontSize: 11, borderRadius: 999, whiteSpace: "nowrap", color: "var(--accent)", borderColor: "var(--accent)" }}>Wireless ✕</Box>
      </Row>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8, marginTop: 10 }}>
        {[0,1,2,3].map(i => (
          <Col key={i} gap={2}>
            <ImgPh w="100%" h={90} />
            <div style={{ fontSize: 11, color: "var(--ink-2)" }}>Brand</div>
            <div style={{ fontSize: 12 }}>Product</div>
            <div style={{ fontSize: 11 }}>$299</div>
          </Col>
        ))}
      </div>
    </div>
  );
}

/* V2 — Top filter bar, big grid */
function PLPV2Desktop() {
  return (
    <div style={{ padding: "0 28px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>◯</span>
        <Row gap={16} style={{ fontSize: 14, color: "var(--ink-2)" }}><span>Shop</span><span>Deals</span><span>Support</span></Row>
        <span style={{ flex: 1, textAlign: "right" }}>⌕ 🛒</span>
      </Row>
      <Row style={{ marginTop: 14, alignItems: "baseline" }}>
        <div className="hand-title" style={{ fontSize: 32, flex: 1 }}>Headphones.</div>
        <div style={{ color: "var(--ink-2)", fontSize: 14 }}>148 products</div>
      </Row>
      {/* filter bar */}
      <Row gap={8} style={{ marginTop: 12, padding: "10px 0", borderTop: "1.5px solid var(--ink)", borderBottom: "1.5px solid var(--ink)" }}>
        {["Brand ▾","Price ▾","Type ▾","Features ▾","Color ▾","Rating ▾"].map(f => (
          <Box key={f} style={{ padding: "4px 12px", fontSize: 13, borderRadius: 4 }}>{f}</Box>
        ))}
        <div style={{ flex: 1 }} />
        <Box style={{ padding: "4px 12px", fontSize: 13, borderRadius: 4 }}>Sort: Popular ▾</Box>
      </Row>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 14, marginTop: 16 }}>
        {Array.from({length:8}).map((_,i) => (
          <Col key={i} gap={4}>
            <div style={{ position: "relative" }}>
              <ImgPh w="100%" h={120} />
              <span style={{ position: "absolute", top: 6, right: 8, fontSize: 14 }}>♡</span>
            </div>
            <div style={{ fontSize: 13 }}>Product Name</div>
            <Row style={{ fontSize: 12, color: "var(--ink-2)", justifyContent: "space-between" }}>
              <span>★ 4.{5+i%3}</span><span>${199+i*20}</span>
            </Row>
          </Col>
        ))}
      </div>
    </div>
  );
}
function PLPV2Mobile() {
  return (
    <div style={{ padding: "0 10px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "8px 0", gap: 6 }}>
        <span>←</span>
        <span className="hand-title" style={{ fontSize: 15, flex: 1 }}>Headphones</span>
        <span>⌕ 🛒</span>
      </Row>
      <Row gap={4} style={{ marginTop: 6, padding: "6px 0", borderTop: "1.2px solid var(--ink-2)", borderBottom: "1.2px solid var(--ink-2)", overflow: "hidden" }}>
        {["Filter","Sort","Brand","Price","Color"].map(f => <Box key={f} style={{ padding: "2px 6px", fontSize: 10, borderRadius: 4, whiteSpace: "nowrap" }}>{f} ▾</Box>)}
      </Row>
      <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 10, marginTop: 10 }}>
        {[0,1,2].map(i => (
          <Row key={i} gap={8}>
            <ImgPh w={90} h={80} />
            <Col gap={2} style={{ flex: 1 }}>
              <div style={{ fontSize: 11, color: "var(--ink-2)" }}>Brand</div>
              <div style={{ fontSize: 13 }}>Product Name</div>
              <div style={{ fontSize: 11, color: "var(--ink-2)" }}>★ 4.6 (231)</div>
              <div style={{ fontSize: 13, marginTop: "auto" }}>$299</div>
            </Col>
          </Row>
        ))}
      </div>
    </div>
  );
}

/* V3 — Visual comparison table (novel) */
function PLPV3Desktop() {
  const specs = ["Price","Battery","Noise cancel","Weight","Warranty"];
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0" }}>
        <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>◯ brand</span>
        <span>⌕ 🛒</span>
      </Row>
      <Row style={{ alignItems: "baseline", marginTop: 6 }}>
        <div className="hand-title" style={{ fontSize: 28, flex: 1 }}>Compare the lineup.</div>
        <Box style={{ padding: "4px 12px", fontSize: 13, borderRadius: 999 }}>Showing 4 of 148 ▾</Box>
      </Row>
      <Row gap={12} style={{ marginTop: 14 }}>
        {/* left label column */}
        <Col gap={12} style={{ width: 140, flexShrink: 0, paddingTop: 160 }}>
          {specs.map(s => (
            <div key={s} style={{ fontSize: 13, color: "var(--ink-2)", height: 32, borderTop: "1px solid var(--ink-3)", paddingTop: 6 }}>{s}</div>
          ))}
          <div style={{ borderTop: "1px solid var(--ink-3)", paddingTop: 10, fontSize: 13, color: "var(--ink-2)" }}>Actions</div>
        </Col>
        {Array.from({length:4}).map((_,i) => (
          <Col key={i} gap={8} style={{ flex: 1 }}>
            <ImgPh w="100%" h={140} label={`Product ${i+1}`} />
            <div className="hand-title" style={{ fontSize: 15 }}>XR-{i+1}00</div>
            <div style={{ fontSize: 12, color: "var(--ink-2)" }}>★ 4.{5+i%3}</div>
            {specs.map((s, j) => (
              <div key={s} style={{ fontSize: 12, height: 32, borderTop: "1px solid var(--ink-3)", paddingTop: 6 }}>
                {j===0?`$${199+i*100}`: j===1?`${20+i*4}h` : j===2?(i>0?"✓":"—") : j===3?`${220+i*10}g` : "2yr"}
              </div>
            ))}
            <div style={{ borderTop: "1px solid var(--ink-3)", paddingTop: 8 }}>
              <Btn fill size="sm" style={{ width: "100%" }}>Add to cart</Btn>
            </div>
          </Col>
        ))}
      </Row>
    </div>
  );
}
function PLPV3Mobile() {
  return (
    <div style={{ padding: "0 10px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "8px 0", gap: 6 }}>
        <span>←</span>
        <span className="hand-title" style={{ fontSize: 14, flex: 1 }}>Compare</span>
        <Box style={{ padding: "2px 8px", fontSize: 10, borderRadius: 999 }}>2 of 4 ▾</Box>
      </Row>
      <Row gap={8} style={{ marginTop: 6 }}>
        {[0,1].map(i => (
          <Col key={i} gap={6} style={{ flex: 1 }}>
            <ImgPh w="100%" h={80} />
            <div className="hand-title" style={{ fontSize: 13 }}>XR-{i+1}00</div>
          </Col>
        ))}
      </Row>
      <Col gap={0} style={{ marginTop: 10 }}>
        {["Price","Battery","ANC","Weight"].map((s, j) => (
          <Row key={s} style={{ fontSize: 11, borderTop: "1px solid var(--ink-3)", padding: "6px 0" }}>
            <span style={{ flex: 1, color: "var(--ink-2)" }}>{s}</span>
            <span style={{ flex: 1 }}>${199+j*50}</span>
            <span style={{ flex: 1 }}>${299+j*50}</span>
          </Row>
        ))}
      </Col>
      <Row gap={6} style={{ marginTop: 10 }}>
        <Btn size="sm" style={{ flex: 1 }}>Add</Btn>
        <Btn fill size="sm" style={{ flex: 1 }}>Add</Btn>
      </Row>
    </div>
  );
}

window.PLPPages = [
  { key: "p1", title: "V1 · Sidebar filters", caption: <>The workhorse: persistent left filter rail, 3-col grid. Familiar, scannable. <b>Best for deep catalogs</b>.</>, desktop: <PLPV1Desktop />, mobile: <PLPV1Mobile />, ann: "safe" },
  { key: "p2", title: "V2 · Top filter bar", caption: <>Horizontal filter dropdowns, 4-col grid. More visual breathing room. Mobile shows list view, not grid. <b>Best for lifestyle browsing</b>.</>, desktop: <PLPV2Desktop />, mobile: <PLPV2Mobile />, ann: "browse" },
  { key: "p3", title: "V3 · Comparison mode (novel)", caption: <>Ditch the grid — let buyers <b>compare 4 products spec-by-spec</b>. Useful for electronics where specs matter. Risk: feels dense.</>, desktop: <PLPV3Desktop />, mobile: <PLPV3Mobile />, ann: "novel" },
];
