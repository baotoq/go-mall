/* Cart + Checkout + Account */

/* ===== CART ===== */
function CartV1Desktop() {
  return (
    <div style={{ padding: "0 32px", height: "100%" }}>
      <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>◯</span>
        <span className="hand-title" style={{ fontSize: 14 }}>Your Bag</span>
      </Row>
      <div className="hand-title" style={{ fontSize: 28, marginTop: 14 }}>Your Bag.</div>
      <div style={{ color: "var(--ink-2)", fontSize: 14, marginBottom: 14 }}>2 items · Free shipping</div>
      <Row gap={24}>
        <Col gap={12} style={{ flex: 2 }}>
          {[0,1].map(i => (
            <Row key={i} gap={12} style={{ paddingBottom: 12, borderBottom: "1px solid var(--ink-3)" }}>
              <ImgPh w={100} h={100} />
              <Col gap={2} style={{ flex: 1 }}>
                <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Brand</div>
                <div className="hand-title" style={{ fontSize: 18 }}>XR-7 Wireless · Midnight</div>
                <div style={{ fontSize: 12, color: "var(--ink-2)" }}>In stock · Ships by Thu</div>
                <Row style={{ marginTop: "auto", alignItems: "center" }}>
                  <div className="stepper"><span>−</span><span>1</span><span>+</span></div>
                  <Row gap={12} style={{ marginLeft: 16, fontSize: 12, color: "var(--ink-2)" }}>
                    <span className="squiggle">Save for later</span>
                    <span className="squiggle">Remove</span>
                  </Row>
                </Row>
              </Col>
              <div className="hand-title" style={{ fontSize: 18 }}>${349-i*100}</div>
            </Row>
          ))}
        </Col>
        <Box style={{ flex: 1, padding: 18, borderRadius: 8, height: "fit-content" }} fill>
          <div className="hand-title" style={{ fontSize: 18, marginBottom: 10 }}>Order Summary</div>
          {[["Subtotal","$548"],["Shipping","Free"],["Estimated tax","$45.21"]].map(([k,v]) => (
            <Row key={k} style={{ fontSize: 13, padding: "4px 0" }}>
              <span style={{ flex: 1, color: "var(--ink-2)" }}>{k}</span><span>{v}</span>
            </Row>
          ))}
          <div style={{ borderTop: "1.5px solid var(--ink)", margin: "10px 0" }} />
          <Row style={{ fontSize: 16 }}>
            <span className="hand-title" style={{ flex: 1 }}>Total</span>
            <span className="hand-title">$593.21</span>
          </Row>
          <Btn fill style={{ width: "100%", marginTop: 14 }}>Checkout</Btn>
          <div style={{ fontSize: 11, color: "var(--ink-2)", textAlign: "center", marginTop: 8 }}>or Apple Pay · Shop Pay · PayPal</div>
        </Box>
      </Row>
    </div>
  );
}
function CartV1Mobile() {
  return (
    <div style={{ height: "100%", position: "relative" }}>
      <Row style={{ padding: "8px 12px", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span style={{ flex: 1 }}>← Bag</span>
      </Row>
      <div style={{ padding: "0 12px" }}>
        {[0,1].map(i => (
          <Row key={i} gap={8} style={{ padding: "10px 0", borderBottom: "1px solid var(--ink-3)" }}>
            <ImgPh w={60} h={60} />
            <Col gap={1} style={{ flex: 1 }}>
              <div style={{ fontSize: 12 }}>XR-{7-i} · Midnight</div>
              <div style={{ fontSize: 10, color: "var(--ink-2)" }}>Ships Thu</div>
              <Row style={{ marginTop: 4, alignItems: "center" }}>
                <div className="stepper" style={{ fontSize: 12 }}><span>−</span><span>1</span><span>+</span></div>
                <span style={{ marginLeft: "auto", fontSize: 13 }} className="hand-title">${349-i*100}</span>
              </Row>
            </Col>
          </Row>
        ))}
      </div>
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, padding: 10, background: "var(--paper)", borderTop: "1.5px solid var(--ink)" }}>
        <Row style={{ fontSize: 13, marginBottom: 6 }}>
          <span style={{ flex: 1 }}>Total</span><span className="hand-title">$593.21</span>
        </Row>
        <Btn fill style={{ width: "100%" }}>Checkout</Btn>
      </div>
    </div>
  );
}

