/* PDP wireframes */

/* V1 — Classic gallery + buy box */
function PDPV1Desktop() {
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>◯</span>
        <span>⌕ 🛒 (2)</span>
      </Row>
      <div style={{ fontSize: 12, color: "var(--ink-2)", margin: "8px 0" }}>Audio › Headphones › XR-7</div>
      <Row gap={18} style={{ marginTop: 6 }}>
        {/* gallery */}
        <Col gap={8} style={{ width: 52 }}>
          {Array.from({length:5}).map((_,i) => <ImgPh key={i} w={52} h={52} />)}
        </Col>
        <div style={{ flex: 1 }}>
          <ImgPh w="100%" h={340} label="hero product shot" />
        </div>
        {/* buy box */}
        <Col gap={10} style={{ width: 240, flexShrink: 0 }}>
          <div style={{ fontSize: 13, color: "var(--ink-2)" }}>Brand</div>
          <div className="hand-title" style={{ fontSize: 22, lineHeight: 1.05 }}>XR-7 Wireless Headphones</div>
          <Row gap={6} style={{ fontSize: 12, color: "var(--ink-2)" }}>
            <span>★★★★★</span><span>4.8 (1,204)</span>
          </Row>
          <div className="hand-title" style={{ fontSize: 22 }}>$349</div>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Color</div>
          <Row gap={6}>
            {["#1a1a1a","#e0dcd1","#2a5a8a","#c8382f"].map(c => <div key={c} style={{ width: 24, height: 24, borderRadius: 999, background: c, border: c==="#1a1a1a" ? "2px solid var(--accent)" : "1.2px solid var(--ink-2)" }} />)}
          </Row>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Memory</div>
          <Row gap={6}>
            {["128GB","256GB","512GB"].map((s,i) => <Box key={s} style={{ padding: "3px 10px", fontSize: 12, borderRadius: 4, ...(i===1 ? { borderColor: "var(--accent)", color: "var(--accent)" } : {}) }}>{s}</Box>)}
          </Row>
          <Btn fill style={{ width: "100%", marginTop: 6 }}>Add to Bag</Btn>
          <Btn accent style={{ width: "100%" }}>Buy now</Btn>
          <div style={{ fontSize: 12, color: "var(--ink-2)", textAlign: "center" }}>Free delivery · Returns within 30 days</div>
        </Col>
      </Row>
      <div className="hand-title" style={{ fontSize: 18, marginTop: 20 }}>Highlights</div>
      <Row gap={12} style={{ marginTop: 8 }}>
        {[0,1,2,3].map(i => <Col key={i} gap={4} style={{ flex: 1 }}><ImgPh w="100%" h={60} /><Scribble lines={2} widths={["s-1","s-3"]} /></Col>)}
      </Row>
    </div>
  );
}
function PDPV1Mobile() {
  return (
    <div style={{ height: "100%", position: "relative", overflow: "hidden" }}>
      <Row style={{ alignItems: "center", padding: "8px 12px", gap: 6 }}>
        <span>←</span><span style={{ flex: 1 }} /><span>♡</span><span>🛒</span>
      </Row>
      <ImgPh w="100%" h={220} label="hero" />
      <Row gap={3} style={{ justifyContent: "center", padding: "6px 0" }}>
        <span style={{ width: 6, height: 6, borderRadius: 999, background: "var(--ink)" }} />
        {[0,1,2].map(i => <span key={i} style={{ width: 6, height: 6, borderRadius: 999, background: "var(--ink-3)" }} />)}
      </Row>
      <div style={{ padding: "0 12px" }}>
        <div style={{ fontSize: 11, color: "var(--ink-2)" }}>Brand</div>
        <div className="hand-title" style={{ fontSize: 18, lineHeight: 1.05 }}>XR-7 Wireless</div>
        <Row gap={4} style={{ fontSize: 11, color: "var(--ink-2)", marginTop: 3 }}>
          <span>★ 4.8</span><span>(1,204)</span>
        </Row>
        <div className="hand-title" style={{ fontSize: 18, marginTop: 4 }}>$349</div>
        <div style={{ fontSize: 11, color: "var(--ink-2)", marginTop: 8 }}>Color · Midnight</div>
        <Row gap={5} style={{ marginTop: 4 }}>
          {["#1a1a1a","#e0dcd1","#2a5a8a","#c8382f"].map((c,i) => <div key={c} style={{ width: 20, height: 20, borderRadius: 999, background: c, border: i===0 ? "2px solid var(--accent)" : "1px solid var(--ink-2)" }} />)}
        </Row>
      </div>
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, padding: 10, background: "var(--paper)", borderTop: "1.5px solid var(--ink)", display: "flex", gap: 8 }}>
        <Btn size="sm" style={{ flex: 1 }}>Add to Bag</Btn>
        <Btn fill size="sm" style={{ flex: 1 }}>Buy now</Btn>
      </div>
    </div>
  );
}

