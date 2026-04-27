# Catalog Storefront Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a vertical slice connecting a go-zero catalog API to a Next.js Apple-style storefront.

**Architecture:** A three-layer go-zero backend (catalog service) serving product data over REST to a Next.js React Server Component frontend that renders alternating light/dark cinematic product sections.

**Tech Stack:** Go, go-zero, ent, Next.js (App Router), Tailwind CSS.

---

### Task 1: Initialize Database & Ent Schema

**Files:**
- Create: `src/catalog/ent/schema/product.go`
- Create: `src/catalog/ent/generate.go`
- Modify: `src/catalog/etc/catalog.yaml`
- Modify: `src/catalog/internal/config/config.go`
- Modify: `src/catalog/internal/svc/servicecontext.go`

- [ ] **Step 1: Define Product Ent Schema**
Write `src/catalog/ent/schema/product.go`:
```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").NotEmpty(),
		field.Int("price").Positive(), // Stored in cents
		field.String("image_url").NotEmpty(),
		field.Enum("theme").Values("light", "dark").Default("light"),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return nil
}
```

- [ ] **Step 2: Generate Ent Code**
Run:
```bash
cd src/catalog
go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema
```

- [ ] **Step 3: Add DB Config to Catalog Service**
Modify `src/catalog/etc/catalog.yaml`:
```yaml
Name: catalog.api
Host: 0.0.0.0
Port: 8888
DB:
  DataSource: "file:catalog.db?mode=memory&cache=shared&_fk=1"
```

Modify `src/catalog/internal/config/config.go`:
```go
package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DB struct {
		DataSource string
	}
}
```

- [ ] **Step 4: Initialize Ent Client in ServiceContext**
Modify `src/catalog/internal/svc/servicecontext.go`:
```go
package svc

import (
	"log"
	
	"go-mall/catalog/ent"
	"go-mall/catalog/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := ent.Open("sqlite3", c.DB.DataSource)
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return &ServiceContext{
		Config: c,
		DB:     client,
	}
}
```

- [ ] **Step 5: Commit**
```bash
git add src/catalog/ent src/catalog/etc src/catalog/internal
git commit -m "feat(catalog): setup ent schema and sqlite db config"
```

---

### Task 2: Implement Catalog Logic

**Files:**
- Modify: `src/catalog/product.api`
- Modify: `src/catalog/internal/logic/productlogic.go` (or similar depending on goctl generation)

- [ ] **Step 1: Define API Spec**
Modify `src/catalog/product.api`:
```api
syntax = "v1"

type Product {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	ImageURL    string `json:"image_url"`
	Theme       string `json:"theme"`
}

type (
	GetProductsReq {}
	GetProductsResp {
		Products []Product `json:"products"`
	}
)

service catalog-api {
	@handler GetProducts
	get /api/v1/products (GetProductsReq) returns (GetProductsResp)
}
```

- [ ] **Step 2: Generate API Code**
Run:
```bash
cd src/catalog
goctl api go -api product.api -dir .
```

- [ ] **Step 3: Implement GetProducts Logic**
Modify `src/catalog/internal/logic/getproductslogic.go`:
```go
package logic

import (
	"context"

	"go-mall/catalog/internal/svc"
	"go-mall/catalog/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductsLogic {
	return &GetProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProductsLogic) GetProducts(req *types.GetProductsReq) (resp *types.GetProductsResp, err error) {
	// Seed some data if empty
	count, _ := l.svcCtx.DB.Product.Query().Count(l.ctx)
	if count == 0 {
		l.svcCtx.DB.Product.Create().
			SetName("iPhone 15 Pro").
			SetDescription("Titanium. So strong. So light. So Pro.").
			SetPrice(99900).
			SetImageURL("/images/iphone15pro.png").
			SetTheme("dark").
			SaveX(l.ctx)
			
		l.svcCtx.DB.Product.Create().
			SetName("MacBook Air 15\"").
			SetDescription("Lean. Mean. M3 machine.").
			SetPrice(129900).
			SetImageURL("/images/macbookair.png").
			SetTheme("light").
			SaveX(l.ctx)
	}

	products, err := l.svcCtx.DB.Product.Query().All(l.ctx)
	if err != nil {
		return nil, err
	}

	var res []types.Product
	for _, p := range products {
		res = append(res, types.Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ImageURL:    p.ImageURL,
			Theme:       string(p.Theme),
		})
	}

	return &types.GetProductsResp{
		Products: res,
	}, nil
}
```

