# E-commerce Storefront & Catalog Vertical Slice

## Overview
Implement a vertical slice of an Apple-inspired e-commerce platform. This includes a basic go-zero catalog service (backend) delivering product data to a Next.js storefront (frontend) featuring the dark/light cinematic rhythm and typography defined in the project's DESIGN.md.

## Architecture

### Backend: Catalog Service (go-zero)
- **Role:** Source of truth for product data.
- **Framework:** go-zero (REST API).
- **Endpoints:**
  - `GET /api/v1/products`: List all products (basic pagination).
  - `GET /api/v1/products/:id`: Get detailed product information.
- **Data Model (ent/SQL):**
  - `Product`: ID, Name, Description, Price, ImageURL, Theme (Light/Dark).
- **Design Constraints:** 
  - Strict three-layer separation (Handler -> Logic -> Model).
  - Use `httpx` for standard error responses.

### Frontend: Storefront (Next.js)
- **Role:** Customer-facing application.
- **Framework:** Next.js (App Router).
- **Data Fetching:** Server Components fetching from the Catalog API.
- **Pages:**
  - `/`: Home page displaying the product catalog.
  - `/product/[id]`: Detailed product view.
- **Design System Implementation (per DESIGN.md):**
  - **Typography:** SF Pro Display/Text, tight line-heights (1.07-1.47), negative tracking.
  - **Colors:** Pure Black (`#000000`), Light Gray (`#f5f5f7`), Near Black (`#1d1d1f`).
  - **Interactive:** Apple Blue (`#0071e3`) exclusively for links, focus states, and primary CTAs.
  - **Components:**
    - Sticky "glass" navigation (`rgba(0,0,0,0.8)` with blur).
    - Hero Sections (Full width, alternating black/light gray).
    - 980px radius pill links for "Learn more" / "Buy".

## Data Flow
1. Next.js Server Component requests `GET /api/v1/products`.
2. go-zero `product-api` handles request, queries database via `ent`.
3. API returns JSON product array.
4. Next.js renders Apple-style full-width sections based on the product's `Theme` attribute (alternating dark/light).

## Error Handling
- **API:** standard go-zero error responses (code/msg format).
- **Frontend:** Next.js `error.tsx` boundary displaying a minimal, on-brand error state (SF Pro, light gray background).

## Testing
- **Backend:** Table-driven unit tests for the logic layer (`internal/logic`).
- **Frontend:** Ensure clean builds and no hydration errors.

## Out of Scope
- Cart and Payment services.
- User authentication.
- Complex product variants or faceted search.