function CartV2Desktop() {
  return (
    <div style={{ height: "100%", display: "flex" }}>
      <div style={{ flex: 1, padding: "0 24px" }}>
        <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
          <span className="hand-title" style={{ fontSize: 16 }}>◯</span>
        </Row>
        <div className="hand-title" style={{ fontSize: 20, marginTop: 14 }}>Continue shopping</div>
        <Row gap={10} style={{ marginTop: 10 }}>
          {[0,1,2,3].map(i => <Col key={i} gap={3} style={{ flex: 1 }}><ImgPh w="100%" h={80} /><div style={{ fontSize: 11 }}>Recommended</div><div style={{ fontSize: 11, color: "var(--ink-2)" }}>$199</div></Col>)}
        </Row>
        <div className="hand-title" style={{ fontSize: 16, marginTop: 16 }}>Saved for later (3)</div>
        <Row gap={8} style={{ marginTop: 8 }}>
          {[0,1,2].map(i => <Box key={i} style={{ flex: 1, padding: 8 }}><ImgPh w="100%" h={50} /><div style={{ fontSize: 11, marginTop: 4 }}>Product</div></Box>)}
        </Row>
      </div>
      {/* slide-in drawer */}
      <div style={{ width: 360, borderLeft: "1.5px solid var(--ink)", padding: "0 20px", background: "var(--paper)" }}>
        <Row style={{ padding: "10px 0", alignItems: "center" }}>
          <span className="hand-title" style={{ fontSize: 18, flex: 1 }}>Your Bag (2)</span>
          <span>✕</span>
        </Row>
        <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Free shipping · Delivers Thu</div>
        {[0,1].map(i => (
          <Row key={i} gap={8} style={{ padding: "10px 0", borderTop: "1px solid var(--ink-3)" }}>
            <ImgPh w={60} h={60} />
            <Col gap={1} style={{ flex: 1 }}>
              <div style={{ fontSize: 12 }}>XR-{7-i} · Midnight</div>
              <div className="stepper" style={{ fontSize: 11, marginTop: 4 }}><span>−</span><span>1</span><span>+</span></div>
            </Col>
            <div style={{ fontSize: 13 }}>${349-i*100}</div>
          </Row>
        ))}
        <div style={{ borderTop: "1.5px solid var(--ink)", marginTop: 10, paddingTop: 10 }}>
          <Row style={{ fontSize: 14, marginBottom: 10 }}><span style={{ flex: 1 }}>Total</span><span className="hand-title">$593.21</span></Row>
          <Btn fill style={{ width: "100%" }}>Checkout</Btn>
          <div style={{ marginTop: 8, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6 }}>
            <Box style={{ padding: "6px", fontSize: 11, textAlign: "center" }}> Pay</Box>
            <Box style={{ padding: "6px", fontSize: 11, textAlign: "center" }}>Shop Pay</Box>
          </div>
        </div>
      </div>
    </div>
  );
}
function CartV2Mobile() {
  return (
    <div style={{ height: "100%", position: "relative" }}>
      {/* background page dimmed */}
      <div style={{ position: "absolute", inset: 0, background: "rgba(0,0,0,0.25)" }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, background: "var(--paper)", borderTop: "1.5px solid var(--ink)", borderRadius: "14px 14px 0 0", padding: "10px 12px", height: "75%" }}>
        <div style={{ width: 40, height: 4, background: "var(--ink-2)", borderRadius: 999, margin: "0 auto 8px" }} />
        <Row>
          <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>Your Bag (2)</span>
          <span>✕</span>
        </Row>
        {[0,1].map(i => (
          <Row key={i} gap={6} style={{ padding: "8px 0", borderTop: "1px solid var(--ink-3)" }}>
            <ImgPh w={50} h={50} />
            <Col gap={1} style={{ flex: 1 }}>
              <div style={{ fontSize: 11 }}>XR-{7-i}</div>
              <div className="stepper" style={{ fontSize: 10, marginTop: 3 }}><span>−</span><span>1</span><span>+</span></div>
            </Col>
            <div style={{ fontSize: 12 }}>${349-i*100}</div>
          </Row>
        ))}
        <div style={{ position: "absolute", bottom: 10, left: 12, right: 12 }}>
          <Btn fill style={{ width: "100%" }}>Checkout · $593</Btn>
        </div>
      </div>
    </div>
  );
}

