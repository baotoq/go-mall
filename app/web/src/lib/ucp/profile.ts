import ucpConfig from "../../../ucp.config.json";

export interface CapabilityDefinition {
  name: string;
  version: string;
}

export interface UCPProfile {
  ucp: {
    version: string;
    services: {
      "dev.ucp.shopping": {
        version: string;
        rest?: { endpoint: string };
        mcp?: { endpoint: string };
      };
    };
    capabilities: CapabilityDefinition[];
  };
}

export function generateProfile(): UCPProfile {
  const domain = process.env.UCP_DOMAIN || "localhost:3000";
  const protocol = domain.startsWith("localhost") ? "http" : "https";
  const base = `${protocol}://${domain}`;

  const allCaps = [
    ...ucpConfig.capabilities.core,
    ...ucpConfig.capabilities.extensions,
  ];

  return {
    ucp: {
      version: ucpConfig.ucp_version,
      services: {
        "dev.ucp.shopping": {
          version: ucpConfig.ucp_version,
          ...(ucpConfig.transports.includes("rest") && {
            rest: { endpoint: `${base}/api/ucp` },
          }),
          ...(ucpConfig.transports.includes("mcp") && {
            mcp: { endpoint: `${base}/api/mcp` },
          }),
        },
      },
      capabilities: allCaps.map((name) => ({
        name,
        version: ucpConfig.ucp_version,
      })),
    },
  };
}
