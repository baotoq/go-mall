for file in src/app/page.tsx src/app/shop/page.tsx src/app/shop/[slug]/page.tsx src/components/header.tsx; do
  echo "export const dynamic = 'force-dynamic';" | cat - "$file" > temp && mv temp "$file"
done