/* ===== CHECKOUT ===== */
function CheckoutV1Desktop() {
  return (
    <div style={{ padding: "0 32px", height: "100%" }}>
      <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>◯</span>
        <span style={{ fontSize: 12, color: "var(--ink-2)" }}>🔒 Secure checkout</span>
      </Row>
      {/* steps */}
      <Row gap={24} style={{ padding: "14px 0", fontSize: 13 }}>
        <span style={{ color: "var(--accent)", fontWeight: 600 }}>① Shipping</span>
        <span style={{ color: "var(--ink-3)" }}>② Delivery</span>
        <span style={{ color: "var(--ink-3)" }}>③ Payment</span>
        <span style={{ color: "var(--ink-3)" }}>④ Review</span>
      </Row>
      <Row gap={24}>
        <Col gap={14} style={{ flex: 2 }}>
          <div className="hand-title" style={{ fontSize: 20 }}>Shipping address</div>
          <Box style={{ padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>Email</Box>
          <Row gap={10}>
            <Box style={{ flex: 1, padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>First name</Box>
            <Box style={{ flex: 1, padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>Last name</Box>
          </Row>
          <Box style={{ padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>Address line 1</Box>
          <Row gap={10}>
            <Box style={{ flex: 2, padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>City</Box>
            <Box style={{ flex: 1, padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>State</Box>
            <Box style={{ flex: 1, padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>Zip</Box>
          </Row>
          <Box style={{ padding: "10px 12px", fontSize: 13, color: "var(--ink-3)" }}>Phone (for delivery)</Box>
          <Btn fill style={{ alignSelf: "flex-start" }}>Continue to delivery ›</Btn>
        </Col>
        <Box style={{ flex: 1, padding: 16, height: "fit-content" }} fill>
          <div className="hand-title" style={{ fontSize: 16, marginBottom: 8 }}>Order (2)</div>
          {[0,1].map(i => (
            <Row key={i} gap={8} style={{ padding: "6px 0", borderTop: "1px solid var(--ink-3)" }}>
              <ImgPh w={40} h={40} />
              <Col gap={0} style={{ flex: 1 }}>
                <div style={{ fontSize: 11 }}>XR-{7-i}</div>
                <div style={{ fontSize: 10, color: "var(--ink-2)" }}>Qty 1</div>
              </Col>
              <div style={{ fontSize: 12 }}>${349-i*100}</div>
            </Row>
          ))}
          <div style={{ borderTop: "1.5px solid var(--ink)", marginTop: 10, paddingTop: 8 }}>
            <Row style={{ fontSize: 12 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>Subtotal</span><span>$548</span></Row>
            <Row style={{ fontSize: 12 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>Shipping</span><span>Free</span></Row>
            <Row style={{ fontSize: 14, marginTop: 6 }}><span className="hand-title" style={{ flex: 1 }}>Total</span><span className="hand-title">$593.21</span></Row>
          </div>
        </Box>
      </Row>
    </div>
  );
}
function CheckoutV1Mobile() {
  return (
    <div style={{ padding: "0 12px", height: "100%" }}>
      <Row style={{ padding: "8px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span style={{ flex: 1 }}>← Checkout</span>
        <span style={{ fontSize: 10, color: "var(--ink-2)" }}>🔒</span>
      </Row>
      <Row gap={6} style={{ fontSize: 10, padding: "8px 0" }}>
        <span style={{ color: "var(--accent)" }}>① Shipping</span>
        <span style={{ color: "var(--ink-3)" }}>②</span>
        <span style={{ color: "var(--ink-3)" }}>③</span>
        <span style={{ color: "var(--ink-3)" }}>④</span>
      </Row>
      <Box style={{ padding: "8px 10px", fontSize: 11, color: "var(--ink-3)", marginBottom: 6 }}>Email</Box>
      <Row gap={6} style={{ marginBottom: 6 }}>
        <Box style={{ flex: 1, padding: "8px 10px", fontSize: 11, color: "var(--ink-3)" }}>First</Box>
        <Box style={{ flex: 1, padding: "8px 10px", fontSize: 11, color: "var(--ink-3)" }}>Last</Box>
      </Row>
      <Box style={{ padding: "8px 10px", fontSize: 11, color: "var(--ink-3)", marginBottom: 6 }}>Address</Box>
      <Row gap={6} style={{ marginBottom: 6 }}>
        <Box style={{ flex: 2, padding: "8px 10px", fontSize: 11, color: "var(--ink-3)" }}>City</Box>
        <Box style={{ flex: 1, padding: "8px 10px", fontSize: 11, color: "var(--ink-3)" }}>Zip</Box>
      </Row>
      <Btn fill style={{ width: "100%", marginTop: 4 }}>Continue ›</Btn>
    </div>
  );
}

function CheckoutV2Desktop() {
  return (
    <div style={{ padding: "0 32px", height: "100%" }}>
      <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>◯</span>
      </Row>
      <div className="hand-title" style={{ fontSize: 22, marginTop: 12 }}>Checkout.</div>
      <Row gap={24} style={{ marginTop: 10 }}>
        <Col gap={14} style={{ flex: 2 }}>
          {/* express */}
          <Row gap={8}>
            <Box style={{ flex: 1, padding: "10px", textAlign: "center" }} fill> Pay</Box>
            <Box style={{ flex: 1, padding: "10px", textAlign: "center" }} fill>G Pay</Box>
            <Box style={{ flex: 1, padding: "10px", textAlign: "center" }} fill>Shop Pay</Box>
          </Row>
          <Row style={{ alignItems: "center", gap: 8, color: "var(--ink-3)", fontSize: 12 }}>
            <div style={{ flex: 1, borderTop: "1px solid var(--ink-3)" }} /><span>or checkout in one page</span><div style={{ flex: 1, borderTop: "1px solid var(--ink-3)" }} />
          </Row>
          {/* single-page accordion */}
          {["Contact","Delivery","Payment"].map((s, i) => (
            <Box key={s} style={{ padding: 14, borderRadius: 8 }}>
              <Row style={{ alignItems: "center" }}>
                <div className="hand-title" style={{ fontSize: 16, flex: 1 }}>{i+1}. {s}</div>
                {i < 1 && <span style={{ fontSize: 11, color: "var(--ink-2)" }} className="squiggle">Edit</span>}
              </Row>
              {i === 0 && <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 4 }}>jane@example.com</div>}
              {i === 1 && (
                <Col gap={6} style={{ marginTop: 8 }}>
                  <Box style={{ padding: "8px 10px", fontSize: 12, color: "var(--ink-3)" }}>Address</Box>
                  <Row gap={6}>
                    <Box style={{ flex: 2, padding: "8px 10px", fontSize: 12, color: "var(--ink-3)" }}>City</Box>
                    <Box style={{ flex: 1, padding: "8px 10px", fontSize: 12, color: "var(--ink-3)" }}>Zip</Box>
                  </Row>
                  <Row gap={6} style={{ marginTop: 4 }}>
                    {["Standard · Free","Express · $12","Same-day · $22"].map((o,k) => (
                      <Box key={o} style={{ flex: 1, padding: 8, fontSize: 11, borderRadius: 4, ...(k===0 ? { borderColor: "var(--accent)", color: "var(--accent)" } : {}) }}>
                        <div>{o}</div>
                      </Box>
                    ))}
                  </Row>
                </Col>
              )}
              {i === 2 && <div style={{ fontSize: 12, color: "var(--ink-3)", marginTop: 6 }}>— locked until above complete —</div>}
            </Box>
          ))}
        </Col>
        <Box style={{ flex: 1, padding: 14, height: "fit-content" }} fill>
          <div className="hand-title" style={{ fontSize: 16 }}>Summary</div>
          <Row style={{ fontSize: 13, marginTop: 8 }}><span style={{ flex: 1 }}>2 items</span><span>$548</span></Row>
          <Row style={{ fontSize: 13 }}><span style={{ flex: 1 }}>Shipping</span><span>Free</span></Row>
          <Row style={{ fontSize: 13 }}><span style={{ flex: 1 }}>Tax</span><span>$45.21</span></Row>
          <div style={{ borderTop: "1.5px solid var(--ink)", margin: "10px 0" }} />
          <Row style={{ fontSize: 16 }}><span className="hand-title" style={{ flex: 1 }}>Total</span><span className="hand-title">$593.21</span></Row>
          <Box style={{ marginTop: 10, padding: 8, fontSize: 11, borderRadius: 4 }} accent>
            PROMO15 — $30 off applied ✓
          </Box>
        </Box>
      </Row>
    </div>
  );
}
function CheckoutV2Mobile() {
  return (
    <div style={{ padding: "0 10px", height: "100%" }}>
      <Row style={{ padding: "8px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span style={{ flex: 1 }}>← Checkout</span>
      </Row>
      <Row gap={4} style={{ marginTop: 8 }}>
        <Box style={{ flex: 1, padding: 6, textAlign: "center", fontSize: 10 }} fill> Pay</Box>
        <Box style={{ flex: 1, padding: 6, textAlign: "center", fontSize: 10 }} fill>Shop</Box>
      </Row>
      <div style={{ fontSize: 10, color: "var(--ink-3)", textAlign: "center", margin: "8px 0" }}>— or —</div>
      {["Contact ✓","Delivery","Payment"].map((s,i) => (
        <Box key={s} style={{ padding: 8, marginBottom: 6 }}>
          <div className="hand-title" style={{ fontSize: 13 }}>{i+1}. {s}</div>
          {i===1 && <Box style={{ padding: "6px 8px", fontSize: 10, color: "var(--ink-3)", marginTop: 4 }}>Address…</Box>}
        </Box>
      ))}
      <Btn fill style={{ width: "100%", marginTop: 6 }}>Place order — $593</Btn>
    </div>
  );
}

/* ===== ACCOUNT ===== */
function AccountV1Desktop() {
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ fontSize: 16, flex: 1 }}>◯</span>
        <span>👤 Jane</span>
      </Row>
      <Row gap={20} style={{ marginTop: 14 }}>
        <Col gap={4} style={{ width: 170, flexShrink: 0 }}>
          <div className="hand-title" style={{ fontSize: 16, marginBottom: 4 }}>Jane Doe</div>
          <div style={{ fontSize: 11, color: "var(--ink-2)", marginBottom: 10 }}>Member since 2023</div>
          {[["Orders", true],["Addresses"],["Payment methods"],["Wishlist"],["Subscriptions"],["Reviews"],["Settings"],["Sign out"]].map(([l, active]) => (
            <div key={l} style={{ fontSize: 13, padding: "6px 8px", borderRadius: 4, background: active ? "rgba(0,0,0,0.05)" : "transparent", color: active ? "var(--ink)" : "var(--ink-2)" }}>{l}</div>
          ))}
        </Col>
        <div style={{ flex: 1 }}>
          <div className="hand-title" style={{ fontSize: 24 }}>Orders</div>
          <Row gap={6} style={{ marginTop: 10 }}>
            {["All (14)","In progress (2)","Delivered","Cancelled"].map((t,i) => (
              <Box key={t} style={{ padding: "4px 10px", fontSize: 12, borderRadius: 999, ...(i===0 ? { borderColor: "var(--ink)", background: "rgba(0,0,0,0.05)" } : {}) }}>{t}</Box>
            ))}
          </Row>
          <Col gap={10} style={{ marginTop: 14 }}>
            {[{n:"#A-10421",d:"Apr 18, 2026",s:"Out for delivery",i:2,p:"$593.21"},{n:"#A-10398",d:"Apr 9, 2026",s:"Delivered",i:1,p:"$129"},{n:"#A-10322",d:"Mar 24, 2026",s:"Delivered",i:3,p:"$849"}].map(o => (
              <Box key={o.n} style={{ padding: 14, borderRadius: 8 }}>
                <Row style={{ alignItems: "baseline" }}>
                  <div className="hand-title" style={{ fontSize: 15 }}>Order {o.n}</div>
                  <div style={{ marginLeft: 12, fontSize: 12, color: "var(--ink-2)" }}>{o.d} · {o.i} items · {o.p}</div>
                  <div style={{ marginLeft: "auto" }}>
                    <Tag accent={o.s.includes("Out")}>{o.s}</Tag>
                  </div>
                </Row>
                <Row gap={6} style={{ marginTop: 10 }}>
                  {Array.from({length:o.i}).map((_,k) => <ImgPh key={k} w={60} h={60} />)}
                  <div style={{ flex: 1 }} />
                  <Btn size="sm">Track</Btn>
                  <Btn size="sm">Invoice</Btn>
                </Row>
              </Box>
            ))}
          </Col>
        </div>
      </Row>
    </div>
  );
}
function AccountV1Mobile() {
  return (
    <div style={{ padding: "0 10px", height: "100%" }}>
      <Row style={{ padding: "8px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ flex: 1, fontSize: 14 }}>Account</span>
        <span style={{ fontSize: 11 }}>⚙</span>
      </Row>
      <div style={{ padding: "10px 0" }}>
        <div className="hand-title" style={{ fontSize: 16 }}>Jane Doe</div>
        <div style={{ fontSize: 11, color: "var(--ink-2)" }}>jane@example.com</div>
      </div>
      <div className="hand-title" style={{ fontSize: 14, marginTop: 4 }}>Recent orders</div>
      <Col gap={6} style={{ marginTop: 6 }}>
        {[{n:"#A-10421",s:"Out for delivery"},{n:"#A-10398",s:"Delivered"}].map(o => (
          <Box key={o.n} style={{ padding: 8 }}>
            <Row><div className="hand-title" style={{ fontSize: 12, flex: 1 }}>{o.n}</div><Tag accent={o.s.includes("Out")}>{o.s}</Tag></Row>
            <Row gap={4} style={{ marginTop: 6 }}>{[0,1].map(i => <ImgPh key={i} w={40} h={40} />)}</Row>
          </Box>
        ))}
      </Col>
      <Col gap={2} style={{ marginTop: 10 }}>
        {["Addresses","Payment","Wishlist","Settings"].map(l => <Box key={l} style={{ padding: "8px 10px", fontSize: 12, display: "flex" }}><span style={{ flex: 1 }}>{l}</span><span>›</span></Box>)}
      </Col>
    </div>
  );
}

function AccountV2Desktop() {
  // order-tracking focused / map-like
  return (
    <div style={{ padding: "0 24px", height: "100%" }}>
      <Row style={{ padding: "10px 0", borderBottom: "1.2px solid var(--ink-2)" }}>
        <span className="hand-title" style={{ flex: 1, fontSize: 16 }}>◯</span>
        <Row gap={14} style={{ fontSize: 13, color: "var(--ink-2)" }}><span>Orders</span><span>Addresses</span><span>Payment</span><span>Settings</span></Row>
      </Row>
      <div className="hand-title" style={{ fontSize: 24, marginTop: 12 }}>Order #A-10421</div>
      <div style={{ fontSize: 13, color: "var(--ink-2)" }}>Placed Apr 18 · 2 items · $593.21</div>
      <Row gap={16} style={{ marginTop: 14 }}>
        {/* timeline */}
        <Col gap={0} style={{ flex: 1 }}>
          <div className="hand-title" style={{ fontSize: 15, marginBottom: 10 }}>Out for delivery — arrives today</div>
          {[{l:"Order placed",d:"Apr 18",done:true},{l:"Packed",d:"Apr 19",done:true},{l:"Shipped",d:"Apr 19",done:true},{l:"Out for delivery",d:"Apr 20, 8:12 AM",current:true},{l:"Delivered",d:"—"}].map((s,i,a) => (
            <Row key={s.l} gap={10} style={{ alignItems: "flex-start" }}>
              <Col gap={0} style={{ alignItems: "center", width: 18 }}>
                <div style={{ width: 14, height: 14, borderRadius: 999, border: "1.5px solid var(--ink)", background: s.done ? "var(--ink)" : s.current ? "var(--accent)" : "transparent", borderColor: s.current ? "var(--accent)" : "var(--ink)" }} />
                {i < a.length-1 && <div style={{ width: 1.5, flex: 1, minHeight: 24, background: s.done ? "var(--ink)" : "var(--ink-3)" }} />}
              </Col>
              <div style={{ paddingBottom: 16 }}>
                <div className="hand-title" style={{ fontSize: 14 }}>{s.l}</div>
                <div style={{ fontSize: 11, color: "var(--ink-2)" }}>{s.d}</div>
              </div>
            </Row>
          ))}
          <Btn accent size="sm" style={{ alignSelf: "flex-start" }}>Track live ›</Btn>
        </Col>
        {/* map + items */}
        <Col gap={12} style={{ flex: 1 }}>
          <Box style={{ height: 180, position: "relative" }} className="hatch">
            <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", color: "var(--ink-2)", fontSize: 13 }}>map · driver position</div>
          </Box>
          <Box style={{ padding: 12 }}>
            <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Delivering to</div>
            <div className="hand-title" style={{ fontSize: 14 }}>221B Baker St · NY 10012</div>
          </Box>
          {[0,1].map(i => (
            <Row key={i} gap={8}>
              <ImgPh w={60} h={60} />
              <Col gap={1} style={{ flex: 1 }}>
                <div className="hand-title" style={{ fontSize: 14 }}>XR-{7-i} · Midnight</div>
                <div style={{ fontSize: 11, color: "var(--ink-2)" }}>Qty 1 · ${349-i*100}</div>
              </Col>
              <Btn size="sm">Return</Btn>
            </Row>
          ))}
        </Col>
      </Row>
    </div>
  );
}
function AccountV2Mobile() {
  return (
    <div style={{ padding: "0 12px", height: "100%" }}>
      <Row style={{ padding: "8px 0" }}><span style={{ flex: 1 }}>← Order #A-10421</span></Row>
      <Box style={{ padding: 8 }} accent>
        <div className="hand-title" style={{ fontSize: 13 }}>Out for delivery</div>
        <div style={{ fontSize: 10 }}>Arrives today · 2–4 PM</div>
      </Box>
      <Box style={{ height: 120, marginTop: 8, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, color: "var(--ink-2)" }} className="hatch">map</Box>
      <Col gap={0} style={{ marginTop: 10 }}>
        {["Placed","Packed","Shipped","Out for delivery","Delivered"].map((l,i) => (
          <Row key={l} gap={6} style={{ alignItems: "center", padding: "4px 0" }}>
            <div style={{ width: 10, height: 10, borderRadius: 999, border: "1px solid var(--ink)", background: i<3 ? "var(--ink)" : i===3 ? "var(--accent)" : "transparent", borderColor: i===3 ? "var(--accent)":"var(--ink)" }} />
            <div style={{ fontSize: 11, flex: 1 }}>{l}</div>
          </Row>
        ))}
      </Col>
      <Btn fill style={{ width: "100%", marginTop: 8 }} size="sm">Track live ›</Btn>
    </div>
  );
}

window.CartPages = [
  { key: "c1", title: "V1 · Full-page bag", caption: <>Dedicated cart page with order summary rail. <b>Standard ecommerce</b> — good for complex orders with promo codes, gift options.</>, desktop: <CartV1Desktop />, mobile: <CartV1Mobile />, ann: "safe" },
  { key: "c2", title: "V2 · Slide-over drawer", caption: <>Cart as a <b>drawer over browsing</b> — don't force a full page change. Better add-to-cart momentum. Pairs with recommendation rail.</>, desktop: <CartV2Desktop />, mobile: <CartV2Mobile />, ann: "flow" },
];
window.CheckoutPages = [
  { key: "k1", title: "V1 · Stepped checkout", caption: <>Four steps: Shipping → Delivery → Payment → Review. <b>Lowest friction for novices</b>; each step is focused. Summary locked on right.</>, desktop: <CheckoutV1Desktop />, mobile: <CheckoutV1Mobile />, ann: "safe" },
  { key: "k2", title: "V2 · Single-page accordion", caption: <>Express-pay row on top, everything else on one page as collapsing sections. <b>Fewer page loads</b>, better for returning customers.</>, desktop: <CheckoutV2Desktop />, mobile: <CheckoutV2Mobile />, ann: "fast" },
];
window.AccountPages = [
  { key: "a1", title: "V1 · Orders hub", caption: <>Left nav, orders list with status tags. The <b>standard "My Account"</b> pattern. Tabs for filter, tracking + invoice quick-actions.</>, desktop: <AccountV1Desktop />, mobile: <AccountV1Mobile />, ann: "safe" },
  { key: "a2", title: "V2 · Order detail w/ live tracking", caption: <>Deep into a single order: <b>vertical timeline + map</b>. This is where post-purchase anxiety gets resolved. Don't hide it two clicks deep.</>, desktop: <AccountV2Desktop />, mobile: <AccountV2Mobile />, ann: "care" },
];
