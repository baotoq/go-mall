import { Button } from "@/components/ui/button";

export default function AccountPage() {
  const mockedOrders = [
    {
      id: "W83921047",
      date: "October 14, 2023",
      total: 1098.0,
      status: "Delivered",
      items: [
        { name: "XR-7 Wireless Headphones", quantity: 1, price: 349.0 },
        { name: "Pro Laptops M3", quantity: 1, price: 749.0 },
      ],
    },
    {
      id: "W71029482",
      date: "September 02, 2023",
      total: 49.0,
      status: "Shipped",
      items: [{ name: "Woven USB-C Cable", quantity: 1, price: 49.0 }],
    },
  ];

  return (
    <div className="mx-auto w-full max-w-4xl px-4 sm:px-6 lg:px-8 py-16 md:py-24 min-h-screen">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12 border-b border-ink/10 pb-8">
        <div>
          <h1 className="text-4xl md:text-5xl font-semibold tracking-tighter text-ink mb-2">
            Your Account.
          </h1>
          <p className="text-lg text-ink-2">
            Manage your orders and preferences.
          </p>
        </div>
        <Button variant="outline" className="rounded-full h-10 px-6">
          Sign Out
        </Button>
      </div>

      <div className="flex flex-col gap-12">
        <section>
          <h2 className="text-2xl font-semibold tracking-tight text-ink mb-6">
            Recent Orders
          </h2>
          <div className="flex flex-col gap-6">
            {mockedOrders.map((order) => (
              <div
                key={order.id}
                className="border border-ink/10 rounded-3xl p-6 md:p-8 bg-surface-2 transition-colors hover:border-ink/20"
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 pb-6 border-b border-ink/10">
                  <div>
                    <p className="text-sm font-medium text-ink-2">
                      Order {order.id}
                    </p>
                    <p className="text-sm text-ink-2">Placed on {order.date}</p>
                  </div>
                  <div className="text-left sm:text-right">
                    <p className="text-lg font-semibold text-ink">
                      ${order.total.toFixed(2)}
                    </p>
                    <div className="inline-flex items-center gap-2 mt-1">
                      <span
                        className={`w-2 h-2 rounded-full ${order.status === "Delivered" ? "bg-green-500" : "bg-blue-500"}`}
                      />
                      <span className="text-sm font-medium text-ink">
                        {order.status}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col gap-4">
                  {order.items.map((item, _idx) => (
                    <div
                      key={item.name}
                      className="flex items-center justify-between text-sm"
                    >
                      <div className="flex items-center gap-4">
                        <div className="h-12 w-12 rounded-lg bg-surface flex items-center justify-center text-ink-3">
                          Image
                        </div>
                        <div>
                          <p className="font-medium text-ink">{item.name}</p>
                          <p className="text-ink-2">Qty: {item.quantity}</p>
                        </div>
                      </div>
                      <p className="font-medium text-ink">
                        ${item.price.toFixed(2)}
                      </p>
                    </div>
                  ))}
                </div>

                <div className="mt-8 pt-6 border-t border-ink/10 flex flex-wrap gap-4">
                  <Button
                    variant="outline"
                    className="rounded-full h-10 px-6 text-sm"
                  >
                    Track Order
                  </Button>
                  <Button
                    variant="outline"
                    className="rounded-full h-10 px-6 text-sm"
                  >
                    Return Items
                  </Button>
                  <Button
                    variant="outline"
                    className="rounded-full h-10 px-6 text-sm"
                  >
                    View Invoice
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="border border-ink/10 rounded-3xl p-8 hover:bg-surface-2 transition-colors cursor-pointer group">
            <h3 className="text-xl font-semibold text-ink mb-2 group-hover:text-accent transition-colors">
              Payment Methods
            </h3>
            <p className="text-sm text-ink-2 mb-6">
              Manage your saved credit cards and payment preferences.
            </p>
            <p className="text-sm font-medium text-ink">Visa ending in 4242</p>
          </div>

          <div className="border border-ink/10 rounded-3xl p-8 hover:bg-surface-2 transition-colors cursor-pointer group">
            <h3 className="text-xl font-semibold text-ink mb-2 group-hover:text-accent transition-colors">
              Shipping Addresses
            </h3>
            <p className="text-sm text-ink-2 mb-6">
              Manage addresses for fast and easy checkout.
            </p>
            <p className="text-sm font-medium text-ink">
              123 Apple Park Way
              <br />
              Cupertino, CA 95014
            </p>
          </div>
        </section>
      </div>
    </div>
  );
}
