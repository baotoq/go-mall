import ucpConfig from "../../../ucp.config.json";

export interface UCPAgentInfo {
  profile: string;
}

export interface NegotiationResult {
  capabilities: string[];
  version: string;
}

const NEGOTIATION_TIMEOUT_MS = 2_000;

// Parses UCP-Agent header: profile="<url>"
export function parseUCPAgent(
  header: string | null | undefined,
): UCPAgentInfo | null {
  if (!header) return null;
  const m = header.match(/profile="([^"]+)"/);
  if (!m) return null;
  return { profile: m[1] };
}

// Caller-supplied URL flows directly into fetch — must reject anything that
// could probe internal services (RFC1918, loopback, link-local, ULA, etc.)
// or downgrade to plain HTTP.
const PRIVATE_IPV4_PATTERNS: RegExp[] = [
  /^0\./, //                this network
  /^10\./, //               RFC1918
  /^127\./, //              loopback
  /^169\.254\./, //         link-local
  /^172\.(1[6-9]|2\d|3[0-1])\./, // RFC1918
  /^192\.168\./, //         RFC1918
  /^100\.(6[4-9]|[7-9]\d|1[01]\d|12[0-7])\./, // CGNAT 100.64.0.0/10
];

const PRIVATE_HOSTNAMES = new Set(["localhost", "0", "0.0.0.0", "::", "::1"]);

function isLiteralIpv6Private(host: string): boolean {
  const h = host.toLowerCase();
  if (h === "::1") return true;
  // fe80::/10 (link-local): fe80–febf
  if (/^fe[89ab]/.test(h)) return true;
  // fc00::/7 (ULA): fc + fd
  if (/^f[cd]/.test(h)) return true;
  // ::ffff:a.b.c.d (IPv4-mapped) — defer to v4 check on the trailing octets
  const v4Mapped = h.match(/^::ffff:([\d.]+)$/);
  if (v4Mapped) return PRIVATE_IPV4_PATTERNS.some((re) => re.test(v4Mapped[1]));
  return false;
}

export function isSafeProfileUrl(raw: string): boolean {
  let u: URL;
  try {
    u = new URL(raw);
  } catch {
    return false;
  }
  if (u.protocol !== "https:") return false;
  // Node URL preserves brackets on IPv6 in `.hostname` — strip them so the
  // host-pattern checks below see the bare address.
  const host = u.hostname.toLowerCase().replace(/^\[|\]$/g, "");
  if (!host) return false;
  if (PRIVATE_HOSTNAMES.has(host)) return false;
  if (PRIVATE_IPV4_PATTERNS.some((re) => re.test(host))) return false;
  if (host.includes(":") && isLiteralIpv6Private(host)) return false;
  return true;
}

// Computes capability intersection with platform profile; falls back to business caps
export async function negotiateCapabilities(
  platformProfileUrl?: string,
): Promise<NegotiationResult> {
  const businessCaps = [
    ...ucpConfig.capabilities.core,
    ...ucpConfig.capabilities.extensions,
  ];
  const fallback: NegotiationResult = {
    capabilities: businessCaps,
    version: ucpConfig.ucp_version,
  };
  if (!platformProfileUrl) return fallback;
  if (!isSafeProfileUrl(platformProfileUrl)) {
    console.warn(
      "[ucp] negotiation: rejected unsafe profile URL (non-https or private host)",
    );
    return fallback;
  }
  try {
    const res = await fetch(platformProfileUrl, {
      headers: { Accept: "application/json" },
      signal: AbortSignal.timeout(NEGOTIATION_TIMEOUT_MS),
      // Prevent followed redirects from re-entering blocked hosts.
      redirect: "error",
    });
    if (!res.ok) return fallback;
    const profile = await res.json();
    const platformCaps: string[] = (profile?.ucp?.capabilities ?? []).map(
      (c: { name: string }) => c.name,
    );
    const intersection = businessCaps.filter((c) => platformCaps.includes(c));
    return {
      capabilities: intersection.length > 0 ? intersection : businessCaps,
      version: ucpConfig.ucp_version,
    };
  } catch {
    return fallback;
  }
}
