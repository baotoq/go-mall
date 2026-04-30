export interface Product {
  id: number
  name: string
  price: number
  category: string
  description: string
  emoji: string
  rating: number
  reviews: number
  badge?: string
}

export interface Category {
  id: string
  name: string
  emoji: string
  count: number
}

export const categories: Category[] = [
  { id: "electronics", name: "Electronics", emoji: "📱", count: 4 },
  { id: "clothing", name: "Clothing", emoji: "👕", count: 4 },
  { id: "books", name: "Books", emoji: "📚", count: 2 },
  { id: "home", name: "Home & Garden", emoji: "🏡", count: 2 },
]

export const products: Product[] = [
  {
    id: 1,
    name: "Wireless Headphones",
    price: 79.99,
    category: "electronics",
    description:
      "High-quality wireless headphones with active noise cancellation and 30-hour battery life. Features premium sound drivers and comfortable over-ear cushions.",
    emoji: "🎧",
    rating: 4.5,
    reviews: 128,
    badge: "Best Seller",
  },
  {
    id: 2,
    name: "Smart Watch",
    price: 199.99,
    category: "electronics",
    description:
      "Feature-rich smartwatch with health tracking, GPS, sleep monitoring, and 7-day battery life. Compatible with iOS and Android.",
    emoji: "⌚",
    rating: 4.3,
    reviews: 89,
  },
  {
    id: 3,
    name: "Laptop Stand",
    price: 39.99,
    category: "electronics",
    description:
      "Adjustable aluminum laptop stand for improved ergonomics and cooling. Folds flat for easy portability. Fits laptops up to 17 inches.",
    emoji: "💻",
    rating: 4.7,
    reviews: 203,
  },
  {
    id: 4,
    name: "Mechanical Keyboard",
    price: 129.99,
    category: "electronics",
    description:
      "Compact TKL mechanical keyboard with RGB backlighting and tactile switches. USB-C connection with detachable cable.",
    emoji: "⌨️",
    rating: 4.6,
    reviews: 156,
    badge: "New",
  },
  {
    id: 5,
    name: "Classic T-Shirt",
    price: 24.99,
    category: "clothing",
    description:
      "Comfortable everyday t-shirt in 100% premium cotton. Pre-shrunk fabric, tagless design. Available in 12 colors and sizes XS–3XL.",
    emoji: "👕",
    rating: 4.4,
    reviews: 312,
  },
  {
    id: 6,
    name: "Running Sneakers",
    price: 89.99,
    category: "clothing",
    description:
      "Lightweight running shoes with responsive foam cushioning and breathable mesh upper. Reflective details for low-light visibility.",
    emoji: "👟",
    rating: 4.5,
    reviews: 267,
    badge: "Popular",
  },
  {
    id: 7,
    name: "Denim Jacket",
    price: 64.99,
    category: "clothing",
    description:
      "Classic denim jacket with a modern slim fit. Two chest pockets, button closure. Perfect for layering in any season.",
    emoji: "🧥",
    rating: 4.2,
    reviews: 94,
  },
  {
    id: 8,
    name: "Wool Beanie",
    price: 18.99,
    category: "clothing",
    description:
      "Warm merino wool beanie in a ribbed knit pattern. Stretchy one-size-fits-all design. Machine washable.",
    emoji: "🧢",
    rating: 4.8,
    reviews: 445,
  },
  {
    id: 9,
    name: "Clean Code",
    price: 34.99,
    category: "books",
    description:
      "A handbook of agile software craftsmanship by Robert C. Martin. Essential reading for every professional developer.",
    emoji: "📗",
    rating: 4.9,
    reviews: 891,
  },
  {
    id: 10,
    name: "Design Patterns",
    price: 44.99,
    category: "books",
    description:
      "Elements of reusable object-oriented software by the Gang of Four. The foundational reference for software design patterns.",
    emoji: "📘",
    rating: 4.7,
    reviews: 654,
    badge: "Classic",
  },
  {
    id: 11,
    name: "Indoor Plant Pot",
    price: 22.99,
    category: "home",
    description:
      "Minimalist ceramic plant pot with drainage hole and matching saucer. Perfect for succulents, herbs, or small houseplants.",
    emoji: "🪴",
    rating: 4.6,
    reviews: 187,
  },
  {
    id: 12,
    name: "Scented Candle Set",
    price: 29.99,
    category: "home",
    description:
      "Set of 3 hand-poured soy wax candles in lavender, vanilla, and cedar scents. 40-hour burn time each. Reusable glass jars.",
    emoji: "🕯️",
    rating: 4.8,
    reviews: 523,
    badge: "Gift Idea",
  },
]

export function getProductById(id: number): Product | undefined {
  return products.find((p) => p.id === id)
}

export function getProductsByCategory(category?: string): Product[] {
  if (!category || category === "all") return products
  return products.filter((p) => p.category === category)
}

export function getFeaturedProducts(): Product[] {
  return products.filter((p) => p.badge).slice(0, 4)
}