/* V2 — Editorial long-form scroll */
function PDPV2Desktop() {
  return (
    <div style={{ height: "100%", overflow: "hidden" }}>
      <Row style={{ alignItems: "center", padding: "10px 24px", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>XR-7</span>
        <Row gap={14} style={{ fontSize: 13, color: "var(--ink-2)" }}><span>Overview</span><span>Specs</span><span>Compare</span><span>Reviews</span></Row>
        <Btn accent size="sm" style={{ marginLeft: 14 }}>Buy — $349</Btn>
      </Row>
      <div style={{ padding: "40px 60px", textAlign: "center", background: "rgba(0,0,0,0.03)" }}>
        <div style={{ fontSize: 13, color: "var(--ink-2)", letterSpacing: "0.1em", textTransform: "uppercase" }}>Introducing</div>
        <div className="hand-title" style={{ fontSize: 56, lineHeight: 0.95, marginTop: 10 }}>XR-7.</div>
        <div className="hand-title" style={{ fontSize: 28, color: "var(--ink-2)", marginTop: 4 }}>Sound you can feel.</div>
        <div style={{ margin: "24px auto 0", width: 360, height: 180, position: "relative" }}>
          <ImgPh w="100%" h="100%" label="hero photography" />
        </div>
        <Row gap={8} style={{ justifyContent: "center", marginTop: 20 }}>
          <Btn fill size="lg">Buy — $349</Btn>
          <Btn size="lg">Watch film ›</Btn>
        </Row>
      </div>
      <Row gap={16} style={{ padding: "20px 60px" }}>
        <div style={{ flex: 1 }}>
          <div className="hand-title" style={{ fontSize: 20 }}>Every note.<br/>Every detail.</div>
        </div>
        <div style={{ flex: 1 }}>
          <Scribble lines={4} widths={["s-1","s-2","s-1","s-3"]} />
        </div>
      </Row>
    </div>
  );
}
function PDPV2Mobile() {
  return (
    <div style={{ height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "8px 10px" }}>
        <span>←</span><span style={{ flex: 1 }} /><Btn accent size="sm">Buy</Btn>
      </Row>
      <div style={{ padding: "16px 14px", textAlign: "center" }}>
        <div style={{ fontSize: 10, color: "var(--ink-2)", letterSpacing: "0.1em" }}>INTRODUCING</div>
        <div className="hand-title" style={{ fontSize: 34, lineHeight: 0.95, marginTop: 6 }}>XR-7.</div>
        <div className="hand-title" style={{ fontSize: 15, color: "var(--ink-2)", marginTop: 4 }}>Sound you can feel.</div>
        <ImgPh w="100%" h={140} style={{ marginTop: 14 }} />
      </div>
      <div style={{ padding: "8px 14px" }}>
        <div className="hand-title" style={{ fontSize: 15 }}>Every note.<br/>Every detail.</div>
        <Scribble lines={3} widths={["s-1","s-2","s-3"]} />
      </div>
    </div>
  );
}

/* V3 — Compact buy-first with sticky summary */
function PDPV3Desktop() {
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ alignItems: "center", padding: "10px 0" }}>
        <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>◯ / XR-7</span>
        <span>⌕ 🛒</span>
      </Row>
      {/* sticky summary bar */}
      <Box style={{ padding: "8px 14px", display: "flex", alignItems: "center", gap: 12, marginBottom: 12, borderRadius: 8 }} fill>
        <div style={{ width: 40, height: 40, border: "1px solid var(--ink-2)", borderRadius: 6 }} className="hatch" />
        <div style={{ flex: 1 }}>
          <div className="hand-title" style={{ fontSize: 14 }}>XR-7 · Midnight · 256GB</div>
          <div style={{ fontSize: 11, color: "var(--ink-2)" }}>In stock · Delivers Thu, Apr 23</div>
        </div>
        <div className="hand-title" style={{ fontSize: 16 }}>$349</div>
        <Btn fill size="sm">Add to Bag</Btn>
      </Box>
      <Row gap={16}>
        <div style={{ flex: 1.5 }}>
          <ImgPh w="100%" h={260} label="swap / 360° spin" />
          <Row gap={4} style={{ marginTop: 6 }}>{[0,1,2,3,4,5].map(i => <div key={i} style={{ flex: 1, height: 34, border: i===1 ? "2px solid var(--accent)" : "1px solid var(--ink-2)", borderRadius: 3, background: "rgba(0,0,0,0.03)" }} />)}</Row>
        </div>
        <Col gap={10} style={{ flex: 1 }}>
          <div className="hand-title" style={{ fontSize: 16 }}>Configure yours.</div>
          {/* options */}
          {[{label: "Color", opts: ["Midnight","Silver","Blue","Red"]}, {label: "Memory", opts: ["128GB","256GB","512GB"]}, {label: "AppleCare", opts: ["None","2 years +$49"]}].map(g => (
            <div key={g.label}>
              <div style={{ fontSize: 12, color: "var(--ink-2)", marginBottom: 4 }}>{g.label}</div>
              <Row gap={6}>
                {g.opts.map((o,i) => <Box key={o} style={{ padding: "4px 8px", fontSize: 11, borderRadius: 4, ...(i===0 ? { borderColor: "var(--accent)", color: "var(--accent)" } : {}) }}>{o}</Box>)}
              </Row>
            </div>
          ))}
          <div style={{ fontSize: 11, color: "var(--ink-2)", borderTop: "1px solid var(--ink-3)", paddingTop: 8 }}>Pay in 4 of $87.25 · No fees</div>
        </Col>
      </Row>
    </div>
  );
}
function PDPV3Mobile() {
  return (
    <div style={{ height: "100%", position: "relative" }}>
      <Row style={{ alignItems: "center", padding: "8px 10px", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span>←</span>
        <span className="hand-title" style={{ fontSize: 13, flex: 1, textAlign: "center" }}>XR-7 · $349</span>
        <span>🛒</span>
      </Row>
      <ImgPh w="100%" h={150} />
      <div style={{ padding: "10px 12px" }}>
        <div className="hand-title" style={{ fontSize: 15 }}>Configure yours.</div>
        {[{l:"Color",opts:["Midnight","Silver","Blue"]},{l:"Memory",opts:["128","256","512"]}].map(g => (
          <div key={g.l} style={{ marginTop: 8 }}>
            <div style={{ fontSize: 10, color: "var(--ink-2)" }}>{g.l}</div>
            <Row gap={4} style={{ marginTop: 3 }}>
              {g.opts.map((o,i) => <Box key={o} style={{ padding: "3px 6px", fontSize: 10, borderRadius: 4, ...(i===0 ? { borderColor: "var(--accent)", color: "var(--accent)" } : {}) }}>{o}</Box>)}
            </Row>
          </div>
        ))}
      </div>
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, padding: 10, background: "var(--paper)", borderTop: "1.5px solid var(--ink)" }}>
        <Btn fill style={{ width: "100%" }}>Add to Bag — $349</Btn>
      </div>
    </div>
  );
}

window.PDPPages = [
  { key: "d1", title: "V1 · Classic gallery + buy box", caption: <>Thumbnails left, hero center, <b>sticky buy box right</b>. The proven pattern from Amazon-era ecommerce. Zero surprises.</>, desktop: <PDPV1Desktop />, mobile: <PDPV1Mobile />, ann: "safe" },
  { key: "d2", title: "V2 · Editorial long-form", caption: <>Full-bleed marketing story — copy stanzas, cinematic product shots. <b>Best for flagship SKUs</b> where the story sells the product.</>, desktop: <PDPV2Desktop />, mobile: <PDPV2Mobile />, ann: "story" },
  { key: "d3", title: "V3 · Configurator-first", caption: <>Sticky summary at the top, options-forward layout. <b>Reduces anxiety</b> for configurable products (color/storage/warranty).</>, desktop: <PDPV3Desktop />, mobile: <PDPV3Mobile />, ann: "convert" },
];
