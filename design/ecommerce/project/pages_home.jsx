/* Home / Landing wireframes */
const { Fragment } = React;

/* V1 — Classic hero + grid */
function HomeV1Desktop() {
  return (
    <div style={{ padding: "0 28px", height: "100%", overflow: "hidden" }}>
      {/* nav */}
      <Row style={{ alignItems: "center", padding: "14px 0", borderBottom: "1.5px solid var(--ink)" }}>
        <div className="hand-title" style={{ fontSize: 22, flex: 1 }}>◯ brand</div>
        <Row gap={16} style={{ color: "var(--ink-2)", fontSize: 16 }}>
          <span>Shop</span><span>Categories</span><span>Deals</span><span>Support</span>
        </Row>
        <Row gap={10} style={{ flex: 1, justifyContent: "flex-end" }}>
          <Box style={{ width: 180, height: 26, display: "flex", alignItems: "center", padding: "0 10px", color: "var(--ink-3)", fontSize: 14, borderRadius: 999 }}>⌕ search</Box>
          <span className="sb-circle">♡</span>
          <span className="sb-circle">⌂</span>
          <span className="sb-circle">🛒</span>
        </Row>
      </Row>
      {/* hero */}
      <Row gap={16} style={{ marginTop: 20 }}>
        <Box style={{ flex: 2, height: 260, position: "relative" }} className="hatch">
          <div style={{ position: "absolute", left: 28, top: 28, maxWidth: 340 }}>
            <div className="hand-title" style={{ fontSize: 34, lineHeight: 1.05 }}>The New<br/>XR-7 Headphones.</div>
            <div style={{ color: "var(--ink-2)", marginTop: 8, fontSize: 16 }}>Sound, redefined.</div>
            <Row gap={8} style={{ marginTop: 14 }}>
              <Btn fill>Buy</Btn>
              <Btn>Learn more ›</Btn>
            </Row>
          </div>
          <div style={{ position: "absolute", right: 20, top: 20, bottom: 20, width: 240, background: "rgba(246,241,231,0.7)", border: "1.5px dashed var(--ink-2)", borderRadius: 4, display: "flex", alignItems: "center", justifyContent: "center", color: "var(--ink-2)" }}>product photo</div>
        </Box>
        <Col gap={14} style={{ flex: 1 }}>
          <Box style={{ height: 123, position: "relative", padding: 14 }} className="hatch">
            <div className="hand-title" style={{ fontSize: 18 }}>Laptops</div>
            <div className="annot" style={{ color: "var(--ink-2)", fontSize: 14 }}>Shop ›</div>
          </Box>
          <Box style={{ height: 123, position: "relative", padding: 14 }} className="hatch">
            <div className="hand-title" style={{ fontSize: 18 }}>Watches</div>
            <div className="annot" style={{ color: "var(--ink-2)", fontSize: 14 }}>Shop ›</div>
          </Box>
        </Col>
      </Row>
      {/* section head */}
      <div className="hand-title" style={{ fontSize: 22, marginTop: 26, marginBottom: 10 }}>New Arrivals</div>
      <Row gap={12}>
        {[0,1,2,3].map(i => (
          <Col key={i} gap={4} style={{ flex: 1 }}>
            <ImgPh w="100%" h={110} />
            <div style={{ fontSize: 15 }}>Product Name</div>
            <div style={{ fontSize: 14, color: "var(--ink-2)" }}>$299</div>
          </Col>
        ))}
      </Row>
      {/* promo strip */}
      <Box style={{ marginTop: 18, height: 56, display: "flex", alignItems: "center", justifyContent: "space-between", padding: "0 20px" }} fill>
        <div className="hand-title" style={{ fontSize: 18 }}>Free shipping over $50</div>
        <Btn accent>Shop all ›</Btn>
      </Box>
    </div>
  );
}

