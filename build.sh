name: Build and Push Docker Image

on:
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Install web dependencies
        run: |
          cd web
          pnpm install

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: 782042369
          password: Wm?r%2/by)5#)Hx
      - name: Build Docker image
        run: |
          cd service
          docker build --progress=plain -t 782042369/top1000-iyuu:latest .

      - name: Push Docker image
        run: |
          docker push 782042369/top1000-iyuu:latest
