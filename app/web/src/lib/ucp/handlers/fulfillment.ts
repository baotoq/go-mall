export function getFulfillmentOptions() {
  return [
    {
      id: "standard",
      name: "Standard Shipping",
      price_cents: 599,
      currency: "USD",
      description: "5-7 business days",
    },
    {
      id: "express",
      name: "Express Shipping",
      price_cents: 1299,
      currency: "USD",
      description: "2-3 business days",
    },
  ];
}
