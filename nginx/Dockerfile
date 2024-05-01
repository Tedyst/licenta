FROM node:latest AS sveltebuilder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json /app/
RUN npm ci
COPY frontend/ /app/frontend
COPY api/ /app/api
RUN npm run build



FROM ubuntu:22.04 as nginxbuilder

RUN apt update \
    && apt upgrade -y \
    && apt install -y libpcre3 libpcre3-dev zlib1g zlib1g-dev openssl libssl-dev wget git gcc make libbrotli-dev

WORKDIR /app
RUN wget https://nginx.org/download/nginx-1.25.3.tar.gz && tar -zxf nginx-1.25.3.tar.gz
RUN git clone https://github.com/google/ngx_brotli && cd ngx_brotli && git submodule update --init --recursive && cd ..
RUN cd nginx-1.25.3 && ./configure --with-compat --add-dynamic-module=../ngx_brotli \
    && make modules

FROM nginx:1.25.3
COPY --from=nginxbuilder /app/nginx-1.25.3/objs/ngx_http_brotli_static_module.so /etc/nginx/modules/
COPY --from=nginxbuilder /app/nginx-1.25.3/objs/ngx_http_brotli_filter_module.so /etc/nginx/modules/
RUN echo "load_module modules/ngx_http_brotli_filter_module.so;\nload_module modules/ngx_http_brotli_static_module.so;\n$(cat /etc/nginx/nginx.conf)" > /etc/nginx/nginx.conf
RUN echo 'brotli on;\n \
brotli_comp_level 6;\n \
brotli_static on;\n \
brotli_types application/atom+xml application/javascript application/json application/rss+xml\n \
          application/vnd.ms-fontobject application/x-font-opentype application/x-font-truetype\n \
          application/x-font-ttf application/x-javascript application/xhtml+xml application/xml\n \
          font/eot font/opentype font/otf font/truetype image/svg+xml image/vnd.microsoft.icon\n \
          image/x-icon image/x-win-bitmap text/css text/javascript text/plain text/xml;' > /etc/nginx/conf.d/brotli.conf

COPY nginx/nginx.conf /etc/nginx/conf.d/svelte.conf.template

COPY --from=sveltebuilder /app/frontend/build /usr/share/nginx/html

RUN rm /etc/nginx/conf.d/default.conf

CMD envsubst '$UPSTREAM' < /etc/nginx/conf.d/svelte.conf.template > /etc/nginx/conf.d/svelte.conf && \
    exec nginx -g 'daemon off;'