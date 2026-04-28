/* Reusable sketch primitives. All components exported to window. */

const { useState, useRef, useEffect } = React;

/* ---------- atoms ---------- */

function Box({ className = "", style, children, fill, dashed, accent, accentFill, hatch, hatchDense, ...rest }) {
  const cls = [
    "sb",
    dashed && "sb-dashed",
    fill && "sb-fill",
    accent && "sb-accent",
    accentFill && "sb-accent-fill",
    hatch && "hatch",
    hatchDense && "hatch-dense",
    className,
  ].filter(Boolean).join(" ");
  return <div className={cls} style={style} {...rest}>{children}</div>;
}

function ImgPh({ w, h, label, style, className = "", dense }) {
  return (
    <div
      className={`sb img-x ${dense ? "hatch-dense" : "hatch"} ${className}`}
      style={{ width: w, height: h, position: "relative", ...style }}
    >
      {label && (
        <div style={{
          position: "absolute", inset: 0, display: "flex",
          alignItems: "center", justifyContent: "center",
          fontFamily: "Caveat, cursive", fontSize: 16, color: "var(--ink-2)",
          background: "rgba(246,241,231,0.55)", padding: "2px 8px",
          borderRadius: 3, width: "fit-content", height: "fit-content",
          margin: "auto", zIndex: 2,
        }}>{label}</div>
      )}
    </div>
  );
}

function Scribble({ lines = 3, widths }) {
  const defaults = ["s-1", "s-2", "s-3", "s-2", "s-4"];
  return (
    <div>
      {Array.from({ length: lines }).map((_, i) => (
        <span key={i} className={`scribble ${widths?.[i] || defaults[i % defaults.length]}`} />
      ))}
    </div>
  );
}

function Btn({ children, fill, accent, accentFill, size = "md", style, ...rest }) {
  const cls = ["sb-btn", fill && "sb-btn-fill", accent && "sb-btn-accent", accentFill && "sb-btn-accent-fill"].filter(Boolean).join(" ");
  const sizeStyle = size === "lg" ? { padding: "9px 20px", fontSize: 18 } : size === "sm" ? { padding: "3px 10px", fontSize: 13 } : {};
  return <button className={cls} style={{ ...sizeStyle, ...style }} {...rest}>{children}</button>;
}

function Tag({ children, accent }) {
  return <span className="sb-tag" style={accent ? { color: "var(--accent)", borderColor: "var(--accent)" } : {}}>{children}</span>;
}

function Annot({ children, style, arrow, color = "var(--accent)" }) {
  return (
    <div className="annot" style={{ color, ...style }}>
      {arrow === "left" && <span>← </span>}
      {children}
      {arrow === "right" && <span> →</span>}
    </div>
  );
}

/* A hand-drawn arrow that goes from (x1,y1) to (x2,y2) with a slight curve */
function Arrow({ from, to, curve = 0.3, color = "var(--accent)", label, labelOffset = [0, -10], strokeWidth = 1.5, style }) {
  const [x1, y1] = from;
  const [x2, y2] = to;
  const mx = (x1 + x2) / 2;
  const my = (y1 + y2) / 2;
  const dx = x2 - x1;
  const dy = y2 - y1;
  // perpendicular offset for control point
  const cx = mx - dy * curve;
  const cy = my + dx * curve;

  const minX = Math.min(x1, x2, cx) - 20;
  const minY = Math.min(y1, y2, cy) - 20;
  const maxX = Math.max(x1, x2, cx) + 20;
  const maxY = Math.max(y1, y2, cy) + 20;

  return (
    <svg
      className="arrow-svg"
      style={{ left: minX, top: minY, width: maxX - minX, height: maxY - minY, ...style }}
      viewBox={`${minX} ${minY} ${maxX - minX} ${maxY - minY}`}
    >
      <defs>
        <marker id={`arrh-${color.replace(/[^a-z0-9]/gi, '')}`} markerWidth="10" markerHeight="10" refX="6" refY="3" orient="auto">
          <path d="M0,0 L6,3 L0,6" fill="none" stroke={color} strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
        </marker>
      </defs>
      <path
        d={`M ${x1} ${y1} Q ${cx} ${cy} ${x2} ${y2}`}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        markerEnd={`url(#arrh-${color.replace(/[^a-z0-9]/gi, '')})`}
      />
      {label && (
        <text
          x={cx + labelOffset[0]}
          y={cy + labelOffset[1]}
          fontFamily="Caveat, cursive"
          fontSize="16"
          fill={color}
          textAnchor="middle"
        >{label}</text>
      )}
    </svg>
  );
}