function HomeV1Mobile() {
  return (
    <div style={{ padding: "0 12px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0", justifyContent: "space-between" }}>
        <span className="hand-title" style={{ fontSize: 18 }}>≡</span>
        <span className="hand-title" style={{ fontSize: 16 }}>◯ brand</span>
        <span>⌕ 🛒</span>
      </Row>
      <Box style={{ height: 30, display: "flex", alignItems: "center", padding: "0 10px", color: "var(--ink-3)", fontSize: 13, borderRadius: 999 }}>⌕ Search products</Box>
      <Box style={{ height: 160, marginTop: 12, padding: 14, position: "relative" }} className="hatch">
        <div className="hand-title" style={{ fontSize: 20, lineHeight: 1.05 }}>The New<br/>XR-7.</div>
        <Btn fill style={{ marginTop: 10 }} size="sm">Buy</Btn>
      </Box>
      <Row gap={8} style={{ marginTop: 12, overflow: "hidden" }}>
        {["Phones","Laptops","Audio","Watch"].map(c => (
          <Box key={c} style={{ padding: "4px 10px", fontSize: 13, whiteSpace: "nowrap", borderRadius: 999 }}>{c}</Box>
        ))}
      </Row>
      <div className="hand-title" style={{ fontSize: 17, marginTop: 16, marginBottom: 8 }}>New Arrivals</div>
      <Row gap={8}>
        <Col gap={2} style={{ flex: 1 }}>
          <ImgPh w="100%" h={100} />
          <div style={{ fontSize: 13 }}>Product</div>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>$299</div>
        </Col>
        <Col gap={2} style={{ flex: 1 }}>
          <ImgPh w="100%" h={100} />
          <div style={{ fontSize: 13 }}>Product</div>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>$199</div>
        </Col>
      </Row>
      {/* mobile tab bar */}
      <div style={{ position: "absolute", bottom: 0, left: 6, right: 6, height: 44, borderTop: "1.5px solid var(--ink-2)", display: "flex", justifyContent: "space-around", alignItems: "center", background: "rgba(246,241,231,0.95)" }}>
        {["⌂","⌕","♡","🛒","👤"].map((i,k) => <span key={k} style={{ fontSize: 16 }}>{i}</span>)}
      </div>
    </div>
  );
}

/* V2 — Full-bleed editorial hero, big type */
function HomeV2Desktop() {
  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Row style={{ alignItems: "center", padding: "10px 24px", borderBottom: "1.5px solid var(--ink)" }}>
        <Row gap={18} style={{ flex: 1, fontSize: 15, color: "var(--ink-2)" }}>
          <span>◯</span><span>Phones</span><span>Laptops</span><span>Audio</span><span>Watches</span><span>Home</span><span>Accessories</span>
        </Row>
        <Row gap={12}>
          <span>⌕</span><span>♡</span><span>🛒 (2)</span>
        </Row>
      </Row>
      {/* full bleed hero */}
      <div style={{ flex: 1, position: "relative", minHeight: 0 }} className="hatch-dense">
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <div style={{ textAlign: "center" }}>
            <div className="hand-title" style={{ fontSize: 72, lineHeight: 0.95, letterSpacing: "-0.02em" }}>Quiet.<br/>Precise.<br/>Yours.</div>
            <div style={{ fontSize: 17, color: "var(--ink-2)", marginTop: 14 }}>The XR-7, now in five new finishes.</div>
            <Row gap={10} style={{ marginTop: 18, justifyContent: "center" }}>
              <Btn fill size="lg">Buy — $349</Btn>
              <Btn size="lg">Learn more ›</Btn>
            </Row>
          </div>
        </div>
        {/* floating product */}
        <div style={{ position: "absolute", right: 40, bottom: 40, width: 140, height: 140, border: "1.5px dashed var(--ink-2)", borderRadius: 999, background: "rgba(246,241,231,0.8)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 13, color: "var(--ink-2)" }}>product</div>
      </div>
      {/* scroll strip teaser */}
      <Row gap={0} style={{ borderTop: "1.5px solid var(--ink)", height: 50, alignItems: "center" }}>
        <div style={{ flex: 1, padding: "0 20px", fontSize: 15, color: "var(--ink-2)" }}>↓ Scroll to explore · Chapter 01 of 04</div>
        <div style={{ padding: "0 20px", fontSize: 15 }}>●○○○</div>
      </Row>
    </div>
  );
}
function HomeV2Mobile() {
  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Row style={{ alignItems: "center", padding: "10px 12px", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span style={{ flex: 1 }}>≡</span>
        <span className="hand-title" style={{ fontSize: 16 }}>◯</span>
        <span style={{ flex: 1, textAlign: "right" }}>⌕ 🛒</span>
      </Row>
      <div style={{ flex: 1, position: "relative" }} className="hatch-dense">
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <div style={{ textAlign: "center", padding: 16 }}>
            <div className="hand-title" style={{ fontSize: 38, lineHeight: 0.95 }}>Quiet.<br/>Precise.<br/>Yours.</div>
            <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 10 }}>XR-7, five finishes.</div>
            <Btn fill style={{ marginTop: 14 }}>Buy — $349</Btn>
          </div>
        </div>
      </div>
      <div style={{ padding: "10px 12px", fontSize: 12, color: "var(--ink-2)", borderTop: "1.2px solid var(--ink-2)" }}>↓ 01 of 04 ●○○○</div>
    </div>
  );
}

