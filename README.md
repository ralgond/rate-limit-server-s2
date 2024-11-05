# rate-limit-server-s2
限流服务器（方案2）

## 架构图
![](images/Rate-Limit-System-Design-5.drawio.png)

## Nginx的配置
```bash
server {
        root /var/www/html;

        # Add index.php to the list if you are using PHP
        index index.html index.htm index.nginx-debian.html;

        server_name _;

        location /auth {
            internal;  # 只允许内部请求
            proxy_pass http://backend_ratelimit$request_uri;  # 鉴权服务的地址
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header Content-Length 0;
            proxy_set_header Content-Type "";
            proxy_set_header Connection "keep-alive";
            proxy_pass_request_body off;
        }

        location / {
            auth_request /auth;  # 调用鉴权服务
            proxy_pass http://backend1;  # 替换为你的后端服务
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header Connection "keep-alive";
            proxy_pass_request_body on;
        }
}
```
## 优势和劣势
### 优势
相对于方案1，rate-limit-server简单，无需再考虑负载均衡、安全性等问题，rate-limit-server只需考虑可用性和性能即可。

### 劣势
rate-limit-server占用了auth_request这个指令，且这个指令会将POST命令改写成GET命令才会发送给rate-limit-server。我可能需要新开发一个新的指令rate_limit来专门处理。
