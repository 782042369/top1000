cd web && pnpm i && pnpm build
cd ../service && docker buildx build --platform linux/amd64,linux/arm64 -t 782042369/top1000-iyuu:v0.1 . --push
