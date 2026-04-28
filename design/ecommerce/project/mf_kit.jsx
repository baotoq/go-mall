/* Mid-fi Apple-styled component kit for ecommerce prototype.
   Exports to window so multi-script Babel setup works. */

const { useState, useEffect, useRef } = React;

/* -------- Product data (shared) -------- */

const PRODUCTS = [
  { id: "xr7", name: "XR-7", tagline: "Wireless Headphones", price: 349, category: "Audio", rating: 4.8, reviews: 1204, color: "#1d1d1f", gradFrom: "#d8d8dc", gradTo: "#f5f5f7", hero: "#eaeaef" },
  { id: "air", name: "Airphones 3", tagline: "Open-ear Buds", price: 199, category: "Audio", rating: 4.6, reviews: 821, color: "#e0dcd1", gradFrom: "#f3ece0", gradTo: "#ffffff", hero: "#f3ece0" },
  { id: "lmb", name: "Luma 16", tagline: "Laptop, M3 Pro", price: 1999, category: "Laptops", rating: 4.9, reviews: 312, color: "#2a2a2c", gradFrom: "#e8e8ed", gradTo: "#f5f5f7", hero: "#e2e3e7" },
  { id: "cir", name: "Circle Watch", tagline: "Fitness & Health", price: 429, category: "Watches", rating: 4.7, reviews: 902, color: "#c8382f", gradFrom: "#fde4d3", gradTo: "#ffffff", hero: "#fde4d3" },
  { id: "phn", name: "Phone 15 Pro", tagline: "Titanium, 256GB", price: 1099, category: "Phones", rating: 4.8, reviews: 2481, color: "#7a776e", gradFrom: "#e3e0d5", gradTo: "#f5f5f7", hero: "#ece9e0" },
  { id: "tab", name: "Slate 11", tagline: "Tablet, Wi-Fi", price: 599, category: "Tablets", rating: 4.7, reviews: 604, color: "#2a5a8a", gradFrom: "#e3f0fa", gradTo: "#ffffff", hero: "#e3f0fa" },
  { id: "cam", name: "Lens One", tagline: "Mirrorless Camera", price: 1299, category: "Cameras", rating: 4.9, reviews: 190, color: "#1a1a1a", gradFrom: "#e8e8ed", gradTo: "#ffffff", hero: "#eaeaef" },
  { id: "spk", name: "Orb Mini", tagline: "Home Speaker", price: 99, category: "Audio", rating: 4.5, reviews: 430, color: "#af52de", gradFrom: "#f1e6f7", gradTo: "#ffffff", hero: "#f1e6f7" },
];

/* -------- Product image — CSS-only generated silhouette --------
   Draws a product-shaped silhouette on a soft gradient. No external assets. */

