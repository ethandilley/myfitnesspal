FROM nginx:alpine
COPY index.html /usr/share/nginx/html/index.html
COPY start.sh /docker-entrypoint.d/40-inject-env.sh
RUN chmod +x /docker-entrypoint.d/40-inject-env.sh
