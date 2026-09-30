FROM node:24-alpine
WORKDIR /app
COPY index.html serve.mjs ./
ENV HOST=0.0.0.0 PORT=4178
USER node
EXPOSE 4178
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=6 CMD node -e "fetch('http://127.0.0.1:4178/').then(response=>process.exit(response.ok?0:1)).catch(()=>process.exit(1))"
CMD ["node", "serve.mjs"]
