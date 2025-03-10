# 构建阶段
FROM node:20 as build

WORKDIR /app

# 更新 npm 到最新版本
RUN npm install -g npm@latest

COPY package.json package-lock.json ./
RUN npm install

COPY . .
# 使用标准的 build 命令
RUN npm run build

# 生产阶段
FROM nginx:stable-alpine

COPY --from=build /app/build /usr/share/nginx/html

# 添加自定义 nginx 配置，将 API 请求代理到后端
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
