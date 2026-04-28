/* Cart V1, Checkout V2, Account V1 */

/* ===== Cart V1 — full page ===== */
function CartV1({ mobile }) {
  const items = [
    { p: PRODUCTS[0], qty: 1, variant: "Midnight · 256GB" },
    { p: PRODUCTS[3], qty: 1, variant: "Crimson · Sport Band" },
  ];
  const subtotal = items.reduce((s, i) => s + i.p.price * i.qty, 0);
  const tax = subtotal * 0.0825;
  const total = subtotal + tax;

  if (mobile) {
    return (
      <div className="frame-scroll" style={{ paddingBottom: 100 }}>
        <div style={{ position: "sticky", top: 0, zIndex: 30, background: "rgba(251,251,253,0.9)", backdropFilter: "blur(20px)", padding: "10px 14px", display: "flex", alignItems: "center", borderBottom: "1px solid var(--hairline)" }}>
          <IconChevL size={20} /><span style={{ flex: 1, textAlign: "center", fontSize: 15, fontWeight: 600 }}>Bag</span><IconUser size={18} />
        </div>
        <div style={{ padding: 16 }}>
          <h1 style={{ fontSize: 24, fontWeight: 600, letterSpacing: "-0.02em", margin: 0 }}>Your Bag.</h1>
          <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 2 }}>{items.length} items · Free shipping</div>
        </div>
        {items.map((it, i) => (
          <div key={i} style={{ padding: "14px 16px", borderTop: "1px solid var(--hairline)", display: "flex", gap: 12 }}>
            <div style={{ width: 72, height: 72, background: it.p.hero, borderRadius: 12, flexShrink: 0 }}>
              <ProductImage product={it.p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 14, fontWeight: 500 }}>{it.p.name}</div>
              <div style={{ fontSize: 11, color: "var(--ink-2)" }}>{it.variant}</div>
              <div style={{ fontSize: 11, color: "var(--ink-2)", marginTop: 2 }}>Ships by Thu</div>
              <div style={{ display: "flex", alignItems: "center", marginTop: 8 }}>
                <div style={{ display: "inline-flex", border: "1px solid var(--hairline)", borderRadius: 980 }}>
                  <button style={{ padding: "4px 8px", background: "transparent", border: 0, cursor: "pointer" }}><IconMinus size={12} /></button>
                  <div style={{ padding: "4px 10px", fontSize: 12 }}>{it.qty}</div>
                  <button style={{ padding: "4px 8px", background: "transparent", border: 0, cursor: "pointer" }}><IconPlus size={12} /></button>
                </div>
                <div style={{ marginLeft: "auto", fontSize: 14, fontWeight: 500 }}>{fmt(it.p.price)}</div>
              </div>
            </div>
          </div>
        ))}
        <div style={{ padding: 16, borderTop: "1px solid var(--hairline)" }}>
          <div style={{ display: "flex", gap: 8, marginBottom: 10 }}>
            <input className="mf-input" placeholder="Promo code" style={{ height: 38, fontSize: 13 }} />
            <Btn kind="secondary">Apply</Btn>
          </div>
          {[["Subtotal", fmt(subtotal)],["Shipping","Free"],["Tax", fmt(tax)]].map(([k,v]) => (
            <div key={k} style={{ display: "flex", padding: "4px 0", fontSize: 14 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>{k}</span><span>{v}</span></div>
          ))}
          <div style={{ height: 1, background: "var(--hairline)", margin: "10px 0" }} />
          <div style={{ display: "flex", fontSize: 17 }}><span style={{ flex: 1, fontWeight: 500 }}>Total</span><span style={{ fontWeight: 600 }}>{fmt(total)}</span></div>
        </div>
        <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, padding: 14, background: "rgba(255,255,255,0.95)", backdropFilter: "blur(20px)", borderTop: "1px solid var(--hairline)" }}>
          <Btn kind="primary" full size="lg">Checkout — {fmt(total)}</Btn>
        </div>
      </div>
    );
  }

  return (
    <div className="frame-scroll">
      <Nav active="store" cartCount={items.length} />
      <section style={{ maxWidth: 1100, margin: "0 auto", padding: "56px 40px 80px" }}>
        <div style={{ textAlign: "center", marginBottom: 32 }}>
          <h1 style={{ fontSize: 48, fontWeight: 600, letterSpacing: "-0.025em", margin: 0 }}>Review your bag.</h1>
          <div style={{ fontSize: 17, color: "var(--ink-2)", marginTop: 8 }}>
            Free shipping and free returns. <a style={{ color: "var(--accent)" }}>See details ›</a>
          </div>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 380px", gap: 40, alignItems: "start" }}>
          <div>
            {items.map((it, i) => (
              <div key={i} style={{ padding: "28px 0", borderTop: i === 0 ? "1px solid var(--hairline)" : "none", borderBottom: "1px solid var(--hairline)", display: "flex", gap: 24 }}>
                <div style={{ width: 170, height: 170, background: it.p.hero, borderRadius: 18, flexShrink: 0, padding: 14 }}>
                  <ProductImage product={it.p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ display: "flex", alignItems: "baseline" }}>
                    <div>
                      <div style={{ fontSize: 19, fontWeight: 600, letterSpacing: "-0.015em" }}>{it.p.name} <span style={{ fontWeight: 400, color: "var(--ink-2)" }}>— {it.p.tagline}</span></div>
                      <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 4 }}>{it.variant}</div>
                    </div>
                    <div style={{ marginLeft: "auto", fontSize: 19, fontWeight: 500 }}>{fmt(it.p.price)}</div>
                  </div>
                  <div style={{ marginTop: 14, display: "flex", gap: 10, alignItems: "center", fontSize: 13, color: "var(--ink-2)" }}>
                    <IconTruck size={16} /><span>In stock · Ships by Thu, Apr 23</span>
                  </div>
                  <div style={{ marginTop: 16, display: "flex", alignItems: "center", gap: 14 }}>
                    <div style={{ display: "inline-flex", border: "1px solid var(--hairline)", borderRadius: 980 }}>
                      <button style={{ padding: "6px 12px", background: "transparent", border: 0, cursor: "pointer" }}><IconMinus size={14} /></button>
                      <div style={{ padding: "6px 14px", borderLeft: "1px solid var(--hairline)", borderRight: "1px solid var(--hairline)", fontSize: 14 }}>{it.qty}</div>
                      <button style={{ padding: "6px 12px", background: "transparent", border: 0, cursor: "pointer" }}><IconPlus size={14} /></button>
                    </div>
                    <a style={{ color: "var(--accent)", fontSize: 13 }}>Save for later</a>
                    <a style={{ color: "var(--accent)", fontSize: 13 }}>Remove</a>
                  </div>
                </div>
              </div>
            ))}

            <div style={{ marginTop: 40 }}>
              <div style={{ fontSize: 22, fontWeight: 600, letterSpacing: "-0.015em" }}>You may also like.</div>
              <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16, marginTop: 16 }}>
                {PRODUCTS.slice(5,8).map(p => (
                  <div key={p.id} style={{ background: p.hero, borderRadius: 18, padding: 16 }}>
                    <ProductImage product={p} />
                    <div style={{ fontSize: 15, fontWeight: 500, marginTop: 8 }}>{p.name}</div>
                    <div style={{ fontSize: 13, color: "var(--ink-2)" }}>From {fmt(p.price)}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <aside className="wash-card" style={{ padding: 24, borderRadius: 22, position: "sticky", top: 64 }}>
            <div style={{ fontSize: 19, fontWeight: 600, letterSpacing: "-0.015em", marginBottom: 14 }}>Order summary</div>
            <div style={{ display: "flex", gap: 8, marginBottom: 14 }}>
              <input className="mf-input" placeholder="Promo code" style={{ height: 40, fontSize: 14 }} />
              <Btn kind="secondary">Apply</Btn>
            </div>
            {[["Subtotal", fmt(subtotal)],["Shipping","Free"],["Estimated tax", fmt(tax)]].map(([k,v]) => (
              <div key={k} style={{ display: "flex", padding: "6px 0", fontSize: 14 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>{k}</span><span style={{ color: "var(--ink)" }}>{v}</span></div>
            ))}
            <div style={{ height: 1, background: "var(--hairline)", margin: "12px 0" }} />
            <div style={{ display: "flex", fontSize: 19, marginBottom: 14 }}>
              <span style={{ flex: 1, fontWeight: 500 }}>Total</span>
              <span style={{ fontWeight: 600 }}>{fmt(total)}</span>
            </div>
            <Btn kind="primary" size="lg" full>Checkout</Btn>
            <div style={{ display: "flex", alignItems: "center", gap: 8, margin: "16px 0 10px", color: "var(--ink-3)", fontSize: 12 }}>
              <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} /> or <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} />
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
              <Btn kind="dark"> Pay</Btn>
              <Btn kind="outline">Shop Pay</Btn>
            </div>
            <div style={{ marginTop: 16, fontSize: 12, color: "var(--ink-2)", display: "flex", gap: 8, alignItems: "center" }}>
              <IconLock size={14} /><span>Secure checkout</span>
            </div>
          </aside>
        </div>
      </section>
    </div>
  );
}

/* ===== Checkout V2 — single-page accordion ===== */
function CheckoutV2({ mobile }) {
  const items = [
    { p: PRODUCTS[0], qty: 1, variant: "Midnight · 256GB" },
    { p: PRODUCTS[3], qty: 1, variant: "Crimson · Sport Band" },
  ];
  const subtotal = items.reduce((s, i) => s + i.p.price * i.qty, 0);
  const discount = 30;
  const tax = (subtotal - discount) * 0.0825;
  const total = subtotal - discount + tax;
  const [open, setOpen] = useState(1); // 0 contact done, 1 delivery open, 2 payment locked

  const Step = ({ i, label, done, body, summary }) => {
    const isOpen = open === i;
    return (
      <div style={{ background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 18, overflow: "hidden" }}>
        <div style={{ padding: "18px 22px", display: "flex", alignItems: "center", gap: 12, cursor: "pointer" }} onClick={() => setOpen(i)}>
          <div style={{ width: 24, height: 24, borderRadius: 999, background: done ? "var(--ink)" : isOpen ? "var(--accent)" : "var(--surface-3)", color: done || isOpen ? "white" : "var(--ink-2)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 12, fontWeight: 600 }}>
            {done ? <IconCheck size={14} stroke="white" /> : i + 1}
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 17, fontWeight: 600, letterSpacing: "-0.015em" }}>{label}</div>
            {done && !isOpen && <div style={{ fontSize: 13, color: "var(--ink-2)", marginTop: 2 }}>{summary}</div>}
          </div>
          {done && !isOpen && <a style={{ color: "var(--accent)", fontSize: 14 }}>Edit</a>}
        </div>
        {isOpen && <div style={{ padding: "0 22px 22px" }}>{body}</div>}
      </div>
    );
  };

  if (mobile) {
    return (
      <div className="frame-scroll">
        <div style={{ padding: "14px 16px", borderBottom: "1px solid var(--hairline)", display: "flex", alignItems: "center", gap: 8 }}>
          <IconChevL size={18} /><span style={{ flex: 1, fontSize: 15, fontWeight: 600 }}>Checkout</span><IconLock size={14} color="var(--ink-2)" />
        </div>
        <div style={{ padding: 14 }}>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8, marginBottom: 10 }}>
            <Btn kind="dark"> Pay</Btn>
            <Btn kind="outline">Shop Pay</Btn>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 8, color: "var(--ink-3)", fontSize: 11, margin: "10px 0" }}>
            <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} /> or <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} />
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            <div style={{ background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 14, padding: 14 }}>
              <div style={{ fontSize: 14, fontWeight: 600, display: "flex", alignItems: "center", gap: 8 }}><IconCheck size={14} /> Contact</div>
              <div style={{ fontSize: 12, color: "var(--ink-2)", marginTop: 4 }}>jane@example.com</div>
            </div>
            <div style={{ background: "var(--surface)", border: "2px solid var(--accent)", borderRadius: 14, padding: 14 }}>
              <div style={{ fontSize: 14, fontWeight: 600 }}>Delivery</div>
              <input className="mf-input" placeholder="Address" style={{ height: 36, fontSize: 12, marginTop: 8 }} />
              <div style={{ display: "grid", gridTemplateColumns: "2fr 1fr", gap: 6, marginTop: 6 }}>
                <input className="mf-input" placeholder="City" style={{ height: 36, fontSize: 12 }} />
                <input className="mf-input" placeholder="Zip" style={{ height: 36, fontSize: 12 }} />
              </div>
              <div style={{ fontSize: 12, color: "var(--ink-2)", marginTop: 10 }}>Shipping method</div>
              {[["Standard","Free","Apr 23"],["Express","$12","Apr 21"]].map(([m,p,d], i) => (
                <div key={m} style={{ display: "flex", alignItems: "center", padding: "8px 10px", marginTop: 4, border: i === 0 ? "2px solid var(--ink)" : "1px solid var(--hairline)", borderRadius: 10, fontSize: 12 }}>
                  <span style={{ flex: 1, fontWeight: 500 }}>{m}</span><span style={{ color: "var(--ink-2)" }}>{d}</span><span style={{ marginLeft: 10 }}>{p}</span>
                </div>
              ))}
            </div>
            <div style={{ background: "var(--surface-3)", borderRadius: 14, padding: 14, color: "var(--ink-3)", fontSize: 12 }}>
              3. Payment — locked
            </div>
          </div>
          <div style={{ background: "var(--surface-3)", borderRadius: 14, padding: 14, marginTop: 14, fontSize: 13 }}>
            <div style={{ fontWeight: 600, marginBottom: 6 }}>Summary</div>
            {[["Subtotal",fmt(subtotal)],["Discount",`-${fmt(discount)}`],["Shipping","Free"],["Tax",fmt(tax)]].map(([k,v]) => (
              <div key={k} style={{ display: "flex", padding: "2px 0", fontSize: 12 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>{k}</span><span>{v}</span></div>
            ))}
            <div style={{ height: 1, background: "var(--hairline)", margin: "8px 0" }} />
            <div style={{ display: "flex", fontWeight: 600 }}><span style={{ flex: 1 }}>Total</span><span>{fmt(total)}</span></div>
          </div>
          <Btn kind="primary" full size="lg" style={{ marginTop: 12 }}>Place order — {fmt(total)}</Btn>
        </div>
      </div>
    );
  }

  return (
    <div className="frame-scroll">
      <div style={{ height: 60, borderBottom: "1px solid var(--hairline)", display: "flex", alignItems: "center", padding: "0 40px" }}>
        <span style={{ fontSize: 18, fontWeight: 500, flex: 1 }}>◯</span>
        <span style={{ fontSize: 13, color: "var(--ink-2)", display: "flex", alignItems: "center", gap: 6 }}><IconLock size={14} /> Secure checkout</span>
      </div>

      <section style={{ maxWidth: 1100, margin: "0 auto", padding: "40px 40px 80px" }}>
        <h1 style={{ fontSize: 40, fontWeight: 600, letterSpacing: "-0.025em", margin: 0 }}>Checkout.</h1>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 380px", gap: 40, marginTop: 28, alignItems: "start" }}>
          <div>
            {/* Express pay */}
            <div style={{ background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 18, padding: 20 }}>
              <div style={{ fontSize: 13, color: "var(--ink-2)", marginBottom: 10 }}>Express checkout</div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 10 }}>
                <Btn kind="dark" size="lg"> Pay</Btn>
                <Btn kind="outline" size="lg" style={{ background: "#5a31f4", color: "white", border: 0 }}>Shop Pay</Btn>
                <Btn kind="outline" size="lg">G Pay</Btn>
              </div>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, margin: "20px 0", color: "var(--ink-3)", fontSize: 13 }}>
              <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} /> or checkout with details <div style={{ flex: 1, height: 1, background: "var(--hairline)" }} />
            </div>

            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              <Step i={0} label="Contact" done summary="jane@example.com · (415) 555-0129" body={null} />
              <Step
                i={1}
                label="Delivery"
                body={
                  <div style={{ display: "flex", flexDirection: "column", gap: 10, marginTop: 4 }}>
                    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
                      <input className="mf-input" placeholder="First name" defaultValue="Jane" />
                      <input className="mf-input" placeholder="Last name" defaultValue="Doe" />
                    </div>
                    <input className="mf-input" placeholder="Address line 1" defaultValue="221B Baker Street" />
                    <input className="mf-input" placeholder="Apartment, suite (optional)" />
                    <div style={{ display: "grid", gridTemplateColumns: "2fr 1fr 1fr", gap: 10 }}>
                      <input className="mf-input" placeholder="City" defaultValue="New York" />
                      <input className="mf-input" placeholder="State" defaultValue="NY" />
                      <input className="mf-input" placeholder="Zip" defaultValue="10012" />
                    </div>
                    <div style={{ marginTop: 8, fontSize: 13, color: "var(--ink-2)" }}>Shipping method</div>
                    {[
                      ["Standard Shipping","2–4 business days","Free"],
                      ["Express","1–2 business days","$12.00"],
                      ["Same-day delivery","By 9 PM today","$22.00"],
                    ].map(([m, s, p], i) => (
                      <div key={m} style={{ display: "flex", alignItems: "center", padding: "14px 16px", border: i === 0 ? "2px solid var(--ink)" : "1px solid var(--hairline)", borderRadius: 14, gap: 12 }}>
                        <span style={{ width: 18, height: 18, borderRadius: 999, border: "1px solid var(--ink-3)", background: i === 0 ? "var(--ink)" : "transparent", display: "inline-flex", alignItems: "center", justifyContent: "center" }}>
                          {i === 0 && <span style={{ width: 7, height: 7, borderRadius: 999, background: "white" }} />}
                        </span>
                        <div style={{ flex: 1 }}>
                          <div style={{ fontSize: 15, fontWeight: 500 }}>{m}</div>
                          <div style={{ fontSize: 13, color: "var(--ink-2)" }}>{s}</div>
                        </div>
                        <div style={{ fontSize: 15, fontWeight: 500 }}>{p}</div>
                      </div>
                    ))}
                    <Btn kind="primary" size="lg" style={{ alignSelf: "flex-start", marginTop: 8 }}>Continue to payment</Btn>
                  </div>
                }
              />
              <Step i={2} label="Payment" body={<div style={{ color: "var(--ink-3)", padding: "14px 0", fontSize: 14 }}>Complete the previous step to enter payment.</div>} />
            </div>
          </div>

          {/* Summary */}
          <aside style={{ background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 22, padding: 24, position: "sticky", top: 24 }}>
            <div style={{ fontSize: 17, fontWeight: 600, letterSpacing: "-0.015em" }}>Order · 2 items</div>
            <div style={{ marginTop: 14, display: "flex", flexDirection: "column", gap: 14 }}>
              {items.map((it, i) => (
                <div key={i} style={{ display: "flex", gap: 12 }}>
                  <div style={{ width: 64, height: 64, background: it.p.hero, borderRadius: 12, position: "relative", flexShrink: 0 }}>
                    <ProductImage product={it.p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
                    <span style={{ position: "absolute", top: -6, right: -6, background: "var(--ink-2)", color: "white", fontSize: 10, minWidth: 18, height: 18, borderRadius: 999, display: "flex", alignItems: "center", justifyContent: "center", fontWeight: 600 }}>{it.qty}</span>
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 13, fontWeight: 500 }}>{it.p.name}</div>
                    <div style={{ fontSize: 11, color: "var(--ink-2)" }}>{it.variant}</div>
                  </div>
                  <div style={{ fontSize: 13, fontWeight: 500 }}>{fmt(it.p.price)}</div>
                </div>
              ))}
            </div>
            <div style={{ height: 1, background: "var(--hairline)", margin: "18px 0" }} />
            <div style={{ display: "flex", gap: 8, marginBottom: 14 }}>
              <input className="mf-input" placeholder="Promo code" style={{ height: 38, fontSize: 13 }} />
              <Btn kind="secondary" size="sm">Apply</Btn>
            </div>
            <div style={{ background: "rgba(52,199,89,0.08)", border: "1px solid rgba(52,199,89,0.25)", borderRadius: 10, padding: "8px 12px", fontSize: 12, color: "#1a7a33", display: "flex", alignItems: "center", gap: 6 }}>
              <IconCheck size={14} stroke="#1a7a33" /> PROMO15 applied — $30 off
            </div>
            <div style={{ marginTop: 14 }}>
              {[["Subtotal", fmt(subtotal)],["Discount", `-${fmt(discount)}`],["Shipping","Free"],["Estimated tax", fmt(tax)]].map(([k, v]) => (
                <div key={k} style={{ display: "flex", padding: "4px 0", fontSize: 14 }}><span style={{ flex: 1, color: "var(--ink-2)" }}>{k}</span><span>{v}</span></div>
              ))}
              <div style={{ height: 1, background: "var(--hairline)", margin: "10px 0" }} />
              <div style={{ display: "flex", fontSize: 19 }}><span style={{ flex: 1, fontWeight: 500 }}>Total</span><span style={{ fontWeight: 600 }}>{fmt(total)}</span></div>
            </div>
          </aside>
        </div>
      </section>
    </div>
  );
}

