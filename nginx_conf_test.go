// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ngconf

import (
	"fmt"
)

func ExampleDecode() {
	const conf = `
	worker_processes auto;
	pid /run/nginx.pid;
	events {
		worker_connections 1024;
	}
	stream {
		upstream backend {
			hash $remote_addr consistent;
		}
		server {
			listen 127.0.0.1:8443;
			proxy_connect_timeout 1s;
			proxy_pass backend;
		}
	}


	http {
			# Basic Settings
			sendfile on;
			tcp_nopush on;
			tcp_nodelay on;
			keepalive_timeout 65;
			types_hash_max_size 2048;
			include /etc/nginx/mime.types;
			default_type application/octet-stream;
			# SSL Settings
			ssl_protocols TLSv1 TLSv1.1 TLSv1.2; # Dropping SSLv3, ref: POODLE
			ssl_prefer_server_ciphers on;
			# Logging Settings
			access_log /var/log/nginx/access.log;
			error_log /var/log/nginx/error.log;
			# Gzip Settings
			gzip on;
			gzip_disable "msie6";
			# Virtual Host Configs
			include /etc/nginx/conf.d/*.conf;
			include /etc/nginx/sites-enabled/*;
	}`

	root, err := Decode(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Remove a config block.
	root.Del("events")

	// Add a new config item.
	upstream := root.Get("stream")[0].Get("upstream")[0]
	upstream.Add("server", "backend1:443", "max_fails=3", "fail_timeout=30s")

	// Get a config item.
	hash := upstream.Get("hash", "$remote_addr")
	fmt.Println(len(hash), "node:", hash[0].Directive, hash[0].Args)

	// Print the whole config.
	fmt.Println("========================= Config =========================")
	fmt.Println(root)

	// Output:
	// 1 node: hash [$remote_addr consistent]
	// ========================= Config =========================
	// worker_processes auto;
	// pid /run/nginx.pid;
	//
	// stream {
	//     upstream backend {
	//         hash $remote_addr consistent;
	//         server backend1:443 max_fails=3 fail_timeout=30s;
	//     }
	//
	//     server {
	//         listen 127.0.0.1:8443;
	//         proxy_connect_timeout 1s;
	//         proxy_pass backend;
	//     }
	// }
	//
	// http {
	//     # Basic Settings sendfile on;
	//     tcp_nopush on;
	//     tcp_nodelay on;
	//     keepalive_timeout 65;
	//     types_hash_max_size 2048;
	//     include /etc/nginx/mime.types;
	//     default_type application/octet-stream;
	//
	//     # SSL Settings ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
	//     # Dropping SSLv3, ref: POODLE ssl_prefer_server_ciphers on;
	//     # Logging Settings access_log /var/log/nginx/access.log;
	//     error_log /var/log/nginx/error.log;
	//
	//     # Gzip Settings gzip on;
	//     gzip_disable "msie6";
	//
	//     # Virtual Host Configs include /etc/nginx/conf.d/*.conf;
	//     include /etc/nginx/sites-enabled/*;
	// }
}