/* V3 — Shop-by-category grid, no hero */
function HomeV3Desktop() {
  const cats = ["Phones","Laptops","Audio","Watches","Cameras","Gaming","Smart Home","Accessories"];
  return (
    <div style={{ padding: "0 28px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "14px 0" }}>
        <div className="hand-title" style={{ fontSize: 22, flex: 1 }}>◯ brand</div>
        <Box style={{ width: 360, height: 32, display: "flex", alignItems: "center", padding: "0 12px", color: "var(--ink-2)", fontSize: 14, borderRadius: 999 }}>⌕ Search 2,431 products</Box>
        <Row gap={12} style={{ flex: 1, justifyContent: "flex-end" }}>
          <span>♡ 3</span><span>👤</span><span>🛒 2</span>
        </Row>
      </Row>
      <Row style={{ alignItems: "baseline", marginTop: 6 }}>
        <div className="hand-title" style={{ fontSize: 30, flex: 1 }}>Shop by category.</div>
        <div style={{ color: "var(--ink-2)", fontSize: 15 }}>View all ›</div>
      </Row>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 12, marginTop: 14 }}>
        {cats.map(c => (
          <Box key={c} style={{ height: 110, padding: 12, display: "flex", flexDirection: "column", justifyContent: "space-between", position: "relative" }} className="hatch">
            <div className="hand-title" style={{ fontSize: 17 }}>{c}</div>
            <div style={{ fontSize: 13, color: "var(--ink-2)" }}>Shop ›</div>
          </Box>
        ))}
      </div>
      <Row style={{ alignItems: "baseline", marginTop: 18 }}>
        <div className="hand-title" style={{ fontSize: 22, flex: 1 }}>Trending</div>
        <div style={{ color: "var(--ink-2)", fontSize: 14 }}>See all ›</div>
      </Row>
      <Row gap={10} style={{ marginTop: 8 }}>
        {[0,1,2,3,4].map(i => (
          <Col key={i} gap={3} style={{ flex: 1 }}>
            <ImgPh w="100%" h={90} />
            <div style={{ fontSize: 13 }}>Product {i+1}</div>
            <div style={{ fontSize: 12, color: "var(--ink-2)" }}>★ 4.{5+i%3} · $199</div>
          </Col>
        ))}
      </Row>
    </div>
  );
}
function HomeV3Mobile() {
  const cats = ["Phones","Laptops","Audio","Watches","Cameras","Gaming"];
  return (
    <div style={{ padding: "0 12px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "8px 0", justifyContent: "space-between" }}>
        <span className="hand-title" style={{ fontSize: 16 }}>◯</span>
        <Box style={{ flex: 1, margin: "0 8px", height: 26, display: "flex", alignItems: "center", padding: "0 10px", color: "var(--ink-3)", fontSize: 12, borderRadius: 999 }}>⌕ Search</Box>
        <span>🛒</span>
      </Row>
      <div className="hand-title" style={{ fontSize: 18, marginTop: 6 }}>Shop by category.</div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8, marginTop: 8 }}>
        {cats.map(c => (
          <Box key={c} style={{ height: 66, padding: 10, position: "relative" }} className="hatch">
            <div className="hand-title" style={{ fontSize: 14 }}>{c}</div>
            <div style={{ position: "absolute", right: 8, bottom: 6, fontSize: 11, color: "var(--ink-2)" }}>›</div>
          </Box>
        ))}
      </div>
      <div className="hand-title" style={{ fontSize: 16, marginTop: 12 }}>Trending</div>
      <Row gap={8} style={{ marginTop: 6 }}>
        {[0,1].map(i => (
          <Col key={i} gap={2} style={{ flex: 1 }}>
            <ImgPh w="100%" h={90} />
            <div style={{ fontSize: 12 }}>Product</div>
            <div style={{ fontSize: 11, color: "var(--ink-2)" }}>$199</div>
          </Col>
        ))}
      </Row>
    </div>
  );
}

window.HomePages = [
  { key: "h1", title: "V1 · Classic hero + grid", caption: <>Two-up marketing hero, category shortcut cards, New Arrivals row. The <b>safe bet</b> — everything ecommerce users expect. Browser nav is light.</>, desktop: <HomeV1Desktop />, mobile: <HomeV1Mobile />, ann: "safe" },
  { key: "h2", title: "V2 · Editorial single-hero", caption: <>Full-bleed hero, narrative scroll chapters. Leans into <b>product storytelling</b> — works when you have one flagship. Nav is minimized.</>, desktop: <HomeV2Desktop />, mobile: <HomeV2Mobile />, ann: "bold" },
  { key: "h3", title: "V3 · Shop-by-category grid", caption: <>No hero. Search-first, category-grid landing — good for stores with <b>broad catalogs</b> and returning customers.</>, desktop: <HomeV3Desktop />, mobile: <HomeV3Mobile />, ann: "utility" },
];