/* ===== Account V1 — orders hub ===== */
function AccountV1({ mobile }) {
  const orders = [
    { n: "A-10421", d: "Apr 18, 2026", status: "Out for delivery", statusColor: "#0071e3", items: [PRODUCTS[0], PRODUCTS[3]], total: 778 },
    { n: "A-10398", d: "Apr 9, 2026", status: "Delivered", statusColor: "#34c759", items: [PRODUCTS[5]], total: 599 },
    { n: "A-10322", d: "Mar 24, 2026", status: "Delivered", statusColor: "#34c759", items: [PRODUCTS[2]], total: 1999 },
    { n: "A-10207", d: "Feb 12, 2026", status: "Returned", statusColor: "#6e6e73", items: [PRODUCTS[7]], total: 99 },
  ];

  if (mobile) {
    return (
      <div className="frame-scroll">
        <div style={{ padding: "14px 16px", borderBottom: "1px solid var(--hairline)", display: "flex", alignItems: "center" }}>
          <span style={{ flex: 1, fontSize: 17, fontWeight: 600 }}>Account</span>
          <IconSearch size={18} />
        </div>
        <div style={{ padding: 16, borderBottom: "1px solid var(--hairline)" }}>
          <div style={{ width: 50, height: 50, borderRadius: 999, background: "var(--wash-ocean)", display: "inline-flex", alignItems: "center", justifyContent: "center", fontSize: 18, fontWeight: 500, color: "var(--accent)" }}>JD</div>
          <div style={{ fontSize: 19, fontWeight: 600, marginTop: 8 }}>Jane Doe</div>
          <div style={{ fontSize: 12, color: "var(--ink-2)" }}>jane@example.com · Member since 2023</div>
        </div>
        <div style={{ padding: 16 }}>
          <div style={{ fontSize: 15, fontWeight: 600, marginBottom: 10 }}>Recent orders</div>
          {orders.slice(0, 3).map(o => (
            <div key={o.n} style={{ padding: 12, marginBottom: 8, border: "1px solid var(--hairline)", borderRadius: 14 }}>
              <div style={{ display: "flex", alignItems: "center" }}>
                <div>
                  <div style={{ fontSize: 13, fontWeight: 600 }}>Order #{o.n}</div>
                  <div style={{ fontSize: 11, color: "var(--ink-2)" }}>{o.d} · {fmt(o.total)}</div>
                </div>
                <span style={{ marginLeft: "auto", fontSize: 11, color: o.statusColor, fontWeight: 600 }}>● {o.status}</span>
              </div>
              <div style={{ display: "flex", gap: 6, marginTop: 10 }}>
                {o.items.map(p => (
                  <div key={p.id} style={{ width: 42, height: 42, background: p.hero, borderRadius: 10 }}>
                    <ProductImage product={p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
        <div style={{ borderTop: "1px solid var(--hairline)" }}>
          {["Addresses","Payment methods","Wishlist (3)","Subscriptions","Settings","Sign out"].map(l => (
            <div key={l} style={{ padding: "14px 16px", display: "flex", alignItems: "center", borderBottom: "1px solid var(--hairline-soft)", fontSize: 14 }}>
              <span style={{ flex: 1 }}>{l}</span><IconChevR size={16} />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="frame-scroll">
      <Nav active="store" />
      <section style={{ maxWidth: 1280, margin: "0 auto", padding: "40px 40px 80px", display: "grid", gridTemplateColumns: "260px 1fr", gap: 40 }}>
        <aside>
          <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 18 }}>
            <div style={{ width: 48, height: 48, borderRadius: 999, background: "var(--wash-ocean)", display: "flex", alignItems: "center", justifyContent: "center", color: "var(--accent)", fontWeight: 600, fontSize: 17 }}>JD</div>
            <div>
              <div style={{ fontSize: 15, fontWeight: 500 }}>Jane Doe</div>
              <div style={{ fontSize: 12, color: "var(--ink-2)" }}>Member since 2023</div>
            </div>
          </div>
          <nav style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {[
              { l: "Orders", active: true },
              { l: "Addresses" },
              { l: "Payment methods" },
              { l: "Wishlist", b: 3 },
              { l: "Subscriptions" },
              { l: "Reviews" },
              { l: "Settings" },
              { l: "Sign out" },
            ].map(i => (
              <a key={i.l} style={{ padding: "9px 14px", borderRadius: 10, fontSize: 14, color: i.active ? "var(--ink)" : "var(--ink-2)", background: i.active ? "var(--surface-3)" : "transparent", display: "flex", alignItems: "center", cursor: "pointer" }}>
                <span style={{ flex: 1, fontWeight: i.active ? 500 : 400 }}>{i.l}</span>
                {i.b && <span style={{ fontSize: 11, color: "var(--ink-3)" }}>{i.b}</span>}
              </a>
            ))}
          </nav>
        </aside>

        <div>
          <div style={{ display: "flex", alignItems: "baseline", marginBottom: 20 }}>
            <h1 style={{ fontSize: 44, fontWeight: 600, letterSpacing: "-0.025em", margin: 0, flex: 1 }}>Orders.</h1>
            <div className="segmented">
              <div className="segmented-item active">All</div>
              <div className="segmented-item">In progress</div>
              <div className="segmented-item">Delivered</div>
              <div className="segmented-item">Returned</div>
            </div>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            {orders.map(o => (
              <article key={o.n} style={{ background: "var(--surface)", border: "1px solid var(--hairline)", borderRadius: 18, overflow: "hidden" }}>
                <div style={{ padding: "16px 22px", background: "var(--surface-3)", borderBottom: "1px solid var(--hairline)", display: "flex", alignItems: "center", gap: 30, fontSize: 13 }}>
                  <div>
                    <div style={{ color: "var(--ink-3)", fontSize: 11, textTransform: "uppercase", letterSpacing: "0.04em", fontWeight: 600 }}>Order placed</div>
                    <div style={{ color: "var(--ink)", marginTop: 2 }}>{o.d}</div>
                  </div>
                  <div>
                    <div style={{ color: "var(--ink-3)", fontSize: 11, textTransform: "uppercase", letterSpacing: "0.04em", fontWeight: 600 }}>Total</div>
                    <div style={{ color: "var(--ink)", marginTop: 2 }}>{fmt(o.total)}</div>
                  </div>
                  <div>
                    <div style={{ color: "var(--ink-3)", fontSize: 11, textTransform: "uppercase", letterSpacing: "0.04em", fontWeight: 600 }}>Order #</div>
                    <div style={{ color: "var(--ink)", marginTop: 2 }}>{o.n}</div>
                  </div>
                  <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: 6 }}>
                    <span style={{ width: 8, height: 8, borderRadius: 999, background: o.statusColor }} />
                    <span style={{ color: o.statusColor, fontWeight: 600, fontSize: 13 }}>{o.status}</span>
                  </div>
                </div>
                <div style={{ padding: "18px 22px", display: "flex", gap: 18, alignItems: "center" }}>
                  {o.items.map((p, i) => (
                    <div key={i} style={{ display: "flex", gap: 12, alignItems: "center" }}>
                      <div style={{ width: 72, height: 72, background: p.hero, borderRadius: 14 }}>
                        <ProductImage product={p} style={{ borderRadius: 0, background: "transparent" }} showShadow={false} />
                      </div>
                      <div>
                        <div style={{ fontSize: 14, fontWeight: 500 }}>{p.name}</div>
                        <div style={{ fontSize: 12, color: "var(--ink-2)" }}>{p.tagline}</div>
                        <a style={{ color: "var(--accent)", fontSize: 12, marginTop: 4, display: "inline-block" }}>Buy again ›</a>
                      </div>
                    </div>
                  ))}
                  <div style={{ marginLeft: "auto", display: "flex", flexDirection: "column", gap: 8 }}>
                    {o.status === "Out for delivery" ? (
                      <Btn kind="primary" size="sm">Track package</Btn>
                    ) : (
                      <Btn kind="outline" size="sm">View order</Btn>
                    )}
                    <Btn kind="ghost" size="sm">Invoice</Btn>
                  </div>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}

window.CartV1 = CartV1;
window.CheckoutV2 = CheckoutV2;
window.AccountV1 = AccountV1;
