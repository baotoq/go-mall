import ucpConfig from "../../../ucp.config.json";

export interface UCPAgentInfo {
  profile: string;
}

export interface NegotiationResult {
  capabilities: string[];
  version: string;
}

// Parses UCP-Agent header: profile="<url>"
export function parseUCPAgent(
  header: string | null | undefined,
): UCPAgentInfo | null {
  if (!header) return null;
  const m = header.match(/profile="([^"]+)"/);
  if (!m) return null;
  return { profile: m[1] };
}

// Computes capability intersection with platform profile; falls back to business caps
export async function negotiateCapabilities(
  platformProfileUrl?: string,
): Promise<NegotiationResult> {
  const businessCaps = [
    ...ucpConfig.capabilities.core,
    ...ucpConfig.capabilities.extensions,
  ];
  if (!platformProfileUrl) {
    return { capabilities: businessCaps, version: ucpConfig.ucp_version };
  }
  try {
    const res = await fetch(platformProfileUrl, {
      headers: { Accept: "application/json" },
    });
    if (!res.ok)
      return { capabilities: businessCaps, version: ucpConfig.ucp_version };
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
    return { capabilities: businessCaps, version: ucpConfig.ucp_version };
  }
}
