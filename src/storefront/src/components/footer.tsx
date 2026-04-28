import Link from "next/link";

export function Footer() {
  return (
    <footer className="border-t border-ink/10 bg-surface-2 mt-auto">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          <div>
            <h3 className="text-xs font-semibold text-ink uppercase tracking-wider mb-4">
              Shop and Learn
            </h3>
            <ul className="space-y-2 text-sm text-ink-2">
              <li>
                <Link href="/shop" className="hover:text-ink transition-colors">
                  Store
                </Link>
              </li>
              <li>
                <Link
                  href="/shop?categoryId=mac"
                  className="hover:text-ink transition-colors"
                >
                  Mac
                </Link>
              </li>
              <li>
                <Link
                  href="/shop?categoryId=ipad"
                  className="hover:text-ink transition-colors"
                >
                  iPad
                </Link>
              </li>
              <li>
                <Link
                  href="/shop?categoryId=iphone"
                  className="hover:text-ink transition-colors"
                >
                  iPhone
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-xs font-semibold text-ink uppercase tracking-wider mb-4">
              Account
            </h3>
            <ul className="space-y-2 text-sm text-ink-2">
              <li>
                <Link
                  href="/account"
                  className="hover:text-ink transition-colors"
                >
                  Manage Your Apple Account
                </Link>
              </li>
              <li>
                <Link
                  href="/account"
                  className="hover:text-ink transition-colors"
                >
                  Apple Store Account
                </Link>
              </li>
              <li>
                <Link
                  href="/account"
                  className="hover:text-ink transition-colors"
                >
                  iCloud.com
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h3 className="text-xs font-semibold text-ink uppercase tracking-wider mb-4">
              About Apple
            </h3>
            <ul className="space-y-2 text-sm text-ink-2">
              <li>
                <Link href="#" className="hover:text-ink transition-colors">
                  Newsroom
                </Link>
              </li>
              <li>
                <Link href="#" className="hover:text-ink transition-colors">
                  Apple Leadership
                </Link>
              </li>
              <li>
                <Link href="#" className="hover:text-ink transition-colors">
                  Career Opportunities
                </Link>
              </li>
              <li>
                <Link href="#" className="hover:text-ink transition-colors">
                  Investors
                </Link>
              </li>
            </ul>
          </div>
        </div>
        <div className="mt-12 pt-8 border-t border-ink/10 flex flex-col md:flex-row justify-between items-center text-xs text-ink-3">
          <p>
            Copyright © {new Date().getFullYear()} Apple Inc. All rights
            reserved.
          </p>
          <div className="flex space-x-4 mt-4 md:mt-0">
            <Link href="#" className="hover:text-ink transition-colors">
              Privacy Policy
            </Link>
            <Link href="#" className="hover:text-ink transition-colors">
              Terms of Use
            </Link>
            <Link href="#" className="hover:text-ink transition-colors">
              Sales and Refunds
            </Link>
            <Link href="#" className="hover:text-ink transition-colors">
              Legal
            </Link>
            <Link href="#" className="hover:text-ink transition-colors">
              Site Map
            </Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
