FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ARG CATALOG_URL=http://localhost:9001
ARG CART_URL=http://localhost:9002
ARG PAYMENT_URL=http://localhost:9003
ENV NEXT_PUBLIC_CATALOG_URL=$CATALOG_URL
ENV NEXT_PUBLIC_CART_URL=$CART_URL
ENV NEXT_PUBLIC_PAYMENT_URL=$PAYMENT_URL
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
EXPOSE 3000
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
CMD ["node", "server.js"]