/* ---------- frames ---------- */

function Frame({ w = 1280, h = 800, kind = "desktop", children, style, label }) {
  // minimal browser/mobile chrome
  if (kind === "desktop") {
    return (
      <div style={{ width: w, flexShrink: 0 }}>
        <div className="wf-frame" style={{ width: w, height: h, ...style }}>
          {/* browser chrome */}
          <div style={{
            height: 28, borderBottom: "1.5px solid var(--ink)",
            display: "flex", alignItems: "center", gap: 6, padding: "0 10px",
            background: "rgba(0,0,0,0.02)",
          }}>
            <span className="sb-circle" style={{ width: 10, height: 10 }} />
            <span className="sb-circle" style={{ width: 10, height: 10 }} />
            <span className="sb-circle" style={{ width: 10, height: 10 }} />
            <div style={{
              flex: 1, height: 16, margin: "0 12px",
              border: "1.2px solid var(--ink-2)", borderRadius: 999,
              background: "rgba(0,0,0,0.02)",
              fontFamily: "JetBrains Mono, monospace", fontSize: 10,
              display: "flex", alignItems: "center", padding: "0 10px",
              color: "var(--ink-3)",
            }}>shop.com{label ? ` / ${label}` : ""}</div>
          </div>
          <div style={{ height: h - 28, overflow: "hidden", position: "relative" }}>
            {children}
          </div>
        </div>
      </div>
    );
  }
  // mobile
  const mW = 300, mH = 620;
  return (
    <div style={{ width: mW, flexShrink: 0 }}>
      <div
        className="wf-frame"
        style={{
          width: mW, height: mH, borderRadius: 28, border: "1.5px solid var(--ink)",
          padding: 6, background: "var(--paper)",
        }}
      >
        <div style={{
          width: "100%", height: "100%", borderRadius: 22,
          border: "1.2px solid var(--ink-2)",
          position: "relative", overflow: "hidden",
        }}>
          {/* notch */}
          <div style={{
            position: "absolute", top: 6, left: "50%", transform: "translateX(-50%)",
            width: 80, height: 18, background: "var(--ink)", borderRadius: 999, zIndex: 10,
          }} />
          {/* status text */}
          <div style={{
            position: "absolute", top: 8, left: 14, fontSize: 10,
            fontFamily: "JetBrains Mono, monospace", color: "var(--ink-2)", zIndex: 5,
          }}>9:41</div>
          <div style={{
            position: "absolute", top: 8, right: 14, fontSize: 10,
            fontFamily: "JetBrains Mono, monospace", color: "var(--ink-2)", zIndex: 5,
          }}>••• ▮</div>
          <div style={{ paddingTop: 32, height: mH - 12, position: "relative" }}>
            {children}
          </div>
        </div>
      </div>
    </div>
  );
}

/* ---------- layout helpers ---------- */

function Row({ children, gap = 8, style, ...rest }) {
  return <div style={{ display: "flex", gap, ...style }} {...rest}>{children}</div>;
}
function Col({ children, gap = 8, style, ...rest }) {
  return <div style={{ display: "flex", flexDirection: "column", gap, ...style }} {...rest}>{children}</div>;
}

/* Section wrapper for a single variation */
function Variation({ title, caption, desktop, mobile, annotations }) {
  return (
    <div className="var-card">
      <div>
        <div className="hand-title" style={{ fontSize: 26, lineHeight: 1.1, color: "var(--ink)" }}>{title}</div>
        {caption && <div className="wf-caption">{caption}</div>}
      </div>
      <div className="var-row" style={{ position: "relative" }}>
        {desktop}
        {mobile}
        {annotations}
      </div>
    </div>
  );
}

Object.assign(window, { Box, ImgPh, Scribble, Btn, Tag, Annot, Arrow, Frame, Row, Col, Variation });