- [ ] **Step 4: Commit**
```bash
git add src/catalog/product.api src/catalog/internal/logic src/catalog/internal/types src/catalog/internal/handler
git commit -m "feat(catalog): implement GetProducts logic with mock seeding"
```

---

### Task 3: Setup Frontend Infrastructure & Tailwind

**Files:**
- Modify: `src/storefront/tailwind.config.ts`
- Modify: `src/storefront/src/app/globals.css`
- Modify: `src/storefront/src/app/layout.tsx`

- [ ] **Step 1: Configure Tailwind for Apple Design System**
Modify `src/storefront/tailwind.config.ts`:
```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        apple: {
          blue: "#0071e3",
          black: "#000000",
          gray: "#f5f5f7",
          textDark: "#1d1d1f",
          linkBlue: "#0066cc",
        },
      },
      fontFamily: {
        display: ['"SF Pro Display"', "Helvetica Neue", "Helvetica", "Arial", "sans-serif"],
        text: ['"SF Pro Text"', "Helvetica Neue", "Helvetica", "Arial", "sans-serif"],
      },
      letterSpacing: {
        tighter: '-0.374px',
        tightest: '-0.28px',
      }
    },
  },
  plugins: [],
};
export default config;
```

- [ ] **Step 2: Base CSS**
Modify `src/storefront/src/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    @apply font-text antialiased;
  }
  h1, h2, h3 {
    @apply font-display tracking-tightest;
  }
}
```

- [ ] **Step 3: Sticky Navigation**
Modify `src/storefront/src/app/layout.tsx` to include the sticky nav:
```tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Apple-Style Storefront",
  description: "E-commerce built with go-zero and Next.js",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="bg-apple-gray text-apple-textDark">
        <nav className="sticky top-0 z-50 h-12 w-full bg-black/80 backdrop-blur-md saturate-180 flex items-center justify-center">
          <div className="text-white text-xs font-text tracking-tighter">Storefront</div>
        </nav>
        <main>{children}</main>
      </body>
    </html>
  );
}
```

- [ ] **Step 4: Commit**
```bash
git add src/storefront/tailwind.config.ts src/storefront/src/app
git commit -m "feat(storefront): configure tailwind and global layout"
```

---

### Task 4: Implement Product Catalog Page

**Files:**
- Modify: `src/storefront/src/app/page.tsx`

- [ ] **Step 1: Fetch and Render Products**
Modify `src/storefront/src/app/page.tsx`:
```tsx
type Product = {
  id: number;
  name: string;
  description: string;
  price: number;
  image_url: string;
  theme: "light" | "dark";
};

async function getProducts(): Promise<Product[]> {
  // In a real setup, handle errors and potentially revalidate
  const res = await fetch("http://localhost:8888/api/v1/products", { cache: 'no-store' });
  if (!res.ok) {
    throw new Error("Failed to fetch products");
  }
  const data = await res.json();
  return data.products;
}

export default async function Home() {
  const products = await getProducts();

  return (
    <div className="w-full flex flex-col items-center">
      {products.map((product) => (
        <section
          key={product.id}
          className={`w-full min-h-[80vh] flex flex-col items-center pt-20 ${
            product.theme === "dark" ? "bg-black text-white" : "bg-apple-gray text-apple-textDark"
          }`}
        >
          <h2 className="text-[56px] font-semibold leading-[1.07] tracking-tightest mb-2 text-center">
            {product.name}
          </h2>
          <p className="text-[21px] font-normal leading-[1.19] mb-6 text-center max-w-[980px]">
            {product.description}
          </p>
          
          <div className="flex gap-4 z-10 mb-12">
            <a href={`/product/${product.id}`} className="px-6 py-3 rounded-full border border-apple-blue text-apple-blue font-text text-[17px] hover:bg-apple-blue hover:text-white transition-colors">
              Learn more {'>'}
            </a>
            <button className="px-6 py-3 rounded-full bg-apple-blue text-white font-text text-[17px]">
              Buy
            </button>
          </div>
          
          {/* Placeholder for image */}
          <div className="relative w-full max-w-[800px] h-[400px] bg-neutral-800/20 rounded-2xl flex items-center justify-center text-sm opacity-50">
             {product.image_url}
          </div>
        </section>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Commit**
```bash
git add src/storefront/src/app/page.tsx
git commit -m "feat(storefront): implement product listing page"
```