function ProductImage({ product, size = "md", style, showShadow = true, className = "" }) {
  const p = product;
  const aspect = size === "hero" ? "1 / 1" : size === "wide" ? "16 / 10" : size === "sq" ? "1/1" : "4 / 3";
  const silhouette = SILHOUETTES[p.id] || SILHOUETTES.xr7;

  return (
    <div
      className={`product-img ${className}`}
      style={{
        aspectRatio: aspect,
        background: `radial-gradient(ellipse at 50% 40%, ${p.gradFrom} 0%, ${p.gradTo} 72%)`,
        borderRadius: 18,
        overflow: "hidden",
        position: "relative",
        ...style,
      }}
    >
      <div style={{
        position: "absolute",
        inset: 0,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}>
        {silhouette(p.color)}
      </div>
      {showShadow && (
        <div style={{
          position: "absolute",
          left: "20%", right: "20%", bottom: "10%",
          height: 14,
          borderRadius: "50%",
          background: "radial-gradient(ellipse at center, rgba(0,0,0,0.14), transparent 70%)",
          filter: "blur(4px)",
        }} />
      )}
    </div>
  );
}

/* SVG product silhouettes — simple, Apple-style product renders in CSS */
const SILHOUETTES = {
  xr7: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "70%", height: "70%" }}>
      {/* headphones */}
      <path d={`M 40 100 Q 40 40 100 40 Q 160 40 160 100`} fill="none" stroke={c} strokeWidth="10" strokeLinecap="round" />
      <rect x="24" y="92" width="34" height="60" rx="14" fill={c} />
      <rect x="142" y="92" width="34" height="60" rx="14" fill={c} />
      <ellipse cx="41" cy="122" rx="10" ry="18" fill="rgba(255,255,255,0.18)" />
      <ellipse cx="159" cy="122" rx="10" ry="18" fill="rgba(255,255,255,0.18)" />
    </svg>
  ),
  air: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "60%", height: "60%" }}>
      {/* earbud pair */}
      <g transform="translate(60 50)">
        <circle cx="0" cy="0" r="20" fill={c} />
        <rect x="-8" y="10" width="16" height="60" rx="6" fill={c} />
      </g>
      <g transform="translate(140 50)">
        <circle cx="0" cy="0" r="20" fill={c} />
        <rect x="-8" y="10" width="16" height="60" rx="6" fill={c} />
      </g>
    </svg>
  ),
  lmb: (c) => (
    <svg viewBox="0 0 220 160" style={{ width: "80%", height: "80%" }}>
      {/* laptop */}
      <rect x="30" y="20" width="160" height="100" rx="6" fill={c} />
      <rect x="36" y="26" width="148" height="88" rx="3" fill="#0a0a0a" />
      <rect x="14" y="118" width="192" height="10" rx="4" fill={c} />
      <rect x="96" y="120" width="28" height="4" rx="2" fill="#0a0a0a" />
    </svg>
  ),
  cir: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "70%", height: "70%" }}>
      <rect x="60" y="50" width="80" height="100" rx="22" fill={c} />
      <rect x="72" y="62" width="56" height="76" rx="12" fill="#0a0a0a" />
      <rect x="94" y="32" width="12" height="20" rx="3" fill={c} opacity="0.7" />
      <rect x="94" y="148" width="12" height="20" rx="3" fill={c} opacity="0.7" />
      <rect x="140" y="86" width="8" height="12" rx="2" fill={c} />
    </svg>
  ),
  phn: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "50%", height: "80%" }}>
      <rect x="60" y="20" width="80" height="160" rx="18" fill={c} />
      <rect x="68" y="28" width="64" height="144" rx="10" fill="#0a0a0a" />
      <rect x="86" y="30" width="28" height="6" rx="3" fill={c} />
    </svg>
  ),
  tab: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "72%", height: "72%" }}>
      <rect x="30" y="28" width="140" height="144" rx="12" fill={c} />
      <rect x="40" y="38" width="120" height="124" rx="4" fill="#0a0a0a" />
    </svg>
  ),
  cam: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "72%", height: "72%" }}>
      <rect x="30" y="60" width="140" height="90" rx="10" fill={c} />
      <rect x="70" y="46" width="34" height="20" rx="3" fill={c} />
      <circle cx="100" cy="105" r="30" fill="#0a0a0a" />
      <circle cx="100" cy="105" r="18" fill="#222" />
      <circle cx="100" cy="105" r="8" fill="#0a0a0a" />
      <circle cx="148" cy="76" r="3" fill="#ff3b30" />
    </svg>
  ),
  spk: (c) => (
    <svg viewBox="0 0 200 200" style={{ width: "55%", height: "55%" }}>
      <circle cx="100" cy="100" r="60" fill={c} />
      <circle cx="100" cy="100" r="40" fill="rgba(0,0,0,0.25)" />
      <circle cx="100" cy="100" r="16" fill={c} />
    </svg>
  ),
};

/* -------- UI atoms -------- */

function Nav({ active = "store", cartCount = 2 }) {
  const items = ["Store","Mac","iPad","iPhone","Watch","Audio","TV","Support"];
  return (
    <nav className="apple-nav">
      <div className="apple-nav-inner">
        <span className="apple-nav-logo" aria-label="brand">◯</span>
        {items.map(i => (
          <a key={i} className="apple-nav-link" href="#" style={i.toLowerCase() === active ? { color: "var(--ink)" } : {}}>{i}</a>
        ))}
        <a className="apple-nav-link" href="#"><IconSearch /></a>
        <a className="apple-nav-link" href="#" style={{ position: "relative" }}>
          <IconBag />
          {cartCount > 0 && <span style={{ position: "absolute", top: -4, right: -8, background: "var(--accent)", color: "white", fontSize: 10, fontWeight: 600, minWidth: 16, height: 16, padding: "0 4px", borderRadius: 999, display: "inline-flex", alignItems: "center", justifyContent: "center" }}>{cartCount}</span>}
        </a>
      </div>
    </nav>
  );
}

/* Lucide-style icons (monoline) */
function Icon({ size = 20, children, stroke = "currentColor", strokeWidth = 1.6 }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round">
      {children}
    </svg>
  );
}
const IconSearch = (p) => <Icon {...p}><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></Icon>;
const IconBag    = (p) => <Icon {...p}><path d="M6 8h12l-1 12H7L6 8Z" /><path d="M9 8V6a3 3 0 0 1 6 0v2" /></Icon>;
const IconHeart  = (p) => <Icon {...p}><path d="M12 20s-7-4.5-7-10a4 4 0 0 1 7-2.5A4 4 0 0 1 19 10c0 5.5-7 10-7 10Z" /></Icon>;
const IconUser   = (p) => <Icon {...p}><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></Icon>;
const IconMenu   = (p) => <Icon {...p}><path d="M3 6h18M3 12h18M3 18h18" /></Icon>;
const IconClose  = (p) => <Icon {...p}><path d="M6 6l12 12M18 6l-12 12" /></Icon>;
const IconPlus   = (p) => <Icon {...p}><path d="M12 5v14M5 12h14" /></Icon>;
const IconMinus  = (p) => <Icon {...p}><path d="M5 12h14" /></Icon>;
const IconChevL  = (p) => <Icon {...p}><path d="M15 6l-6 6 6 6" /></Icon>;
const IconChevR  = (p) => <Icon {...p}><path d="M9 6l6 6-6 6" /></Icon>;
const IconChevD  = (p) => <Icon {...p}><path d="M6 9l6 6 6-6" /></Icon>;
const IconCheck  = (p) => <Icon {...p}><path d="M4 12l5 5L20 6" /></Icon>;
const IconStar   = (p) => <Icon {...p}><path d="M12 3l2.6 5.8 6.4.6-4.8 4.4 1.4 6.2L12 17l-5.6 3 1.4-6.2L3 9.4l6.4-.6L12 3Z" /></Icon>;
const IconStarFill = ({ size=14 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24"><path d="M12 3l2.6 5.8 6.4.6-4.8 4.4 1.4 6.2L12 17l-5.6 3 1.4-6.2L3 9.4l6.4-.6L12 3Z" fill="#1d1d1f" /></svg>
);
const IconFilter = (p) => <Icon {...p}><path d="M3 5h18M6 12h12M10 19h4" /></Icon>;
const IconTruck  = (p) => <Icon {...p}><path d="M3 7h11v10H3zM14 10h5l2 3v4h-7" /><circle cx="7" cy="18" r="2" /><circle cx="17" cy="18" r="2" /></Icon>;
const IconShield = (p) => <Icon {...p}><path d="M12 3l8 3v6c0 5-3.5 8.5-8 9-4.5-.5-8-4-8-9V6l8-3Z" /></Icon>;
const IconReturn = (p) => <Icon {...p}><path d="M9 14l-4-4 4-4" /><path d="M5 10h10a5 5 0 0 1 0 10h-3" /></Icon>;
const IconLock   = (p) => <Icon {...p}><rect x="5" y="11" width="14" height="10" rx="2" /><path d="M8 11V8a4 4 0 0 1 8 0v3" /></Icon>;
const IconMap    = (p) => <Icon {...p}><path d="M9 4 3 6v14l6-2 6 2 6-2V4l-6 2-6-2z M9 4v14 M15 6v14" /></Icon>;
const IconBox    = (p) => <Icon {...p}><path d="M3 7l9-4 9 4v10l-9 4-9-4V7z M3 7l9 4 9-4 M12 11v10" /></Icon>;

/* Button */
function Btn({ children, kind = "secondary", size = "md", full, style, ...rest }) {
  const base = {
    fontFamily: "var(--font-text)",
    fontWeight: 400,
    letterSpacing: "-0.022em",
    border: 0,
    cursor: "pointer",
    transition: "all var(--dur-fast) var(--ease-standard)",
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    gap: 6,
  };
  const sizes = {
    sm: { padding: "6px 14px", fontSize: 13, borderRadius: 980, minHeight: 28 },
    md: { padding: "8px 18px", fontSize: 15, borderRadius: 980, minHeight: 36 },
    lg: { padding: "12px 26px", fontSize: 17, borderRadius: 980, minHeight: 44 },
  };
  const kinds = {
    primary: { background: "var(--accent)", color: "white" },
    secondary: { background: "rgba(0,0,0,0.04)", color: "var(--ink)", border: "1px solid transparent" },
    outline: { background: "transparent", color: "var(--ink)", border: "1px solid var(--ink)" },
    dark: { background: "var(--ink)", color: "white" },
    ghost: { background: "transparent", color: "var(--accent)" },
  };
  return (
    <button style={{ ...base, ...sizes[size], ...kinds[kind], width: full ? "100%" : undefined, ...style }} {...rest}>
      {children}
    </button>
  );
}

/* Apple chevron link */
function ChevronLink({ children, color = "var(--accent)", style, ...rest }) {
  return <a style={{ color, fontSize: 17, letterSpacing: "-0.022em", textDecoration: "none", ...style }} {...rest}>{children} ›</a>;
}

/* Rating row */
function Rating({ value, count, size = 14 }) {
  return (
    <span style={{ display: "inline-flex", alignItems: "center", gap: 4, fontSize: 13, color: "var(--ink-2)" }}>
      <IconStarFill size={size} />
      <span style={{ color: "var(--ink)" }}>{value}</span>
      {count != null && <span>({count.toLocaleString()})</span>}
    </span>
  );
}

/* Price formatter */
const fmt = n => `$${n.toLocaleString("en-US", { minimumFractionDigits: n % 1 ? 2 : 0 })}`;

/* Device frames */
function DesktopFrame({ children, url = "shop.com", w = 1280, h = 820, style }) {
  return (
    <div style={{ width: w, flexShrink: 0 }}>
      <div style={{
        width: w, height: h,
        borderRadius: 14,
        border: "1px solid var(--hairline)",
        background: "var(--surface)",
        overflow: "hidden",
        boxShadow: "0 20px 50px rgba(0,0,0,0.08)",
        ...style,
      }}>
        <div style={{
          height: 32, borderBottom: "1px solid var(--hairline)",
          display: "flex", alignItems: "center", gap: 6, padding: "0 12px",
          background: "var(--surface-2)",
        }}>
          <span style={{ width: 10, height: 10, borderRadius: 999, background: "#ff5f57" }} />
          <span style={{ width: 10, height: 10, borderRadius: 999, background: "#febc2e" }} />
          <span style={{ width: 10, height: 10, borderRadius: 999, background: "#28c840" }} />
          <div style={{
            flex: 1, margin: "0 14px", padding: "3px 12px",
            background: "var(--surface)", border: "1px solid var(--hairline)",
            borderRadius: 6, fontSize: 11, color: "var(--ink-3)",
            fontFamily: "var(--font-mono)", textAlign: "center",
          }}>{url}</div>
        </div>
        <div style={{ height: h - 32, overflow: "hidden", position: "relative" }}>
          {children}
        </div>
      </div>
    </div>
  );
}

function MobileFrame({ children, w = 340, h = 720, style }) {
  return (
    <div style={{ width: w, flexShrink: 0 }}>
      <div style={{
        width: w, height: h,
        borderRadius: 44,
        border: "1.5px solid #0a0a0a",
        padding: 8,
        background: "#0a0a0a",
        boxShadow: "0 20px 50px rgba(0,0,0,0.15)",
        ...style,
      }}>
        <div style={{
          width: "100%", height: "100%", borderRadius: 36,
          background: "var(--surface)",
          position: "relative", overflow: "hidden",
        }}>
          {/* dynamic island */}
          <div style={{
            position: "absolute", top: 10, left: "50%", transform: "translateX(-50%)",
            width: 100, height: 28, background: "#0a0a0a", borderRadius: 999, zIndex: 50,
          }} />
          {/* status bar */}
          <div style={{
            position: "absolute", top: 14, left: 24, fontSize: 13, fontWeight: 600, zIndex: 40, fontFamily: "var(--font-display)",
          }}>9:41</div>
          <div style={{
            position: "absolute", top: 14, right: 24, display: "flex", gap: 4, alignItems: "center", zIndex: 40,
          }}>
            <svg width="16" height="10" viewBox="0 0 16 10"><path d="M1 7h2v2H1zM5 5h2v4H5zM9 3h2v6H9zM13 1h2v8h-2z" fill="#1d1d1f"/></svg>
            <svg width="16" height="10" viewBox="0 0 24 10"><rect x="1" y="1" width="18" height="8" rx="2" stroke="#1d1d1f" fill="none" /><rect x="3" y="3" width="14" height="4" fill="#1d1d1f" rx="0.5"/><rect x="20" y="3.5" width="2" height="3" fill="#1d1d1f" rx="0.5"/></svg>
          </div>
          <div style={{ paddingTop: 50, height: h, position: "relative" }}>
            {children}
          </div>
        </div>
      </div>
    </div>
  );
}

/* Variation wrapper */
function Variation({ title, caption, desktop, mobile }) {
  return (
    <div className="variation-card">
      <div>
        <div style={{ fontFamily: "var(--font-display)", fontSize: 32, fontWeight: 600, letterSpacing: "-0.015em", color: "var(--ink)" }}>{title}</div>
        {caption && <div style={{ fontFamily: "var(--font-text)", fontSize: 17, color: "var(--ink-2)", letterSpacing: "-0.022em", maxWidth: 820, marginTop: 6 }}>{caption}</div>}
      </div>
      <div style={{ display: "flex", gap: 32, alignItems: "flex-start", marginTop: 18, flexWrap: "wrap" }}>
        {desktop}
        {mobile}
      </div>
    </div>
  );
}

Object.assign(window, {
  PRODUCTS, ProductImage, SILHOUETTES,
  Nav, DesktopFrame, MobileFrame, Variation,
  Btn, ChevronLink, Rating, fmt,
  IconSearch, IconBag, IconHeart, IconUser, IconMenu, IconClose, IconPlus, IconMinus,
  IconChevL, IconChevR, IconChevD, IconCheck, IconStar, IconStarFill, IconFilter,
  IconTruck, IconShield, IconReturn, IconLock, IconMap, IconBox,
});
